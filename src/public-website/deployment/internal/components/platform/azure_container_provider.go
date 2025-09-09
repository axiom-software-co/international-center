package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type AzureContainerProviderArgs struct {
	Environment          string
	NetworkConfiguration pulumi.MapOutput
	ContainerRegistry    pulumi.MapOutput
	ResourceGroupName    string
	ContainerEnvironment string
}

type AzureContainerProviderComponent struct {
	pulumi.ResourceState

	ResourceGroupName       pulumi.StringOutput `pulumi:"resourceGroupName"`
	ContainerEnvironment    pulumi.StringOutput `pulumi:"containerEnvironment"`
	ContainerRegistry       pulumi.StringOutput `pulumi:"containerRegistry"`
	ManagedEnvironmentID    pulumi.StringOutput `pulumi:"managedEnvironmentId"`
	NetworkConfiguration    pulumi.MapOutput    `pulumi:"networkConfiguration"`
	ScalingConfiguration    pulumi.MapOutput    `pulumi:"scalingConfiguration"`
	HealthChecker           *UnifiedHealthChecker
	DaprManager             *UnifiedDaprSidecarManager
}

type AzureContainerAppSpec struct {
	Name                 string
	Image                string
	Port                 int
	AppID                string
	Environment          map[string]string
	HealthEndpoint       string
	ResourceLimits       map[string]interface{}
	ScalingRules         map[string]interface{}
	IngressConfiguration map[string]interface{}
	TrafficSplitting     map[string]interface{}
	DaprConfig          map[string]interface{}
}

type ContainerAppRevision struct {
	Name         string `json:"name"`
	CreatedTime  string `json:"createdTime"`
	Active       bool   `json:"active"`
	TrafficWeight int    `json:"trafficWeight"`
}

// Ensure AzureContainerProviderComponent implements ContainerProvider interface
var _ ContainerProvider = (*AzureContainerProviderComponent)(nil)
var _ DaprProvider = (*AzureContainerProviderComponent)(nil)
var _ ContainerHealthChecker = (*AzureContainerProviderComponent)(nil)
var _ DaprSidecarInjector = (*AzureContainerProviderComponent)(nil)

func NewAzureContainerProviderComponent(ctx *pulumi.Context, name string, args *AzureContainerProviderArgs, opts ...pulumi.ResourceOption) (*AzureContainerProviderComponent, error) {
	component := &AzureContainerProviderComponent{
		HealthChecker: NewUnifiedHealthChecker(),
		DaprManager:   NewUnifiedDaprSidecarManager(args.ContainerEnvironment),
	}
	
	if canRegister(ctx) {
		err := ctx.RegisterComponentResource("international-center:platform:AzureContainerProvider", name, component, opts...)
		if err != nil {
			return nil, err
		}
	}

	resourceGroupName := pulumi.String(args.ResourceGroupName).ToStringOutput()
	containerEnvironment := pulumi.String(args.ContainerEnvironment).ToStringOutput()
	containerRegistry := pulumi.String("registry.azurecr.io").ToStringOutput()

	// Create managed environment ID based on environment
	var managedEnvironmentID pulumi.StringOutput
	switch args.Environment {
	case "staging":
		managedEnvironmentID = pulumi.String(fmt.Sprintf("/subscriptions/{subscription}/resourceGroups/%s/providers/Microsoft.App/managedEnvironments/staging-containerapp-env", args.ResourceGroupName)).ToStringOutput()
	case "production":
		managedEnvironmentID = pulumi.String(fmt.Sprintf("/subscriptions/{subscription}/resourceGroups/%s/providers/Microsoft.App/managedEnvironments/production-containerapp-env", args.ResourceGroupName)).ToStringOutput()
	default:
		managedEnvironmentID = pulumi.String("").ToStringOutput()
	}

	// Scaling configuration based on environment
	var scalingConfiguration pulumi.MapOutput
	switch args.Environment {
	case "staging":
		scalingConfiguration = pulumi.Map{
			"min_replicas": pulumi.Int(1),
			"max_replicas": pulumi.Int(10),
			"scale_rules": pulumi.Map{
				"http_requests": pulumi.Map{
					"concurrent_requests": pulumi.Int(10),
				},
				"cpu_utilization": pulumi.Map{
					"utilization": pulumi.Int(70),
				},
			},
		}.ToMapOutput()
	case "production":
		scalingConfiguration = pulumi.Map{
			"min_replicas": pulumi.Int(3),
			"max_replicas": pulumi.Int(50),
			"scale_rules": pulumi.Map{
				"http_requests": pulumi.Map{
					"concurrent_requests": pulumi.Int(100),
				},
				"cpu_utilization": pulumi.Map{
					"utilization": pulumi.Int(80),
				},
			},
		}.ToMapOutput()
	default:
		scalingConfiguration = pulumi.Map{}.ToMapOutput()
	}

	component.ResourceGroupName = resourceGroupName
	component.ContainerEnvironment = containerEnvironment
	component.ContainerRegistry = containerRegistry
	component.ManagedEnvironmentID = managedEnvironmentID
	component.NetworkConfiguration = args.NetworkConfiguration
	component.ScalingConfiguration = scalingConfiguration

	if canRegister(ctx) {
		if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
			"resourceGroupName":     component.ResourceGroupName,
			"containerEnvironment":  component.ContainerEnvironment,
			"containerRegistry":     component.ContainerRegistry,
			"managedEnvironmentId":  component.ManagedEnvironmentID,
			"networkConfiguration":  component.NetworkConfiguration,
			"scalingConfiguration":  component.ScalingConfiguration,
		}); err != nil {
			return nil, err
		}
	}

	return component, nil
}

