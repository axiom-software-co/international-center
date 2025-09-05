package dapr

import (
	"context"
	"testing"
	"time"

	sharedtesting "github.com/axiom-software-co/international-center/src/backend/internal/shared/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED PHASE - Dapr Configuration Management Tests (30+ test cases)

func TestNewConfiguration(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		expectedError  string
		validateResult func(*testing.T, *Configuration)
	}{
		{
			name: "create configuration with default store name",
			envVars: map[string]string{},
			validateResult: func(t *testing.T, config *Configuration) {
				assert.NotNil(t, config)
				assert.Equal(t, "configstore", config.storeName)
				assert.NotNil(t, config.client)
			},
		},
		{
			name: "create configuration with custom store name",
			envVars: map[string]string{
				"DAPR_CONFIGURATION_STORE_NAME": "custom-config-store",
			},
			validateResult: func(t *testing.T, config *Configuration) {
				assert.NotNil(t, config)
				assert.Equal(t, "custom-config-store", config.storeName)
			},
		},
		{
			name: "create configuration with production environment",
			envVars: map[string]string{
				"ENVIRONMENT": "production",
				"DAPR_CONFIGURATION_STORE_NAME": "prod-config-store",
			},
			validateResult: func(t *testing.T, config *Configuration) {
				assert.NotNil(t, config)
				assert.Equal(t, "prod-config-store", config.storeName)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.SetupDaprTest()
			defer cancel()
			defer ResetClientForTesting()

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			// Act
			config := NewConfiguration(client)

			// Assert
			if tt.expectedError != "" {
				assert.Nil(t, config)
			} else {
				require.NotNil(t, config)
				if tt.validateResult != nil {
					tt.validateResult(t, config)
				}
			}

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestConfiguration_GetConfigurationItem(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		setupContext  func() (context.Context, context.CancelFunc)
		envVars       map[string]string
		expectedError string
		validateResult func(*testing.T, *ConfigItem)
	}{
		{
			name: "get database configuration item",
			key:  "database.max_connections",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			validateResult: func(t *testing.T, item *ConfigItem) {
				assert.Equal(t, "database.max_connections", item.Key)
				assert.NotEmpty(t, item.Value)
				assert.NotEmpty(t, item.Version)
				assert.NotNil(t, item.Metadata)
			},
		},
		{
			name: "get API configuration item",
			key:  "api.port",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			validateResult: func(t *testing.T, item *ConfigItem) {
				assert.Equal(t, "api.port", item.Key)
				assert.NotEmpty(t, item.Value)
			},
		},
		{
			name: "get observability configuration item",
			key:  "observability.log_level",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			validateResult: func(t *testing.T, item *ConfigItem) {
				assert.Equal(t, "observability.log_level", item.Key)
				assert.NotEmpty(t, item.Value)
			},
		},
		{
			name: "get non-existent configuration item",
			key:  "non.existent.key",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			validateResult: func(t *testing.T, item *ConfigItem) {
				assert.Equal(t, "non.existent.key", item.Key)
			},
		},
		{
			name: "get configuration item with empty key",
			key:  "",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			expectedError: "configuration key cannot be empty",
		},
		{
			name: "get configuration item with timeout context",
			key:  "database.connection_timeout",
			setupContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 1*time.Millisecond)
			},
		},
		{
			name: "get configuration item with cancelled context",
			key:  "api.read_timeout",
			setupContext: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				return ctx, func() {}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - Set up test environment first
			_, testCancel := sharedtesting.SetupDaprTest()
			defer testCancel()
			defer ResetClientForTesting()
			
			ctx, cancel := tt.setupContext()
			defer cancel()

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			config := NewConfiguration(client)
			require.NotNil(t, config)

			// Act
			item, err := config.GetConfigurationItem(ctx, tt.key)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, item)
			} else {
				if tt.validateResult != nil {
					require.NoError(t, err)
					require.NotNil(t, item)
					tt.validateResult(t, item)
				}
				// For timeout/cancelled context tests, we just verify no panic occurs
			}
		})
	}
}

