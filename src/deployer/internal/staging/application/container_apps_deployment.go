package application

import (
	"context"
	"fmt"

	"github.com/pulumi/pulumi-azure-native/sdk/v2/go/azurenative/app"
	"github.com/pulumi/pulumi-azure-native/sdk/v2/go/azurenative/resources"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/international-center/src/deployer/internal/staging/infrastructure"
)

type ContainerAppsDeployment struct {
	containerAppsStack *infrastructure.AzureContainerAppsStack
	databaseStack     *infrastructure.AzureDatabaseStack
	storageStack      *infrastructure.AzureStorageStack
	vaultStack        *infrastructure.VaultCloudStack
	grafanaStack      *infrastructure.GrafanaCloudStack
	deployedApps      map[string]*app.ContainerApp
	deployedGateways  map[string]*app.ContainerApp
}

type DeploymentConfiguration struct {
	Environment          string
	ContainerRegistry    string
	ImageTag            string
	DatabaseConnections map[string]string
	EnableMetrics       bool
	EnableTracing       bool
	ScalingRules        map[string]*ScalingConfiguration
	SecuritySettings    *SecurityConfiguration
}

type ScalingConfiguration struct {
	MinReplicas     int
	MaxReplicas     int
	ConcurrentRequests int
	CPUUtilization  int
	MemoryUtilization int
}

type SecurityConfiguration struct {
	EnableHTTPS        bool
	RequireAuthentication bool
	AllowedOrigins     []string
	RateLimits         map[string]int
	EnableCSRF         bool
}

func NewContainerAppsDeployment(
	containerAppsStack *infrastructure.AzureContainerAppsStack,
	databaseStack *infrastructure.AzureDatabaseStack,
	storageStack *infrastructure.AzureStorageStack,
	vaultStack *infrastructure.VaultCloudStack,
	grafanaStack *infrastructure.GrafanaCloudStack,
) *ContainerAppsDeployment {
	return &ContainerAppsDeployment{
		containerAppsStack: containerAppsStack,
		databaseStack:     databaseStack,
		storageStack:      storageStack,
		vaultStack:        vaultStack,
		grafanaStack:      grafanaStack,
		deployedApps:      make(map[string]*app.ContainerApp),
		deployedGateways:  make(map[string]*app.ContainerApp),
	}
}

func (deployment *ContainerAppsDeployment) Deploy(ctx *pulumi.Context) error {
	config := deployment.getDeploymentConfiguration()

	if err := deployment.deployApiServices(ctx, config); err != nil {
		return fmt.Errorf("failed to deploy API services: %w", err)
	}

	if err := deployment.deployGatewayServices(ctx, config); err != nil {
		return fmt.Errorf("failed to deploy gateway services: %w", err)
	}

	return nil
}

func (deployment *ContainerAppsDeployment) getDeploymentConfiguration() *DeploymentConfiguration {
	return &DeploymentConfiguration{
		Environment:       "staging",
		ContainerRegistry: "internationalcenterregistry.azurecr.io",
		ImageTag:         "staging",
		DatabaseConnections: map[string]string{
			"identity": "Server=international-center-staging-db.postgres.database.azure.com;Database=identity_staging;Port=5432;User Id=dbadmin;Password={vault:database-admin-password};Ssl Mode=Require;",
			"content":  "Server=international-center-staging-db.postgres.database.azure.com;Database=content_staging;Port=5432;User Id=dbadmin;Password={vault:database-admin-password};Ssl Mode=Require;",
			"services": "Server=international-center-staging-db.postgres.database.azure.com;Database=services_staging;Port=5432;User Id=dbadmin;Password={vault:database-admin-password};Ssl Mode=Require;",
		},
		EnableMetrics: true,
		EnableTracing: true,
		ScalingRules: map[string]*ScalingConfiguration{
			"identity-api": {
				MinReplicas:        2,
				MaxReplicas:        15,
				ConcurrentRequests: 25,
				CPUUtilization:     70,
				MemoryUtilization:  80,
			},
			"content-api": {
				MinReplicas:        2,
				MaxReplicas:        20,
				ConcurrentRequests: 30,
				CPUUtilization:     70,
				MemoryUtilization:  80,
			},
			"services-api": {
				MinReplicas:        2,
				MaxReplicas:        10,
				ConcurrentRequests: 20,
				CPUUtilization:     70,
				MemoryUtilization:  80,
			},
			"public-gateway": {
				MinReplicas:        3,
				MaxReplicas:        25,
				ConcurrentRequests: 100,
				CPUUtilization:     65,
				MemoryUtilization:  75,
			},
			"admin-gateway": {
				MinReplicas:        2,
				MaxReplicas:        15,
				ConcurrentRequests: 50,
				CPUUtilization:     70,
				MemoryUtilization:  80,
			},
		},
		SecuritySettings: &SecurityConfiguration{
			EnableHTTPS:           true,
			RequireAuthentication: true,
			AllowedOrigins: []string{
				"https://staging.international-center.com",
				"https://app-staging.international-center.com",
				"https://admin-staging.international-center.com",
			},
			RateLimits: map[string]int{
				"public":  1000, // requests per minute
				"admin":   500,
				"api":     2000,
			},
			EnableCSRF: true,
		},
	}
}

