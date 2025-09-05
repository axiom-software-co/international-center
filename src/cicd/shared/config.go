package shared

import (
	"fmt"
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// EnvironmentConfig represents environment-specific configuration
type EnvironmentConfig struct {
	Environment string
	Azure       AzureConfig
	Cloudflare  CloudflareConfig
	HashiCorp   HashiCorpConfig
	Upstash     UpstashConfig
	Grafana     GrafanaConfig
}

// AzureConfig contains Azure-specific configuration
type AzureConfig struct {
	SubscriptionID string
	TenantID       string
	ClientID       string
	ClientSecret   string
	ResourceGroup  string
	Location       string
}

// CloudflareConfig contains Cloudflare-specific configuration
type CloudflareConfig struct {
	APIToken   string
	AccountID  string
	ZoneID     string
	DomainName string
}

// HashiCorpConfig contains HashiCorp Vault Cloud configuration
type HashiCorpConfig struct {
	OrganizationID string
	ProjectID      string
	AppName        string
	ClientID       string
	ClientSecret   string
}

// UpstashConfig contains Upstash Redis configuration
type UpstashConfig struct {
	APIKey    string
	Email     string
	DatabaseID string
	Endpoint   string
	Password   string
}

// GrafanaConfig contains Grafana Cloud configuration
type GrafanaConfig struct {
	APIKey    string
	StackSlug string
	OrgSlug   string
	URL       string
}

// LoadEnvironmentConfig loads configuration for the specified environment
func LoadEnvironmentConfig(ctx *pulumi.Context, environment string) (*EnvironmentConfig, error) {
	cfg := config.New(ctx, "")

	envConfig := &EnvironmentConfig{
		Environment: environment,
	}

	// Load Azure configuration
	if err := loadAzureConfig(cfg, envConfig, environment); err != nil {
		return nil, fmt.Errorf("failed to load Azure config: %w", err)
	}

	// Load Cloudflare configuration
	if err := loadCloudflareConfig(cfg, envConfig, environment); err != nil {
		return nil, fmt.Errorf("failed to load Cloudflare config: %w", err)
	}

	// Load HashiCorp configuration
	if err := loadHashiCorpConfig(cfg, envConfig, environment); err != nil {
		return nil, fmt.Errorf("failed to load HashiCorp config: %w", err)
	}

	// Load Upstash configuration
	if err := loadUpstashConfig(cfg, envConfig, environment); err != nil {
		return nil, fmt.Errorf("failed to load Upstash config: %w", err)
	}

	// Load Grafana configuration
	if err := loadGrafanaConfig(cfg, envConfig, environment); err != nil {
		return nil, fmt.Errorf("failed to load Grafana config: %w", err)
	}

	return envConfig, nil
}

// loadAzureConfig loads Azure-specific configuration
func loadAzureConfig(cfg *config.Config, envConfig *EnvironmentConfig, environment string) error {
	envConfig.Azure = AzureConfig{
		SubscriptionID: getSecretOrDefault(cfg, "azure:subscriptionId", os.Getenv("AZURE_SUBSCRIPTION_ID")),
		TenantID:       getSecretOrDefault(cfg, "azure:tenantId", os.Getenv("AZURE_TENANT_ID")),
		ClientID:       getSecretOrDefault(cfg, "azure:clientId", os.Getenv("AZURE_CLIENT_ID")),
		ClientSecret:   getSecretOrDefault(cfg, "azure:clientSecret", os.Getenv("AZURE_CLIENT_SECRET")),
		ResourceGroup:  getOrDefault(cfg, "azure:resourceGroup", fmt.Sprintf("international-center-%s", environment)),
		Location:       getOrDefault(cfg, "azure:location", "East US 2"),
	}

	return nil
}

// loadCloudflareConfig loads Cloudflare-specific configuration
func loadCloudflareConfig(cfg *config.Config, envConfig *EnvironmentConfig, environment string) error {
	domainName := "international-center.org"
	if environment == "staging" {
		domainName = "staging.international-center.org"
	}

	envConfig.Cloudflare = CloudflareConfig{
		APIToken:   getSecretOrDefault(cfg, "cloudflare:apiToken", os.Getenv("CLOUDFLARE_API_TOKEN")),
		AccountID:  getSecretOrDefault(cfg, "cloudflare:accountId", os.Getenv("CLOUDFLARE_ACCOUNT_ID")),
		ZoneID:     getSecretOrDefault(cfg, "cloudflare:zoneId", os.Getenv("CLOUDFLARE_ZONE_ID")),
		DomainName: domainName,
	}

	return nil
}

// loadHashiCorpConfig loads HashiCorp Vault Cloud configuration
func loadHashiCorpConfig(cfg *config.Config, envConfig *EnvironmentConfig, environment string) error {
	envConfig.HashiCorp = HashiCorpConfig{
		OrganizationID: getSecretOrDefault(cfg, "hashicorp:organizationId", os.Getenv("HCP_ORGANIZATION_ID")),
		ProjectID:      getSecretOrDefault(cfg, "hashicorp:projectId", os.Getenv("HCP_PROJECT_ID")),
		AppName:        fmt.Sprintf("international-center-%s", environment),
		ClientID:       getSecretOrDefault(cfg, "hashicorp:clientId", os.Getenv("HCP_CLIENT_ID")),
		ClientSecret:   getSecretOrDefault(cfg, "hashicorp:clientSecret", os.Getenv("HCP_CLIENT_SECRET")),
	}

	return nil
}

// loadUpstashConfig loads Upstash Redis configuration
func loadUpstashConfig(cfg *config.Config, envConfig *EnvironmentConfig, environment string) error {
	envConfig.Upstash = UpstashConfig{
		APIKey:     getSecretOrDefault(cfg, "upstash:apiKey", os.Getenv("UPSTASH_API_KEY")),
		Email:      getSecretOrDefault(cfg, "upstash:email", os.Getenv("UPSTASH_EMAIL")),
		DatabaseID: getSecretOrDefault(cfg, "upstash:databaseId", os.Getenv("UPSTASH_DATABASE_ID")),
		Endpoint:   getSecretOrDefault(cfg, "upstash:endpoint", os.Getenv("UPSTASH_REDIS_REST_URL")),
		Password:   getSecretOrDefault(cfg, "upstash:password", os.Getenv("UPSTASH_REDIS_REST_TOKEN")),
	}

	return nil
}

// loadGrafanaConfig loads Grafana Cloud configuration
func loadGrafanaConfig(cfg *config.Config, envConfig *EnvironmentConfig, environment string) error {
	stackSlug := fmt.Sprintf("international-center-%s", environment)
	if environment == "production" {
		stackSlug = "international-center-production"
	}

	envConfig.Grafana = GrafanaConfig{
		APIKey:    getSecretOrDefault(cfg, "grafana:apiKey", os.Getenv("GRAFANA_CLOUD_API_KEY")),
		StackSlug: stackSlug,
		OrgSlug:   "international-center",
		URL:       fmt.Sprintf("https://%s.grafana.net", stackSlug),
	}

	return nil
}

// getSecretOrDefault gets a secret value from Pulumi config or falls back to default
func getSecretOrDefault(cfg *config.Config, key, defaultValue string) string {
	// For secrets, we might get them as regular config values in development
	// In production, these would be actual secrets
	value := cfg.Get(key)
	if value != "" {
		return value
	}
	return defaultValue
}

// getOrDefault gets a regular config value or falls back to default
func getOrDefault(cfg *config.Config, key, defaultValue string) string {
	value := cfg.Get(key)
	if value != "" {
		return value
	}
	return defaultValue
}

// ValidateConfiguration validates that all required configuration is present
func ValidateConfiguration(envConfig *EnvironmentConfig, environment string) error {
	var missingConfigs []string

	// Validate Azure configuration (required for staging and production)
	if environment != "development" {
		if envConfig.Azure.SubscriptionID == "" {
			missingConfigs = append(missingConfigs, "Azure Subscription ID")
		}
		if envConfig.Azure.TenantID == "" {
			missingConfigs = append(missingConfigs, "Azure Tenant ID")
		}
		if envConfig.Azure.ClientID == "" {
			missingConfigs = append(missingConfigs, "Azure Client ID")
		}
		if envConfig.Azure.ClientSecret == "" {
			missingConfigs = append(missingConfigs, "Azure Client Secret")
		}
	}

	// Validate Cloudflare configuration (required for staging and production website)
	if environment != "development" {
		if envConfig.Cloudflare.APIToken == "" {
			missingConfigs = append(missingConfigs, "Cloudflare API Token")
		}
		if envConfig.Cloudflare.AccountID == "" {
			missingConfigs = append(missingConfigs, "Cloudflare Account ID")
		}
		if envConfig.Cloudflare.ZoneID == "" {
			missingConfigs = append(missingConfigs, "Cloudflare Zone ID")
		}
	}

	// Validate HashiCorp configuration (required for staging and production secrets)
	if environment != "development" {
		if envConfig.HashiCorp.OrganizationID == "" {
			missingConfigs = append(missingConfigs, "HashiCorp Organization ID")
		}
		if envConfig.HashiCorp.ProjectID == "" {
			missingConfigs = append(missingConfigs, "HashiCorp Project ID")
		}
		if envConfig.HashiCorp.ClientID == "" {
			missingConfigs = append(missingConfigs, "HashiCorp Client ID")
		}
		if envConfig.HashiCorp.ClientSecret == "" {
			missingConfigs = append(missingConfigs, "HashiCorp Client Secret")
		}
	}

	// Validate Grafana configuration (required for staging and production observability)
	if environment != "development" {
		if envConfig.Grafana.APIKey == "" {
			missingConfigs = append(missingConfigs, "Grafana Cloud API Key")
		}
	}

	if len(missingConfigs) > 0 {
		return fmt.Errorf("missing required configuration for %s environment: %v", environment, missingConfigs)
	}

	return nil
}