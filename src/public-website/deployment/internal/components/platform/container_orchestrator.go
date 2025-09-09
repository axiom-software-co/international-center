package platform

import (
	"context"
	"fmt"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ContainerOrchestratorArgs struct {
	Environment           string
	InfrastructureOutputs pulumi.Map
	PlatformOutputs      pulumi.Map
}

type ContainerOrchestratorComponent struct {
	pulumi.ResourceState

	OrchestrationEngine   pulumi.StringOutput `pulumi:"orchestrationEngine"`
	DeploymentStrategy    pulumi.StringOutput `pulumi:"deploymentStrategy"`
	HealthCheckConfig     pulumi.MapOutput    `pulumi:"healthCheckConfig"`
	DependencyGraph       pulumi.MapOutput    `pulumi:"dependencyGraph"`
	ContainerRegistry     pulumi.MapOutput    `pulumi:"containerRegistry"`
	NetworkConfiguration  pulumi.MapOutput    `pulumi:"networkConfiguration"`
	PodmanProvider        *PodmanProviderComponent
	AzureProvider         *AzureContainerProviderComponent
}

type ContainerDeploymentOrder struct {
	Phase         string   `json:"phase"`
	Containers    []string `json:"containers"`
	Dependencies  []string `json:"dependencies"`
	HealthChecks  []string `json:"health_checks"`
	TimeoutMinutes int     `json:"timeout_minutes"`
}

func NewContainerOrchestratorComponent(ctx *pulumi.Context, name string, args *ContainerOrchestratorArgs, opts ...pulumi.ResourceOption) (*ContainerOrchestratorComponent, error) {
	component := &ContainerOrchestratorComponent{}
	
	// Safe registration for mock contexts
	if canRegister(ctx) {
		err := ctx.RegisterComponentResource("international-center:platform:ContainerOrchestrator", name, component, opts...)
		if err != nil {
			return nil, err
		}
	}

	var orchestrationEngine, deploymentStrategy pulumi.StringOutput
	var healthCheckConfig, dependencyGraph, containerRegistry, networkConfiguration pulumi.MapOutput

	switch args.Environment {
	case "development":
		orchestrationEngine = pulumi.String("podman").ToStringOutput()
		deploymentStrategy = pulumi.String("sequential").ToStringOutput()
		
		healthCheckConfig = pulumi.Map{
			"enabled":            pulumi.Bool(true),
			"check_interval":     pulumi.String("30s"),
			"timeout":            pulumi.String("10s"),
			"retries":            pulumi.Int(3),
			"start_period":       pulumi.String("60s"),
			"failure_threshold":  pulumi.Int(3),
			"success_threshold":  pulumi.Int(1),
		}.ToMapOutput()

		dependencyGraph = pulumi.Map{
			"infrastructure": pulumi.Array{},
			"platform": pulumi.Array{
				pulumi.String("infrastructure"),
			},
			"dapr": pulumi.Array{
				pulumi.String("infrastructure"),
				pulumi.String("platform"),
			},
			"services": pulumi.Array{
				pulumi.String("infrastructure"),
				pulumi.String("platform"),
				pulumi.String("dapr"),
			},
			"websites": pulumi.Array{
				pulumi.String("infrastructure"),
				pulumi.String("platform"),
				pulumi.String("dapr"),
				pulumi.String("services"),
			},
		}.ToMapOutput()

		containerRegistry = pulumi.Map{
			"type":      pulumi.String("local"),
			"namespace": pulumi.String("localhost"),
			"images": pulumi.Map{
				"dapr":            pulumi.String("daprio/dapr:latest"),
				"public_gateway":  pulumi.String("localhost/backend/public-gateway:latest"),
				"admin_gateway":   pulumi.String("localhost/backend/admin-gateway:latest"),
				"content_service": pulumi.String("localhost/backend/content:latest"),
				"directus":        pulumi.String("directus/directus:latest"),
			},
		}.ToMapOutput()

		networkConfiguration = pulumi.Map{
			"driver":      pulumi.String("bridge"),
			"subnet":      pulumi.String("172.20.0.0/16"),
			"gateway":     pulumi.String("172.20.0.1"),
			"dns_servers": pulumi.Array{
				pulumi.String("8.8.8.8"),
				pulumi.String("8.8.4.4"),
			},
			"port_ranges": pulumi.Map{
				"dapr":         pulumi.String("50000-50100"),
				"gateways":     pulumi.String("9000-9100"),
				"services":     pulumi.String("3000-3100"),
				"admin_portal": pulumi.String("8055"),
			},
		}.ToMapOutput()

	case "staging":
		orchestrationEngine = pulumi.String("azure_container_apps").ToStringOutput()
		deploymentStrategy = pulumi.String("blue_green").ToStringOutput()
		
		healthCheckConfig = pulumi.Map{
			"enabled":            pulumi.Bool(true),
			"check_interval":     pulumi.String("15s"),
			"timeout":            pulumi.String("5s"),
			"retries":            pulumi.Int(5),
			"start_period":       pulumi.String("90s"),
			"failure_threshold":  pulumi.Int(3),
			"success_threshold":  pulumi.Int(2),
		}.ToMapOutput()

		dependencyGraph = pulumi.Map{
			"infrastructure": pulumi.Array{},
			"platform": pulumi.Array{
				pulumi.String("infrastructure"),
			},
			"dapr": pulumi.Array{
				pulumi.String("infrastructure"),
				pulumi.String("platform"),
			},
			"services": pulumi.Array{
				pulumi.String("infrastructure"),
				pulumi.String("platform"),
				pulumi.String("dapr"),
			},
			"websites": pulumi.Array{
				pulumi.String("infrastructure"),
				pulumi.String("platform"),
				pulumi.String("dapr"),
				pulumi.String("services"),
			},
		}.ToMapOutput()

		containerRegistry = pulumi.Map{
			"type":      pulumi.String("azure_container_registry"),
			"namespace": pulumi.String("registry.azurecr.io"),
			"images": pulumi.Map{
				"dapr":            pulumi.String("mcr.microsoft.com/dapr/dapr:latest"),
				"public_gateway":  pulumi.String("registry.azurecr.io/backend/public-gateway:staging"),
				"admin_gateway":   pulumi.String("registry.azurecr.io/backend/admin-gateway:staging"),
				"content_service": pulumi.String("registry.azurecr.io/backend/content:staging"),
				"directus":        pulumi.String("registry.azurecr.io/admin/directus:staging"),
			},
		}.ToMapOutput()

		networkConfiguration = pulumi.Map{
			"driver":          pulumi.String("azure_virtual_network"),
			"vnet_integration": pulumi.Bool(true),
			"subnet_delegation": pulumi.String("Microsoft.App/environments"),
			"dns_configuration": pulumi.Map{
				"private_dns_zone": pulumi.String("staging.internal"),
				"custom_dns":       pulumi.Bool(true),
			},
			"security_groups": pulumi.Array{
				pulumi.String("allow-http-https"),
				pulumi.String("allow-dapr-ports"),
			},
		}.ToMapOutput()

	case "production":
		orchestrationEngine = pulumi.String("azure_container_apps").ToStringOutput()
		deploymentStrategy = pulumi.String("canary").ToStringOutput()
		
		healthCheckConfig = pulumi.Map{
			"enabled":            pulumi.Bool(true),
			"check_interval":     pulumi.String("10s"),
			"timeout":            pulumi.String("3s"),
			"retries":            pulumi.Int(10),
			"start_period":       pulumi.String("120s"),
			"failure_threshold":  pulumi.Int(5),
			"success_threshold":  pulumi.Int(3),
		}.ToMapOutput()

		dependencyGraph = pulumi.Map{
			"infrastructure": pulumi.Array{},
			"platform": pulumi.Array{
				pulumi.String("infrastructure"),
			},
			"dapr": pulumi.Array{
				pulumi.String("infrastructure"),
				pulumi.String("platform"),
			},
			"services": pulumi.Array{
				pulumi.String("infrastructure"),
				pulumi.String("platform"),
				pulumi.String("dapr"),
			},
			"websites": pulumi.Array{
				pulumi.String("infrastructure"),
				pulumi.String("platform"),
				pulumi.String("dapr"),
				pulumi.String("services"),
			},
		}.ToMapOutput()

		containerRegistry = pulumi.Map{
			"type":      pulumi.String("azure_container_registry"),
			"namespace": pulumi.String("registry.azurecr.io"),
			"images": pulumi.Map{
				"dapr":            pulumi.String("mcr.microsoft.com/dapr/dapr:1.12.0"), // Fixed version for production
				"public_gateway":  pulumi.String("registry.azurecr.io/backend/public-gateway:production"),
				"admin_gateway":   pulumi.String("registry.azurecr.io/backend/admin-gateway:production"),
				"content_service": pulumi.String("registry.azurecr.io/backend/content:production"),
				"directus":        pulumi.String("registry.azurecr.io/admin/directus:production"),
			},
		}.ToMapOutput()

		networkConfiguration = pulumi.Map{
			"driver":          pulumi.String("azure_virtual_network"),
			"vnet_integration": pulumi.Bool(true),
			"subnet_delegation": pulumi.String("Microsoft.App/environments"),
			"dns_configuration": pulumi.Map{
				"private_dns_zone": pulumi.String("production.internal"),
				"custom_dns":       pulumi.Bool(true),
			},
			"security_groups": pulumi.Array{
				pulumi.String("allow-http-https-secure"),
				pulumi.String("allow-dapr-ports-secure"),
				pulumi.String("deny-all-other"),
			},
			"traffic_splitting": pulumi.Map{
				"enabled":         pulumi.Bool(true),
				"canary_weight":   pulumi.Int(10),
				"stable_weight":   pulumi.Int(90),
			},
		}.ToMapOutput()

	default:
		return nil, fmt.Errorf("unsupported environment: %s", args.Environment)
	}

	component.OrchestrationEngine = orchestrationEngine
	component.DeploymentStrategy = deploymentStrategy
	component.HealthCheckConfig = healthCheckConfig
	component.DependencyGraph = dependencyGraph
	component.ContainerRegistry = containerRegistry
	component.NetworkConfiguration = networkConfiguration

	// Initialize container provider based on environment
	switch args.Environment {
	case "development":
		podmanProvider, err := NewPodmanProviderComponent(ctx, "podman-provider", &PodmanProviderArgs{
			Environment:          args.Environment,
			NetworkConfiguration: networkConfiguration,
			ContainerRegistry:    containerRegistry,
		}, pulumi.Parent(component))
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Podman provider: %w", err)
		}
		component.PodmanProvider = podmanProvider
		component.AzureProvider = nil
	case "staging":
		azureProvider, err := NewAzureContainerProviderComponent(ctx, "azure-provider", &AzureContainerProviderArgs{
			Environment:          args.Environment,
			NetworkConfiguration: networkConfiguration,
			ContainerRegistry:    containerRegistry,
			ResourceGroupName:    "international-center-staging-rg",
			ContainerEnvironment: "international-center-staging-env",
		}, pulumi.Parent(component))
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Azure provider for staging: %w", err)
		}
		component.AzureProvider = azureProvider
		component.PodmanProvider = nil
	case "production":
		azureProvider, err := NewAzureContainerProviderComponent(ctx, "azure-provider", &AzureContainerProviderArgs{
			Environment:          args.Environment,
			NetworkConfiguration: networkConfiguration,
			ContainerRegistry:    containerRegistry,
			ResourceGroupName:    "international-center-production-rg",
			ContainerEnvironment: "international-center-production-env",
		}, pulumi.Parent(component))
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Azure provider for production: %w", err)
		}
		component.AzureProvider = azureProvider
		component.PodmanProvider = nil
	default:
		component.PodmanProvider = nil
		component.AzureProvider = nil
	}

	if canRegister(ctx) {
		outputMap := pulumi.Map{
			"orchestrationEngine":   component.OrchestrationEngine,
			"deploymentStrategy":    component.DeploymentStrategy,
			"healthCheckConfig":     component.HealthCheckConfig,
			"dependencyGraph":       component.DependencyGraph,
			"containerRegistry":     component.ContainerRegistry,
			"networkConfiguration":  component.NetworkConfiguration,
		}
		
		if err := ctx.RegisterResourceOutputs(component, outputMap); err != nil {
			return nil, err
		}
	}

	return component, nil
}

