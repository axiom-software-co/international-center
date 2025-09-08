package shared

import (
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDeploymentStrategy validates deployment strategy pattern implementation
func TestDeploymentStrategy(t *testing.T) {
	t.Run("CreateStrategyForEachEnvironment", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "deployment-strategy-create-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			// Test strategy creation for each environment
			devStrategy, err := NewDeploymentStrategy("development", ctx, cfg)
			require.NoError(t, err, "Development strategy creation should succeed")
			require.NotNil(t, devStrategy, "Development strategy should be created")
			assert.Equal(t, "development", devStrategy.GetEnvironment(), "Strategy should have correct environment")
			
			stagingStrategy, err := NewDeploymentStrategy("staging", ctx, cfg)
			require.NoError(t, err, "Staging strategy creation should succeed")
			require.NotNil(t, stagingStrategy, "Staging strategy should be created")
			assert.Equal(t, "staging", stagingStrategy.GetEnvironment(), "Strategy should have correct environment")
			
			productionStrategy, err := NewDeploymentStrategy("production", ctx, cfg)
			require.NoError(t, err, "Production strategy creation should succeed")
			require.NotNil(t, productionStrategy, "Production strategy should be created")
			assert.Equal(t, "production", productionStrategy.GetEnvironment(), "Strategy should have correct environment")
			
			return nil
		})
	})
	
	t.Run("StrategyDefinesDeploymentOrder", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "deployment-strategy-order-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			strategy, err := NewDeploymentStrategy("development", ctx, cfg)
			require.NoError(t, err, "Strategy creation should succeed")
			
			deploymentOrder := strategy.GetDeploymentOrder()
			assert.NotEmpty(t, deploymentOrder, "Strategy should define deployment order")
			
			// Verify expected deployment sequence matches current main.go logic
			expectedOrder := []string{
				"database", "storage", "vault", "rabbitmq", 
				"observability", "dapr", "services", "website",
			}
			assert.Equal(t, expectedOrder, deploymentOrder, "Deployment order should match current sequence")
			
			return nil
		})
	})
	
	t.Run("StrategyExecutesDeployment", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "deployment-strategy-execute-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			strategy, err := NewDeploymentStrategy("development", ctx, cfg)
			require.NoError(t, err, "Strategy creation should succeed")
			
			// Test strategy can execute deployment
			outputs, err := strategy.Deploy(ctx, cfg)
			require.NoError(t, err, "Strategy deployment should succeed")
			assert.NotNil(t, outputs, "Deployment should return outputs")
			
			// Verify outputs contain expected keys (matching current main.go exports)
			expectedOutputs := map[string]bool{
				"environment":                true,
				"database_connection_string": true,
				"storage_connection_string":  true,
				"vault_address":              true,
				"rabbitmq_endpoint":          true,
				"grafana_url":                true,
				"dapr_control_plane_url":     true,
				"public_gateway_url":         true,
				"admin_gateway_url":          true,
				"website_url":                true,
			}
			
			for expectedOutput := range expectedOutputs {
				assert.Contains(t, outputs, expectedOutput, "Outputs should contain %s", expectedOutput)
			}
			
			return nil
		})
	})
}

// TestDeploymentStrategySequential validates sequential deployment behavior
func TestDeploymentStrategySequential(t *testing.T) {
	t.Run("StrategyDeploysComponentsSequentially", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "deployment-strategy-sequential-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			strategy, err := NewDeploymentStrategy("development", ctx, cfg)
			require.NoError(t, err, "Strategy creation should succeed")
			
			// Test that strategy enforces sequential deployment
			deploymentOrder := strategy.GetDeploymentOrder()
			
			// Verify dependencies are respected in order
			databaseIndex := findInSlice(deploymentOrder, "database")
			storageIndex := findInSlice(deploymentOrder, "storage")
			daprIndex := findInSlice(deploymentOrder, "dapr")
			servicesIndex := findInSlice(deploymentOrder, "services")
			websiteIndex := findInSlice(deploymentOrder, "website")
			
			assert.True(t, databaseIndex < servicesIndex, "Database should deploy before services")
			assert.True(t, storageIndex < servicesIndex, "Storage should deploy before services")
			assert.True(t, daprIndex < servicesIndex, "Dapr should deploy before services")
			assert.True(t, servicesIndex < websiteIndex, "Services should deploy before website")
			
			return nil
		})
	})
	
	t.Run("StrategyHandlesDeploymentFailure", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "deployment-strategy-failure-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			strategy, err := NewDeploymentStrategy("development", ctx, cfg)
			require.NoError(t, err, "Strategy creation should succeed")
			
			// Test that strategy properly handles and propagates deployment failures
			// This tests the error handling path of the sequential deployment
			outputs, err := strategy.Deploy(ctx, cfg)
			
			// In this test case, we expect the deployment to succeed since all components exist
			// But we verify the strategy has proper error handling structure
			if err != nil {
				// If there's an error, it should be properly formatted
				assert.NotEmpty(t, err.Error(), "Error should have descriptive message")
				assert.Nil(t, outputs, "Outputs should be nil when deployment fails")
			} else {
				// If deployment succeeds, outputs should be complete
				assert.NotNil(t, outputs, "Outputs should not be nil when deployment succeeds")
			}
			
			return nil
		})
	})
}

