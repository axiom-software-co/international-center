package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/notifications"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	sharedtesting "github.com/axiom-software-co/international-center/src/backend/internal/shared/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockServiceInvocationForProxy provides enhanced mock for service proxy unit tests
type MockServiceInvocationForProxy struct {
	responses           map[string]*dapr.ServiceResponse
	failures            map[string]error
	healthChecks        map[string]bool
	invocations         []ProxyMockInvocation
	lastInvocationError error
}

type ProxyMockInvocation struct {
	Method      string
	HTTPVerb    string
	Data        []byte
	Timestamp   time.Time
	ServiceType string // "content", "services", "notification"
}

func NewMockServiceInvocationForProxy() *MockServiceInvocationForProxy {
	return &MockServiceInvocationForProxy{
		responses:    make(map[string]*dapr.ServiceResponse),
		failures:     make(map[string]error),
		healthChecks: make(map[string]bool),
		invocations:  make([]ProxyMockInvocation, 0),
	}
}

// SetResponse sets a mock response for a specific method
func (m *MockServiceInvocationForProxy) SetResponse(method string, response *dapr.ServiceResponse) {
	m.responses[method] = response
}

// SetFailure sets a mock failure for a specific method
func (m *MockServiceInvocationForProxy) SetFailure(method string, err error) {
	m.failures[method] = err
}

// SetHealthStatus sets health check status for a service
func (m *MockServiceInvocationForProxy) SetHealthStatus(service string, healthy bool) {
	m.healthChecks[service] = healthy
}

// GetInvocations returns all recorded invocations
func (m *MockServiceInvocationForProxy) GetInvocations() []ProxyMockInvocation {
	return m.invocations
}

// GetLastInvocation returns the most recent invocation
func (m *MockServiceInvocationForProxy) GetLastInvocation() *ProxyMockInvocation {
	if len(m.invocations) == 0 {
		return nil
	}
	return &m.invocations[len(m.invocations)-1]
}

// InvokeContentAPI mocks content API invocation for proxy tests
func (m *MockServiceInvocationForProxy) InvokeContentAPI(ctx context.Context, method, httpVerb string, data []byte) (*dapr.ServiceResponse, error) {
	// Record invocation
	m.invocations = append(m.invocations, ProxyMockInvocation{
		Method:      method,
		HTTPVerb:    httpVerb,
		Data:        data,
		Timestamp:   time.Now(),
		ServiceType: "content",
	})

	// Check for failures first
	if err, exists := m.failures[method]; exists {
		m.lastInvocationError = err
		return nil, err
	}

	// Return mock response if available
	if response, exists := m.responses[method]; exists {
		return response, nil
	}

	// Default response
	return &dapr.ServiceResponse{
		Data:        []byte(`{"message": "content response"}`),
		ContentType: "application/json",
		StatusCode:  200,
		Headers:     make(map[string]string),
	}, nil
}

// InvokeInquiriesAPI mocks inquiries API invocation for proxy tests
func (m *MockServiceInvocationForProxy) InvokeInquiriesAPI(ctx context.Context, method, httpVerb string, data []byte) (*dapr.ServiceResponse, error) {
	// Record invocation
	m.invocations = append(m.invocations, ProxyMockInvocation{
		Method:      method,
		HTTPVerb:    httpVerb,
		Data:        data,
		Timestamp:   time.Now(),
		ServiceType: "services",
	})

	// Check for failures first
	if err, exists := m.failures[method]; exists {
		m.lastInvocationError = err
		return nil, err
	}

	// Return mock response if available
	if response, exists := m.responses[method]; exists {
		return response, nil
	}

	// Default response
	return &dapr.ServiceResponse{
		Data:        []byte(`{"message": "services response"}`),
		ContentType: "application/json",
		StatusCode:  200,
		Headers:     make(map[string]string),
	}, nil
}

