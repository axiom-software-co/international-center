package events

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
)

// EventsRepositoryInterface defines the interface for events domain data operations
type EventsRepositoryInterface interface {
	// Event operations
	SaveEvent(ctx context.Context, event *Event) error
	GetEvent(ctx context.Context, eventID string) (*Event, error)
	DeleteEvent(ctx context.Context, eventID string, userID string) error
	
	// Event category operations
	SaveEventCategory(ctx context.Context, category *EventCategory) error
	GetEventCategory(ctx context.Context, categoryID string) (*EventCategory, error)
	DeleteEventCategory(ctx context.Context, categoryID string, userID string) error
	GetDefaultUnassignedCategory(ctx context.Context) (*EventCategory, error)
	
	// Featured event operations
	SaveFeaturedEvent(ctx context.Context, featuredEvent *FeaturedEvent) error
	GetFeaturedEvent(ctx context.Context) (*FeaturedEvent, error)
	DeleteFeaturedEvent(ctx context.Context) error
	
	// Event registration operations
	GetEventRegistrations(ctx context.Context, eventID string) ([]*EventRegistration, error)
	
	// Audit operations
	PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error
}

// EventsService implements business logic for events operations
type EventsService struct {
	repository EventsRepositoryInterface
}

// NewEventsService creates a new events service
func NewEventsService(repository EventsRepositoryInterface) *EventsService {
	return &EventsService{
		repository: repository,
	}
}

// AdminCreateEvent creates a new event (admin only)
func (s *EventsService) AdminCreateEvent(ctx context.Context, request AdminCreateEventRequest, userID string) (*Event, error) {
	// Validate admin authentication
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required to create events")
	}

	// Validate request
	if err := s.validateCreateEventRequest(request); err != nil {
		return nil, err
	}

	// Verify category exists
	_, err := s.repository.GetEventCategory(ctx, request.CategoryID)
	if err != nil {
		if domain.IsNotFoundError(err) {
			return nil, domain.NewNotFoundError("event category", request.CategoryID)
		}
		return nil, domain.WrapError(err, "failed to validate category")
	}

	// Parse event date
	eventDate, err := time.Parse("2006-01-02", request.EventDate)
	if err != nil {
		return nil, domain.NewValidationError("invalid event date format, use YYYY-MM-DD")
	}

	// Parse event type
	eventType := EventType(request.EventType)
	if !IsValidEventType(string(eventType)) {
		return nil, domain.NewValidationError("invalid event type")
	}

	// Create event entity
	event, err := NewEvent(request.Title, request.Description, request.CategoryID, request.Location, eventDate, eventType, userID)
	if err != nil {
		return nil, domain.WrapError(err, "failed to create event entity")
	}

	// Set optional fields
	if request.Content != nil {
		event.Content = request.Content
	}

	if request.ImageURL != nil {
		if err := event.SetImageURL(*request.ImageURL, userID); err != nil {
			return nil, err
		}
	}

	if request.VirtualLink != nil {
		if err := event.SetVirtualLink(*request.VirtualLink, userID); err != nil {
			return nil, err
		}
	}

	if request.OrganizerName != nil {
		event.OrganizerName = request.OrganizerName
	}

	if request.EventTime != nil {
		if err := s.validateTimeFormat(*request.EventTime); err != nil {
			return nil, err
		}
		event.EventTime = request.EventTime
	}

	if request.EndDate != nil {
		endDate, parseErr := time.Parse("2006-01-02", *request.EndDate)
		if parseErr != nil {
			return nil, domain.NewValidationError("invalid end date format, use YYYY-MM-DD")
		}
		if endDate.Before(eventDate) {
			return nil, domain.NewValidationError("end date cannot be before event date")
		}
		event.EndDate = &endDate
	}

	if request.EndTime != nil {
		if err := s.validateTimeFormat(*request.EndTime); err != nil {
			return nil, err
		}
		event.EndTime = request.EndTime
	}

	if request.MaxCapacity != nil {
		if err := event.SetMaxCapacity(request.MaxCapacity, userID); err != nil {
			return nil, err
		}
	}

	if request.RegistrationDeadline != nil {
		deadline, parseErr := time.Parse(time.RFC3339, *request.RegistrationDeadline)
		if parseErr != nil {
			return nil, domain.NewValidationError("invalid registration deadline format, use YYYY-MM-DDTHH:MM:SSZ")
		}
		if err := event.SetRegistrationDeadline(&deadline, userID); err != nil {
			return nil, err
		}
	}

	if request.PriorityLevel != nil {
		priorityLevel := PriorityLevel(*request.PriorityLevel)
		if err := event.SetPriorityLevel(priorityLevel, userID); err != nil {
			return nil, err
		}
	}

	if len(request.Tags) > 0 {
		if err := event.AddTags(request.Tags, userID); err != nil {
			return nil, err
		}
	}

	// Save event to repository
	if err := s.repository.SaveEvent(ctx, event); err != nil {
		return nil, domain.WrapError(err, "failed to save event")
	}

	// Publish audit event
	if err := s.repository.PublishAuditEvent(ctx, domain.EntityTypeEvent, event.EventID, domain.AuditEventInsert, userID, nil, event); err != nil {
		// Log error but don't fail the operation
	}

	return event, nil
}

