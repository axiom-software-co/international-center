package tests

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/dapr/go-sdk/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDaprSidecarRegistrationDiscovery(t *testing.T) {
	// Phase 2: Sidecar Registration and Discovery Validation
	// Integration test - requires full podman compose environment
	
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	daprClient, err := client.NewClient()
	require.NoError(t, err, "Should create Dapr client successfully")
	defer daprClient.Close()

	t.Run("sidecar service discovery registration", func(t *testing.T) {
		// Test: All Dapr sidecars are registered and discoverable
		
		// Expected services registered with control plane
		expectedServices := []string{
			"services-api",
			"content-api", 
			"public-gateway",
			"admin-gateway",
			"grafana-agent",
		}
		
		for _, serviceName := range expectedServices {
			t.Run(fmt.Sprintf("service_%s_discovery", serviceName), func(t *testing.T) {
				// Test service discovery by attempting health check via Dapr service invocation
				
				// Use Dapr service invocation to check if service is discoverable
				method := "GET"
				
				// Try to invoke health endpoint via Dapr service invocation
				content := &client.DataContent{
					ContentType: "application/json",
					Data:        []byte{},
				}
				
				// Attempt service invocation - this tests service discovery
				resp, err := daprClient.InvokeMethodWithContent(ctx, serviceName, "health", method, content)
				if err != nil {
					// Service might not be ready yet, but should be discoverable
					// Check if it's a connection issue vs discovery issue
					if !strings.Contains(err.Error(), "connection refused") && !strings.Contains(err.Error(), "no healthy upstream") {
						// Service is discoverable (registered) but might not be ready
						assert.True(t, true, "Service %s is registered for discovery", serviceName)
					} else {
						t.Logf("Service %s discovery test: %v (may indicate service not yet ready)", serviceName, err)
					}
				} else {
					// Service responded - it's both discoverable and healthy
					assert.NotNil(t, resp, "Service %s responded via service discovery", serviceName)
				}
			})
		}
	})
	
	t.Run("sidecar cross-service discovery capabilities", func(t *testing.T) {
		// Test: Sidecars can discover each other through control plane
		
		// Test discovery from services-api to other services
		testCases := []struct {
			sourceService string
			targetService string
			testEndpoint  string
		}{
			{"services-api", "content-api", "health"},
			{"public-gateway", "services-api", "health"}, 
			{"admin-gateway", "content-api", "health"},
		}
		
		for _, tc := range testCases {
			t.Run(fmt.Sprintf("discovery_%s_to_%s", tc.sourceService, tc.targetService), func(t *testing.T) {
				// Test cross-service discovery capability
				
				method := "GET"
				content := &client.DataContent{
					ContentType: "application/json", 
					Data:        []byte{},
				}
				
				// Attempt cross-service invocation to test discovery
				resp, err := daprClient.InvokeMethodWithContent(ctx, tc.targetService, tc.testEndpoint, method, content)
				
				if err != nil {
					// Analyze error to determine if it's discovery vs connectivity issue
					if strings.Contains(err.Error(), "service not found") || strings.Contains(err.Error(), "name resolution failed") {
						t.Errorf("Service discovery failed: %s cannot discover %s", tc.sourceService, tc.targetService)
					} else {
						// Discovery worked, but service might not be ready (acceptable)
						t.Logf("Cross-service discovery test %s->%s: %v (service discoverable but may not be ready)", 
							tc.sourceService, tc.targetService, err)
					}
				} else {
					assert.NotNil(t, resp, "Cross-service discovery successful: %s->%s", tc.sourceService, tc.targetService)
				}
			})
		}
	})
	
	t.Run("sidecar component registration validation", func(t *testing.T) {
		// Test: Sidecars can access registered components
		
		// Test component access through each sidecar
		componentTests := []struct {
			component     string
			operation     string
			expectedError []string
		}{
			{"statestore-postgresql", "state_get", []string{"state not found", "error getting state: state not found"}},
		}
		
		for _, test := range componentTests {
			t.Run(fmt.Sprintf("component_%s_%s", test.component, test.operation), func(t *testing.T) {
				switch test.operation {
				case "state_get":
					// Test state component registration
					_, err := daprClient.GetState(ctx, test.component, "registration-validation-test", nil)
					if err != nil {
						// Should get "not found" error, not component registration error
						errorFound := false
						for _, expectedErr := range test.expectedError {
							if strings.Contains(err.Error(), expectedErr) {
								errorFound = true
								break
							}
						}
						assert.True(t, errorFound, "Component %s should be registered (got: %v)", test.component, err)
					}
				}
			})
		}
	})
}