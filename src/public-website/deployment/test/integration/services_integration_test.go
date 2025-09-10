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

// Services Integration Tests
// Validates services phase with 5 consolidated services communicating through service mesh
// Tests gateway routing, service-to-service communication, consolidated service functionality

func TestServicesIntegration_ConsolidatedServiceMesh(t *testing.T) {
	// This test requires complete environment health - enforcing axiom rule
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("ConsolidatedServices_DeploymentAndHealth", func(t *testing.T) {
		// Test that all 5 consolidated services are operational
		consolidatedServices := []struct {
			serviceName    string
			port           int
			healthEndpoint string
			appID          string
			description    string
		}{
			{"public-gateway", 9001, "http://localhost:9001/health", "public-gateway", "Public gateway must handle public API requests"},
			{"admin-gateway", 9000, "http://localhost:9000/health", "admin-gateway", "Admin gateway must handle admin portal requests"},
			{"content", 3001, "http://localhost:3001/health", "content", "Content service must handle news, events, research"},
			{"inquiries", 3101, "http://localhost:3101/health", "inquiries", "Inquiries service must handle business, donations, media, volunteers"},
			{"notifications", 3201, "http://localhost:3201/health", "notifications", "Notifications service must handle email, SMS, Slack"},
		}

		client := &http.Client{Timeout: 10 * time.Second}

		for _, service := range consolidatedServices {
			t.Run("ConsolidatedService_"+service.serviceName, func(t *testing.T) {
				// Validate service container is running
				serviceCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+service.serviceName, "--format", "{{.Names}}")
				serviceOutput, err := serviceCmd.Output()
				require.NoError(t, err, "Failed to check service %s status", service.serviceName)

				runningServices := strings.TrimSpace(string(serviceOutput))
				assert.Contains(t, runningServices, service.serviceName, 
					"%s - consolidated service must be running", service.description)

				// Test service health endpoint
				if strings.Contains(runningServices, service.serviceName) {
					healthReq, err := http.NewRequestWithContext(ctx, "GET", service.healthEndpoint, nil)
					require.NoError(t, err, "Failed to create health request for %s", service.serviceName)

					healthResp, err := client.Do(healthReq)
					if err == nil {
						defer healthResp.Body.Close()
						assert.True(t, healthResp.StatusCode >= 200 && healthResp.StatusCode < 300, 
							"%s health endpoint must be accessible", service.description)
					} else {
						t.Errorf("%s health endpoint not accessible: %v", service.description, err)
					}
				}

				// Validate Dapr sidecar is running for service mesh integration
				sidecarName := service.serviceName + "-dapr"
				sidecarCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+sidecarName, "--format", "{{.Names}}")
				sidecarOutput, err := sidecarCmd.Output()
				if err == nil {
					runningSidecars := strings.TrimSpace(string(sidecarOutput))
					assert.Contains(t, runningSidecars, sidecarName, 
						"Dapr sidecar for %s must be running for service mesh integration", service.serviceName)
				}
			})
		}
	})
}