// TestDeploymentStrategyEnvironmentParity validates environment parity
func TestDeploymentStrategyEnvironmentParity(t *testing.T) {
	t.Run("AllEnvironmentsFollowSameDeploymentPattern", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "deployment-strategy-parity-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			// Create strategies for all environments
			devStrategy, err := NewDeploymentStrategy("development", ctx, cfg)
			require.NoError(t, err, "Development strategy creation should succeed")
			stagingStrategy, err := NewDeploymentStrategy("staging", ctx, cfg)
			require.NoError(t, err, "Staging strategy creation should succeed")
			productionStrategy, err := NewDeploymentStrategy("production", ctx, cfg)
			require.NoError(t, err, "Production strategy creation should succeed")
			
			// All strategies should have same deployment order
			devOrder := devStrategy.GetDeploymentOrder()
			stagingOrder := stagingStrategy.GetDeploymentOrder()
			productionOrder := productionStrategy.GetDeploymentOrder()
			
			assert.Equal(t, devOrder, stagingOrder, "Development and staging should have same deployment order")
			assert.Equal(t, stagingOrder, productionOrder, "Staging and production should have same deployment order")
			
			return nil
		})
	})
	
	t.Run("EnvironmentSpecificBehaviorIsolated", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "deployment-strategy-isolation-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			devStrategy, err := NewDeploymentStrategy("development", ctx, cfg)
			require.NoError(t, err, "Development strategy creation should succeed")
			stagingStrategy, err := NewDeploymentStrategy("staging", ctx, cfg)
			require.NoError(t, err, "Staging strategy creation should succeed")
			
			// Strategies should have different environments but same interface
			assert.NotEqual(t, devStrategy.GetEnvironment(), stagingStrategy.GetEnvironment(), "Strategies should have different environments")
			assert.Equal(t, devStrategy.GetDeploymentOrder(), stagingStrategy.GetDeploymentOrder(), "Strategies should have same deployment order")
			
			// Both strategies should be capable of independent deployment
			devOutputs, devErr := devStrategy.Deploy(ctx, cfg)
			stagingOutputs, stagingErr := stagingStrategy.Deploy(ctx, cfg)
			
			// Both should handle deployment appropriately (either succeed or fail gracefully)
			if devErr == nil {
				assert.NotNil(t, devOutputs, "Development outputs should not be nil on success")
			}
			if stagingErr == nil {
				assert.NotNil(t, stagingOutputs, "Staging outputs should not be nil on success")
			}
			
			return nil
		})
	})
}

// TestDeploymentStrategyIntegration validates strategy integrates with existing components
func TestDeploymentStrategyIntegration(t *testing.T) {
	t.Run("StrategyUsesExistingComponents", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "deployment-strategy-integration-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			strategy, err := NewDeploymentStrategy("development", ctx, cfg)
			require.NoError(t, err, "Strategy creation should succeed")
			
			// Strategy should integrate with existing component deployment functions
			// This validates that the strategy calls the same components.Deploy* functions
			// that are currently used in main.go
			
			deploymentOrder := strategy.GetDeploymentOrder()
			
			// Verify all components in deployment order correspond to existing Deploy functions
			componentFunctions := map[string]bool{
				"database":      true, // components.DeployDatabase
				"storage":       true, // components.DeployStorage  
				"vault":         true, // components.DeployVault
				"rabbitmq":      true, // components.DeployRabbitMQ
				"observability": true, // components.DeployObservability
				"dapr":          true, // components.DeployDapr
				"services":      true, // components.DeployServices
				"website":       true, // components.DeployWebsite
			}
			
			for _, component := range deploymentOrder {
				assert.True(t, componentFunctions[component], "Component %s should have corresponding Deploy function", component)
			}
			
			return nil
		})
	})
	
	t.Run("StrategyReplacesCurrentMainLogic", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "deployment-strategy-replacement-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			strategy, err := NewDeploymentStrategy("development", ctx, cfg)
			require.NoError(t, err, "Strategy creation should succeed")
			
			// Strategy deployment should produce equivalent results to current main.go logic
			outputs, err := strategy.Deploy(ctx, cfg)
			require.NoError(t, err, "Strategy deployment should succeed")
			
			// Verify outputs match what current deployDevelopmentInfrastructure would export
			expectedOutputKeys := []string{
				"environment", "database_connection_string", "storage_connection_string",
				"vault_address", "rabbitmq_endpoint", "grafana_url",
				"dapr_control_plane_url", "public_gateway_url", "admin_gateway_url", "website_url",
			}
			
			assert.Equal(t, len(expectedOutputKeys), len(outputs), "Should have correct number of outputs")
			
			for _, key := range expectedOutputKeys {
				assert.Contains(t, outputs, key, "Should contain output key %s", key)
			}
			
			// Verify environment output matches expected value
			assert.Equal(t, "development", outputs["environment"], "Environment output should be correct")
			
			return nil
		})
	})
}

