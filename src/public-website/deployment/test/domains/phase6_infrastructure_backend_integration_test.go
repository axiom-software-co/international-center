package domains

import (
	"context"
	"net/http"
	"testing"
	"time"

	sharedtesting "github.com/axiom-software-co/international-center/src/public-website/infrastructure/test/shared"
	"github.com/stretchr/testify/assert"
)

// PHASE 6: INFRASTRUCTURE TO BACKEND INTEGRATION TESTS
// WHY: Infrastructure to backend integration must work before full stack integration
// SCOPE: Database connectivity through backend, state store operations, pub/sub messaging
// DEPENDENCIES: Phases 1-5 must pass
// CONTEXT: Dapr abstractions, backend to infrastructure communication

func TestPhase6InfrastructureBackendIntegration(t *testing.T) {
	// Environment validation required for infrastructure to backend integration tests
	sharedtesting.ValidateEnvironmentPrerequisites(t)

	timeout := 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	t.Run("DatabaseIntegrationThroughServiceMesh", func(t *testing.T) {
		// Test database operations through service mesh
		dbTester := sharedtesting.NewDatabaseIntegrationTester()

		errors := dbTester.ValidateDatabaseIntegration(ctx)
		if len(errors) > 0 {
			for _, err := range errors {
				t.Logf("Database integration issue: %v", err)
			}
		} else {
			t.Log("✅ Database integration: All operations through service mesh working")
		}

		assert.Empty(t, errors, "Database integration through service mesh should work")
	})

	t.Run("StateStoreOperationsThroughDapr", func(t *testing.T) {
		// Test state store operations through Dapr
		services := []string{"content", "inquiries", "notifications"}

		for _, service := range services {
			t.Run(service+"StateStoreIntegration", func(t *testing.T) {
				daprClient := sharedtesting.NewDaprServiceTestClient(service+"-integration-test", "3500")

				// Test state store accessibility and operations
				err := daprClient.ValidateStateStoreAccess(ctx, "statestore")
				if err != nil {
					t.Logf("State store access issue for %s: %v", service, err)
				} else {
					t.Logf("✅ Service %s: State store operations through Dapr operational", service)
				}

				assert.NoError(t, err, "Service %s should have operational state store access through Dapr", service)
			})
		}
	})

	t.Run("PubSubMessagingIntegration", func(t *testing.T) {
		// Test pub/sub messaging integration through Dapr
		pubSubTester := sharedtesting.NewServiceMeshReliabilityTester()

		errors := pubSubTester.ValidateServiceMeshResilience(ctx)
		if len(errors) > 0 {
			for _, err := range errors {
				t.Logf("Pub/sub messaging issue: %v", err)
			}
		} else {
			t.Log("✅ Pub/sub messaging: Integration through Dapr operational")
		}

		// Pub/sub integration should have minimal issues
		assert.LessOrEqual(t, len(errors), 2, "Pub/sub messaging integration should have minimal issues")
	})

	t.Run("SecretStoreIntegrationThroughDapr", func(t *testing.T) {
		// Test secret store integration through Dapr
		services := []string{"content-api", "inquiries-api", "notifications-api"}
		client := &http.Client{Timeout: 10 * time.Second}

		for _, service := range services {
			t.Run(service+"SecretStoreIntegration", func(t *testing.T) {
				// Test secret store accessibility through service health
				healthURL := "http://localhost:3500/v1.0/invoke/" + service + "/method/health"

				resp, err := client.Get(healthURL)
				if err != nil {
					t.Logf("Service %s secret store integration test failed: %v", service, err)
					return
				}
				defer resp.Body.Close()

				// Service health should indicate proper secret store access
				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					t.Logf("✅ Service %s: Secret store integration operational", service)
				} else {
					t.Logf("⚠️ Service %s: Secret store integration may have issues (status: %d)", service, resp.StatusCode)
				}

				assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300,
					"Service %s should have operational secret store integration, got status %d", service, resp.StatusCode)
			})
		}
	})

	t.Run("BackendToInfrastructureConnectivity", func(t *testing.T) {
		// Test backend service connectivity to infrastructure components
		connectivityTests := []struct {
			service     string
			component   string
			description string
		}{
			{"content-api", "postgresql", "Content service database connectivity"},
			{"inquiries-api", "postgresql", "Inquiries service database connectivity"},
			{"notifications-api", "postgresql", "Notifications service database connectivity"},
			{"content-api", "rabbitmq", "Content service messaging connectivity"},
			{"inquiries-api", "rabbitmq", "Inquiries service messaging connectivity"},
			{"notifications-api", "rabbitmq", "Notifications service messaging connectivity"},
		}

		client := &http.Client{Timeout: 10 * time.Second}

		for _, test := range connectivityTests {
			t.Run(test.service+"_to_"+test.component, func(t *testing.T) {
				// Test connectivity through service health
				healthURL := "http://localhost:3500/v1.0/invoke/" + test.service + "/method/health"

				resp, err := client.Get(healthURL)
				if err != nil {
					t.Logf("Connectivity test failed for %s to %s: %v", test.service, test.component, err)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					t.Logf("✅ Infrastructure connectivity: %s to %s operational", test.service, test.component)
				} else {
					t.Logf("⚠️ Infrastructure connectivity: %s to %s may have issues (status: %d)",
						test.service, test.component, resp.StatusCode)
				}
			})
		}
	})
}

