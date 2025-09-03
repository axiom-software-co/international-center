package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED PHASE - Services Domain Entity Tests

func TestServiceEntityCreation(t *testing.T) {

	t.Run("create service with required fields", func(t *testing.T) {
		// Test: Service creation with minimum required fields
		service, err := NewService(
			"Test Service",
			"Service description",
			"test-service",
			"mobile_service",
		)

		require.NoError(t, err, "Should create service with valid required fields")
		assert.Equal(t, "Test Service", service.Title)
		assert.Equal(t, "Service description", service.Description)
		assert.Equal(t, "test-service", service.Slug)
		assert.Equal(t, DeliveryMobilService, service.DeliveryMode)
		assert.Equal(t, PublishingStatusDraft, service.PublishingStatus)
		assert.False(t, service.IsDeleted)
		assert.NotZero(t, service.CreatedOn)
	})

	t.Run("validate required fields", func(t *testing.T) {
		// Test: Service creation fails with missing required fields
		_, err := NewService("", "description", "slug", "mobile_service")
		assert.Error(t, err, "Should fail when title is empty")

		_, err = NewService("Title", "", "slug", "mobile_service")
		assert.Error(t, err, "Should fail when description is empty")

		_, err = NewService("Title", "description", "", "mobile_service")
		assert.Error(t, err, "Should fail when slug is empty")

		_, err = NewService("Title", "description", "slug", "invalid_mode")
		assert.Error(t, err, "Should fail when delivery mode is invalid")
	})

	t.Run("validate slug format", func(t *testing.T) {
		// Test: Slug validation rules
		validSlugs := []string{"test-service", "service-123", "my-health-service"}
		for _, slug := range validSlugs {
			_, err := NewService("Title", "Description", slug, "mobile_service")
			assert.NoError(t, err, "Should accept valid slug: %s", slug)
		}

		invalidSlugs := []string{"Test Service", "service_123", "service!", ""}
		for _, slug := range invalidSlugs {
			_, err := NewService("Title", "Description", slug, "mobile_service")
			assert.Error(t, err, "Should reject invalid slug: %s", slug)
		}
	})
}

func TestServicePublishingWorkflow(t *testing.T) {

	t.Run("draft to published transition", func(t *testing.T) {
		// Test: Publishing workflow state machine
		service, _ := NewService("Test Service", "Description", "test-service", "mobile_service")
		
		err := service.Publish("admin-user")
		require.NoError(t, err, "Should allow publishing from draft status")
		assert.Equal(t, PublishingStatusPublished, service.PublishingStatus)
		assert.Equal(t, "admin-user", service.ModifiedBy)
		assert.NotZero(t, service.ModifiedOn)
	})

	t.Run("published to archived transition", func(t *testing.T) {
		// Test: Archiving workflow
		service, _ := NewService("Test Service", "Description", "test-service", "mobile_service")
		service.Publish("admin-user")
		
		err := service.Archive("admin-user")
		require.NoError(t, err, "Should allow archiving from published status")
		assert.Equal(t, PublishingStatusArchived, service.PublishingStatus)
	})

	t.Run("invalid state transitions", func(t *testing.T) {
		// Test: Invalid state transition prevention
		service, _ := NewService("Test Service", "Description", "test-service", "mobile_service")
		
		err := service.Archive("admin-user")
		assert.Error(t, err, "Should not allow archiving from draft status")
	})
}

func TestServiceCategoryAssignment(t *testing.T) {

	t.Run("assign valid category", func(t *testing.T) {
		// Test: Category assignment
		service, _ := NewService("Test Service", "Description", "test-service", "mobile_service")
		categoryID := "550e8400-e29b-41d4-a716-446655440000"
		
		err := service.AssignCategory(categoryID, "admin-user")
		require.NoError(t, err, "Should assign valid category")
		assert.Equal(t, categoryID, service.CategoryID)
	})

	t.Run("validate category ID format", func(t *testing.T) {
		// Test: Category ID validation
		service, _ := NewService("Test Service", "Description", "test-service", "mobile_service")
		
		err := service.AssignCategory("invalid-uuid", "admin-user")
		assert.Error(t, err, "Should reject invalid UUID format")
	})
}