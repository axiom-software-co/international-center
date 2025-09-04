package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ConfigManager provides centralized configuration management across all environments
type ConfigManager struct {
	environment Environment
	envConfig   *EnvironmentConfig
	pulumiConfig *PulumiConfig
	
	// Cached configuration values to avoid repeated environment variable access
	databaseConfig *RuntimeDatabaseConfig
	redisConfig    *RuntimeRedisConfig
	storageConfig  *RuntimeStorageConfig
	vaultConfig    *RuntimeVaultConfig
	daprConfig     *RuntimeDaprConfig
	serviceConfig  *RuntimeServiceConfig
}

// RuntimeDatabaseConfig holds runtime database configuration values
type RuntimeDatabaseConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
	URL      string
	ContainerHost string
	ContainerURL  string
}

// RuntimeRedisConfig holds runtime Redis configuration values  
type RuntimeRedisConfig struct {
	Host          string
	Port          int
	Password      string
	Address       string
	ContainerHost string
	ContainerAddr string
}

// RuntimeStorageConfig holds runtime storage configuration values
type RuntimeStorageConfig struct {
	AzuriteHost        string
	AzuritePort        string
	BlobEndpoint       string
	ConnectionString   string
}

// RuntimeVaultConfig holds runtime Vault configuration values
type RuntimeVaultConfig struct {
	Address   string
	Token     string
	Namespace string
}

// RuntimeDaprConfig holds runtime Dapr configuration values
type RuntimeDaprConfig struct {
	HTTPEndpoint      string
	GRPCEndpoint      string
	PlacementEndpoint string
	ComponentsPath    string
}

// RuntimeServiceConfig holds runtime service configuration values
type RuntimeServiceConfig struct {
	Host             string
	ContentAPIURL    string
	ServicesAPIURL   string
	PublicGatewayURL string
	AdminGatewayURL  string
}

// NewConfigManager creates a new centralized configuration manager
func NewConfigManager(ctx *pulumi.Context) (*ConfigManager, error) {
	// Load environment variables from .env file first
	if err := loadEnvironmentFile(); err != nil {
		ctx.Log.Warn(fmt.Sprintf("Failed to load environment file: %v", err), nil)
	}
	
	// Detect environment
	env, err := DetectEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to detect environment: %w", err)
	}
	
	// Get environment-specific configuration
	envConfig, err := GetEnvironmentConfig(env)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment config: %w", err)
	}
	
	// Create Pulumi configuration wrapper
	pulumiConfig := NewPulumiConfig(ctx, env)
	
	manager := &ConfigManager{
		environment:  env,
		envConfig:    envConfig,
		pulumiConfig: pulumiConfig,
	}
	
	// Initialize runtime configurations
	if err := manager.initializeRuntimeConfigs(); err != nil {
		return nil, fmt.Errorf("failed to initialize runtime configs: %w", err)
	}
	
	return manager, nil
}

// loadEnvironmentFile loads environment variables from the appropriate .env file
func loadEnvironmentFile() error {
	// Determine which .env file to load based on environment
	envFile := ".env.development" // Default to development
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		switch strings.ToLower(env) {
		case "staging", "stage":
			envFile = ".env.staging"
		case "production", "prod":
			envFile = ".env.production"
		}
	}
	
	file, err := os.Open(envFile)
	if err != nil {
		return nil // File doesn't exist, continue without loading
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// Remove quotes if present
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
		   (strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}
		
		// Only set if not already set (environment variables take precedence)
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
	
	return scanner.Err()
}

