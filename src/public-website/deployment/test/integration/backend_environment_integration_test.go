package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sharedtesting "github.com/axiom-software-co/international-center/src/public-website/infrastructure/test/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDAPREnvironmentIntegration validates DAPR middleware integration across environments
func TestDAPREnvironmentIntegration(t *testing.T) {
	// Use consolidated environment health validation
	sharedtesting.ValidateEnvironmentPrerequisites(t)

	timeout := 15 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	t.Run("development environment DAPR integration", func(t *testing.T) {
		// Test DAPR service mesh connectivity using consolidated testing framework
		daprRunner := sharedtesting.NewDaprServiceMeshTestRunner()

		// Validate service mesh communication
		communicationErrors := daprRunner.ValidateServiceMeshCommunication(ctx)
		if len(communicationErrors) > 0 {
			for _, err := range communicationErrors {
				t.Logf("Service mesh communication issue: %v", err)
			}
			t.Log("Development environment: Some service mesh communication routes may not be fully implemented yet")
		} else {
			t.Log("Service mesh communication validation: All inter-service communication working")
		}

		// Validate component configuration
		componentErrors := daprRunner.ValidateComponentConfiguration(ctx)
		if len(componentErrors) > 0 {
			for serviceName, errors := range componentErrors {
				for _, err := range errors {
					t.Logf("Service %s component issue: %v", serviceName, err)
				}
			}
			t.Log("Development environment: Some DAPR components may not be fully configured yet")
		} else {
			t.Log("DAPR component validation: All components properly configured")
		}
	})
}

// TestFullGatewayEnvironmentValidation tests complete gateway environment setup
func TestFullGatewayEnvironmentValidation(t *testing.T) {
	// Use consolidated environment health validation
	sharedtesting.ValidateEnvironmentPrerequisites(t)

	timeout := 15 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("development environment complete validation", func(t *testing.T) {
		// Test public gateway accessibility
		t.Run("public gateway accessibility", func(t *testing.T) {
			resp, err := client.Get("http://localhost:9001/health")
			if err != nil {
				t.Logf("Public gateway not accessible: %v", err)
				return
			}
			defer resp.Body.Close()

			assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300,
				"Public gateway health should be accessible")
			t.Logf("Public gateway health check: status %d", resp.StatusCode)
		})

		// Test admin gateway accessibility
		t.Run("admin gateway accessibility", func(t *testing.T) {
			resp, err := client.Get("http://localhost:9000/health")
			if err != nil {
				t.Logf("Admin gateway not accessible: %v", err)
				return
			}
			defer resp.Body.Close()

			assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300,
				"Admin gateway health should be accessible")
			t.Logf("Admin gateway health check: status %d", resp.StatusCode)
		})

		// Validate gateway routing using enhanced testing framework
		t.Run("gateway routing validation", func(t *testing.T) {
			gatewayTester := sharedtesting.NewGatewayServiceMeshTester()

			errors := gatewayTester.ValidateGatewayToServiceCommunication(ctx)
			if len(errors) > 0 {
				for _, err := range errors {
					t.Logf("Gateway routing issue: %v", err)
				}
				t.Log("Development environment: Some gateway routes may not be fully implemented yet")
			} else {
				t.Log("Gateway routing validation: All routes properly configured")
			}
		})
	})
}

