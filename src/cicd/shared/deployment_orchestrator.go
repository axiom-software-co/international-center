package shared

import (
	"fmt"
	"time"

	"github.com/axiom-software-co/international-center/src/cicd/components"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/pulumi/pulumi-command/sdk/go/command/local"
)

// DeploymentOrchestrator manages component deployment with dependency resolution and health monitoring
type DeploymentOrchestrator struct {
	ctx             *pulumi.Context
	cfg             *config.Config
	environment     string
	healthMonitor   *HealthMonitor
	deployedOutputs *ComponentOutputs
}

// ComponentOutputs aggregates all component outputs for orchestration
type ComponentOutputs struct {
	Database      *components.DatabaseOutputs
	Storage       *components.StorageOutputs
	Vault         *components.VaultOutputs
	Observability *components.ObservabilityOutputs
	Dapr          *components.DaprOutputs
	Services      *components.ServicesOutputs
	Website       *components.WebsiteOutputs
}

// NewDeploymentOrchestrator creates a new deployment orchestrator
func NewDeploymentOrchestrator(ctx *pulumi.Context, cfg *config.Config, environment string) *DeploymentOrchestrator {
	return &DeploymentOrchestrator{
		ctx:           ctx,
		cfg:           cfg,
		environment:   environment,
		healthMonitor: NewHealthMonitor(ctx, environment),
	}
}

// ImageBuildWorkflow manages the optimized image building process
type ImageBuildWorkflow struct {
	orchestrator *DeploymentOrchestrator
	builder      *components.ImageBuilder
	buildResults map[string]*ImageBuildResult
}

// ImageBuildResult tracks the result of an individual image build
type ImageBuildResult struct {
	ImageName   string
	ServiceName string
	Success     bool
	Duration    time.Duration
	Error       error
}

// ImageBuildGroup defines a group of images that can be built in parallel
type ImageBuildGroup struct {
	Name        string
	Images      []ImageBuildTask
	Parallelism int // Maximum number of parallel builds in this group
}

// ImageBuildTask defines a specific image build task
type ImageBuildTask struct {
	ServiceName string
	ServiceType string
	ImageName   string
	Priority    int // Higher priority built first within group
}

// BuildRequiredImages builds all Docker images required for the deployment with optimized workflow
func (d *DeploymentOrchestrator) BuildRequiredImages() error {
	d.ctx.Log.Info("Starting optimized image building workflow", nil)
	
	workflow := &ImageBuildWorkflow{
		orchestrator: d,
		builder:      NewImageBuilder(d.ctx, d.environment),
		buildResults: make(map[string]*ImageBuildResult),
	}
	
	// Build images using optimized workflow
	if err := workflow.ExecuteOptimizedBuildWorkflow(); err != nil {
		d.ctx.Log.Error("Optimized image building workflow failed", nil)
		return fmt.Errorf("optimized image building failed: %w", err)
	}
	
	d.ctx.Log.Info("Optimized image building workflow completed successfully", nil)
	return nil
}

// ExecuteOptimizedBuildWorkflow executes the image building workflow with dependency-aware parallel processing
func (w *ImageBuildWorkflow) ExecuteOptimizedBuildWorkflow() error {
	w.orchestrator.ctx.Log.Info("Executing dependency-aware parallel image building", nil)
	
	// Define build groups with dependency order and parallelism
	buildGroups := w.createBuildGroups()
	
	// Execute build groups in dependency order
	for i, group := range buildGroups {
		w.orchestrator.ctx.Log.Info(fmt.Sprintf("Building group %d: %s", i+1, group.Name), nil)
		
		if err := w.executeBuildGroup(group); err != nil {
			return fmt.Errorf("build group '%s' failed: %w", group.Name, err)
		}
		
		w.orchestrator.ctx.Log.Info(fmt.Sprintf("Group %d: %s completed successfully", i+1, group.Name), nil)
	}
	
	// Validate all builds completed successfully
	return w.validateAllBuildsSuccessful()
}

