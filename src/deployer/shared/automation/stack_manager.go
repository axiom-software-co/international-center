package automation

import (
	"context"
	"fmt"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// StackManager manages Pulumi stacks programmatically for CICD workflows
type StackManager struct {
	projectName   string
	workspaceDir  string
	environments  map[string]*EnvironmentConfig
	secretsManager *SecretsManager
	eventHandler  *DeploymentEventHandler
}

// EnvironmentConfig defines configuration for each environment
type EnvironmentConfig struct {
	Name              string
	Description       string
	Config            map[string]string
	Secrets           map[string]string
	AllowDestroy      bool
	RequireApproval   bool
	BackendURL        string
	Tags              map[string]string
	RefreshBeforeDeploy bool
	ParallelDeployment int
}

// DeploymentResult represents the result of a deployment operation
type DeploymentResult struct {
	Environment   string
	StackName     string
	Operation     string
	Status        DeploymentStatus
	Duration      time.Duration
	Summary       *auto.UpdateSummary
	Outputs       auto.OutputMap
	Error         error
	Timestamp     time.Time
	Resources     []ResourceInfo
}

// DeploymentStatus represents the status of a deployment
type DeploymentStatus string

const (
	DeploymentStatusPending   DeploymentStatus = "pending"
	DeploymentStatusRunning   DeploymentStatus = "running"
	DeploymentStatusSucceeded DeploymentStatus = "succeeded"
	DeploymentStatusFailed    DeploymentStatus = "failed"
	DeploymentStatusCancelled DeploymentStatus = "cancelled"
)

// ResourceInfo provides information about deployed resources
type ResourceInfo struct {
	Type     string
	Name     string
	URN      string
	Status   string
	Provider string
}

// NewStackManager creates a new stack manager for CICD workflows
func NewStackManager(projectName, workspaceDir string) *StackManager {
	return &StackManager{
		projectName:   projectName,
		workspaceDir:  workspaceDir,
		environments:  make(map[string]*EnvironmentConfig),
		secretsManager: NewSecretsManager(),
		eventHandler:  NewDeploymentEventHandler(),
	}
}

// RegisterEnvironment registers an environment configuration
func (sm *StackManager) RegisterEnvironment(env *EnvironmentConfig) {
	sm.environments[env.Name] = env
}

// GetEnvironments returns all registered environments
func (sm *StackManager) GetEnvironments() map[string]*EnvironmentConfig {
	return sm.environments
}

// CreateStack creates a new Pulumi stack programmatically
func (sm *StackManager) CreateStack(ctx context.Context, environment string, program pulumi.RunFunc) (*auto.Stack, error) {
	envConfig, exists := sm.environments[environment]
	if !exists {
		return nil, fmt.Errorf("environment %s not registered", environment)
	}

	stackName := fmt.Sprintf("%s-%s", sm.projectName, environment)
	
	// Create workspace and stack
	ws, err := auto.NewLocalWorkspace(ctx,
		auto.Project(workspace.Project{
			Name:        tokens.PackageName(sm.projectName),
			Description: &[]string{fmt.Sprintf("Infrastructure for %s environment", environment)}[0],
			Runtime:     workspace.NewProjectRuntimeInfo("go", nil),
		}),
		auto.WorkDir(sm.workspaceDir),
		auto.Program(program),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	stack, err := auto.UpsertStack(ctx, stackName, ws)
	if err != nil {
		return nil, fmt.Errorf("failed to create/select stack: %w", err)
	}

	// Configure stack
	if err := sm.configureStack(ctx, stack, envConfig); err != nil {
		return nil, fmt.Errorf("failed to configure stack: %w", err)
	}

	sm.eventHandler.OnStackCreated(environment, stackName)
	return &stack, nil
}

// configureStack configures a stack with environment-specific settings
func (sm *StackManager) configureStack(ctx context.Context, stack auto.Stack, envConfig *EnvironmentConfig) error {
	// Set configuration values
	for key, value := range envConfig.Config {
		if err := stack.SetConfig(ctx, key, auto.ConfigValue{Value: value}); err != nil {
			return fmt.Errorf("failed to set config %s: %w", key, err)
		}
	}

	// Set secrets
	for key, value := range envConfig.Secrets {
		if err := stack.SetConfig(ctx, key, auto.ConfigValue{Value: value, Secret: true}); err != nil {
			return fmt.Errorf("failed to set secret %s: %w", key, err)
		}
	}

	// Set backend URL if specified
	if envConfig.BackendURL != "" {
		// Backend URL configuration would be handled during workspace creation
	}

	return nil
}

// Deploy deploys infrastructure to the specified environment
func (sm *StackManager) Deploy(ctx context.Context, environment string, program pulumi.RunFunc) (*DeploymentResult, error) {
	startTime := time.Now()
	result := &DeploymentResult{
		Environment: environment,
		Operation:   "deploy",
		Status:      DeploymentStatusRunning,
		Timestamp:   startTime,
	}

	sm.eventHandler.OnDeploymentStarted(environment, result)

	// Create or get stack
	stack, err := sm.CreateStack(ctx, environment, program)
	if err != nil {
		result.Status = DeploymentStatusFailed
		result.Error = err
		result.Duration = time.Since(startTime)
		sm.eventHandler.OnDeploymentCompleted(environment, result)
		return result, err
	}

	result.StackName = stack.Name()

	// Configure deployment options
	envConfig := sm.environments[environment]
	upOptions := []optup.Option{
		optup.ProgressStreams(sm.eventHandler.GetProgressStreams()),
	}

	if envConfig.RefreshBeforeDeploy {
		upOptions = append(upOptions, optup.Refresh())
	}

	if envConfig.ParallelDeployment > 0 {
		upOptions = append(upOptions, optup.Parallel(envConfig.ParallelDeployment))
	}

	// Perform deployment
	upResult, err := stack.Up(ctx, upOptions...)
	result.Duration = time.Since(startTime)
	result.Summary = &upResult.Summary

	if err != nil {
		result.Status = DeploymentStatusFailed
		result.Error = err
		sm.eventHandler.OnDeploymentCompleted(environment, result)
		return result, err
	}

	// Get outputs
	outputs, err := stack.Outputs(ctx)
	if err != nil {
		result.Status = DeploymentStatusFailed
		result.Error = fmt.Errorf("failed to get outputs: %w", err)
		sm.eventHandler.OnDeploymentCompleted(environment, result)
		return result, result.Error
	}

	result.Outputs = outputs
	result.Status = DeploymentStatusSucceeded
	result.Resources = sm.extractResourceInfo(upResult.Summary)

	sm.eventHandler.OnDeploymentCompleted(environment, result)
	return result, nil
}

// Destroy destroys infrastructure in the specified environment
func (sm *StackManager) Destroy(ctx context.Context, environment string, program pulumi.RunFunc) (*DeploymentResult, error) {
	startTime := time.Now()
	result := &DeploymentResult{
		Environment: environment,
		Operation:   "destroy",
		Status:      DeploymentStatusRunning,
		Timestamp:   startTime,
	}

	envConfig, exists := sm.environments[environment]
	if !exists {
		result.Status = DeploymentStatusFailed
		result.Error = fmt.Errorf("environment %s not registered", environment)
		result.Duration = time.Since(startTime)
		return result, result.Error
	}

	if !envConfig.AllowDestroy {
		result.Status = DeploymentStatusFailed
		result.Error = fmt.Errorf("destroy operation not allowed for environment %s", environment)
		result.Duration = time.Since(startTime)
		return result, result.Error
	}

	sm.eventHandler.OnDestroyStarted(environment, result)

	// Get stack
	stack, err := sm.CreateStack(ctx, environment, program)
	if err != nil {
		result.Status = DeploymentStatusFailed
		result.Error = err
		result.Duration = time.Since(startTime)
		sm.eventHandler.OnDeploymentCompleted(environment, result)
		return result, err
	}

	result.StackName = stack.Name()

	// Perform destroy
	destroyResult, err := stack.Destroy(ctx, 
		optdestroy.ProgressStreams(sm.eventHandler.GetProgressStreams()),
	)
	result.Duration = time.Since(startTime)
	result.Summary = &destroyResult.Summary

	if err != nil {
		result.Status = DeploymentStatusFailed
		result.Error = err
		sm.eventHandler.OnDeploymentCompleted(environment, result)
		return result, err
	}

	result.Status = DeploymentStatusSucceeded
	result.Resources = sm.extractResourceInfo(destroyResult.Summary)

	sm.eventHandler.OnDeploymentCompleted(environment, result)
	return result, nil
}

// GetStackInfo returns information about a specific stack
func (sm *StackManager) GetStackInfo(ctx context.Context, environment string, program pulumi.RunFunc) (*auto.Stack, auto.OutputMap, error) {
	stack, err := sm.CreateStack(ctx, environment, program)
	if err != nil {
		return nil, nil, err
	}

	outputs, err := stack.Outputs(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get outputs: %w", err)
	}

	return stack, outputs, nil
}

// ListStacks lists all available stacks
func (sm *StackManager) ListStacks(ctx context.Context) ([]string, error) {
	var stacks []string
	for env := range sm.environments {
		stackName := fmt.Sprintf("%s-%s", sm.projectName, env)
		stacks = append(stacks, stackName)
	}
	return stacks, nil
}

// extractResourceInfo extracts resource information from update summary
func (sm *StackManager) extractResourceInfo(summary auto.UpdateSummary) []ResourceInfo {
	var resources []ResourceInfo
	
	if summary.ResourceChanges != nil {
		for urn, change := range *summary.ResourceChanges {
			resources = append(resources, ResourceInfo{
				URN:    urn,
				Status: string(change),
				Type:   sm.extractResourceType(urn),
				Name:   sm.extractResourceName(urn),
			})
		}
	}
	
	return resources
}

// extractResourceType extracts resource type from URN
func (sm *StackManager) extractResourceType(urn string) string {
	// URN format: urn:pulumi:stack::project::type::name
	// This would parse the URN to extract the type
	return "resource" // Simplified for now
}

// extractResourceName extracts resource name from URN
func (sm *StackManager) extractResourceName(urn string) string {
	// URN format: urn:pulumi:stack::project::type::name  
	// This would parse the URN to extract the name
	return "resource-name" // Simplified for now
}

// ValidateEnvironmentConfig validates environment configuration
func (sm *StackManager) ValidateEnvironmentConfig(env *EnvironmentConfig) error {
	if env.Name == "" {
		return fmt.Errorf("environment name is required")
	}

	// Validate production-specific requirements
	if env.Name == "production" {
		if env.AllowDestroy {
			return fmt.Errorf("production environment should not allow destroy operations")
		}
		if !env.RequireApproval {
			return fmt.Errorf("production environment must require approval")
		}
	}

	// Validate staging-specific requirements
	if env.Name == "staging" {
		if !env.AllowDestroy {
			return fmt.Errorf("staging environment should allow destroy operations for flexibility")
		}
	}

	return nil
}