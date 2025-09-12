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

// RED PHASE: Dapr Component Configuration Tests
// These tests validate that Dapr state store, pub/sub, and other components are properly configured

func TestDaprComponentConfiguration_StateStoreAvailability(t *testing.T) {
	// This test requires complete environment health - enforcing axiom rule
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test Dapr state store component availability
	t.Run("StateStore_ComponentAccess", func(t *testing.T) {
		// Test state store component through Dapr control plane
		client := &http.Client{Timeout: 10 * time.Second}
		
		// Check Dapr components endpoint
		componentsURL := "http://localhost:3500/v1.0/components"
		req, err := http.NewRequestWithContext(ctx, "GET", componentsURL, nil)
		require.NoError(t, err, "Failed to create components request")

		resp, err := client.Do(req)
		require.NoError(t, err, "Dapr components endpoint must be accessible")
		defer resp.Body.Close()

		assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300,
			"Dapr components endpoint must be operational")

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "Failed to read components response")

		var components []map[string]interface{}
		err = json.Unmarshal(body, &components)
		require.NoError(t, err, "Components response must be valid JSON")

		// Should have state store component configured
		hasStateStore := false
		for _, component := range components {
			if componentType, exists := component["type"]; exists {
				if strings.Contains(strings.ToLower(componentType.(string)), "state") {
					hasStateStore = true
					break
				}
			}
		}

		assert.True(t, hasStateStore, 
			"Dapr must have state store component configured for service data persistence")

		t.Logf("Dapr components found: %d", len(components))
		for _, component := range components {
			if name, exists := component["name"]; exists {
				if componentType, exists := component["type"]; exists {
					t.Logf("Component: %s, Type: %s", name, componentType)
				}
			}
		}
	})

	// Test state store functionality
	t.Run("StateStore_Functionality", func(t *testing.T) {
		client := &http.Client{Timeout: 10 * time.Second}
		
		// Test state store through Dapr state API
		stateURL := "http://localhost:3500/v1.0/state/statestore/test-key"
		
		// Try to read from state store
		req, err := http.NewRequestWithContext(ctx, "GET", stateURL, nil)
		require.NoError(t, err, "Failed to create state store request")

		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			// State store should be accessible (404 for missing key is acceptable)
			assert.True(t, resp.StatusCode == 404 || (resp.StatusCode >= 200 && resp.StatusCode < 300),
				"Dapr state store must be accessible for service data operations")
		} else {
			t.Errorf("Dapr state store not accessible: %v", err)
		}
	})
}

func TestDaprComponentConfiguration_PubSubAvailability(t *testing.T) {
	// Test Dapr pub/sub component configuration
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test pub/sub component availability
	t.Run("PubSub_ComponentConfiguration", func(t *testing.T) {
		client := &http.Client{Timeout: 10 * time.Second}
		
		// Test pub/sub through Dapr publish endpoint
		publishURL := "http://localhost:3500/v1.0/publish/pubsub/test-topic"
		testMessage := `{"message": "test"}`
		
		req, err := http.NewRequestWithContext(ctx, "POST", publishURL, strings.NewReader(testMessage))
		require.NoError(t, err, "Failed to create pub/sub request")
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			// Pub/sub should be accessible (even if component not configured, should get proper error)
			assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500,
				"Dapr pub/sub endpoint must be accessible")
		} else {
			t.Errorf("Dapr pub/sub not accessible: %v", err)
		}
	})

	// Test pub/sub component integration with messaging infrastructure
	t.Run("PubSub_MessagingIntegration", func(t *testing.T) {
		// Verify RabbitMQ is available for pub/sub integration
		cmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name=rabbitmq", "--format", "{{.Names}}")
		output, err := cmd.Output()
		require.NoError(t, err, "Failed to check RabbitMQ container")

		runningContainers := strings.TrimSpace(string(output))
		assert.Contains(t, runningContainers, "rabbitmq",
			"RabbitMQ must be running for Dapr pub/sub component integration")
	})
}

