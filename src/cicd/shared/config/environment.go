package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Environment represents the deployment environment
type Environment string

const (
	EnvironmentDevelopment Environment = "development"
	EnvironmentStaging     Environment = "staging"
	EnvironmentProduction  Environment = "production"
)

// EnvironmentConfig holds environment-specific configuration
type EnvironmentConfig struct {
	Environment Environment `json:"environment"`
	Region      string      `json:"region"`
	ProjectName string      `json:"project_name"`
	StackName   string      `json:"stack_name"`
	Tags        map[string]string `json:"tags"`
	
	// Resource naming configuration
	ResourcePrefix string `json:"resource_prefix"`
	ResourceSuffix string `json:"resource_suffix"`
	
	// Network configuration
	NetworkCIDR   string `json:"network_cidr"`
	SubnetCIDRs   []string `json:"subnet_cidrs"`
	
	// Security configuration
	AllowedIPs    []string `json:"allowed_ips"`
	EnableHTTPS   bool     `json:"enable_https"`
	
	// Resource limits
	ResourceLimits ResourceLimits `json:"resource_limits"`
	
	// Feature flags
	Features FeatureFlags `json:"features"`
}

// ResourceLimits defines resource constraints per environment
type ResourceLimits struct {
	MaxContainers     int   `json:"max_containers"`
	MaxCPUCores      int   `json:"max_cpu_cores"`
	MaxMemoryGB      int   `json:"max_memory_gb"`
	MaxStorageGB     int   `json:"max_storage_gb"`
	MaxConnections   int   `json:"max_connections"`
}

// FeatureFlags defines feature toggles per environment
type FeatureFlags struct {
	EnableDebugLogging    bool `json:"enable_debug_logging"`
	EnableMetricsExport   bool `json:"enable_metrics_export"`
	EnableTracingExport   bool `json:"enable_tracing_export"`
	EnableBackupValidation bool `json:"enable_backup_validation"`
	EnableCacheLayer      bool `json:"enable_cache_layer"`
	EnableAuditLogging    bool `json:"enable_audit_logging"`
}

// DetectEnvironment detects the current environment from environment variables
func DetectEnvironment() (Environment, error) {
	envStr := strings.ToLower(strings.TrimSpace(os.Getenv("ENVIRONMENT")))
	if envStr == "" {
		// Check alternative environment variable names
		if alt := os.Getenv("DEPLOY_ENV"); alt != "" {
			envStr = strings.ToLower(strings.TrimSpace(alt))
		} else if alt := os.Getenv("NODE_ENV"); alt != "" {
			envStr = strings.ToLower(strings.TrimSpace(alt))
		}
	}
	
	if envStr == "" {
		return "", errors.New("environment not specified - set ENVIRONMENT variable")
	}
	
	switch envStr {
	case "development", "dev", "local":
		return EnvironmentDevelopment, nil
	case "staging", "stage", "test":
		return EnvironmentStaging, nil
	case "production", "prod":
		return EnvironmentProduction, nil
	default:
		return "", fmt.Errorf("unknown environment: %s", envStr)
	}
}

// ValidateEnvironment validates the environment configuration
func ValidateEnvironment(env Environment) error {
	switch env {
	case EnvironmentDevelopment, EnvironmentStaging, EnvironmentProduction:
		return nil
	default:
		return fmt.Errorf("invalid environment: %s", env)
	}
}

// GetEnvironmentConfig returns environment-specific configuration
func GetEnvironmentConfig(env Environment) (*EnvironmentConfig, error) {
	if err := ValidateEnvironment(env); err != nil {
		return nil, err
	}
	
	baseConfig := &EnvironmentConfig{
		Environment: env,
		ProjectName: getRequiredEnv("PROJECT_NAME", "international-center"),
		Region:      getRequiredEnv("REGION", "eastus"),
		Tags: map[string]string{
			"project":     "international-center",
			"environment": string(env),
			"managed-by":  "pulumi",
		},
	}
	
	// Environment-specific configuration
	switch env {
	case EnvironmentDevelopment:
		return configureDevelopment(baseConfig), nil
	case EnvironmentStaging:
		return configureStaging(baseConfig), nil
	case EnvironmentProduction:
		return configureProduction(baseConfig), nil
	default:
		return nil, fmt.Errorf("unsupported environment: %s", env)
	}
}

