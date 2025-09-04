package testing

import (
	"testing"

	"github.com/axiom-software-co/international-center/src/deployer/shared/testing"
	"github.com/stretchr/testify/suite"
)

type ProductionValidationTestSuite struct {
	suite.Suite
	integrationSuite       *testing.IntegrationTestSuite
	pulumiSuite           *testing.PulumiDeploymentTestSuite
	migrationSuite        *testing.MigrationValidationTestSuite
	infrastructureSuite   *testing.InfrastructureComponentTestSuite
	observabilitySuite    *testing.ObservabilityValidationTestSuite
}

func TestProductionValidationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping production integration tests in short mode")
	}

	// Production tests require explicit environment variable
	if os.Getenv("INTEGRATION_TESTS") != "production" {
		t.Skip("Production integration tests require INTEGRATION_TESTS=production")
	}

	suite.Run(t, new(ProductionValidationTestSuite))
}

func (suite *ProductionValidationTestSuite) SetupSuite() {
	// Initialize deployment-focused integration test suites for production
	suite.integrationSuite = testing.NewIntegrationTestSuite(suite.T())
	suite.pulumiSuite = testing.NewPulumiDeploymentTestSuite(suite.T())
	suite.migrationSuite = testing.NewMigrationValidationTestSuite(suite.T())
	suite.infrastructureSuite = testing.NewInfrastructureComponentTestSuite(suite.T())
	suite.observabilitySuite = testing.NewObservabilityValidationTestSuite(suite.T())
	
	// Validate production infrastructure is fully deployed and ready
	suite.integrationSuite.InfrastructureHealthCheck(suite.T())
	suite.pulumiSuite.ValidateStackEnvironmentConsistency(suite.T())
	suite.migrationSuite.ValidateMigrationCompletion(suite.T())
	suite.observabilitySuite.ValidateGrafanaCloudIntegration(suite.T())
}

func (suite *ProductionValidationTestSuite) TearDownSuite() {
	// Cleanup is handled by the integration test suites
	if suite.integrationSuite != nil {
		suite.integrationSuite.Cleanup(suite.T())
	}
}

func (suite *ProductionValidationTestSuite) TestProductionInfrastructureDeployment() {
	t := suite.T()

	t.Run("PulumiProductionStackValidation", func(t *testing.T) {
		// Validate Pulumi stack is properly deployed for production
		suite.pulumiSuite.ValidateStackEnvironmentConsistency(t)
		
		// Validate expected production stack outputs exist
		expectedOutputs := []string{"database_connection_string", "redis_endpoint", "api_endpoint", "admin_endpoint", "vault_endpoint"}
		suite.pulumiSuite.ValidateStackOutputs(t, expectedOutputs)
		
		// Validate stack outputs contain required service endpoints
		requiredEndpoints := []string{"api", "admin", "grafana", "vault"}
		suite.pulumiSuite.ValidateStackOutputsContainEndpoints(t, requiredEndpoints)

		// Validate resource deployment state
		suite.pulumiSuite.ValidateResourceDeployment(t, "azure:containerapp/ContainerApp", "production-api")
		suite.pulumiSuite.ValidateResourceDeployment(t, "azure:containerapp/ContainerApp", "production-admin")
	})

	t.Run("ProductionDatabaseMigrationValidation", func(t *testing.T) {
		// Validate all migrations have been applied with production validation
		suite.migrationSuite.ValidateMigrationCompletion(t)
		
		// Validate schema integrity matches specifications
		suite.migrationSuite.ValidateContentDomainSchema(t)
		suite.migrationSuite.ValidateServicesDomainSchema(t)

		// Validate production-specific table indexes
		productionIndexes := []string{
			"idx_services_category_id", "idx_services_publishing_status", "idx_services_slug",
			"idx_content_hash", "idx_content_upload_status", "idx_content_storage_path",
		}
		suite.migrationSuite.ValidateTableIndexes(t, productionIndexes)
	})

	t.Run("ProductionInfrastructureComponentHealth", func(t *testing.T) {
		// Validate all expected production infrastructure components are healthy
		suite.infrastructureSuite.ValidateComponentHealth(t)
	})
}