// InvokeNewsAPI mocks news API invocation for proxy tests
func (m *MockServiceInvocationForProxy) InvokeNewsAPI(ctx context.Context, method, httpVerb string, data []byte) (*dapr.ServiceResponse, error) {
	// Record invocation
	m.invocations = append(m.invocations, ProxyMockInvocation{
		Method:      method,
		HTTPVerb:    httpVerb,
		Data:        data,
		Timestamp:   time.Now(),
		ServiceType: "news",
	})

	// Check for failures first
	if err, exists := m.failures[method]; exists {
		m.lastInvocationError = err
		return nil, err
	}

	// Return mock response if available
	if response, exists := m.responses[method]; exists {
		return response, nil
	}

	// Default response
	return &dapr.ServiceResponse{
		Data:        []byte(`{"message": "news response"}`),
		ContentType: "application/json",
		StatusCode:  200,
		Headers:     make(map[string]string),
	}, nil
}

// InvokeNotificationAPI mocks notification API invocation for proxy tests
func (m *MockServiceInvocationForProxy) InvokeNotificationAPI(ctx context.Context, method, httpVerb string, data []byte) (*dapr.ServiceResponse, error) {
	// Record invocation
	m.invocations = append(m.invocations, ProxyMockInvocation{
		Method:      method,
		HTTPVerb:    httpVerb,
		Data:        data,
		Timestamp:   time.Now(),
		ServiceType: "notification",
	})

	// Check for failures first
	if err, exists := m.failures[method]; exists {
		m.lastInvocationError = err
		return nil, err
	}

	// Return mock response if available
	if response, exists := m.responses[method]; exists {
		return response, nil
	}

	// Default response based on endpoint
	switch {
	case strings.Contains(method, "/subscribers"):
		subscribersResponse := map[string]interface{}{
			"subscribers": []map[string]interface{}{
				{
					"id":     "sub-123",
					"email":  "user@example.com",
					"status": "active",
					"preferences": map[string]interface{}{
						"email": true,
						"sms":   false,
					},
				},
			},
			"total": 1,
		}
		responseData, _ := json.Marshal(subscribersResponse)
		return &dapr.ServiceResponse{
			Data:        responseData,
			ContentType: "application/json",
			StatusCode:  200,
			Headers:     make(map[string]string),
		}, nil

	case strings.Contains(method, "/templates"):
		templatesResponse := map[string]interface{}{
			"templates": []map[string]interface{}{
				{
					"id":          "tpl-456",
					"name":        "Welcome Email",
					"subject":     "Welcome to our service",
					"type":        "email",
					"is_active":   true,
				},
			},
		}
		responseData, _ := json.Marshal(templatesResponse)
		return &dapr.ServiceResponse{
			Data:        responseData,
			ContentType: "application/json",
			StatusCode:  200,
			Headers:     make(map[string]string),
		}, nil

	default:
		// Default notification response
		notificationResponse := map[string]interface{}{
			"notifications": []map[string]interface{}{
				{
					"id":        "notif-789",
					"type":      "email",
					"status":    "sent",
					"recipient": "user@example.com",
					"subject":   "Test Notification",
				},
			},
		}
		responseData, _ := json.Marshal(notificationResponse)
		return &dapr.ServiceResponse{
			Data:        responseData,
			ContentType: "application/json",
			StatusCode:  200,
			Headers:     make(map[string]string),
		}, nil
	}
}

// Health check methods
func (m *MockServiceInvocationForProxy) CheckContentAPIHealth(ctx context.Context) (bool, error) {
	if err, exists := m.failures["CheckContentAPIHealth"]; exists {
		return false, err
	}
	if healthy, exists := m.healthChecks["content-api"]; exists {
		return healthy, nil
	}
	return true, nil
}

func (m *MockServiceInvocationForProxy) CheckInquiriesAPIHealth(ctx context.Context) (bool, error) {
	if err, exists := m.failures["CheckInquiriesAPIHealth"]; exists {
		return false, err
	}
	if healthy, exists := m.healthChecks["inquiries-api"]; exists {
		return healthy, nil
	}
	return true, nil
}

func (m *MockServiceInvocationForProxy) CheckNewsAPIHealth(ctx context.Context) (bool, error) {
	if err, exists := m.failures["CheckNewsAPIHealth"]; exists {
		return false, err
	}
	if healthy, exists := m.healthChecks["news-api"]; exists {
		return healthy, nil
	}
	return true, nil
}