// ExecuteContainerDeployment orchestrates the container deployment sequence
func (component *ContainerOrchestratorComponent) ExecuteContainerDeployment(ctx context.Context, phase string) error {
	deploymentOrder := component.getDeploymentOrder(phase)
	
	// Validate dependencies are met
	if err := component.validateDependencies(ctx, deploymentOrder.Dependencies); err != nil {
		return fmt.Errorf("dependency validation failed for phase %s: %w", phase, err)
	}

	// Execute container deployments based on orchestration engine
	engineType := extractStringFromPulumiOutput(component.OrchestrationEngine)
	
	switch engineType {
	case "podman":
		return component.deployWithPodman(ctx, deploymentOrder)
	case "azure_container_apps":
		return component.deployWithAzureContainerApps(ctx, deploymentOrder)
	default:
		return fmt.Errorf("unsupported orchestration engine: %s", engineType)
	}
}

// getDeploymentOrder returns the deployment order for a specific phase
func (component *ContainerOrchestratorComponent) getDeploymentOrder(phase string) *ContainerDeploymentOrder {
	switch phase {
	case "infrastructure":
		return &ContainerDeploymentOrder{
			Phase:          "infrastructure",
			Containers:     []string{"postgres", "redis", "vault", "rabbitmq", "grafana"},
			Dependencies:   []string{},
			HealthChecks:   []string{"database", "cache", "secrets", "messaging", "monitoring"},
			TimeoutMinutes: 10,
		}
	case "platform":
		return &ContainerDeploymentOrder{
			Phase:          "platform",
			Containers:     []string{"dapr-control-plane", "dapr-placement", "dapr-sentry"},
			Dependencies:   []string{"infrastructure"},
			HealthChecks:   []string{"dapr-control-plane", "dapr-placement", "dapr-sentry"},
			TimeoutMinutes: 5,
		}
	case "services":
		return &ContainerDeploymentOrder{
			Phase:          "services",
			Containers:     []string{"public-gateway", "admin-gateway", "content-news", "content-events", "content-research", "inquiries-business", "inquiries-donations", "inquiries-media", "inquiries-volunteers", "notification-service"},
			Dependencies:   []string{"infrastructure", "platform"},
			HealthChecks:   []string{"gateway-health", "content-health", "inquiries-health", "notifications-health"},
			TimeoutMinutes: 15,
		}
	case "websites":
		return &ContainerDeploymentOrder{
			Phase:          "websites",
			Containers:     []string{"admin-portal"},
			Dependencies:   []string{"infrastructure", "platform", "services"},
			HealthChecks:   []string{"admin-portal-health"},
			TimeoutMinutes: 5,
		}
	default:
		return &ContainerDeploymentOrder{
			Phase:          "unknown",
			Containers:     []string{},
			Dependencies:   []string{},
			HealthChecks:   []string{},
			TimeoutMinutes: 5,
		}
	}
}

