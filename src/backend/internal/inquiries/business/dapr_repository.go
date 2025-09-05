package business

import (
	"context"
	"fmt"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/google/uuid"
)

// BusinessRepository implements business inquiry data access using Dapr state store and pub/sub
type BusinessRepository struct {
	stateStore *dapr.StateStore
	bindings   *dapr.Bindings
	pubsub     *dapr.PubSub
}

// NewBusinessRepository creates a new business repository
func NewBusinessRepository(client *dapr.Client) *BusinessRepository {
	return &BusinessRepository{
		stateStore: dapr.NewStateStore(client),
		bindings:   dapr.NewBindings(client),
		pubsub:     dapr.NewPubSub(client),
	}
}

// Business inquiry operations

// SaveInquiry saves business inquiry to Dapr state store
func (r *BusinessRepository) SaveInquiry(ctx context.Context, inquiry *BusinessInquiry) error {
	key := r.stateStore.CreateKey("business", "inquiry", inquiry.InquiryID)
	
	err := r.stateStore.Save(ctx, key, inquiry, nil)
	if err != nil {
		return fmt.Errorf("failed to save business inquiry %s: %w", inquiry.InquiryID, err)
	}

	// Create index for status search
	statusKey := r.stateStore.CreateIndexKey("business", "inquiry", "status", string(inquiry.Status))
	statusIndex := map[string]string{"inquiry_id": inquiry.InquiryID}
	
	err = r.stateStore.Save(ctx, statusKey, statusIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create status index for business inquiry %s: %w", inquiry.InquiryID, err)
	}

	// Create index for priority search
	priorityKey := r.stateStore.CreateIndexKey("business", "inquiry", "priority", string(inquiry.Priority))
	priorityIndex := map[string]string{"inquiry_id": inquiry.InquiryID}
	
	err = r.stateStore.Save(ctx, priorityKey, priorityIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create priority index for business inquiry %s: %w", inquiry.InquiryID, err)
	}

	// Create index for inquiry type search
	typeKey := r.stateStore.CreateIndexKey("business", "inquiry", "type", string(inquiry.InquiryType))
	typeIndex := map[string]string{"inquiry_id": inquiry.InquiryID}
	
	err = r.stateStore.Save(ctx, typeKey, typeIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create type index for business inquiry %s: %w", inquiry.InquiryID, err)
	}

	// Create index for industry search (if industry is provided)
	if inquiry.Industry != nil && *inquiry.Industry != "" {
		industryKey := r.stateStore.CreateIndexKey("business", "inquiry", "industry", *inquiry.Industry)
		industryIndex := map[string]string{"inquiry_id": inquiry.InquiryID}
		
		err = r.stateStore.Save(ctx, industryKey, industryIndex, nil)
		if err != nil {
			return fmt.Errorf("failed to create industry index for business inquiry %s: %w", inquiry.InquiryID, err)
		}
	}

	// Create index for organization search
	orgKey := r.stateStore.CreateIndexKey("business", "inquiry", "organization", inquiry.OrganizationName)
	orgIndex := map[string]string{"inquiry_id": inquiry.InquiryID}
	
	err = r.stateStore.Save(ctx, orgKey, orgIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create organization index for business inquiry %s: %w", inquiry.InquiryID, err)
	}

	// Create index for creation date search
	dateKey := r.stateStore.CreateIndexKey("business", "inquiry", "created_date", inquiry.CreatedAt.Format("2006-01-02"))
	dateIndex := map[string]string{"inquiry_id": inquiry.InquiryID}
	
	err = r.stateStore.Save(ctx, dateKey, dateIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create date index for business inquiry %s: %w", inquiry.InquiryID, err)
	}

	return nil
}

// GetInquiry retrieves a business inquiry by ID from Dapr state store
func (r *BusinessRepository) GetInquiry(ctx context.Context, inquiryID string) (*BusinessInquiry, error) {
	key := r.stateStore.CreateKey("business", "inquiry", inquiryID)
	
	var inquiry BusinessInquiry
	found, err := r.stateStore.Get(ctx, key, &inquiry)
	if err != nil {
		return nil, fmt.Errorf("failed to get business inquiry %s: %w", inquiryID, err)
	}
	
	if !found {
		return nil, domain.NewNotFoundError("business inquiry", inquiryID)
	}
	
	// Check if inquiry is soft deleted
	if inquiry.IsDeleted {
		return nil, domain.NewNotFoundError("business inquiry", inquiryID)
	}
	
	return &inquiry, nil
}

// DeleteInquiry soft deletes a business inquiry
func (r *BusinessRepository) DeleteInquiry(ctx context.Context, inquiryID string, userID string) error {
	inquiry, err := r.GetInquiry(ctx, inquiryID)
	if err != nil {
		return err
	}

	inquiry.IsDeleted = true
	now := time.Now().UTC()
	inquiry.DeletedAt = &now
	inquiry.UpdatedBy = userID
	inquiry.UpdatedAt = now

	return r.SaveInquiry(ctx, inquiry)
}

// ListInquiries retrieves business inquiries with filtering
func (r *BusinessRepository) ListInquiries(ctx context.Context, filters InquiryFilters) ([]*BusinessInquiry, error) {
	// For this minimal GREEN phase implementation, return empty list
	// In a production system, this would use proper database queries or indexed searches
	// This will be improved in the REFACTOR phase
	
	return []*BusinessInquiry{}, nil
}

// PublishAuditEvent publishes an audit event for compliance logging
func (r *BusinessRepository) PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error {
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
		Source:      "business-admin-api",
		Type:        "audit.event",
		Subject:     fmt.Sprintf("audit.%s.%s", string(entityType), string(operationType)),
		Time:        time.Now(),
	}

	err := r.pubsub.PublishEvent(ctx, "audit-events", eventMessage)
	if err != nil {
		return fmt.Errorf("failed to publish business inquiry audit event: %w", err)
	}

	return nil
}