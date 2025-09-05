package infrastructure

import (
	"fmt"
	"time"

	"github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ServiceContainerConfig defines the configuration for a standardized service container
type ServiceContainerConfig struct {
	// Basic Configuration
	ServiceName     string
	Environment     string
	Image           string
	Tag             string
	
	// Network Configuration
	InternalPort    int
	ExternalPort    int
	Networks        []string
	
	// Dapr Configuration
	DaprHTTPPort    int
	DaprGRPCPort    int
	
	// Environment Variables
	BaseEnvironment map[string]pulumi.StringInput
	ServiceSpecific map[string]pulumi.StringInput
	
	// Health Check Configuration
	HealthPath      string
	HealthInterval  time.Duration
	HealthTimeout   time.Duration
	HealthRetries   int
	
	// Dependencies
	Dependencies    []pulumi.Resource
	
	// Metadata
	Component       string
	Labels          map[string]string
}

// DaprSidecarConfig defines the configuration for a standardized Dapr sidecar
type DaprSidecarConfig struct {
	// Basic Configuration
	AppID           string
	Environment     string
	AppPort         int
	
	// Port Configuration  
	HTTPPort        int
	GRPCPort        int
	
	// Infrastructure
	Networks        []string
	ComponentsVolume *docker.Volume
	PlacementContainer *docker.Container
	RedisContainer  *docker.Container
	
	// Dapr Configuration
	DaprVersion     string
	ComponentsPath  string
	LogLevel        string
	
	// Metadata
	Labels          map[string]string
}

// NetworkConfig defines the configuration for standardized networks
type NetworkConfig struct {
	Name        string
	Environment string
	Component   string
	Driver      string
	Subnet      string
	Gateway     string
	Labels      map[string]string
}

// ContainerTemplateFactory provides standardized container deployment patterns
type ContainerTemplateFactory struct {
	ctx         *pulumi.Context
	environment string
}

// NewContainerTemplateFactory creates a new container template factory
func NewContainerTemplateFactory(ctx *pulumi.Context, environment string) *ContainerTemplateFactory {
	return &ContainerTemplateFactory{
		ctx:         ctx,
		environment: environment,
	}
}

// CreateServiceContainer creates a service container using the standardized template
func (ctf *ContainerTemplateFactory) CreateServiceContainer(config ServiceContainerConfig) (*docker.Container, error) {
	// Build complete environment variables
	environment := ctf.buildEnvironmentVariables(config)
	
	// Build health check configuration
	healthCheck := ctf.buildHealthCheck(config)
	
	// Build labels
	labels := ctf.buildLabels(config)
	
	// Create the container
	container, err := docker.NewContainer(ctf.ctx, fmt.Sprintf("%s-%s", config.Environment, config.ServiceName), &docker.ContainerArgs{
		Name:    pulumi.Sprintf("%s-%s", config.Environment, config.ServiceName),
		Image:   pulumi.Sprintf("%s:%s", config.Image, config.Tag),
		Restart: pulumi.String("unless-stopped"),
		
		Envs: environment,
		
		Ports: docker.ContainerPortArray{
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(config.InternalPort),
				External: pulumi.Int(config.ExternalPort),
				Protocol: pulumi.String("tcp"),
			},
		},
		
		NetworksAdvanced: ctf.buildNetworkConfiguration(config.Networks),
		
		Healthcheck: &docker.ContainerHealthcheckArgs{
			Tests:       pulumi.StringArray{pulumi.String(healthCheck)},
			Interval:    pulumi.String(config.HealthInterval.String()),
			Timeout:     pulumi.String(config.HealthTimeout.String()),
			Retries:     pulumi.Int(config.HealthRetries),
			StartPeriod: pulumi.String("30s"),
		},
		
		Labels: labels,
	}, pulumi.DependsOn(config.Dependencies))
	
	if err != nil {
		return nil, fmt.Errorf("failed to create service container %s: %w", config.ServiceName, err)
	}
	
	return container, nil
}

