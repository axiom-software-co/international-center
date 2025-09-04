package domain

import (
	"time"

	"github.com/google/uuid"
)

// AuditEventType represents the type of audit event
type AuditEventType string

const (
	AuditEventInsert AuditEventType = "INSERT"
	AuditEventUpdate AuditEventType = "UPDATE" 
	AuditEventDelete AuditEventType = "DELETE"
	AuditEventPublish AuditEventType = "PUBLISH"
	AuditEventArchive AuditEventType = "ARCHIVE"
	AuditEventAccess AuditEventType = "ACCESS"
)

// EntityType represents the type of entity being audited
type EntityType string

const (
	EntityTypeService  EntityType = "service"
	EntityTypeServiceCategory EntityType = "service_category"
	EntityTypeCategory EntityType = "category" 
	EntityTypeFeatured EntityType = "featured_category"
	EntityTypeContent  EntityType = "content"
	EntityTypeUser     EntityType = "user"
	EntityTypeMigration EntityType = "migration"
	EntityTypeNews     EntityType = "news"
	EntityTypeNewsCategory EntityType = "news_category"
	EntityTypeFeaturedNews EntityType = "featured_news"
	EntityTypeResearch EntityType = "research"
	EntityTypeResearchCategory EntityType = "research_category"
	EntityTypeFeaturedResearch EntityType = "featured_research"
	EntityTypeEvent EntityType = "event"
	EntityTypeEventCategory EntityType = "event_category"
	EntityTypeFeaturedEvent EntityType = "featured_event"
	EntityTypeBusiness EntityType = "business"
	EntityTypeDonations EntityType = "donations"
	EntityTypeMedia EntityType = "media"
	EntityTypeVolunteerApplication EntityType = "volunteer_application"
)

// AuditEvent represents a complete audit event for compliance
type AuditEvent struct {
	AuditID       string                 `json:"audit_id"`
	EntityType    EntityType             `json:"entity_type"`
	EntityID      string                 `json:"entity_id"`
	OperationType AuditEventType         `json:"operation_type"`
	AuditTime     time.Time              `json:"audit_timestamp"`
	UserID        string                 `json:"user_id"`
	CorrelationID string                 `json:"correlation_id"`
	TraceID       string                 `json:"trace_id"`
	DataSnapshot  *AuditDataSnapshot     `json:"data_snapshot"`
	Environment   string                 `json:"environment"`
	AppVersion    string                 `json:"app_version"`
	RequestURL    string                 `json:"request_url,omitempty"`
	IPAddress     string                 `json:"ip_address,omitempty"`
	UserAgent     string                 `json:"user_agent,omitempty"`
}

// AuditDataSnapshot contains before and after data for the audited entity
type AuditDataSnapshot struct {
	Before interface{} `json:"before,omitempty"`
	After  interface{} `json:"after,omitempty"`
}

// NewAuditEvent creates a new audit event with generated ID and timestamp
func NewAuditEvent(entityType EntityType, entityID string, operationType AuditEventType, userID string) *AuditEvent {
	return &AuditEvent{
		AuditID:       uuid.New().String(),
		EntityType:    entityType,
		EntityID:      entityID,
		OperationType: operationType,
		AuditTime:     time.Now().UTC(),
		UserID:        userID,
		CorrelationID: uuid.New().String(),
		DataSnapshot:  &AuditDataSnapshot{},
	}
}

// SetTraceContext sets tracing information for the audit event
func (a *AuditEvent) SetTraceContext(correlationID, traceID string) {
	a.CorrelationID = correlationID
	a.TraceID = traceID
}

// SetRequestContext sets HTTP request context information
func (a *AuditEvent) SetRequestContext(requestURL, ipAddress, userAgent string) {
	a.RequestURL = requestURL
	a.IPAddress = ipAddress
	a.UserAgent = userAgent
}

// SetEnvironmentContext sets environment and version information
func (a *AuditEvent) SetEnvironmentContext(environment, appVersion string) {
	a.Environment = environment
	a.AppVersion = appVersion
}

// SetDataSnapshot sets the before and after data for the audit event
func (a *AuditEvent) SetDataSnapshot(before, after interface{}) {
	a.DataSnapshot = &AuditDataSnapshot{
		Before: before,
		After:  after,
	}
}

// SetBeforeData sets only the before data (for delete operations)
func (a *AuditEvent) SetBeforeData(before interface{}) {
	if a.DataSnapshot == nil {
		a.DataSnapshot = &AuditDataSnapshot{}
	}
	a.DataSnapshot.Before = before
}

// SetAfterData sets only the after data (for insert operations)
func (a *AuditEvent) SetAfterData(after interface{}) {
	if a.DataSnapshot == nil {
		a.DataSnapshot = &AuditDataSnapshot{}
	}
	a.DataSnapshot.After = after
}

// Validate ensures the audit event has required fields
func (a *AuditEvent) Validate() error {
	if a.AuditID == "" {
		return NewValidationError("audit_id is required")
	}
	
	if a.EntityType == "" {
		return NewValidationError("entity_type is required")
	}
	
	if a.EntityID == "" {
		return NewValidationError("entity_id is required") 
	}
	
	if a.OperationType == "" {
		return NewValidationError("operation_type is required")
	}
	
	if a.AuditTime.IsZero() {
		return NewValidationError("audit_timestamp is required")
	}
	
	if a.UserID == "" {
		return NewValidationError("user_id is required")
	}
	
	return nil
}

// IsValidEntityType checks if the entity type is valid
func IsValidEntityType(entityType EntityType) bool {
	switch entityType {
	case EntityTypeService, EntityTypeServiceCategory, EntityTypeCategory, EntityTypeFeatured, EntityTypeContent, EntityTypeUser, EntityTypeMigration, EntityTypeNews, EntityTypeNewsCategory, EntityTypeFeaturedNews, EntityTypeResearch, EntityTypeResearchCategory, EntityTypeFeaturedResearch, EntityTypeEvent, EntityTypeEventCategory, EntityTypeFeaturedEvent, EntityTypeBusiness, EntityTypeDonations, EntityTypeMedia, EntityTypeVolunteerApplication:
		return true
	default:
		return false
	}
}

// IsValidOperationType checks if the operation type is valid
func IsValidOperationType(operationType AuditEventType) bool {
	switch operationType {
	case AuditEventInsert, AuditEventUpdate, AuditEventDelete, AuditEventPublish, AuditEventArchive, AuditEventAccess:
		return true
	default:
		return false
	}
}