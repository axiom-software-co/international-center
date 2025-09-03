package application

import (
	"context"
	"fmt"

	"github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type APIDeployment struct {
	ctx         *pulumi.Context
	config      *config.Config
	networkName string
	environment string
}

type APIDeploymentResult struct {
	IdentityAPIContainer     *docker.Container
	IdentityDaprContainer    *docker.Container
	ContentAPIContainer      *docker.Container
	ContentDaprContainer     *docker.Container
	ServicesAPIContainer     *docker.Container
	ServicesDaprContainer    *docker.Container
	APINetwork              *docker.Network
	DaprComponentsVolume    *docker.Volume
}

type APIConfiguration struct {
	ImageRegistry    string
	ImageTag         string
	DatabaseURL      string
	RedisURL         string
	VaultURL         string
	StorageURL       string
	DaprHTTPPort     int
	DaprGRPCPort     int
	EnableDebug      bool
	EnableProfiling  bool
	LogLevel         string
}

func NewAPIDeployment(ctx *pulumi.Context, config *config.Config, networkName, environment string) *APIDeployment {
	return &APIDeployment{
		ctx:         ctx,
		config:      config,
		networkName: networkName,
		environment: environment,
	}
}

func (ad *APIDeployment) Deploy(ctx context.Context) (*APIDeploymentResult, error) {
	result := &APIDeploymentResult{}

	var err error

	result.APINetwork, err = ad.createAPINetwork()
	if err != nil {
		return nil, fmt.Errorf("failed to create API network: %w", err)
	}

	result.DaprComponentsVolume, err = ad.createDaprComponentsVolume()
	if err != nil {
		return nil, fmt.Errorf("failed to create Dapr components volume: %w", err)
	}

	apiConfig := ad.getAPIConfiguration()

	result.IdentityAPIContainer, result.IdentityDaprContainer, err = ad.deployIdentityAPI(result, apiConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy Identity API: %w", err)
	}

	result.ContentAPIContainer, result.ContentDaprContainer, err = ad.deployContentAPI(result, apiConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy Content API: %w", err)
	}

	result.ServicesAPIContainer, result.ServicesDaprContainer, err = ad.deployServicesAPI(result, apiConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy Services API: %w", err)
	}

	return result, nil
}

func (ad *APIDeployment) createAPINetwork() (*docker.Network, error) {
	network, err := docker.NewNetwork(ad.ctx, "api-network", &docker.NetworkArgs{
		Name:   pulumi.Sprintf("%s-api-network", ad.environment),
		Driver: pulumi.String("bridge"),
		Options: pulumi.StringMap{
			"com.docker.network.driver.mtu": pulumi.String("1500"),
		},
		Labels: pulumi.StringMap{
			"environment": pulumi.String(ad.environment),
			"component":   pulumi.String("api"),
			"managed-by":  pulumi.String("pulumi"),
		},
	})
	if err != nil {
		return nil, err
	}

	return network, nil
}

func (ad *APIDeployment) createDaprComponentsVolume() (*docker.Volume, error) {
	volume, err := docker.NewVolume(ad.ctx, "api-dapr-components", &docker.VolumeArgs{
		Name:   pulumi.Sprintf("%s-api-dapr-components", ad.environment),
		Driver: pulumi.String("local"),
		Labels: pulumi.StringMap{
			"environment": pulumi.String(ad.environment),
			"component":   pulumi.String("dapr"),
			"data-type":   pulumi.String("configuration"),
		},
	})
	if err != nil {
		return nil, err
	}

	return volume, nil
}

