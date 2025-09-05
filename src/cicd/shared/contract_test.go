package shared

import (
	"testing"

	"github.com/axiom-software-co/international-center/src/cicd/components"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestComponentIntegrationContracts validates component interfaces work correctly together
func TestComponentIntegrationContracts(t *testing.T) {
	environments := []string{"development", "staging", "production"}
	framework := NewContractTestingFramework("project", "stack")

	for _, env := range environments {
		t.Run("Environment_"+env, func(t *testing.T) {
			framework.RunComponentContractTest(t, env, func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
				// Load environment configuration
				_, err := LoadEnvironmentConfig(ctx, env)
				require.NoError(t, err)

				// Deploy components in dependency order
				databaseOutputs, err := components.DeployDatabase(ctx, cfg, env)
				require.NoError(t, err)
				ValidateComponentOutputs(t, "database", databaseOutputs, env)

				storageOutputs, err := components.DeployStorage(ctx, cfg, env)
				require.NoError(t, err)
				ValidateComponentOutputs(t, "storage", storageOutputs, env)

				vaultOutputs, err := components.DeployVault(ctx, cfg, env)
				require.NoError(t, err)
				ValidateComponentOutputs(t, "vault", vaultOutputs, env)

				observabilityOutputs, err := components.DeployObservability(ctx, cfg, env)
				require.NoError(t, err)
				ValidateComponentOutputs(t, "observability", observabilityOutputs, env)

				daprOutputs, err := components.DeployDapr(ctx, cfg, env)
				require.NoError(t, err)
				ValidateComponentOutputs(t, "dapr", daprOutputs, env)

				// Services component should be able to consume all previous outputs
				servicesOutputs, err := components.DeployServices(ctx, cfg, env)
				require.NoError(t, err)
				ValidateComponentOutputs(t, "services", servicesOutputs, env)

				// Website component should be able to consume services outputs
				websiteOutputs, err := components.DeployWebsite(ctx, cfg, env)
				require.NoError(t, err)
				ValidateComponentOutputs(t, "website", websiteOutputs, env)

				// Validate component integrations
				ValidateComponentIntegration(t, "database", databaseOutputs, "services", servicesOutputs, env)
				ValidateComponentIntegration(t, "storage", storageOutputs, "services", servicesOutputs, env)
				ValidateComponentIntegration(t, "vault", vaultOutputs, "services", servicesOutputs, env)
				ValidateComponentIntegration(t, "vault", vaultOutputs, "dapr", daprOutputs, env)
				ValidateComponentIntegration(t, "dapr", daprOutputs, "services", servicesOutputs, env)
				ValidateComponentIntegration(t, "services", servicesOutputs, "website", websiteOutputs, env)
				ValidateComponentIntegration(t, "observability", observabilityOutputs, "services", servicesOutputs, env)

				return nil
			})
		})
	}
}

// TestDatabaseServiceIntegration validates database component outputs are correctly consumed by services
func TestDatabaseServiceIntegration(t *testing.T) {
	environments := []string{"development", "staging", "production"}
	framework := NewContractTestingFramework("project", "stack")

	for _, env := range environments {
		t.Run("Environment_"+env, func(t *testing.T) {
			framework.RunComponentContractTest(t, env, func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
				// Deploy database component
				databaseOutputs, err := components.DeployDatabase(ctx, cfg, env)
				require.NoError(t, err)

				// Verify services can consume database connection string
				assert.NotNil(t, databaseOutputs.ConnectionString)

				// Contract: Services component requires database connection string
				servicesOutputs, err := components.DeployServices(ctx, cfg, env)
				require.NoError(t, err)

				// Verify services component references database outputs
				ValidateComponentIntegration(t, "database", databaseOutputs, "services", servicesOutputs, env)

				return nil
			})
		})
	}
}

