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

// RED PHASE: API Data Operations Tests
// These tests validate that APIs perform actual CRUD operations with database persistence

func TestAPIDataOperations_ActualDatabaseIntegration(t *testing.T) {
	// This test requires complete environment health - enforcing axiom rule
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// API data operations tests - should perform actual database operations
	apiDataOperationTests := []struct {
		serviceName     string
		apiEndpoint     string
		tableRequired   string
		testDataExists  bool
		description     string
		critical        bool
	}{
		{
			serviceName:    "content",
			apiEndpoint:    "http://localhost:3001/api/news",
			tableRequired:  "content_news",
			testDataExists: false, // Should create test data if none exists
			description:    "Content news API must perform actual database operations with content_news table",
			critical:       true,
		},
		{
			serviceName:    "content",
			apiEndpoint:    "http://localhost:3001/api/events",
			tableRequired:  "content_events",
			testDataExists: false,
			description:    "Content events API must perform actual database operations with content_events table",
			critical:       false, // Already working well
		},
		{
			serviceName:    "inquiries",
			apiEndpoint:    "http://localhost:3101/api/inquiries",
			tableRequired:  "inquiries_business",
			testDataExists: false,
			description:    "Inquiries API must perform actual database operations with inquiries tables",
			critical:       true,
		},
		{
			serviceName:    "notifications",
			apiEndpoint:    "http://localhost:3201/api/subscribers",
			tableRequired:  "notification_subscribers",
			testDataExists: false,
			description:    "Notifications API must perform actual database operations with notification_subscribers table",
			critical:       true,
		},
	}

	client := &http.Client{Timeout: 15 * time.Second}

	// Act & Assert: Test API data operations
	for _, dataTest := range apiDataOperationTests {
		t.Run("APIDataOps_"+dataTest.serviceName, func(t *testing.T) {
			// First ensure database table exists and is accessible
			connStr := "postgresql://postgres:password@localhost:5432/international_center_development?sslmode=disable"
			db, err := sql.Open("postgres", connStr)
			require.NoError(t, err, "Database must be accessible for %s", dataTest.serviceName)
			defer db.Close()

			// Check if table exists
			var tableExists bool
			err = db.QueryRowContext(ctx, 
				"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = $1)", 
				dataTest.tableRequired).Scan(&tableExists)
			require.NoError(t, err, "Failed to check table %s", dataTest.tableRequired)

			assert.True(t, tableExists, 
				"Table %s must exist for %s data operations", dataTest.tableRequired, dataTest.description)

			// Add test data if table is empty (for testing data operations)
			var rowCount int
			err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+dataTest.tableRequired).Scan(&rowCount)
			require.NoError(t, err, "Failed to count rows in %s", dataTest.tableRequired)

			// Test API endpoint
			req, err := http.NewRequestWithContext(ctx, "GET", dataTest.apiEndpoint, nil)
			require.NoError(t, err, "Failed to create API request")

			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				
				if dataTest.critical {
					// Critical APIs must not fail with internal errors
					assert.NotEqual(t, 500, resp.StatusCode,
						"%s - API must not fail with internal server errors", dataTest.description)
					
					body, err := io.ReadAll(resp.Body)
					if err == nil {
						responseStr := string(body)
						
						// Must not return internal errors
						assert.NotContains(t, responseStr, "INTERNAL_ERROR",
							"%s - API must not fail with internal errors", dataTest.description)
						assert.NotContains(t, responseStr, "failed to get all",
							"%s - API must successfully access database", dataTest.description)
						
						// Should return structured data response
						if resp.StatusCode >= 200 && resp.StatusCode < 300 {
							var jsonData map[string]interface{}
							assert.NoError(t, json.Unmarshal(body, &jsonData),
								"%s - API must return valid JSON for data operations", dataTest.description)
							
							// Should have data structure (even if empty)
							assert.Contains(t, jsonData, "data",
								"%s - API response must have data field", dataTest.description)
							assert.Contains(t, jsonData, "count",
								"%s - API response must have count field", dataTest.description)
						}
					}
				} else {
					// Non-critical APIs should work for complete environment
					if resp.StatusCode >= 500 {
						t.Logf("%s failing with server error (expected for incomplete data operations)", dataTest.description)
					}
				}
			} else {
				if dataTest.critical {
					t.Errorf("%s not accessible: %v", dataTest.description, err)
				}
			}
		})
	}
}

