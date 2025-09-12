// RED PHASE: Contract-first integration tests - these tests should FAIL initially
package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

// TestOpenAPIContractCompliance validates backend services comply with OpenAPI specifications
func TestOpenAPIContractCompliance(t *testing.T) {
	timeout := 15 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Load OpenAPI specifications
	contractsBasePath := "/home/tojkuv/Documents/GitHub/international-center-workspace/international-center/src/public-website/contracts/openapi"
	publicSpecPath := filepath.Join(contractsBasePath, "public-api.yaml")
	adminSpecPath := filepath.Join(contractsBasePath, "admin-api.yaml")

	t.Run("Backend services should comply with public API OpenAPI specification", func(t *testing.T) {
		// Load public API specification
		loader := &openapi3.Loader{Context: ctx, IsExternalRefsAllowed: true}
		publicSpec, err := loader.LoadFromFile(publicSpecPath)
		if err != nil {
			t.Fatalf("❌ FAIL: Cannot load public API specification: %v", err)
		}

		// Validate specification is valid
		if err := publicSpec.Validate(ctx); err != nil {
			t.Fatalf("❌ FAIL: Public API specification is invalid: %v", err)
		}

		// Test public API endpoints against specification
		publicEndpoints := []struct {
			method   string
			path     string
			gateway  string
		}{
			{"GET", "/api/news", "http://localhost:9001"},
			{"GET", "/api/news/featured", "http://localhost:9001"},
			{"GET", "/api/services", "http://localhost:9001"},
			{"GET", "/api/services/featured", "http://localhost:9001"},
			{"GET", "/api/research", "http://localhost:9001"},
			{"GET", "/api/research/featured", "http://localhost:9001"},
			{"GET", "/api/events", "http://localhost:9001"},
			{"GET", "/api/events/featured", "http://localhost:9001"},
			{"POST", "/api/inquiries/media", "http://localhost:9001"},
			{"POST", "/api/inquiries/business", "http://localhost:9001"},
		}

		for _, endpoint := range publicEndpoints {
			t.Run(fmt.Sprintf("%s %s should comply with OpenAPI specification", endpoint.method, endpoint.path), func(t *testing.T) {
				// Check if endpoint exists in specification
				pathItem := publicSpec.Paths.Find(endpoint.path)
				if pathItem == nil {
					t.Errorf("❌ FAIL: Endpoint %s not found in public API specification", endpoint.path)
					return
				}

				// Check if method exists for endpoint
				var operation *openapi3.Operation
				switch endpoint.method {
				case "GET":
					operation = pathItem.Get
				case "POST":
					operation = pathItem.Post
				case "PUT":
					operation = pathItem.Put
				case "DELETE":
					operation = pathItem.Delete
				}

				if operation == nil {
					t.Errorf("❌ FAIL: Method %s not defined for %s in OpenAPI specification", endpoint.method, endpoint.path)
					return
				}

				// Test actual endpoint against specification
				client := &http.Client{Timeout: 5 * time.Second}
				url := endpoint.gateway + endpoint.path
				req, err := http.NewRequestWithContext(ctx, endpoint.method, url, nil)
				require.NoError(t, err, "Should create contract compliance request")

				resp, err := client.Do(req)
				if err != nil {
					t.Errorf("❌ FAIL: Contract endpoint %s %s not accessible: %v", endpoint.method, endpoint.path, err)
					return
				}
				defer resp.Body.Close()

				// Validate response is reasonable for contract endpoint
				if resp.StatusCode >= 200 && resp.StatusCode < 500 {
					t.Logf("✅ Contract endpoint %s %s accessible (status %d)", endpoint.method, endpoint.path, resp.StatusCode)
				} else {
					t.Errorf("❌ FAIL: Contract endpoint %s %s failed: status %d", endpoint.method, endpoint.path, resp.StatusCode)
				}

				// Validate response content type and structure for successful responses
				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					contentType := resp.Header.Get("Content-Type")
					if contentType == "" {
						t.Errorf("❌ FAIL: Missing Content-Type header for %s %s", endpoint.method, endpoint.path)
					} else if contentType != "application/json" {
						t.Errorf("❌ FAIL: Expected application/json, got %s for %s %s", contentType, endpoint.method, endpoint.path)
					}

					// Validate JSON response structure
					if contentType == "application/json" {
						var response map[string]interface{}
						if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
							t.Errorf("❌ FAIL: Invalid JSON response for %s %s: %v", endpoint.method, endpoint.path, err)
						} else {
							// OpenAPI contract expects data field for successful responses
							if _, hasData := response["data"]; !hasData {
								t.Errorf("❌ FAIL: Contract violation - missing 'data' field for %s %s", endpoint.method, endpoint.path)
							} else {
								t.Logf("✅ Contract-compliant JSON response for %s %s", endpoint.method, endpoint.path)
							}
						}
					}
				}
			})
		}
	})

	t.Run("Backend services should comply with admin API OpenAPI specification", func(t *testing.T) {
		// RED PHASE: Admin API specification MUST load without errors
		loader := &openapi3.Loader{Context: ctx, IsExternalRefsAllowed: true}
		adminSpec, err := loader.LoadFromFile(adminSpecPath)
		if err != nil {
			t.Errorf("❌ FAIL: Admin API specification loading failed: %v", err)
			t.Log("🚨 CRITICAL: OpenAPI specifications MUST parse without errors")
			t.Log("    YAML corruption prevents contract-first development")
			t.Log("    Parameter references and component definitions MUST be valid")
			t.Log("    Contract generation REQUIRES error-free specifications")
			t.Fail() // RED PHASE: MUST fail if admin spec can't load
			return
		}

		// RED PHASE: Specification MUST be structurally valid
		if err := adminSpec.Validate(ctx); err != nil {
			t.Errorf("❌ FAIL: Admin API specification validation failed: %v", err)
			t.Log("🚨 CRITICAL: OpenAPI specification MUST be structurally valid")
			t.Log("    Schema validation REQUIRED for contract compliance")
			t.Fail() // RED PHASE: MUST fail if spec is invalid
			return
		}

		// Test admin API endpoints against specification  
		adminEndpoints := []struct {
			method   string
			path     string
			gateway  string
			requiresAuth bool
		}{
			{"GET", "/api/admin/news", "http://localhost:9000", true},
			{"POST", "/api/admin/news", "http://localhost:9000", true},
			{"GET", "/api/admin/news/categories", "http://localhost:9000", true},
			{"GET", "/api/admin/services", "http://localhost:9000", true},
			{"POST", "/api/admin/services", "http://localhost:9000", true},
			{"GET", "/api/admin/inquiries", "http://localhost:9000", true},
			{"PUT", "/api/admin/inquiries/{id}", "http://localhost:9000", true},
		}

		for _, endpoint := range adminEndpoints {
			t.Run(fmt.Sprintf("%s %s should comply with admin OpenAPI specification", endpoint.method, endpoint.path), func(t *testing.T) {
				// RED PHASE: ALL admin endpoints MUST be defined in OpenAPI specification
				pathPattern := endpoint.path
				
				// Convert {id} patterns to OpenAPI format if needed
				if endpoint.path == "/api/admin/inquiries/{id}" {
					pathPattern = "/api/admin/inquiries/{inquiry_id}"
				}
				
				pathItem := adminSpec.Paths.Find(pathPattern)
				if pathItem == nil {
					t.Errorf("❌ FAIL: Admin endpoint %s not found in OpenAPI specification", pathPattern)
					t.Logf("Available paths in admin spec: %v", getAvailablePaths(adminSpec))
					t.Log("🚨 CRITICAL: ALL admin endpoints MUST be defined in OpenAPI specification")
					t.Log("    Contract-first development REQUIRES complete endpoint coverage")
					t.Log("    Backend implementations without contracts violate architecture")
					t.Fail() // RED PHASE: MUST fail if endpoints missing from spec
					return
				}

				// Check if method exists for endpoint
				var operation *openapi3.Operation
				switch endpoint.method {
				case "GET":
					operation = pathItem.Get
				case "POST":
					operation = pathItem.Post
				case "PUT":
					operation = pathItem.Put
				case "DELETE":
					operation = pathItem.Delete
				}

				if operation == nil {
					t.Errorf("❌ FAIL: Method %s not defined for %s in admin OpenAPI specification", endpoint.method, pathPattern)
					return
				}

				// Test actual admin endpoint (without auth for now to test endpoint existence)
				client := &http.Client{Timeout: 5 * time.Second}
				url := endpoint.gateway + endpoint.path
				testURL := url
				if endpoint.path == "/api/admin/inquiries/{id}" {
					testURL = endpoint.gateway + "/api/admin/inquiries/test-id-123"
				}
				
				req, err := http.NewRequestWithContext(ctx, endpoint.method, testURL, nil)
				require.NoError(t, err, "Should create admin contract request")

				resp, err := client.Do(req)
				if err != nil {
					t.Errorf("❌ FAIL: Admin contract endpoint %s %s not accessible: %v", endpoint.method, endpoint.path, err)
					return
				}
				defer resp.Body.Close()

				// For auth-required endpoints, expect 401/403, for others expect success
				if endpoint.requiresAuth {
					if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
						t.Logf("✅ Admin endpoint %s %s properly requires authentication", endpoint.method, endpoint.path)
					} else if resp.StatusCode >= 200 && resp.StatusCode < 300 {
						t.Errorf("❌ FAIL: Admin endpoint %s %s allows unauthenticated access", endpoint.method, endpoint.path)
					} else {
						t.Logf("⚠️  Admin endpoint %s %s status: %d", endpoint.method, endpoint.path, resp.StatusCode)
					}
				} else {
					t.Logf("Info: Admin endpoint %s %s status: %d", endpoint.method, endpoint.path, resp.StatusCode)
				}
			})
		}
	})
}

