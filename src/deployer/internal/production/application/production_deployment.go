package application

import (
	"context"
	"fmt"

	"github.com/pulumi/pulumi-azure-native/sdk/v2/go/azurenative/app"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/international-center/src/deployer/internal/production/infrastructure"
)

type ProductionDeployment struct {
	containerAppsStack *infrastructure.AzureProductionAppsStack
	databaseStack     *infrastructure.AzureProductionDatabaseStack
	storageStack      *infrastructure.AzureProductionStorageStack
	vaultStack        *infrastructure.VaultProductionStack
	grafanaStack      *infrastructure.GrafanaProductionStack
	deployedApps      map[string]*app.ContainerApp
	deployedGateways  map[string]*app.ContainerApp
	securityConfig    *ProductionSecurityConfiguration
	complianceConfig  *ProductionComplianceConfiguration
}

type ProductionDeploymentConfiguration struct {
	Environment                string
	ContainerRegistry         string
	ImageTag                  string
	DatabaseConnections       map[string]string
	ReadReplicaConnections    map[string]string
	EnableMetrics             bool
	EnableTracing             bool
	EnableAuditLogging        bool
	EnableSecurityScanning    bool
	EnableComplianceMonitoring bool
	ScalingRules              map[string]*ProductionScalingConfiguration
	SecuritySettings          *ProductionSecurityConfiguration
	ComplianceSettings        *ProductionComplianceConfiguration
}

type ProductionScalingConfiguration struct {
	MinReplicas              int
	MaxReplicas              int
	TargetCPUUtilization     int
	TargetMemoryUtilization  int
	ConcurrentRequests       int
	ScaleUpCooldown         string
	ScaleDownCooldown       string
	EnablePredictiveScaling  bool
}

type ProductionSecurityConfiguration struct {
	EnforceHTTPS                    bool
	RequireAuthentication           bool
	EnableAdvancedThreatProtection  bool
	EnableSecurityHeaders           bool
	AllowedOrigins                  []string
	RateLimits                      map[string]int
	EnableCSRF                      bool
	EnableCORS                      bool
	EnableIPWhitelisting            bool
	AllowedIPRanges                 []string
	EnableCertificatePinning        bool
	SecurityScanSchedule            string
	VulnerabilityThreshold          string
}

type ProductionComplianceConfiguration struct {
	EnableAuditLogging         bool
	AuditRetentionDays         int
	EnableDataClassification   bool
	EnableAccessLogging        bool
	EnableIntegrityChecks      bool
	ComplianceFrameworks       []string
	DataRetentionPolicies      map[string]int
	EnableEncryptionAtRest     bool
	EnableEncryptionInTransit  bool
	RequireApprovalWorkflows   bool
}

func NewProductionDeployment(
	containerAppsStack *infrastructure.AzureProductionAppsStack,
	databaseStack *infrastructure.AzureProductionDatabaseStack,
	storageStack *infrastructure.AzureProductionStorageStack,
	vaultStack *infrastructure.VaultProductionStack,
	grafanaStack *infrastructure.GrafanaProductionStack,
) *ProductionDeployment {
	return &ProductionDeployment{
		containerAppsStack: containerAppsStack,
		databaseStack:     databaseStack,
		storageStack:      storageStack,
		vaultStack:        vaultStack,
		grafanaStack:      grafanaStack,
		deployedApps:      make(map[string]*app.ContainerApp),
		deployedGateways:  make(map[string]*app.ContainerApp),
		securityConfig:    getProductionSecurityConfiguration(),
		complianceConfig:  getProductionComplianceConfiguration(),
	}
}

func (deployment *ProductionDeployment) Deploy(ctx *pulumi.Context) error {
	config := deployment.getProductionDeploymentConfiguration()

	if err := deployment.deployProductionApiServices(ctx, config); err != nil {
		return fmt.Errorf("failed to deploy production API services: %w", err)
	}

	if err := deployment.deployProductionGatewayServices(ctx, config); err != nil {
		return fmt.Errorf("failed to deploy production gateway services: %w", err)
	}

	if err := deployment.configureProductionMonitoring(ctx, config); err != nil {
		return fmt.Errorf("failed to configure production monitoring: %w", err)
	}

	if err := deployment.enableSecurityCompliance(ctx, config); err != nil {
		return fmt.Errorf("failed to enable security compliance: %w", err)
	}

	return nil
}