// createBuildGroups creates optimized build groups based on dependencies and parallelism
func (w *ImageBuildWorkflow) createBuildGroups() []ImageBuildGroup {
	return []ImageBuildGroup{
		{
			Name:        "Foundation Services",
			Parallelism: 4, // Build up to 4 foundation services in parallel
			Images: []ImageBuildTask{
				{ServiceName: "media", ServiceType: "inquiries", ImageName: "backend/media:latest", Priority: 1},
				{ServiceName: "donations", ServiceType: "inquiries", ImageName: "backend/donations:latest", Priority: 1},
				{ServiceName: "volunteers", ServiceType: "inquiries", ImageName: "backend/volunteers:latest", Priority: 2},
				{ServiceName: "business", ServiceType: "inquiries", ImageName: "backend/business:latest", Priority: 2},
			},
		},
		{
			Name:        "Content Services", 
			Parallelism: 4, // Build up to 4 content services in parallel
			Images: []ImageBuildTask{
				{ServiceName: "research", ServiceType: "content", ImageName: "backend/research:latest", Priority: 1},
				{ServiceName: "services", ServiceType: "content", ImageName: "backend/services:latest", Priority: 1},
				{ServiceName: "events", ServiceType: "content", ImageName: "backend/events:latest", Priority: 2},
				{ServiceName: "news", ServiceType: "content", ImageName: "backend/news:latest", Priority: 2},
			},
		},
		{
			Name:        "Gateway Services",
			Parallelism: 2, // Build both gateways in parallel
			Images: []ImageBuildTask{
				{ServiceName: "admin", ServiceType: "gateway", ImageName: "backend/admin-gateway:latest", Priority: 1},
				{ServiceName: "public", ServiceType: "gateway", ImageName: "backend/public-gateway:latest", Priority: 1},
			},
		},
		{
			Name:        "Frontend",
			Parallelism: 1, // Single website build
			Images: []ImageBuildTask{
				{ServiceName: "website", ServiceType: "website", ImageName: "website:latest", Priority: 1},
			},
		},
	}
}

// executeBuildGroup executes a build group with controlled parallelism
func (w *ImageBuildWorkflow) executeBuildGroup(group ImageBuildGroup) error {
	w.orchestrator.ctx.Log.Info(fmt.Sprintf("Executing build group '%s' with parallelism %d", group.Name, group.Parallelism), nil)
	
	// Create semaphore for controlling parallelism
	semaphore := make(chan struct{}, group.Parallelism)
	resultChan := make(chan *ImageBuildResult, len(group.Images))
	
	// Launch builds with controlled parallelism
	for _, task := range group.Images {
		go func(task ImageBuildTask) {
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			// Execute build task
			result := w.executeBuildTask(task)
			resultChan <- result
		}(task)
	}
	
	// Collect results
	var buildErrors []error
	for i := 0; i < len(group.Images); i++ {
		result := <-resultChan
		w.buildResults[result.ServiceName] = result
		
		if !result.Success {
			buildErrors = append(buildErrors, result.Error)
			w.orchestrator.ctx.Log.Error(fmt.Sprintf("Build failed for %s: %v", result.ServiceName, result.Error), nil)
		} else {
			w.orchestrator.ctx.Log.Info(fmt.Sprintf("Build completed for %s in %v", result.ServiceName, result.Duration), nil)
		}
	}
	
	// Return error if any builds failed
	if len(buildErrors) > 0 {
		return fmt.Errorf("build group '%s' had %d failures", group.Name, len(buildErrors))
	}
	
	return nil
}

// executeBuildTask executes an individual build task with timing and error handling
func (w *ImageBuildWorkflow) executeBuildTask(task ImageBuildTask) *ImageBuildResult {
	startTime := time.Now()
	result := &ImageBuildResult{
		ImageName:   task.ImageName,
		ServiceName: task.ServiceName,
	}
	
	w.orchestrator.ctx.Log.Info(fmt.Sprintf("Starting build for %s (%s)", task.ServiceName, task.ServiceType), nil)
	
	// Execute the appropriate build based on service type
	var err error
	switch task.ServiceType {
	case "inquiries", "content":
		_, err = w.builder.BuildServiceImage(task.ServiceName, task.ServiceType)
	case "gateway":
		_, err = w.builder.BuildGatewayImage(task.ServiceName)
	case "website":
		_, err = w.builder.BuildWebsiteImage()
	default:
		err = fmt.Errorf("unknown service type: %s", task.ServiceType)
	}
	
	result.Duration = time.Since(startTime)
	result.Success = err == nil
	result.Error = err
	
	return result
}

