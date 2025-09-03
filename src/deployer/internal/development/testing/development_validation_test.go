package testing

import (
	"testing"

	"github.com/axiom-software-co/international-center/src/deployer/internal/shared/testing"
	"github.com/stretchr/testify/suite"
)

type DevelopmentValidationTestSuite struct {
	suite.Suite
	integrationSuite         *testing.IntegrationTestSuite
	pulumiSuite             *testing.PulumiDeploymentTestSuite
	migrationSuite          *testing.MigrationValidationTestSuite
	infrastructureSuite     *testing.InfrastructureComponentTestSuite
}

func TestDevelopmentValidationSuite(t *testing.T) {
	suite.Run(t, new(DevelopmentValidationTestSuite))
}

func (suite *DevelopmentValidationTestSuite) SetupSuite() {
	// Initialize deployment-focused integration test suites
	suite.integrationSuite = testing.NewIntegrationTestSuite(suite.T())
	suite.pulumiSuite = testing.NewPulumiDeploymentTestSuite(suite.T())
	suite.migrationSuite = testing.NewMigrationValidationTestSuite(suite.T())
	suite.infrastructureSuite = testing.NewInfrastructureComponentTestSuite(suite.T())
	
	// Validate development infrastructure is fully deployed and ready
	suite.integrationSuite.InfrastructureHealthCheck(suite.T())
	suite.pulumiSuite.ValidateStackEnvironmentConsistency(suite.T())
	suite.migrationSuite.ValidateMigrationCompletion(suite.T())
}

func (suite *DevelopmentValidationTestSuite) TearDownSuite() {
	// Cleanup is handled by the integration test suites
	if suite.integrationSuite != nil {
		suite.integrationSuite.Cleanup(suite.T())
	}
}

func (suite *DevelopmentValidationTestSuite) TestDevelopmentInfrastructureDeployment() {
	t := suite.T()

	t.Run("PulumiStackValidation", func(t *testing.T) {
		// Validate Pulumi stack is properly deployed for development
		suite.pulumiSuite.ValidateStackEnvironmentConsistency(t)
		
		// Validate expected development stack outputs exist
		expectedOutputs := []string{"database_connection_string", "redis_endpoint"}
		suite.pulumiSuite.ValidateStackOutputs(t, expectedOutputs)
	})

	t.Run("DatabaseMigrationValidation", func(t *testing.T) {
		// Validate all migrations have been applied
		suite.migrationSuite.ValidateMigrationCompletion(t)
		
		// Validate content domain schema matches specification
		suite.migrationSuite.ValidateContentDomainSchema(t)
		
		// Validate services domain schema matches specification  
		suite.migrationSuite.ValidateServicesDomainSchema(t)
	})

	t.Run("InfrastructureComponentHealth", func(t *testing.T) {
		// Validate all expected infrastructure components are healthy
		suite.infrastructureSuite.ValidateComponentHealth(t)
	})
}

func (suite *DevelopmentValidationTestSuite) TestDevelopmentInfrastructureConnectivity() {
	t := suite.T()

	t.Run("CoreInfrastructureHealthCheck", func(t *testing.T) {
		// Perform comprehensive infrastructure health validation
		suite.integrationSuite.InfrastructureHealthCheck(t)
	})
}

func (suite *DevelopmentValidationTestSuite) TestDevelopmentInfrastructureFunctionality() {
	t := suite.T()

	t.Run("RedisStateManagement", func(t *testing.T) {
		// Test Redis state management capabilities
		testKey := "development-infrastructure-test"
		testValue := "development-validation"
		
		suite.integrationSuite.SaveTestState(t, testKey, testValue)
		retrieved, found := suite.integrationSuite.GetTestState(t, testKey)
		suite.Require().True(found, "Test state should be retrievable")
		suite.Require().Equal(testValue, retrieved, "Retrieved state should match saved state")
	})
}

func (suite *DevelopmentValidationTestSuite) TestDevelopmentEventSystem() {
	t := suite.T()

	t.Run("EventPublishing", func(t *testing.T) {
		// Test deployment event publishing capability
		testChannel := "development.infrastructure.test"
		testMessage := `{"type": "validation", "status": "success"}`
		
		suite.integrationSuite.PublishTestEvent(t, testChannel, testMessage)
	})
}

func (suite *DevelopmentValidationTestSuite) TestDevelopmentEndpointAccessibility() {
	t := suite.T()

	t.Run("InfrastructureEndpoints", func(t *testing.T) {
		// Test infrastructure endpoint accessibility
		if suite.integrationSuite.Environment.APIEndpoint != "" {
			resp := suite.integrationSuite.InvokeHTTPEndpoint(t, "GET", suite.integrationSuite.Environment.APIEndpoint+"/health", nil)
			defer resp.Body.Close()
			suite.Require().True(resp.StatusCode < 500, "Infrastructure endpoints should be accessible")
		}
	})
}

func (suite *DevelopmentValidationTestSuite) TestDevelopmentDeploymentConfiguration() {
	t := suite.T()

	t.Run("EnvironmentConfiguration", func(t *testing.T) {
		// Validate environment-specific configuration
		suite.Require().Equal("development", suite.integrationSuite.Environment.Environment,
			"Environment should be configured as development")
		suite.Require().NotEmpty(suite.integrationSuite.Environment.DatabaseURL,
			"Database URL should be configured for development")
		suite.Require().NotEmpty(suite.integrationSuite.Environment.RedisAddr,
			"Redis address should be configured for development")
	})
}

func (suite *DevelopmentValidationTestSuite) TestDevelopmentSchemaValidation() {
	t := suite.T()

	t.Run("DatabaseSchemaIntegrity", func(t *testing.T) {
		// Comprehensive schema validation using migration suite
		suite.migrationSuite.ValidateContentDomainSchema(t)
		suite.migrationSuite.ValidateServicesDomainSchema(t)
	})
}

func (suite *DevelopmentValidationTestSuite) TestDevelopmentInfrastructureStabilization() {
	t := suite.T()

	t.Run("InfrastructureStabilization", func(t *testing.T) {
		// Wait for infrastructure to stabilize after validation
		suite.integrationSuite.WaitForInfrastructureStabilization(5000) // 5 seconds for development
	})
}

func (suite *DevelopmentValidationTestSuite) TestDevelopmentOverallSystemReadiness() {
	t := suite.T()

	t.Run("ComprehensiveSystemValidation", func(t *testing.T) {
		// Final comprehensive validation of development infrastructure
		suite.integrationSuite.InfrastructureHealthCheck(t)
		suite.pulumiSuite.ValidateStackEnvironmentConsistency(t)
		suite.migrationSuite.ValidateMigrationCompletion(t)
		suite.infrastructureSuite.ValidateComponentHealth(t)
		
		t.Logf("✅ Development infrastructure validation completed successfully")
	})
}