func (m *MockServiceInvocationForProxy) CheckNotificationAPIHealth(ctx context.Context) (bool, error) {
	if err, exists := m.failures["CheckNotificationAPIHealth"]; exists {
		return false, err
	}
	if healthy, exists := m.healthChecks["notification-api"]; exists {
		return healthy, nil
	}
	return true, nil
}

// Metrics methods
func (m *MockServiceInvocationForProxy) GetContentAPIMetrics(ctx context.Context) (map[string]interface{}, error) {
	if err, exists := m.failures["GetContentAPIMetrics"]; exists {
		return nil, err
	}
	return map[string]interface{}{
		"status":   "healthy",
		"requests": len(m.getInvocationsByType("content")),
		"uptime":   "2h15m",
	}, nil
}

func (m *MockServiceInvocationForProxy) GetInquiriesAPIMetrics(ctx context.Context) (map[string]interface{}, error) {
	if err, exists := m.failures["GetInquiriesAPIMetrics"]; exists {
		return nil, err
	}
	return map[string]interface{}{
		"status":   "healthy",
		"requests": len(m.getInvocationsByType("services")),
		"uptime":   "2h10m",
	}, nil
}

func (m *MockServiceInvocationForProxy) GetNewsAPIMetrics(ctx context.Context) (map[string]interface{}, error) {
	if err, exists := m.failures["GetNewsAPIMetrics"]; exists {
		return nil, err
	}
	return map[string]interface{}{
		"status":   "healthy",
		"requests": len(m.getInvocationsByType("news")),
		"uptime":   "1h45m",
	}, nil
}

func (m *MockServiceInvocationForProxy) GetNotificationAPIMetrics(ctx context.Context) (map[string]interface{}, error) {
	if err, exists := m.failures["GetNotificationAPIMetrics"]; exists {
		return nil, err
	}
	return map[string]interface{}{
		"status":             "healthy",
		"requests":           len(m.getInvocationsByType("notification")),
		"uptime":             "2h05m",
		"notifications_sent": 42,
		"active_subscribers": 156,
		"templates_active":   8,
	}, nil
}

// Helper method to filter invocations by service type
func (m *MockServiceInvocationForProxy) getInvocationsByType(serviceType string) []ProxyMockInvocation {
	var filtered []ProxyMockInvocation
	for _, inv := range m.invocations {
		if inv.ServiceType == serviceType {
			filtered = append(filtered, inv)
		}
	}
	return filtered
}

// Test helper functions

func createProxyTestConfiguration() *GatewayConfiguration {
	return &GatewayConfiguration{
		Name:        "proxy-test-gateway",
		Type:        GatewayTypeAdmin,
		Port:        8080,
		Environment: "test",
		Version:     "1.0.0",
		
		ServiceRouting: ServiceRoutingConfig{
			ContentAPIEnabled:      true,
			ServicesAPIEnabled:     true,
			NotificationAPIEnabled: true,
			NewsAPIEnabled:         false,
		},
		
		CacheControl: CacheControlConfig{
			Enabled: false, // Disabled for testing
			MaxAge:  0,
		},
		
		Timeouts: TimeoutConfig{
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			RequestTimeout:  30 * time.Second,
			ShutdownTimeout: 15 * time.Second,
		},
	}
}

func createTestServiceProxyWithMock(mock *MockServiceInvocationForProxy) *ServiceProxy {
	config := createProxyTestConfiguration()
	return NewServiceProxyWithInvocation(mock, config)
}

// Unit Tests for ServiceProxy Notification API Integration

