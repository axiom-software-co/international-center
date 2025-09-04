package infrastructure

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// InfrastructureFactory is the main factory interface that combines all environment-specific factories
type InfrastructureFactory interface {
	DatabaseFactory
	StorageFactory
	DaprFactory
	VaultFactory
	ObservabilityFactory
	ServiceFactory
}

// InfrastructureFactoryConfig holds common configuration for all factory implementations
type InfrastructoryFactoryConfig struct {
	Environment string
	NetworkName string
	Region      string
	Tags        map[string]string
}

func GetInfrastructureFactoryConfig(environment string, config *config.Config) *InfrastructoryFactoryConfig {
	return &InfrastructoryFactoryConfig{
		Environment: environment,
		NetworkName: environment + "-network",
		Region:      config.Get("azure:location"),
		Tags: map[string]string{
			"Environment": environment,
			"ManagedBy":   "pulumi",
			"Project":     "international-center",
		},
	}
}