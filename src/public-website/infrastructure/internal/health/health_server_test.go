package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHealthServer(t *testing.T) {
	// Arrange
	testCases := []struct {
		name        string
		port        int
		environment string
	}{
		{
			name:        "Valid development server on port 8080",
			port:        8080,
			environment: "development",
		},
		{
			name:        "Valid staging server on port 9000",
			port:        9000,
			environment: "staging",
		},
		{
			name:        "Valid production server on port 8000",
			port:        8000,
			environment: "production",
		},
		{
			name:        "Server with port 0",
			port:        0,
			environment: "development",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			server := NewHealthServer(tc.port, tc.environment)

			// Assert
			assert.NotNil(t, server)
			assert.Equal(t, tc.port, server.port)
			assert.NotNil(t, server.orchestrator)
			assert.Nil(t, server.server) // HTTP server not started yet
		})
	}
}

func TestHealthServer_StartAndStop(t *testing.T) {
	// Arrange
	server := NewHealthServer(0, "development") // Use port 0 for automatic assignment
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Act - Start server
	err := server.Start(ctx)
	require.NoError(t, err)
	
	// Assert - Server should be running
	assert.NotNil(t, server.server)

	// Act - Stop server
	err = server.Stop(ctx)

	// Assert - Should stop without error
	assert.NoError(t, err)
}