// validateDependencies validates that all required dependencies are healthy
func (component *ContainerOrchestratorComponent) validateDependencies(ctx context.Context, dependencies []string) error {
	for _, dependency := range dependencies {
		if err := component.validateDependencyHealth(ctx, dependency); err != nil {
			return fmt.Errorf("dependency %s is not healthy: %w", dependency, err)
		}
	}
	return nil
}

// validateDependencyHealth validates that a specific dependency is healthy
func (component *ContainerOrchestratorComponent) validateDependencyHealth(ctx context.Context, dependency string) error {
	// This would implement actual dependency health validation
	// For now, this is a placeholder
	switch dependency {
	case "infrastructure":
		return component.validateInfrastructureHealth(ctx)
	case "platform":
		return component.validatePlatformHealth(ctx)
	case "dapr":
		return component.validateDaprHealth(ctx)
	case "services":
		return component.validateServicesHealth(ctx)
	default:
		return fmt.Errorf("unknown dependency: %s", dependency)
	}
}

// deployWithPodman deploys containers using Podman for development
func (component *ContainerOrchestratorComponent) deployWithPodman(ctx context.Context, order *ContainerDeploymentOrder) error {
	// TODO: Implement Podman deployment orchestration
	// This should:
	// 1. Create container network if needed
	// 2. Pull required images
	// 3. Start containers in correct order
	// 4. Wait for each container to be healthy before proceeding
	// 5. Configure Dapr sidecars
	// 6. Validate all containers are running and healthy

	for _, containerID := range order.Containers {
		if err := component.deployPodmanContainer(ctx, containerID); err != nil {
			return fmt.Errorf("failed to deploy container %s: %w", containerID, err)
		}
		
		// Wait for container health check
		if err := component.waitForContainerHealth(ctx, containerID, 60*time.Second); err != nil {
			return fmt.Errorf("container %s failed health check: %w", containerID, err)
		}
	}

	return nil
}