// TestStorageServiceIntegration validates storage component outputs are correctly consumed by services
func TestStorageServiceIntegration(t *testing.T) {
	environments := []string{"development", "staging", "production"}
	framework := NewContractTestingFramework("project", "stack")

	for _, env := range environments {
		t.Run("Environment_"+env, func(t *testing.T) {
			framework.RunComponentContractTest(t, env, func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
				// Deploy storage component
				storageOutputs, err := components.DeployStorage(ctx, cfg, env)
				require.NoError(t, err)

				// Contract: Services component requires storage connection string
				servicesOutputs, err := components.DeployServices(ctx, cfg, env)
				require.NoError(t, err)

				// Verify services component references storage outputs
				ValidateComponentIntegration(t, "storage", storageOutputs, "services", servicesOutputs, env)

				return nil
			})
		})
	}
}

// TestVaultServiceIntegration validates vault component outputs are correctly consumed by services
func TestVaultServiceIntegration(t *testing.T) {
	environments := []string{"development", "staging", "production"}
	framework := NewContractTestingFramework("project", "stack")

	for _, env := range environments {
		t.Run("Environment_"+env, func(t *testing.T) {
			framework.RunComponentContractTest(t, env, func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
				// Deploy vault component
				vaultOutputs, err := components.DeployVault(ctx, cfg, env)
				require.NoError(t, err)

				// Contract: Services component requires vault secret store configuration
				servicesOutputs, err := components.DeployServices(ctx, cfg, env)
				require.NoError(t, err)

				// Verify services component references vault outputs
				ValidateComponentIntegration(t, "vault", vaultOutputs, "services", servicesOutputs, env)

				return nil
			})
		})
	}
}

// TestWebsiteServiceIntegration validates website component correctly integrates with services
func TestWebsiteServiceIntegration(t *testing.T) {
	environments := []string{"development", "staging", "production"}
	framework := NewContractTestingFramework("project", "stack")

	for _, env := range environments {
		t.Run("Environment_"+env, func(t *testing.T) {
			framework.RunComponentContractTest(t, env, func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
				// Deploy services component first
				servicesOutputs, err := components.DeployServices(ctx, cfg, env)
				require.NoError(t, err)

				// Deploy website component
				websiteOutputs, err := components.DeployWebsite(ctx, cfg, env)
				require.NoError(t, err)

				// Contract: Website component should reference API gateway URL from services
				ValidateComponentIntegration(t, "services", servicesOutputs, "website", websiteOutputs, env)

				return nil
			})
		})
	}
}

// TestEnvironmentParity validates that components maintain environment parity across dev/staging/prod
func TestEnvironmentParity(t *testing.T) {
	framework := NewContractTestingFramework("project", "stack")
	
	// Test database component parity
	t.Run("Database_Environment_Parity", func(t *testing.T) {
		var devOutputs, stagingOutputs, prodOutputs *components.DatabaseOutputs
		
		// Deploy database in all environments
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			outputs, err := components.DeployDatabase(ctx, cfg, env)
			devOutputs = outputs
			return err
		})
		
		framework.RunComponentContractTest(t, "staging", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			outputs, err := components.DeployDatabase(ctx, cfg, env)
			stagingOutputs = outputs
			return err
		})
		
		framework.RunComponentContractTest(t, "production", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			outputs, err := components.DeployDatabase(ctx, cfg, env)
			prodOutputs = outputs
			return err
		})
		
		// Validate parity
		ValidateEnvironmentParity(t, "database", devOutputs, stagingOutputs, prodOutputs)
	})
	
	// Test services component parity
	t.Run("Services_Environment_Parity", func(t *testing.T) {
		var devOutputs, stagingOutputs, prodOutputs *components.ServicesOutputs
		
		// Deploy services in all environments
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			outputs, err := components.DeployServices(ctx, cfg, env)
			devOutputs = outputs
			return err
		})
		
		framework.RunComponentContractTest(t, "staging", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			outputs, err := components.DeployServices(ctx, cfg, env)
			stagingOutputs = outputs
			return err
		})
		
		framework.RunComponentContractTest(t, "production", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			outputs, err := components.DeployServices(ctx, cfg, env)
			prodOutputs = outputs
			return err
		})
		
		// Validate parity
		ValidateEnvironmentParity(t, "services", devOutputs, stagingOutputs, prodOutputs)
	})
}