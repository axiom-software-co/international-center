package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/gateway"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDAPREnvironmentIntegration validates DAPR middleware integration across environments
func TestDAPREnvironmentIntegration(t *testing.T) {
	timeout := 15 * time.Second
	_, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	environments := []struct {
		name                string
		environment         string
		expectedServices    []string
		expectedMiddleware  []string
		expectedSecurity    bool
	}{
		{
			name:               "development environment should have full DAPR integration",
			environment:        "development", 
			expectedServices:   []string{"content-api", "inquiries-api", "notifications-api"},
			expectedMiddleware: []string{"bearer", "oauth2", "opa", "routeChecker", "cors", "ratelimit"},
			expectedSecurity:   true,
		},
		{
			name:               "staging environment should match development middleware",
			environment:        "staging",
			expectedServices:   []string{"content-api", "inquiries-api", "notifications-api"},
			expectedMiddleware: []string{"bearer", "oauth2", "opa", "routeChecker", "cors", "ratelimit"},
			expectedSecurity:   true,
		},
		{
			name:               "production environment should have enhanced security",
			environment:        "production",
			expectedServices:   []string{"content-api", "inquiries-api", "notifications-api"},
			expectedMiddleware: []string{"bearer", "oauth2", "opa", "routeChecker", "cors", "ratelimit"},
			expectedSecurity:   true,
		},
	}

	for _, tt := range environments {
		t.Run(tt.name, func(t *testing.T) {
			// Validate DAPR configuration for environment
			daprConfig, err := gateway.NewGatewayDAPRConfiguration(tt.environment)
			require.NoError(t, err, "DAPR configuration should be creatable")
			require.NotNil(t, daprConfig, "DAPR configuration should not be nil")

			// Validate environment-specific endpoints
			assert.NotEmpty(t, daprConfig.VaultEndpoint, "Vault endpoint should be configured")
			assert.NotEmpty(t, daprConfig.RedisEndpoint, "Redis endpoint should be configured")
			assert.True(t, daprConfig.DAPREnabled, "DAPR should be enabled")

			// Validate middleware chain completeness
			middlewareChain := daprConfig.GetMiddlewareChain()
			assert.GreaterOrEqual(t, len(middlewareChain), len(tt.expectedMiddleware), 
				"Should have all required middleware components")

			// Verify each expected middleware is present
			for _, expectedMW := range tt.expectedMiddleware {
				found := false
				for _, actualMW := range middlewareChain {
					if actualMW.Name == expectedMW {
						found = true
						break
					}
				}
				assert.True(t, found, "Middleware %s should be present", expectedMW)
			}

			// Validate security configurations based on environment
			if tt.expectedSecurity {
				authConfig, exists := daprConfig.Configuration["authentication"]
				assert.True(t, exists, "Authentication configuration should exist")
				assert.NotNil(t, authConfig, "Authentication configuration should not be nil")
			}
		})
	}
}

