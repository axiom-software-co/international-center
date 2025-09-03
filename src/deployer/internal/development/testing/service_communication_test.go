package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServiceToServiceCommunication validates Dapr service invocation between APIs
func TestServiceToServiceCommunication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
	defer cancel()

	// This test validates that the content API can call the services API via Dapr service invocation
	t.Run("ContentAPI_to_ServicesAPI_Communication", func(t *testing.T) {
		// Arrange
		contentAPIURL := getRequiredEnvVar(t, "CONTENT_API_URL")
		servicesAPIURL := getRequiredEnvVar(t, "SERVICES_API_URL")

		// Act - Make request to content API that should invoke services API
		endpointURL := fmt.Sprintf("%s/api/v1/content/services-integration", contentAPIURL)
		resp, err := makeHTTPRequest(ctx, "GET", endpointURL, nil)

		// Assert - Communication should work via Dapr service invocation
		require.NoError(t, err, "Content API should be able to invoke Services API via Dapr")
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Service-to-service call should succeed")

		// Verify response contains data from both services
		var responseData map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&responseData)
		require.NoError(t, err, "Response should be valid JSON")
		
		assert.Contains(t, responseData, "content_service", "Response should include content service data")
		assert.Contains(t, responseData, "services_service", "Response should include services service data")

		// Verify Dapr service invocation headers are present
		assert.NotEmpty(t, resp.Header.Get("X-Correlation-ID"), "Correlation ID should be present")
		
		t.Logf("Services API URL validated: %s", servicesAPIURL)
	})

	// This test validates that the services API can call the content API via Dapr service invocation
	t.Run("ServicesAPI_to_ContentAPI_Communication", func(t *testing.T) {
		// Arrange
		servicesAPIURL := getRequiredEnvVar(t, "SERVICES_API_URL")
		contentAPIURL := getRequiredEnvVar(t, "CONTENT_API_URL")

		// Act - Make request to services API that should invoke content API
		endpointURL := fmt.Sprintf("%s/api/v1/services/content-integration", servicesAPIURL)
		resp, err := makeHTTPRequest(ctx, "GET", endpointURL, nil)

		// Assert - Communication should work via Dapr service invocation
		require.NoError(t, err, "Services API should be able to invoke Content API via Dapr")
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Service-to-service call should succeed")

		// Verify response contains data from both services
		var responseData map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&responseData)
		require.NoError(t, err, "Response should be valid JSON")
		
		assert.Contains(t, responseData, "services_service", "Response should include services service data")
		assert.Contains(t, responseData, "content_service", "Response should include content service data")

		// Verify Dapr service invocation headers are present
		assert.NotEmpty(t, resp.Header.Get("X-Correlation-ID"), "Correlation ID should be present")
		
		t.Logf("Content API URL validated: %s", contentAPIURL)
	})
}

// TestGatewayToAPIProxying validates gateway proxying to backend APIs
func TestGatewayToAPIProxying(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
	defer cancel()

	// Test public gateway proxying to content and services APIs
	t.Run("PublicGateway_Proxying", func(t *testing.T) {
		// Arrange
		publicGatewayURL := getRequiredEnvVar(t, "PUBLIC_GATEWAY_URL")

		// Test proxying to services API
		t.Run("Proxy_to_ServicesAPI", func(t *testing.T) {
			endpointURL := fmt.Sprintf("%s/api/v1/services", publicGatewayURL)
			resp, err := makeHTTPRequest(ctx, "GET", endpointURL, nil)

			require.NoError(t, err, "Public gateway should proxy to services API")
			assert.Equal(t, http.StatusOK, resp.StatusCode, "Services API request via gateway should succeed")

			// Verify gateway security headers are present
			assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"), "Security headers should be applied")
			assert.Equal(t, "DENY", resp.Header.Get("X-Frame-Options"), "Security headers should be applied") 
			assert.Equal(t, "1; mode=block", resp.Header.Get("X-XSS-Protection"), "Security headers should be applied")
		})

		// Test proxying to content API
		t.Run("Proxy_to_ContentAPI", func(t *testing.T) {
			endpointURL := fmt.Sprintf("%s/api/v1/content", publicGatewayURL)
			resp, err := makeHTTPRequest(ctx, "GET", endpointURL, nil)

			require.NoError(t, err, "Public gateway should proxy to content API")
			assert.Equal(t, http.StatusOK, resp.StatusCode, "Content API request via gateway should succeed")

			// Verify gateway security headers are present
			assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"), "Security headers should be applied")
			assert.Equal(t, "DENY", resp.Header.Get("X-Frame-Options"), "Security headers should be applied")
			assert.Equal(t, "1; mode=block", resp.Header.Get("X-XSS-Protection"), "Security headers should be applied")
		})
	})

	// Test admin gateway proxying to backend APIs
	t.Run("AdminGateway_Proxying", func(t *testing.T) {
		// Arrange
		adminGatewayURL := getRequiredEnvVar(t, "ADMIN_GATEWAY_URL")

		// Test proxying to services API with admin prefix
		t.Run("Proxy_to_AdminServicesAPI", func(t *testing.T) {
			endpointURL := fmt.Sprintf("%s/admin/api/v1/services", adminGatewayURL)
			resp, err := makeHTTPRequest(ctx, "GET", endpointURL, nil)

			require.NoError(t, err, "Admin gateway should proxy to services API")
			assert.Equal(t, http.StatusOK, resp.StatusCode, "Admin services API request via gateway should succeed")

			// Verify admin gateway security headers are present
			assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"), "Security headers should be applied")
			assert.Equal(t, "DENY", resp.Header.Get("X-Frame-Options"), "Security headers should be applied")
			assert.Equal(t, "1; mode=block", resp.Header.Get("X-XSS-Protection"), "Security headers should be applied")
		})

		// Test proxying to content API with admin prefix  
		t.Run("Proxy_to_AdminContentAPI", func(t *testing.T) {
			endpointURL := fmt.Sprintf("%s/admin/api/v1/content", adminGatewayURL)
			resp, err := makeHTTPRequest(ctx, "GET", endpointURL, nil)

			require.NoError(t, err, "Admin gateway should proxy to content API")
			assert.Equal(t, http.StatusOK, resp.StatusCode, "Admin content API request via gateway should succeed")

			// Verify admin gateway security headers are present
			assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"), "Security headers should be applied")
			assert.Equal(t, "DENY", resp.Header.Get("X-Frame-Options"), "Security headers should be applied")
			assert.Equal(t, "1; mode=block", resp.Header.Get("X-XSS-Protection"), "Security headers should be applied")
		})
	})
}