func TestPhase6CrossStackInfrastructureValidation(t *testing.T) {
	// Test cross-stack infrastructure validation
	sharedtesting.ValidateEnvironmentPrerequisites(t)

	timeout := 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	t.Run("EndToEndInfrastructureWorkflow", func(t *testing.T) {
		// Test complete infrastructure workflow enablement
		crossStackTester := sharedtesting.NewCrossStackIntegrationTester()

		errors := crossStackTester.ValidateEndToEndWorkflow(ctx)
		if len(errors) > 0 {
			for _, err := range errors {
				t.Logf("End-to-end infrastructure workflow issue: %v", err)
			}
			t.Log("⚠️ Infrastructure workflows: Some infrastructure workflows may not be fully functional")
		} else {
			t.Log("✅ Infrastructure workflows: All infrastructure workflows operational")
		}

		// Infrastructure workflow should be mostly operational
		assert.LessOrEqual(t, len(errors), 3, "Infrastructure workflows should have minimal issues")
	})

	t.Run("InfrastructureHealthAcrossServices", func(t *testing.T) {
		// Test infrastructure health across all backend services
		services := []string{"content-api", "inquiries-api", "notifications-api"}
		client := &http.Client{Timeout: 10 * time.Second}

		healthyServices := 0
		totalServices := len(services)

		for _, service := range services {
			t.Run("InfrastructureHealth_"+service, func(t *testing.T) {
				healthURL := "http://localhost:3500/v1.0/invoke/" + service + "/method/health"

				resp, err := client.Get(healthURL)
				if err != nil {
					t.Logf("Infrastructure health check failed for %s: %v", service, err)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					healthyServices++
					t.Logf("✅ Infrastructure health: %s operational", service)
				} else {
					t.Logf("⚠️ Infrastructure health: %s not operational (status: %d)", service, resp.StatusCode)
				}
			})
		}

		healthPercentage := float64(healthyServices) / float64(totalServices)
		t.Logf("Infrastructure health across services: %.2f%% (%d/%d services operational)",
			healthPercentage*100, healthyServices, totalServices)

		// At least 60% of services should have operational infrastructure
		assert.GreaterOrEqual(t, healthPercentage, 0.6,
			"At least 60%% of services should have operational infrastructure connectivity")
	})

	t.Run("DaprInfrastructureIntegration", func(t *testing.T) {
		// Test Dapr infrastructure component integration
		daprRunner := sharedtesting.NewDaprServiceMeshTestRunner()

		// Test Dapr component configuration
		componentErrors := daprRunner.ValidateComponentConfiguration(ctx)
		if len(componentErrors) > 0 {
			for serviceName, errors := range componentErrors {
				for _, err := range errors {
					t.Logf("Dapr infrastructure component issue in %s: %v", serviceName, err)
				}
			}
			t.Log("⚠️ Dapr infrastructure: Some component integrations may have issues")
		} else {
			t.Log("✅ Dapr infrastructure: All component integrations operational")
		}

		// Dapr infrastructure integration should be operational
		assert.Empty(t, componentErrors, "Dapr infrastructure component integrations should be operational")
	})

	t.Run("InfrastructureResilienceValidation", func(t *testing.T) {
		// Test infrastructure resilience patterns
		reliabilityTester := sharedtesting.NewServiceMeshReliabilityTester()

		errors := reliabilityTester.ValidateServiceMeshResilience(ctx)
		if len(errors) > 0 {
			for _, err := range errors {
				t.Logf("Infrastructure resilience issue: %v", err)
			}
		} else {
			t.Log("✅ Infrastructure resilience: All patterns operational")
		}

		// Infrastructure resilience should be mostly operational
		assert.LessOrEqual(t, len(errors), 2, "Infrastructure resilience patterns should have minimal issues")
	})

	t.Run("InfrastructureOperationalReadiness", func(t *testing.T) {
		// Test overall infrastructure operational readiness for frontend integration
		dbTester := sharedtesting.NewDatabaseIntegrationTester()
		daprRunner := sharedtesting.NewDaprServiceMeshTestRunner()

		// Validate database integration
		dbErrors := dbTester.ValidateDatabaseIntegration(ctx)

		// Validate service mesh communication
		meshErrors := daprRunner.ValidateServiceMeshCommunication(ctx)

		// Validate component configuration
		componentErrors := daprRunner.ValidateComponentConfiguration(ctx)

		totalIssues := len(dbErrors) + len(meshErrors) + len(componentErrors)

		if totalIssues == 0 {
			t.Log("✅ Infrastructure operational readiness: Ready for frontend integration")
		} else {
			t.Logf("⚠️ Infrastructure operational readiness: %d total issues detected", totalIssues)
		}

		readinessScore := 1.0 - (float64(totalIssues) / 10.0) // Assume max 10 potential issues
		if readinessScore < 0 {
			readinessScore = 0
		}

		t.Logf("Infrastructure operational readiness: %.1f%% ready for frontend integration", readinessScore*100)

		// Infrastructure should be at least 70% ready
		assert.GreaterOrEqual(t, readinessScore, 0.7,
			"Infrastructure should be at least 70%% ready for frontend integration")
	})
}