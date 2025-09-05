package events

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to assert error type
func assertErrorType(t *testing.T, err error, expectedType string) {
	switch expectedType {
	case "validation":
		assert.True(t, domain.IsValidationError(err), "expected validation error")
	case "not_found":
		assert.True(t, domain.IsNotFoundError(err), "expected not found error")
	case "unauthorized":
		assert.True(t, domain.IsUnauthorizedError(err), "expected unauthorized error")
	case "forbidden":
		assert.True(t, domain.IsForbiddenError(err), "expected forbidden error")
	case "conflict":
		assert.True(t, domain.IsConflictError(err), "expected conflict error")
	default:
		t.Fatalf("unknown error type: %s", expectedType)
	}
}

// MockEventsRepository provides mock implementation for unit tests
type MockEventsRepository struct {
	events             map[string]*Event
	categories         map[string]*EventCategory
	featuredEvent      *FeaturedEvent
	registrations      map[string]*EventRegistration
	auditEvents        []MockAuditEvent
	failures           map[string]error
}

type MockAuditEvent struct {
	EntityType    domain.EntityType
	EntityID      string
	OperationType domain.AuditEventType
	UserID        string
	Before        interface{}
	After         interface{}
}

func NewMockEventsRepository() *MockEventsRepository {
	return &MockEventsRepository{
		events:        make(map[string]*Event),
		categories:    make(map[string]*EventCategory),
		registrations: make(map[string]*EventRegistration),
		auditEvents:   make([]MockAuditEvent, 0),
		failures:      make(map[string]error),
	}
}

// SetFailure sets a mock failure for specific operations
func (m *MockEventsRepository) SetFailure(operation string, err error) {
	m.failures[operation] = err
}

// GetAuditEvents returns all mock audit events
func (m *MockEventsRepository) GetAuditEvents() []MockAuditEvent {
	return m.auditEvents
}

// Repository interface methods
func (m *MockEventsRepository) SaveEvent(ctx context.Context, event *Event) error {
	if err := m.failures["SaveEvent"]; err != nil {
		return err
	}
	m.events[event.EventID] = event
	return nil
}

func (m *MockEventsRepository) GetEvent(ctx context.Context, eventID string) (*Event, error) {
	if err := m.failures["GetEvent"]; err != nil {
		return nil, err
	}
	event, exists := m.events[eventID]
	if !exists {
		return nil, domain.NewNotFoundError("event", eventID)
	}
	return event, nil
}

func (m *MockEventsRepository) DeleteEvent(ctx context.Context, eventID string, userID string) error {
	if err := m.failures["DeleteEvent"]; err != nil {
		return err
	}
	event, exists := m.events[eventID]
	if !exists {
		return domain.NewNotFoundError("event", eventID)
	}
	event.IsDeleted = true
	event.DeletedOn = &[]time.Time{time.Now()}[0]
	event.DeletedBy = &userID
	return nil
}

func (m *MockEventsRepository) SaveEventCategory(ctx context.Context, category *EventCategory) error {
	if err := m.failures["SaveEventCategory"]; err != nil {
		return err
	}
	m.categories[category.CategoryID] = category
	return nil
}

func (m *MockEventsRepository) GetEventCategory(ctx context.Context, categoryID string) (*EventCategory, error) {
	if err := m.failures["GetEventCategory"]; err != nil {
		return nil, err
	}
	category, exists := m.categories[categoryID]
	if !exists {
		return nil, domain.NewNotFoundError("category", categoryID)
	}
	return category, nil
}

func (m *MockEventsRepository) DeleteEventCategory(ctx context.Context, categoryID string, userID string) error {
	if err := m.failures["DeleteEventCategory"]; err != nil {
		return err
	}
	category, exists := m.categories[categoryID]
	if !exists {
		return domain.NewNotFoundError("category", categoryID)
	}
	category.IsDeleted = true
	category.DeletedOn = &[]time.Time{time.Now()}[0]
	category.DeletedBy = &userID
	return nil
}

func (m *MockEventsRepository) GetDefaultUnassignedCategory(ctx context.Context) (*EventCategory, error) {
	if err := m.failures["GetDefaultUnassignedCategory"]; err != nil {
		return nil, err
	}
	for _, category := range m.categories {
		if category.IsDefaultUnassigned && !category.IsDeleted {
			return category, nil
		}
	}
	return nil, domain.NewNotFoundError("default unassigned category", "unassigned")
}

func (m *MockEventsRepository) SaveFeaturedEvent(ctx context.Context, featuredEvent *FeaturedEvent) error {
	if err := m.failures["SaveFeaturedEvent"]; err != nil {
		return err
	}
	m.featuredEvent = featuredEvent
	return nil
}

