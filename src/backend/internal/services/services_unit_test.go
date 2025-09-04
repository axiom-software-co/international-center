package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	sharedtesting "github.com/axiom-software-co/international-center/src/backend/internal/shared/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockServicesRepository provides mock implementation for unit tests
type MockServicesRepository struct {
	services         map[string]*Service
	categories       map[string]*ServiceCategory
	featuredCategories map[string]*FeaturedCategory
	auditEvents      []MockAuditEvent
	failures         map[string]error
	blobs            map[string][]byte
}

type MockAuditEvent struct {
	EntityType    domain.EntityType
	EntityID      string
	OperationType domain.AuditEventType
	UserID        string
	Before        interface{}
	After         interface{}
}

func NewMockServicesRepository() *MockServicesRepository {
	return &MockServicesRepository{
		services:         make(map[string]*Service),
		categories:       make(map[string]*ServiceCategory),
		featuredCategories: make(map[string]*FeaturedCategory),
		auditEvents:      make([]MockAuditEvent, 0),
		failures:         make(map[string]error),
		blobs:            make(map[string][]byte),
	}
}

// SetFailure sets a mock failure for specific operations
func (m *MockServicesRepository) SetFailure(operation string, err error) {
	m.failures[operation] = err
}

// GetAuditEvents returns all mock audit events
func (m *MockServicesRepository) GetAuditEvents() []MockAuditEvent {
	return m.auditEvents
}

// Service repository methods
func (m *MockServicesRepository) SaveService(ctx context.Context, service *Service) error {
	if err, exists := m.failures["SaveService"]; exists {
		return err
	}
	m.services[service.ServiceID] = service
	return nil
}

func (m *MockServicesRepository) GetService(ctx context.Context, serviceID string) (*Service, error) {
	if err, exists := m.failures["GetService"]; exists {
		return nil, err
	}
	service, exists := m.services[serviceID]
	if !exists || service.IsDeleted {
		return nil, domain.NewNotFoundError("service", serviceID)
	}
	return service, nil
}

func (m *MockServicesRepository) GetServiceBySlug(ctx context.Context, slug string) (*Service, error) {
	if err, exists := m.failures["GetServiceBySlug"]; exists {
		return nil, err
	}
	for _, service := range m.services {
		if service.Slug == slug && !service.IsDeleted {
			return service, nil
		}
	}
	return nil, domain.NewNotFoundError("service", slug)
}

func (m *MockServicesRepository) GetAllServices(ctx context.Context) ([]*Service, error) {
	if err, exists := m.failures["GetAllServices"]; exists {
		return nil, err
	}
	var services []*Service
	for _, service := range m.services {
		if !service.IsDeleted {
			services = append(services, service)
		}
	}
	return services, nil
}

func (m *MockServicesRepository) GetServicesByCategory(ctx context.Context, categoryID string) ([]*Service, error) {
	if err, exists := m.failures["GetServicesByCategory"]; exists {
		return nil, err
	}
	var services []*Service
	for _, service := range m.services {
		if service.CategoryID == categoryID && !service.IsDeleted {
			services = append(services, service)
		}
	}
	return services, nil
}

func (m *MockServicesRepository) GetServicesByPublishingStatus(ctx context.Context, status PublishingStatus) ([]*Service, error) {
	if err, exists := m.failures["GetServicesByPublishingStatus"]; exists {
		return nil, err
	}
	var services []*Service
	for _, service := range m.services {
		if service.PublishingStatus == status && !service.IsDeleted {
			services = append(services, service)
		}
	}
	return services, nil
}

func (m *MockServicesRepository) SearchServices(ctx context.Context, searchTerm string) ([]*Service, error) {
	if err, exists := m.failures["SearchServices"]; exists {
		return nil, err
	}
	var services []*Service
	for _, service := range m.services {
		if !service.IsDeleted {
			services = append(services, service)
		}
	}
	return services, nil
}

func (m *MockServicesRepository) DeleteService(ctx context.Context, serviceID string, userID string) error {
	if err, exists := m.failures["DeleteService"]; exists {
		return err
	}
	service, exists := m.services[serviceID]
	if !exists {
		return domain.NewNotFoundError("service", serviceID)
	}
	service.Delete(userID)
	return nil
}

// Category repository methods
func (m *MockServicesRepository) GetServiceCategory(ctx context.Context, categoryID string) (*ServiceCategory, error) {
	if err, exists := m.failures["GetServiceCategory"]; exists {
		return nil, err
	}
	category, exists := m.categories[categoryID]
	if !exists || category.IsDeleted {
		return nil, domain.NewNotFoundError("category", categoryID)
	}
	return category, nil
}

func (m *MockServicesRepository) GetServiceCategoryBySlug(ctx context.Context, slug string) (*ServiceCategory, error) {
	if err, exists := m.failures["GetServiceCategoryBySlug"]; exists {
		return nil, err
	}
	for _, category := range m.categories {
		if category.Slug == slug && !category.IsDeleted {
			return category, nil
		}
	}
	return nil, domain.NewNotFoundError("category", slug)
}

func (m *MockServicesRepository) GetAllServiceCategories(ctx context.Context) ([]*ServiceCategory, error) {
	if err, exists := m.failures["GetAllServiceCategories"]; exists {
		return nil, err
	}
	var categories []*ServiceCategory
	for _, category := range m.categories {
		if !category.IsDeleted {
			categories = append(categories, category)
		}
	}
	return categories, nil
}

// Featured category repository methods
func (m *MockServicesRepository) GetAllFeaturedCategories(ctx context.Context) ([]*FeaturedCategory, error) {
	if err, exists := m.failures["GetAllFeaturedCategories"]; exists {
		return nil, err
	}
	var featured []*FeaturedCategory
	for _, fc := range m.featuredCategories {
		featured = append(featured, fc)
	}
	return featured, nil
}

