package contracts

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/gateway"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Gateway Authentication Middleware Contract Tests
//
// WHY: Production-ready API gateways require secure route protection with RBAC integration
// SCOPE: Route protection, admin vs public differentiation, DAPR OPA policy enforcement
// DEPENDENCIES: Authentication service contracts, JWT middleware contracts
// CONTEXT: Admin and public gateway separation with DAPR sidecar middleware chain

func TestPublicGatewayAuthentication_Contract(t *testing.T) {
	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	tests := []struct {
		name           string
		method         string
		path           string
		headers        map[string]string
		expectedStatus int
		expectAuth     bool
		expectCORS     bool
	}{
		{
			name:           "public services endpoint should allow anonymous access",
			method:         "GET",
			path:           "/api/v1/services",
			headers:        map[string]string{},
			expectedStatus: http.StatusOK,
			expectAuth:     false,
			expectCORS:     true,
		},
		{
			name:           "public news endpoint should allow anonymous access",
			method:         "GET",
			path:           "/api/v1/news",
			headers:        map[string]string{},
			expectedStatus: http.StatusOK,
			expectAuth:     false,
			expectCORS:     true,
		},
		{
			name:           "public inquiry submission should allow anonymous access",
			method:         "POST",
			path:           "/api/v1/inquiries/media",
			headers:        map[string]string{"Content-Type": "application/json"},
			expectedStatus: http.StatusAccepted,
			expectAuth:     false,
			expectCORS:     true,
		},
		{
			name:           "health check should allow anonymous access",
			method:         "GET",
			path:           "/health",
			headers:        map[string]string{},
			expectedStatus: http.StatusOK,
			expectAuth:     false,
			expectCORS:     false,
		},
		{
			name:           "admin routes should be blocked on public gateway",
			method:         "GET",
			path:           "/admin/api/v1/services",
			headers:        map[string]string{},
			expectedStatus: http.StatusNotFound,
			expectAuth:     false,
			expectCORS:     false,
		},
		{
			name:           "CORS preflight should be handled correctly",
			method:         "OPTIONS",
			path:           "/api/v1/services",
			headers:        map[string]string{"Origin": "https://international-center.dev"},
			expectedStatus: http.StatusOK,
			expectAuth:     false,
			expectCORS:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Contract: Public gateway must be creatable with proper configuration
			publicGateway, err := gateway.NewPublicGateway(ctx, "development")
			require.NoError(t, err, "Public gateway creation should succeed")
			require.NotNil(t, publicGateway, "Public gateway should not be nil")

			// Contract: Gateway should handle HTTP requests according to routing rules
			req := httptest.NewRequest(tt.method, tt.path, nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			recorder := httptest.NewRecorder()
			publicGateway.ServeHTTP(recorder, req)

			// Contract: Response status should match expected behavior
			assert.Equal(t, tt.expectedStatus, recorder.Code, "Response status should match expected")

			// Contract: Authentication headers should be handled appropriately
			if !tt.expectAuth {
				assert.Empty(t, recorder.Header().Get("WWW-Authenticate"), "WWW-Authenticate header should not be set for public routes")
			}

			// Contract: CORS headers should be set for appropriate routes
			if tt.expectCORS {
				assert.NotEmpty(t, recorder.Header().Get("Access-Control-Allow-Origin"), "CORS headers should be set for public API routes")
				assert.NotEmpty(t, recorder.Header().Get("Access-Control-Allow-Methods"), "CORS methods should be set")
			}

			// Contract: Security headers should always be present
			assert.Equal(t, "nosniff", recorder.Header().Get("X-Content-Type-Options"), "X-Content-Type-Options should be set")
			assert.Equal(t, "DENY", recorder.Header().Get("X-Frame-Options"), "X-Frame-Options should be set")
			assert.Equal(t, "1; mode=block", recorder.Header().Get("X-XSS-Protection"), "X-XSS-Protection should be set")
		})
	}
}

