package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/google/uuid"
)

// DaprSubscriberRepository implements SubscriberRepository using Dapr state store
type DaprSubscriberRepository struct {
	stateStore *dapr.StateStore
	bindings   *dapr.Bindings
	pubsub     *dapr.PubSub
	logger     *slog.Logger
}

// NewDaprSubscriberRepository creates a new Dapr-based subscriber repository
func NewDaprSubscriberRepository(client *dapr.Client, logger *slog.Logger) SubscriberRepository {
	return &DaprSubscriberRepository{
		stateStore: dapr.NewStateStore(client),
		bindings:   dapr.NewBindings(client),
		pubsub:     dapr.NewPubSub(client),
		logger:     logger,
	}
}

// CreateSubscriber creates a new notification subscriber in Dapr state store
func (r *DaprSubscriberRepository) CreateSubscriber(ctx context.Context, subscriber *NotificationSubscriber) error {
	// Check if email already exists
	exists, err := r.CheckEmailExists(ctx, subscriber.Email, nil)
	if err != nil {
		return err
	}
	if exists {
		return domain.NewValidationError("email address already exists")
	}

	// Save subscriber to state store
	key := r.stateStore.CreateKey("notifications", "subscriber", subscriber.SubscriberID)
	err = r.stateStore.Save(ctx, key, subscriber, nil)
	if err != nil {
		r.logger.Error("Failed to create subscriber",
			"subscriber_id", subscriber.SubscriberID,
			"email", subscriber.Email,
			"error", err)
		return fmt.Errorf("failed to save subscriber %s: %w", subscriber.SubscriberID, err)
	}

	// Create email index for uniqueness checking
	emailKey := r.stateStore.CreateIndexKey("notifications", "subscriber", "email", subscriber.Email)
	emailIndex := map[string]string{"subscriber_id": subscriber.SubscriberID}
	err = r.stateStore.Save(ctx, emailKey, emailIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create email index for subscriber %s: %w", subscriber.SubscriberID, err)
	}

	// Create status index
	statusKey := r.stateStore.CreateIndexKey("notifications", "subscriber", "status", string(subscriber.Status))
	statusIndex := map[string]string{"subscriber_id": subscriber.SubscriberID}
	err = r.stateStore.Save(ctx, statusKey, statusIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create status index for subscriber %s: %w", subscriber.SubscriberID, err)
	}

	// Create event type indexes
	for _, eventType := range subscriber.EventTypes {
		eventTypeKey := r.stateStore.CreateIndexKey("notifications", "subscriber", "event_type", string(eventType))
		eventTypeIndex := map[string]string{"subscriber_id": subscriber.SubscriberID}
		err = r.stateStore.Save(ctx, eventTypeKey, eventTypeIndex, nil)
		if err != nil {
			return fmt.Errorf("failed to create event type index for subscriber %s: %w", subscriber.SubscriberID, err)
		}
	}

	// Create priority index
	priorityKey := r.stateStore.CreateIndexKey("notifications", "subscriber", "priority", string(subscriber.PriorityThreshold))
	priorityIndex := map[string]string{"subscriber_id": subscriber.SubscriberID}
	err = r.stateStore.Save(ctx, priorityKey, priorityIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create priority index for subscriber %s: %w", subscriber.SubscriberID, err)
	}

	r.logger.Info("Created new subscriber",
		"subscriber_id", subscriber.SubscriberID,
		"email", subscriber.Email)

	return nil
}

// GetSubscriber retrieves a subscriber by ID from Dapr state store
func (r *DaprSubscriberRepository) GetSubscriber(ctx context.Context, subscriberID string) (*NotificationSubscriber, error) {
	// Validate UUID format
	if _, err := uuid.Parse(subscriberID); err != nil {
		return nil, domain.NewValidationError("invalid subscriber ID format")
	}

	key := r.stateStore.CreateKey("notifications", "subscriber", subscriberID)
	
	var subscriber NotificationSubscriber
	found, err := r.stateStore.Get(ctx, key, &subscriber)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriber %s: %w", subscriberID, err)
	}
	
	if !found || subscriber.IsDeleted {
		return nil, domain.NewNotFoundError("subscriber", subscriberID)
	}
	
	return &subscriber, nil
}

