package infrastructure

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	sharedconfig "github.com/axiom-software-co/international-center/src/deployer/shared/config"
	shared "github.com/axiom-software-co/international-center/src/deployer/shared/infrastructure"
)

type DevelopmentWebsiteStack struct {
	ctx             *pulumi.Context
	config          *config.Config
	configManager   *sharedconfig.ConfigManager
	containerRuntime *shared.ContainerRuntime
	environment     string
	errorHandler    *shared.ErrorHandler
	websiteConfig   *shared.WebsiteConfiguration // Cache configuration
	projectRoot     string
	
	// Outputs
	WebsiteURL         pulumi.StringOutput `pulumi:"websiteUrl"`
	WebsiteName        pulumi.StringOutput `pulumi:"websiteName"`
	DeploymentStatus   pulumi.StringOutput `pulumi:"deploymentStatus"`
}

type DevelopmentWebsiteDeployment struct {
	ctx                  *pulumi.Context
	websiteConfiguration *shared.WebsiteConfiguration
	cdnResources        *DevelopmentWebsiteCDNResources // Cache CDN configuration
	
	// Container Resources
	WebsiteImage      *docker.Image
	WebsiteContainer  *docker.Container
	WebsiteNetwork    *docker.Network
	DaprSidecar       *docker.Container
	
	// Outputs
	PrimaryURL         pulumi.StringOutput `pulumi:"primaryUrl"`
	PreviewURL         pulumi.StringOutput `pulumi:"previewUrl"`
	DeploymentStatus   pulumi.StringOutput `pulumi:"deploymentStatus"`
	GatewayEndpoint    string
}

// CDNResources implementation for development environment
type DevelopmentWebsiteCDNResources struct {
	distributionID     pulumi.StringOutput
	cacheConfig        shared.WebsiteCacheConfig
	sslCertificate     pulumi.StringOutput
	domainConfig       shared.WebsiteDomainConfig
}

func NewWebsiteStack(ctx *pulumi.Context, config *config.Config, environment string) shared.WebsiteStack {
	errorHandler := shared.NewErrorHandler(ctx, environment, "website")
	
	// Create ConfigManager for centralized configuration
	configManager, err := sharedconfig.NewConfigManager(ctx)
	if err != nil {
		configErr := shared.NewConfigurationError("create_config_manager", "website", environment, "ConfigManager", err)
		errorHandler.HandleError(configErr)
		configManager = nil // Fallback to legacy configuration
	}
	
	// Initialize container runtime for website deployment
	containerRuntime := shared.NewContainerRuntime(ctx, environment, "development-network", 3000)
	
	// Determine project root (get working directory)
	projectRoot := "."
	
	// Cache website configuration for performance
	websiteConfig := shared.GetWebsiteConfiguration(environment, config)
	if websiteConfig == nil {
		configErr := shared.NewConfigurationError("get_website_config", "website", environment, "WebsiteConfiguration", fmt.Errorf("failed to load website configuration"))
		errorHandler.HandleError(configErr)
		// Provide default configuration as fallback
		websiteConfig = &shared.WebsiteConfiguration{
			Environment:           environment,
			ProjectName:          "international-center-" + environment,
			BuildCommand:         "npm run build",
			BuildOutput:          "dist",
			SourcePath:           "website/",
			GatewayEndpoint:      "http://localhost:8080",
			EnableSSL:            false,
			EnableCaching:        false,
			CacheMaxAge:          300,
			PreviewDeployments:   true,
			AutoDeployBranches:   []string{"main"},
			EnvironmentVariables: map[string]string{
				"NODE_ENV": environment,
				"API_BASE_URL": "http://localhost:8080",
			},
		}
	}
	
	component := &DevelopmentWebsiteStack{
		ctx:             ctx,
		config:          config,
		configManager:   configManager,
		containerRuntime: containerRuntime,
		environment:     environment,
		errorHandler:    errorHandler,
		websiteConfig:   websiteConfig,
		projectRoot:     projectRoot,
	}
	
	return component
}

