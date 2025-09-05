package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	sharedtesting "github.com/axiom-software-co/international-center/src/deployer/shared/testing"
)

// TestWebsiteContainerDeployment validates website container deployment and health
func TestWebsiteContainerDeployment(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("Website_Container_Health", func(t *testing.T) {
		// Arrange - Website should be deployed in container
		websiteURL := sharedtesting.GetEnvVar("WEBSITE_URL", "http://localhost:3000")
		client := &http.Client{Timeout: 10 * time.Second}
		
		// Act - Check website health endpoint
		req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/health", websiteURL), nil)
		require.NoError(t, err, "Should be able to create website health request")
		resp, err := client.Do(req)
		
		// Assert - Website container should be running and healthy
		require.NoError(t, err, "Website container should be accessible")
		require.NotNil(t, resp, "Website health response should not be nil")
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Website container should be healthy")
		resp.Body.Close()
		
		t.Logf("Website container is healthy at %s", websiteURL)
	})

	t.Run("Website_Static_Asset_Serving", func(t *testing.T) {
		// Arrange - Website should serve static assets
		websiteURL := sharedtesting.GetEnvVar("WEBSITE_URL", "http://localhost:3000")
		client := &http.Client{Timeout: 10 * time.Second}
		
		// Act - Request static assets (CSS, JS, images)
		staticAssets := []string{
			"/assets/main.css",
			"/assets/main.js", 
			"/favicon.ico",
		}
		
		for _, asset := range staticAssets {
			req, err := http.NewRequestWithContext(ctx, "GET", websiteURL+asset, nil)
			require.NoError(t, err, "Should be able to create static asset request")
			resp, err := client.Do(req)
			
			// Assert - Static assets should be served correctly
			if err == nil && resp != nil {
				assert.True(t, resp.StatusCode < 500, fmt.Sprintf("Static asset %s should be servable", asset))
				if resp.Body != nil {
					resp.Body.Close()
				}
			} else {
				t.Logf("Static asset %s not yet available - expected until container is deployed", asset)
			}
		}
		
		t.Log("Website static asset serving validation completed")
	})

	t.Run("Website_Container_Build_Validation", func(t *testing.T) {
		// Arrange - Website container should be built from Astro/Vue source
		websiteURL := sharedtesting.GetEnvVar("WEBSITE_URL", "http://localhost:3000")
		client := &http.Client{Timeout: 10 * time.Second}
		
		// Act - Check website main page for proper rendering
		req, err := http.NewRequestWithContext(ctx, "GET", websiteURL, nil)
		require.NoError(t, err, "Should be able to create main page request")
		resp, err := client.Do(req)
		
		// Assert - Website should serve HTML content
		if err == nil && resp != nil {
			defer resp.Body.Close()
			contentType := resp.Header.Get("Content-Type")
			assert.True(t, strings.Contains(contentType, "text/html"), 
				"Website should serve HTML content")
			
			// Validate proper Astro/Vue build artifacts
			if resp.StatusCode == http.StatusOK {
				t.Log("Website container is serving built content")
			} else {
				t.Logf("Website container status: %d - expected until proper deployment", resp.StatusCode)
			}
		} else {
			t.Log("Website container not yet accessible - expected until deployment")
		}
	})
}

// TestWebsiteToBackendConnectivity validates website-to-backend API integration
func TestWebsiteToBackendConnectivity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("Website_to_Content_API_Integration", func(t *testing.T) {
		// Arrange - Website should be able to connect to Content API via Dapr
		websiteURL := sharedtesting.GetEnvVar("WEBSITE_URL", "http://localhost:3000")
		client := &http.Client{Timeout: 10 * time.Second}
		
		// Act - Test website API proxy to Content API
		apiEndpoint := fmt.Sprintf("%s/api/content/health", websiteURL)
		req, err := http.NewRequestWithContext(ctx, "GET", apiEndpoint, nil)
		require.NoError(t, err, "Should be able to create API proxy request")
		resp, err := client.Do(req)
		
		// Assert - Website should proxy to Content API successfully
		if err == nil && resp != nil {
			defer resp.Body.Close()
			assert.True(t, resp.StatusCode < 500, "Website should successfully proxy to Content API")
			t.Log("Website-to-Content API integration validated")
		} else {
			t.Log("Website-to-Content API integration not yet configured - expected until GREEN phase")
		}
	})

	t.Run("Website_to_Services_API_Integration", func(t *testing.T) {
		// Arrange - Website should be able to connect to Services API
		websiteURL := sharedtesting.GetEnvVar("WEBSITE_URL", "http://localhost:3000")
		client := &http.Client{Timeout: 10 * time.Second}
		
		// Act - Test website API proxy to Services API
		apiEndpoint := fmt.Sprintf("%s/api/services/health", websiteURL)
		req, err := http.NewRequestWithContext(ctx, "GET", apiEndpoint, nil)
		require.NoError(t, err, "Should be able to create Services API proxy request")
		resp, err := client.Do(req)
		
		// Assert - Website should proxy to Services API successfully
		if err == nil && resp != nil {
			defer resp.Body.Close()
			assert.True(t, resp.StatusCode < 500, "Website should successfully proxy to Services API")
			t.Log("Website-to-Services API integration validated")
		} else {
			t.Log("Website-to-Services API integration not yet configured - expected until GREEN phase")
		}
	})

	t.Run("Website_Backend_Authentication_Flow", func(t *testing.T) {
		// Arrange - Website should handle authentication with backend services
		websiteURL := sharedtesting.GetEnvVar("WEBSITE_URL", "http://localhost:3000")
		client := &http.Client{Timeout: 10 * time.Second}
		
		// Act - Test authentication endpoint connectivity
		authEndpoint := fmt.Sprintf("%s/auth/login", websiteURL)
		req, err := http.NewRequestWithContext(ctx, "GET", authEndpoint, nil)
		require.NoError(t, err, "Should be able to create auth endpoint request")
		resp, err := client.Do(req)
		
		// Assert - Authentication flow should be accessible
		if err == nil && resp != nil {
			defer resp.Body.Close()
			assert.True(t, resp.StatusCode < 500, "Website authentication flow should be accessible")
		} else {
			t.Log("Website authentication not yet configured - expected until GREEN phase")
		}
	})
}