func (m *MockServicesRepository) GetAdminFeaturedCategories(ctx context.Context) ([]*FeaturedCategory, error) {
	if err, exists := m.failures["GetAdminFeaturedCategories"]; exists {
		return nil, err
	}
	// For admin, return same as GetAllFeaturedCategories
	return m.GetAllFeaturedCategories(ctx)
}

func (m *MockServicesRepository) GetFeaturedCategoryByPosition(ctx context.Context, position int) (*FeaturedCategory, error) {
	if err, exists := m.failures["GetFeaturedCategoryByPosition"]; exists {
		return nil, err
	}
	for _, fc := range m.featuredCategories {
		if fc.FeaturePosition == position {
			return fc, nil
		}
	}
	return nil, domain.NewNotFoundError("featured_category", fmt.Sprintf("position-%d", position))
}

// Content repository methods
func (m *MockServicesRepository) CreateServiceContentBlobURL(ctx context.Context, contentURL string, expirationMinutes int) (string, error) {
	if err, exists := m.failures["CreateServiceContentBlobURL"]; exists {
		return "", err
	}
	return "https://mock-blob-storage.com/download/" + contentURL, nil
}

func (m *MockServicesRepository) UploadServiceContentBlob(ctx context.Context, storagePath string, content []byte, contentType string) error {
	if err, exists := m.failures["UploadServiceContentBlob"]; exists {
		return err
	}
	return nil
}

// Audit repository methods
func (m *MockServicesRepository) PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, before interface{}, after interface{}) error {
	if err, exists := m.failures["PublishAuditEvent"]; exists {
		return err
	}
	m.auditEvents = append(m.auditEvents, MockAuditEvent{
		EntityType:    entityType,
		EntityID:      entityID,
		OperationType: operationType,
		UserID:        userID,
		Before:        before,
		After:         after,
	})
	return nil
}

// SaveServiceCategory mocks saving a service category
func (m *MockServicesRepository) SaveServiceCategory(ctx context.Context, category *ServiceCategory) error {
	if err, exists := m.failures["SaveServiceCategory"]; exists {
		return err
	}
	m.categories[category.CategoryID] = category
	return nil
}

// GetDefaultUnassignedCategory mocks getting default unassigned category
func (m *MockServicesRepository) GetDefaultUnassignedCategory(ctx context.Context) (*ServiceCategory, error) {
	if err, exists := m.failures["GetDefaultUnassignedCategory"]; exists {
		return nil, err
	}
	for _, category := range m.categories {
		if category.IsDefaultUnassigned {
			return category, nil
		}
	}
	// Return a default category if none exists
	defaultCat, _ := NewServiceCategory("Unassigned", "unassigned", true, "system")
	return defaultCat, nil
}

// DeleteServiceCategory mocks deleting a service category
func (m *MockServicesRepository) DeleteServiceCategory(ctx context.Context, categoryID string, userID string) error {
	if err, exists := m.failures["DeleteServiceCategory"]; exists {
		return err
	}
	category, exists := m.categories[categoryID]
	if !exists {
		return domain.NewNotFoundError("service_category", categoryID)
	}
	category.Delete(userID)
	return nil
}

// SaveFeaturedCategory mocks saving a featured category
func (m *MockServicesRepository) SaveFeaturedCategory(ctx context.Context, featured *FeaturedCategory) error {
	if err, exists := m.failures["SaveFeaturedCategory"]; exists {
		return err
	}
	m.featuredCategories[featured.FeaturedCategoryID] = featured
	return nil
}

// GetFeaturedCategory mocks getting a featured category by ID
func (m *MockServicesRepository) GetFeaturedCategory(ctx context.Context, featuredCategoryID string) (*FeaturedCategory, error) {
	if err, exists := m.failures["GetFeaturedCategory"]; exists {
		return nil, err
	}
	featured, exists := m.featuredCategories[featuredCategoryID]
	if !exists {
		return nil, domain.NewNotFoundError("featured_category", featuredCategoryID)
	}
	return featured, nil
}

// DeleteFeaturedCategory mocks deleting a featured category
func (m *MockServicesRepository) DeleteFeaturedCategory(ctx context.Context, featuredCategoryID string) error {
	if err, exists := m.failures["DeleteFeaturedCategory"]; exists {
		return err
	}
	_, exists := m.featuredCategories[featuredCategoryID]
	if !exists {
		return domain.NewNotFoundError("featured_category", featuredCategoryID)
	}
	delete(m.featuredCategories, featuredCategoryID)
	return nil
}

// DownloadServiceContentBlob mocks downloading a blob
func (m *MockServicesRepository) DownloadServiceContentBlob(ctx context.Context, storagePath string) ([]byte, error) {
	if err, exists := m.failures["DownloadServiceContentBlob"]; exists {
		return nil, err
	}
	if blob, exists := m.blobs[storagePath]; exists {
		return blob, nil
	}
	return nil, domain.NewNotFoundError("blob", storagePath)
}

// Admin audit repository methods - these will be implemented in the Green phase