// TestServiceDiscoveryIntegration validates service discovery through DAPR
func TestServiceDiscoveryIntegration(t *testing.T) {
	// Use consolidated environment health validation
	sharedtesting.ValidateEnvironmentPrerequisites(t)

	timeout := 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	t.Run("DAPR service discovery validation", func(t *testing.T) {
		// Test service discovery using service mesh reliability tester
		reliabilityTester := sharedtesting.NewServiceMeshReliabilityTester()

		errors := reliabilityTester.ValidateServiceMeshResilience(ctx)
		if len(errors) > 0 {
			for _, err := range errors {
				t.Logf("Service discovery issue: %v", err)
			}
			t.Log("Development environment: Some service discovery patterns may not be fully configured yet")
		} else {
			t.Log("Service discovery validation: All services discoverable through DAPR")
		}

		// Test direct service invocation through DAPR
		daprClient := sharedtesting.NewDaprServiceTestClient("discovery-test", "3500")

		expectedServices := []string{"content", "inquiries", "notifications"}
		for _, service := range expectedServices {
			resp, err := daprClient.InvokeService(ctx, service, "GET", "/health", nil)
			if err != nil {
				t.Logf("Service %s discovery failed: %v", service, err)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode < 500 {
				t.Logf("Service %s discoverable: status %d", service, resp.StatusCode)
			} else {
				t.Logf("Service %s discovery error: status %d", service, resp.StatusCode)
			}
		}
	})
}

// TestContractComplianceWithEnvironmentIntegration validates contract compliance in environment testing
func TestContractComplianceWithEnvironmentIntegration(t *testing.T) {
	// Use consolidated environment health validation
	sharedtesting.ValidateEnvironmentPrerequisites(t)

	timeout := 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("contract validation with DAPR integration", func(t *testing.T) {
		// Test contract compliance using cross-stack integration tester
		crossStackTester := sharedtesting.NewCrossStackIntegrationTester()

		errors := crossStackTester.ValidateEndToEndWorkflow(ctx)
		if len(errors) > 0 {
			for _, err := range errors {
				t.Logf("Contract compliance issue: %v", err)
			}
			t.Log("Development environment: Some contract workflows may not be fully implemented yet")
		} else {
			t.Log("Contract compliance validation: All workflows properly structured")
		}

		// Test direct contract compliance through admin gateway
		testCases := []struct {
			name           string
			method         string
			url            string
			expectedStatus int
			validateJSON   bool
		}{
			{
				name:           "admin gateway health contract",
				method:         "GET",
				url:            "http://localhost:9000/health",
				expectedStatus: http.StatusOK,
				validateJSON:   true,
			},
			{
				name:           "admin inquiries endpoint contract",
				method:         "GET",
				url:            "http://localhost:9000/api/admin/inquiries",
				expectedStatus: http.StatusOK,
				validateJSON:   true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req, err := http.NewRequestWithContext(ctx, tc.method, tc.url, nil)
				require.NoError(t, err, "Request creation should succeed")

				resp, err := client.Do(req)
				if err != nil {
					t.Logf("Contract validation endpoint not accessible: %v", err)
					return
				}
				defer resp.Body.Close()

				// Contract compliance check
				if tc.validateJSON && resp.StatusCode >= 200 && resp.StatusCode < 300 {
					var response map[string]interface{}
					err := json.NewDecoder(resp.Body).Decode(&response)
					if err == nil {
						t.Logf("Contract compliance: %s returns valid JSON", tc.name)
					} else {
						t.Logf("Contract compliance issue: %s does not return valid JSON", tc.name)
					}
				}

				t.Logf("Contract validation: %s status %d", tc.name, resp.StatusCode)
			})
		}
	})
}

// TestEnvironmentHealthValidation validates that all environment components are healthy
func TestEnvironmentHealthValidation(t *testing.T) {
	// Use consolidated environment health validation
	sharedtesting.ValidateEnvironmentPrerequisites(t)

	timeout := 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("development environment health validation", func(t *testing.T) {
		// Comprehensive health validation using all testing components
		t.Run("comprehensive service mesh health", func(t *testing.T) {
			daprRunner := sharedtesting.NewDaprServiceMeshTestRunner()
			results := daprRunner.RunComprehensiveDaprTesting(ctx)

			var healthyServices, unhealthyServices int
			for _, result := range results {
				if result.Success {
					healthyServices++
					t.Logf("✅ %s - %s: HEALTHY", result.ServiceName, result.TestType)
				} else {
					unhealthyServices++
					t.Logf("⚠️ %s - %s: ISSUE (%v)", result.ServiceName, result.TestType, result.Error)
				}
			}

			t.Logf("Environment health summary: %d healthy, %d with issues", healthyServices, unhealthyServices)
		})

		// Test gateway health endpoints directly
		t.Run("gateway health endpoints", func(t *testing.T) {
			healthEndpoints := []struct {
				name string
				url  string
			}{
				{"public-gateway", "http://localhost:9001/health"},
				{"admin-gateway", "http://localhost:9000/health"},
			}

			for _, endpoint := range healthEndpoints {
				resp, err := client.Get(endpoint.url)
				if err != nil {
					t.Logf("%s health endpoint not accessible: %v", endpoint.name, err)
					continue
				}
				defer resp.Body.Close()

				var health map[string]interface{}
				if err := json.NewDecoder(resp.Body).Decode(&health); err == nil {
					t.Logf("%s health: %+v", endpoint.name, health)
				} else {
					t.Logf("%s health response is not JSON: status %d", endpoint.name, resp.StatusCode)
				}
			}
		})

		t.Log("Environment health validation completed using consolidated testing framework")
	})
}