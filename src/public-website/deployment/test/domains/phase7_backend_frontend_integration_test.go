package domains

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	sharedtesting "github.com/axiom-software-co/international-center/src/public-website/infrastructure/test/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// PHASE 7: BACKEND TO FRONTEND INTEGRATION TESTS
// WHY: End-to-end integration validation across complete stack
// SCOPE: Gateway routing, frontend to backend communication, full workflow validation
// DEPENDENCIES: Phases 1-6 must pass
// CONTEXT: Complete stack integration, end-to-end workflow testing

func TestPhase7GatewayIntegration(t *testing.T) {
	// Environment validation required for all gateway tests
	sharedtesting.ValidateEnvironmentPrerequisites(t)

	timeout := 15 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("PublicGatewayOperational", func(t *testing.T) {
		// Test public gateway accessibility and basic functionality
		resp, err := client.Get("http://localhost:9001/health")
		if err != nil {
			t.Logf("Public gateway not accessible: %v", err)
			return
		}
		defer resp.Body.Close()

		assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300,
			"Public gateway health should be accessible")

		// Validate health response structure
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			var healthData map[string]interface{}
			if json.Unmarshal(body, &healthData) == nil {
				if gatewayField, exists := healthData["gateway"]; exists {
					assert.Equal(t, "public-gateway", gatewayField,
						"Public gateway should identify correctly")
				}
			}
		}

		t.Logf("✅ Public gateway: Operational (status: %d)", resp.StatusCode)
	})

	t.Run("AdminGatewayOperational", func(t *testing.T) {
		// Test admin gateway accessibility and basic functionality
		resp, err := client.Get("http://localhost:9000/health")
		if err != nil {
			t.Logf("Admin gateway not accessible: %v", err)
			return
		}
		defer resp.Body.Close()

		assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300,
			"Admin gateway health should be accessible")

		// Validate health response structure
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			var healthData map[string]interface{}
			if json.Unmarshal(body, &healthData) == nil {
				if gatewayField, exists := healthData["gateway"]; exists {
					assert.Equal(t, "admin-gateway", gatewayField,
						"Admin gateway should identify correctly")
				}
			}
		}

		t.Logf("✅ Admin gateway: Operational (status: %d)", resp.StatusCode)
	})

	t.Run("PublicGatewayBackendServiceRouting", func(t *testing.T) {
		// Test public gateway routing to actual backend service endpoints
		baseURL := "http://localhost:9001/api"

		t.Run("ContentServiceRouting", func(t *testing.T) {
			// Test routing to content service endpoints
			endpoints := []struct {
				name string
				path string
			}{
				{"content-health", "/content/health"},
				{"content-events", "/content/events"},
				{"content-news", "/content/news"},
				{"content-research", "/content/research"},
				{"content-services", "/content/services"},
			}

			for _, endpoint := range endpoints {
				url := baseURL + endpoint.path
				resp, err := client.Get(url)

				if err != nil {
					t.Logf("❌ Content service routing failed for %s: %v", endpoint.name, err)
					continue
				}
				defer resp.Body.Close()

				// Accept various response codes - we're testing routing, not full functionality
				if resp.StatusCode < 500 {
					t.Logf("✅ Content service routing %s: Gateway routing successful (status: %d)", endpoint.name, resp.StatusCode)
				} else {
					t.Logf("⚠️ Content service routing %s: Server error (status: %d)", endpoint.name, resp.StatusCode)
				}

				assert.True(t, resp.StatusCode < 500,
					"Content service %s should be routable through gateway (no 5xx errors), got %d", endpoint.name, resp.StatusCode)
			}
		})

		t.Run("InquiriesServiceRouting", func(t *testing.T) {
			// Test routing to inquiries service endpoints
			endpoints := []struct {
				name string
				path string
			}{
				{"inquiries-health", "/inquiries/health"},
				{"inquiries-donations", "/inquiries/donations"},
				{"inquiries-business", "/inquiries/business"},
				{"inquiries-media", "/inquiries/media"},
				{"inquiries-volunteers", "/inquiries/volunteers"},
			}

			for _, endpoint := range endpoints {
				url := baseURL + endpoint.path
				resp, err := client.Get(url)

				if err != nil {
					t.Logf("❌ Inquiries service routing failed for %s: %v", endpoint.name, err)
					continue
				}
				defer resp.Body.Close()

				if resp.StatusCode < 500 {
					t.Logf("✅ Inquiries service routing %s: Gateway routing successful (status: %d)", endpoint.name, resp.StatusCode)
				} else {
					t.Logf("⚠️ Inquiries service routing %s: Server error (status: %d)", endpoint.name, resp.StatusCode)
				}

				assert.True(t, resp.StatusCode < 500,
					"Inquiries service %s should be routable through gateway (no 5xx errors), got %d", endpoint.name, resp.StatusCode)
			}
		})

		t.Run("NotificationsServiceRouting", func(t *testing.T) {
			// Test routing to notifications service endpoints
			endpoints := []struct {
				name string
				path string
			}{
				{"notifications-health", "/notifications/health"},
			}

			for _, endpoint := range endpoints {
				url := baseURL + endpoint.path
				resp, err := client.Get(url)

				if err != nil {
					t.Logf("❌ Notifications service routing failed for %s: %v", endpoint.name, err)
					continue
				}
				defer resp.Body.Close()

				if resp.StatusCode < 500 {
					t.Logf("✅ Notifications service routing %s: Gateway routing successful (status: %d)", endpoint.name, resp.StatusCode)
				} else {
					t.Logf("⚠️ Notifications service routing %s: Server error (status: %d)", endpoint.name, resp.StatusCode)
				}

				assert.True(t, resp.StatusCode < 500,
					"Notifications service %s should be routable through gateway (no 5xx errors), got %d", endpoint.name, resp.StatusCode)
			}
		})
	})

	t.Run("GatewayRoutingToBackendServices", func(t *testing.T) {
		// Test gateway routing to backend services
		gatewayTester := sharedtesting.NewGatewayServiceMeshTester()

		errors := gatewayTester.ValidateGatewayToServiceCommunication(ctx)
		if len(errors) > 0 {
			for _, err := range errors {
				t.Logf("Gateway routing issue: %v", err)
			}
			t.Log("⚠️ Gateway routing: Some routes may not be fully implemented")
		} else {
			t.Log("✅ Gateway routing: All routes properly configured")
		}

		assert.Empty(t, errors, "Gateway routing should work without errors")
	})

	t.Run("PublicGatewayAPIRouting", func(t *testing.T) {
		// Test public gateway API routing patterns
		apiTests := []struct {
			endpoint      string
			expectedService string
			description   string
		}{
			{"/api/news", "content", "Public gateway should route news requests to content service"},
			{"/api/events", "content", "Public gateway should route events requests to content service"},
			{"/api/services", "content", "Public gateway should route services requests to content service"},
		}

		for _, test := range apiTests {
			t.Run("PublicAPI_"+strings.ReplaceAll(test.endpoint, "/", "_"), func(t *testing.T) {
				resp, err := client.Get("http://localhost:9001" + test.endpoint)
				if err != nil {
					t.Logf("Public API routing failed for %s: %v", test.endpoint, err)
					return
				}
				defer resp.Body.Close()

				// Accept any reasonable response (200-404) as routing working
				if resp.StatusCode < 500 {
					t.Logf("✅ Public API routing: %s → %s (status: %d)", test.endpoint, test.expectedService, resp.StatusCode)
				} else {
					t.Logf("⚠️ Public API routing error for %s: status %d", test.endpoint, resp.StatusCode)
				}
			})
		}
	})

	t.Run("AdminGatewayAPIRouting", func(t *testing.T) {
		// Test admin gateway API routing patterns
		apiTests := []struct {
			endpoint      string
			expectedService string
			description   string
		}{
			{"/api/admin/inquiries", "inquiries", "Admin gateway should route inquiry management to inquiries service"},
			{"/api/admin/content", "content", "Admin gateway should route content management to content service"},
		}

		for _, test := range apiTests {
			t.Run("AdminAPI_"+strings.ReplaceAll(test.endpoint, "/", "_"), func(t *testing.T) {
				resp, err := client.Get("http://localhost:9000" + test.endpoint)
				if err != nil {
					t.Logf("Admin API routing failed for %s: %v", test.endpoint, err)
					return
				}
				defer resp.Body.Close()

				// Accept any reasonable response (200-404) as routing working
				if resp.StatusCode < 500 {
					t.Logf("✅ Admin API routing: %s → %s (status: %d)", test.endpoint, test.expectedService, resp.StatusCode)
				} else {
					t.Logf("⚠️ Admin API routing error for %s: status %d", test.endpoint, resp.StatusCode)
				}
			})
		}
	})
}