func (deployment *ProductionDeployment) getProductionDeploymentConfiguration() *ProductionDeploymentConfiguration {
	return &ProductionDeploymentConfiguration{
		Environment:       "production",
		ContainerRegistry: "internationalcenterregistry.azurecr.io",
		ImageTag:         "production",
		DatabaseConnections: map[string]string{
			"identity": "Server=international-center-production-db.postgres.database.azure.com;Database=identity_production;Port=5432;User Id=dbadmin;Password={vault:database-admin-password};Ssl Mode=Require;Trust Server Certificate=false;",
			"content":  "Server=international-center-production-db.postgres.database.azure.com;Database=content_production;Port=5432;User Id=dbadmin;Password={vault:database-admin-password};Ssl Mode=Require;Trust Server Certificate=false;",
			"services": "Server=international-center-production-db.postgres.database.azure.com;Database=services_production;Port=5432;User Id=dbadmin;Password={vault:database-admin-password};Ssl Mode=Require;Trust Server Certificate=false;",
		},
		ReadReplicaConnections: map[string]string{
			"identity": "Server=international-center-production-db-replica.postgres.database.azure.com;Database=identity_production;Port=5432;User Id=dbadmin;Password={vault:database-admin-password};Ssl Mode=Require;Trust Server Certificate=false;ApplicationName=ReadReplica;",
			"content":  "Server=international-center-production-db-replica.postgres.database.azure.com;Database=content_production;Port=5432;User Id=dbadmin;Password={vault:database-admin-password};Ssl Mode=Require;Trust Server Certificate=false;ApplicationName=ReadReplica;",
			"services": "Server=international-center-production-db-replica.postgres.database.azure.com;Database=services_production;Port=5432;User Id=dbadmin;Password={vault:database-admin-password};Ssl Mode=Require;Trust Server Certificate=false;ApplicationName=ReadReplica;",
		},
		EnableMetrics:              true,
		EnableTracing:              true,
		EnableAuditLogging:         true,
		EnableSecurityScanning:     true,
		EnableComplianceMonitoring: true,
		ScalingRules: map[string]*ProductionScalingConfiguration{
			"identity-api": {
				MinReplicas:              5,  // High availability
				MaxReplicas:              100,
				TargetCPUUtilization:     50, // Conservative threshold
				TargetMemoryUtilization:  60,
				ConcurrentRequests:       25,
				ScaleUpCooldown:         "30s",
				ScaleDownCooldown:       "300s", // Conservative scale-down
				EnablePredictiveScaling:  true,
			},
			"content-api": {
				MinReplicas:              5,
				MaxReplicas:              150, // Higher capacity for content
				TargetCPUUtilization:     50,
				TargetMemoryUtilization:  60,
				ConcurrentRequests:       30,
				ScaleUpCooldown:         "30s",
				ScaleDownCooldown:       "300s",
				EnablePredictiveScaling:  true,
			},
			"services-api": {
				MinReplicas:              3,
				MaxReplicas:              75,
				TargetCPUUtilization:     50,
				TargetMemoryUtilization:  60,
				ConcurrentRequests:       20,
				ScaleUpCooldown:         "30s",
				ScaleDownCooldown:       "300s",
				EnablePredictiveScaling:  true,
			},
			"public-gateway": {
				MinReplicas:              10, // Very high availability for public gateway
				MaxReplicas:              200,
				TargetCPUUtilization:     40, // Very conservative
				TargetMemoryUtilization:  50,
				ConcurrentRequests:       100,
				ScaleUpCooldown:         "15s", // Fast scale-up for traffic spikes
				ScaleDownCooldown:       "600s", // Very slow scale-down
				EnablePredictiveScaling:  true,
			},
			"admin-gateway": {
				MinReplicas:              5,
				MaxReplicas:              50,
				TargetCPUUtilization:     50,
				TargetMemoryUtilization:  60,
				ConcurrentRequests:       50,
				ScaleUpCooldown:         "30s",
				ScaleDownCooldown:       "300s",
				EnablePredictiveScaling:  true,
			},
		},
		SecuritySettings:   deployment.securityConfig,
		ComplianceSettings: deployment.complianceConfig,
	}
}