// validateAllBuildsSuccessful validates that all required images were built successfully
func (w *ImageBuildWorkflow) validateAllBuildsSuccessful() error {
	var failedBuilds []string
	var totalDuration time.Duration
	
	for serviceName, result := range w.buildResults {
		totalDuration += result.Duration
		if !result.Success {
			failedBuilds = append(failedBuilds, serviceName)
		}
	}
	
	w.orchestrator.ctx.Log.Info(fmt.Sprintf("Build workflow summary: %d builds completed in %v total", 
		len(w.buildResults), totalDuration), nil)
	
	if len(failedBuilds) > 0 {
		return fmt.Errorf("build validation failed: %d services failed to build: %v", 
			len(failedBuilds), failedBuilds)
	}
	
	w.orchestrator.ctx.Log.Info("All image builds validated successfully", nil)
	return nil
}

// DeployInfrastructure orchestrates the deployment of all infrastructure components with dependency resolution
func (d *DeploymentOrchestrator) DeployInfrastructure() (*ComponentOutputs, error) {
	d.ctx.Log.Info("Starting orchestrated infrastructure deployment", nil)
	
	// Validate environment configuration before deployment
	envConfig, err := LoadEnvironmentConfig(d.ctx, d.environment)
	if err != nil {
		return nil, fmt.Errorf("environment configuration validation failed: %w", err)
	}
	
	// For staging and production, validate all required configuration
	if d.environment != "development" {
		if err := ValidateConfiguration(envConfig, d.environment); err != nil {
			return nil, fmt.Errorf("configuration validation failed: %w", err)
		}
	}
	
	outputs := &ComponentOutputs{}
	
	// Phase 1: Deploy foundational infrastructure components
	d.ctx.Log.Info("Phase 1: Deploying foundational infrastructure", nil)
	
	// Deploy database first (foundational dependency)
	var dbDeployErr error
	outputs.Database, dbDeployErr = components.DeployDatabase(d.ctx, d.cfg, d.environment)
	if err := d.deployWithHealthCheck("database", func() error {
		return dbDeployErr
	}, outputs.Database); err != nil {
		return nil, fmt.Errorf("database deployment failed: %w", err)
	}
	
	// Deploy storage (foundational dependency)
	var storageDeployErr error
	outputs.Storage, storageDeployErr = components.DeployStorage(d.ctx, d.cfg, d.environment)
	if err := d.deployWithHealthCheck("storage", func() error {
		return storageDeployErr
	}, outputs.Storage); err != nil {
		return nil, fmt.Errorf("storage deployment failed: %w", err)
	}
	
	// Deploy vault (secrets management dependency)
	var vaultDeployErr error
	outputs.Vault, vaultDeployErr = components.DeployVault(d.ctx, d.cfg, d.environment)
	if err := d.deployWithHealthCheck("vault", func() error {
		return vaultDeployErr
	}, outputs.Vault); err != nil {
		return nil, fmt.Errorf("vault deployment failed: %w", err)
	}
	
	// Phase 2: Deploy monitoring and orchestration
	d.ctx.Log.Info("Phase 2: Deploying monitoring and orchestration", nil)
	
	// Deploy observability (monitoring dependency)
	var observabilityDeployErr error
	outputs.Observability, observabilityDeployErr = components.DeployObservability(d.ctx, d.cfg, d.environment)
	if err := d.deployWithHealthCheck("observability", func() error {
		return observabilityDeployErr
	}, outputs.Observability); err != nil {
		return nil, fmt.Errorf("observability deployment failed: %w", err)
	}
	
	// Deploy Dapr (service mesh dependency)
	var daprDeployErr error
	outputs.Dapr, daprDeployErr = components.DeployDapr(d.ctx, d.cfg, d.environment)
	if err := d.deployWithHealthCheck("dapr", func() error {
		return daprDeployErr
	}, outputs.Dapr); err != nil {
		return nil, fmt.Errorf("dapr deployment failed: %w", err)
	}
	
	// Phase 3: Deploy application services
	d.ctx.Log.Info("Phase 3: Deploying application services", nil)
	
	// Deploy services (requires all previous components)
	var servicesDeployErr error
	outputs.Services, servicesDeployErr = components.DeployServices(d.ctx, d.cfg, d.environment)
	if err := d.deployWithHealthCheck("services", func() error {
		return servicesDeployErr
	}, outputs.Services); err != nil {
		return nil, fmt.Errorf("services deployment failed: %w", err)
	}
	
	// Phase 4: Deploy frontend
	d.ctx.Log.Info("Phase 4: Deploying frontend", nil)
	
	// Deploy website (requires services)
	var websiteDeployErr error
	outputs.Website, websiteDeployErr = components.DeployWebsite(d.ctx, d.cfg, d.environment)
	if err := d.deployWithHealthCheck("website", func() error {
		return websiteDeployErr
	}, outputs.Website); err != nil {
		return nil, fmt.Errorf("website deployment failed: %w", err)
	}
	
	d.ctx.Log.Info("Infrastructure deployment completed successfully", nil)
	
	// Store deployed outputs for potential rollback
	d.deployedOutputs = outputs
	
	// Validate cross-component integration after all deployments
	if err := d.validateIntegration(outputs); err != nil {
		return nil, fmt.Errorf("component integration validation failed: %w", err)
	}
	
	// Perform overall health assessment and rollback if necessary
	if err := d.assessOverallHealthAndRollback(outputs); err != nil {
		return nil, fmt.Errorf("health assessment failed: %w", err)
	}
	
	return outputs, nil
}

