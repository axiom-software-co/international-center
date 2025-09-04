package main

import (
	"context"
	"fmt"

	devinfra "github.com/axiom-software-co/international-center/src/deployer/development/infrastructure"
	prodinfra "github.com/axiom-software-co/international-center/src/deployer/production/infrastructure"
	staginginfra "github.com/axiom-software-co/international-center/src/deployer/staging/infrastructure"
	"github.com/axiom-software-co/international-center/src/deployer/shared/config"
	sharedinfra "github.com/axiom-software-co/international-center/src/deployer/shared/infrastructure"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Initialize centralized configuration manager
		configManager, err := config.NewConfigManager(ctx)
		if err != nil {
			return fmt.Errorf("failed to initialize configuration manager: %w", err)
		}
		
		// Validate configuration
		if err := configManager.ValidateConfiguration(); err != nil {
			return fmt.Errorf("configuration validation failed: %w", err)
		}
		
		// Get environment from config manager
		environment := string(configManager.GetEnvironment())
		pulumiConfig := configManager.GetPulumiConfig().GetUnderlyingConfig()

		// Select infrastructure factory based on environment
		var factory sharedinfra.InfrastructureFactory
		switch configManager.GetEnvironment() {
		case config.EnvironmentDevelopment:
			factory = devinfra.NewDevelopmentInfrastructureFactory()
		case config.EnvironmentStaging:
			factory = staginginfra.NewStagingInfrastructureFactory()
		case config.EnvironmentProduction:
			factory = prodinfra.NewProductionInfrastructureFactory()
		default:
			return fmt.Errorf("unsupported environment: %s", environment)
		}

		// Deploy Database Stack using factory
		databaseStack := factory.CreateDatabaseStack(ctx, pulumiConfig, environment)
		databaseDeployment, err := databaseStack.Deploy(context.Background())
		if err != nil {
			return fmt.Errorf("failed to deploy database stack: %w", err)
		}

		// Deploy Dapr Stack using factory
		daprStack := factory.CreateDaprStack(ctx, pulumiConfig, environment)
		daprDeployment, err := daprStack.Deploy(context.Background())
		if err != nil {
			return fmt.Errorf("failed to deploy Dapr stack: %w", err)
		}

		// Deploy Storage Stack using factory
		storageStack := factory.CreateStorageStack(ctx, pulumiConfig, environment)
		storageDeployment, err := storageStack.Deploy(context.Background())
		if err != nil {
			return fmt.Errorf("failed to deploy storage stack: %w", err)
		}

		// Deploy Vault Stack using factory
		vaultStack := factory.CreateVaultStack(ctx, pulumiConfig, environment)
		vaultDeployment, err := vaultStack.Deploy(context.Background())
		if err != nil {
			return fmt.Errorf("failed to deploy Vault stack: %w", err)
		}

		// Deploy Observability Stack using factory
		observabilityStack := factory.CreateObservabilityStack(ctx, pulumiConfig, environment)
		observabilityDeployment, err := observabilityStack.Deploy(context.Background())
		if err != nil {
			return fmt.Errorf("failed to deploy observability stack: %w", err)
		}

		// Export environment variables for integration tests
		dbHost, dbPort, dbName, dbUser := databaseStack.GetConnectionInfo()
		ctx.Export("DATABASE_URL", pulumi.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
			dbUser, configManager.GetPulumiConfig().GetRequiredString("postgres_password"), dbHost, dbPort, dbName))

		redisHost, redisPort, redisPassword := daprStack.GetRedisConnectionInfo()
		ctx.Export("REDIS_URL", pulumi.Sprintf("redis://%s:%d", redisHost, redisPort))
		ctx.Export("REDIS_PORT", pulumi.Int(redisPort))
		ctx.Export("REDIS_PASSWORD", pulumi.String(redisPassword))

		vaultURL, vaultToken, _ := vaultStack.GetVaultConnectionInfo()
		ctx.Export("VAULT_URL", pulumi.String(vaultURL))
		ctx.Export("VAULT_TOKEN", pulumi.String(vaultToken))

		azuriteURL := storageStack.GetBlobStorageEndpoint()
		ctx.Export("AZURITE_URL", pulumi.String(azuriteURL))

		observabilityEndpoints := observabilityStack.GetObservabilityEndpoints()
		ctx.Export("GRAFANA_URL", pulumi.String(observabilityEndpoints["grafana"]))
		ctx.Export("LOKI_URL", pulumi.String(observabilityEndpoints["loki"]))
		ctx.Export("PROMETHEUS_URL", pulumi.String(observabilityEndpoints["prometheus"]))

		// Dapr endpoints
		daprEndpoints := daprStack.GetDaprEndpoints()
		ctx.Export("DAPR_HTTP_ENDPOINT", pulumi.String(daprEndpoints["http"]))
		ctx.Export("DAPR_GRPC_ENDPOINT", pulumi.String(daprEndpoints["grpc"]))
		ctx.Export("DAPR_PLACEMENT_ENDPOINT", pulumi.String(daprEndpoints["placement"]))

		// Service endpoints from configuration manager
		serviceConfig := configManager.GetServiceConfig()
		ctx.Export("SERVICE_HOST", pulumi.String(serviceConfig.Host))
		ctx.Export("CONTENT_API_URL", pulumi.String(configManager.GetPulumiConfig().GetString("content_api_url")))
		ctx.Export("SERVICES_API_URL", pulumi.String(configManager.GetPulumiConfig().GetString("services_api_url")))
		ctx.Export("PUBLIC_GATEWAY_URL", pulumi.String(configManager.GetPulumiConfig().GetString("public_gateway_url")))
		ctx.Export("ADMIN_GATEWAY_URL", pulumi.String(configManager.GetPulumiConfig().GetString("admin_gateway_url")))

		// Generate Dapr components after all infrastructure is deployed
		err = daprStack.GenerateDaprComponents(context.Background(), daprDeployment)
		if err != nil {
			return fmt.Errorf("failed to generate Dapr components: %w", err)
		}

		// Initialize Vault secrets after deployment
		err = vaultStack.InitializeSecrets(context.Background(), vaultDeployment)
		if err != nil {
			return fmt.Errorf("failed to initialize Vault secrets: %w", err)
		}

		// Create Vault policies
		err = vaultStack.CreatePolicies(context.Background(), vaultDeployment)
		if err != nil {
			return fmt.Errorf("failed to create Vault policies: %w", err)
		}

		// Configure observability data sources
		err = observabilityStack.ConfigureDataSources(context.Background(), observabilityDeployment)
		if err != nil {
			return fmt.Errorf("failed to configure observability data sources: %w", err)
		}

		// Create storage containers
		err = storageStack.CreateStorageContainers(context.Background(), storageDeployment)
		if err != nil {
			return fmt.Errorf("failed to create storage containers: %w", err)
		}

		// Validate all deployments
		if err := databaseStack.ValidateDeployment(context.Background(), databaseDeployment); err != nil {
			return fmt.Errorf("database deployment validation failed: %w", err)
		}

		if err := daprStack.ValidateDeployment(context.Background(), daprDeployment); err != nil {
			return fmt.Errorf("Dapr deployment validation failed: %w", err)
		}

		if err := storageStack.ValidateDeployment(context.Background(), storageDeployment); err != nil {
			return fmt.Errorf("storage deployment validation failed: %w", err)
		}

		if err := vaultStack.ValidateDeployment(context.Background(), vaultDeployment); err != nil {
			return fmt.Errorf("Vault deployment validation failed: %w", err)
		}

		if err := observabilityStack.ValidateDeployment(context.Background(), observabilityDeployment); err != nil {
			return fmt.Errorf("observability deployment validation failed: %w", err)
		}

		return nil
	})
}