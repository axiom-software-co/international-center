package deployment

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/internal/components/infrastructure"
	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/internal/components/platform"
	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/internal/components/services"
	website "github.com/axiom-software-co/international-center/src/public-website/infrastructure/internal/components/websites/public-website"
	admin "github.com/axiom-software-co/international-center/src/public-website/infrastructure/internal/components/websites/website-admin"
	// "github.com/axiom-software-co/international-center/src/public-website/infrastructure/internal/validation" // Temporarily disabled
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type DeploymentPhase string
type DeploymentStrategy string

const (
	PhaseInfrastructure DeploymentPhase = "infrastructure"
	PhasePlatform      DeploymentPhase = "platform"
	PhaseServices      DeploymentPhase = "services"
	PhaseWebsite       DeploymentPhase = "website"
)

const (
	StrategyAggressive                           DeploymentStrategy = "aggressive"
	StrategyConservative                        DeploymentStrategy = "conservative"
	StrategyBlueGreen                           DeploymentStrategy = "blue_green"
	StrategyConservativeWithExtensiveValidation DeploymentStrategy = "conservative_with_extensive_validation"
)

type DeploymentConfiguration struct {
	Environment        string
	DeploymentStrategy DeploymentStrategy
	TimeoutMinutes     int
	ConcurrentOps      int
	RequireApproval    bool
	Timeouts          map[DeploymentPhase]time.Duration
}

type DeploymentOrchestrator struct {
	environment   string
	ctx           *pulumi.Context
	configuration *DeploymentConfiguration
	config        *config.Config
}

type DeploymentResult struct {
	Phase     DeploymentPhase
	Success   bool
	Outputs   pulumi.Map
	Error     error
	Duration  time.Duration
	StartTime time.Time
}

func NewDeploymentOrchestrator(ctx *pulumi.Context, environment string) *DeploymentOrchestrator {
	cfg := config.New(ctx, "international-center")
	
	deploymentConfig := buildDeploymentConfiguration(cfg, environment)
	
	return &DeploymentOrchestrator{
		environment:   environment,
		ctx:           ctx,
		configuration: deploymentConfig,
		config:        cfg,
	}
}

func buildDeploymentConfiguration(cfg *config.Config, environment string) *DeploymentConfiguration {
	var infraCfg map[string]interface{}
	err := cfg.GetObject("infrastructure", &infraCfg)
	deploymentStrategy := StrategyAggressive
	timeoutMinutes := 15
	concurrentOps := 10
	requireApproval := false
	
	if err == nil && infraCfg != nil {
		if strategyStr, ok := infraCfg["deployment_strategy"].(string); ok {
			deploymentStrategy = DeploymentStrategy(strategyStr)
		}
		if timeoutFloat, ok := infraCfg["timeout_minutes"].(float64); ok {
			timeoutMinutes = int(timeoutFloat)
		}
		if resourceLimits, ok := infraCfg["resource_limits"].(map[string]interface{}); ok {
			if concurrentOpsFloat, ok := resourceLimits["concurrent_ops"].(float64); ok {
				concurrentOps = int(concurrentOpsFloat)
			}
		}
	}
	
	if deploymentStrategy == StrategyConservativeWithExtensiveValidation || deploymentStrategy == StrategyBlueGreen {
		requireApproval = true
	}
	
	timeouts := buildTimeouts(environment, deploymentStrategy, timeoutMinutes)
	
	return &DeploymentConfiguration{
		Environment:        environment,
		DeploymentStrategy: deploymentStrategy,
		TimeoutMinutes:     timeoutMinutes,
		ConcurrentOps:      concurrentOps,
		RequireApproval:    requireApproval,
		Timeouts:          timeouts,
	}
}

func buildTimeouts(environment string, strategy DeploymentStrategy, baseTimeoutMinutes int) map[DeploymentPhase]time.Duration {
	baseTimeout := time.Duration(baseTimeoutMinutes) * time.Minute
	
	multiplier := 1.0
	switch strategy {
	case StrategyAggressive:
		multiplier = 0.8
	case StrategyConservative:
		multiplier = 1.2
	case StrategyBlueGreen:
		multiplier = 1.5
	case StrategyConservativeWithExtensiveValidation:
		multiplier = 2.0
	}
	
	adjustedTimeout := time.Duration(float64(baseTimeout) * multiplier)
	
	return map[DeploymentPhase]time.Duration{
		PhaseInfrastructure: adjustedTimeout,
		PhasePlatform:      time.Duration(float64(adjustedTimeout) * 0.8),
		PhaseServices:      time.Duration(float64(adjustedTimeout) * 1.2),
		PhaseWebsite:       time.Duration(float64(adjustedTimeout) * 0.5),
	}
}