// DeployContainerApp deploys a container app to Azure Container Apps
func (a *AzureContainerProviderComponent) DeployContainerApp(ctx context.Context, spec *AzureContainerAppSpec) error {
	// Create container app configuration
	containerAppConfig := a.buildContainerAppConfiguration(spec)

	// Convert configuration to JSON for Azure CLI
	configJSON, err := json.Marshal(containerAppConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal container app config: %w", err)
	}

	// Write configuration to temporary file
	configFile := fmt.Sprintf("/tmp/%s-config.json", spec.Name)
	if err := a.writeConfigFile(configFile, string(configJSON)); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Deploy using Azure CLI
	cmd := exec.CommandContext(ctx, "az", "containerapp", "create",
		"--name", spec.Name,
		"--resource-group", a.getResourceGroupName(),
		"--environment", a.getContainerEnvironment(),
		"--yaml", configFile,
		"--output", "json")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to deploy container app %s: %w, output: %s", spec.Name, err, output)
	}

	return nil
}

// UpdateContainerAppRevision creates a new revision for blue-green or canary deployment
func (a *AzureContainerProviderComponent) UpdateContainerAppRevision(ctx context.Context, spec *AzureContainerAppSpec, trafficPercentage int) error {
	revisionName := fmt.Sprintf("%s-%d", spec.Name, time.Now().Unix())
	
	// Create new revision
	containerAppConfig := a.buildContainerAppConfiguration(spec)
	containerAppConfig["revision"] = map[string]interface{}{
		"suffix": fmt.Sprintf("%d", time.Now().Unix()),
	}

	configJSON, err := json.Marshal(containerAppConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal container app config: %w", err)
	}

	configFile := fmt.Sprintf("/tmp/%s-revision-config.json", spec.Name)
	if err := a.writeConfigFile(configFile, string(configJSON)); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Update container app with new revision
	cmd := exec.CommandContext(ctx, "az", "containerapp", "revision", "copy",
		"--name", spec.Name,
		"--resource-group", a.getResourceGroupName(),
		"--from-revision", "latest",
		"--yaml", configFile,
		"--output", "json")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create revision for %s: %w, output: %s", spec.Name, err, output)
	}

	// Configure traffic splitting
	if err := a.ConfigureTrafficSplitting(ctx, spec.Name, revisionName, trafficPercentage); err != nil {
		return fmt.Errorf("failed to configure traffic splitting: %w", err)
	}

	return nil
}