func getProductionSecurityConfiguration() *ProductionSecurityConfiguration {
	return &ProductionSecurityConfiguration{
		EnforceHTTPS:                   true,
		RequireAuthentication:          true,
		EnableAdvancedThreatProtection: true,
		EnableSecurityHeaders:          true,
		AllowedOrigins: []string{
			"https://www.international-center.com",
			"https://app.international-center.com",
			"https://admin.international-center.com",
		},
		RateLimits: map[string]int{
			"public":          1000, // 1000 requests per minute for public
			"admin":           100,  // 100 requests per minute for admin
			"authenticated":   500,  // 500 requests per minute for authenticated users
			"anonymous":       50,   // 50 requests per minute for anonymous
		},
		EnableCSRF:               true,
		EnableCORS:               true,
		EnableIPWhitelisting:     true,
		AllowedIPRanges: []string{
			"10.0.0.0/8",     // Private networks
			"172.16.0.0/12",  // Private networks
			"192.168.0.0/16", // Private networks
		},
		EnableCertificatePinning: true,
		SecurityScanSchedule:     "0 2 * * *", // Daily at 2 AM
		VulnerabilityThreshold:   "MEDIUM",    // Alert on medium+ vulnerabilities
	}
}

func getProductionComplianceConfiguration() *ProductionComplianceConfiguration {
	return &ProductionComplianceConfiguration{
		EnableAuditLogging:       true,
		AuditRetentionDays:      2555, // 7 years retention
		EnableDataClassification: true,
		EnableAccessLogging:     true,
		EnableIntegrityChecks:   true,
		ComplianceFrameworks: []string{
			"SOC2-Type2",
			"ISO27001",
			"GDPR",
			"CCPA",
		},
		DataRetentionPolicies: map[string]int{
			"audit_logs":      2555, // 7 years
			"access_logs":     365,  // 1 year
			"application_logs": 90,   // 3 months
			"metrics":         365,  // 1 year
		},
		EnableEncryptionAtRest:     true,
		EnableEncryptionInTransit:  true,
		RequireApprovalWorkflows:   true,
	}
}

func (deployment *ProductionDeployment) deployProductionApiServices(ctx *pulumi.Context, config *ProductionDeploymentConfiguration) error {
	apis := []string{"identity-api", "content-api", "services-api"}
	
	for _, apiName := range apis {
		if err := deployment.deployProductionApiService(ctx, apiName, config); err != nil {
			return fmt.Errorf("failed to deploy production %s: %w", apiName, err)
		}
	}

	return nil
}

func (deployment *ProductionDeployment) deployProductionApiService(ctx *pulumi.Context, serviceName string, config *ProductionDeploymentConfiguration) error {
	scalingConfig := config.ScalingRules[serviceName]
	
	containerApp, err := app.NewContainerApp(ctx, fmt.Sprintf("production-%s", serviceName), &app.ContainerAppArgs{
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
			},
			Dapr: &app.DaprArgs{
				Enabled:     pulumi.Bool(true),
				AppId:       pulumi.String(serviceName),
				AppPort:     pulumi.Int(8080),
				AppProtocol: pulumi.String("http"),
				EnableApiLogging: pulumi.Bool(true),
				LogLevel:    pulumi.String("error"), // Production log level
			},
			Secrets: app.SecretArray{
				&app.SecretArgs{
					Name:        pulumi.String("database-connection"),
					KeyVaultUrl: deployment.vaultStack.GetVaultUri(),
					Identity:    pulumi.String("system"),
				},
				&app.SecretArgs{
					Name:        pulumi.String("database-read-replica-connection"),
					KeyVaultUrl: deployment.vaultStack.GetVaultUri(),
					Identity:    pulumi.String("system"),
				},
				&app.SecretArgs{
					Name:        pulumi.String("redis-connection"),
					KeyVaultUrl: deployment.vaultStack.GetVaultUri(),
					Identity:    pulumi.String("system"),
				},
				&app.SecretArgs{
					Name:        pulumi.String("encryption-key"),
					KeyVaultUrl: deployment.vaultStack.GetVaultUri(),
					Identity:    pulumi.String("system"),
				},
				&app.SecretArgs{
					Name:        pulumi.String("audit-signing-key"),
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
								"value": pulumi.String(fmt.Sprintf("%d", scalingConfig.TargetCPUUtilization)),
							},
						},
					},
					&app.ScaleRuleArgs{
						Name: pulumi.String("memory-utilization"),
						Custom: &app.CustomScaleRuleArgs{
							Type: pulumi.String("memory"),
							Metadata: pulumi.StringMap{
								"type":  pulumi.String("Utilization"),
								"value": pulumi.String(fmt.Sprintf("%d", scalingConfig.TargetMemoryUtilization)),
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
						Cpu:    pulumi.Float64(1.5), // Higher resources for production
						Memory: pulumi.String("3Gi"),
					},
					Env: deployment.getProductionServiceEnvironmentVariables(serviceName, config),
					Probes: app.ContainerAppProbeArray{
						&app.ContainerAppProbeArgs{
							Type: pulumi.String("Liveness"),
							HttpGet: &app.ContainerAppProbeHttpGetArgs{
								Path:   pulumi.String("/health"),
								Port:   pulumi.Int(8080),
								Scheme: pulumi.String("HTTP"),
							},
							InitialDelaySeconds: pulumi.Int(90),  // Longer startup time for production
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
							InitialDelaySeconds: pulumi.Int(45),
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
							InitialDelaySeconds: pulumi.Int(30),
							PeriodSeconds:      pulumi.Int(10),
							TimeoutSeconds:     pulumi.Int(5),
							FailureThreshold:   pulumi.Int(90), // Very generous startup timeout
						},
					},
				},
			},
		},
		Tags: pulumi.StringMap{
			"environment":                 pulumi.String("production"),
			"service":                    pulumi.String(serviceName),
			"project":                    pulumi.String("international-center"),
			"tier":                       pulumi.String("api"),
			"compliance":                 pulumi.String("required"),
			"backup-required":            pulumi.String("true"),
			"security-scanning-required": pulumi.String("true"),
			"audit-logging-enabled":      pulumi.String("true"),
		},
	})
	if err != nil {
		return err
	}

	deployment.deployedApps[serviceName] = containerApp
	return nil
}

