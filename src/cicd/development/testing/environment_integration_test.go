package testing

import (
	"context"
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/cicd/components"
	"github.com/axiom-software-co/international-center/src/cicd/migration"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDevelopmentEnvironment_FullDeployment validates complete development environment deployment
func TestDevelopmentEnvironment_FullDeployment(t *testing.T) {
	timeout := 30 * time.Second
	
	_, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	err := pulumi.RunErr(func(pulumiCtx *pulumi.Context) error {
		cfg := config.New(pulumiCtx, "")
		environment := "development"
		
		// Deploy all infrastructure components in dependency order
		databaseOutputs, err := components.DeployDatabase(pulumiCtx, cfg, environment)
		require.NoError(t, err)
		
		storageOutputs, err := components.DeployStorage(pulumiCtx, cfg, environment)
		require.NoError(t, err)
		
		vaultOutputs, err := components.DeployVault(pulumiCtx, cfg, environment)
		require.NoError(t, err)
		
		observabilityOutputs, err := components.DeployObservability(pulumiCtx, cfg, environment)
		require.NoError(t, err)
		
		daprOutputs, err := components.DeployDapr(pulumiCtx, cfg, environment)
		require.NoError(t, err)
		
		servicesOutputs, err := components.DeployServices(pulumiCtx, cfg, environment)
		require.NoError(t, err)
		
		websiteOutputs, err := components.DeployWebsite(pulumiCtx, cfg, environment)
		require.NoError(t, err)
		
		// Execute database migrations
		migrationRunner, err := migration.NewMigrationRunner(pulumiCtx, cfg, environment, databaseOutputs)
		require.NoError(t, err)
		
		migrationResult, err := migrationRunner.ExecuteMigration(pulumiCtx)
		require.NoError(t, err)
		
		// Validate schema consistency
		schemaValidator, err := migration.NewSchemaValidator(pulumiCtx, cfg, environment, databaseOutputs)
		require.NoError(t, err)
		
		domains := []string{"business", "donations", "events", "media", "news", "research", "services", "volunteers"}
		for _, domain := range domains {
			validationResult, err := schemaValidator.ValidateDomainSchema(pulumiCtx, domain)
			require.NoError(t, err, "Schema validation failed for domain %s", domain)
			
			validationResult.IsValid.ApplyT(func(isValid bool) error {
				assert.True(t, isValid, "Domain %s schema should be valid", domain)
				return nil
			})
		}
		
		// Validate component integration
		validateComponentIntegration(t, pulumiCtx, &ComponentOutputs{
			Database:      databaseOutputs,
			Storage:       storageOutputs,
			Vault:         vaultOutputs,
			Observability: observabilityOutputs,
			Dapr:          daprOutputs,
			Services:      servicesOutputs,
			Website:       websiteOutputs,
			Migration:     migrationResult,
		})
		
		return nil
	}, pulumi.WithMocks("integration-test", "integration-stack", &IntegrationTestMocks{}))
	
	require.NoError(t, err)
}

// TestDevelopmentEnvironment_ServiceCommunication validates service-to-service communication
func TestDevelopmentEnvironment_ServiceCommunication(t *testing.T) {
	timeout := 30 * time.Second
	
	_, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	err := pulumi.RunErr(func(pulumiCtx *pulumi.Context) error {
		cfg := config.New(pulumiCtx, "")
		environment := "development"
		
		// Deploy services and dapr components
		databaseOutputs, err := components.DeployDatabase(pulumiCtx, cfg, environment)
		require.NoError(t, err)
		
		daprOutputs, err := components.DeployDapr(pulumiCtx, cfg, environment)
		require.NoError(t, err)
		
		servicesOutputs, err := components.DeployServices(pulumiCtx, cfg, environment)
		require.NoError(t, err)
		
		// Validate service communication through Dapr
		validateServiceCommunication(t, pulumiCtx, &ServiceCommunicationTest{
			Database: databaseOutputs,
			Dapr:     daprOutputs,
			Services: servicesOutputs,
		})
		
		return nil
	}, pulumi.WithMocks("integration-test", "integration-stack", &IntegrationTestMocks{}))
	
	require.NoError(t, err)
}

// TestDevelopmentEnvironment_DatabaseMigrationFlow validates end-to-end migration workflow
func TestDevelopmentEnvironment_DatabaseMigrationFlow(t *testing.T) {
	timeout := 30 * time.Second
	
	_, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	err := pulumi.RunErr(func(pulumiCtx *pulumi.Context) error {
		cfg := config.New(pulumiCtx, "")
		environment := "development"
		
		// Deploy database
		databaseOutputs, err := components.DeployDatabase(pulumiCtx, cfg, environment)
		require.NoError(t, err)
		
		// Execute migration workflow
		migrationRunner, err := migration.NewMigrationRunner(pulumiCtx, cfg, environment, databaseOutputs)
		require.NoError(t, err)
		
		migrationResult, err := migrationRunner.ExecuteMigration(pulumiCtx)
		require.NoError(t, err)
		
		// Validate migration results
		migrationResult.ExecutionStatus.ApplyT(func(status string) error {
			assert.Equal(t, "completed", status, "Migration should complete successfully")
			return nil
		})
		
		migrationResult.MigrationsRun.ApplyT(func(count int) error {
			assert.Equal(t, 8, count, "Should run migrations for all 8 domains")
			return nil
		})
		
		// Validate all domain schemas after migration
		schemaValidator, err := migration.NewSchemaValidator(pulumiCtx, cfg, environment, databaseOutputs)
		require.NoError(t, err)
		
		domains := []string{"business", "donations", "events", "media", "news", "research", "services", "volunteers"}
		for _, domain := range domains {
			result, err := schemaValidator.ValidateDomainSchema(pulumiCtx, domain)
			require.NoError(t, err)
			
			result.IsValid.ApplyT(func(isValid bool) error {
				assert.True(t, isValid, "Domain %s schema should be valid after migration", domain)
				return nil
			})
		}
		
		return nil
	}, pulumi.WithMocks("integration-test", "integration-stack", &IntegrationTestMocks{}))
	
	require.NoError(t, err)
}

// ComponentOutputs aggregates all component outputs for integration testing
type ComponentOutputs struct {
	Database      *components.DatabaseOutputs
	Storage       *components.StorageOutputs
	Vault         *components.VaultOutputs
	Observability *components.ObservabilityOutputs
	Dapr          *components.DaprOutputs
	Services      *components.ServicesOutputs
	Website       *components.WebsiteOutputs
	Migration     *migration.MigrationExecutionResult
}

// ServiceCommunicationTest holds components needed for service communication testing
type ServiceCommunicationTest struct {
	Database *components.DatabaseOutputs
	Dapr     *components.DaprOutputs
	Services *components.ServicesOutputs
}

// validateComponentIntegration validates that all infrastructure components integrate correctly
func validateComponentIntegration(t *testing.T, ctx *pulumi.Context, outputs *ComponentOutputs) {
	// Database integration validation
	outputs.Database.ConnectionString.ApplyT(func(connStr string) error {
		assert.NotEmpty(t, connStr, "Database connection string should be provided")
		return nil
	})
	
	// Storage integration validation  
	outputs.Storage.ConnectionString.ApplyT(func(connStr string) error {
		assert.NotEmpty(t, connStr, "Storage connection string should be provided")
		return nil
	})
	
	// Vault integration validation
	outputs.Vault.VaultAddress.ApplyT(func(url string) error {
		assert.NotEmpty(t, url, "Vault address should be provided")
		return nil
	})
	
	// Services integration validation
	pulumi.All(outputs.Services.PublicGatewayURL, outputs.Services.AdminGatewayURL).ApplyT(func(args []interface{}) error {
		publicURL := args[0].(string)
		adminURL := args[1].(string)
		
		assert.NotEmpty(t, publicURL, "Public gateway URL should be provided")
		assert.NotEmpty(t, adminURL, "Admin gateway URL should be provided")
		assert.NotEqual(t, publicURL, adminURL, "Gateway URLs should be different")
		return nil
	})
	
	// Website integration validation
	outputs.Website.ServerURL.ApplyT(func(url string) error {
		assert.NotEmpty(t, url, "Website server URL should be provided")
		return nil
	})
	
	// Migration integration validation
	outputs.Migration.ExecutionStatus.ApplyT(func(status string) error {
		assert.Equal(t, "completed", status, "Migration should complete successfully")
		return nil
	})
}

// validateServiceCommunication validates Dapr service-to-service communication
func validateServiceCommunication(t *testing.T, ctx *pulumi.Context, test *ServiceCommunicationTest) {
	// Validate Dapr control plane is accessible
	test.Dapr.ControlPlaneURL.ApplyT(func(url string) error {
		assert.NotEmpty(t, url, "Dapr control plane URL should be provided")
		assert.Contains(t, url, "http", "Control plane URL should be HTTP endpoint")
		return nil
	})
	
	// Validate Dapr middleware is configured
	test.Dapr.MiddlewareEnabled.ApplyT(func(enabled bool) error {
		assert.True(t, enabled, "Dapr middleware should be enabled for service communication")
		return nil
	})
	
	// Validate Dapr policies are configured
	test.Dapr.PolicyEnabled.ApplyT(func(enabled bool) error {
		assert.True(t, enabled, "Dapr policies should be enabled for secure communication")
		return nil
	})
	
	// Validate deployment type matches environment
	test.Dapr.DeploymentType.ApplyT(func(deploymentType string) error {
		assert.Equal(t, "self_hosted", deploymentType, "Development should use self-hosted Dapr")
		return nil
	})
}