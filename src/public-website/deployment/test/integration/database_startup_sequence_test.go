package integration

import (
	"context"
	"net"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED PHASE: These tests validate proper startup sequencing
// Database containers must be running and accepting connections before migrations execute

func TestDatabaseStartupSequence_ContainerRunningBeforeMigrations(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Arrange: Check prerequisites
	_, err := exec.LookPath("podman")
	if err != nil {
		t.Skip("Podman not available - skipping database startup sequence test")
	}

	// Act: Verify database container is running
	cmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name=postgresql", "--format", "{{.Status}}")
	output, err := cmd.Output()
	require.NoError(t, err, "Failed to check PostgreSQL container status")

	status := strings.TrimSpace(string(output))

	// Assert: Database container must be in running state
	assert.Contains(t, status, "Up", "PostgreSQL container must be running before any migrations attempt to connect")

	// Additional validation: Database must be accepting connections
	if strings.Contains(status, "Up") {
		// Test connection to database port
		conn, err := net.DialTimeout("tcp", "localhost:5432", 2*time.Second)
		if err == nil {
			defer conn.Close()
			assert.True(t, true, "Database connection successful")
		} else {
			t.Errorf("Database container is running but not accepting connections on port 5432: %v", err)
		}
	}
}

func TestDatabaseStartupSequence_ReadinessBeforeMigrationExecution(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Arrange: Check prerequisites
	_, err := exec.LookPath("podman")
	if err != nil {
		t.Skip("Podman not available - skipping database readiness test")
	}

	// Act: Test database readiness using pg_isready
	cmd := exec.CommandContext(ctx, "podman", "exec", "postgresql", "pg_isready", "-h", "localhost", "-p", "5432", "-U", "postgres")
	err = cmd.Run()

	// Assert: Database should be ready to accept connections
	assert.NoError(t, err, "PostgreSQL must be ready to accept connections before migrations run")

	// Additional validation: Test that we can actually connect to the database
	connectCmd := exec.CommandContext(ctx, "podman", "exec", "postgresql", "psql", "-h", "localhost", "-U", "postgres", "-d", "international_center_development", "-c", "SELECT 1;")
	err = connectCmd.Run()
	assert.NoError(t, err, "Should be able to execute SQL queries on the database before migrations run")
}

func TestInfrastructureContainerDependencies_StartupOrder(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Arrange: Check prerequisites
	_, err := exec.LookPath("podman")
	if err != nil {
		t.Skip("Podman not available - skipping dependency startup order test")
	}

	// Define infrastructure containers and their dependencies
	containerDependencies := map[string][]string{
		"postgresql":         {}, // No dependencies - should start first
		"vault":              {}, // No dependencies
		"rabbitmq":           {}, // No dependencies  
		"dapr-placement":     {}, // No dependencies
		"dapr-control-plane": {"dapr-placement"}, // Depends on placement service
	}

	// Act & Assert: Validate startup order
	for container, dependencies := range containerDependencies {
		t.Run("StartupOrder_"+container, func(t *testing.T) {
			// Check if main container is running
			cmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+container, "--format", "{{.Names}}")
			output, err := cmd.Output()
			require.NoError(t, err, "Failed to check container %s status", container)

			runningContainers := strings.TrimSpace(string(output))
			assert.Contains(t, runningContainers, container, "Container %s should be running", container)

			// If container is running, validate its dependencies are also running
			if strings.Contains(runningContainers, container) {
				for _, dependency := range dependencies {
					depCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+dependency, "--format", "{{.Names}}")
					depOutput, err := depCmd.Output()
					require.NoError(t, err, "Failed to check dependency %s status", dependency)

					runningDeps := strings.TrimSpace(string(depOutput))
					assert.Contains(t, runningDeps, dependency, 
						"Dependency %s must be running before %s can start", dependency, container)
				}
			}
		})
	}
}

func TestServiceContainerStartupSequence_AfterInfrastructure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Arrange: Check prerequisites
	_, err := exec.LookPath("podman")
	if err != nil {
		t.Skip("Podman not available - skipping service startup sequence test")
	}

	// Infrastructure containers that must be running before services start
	requiredInfrastructure := []string{
		"postgresql",
		"dapr-control-plane",
		"vault",
		"rabbitmq",
	}

	// Service containers that depend on infrastructure
	serviceContainers := []string{
		"public-gateway",
		"admin-gateway",
		"content-news",
		"content-events", 
		"content-research",
	}

	// Act: Check infrastructure containers are running
	for _, infraContainer := range requiredInfrastructure {
		cmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+infraContainer, "--format", "{{.Names}}")
		output, err := cmd.Output()
		require.NoError(t, err, "Failed to check infrastructure container %s", infraContainer)

		runningContainers := strings.TrimSpace(string(output))
		assert.Contains(t, runningContainers, infraContainer, 
			"Infrastructure container %s must be running before services start", infraContainer)
	}

	// Assert: Service containers should only start after infrastructure is ready
	for _, serviceContainer := range serviceContainers {
		t.Run("ServiceAfterInfra_"+serviceContainer, func(t *testing.T) {
			// If service container is running, all infrastructure should already be running
			cmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+serviceContainer, "--format", "{{.Names}}")
			output, err := cmd.Output()
			require.NoError(t, err, "Failed to check service container %s", serviceContainer)

			runningServices := strings.TrimSpace(string(output))
			
			if strings.Contains(runningServices, serviceContainer) {
				// Service is running - validate all required infrastructure is also running
				for _, infraContainer := range requiredInfrastructure {
					infraCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+infraContainer, "--format", "{{.Names}}")
					infraOutput, err := infraCmd.Output()
					require.NoError(t, err, "Failed to check infrastructure dependency %s", infraContainer)

					runningInfra := strings.TrimSpace(string(infraOutput))
					assert.Contains(t, runningInfra, infraContainer, 
						"Infrastructure %s must be running when service %s is running", infraContainer, serviceContainer)
				}
			} else {
				t.Logf("Service container %s not yet running (expected in RED phase)", serviceContainer)
			}
		})
	}
}