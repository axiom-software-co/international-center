package infrastructure

import (
	"fmt"

	"github.com/pulumi/pulumi-azure-native-sdk/app/v3"
	"github.com/pulumi/pulumi-azure-native-sdk/network/v3"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type AzureProductionAppsStack struct {
	resourceGroup        *resources.ResourceGroup
	vnet                *network.VirtualNetwork
	containerAppsSubnet *network.Subnet
	privateSubnet       *network.Subnet
	environment         *app.ManagedEnvironment
	apps                map[string]*app.ContainerApp
	daprComponents      map[string]*app.DaprComponent
	// TODO: Fix security.Setting API compatibility in Azure Native SDK v1.104.0
	// securityCenter      *security.Setting // Removed due to API changes
	networkSecurityGroup *network.NetworkSecurityGroup
}

func NewAzureProductionAppsStack() *AzureProductionAppsStack {
	return &AzureProductionAppsStack{
		apps:          make(map[string]*app.ContainerApp),
		daprComponents: make(map[string]*app.DaprComponent),
	}
}

func (stack *AzureProductionAppsStack) Deploy(ctx *pulumi.Context) error {
	if err := stack.createResourceGroup(ctx); err != nil {
		return fmt.Errorf("failed to create production resource group: %w", err)
	}

	if err := stack.createNetworkInfrastructure(ctx); err != nil {
		return fmt.Errorf("failed to create network infrastructure: %w", err)
	}

	if err := stack.enableSecurityCenter(ctx); err != nil {
		return fmt.Errorf("failed to enable security center: %w", err)
	}

	if err := stack.createContainerAppsEnvironment(ctx); err != nil {
		return fmt.Errorf("failed to create container apps environment: %w", err)
	}

	if err := stack.createProductionDaprComponents(ctx); err != nil {
		return fmt.Errorf("failed to create production Dapr components: %w", err)
	}

	if err := stack.createProductionApiContainerApps(ctx); err != nil {
		return fmt.Errorf("failed to create production API container apps: %w", err)
	}

	if err := stack.createProductionGatewayContainerApps(ctx); err != nil {
		return fmt.Errorf("failed to create production gateway container apps: %w", err)
	}

	return nil
}

func (stack *AzureProductionAppsStack) createResourceGroup(ctx *pulumi.Context) error {
	rg, err := resources.NewResourceGroup(ctx, "production-rg", &resources.ResourceGroupArgs{
		ResourceGroupName: pulumi.String("international-center-production"),
		Location:         pulumi.String("East US 2"),
		Tags: pulumi.StringMap{
			"environment":      pulumi.String("production"),
			"project":         pulumi.String("international-center"),
			"managed-by":      pulumi.String("pulumi"),
			"compliance":      pulumi.String("required"),
			"backup-required": pulumi.String("true"),
			"tier":           pulumi.String("production"),
		},
	})
	if err != nil {
		return err
	}

	stack.resourceGroup = rg
	return nil
}

func (stack *AzureProductionAppsStack) createNetworkInfrastructure(ctx *pulumi.Context) error {
	nsg, err := network.NewNetworkSecurityGroup(ctx, "production-nsg", &network.NetworkSecurityGroupArgs{
		ResourceGroupName:        stack.resourceGroup.Name,
		NetworkSecurityGroupName: pulumi.String("international-center-production-nsg"),
		Location:                stack.resourceGroup.Location,
		SecurityRules: network.SecurityRuleTypeArray{
			&network.SecurityRuleTypeArgs{
				Name:                     pulumi.String("AllowHTTPS"),
				Protocol:                 pulumi.String(string(network.SecurityRuleProtocolTcp)),
				SourcePortRange:          pulumi.String("*"),
				DestinationPortRange:     pulumi.String("443"),
				SourceAddressPrefix:      pulumi.String("*"),
				DestinationAddressPrefix: pulumi.String("*"),
				Access:                   pulumi.String(string(network.SecurityRuleAccessAllow)),
				Priority:                 pulumi.Int(100),
				Direction:                pulumi.String(string(network.SecurityRuleDirectionInbound)),
			},
			&network.SecurityRuleTypeArgs{
				Name:                     pulumi.String("AllowHTTP"),
				Protocol:                 pulumi.String(string(network.SecurityRuleProtocolTcp)),
				SourcePortRange:          pulumi.String("*"),
				DestinationPortRange:     pulumi.String("80"),
				SourceAddressPrefix:      pulumi.String("*"),
				DestinationAddressPrefix: pulumi.String("*"),
				Access:                   pulumi.String(string(network.SecurityRuleAccessAllow)),
				Priority:                 pulumi.Int(110),
				Direction:                pulumi.String(string(network.SecurityRuleDirectionInbound)),
			},
			&network.SecurityRuleTypeArgs{
				Name:                     pulumi.String("DenyAllInbound"),
				Protocol:                 pulumi.String(string(network.SecurityRuleProtocolAsterisk)),
				SourcePortRange:          pulumi.String("*"),
				DestinationPortRange:     pulumi.String("*"),
				SourceAddressPrefix:      pulumi.String("*"),
				DestinationAddressPrefix: pulumi.String("*"),
				Access:                   pulumi.String(string(network.SecurityRuleAccessDeny)),
				Priority:                 pulumi.Int(4096),
				Direction:                pulumi.String(string(network.SecurityRuleDirectionInbound)),
			},
		},
		Tags: pulumi.StringMap{
			"environment": pulumi.String("production"),
			"project":     pulumi.String("international-center"),
		},
	})
	if err != nil {
		return err
	}

	vnet, err := network.NewVirtualNetwork(ctx, "production-vnet", &network.VirtualNetworkArgs{
		ResourceGroupName:    stack.resourceGroup.Name,
		VirtualNetworkName:   pulumi.String("international-center-production-vnet"),
		Location:            stack.resourceGroup.Location,
		AddressSpace: &network.AddressSpaceArgs{
			AddressPrefixes: pulumi.StringArray{
				pulumi.String("10.0.0.0/16"),
			},
		},
		Tags: pulumi.StringMap{
			"environment": pulumi.String("production"),
			"project":     pulumi.String("international-center"),
		},
	})
	if err != nil {
		return err
	}

	containerAppsSubnet, err := network.NewSubnet(ctx, "production-container-apps-subnet", &network.SubnetArgs{
		ResourceGroupName:  stack.resourceGroup.Name,
		VirtualNetworkName: vnet.Name,
		SubnetName:         pulumi.String("container-apps-subnet"),
		AddressPrefix:      pulumi.String("10.0.1.0/23"),
		NetworkSecurityGroup: &network.NetworkSecurityGroupTypeArgs{
			Id: nsg.ID(),
		},
		ServiceEndpoints: network.ServiceEndpointPropertiesFormatArray{
			&network.ServiceEndpointPropertiesFormatArgs{
				Service: pulumi.String("Microsoft.KeyVault"),
			},
			&network.ServiceEndpointPropertiesFormatArgs{
				Service: pulumi.String("Microsoft.Storage"),
			},
			&network.ServiceEndpointPropertiesFormatArgs{
				Service: pulumi.String("Microsoft.Sql"),
			},
		},
		PrivateEndpointNetworkPolicies:    pulumi.String("Disabled"),
		PrivateLinkServiceNetworkPolicies: pulumi.String("Enabled"),
	})
	if err != nil {
		return err
	}

	privateSubnet, err := network.NewSubnet(ctx, "production-private-subnet", &network.SubnetArgs{
		ResourceGroupName:  stack.resourceGroup.Name,
		VirtualNetworkName: vnet.Name,
		SubnetName:         pulumi.String("private-subnet"),
		AddressPrefix:      pulumi.String("10.0.3.0/24"),
		NetworkSecurityGroup: &network.NetworkSecurityGroupTypeArgs{
			Id: nsg.ID(),
		},
		PrivateEndpointNetworkPolicies:    pulumi.String("Disabled"),
		PrivateLinkServiceNetworkPolicies: pulumi.String("Enabled"),
	})
	if err != nil {
		return err
	}

	stack.networkSecurityGroup = nsg
	stack.vnet = vnet
	stack.containerAppsSubnet = containerAppsSubnet
	stack.privateSubnet = privateSubnet
	return nil
}

func (stack *AzureProductionAppsStack) enableSecurityCenter(ctx *pulumi.Context) error {
	// TODO: Fix security.NewSetting API compatibility in Azure Native SDK v1.104.0
	// Security Center settings API has changed significantly
	// securitySetting, err := security.NewSetting(ctx, "production-security-center", &security.SettingArgs{
	// 	SettingName: pulumi.String("MCAS"),
	// 	Enabled:     pulumi.Bool(true),
	// 	Kind:        pulumi.String("AlertSyncSettings"),
	// })
	// if err != nil {
	// 	return err
	// }
	// stack.securityCenter = securitySetting
	return nil
}

func (stack *AzureProductionAppsStack) createContainerAppsEnvironment(ctx *pulumi.Context) error {
	env, err := app.NewManagedEnvironment(ctx, "production-env", &app.ManagedEnvironmentArgs{
		ResourceGroupName:    stack.resourceGroup.Name,
		EnvironmentName:      pulumi.String("international-center-production"),
		Location:            stack.resourceGroup.Location,
		VnetConfiguration: &app.VnetConfigurationArgs{
			InfrastructureSubnetId: stack.containerAppsSubnet.ID(),
		},
		AppLogsConfiguration: &app.AppLogsConfigurationArgs{
			Destination: pulumi.String("log-analytics"),
			LogAnalyticsConfiguration: &app.LogAnalyticsConfigurationArgs{
				CustomerId: pulumi.String(""), // From environment variable
				SharedKey:  pulumi.String(""), // From Key Vault
			},
		},
		Tags: pulumi.StringMap{
			"environment": pulumi.String("production"),
			"project":     pulumi.String("international-center"),
			"tier":        pulumi.String("production"),
		},
	})
	if err != nil {
		return err
	}

	stack.environment = env
	return nil
}

func (stack *AzureProductionAppsStack) createProductionDaprComponents(ctx *pulumi.Context) error {
	if err := stack.createProductionRedisPubSubComponent(ctx); err != nil {
		return err
	}

	if err := stack.createProductionSecretStoreComponent(ctx); err != nil {
		return err
	}

	if err := stack.createProductionStateStoreComponent(ctx); err != nil {
		return err
	}

	if err := stack.createProductionBindingsComponent(ctx); err != nil {
		return err
	}

	return nil
}

func (stack *AzureProductionAppsStack) createProductionRedisPubSubComponent(ctx *pulumi.Context) error {
	component, err := app.NewDaprComponent(ctx, "production-pubsub", &app.DaprComponentArgs{
		ResourceGroupName:    stack.resourceGroup.Name,
		EnvironmentName:     stack.environment.Name,
		ComponentName:       pulumi.String("pubsub"),
		ComponentType:       pulumi.String("pubsub.redis"),
		Version:            pulumi.String("v1"),
		Metadata: app.DaprMetadataArray{
			&app.DaprMetadataArgs{
				Name:  pulumi.String("redisHost"),
				Value: pulumi.String("redis-production.redis.cache.windows.net:6380"),
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
				Value: pulumi.String("5"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("maxRetryBackoff"),
				Value: pulumi.String("5s"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("dialTimeout"),
				Value: pulumi.String("10s"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("readTimeout"),
				Value: pulumi.String("30s"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("writeTimeout"),
				Value: pulumi.String("30s"),
			},
		},
		Secrets: app.SecretArray{
			&app.SecretArgs{
				Name:        pulumi.String("redis-password"),
				Value: pulumi.String(""), // TODO: Fix KeyVault integration in Azure Native SDK v1.104.0
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

func (stack *AzureProductionAppsStack) createProductionSecretStoreComponent(ctx *pulumi.Context) error {
	component, err := app.NewDaprComponent(ctx, "production-secretstore", &app.DaprComponentArgs{
		ResourceGroupName:    stack.resourceGroup.Name,
		EnvironmentName:     stack.environment.Name,
		ComponentName:       pulumi.String("secretstore"),
		ComponentType:       pulumi.String("secretstores.azure.keyvault"),
		Version:            pulumi.String("v1"),
		Metadata: app.DaprMetadataArray{
			&app.DaprMetadataArgs{
				Name:  pulumi.String("vaultName"),
				Value: pulumi.String("international-center-production-kv"),
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

	stack.daprComponents["secretstore"] = component
	return nil
}

func (stack *AzureProductionAppsStack) createProductionStateStoreComponent(ctx *pulumi.Context) error {
	component, err := app.NewDaprComponent(ctx, "production-statestore", &app.DaprComponentArgs{
		ResourceGroupName:    stack.resourceGroup.Name,
		EnvironmentName:     stack.environment.Name,
		ComponentName:       pulumi.String("statestore"),
		ComponentType:       pulumi.String("state.redis"),
		Version:            pulumi.String("v1"),
		Metadata: app.DaprMetadataArray{
			&app.DaprMetadataArgs{
				Name:  pulumi.String("redisHost"),
				Value: pulumi.String("redis-production.redis.cache.windows.net:6380"),
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
				Value: pulumi.String("production"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("actorStateStore"),
				Value: pulumi.String("true"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("redisType"),
				Value: pulumi.String("cluster"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("dialTimeout"),
				Value: pulumi.String("10s"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("readTimeout"),
				Value: pulumi.String("30s"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("writeTimeout"),
				Value: pulumi.String("30s"),
			},
		},
		Secrets: app.SecretArray{
			&app.SecretArgs{
				Name:        pulumi.String("redis-password"),
				Value: pulumi.String(""), // TODO: Fix KeyVault integration in Azure Native SDK v1.104.0
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

func (stack *AzureProductionAppsStack) createProductionBindingsComponent(ctx *pulumi.Context) error {
	component, err := app.NewDaprComponent(ctx, "production-storage-binding", &app.DaprComponentArgs{
		ResourceGroupName:    stack.resourceGroup.Name,
		EnvironmentName:     stack.environment.Name,
		ComponentName:       pulumi.String("storage-binding"),
		ComponentType:       pulumi.String("bindings.azure.blobstorage"),
		Version:            pulumi.String("v1"),
		Metadata: app.DaprMetadataArray{
			&app.DaprMetadataArgs{
				Name:  pulumi.String("storageAccount"),
				Value: pulumi.String("internationalcenterproduction"),
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
				Value: pulumi.String("5"),
			},
		},
		Secrets: app.SecretArray{
			&app.SecretArgs{
				Name:        pulumi.String("storage-access-key"),
				Value: pulumi.String(""), // TODO: Fix KeyVault integration in Azure Native SDK v1.104.0
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

func (stack *AzureProductionAppsStack) createProductionApiContainerApps(ctx *pulumi.Context) error {
	apis := []string{"identity-api", "content-api", "services-api"}
	
	for _, apiName := range apis {
		if err := stack.createProductionApiContainerApp(ctx, apiName); err != nil {
			return fmt.Errorf("failed to create production %s: %w", apiName, err)
		}
	}

	return nil
}

func (stack *AzureProductionAppsStack) createProductionApiContainerApp(ctx *pulumi.Context, apiName string) error {
	containerApp, err := app.NewContainerApp(ctx, fmt.Sprintf("production-%s", apiName), &app.ContainerAppArgs{
		ResourceGroupName:    stack.resourceGroup.Name,
		ContainerAppName:     pulumi.String(apiName),
		ManagedEnvironmentId: stack.environment.ID(),
		Configuration: &app.ConfigurationArgs{
			Ingress: &app.IngressArgs{
				External:   pulumi.Bool(false),
				TargetPort: pulumi.Int(8080),
				Transport:  pulumi.String("http"),
				Traffic: app.TrafficWeightArray{
					&app.TrafficWeightArgs{
						RevisionName: pulumi.String(""),
						Weight:       pulumi.Int(100),
					},
				},
			},
			Dapr: &app.DaprArgs{
				Enabled:     pulumi.Bool(true),
				AppId:       pulumi.String(apiName),
				AppPort:     pulumi.Int(8080),
				AppProtocol: pulumi.String("http"),
				// TODO: Fix Dapr logging configuration in Azure Native SDK v1.104.0
				// EnableApiLogging: pulumi.Bool(true),
				// LogLevel:    pulumi.String("warn"),
			},
			Secrets: app.SecretArray{
				&app.SecretArgs{
					Name:  pulumi.String("database-connection"),
					Value: pulumi.String(""), // TODO: Fix KeyVault integration in Azure Native SDK v1.104.0
				},
			},
			Registries: app.RegistryCredentialsArray{
				&app.RegistryCredentialsArgs{
					Server:   pulumi.String("internationalcenterregistry.azurecr.io"),
					Username: pulumi.String(""),
					Identity: pulumi.String("system"),
				},
			},
		},
		Template: &app.TemplateArgs{
			Scale: &app.ScaleArgs{
				MinReplicas: pulumi.Int(3), // Higher availability for production
				MaxReplicas: pulumi.Int(50), // Higher scale capacity
				Rules: app.ScaleRuleArray{
					&app.ScaleRuleArgs{
						Name: pulumi.String("http-requests"),
						Http: &app.HttpScaleRuleArgs{
							Metadata: pulumi.StringMap{
								"concurrentRequests": pulumi.String("50"), // Conservative for production
							},
						},
					},
					&app.ScaleRuleArgs{
						Name: pulumi.String("cpu-utilization"),
						Custom: &app.CustomScaleRuleArgs{
							Type: pulumi.String("cpu"),
							Metadata: pulumi.StringMap{
								"type":  pulumi.String("Utilization"),
								"value": pulumi.String("60"), // Lower threshold for production
							},
						},
					},
					&app.ScaleRuleArgs{
						Name: pulumi.String("memory-utilization"),
						Custom: &app.CustomScaleRuleArgs{
							Type: pulumi.String("memory"),
							Metadata: pulumi.StringMap{
								"type":  pulumi.String("Utilization"),
								"value": pulumi.String("70"), // Lower threshold for production
							},
						},
					},
				},
			},
			Containers: app.ContainerArray{
				&app.ContainerArgs{
					Name:  pulumi.String(apiName),
					Image: pulumi.String(fmt.Sprintf("internationalcenterregistry.azurecr.io/%s:production", apiName)),
					Resources: &app.ContainerResourcesArgs{
						Cpu:    pulumi.Float64(1.0),
						Memory: pulumi.String("2Gi"),
					},
					Env: app.EnvironmentVarArray{
						&app.EnvironmentVarArgs{
							Name:  pulumi.String("ENVIRONMENT"),
							Value: pulumi.String("production"),
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
						&app.EnvironmentVarArgs{
							Name:  pulumi.String("LOG_LEVEL"),
							Value: pulumi.String("INFO"),
						},
						&app.EnvironmentVarArgs{
							Name:  pulumi.String("ENABLE_METRICS"),
							Value: pulumi.String("true"),
						},
						&app.EnvironmentVarArgs{
							Name:  pulumi.String("ENABLE_TRACING"),
							Value: pulumi.String("true"),
						},
						&app.EnvironmentVarArgs{
							Name:  pulumi.String("ENABLE_AUDIT_LOGGING"),
							Value: pulumi.String("true"),
						},
					},
					Probes: app.ContainerAppProbeArray{
						&app.ContainerAppProbeArgs{
							Type: pulumi.String("Liveness"),
							HttpGet: &app.ContainerAppProbeHttpGetArgs{
								Path:   pulumi.String("/health"),
								Port:   pulumi.Int(8080),
								Scheme: pulumi.String("HTTP"),
							},
							InitialDelaySeconds: pulumi.Int(60),
							PeriodSeconds:      pulumi.Int(30),
							TimeoutSeconds:     pulumi.Int(10),
							FailureThreshold:   pulumi.Int(3),
							SuccessThreshold:   pulumi.Int(1),
						},
						&app.ContainerAppProbeArgs{
							Type: pulumi.String("Readiness"),
							HttpGet: &app.ContainerAppProbeHttpGetArgs{
								Path:   pulumi.String("/ready"),
								Port:   pulumi.Int(8080),
								Scheme: pulumi.String("HTTP"),
							},
							InitialDelaySeconds: pulumi.Int(30),
							PeriodSeconds:      pulumi.Int(15),
							TimeoutSeconds:     pulumi.Int(5),
							FailureThreshold:   pulumi.Int(3),
							SuccessThreshold:   pulumi.Int(1),
						},
						&app.ContainerAppProbeArgs{
							Type: pulumi.String("Startup"),
							HttpGet: &app.ContainerAppProbeHttpGetArgs{
								Path:   pulumi.String("/startup"),
								Port:   pulumi.Int(8080),
								Scheme: pulumi.String("HTTP"),
							},
							InitialDelaySeconds: pulumi.Int(15),
							PeriodSeconds:      pulumi.Int(10),
							TimeoutSeconds:     pulumi.Int(5),
							FailureThreshold:   pulumi.Int(60),
						},
					},
				},
			},
		},
		Tags: pulumi.StringMap{
			"environment":      pulumi.String("production"),
			"service":         pulumi.String(apiName),
			"project":         pulumi.String("international-center"),
			"tier":           pulumi.String("api"),
			"compliance":     pulumi.String("required"),
			"backup-required": pulumi.String("true"),
		},
	})
	if err != nil {
		return err
	}

	stack.apps[apiName] = containerApp
	return nil
}

func (stack *AzureProductionAppsStack) createProductionGatewayContainerApps(ctx *pulumi.Context) error {
	gateways := []string{"public-gateway", "admin-gateway"}
	
	for _, gatewayName := range gateways {
		if err := stack.createProductionGatewayContainerApp(ctx, gatewayName); err != nil {
			return fmt.Errorf("failed to create production %s: %w", gatewayName, err)
		}
	}

	return nil
}

func (stack *AzureProductionAppsStack) createProductionGatewayContainerApp(ctx *pulumi.Context, gatewayName string) error {
	isPublic := gatewayName == "public-gateway"
	
	containerApp, err := app.NewContainerApp(ctx, fmt.Sprintf("production-%s", gatewayName), &app.ContainerAppArgs{
		ResourceGroupName:    stack.resourceGroup.Name,
		ContainerAppName:     pulumi.String(gatewayName),
		ManagedEnvironmentId: stack.environment.ID(),
		Configuration: &app.ConfigurationArgs{
			Ingress: &app.IngressArgs{
				External:   pulumi.Bool(isPublic),
				TargetPort: pulumi.Int(8080),
				Transport:  pulumi.String("http"),
				Traffic: app.TrafficWeightArray{
					&app.TrafficWeightArgs{
						RevisionName: pulumi.String(""),
						Weight:       pulumi.Int(100),
					},
				},
				CustomDomains: func() app.CustomDomainArray {
					if isPublic {
						return app.CustomDomainArray{
							&app.CustomDomainArgs{
								Name: pulumi.String("api.international-center.com"),
								CertificateId: pulumi.String(""),
								BindingType: pulumi.String("SniEnabled"),
							},
						}
					} else {
						return app.CustomDomainArray{
							&app.CustomDomainArgs{
								Name: pulumi.String("admin-api.international-center.com"),
								CertificateId: pulumi.String(""),
								BindingType: pulumi.String("SniEnabled"),
							},
						}
					}
				}(),
				CorsPolicy: &app.CorsPolicyArgs{
					AllowedOrigins: func() pulumi.StringArray {
						if isPublic {
							return pulumi.StringArray{
								pulumi.String("https://www.international-center.com"),
								pulumi.String("https://app.international-center.com"),
							}
						} else {
							return pulumi.StringArray{
								pulumi.String("https://admin.international-center.com"),
							}
						}
					}(),
					AllowedMethods: pulumi.StringArray{
						pulumi.String("GET"),
						pulumi.String("POST"),
						pulumi.String("PUT"),
						pulumi.String("DELETE"),
						pulumi.String("PATCH"),
						pulumi.String("OPTIONS"),
					},
					AllowedHeaders: pulumi.StringArray{
						pulumi.String("Content-Type"),
						pulumi.String("Authorization"),
						pulumi.String("X-Requested-With"),
						pulumi.String("X-Correlation-ID"),
						pulumi.String("X-Trace-ID"),
					},
					MaxAge: pulumi.Int(3600),
					AllowCredentials: pulumi.Bool(true),
				},
			},
			Dapr: &app.DaprArgs{
				Enabled:     pulumi.Bool(true),
				AppId:       pulumi.String(gatewayName),
				AppPort:     pulumi.Int(8080),
				AppProtocol: pulumi.String("http"),
				// TODO: Fix Dapr logging configuration in Azure Native SDK v1.104.0
				// EnableApiLogging: pulumi.Bool(true),
				// LogLevel:    pulumi.String("warn"),
			},
			Secrets: app.SecretArray{
				&app.SecretArgs{
					Name:        pulumi.String("jwt-secret"),
					Value: pulumi.String(""), // TODO: Fix KeyVault integration in Azure Native SDK v1.104.0
				},
				&app.SecretArgs{
					Name:        pulumi.String("api-keys"),
					Value: pulumi.String(""), // TODO: Fix KeyVault integration in Azure Native SDK v1.104.0
				},
			},
			Registries: app.RegistryCredentialsArray{
				&app.RegistryCredentialsArgs{
					Server:   pulumi.String("internationalcenterregistry.azurecr.io"),
					Username: pulumi.String(""),
					Identity: pulumi.String("system"),
				},
			},
		},
		Template: &app.TemplateArgs{
			Scale: &app.ScaleArgs{
				MinReplicas: pulumi.Int(5), // Higher availability for production gateways
				MaxReplicas: pulumi.Int(100), // Higher scale capacity for production
				Rules: app.ScaleRuleArray{
					&app.ScaleRuleArgs{
						Name: pulumi.String("http-requests"),
						Http: &app.HttpScaleRuleArgs{
							Metadata: pulumi.StringMap{
								"concurrentRequests": pulumi.String("200"), // Higher capacity for production
							},
						},
					},
					&app.ScaleRuleArgs{
						Name: pulumi.String("cpu-utilization"),
						Custom: &app.CustomScaleRuleArgs{
							Type: pulumi.String("cpu"),
							Metadata: pulumi.StringMap{
								"type":  pulumi.String("Utilization"),
								"value": pulumi.String("50"), // Lower threshold for production
							},
						},
					},
				},
			},
			Containers: app.ContainerArray{
				&app.ContainerArgs{
					Name:  pulumi.String(gatewayName),
					Image: pulumi.String(fmt.Sprintf("internationalcenterregistry.azurecr.io/%s:production", gatewayName)),
					Resources: &app.ContainerResourcesArgs{
						Cpu:    pulumi.Float64(1.5),
						Memory: pulumi.String("3Gi"),
					},
					Env: app.EnvironmentVarArray{
						&app.EnvironmentVarArgs{
							Name:  pulumi.String("ENVIRONMENT"),
							Value: pulumi.String("production"),
						},
						&app.EnvironmentVarArgs{
							Name:  pulumi.String("GATEWAY_TYPE"),
							Value: pulumi.String(gatewayName),
						},
						&app.EnvironmentVarArgs{
							Name:      pulumi.String("JWT_SECRET"),
							SecretRef: pulumi.String("jwt-secret"),
						},
						&app.EnvironmentVarArgs{
							Name:      pulumi.String("API_KEYS"),
							SecretRef: pulumi.String("api-keys"),
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
							Name:  pulumi.String("ENABLE_HTTPS"),
							Value: pulumi.String("true"),
						},
						&app.EnvironmentVarArgs{
							Name:  pulumi.String("REQUIRE_AUTH"),
							Value: pulumi.String("true"),
						},
						&app.EnvironmentVarArgs{
							Name:  pulumi.String("ENABLE_CSRF"),
							Value: pulumi.String("true"),
						},
						&app.EnvironmentVarArgs{
							Name:  pulumi.String("ENABLE_AUDIT_LOGGING"),
							Value: pulumi.String("true"),
						},
						&app.EnvironmentVarArgs{
							Name: pulumi.String("RATE_LIMIT"),
							Value: func() pulumi.String {
								if isPublic {
									return pulumi.String("1000") // 1000 req/min for public gateway
								}
								return pulumi.String("100") // 100 req/min for admin gateway
							}(),
						},
					},
					Probes: app.ContainerAppProbeArray{
						&app.ContainerAppProbeArgs{
							Type: pulumi.String("Liveness"),
							HttpGet: &app.ContainerAppProbeHttpGetArgs{
								Path:   pulumi.String("/health"),
								Port:   pulumi.Int(8080),
								Scheme: pulumi.String("HTTP"),
							},
							InitialDelaySeconds: pulumi.Int(90),
							PeriodSeconds:      pulumi.Int(30),
							TimeoutSeconds:     pulumi.Int(15),
							FailureThreshold:   pulumi.Int(3),
							SuccessThreshold:   pulumi.Int(1),
						},
						&app.ContainerAppProbeArgs{
							Type: pulumi.String("Readiness"),
							HttpGet: &app.ContainerAppProbeHttpGetArgs{
								Path:   pulumi.String("/ready"),
								Port:   pulumi.Int(8080),
								Scheme: pulumi.String("HTTP"),
							},
							InitialDelaySeconds: pulumi.Int(45),
							PeriodSeconds:      pulumi.Int(15),
							TimeoutSeconds:     pulumi.Int(10),
							FailureThreshold:   pulumi.Int(3),
							SuccessThreshold:   pulumi.Int(1),
						},
					},
				},
			},
		},
		Tags: pulumi.StringMap{
			"environment":      pulumi.String("production"),
			"service":         pulumi.String(gatewayName),
			"project":         pulumi.String("international-center"),
			"tier":           pulumi.String("gateway"),
			"compliance":     pulumi.String("required"),
			"backup-required": pulumi.String("true"),
		},
	})
	if err != nil {
		return err
	}

	stack.apps[gatewayName] = containerApp
	return nil
}

func (stack *AzureProductionAppsStack) GetResourceGroup() *resources.ResourceGroup {
	return stack.resourceGroup
}

func (stack *AzureProductionAppsStack) GetEnvironment() *app.ManagedEnvironment {
	return stack.environment
}

func (stack *AzureProductionAppsStack) GetVirtualNetwork() *network.VirtualNetwork {
	return stack.vnet
}

func (stack *AzureProductionAppsStack) GetPrivateSubnet() *network.Subnet {
	return stack.privateSubnet
}

func (stack *AzureProductionAppsStack) GetContainerApp(name string) *app.ContainerApp {
	return stack.apps[name]
}

func (stack *AzureProductionAppsStack) GetDaprComponent(name string) *app.DaprComponent {
	return stack.daprComponents[name]
}