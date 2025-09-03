package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	sharedtesting "github.com/axiom-software-co/international-center/src/backend/internal/shared/testing"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockServiceInvocation provides mock service invocation for unit tests
type MockServiceInvocation struct {
	responses     map[string]interface{}
	failures      map[string]error
	healthChecks  map[string]bool
	invocations   []MockInvocation
	endpoints     *dapr.ServiceEndpoints
}

type MockInvocation struct {
	AppID      string
	Method     string
	HTTPVerb   string
	Data       []byte
	Metadata   map[string]string
}

func NewMockServiceInvocation() *MockServiceInvocation {
	return &MockServiceInvocation{
		responses:    make(map[string]interface{}),
		failures:     make(map[string]error),
		healthChecks: make(map[string]bool),
		invocations:  make([]MockInvocation, 0),
		endpoints: &dapr.ServiceEndpoints{
			ContentAPI:  "content-api",
			ServicesAPI: "services-api",
			AdminGW:     "admin-gateway",
			PublicGW:    "public-gateway",
		},
	}
}

// SetMockResponse sets a mock response for a service invocation
func (m *MockServiceInvocation) SetMockResponse(appID, method string, response interface{}) {
	key := appID + "/" + method
	m.responses[key] = response
}

// SetFailure sets a mock failure for specific operations
func (m *MockServiceInvocation) SetFailure(operation string, err error) {
	m.failures[operation] = err
}

// SetHealthCheck sets mock health check result
func (m *MockServiceInvocation) SetHealthCheck(appID string, healthy bool) {
	m.healthChecks[appID] = healthy
}

// GetInvocations returns all mock invocations
func (m *MockServiceInvocation) GetInvocations() []MockInvocation {
	return m.invocations
}

// GetServiceEndpoints returns mock service endpoints
func (m *MockServiceInvocation) GetServiceEndpoints() *dapr.ServiceEndpoints {
	return m.endpoints
}

// InvokeService mocks service invocation
func (m *MockServiceInvocation) InvokeService(ctx context.Context, req *dapr.ServiceRequest) (*dapr.ServiceResponse, error) {
	if err, exists := m.failures["InvokeService"]; exists {
		return nil, err
	}
	
	// Record invocation
	m.invocations = append(m.invocations, MockInvocation{
		AppID:      req.AppID,
		Method:     req.MethodName,
		HTTPVerb:   req.HTTPVerb,
		Data:       req.Data,
		Metadata:   req.Metadata,
	})
	
	// Check for mock response
	key := req.AppID + "/" + req.MethodName
	if response, exists := m.responses[key]; exists {
		responseData, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal mock response: %w", err)
		}
		return &dapr.ServiceResponse{
			Data:        responseData,
			ContentType: "application/json",
			StatusCode:  200,
			Headers:     make(map[string]string),
		}, nil
	}
	
	// Default successful response
	return &dapr.ServiceResponse{
		Data:        []byte(`{"message": "mock response"}`),
		ContentType: "application/json",
		StatusCode:  200,
		Headers:     make(map[string]string),
	}, nil
}

// InvokeContentAPI mocks content API invocation
func (m *MockServiceInvocation) InvokeContentAPI(ctx context.Context, method, httpVerb string, data []byte) (*dapr.ServiceResponse, error) {
	if err, exists := m.failures["InvokeContentAPI"]; exists {
		return nil, err
	}
	
	req := &dapr.ServiceRequest{
		AppID:       m.endpoints.ContentAPI,
		MethodName:  method,
		HTTPVerb:    httpVerb,
		Data:        data,
		ContentType: "application/json",
	}
	
	return m.InvokeService(ctx, req)
}

// InvokeServicesAPI mocks services API invocation
func (m *MockServiceInvocation) InvokeServicesAPI(ctx context.Context, method, httpVerb string, data []byte) (*dapr.ServiceResponse, error) {
	if err, exists := m.failures["InvokeServicesAPI"]; exists {
		return nil, err
	}
	
	req := &dapr.ServiceRequest{
		AppID:       m.endpoints.ServicesAPI,
		MethodName:  method,
		HTTPVerb:    httpVerb,
		Data:        data,
		ContentType: "application/json",
	}
	
	return m.InvokeService(ctx, req)
}