func TestDaprComponentConfiguration_ServiceStateIntegration(t *testing.T) {
	// Test that services can integrate with Dapr state store for data operations
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Services that should be able to use Dapr state store
	serviceStateTests := []struct {
		serviceName     string
		healthEndpoint  string
		dataEndpoint    string
		stateOperation  string
		description     string
	}{
		{
			serviceName:    "content",
			healthEndpoint: "http://localhost:3500/v1.0/invoke/content/method/health",
			dataEndpoint:   "http://localhost:3500/v1.0/invoke/content/method/api/news",
			stateOperation: "news data persistence",
			description:    "Content service must integrate with Dapr state store for news data operations",
		},
		{
			serviceName:    "inquiries",
			healthEndpoint: "http://localhost:3500/v1.0/invoke/inquiries/method/health",
			dataEndpoint:   "http://localhost:3500/v1.0/invoke/inquiries/method/api/inquiries",
			stateOperation: "inquiry data persistence",
			description:    "Inquiries service must integrate with Dapr state store for inquiry data operations",
		},
		{
			serviceName:    "notifications",
			healthEndpoint: "http://localhost:3500/v1.0/invoke/notifications/method/health",
			dataEndpoint:   "http://localhost:3500/v1.0/invoke/notifications/method/api/subscribers",
			stateOperation: "subscriber data persistence",
			description:    "Notifications service must integrate with Dapr state store for subscriber data operations",
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Act & Assert: Test service state integration
	for _, stateTest := range serviceStateTests {
		t.Run("ServiceState_"+stateTest.serviceName, func(t *testing.T) {
			// Verify service is healthy
			healthReq, err := http.NewRequestWithContext(ctx, "GET", stateTest.healthEndpoint, nil)
			require.NoError(t, err, "Failed to create service health request")

			healthResp, err := client.Do(healthReq)
			require.NoError(t, err, "Service health must be accessible")
			defer healthResp.Body.Close()

			assert.True(t, healthResp.StatusCode >= 200 && healthResp.StatusCode < 300,
				"Service %s must be healthy for state integration testing", stateTest.serviceName)

			// Test service data operations (should work with Dapr state store)
			dataReq, err := http.NewRequestWithContext(ctx, "GET", stateTest.dataEndpoint, nil)
			require.NoError(t, err, "Failed to create service data request")

			dataResp, err := client.Do(dataReq)
			require.NoError(t, err, "Service data endpoint must be accessible")
			defer dataResp.Body.Close()

			assert.True(t, dataResp.StatusCode >= 200 && dataResp.StatusCode < 300,
				"%s - service data operations must be functional", stateTest.description)

			// Validate response structure indicates state management readiness
			body, err := io.ReadAll(dataResp.Body)
			if err == nil {
				var jsonData map[string]interface{}
				assert.NoError(t, json.Unmarshal(body, &jsonData),
					"%s - service must return JSON for %s", stateTest.description, stateTest.stateOperation)

				// Should have data structure ready for state operations
				assert.Contains(t, jsonData, "data",
					"%s - service response must have data field for state operations", stateTest.description)
				assert.Contains(t, jsonData, "count",
					"%s - service response must have count field for state operations", stateTest.description)
			}
		})
	}
}

func TestDaprComponentConfiguration_CrossServiceCommunication(t *testing.T) {
	// Test that cross-service communication works through Dapr service mesh
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Cross-service communication patterns that should work through Dapr
	crossServiceTests := []struct {
		communicationPattern string
		sourceService        string
		targetService        string
		testEndpoint         string
		description          string
	}{
		{
			communicationPattern: "content-to-notifications",
			sourceService:        "content",
			targetService:        "notifications",
			testEndpoint:         "http://localhost:3500/v1.0/invoke/notifications/method/health",
			description:          "Content service must communicate with notifications through Dapr for event publishing",
		},
		{
			communicationPattern: "inquiries-to-notifications",
			sourceService:        "inquiries",
			targetService:        "notifications",
			testEndpoint:         "http://localhost:3500/v1.0/invoke/notifications/method/health",
			description:          "Inquiries service must communicate with notifications through Dapr for inquiry alerts",
		},
		{
			communicationPattern: "gateway-to-content",
			sourceService:        "public-gateway",
			targetService:        "content",
			testEndpoint:         "http://localhost:3500/v1.0/invoke/content/method/health",
			description:          "Public gateway must communicate with content service through Dapr for API routing",
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Act & Assert: Test cross-service communication
	for _, communication := range crossServiceTests {
		t.Run("CrossService_"+communication.communicationPattern, func(t *testing.T) {
			// Test that target service is accessible through service mesh
			req, err := http.NewRequestWithContext(ctx, "GET", communication.testEndpoint, nil)
			require.NoError(t, err, "Failed to create cross-service communication request")

			resp, err := client.Do(req)
			require.NoError(t, err, "Cross-service communication must be accessible")
			defer resp.Body.Close()

			assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300,
				"%s - cross-service communication must be functional", communication.description)

			// Validate service mesh communication works
			body, err := io.ReadAll(resp.Body)
			if err == nil {
				var jsonData interface{}
				assert.NoError(t, json.Unmarshal(body, &jsonData),
					"%s - cross-service communication must return valid JSON", communication.description)
			}
		})
	}
}

// validateEnvironmentPrerequisites ensures environment health before integration testing
func validateEnvironmentPrerequisites(t *testing.T) {
	// Check critical infrastructure, platform, service, and gateway components are running
	criticalContainers := []string{"postgresql", "dapr-control-plane", "content", "inquiries", "notifications", "public-gateway", "admin-gateway"}
	
	for _, container := range criticalContainers {
		cmd := exec.Command("podman", "ps", "--filter", "name="+container, "--format", "{{.Names}}")
		output, err := cmd.Output()
		require.NoError(t, err, "Failed to check critical container %s", container)

		if !strings.Contains(string(output), container) {
			t.Skipf("Critical container %s not running - environment not ready for integration testing", container)
		}
	}
}