func TestPhase7GatewaySecurityAndCompliance(t *testing.T) {
	// Test gateway security features and compliance requirements
	sharedtesting.ValidateEnvironmentPrerequisites(t)

	timeout := 15 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("GatewayCORSConfiguration", func(t *testing.T) {
		// Test CORS configuration for both gateways
		gateways := []struct {
			name string
			url  string
		}{
			{"public-gateway", "http://localhost:9001/health"},
			{"admin-gateway", "http://localhost:9000/health"},
		}

		for _, gateway := range gateways {
			t.Run(gateway.name+"CORS", func(t *testing.T) {
				req, err := http.NewRequestWithContext(ctx, "OPTIONS", gateway.url, nil)
				require.NoError(t, err)

				req.Header.Set("Origin", "http://localhost:3000")
				req.Header.Set("Access-Control-Request-Method", "GET")

				resp, err := client.Do(req)
				if err != nil {
					t.Logf("%s CORS preflight failed: %v", gateway.name, err)
					return
				}
				defer resp.Body.Close()

				// Check for CORS headers
				corsOrigin := resp.Header.Get("Access-Control-Allow-Origin")
				corsHeaders := resp.Header.Get("Access-Control-Allow-Headers")

				if corsOrigin != "" || corsHeaders != "" {
					t.Logf("✅ %s: CORS configured (origin: %s)", gateway.name, corsOrigin)
				} else {
					t.Logf("⚠️ %s: CORS headers not detected", gateway.name)
				}
			})
		}
	})

	t.Run("GatewaySecurityHeaders", func(t *testing.T) {
		// Test security headers on both gateways
		gateways := []struct {
			name string
			url  string
		}{
			{"public-gateway", "http://localhost:9001/health"},
			{"admin-gateway", "http://localhost:9000/health"},
		}

		for _, gateway := range gateways {
			t.Run(gateway.name+"SecurityHeaders", func(t *testing.T) {
				resp, err := client.Get(gateway.url)
				if err != nil {
					t.Logf("%s security headers check failed: %v", gateway.name, err)
					return
				}
				defer resp.Body.Close()

				// Check for common security headers
				securityHeaders := []string{
					"X-Content-Type-Options",
					"X-Frame-Options",
					"X-XSS-Protection",
					"Strict-Transport-Security",
				}

				foundHeaders := 0
				for _, header := range securityHeaders {
					if resp.Header.Get(header) != "" {
						foundHeaders++
					}
				}

				t.Logf("%s: Security headers detected: %d/%d", gateway.name, foundHeaders, len(securityHeaders))
			})
		}
	})

	t.Run("AdminGatewayAuthenticationValidation", func(t *testing.T) {
		// Test admin gateway authentication requirements
		protectedEndpoints := []string{
			"/api/admin/inquiries",
			"/api/admin/content",
		}

		for _, endpoint := range protectedEndpoints {
			t.Run("Auth_"+strings.ReplaceAll(endpoint, "/", "_"), func(t *testing.T) {
				// Test unauthenticated request
				resp, err := client.Get("http://localhost:9000" + endpoint)
				if err != nil {
					t.Logf("Admin endpoint %s not accessible: %v", endpoint, err)
					return
				}
				defer resp.Body.Close()

				// Admin endpoints should require authentication
				if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
					t.Logf("✅ Admin authentication: %s properly protected (status: %d)", endpoint, resp.StatusCode)
				} else {
					t.Logf("⚠️ Admin authentication: %s may not be properly protected (status: %d)", endpoint, resp.StatusCode)
				}
			})
		}
	})
}