func (deployment *ContainerAppsDeployment) deployApiServices(ctx *pulumi.Context, config *DeploymentConfiguration) error {
	apis := []string{"identity-api", "content-api", "services-api"}
	
	for _, apiName := range apis {
		if err := deployment.deployApiService(ctx, apiName, config); err != nil {
			return fmt.Errorf("failed to deploy %s: %w", apiName, err)
		}
	}

	return nil
}

func (deployment *ContainerAppsDeployment) deployApiService(ctx *pulumi.Context, serviceName string, config *DeploymentConfiguration) error {
	scalingConfig := config.ScalingRules[serviceName]
	
	containerApp, err := app.NewContainerApp(ctx, fmt.Sprintf("staging-%s", serviceName), &app.ContainerAppArgs{
		ResourceGroupName:    deployment.containerAppsStack.GetResourceGroup().Name,
		ContainerAppName:     pulumi.String(serviceName),
		ManagedEnvironmentId: deployment.containerAppsStack.GetEnvironment().ID(),
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
				CustomDomains: app.CustomDomainArray{},
			},
			Dapr: &app.DaprArgs{
				Enabled:     pulumi.Bool(true),
				AppId:       pulumi.String(serviceName),
				AppPort:     pulumi.Int(8080),
				AppProtocol: pulumi.String("http"),
				EnableApiLogging: pulumi.Bool(true),
				LogLevel:    pulumi.String("info"),
			},
			Secrets: app.SecretArray{
				&app.SecretArgs{
					Name:  pulumi.String("database-connection"),
					Value: pulumi.String(config.DatabaseConnections[deployment.getDomainForService(serviceName)]),
				},
				&app.SecretArgs{
					Name:      pulumi.String("redis-connection"),
					KeyVaultUrl: deployment.vaultStack.GetVaultUri(),
					Identity:    pulumi.String("system"),
				},
			},
			Registries: app.RegistryCredentialsArray{
				&app.RegistryCredentialsArgs{
					Server:   pulumi.String(config.ContainerRegistry),
					Username: pulumi.String(""), // Managed identity
					Identity: pulumi.String("system"),
				},
			},
		},
		Template: &app.TemplateArgs{
			Scale: &app.ScaleArgs{
				MinReplicas: pulumi.Int(scalingConfig.MinReplicas),
				MaxReplicas: pulumi.Int(scalingConfig.MaxReplicas),
				Rules: app.ScaleRuleArray{
					&app.ScaleRuleArgs{
						Name: pulumi.String("http-requests"),
						Http: &app.HttpScaleRuleArgs{
							Metadata: pulumi.StringMap{
								"concurrentRequests": pulumi.String(fmt.Sprintf("%d", scalingConfig.ConcurrentRequests)),
							},
						},
					},
					&app.ScaleRuleArgs{
						Name: pulumi.String("cpu-utilization"),
						Custom: &app.CustomScaleRuleArgs{
							Type: pulumi.String("cpu"),
							Metadata: pulumi.StringMap{
								"type":            pulumi.String("Utilization"),
								"value":           pulumi.String(fmt.Sprintf("%d", scalingConfig.CPUUtilization)),
							},
						},
					},
					&app.ScaleRuleArgs{
						Name: pulumi.String("memory-utilization"),
						Custom: &app.CustomScaleRuleArgs{
							Type: pulumi.String("memory"),
							Metadata: pulumi.StringMap{
								"type":  pulumi.String("Utilization"),
								"value": pulumi.String(fmt.Sprintf("%d", scalingConfig.MemoryUtilization)),
							},
						},
					},
				},
			},
			Containers: app.ContainerArray{
				&app.ContainerArgs{
					Name:  pulumi.String(serviceName),
					Image: pulumi.String(fmt.Sprintf("%s/%s:%s", config.ContainerRegistry, serviceName, config.ImageTag)),
					Resources: &app.ContainerResourcesArgs{
						Cpu:    pulumi.Float64(0.75),
						Memory: pulumi.String("1.5Gi"),
					},
					Env: deployment.getServiceEnvironmentVariables(serviceName, config),
					Probes: app.ContainerAppProbeArray{
						&app.ContainerAppProbeArgs{
							Type: pulumi.String("Liveness"),
							HttpGet: &app.ContainerAppProbeHttpGetArgs{
								Path:   pulumi.String("/health"),
								Port:   pulumi.Int(8080),
								Scheme: pulumi.String("HTTP"),
							},
							InitialDelaySeconds: pulumi.Int(45),
							PeriodSeconds:      pulumi.Int(15),
							TimeoutSeconds:     pulumi.Int(5),
							FailureThreshold:   pulumi.Int(5),
						},
						&app.ContainerAppProbeArgs{
							Type: pulumi.String("Readiness"),
							HttpGet: &app.ContainerAppProbeHttpGetArgs{
								Path:   pulumi.String("/ready"),
								Port:   pulumi.Int(8080),
								Scheme: pulumi.String("HTTP"),
							},
							InitialDelaySeconds: pulumi.Int(10),
							PeriodSeconds:      pulumi.Int(10),
							TimeoutSeconds:     pulumi.Int(3),
							FailureThreshold:   pulumi.Int(3),
						},
						&app.ContainerAppProbeArgs{
							Type: pulumi.String("Startup"),
							HttpGet: &app.ContainerAppProbeHttpGetArgs{
								Path:   pulumi.String("/startup"),
								Port:   pulumi.Int(8080),
								Scheme: pulumi.String("HTTP"),
							},
							InitialDelaySeconds: pulumi.Int(10),
							PeriodSeconds:      pulumi.Int(5),
							TimeoutSeconds:     pulumi.Int(3),
							FailureThreshold:   pulumi.Int(30),
						},
					},
				},
			},
		},
		Tags: pulumi.StringMap{
			"environment": pulumi.String("staging"),
			"service":     pulumi.String(serviceName),
			"project":     pulumi.String("international-center"),
			"tier":        pulumi.String("api"),
		},
	})
	if err != nil {
		return err
	}

	deployment.deployedApps[serviceName] = containerApp
	return nil
}

