package integration

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED PHASE: Complete Admin Gateway Tests
// These tests validate that admin gateway routes all admin API requests properly

func TestCompleteAdminGateway_AllAdminAPIRoutes(t *testing.T) {
	// This test requires complete environment health - enforcing axiom rule
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Complete admin API routes that must work for admin portal functionality
	adminAPIRouteTests := []struct {
		routeName        string
		adminEndpoint    string
		backendService   string
		backendEndpoint  string
		currentStatus    string
		description      string
		critical         bool
	}{
		{
			routeName:       "admin-inquiries",
			adminEndpoint:   "http://localhost:9000/api/admin/inquiries",
			backendService:  "inquiries",
			backendEndpoint: "http://localhost:3101/api/inquiries",
			currentStatus:   "GATEWAY_ERROR",
			description:     "Admin gateway must route inquiries API for admin portal inquiry management",
			critical:        true,
		},
		{
			routeName:       "admin-subscribers",
			adminEndpoint:   "http://localhost:9000/api/admin/subscribers",
			backendService:  "notifications",
			backendEndpoint: "http://localhost:3201/api/subscribers",
			currentStatus:   "GATEWAY_ERROR",
			description:     "Admin gateway must route subscribers API for admin portal subscriber management",
			critical:        true,
		},
		{
			routeName:       "admin-content",
			adminEndpoint:   "http://localhost:9000/api/admin/content",
			backendService:  "content",
			backendEndpoint: "http://localhost:3001/api/news",
			currentStatus:   "unknown",
			description:     "Admin gateway must route content API for admin portal content management",
			critical:        false,
		},
	}

	client := &http.Client{Timeout: 15 * time.Second}

	// Act & Assert: Test all admin API routes
	for _, route := range adminAPIRouteTests {
		t.Run("AdminAPIRoute_"+route.routeName, func(t *testing.T) {
			// First verify backend service is accessible
			backendReq, err := http.NewRequestWithContext(ctx, "GET", route.backendEndpoint, nil)
			require.NoError(t, err, "Failed to create backend request")

			backendResp, err := client.Do(backendReq)
			require.NoError(t, err, "Backend service %s must be accessible", route.backendService)
			defer backendResp.Body.Close()

			assert.True(t, backendResp.StatusCode >= 200 && backendResp.StatusCode < 300,
				"Backend service %s must be functional for admin routing", route.backendService)

			// Test admin gateway routing
			adminReq, err := http.NewRequestWithContext(ctx, "GET", route.adminEndpoint, nil)
			require.NoError(t, err, "Failed to create admin gateway request")

			adminResp, err := client.Do(adminReq)
			if err == nil {
				defer adminResp.Body.Close()

				body, err := io.ReadAll(adminResp.Body)
				if err == nil {
					responseStr := string(body)

					if route.critical {
						// Critical admin routes must not return GATEWAY_ERROR
						assert.NotContains(t, responseStr, "GATEWAY_ERROR",
							"%s - admin gateway must not return processing errors", route.description)
						assert.NotContains(t, responseStr, "Gateway processing error occurred",
							"%s - admin routing must be functional", route.description)
						
						// Should return backend service response or proper proxy response
						if adminResp.StatusCode >= 200 && adminResp.StatusCode < 300 {
							var jsonData interface{}
							assert.NoError(t, json.Unmarshal(body, &jsonData),
								"%s - admin gateway must return valid JSON from backend", route.description)
						}

						// Should not return generic errors
						assert.True(t, adminResp.StatusCode >= 200 && adminResp.StatusCode < 500,
							"%s - admin API route must be accessible", route.description)
					} else {
						// Non-critical routes should work for complete admin functionality
						if strings.Contains(responseStr, "GATEWAY_ERROR") {
							t.Logf("%s returning gateway error (expected for incomplete admin routing)", route.description)
						}
					}
				}
			} else {
				if route.critical {
					t.Errorf("%s not accessible: %v", route.description, err)
				}
			}
		})
	}
}

