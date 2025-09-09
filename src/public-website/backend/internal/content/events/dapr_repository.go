package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/google/uuid"
)

// EventsRepository implements events data access using Dapr state store and bindings
type EventsRepository struct {
	stateStore *dapr.StateStore
	bindings   *dapr.Bindings
	pubsub     *dapr.PubSub
}

// NewEventsRepository creates a new events repository
func NewEventsRepository(client *dapr.Client) *EventsRepository {
	return &EventsRepository{
		stateStore: dapr.NewStateStore(client),
		bindings:   dapr.NewBindings(client),
		pubsub:     dapr.NewPubSub(client),
	}
}

// Event operations

// SaveEvent saves event to Dapr state store
func (r *EventsRepository) SaveEvent(ctx context.Context, event *Event) error {
	key := r.stateStore.CreateKey("events", "event", event.EventID)
	
	err := r.stateStore.Save(ctx, key, event, nil)
	if err != nil {
		return fmt.Errorf("failed to save event %s: %w", event.EventID, err)
	}

	// Create index for slug search
	slugKey := r.stateStore.CreateIndexKey("events", "event", "slug", event.Slug)
	slugIndex := map[string]string{"event_id": event.EventID}
	
	err = r.stateStore.Save(ctx, slugKey, slugIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create slug index for event %s: %w", event.EventID, err)
	}

	// Create index for category search
	categoryKey := r.stateStore.CreateIndexKey("events", "event", "category", event.CategoryID)
	categoryIndex := map[string]string{"event_id": event.EventID}
	
	err = r.stateStore.Save(ctx, categoryKey, categoryIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create category index for event %s: %w", event.EventID, err)
	}

	// Create index for event type search
	typeKey := r.stateStore.CreateIndexKey("events", "event", "type", string(event.EventType))
	typeIndex := map[string]string{"event_id": event.EventID}
	
	err = r.stateStore.Save(ctx, typeKey, typeIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create type index for event %s: %w", event.EventID, err)
	}

	// Create index for publishing status search
	statusKey := r.stateStore.CreateIndexKey("events", "event", "status", string(event.PublishingStatus))
	statusIndex := map[string]string{"event_id": event.EventID}
	
	err = r.stateStore.Save(ctx, statusKey, statusIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create status index for event %s: %w", event.EventID, err)
	}

	// Create index for event date search
	dateKey := r.stateStore.CreateIndexKey("events", "event", "date", event.EventDate.Format("2006-01-02"))
	dateIndex := map[string]string{"event_id": event.EventID}
	
	err = r.stateStore.Save(ctx, dateKey, dateIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create date index for event %s: %w", event.EventID, err)
	}

	return nil
}

// GetEvent retrieves event from Dapr state store
func (r *EventsRepository) GetEvent(ctx context.Context, eventID string) (*Event, error) {
	key := r.stateStore.CreateKey("events", "event", eventID)
	
	var event Event
	found, err := r.stateStore.Get(ctx, key, &event)
	if err != nil {
		return nil, fmt.Errorf("failed to get event %s: %w", eventID, err)
	}
	
	if !found {
		return nil, domain.NewNotFoundError("event", eventID)
	}

	return &event, nil
}

// DeleteEvent soft deletes event in Dapr state store
func (r *EventsRepository) DeleteEvent(ctx context.Context, eventID string, userID string) error {
	// Get existing event
	event, err := r.GetEvent(ctx, eventID)
	if err != nil {
		return err
	}

	// Mark as deleted
	event.IsDeleted = true
	event.DeletedOn = &[]time.Time{time.Now()}[0]
	event.DeletedBy = &userID

	// Save updated event
	return r.SaveEvent(ctx, event)
}

// Event category operations

// SaveEventCategory saves event category to Dapr state store
func (r *EventsRepository) SaveEventCategory(ctx context.Context, category *EventCategory) error {
	key := r.stateStore.CreateKey("events", "category", category.CategoryID)
	
	err := r.stateStore.Save(ctx, key, category, nil)
	if err != nil {
		return fmt.Errorf("failed to save event category %s: %w", category.CategoryID, err)
	}

	// Create index for slug search
	slugKey := r.stateStore.CreateIndexKey("events", "category", "slug", category.Slug)
	slugIndex := map[string]string{"category_id": category.CategoryID}
	
	err = r.stateStore.Save(ctx, slugKey, slugIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create slug index for event category %s: %w", category.CategoryID, err)
	}

	// Create index for default unassigned category
	if category.IsDefaultUnassigned {
		defaultKey := r.stateStore.CreateIndexKey("events", "category", "default", "true")
		defaultIndex := map[string]string{"category_id": category.CategoryID}
		
		err = r.stateStore.Save(ctx, defaultKey, defaultIndex, nil)
		if err != nil {
			return fmt.Errorf("failed to create default index for event category %s: %w", category.CategoryID, err)
		}
	}

	return nil
}