// GetServiceAudit mocks getting audit trail for a service
func (m *MockServicesRepository) GetServiceAudit(ctx context.Context, serviceID string, limit int, offset int) ([]*ServiceAuditEvent, error) {
	if err, exists := m.failures["GetServiceAudit"]; exists {
		return nil, err
	}
	// Return mock audit events for this service
	events := []*ServiceAuditEvent{}
	for _, auditEvent := range m.auditEvents {
		if auditEvent.EntityID == serviceID && auditEvent.EntityType == domain.EntityTypeService {
			events = append(events, &ServiceAuditEvent{
				AuditID:       fmt.Sprintf("audit-%s-%d", serviceID, len(events)+1),
				EntityType:    string(auditEvent.EntityType),
				EntityID:      auditEvent.EntityID,
				OperationType: string(auditEvent.OperationType),
				AuditTimestamp: time.Now().UTC().Add(-time.Duration(len(events)) * time.Hour),
				UserID:        auditEvent.UserID,
				DataSnapshot: AuditDataSnapshot{
					Before: auditEvent.Before,
					After:  auditEvent.After,
				},
				Environment: "development",
			})
		}
	}
	return events, nil
}

// GetServiceCategoryAudit mocks getting audit trail for a service category
func (m *MockServicesRepository) GetServiceCategoryAudit(ctx context.Context, categoryID string, limit int, offset int) ([]*ServiceAuditEvent, error) {
	if err, exists := m.failures["GetServiceCategoryAudit"]; exists {
		return nil, err
	}
	// Return mock audit events for this category
	var events []*ServiceAuditEvent
	for _, auditEvent := range m.auditEvents {
		if auditEvent.EntityID == categoryID && auditEvent.EntityType == domain.EntityTypeServiceCategory {
			events = append(events, &ServiceAuditEvent{
				AuditID:       fmt.Sprintf("audit-%s-%d", categoryID, len(events)+1),
				EntityType:    string(auditEvent.EntityType),
				EntityID:      auditEvent.EntityID,
				OperationType: string(auditEvent.OperationType),
				AuditTimestamp: time.Now().UTC().Add(-time.Duration(len(events)) * time.Hour),
				UserID:        auditEvent.UserID,
				DataSnapshot: AuditDataSnapshot{
					Before: auditEvent.Before,
					After:  auditEvent.After,
				},
				Environment: "development",
			})
		}
	}
	return events, nil
}

// Test helper functions
func createTestService(userID string) *Service {
	service, _ := NewService("Test Service", "Test Description", "test-service", "category-1", DeliveryModeMobile, userID)
	return service
}

func createTestCategory(userID string) *ServiceCategory {
	category, _ := NewServiceCategory("Test Category", "test-category", false, userID)
	return category
}

// Unit Tests for ServicesService