// GetSubscriberByEmail retrieves a subscriber by email address from Dapr state store
func (r *DaprSubscriberRepository) GetSubscriberByEmail(ctx context.Context, email string) (*NotificationSubscriber, error) {
	// Use email index to find subscriber ID
	emailKey := r.stateStore.CreateIndexKey("notifications", "subscriber", "email", email)
	
	var emailIndex map[string]string
	found, err := r.stateStore.Get(ctx, emailKey, &emailIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to query subscriber by email %s: %w", email, err)
	}
	
	if !found {
		return nil, domain.NewNotFoundError("subscriber", email)
	}
	
	subscriberID := emailIndex["subscriber_id"]
	if subscriberID == "" {
		return nil, domain.NewNotFoundError("subscriber", email)
	}
	
	// Get the actual subscriber
	return r.GetSubscriber(ctx, subscriberID)
}

// UpdateSubscriber updates an existing subscriber in Dapr state store
func (r *DaprSubscriberRepository) UpdateSubscriber(ctx context.Context, subscriber *NotificationSubscriber) error {
	// Get existing subscriber to check if it exists
	existing, err := r.GetSubscriber(ctx, subscriber.SubscriberID)
	if err != nil {
		return err
	}

	// Check if email change conflicts with another subscriber
	if existing.Email != subscriber.Email {
		exists, err := r.CheckEmailExists(ctx, subscriber.Email, &subscriber.SubscriberID)
		if err != nil {
			return err
		}
		if exists {
			return domain.NewValidationError("email address already exists")
		}

		// Remove old email index
		oldEmailKey := r.stateStore.CreateIndexKey("notifications", "subscriber", "email", existing.Email)
		err = r.stateStore.Delete(ctx, oldEmailKey, nil)
		if err != nil {
			return fmt.Errorf("failed to remove old email index for subscriber %s: %w", subscriber.SubscriberID, err)
		}

		// Create new email index
		newEmailKey := r.stateStore.CreateIndexKey("notifications", "subscriber", "email", subscriber.Email)
		emailIndex := map[string]string{"subscriber_id": subscriber.SubscriberID}
		err = r.stateStore.Save(ctx, newEmailKey, emailIndex, nil)
		if err != nil {
			return fmt.Errorf("failed to create new email index for subscriber %s: %w", subscriber.SubscriberID, err)
		}
	}

	// Update status index if changed
	if existing.Status != subscriber.Status {
		// Remove old status index
		oldStatusKey := r.stateStore.CreateIndexKey("notifications", "subscriber", "status", string(existing.Status))
		err = r.stateStore.Delete(ctx, oldStatusKey, nil)
		if err != nil {
			return fmt.Errorf("failed to remove old status index for subscriber %s: %w", subscriber.SubscriberID, err)
		}

		// Create new status index
		newStatusKey := r.stateStore.CreateIndexKey("notifications", "subscriber", "status", string(subscriber.Status))
		statusIndex := map[string]string{"subscriber_id": subscriber.SubscriberID}
		err = r.stateStore.Save(ctx, newStatusKey, statusIndex, nil)
		if err != nil {
			return fmt.Errorf("failed to create new status index for subscriber %s: %w", subscriber.SubscriberID, err)
		}
	}

	// Update priority index if changed
	if existing.PriorityThreshold != subscriber.PriorityThreshold {
		// Remove old priority index
		oldPriorityKey := r.stateStore.CreateIndexKey("notifications", "subscriber", "priority", string(existing.PriorityThreshold))
		err = r.stateStore.Delete(ctx, oldPriorityKey, nil)
		if err != nil {
			return fmt.Errorf("failed to remove old priority index for subscriber %s: %w", subscriber.SubscriberID, err)
		}

		// Create new priority index
		newPriorityKey := r.stateStore.CreateIndexKey("notifications", "subscriber", "priority", string(subscriber.PriorityThreshold))
		priorityIndex := map[string]string{"subscriber_id": subscriber.SubscriberID}
		err = r.stateStore.Save(ctx, newPriorityKey, priorityIndex, nil)
		if err != nil {
			return fmt.Errorf("failed to create new priority index for subscriber %s: %w", subscriber.SubscriberID, err)
		}
	}

	// Update the subscriber
	key := r.stateStore.CreateKey("notifications", "subscriber", subscriber.SubscriberID)
	err = r.stateStore.Save(ctx, key, subscriber, nil)
	if err != nil {
		return fmt.Errorf("failed to update subscriber %s: %w", subscriber.SubscriberID, err)
	}

	r.logger.Info("Updated subscriber",
		"subscriber_id", subscriber.SubscriberID,
		"email", subscriber.Email)

	return nil
}