// TestDaprServiceInvocation validates Dapr service invocation is working correctly
func TestDaprServiceInvocation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
	defer cancel()

	// Test direct Dapr service invocation to content API
	t.Run("Direct_ContentAPI_Invocation", func(t *testing.T) {
		// Arrange
		daprHTTPPort := getRequiredEnvVar(t, "DAPR_HTTP_PORT")
		serviceHost := getRequiredEnvVar(t, "SERVICE_HOST")
		
		// Act - Invoke content API directly via Dapr
		daprURL := fmt.Sprintf("http://%s:%s/v1.0/invoke/content-api/method/api/v1/content", serviceHost, daprHTTPPort)
		resp, err := makeHTTPRequest(ctx, "GET", daprURL, nil)

		// Assert - Dapr service invocation should work
		require.NoError(t, err, "Dapr service invocation to content API should work")
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Content API should respond via Dapr service invocation")

		// Verify Dapr specific headers
		assert.NotEmpty(t, resp.Header.Get("Dapr-App-Id"), "Dapr app ID header should be present")
	})

	// Test direct Dapr service invocation to services API
	t.Run("Direct_ServicesAPI_Invocation", func(t *testing.T) {
		// Arrange
		daprHTTPPort := getRequiredEnvVar(t, "DAPR_HTTP_PORT")
		serviceHost := getRequiredEnvVar(t, "SERVICE_HOST")
		
		// Act - Invoke services API directly via Dapr
		daprURL := fmt.Sprintf("http://%s:%s/v1.0/invoke/services-api/method/api/v1/services", serviceHost, daprHTTPPort)
		resp, err := makeHTTPRequest(ctx, "GET", daprURL, nil)

		// Assert - Dapr service invocation should work
		require.NoError(t, err, "Dapr service invocation to services API should work")
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Services API should respond via Dapr service invocation")

		// Verify Dapr specific headers
		assert.NotEmpty(t, resp.Header.Get("Dapr-App-Id"), "Dapr app ID header should be present")
	})
}

// TestHealthEndpoints validates health endpoints are accessible via service communication
func TestHealthEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
	defer cancel()

	// Test health endpoints for all services
	services := map[string]string{
		"Content API":      getRequiredEnvVar(t, "CONTENT_API_URL"),
		"Services API":     getRequiredEnvVar(t, "SERVICES_API_URL"),
		"Public Gateway":   getRequiredEnvVar(t, "PUBLIC_GATEWAY_URL"),
		"Admin Gateway":    getRequiredEnvVar(t, "ADMIN_GATEWAY_URL"),
	}

	for serviceName, serviceURL := range services {
		t.Run(fmt.Sprintf("%s_Health", serviceName), func(t *testing.T) {
			// Act - Check health endpoint
			healthURL := fmt.Sprintf("%s/health", serviceURL)
			resp, err := makeHTTPRequest(ctx, "GET", healthURL, nil)

			// Assert - Health endpoint should be accessible and healthy
			require.NoError(t, err, "%s health endpoint should be accessible", serviceName)
			assert.Equal(t, http.StatusOK, resp.StatusCode, "%s should be healthy", serviceName)

			// Verify health response format
			var healthData map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&healthData)
			require.NoError(t, err, "Health response should be valid JSON")
			
			assert.Equal(t, "ok", healthData["status"], "Health status should be 'ok'")
			assert.Contains(t, healthData, "timestamp", "Health response should include timestamp")
		})
	}
}

// makeHTTPRequest creates HTTP request with proper timeout and context
func makeHTTPRequest(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add required headers for API requests
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "development-integration-tests/1.0")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	return resp, nil
}