func TestServiceProxy_InvokeNotificationAPI_SubscriberManagement(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		httpVerb         string
		requestData      interface{}
		setupMock        func(*MockServiceInvocationForProxy)
		expectedError    string
		validateResponse func(*testing.T, interface{}, *MockServiceInvocationForProxy)
	}{
		{
			name:     "successfully get all subscribers",
			method:   "/api/v1/notifications/subscribers",
			httpVerb: "GET",
			setupMock: func(mock *MockServiceInvocationForProxy) {
				// Use default mock response for subscribers
			},
			validateResponse: func(t *testing.T, response interface{}, mock *MockServiceInvocationForProxy) {
				// Verify invocation was recorded
				invocations := mock.GetInvocations()
				assert.Len(t, invocations, 1)
				assert.Equal(t, "notification", invocations[0].ServiceType)
				assert.Equal(t, "/api/v1/notifications/subscribers", invocations[0].Method)
				assert.Equal(t, "GET", invocations[0].HTTPVerb)

				// Verify response structure
				responseMap, ok := response.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, responseMap, "subscribers")
				
				subscribers, ok := responseMap["subscribers"].([]interface{})
				require.True(t, ok)
				assert.Len(t, subscribers, 1)
			},
		},
		{
			name:     "successfully create new subscriber",
			method:   "/api/v1/notifications/subscribers",
			httpVerb: "POST",
			requestData: CreateSubscriberRequest{
				SubscriberName: "Test User",
				Email: "newuser@example.com",
				EventTypes: []notifications.EventType{notifications.EventTypeInquiryBusiness},
				NotificationMethods: []notifications.NotificationMethod{notifications.NotificationMethodEmail},
				NotificationSchedule: notifications.ScheduleImmediate,
				PriorityThreshold: notifications.PriorityMedium,
				CreatedBy: "test-user",
			},
			setupMock: func(mock *MockServiceInvocationForProxy) {
				createdSubscriber := map[string]interface{}{
					"subscriber": map[string]interface{}{
						"id":      "sub-new",
						"email":   "newuser@example.com",
						"status":  "active",
						"created": time.Now().UTC().Format(time.RFC3339),
						"preferences": map[string]interface{}{
							"email": true,
							"sms":   false,
							"slack": false,
						},
					},
				}
				responseData, _ := json.Marshal(createdSubscriber)
				mock.SetResponse("/api/v1/notifications/subscribers", &dapr.ServiceResponse{
					Data:        responseData,
					ContentType: "application/json",
					StatusCode:  201,
					Headers:     make(map[string]string),
				})
			},
			validateResponse: func(t *testing.T, response interface{}, mock *MockServiceInvocationForProxy) {
				// Verify POST request was made with data
				lastInvocation := mock.GetLastInvocation()
				require.NotNil(t, lastInvocation)
				assert.Equal(t, "POST", lastInvocation.HTTPVerb)
				assert.NotEmpty(t, lastInvocation.Data)

				// Verify request data structure
				var requestData CreateSubscriberRequest
				err := json.Unmarshal(lastInvocation.Data, &requestData)
				require.NoError(t, err)
				assert.Equal(t, "newuser@example.com", requestData.Email)
				assert.Equal(t, "Test User", requestData.SubscriberName)
				assert.Contains(t, requestData.NotificationMethods, notifications.NotificationMethodEmail)

				// Verify response
				responseMap, ok := response.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, responseMap, "subscriber")
			},
		},
		{
			name:     "successfully update subscriber preferences",
			method:   "/api/v1/notifications/subscribers/sub-123",
			httpVerb: "PUT",
			requestData: func() UpdateSubscriberRequest {
				status := notifications.SubscriberStatusActive
				return UpdateSubscriberRequest{
					NotificationMethods: []notifications.NotificationMethod{
						notifications.NotificationMethodEmail,
						notifications.NotificationMethodSMS,
					},
					Status: &status,
					UpdatedBy: "test-user",
				}
			}(),
			setupMock: func(mock *MockServiceInvocationForProxy) {
				updatedSubscriber := map[string]interface{}{
					"subscriber": map[string]interface{}{
						"id":     "sub-123",
						"email":  "user@example.com",
						"status": "active",
						"preferences": map[string]interface{}{
							"email": true,
							"sms":   true,
							"slack": false,
						},
						"updated": time.Now().UTC().Format(time.RFC3339),
					},
				}
				responseData, _ := json.Marshal(updatedSubscriber)
				mock.SetResponse("/api/v1/notifications/subscribers/sub-123", &dapr.ServiceResponse{
					Data:        responseData,
					ContentType: "application/json",
					StatusCode:  200,
					Headers:     make(map[string]string),
				})
			},
			validateResponse: func(t *testing.T, response interface{}, mock *MockServiceInvocationForProxy) {
				// Verify PUT request with correct path
				lastInvocation := mock.GetLastInvocation()
				require.NotNil(t, lastInvocation)
				assert.Equal(t, "/api/v1/notifications/subscribers/sub-123", lastInvocation.Method)
				assert.Equal(t, "PUT", lastInvocation.HTTPVerb)

				// Verify response
				responseMap, ok := response.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, responseMap, "subscriber")
			},
		},
		{
			name:     "handle notification service failure",
			method:   "/api/v1/notifications/subscribers",
			httpVerb: "GET",
			setupMock: func(mock *MockServiceInvocationForProxy) {
				mock.SetFailure("/api/v1/notifications/subscribers", domain.NewDependencyError("notification service unavailable", nil))
			},
			expectedError: "notification service unavailable",
		},
		{
			name:     "handle invalid subscriber data",
			method:   "/api/v1/notifications/subscribers",
			httpVerb: "POST",
			requestData: CreateSubscriberRequest{
				Email: "", // Invalid empty email
			},
			setupMock: func(mock *MockServiceInvocationForProxy) {
				mock.SetFailure("/api/v1/notifications/subscribers", domain.NewValidationError("email is required"))
			},
			expectedError: "email is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			mockInvocation := NewMockServiceInvocationForProxy()
			tt.setupMock(mockInvocation)

			serviceProxy := createTestServiceProxyWithMock(mockInvocation)

			// Prepare request data
			var requestData []byte
			if tt.requestData != nil {
				var err error
				requestData, err = json.Marshal(tt.requestData)
				require.NoError(t, err)
			}

			// Act
			response, err := serviceProxy.invokeNotificationAPI(ctx, tt.httpVerb, tt.method, requestData, make(map[string]string))

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, response)

				if tt.validateResponse != nil {
					tt.validateResponse(t, response, mockInvocation)
				}
			}
		})
	}
}