// CreateDaprSidecar creates a Dapr sidecar container using the standardized template
func (ctf *ContainerTemplateFactory) CreateDaprSidecar(config DaprSidecarConfig) (*docker.Container, error) {
	// Build Dapr command
	command := ctf.buildDaprCommand(config)
	
	// Build port configuration
	ports := ctf.buildDaprPorts(config)
	
	// Build volume mounts
	mounts := ctf.buildDaprMounts(config)
	
	// Build labels
	labels := ctf.buildDaprLabels(config)
	
	// Create the Dapr sidecar container
	container, err := docker.NewContainer(ctf.ctx, fmt.Sprintf("%s-%s-dapr-sidecar", config.Environment, config.AppID), &docker.ContainerArgs{
		Name:    pulumi.Sprintf("%s-%s-dapr-sidecar", config.Environment, config.AppID),
		Image:   pulumi.Sprintf("daprio/daprd:%s", config.DaprVersion),
		Restart: pulumi.String("unless-stopped"),
		
		Command: command,
		Ports:   ports,
		Mounts:  mounts,
		
		NetworksAdvanced: ctf.buildNetworkConfiguration(config.Networks),
		
		Labels: labels,
	}, pulumi.DependsOn([]pulumi.Resource{
		config.PlacementContainer,
		config.RedisContainer,
	}))
	
	if err != nil {
		return nil, fmt.Errorf("failed to create Dapr sidecar for %s: %w", config.AppID, err)
	}
	
	return container, nil
}

