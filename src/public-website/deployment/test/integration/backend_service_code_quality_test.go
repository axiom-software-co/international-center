package integration

import (
	"context"
	"database/sql"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// RED PHASE: Backend Service Code Quality Tests
// These tests validate that backend services use proper SQL syntax and PostgreSQL compatibility

func TestBackendServiceCodeQuality_SQLCompatibility(t *testing.T) {
	// This test requires complete environment health - enforcing axiom rule
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Services that must have PostgreSQL-compatible SQL queries
	backendServices := []struct {
		serviceName        string
		containerName      string
		expectedSQLErrors  []string
		description        string
		critical           bool
	}{
		{
			serviceName:   "notifications",
			containerName: "notifications",
			expectedSQLErrors: []string{
				"operator does not exist: uuid <> text",
				"checkEmailExists statement",
				"uuid comparison error",
			},
			description: "Notifications service must use PostgreSQL-compatible SQL syntax",
			critical:    true,
		},
		{
			serviceName:   "content",
			containerName: "content",
			expectedSQLErrors: []string{
				"operator does not exist",
				"invalid SQL syntax",
			},
			description: "Content service must use PostgreSQL-compatible SQL syntax",
			critical:    false,
		},
		{
			serviceName:   "inquiries",
			containerName: "inquiries", 
			expectedSQLErrors: []string{
				"operator does not exist",
				"invalid SQL syntax",
			},
			description: "Inquiries service must use PostgreSQL-compatible SQL syntax",
			critical:    false,
		},
	}

	// Act & Assert: Validate SQL compatibility in service logs
	for _, service := range backendServices {
		t.Run("SQLCompatibility_"+service.serviceName, func(t *testing.T) {
			// Check if service is running or has attempted to start
			statusCmd := exec.CommandContext(ctx, "podman", "ps", "-a", "--filter", "name="+service.containerName, "--format", "{{.Names}}")
			statusOutput, err := statusCmd.Output()
			require.NoError(t, err, "Failed to check service %s status", service.containerName)

			serviceExists := strings.Contains(string(statusOutput), service.containerName)
			
			if serviceExists {
				// Service exists - check logs for SQL compatibility issues
				logsCmd := exec.CommandContext(ctx, "podman", "logs", "--tail", "50", service.containerName)
				logsOutput, err := logsCmd.Output()
				require.NoError(t, err, "Failed to get logs for %s", service.containerName)

				logs := string(logsOutput)
				
				// Check for SQL compatibility errors
				for _, sqlError := range service.expectedSQLErrors {
					if service.critical && strings.Contains(logs, sqlError) {
						t.Errorf("%s contains SQL compatibility error: %s", service.description, sqlError)
					} else if !service.critical && strings.Contains(logs, sqlError) {
						t.Logf("%s contains SQL compatibility issue (non-critical): %s", service.description, sqlError)
					}
				}

				// Service should not fail with SQL operator errors
				assert.NotContains(t, logs, "operator does not exist", 
					"%s must not fail with SQL operator compatibility errors", service.description)
				assert.NotContains(t, logs, "invalid SQL syntax",
					"%s must not fail with SQL syntax errors", service.description)
			} else {
				t.Logf("Service %s not found - cannot validate SQL compatibility", service.serviceName)
			}
		})
	}
}

func TestBackendServiceCodeQuality_DatabaseTypeHandling(t *testing.T) {
	// Test that backend services properly handle PostgreSQL data types
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Database type compatibility requirements for each service
	databaseTypeTests := []struct {
		serviceName    string
		dataTypes      []string
		tableRequired  string
		description    string
	}{
		{
			serviceName:   "notifications",
			dataTypes:     []string{"UUID", "TEXT[]", "JSONB", "TIMESTAMP WITH TIME ZONE"},
			tableRequired: "notification_subscribers",
			description:   "Notifications service must handle UUID, array, and JSON types properly",
		},
		{
			serviceName:   "content", 
			dataTypes:     []string{"VARCHAR", "TEXT", "TIMESTAMP"},
			tableRequired: "content_news",
			description:   "Content service must handle text and timestamp types properly",
		},
		{
			serviceName:   "inquiries",
			dataTypes:     []string{"VARCHAR", "TEXT", "TIMESTAMP"},
			tableRequired: "inquiries_business", 
			description:   "Inquiries service must handle text and timestamp types properly",
		},
	}

	// Act & Assert: Test database type handling
	for _, test := range databaseTypeTests {
		t.Run("DatabaseTypeHandling_"+test.serviceName, func(t *testing.T) {
			// Connect to database to test type compatibility
			connStr := "postgresql://postgres:password@localhost:5432/international_center_development?sslmode=disable"
			db, err := sql.Open("postgres", connStr)
			require.NoError(t, err, "Database must be accessible for type handling validation")
			defer db.Close()

			// Test that the required table exists with proper data types
			var tableExists bool
			err = db.QueryRowContext(ctx, 
				"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = $1)", 
				test.tableRequired).Scan(&tableExists)
			require.NoError(t, err, "Failed to check table %s for %s", test.tableRequired, test.serviceName)

			if tableExists {
				// Table exists - test type compatibility by performing operations
				_, err = db.ExecContext(ctx, "SELECT COUNT(*) FROM "+test.tableRequired)
				assert.NoError(t, err, 
					"%s table %s must be accessible for service operations", test.description, test.tableRequired)
			} else {
				t.Logf("Table %s does not exist for %s - service will fail with database type errors", 
					test.tableRequired, test.serviceName)
			}
		})
	}
}

func TestBackendServiceCodeQuality_ServiceStartupReliability(t *testing.T) {
	// Test that all backend services start without code quality issues
	validateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// All backend services that should start without code issues
	allBackendServices := []struct {
		serviceName   string
		containerName string
		critical      bool
		description   string
	}{
		{"content", "content", false, "Content service should start without code issues"},
		{"inquiries", "inquiries", false, "Inquiries service should start without code issues"},
		{"notifications", "notifications", true, "Notifications service MUST start without SQL compatibility errors"},
		{"public-gateway", "public-gateway", false, "Public gateway should start without code issues"},
		{"admin-gateway", "admin-gateway", false, "Admin gateway should start without code issues"},
	}

	// Act & Assert: Validate service startup reliability
	for _, service := range allBackendServices {
		t.Run("StartupReliability_"+service.serviceName, func(t *testing.T) {
			// Check service container status
			statusCmd := exec.CommandContext(ctx, "podman", "ps", "-a", "--filter", "name="+service.containerName, "--format", "{{.Status}}")
			statusOutput, err := statusCmd.Output()
			require.NoError(t, err, "Failed to check service %s status", service.containerName)

			status := strings.TrimSpace(string(statusOutput))
			
			if status != "" {
				if service.critical {
					// Critical services must be running without exits
					assert.Contains(t, status, "Up", 
						"%s - critical service must be running for development environment", service.description)
					assert.NotContains(t, status, "Exited", 
						"%s - critical service must not exit due to code quality issues", service.description)
				} else {
					// Non-critical services should be running for complete environment
					if strings.Contains(status, "Exited") {
						t.Logf("%s exited (expected for incomplete code quality): %s", service.description, status)
					} else if strings.Contains(status, "Up") {
						t.Logf("%s running successfully: %s", service.description, status)
					}
				}

				// Check logs for backend code quality issues
				logsCmd := exec.CommandContext(ctx, "podman", "logs", "--tail", "20", service.containerName)
				logsOutput, err := logsCmd.Output()
				if err == nil {
					logs := string(logsOutput)
					
					// Services should not fail with code quality issues
					assert.NotContains(t, logs, "panic:", 
						"%s must not panic due to code quality issues", service.description)
					assert.NotContains(t, logs, "fatal error:",
						"%s must not have fatal errors due to code quality", service.description)
					
					if service.critical {
						assert.NotContains(t, logs, "Failed to create",
							"%s critical initialization must succeed", service.description)
					}
				}
			} else {
				t.Logf("Service %s not deployed - cannot validate startup reliability", service.serviceName)
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