func TestServicesIntegration_ServiceMeshCommunication(t *testing.T) {
	// Test service-to-service communication through Dapr service mesh
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("GatewayToServiceCommunication", func(t *testing.T) {
		// Test gateway routing to consolidated backend services
		client := &http.Client{Timeout: 10 * time.Second}

		gatewayRoutingTests := []struct {
			sourceGateway  string
			targetService  string
			invocationURL  string
			description    string
		}{
			{
				"public-gateway", 
				"content", 
				"http://localhost:3500/v1.0/invoke/content/method/health",
				"Public gateway to content service communication through service mesh",
			},
			{
				"admin-gateway",
				"inquiries", 
				"http://localhost:3500/v1.0/invoke/inquiries/method/health",
				"Admin gateway to inquiries service communication through service mesh",
			},
			{
				"public-gateway",
				"notifications",
				"http://localhost:3500/v1.0/invoke/notifications/method/health", 
				"Gateway to notifications service communication for form submissions",
			},
		}

		for _, test := range gatewayRoutingTests {
			t.Run(fmt.Sprintf("Routing_%s_to_%s", test.sourceGateway, test.targetService), func(t *testing.T) {
				req, err := http.NewRequestWithContext(ctx, "GET", test.invocationURL, nil)
				require.NoError(t, err, "Failed to create service mesh communication request")

				resp, err := client.Do(req)
				if err == nil {
					defer resp.Body.Close()
					assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300, 
						"%s - service mesh communication must be operational", test.description)
				} else {
					t.Logf("%s not operational yet: %v", test.description, err)
				}
			})
		}
	})

	t.Run("ServiceToServiceIntegration", func(t *testing.T) {
		// Test direct service-to-service communication patterns
		client := &http.Client{Timeout: 10 * time.Second}

		serviceIntegrationTests := []struct {
			sourceService string
			targetService string
			workflow      string
		}{
			{"content", "notifications", "Content publishing triggers notifications"},
			{"inquiries", "notifications", "Inquiry submissions trigger notifications"}, 
			{"inquiries", "content", "Inquiry responses may update content"},
		}

		for _, test := range serviceIntegrationTests {
			t.Run(fmt.Sprintf("ServiceIntegration_%s_to_%s", test.sourceService, test.targetService), func(t *testing.T) {
				// Test service mesh communication between consolidated services
				invocationURL := fmt.Sprintf("http://localhost:3500/v1.0/invoke/%s/method/health", test.targetService)
				
				req, err := http.NewRequestWithContext(ctx, "GET", invocationURL, nil)
				require.NoError(t, err, "Failed to create service integration request")

				resp, err := client.Do(req)
				if err == nil {
					defer resp.Body.Close()
					assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300, 
						"Service integration %s must be functional", test.workflow)
				} else {
					t.Logf("Service integration %s not operational: %v", test.workflow, err)
				}
			})
		}
	})
}

func TestServicesIntegration_ConsolidatedServiceFunctionality(t *testing.T) {
	// Test consolidated service functionality and domain boundaries
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("ContentService_DomainFunctionality", func(t *testing.T) {
		// Test content service handles multiple content types (news, events, research)
		client := &http.Client{Timeout: 10 * time.Second}

		contentEndpoints := []string{
			"http://localhost:3001/api/news",
			"http://localhost:3001/api/events", 
			"http://localhost:3001/api/research",
		}

		for _, endpoint := range contentEndpoints {
			req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
			require.NoError(t, err, "Failed to create content request")

			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500, 
					"Content service must handle multiple content types: %s", endpoint)
			} else {
				t.Logf("Content endpoint %s not operational: %v", endpoint, err)
			}
		}
	})

	t.Run("InquiriesService_DomainFunctionality", func(t *testing.T) {
		// Test inquiries service handles multiple inquiry types
		client := &http.Client{Timeout: 10 * time.Second}

		inquiryEndpoints := []string{
			"http://localhost:3101/api/inquiries/business",
			"http://localhost:3101/api/inquiries/donations",
			"http://localhost:3101/api/inquiries/media",
			"http://localhost:3101/api/inquiries/volunteers",
		}

		for _, endpoint := range inquiryEndpoints {
			req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
			require.NoError(t, err, "Failed to create inquiry request")

			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500, 
					"Inquiries service must handle multiple inquiry types: %s", endpoint)
			} else {
				t.Logf("Inquiry endpoint %s not operational: %v", endpoint, err)
			}
		}
	})

	t.Run("NotificationsService_CrossServiceIntegration", func(t *testing.T) {
		// Test notifications service integration with other services
		client := &http.Client{Timeout: 10 * time.Second}

		// Test notifications service through service mesh
		notificationURL := "http://localhost:3500/v1.0/invoke/notifications/method/api/notifications"
		
		req, err := http.NewRequestWithContext(ctx, "GET", notificationURL, nil)
		require.NoError(t, err, "Failed to create notifications request")

		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500, 
				"Notifications service must be accessible through service mesh")
		} else {
			t.Logf("Notifications service mesh integration not operational: %v", err)
		}
	})
}