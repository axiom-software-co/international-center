package main

import (
	"github.com/axiom-software-co/international-center/src/cicd/components"
	"github.com/axiom-software-co/international-center/src/cicd/shared"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		environment := cfg.Require("environment")
		
		ctx.Log.Info("Starting International Center infrastructure deployment", nil)
		ctx.Log.Info("Environment: "+environment, nil)
		
		// TODO: Component deployment will be implemented during TDD Phase 2
		// This placeholder demonstrates single-program architecture
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

// deployDevelopmentInfrastructure deploys complete infrastructure for development environment
func deployDevelopmentInfrastructure(ctx *pulumi.Context, cfg *config.Config) error {
	environment := "development"
	
	// Load and validate environment configuration
	_, err := shared.LoadEnvironmentConfig(ctx, environment)
	if err != nil {
		return err
	}
	
	// Deploy components in dependency order
	ctx.Log.Info("Deploying database component", nil)
	databaseOutputs, err := components.DeployDatabase(ctx, cfg, environment)
	if err != nil {
		return err
	}

	ctx.Log.Info("Deploying storage component", nil)
	storageOutputs, err := components.DeployStorage(ctx, cfg, environment)
	if err != nil {
		return err
	}

	ctx.Log.Info("Deploying vault component", nil)
	vaultOutputs, err := components.DeployVault(ctx, cfg, environment)
	if err != nil {
		return err
	}

	ctx.Log.Info("Deploying observability component", nil)
	observabilityOutputs, err := components.DeployObservability(ctx, cfg, environment)
	if err != nil {
		return err
	}

	ctx.Log.Info("Deploying dapr component", nil)
	daprOutputs, err := components.DeployDapr(ctx, cfg, environment)
	if err != nil {
		return err
	}

	ctx.Log.Info("Deploying services component", nil)
	_, err = components.DeployServices(ctx, cfg, environment)
	if err != nil {
		return err
	}

	ctx.Log.Info("Deploying website component", nil)
	websiteOutputs, err := components.DeployWebsite(ctx, cfg, environment)
	if err != nil {
		return err
	}

	// Export key outputs for development
	ctx.Export("environment", pulumi.String(environment))
	ctx.Export("database_connection_string", databaseOutputs.ConnectionString)
	ctx.Export("storage_connection_string", storageOutputs.ConnectionString)
	ctx.Export("vault_address", vaultOutputs.VaultAddress)
	ctx.Export("grafana_url", observabilityOutputs.GrafanaURL)
	ctx.Export("dapr_control_plane_url", daprOutputs.ControlPlaneURL)
	ctx.Export("website_url", websiteOutputs.ServerURL)

	ctx.Log.Info("Development infrastructure deployment completed successfully", nil)
	return nil
}

func deployStagingInfrastructure(ctx *pulumi.Context, cfg *config.Config) error {
	environment := "staging"
	
	// Load and validate environment configuration
	_, err := shared.LoadEnvironmentConfig(ctx, environment)
	if err != nil {
		return err
	}
	
	// Load configuration for validation
	envConfig, err := shared.LoadEnvironmentConfig(ctx, environment)
	if err != nil {
		return err
	}
	
	// Validate required configuration for staging
	if err := shared.ValidateConfiguration(envConfig, environment); err != nil {
		return err
	}

	// Deploy components in dependency order
	ctx.Log.Info("Deploying database component", nil)
	databaseOutputs, err := components.DeployDatabase(ctx, cfg, environment)
	if err != nil {
		return err
	}

	ctx.Log.Info("Deploying storage component", nil)
	storageOutputs, err := components.DeployStorage(ctx, cfg, environment)
	if err != nil {
		return err
	}

	ctx.Log.Info("Deploying vault component", nil)
	vaultOutputs, err := components.DeployVault(ctx, cfg, environment)
	if err != nil {
		return err
	}

	ctx.Log.Info("Deploying observability component", nil)
	observabilityOutputs, err := components.DeployObservability(ctx, cfg, environment)
	if err != nil {
		return err
	}

	ctx.Log.Info("Deploying dapr component", nil)
	daprOutputs, err := components.DeployDapr(ctx, cfg, environment)
	if err != nil {
		return err
	}

	ctx.Log.Info("Deploying services component", nil)
	_, err = components.DeployServices(ctx, cfg, environment)
	if err != nil {
		return err
	}

	ctx.Log.Info("Deploying website component", nil)
	websiteOutputs, err := components.DeployWebsite(ctx, cfg, environment)
	if err != nil {
		return err
	}

	// Export key outputs for staging
	ctx.Export("environment", pulumi.String(environment))
	ctx.Export("database_connection_string", databaseOutputs.ConnectionString)
	ctx.Export("storage_connection_string", storageOutputs.ConnectionString)
	ctx.Export("vault_address", vaultOutputs.VaultAddress)
	ctx.Export("grafana_url", observabilityOutputs.GrafanaURL)
	ctx.Export("dapr_control_plane_url", daprOutputs.ControlPlaneURL)
	ctx.Export("website_url", websiteOutputs.ServerURL)

	ctx.Log.Info("Staging infrastructure deployment completed successfully", nil)
	return nil
}

func deployProductionInfrastructure(ctx *pulumi.Context, cfg *config.Config) error {
	environment := "production"
	
	// Load and validate environment configuration
	envConfig, err := shared.LoadEnvironmentConfig(ctx, environment)
	if err != nil {
		return err
	}
	
	// Validate required configuration for production
	if err := shared.ValidateConfiguration(envConfig, environment); err != nil {
		return err
	}

	// Deploy components in dependency order with production policies
	ctx.Log.Info("Deploying database component", nil)
	databaseOutputs, err := components.DeployDatabase(ctx, cfg, environment)
	if err != nil {
		return err
	}

	ctx.Log.Info("Deploying storage component", nil)
	storageOutputs, err := components.DeployStorage(ctx, cfg, environment)
	if err != nil {
		return err
	}

	ctx.Log.Info("Deploying vault component", nil)
	vaultOutputs, err := components.DeployVault(ctx, cfg, environment)
	if err != nil {
		return err
	}

	ctx.Log.Info("Deploying observability component", nil)
	observabilityOutputs, err := components.DeployObservability(ctx, cfg, environment)
	if err != nil {
		return err
	}

	ctx.Log.Info("Deploying dapr component", nil)
	daprOutputs, err := components.DeployDapr(ctx, cfg, environment)
	if err != nil {
		return err
	}

	ctx.Log.Info("Deploying services component", nil)
	_, err = components.DeployServices(ctx, cfg, environment)
	if err != nil {
		return err
	}

	ctx.Log.Info("Deploying website component", nil)
	websiteOutputs, err := components.DeployWebsite(ctx, cfg, environment)
	if err != nil {
		return err
	}

	// Export key outputs for production
	ctx.Export("environment", pulumi.String(environment))
	ctx.Export("database_connection_string", databaseOutputs.ConnectionString)
	ctx.Export("storage_connection_string", storageOutputs.ConnectionString)
	ctx.Export("vault_address", vaultOutputs.VaultAddress)
	ctx.Export("grafana_url", observabilityOutputs.GrafanaURL)
	ctx.Export("dapr_control_plane_url", daprOutputs.ControlPlaneURL)
	ctx.Export("website_url", websiteOutputs.ServerURL)

	ctx.Log.Info("Production infrastructure deployment completed successfully", nil)
	return nil
}