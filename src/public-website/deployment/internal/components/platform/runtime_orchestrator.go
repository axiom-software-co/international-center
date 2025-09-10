package platform

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// RuntimeOrchestrator bridges the gap between Pulumi configuration and actual container execution
// It reads deployment configurations and executes the actual container runtime commands
type RuntimeOrchestrator struct {
	Environment     string
	PodmanProvider  *PodmanProviderComponent
	AzureProvider   *AzureContainerProviderComponent
	HealthChecker   *UnifiedHealthChecker
	DaprManager     *UnifiedDaprSidecarManager
}

// RuntimeExecutionArgs contains arguments for runtime execution
type RuntimeExecutionArgs struct {
	Environment           string
	InfrastructureOutputs pulumi.Map
	PlatformOutputs       pulumi.Map
	ServicesOutputs       pulumi.Map
	ExecutionContext      context.Context
	ExecutionTimeout      time.Duration
}

// ContainerExecutionPlan defines the execution plan for containers
type ContainerExecutionPlan struct {
	InfrastructureContainers []string
	PlatformContainers       []string
	ServiceContainers        []string
	ExecutionOrder           []string
	DependencyGraph          map[string][]string
}

// NewRuntimeOrchestrator creates a new runtime orchestrator for the given environment
func NewRuntimeOrchestrator(environment string) *RuntimeOrchestrator {
	orchestrator := &RuntimeOrchestrator{
		Environment:   environment,
		HealthChecker: NewUnifiedHealthChecker(),
		DaprManager:   NewUnifiedDaprSidecarManager(environment),
	}

	// Initialize providers based on environment
	switch environment {
	case "development":
		orchestrator.PodmanProvider, _ = NewPodmanProviderComponent(nil, "runtime-podman", &PodmanProviderArgs{
			Environment: environment,
		})
	case "staging", "production":
		orchestrator.AzureProvider, _ = NewAzureContainerProviderComponent(nil, "runtime-azure", &AzureContainerProviderArgs{
			ContainerEnvironment: environment,
		})
	}

	return orchestrator
}

// ExecuteRuntimeDeployment executes the actual container deployment based on Pulumi configurations
func (r *RuntimeOrchestrator) ExecuteRuntimeDeployment(ctx context.Context, args *RuntimeExecutionArgs) error {
	log.Printf("Starting runtime container execution for environment: %s", r.Environment)

	// Build execution plan based on deployment outputs
	plan, err := r.buildExecutionPlan(args)
	if err != nil {
		return fmt.Errorf("failed to build execution plan: %w", err)
	}

	// Execute containers in proper dependency order
	if err := r.executeContainersPlan(ctx, plan, args); err != nil {
		return fmt.Errorf("failed to execute containers plan: %w", err)
	}

	// Validate all containers are running and healthy
	if err := r.validateRuntimeHealth(ctx, plan); err != nil {
		return fmt.Errorf("runtime health validation failed: %w", err)
	}

	log.Printf("Runtime container execution completed successfully for environment: %s", r.Environment)
	return nil
}

// buildExecutionPlan creates the container execution plan based on deployment outputs
func (r *RuntimeOrchestrator) buildExecutionPlan(args *RuntimeExecutionArgs) (*ContainerExecutionPlan, error) {
	plan := &ContainerExecutionPlan{
		InfrastructureContainers: r.getInfrastructureContainerList(),
		PlatformContainers:       r.getPlatformContainerList(),
		ServiceContainers:        r.getServiceContainerList(),
		DependencyGraph:          r.buildDependencyGraph(),
	}

	// Include infrastructure, platform, and service containers in execution order
	allContainers := make(map[string][]string)
	
	// Add infrastructure containers
	for _, container := range plan.InfrastructureContainers {
		allContainers[container] = plan.DependencyGraph[container]
	}
	
	// Add platform containers
	for _, container := range plan.PlatformContainers {
		allContainers[container] = plan.DependencyGraph[container]
	}

	// Add service containers  
	for _, container := range plan.ServiceContainers {
		allContainers[container] = plan.DependencyGraph[container]
	}

	// Build execution order based on all container dependencies
	executionOrder, err := r.calculateExecutionOrder(allContainers)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate execution order: %w", err)
	}
	plan.ExecutionOrder = executionOrder

	return plan, nil
}

