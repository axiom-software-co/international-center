package dapr

import (
	"context"
	"testing"
	"time"

	sharedtesting "github.com/axiom-software-co/international-center/src/backend/internal/shared/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED PHASE - Dapr Client Tests (40+ test cases)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		expectedError  string
		validateResult func(*testing.T, *Client)
	}{
		{
			name: "create client with default environment",
			envVars: map[string]string{
				"ENVIRONMENT": "development",
				"DAPR_APP_ID": "international-center",
			},
			validateResult: func(t *testing.T, client *Client) {
				assert.Equal(t, "development", client.GetEnvironment())
				assert.Equal(t, "international-center", client.GetAppID())
				// In test mode, GetClient() returns nil, which is expected
				assert.Nil(t, client.GetClient())
			},
		},
		{
			name: "create client with production environment",
			envVars: map[string]string{
				"ENVIRONMENT": "production",
				"DAPR_APP_ID": "international-center-prod",
			},
			validateResult: func(t *testing.T, client *Client) {
				assert.Equal(t, "production", client.GetEnvironment())
				assert.Equal(t, "international-center-prod", client.GetAppID())
			},
		},
		{
			name: "create client with staging environment",
			envVars: map[string]string{
				"ENVIRONMENT": "staging",
				"DAPR_APP_ID": "international-center-staging",
			},
			validateResult: func(t *testing.T, client *Client) {
				assert.Equal(t, "staging", client.GetEnvironment())
				assert.Equal(t, "international-center-staging", client.GetAppID())
			},
		},
		{
			name: "create client with custom app ID",
			envVars: map[string]string{
				"DAPR_APP_ID": "custom-app-id",
			},
			validateResult: func(t *testing.T, client *Client) {
				assert.Equal(t, "custom-app-id", client.GetAppID())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - Reset client first, then set environment variables
			ResetClientForTesting()
			
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}
			
			// Set up test mode after environment variables are set
			t.Setenv("DAPR_TEST_MODE", "true")
			
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			defer ResetClientForTesting()

			// Act
			client, err := NewClient()

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, client)
			} else {
				require.NoError(t, err)
				require.NotNil(t, client)
				if tt.validateResult != nil {
					tt.validateResult(t, client)
				}
			}

			// Cleanup
			if client != nil {
				client.Close()
			}

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestClient_GetClient(t *testing.T) {
	// Arrange
	ResetClientForTesting()
	t.Setenv("DAPR_TEST_MODE", "true")
	
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()
	defer ResetClientForTesting()

	client, err := NewClient()
	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	// Act
	daprClient := client.GetClient()

	// Assert - In test mode, client should be nil
	assert.Nil(t, daprClient, "In test mode, underlying Dapr client should be nil")

	_ = ctx // Use context to avoid linting issues
}