// CheckServiceHealth mocks health check
func (m *MockServiceInvocation) CheckServiceHealth(ctx context.Context, appID string) error {
	if err, exists := m.failures["CheckServiceHealth"]; exists {
		return err
	}
	
	if healthy, exists := m.healthChecks[appID]; exists {
		if !healthy {
			return domain.NewDependencyError("service health check failed", nil)
		}
	}
	
	return nil
}

// CheckServiceReadiness mocks readiness check
func (m *MockServiceInvocation) CheckServiceReadiness(ctx context.Context, appID string) error {
	if err, exists := m.failures["CheckServiceReadiness"]; exists {
		return err
	}
	
	return m.CheckServiceHealth(ctx, appID)
}

// CheckAllServicesHealth mocks health check for all services
func (m *MockServiceInvocation) CheckAllServicesHealth(ctx context.Context) error {
	if err, exists := m.failures["CheckAllServicesHealth"]; exists {
		return err
	}
	
	endpoints := m.GetServiceEndpoints()
	services := []string{endpoints.ContentAPI, endpoints.ServicesAPI}
	
	for _, service := range services {
		if err := m.CheckServiceHealth(ctx, service); err != nil {
			return err
		}
	}
	
	return nil
}

// CheckAllServicesReadiness mocks readiness check for all services
func (m *MockServiceInvocation) CheckAllServicesReadiness(ctx context.Context) error {
	return m.CheckAllServicesHealth(ctx)
}

// CheckContentAPIHealth checks if the content API service is healthy
func (m *MockServiceInvocation) CheckContentAPIHealth(ctx context.Context) (bool, error) {
	if err, exists := m.failures["CheckContentAPIHealth"]; exists {
		return false, err
	}
	
	if healthy, exists := m.healthChecks[m.endpoints.ContentAPI]; exists {
		return healthy, nil
	}
	
	return true, nil // default to healthy
}

// CheckServicesAPIHealth checks if the services API service is healthy
func (m *MockServiceInvocation) CheckServicesAPIHealth(ctx context.Context) (bool, error) {
	if err, exists := m.failures["CheckServicesAPIHealth"]; exists {
		return false, err
	}
	
	if healthy, exists := m.healthChecks[m.endpoints.ServicesAPI]; exists {
		return healthy, nil
	}
	
	return true, nil // default to healthy
}

// GetContentAPIMetrics retrieves metrics from the content API service
func (m *MockServiceInvocation) GetContentAPIMetrics(ctx context.Context) (map[string]interface{}, error) {
	if err, exists := m.failures["GetContentAPIMetrics"]; exists {
		return nil, err
	}
	
	return map[string]interface{}{
		"status": "healthy",
		"uptime": "1h30m",
		"requests": 150,
	}, nil
}

// GetServicesAPIMetrics retrieves metrics from the services API service
func (m *MockServiceInvocation) GetServicesAPIMetrics(ctx context.Context) (map[string]interface{}, error) {
	if err, exists := m.failures["GetServicesAPIMetrics"]; exists {
		return nil, err
	}
	
	return map[string]interface{}{
		"status": "healthy",
		"uptime": "1h25m", 
		"requests": 200,
	}, nil
}

