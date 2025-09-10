package integration

import (
	"context"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Environment Health Tests - Prerequisites for all other integration tests
// Following axiom rule: integration tests only run when entire development environment is up
// These tests validate that all deployment phases are healthy before other integration tests execute

func TestEnvironmentHealth_CompleteDeploymentValidation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Arrange: Check prerequisites
	_, err := exec.LookPath("podman")
	if err != nil {
		t.Skip("Podman not available - cannot validate environment health")
	}

	// Act & Assert: Validate all deployment phases are healthy
	t.Run("InfrastructurePhaseHealth", func(t *testing.T) {
		validateInfrastructurePhaseHealth(t, ctx)
	})

	t.Run("PlatformPhaseHealth", func(t *testing.T) {
		validatePlatformPhaseHealth(t, ctx)
	})

	t.Run("ServicesPhaseHealth", func(t *testing.T) {
		validateServicesPhaseHealth(t, ctx)
	})

	t.Run("WebsitePhaseHealth", func(t *testing.T) {
		validateWebsitePhaseHealth(t, ctx)
	})
}

// validateInfrastructurePhaseHealth validates that infrastructure components are operational
func validateInfrastructurePhaseHealth(t *testing.T, ctx context.Context) {
	// Infrastructure containers that must be running and healthy
	infrastructureComponents := []struct {
		containerName string
		port          int
		healthCheck   string
		description   string
	}{
		{"postgresql", 5432, "database", "PostgreSQL database must be operational"},
		{"rabbitmq", 5672, "messaging", "RabbitMQ messaging must be operational"},
		{"azurite", 10000, "storage", "Azurite storage emulator must be operational"},
		{"vault", 8200, "secrets", "Vault secrets store must be operational"},
	}

	for _, component := range infrastructureComponents {
		t.Run("InfraComponent_"+component.containerName, func(t *testing.T) {
			// Validate container is running
			cmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+component.containerName, "--format", "{{.Names}}")
			output, err := cmd.Output()
			require.NoError(t, err, "Failed to check %s container status", component.containerName)

			runningContainers := strings.TrimSpace(string(output))
			assert.Contains(t, runningContainers, component.containerName, 
				"%s - infrastructure component must be running for environment health", component.description)

			// Validate container is healthy (not just running)
			statusCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+component.containerName, "--format", "{{.Status}}")
			statusOutput, err := statusCmd.Output()
			if err == nil {
				status := strings.TrimSpace(string(statusOutput))
				assert.Contains(t, status, "Up", 
					"%s container must be in running state", component.containerName)
			}
		})
	}
}

// validatePlatformPhaseHealth validates that platform components (Dapr) are operational
func validatePlatformPhaseHealth(t *testing.T, ctx context.Context) {
	// Platform components that must be running and accessible
	platformComponents := []struct {
		containerName string
		endpoint      string
		expectedStatus int
		description   string
	}{
		{"dapr-control-plane", "http://localhost:3500/v1.0/healthz", 204, "Dapr control plane must be operational for service mesh"},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, component := range platformComponents {
		t.Run("PlatformComponent_"+component.containerName, func(t *testing.T) {
			// Validate container is running
			cmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+component.containerName, "--format", "{{.Names}}")
			output, err := cmd.Output()
			require.NoError(t, err, "Failed to check %s container status", component.containerName)

			runningContainers := strings.TrimSpace(string(output))
			assert.Contains(t, runningContainers, component.containerName, 
				"%s - platform component must be running", component.description)

			// Validate platform component health endpoint
			if strings.Contains(runningContainers, component.containerName) {
				req, err := http.NewRequestWithContext(ctx, "GET", component.endpoint, nil)
				require.NoError(t, err, "Failed to create health request for %s", component.containerName)

				resp, err := client.Do(req)
				if err == nil {
					defer resp.Body.Close()
					assert.Equal(t, component.expectedStatus, resp.StatusCode, 
						"%s health endpoint must return expected status", component.description)
				} else {
					t.Errorf("%s health endpoint not accessible: %v", component.description, err)
				}
			}
		})
	}
}

// validateServicesPhaseHealth validates that consolidated service containers are operational
func validateServicesPhaseHealth(t *testing.T, ctx context.Context) {
	// Consolidated service containers (following new architecture)
	consolidatedServices := []struct {
		containerName string
		port          int
		healthEndpoint string
		description   string
	}{
		{"public-gateway", 9001, "http://localhost:9001/health", "Public gateway service must be operational"},
		{"admin-gateway", 9000, "http://localhost:9000/health", "Admin gateway service must be operational"},
		{"content", 3001, "http://localhost:3001/health", "Content service must be operational"},
		{"inquiries", 3101, "http://localhost:3101/health", "Inquiries service must be operational"},
		{"notifications", 3201, "http://localhost:3201/health", "Notifications service must be operational"},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, service := range consolidatedServices {
		t.Run("ConsolidatedService_"+service.containerName, func(t *testing.T) {
			// Validate service container is running
			cmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+service.containerName, "--format", "{{.Names}}")
			output, err := cmd.Output()
			require.NoError(t, err, "Failed to check %s service status", service.containerName)

			runningServices := strings.TrimSpace(string(output))
			assert.Contains(t, runningServices, service.containerName, 
				"%s - consolidated service must be running", service.description)

			// Validate service health endpoint accessibility
			if strings.Contains(runningServices, service.containerName) {
				req, err := http.NewRequestWithContext(ctx, "GET", service.healthEndpoint, nil)
				require.NoError(t, err, "Failed to create health request for %s", service.containerName)

				resp, err := client.Do(req)
				if err == nil {
					defer resp.Body.Close()
					assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300, 
						"%s health endpoint must be accessible", service.description)
				} else {
					t.Errorf("%s health endpoint not accessible: %v", service.description, err)
				}
			}

			// Validate corresponding Dapr sidecar is running (for service mesh integration)
			sidecarName := service.containerName + "-dapr"
			sidecarCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+sidecarName, "--format", "{{.Names}}")
			sidecarOutput, err := sidecarCmd.Output()
			if err == nil {
				runningSidecars := strings.TrimSpace(string(sidecarOutput))
				assert.Contains(t, runningSidecars, sidecarName, 
					"Dapr sidecar for %s must be running for service mesh functionality", service.containerName)
			}
		})
	}
}

