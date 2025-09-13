package integration

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED PHASE: Service Mesh Reliability Tests
// Tests validate service mesh communication reliability and port allocation

func TestServiceMeshReliability_ConsolidatedServiceCommunication(t *testing.T) {
	sharedtesting.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test service-to-service communication through Dapr
	serviceCommunicationTests := []struct {
		sourceService string
		targetService string
		description   string
	}{
		{"public-gateway", "content", "Public gateway to content service communication"},
		{"admin-gateway", "inquiries", "Admin gateway to inquiries service communication"},
		{"content", "notifications", "Content to notifications service communication"},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	for _, test := range serviceCommunicationTests {
		t.Run(fmt.Sprintf("ServiceCommunication_%s_to_%s", test.sourceService, test.targetService), func(t *testing.T) {
			serviceInvocationURL := fmt.Sprintf("http://localhost:3500/v1.0/invoke/%s/method/health", test.targetService)
			
			req, err := http.NewRequestWithContext(ctx, "GET", serviceInvocationURL, nil)
			require.NoError(t, err, "Failed to create service invocation request")

			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300, 
					"%s - service mesh communication must be functional", test.description)
			} else {
				t.Logf("%s not working (expected in RED phase): %v", test.description, err)
				t.Fail()
			}
		})
	}
}

func TestServiceMeshReliability_PortAllocation(t *testing.T) {
	sharedtesting.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Expected unique port allocations for sidecars
	expectedPorts := map[int]string{
		50010: "public-gateway-dapr",
		50020: "admin-gateway-dapr", 
		50030: "content-dapr",
		50040: "inquiries-dapr",
		50050: "notifications-dapr",
	}

	// Test that each port is allocated to the correct sidecar
	for port, sidecarName := range expectedPorts {
		t.Run(fmt.Sprintf("PortAllocation_%d_%s", port, sidecarName), func(t *testing.T) {
			// Test port accessibility for sidecar
			client := &http.Client{Timeout: 5 * time.Second}
			healthURL := fmt.Sprintf("http://localhost:%d/v1.0/healthz", port)
			
			req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
			require.NoError(t, err, "Failed to create port health request")

			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300, 
					"Port %d must be accessible for %s", port, sidecarName)
			} else {
				t.Logf("Port %d not accessible for %s (expected in RED phase): %v", port, sidecarName, err)
				t.Fail()
			}
		})
	}
}