package dapr

import (
	"context"
	"testing"
	"time"

	sharedtesting "github.com/axiom-software-co/international-center/src/backend/internal/shared/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED PHASE - Dapr Secrets Management Tests (40+ test cases)

func TestNewSecrets(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		expectedError  string
		validateResult func(*testing.T, *Secrets)
	}{
		{
			name: "create secrets manager with default environment",
			envVars: map[string]string{},
			validateResult: func(t *testing.T, secrets *Secrets) {
				assert.NotNil(t, secrets)
				assert.NotNil(t, secrets.client)
			},
		},
		{
			name: "create secrets manager with production environment",
			envVars: map[string]string{
				"ENVIRONMENT": "production",
				"DAPR_SECRET_STORE_NAME": "azure-keyvault",
			},
			validateResult: func(t *testing.T, secrets *Secrets) {
				assert.NotNil(t, secrets)
				assert.NotNil(t, secrets.client)
			},
		},
		{
			name: "create secrets manager with staging environment",
			envVars: map[string]string{
				"ENVIRONMENT": "staging",
				"DAPR_SECRET_STORE_NAME": "staging-secrets",
			},
			validateResult: func(t *testing.T, secrets *Secrets) {
				assert.NotNil(t, secrets)
				assert.NotNil(t, secrets.client)
			},
		},
		{
			name: "create secrets manager with custom secret store",
			envVars: map[string]string{
				"DAPR_SECRET_STORE_NAME": "custom-secret-store",
			},
			validateResult: func(t *testing.T, secrets *Secrets) {
				assert.NotNil(t, secrets)
				assert.NotNil(t, secrets.client)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			// Reset client for testing and enable test mode
			ResetClientForTesting()
			t.Setenv("DAPR_TEST_MODE", "true")
			
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			// Act
			secrets := NewSecrets(client)

			// Assert
			if tt.expectedError != "" {
				assert.Nil(t, secrets)
			} else {
				require.NotNil(t, secrets)
				if tt.validateResult != nil {
					tt.validateResult(t, secrets)
				}
			}

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestSecrets_GetSecret(t *testing.T) {
	tests := []struct {
		name          string
		secretName    string
		setupContext  func() (context.Context, context.CancelFunc)
		envVars       map[string]string
		expectedError string
		validateResult func(*testing.T, string)
	}{
		{
			name:       "get database connection string",
			secretName: "database-connection-string",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, secret string) {
				assert.NotEmpty(t, secret)
			},
		},
		{
			name:       "get Redis connection string",
			secretName: "redis-connection-string",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, secret string) {
				assert.NotEmpty(t, secret)
			},
		},
		{
			name:       "get OAuth2 client secret",
			secretName: "oauth2-client-secret",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, secret string) {
				assert.NotEmpty(t, secret)
				assert.True(t, len(secret) >= 32) // OAuth2 secrets should be substantial
			},
		},
		{
			name:       "get API key secret",
			secretName: "external-api-key",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, secret string) {
				assert.NotEmpty(t, secret)
			},
		},
		{
			name:       "get non-existent secret",
			secretName: "non-existent-secret",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			expectedError: "secret not found",
		},
		{
			name:       "get secret with empty name",
			secretName: "",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			expectedError: "secret name cannot be empty",
		},
		{
			name:       "get secret with timeout context",
			secretName: "database-connection-string",
			setupContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 1*time.Millisecond)
			},
		},
		{
			name:       "get secret with cancelled context",
			secretName: "database-connection-string",
			setupContext: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				return ctx, func() {}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := tt.setupContext()
			defer cancel()

			// Reset client for testing and enable test mode
			ResetClientForTesting()
			t.Setenv("DAPR_TEST_MODE", "true")
			
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			secrets := NewSecrets(client)
			require.NotNil(t, secrets)

			// Act
			secret, err := secrets.GetSecret(ctx, tt.secretName)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Empty(t, secret)
			} else {
				if tt.validateResult != nil {
					require.NoError(t, err)
					tt.validateResult(t, secret)
				}
				// For timeout/cancelled context tests, we just verify no panic occurs
			}
		})
	}
}

