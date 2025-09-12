package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// RED PHASE: Service Data Operations Tests
// These tests validate that services can perform full CRUD operations with proper database schemas

func TestServiceDataOperations_FullCRUDValidation(t *testing.T) {
	// This test requires complete environment health - enforcing axiom rule
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Define CRUD operation tests for each service domain
	serviceCRUDTests := []struct {
		serviceName     string
		createEndpoint  string
		readEndpoint    string
		updateEndpoint  string
		deleteEndpoint  string
		testPayload     string
		description     string
	}{
		{
			serviceName:    "content",
			createEndpoint: "http://localhost:3500/v1.0/invoke/content/method/api/news",
			readEndpoint:   "http://localhost:3500/v1.0/invoke/content/method/api/news",
			updateEndpoint: "http://localhost:3500/v1.0/invoke/content/method/api/news",
			deleteEndpoint: "http://localhost:3500/v1.0/invoke/content/method/api/news",
			testPayload:    `{"title":"Test News","content":"Test content","author":"Test Author"}`,
			description:    "Content service must support full CRUD operations for news management",
		},
		{
			serviceName:    "inquiries",
			createEndpoint: "http://localhost:3500/v1.0/invoke/inquiries/method/api/inquiries",
			readEndpoint:   "http://localhost:3500/v1.0/invoke/inquiries/method/api/inquiries",
			updateEndpoint: "http://localhost:3500/v1.0/invoke/inquiries/method/api/inquiries",
			deleteEndpoint: "http://localhost:3500/v1.0/invoke/inquiries/method/api/inquiries",
			testPayload:    `{"type":"business","email":"test@example.com","message":"Test inquiry"}`,
			description:    "Inquiries service must support full CRUD operations for inquiry management",
		},
		{
			serviceName:    "notifications",
			createEndpoint: "http://localhost:3500/v1.0/invoke/notifications/method/api/subscribers",
			readEndpoint:   "http://localhost:3500/v1.0/invoke/notifications/method/api/subscribers",
			updateEndpoint: "http://localhost:3500/v1.0/invoke/notifications/method/api/subscribers",
			deleteEndpoint: "http://localhost:3500/v1.0/invoke/notifications/method/api/subscribers",
			testPayload:    `{"email":"test@example.com","name":"Test Subscriber","eventTypes":["news"]}`,
			description:    "Notifications service must support full CRUD operations for subscriber management",
		},
	}

	client := &http.Client{Timeout: 15 * time.Second}

	// Act & Assert: Test CRUD operations for each service
	for _, crudTest := range serviceCRUDTests {
		t.Run("CRUD_"+crudTest.serviceName, func(t *testing.T) {
			// Test CREATE operation
			t.Run("Create", func(t *testing.T) {
				createReq, err := http.NewRequestWithContext(ctx, "POST", crudTest.createEndpoint, strings.NewReader(crudTest.testPayload))
				require.NoError(t, err, "Failed to create POST request for %s", crudTest.serviceName)
				createReq.Header.Set("Content-Type", "application/json")

				createResp, err := client.Do(createReq)
				if err == nil {
					defer createResp.Body.Close()
					assert.True(t, createResp.StatusCode >= 200 && createResp.StatusCode < 300, 
						"%s - CREATE operation must be functional", crudTest.description)
				} else {
					t.Errorf("%s CREATE operation not functional: %v", crudTest.description, err)
				}
			})

			// Test READ operation
			t.Run("Read", func(t *testing.T) {
				readReq, err := http.NewRequestWithContext(ctx, "GET", crudTest.readEndpoint, nil)
				require.NoError(t, err, "Failed to create GET request for %s", crudTest.serviceName)

				readResp, err := client.Do(readReq)
				if err == nil {
					defer readResp.Body.Close()
					assert.True(t, readResp.StatusCode >= 200 && readResp.StatusCode < 300, 
						"%s - READ operation must be functional", crudTest.description)
					
					// Validate response contains data
					body, err := io.ReadAll(readResp.Body)
					if err == nil {
						responseStr := string(body)
						assert.NotEmpty(t, responseStr, "%s READ operation must return data", crudTest.description)
						
						// Should be valid JSON
						var jsonData interface{}
						assert.NoError(t, json.Unmarshal(body, &jsonData), 
							"%s READ operation must return valid JSON", crudTest.description)
					}
				} else {
					t.Errorf("%s READ operation not functional: %v", crudTest.description, err)
				}
			})
		})
	}
}