// AdminUpdateEvent updates an existing event (admin only)
func (s *EventsService) AdminUpdateEvent(ctx context.Context, eventID string, request AdminUpdateEventRequest, userID string) (*Event, error) {
	// Validate admin authentication
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required to update events")
	}

	// Get existing event
	event, err := s.repository.GetEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Store original data for audit
	originalEvent := *event

	// Update fields if provided
	if request.Title != nil {
		if err := event.UpdateEvent(request.Title, nil, nil, nil, userID); err != nil {
			return nil, err
		}
	}

	if request.Description != nil {
		if err := event.UpdateEvent(nil, request.Description, nil, nil, userID); err != nil {
			return nil, err
		}
	}

	if request.Location != nil {
		if err := event.UpdateEvent(nil, nil, request.Location, nil, userID); err != nil {
			return nil, err
		}
	}

	if request.EventDate != nil {
		eventDate, parseErr := time.Parse("2006-01-02", *request.EventDate)
		if parseErr != nil {
			return nil, domain.NewValidationError("invalid event date format, use YYYY-MM-DD")
		}
		if err := event.UpdateEvent(nil, nil, nil, &eventDate, userID); err != nil {
			return nil, err
		}
	}

	if request.Content != nil {
		event.Content = request.Content
		event.ModifiedOn = &[]time.Time{time.Now()}[0]
		event.ModifiedBy = &userID
	}

	if request.ImageURL != nil {
		if err := event.SetImageURL(*request.ImageURL, userID); err != nil {
			return nil, err
		}
	}

	if request.VirtualLink != nil {
		if err := event.SetVirtualLink(*request.VirtualLink, userID); err != nil {
			return nil, err
		}
	}

	if request.OrganizerName != nil {
		event.OrganizerName = request.OrganizerName
		event.ModifiedOn = &[]time.Time{time.Now()}[0]
		event.ModifiedBy = &userID
	}

	if request.EventTime != nil {
		if err := s.validateTimeFormat(*request.EventTime); err != nil {
			return nil, err
		}
		event.EventTime = request.EventTime
		event.ModifiedOn = &[]time.Time{time.Now()}[0]
		event.ModifiedBy = &userID
	}

	if request.EndDate != nil {
		endDate, parseErr := time.Parse("2006-01-02", *request.EndDate)
		if parseErr != nil {
			return nil, domain.NewValidationError("invalid end date format, use YYYY-MM-DD")
		}
		if endDate.Before(event.EventDate) {
			return nil, domain.NewValidationError("end date cannot be before event date")
		}
		event.EndDate = &endDate
		event.ModifiedOn = &[]time.Time{time.Now()}[0]
		event.ModifiedBy = &userID
	}

	if request.EndTime != nil {
		if err := s.validateTimeFormat(*request.EndTime); err != nil {
			return nil, err
		}
		event.EndTime = request.EndTime
		event.ModifiedOn = &[]time.Time{time.Now()}[0]
		event.ModifiedBy = &userID
	}

	if request.MaxCapacity != nil {
		if err := event.SetMaxCapacity(request.MaxCapacity, userID); err != nil {
			return nil, err
		}
	}

	if request.RegistrationDeadline != nil {
		deadline, parseErr := time.Parse(time.RFC3339, *request.RegistrationDeadline)
		if parseErr != nil {
			return nil, domain.NewValidationError("invalid registration deadline format, use YYYY-MM-DDTHH:MM:SSZ")
		}
		if err := event.SetRegistrationDeadline(&deadline, userID); err != nil {
			return nil, err
		}
	}

	if request.EventType != nil {
		eventType := EventType(*request.EventType)
		if err := event.SetEventType(eventType, userID); err != nil {
			return nil, err
		}
	}

	if request.PriorityLevel != nil {
		priorityLevel := PriorityLevel(*request.PriorityLevel)
		if err := event.SetPriorityLevel(priorityLevel, userID); err != nil {
			return nil, err
		}
	}

	if len(request.Tags) > 0 {
		if err := event.AddTags(request.Tags, userID); err != nil {
			return nil, err
		}
	}

	if request.CategoryID != nil {
		// Verify new category exists
		_, categoryErr := s.repository.GetEventCategory(ctx, *request.CategoryID)
		if categoryErr != nil {
			if domain.IsNotFoundError(categoryErr) {
				return nil, domain.NewNotFoundError("event category", *request.CategoryID)
			}
			return nil, domain.WrapError(categoryErr, "failed to validate category")
		}
		event.CategoryID = *request.CategoryID
		event.ModifiedOn = &[]time.Time{time.Now()}[0]
		event.ModifiedBy = &userID
	}

	// Save updated event
	if err := s.repository.SaveEvent(ctx, event); err != nil {
		return nil, domain.WrapError(err, "failed to save updated event")
	}

	// Publish audit event
	if err := s.repository.PublishAuditEvent(ctx, domain.EntityTypeEvent, event.EventID, domain.AuditEventUpdate, userID, &originalEvent, event); err != nil {
		// Log error but don't fail the operation
	}

	return event, nil
}

