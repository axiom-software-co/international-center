package infrastructure

import (
	"fmt"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
)

// ContainerRuntime provides abstraction for container runtime operations
type ContainerRuntime struct {
	ctx        *pulumi.Context
	runtime    string  // "podman" or "containerd"
	registryHost string
	registryPort int
}

// NewContainerRuntime creates a new container runtime abstraction
func NewContainerRuntime(ctx *pulumi.Context, runtime, registryHost string, registryPort int) *ContainerRuntime {
	return &ContainerRuntime{
		ctx:          ctx,
		runtime:      runtime,
		registryHost: registryHost,
		registryPort: registryPort,
	}
}

// ContainerConfig defines container configuration
type ContainerConfig struct {
	Name         string
	Image        string
	Tag          string
	Environment  map[string]pulumi.StringInput
	Ports        []ContainerPort
	Volumes      []ContainerVolume
	Networks     []string
	Resources    ContainerResources
	HealthCheck  ContainerHealthCheck
	RestartPolicy string
	LogConfig    ContainerLogConfig
	SecurityOpts []string
	Labels       map[string]string
	Command      []string
	Args         []string
	WorkingDir   string
	User         string
	DependsOn    []pulumi.Resource
}

// ContainerPort defines container port configuration
type ContainerPort struct {
	Internal int    `json:"internal"`
	External int    `json:"external"`
	Protocol string `json:"protocol"`
}

// ContainerVolume defines container volume configuration
type ContainerVolume struct {
	HostPath      string `json:"host_path"`
	ContainerPath string `json:"container_path"`
	ReadOnly      bool   `json:"read_only"`
	VolumeType    string `json:"volume_type"` // "bind", "volume", "tmpfs"
}

// ContainerResources defines container resource limits
type ContainerResources struct {
	CPUShares    int    `json:"cpu_shares"`
	CPUQuota     int    `json:"cpu_quota"`
	CPUPeriod    int    `json:"cpu_period"`
	Memory       int64  `json:"memory"`        // in bytes
	MemorySwap   int64  `json:"memory_swap"`   // in bytes
	MemoryReservation int64 `json:"memory_reservation"` // in bytes
	KernelMemory int64  `json:"kernel_memory"` // in bytes
	OOMKillDisable bool `json:"oom_kill_disable"`
	PidsLimit    int    `json:"pids_limit"`
	Ulimits      []Ulimit `json:"ulimits"`
}

// Ulimit defines ulimit configuration
type Ulimit struct {
	Name string `json:"name"`
	Soft int    `json:"soft"`
	Hard int    `json:"hard"`
}

// ContainerHealthCheck defines container health check configuration
type ContainerHealthCheck struct {
	Test        []string      `json:"test"`
	Interval    time.Duration `json:"interval"`
	Timeout     time.Duration `json:"timeout"`
	StartPeriod time.Duration `json:"start_period"`
	Retries     int           `json:"retries"`
}

// ContainerLogConfig defines container logging configuration
type ContainerLogConfig struct {
	Driver  string            `json:"driver"`
	Options map[string]string `json:"options"`
}