func TestServiceProxy_InvokeNotificationAPI_Templates(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		httpVerb         string
		setupMock        func(*MockServiceInvocationForProxy)
		validateResponse func(*testing.T, interface{}, *MockServiceInvocationForProxy)
	}{
		{
			name:     "successfully get notification templates",
			method:   "/api/v1/notifications/templates",
			httpVerb: "GET",
			setupMock: func(mock *MockServiceInvocationForProxy) {
				// Use default mock response for templates
			},
			validateResponse: func(t *testing.T, response interface{}, mock *MockServiceInvocationForProxy) {
				// Verify invocation
				invocations := mock.GetInvocations()
				assert.Len(t, invocations, 1)
				assert.Equal(t, "notification", invocations[0].ServiceType)

				// Verify templates response structure
				responseMap, ok := response.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, responseMap, "templates")

				templates, ok := responseMap["templates"].([]interface{})
				require.True(t, ok)
				assert.Len(t, templates, 1)

				template := templates[0].(map[string]interface{})
				assert.Contains(t, template, "id")
				assert.Contains(t, template, "name")
				assert.Contains(t, template, "type")
				assert.Equal(t, "email", template["type"])
			},
		},
		{
			name:     "successfully get specific template",
			method:   "/api/v1/notifications/templates/tpl-456",
			httpVerb: "GET",
			setupMock: func(mock *MockServiceInvocationForProxy) {
				templateResponse := map[string]interface{}{
					"template": map[string]interface{}{
						"id":         "tpl-456",
						"name":       "Welcome Email",
						"subject":    "Welcome to our service",
						"type":       "email",
						"content":    "Welcome {{.Name}} to our service!",
						"is_active":  true,
						"created_at": time.Now().UTC().Format(time.RFC3339),
					},
				}
				responseData, _ := json.Marshal(templateResponse)
				mock.SetResponse("/api/v1/notifications/templates/tpl-456", &dapr.ServiceResponse{
					Data:        responseData,
					ContentType: "application/json",
					StatusCode:  200,
					Headers:     make(map[string]string),
				})
			},
			validateResponse: func(t *testing.T, response interface{}, mock *MockServiceInvocationForProxy) {
				lastInvocation := mock.GetLastInvocation()
				require.NotNil(t, lastInvocation)
				assert.Equal(t, "/api/v1/notifications/templates/tpl-456", lastInvocation.Method)

				responseMap, ok := response.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, responseMap, "template")

				template := responseMap["template"].(map[string]interface{})
				assert.Equal(t, "tpl-456", template["id"])
				assert.Equal(t, "Welcome Email", template["name"])
				assert.Contains(t, template, "content")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			mockInvocation := NewMockServiceInvocationForProxy()
			tt.setupMock(mockInvocation)

			serviceProxy := createTestServiceProxyWithMock(mockInvocation)

			// Act
			response, err := serviceProxy.invokeNotificationAPI(ctx, tt.httpVerb, tt.method, nil, make(map[string]string))

			// Assert
			require.NoError(t, err)
			assert.NotNil(t, response)

			if tt.validateResponse != nil {
				tt.validateResponse(t, response, mockInvocation)
			}
		})
	}
}