// executeContainersPlan executes containers according to the plan
func (r *RuntimeOrchestrator) executeContainersPlan(ctx context.Context, plan *ContainerExecutionPlan, args *RuntimeExecutionArgs) error {
	switch r.Environment {
	case "development":
		return r.executePodmanContainers(ctx, plan, args)
	case "staging", "production":
		return r.executeAzureContainers(ctx, plan, args)
	default:
		return fmt.Errorf("unsupported environment for runtime execution: %s", r.Environment)
	}
}

// executePodmanContainers executes containers using Podman for development environment
func (r *RuntimeOrchestrator) executePodmanContainers(ctx context.Context, plan *ContainerExecutionPlan, args *RuntimeExecutionArgs) error {
	if r.PodmanProvider == nil {
		return fmt.Errorf("Podman provider not initialized")
	}

	// Initialize Podman network first
	if err := r.PodmanProvider.CreateNetwork(ctx); err != nil {
		return fmt.Errorf("failed to create Podman network: %w", err)
	}

	// Execute containers in dependency order (infrastructure, platform, and services)
	for _, containerID := range plan.ExecutionOrder {
		log.Printf("Deploying container: %s", containerID)
		
		// Deploy all container types (infrastructure, platform, and services)
		
		// Get container specification
		spec, err := r.getContainerSpecification(containerID, args)
		if err != nil {
			return fmt.Errorf("failed to get specification for container %s: %w", containerID, err)
		}

		// Pull container image
		if err := r.PodmanProvider.PullImage(ctx, spec.Image); err != nil {
			return fmt.Errorf("failed to pull image for %s: %w", containerID, err)
		}

		// Deploy the main container
		if err := r.PodmanProvider.DeployContainer(ctx, spec); err != nil {
			return fmt.Errorf("failed to deploy container %s: %w", containerID, err)
		}

		// For standalone mode, services connect directly to Dapr control plane
		// No individual sidecars needed in development environment
		if r.isServiceContainer(containerID) && spec.DaprEnabled {
			log.Printf("Service %s configured for standalone Dapr mode - connecting to control plane", containerID)
		}

		// Wait for container to be healthy before proceeding to next
		if err := r.PodmanProvider.WaitForContainerHealth(ctx, containerID, 60*time.Second); err != nil {
			log.Printf("Warning: Container %s may not be healthy yet: %v", containerID, err)
		}

		log.Printf("Successfully deployed container: %s", containerID)
	}

	return nil
}

// executeAzureContainers executes containers using Azure Container Apps for staging/production
func (r *RuntimeOrchestrator) executeAzureContainers(ctx context.Context, plan *ContainerExecutionPlan, args *RuntimeExecutionArgs) error {
	if r.AzureProvider == nil {
		return fmt.Errorf("Azure provider not initialized")
	}

	// Execute containers in dependency order
	for _, containerID := range plan.ExecutionOrder {
		log.Printf("Deploying Azure container: %s", containerID)
		
		// Get container specification
		spec, err := r.getContainerSpecification(containerID, args)
		if err != nil {
			return fmt.Errorf("failed to get specification for container %s: %w", containerID, err)
		}

		// Deploy to Azure Container Apps
		if err := r.AzureProvider.DeployContainer(ctx, spec); err != nil {
			return fmt.Errorf("failed to deploy Azure container %s: %w", containerID, err)
		}

		// Azure automatically handles Dapr sidecars, so no explicit sidecar deployment needed

		log.Printf("Successfully deployed Azure container: %s", containerID)
	}

	return nil
}

