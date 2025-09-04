package infrastructure

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	sharedinfra "github.com/axiom-software-co/international-center/src/deployer/shared/infrastructure"
)

// StagingInfrastructureFactory implements all factory interfaces for staging environment
// Uses Azure managed services instead of containers for better reliability and scalability
type StagingInfrastructureFactory struct{}

func NewStagingInfrastructureFactory() *StagingInfrastructureFactory {
	return &StagingInfrastructureFactory{}
}

// Database Factory Implementation - Uses Azure Database for PostgreSQL
func (f *StagingInfrastructureFactory) CreateDatabaseStack(ctx *pulumi.Context, config *config.Config, environment string) sharedinfra.DatabaseStack {
	// TODO: Implement proper shared interface compatibility for AzureDatabaseStack
	return nil // NewAzureDatabaseStack requires resource group parameter
}

// Storage Factory Implementation - Uses Azure Storage Account
func (f *StagingInfrastructureFactory) CreateStorageStack(ctx *pulumi.Context, config *config.Config, environment string) sharedinfra.StorageStack {
	// TODO: Implement proper shared interface compatibility for AzureStorageStack
	return nil // NewAzureStorageStack requires resource group parameter
}

// Dapr Factory Implementation - Uses Azure Container Apps with Dapr
func (f *StagingInfrastructureFactory) CreateDaprStack(ctx *pulumi.Context, config *config.Config, environment string) sharedinfra.DaprStack {
	// TODO: Implement proper shared interface compatibility
	return nil // Stack not yet implemented
}

// Vault Factory Implementation - Uses Azure Key Vault
func (f *StagingInfrastructureFactory) CreateVaultStack(ctx *pulumi.Context, config *config.Config, environment string) sharedinfra.VaultStack {
	// TODO: Implement proper shared interface compatibility for VaultCloudStack
	return nil // NewVaultCloudStack requires resource group parameter
}

// Observability Factory Implementation - Uses Azure Monitor, Application Insights
func (f *StagingInfrastructureFactory) CreateObservabilityStack(ctx *pulumi.Context, config *config.Config, environment string) sharedinfra.ObservabilityStack {
	// TODO: Implement proper shared interface compatibility
	return nil // Stack not yet implemented
}

// Service Factory Implementation - Uses Azure Container Apps
func (f *StagingInfrastructureFactory) CreateServiceStack(ctx *pulumi.Context, config *config.Config, environment string) sharedinfra.ServiceStack {
	// TODO: Implement proper shared interface compatibility
	return nil // Stack not yet implemented
}

// Verify that StagingInfrastructureFactory implements the shared InfrastructureFactory interface
var _ sharedinfra.InfrastructureFactory = (*StagingInfrastructureFactory)(nil)
var _ sharedinfra.DatabaseFactory = (*StagingInfrastructureFactory)(nil)
var _ sharedinfra.StorageFactory = (*StagingInfrastructureFactory)(nil)
var _ sharedinfra.DaprFactory = (*StagingInfrastructureFactory)(nil)
var _ sharedinfra.VaultFactory = (*StagingInfrastructureFactory)(nil)
var _ sharedinfra.ObservabilityFactory = (*StagingInfrastructureFactory)(nil)
var _ sharedinfra.ServiceFactory = (*StagingInfrastructureFactory)(nil)