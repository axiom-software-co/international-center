package integration

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED PHASE: Service Sidecar Connectivity Tests
// These tests validate that services can connect to their specific Dapr sidecars
// without hardcoded addresses and that sidecars are properly deployed without port conflicts

func TestServiceSidecarConnectivity_ConsolidatedServiceReliability(t *testing.T) {
	// This test requires complete environment health - enforcing axiom rule
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Define consolidated service connectivity requirements
	consolidatedServices := []struct {
		serviceName      string
		sidecarName      string
		servicePort      int
		sidecarPort      int
		expectedDaprHost string
		expectedDaprPort int
	}{
		{"public-gateway", "public-gateway-dapr", 9001, 50010, "public-gateway-dapr", 3500},
		{"admin-gateway", "admin-gateway-dapr", 9000, 50020, "admin-gateway-dapr", 3500},
		{"content", "content-dapr", 3001, 50030, "content-dapr", 3500},
		{"inquiries", "inquiries-dapr", 3101, 50040, "inquiries-dapr", 3500},
		{"notifications", "notifications-dapr", 3201, 50050, "notifications-dapr", 3500},
	}

	// Act & Assert: Validate each consolidated service and sidecar connectivity
	for _, service := range consolidatedServices {
		t.Run("ServiceConnectivity_"+service.serviceName, func(t *testing.T) {
			// Validate service container is running (not exited)
			serviceCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+service.serviceName, "--format", "{{.Status}}")
			serviceOutput, err := serviceCmd.Output()
			require.NoError(t, err, "Failed to check service %s status", service.serviceName)

			serviceStatus := strings.TrimSpace(string(serviceOutput))
			assert.Contains(t, serviceStatus, "Up", 
				"Service %s must be running (not exited) for sidecar connectivity", service.serviceName)
			assert.NotContains(t, serviceStatus, "Exited", 
				"Service %s must not exit due to Dapr connection failures", service.serviceName)

			// Validate corresponding Dapr sidecar is running
			sidecarCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+service.sidecarName, "--format", "{{.Names}}")
			sidecarOutput, err := sidecarCmd.Output()
			require.NoError(t, err, "Failed to check sidecar %s status", service.sidecarName)

			runningSidecars := strings.TrimSpace(string(sidecarOutput))
			assert.Contains(t, runningSidecars, service.sidecarName, 
				"Dapr sidecar %s must be running for service connectivity", service.sidecarName)

			// If both are running, validate connectivity between service and sidecar
			if strings.Contains(serviceStatus, "Up") && strings.Contains(runningSidecars, service.sidecarName) {
				// Test network connectivity from service to its sidecar
				connectCmd := exec.CommandContext(ctx, "podman", "exec", service.serviceName, "nc", "-z", service.expectedDaprHost, fmt.Sprintf("%d", service.expectedDaprPort))
				connectErr := connectCmd.Run()
				assert.NoError(t, connectErr, 
					"Service %s must be able to connect to its sidecar %s on port %d", service.serviceName, service.expectedDaprHost, service.expectedDaprPort)
			}
		})
	}
}

