package shared

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// EnvironmentFactory creates deployment strategies for different environments
type EnvironmentFactory struct{}

// NewEnvironmentFactory creates a new environment factory
func NewEnvironmentFactory() *EnvironmentFactory {
	return &EnvironmentFactory{}
}

// CreateDeploymentStrategy creates a deployment strategy for the specified environment
func (ef *EnvironmentFactory) CreateDeploymentStrategy(environment string, ctx *pulumi.Context, cfg *config.Config) (DeploymentStrategy, error) {
	// Validate environment
	validEnvironments := map[string]bool{
		"development": true,
		"staging":     true,
		"production":  true,
	}

	if !validEnvironments[environment] {
		return nil, fmt.Errorf("unknown environment: %s", environment)
	}

	// Create deployment strategy using the factory pattern
	strategy, err := NewDeploymentStrategy(environment, ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment strategy for environment %s: %w", environment, err)
	}

	return strategy, nil
}