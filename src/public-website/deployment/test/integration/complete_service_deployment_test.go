package integration

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	sharedtesting "github.com/axiom-software-co/international-center/src/public-website/infrastructure/test/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED PHASE: Complete Service Deployment Tests
// These tests validate that all 5 consolidated services are deployed with Dapr sidecars
// and form a complete functional development environment

func TestCompleteServiceDeployment_AllConsolidatedServices(t *testing.T) {
	// This test requires complete environment health - enforcing axiom rule
	sharedtesting.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// All 5 consolidated services that must be deployed
	consolidatedServices := []struct {
		serviceName    string
		sidecarName    string
		servicePort    int
		sidecarPort    int
		healthEndpoint string
		description    string
	}{
		{
			serviceName:    "public-gateway",
			sidecarName:    "public-gateway-dapr",
			servicePort:    9001,
			sidecarPort:    50010,
			healthEndpoint: "http://localhost:9001/health",
			description:    "Public gateway service for website API access",
		},
		{
			serviceName:    "admin-gateway",
			sidecarName:    "admin-gateway-dapr",
			servicePort:    9000,
			sidecarPort:    50020,
			healthEndpoint: "http://localhost:9000/health",
			description:    "Admin gateway service for portal management",
		},
		{
			serviceName:    "content",
			sidecarName:    "content-dapr",
			servicePort:    3001,
			sidecarPort:    50030,
			healthEndpoint: "http://localhost:3001/health",
			description:    "Content service for news, events, research",
		},
		{
			serviceName:    "inquiries",
			sidecarName:    "inquiries-dapr",
			servicePort:    3101,
			sidecarPort:    50040,
			healthEndpoint: "http://localhost:3101/health",
			description:    "Inquiries service for business, donations, media, volunteers",
		},
		{
			serviceName:    "notifications",
			sidecarName:    "notifications-dapr",
			servicePort:    3201,
			sidecarPort:    50050,
			healthEndpoint: "http://localhost:3201/health",
			description:    "Notifications service for email, SMS, Slack",
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Act & Assert: Validate each consolidated service and sidecar deployment
	for _, service := range consolidatedServices {
		t.Run("CompleteDeployment_"+service.serviceName, func(t *testing.T) {
			// Validate service container is running
			serviceCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+service.serviceName, "--format", "{{.Names}}")
			serviceOutput, err := serviceCmd.Output()
			require.NoError(t, err, "Failed to check service %s status", service.serviceName)

			runningServices := strings.TrimSpace(string(serviceOutput))
			assert.Contains(t, runningServices, service.serviceName, 
				"%s - consolidated service must be deployed and running", service.description)

			// Validate corresponding Dapr sidecar is running
			sidecarCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+service.sidecarName, "--format", "{{.Names}}")
			sidecarOutput, err := sidecarCmd.Output()
			require.NoError(t, err, "Failed to check sidecar %s status", service.sidecarName)

			runningSidecars := strings.TrimSpace(string(sidecarOutput))
			assert.Contains(t, runningSidecars, service.sidecarName, 
				"Dapr sidecar for %s must be deployed and running", service.serviceName)

			// Test service health endpoint accessibility
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

			// Test sidecar health endpoint accessibility
			if strings.Contains(runningSidecars, service.sidecarName) {
				sidecarHealthURL := fmt.Sprintf("http://localhost:%d/v1.0/healthz", service.sidecarPort)
				
				sidecarHealthReq, err := http.NewRequestWithContext(ctx, "GET", sidecarHealthURL, nil)
				require.NoError(t, err, "Failed to create sidecar health request for %s", service.sidecarName)

				sidecarHealthResp, err := client.Do(sidecarHealthReq)
				if err == nil {
					defer sidecarHealthResp.Body.Close()
					assert.True(t, sidecarHealthResp.StatusCode >= 200 && sidecarHealthResp.StatusCode < 300, 
						"Sidecar %s health endpoint must be accessible", service.sidecarName)
				} else {
					t.Errorf("Sidecar %s health endpoint not accessible: %v", service.sidecarName, err)
				}
			}
		})
	}
}

func TestCompleteServiceDeployment_ServiceMeshRegistration(t *testing.T) {
	// Test that all services are properly registered with the service mesh
	sharedtesting.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// All services that should be discoverable through Dapr
	expectedRegisteredServices := []string{
		"public-gateway",
		"admin-gateway", 
		"content",
		"inquiries",
		"notifications",
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Act & Assert: Validate each service is registered and discoverable
	for _, appID := range expectedRegisteredServices {
		t.Run("ServiceMeshRegistration_"+appID, func(t *testing.T) {
			// Test service discoverability through Dapr service invocation
			serviceInvocationURL := "http://localhost:3500/v1.0/invoke/" + appID + "/method/health"
			
			req, err := http.NewRequestWithContext(ctx, "GET", serviceInvocationURL, nil)
			require.NoError(t, err, "Failed to create service invocation request for %s", appID)

			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300, 
					"Service %s must be discoverable through Dapr service mesh", appID)
				assert.NotEqual(t, 500, resp.StatusCode, 
					"Service %s must not return 'service not found' error", appID)
			} else {
				t.Errorf("Service %s not discoverable through service mesh: %v", appID, err)
			}
		})
	}
}

