// RED PHASE: TypeScript client integration tests - these tests should FAIL initially
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestTypeScriptClientBackendIntegration validates generated TypeScript clients work with backend
func TestTypeScriptClientBackendIntegration(t *testing.T) {
	timeout := 15 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	t.Run("Generated TypeScript clients should successfully call backend endpoints", func(t *testing.T) {
		// Test that TypeScript client endpoints match actual backend implementations
		
		clientEndpointTests := []struct {
			clientMethod   string
			backendPath    string
			gateway        string
			expectedFields []string
		}{
			{
				clientMethod:   "newsApi.getNews()",
				backendPath:    "/api/news",
				gateway:        "http://localhost:9001",
				expectedFields: []string{"data", "pagination"},
			},
			{
				clientMethod:   "newsApi.getFeaturedNews()",
				backendPath:    "/api/news/featured",
				gateway:        "http://localhost:9001",
				expectedFields: []string{"data"},
			},
			{
				clientMethod:   "servicesApi.getServices()",
				backendPath:    "/api/services",
				gateway:        "http://localhost:9001",
				expectedFields: []string{"data", "pagination"},
			},
			{
				clientMethod:   "servicesApi.getFeaturedServices()",
				backendPath:    "/api/services/featured",
				gateway:        "http://localhost:9001",
				expectedFields: []string{"data"},
			},
			{
				clientMethod:   "researchApi.getResearch()",
				backendPath:    "/api/research",
				gateway:        "http://localhost:9001",
				expectedFields: []string{"data", "pagination"},
			},
			{
				clientMethod:   "eventsApi.getEvents()",
				backendPath:    "/api/events",
				gateway:        "http://localhost:9001",
				expectedFields: []string{"data", "pagination"},
			},
		}

		for _, test := range clientEndpointTests {
			t.Run(test.clientMethod+" should match backend "+test.backendPath, func(t *testing.T) {
				// Test backend endpoint that TypeScript client calls
				client := &http.Client{Timeout: 5 * time.Second}
				url := test.gateway + test.backendPath
				
				req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
				require.NoError(t, err, "Should create client endpoint request")
				
				resp, err := client.Do(req)
				if err != nil {
					t.Errorf("❌ FAIL: Backend endpoint for %s not accessible: %v", test.clientMethod, err)
					t.Log("    TypeScript clients expect backend endpoints to be accessible")
					return
				}
				defer resp.Body.Close()
				
				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					// Parse response to validate structure matches TypeScript client expectations
					var response map[string]interface{}
					if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
						t.Errorf("❌ FAIL: Backend response not valid JSON for %s", test.clientMethod)
						return
					}
					
					// Validate response has fields that TypeScript client expects
					for _, field := range test.expectedFields {
						if _, exists := response[field]; !exists {
							t.Errorf("❌ FAIL: TypeScript client %s expects '%s' field, missing from backend response", test.clientMethod, field)
						} else {
							t.Logf("✅ Backend provides '%s' field for %s", field, test.clientMethod)
						}
					}
					
					// Validate response structure for TypeScript client consumption
					if dataField, hasData := response["data"]; hasData {
						if dataArray, ok := dataField.([]interface{}); ok {
							if len(dataArray) > 0 {
								// Check first item structure for TypeScript type compatibility
								firstItem, ok := dataArray[0].(map[string]interface{})
								if !ok {
									t.Errorf("❌ FAIL: Backend data structure not compatible with TypeScript client for %s", test.clientMethod)
								} else {
									t.Logf("✅ Backend data structure compatible with TypeScript client for %s", test.clientMethod)
									t.Logf("Sample fields: %v", getMapKeys(firstItem))
								}
							} else {
								t.Logf("⚠️  Backend returns empty data for %s", test.clientMethod)
							}
						} else {
							t.Errorf("❌ FAIL: Backend data field not array for %s", test.clientMethod)
						}
					}
				} else if resp.StatusCode == http.StatusNotFound {
					t.Errorf("❌ FAIL: Backend endpoint for %s not found (404)", test.clientMethod)
					t.Log("    TypeScript clients generated for non-existent backend endpoints")
				} else {
					t.Errorf("❌ FAIL: Backend endpoint for %s failed: status %d", test.clientMethod, resp.StatusCode)
				}
			})
		}
	})

	t.Run("TypeScript client POST operations should match backend implementations", func(t *testing.T) {
		// Test that TypeScript client POST operations work with backend
		
		postClientTests := []struct {
			clientMethod   string
			backendPath    string
			gateway        string
			requiresAuth   bool
		}{
			{
				clientMethod: "inquiriesApi.submitMediaInquiry()",
				backendPath:  "/api/inquiries/media",
				gateway:      "http://localhost:9001",
				requiresAuth: false,
			},
			{
				clientMethod: "inquiriesApi.submitBusinessInquiry()",
				backendPath:  "/api/inquiries/business",
				gateway:      "http://localhost:9001",
				requiresAuth: false,
			},
		}

		for _, test := range postClientTests {
			t.Run(test.clientMethod+" should work with backend "+test.backendPath, func(t *testing.T) {
				// Test POST endpoint that TypeScript client uses
				client := &http.Client{Timeout: 5 * time.Second}
				url := test.gateway + test.backendPath
				
				// Create minimal valid request body for testing
				requestData := map[string]interface{}{
					"first_name": "Test",
					"last_name":  "User",
					"email":      "test@example.com",
					"phone":      "123-456-7890",
					"message":    "Test inquiry message",
				}
				
				requestJSON, err := json.Marshal(requestData)
				require.NoError(t, err, "Should marshal request data")
				
				req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(requestJSON))
				require.NoError(t, err, "Should create POST request")
				req.Header.Set("Content-Type", "application/json")
				
				resp, err := client.Do(req)
				if err != nil {
					t.Errorf("❌ FAIL: Backend POST endpoint for %s not accessible: %v", test.clientMethod, err)
					return
				}
				defer resp.Body.Close()
				
				// Validate response for TypeScript client compatibility
				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					// Successful POST should return appropriate response
					var response map[string]interface{}
					if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
						t.Logf("✅ Backend POST endpoint for %s returns JSON response", test.clientMethod)
						
						// TypeScript clients expect specific response structure
						if _, hasData := response["data"]; hasData {
							t.Logf("✅ Backend POST response has 'data' field for %s", test.clientMethod)
						} else {
							t.Errorf("❌ FAIL: Backend POST response missing 'data' field for %s", test.clientMethod)
						}
					} else {
						t.Errorf("❌ FAIL: Backend POST response not valid JSON for %s", test.clientMethod)
					}
				} else if resp.StatusCode == http.StatusBadRequest {
					t.Logf("⚠️  Backend validation working for %s (400 - validation error)", test.clientMethod)
				} else if resp.StatusCode == http.StatusNotFound {
					t.Errorf("❌ FAIL: Backend POST endpoint for %s not found (404)", test.clientMethod)
				} else {
					t.Errorf("❌ FAIL: Backend POST endpoint for %s failed: status %d", test.clientMethod, resp.StatusCode)
				}
			})
		}
	})

	t.Run("TypeScript client error handling should match backend error responses", func(t *testing.T) {
		// Test that backend error responses are compatible with TypeScript client error handling
		
		errorTests := []struct {
			description    string
			endpoint       string
			gateway        string
			method         string
			expectedStatus int
			invalidData    map[string]interface{}
		}{
			{
				description:    "Invalid inquiry submission should return TypeScript-compatible error",
				endpoint:       "/api/inquiries/media",
				gateway:        "http://localhost:9001",
				method:         "POST",
				expectedStatus: 400,
				invalidData:    map[string]interface{}{"invalid": "data"},
			},
		}

		for _, test := range errorTests {
			t.Run(test.description, func(t *testing.T) {
				client := &http.Client{Timeout: 5 * time.Second}
				url := test.gateway + test.endpoint
				
				// Send invalid request to test error handling
				invalidJSON, err := json.Marshal(test.invalidData)
				require.NoError(t, err, "Should marshal invalid data")
				
				req, err := http.NewRequestWithContext(ctx, test.method, url, bytes.NewReader(invalidJSON))
				require.NoError(t, err, "Should create error test request")
				req.Header.Set("Content-Type", "application/json")
				
				resp, err := client.Do(req)
				if err != nil {
					t.Errorf("❌ FAIL: Backend error endpoint not accessible: %v", err)
					return
				}
				defer resp.Body.Close()
				
				// Validate error response structure for TypeScript client compatibility
				if resp.StatusCode >= 400 && resp.StatusCode < 500 {
					var errorResponse map[string]interface{}
					if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
						// TypeScript clients expect error responses to have specific structure
						if _, hasError := errorResponse["error"]; hasError {
							t.Logf("✅ Backend error response compatible with TypeScript client")
						} else {
							t.Errorf("❌ FAIL: Backend error response missing 'error' field for TypeScript client")
						}
					} else {
						t.Errorf("❌ FAIL: Backend error response not valid JSON for TypeScript client")
					}
				} else {
					t.Errorf("❌ FAIL: Backend should return validation error for invalid data, got status %d", resp.StatusCode)
				}
			})
		}
	})
}

