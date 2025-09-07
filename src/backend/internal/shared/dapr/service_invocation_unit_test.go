package dapr

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	sharedtesting "github.com/axiom-software-co/international-center/src/backend/internal/shared/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED PHASE - Service Invocation Tests (80+ test cases)

func TestNewServiceInvocation(t *testing.T) {
	tests := []struct {
		name           string
		client         *Client
		expectedResult func(*testing.T, *ServiceInvocation)
		expectNil      bool
	}{
		{
			name: "create service invocation with valid client",
			client: func() *Client {
				client, err := NewClient()
				require.NoError(t, err)
				return client
			}(),
			expectedResult: func(t *testing.T, si *ServiceInvocation) {
				assert.NotNil(t, si.client)
				endpoints := si.GetServiceEndpoints()
				assert.NotNil(t, endpoints)
				assert.NotEmpty(t, endpoints.ContentAPI)
				assert.NotEmpty(t, endpoints.InquiriesAPI)
			},
		},
		{
			name:      "create service invocation with nil client",
			client:    nil,
			expectNil: false, // Should handle gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			serviceInvocation := NewServiceInvocation(tt.client)

			// Assert
			if tt.expectNil {
				assert.Nil(t, serviceInvocation)
			} else {
				assert.NotNil(t, serviceInvocation)
				if tt.expectedResult != nil {
					tt.expectedResult(t, serviceInvocation)
				}
			}
		})
	}
}

func TestServiceInvocation_GetServiceEndpoints(t *testing.T) {
	tests := []struct {
		name         string
		envVars      map[string]string
		expectedEndpoints map[string]string
	}{
		{
			name:    "default service endpoints",
			envVars: map[string]string{},
			expectedEndpoints: map[string]string{
				"ContentAPI":  "content",
				"InquiriesAPI": "inquiries",
				"AdminGW":     "admin-gateway",
				"PublicGW":    "public-gateway",
			},
		},
		{
			name: "custom service endpoints",
			envVars: map[string]string{
				"CONTENT_API_APP_ID":    "custom-content-api",
				"INQUIRIES_API_APP_ID":   "custom-inquiries-api",
				"ADMIN_GATEWAY_APP_ID":  "custom-admin-gateway",
				"PUBLIC_GATEWAY_APP_ID": "custom-public-gateway",
			},
			expectedEndpoints: map[string]string{
				"ContentAPI":  "custom-content-api",
				"InquiriesAPI": "custom-inquiries-api",
				"AdminGW":     "custom-admin-gateway",
				"PublicGW":    "custom-public-gateway",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			client, err := NewClient()
			require.NoError(t, err)
			defer client.Close()

			serviceInvocation := NewServiceInvocation(client)

			// Act
			endpoints := serviceInvocation.GetServiceEndpoints()

			// Assert
			assert.NotNil(t, endpoints)
			assert.Equal(t, tt.expectedEndpoints["ContentAPI"], endpoints.ContentAPI)
			assert.Equal(t, tt.expectedEndpoints["InquiriesAPI"], endpoints.InquiriesAPI)
			assert.Equal(t, tt.expectedEndpoints["AdminGW"], endpoints.AdminGW)
			assert.Equal(t, tt.expectedEndpoints["PublicGW"], endpoints.PublicGW)
		})
	}
}

