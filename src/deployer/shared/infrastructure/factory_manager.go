package infrastructure

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	devinfra "github.com/axiom-software-co/international-center/src/deployer/development/infrastructure"
	staginginfra "github.com/axiom-software-co/international-center/src/deployer/staging/infrastructure"
	prodinfra "github.com/axiom-software-co/international-center/src/deployer/production/infrastructure"
)

// InfrastructureFactoryManager manages factory selection based on environment
type InfrastructureFactoryManager struct {
	environment string
	factory     InfrastructureFactory
}

// InfrastructureFactory represents the composite factory interface for all infrastructure components
type InfrastructureFactory interface {
	DatabaseFactory
	StorageFactory
	DaprFactory
	VaultFactory
	ObservabilityFactory
	ServiceFactory
}

// NewInfrastructureFactoryManager creates a new factory manager
func NewInfrastructureFactoryManager(environment string) (*InfrastructureFactoryManager, error) {
	var factory InfrastructureFactory
	
	switch environment {
	case "development", "dev", "local":
		factory = devinfra.NewDevelopmentInfrastructureFactory()
	case "staging", "stage", "test":
		factory = staginginfra.NewStagingInfrastructureFactory()
	case "production", "prod":
		factory = prodinfra.NewProductionInfrastructureFactory()
	default:
		return nil, fmt.Errorf("unsupported environment: %s", environment)
	}
	
	return &InfrastructureFactoryManager{
		environment: environment,
		factory:     factory,
	}, nil
}

// GetEnvironment returns the current environment
func (fm *InfrastructureFactoryManager) GetEnvironment() string {
	return fm.environment
}

// Database Factory Methods
func (fm *InfrastructureFactoryManager) CreateDatabaseStack(ctx *pulumi.Context, config *config.Config, environment string) DatabaseStack {
	return fm.factory.CreateDatabaseStack(ctx, config, environment)
}

// Storage Factory Methods
func (fm *InfrastructureFactoryManager) CreateStorageStack(ctx *pulumi.Context, config *config.Config, environment string) StorageStack {
	return fm.factory.CreateStorageStack(ctx, config, environment)
}

// Dapr Factory Methods
func (fm *InfrastructureFactoryManager) CreateDaprStack(ctx *pulumi.Context, config *config.Config, environment string) DaprStack {
	return fm.factory.CreateDaprStack(ctx, config, environment)
}

// Vault Factory Methods
func (fm *InfrastructureFactoryManager) CreateVaultStack(ctx *pulumi.Context, config *config.Config, environment string) VaultStack {
	return fm.factory.CreateVaultStack(ctx, config, environment)
}

// Observability Factory Methods
func (fm *InfrastructureFactoryManager) CreateObservabilityStack(ctx *pulumi.Context, config *config.Config, environment string) ObservabilityStack {
	return fm.factory.CreateObservabilityStack(ctx, config, environment)
}

// Service Factory Methods
func (fm *InfrastructureFactoryManager) CreateServiceStack(ctx *pulumi.Context, config *config.Config, environment string) ServiceStack {
	return fm.factory.CreateServiceStack(ctx, config, environment)
}

// DeploymentStrategy represents the deployment strategy for the environment
type DeploymentStrategy struct {
	Environment           string
	UsesContainers        bool
	UsesManagedServices   bool
	SupportsHighAvailability bool
	SupportsGeoReplication bool
	RequiresCompliance    bool
	InfrastructureType    string // "development", "staging", "production"
}

// GetDeploymentStrategy returns the deployment strategy for the environment
func (fm *InfrastructureFactoryManager) GetDeploymentStrategy() DeploymentStrategy {
	switch fm.environment {
	case "development", "dev", "local":
		return DeploymentStrategy{
			Environment:              "development",
			UsesContainers:           true,
			UsesManagedServices:      false,
			SupportsHighAvailability: false,
			SupportsGeoReplication:   false,
			RequiresCompliance:       false,
			InfrastructureType:       "container-based",
		}
	case "staging", "stage", "test":
		return DeploymentStrategy{
			Environment:              "staging",
			UsesContainers:           false,
			UsesManagedServices:      true,
			SupportsHighAvailability: true,
			SupportsGeoReplication:   true,
			RequiresCompliance:       true,
			InfrastructureType:       "azure-managed",
		}
	case "production", "prod":
		return DeploymentStrategy{
			Environment:              "production",
			UsesContainers:           false,
			UsesManagedServices:      true,
			SupportsHighAvailability: true,
			SupportsGeoReplication:   true,
			RequiresCompliance:       true,
			InfrastructureType:       "azure-managed-ha",
		}
	default:
		return DeploymentStrategy{
			Environment:              fm.environment,
			UsesContainers:           true,
			UsesManagedServices:      false,
			SupportsHighAvailability: false,
			SupportsGeoReplication:   false,
			RequiresCompliance:       false,
			InfrastructureType:       "container-based",
		}
	}
}

// ValidateEnvironmentCapabilities validates that the requested capabilities are supported
func (fm *InfrastructureFactoryManager) ValidateEnvironmentCapabilities(requiredCapabilities []string) error {
	strategy := fm.GetDeploymentStrategy()
	
	for _, capability := range requiredCapabilities {
		switch capability {
		case "high-availability":
			if !strategy.SupportsHighAvailability {
				return fmt.Errorf("high availability not supported in %s environment", strategy.Environment)
			}
		case "geo-replication":
			if !strategy.SupportsGeoReplication {
				return fmt.Errorf("geo-replication not supported in %s environment", strategy.Environment)
			}
		case "compliance":
			if !strategy.RequiresCompliance {
				return fmt.Errorf("compliance features not enabled in %s environment", strategy.Environment)
			}
		case "managed-services":
			if !strategy.UsesManagedServices {
				return fmt.Errorf("managed services not available in %s environment", strategy.Environment)
			}
		case "containers":
			if !strategy.UsesContainers {
				return fmt.Errorf("container-based deployment not used in %s environment", strategy.Environment)
			}
		default:
			return fmt.Errorf("unknown capability: %s", capability)
		}
	}
	
	return nil
}

// GetSupportedCapabilities returns the list of supported capabilities for the current environment
func (fm *InfrastructureFactoryManager) GetSupportedCapabilities() []string {
	strategy := fm.GetDeploymentStrategy()
	capabilities := []string{}
	
	if strategy.UsesContainers {
		capabilities = append(capabilities, "containers")
	}
	if strategy.UsesManagedServices {
		capabilities = append(capabilities, "managed-services")
	}
	if strategy.SupportsHighAvailability {
		capabilities = append(capabilities, "high-availability")
	}
	if strategy.SupportsGeoReplication {
		capabilities = append(capabilities, "geo-replication")
	}
	if strategy.RequiresCompliance {
		capabilities = append(capabilities, "compliance")
	}
	
	return capabilities
}

// Verify that InfrastructureFactoryManager implements all factory interfaces
var _ InfrastructureFactory = (*InfrastructureFactoryManager)(nil)
var _ DatabaseFactory = (*InfrastructureFactoryManager)(nil)
var _ StorageFactory = (*InfrastructureFactoryManager)(nil)
var _ DaprFactory = (*InfrastructureFactoryManager)(nil)
var _ VaultFactory = (*InfrastructureFactoryManager)(nil)
var _ ObservabilityFactory = (*InfrastructureFactoryManager)(nil)
var _ ServiceFactory = (*InfrastructureFactoryManager)(nil)