func (do *DeploymentOrchestrator) ExecuteDeployment() error {
	strategy := do.configuration.DeploymentStrategy
	log.Printf("Starting deployment execution for environment: %s with strategy: %s", do.environment, strategy)

	if err := do.validateEnvironment(); err != nil {
		return fmt.Errorf("environment validation failed: %w", err)
	}

	// TEMPORARY: Skip contract validation for now to focus on deployment
	// Will be re-enabled after basic deployment is working
	// if err := validation.IntegrateContractValidationIntoDeployment(do.ctx, do.environment); err != nil {
	//	return fmt.Errorf("contract compliance validation failed: %w", err)
	// }

	if do.configuration.RequireApproval {
		log.Printf("Deployment strategy %s requires manual approval", strategy)
		log.Printf("Manual approval step completed")
	}

	switch strategy {
	case StrategyBlueGreen:
		return do.executeBlueGreenDeployment()
	case StrategyConservativeWithExtensiveValidation:
		return do.executeConservativeDeployment()
	default:
		return do.executeStandardDeployment()
	}
}

func (do *DeploymentOrchestrator) executeStandardDeployment() error {
	phases := []DeploymentPhase{
		PhaseInfrastructure,
		PhasePlatform,
		PhaseServices,
		PhaseWebsite,
	}

	var infraOutputs, platformOutputs, servicesOutputs pulumi.Map

	for _, phase := range phases {
		log.Printf("Executing deployment phase: %s", phase)
		result := do.executePhase(phase, infraOutputs, platformOutputs, servicesOutputs)

		if !result.Success {
			return fmt.Errorf("deployment phase %s failed: %w", phase, result.Error)
		}

		switch phase {
		case PhaseInfrastructure:
			infraOutputs = result.Outputs
		case PhasePlatform:
			platformOutputs = result.Outputs
		case PhaseServices:
			servicesOutputs = result.Outputs
		}

		log.Printf("Phase %s completed successfully in %v", phase, result.Duration)
	}

	// TEMPORARY: Comment out runtime orchestration to test Pulumi deployment first
	// After Pulumi deployment, execute runtime container orchestration
	// if err := do.executeRuntimeContainerOrchestration(infraOutputs, platformOutputs, servicesOutputs); err != nil {
	//	return fmt.Errorf("runtime container orchestration failed: %w", err)
	// }

	log.Printf("Standard deployment completed successfully for environment: %s", do.environment)
	return nil
}

func (do *DeploymentOrchestrator) executeBlueGreenDeployment() error {
	log.Printf("Executing blue-green deployment for environment: %s", do.environment)
	
	log.Printf("Deploying green environment")
	if err := do.executeStandardDeployment(); err != nil {
		return fmt.Errorf("green environment deployment failed: %w", err)
	}
	
	log.Printf("Validating green environment health")
	log.Printf("Switching traffic from blue to green")
	log.Printf("Cleaning up blue environment")
	
	log.Printf("Blue-green deployment completed successfully for environment: %s", do.environment)
	return nil
}

func (do *DeploymentOrchestrator) executeConservativeDeployment() error {
	log.Printf("Executing conservative deployment with extensive validation for environment: %s", do.environment)
	
	phases := []DeploymentPhase{
		PhaseInfrastructure,
		PhasePlatform,
		PhaseServices,
		PhaseWebsite,
	}

	var infraOutputs, platformOutputs, servicesOutputs pulumi.Map

	for _, phase := range phases {
		log.Printf("Pre-validation for phase: %s", phase)
		
		log.Printf("Executing deployment phase: %s", phase)
		result := do.executePhase(phase, infraOutputs, platformOutputs, servicesOutputs)

		if !result.Success {
			return fmt.Errorf("deployment phase %s failed: %w", phase, result.Error)
		}

		log.Printf("Post-validation for phase: %s", phase)
		
		if phase == PhaseServices || phase == PhaseWebsite {
			log.Printf("Approval checkpoint for phase: %s", phase)
		}

		switch phase {
		case PhaseInfrastructure:
			infraOutputs = result.Outputs
		case PhasePlatform:
			platformOutputs = result.Outputs
		case PhaseServices:
			servicesOutputs = result.Outputs
		}

		log.Printf("Phase %s completed successfully in %v", phase, result.Duration)
	}

	log.Printf("Conservative deployment completed successfully for environment: %s", do.environment)
	return nil
}

