package infrastructure

import (
	"context"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type WebsiteStack interface {
	Deploy(ctx context.Context) (WebsiteDeployment, error)
	GetWebsiteURL() pulumi.StringOutput
	GetDeploymentConfiguration() *WebsiteConfiguration
	ValidateDeployment(ctx context.Context, deployment WebsiteDeployment) error
	ConfigureGatewayIntegration(gatewayEndpoint string) error
}

type WebsiteDeployment interface {
	GetPrimaryURL() pulumi.StringOutput
	GetPreviewURL() pulumi.StringOutput
	GetDeploymentStatus() pulumi.StringOutput
	GetCDNConfiguration() WebsiteCDNResources
	GetGatewayEndpoint() string
}

type WebsiteCDNResources interface {
	GetDistributionID() pulumi.StringOutput
	GetCacheConfiguration() WebsiteCacheConfig
	GetSSLCertificate() pulumi.StringOutput
	GetDomainConfiguration() WebsiteDomainConfig
}

type WebsiteFactory interface {
	CreateWebsiteStack(ctx *pulumi.Context, config *config.Config, environment string) WebsiteStack
}

type WebsiteConfiguration struct {
	Environment           string
	ProjectName          string
	RepositoryURL        string
	BuildCommand         string
	BuildOutput          string
	SourcePath           string
	GatewayEndpoint      string
	CustomDomain         string
	EnableSSL            bool
	EnableCaching        bool
	CacheMaxAge          int
	PreviewDeployments   bool
	AutoDeployBranches   []string
	EnvironmentVariables map[string]string
}

// WebsiteCacheConfig represents caching configuration for website deployment
type WebsiteCacheConfig struct {
	Enabled           bool
	MaxAge            int
	StaticAssetsMaxAge int
	APIResponseMaxAge int
	CompressionEnabled bool
}

// WebsiteDomainConfig represents domain configuration for website deployment  
type WebsiteDomainConfig struct {
	CustomDomain      string
	SubdomainPrefix   string
	SSLCertificate    string
	DNSConfiguration  map[string]string
}

func GetWebsiteConfiguration(environment string, config *config.Config) *WebsiteConfiguration {
	switch environment {
	case "development":
		return &WebsiteConfiguration{
			Environment:           "development",
			ProjectName:          "international-center-dev",
			RepositoryURL:        config.Get("website_repo_url"),
			BuildCommand:         "npm run build",
			BuildOutput:          "dist",
			SourcePath:           "website/",
			GatewayEndpoint:      "http://localhost:8080",
			CustomDomain:         "",
			EnableSSL:            false,
			EnableCaching:        false,
			CacheMaxAge:          300, // 5 minutes for dev
			PreviewDeployments:   true,
			AutoDeployBranches:   []string{"main", "develop"},
			EnvironmentVariables: map[string]string{
				"NODE_ENV": "development",
				"API_BASE_URL": "http://localhost:8080",
			},
		}
	case "staging":
		return &WebsiteConfiguration{
			Environment:           "staging",
			ProjectName:          "international-center-staging",
			RepositoryURL:        "",
			BuildCommand:         "npm run build:staging",
			BuildOutput:          "dist",
			SourcePath:           "website/",
			GatewayEndpoint:      "",
			CustomDomain:         "staging.internationalcenter.org",
			EnableSSL:            true,
			EnableCaching:        true,
			CacheMaxAge:          1800, // 30 minutes for staging
			PreviewDeployments:   true,
			AutoDeployBranches:   []string{"staging"},
			EnvironmentVariables: map[string]string{
				"NODE_ENV": "staging",
				"API_BASE_URL": "",
			},
		}
	case "production":
		return &WebsiteConfiguration{
			Environment:           "production",
			ProjectName:          "international-center-production",
			RepositoryURL:        "",
			BuildCommand:         "npm run build:production",
			BuildOutput:          "dist",
			SourcePath:           "website/",
			GatewayEndpoint:      "",
			CustomDomain:         "www.internationalcenter.org",
			EnableSSL:            true,
			EnableCaching:        true,
			CacheMaxAge:          3600, // 1 hour for production
			PreviewDeployments:   false,
			AutoDeployBranches:   []string{"main"},
			EnvironmentVariables: map[string]string{
				"NODE_ENV": "production",
				"API_BASE_URL": "",
			},
		}
	default:
		return &WebsiteConfiguration{
			Environment:           environment,
			ProjectName:          "international-center-" + environment,
			RepositoryURL:        "",
			BuildCommand:         "npm run build",
			BuildOutput:          "dist",
			SourcePath:           "website/",
			GatewayEndpoint:      "",
			CustomDomain:         "",
			EnableSSL:            true,
			EnableCaching:        true,
			CacheMaxAge:          1800,
			PreviewDeployments:   true,
			AutoDeployBranches:   []string{"main"},
			EnvironmentVariables: map[string]string{
				"NODE_ENV": environment,
				"API_BASE_URL": "",
			},
		}
	}
}

// WebsiteMetrics defines website performance and availability metrics for environment-specific policies
type WebsiteMetrics struct {
	MaxBuildTimeMinutes    int
	MaxDeployTimeMinutes   int
	MaxPageLoadTimeMS      int
	RequiredUptimePercent  float64
	MaxBandwidthGB         int
	EnablePerformanceMonitoring bool
	EnableAccessLogs       bool
	EnableErrorTracking    bool
}

func GetWebsiteMetrics(environment string) WebsiteMetrics {
	switch environment {
	case "development":
		return WebsiteMetrics{
			MaxBuildTimeMinutes:    10,
			MaxDeployTimeMinutes:   5,
			MaxPageLoadTimeMS:      5000,
			RequiredUptimePercent:  95.0,
			MaxBandwidthGB:         10,
			EnablePerformanceMonitoring: false,
			EnableAccessLogs:       false,
			EnableErrorTracking:    true,
		}
	case "staging":
		return WebsiteMetrics{
			MaxBuildTimeMinutes:    15,
			MaxDeployTimeMinutes:   10,
			MaxPageLoadTimeMS:      3000,
			RequiredUptimePercent:  98.0,
			MaxBandwidthGB:         100,
			EnablePerformanceMonitoring: true,
			EnableAccessLogs:       true,
			EnableErrorTracking:    true,
		}
	case "production":
		return WebsiteMetrics{
			MaxBuildTimeMinutes:    20,
			MaxDeployTimeMinutes:   15,
			MaxPageLoadTimeMS:      2000,
			RequiredUptimePercent:  99.9,
			MaxBandwidthGB:         1000,
			EnablePerformanceMonitoring: true,
			EnableAccessLogs:       true,
			EnableErrorTracking:    true,
		}
	default:
		return WebsiteMetrics{
			MaxBuildTimeMinutes:    10,
			MaxDeployTimeMinutes:   5,
			MaxPageLoadTimeMS:      4000,
			RequiredUptimePercent:  97.0,
			MaxBandwidthGB:         50,
			EnablePerformanceMonitoring: false,
			EnableAccessLogs:       true,
			EnableErrorTracking:    true,
		}
	}
}