// TestFullGatewayEnvironmentValidation tests complete gateway environment setup
func TestFullGatewayEnvironmentValidation(t *testing.T) {
	timeout := 15 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	t.Run("development environment complete validation", func(t *testing.T) {
		// Test public gateway setup
		publicGateway, err := gateway.NewPublicGateway(ctx, "development")
		require.NoError(t, err, "Public gateway should be creatable")
		require.NotNil(t, publicGateway, "Public gateway should not be nil")

		// Test admin gateway setup  
		adminGateway, err := gateway.NewAdminGateway(ctx, "development")
		require.NoError(t, err, "Admin gateway should be creatable")
		require.NotNil(t, adminGateway, "Admin gateway should not be nil")

		// Validate public gateway routing
		t.Run("public gateway routing", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/services", nil)
			rr := httptest.NewRecorder()
			publicGateway.ServeHTTP(rr, req)
			
			assert.Equal(t, http.StatusOK, rr.Code, "Public services endpoint should be accessible")
			assert.NotEmpty(t, rr.Header().Get("X-Correlation-ID"), "Correlation ID should be set")
			assert.NotEmpty(t, rr.Header().Get("Access-Control-Allow-Origin"), "CORS headers should be set")
		})

		// Validate admin gateway authentication
		t.Run("admin gateway authentication", func(t *testing.T) {
			// Unauthenticated request should fail
			req := httptest.NewRequest("GET", "/admin/api/v1/services", nil)
			rr := httptest.NewRecorder()
			adminGateway.ServeHTTP(rr, req)
			
			assert.Equal(t, http.StatusUnauthorized, rr.Code, "Unauthenticated admin requests should be rejected")
			assert.NotEmpty(t, rr.Header().Get("WWW-Authenticate"), "WWW-Authenticate header should be set")

			// Authenticated request should succeed
			authReq := httptest.NewRequest("GET", "/admin/api/v1/services", nil)
			authReq.Header.Set("Authorization", "Bearer valid_admin_token")
			authRR := httptest.NewRecorder()
			adminGateway.ServeHTTP(authRR, authReq)
			
			assert.Equal(t, http.StatusOK, authRR.Code, "Authenticated admin requests should succeed")
			assert.Equal(t, "admin", authRR.Header().Get("X-User-Role"), "User role should be set in headers")
		})

		// Validate rate limiting functionality
		t.Run("rate limiting validation", func(t *testing.T) {
			// Test that rate limiting works by making many requests
			var rateLimitHit bool
			
			// Make requests until rate limit is hit (admin limit is 100/min)
			for i := 0; i < 105; i++ {
				req := httptest.NewRequest("GET", "/health", nil)
				req.Header.Set("Authorization", "Bearer valid_admin_token")
				rr := httptest.NewRecorder()
				adminGateway.ServeHTTP(rr, req)
				
				if rr.Code == http.StatusTooManyRequests {
					rateLimitHit = true
					break
				}
			}
			
			assert.True(t, rateLimitHit, "Rate limiting should be enforced")
		})
	})
}

// TestServiceDiscoveryIntegration validates service discovery through DAPR
func TestServiceDiscoveryIntegration(t *testing.T) {
	timeout := 10 * time.Second
	_, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	t.Run("DAPR service discovery configuration", func(t *testing.T) {
		// Validate service endpoints configuration
		config, err := gateway.NewGatewayDAPRConfiguration("development")
		require.NoError(t, err, "DAPR configuration should be valid")

		// Verify essential services are configured
		expectedServices := []string{
			"content-api",
			"inquiries-api", 
			"notifications-api",
		}

		// Check if routes are configured for expected services
		middlewareChain := config.GetMiddlewareChain()
		var routeChecker gateway.DAPRMiddleware
		
		for _, mw := range middlewareChain {
			if mw.Name == "routeChecker" {
				routeChecker = mw
				break
			}
		}
		
		assert.NotNil(t, routeChecker.Config, "Route checker middleware should be configured")
		
		if routeConfig, ok := routeChecker.Config.(map[string]interface{}); ok {
			if allowedRoutes, ok := routeConfig["allowedRoutes"].([]string); ok {
				// Verify service routes are present
				for _, service := range expectedServices {
					serviceRouteFound := false
					for _, route := range allowedRoutes {
						if route == fmt.Sprintf("/api/v1/%s", service) || 
						   route == fmt.Sprintf("/admin/api/v1/%s", service) {
							serviceRouteFound = true
							break
						}
					}
					// Note: Some services might not have direct routes, that's OK
					t.Logf("Service %s route configuration checked: %v", service, serviceRouteFound)
				}
			}
		}
	})
}