// GetEventCategory retrieves event category from Dapr state store
func (r *EventsRepository) GetEventCategory(ctx context.Context, categoryID string) (*EventCategory, error) {
	key := r.stateStore.CreateKey("events", "category", categoryID)
	
	var category EventCategory
	found, err := r.stateStore.Get(ctx, key, &category)
	if err != nil {
		return nil, fmt.Errorf("failed to get event category %s: %w", categoryID, err)
	}
	
	if !found {
		return nil, domain.NewNotFoundError("event category", categoryID)
	}

	return &category, nil
}

// DeleteEventCategory soft deletes event category in Dapr state store
func (r *EventsRepository) DeleteEventCategory(ctx context.Context, categoryID string, userID string) error {
	// Get existing category
	category, err := r.GetEventCategory(ctx, categoryID)
	if err != nil {
		return err
	}

	// Cannot delete default unassigned category
	if category.IsDefaultUnassigned {
		return domain.NewValidationError("cannot delete default unassigned category")
	}

	// Mark as deleted
	category.IsDeleted = true
	category.DeletedOn = &[]time.Time{time.Now()}[0]
	category.DeletedBy = &userID

	// Save updated category
	return r.SaveEventCategory(ctx, category)
}

// GetDefaultUnassignedCategory retrieves the default unassigned category
func (r *EventsRepository) GetDefaultUnassignedCategory(ctx context.Context) (*EventCategory, error) {
	// Use index to find default unassigned category
	defaultKey := r.stateStore.CreateIndexKey("events", "category", "default", "true")
	
	var defaultIndex map[string]string
	found, err := r.stateStore.Get(ctx, defaultKey, &defaultIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get default unassigned category index: %w", err)
	}
	
	if !found {
		return nil, domain.NewNotFoundError("default unassigned category", "unassigned")
	}

	categoryID, exists := defaultIndex["category_id"]
	if !exists {
		return nil, domain.NewInternalError("invalid default category index", nil)
	}

	return r.GetEventCategory(ctx, categoryID)
}

// Featured event operations

// SaveFeaturedEvent saves featured event to Dapr state store
func (r *EventsRepository) SaveFeaturedEvent(ctx context.Context, featuredEvent *FeaturedEvent) error {
	key := r.stateStore.CreateKey("events", "featured", "current")
	
	err := r.stateStore.Save(ctx, key, featuredEvent, nil)
	if err != nil {
		return fmt.Errorf("failed to save featured event %s: %w", featuredEvent.FeaturedEventID, err)
	}

	// Create index for event ID
	eventKey := r.stateStore.CreateIndexKey("events", "featured", "event", featuredEvent.EventID)
	eventIndex := map[string]string{"featured_event_id": featuredEvent.FeaturedEventID}
	
	err = r.stateStore.Save(ctx, eventKey, eventIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create event index for featured event %s: %w", featuredEvent.FeaturedEventID, err)
	}

	return nil
}

// GetFeaturedEvent retrieves the current featured event from Dapr state store
func (r *EventsRepository) GetFeaturedEvent(ctx context.Context) (*FeaturedEvent, error) {
	key := r.stateStore.CreateKey("events", "featured", "current")
	
	var featuredEvent FeaturedEvent
	found, err := r.stateStore.Get(ctx, key, &featuredEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to get featured event: %w", err)
	}
	
	if !found {
		return nil, domain.NewNotFoundError("featured event", "current")
	}

	return &featuredEvent, nil
}

// DeleteFeaturedEvent removes the current featured event from Dapr state store
func (r *EventsRepository) DeleteFeaturedEvent(ctx context.Context) error {
	key := r.stateStore.CreateKey("events", "featured", "current")
	
	err := r.stateStore.Delete(ctx, key, nil)
	if err != nil {
		return fmt.Errorf("failed to delete featured event: %w", err)
	}

	return nil
}

// Event registration operations

// GetEventRegistrations retrieves all registrations for an event
func (r *EventsRepository) GetEventRegistrations(ctx context.Context, eventID string) ([]*EventRegistration, error) {
	// Use query to find all registrations for this event
	query := map[string]interface{}{
		"filter": map[string]interface{}{
			"EQ": map[string]interface{}{
				"event_id":   eventID,
				"is_deleted": false,
			},
		},
	}

	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal registration query: %w", err)
	}

	results, err := r.stateStore.Query(ctx, string(queryBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to query event registrations: %w", err)
	}

	var registrations []*EventRegistration
	for _, result := range results {
		var registration EventRegistration
		if err := json.Unmarshal(result.Value, &registration); err != nil {
			return nil, fmt.Errorf("failed to unmarshal registration: %w", err)
		}
		registrations = append(registrations, &registration)
	}

	return registrations, nil
}

// Audit operations

