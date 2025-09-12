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

// RED PHASE: Complete Gateway Functionality Tests
// These tests validate that both public and admin gateways are operational with proper API routing

func TestCompleteGatewayFunctionality_BothGatewaysOperational(t *testing.T) {
	// This test requires complete environment health - enforcing axiom rule
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Both gateways that must be operational
	gatewayTests := []struct {
		gatewayName     string
		containerName   string
		healthEndpoint  string
		port            int
		expectedType    string
		description     string
		critical        bool
	}{
		{
			gatewayName:    "public-gateway",
			containerName:  "public-gateway",
			healthEndpoint: "http://localhost:9001/health",
			port:           9001,
			expectedType:   "public",
			description:    "Public gateway must be operational for website frontend",
			critical:       true,
		},
		{
			gatewayName:    "admin-gateway",
			containerName:  "admin-gateway",
			healthEndpoint: "http://localhost:9000/health",
			port:           9000,
			expectedType:   "admin",
			description:    "Admin gateway must be operational for admin portal",
			critical:       true,
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Act & Assert: Validate both gateways are operational
	for _, gateway := range gatewayTests {
		t.Run("GatewayOperational_"+gateway.gatewayName, func(t *testing.T) {
			// Check if gateway container is running
			statusCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+gateway.containerName, "--format", "{{.Status}}")
			statusOutput, err := statusCmd.Output()
			require.NoError(t, err, "Failed to check gateway %s status", gateway.containerName)

			status := strings.TrimSpace(string(statusOutput))
			
			if gateway.critical {
				// Critical gateways must be running
				assert.Contains(t, status, "Up", 
					"%s - critical gateway must be running", gateway.description)
				assert.NotContains(t, status, "Exited", 
					"%s - critical gateway must not exit due to configuration issues", gateway.description)
			}

			// If gateway is running, test health endpoint
			if strings.Contains(status, "Up") {
				healthReq, err := http.NewRequestWithContext(ctx, "GET", gateway.healthEndpoint, nil)
				require.NoError(t, err, "Failed to create health request for %s", gateway.gatewayName)

				healthResp, err := client.Do(healthReq)
				if err == nil {
					defer healthResp.Body.Close()
					assert.True(t, healthResp.StatusCode >= 200 && healthResp.StatusCode < 300,
						"%s health endpoint must be accessible", gateway.description)

					// Validate health response structure
					body, err := io.ReadAll(healthResp.Body)
					if err == nil {
						var healthData map[string]interface{}
						assert.NoError(t, json.Unmarshal(body, &healthData),
							"%s health endpoint must return valid JSON", gateway.description)

						// Validate gateway type in response
						if gatewayField, exists := healthData["gateway"]; exists {
							assert.Equal(t, gateway.expectedType+"-gateway", gatewayField,
								"%s must identify as %s gateway", gateway.description, gateway.expectedType)
						}
					}
				} else {
					if gateway.critical {
						t.Errorf("%s health endpoint not accessible: %v", gateway.description, err)
					}
				}
			} else if gateway.critical {
				t.Logf("%s not running (critical for development environment): %s", gateway.description, status)
			}
		})
	}
}

func TestCompleteGatewayFunctionality_APIRoutingToBackendServices(t *testing.T) {
	// Test that both gateways can route to backend services properly
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Complete API routing scenarios for both gateways
	apiRoutingTests := []struct {
		gatewayName     string
		apiEndpoint     string
		backendService  string
		backendEndpoint string
		description     string
		critical        bool
	}{
		{
			gatewayName:     "public-gateway",
			apiEndpoint:     "http://localhost:9001/api/news",
			backendService:  "content",
			backendEndpoint: "http://localhost:3001/api/news",
			description:     "Public gateway must route news API to content service",
			critical:        true,
		},
		{
			gatewayName:     "public-gateway",
			apiEndpoint:     "http://localhost:9001/api/events",
			backendService:  "content",
			backendEndpoint: "http://localhost:3001/api/events",
			description:     "Public gateway must route events API to content service",
			critical:        true,
		},
		{
			gatewayName:     "admin-gateway",
			apiEndpoint:     "http://localhost:9000/api/admin/inquiries",
			backendService:  "inquiries",
			backendEndpoint: "http://localhost:3101/api/inquiries",
			description:     "Admin gateway must route inquiries API to inquiries service",
			critical:        true,
		},
		{
			gatewayName:     "admin-gateway",
			apiEndpoint:     "http://localhost:9000/api/admin/subscribers",
			backendService:  "notifications",
			backendEndpoint: "http://localhost:3201/api/subscribers",
			description:     "Admin gateway must route subscribers API to notifications service",
			critical:        false,
		},
	}

	client := &http.Client{Timeout: 15 * time.Second}

	// Act & Assert: Test complete API routing functionality
	for _, routing := range apiRoutingTests {
		t.Run("APIRouting_"+routing.gatewayName+"_"+routing.backendService, func(t *testing.T) {
			// First verify backend service is accessible
			backendReq, err := http.NewRequestWithContext(ctx, "GET", routing.backendEndpoint, nil)
			require.NoError(t, err, "Failed to create backend request")

			backendResp, err := client.Do(backendReq)
			require.NoError(t, err, "Backend service %s must be accessible", routing.backendService)
			defer backendResp.Body.Close()

			assert.True(t, backendResp.StatusCode >= 200 && backendResp.StatusCode < 300,
				"Backend service %s must be functional", routing.backendService)

			// Test gateway routing to backend service
			gatewayReq, err := http.NewRequestWithContext(ctx, "GET", routing.apiEndpoint, nil)
			require.NoError(t, err, "Failed to create gateway request")

			gatewayResp, err := client.Do(gatewayReq)
			if err == nil {
				defer gatewayResp.Body.Close()

				if routing.critical {
					// Critical routes must work for development environment
					assert.True(t, gatewayResp.StatusCode >= 200 && gatewayResp.StatusCode < 500,
						"%s - critical API routing must be functional", routing.description)

					body, err := io.ReadAll(gatewayResp.Body)
					if err == nil {
						responseStr := string(body)
						
						// Must not return ROUTE_NOT_FOUND errors
						assert.NotContains(t, responseStr, "ROUTE_NOT_FOUND",
							"%s - gateway must route to backend service", routing.description)
						assert.NotContains(t, responseStr, "route was not found",
							"%s - API routing must be configured", routing.description)
						
						// Must not return path parsing errors
						assert.NotContains(t, responseStr, "invalid API path format",
							"%s - API path parsing must work correctly", routing.description)
						assert.NotContains(t, responseStr, "unknown service",
							"%s - service discovery must work", routing.description)
						
						// Should return backend service response or proper gateway proxy response
						if gatewayResp.StatusCode >= 200 && gatewayResp.StatusCode < 300 {
							var jsonData interface{}
							assert.NoError(t, json.Unmarshal(body, &jsonData),
								"%s - gateway must return JSON response from backend", routing.description)
						}
					}
				} else {
					// Non-critical routes should work for complete environment
					body, err := io.ReadAll(gatewayResp.Body)
					if err == nil {
						responseStr := string(body)
						if strings.Contains(responseStr, "ROUTE_NOT_FOUND") || strings.Contains(responseStr, "unknown service") {
							t.Logf("%s not functional yet (expected for incomplete routing)", routing.description)
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

func TestCompleteGatewayFunctionality_FrontendWorkflowEnablement(t *testing.T) {
	// Test that complete frontend development workflow is enabled through gateways
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Frontend workflow scenarios that must work for development
	frontendWorkflowTests := []struct {
		workflow        string
		gatewayEndpoint string
		frontendUse     string
		description     string
		critical        bool
	}{
		{
			workflow:        "website-content-consumption",
			gatewayEndpoint: "http://localhost:9001/api/news",
			frontendUse:     "public website news section",
			description:     "Website must consume news content through public gateway",
			critical:        true,
		},
		{
			workflow:        "website-events-consumption",
			gatewayEndpoint: "http://localhost:9001/api/events",
			frontendUse:     "public website events section",
			description:     "Website must consume events content through public gateway",
			critical:        true,
		},
		{
			workflow:        "admin-inquiry-management",
			gatewayEndpoint: "http://localhost:9000/api/admin/inquiries",
			frontendUse:     "admin portal inquiry management",
			description:     "Admin portal must manage inquiries through admin gateway",
			critical:        false,
		},
		{
			workflow:        "admin-subscriber-management",
			gatewayEndpoint: "http://localhost:9000/api/admin/subscribers",
			frontendUse:     "admin portal subscriber management",
			description:     "Admin portal must manage subscribers through admin gateway",
			critical:        false,
		},
	}

	client := &http.Client{Timeout: 15 * time.Second}

	// Act & Assert: Test frontend development workflow enablement
	for _, workflow := range frontendWorkflowTests {
		t.Run("FrontendWorkflow_"+workflow.workflow, func(t *testing.T) {
			req, err := http.NewRequestWithContext(ctx, "GET", workflow.gatewayEndpoint, nil)
			require.NoError(t, err, "Failed to create frontend workflow request")

			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()

				if workflow.critical {
					// Critical workflows must work for development environment
					assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500,
						"%s - critical workflow must be accessible", workflow.description)

					body, err := io.ReadAll(resp.Body)
					if err == nil {
						responseStr := string(body)
						
						// Must enable frontend development (not return routing errors)
						assert.NotContains(t, responseStr, "ROUTE_NOT_FOUND",
							"%s - frontend workflow must not encounter routing errors", workflow.description)
						assert.NotContains(t, responseStr, "invalid API path format",
							"%s - API path parsing must support frontend use case", workflow.description)
						
						// Should return data suitable for frontend consumption
						if resp.StatusCode >= 200 && resp.StatusCode < 300 {
							var jsonData interface{}
							assert.NoError(t, json.Unmarshal(body, &jsonData),
								"%s - must return JSON for %s", workflow.description, workflow.frontendUse)
						}
					}
				} else {
					// Non-critical workflows should work for complete environment
					body, err := io.ReadAll(resp.Body)
					if err == nil {
						responseStr := string(body)
						if strings.Contains(responseStr, "ROUTE_NOT_FOUND") {
							t.Logf("%s not functional for %s (expected for incomplete routing)", workflow.description, workflow.frontendUse)
						}
					}
				}
			} else {
				if workflow.critical {
					t.Errorf("%s not accessible for %s: %v", workflow.description, workflow.frontendUse, err)
				}
			}
		})
	}
}

func TestCompleteGatewayFunctionality_ServiceMeshIntegration(t *testing.T) {
	// Test that gateways integrate properly with service mesh for backend communication
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Service mesh integration tests through gateways
	serviceMeshIntegrationTests := []struct {
		gatewayName      string
		serviceEndpoint  string
		targetService    string
		targetMethod     string
		description      string
	}{
		{
			gatewayName:     "admin-gateway",
			serviceEndpoint: "http://localhost:9000/health",
			targetService:   "inquiries",
			targetMethod:    "health",
			description:     "Admin gateway must integrate with service mesh to reach inquiries service",
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Act & Assert: Test service mesh integration
	for _, integration := range serviceMeshIntegrationTests {
		t.Run("ServiceMeshIntegration_"+integration.gatewayName, func(t *testing.T) {
			// Test gateway health (verifies service mesh integration)
			gatewayReq, err := http.NewRequestWithContext(ctx, "GET", integration.serviceEndpoint, nil)
			require.NoError(t, err, "Failed to create service mesh integration request")

			gatewayResp, err := client.Do(gatewayReq)
			require.NoError(t, err, "Gateway service mesh integration must be accessible")
			defer gatewayResp.Body.Close()

			assert.True(t, gatewayResp.StatusCode >= 200 && gatewayResp.StatusCode < 300,
				"%s - service mesh integration must be functional", integration.description)

			// Validate response indicates healthy backend services
			body, err := io.ReadAll(gatewayResp.Body)
			require.NoError(t, err, "Failed to read service mesh integration response")

			responseStr := string(body)
			
			// Should report healthy backend services
			assert.Contains(t, responseStr, "backend_services",
				"%s - must report backend service status", integration.description)
			assert.Contains(t, responseStr, "healthy",
				"%s - backend services must be healthy through service mesh", integration.description)

			// Test direct service mesh access to verify target service
			directServiceMeshURL := "http://localhost:3500/v1.0/invoke/" + integration.targetService + "/method/" + integration.targetMethod
			meshReq, err := http.NewRequestWithContext(ctx, "GET", directServiceMeshURL, nil)
			require.NoError(t, err, "Failed to create direct service mesh request")

			meshResp, err := client.Do(meshReq)
			if err == nil {
				defer meshResp.Body.Close()
				assert.True(t, meshResp.StatusCode >= 200 && meshResp.StatusCode < 300,
					"Target service %s must be accessible through service mesh", integration.targetService)
			}
		})
	}
}

// validateEnvironmentPrerequisites ensures environment health before integration testing
func validateEnvironmentPrerequisites(t *testing.T) {
	// Check critical infrastructure, platform, and service components are running
	criticalContainers := []string{"postgresql", "dapr-control-plane", "content", "inquiries", "notifications"}
	
	for _, container := range criticalContainers {
		cmd := exec.Command("podman", "ps", "--filter", "name="+container, "--format", "{{.Names}}")
		output, err := cmd.Output()
		require.NoError(t, err, "Failed to check critical container %s", container)

		if !strings.Contains(string(output), container) {
			t.Skipf("Critical container %s not running - environment not ready for integration testing", container)
		}
	}
}