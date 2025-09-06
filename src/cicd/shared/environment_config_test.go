package shared

import (
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnvironmentConfiguration validates environment-specific configuration management
func TestEnvironmentConfiguration(t *testing.T) {
	t.Run("LoadValidEnvironmentConfiguration", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "environment-config-valid-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			// Test loading valid environment configurations
			devConfig, err := LoadEnvironmentConfiguration("development", cfg)
			require.NoError(t, err, "Development config should load successfully")
			assert.NotNil(t, devConfig, "Development config should not be nil")
			assert.Equal(t, "development", devConfig.Environment, "Config should have correct environment")
			
			stagingConfig, err := LoadEnvironmentConfiguration("staging", cfg)
			require.NoError(t, err, "Staging config should load successfully")
			assert.NotNil(t, stagingConfig, "Staging config should not be nil")
			assert.Equal(t, "staging", stagingConfig.Environment, "Config should have correct environment")
			
			productionConfig, err := LoadEnvironmentConfiguration("production", cfg)
			require.NoError(t, err, "Production config should load successfully")
			assert.NotNil(t, productionConfig, "Production config should not be nil")
			assert.Equal(t, "production", productionConfig.Environment, "Config should have correct environment")
			
			return nil
		})
	})
	
	t.Run("RejectInvalidEnvironmentConfiguration", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "environment-config-invalid-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			// Test rejection of invalid environment
			invalidConfig, err := LoadEnvironmentConfiguration("invalid-env", cfg)
			require.Error(t, err, "Invalid environment should be rejected")
			assert.Nil(t, invalidConfig, "Invalid config should be nil")
			assert.Contains(t, err.Error(), "unsupported environment", "Error should indicate unsupported environment")
			
			// Test rejection of empty environment
			emptyConfig, err := LoadEnvironmentConfiguration("", cfg)
			require.Error(t, err, "Empty environment should be rejected")
			assert.Nil(t, emptyConfig, "Empty config should be nil")
			
			return nil
		})
	})
	
	t.Run("EnvironmentConfigurationStructure", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "environment-config-structure-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			config, err := LoadEnvironmentConfiguration("development", cfg)
			require.NoError(t, err, "Configuration should load successfully")
			
			// Verify configuration structure contains required fields
			assert.NotEmpty(t, config.Environment, "Environment field should not be empty")
			assert.NotNil(t, config.DeploymentOrder, "DeploymentOrder should not be nil")
			assert.NotEmpty(t, config.DeploymentOrder, "DeploymentOrder should not be empty")
			assert.NotNil(t, config.OutputMappings, "OutputMappings should not be nil")
			assert.NotNil(t, config.ComponentSettings, "ComponentSettings should not be nil")
			
			// Verify deployment order contains expected components
			expectedComponents := []string{"database", "storage", "vault", "redis", "rabbitmq", "observability", "dapr", "services", "website"}
			assert.Equal(t, expectedComponents, config.DeploymentOrder, "DeploymentOrder should match expected components")
			
			// Verify output mappings contain required keys
			expectedOutputs := []string{
				"environment", "database_connection_string", "storage_connection_string",
				"vault_address", "redis_endpoint", "rabbitmq_endpoint", "grafana_url",
				"dapr_control_plane_url", "public_gateway_url", "admin_gateway_url", "website_url",
			}
			for _, output := range expectedOutputs {
				assert.Contains(t, config.OutputMappings, output, "OutputMappings should contain %s", output)
			}
			
			return nil
		})
	})
}

// TestEnvironmentConfigurationValidation validates configuration validation logic
func TestEnvironmentConfigurationValidation(t *testing.T) {
	t.Run("ValidateCompleteConfiguration", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "config-validation-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			config, err := LoadEnvironmentConfiguration("development", cfg)
			require.NoError(t, err, "Configuration should load successfully")
			
			// Test configuration validation
			err = config.Validate()
			assert.NoError(t, err, "Valid configuration should pass validation")
			
			return nil
		})
	})
	
	t.Run("DetectInvalidConfiguration", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "config-validation-invalid-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			// Test configuration with missing deployment order
			invalidConfig := &EnvironmentConfiguration{
				Environment:       "test",
				DeploymentOrder:   []string{}, // Empty deployment order should be invalid
				OutputMappings:    make(map[string]string),
				ComponentSettings: make(map[string]interface{}),
			}
			
			err := invalidConfig.Validate()
			assert.Error(t, err, "Configuration with empty deployment order should be invalid")
			assert.Contains(t, err.Error(), "deployment order cannot be empty", "Error should indicate empty deployment order")
			
			// Test configuration with missing output mappings
			invalidConfig2 := &EnvironmentConfiguration{
				Environment:       "test",
				DeploymentOrder:   []string{"database", "storage"},
				OutputMappings:    make(map[string]string), // Empty output mappings should be invalid
				ComponentSettings: make(map[string]interface{}),
			}
			
			err = invalidConfig2.Validate()
			assert.Error(t, err, "Configuration with missing output mappings should be invalid")
			
			return nil
		})
	})
}

// TestEnvironmentConfigurationDefaults validates default configuration behavior
func TestEnvironmentConfigurationDefaults(t *testing.T) {
	t.Run("ApplyEnvironmentDefaults", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "config-defaults-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			// Test that each environment gets appropriate defaults
			devConfig, err := LoadEnvironmentConfiguration("development", cfg)
			require.NoError(t, err, "Development config should load")
			
			stagingConfig, err := LoadEnvironmentConfiguration("staging", cfg)
			require.NoError(t, err, "Staging config should load")
			
			productionConfig, err := LoadEnvironmentConfiguration("production", cfg)
			require.NoError(t, err, "Production config should load")
			
			// All environments should have same deployment order (no environment-specific order)
			assert.Equal(t, devConfig.DeploymentOrder, stagingConfig.DeploymentOrder, "Dev and staging should have same deployment order")
			assert.Equal(t, stagingConfig.DeploymentOrder, productionConfig.DeploymentOrder, "Staging and production should have same deployment order")
			
			// All environments should have same output mappings
			assert.Equal(t, len(devConfig.OutputMappings), len(stagingConfig.OutputMappings), "Dev and staging should have same number of outputs")
			assert.Equal(t, len(stagingConfig.OutputMappings), len(productionConfig.OutputMappings), "Staging and production should have same number of outputs")
			
			// Component settings may differ between environments (this is expected)
			assert.NotNil(t, devConfig.ComponentSettings, "Development should have component settings")
			assert.NotNil(t, stagingConfig.ComponentSettings, "Staging should have component settings")
			assert.NotNil(t, productionConfig.ComponentSettings, "Production should have component settings")
			
			return nil
		})
	})
	
	t.Run("ConfigurationConsistency", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "config-consistency-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			// Test that configuration loading is consistent
			config1, err := LoadEnvironmentConfiguration("development", cfg)
			require.NoError(t, err, "First config load should succeed")
			
			config2, err := LoadEnvironmentConfiguration("development", cfg)
			require.NoError(t, err, "Second config load should succeed")
			
			// Configurations should be identical for same environment
			assert.Equal(t, config1.Environment, config2.Environment, "Environments should match")
			assert.Equal(t, config1.DeploymentOrder, config2.DeploymentOrder, "Deployment orders should match")
			assert.Equal(t, len(config1.OutputMappings), len(config2.OutputMappings), "Output mappings should have same length")
			
			return nil
		})
	})
}