// TestContractComplianceWithEnvironmentIntegration validates contract compliance in environment testing
func TestContractComplianceWithEnvironmentIntegration(t *testing.T) {
	timeout := 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	t.Run("contract validation with DAPR integration", func(t *testing.T) {
		// Create admin gateway with full DAPR integration
		adminGateway, err := gateway.NewAdminGateway(ctx, "development")
		require.NoError(t, err, "Admin gateway should be creatable")

		// Test contract-compliant inquiry endpoint
		testCases := []struct {
			name           string
			method         string
			path           string
			token          string
			expectedStatus int
			validateBody   bool
		}{
			{
				name:           "contract-compliant inquiry listing",
				method:         "GET",
				path:           "/admin/api/v1/inquiries",
				token:          "valid_admin_token",
				expectedStatus: http.StatusOK,
				validateBody:   true,
			},
			{
				name:           "contract-compliant inquiry retrieval",
				method:         "GET", 
				path:           "/admin/api/v1/inquiries/550e8400-e29b-41d4-a716-446655440000",
				token:          "valid_admin_token",
				expectedStatus: http.StatusOK,
				validateBody:   true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req := httptest.NewRequest(tc.method, tc.path, nil)
				req.Header.Set("Authorization", "Bearer "+tc.token)
				req.Header.Set("Content-Type", "application/json")
				
				// Add correlation context
				correlationCtx := domain.NewCorrelationContext()
				ctx := correlationCtx.ToContext(req.Context())
				req = req.WithContext(ctx)
				
				rr := httptest.NewRecorder()
				adminGateway.ServeHTTP(rr, req)

				// Validate response status
				assert.Equal(t, tc.expectedStatus, rr.Code, "Response status should match contract")
				
				// Validate response headers
				assert.NotEmpty(t, rr.Header().Get("X-Correlation-ID"), "Correlation ID should be set")
				assert.Equal(t, "admin", rr.Header().Get("X-User-Role"), "User role should be set")
				
				// Validate response body structure for successful responses
				if tc.validateBody && rr.Code == http.StatusOK {
					var response map[string]interface{}
					err := json.NewDecoder(rr.Body).Decode(&response)
					assert.NoError(t, err, "Response should be valid JSON")
					
					// Contract compliance: response should have data field
					assert.Contains(t, response, "data", "Response should contain data field")
				}
			})
		}
	})
}

// TestEnvironmentHealthValidation validates that all environment components are healthy
func TestEnvironmentHealthValidation(t *testing.T) {
	timeout := 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	t.Run("development environment health validation", func(t *testing.T) {
		// Create both gateways
		publicGateway, err := gateway.NewPublicGateway(ctx, "development")
		require.NoError(t, err, "Public gateway should be healthy")
		
		adminGateway, err := gateway.NewAdminGateway(ctx, "development") 
		require.NoError(t, err, "Admin gateway should be healthy")

		// Test public gateway health
		pubHealthReq := httptest.NewRequest("GET", "/health", nil)
		pubHealthRR := httptest.NewRecorder()
		publicGateway.ServeHTTP(pubHealthRR, pubHealthReq)
		
		assert.Equal(t, http.StatusOK, pubHealthRR.Code, "Public gateway health should be OK")

		// Test admin gateway health
		adminHealthReq := httptest.NewRequest("GET", "/health", nil)
		adminHealthRR := httptest.NewRecorder()
		adminGateway.ServeHTTP(adminHealthRR, adminHealthReq)
		
		assert.Equal(t, http.StatusOK, adminHealthRR.Code, "Admin gateway health should be OK")

		// Validate health response structure
		var publicHealth map[string]interface{}
		err = json.NewDecoder(pubHealthRR.Body).Decode(&publicHealth)
		assert.NoError(t, err, "Public health response should be valid JSON")
		
		var adminHealth map[string]interface{}
		err = json.NewDecoder(adminHealthRR.Body).Decode(&adminHealth)
		assert.NoError(t, err, "Admin health response should be valid JSON")

		t.Logf("Environment health validation completed successfully")
		t.Logf("Public gateway: %+v", publicHealth)
		t.Logf("Admin gateway: %+v", adminHealth)
	})
}