// initializeRuntimeConfigs initializes all runtime configuration caches
func (cm *ConfigManager) initializeRuntimeConfigs() error {
	var err error
	
	// Initialize database configuration
	if cm.databaseConfig, err = cm.loadDatabaseConfig(); err != nil {
		return fmt.Errorf("failed to load database config: %w", err)
	}
	
	// Initialize Redis configuration
	if cm.redisConfig, err = cm.loadRedisConfig(); err != nil {
		return fmt.Errorf("failed to load Redis config: %w", err)
	}
	
	// Initialize storage configuration
	if cm.storageConfig, err = cm.loadStorageConfig(); err != nil {
		return fmt.Errorf("failed to load storage config: %w", err)
	}
	
	// Initialize Vault configuration
	if cm.vaultConfig, err = cm.loadVaultConfig(); err != nil {
		return fmt.Errorf("failed to load Vault config: %w", err)
	}
	
	// Initialize Dapr configuration
	if cm.daprConfig, err = cm.loadDaprConfig(); err != nil {
		return fmt.Errorf("failed to load Dapr config: %w", err)
	}
	
	// Initialize service configuration
	if cm.serviceConfig, err = cm.loadServiceConfig(); err != nil {
		return fmt.Errorf("failed to load service config: %w", err)
	}
	
	return nil
}

// loadDatabaseConfig loads database configuration from environment variables
func (cm *ConfigManager) loadDatabaseConfig() (*RuntimeDatabaseConfig, error) {
	config := &RuntimeDatabaseConfig{
		Host:          cm.getEnvWithDefault("DATABASE_HOST", "localhost"),
		Database:      cm.getEnvWithDefault("DATABASE_NAME", "international_center_dev"),
		User:          cm.getEnvWithDefault("DATABASE_USER", "dev_user"),
		Password:      cm.getEnvWithDefault("DATABASE_PASSWORD", "dev_password"),
		URL:           os.Getenv("DATABASE_URL"),
		ContainerHost: cm.getEnvWithDefault("POSTGRES_CONTAINER_HOST", "localhost"),
		ContainerURL:  os.Getenv("DATABASE_CONTAINER_URL"),
	}
	
	// Parse port with default
	portStr := cm.getEnvWithDefault("DATABASE_PORT", "5432")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid DATABASE_PORT '%s': %w", portStr, err)
	}
	config.Port = port
	
	// Generate URL if not provided
	if config.URL == "" {
		config.URL = fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
			config.User, config.Password, config.Host, config.Port, config.Database)
	}
	
	// Generate container URL if not provided
	if config.ContainerURL == "" {
		config.ContainerURL = fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
			config.User, config.Password, config.ContainerHost, config.Port, config.Database)
	}
	
	return config, nil
}

// loadRedisConfig loads Redis configuration from environment variables
func (cm *ConfigManager) loadRedisConfig() (*RuntimeRedisConfig, error) {
	config := &RuntimeRedisConfig{
		Host:          cm.getEnvWithDefault("REDIS_HOST", "localhost"),
		Password:      cm.getEnvWithDefault("REDIS_PASSWORD", "dev-redis-password"),
		Address:       os.Getenv("REDIS_ADDR"),
		ContainerHost: cm.getEnvWithDefault("REDIS_CONTAINER_HOST", "localhost"),
		ContainerAddr: os.Getenv("REDIS_CONTAINER_ADDR"),
	}
	
	// Parse port with default
	portStr := cm.getEnvWithDefault("REDIS_PORT", "6379")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_PORT '%s': %w", portStr, err)
	}
	config.Port = port
	
	// Generate address if not provided
	if config.Address == "" {
		config.Address = fmt.Sprintf("%s:%d", config.Host, config.Port)
	}
	
	// Generate container address if not provided
	if config.ContainerAddr == "" {
		config.ContainerAddr = fmt.Sprintf("%s:%d", config.ContainerHost, config.Port)
	}
	
	return config, nil
}

// loadStorageConfig loads storage configuration from environment variables
func (cm *ConfigManager) loadStorageConfig() (*RuntimeStorageConfig, error) {
	azuritePort := cm.getEnvWithDefault("AZURITE_PORT", "10000")
	azuriteHost := cm.getEnvWithDefault("AZURITE_CONTAINER_HOST", "localhost")
	
	config := &RuntimeStorageConfig{
		AzuriteHost: azuriteHost,
		AzuritePort: azuritePort,
		BlobEndpoint: fmt.Sprintf("http://%s:%s/devstoreaccount1", azuriteHost, azuritePort),
		ConnectionString: fmt.Sprintf("DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://%s:%s/devstoreaccount1;", azuriteHost, azuritePort),
	}
	
	return config, nil
}