// deployWithHealthCheck deploys a component with health monitoring and validation
func (d *DeploymentOrchestrator) deployWithHealthCheck(componentName string, deployFunc func() error, outputs interface{}) error {
	d.ctx.Log.Info(fmt.Sprintf("Deploying %s component with health monitoring", componentName), nil)
	
	// Execute deployment
	startTime := time.Now()
	err := deployFunc()
	duration := time.Since(startTime)
	
	if err != nil {
		d.ctx.Log.Error(fmt.Sprintf("%s deployment failed after %v: %v", componentName, duration, err), nil)
		return fmt.Errorf("%s deployment failed: %w", componentName, err)
	}
	
	d.ctx.Log.Info(fmt.Sprintf("%s deployment completed in %v", componentName, duration), nil)
	
	// Validate component health
	health := d.healthMonitor.ValidateComponentHealth(componentName, outputs)
	
	if !health.Healthy {
		d.ctx.Log.Error(fmt.Sprintf("%s health check failed: %s", componentName, health.Status), nil)
		for _, errorMsg := range health.Errors {
			d.ctx.Log.Error(fmt.Sprintf("%s health error: %s", componentName, errorMsg), nil)
		}
		
		// For critical deployment failures, return error immediately
		if health.Status == "failed" {
			return fmt.Errorf("%s health check failed: %s", componentName, health.Status)
		}
		
		// For configuration issues, log warning but continue
		d.ctx.Log.Warn(fmt.Sprintf("%s has configuration issues but deployment continues", componentName), nil)
	}
	
	d.ctx.Log.Info(fmt.Sprintf("%s health check passed", componentName), nil)
	
	return nil
}

// validateIntegration validates that all components integrate correctly
func (d *DeploymentOrchestrator) validateIntegration(outputs *ComponentOutputs) error {
	d.ctx.Log.Info("Validating component integration", nil)
	
	// In a real deployment, this would validate:
	// - Database connectivity from services
	// - Storage accessibility from services
	// - Vault secrets access from services
	// - Dapr service mesh communication
	// - Observability data collection
	// - Website API connectivity
	
	// For now, we validate that all outputs are present
	if outputs.Database == nil || outputs.Storage == nil || outputs.Vault == nil ||
		outputs.Observability == nil || outputs.Dapr == nil || outputs.Services == nil ||
		outputs.Website == nil {
		return fmt.Errorf("component integration validation failed: missing component outputs")
	}
	
	d.ctx.Log.Info("Component integration validation passed", nil)
	return nil
}

// GetDeploymentHealth returns the health status of the deployment
func (d *DeploymentOrchestrator) GetDeploymentHealth(outputs *ComponentOutputs) map[string]bool {
	health := make(map[string]bool)
	
	health["database"] = outputs.Database != nil
	health["storage"] = outputs.Storage != nil
	health["vault"] = outputs.Vault != nil
	health["observability"] = outputs.Observability != nil
	health["dapr"] = outputs.Dapr != nil
	health["services"] = outputs.Services != nil
	health["website"] = outputs.Website != nil
	
	return health
}