func TestSecrets_GetSecrets(t *testing.T) {
	tests := []struct {
		name          string
		secretNames   []string
		setupContext  func() (context.Context, context.CancelFunc)
		envVars       map[string]string
		expectedError string
		validateResult func(*testing.T, map[string]string)
	}{
		{
			name: "get multiple database secrets",
			secretNames: []string{
				"database-connection-string",
				"database-password",
				"database-username",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, secrets map[string]string) {
				assert.NotNil(t, secrets)
				// Should return empty map if secrets don't exist
			},
		},
		{
			name: "get multiple OAuth2 secrets",
			secretNames: []string{
				"oauth2-client-id",
				"oauth2-client-secret",
				"oauth2-tenant-id",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, secrets map[string]string) {
				assert.NotNil(t, secrets)
			},
		},
		{
			name: "get multiple secrets with some missing",
			secretNames: []string{
				"database-connection-string",
				"non-existent-secret",
				"redis-connection-string",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, secrets map[string]string) {
				assert.NotNil(t, secrets)
				// Should return only existing secrets
			},
		},
		{
			name:        "get secrets with empty names list",
			secretNames: []string{},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, secrets map[string]string) {
				assert.NotNil(t, secrets)
				assert.Len(t, secrets, 0)
			},
		},
		{
			name:        "get secrets with nil names list",
			secretNames: nil,
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, secrets map[string]string) {
				assert.NotNil(t, secrets)
				assert.Len(t, secrets, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := tt.setupContext()
			defer cancel()

			// Reset client for testing and enable test mode
			ResetClientForTesting()
			t.Setenv("DAPR_TEST_MODE", "true")
			
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			secrets := NewSecrets(client)
			require.NotNil(t, secrets)

			// Act
			secretsMap, err := secrets.GetSecrets(ctx, tt.secretNames)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, secretsMap)
			} else {
				require.NoError(t, err)
				require.NotNil(t, secretsMap)
				if tt.validateResult != nil {
					tt.validateResult(t, secretsMap)
				}
			}
		})
	}
}

func TestSecrets_GetDatabaseConnectionString(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func() (context.Context, context.CancelFunc)
		envVars       map[string]string
		expectedError string
		validateResult func(*testing.T, string)
	}{
		{
			name: "get database connection string with default key",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, connectionString string) {
				assert.NotEmpty(t, connectionString)
			},
		},
		{
			name: "get database connection string with custom key",
			envVars: map[string]string{
				"DATABASE_CONNECTION_SECRET_KEY": "custom-db-connection",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, connectionString string) {
				assert.NotEmpty(t, connectionString)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := tt.setupContext()
			defer cancel()

			// Reset client for testing and enable test mode
			ResetClientForTesting()
			t.Setenv("DAPR_TEST_MODE", "true")
			
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			secrets := NewSecrets(client)
			require.NotNil(t, secrets)

			// Act
			connectionString, err := secrets.GetDatabaseConnectionString(ctx)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Empty(t, connectionString)
			} else {
				if tt.validateResult != nil {
					require.NoError(t, err)
					tt.validateResult(t, connectionString)
				}
			}
		})
	}
}

func TestSecrets_GetRedisConnectionString(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func() (context.Context, context.CancelFunc)
		envVars       map[string]string
		expectedError string
		validateResult func(*testing.T, string)
	}{
		{
			name: "get Redis connection string with default key",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, connectionString string) {
				assert.NotEmpty(t, connectionString)
			},
		},
		{
			name: "get Redis connection string with custom key",
			envVars: map[string]string{
				"REDIS_CONNECTION_SECRET_KEY": "custom-redis-connection",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, connectionString string) {
				assert.NotEmpty(t, connectionString)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := tt.setupContext()
			defer cancel()

			// Reset client for testing and enable test mode
			ResetClientForTesting()
			t.Setenv("DAPR_TEST_MODE", "true")
			
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			secrets := NewSecrets(client)
			require.NotNil(t, secrets)

			// Act
			connectionString, err := secrets.GetRedisConnectionString(ctx)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Empty(t, connectionString)
			} else {
				if tt.validateResult != nil {
					require.NoError(t, err)
					tt.validateResult(t, connectionString)
				}
			}
		})
	}
}

func TestSecrets_GetOAuth2ClientSecret(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func() (context.Context, context.CancelFunc)
		envVars       map[string]string
		expectedError string
		validateResult func(*testing.T, string)
	}{
		{
			name: "get OAuth2 client secret with default key",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, clientSecret string) {
				assert.NotEmpty(t, clientSecret)
			},
		},
		{
			name: "get OAuth2 client secret with custom key",
			envVars: map[string]string{
				"OAUTH2_CLIENT_SECRET_KEY": "custom-oauth2-secret",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, clientSecret string) {
				assert.NotEmpty(t, clientSecret)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := tt.setupContext()
			defer cancel()

			// Reset client for testing and enable test mode
			ResetClientForTesting()
			t.Setenv("DAPR_TEST_MODE", "true")
			
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			secrets := NewSecrets(client)
			require.NotNil(t, secrets)

			// Act
			clientSecret, err := secrets.GetOAuth2ClientSecret(ctx)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Empty(t, clientSecret)
			} else {
				if tt.validateResult != nil {
					require.NoError(t, err)
					tt.validateResult(t, clientSecret)
				}
			}
		})
	}
}

