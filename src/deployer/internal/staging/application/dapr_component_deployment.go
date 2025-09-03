package application

import (
	"context"
	"fmt"

	"github.com/pulumi/pulumi-azure-native/sdk/v2/go/azurenative/app"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/axiom-software-co/international-center/src/deployer/internal/staging/infrastructure"
)

type DaprComponentDeployment struct {
	containerAppsStack *infrastructure.AzureContainerAppsStack
	databaseStack     *infrastructure.AzureDatabaseStack
	storageStack      *infrastructure.AzureStorageStack
	vaultStack        *infrastructure.VaultCloudStack
	deployedComponents map[string]*app.DaprComponent
}

type ComponentConfiguration struct {
	Environment         string
	RedisHost          string
	RedisPassword      string
	VaultName          string
	StorageAccountName string
	ServiceBusConnection string
}

func NewDaprComponentDeployment(
	containerAppsStack *infrastructure.AzureContainerAppsStack,
	databaseStack *infrastructure.AzureDatabaseStack,
	storageStack *infrastructure.AzureStorageStack,
	vaultStack *infrastructure.VaultCloudStack,
) *DaprComponentDeployment {
	return &DaprComponentDeployment{
		containerAppsStack: containerAppsStack,
		databaseStack:     databaseStack,
		storageStack:      storageStack,
		vaultStack:        vaultStack,
		deployedComponents: make(map[string]*app.DaprComponent),
	}
}

func (deployment *DaprComponentDeployment) Deploy(ctx *pulumi.Context) error {
	config := deployment.getComponentConfiguration()

	components := []struct {
		name       string
		deployFunc func(*pulumi.Context, *ComponentConfiguration) error
	}{
		{"pubsub-redis", deployment.deployRedisPubSubComponent},
		{"statestore-redis", deployment.deployRedisStateStoreComponent},
		{"secretstore-keyvault", deployment.deployKeyVaultSecretStoreComponent},
		{"binding-storage", deployment.deployStorageBindingComponent},
		{"binding-servicebus", deployment.deployServiceBusBindingComponent},
		{"binding-cosmos", deployment.deployCosmosBindingComponent},
		{"middleware-ratelimit", deployment.deployRateLimitMiddleware},
		{"middleware-oauth", deployment.deployOAuthMiddleware},
		{"middleware-cors", deployment.deployCORSMiddleware},
	}

	for _, component := range components {
		if err := component.deployFunc(ctx, config); err != nil {
			return fmt.Errorf("failed to deploy %s component: %w", component.name, err)
		}
	}

	return nil
}

func (deployment *DaprComponentDeployment) getComponentConfiguration() *ComponentConfiguration {
	return &ComponentConfiguration{
		Environment:          "staging",
		RedisHost:           "redis-staging.redis.cache.windows.net:6380",
		RedisPassword:       "{vault:redis-connection-string}",
		VaultName:           "international-center-staging-kv",
		StorageAccountName:  "internationalcenterstaging",
		ServiceBusConnection: "{vault:servicebus-connection-string}",
	}
}