// validateRuntimeHealth validates that all deployed containers are running and healthy
func (r *RuntimeOrchestrator) validateRuntimeHealth(ctx context.Context, plan *ContainerExecutionPlan) error {
	// In standalone mode, validate infrastructure, platform, and services (no individual sidecars)
	containersToValidate := append(append(plan.InfrastructureContainers, plan.PlatformContainers...), plan.ServiceContainers...)

	switch r.Environment {
	case "development":
		if r.PodmanProvider == nil {
			return fmt.Errorf("Podman provider not available for health validation")
		}
		
		// Check deployed containers using Podman provider health checker
		results := r.HealthChecker.CheckMultipleContainers(ctx, containersToValidate, r.PodmanProvider)
		healthy, unhealthy, issues := r.HealthChecker.GetHealthSummary(results)
		
		if unhealthy > 0 {
			log.Printf("Runtime health validation: %d unhealthy containers: %v", unhealthy, issues)
			// Continue execution even with unhealthy containers for development
		}
		
		log.Printf("Runtime health validation completed: %d healthy, %d unhealthy containers", healthy, unhealthy)
		return nil

	case "staging", "production":
		if r.AzureProvider == nil {
			return fmt.Errorf("Azure provider not available for health validation")
		}
		
		// Check deployed containers using Azure provider health checker  
		results := r.HealthChecker.CheckMultipleContainers(ctx, containersToValidate, r.AzureProvider)
		healthy, unhealthy, issues := r.HealthChecker.GetHealthSummary(results)
		
		if unhealthy > 0 {
			return fmt.Errorf("runtime health validation failed: %d unhealthy containers: %v", unhealthy, issues)
		}
		
		log.Printf("Runtime health validation successful: %d healthy containers", healthy)
		return nil

	default:
		return fmt.Errorf("unsupported environment for health validation: %s", r.Environment)
	}
}

// getInfrastructureContainerList returns list of infrastructure containers that need to be deployed
func (r *RuntimeOrchestrator) getInfrastructureContainerList() []string {
	return []string{
		"postgresql",
		"vault", 
		"rabbitmq",
		"azurite",
	}
}

// getPlatformContainerList returns list of platform containers that need to be deployed  
func (r *RuntimeOrchestrator) getPlatformContainerList() []string {
	switch r.Environment {
	case "development":
		// For development, use simplified Dapr setup with just control plane
		return []string{
			"dapr-control-plane",
		}
	default:
		// For staging/production, use full Dapr control plane setup
		return []string{
			"dapr-placement",
			"dapr-control-plane",
			"dapr-sentry",
		}
	}
}

// getServiceContainerList returns list of consolidated service containers that need to be deployed
func (r *RuntimeOrchestrator) getServiceContainerList() []string {
	return []string{
		"public-gateway",
		"admin-gateway", 
		"content",
		"inquiries",
		"notifications",
	}
}

// buildDependencyGraph creates dependency relationships between containers
func (r *RuntimeOrchestrator) buildDependencyGraph() map[string][]string {
	dependencies := map[string][]string{
		// Infrastructure containers (no dependencies)
		"postgresql": {},
		"vault":      {},
		"rabbitmq":   {},
		"azurite":    {},
	}

	// Add environment-specific platform dependencies
	switch r.Environment {
	case "development":
		// Simplified Dapr setup for development
		dependencies["dapr-control-plane"] = []string{"postgresql"}
		
		// Consolidated service containers depend on simplified control plane
		dependencies["public-gateway"] = []string{"dapr-control-plane", "postgresql", "vault"}
		dependencies["admin-gateway"] = []string{"dapr-control-plane", "postgresql", "vault"}
		dependencies["content"] = []string{"dapr-control-plane", "postgresql"}
		dependencies["inquiries"] = []string{"dapr-control-plane", "postgresql"}
		dependencies["notifications"] = []string{"dapr-control-plane", "rabbitmq", "postgresql"}
		
	default:
		// Full Dapr control plane setup for staging/production
		dependencies["dapr-placement"] = []string{"postgresql", "vault"}
		dependencies["dapr-control-plane"] = []string{"dapr-placement"}
		dependencies["dapr-sentry"] = []string{"dapr-placement"}
		
		// Consolidated service containers depend on full control plane
		dependencies["public-gateway"] = []string{"dapr-control-plane", "postgresql", "vault"}
		dependencies["admin-gateway"] = []string{"dapr-control-plane", "postgresql", "vault"}
		dependencies["content"] = []string{"dapr-control-plane", "postgresql"}
		dependencies["inquiries"] = []string{"dapr-control-plane", "postgresql"}
		dependencies["notifications"] = []string{"dapr-control-plane", "rabbitmq", "postgresql"}
	}

	return dependencies
}

