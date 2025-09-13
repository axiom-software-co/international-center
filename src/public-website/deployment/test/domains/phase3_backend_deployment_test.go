package domains

import (
	"context"
	"net/http"
	"testing"
	"time"

	sharedtesting "github.com/axiom-software-co/international-center/src/public-website/infrastructure/test/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// PHASE 3: BACKEND DEPLOYMENT TESTS
// WHY: Backend services must be deployed and operational before frontend integration
// SCOPE: Backend API services, Dapr sidecars, service mesh communication
// DEPENDENCIES: Phases 1-2 (infrastructure, database) must pass
// CONTEXT: content-api, inquiries-api, notifications-api deployment validation

func TestPhase3ServiceMeshIntegration(t *testing.T) {
	// Environment validation required for all service mesh tests
	sharedtesting.ValidateEnvironmentPrerequisites(t)

	timeout := 15 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	t.Run("DaprServiceMeshCommunication", func(t *testing.T) {
		// Test inter-service communication through Dapr service mesh
		daprRunner := sharedtesting.NewDaprServiceMeshTestRunner()

		errors := daprRunner.ValidateServiceMeshCommunication(ctx)
		if len(errors) > 0 {
			for _, err := range errors {
				t.Logf("Service mesh communication issue: %v", err)
			}
		} else {
			t.Log("✅ Service mesh communication: All services communicating properly")
		}

		assert.Empty(t, errors, "Service mesh communication should be successful")

		// Test comprehensive Dapr functionality
		results := daprRunner.RunComprehensiveDaprTesting(ctx)

		var successCount, failureCount int
		for _, result := range results {
			if result.Success {
				successCount++
				t.Logf("✅ %s - %s: PASSED (duration: %v)", result.ServiceName, result.TestType, result.Duration)
			} else {
				failureCount++
				t.Logf("⚠️ %s - %s: FAILED (%v) (duration: %v)", result.ServiceName, result.TestType, result.Error, result.Duration)
			}
		}

		t.Logf("Service mesh validation: %d passed, %d failed", successCount, failureCount)
		require.Greater(t, successCount, 0, "At least some service mesh tests should pass")
		assert.Equal(t, 0, failureCount, "No service mesh tests should fail")
	})

	t.Run("DaprComponentConfiguration", func(t *testing.T) {
		// Test Dapr component configuration (state store, pub/sub, secrets)
		daprRunner := sharedtesting.NewDaprServiceMeshTestRunner()

		componentErrors := daprRunner.ValidateComponentConfiguration(ctx)
		if len(componentErrors) > 0 {
			for serviceName, errors := range componentErrors {
				for _, err := range errors {
					t.Logf("Component configuration issue in %s: %v", serviceName, err)
				}
			}
		} else {
			t.Log("✅ Dapr components: All components properly configured")
		}

		assert.Empty(t, componentErrors, "All Dapr components should be properly configured")
	})

	t.Run("ServiceMeshResilience", func(t *testing.T) {
		// Test service mesh resilience patterns (circuit breaker, retry, discovery)
		reliabilityTester := sharedtesting.NewServiceMeshReliabilityTester()

		errors := reliabilityTester.ValidateServiceMeshResilience(ctx)
		if len(errors) > 0 {
			for _, err := range errors {
				t.Logf("Service mesh resilience issue: %v", err)
			}
		} else {
			t.Log("✅ Service mesh resilience: All patterns operational")
		}

		assert.Empty(t, errors, "Service mesh resilience patterns should be operational")
	})

	t.Run("BackendServiceAPICommunication", func(t *testing.T) {
		// Test actual backend service API endpoints through Dapr service mesh
		client := &http.Client{Timeout: 10 * time.Second}

		t.Run("ContentServiceAPI", func(t *testing.T) {
			// Test content service actual API endpoints
			endpoints := []struct {
				name string
				url  string
			}{
				{"content-health", "http://localhost:3500/v1.0/invoke/content-api/method/health"},
				{"content-ready", "http://localhost:3500/v1.0/invoke/content-api/method/health/ready"},
			}

			for _, endpoint := range endpoints {
				resp, err := client.Get(endpoint.url)
				require.NoError(t, err, "Should be able to invoke content-api %s endpoint", endpoint.name)
				defer resp.Body.Close()

				assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300,
					"Content API %s should respond with success status, got %d", endpoint.name, resp.StatusCode)

				t.Logf("✅ Content API %s: Responding correctly (status: %d)", endpoint.name, resp.StatusCode)
			}
		})

		t.Run("InquiriesServiceAPI", func(t *testing.T) {
			// Test inquiries service actual API endpoints
			endpoints := []struct {
				name string
				url  string
			}{
				{"inquiries-health", "http://localhost:3500/v1.0/invoke/inquiries-api/method/health"},
				{"inquiries-ready", "http://localhost:3500/v1.0/invoke/inquiries-api/method/health/ready"},
			}

			for _, endpoint := range endpoints {
				resp, err := client.Get(endpoint.url)
				require.NoError(t, err, "Should be able to invoke inquiries-api %s endpoint", endpoint.name)
				defer resp.Body.Close()

				assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300,
					"Inquiries API %s should respond with success status, got %d", endpoint.name, resp.StatusCode)

				t.Logf("✅ Inquiries API %s: Responding correctly (status: %d)", endpoint.name, resp.StatusCode)
			}
		})

		t.Run("NotificationsServiceAPI", func(t *testing.T) {
			// Test notifications service actual API endpoints
			endpoints := []struct {
				name string
				url  string
			}{
				{"notifications-health", "http://localhost:3500/v1.0/invoke/notifications-api/method/health"},
				{"notifications-ready", "http://localhost:3500/v1.0/invoke/notifications-api/method/health/ready"},
			}

			for _, endpoint := range endpoints {
				resp, err := client.Get(endpoint.url)
				require.NoError(t, err, "Should be able to invoke notifications-api %s endpoint", endpoint.name)
				defer resp.Body.Close()

				assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300,
					"Notifications API %s should respond with success status, got %d", endpoint.name, resp.StatusCode)

				t.Logf("✅ Notifications API %s: Responding correctly (status: %d)", endpoint.name, resp.StatusCode)
			}
		})
	})

	t.Run("ServiceSidecarConnectivity", func(t *testing.T) {
		// Test service-to-sidecar connectivity for all services
		services := []string{"content", "inquiries", "notifications"}
		client := &http.Client{Timeout: 10 * time.Second}

		for _, service := range services {
			t.Run(service+"SidecarConnectivity", func(t *testing.T) {
				// Test sidecar health endpoint
				sidecarURL := "http://localhost:3500/v1.0/invoke/" + service + "/method/health"

				resp, err := client.Get(sidecarURL)
				require.NoError(t, err, "Should be able to connect to %s sidecar", service)
				defer resp.Body.Close()

				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					t.Logf("✅ Service %s: Sidecar connectivity operational", service)
				} else {
					t.Logf("⚠️ Service %s: Sidecar connectivity issue (status: %d)", service, resp.StatusCode)
				}

				assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300,
					"Service %s sidecar should respond with success status, got %d", service, resp.StatusCode)
			})
		}
	})

	t.Run("ServiceStateStoreOperations", func(t *testing.T) {
		// Test state store operations through Dapr
		services := []string{"content", "inquiries", "notifications"}

		for _, service := range services {
			t.Run(service+"StateStore", func(t *testing.T) {
				daprClient := sharedtesting.NewDaprServiceTestClient(service+"-test", "3500")

				// Test state store accessibility
				err := daprClient.ValidateStateStoreAccess(ctx, "statestore")
				if err != nil {
					t.Logf("State store access issue for %s: %v", service, err)
				} else {
					t.Logf("✅ Service %s: State store access operational", service)
				}

				assert.NoError(t, err, "Service %s should have state store access", service)
			})
		}
	})
}