// ConfigureTrafficSplitting configures traffic splitting between revisions
func (a *AzureContainerProviderComponent) ConfigureTrafficSplitting(ctx context.Context, appName, newRevision string, newTrafficPercentage int) error {
	// Get current revisions
	revisions, err := a.GetActiveRevisions(ctx, appName)
	if err != nil {
		return fmt.Errorf("failed to get active revisions: %w", err)
	}

	// Build traffic configuration
	trafficConfig := []map[string]interface{}{}
	
	// Add new revision with specified percentage
	trafficConfig = append(trafficConfig, map[string]interface{}{
		"revisionName": newRevision,
		"weight":       newTrafficPercentage,
	})

	// Distribute remaining traffic among other active revisions
	remainingTraffic := 100 - newTrafficPercentage
	activeCount := len(revisions) - 1 // Exclude the new revision
	
	if activeCount > 0 {
		weightPerRevision := remainingTraffic / activeCount
		for _, revision := range revisions {
			if revision.Name != newRevision {
				trafficConfig = append(trafficConfig, map[string]interface{}{
					"revisionName": revision.Name,
					"weight":       weightPerRevision,
				})
			}
		}
	}

	// Apply traffic configuration
	trafficJSON, err := json.Marshal(trafficConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal traffic config: %w", err)
	}

	cmd := exec.CommandContext(ctx, "az", "containerapp", "ingress", "traffic", "set",
		"--name", appName,
		"--resource-group", a.getResourceGroupName(),
		"--revision-weight", string(trafficJSON),
		"--output", "json")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to configure traffic splitting: %w, output: %s", err, output)
	}

	return nil
}

// GetActiveRevisions retrieves active revisions for a container app
func (a *AzureContainerProviderComponent) GetActiveRevisions(ctx context.Context, appName string) ([]ContainerAppRevision, error) {
	cmd := exec.CommandContext(ctx, "az", "containerapp", "revision", "list",
		"--name", appName,
		"--resource-group", a.getResourceGroupName(),
		"--output", "json")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list revisions: %w", err)
	}

	var revisions []ContainerAppRevision
	if err := json.Unmarshal(output, &revisions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal revisions: %w", err)
	}

	// Filter active revisions
	activeRevisions := []ContainerAppRevision{}
	for _, revision := range revisions {
		if revision.Active {
			activeRevisions = append(activeRevisions, revision)
		}
	}

	return activeRevisions, nil
}

// WaitForRevisionReady waits for a container app revision to be ready
func (a *AzureContainerProviderComponent) WaitForRevisionReady(ctx context.Context, appName, revisionName string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for revision %s to be ready", revisionName)
		case <-ticker.C:
			cmd := exec.CommandContext(ctx, "az", "containerapp", "revision", "show",
				"--name", appName,
				"--resource-group", a.getResourceGroupName(),
				"--revision", revisionName,
				"--query", "properties.provisioningState",
				"--output", "tsv")

			output, err := cmd.Output()
			if err != nil {
				continue // Revision might not be ready yet
			}

			state := strings.TrimSpace(string(output))
			switch state {
			case "Succeeded":
				return nil
			case "Failed":
				return fmt.Errorf("revision %s failed to deploy", revisionName)
			default:
				continue // Still provisioning
			}
		}
	}
}

// WaitForContainerAppHealth waits for a container app to be healthy
func (a *AzureContainerProviderComponent) WaitForContainerAppHealth(ctx context.Context, appName string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for container app %s to be healthy", appName)
		case <-ticker.C:
			// Check container app status
			cmd := exec.CommandContext(ctx, "az", "containerapp", "show",
				"--name", appName,
				"--resource-group", a.getResourceGroupName(),
				"--query", "properties.provisioningState",
				"--output", "tsv")

			output, err := cmd.Output()
			if err != nil {
				continue
			}

			state := strings.TrimSpace(string(output))
			switch state {
			case "Succeeded":
				// Additional health check via HTTP if health endpoint is available
				return a.validateContainerAppHealth(ctx, appName)
			case "Failed":
				return fmt.Errorf("container app %s failed to deploy", appName)
			default:
				continue // Still provisioning
			}
		}
	}
}