func (deployment *ContainerAppsDeployment) deployGatewayServices(ctx *pulumi.Context, config *DeploymentConfiguration) error {
	gateways := []string{"public-gateway", "admin-gateway"}
	
	for _, gatewayName := range gateways {
		if err := deployment.deployGatewayService(ctx, gatewayName, config); err != nil {
			return fmt.Errorf("failed to deploy %s: %w", gatewayName, err)
		}
	}

	return nil
}

func (deployment *ContainerAppsDeployment) deployGatewayService(ctx *pulumi.Context, gatewayName string, config *DeploymentConfiguration) error {
	isPublic := gatewayName == "public-gateway"
	scalingConfig := config.ScalingRules[gatewayName]
	
	containerApp, err := app.NewContainerApp(ctx, fmt.Sprintf("staging-%s", gatewayName), &app.ContainerAppArgs{
		ResourceGroupName:    deployment.containerAppsStack.GetResourceGroup().Name,
		ContainerAppName:     pulumi.String(gatewayName),
		ManagedEnvironmentId: deployment.containerAppsStack.GetEnvironment().ID(),
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
								Name: pulumi.String("api-staging.international-center.com"),
								CertificateId: pulumi.String(""),
								BindingType: pulumi.String("SniEnabled"),
							},
						}
					} else {
						return app.CustomDomainArray{
							&app.CustomDomainArgs{
								Name: pulumi.String("admin-api-staging.international-center.com"),
								CertificateId: pulumi.String(""),
								BindingType: pulumi.String("SniEnabled"),
							},
						}
					}
				}(),
				CorsPolicy: &app.CorsPolicyArgs{
					AllowedOrigins: func() pulumi.StringArray {
						origins := pulumi.StringArray{}
						for _, origin := range config.SecuritySettings.AllowedOrigins {
							origins = append(origins, pulumi.String(origin))
						}
						return origins
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
						pulumi.String("*"),
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
				EnableApiLogging: pulumi.Bool(true),
				LogLevel:    pulumi.String("info"),
			},
			Secrets: app.SecretArray{
				&app.SecretArgs{
					Name:        pulumi.String("jwt-secret"),
					KeyVaultUrl: deployment.vaultStack.GetVaultUri(),
					Identity:    pulumi.String("system"),
				},
				&app.SecretArgs{
					Name:        pulumi.String("api-keys"),
					KeyVaultUrl: deployment.vaultStack.GetVaultUri(),
					Identity:    pulumi.String("system"),
				},
			},
			Registries: app.RegistryCredentialsArray{
				&app.RegistryCredentialsArgs{
					Server:   pulumi.String(config.ContainerRegistry),
					Username: pulumi.String(""),
					Identity: pulumi.String("system"),
				},
			},
		},
		Template: &app.TemplateArgs{
			Scale: &app.ScaleArgs{
				MinReplicas: pulumi.Int(scalingConfig.MinReplicas),
				MaxReplicas: pulumi.Int(scalingConfig.MaxReplicas),
				Rules: app.ScaleRuleArray{
					&app.ScaleRuleArgs{
						Name: pulumi.String("http-requests"),
						Http: &app.HttpScaleRuleArgs{
							Metadata: pulumi.StringMap{
								"concurrentRequests": pulumi.String(fmt.Sprintf("%d", scalingConfig.ConcurrentRequests)),
							},
						},
					},
					&app.ScaleRuleArgs{
						Name: pulumi.String("cpu-utilization"),
						Custom: &app.CustomScaleRuleArgs{
							Type: pulumi.String("cpu"),
							Metadata: pulumi.StringMap{
								"type":  pulumi.String("Utilization"),
								"value": pulumi.String(fmt.Sprintf("%d", scalingConfig.CPUUtilization)),
							},
						},
					},
				},
			},
			Containers: app.ContainerArray{
				&app.ContainerArgs{
					Name:  pulumi.String(gatewayName),
					Image: pulumi.String(fmt.Sprintf("%s/%s:%s", config.ContainerRegistry, gatewayName, config.ImageTag)),
					Resources: &app.ContainerResourcesArgs{
						Cpu:    pulumi.Float64(1.0),
						Memory: pulumi.String("2Gi"),
					},
					Env: deployment.getGatewayEnvironmentVariables(gatewayName, config),
					Probes: app.ContainerAppProbeArray{
						&app.ContainerAppProbeArgs{
							Type: pulumi.String("Liveness"),
							HttpGet: &app.ContainerAppProbeHttpGetArgs{
								Path:   pulumi.String("/health"),
								Port:   pulumi.Int(8080),
								Scheme: pulumi.String("HTTP"),
							},
							InitialDelaySeconds: pulumi.Int(60),
							PeriodSeconds:      pulumi.Int(20),
							TimeoutSeconds:     pulumi.Int(10),
							FailureThreshold:   pulumi.Int(3),
						},
						&app.ContainerAppProbeArgs{
							Type: pulumi.String("Readiness"),
							HttpGet: &app.ContainerAppProbeHttpGetArgs{
								Path:   pulumi.String("/ready"),
								Port:   pulumi.Int(8080),
								Scheme: pulumi.String("HTTP"),
							},
							InitialDelaySeconds: pulumi.Int(15),
							PeriodSeconds:      pulumi.Int(10),
							TimeoutSeconds:     pulumi.Int(5),
							FailureThreshold:   pulumi.Int(3),
						},
					},
				},
			},
		},
		Tags: pulumi.StringMap{
			"environment": pulumi.String("staging"),
			"service":     pulumi.String(gatewayName),
			"project":     pulumi.String("international-center"),
			"tier":        pulumi.String("gateway"),
		},
	})
	if err != nil {
		return err
	}

	deployment.deployedGateways[gatewayName] = containerApp
	return nil
}

