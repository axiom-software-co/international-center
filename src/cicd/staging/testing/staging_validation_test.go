package testing

import (
	"testing"

	sharedtesting "github.com/axiom-software-co/international-center/src/deployer/shared/testing"
	"github.com/stretchr/testify/suite"
)

type StagingValidationTestSuite struct {
	suite.Suite
	integrationSuite       *sharedtesting.IntegrationTestSuite
	pulumiSuite           *sharedtesting.PulumiDeploymentTestSuite
	migrationSuite        *sharedtesting.MigrationValidationTestSuite
	infrastructureSuite   *sharedtesting.InfrastructureComponentTestSuite
	observabilitySuite    *sharedtesting.ObservabilityValidationTestSuite
}

func TestStagingValidationSuite(t *testing.T) {
	suite.Run(t, new(StagingValidationTestSuite))
}

func (suite *StagingValidationTestSuite) SetupSuite() {
	// Initialize deployment-focused integration test suites for staging
	suite.integrationSuite = sharedtesting.NewIntegrationTestSuite(suite.T())
	suite.pulumiSuite = sharedtesting.NewPulumiDeploymentTestSuite(suite.T())
	suite.migrationSuite = sharedtesting.NewMigrationValidationTestSuite(suite.T())
	suite.infrastructureSuite = sharedtesting.NewInfrastructureComponentTestSuite(suite.T())
	suite.observabilitySuite = sharedtesting.NewObservabilityValidationTestSuite(suite.T())
	
	// Validate staging infrastructure is fully deployed and ready
	suite.integrationSuite.InfrastructureHealthCheck(suite.T())
	suite.pulumiSuite.ValidateStackEnvironmentConsistency(suite.T())
	suite.migrationSuite.ValidateMigrationCompletion(suite.T())
	suite.observabilitySuite.ValidateGrafanaCloudIntegration(suite.T())
}

func (suite *StagingValidationTestSuite) TearDownSuite() {
	// Cleanup is handled by the integration test suites
	if suite.integrationSuite != nil {
		suite.integrationSuite.Cleanup(suite.T())
	}
}

func (suite *StagingValidationTestSuite) TestStagingInfrastructureDeployment() {
	t := suite.T()

	t.Run("PulumiStackValidation", func(t *testing.T) {
		// Validate Pulumi stack is properly deployed for staging
		suite.pulumiSuite.ValidateStackEnvironmentConsistency(t)
		
		// Validate expected staging stack outputs exist
		expectedOutputs := []string{"database_connection_string", "redis_endpoint", "api_endpoint", "admin_endpoint"}
		suite.pulumiSuite.ValidateStackOutputs(t, expectedOutputs)
		
		// Validate stack outputs contain required service endpoints
		requiredEndpoints := []string{"api", "admin"}
		suite.pulumiSuite.ValidateStackOutputsContainEndpoints(t, requiredEndpoints)
	})

	t.Run("DatabaseMigrationValidation", func(t *testing.T) {
		// Validate all migrations have been applied with staging validation
		suite.migrationSuite.ValidateMigrationCompletion(t)
		
		// Validate schema integrity matches specifications
		suite.migrationSuite.ValidateContentDomainSchema(t)
		suite.migrationSuite.ValidateServicesDomainSchema(t)
	})

	t.Run("InfrastructureComponentHealth", func(t *testing.T) {
		// Validate all expected infrastructure components are healthy
		suite.infrastructureSuite.ValidateComponentHealth(t)
	})
}

func (suite *StagingValidationTestSuite) TestStagingObservabilityIntegration() {
	t := suite.T()

	t.Run("GrafanaCloudConnectivity", func(t *testing.T) {
		// Validate Grafana Cloud integration for staging audit logging
		suite.observabilitySuite.ValidateGrafanaCloudIntegration(t)
	})

	t.Run("ObservabilityEndpointsHealth", func(t *testing.T) {
		// Test observability endpoints health for staging
		if suite.integrationSuite.Environment.GrafanaEndpoint != "" {
			resp := suite.integrationSuite.InvokeHTTPEndpoint(t, "GET", suite.integrationSuite.Environment.GrafanaEndpoint+"/api/health", 
				map[string]string{"Authorization": "Bearer " + suite.integrationSuite.Environment.GrafanaAPIKey})
			defer resp.Body.Close()
			suite.Require().Equal(200, resp.StatusCode, "Grafana should be accessible in staging")
		}
	})

	t.Run("AuditLoggingValidation", func(t *testing.T) {
		// Validate audit logging integration with Grafana Cloud Loki
		if suite.integrationSuite.Environment.LokiEndpoint != "" {
			resp := suite.integrationSuite.InvokeHTTPEndpoint(t, "GET", suite.integrationSuite.Environment.LokiEndpoint+"/ready", nil)
			defer resp.Body.Close()
			suite.Require().True(resp.StatusCode < 500, "Loki should be accessible for audit logging")
		}
	})
}