func (ad *APIDeployment) deployIdentityAPI(result *APIDeploymentResult, config *APIConfiguration) (*docker.Container, *docker.Container, error) {
	identityPort := ad.config.RequireInt("identity_api_port")

	envVars := pulumi.StringArray{
		pulumi.String("ENVIRONMENT=" + ad.environment),
		pulumi.String("LOG_LEVEL=" + config.LogLevel),
		pulumi.String("DATABASE_URL=" + config.DatabaseURL),
		pulumi.String("REDIS_URL=" + config.RedisURL),
		pulumi.String("VAULT_URL=" + config.VaultURL),
		pulumi.String("DAPR_HTTP_ENDPOINT=http://localhost:" + fmt.Sprintf("%d", config.DaprHTTPPort)),
		pulumi.String("DAPR_GRPC_ENDPOINT=localhost:" + fmt.Sprintf("%d", config.DaprGRPCPort)),
		pulumi.String("SERVICE_NAME=identity-api"),
		pulumi.String("SERVICE_PORT=" + fmt.Sprintf("%d", identityPort)),
	}

	if config.EnableDebug {
		envVars = append(envVars, pulumi.String("DEBUG=true"))
	}

	if config.EnableProfiling {
		envVars = append(envVars, pulumi.String("PROFILING=true"))
	}

	identityContainer, err := docker.NewContainer(ad.ctx, "identity-api", &docker.ContainerArgs{
		Name:    pulumi.Sprintf("%s-identity-api", ad.environment),
		Image:   pulumi.Sprintf("%s/identity-api:%s", config.ImageRegistry, config.ImageTag),
		Restart: pulumi.String("unless-stopped"),

		Envs: envVars,

		Ports: docker.ContainerPortArray{
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(identityPort),
				External: pulumi.Int(identityPort),
				Protocol: pulumi.String("tcp"),
			},
		},

		NetworksAdvanced: docker.ContainerNetworksAdvancedArray{
			&docker.ContainerNetworksAdvancedArgs{
				Name: result.APINetwork.Name,
				Aliases: pulumi.StringArray{
					pulumi.String("identity-api"),
				},
			},
		},

		Healthcheck: &docker.ContainerHealthcheckArgs{
			Tests: pulumi.StringArray{
				pulumi.String("CMD-SHELL"),
				pulumi.Sprintf("curl -f http://localhost:%d/health || exit 1", identityPort),
			},
			Interval: pulumi.String("30s"),
			Timeout:  pulumi.String("10s"),
			Retries:  pulumi.Int(3),
			StartPeriod: pulumi.String("60s"),
		},

		Labels: pulumi.StringMap{
			"environment": pulumi.String(ad.environment),
			"component":   pulumi.String("identity-api"),
			"service":     pulumi.String("api"),
			"managed-by":  pulumi.String("pulumi"),
			"dapr-app-id": pulumi.String("identity-api"),
		},

		LogDriver: pulumi.String("json-file"),
		LogOpts: pulumi.StringMap{
			"max-size": pulumi.String("10m"),
			"max-file": pulumi.String("3"),
		},
	})
	if err != nil {
		return nil, nil, err
	}

	identityDaprContainer, err := ad.deployDaprSidecar("identity", identityPort, result, config)
	if err != nil {
		return nil, nil, err
	}

	return identityContainer, identityDaprContainer, nil
}

func (ad *APIDeployment) deployContentAPI(result *APIDeploymentResult, config *APIConfiguration) (*docker.Container, *docker.Container, error) {
	contentPort := ad.config.RequireInt("content_api_port")

	envVars := pulumi.StringArray{
		pulumi.String("ENVIRONMENT=" + ad.environment),
		pulumi.String("LOG_LEVEL=" + config.LogLevel),
		pulumi.String("DATABASE_URL=" + config.DatabaseURL),
		pulumi.String("REDIS_URL=" + config.RedisURL),
		pulumi.String("VAULT_URL=" + config.VaultURL),
		pulumi.String("STORAGE_URL=" + config.StorageURL),
		pulumi.String("DAPR_HTTP_ENDPOINT=http://localhost:" + fmt.Sprintf("%d", config.DaprHTTPPort)),
		pulumi.String("DAPR_GRPC_ENDPOINT=localhost:" + fmt.Sprintf("%d", config.DaprGRPCPort)),
		pulumi.String("SERVICE_NAME=content-api"),
		pulumi.String("SERVICE_PORT=" + fmt.Sprintf("%d", contentPort)),
	}

	if config.EnableDebug {
		envVars = append(envVars, pulumi.String("DEBUG=true"))
	}

	if config.EnableProfiling {
		envVars = append(envVars, pulumi.String("PROFILING=true"))
	}

	contentContainer, err := docker.NewContainer(ad.ctx, "content-api", &docker.ContainerArgs{
		Name:    pulumi.Sprintf("%s-content-api", ad.environment),
		Image:   pulumi.Sprintf("%s/content-api:%s", config.ImageRegistry, config.ImageTag),
		Restart: pulumi.String("unless-stopped"),

		Envs: envVars,

		Ports: docker.ContainerPortArray{
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(contentPort),
				External: pulumi.Int(contentPort),
				Protocol: pulumi.String("tcp"),
			},
		},

		NetworksAdvanced: docker.ContainerNetworksAdvancedArray{
			&docker.ContainerNetworksAdvancedArgs{
				Name: result.APINetwork.Name,
				Aliases: pulumi.StringArray{
					pulumi.String("content-api"),
				},
			},
		},

		Healthcheck: &docker.ContainerHealthcheckArgs{
			Tests: pulumi.StringArray{
				pulumi.String("CMD-SHELL"),
				pulumi.Sprintf("curl -f http://localhost:%d/health || exit 1", contentPort),
			},
			Interval: pulumi.String("30s"),
			Timeout:  pulumi.String("10s"),
			Retries:  pulumi.Int(3),
			StartPeriod: pulumi.String("60s"),
		},

		Labels: pulumi.StringMap{
			"environment": pulumi.String(ad.environment),
			"component":   pulumi.String("content-api"),
			"service":     pulumi.String("api"),
			"managed-by":  pulumi.String("pulumi"),
			"dapr-app-id": pulumi.String("content-api"),
		},

		LogDriver: pulumi.String("json-file"),
		LogOpts: pulumi.StringMap{
			"max-size": pulumi.String("10m"),
			"max-file": pulumi.String("3"),
		},
	})
	if err != nil {
		return nil, nil, err
	}

	contentDaprContainer, err := ad.deployDaprSidecar("content", contentPort, result, config)
	if err != nil {
		return nil, nil, err
	}

	return contentContainer, contentDaprContainer, nil
}