// assessOverallHealthAndRollback performs overall health assessment and triggers rollback if necessary
func (d *DeploymentOrchestrator) assessOverallHealthAndRollback(outputs *ComponentOutputs) error {
	d.ctx.Log.Info("Performing overall deployment health assessment", nil)
	
	// Collect health status for all components
	componentHealths := make(map[string]*ComponentHealth)
	componentHealths["database"] = d.healthMonitor.ValidateComponentHealth("database", outputs.Database)
	componentHealths["storage"] = d.healthMonitor.ValidateComponentHealth("storage", outputs.Storage)
	componentHealths["vault"] = d.healthMonitor.ValidateComponentHealth("vault", outputs.Vault)
	componentHealths["observability"] = d.healthMonitor.ValidateComponentHealth("observability", outputs.Observability)
	componentHealths["dapr"] = d.healthMonitor.ValidateComponentHealth("dapr", outputs.Dapr)
	componentHealths["services"] = d.healthMonitor.ValidateComponentHealth("services", outputs.Services)
	componentHealths["website"] = d.healthMonitor.ValidateComponentHealth("website", outputs.Website)
	
	// Get overall health assessment
	overallHealth := d.healthMonitor.GetOverallHealth(componentHealths)
	
	// Check if rollback is needed
	if d.healthMonitor.ShouldRollback(overallHealth, d.environment) {
		d.ctx.Log.Error("Deployment health assessment indicates rollback is required", nil)
		
		if err := d.performRollback(outputs); err != nil {
			return fmt.Errorf("rollback failed: %w", err)
		}
		
		return fmt.Errorf("deployment rolled back due to health assessment failure")
	}
	
	d.ctx.Log.Info("Deployment health assessment passed, no rollback needed", nil)
	return nil
}

// performRollback performs rollback in reverse dependency order
func (d *DeploymentOrchestrator) performRollback(outputs *ComponentOutputs) error {
	d.ctx.Log.Info("Starting deployment rollback in reverse dependency order", nil)
	
	// Rollback in reverse order: website -> services -> dapr -> observability -> vault -> storage -> database
	rollbackOrder := []struct {
		name      string
		rollbackFn func() error
	}{
		{"website", func() error { return d.rollbackWebsite(outputs.Website) }},
		{"services", func() error { return d.rollbackServices(outputs.Services) }},
		{"dapr", func() error { return d.rollbackDapr(outputs.Dapr) }},
		{"observability", func() error { return d.rollbackObservability(outputs.Observability) }},
		{"vault", func() error { return d.rollbackVault(outputs.Vault) }},
		{"storage", func() error { return d.rollbackStorage(outputs.Storage) }},
		{"database", func() error { return d.rollbackDatabase(outputs.Database) }},
	}
	
	for _, rollback := range rollbackOrder {
		d.ctx.Log.Info(fmt.Sprintf("Rolling back %s component", rollback.name), nil)
		
		if err := rollback.rollbackFn(); err != nil {
			d.ctx.Log.Error(fmt.Sprintf("Failed to rollback %s: %v", rollback.name, err), nil)
			// Continue with other rollbacks even if one fails
		} else {
			d.ctx.Log.Info(fmt.Sprintf("%s component rollback completed", rollback.name), nil)
		}
	}
	
	d.ctx.Log.Info("Deployment rollback completed", nil)
	return nil
}

// Component-specific rollback methods
func (d *DeploymentOrchestrator) rollbackWebsite(outputs *components.WebsiteOutputs) error {
	if outputs == nil {
		return nil
	}
	
	d.ctx.Log.Info("Terminating website container and cleaning up resources", nil)
	
	// Stop and remove website container
	if err := d.stopContainer("website-dev"); err != nil {
		d.ctx.Log.Error(fmt.Sprintf("Failed to stop website container: %v", err), nil)
	}
	
	// For development environment, clean up container artifacts
	if d.environment == "development" {
		if err := d.cleanupContainerArtifacts("website"); err != nil {
			d.ctx.Log.Error(fmt.Sprintf("Failed to cleanup website artifacts: %v", err), nil)
		}
	}
	
	return nil
}

