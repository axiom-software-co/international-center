package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED PHASE: Cross-Service Communication Tests
// These tests validate service-to-service communication between all services through service mesh

func TestCrossServiceCommunication_ServiceMeshIntegration(t *testing.T) {
	// This test requires complete environment health - enforcing axiom rule
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Define all possible service communication patterns
	serviceCommunicationMatrix := []struct {
		sourceService string
		targetService string
		method        string
		workflow      string
		critical      bool
	}{
		// Gateway to service communication (critical for website functionality)
		{"public-gateway", "content", "health", "Website content delivery", true},
		{"public-gateway", "notifications", "health", "Website form submissions", true},
		{"admin-gateway", "content", "health", "Admin content management", true},
		{"admin-gateway", "inquiries", "health", "Admin inquiry management", true},
		{"admin-gateway", "notifications", "health", "Admin notification triggers", true},
		
		// Service to service communication (for domain integration)
		{"content", "notifications", "health", "Content publishing notifications", false},
		{"inquiries", "notifications", "health", "Inquiry submission notifications", false},
		{"content", "inquiries", "health", "Content inquiry integration", false},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Act & Assert: Test all service mesh communication patterns
	for _, comm := range serviceCommunicationMatrix {
		t.Run(fmt.Sprintf("ServiceMesh_%s_to_%s", comm.sourceService, comm.targetService), func(t *testing.T) {
			// Test service invocation through Dapr service mesh
			serviceInvocationURL := fmt.Sprintf("http://localhost:3500/v1.0/invoke/%s/method/%s", comm.targetService, comm.method)
			
			req, err := http.NewRequestWithContext(ctx, "GET", serviceInvocationURL, nil)
			require.NoError(t, err, "Failed to create service mesh communication request")

			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				
				if comm.critical {
					// Critical communications must work for basic development environment
					assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300, 
						"Critical workflow '%s' - service mesh communication must be functional", comm.workflow)
					assert.NotEqual(t, 500, resp.StatusCode, 
						"Critical workflow '%s' - target service %s must be discoverable", comm.workflow, comm.targetService)
				} else {
					// Non-critical communications should work for complete environment
					assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500, 
						"Workflow '%s' - service mesh communication should be accessible", comm.workflow)
				}

				// Validate response indicates proper service mesh routing
				body, err := io.ReadAll(resp.Body)
				if err == nil {
					responseStr := string(body)
					assert.NotContains(t, responseStr, "couldn't find service", 
						"Service %s must be registered with Dapr for workflow '%s'", comm.targetService, comm.workflow)
				}
			} else {
				if comm.critical {
					t.Errorf("Critical workflow '%s' not functional: %v", comm.workflow, err)
				} else {
					t.Logf("Workflow '%s' not functional (expected for incomplete deployment): %v", comm.workflow, err)
				}
			}
		})
	}
}