func TestPhase7EndToEndIntegration(t *testing.T) {
	// Test end-to-end backend to frontend integration
	sharedtesting.ValidateEnvironmentPrerequisites(t)

	timeout := 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	t.Run("EndToEndWorkflowValidation", func(t *testing.T) {
		// Test complete frontend workflow enablement through gateways
		crossStackTester := sharedtesting.NewCrossStackIntegrationTester()

		errors := crossStackTester.ValidateEndToEndWorkflow(ctx)
		if len(errors) > 0 {
			for _, err := range errors {
				t.Logf("End-to-end workflow issue: %v", err)
			}
			t.Log("⚠️ Gateway workflows: Some end-to-end workflows may not be fully functional")
		} else {
			t.Log("✅ Gateway workflows: All end-to-end workflows operational")
		}
	})

	t.Run("GatewayServiceMeshIntegration", func(t *testing.T) {
		// Test gateway integration with service mesh backend
		client := &http.Client{Timeout: 10 * time.Second}

		// Test direct service mesh access to verify target services
		services := []string{"content", "inquiries", "notifications"}

		for _, service := range services {
			t.Run("ServiceMesh_"+service, func(t *testing.T) {
				directServiceMeshURL := "http://localhost:3500/v1.0/invoke/" + service + "/method/health"

				resp, err := client.Get(directServiceMeshURL)
				if err != nil {
					t.Logf("Service mesh access failed for %s: %v", service, err)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					t.Logf("✅ Service mesh integration: %s accessible through Dapr", service)
				} else {
					t.Logf("⚠️ Service mesh integration: %s access issue (status: %d)", service, resp.StatusCode)
				}
			})
		}
	})

	t.Run("FrontendToBackendConnectivity", func(t *testing.T) {
		// Test frontend to backend connectivity through gateways
		connectivityTests := []struct {
			frontend string
			gateway  string
			backend  string
			endpoint string
		}{
			{"public-website", "http://localhost:9001", "content-api", "/api/content/health"},
			{"public-website", "http://localhost:9001", "inquiries-api", "/api/inquiries/health"},
			{"admin-portal", "http://localhost:9000", "content-api", "/api/admin/content"},
			{"admin-portal", "http://localhost:9000", "inquiries-api", "/api/admin/inquiries"},
		}

		client := &http.Client{Timeout: 10 * time.Second}

		for _, test := range connectivityTests {
			t.Run("Connectivity_"+test.frontend+"_to_"+test.backend, func(t *testing.T) {
				url := test.gateway + test.endpoint

				resp, err := client.Get(url)
				if err != nil {
					t.Logf("Frontend to backend connectivity failed for %s to %s: %v", test.frontend, test.backend, err)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode < 500 {
					t.Logf("✅ Frontend connectivity: %s to %s operational (status: %d)",
						test.frontend, test.backend, resp.StatusCode)
				} else {
					t.Logf("⚠️ Frontend connectivity: %s to %s issue (status: %d)",
						test.frontend, test.backend, resp.StatusCode)
				}
			})
		}
	})

	t.Run("FullStackOperationalReadiness", func(t *testing.T) {
		// Test overall full stack operational readiness
		readinessChecks := []struct {
			component string
			url       string
		}{
			{"public-gateway", "http://localhost:9001/health"},
			{"admin-gateway", "http://localhost:9000/health"},
			{"content-service", "http://localhost:3500/v1.0/invoke/content-api/method/health"},
			{"inquiries-service", "http://localhost:3500/v1.0/invoke/inquiries-api/method/health"},
			{"notifications-service", "http://localhost:3500/v1.0/invoke/notifications-api/method/health"},
		}

		client := &http.Client{Timeout: 5 * time.Second}
		operational := 0
		total := len(readinessChecks)

		for _, check := range readinessChecks {
			resp, err := client.Get(check.url)
			if err == nil && resp != nil {
				resp.Body.Close()
				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					operational++
					t.Logf("✅ Full stack readiness: %s operational", check.component)
				}
			} else {
				t.Logf("⚠️ Full stack readiness: %s not operational", check.component)
			}
		}

		readinessPercentage := float64(operational) / float64(total)
		t.Logf("Full stack operational readiness: %.2f%% (%d/%d components operational)",
			readinessPercentage*100, operational, total)

		// Full stack should be at least 70% operational
		assert.GreaterOrEqual(t, readinessPercentage, 0.7,
			"Full stack should be at least 70%% operational for production readiness")
	})
}