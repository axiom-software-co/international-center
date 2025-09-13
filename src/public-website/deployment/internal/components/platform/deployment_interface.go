package platform

import (
	"context"
	"fmt"
	"time"
)

// ContainerProvider defines the interface for container deployment providers
type ContainerProvider interface {
	// Core container operations
	DeployContainer(ctx context.Context, spec *ContainerSpec) error
	StopContainer(ctx context.Context, containerName string) error
	WaitForContainerHealth(ctx context.Context, containerName string, timeout time.Duration) error
	GetContainerLogs(ctx context.Context, containerName string, lines int) (string, error)
	IsContainerRunning(ctx context.Context, containerName string) (bool, error)
	
	// Image management
	PullImage(ctx context.Context, image string) error
	
	// Network and lifecycle management
	Initialize(ctx context.Context) error
	Cleanup(ctx context.Context) error
	ListContainers(ctx context.Context) ([]string, error)
}

// DaprProvider defines the interface for Dapr sidecar management
type DaprProvider interface {
	DeployDaprSidecar(ctx context.Context, spec *ContainerSpec) error
	ValidateDaprConfiguration(ctx context.Context, appID string) error
	GetDaprHealth(ctx context.Context, appID string) error
}

// ContainerSpec defines a unified container specification
type ContainerSpec struct {
	// Basic container configuration
	Name           string
	Image          string
	Port           int
	Command        []string
	Environment    map[string]string
	HealthEndpoint string
	ResourceLimits ResourceLimits
	Volumes        []VolumeMount
	
	// Dapr configuration
	DaprEnabled    bool
	DaprAppID      string
	DaprPort       int
	DaprConfig     map[string]interface{}
	
	// Provider-specific configuration
	PodmanConfig *PodmanContainerConfig
	AzureConfig  *AzureContainerConfig
}

// ResourceLimits defines resource constraints for containers
type ResourceLimits struct {
	CPU           string
	Memory        string
	CPURequest    string
	MemoryRequest string
}

// PodmanContainerConfig contains Podman-specific configuration
type PodmanContainerConfig struct {
	NetworkName    string
	HealthCheck    *HealthCheckConfig
	RestartPolicy  string
	SecurityOpts   []string
}

// AzureContainerConfig contains Azure Container Apps-specific configuration
type AzureContainerConfig struct {
	ResourceGroupName       string
	ContainerEnvironment    string
	ScalingRules           map[string]interface{}
	IngressConfiguration   map[string]interface{}
	TrafficSplitting       map[string]interface{}
	RevisionSuffix         string
}

// HealthCheckConfig defines container health check parameters
type HealthCheckConfig struct {
	Command       string
	Interval      time.Duration
	Timeout       time.Duration
	Retries       int
	StartPeriod   time.Duration
}

// DeploymentResult contains the result of a container deployment
type DeploymentResult struct {
	ContainerID   string
	Status        string
	Endpoint      string
	HealthStatus  string
	Message       string
	Timestamp     time.Time
}

// ContainerStatus represents the current state of a container
type ContainerStatus struct {
	Name          string
	State         string
	Health        string
	RestartCount  int
	StartTime     time.Time
	Image         string
	Ports         []PortMapping
}

// PortMapping represents a port mapping configuration
type PortMapping struct {
	ContainerPort int
	HostPort      int
	Protocol      string
}

// VolumeMount represents a volume mount configuration
type VolumeMount struct {
	HostPath      string
	ContainerPath string
	ReadOnly      bool
}

// BaseContainerProvider provides common functionality for container providers
type BaseContainerProvider struct {
	Environment      string
	ResourceLimits   ResourceLimits
	HealthTimeout    time.Duration
	DeploymentTimeout time.Duration
}

// NewContainerSpec creates a new container specification with defaults
func NewContainerSpec(name, image string, port int) *ContainerSpec {
	return &ContainerSpec{
		Name:           name,
		Image:          image,
		Port:           port,
		Environment:    make(map[string]string),
		HealthEndpoint: fmt.Sprintf("http://localhost:%d/health", port),
		ResourceLimits: ResourceLimits{
			CPU:           "500m",
			Memory:        "256Mi",
			CPURequest:    "100m",
			MemoryRequest: "128Mi",
		},
		DaprEnabled: false,
		DaprConfig:  make(map[string]interface{}),
	}
}

