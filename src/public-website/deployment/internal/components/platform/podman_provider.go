package platform

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type PodmanProviderArgs struct {
	Environment          string
	NetworkConfiguration pulumi.MapOutput
	ContainerRegistry    pulumi.MapOutput
}

type PodmanProviderComponent struct {
	pulumi.ResourceState

	NetworkName     pulumi.StringOutput `pulumi:"networkName"`
	ContainerImages pulumi.MapOutput    `pulumi:"containerImages"`
	PodmanCommands  pulumi.MapOutput    `pulumi:"podmanCommands"`
	HealthChecker   *UnifiedHealthChecker
	DaprManager     *UnifiedDaprSidecarManager
}

// Ensure PodmanProviderComponent implements ContainerProvider interface
var _ ContainerProvider = (*PodmanProviderComponent)(nil)
var _ DaprProvider = (*PodmanProviderComponent)(nil)
var _ ContainerHealthChecker = (*PodmanProviderComponent)(nil)
var _ DaprSidecarInjector = (*PodmanProviderComponent)(nil)

func NewPodmanProviderComponent(ctx *pulumi.Context, name string, args *PodmanProviderArgs, opts ...pulumi.ResourceOption) (*PodmanProviderComponent, error) {
	component := &PodmanProviderComponent{
		HealthChecker: NewUnifiedHealthChecker(),
		DaprManager:   NewUnifiedDaprSidecarManager(args.Environment),
	}
	
	if canRegister(ctx) {
		err := ctx.RegisterComponentResource("international-center:platform:PodmanProvider", name, component, opts...)
		if err != nil {
			return nil, err
		}
	}

	networkName := pulumi.String("international-center-dev").ToStringOutput()
	
	containerImages := pulumi.Map{
		"dapr_control_plane": pulumi.String("daprio/dapr:latest"),
		"dapr_placement":     pulumi.String("daprio/dapr:latest"),
		"dapr_sentry":        pulumi.String("daprio/dapr:latest"),
		"public_gateway":     pulumi.String("localhost/backend/public-gateway:latest"),
		"admin_gateway":      pulumi.String("localhost/backend/admin-gateway:latest"),
		"content_service":    pulumi.String("localhost/backend/content:latest"),
		"inquiries_service":  pulumi.String("localhost/backend/inquiries:latest"),
		"notifications_service": pulumi.String("localhost/backend/notifications:latest"),
	}.ToMapOutput()

	podmanCommands := pulumi.Map{
		"network_create": pulumi.String("podman network create --driver bridge --subnet 172.20.0.0/16 --gateway 172.20.0.1 international-center-dev"),
		"network_inspect": pulumi.String("podman network inspect international-center-dev"),
		"network_cleanup": pulumi.String("podman network rm international-center-dev"),
	}.ToMapOutput()

	component.NetworkName = networkName
	component.ContainerImages = containerImages
	component.PodmanCommands = podmanCommands

	if canRegister(ctx) {
		if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
			"networkName":     component.NetworkName,
			"containerImages": component.ContainerImages,
			"podmanCommands":  component.PodmanCommands,
		}); err != nil {
			return nil, err
		}
	}

	return component, nil
}

// CreateNetwork creates the Podman network for container communication
func (p *PodmanProviderComponent) CreateNetwork(ctx context.Context) error {
	// Check if network already exists
	checkCmd := exec.CommandContext(ctx, "podman", "network", "inspect", "international-center-dev")
	if err := checkCmd.Run(); err == nil {
		// Network already exists
		return nil
	}

	// Create the network
	createCmd := exec.CommandContext(ctx, "podman", "network", "create",
		"--driver", "bridge",
		"--subnet", "172.20.0.0/16",
		"--gateway", "172.20.0.1",
		"international-center-dev")
	
	if err := createCmd.Run(); err != nil {
		return fmt.Errorf("failed to create Podman network: %w", err)
	}

	return nil
}