// deployWithAzureContainerApps deploys containers using Azure Container Apps for staging/production
func (component *ContainerOrchestratorComponent) deployWithAzureContainerApps(ctx context.Context, order *ContainerDeploymentOrder) error {
	// TODO: Implement Azure Container Apps deployment orchestration
	// This should:
	// 1. Create Container App Environment if needed
	// 2. Configure Dapr components
	// 3. Deploy Container Apps with proper scaling rules
	// 4. Configure ingress and networking
	// 5. Set up monitoring and alerting
	// 6. Validate all apps are running and healthy

	for _, containerID := range order.Containers {
		if err := component.deployAzureContainerApp(ctx, containerID); err != nil {
			return fmt.Errorf("failed to deploy container app %s: %w", containerID, err)
		}
		
		// Wait for container app to be ready
		if err := component.waitForContainerAppHealth(ctx, containerID, 180*time.Second); err != nil {
			return fmt.Errorf("container app %s failed health check: %w", containerID, err)
		}
	}

	return nil
}

// Helper methods for container deployment (stubs for now)

func (component *ContainerOrchestratorComponent) validateInfrastructureHealth(ctx context.Context) error {
	// TODO: Implement infrastructure health validation
	return nil
}

func (component *ContainerOrchestratorComponent) validatePlatformHealth(ctx context.Context) error {
	// TODO: Implement platform health validation  
	return nil
}