func TestServicesService_GetService(t *testing.T) {
	tests := []struct {
		name          string
		serviceID     string
		userID        string
		setupMock     func(*MockServicesRepository)
		expectedError string
		validateResult func(*testing.T, *Service)
	}{
		{
			name:      "successfully get published service without auth",
			serviceID: "service-1",
			userID:    "",
			setupMock: func(repo *MockServicesRepository) {
				service := createTestService("creator-1")
				service.ServiceID = "service-1"
				service.PublishingStatus = PublishingStatusPublished
				repo.services["service-1"] = service
			},
			validateResult: func(t *testing.T, service *Service) {
				assert.Equal(t, "service-1", service.ServiceID)
				assert.Equal(t, PublishingStatusPublished, service.PublishingStatus)
			},
		},
		{
			name:      "successfully get draft service with creator auth",
			serviceID: "service-2",
			userID:    "creator-1",
			setupMock: func(repo *MockServicesRepository) {
				service := createTestService("creator-1")
				service.ServiceID = "service-2"
				service.PublishingStatus = PublishingStatusDraft
				repo.services["service-2"] = service
			},
			validateResult: func(t *testing.T, service *Service) {
				assert.Equal(t, "service-2", service.ServiceID)
				assert.Equal(t, PublishingStatusDraft, service.PublishingStatus)
			},
		},
		{
			name:          "fail with empty service ID",
			serviceID:     "",
			userID:        "user-1",
			setupMock:     func(repo *MockServicesRepository) {},
			expectedError: "service ID cannot be empty",
		},
		{
			name:      "fail accessing draft service without auth",
			serviceID: "service-3",
			userID:    "",
			setupMock: func(repo *MockServicesRepository) {
				service := createTestService("creator-1")
				service.ServiceID = "service-3"
				service.PublishingStatus = PublishingStatusDraft
				repo.services["service-3"] = service
			},
			expectedError: "authentication required",
		},
		{
			name:      "fail accessing draft service with wrong user",
			serviceID: "service-4",
			userID:    "different-user",
			setupMock: func(repo *MockServicesRepository) {
				service := createTestService("creator-1")
				service.ServiceID = "service-4"
				service.PublishingStatus = PublishingStatusDraft
				repo.services["service-4"] = service
			},
			expectedError: "insufficient permissions",
		},
		{
			name:      "fail when service not found",
			serviceID: "nonexistent",
			userID:    "user-1",
			setupMock: func(repo *MockServicesRepository) {},
			expectedError: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			
			mockRepo := NewMockServicesRepository()
			tt.setupMock(mockRepo)
			service := NewServicesService(mockRepo)

			// Act
			result, err := service.GetService(ctx, tt.serviceID, tt.userID)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

func TestServicesService_CreateService(t *testing.T) {
	tests := []struct {
		name           string
		title          string
		description    string
		slug           string
		categoryID     string
		deliveryMode   DeliveryMode
		userID         string
		setupMock      func(*MockServicesRepository)
		expectedError  string
		validateResult func(*testing.T, *Service, *MockServicesRepository)
	}{
		{
			name:         "successfully create service",
			title:        "New Service",
			description:  "New Description",
			slug:         "new-service",
			categoryID:   "category-1",
			deliveryMode: DeliveryModeMobile,
			userID:       "creator-1",
			setupMock: func(repo *MockServicesRepository) {
				category := createTestCategory("admin")
				category.CategoryID = "category-1"
				repo.categories["category-1"] = category
			},
			validateResult: func(t *testing.T, service *Service, repo *MockServicesRepository) {
				assert.Equal(t, "New Service", service.Title)
				assert.Equal(t, "New Description", service.Description)
				assert.Equal(t, "new-service", service.Slug)
				assert.Equal(t, "category-1", service.CategoryID)
				assert.Equal(t, DeliveryModeMobile, service.DeliveryMode)
				assert.Equal(t, "creator-1", service.CreatedBy)
				assert.Equal(t, PublishingStatusDraft, service.PublishingStatus)
				
				// Verify service was saved
				savedService, exists := repo.services[service.ServiceID]
				assert.True(t, exists)
				assert.Equal(t, service.ServiceID, savedService.ServiceID)
				
				// Verify audit event was published
				auditEvents := repo.GetAuditEvents()
				assert.Len(t, auditEvents, 1)
				assert.Equal(t, domain.EntityTypeService, auditEvents[0].EntityType)
				assert.Equal(t, service.ServiceID, auditEvents[0].EntityID)
				assert.Equal(t, domain.AuditEventInsert, auditEvents[0].OperationType)
				assert.Equal(t, "creator-1", auditEvents[0].UserID)
			},
		},
		{
			name:          "fail with empty title",
			title:         "",
			description:   "Description",
			slug:          "slug",
			categoryID:    "category-1",
			deliveryMode:  DeliveryModeMobile,
			userID:        "user-1",
			setupMock:     func(repo *MockServicesRepository) {},
			expectedError: "title cannot be empty",
		},
		{
			name:          "fail with empty description",
			title:         "Title",
			description:   "",
			slug:          "slug",
			categoryID:    "category-1",
			deliveryMode:  DeliveryModeMobile,
			userID:        "user-1",
			setupMock:     func(repo *MockServicesRepository) {},
			expectedError: "description cannot be empty",
		},
		{
			name:          "fail with invalid slug",
			title:         "Title",
			description:   "Description",
			slug:          "INVALID SLUG!",
			categoryID:    "category-1",
			deliveryMode:  DeliveryModeMobile,
			userID:        "user-1",
			setupMock:     func(repo *MockServicesRepository) {},
			expectedError: "slug must contain only lowercase letters, numbers, and hyphens",
		},
		{
			name:         "fail when category not found",
			title:        "Title",
			description:  "Description",
			slug:         "slug",
			categoryID:   "nonexistent",
			deliveryMode: DeliveryModeMobile,
			userID:       "user-1",
			setupMock:    func(repo *MockServicesRepository) {},
			expectedError: "not found",
		},
		{
			name:         "fail when slug already exists",
			title:        "Title",
			description:  "Description",
			slug:         "existing-slug",
			categoryID:   "category-1",
			deliveryMode: DeliveryModeMobile,
			userID:       "user-1",
			setupMock: func(repo *MockServicesRepository) {
				category := createTestCategory("admin")
				category.CategoryID = "category-1"
				repo.categories["category-1"] = category
				
				existingService := createTestService("other-user")
				existingService.Slug = "existing-slug"
				repo.services["existing"] = existingService
			},
			expectedError: "already exists",
		},
		{
			name:          "fail with invalid delivery mode",
			title:         "Title",
			description:   "Description",
			slug:          "slug",
			categoryID:    "category-1",
			deliveryMode:  "invalid_mode",
			userID:        "user-1",
			setupMock:     func(repo *MockServicesRepository) {},
			expectedError: "invalid delivery mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			
			mockRepo := NewMockServicesRepository()
			tt.setupMock(mockRepo)
			service := NewServicesService(mockRepo)

			// Act
			result, err := service.CreateService(ctx, tt.title, tt.description, tt.slug, tt.categoryID, tt.deliveryMode, tt.userID)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result, mockRepo)
				}
			}
		})
	}
}

func TestServicesService_PublishService(t *testing.T) {
	tests := []struct {
		name           string
		serviceID      string
		userID         string
		setupMock      func(*MockServicesRepository)
		expectedError  string
		validateResult func(*testing.T, *Service, *MockServicesRepository)
	}{
		{
			name:      "successfully publish draft service",
			serviceID: "service-1",
			userID:    "creator-1",
			setupMock: func(repo *MockServicesRepository) {
				service := createTestService("creator-1")
				service.ServiceID = "service-1"
				service.PublishingStatus = PublishingStatusDraft
				repo.services["service-1"] = service
			},
			validateResult: func(t *testing.T, service *Service, repo *MockServicesRepository) {
				assert.Equal(t, PublishingStatusPublished, service.PublishingStatus)
				assert.Equal(t, "creator-1", service.ModifiedBy)
				assert.NotNil(t, service.ModifiedOn)
				
				// Verify audit event was published
				auditEvents := repo.GetAuditEvents()
				assert.Len(t, auditEvents, 1)
				assert.Equal(t, domain.AuditEventUpdate, auditEvents[0].OperationType)
			},
		},
		{
			name:      "fail when service not found",
			serviceID: "nonexistent",
			userID:    "user-1",
			setupMock: func(repo *MockServicesRepository) {},
			expectedError: "not found",
		},
		{
			name:      "fail when user lacks permission",
			serviceID: "service-2",
			userID:    "different-user",
			setupMock: func(repo *MockServicesRepository) {
				service := createTestService("creator-1")
				service.ServiceID = "service-2"
				service.PublishingStatus = PublishingStatusDraft
				repo.services["service-2"] = service
			},
			expectedError: "insufficient permissions",
		},
		{
			name:      "fail when service already published",
			serviceID: "service-3",
			userID:    "creator-1",
			setupMock: func(repo *MockServicesRepository) {
				service := createTestService("creator-1")
				service.ServiceID = "service-3"
				service.PublishingStatus = PublishingStatusPublished
				repo.services["service-3"] = service
			},
			expectedError: "can only publish services with draft status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			
			mockRepo := NewMockServicesRepository()
			tt.setupMock(mockRepo)
			service := NewServicesService(mockRepo)

			// Act
			result, err := service.PublishService(ctx, tt.serviceID, tt.userID)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result, mockRepo)
				}
			}
		})
	}
}