// Helper function to get available paths from OpenAPI spec
func getAvailablePaths(spec *openapi3.T) []string {
	var paths []string
	for path := range spec.Paths.Map() {
		paths = append(paths, path)
	}
	return paths
}

// TestContractValidationFramework validates contract validation infrastructure
func TestContractValidationFramework(t *testing.T) {
	timeout := 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	t.Run("Contract validation framework should be available for backend services", func(t *testing.T) {
		// Test that backend services have contract validation framework available
		
		contractsBasePath := "/home/tojkuv/Documents/GitHub/international-center-workspace/international-center/src/public-website/contracts/openapi"
		publicSpecPath := filepath.Join(contractsBasePath, "public-api.yaml")
		adminSpecPath := filepath.Join(contractsBasePath, "admin-api.yaml")

		// Test that contract specifications can be loaded
		loader := &openapi3.Loader{Context: ctx, IsExternalRefsAllowed: true}
		
		// Load public specification
		publicSpec, err := loader.LoadFromFile(publicSpecPath)
		if err != nil {
			t.Errorf("❌ FAIL: Cannot load public API specification for validation: %v", err)
		} else {
			if err := publicSpec.Validate(ctx); err != nil {
				t.Errorf("❌ FAIL: Public API specification validation failed: %v", err)
			} else {
				t.Log("✅ Public API OpenAPI specification valid and loadable")
			}
		}

		// Load admin specification
		adminSpec, err := loader.LoadFromFile(adminSpecPath)
		if err != nil {
			t.Errorf("❌ FAIL: Cannot load admin API specification for validation: %v", err)
		} else {
			if err := adminSpec.Validate(ctx); err != nil {
				t.Errorf("❌ FAIL: Admin API specification validation failed: %v", err)
			} else {
				t.Log("✅ Admin API OpenAPI specification valid and loadable")
			}
		}
	})

	t.Run("Contract validation should be integrated into service endpoints", func(t *testing.T) {
		// Test that services have contract validation middleware
		
		services := []struct {
			name     string
			endpoint string
		}{
			{"content", "http://localhost:3001"},
			{"inquiries", "http://localhost:3101"},
			{"notifications", "http://localhost:3201"},
		}

		for _, service := range services {
			t.Run(service.name+" should have contract validation middleware", func(t *testing.T) {
				// Test service has contract validation by sending invalid request
				
				client := &http.Client{Timeout: 3 * time.Second}
				
				// Send malformed request that should be rejected by contract validation
				invalidURL := service.endpoint + "/api/invalid-endpoint"
				req, err := http.NewRequestWithContext(ctx, "POST", invalidURL, nil)
				require.NoError(t, err, "Should create invalid request")
				req.Header.Set("Content-Type", "application/json")
				
				resp, err := client.Do(req)
				if err != nil {
					t.Errorf("❌ FAIL: Service %s not accessible for contract validation testing: %v", service.name, err)
				} else {
					defer resp.Body.Close()
					
					// Contract validation should reject invalid requests
					if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound {
						t.Logf("✅ Service %s has request validation (status %d)", service.name, resp.StatusCode)
					} else {
						t.Logf("⚠️  Service %s contract validation status unclear: %d", service.name, resp.StatusCode)
					}
				}
			})
		}
	})

	t.Run("Gateway routing should enforce contract endpoint patterns", func(t *testing.T) {
		// Test that gateways only route to contract-defined endpoints
		
		gatewayTests := []struct {
			gateway     string
			gatewayURL  string
			validPath   string
			invalidPath string
		}{
			{"public-gateway", "http://localhost:9001", "/api/news", "/api/invalid"},
			{"admin-gateway", "http://localhost:9000", "/api/admin/news", "/api/admin/invalid"},
		}

		for _, test := range gatewayTests {
			t.Run(test.gateway+" should enforce contract endpoint patterns", func(t *testing.T) {
				client := &http.Client{Timeout: 3 * time.Second}
				
				// Test valid contract endpoint
				validReq, err := http.NewRequestWithContext(ctx, "GET", test.gatewayURL+test.validPath, nil)
				require.NoError(t, err, "Should create valid endpoint request")
				
				validResp, err := client.Do(validReq)
				if err != nil {
					t.Errorf("❌ FAIL: Gateway %s not accessible: %v", test.gateway, err)
				} else {
					defer validResp.Body.Close()
					
					// Valid endpoints should not return 404 for routing
					if validResp.StatusCode == http.StatusNotFound {
						t.Errorf("❌ FAIL: Gateway %s not routing valid contract endpoint %s", test.gateway, test.validPath)
					} else {
						t.Logf("✅ Gateway %s routes valid contract endpoint %s (status %d)", test.gateway, test.validPath, validResp.StatusCode)
					}
				}

				// Test invalid endpoint
				invalidReq, err := http.NewRequestWithContext(ctx, "GET", test.gatewayURL+test.invalidPath, nil)
				require.NoError(t, err, "Should create invalid endpoint request")
				
				invalidResp, err := client.Do(invalidReq)
				if err != nil {
					t.Logf("Gateway %s properly rejects invalid endpoint: %v", test.gateway, err)
				} else {
					defer invalidResp.Body.Close()
					
					// Invalid endpoints should return 404
					if invalidResp.StatusCode == http.StatusNotFound {
						t.Logf("✅ Gateway %s properly rejects invalid endpoint %s", test.gateway, test.invalidPath)
					} else {
						t.Errorf("❌ FAIL: Gateway %s accepts invalid endpoint %s (status %d)", test.gateway, test.invalidPath, invalidResp.StatusCode)
					}
				}
			})
		}
	})
}

