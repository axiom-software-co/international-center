package testing

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	sharedtesting "github.com/axiom-software-co/international-center/src/deployer/shared/testing"
)

// TestBackendServiceDeployment validates backend service containers with Dapr sidecars
func TestBackendServiceDeployment(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("Content_API_Service_Deployment", func(t *testing.T) {
		// Arrange - Content API should be deployed with Dapr sidecar
		contentAPIURL := sharedtesting.GetEnvVar("CONTENT_API_URL", "http://localhost:8001")
		daprHTTPPort := sharedtesting.GetEnvVar("DAPR_HTTP_PORT", "3500")
		
		// Act - Attempt to connect to Content API service
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/health", contentAPIURL), nil)
		require.NoError(t, err, "Should be able to create HTTP request")
		resp, err := client.Do(req)
		
		// Assert - Content API should be accessible and healthy
		require.NoError(t, err, "Content API service should be accessible")
		require.NotNil(t, resp, "HTTP response should not be nil")
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Content API should return healthy status")
		resp.Body.Close()
		
		// Act - Verify Dapr sidecar is accessible
		daprReq, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://localhost:%s/v1.0/healthz", daprHTTPPort), nil)
		require.NoError(t, err, "Should be able to create Dapr HTTP request")
		daprResp, err := client.Do(daprReq)
		require.NoError(t, err, "Dapr sidecar should be accessible for Content API")
		assert.Equal(t, http.StatusOK, daprResp.StatusCode, "Dapr sidecar should be healthy")
		daprResp.Body.Close()
		
		t.Logf("Content API service deployed successfully at %s", contentAPIURL)
	})

	t.Run("Services_API_Service_Deployment", func(t *testing.T) {
		// Arrange - Services API should be deployed with Dapr sidecar
		servicesAPIURL := sharedtesting.GetEnvVar("SERVICES_API_URL", "http://localhost:8002")
		
		// Act - Attempt to connect to Services API service
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/health", servicesAPIURL), nil)
		require.NoError(t, err, "Should be able to create HTTP request")
		resp, err := client.Do(req)
		
		// Assert - Services API should be accessible and healthy
		require.NoError(t, err, "Services API service should be accessible")
		require.NotNil(t, resp, "HTTP response should not be nil")
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Services API should return healthy status")
		resp.Body.Close()
		
		t.Logf("Services API service deployed successfully at %s", servicesAPIURL)
	})

	t.Run("Public_Gateway_Service_Deployment", func(t *testing.T) {
		// Arrange - Public Gateway should be deployed with Dapr sidecar
		publicGatewayURL := sharedtesting.GetEnvVar("PUBLIC_GATEWAY_URL", "http://localhost:8003")
		
		// Act - Attempt to connect to Public Gateway service
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/health", publicGatewayURL), nil)
		require.NoError(t, err, "Should be able to create HTTP request")
		resp, err := client.Do(req)
		
		// Assert - Public Gateway should be accessible and healthy
		require.NoError(t, err, "Public Gateway service should be accessible")
		require.NotNil(t, resp, "HTTP response should not be nil")
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Public Gateway should return healthy status")
		resp.Body.Close()
		
		t.Logf("Public Gateway service deployed successfully at %s", publicGatewayURL)
	})

	t.Run("Admin_Gateway_Service_Deployment", func(t *testing.T) {
		// Arrange - Admin Gateway should be deployed with Dapr sidecar
		adminGatewayURL := sharedtesting.GetEnvVar("ADMIN_GATEWAY_URL", "http://localhost:8004")
		
		// Act - Attempt to connect to Admin Gateway service
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/health", adminGatewayURL), nil)
		require.NoError(t, err, "Should be able to create HTTP request")
		resp, err := client.Do(req)
		
		// Assert - Admin Gateway should be accessible and healthy
		require.NoError(t, err, "Admin Gateway service should be accessible")
		require.NotNil(t, resp, "HTTP response should not be nil")
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Admin Gateway should return healthy status")
		resp.Body.Close()
		
		t.Logf("Admin Gateway service deployed successfully at %s", adminGatewayURL)
	})
}

// TestServiceToServiceCommunication validates Dapr service invocation between services
func TestServiceOrchestrationCommunication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("Gateway_to_Backend_Communication", func(t *testing.T) {
		// Arrange - Public Gateway should be able to invoke Content API via Dapr
		daprHTTPPort := sharedtesting.GetEnvVar("DAPR_HTTP_PORT", "3500")
		
		// Act - Invoke Content API through Dapr service invocation
		client := &http.Client{Timeout: 10 * time.Second}
		invokeURL := fmt.Sprintf("http://localhost:%s/v1.0/invoke/content-api/method/health", daprHTTPPort)
		req, err := http.NewRequestWithContext(ctx, "GET", invokeURL, nil)
		require.NoError(t, err, "Should be able to create service invocation request")
		resp, err := client.Do(req)
		
		// Assert - Service invocation should work
		require.NoError(t, err, "Dapr service invocation should work between Gateway and Content API")
		require.NotNil(t, resp, "Service invocation response should not be nil")
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Content API should respond via service invocation")
		resp.Body.Close()
		
		t.Log("Service-to-service communication validated via Dapr service invocation")
	})

	t.Run("Cross_Service_API_Communication", func(t *testing.T) {
		// Arrange - Services API should be able to invoke Content API via Dapr
		daprHTTPPort := sharedtesting.GetEnvVar("DAPR_HTTP_PORT", "3500")
		
		// Act - Test cross-service communication
		client := &http.Client{Timeout: 10 * time.Second}
		invokeURL := fmt.Sprintf("http://localhost:%s/v1.0/invoke/services-api/method/health", daprHTTPPort)
		req, err := http.NewRequestWithContext(ctx, "GET", invokeURL, nil)
		require.NoError(t, err, "Should be able to create cross-service request")
		resp, err := client.Do(req)
		
		// Assert - Cross-service communication should work
		require.NoError(t, err, "Cross-service communication should work via Dapr")
		require.NotNil(t, resp, "Cross-service response should not be nil")
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Services API should respond via service invocation")
		resp.Body.Close()
		
		t.Log("Cross-service API communication validated")
	})
}

// TestServiceContainerHealth validates individual service container health
func TestServiceContainerHealth(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	services := map[string]string{
		"content-api":     sharedtesting.GetEnvVar("CONTENT_API_URL", "http://localhost:8001"),
		"services-api":    sharedtesting.GetEnvVar("SERVICES_API_URL", "http://localhost:8002"),
		"public-gateway":  sharedtesting.GetEnvVar("PUBLIC_GATEWAY_URL", "http://localhost:8003"),
		"admin-gateway":   sharedtesting.GetEnvVar("ADMIN_GATEWAY_URL", "http://localhost:8004"),
	}

	for serviceName, serviceURL := range services {
		t.Run(fmt.Sprintf("%s_Container_Health", serviceName), func(t *testing.T) {
			// Arrange - Service container should be running and healthy
			client := &http.Client{Timeout: 5 * time.Second}
			
			// Act - Check service health endpoint
			healthURL := fmt.Sprintf("%s/health", serviceURL)
			req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
			require.NoError(t, err, "Should be able to create health check request")
			resp, err := client.Do(req)
			
			// Assert - Service should be healthy
			require.NoError(t, err, fmt.Sprintf("%s service container should be accessible", serviceName))
			require.NotNil(t, resp, "Health check response should not be nil")
			assert.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("%s service should be healthy", serviceName))
			
			if resp.Body != nil {
				resp.Body.Close()
			}
			
			t.Logf("%s service container is healthy at %s", serviceName, serviceURL)
		})
	}
}