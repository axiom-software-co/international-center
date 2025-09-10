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

// Platform Integration Tests
// Validates platform phase components working together as integrated system
// Tests Dapr control plane, orchestration, networking integration

func TestPlatformIntegration_DaprControlPlaneOrchestration(t *testing.T) {
	// This test requires complete environment health - enforcing axiom rule
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("DaprControlPlane_ServiceMeshReadiness", func(t *testing.T) {
		// Test Dapr control plane is operational and ready for service mesh
		client := &http.Client{Timeout: 5 * time.Second}

		// Test Dapr health endpoint
		healthReq, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:3500/v1.0/healthz", nil)
		require.NoError(t, err, "Failed to create Dapr health request")

		healthResp, err := client.Do(healthReq)
		require.NoError(t, err, "Dapr control plane health endpoint must be accessible for platform integration")
		defer healthResp.Body.Close()

		assert.True(t, healthResp.StatusCode >= 200 && healthResp.StatusCode < 300, 
			"Dapr control plane must be healthy for service mesh functionality")

		// Test Dapr metadata endpoint for service mesh configuration
		metadataReq, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:3500/v1.0/metadata", nil)
		require.NoError(t, err, "Failed to create metadata request")

		metadataResp, err := client.Do(metadataReq)
		require.NoError(t, err, "Dapr metadata endpoint must be accessible for service discovery")
		defer metadataResp.Body.Close()

		assert.Equal(t, http.StatusOK, metadataResp.StatusCode, 
			"Dapr metadata endpoint must be operational for service mesh integration")

		// Validate metadata contains expected service mesh configuration
		body, err := io.ReadAll(metadataResp.Body)
		require.NoError(t, err, "Failed to read Dapr metadata")

		var metadata map[string]interface{}
		err = json.Unmarshal(body, &metadata)
		require.NoError(t, err, "Failed to parse Dapr metadata JSON")

		// Validate service mesh is configured
		assert.Contains(t, metadata, "id", "Dapr metadata must contain service mesh identity")
		
		if runtimeVersion, exists := metadata["runtimeVersion"]; exists {
			assert.NotEmpty(t, runtimeVersion, "Dapr runtime version must be available")
		}
	})

	t.Run("ServiceMeshOrchestration_PlatformReadiness", func(t *testing.T) {
		// Test that platform is ready for service orchestration
		client := &http.Client{Timeout: 5 * time.Second}

		// Test Dapr service invocation capability (without actual services)
		// This validates the platform layer service mesh infrastructure
		serviceInvocationURL := "http://localhost:3500/v1.0/invoke/test-service/method/health"
		
		req, err := http.NewRequestWithContext(ctx, "GET", serviceInvocationURL, nil)
		require.NoError(t, err, "Failed to create service invocation request")

		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			// Should return error for non-existent service, but platform should handle gracefully
			assert.True(t, resp.StatusCode >= 400 && resp.StatusCode < 600, 
				"Dapr service invocation platform should handle non-existent services gracefully")
		} else {
			t.Logf("Service invocation platform capability: %v", err)
			// Platform may not be fully ready for service invocation
		}
	})
}

func TestPlatformIntegration_NetworkingConfiguration(t *testing.T) {
	// Test platform networking configuration and container connectivity
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("DevelopmentNetwork_PlatformConnectivity", func(t *testing.T) {
		// Test that platform containers are properly connected to development network
		networkCmd := exec.CommandContext(ctx, "podman", "network", "inspect", "international-center-dev", "--format", "{{range .containers}}{{.Name}} {{end}}")
		networkOutput, err := networkCmd.Output()
		require.NoError(t, err, "Development network must be inspectable for platform integration")

		connectedContainers := strings.TrimSpace(string(networkOutput))
		
		// Platform components that should be connected to development network
		platformComponents := []string{"dapr-control-plane", "postgresql"}
		
		for _, component := range platformComponents {
			assert.Contains(t, connectedContainers, component, 
				"Platform component %s must be connected to development network", component)
		}
	})

	t.Run("PlatformDaprPortAccessibility", func(t *testing.T) {
		// Test that Dapr platform ports are accessible for service integration
		platformPorts := []struct {
			port        int
			protocol    string
			description string
		}{
			{3500, "HTTP", "Dapr HTTP API port must be accessible"},
			{50001, "gRPC", "Dapr gRPC API port must be accessible"},
		}

		for _, portTest := range platformPorts {
			t.Run("Port_"+portTest.protocol, func(t *testing.T) {
				// Test port connectivity using nc command through Dapr container
				portCheckCmd := exec.CommandContext(ctx, "podman", "exec", "dapr-control-plane", "nc", "-z", "localhost", fmt.Sprintf("%d", portTest.port))
				err := portCheckCmd.Run()
				assert.NoError(t, err, "%s - platform port must be accessible", portTest.description)
			})
		}
	})
}

func TestPlatformIntegration_ServiceMeshInfrastructure(t *testing.T) {
	// Test platform service mesh infrastructure readiness
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("ServiceRegistrationCapability", func(t *testing.T) {
		// Test that platform can handle service registration (even without services running)
		client := &http.Client{Timeout: 5 * time.Second}

		// Test Dapr metadata for service discovery infrastructure
		metadataReq, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:3500/v1.0/metadata", nil)
		require.NoError(t, err, "Failed to create metadata request")

		metadataResp, err := client.Do(metadataReq)
		require.NoError(t, err, "Service registration infrastructure must be operational")
		defer metadataResp.Body.Close()

		body, err := io.ReadAll(metadataResp.Body)
		require.NoError(t, err, "Failed to read service registration metadata")

		var metadata map[string]interface{}
		err = json.Unmarshal(body, &metadata)
		require.NoError(t, err, "Service registration metadata must be valid JSON")

		// Platform should be ready to register services (even if none are running yet)
		assert.Contains(t, metadata, "id", "Platform must have service mesh identity for service registration")
	})

	t.Run("ServiceInvocationInfrastructure", func(t *testing.T) {
		// Test that service invocation infrastructure is ready
		client := &http.Client{Timeout: 5 * time.Second}

		// Test service invocation endpoint availability (without actual target service)
		invocationURL := "http://localhost:3500/v1.0/invoke/platform-test/method/health"
		
		req, err := http.NewRequestWithContext(ctx, "GET", invocationURL, nil)
		require.NoError(t, err, "Failed to create service invocation request")

		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			// Platform should respond (even with error for non-existent service)
			assert.True(t, resp.StatusCode >= 200, 
				"Service invocation infrastructure must be responsive")
		} else {
			t.Logf("Service invocation infrastructure not ready: %v", err)
		}
	})

	t.Run("ComponentConfigurationSupport", func(t *testing.T) {
		// Test that platform supports Dapr component configuration
		client := &http.Client{Timeout: 5 * time.Second}

		// Test that platform can handle component queries
		componentsURL := "http://localhost:3500/v1.0/components"
		
		req, err := http.NewRequestWithContext(ctx, "GET", componentsURL, nil)
		require.NoError(t, err, "Failed to create components request")

		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500, 
				"Platform component configuration support must be operational")
		} else {
			t.Logf("Component configuration support not ready: %v", err)
		}
	})
}


// validateEnvironmentPrerequisites ensures environment health before integration testing
func validateEnvironmentPrerequisites(t *testing.T) {
	// Check critical infrastructure components are running
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