func TestSecrets_ClearCache(t *testing.T) {
	// Arrange
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	// Reset client for testing and enable test mode
	ResetClientForTesting()
	t.Setenv("DAPR_TEST_MODE", "true")
	
	client, err := NewClient()
	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	secrets := NewSecrets(client)
	require.NotNil(t, secrets)

	// Act
	secrets.ClearCache()

	// Assert - Should not panic
	assert.NotNil(t, secrets)

	_ = ctx // Use context to avoid linting issues
}

func TestSecrets_ClearSecretFromCache(t *testing.T) {
	tests := []struct {
		name       string
		secretKey  string
	}{
		{
			name:      "clear existing secret from cache",
			secretKey: "test-secret-key",
		},
		{
			name:      "clear non-existent secret from cache",
			secretKey: "non-existent-key",
		},
		{
			name:      "clear secret with empty key",
			secretKey: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			// Enable test mode to use mock Dapr client
			t.Setenv("DAPR_TEST_MODE", "true")
			
			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			secrets := NewSecrets(client)
			require.NotNil(t, secrets)

			// Act
			secrets.ClearSecretFromCache(tt.secretKey)

			// Assert - Should not panic
			assert.NotNil(t, secrets)

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestSecrets_RefreshSecret(t *testing.T) {
	tests := []struct {
		name          string
		secretKey     string
		setupContext  func() (context.Context, context.CancelFunc)
		expectedError string
		validateResult func(*testing.T, string)
	}{
		{
			name:      "refresh existing secret",
			secretKey: "database-connection-string",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, secret string) {
				assert.NotEmpty(t, secret)
			},
		},
		{
			name:      "refresh non-existent secret",
			secretKey: "non-existent-secret",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			expectedError: "secret not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := tt.setupContext()
			defer cancel()

			// Enable test mode to use mock Dapr client
			t.Setenv("DAPR_TEST_MODE", "true")
			
			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			secrets := NewSecrets(client)
			require.NotNil(t, secrets)

			// Act
			secret, err := secrets.RefreshSecret(ctx, tt.secretKey)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Empty(t, secret)
			} else {
				if tt.validateResult != nil {
					require.NoError(t, err)
					tt.validateResult(t, secret)
				}
			}
		})
	}
}

func TestSecrets_HealthCheck(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func() (context.Context, context.CancelFunc)
		envVars       map[string]string
		expectedError string
		expectHealthy bool
	}{
		{
			name: "health check with valid secrets manager",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
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
			// Arrange
			ctx, cancel := tt.setupContext()
			defer cancel()

			// Reset client for testing and enable test mode
			ResetClientForTesting()
			t.Setenv("DAPR_TEST_MODE", "true")
			
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			secrets := NewSecrets(client)
			require.NotNil(t, secrets)

			// Act
			err = secrets.HealthCheck(ctx)

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

func TestSecrets_Concurrent_Access(t *testing.T) {
	// Arrange
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	// Reset client for testing and enable test mode
	ResetClientForTesting()
	t.Setenv("DAPR_TEST_MODE", "true")
	
	client, err := NewClient()
	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	secrets := NewSecrets(client)
	require.NotNil(t, secrets)

	// Act - Test concurrent access to secrets methods
	const numGoroutines = 10
	done := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			// Test concurrent calls to various secrets methods
			secretName := "concurrent-test-secret"
			_, err1 := secrets.GetSecret(ctx, secretName)
			
			secrets.ClearSecretFromCache(secretName)
			
			if err1 != nil {
				done <- err1
				return
			}
			done <- nil
		}(i)
	}

	// Assert - All goroutines should complete without error
	for i := 0; i < numGoroutines; i++ {
		select {
		case err := <-done:
			// Some operations may fail in test environment, but should not panic
			if err != nil {
				t.Logf("Goroutine %d completed with error: %v", i, err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for goroutine completion")
		}
	}
}

func TestSecrets_Error_Handling(t *testing.T) {
	tests := []struct {
		name          string
		setupTest     func(*testing.T) (*Secrets, context.Context, context.CancelFunc)
		operation     func(context.Context, *Secrets) error
		expectedError string
	}{
		{
			name: "get secret with nil context should not panic",
			setupTest: func(t *testing.T) (*Secrets, context.Context, context.CancelFunc) {
				// Reset client for testing and enable test mode
				ResetClientForTesting()
				t.Setenv("DAPR_TEST_MODE", "true")
				
				client, err := NewClient()
				require.NoError(t, err)
				secrets := NewSecrets(client)
				return secrets, nil, func() { client.Close() }
			},
			operation: func(ctx context.Context, secrets *Secrets) error {
				_, err := secrets.GetSecret(ctx, "test-secret")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			secrets, ctx, cleanup := tt.setupTest(t)
			defer cleanup()

			// Act & Assert - Should not panic
			assert.NotPanics(t, func() {
				err := tt.operation(ctx, secrets)
				if tt.expectedError != "" {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			})
		})
	}
}