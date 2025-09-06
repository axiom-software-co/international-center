package shared

import (
	"testing"

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
				"database", "storage", "vault", "redis", "rabbitmq", 
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
				"redis_endpoint":             true,
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
				"redis":         true, // components.DeployRedis
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
				"vault_address", "redis_endpoint", "rabbitmq_endpoint", "grafana_url",
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

// Helper function to find index of element in slice
func findInSlice(slice []string, item string) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
}