func TestCrossServiceCommunication_DirectAPIAccess(t *testing.T) {
	// Test that services are accessible through both direct and service mesh routing
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// API endpoints that should be accessible directly and through service mesh
	apiEndpoints := []struct {
		serviceName    string
		directURL      string
		serviceMeshURL string
		endpoint       string
		description    string
	}{
		{
			serviceName:    "content",
			directURL:      "http://localhost:3001/api/news",
			serviceMeshURL: "http://localhost:3500/v1.0/invoke/content/method/api/news",
			endpoint:       "news",
			description:    "Content news API access",
		},
		{
			serviceName:    "content",
			directURL:      "http://localhost:3001/api/events", 
			serviceMeshURL: "http://localhost:3500/v1.0/invoke/content/method/api/events",
			endpoint:       "events",
			description:    "Content events API access",
		},
		{
			serviceName:    "inquiries",
			directURL:      "http://localhost:3101/api/inquiries",
			serviceMeshURL: "http://localhost:3500/v1.0/invoke/inquiries/method/api/inquiries",
			endpoint:       "inquiries",
			description:    "Inquiries API access",
		},
		{
			serviceName:    "notifications",
			directURL:      "http://localhost:3201/api/notifications",
			serviceMeshURL: "http://localhost:3500/v1.0/invoke/notifications/method/api/notifications", 
			endpoint:       "notifications",
			description:    "Notifications API access",
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Act & Assert: Test API accessibility through both routes
	for _, api := range apiEndpoints {
		t.Run("APIAccess_"+api.serviceName+"_"+api.endpoint, func(t *testing.T) {
			// Test direct API access
			directReq, err := http.NewRequestWithContext(ctx, "GET", api.directURL, nil)
			require.NoError(t, err, "Failed to create direct API request")

			directResp, err := client.Do(directReq)
			if err == nil {
				defer directResp.Body.Close()
				assert.True(t, directResp.StatusCode >= 200 && directResp.StatusCode < 500, 
					"%s - direct API access should be functional", api.description)
			} else {
				t.Logf("%s direct access not functional: %v", api.description, err)
			}

			// Test service mesh API access
			meshReq, err := http.NewRequestWithContext(ctx, "GET", api.serviceMeshURL, nil)
			require.NoError(t, err, "Failed to create service mesh API request")

			meshResp, err := client.Do(meshReq)
			if err == nil {
				defer meshResp.Body.Close()
				assert.True(t, meshResp.StatusCode >= 200 && meshResp.StatusCode < 500, 
					"%s - service mesh API access should be functional", api.description)
					
				// Validate response is not service discovery error
				body, err := io.ReadAll(meshResp.Body)
				if err == nil {
					responseStr := string(body)
					assert.NotContains(t, responseStr, "couldn't find service", 
						"%s - service must be discoverable through service mesh", api.description)
				}
			} else {
				t.Errorf("%s service mesh access not functional: %v", api.description, err)
			}
		})
	}
}

func TestCrossServiceCommunication_GatewayRouting(t *testing.T) {
	// Test gateway routing to backend services (critical for website functionality)
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Gateway routing patterns that must work
	gatewayRoutingTests := []struct {
		gatewayName   string
		gatewayURL    string
		targetService string
		routePath     string
		description   string
	}{
		{
			gatewayName:   "public-gateway",
			gatewayURL:    "http://localhost:9001",
			targetService: "content",
			routePath:     "/api/news",
			description:   "Public gateway routing to content service for website",
		},
		{
			gatewayName:   "public-gateway",
			gatewayURL:    "http://localhost:9001",
			targetService: "content",
			routePath:     "/api/events",
			description:   "Public gateway routing to events for website",
		},
		{
			gatewayName:   "admin-gateway",
			gatewayURL:    "http://localhost:9000",
			targetService: "inquiries",
			routePath:     "/api/admin/inquiries",
			description:   "Admin gateway routing to inquiries for portal management",
		},
		{
			gatewayName:   "admin-gateway",
			gatewayURL:    "http://localhost:9000",
			targetService: "content",
			routePath:     "/api/admin/content",
			description:   "Admin gateway routing to content for portal management",
		},
	}

	client := &http.Client{Timeout: 15 * time.Second}

	// Act & Assert: Test gateway routing functionality
	for _, routing := range gatewayRoutingTests {
		t.Run("GatewayRouting_"+routing.gatewayName+"_to_"+routing.targetService, func(t *testing.T) {
			// Test gateway routing endpoint
			routingURL := routing.gatewayURL + routing.routePath
			
			req, err := http.NewRequestWithContext(ctx, "GET", routingURL, nil)
			require.NoError(t, err, "Failed to create gateway routing request")

			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500, 
					"%s - gateway routing must be functional", routing.description)
				
				// Test that gateway can reach target service through service mesh
				serviceMeshURL := fmt.Sprintf("http://localhost:3500/v1.0/invoke/%s/method/health", routing.targetService)
				meshReq, err := http.NewRequestWithContext(ctx, "GET", serviceMeshURL, nil)
				require.NoError(t, err, "Failed to create service mesh verification request")

				meshResp, err := client.Do(meshReq)
				if err == nil {
					defer meshResp.Body.Close()
					assert.True(t, meshResp.StatusCode >= 200 && meshResp.StatusCode < 300, 
						"Target service %s must be accessible for gateway routing", routing.targetService)
				}
			} else {
				t.Errorf("%s not functional: %v", routing.description, err)
			}
		})
	}
}

// validateEnvironmentPrerequisites ensures environment health before integration testing
func validateEnvironmentPrerequisites(t *testing.T) {
	// Check critical infrastructure and platform components are running
	criticalContainers := []string{"postgresql", "dapr-control-plane"}
	
	for _, container := range criticalContainers {
		cmd := exec.Command("podman", "ps", "--filter", "name="+container, "--format", "{{.Names}}")
		output, err := cmd.Output()
		require.NoError(t, err, "Failed to check critical container %s", container)

		if !strings.Contains(string(output), container) {
			t.Skipf("Critical container %s not running - environment not ready for integration testing", container)
		}
	}
}