// AdminDeleteEvent soft deletes an event (admin only)
func (s *EventsService) AdminDeleteEvent(ctx context.Context, eventID string, userID string) error {
	// Validate admin authentication
	if !IsAdminUser(userID) {
		return domain.NewUnauthorizedError("admin privileges required to delete events")
	}

	// Get existing event for audit
	event, err := s.repository.GetEvent(ctx, eventID)
	if err != nil {
		return err
	}

	// Store original data for audit
	originalEvent := *event

	// Delete event
	if err := s.repository.DeleteEvent(ctx, eventID, userID); err != nil {
		return domain.WrapError(err, "failed to delete event")
	}

	// Remove as featured event if it was featured
	featuredEvent, featuredErr := s.repository.GetFeaturedEvent(ctx)
	if featuredErr == nil && featuredEvent.EventID == eventID {
		if err := s.repository.DeleteFeaturedEvent(ctx); err != nil {
			// Log error but don't fail the operation
		}
	}

	// Publish audit event
	if err := s.repository.PublishAuditEvent(ctx, domain.EntityTypeEvent, event.EventID, domain.AuditEventDelete, userID, &originalEvent, nil); err != nil {
		// Log error but don't fail the operation
	}

	return nil
}

// AdminPublishEvent publishes a draft event (admin only)
func (s *EventsService) AdminPublishEvent(ctx context.Context, eventID string, userID string) (*Event, error) {
	// Validate admin authentication
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required to publish events")
	}

	// Get existing event
	event, err := s.repository.GetEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Store original data for audit
	originalEvent := *event

	// Publish event
	if err := event.PublishEvent(userID); err != nil {
		return nil, err
	}

	// Save updated event
	if err := s.repository.SaveEvent(ctx, event); err != nil {
		return nil, domain.WrapError(err, "failed to save published event")
	}

	// Publish audit event
	if err := s.repository.PublishAuditEvent(ctx, domain.EntityTypeEvent, event.EventID, domain.AuditEventUpdate, userID, &originalEvent, event); err != nil {
		// Log error but don't fail the operation
	}

	return event, nil
}

