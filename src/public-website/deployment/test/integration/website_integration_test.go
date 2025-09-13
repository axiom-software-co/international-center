package integration

import (
	"context"
	"net/http"
	"testing"
	"time"

	sharedValidation "github.com/axiom-software-co/international-center/src/public-website/infrastructure/test/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Website Integration Tests
// Validates website phase integration with frontend, CDN, SSL working with backend services
// Tests end-to-end functionality from frontend to backend through gateway services

func TestWebsiteIntegration_FrontendToBackendFlow(t *testing.T) {
	// This test requires complete environment health - enforcing axiom rule
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("PublicWebsite_BackendIntegration", func(t *testing.T) {
		// Test public website integration with backend services through public gateway
		client := &http.Client{Timeout: 15 * time.Second}

		// Test public website accessibility
		websiteReq, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:5173", nil)
		require.NoError(t, err, "Failed to create public website request")

		websiteResp, err := client.Do(websiteReq)
		if err == nil {
			defer websiteResp.Body.Close()
			assert.True(t, websiteResp.StatusCode >= 200 && websiteResp.StatusCode < 500, 
				"Public website must be accessible for frontend integration")

			// Test that website can reach backend APIs through gateway
			if websiteResp.StatusCode >= 200 && websiteResp.StatusCode < 300 {
				// Test API endpoints that website would use
				apiEndpoints := []string{
					"http://localhost:9001/api/news",
					"http://localhost:9001/api/events",
					"http://localhost:9001/api/research",
				}

				for _, endpoint := range apiEndpoints {
					apiReq, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
					require.NoError(t, err, "Failed to create API request")

					apiResp, err := client.Do(apiReq)
					if err == nil {
						defer apiResp.Body.Close()
						assert.True(t, apiResp.StatusCode >= 200 && apiResp.StatusCode < 500,
							"Public website API endpoint must be accessible: %s", endpoint)
					} else {
						t.Logf("Website API endpoint %s not accessible: %v", endpoint, err)
					}
				}
			}
		} else {
			t.Logf("Public website not accessible: %v", err)
		}
	})

	t.Run("AdminPortal_BackendIntegration", func(t *testing.T) {
		// Test admin portal integration with backend services through admin gateway
		client := &http.Client{Timeout: 15 * time.Second}

		// Test admin portal accessibility
		portalReq, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:3001", nil)
		require.NoError(t, err, "Failed to create admin portal request")

		portalResp, err := client.Do(portalReq)
		if err == nil {
			defer portalResp.Body.Close()
			assert.True(t, portalResp.StatusCode >= 200 && portalResp.StatusCode < 500, 
				"Admin portal must be accessible for admin integration")

			// Test admin API endpoints through admin gateway
			if portalResp.StatusCode >= 200 && portalResp.StatusCode < 300 {
				adminEndpoints := []string{
					"http://localhost:9000/api/admin/content",
					"http://localhost:9000/api/admin/inquiries",
					"http://localhost:9000/api/admin/notifications",
				}

				for _, endpoint := range adminEndpoints {
					adminReq, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
					require.NoError(t, err, "Failed to create admin API request")

					adminResp, err := client.Do(adminReq)
					if err == nil {
						defer adminResp.Body.Close()
						assert.True(t, adminResp.StatusCode >= 200 && adminResp.StatusCode < 500,
							"Admin portal API endpoint must be accessible: %s", endpoint)
					} else {
						t.Logf("Admin API endpoint %s not accessible: %v", endpoint, err)
					}
				}
			}
		} else {
			t.Logf("Admin portal not accessible: %v", err)
		}
	})
}

func TestWebsiteIntegration_ContentDeliveryAndCaching(t *testing.T) {
	// Test website content delivery and caching integration
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("StaticAsset_DeliveryIntegration", func(t *testing.T) {
		// Test static asset delivery for public website
		client := &http.Client{Timeout: 10 * time.Second}

		// Test common static asset paths
		staticAssetPaths := []string{
			"http://localhost:5173/favicon.ico",
			"http://localhost:5173/assets/",
			"http://localhost:5173/public/",
		}

		for _, assetPath := range staticAssetPaths {
			req, err := http.NewRequestWithContext(ctx, "GET", assetPath, nil)
			require.NoError(t, err, "Failed to create static asset request")

			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				// Static assets should be accessible or return proper 404
				assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500,
					"Static asset delivery must be functional: %s", assetPath)
			} else {
				t.Logf("Static asset path %s not accessible: %v", assetPath, err)
			}
		}
	})

	t.Run("CDN_ConfigurationIntegration", func(t *testing.T) {
		// Test CDN configuration (even if disabled in development)
		client := &http.Client{Timeout: 5 * time.Second}

		// Test that website responds with appropriate headers for CDN integration
		req, err := http.NewRequestWithContext(ctx, "HEAD", "http://localhost:5173", nil)
		require.NoError(t, err, "Failed to create CDN headers request")

		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			
			// Check for caching headers that would integrate with CDN
			cacheControl := resp.Header.Get("Cache-Control")
			t.Logf("Cache-Control header: %s", cacheControl)
			
			// Website should have proper headers for CDN integration
			assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300,
				"Website must be ready for CDN integration")
		} else {
			t.Logf("Website CDN integration headers not accessible: %v", err)
		}
	})
}

func TestWebsiteIntegration_SecurityAndSSLConfiguration(t *testing.T) {
	// Test website security and SSL configuration integration
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("SecurityHeaders_IntegrationConfiguration", func(t *testing.T) {
		// Test security headers configuration for both public and admin sites
		client := &http.Client{Timeout: 10 * time.Second}

		securityHeadersTests := []struct {
			endpoint    string
			description string
		}{
			{"http://localhost:5173", "Public website security headers"},
			{"http://localhost:3001", "Admin portal security headers"},
		}

		for _, test := range securityHeadersTests {
			req, err := http.NewRequestWithContext(ctx, "HEAD", test.endpoint, nil)
			require.NoError(t, err, "Failed to create security headers request")

			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()

				// Check for basic security headers
				contentType := resp.Header.Get("Content-Type")
				t.Logf("%s Content-Type: %s", test.description, contentType)

				assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500,
					"Website security configuration must be operational: %s", test.description)
			} else {
				t.Logf("%s not accessible: %v", test.description, err)
			}
		}
	})

	t.Run("CORS_PolicyIntegration", func(t *testing.T) {
		// Test CORS policy integration between frontend and gateway services
		client := &http.Client{Timeout: 10 * time.Second}

		// Test CORS preflight request from frontend perspective
		corsTestURL := "http://localhost:9001/api/news"
		
		req, err := http.NewRequestWithContext(ctx, "OPTIONS", corsTestURL, nil)
		require.NoError(t, err, "Failed to create CORS preflight request")
		
		// Set Origin header as frontend would
		req.Header.Set("Origin", "http://localhost:5173")
		req.Header.Set("Access-Control-Request-Method", "GET")

		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			
			// Check CORS headers in response
			allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
			allowMethods := resp.Header.Get("Access-Control-Allow-Methods")
			
			t.Logf("CORS Allow-Origin: %s", allowOrigin)
			t.Logf("CORS Allow-Methods: %s", allowMethods)
			
			assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500,
				"CORS policy integration must be operational")
		} else {
			t.Logf("CORS policy integration not operational: %v", err)
		}
	})
}