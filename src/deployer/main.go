package main

import (
	"context"
	"fmt"

	"github.com/axiom-software-co/international-center/src/deployer/development/infrastructure"
	"github.com/axiom-software-co/international-center/src/deployer/shared/config"
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
		
		// Get environment and network name from config manager
		environment := string(configManager.GetEnvironment())
		networkName := configManager.GetNetworkName()

		// Deploy Database Stack (PostgreSQL)
		databaseStack := infrastructure.NewDatabaseStack(ctx, configManager.GetPulumiConfig().GetUnderlyingConfig(), networkName, environment)
		databaseDeployment, err := databaseStack.Deploy(context.Background())
		if err != nil {
			return fmt.Errorf("failed to deploy database stack: %w", err)
		}

		// Deploy Dapr Stack (Redis + Dapr control plane)
		daprStack := infrastructure.NewDaprStack(ctx, configManager.GetPulumiConfig().GetUnderlyingConfig(), networkName, environment)
		daprDeployment, err := daprStack.Deploy(context.Background())
		if err != nil {
			return fmt.Errorf("failed to deploy Dapr stack: %w", err)
		}

		// Deploy Storage Stack (Azurite)
		storageStack := infrastructure.NewStorageStack(ctx, configManager.GetPulumiConfig().GetUnderlyingConfig(), networkName, environment)
		storageDeployment, err := storageStack.Deploy(context.Background())
		if err != nil {
			return fmt.Errorf("failed to deploy storage stack: %w", err)
		}

		// Deploy Vault Stack (HashiCorp Vault)
		vaultStack := infrastructure.NewVaultStack(ctx, configManager.GetPulumiConfig().GetUnderlyingConfig(), networkName, environment)
		vaultDeployment, err := vaultStack.Deploy(context.Background())
		if err != nil {
			return fmt.Errorf("failed to deploy Vault stack: %w", err)
		}

		// Deploy Observability Stack (Grafana + Loki + Prometheus)
		observabilityStack := infrastructure.NewObservabilityStack(ctx, configManager.GetPulumiConfig().GetUnderlyingConfig(), networkName, environment)
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

		// Service endpoints (these would be set by the service deployment)
		ctx.Export("SERVICE_HOST", pulumi.String("localhost"))
		ctx.Export("CONTENT_API_URL", pulumi.String("http://localhost:8081"))
		ctx.Export("SERVICES_API_URL", pulumi.String("http://localhost:8082"))
		ctx.Export("PUBLIC_GATEWAY_URL", pulumi.String("http://localhost:8080"))
		ctx.Export("ADMIN_GATEWAY_URL", pulumi.String("http://localhost:8090"))

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