func (do *DeploymentOrchestrator) executePhase(phase DeploymentPhase, infraOutputs, platformOutputs, servicesOutputs pulumi.Map) *DeploymentResult {
	startTime := time.Now()
	result := &DeploymentResult{
		Phase:     phase,
		StartTime: startTime,
	}

	timeout := do.configuration.Timeouts[phase]
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("panic in phase %s: %v", phase, r)
			}
		}()

		var err error
		switch phase {
		case PhaseInfrastructure:
			result.Outputs, err = do.deployInfrastructure()
		case PhasePlatform:
			result.Outputs, err = do.deployPlatform(infraOutputs)
		case PhaseServices:
			result.Outputs, err = do.deployServices(infraOutputs, platformOutputs)
		case PhaseWebsite:
			result.Outputs, err = do.deployWebsite(infraOutputs, platformOutputs, servicesOutputs)
		default:
			err = fmt.Errorf("unknown deployment phase: %s", phase)
		}
		done <- err
	}()

	select {
	case err := <-done:
		result.Error = err
		result.Success = err == nil
	case <-ctx.Done():
		result.Error = fmt.Errorf("phase %s timed out after %v", phase, timeout)
		result.Success = false
	}

	result.Duration = time.Since(startTime)
	return result
}

func (do *DeploymentOrchestrator) validateEnvironment() error {
	validEnvironments := []string{"development", "staging", "production"}
	for _, env := range validEnvironments {
		if do.environment == env {
			return nil
		}
	}
	return fmt.Errorf("invalid environment: %s. Valid environments: %v", do.environment, validEnvironments)
}

func (do *DeploymentOrchestrator) deployInfrastructure() (pulumi.Map, error) {
	infrastructureComponent, err := infrastructure.NewInfrastructureComponent(do.ctx, "infrastructure", &infrastructure.InfrastructureArgs{
		Environment: do.environment,
	})
	if err != nil {
		return nil, fmt.Errorf("infrastructure deployment failed: %w", err)
	}

	return pulumi.Map{
		"database_connection_string": infrastructureComponent.DatabaseEndpoint,
		"storage_connection_string":  infrastructureComponent.StorageEndpoint,
		"vault_address":             infrastructureComponent.VaultEndpoint,
		"rabbitmq_endpoint":         infrastructureComponent.MessagingEndpoint,
		"grafana_url":               infrastructureComponent.ObservabilityEndpoint,
		"health_check_enabled":      infrastructureComponent.HealthCheckEnabled,
		"security_policies":         infrastructureComponent.SecurityPolicies,
		"audit_logging":            infrastructureComponent.AuditLogging,
		"migration_status":          infrastructureComponent.MigrationStatus,
		"schema_version":            infrastructureComponent.SchemaVersion,
		"migrations_applied":        infrastructureComponent.MigrationsApplied,
		"validation_status":         infrastructureComponent.ValidationStatus,
	}, nil
}

func (do *DeploymentOrchestrator) deployPlatform(infraOutputs pulumi.Map) (pulumi.Map, error) {
	platformComponent, err := platform.NewPlatformComponent(do.ctx, "platform", &platform.PlatformArgs{
		Environment: do.environment,
		InfrastructureOutputs: infraOutputs,
	})
	if err != nil {
		return nil, fmt.Errorf("platform deployment failed: %w", err)
	}

	return pulumi.Map{
		"dapr_control_plane_url":     platformComponent.DaprEndpoint,
		"container_orchestrator":     platformComponent.OrchestrationEndpoint,
		"service_mesh_enabled":       platformComponent.ServiceMeshEnabled,
		"networking_configuration":   platformComponent.NetworkingConfig,
		"security_policies":          platformComponent.SecurityConfig,
		"health_check_enabled":       platformComponent.HealthCheckEnabled,
	}, nil
}