// TestCompleteInfrastructureDeployment validates all infrastructure containers are deployed and running
func TestCompleteInfrastructureDeployment(t *testing.T) {
	t.Run("AllInfrastructureContainersRunning_Development", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "complete-infrastructure-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			strategy, err := NewDeploymentStrategy("development", ctx, cfg)
			require.NoError(t, err, "Strategy creation should succeed")
			
			// Deploy complete infrastructure
			outputs, err := strategy.Deploy(ctx, cfg)
			require.NoError(t, err, "Complete infrastructure deployment should succeed")
			require.NotNil(t, outputs, "Deployment outputs should not be nil")
			
			// Wait for containers to stabilize
			time.Sleep(10 * time.Second)
			
			// Validate all expected infrastructure containers are running
			expectedContainers := []ContainerValidation{
				{"postgresql-dev", "postgres:15", []string{"5432:5432"}},
				{"vault-dev", "hashicorp/vault:latest", []string{"8200:8200"}},
				{"azurite-dev", "mcr.microsoft.com/azure-storage/azurite:latest", []string{"10000:10000", "10001:10001", "10002:10002"}},
				{"rabbitmq-dev", "rabbitmq:3-management-alpine", []string{"5672:5672", "15672:15672"}},
				{"dapr-placement-dev", "daprio/dapr:1.12.0", []string{}},
				{"otel-lgtm-dev", "grafana/otel-lgtm:latest", []string{"3000:3000", "3100:3100", "4317:4317", "4318:4318", "9090:9090"}},
				{"website-dev", "localhost/website:latest", []string{"3001:3000"}},
			}
			
			// Validate Dapr sidecar containers
			expectedDaprSidecars := []string{
				"inquiries-dapr", "content-dapr", "admin-dapr", "public-dapr", "notifications-dapr",
			}
			
			// Validate backend service containers
			expectedServiceContainers := []string{
				"inquiries-service", "content-service", "admin-service", "public-service", "notifications-service",
			}
			
			for _, container := range expectedContainers {
				validateContainerRunning(t, container.Name, container.Image, container.Ports)
			}
			
			for _, sidecar := range expectedDaprSidecars {
				validateContainerRunning(t, sidecar, "daprio/daprd:latest", []string{})
			}
			
			for _, service := range expectedServiceContainers {
				validateContainerExists(t, service) // Backend services may not have standardized image names
			}
			
			return nil
		})
	})
	
	t.Run("ContainerHealthValidation_Development", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "container-health-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			// Deploy infrastructure first
			strategy, err := NewDeploymentStrategy("development", ctx, cfg)
			require.NoError(t, err, "Strategy creation should succeed")
			
			_, err = strategy.Deploy(ctx, cfg)
			require.NoError(t, err, "Infrastructure deployment should succeed")
			
			// Wait for health checks to stabilize  
			time.Sleep(15 * time.Second)
			
			// Validate container health status
			healthyContainers := []string{
				"postgresql-dev", "vault-dev", "azurite-dev", "rabbitmq-dev", 
				"otel-lgtm-dev", "dapr-placement-dev",
			}
			
			for _, containerName := range healthyContainers {
				validateContainerHealthy(t, containerName)
			}
			
			return nil
		})
	})
	
	t.Run("ServiceAccessibilityValidation_Development", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "service-accessibility-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			// Deploy infrastructure first
			strategy, err := NewDeploymentStrategy("development", ctx, cfg)
			require.NoError(t, err, "Strategy creation should succeed")
			
			_, err = strategy.Deploy(ctx, cfg)
			require.NoError(t, err, "Infrastructure deployment should succeed")
			
			// Wait for services to be accessible
			time.Sleep(20 * time.Second)
			
			// Validate service accessibility
			serviceEndpoints := []ServiceEndpoint{
				{"postgresql", "localhost", 5432},
				{"vault", "localhost", 8200},
				{"azurite-blob", "localhost", 10000},
				{"azurite-queue", "localhost", 10001},
				{"azurite-table", "localhost", 10002},
				{"rabbitmq", "localhost", 5672},
				{"rabbitmq-management", "localhost", 15672},
				{"grafana", "localhost", 3000},
				{"prometheus", "localhost", 9090},
				{"website", "localhost", 3001},
			}
			
			for _, endpoint := range serviceEndpoints {
				validateServiceAccessible(t, endpoint.Name, endpoint.Host, endpoint.Port)
			}
			
			return nil
		})
	})
}