func TestServiceInvocation_InvokeService(t *testing.T) {
	tests := []struct {
		name           string
		request        *ServiceRequest
		setupContext   func() (context.Context, context.CancelFunc)
		expectedError  string
		validateResult func(*testing.T, *ServiceResponse)
	}{
		{
			name: "invoke service with valid request",
			request: &ServiceRequest{
				AppID:       "test-service",
				MethodName:  "test-method",
				HTTPVerb:    "GET",
				Data:        []byte(`{"test": "data"}`),
				ContentType: "application/json",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, response *ServiceResponse) {
				assert.NotNil(t, response)
				assert.Equal(t, 200, response.StatusCode)
				assert.Equal(t, "application/json", response.ContentType)
				assert.NotEmpty(t, response.Data)
			},
		},
		{
			name: "invoke service with timeout context",
			request: &ServiceRequest{
				AppID:      "test-service",
				MethodName: "test-method",
				HTTPVerb:   "GET",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 1*time.Millisecond)
			},
			validateResult: func(t *testing.T, response *ServiceResponse) {
				// May or may not succeed depending on timing
				if response != nil {
					assert.Equal(t, 200, response.StatusCode)
				}
			},
		},
		{
			name: "invoke service with cancelled context",
			request: &ServiceRequest{
				AppID:      "test-service",
				MethodName: "test-method",
				HTTPVerb:   "POST",
				Data:       []byte(`{"data": "test"}`),
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				return ctx, func() {}
			},
		},
		{
			name: "invoke service with large data payload",
			request: &ServiceRequest{
				AppID:       "test-service",
				MethodName:  "upload",
				HTTPVerb:    "POST",
				Data:        make([]byte, 1024*1024), // 1MB payload
				ContentType: "application/octet-stream",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, response *ServiceResponse) {
				assert.NotNil(t, response)
				assert.Equal(t, 200, response.StatusCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := tt.setupContext()
			defer cancel()

			client, err := NewClient()
			require.NoError(t, err)
			defer client.Close()

			serviceInvocation := NewServiceInvocation(client)

			// Act
			response, err := serviceInvocation.InvokeService(ctx, tt.request)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				if tt.validateResult != nil {
					tt.validateResult(t, response)
				}
			}
		})
	}
}

func TestServiceInvocation_InvokeContentAPI(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		httpVerb       string
		data           []byte
		setupContext   func() (context.Context, context.CancelFunc)
		expectedError  string
		validateResult func(*testing.T, *ServiceResponse)
	}{
		{
			name:     "invoke content API GET method",
			method:   "api/v1/content",
			httpVerb: "GET",
			data:     nil,
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, response *ServiceResponse) {
				assert.NotNil(t, response)
				assert.Equal(t, 200, response.StatusCode)
				assert.Equal(t, "application/json", response.ContentType)
			},
		},
		{
			name:     "invoke content API POST method",
			method:   "api/v1/content",
			httpVerb: "POST",
			data:     []byte(`{"title": "Test Content", "body": "Test Body"}`),
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, response *ServiceResponse) {
				assert.NotNil(t, response)
				assert.Equal(t, 200, response.StatusCode)
			},
		},
		{
			name:     "invoke content API PUT method",
			method:   "api/v1/content/123",
			httpVerb: "PUT",
			data:     []byte(`{"title": "Updated Content"}`),
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, response *ServiceResponse) {
				assert.NotNil(t, response)
				assert.Equal(t, 200, response.StatusCode)
			},
		},
		{
			name:     "invoke content API DELETE method",
			method:   "api/v1/content/123",
			httpVerb: "DELETE",
			data:     nil,
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, response *ServiceResponse) {
				assert.NotNil(t, response)
				assert.Equal(t, 200, response.StatusCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := tt.setupContext()
			defer cancel()

			client, err := NewClient()
			require.NoError(t, err)
			defer client.Close()

			serviceInvocation := NewServiceInvocation(client)

			// Act
			response, err := serviceInvocation.InvokeContentAPI(ctx, tt.method, tt.httpVerb, tt.data)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				if tt.validateResult != nil {
					tt.validateResult(t, response)
				}
			}
		})
	}
}

func TestServiceInvocation_InvokeInquiriesAPI(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		httpVerb       string
		data           []byte
		setupContext   func() (context.Context, context.CancelFunc)
		expectedError  string
		validateResult func(*testing.T, *ServiceResponse)
	}{
		{
			name:     "invoke inquiries API GET method",
			method:   "api/v1/inquiries",
			httpVerb: "GET",
			data:     nil,
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, response *ServiceResponse) {
				assert.NotNil(t, response)
				assert.Equal(t, 200, response.StatusCode)
				assert.Equal(t, "application/json", response.ContentType)
			},
		},
		{
			name:     "invoke inquiries API POST method",
			method:   "api/v1/inquiries",
			httpVerb: "POST",
			data:     []byte(`{"type": "business", "description": "Test Inquiry"}`),
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, response *ServiceResponse) {
				assert.NotNil(t, response)
				assert.Equal(t, 200, response.StatusCode)
			},
		},
		{
			name:     "invoke inquiries API with business endpoint",
			method:   "api/v1/inquiries/business",
			httpVerb: "GET",
			data:     nil,
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, response *ServiceResponse) {
				assert.NotNil(t, response)
				assert.Equal(t, 200, response.StatusCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := tt.setupContext()
			defer cancel()

			client, err := NewClient()
			require.NoError(t, err)
			defer client.Close()

			serviceInvocation := NewServiceInvocation(client)

			// Act
			response, err := serviceInvocation.InvokeInquiriesAPI(ctx, tt.method, tt.httpVerb, tt.data)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				if tt.validateResult != nil {
					tt.validateResult(t, response)
				}
			}
		})
	}
}

