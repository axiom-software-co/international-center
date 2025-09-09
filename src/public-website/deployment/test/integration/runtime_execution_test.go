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

// RED PHASE: These tests validate actual container execution, not just Pulumi configuration
// They should FAIL initially because containers are configured but not running

func TestActualPodmanContainerExecution_DatabaseContainer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Arrange: Check if we have Podman available
	_, err := exec.LookPath("podman")
	if err != nil {
		t.Skip("Podman not available - skipping container execution test")
	}

	// Act: Check if PostgreSQL container is actually running via Podman
	cmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name=postgresql", "--format", "{{.Names}}")
	output, err := cmd.Output()
	require.NoError(t, err, "Failed to check Podman container status")

	runningContainers := strings.TrimSpace(string(output))

	// Assert: PostgreSQL container should be running
	assert.Contains(t, runningContainers, "postgresql", "PostgreSQL container is not running in Podman")

	// Additional validation: Container should be accessible on port 5432
	if strings.Contains(runningContainers, "postgresql") {
		// Test database connectivity
		portCheckCmd := exec.CommandContext(ctx, "podman", "exec", "postgresql", "pg_isready", "-h", "localhost", "-p", "5432")
		err := portCheckCmd.Run()
		assert.NoError(t, err, "PostgreSQL container is running but not accepting connections")
	}
}

func TestActualPodmanContainerExecution_DaprControlPlane(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Arrange: Check if we have Podman available
	_, err := exec.LookPath("podman")
	if err != nil {
		t.Skip("Podman not available - skipping container execution test")
	}

	// Act: Check if Dapr control plane container is actually running
	cmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name=dapr-control-plane", "--format", "{{.Names}}")
	output, err := cmd.Output()
	require.NoError(t, err, "Failed to check Dapr control plane container status")

	runningContainers := strings.TrimSpace(string(output))

	// Assert: Dapr control plane should be running
	assert.Contains(t, runningContainers, "dapr-control-plane", "Dapr control plane container is not running in Podman")

	// Additional validation: Dapr health endpoint should be accessible
	if strings.Contains(runningContainers, "dapr-control-plane") {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get("http://localhost:3500/v1.0/healthz")
		if err == nil {
			defer resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode, "Dapr control plane health endpoint not responding correctly")
		} else {
			t.Logf("Dapr health endpoint not yet accessible: %v", err)
		}
	}
}

func TestActualPodmanContainerExecution_AllServiceContainers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Arrange: Check if we have Podman available
	_, err := exec.LookPath("podman")
	if err != nil {
		t.Skip("Podman not available - skipping container execution test")
	}

	// Define expected service containers based on deployment configuration
	expectedContainers := []string{
		"public-gateway",
		"admin-gateway", 
		"content-news",
		"content-events",
		"content-research",
		"inquiries-business",
		"inquiries-donations",
		"inquiries-media", 
		"inquiries-volunteers",
		"notification-service",
	}

	// Act & Assert: Check each service container is running
	for _, containerName := range expectedContainers {
		t.Run("Container_"+containerName, func(t *testing.T) {
			cmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+containerName, "--format", "{{.Names}}")
			output, err := cmd.Output()
			require.NoError(t, err, "Failed to check container status for %s", containerName)

			runningContainers := strings.TrimSpace(string(output))
			assert.Contains(t, runningContainers, containerName, "Service container %s is not running in Podman", containerName)

			// Test corresponding Dapr sidecar is also running
			sidecarName := containerName + "-dapr"
			sidecarCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+sidecarName, "--format", "{{.Names}}")
			sidecarOutput, err := sidecarCmd.Output()
			if err == nil {
				runningSidecars := strings.TrimSpace(string(sidecarOutput))
				assert.Contains(t, runningSidecars, sidecarName, "Dapr sidecar %s is not running for service %s", sidecarName, containerName)
			}
		})
	}
}

func TestActualServiceHealthEndpoints_HTTPAccessibility(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Arrange: Define expected service health endpoints from deployment outputs
	serviceEndpoints := map[string]string{
		"public_gateway":       "http://127.0.0.1:9001/health",
		"admin_gateway":        "http://127.0.0.1:9000/health", 
		"content_service":      "http://localhost:8001/health",
		"inquiries_service":    "http://localhost:8002/health",
		"notifications_service": "http://localhost:8003/health",
		"dapr_control_plane":   "http://localhost:3500/v1.0/healthz",
	}

	client := &http.Client{Timeout: 5 * time.Second}

	// Act & Assert: Test each service endpoint is accessible
	for serviceName, endpoint := range serviceEndpoints {
		t.Run("HealthEndpoint_"+serviceName, func(t *testing.T) {
			req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
			require.NoError(t, err, "Failed to create HTTP request for %s", serviceName)

			resp, err := client.Do(req)
			if err != nil {
				t.Logf("Service %s health endpoint %s not accessible: %v", serviceName, endpoint, err)
				// This should fail in RED phase - services aren't running yet
				t.Fail()
				return
			}
			defer resp.Body.Close()

			// Assert: Health endpoint should return success status
			assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300, 
				"Service %s health endpoint %s returned status %d", serviceName, endpoint, resp.StatusCode)
		})
	}
}

func TestContainerStartupSequencing_DatabaseBeforeMigrations(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Arrange: Check if we have Podman available
	_, err := exec.LookPath("podman")
	if err != nil {
		t.Skip("Podman not available - skipping container sequencing test")
	}

	// Act: Check if database container is running
	dbCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name=postgresql", "--format", "{{.Names}}")
	dbOutput, err := dbCmd.Output()
	require.NoError(t, err, "Failed to check database container status")

	dbRunning := strings.TrimSpace(string(dbOutput))
	
	// Assert: Database should be running before migrations are attempted
	assert.Contains(t, dbRunning, "postgresql", "Database container must be running before migrations can execute")

	// If database is running, validate it accepts connections before any migration logic runs
	if strings.Contains(dbRunning, "postgresql") {
		// Test database readiness
		readinessCmd := exec.CommandContext(ctx, "podman", "exec", "postgresql", "pg_isready", "-h", "localhost", "-p", "5432")
		err := readinessCmd.Run()
		assert.NoError(t, err, "Database container is running but not ready to accept connections - migrations would fail")
	}
}

func TestActualNetworkConnectivity_ServiceToServiceCommunication(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Arrange: Check if we have Podman available and network exists
	_, err := exec.LookPath("podman")
	if err != nil {
		t.Skip("Podman not available - skipping network connectivity test")
	}

	// Act: Check if international-center-dev network exists
	networkCmd := exec.CommandContext(ctx, "podman", "network", "inspect", "international-center-dev")
	err = networkCmd.Run()

	// Assert: Network should exist for service communication
	assert.NoError(t, err, "international-center-dev network should exist for service-to-service communication")

	// Act: Check if containers are connected to the network
	containersCmd := exec.CommandContext(ctx, "podman", "network", "inspect", "international-center-dev", "--format", "{{range .containers}}{{.Name}} {{end}}")
	output, err := containersCmd.Output()
	if err == nil {
		connectedContainers := strings.TrimSpace(string(output))
		
		// Assert: Key containers should be connected to the network
		expectedInNetwork := []string{"postgresql", "dapr-control-plane", "public-gateway", "admin-gateway"}
		for _, container := range expectedInNetwork {
			assert.Contains(t, connectedContainers, container, "Container %s should be connected to international-center-dev network", container)
		}
	} else {
		t.Logf("Could not inspect network containers: %v", err)
		t.Fail()
	}
}