func TestAdminGatewayAuthentication_Contract(t *testing.T) {
	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	tests := []struct {
		name           string
		method         string
		path           string
		headers        map[string]string
		expectedStatus int
		expectAuth     bool
		expectAudit    bool
	}{
		{
			name:           "admin services endpoint should require authentication",
			method:         "GET",
			path:           "/admin/api/v1/services",
			headers:        map[string]string{},
			expectedStatus: http.StatusUnauthorized,
			expectAuth:     true,
			expectAudit:    true,
		},
		{
			name:           "authenticated admin should access services endpoint",
			method:         "GET",
			path:           "/admin/api/v1/services",
			headers:        map[string]string{"Authorization": "Bearer valid_admin_token"},
			expectedStatus: http.StatusOK,
			expectAuth:     true,
			expectAudit:    true,
		},
		{
			name:           "admin news management should require authentication",
			method:         "POST",
			path:           "/admin/api/v1/news",
			headers:        map[string]string{"Content-Type": "application/json"},
			expectedStatus: http.StatusUnauthorized,
			expectAuth:     true,
			expectAudit:    true,
		},
		{
			name:           "authenticated editor should access news endpoint",
			method:         "POST",
			path:           "/admin/api/v1/news",
			headers:        map[string]string{
				"Authorization": "Bearer valid_editor_token",
				"Content-Type":  "application/json",
			},
			expectedStatus: http.StatusCreated,
			expectAuth:     true,
			expectAudit:    true,
		},
		{
			name:           "insufficient permissions should be denied",
			method:         "DELETE",
			path:           "/admin/api/v1/services/123",
			headers:        map[string]string{"Authorization": "Bearer valid_viewer_token"},
			expectedStatus: http.StatusForbidden,
			expectAuth:     true,
			expectAudit:    true,
		},
		{
			name:           "expired token should be rejected",
			method:         "GET",
			path:           "/admin/api/v1/services",
			headers:        map[string]string{"Authorization": "Bearer expired_admin_token"},
			expectedStatus: http.StatusUnauthorized,
			expectAuth:     true,
			expectAudit:    true,
		},
		{
			name:           "health check should not require authentication",
			method:         "GET",
			path:           "/health",
			headers:        map[string]string{},
			expectedStatus: http.StatusOK,
			expectAuth:     false,
			expectAudit:    false,
		},
		{
			name:           "public routes should be blocked on admin gateway",
			method:         "GET",
			path:           "/api/v1/services",
			headers:        map[string]string{},
			expectedStatus: http.StatusNotFound,
			expectAuth:     false,
			expectAudit:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Contract: Admin gateway must be creatable with proper configuration
			adminGateway, err := gateway.NewAdminGateway(ctx, "development")
			require.NoError(t, err, "Admin gateway creation should succeed")
			require.NotNil(t, adminGateway, "Admin gateway should not be nil")

			// Contract: Gateway should handle HTTP requests according to routing and auth rules
			req := httptest.NewRequest(tt.method, tt.path, nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			recorder := httptest.NewRecorder()
			adminGateway.ServeHTTP(recorder, req)

			// Contract: Response status should match expected behavior
			assert.Equal(t, tt.expectedStatus, recorder.Code, "Response status should match expected")

			// Contract: Authentication challenges should be provided when required
			if tt.expectAuth && recorder.Code == http.StatusUnauthorized {
				assert.NotEmpty(t, recorder.Header().Get("WWW-Authenticate"), "WWW-Authenticate header should be set for unauthorized requests")
			}

			// Contract: Security headers should always be present
			assert.Equal(t, "nosniff", recorder.Header().Get("X-Content-Type-Options"), "X-Content-Type-Options should be set")
			assert.Equal(t, "DENY", recorder.Header().Get("X-Frame-Options"), "X-Frame-Options should be set")
			assert.Equal(t, "1; mode=block", recorder.Header().Get("X-XSS-Protection"), "X-XSS-Protection should be set")

			// Contract: Audit logging should occur for admin operations
			if tt.expectAudit {
				// Verify correlation ID is set for audit trail
				correlationID := recorder.Header().Get("X-Correlation-ID")
				assert.NotEmpty(t, correlationID, "Correlation ID should be set for audit trail")
			}
		})
	}
}