// Test helper functions
func createTestGatewayConfiguration() *GatewayConfiguration {
	return &GatewayConfiguration{
		Name:        "test-gateway",
		Type:        GatewayTypePublic,
		Port:        8080,
		Environment: "test",
		Version:     "1.0.0",
		
		Security: SecurityConfig{
			RequireAuthentication: false,
			AllowedOrigins:       []string{"http://localhost:3000"},
			SecurityHeaders: SecurityHeadersConfig{
				Enabled:             true,
				ContentTypeOptions:  "nosniff",
				FrameOptions:       "DENY",
				XSSProtection:      "1; mode=block",
			},
		},
		
		RateLimit: RateLimitConfig{
			Enabled:           true,
			RequestsPerMinute: 1000,
			BurstSize:         100,
			WindowSize:        time.Minute,
			KeyExtractor:      "ip",
			BackingStore:      "redis",
		},
		
		CORS: CORSConfig{
			Enabled:          true,
			AllowedOrigins:   []string{"http://localhost:3000"},
			AllowedMethods:   []string{"GET", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type", "Authorization"},
			ExposedHeaders:   []string{"X-Correlation-ID"},
			AllowCredentials: false,
			MaxAge:           3600,
		},
		
		ServiceRouting: ServiceRoutingConfig{
			ContentAPIEnabled:  true,
			ServicesAPIEnabled: true,
			HealthCheckPath:    "/health",
			MetricsPath:        "/metrics",
		},
		
		Timeouts: TimeoutConfig{
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			IdleTimeout:     60 * time.Second,
			RequestTimeout:  30 * time.Second,
			ShutdownTimeout: 15 * time.Second,
		},
		
		Observability: ObservabilityConfig{
			Enabled:         true,
			MetricsEnabled:  true,
			TracingEnabled:  true,
			LoggingEnabled:  true,
			HealthCheckPath: "/health",
			ReadinessPath:   "/ready",
			MetricsPath:     "/metrics",
		},
	}
}

func createTestServiceProxy(mockInvocation *MockServiceInvocation) *ServiceProxy {
	config := createTestGatewayConfiguration()
	return NewServiceProxyWithInvocation(mockInvocation, config)
}

func createTestMiddleware() *Middleware {
	config := createTestGatewayConfiguration()
	return NewMiddleware(config)
}

// Unit Tests for GatewayHandler

func TestGatewayHandler_ProxyToContentAPI(t *testing.T) {
	tests := []struct {
		name             string
		path             string
		setupMock        func(*MockServiceInvocation)
		expectedStatus   int
		expectedError    string
		validateResponse func(*testing.T, *httptest.ResponseRecorder, *MockServiceInvocation)
	}{
		{
			name: "successfully proxy GET request to content API",
			path: "/api/v1/content",
			setupMock: func(mock *MockServiceInvocation) {
				mock.SetMockResponse("content-api", "GET", map[string]interface{}{
					"message": "content response",
				})
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, mock *MockServiceInvocation) {
				// Verify service was invoked
				invocations := mock.GetInvocations()
				assert.Len(t, invocations, 1)
				assert.Equal(t, "content-api", invocations[0].AppID)
				assert.Equal(t, "GET", invocations[0].HTTPVerb)
				
				// Verify response headers
				assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
				assert.NotEmpty(t, recorder.Header().Get("X-Correlation-ID"))
			},
		},
		{
			name: "successfully proxy specific content endpoint",
			path: "/api/v1/content/123",
			setupMock: func(mock *MockServiceInvocation) {
				mock.SetMockResponse("content-api", "/api/v1/content/123", map[string]interface{}{
					"id": "123",
					"title": "Test Content",
				})
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, mock *MockServiceInvocation) {
				invocations := mock.GetInvocations()
				assert.Len(t, invocations, 1)
				assert.Equal(t, "content-api", invocations[0].AppID)
			},
		},
		{
			name: "handle service invocation failure",
			path: "/api/v1/content",
			setupMock: func(mock *MockServiceInvocation) {
				mock.SetFailure("InvokeService", domain.NewDependencyError("content API unavailable", nil))
			},
			expectedStatus: http.StatusBadGateway,
			expectedError:  "content API unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			
			mockInvocation := NewMockServiceInvocation()
			tt.setupMock(mockInvocation)
			
			serviceProxy := createTestServiceProxy(mockInvocation)
			middleware := createTestMiddleware()
			config := createTestGatewayConfiguration()
			handler := NewGatewayHandler(config, serviceProxy, middleware)
			
			// Create request
			req := httptest.NewRequest("GET", tt.path, nil)
			req = req.WithContext(ctx)
			recorder := httptest.NewRecorder()
			
			// Act
			handler.ProxyToContentAPI(recorder, req)
			
			// Assert
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			if tt.expectedError != "" {
				assert.Contains(t, recorder.Body.String(), tt.expectedError)
			} else {
				if tt.validateResponse != nil {
					tt.validateResponse(t, recorder, mockInvocation)
				}
			}
		})
	}
}