func (m *MockEventsRepository) GetFeaturedEvent(ctx context.Context) (*FeaturedEvent, error) {
	if err := m.failures["GetFeaturedEvent"]; err != nil {
		return nil, err
	}
	if m.featuredEvent == nil {
		return nil, domain.NewNotFoundError("featured event", "current")
	}
	return m.featuredEvent, nil
}

func (m *MockEventsRepository) DeleteFeaturedEvent(ctx context.Context) error {
	if err := m.failures["DeleteFeaturedEvent"]; err != nil {
		return err
	}
	m.featuredEvent = nil
	return nil
}

func (m *MockEventsRepository) GetEventRegistrations(ctx context.Context, eventID string) ([]*EventRegistration, error) {
	if err := m.failures["GetEventRegistrations"]; err != nil {
		return nil, err
	}
	var registrations []*EventRegistration
	for _, registration := range m.registrations {
		if registration.EventID == eventID && !registration.IsDeleted {
			registrations = append(registrations, registration)
		}
	}
	return registrations, nil
}

func (m *MockEventsRepository) PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error {
	if err := m.failures["PublishAuditEvent"]; err != nil {
		return err
	}
	m.auditEvents = append(m.auditEvents, MockAuditEvent{
		EntityType:    entityType,
		EntityID:      entityID,
		OperationType: operationType,
		UserID:        userID,
		Before:        beforeData,
		After:         afterData,
	})
	return nil
}

// Test helper functions
func createTestEvent(eventID, title, categoryID, userID string) *Event {
	eventDate := time.Date(2024, 12, 15, 0, 0, 0, 0, time.UTC)
	event := &Event{
		EventID:          eventID,
		Title:            title,
		Description:      "Test event description",
		Slug:             "test-event",
		CategoryID:       categoryID,
		EventDate:        eventDate,
		Location:         "Test Location",
		EventType:        EventTypeWorkshop,
		PublishingStatus: PublishingStatusDraft,
		RegistrationStatus: RegistrationStatusOpen,
		PriorityLevel:    PriorityLevelNormal,
		CreatedOn:        time.Now(),
		CreatedBy:        &userID,
		IsDeleted:        false,
	}
	return event
}

func createTestEventCategory(categoryID, name, userID string) *EventCategory {
	return &EventCategory{
		CategoryID:          categoryID,
		Name:                name,
		Slug:                "test-category",
		Description:         &[]string{"Test category description"}[0],
		IsDefaultUnassigned: false,
		CreatedOn:           time.Now(),
		CreatedBy:           &userID,
		IsDeleted:           false,
	}
}

func createDefaultUnassignedCategory(categoryID, userID string) *EventCategory {
	return &EventCategory{
		CategoryID:          categoryID,
		Name:                "Unassigned",
		Slug:                "unassigned",
		Description:         &[]string{"Default unassigned category"}[0],
		IsDefaultUnassigned: true,
		CreatedOn:           time.Now(),
		CreatedBy:           &userID,
		IsDeleted:           false,
	}
}