func TestPhase3ServiceMeshOperationalReadiness(t *testing.T) {
	// Test service mesh operational readiness for deployment
	sharedtesting.ValidateEnvironmentPrerequisites(t)

	timeout := 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	t.Run("ServiceMeshDeploymentReadiness", func(t *testing.T) {
		// Comprehensive service mesh readiness validation
		daprRunner := sharedtesting.NewDaprServiceMeshTestRunner()

		// Test all aspects of service mesh functionality
		communicationErrors := daprRunner.ValidateServiceMeshCommunication(ctx)
		componentErrors := daprRunner.ValidateComponentConfiguration(ctx)

		totalIssues := len(communicationErrors) + len(componentErrors)

		if totalIssues == 0 {
			t.Log("✅ Service mesh deployment ready: All components operational")
		} else {
			t.Logf("⚠️ Service mesh deployment issues detected: %d total issues", totalIssues)

			// Log communication issues
			for _, err := range communicationErrors {
				t.Logf("Communication issue: %v", err)
			}

			// Log component issues
			for serviceName, errors := range componentErrors {
				for _, err := range errors {
					t.Logf("Component issue in %s: %v", serviceName, err)
				}
			}
		}

		assert.Equal(t, 0, totalIssues, "Service mesh should be deployment ready with no issues")
	})

	t.Run("ServiceDiscoveryOperational", func(t *testing.T) {
		// Test service discovery through Dapr
		services := []string{"content", "inquiries", "notifications"}

		for _, service := range services {
			daprClient := sharedtesting.NewDaprServiceTestClient("discovery-test", "3500")

			resp, err := daprClient.InvokeService(ctx, service, "GET", "/health", nil)
			if err != nil {
				t.Logf("Service discovery failed for %s: %v", service, err)
				require.NoError(t, err, "Should be able to discover service %s", service)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode < 500 {
				t.Logf("✅ Service discovery: %s discoverable", service)
			} else {
				t.Logf("⚠️ Service discovery: %s discovery error (status: %d)", service, resp.StatusCode)
			}

			assert.True(t, resp.StatusCode < 500,
				"Service %s should be discoverable without server error, got status %d", service, resp.StatusCode)
		}
	})
}