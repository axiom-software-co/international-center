package main

import (
	"fmt"
	
	"github.com/axiom-software-co/international-center/src/cicd/shared"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "international-center-cicd")
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

// deployDevelopmentInfrastructure deploys complete infrastructure for development environment using orchestration
func deployDevelopmentInfrastructure(ctx *pulumi.Context, cfg *config.Config) error {
	environment := "development"
	
	// Create deployment orchestrator
	orchestrator := shared.NewDeploymentOrchestrator(ctx, cfg, environment)
	
	// Deploy infrastructure with orchestration and health monitoring
	outputs, err := orchestrator.DeployInfrastructure()
	if err != nil {
		return err
	}
	
	// Export key outputs for development
	ctx.Export("environment", pulumi.String(environment))
	ctx.Export("database_connection_string", outputs.Database.ConnectionString)
	ctx.Export("storage_connection_string", outputs.Storage.ConnectionString)
	ctx.Export("vault_address", outputs.Vault.VaultAddress)
	ctx.Export("grafana_url", outputs.Observability.GrafanaURL)
	ctx.Export("dapr_control_plane_url", outputs.Dapr.ControlPlaneURL)
	ctx.Export("public_gateway_url", outputs.Services.PublicGatewayURL)
	ctx.Export("admin_gateway_url", outputs.Services.AdminGatewayURL)
	ctx.Export("website_url", outputs.Website.ServerURL)
	
	// Export deployment health status
	health := orchestrator.GetDeploymentHealth(outputs)
	for component, healthy := range health {
		ctx.Export(fmt.Sprintf("%s_healthy", component), pulumi.Bool(healthy))
	}

	ctx.Log.Info("Development infrastructure deployment completed successfully", nil)
	return nil
}

func deployStagingInfrastructure(ctx *pulumi.Context, cfg *config.Config) error {
	environment := "staging"
	
	// Create deployment orchestrator
	orchestrator := shared.NewDeploymentOrchestrator(ctx, cfg, environment)
	
	// Deploy infrastructure with orchestration and health monitoring
	outputs, err := orchestrator.DeployInfrastructure()
	if err != nil {
		return err
	}
	
	// Export key outputs for staging
	ctx.Export("environment", pulumi.String(environment))
	ctx.Export("database_connection_string", outputs.Database.ConnectionString)
	ctx.Export("storage_connection_string", outputs.Storage.ConnectionString)
	ctx.Export("vault_address", outputs.Vault.VaultAddress)
	ctx.Export("grafana_url", outputs.Observability.GrafanaURL)
	ctx.Export("dapr_control_plane_url", outputs.Dapr.ControlPlaneURL)
	ctx.Export("public_gateway_url", outputs.Services.PublicGatewayURL)
	ctx.Export("admin_gateway_url", outputs.Services.AdminGatewayURL)
	ctx.Export("website_url", outputs.Website.ServerURL)
	
	// Export deployment health status
	health := orchestrator.GetDeploymentHealth(outputs)
	for component, healthy := range health {
		ctx.Export(fmt.Sprintf("%s_healthy", component), pulumi.Bool(healthy))
	}

	ctx.Log.Info("Staging infrastructure deployment completed successfully", nil)
	return nil
}

func deployProductionInfrastructure(ctx *pulumi.Context, cfg *config.Config) error {
	environment := "production"
	
	// Create deployment orchestrator
	orchestrator := shared.NewDeploymentOrchestrator(ctx, cfg, environment)
	
	// Deploy infrastructure with orchestration and health monitoring
	outputs, err := orchestrator.DeployInfrastructure()
	if err != nil {
		return err
	}
	
	// Export key outputs for production
	ctx.Export("environment", pulumi.String(environment))
	ctx.Export("database_connection_string", outputs.Database.ConnectionString)
	ctx.Export("storage_connection_string", outputs.Storage.ConnectionString)
	ctx.Export("vault_address", outputs.Vault.VaultAddress)
	ctx.Export("grafana_url", outputs.Observability.GrafanaURL)
	ctx.Export("dapr_control_plane_url", outputs.Dapr.ControlPlaneURL)
	ctx.Export("public_gateway_url", outputs.Services.PublicGatewayURL)
	ctx.Export("admin_gateway_url", outputs.Services.AdminGatewayURL)
	ctx.Export("website_url", outputs.Website.ServerURL)
	
	// Export deployment health status
	health := orchestrator.GetDeploymentHealth(outputs)
	for component, healthy := range health {
		ctx.Export(fmt.Sprintf("%s_healthy", component), pulumi.Bool(healthy))
	}

	ctx.Log.Info("Production infrastructure deployment completed successfully", nil)
	return nil
}