func (component *ContainerOrchestratorComponent) validateDaprHealth(ctx context.Context) error {
	// TODO: Implement Dapr health validation
	return nil
}

func (component *ContainerOrchestratorComponent) validateServicesHealth(ctx context.Context) error {
	// TODO: Implement services health validation
	return nil
}

func (component *ContainerOrchestratorComponent) deployPodmanContainer(ctx context.Context, containerID string) error {
	if component.PodmanProvider == nil {
		return fmt.Errorf("Podman provider not initialized for container %s", containerID)
	}

	// Get container specification based on containerID
	spec, err := component.getContainerSpec(containerID)
	if err != nil {
		return fmt.Errorf("failed to get container spec for %s: %w", containerID, err)
	}

	// Pull container image if needed
	if err := component.PodmanProvider.PullImage(ctx, spec.Image); err != nil {
		return fmt.Errorf("failed to pull image for %s: %w", containerID, err)
	}

	// Deploy the main container
	if err := component.PodmanProvider.DeployContainer(ctx, spec); err != nil {
		return fmt.Errorf("failed to deploy container %s: %w", containerID, err)
	}

	// Deploy Dapr sidecar if this is a service container (not infrastructure)
	if component.isServiceContainer(containerID) {
		if err := component.PodmanProvider.DeployDaprSidecar(ctx, spec); err != nil {
			return fmt.Errorf("failed to deploy Dapr sidecar for %s: %w", containerID, err)
		}
	}

	return nil
}

func (component *ContainerOrchestratorComponent) deployAzureContainerApp(ctx context.Context, containerID string) error {
	if component.AzureProvider == nil {
		return fmt.Errorf("Azure provider not initialized for container %s", containerID)
	}

	// Get container specification
	spec, err := component.getContainerSpec(containerID)
	if err != nil {
		return fmt.Errorf("failed to get container spec for %s: %w", containerID, err)
	}

	// Configure Azure-specific scaling rules based on environment
	environment := component.getEnvironment()
	if spec.AzureConfig == nil {
		spec.AzureConfig = &AzureContainerConfig{
			ScalingRules: make(map[string]interface{}),
		}
	}

	switch environment {
	case "staging":
		spec.AzureConfig.ScalingRules["http_requests"] = map[string]interface{}{
			"concurrent_requests": 20,
		}
	case "production":
		spec.AzureConfig.ScalingRules["http_requests"] = map[string]interface{}{
			"concurrent_requests": 100,
		}
	default:
		spec.AzureConfig.ScalingRules["http_requests"] = map[string]interface{}{
			"concurrent_requests": 10,
		}
	}

	// Use the unified interface to deploy the container
	if err := component.AzureProvider.DeployContainer(ctx, spec); err != nil {
		return fmt.Errorf("failed to deploy Azure Container App %s: %w", containerID, err)
	}

	return nil
}