// calculateExecutionOrder determines the order containers should be started based on dependencies
func (r *RuntimeOrchestrator) calculateExecutionOrder(dependencies map[string][]string) ([]string, error) {
	var order []string
	visited := make(map[string]bool)
	visiting := make(map[string]bool)

	var visit func(container string) error
	visit = func(container string) error {
		if visiting[container] {
			return fmt.Errorf("circular dependency detected involving container: %s", container)
		}
		if visited[container] {
			return nil
		}

		visiting[container] = true
		
		// Visit all dependencies first
		for _, dep := range dependencies[container] {
			if err := visit(dep); err != nil {
				return err
			}
		}
		
		visiting[container] = false
		visited[container] = true
		order = append(order, container)
		return nil
	}

	// Visit all containers
	for container := range dependencies {
		if err := visit(container); err != nil {
			return nil, fmt.Errorf("failed to calculate execution order: %w", err)
		}
	}

	return order, nil
}

// getContainerSpecification returns container specification for a given container ID
func (r *RuntimeOrchestrator) getContainerSpecification(containerID string, args *RuntimeExecutionArgs) (*ContainerSpec, error) {
	// Configure container based on type and deployment outputs
	switch {
	case r.isInfrastructureContainer(containerID):
		return r.buildInfrastructureContainerSpec(containerID, args)
	case r.isPlatformContainer(containerID):
		return r.buildPlatformContainerSpec(containerID, args)
	case r.isServiceContainer(containerID):
		return r.buildServiceContainerSpec(containerID, args)
	default:
		return nil, fmt.Errorf("unknown container type for: %s", containerID)
	}
}

// buildInfrastructureContainerSpec builds container spec for infrastructure containers
func (r *RuntimeOrchestrator) buildInfrastructureContainerSpec(containerID string, args *RuntimeExecutionArgs) (*ContainerSpec, error) {
	switch containerID {
	case "postgresql":
		spec := NewContainerSpecBuilder(containerID, "postgres:15", 5432).
			WithEnvironment(map[string]string{
				"POSTGRES_USER":     "postgres",
				"POSTGRES_PASSWORD": "password",
				"POSTGRES_DB":       "international_center_development",
			}).
			WithResourceLimits("0.5", "256m").
			WithHealthEndpoint("postgresql://localhost:5432")
		return spec.Build()

	case "vault":
		spec := NewContainerSpecBuilder(containerID, "hashicorp/vault:latest", 8200).
			WithEnvironment(map[string]string{
				"VAULT_DEV_ROOT_TOKEN_ID":   "development",
				"VAULT_DEV_LISTEN_ADDRESS": "0.0.0.0:8200",
			}).
			WithResourceLimits("0.2", "128m").
			WithHealthEndpoint("http://localhost:8200/v1/sys/health")
		return spec.Build()

	case "rabbitmq":
		spec := NewContainerSpecBuilder(containerID, "rabbitmq:3-management", 5672).
			WithEnvironment(map[string]string{
				"RABBITMQ_DEFAULT_USER": "guest",
				"RABBITMQ_DEFAULT_PASS": "guest",
			}).
			WithResourceLimits("0.3", "256m").
			WithHealthEndpoint("http://localhost:15672/api/overview")
		return spec.Build()

	case "azurite":
		spec := NewContainerSpecBuilder(containerID, "mcr.microsoft.com/azure-storage/azurite:latest", 10000).
			WithResourceLimits("0.2", "128m").
			WithHealthEndpoint("http://localhost:10000")
		return spec.Build()

	default:
		return nil, fmt.Errorf("unknown infrastructure container: %s", containerID)
	}
}

// buildPlatformContainerSpec builds container spec for platform containers (Dapr)
func (r *RuntimeOrchestrator) buildPlatformContainerSpec(containerID string, args *RuntimeExecutionArgs) (*ContainerSpec, error) {
	switch containerID {
	case "dapr-placement":
		spec := NewContainerSpecBuilder(containerID, "daprio/dapr:latest", 50005).
			WithCommand([]string{"/placement", "--port", "50005"}).
			WithResourceLimits("0.2", "128m")
		return spec.Build()

	case "dapr-control-plane":
		// Use standalone mode for development - simpler setup without placement service
		spec := NewContainerSpecBuilder(containerID, "daprio/dapr:latest", 3500).
			WithCommand([]string{"/daprd", "--mode", "standalone", "--dapr-http-port", "3500", "--dapr-grpc-port", "50001", "--log-level", "info", "--app-id", "control-plane"}).
			WithResourceLimits("0.5", "256m").
			WithHealthEndpoint("http://localhost:3500/v1.0/healthz")
		return spec.Build()

	case "dapr-sentry":
		spec := NewContainerSpecBuilder(containerID, "daprio/dapr:latest", 50003).
			WithCommand([]string{"/sentry", "--port", "50003", "--log-level", "info"}).
			WithResourceLimits("0.2", "128m")
		return spec.Build()

	default:
		return nil, fmt.Errorf("unknown platform container: %s", containerID)
	}
}

