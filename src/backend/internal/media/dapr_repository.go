package media

import (
	"context"
	"fmt"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/google/uuid"
)

// MediaRepository implements media inquiry data access using Dapr state store and pub/sub
type MediaRepository struct {
	stateStore *dapr.StateStore
	bindings   *dapr.Bindings
	pubsub     *dapr.PubSub
}

// NewMediaRepository creates a new media repository
func NewMediaRepository(client *dapr.Client) *MediaRepository {
	return &MediaRepository{
		stateStore: dapr.NewStateStore(client),
		bindings:   dapr.NewBindings(client),
		pubsub:     dapr.NewPubSub(client),
	}
}

// Media inquiry operations

// SaveInquiry saves media inquiry to Dapr state store
func (r *MediaRepository) SaveInquiry(ctx context.Context, inquiry *MediaInquiry) error {
	key := r.stateStore.CreateKey("media", "inquiry", inquiry.InquiryID)
	
	err := r.stateStore.Save(ctx, key, inquiry, nil)
	if err != nil {
		return fmt.Errorf("failed to save media inquiry %s: %w", inquiry.InquiryID, err)
	}

	// Create index for status search
	statusKey := r.stateStore.CreateIndexKey("media", "inquiry", "status", string(inquiry.Status))
	statusIndex := map[string]string{"inquiry_id": inquiry.InquiryID}
	
	err = r.stateStore.Save(ctx, statusKey, statusIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create status index for media inquiry %s: %w", inquiry.InquiryID, err)
	}

	// Create index for priority search
	priorityKey := r.stateStore.CreateIndexKey("media", "inquiry", "priority", string(inquiry.Priority))
	priorityIndex := map[string]string{"inquiry_id": inquiry.InquiryID}
	
	err = r.stateStore.Save(ctx, priorityKey, priorityIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create priority index for media inquiry %s: %w", inquiry.InquiryID, err)
	}

	// Create index for urgency search
	urgencyKey := r.stateStore.CreateIndexKey("media", "inquiry", "urgency", string(inquiry.Urgency))
	urgencyIndex := map[string]string{"inquiry_id": inquiry.InquiryID}
	
	err = r.stateStore.Save(ctx, urgencyKey, urgencyIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create urgency index for media inquiry %s: %w", inquiry.InquiryID, err)
	}

	// Create index for media type search (if media type is provided)
	if inquiry.MediaType != nil {
		mediaTypeKey := r.stateStore.CreateIndexKey("media", "inquiry", "media_type", string(*inquiry.MediaType))
		mediaTypeIndex := map[string]string{"inquiry_id": inquiry.InquiryID}
		
		err = r.stateStore.Save(ctx, mediaTypeKey, mediaTypeIndex, nil)
		if err != nil {
			return fmt.Errorf("failed to create media type index for media inquiry %s: %w", inquiry.InquiryID, err)
		}
	}

	// Create index for outlet search
	outletKey := r.stateStore.CreateIndexKey("media", "inquiry", "outlet", inquiry.Outlet)
	outletIndex := map[string]string{"inquiry_id": inquiry.InquiryID}
	
	err = r.stateStore.Save(ctx, outletKey, outletIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create outlet index for media inquiry %s: %w", inquiry.InquiryID, err)
	}

	// Create index for deadline search (if deadline is provided)
	if inquiry.Deadline != nil {
		deadlineKey := r.stateStore.CreateIndexKey("media", "inquiry", "deadline", inquiry.Deadline.Format("2006-01-02"))
		deadlineIndex := map[string]string{"inquiry_id": inquiry.InquiryID}
		
		err = r.stateStore.Save(ctx, deadlineKey, deadlineIndex, nil)
		if err != nil {
			return fmt.Errorf("failed to create deadline index for media inquiry %s: %w", inquiry.InquiryID, err)
		}
	}

	// Create index for email search
	emailKey := r.stateStore.CreateIndexKey("media", "inquiry", "email", inquiry.Email)
	emailIndex := map[string]string{"inquiry_id": inquiry.InquiryID}
	
	err = r.stateStore.Save(ctx, emailKey, emailIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create email index for media inquiry %s: %w", inquiry.InquiryID, err)
	}

	// Create index for contact name search
	contactKey := r.stateStore.CreateIndexKey("media", "inquiry", "contact_name", inquiry.ContactName)
	contactIndex := map[string]string{"inquiry_id": inquiry.InquiryID}
	
	err = r.stateStore.Save(ctx, contactKey, contactIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create contact name index for media inquiry %s: %w", inquiry.InquiryID, err)
	}

	// Create index for creation date search
	dateKey := r.stateStore.CreateIndexKey("media", "inquiry", "created_date", inquiry.CreatedAt.Format("2006-01-02"))
	dateIndex := map[string]string{"inquiry_id": inquiry.InquiryID}
	
	err = r.stateStore.Save(ctx, dateKey, dateIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create date index for media inquiry %s: %w", inquiry.InquiryID, err)
	}

	return nil
}

// GetInquiry retrieves a media inquiry by ID from Dapr state store
func (r *MediaRepository) GetInquiry(ctx context.Context, inquiryID string) (*MediaInquiry, error) {
	key := r.stateStore.CreateKey("media", "inquiry", inquiryID)
	
	var inquiry MediaInquiry
	found, err := r.stateStore.Get(ctx, key, &inquiry)
	if err != nil {
		return nil, fmt.Errorf("failed to get media inquiry %s: %w", inquiryID, err)
	}
	
	if !found {
		return nil, domain.NewNotFoundError("media inquiry", inquiryID)
	}
	
	// Check if inquiry is soft deleted
	if inquiry.IsDeleted {
		return nil, domain.NewNotFoundError("media inquiry", inquiryID)
	}
	
	return &inquiry, nil
}

// DeleteInquiry soft deletes a media inquiry
func (r *MediaRepository) DeleteInquiry(ctx context.Context, inquiryID string, userID string) error {
	inquiry, err := r.GetInquiry(ctx, inquiryID)
	if err != nil {
		return err
	}

	inquiry.IsDeleted = true
	now := time.Now()
	inquiry.DeletedAt = &now
	inquiry.UpdatedBy = userID
	inquiry.UpdatedAt = now

	return r.SaveInquiry(ctx, inquiry)
}

// ListInquiries retrieves media inquiries with filtering
func (r *MediaRepository) ListInquiries(ctx context.Context, filters InquiryFilters) ([]*MediaInquiry, error) {
	// For this minimal GREEN phase implementation, return empty list
	// In a production system, this would use proper database queries or indexed searches
	// This will be improved in the REFACTOR phase
	
	return []*MediaInquiry{}, nil
}

// PublishAuditEvent publishes an audit event for compliance logging
func (r *MediaRepository) PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error {
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
		Source:      "media-admin-api",
		Type:        "audit.event",
		Subject:     fmt.Sprintf("audit.%s.%s", string(entityType), string(operationType)),
		Time:        time.Now(),
	}

	err := r.pubsub.PublishEvent(ctx, "audit-events", eventMessage)
	if err != nil {
		return fmt.Errorf("failed to publish media inquiry audit event: %w", err)
	}

	return nil
}