// configureDevelopment configures development environment
func configureDevelopment(base *EnvironmentConfig) *EnvironmentConfig {
	base.StackName = getRequiredEnv("STACK_NAME", "dev")
	base.ResourcePrefix = "dev"
	base.ResourceSuffix = ""
	base.NetworkCIDR = "10.0.0.0/16"
	base.SubnetCIDRs = []string{"10.0.1.0/24", "10.0.2.0/24"}
	base.AllowedIPs = []string{"0.0.0.0/0"} // Allow all for development
	base.EnableHTTPS = false
	
	base.ResourceLimits = ResourceLimits{
		MaxContainers:  10,
		MaxCPUCores:   4,
		MaxMemoryGB:   8,
		MaxStorageGB:  50,
		MaxConnections: 100,
	}
	
	base.Features = FeatureFlags{
		EnableDebugLogging:    true,
		EnableMetricsExport:   true,
		EnableTracingExport:   true,
		EnableBackupValidation: false, // Skip backup validation in dev
		EnableCacheLayer:      true,
		EnableAuditLogging:    false, // Minimal audit logging in dev
	}
	
	return base
}

// configureStaging configures staging environment
func configureStaging(base *EnvironmentConfig) *EnvironmentConfig {
	base.StackName = getRequiredEnv("STACK_NAME", "staging")
	base.ResourcePrefix = "staging"
	base.ResourceSuffix = ""
	base.NetworkCIDR = "10.1.0.0/16"
	base.SubnetCIDRs = []string{"10.1.1.0/24", "10.1.2.0/24"}
	base.AllowedIPs = getStringSliceEnv("ALLOWED_IPS", []string{"10.0.0.0/8"})
	base.EnableHTTPS = true
	
	base.ResourceLimits = ResourceLimits{
		MaxContainers:  20,
		MaxCPUCores:   8,
		MaxMemoryGB:   16,
		MaxStorageGB:  100,
		MaxConnections: 500,
	}
	
	base.Features = FeatureFlags{
		EnableDebugLogging:    false,
		EnableMetricsExport:   true,
		EnableTracingExport:   true,
		EnableBackupValidation: true,
		EnableCacheLayer:      true,
		EnableAuditLogging:    true,
	}
	
	return base
}

// configureProduction configures production environment
func configureProduction(base *EnvironmentConfig) *EnvironmentConfig {
	base.StackName = getRequiredEnv("STACK_NAME", "production")
	base.ResourcePrefix = "prod"
	base.ResourceSuffix = ""
	base.NetworkCIDR = "10.2.0.0/16"
	base.SubnetCIDRs = []string{"10.2.1.0/24", "10.2.2.0/24"}
	base.AllowedIPs = getStringSliceEnv("ALLOWED_IPS", []string{}) // Must be explicitly configured
	base.EnableHTTPS = true
	
	base.ResourceLimits = ResourceLimits{
		MaxContainers:  50,
		MaxCPUCores:   16,
		MaxMemoryGB:   32,
		MaxStorageGB:  500,
		MaxConnections: 2000,
	}
	
	base.Features = FeatureFlags{
		EnableDebugLogging:    false,
		EnableMetricsExport:   true,
		EnableTracingExport:   true,
		EnableBackupValidation: true,
		EnableCacheLayer:      true,
		EnableAuditLogging:    true,
	}
	
	return base
}

// IsValidEnvironment checks if the environment is valid
func IsValidEnvironment(env Environment) bool {
	return ValidateEnvironment(env) == nil
}

// IsDevelopment returns true if environment is development
func (e Environment) IsDevelopment() bool {
	return e == EnvironmentDevelopment
}

// IsStaging returns true if environment is staging
func (e Environment) IsStaging() bool {
	return e == EnvironmentStaging
}

// IsProduction returns true if environment is production
func (e Environment) IsProduction() bool {
	return e == EnvironmentProduction
}

// RequiresApproval returns true if environment requires manual approval
func (e Environment) RequiresApproval() bool {
	return e == EnvironmentProduction
}

// AllowsAggressiveMigration returns true if aggressive migrations are allowed
func (e Environment) AllowsAggressiveMigration() bool {
	return e == EnvironmentDevelopment
}

// AllowsAutoRollback returns true if automatic rollback is allowed
func (e Environment) AllowsAutoRollback() bool {
	return e == EnvironmentDevelopment
}

// Helper functions

func getRequiredEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getStringSliceEnv(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}