func TestHealthServer_handleStatus(t *testing.T) {
	// Arrange
	server := NewHealthServer(8080, "development")
	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()

	// Act
	server.handleStatus(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Contains(t, response, "container_status")
	assert.Contains(t, response, "timestamp")
	
	// Verify container_status is a map
	containerStatus, ok := response["container_status"].(map[string]interface{})
	assert.True(t, ok)
	assert.NotNil(t, containerStatus)
}

func TestHealthServer_handleOverallHealth(t *testing.T) {
	// Arrange
	server := NewHealthServer(8080, "development")
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Act
	server.handleOverallHealth(w, req)

	// Assert
	assert.Contains(t, []int{http.StatusOK, http.StatusServiceUnavailable}, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response OverallHealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.NotEmpty(t, response.Status)
	assert.Equal(t, "development", response.Environment)
	assert.Greater(t, response.TotalServices, 0)
	assert.GreaterOrEqual(t, response.HealthyCount, 0)
	assert.GreaterOrEqual(t, response.UnhealthyCount, 0)
	assert.Equal(t, response.TotalServices, response.HealthyCount+response.UnhealthyCount)
	assert.NotNil(t, response.Services)
	assert.False(t, response.Timestamp.IsZero())
}

func TestHealthServer_handleComponentHealth(t *testing.T) {
	// Arrange
	testCases := []struct {
		name          string
		path          string
		expectedCode  int
		componentName string
	}{
		{
			name:          "Database component health check",
			path:          "/health/database_connection_string",
			expectedCode:  http.StatusServiceUnavailable, // In unit tests, containers won't be running
			componentName: "database_connection_string",
		},
		{
			name:          "Storage component health check",
			path:          "/health/storage_connection_string",
			expectedCode:  http.StatusServiceUnavailable,
			componentName: "storage_connection_string",
		},
		{
			name:          "Vault component health check",
			path:          "/health/vault_address",
			expectedCode:  http.StatusServiceUnavailable,
			componentName: "vault_address",
		},
		{
			name:          "Unknown component health check",
			path:          "/health/unknown_component",
			expectedCode:  http.StatusServiceUnavailable,
			componentName: "unknown_component",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			server := NewHealthServer(8080, "development")
			req := httptest.NewRequest("GET", tc.path, nil)
			w := httptest.NewRecorder()

			// Act
			server.handleComponentHealth(w, req)

			// Assert
			assert.Equal(t, tc.expectedCode, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response HealthResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			
			assert.Equal(t, tc.componentName, response.Component)
			assert.NotEmpty(t, response.Status)
			assert.NotNil(t, response.Details)
			assert.False(t, response.Timestamp.IsZero())
		})
	}
}

func TestHealthServer_checkComponentHealth(t *testing.T) {
	// Arrange
	testCases := []struct {
		name         string
		component    string
		expectedHealthy bool
	}{
		{
			name:         "Known component - database",
			component:    "database_connection_string",
			expectedHealthy: false, // In unit tests, containers won't be running
		},
		{
			name:         "Known component - storage",
			component:    "storage_connection_string",
			expectedHealthy: false,
		},
		{
			name:         "Known component - vault",
			component:    "vault_address",
			expectedHealthy: false,
		},
		{
			name:         "Unknown component",
			component:    "unknown_component",
			expectedHealthy: false,
		},
		{
			name:         "Empty component name",
			component:    "",
			expectedHealthy: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			server := NewHealthServer(8080, "development")

			// Act
			response := server.checkComponentHealth(tc.component)

			// Assert
			assert.Equal(t, tc.component, response.Component)
			assert.Equal(t, tc.expectedHealthy, response.Healthy)
			assert.NotEmpty(t, response.Status)
			assert.NotNil(t, response.Details)
			assert.False(t, response.Timestamp.IsZero())
			
			if tc.expectedHealthy {
				assert.Equal(t, "healthy", response.Status)
			} else {
				assert.Contains(t, []string{"unhealthy", "unknown"}, response.Status)
			}
		})
	}
}

func TestHealthServer_checkOverallHealth(t *testing.T) {
	// Arrange
	server := NewHealthServer(8080, "development")

	// Act
	response := server.checkOverallHealth()

	// Assert
	assert.NotEmpty(t, response.Status)
	assert.Equal(t, "development", response.Environment)
	assert.Greater(t, response.TotalServices, 0)
	assert.GreaterOrEqual(t, response.HealthyCount, 0)
	assert.GreaterOrEqual(t, response.UnhealthyCount, 0)
	assert.Equal(t, response.TotalServices, response.HealthyCount+response.UnhealthyCount)
	assert.NotNil(t, response.Services)
	assert.False(t, response.Timestamp.IsZero())

	// In unit tests, most services will be unhealthy since containers aren't running
	expectedComponents := []string{
		"database_connection_string",
		"storage_connection_string",
		"vault_address",
		"rabbitmq_endpoint",
		"grafana_url",
		"website_url",
	}
	
	for _, component := range expectedComponents {
		assert.Contains(t, response.Services, component)
		service := response.Services[component]
		assert.Equal(t, component, service.Component)
		assert.NotNil(t, service.Details)
		assert.False(t, service.Timestamp.IsZero())
	}

	// Verify status logic
	if response.HealthyCount == response.TotalServices {
		assert.Equal(t, "healthy", response.Status)
	} else if response.HealthyCount > 0 {
		assert.Equal(t, "partially_healthy", response.Status)
	} else {
		assert.Equal(t, "unhealthy", response.Status)
	}
}

func TestHealthServer_mapComponentToContainer(t *testing.T) {
	// Arrange
	server := NewHealthServer(8080, "development")
	
	testCases := []struct {
		name              string
		component         string
		expectedContainer string
	}{
		{
			name:              "Database component maps to postgres",
			component:         "database_connection_string",
			expectedContainer: "postgres",
		},
		{
			name:              "Storage component maps to azurite",
			component:         "storage_connection_string",
			expectedContainer: "azurite",
		},
		{
			name:              "Vault component maps to vault",
			component:         "vault_address",
			expectedContainer: "vault",
		},
		{
			name:              "RabbitMQ component maps to rabbitmq",
			component:         "rabbitmq_endpoint",
			expectedContainer: "rabbitmq",
		},
		{
			name:              "Grafana component maps to grafana",
			component:         "grafana_url",
			expectedContainer: "grafana",
		},
		{
			name:              "Website component maps to empty (not containerized yet)",
			component:         "website_url",
			expectedContainer: "",
		},
		{
			name:              "Unknown component maps to empty",
			component:         "unknown_component",
			expectedContainer: "",
		},
		{
			name:              "Empty component maps to empty",
			component:         "",
			expectedContainer: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := server.mapComponentToContainer(tc.component)

			// Assert
			assert.Equal(t, tc.expectedContainer, result)
		})
	}
}

func TestHealthServer_isContainerHealthy(t *testing.T) {
	// Arrange
	server := NewHealthServer(8080, "development")
	
	testCases := []struct {
		name          string
		containerName string
		expectedHealthy bool
	}{
		{
			name:          "Valid container name but not running (unit test)",
			containerName: "postgres",
			expectedHealthy: false, // In unit tests, no containers are actually running
		},
		{
			name:          "Another valid container name but not running",
			containerName: "rabbitmq",
			expectedHealthy: false,
		},
		{
			name:          "Empty container name",
			containerName: "",
			expectedHealthy: false,
		},
		{
			name:          "Non-existent container",
			containerName: "non-existent-container",
			expectedHealthy: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := server.isContainerHealthy(tc.containerName)

			// Assert
			assert.Equal(t, tc.expectedHealthy, result)
		})
	}
}

