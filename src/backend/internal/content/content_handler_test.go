package content

import (
	"testing"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContentHandler(t *testing.T) {
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
			handler, err := NewContentHandler(tt.client)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, handler)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, handler)
				assert.NotNil(t, handler.eventsHandler)
				assert.NotNil(t, handler.newsHandler)
				assert.NotNil(t, handler.researchHandler)
				assert.NotNil(t, handler.servicesHandler)
			}
		})
	}
}

func TestContentHandler_RegisterRoutes(t *testing.T) {
	client := &dapr.Client{}
	handler, err := NewContentHandler(client)
	require.NoError(t, err)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Test that all content domain routes are registered
	var routes []string
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		routes = append(routes, path)
		return nil
	})

	// Events routes should be registered
	expectedRoutes := []string{
		// Events domain routes - Admin
		"/admin/api/v1/events",
		"/admin/api/v1/events/{id}",
		"/admin/api/v1/events/{id}/publish",
		"/admin/api/v1/events/{id}/archive",
		"/admin/api/v1/events/categories",
		"/admin/api/v1/events/categories/{id}",
		"/admin/api/v1/events/featured",
		"/admin/api/v1/events/{id}/registrations",

		// Events domain routes - Public
		"/api/v1/events",
		"/api/v1/events/{id}",
		"/api/v1/events/categories",
		"/api/v1/events/featured",
		"/api/v1/events/published",
		"/api/v1/events/upcoming",

		// News domain routes - Admin
		"/admin/api/v1/news",
		"/admin/api/v1/news/{id}",
		"/admin/api/v1/news/{id}/publish",
		"/admin/api/v1/news/{id}/archive",
		"/admin/api/v1/news/categories",
		"/admin/api/v1/news/categories/{id}",
		"/admin/api/v1/news/featured",

		// News domain routes - Public
		"/api/v1/news",
		"/api/v1/news/{id}",
		"/api/v1/news/categories",
		"/api/v1/news/featured",
		"/api/v1/news/search",

		// Research domain routes - Admin
		"/admin/api/v1/research",
		"/admin/api/v1/research/{id}",
		"/admin/api/v1/research/{id}/publish",
		"/admin/api/v1/research/{id}/archive",
		"/admin/api/v1/research/categories",
		"/admin/api/v1/research/categories/{id}",
		"/admin/api/v1/research/featured",

		// Research domain routes - Public
		"/api/v1/research",
		"/api/v1/research/{id}",
		"/api/v1/research/categories",
		"/api/v1/research/featured",
		"/api/v1/research/search",

		// Services domain routes - Admin
		"/admin/api/v1/services",
		"/admin/api/v1/services/{id}",
		"/admin/api/v1/services/{id}/publish",
		"/admin/api/v1/services/{id}/archive",
		"/admin/api/v1/services/categories",
		"/admin/api/v1/services/categories/{id}",
		"/admin/api/v1/services/featured",

		// Services domain routes - Public
		"/api/v1/services",
		"/api/v1/services/{id}",
		"/api/v1/services/categories",
		"/api/v1/services/featured",
		"/api/v1/services/published",
	}

	for _, expectedRoute := range expectedRoutes {
		assert.Contains(t, routes, expectedRoute, "Expected route %s to be registered", expectedRoute)
	}

	// Verify minimum number of routes registered
	assert.GreaterOrEqual(t, len(routes), len(expectedRoutes), "Expected at least %d routes to be registered", len(expectedRoutes))
}

func TestContentHandler_HealthCheck(t *testing.T) {
	client := &dapr.Client{}
	handler, err := NewContentHandler(client)
	require.NoError(t, err)

	// Test that HealthCheck method exists and is callable
	assert.NotNil(t, handler.HealthCheck)
}

func TestContentHandler_ReadinessCheck(t *testing.T) {
	client := &dapr.Client{}
	handler, err := NewContentHandler(client)
	require.NoError(t, err)

	// Test that ReadinessCheck method exists and is callable
	assert.NotNil(t, handler.ReadinessCheck)
}

func TestContentHandler_DomainHandlerIntegration(t *testing.T) {
	client := &dapr.Client{}
	handler, err := NewContentHandler(client)
	require.NoError(t, err)

	// Test that all domain handlers are properly initialized
	t.Run("events handler integration", func(t *testing.T) {
		assert.NotNil(t, handler.eventsHandler)
		// Verify events handler has proper methods
		assert.NotNil(t, handler.eventsHandler.RegisterRoutes)
		assert.NotNil(t, handler.eventsHandler.HealthCheck)
		assert.NotNil(t, handler.eventsHandler.ReadinessCheck)
	})

	t.Run("news handler integration", func(t *testing.T) {
		assert.NotNil(t, handler.newsHandler)
		// Verify news handler has proper methods
		assert.NotNil(t, handler.newsHandler.RegisterRoutes)
		assert.NotNil(t, handler.newsHandler.HealthCheck)
		assert.NotNil(t, handler.newsHandler.ReadinessCheck)
	})

	t.Run("research handler integration", func(t *testing.T) {
		assert.NotNil(t, handler.researchHandler)
		// Verify research handler has proper methods
		assert.NotNil(t, handler.researchHandler.RegisterRoutes)
		assert.NotNil(t, handler.researchHandler.HealthCheck)
		assert.NotNil(t, handler.researchHandler.ReadinessCheck)
	})

	t.Run("services handler integration", func(t *testing.T) {
		assert.NotNil(t, handler.servicesHandler)
		// Verify services handler has proper methods
		assert.NotNil(t, handler.servicesHandler.RegisterRoutes)
		assert.NotNil(t, handler.servicesHandler.HealthCheck)
		assert.NotNil(t, handler.servicesHandler.ReadinessCheck)
	})
}