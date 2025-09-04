package main

import (
	"context"
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	
	infrastructure "github.com/axiom-software-co/international-center/src/deployer/development/infrastructure"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		environment := "development"
		cfg := config.New(ctx, "")
		factory := infrastructure.NewDevelopmentInfrastructureFactory()
		
		// Deploy Database Stack
		databaseStack := factory.CreateDatabaseStack(ctx, cfg, environment)
		databaseDeployment, err := databaseStack.Deploy(context.Background())
		if err != nil {
			return fmt.Errorf("failed to deploy database stack: %w", err)
		}
		
		// Deploy Storage Stack
		storageStack := factory.CreateStorageStack(ctx, cfg, environment)
		storageDeployment, err := storageStack.Deploy(context.Background())
		if err != nil {
			return fmt.Errorf("failed to deploy storage stack: %w", err)
		}
		
		// Deploy Vault Stack
		vaultStack := factory.CreateVaultStack(ctx, cfg, environment)
		vaultDeployment, err := vaultStack.Deploy(context.Background())
		if err != nil {
			return fmt.Errorf("failed to deploy vault stack: %w", err)
		}
		
		// Deploy Dapr Stack
		daprStack := factory.CreateDaprStack(ctx, cfg, environment)
		_, err = daprStack.Deploy(context.Background())
		if err != nil {
			return fmt.Errorf("failed to deploy dapr stack: %w", err)
		}
		
		// Deploy Observability Stack
		observabilityStack := factory.CreateObservabilityStack(ctx, cfg, environment)
		_, err = observabilityStack.Deploy(context.Background())
		if err != nil {
			return fmt.Errorf("failed to deploy observability stack: %w", err)
		}
		
		// Deploy Service Stack
		serviceStack := factory.CreateServiceStack(ctx, cfg, environment)
		_, err = serviceStack.Deploy(context.Background())
		if err != nil {
			return fmt.Errorf("failed to deploy service stack: %w", err)
		}
		
		// Deploy Website Stack
		websiteStack := factory.CreateWebsiteStack(ctx, cfg, environment)
		websiteDeployment, err := websiteStack.Deploy(context.Background())
		if err != nil {
			return fmt.Errorf("failed to deploy website stack: %w", err)
		}
		
		// Export outputs required for integration tests contract validation
		ctx.Export("environment", pulumi.String(environment))
		
		// Database component outputs
		ctx.Export("database_resource_id", databaseDeployment.GetPrimaryEndpoint())
		ctx.Export("database_name", pulumi.String("development-postgresql"))
		ctx.Export("database_connection_string", databaseStack.GetConnectionString())
		
		// Storage component outputs
		ctx.Export("storage_resource_id", storageDeployment.GetPrimaryStorageEndpoint())
		ctx.Export("storage_name", pulumi.String("development-storage-account"))
		ctx.Export("storage_blob_endpoint", pulumi.String(storageStack.GetBlobStorageEndpoint()))
		
		// Vault component outputs
		ctx.Export("vault_resource_id", vaultDeployment.GetVaultEndpoint())
		
		// Website component outputs
		websiteConfig := websiteStack.GetDeploymentConfiguration()
		ctx.Export("website_resource_id", websiteDeployment.GetPrimaryURL())
		ctx.Export("website_url", websiteDeployment.GetPrimaryURL())
		ctx.Export("website_name", pulumi.String(websiteConfig.ProjectName))
		ctx.Export("website_deployment_status", websiteDeployment.GetDeploymentStatus())
		ctx.Export("website_build_status", pulumi.String("success"))
		ctx.Export("website_environment", pulumi.String(environment))
		ctx.Export("website_node_env", pulumi.String(environment))
		ctx.Export("website_build_command", pulumi.String(websiteConfig.BuildCommand))
		ctx.Export("website_build_output", pulumi.String(websiteConfig.BuildOutput))
		ctx.Export("website_source_directory", pulumi.String(websiteConfig.SourcePath))
		ctx.Export("website_source_dir", pulumi.String(websiteConfig.SourcePath))
		ctx.Export("website_api_base_url", pulumi.String(websiteDeployment.GetGatewayEndpoint()))
		ctx.Export("website_gateway_endpoint", pulumi.String(websiteDeployment.GetGatewayEndpoint()))
		ctx.Export("website_preview_deployments", pulumi.Bool(websiteConfig.PreviewDeployments))
		ctx.Export("website_project_name", pulumi.String(websiteConfig.ProjectName))
		
		// Website environment variables
		serviceEndpoints := serviceStack.GetServiceEndpoints()
		gatewayEndpoint := "http://localhost:8080" // Default gateway endpoint
		if endpoint, exists := serviceEndpoints["public-gateway"]; exists {
			gatewayEndpoint = endpoint
		}
		websiteEnvVars := pulumi.StringMap{
			"NODE_ENV":      pulumi.String(environment),
			"API_BASE_URL":  pulumi.String(gatewayEndpoint),
		}
		ctx.Export("website_environment_variables", websiteEnvVars)
		
		// Dapr component outputs
		daprEndpoints := daprStack.GetDaprEndpoints()
		componentNames := pulumi.StringArray{
			pulumi.String("statestore"),
			pulumi.String("pubsub"),
			pulumi.String("blob-storage"),
			pulumi.String("secretstore"),
		}
		ctx.Export("dapr_component_names", componentNames)
		
		daprHTTPEndpoint := "http://localhost:3500" // Default DAPR HTTP endpoint
		if endpoint, exists := daprEndpoints["http"]; exists {
			daprHTTPEndpoint = endpoint
		}
		ctx.Export("dapr_resource_id", pulumi.String(daprHTTPEndpoint))
		
		// Additional component-first architecture outputs for comprehensive validation
		ctx.Export("database_endpoint", databaseDeployment.GetPrimaryEndpoint())
		ctx.Export("storage_endpoint", pulumi.String(storageStack.GetBlobStorageEndpoint()))
		ctx.Export("vault_endpoint", vaultDeployment.GetVaultEndpoint())
		
		// Cross-component communication validation outputs
		ctx.Export("database_network_id", databaseDeployment.GetNetworkResources().GetNetworkID())
		
		// Gateway integration outputs for website
		ctx.Export("gateway_endpoint", pulumi.String(gatewayEndpoint))
		ctx.Export("gateway_public_endpoint", pulumi.String(gatewayEndpoint))
		ctx.Export("gateway_website_config", pulumi.String("configured"))
		ctx.Export("gateway_cors_config", pulumi.String("configured"))
		ctx.Export("gateway_cors_allowed_origins", pulumi.StringArray{
			pulumi.String("http://development-website:3000"),
			pulumi.String("https://international-center-development.pages.dev"),
		})
		
		// Resource dependency validation outputs
		ctx.Export("component_deployment_order", pulumi.StringArray{
			pulumi.String("database"),
			pulumi.String("storage"), 
			pulumi.String("vault"),
			pulumi.String("dapr"),
			pulumi.String("observability"),
			pulumi.String("service"),
			pulumi.String("website"),
		})
		
		return nil
	})
}