// loadVaultConfig loads Vault configuration from environment variables
func (cm *ConfigManager) loadVaultConfig() (*RuntimeVaultConfig, error) {
	config := &RuntimeVaultConfig{
		Address:   cm.getEnvWithDefault("VAULT_ADDR", "http://localhost:8200"),
		Token:     os.Getenv("VAULT_TOKEN"),
		Namespace: os.Getenv("VAULT_NAMESPACE"),
	}
	
	return config, nil
}

// loadDaprConfig loads Dapr configuration from environment variables
func (cm *ConfigManager) loadDaprConfig() (*RuntimeDaprConfig, error) {
	config := &RuntimeDaprConfig{
		HTTPEndpoint:      cm.getEnvWithDefault("DAPR_HTTP_ENDPOINT", "localhost:3500"),
		GRPCEndpoint:      cm.getEnvWithDefault("DAPR_GRPC_ENDPOINT", "localhost:50001"),
		PlacementEndpoint: cm.getEnvWithDefault("DAPR_PLACEMENT_ENDPOINT", "localhost:50005"),
		ComponentsPath:    cm.getEnvWithDefault("DAPR_COMPONENTS_PATH", "/tmp/dapr-components"),
	}
	
	return config, nil
}

// loadServiceConfig loads service configuration from environment variables
func (cm *ConfigManager) loadServiceConfig() (*RuntimeServiceConfig, error) {
	config := &RuntimeServiceConfig{
		Host:             cm.getEnvWithDefault("SERVICE_HOST", "localhost"),
		ContentAPIURL:    cm.getEnvWithDefault("CONTENT_API_URL", "http://localhost:8080"),
		ServicesAPIURL:   cm.getEnvWithDefault("SERVICES_API_URL", "http://localhost:8081"),
		PublicGatewayURL: cm.getEnvWithDefault("PUBLIC_GATEWAY_URL", "http://localhost:8082"),
		AdminGatewayURL:  cm.getEnvWithDefault("ADMIN_GATEWAY_URL", "http://localhost:8083"),
	}
	
	return config, nil
}

// getEnvWithDefault gets an environment variable with a default value
func (cm *ConfigManager) getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Getter methods for different configuration aspects

// GetEnvironment returns the detected environment
func (cm *ConfigManager) GetEnvironment() Environment {
	return cm.environment
}

// GetEnvironmentConfig returns the environment-specific configuration
func (cm *ConfigManager) GetEnvironmentConfig() *EnvironmentConfig {
	return cm.envConfig
}

// GetPulumiConfig returns the Pulumi configuration wrapper
func (cm *ConfigManager) GetPulumiConfig() *PulumiConfig {
	return cm.pulumiConfig
}


// GetDatabaseConfig returns the runtime database configuration
func (cm *ConfigManager) GetDatabaseConfig() *RuntimeDatabaseConfig {
	return cm.databaseConfig
}

// GetRedisConfig returns the runtime Redis configuration
func (cm *ConfigManager) GetRedisConfig() *RuntimeRedisConfig {
	return cm.redisConfig
}

// GetStorageConfig returns the runtime storage configuration
func (cm *ConfigManager) GetStorageConfig() *RuntimeStorageConfig {
	return cm.storageConfig
}

// GetVaultConfig returns the runtime Vault configuration
func (cm *ConfigManager) GetVaultConfig() *RuntimeVaultConfig {
	return cm.vaultConfig
}

// GetDaprConfig returns the runtime Dapr configuration
func (cm *ConfigManager) GetDaprConfig() *RuntimeDaprConfig {
	return cm.daprConfig
}

// GetServiceConfig returns the runtime service configuration
func (cm *ConfigManager) GetServiceConfig() *RuntimeServiceConfig {
	return cm.serviceConfig
}

// Validation methods

// ValidateConfiguration validates all configuration settings
func (cm *ConfigManager) ValidateConfiguration() error {
	if err := ValidateEnvironment(cm.environment); err != nil {
		return fmt.Errorf("invalid environment configuration: %w", err)
	}
	
	if err := cm.validateDatabaseConfig(); err != nil {
		return fmt.Errorf("invalid database configuration: %w", err)
	}
	
	if err := cm.validateRedisConfig(); err != nil {
		return fmt.Errorf("invalid Redis configuration: %w", err)
	}
	
	return nil
}