func (component *ContainerOrchestratorComponent) waitForContainerHealth(ctx context.Context, containerID string, timeout time.Duration) error {
	if component.PodmanProvider != nil {
		// Wait for main container health
		if err := component.PodmanProvider.WaitForContainerHealth(ctx, containerID, timeout); err != nil {
			return fmt.Errorf("main container %s failed health check: %w", containerID, err)
		}

		// Wait for Dapr sidecar health if this is a service container
		if component.isServiceContainer(containerID) {
			if err := component.PodmanProvider.GetDaprHealth(ctx, containerID); err != nil {
				return fmt.Errorf("Dapr sidecar for %s failed health check: %w", containerID, err)
			}
		}
	} else if component.AzureProvider != nil {
		// Wait for the container app to be healthy (includes Dapr sidecar)
		if err := component.AzureProvider.WaitForContainerHealth(ctx, containerID, timeout); err != nil {
			return fmt.Errorf("container %s failed health check: %w", containerID, err)
		}

		// Validate Dapr health if this is a service container
		if component.isServiceContainer(containerID) {
			if err := component.AzureProvider.GetDaprHealth(ctx, containerID); err != nil {
				return fmt.Errorf("Dapr sidecar for %s failed health check: %w", containerID, err)
			}
		}
	} else {
		return fmt.Errorf("no container provider initialized for %s", containerID)
	}

	return nil
}

func (component *ContainerOrchestratorComponent) waitForContainerAppHealth(ctx context.Context, containerID string, timeout time.Duration) error {
	// Delegate to the unified health checking method
	return component.waitForContainerHealth(ctx, containerID, timeout)
}