func (deployment *ContainerAppsDeployment) getServiceEnvironmentVariables(serviceName string, config *DeploymentConfiguration) app.EnvironmentVarArray {
	baseVars := app.EnvironmentVarArray{
		&app.EnvironmentVarArgs{
			Name:  pulumi.String("ENVIRONMENT"),
			Value: pulumi.String(config.Environment),
		},
		&app.EnvironmentVarArgs{
			Name:      pulumi.String("DATABASE_CONNECTION"),
			SecretRef: pulumi.String("database-connection"),
		},
		&app.EnvironmentVarArgs{
			Name:      pulumi.String("REDIS_CONNECTION"),
			SecretRef: pulumi.String("redis-connection"),
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
			Name:  pulumi.String("SERVICE_NAME"),
			Value: pulumi.String(serviceName),
		},
		&app.EnvironmentVarArgs{
			Name:  pulumi.String("ENABLE_METRICS"),
			Value: pulumi.String(fmt.Sprintf("%t", config.EnableMetrics)),
		},
		&app.EnvironmentVarArgs{
			Name:  pulumi.String("ENABLE_TRACING"),
			Value: pulumi.String(fmt.Sprintf("%t", config.EnableTracing)),
		},
		&app.EnvironmentVarArgs{
			Name:  pulumi.String("PROMETHEUS_ENDPOINT"),
			Value: deployment.grafanaStack.GetPrometheusUrl(),
		},
		&app.EnvironmentVarArgs{
			Name:  pulumi.String("LOKI_ENDPOINT"),
			Value: deployment.grafanaStack.GetLokiUrl(),
		},
	}

	return baseVars
}