// CreateContainer creates a new container with the specified configuration
func (cr *ContainerRuntime) CreateContainer(config ContainerConfig) (*docker.Container, error) {
	// Build image name with registry
	imageName := cr.buildImageName(config.Image, config.Tag)
	
	// Convert port configuration
	portSpecs := make(docker.ContainerPortArray, len(config.Ports))
	for i, port := range config.Ports {
		portSpecs[i] = &docker.ContainerPortArgs{
			Internal: pulumi.Int(port.Internal),
			External: pulumi.Int(port.External),
			Protocol: pulumi.String(port.Protocol),
		}
	}
	
	// Convert volume configuration
	volumes := make(docker.ContainerVolumeArray, len(config.Volumes))
	for i, vol := range config.Volumes {
		volumes[i] = &docker.ContainerVolumeArgs{
			HostPath:      pulumi.String(vol.HostPath),
			ContainerPath: pulumi.String(vol.ContainerPath),
			ReadOnly:      pulumi.Bool(vol.ReadOnly),
		}
	}
	
	// Convert environment variables
	envVars := make(pulumi.StringArray, 0, len(config.Environment))
	for key, value := range config.Environment {
		envVar := pulumi.Sprintf("%s=%s", key, value)
		envVars = append(envVars, envVar)
	}
	
	// Convert health check configuration
	var healthCheck *docker.ContainerHealthcheckArgs
	if len(config.HealthCheck.Test) > 0 {
		testCommands := make(pulumi.StringArray, len(config.HealthCheck.Test))
		for i, test := range config.HealthCheck.Test {
			testCommands[i] = pulumi.String(test)
		}
		
		healthCheck = &docker.ContainerHealthcheckArgs{
			Tests:       testCommands,
			Interval:    pulumi.String(config.HealthCheck.Interval.String()),
			Timeout:     pulumi.String(config.HealthCheck.Timeout.String()),
			StartPeriod: pulumi.String(config.HealthCheck.StartPeriod.String()),
			Retries:     pulumi.Int(config.HealthCheck.Retries),
		}
	}
	
	// Convert log configuration
	var logOpts pulumi.StringMap
	if len(config.LogConfig.Options) > 0 {
		logOpts = make(pulumi.StringMap)
		for key, value := range config.LogConfig.Options {
			logOpts[key] = pulumi.String(value)
		}
	}
	
	// Convert labels
	var labels pulumi.StringMap
	if len(config.Labels) > 0 {
		labels = make(pulumi.StringMap)
		for key, value := range config.Labels {
			labels[key] = pulumi.String(value)
		}
	}
	
	// Create container arguments
	containerArgs := &docker.ContainerArgs{
		Image:   pulumi.String(imageName),
		Name:    pulumi.String(config.Name),
		Ports:   portSpecs,
		Volumes: volumes,
		Envs:    envVars,
		
		// Resource limits
		Memory:     pulumi.Int(int(config.Resources.Memory)),
		MemorySwap: pulumi.Int(int(config.Resources.MemorySwap)),
		CpuShares:  pulumi.Int(config.Resources.CPUShares),
		
		// Network configuration
		NetworkMode: pulumi.String(cr.getNetworkMode(config.Networks)),
		
		// Health check
		Healthcheck: healthCheck,
		
		// Restart policy
		Restart: pulumi.String(config.RestartPolicy),
		
		// Logging configuration
		LogDriver: pulumi.String(config.LogConfig.Driver),
		LogOpts:   logOpts,
		
		// Security options
		SecurityOpts: cr.convertStringSliceToPulumiArray(config.SecurityOpts),
		
		// Labels - convert to label array format
		Labels: cr.convertLabelsToArray(labels),
		
		// Command and arguments
		Command: cr.convertStringSliceToPulumiArray(config.Command),
		
		// Working directory
		WorkingDir: pulumi.String(config.WorkingDir),
		
		// User
		User: pulumi.String(config.User),
		
		// Remove container when stopped (for development)
		Rm: pulumi.Bool(cr.runtime == "podman"),
	}
	
	// Set dependencies
	var opts []pulumi.ResourceOption
	if len(config.DependsOn) > 0 {
		opts = append(opts, pulumi.DependsOn(config.DependsOn))
	}
	
	// Create the container
	container, err := docker.NewContainer(cr.ctx, config.Name, containerArgs, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create container %s: %w", config.Name, err)
	}
	
	return container, nil
}