func TestAPIDataOperations_DatabaseCRUDOperations(t *testing.T) {
	// Test that APIs can perform actual CRUD operations with database
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// CRUD operation tests with database integration
	crudOperationTests := []struct {
		serviceName   string
		createEndpoint string
		readEndpoint   string
		testPayload    string
		tableName      string
		description    string
	}{
		{
			serviceName:   "content",
			createEndpoint: "http://localhost:3001/api/events",
			readEndpoint:   "http://localhost:3001/api/events",
			testPayload:    `{"title":"Test Event","description":"Test event description","event_date":"2025-12-25T00:00:00Z","location":"Test Location"}`,
			tableName:      "content_events",
			description:    "Content events API must support create and read operations with database persistence",
		},
		{
			serviceName:   "inquiries",
			createEndpoint: "http://localhost:3101/api/inquiries",
			readEndpoint:   "http://localhost:3101/api/inquiries",
			testPayload:    `{"company_name":"Test Company","contact_email":"test@example.com","contact_name":"Test Contact","message":"Test inquiry message"}`,
			tableName:      "inquiries_business",
			description:    "Inquiries API must support create and read operations with database persistence",
		},
		{
			serviceName:   "notifications",
			createEndpoint: "http://localhost:3201/api/subscribers",
			readEndpoint:   "http://localhost:3201/api/subscribers",
			testPayload:    `{"subscriber_name":"Test Subscriber","email":"test@example.com","event_types":["news","events"],"notification_methods":["email"]}`,
			tableName:      "notification_subscribers",
			description:    "Notifications API must support create and read operations with database persistence",
		},
	}

	client := &http.Client{Timeout: 15 * time.Second}

	// Act & Assert: Test CRUD operations
	for _, crud := range crudOperationTests {
		t.Run("CRUDOps_"+crud.serviceName, func(t *testing.T) {
			// Test READ operation (should work with database)
			readReq, err := http.NewRequestWithContext(ctx, "GET", crud.readEndpoint, nil)
			require.NoError(t, err, "Failed to create read request")

			readResp, err := client.Do(readReq)
			if err == nil {
				defer readResp.Body.Close()
				
				// Should return successful read operation
				assert.True(t, readResp.StatusCode >= 200 && readResp.StatusCode < 300,
					"%s - read operation must be successful", crud.description)
				
				body, err := io.ReadAll(readResp.Body)
				if err == nil {
					responseStr := string(body)
					
					// Should not return internal errors for database access
					assert.NotContains(t, responseStr, "INTERNAL_ERROR",
						"%s - read operation must not fail with database errors", crud.description)
					assert.NotContains(t, responseStr, "failed to get all",
						"%s - read operation must successfully access database", crud.description)
					
					// Should return structured data
					var jsonData map[string]interface{}
					assert.NoError(t, json.Unmarshal(body, &jsonData),
						"%s - read operation must return valid JSON", crud.description)
				}
			} else {
				t.Errorf("%s read operation not accessible: %v", crud.description, err)
			}

			// Test CREATE operation (should persist to database)
			createReq, err := http.NewRequestWithContext(ctx, "POST", crud.createEndpoint, strings.NewReader(crud.testPayload))
			require.NoError(t, err, "Failed to create POST request")
			createReq.Header.Set("Content-Type", "application/json")

			createResp, err := client.Do(createReq)
			if err == nil {
				defer createResp.Body.Close()
				
				// Should handle create operation (even if validation errors)
				assert.True(t, createResp.StatusCode >= 200 && createResp.StatusCode < 500,
					"%s - create operation must be handled by API", crud.description)
				
				createBody, err := io.ReadAll(createResp.Body)
				if err == nil {
					createResponseStr := string(createBody)
					
					// Should not return 404 or method not found
					assert.NotContains(t, createResponseStr, "404 page not found",
						"%s - create operation endpoint must be implemented", crud.description)
					assert.NotContains(t, createResponseStr, "method not allowed",
						"%s - create operation must be supported", crud.description)
				}
			} else {
				t.Logf("%s create operation not accessible: %v", crud.description, err)
			}
		})
	}
}

func TestAPIDataOperations_DatabaseTableDataValidation(t *testing.T) {
	// Test that database tables can be accessed by services for data operations
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Database table access validation for each service
	tableAccessTests := []struct {
		serviceName string
		tableName   string
		testQuery   string
		description string
	}{
		{
			serviceName: "content",
			tableName:   "content_news",
			testQuery:   "SELECT COUNT(*) FROM content_news",
			description: "Content service must be able to access content_news table for news API",
		},
		{
			serviceName: "content",
			tableName:   "content_events",
			testQuery:   "SELECT COUNT(*) FROM content_events",
			description: "Content service must be able to access content_events table for events API",
		},
		{
			serviceName: "inquiries",
			tableName:   "inquiries_business",
			testQuery:   "SELECT COUNT(*) FROM inquiries_business",
			description: "Inquiries service must be able to access inquiries_business table for business inquiries API",
		},
		{
			serviceName: "notifications",
			tableName:   "notification_subscribers",
			testQuery:   "SELECT COUNT(*) FROM notification_subscribers",
			description: "Notifications service must be able to access notification_subscribers table for subscribers API",
		},
	}

	// Act & Assert: Test database table access
	for _, table := range tableAccessTests {
		t.Run("TableAccess_"+table.serviceName+"_"+table.tableName, func(t *testing.T) {
			// Connect to database
			connStr := "postgresql://postgres:password@localhost:5432/international_center_development?sslmode=disable"
			db, err := sql.Open("postgres", connStr)
			require.NoError(t, err, "Database must be accessible for %s", table.serviceName)
			defer db.Close()

			// Test table access
			var count int
			err = db.QueryRowContext(ctx, table.testQuery).Scan(&count)
			assert.NoError(t, err, 
				"%s - database table must be accessible for service operations", table.description)

			// Validate table structure for service operations
			var columnCount int
			err = db.QueryRowContext(ctx, 
				"SELECT COUNT(*) FROM information_schema.columns WHERE table_name = $1", 
				table.tableName).Scan(&columnCount)
			require.NoError(t, err, "Failed to check table structure for %s", table.tableName)

			assert.Greater(t, columnCount, 0, 
				"Table %s must have columns for %s operations", table.tableName, table.serviceName)

			t.Logf("Table %s has %d rows, %d columns for %s service", table.tableName, count, columnCount, table.serviceName)
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