func (deployment *ProductionDeployment) deployProductionGatewayServices(ctx *pulumi.Context, config *ProductionDeploymentConfiguration) error {
	gateways := []string{"public-gateway", "admin-gateway"}
	
	for _, gatewayName := range gateways {
		if err := deployment.deployProductionGatewayService(ctx, gatewayName, config); err != nil {
			return fmt.Errorf("failed to deploy production %s: %w", gatewayName, err)
		}
	}

	return nil
}

func (deployment *ProductionDeployment) deployProductionGatewayService(ctx *pulumi.Context, gatewayName string, config *ProductionDeploymentConfiguration) error {
	isPublic := gatewayName == "public-gateway"
	scalingConfig := config.ScalingRules[gatewayName]
	
	containerApp, err := app.NewContainerApp(ctx, fmt.Sprintf("production-%s", gatewayName), &app.ContainerAppArgs{
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
								Name: pulumi.String("api.international-center.com"),
								CertificateId: pulumi.String(""), // SSL certificate from Key Vault
								BindingType: pulumi.String("SniEnabled"),
							},
						}
					} else {
						return app.CustomDomainArray{
							&app.CustomDomainArgs{
								Name: pulumi.String("admin-api.international-center.com"),
								CertificateId: pulumi.String(""), // SSL certificate from Key Vault
								BindingType: pulumi.String("SniEnabled"),
							},
						}
					}
				}(),
				CorsPolicy: &app.CorsPolicyArgs{
					AllowedOrigins: func() pulumi.StringArray {
						origins := pulumi.StringArray{}
						for _, origin := range config.SecuritySettings.AllowedOrigins {
							if isPublic && (origin == "https://www.international-center.com" || origin == "https://app.international-center.com") {
								origins = append(origins, pulumi.String(origin))
							} else if !isPublic && origin == "https://admin.international-center.com" {
								origins = append(origins, pulumi.String(origin))
							}
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
						pulumi.String("Content-Type"),
						pulumi.String("Authorization"),
						pulumi.String("X-Requested-With"),
						pulumi.String("X-Correlation-ID"),
						pulumi.String("X-Trace-ID"),
						pulumi.String("X-User-ID"),
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
				LogLevel:    pulumi.String("error"), // Production log level
			},
			Secrets: app.SecretArray{
				&app.SecretArgs{
					Name:        pulumi.String("jwt-signing-key"),
					KeyVaultUrl: deployment.vaultStack.GetVaultUri(),
					Identity:    pulumi.String("system"),
				},
				&app.SecretArgs{
					Name:        pulumi.String("api-keys"),
					KeyVaultUrl: deployment.vaultStack.GetVaultUri(),
					Identity:    pulumi.String("system"),
				},
				&app.SecretArgs{
					Name:        pulumi.String("encryption-key"),
					KeyVaultUrl: deployment.vaultStack.GetVaultUri(),
					Identity:    pulumi.String("system"),
				},
				&app.SecretArgs{
					Name:        pulumi.String("audit-signing-key"),
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
								"value": pulumi.String(fmt.Sprintf("%d", scalingConfig.TargetCPUUtilization)),
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
						Cpu:    pulumi.Float64(2.0), // Higher resources for production gateways
						Memory: pulumi.String("4Gi"),
					},
					Env: deployment.getProductionGatewayEnvironmentVariables(gatewayName, config),
					Probes: app.ContainerAppProbeArray{
						&app.ContainerAppProbeArgs{
							Type: pulumi.String("Liveness"),
							HttpGet: &app.ContainerAppProbeHttpGetArgs{
								Path:   pulumi.String("/health"),
								Port:   pulumi.Int(8080),
								Scheme: pulumi.String("HTTP"),
							},
							InitialDelaySeconds: pulumi.Int(120), // Longer startup for production gateways
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
							InitialDelaySeconds: pulumi.Int(60),
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
			"environment":                 pulumi.String("production"),
			"service":                    pulumi.String(gatewayName),
			"project":                    pulumi.String("international-center"),
			"tier":                       pulumi.String("gateway"),
			"compliance":                 pulumi.String("required"),
			"backup-required":            pulumi.String("true"),
			"security-scanning-required": pulumi.String("true"),
			"audit-logging-enabled":      pulumi.String("true"),
			"public-facing":              pulumi.String(fmt.Sprintf("%t", isPublic)),
		},
	})
	if err != nil {
		return err
	}

	deployment.deployedGateways[gatewayName] = containerApp
	return nil
}

func (deployment *ProductionDeployment) getProductionServiceEnvironmentVariables(serviceName string, config *ProductionDeploymentConfiguration) app.EnvironmentVarArray {
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
			Name:      pulumi.String("DATABASE_READ_REPLICA_CONNECTION"),
			SecretRef: pulumi.String("database-read-replica-connection"),
		},
		&app.EnvironmentVarArgs{
			Name:      pulumi.String("REDIS_CONNECTION"),
			SecretRef: pulumi.String("redis-connection"),
		},
		&app.EnvironmentVarArgs{
			Name:      pulumi.String("ENCRYPTION_KEY"),
			SecretRef: pulumi.String("encryption-key"),
		},
		&app.EnvironmentVarArgs{
			Name:      pulumi.String("AUDIT_SIGNING_KEY"),
			SecretRef: pulumi.String("audit-signing-key"),
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
			Name:  pulumi.String("LOG_LEVEL"),
			Value: pulumi.String("INFO"),
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
			Name:  pulumi.String("ENABLE_AUDIT_LOGGING"),
			Value: pulumi.String(fmt.Sprintf("%t", config.EnableAuditLogging)),
		},
		&app.EnvironmentVarArgs{
			Name:  pulumi.String("ENABLE_SECURITY_SCANNING"),
			Value: pulumi.String(fmt.Sprintf("%t", config.EnableSecurityScanning)),
		},
		&app.EnvironmentVarArgs{
			Name:  pulumi.String("ENABLE_COMPLIANCE_MONITORING"),
			Value: pulumi.String(fmt.Sprintf("%t", config.EnableComplianceMonitoring)),
		},
		&app.EnvironmentVarArgs{
			Name:  pulumi.String("PROMETHEUS_ENDPOINT"),
			Value: deployment.grafanaStack.GetPrometheusUrl(),
		},
		&app.EnvironmentVarArgs{
			Name:  pulumi.String("LOKI_ENDPOINT"),
			Value: deployment.grafanaStack.GetLokiUrl(),
		},
		&app.EnvironmentVarArgs{
			Name:  pulumi.String("TEMPO_ENDPOINT"),
			Value: deployment.grafanaStack.GetTempoUrl(),
		},
		&app.EnvironmentVarArgs{
			Name:  pulumi.String("AUDIT_RETENTION_DAYS"),
			Value: pulumi.String(fmt.Sprintf("%d", deployment.complianceConfig.AuditRetentionDays)),
		},
		&app.EnvironmentVarArgs{
			Name:  pulumi.String("ENABLE_ENCRYPTION_AT_REST"),
			Value: pulumi.String(fmt.Sprintf("%t", deployment.complianceConfig.EnableEncryptionAtRest)),
		},
		&app.EnvironmentVarArgs{
			Name:  pulumi.String("ENABLE_ENCRYPTION_IN_TRANSIT"),
			Value: pulumi.String(fmt.Sprintf("%t", deployment.complianceConfig.EnableEncryptionInTransit)),
		},
	}

	return baseVars
}

