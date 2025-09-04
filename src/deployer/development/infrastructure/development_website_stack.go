package infrastructure

import (
	"context"
	"fmt"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	sharedconfig "github.com/axiom-software-co/international-center/src/deployer/shared/config"
	shared "github.com/axiom-software-co/international-center/src/deployer/shared/infrastructure"
)

type DevelopmentWebsiteStack struct {
	ctx             *pulumi.Context
	config          *config.Config
	configManager   *sharedconfig.ConfigManager
	environment     string
	errorHandler    *shared.ErrorHandler
	websiteConfig   *shared.WebsiteConfiguration // Cache configuration
	
	// Outputs
	WebsiteURL         pulumi.StringOutput `pulumi:"websiteUrl"`
	WebsiteName        pulumi.StringOutput `pulumi:"websiteName"`
	DeploymentStatus   pulumi.StringOutput `pulumi:"deploymentStatus"`
}

type DevelopmentWebsiteDeployment struct {
	ctx                  *pulumi.Context
	websiteConfiguration *shared.WebsiteConfiguration
	cdnResources        *DevelopmentWebsiteCDNResources // Cache CDN configuration
	
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
		ctx:           ctx,
		config:        config,
		configManager: configManager,
		environment:   environment,
		errorHandler:  errorHandler,
		websiteConfig: websiteConfig,
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

// createWebsiteDeployment creates the website deployment resource
func (ws *DevelopmentWebsiteStack) createWebsiteDeployment(config *shared.WebsiteConfiguration) (*DevelopmentWebsiteDeployment, error) {
	// Initialize CDN configuration for development environment
	cdnResources := &DevelopmentWebsiteCDNResources{
		distributionID: pulumi.String("dev-distribution-id").ToStringOutput(),
		cacheConfig: shared.WebsiteCacheConfig{
			Enabled:           config.EnableCaching, // Use config value
			MaxAge:            config.CacheMaxAge,   // Use config value
			StaticAssetsMaxAge: 3600, // 1 hour
			APIResponseMaxAge:  0,    // No API response caching in dev
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
		cdnResources:        cdnResources, // Cache CDN configuration
		GatewayEndpoint:     config.GatewayEndpoint,
	}
	
	// Set development-specific deployment outputs with error handling
	if config.ProjectName == "" {
		return nil, fmt.Errorf("project name cannot be empty")
	}
	
	deployment.PrimaryURL = pulumi.Sprintf("https://%s.pages.dev", config.ProjectName).ToStringOutput()
	deployment.PreviewURL = pulumi.Sprintf("https://%s-preview.pages.dev", config.ProjectName).ToStringOutput()
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