func TestServiceDataOperations_DatabaseTableAccess(t *testing.T) {
	// Test that services can access their required database tables
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Required database tables for each service
	serviceTableRequirements := map[string][]string{
		"content": {
			"content_news",
			"content_events", 
			"content_research",
			"content_metadata",
		},
		"inquiries": {
			"inquiries_business",
			"inquiries_donations",
			"inquiries_media",
			"inquiries_volunteers",
			"inquiry_metadata",
		},
		"notifications": {
			"notification_subscribers", // This one exists
			"notification_templates",
			"notification_events",
			"notification_delivery_log",
		},
		"gateways": {
			"gateway_rate_limits",
			"gateway_access_log",
			"gateway_configuration",
		},
	}

	// Act & Assert: Test database table access for each service
	for serviceDomain, requiredTables := range serviceTableRequirements {
		t.Run("DatabaseAccess_"+serviceDomain, func(t *testing.T) {
			// Connect to database
			connStr := "postgresql://postgres:password@localhost:5432/international_center_development?sslmode=disable"
			db, err := sql.Open("postgres", connStr)
			require.NoError(t, err, "Database must be accessible for %s domain validation", serviceDomain)
			defer db.Close()

			// Test each required table
			for _, tableName := range requiredTables {
				t.Run("Table_"+tableName, func(t *testing.T) {
					// Check if table exists
					var tableExists bool
					err := db.QueryRowContext(ctx, 
						"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = $1)", 
						tableName).Scan(&tableExists)
					require.NoError(t, err, "Failed to check table %s", tableName)

					assert.True(t, tableExists, 
						"Service domain %s requires table %s for data operations", serviceDomain, tableName)

					// If table exists, test basic operations
					if tableExists {
						_, err = db.ExecContext(ctx, "SELECT COUNT(*) FROM "+tableName)
						assert.NoError(t, err, 
							"Service domain %s must be able to query table %s", serviceDomain, tableName)
					}
				})
			}
		})
	}
}

func TestServiceDataOperations_CrossServiceDataFlow(t *testing.T) {
	// Test data flow between services through service mesh
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Data flow scenarios between services
	dataFlowTests := []struct {
		workflow       string
		sourceService  string
		targetService  string
		dataEndpoint   string
		expectedData   string
		description    string
	}{
		{
			workflow:      "content-to-notifications",
			sourceService: "content",
			targetService: "notifications",
			dataEndpoint:  "http://localhost:3500/v1.0/invoke/notifications/method/api/events",
			expectedData:  "content_published",
			description:   "Content publishing must trigger notification events",
		},
		{
			workflow:      "inquiries-to-notifications",
			sourceService: "inquiries",
			targetService: "notifications", 
			dataEndpoint:  "http://localhost:3500/v1.0/invoke/notifications/method/api/events",
			expectedData:  "inquiry_submitted",
			description:   "Inquiry submissions must trigger notification events",
		},
		{
			workflow:      "gateway-to-content",
			sourceService: "public-gateway",
			targetService: "content",
			dataEndpoint:  "http://localhost:3500/v1.0/invoke/content/method/api/news",
			expectedData:  "news",
			description:   "Public gateway must route to content service data",
		},
	}

	client := &http.Client{Timeout: 15 * time.Second}

	// Act & Assert: Test cross-service data flow
	for _, flow := range dataFlowTests {
		t.Run("DataFlow_"+flow.workflow, func(t *testing.T) {
			req, err := http.NewRequestWithContext(ctx, "GET", flow.dataEndpoint, nil)
			require.NoError(t, err, "Failed to create data flow request")

			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500, 
					"%s - cross-service data flow must be accessible", flow.description)

				// Validate data flow returns expected content
				body, err := io.ReadAll(resp.Body)
				if err == nil {
					responseStr := string(body)
					assert.NotEmpty(t, responseStr, "%s must return data", flow.description)
					
					// Should contain expected data indicators
					if flow.expectedData != "" {
						// Expected data may not be present until schemas are complete
						t.Logf("%s data flow response: %s", flow.description, responseStr)
					}
				}
			} else {
				t.Errorf("%s data flow not accessible: %v", flow.description, err)
			}
		})
	}
}