// DeleteSubscriber soft deletes a subscriber
func (r *DaprSubscriberRepository) DeleteSubscriber(ctx context.Context, subscriberID string, deletedBy string) error {
	subscriber, err := r.GetSubscriber(ctx, subscriberID)
	if err != nil {
		return err
	}

	subscriber.IsDeleted = true
	now := time.Now().UTC()
	subscriber.DeletedAt = &now
	subscriber.UpdatedBy = deletedBy
	subscriber.UpdatedAt = now

	// Save updated subscriber
	key := r.stateStore.CreateKey("notifications", "subscriber", subscriberID)
	err = r.stateStore.Save(ctx, key, subscriber, nil)
	if err != nil {
		return fmt.Errorf("failed to delete subscriber %s: %w", subscriberID, err)
	}

	r.logger.Info("Deleted subscriber", "subscriber_id", subscriberID)
	return nil
}

// ListSubscribers retrieves a paginated list of subscribers
func (r *DaprSubscriberRepository) ListSubscribers(ctx context.Context, status *SubscriberStatus, limit, offset int) ([]*NotificationSubscriber, int, error) {
	if limit <= 0 {
		return nil, 0, domain.NewValidationError("invalid limit parameter")
	}

	if offset < 0 {
		return nil, 0, domain.NewValidationError("invalid offset parameter")
	}

	// Build query for Dapr state store
	var query string
	if status != nil {
		query = fmt.Sprintf(`{
			"filter": {
				"AND": [
					{"EQ": {"status": "%s"}},
					{"EQ": {"is_deleted": false}}
				]
			},
			"sort": [
				{
					"key": "created_at",
					"order": "DESC"
				}
			]
		}`, *status)
	} else {
		query = `{
			"filter": {
				"EQ": {"is_deleted": false}
			},
			"sort": [
				{
					"key": "created_at",
					"order": "DESC"
				}
			]
		}`
	}

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query subscribers: %w", err)
	}

	var allSubscribers []*NotificationSubscriber
	for _, result := range results {
		var subscriber NotificationSubscriber
		err = json.Unmarshal(result.Value, &subscriber)
		if err != nil {
			continue
		}
		if !subscriber.IsDeleted {
			allSubscribers = append(allSubscribers, &subscriber)
		}
	}

	total := len(allSubscribers)

	// Apply pagination
	start := offset
	if start >= total {
		return []*NotificationSubscriber{}, total, nil
	}

	end := start + limit
	if end > total {
		end = total
	}

	subscribers := allSubscribers[start:end]
	return subscribers, total, nil
}

// GetSubscribersByEventType retrieves active subscribers for a specific event type
func (r *DaprSubscriberRepository) GetSubscribersByEventType(ctx context.Context, eventType EventType) ([]*NotificationSubscriber, error) {
	// Use event type index to find subscriber IDs
	eventTypeKey := r.stateStore.CreateIndexKey("notifications", "subscriber", "event_type", string(eventType))
	
	var eventTypeIndex map[string]string
	found, err := r.stateStore.Get(ctx, eventTypeKey, &eventTypeIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to query subscribers by event type %s: %w", eventType, err)
	}
	
	if !found {
		return []*NotificationSubscriber{}, nil
	}

	subscriberID := eventTypeIndex["subscriber_id"]
	if subscriberID == "" {
		return []*NotificationSubscriber{}, nil
	}

	// Get the actual subscriber
	subscriber, err := r.GetSubscriber(ctx, subscriberID)
	if err != nil {
		if domain.IsNotFoundError(err) {
			return []*NotificationSubscriber{}, nil
		}
		return nil, err
	}

	// Only return active subscribers
	if subscriber.Status == SubscriberStatusActive {
		return []*NotificationSubscriber{subscriber}, nil
	}

	return []*NotificationSubscriber{}, nil
}