// buildServiceContainerSpec builds container spec for service containers
func (r *RuntimeOrchestrator) buildServiceContainerSpec(containerID string, args *RuntimeExecutionArgs) (*ContainerSpec, error) {
	switch containerID {
	case "public-gateway":
		spec := NewContainerSpecBuilder(containerID, "localhost/backend/public-gateway:latest", 9001).
			WithEnvironment(map[string]string{
				"SERVICE_TYPE":     "gateway",
				"GATEWAY_TYPE":    "public",
				"DAPR_HOST":       "dapr-control-plane",
				"DAPR_HTTP_PORT":  "3500",
				"DAPR_GRPC_PORT":  "50001",
			}).
			WithResourceLimits("0.5", "256m").
			WithHealthEndpoint("http://localhost:9001/health").
			WithDapr("public-gateway", 9001)
		return spec.Build()

	case "admin-gateway":
		spec := NewContainerSpecBuilder(containerID, "localhost/backend/admin-gateway:latest", 9000).
			WithEnvironment(map[string]string{
				"SERVICE_TYPE":     "gateway",
				"GATEWAY_TYPE":    "admin",
				"DAPR_HOST":       "dapr-control-plane",
				"DAPR_HTTP_PORT":  "3500",
				"DAPR_GRPC_PORT":  "50001",
			}).
			WithResourceLimits("0.5", "256m").
			WithHealthEndpoint("http://localhost:9000/health").
			WithDapr("admin-gateway", 9000)
		return spec.Build()

	case "content":
		spec := NewContainerSpecBuilder(containerID, "localhost/backend/content:latest", 3001).
			WithEnvironment(map[string]string{
				"SERVICE_TYPE":     "content",
				"DAPR_HOST":       "dapr-control-plane",
				"DAPR_HTTP_PORT":  "3500",
				"DAPR_GRPC_PORT":  "50001",
			}).
			WithResourceLimits("0.5", "512m").
			WithHealthEndpoint("http://localhost:3001/health").
			WithDapr("content", 3001)
		return spec.Build()

	case "inquiries":
		spec := NewContainerSpecBuilder(containerID, "localhost/backend/inquiries:latest", 3101).
			WithEnvironment(map[string]string{
				"SERVICE_TYPE":     "inquiries",
				"DAPR_HOST":       "dapr-control-plane",
				"DAPR_HTTP_PORT":  "3500",
				"DAPR_GRPC_PORT":  "50001",
			}).
			WithResourceLimits("0.5", "512m").
			WithHealthEndpoint("http://localhost:3101/health").
			WithDapr("inquiries", 3101)
		return spec.Build()

	case "notifications":
		spec := NewContainerSpecBuilder(containerID, "localhost/backend/notifications:latest", 3201).
			WithEnvironment(map[string]string{
				"SERVICE_TYPE":     "notifications",
				"DAPR_HOST":       "dapr-control-plane",
				"DAPR_HTTP_PORT":  "3500",
				"DAPR_GRPC_PORT":  "50001",
			}).
			WithResourceLimits("0.5", "512m").
			WithHealthEndpoint("http://localhost:3201/health").
			WithDapr("notifications", 3201)
		return spec.Build()

	default:
		return nil, fmt.Errorf("unknown service container: %s", containerID)
	}
}

// Helper methods for container classification

func (r *RuntimeOrchestrator) isInfrastructureContainer(containerID string) bool {
	infra := r.getInfrastructureContainerList()
	for _, container := range infra {
		if container == containerID {
			return true
		}
	}
	return false
}

func (r *RuntimeOrchestrator) isPlatformContainer(containerID string) bool {
	platform := r.getPlatformContainerList()
	for _, container := range platform {
		if container == containerID {
			return true
		}
	}
	return false
}

func (r *RuntimeOrchestrator) isServiceContainer(containerID string) bool {
	services := r.getServiceContainerList()
	for _, container := range services {
		if container == containerID {
			return true
		}
	}
	return false
}

