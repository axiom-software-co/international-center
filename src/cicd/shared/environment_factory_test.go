package shared

import (
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnvironmentFactory validates factory pattern for infrastructure deployment
func TestEnvironmentFactory(t *testing.T) {
	t.Run("CreateDeploymentStrategy_ValidEnvironments", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "environment-factory-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			factory := NewEnvironmentFactory()
			
			// Test factory creates deployment strategy for each valid environment
			devStrategy, err := factory.CreateDeploymentStrategy("development", ctx, cfg)
			require.NoError(t, err, "Factory should create development strategy")
			assert.NotNil(t, devStrategy, "Development strategy should not be nil")
			
			stagingStrategy, err := factory.CreateDeploymentStrategy("staging", ctx, cfg)
			require.NoError(t, err, "Factory should create staging strategy")
			assert.NotNil(t, stagingStrategy, "Staging strategy should not be nil")
			
			productionStrategy, err := factory.CreateDeploymentStrategy("production", ctx, cfg)
			require.NoError(t, err, "Factory should create production strategy")
			assert.NotNil(t, productionStrategy, "Production strategy should not be nil")
			
			return nil
		})
	})
	
	t.Run("CreateDeploymentStrategy_InvalidEnvironment", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "environment-factory-invalid-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			factory := NewEnvironmentFactory()
			
			// Test factory rejects invalid environment
			invalidStrategy, err := factory.CreateDeploymentStrategy("invalid-env", ctx, cfg)
			require.Error(t, err, "Factory should reject invalid environment")
			assert.Nil(t, invalidStrategy, "Invalid strategy should be nil")
			assert.Contains(t, err.Error(), "unknown environment", "Error should indicate unknown environment")
			
			return nil
		})
	})
	
	t.Run("FactoryProducesConsistentStrategies", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "environment-factory-consistency-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			factory := NewEnvironmentFactory()
			
			// Test factory produces consistent strategies for same environment
			strategy1, err := factory.CreateDeploymentStrategy("development", ctx, cfg)
			require.NoError(t, err, "First strategy creation should succeed")
			
			strategy2, err := factory.CreateDeploymentStrategy("development", ctx, cfg)
			require.NoError(t, err, "Second strategy creation should succeed")
			
			// Strategies should be equivalent for same environment
			assert.Equal(t, strategy1.GetEnvironment(), strategy2.GetEnvironment(), "Strategies should have same environment")
			assert.Equal(t, strategy1.GetDeploymentOrder(), strategy2.GetDeploymentOrder(), "Strategies should have same deployment order")
			
			return nil
		})
	})
}

// TestDeploymentStrategyContract validates deployment strategy interface
func TestDeploymentStrategyContract(t *testing.T) {
	t.Run("StrategyImplementsRequiredInterface", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "deployment-strategy-contract-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			factory := NewEnvironmentFactory()
			strategy, err := factory.CreateDeploymentStrategy("development", ctx, cfg)
			require.NoError(t, err, "Strategy creation should succeed")
			
			// Test strategy implements required methods
			environment := strategy.GetEnvironment()
			assert.Equal(t, "development", environment, "Strategy should return correct environment")
			
			deploymentOrder := strategy.GetDeploymentOrder()
			assert.NotEmpty(t, deploymentOrder, "Strategy should define deployment order")
			
			// Verify expected components in deployment order
			expectedComponents := []string{"database", "storage", "vault", "redis", "rabbitmq", "observability", "dapr", "services", "website"}
			assert.Equal(t, expectedComponents, deploymentOrder, "Strategy should define correct deployment order")
			
			// Test strategy can execute deployment
			outputs, err := strategy.Deploy(ctx, cfg)
			require.NoError(t, err, "Strategy deployment should succeed")
			assert.NotNil(t, outputs, "Strategy should return deployment outputs")
			
			return nil
		})
	})
	
	t.Run("EnvironmentSpecificConfiguration", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "environment-config-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			factory := NewEnvironmentFactory()
			
			// Test each environment has specific configuration
			devStrategy, err := factory.CreateDeploymentStrategy("development", ctx, cfg)
			require.NoError(t, err, "Development strategy should be created")
			
			stagingStrategy, err := factory.CreateDeploymentStrategy("staging", ctx, cfg)
			require.NoError(t, err, "Staging strategy should be created")
			
			productionStrategy, err := factory.CreateDeploymentStrategy("production", ctx, cfg)
			require.NoError(t, err, "Production strategy should be created")
			
			// Verify environment-specific behavior
			assert.Equal(t, "development", devStrategy.GetEnvironment())
			assert.Equal(t, "staging", stagingStrategy.GetEnvironment())
			assert.Equal(t, "production", productionStrategy.GetEnvironment())
			
			// All strategies should follow same deployment pattern
			assert.Equal(t, devStrategy.GetDeploymentOrder(), stagingStrategy.GetDeploymentOrder())
			assert.Equal(t, stagingStrategy.GetDeploymentOrder(), productionStrategy.GetDeploymentOrder())
			
			return nil
		})
	})
}

// TestEnvironmentFactoryIntegration validates factory integrates with existing infrastructure
func TestEnvironmentFactoryIntegration(t *testing.T) {
	t.Run("FactoryReplacesCurrentDeploymentLogic", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "factory-integration-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			factory := NewEnvironmentFactory()
			strategy, err := factory.CreateDeploymentStrategy("development", ctx, cfg)
			require.NoError(t, err, "Factory should create strategy")
			
			// Test factory-based deployment produces same outputs as current logic
			outputs, err := strategy.Deploy(ctx, cfg)
			require.NoError(t, err, "Factory-based deployment should succeed")
			
			// Verify all expected outputs exist
			expectedOutputs := []string{
				"environment", "database_connection_string", "storage_connection_string",
				"vault_address", "redis_endpoint", "rabbitmq_endpoint", "grafana_url",
				"dapr_control_plane_url", "public_gateway_url", "admin_gateway_url", "website_url",
			}
			
			for _, expectedOutput := range expectedOutputs {
				assert.Contains(t, outputs, expectedOutput, "Output should contain %s", expectedOutput)
			}
			
			return nil
		})
	})
}