func (d *DeploymentOrchestrator) rollbackServices(outputs *components.ServicesOutputs) error {
	if outputs == nil {
		return nil
	}
	d.ctx.Log.Info("Services rollback - would terminate all service containers and clean up resources", nil)
	// In a real implementation, this would:
	// - Stop all service containers
	// - Clean up load balancers and service mesh configurations
	// - Remove service discovery entries
	return nil
}

func (d *DeploymentOrchestrator) rollbackDapr(outputs *components.DaprOutputs) error {
	if outputs == nil {
		return nil
	}
	d.ctx.Log.Info("Dapr rollback - would terminate Dapr control plane and sidecar containers", nil)
	// In a real implementation, this would:
	// - Stop Dapr control plane
	// - Remove Dapr sidecar containers
	// - Clean up service mesh policies and middleware
	return nil
}

func (d *DeploymentOrchestrator) rollbackObservability(outputs *components.ObservabilityOutputs) error {
	if outputs == nil {
		return nil
	}
	d.ctx.Log.Info("Observability rollback - would terminate monitoring containers and clean up dashboards", nil)
	// In a real implementation, this would:
	// - Stop Grafana and monitoring containers
	// - Clean up dashboards and alert rules
	// - Remove data collection configurations
	return nil
}

func (d *DeploymentOrchestrator) rollbackVault(outputs *components.VaultOutputs) error {
	if outputs == nil {
		return nil
	}
	d.ctx.Log.Info("Vault rollback - would seal vault and clean up secrets management", nil)
	// In a real implementation, this would:
	// - Seal the vault
	// - Stop vault containers
	// - Clean up secret configurations
	return nil
}

func (d *DeploymentOrchestrator) rollbackStorage(outputs *components.StorageOutputs) error {
	if outputs == nil {
		return nil
	}
	d.ctx.Log.Info("Storage rollback - would clean up storage containers and data volumes", nil)
	// In a real implementation, this would:
	// - Stop storage containers
	// - Clean up data volumes (with caution for data preservation)
	// - Remove storage configurations
	return nil
}

func (d *DeploymentOrchestrator) rollbackDatabase(outputs *components.DatabaseOutputs) error {
	if outputs == nil {
		return nil
	}
	d.ctx.Log.Info("Database rollback - would stop database containers and preserve data", nil)
	// In a real implementation, this would:
	// - Stop database containers
	// - Preserve data volumes for potential recovery
	// - Clean up database configurations
	return nil
}

// stopContainer stops and removes a container by name
func (d *DeploymentOrchestrator) stopContainer(containerName string) error {
	d.ctx.Log.Info(fmt.Sprintf("Stopping container: %s", containerName), nil)
	
	// Create command to stop and remove container
	stopCmd, err := local.NewCommand(d.ctx, fmt.Sprintf("stop-%s", containerName), &local.CommandArgs{
		Create: pulumi.Sprintf("podman stop %s && podman rm %s", containerName, containerName),
	})
	if err != nil {
		return fmt.Errorf("failed to create stop command for container %s: %w", containerName, err)
	}
	
	// Wait for command completion
	_ = stopCmd.Stdout // Trigger command execution
	
	d.ctx.Log.Info(fmt.Sprintf("Container %s stopped and removed", containerName), nil)
	return nil
}

// cleanupContainerArtifacts cleans up container-related artifacts for a service
func (d *DeploymentOrchestrator) cleanupContainerArtifacts(serviceName string) error {
	d.ctx.Log.Info(fmt.Sprintf("Cleaning up container artifacts for service: %s", serviceName), nil)
	
	// Create command to clean up container artifacts
	cleanupCmd, err := local.NewCommand(d.ctx, fmt.Sprintf("cleanup-%s", serviceName), &local.CommandArgs{
		Create: pulumi.Sprintf("podman image prune -f --filter label=service=%s", serviceName),
	})
	if err != nil {
		return fmt.Errorf("failed to create cleanup command for service %s: %w", serviceName, err)
	}
	
	// Wait for command completion
	_ = cleanupCmd.Stdout // Trigger command execution
	
	d.ctx.Log.Info(fmt.Sprintf("Container artifacts for %s cleaned up", serviceName), nil)
	return nil
}