func TestServiceSidecarConnectivity_EnvironmentVariableConfiguration(t *testing.T) {
	// Test that services have proper environment variables for sidecar connectivity
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Expected Dapr environment variables for each service
	expectedDaprConfig := map[string]struct {
		daprHost string
		daprPort string
	}{
		"public-gateway": {"public-gateway-dapr", "3500"},
		"admin-gateway":  {"admin-gateway-dapr", "3500"},
		"content":        {"content-dapr", "3500"},
		"inquiries":      {"inquiries-dapr", "3500"},
		"notifications":  {"notifications-dapr", "3500"},
	}

	// Act & Assert: Validate environment variable configuration
	for serviceName, config := range expectedDaprConfig {
		t.Run("EnvConfig_"+serviceName, func(t *testing.T) {
			// Check if service container is running
			serviceCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+serviceName, "--format", "{{.Names}}")
			serviceOutput, err := serviceCmd.Output()
			require.NoError(t, err, "Failed to check service %s", serviceName)

			runningServices := strings.TrimSpace(string(serviceOutput))
			
			if strings.Contains(runningServices, serviceName) {
				// Service is running - validate environment variables
				envCmd := exec.CommandContext(ctx, "podman", "exec", serviceName, "env")
				envOutput, err := envCmd.Output()
				require.NoError(t, err, "Failed to get environment for %s", serviceName)

				envVars := string(envOutput)
				
				// Validate DAPR_HTTP_ENDPOINT points to specific sidecar
				expectedDaprEndpoint := "DAPR_HTTP_ENDPOINT=http://" + config.daprHost + ":" + config.daprPort
				assert.Contains(t, envVars, expectedDaprEndpoint, 
					"Service %s must have correct DAPR_HTTP_ENDPOINT pointing to its sidecar", serviceName)

				// Validate DAPR_GRPC_ENDPOINT points to specific sidecar  
				expectedDaprGrpcEndpoint := "DAPR_GRPC_ENDPOINT=" + config.daprHost + ":50001"
				assert.Contains(t, envVars, expectedDaprGrpcEndpoint,
					"Service %s must have correct DAPR_GRPC_ENDPOINT pointing to its sidecar", serviceName)

				// Service should NOT have hardcoded localhost addresses
				assert.NotContains(t, envVars, "127.0.0.1:50001", 
					"Service %s must not use hardcoded Dapr client address", serviceName)
			} else {
				t.Logf("Service %s not running - cannot validate environment configuration", serviceName)
			}
		})
	}
}

func TestServiceSidecarConnectivity_PortAllocationConflicts(t *testing.T) {
	// Test that Dapr sidecar ports are allocated without conflicts
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Expected sidecar port allocations (should be unique and not conflict)
	expectedSidecarPorts := map[string]int{
		"public-gateway-dapr": 50010,
		"admin-gateway-dapr":  50020,
		"content-dapr":        50030,
		"inquiries-dapr":      50040,
		"notifications-dapr":  50050,
	}

	// Track allocated ports to detect conflicts
	allocatedPorts := make(map[int]string)

	// Act & Assert: Validate port allocations are unique and accessible
	for sidecarName, expectedPort := range expectedSidecarPorts {
		t.Run("PortAllocation_"+sidecarName, func(t *testing.T) {
			// Check for port conflicts in allocation
			if conflictingService, exists := allocatedPorts[expectedPort]; exists {
				t.Errorf("Port conflict: %s and %s both allocated port %d", sidecarName, conflictingService, expectedPort)
			}
			allocatedPorts[expectedPort] = sidecarName

			// Check if sidecar container is running
			sidecarCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+sidecarName, "--format", "{{.Names}}")
			sidecarOutput, err := sidecarCmd.Output()
			require.NoError(t, err, "Failed to check sidecar %s", sidecarName)

			runningSidecars := strings.TrimSpace(string(sidecarOutput))
			assert.Contains(t, runningSidecars, sidecarName, 
				"Sidecar %s must be running on allocated port %d", sidecarName, expectedPort)

			// If sidecar is running, validate port is accessible
			if strings.Contains(runningSidecars, sidecarName) {
				// Test sidecar health endpoint on allocated port
				client := &http.Client{Timeout: 5 * time.Second}
				sidecarHealthURL := fmt.Sprintf("http://localhost:%d/v1.0/healthz", expectedPort)
				
				req, err := http.NewRequestWithContext(ctx, "GET", sidecarHealthURL, nil)
				require.NoError(t, err, "Failed to create sidecar health request")

				resp, err := client.Do(req)
				if err == nil {
					defer resp.Body.Close()
					assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300, 
						"Sidecar %s must have accessible health endpoint on port %d", sidecarName, expectedPort)
				} else {
					t.Logf("Sidecar %s health endpoint not accessible on port %d: %v", sidecarName, expectedPort, err)
				}
			}
		})
	}
}