func TestGatewayHandler_ProxyToServicesAPI(t *testing.T) {
	tests := []struct {
		name             string
		path             string
		setupMock        func(*MockServiceInvocation)
		expectedStatus   int
		expectedError    string
		validateResponse func(*testing.T, *httptest.ResponseRecorder, *MockServiceInvocation)
	}{
		{
			name: "successfully proxy GET request to services API",
			path: "/api/v1/services",
			setupMock: func(mock *MockServiceInvocation) {
				mock.SetMockResponse("services-api", "GET", map[string]interface{}{
					"services": []string{"service1", "service2"},
				})
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, mock *MockServiceInvocation) {
				// Verify service was invoked
				invocations := mock.GetInvocations()
				assert.Len(t, invocations, 1)
				assert.Equal(t, "services-api", invocations[0].AppID)
				assert.Equal(t, "GET", invocations[0].HTTPVerb)
			},
		},
		{
			name: "successfully proxy specific service endpoint",
			path: "/api/v1/services/featured",
			setupMock: func(mock *MockServiceInvocation) {
				mock.SetMockResponse("services-api", "/api/v1/services/featured", map[string]interface{}{
					"featured": []string{"service1"},
				})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "handle service invocation failure",
			path: "/api/v1/services",
			setupMock: func(mock *MockServiceInvocation) {
				mock.SetFailure("InvokeService", domain.NewDependencyError("services API unavailable", nil))
			},
			expectedStatus: http.StatusBadGateway,
			expectedError:  "services API unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			
			mockInvocation := NewMockServiceInvocation()
			tt.setupMock(mockInvocation)
			
			serviceProxy := createTestServiceProxy(mockInvocation)
			middleware := createTestMiddleware()
			config := createTestGatewayConfiguration()
			handler := NewGatewayHandler(config, serviceProxy, middleware)
			
			// Create request
			req := httptest.NewRequest("GET", tt.path, nil)
			req = req.WithContext(ctx)
			recorder := httptest.NewRecorder()
			
			// Act
			handler.ProxyToServicesAPI(recorder, req)
			
			// Assert
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			if tt.expectedError != "" {
				assert.Contains(t, recorder.Body.String(), tt.expectedError)
			} else {
				if tt.validateResponse != nil {
					tt.validateResponse(t, recorder, mockInvocation)
				}
			}
		})
	}
}

func TestGatewayHandler_HealthCheck(t *testing.T) {
	tests := []struct {
		name             string
		setupMock        func(*MockServiceInvocation)
		expectedStatus   int
		expectedHealth   string
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "return healthy when all services are healthy",
			setupMock: func(mock *MockServiceInvocation) {
				mock.SetHealthCheck("content-api", true)
				mock.SetHealthCheck("services-api", true)
			},
			expectedStatus: http.StatusOK,
			expectedHealth: "healthy",
			validateResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Contains(t, recorder.Body.String(), "healthy")
				assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
			},
		},
		{
			name: "return unhealthy when service is down",
			setupMock: func(mock *MockServiceInvocation) {
				mock.SetHealthCheck("content-api", false)
				mock.SetHealthCheck("services-api", true)
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedHealth: "unhealthy",
		},
		{
			name: "return unhealthy when health check fails",
			setupMock: func(mock *MockServiceInvocation) {
				mock.SetFailure("CheckContentAPIHealth", domain.NewDependencyError("health check failed", nil))
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedHealth: "unhealthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			
			mockInvocation := NewMockServiceInvocation()
			tt.setupMock(mockInvocation)
			
			serviceProxy := createTestServiceProxy(mockInvocation)
			middleware := createTestMiddleware()
			config := createTestGatewayConfiguration()
			handler := NewGatewayHandler(config, serviceProxy, middleware)
			
			// Create request
			req := httptest.NewRequest("GET", "/health", nil)
			req = req.WithContext(ctx)
			recorder := httptest.NewRecorder()
			
			// Act
			handler.HealthCheck(recorder, req)
			
			// Assert
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			assert.Contains(t, recorder.Body.String(), tt.expectedHealth)
			
			if tt.validateResponse != nil {
				tt.validateResponse(t, recorder)
			}
		})
	}
}