func TestCompleteAdminGateway_AdminPortalRequiredFunctionality(t *testing.T) {
	// Test admin gateway functionality required for admin portal development
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Admin portal functionality requirements
	adminPortalRequirements := []struct {
		functionality   string
		testEndpoint    string
		requiredFeature string
		description     string
	}{
		{
			functionality:   "inquiry-management",
			testEndpoint:    "http://localhost:9000/api/admin/inquiries",
			requiredFeature: "CRUD operations for inquiries",
			description:     "Admin portal must manage inquiries through admin gateway",
		},
		{
			functionality:   "subscriber-management",
			testEndpoint:    "http://localhost:9000/api/admin/subscribers",
			requiredFeature: "CRUD operations for subscribers",
			description:     "Admin portal must manage subscribers through admin gateway",
		},
		{
			functionality:   "content-management",
			testEndpoint:    "http://localhost:9000/health",
			requiredFeature: "Access to content services",
			description:     "Admin portal must access content management functionality",
		},
		{
			functionality:   "notification-management",
			testEndpoint:    "http://localhost:9000/health",
			requiredFeature: "Access to notification services",
			description:     "Admin portal must access notification management functionality",
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Act & Assert: Test admin portal functionality requirements
	for _, requirement := range adminPortalRequirements {
		t.Run("AdminPortalFunc_"+requirement.functionality, func(t *testing.T) {
			req, err := http.NewRequestWithContext(ctx, "GET", requirement.testEndpoint, nil)
			require.NoError(t, err, "Failed to create admin portal functionality request")

			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				
				// Admin functionality must be accessible
				assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500,
					"%s - admin functionality must be accessible", requirement.description)

				body, err := io.ReadAll(resp.Body)
				if err == nil {
					responseStr := string(body)
					
					// Must not return generic gateway errors for admin functionality
					assert.NotContains(t, responseStr, "GATEWAY_ERROR",
						"%s - admin functionality must not return generic gateway errors", requirement.description)
					
					// Should return structured response for admin portal consumption
					var jsonData interface{}
					assert.NoError(t, json.Unmarshal(body, &jsonData),
						"%s - must return JSON for admin portal %s", requirement.description, requirement.requiredFeature)
				}
			} else {
				t.Errorf("%s not accessible for %s: %v", requirement.description, requirement.requiredFeature, err)
			}
		})
	}
}