func (deployment *DaprComponentDeployment) deployRedisPubSubComponent(ctx *pulumi.Context, config *ComponentConfiguration) error {
	component, err := app.NewDaprComponent(ctx, "staging-pubsub-redis", &app.DaprComponentArgs{
		ResourceGroupName:    deployment.containerAppsStack.GetResourceGroup().Name,
		EnvironmentName:     deployment.containerAppsStack.GetEnvironment().Name,
		ComponentName:       pulumi.String("pubsub"),
		ComponentType:       pulumi.String("pubsub.redis"),
		Version:            pulumi.String("v1"),
		Metadata: app.DaprMetadataArray{
			&app.DaprMetadataArgs{
				Name:  pulumi.String("redisHost"),
				Value: pulumi.String(config.RedisHost),
			},
			&app.DaprMetadataArgs{
				Name:      pulumi.String("redisPassword"),
				SecretRef: pulumi.String("redis-password"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("enableTLS"),
				Value: pulumi.String("true"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("failover"),
				Value: pulumi.String("true"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("sentinelMasterName"),
				Value: pulumi.String("mymaster"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("maxRetries"),
				Value: pulumi.String("3"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("maxRetryBackoff"),
				Value: pulumi.String("2s"),
			},
		},
		Secrets: app.SecretArray{
			&app.SecretArgs{
				Name:        pulumi.String("redis-password"),
				KeyVaultUrl: deployment.vaultStack.GetVaultUri(),
				Identity:    pulumi.String("system"),
			},
		},
		Scopes: pulumi.StringArray{
			pulumi.String("identity-api"),
			pulumi.String("content-api"),
			pulumi.String("services-api"),
			pulumi.String("public-gateway"),
			pulumi.String("admin-gateway"),
		},
	})
	if err != nil {
		return err
	}

	deployment.deployedComponents["pubsub"] = component
	return nil
}

func (deployment *DaprComponentDeployment) deployRedisStateStoreComponent(ctx *pulumi.Context, config *ComponentConfiguration) error {
	component, err := app.NewDaprComponent(ctx, "staging-statestore-redis", &app.DaprComponentArgs{
		ResourceGroupName:    deployment.containerAppsStack.GetResourceGroup().Name,
		EnvironmentName:     deployment.containerAppsStack.GetEnvironment().Name,
		ComponentName:       pulumi.String("statestore"),
		ComponentType:       pulumi.String("state.redis"),
		Version:            pulumi.String("v1"),
		Metadata: app.DaprMetadataArray{
			&app.DaprMetadataArgs{
				Name:  pulumi.String("redisHost"),
				Value: pulumi.String(config.RedisHost),
			},
			&app.DaprMetadataArgs{
				Name:      pulumi.String("redisPassword"),
				SecretRef: pulumi.String("redis-password"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("enableTLS"),
				Value: pulumi.String("true"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("keyPrefix"),
				Value: pulumi.String("staging"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("actorStateStore"),
				Value: pulumi.String("true"),
			},
		},
		Secrets: app.SecretArray{
			&app.SecretArgs{
				Name:        pulumi.String("redis-password"),
				KeyVaultUrl: deployment.vaultStack.GetVaultUri(),
				Identity:    pulumi.String("system"),
			},
		},
		Scopes: pulumi.StringArray{
			pulumi.String("identity-api"),
			pulumi.String("content-api"),
			pulumi.String("services-api"),
		},
	})
	if err != nil {
		return err
	}

	deployment.deployedComponents["statestore"] = component
	return nil
}

func (deployment *DaprComponentDeployment) deployKeyVaultSecretStoreComponent(ctx *pulumi.Context, config *ComponentConfiguration) error {
	component, err := app.NewDaprComponent(ctx, "staging-secretstore-keyvault", &app.DaprComponentArgs{
		ResourceGroupName:    deployment.containerAppsStack.GetResourceGroup().Name,
		EnvironmentName:     deployment.containerAppsStack.GetEnvironment().Name,
		ComponentName:       pulumi.String("secretstore"),
		ComponentType:       pulumi.String("secretstores.azure.keyvault"),
		Version:            pulumi.String("v1"),
		Metadata: app.DaprMetadataArray{
			&app.DaprMetadataArgs{
				Name:  pulumi.String("vaultName"),
				Value: pulumi.String(config.VaultName),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("azureTenantId"),
				Value: pulumi.String(""), // From environment variables
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("azureClientId"),
				Value: pulumi.String(""), // From environment variables
			},
			&app.DaprMetadataArgs{
				Name:      pulumi.String("azureClientSecret"),
				SecretRef: pulumi.String("azure-client-secret"),
			},
		},
		Secrets: app.SecretArray{
			&app.SecretArgs{
				Name:  pulumi.String("azure-client-secret"),
				Value: pulumi.String(""), // Retrieved from deployment environment
			},
		},
		Scopes: pulumi.StringArray{
			pulumi.String("identity-api"),
			pulumi.String("content-api"),
			pulumi.String("services-api"),
			pulumi.String("public-gateway"),
			pulumi.String("admin-gateway"),
		},
	})
	if err != nil {
		return err
	}

	deployment.deployedComponents["secretstore"] = component
	return nil
}

func (deployment *DaprComponentDeployment) deployStorageBindingComponent(ctx *pulumi.Context, config *ComponentConfiguration) error {
	component, err := app.NewDaprComponent(ctx, "staging-binding-storage", &app.DaprComponentArgs{
		ResourceGroupName:    deployment.containerAppsStack.GetResourceGroup().Name,
		EnvironmentName:     deployment.containerAppsStack.GetEnvironment().Name,
		ComponentName:       pulumi.String("storage-binding"),
		ComponentType:       pulumi.String("bindings.azure.blobstorage"),
		Version:            pulumi.String("v1"),
		Metadata: app.DaprMetadataArray{
			&app.DaprMetadataArgs{
				Name:  pulumi.String("storageAccount"),
				Value: pulumi.String(config.StorageAccountName),
			},
			&app.DaprMetadataArgs{
				Name:      pulumi.String("storageAccessKey"),
				SecretRef: pulumi.String("storage-access-key"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("container"),
				Value: pulumi.String("content"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("decodeBase64"),
				Value: pulumi.String("false"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("getBlobRetryCount"),
				Value: pulumi.String("3"),
			},
		},
		Secrets: app.SecretArray{
			&app.SecretArgs{
				Name:        pulumi.String("storage-access-key"),
				KeyVaultUrl: deployment.vaultStack.GetVaultUri(),
				Identity:    pulumi.String("system"),
			},
		},
		Scopes: pulumi.StringArray{
			pulumi.String("content-api"),
		},
	})
	if err != nil {
		return err
	}

	deployment.deployedComponents["storage-binding"] = component
	return nil
}

func (deployment *DaprComponentDeployment) deployServiceBusBindingComponent(ctx *pulumi.Context, config *ComponentConfiguration) error {
	component, err := app.NewDaprComponent(ctx, "staging-binding-servicebus", &app.DaprComponentArgs{
		ResourceGroupName:    deployment.containerAppsStack.GetResourceGroup().Name,
		EnvironmentName:     deployment.containerAppsStack.GetEnvironment().Name,
		ComponentName:       pulumi.String("servicebus-binding"),
		ComponentType:       pulumi.String("bindings.azure.servicebusqueues"),
		Version:            pulumi.String("v1"),
		Metadata: app.DaprMetadataArray{
			&app.DaprMetadataArgs{
				Name:      pulumi.String("connectionString"),
				SecretRef: pulumi.String("servicebus-connection-string"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("queueName"),
				Value: pulumi.String("content-processing"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("ttlInSeconds"),
				Value: pulumi.String("60"),
			},
		},
		Secrets: app.SecretArray{
			&app.SecretArgs{
				Name:        pulumi.String("servicebus-connection-string"),
				KeyVaultUrl: deployment.vaultStack.GetVaultUri(),
				Identity:    pulumi.String("system"),
			},
		},
		Scopes: pulumi.StringArray{
			pulumi.String("content-api"),
			pulumi.String("services-api"),
		},
	})
	if err != nil {
		return err
	}

	deployment.deployedComponents["servicebus-binding"] = component
	return nil
}

func (deployment *DaprComponentDeployment) deployCosmosBindingComponent(ctx *pulumi.Context, config *ComponentConfiguration) error {
	component, err := app.NewDaprComponent(ctx, "staging-binding-cosmos", &app.DaprComponentArgs{
		ResourceGroupName:    deployment.containerAppsStack.GetResourceGroup().Name,
		EnvironmentName:     deployment.containerAppsStack.GetEnvironment().Name,
		ComponentName:       pulumi.String("cosmos-binding"),
		ComponentType:       pulumi.String("bindings.azure.cosmosdb"),
		Version:            pulumi.String("v1"),
		Metadata: app.DaprMetadataArray{
			&app.DaprMetadataArgs{
				Name:  pulumi.String("url"),
				Value: pulumi.String("https://international-center-staging-cosmos.documents.azure.com:443/"),
			},
			&app.DaprMetadataArgs{
				Name:      pulumi.String("masterKey"),
				SecretRef: pulumi.String("cosmos-master-key"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("database"),
				Value: pulumi.String("analytics"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("collection"),
				Value: pulumi.String("events"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("partitionKey"),
				Value: pulumi.String("partitionKey"),
			},
		},
		Secrets: app.SecretArray{
			&app.SecretArgs{
				Name:        pulumi.String("cosmos-master-key"),
				KeyVaultUrl: deployment.vaultStack.GetVaultUri(),
				Identity:    pulumi.String("system"),
			},
		},
		Scopes: pulumi.StringArray{
			pulumi.String("services-api"),
		},
	})
	if err != nil {
		return err
	}

	deployment.deployedComponents["cosmos-binding"] = component
	return nil
}

func (deployment *DaprComponentDeployment) deployRateLimitMiddleware(ctx *pulumi.Context, config *ComponentConfiguration) error {
	component, err := app.NewDaprComponent(ctx, "staging-middleware-ratelimit", &app.DaprComponentArgs{
		ResourceGroupName:    deployment.containerAppsStack.GetResourceGroup().Name,
		EnvironmentName:     deployment.containerAppsStack.GetEnvironment().Name,
		ComponentName:       pulumi.String("ratelimit"),
		ComponentType:       pulumi.String("middleware.http.ratelimit"),
		Version:            pulumi.String("v1"),
		Metadata: app.DaprMetadataArray{
			&app.DaprMetadataArgs{
				Name:  pulumi.String("maxRequestsPerSecond"),
				Value: pulumi.String("100"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("statusCode"),
				Value: pulumi.String("429"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("statusMessage"),
				Value: pulumi.String("Too Many Requests"),
			},
		},
		Scopes: pulumi.StringArray{
			pulumi.String("public-gateway"),
			pulumi.String("admin-gateway"),
		},
	})
	if err != nil {
		return err
	}

	deployment.deployedComponents["ratelimit"] = component
	return nil
}

func (deployment *DaprComponentDeployment) deployOAuthMiddleware(ctx *pulumi.Context, config *ComponentConfiguration) error {
	component, err := app.NewDaprComponent(ctx, "staging-middleware-oauth", &app.DaprComponentArgs{
		ResourceGroupName:    deployment.containerAppsStack.GetResourceGroup().Name,
		EnvironmentName:     deployment.containerAppsStack.GetEnvironment().Name,
		ComponentName:       pulumi.String("oauth2"),
		ComponentType:       pulumi.String("middleware.http.oauth2"),
		Version:            pulumi.String("v1"),
		Metadata: app.DaprMetadataArray{
			&app.DaprMetadataArgs{
				Name:  pulumi.String("clientId"),
				Value: pulumi.String(""), // From Azure AD App Registration
			},
			&app.DaprMetadataArgs{
				Name:      pulumi.String("clientSecret"),
				SecretRef: pulumi.String("oauth-client-secret"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("scopes"),
				Value: pulumi.String("openid profile email"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("authURL"),
				Value: pulumi.String("https://login.microsoftonline.com/common/oauth2/v2.0/authorize"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("tokenURL"),
				Value: pulumi.String("https://login.microsoftonline.com/common/oauth2/v2.0/token"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("redirectURL"),
				Value: pulumi.String("https://admin-api-staging.international-center.com/auth/callback"),
			},
		},
		Secrets: app.SecretArray{
			&app.SecretArgs{
				Name:        pulumi.String("oauth-client-secret"),
				KeyVaultUrl: deployment.vaultStack.GetVaultUri(),
				Identity:    pulumi.String("system"),
			},
		},
		Scopes: pulumi.StringArray{
			pulumi.String("admin-gateway"),
		},
	})
	if err != nil {
		return err
	}

	deployment.deployedComponents["oauth2"] = component
	return nil
}

func (deployment *DaprComponentDeployment) deployCORSMiddleware(ctx *pulumi.Context, config *ComponentConfiguration) error {
	component, err := app.NewDaprComponent(ctx, "staging-middleware-cors", &app.DaprComponentArgs{
		ResourceGroupName:    deployment.containerAppsStack.GetResourceGroup().Name,
		EnvironmentName:     deployment.containerAppsStack.GetEnvironment().Name,
		ComponentName:       pulumi.String("cors"),
		ComponentType:       pulumi.String("middleware.http.cors"),
		Version:            pulumi.String("v1"),
		Metadata: app.DaprMetadataArray{
			&app.DaprMetadataArgs{
				Name:  pulumi.String("allowOrigins"),
				Value: pulumi.String("https://staging.international-center.com,https://app-staging.international-center.com,https://admin-staging.international-center.com"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("allowMethods"),
				Value: pulumi.String("GET,POST,PUT,DELETE,PATCH,OPTIONS"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("allowHeaders"),
				Value: pulumi.String("Content-Type,Authorization,X-Requested-With"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("maxAge"),
				Value: pulumi.String("3600"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("allowCredentials"),
				Value: pulumi.String("true"),
			},
		},
		Scopes: pulumi.StringArray{
			pulumi.String("public-gateway"),
			pulumi.String("admin-gateway"),
		},
	})
	if err != nil {
		return err
	}

	deployment.deployedComponents["cors"] = component
	return nil
}

func (deployment *DaprComponentDeployment) GetDeployedComponent(name string) *app.DaprComponent {
	return deployment.deployedComponents[name]
}

func (deployment *DaprComponentDeployment) ListDeployedComponents() map[string]*app.DaprComponent {
	return deployment.deployedComponents
}