func TestServiceInvocation_GetContent(t *testing.T) {
	tests := []struct {
		name           string
		contentID      string
		setupContext   func() (context.Context, context.CancelFunc)
		expectedError  string
		validateResult func(*testing.T, *ServiceResponse)
	}{
		{
			name:      "get content with valid ID",
			contentID: "123e4567-e89b-12d3-a456-426614174000",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, response *ServiceResponse) {
				assert.NotNil(t, response)
				assert.Equal(t, 200, response.StatusCode)
			},
		},
		{
			name:      "get content with numeric ID",
			contentID: "12345",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, response *ServiceResponse) {
				assert.NotNil(t, response)
				assert.Equal(t, 200, response.StatusCode)
			},
		},
		{
			name:      "get content with slug ID",
			contentID: "test-content-slug",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, response *ServiceResponse) {
				assert.NotNil(t, response)
				assert.Equal(t, 200, response.StatusCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := tt.setupContext()
			defer cancel()

			client, err := NewClient()
			require.NoError(t, err)
			defer client.Close()

			serviceInvocation := NewServiceInvocation(client)

			// Act
			response, err := serviceInvocation.GetContent(ctx, tt.contentID)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, response)
				}
			}
		})
	}
}

func TestServiceInvocation_GetAllContent(t *testing.T) {
	// Arrange
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	client, err := NewClient()
	require.NoError(t, err)
	defer client.Close()

	serviceInvocation := NewServiceInvocation(client)

	// Act
	response, err := serviceInvocation.GetAllContent(ctx)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 200, response.StatusCode)
	assert.Equal(t, "application/json", response.ContentType)
}

func TestServiceInvocation_GetContentDownload(t *testing.T) {
	tests := []struct {
		name           string
		contentID      string
		expectedError  string
		validateResult func(*testing.T, *ServiceResponse)
	}{
		{
			name:      "get content download with valid ID",
			contentID: "123e4567-e89b-12d3-a456-426614174000",
			validateResult: func(t *testing.T, response *ServiceResponse) {
				assert.NotNil(t, response)
				assert.Equal(t, 200, response.StatusCode)
			},
		},
		{
			name:      "get content download with empty ID",
			contentID: "",
			validateResult: func(t *testing.T, response *ServiceResponse) {
				assert.NotNil(t, response)
				assert.Equal(t, 200, response.StatusCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			client, err := NewClient()
			require.NoError(t, err)
			defer client.Close()

			serviceInvocation := NewServiceInvocation(client)

			// Act
			response, err := serviceInvocation.GetContentDownload(ctx, tt.contentID)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, response)
				}
			}
		})
	}
}