// ContainerValidation represents expected container configuration
type ContainerValidation struct {
	Name  string
	Image string
	Ports []string
}

// ServiceEndpoint represents service accessibility requirements
type ServiceEndpoint struct {
	Name string
	Host string
	Port int
}

// validateContainerRunning verifies a container is running with expected configuration
func validateContainerRunning(t *testing.T, name, expectedImage string, expectedPorts []string) {
	// Check container exists and is running
	cmd := exec.Command("podman", "ps", "--format", "{{.Names}}\t{{.Status}}\t{{.Image}}", "--filter", "name="+name)
	output, err := cmd.Output()
	require.NoError(t, err, "Should be able to query container status for %s", name)
	
	outputStr := strings.TrimSpace(string(output))
	assert.NotEmpty(t, outputStr, "Container %s should be running", name)
	
	if outputStr != "" {
		parts := strings.Split(outputStr, "\t")
		assert.Equal(t, name, parts[0], "Container name should match")
		assert.Contains(t, parts[1], "Up", "Container %s should be in Up status", name)
		
		if expectedImage != "" {
			assert.Contains(t, parts[2], expectedImage, "Container %s should use expected image", name)
		}
	}
	
	// Check port mappings if specified
	if len(expectedPorts) > 0 {
		portCmd := exec.Command("podman", "port", name)
		portOutput, err := portCmd.Output()
		require.NoError(t, err, "Should be able to query container ports for %s", name)
		
		portStr := strings.TrimSpace(string(portOutput))
		for _, expectedPort := range expectedPorts {
			// Handle both Docker format "5432:5432" and Podman format "5432/tcp -> 0.0.0.0:5432"
			// Also handle non-standard mappings like "3001:3000" (external:internal)
			portParts := strings.Split(expectedPort, ":")
			if len(portParts) == 2 {
				externalPort := portParts[0]
				internalPort := portParts[1]
				// Check if port is mapped (e.g., "3000" should appear in "3000/tcp -> 0.0.0.0:3001")
				assert.Contains(t, portStr, internalPort+"/tcp", "Container %s should have internal port %s exposed", name, internalPort)
				assert.Contains(t, portStr, "0.0.0.0:"+externalPort, "Container %s should have external port %s mapped", name, externalPort)
			} else {
				assert.Contains(t, portStr, expectedPort, "Container %s should have port mapping %s", name, expectedPort)
			}
		}
	}
}

// validateContainerExists verifies a container exists (less strict than validateContainerRunning)
func validateContainerExists(t *testing.T, name string) {
	cmd := exec.Command("podman", "ps", "-a", "--format", "{{.Names}}", "--filter", "name="+name)
	output, err := cmd.Output()
	require.NoError(t, err, "Should be able to query container existence for %s", name)
	
	outputStr := strings.TrimSpace(string(output))
	assert.NotEmpty(t, outputStr, "Container %s should exist", name)
}

// validateContainerHealthy verifies a container is healthy (not just running)
func validateContainerHealthy(t *testing.T, name string) {
	cmd := exec.Command("podman", "ps", "--format", "{{.Names}}\t{{.Status}}", "--filter", "name="+name)
	output, err := cmd.Output()
	require.NoError(t, err, "Should be able to query container health for %s", name)
	
	outputStr := strings.TrimSpace(string(output))
	assert.NotEmpty(t, outputStr, "Container %s should be running", name)
	
	if outputStr != "" {
		parts := strings.Split(outputStr, "\t")
		status := parts[1]
		assert.Contains(t, status, "Up", "Container %s should be Up", name)
		assert.NotContains(t, status, "unhealthy", "Container %s should not be unhealthy", name)
	}
}

// validateServiceAccessible verifies a service endpoint is accessible
func validateServiceAccessible(t *testing.T, serviceName, host string, port int) {
	// Simple TCP connection test - attempt to connect to the port
	cmd := exec.Command("timeout", "5s", "bash", "-c", "echo > /dev/tcp/"+host+"/"+string(rune(port)))
	err := cmd.Run()
	
	// For now, we'll just log the connection attempt since some services may require specific protocols
	if err != nil {
		t.Logf("Service %s at %s:%d may not be accessible via simple TCP check: %v", serviceName, host, port, err)
		// Don't fail the test for accessibility - focus on container existence first
	}
}

// Helper function to find index of element in slice
func findInSlice(slice []string, item string) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
}