func TestServiceProxy_HealthCheck_WithNotificationService(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockServiceInvocationForProxy)
		expectedError  string
		validateHealth func(*testing.T, *MockServiceInvocationForProxy)
	}{
		{
			name: "all services healthy including notification service",
			setupMock: func(mock *MockServiceInvocationForProxy) {
				mock.SetHealthStatus("content-api", true)
				mock.SetHealthStatus("inquiries-api", true)
				mock.SetHealthStatus("notification-api", true)
			},
			validateHealth: func(t *testing.T, mock *MockServiceInvocationForProxy) {
				// Verify no error and all health checks were called
			},
		},
		{
			name: "notification service unhealthy",
			setupMock: func(mock *MockServiceInvocationForProxy) {
				mock.SetHealthStatus("content-api", true)
				mock.SetHealthStatus("inquiries-api", true)
				mock.SetHealthStatus("notification-api", false)
			},
			expectedError: "notification API health check failed",
		},
		{
			name: "notification service health check error",
			setupMock: func(mock *MockServiceInvocationForProxy) {
				mock.SetHealthStatus("content-api", true)
				mock.SetHealthStatus("inquiries-api", true)
				mock.SetFailure("CheckNotificationAPIHealth", fmt.Errorf("connection timeout"))
			},
			expectedError: "notification API health check failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			mockInvocation := NewMockServiceInvocationForProxy()
			tt.setupMock(mockInvocation)

			serviceProxy := createTestServiceProxyWithMock(mockInvocation)

			// Act
			err := serviceProxy.HealthCheck(ctx)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				if tt.validateHealth != nil {
					tt.validateHealth(t, mockInvocation)
				}
			}
		})
	}
}