func (ws *DevelopmentWebsiteStack) Deploy(ctx context.Context) (shared.WebsiteDeployment, error) {
	// Use cached configuration for better performance
	websiteConfig := ws.websiteConfig
	
	// Validate configuration before deployment
	if err := ws.validateConfiguration(websiteConfig); err != nil {
		deploymentErr := shared.NewDeploymentError("validate_config", "website", ws.environment, err)
		ws.errorHandler.HandleError(deploymentErr)
		return nil, fmt.Errorf("website configuration validation failed: %w", err)
	}
	
	// For development environment, we simulate Cloudflare Pages deployment
	// In a real implementation, this would use the Cloudflare Pulumi provider
	deployment, err := ws.createWebsiteDeployment(websiteConfig)
	if err != nil {
		deploymentErr := shared.NewDeploymentError("create_deployment", "website", ws.environment, err)
		ws.errorHandler.HandleError(deploymentErr)
		return nil, fmt.Errorf("failed to create website deployment: %w", err)
	}
	
	// Set stack outputs for component-first architecture
	ws.WebsiteURL = deployment.GetPrimaryURL()
	ws.WebsiteName = pulumi.String(websiteConfig.ProjectName).ToStringOutput()
	ws.DeploymentStatus = deployment.GetDeploymentStatus()
	
	return deployment, nil
}

// validateConfiguration validates the website configuration
func (ws *DevelopmentWebsiteStack) validateConfiguration(config *shared.WebsiteConfiguration) error {
	if config == nil {
		return fmt.Errorf("website configuration is nil")
	}
	if config.ProjectName == "" {
		return fmt.Errorf("project name is required")
	}
	if config.BuildCommand == "" {
		return fmt.Errorf("build command is required")
	}
	if config.BuildOutput == "" {
		return fmt.Errorf("build output directory is required")
	}
	return nil
}

// createWebsiteDeployment creates the website deployment resource with actual containers
func (ws *DevelopmentWebsiteStack) createWebsiteDeployment(config *shared.WebsiteConfiguration) (*DevelopmentWebsiteDeployment, error) {
	if ws.containerRuntime == nil {
		return nil, fmt.Errorf("container runtime not initialized")
	}

	// Create website network
	websiteNetwork, err := docker.NewNetwork(ws.ctx, "website-network", &docker.NetworkArgs{
		Name:     pulumi.Sprintf("%s-website-network", ws.environment),
		Driver:   pulumi.String("bridge"),
		Internal: pulumi.Bool(false),
		Labels: docker.NetworkLabelArray{
			&docker.NetworkLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(ws.environment),
			},
			&docker.NetworkLabelArgs{
				Label: pulumi.String("service"),
				Value: pulumi.String("website"),
			},
			&docker.NetworkLabelArgs{
				Label: pulumi.String("managed-by"),
				Value: pulumi.String("pulumi"),
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create website network: %w", err)
	}

	// Build website image
	websiteImage, err := ws.buildWebsiteImage()
	if err != nil {
		return nil, fmt.Errorf("failed to build website image: %w", err)
	}

	// Deploy website container
	websiteContainer, err := ws.deployWebsiteContainer(websiteImage, websiteNetwork, config)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy website container: %w", err)
	}

	// Deploy Dapr sidecar for service mesh integration
	daprSidecar, err := ws.deployDaprSidecar(websiteContainer, websiteNetwork)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy Dapr sidecar: %w", err)
	}

	// Initialize CDN configuration for development environment
	cdnResources := &DevelopmentWebsiteCDNResources{
		distributionID: pulumi.String("dev-distribution-id").ToStringOutput(),
		cacheConfig: shared.WebsiteCacheConfig{
			Enabled:           config.EnableCaching,
			MaxAge:            config.CacheMaxAge,
			StaticAssetsMaxAge: 3600,
			APIResponseMaxAge:  0,
			CompressionEnabled: false,
		},
		sslCertificate: pulumi.String("dev-ssl-cert").ToStringOutput(),
		domainConfig: shared.WebsiteDomainConfig{
			CustomDomain:      config.CustomDomain,
			SubdomainPrefix:   "dev",
			SSLCertificate:    "dev-ssl-cert",
			DNSConfiguration:  map[string]string{},
		},
	}

	deployment := &DevelopmentWebsiteDeployment{
		ctx:                  ws.ctx,
		websiteConfiguration: config,
		cdnResources:        cdnResources,
		WebsiteImage:        websiteImage,
		WebsiteContainer:    websiteContainer,
		WebsiteNetwork:      websiteNetwork,
		DaprSidecar:         daprSidecar,
		GatewayEndpoint:     config.GatewayEndpoint,
	}

	// Set container-based deployment outputs
	deployment.PrimaryURL = pulumi.String("http://localhost:3000").ToStringOutput()
	deployment.PreviewURL = pulumi.String("http://localhost:3000").ToStringOutput()
	deployment.DeploymentStatus = pulumi.String("deployed").ToStringOutput()

	return deployment, nil
}

