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

// RED PHASE: Gateway Backend Integration Tests
// These tests validate that gateways route requests to backend service APIs

func TestGatewayBackendIntegration_APIRouting(t *testing.T) {
	// This test requires complete environment health - enforcing axiom rule
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Gateway API routing tests that should work (not return ROUTE_NOT_FOUND)
	gatewayRoutingTests := []struct {
		gatewayName     string
		gatewayEndpoint string
		backendService  string
		backendEndpoint string
		description     string
		critical        bool
	}{
		{
			gatewayName:     "public-gateway",
			gatewayEndpoint: "http://localhost:9001/api/news",
			backendService:  "content",
			backendEndpoint: "http://localhost:3001/api/news",
			description:     "Public gateway must route news API to content service",
			critical:        true,
		},
		{
			gatewayName:     "public-gateway", 
			gatewayEndpoint: "http://localhost:9001/api/events",
			backendService:  "content",
			backendEndpoint: "http://localhost:3001/api/events",
			description:     "Public gateway must route events API to content service",
			critical:        true,
		},
		{
			gatewayName:     "admin-gateway",
			gatewayEndpoint: "http://localhost:9000/api/admin/inquiries",
			backendService:  "inquiries",
			backendEndpoint: "http://localhost:3101/api/inquiries",
			description:     "Admin gateway must route inquiries API to inquiries service",
			critical:        true,
		},
		{
			gatewayName:     "admin-gateway",
			gatewayEndpoint: "http://localhost:9000/api/admin/subscribers",
			backendService:  "notifications",
			backendEndpoint: "http://localhost:3201/api/subscribers",
			description:     "Admin gateway must route subscribers API to notifications service",
			critical:        false,
		},
	}

	client := &http.Client{Timeout: 15 * time.Second}

	// Act & Assert: Test gateway routing to backend services
	for _, routing := range gatewayRoutingTests {
		t.Run("GatewayRouting_"+routing.gatewayName+"_to_"+routing.backendService, func(t *testing.T) {
			// First verify backend service endpoint is working
			backendReq, err := http.NewRequestWithContext(ctx, "GET", routing.backendEndpoint, nil)
			require.NoError(t, err, "Failed to create backend API request")

			backendResp, err := client.Do(backendReq)
			require.NoError(t, err, "Backend service %s must be accessible", routing.backendService)
			defer backendResp.Body.Close()

			assert.True(t, backendResp.StatusCode >= 200 && backendResp.StatusCode < 300,
				"Backend service %s API must be functional for gateway routing", routing.backendService)

			// Now test gateway routing to backend service
			gatewayReq, err := http.NewRequestWithContext(ctx, "GET", routing.gatewayEndpoint, nil)
			require.NoError(t, err, "Failed to create gateway routing request")

			gatewayResp, err := client.Do(gatewayReq)
			if err == nil {
				defer gatewayResp.Body.Close()

				if routing.critical {
					// Critical routes must not return ROUTE_NOT_FOUND
					assert.NotEqual(t, 404, gatewayResp.StatusCode,
						"%s - gateway must route to backend (not 404)", routing.description)
					
					body, err := io.ReadAll(gatewayResp.Body)
					if err == nil {
						responseStr := string(body)
						
						// Must not return ROUTE_NOT_FOUND error
						assert.NotContains(t, responseStr, "ROUTE_NOT_FOUND",
							"%s - gateway must route to backend service", routing.description)
						assert.NotContains(t, responseStr, "route was not found",
							"%s - gateway routing must be configured", routing.description)
						
						// Should return backend service response (not gateway error)
						var jsonData map[string]interface{}
						if json.Unmarshal(body, &jsonData) == nil {
							if errorData, hasError := jsonData["error"]; hasError {
								if errorMap, isMap := errorData.(map[string]interface{}); isMap {
									if code, hasCode := errorMap["code"]; hasCode && code == "ROUTE_NOT_FOUND" {
										t.Errorf("%s - gateway returning ROUTE_NOT_FOUND, routing not configured", routing.description)
									}
								}
							} else {
								// Should contain backend service response indicators
								assert.True(t, gatewayResp.StatusCode >= 200 && gatewayResp.StatusCode < 300,
									"%s - gateway routing should return backend service response", routing.description)
							}
						}
					}
				} else {
					// Non-critical routes should be configured for complete environment
					body, err := io.ReadAll(gatewayResp.Body)
					if err == nil {
						responseStr := string(body)
						if strings.Contains(responseStr, "ROUTE_NOT_FOUND") {
							t.Logf("%s not configured yet (expected for incomplete gateway routing)", routing.description)
						}
					}
				}
			} else {
				if routing.critical {
					t.Errorf("%s not accessible: %v", routing.description, err)
				}
			}
		})
	}
}