// WithDapr enables Dapr sidecar for the container
func (spec *ContainerSpec) WithDapr(appID string, daprPort int) *ContainerSpec {
	spec.DaprEnabled = true
	spec.DaprAppID = appID
	spec.DaprPort = daprPort
	spec.DaprConfig = map[string]interface{}{
		"app_port":                 spec.Port,
		"placement_host_address":   "localhost:50005",
		"log_level":               "debug",
		"enable_profiling":        true,
		"enable_metrics":          true,
		"metrics_port":            "9090",
	}
	return spec
}

// WithEnvironment adds environment variables to the container
func (spec *ContainerSpec) WithEnvironment(env map[string]string) *ContainerSpec {
	for k, v := range env {
		spec.Environment[k] = v
	}
	return spec
}

// WithResourceLimits sets resource limits for the container
func (spec *ContainerSpec) WithResourceLimits(cpu, memory string) *ContainerSpec {
	spec.ResourceLimits.CPU = cpu
	spec.ResourceLimits.Memory = memory
	return spec
}

// WithPodmanConfig adds Podman-specific configuration
func (spec *ContainerSpec) WithPodmanConfig(config *PodmanContainerConfig) *ContainerSpec {
	spec.PodmanConfig = config
	return spec
}

// WithAzureConfig adds Azure Container Apps-specific configuration
func (spec *ContainerSpec) WithAzureConfig(config *AzureContainerConfig) *ContainerSpec {
	spec.AzureConfig = config
	return spec
}

// Validate checks if the container specification is valid
func (spec *ContainerSpec) Validate() error {
	if spec.Name == "" {
		return fmt.Errorf("container name is required")
	}
	if spec.Image == "" {
		return fmt.Errorf("container image is required")
	}
	if spec.Port <= 0 {
		return fmt.Errorf("valid port number is required")
	}
	if spec.DaprEnabled && spec.DaprAppID == "" {
		return fmt.Errorf("Dapr app ID is required when Dapr is enabled")
	}
	return nil
}

// GetDaprHTTPEndpoint returns the Dapr HTTP endpoint for the container
func (spec *ContainerSpec) GetDaprHTTPEndpoint() string {
	if !spec.DaprEnabled {
		return ""
	}
	return fmt.Sprintf("http://localhost:%d", spec.DaprPort)
}

// GetHealthEndpoint returns the health check endpoint for the container
func (spec *ContainerSpec) GetHealthEndpoint() string {
	if spec.HealthEndpoint != "" {
		return spec.HealthEndpoint
	}
	return fmt.Sprintf("http://localhost:%d/health", spec.Port)
}

// IsServiceContainer determines if this container runs a service (vs infrastructure)
func (spec *ContainerSpec) IsServiceContainer() bool {
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
	return serviceContainers[spec.Name]
}

