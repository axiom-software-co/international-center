package integration

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED PHASE: Service Method Invocation Tests
// These tests validate that gateways invoke backend service methods correctly (not entity lookups)

func TestServiceMethodInvocation_MethodCallsNotEntityLookups(t *testing.T) {
	// This test requires complete environment health - enforcing axiom rule
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Service method invocation tests (should call methods, not search for entities)
	methodInvocationTests := []struct {
		gatewayName      string
		gatewayEndpoint  string
		targetService    string
		expectedMethod   string
		currentError     string
		description      string
		critical         bool
	}{
		{
			gatewayName:     "public-gateway",
			gatewayEndpoint: "http://localhost:9001/api/news",
			targetService:   "content",
			expectedMethod:  "api/news",
			currentError:    "content API endpoint with ID /api/news not found",
			description:     "Public gateway must invoke content service api/news method (not search for entity)",
			critical:        true,
		},
		{
			gatewayName:     "public-gateway",
			gatewayEndpoint: "http://localhost:9001/api/events",
			targetService:   "content",
			expectedMethod:  "api/events",
			currentError:    "content API endpoint with ID /api/events not found",
			description:     "Public gateway must invoke content service api/events method (not search for entity)",
			critical:        true,
		},
		{
			gatewayName:     "admin-gateway",
			gatewayEndpoint: "http://localhost:9000/api/admin/inquiries",
			targetService:   "inquiries",
			expectedMethod:  "api/inquiries",
			currentError:    "inquiries API endpoint with ID /api/admin/inquiries not found",
			description:     "Admin gateway must invoke inquiries service api/inquiries method (not search for entity)",
			critical:        true,
		},
		{
			gatewayName:     "admin-gateway",
			gatewayEndpoint: "http://localhost:9000/api/admin/subscribers",
			targetService:   "notifications",
			expectedMethod:  "api/subscribers",
			currentError:    "notifications API endpoint with ID /api/admin/subscribers not found",
			description:     "Admin gateway must invoke notifications service api/subscribers method (not search for entity)",
			critical:        false,
		},
	}

	client := &http.Client{Timeout: 15 * time.Second}

	// Act & Assert: Test service method invocation (not entity lookup)
	for _, invocation := range methodInvocationTests {
		t.Run("MethodInvocation_"+invocation.gatewayName+"_"+invocation.targetService, func(t *testing.T) {
			// First verify that direct service method works
			directServiceURL := "http://localhost:3500/v1.0/invoke/" + invocation.targetService + "/method/" + invocation.expectedMethod
			directReq, err := http.NewRequestWithContext(ctx, "GET", directServiceURL, nil)
			require.NoError(t, err, "Failed to create direct service method request")

			directResp, err := client.Do(directReq)
			require.NoError(t, err, "Direct service method must be accessible")
			defer directResp.Body.Close()

			assert.True(t, directResp.StatusCode >= 200 && directResp.StatusCode < 300,
				"Target service %s method %s must be functional", invocation.targetService, invocation.expectedMethod)

			// Now test gateway service method invocation
			gatewayReq, err := http.NewRequestWithContext(ctx, "GET", invocation.gatewayEndpoint, nil)
			require.NoError(t, err, "Failed to create gateway method invocation request")

			gatewayResp, err := client.Do(gatewayReq)
			if err == nil {
				defer gatewayResp.Body.Close()

				body, err := io.ReadAll(gatewayResp.Body)
				if err == nil {
					responseStr := string(body)
					
					if invocation.critical {
						// Critical invocations must not return entity lookup errors
						assert.NotContains(t, responseStr, "API endpoint with ID",
							"%s - gateway must invoke service method (not search for entity with path as ID)", invocation.description)
						assert.NotContains(t, responseStr, invocation.currentError,
							"%s - gateway must not treat API path as entity ID", invocation.description)
						
						// Should return backend service method response (not entity not found)
						if gatewayResp.StatusCode >= 200 && gatewayResp.StatusCode < 300 {
							var jsonData interface{}
							assert.NoError(t, json.Unmarshal(body, &jsonData),
								"%s - must return backend service method response", invocation.description)
						} else {
							// Even error responses should be from service method (not entity lookup)
							assert.NotEqual(t, 404, gatewayResp.StatusCode,
								"%s - service method should not return 404 entity not found", invocation.description)
						}
					} else {
						// Non-critical invocations should work for complete environment
						if strings.Contains(responseStr, "API endpoint with ID") {
							t.Logf("%s still using entity lookup (expected until method invocation fixed)", invocation.description)
						}
					}
				}
			} else {
				if invocation.critical {
					t.Errorf("%s not accessible: %v", invocation.description, err)
				}
			}
		})
	}
}