// getContainerSpec returns the container specification for a given containerID
func (component *ContainerOrchestratorComponent) getContainerSpec(containerID string) (*ContainerSpec, error) {
	specs := map[string]*ContainerSpec{
		"dapr-control-plane": {
			Name:           "dapr-control-plane",
			Image:          "daprio/dapr:latest",
			Port:           3500,
			Environment: map[string]string{
				"DAPR_HOST": "0.0.0.0",
			},
			HealthEndpoint: "http://localhost:3500/v1.0/healthz",
			ResourceLimits: ResourceLimits{
				CPU:    "500m",
				Memory: "256Mi",
			},
			DaprEnabled: false, // This is infrastructure, not a Dapr app
			DaprAppID:   "dapr-control-plane",
			DaprPort:    50001,
			DaprConfig: map[string]interface{}{
				"placement_host_address": "localhost:50005",
			},
		},
		"dapr-placement": {
			Name:           "dapr-placement",
			Image:          "daprio/dapr:latest",
			Port:           50005,
			Environment: map[string]string{
				"DAPR_HOST": "0.0.0.0",
			},
			HealthEndpoint: "http://localhost:50005/v1.0/healthz",
			ResourceLimits: ResourceLimits{
				CPU:    "500m",
				Memory: "256Mi",
			},
			DaprEnabled: false, // Infrastructure component
			DaprAppID:   "dapr-placement",
		},
		"dapr-sentry": {
			Name:           "dapr-sentry",
			Image:          "daprio/dapr:latest",
			Port:           50003,
			Environment: map[string]string{
				"DAPR_HOST": "0.0.0.0",
			},
			HealthEndpoint: "http://localhost:50003/v1.0/healthz",
			ResourceLimits: ResourceLimits{
				CPU:    "500m",
				Memory: "256Mi",
			},
			DaprEnabled: false, // Infrastructure component
			DaprAppID:   "dapr-sentry",
		},
		"public-gateway": {
			Name:           "public-gateway",
			Image:          "localhost/backend/public-gateway:latest",
			Port:           9001,
			Environment: map[string]string{
				"DATABASE_CONNECTION_STRING": "postgresql://postgres:5432/development",
				"DAPR_HTTP_ENDPOINT":         "http://localhost:3500",
				"ENVIRONMENT":                "development",
				"LOG_LEVEL":                  "debug",
			},
			HealthEndpoint: "http://localhost:9001/health",
			ResourceLimits: ResourceLimits{
				CPU:    "500m",
				Memory: "256Mi",
			},
			DaprEnabled: true,
			DaprAppID:   "public-gateway",
			DaprPort:    50001,
			DaprConfig: map[string]interface{}{
				"placement_host_address": "localhost:50005",
			},
		},
		"admin-gateway": {
			Name:           "admin-gateway",
			Image:          "localhost/backend/admin-gateway:latest",
			Port:           9000,
			Environment: map[string]string{
				"DATABASE_CONNECTION_STRING": "postgresql://postgres:5432/development",
				"DAPR_HTTP_ENDPOINT":         "http://localhost:3500",
				"ENVIRONMENT":                "development",
				"LOG_LEVEL":                  "debug",
			},
			HealthEndpoint: "http://localhost:9000/health",
			ResourceLimits: ResourceLimits{
				CPU:    "500m",
				Memory: "256Mi",
			},
			DaprEnabled: true,
			DaprAppID:   "admin-gateway",
			DaprPort:    50000,
			DaprConfig: map[string]interface{}{
				"placement_host_address": "localhost:50005",
			},
		},
		"content-news": {
			Name:           "content-news",
			Image:          "localhost/backend/content:latest",
			Port:           3001,
			Environment: map[string]string{
				"DATABASE_CONNECTION_STRING": "postgresql://postgres:5432/development",
				"DAPR_HTTP_ENDPOINT":         "http://localhost:3500",
				"ENVIRONMENT":                "development",
				"SERVICE_TYPE":               "news",
				"LOG_LEVEL":                  "debug",
			},
			HealthEndpoint: "http://localhost:3001/health",
			ResourceLimits: ResourceLimits{
				CPU:    "500m",
				Memory: "256Mi",
			},
			DaprEnabled: true,
			DaprAppID:   "content-news",
			DaprPort:    50011,
			DaprConfig: map[string]interface{}{
				"placement_host_address": "localhost:50005",
			},
		},
		"content-events": {
			Name:           "content-events",
			Image:          "localhost/backend/content:latest",
			Port:           3002,
			Environment: map[string]string{
				"DATABASE_CONNECTION_STRING": "postgresql://postgres:5432/development",
				"DAPR_HTTP_ENDPOINT":         "http://localhost:3500",
				"ENVIRONMENT":                "development",
				"SERVICE_TYPE":               "events",
				"LOG_LEVEL":                  "debug",
			},
			HealthEndpoint: "http://localhost:3002/health",
			ResourceLimits: ResourceLimits{
				CPU:    "500m",
				Memory: "256Mi",
			},
			DaprEnabled: true,
			DaprAppID:   "content-events",
			DaprPort:    50012,
			DaprConfig: map[string]interface{}{
				"placement_host_address": "localhost:50005",
			},
		},
		"content-research": {
			Name:           "content-research",
			Image:          "localhost/backend/content:latest",
			Port:           3003,
			Environment: map[string]string{
				"DATABASE_CONNECTION_STRING": "postgresql://postgres:5432/development",
				"DAPR_HTTP_ENDPOINT":         "http://localhost:3500",
				"ENVIRONMENT":                "development",
				"SERVICE_TYPE":               "research",
				"LOG_LEVEL":                  "debug",
			},
			HealthEndpoint: "http://localhost:3003/health",
			ResourceLimits: ResourceLimits{
				CPU:    "500m",
				Memory: "256Mi",
			},
			DaprEnabled: true,
			DaprAppID:   "content-research",
			DaprPort:    50013,
			DaprConfig: map[string]interface{}{
				"placement_host_address": "localhost:50005",
			},
		},
		"inquiries-business": {
			Name:           "inquiries-business",
			Image:          "localhost/backend/inquiries:latest",
			Port:           3101,
			Environment: map[string]string{
				"DATABASE_CONNECTION_STRING": "postgresql://postgres:5432/development",
				"DAPR_HTTP_ENDPOINT":         "http://localhost:3500",
				"ENVIRONMENT":                "development",
				"SERVICE_TYPE":               "business",
				"LOG_LEVEL":                  "debug",
			},
			HealthEndpoint: "http://localhost:3101/health",
			ResourceLimits: ResourceLimits{
				CPU:    "500m",
				Memory: "256Mi",
			},
			DaprEnabled: true,
			DaprAppID:   "inquiries-business",
			DaprPort:    50021,
			DaprConfig: map[string]interface{}{
				"placement_host_address": "localhost:50005",
			},
		},
		"inquiries-donations": {
			Name:           "inquiries-donations",
			Image:          "localhost/backend/inquiries:latest",
			Port:           3102,
			Environment: map[string]string{
				"DATABASE_CONNECTION_STRING": "postgresql://postgres:5432/development",
				"DAPR_HTTP_ENDPOINT":         "http://localhost:3500",
				"ENVIRONMENT":                "development",
				"SERVICE_TYPE":               "donations",
				"LOG_LEVEL":                  "debug",
			},
			HealthEndpoint: "http://localhost:3102/health",
			ResourceLimits: ResourceLimits{
				CPU:    "500m",
				Memory: "256Mi",
			},
			DaprEnabled: true,
			DaprAppID:   "inquiries-donations",
			DaprPort:    50022,
			DaprConfig: map[string]interface{}{
				"placement_host_address": "localhost:50005",
			},
		},
		"inquiries-media": {
			Name:           "inquiries-media",
			Image:          "localhost/backend/inquiries:latest",
			Port:           3103,
			Environment: map[string]string{
				"DATABASE_CONNECTION_STRING": "postgresql://postgres:5432/development",
				"DAPR_HTTP_ENDPOINT":         "http://localhost:3500",
				"ENVIRONMENT":                "development",
				"SERVICE_TYPE":               "media",
				"LOG_LEVEL":                  "debug",
			},
			HealthEndpoint: "http://localhost:3103/health",
			ResourceLimits: ResourceLimits{
				CPU:    "500m",
				Memory: "256Mi",
			},
			DaprEnabled: true,
			DaprAppID:   "inquiries-media",
			DaprPort:    50023,
			DaprConfig: map[string]interface{}{
				"placement_host_address": "localhost:50005",
			},
		},
		"inquiries-volunteers": {
			Name:           "inquiries-volunteers",
			Image:          "localhost/backend/inquiries:latest",
			Port:           3104,
			Environment: map[string]string{
				"DATABASE_CONNECTION_STRING": "postgresql://postgres:5432/development",
				"DAPR_HTTP_ENDPOINT":         "http://localhost:3500",
				"ENVIRONMENT":                "development",
				"SERVICE_TYPE":               "volunteers",
				"LOG_LEVEL":                  "debug",
			},
			HealthEndpoint: "http://localhost:3104/health",
			ResourceLimits: ResourceLimits{
				CPU:    "500m",
				Memory: "256Mi",
			},
			DaprEnabled: true,
			DaprAppID:   "inquiries-volunteers",
			DaprPort:    50024,
			DaprConfig: map[string]interface{}{
				"placement_host_address": "localhost:50005",
			},
		},
		"notification-service": {
			Name:           "notification-service",
			Image:          "localhost/backend/notifications:latest",
			Port:           3201,
			Environment: map[string]string{
				"DATABASE_CONNECTION_STRING": "postgresql://postgres:5432/development",
				"DAPR_HTTP_ENDPOINT":         "http://localhost:3500",
				"ENVIRONMENT":                "development",
				"LOG_LEVEL":                  "debug",
			},
			HealthEndpoint: "http://localhost:3201/health",
			ResourceLimits: ResourceLimits{
				CPU:    "500m",
				Memory: "256Mi",
			},
			DaprEnabled: true,
			DaprAppID:   "notification-service",
			DaprPort:    50031,
			DaprConfig: map[string]interface{}{
				"placement_host_address": "localhost:50005",
			},
		},
	}

	spec, exists := specs[containerID]
	if !exists {
		return nil, fmt.Errorf("container specification not found for %s", containerID)
	}

	return spec, nil
}