func TestEventsService_AdminCreateEvent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockEventsRepository)
		userID    string
		request   AdminCreateEventRequest
		wantError bool
		errorType string
	}{
		{
			name: "successfully create event with valid data",
			setupFunc: func(repo *MockEventsRepository) {
				category := createTestEventCategory("550e8400-e29b-41d4-a716-446655440001", "Test Category", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.categories[category.CategoryID] = category
			},
			userID: "admin-550e8400-e29b-41d4-a716-446655440003",
			request: AdminCreateEventRequest{
				Title:       "Test Event",
				Description: "Test event description",
				CategoryID:  "550e8400-e29b-41d4-a716-446655440001",
				EventDate:   "2024-12-15",
				EventTime:   &[]string{"10:00"}[0],
				Location:    "Test Location",
				EventType:   string(EventTypeWorkshop),
			},
			wantError: false,
		},
		{
			name:      "return validation error for empty title",
			setupFunc: func(repo *MockEventsRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			request: AdminCreateEventRequest{
				Title:       "",
				Description: "Test event description",
				CategoryID:  "550e8400-e29b-41d4-a716-446655440001",
				EventDate:   "2024-12-15",
				Location:    "Test Location",
				EventType:   string(EventTypeWorkshop),
			},
			wantError: true,
			errorType: "validation",
		},
		{
			name:      "return unauthorized error for non-admin user",
			setupFunc: func(repo *MockEventsRepository) {},
			userID:    "regular-user-id",
			request: AdminCreateEventRequest{
				Title:       "Test Event",
				Description: "Test event description",
				CategoryID:  "550e8400-e29b-41d4-a716-446655440001",
				EventDate:   "2024-12-15",
				Location:    "Test Location",
				EventType:   string(EventTypeWorkshop),
			},
			wantError: true,
			errorType: "unauthorized",
		},
		{
			name: "return validation error for invalid category",
			setupFunc: func(repo *MockEventsRepository) {
				repo.SetFailure("GetEventCategory", domain.NewNotFoundError("category", "invalid-category"))
			},
			userID: "admin-550e8400-e29b-41d4-a716-446655440003",
			request: AdminCreateEventRequest{
				Title:       "Test Event",
				Description: "Test event description",
				CategoryID:  "invalid-category",
				EventDate:   "2024-12-15",
				Location:    "Test Location",
				EventType:   string(EventTypeWorkshop),
			},
			wantError: true,
			errorType: "not_found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockEventsRepository()
			tt.setupFunc(repo)
			service := NewEventsService(repo)

			event, err := service.AdminCreateEvent(ctx, tt.request, tt.userID)

			if tt.wantError {
				require.Error(t, err)
				assertErrorType(t, err, tt.errorType)
			} else {
				require.NoError(t, err)
				require.NotNil(t, event)
				assert.Equal(t, tt.request.Title, event.Title)
				assert.Equal(t, tt.request.Description, event.Description)
				assert.Equal(t, tt.request.CategoryID, event.CategoryID)
				assert.Equal(t, PublishingStatusDraft, event.PublishingStatus)
				assert.Equal(t, EventType(tt.request.EventType), event.EventType)
			}
		})
	}
}