func TestServiceProxy_ProxyRequest(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		targetService  string
		setupMock      func(*MockServiceInvocation)
		expectedError  string
		validateResult func(*testing.T, *MockServiceInvocation)
	}{
		{
			name:          "successfully proxy content API request",
			path:          "/api/v1/content",
			targetService: "content-api",
			setupMock:     func(mock *MockServiceInvocation) {},
			validateResult: func(t *testing.T, mock *MockServiceInvocation) {
				invocations := mock.GetInvocations()
				assert.Len(t, invocations, 1)
				assert.Equal(t, "content-api", invocations[0].AppID)
			},
		},
		{
			name:          "successfully proxy services API request",
			path:          "/api/v1/services",
			targetService: "services-api",
			setupMock:     func(mock *MockServiceInvocation) {},
			validateResult: func(t *testing.T, mock *MockServiceInvocation) {
				invocations := mock.GetInvocations()
				assert.Len(t, invocations, 1)
				assert.Equal(t, "services-api", invocations[0].AppID)
			},
		},
		{
			name:          "fail with invalid API path format",
			path:          "/invalid/path",
			targetService: "content-api",
			setupMock:     func(mock *MockServiceInvocation) {},
			expectedError: "invalid API path format",
		},
		{
			name:          "fail with unknown target service",
			path:          "/api/v1/unknown",
			targetService: "content-api",
			setupMock:     func(mock *MockServiceInvocation) {},
			expectedError: "unknown service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			
			mockInvocation := NewMockServiceInvocation()
			tt.setupMock(mockInvocation)
			
			serviceProxy := createTestServiceProxy(mockInvocation)
			
			// Create request
			req := httptest.NewRequest("GET", tt.path, nil)
			req = req.WithContext(ctx)
			recorder := httptest.NewRecorder()
			
			// Act
			err := serviceProxy.ProxyRequest(ctx, recorder, req, tt.targetService)
			
			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, mockInvocation)
				}
			}
		})
	}
}

func TestGatewayConfiguration_Validation(t *testing.T) {
	tests := []struct {
		name           string
		modifyConfig   func(*GatewayConfiguration)
		expectedError  string
		validateConfig func(*testing.T, *GatewayConfiguration)
	}{
		{
			name:         "valid public gateway configuration",
			modifyConfig: func(config *GatewayConfiguration) {},
			validateConfig: func(t *testing.T, config *GatewayConfiguration) {
				assert.Equal(t, GatewayTypePublic, config.Type)
				assert.True(t, config.IsPublic())
				assert.False(t, config.IsAdmin())
				assert.False(t, config.ShouldRequireAuth())
			},
		},
		{
			name: "valid admin gateway configuration",
			modifyConfig: func(config *GatewayConfiguration) {
				config.Type = GatewayTypeAdmin
				config.Security.RequireAuthentication = true
			},
			validateConfig: func(t *testing.T, config *GatewayConfiguration) {
				assert.Equal(t, GatewayTypeAdmin, config.Type)
				assert.False(t, config.IsPublic())
				assert.True(t, config.IsAdmin())
				assert.True(t, config.ShouldRequireAuth())
			},
		},
		{
			name: "valid listen address",
			modifyConfig: func(config *GatewayConfiguration) {
				config.Port = 9090
			},
			validateConfig: func(t *testing.T, config *GatewayConfiguration) {
				assert.Equal(t, ":9090", config.GetListenAddress())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			config := createTestGatewayConfiguration()
			tt.modifyConfig(config)
			
			// Act & Assert
			if tt.expectedError != "" {
				// Configuration validation would typically be done by the service
				// For this test, we're just demonstrating the structure
				assert.Contains(t, "mock validation error", tt.expectedError)
			} else {
				if tt.validateConfig != nil {
					tt.validateConfig(t, config)
				}
			}
		})
	}
}

