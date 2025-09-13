package integration

import (
	"context"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	sharedValidation "github.com/axiom-software-co/international-center/src/public-website/deployment/test/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED PHASE: Service Environment Configuration Tests
// These tests validate that all services have required environment variables and configuration
// for their specific domain functionality

func TestServiceEnvironmentConfiguration_RequiredVariables(t *testing.T) {
	// This test requires complete environment health - enforcing axiom rule
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Define required environment variables for each consolidated service
	serviceEnvironmentRequirements := map[string]struct {
		serviceName        string
		requiredEnvVars    []string
		conditionalEnvVars map[string]string
		description        string
	}{
		"public-gateway": {
			serviceName: "public-gateway",
			requiredEnvVars: []string{
				"DAPR_APP_ID",
				"PORT",
				"PUBLIC_GATEWAY_PORT",
				"DATABASE_URL",
			},
			conditionalEnvVars: map[string]string{
				"ENVIRONMENT": "development",
			},
			description: "Public gateway must have gateway-specific and database configuration",
		},
		"admin-gateway": {
			serviceName: "admin-gateway",
			requiredEnvVars: []string{
				"DAPR_APP_ID",
				"PORT", 
				"ADMIN_GATEWAY_PORT",
				"DATABASE_URL",
			},
			conditionalEnvVars: map[string]string{
				"ENVIRONMENT": "development",
			},
			description: "Admin gateway must have gateway-specific and database configuration",
		},
		"content": {
			serviceName: "content",
			requiredEnvVars: []string{
				"DAPR_APP_ID",
				"PORT",
				"DATABASE_URL",
			},
			conditionalEnvVars: map[string]string{
				"ENVIRONMENT": "development",
			},
			description: "Content service must have database configuration for content management",
		},
		"inquiries": {
			serviceName: "inquiries",
			requiredEnvVars: []string{
				"DAPR_APP_ID",
				"PORT",
				"DATABASE_URL",
			},
			conditionalEnvVars: map[string]string{
				"ENVIRONMENT": "development",
			},
			description: "Inquiries service must have database configuration for inquiry management",
		},
		"notifications": {
			serviceName: "notifications",
			requiredEnvVars: []string{
				"DAPR_APP_ID",
				"PORT",
				"DATABASE_URL",
				"RABBITMQ_URL",
			},
			conditionalEnvVars: map[string]string{
				"ENVIRONMENT": "development",
			},
			description: "Notifications service must have database and messaging configuration",
		},
	}

	// Act & Assert: Validate environment configuration for each service
	for serviceName, requirements := range serviceEnvironmentRequirements {
		t.Run("EnvConfig_"+serviceName, func(t *testing.T) {
			// Check if service container is running
			serviceCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+requirements.serviceName, "--format", "{{.Names}}")
			serviceOutput, err := serviceCmd.Output()
			require.NoError(t, err, "Failed to check service %s status", requirements.serviceName)

			runningServices := strings.TrimSpace(string(serviceOutput))

			if strings.Contains(runningServices, requirements.serviceName) {
				// Service is running - validate environment variables
				envCmd := exec.CommandContext(ctx, "podman", "exec", requirements.serviceName, "env")
				envOutput, err := envCmd.Output()
				require.NoError(t, err, "Failed to get environment for %s", requirements.serviceName)

				envVars := string(envOutput)
				
				// Validate required environment variables
				for _, requiredVar := range requirements.requiredEnvVars {
					assert.Contains(t, envVars, requiredVar+"=", 
						"%s must have %s environment variable", requirements.description, requiredVar)
				}

				// Validate conditional environment variables
				for condVar, expectedValue := range requirements.conditionalEnvVars {
					expectedEnvVar := condVar + "=" + expectedValue
					assert.Contains(t, envVars, expectedEnvVar,
						"%s must have %s for proper configuration", requirements.description, expectedEnvVar)
				}

				// Check service logs for configuration errors
				logsCmd := exec.CommandContext(ctx, "podman", "logs", "--tail", "20", requirements.serviceName)
				logsOutput, err := logsCmd.Output()
				if err == nil {
					logs := string(logsOutput)
					
					// Service should not fail with configuration errors
					assert.NotContains(t, logs, "environment variable is required",
						"%s must not fail with missing environment variable errors", requirements.description)
					assert.NotContains(t, logs, "Invalid configuration",
						"%s must not fail with configuration validation errors", requirements.description)
					assert.NotContains(t, logs, "database config validation failed",
						"%s must not fail with database configuration errors", requirements.description)
				}
			} else {
				t.Logf("Service %s not running - cannot validate environment configuration", requirements.serviceName)
			}
		})
	}
}