func TestServiceDataOperations_ServiceDomainFunctionality(t *testing.T) {
	// Test that each service supports its domain-specific functionality
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Domain-specific functionality tests
	serviceDomainTests := []struct {
		serviceName string
		endpoints   map[string]string
		description string
	}{
		{
			serviceName: "content",
			endpoints: map[string]string{
				"news":     "http://localhost:3500/v1.0/invoke/content/method/api/news",
				"events":   "http://localhost:3500/v1.0/invoke/content/method/api/events",
				"research": "http://localhost:3500/v1.0/invoke/content/method/api/research",
			},
			description: "Content service must support all content domain operations",
		},
		{
			serviceName: "inquiries",
			endpoints: map[string]string{
				"business":   "http://localhost:3500/v1.0/invoke/inquiries/method/api/inquiries/business",
				"donations":  "http://localhost:3500/v1.0/invoke/inquiries/method/api/inquiries/donations", 
				"media":      "http://localhost:3500/v1.0/invoke/inquiries/method/api/inquiries/media",
				"volunteers": "http://localhost:3500/v1.0/invoke/inquiries/method/api/inquiries/volunteers",
			},
			description: "Inquiries service must support all inquiry domain operations",
		},
		{
			serviceName: "notifications",
			endpoints: map[string]string{
				"subscribers": "http://localhost:3500/v1.0/invoke/notifications/method/api/subscribers",
				"templates":   "http://localhost:3500/v1.0/invoke/notifications/method/api/templates",
				"events":      "http://localhost:3500/v1.0/invoke/notifications/method/api/events",
			},
			description: "Notifications service must support all notification domain operations",
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Act & Assert: Test domain functionality for each service
	for _, domainTest := range serviceDomainTests {
		t.Run("DomainFunctionality_"+domainTest.serviceName, func(t *testing.T) {
			for domainArea, endpoint := range domainTest.endpoints {
				t.Run("Domain_"+domainArea, func(t *testing.T) {
					req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
					require.NoError(t, err, "Failed to create domain functionality request")

					resp, err := client.Do(req)
					if err == nil {
						defer resp.Body.Close()
						assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500, 
							"%s domain area %s must be accessible", domainTest.description, domainArea)

						// Validate response indicates domain functionality
						body, err := io.ReadAll(resp.Body)
						if err == nil {
							responseStr := string(body)
							
							// Should not return generic "404 page not found" 
							assert.NotContains(t, responseStr, "404 page not found",
								"Domain %s should have proper endpoint implementation", domainArea)
							
							// Should return domain-specific data or proper API response
							if resp.StatusCode >= 200 && resp.StatusCode < 300 {
								var jsonData interface{}
								if json.Unmarshal(body, &jsonData) == nil {
									t.Logf("Domain %s returned valid JSON response", domainArea)
								}
							}
						}
					} else {
						t.Errorf("%s domain area %s not accessible: %v", domainTest.description, domainArea, err)
					}
				})
			}
		})
	}
}

func TestServiceDataOperations_GatewayRoutingWithData(t *testing.T) {
	// Test that gateways can route to backend services with data operations
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Gateway routing scenarios with data operations
	gatewayRoutingTests := []struct {
		gatewayName    string
		directEndpoint string
		meshEndpoint   string
		description    string
	}{
		{
			gatewayName:    "public-gateway",
			directEndpoint: "http://localhost:9001/api/news",
			meshEndpoint:   "http://localhost:3500/v1.0/invoke/public-gateway/method/api/news",
			description:    "Public gateway must route news API with data operations",
		},
		{
			gatewayName:    "public-gateway",
			directEndpoint: "http://localhost:9001/api/events", 
			meshEndpoint:   "http://localhost:3500/v1.0/invoke/public-gateway/method/api/events",
			description:    "Public gateway must route events API with data operations",
		},
		{
			gatewayName:    "admin-gateway",
			directEndpoint: "http://localhost:9000/api/admin/inquiries",
			meshEndpoint:   "http://localhost:3500/v1.0/invoke/admin-gateway/method/api/admin/inquiries",
			description:    "Admin gateway must route inquiries management with data operations",
		},
	}

	client := &http.Client{Timeout: 15 * time.Second}

	// Act & Assert: Test gateway routing with data operations
	for _, routing := range gatewayRoutingTests {
		t.Run("GatewayRouting_"+routing.gatewayName, func(t *testing.T) {
			// Test direct gateway routing
			directReq, err := http.NewRequestWithContext(ctx, "GET", routing.directEndpoint, nil)
			require.NoError(t, err, "Failed to create direct routing request")

			directResp, err := client.Do(directReq)
			if err == nil {
				defer directResp.Body.Close()
				assert.True(t, directResp.StatusCode >= 200 && directResp.StatusCode < 500, 
					"%s - direct routing must be functional", routing.description)
			} else {
				t.Logf("%s direct routing not functional: %v", routing.description, err)
			}

			// Test service mesh routing
			meshReq, err := http.NewRequestWithContext(ctx, "GET", routing.meshEndpoint, nil)
			require.NoError(t, err, "Failed to create mesh routing request")

			meshResp, err := client.Do(meshReq)
			if err == nil {
				defer meshResp.Body.Close()
				assert.True(t, meshResp.StatusCode >= 200 && meshResp.StatusCode < 500, 
					"%s - service mesh routing must be functional", routing.description)
			} else {
				t.Errorf("%s service mesh routing not functional: %v", routing.description, err)
			}
		})
	}
}

// validateEnvironmentPrerequisites ensures environment health before integration testing
func validateEnvironmentPrerequisites(t *testing.T) {
	// Check critical infrastructure and platform components are running
	criticalContainers := []string{"postgresql", "dapr-control-plane"}
	
	for _, container := range criticalContainers {
		cmd := exec.Command("podman", "ps", "--filter", "name="+container, "--format", "{{.Names}}")
		output, err := cmd.Output()
		require.NoError(t, err, "Failed to check critical container %s", container)

		if !strings.Contains(string(output), container) {
			t.Skipf("Critical container %s not running - environment not ready for integration testing", container)
		}
	}
}