// isServiceContainer determines if a container requires a Dapr sidecar
func (component *ContainerOrchestratorComponent) isServiceContainer(containerID string) bool {
	serviceContainers := map[string]bool{
		"public-gateway":         true,
		"admin-gateway":          true,
		"content-news":           true,
		"content-events":         true,
		"content-research":       true,
		"inquiries-business":     true,
		"inquiries-donations":    true,
		"inquiries-media":        true,
		"inquiries-volunteers":   true,
		"notification-service":   true,
	}

	return serviceContainers[containerID]
}

// getEnvironment returns the environment for this orchestrator component
func (component *ContainerOrchestratorComponent) getEnvironment() string {
	// In real deployment, this would resolve the environment from component configuration
	// For now, determine based on which provider is active
	if component.PodmanProvider != nil {
		return "development"
	} else if component.AzureProvider != nil {
		// This could be resolved from the AzureProvider configuration
		// For now, assume staging if Azure provider is active
		return "staging"
	}
	return "development"
}

// Helper function to extract string from Pulumi output (simplified for this context)
func extractStringFromPulumiOutput(output pulumi.StringOutput) string {
	// In real deployment, this would properly resolve the Pulumi output
	// For now, return a default based on common patterns
	return "podman" // This would be resolved properly in actual deployment context
}