func TestServiceSidecarConnectivity_ServiceMeshRegistration(t *testing.T) {
	// Test that services are properly registered with Dapr service mesh
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Expected service registrations in Dapr service mesh
	expectedServiceRegistrations := []string{
		"public-gateway",
		"admin-gateway", 
		"content",
		"inquiries",
		"notifications",
	}

	t.Run("ServiceMeshRegistration_DaprDiscovery", func(t *testing.T) {
		// Test service discovery through Dapr control plane
		client := &http.Client{Timeout: 10 * time.Second}
		metadataReq, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:3500/v1.0/metadata", nil)
		require.NoError(t, err, "Failed to create metadata request")

		metadataResp, err := client.Do(metadataReq)
		require.NoError(t, err, "Dapr metadata endpoint must be accessible for service registration validation")
		defer metadataResp.Body.Close()

		assert.True(t, metadataResp.StatusCode >= 200 && metadataResp.StatusCode < 300, 
			"Dapr service discovery must be operational")

		// Note: In RED phase, services may not be properly registered yet
		// This test validates the infrastructure is ready for service registration
	})

	// Test service invocation capability for each expected service
	for _, serviceName := range expectedServiceRegistrations {
		t.Run("ServiceInvocation_"+serviceName, func(t *testing.T) {
			client := &http.Client{Timeout: 10 * time.Second}
			
			// Test service invocation through Dapr (should work when service is properly registered)
			serviceInvocationURL := fmt.Sprintf("http://localhost:3500/v1.0/invoke/%s/method/health", serviceName)
			
			req, err := http.NewRequestWithContext(ctx, "GET", serviceInvocationURL, nil)
			require.NoError(t, err, "Failed to create service invocation request")

			resp, err := client.Do(req)
			if err != nil {
				t.Logf("Service %s not accessible through service mesh (expected in RED phase): %v", serviceName, err)
				t.Fail() // Should fail until service mesh connectivity is fixed
				return
			}
			defer resp.Body.Close()

			// Service should be accessible through Dapr service invocation
			assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300, 
				"Service %s must be accessible through Dapr service invocation", serviceName)
		})
	}
}