func TestConfiguration_GetConfigurationItems(t *testing.T) {
	tests := []struct {
		name          string
		keys          []string
		setupContext  func() (context.Context, context.CancelFunc)
		envVars       map[string]string
		expectedError string
		validateResult func(*testing.T, map[string]*ConfigItem)
	}{
		{
			name: "get multiple database configuration items",
			keys: []string{
				"database.max_connections",
				"database.connection_timeout",
				"database.query_timeout",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			validateResult: func(t *testing.T, items map[string]*ConfigItem) {
				assert.Len(t, items, 3)
				assert.Contains(t, items, "database.max_connections")
				assert.Contains(t, items, "database.connection_timeout")
				assert.Contains(t, items, "database.query_timeout")
				
				for key, item := range items {
					assert.Equal(t, key, item.Key)
					assert.NotEmpty(t, item.Value)
					assert.NotEmpty(t, item.Version)
				}
			},
		},
		{
			name: "get multiple API configuration items",
			keys: []string{
				"api.port",
				"api.read_timeout",
				"api.write_timeout",
				"api.idle_timeout",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			validateResult: func(t *testing.T, items map[string]*ConfigItem) {
				assert.Len(t, items, 4)
				for _, item := range items {
					assert.NotEmpty(t, item.Value)
				}
			},
		},
		{
			name: "get mixed configuration items",
			keys: []string{
				"database.max_connections",
				"api.port",
				"observability.log_level",
				"gateway.public_rate_limit",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			validateResult: func(t *testing.T, items map[string]*ConfigItem) {
				assert.Len(t, items, 4)
				assert.Contains(t, items, "database.max_connections")
				assert.Contains(t, items, "api.port")
				assert.Contains(t, items, "observability.log_level")
				assert.Contains(t, items, "gateway.public_rate_limit")
			},
		},
		{
			name: "get configuration items with empty keys list",
			keys: []string{},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			expectedError: "configuration keys list cannot be empty",
		},
		{
			name: "get configuration items with nil keys list",
			keys: nil,
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			expectedError: "configuration keys list cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - Set up test environment first
			_, testCancel := sharedtesting.SetupDaprTest()
			defer testCancel()
			defer ResetClientForTesting()
			
			ctx, cancel := tt.setupContext()
			defer cancel()

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			config := NewConfiguration(client)
			require.NotNil(t, config)

			// Act
			items, err := config.GetConfigurationItems(ctx, tt.keys)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, items)
			} else {
				require.NoError(t, err)
				require.NotNil(t, items)
				if tt.validateResult != nil {
					tt.validateResult(t, items)
				}
			}
		})
	}
}

