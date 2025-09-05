package infrastructure

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	sharedinfra "github.com/axiom-software-co/international-center/src/deployer/shared/infrastructure"
)

// DevelopmentInfrastructureFactory implements all factory interfaces for development environment
type DevelopmentInfrastructureFactory struct{}

func NewDevelopmentInfrastructureFactory() *DevelopmentInfrastructureFactory {
	return &DevelopmentInfrastructureFactory{}
}

// Database Factory Implementation
func (f *DevelopmentInfrastructureFactory) CreateDatabaseStack(ctx *pulumi.Context, config *config.Config, environment string) sharedinfra.DatabaseStack {
	return NewDatabaseStack(ctx, config, "development-network", environment)
}

// Storage Factory Implementation  
func (f *DevelopmentInfrastructureFactory) CreateStorageStack(ctx *pulumi.Context, config *config.Config, environment string) sharedinfra.StorageStack {
	return NewStorageStack(ctx, config, "development-network", environment)
}

// Dapr Factory Implementation
func (f *DevelopmentInfrastructureFactory) CreateDaprStack(ctx *pulumi.Context, config *config.Config, environment string) sharedinfra.DaprStack {
	return NewDaprStack(ctx, config, "development-network", environment)
}

// Vault Factory Implementation
func (f *DevelopmentInfrastructureFactory) CreateVaultStack(ctx *pulumi.Context, config *config.Config, environment string) sharedinfra.VaultStack {
	return NewVaultStack(ctx, config, "development-network", environment)
}

// Observability Factory Implementation
func (f *DevelopmentInfrastructureFactory) CreateObservabilityStack(ctx *pulumi.Context, config *config.Config, environment string) sharedinfra.ObservabilityStack {
	return NewObservabilityStack(ctx, config, "development-network", environment)
}

// Service Factory Implementation
func (f *DevelopmentInfrastructureFactory) CreateServiceStack(ctx *pulumi.Context, config *config.Config, environment string) sharedinfra.ServiceStack {
	// For development factory, we create a service stack without external dependencies
	// The service stack will handle its own dapr dependency internally
	return NewServiceStack(ctx, config, nil, "development-network", environment, ".")
}

// Website Factory Implementation
func (f *DevelopmentInfrastructureFactory) CreateWebsiteStack(ctx *pulumi.Context, config *config.Config, environment string) sharedinfra.WebsiteStack {
	return NewWebsiteStack(ctx, config, environment)
}

// Verify that DevelopmentInfrastructureFactory implements the shared InfrastructureFactory interface
var _ sharedinfra.InfrastructureFactory = (*DevelopmentInfrastructureFactory)(nil)
var _ sharedinfra.DatabaseFactory = (*DevelopmentInfrastructureFactory)(nil)
var _ sharedinfra.StorageFactory = (*DevelopmentInfrastructureFactory)(nil)
var _ sharedinfra.DaprFactory = (*DevelopmentInfrastructureFactory)(nil)
var _ sharedinfra.VaultFactory = (*DevelopmentInfrastructureFactory)(nil)
var _ sharedinfra.ObservabilityFactory = (*DevelopmentInfrastructureFactory)(nil)
var _ sharedinfra.ServiceFactory = (*DevelopmentInfrastructureFactory)(nil)
var _ sharedinfra.WebsiteFactory = (*DevelopmentInfrastructureFactory)(nil)