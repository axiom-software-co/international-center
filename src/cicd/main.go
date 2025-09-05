package main

import (
	"github.com/axiom-software-co/international-center/src/cicd/components"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "international-center-cicd")
		environment := cfg.Require("environment")
		
		ctx.Log.Info("Starting International Center infrastructure deployment", nil)
		ctx.Log.Info("Environment: "+environment, nil)
		
		switch environment {
		case "development":
			return deployDevelopmentInfrastructure(ctx, cfg)
		case "staging":
			return deployStagingInfrastructure(ctx, cfg)
		case "production":
			return deployProductionInfrastructure(ctx, cfg)
		default:
			ctx.Log.Error("Unknown environment: "+environment, nil)
			return nil
		}
	})
}

// deployDevelopmentInfrastructure deploys infrastructure components sequentially for development
func deployDevelopmentInfrastructure(ctx *pulumi.Context, cfg *config.Config) error {
	environment := "development"
	
	// 1. Deploy database
	ctx.Log.Info("Deploying database component", nil)
	database, err := components.DeployDatabase(ctx, cfg, environment)
	if err != nil {
		return err
	}
	
	// 2. Deploy storage
	ctx.Log.Info("Deploying storage component", nil)
	storage, err := components.DeployStorage(ctx, cfg, environment)
	if err != nil {
		return err
	}
	
	// 3. Deploy vault
	ctx.Log.Info("Deploying vault component", nil)
	vault, err := components.DeployVault(ctx, cfg, environment)
	if err != nil {
		return err
	}
	
	// 4. Deploy redis (caching)
	ctx.Log.Info("Deploying redis component", nil)
	redis, err := components.DeployRedis(ctx, cfg, environment)
	if err != nil {
		return err
	}
	
	// 5. Deploy rabbitmq (pub/sub)
	ctx.Log.Info("Deploying rabbitmq component", nil)
	rabbitmq, err := components.DeployRabbitMQ(ctx, cfg, environment)
	if err != nil {
		return err
	}
	
	// 6. Deploy observability
	ctx.Log.Info("Deploying observability component", nil)
	observability, err := components.DeployObservability(ctx, cfg, environment)
	if err != nil {
		return err
	}
	
	// 7. Deploy dapr
	ctx.Log.Info("Deploying dapr component", nil)
	dapr, err := components.DeployDapr(ctx, cfg, environment)
	if err != nil {
		return err
	}
	
	// 8. Deploy services (requires dapr)
	ctx.Log.Info("Deploying services component", nil)
	services, err := components.DeployServices(ctx, cfg, environment)
	if err != nil {
		return err
	}
	
	// 9. Deploy website (requires services)
	ctx.Log.Info("Deploying website component", nil)
	website, err := components.DeployWebsite(ctx, cfg, environment)
	if err != nil {
		return err
	}
	
	// Export outputs
	ctx.Export("environment", pulumi.String(environment))
	ctx.Export("database_connection_string", database.ConnectionString)
	ctx.Export("storage_connection_string", storage.ConnectionString)
	ctx.Export("vault_address", vault.VaultAddress)
	ctx.Export("redis_endpoint", redis.Endpoint)
	ctx.Export("rabbitmq_endpoint", rabbitmq.Endpoint)
	ctx.Export("grafana_url", observability.GrafanaURL)
	ctx.Export("dapr_control_plane_url", dapr.ControlPlaneURL)
	ctx.Export("public_gateway_url", services.PublicGatewayURL)
	ctx.Export("admin_gateway_url", services.AdminGatewayURL)
	ctx.Export("website_url", website.ServerURL)

	ctx.Log.Info("Development infrastructure deployment completed successfully", nil)
	return nil
}