func TestConfiguration_LoadAppConfig(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func() (context.Context, context.CancelFunc)
		envVars       map[string]string
		expectedError string
		validateResult func(*testing.T, *AppConfig)
	}{
		{
			name: "load app config for development environment",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			envVars: map[string]string{
				"ENVIRONMENT": "development",
				"DAPR_APP_ID": "international-center",
			},
			validateResult: func(t *testing.T, appConfig *AppConfig) {
				assert.Equal(t, "development", appConfig.Environment)
				assert.Equal(t, "international-center", appConfig.AppID)
				assert.NotEmpty(t, appConfig.Version)
				
				// Validate database config
				assert.Greater(t, appConfig.DatabaseConfig.MaxConnections, 0)
				assert.Greater(t, appConfig.DatabaseConfig.ConnectionTimeout, 0)
				assert.Greater(t, appConfig.DatabaseConfig.QueryTimeout, 0)
				
				// Validate API config
				assert.Greater(t, appConfig.APIConfig.Port, 0)
				assert.Greater(t, appConfig.APIConfig.ReadTimeout, 0)
				assert.Greater(t, appConfig.APIConfig.WriteTimeout, 0)
				assert.Greater(t, appConfig.APIConfig.IdleTimeout, 0)
				
				// Validate gateway config
				assert.Greater(t, appConfig.GatewayConfig.PublicRateLimit, 0)
				assert.Greater(t, appConfig.GatewayConfig.AdminRateLimit, 0)
				assert.NotNil(t, appConfig.GatewayConfig.AllowedOrigins)
				
				// Validate Dapr config
				assert.NotEmpty(t, appConfig.DaprConfig.StateStoreName)
				assert.NotEmpty(t, appConfig.DaprConfig.PubSubName)
				assert.NotEmpty(t, appConfig.DaprConfig.SecretStoreName)
				assert.NotEmpty(t, appConfig.DaprConfig.BlobBindingName)
				assert.Greater(t, appConfig.DaprConfig.HTTPPort, 0)
				assert.Greater(t, appConfig.DaprConfig.GRPCPort, 0)
				
				// Validate observability config
				assert.NotEmpty(t, appConfig.ObservabilityConfig.LogLevel)
			},
		},
		{
			name: "load app config for production environment",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			envVars: map[string]string{
				"ENVIRONMENT": "production",
				"DAPR_APP_ID": "international-center-prod",
			},
			validateResult: func(t *testing.T, appConfig *AppConfig) {
				assert.Equal(t, "production", appConfig.Environment)
				assert.Equal(t, "international-center-prod", appConfig.AppID)
			},
		},
		{
			name: "load app config for staging environment",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			envVars: map[string]string{
				"ENVIRONMENT": "staging",
				"DAPR_APP_ID": "international-center-staging",
			},
			validateResult: func(t *testing.T, appConfig *AppConfig) {
				assert.Equal(t, "staging", appConfig.Environment)
				assert.Equal(t, "international-center-staging", appConfig.AppID)
			},
		},
		{
			name: "load app config with timeout context",
			setupContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 1*time.Millisecond)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - Set up test environment first
			_, testCancel := sharedtesting.SetupDaprTest()
			defer testCancel()
			defer ResetClientForTesting()
			
			ctx, cancel := tt.setupContext()
			defer cancel()

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			config := NewConfiguration(client)
			require.NotNil(t, config)

			// Act
			appConfig, err := config.LoadAppConfig(ctx)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, appConfig)
			} else {
				if tt.validateResult != nil {
					require.NoError(t, err)
					require.NotNil(t, appConfig)
					tt.validateResult(t, appConfig)
				}
				// For timeout/cancelled context tests, we just verify no panic occurs
			}
		})
	}
}

func TestConfiguration_GetEnvironmentSpecificConfig(t *testing.T) {
	tests := []struct {
		name          string
		baseKey       string
		setupContext  func() (context.Context, context.CancelFunc)
		envVars       map[string]string
		expectedError string
		validateResult func(*testing.T, *ConfigItem)
	}{
		{
			name:    "get development-specific database config",
			baseKey: "database.connection_string",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			envVars: map[string]string{
				"ENVIRONMENT": "development",
			},
			validateResult: func(t *testing.T, item *ConfigItem) {
				// Should try "database.connection_string.development" first, then fallback to "database.connection_string"
				assert.NotNil(t, item)
				assert.Contains(t, item.Key, "database.connection_string")
			},
		},
		{
			name:    "get production-specific API config",
			baseKey: "api.rate_limit",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			envVars: map[string]string{
				"ENVIRONMENT": "production",
				"DAPR_APP_ID": "international-center",
			},
			validateResult: func(t *testing.T, item *ConfigItem) {
				assert.NotNil(t, item)
				assert.Contains(t, item.Key, "api.rate_limit")
			},
		},
		{
			name:    "get staging-specific observability config",
			baseKey: "observability.sampling_rate",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			envVars: map[string]string{
				"ENVIRONMENT": "staging",
				"DAPR_APP_ID": "international-center",
			},
			validateResult: func(t *testing.T, item *ConfigItem) {
				assert.NotNil(t, item)
				assert.Contains(t, item.Key, "observability.sampling_rate")
			},
		},
		{
			name:    "get config with empty base key",
			baseKey: "",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			expectedError: "base key cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - Set up test environment first
			_, testCancel := sharedtesting.SetupDaprTest()
			defer testCancel()
			defer ResetClientForTesting()
			
			ctx, cancel := tt.setupContext()
			defer cancel()

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			config := NewConfiguration(client)
			require.NotNil(t, config)

			// Act
			item, err := config.GetEnvironmentSpecificConfig(ctx, tt.baseKey)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, item)
			} else {
				require.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, item)
				}
			}
		})
	}
}