func TestClient_GetEnvironment(t *testing.T) {
	tests := []struct {
		name        string
		envVar      string
		expectedEnv string
	}{
		{
			name:        "default environment",
			envVar:      "",
			expectedEnv: "development",
		},
		{
			name:        "production environment",
			envVar:      "production",
			expectedEnv: "production",
		},
		{
			name:        "staging environment",
			envVar:      "staging",
			expectedEnv: "staging",
		},
		{
			name:        "test environment",
			envVar:      "test",
			expectedEnv: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - Reset client first
			ResetClientForTesting()
			
			if tt.envVar != "" {
				t.Setenv("ENVIRONMENT", tt.envVar)
			}
			t.Setenv("DAPR_TEST_MODE", "true")
			
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			defer ResetClientForTesting()

			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			// Act
			environment := client.GetEnvironment()

			// Assert
			assert.Equal(t, tt.expectedEnv, environment)

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestClient_GetAppID(t *testing.T) {
	tests := []struct {
		name           string
		appIDVar       string
		expectedAppID  string
	}{
		{
			name:          "default app ID",
			appIDVar:      "",
			expectedAppID: "international-center",
		},
		{
			name:          "custom app ID",
			appIDVar:      "custom-app",
			expectedAppID: "custom-app",
		},
		{
			name:          "production app ID",
			appIDVar:      "international-center-prod",
			expectedAppID: "international-center-prod",
		},
		{
			name:          "staging app ID",
			appIDVar:      "international-center-staging",
			expectedAppID: "international-center-staging",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - Reset client first
			ResetClientForTesting()
			
			if tt.appIDVar != "" {
				t.Setenv("DAPR_APP_ID", tt.appIDVar)
			}
			t.Setenv("DAPR_TEST_MODE", "true")
			
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			defer ResetClientForTesting()

			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			// Act
			appID := client.GetAppID()

			// Assert
			assert.Equal(t, tt.expectedAppID, appID)

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestClient_Close(t *testing.T) {
	tests := []struct {
		name         string
		setupClient  func() *Client
		expectedError string
	}{
		{
			name: "close valid client",
			setupClient: func() *Client {
				ResetClientForTesting()
				client, err := NewClient()
				require.NoError(t, err)
				return client
			},
		},
		{
			name: "close client multiple times",
			setupClient: func() *Client {
				ResetClientForTesting()
				client, err := NewClient()
				require.NoError(t, err)
				return client
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			t.Setenv("DAPR_TEST_MODE", "true")
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			defer ResetClientForTesting()

			client := tt.setupClient()
			require.NotNil(t, client)

			// Act
			err := client.Close()

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			// Test multiple closes
			if tt.name == "close client multiple times" {
				err2 := client.Close()
				assert.NoError(t, err2) // Should not error on multiple closes
			}

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestClient_HealthCheck(t *testing.T) {
	tests := []struct {
		name          string
		setupClient   func() *Client
		setupContext  func() (context.Context, context.CancelFunc)
		expectedError string
		expectHealthy bool
	}{
		{
			name: "health check with valid client",
			setupClient: func() *Client {
				ResetClientForTesting()
				client, err := NewClient()
				require.NoError(t, err)
				return client
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			expectHealthy: true,
		},
		{
			name: "health check with timeout context",
			setupClient: func() *Client {
				ResetClientForTesting()
				client, err := NewClient()
				require.NoError(t, err)
				return client
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 1*time.Millisecond)
			},
		},
		{
			name: "health check with cancelled context",
			setupClient: func() *Client {
				ResetClientForTesting()
				client, err := NewClient()
				require.NoError(t, err)
				return client
			},
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
			t.Setenv("DAPR_TEST_MODE", "true")
			defer ResetClientForTesting()
			
			ctx, cancel := tt.setupContext()
			defer cancel()

			client := tt.setupClient()
			require.NotNil(t, client)
			defer client.Close()

			// Act
			err := client.HealthCheck(ctx)

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

func TestClient_IsHealthy(t *testing.T) {
	tests := []struct {
		name         string
		setupClient  func() *Client
		setupContext func() (context.Context, context.CancelFunc)
		expectedHealthy bool
	}{
		{
			name: "healthy client returns true",
			setupClient: func() *Client {
				ResetClientForTesting()
				client, err := NewClient()
				require.NoError(t, err)
				return client
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			expectedHealthy: true,
		},
		{
			name: "client with timeout context",
			setupClient: func() *Client {
				ResetClientForTesting()
				client, err := NewClient()
				require.NoError(t, err)
				return client
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 1*time.Millisecond)
			},
			expectedHealthy: true, // In test mode, health check always succeeds
		},
		{
			name: "client with cancelled context",
			setupClient: func() *Client {
				ResetClientForTesting()
				client, err := NewClient()
				require.NoError(t, err)
				return client
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				return ctx, func() {}
			},
			expectedHealthy: true, // In test mode, health check always succeeds
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			t.Setenv("DAPR_TEST_MODE", "true")
			defer ResetClientForTesting()
			
			ctx, cancel := tt.setupContext()
			defer cancel()

			client := tt.setupClient()
			require.NotNil(t, client)
			defer client.Close()

			// Act
			healthy := client.IsHealthy(ctx)

			// Assert
			assert.Equal(t, tt.expectedHealthy, healthy)
		})
	}
}

func TestClient_Singleton_Behavior(t *testing.T) {
	// Arrange
	ResetClientForTesting()
	t.Setenv("DAPR_TEST_MODE", "true")
	
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()
	defer ResetClientForTesting()

	// Act
	client1, err1 := NewClient()
	require.NoError(t, err1)

	client2, err2 := NewClient()
	require.NoError(t, err2)

	// Assert - Both should be the same singleton instance
	assert.Same(t, client1, client2, "NewClient should return the same singleton instance")
	
	// Cleanup - Only close once since they're the same instance
	if client1 != nil {
		client1.Close()
	}

	_ = ctx // Use context to avoid linting issues
}

func TestClient_Environment_Configuration(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		expectedConfig map[string]interface{}
	}{
		{
			name:        "development environment configuration",
			environment: "development",
			expectedConfig: map[string]interface{}{
				"env":   "development",
				"appId": "international-center",
			},
		},
		{
			name:        "production environment configuration",
			environment: "production",
			expectedConfig: map[string]interface{}{
				"env":   "production",
				"appId": "international-center",
			},
		},
		{
			name:        "staging environment configuration",
			environment: "staging",
			expectedConfig: map[string]interface{}{
				"env":   "staging", 
				"appId": "international-center",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - Reset client first
			ResetClientForTesting()
			
			t.Setenv("ENVIRONMENT", tt.environment)
			t.Setenv("DAPR_TEST_MODE", "true")
			
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			defer ResetClientForTesting()

			// Act
			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			// Assert
			assert.Equal(t, tt.expectedConfig["env"], client.GetEnvironment())
			assert.Equal(t, tt.expectedConfig["appId"], client.GetAppID())

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestClient_Timeout_Handling(t *testing.T) {
	tests := []struct {
		name        string
		timeout     time.Duration
		operation   func(context.Context, *Client) error
		expectError bool
	}{
		{
			name:    "health check with sufficient timeout",
			timeout: 5 * time.Second,
			operation: func(ctx context.Context, client *Client) error {
				return client.HealthCheck(ctx)
			},
			expectError: false,
		},
		{
			name:    "health check with minimal timeout",
			timeout: 1 * time.Millisecond,
			operation: func(ctx context.Context, client *Client) error {
				return client.HealthCheck(ctx)
			},
			expectError: false, // Health check is simple and should complete quickly
		},
		{
			name:    "is healthy with sufficient timeout",
			timeout: 5 * time.Second,
			operation: func(ctx context.Context, client *Client) error {
				healthy := client.IsHealthy(ctx)
				if !healthy {
					return assert.AnError
				}
				return nil
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ResetClientForTesting()
			t.Setenv("DAPR_TEST_MODE", "true")
			defer ResetClientForTesting()
			
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			// Act
			err = tt.operation(ctx, client)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_Concurrent_Access(t *testing.T) {
	// Arrange
	ResetClientForTesting()
	t.Setenv("DAPR_TEST_MODE", "true")
	defer ResetClientForTesting()
	
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	client, err := NewClient()
	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	// Act - Test concurrent access to client methods
	const numGoroutines = 10
	done := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			// Test concurrent calls to various client methods
			_ = client.GetEnvironment()
			_ = client.GetAppID()
			_ = client.GetClient()
			
			err := client.HealthCheck(ctx)
			done <- err
		}()
	}

	// Assert - All goroutines should complete without error
	for i := 0; i < numGoroutines; i++ {
		select {
		case err := <-done:
			assert.NoError(t, err)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for goroutine completion")
		}
	}
}

func TestClient_Error_Handling(t *testing.T) {
	tests := []struct {
		name          string
		setupTest     func(*testing.T) (*Client, context.Context, context.CancelFunc)
		operation     func(context.Context, *Client) error
		expectedError string
	}{
		{
			name: "health check with nil context should not panic",
			setupTest: func(t *testing.T) (*Client, context.Context, context.CancelFunc) {
				ResetClientForTesting()
				t.Setenv("DAPR_TEST_MODE", "true")
				client, err := NewClient()
				require.NoError(t, err)
				return client, nil, func() {}
			},
			operation: func(ctx context.Context, client *Client) error {
				// This should not panic even with nil context
				return client.HealthCheck(ctx)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			defer ResetClientForTesting()
			client, ctx, cancel := tt.setupTest(t)
			defer cancel()
			if client != nil {
				defer client.Close()
			}

			// Act & Assert - Should not panic
			assert.NotPanics(t, func() {
				err := tt.operation(ctx, client)
				if tt.expectedError != "" {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			})
		})
	}
}