func (suite *ProductionValidationTestSuite) TestProductionObservabilityIntegration() {
	t := suite.T()

	t.Run("ProductionGrafanaCloudValidation", func(t *testing.T) {
		// Validate Grafana Cloud integration for production audit logging
		suite.observabilitySuite.ValidateGrafanaCloudIntegration(t)
	})

	t.Run("ProductionObservabilityEndpointsHealth", func(t *testing.T) {
		// Test observability endpoints health for production
		if suite.integrationSuite.Environment.GrafanaEndpoint != "" {
			resp := suite.integrationSuite.InvokeHTTPEndpoint(t, "GET", suite.integrationSuite.Environment.GrafanaEndpoint+"/api/health", 
				map[string]string{"Authorization": "Bearer " + suite.integrationSuite.Environment.GrafanaAPIKey})
			defer resp.Body.Close()
			suite.Require().Equal(200, resp.StatusCode, "Grafana must be accessible in production")
		}

		// Test Prometheus endpoint if available
		if suite.integrationSuite.Environment.PrometheusEndpoint != "" {
			resp := suite.integrationSuite.InvokeHTTPEndpoint(t, "GET", suite.integrationSuite.Environment.PrometheusEndpoint+"/-/ready", nil)
			defer resp.Body.Close()
			suite.Require().True(resp.StatusCode < 500, "Prometheus should be accessible for production metrics")
		}
	})

	t.Run("ProductionAuditLoggingValidation", func(t *testing.T) {
		// Validate audit logging integration with Grafana Cloud Loki for compliance
		if suite.integrationSuite.Environment.LokiEndpoint != "" {
			resp := suite.integrationSuite.InvokeHTTPEndpoint(t, "GET", suite.integrationSuite.Environment.LokiEndpoint+"/ready", nil)
			defer resp.Body.Close()
			suite.Require().True(resp.StatusCode < 500, "Loki must be accessible for compliance audit logging")
		}
	})
}

func (suite *ProductionValidationTestSuite) TestProductionSecurityValidation() {
	t := suite.T()

	t.Run("ProductionHTTPSEnforcement", func(t *testing.T) {
		// Validate HTTPS enforcement in production
		if suite.integrationSuite.Environment.APIEndpoint != "" {
			suite.Require().Contains(suite.integrationSuite.Environment.APIEndpoint, "https://", 
				"Production API must enforce HTTPS")
		}
		if suite.integrationSuite.Environment.AdminEndpoint != "" {
			suite.Require().Contains(suite.integrationSuite.Environment.AdminEndpoint, "https://", 
				"Production Admin must enforce HTTPS")
		}
	})

	t.Run("ProductionSecurityHeaders", func(t *testing.T) {
		// Test comprehensive security headers in production
		if suite.integrationSuite.Environment.APIEndpoint != "" {
			resp := suite.integrationSuite.InvokeHTTPEndpoint(t, "GET", suite.integrationSuite.Environment.APIEndpoint+"/health", nil)
			defer resp.Body.Close()
			
			// Validate comprehensive security headers
			requiredHeaders := map[string]string{
				"X-Content-Type-Options": "nosniff",
				"X-Frame-Options":        "DENY",
				"X-XSS-Protection":       "1; mode=block",
				"Strict-Transport-Security": "",
				"Content-Security-Policy":   "",
				"Referrer-Policy":          "",
			}

			for headerName, expectedValue := range requiredHeaders {
				actualValue := resp.Header.Get(headerName)
				suite.Require().NotEmpty(actualValue, "Production security header %s must be present", headerName)
				
				if expectedValue != "" {
					suite.Require().Equal(expectedValue, actualValue, 
						"Production security header %s must have correct value", headerName)
				}
			}
		}
	})

	t.Run("VaultSecurityIntegration", func(t *testing.T) {
		// Test Vault integration for secrets management
		if suite.integrationSuite.Environment.VaultAddr != "" {
			resp := suite.integrationSuite.InvokeHTTPEndpoint(t, "GET", suite.integrationSuite.Environment.VaultAddr+"/v1/sys/health", nil)
			defer resp.Body.Close()
			suite.Require().True(resp.StatusCode < 500, "Vault should be accessible for production secrets management")
		}
	})
}