// DeployContainer deploys a container using Podman with the specified configuration
func (p *PodmanProviderComponent) DeployContainer(ctx context.Context, spec *ContainerSpec) error {
	// Ensure network exists
	if err := p.CreateNetwork(ctx); err != nil {
		return fmt.Errorf("failed to create network: %w", err)
	}

	// Stop and remove existing container if it exists
	if err := p.StopContainer(ctx, spec.Name); err != nil {
		// Continue even if stopping fails - container might not exist
	}

	// Build Podman run command
	args := []string{"run", "-d", "--name", spec.Name}
	
	// Add network configuration
	networkName := "international-center-dev"
	if spec.PodmanConfig != nil && spec.PodmanConfig.NetworkName != "" {
		networkName = spec.PodmanConfig.NetworkName
	}
	args = append(args, "--network", networkName)

	// Add port mappings
	if spec.Port > 0 {
		args = append(args, "-p", fmt.Sprintf("%d:%d", spec.Port, spec.Port))
	}
	
	if spec.DaprEnabled && spec.DaprPort > 0 {
		args = append(args, "-p", fmt.Sprintf("%d:%d", spec.DaprPort, spec.DaprPort))
	}

	// Add environment variables
	for key, value := range spec.Environment {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	// Add resource limits
	if spec.ResourceLimits.CPU != "" {
		args = append(args, "--cpus", spec.ResourceLimits.CPU)
	}
	
	if spec.ResourceLimits.Memory != "" {
		args = append(args, "--memory", spec.ResourceLimits.Memory)
	}

	// Add health check if specified
	healthEndpoint := spec.GetHealthEndpoint()
	if healthEndpoint != "" {
		healthCheckCmd := fmt.Sprintf("curl -f %s || exit 1", healthEndpoint)
		args = append(args, "--health-cmd", healthCheckCmd)
		args = append(args, "--health-interval", "30s")
		args = append(args, "--health-timeout", "10s")
		args = append(args, "--health-retries", "3")
		args = append(args, "--health-start-period", "60s")
	}

	// Add image
	args = append(args, spec.Image)

	// Execute podman run command
	cmd := exec.CommandContext(ctx, "podman", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to deploy container %s: %w, output: %s", spec.Name, err, output)
	}

	return nil
}

// DeployDaprSidecar deploys a Dapr sidecar container alongside the main container
func (p *PodmanProviderComponent) DeployDaprSidecar(ctx context.Context, spec *ContainerSpec) error {
	// Validate and enrich container spec with Dapr configuration
	if err := p.DaprManager.EnrichContainerSpecWithDapr(spec); err != nil {
		return fmt.Errorf("failed to enrich container spec with Dapr config: %w", err)
	}

	// Use unified manager to inject sidecar
	return p.InjectSidecar(ctx, spec, p.DaprManager.BuildDefaultDaprConfig(spec.DaprAppID, spec.Port))
}

// StopContainer stops and removes a container
func (p *PodmanProviderComponent) StopContainer(ctx context.Context, containerName string) error {
	// Stop container
	stopCmd := exec.CommandContext(ctx, "podman", "stop", containerName)
	stopCmd.Run() // Ignore errors - container might not be running

	// Remove container
	rmCmd := exec.CommandContext(ctx, "podman", "rm", containerName)
	rmCmd.Run() // Ignore errors - container might not exist

	return nil
}

// WaitForContainerHealth waits for a container to become healthy
func (p *PodmanProviderComponent) WaitForContainerHealth(ctx context.Context, containerName string, timeout time.Duration) error {
	return p.HealthChecker.WaitForContainerHealth(ctx, containerName, p, timeout)
}

// GetContainerLogs retrieves logs from a container
func (p *PodmanProviderComponent) GetContainerLogs(ctx context.Context, containerName string, lines int) (string, error) {
	cmd := exec.CommandContext(ctx, "podman", "logs", "--tail", strconv.Itoa(lines), containerName)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get logs for container %s: %w", containerName, err)
	}

	return string(output), nil
}