func (deployment *ContainerAppsDeployment) getGatewayEnvironmentVariables(gatewayName string, config *DeploymentConfiguration) app.EnvironmentVarArray {
	isPublic := gatewayName == "public-gateway"
	
	baseVars := app.EnvironmentVarArray{
		&app.EnvironmentVarArgs{
			Name:  pulumi.String("ENVIRONMENT"),
			Value: pulumi.String(config.Environment),
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
			Value: pulumi.String(fmt.Sprintf("%t", config.SecuritySettings.EnableHTTPS)),
		},
		&app.EnvironmentVarArgs{
			Name:  pulumi.String("REQUIRE_AUTH"),
			Value: pulumi.String(fmt.Sprintf("%t", config.SecuritySettings.RequireAuthentication)),
		},
		&app.EnvironmentVarArgs{
			Name:  pulumi.String("ENABLE_CSRF"),
			Value: pulumi.String(fmt.Sprintf("%t", config.SecuritySettings.EnableCSRF)),
		},
		&app.EnvironmentVarArgs{
			Name: pulumi.String("RATE_LIMIT"),
			Value: func() pulumi.String {
				if isPublic {
					return pulumi.String(fmt.Sprintf("%d", config.SecuritySettings.RateLimits["public"]))
				}
				return pulumi.String(fmt.Sprintf("%d", config.SecuritySettings.RateLimits["admin"]))
			}(),
		},
		&app.EnvironmentVarArgs{
			Name:  pulumi.String("IDENTITY_API_ENDPOINT"),
			Value: pulumi.String("http://identity-api:8080"),
		},
		&app.EnvironmentVarArgs{
			Name:  pulumi.String("CONTENT_API_ENDPOINT"),
			Value: pulumi.String("http://content-api:8080"),
		},
		&app.EnvironmentVarArgs{
			Name:  pulumi.String("SERVICES_API_ENDPOINT"),
			Value: pulumi.String("http://services-api:8080"),
		},
	}

	return baseVars
}

func (deployment *ContainerAppsDeployment) getDomainForService(serviceName string) string {
	domainMap := map[string]string{
		"identity-api": "identity",
		"content-api":  "content",
		"services-api": "services",
	}
	
	return domainMap[serviceName]
}

func (deployment *ContainerAppsDeployment) GetDeployedApp(name string) *app.ContainerApp {
	return deployment.deployedApps[name]
}

func (deployment *ContainerAppsDeployment) GetDeployedGateway(name string) *app.ContainerApp {
	return deployment.deployedGateways[name]
}