func TestServiceSidecarConnectivity_ContainerNetworkingIntegration(t *testing.T) {
	// Test that service containers and sidecars are properly connected to development network
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("NetworkConnectivity_ServiceAndSidecarPairs", func(t *testing.T) {
		// Test network connectivity between service/sidecar pairs
		serviceSidecarPairs := []struct {
			serviceName string
			sidecarName string
		}{
			{"public-gateway", "public-gateway-dapr"},
			{"admin-gateway", "admin-gateway-dapr"},
			{"content", "content-dapr"},
			{"inquiries", "inquiries-dapr"},
			{"notifications", "notifications-dapr"},
		}

		for _, pair := range serviceSidecarPairs {
			t.Run("NetworkPair_"+pair.serviceName, func(t *testing.T) {
				// Check if both service and sidecar are connected to development network
				networkCmd := exec.CommandContext(ctx, "podman", "network", "inspect", "international-center-dev", "--format", "{{range .containers}}{{.Name}} {{end}}")
				networkOutput, err := networkCmd.Output()
				require.NoError(t, err, "Failed to inspect development network")

				connectedContainers := strings.TrimSpace(string(networkOutput))
				
				// Both service and sidecar must be connected to development network
				assert.Contains(t, connectedContainers, pair.serviceName, 
					"Service %s must be connected to development network", pair.serviceName)
				assert.Contains(t, connectedContainers, pair.sidecarName, 
					"Sidecar %s must be connected to development network", pair.sidecarName)

				// If both are connected, test inter-container communication
				if strings.Contains(connectedContainers, pair.serviceName) && strings.Contains(connectedContainers, pair.sidecarName) {
					// Test that service can reach its sidecar through container networking
					pingCmd := exec.CommandContext(ctx, "podman", "exec", pair.serviceName, "nc", "-z", pair.sidecarName, "3500")
					pingErr := pingCmd.Run()
					assert.NoError(t, pingErr, 
						"Service %s must be able to reach sidecar %s through container networking", pair.serviceName, pair.sidecarName)
				}
			})
		}
	})

	t.Run("CrossServiceConnectivity_ThroughNetwork", func(t *testing.T) {
		// Test that services can reach each other through development network (for service mesh)
		services := []string{"public-gateway", "admin-gateway", "content", "inquiries", "notifications"}
		
		for i, sourceService := range services {
			for j, targetService := range services {
				if i != j { // Don't test service connecting to itself
					t.Run(fmt.Sprintf("CrossService_%s_to_%s", sourceService, targetService), func(t *testing.T) {
						// Check if both services are running
						sourceCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+sourceService, "--format", "{{.Names}}")
						sourceOutput, err := sourceCmd.Output()
						require.NoError(t, err, "Failed to check source service %s", sourceService)

						targetCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+targetService, "--format", "{{.Names}}")
						targetOutput, err := targetCmd.Output()
						require.NoError(t, err, "Failed to check target service %s", targetService)

						sourceRunning := strings.Contains(string(sourceOutput), sourceService)
						targetRunning := strings.Contains(string(targetOutput), targetService)

						if sourceRunning && targetRunning {
							// Test cross-service connectivity through network
							crossConnectCmd := exec.CommandContext(ctx, "podman", "exec", sourceService, "nc", "-z", targetService, "8080")
							crossConnectErr := crossConnectCmd.Run()
							assert.NoError(t, crossConnectErr, 
								"Service %s must be able to reach service %s through container network for service mesh", sourceService, targetService)
						} else {
							t.Logf("Cross-service connectivity test skipped: %s running=%v, %s running=%v", 
								sourceService, sourceRunning, targetService, targetRunning)
						}
					})
				}
			}
		}
	})
}

func TestServiceSidecarConnectivity_DaprClientConfiguration(t *testing.T) {
	// Test that service containers have correct Dapr client configuration
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Services that should have Dapr client configuration
	daprEnabledServices := []struct {
		serviceName string
		appID       string
	}{
		{"public-gateway", "public-gateway"},
		{"admin-gateway", "admin-gateway"},
		{"content", "content"},
		{"inquiries", "inquiries"},
		{"notifications", "notifications"},
	}

	// Act & Assert: Validate Dapr client configuration
	for _, service := range daprEnabledServices {
		t.Run("DaprClientConfig_"+service.serviceName, func(t *testing.T) {
			// Check if service is running
			serviceCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+service.serviceName, "--format", "{{.Names}}")
			serviceOutput, err := serviceCmd.Output()
			require.NoError(t, err, "Failed to check service %s", service.serviceName)

			runningServices := strings.TrimSpace(string(serviceOutput))

			if strings.Contains(runningServices, service.serviceName) {
				// Check service logs for Dapr client initialization
				logsCmd := exec.CommandContext(ctx, "podman", "logs", "--tail", "20", service.serviceName)
				logsOutput, err := logsCmd.Output()
				require.NoError(t, err, "Failed to get logs for %s", service.serviceName)

				logs := string(logsOutput)
				
				// Service should not fail with Dapr connection errors
				assert.NotContains(t, logs, "context deadline exceeded", 
					"Service %s must not fail with Dapr connection timeouts", service.serviceName)
				assert.NotContains(t, logs, "Failed to create Dapr client", 
					"Service %s must successfully create Dapr client connection", service.serviceName)

				// Service should successfully initialize Dapr client
				if !strings.Contains(logs, "context deadline exceeded") && !strings.Contains(logs, "Failed to create Dapr client") {
					assert.Contains(t, logs, "dapr client initializing", 
						"Service %s should show Dapr client initialization", service.serviceName)
				}
			} else {
				t.Logf("Service %s not running - cannot validate Dapr client configuration", service.serviceName)
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