// validateContainerAppHealth performs HTTP health check on the container app
func (a *AzureContainerProviderComponent) validateContainerAppHealth(ctx context.Context, appName string) error {
	// Get container app FQDN
	cmd := exec.CommandContext(ctx, "az", "containerapp", "show",
		"--name", appName,
		"--resource-group", a.getResourceGroupName(),
		"--query", "properties.configuration.ingress.fqdn",
		"--output", "tsv")

	output, err := cmd.Output()
	if err != nil {
		return nil // No ingress configured, assume healthy if provisioned
	}

	fqdn := strings.TrimSpace(string(output))
	if fqdn == "" {
		return nil // No FQDN available
	}

	// Perform health check
	healthURL := fmt.Sprintf("https://%s/health", fqdn)
	healthCmd := exec.CommandContext(ctx, "curl", "-f", "-s", "--connect-timeout", "10", healthURL)
	if err := healthCmd.Run(); err != nil {
		return fmt.Errorf("health check failed for %s at %s: %w", appName, healthURL, err)
	}

	return nil
}

// DeleteRevision deletes a specific revision
func (a *AzureContainerProviderComponent) DeleteRevision(ctx context.Context, appName, revisionName string) error {
	cmd := exec.CommandContext(ctx, "az", "containerapp", "revision", "deactivate",
		"--name", appName,
		"--resource-group", a.getResourceGroupName(),
		"--revision", revisionName)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to deactivate revision %s: %w, output: %s", revisionName, err, output)
	}

	return nil
}

// GetContainerAppLogs retrieves logs from a container app
func (a *AzureContainerProviderComponent) GetContainerAppLogs(ctx context.Context, appName string, lines int) (string, error) {
	cmd := exec.CommandContext(ctx, "az", "containerapp", "logs", "show",
		"--name", appName,
		"--resource-group", a.getResourceGroupName(),
		"--tail", strconv.Itoa(lines),
		"--output", "table")

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get logs for %s: %w", appName, err)
	}

	return string(output), nil
}

// buildContainerAppConfiguration builds the Azure Container App configuration
func (a *AzureContainerProviderComponent) buildContainerAppConfiguration(spec *AzureContainerAppSpec) map[string]interface{} {
	config := map[string]interface{}{
		"properties": map[string]interface{}{
			"configuration": map[string]interface{}{
				"activeRevisionsMode": "Multiple",
				"ingress": map[string]interface{}{
					"external":   true,
					"targetPort": spec.Port,
					"traffic": []map[string]interface{}{
						{
							"weight":        100,
							"latestRevision": true,
						},
					},
				},
				"dapr": map[string]interface{}{
					"enabled": true,
					"appId":   spec.AppID,
					"appPort": spec.Port,
				},
			},
			"template": map[string]interface{}{
				"containers": []map[string]interface{}{
					{
						"name":  spec.Name,
						"image": spec.Image,
						"resources": spec.ResourceLimits,
						"env":     a.buildEnvironmentVariables(spec.Environment),
					},
				},
				"scale": a.buildScaleConfiguration(spec.ScalingRules),
			},
		},
	}

	// Add traffic splitting if configured
	if spec.TrafficSplitting != nil {
		config["properties"].(map[string]interface{})["configuration"].(map[string]interface{})["ingress"].(map[string]interface{})["traffic"] = spec.TrafficSplitting
	}

	return config
}

// buildEnvironmentVariables builds environment variables for Azure Container Apps
func (a *AzureContainerProviderComponent) buildEnvironmentVariables(env map[string]string) []map[string]interface{} {
	envVars := []map[string]interface{}{}
	for key, value := range env {
		envVars = append(envVars, map[string]interface{}{
			"name":  key,
			"value": value,
		})
	}
	return envVars
}