func TestServiceMethodInvocation_ServiceMeshDirectVsGatewayComparison(t *testing.T) {
	// Test comparison between direct service mesh calls (working) and gateway calls (entity lookup errors)
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Comparison tests between working direct calls and failing gateway calls
	methodComparisonTests := []struct {
		serviceName     string
		methodName      string
		directURL       string
		gatewayURL      string
		description     string
	}{
		{
			serviceName: "content",
			methodName:  "api/news",
			directURL:   "http://localhost:3500/v1.0/invoke/content/method/api/news",
			gatewayURL:  "http://localhost:9001/api/news",
			description: "Content news API - direct service mesh working, gateway entity lookup failing",
		},
		{
			serviceName: "content",
			methodName:  "api/events",
			directURL:   "http://localhost:3500/v1.0/invoke/content/method/api/events",
			gatewayURL:  "http://localhost:9001/api/events",
			description: "Content events API - direct service mesh working, gateway entity lookup failing",
		},
		{
			serviceName: "inquiries",
			methodName:  "api/inquiries",
			directURL:   "http://localhost:3500/v1.0/invoke/inquiries/method/api/inquiries",
			gatewayURL:  "http://localhost:9000/api/admin/inquiries",
			description: "Inquiries API - direct service mesh working, gateway entity lookup failing",
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Act & Assert: Compare direct vs gateway method invocation
	for _, comparison := range methodComparisonTests {
		t.Run("MethodComparison_"+comparison.serviceName+"_"+comparison.methodName, func(t *testing.T) {
			// Test direct service mesh method call (should work)
			directReq, err := http.NewRequestWithContext(ctx, "GET", comparison.directURL, nil)
			require.NoError(t, err, "Failed to create direct method call")

			directResp, err := client.Do(directReq)
			require.NoError(t, err, "Direct service mesh method call must work")
			defer directResp.Body.Close()

			assert.True(t, directResp.StatusCode >= 200 && directResp.StatusCode < 300,
				"Direct service mesh method %s must be functional", comparison.methodName)

			directBody, err := io.ReadAll(directResp.Body)
			require.NoError(t, err, "Failed to read direct method response")

			// Test gateway method call (currently failing with entity lookup)
			gatewayReq, err := http.NewRequestWithContext(ctx, "GET", comparison.gatewayURL, nil)
			require.NoError(t, err, "Failed to create gateway method call")

			gatewayResp, err := client.Do(gatewayReq)
			require.NoError(t, err, "Gateway method call must be accessible")
			defer gatewayResp.Body.Close()

			gatewayBody, err := io.ReadAll(gatewayResp.Body)
			require.NoError(t, err, "Failed to read gateway method response")

			// Gateway should return same type of response as direct service mesh call
			if gatewayResp.StatusCode >= 200 && gatewayResp.StatusCode < 300 {
				// Both should be successful JSON responses
				var directData, gatewayData interface{}
				assert.NoError(t, json.Unmarshal(directBody, &directData),
					"Direct service mesh call must return JSON")
				assert.NoError(t, json.Unmarshal(gatewayBody, &gatewayData),
					"Gateway call must return JSON (not entity lookup error)")
			} else {
				// Gateway should not return entity lookup errors
				gatewayResponseStr := string(gatewayBody)
				assert.NotContains(t, gatewayResponseStr, "API endpoint with ID",
					"%s - gateway must invoke method (not search for entity)", comparison.description)
				assert.NotContains(t, gatewayResponseStr, "ENTITY_NOT_FOUND",
					"%s - gateway must not treat method call as entity lookup", comparison.description)
			}

			t.Logf("Direct service mesh: %d, Gateway: %d", directResp.StatusCode, gatewayResp.StatusCode)
			t.Logf("Direct response: %s", string(directBody))
			t.Logf("Gateway response: %s", string(gatewayBody))
		})
	}
}

func TestServiceMethodInvocation_BackendServiceMethodAvailability(t *testing.T) {
	// Test that backend services have the methods that gateways are trying to invoke
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Backend service methods that gateways should be able to invoke
	serviceMethodTests := []struct {
		serviceName string
		methodName  string
		directURL   string
		description string
	}{
		{
			serviceName: "content",
			methodName:  "api/news",
			directURL:   "http://localhost:3500/v1.0/invoke/content/method/api/news",
			description: "Content service must have api/news method available for gateway invocation",
		},
		{
			serviceName: "content",
			methodName:  "api/events",
			directURL:   "http://localhost:3500/v1.0/invoke/content/method/api/events",
			description: "Content service must have api/events method available for gateway invocation",
		},
		{
			serviceName: "inquiries",
			methodName:  "api/inquiries",
			directURL:   "http://localhost:3500/v1.0/invoke/inquiries/method/api/inquiries",
			description: "Inquiries service must have api/inquiries method available for gateway invocation",
		},
		{
			serviceName: "notifications",
			methodName:  "api/subscribers",
			directURL:   "http://localhost:3500/v1.0/invoke/notifications/method/api/subscribers",
			description: "Notifications service must have api/subscribers method available for gateway invocation",
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Act & Assert: Validate backend service methods are available
	for _, method := range serviceMethodTests {
		t.Run("ServiceMethod_"+method.serviceName+"_"+method.methodName, func(t *testing.T) {
			req, err := http.NewRequestWithContext(ctx, "GET", method.directURL, nil)
			require.NoError(t, err, "Failed to create service method request")

			resp, err := client.Do(req)
			require.NoError(t, err, "Service method must be accessible")
			defer resp.Body.Close()

			assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300,
				"%s - service method must be available for gateway invocation", method.description)

			// Validate method returns structured response (not entity lookup error)
			body, err := io.ReadAll(resp.Body)
			if err == nil {
				responseStr := string(body)
				
				// Should not return entity lookup errors
				assert.NotContains(t, responseStr, "API endpoint with ID",
					"%s - backend service method should not return entity lookup errors", method.description)
				assert.NotContains(t, responseStr, "ENTITY_NOT_FOUND",
					"%s - backend service method should be available", method.description)
				
				// Should return valid JSON from service method
				var jsonData interface{}
				assert.NoError(t, json.Unmarshal(body, &jsonData),
					"%s - service method must return valid JSON response", method.description)
			}
		})
	}
}

func TestServiceMethodInvocation_GatewayServiceInvocationBehavior(t *testing.T) {
	// Test current gateway service invocation behavior vs expected behavior
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Gateway invocation behavior tests
	invocationBehaviorTests := []struct {
		gatewayName        string
		gatewayEndpoint    string
		targetService      string
		expectedBehavior   string
		currentBehavior    string
		description        string
	}{
		{
			gatewayName:      "public-gateway",
			gatewayEndpoint:  "http://localhost:9001/api/news",
			targetService:    "content",
			expectedBehavior: "invoke content service api/news method",
			currentBehavior:  "search for content entity with ID '/api/news'",
			description:      "Public gateway service invocation behavior must be method call not entity lookup",
		},
		{
			gatewayName:      "admin-gateway",
			gatewayEndpoint:  "http://localhost:9000/api/admin/inquiries",
			targetService:    "inquiries",
			expectedBehavior: "invoke inquiries service api/inquiries method",
			currentBehavior:  "search for inquiries entity with ID '/api/admin/inquiries'",
			description:      "Admin gateway service invocation behavior must be method call not entity lookup",
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Act & Assert: Test service invocation behavior
	for _, behavior := range invocationBehaviorTests {
		t.Run("InvocationBehavior_"+behavior.gatewayName+"_"+behavior.targetService, func(t *testing.T) {
			req, err := http.NewRequestWithContext(ctx, "GET", behavior.gatewayEndpoint, nil)
			require.NoError(t, err, "Failed to create gateway invocation request")

			resp, err := client.Do(req)
			require.NoError(t, err, "Gateway invocation must be accessible")
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err, "Failed to read gateway invocation response")

			responseStr := string(body)
			
			// Should NOT exhibit current (incorrect) behavior
			assert.NotContains(t, responseStr, "API endpoint with ID",
				"%s - must not exhibit current behavior: %s", behavior.description, behavior.currentBehavior)
			assert.NotContains(t, responseStr, "ENTITY_NOT_FOUND",
				"%s - must not search for entities by API path", behavior.description)
			
			// Should exhibit expected behavior (method invocation)
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				var jsonData interface{}
				assert.NoError(t, json.Unmarshal(body, &jsonData),
					"%s - must exhibit expected behavior: %s", behavior.description, behavior.expectedBehavior)
			} else {
				// Even error responses should be from service method (not entity lookup)
				t.Logf("%s current behavior: %s (expected: %s)", behavior.description, behavior.currentBehavior, behavior.expectedBehavior)
				t.Fail() // Should fail until GREEN phase fixes method invocation
			}
		})
	}
}

// validateEnvironmentPrerequisites ensures environment health before integration testing
func validateEnvironmentPrerequisites(t *testing.T) {
	// Check critical infrastructure, platform, service, and gateway components are running
	criticalContainers := []string{"postgresql", "dapr-control-plane", "content", "inquiries", "notifications", "admin-gateway"}
	
	for _, container := range criticalContainers {
		cmd := exec.Command("podman", "ps", "--filter", "name="+container, "--format", "{{.Names}}")
		output, err := cmd.Output()
		require.NoError(t, err, "Failed to check critical container %s", container)

		if !strings.Contains(string(output), container) {
			t.Skipf("Critical container %s not running - environment not ready for integration testing", container)
		}
	}
}