func TestServicesService_GetAllServices(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		setupMock      func(*MockServicesRepository)
		expectedError  string
		validateResult func(*testing.T, []*Service)
	}{
		{
			name:   "successfully get published services for anonymous user",
			userID: "",
			setupMock: func(repo *MockServicesRepository) {
				// Published service - should be accessible
				publishedService := createTestService("creator-1")
				publishedService.ServiceID = "published"
				publishedService.PublishingStatus = PublishingStatusPublished
				repo.services["published"] = publishedService
				
				// Draft service - should not be accessible anonymously
				draftService := createTestService("creator-1")
				draftService.ServiceID = "draft"
				draftService.PublishingStatus = PublishingStatusDraft
				repo.services["draft"] = draftService
			},
			validateResult: func(t *testing.T, services []*Service) {
				assert.Len(t, services, 1)
				assert.Equal(t, "published", services[0].ServiceID)
				assert.Equal(t, PublishingStatusPublished, services[0].PublishingStatus)
			},
		},
		{
			name:   "successfully get all accessible services for authenticated user",
			userID: "creator-1",
			setupMock: func(repo *MockServicesRepository) {
				// Published service - should be accessible
				publishedService := createTestService("creator-1")
				publishedService.ServiceID = "published"
				publishedService.PublishingStatus = PublishingStatusPublished
				repo.services["published"] = publishedService
				
				// Own draft service - should be accessible
				ownDraftService := createTestService("creator-1")
				ownDraftService.ServiceID = "own-draft"
				ownDraftService.PublishingStatus = PublishingStatusDraft
				repo.services["own-draft"] = ownDraftService
				
				// Other user's draft service - should not be accessible
				otherDraftService := createTestService("other-creator")
				otherDraftService.ServiceID = "other-draft"
				otherDraftService.PublishingStatus = PublishingStatusDraft
				repo.services["other-draft"] = otherDraftService
			},
			validateResult: func(t *testing.T, services []*Service) {
				assert.Len(t, services, 2)
				serviceIDs := make([]string, len(services))
				for i, svc := range services {
					serviceIDs[i] = svc.ServiceID
				}
				assert.Contains(t, serviceIDs, "published")
				assert.Contains(t, serviceIDs, "own-draft")
			},
		},
		{
			name:   "return empty array when no accessible services",
			userID: "",
			setupMock: func(repo *MockServicesRepository) {
				// Only draft service - not accessible anonymously
				draftService := createTestService("creator-1")
				draftService.ServiceID = "draft"
				draftService.PublishingStatus = PublishingStatusDraft
				repo.services["draft"] = draftService
			},
			validateResult: func(t *testing.T, services []*Service) {
				assert.Len(t, services, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			
			mockRepo := NewMockServicesRepository()
			tt.setupMock(mockRepo)
			service := NewServicesService(mockRepo)

			// Act
			result, err := service.GetAllServices(ctx, tt.userID)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

func TestServicesService_Timeout(t *testing.T) {
	// Test that context timeout is respected (5 seconds for unit tests)
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()
	
	// Verify context has 5 second timeout
	deadline, hasDeadline := ctx.Deadline()
	require.True(t, hasDeadline)
	assert.True(t, time.Until(deadline) <= 5*time.Second)
	assert.True(t, time.Until(deadline) > 4*time.Second) // Allow some margin
}

// Admin Audit Endpoint Unit Tests - RED PHASE (Failing Tests)

func TestServicesService_GetServiceAudit(t *testing.T) {
	tests := []struct {
		name           string
		serviceID      string
		userID         string
		limit          int
		offset         int
		setupMock      func(*MockServicesRepository)
		expectedError  string
		validateResult func(*testing.T, []*ServiceAuditEvent)
	}{
		{
			name:      "successfully get service audit trail",
			serviceID: "service-1",
			userID:    "admin-1",
			limit:     10,
			offset:    0,
			setupMock: func(repo *MockServicesRepository) {
				// Create a service with some audit events
				service := createTestService("creator-1")
				service.ServiceID = "service-1"
				repo.services["service-1"] = service
				
				// Add audit events
				repo.auditEvents = []MockAuditEvent{
					{
						EntityType:    domain.EntityTypeService,
						EntityID:      "service-1",
						OperationType: domain.AuditEventInsert,
						UserID:        "creator-1",
						Before:        nil,
						After:         service,
					},
					{
						EntityType:    domain.EntityTypeService,
						EntityID:      "service-1",
						OperationType: domain.AuditEventUpdate,
						UserID:        "creator-1",
						Before:        service,
						After:         service,
					},
				}
			},
			validateResult: func(t *testing.T, events []*ServiceAuditEvent) {
				assert.Len(t, events, 2)
				assert.Equal(t, "service-1", events[0].EntityID)
				assert.Equal(t, "service", events[0].EntityType)
				assert.Equal(t, "creator-1", events[0].UserID)
				assert.Equal(t, "development", events[0].Environment)
			},
		},
		{
			name:          "fail with empty service ID",
			serviceID:     "",
			userID:        "admin-1",
			limit:         10,
			offset:        0,
			setupMock:     func(repo *MockServicesRepository) {},
			expectedError: "service ID cannot be empty",
		},
		{
			name:          "fail without admin authentication",
			serviceID:     "service-1",
			userID:        "",
			limit:         10,
			offset:        0,
			setupMock:     func(repo *MockServicesRepository) {},
			expectedError: "admin authentication required",
		},
		{
			name:      "return empty array for service with no audit events",
			serviceID: "service-2",
			userID:    "admin-1",
			limit:     10,
			offset:    0,
			setupMock: func(repo *MockServicesRepository) {
				service := createTestService("creator-1")
				service.ServiceID = "service-2"
				repo.services["service-2"] = service
			},
			validateResult: func(t *testing.T, events []*ServiceAuditEvent) {
				assert.Len(t, events, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			
			mockRepo := NewMockServicesRepository()
			tt.setupMock(mockRepo)
			service := NewServicesService(mockRepo)

			// Act - this will fail until we implement GetServiceAudit method
			result, err := service.GetServiceAudit(ctx, tt.serviceID, tt.userID, tt.limit, tt.offset)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

func TestServicesService_GetServiceCategoryAudit(t *testing.T) {
	tests := []struct {
		name           string
		categoryID     string
		userID         string
		limit          int
		offset         int
		setupMock      func(*MockServicesRepository)
		expectedError  string
		validateResult func(*testing.T, []*ServiceAuditEvent)
	}{
		{
			name:       "successfully get category audit trail",
			categoryID: "category-1",
			userID:     "admin-1",
			limit:      10,
			offset:     0,
			setupMock: func(repo *MockServicesRepository) {
				category := createTestCategory("creator-1")
				category.CategoryID = "category-1"
				repo.categories["category-1"] = category
				
				// Add audit events
				repo.auditEvents = []MockAuditEvent{
					{
						EntityType:    domain.EntityTypeServiceCategory,
						EntityID:      "category-1",
						OperationType: domain.AuditEventInsert,
						UserID:        "creator-1",
						Before:        nil,
						After:         category,
					},
				}
			},
			validateResult: func(t *testing.T, events []*ServiceAuditEvent) {
				assert.Len(t, events, 1)
				assert.Equal(t, "category-1", events[0].EntityID)
				assert.Equal(t, "service_category", events[0].EntityType)
				assert.Equal(t, "creator-1", events[0].UserID)
			},
		},
		{
			name:          "fail with empty category ID",
			categoryID:    "",
			userID:        "admin-1",
			limit:         10,
			offset:        0,
			setupMock:     func(repo *MockServicesRepository) {},
			expectedError: "category ID cannot be empty",
		},
		{
			name:          "fail without admin authentication",
			categoryID:    "category-1",
			userID:        "",
			limit:         10,
			offset:        0,
			setupMock:     func(repo *MockServicesRepository) {},
			expectedError: "admin authentication required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			
			mockRepo := NewMockServicesRepository()
			tt.setupMock(mockRepo)
			service := NewServicesService(mockRepo)

			// Act - this will fail until we implement GetServiceCategoryAudit method
			result, err := service.GetServiceCategoryAudit(ctx, tt.categoryID, tt.userID, tt.limit, tt.offset)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

func TestServicesService_GetAdminFeaturedCategories(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		setupMock      func(*MockServicesRepository)
		expectedError  string
		validateResult func(*testing.T, []*FeaturedCategory)
	}{
		{
			name:   "successfully get admin view of featured categories",
			userID: "admin-1",
			setupMock: func(repo *MockServicesRepository) {
				category1 := createTestCategory("admin")
				category1.CategoryID = "cat-1"
				repo.categories["cat-1"] = category1
				
				category2 := createTestCategory("admin")
				category2.CategoryID = "cat-2"
				repo.categories["cat-2"] = category2
				
				featured1, _ := NewFeaturedCategory("cat-1", 1, "admin")
				featured2, _ := NewFeaturedCategory("cat-2", 2, "admin")
				
				repo.featuredCategories[featured1.FeaturedCategoryID] = featured1
				repo.featuredCategories[featured2.FeaturedCategoryID] = featured2
			},
			validateResult: func(t *testing.T, featured []*FeaturedCategory) {
				assert.Len(t, featured, 2)
				// Should include detailed admin information
				assert.NotEmpty(t, featured[0].CreatedBy)
				assert.NotNil(t, featured[0].CreatedOn)
			},
		},
		{
			name:          "fail without admin authentication",
			userID:        "",
			setupMock:     func(repo *MockServicesRepository) {},
			expectedError: "admin authentication required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			
			mockRepo := NewMockServicesRepository()
			tt.setupMock(mockRepo)
			service := NewServicesService(mockRepo)

			// Act - this will fail until we implement GetAdminFeaturedCategories method
			result, err := service.GetAdminFeaturedCategories(ctx, tt.userID)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

// Admin CRUD Operations Tests

func TestServicesService_AdminCreateService(t *testing.T) {
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	tests := []struct {
		name      string
		service   *Service
		userID    string
		setupFunc func(*MockServicesRepository)
		wantErr   bool
	}{
		{
			name: "successfully create new service",
			service: &Service{
				Title:            "New Healthcare Service",
				Description:      "Description for new healthcare service with sufficient length to meet validation requirements",
				Slug:             "new-healthcare-service",
				CategoryID:       "550e8400-e29b-41d4-a716-446655440002",
				DeliveryMode:     DeliveryModeOutpatient,
				PublishingStatus: PublishingStatusDraft,
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockServicesRepository) {},
			wantErr:   false,
		},
		{
			name: "return validation error for invalid service",
			service: &Service{
				Title: "", // Invalid: empty title
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockServicesRepository) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockServicesRepository()
			tt.setupFunc(repo)
			service := NewServicesService(repo)

			err := service.AdminCreateService(ctx, tt.service, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, repo.auditEvents, 1)
				assert.Equal(t, domain.AuditEventInsert, repo.auditEvents[0].OperationType)
			}
		})
	}
}

func TestServicesService_AdminUpdateService(t *testing.T) {
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	tests := []struct {
		name      string
		service   *Service
		userID    string
		setupFunc func(*MockServicesRepository)
		wantErr   bool
	}{
		{
			name: "successfully update existing service",
			service: &Service{
				ServiceID:        "550e8400-e29b-41d4-a716-446655440001",
				Title:            "Updated Service Title",
				Description:      "Updated description for service with sufficient length to meet validation requirements",
				Slug:             "updated-service-title",
				CategoryID:       "550e8400-e29b-41d4-a716-446655440002",
				DeliveryMode:     DeliveryModeOutpatient,
				PublishingStatus: PublishingStatusDraft,
			},
			userID: "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockServicesRepository) {
				service := createTestService("admin-550e8400-e29b-41d4-a716-446655440003")
				service.ServiceID = "550e8400-e29b-41d4-a716-446655440001"
				repo.services["550e8400-e29b-41d4-a716-446655440001"] = service
			},
			wantErr: false,
		},
		{
			name: "return not found error for non-existent service",
			service: &Service{
				ServiceID: "550e8400-e29b-41d4-a716-446655440999",
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockServicesRepository) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockServicesRepository()
			tt.setupFunc(repo)
			service := NewServicesService(repo)

			err := service.AdminUpdateService(ctx, tt.service, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, repo.auditEvents, 1)
				assert.Equal(t, domain.AuditEventUpdate, repo.auditEvents[0].OperationType)
			}
		})
	}
}

func TestServicesService_AdminDeleteService(t *testing.T) {
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	tests := []struct {
		name      string
		serviceID string
		userID    string
		setupFunc func(*MockServicesRepository)
		wantErr   bool
	}{
		{
			name:      "successfully delete existing service",
			serviceID: "550e8400-e29b-41d4-a716-446655440001",
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockServicesRepository) {
				service := createTestService("admin-550e8400-e29b-41d4-a716-446655440003")
				service.ServiceID = "550e8400-e29b-41d4-a716-446655440001"
				repo.services["550e8400-e29b-41d4-a716-446655440001"] = service
			},
			wantErr: false,
		},
		{
			name:      "return not found error for non-existent service",
			serviceID: "550e8400-e29b-41d4-a716-446655440999",
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockServicesRepository) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockServicesRepository()
			tt.setupFunc(repo)
			service := NewServicesService(repo)

			err := service.AdminDeleteService(ctx, tt.serviceID, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, repo.auditEvents, 1)
				assert.Equal(t, domain.AuditEventDelete, repo.auditEvents[0].OperationType)
			}
		})
	}
}

func TestServicesService_AdminPublishService(t *testing.T) {
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	tests := []struct {
		name      string
		serviceID string
		userID    string
		setupFunc func(*MockServicesRepository)
		wantErr   bool
	}{
		{
			name:      "successfully publish draft service",
			serviceID: "550e8400-e29b-41d4-a716-446655440001",
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockServicesRepository) {
				service := createTestService("admin-550e8400-e29b-41d4-a716-446655440003")
				service.ServiceID = "550e8400-e29b-41d4-a716-446655440001"
				service.PublishingStatus = PublishingStatusDraft
				repo.services["550e8400-e29b-41d4-a716-446655440001"] = service
			},
			wantErr: false,
		},
		{
			name:      "return validation error for service without required fields",
			serviceID: "550e8400-e29b-41d4-a716-446655440001",
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockServicesRepository) {
				service := createTestService("admin-550e8400-e29b-41d4-a716-446655440003")
				service.ServiceID = "550e8400-e29b-41d4-a716-446655440001"
				service.Description = "" // Missing description for publication
				service.PublishingStatus = PublishingStatusDraft
				repo.services["550e8400-e29b-41d4-a716-446655440001"] = service
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockServicesRepository()
			tt.setupFunc(repo)
			service := NewServicesService(repo)

			err := service.AdminPublishService(ctx, tt.serviceID, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, repo.auditEvents, 1)
				assert.Equal(t, domain.AuditEventPublish, repo.auditEvents[0].OperationType)
			}
		})
	}
}

func TestServicesService_AdminArchiveService(t *testing.T) {
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	tests := []struct {
		name      string
		serviceID string
		userID    string
		setupFunc func(*MockServicesRepository)
		wantErr   bool
	}{
		{
			name:      "successfully archive published service",
			serviceID: "550e8400-e29b-41d4-a716-446655440001",
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockServicesRepository) {
				service := createTestService("admin-550e8400-e29b-41d4-a716-446655440003")
				service.ServiceID = "550e8400-e29b-41d4-a716-446655440001"
				service.PublishingStatus = PublishingStatusPublished
				repo.services["550e8400-e29b-41d4-a716-446655440001"] = service
			},
			wantErr: false,
		},
		{
			name:      "return not found error for non-existent service",
			serviceID: "550e8400-e29b-41d4-a716-446655440999",
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockServicesRepository) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockServicesRepository()
			tt.setupFunc(repo)
			service := NewServicesService(repo)

			err := service.AdminArchiveService(ctx, tt.serviceID, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, repo.auditEvents, 1)
				assert.Equal(t, domain.AuditEventArchive, repo.auditEvents[0].OperationType)
			}
		})
	}
}