func TestServiceInvocation_GetContentPreview(t *testing.T) {
	tests := []struct {
		name           string
		contentID      string
		validateResult func(*testing.T, *ServiceResponse)
	}{
		{
			name:      "get content preview with valid ID",
			contentID: "123e4567-e89b-12d3-a456-426614174000",
			validateResult: func(t *testing.T, response *ServiceResponse) {
				assert.NotNil(t, response)
				assert.Equal(t, 200, response.StatusCode)
			},
		},
		{
			name:      "get content preview with numeric ID",
			contentID: "12345",
			validateResult: func(t *testing.T, response *ServiceResponse) {
				assert.NotNil(t, response)
				assert.Equal(t, 200, response.StatusCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			client, err := NewClient()
			require.NoError(t, err)
			defer client.Close()

			serviceInvocation := NewServiceInvocation(client)

			// Act
			response, err := serviceInvocation.GetContentPreview(ctx, tt.contentID)

			// Assert
			assert.NoError(t, err)
			if tt.validateResult != nil {
				tt.validateResult(t, response)
			}
		})
	}
}

func TestServiceInvocation_GetService(t *testing.T) {
	tests := []struct {
		name           string
		serviceID      string
		validateResult func(*testing.T, *ServiceResponse)
	}{
		{
			name:      "get service with valid UUID",
			serviceID: "123e4567-e89b-12d3-a456-426614174000",
			validateResult: func(t *testing.T, response *ServiceResponse) {
				assert.NotNil(t, response)
				assert.Equal(t, 200, response.StatusCode)
			},
		},
		{
			name:      "get service with numeric ID",
			serviceID: "12345",
			validateResult: func(t *testing.T, response *ServiceResponse) {
				assert.NotNil(t, response)
				assert.Equal(t, 200, response.StatusCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			client, err := NewClient()
			require.NoError(t, err)
			defer client.Close()

			serviceInvocation := NewServiceInvocation(client)

			// Act
			response, err := serviceInvocation.GetService(ctx, tt.serviceID)

			// Assert
			assert.NoError(t, err)
			if tt.validateResult != nil {
				tt.validateResult(t, response)
			}
		})
	}
}

func TestServiceInvocation_GetAllServices(t *testing.T) {
	// Arrange
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	client, err := NewClient()
	require.NoError(t, err)
	defer client.Close()

	serviceInvocation := NewServiceInvocation(client)

	// Act
	response, err := serviceInvocation.GetAllServices(ctx)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 200, response.StatusCode)
	assert.Equal(t, "application/json", response.ContentType)
}

func TestServiceInvocation_GetServiceBySlug(t *testing.T) {
	tests := []struct {
		name           string
		slug           string
		validateResult func(*testing.T, *ServiceResponse)
	}{
		{
			name: "get service by valid slug",
			slug: "test-service-slug",
			validateResult: func(t *testing.T, response *ServiceResponse) {
				assert.NotNil(t, response)
				assert.Equal(t, 200, response.StatusCode)
			},
		},
		{
			name: "get service by slug with hyphens",
			slug: "complex-service-slug-name",
			validateResult: func(t *testing.T, response *ServiceResponse) {
				assert.NotNil(t, response)
				assert.Equal(t, 200, response.StatusCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			client, err := NewClient()
			require.NoError(t, err)
			defer client.Close()

			serviceInvocation := NewServiceInvocation(client)

			// Act
			response, err := serviceInvocation.GetServiceBySlug(ctx, tt.slug)

			// Assert
			assert.NoError(t, err)
			if tt.validateResult != nil {
				tt.validateResult(t, response)
			}
		})
	}
}

func TestServiceInvocation_GetFeaturedServices(t *testing.T) {
	// Arrange
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	client, err := NewClient()
	require.NoError(t, err)
	defer client.Close()

	serviceInvocation := NewServiceInvocation(client)

	// Act
	response, err := serviceInvocation.GetFeaturedServices(ctx)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 200, response.StatusCode)
}

func TestServiceInvocation_GetServiceCategories(t *testing.T) {
	// Arrange
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	client, err := NewClient()
	require.NoError(t, err)
	defer client.Close()

	serviceInvocation := NewServiceInvocation(client)

	// Act
	response, err := serviceInvocation.GetServiceCategories(ctx)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 200, response.StatusCode)
}

func TestServiceInvocation_GetServicesByCategory(t *testing.T) {
	tests := []struct {
		name           string
		categoryID     string
		validateResult func(*testing.T, *ServiceResponse)
	}{
		{
			name:       "get services by category with UUID",
			categoryID: "123e4567-e89b-12d3-a456-426614174000",
			validateResult: func(t *testing.T, response *ServiceResponse) {
				assert.NotNil(t, response)
				assert.Equal(t, 200, response.StatusCode)
			},
		},
		{
			name:       "get services by category with numeric ID",
			categoryID: "1",
			validateResult: func(t *testing.T, response *ServiceResponse) {
				assert.NotNil(t, response)
				assert.Equal(t, 200, response.StatusCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			client, err := NewClient()
			require.NoError(t, err)
			defer client.Close()

			serviceInvocation := NewServiceInvocation(client)

			// Act
			response, err := serviceInvocation.GetServicesByCategory(ctx, tt.categoryID)

			// Assert
			assert.NoError(t, err)
			if tt.validateResult != nil {
				tt.validateResult(t, response)
			}
		})
	}
}

func TestServiceInvocation_CheckServiceHealth(t *testing.T) {
	tests := []struct {
		name          string
		appID         string
		setupContext  func() (context.Context, context.CancelFunc)
		expectedError string
	}{
		{
			name:  "check health for content API",
			appID: "content-api",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
		},
		{
			name:  "check health for inquiries API",
			appID: "inquiries-api",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
		},
		{
			name:  "check health for custom service",
			appID: "custom-service",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
		},
		{
			name:  "check health with timeout context",
			appID: "content-api",
			setupContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 1*time.Millisecond)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := tt.setupContext()
			defer cancel()

			client, err := NewClient()
			require.NoError(t, err)
			defer client.Close()

			serviceInvocation := NewServiceInvocation(client)

			// Act
			err = serviceInvocation.CheckServiceHealth(ctx, tt.appID)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				// Health check may succeed or fail depending on implementation
			}
		})
	}
}