func (ws *DevelopmentWebsiteStack) GetWebsiteURL() pulumi.StringOutput {
	return ws.WebsiteURL
}

func (ws *DevelopmentWebsiteStack) GetDeploymentConfiguration() *shared.WebsiteConfiguration {
	// Return cached configuration for better performance
	return ws.websiteConfig
}

func (ws *DevelopmentWebsiteStack) ValidateDeployment(ctx context.Context, deployment shared.WebsiteDeployment) error {
	// Development environment validation
	if deployment == nil {
		return fmt.Errorf("website deployment is nil")
	}
	
	// Validate deployment has required outputs
	primaryURL := deployment.GetPrimaryURL()
	if primaryURL.ElementType() == nil {
		return fmt.Errorf("website deployment missing primary URL")
	}
	
	return nil
}

func (ws *DevelopmentWebsiteStack) ConfigureGatewayIntegration(gatewayEndpoint string) error {
	// Configure website to use the provided gateway endpoint
	// In development, this would typically be localhost
	if gatewayEndpoint == "" {
		return fmt.Errorf("gateway endpoint cannot be empty")
	}
	
	// Validate gateway endpoint format for development environment
	if ws.environment == "development" && !strings.Contains(gatewayEndpoint, "localhost") && !strings.Contains(gatewayEndpoint, "127.0.0.1") {
		ws.errorHandler.HandleError(shared.NewValidationError("invalid_dev_gateway", "website", ws.environment, "development gateway should use localhost"))
	}
	
	// Update cached configuration with gateway endpoint
	ws.websiteConfig.GatewayEndpoint = gatewayEndpoint
	
	// Update environment variables to reflect the new gateway endpoint
	if ws.websiteConfig.EnvironmentVariables == nil {
		ws.websiteConfig.EnvironmentVariables = make(map[string]string)
	}
	ws.websiteConfig.EnvironmentVariables["API_BASE_URL"] = gatewayEndpoint
	
	return nil
}

// DevelopmentWebsiteDeployment methods
func (d *DevelopmentWebsiteDeployment) GetPrimaryURL() pulumi.StringOutput {
	return d.PrimaryURL
}

func (d *DevelopmentWebsiteDeployment) GetPreviewURL() pulumi.StringOutput {
	return d.PreviewURL
}

func (d *DevelopmentWebsiteDeployment) GetDeploymentStatus() pulumi.StringOutput {
	return d.DeploymentStatus
}

func (d *DevelopmentWebsiteDeployment) GetCDNConfiguration() shared.WebsiteCDNResources {
	// Return cached CDN configuration for better performance
	return d.cdnResources
}

func (d *DevelopmentWebsiteDeployment) GetGatewayEndpoint() string {
	return d.GatewayEndpoint
}

// DevelopmentWebsiteCDNResources methods
func (cdn *DevelopmentWebsiteCDNResources) GetDistributionID() pulumi.StringOutput {
	return cdn.distributionID
}

func (cdn *DevelopmentWebsiteCDNResources) GetCacheConfiguration() shared.WebsiteCacheConfig {
	return cdn.cacheConfig
}

func (cdn *DevelopmentWebsiteCDNResources) GetSSLCertificate() pulumi.StringOutput {
	return cdn.sslCertificate
}

func (cdn *DevelopmentWebsiteCDNResources) GetDomainConfiguration() shared.WebsiteDomainConfig {
	return cdn.domainConfig
}

// buildWebsiteImage builds the website container image
func (ws *DevelopmentWebsiteStack) buildWebsiteImage() (*docker.Image, error) {
	return ws.containerRuntime.BuildImage(
		"website",
		ws.projectRoot,
		filepath.Join(ws.projectRoot, "src/website/Containerfile"),
		map[string]string{
			"APP_VERSION": ws.config.Get("app_version"),
			"BUILD_TIME":  fmt.Sprintf("%d", time.Now().Unix()),
			"GIT_COMMIT":  "development",
		},
	)
}