func TestServicesService_AdminCreateServiceCategory(t *testing.T) {
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	tests := []struct {
		name      string
		category  *ServiceCategory
		userID    string
		setupFunc func(*MockServicesRepository)
		wantErr   bool
	}{
		{
			name: "successfully create new service category",
			category: &ServiceCategory{
				Name: "New Healthcare Category",
				Slug: "new-healthcare-category",
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockServicesRepository) {},
			wantErr:   false,
		},
		{
			name: "return validation error for invalid category",
			category: &ServiceCategory{
				Name: "", // Invalid: empty name
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockServicesRepository) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockServicesRepository()
			tt.setupFunc(repo)
			service := NewServicesService(repo)

			err := service.AdminCreateServiceCategory(ctx, tt.category, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, repo.auditEvents, 1)
				assert.Equal(t, domain.AuditEventInsert, repo.auditEvents[0].OperationType)
			}
		})
	}
}

func TestServicesService_AdminUpdateServiceCategory(t *testing.T) {
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	tests := []struct {
		name      string
		category  *ServiceCategory
		userID    string
		setupFunc func(*MockServicesRepository)
		wantErr   bool
	}{
		{
			name: "successfully update existing service category",
			category: &ServiceCategory{
				CategoryID: "550e8400-e29b-41d4-a716-446655440002",
				Name:       "Updated Category Name",
				Slug:       "updated-category-name",
			},
			userID: "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockServicesRepository) {
				category := createTestCategory("admin-550e8400-e29b-41d4-a716-446655440003")
				category.CategoryID = "550e8400-e29b-41d4-a716-446655440002"
				repo.categories["550e8400-e29b-41d4-a716-446655440002"] = category
			},
			wantErr: false,
		},
		{
			name: "return not found error for non-existent category",
			category: &ServiceCategory{
				CategoryID: "550e8400-e29b-41d4-a716-446655440999",
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockServicesRepository) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockServicesRepository()
			tt.setupFunc(repo)
			service := NewServicesService(repo)

			err := service.AdminUpdateServiceCategory(ctx, tt.category, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, repo.auditEvents, 1)
				assert.Equal(t, domain.AuditEventUpdate, repo.auditEvents[0].OperationType)
			}
		})
	}
}