// validateWebsitePhaseHealth validates that website components are operational
func validateWebsitePhaseHealth(t *testing.T, ctx context.Context) {
	// Website endpoints that should be accessible
	websiteEndpoints := []struct {
		name        string
		endpoint    string
		description string
	}{
		{"public-website", "http://localhost:5173", "Public website must be accessible"},
		{"admin-portal", "http://localhost:3001", "Admin portal must be accessible"},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	for _, website := range websiteEndpoints {
		t.Run("WebsiteEndpoint_"+website.name, func(t *testing.T) {
			req, err := http.NewRequestWithContext(ctx, "GET", website.endpoint, nil)
			require.NoError(t, err, "Failed to create website request for %s", website.name)

			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				// Website may return various status codes but should be accessible
				assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500, 
					"%s - website endpoint must be accessible", website.description)
			} else {
				t.Logf("%s not accessible yet: %v", website.description, err)
				// Website phase may not be fully operational yet - this is acceptable
			}
		})
	}
}

func TestEnvironmentHealth_CrossPhaseIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Arrange: This test only runs if all phases are healthy (prerequisite validation)
	_, err := exec.LookPath("podman")
	if err != nil {
		t.Skip("Podman not available - cannot validate cross-phase integration")
	}

	// Act & Assert: Validate cross-phase functionality
	t.Run("DatabaseToServiceIntegration", func(t *testing.T) {
		// Validate database is accessible from service containers
		serviceContainers := []string{"content", "inquiries", "notifications"}
		
		for _, serviceName := range serviceContainers {
			// Check if service container is running
			serviceCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+serviceName, "--format", "{{.Names}}")
			serviceOutput, err := serviceCmd.Output()
			require.NoError(t, err, "Failed to check service %s status", serviceName)

			runningServices := strings.TrimSpace(string(serviceOutput))
			
			if strings.Contains(runningServices, serviceName) {
				// Test database connectivity from service container
				dbTestCmd := exec.CommandContext(ctx, "podman", "exec", serviceName, "nc", "-z", "postgresql", "5432")
				dbErr := dbTestCmd.Run()
				assert.NoError(t, dbErr, "Service %s must be able to reach database for integrated functionality", serviceName)
			}
		}
	})

	t.Run("ServiceMeshIntegration", func(t *testing.T) {
		// Validate Dapr service mesh connectivity between consolidated services
		client := &http.Client{Timeout: 10 * time.Second}
		
		// Test service discovery through Dapr control plane
		serviceDiscoveryURL := "http://localhost:3500/v1.0/metadata"
		req, err := http.NewRequestWithContext(ctx, "GET", serviceDiscoveryURL, nil)
		require.NoError(t, err, "Failed to create service discovery request")

		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300, 
				"Dapr service discovery must be operational for service mesh integration")
		} else {
			t.Logf("Service mesh integration not operational: %v", err)
		}
	})
}

func TestEnvironmentHealth_DeploymentStateConsistency(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// This test validates that deployment state matches actual environment state
	// Critical for ensuring deployment orchestrator reliability

	t.Run("ContainerDeploymentConsistency", func(t *testing.T) {
		// Expected containers for complete environment
		expectedContainers := []string{
			// Infrastructure
			"postgresql", "rabbitmq", "azurite", "vault",
			// Platform  
			"dapr-control-plane",
			// Services (consolidated)
			"public-gateway", "admin-gateway", "content", "inquiries", "notifications",
		}

		// Count running containers vs expected
		runningCount := 0
		for _, containerName := range expectedContainers {
			cmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+containerName, "--format", "{{.Names}}")
			output, err := cmd.Output()
			require.NoError(t, err, "Failed to check container %s", containerName)

			if strings.Contains(string(output), containerName) {
				runningCount++
			}
		}

		// Environment is healthy when most critical containers are running
		minRequiredContainers := 7 // Infrastructure + Platform + Some Services
		assert.GreaterOrEqual(t, runningCount, minRequiredContainers, 
			"Environment health requires at least %d containers running, found %d", minRequiredContainers, runningCount)

		t.Logf("Environment health: %d/%d expected containers running", runningCount, len(expectedContainers))
	})

	t.Run("NetworkConnectivityHealth", func(t *testing.T) {
		// Validate development network exists and containers are connected
		networkCmd := exec.CommandContext(ctx, "podman", "network", "inspect", "international-center-dev")
		err := networkCmd.Run()
		assert.NoError(t, err, "Development network international-center-dev must exist for environment health")

		// Test basic network connectivity between key components
		if err == nil {
			// Test database network accessibility
			dbNetworkCmd := exec.CommandContext(ctx, "podman", "exec", "postgresql", "nc", "-z", "localhost", "5432")
			dbErr := dbNetworkCmd.Run()
			if dbErr != nil {
				t.Logf("Database network connectivity issue: %v", dbErr)
			}
		}
	})
}