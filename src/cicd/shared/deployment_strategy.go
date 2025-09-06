package shared

import (
	"github.com/axiom-software-co/international-center/src/cicd/components"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// DeploymentStrategy defines the interface for environment-specific deployment strategies
type DeploymentStrategy interface {
	GetEnvironment() string
	GetDeploymentOrder() []string
	Deploy(ctx *pulumi.Context, cfg *config.Config) (map[string]interface{}, error)
}

// EnvironmentDeploymentStrategy implements DeploymentStrategy for a specific environment
type EnvironmentDeploymentStrategy struct {
	config *EnvironmentConfiguration
}

// NewDeploymentStrategy creates a new deployment strategy for the specified environment
func NewDeploymentStrategy(environment string, ctx *pulumi.Context, cfg *config.Config) (DeploymentStrategy, error) {
	envConfig, err := LoadEnvironmentConfiguration(environment, cfg)
	if err != nil {
		return nil, err
	}

	return &EnvironmentDeploymentStrategy{
		config: envConfig,
	}, nil
}

// GetEnvironment returns the environment this strategy deploys to
func (eds *EnvironmentDeploymentStrategy) GetEnvironment() string {
	return eds.config.Environment
}

// GetDeploymentOrder returns the ordered list of components to deploy
func (eds *EnvironmentDeploymentStrategy) GetDeploymentOrder() []string {
	return eds.config.DeploymentOrder
}

// Deploy executes the deployment strategy and returns outputs
func (eds *EnvironmentDeploymentStrategy) Deploy(ctx *pulumi.Context, cfg *config.Config) (map[string]interface{}, error) {
	environment := eds.config.Environment
	outputs := make(map[string]interface{})

	// Deploy each component in order based on the current main.go logic
	for _, componentName := range eds.config.DeploymentOrder {
		ctx.Log.Info("Deploying "+componentName+" component", nil)

		switch componentName {
		case "database":
			database, err := components.DeployDatabase(ctx, cfg, environment)
			if err != nil {
				return nil, err
			}
			outputs["database_connection_string"] = database.ConnectionString

		case "storage":
			storage, err := components.DeployStorage(ctx, cfg, environment)
			if err != nil {
				return nil, err
			}
			outputs["storage_connection_string"] = storage.ConnectionString

		case "vault":
			vault, err := components.DeployVault(ctx, cfg, environment)
			if err != nil {
				return nil, err
			}
			outputs["vault_address"] = vault.VaultAddress

		case "redis":
			redis, err := components.DeployRedis(ctx, cfg, environment)
			if err != nil {
				return nil, err
			}
			outputs["redis_endpoint"] = redis.Endpoint

		case "rabbitmq":
			rabbitmq, err := components.DeployRabbitMQ(ctx, cfg, environment)
			if err != nil {
				return nil, err
			}
			outputs["rabbitmq_endpoint"] = rabbitmq.Endpoint

		case "observability":
			observability, err := components.DeployObservability(ctx, cfg, environment)
			if err != nil {
				return nil, err
			}
			outputs["grafana_url"] = observability.GrafanaURL

		case "dapr":
			dapr, err := components.DeployDapr(ctx, cfg, environment)
			if err != nil {
				return nil, err
			}
			outputs["dapr_control_plane_url"] = dapr.ControlPlaneURL

		case "services":
			services, err := components.DeployServices(ctx, cfg, environment)
			if err != nil {
				return nil, err
			}
			outputs["public_gateway_url"] = services.PublicGatewayURL
			outputs["admin_gateway_url"] = services.AdminGatewayURL

		case "website":
			website, err := components.DeployWebsite(ctx, cfg, environment)
			if err != nil {
				return nil, err
			}
			outputs["website_url"] = website.ServerURL
		}
	}

	// Add environment to outputs
	outputs["environment"] = environment

	ctx.Log.Info(environment+" infrastructure deployment completed successfully", nil)
	return outputs, nil
}