func TestServicesService_AdminDeleteServiceCategory(t *testing.T) {
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	tests := []struct {
		name       string
		categoryID string
		userID     string
		setupFunc  func(*MockServicesRepository)
		wantErr    bool
	}{
		{
			name:       "successfully delete existing service category",
			categoryID: "550e8400-e29b-41d4-a716-446655440002",
			userID:     "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockServicesRepository) {
				category := createTestCategory("admin-550e8400-e29b-41d4-a716-446655440003")
				category.CategoryID = "550e8400-e29b-41d4-a716-446655440002"
				category.IsDefaultUnassigned = false
				repo.categories["550e8400-e29b-41d4-a716-446655440002"] = category
			},
			wantErr: false,
		},
		{
			name:       "return validation error for default unassigned category",
			categoryID: "550e8400-e29b-41d4-a716-446655440002",
			userID:     "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockServicesRepository) {
				category := createTestCategory("admin-550e8400-e29b-41d4-a716-446655440003")
				category.CategoryID = "550e8400-e29b-41d4-a716-446655440002"
				category.IsDefaultUnassigned = true
				repo.categories["550e8400-e29b-41d4-a716-446655440002"] = category
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockServicesRepository()
			tt.setupFunc(repo)
			service := NewServicesService(repo)

			err := service.AdminDeleteServiceCategory(ctx, tt.categoryID, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, repo.auditEvents, 1)
				assert.Equal(t, domain.AuditEventDelete, repo.auditEvents[0].OperationType)
			}
		})
	}
}