func TestServiceEnvironmentConfiguration_DatabaseConnectivity(t *testing.T) {
	// Test that services requiring database access have proper database configuration
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Services that require database connectivity
	databaseDependentServices := []struct {
		serviceName    string
		description    string
		requiresDB     bool
	}{
		{"content", "Content service requires database for content storage", true},
		{"inquiries", "Inquiries service requires database for inquiry storage", true},
		{"notifications", "Notifications service requires database for notification tracking", true},
		{"public-gateway", "Public gateway requires database for caching and state", true},
		{"admin-gateway", "Admin gateway requires database for admin operations", true},
	}

	// Act & Assert: Validate database configuration
	for _, service := range databaseDependentServices {
		t.Run("DatabaseConfig_"+service.serviceName, func(t *testing.T) {
			// Check if service is running
			serviceCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+service.serviceName, "--format", "{{.Names}}")
			serviceOutput, err := serviceCmd.Output()
			require.NoError(t, err, "Failed to check service %s", service.serviceName)

			runningServices := strings.TrimSpace(string(serviceOutput))

			if strings.Contains(runningServices, service.serviceName) && service.requiresDB {
				// Service is running and requires database - validate configuration
				envCmd := exec.CommandContext(ctx, "podman", "exec", service.serviceName, "env")
				envOutput, err := envCmd.Output()
				require.NoError(t, err, "Failed to get environment for %s", service.serviceName)

				envVars := string(envOutput)
				
				// Service must have database URL configuration
				assert.Contains(t, envVars, "DATABASE_URL=", 
					"%s must have DATABASE_URL environment variable", service.description)

				// Validate database URL points to deployed PostgreSQL
				assert.Contains(t, envVars, "postgresql://", 
					"%s DATABASE_URL must be PostgreSQL connection string", service.description)
				assert.Contains(t, envVars, "localhost:5432", 
					"%s DATABASE_URL must point to deployed PostgreSQL container", service.description)

			} else if !service.requiresDB {
				t.Logf("Service %s does not require database configuration", service.serviceName)
			} else {
				t.Logf("Service %s not running - cannot validate database configuration", service.serviceName)
			}
		})
	}
}

func TestServiceEnvironmentConfiguration_ServiceSpecificSettings(t *testing.T) {
	// Test service-specific configuration requirements beyond basic Dapr and database
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Service-specific configuration requirements
	serviceSpecificRequirements := map[string]struct {
		serviceName      string
		specificSettings map[string]string
		description      string
	}{
		"public-gateway": {
			serviceName: "public-gateway",
			specificSettings: map[string]string{
				"PUBLIC_GATEWAY_PORT": "9001",
				"GATEWAY_TYPE":        "public",
			},
			description: "Public gateway requires gateway-specific port and type configuration",
		},
		"admin-gateway": {
			serviceName: "admin-gateway", 
			specificSettings: map[string]string{
				"ADMIN_GATEWAY_PORT": "9000",
				"GATEWAY_TYPE":       "admin",
			},
			description: "Admin gateway requires gateway-specific port and type configuration",
		},
		"notifications": {
			serviceName: "notifications",
			specificSettings: map[string]string{
				"RABBITMQ_URL": "amqp://guest:guest@rabbitmq:5672/",
			},
			description: "Notifications service requires messaging configuration for RabbitMQ",
		},
	}

	// Act & Assert: Validate service-specific configuration
	for serviceName, requirements := range serviceSpecificRequirements {
		t.Run("SpecificConfig_"+serviceName, func(t *testing.T) {
			// Check if service is running
			serviceCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+requirements.serviceName, "--format", "{{.Names}}")
			serviceOutput, err := serviceCmd.Output()
			require.NoError(t, err, "Failed to check service %s", requirements.serviceName)

			runningServices := strings.TrimSpace(string(serviceOutput))

			if strings.Contains(runningServices, requirements.serviceName) {
				// Service is running - validate specific settings
				envCmd := exec.CommandContext(ctx, "podman", "exec", requirements.serviceName, "env")
				envOutput, err := envCmd.Output()
				require.NoError(t, err, "Failed to get environment for %s", requirements.serviceName)

				envVars := string(envOutput)
				
				// Validate service-specific settings
				for setting, expectedValue := range requirements.specificSettings {
					expectedEnvVar := setting + "=" + expectedValue
					assert.Contains(t, envVars, expectedEnvVar,
						"%s must have %s", requirements.description, expectedEnvVar)
				}
			} else {
				t.Logf("Service %s not running - cannot validate specific configuration", requirements.serviceName)
			}
		})
	}
}

func TestServiceEnvironmentConfiguration_ServiceStartupReliability(t *testing.T) {
	// Test that services start reliably without configuration errors
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// All services that should start without configuration errors
	consolidatedServices := []string{
		"public-gateway",
		"admin-gateway",
		"content",
		"inquiries", 
		"notifications",
	}

	// Act & Assert: Validate service startup reliability
	for _, serviceName := range consolidatedServices {
		t.Run("StartupReliability_"+serviceName, func(t *testing.T) {
			// Check service status
			statusCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+serviceName, "--format", "{{.Status}}")
			statusOutput, err := statusCmd.Output()
			require.NoError(t, err, "Failed to check service %s status", serviceName)

			status := strings.TrimSpace(string(statusOutput))

			if status != "" {
				// Service container exists - validate it's not exited due to config errors
				assert.Contains(t, status, "Up", 
					"Service %s must be running with proper configuration", serviceName)
				assert.NotContains(t, status, "Exited", 
					"Service %s must not exit due to configuration errors", serviceName)

				// If running, test health endpoint accessibility
				if strings.Contains(status, "Up") {
					// Determine service port based on service name
					servicePort := "8080" // Default internal port
					switch serviceName {
					case "public-gateway":
						servicePort = "9001"
					case "admin-gateway":
						servicePort = "9000" 
					case "content":
						servicePort = "3001"
					case "inquiries":
						servicePort = "3101"
					case "notifications":
						servicePort = "3201"
					}

					healthURL := "http://localhost:" + servicePort + "/health"
					client := &http.Client{Timeout: 5 * time.Second}
					
					healthReq, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
					require.NoError(t, err, "Failed to create health request")

					healthResp, err := client.Do(healthReq)
					if err == nil {
						defer healthResp.Body.Close()
						assert.True(t, healthResp.StatusCode >= 200 && healthResp.StatusCode < 300,
							"Service %s health endpoint must be accessible with proper configuration", serviceName)
					} else {
						t.Errorf("Service %s health endpoint not accessible - configuration issue: %v", serviceName, err)
					}
				}
			} else {
				t.Logf("Service %s not deployed yet (expected for incomplete configuration)", serviceName)
			}
		})
	}
}