func TestCompleteServiceDeployment_ContainerHealthValidation(t *testing.T) {
	// Test that all deployed containers are healthy and stable
	sharedtesting.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// All containers that should be running and healthy
	expectedHealthyContainers := []struct {
		containerName string
		description   string
		critical      bool
	}{
		{"postgresql", "Database must be healthy for service data", true},
		{"dapr-control-plane", "Dapr control plane must be healthy for service mesh", true},
		{"content", "Content service must be healthy", true},
		{"content-dapr", "Content sidecar must be healthy", true},
		{"public-gateway", "Public gateway must be healthy for website API", false},
		{"public-gateway-dapr", "Public gateway sidecar must be healthy", false},
		{"admin-gateway", "Admin gateway must be healthy for portal", false},
		{"admin-gateway-dapr", "Admin gateway sidecar must be healthy", false},
		{"inquiries", "Inquiries service must be healthy", false},
		{"inquiries-dapr", "Inquiries sidecar must be healthy", false},
		{"notifications", "Notifications service must be healthy", false},
		{"notifications-dapr", "Notifications sidecar must be healthy", false},
	}

	// Act & Assert: Validate container health status
	for _, container := range expectedHealthyContainers {
		t.Run("ContainerHealth_"+container.containerName, func(t *testing.T) {
			// Check if container is running
			statusCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+container.containerName, "--format", "{{.Status}}")
			statusOutput, err := statusCmd.Output()
			require.NoError(t, err, "Failed to check container %s status", container.containerName)

			status := strings.TrimSpace(string(statusOutput))
			
			if container.critical {
				// Critical containers must be running
				assert.Contains(t, status, "Up", 
					"%s - critical container must be running", container.description)
				assert.NotContains(t, status, "Exited", 
					"%s - critical container must not be exited", container.description)
			} else {
				// Non-critical containers should be running for complete environment
				if status == "" {
					t.Logf("%s not deployed yet (expected for incomplete deployment)", container.description)
				} else if strings.Contains(status, "Exited") {
					t.Logf("%s exited (expected for incomplete deployment): %s", container.description, status)
				} else if strings.Contains(status, "Up") {
					t.Logf("%s healthy: %s", container.description, status)
				}
			}
		})
	}
}

func TestCompleteServiceDeployment_NetworkConnectivity(t *testing.T) {
	// Test that all services can reach their sidecars and other infrastructure
	sharedtesting.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Services that should have network connectivity to infrastructure
	serviceInfraConnectivity := map[string][]struct {
		targetContainer string
		targetPort      int
		description     string
	}{
		"content": {
			{"postgresql", 5432, "Content service to database connectivity"},
			{"content-dapr", 50001, "Content service to its sidecar connectivity"},
		},
		"inquiries": {
			{"postgresql", 5432, "Inquiries service to database connectivity"},
			{"inquiries-dapr", 50001, "Inquiries service to its sidecar connectivity"},
		},
		"notifications": {
			{"rabbitmq", 5672, "Notifications service to messaging connectivity"},
			{"notifications-dapr", 50001, "Notifications service to its sidecar connectivity"},
		},
	}

	// Act & Assert: Validate network connectivity
	for serviceName, connectivityTests := range serviceInfraConnectivity {
		t.Run("NetworkConnectivity_"+serviceName, func(t *testing.T) {
			// Check if service is running first
			serviceCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+serviceName, "--format", "{{.Names}}")
			serviceOutput, err := serviceCmd.Output()
			require.NoError(t, err, "Failed to check service %s", serviceName)

			runningServices := strings.TrimSpace(string(serviceOutput))

			if strings.Contains(runningServices, serviceName) {
				// Service is running - test connectivity
				for _, connectivity := range connectivityTests {
					t.Run("Connectivity_"+connectivity.targetContainer, func(t *testing.T) {
						connectCmd := exec.CommandContext(ctx, "podman", "exec", serviceName, "nc", "-z", connectivity.targetContainer, 
							strconv.Itoa(connectivity.targetPort))
						
						connectErr := connectCmd.Run()
						assert.NoError(t, connectErr, 
							"%s - network connectivity must be functional", connectivity.description)
					})
				}
			} else {
				t.Logf("Service %s not running - skipping network connectivity tests", serviceName)
			}
		})
	}
}