func (ad *APIDeployment) deployServicesAPI(result *APIDeploymentResult, config *APIConfiguration) (*docker.Container, *docker.Container, error) {
	servicesPort := ad.config.RequireInt("services_api_port")

	envVars := pulumi.StringArray{
		pulumi.String("ENVIRONMENT=" + ad.environment),
		pulumi.String("LOG_LEVEL=" + config.LogLevel),
		pulumi.String("DATABASE_URL=" + config.DatabaseURL),
		pulumi.String("REDIS_URL=" + config.RedisURL),
		pulumi.String("VAULT_URL=" + config.VaultURL),
		pulumi.String("STORAGE_URL=" + config.StorageURL),
		pulumi.String("DAPR_HTTP_ENDPOINT=http://localhost:" + fmt.Sprintf("%d", config.DaprHTTPPort)),
		pulumi.String("DAPR_GRPC_ENDPOINT=localhost:" + fmt.Sprintf("%d", config.DaprGRPCPort)),
		pulumi.String("SERVICE_NAME=services-api"),
		pulumi.String("SERVICE_PORT=" + fmt.Sprintf("%d", servicesPort)),
		pulumi.String("CONTENT_API_URL=http://content-api:" + fmt.Sprintf("%d", ad.config.RequireInt("content_api_port"))),
	}

	if config.EnableDebug {
		envVars = append(envVars, pulumi.String("DEBUG=true"))
	}

	if config.EnableProfiling {
		envVars = append(envVars, pulumi.String("PROFILING=true"))
	}

	servicesContainer, err := docker.NewContainer(ad.ctx, "services-api", &docker.ContainerArgs{
		Name:    pulumi.Sprintf("%s-services-api", ad.environment),
		Image:   pulumi.Sprintf("%s/services-api:%s", config.ImageRegistry, config.ImageTag),
		Restart: pulumi.String("unless-stopped"),

		Envs: envVars,

		Ports: docker.ContainerPortArray{
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(servicesPort),
				External: pulumi.Int(servicesPort),
				Protocol: pulumi.String("tcp"),
			},
		},

		NetworksAdvanced: docker.ContainerNetworksAdvancedArray{
			&docker.ContainerNetworksAdvancedArgs{
				Name: result.APINetwork.Name,
				Aliases: pulumi.StringArray{
					pulumi.String("services-api"),
				},
			},
		},

		Healthcheck: &docker.ContainerHealthcheckArgs{
			Tests: pulumi.StringArray{
				pulumi.String("CMD-SHELL"),
				pulumi.Sprintf("curl -f http://localhost:%d/health || exit 1", servicesPort),
			},
			Interval: pulumi.String("30s"),
			Timeout:  pulumi.String("10s"),
			Retries:  pulumi.Int(3),
			StartPeriod: pulumi.String("60s"),
		},

		Labels: pulumi.StringMap{
			"environment": pulumi.String(ad.environment),
			"component":   pulumi.String("services-api"),
			"service":     pulumi.String("api"),
			"managed-by":  pulumi.String("pulumi"),
			"dapr-app-id": pulumi.String("services-api"),
		},

		LogDriver: pulumi.String("json-file"),
		LogOpts: pulumi.StringMap{
			"max-size": pulumi.String("10m"),
			"max-file": pulumi.String("3"),
		},
	})
	if err != nil {
		return nil, nil, err
	}

	servicesDaprContainer, err := ad.deployDaprSidecar("services", servicesPort, result, config)
	if err != nil {
		return nil, nil, err
	}

	return servicesContainer, servicesDaprContainer, nil
}