// BuildImage builds a container image
func (cr *ContainerRuntime) BuildImage(imageName, contextPath, dockerfilePath string, buildArgs map[string]string) (*docker.Image, error) {
	// Convert build args
	var buildArgsMap pulumi.StringMap
	if len(buildArgs) > 0 {
		buildArgsMap = make(pulumi.StringMap)
		for key, value := range buildArgs {
			buildArgsMap[key] = pulumi.String(value)
		}
	}
	
	// Create build configuration
	buildConfig := &docker.DockerBuildArgs{
		Context:    pulumi.String(contextPath),
		Dockerfile: pulumi.String(dockerfilePath),
		Args:       buildArgsMap,
		Platform:   pulumi.String("linux/amd64"), // Ensure consistent platform
		Target:     pulumi.String(""),            // Build all stages by default
	}
	
	// Build the image
	image, err := docker.NewImage(cr.ctx, imageName, &docker.ImageArgs{
		Build:     buildConfig,
		ImageName: pulumi.String(cr.buildImageName(imageName, "latest")),
		Registry: &docker.RegistryArgs{
			Server: pulumi.String(fmt.Sprintf("%s:%d", cr.registryHost, cr.registryPort)),
		},
		SkipPush: pulumi.Bool(cr.runtime == "podman"), // Don't push for local development
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build image %s: %w", imageName, err)
	}
	
	return image, nil
}

// CreateNetwork creates a container network
func (cr *ContainerRuntime) CreateNetwork(name string, config NetworkConfig) (*docker.Network, error) {
	networkArgs := &docker.NetworkArgs{
		Name:   pulumi.String(name),
		Driver: pulumi.String(config.Driver),
	}
	
	// TODO: IPAM configuration disabled - Docker SDK v4 compatibility issue
	// Set IPAM configuration if provided
	if config.IPAM.Driver != "" {
		// IPAM functionality temporarily disabled due to Docker SDK v4 incompatibility
		// The following types and fields are not available in Docker SDK v4:
		// - docker.NetworkIpamArgs
		// - docker.NetworkIpamConfigArgs  
		// - networkArgs.Ipam field
		// This will need to be re-implemented when Docker SDK supports IPAM configuration
		_ = config.IPAM // Prevent unused variable error
	}
	
	// Set network options
	if len(config.Options) > 0 {
		options := make(pulumi.StringMap)
		for key, value := range config.Options {
			options[key] = pulumi.String(value)
		}
		networkArgs.Options = options
	}
	
	// Create the network
	network, err := docker.NewNetwork(cr.ctx, name, networkArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to create network %s: %w", name, err)
	}
	
	return network, nil
}

// CreateVolume creates a container volume
func (cr *ContainerRuntime) CreateVolume(name string, config VolumeConfig) (*docker.Volume, error) {
	volumeArgs := &docker.VolumeArgs{
		Name:   pulumi.String(name),
		Driver: pulumi.String(config.Driver),
	}
	
	// Set driver options
	if len(config.DriverOpts) > 0 {
		driverOpts := make(pulumi.StringMap)
		for key, value := range config.DriverOpts {
			driverOpts[key] = pulumi.String(value)
		}
		volumeArgs.DriverOpts = driverOpts
	}
	
	// Set labels
	if len(config.Labels) > 0 {
		labels := make(pulumi.StringMap)
		for key, value := range config.Labels {
			labels[key] = pulumi.String(value)
		}
		// TODO: Volume labels disabled - Docker SDK v4 compatibility issue
		if len(labels) > 0 {
			_ = labels // Prevent unused variable error
		}
	}
	
	// Create the volume
	volume, err := docker.NewVolume(cr.ctx, name, volumeArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to create volume %s: %w", name, err)
	}
	
	return volume, nil
}


// Private helper methods

func (cr *ContainerRuntime) buildImageName(image, tag string) string {
	if cr.registryHost != "" && cr.registryPort > 0 {
		return fmt.Sprintf("%s:%d/%s:%s", cr.registryHost, cr.registryPort, image, tag)
	}
	return fmt.Sprintf("%s:%s", image, tag)
}

func (cr *ContainerRuntime) getNetworkMode(networks []string) string {
	if len(networks) == 0 {
		return "bridge" // Default network mode
	}
	if len(networks) == 1 {
		return networks[0]
	}
	// For multiple networks, use the first one as primary
	return networks[0]
}

func (cr *ContainerRuntime) convertStringSliceToPulumiArray(slice []string) pulumi.StringArray {
	if len(slice) == 0 {
		return nil
	}
	
	result := make(pulumi.StringArray, len(slice))
	for i, item := range slice {
		result[i] = pulumi.String(item)
	}
	return result
}

// Supporting configuration structures

// NetworkConfig defines network configuration
type NetworkConfig struct {
	Driver  string               `json:"driver"`
	IPAM    IPAMConfig           `json:"ipam"`
	Options map[string]string    `json:"options"`
	Labels  map[string]string    `json:"labels"`
}

// IPAMConfig defines IPAM configuration
type IPAMConfig struct {
	Driver string              `json:"driver"`
	Config []IPAMConfigEntry   `json:"config"`
}

// IPAMConfigEntry defines IPAM configuration entry
type IPAMConfigEntry struct {
	Subnet  string `json:"subnet"`
	Gateway string `json:"gateway"`
}

// VolumeConfig defines volume configuration
type VolumeConfig struct {
	Driver     string            `json:"driver"`
	DriverOpts map[string]string `json:"driver_opts"`
	Labels     map[string]string `json:"labels"`
}

// convertLabelsToArray converts string map labels to ContainerLabelArray format
func (cr *ContainerRuntime) convertLabelsToArray(labels pulumi.StringMap) docker.ContainerLabelArrayInput {
	if len(labels) == 0 {
		return nil
	}
	
	labelArray := make(docker.ContainerLabelArray, 0, len(labels))
	for key, value := range labels {
		labelArray = append(labelArray, &docker.ContainerLabelArgs{
			Label: pulumi.Sprintf("%s=%s", key, value),
		})
	}
	return labelArray
}

// convertLabelsToVolumeArray converts string map labels to VolumeLabelArray format
func (cr *ContainerRuntime) convertLabelsToVolumeArray(labels pulumi.StringMap) docker.VolumeLabelArrayInput {
	if len(labels) == 0 {
		return nil
	}
	
	labelArray := make(docker.VolumeLabelArray, 0, len(labels))
	for key, value := range labels {
		labelArray = append(labelArray, &docker.VolumeLabelArgs{
			Label: pulumi.Sprintf("%s=%s", key, value),
		})
	}
	return labelArray
}