// deployWebsiteContainer deploys the website container
func (ws *DevelopmentWebsiteStack) deployWebsiteContainer(image *docker.Image, network *docker.Network, config *shared.WebsiteConfiguration) (*docker.Container, error) {
	return docker.NewContainer(ws.ctx, "website-container", &docker.ContainerArgs{
		Name:    pulumi.Sprintf("%s-website", ws.environment),
		Image:   image.ImageName,
		Restart: pulumi.String("unless-stopped"),
		
		Ports: docker.ContainerPortArray{
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(3000),
				External: pulumi.Int(3000),
				Protocol: pulumi.String("tcp"),
			},
		},
		
		Envs: pulumi.StringArray{
			pulumi.String("NODE_ENV=development"),
			pulumi.String("HOST=0.0.0.0"),
			pulumi.String("PORT=3000"),
			pulumi.String("ASTRO_TELEMETRY_DISABLED=1"),
			pulumi.Sprintf("API_BASE_URL=%s", config.GatewayEndpoint),
			pulumi.String("SERVICES_API_URL=http://localhost:8081"),
			pulumi.String("DAPR_HTTP_PORT=3500"),
			pulumi.String("DAPR_APP_ID=website"),
			pulumi.Sprintf("NETWORK_NAME=%s-website-network", ws.environment),
		},
		
		NetworksAdvanced: docker.ContainerNetworksAdvancedArray{
			&docker.ContainerNetworksAdvancedArgs{
				Name: network.Name,
			},
		},
		
		Labels: docker.ContainerLabelArray{
			&docker.ContainerLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(ws.environment),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("service"),
				Value: pulumi.String("website"),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("managed-by"),
				Value: pulumi.String("pulumi"),
			},
		},
		
		Healthcheck: &docker.ContainerHealthcheckArgs{
			Tests: pulumi.StringArray{
				pulumi.String("CMD"),
				pulumi.String("curl"),
				pulumi.String("-f"),
				pulumi.String("http://localhost:3000/health"),
			},
			Interval:    pulumi.String("30s"),
			Timeout:     pulumi.String("10s"),
			StartPeriod: pulumi.String("5s"),
			Retries:     pulumi.Int(3),
		},
	})
}

// deployDaprSidecar deploys the Dapr sidecar container for service mesh integration
func (ws *DevelopmentWebsiteStack) deployDaprSidecar(websiteContainer *docker.Container, network *docker.Network) (*docker.Container, error) {
	daprVersion := ws.config.Get("dapr_version")
	if daprVersion == "" {
		daprVersion = "1.12.0"
	}
	
	return docker.NewContainer(ws.ctx, "website-dapr-sidecar", &docker.ContainerArgs{
		Name:    pulumi.Sprintf("%s-website-dapr-sidecar", ws.environment),
		Image:   pulumi.Sprintf("daprio/daprd:%s", daprVersion),
		Restart: pulumi.String("unless-stopped"),
		
		Command: pulumi.StringArray{
			pulumi.String("./daprd"),
			pulumi.String("--app-id=website"),
			pulumi.String("--app-port=3000"),
			pulumi.String("--dapr-http-port=3500"),
			pulumi.String("--dapr-grpc-port=50001"),
			pulumi.String("--placement-host-address=dapr-placement:50005"),
			pulumi.String("--components-path=/components"),
			pulumi.String("--log-level=info"),
			pulumi.String("--app-ssl=false"),
		},
		
		Ports: docker.ContainerPortArray{
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(3500),
				External: pulumi.Int(3500),
				Protocol: pulumi.String("tcp"),
			},
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(50001),
				External: pulumi.Int(50001),
				Protocol: pulumi.String("tcp"),
			},
		},
		
		NetworksAdvanced: docker.ContainerNetworksAdvancedArray{
			&docker.ContainerNetworksAdvancedArgs{
				Name: network.Name,
			},
		},
		
		Labels: docker.ContainerLabelArray{
			&docker.ContainerLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(ws.environment),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("service"),
				Value: pulumi.String("website-dapr"),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("managed-by"),
				Value: pulumi.String("pulumi"),
			},
		},
		
		// DependsOn not supported in this Pulumi Docker version
	})
}