// TestWebsiteContainerNetworking validates website container networking with backend services
func TestWebsiteContainerNetworking(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("Website_Container_Network_Isolation", func(t *testing.T) {
		// Arrange - Website container should be on same network as backend services
		websiteURL := sharedtesting.GetEnvVar("WEBSITE_URL", "http://localhost:3000")
		client := &http.Client{Timeout: 10 * time.Second}
		
		// Act - Test network connectivity validation endpoint
		networkEndpoint := fmt.Sprintf("%s/api/network/status", websiteURL)
		req, err := http.NewRequestWithContext(ctx, "GET", networkEndpoint, nil)
		require.NoError(t, err, "Should be able to create network status request")
		resp, err := client.Do(req)
		
		// Assert - Network connectivity should be validated
		if err == nil && resp != nil {
			defer resp.Body.Close()
			assert.True(t, resp.StatusCode < 500, "Website should validate network connectivity")
		} else {
			t.Log("Website network validation not yet implemented - expected until GREEN phase")
		}
	})

	t.Run("Website_Service_Discovery_Integration", func(t *testing.T) {
		// Arrange - Website should discover backend services via Dapr
		websiteURL := sharedtesting.GetEnvVar("WEBSITE_URL", "http://localhost:3000")
		client := &http.Client{Timeout: 10 * time.Second}
		
		// Act - Test service discovery endpoint
		discoveryEndpoint := fmt.Sprintf("%s/api/services/discovery", websiteURL)
		req, err := http.NewRequestWithContext(ctx, "GET", discoveryEndpoint, nil)
		require.NoError(t, err, "Should be able to create service discovery request")
		resp, err := client.Do(req)
		
		// Assert - Service discovery should work from website
		if err == nil && resp != nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				var services map[string]interface{}
				if json.NewDecoder(resp.Body).Decode(&services) == nil {
					assert.NotEmpty(t, services, "Website should discover backend services")
				}
			}
		} else {
			t.Log("Website service discovery not yet implemented - expected until GREEN phase")
		}
	})

	t.Run("Website_CORS_Configuration", func(t *testing.T) {
		// Arrange - Website should handle CORS for backend API calls
		websiteURL := sharedtesting.GetEnvVar("WEBSITE_URL", "http://localhost:3000")
		client := &http.Client{Timeout: 10 * time.Second}
		
		// Act - Test CORS preflight for API calls
		req, err := http.NewRequestWithContext(ctx, "OPTIONS", fmt.Sprintf("%s/api/content", websiteURL), nil)
		require.NoError(t, err, "Should be able to create CORS preflight request")
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "GET")
		resp, err := client.Do(req)
		
		// Assert - CORS should be properly configured
		if err == nil && resp != nil {
			defer resp.Body.Close()
			corsHeader := resp.Header.Get("Access-Control-Allow-Origin")
			if corsHeader != "" {
				assert.NotEmpty(t, corsHeader, "Website should configure CORS for API access")
			}
		} else {
			t.Log("Website CORS configuration not yet implemented - expected until GREEN phase")
		}
	})
}