func TestServiceProxy_GetServiceMetrics_WithNotificationService(t *testing.T) {
	tests := []struct {
		name            string
		setupMock       func(*MockServiceInvocationForProxy)
		expectedError   string
		validateMetrics func(*testing.T, map[string]interface{}, *MockServiceInvocationForProxy)
	}{
		{
			name: "successfully get metrics from all services including notification service",
			setupMock: func(mock *MockServiceInvocationForProxy) {
				// Simulate some invocations to verify metrics
				mock.InvokeNotificationAPI(context.Background(), "/api/v1/notifications/subscribers", "GET", nil)
				mock.InvokeNotificationAPI(context.Background(), "/api/v1/notifications/templates", "GET", nil)
			},
			validateMetrics: func(t *testing.T, metrics map[string]interface{}, mock *MockServiceInvocationForProxy) {
				// Verify gateway metrics structure
				assert.Contains(t, metrics, "gateway")
				gatewayMetrics := metrics["gateway"].(map[string]interface{})
				assert.Contains(t, gatewayMetrics, "uptime")
				assert.Contains(t, gatewayMetrics, "version")

				// Verify notification service metrics if present
				if notifMetrics, exists := metrics["notification_api"]; exists {
					notifMap := notifMetrics.(map[string]interface{})
					assert.Contains(t, notifMap, "status")
					assert.Contains(t, notifMap, "notifications_sent")
					assert.Contains(t, notifMap, "active_subscribers")
					assert.Equal(t, "healthy", notifMap["status"])
					
					// Verify request count reflects actual invocations
					assert.Equal(t, 2, notifMap["requests"]) // 2 invocations made in setup
				}
			},
		},
		{
			name: "handle notification service metrics failure gracefully",
			setupMock: func(mock *MockServiceInvocationForProxy) {
				mock.SetFailure("GetNotificationAPIMetrics", fmt.Errorf("metrics service unavailable"))
			},
			validateMetrics: func(t *testing.T, metrics map[string]interface{}, mock *MockServiceInvocationForProxy) {
				// Should still return gateway metrics even if notification metrics fail
				assert.Contains(t, metrics, "gateway")
				
				// Notification metrics might not be present, but that's ok
				// The gateway should handle service-specific failures gracefully
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			mockInvocation := NewMockServiceInvocationForProxy()
			tt.setupMock(mockInvocation)

			serviceProxy := createTestServiceProxyWithMock(mockInvocation)

			// Act
			metrics, err := serviceProxy.GetServiceMetrics(ctx)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, metrics)
				
				if tt.validateMetrics != nil {
					tt.validateMetrics(t, metrics, mockInvocation)
				}
			}
		})
	}
}

func TestServiceProxy_ProxyRequest_NotificationRouting(t *testing.T) {
	tests := []struct {
		name           string
		requestPath    string
		targetService  string
		setupMock      func(*MockServiceInvocationForProxy)
		expectedError  string
		validateResult func(*testing.T, *httptest.ResponseRecorder, *MockServiceInvocationForProxy)
	}{
		{
			name:          "route notification API request correctly",
			requestPath:   "/api/v1/notifications",
			targetService: "notification-api",
			setupMock:     func(mock *MockServiceInvocationForProxy) {},
			validateResult: func(t *testing.T, recorder *httptest.ResponseRecorder, mock *MockServiceInvocationForProxy) {
				// Verify the request was routed to notification service
				invocations := mock.GetInvocations()
				assert.Len(t, invocations, 1)
				assert.Equal(t, "notification", invocations[0].ServiceType)
				assert.Equal(t, "/api/v1/notifications", invocations[0].Method)
				
				// Verify response headers
				assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
				assert.NotEmpty(t, recorder.Header().Get("X-Correlation-ID"))
			},
		},
		{
			name:          "route subscriber management request correctly",
			requestPath:   "/api/v1/notifications/subscribers/123",
			targetService: "notification-api",
			setupMock:     func(mock *MockServiceInvocationForProxy) {},
			validateResult: func(t *testing.T, recorder *httptest.ResponseRecorder, mock *MockServiceInvocationForProxy) {
				invocations := mock.GetInvocations()
				assert.Len(t, invocations, 1)
				assert.Equal(t, "/api/v1/notifications/subscribers/123", invocations[0].Method)
			},
		},
		{
			name:          "handle unknown notification service path",
			requestPath:   "/api/v1/notifications/unknown/endpoint",
			targetService: "notification-api",
			setupMock:     func(mock *MockServiceInvocationForProxy) {},
			validateResult: func(t *testing.T, recorder *httptest.ResponseRecorder, mock *MockServiceInvocationForProxy) {
				// Should still attempt to route - let the notification service handle unknown paths
				invocations := mock.GetInvocations()
				assert.Len(t, invocations, 1)
				assert.Equal(t, "/api/v1/notifications/unknown/endpoint", invocations[0].Method)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			mockInvocation := NewMockServiceInvocationForProxy()
			tt.setupMock(mockInvocation)

			serviceProxy := createTestServiceProxyWithMock(mockInvocation)

			// Create HTTP request
			req := httptest.NewRequest("GET", tt.requestPath, nil)
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
					tt.validateResult(t, recorder, mockInvocation)
				}
			}
		})
	}
}