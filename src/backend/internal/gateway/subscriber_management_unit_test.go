package gateway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/axiom-software-co/international-center/src/backend/internal/notifications"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
	sharedtesting "github.com/axiom-software-co/international-center/src/backend/internal/shared/testing"
	"github.com/gorilla/mux"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)



// Domain Model Validation Tests

// Gateway HTTP Handler Tests for Notification Management

func TestNotificationGatewayHandler_CreateSubscriber(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockResponse   *dapr.ServiceResponse
		mockError      error
		expectedStatus int
		expectedError  string
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successfully create subscriber via gateway",
			requestBody: CreateSubscriberRequest{
				SubscriberName:       "API Gateway Test",
				Email:                "gateway@example.com",
				EventTypes:           []notifications.EventType{notifications.EventTypeInquiryBusiness},
				NotificationMethods:  []notifications.NotificationMethod{notifications.NotificationMethodEmail},
				NotificationSchedule: notifications.ScheduleImmediate,
				PriorityThreshold:    notifications.PriorityMedium,
				CreatedBy:            "gateway-test",
			},
			mockResponse: &dapr.ServiceResponse{
				Data: []byte(`{"subscriber_id":"123","email":"gateway@example.com","status":"active"}`),
			},
			expectedStatus: http.StatusCreated,
			validateResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var response notifications.NotificationSubscriber
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "gateway@example.com", response.Email)
				assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
			},
		},
		{
			name: "fail to create subscriber - notification service error",
			requestBody: CreateSubscriberRequest{
				SubscriberName: "Failed User",
				Email:         "fail@example.com",
				EventTypes:    []notifications.EventType{notifications.EventTypeInquiryBusiness},
				NotificationMethods: []notifications.NotificationMethod{notifications.NotificationMethodEmail},
				NotificationSchedule: notifications.ScheduleImmediate,
				PriorityThreshold: notifications.PriorityMedium,
				CreatedBy: "gateway-test",
			},
			mockError:      fmt.Errorf("notification service unavailable"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "notification service unavailable",
		},
		{
			name: "fail with invalid JSON request",
			requestBody:    `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid JSON format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			
			// Create mock service invocation
			mockService := NewMockServiceInvocation()
			
			if tt.mockError != nil {
				mockService.SetFailure("InvokeNotificationAPI", tt.mockError)
			} else if tt.mockResponse != nil {
				// Convert DAPR response to expected format
				var responseData interface{}
				json.Unmarshal(tt.mockResponse.Data, &responseData)
				mockService.SetMockResponse("notification-api", "/api/v1/notifications/subscribers", responseData)
			}
			
			// Create gateway configuration
			config := NewAdminGatewayConfiguration()
			
			// Create subscriber service and handler
			subscriberService := NewDefaultSubscriberService(nil) // Using nil repository for this test
			handler := NewSubscriberHandler(subscriberService, config)
			
			var requestBody []byte
			var err error
			
			if str, ok := tt.requestBody.(string); ok {
				requestBody = []byte(str)
			} else {
				requestBody, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}
			
			req := httptest.NewRequest("POST", "/admin/api/v1/notifications/subscribers", bytes.NewBuffer(requestBody))
			req = req.WithContext(ctx)
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			
			// Act - This will fail because handler doesn't exist yet
			handler.CreateSubscriber(recorder, req)
			
			// Assert
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			if tt.expectedError != "" {
				assert.Contains(t, recorder.Body.String(), tt.expectedError)
			} else {
				if tt.validateResponse != nil {
					tt.validateResponse(t, recorder)
				}
			}
		})
	}
}

func TestNotificationGatewayHandler_GetSubscriber(t *testing.T) {
	tests := []struct {
		name           string
		subscriberID   string
		mockResponse   *dapr.ServiceResponse
		mockError      error
		expectedStatus int
		expectedError  string
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:         "successfully get subscriber via gateway",
			subscriberID: uuid.New().String(),
			mockResponse: &dapr.ServiceResponse{
				Data: []byte(`{"subscriber_id":"123","email":"test@example.com","status":"active"}`),
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var response notifications.NotificationSubscriber
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "test@example.com", response.Email)
				assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
			},
		},
		{
			name:           "fail to get non-existing subscriber",
			subscriberID:   uuid.New().String(),
			mockError:      fmt.Errorf("subscriber not found"),
			expectedStatus: http.StatusNotFound,
			expectedError:  "subscriber not found",
		},
		{
			name:           "fail with invalid subscriber ID format",
			subscriberID:   "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid subscriber ID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			
			// Create mock service invocation
			mockService := NewMockServiceInvocation()
			
			if tt.mockError != nil {
				mockService.SetFailure("InvokeNotificationAPI", tt.mockError)
			} else if tt.mockResponse != nil {
				// Convert DAPR response to expected format
				var responseData interface{}
				json.Unmarshal(tt.mockResponse.Data, &responseData)
				mockService.SetMockResponse("notification-api", "/api/v1/notifications/subscribers", responseData)
			}
			
			// Create gateway configuration
			config := NewAdminGatewayConfiguration()
			
			// Create subscriber service and handler
			subscriberService := NewDefaultSubscriberService(nil) // Using nil repository for this test
			handler := NewSubscriberHandler(subscriberService, config)
			
			req := httptest.NewRequest("GET", fmt.Sprintf("/admin/api/v1/notifications/subscribers/%s", tt.subscriberID), nil)
			req = req.WithContext(ctx)
			req = mux.SetURLVars(req, map[string]string{"id": tt.subscriberID})
			recorder := httptest.NewRecorder()
			
			// Act - This will fail because handler doesn't exist yet
			handler.GetSubscriber(recorder, req)
			
			// Assert
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			if tt.expectedError != "" {
				assert.Contains(t, recorder.Body.String(), tt.expectedError)
			} else {
				if tt.validateResponse != nil {
					tt.validateResponse(t, recorder)
				}
			}
		})
	}
}

func TestNotificationGatewayHandler_ListSubscribers(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockResponse   *dapr.ServiceResponse
		mockError      error
		expectedStatus int
		expectedError  string
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:        "successfully list all subscribers",
			queryParams: "",
			mockResponse: &dapr.ServiceResponse{
				Data: []byte(`{"subscribers":[{"subscriber_id":"123","email":"test@example.com"}],"total":1,"page":1,"page_size":10}`),
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var response struct {
					Subscribers []*notifications.NotificationSubscriber `json:"subscribers"`
					Total       int                                     `json:"total"`
					Page        int                                     `json:"page"`
					PageSize    int                                     `json:"page_size"`
				}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, len(response.Subscribers), 0)
				assert.GreaterOrEqual(t, response.Total, 0)
			},
		},
		{
			name:           "fail with notification service error",
			queryParams:    "",
			mockError:      fmt.Errorf("notification service error"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "notification service error",
		},
		{
			name:           "fail with invalid page parameter",
			queryParams:    "?page=-1",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid page parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			
			// Create mock service invocation
			mockService := NewMockServiceInvocation()
			
			if tt.mockError != nil {
				mockService.SetFailure("InvokeNotificationAPI", tt.mockError)
			} else if tt.mockResponse != nil {
				// Convert DAPR response to expected format
				var responseData interface{}
				json.Unmarshal(tt.mockResponse.Data, &responseData)
				mockService.SetMockResponse("notification-api", "/api/v1/notifications/subscribers", responseData)
			}
			
			// Create gateway configuration
			config := NewAdminGatewayConfiguration()
			
			// Create subscriber service and handler
			subscriberService := NewDefaultSubscriberService(nil) // Using nil repository for this test
			handler := NewSubscriberHandler(subscriberService, config)
			
			req := httptest.NewRequest("GET", "/admin/api/v1/notifications/subscribers"+tt.queryParams, nil)
			req = req.WithContext(ctx)
			recorder := httptest.NewRecorder()
			
			// Act - This will fail because handler doesn't exist yet
			handler.ListSubscribers(recorder, req)
			
			// Assert
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			if tt.expectedError != "" {
				assert.Contains(t, recorder.Body.String(), tt.expectedError)
			} else {
				if tt.validateResponse != nil {
					tt.validateResponse(t, recorder)
				}
			}
		})
	}
}

func TestNotificationGatewayHandler_UpdateSubscriber(t *testing.T) {
	tests := []struct {
		name           string
		subscriberID   string
		requestBody    interface{}
		mockResponse   *dapr.ServiceResponse
		mockError      error
		expectedStatus int
		expectedError  string
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:         "successfully update subscriber via gateway",
			subscriberID: uuid.New().String(),
			requestBody: func() UpdateSubscriberRequest {
				status := notifications.SubscriberStatusActive
				return UpdateSubscriberRequest{
					SubscriberName: stringPtr("Updated User"),
					Email:          stringPtr("updated@example.com"),
					Status:         &status,
					UpdatedBy:      "gateway-admin",
				}
			}(),
			mockResponse: &dapr.ServiceResponse{
				Data: []byte(`{"subscriber_id":"123","email":"updated@example.com","status":"active"}`),
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var response notifications.NotificationSubscriber
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "updated@example.com", response.Email)
			},
		},
		{
			name:           "fail to update non-existing subscriber",
			subscriberID:   uuid.New().String(),
			requestBody:    UpdateSubscriberRequest{UpdatedBy: "admin"},
			mockError:      fmt.Errorf("subscriber not found"),
			expectedStatus: http.StatusNotFound,
			expectedError:  "subscriber not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			
			// Create mock service invocation
			mockService := NewMockServiceInvocation()
			
			if tt.mockError != nil {
				mockService.SetFailure("InvokeNotificationAPI", tt.mockError)
			} else if tt.mockResponse != nil {
				// Convert DAPR response to expected format
				var responseData interface{}
				json.Unmarshal(tt.mockResponse.Data, &responseData)
				mockService.SetMockResponse("notification-api", "/api/v1/notifications/subscribers", responseData)
			}
			
			// Create gateway configuration
			config := NewAdminGatewayConfiguration()
			
			// Create subscriber service and handler
			subscriberService := NewDefaultSubscriberService(nil) // Using nil repository for this test
			handler := NewSubscriberHandler(subscriberService, config)
			
			requestBody, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)
			
			req := httptest.NewRequest("PUT", fmt.Sprintf("/admin/api/v1/notifications/subscribers/%s", tt.subscriberID), bytes.NewBuffer(requestBody))
			req = req.WithContext(ctx)
			req.Header.Set("Content-Type", "application/json")
			req = mux.SetURLVars(req, map[string]string{"id": tt.subscriberID})
			recorder := httptest.NewRecorder()
			
			// Act - This will fail because handler doesn't exist yet
			handler.UpdateSubscriber(recorder, req)
			
			// Assert
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			if tt.expectedError != "" {
				assert.Contains(t, recorder.Body.String(), tt.expectedError)
			} else {
				if tt.validateResponse != nil {
					tt.validateResponse(t, recorder)
				}
			}
		})
	}
}

func TestNotificationGatewayHandler_DeleteSubscriber(t *testing.T) {
	tests := []struct {
		name           string
		subscriberID   string
		mockError      error
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successfully delete subscriber via gateway",
			subscriberID:   uuid.New().String(),
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "fail to delete non-existing subscriber",
			subscriberID:   uuid.New().String(),
			mockError:      fmt.Errorf("subscriber not found"),
			expectedStatus: http.StatusNotFound,
			expectedError:  "subscriber not found",
		},
		{
			name:           "fail with invalid subscriber ID format",
			subscriberID:   "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid subscriber ID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			
			// Create mock service invocation
			mockService := NewMockServiceInvocation()
			
			if tt.mockError != nil {
				mockService.SetFailure("InvokeNotificationAPI", tt.mockError)
			} else {
				mockService.SetMockResponse("notification-api", "/api/v1/notifications/subscribers", map[string]interface{}{})
			}
			
			// Create gateway configuration
			config := NewAdminGatewayConfiguration()
			
			// Create subscriber service and handler
			subscriberService := NewDefaultSubscriberService(nil) // Using nil repository for this test
			handler := NewSubscriberHandler(subscriberService, config)
			
			req := httptest.NewRequest("DELETE", fmt.Sprintf("/admin/api/v1/notifications/subscribers/%s", tt.subscriberID), nil)
			req = req.WithContext(ctx)
			req = mux.SetURLVars(req, map[string]string{"id": tt.subscriberID})
			recorder := httptest.NewRecorder()
			
			// Act - This will fail because handler doesn't exist yet
			handler.DeleteSubscriber(recorder, req)
			
			// Assert
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			if tt.expectedError != "" {
				assert.Contains(t, recorder.Body.String(), tt.expectedError)
			}
		})
	}
}

func TestNotificationGatewayHandler_HealthCheck(t *testing.T) {
	tests := []struct {
		name           string
		mockError      error
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successfully get notification service health",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "fail when notification service unhealthy",
			mockError:      fmt.Errorf("notification service unhealthy"),
			expectedStatus: http.StatusServiceUnavailable,
			expectedError:  "notification service unhealthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			
			// Create gateway configuration
			config := NewAdminGatewayConfiguration()
			
			// Create subscriber service and handler
			subscriberService := NewDefaultSubscriberService(nil) // Using nil repository for this test
			handler := NewSubscriberHandler(subscriberService, config)
			
			req := httptest.NewRequest("GET", "/admin/api/v1/notifications/health", nil)
			req = req.WithContext(ctx)
			recorder := httptest.NewRecorder()
			
			// Act
			handler.SubscriberHealthCheck(recorder, req)
			
			// Assert
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			if tt.expectedError != "" {
				assert.Contains(t, recorder.Body.String(), tt.expectedError)
			}
		})
	}
}

// Helper functions
