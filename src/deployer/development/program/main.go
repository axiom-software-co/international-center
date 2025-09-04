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
		cfg := config.New(ctx, "")
		environment := "development"
		
		// Create the development infrastructure factory
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
		
		// Export outputs required for integration tests contract validation
		ctx.Export("environment", pulumi.String(environment))
		
		// Database component outputs
		ctx.Export("database_resource_id", databaseDeployment.GetPrimaryEndpoint())
		ctx.Export("database_name", pulumi.String("development-postgresql"))
		ctx.Export("database_connection_string", databaseStack.GetConnectionString())
		
		// Storage component outputs
		ctx.Export("storage_resource_id", storageDeployment.GetPrimaryStorageEndpoint())
		
		// Vault component outputs
		ctx.Export("vault_resource_id", vaultDeployment.GetVaultEndpoint())
		
		// Dapr component outputs
		componentNames := pulumi.StringArray{
			pulumi.String("statestore"),
			pulumi.String("pubsub"),
			pulumi.String("blob-storage"),
			pulumi.String("secretstore"),
		}
		ctx.Export("dapr_component_names", componentNames)
		
		// Additional component-first architecture outputs for comprehensive validation
		ctx.Export("database_endpoint", databaseDeployment.GetPrimaryEndpoint())
		ctx.Export("storage_endpoint", pulumi.String(storageStack.GetBlobStorageEndpoint()))
		ctx.Export("vault_endpoint", vaultDeployment.GetVaultEndpoint())
		
		// Cross-component communication validation outputs
		ctx.Export("database_network_id", databaseDeployment.GetNetworkResources().GetNetworkID())
		
		// Resource dependency validation outputs
		ctx.Export("component_deployment_order", pulumi.StringArray{
			pulumi.String("database"),
			pulumi.String("storage"), 
			pulumi.String("vault"),
			pulumi.String("dapr"),
		})
		
		return nil
	})
}