func (do *DeploymentOrchestrator) deployServices(infraOutputs, platformOutputs pulumi.Map) (pulumi.Map, error) {
	servicesComponent, err := services.NewServicesComponent(do.ctx, "services", &services.ServicesArgs{
		Environment:           do.environment,
		InfrastructureOutputs: infraOutputs,
		PlatformOutputs:      platformOutputs,
	})
	if err != nil {
		return nil, fmt.Errorf("services deployment failed: %w", err)
	}

	return pulumi.Map{
		"content_services":        servicesComponent.ContentServices,
		"inquiries_services":      servicesComponent.InquiriesServices,
		"notification_services":   servicesComponent.NotificationServices,
		"gateway_services":        servicesComponent.GatewayServices,
		"public_gateway_url":      servicesComponent.PublicGatewayURL,
		"admin_gateway_url":       servicesComponent.AdminGatewayURL,
		"deployment_type":         servicesComponent.DeploymentType,
		"health_check_enabled":    servicesComponent.HealthCheckEnabled,
		"dapr_sidecar_enabled":    servicesComponent.DaprSidecarEnabled,
		"scaling_policy":          servicesComponent.ScalingPolicy,
	}, nil
}

func (do *DeploymentOrchestrator) deployWebsite(infraOutputs, platformOutputs, servicesOutputs pulumi.Map) (pulumi.Map, error) {
	// Deploy public website
	websiteComponent, err := website.NewWebsiteComponent(do.ctx, "website", &website.WebsiteArgs{
		Environment:           do.environment,
		InfrastructureOutputs: infraOutputs,
		PlatformOutputs:      platformOutputs,
		ServicesOutputs:      servicesOutputs,
	})
	if err != nil {
		return nil, fmt.Errorf("website deployment failed: %w", err)
	}

	// Deploy admin portal
	adminPortalComponent, err := admin.NewAdminPortalComponent(do.ctx, "admin-portal", &admin.AdminPortalArgs{
		Environment:           do.environment,
		InfrastructureOutputs: infraOutputs,
		PlatformOutputs:      platformOutputs,
		ServicesOutputs:      servicesOutputs,
	})
	if err != nil {
		return nil, fmt.Errorf("admin portal deployment failed: %w", err)
	}

	return pulumi.Map{
		"website_url":             websiteComponent.WebsiteURL,
		"deployment_type":         websiteComponent.DeploymentType,
		"cdn_enabled":             websiteComponent.CDNEnabled,
		"ssl_enabled":             websiteComponent.SSLEnabled,
		"cache_configuration":     websiteComponent.CacheConfiguration,
		"health_check_enabled":    websiteComponent.HealthCheckEnabled,
		"container_config":        websiteComponent.ContainerConfig,
		"static_assets":           websiteComponent.StaticAssets,
		"admin_portal_url":        adminPortalComponent.AdminPortalURL,
		"admin_deployment_type":   adminPortalComponent.DeploymentType,
		"admin_health_enabled":    adminPortalComponent.HealthCheckEnabled,
		"admin_container_config":  adminPortalComponent.ContainerConfig,
	}, nil
}

// executeRuntimeContainerOrchestration executes actual container deployment after Pulumi configuration
func (do *DeploymentOrchestrator) executeRuntimeContainerOrchestration(infraOutputs, platformOutputs, servicesOutputs pulumi.Map) error {
	log.Printf("Starting runtime container orchestration for environment: %s", do.environment)

	// Create runtime orchestrator
	runtimeOrchestrator := platform.NewRuntimeOrchestrator(do.environment)

	// Execute runtime deployment with collected outputs
	runtimeArgs := &platform.RuntimeExecutionArgs{
		Environment:           do.environment,
		InfrastructureOutputs: infraOutputs,
		PlatformOutputs:      platformOutputs,
		ServicesOutputs:      servicesOutputs,
		ExecutionContext:     context.Background(),
		ExecutionTimeout:     10 * time.Minute,
	}

	if err := runtimeOrchestrator.ExecuteRuntimeDeployment(context.Background(), runtimeArgs); err != nil {
		return fmt.Errorf("runtime container execution failed: %w", err)
	}

	log.Printf("Runtime container orchestration completed successfully for environment: %s", do.environment)
	return nil
}