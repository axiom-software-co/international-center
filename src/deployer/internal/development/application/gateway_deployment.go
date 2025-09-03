package application

import (
	"context"
	"fmt"

	"github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type GatewayDeployment struct {
	ctx         *pulumi.Context
	config      *config.Config
	networkName string
	environment string
}

type GatewayDeploymentResult struct {
	PublicGatewayContainer    *docker.Container
	PublicDaprContainer       *docker.Container
	AdminGatewayContainer     *docker.Container
	AdminDaprContainer        *docker.Container
	GatewayNetwork           *docker.Network
	GatewayConfigVolume      *docker.Volume
	DaprComponentsVolume     *docker.Volume
}

type GatewayConfiguration struct {
	ImageRegistry      string
	ImageTag           string
	RedisURL           string
	VaultURL           string
	DaprHTTPPort       int
	DaprGRPCPort       int
	EnableDebug        bool
	EnableProfiling    bool
	LogLevel           string
	CorsOrigins        []string
	RateLimitEnabled   bool
	AuthEnabled        bool
	IdentityAPIURL     string
	ContentAPIURL      string
	ServicesAPIURL     string
}

func NewGatewayDeployment(ctx *pulumi.Context, config *config.Config, networkName, environment string) *GatewayDeployment {
	return &GatewayDeployment{
		ctx:         ctx,
		config:      config,
		networkName: networkName,
		environment: environment,
	}
}

func (gd *GatewayDeployment) Deploy(ctx context.Context) (*GatewayDeploymentResult, error) {
	result := &GatewayDeploymentResult{}

	var err error

	result.GatewayNetwork, err = gd.createGatewayNetwork()
	if err != nil {
		return nil, fmt.Errorf("failed to create gateway network: %w", err)
	}

	result.GatewayConfigVolume, err = gd.createGatewayConfigVolume()
	if err != nil {
		return nil, fmt.Errorf("failed to create gateway config volume: %w", err)
	}

	result.DaprComponentsVolume, err = gd.createDaprComponentsVolume()
	if err != nil {
		return nil, fmt.Errorf("failed to create Dapr components volume: %w", err)
	}

	gatewayConfig := gd.getGatewayConfiguration()

	result.PublicGatewayContainer, result.PublicDaprContainer, err = gd.deployPublicGateway(result, gatewayConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy Public Gateway: %w", err)
	}

	result.AdminGatewayContainer, result.AdminDaprContainer, err = gd.deployAdminGateway(result, gatewayConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy Admin Gateway: %w", err)
	}

	return result, nil
}

func (gd *GatewayDeployment) createGatewayNetwork() (*docker.Network, error) {
	network, err := docker.NewNetwork(gd.ctx, "gateway-network", &docker.NetworkArgs{
		Name:   pulumi.Sprintf("%s-gateway-network", gd.environment),
		Driver: pulumi.String("bridge"),
		Options: pulumi.StringMap{
			"com.docker.network.driver.mtu": pulumi.String("1500"),
		},
		Labels: pulumi.StringMap{
			"environment": pulumi.String(gd.environment),
			"component":   pulumi.String("gateway"),
			"managed-by":  pulumi.String("pulumi"),
		},
	})
	if err != nil {
		return nil, err
	}

	return network, nil
}

func (gd *GatewayDeployment) createGatewayConfigVolume() (*docker.Volume, error) {
	volume, err := docker.NewVolume(gd.ctx, "gateway-config", &docker.VolumeArgs{
		Name:   pulumi.Sprintf("%s-gateway-config", gd.environment),
		Driver: pulumi.String("local"),
		Labels: pulumi.StringMap{
			"environment": pulumi.String(gd.environment),
			"component":   pulumi.String("gateway"),
			"data-type":   pulumi.String("configuration"),
		},
	})
	if err != nil {
		return nil, err
	}

	return volume, nil
}

func (gd *GatewayDeployment) createDaprComponentsVolume() (*docker.Volume, error) {
	volume, err := docker.NewVolume(gd.ctx, "gateway-dapr-components", &docker.VolumeArgs{
		Name:   pulumi.Sprintf("%s-gateway-dapr-components", gd.environment),
		Driver: pulumi.String("local"),
		Labels: pulumi.StringMap{
			"environment": pulumi.String(gd.environment),
			"component":   pulumi.String("dapr"),
			"data-type":   pulumi.String("configuration"),
		},
	})
	if err != nil {
		return nil, err
	}

	return volume, nil
}