func TestEventsService_AdminUpdateEvent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockEventsRepository)
		userID    string
		eventID   string
		request   AdminUpdateEventRequest
		wantError bool
		errorType string
	}{
		{
			name: "successfully update event with valid data",
			setupFunc: func(repo *MockEventsRepository) {
				category := createTestEventCategory("550e8400-e29b-41d4-a716-446655440001", "Test Category", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.categories[category.CategoryID] = category
				event := createTestEvent("550e8400-e29b-41d4-a716-446655440002", "Original Title", category.CategoryID, "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.events[event.EventID] = event
			},
			userID:  "admin-550e8400-e29b-41d4-a716-446655440003",
			eventID: "550e8400-e29b-41d4-a716-446655440002",
			request: AdminUpdateEventRequest{
				Title:       &[]string{"Updated Title"}[0],
				Description: &[]string{"Updated description"}[0],
				Location:    &[]string{"Updated Location"}[0],
			},
			wantError: false,
		},
		{
			name:      "return not found error for non-existent event",
			setupFunc: func(repo *MockEventsRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			eventID:   "non-existent-id",
			request: AdminUpdateEventRequest{
				Title: &[]string{"Updated Title"}[0],
			},
			wantError: true,
			errorType: "not_found",
		},
		{
			name:      "return unauthorized error for non-admin user",
			setupFunc: func(repo *MockEventsRepository) {},
			userID:    "regular-user-id",
			eventID:   "550e8400-e29b-41d4-a716-446655440002",
			request: AdminUpdateEventRequest{
				Title: &[]string{"Updated Title"}[0],
			},
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockEventsRepository()
			tt.setupFunc(repo)
			service := NewEventsService(repo)

			event, err := service.AdminUpdateEvent(ctx, tt.eventID, tt.request, tt.userID)

			if tt.wantError {
				require.Error(t, err)
				assertErrorType(t, err, tt.errorType)
			} else {
				require.NoError(t, err)
				require.NotNil(t, event)
				if tt.request.Title != nil {
					assert.Equal(t, *tt.request.Title, event.Title)
				}
				if tt.request.Description != nil {
					assert.Equal(t, *tt.request.Description, event.Description)
				}
			}
		})
	}
}

func TestEventsService_AdminDeleteEvent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockEventsRepository)
		userID    string
		eventID   string
		wantError bool
		errorType string
	}{
		{
			name: "successfully delete event",
			setupFunc: func(repo *MockEventsRepository) {
				category := createTestEventCategory("550e8400-e29b-41d4-a716-446655440001", "Test Category", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.categories[category.CategoryID] = category
				event := createTestEvent("550e8400-e29b-41d4-a716-446655440002", "Test Event", category.CategoryID, "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.events[event.EventID] = event
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			eventID:   "550e8400-e29b-41d4-a716-446655440002",
			wantError: false,
		},
		{
			name:      "return not found error for non-existent event",
			setupFunc: func(repo *MockEventsRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			eventID:   "non-existent-id",
			wantError: true,
			errorType: "not_found",
		},
		{
			name:      "return unauthorized error for non-admin user",
			setupFunc: func(repo *MockEventsRepository) {},
			userID:    "regular-user-id",
			eventID:   "550e8400-e29b-41d4-a716-446655440002",
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockEventsRepository()
			tt.setupFunc(repo)
			service := NewEventsService(repo)

			err := service.AdminDeleteEvent(ctx, tt.eventID, tt.userID)

			if tt.wantError {
				require.Error(t, err)
				assertErrorType(t, err, tt.errorType)
			} else {
				require.NoError(t, err)
				// Verify event is soft deleted
				event, getErr := repo.GetEvent(ctx, tt.eventID)
				require.NoError(t, getErr)
				assert.True(t, event.IsDeleted)
				assert.NotNil(t, event.DeletedOn)
				assert.Equal(t, tt.userID, *event.DeletedBy)
			}
		})
	}
}

func TestEventsService_AdminPublishEvent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockEventsRepository)
		userID    string
		eventID   string
		wantError bool
		errorType string
	}{
		{
			name: "successfully publish draft event",
			setupFunc: func(repo *MockEventsRepository) {
				category := createTestEventCategory("550e8400-e29b-41d4-a716-446655440001", "Test Category", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.categories[category.CategoryID] = category
				event := createTestEvent("550e8400-e29b-41d4-a716-446655440002", "Test Event", category.CategoryID, "admin-550e8400-e29b-41d4-a716-446655440003")
				event.PublishingStatus = PublishingStatusDraft
				repo.events[event.EventID] = event
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			eventID:   "550e8400-e29b-41d4-a716-446655440002",
			wantError: false,
		},
		{
			name: "return validation error for already published event",
			setupFunc: func(repo *MockEventsRepository) {
				category := createTestEventCategory("550e8400-e29b-41d4-a716-446655440001", "Test Category", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.categories[category.CategoryID] = category
				event := createTestEvent("550e8400-e29b-41d4-a716-446655440002", "Test Event", category.CategoryID, "admin-550e8400-e29b-41d4-a716-446655440003")
				event.PublishingStatus = PublishingStatusPublished
				repo.events[event.EventID] = event
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			eventID:   "550e8400-e29b-41d4-a716-446655440002",
			wantError: true,
			errorType: "validation",
		},
		{
			name:      "return not found error for non-existent event",
			setupFunc: func(repo *MockEventsRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			eventID:   "non-existent-id",
			wantError: true,
			errorType: "not_found",
		},
		{
			name:      "return unauthorized error for non-admin user",
			setupFunc: func(repo *MockEventsRepository) {},
			userID:    "regular-user-id",
			eventID:   "550e8400-e29b-41d4-a716-446655440002",
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockEventsRepository()
			tt.setupFunc(repo)
			service := NewEventsService(repo)

			event, err := service.AdminPublishEvent(ctx, tt.eventID, tt.userID)

			if tt.wantError {
				require.Error(t, err)
				assertErrorType(t, err, tt.errorType)
			} else {
				require.NoError(t, err)
				require.NotNil(t, event)
				assert.Equal(t, PublishingStatusPublished, event.PublishingStatus)
			}
		})
	}
}

func TestEventsService_AdminArchiveEvent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockEventsRepository)
		userID    string
		eventID   string
		wantError bool
		errorType string
	}{
		{
			name: "successfully archive published event",
			setupFunc: func(repo *MockEventsRepository) {
				category := createTestEventCategory("550e8400-e29b-41d4-a716-446655440001", "Test Category", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.categories[category.CategoryID] = category
				event := createTestEvent("550e8400-e29b-41d4-a716-446655440002", "Test Event", category.CategoryID, "admin-550e8400-e29b-41d4-a716-446655440003")
				event.PublishingStatus = PublishingStatusPublished
				repo.events[event.EventID] = event
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			eventID:   "550e8400-e29b-41d4-a716-446655440002",
			wantError: false,
		},
		{
			name: "return validation error for draft event",
			setupFunc: func(repo *MockEventsRepository) {
				category := createTestEventCategory("550e8400-e29b-41d4-a716-446655440001", "Test Category", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.categories[category.CategoryID] = category
				event := createTestEvent("550e8400-e29b-41d4-a716-446655440002", "Test Event", category.CategoryID, "admin-550e8400-e29b-41d4-a716-446655440003")
				event.PublishingStatus = PublishingStatusDraft
				repo.events[event.EventID] = event
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			eventID:   "550e8400-e29b-41d4-a716-446655440002",
			wantError: true,
			errorType: "validation",
		},
		{
			name:      "return not found error for non-existent event",
			setupFunc: func(repo *MockEventsRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			eventID:   "non-existent-id",
			wantError: true,
			errorType: "not_found",
		},
		{
			name:      "return unauthorized error for non-admin user",
			setupFunc: func(repo *MockEventsRepository) {},
			userID:    "regular-user-id",
			eventID:   "550e8400-e29b-41d4-a716-446655440002",
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockEventsRepository()
			tt.setupFunc(repo)
			service := NewEventsService(repo)

			event, err := service.AdminArchiveEvent(ctx, tt.eventID, tt.userID)

			if tt.wantError {
				require.Error(t, err)
				assertErrorType(t, err, tt.errorType)
			} else {
				require.NoError(t, err)
				require.NotNil(t, event)
				assert.Equal(t, PublishingStatusArchived, event.PublishingStatus)
			}
		})
	}
}

func TestEventsService_AdminCreateEventCategory(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockEventsRepository)
		userID    string
		request   AdminCreateEventCategoryRequest
		wantError bool
		errorType string
	}{
		{
			name:      "successfully create event category",
			setupFunc: func(repo *MockEventsRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			request: AdminCreateEventCategoryRequest{
				Name:        "Test Category",
				Description: "Test category description",
			},
			wantError: false,
		},
		{
			name:      "return validation error for empty name",
			setupFunc: func(repo *MockEventsRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			request: AdminCreateEventCategoryRequest{
				Name:        "",
				Description: "Test category description",
			},
			wantError: true,
			errorType: "validation",
		},
		{
			name:      "return unauthorized error for non-admin user",
			setupFunc: func(repo *MockEventsRepository) {},
			userID:    "regular-user-id",
			request: AdminCreateEventCategoryRequest{
				Name:        "Test Category",
				Description: "Test category description",
			},
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockEventsRepository()
			tt.setupFunc(repo)
			service := NewEventsService(repo)

			category, err := service.AdminCreateEventCategory(ctx, tt.request, tt.userID)

			if tt.wantError {
				require.Error(t, err)
				assertErrorType(t, err, tt.errorType)
			} else {
				require.NoError(t, err)
				require.NotNil(t, category)
				assert.Equal(t, tt.request.Name, category.Name)
				assert.Equal(t, tt.request.Description, *category.Description)
				assert.False(t, category.IsDefaultUnassigned)
			}
		})
	}
}

func TestEventsService_AdminDeleteEventCategory(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockEventsRepository)
		userID    string
		categoryID string
		wantError bool
		errorType string
	}{
		{
			name: "successfully delete category and reassign events to default",
			setupFunc: func(repo *MockEventsRepository) {
				// Create default unassigned category
				defaultCategory := createDefaultUnassignedCategory("550e8400-e29b-41d4-a716-446655440001", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.categories[defaultCategory.CategoryID] = defaultCategory

				// Create category to be deleted
				categoryToDelete := createTestEventCategory("550e8400-e29b-41d4-a716-446655440002", "Category To Delete", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.categories[categoryToDelete.CategoryID] = categoryToDelete

				// Create event in category to be deleted
				event := createTestEvent("550e8400-e29b-41d4-a716-446655440003", "Test Event", categoryToDelete.CategoryID, "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.events[event.EventID] = event
			},
			userID:     "admin-550e8400-e29b-41d4-a716-446655440003",
			categoryID: "550e8400-e29b-41d4-a716-446655440002",
			wantError:  false,
		},
		{
			name: "return validation error for default unassigned category",
			setupFunc: func(repo *MockEventsRepository) {
				defaultCategory := createDefaultUnassignedCategory("550e8400-e29b-41d4-a716-446655440001", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.categories[defaultCategory.CategoryID] = defaultCategory
			},
			userID:     "admin-550e8400-e29b-41d4-a716-446655440003",
			categoryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError:  true,
			errorType:  "validation",
		},
		{
			name:       "return not found error for non-existent category",
			setupFunc:  func(repo *MockEventsRepository) {},
			userID:     "admin-550e8400-e29b-41d4-a716-446655440003",
			categoryID: "non-existent-id",
			wantError:  true,
			errorType:  "not_found",
		},
		{
			name:       "return unauthorized error for non-admin user",
			setupFunc:  func(repo *MockEventsRepository) {},
			userID:     "regular-user-id",
			categoryID: "550e8400-e29b-41d4-a716-446655440002",
			wantError:  true,
			errorType:  "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockEventsRepository()
			tt.setupFunc(repo)
			service := NewEventsService(repo)

			err := service.AdminDeleteEventCategory(ctx, tt.categoryID, tt.userID)

			if tt.wantError {
				require.Error(t, err)
				assertErrorType(t, err, tt.errorType)
			} else {
				require.NoError(t, err)
				// Verify category is soft deleted
				category, getErr := repo.GetEventCategory(ctx, tt.categoryID)
				require.NoError(t, getErr)
				assert.True(t, category.IsDeleted)
				assert.NotNil(t, category.DeletedOn)
				assert.Equal(t, tt.userID, *category.DeletedBy)
			}
		})
	}
}

func TestEventsService_AdminSetFeaturedEvent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockEventsRepository)
		userID    string
		eventID   string
		wantError bool
		errorType string
	}{
		{
			name: "successfully set featured event",
			setupFunc: func(repo *MockEventsRepository) {
				category := createTestEventCategory("550e8400-e29b-41d4-a716-446655440001", "Test Category", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.categories[category.CategoryID] = category
				event := createTestEvent("550e8400-e29b-41d4-a716-446655440002", "Test Event", category.CategoryID, "admin-550e8400-e29b-41d4-a716-446655440003")
				event.PublishingStatus = PublishingStatusPublished
				repo.events[event.EventID] = event
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			eventID:   "550e8400-e29b-41d4-a716-446655440002",
			wantError: false,
		},
		{
			name: "return validation error for draft event",
			setupFunc: func(repo *MockEventsRepository) {
				category := createTestEventCategory("550e8400-e29b-41d4-a716-446655440001", "Test Category", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.categories[category.CategoryID] = category
				event := createTestEvent("550e8400-e29b-41d4-a716-446655440002", "Test Event", category.CategoryID, "admin-550e8400-e29b-41d4-a716-446655440003")
				event.PublishingStatus = PublishingStatusDraft
				repo.events[event.EventID] = event
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			eventID:   "550e8400-e29b-41d4-a716-446655440002",
			wantError: true,
			errorType: "validation",
		},
		{
			name: "return validation error for event in default unassigned category",
			setupFunc: func(repo *MockEventsRepository) {
				defaultCategory := createDefaultUnassignedCategory("550e8400-e29b-41d4-a716-446655440001", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.categories[defaultCategory.CategoryID] = defaultCategory
				event := createTestEvent("550e8400-e29b-41d4-a716-446655440002", "Test Event", defaultCategory.CategoryID, "admin-550e8400-e29b-41d4-a716-446655440003")
				event.PublishingStatus = PublishingStatusPublished
				repo.events[event.EventID] = event
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			eventID:   "550e8400-e29b-41d4-a716-446655440002",
			wantError: true,
			errorType: "validation",
		},
		{
			name:      "return not found error for non-existent event",
			setupFunc: func(repo *MockEventsRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			eventID:   "non-existent-id",
			wantError: true,
			errorType: "not_found",
		},
		{
			name:      "return unauthorized error for non-admin user",
			setupFunc: func(repo *MockEventsRepository) {},
			userID:    "regular-user-id",
			eventID:   "550e8400-e29b-41d4-a716-446655440002",
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockEventsRepository()
			tt.setupFunc(repo)
			service := NewEventsService(repo)

			featuredEvent, err := service.AdminSetFeaturedEvent(ctx, tt.eventID, tt.userID)

			if tt.wantError {
				require.Error(t, err)
				assertErrorType(t, err, tt.errorType)
			} else {
				require.NoError(t, err)
				require.NotNil(t, featuredEvent)
				assert.Equal(t, tt.eventID, featuredEvent.EventID)
			}
		})
	}
}

func TestEventsService_AdminGetEventRegistrations(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockEventsRepository)
		userID    string
		eventID   string
		wantError bool
		errorType string
		wantCount int
	}{
		{
			name: "successfully get event registrations",
			setupFunc: func(repo *MockEventsRepository) {
				category := createTestEventCategory("550e8400-e29b-41d4-a716-446655440001", "Test Category", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.categories[category.CategoryID] = category
				event := createTestEvent("550e8400-e29b-41d4-a716-446655440002", "Test Event", category.CategoryID, "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.events[event.EventID] = event

				// Add registrations
				registration1 := &EventRegistration{
					RegistrationID:    "550e8400-e29b-41d4-a716-446655440003",
					EventID:           event.EventID,
					ParticipantName:   "John Doe",
					ParticipantEmail:  "john@example.com",
					RegistrationStatus: RegistrationStatusRegistered,
					CreatedOn:         time.Now(),
					IsDeleted:         false,
				}
				registration2 := &EventRegistration{
					RegistrationID:    "550e8400-e29b-41d4-a716-446655440004",
					EventID:           event.EventID,
					ParticipantName:   "Jane Smith",
					ParticipantEmail:  "jane@example.com",
					RegistrationStatus: RegistrationStatusConfirmed,
					CreatedOn:         time.Now(),
					IsDeleted:         false,
				}
				repo.registrations[registration1.RegistrationID] = registration1
				repo.registrations[registration2.RegistrationID] = registration2
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			eventID:   "550e8400-e29b-41d4-a716-446655440002",
			wantError: false,
			wantCount: 2,
		},
		{
			name:      "return not found error for non-existent event",
			setupFunc: func(repo *MockEventsRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			eventID:   "non-existent-id",
			wantError: true,
			errorType: "not_found",
		},
		{
			name:      "return unauthorized error for non-admin user",
			setupFunc: func(repo *MockEventsRepository) {},
			userID:    "regular-user-id",
			eventID:   "550e8400-e29b-41d4-a716-446655440002",
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockEventsRepository()
			tt.setupFunc(repo)
			service := NewEventsService(repo)

			registrations, err := service.AdminGetEventRegistrations(ctx, tt.eventID, tt.userID)

			if tt.wantError {
				require.Error(t, err)
				assertErrorType(t, err, tt.errorType)
			} else {
				require.NoError(t, err)
				require.NotNil(t, registrations)
				assert.Len(t, registrations, tt.wantCount)
			}
		})
	}
}

// RED PHASE - Domain enum validation tests (will fail until IsValid methods are implemented)

func TestEventType_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		want      bool
	}{
		{
			name:      "valid workshop",
			eventType: EventTypeWorkshop,
			want:      true,
		},
		{
			name:      "valid seminar",
			eventType: EventTypeSeminar,
			want:      true,
		},
		{
			name:      "valid webinar",
			eventType: EventTypeWebinar,
			want:      true,
		},
		{
			name:      "valid conference",
			eventType: EventTypeConference,
			want:      true,
		},
		{
			name:      "valid fundraiser",
			eventType: EventTypeFundraiser,
			want:      true,
		},
		{
			name:      "valid community",
			eventType: EventTypeCommunity,
			want:      true,
		},
		{
			name:      "valid medical",
			eventType: EventTypeMedical,
			want:      true,
		},
		{
			name:      "valid educational",
			eventType: EventTypeEducational,
			want:      true,
		},
		{
			name:      "invalid empty event type",
			eventType: EventType(""),
			want:      false,
		},
		{
			name:      "invalid unknown event type",
			eventType: EventType("unknown_type"),
			want:      false,
		},
		{
			name:      "invalid mixed case event type",
			eventType: EventType("Workshop"),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until we implement EventType.IsValid() method
			got := tt.eventType.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPublishingStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status PublishingStatus
		want   bool
	}{
		{
			name:   "valid draft status",
			status: PublishingStatusDraft,
			want:   true,
		},
		{
			name:   "valid published status",
			status: PublishingStatusPublished,
			want:   true,
		},
		{
			name:   "valid archived status",
			status: PublishingStatusArchived,
			want:   true,
		},
		{
			name:   "invalid empty status",
			status: PublishingStatus(""),
			want:   false,
		},
		{
			name:   "invalid unknown status",
			status: PublishingStatus("unknown_status"),
			want:   false,
		},
		{
			name:   "invalid mixed case status",
			status: PublishingStatus("DRAFT"),
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until we implement PublishingStatus.IsValid() method
			got := tt.status.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRegistrationStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status RegistrationStatus
		want   bool
	}{
		{
			name:   "valid open status",
			status: RegistrationStatusOpen,
			want:   true,
		},
		{
			name:   "valid registration_required status",
			status: RegistrationStatusRegistrationRequired,
			want:   true,
		},
		{
			name:   "valid full status",
			status: RegistrationStatusFull,
			want:   true,
		},
		{
			name:   "valid cancelled status",
			status: RegistrationStatusCancelled,
			want:   true,
		},
		{
			name:   "valid registered status",
			status: RegistrationStatusRegistered,
			want:   true,
		},
		{
			name:   "valid confirmed status",
			status: RegistrationStatusConfirmed,
			want:   true,
		},
		{
			name:   "valid no_show status",
			status: RegistrationStatusNoShow,
			want:   true,
		},
		{
			name:   "invalid empty status",
			status: RegistrationStatus(""),
			want:   false,
		},
		{
			name:   "invalid unknown status",
			status: RegistrationStatus("unknown_status"),
			want:   false,
		},
		{
			name:   "invalid mixed case status",
			status: RegistrationStatus("OPEN"),
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until we implement RegistrationStatus.IsValid() method
			got := tt.status.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPriorityLevel_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		level PriorityLevel
		want  bool
	}{
		{
			name:  "valid low priority",
			level: PriorityLevelLow,
			want:  true,
		},
		{
			name:  "valid normal priority",
			level: PriorityLevelNormal,
			want:  true,
		},
		{
			name:  "valid high priority",
			level: PriorityLevelHigh,
			want:  true,
		},
		{
			name:  "valid urgent priority",
			level: PriorityLevelUrgent,
			want:  true,
		},
		{
			name:  "invalid empty priority",
			level: PriorityLevel(""),
			want:  false,
		},
		{
			name:  "invalid unknown priority",
			level: PriorityLevel("unknown_priority"),
			want:  false,
		},
		{
			name:  "invalid mixed case priority",
			level: PriorityLevel("LOW"),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until we implement PriorityLevel.IsValid() method
			got := tt.level.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

// Domain validation helper tests (will fail until comprehensive validation helpers are implemented)

func TestValidateEventTitle(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		wantError   bool
		errorMsg    string
	}{
		{
			name:      "valid event title",
			title:     "Healthcare Innovation Workshop",
			wantError: false,
		},
		{
			name:      "valid title with maximum length",
			title:     strings.Repeat("a", 255),
			wantError: false,
		},
		{
			name:      "valid title with minimum length",
			title:     "ab",
			wantError: false,
		},
		{
			name:      "invalid empty title",
			title:     "",
			wantError: true,
			errorMsg:  "title is required",
		},
		{
			name:      "invalid whitespace-only title",
			title:     "   ",
			wantError: true,
			errorMsg:  "title is required",
		},
		{
			name:      "invalid title too short",
			title:     "a",
			wantError: true,
			errorMsg:  "title must be between 2 and 255 characters",
		},
		{
			name:      "invalid title too long",
			title:     strings.Repeat("a", 256),
			wantError: true,
			errorMsg:  "title must be between 2 and 255 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until we implement validateEventTitle helper function
			err := validateEventTitle(tt.title)
			
			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateEventDescription(t *testing.T) {
	tests := []struct {
		name        string
		description string
		wantError   bool
		errorMsg    string
	}{
		{
			name:        "valid event description",
			description: "Join us for an innovative healthcare workshop focusing on cutting-edge medical technologies and patient care strategies.",
			wantError:   false,
		},
		{
			name:        "valid description with minimum length",
			description: "abcdefghij", // 10 characters
			wantError:   false,
		},
		{
			name:        "invalid empty description",
			description: "",
			wantError:   true,
			errorMsg:    "description is required",
		},
		{
			name:        "invalid whitespace-only description",
			description: "   ",
			wantError:   true,
			errorMsg:    "description is required",
		},
		{
			name:        "invalid description too short",
			description: "short",
			wantError:   true,
			errorMsg:    "description must be between 10 and 2000 characters",
		},
		{
			name:        "invalid description too long",
			description: strings.Repeat("a", 2001),
			wantError:   true,
			errorMsg:    "description must be between 10 and 2000 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until we implement validateEventDescription helper function
			err := validateEventDescription(tt.description)
			
			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateEventLocation(t *testing.T) {
	tests := []struct {
		name      string
		location  string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid event location",
			location:  "Main Conference Hall, Building A",
			wantError: false,
		},
		{
			name:      "valid virtual location",
			location:  "Virtual Event",
			wantError: false,
		},
		{
			name:      "valid location minimum length",
			location:  "ab",
			wantError: false,
		},
		{
			name:      "invalid empty location",
			location:  "",
			wantError: true,
			errorMsg:  "location is required",
		},
		{
			name:      "invalid whitespace-only location",
			location:  "   ",
			wantError: true,
			errorMsg:  "location is required",
		},
		{
			name:      "invalid location too short",
			location:  "a",
			wantError: true,
			errorMsg:  "location must be between 2 and 255 characters",
		},
		{
			name:      "invalid location too long",
			location:  strings.Repeat("a", 256),
			wantError: true,
			errorMsg:  "location must be between 2 and 255 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until we implement validateEventLocation helper function
			err := validateEventLocation(tt.location)
			
			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}