// CreateNetwork creates a network using the standardized template
func (ctf *ContainerTemplateFactory) CreateNetwork(config NetworkConfig) (*docker.Network, error) {
	network, err := docker.NewNetwork(ctf.ctx, fmt.Sprintf("%s-%s-network", config.Environment, config.Component), &docker.NetworkArgs{
		Name:   pulumi.Sprintf("%s-%s-network", config.Environment, config.Component),
		Driver: pulumi.String(config.Driver),
		IpamConfigs: docker.NetworkIpamConfigArray{
			&docker.NetworkIpamConfigArgs{
				Subnet:  pulumi.String(config.Subnet),
				Gateway: pulumi.String(config.Gateway),
			},
		},
		Options: pulumi.StringMap{
			"com.docker.network.bridge.name": pulumi.Sprintf("br-%s", config.Component),
			"com.docker.network.driver.mtu":  pulumi.String("1500"),
		},
		Labels: ctf.buildNetworkLabels(config),
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to create network %s: %w", config.Name, err)
	}
	
	return network, nil
}

// Helper methods for building configuration

func (ctf *ContainerTemplateFactory) buildEnvironmentVariables(config ServiceContainerConfig) pulumi.StringArray {
	envVars := make([]pulumi.StringInput, 0)
	
	// Add base environment variables
	for key, value := range config.BaseEnvironment {
		envVars = append(envVars, pulumi.Sprintf("%s=%s", key, value))
	}
	
	// Add service-specific environment variables
	for key, value := range config.ServiceSpecific {
		envVars = append(envVars, pulumi.Sprintf("%s=%s", key, value))
	}
	
	// Add standard environment variables
	envVars = append(envVars, pulumi.Sprintf("ENVIRONMENT=%s", config.Environment))
	envVars = append(envVars, pulumi.Sprintf("SERVICE_NAME=%s", config.ServiceName))
	envVars = append(envVars, pulumi.Sprintf("DAPR_HTTP_PORT=%d", config.DaprHTTPPort))
	envVars = append(envVars, pulumi.Sprintf("DAPR_GRPC_PORT=%d", config.DaprGRPCPort))
	
	return pulumi.StringArray(envVars)
}

func (ctf *ContainerTemplateFactory) buildHealthCheck(config ServiceContainerConfig) string {
	if config.HealthPath == "" {
		config.HealthPath = "/health"
	}
	return fmt.Sprintf("CMD-SHELL wget --no-verbose --tries=1 --spider http://localhost:%d%s || exit 1", 
		config.InternalPort, config.HealthPath)
}

func (ctf *ContainerTemplateFactory) buildLabels(config ServiceContainerConfig) docker.ContainerLabelArray {
	labels := docker.ContainerLabelArray{}
	
	// Standard labels
	labels = append(labels, &docker.ContainerLabelArgs{
		Label: pulumi.String("environment"),
		Value: pulumi.String(config.Environment),
	})
	labels = append(labels, &docker.ContainerLabelArgs{
		Label: pulumi.String("service"),
		Value: pulumi.String(config.ServiceName),
	})
	labels = append(labels, &docker.ContainerLabelArgs{
		Label: pulumi.String("component"),
		Value: pulumi.String(config.Component),
	})
	labels = append(labels, &docker.ContainerLabelArgs{
		Label: pulumi.String("managed-by"),
		Value: pulumi.String("pulumi"),
	})
	
	// Custom labels
	for key, value := range config.Labels {
		labels = append(labels, &docker.ContainerLabelArgs{
			Label: pulumi.String(key),
			Value: pulumi.String(value),
		})
	}
	
	return labels
}

func (ctf *ContainerTemplateFactory) buildNetworkConfiguration(networks []string) docker.ContainerNetworksAdvancedArray {
	networkConfig := docker.ContainerNetworksAdvancedArray{}
	
	for _, network := range networks {
		networkConfig = append(networkConfig, &docker.ContainerNetworksAdvancedArgs{
			Name: pulumi.String(network),
		})
	}
	
	return networkConfig
}

func (ctf *ContainerTemplateFactory) buildDaprCommand(config DaprSidecarConfig) pulumi.StringArray {
	return pulumi.StringArray{
		pulumi.String("./daprd"),
		pulumi.Sprintf("--app-id=%s", config.AppID),
		pulumi.Sprintf("--app-port=%d", config.AppPort),
		pulumi.Sprintf("--dapr-http-port=%d", config.HTTPPort),
		pulumi.Sprintf("--dapr-grpc-port=%d", config.GRPCPort),
		pulumi.String("--placement-host-address=dapr-placement:50005"),
		pulumi.Sprintf("--components-path=%s", config.ComponentsPath),
		pulumi.Sprintf("--log-level=%s", config.LogLevel),
		pulumi.String("--app-ssl=false"),
	}
}

func (ctf *ContainerTemplateFactory) buildDaprPorts(config DaprSidecarConfig) docker.ContainerPortArray {
	return docker.ContainerPortArray{
		&docker.ContainerPortArgs{
			Internal: pulumi.Int(config.HTTPPort),
			External: pulumi.Int(config.HTTPPort),
			Protocol: pulumi.String("tcp"),
		},
		&docker.ContainerPortArgs{
			Internal: pulumi.Int(config.GRPCPort),
			External: pulumi.Int(config.GRPCPort),
			Protocol: pulumi.String("tcp"),
		},
	}
}

func (ctf *ContainerTemplateFactory) buildDaprMounts(config DaprSidecarConfig) docker.ContainerMountArray {
	return docker.ContainerMountArray{
		&docker.ContainerMountArgs{
			Type:   pulumi.String("volume"),
			Source: config.ComponentsVolume.Name,
			Target: pulumi.String(config.ComponentsPath),
		},
	}
}

func (ctf *ContainerTemplateFactory) buildDaprLabels(config DaprSidecarConfig) docker.ContainerLabelArray {
	labels := docker.ContainerLabelArray{
		&docker.ContainerLabelArgs{
			Label: pulumi.String("environment"),
			Value: pulumi.String(config.Environment),
		},
		&docker.ContainerLabelArgs{
			Label: pulumi.String("component"),
			Value: pulumi.String("dapr-sidecar"),
		},
		&docker.ContainerLabelArgs{
			Label: pulumi.String("app-id"),
			Value: pulumi.String(config.AppID),
		},
		&docker.ContainerLabelArgs{
			Label: pulumi.String("managed-by"),
			Value: pulumi.String("pulumi"),
		},
	}
	
	// Custom labels
	for key, value := range config.Labels {
		labels = append(labels, &docker.ContainerLabelArgs{
			Label: pulumi.String(key),
			Value: pulumi.String(value),
		})
	}
	
	return labels
}

func (ctf *ContainerTemplateFactory) buildNetworkLabels(config NetworkConfig) docker.NetworkLabelArray {
	return docker.NetworkLabelArray{
		&docker.NetworkLabelArgs{
			Label: pulumi.String("environment"),
			Value: pulumi.String(config.Environment),
		},
		&docker.NetworkLabelArgs{
			Label: pulumi.String("component"),
			Value: pulumi.String(config.Component),
		},
		&docker.NetworkLabelArgs{
			Label: pulumi.String("managed-by"),
			Value: pulumi.String("pulumi"),
		},
	}
}