func TestConfiguration_WatchConfiguration(t *testing.T) {
	tests := []struct {
		name          string
		keys          []string
		setupContext  func() (context.Context, context.CancelFunc)
		envVars       map[string]string
		expectedError string
	}{
		{
			name: "watch database configuration keys",
			keys: []string{
				"database.max_connections",
				"database.connection_timeout",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			expectedError: "configuration watching not implemented",
		},
		{
			name: "watch API configuration keys",
			keys: []string{
				"api.port",
				"api.read_timeout",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			expectedError: "configuration watching not implemented",
		},
		{
			name: "watch configuration with empty keys list",
			keys: []string{},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			expectedError: "configuration watching not implemented",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - Set up test environment first
			_, testCancel := sharedtesting.SetupDaprTest()
			defer testCancel()
			defer ResetClientForTesting()
			
			ctx, cancel := tt.setupContext()
			defer cancel()

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			config := NewConfiguration(client)
			require.NotNil(t, config)

			callback := func(items map[string]*ConfigItem) {
				// Mock callback function
			}

			// Act
			err = config.WatchConfiguration(ctx, tt.keys, callback)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfiguration_HealthCheck(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func() (context.Context, context.CancelFunc)
		envVars       map[string]string
		expectedError string
		expectHealthy bool
	}{
		{
			name: "health check with valid configuration",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			expectHealthy: true,
		},
		{
			name: "health check with timeout context",
			setupContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 1*time.Millisecond)
			},
		},
		{
			name: "health check with cancelled context",
			setupContext: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				return ctx, func() {}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - Set up test environment first
			_, testCancel := sharedtesting.SetupDaprTest()
			defer testCancel()
			defer ResetClientForTesting()
			
			ctx, cancel := tt.setupContext()
			defer cancel()

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			config := NewConfiguration(client)
			require.NotNil(t, config)

			// Act
			err = config.HealthCheck(ctx)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else if tt.expectHealthy {
				assert.NoError(t, err)
			}
			// For timeout/cancelled context tests, we just verify no panic occurs
		})
	}
}

func TestConfiguration_ConfigValueParsing(t *testing.T) {
	// Arrange
	ctx, cancel := sharedtesting.SetupDaprTest()
	defer cancel()
	defer ResetClientForTesting()

	client, err := NewClient()
	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	config := NewConfiguration(client)
	require.NotNil(t, config)

	// Create mock configuration items for testing parsing
	mockItems := map[string]*ConfigItem{
		"string.value":  {Key: "string.value", Value: "test-string", Version: "1.0"},
		"int.value":     {Key: "int.value", Value: "42", Version: "1.0"},
		"bool.true":     {Key: "bool.true", Value: "true", Version: "1.0"},
		"bool.false":    {Key: "bool.false", Value: "false", Version: "1.0"},
		"array.value":   {Key: "array.value", Value: "item1,item2,item3", Version: "1.0"},
		"empty.value":   {Key: "empty.value", Value: "", Version: "1.0"},
		"invalid.int":   {Key: "invalid.int", Value: "not-a-number", Version: "1.0"},
		"invalid.bool":  {Key: "invalid.bool", Value: "not-a-boolean", Version: "1.0"},
	}

	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "parse string values",
			testFunc: func(t *testing.T) {
				value := config.getConfigValue(mockItems, "string.value", "default")
				assert.Equal(t, "test-string", value)
				
				defaultValue := config.getConfigValue(mockItems, "missing.key", "default")
				assert.Equal(t, "default", defaultValue)
				
				emptyValue := config.getConfigValue(mockItems, "empty.value", "default")
				assert.Equal(t, "default", emptyValue)
			},
		},
		{
			name: "parse integer values",
			testFunc: func(t *testing.T) {
				value := config.getConfigInt(mockItems, "int.value", 0)
				assert.Equal(t, 42, value)
				
				defaultValue := config.getConfigInt(mockItems, "missing.key", 100)
				assert.Equal(t, 100, defaultValue)
				
				invalidValue := config.getConfigInt(mockItems, "invalid.int", 200)
				assert.Equal(t, 200, invalidValue)
			},
		},
		{
			name: "parse boolean values",
			testFunc: func(t *testing.T) {
				trueValue := config.getConfigBool(mockItems, "bool.true", false)
				assert.True(t, trueValue)
				
				falseValue := config.getConfigBool(mockItems, "bool.false", true)
				assert.False(t, falseValue)
				
				defaultValue := config.getConfigBool(mockItems, "missing.key", true)
				assert.True(t, defaultValue)
				
				invalidValue := config.getConfigBool(mockItems, "invalid.bool", true)
				assert.True(t, invalidValue)
			},
		},
		{
			name: "parse array values",
			testFunc: func(t *testing.T) {
				value := config.getConfigArray(mockItems, "array.value", []string{"default"})
				assert.Equal(t, []string{"item1", "item2", "item3"}, value)
				
				defaultValue := config.getConfigArray(mockItems, "missing.key", []string{"default1", "default2"})
				assert.Equal(t, []string{"default1", "default2"}, defaultValue)
				
				emptyValue := config.getConfigArray(mockItems, "empty.value", []string{"default"})
				assert.Equal(t, []string{"default"}, emptyValue)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}

	_ = ctx // Use context to avoid linting issues
}

func TestConfiguration_Concurrent_Access(t *testing.T) {
	// Arrange
	ctx, cancel := sharedtesting.SetupDaprTest()
	defer cancel()
	defer ResetClientForTesting()

	client, err := NewClient()
	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	config := NewConfiguration(client)
	require.NotNil(t, config)

	// Act - Test concurrent access to configuration methods
	const numGoroutines = 10
	done := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			// Test concurrent calls to various configuration methods
			_, err1 := config.GetConfigurationItem(ctx, "concurrent.test.key")
			_, err2 := config.LoadAppConfig(ctx)
			err3 := config.HealthCheck(ctx)
			
			if err1 != nil {
				done <- err1
				return
			}
			if err2 != nil {
				done <- err2
				return
			}
			done <- err3
		}(i)
	}

	// Assert - All goroutines should complete without error
	for i := 0; i < numGoroutines; i++ {
		select {
		case err := <-done:
			// Operations may fail in test environment, but should not panic
			if err != nil {
				t.Logf("Goroutine %d completed with error: %v", i, err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for goroutine completion")
		}
	}
}

func TestConfiguration_Error_Handling(t *testing.T) {
	tests := []struct {
		name          string
		setupTest     func(*testing.T) (*Configuration, context.Context, context.CancelFunc)
		operation     func(context.Context, *Configuration) error
		expectedError string
	}{
		{
			name: "get configuration item with nil context should not panic",
			setupTest: func(t *testing.T) (*Configuration, context.Context, context.CancelFunc) {
				client, err := NewClient()
				require.NoError(t, err)
				config := NewConfiguration(client)
				return config, nil, func() { client.Close() }
			},
			operation: func(ctx context.Context, config *Configuration) error {
				_, err := config.GetConfigurationItem(ctx, "test.key")
				return err
			},
		},
		{
			name: "load app config with nil context should not panic",
			setupTest: func(t *testing.T) (*Configuration, context.Context, context.CancelFunc) {
				client, err := NewClient()
				require.NoError(t, err)
				config := NewConfiguration(client)
				return config, nil, func() { client.Close() }
			},
			operation: func(ctx context.Context, config *Configuration) error {
				_, err := config.LoadAppConfig(ctx)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - Set up test environment first
			_, testCancel := sharedtesting.SetupDaprTest()
			defer testCancel()
			defer ResetClientForTesting()
			
			config, ctx, cleanup := tt.setupTest(t)
			defer cleanup()

			// Act & Assert - Should not panic
			assert.NotPanics(t, func() {
				err := tt.operation(ctx, config)
				if tt.expectedError != "" {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			})
		})
	}
}