func deployStagingInfrastructure(ctx *pulumi.Context, cfg *config.Config) error {
	environment := "staging"
	
	// Sequential deployment for staging
	database, err := components.DeployDatabase(ctx, cfg, environment)
	if err != nil {
		return err
	}
	
	storage, err := components.DeployStorage(ctx, cfg, environment)
	if err != nil {
		return err
	}
	
	vault, err := components.DeployVault(ctx, cfg, environment)
	if err != nil {
		return err
	}
	
	redis, err := components.DeployRedis(ctx, cfg, environment)
	if err != nil {
		return err
	}
	
	rabbitmq, err := components.DeployRabbitMQ(ctx, cfg, environment)
	if err != nil {
		return err
	}
	
	observability, err := components.DeployObservability(ctx, cfg, environment)
	if err != nil {
		return err
	}
	
	dapr, err := components.DeployDapr(ctx, cfg, environment)
	if err != nil {
		return err
	}
	
	services, err := components.DeployServices(ctx, cfg, environment)
	if err != nil {
		return err
	}
	
	website, err := components.DeployWebsite(ctx, cfg, environment)
	if err != nil {
		return err
	}
	
	// Export outputs
	ctx.Export("environment", pulumi.String(environment))
	ctx.Export("database_connection_string", database.ConnectionString)
	ctx.Export("storage_connection_string", storage.ConnectionString)
	ctx.Export("vault_address", vault.VaultAddress)
	ctx.Export("redis_endpoint", redis.Endpoint)
	ctx.Export("rabbitmq_endpoint", rabbitmq.Endpoint)
	ctx.Export("grafana_url", observability.GrafanaURL)
	ctx.Export("dapr_control_plane_url", dapr.ControlPlaneURL)
	ctx.Export("public_gateway_url", services.PublicGatewayURL)
	ctx.Export("admin_gateway_url", services.AdminGatewayURL)
	ctx.Export("website_url", website.ServerURL)

	ctx.Log.Info("Staging infrastructure deployment completed successfully", nil)
	return nil
}

func deployProductionInfrastructure(ctx *pulumi.Context, cfg *config.Config) error {
	environment := "production"
	
	// Sequential deployment for production
	database, err := components.DeployDatabase(ctx, cfg, environment)
	if err != nil {
		return err
	}
	
	storage, err := components.DeployStorage(ctx, cfg, environment)
	if err != nil {
		return err
	}
	
	vault, err := components.DeployVault(ctx, cfg, environment)
	if err != nil {
		return err
	}
	
	redis, err := components.DeployRedis(ctx, cfg, environment)
	if err != nil {
		return err
	}
	
	rabbitmq, err := components.DeployRabbitMQ(ctx, cfg, environment)
	if err != nil {
		return err
	}
	
	observability, err := components.DeployObservability(ctx, cfg, environment)
	if err != nil {
		return err
	}
	
	dapr, err := components.DeployDapr(ctx, cfg, environment)
	if err != nil {
		return err
	}
	
	services, err := components.DeployServices(ctx, cfg, environment)
	if err != nil {
		return err
	}
	
	website, err := components.DeployWebsite(ctx, cfg, environment)
	if err != nil {
		return err
	}
	
	// Export outputs
	ctx.Export("environment", pulumi.String(environment))
	ctx.Export("database_connection_string", database.ConnectionString)
	ctx.Export("storage_connection_string", storage.ConnectionString)
	ctx.Export("vault_address", vault.VaultAddress)
	ctx.Export("redis_endpoint", redis.Endpoint)
	ctx.Export("rabbitmq_endpoint", rabbitmq.Endpoint)
	ctx.Export("grafana_url", observability.GrafanaURL)
	ctx.Export("dapr_control_plane_url", dapr.ControlPlaneURL)
	ctx.Export("public_gateway_url", services.PublicGatewayURL)
	ctx.Export("admin_gateway_url", services.AdminGatewayURL)
	ctx.Export("website_url", website.ServerURL)

	ctx.Log.Info("Production infrastructure deployment completed successfully", nil)
	return nil
}