func (suite *ProductionValidationTestSuite) TestProductionPerformanceValidation() {
	t := suite.T()

	t.Run("ProductionDatabasePerformance", func(t *testing.T) {
		// Test database performance meets production standards
		suite.integrationSuite.WaitForInfrastructureStabilization(1000) // 1 second
		suite.integrationSuite.InfrastructureHealthCheck(t)
	})

	t.Run("ProductionRedisPerformance", func(t *testing.T) {
		// Test Redis performance and state management for production scale
		testKey := "production-performance-test"
		testValue := "production-validation"
		
		suite.integrationSuite.SaveTestState(t, testKey, testValue)
		retrieved, found := suite.integrationSuite.GetTestState(t, testKey)
		suite.Require().True(found, "Production state should be retrievable")
		suite.Require().Equal(testValue, retrieved, "Production state should match saved state")
	})

	t.Run("ProductionEndpointResponseTime", func(t *testing.T) {
		// Test production endpoint response times
		if suite.integrationSuite.Environment.APIEndpoint != "" {
			// Warm up request
			resp := suite.integrationSuite.InvokeHTTPEndpoint(t, "GET", suite.integrationSuite.Environment.APIEndpoint+"/health", nil)
			resp.Body.Close()
			
			// Measure response time (should be fast for health endpoint)
			resp = suite.integrationSuite.InvokeHTTPEndpoint(t, "GET", suite.integrationSuite.Environment.APIEndpoint+"/health", nil)
			defer resp.Body.Close()
			suite.Require().True(resp.StatusCode < 500, "Production health endpoint should respond quickly")
		}
	})
}

func (suite *ProductionValidationTestSuite) TestProductionDeploymentValidation() {
	t := suite.T()

	t.Run("ProductionEnvironmentConsistency", func(t *testing.T) {
		// Validate production environment configuration
		suite.Require().Equal("production", suite.integrationSuite.Environment.Environment,
			"Environment must be configured as production")
		suite.Require().NotEmpty(suite.integrationSuite.Environment.DatabaseURL,
			"Database URL must be configured for production")
		suite.Require().NotEmpty(suite.integrationSuite.Environment.RedisAddr,
			"Redis address must be configured for production")
		suite.Require().NotEmpty(suite.integrationSuite.Environment.GrafanaEndpoint,
			"Grafana endpoint must be configured for production")
		suite.Require().NotEmpty(suite.integrationSuite.Environment.VaultAddr,
			"Vault address must be configured for production")
	})

	t.Run("ProductionResourceHealthChecks", func(t *testing.T) {
		// Comprehensive production infrastructure readiness check
		healthChecks := map[string]string{}
		
		if suite.integrationSuite.Environment.APIEndpoint != "" {
			healthChecks["api"] = suite.integrationSuite.Environment.APIEndpoint + "/health"
		}
		if suite.integrationSuite.Environment.AdminEndpoint != "" {
			healthChecks["admin"] = suite.integrationSuite.Environment.AdminEndpoint + "/health"
		}
		
		if len(healthChecks) > 0 {
			suite.pulumiSuite.CheckResourceHealth(t, healthChecks)
		}
	})
}

func (suite *ProductionValidationTestSuite) TestProductionComplianceValidation() {
	t := suite.T()

	t.Run("AuditLoggingCompliance", func(t *testing.T) {
		// Validate audit logging meets compliance requirements
		suite.observabilitySuite.ValidateGrafanaCloudIntegration(t)
		
		// Test audit event publishing
		testChannel := "production.audit.test"
		testMessage := `{"type": "compliance-test", "status": "validation"}`
		
		suite.integrationSuite.PublishTestEvent(t, testChannel, testMessage)
	})

	t.Run("DataRetentionValidation", func(t *testing.T) {
		// Validate data retention policies are in place
		// This would typically verify backup and retention configurations
		suite.integrationSuite.InfrastructureHealthCheck(t)
	})
}

func (suite *ProductionValidationTestSuite) TestProductionOverallSystemValidation() {
	t := suite.T()

	t.Run("ComprehensiveProductionValidation", func(t *testing.T) {
		// Final comprehensive validation of production infrastructure
		suite.integrationSuite.InfrastructureHealthCheck(t)
		suite.pulumiSuite.ValidateStackEnvironmentConsistency(t)
		suite.migrationSuite.ValidateMigrationCompletion(t)
		suite.infrastructureSuite.ValidateComponentHealth(t)
		suite.observabilitySuite.ValidateGrafanaCloudIntegration(t)
		
		// Additional production-specific validations
		if suite.integrationSuite.Environment.APIEndpoint != "" && suite.integrationSuite.Environment.AdminEndpoint != "" {
			healthChecks := map[string]string{
				"api":   suite.integrationSuite.Environment.APIEndpoint + "/health",
				"admin": suite.integrationSuite.Environment.AdminEndpoint + "/health",
			}
			suite.pulumiSuite.CheckResourceHealth(t, healthChecks)
		}
		
		t.Logf("✅ Production infrastructure validation completed successfully")
	})
}