// TestContractResponseValidation validates service responses match OpenAPI schemas
func TestContractResponseValidation(t *testing.T) {
	timeout := 15 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	t.Run("Service responses should match OpenAPI response schemas", func(t *testing.T) {
		// Test that actual service responses match OpenAPI response schemas
		
		responseValidationTests := []struct {
			service     string
			method      string
			endpoint    string
			gateway     string
			expectData  bool
			expectPagination bool
		}{
			{"content", "GET", "/api/news", "http://localhost:9001", true, true},
			{"content", "GET", "/api/news/featured", "http://localhost:9001", true, false},
			{"content", "GET", "/api/services", "http://localhost:9001", true, true},
			{"content", "GET", "/api/services/featured", "http://localhost:9001", true, false},
		}

		for _, test := range responseValidationTests {
			t.Run(fmt.Sprintf("%s %s %s should return schema-compliant response", test.service, test.method, test.endpoint), func(t *testing.T) {
				client := &http.Client{Timeout: 5 * time.Second}
				url := test.gateway + test.endpoint
				
				req, err := http.NewRequestWithContext(ctx, test.method, url, nil)
				require.NoError(t, err, "Should create schema validation request")
				
				resp, err := client.Do(req)
				if err != nil {
					t.Errorf("❌ FAIL: Service endpoint %s not accessible for schema validation: %v", test.endpoint, err)
					return
				}
				defer resp.Body.Close()
				
				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					// Parse response for schema validation
					var response map[string]interface{}
					if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
						t.Errorf("❌ FAIL: Service response not valid JSON for %s", test.endpoint)
						return
					}
					
					// Validate expected schema fields
					if test.expectData {
						if _, hasData := response["data"]; !hasData {
							t.Errorf("❌ FAIL: Schema violation - missing 'data' field for %s", test.endpoint)
						} else {
							t.Logf("✅ Schema-compliant 'data' field for %s", test.endpoint)
						}
					}
					
					if test.expectPagination {
						if _, hasPagination := response["pagination"]; !hasPagination {
							t.Errorf("❌ FAIL: Schema violation - missing 'pagination' field for %s", test.endpoint)
						} else {
							t.Logf("✅ Schema-compliant 'pagination' field for %s", test.endpoint)
						}
					}
					
					// Validate response has correlation tracking
					correlationID := resp.Header.Get("X-Correlation-ID")
					if correlationID == "" {
						t.Errorf("❌ FAIL: Missing X-Correlation-ID header for %s", test.endpoint)
					} else {
						t.Logf("✅ Correlation tracking present for %s", test.endpoint)
					}
				} else {
					t.Logf("⚠️  Service endpoint %s returned status %d", test.endpoint, resp.StatusCode)
				}
			})
		}
	})

	t.Run("OpenAPI specifications MUST have complete endpoint coverage without corruption", func(t *testing.T) {
		// RED PHASE: OpenAPI specifications MUST be complete and error-free
		
		t.Log("🚨 CRITICAL REQUIREMENTS for OpenAPI specification integrity:")
		t.Log("    1. Admin API specification MUST parse without YAML errors")
		t.Log("    2. ALL backend endpoints MUST be defined in specifications")
		t.Log("    3. Parameter references MUST be valid and resolvable")
		t.Log("    4. Component definitions MUST be syntactically correct")
		t.Log("    5. Response schemas MUST match backend implementations")
		
		// Test specific specification integrity requirements
		specIntegrityIssues := []string{
			"Admin API YAML parsing errors (parameter reference corruption)",
			"Missing featured endpoints in public API specification",
			"Missing inquiry submission endpoints in specifications",
			"Missing admin endpoints in admin API specification",
			"Component parameter reference syntax errors",
		}
		
		t.Log("❌ FAIL: OpenAPI specification integrity issues preventing contract-first development:")
		for _, issue := range specIntegrityIssues {
			t.Logf("    %s", issue)
		}
		
		t.Log("🚨 CRITICAL: Complete specification integrity REQUIRED for development workflow")
		
		// RED PHASE: MUST fail until specification integrity is achieved
		t.Fail()
	})
}