// buildScaleConfiguration builds scaling configuration for Azure Container Apps
func (a *AzureContainerProviderComponent) buildScaleConfiguration(scalingRules map[string]interface{}) map[string]interface{} {
	if scalingRules == nil {
		// Default scaling configuration
		return map[string]interface{}{
			"minReplicas": 1,
			"maxReplicas": 10,
		}
	}

	scaleConfig := map[string]interface{}{
		"minReplicas": 1,
		"maxReplicas": 10,
		"rules":       []map[string]interface{}{},
	}

	// Add HTTP scaling rule if configured
	if httpRule, exists := scalingRules["http_requests"]; exists {
		if httpConfig, ok := httpRule.(map[string]interface{}); ok {
			if concurrentReqs, ok := httpConfig["concurrent_requests"].(int); ok {
				rule := map[string]interface{}{
					"name": "http-scale-rule",
					"http": map[string]interface{}{
						"metadata": map[string]interface{}{
							"concurrentRequests": fmt.Sprintf("%d", concurrentReqs),
						},
					},
				}
				scaleConfig["rules"] = append(scaleConfig["rules"].([]map[string]interface{}), rule)
			}
		}
	}

	return scaleConfig
}

// writeConfigFile writes configuration to a temporary file
func (a *AzureContainerProviderComponent) writeConfigFile(filename, content string) error {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("cat > %s << 'EOF'\n%s\nEOF", filename, content))
	return cmd.Run()
}

// Helper methods to get configuration values
func (a *AzureContainerProviderComponent) getResourceGroupName() string {
	// This would resolve the Pulumi output in actual deployment
	return "international-center-rg" // Default for now
}

func (a *AzureContainerProviderComponent) getContainerEnvironment() string {
	// This would resolve the Pulumi output in actual deployment
	return "international-center-env" // Default for now
}

// ContainerProvider interface implementation

// DeployContainer deploys a container using Azure Container Apps
func (a *AzureContainerProviderComponent) DeployContainer(ctx context.Context, spec *ContainerSpec) error {
	azureSpec := a.convertToAzureSpec(spec)
	return a.DeployContainerApp(ctx, azureSpec)
}

// StopContainer stops a container app by scaling it to zero
func (a *AzureContainerProviderComponent) StopContainer(ctx context.Context, containerName string) error {
	cmd := exec.CommandContext(ctx, "az", "containerapp", "update",
		"--name", containerName,
		"--resource-group", a.getResourceGroupName(),
		"--min-replicas", "0",
		"--max-replicas", "0",
		"--output", "json")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop container %s: %w, output: %s", containerName, err, output)
	}

	return nil
}

// WaitForContainerHealth waits for a container app to be healthy
func (a *AzureContainerProviderComponent) WaitForContainerHealth(ctx context.Context, containerName string, timeout time.Duration) error {
	return a.HealthChecker.WaitForContainerHealth(ctx, containerName, a, timeout)
}

// GetContainerLogs retrieves logs from a container app
func (a *AzureContainerProviderComponent) GetContainerLogs(ctx context.Context, containerName string, lines int) (string, error) {
	return a.GetContainerAppLogs(ctx, containerName, lines)
}

// IsContainerRunning checks if a container app is running
func (a *AzureContainerProviderComponent) IsContainerRunning(ctx context.Context, containerName string) (bool, error) {
	cmd := exec.CommandContext(ctx, "az", "containerapp", "show",
		"--name", containerName,
		"--resource-group", a.getResourceGroupName(),
		"--query", "properties.provisioningState",
		"--output", "tsv")

	output, err := cmd.Output()
	if err != nil {
		return false, nil // Container doesn't exist
	}

	state := strings.TrimSpace(string(output))
	return state == "Succeeded", nil
}

// PullImage pulls an image (not needed for Azure Container Apps as they pull automatically)
func (a *AzureContainerProviderComponent) PullImage(ctx context.Context, image string) error {
	// Azure Container Apps automatically pull images during deployment
	return nil
}