// PublishAuditEvent publishes audit event to Grafana Cloud Loki via Dapr pub/sub
func (r *EventsRepository) PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error {
	correlationID := domain.GetCorrelationID(ctx)
	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	auditEvent := domain.AuditEvent{
		AuditID:       uuid.New().String(), // Generate unique audit ID
		EntityType:    entityType,
		EntityID:      entityID,
		OperationType: operationType,
		AuditTime:     time.Now(),
		UserID:        userID,
		CorrelationID: correlationID,
		TraceID:       domain.GetTraceID(ctx),
		DataSnapshot: &domain.AuditDataSnapshot{
			Before: beforeData,
			After:  afterData,
		},
		Environment: "development", // This should come from configuration
	}

	// Publish to Grafana audit events topic
	daprAuditEvent := &dapr.AuditEvent{
		AuditID:       auditEvent.AuditID,
		EntityType:    string(auditEvent.EntityType),
		EntityID:      auditEvent.EntityID,
		OperationType: string(auditEvent.OperationType),
		AuditTime:     auditEvent.AuditTime,
		UserID:        auditEvent.UserID,
		CorrelationID: auditEvent.CorrelationID,
		TraceID:       auditEvent.TraceID,
		DataSnapshot:  map[string]interface{}{
			"before": auditEvent.DataSnapshot.Before,
			"after":  auditEvent.DataSnapshot.After,
		},
		Environment:   auditEvent.Environment,
	}
	err := r.pubsub.PublishAuditEvent(ctx, daprAuditEvent)
	if err != nil {
		return fmt.Errorf("failed to publish audit event for %s %s: %w", entityType, entityID, err)
	}

	return nil
}

// Search and query operations

// SearchEvents searches for events by various criteria
func (r *EventsRepository) SearchEvents(ctx context.Context, criteria EventSearchCriteria) ([]*Event, error) {
	// Build query based on criteria
	filters := make(map[string]interface{})

	if criteria.Title != "" {
		filters["title"] = map[string]interface{}{
			"LIKE": fmt.Sprintf("%%%s%%", criteria.Title),
		}
	}

	if criteria.CategoryID != "" {
		filters["category_id"] = criteria.CategoryID
	}

	if criteria.EventType != "" {
		filters["event_type"] = criteria.EventType
	}

	if criteria.PublishingStatus != "" {
		filters["publishing_status"] = criteria.PublishingStatus
	}

	if criteria.EventDateFrom != nil {
		filters["event_date"] = map[string]interface{}{
			"GTE": criteria.EventDateFrom.Format(time.RFC3339),
		}
	}

	if criteria.EventDateTo != nil {
		if existingDateFilter, exists := filters["event_date"]; exists {
			if dateMap, ok := existingDateFilter.(map[string]interface{}); ok {
				dateMap["LTE"] = criteria.EventDateTo.Format(time.RFC3339)
			}
		} else {
			filters["event_date"] = map[string]interface{}{
				"LTE": criteria.EventDateTo.Format(time.RFC3339),
			}
		}
	}

	// Always filter out deleted events unless specifically requested
	if !criteria.IncludeDeleted {
		filters["is_deleted"] = false
	}

	query := map[string]interface{}{
		"filter": map[string]interface{}{
			"EQ": filters,
		},
	}

	if criteria.Limit > 0 {
		query["page"] = map[string]interface{}{
			"limit": criteria.Limit,
		}
		if criteria.Offset > 0 {
			query["page"].(map[string]interface{})["token"] = fmt.Sprintf("%d", criteria.Offset)
		}
	}

	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event search query: %w", err)
	}

	results, err := r.stateStore.Query(ctx, string(queryBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to search events: %w", err)
	}

	var events []*Event
	for _, result := range results {
		var event Event
		if err := json.Unmarshal(result.Value, &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event: %w", err)
		}
		events = append(events, &event)
	}

	return events, nil
}

// GetEventsByCategory retrieves all events in a specific category
func (r *EventsRepository) GetEventsByCategory(ctx context.Context, categoryID string, includeDeleted bool) ([]*Event, error) {
	criteria := EventSearchCriteria{
		CategoryID:      categoryID,
		IncludeDeleted:  includeDeleted,
	}
	
	return r.SearchEvents(ctx, criteria)
}

// GetPublishedEvents retrieves all published events
func (r *EventsRepository) GetPublishedEvents(ctx context.Context) ([]*Event, error) {
	criteria := EventSearchCriteria{
		PublishingStatus: string(PublishingStatusPublished),
		IncludeDeleted:   false,
	}
	
	return r.SearchEvents(ctx, criteria)
}

// GetUpcomingEvents retrieves all upcoming events
func (r *EventsRepository) GetUpcomingEvents(ctx context.Context) ([]*Event, error) {
	now := time.Now()
	criteria := EventSearchCriteria{
		PublishingStatus: string(PublishingStatusPublished),
		EventDateFrom:    &now,
		IncludeDeleted:   false,
	}
	
	return r.SearchEvents(ctx, criteria)
}

// Supporting types

// EventSearchCriteria defines criteria for searching events
type EventSearchCriteria struct {
	Title            string
	CategoryID       string
	EventType        string
	PublishingStatus string
	EventDateFrom    *time.Time
	EventDateTo      *time.Time
	IncludeDeleted   bool
	Limit            int
	Offset           int
}