func TestGatewayBackendIntegration_ServiceMeshProxying(t *testing.T) {
	// Test that gateways can proxy requests through service mesh to backend services
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Service mesh proxying tests through gateways
	serviceMeshProxyTests := []struct {
		gatewayName    string
		gatewayPath    string
		targetService  string
		targetEndpoint string
		description    string
	}{
		{
			gatewayName:    "public-gateway",
			gatewayPath:    "/api/news",
			targetService:  "content",
			targetEndpoint: "api/news",
			description:    "Public gateway must proxy news requests through service mesh to content service",
		},
		{
			gatewayName:    "admin-gateway",
			gatewayPath:    "/api/admin/inquiries",
			targetService:  "inquiries", 
			targetEndpoint: "api/inquiries",
			description:    "Admin gateway must proxy inquiry requests through service mesh to inquiries service",
		},
	}

	client := &http.Client{Timeout: 15 * time.Second}

	// Act & Assert: Test service mesh proxying
	for _, proxy := range serviceMeshProxyTests {
		t.Run("ServiceMeshProxy_"+proxy.gatewayName+"_to_"+proxy.targetService, func(t *testing.T) {
			// Test that target service is accessible through service mesh
			serviceMeshURL := "http://localhost:3500/v1.0/invoke/" + proxy.targetService + "/method/" + proxy.targetEndpoint
			
			meshReq, err := http.NewRequestWithContext(ctx, "GET", serviceMeshURL, nil)
			require.NoError(t, err, "Failed to create service mesh request")

			meshResp, err := client.Do(meshReq)
			require.NoError(t, err, "Service mesh must be accessible for %s", proxy.targetService)
			defer meshResp.Body.Close()

			assert.True(t, meshResp.StatusCode >= 200 && meshResp.StatusCode < 300,
				"Target service %s must be accessible through service mesh", proxy.targetService)

			// Test that gateway can proxy to the same service
			gatewayURL := "http://localhost:9001" + proxy.gatewayPath
			if proxy.gatewayName == "admin-gateway" {
				gatewayURL = "http://localhost:9000" + proxy.gatewayPath
			}
			
			gatewayReq, err := http.NewRequestWithContext(ctx, "GET", gatewayURL, nil)
			require.NoError(t, err, "Failed to create gateway proxy request")

			gatewayResp, err := client.Do(gatewayReq)
			if err == nil {
				defer gatewayResp.Body.Close()
				
				body, err := io.ReadAll(gatewayResp.Body)
				if err == nil {
					responseStr := string(body)
					
					// Gateway should proxy to backend service (not return routing error)
					assert.NotContains(t, responseStr, "ROUTE_NOT_FOUND",
						"%s - gateway must proxy through service mesh", proxy.description)
					
					// Should return backend service response
					if gatewayResp.StatusCode >= 200 && gatewayResp.StatusCode < 300 {
						var jsonData interface{}
						assert.NoError(t, json.Unmarshal(body, &jsonData),
							"%s - gateway must return backend service JSON response", proxy.description)
					}
				}
			} else {
				t.Errorf("%s gateway proxying not accessible: %v", proxy.description, err)
			}
		})
	}
}

func TestGatewayBackendIntegration_FrontendWorkflowEnablement(t *testing.T) {
	// Test that gateways enable complete frontend development workflow
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Frontend workflow scenarios that must work through gateways
	frontendWorkflowTests := []struct {
		workflow        string
		gatewayEndpoint string
		expectedData    string
		description     string
	}{
		{
			workflow:        "website-news-display",
			gatewayEndpoint: "http://localhost:9001/api/news",
			expectedData:    "news",
			description:     "Website news display workflow through public gateway",
		},
		{
			workflow:        "website-events-display",
			gatewayEndpoint: "http://localhost:9001/api/events",
			expectedData:    "events",
			description:     "Website events display workflow through public gateway",
		},
		{
			workflow:        "admin-inquiry-management",
			gatewayEndpoint: "http://localhost:9000/api/admin/inquiries",
			expectedData:    "inquiries",
			description:     "Admin inquiry management workflow through admin gateway",
		},
	}

	client := &http.Client{Timeout: 15 * time.Second}

	// Act & Assert: Test frontend workflow enablement
	for _, workflow := range frontendWorkflowTests {
		t.Run("FrontendWorkflow_"+workflow.workflow, func(t *testing.T) {
			req, err := http.NewRequestWithContext(ctx, "GET", workflow.gatewayEndpoint, nil)
			require.NoError(t, err, "Failed to create frontend workflow request")

			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				
				body, err := io.ReadAll(resp.Body)
				if err == nil {
					responseStr := string(body)
					
					// Frontend workflow must not encounter routing errors
					assert.NotContains(t, responseStr, "ROUTE_NOT_FOUND",
						"%s - frontend workflow must not encounter routing errors", workflow.description)
					assert.NotContains(t, responseStr, "route was not found",
						"%s - gateway routing must enable frontend workflow", workflow.description)
					
					// Should enable frontend development (return data or proper API responses)
					if resp.StatusCode >= 200 && resp.StatusCode < 300 {
						var jsonData interface{}
						assert.NoError(t, json.Unmarshal(body, &jsonData),
							"%s - must return JSON data for frontend consumption", workflow.description)
						
						t.Logf("%s workflow response: %s", workflow.description, responseStr)
					} else {
						t.Logf("%s workflow status: %d, response: %s", workflow.description, resp.StatusCode, responseStr)
					}
				}
			} else {
				t.Errorf("%s not accessible for frontend workflow: %v", workflow.description, err)
			}
		})
	}
}

// validateEnvironmentPrerequisites ensures environment health before integration testing
func validateEnvironmentPrerequisites(t *testing.T) {
	// Check critical infrastructure, platform, and service components are running
	criticalContainers := []string{"postgresql", "dapr-control-plane", "public-gateway", "admin-gateway", "content", "inquiries", "notifications"}
	
	for _, container := range criticalContainers {
		cmd := exec.Command("podman", "ps", "--filter", "name="+container, "--format", "{{.Names}}")
		output, err := cmd.Output()
		require.NoError(t, err, "Failed to check critical container %s", container)

		if !strings.Contains(string(output), container) {
			t.Skipf("Critical container %s not running - environment not ready for integration testing", container)
		}
	}
}