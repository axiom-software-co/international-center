package testing

import (
	"context"
	"testing"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/stretchr/testify/require"
	
	sharedconfig "github.com/axiom-software-co/international-center/src/deployer/shared/config"
	"github.com/axiom-software-co/international-center/src/deployer/shared/infrastructure"
)

const (
	IntegrationTestTimeout = 2 * time.Minute
	HealthCheckTimeout     = 30 * time.Second
)

// InfrastructureTestSuite provides contract-first testing using real infrastructure deployments
type InfrastructureTestSuite struct {
	t               *testing.T
	environment     string
	factoryManager  *infrastructure.InfrastructureFactoryManager
	configManager   *sharedconfig.ConfigManager
	ctx             context.Context
	cancel          context.CancelFunc
	pulumiContext   *pulumi.Context
	pulumiConfig    *config.Config
}

// NewInfrastructureTestSuite creates a new test suite that uses real infrastructure through factory patterns
func NewInfrastructureTestSuite(t *testing.T, environment string) *InfrastructureTestSuite {
	if testing.Short() {
		t.Skip("Skipping infrastructure integration tests in short mode")
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), IntegrationTestTimeout)
	
	// Create factory manager for the environment
	factoryManager, err := infrastructure.NewInfrastructureFactoryManager(environment)
	require.NoError(t, err, "Failed to create infrastructure factory manager")
	
	return &InfrastructureTestSuite{
		t:              t,
		environment:    environment,
		factoryManager: factoryManager,
		ctx:            ctx,
		cancel:         cancel,
	}
}

// Setup initializes the test suite with configuration and Pulumi context
func (suite *InfrastructureTestSuite) Setup() {
	// Initialize ConfigManager
	configManager, err := sharedconfig.NewConfigManagerFromEnv()
	require.NoError(suite.t, err, "Failed to initialize ConfigManager from environment")
	suite.configManager = configManager
	
	// Get Pulumi configuration
	suite.pulumiConfig = configManager.GetPulumiConfig().GetUnderlyingConfig()
	require.NotNil(suite.t, suite.pulumiConfig, "Pulumi configuration must be available")
	
	suite.t.Logf("Infrastructure test suite initialized for environment: %s", suite.environment)
}

// Teardown cleans up the test suite
func (suite *InfrastructureTestSuite) Teardown() {
	if suite.cancel != nil {
		suite.cancel()
	}
}

// RequireEnvironmentRunning verifies that the complete development environment is running
func (suite *InfrastructureTestSuite) RequireEnvironmentRunning() {
	suite.t.Helper()
	
	// Verify that all required environment variables are set
	suite.requireEnvironmentVariable("DATABASE_URL")
	suite.requireEnvironmentVariable("REDIS_URL") 
	suite.requireEnvironmentVariable("VAULT_URL")
	suite.requireEnvironmentVariable("AZURITE_URL")
	suite.requireEnvironmentVariable("GRAFANA_URL")
	suite.requireEnvironmentVariable("DAPR_HTTP_ENDPOINT")
	
	suite.t.Logf("All required environment variables are set for %s environment", suite.environment)
}

// requireEnvironmentVariable validates that a required environment variable is set
func (suite *InfrastructureTestSuite) requireEnvironmentVariable(key string) {
	value, exists := suite.configManager.GetEnvironmentVariable(key)
	require.True(suite.t, exists, "Required environment variable %s must be set when development environment is running", key)
	require.NotEmpty(suite.t, value, "Required environment variable %s must not be empty", key)
}

// GetDatabaseStack returns the database stack using the factory pattern
func (suite *InfrastructureTestSuite) GetDatabaseStack() infrastructure.DatabaseStack {
	return suite.factoryManager.CreateDatabaseStack(suite.pulumiContext, suite.pulumiConfig, suite.environment)
}

// GetStorageStack returns the storage stack using the factory pattern
func (suite *InfrastructureTestSuite) GetStorageStack() infrastructure.StorageStack {
	return suite.factoryManager.CreateStorageStack(suite.pulumiContext, suite.pulumiConfig, suite.environment)
}

// GetDaprStack returns the dapr stack using the factory pattern
func (suite *InfrastructureTestSuite) GetDaprStack() infrastructure.DaprStack {
	return suite.factoryManager.CreateDaprStack(suite.pulumiContext, suite.pulumiConfig, suite.environment)
}

// GetVaultStack returns the vault stack using the factory pattern
func (suite *InfrastructureTestSuite) GetVaultStack() infrastructure.VaultStack {
	return suite.factoryManager.CreateVaultStack(suite.pulumiContext, suite.pulumiConfig, suite.environment)
}

// GetObservabilityStack returns the observability stack using the factory pattern
func (suite *InfrastructureTestSuite) GetObservabilityStack() infrastructure.ObservabilityStack {
	return suite.factoryManager.CreateObservabilityStack(suite.pulumiContext, suite.pulumiConfig, suite.environment)
}

// GetServiceStack returns the service stack using the factory pattern
func (suite *InfrastructureTestSuite) GetServiceStack() infrastructure.ServiceStack {
	return suite.factoryManager.CreateServiceStack(suite.pulumiContext, suite.pulumiConfig, suite.environment)
}

// RequireCapability checks if the environment supports a specific capability and skips if not
func (suite *InfrastructureTestSuite) RequireCapability(capability string) {
	suite.t.Helper()
	
	err := suite.factoryManager.ValidateEnvironmentCapabilities([]string{capability})
	if err != nil {
		suite.t.Skipf("Environment %s does not support required capability: %s", suite.environment, capability)
	}
}

// GetEnvironmentCapabilities returns the supported capabilities for the test environment
func (suite *InfrastructureTestSuite) GetEnvironmentCapabilities() []string {
	return suite.factoryManager.GetSupportedCapabilities()
}

// GetDeploymentStrategy returns the deployment strategy for the environment
func (suite *InfrastructureTestSuite) GetDeploymentStrategy() infrastructure.DeploymentStrategy {
	return suite.factoryManager.GetDeploymentStrategy()
}

// Context returns the test context with timeout
func (suite *InfrastructureTestSuite) Context() context.Context {
	return suite.ctx
}

// Environment returns the test environment name
func (suite *InfrastructureTestSuite) Environment() string {
	return suite.environment
}

// ConfigManager returns the configuration manager
func (suite *InfrastructureTestSuite) ConfigManager() *sharedconfig.ConfigManager {
	return suite.configManager
}

// TestingT returns the testing.T instance
func (suite *InfrastructureTestSuite) TestingT() *testing.T {
	return suite.t
}