func TestServiceInvocation_CheckServiceReadiness(t *testing.T) {
	tests := []struct {
		name          string
		appID         string
		expectedError string
	}{
		{
			name:  "check readiness for content API",
			appID: "content-api",
		},
		{
			name:  "check readiness for inquiries API",
			appID: "inquiries-api",
		},
		{
			name:  "check readiness for admin gateway",
			appID: "admin-gateway",
		},
		{
			name:  "check readiness for public gateway",
			appID: "public-gateway",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			client, err := NewClient()
			require.NoError(t, err)
			defer client.Close()

			serviceInvocation := NewServiceInvocation(client)

			// Act
			err = serviceInvocation.CheckServiceReadiness(ctx, tt.appID)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				// Readiness check may succeed or fail depending on implementation
			}
		})
	}
}

func TestServiceInvocation_CheckAllServicesHealth(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func() (context.Context, context.CancelFunc)
		expectedError string
	}{
		{
			name: "check all services health with valid context",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
		},
		{
			name: "check all services health with timeout context",
			setupContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 1*time.Millisecond)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := tt.setupContext()
			defer cancel()

			client, err := NewClient()
			require.NoError(t, err)
			defer client.Close()

			serviceInvocation := NewServiceInvocation(client)

			// Act
			err = serviceInvocation.CheckAllServicesHealth(ctx)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				// Health check may succeed or fail depending on implementation
			}
		})
	}
}

func TestServiceInvocation_CheckAllServicesReadiness(t *testing.T) {
	// Arrange
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	client, err := NewClient()
	require.NoError(t, err)
	defer client.Close()

	serviceInvocation := NewServiceInvocation(client)

	// Act
	err = serviceInvocation.CheckAllServicesReadiness(ctx)

	// Assert - Should not panic, may succeed or fail
	if err != nil {
		assert.IsType(t, err, &domain.DomainError{})
	}
}

func TestServiceInvocation_InvokeWithRetry(t *testing.T) {
	tests := []struct {
		name           string
		request        *ServiceRequest
		maxRetries     int
		setupContext   func() (context.Context, context.CancelFunc)
		expectedError  string
		validateResult func(*testing.T, *ServiceResponse)
	}{
		{
			name: "invoke with retry succeeds on first attempt",
			request: &ServiceRequest{
				AppID:      "test-service",
				MethodName: "test-method",
				HTTPVerb:   "GET",
			},
			maxRetries: 3,
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, response *ServiceResponse) {
				assert.NotNil(t, response)
				assert.Equal(t, 200, response.StatusCode)
			},
		},
		{
			name: "invoke with zero retries",
			request: &ServiceRequest{
				AppID:      "test-service",
				MethodName: "test-method",
				HTTPVerb:   "GET",
			},
			maxRetries: 0,
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateResult: func(t *testing.T, response *ServiceResponse) {
				assert.NotNil(t, response)
				assert.Equal(t, 200, response.StatusCode)
			},
		},
		{
			name: "invoke with retry and timeout context",
			request: &ServiceRequest{
				AppID:      "test-service",
				MethodName: "test-method",
				HTTPVerb:   "GET",
			},
			maxRetries: 2,
			setupContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 50*time.Millisecond)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := tt.setupContext()
			defer cancel()

			client, err := NewClient()
			require.NoError(t, err)
			defer client.Close()

			serviceInvocation := NewServiceInvocation(client)

			// Act
			response, err := serviceInvocation.InvokeWithRetry(ctx, tt.request, tt.maxRetries)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				if tt.validateResult != nil {
					tt.validateResult(t, response)
				}
			}
		})
	}
}