// IsContainerRunning checks if a container is currently running
func (p *PodmanProviderComponent) IsContainerRunning(ctx context.Context, containerName string) (bool, error) {
	cmd := exec.CommandContext(ctx, "podman", "inspect", "--format", "{{.State.Status}}", containerName)
	output, err := cmd.Output()
	if err != nil {
		return false, nil // Container doesn't exist
	}

	status := strings.TrimSpace(string(output))
	return status == "running", nil
}

// ListContainers lists all containers with the international-center prefix
func (p *PodmanProviderComponent) ListContainers(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "podman", "ps", "-a", "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	containers := []string{}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && (strings.Contains(line, "gateway") || strings.Contains(line, "content") || 
			strings.Contains(line, "inquiries") || strings.Contains(line, "notifications") || 
			strings.Contains(line, "dapr")) {
			containers = append(containers, line)
		}
	}

	return containers, nil
}

// Initialize initializes the container provider (creates network, etc.)
func (p *PodmanProviderComponent) Initialize(ctx context.Context) error {
	return p.CreateNetwork(ctx)
}

// Cleanup performs cleanup of containers and networks
func (p *PodmanProviderComponent) Cleanup(ctx context.Context) error {
	// Stop all containers first
	containers, err := p.ListContainers(ctx)
	if err != nil {
		return fmt.Errorf("failed to list containers for cleanup: %w", err)
	}
	
	for _, containerName := range containers {
		if err := p.StopContainer(ctx, containerName); err != nil {
			// Continue cleanup even if some containers fail to stop
			continue
		}
	}
	
	// Remove network
	return p.CleanupNetwork(ctx)
}

// CleanupNetwork removes the development network
func (p *PodmanProviderComponent) CleanupNetwork(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "podman", "network", "rm", "international-center-dev")
	return cmd.Run() // Ignore errors if network doesn't exist
}

// ValidateDaprConfiguration validates Dapr configuration for an app
func (p *PodmanProviderComponent) ValidateDaprConfiguration(ctx context.Context, appID string) error {
	if appID == "" {
		return fmt.Errorf("Dapr app ID cannot be empty")
	}
	
	// Check if Dapr control plane is running
	running, err := p.IsContainerRunning(ctx, "dapr-control-plane")
	if err != nil {
		return fmt.Errorf("failed to check Dapr control plane status: %w", err)
	}
	
	if !running {
		return fmt.Errorf("Dapr control plane is not running")
	}
	
	return nil
}

// GetDaprHealth checks the health of Dapr sidecar for an app
func (p *PodmanProviderComponent) GetDaprHealth(ctx context.Context, appID string) error {
	// Use unified health checker for Dapr health validation
	daprEndpoint := "http://localhost:3500" // Default Dapr HTTP port
	return p.HealthChecker.ValidateDaprHealth(ctx, appID, daprEndpoint)
}

// ContainerHealthChecker interface implementation

// CheckContainerStatus checks if the container is in a running state
func (p *PodmanProviderComponent) CheckContainerStatus(ctx context.Context, containerName string) (string, error) {
	cmd := exec.CommandContext(ctx, "podman", "inspect", "--format", "{{.State.Status}}", containerName)
	output, err := cmd.Output()
	if err != nil {
		return "not_found", nil // Container doesn't exist
	}

	status := strings.TrimSpace(string(output))
	return status, nil
}