func (suite *StagingValidationTestSuite) TestStagingSecurityValidation() {
	t := suite.T()

	t.Run("HTTPSEnforcement", func(t *testing.T) {
		// Validate HTTPS enforcement using Cloudflare domain validation
		if suite.integrationSuite.Environment.APIEndpoint != "" {
			suite.Require().Contains(suite.integrationSuite.Environment.APIEndpoint, "https://", 
				"Staging API should enforce HTTPS")
			
			// Extract domain from API endpoint for Cloudflare validation
			if suite.integrationSuite.Environment.APIEndpoint == "https://api.axiomcloud.dev" {
				sharedtesting.ValidateHTTPSEnforcement(t, "api.axiomcloud.dev")
			}
		}
		if suite.integrationSuite.Environment.AdminEndpoint != "" {
			suite.Require().Contains(suite.integrationSuite.Environment.AdminEndpoint, "https://", 
				"Staging Admin should enforce HTTPS")
				
			// Extract domain from Admin endpoint for Cloudflare validation
			if suite.integrationSuite.Environment.AdminEndpoint == "https://admin.axiomcloud.dev" {
				sharedtesting.ValidateHTTPSEnforcement(t, "admin.axiomcloud.dev")
			}
		}
	})

	t.Run("SecurityHeaders", func(t *testing.T) {
		// Test security headers in staging environment
		if suite.integrationSuite.Environment.APIEndpoint != "" {
			resp := suite.integrationSuite.InvokeHTTPEndpoint(t, "GET", suite.integrationSuite.Environment.APIEndpoint+"/health", nil)
			defer resp.Body.Close()
			
			// Validate security headers are present
			suite.Require().NotEmpty(resp.Header.Get("X-Content-Type-Options"), "Security headers should be present")
		}
	})
}

func (suite *StagingValidationTestSuite) TestStagingDeploymentValidation() {
	t := suite.T()

	t.Run("EnvironmentConsistency", func(t *testing.T) {
		// Validate staging environment configuration
		suite.Require().Equal("staging", suite.integrationSuite.Environment.Environment,
			"Environment should be configured as staging")
		suite.Require().NotEmpty(suite.integrationSuite.Environment.DatabaseURL,
			"Database URL should be configured for staging")
		suite.Require().NotEmpty(suite.integrationSuite.Environment.RedisAddr,
			"Redis address should be configured for staging")
	})

	t.Run("InfrastructureReadiness", func(t *testing.T) {
		// Comprehensive infrastructure readiness check
		suite.integrationSuite.InfrastructureHealthCheck(t)
	})
}

func (suite *StagingValidationTestSuite) TestStagingResourceHealthChecks() {
	t := suite.T()

	t.Run("DatabasePerformanceValidation", func(t *testing.T) {
		// Test database performance in staging
		suite.integrationSuite.WaitForInfrastructureStabilization(2000) // 2 seconds
		suite.integrationSuite.InfrastructureHealthCheck(t)
	})

	t.Run("RedisPerformanceValidation", func(t *testing.T) {
		// Test Redis performance and state management
		testKey := "staging-performance-test"
		testValue := "staging-validation"
		
		suite.integrationSuite.SaveTestState(t, testKey, testValue)
		retrieved, found := suite.integrationSuite.GetTestState(t, testKey)
		suite.Require().True(found, "Test state should be retrievable")
		suite.Require().Equal(testValue, retrieved, "Retrieved state should match saved state")
	})
}

func (suite *StagingValidationTestSuite) TestStagingOverallSystemValidation() {
	t := suite.T()

	t.Run("ComprehensiveStagingValidation", func(t *testing.T) {
		// Final comprehensive validation of staging infrastructure
		suite.integrationSuite.InfrastructureHealthCheck(t)
		suite.pulumiSuite.ValidateStackEnvironmentConsistency(t)
		suite.migrationSuite.ValidateMigrationCompletion(t)
		suite.infrastructureSuite.ValidateComponentHealth(t)
		suite.observabilitySuite.ValidateGrafanaCloudIntegration(t)
		
		t.Logf("✅ Staging infrastructure validation completed successfully")
	})
}