func (deployment *ProductionDeployment) getProductionGatewayEnvironmentVariables(gatewayName string, config *ProductionDeploymentConfiguration) app.EnvironmentVarArray {
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
			Name:      pulumi.String("JWT_SIGNING_KEY"),
			SecretRef: pulumi.String("jwt-signing-key"),
		},
		&app.EnvironmentVarArgs{
			Name:      pulumi.String("API_KEYS"),
			SecretRef: pulumi.String("api-keys"),
		},
		&app.EnvironmentVarArgs{
			Name:      pulumi.String("ENCRYPTION_KEY"),
			SecretRef: pulumi.String("encryption-key"),
		},
		&app.EnvironmentVarArgs{
			Name:      pulumi.String("AUDIT_SIGNING_KEY"),
			SecretRef: pulumi.String("audit-signing-key"),
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
			Name:  pulumi.String("ENFORCE_HTTPS"),
			Value: pulumi.String(fmt.Sprintf("%t", config.SecuritySettings.EnforceHTTPS)),
		},
		&app.EnvironmentVarArgs{
			Name:  pulumi.String("REQUIRE_AUTHENTICATION"),
			Value: pulumi.String(fmt.Sprintf("%t", config.SecuritySettings.RequireAuthentication)),
		},
		&app.EnvironmentVarArgs{
			Name:  pulumi.String("ENABLE_CSRF"),
			Value: pulumi.String(fmt.Sprintf("%t", config.SecuritySettings.EnableCSRF)),
		},
		&app.EnvironmentVarArgs{
			Name:  pulumi.String("ENABLE_AUDIT_LOGGING"),
			Value: pulumi.String(fmt.Sprintf("%t", config.EnableAuditLogging)),
		},
		&app.EnvironmentVarArgs{
			Name:  pulumi.String("ENABLE_ADVANCED_THREAT_PROTECTION"),
			Value: pulumi.String(fmt.Sprintf("%t", config.SecuritySettings.EnableAdvancedThreatProtection)),
		},
		&app.EnvironmentVarArgs{
			Name:  pulumi.String("ENABLE_SECURITY_HEADERS"),
			Value: pulumi.String(fmt.Sprintf("%t", config.SecuritySettings.EnableSecurityHeaders)),
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
		&app.EnvironmentVarArgs{
			Name:  pulumi.String("VULNERABILITY_THRESHOLD"),
			Value: pulumi.String(config.SecuritySettings.VulnerabilityThreshold),
		},
		&app.EnvironmentVarArgs{
			Name:  pulumi.String("SECURITY_SCAN_SCHEDULE"),
			Value: pulumi.String(config.SecuritySettings.SecurityScanSchedule),
		},
	}

	return baseVars
}

func (deployment *ProductionDeployment) configureProductionMonitoring(ctx *pulumi.Context, config *ProductionDeploymentConfiguration) error {
	return nil
}

func (deployment *ProductionDeployment) enableSecurityCompliance(ctx *pulumi.Context, config *ProductionDeploymentConfiguration) error {
	return nil
}

func (deployment *ProductionDeployment) GetDeployedApp(name string) *app.ContainerApp {
	return deployment.deployedApps[name]
}

func (deployment *ProductionDeployment) GetDeployedGateway(name string) *app.ContainerApp {
	return deployment.deployedGateways[name]
}

func (deployment *ProductionDeployment) GetSecurityConfiguration() *ProductionSecurityConfiguration {
	return deployment.securityConfig
}

func (deployment *ProductionDeployment) GetComplianceConfiguration() *ProductionComplianceConfiguration {
	return deployment.complianceConfig
}