// validateDatabaseConfig validates database configuration
func (cm *ConfigManager) validateDatabaseConfig() error {
	if cm.databaseConfig.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if cm.databaseConfig.Port <= 0 {
		return fmt.Errorf("database port must be positive")
	}
	if cm.databaseConfig.Database == "" {
		return fmt.Errorf("database name is required")
	}
	if cm.databaseConfig.User == "" {
		return fmt.Errorf("database user is required")
	}
	return nil
}

// validateRedisConfig validates Redis configuration
func (cm *ConfigManager) validateRedisConfig() error {
	if cm.redisConfig.Host == "" {
		return fmt.Errorf("Redis host is required")
	}
	if cm.redisConfig.Port <= 0 {
		return fmt.Errorf("Redis port must be positive")
	}
	return nil
}

// Utility methods

// GetNetworkName returns the environment-specific network name
func (cm *ConfigManager) GetNetworkName() string {
	return fmt.Sprintf("%s-network", cm.environment)
}

// GetResourcePrefix returns the resource prefix for the environment
func (cm *ConfigManager) GetResourcePrefix() string {
	return cm.envConfig.ResourcePrefix
}

// GetResourceSuffix returns the resource suffix for the environment
func (cm *ConfigManager) GetResourceSuffix() string {
	return cm.envConfig.ResourceSuffix
}

// GetProjectName returns the project name
func (cm *ConfigManager) GetProjectName() string {
	return cm.envConfig.ProjectName
}

// GetStackName returns the stack name
func (cm *ConfigManager) GetStackName() string {
	return cm.envConfig.StackName
}

// IsDevelopment returns true if the environment is development
func (cm *ConfigManager) IsDevelopment() bool {
	return cm.environment.IsDevelopment()
}

// IsStaging returns true if the environment is staging
func (cm *ConfigManager) IsStaging() bool {
	return cm.environment.IsStaging()
}

// IsProduction returns true if the environment is production
func (cm *ConfigManager) IsProduction() bool {
	return cm.environment.IsProduction()
}

// NewConfigManagerFromEnv creates a ConfigManager from environment variables (for testing)
func NewConfigManagerFromEnv() (*ConfigManager, error) {
	// Detect environment from environment variables
	environment, err := DetectEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to detect environment: %w", err)
	}
	
	// Load environment file first
	if err := loadEnvironmentFile(); err != nil {
		// Log warning but continue
		fmt.Printf("Warning: Failed to load environment file: %v\n", err)
	}
	
	// Get environment configuration
	envConfig, err := GetEnvironmentConfig(environment)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment configuration: %w", err)
	}
	
	// Create ConfigManager with runtime configurations only
	cm := &ConfigManager{
		environment: environment,
		envConfig:   envConfig,
	}
	
	// Load runtime configurations
	databaseConfig, err := cm.loadDatabaseConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load database configuration: %w", err)
	}
	cm.databaseConfig = databaseConfig
	
	redisConfig, err := cm.loadRedisConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load Redis configuration: %w", err)
	}
	cm.redisConfig = redisConfig
	
	storageConfig, err := cm.loadStorageConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load storage configuration: %w", err)
	}
	cm.storageConfig = storageConfig
	
	vaultConfig, err := cm.loadVaultConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load Vault configuration: %w", err)
	}
	cm.vaultConfig = vaultConfig
	
	daprConfig, err := cm.loadDaprConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load Dapr configuration: %w", err)
	}
	cm.daprConfig = daprConfig
	
	serviceConfig, err := cm.loadServiceConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load service configuration: %w", err)
	}
	cm.serviceConfig = serviceConfig
	
	return cm, nil
}

// GetEnvironmentVariable returns an environment variable value and whether it exists
func (cm *ConfigManager) GetEnvironmentVariable(key string) (string, bool) {
	value := os.Getenv(key)
	return value, value != ""
}