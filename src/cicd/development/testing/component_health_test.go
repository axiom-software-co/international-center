package testing

import (
	"context"
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/cicd/components"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestComponentHealth_Database validates database component health before integration testing
func TestComponentHealth_Database(t *testing.T) {
	timeout := 15 * time.Second
	
	_, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	err := pulumi.RunErr(func(pulumiCtx *pulumi.Context) error {
		cfg := config.New(pulumiCtx, "")
		environment := "development"
		
		// Deploy database component
		databaseOutputs, err := components.DeployDatabase(pulumiCtx, cfg, environment)
		require.NoError(t, err)
		
		// Validate database health
		validateDatabaseHealth(t, pulumiCtx, databaseOutputs)
		
		return nil
	}, pulumi.WithMocks("health-test", "health-stack", &HealthValidationMocks{}))
	
	require.NoError(t, err)
}

// TestComponentHealth_Storage validates storage component health before integration testing
func TestComponentHealth_Storage(t *testing.T) {
	timeout := 15 * time.Second
	
	_, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	err := pulumi.RunErr(func(pulumiCtx *pulumi.Context) error {
		cfg := config.New(pulumiCtx, "")
		environment := "development"
		
		// Deploy storage component
		storageOutputs, err := components.DeployStorage(pulumiCtx, cfg, environment)
		require.NoError(t, err)
		
		// Validate storage health
		validateStorageHealth(t, pulumiCtx, storageOutputs)
		
		return nil
	}, pulumi.WithMocks("health-test", "health-stack", &HealthValidationMocks{}))
	
	require.NoError(t, err)
}

// TestComponentHealth_Vault validates vault component health before integration testing
func TestComponentHealth_Vault(t *testing.T) {
	timeout := 15 * time.Second
	
	_, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	err := pulumi.RunErr(func(pulumiCtx *pulumi.Context) error {
		cfg := config.New(pulumiCtx, "")
		environment := "development"
		
		// Deploy vault component
		vaultOutputs, err := components.DeployVault(pulumiCtx, cfg, environment)
		require.NoError(t, err)
		
		// Validate vault health
		validateVaultHealth(t, pulumiCtx, vaultOutputs)
		
		return nil
	}, pulumi.WithMocks("health-test", "health-stack", &HealthValidationMocks{}))
	
	require.NoError(t, err)
}

// TestComponentHealth_Observability validates observability component health before integration testing
func TestComponentHealth_Observability(t *testing.T) {
	timeout := 15 * time.Second
	
	_, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	err := pulumi.RunErr(func(pulumiCtx *pulumi.Context) error {
		cfg := config.New(pulumiCtx, "")
		environment := "development"
		
		// Deploy observability component
		observabilityOutputs, err := components.DeployObservability(pulumiCtx, cfg, environment)
		require.NoError(t, err)
		
		// Validate observability health
		validateObservabilityHealth(t, pulumiCtx, observabilityOutputs)
		
		return nil
	}, pulumi.WithMocks("health-test", "health-stack", &HealthValidationMocks{}))
	
	require.NoError(t, err)
}

// TestComponentHealth_Dapr validates dapr component health before integration testing
func TestComponentHealth_Dapr(t *testing.T) {
	timeout := 15 * time.Second
	
	_, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	err := pulumi.RunErr(func(pulumiCtx *pulumi.Context) error {
		cfg := config.New(pulumiCtx, "")
		environment := "development"
		
		// Deploy dapr component
		daprOutputs, err := components.DeployDapr(pulumiCtx, cfg, environment)
		require.NoError(t, err)
		
		// Validate dapr health
		validateDaprHealth(t, pulumiCtx, daprOutputs)
		
		return nil
	}, pulumi.WithMocks("health-test", "health-stack", &HealthValidationMocks{}))
	
	require.NoError(t, err)
}

// TestComponentHealth_Services validates services component health before integration testing
func TestComponentHealth_Services(t *testing.T) {
	timeout := 15 * time.Second
	
	_, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	err := pulumi.RunErr(func(pulumiCtx *pulumi.Context) error {
		cfg := config.New(pulumiCtx, "")
		environment := "development"
		
		// Deploy services component
		servicesOutputs, err := components.DeployServices(pulumiCtx, cfg, environment)
		require.NoError(t, err)
		
		// Validate services health
		validateServicesHealth(t, pulumiCtx, servicesOutputs)
		
		return nil
	}, pulumi.WithMocks("health-test", "health-stack", &HealthValidationMocks{}))
	
	require.NoError(t, err)
}

// TestComponentHealth_Website validates website component health before integration testing
func TestComponentHealth_Website(t *testing.T) {
	timeout := 15 * time.Second
	
	_, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	err := pulumi.RunErr(func(pulumiCtx *pulumi.Context) error {
		cfg := config.New(pulumiCtx, "")
		environment := "development"
		
		// Deploy website component
		websiteOutputs, err := components.DeployWebsite(pulumiCtx, cfg, environment)
		require.NoError(t, err)
		
		// Validate website health
		validateWebsiteHealth(t, pulumiCtx, websiteOutputs)
		
		return nil
	}, pulumi.WithMocks("health-test", "health-stack", &HealthValidationMocks{}))
	
	require.NoError(t, err)
}

// TestComponentHealth_AllComponents validates health of all components together
func TestComponentHealth_AllComponents(t *testing.T) {
	timeout := 30 * time.Second
	
	_, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	err := pulumi.RunErr(func(pulumiCtx *pulumi.Context) error {
		cfg := config.New(pulumiCtx, "")
		environment := "development"
		
		// Deploy all components
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
		
		// Validate all component health
		validateAllComponentsHealth(t, pulumiCtx, &AllComponentsHealth{
			Database:      databaseOutputs,
			Storage:       storageOutputs,
			Vault:         vaultOutputs,
			Observability: observabilityOutputs,
			Dapr:          daprOutputs,
			Services:      servicesOutputs,
			Website:       websiteOutputs,
		})
		
		return nil
	}, pulumi.WithMocks("health-test", "health-stack", &HealthValidationMocks{}))
	
	require.NoError(t, err)
}

// AllComponentsHealth aggregates all component outputs for health validation
type AllComponentsHealth struct {
	Database      *components.DatabaseOutputs
	Storage       *components.StorageOutputs
	Vault         *components.VaultOutputs
	Observability *components.ObservabilityOutputs
	Dapr          *components.DaprOutputs
	Services      *components.ServicesOutputs
	Website       *components.WebsiteOutputs
}

// validateDatabaseHealth validates database component is healthy and accessible
func validateDatabaseHealth(t *testing.T, ctx *pulumi.Context, outputs *components.DatabaseOutputs) {
	outputs.ConnectionString.ApplyT(func(connStr string) error {
		assert.NotEmpty(t, connStr, "Database connection string should be provided")
		assert.Contains(t, connStr, "postgresql://", "Connection string should be PostgreSQL format")
		return nil
	})
	
	outputs.Port.ApplyT(func(port int) error {
		assert.Equal(t, 5432, port, "Database should run on PostgreSQL default port")
		return nil
	})
	
	outputs.DatabaseName.ApplyT(func(name string) error {
		assert.NotEmpty(t, name, "Database name should be provided")
		return nil
	})
}

// validateStorageHealth validates storage component is healthy and accessible
func validateStorageHealth(t *testing.T, ctx *pulumi.Context, outputs *components.StorageOutputs) {
	outputs.ConnectionString.ApplyT(func(connStr string) error {
		assert.NotEmpty(t, connStr, "Storage connection string should be provided")
		return nil
	})
	
	outputs.StorageType.ApplyT(func(storageType string) error {
		assert.Contains(t, []string{"blob_storage", "azurite"}, storageType, "Storage type should be valid")
		return nil
	})
}

// validateVaultHealth validates vault component is healthy and accessible
func validateVaultHealth(t *testing.T, ctx *pulumi.Context, outputs *components.VaultOutputs) {
	outputs.VaultAddress.ApplyT(func(address string) error {
		assert.NotEmpty(t, address, "Vault address should be provided")
		assert.Contains(t, address, "http", "Vault address should be HTTP endpoint")
		return nil
	})
}

// validateObservabilityHealth validates observability component is healthy and accessible
func validateObservabilityHealth(t *testing.T, ctx *pulumi.Context, outputs *components.ObservabilityOutputs) {
	outputs.GrafanaURL.ApplyT(func(url string) error {
		assert.NotEmpty(t, url, "Grafana URL should be provided")
		assert.Contains(t, url, "http", "Grafana URL should be HTTP endpoint")
		return nil
	})
	
	outputs.AlertingEnabled.ApplyT(func(enabled bool) error {
		assert.True(t, enabled, "Alerting should be enabled for health monitoring")
		return nil
	})
}

// validateDaprHealth validates dapr component is healthy and accessible
func validateDaprHealth(t *testing.T, ctx *pulumi.Context, outputs *components.DaprOutputs) {
	outputs.ControlPlaneURL.ApplyT(func(url string) error {
		assert.NotEmpty(t, url, "Dapr control plane URL should be provided")
		assert.Contains(t, url, "http", "Control plane URL should be HTTP endpoint")
		return nil
	})
	
	outputs.MiddlewareEnabled.ApplyT(func(enabled bool) error {
		assert.True(t, enabled, "Dapr middleware should be enabled")
		return nil
	})
	
	outputs.PolicyEnabled.ApplyT(func(enabled bool) error {
		assert.True(t, enabled, "Dapr policies should be enabled")
		return nil
	})
}

// validateServicesHealth validates services component is healthy and accessible
func validateServicesHealth(t *testing.T, ctx *pulumi.Context, outputs *components.ServicesOutputs) {
	pulumi.All(outputs.PublicGatewayURL, outputs.AdminGatewayURL).ApplyT(func(args []interface{}) error {
		publicURL := args[0].(string)
		adminURL := args[1].(string)
		
		assert.NotEmpty(t, publicURL, "Public gateway URL should be provided")
		assert.NotEmpty(t, adminURL, "Admin gateway URL should be provided")
		assert.NotEqual(t, publicURL, adminURL, "Gateway URLs should be different")
		assert.Contains(t, publicURL, "http", "Public gateway URL should be HTTP endpoint")
		assert.Contains(t, adminURL, "http", "Admin gateway URL should be HTTP endpoint")
		return nil
	})
	
	outputs.HealthCheckEnabled.ApplyT(func(enabled bool) error {
		assert.True(t, enabled, "Health checks should be enabled")
		return nil
	})
	
	outputs.DaprSidecarEnabled.ApplyT(func(enabled bool) error {
		assert.True(t, enabled, "Dapr sidecar should be enabled for services")
		return nil
	})
}

// validateWebsiteHealth validates website component is healthy and accessible
func validateWebsiteHealth(t *testing.T, ctx *pulumi.Context, outputs *components.WebsiteOutputs) {
	outputs.ServerURL.ApplyT(func(url string) error {
		assert.NotEmpty(t, url, "Website server URL should be provided")
		assert.Contains(t, url, "http", "Website URL should be HTTP endpoint")
		return nil
	})
	
	outputs.CDNEnabled.ApplyT(func(enabled bool) error {
		assert.True(t, enabled, "CDN should be enabled for performance")
		return nil
	})
}

// validateAllComponentsHealth validates that all components are healthy and integrated
func validateAllComponentsHealth(t *testing.T, ctx *pulumi.Context, health *AllComponentsHealth) {
	// Individual component health validation
	validateDatabaseHealth(t, ctx, health.Database)
	validateStorageHealth(t, ctx, health.Storage)
	validateVaultHealth(t, ctx, health.Vault)
	validateObservabilityHealth(t, ctx, health.Observability)
	validateDaprHealth(t, ctx, health.Dapr)
	validateServicesHealth(t, ctx, health.Services)
	validateWebsiteHealth(t, ctx, health.Website)
	
	// Cross-component integration health validation
	validateComponentIntegrationHealth(t, ctx, health)
}

// validateComponentIntegrationHealth validates that components integrate correctly for health
func validateComponentIntegrationHealth(t *testing.T, ctx *pulumi.Context, health *AllComponentsHealth) {
	// Validate database and services integration
	pulumi.All(health.Database.ConnectionString, health.Services.HealthCheckEnabled).ApplyT(func(args []interface{}) error {
		dbConnStr := args[0].(string)
		healthEnabled := args[1].(bool)
		
		assert.NotEmpty(t, dbConnStr, "Database connection should be available for services")
		assert.True(t, healthEnabled, "Services health check should validate database connectivity")
		return nil
	})
	
	// Validate vault and dapr integration
	pulumi.All(health.Vault.VaultAddress, health.Dapr.MiddlewareEnabled).ApplyT(func(args []interface{}) error {
		vaultAddress := args[0].(string)
		middlewareEnabled := args[1].(bool)
		
		assert.NotEmpty(t, vaultAddress, "Vault should be accessible for secrets")
		assert.True(t, middlewareEnabled, "Dapr middleware should be enabled for vault integration")
		return nil
	})
	
	// Validate observability integration
	pulumi.All(health.Observability.GrafanaURL, health.Services.ObservabilityEnabled).ApplyT(func(args []interface{}) error {
		grafanaURL := args[0].(string)
		obsEnabled := args[1].(bool)
		
		assert.NotEmpty(t, grafanaURL, "Observability should be accessible")
		assert.True(t, obsEnabled, "Services should integrate with observability")
		return nil
	})
}