// Clone creates a deep copy of the container specification
func (spec *ContainerSpec) Clone() *ContainerSpec {
	clone := &ContainerSpec{
		Name:           spec.Name,
		Image:          spec.Image,
		Port:           spec.Port,
		Command:        make([]string, len(spec.Command)),
		Environment:    make(map[string]string),
		HealthEndpoint: spec.HealthEndpoint,
		ResourceLimits: spec.ResourceLimits,
		Volumes:        make([]VolumeMount, len(spec.Volumes)),
		DaprEnabled:    spec.DaprEnabled,
		DaprAppID:      spec.DaprAppID,
		DaprPort:       spec.DaprPort,
		DaprConfig:     make(map[string]interface{}),
	}
	
	// Copy command arguments
	copy(clone.Command, spec.Command)
	
	// Copy volume mounts
	copy(clone.Volumes, spec.Volumes)
	
	// Copy environment variables
	for k, v := range spec.Environment {
		clone.Environment[k] = v
	}
	
	// Copy Dapr config
	for k, v := range spec.DaprConfig {
		clone.DaprConfig[k] = v
	}
	
	// Copy provider-specific configs
	if spec.PodmanConfig != nil {
		clone.PodmanConfig = &PodmanContainerConfig{
			NetworkName:   spec.PodmanConfig.NetworkName,
			RestartPolicy: spec.PodmanConfig.RestartPolicy,
			SecurityOpts:  append([]string{}, spec.PodmanConfig.SecurityOpts...),
		}
		if spec.PodmanConfig.HealthCheck != nil {
			clone.PodmanConfig.HealthCheck = &HealthCheckConfig{
				Command:     spec.PodmanConfig.HealthCheck.Command,
				Interval:    spec.PodmanConfig.HealthCheck.Interval,
				Timeout:     spec.PodmanConfig.HealthCheck.Timeout,
				Retries:     spec.PodmanConfig.HealthCheck.Retries,
				StartPeriod: spec.PodmanConfig.HealthCheck.StartPeriod,
			}
		}
	}
	
	if spec.AzureConfig != nil {
		clone.AzureConfig = &AzureContainerConfig{
			ResourceGroupName:    spec.AzureConfig.ResourceGroupName,
			ContainerEnvironment: spec.AzureConfig.ContainerEnvironment,
			RevisionSuffix:       spec.AzureConfig.RevisionSuffix,
			ScalingRules:        make(map[string]interface{}),
			IngressConfiguration: make(map[string]interface{}),
			TrafficSplitting:    make(map[string]interface{}),
		}
		
		// Copy maps
		for k, v := range spec.AzureConfig.ScalingRules {
			clone.AzureConfig.ScalingRules[k] = v
		}
		for k, v := range spec.AzureConfig.IngressConfiguration {
			clone.AzureConfig.IngressConfiguration[k] = v
		}
		for k, v := range spec.AzureConfig.TrafficSplitting {
			clone.AzureConfig.TrafficSplitting[k] = v
		}
	}
	
	return clone
}

// ContainerSpecBuilder provides a fluent interface for building container specifications
type ContainerSpecBuilder struct {
	spec *ContainerSpec
}

// NewContainerSpecBuilder creates a new builder for container specifications
func NewContainerSpecBuilder(name, image string, port int) *ContainerSpecBuilder {
	return &ContainerSpecBuilder{
		spec: NewContainerSpec(name, image, port),
	}
}

// WithDapr enables Dapr configuration
func (b *ContainerSpecBuilder) WithDapr(appID string, daprPort int) *ContainerSpecBuilder {
	b.spec.WithDapr(appID, daprPort)
	return b
}

// WithEnvironment adds environment variables
func (b *ContainerSpecBuilder) WithEnvironment(env map[string]string) *ContainerSpecBuilder {
	b.spec.WithEnvironment(env)
	return b
}

// WithResourceLimits sets resource limits
func (b *ContainerSpecBuilder) WithResourceLimits(cpu, memory string) *ContainerSpecBuilder {
	b.spec.WithResourceLimits(cpu, memory)
	return b
}

// WithHealthEndpoint sets a custom health endpoint
func (b *ContainerSpecBuilder) WithHealthEndpoint(endpoint string) *ContainerSpecBuilder {
	b.spec.HealthEndpoint = endpoint
	return b
}

// WithCommand sets the container command
func (b *ContainerSpecBuilder) WithCommand(command []string) *ContainerSpecBuilder {
	b.spec.Command = make([]string, len(command))
	copy(b.spec.Command, command)
	return b
}

// WithVolumeMount adds a volume mount
func (b *ContainerSpecBuilder) WithVolumeMount(hostPath, containerPath string, readOnly bool) *ContainerSpecBuilder {
	b.spec.Volumes = append(b.spec.Volumes, VolumeMount{
		HostPath:      hostPath,
		ContainerPath: containerPath,
		ReadOnly:      readOnly,
	})
	return b
}

// Build returns the completed container specification
func (b *ContainerSpecBuilder) Build() (*ContainerSpec, error) {
	if err := b.spec.Validate(); err != nil {
		return nil, fmt.Errorf("invalid container specification: %w", err)
	}
	return b.spec.Clone(), nil
}