// GetContainerEndpoint returns the health endpoint for a container
func (p *PodmanProviderComponent) GetContainerEndpoint(containerName string) string {
	// Map container names to their health endpoints
	healthEndpoints := map[string]string{
		"dapr-control-plane":   "http://localhost:3500/v1.0/healthz",
		"dapr-placement":       "http://localhost:50005/v1.0/healthz", 
		"dapr-sentry":          "http://localhost:50003/v1.0/healthz",
		"public-gateway":       "http://localhost:9001/health",
		"admin-gateway":        "http://localhost:9000/health",
		"content-news":         "http://localhost:3001/health",
		"content-events":       "http://localhost:3002/health",
		"content-research":     "http://localhost:3003/health",
		"inquiries-business":   "http://localhost:3101/health",
		"inquiries-donations":  "http://localhost:3102/health",
		"inquiries-media":      "http://localhost:3103/health",
		"inquiries-volunteers": "http://localhost:3104/health",
		"notification-service": "http://localhost:3201/health",
	}
	
	return healthEndpoints[containerName]
}

// PullImage pulls a container image if it doesn't exist locally
func (p *PodmanProviderComponent) PullImage(ctx context.Context, image string) error {
	// Check if image exists locally
	checkCmd := exec.CommandContext(ctx, "podman", "image", "inspect", image)
	if err := checkCmd.Run(); err == nil {
		return nil // Image already exists
	}

	// Pull the image
	pullCmd := exec.CommandContext(ctx, "podman", "pull", image)
	output, err := pullCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w, output: %s", image, err, output)
	}

	return nil
}

// DaprSidecarInjector interface implementation

// InjectSidecar injects a Dapr sidecar for the given container specification
func (p *PodmanProviderComponent) InjectSidecar(ctx context.Context, spec *ContainerSpec, config *DaprSidecarConfig) error {
	if err := p.ValidateSidecarConfig(config); err != nil {
		return fmt.Errorf("invalid sidecar configuration: %w", err)
	}

	sidecarName := p.GetSidecarName(config.AppID)
	
	// Stop and remove existing sidecar if it exists
	if err := p.StopContainer(ctx, sidecarName); err != nil {
		// Continue even if stopping fails
	}

	// Build Podman run command for Dapr sidecar
	args := []string{"run", "-d", "--name", sidecarName}
	
	// Add network configuration
	args = append(args, "--network", "international-center-dev")

	// Add Dapr sidecar specific ports
	args = append(args, "-p", fmt.Sprintf("%d:3500", config.DaprHTTPPort))      // Dapr HTTP
	args = append(args, "-p", fmt.Sprintf("%d:50001", config.DaprGRPCPort))     // Dapr GRPC
	args = append(args, "-p", fmt.Sprintf("%d:9090", config.DaprMetricsPort))   // Metrics
	args = append(args, "-p", fmt.Sprintf("%d:7777", config.DaprProfilePort))   // Profile

	// Add resource limits for sidecar
	args = append(args, "--cpus", config.ResourceLimits.CPU)
	args = append(args, "--memory", config.ResourceLimits.Memory)

	// Use Dapr image
	args = append(args, "daprio/dapr:latest")

	// Add Dapr command arguments from unified manager
	daprArgs := p.DaprManager.BuildDaprCommand(config)
	args = append(args, daprArgs...)

	// Execute podman run command for Dapr sidecar
	cmd := exec.CommandContext(ctx, "podman", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to deploy Dapr sidecar for %s: %w, output: %s", config.AppID, err, output)
	}

	return nil
}

// ValidateSidecarConfig validates the Dapr sidecar configuration
func (p *PodmanProviderComponent) ValidateSidecarConfig(config *DaprSidecarConfig) error {
	if config.AppID == "" {
		return fmt.Errorf("Dapr app ID is required")
	}
	
	if config.AppPort <= 0 {
		return fmt.Errorf("valid application port is required")
	}
	
	if config.DaprHTTPPort <= 0 {
		return fmt.Errorf("valid Dapr HTTP port is required")
	}
	
	if config.DaprGRPCPort <= 0 {
		return fmt.Errorf("valid Dapr GRPC port is required")
	}
	
	return nil
}

// GetSidecarName returns the name of the sidecar container
func (p *PodmanProviderComponent) GetSidecarName(appID string) string {
	return fmt.Sprintf("%s-dapr", appID)
}