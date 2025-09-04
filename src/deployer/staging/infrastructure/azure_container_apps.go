package infrastructure

import (
	"context"
	"fmt"

	"github.com/pulumi/pulumi-azure-native/sdk/v2/go/azurenative/app"
	"github.com/pulumi/pulumi-azure-native/sdk/v2/go/azurenative/resources"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type AzureContainerAppsStack struct {
	resourceGroup *resources.ResourceGroup
	environment   *app.ManagedEnvironment
	apps          map[string]*app.ContainerApp
	daprComponents map[string]*app.DaprComponent
}

func NewAzureContainerAppsStack() *AzureContainerAppsStack {
	return &AzureContainerAppsStack{
		apps:          make(map[string]*app.ContainerApp),
		daprComponents: make(map[string]*app.DaprComponent),
	}
}

func (stack *AzureContainerAppsStack) Deploy(ctx *pulumi.Context) error {
	if err := stack.createResourceGroup(ctx); err != nil {
		return fmt.Errorf("failed to create resource group: %w", err)
	}

	if err := stack.createContainerAppsEnvironment(ctx); err != nil {
		return fmt.Errorf("failed to create container apps environment: %w", err)
	}

	if err := stack.createDaprComponents(ctx); err != nil {
		return fmt.Errorf("failed to create Dapr components: %w", err)
	}

	if err := stack.createApiContainerApps(ctx); err != nil {
		return fmt.Errorf("failed to create API container apps: %w", err)
	}

	if err := stack.createGatewayContainerApps(ctx); err != nil {
		return fmt.Errorf("failed to create gateway container apps: %w", err)
	}

	return nil
}

func (stack *AzureContainerAppsStack) createResourceGroup(ctx *pulumi.Context) error {
	rg, err := resources.NewResourceGroup(ctx, "staging-rg", &resources.ResourceGroupArgs{
		ResourceGroupName: pulumi.String("international-center-staging"),
		Location:         pulumi.String("East US 2"),
		Tags: pulumi.StringMap{
			"environment": pulumi.String("staging"),
			"project":     pulumi.String("international-center"),
			"managed-by":  pulumi.String("pulumi"),
		},
	})
	if err != nil {
		return err
	}

	stack.resourceGroup = rg
	return nil
}

func (stack *AzureContainerAppsStack) createContainerAppsEnvironment(ctx *pulumi.Context) error {
	env, err := app.NewManagedEnvironment(ctx, "staging-env", &app.ManagedEnvironmentArgs{
		ResourceGroupName:    stack.resourceGroup.Name,
		EnvironmentName:      pulumi.String("international-center-staging"),
		Location:            stack.resourceGroup.Location,
		DaprAIInstrumentationKey: pulumi.String(""),
		DaprAIConnectionString:   pulumi.String(""),
		AppLogsConfiguration: &app.AppLogsConfigurationArgs{
			Destination: pulumi.String("log-analytics"),
		},
		Tags: pulumi.StringMap{
			"environment": pulumi.String("staging"),
			"project":     pulumi.String("international-center"),
		},
	})
	if err != nil {
		return err
	}

	stack.environment = env
	return nil
}

func (stack *AzureContainerAppsStack) createDaprComponents(ctx *pulumi.Context) error {
	if err := stack.createRedisPubSubComponent(ctx); err != nil {
		return err
	}

	if err := stack.createSecretStoreComponent(ctx); err != nil {
		return err
	}

	if err := stack.createStateStoreComponent(ctx); err != nil {
		return err
	}

	if err := stack.createBindingsComponent(ctx); err != nil {
		return err
	}

	return nil
}

func (stack *AzureContainerAppsStack) createRedisPubSubComponent(ctx *pulumi.Context) error {
	component, err := app.NewDaprComponent(ctx, "staging-pubsub", &app.DaprComponentArgs{
		ResourceGroupName:    stack.resourceGroup.Name,
		EnvironmentName:     stack.environment.Name,
		ComponentName:       pulumi.String("pubsub"),
		ComponentType:       pulumi.String("pubsub.redis"),
		Version:            pulumi.String("v1"),
		Metadata: app.DaprMetadataArray{
			&app.DaprMetadataArgs{
				Name:  pulumi.String("redisHost"),
				Value: pulumi.String("redis-staging.redis.cache.windows.net:6380"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("redisPassword"),
				SecretRef: pulumi.String("redis-password"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("enableTLS"),
				Value: pulumi.String("true"),
			},
		},
		Secrets: app.SecretArray{
			&app.SecretArgs{
				Name:  pulumi.String("redis-password"),
				Value: pulumi.String(""), // Retrieved from Key Vault
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

	stack.daprComponents["pubsub"] = component
	return nil
}

func (stack *AzureContainerAppsStack) createSecretStoreComponent(ctx *pulumi.Context) error {
	component, err := app.NewDaprComponent(ctx, "staging-secretstore", &app.DaprComponentArgs{
		ResourceGroupName:    stack.resourceGroup.Name,
		EnvironmentName:     stack.environment.Name,
		ComponentName:       pulumi.String("secretstore"),
		ComponentType:       pulumi.String("secretstores.azure.keyvault"),
		Version:            pulumi.String("v1"),
		Metadata: app.DaprMetadataArray{
			&app.DaprMetadataArgs{
				Name:  pulumi.String("vaultName"),
				Value: pulumi.String("international-center-staging-kv"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("azureTenantId"),
				Value: pulumi.String(""), // From environment
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("azureClientId"),
				Value: pulumi.String(""), // From environment
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("azureClientSecret"),
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

	stack.daprComponents["secretstore"] = component
	return nil
}

func (stack *AzureContainerAppsStack) createStateStoreComponent(ctx *pulumi.Context) error {
	component, err := app.NewDaprComponent(ctx, "staging-statestore", &app.DaprComponentArgs{
		ResourceGroupName:    stack.resourceGroup.Name,
		EnvironmentName:     stack.environment.Name,
		ComponentName:       pulumi.String("statestore"),
		ComponentType:       pulumi.String("state.redis"),
		Version:            pulumi.String("v1"),
		Metadata: app.DaprMetadataArray{
			&app.DaprMetadataArgs{
				Name:  pulumi.String("redisHost"),
				Value: pulumi.String("redis-staging.redis.cache.windows.net:6380"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("redisPassword"),
				SecretRef: pulumi.String("redis-password"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("enableTLS"),
				Value: pulumi.String("true"),
			},
		},
		Secrets: app.SecretArray{
			&app.SecretArgs{
				Name:  pulumi.String("redis-password"),
				Value: pulumi.String(""), // Retrieved from Key Vault
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

	stack.daprComponents["statestore"] = component
	return nil
}

func (stack *AzureContainerAppsStack) createBindingsComponent(ctx *pulumi.Context) error {
	component, err := app.NewDaprComponent(ctx, "staging-storage-binding", &app.DaprComponentArgs{
		ResourceGroupName:    stack.resourceGroup.Name,
		EnvironmentName:     stack.environment.Name,
		ComponentName:       pulumi.String("storage-binding"),
		ComponentType:       pulumi.String("bindings.azure.blobstorage"),
		Version:            pulumi.String("v1"),
		Metadata: app.DaprMetadataArray{
			&app.DaprMetadataArgs{
				Name:  pulumi.String("storageAccount"),
				Value: pulumi.String("internationalcenterstaging"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("storageAccessKey"),
				SecretRef: pulumi.String("storage-access-key"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("container"),
				Value: pulumi.String("content"),
			},
		},
		Secrets: app.SecretArray{
			&app.SecretArgs{
				Name:  pulumi.String("storage-access-key"),
				Value: pulumi.String(""), // Retrieved from Key Vault
			},
		},
		Scopes: pulumi.StringArray{
			pulumi.String("content-api"),
		},
	})
	if err != nil {
		return err
	}

	stack.daprComponents["storage-binding"] = component
	return nil
}

func (stack *AzureContainerAppsStack) createApiContainerApps(ctx *pulumi.Context) error {
	apis := []string{"identity-api", "content-api", "services-api"}
	
	for _, apiName := range apis {
		if err := stack.createApiContainerApp(ctx, apiName); err != nil {
			return fmt.Errorf("failed to create %s: %w", apiName, err)
		}
	}

	return nil
}

func (stack *AzureContainerAppsStack) createApiContainerApp(ctx *pulumi.Context, apiName string) error {
	containerApp, err := app.NewContainerApp(ctx, fmt.Sprintf("staging-%s", apiName), &app.ContainerAppArgs{
		ResourceGroupName:    stack.resourceGroup.Name,
		ContainerAppName:     pulumi.String(apiName),
		ManagedEnvironmentId: stack.environment.ID(),
		Configuration: &app.ConfigurationArgs{
			Ingress: &app.IngressArgs{
				External:   pulumi.Bool(false), // Internal only, accessed via gateways
				TargetPort: pulumi.Int(8080),
				Traffic: app.TrafficWeightArray{
					&app.TrafficWeightArgs{
						RevisionName: pulumi.String(""), // Latest revision
						Weight:       pulumi.Int(100),
					},
				},
			},
			Dapr: &app.DaprArgs{
				Enabled: pulumi.Bool(true),
				AppId:   pulumi.String(apiName),
				AppPort: pulumi.Int(8080),
				AppProtocol: pulumi.String("http"),
			},
			Secrets: app.SecretArray{
				&app.SecretArgs{
					Name:  pulumi.String("database-connection"),
					Value: pulumi.String(""), // Retrieved from Key Vault
				},
			},
		},
		Template: &app.TemplateArgs{
			Scale: &app.ScaleArgs{
				MinReplicas: pulumi.Int(2),
				MaxReplicas: pulumi.Int(10),
				Rules: app.ScaleRuleArray{
					&app.ScaleRuleArgs{
						Name: pulumi.String("http-requests"),
						Http: &app.HttpScaleRuleArgs{
							Metadata: pulumi.StringMap{
								"concurrentRequests": pulumi.String("30"),
							},
						},
					},
				},
			},
			Containers: app.ContainerArray{
				&app.ContainerArgs{
					Name:  pulumi.String(apiName),
					Image: pulumi.String(fmt.Sprintf("international-center/%s:staging", apiName)),
					Resources: &app.ContainerResourcesArgs{
						Cpu:    pulumi.Float64(0.5),
						Memory: pulumi.String("1Gi"),
					},
					Env: app.EnvironmentVarArray{
						&app.EnvironmentVarArgs{
							Name:      pulumi.String("ENVIRONMENT"),
							Value:     pulumi.String("staging"),
						},
						&app.EnvironmentVarArgs{
							Name:      pulumi.String("DATABASE_CONNECTION"),
							SecretRef: pulumi.String("database-connection"),
						},
						&app.EnvironmentVarArgs{
							Name:  pulumi.String("DAPR_HTTP_PORT"),
							Value: pulumi.String("3500"),
						},
						&app.EnvironmentVarArgs{
							Name:  pulumi.String("DAPR_GRPC_PORT"),
							Value: pulumi.String("50001"),
						},
					},
					Probes: app.ContainerAppProbeArray{
						&app.ContainerAppProbeArgs{
							Type: pulumi.String("Liveness"),
							HttpGet: &app.ContainerAppProbeHttpGetArgs{
								Path: pulumi.String("/health"),
								Port: pulumi.Int(8080),
							},
							InitialDelaySeconds: pulumi.Int(30),
							PeriodSeconds:      pulumi.Int(10),
						},
						&app.ContainerAppProbeArgs{
							Type: pulumi.String("Readiness"),
							HttpGet: &app.ContainerAppProbeHttpGetArgs{
								Path: pulumi.String("/ready"),
								Port: pulumi.Int(8080),
							},
							InitialDelaySeconds: pulumi.Int(5),
							PeriodSeconds:      pulumi.Int(5),
						},
					},
				},
			},
		},
		Tags: pulumi.StringMap{
			"environment": pulumi.String("staging"),
			"service":     pulumi.String(apiName),
			"project":     pulumi.String("international-center"),
		},
	})
	if err != nil {
		return err
	}

	stack.apps[apiName] = containerApp
	return nil
}

func (stack *AzureContainerAppsStack) createGatewayContainerApps(ctx *pulumi.Context) error {
	gateways := []string{"public-gateway", "admin-gateway"}
	
	for _, gatewayName := range gateways {
		if err := stack.createGatewayContainerApp(ctx, gatewayName); err != nil {
			return fmt.Errorf("failed to create %s: %w", gatewayName, err)
		}
	}

	return nil
}

func (stack *AzureContainerAppsStack) createGatewayContainerApp(ctx *pulumi.Context, gatewayName string) error {
	isPublic := gatewayName == "public-gateway"
	
	containerApp, err := app.NewContainerApp(ctx, fmt.Sprintf("staging-%s", gatewayName), &app.ContainerAppArgs{
		ResourceGroupName:    stack.resourceGroup.Name,
		ContainerAppName:     pulumi.String(gatewayName),
		ManagedEnvironmentId: stack.environment.ID(),
		Configuration: &app.ConfigurationArgs{
			Ingress: &app.IngressArgs{
				External:   pulumi.Bool(isPublic), // Public gateway is external
				TargetPort: pulumi.Int(8080),
				Traffic: app.TrafficWeightArray{
					&app.TrafficWeightArgs{
						RevisionName: pulumi.String(""), // Latest revision
						Weight:       pulumi.Int(100),
					},
				},
				CustomDomains: func() app.CustomDomainArray {
					if isPublic {
						return app.CustomDomainArray{
							&app.CustomDomainArgs{
								Name: pulumi.String("api-staging.international-center.com"),
								CertificateId: pulumi.String(""), // SSL certificate
							},
						}
					}
					return nil
				}(),
			},
			Dapr: &app.DaprArgs{
				Enabled: pulumi.Bool(true),
				AppId:   pulumi.String(gatewayName),
				AppPort: pulumi.Int(8080),
				AppProtocol: pulumi.String("http"),
			},
		},
		Template: &app.TemplateArgs{
			Scale: &app.ScaleArgs{
				MinReplicas: pulumi.Int(3), // Higher availability for gateways
				MaxReplicas: pulumi.Int(20),
				Rules: app.ScaleRuleArray{
					&app.ScaleRuleArgs{
						Name: pulumi.String("http-requests"),
						Http: &app.HttpScaleRuleArgs{
							Metadata: pulumi.StringMap{
								"concurrentRequests": pulumi.String("100"),
							},
						},
					},
				},
			},
			Containers: app.ContainerArray{
				&app.ContainerArgs{
					Name:  pulumi.String(gatewayName),
					Image: pulumi.String(fmt.Sprintf("international-center/%s:staging", gatewayName)),
					Resources: &app.ContainerResourcesArgs{
						Cpu:    pulumi.Float64(0.75),
						Memory: pulumi.String("1.5Gi"),
					},
					Env: app.EnvironmentVarArray{
						&app.EnvironmentVarArgs{
							Name:  pulumi.String("ENVIRONMENT"),
							Value: pulumi.String("staging"),
						},
						&app.EnvironmentVarArgs{
							Name:  pulumi.String("GATEWAY_TYPE"),
							Value: pulumi.String(gatewayName),
						},
						&app.EnvironmentVarArgs{
							Name:  pulumi.String("DAPR_HTTP_PORT"),
							Value: pulumi.String("3500"),
						},
						&app.EnvironmentVarArgs{
							Name:  pulumi.String("DAPR_GRPC_PORT"),
							Value: pulumi.String("50001"),
						},
						&app.EnvironmentVarArgs{
							Name:  pulumi.String("CORS_ALLOWED_ORIGINS"),
							Value: func() pulumi.String {
								if isPublic {
									return pulumi.String("https://staging.international-center.com,https://app-staging.international-center.com")
								}
								return pulumi.String("https://admin-staging.international-center.com")
							}(),
						},
					},
					Probes: app.ContainerAppProbeArray{
						&app.ContainerAppProbeArgs{
							Type: pulumi.String("Liveness"),
							HttpGet: &app.ContainerAppProbeHttpGetArgs{
								Path: pulumi.String("/health"),
								Port: pulumi.Int(8080),
							},
							InitialDelaySeconds: pulumi.Int(30),
							PeriodSeconds:      pulumi.Int(10),
						},
						&app.ContainerAppProbeArgs{
							Type: pulumi.String("Readiness"),
							HttpGet: &app.ContainerAppProbeHttpGetArgs{
								Path: pulumi.String("/ready"),
								Port: pulumi.Int(8080),
							},
							InitialDelaySeconds: pulumi.Int(5),
							PeriodSeconds:      pulumi.Int(5),
						},
					},
				},
			},
		},
		Tags: pulumi.StringMap{
			"environment": pulumi.String("staging"),
			"service":     pulumi.String(gatewayName),
			"project":     pulumi.String("international-center"),
		},
	})
	if err != nil {
		return err
	}

	stack.apps[gatewayName] = containerApp
	return nil
}

func (stack *AzureContainerAppsStack) GetResourceGroup() *resources.ResourceGroup {
	return stack.resourceGroup
}

func (stack *AzureContainerAppsStack) GetEnvironment() *app.ManagedEnvironment {
	return stack.environment
}

func (stack *AzureContainerAppsStack) GetContainerApp(name string) *app.ContainerApp {
	return stack.apps[name]
}

func (stack *AzureContainerAppsStack) GetDaprComponent(name string) *app.DaprComponent {
	return stack.daprComponents[name]
}