// Helper function to get map keys
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// TestFrontendContractClientIntegration validates frontend contract clients work with backend
func TestFrontendContractClientIntegration(t *testing.T) {
	timeout := 15 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	t.Run("Frontend applications should be able to consume backend via contract clients", func(t *testing.T) {
		// Test that frontend applications can successfully consume backend APIs
		
		frontendIntegrationTests := []struct {
			frontend        string
			frontendURL     string
			backendEndpoint string
			gateway         string
		}{
			{
				frontend:        "public-website",
				frontendURL:     "http://localhost:3000",
				backendEndpoint: "/api/news",
				gateway:         "http://localhost:9001",
			},
			{
				frontend:        "admin-portal",
				frontendURL:     "http://localhost:3002",
				backendEndpoint: "/api/admin/news",
				gateway:         "http://localhost:9000",
			},
		}

		for _, test := range frontendIntegrationTests {
			t.Run(test.frontend+" should successfully consume backend via contract clients", func(t *testing.T) {
				client := &http.Client{Timeout: 5 * time.Second}
				
				// Test frontend application is accessible
				frontendReq, err := http.NewRequestWithContext(ctx, "GET", test.frontendURL, nil)
				require.NoError(t, err, "Should create frontend request")
				
				frontendResp, err := client.Do(frontendReq)
				if err != nil {
					t.Errorf("❌ FAIL: Frontend %s not accessible: %v", test.frontend, err)
					return
				}
				defer frontendResp.Body.Close()
				
				if frontendResp.StatusCode != http.StatusOK {
					t.Errorf("❌ FAIL: Frontend %s not serving: status %d", test.frontend, frontendResp.StatusCode)
					return
				}
				
				// Test backend endpoint that frontend contract client uses
				backendReq, err := http.NewRequestWithContext(ctx, "GET", test.gateway+test.backendEndpoint, nil)
				require.NoError(t, err, "Should create backend request")
				
				backendResp, err := client.Do(backendReq)
				if err != nil {
					t.Errorf("❌ FAIL: Backend endpoint for %s not accessible: %v", test.frontend, err)
					return
				}
				defer backendResp.Body.Close()
				
				if backendResp.StatusCode >= 200 && backendResp.StatusCode < 300 {
					t.Logf("✅ Frontend %s can access backend endpoint %s", test.frontend, test.backendEndpoint)
				} else {
					t.Errorf("❌ FAIL: Backend endpoint for %s failed: status %d", test.frontend, backendResp.StatusCode)
				}
			})
		}
	})

	t.Run("Contract client type definitions should match backend response structures", func(t *testing.T) {
		// Test that TypeScript client types match actual backend response structures
		
		typeValidationTests := []struct {
			endpoint        string
			gateway         string
			expectedTypes   []string
			expectedStructure map[string]string
		}{
			{
				endpoint:      "/api/news",
				gateway:       "http://localhost:9001", 
				expectedTypes: []string{"NewsArticle", "NewsCategory"},
				expectedStructure: map[string]string{
					"news_id":             "string",
					"title":               "string",
					"summary":             "string",
					"content":             "string",
					"category_id":         "string",
					"news_type":           "string",
					"priority_level":      "string",
					"publishing_status":   "string",
					"publication_timestamp": "string",
					"created_on":          "string",
					"slug":                "string",
				},
			},
			{
				endpoint:      "/api/services",
				gateway:       "http://localhost:9001",
				expectedTypes: []string{"Service", "ServiceCategory"},
				expectedStructure: map[string]string{
					"service_id":     "string",
					"title":          "string",
					"description":    "string",
					"category_id":    "string",
					"service_type":   "string",
					"availability":   "string",
					"created_on":     "string",
					"slug":           "string",
				},
			},
		}

		for _, test := range typeValidationTests {
			t.Run("Backend "+test.endpoint+" should return TypeScript-compatible data", func(t *testing.T) {
				client := &http.Client{Timeout: 5 * time.Second}
				url := test.gateway + test.endpoint
				
				req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
				require.NoError(t, err, "Should create type validation request")
				
				resp, err := client.Do(req)
				if err != nil {
					t.Errorf("❌ FAIL: Backend endpoint for type validation not accessible: %v", err)
					return
				}
				defer resp.Body.Close()
				
				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					var response map[string]interface{}
					if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
						t.Errorf("❌ FAIL: Backend response not JSON for type validation")
						return
					}
					
					// Check data array structure
					if dataField, hasData := response["data"]; hasData {
						if dataArray, ok := dataField.([]interface{}); ok && len(dataArray) > 0 {
							firstItem, ok := dataArray[0].(map[string]interface{})
							if !ok {
								t.Errorf("❌ FAIL: Backend data items not objects for TypeScript compatibility")
								return
							}
							
							// Validate field types match TypeScript expectations
							typeMatchErrors := 0
							for field, expectedType := range test.expectedStructure {
								if value, exists := firstItem[field]; exists {
									actualType := getJSONType(value)
									if actualType != expectedType {
										t.Errorf("❌ FAIL: Field '%s' type mismatch - TypeScript expects %s, backend returns %s", field, expectedType, actualType)
										typeMatchErrors++
									}
								} else {
									t.Errorf("❌ FAIL: Required field '%s' missing from backend response", field)
									typeMatchErrors++
								}
							}
							
							if typeMatchErrors == 0 {
								t.Logf("✅ Backend response structure matches TypeScript client expectations")
							} else {
								t.Errorf("❌ FAIL: %d type mismatches between backend and TypeScript client", typeMatchErrors)
							}
						} else {
							t.Error("❌ FAIL: Backend data field not array or empty")
						}
					} else {
						t.Error("❌ FAIL: Backend response missing data field")
					}
				} else {
					t.Errorf("❌ FAIL: Backend type validation endpoint failed: status %d", resp.StatusCode)
				}
			})
		}
	})

	t.Run("Generated client authentication should work with backend authentication", func(t *testing.T) {
		// Test that TypeScript client authentication patterns work with backend
		
		authTests := []struct {
			endpoint       string
			gateway        string
			requiresAuth   bool
			clientType     string
		}{
			{"/api/news", "http://localhost:9001", false, "public-client"},
			{"/api/admin/news", "http://localhost:9000", true, "admin-client"},
			{"/api/admin/inquiries", "http://localhost:9000", true, "admin-client"},
		}

		for _, test := range authTests {
			t.Run(test.clientType+" authentication should work with "+test.endpoint, func(t *testing.T) {
				client := &http.Client{Timeout: 5 * time.Second}
				url := test.gateway + test.endpoint
				
				// Test unauthenticated request
				req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
				require.NoError(t, err, "Should create auth test request")
				
				resp, err := client.Do(req)
				if err != nil {
					t.Errorf("❌ FAIL: Backend auth endpoint not accessible: %v", err)
					return
				}
				defer resp.Body.Close()
				
				if test.requiresAuth {
					// Should require authentication
					if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
						t.Logf("✅ Backend properly requires authentication for %s", test.endpoint)
					} else if resp.StatusCode == http.StatusOK {
						t.Errorf("❌ FAIL: Backend allows unauthenticated access to %s", test.endpoint)
						t.Log("    TypeScript admin clients expect authentication to be required")
					} else {
						t.Logf("⚠️  Backend authentication status for %s: %d", test.endpoint, resp.StatusCode)
					}
				} else {
					// Should allow anonymous access
					if resp.StatusCode == http.StatusOK {
						t.Logf("✅ Backend allows anonymous access for %s", test.endpoint)
					} else {
						t.Logf("⚠️  Backend anonymous access status for %s: %d", test.endpoint, resp.StatusCode)
					}
				}
			})
		}
	})
}

