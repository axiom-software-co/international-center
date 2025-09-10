package inquiries

import (
	"testing"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInquiriesHandler(t *testing.T) {
	tests := []struct {
		name        string
		client      *dapr.Client
		expectError bool
		errorMsg    string
	}{
		{
			name:        "successful creation with valid dapr client",
			client:      &dapr.Client{},
			expectError: false,
		},
		{
			name:        "failure with nil dapr client",
			client:      nil,
			expectError: true,
			errorMsg:    "dapr client cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := NewInquiriesHandler(tt.client)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, handler)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, handler)
				assert.NotNil(t, handler.businessHandler)
				assert.NotNil(t, handler.donationsHandler)
				assert.NotNil(t, handler.mediaHandler)
				assert.NotNil(t, handler.volunteersHandler)
			}
		})
	}
}

func TestInquiriesHandler_RegisterRoutes(t *testing.T) {
	client := &dapr.Client{}
	handler, err := NewInquiriesHandler(client)
	require.NoError(t, err)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Test that all inquiries domain routes are registered
	var routes []string
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		routes = append(routes, path)
		return nil
	})

	// Business routes should be registered
	expectedRoutes := []string{
		// Business domain routes
		"/admin/api/v1/business/inquiries",
		"/admin/api/v1/business/inquiries/{id}",
		"/admin/api/v1/business/inquiries/{id}/acknowledge",
		"/admin/api/v1/business/inquiries/{id}/resolve",
		"/admin/api/v1/business/inquiries/{id}/close",
		"/admin/api/v1/business/inquiries/{id}/priority",

		// Donations domain routes
		"/admin/api/v1/donations/inquiries",
		"/admin/api/v1/donations/inquiries/{id}",
		"/admin/api/v1/donations/inquiries/{id}/acknowledge",
		"/admin/api/v1/donations/inquiries/{id}/resolve",
		"/admin/api/v1/donations/inquiries/{id}/close",
		"/admin/api/v1/donations/inquiries/{id}/priority",

		// Media domain routes
		// Media inquiry routes replaced by contract-compliant inquiry endpoints
		"/admin/api/v1/inquiries",
		"/admin/api/v1/inquiries/{id}",

		// Volunteers domain routes
		"/api/v1/volunteers/applications",
		"/admin/api/v1/volunteers/applications",
		"/admin/api/v1/volunteers/applications/{id}",
		"/admin/api/v1/volunteers/applications/search",
		"/admin/api/v1/volunteers/applications/status/{status}",
		"/admin/api/v1/volunteers/applications/interest/{interest}",
		"/admin/api/v1/volunteers/applications/{id}/status",
		"/admin/api/v1/volunteers/applications/{id}/priority",
		"/admin/api/v1/volunteers/applications/{id}/audit",
	}

	for _, expectedRoute := range expectedRoutes {
		assert.Contains(t, routes, expectedRoute, "Expected route %s to be registered", expectedRoute)
	}

	// Verify minimum number of routes registered
	assert.GreaterOrEqual(t, len(routes), len(expectedRoutes), "Expected at least %d routes to be registered", len(expectedRoutes))
}

func TestInquiriesHandler_HealthCheck(t *testing.T) {
	client := &dapr.Client{}
	handler, err := NewInquiriesHandler(client)
	require.NoError(t, err)

	// Test that HealthCheck method exists and is callable
	assert.NotNil(t, handler.HealthCheck)
}

func TestInquiriesHandler_ReadinessCheck(t *testing.T) {
	client := &dapr.Client{}
	handler, err := NewInquiriesHandler(client)
	require.NoError(t, err)

	// Test that ReadinessCheck method exists and is callable
	assert.NotNil(t, handler.ReadinessCheck)
}

func TestInquiriesHandler_DomainHandlerIntegration(t *testing.T) {
	client := &dapr.Client{}
	handler, err := NewInquiriesHandler(client)
	require.NoError(t, err)

	// Test that all domain handlers are properly initialized
	t.Run("business handler integration", func(t *testing.T) {
		assert.NotNil(t, handler.businessHandler)
		// Verify business handler has proper methods
		assert.NotNil(t, handler.businessHandler.RegisterRoutes)
		assert.NotNil(t, handler.businessHandler.HealthCheck)
		assert.NotNil(t, handler.businessHandler.ReadinessCheck)
	})

	t.Run("donations handler integration", func(t *testing.T) {
		assert.NotNil(t, handler.donationsHandler)
		// Verify donations handler has proper methods
		assert.NotNil(t, handler.donationsHandler.RegisterRoutes)
		assert.NotNil(t, handler.donationsHandler.HealthCheck)
		assert.NotNil(t, handler.donationsHandler.ReadinessCheck)
	})

	t.Run("media handler integration", func(t *testing.T) {
		assert.NotNil(t, handler.mediaHandler)
		// Verify media handler has proper methods
		assert.NotNil(t, handler.mediaHandler.RegisterRoutes)
		assert.NotNil(t, handler.mediaHandler.HealthCheck)
		assert.NotNil(t, handler.mediaHandler.ReadinessCheck)
	})

	t.Run("volunteers handler integration", func(t *testing.T) {
		assert.NotNil(t, handler.volunteersHandler)
		// Verify volunteers handler has proper methods
		assert.NotNil(t, handler.volunteersHandler.RegisterRoutes)
		assert.NotNil(t, handler.volunteersHandler.HealthCheck)
		assert.NotNil(t, handler.volunteersHandler.ReadinessCheck)
	})
}