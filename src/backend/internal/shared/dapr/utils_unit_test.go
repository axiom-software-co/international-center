package dapr

import (
	"testing"

	sharedtesting "github.com/axiom-software-co/international-center/src/backend/internal/shared/testing"
	"github.com/stretchr/testify/assert"
)

// RED PHASE - Dapr Utility Functions Tests (40+ test cases)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		setEnv       bool
		expected     string
	}{
		{
			name:         "get existing environment variable",
			key:          "TEST_EXISTING_VAR",
			defaultValue: "default",
			envValue:     "actual-value",
			setEnv:       true,
			expected:     "actual-value",
		},
		{
			name:         "get non-existing environment variable returns default",
			key:          "TEST_NON_EXISTING_VAR",
			defaultValue: "default-value",
			envValue:     "",
			setEnv:       false,
			expected:     "default-value",
		},
		{
			name:         "get environment variable with empty value returns default",
			key:          "TEST_EMPTY_VAR",
			defaultValue: "default-for-empty",
			envValue:     "",
			setEnv:       true,
			expected:     "default-for-empty",
		},
		{
			name:         "get environment variable with whitespace value",
			key:          "TEST_WHITESPACE_VAR",
			defaultValue: "default",
			envValue:     "  whitespace-value  ",
			setEnv:       true,
			expected:     "  whitespace-value  ",
		},
		{
			name:         "get DAPR_APP_ID environment variable",
			key:          "DAPR_APP_ID",
			defaultValue: "international-center",
			envValue:     "custom-app-id",
			setEnv:       true,
			expected:     "custom-app-id",
		},
		{
			name:         "get ENVIRONMENT with production value",
			key:          "ENVIRONMENT",
			defaultValue: "development",
			envValue:     "production",
			setEnv:       true,
			expected:     "production",
		},
		{
			name:         "get DAPR_STATE_STORE_NAME with default",
			key:          "DAPR_STATE_STORE_NAME",
			defaultValue: "statestore-postgresql",
			envValue:     "",
			setEnv:       false,
			expected:     "statestore-postgresql",
		},
		{
			name:         "get DAPR_PUBSUB_NAME with custom value",
			key:          "DAPR_PUBSUB_NAME",
			defaultValue: "pubsub-redis",
			envValue:     "custom-pubsub",
			setEnv:       true,
			expected:     "custom-pubsub",
		},
		{
			name:         "get SECRET_STORE_NAME with default",
			key:          "DAPR_SECRET_STORE_NAME",
			defaultValue: "local-secrets",
			envValue:     "",
			setEnv:       false,
			expected:     "local-secrets",
		},
		{
			name:         "get DAPR_BLOB_BINDING_NAME with staging value",
			key:          "DAPR_BLOB_BINDING_NAME",
			defaultValue: "blob-storage",
			envValue:     "staging-blob-storage",
			setEnv:       true,
			expected:     "staging-blob-storage",
		},
		{
			name:         "get configuration store name with default",
			key:          "DAPR_CONFIGURATION_STORE_NAME",
			defaultValue: "configstore",
			envValue:     "",
			setEnv:       false,
			expected:     "configstore",
		},
		{
			name:         "get environment variable with special characters",
			key:          "TEST_SPECIAL_CHARS",
			defaultValue: "default",
			envValue:     "value-with-special!@#$%^&*()_+",
			setEnv:       true,
			expected:     "value-with-special!@#$%^&*()_+",
		},
		{
			name:         "get environment variable with numbers",
			key:          "TEST_NUMERIC_VALUE",
			defaultValue: "default",
			envValue:     "12345",
			setEnv:       true,
			expected:     "12345",
		},
		{
			name:         "get environment variable with mixed alphanumeric",
			key:          "TEST_MIXED_VALUE",
			defaultValue: "default",
			envValue:     "abc123def456",
			setEnv:       true,
			expected:     "abc123def456",
		},
		{
			name:         "get environment variable with path value",
			key:          "TEST_PATH_VALUE",
			defaultValue: "/default/path",
			envValue:     "/custom/path/to/resource",
			setEnv:       true,
			expected:     "/custom/path/to/resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			if tt.setEnv {
				t.Setenv(tt.key, tt.envValue)
			}

			// Act
			result := getEnv(tt.key, tt.defaultValue)

			// Assert
			assert.Equal(t, tt.expected, result)

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int
		envValue     string
		setEnv       bool
		expected     int
	}{
		{
			name:         "get existing integer environment variable",
			key:          "TEST_INT_EXISTING",
			defaultValue: 100,
			envValue:     "42",
			setEnv:       true,
			expected:     42,
		},
		{
			name:         "get non-existing integer environment variable returns default",
			key:          "TEST_INT_NON_EXISTING",
			defaultValue: 500,
			envValue:     "",
			setEnv:       false,
			expected:     500,
		},
		{
			name:         "get environment variable with empty value returns default",
			key:          "TEST_INT_EMPTY",
			defaultValue: 200,
			envValue:     "",
			setEnv:       true,
			expected:     200,
		},
		{
			name:         "get environment variable with zero value",
			key:          "TEST_INT_ZERO",
			defaultValue: 300,
			envValue:     "0",
			setEnv:       true,
			expected:     300, // Zero is not > 0, so should return default
		},
		{
			name:         "get environment variable with negative value returns default",
			key:          "TEST_INT_NEGATIVE",
			defaultValue: 400,
			envValue:     "-50",
			setEnv:       true,
			expected:     400, // Negative parsed as 0, not > 0, so should return default
		},
		{
			name:         "get environment variable with invalid integer returns default",
			key:          "TEST_INT_INVALID",
			defaultValue: 600,
			envValue:     "not-a-number",
			setEnv:       true,
			expected:     600,
		},
		{
			name:         "get environment variable with large integer",
			key:          "TEST_INT_LARGE",
			defaultValue: 1,
			envValue:     "999999",
			setEnv:       true,
			expected:     999999,
		},
		{
			name:         "get DAPR_HTTP_PORT with default",
			key:          "DAPR_HTTP_PORT",
			defaultValue: 3500,
			envValue:     "",
			setEnv:       false,
			expected:     3500,
		},
		{
			name:         "get DAPR_GRPC_PORT with custom value",
			key:          "DAPR_GRPC_PORT",
			defaultValue: 50001,
			envValue:     "60001",
			setEnv:       true,
			expected:     60001,
		},
		{
			name:         "get API_PORT with default",
			key:          "API_PORT",
			defaultValue: 8080,
			envValue:     "",
			setEnv:       false,
			expected:     8080,
		},
		{
			name:         "get DATABASE_MAX_CONNECTIONS with custom value",
			key:          "DATABASE_MAX_CONNECTIONS",
			defaultValue: 25,
			envValue:     "50",
			setEnv:       true,
			expected:     50,
		},
		{
			name:         "get API_TIMEOUT with default",
			key:          "API_TIMEOUT",
			defaultValue: 30,
			envValue:     "",
			setEnv:       false,
			expected:     30,
		},
		{
			name:         "get RATE_LIMIT with custom value",
			key:          "RATE_LIMIT",
			defaultValue: 1000,
			envValue:     "2000",
			setEnv:       true,
			expected:     2000,
		},
		{
			name:         "get environment variable with mixed characters returns default",
			key:          "TEST_INT_MIXED",
			defaultValue: 999,
			envValue:     "123abc456",
			setEnv:       true,
			expected:     999, // Should fail parsing and return default
		},
		{
			name:         "get environment variable with leading zeros",
			key:          "TEST_INT_LEADING_ZEROS",
			defaultValue: 1,
			envValue:     "00123",
			setEnv:       true,
			expected:     123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			if tt.setEnv {
				t.Setenv(tt.key, tt.envValue)
			}

			// Act
			result := getEnvInt(tt.key, tt.defaultValue)

			// Assert
			assert.Equal(t, tt.expected, result)

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "parse valid positive integer",
			input:    "123",
			expected: 123,
		},
		{
			name:     "parse zero",
			input:    "0",
			expected: 0,
		},
		{
			name:     "parse single digit",
			input:    "5",
			expected: 5,
		},
		{
			name:     "parse large number",
			input:    "999999",
			expected: 999999,
		},
		{
			name:     "parse number with leading zeros",
			input:    "00123",
			expected: 123,
		},
		{
			name:     "parse maximum safe integer representation",
			input:    "2147483647", // Max int32
			expected: 2147483647,
		},
		{
			name:     "parse empty string returns zero",
			input:    "",
			expected: 0,
		},
		{
			name:     "parse string with letters returns zero",
			input:    "abc",
			expected: 0,
		},
		{
			name:     "parse string with mixed characters returns zero",
			input:    "123abc",
			expected: 0, // Should stop at first non-digit
		},
		{
			name:     "parse string with special characters returns zero",
			input:    "123!@#",
			expected: 0, // Should stop at first non-digit
		},
		{
			name:     "parse negative sign returns zero",
			input:    "-123",
			expected: 0, // Should stop at first non-digit (-)
		},
		{
			name:     "parse plus sign returns zero",
			input:    "+123",
			expected: 0, // Should stop at first non-digit (+)
		},
		{
			name:     "parse decimal number returns zero",
			input:    "123.45",
			expected: 0, // Should stop at first non-digit (.)
		},
		{
			name:     "parse whitespace returns zero",
			input:    "   ",
			expected: 0,
		},
		{
			name:     "parse string with leading whitespace returns zero",
			input:    " 123",
			expected: 0, // Should stop at first non-digit (space)
		},
		{
			name:     "parse hexadecimal returns zero",
			input:    "0xFF",
			expected: 0, // Should stop at first non-digit after 0
		},
		{
			name:     "parse scientific notation returns zero",
			input:    "1e5",
			expected: 1, // Should parse "1" then stop at 'e'
		},
		{
			name:     "parse very large number that could overflow",
			input:    "999999999",
			expected: 999999999, // Large but within int bounds
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			// Act
			result := parseInt(tt.input)

			// Assert
			assert.Equal(t, tt.expected, result)

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestUtilityFunctions_Integration(t *testing.T) {
	tests := []struct {
		name          string
		envVars       map[string]string
		testOperation func(*testing.T)
	}{
		{
			name: "integration test with multiple environment variables",
			envVars: map[string]string{
				"DAPR_APP_ID":       "test-app",
				"DAPR_HTTP_PORT":    "3500",
				"DAPR_GRPC_PORT":    "50001",
				"API_PORT":          "8080",
				"ENVIRONMENT":       "test",
			},
			testOperation: func(t *testing.T) {
				// Test string environment variables
				appID := getEnv("DAPR_APP_ID", "default-app")
				assert.Equal(t, "test-app", appID)
				
				environment := getEnv("ENVIRONMENT", "development")
				assert.Equal(t, "test", environment)
				
				// Test integer environment variables
				daprHTTPPort := getEnvInt("DAPR_HTTP_PORT", 3000)
				assert.Equal(t, 3500, daprHTTPPort)
				
				daprGRPCPort := getEnvInt("DAPR_GRPC_PORT", 50000)
				assert.Equal(t, 50001, daprGRPCPort)
				
				apiPort := getEnvInt("API_PORT", 8000)
				assert.Equal(t, 8080, apiPort)
				
				// Test non-existing environment variables with defaults
				nonExisting := getEnv("NON_EXISTING_VAR", "default")
				assert.Equal(t, "default", nonExisting)
				
				nonExistingInt := getEnvInt("NON_EXISTING_INT", 999)
				assert.Equal(t, 999, nonExistingInt)
			},
		},
		{
			name: "integration test with production-like configuration",
			envVars: map[string]string{
				"ENVIRONMENT":                     "production",
				"DAPR_APP_ID":                     "international-center-prod",
				"DAPR_STATE_STORE_NAME":           "statestore-postgresql-prod",
				"DAPR_PUBSUB_NAME":                "pubsub-redis-prod",
				"DAPR_SECRET_STORE_NAME":          "azure-keyvault",
				"DAPR_BLOB_BINDING_NAME":          "azure-blob-storage",
				"DAPR_CONFIGURATION_STORE_NAME":   "azure-config-store",
				"DATABASE_MAX_CONNECTIONS":        "100",
				"API_PORT":                        "8443",
				"RATE_LIMIT":                      "5000",
			},
			testOperation: func(t *testing.T) {
				// Test production environment configuration
				env := getEnv("ENVIRONMENT", "development")
				assert.Equal(t, "production", env)
				
				appID := getEnv("DAPR_APP_ID", "international-center")
				assert.Equal(t, "international-center-prod", appID)
				
				stateStore := getEnv("DAPR_STATE_STORE_NAME", "statestore-postgresql")
				assert.Equal(t, "statestore-postgresql-prod", stateStore)
				
				pubsub := getEnv("DAPR_PUBSUB_NAME", "pubsub-redis")
				assert.Equal(t, "pubsub-redis-prod", pubsub)
				
				secretStore := getEnv("DAPR_SECRET_STORE_NAME", "local-secrets")
				assert.Equal(t, "azure-keyvault", secretStore)
				
				blobBinding := getEnv("DAPR_BLOB_BINDING_NAME", "blob-storage")
				assert.Equal(t, "azure-blob-storage", blobBinding)
				
				configStore := getEnv("DAPR_CONFIGURATION_STORE_NAME", "configstore")
				assert.Equal(t, "azure-config-store", configStore)
				
				// Test integer configurations
				maxConnections := getEnvInt("DATABASE_MAX_CONNECTIONS", 25)
				assert.Equal(t, 100, maxConnections)
				
				apiPort := getEnvInt("API_PORT", 8080)
				assert.Equal(t, 8443, apiPort)
				
				rateLimit := getEnvInt("RATE_LIMIT", 1000)
				assert.Equal(t, 5000, rateLimit)
			},
		},
		{
			name: "integration test with staging configuration",
			envVars: map[string]string{
				"ENVIRONMENT":             "staging",
				"DAPR_APP_ID":             "international-center-staging",
				"DAPR_HTTP_PORT":          "3501",
				"DAPR_GRPC_PORT":          "50002",
				"DATABASE_MAX_CONNECTIONS": "50",
				"API_READ_TIMEOUT":        "20",
				"API_WRITE_TIMEOUT":       "25",
			},
			testOperation: func(t *testing.T) {
				// Test staging environment configuration
				env := getEnv("ENVIRONMENT", "development")
				assert.Equal(t, "staging", env)
				
				appID := getEnv("DAPR_APP_ID", "international-center")
				assert.Equal(t, "international-center-staging", appID)
				
				httpPort := getEnvInt("DAPR_HTTP_PORT", 3500)
				assert.Equal(t, 3501, httpPort)
				
				grpcPort := getEnvInt("DAPR_GRPC_PORT", 50001)
				assert.Equal(t, 50002, grpcPort)
				
				maxConnections := getEnvInt("DATABASE_MAX_CONNECTIONS", 25)
				assert.Equal(t, 50, maxConnections)
				
				readTimeout := getEnvInt("API_READ_TIMEOUT", 15)
				assert.Equal(t, 20, readTimeout)
				
				writeTimeout := getEnvInt("API_WRITE_TIMEOUT", 15)
				assert.Equal(t, 25, writeTimeout)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			// Act & Assert
			tt.testOperation(t)

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestUtilityFunctions_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		testOperation func(*testing.T)
	}{
		{
			name: "test empty and nil inputs",
			testOperation: func(t *testing.T) {
				// Test getEnv with empty key
				result := getEnv("", "default")
				assert.Equal(t, "default", result)
				
				// Test getEnvInt with empty key
				intResult := getEnvInt("", 42)
				assert.Equal(t, 42, intResult)
				
				// Test parseInt with empty string
				parseResult := parseInt("")
				assert.Equal(t, 0, parseResult)
			},
		},
		{
			name: "test very long strings",
			testOperation: func(t *testing.T) {
				longString := "this-is-a-very-long-environment-variable-value-that-exceeds-normal-length-expectations-and-tests-boundary-conditions"
				longKey := "VERY_LONG_ENVIRONMENT_VARIABLE_KEY_NAME"
				
				// Set and test long environment variable
				t.Setenv(longKey, longString)
				result := getEnv(longKey, "default")
				assert.Equal(t, longString, result)
				
				// Test parsing very long number string
				longNumberString := "123456789012345678901234567890"
				parseResult := parseInt(longNumberString)
				// This might overflow, but we're testing the function behavior
				assert.Greater(t, parseResult, 0)
			},
		},
		{
			name: "test unicode and special characters",
			testOperation: func(t *testing.T) {
				// Test environment variable with unicode characters
				unicodeValue := "测试-value-тест-🚀"
				t.Setenv("UNICODE_TEST", unicodeValue)
				result := getEnv("UNICODE_TEST", "default")
				assert.Equal(t, unicodeValue, result)
				
				// Test environment variable with special characters
				specialValue := "value!@#$%^&*()_+-=[]{}|;':\",./<>?"
				t.Setenv("SPECIAL_CHARS_TEST", specialValue)
				specialResult := getEnv("SPECIAL_CHARS_TEST", "default")
				assert.Equal(t, specialValue, specialResult)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			// Act & Assert
			tt.testOperation(t)

			_ = ctx // Use context to avoid linting issues
		})
	}
}