// Property-based tests
func TestHealthServer_Properties(t *testing.T) {
	// Property: All created servers should have valid orchestrators
	t.Run("Property_NewServers_AlwaysHaveValidOrchestrators", func(t *testing.T) {
		// Arrange
		environments := []string{"development", "staging", "production"}
		ports := []int{8080, 9000, 3000, 0}

		// Act & Assert
		for _, env := range environments {
			for _, port := range ports {
				server := NewHealthServer(port, env)
				assert.NotNil(t, server.orchestrator, "Server with env %s and port %d should have orchestrator", env, port)
				assert.Equal(t, port, server.port, "Server should preserve port configuration")
			}
		}
	})

	// Property: Status endpoint always returns valid JSON
	t.Run("Property_StatusEndpoint_AlwaysReturnsValidJSON", func(t *testing.T) {
		// Arrange
		server := NewHealthServer(8080, "development")
		
		// Test multiple requests
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest("GET", "/status", nil)
			w := httptest.NewRecorder()

			// Act
			server.handleStatus(w, req)

			// Assert - Property: always returns valid JSON
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err, "Status endpoint should always return valid JSON")
			assert.Contains(t, response, "container_status")
			assert.Contains(t, response, "timestamp")
		}
	})

	// Property: Component health checks always have required fields
	t.Run("Property_ComponentHealth_AlwaysHasRequiredFields", func(t *testing.T) {
		// Arrange
		server := NewHealthServer(8080, "development")
		components := []string{
			"database_connection_string",
			"storage_connection_string",
			"vault_address",
			"rabbitmq_endpoint",
			"grafana_url",
			"website_url",
			"unknown_component",
			"",
		}

		// Act & Assert - Property: all responses have required fields
		for _, component := range components {
			response := server.checkComponentHealth(component)
			
			assert.Equal(t, component, response.Component, "Component field should match input")
			assert.NotEmpty(t, response.Status, "Status should never be empty")
			assert.NotNil(t, response.Details, "Details should never be nil")
			assert.False(t, response.Timestamp.IsZero(), "Timestamp should be set")
			assert.Contains(t, []bool{true, false}, response.Healthy, "Healthy should be boolean")
		}
	})
}

// Error handling and edge cases
func TestHealthServer_ErrorHandling(t *testing.T) {
	// Test server stop when not started
	t.Run("StopServer_WhenNotStarted_DoesNotError", func(t *testing.T) {
		// Arrange
		server := NewHealthServer(8080, "development")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Act
		err := server.Stop(ctx)

		// Assert
		assert.NoError(t, err)
	})

	// Test concurrent access to health endpoints
	t.Run("HealthEndpoints_ConcurrentAccess_ThreadSafe", func(t *testing.T) {
		// Arrange
		server := NewHealthServer(8080, "development")
		
		// Act - Make concurrent requests
		responses := make(chan *httptest.ResponseRecorder, 10)
		for i := 0; i < 10; i++ {
			go func() {
				req := httptest.NewRequest("GET", "/health", nil)
				w := httptest.NewRecorder()
				server.handleOverallHealth(w, req)
				responses <- w
			}()
		}

		// Assert - All responses should be valid
		for i := 0; i < 10; i++ {
			w := <-responses
			assert.Contains(t, []int{http.StatusOK, http.StatusServiceUnavailable}, w.Code)
			
			var response OverallHealthResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err, "Concurrent request %d should return valid JSON", i)
		}
	})

	// Test invalid component names
	t.Run("ComponentHealth_InvalidNames_HandledGracefully", func(t *testing.T) {
		// Arrange
		server := NewHealthServer(8080, "development")
		invalidNames := []string{
			"../../../etc/passwd",
			"DROP TABLE users;",
			"<script>alert('xss')</script>",
			"component with spaces",
			"component/with/slashes",
		}

		// Act & Assert
		for _, name := range invalidNames {
			response := server.checkComponentHealth(name)
			assert.Equal(t, name, response.Component)
			assert.False(t, response.Healthy)
			assert.NotEmpty(t, response.Status)
			assert.NotNil(t, response.Details)
		}
	})
}

// Benchmark tests for performance validation
func BenchmarkHealthServer_NewHealthServer(b *testing.B) {
	// Act
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server := NewHealthServer(8080, "development")
		_ = server // Prevent compiler optimization
	}
}

func BenchmarkHealthServer_HandleStatus(b *testing.B) {
	// Arrange
	server := NewHealthServer(8080, "development")
	req := httptest.NewRequest("GET", "/status", nil)

	// Act
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		server.handleStatus(w, req)
	}
}

func BenchmarkHealthServer_CheckComponentHealth(b *testing.B) {
	// Arrange
	server := NewHealthServer(8080, "development")

	// Act
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server.checkComponentHealth("database_connection_string")
	}
}

func BenchmarkHealthServer_CheckOverallHealth(b *testing.B) {
	// Arrange
	server := NewHealthServer(8080, "development")

	// Act
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server.checkOverallHealth()
	}
}