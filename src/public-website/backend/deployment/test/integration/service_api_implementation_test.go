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

// RED PHASE: Service API Implementation Tests
// These tests validate that service APIs implement actual endpoints beyond health checks

func TestServiceAPIImplementation_ActualEndpointsAvailable(t *testing.T) {
	// This test requires complete environment health - enforcing axiom rule
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Expected API endpoints that should be implemented in each service
	serviceAPIEndpoints := []struct {
		serviceName    string
		directEndpoint string
		meshEndpoint   string
		method         string
		expectedStatus int
		description    string
		critical       bool
	}{
		{
			serviceName:    "content",
			directEndpoint: "http://localhost:3001/api/news",
			meshEndpoint:   "http://localhost:3500/v1.0/invoke/content/method/api/news",
			method:         "GET",
			expectedStatus: 200,
			description:    "Content service must implement news API endpoint",
			critical:       true,
		},
		{
			serviceName:    "content",
			directEndpoint: "http://localhost:3001/api/events",
			meshEndpoint:   "http://localhost:3500/v1.0/invoke/content/method/api/events",
			method:         "GET",
			expectedStatus: 200,
			description:    "Content service must implement events API endpoint",
			critical:       true,
		},
		{
			serviceName:    "inquiries",
			directEndpoint: "http://localhost:3101/api/inquiries",
			meshEndpoint:   "http://localhost:3500/v1.0/invoke/inquiries/method/api/inquiries",
			method:         "GET",
			expectedStatus: 200,
			description:    "Inquiries service must implement inquiries API endpoint",
			critical:       true,
		},
		{
			serviceName:    "notifications",
			directEndpoint: "http://localhost:3201/api/subscribers",
			meshEndpoint:   "http://localhost:3500/v1.0/invoke/notifications/method/api/subscribers",
			method:         "GET",
			expectedStatus: 200,
			description:    "Notifications service must implement subscribers API endpoint",
			critical:       true,
		},
		{
			serviceName:    "public-gateway",
			directEndpoint: "http://localhost:9001/api/news",
			meshEndpoint:   "http://localhost:3500/v1.0/invoke/public-gateway/method/api/news",
			method:         "GET",
			expectedStatus: 200,
			description:    "Public gateway must route to news API endpoint",
			critical:       false,
		},
		{
			serviceName:    "admin-gateway",
			directEndpoint: "http://localhost:9000/api/admin/content",
			meshEndpoint:   "http://localhost:3500/v1.0/invoke/admin-gateway/method/api/admin/content",
			method:         "GET",
			expectedStatus: 200,
			description:    "Admin gateway must provide content management API",
			critical:       false,
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Act & Assert: Test API endpoint implementation
	for _, endpoint := range serviceAPIEndpoints {
		t.Run("APIEndpoint_"+endpoint.serviceName+"_"+strings.ReplaceAll(endpoint.directEndpoint[strings.LastIndex(endpoint.directEndpoint, "/"):], "/", ""), func(t *testing.T) {
			// Test direct API endpoint
			directReq, err := http.NewRequestWithContext(ctx, endpoint.method, endpoint.directEndpoint, nil)
			require.NoError(t, err, "Failed to create direct API request")

			directResp, err := client.Do(directReq)
			if err == nil {
				defer directResp.Body.Close()
				
				if endpoint.critical {
					// Critical endpoints must not return 404
					assert.NotEqual(t, 404, directResp.StatusCode, 
						"%s - critical API endpoint must be implemented (not 404)", endpoint.description)
					assert.True(t, directResp.StatusCode >= 200 && directResp.StatusCode < 500,
						"%s - API endpoint must be accessible", endpoint.description)
				} else {
					// Non-critical endpoints should be implemented for complete functionality
					if directResp.StatusCode == 404 {
						t.Logf("%s not implemented yet (expected for incomplete API layer)", endpoint.description)
					} else {
						assert.True(t, directResp.StatusCode >= 200 && directResp.StatusCode < 500,
							"%s - API endpoint should be functional", endpoint.description)
					}
				}

				// Validate response is not generic 404
				body, err := io.ReadAll(directResp.Body)
				if err == nil {
					responseStr := string(body)
					
					if endpoint.critical {
						assert.NotContains(t, responseStr, "404 page not found",
							"%s - critical endpoint must have proper implementation", endpoint.description)
					}

					// API should return JSON data (not HTML 404 page)
					if directResp.StatusCode >= 200 && directResp.StatusCode < 300 {
						var jsonData interface{}
						assert.NoError(t, json.Unmarshal(body, &jsonData),
							"%s - API endpoint must return valid JSON", endpoint.description)
					}
				}
			} else {
				if endpoint.critical {
					t.Errorf("%s not accessible: %v", endpoint.description, err)
				} else {
					t.Logf("%s not accessible (expected for incomplete implementation): %v", endpoint.description, err)
				}
			}

			// Test service mesh API access
			meshReq, err := http.NewRequestWithContext(ctx, endpoint.method, endpoint.meshEndpoint, nil)
			require.NoError(t, err, "Failed to create service mesh API request")

			meshResp, err := client.Do(meshReq)
			if err == nil {
				defer meshResp.Body.Close()
				
				if endpoint.critical {
					assert.True(t, meshResp.StatusCode >= 200 && meshResp.StatusCode < 500,
						"%s - service mesh API access must be functional", endpoint.description)
					
					meshBody, err := io.ReadAll(meshResp.Body)
					if err == nil {
						meshResponseStr := string(meshBody)
						assert.NotContains(t, meshResponseStr, "404 page not found",
							"%s - service mesh endpoint must have proper implementation", endpoint.description)
					}
				}
			} else {
				if endpoint.critical {
					t.Errorf("%s service mesh access not functional: %v", endpoint.description, err)
				}
			}
		})
	}
}

func TestServiceAPIImplementation_DataPersistenceOperations(t *testing.T) {
	// Test that API endpoints support actual data operations with database
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Data persistence operation tests
	dataPersistenceTests := []struct {
		serviceName   string
		endpoint      string
		testData      string
		operation     string
		description   string
	}{
		{
			serviceName: "content",
			endpoint:    "http://localhost:3001/api/news",
			testData:    `{"title":"Development Test","content":"Test content","author":"Test System"}`,
			operation:   "CREATE",
			description: "Content service API must support creating news with database persistence",
		},
		{
			serviceName: "inquiries",
			endpoint:    "http://localhost:3101/api/inquiries/business",
			testData:    `{"company_name":"Test Company","contact_email":"test@example.com","message":"Test inquiry"}`,
			operation:   "CREATE",
			description: "Inquiries service API must support creating business inquiries with database persistence",
		},
		{
			serviceName: "notifications",
			endpoint:    "http://localhost:3201/api/subscribers",
			testData:    `{"email":"test@example.com","name":"Test Subscriber","event_types":["news","events"]}`,
			operation:   "CREATE",
			description: "Notifications service API must support creating subscribers with database persistence",
		},
	}

	client := &http.Client{Timeout: 15 * time.Second}

	// Act & Assert: Test data persistence operations
	for _, persistenceTest := range dataPersistenceTests {
		t.Run("DataPersistence_"+persistenceTest.serviceName+"_"+persistenceTest.operation, func(t *testing.T) {
			// Test POST operation for data creation
			createReq, err := http.NewRequestWithContext(ctx, "POST", persistenceTest.endpoint, strings.NewReader(persistenceTest.testData))
			require.NoError(t, err, "Failed to create data persistence request")
			createReq.Header.Set("Content-Type", "application/json")

			createResp, err := client.Do(createReq)
			if err == nil {
				defer createResp.Body.Close()
				
				// Should not return 404 for implemented endpoints
				assert.NotEqual(t, 404, createResp.StatusCode,
					"%s - API endpoint must be implemented for data operations", persistenceTest.description)
				
				// Should handle data operations (even if validation errors occur)
				assert.True(t, createResp.StatusCode >= 200 && createResp.StatusCode < 500,
					"%s - API must handle data persistence operations", persistenceTest.description)

				body, err := io.ReadAll(createResp.Body)
				if err == nil {
					responseStr := string(body)
					
					// Should not return generic 404 page
					assert.NotContains(t, responseStr, "404 page not found",
						"%s - API endpoint must have proper handler implementation", persistenceTest.description)
					
					// Should return structured response (JSON)
					if createResp.StatusCode >= 200 && createResp.StatusCode < 300 {
						var jsonData interface{}
						assert.NoError(t, json.Unmarshal(body, &jsonData),
							"%s - API must return JSON response for data operations", persistenceTest.description)
					}
				}
			} else {
				t.Errorf("%s not accessible for data operations: %v", persistenceTest.description, err)
			}
		})
	}
}

func TestServiceAPIImplementation_FrontendRequiredEndpoints(t *testing.T) {
	// Test that services implement endpoints required for frontend consumption
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Frontend-required API endpoints
	frontendRequiredAPIs := []struct {
		serviceName    string
		endpoint       string
		frontendUsage  string
		responseFormat string
		description    string
	}{
		{
			serviceName:   "content",
			endpoint:      "http://localhost:3001/api/news",
			frontendUsage: "website news section",
			responseFormat: "JSON array",
			description:   "Content news API required for website news section display",
		},
		{
			serviceName:   "content",
			endpoint:      "http://localhost:3001/api/events",
			frontendUsage: "website events section",
			responseFormat: "JSON array",
			description:   "Content events API required for website events section display",
		},
		{
			serviceName:   "inquiries",
			endpoint:      "http://localhost:3101/api/inquiries",
			frontendUsage: "website contact forms",
			responseFormat: "JSON object",
			description:   "Inquiries API required for website contact form submission",
		},
		{
			serviceName:   "notifications",
			endpoint:      "http://localhost:3201/api/subscribers",
			frontendUsage: "newsletter subscription",
			responseFormat: "JSON object",
			description:   "Notifications API required for newsletter subscription functionality",
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Act & Assert: Test frontend-required API endpoints
	for _, frontendAPI := range frontendRequiredAPIs {
		t.Run("FrontendAPI_"+frontendAPI.serviceName, func(t *testing.T) {
			req, err := http.NewRequestWithContext(ctx, "GET", frontendAPI.endpoint, nil)
			require.NoError(t, err, "Failed to create frontend API request")

			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				
				// Frontend APIs must be accessible (not 404)
				assert.NotEqual(t, 404, resp.StatusCode,
					"%s - frontend API must be implemented", frontendAPI.description)
				
				// Should return appropriate status for frontend consumption
				assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500,
					"%s - frontend API must be accessible", frontendAPI.description)

				body, err := io.ReadAll(resp.Body)
				if err == nil {
					responseStr := string(body)
					
					// Must not return generic 404 page
					assert.NotContains(t, responseStr, "404 page not found",
						"%s - must have actual API implementation for %s", frontendAPI.description, frontendAPI.frontendUsage)
					
					// Should return structured data for frontend consumption
					if resp.StatusCode >= 200 && resp.StatusCode < 300 {
						var jsonData interface{}
						assert.NoError(t, json.Unmarshal(body, &jsonData),
							"%s - must return %s for %s", frontendAPI.description, frontendAPI.responseFormat, frontendAPI.frontendUsage)
					}
				}
			} else {
				t.Errorf("%s not accessible for %s: %v", frontendAPI.description, frontendAPI.frontendUsage, err)
			}
		})
	}
}

func TestServiceAPIImplementation_ServiceHealthVsAPIConsistency(t *testing.T) {
	// Test consistency between service health (working) and API implementation (missing)
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Services that have health endpoints working but API endpoints missing
	serviceConsistencyTests := []struct {
		serviceName     string
		healthEndpoint  string
		apiEndpoint     string
		description     string
	}{
		{
			serviceName:    "content",
			healthEndpoint: "http://localhost:3500/v1.0/invoke/content/method/health",
			apiEndpoint:    "http://localhost:3500/v1.0/invoke/content/method/api/news",
			description:    "Content service health working but news API missing",
		},
		{
			serviceName:    "inquiries", 
			healthEndpoint: "http://localhost:3500/v1.0/invoke/inquiries/method/health",
			apiEndpoint:    "http://localhost:3500/v1.0/invoke/inquiries/method/api/inquiries",
			description:    "Inquiries service health working but inquiries API missing",
		},
		{
			serviceName:    "notifications",
			healthEndpoint: "http://localhost:3500/v1.0/invoke/notifications/method/health",
			apiEndpoint:    "http://localhost:3500/v1.0/invoke/notifications/method/api/subscribers",
			description:    "Notifications service health working but subscribers API missing",
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Act & Assert: Test health vs API consistency
	for _, consistency := range serviceConsistencyTests {
		t.Run("Consistency_"+consistency.serviceName, func(t *testing.T) {
			// Test health endpoint (should work)
			healthReq, err := http.NewRequestWithContext(ctx, "GET", consistency.healthEndpoint, nil)
			require.NoError(t, err, "Failed to create health request")

			healthResp, err := client.Do(healthReq)
			require.NoError(t, err, "Health endpoint must be accessible")
			defer healthResp.Body.Close()

			assert.True(t, healthResp.StatusCode >= 200 && healthResp.StatusCode < 300,
				"Service %s health endpoint must be working", consistency.serviceName)

			// Test API endpoint (should fail in RED phase)
			apiReq, err := http.NewRequestWithContext(ctx, "GET", consistency.apiEndpoint, nil)
			require.NoError(t, err, "Failed to create API request")

			apiResp, err := client.Do(apiReq)
			if err == nil {
				defer apiResp.Body.Close()
				
				body, err := io.ReadAll(apiResp.Body)
				if err == nil {
					responseStr := string(body)
					
					// In RED phase, API endpoints should return 404 or similar (not implemented)
					if strings.Contains(responseStr, "404 page not found") || apiResp.StatusCode == 404 {
						t.Logf("%s - API endpoint not implemented yet (expected in RED phase)", consistency.description)
						t.Fail() // Should fail until GREEN phase implements APIs
					} else {
						t.Logf("%s - API endpoint appears to be implemented", consistency.description)
					}
				}
			} else {
				t.Logf("%s - API endpoint not accessible: %v", consistency.description, err)
			}
		})
	}
}

// validateEnvironmentPrerequisites ensures environment health before integration testing
func validateEnvironmentPrerequisites(t *testing.T) {
	// Check critical infrastructure and platform components are running
	criticalContainers := []string{"postgresql", "dapr-control-plane", "content", "inquiries", "notifications"}
	
	for _, container := range criticalContainers {
		cmd := exec.Command("podman", "ps", "--filter", "name="+container, "--format", "{{.Names}}")
		output, err := cmd.Output()
		require.NoError(t, err, "Failed to check critical container %s", container)

		if !strings.Contains(string(output), container) {
			t.Skipf("Critical container %s not running - environment not ready for integration testing", container)
		}
	}
}