// AdminArchiveEvent archives a published event (admin only)
func (s *EventsService) AdminArchiveEvent(ctx context.Context, eventID string, userID string) (*Event, error) {
	// Validate admin authentication
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required to archive events")
	}

	// Get existing event
	event, err := s.repository.GetEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Store original data for audit
	originalEvent := *event

	// Archive event
	if err := event.ArchiveEvent(userID); err != nil {
		return nil, err
	}

	// Save updated event
	if err := s.repository.SaveEvent(ctx, event); err != nil {
		return nil, domain.WrapError(err, "failed to save archived event")
	}

	// Remove as featured event if it was featured
	featuredEvent, featuredErr := s.repository.GetFeaturedEvent(ctx)
	if featuredErr == nil && featuredEvent.EventID == eventID {
		if err := s.repository.DeleteFeaturedEvent(ctx); err != nil {
			// Log error but don't fail the operation
		}
	}

	// Publish audit event
	if err := s.repository.PublishAuditEvent(ctx, domain.EntityTypeEvent, event.EventID, domain.AuditEventUpdate, userID, &originalEvent, event); err != nil {
		// Log error but don't fail the operation
	}

	return event, nil
}

// AdminCreateEventCategory creates a new event category (admin only)
func (s *EventsService) AdminCreateEventCategory(ctx context.Context, request AdminCreateEventCategoryRequest, userID string) (*EventCategory, error) {
	// Validate admin authentication
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required to create event categories")
	}

	// Validate request
	if err := s.validateCreateEventCategoryRequest(request); err != nil {
		return nil, err
	}

	// Create category entity
	category, err := NewEventCategory(request.Name, request.Description, userID)
	if err != nil {
		return nil, domain.WrapError(err, "failed to create event category entity")
	}

	// Save category to repository
	if err := s.repository.SaveEventCategory(ctx, category); err != nil {
		return nil, domain.WrapError(err, "failed to save event category")
	}

	// Publish audit event
	if err := s.repository.PublishAuditEvent(ctx, domain.EntityTypeEventCategory, category.CategoryID, domain.AuditEventInsert, userID, nil, category); err != nil {
		// Log error but don't fail the operation
	}

	return category, nil
}

// AdminDeleteEventCategory soft deletes an event category and reassigns events (admin only)
func (s *EventsService) AdminDeleteEventCategory(ctx context.Context, categoryID string, userID string) error {
	// Validate admin authentication
	if !IsAdminUser(userID) {
		return domain.NewUnauthorizedError("admin privileges required to delete event categories")
	}

	// Get existing category
	category, err := s.repository.GetEventCategory(ctx, categoryID)
	if err != nil {
		return err
	}

	// Cannot delete default unassigned category
	if category.IsDefaultUnassigned {
		return domain.NewValidationError("cannot delete default unassigned category")
	}

	// Store original data for audit
	originalCategory := *category

	// Get default unassigned category for reassignment
	_, err = s.repository.GetDefaultUnassignedCategory(ctx)
	if err != nil {
		return domain.WrapError(err, "failed to get default unassigned category")
	}

	// TODO: Reassign all events in this category to default unassigned category
	// This would require additional repository methods to list and update events by category

	// Delete category
	if err := s.repository.DeleteEventCategory(ctx, categoryID, userID); err != nil {
		return domain.WrapError(err, "failed to delete event category")
	}

	// Publish audit event
	if err := s.repository.PublishAuditEvent(ctx, domain.EntityTypeEventCategory, category.CategoryID, domain.AuditEventDelete, userID, &originalCategory, nil); err != nil {
		// Log error but don't fail the operation
	}

	return nil
}