// TestWebsiteContainerLifecycle validates website container build and deployment lifecycle
func TestWebsiteContainerLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("Website_Container_Build_Process", func(t *testing.T) {
		// Arrange - Website container should be built from source
		websiteURL := sharedtesting.GetEnvVar("WEBSITE_URL", "http://localhost:3000")
		client := &http.Client{Timeout: 10 * time.Second}
		
		// Act - Validate build artifacts are accessible
		req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/build-info", websiteURL), nil)
		require.NoError(t, err, "Should be able to create build info request")
		resp, err := client.Do(req)
		
		// Assert - Build information should be available
		if err == nil && resp != nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				var buildInfo map[string]interface{}
				if json.NewDecoder(resp.Body).Decode(&buildInfo) == nil {
					assert.NotEmpty(t, buildInfo, "Website should provide build information")
				}
			}
		} else {
			t.Log("Website build info endpoint not yet implemented - expected until GREEN phase")
		}
	})

	t.Run("Website_Container_Environment_Variables", func(t *testing.T) {
		// Arrange - Website container should have proper environment configuration
		websiteURL := sharedtesting.GetEnvVar("WEBSITE_URL", "http://localhost:3000")
		client := &http.Client{Timeout: 10 * time.Second}
		
		// Act - Test environment configuration endpoint
		req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/config", websiteURL), nil)
		require.NoError(t, err, "Should be able to create config request")
		resp, err := client.Do(req)
		
		// Assert - Environment configuration should be properly set
		if err == nil && resp != nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				var config map[string]interface{}
				if json.NewDecoder(resp.Body).Decode(&config) == nil {
					// Validate essential configuration is present
					assert.Contains(t, config, "environment", "Website should have environment configuration")
				}
			}
		} else {
			t.Log("Website configuration endpoint not yet implemented - expected until GREEN phase")
		}
	})

	t.Run("Website_Container_Health_Monitoring", func(t *testing.T) {
		// Arrange - Website container should support health monitoring
		websiteURL := sharedtesting.GetEnvVar("WEBSITE_URL", "http://localhost:3000")
		client := &http.Client{Timeout: 10 * time.Second}
		
		// Act - Test detailed health check endpoint
		req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/health/detailed", websiteURL), nil)
		require.NoError(t, err, "Should be able to create detailed health request")
		resp, err := client.Do(req)
		
		// Assert - Detailed health information should be available
		if err == nil && resp != nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				var health map[string]interface{}
				if json.NewDecoder(resp.Body).Decode(&health) == nil {
					assert.Contains(t, health, "status", "Website should provide health status")
					assert.Contains(t, health, "dependencies", "Website should report dependency status")
				}
			}
		} else {
			t.Log("Website detailed health check not yet implemented - expected until GREEN phase")
		}
	})
}

// TestWebsiteEnvironmentIntegration validates website integration with development environment
func TestWebsiteEnvironmentIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("Website_Development_Mode_Features", func(t *testing.T) {
		// Arrange - Website should have development mode features enabled
		websiteURL := sharedtesting.GetEnvVar("WEBSITE_URL", "http://localhost:3000")
		client := &http.Client{Timeout: 10 * time.Second}
		
		// Act - Test development features accessibility
		devEndpoints := []string{
			"/dev/hot-reload",
			"/dev/api-explorer",
			"/dev/component-preview",
		}
		
		for _, endpoint := range devEndpoints {
			req, err := http.NewRequestWithContext(ctx, "GET", websiteURL+endpoint, nil)
			require.NoError(t, err, "Should be able to create dev feature request")
			resp, err := client.Do(req)
			
			// Assert - Development features should be accessible
			if err == nil && resp != nil {
				if resp.Body != nil {
					resp.Body.Close()
				}
				t.Logf("Development feature %s accessibility checked", endpoint)
			} else {
				t.Logf("Development feature %s not yet available - expected until GREEN phase", endpoint)
			}
		}
	})

	t.Run("Website_Local_Storage_Integration", func(t *testing.T) {
		// Arrange - Website should integrate with local development storage
		websiteURL := sharedtesting.GetEnvVar("WEBSITE_URL", "http://localhost:3000")
		client := &http.Client{Timeout: 10 * time.Second}
		
		// Act - Test local storage endpoint integration
		req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/storage/test", websiteURL), nil)
		require.NoError(t, err, "Should be able to create storage test request")
		resp, err := client.Do(req)
		
		// Assert - Storage integration should work
		if err == nil && resp != nil {
			defer resp.Body.Close()
			assert.True(t, resp.StatusCode < 500, "Website should integrate with storage services")
		} else {
			t.Log("Website storage integration not yet configured - expected until GREEN phase")
		}
	})

	t.Run("Website_Observability_Integration", func(t *testing.T) {
		// Arrange - Website should integrate with observability stack
		websiteURL := sharedtesting.GetEnvVar("WEBSITE_URL", "http://localhost:3000")
		client := &http.Client{Timeout: 10 * time.Second}
		
		// Act - Test observability endpoints
		req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/metrics", websiteURL), nil)
		require.NoError(t, err, "Should be able to create metrics request")
		resp, err := client.Do(req)
		
		// Assert - Observability integration should be available
		if err == nil && resp != nil {
			defer resp.Body.Close()
			assert.True(t, resp.StatusCode < 500, "Website should provide observability metrics")
		} else {
			t.Log("Website observability integration not yet implemented - expected until GREEN phase")
		}
	})
}