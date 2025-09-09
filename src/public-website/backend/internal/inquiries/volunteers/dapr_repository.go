package volunteers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
)

// VolunteerRepository implements volunteer data access using Dapr state store and bindings
type VolunteerRepository struct {
	stateStore *dapr.StateStore
	bindings   *dapr.Bindings
	pubsub     *dapr.PubSub
}

// NewVolunteerRepository creates a new volunteer repository with Dapr integration
func NewVolunteerRepository(stateStore *dapr.StateStore, bindings *dapr.Bindings, pubsub *dapr.PubSub) VolunteerRepositoryInterface {
	return &VolunteerRepository{
		stateStore: stateStore,
		bindings:   bindings,
		pubsub:     pubsub,
	}
}

// Volunteer application operations

// SaveVolunteerApplication saves a volunteer application to the state store
func (r *VolunteerRepository) SaveVolunteerApplication(ctx context.Context, application *VolunteerApplication) error {
	key := r.stateStore.CreateKey("volunteers", "volunteer_application", application.ApplicationID)
	
	err := r.stateStore.Save(ctx, key, application, nil)
	if err != nil {
		return fmt.Errorf("failed to save volunteer application %s: %w", application.ApplicationID, err)
	}

	return nil
}

// GetVolunteerApplication retrieves volunteer application by ID from state store
func (r *VolunteerRepository) GetVolunteerApplication(ctx context.Context, applicationID string) (*VolunteerApplication, error) {
	key := r.stateStore.CreateKey("volunteers", "volunteer_application", applicationID)
	
	var application VolunteerApplication
	found, err := r.stateStore.Get(ctx, key, &application)
	if err != nil {
		return nil, fmt.Errorf("failed to get volunteer application %s: %w", applicationID, err)
	}
	
	if !found || application.IsDeleted {
		return nil, domain.NewNotFoundError("volunteer_application", applicationID)
	}
	
	return &application, nil
}

// GetAllVolunteerApplications retrieves all volunteer applications from state store with pagination
func (r *VolunteerRepository) GetAllVolunteerApplications(ctx context.Context, limit, offset int) ([]*VolunteerApplication, error) {
	query := `{
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

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all volunteer applications: %w", err)
	}

	var applicationsList []*VolunteerApplication
	count := 0
	for _, result := range results {
		// Apply offset
		if count < offset {
			count++
			continue
		}
		
		// Apply limit
		if len(applicationsList) >= limit {
			break
		}

		var application VolunteerApplication
		err = json.Unmarshal(result.Value, &application)
		if err != nil {
			continue // Skip invalid records
		}
		if !application.IsDeleted {
			applicationsList = append(applicationsList, &application)
		}
		count++
	}

	return applicationsList, nil
}

// GetVolunteerApplicationsByStatus retrieves applications by status from state store with pagination
func (r *VolunteerRepository) GetVolunteerApplicationsByStatus(ctx context.Context, status ApplicationStatus, limit, offset int) ([]*VolunteerApplication, error) {
	query := fmt.Sprintf(`{
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
	}`, status)

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query volunteer applications by status %s: %w", status, err)
	}

	var applicationsList []*VolunteerApplication
	count := 0
	for _, result := range results {
		// Apply offset
		if count < offset {
			count++
			continue
		}
		
		// Apply limit
		if len(applicationsList) >= limit {
			break
		}

		var application VolunteerApplication
		err = json.Unmarshal(result.Value, &application)
		if err != nil {
			continue
		}
		if !application.IsDeleted && application.Status == status {
			applicationsList = append(applicationsList, &application)
		}
		count++
	}

	return applicationsList, nil
}

// GetVolunteerApplicationsByPriority retrieves applications by priority from state store with pagination
func (r *VolunteerRepository) GetVolunteerApplicationsByPriority(ctx context.Context, priority ApplicationPriority, limit, offset int) ([]*VolunteerApplication, error) {
	query := fmt.Sprintf(`{
		"filter": {
			"AND": [
				{"EQ": {"priority": "%s"}},
				{"EQ": {"is_deleted": false}}
			]
		},
		"sort": [
			{
				"key": "created_at", 
				"order": "DESC"
			}
		]
	}`, priority)

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query volunteer applications by priority %s: %w", priority, err)
	}

	var applicationsList []*VolunteerApplication
	count := 0
	for _, result := range results {
		// Apply offset
		if count < offset {
			count++
			continue
		}
		
		// Apply limit
		if len(applicationsList) >= limit {
			break
		}

		var application VolunteerApplication
		err = json.Unmarshal(result.Value, &application)
		if err != nil {
			continue
		}
		if !application.IsDeleted && application.Priority == priority {
			applicationsList = append(applicationsList, &application)
		}
		count++
	}

	return applicationsList, nil
}

// GetVolunteerApplicationsByInterest retrieves applications by volunteer interest from state store with pagination
func (r *VolunteerRepository) GetVolunteerApplicationsByInterest(ctx context.Context, interest VolunteerInterest, limit, offset int) ([]*VolunteerApplication, error) {
	query := fmt.Sprintf(`{
		"filter": {
			"AND": [
				{"EQ": {"volunteer_interest": "%s"}},
				{"EQ": {"is_deleted": false}}
			]
		},
		"sort": [
			{
				"key": "created_at", 
				"order": "DESC"
			}
		]
	}`, interest)

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query volunteer applications by interest %s: %w", interest, err)
	}

	var applicationsList []*VolunteerApplication
	count := 0
	for _, result := range results {
		// Apply offset
		if count < offset {
			count++
			continue
		}
		
		// Apply limit
		if len(applicationsList) >= limit {
			break
		}

		var application VolunteerApplication
		err = json.Unmarshal(result.Value, &application)
		if err != nil {
			continue
		}
		if !application.IsDeleted && application.VolunteerInterest == interest {
			applicationsList = append(applicationsList, &application)
		}
		count++
	}

	return applicationsList, nil
}

// SearchVolunteerApplications searches volunteer applications by query terms
func (r *VolunteerRepository) SearchVolunteerApplications(ctx context.Context, searchTerm string, limit, offset int) ([]*VolunteerApplication, error) {
	query := `{
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

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search volunteer applications: %w", err)
	}

	var applicationsList []*VolunteerApplication
	count := 0
	searchLower := strings.ToLower(searchTerm)

	for _, result := range results {
		var application VolunteerApplication
		err = json.Unmarshal(result.Value, &application)
		if err != nil {
			continue
		}

		if application.IsDeleted {
			continue
		}

		// Search in name, email, and motivation
		fullName := strings.ToLower(application.FirstName + " " + application.LastName)
		email := strings.ToLower(application.Email)
		motivation := strings.ToLower(application.Motivation)

		if strings.Contains(fullName, searchLower) ||
		   strings.Contains(email, searchLower) ||
		   strings.Contains(motivation, searchLower) {

			// Apply offset
			if count < offset {
				count++
				continue
			}
			
			// Apply limit
			if len(applicationsList) >= limit {
				break
			}

			applicationsList = append(applicationsList, &application)
		}
		count++
	}

	return applicationsList, nil
}

