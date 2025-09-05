package donations

import (
	"context"
	"fmt"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/google/uuid"
)

// DonationsRepository implements donations inquiry data access using Dapr state store and pub/sub
type DonationsRepository struct {
	stateStore *dapr.StateStore
	bindings   *dapr.Bindings
	pubsub     *dapr.PubSub
}

// NewDonationsRepository creates a new donations repository
func NewDonationsRepository(client *dapr.Client) *DonationsRepository {
	return &DonationsRepository{
		stateStore: dapr.NewStateStore(client),
		bindings:   dapr.NewBindings(client),
		pubsub:     dapr.NewPubSub(client),
	}
}

// Donations inquiry operations

// SaveInquiry saves donations inquiry to Dapr state store
func (r *DonationsRepository) SaveInquiry(ctx context.Context, inquiry *DonationsInquiry) error {
	key := r.stateStore.CreateKey("donations", "inquiry", inquiry.InquiryID)
	
	err := r.stateStore.Save(ctx, key, inquiry, nil)
	if err != nil {
		return fmt.Errorf("failed to save donations inquiry %s: %w", inquiry.InquiryID, err)
	}

	// Create index for status search
	statusKey := r.stateStore.CreateIndexKey("donations", "inquiry", "status", string(inquiry.Status))
	statusIndex := map[string]string{"inquiry_id": inquiry.InquiryID}
	
	err = r.stateStore.Save(ctx, statusKey, statusIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create status index for donations inquiry %s: %w", inquiry.InquiryID, err)
	}

	// Create index for priority search
	priorityKey := r.stateStore.CreateIndexKey("donations", "inquiry", "priority", string(inquiry.Priority))
	priorityIndex := map[string]string{"inquiry_id": inquiry.InquiryID}
	
	err = r.stateStore.Save(ctx, priorityKey, priorityIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create priority index for donations inquiry %s: %w", inquiry.InquiryID, err)
	}

	// Create index for donor type search
	donorTypeKey := r.stateStore.CreateIndexKey("donations", "inquiry", "donor_type", string(inquiry.DonorType))
	donorTypeIndex := map[string]string{"inquiry_id": inquiry.InquiryID}
	
	err = r.stateStore.Save(ctx, donorTypeKey, donorTypeIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create donor type index for donations inquiry %s: %w", inquiry.InquiryID, err)
	}

	// Create index for interest area search (if interest area is provided)
	if inquiry.InterestArea != nil {
		interestKey := r.stateStore.CreateIndexKey("donations", "inquiry", "interest_area", string(*inquiry.InterestArea))
		interestIndex := map[string]string{"inquiry_id": inquiry.InquiryID}
		
		err = r.stateStore.Save(ctx, interestKey, interestIndex, nil)
		if err != nil {
			return fmt.Errorf("failed to create interest area index for donations inquiry %s: %w", inquiry.InquiryID, err)
		}
	}

	// Create index for amount range search (if amount range is provided)
	if inquiry.PreferredAmountRange != nil {
		amountKey := r.stateStore.CreateIndexKey("donations", "inquiry", "amount_range", string(*inquiry.PreferredAmountRange))
		amountIndex := map[string]string{"inquiry_id": inquiry.InquiryID}
		
		err = r.stateStore.Save(ctx, amountKey, amountIndex, nil)
		if err != nil {
			return fmt.Errorf("failed to create amount range index for donations inquiry %s: %w", inquiry.InquiryID, err)
		}
	}

	// Create index for donation frequency search (if donation frequency is provided)
	if inquiry.DonationFrequency != nil {
		frequencyKey := r.stateStore.CreateIndexKey("donations", "inquiry", "donation_frequency", string(*inquiry.DonationFrequency))
		frequencyIndex := map[string]string{"inquiry_id": inquiry.InquiryID}
		
		err = r.stateStore.Save(ctx, frequencyKey, frequencyIndex, nil)
		if err != nil {
			return fmt.Errorf("failed to create donation frequency index for donations inquiry %s: %w", inquiry.InquiryID, err)
		}
	}

	// Create index for organization search (if organization is provided)
	if inquiry.Organization != nil && *inquiry.Organization != "" {
		orgKey := r.stateStore.CreateIndexKey("donations", "inquiry", "organization", *inquiry.Organization)
		orgIndex := map[string]string{"inquiry_id": inquiry.InquiryID}
		
		err = r.stateStore.Save(ctx, orgKey, orgIndex, nil)
		if err != nil {
			return fmt.Errorf("failed to create organization index for donations inquiry %s: %w", inquiry.InquiryID, err)
		}
	}

	// Create index for email search
	emailKey := r.stateStore.CreateIndexKey("donations", "inquiry", "email", inquiry.Email)
	emailIndex := map[string]string{"inquiry_id": inquiry.InquiryID}
	
	err = r.stateStore.Save(ctx, emailKey, emailIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create email index for donations inquiry %s: %w", inquiry.InquiryID, err)
	}

	// Create index for creation date search
	dateKey := r.stateStore.CreateIndexKey("donations", "inquiry", "created_date", inquiry.CreatedAt.Format("2006-01-02"))
	dateIndex := map[string]string{"inquiry_id": inquiry.InquiryID}
	
	err = r.stateStore.Save(ctx, dateKey, dateIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create date index for donations inquiry %s: %w", inquiry.InquiryID, err)
	}

	return nil
}

// GetInquiry retrieves a donations inquiry by ID from Dapr state store
func (r *DonationsRepository) GetInquiry(ctx context.Context, inquiryID string) (*DonationsInquiry, error) {
	key := r.stateStore.CreateKey("donations", "inquiry", inquiryID)
	
	var inquiry DonationsInquiry
	found, err := r.stateStore.Get(ctx, key, &inquiry)
	if err != nil {
		return nil, fmt.Errorf("failed to get donations inquiry %s: %w", inquiryID, err)
	}
	
	if !found {
		return nil, domain.NewNotFoundError("donations inquiry", inquiryID)
	}
	
	// Check if inquiry is soft deleted
	if inquiry.IsDeleted {
		return nil, domain.NewNotFoundError("donations inquiry", inquiryID)
	}
	
	return &inquiry, nil
}

// DeleteInquiry soft deletes a donations inquiry
func (r *DonationsRepository) DeleteInquiry(ctx context.Context, inquiryID string, userID string) error {
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

// ListInquiries retrieves donations inquiries with filtering
func (r *DonationsRepository) ListInquiries(ctx context.Context, filters InquiryFilters) ([]*DonationsInquiry, error) {
	// For this minimal GREEN phase implementation, return empty list
	// In a production system, this would use proper database queries or indexed searches
	// This will be improved in the REFACTOR phase
	
	return []*DonationsInquiry{}, nil
}

// PublishAuditEvent publishes an audit event for compliance logging
func (r *DonationsRepository) PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error {
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
		Source:      "donations-admin-api",
		Type:        "audit.event",
		Subject:     fmt.Sprintf("audit.%s.%s", string(entityType), string(operationType)),
		Time:        time.Now(),
	}

	err := r.pubsub.PublishEvent(ctx, "audit-events", eventMessage)
	if err != nil {
		return fmt.Errorf("failed to publish donations inquiry audit event: %w", err)
	}

	return nil
}