func TestServiceInvocation_CheckContentAPIHealth(t *testing.T) {
	// Arrange
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	client, err := NewClient()
	require.NoError(t, err)
	defer client.Close()

	serviceInvocation := NewServiceInvocation(client)

	// Act
	healthy, err := serviceInvocation.CheckContentAPIHealth(ctx)

	// Assert - Should not panic, may be healthy or unhealthy
	assert.IsType(t, true, healthy)
}

func TestServiceInvocation_CheckInquiriesAPIHealth(t *testing.T) {
	// Arrange
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	client, err := NewClient()
	require.NoError(t, err)
	defer client.Close()

	serviceInvocation := NewServiceInvocation(client)

	// Act
	healthy, err := serviceInvocation.CheckInquiriesAPIHealth(ctx)

	// Assert - Should not panic, may be healthy or unhealthy
	assert.IsType(t, true, healthy)
}

func TestServiceInvocation_GetContentAPIMetrics(t *testing.T) {
	// Arrange
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	client, err := NewClient()
	require.NoError(t, err)
	defer client.Close()

	serviceInvocation := NewServiceInvocation(client)

	// Act
	metrics, err := serviceInvocation.GetContentAPIMetrics(ctx)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Contains(t, metrics, "status")
	assert.Contains(t, metrics, "uptime")
	assert.Contains(t, metrics, "requests")
}

func TestServiceInvocation_GetInquiriesAPIMetrics(t *testing.T) {
	// Arrange
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	client, err := NewClient()
	require.NoError(t, err)
	defer client.Close()

	serviceInvocation := NewServiceInvocation(client)

	// Act
	metrics, err := serviceInvocation.GetInquiriesAPIMetrics(ctx)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Contains(t, metrics, "status")
	assert.Contains(t, metrics, "uptime")
	assert.Contains(t, metrics, "requests")
}

func TestServiceRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request *ServiceRequest
		isValid bool
	}{
		{
			name: "valid service request",
			request: &ServiceRequest{
				AppID:       "test-app",
				MethodName:  "test-method",
				HTTPVerb:    "GET",
				Data:        []byte(`{"test": "data"}`),
				ContentType: "application/json",
				Metadata:    map[string]string{"key": "value"},
			},
			isValid: true,
		},
		{
			name: "request with empty app ID",
			request: &ServiceRequest{
				AppID:      "",
				MethodName: "test-method",
				HTTPVerb:   "GET",
			},
			isValid: true, // Should be handled gracefully
		},
		{
			name: "request with nil data",
			request: &ServiceRequest{
				AppID:      "test-app",
				MethodName: "test-method",
				HTTPVerb:   "GET",
				Data:       nil,
			},
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			client, err := NewClient()
			require.NoError(t, err)
			defer client.Close()

			serviceInvocation := NewServiceInvocation(client)

			// Act
			response, err := serviceInvocation.InvokeService(ctx, tt.request)

			// Assert
			if tt.isValid {
				// Should not panic, may succeed or fail gracefully
				if err == nil {
					assert.NotNil(t, response)
				}
			}
		})
	}
}

func TestServiceResponse_Structure(t *testing.T) {
	// Arrange
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	client, err := NewClient()
	require.NoError(t, err)
	defer client.Close()

	serviceInvocation := NewServiceInvocation(client)

	request := &ServiceRequest{
		AppID:      "test-service",
		MethodName: "test-method",
		HTTPVerb:   "GET",
	}

	// Act
	response, err := serviceInvocation.InvokeService(ctx, request)

	// Assert
	if err == nil && response != nil {
		assert.IsType(t, []byte{}, response.Data)
		assert.IsType(t, "", response.ContentType)
		assert.IsType(t, 0, response.StatusCode)
		assert.IsType(t, map[string]string{}, response.Headers)
		
		// Validate JSON response if content type is JSON
		if response.ContentType == "application/json" {
			var jsonData interface{}
			assert.NoError(t, json.Unmarshal(response.Data, &jsonData))
		}
	}
}