// Initialize initializes the Azure Container Apps environment
func (a *AzureContainerProviderComponent) Initialize(ctx context.Context) error {
	// Azure Container Apps environment should already be created by infrastructure
	// Validate that the environment exists
	cmd := exec.CommandContext(ctx, "az", "containerapp", "env", "show",
		"--name", a.getContainerEnvironment(),
		"--resource-group", a.getResourceGroupName(),
		"--output", "json")

	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Azure Container Apps environment %s not found: %w", a.getContainerEnvironment(), err)
	}

	return nil
}

// Cleanup performs cleanup of container apps
func (a *AzureContainerProviderComponent) Cleanup(ctx context.Context) error {
	// List all container apps
	containers, err := a.ListContainers(ctx)
	if err != nil {
		return fmt.Errorf("failed to list containers for cleanup: %w", err)
	}

	// Stop all container apps
	for _, containerName := range containers {
		if err := a.StopContainer(ctx, containerName); err != nil {
			// Continue cleanup even if some containers fail to stop
			continue
		}
	}

	return nil
}

// ListContainers lists all container apps in the resource group
func (a *AzureContainerProviderComponent) ListContainers(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "az", "containerapp", "list",
		"--resource-group", a.getResourceGroupName(),
		"--query", "[].name",
		"--output", "tsv")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list container apps: %w", err)
	}

	containers := []string{}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			containers = append(containers, line)
		}
	}

	return containers, nil
}

// DaprProvider interface implementation

// DeployDaprSidecar is handled automatically by Azure Container Apps Dapr extension
func (a *AzureContainerProviderComponent) DeployDaprSidecar(ctx context.Context, spec *ContainerSpec) error {
	// Validate and enrich container spec with Dapr configuration using unified manager
	if err := a.DaprManager.EnrichContainerSpecWithDapr(spec); err != nil {
		return fmt.Errorf("failed to enrich container spec with Dapr config: %w", err)
	}

	// Azure Container Apps handles Dapr sidecar deployment automatically
	// when Dapr is enabled in the container app configuration
	// The sidecar injection is handled by the platform, not manually deployed
	return nil
}

// ValidateDaprConfiguration validates Dapr configuration for Azure Container Apps
func (a *AzureContainerProviderComponent) ValidateDaprConfiguration(ctx context.Context, appID string) error {
	if appID == "" {
		return fmt.Errorf("Dapr app ID cannot be empty")
	}

	// Validate that the container app has Dapr enabled
	cmd := exec.CommandContext(ctx, "az", "containerapp", "show",
		"--name", appID,
		"--resource-group", a.getResourceGroupName(),
		"--query", "properties.configuration.dapr.enabled",
		"--output", "tsv")

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check Dapr configuration for %s: %w", appID, err)
	}

	enabled := strings.TrimSpace(string(output))
	if enabled != "true" {
		return fmt.Errorf("Dapr is not enabled for container app %s", appID)
	}

	return nil
}

// GetDaprHealth checks the health of Dapr sidecar via Azure Container Apps
func (a *AzureContainerProviderComponent) GetDaprHealth(ctx context.Context, appID string) error {
	// Get the container app's Dapr endpoint
	daprEndpoint := a.GetContainerEndpoint(appID)
	if daprEndpoint == "" {
		return fmt.Errorf("no Dapr endpoint available for %s", appID)
	}
	
	// Extract base URL for Dapr health check
	if strings.Contains(daprEndpoint, "/health") {
		daprEndpoint = strings.Replace(daprEndpoint, "/health", "", 1)
	}
	
	return a.HealthChecker.ValidateDaprHealth(ctx, appID, daprEndpoint)
}

// ContainerHealthChecker interface implementation

// CheckContainerStatus checks if the container app is in a running state
func (a *AzureContainerProviderComponent) CheckContainerStatus(ctx context.Context, containerName string) (string, error) {
	cmd := exec.CommandContext(ctx, "az", "containerapp", "show",
		"--name", containerName,
		"--resource-group", a.getResourceGroupName(),
		"--query", "properties.provisioningState",
		"--output", "tsv")

	output, err := cmd.Output()
	if err != nil {
		return "not_found", nil // Container app doesn't exist
	}

	status := strings.TrimSpace(string(output))
	return status, nil
}