// DeleteVolunteerApplication soft deletes a volunteer application
func (r *VolunteerRepository) DeleteVolunteerApplication(ctx context.Context, applicationID string) error {
	// Get the existing application first
	application, err := r.GetVolunteerApplication(ctx, applicationID)
	if err != nil {
		return err
	}

	// Mark as deleted
	application.IsDeleted = true
	now := time.Now()
	application.DeletedAt = &now

	// Save the soft-deleted application
	return r.SaveVolunteerApplication(ctx, application)
}

// Audit operations

// GetVolunteerApplicationAudit retrieves audit events for a volunteer application
func (r *VolunteerRepository) GetVolunteerApplicationAudit(ctx context.Context, applicationID string, userID string, limit int, offset int) ([]*domain.AuditEvent, error) {
	// Query audit events from the audit topic via Dapr pubsub
	// This is a simplified implementation - in production this would query from Grafana Cloud Loki
	
	// Create mock audit events for testing
	var auditEvents []*domain.AuditEvent
	
	// This would typically query from external audit store (Grafana Cloud Loki)
	// For now return empty slice as audit events are published but not stored locally
	
	return auditEvents, nil
}

// PublishAuditEvent publishes an audit event via Dapr pubsub to Grafana Cloud Loki
func (r *VolunteerRepository) PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error {
	auditEvent := domain.NewAuditEvent(entityType, entityID, operationType, userID)
	auditEvent.SetDataSnapshot(beforeData, afterData)
	auditEvent.SetEnvironmentContext("development", "1.0.0") // TODO: Get from config

	// Convert domain.AuditEvent to Dapr AuditEvent structure  
	daprAuditEvent := &dapr.AuditEvent{
		AuditID:       auditEvent.AuditID,
		EntityType:    string(auditEvent.EntityType),
		EntityID:      auditEvent.EntityID,
		OperationType: string(auditEvent.OperationType),
		AuditTime:     auditEvent.AuditTime,
		UserID:        auditEvent.UserID,
		CorrelationID: auditEvent.CorrelationID,
		TraceID:       auditEvent.TraceID,
		Environment:   auditEvent.Environment,
		AppVersion:    auditEvent.AppVersion,
		RequestURL:    auditEvent.RequestURL,
		DataSnapshot:  make(map[string]interface{}),
	}

	if auditEvent.DataSnapshot != nil {
		if auditEvent.DataSnapshot.Before != nil {
			daprAuditEvent.DataSnapshot["before"] = auditEvent.DataSnapshot.Before
		}
		if auditEvent.DataSnapshot.After != nil {
			daprAuditEvent.DataSnapshot["after"] = auditEvent.DataSnapshot.After
		}
	}

	// Publish audit event to Grafana Cloud Loki via Dapr pubsub
	err := r.pubsub.PublishAuditEvent(ctx, daprAuditEvent)
	if err != nil {
		return fmt.Errorf("failed to publish audit event for %s %s: %w", entityType, entityID, err)
	}

	return nil
}