func (gd *GatewayDeployment) deployPublicGateway(result *GatewayDeploymentResult, config *GatewayConfiguration) (*docker.Container, *docker.Container, error) {
	publicPort := gd.config.RequireInt("public_gateway_port")

	envVars := pulumi.StringArray{
		pulumi.String("ENVIRONMENT=" + gd.environment),
		pulumi.String("LOG_LEVEL=" + config.LogLevel),
		pulumi.String("REDIS_URL=" + config.RedisURL),
		pulumi.String("VAULT_URL=" + config.VaultURL),
		pulumi.String("DAPR_HTTP_ENDPOINT=http://localhost:" + fmt.Sprintf("%d", config.DaprHTTPPort)),
		pulumi.String("DAPR_GRPC_ENDPOINT=localhost:" + fmt.Sprintf("%d", config.DaprGRPCPort)),
		pulumi.String("SERVICE_NAME=public-gateway"),
		pulumi.String("SERVICE_PORT=" + fmt.Sprintf("%d", publicPort)),
		pulumi.String("GATEWAY_TYPE=public"),
		pulumi.String("CONTENT_API_URL=" + config.ContentAPIURL),
		pulumi.String("SERVICES_API_URL=" + config.ServicesAPIURL),
		pulumi.String("RATE_LIMIT_ENABLED=" + fmt.Sprintf("%t", config.RateLimitEnabled)),
		pulumi.String("RATE_LIMIT_REQUESTS_PER_MINUTE=1000"),
		pulumi.String("CORS_ENABLED=true"),
		pulumi.String("CORS_ORIGINS=http://localhost:3000,http://localhost:8080"),
	}

	if config.EnableDebug {
		envVars = append(envVars, pulumi.String("DEBUG=true"))
	}

	if config.EnableProfiling {
		envVars = append(envVars, pulumi.String("PROFILING=true"))
	}

	publicGatewayContainer, err := docker.NewContainer(gd.ctx, "public-gateway", &docker.ContainerArgs{
		Name:    pulumi.Sprintf("%s-public-gateway", gd.environment),
		Image:   pulumi.Sprintf("%s/public-gateway:%s", config.ImageRegistry, config.ImageTag),
		Restart: pulumi.String("unless-stopped"),

		Envs: envVars,

		Ports: docker.ContainerPortArray{
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(publicPort),
				External: pulumi.Int(publicPort),
				Protocol: pulumi.String("tcp"),
			},
		},

		Mounts: docker.ContainerMountArray{
			&docker.ContainerMountArgs{
				Type:   pulumi.String("volume"),
				Source: result.GatewayConfigVolume.Name,
				Target: pulumi.String("/config"),
			},
		},

		NetworksAdvanced: docker.ContainerNetworksAdvancedArray{
			&docker.ContainerNetworksAdvancedArgs{
				Name: result.GatewayNetwork.Name,
				Aliases: pulumi.StringArray{
					pulumi.String("public-gateway"),
				},
			},
		},

		Healthcheck: &docker.ContainerHealthcheckArgs{
			Tests: pulumi.StringArray{
				pulumi.String("CMD-SHELL"),
				pulumi.Sprintf("curl -f http://localhost:%d/health || exit 1", publicPort),
			},
			Interval: pulumi.String("30s"),
			Timeout:  pulumi.String("10s"),
			Retries:  pulumi.Int(3),
			StartPeriod: pulumi.String("60s"),
		},

		Labels: pulumi.StringMap{
			"environment": pulumi.String(gd.environment),
			"component":   pulumi.String("public-gateway"),
			"service":     pulumi.String("gateway"),
			"managed-by":  pulumi.String("pulumi"),
			"dapr-app-id": pulumi.String("public-gateway"),
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

	publicDaprContainer, err := gd.deployDaprSidecar("public-gateway", publicPort, result, config)
	if err != nil {
		return nil, nil, err
	}

	return publicGatewayContainer, publicDaprContainer, nil
}

func (gd *GatewayDeployment) deployAdminGateway(result *GatewayDeploymentResult, config *GatewayConfiguration) (*docker.Container, *docker.Container, error) {
	adminPort := gd.config.RequireInt("admin_gateway_port")

	envVars := pulumi.StringArray{
		pulumi.String("ENVIRONMENT=" + gd.environment),
		pulumi.String("LOG_LEVEL=" + config.LogLevel),
		pulumi.String("REDIS_URL=" + config.RedisURL),
		pulumi.String("VAULT_URL=" + config.VaultURL),
		pulumi.String("DAPR_HTTP_ENDPOINT=http://localhost:" + fmt.Sprintf("%d", config.DaprHTTPPort)),
		pulumi.String("DAPR_GRPC_ENDPOINT=localhost:" + fmt.Sprintf("%d", config.DaprGRPCPort)),
		pulumi.String("SERVICE_NAME=admin-gateway"),
		pulumi.String("SERVICE_PORT=" + fmt.Sprintf("%d", adminPort)),
		pulumi.String("GATEWAY_TYPE=admin"),
		pulumi.String("IDENTITY_API_URL=" + config.IdentityAPIURL),
		pulumi.String("CONTENT_API_URL=" + config.ContentAPIURL),
		pulumi.String("SERVICES_API_URL=" + config.ServicesAPIURL),
		pulumi.String("AUTH_ENABLED=" + fmt.Sprintf("%t", config.AuthEnabled)),
		pulumi.String("RATE_LIMIT_ENABLED=" + fmt.Sprintf("%t", config.RateLimitEnabled)),
		pulumi.String("RATE_LIMIT_REQUESTS_PER_MINUTE=100"),
		pulumi.String("AUDIT_LOGGING_ENABLED=true"),
		pulumi.String("RBAC_ENABLED=true"),
		pulumi.String("CORS_ENABLED=true"),
		pulumi.String("CORS_ORIGINS=http://localhost:3000,http://localhost:8080"),
	}

	if config.EnableDebug {
		envVars = append(envVars, pulumi.String("DEBUG=true"))
	}

	if config.EnableProfiling {
		envVars = append(envVars, pulumi.String("PROFILING=true"))
	}

	adminGatewayContainer, err := docker.NewContainer(gd.ctx, "admin-gateway", &docker.ContainerArgs{
		Name:    pulumi.Sprintf("%s-admin-gateway", gd.environment),
		Image:   pulumi.Sprintf("%s/admin-gateway:%s", config.ImageRegistry, config.ImageTag),
		Restart: pulumi.String("unless-stopped"),

		Envs: envVars,

		Ports: docker.ContainerPortArray{
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(adminPort),
				External: pulumi.Int(adminPort),
				Protocol: pulumi.String("tcp"),
			},
		},

		Mounts: docker.ContainerMountArray{
			&docker.ContainerMountArgs{
				Type:   pulumi.String("volume"),
				Source: result.GatewayConfigVolume.Name,
				Target: pulumi.String("/config"),
			},
		},

		NetworksAdvanced: docker.ContainerNetworksAdvancedArray{
			&docker.ContainerNetworksAdvancedArgs{
				Name: result.GatewayNetwork.Name,
				Aliases: pulumi.StringArray{
					pulumi.String("admin-gateway"),
				},
			},
		},

		Healthcheck: &docker.ContainerHealthcheckArgs{
			Tests: pulumi.StringArray{
				pulumi.String("CMD-SHELL"),
				pulumi.Sprintf("curl -f http://localhost:%d/health || exit 1", adminPort),
			},
			Interval: pulumi.String("30s"),
			Timeout:  pulumi.String("10s"),
			Retries:  pulumi.Int(3),
			StartPeriod: pulumi.String("60s"),
		},

		Labels: pulumi.StringMap{
			"environment": pulumi.String(gd.environment),
			"component":   pulumi.String("admin-gateway"),
			"service":     pulumi.String("gateway"),
			"managed-by":  pulumi.String("pulumi"),
			"dapr-app-id": pulumi.String("admin-gateway"),
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

	adminDaprContainer, err := gd.deployDaprSidecar("admin-gateway", adminPort, result, config)
	if err != nil {
		return nil, nil, err
	}

	return adminGatewayContainer, adminDaprContainer, nil
}

func (gd *GatewayDeployment) deployDaprSidecar(appName string, appPort int, result *GatewayDeploymentResult, config *GatewayConfiguration) (*docker.Container, error) {
	daprVersion := gd.config.Get("dapr_version")
	if daprVersion == "" {
		daprVersion = "1.12.0"
	}

	daprContainer, err := docker.NewContainer(gd.ctx, fmt.Sprintf("%s-dapr", appName), &docker.ContainerArgs{
		Name:    pulumi.Sprintf("%s-%s-dapr", gd.environment, appName),
		Image:   pulumi.Sprintf("daprio/daprd:%s", daprVersion),
		Restart: pulumi.String("unless-stopped"),

		Command: pulumi.StringArray{
			pulumi.String("./daprd"),
			pulumi.Sprintf("--app-id=%s", appName),
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
				External: pulumi.Int(config.DaprHTTPPort + appPort - 8080),
				Protocol: pulumi.String("tcp"),
			},
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(config.DaprGRPCPort),
				External: pulumi.Int(config.DaprGRPCPort + appPort - 8080),
				Protocol: pulumi.String("tcp"),
			},
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(9090),
				External: pulumi.Int(9090 + appPort - 8080),
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
				Name: result.GatewayNetwork.Name,
				Aliases: pulumi.StringArray{
					pulumi.Sprintf("%s-dapr", appName),
				},
			},
		},

		Labels: pulumi.StringMap{
			"environment": pulumi.String(gd.environment),
			"component":   pulumi.String("dapr"),
			"service":     pulumi.String("sidecar"),
			"app":         pulumi.String(appName),
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

func (gd *GatewayDeployment) getGatewayConfiguration() *GatewayConfiguration {
	identityPort := gd.config.RequireInt("identity_api_port")
	contentPort := gd.config.RequireInt("content_api_port")
	servicesPort := gd.config.RequireInt("services_api_port")

	return &GatewayConfiguration{
		ImageRegistry:      gd.config.Get("container_registry"),
		ImageTag:           gd.config.Get("image_tag"),
		RedisURL:           gd.config.Require("redis_url"),
		VaultURL:           gd.config.Require("vault_url"),
		DaprHTTPPort:       gd.config.RequireInt("dapr_http_port"),
		DaprGRPCPort:       gd.config.RequireInt("dapr_grpc_port"),
		EnableDebug:        gd.config.GetBool("enable_debugging"),
		EnableProfiling:    gd.config.GetBool("enable_profiling"),
		LogLevel:           gd.config.Get("log_level"),
		CorsOrigins:        []string{"http://localhost:3000", "http://localhost:8080"},
		RateLimitEnabled:   true,
		AuthEnabled:        true,
		IdentityAPIURL:     fmt.Sprintf("http://identity-api:%d", identityPort),
		ContentAPIURL:      fmt.Sprintf("http://content-api:%d", contentPort),
		ServicesAPIURL:     fmt.Sprintf("http://services-api:%d", servicesPort),
	}
}

func (gd *GatewayDeployment) ValidateDeployment(ctx context.Context, result *GatewayDeploymentResult) error {
	if result.PublicGatewayContainer == nil {
		return fmt.Errorf("Public Gateway container is not deployed")
	}

	if result.PublicDaprContainer == nil {
		return fmt.Errorf("Public Gateway Dapr sidecar container is not deployed")
	}

	if result.AdminGatewayContainer == nil {
		return fmt.Errorf("Admin Gateway container is not deployed")
	}

	if result.AdminDaprContainer == nil {
		return fmt.Errorf("Admin Gateway Dapr sidecar container is not deployed")
	}

	return nil
}

func (gd *GatewayDeployment) GetGatewayEndpoints() map[string]string {
	publicPort := gd.config.RequireInt("public_gateway_port")
	adminPort := gd.config.RequireInt("admin_gateway_port")

	return map[string]string{
		"public": fmt.Sprintf("http://localhost:%d", publicPort),
		"admin":  fmt.Sprintf("http://localhost:%d", adminPort),
	}
}