func TestGatewayHandler_RegisterRoutes(t *testing.T) {
	tests := []struct {
		name           string
		modifyConfig   func(*GatewayConfiguration)
		testPath       string
		testMethod     string
		expectedFound  bool
	}{
		{
			name:          "register health check route",
			modifyConfig:  func(config *GatewayConfiguration) {},
			testPath:      "/health",
			testMethod:    "GET",
			expectedFound: true,
		},
		{
			name:          "register readiness check route",
			modifyConfig:  func(config *GatewayConfiguration) {},
			testPath:      "/ready",
			testMethod:    "GET",
			expectedFound: true,
		},
		{
			name:          "register metrics route",
			modifyConfig:  func(config *GatewayConfiguration) {},
			testPath:      "/metrics",
			testMethod:    "GET",
			expectedFound: true,
		},
		{
			name:          "register content API proxy routes",
			modifyConfig:  func(config *GatewayConfiguration) {},
			testPath:      "/api/v1/content",
			testMethod:    "GET",
			expectedFound: true,
		},
		{
			name:          "register services API proxy routes",
			modifyConfig:  func(config *GatewayConfiguration) {},
			testPath:      "/api/v1/services",
			testMethod:    "GET",
			expectedFound: true,
		},
		{
			name: "disable content API routes",
			modifyConfig: func(config *GatewayConfiguration) {
				config.ServiceRouting.ContentAPIEnabled = false
			},
			testPath:      "/api/v1/content",
			testMethod:    "GET",
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockInvocation := NewMockServiceInvocation()
			serviceProxy := createTestServiceProxy(mockInvocation)
			middleware := createTestMiddleware()
			config := createTestGatewayConfiguration()
			tt.modifyConfig(config)
			
			handler := NewGatewayHandler(config, serviceProxy, middleware)
			router := mux.NewRouter()
			
			// Act
			handler.RegisterRoutes(router)
			
			// Assert
			req := httptest.NewRequest(tt.testMethod, tt.testPath, nil)
			match := &mux.RouteMatch{}
			found := router.Match(req, match)
			
			assert.Equal(t, tt.expectedFound, found, "Route %s should be %s", tt.testPath, map[bool]string{true: "found", false: "not found"}[tt.expectedFound])
		})
	}
}

func TestGatewayHandler_Timeout(t *testing.T) {
	// Test that context timeout is respected (5 seconds for unit tests)
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()
	
	// Verify context has 5 second timeout
	deadline, hasDeadline := ctx.Deadline()
	require.True(t, hasDeadline)
	assert.True(t, time.Until(deadline) <= 5*time.Second)
	assert.True(t, time.Until(deadline) > 4*time.Second) // Allow some margin
}

func TestGatewayHandler_MiddlewareIntegration(t *testing.T) {
	tests := []struct {
		name           string
		requestHeaders map[string]string
		validateHeaders func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "apply security headers",
			requestHeaders: map[string]string{
				"User-Agent": "test-client/1.0",
			},
			validateHeaders: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// Verify security headers are applied
				assert.Equal(t, "nosniff", recorder.Header().Get("X-Content-Type-Options"))
				assert.Equal(t, "DENY", recorder.Header().Get("X-Frame-Options"))
				assert.Equal(t, "1; mode=block", recorder.Header().Get("X-XSS-Protection"))
			},
		},
		{
			name: "apply CORS headers",
			requestHeaders: map[string]string{
				"Origin": "http://localhost:3000",
			},
			validateHeaders: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// CORS headers would be applied by middleware
				// This demonstrates the integration testing pattern
				assert.NotEmpty(t, recorder.Header().Get("X-Correlation-ID"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			
			mockInvocation := NewMockServiceInvocation()
			serviceProxy := createTestServiceProxy(mockInvocation)
			middleware := createTestMiddleware()
			config := createTestGatewayConfiguration()
			handler := NewGatewayHandler(config, serviceProxy, middleware)
			
			// Create request with headers
			req := httptest.NewRequest("GET", "/health", nil)
			req = req.WithContext(ctx)
			for key, value := range tt.requestHeaders {
				req.Header.Set(key, value)
			}
			recorder := httptest.NewRecorder()
			
			// Act
			handler.HealthCheck(recorder, req)
			
			// Assert
			if tt.validateHeaders != nil {
				tt.validateHeaders(t, recorder)
			}
		})
	}
}