// GetActiveSubscribersByPriority retrieves active subscribers by priority threshold
func (r *DaprSubscriberRepository) GetActiveSubscribersByPriority(ctx context.Context, priority PriorityThreshold) ([]*NotificationSubscriber, error) {
	// Query for active subscribers
	query := `{
		"filter": {
			"AND": [
				{"EQ": {"status": "active"}},
				{"EQ": {"is_deleted": false}}
			]
		},
		"sort": [
			{
				"key": "created_at",
				"order": "ASC"
			}
		]
	}`

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active subscribers: %w", err)
	}

	// Priority levels: low=1, medium=2, high=3, urgent=4
	priorityLevels := map[PriorityThreshold]int{
		PriorityLow:    1,
		PriorityMedium: 2,
		PriorityHigh:   3,
		PriorityUrgent: 4,
	}

	eventLevel, exists := priorityLevels[priority]
	if !exists {
		return nil, domain.NewValidationError("invalid priority threshold")
	}

	var subscribers []*NotificationSubscriber
	for _, result := range results {
		var subscriber NotificationSubscriber
		err = json.Unmarshal(result.Value, &subscriber)
		if err != nil {
			continue
		}
		
		if !subscriber.IsDeleted && subscriber.Status == SubscriberStatusActive {
			subscriberLevel := priorityLevels[subscriber.PriorityThreshold]
			if subscriberLevel <= eventLevel {
				subscribers = append(subscribers, &subscriber)
			}
		}
	}

	return subscribers, nil
}

// CheckEmailExists checks if an email address already exists
func (r *DaprSubscriberRepository) CheckEmailExists(ctx context.Context, email string, excludeID *string) (bool, error) {
	emailKey := r.stateStore.CreateIndexKey("notifications", "subscriber", "email", email)
	
	var emailIndex map[string]string
	found, err := r.stateStore.Get(ctx, emailKey, &emailIndex)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}
	
	if !found {
		return false, nil
	}

	subscriberID := emailIndex["subscriber_id"]
	if subscriberID == "" {
		return false, nil
	}

	// If we're excluding a specific ID, check if it matches
	if excludeID != nil && subscriberID == *excludeID {
		return false, nil
	}

	// Check if the subscriber is not deleted
	subscriber, err := r.GetSubscriber(ctx, subscriberID)
	if err != nil {
		if domain.IsNotFoundError(err) {
			return false, nil
		}
		return false, err
	}

	return !subscriber.IsDeleted, nil
}

// HealthCheck performs a health check on the Dapr state store connection
func (r *DaprSubscriberRepository) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Try to perform a simple state store operation
	testKey := r.stateStore.CreateKey("notifications", "health", "check")
	healthData := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"status":    "healthy",
	}

	err := r.stateStore.Save(ctx, testKey, healthData, nil)
	if err != nil {
		return domain.NewDependencyError("Dapr state store health check failed", err)
	}

	// Clean up test data
	err = r.stateStore.Delete(ctx, testKey, nil)
	if err != nil {
		r.logger.Warn("Failed to clean up health check test data", "error", err)
	}

	return nil
}

// PublishAuditEvent publishes an audit event for compliance logging
func (r *DaprSubscriberRepository) PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error {
	correlationID := domain.GetCorrelationID(ctx)
	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	auditEvent := domain.AuditEvent{
		AuditID:       uuid.New().String(),
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
		DataSnapshot: map[string]interface{}{
			"before": auditEvent.DataSnapshot.Before,
			"after":  auditEvent.DataSnapshot.After,
		},
		Environment: auditEvent.Environment,
	}

	eventMessage := &dapr.EventMessage{
		Topic: "audit-events",
		Data: map[string]interface{}{
			"audit_event": daprAuditEvent,
		},
		Metadata: map[string]string{
			"entity_type":    string(entityType),
			"operation_type": string(operationType),
			"user_id":        userID,
		},
		ContentType: "application/json",
		Source:      "notifications-api",
		Type:        "audit.event",
		Subject:     fmt.Sprintf("audit.%s.%s", string(entityType), string(operationType)),
		Time:        time.Now(),
	}

	err := r.pubsub.PublishEvent(ctx, "audit-events", eventMessage)
	if err != nil {
		return fmt.Errorf("failed to publish notifications audit event: %w", err)
	}

	return nil
}