// AdminSetFeaturedEvent sets an event as the featured event (admin only)
func (s *EventsService) AdminSetFeaturedEvent(ctx context.Context, eventID string, userID string) (*FeaturedEvent, error) {
	// Validate admin authentication
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required to set featured events")
	}

	// Get event and validate it can be featured
	event, err := s.repository.GetEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Check if event can be featured
	if err := event.CanBeFeatured(); err != nil {
		return nil, err
	}

	// Check if event is in default unassigned category
	category, err := s.repository.GetEventCategory(ctx, event.CategoryID)
	if err != nil {
		return nil, domain.WrapError(err, "failed to get event category")
	}

	if category.IsDefaultUnassigned {
		return nil, domain.NewValidationError("cannot feature event from default unassigned category")
	}

	// Remove existing featured event if any
	if err := s.repository.DeleteFeaturedEvent(ctx); err != nil {
		// Ignore not found errors for featured event
		if !domain.IsNotFoundError(err) {
			return nil, domain.WrapError(err, "failed to remove existing featured event")
		}
	}

	// Create new featured event
	featuredEvent, err := NewFeaturedEvent(eventID, userID)
	if err != nil {
		return nil, domain.WrapError(err, "failed to create featured event entity")
	}

	// Save featured event
	if err := s.repository.SaveFeaturedEvent(ctx, featuredEvent); err != nil {
		return nil, domain.WrapError(err, "failed to save featured event")
	}

	// Publish audit event
	if err := s.repository.PublishAuditEvent(ctx, domain.EntityTypeFeaturedEvent, featuredEvent.FeaturedEventID, domain.AuditEventInsert, userID, nil, featuredEvent); err != nil {
		// Log error but don't fail the operation
	}

	return featuredEvent, nil
}

// AdminGetEventRegistrations gets all registrations for an event (admin only)
func (s *EventsService) AdminGetEventRegistrations(ctx context.Context, eventID string, userID string) ([]*EventRegistration, error) {
	// Validate admin authentication
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required to view event registrations")
	}

	// Verify event exists
	_, err := s.repository.GetEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Get event registrations
	registrations, err := s.repository.GetEventRegistrations(ctx, eventID)
	if err != nil {
		return nil, domain.WrapError(err, "failed to get event registrations")
	}

	return registrations, nil
}

// Private validation helper methods

func (s *EventsService) validateCreateEventRequest(request AdminCreateEventRequest) error {
	if strings.TrimSpace(request.Title) == "" {
		return domain.NewValidationError("event title cannot be empty")
	}

	if strings.TrimSpace(request.Description) == "" {
		return domain.NewValidationError("event description cannot be empty")
	}

	if strings.TrimSpace(request.CategoryID) == "" {
		return domain.NewValidationError("category ID cannot be empty")
	}

	if strings.TrimSpace(request.Location) == "" {
		return domain.NewValidationError("event location cannot be empty")
	}

	if strings.TrimSpace(request.EventDate) == "" {
		return domain.NewValidationError("event date cannot be empty")
	}

	if strings.TrimSpace(request.EventType) == "" {
		return domain.NewValidationError("event type cannot be empty")
	}

	return nil
}

func (s *EventsService) validateCreateEventCategoryRequest(request AdminCreateEventCategoryRequest) error {
	if strings.TrimSpace(request.Name) == "" {
		return domain.NewValidationError("category name cannot be empty")
	}

	if strings.TrimSpace(request.Description) == "" {
		return domain.NewValidationError("category description cannot be empty")
	}

	return nil
}

func (s *EventsService) validateTimeFormat(timeStr string) error {
	// Expected format: HH:MM
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return domain.NewValidationError("time must be in HH:MM format")
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return domain.NewValidationError("invalid hour, must be 00-23")
	}

	minute, err := strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return domain.NewValidationError("invalid minute, must be 00-59")
	}

	return nil
}