func TestServicesService_AdminSetFeaturedCategories(t *testing.T) {
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	tests := []struct {
		name        string
		categoryIDs []string
		userID      string
		setupFunc   func(*MockServicesRepository)
		wantErr     bool
	}{
		{
			name:        "successfully set featured categories",
			categoryIDs: []string{"550e8400-e29b-41d4-a716-446655440002", "550e8400-e29b-41d4-a716-446655440007"},
			userID:      "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockServicesRepository) {
				category1 := createTestCategory("admin-550e8400-e29b-41d4-a716-446655440003")
				category1.CategoryID = "550e8400-e29b-41d4-a716-446655440002"
				repo.categories["550e8400-e29b-41d4-a716-446655440002"] = category1

				category2 := createTestCategory("admin-550e8400-e29b-41d4-a716-446655440003")
				category2.CategoryID = "550e8400-e29b-41d4-a716-446655440007"
				repo.categories["550e8400-e29b-41d4-a716-446655440007"] = category2
			},
			wantErr: false,
		},
		{
			name:        "return validation error for deleted category",
			categoryIDs: []string{"550e8400-e29b-41d4-a716-446655440002"},
			userID:      "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockServicesRepository) {
				category := createTestCategory("admin-550e8400-e29b-41d4-a716-446655440003")
				category.CategoryID = "550e8400-e29b-41d4-a716-446655440002"
				category.IsDeleted = true // Make category soft-deleted
				repo.categories["550e8400-e29b-41d4-a716-446655440002"] = category
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockServicesRepository()
			tt.setupFunc(repo)
			service := NewServicesService(repo)

			err := service.AdminSetFeaturedCategories(ctx, tt.categoryIDs, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Should have audit events for setting featured categories
				assert.True(t, len(repo.auditEvents) >= 1)
			}
		})
	}
}