func TestRoleBasedAccessControl_Contract(t *testing.T) {
	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	tests := []struct {
		name           string
		token          string
		method         string
		path           string
		expectedStatus int
		expectedRole   string
	}{
		{
			name:           "admin should access all admin endpoints",
			token:          "valid_admin_token",
			method:         "DELETE",
			path:           "/admin/api/v1/services/123",
			expectedStatus: http.StatusOK,
			expectedRole:   "admin",
		},
		{
			name:           "content_editor should access content endpoints",
			token:          "valid_editor_token",
			method:         "POST",
			path:           "/admin/api/v1/news",
			expectedStatus: http.StatusCreated,
			expectedRole:   "content_editor",
		},
		{
			name:           "content_editor should not access admin-only endpoints",
			token:          "valid_editor_token",
			method:         "DELETE",
			path:           "/admin/api/v1/services/123",
			expectedStatus: http.StatusForbidden,
			expectedRole:   "content_editor",
		},
		{
			name:           "viewer should access read-only endpoints",
			token:          "valid_viewer_token",
			method:         "GET",
			path:           "/admin/api/v1/services",
			expectedStatus: http.StatusOK,
			expectedRole:   "viewer",
		},
		{
			name:           "viewer should not access write endpoints",
			token:          "valid_viewer_token",
			method:         "POST",
			path:           "/admin/api/v1/news",
			expectedStatus: http.StatusForbidden,
			expectedRole:   "viewer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Contract: RBAC should work with DAPR OPA policy integration
			adminGateway, err := gateway.NewAdminGateway(ctx, "development")
			require.NoError(t, err)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Authorization", "Bearer "+tt.token)

			recorder := httptest.NewRecorder()
			adminGateway.ServeHTTP(recorder, req)

			// Contract: Role-based access should be enforced
			assert.Equal(t, tt.expectedStatus, recorder.Code, "RBAC should enforce expected status")

			// Contract: User role should be logged for successful requests
			if recorder.Code == http.StatusOK || recorder.Code == http.StatusCreated {
				userRole := recorder.Header().Get("X-User-Role")
				assert.Equal(t, tt.expectedRole, userRole, "User role should be set in response headers")
			}
		})
	}
}

func TestRateLimitingIntegration_Contract(t *testing.T) {
	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	tests := []struct {
		name           string
		gatewayType    string
		requestCount   int
		expectedLimit  bool
		limitType      string
	}{
		{
			name:          "public gateway should enforce IP-based rate limiting",
			gatewayType:   "public",
			requestCount:  1001, // Exceeds 1000/min limit
			expectedLimit: true,
			limitType:     "ip",
		},
		{
			name:          "admin gateway should enforce user-based rate limiting",
			gatewayType:   "admin",
			requestCount:  101, // Exceeds 100/min limit
			expectedLimit: true,
			limitType:     "user",
		},
		{
			name:          "public gateway should allow requests under limit",
			gatewayType:   "public",
			requestCount:  10, // Under 1000/min limit
			expectedLimit: false,
			limitType:     "ip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gatewayHandler http.Handler
			var err error

			// Contract: Rate limiting should be configured per gateway type
			if tt.gatewayType == "public" {
				gatewayHandler, err = gateway.NewPublicGateway(ctx, "development")
			} else {
				gatewayHandler, err = gateway.NewAdminGateway(ctx, "development")
			}
			require.NoError(t, err, "Gateway creation should succeed")

			// Simulate multiple requests to trigger rate limiting
			var lastStatus int
			for i := 0; i < tt.requestCount; i++ {
				req := httptest.NewRequest("GET", "/health", nil)
				if tt.gatewayType == "admin" {
					req.Header.Set("Authorization", "Bearer valid_admin_token")
				}

				recorder := httptest.NewRecorder()
				gatewayHandler.ServeHTTP(recorder, req)
				lastStatus = recorder.Code

				// Break early if rate limited
				if recorder.Code == http.StatusTooManyRequests {
					break
				}
			}

			// Contract: Rate limiting should be enforced according to configuration
			if tt.expectedLimit {
				assert.Equal(t, http.StatusTooManyRequests, lastStatus, "Rate limiting should be enforced")
			} else {
				assert.NotEqual(t, http.StatusTooManyRequests, lastStatus, "Requests under limit should be allowed")
			}
		})
	}
}