func (ad *APIDeployment) deployDaprSidecar(appName string, appPort int, result *APIDeploymentResult, config *APIConfiguration) (*docker.Container, error) {
	daprVersion := ad.config.Get("dapr_version")
	if daprVersion == "" {
		daprVersion = "1.12.0"
	}

	daprContainer, err := docker.NewContainer(ad.ctx, fmt.Sprintf("%s-dapr", appName), &docker.ContainerArgs{
		Name:    pulumi.Sprintf("%s-%s-dapr", ad.environment, appName),
		Image:   pulumi.Sprintf("daprio/daprd:%s", daprVersion),
		Restart: pulumi.String("unless-stopped"),

		Command: pulumi.StringArray{
			pulumi.String("./daprd"),
			pulumi.Sprintf("--app-id=%s-api", appName),
			pulumi.Sprintf("--app-port=%d", appPort),
			pulumi.String("--dapr-http-port=" + fmt.Sprintf("%d", config.DaprHTTPPort)),
			pulumi.String("--dapr-grpc-port=" + fmt.Sprintf("%d", config.DaprGRPCPort)),
			pulumi.String("--placement-host-address=dapr-placement:50005"),
			pulumi.String("--components-path=/components"),
			pulumi.String("--log-level=info"),
			pulumi.String("--enable-profiling"),
			pulumi.String("--enable-metrics"),
			pulumi.String("--metrics-port=9090"),
		},

		Ports: docker.ContainerPortArray{
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(config.DaprHTTPPort),
				External: pulumi.Int(config.DaprHTTPPort + appPort - 8001),
				Protocol: pulumi.String("tcp"),
			},
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(config.DaprGRPCPort),
				External: pulumi.Int(config.DaprGRPCPort + appPort - 8001),
				Protocol: pulumi.String("tcp"),
			},
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(9090),
				External: pulumi.Int(9090 + appPort - 8001),
				Protocol: pulumi.String("tcp"),
			},
		},

		Mounts: docker.ContainerMountArray{
			&docker.ContainerMountArgs{
				Type:   pulumi.String("volume"),
				Source: result.DaprComponentsVolume.Name,
				Target: pulumi.String("/components"),
			},
		},

		NetworksAdvanced: docker.ContainerNetworksAdvancedArray{
			&docker.ContainerNetworksAdvancedArgs{
				Name: result.APINetwork.Name,
				Aliases: pulumi.StringArray{
					pulumi.Sprintf("%s-dapr", appName),
				},
			},
		},

		Labels: pulumi.StringMap{
			"environment": pulumi.String(ad.environment),
			"component":   pulumi.String("dapr"),
			"service":     pulumi.String("sidecar"),
			"app":         pulumi.String(fmt.Sprintf("%s-api", appName)),
			"managed-by":  pulumi.String("pulumi"),
		},

		LogDriver: pulumi.String("json-file"),
		LogOpts: pulumi.StringMap{
			"max-size": pulumi.String("10m"),
			"max-file": pulumi.String("3"),
		},
	})
	if err != nil {
		return nil, err
	}

	return daprContainer, nil
}

func (ad *APIDeployment) getAPIConfiguration() *APIConfiguration {
	return &APIConfiguration{
		ImageRegistry:   ad.config.Get("container_registry"),
		ImageTag:        ad.config.Get("image_tag"),
		DatabaseURL:     ad.config.Require("database_url"),
		RedisURL:        ad.config.Require("redis_url"),
		VaultURL:        ad.config.Require("vault_url"),
		StorageURL:      ad.config.Require("storage_url"),
		DaprHTTPPort:    ad.config.RequireInt("dapr_http_port"),
		DaprGRPCPort:    ad.config.RequireInt("dapr_grpc_port"),
		EnableDebug:     ad.config.GetBool("enable_debugging"),
		EnableProfiling: ad.config.GetBool("enable_profiling"),
		LogLevel:        ad.config.Get("log_level"),
	}
}

func (ad *APIDeployment) ValidateDeployment(ctx context.Context, result *APIDeploymentResult) error {
	if result.IdentityAPIContainer == nil {
		return fmt.Errorf("Identity API container is not deployed")
	}

	if result.IdentityDaprContainer == nil {
		return fmt.Errorf("Identity Dapr sidecar container is not deployed")
	}

	if result.ContentAPIContainer == nil {
		return fmt.Errorf("Content API container is not deployed")
	}

	if result.ContentDaprContainer == nil {
		return fmt.Errorf("Content Dapr sidecar container is not deployed")
	}

	if result.ServicesAPIContainer == nil {
		return fmt.Errorf("Services API container is not deployed")
	}

	if result.ServicesDaprContainer == nil {
		return fmt.Errorf("Services Dapr sidecar container is not deployed")
	}

	return nil
}

func (ad *APIDeployment) GetAPIEndpoints() map[string]string {
	identityPort := ad.config.RequireInt("identity_api_port")
	contentPort := ad.config.RequireInt("content_api_port")
	servicesPort := ad.config.RequireInt("services_api_port")

	return map[string]string{
		"identity": fmt.Sprintf("http://localhost:%d", identityPort),
		"content":  fmt.Sprintf("http://localhost:%d", contentPort),
		"services": fmt.Sprintf("http://localhost:%d", servicesPort),
	}
}