// Helper function to determine JSON type for TypeScript compatibility
func getJSONType(value interface{}) string {
	switch value.(type) {
	case string:
		return "string"
	case float64, int, int64:
		return "number"
	case bool:
		return "boolean"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	case nil:
		return "null"
	default:
		return "unknown"
	}
}

// TestContractClientGenerationIntegrity validates contract client generation process
func TestContractClientGenerationIntegrity(t *testing.T) {
	t.Run("Contract client generation should be consistent with backend implementations", func(t *testing.T) {
		// Test that contract client generation process is working correctly
		
		t.Log("❌ FAIL: Contract client generation validation not implemented")
		t.Log("    Need to validate TypeScript clients are generated from correct OpenAPI specifications")
		t.Log("    Need to validate generated clients match current backend API implementations")
		t.Log("    Need to validate client generation process maintains type safety")
		
		// This test should fail until client generation validation is implemented
		t.Fail()
	})

	t.Run("Frontend build process should include contract client validation", func(t *testing.T) {
		// Test that frontend build process validates contract clients
		
		t.Log("❌ FAIL: Frontend build contract validation not implemented")
		t.Log("    Frontend build should validate contract clients match backend")
		t.Log("    Build should fail if contract clients are out of sync with backend")
		
		// This test should fail until frontend build validation is implemented
		t.Fail()
	})
}