func TestSecurityHeadersIntegration_Contract(t *testing.T) {
	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	gatewayTypes := []struct {
		name    string
		gateway func(context.Context, string) (http.Handler, error)
	}{
		{"public", func(ctx context.Context, env string) (http.Handler, error) { return gateway.NewPublicGateway(ctx, env) }},
		{"admin", func(ctx context.Context, env string) (http.Handler, error) { return gateway.NewAdminGateway(ctx, env) }},
	}

	for _, gt := range gatewayTypes {
		t.Run(gt.name+"_gateway_security_headers", func(t *testing.T) {
			// Contract: Security headers should be consistently applied
			gatewayHandler, err := gt.gateway(ctx, "development")
			require.NoError(t, err)

			req := httptest.NewRequest("GET", "/health", nil)
			recorder := httptest.NewRecorder()
			gatewayHandler.ServeHTTP(recorder, req)

			// Contract: Essential security headers must be present
			expectedHeaders := map[string]string{
				"X-Content-Type-Options":           "nosniff",
				"X-Frame-Options":                  "DENY",
				"X-XSS-Protection":                 "1; mode=block",
				"Strict-Transport-Security":        "max-age=31536000; includeSubDomains",
				"Content-Security-Policy":          "default-src 'self'; object-src 'none'",
				"Referrer-Policy":                  "strict-origin-when-cross-origin",
			}

			for header, expectedValue := range expectedHeaders {
				actualValue := recorder.Header().Get(header)
				assert.Equal(t, expectedValue, actualValue, "Security header %s should be set correctly", header)
			}

			// Contract: Gateway should set correlation ID for tracing
			correlationID := recorder.Header().Get("X-Correlation-ID")
			assert.NotEmpty(t, correlationID, "Correlation ID should be set for request tracing")
		})
	}
}

func TestDAPRMiddlewareIntegration_Contract(t *testing.T) {
	timeout := 5 * time.Second
	_, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	tests := []struct {
		name            string
		environment     string
		expectedDAPR    bool
		expectedVault   string
		expectedRedis   string
	}{
		{
			name:          "development environment should use local DAPR services",
			environment:   "development",
			expectedDAPR:  true,
			expectedVault: "http://vault-dev:8200",
			expectedRedis: "redis-dev:6379",
		},
		{
			name:          "staging environment should use hosted DAPR services",
			environment:   "staging",
			expectedDAPR:  true,
			expectedVault: "https://vault-staging.axiomcloud.dev",
			expectedRedis: "redis-staging.upstash.io:6379",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Contract: DAPR integration should be environment-specific
			config, err := gateway.NewGatewayDAPRConfiguration(tt.environment)
			require.NoError(t, err, "DAPR configuration should be creatable")
			require.NotNil(t, config, "DAPR configuration should not be nil")

			// Contract: DAPR middleware chain should be properly configured
			assert.Equal(t, tt.expectedDAPR, config.DAPREnabled, "DAPR should be enabled for environment")
			assert.Equal(t, tt.expectedVault, config.VaultEndpoint, "Vault endpoint should match environment")
			assert.Equal(t, tt.expectedRedis, config.RedisEndpoint, "Redis endpoint should match environment")

			// Contract: DAPR middleware components should be configured
			middlewareChain := config.GetMiddlewareChain()
			require.NotEmpty(t, middlewareChain, "Middleware chain should not be empty")

			// Verify expected middleware components
			expectedMiddlewares := []string{"bearer", "oauth2", "opa", "ratelimit", "cors"}
			for _, expectedMW := range expectedMiddlewares {
				found := false
				for _, actualMW := range middlewareChain {
					if actualMW.Name == expectedMW {
						found = true
						break
					}
				}
				assert.True(t, found, "Middleware %s should be present in chain", expectedMW)
			}
		})
	}
}