// GetContainerEndpoint returns the health endpoint for a container app
func (a *AzureContainerProviderComponent) GetContainerEndpoint(containerName string) string {
	// Get container app FQDN for health check
	cmd := exec.Command("az", "containerapp", "show",
		"--name", containerName,
		"--resource-group", a.getResourceGroupName(),
		"--query", "properties.configuration.ingress.fqdn",
		"--output", "tsv")

	output, err := cmd.Output()
	if err != nil {
		return "" // No ingress configured
	}

	fqdn := strings.TrimSpace(string(output))
	if fqdn == "" {
		return "" // No FQDN available
	}

	// Map container names to their health endpoints
	healthPaths := map[string]string{
		"dapr-control-plane":   "/v1.0/healthz",
		"dapr-placement":       "/v1.0/healthz",
		"dapr-sentry":          "/v1.0/healthz",
		"public-gateway":       "/health",
		"admin-gateway":        "/health",
		"content-news":         "/health",
		"content-events":       "/health",
		"content-research":     "/health",
		"inquiries-business":   "/health",
		"inquiries-donations":  "/health",
		"inquiries-media":      "/health",
		"inquiries-volunteers": "/health",
		"notification-service": "/health",
	}

	if path, exists := healthPaths[containerName]; exists {
		return fmt.Sprintf("https://%s%s", fqdn, path)
	}

	return fmt.Sprintf("https://%s/health", fqdn)
}

// convertToAzureSpec converts unified ContainerSpec to Azure-specific spec
func (a *AzureContainerProviderComponent) convertToAzureSpec(spec *ContainerSpec) *AzureContainerAppSpec {
	azureSpec := &AzureContainerAppSpec{
		Name:           spec.Name,
		Image:          spec.Image,
		Port:           spec.Port,
		AppID:          spec.DaprAppID,
		Environment:    spec.Environment,
		HealthEndpoint: spec.GetHealthEndpoint(),
		ResourceLimits: map[string]interface{}{
			"cpu":    spec.ResourceLimits.CPU,
			"memory": spec.ResourceLimits.Memory,
		},
		DaprConfig: spec.DaprConfig,
	}

	// Add Azure-specific configuration if present
	if spec.AzureConfig != nil {
		azureSpec.ScalingRules = spec.AzureConfig.ScalingRules
		azureSpec.IngressConfiguration = spec.AzureConfig.IngressConfiguration
		azureSpec.TrafficSplitting = spec.AzureConfig.TrafficSplitting
	}

	return azureSpec
}

// DaprSidecarInjector interface implementation

// InjectSidecar for Azure Container Apps - handled automatically by the platform
func (a *AzureContainerProviderComponent) InjectSidecar(ctx context.Context, spec *ContainerSpec, config *DaprSidecarConfig) error {
	if err := a.ValidateSidecarConfig(config); err != nil {
		return fmt.Errorf("invalid sidecar configuration: %w", err)
	}

	// Azure Container Apps automatically injects Dapr sidecars when Dapr is enabled
	// No manual sidecar deployment is required - this is handled by the platform
	return nil
}

// ValidateSidecarConfig validates the Dapr sidecar configuration for Azure Container Apps
func (a *AzureContainerProviderComponent) ValidateSidecarConfig(config *DaprSidecarConfig) error {
	if config.AppID == "" {
		return fmt.Errorf("Dapr app ID is required")
	}
	
	if config.AppPort <= 0 {
		return fmt.Errorf("valid application port is required")
	}
	
	// Azure Container Apps has different port requirements than Podman
	// but we still validate basic configuration consistency
	return nil
}

// GetSidecarName returns the name of the sidecar container (Azure handles naming automatically)
func (a *AzureContainerProviderComponent) GetSidecarName(appID string) string {
	// Azure Container Apps automatically names Dapr sidecars
	// This follows their internal naming convention
	return fmt.Sprintf("%s--dapr", appID)
}