func TestCompleteAdminGateway_ServiceInvocationConsistency(t *testing.T) {
	// Test consistency between public gateway (working) and admin gateway (errors)
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Comparison tests between working public gateway and admin gateway issues
	gatewayConsistencyTests := []struct {
		testName         string
		publicEndpoint   string
		adminEndpoint    string
		targetService    string
		publicStatus     string
		adminStatus      string
		description      string
	}{
		{
			testName:       "content-api-access",
			publicEndpoint: "http://localhost:9001/api/events",
			adminEndpoint:  "http://localhost:9000/api/admin/content",
			targetService:  "content",
			publicStatus:   "working",
			adminStatus:    "unknown",
			description:    "Content API access should work consistently through both gateways",
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Act & Assert: Test gateway consistency
	for _, consistency := range gatewayConsistencyTests {
		t.Run("GatewayConsistency_"+consistency.testName, func(t *testing.T) {
			// Test public gateway (should work)
			publicReq, err := http.NewRequestWithContext(ctx, "GET", consistency.publicEndpoint, nil)
			require.NoError(t, err, "Failed to create public gateway request")

			publicResp, err := client.Do(publicReq)
			require.NoError(t, err, "Public gateway must be accessible")
			defer publicResp.Body.Close()

			assert.True(t, publicResp.StatusCode >= 200 && publicResp.StatusCode < 300,
				"Public gateway %s access must work", consistency.targetService)

			// Test admin gateway (currently failing)
			adminReq, err := http.NewRequestWithContext(ctx, "GET", consistency.adminEndpoint, nil)
			require.NoError(t, err, "Failed to create admin gateway request")

			adminResp, err := client.Do(adminReq)
			if err == nil {
				defer adminResp.Body.Close()
				
				body, err := io.ReadAll(adminResp.Body)
				if err == nil {
					responseStr := string(body)
					
					// Admin gateway should work consistently with public gateway
					assert.NotContains(t, responseStr, "GATEWAY_ERROR",
						"%s - admin gateway should work consistently with public gateway", consistency.description)
					
					// Should return similar response structure as public gateway
					if adminResp.StatusCode >= 200 && adminResp.StatusCode < 300 {
						var jsonData interface{}
						assert.NoError(t, json.Unmarshal(body, &jsonData),
							"%s - admin gateway should return JSON like public gateway", consistency.description)
					}
				}
			} else {
				t.Errorf("%s admin gateway not accessible: %v", consistency.description, err)
			}
		})
	}
}

func TestCompleteAdminGateway_BackendServiceAccessibility(t *testing.T) {
	// Test that admin gateway can access all required backend services
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Backend services that admin gateway must access
	adminBackendAccessTests := []struct {
		serviceName    string
		serviceHealth  string
		serviceAPI     string
		description    string
	}{
		{
			serviceName:   "inquiries",
			serviceHealth: "http://localhost:3500/v1.0/invoke/inquiries/method/health",
			serviceAPI:    "http://localhost:3500/v1.0/invoke/inquiries/method/api/inquiries",
			description:   "Admin gateway must access inquiries service for inquiry management",
		},
		{
			serviceName:   "notifications",
			serviceHealth: "http://localhost:3500/v1.0/invoke/notifications/method/health",
			serviceAPI:    "http://localhost:3500/v1.0/invoke/notifications/method/api/subscribers",
			description:   "Admin gateway must access notifications service for subscriber management",
		},
		{
			serviceName:   "content",
			serviceHealth: "http://localhost:3500/v1.0/invoke/content/method/health",
			serviceAPI:    "http://localhost:3500/v1.0/invoke/content/method/api/news",
			description:   "Admin gateway must access content service for content management",
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Act & Assert: Test backend service accessibility
	for _, access := range adminBackendAccessTests {
		t.Run("BackendAccess_"+access.serviceName, func(t *testing.T) {
			// Test service health (should work)
			healthReq, err := http.NewRequestWithContext(ctx, "GET", access.serviceHealth, nil)
			require.NoError(t, err, "Failed to create service health request")

			healthResp, err := client.Do(healthReq)
			require.NoError(t, err, "Service health must be accessible")
			defer healthResp.Body.Close()

			assert.True(t, healthResp.StatusCode >= 200 && healthResp.StatusCode < 300,
				"Service %s health must be accessible for admin gateway integration", access.serviceName)

			// Test service API (should work through service mesh)
			apiReq, err := http.NewRequestWithContext(ctx, "GET", access.serviceAPI, nil)
			require.NoError(t, err, "Failed to create service API request")

			apiResp, err := client.Do(apiReq)
			require.NoError(t, err, "Service API must be accessible")
			defer apiResp.Body.Close()

			assert.True(t, apiResp.StatusCode >= 200 && apiResp.StatusCode < 300,
				"%s - backend service API must be accessible for admin gateway", access.description)

			// Validate service returns structured data
			body, err := io.ReadAll(apiResp.Body)
			if err == nil {
				var jsonData interface{}
				assert.NoError(t, json.Unmarshal(body, &jsonData),
					"Backend service %s must return JSON for admin gateway integration", access.serviceName)
			}
		})
	}
}

func TestCompleteAdminGateway_AdminVsPublicGatewayComparison(t *testing.T) {
	// Compare admin gateway issues with working public gateway patterns
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Gateway comparison tests
	gatewayComparisonTests := []struct {
		comparisonName  string
		publicEndpoint  string
		adminEndpoint   string
		targetService   string
		description     string
	}{
		{
			comparisonName: "content-service-access",
			publicEndpoint: "http://localhost:9001/api/events",
			adminEndpoint:  "http://localhost:9000/api/admin/content",
			targetService:  "content",
			description:    "Content service access comparison between working public and failing admin gateway",
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Act & Assert: Compare gateway behaviors
	for _, comparison := range gatewayComparisonTests {
		t.Run("GatewayComparison_"+comparison.comparisonName, func(t *testing.T) {
			// Test public gateway (working pattern)
			publicReq, err := http.NewRequestWithContext(ctx, "GET", comparison.publicEndpoint, nil)
			require.NoError(t, err, "Failed to create public gateway request")

			publicResp, err := client.Do(publicReq)
			require.NoError(t, err, "Public gateway must be accessible")
			defer publicResp.Body.Close()

			publicBody, err := io.ReadAll(publicResp.Body)
			require.NoError(t, err, "Failed to read public gateway response")

			// Test admin gateway (problematic pattern)
			adminReq, err := http.NewRequestWithContext(ctx, "GET", comparison.adminEndpoint, nil)
			require.NoError(t, err, "Failed to create admin gateway request")

			adminResp, err := client.Do(adminReq)
			if err == nil {
				defer adminResp.Body.Close()
				
				adminBody, err := io.ReadAll(adminResp.Body)
				if err == nil {
					publicResponseStr := string(publicBody)
					adminResponseStr := string(adminBody)
					
					// Admin gateway should work like public gateway (not return gateway errors)
					assert.NotContains(t, adminResponseStr, "GATEWAY_ERROR",
						"%s - admin gateway should work like public gateway", comparison.description)
					
					// Both should access the same target service successfully
					t.Logf("Public gateway response: %s", publicResponseStr)
					t.Logf("Admin gateway response: %s", adminResponseStr)
					
					// Admin gateway should return service response (not processing error)
					if strings.Contains(adminResponseStr, "GATEWAY_ERROR") {
						t.Errorf("%s - admin gateway failing while public gateway works", comparison.description)
					}
				}
			} else {
				t.Errorf("%s admin gateway not accessible: %v", comparison.description, err)
			}
		})
	}
}

// validateEnvironmentPrerequisites ensures environment health before integration testing
func validateEnvironmentPrerequisites(t *testing.T) {
	// Check critical infrastructure, platform, service, and gateway components are running
	criticalContainers := []string{"postgresql", "dapr-control-plane", "content", "inquiries", "notifications", "public-gateway", "admin-gateway"}
	
	for _, container := range criticalContainers {
		cmd := exec.Command("podman", "ps", "--filter", "name="+container, "--format", "{{.Names}}")
		output, err := cmd.Output()
		require.NoError(t, err, "Failed to check critical container %s", container)

		if !strings.Contains(string(output), container) {
			t.Skipf("Critical container %s not running - environment not ready for integration testing", container)
		}
	}
}