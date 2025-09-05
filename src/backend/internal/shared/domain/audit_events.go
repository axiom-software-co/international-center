package domain

import (
	"time"

	"github.com/google/uuid"
)

// AuditEventType represents the type of operation being audited for compliance tracking.
// Each type corresponds to a specific database operation that requires audit trail documentation
// for regulatory compliance and security monitoring.
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
	// Services Domain
	EntityTypeService  EntityType = "service"
	EntityTypeServiceCategory EntityType = "service_category"
	EntityTypeFeaturedCategory EntityType = "featured_category"
	
	// News Domain  
	EntityTypeNews     EntityType = "news"
	EntityTypeNewsCategory EntityType = "news_category"
	EntityTypeFeaturedNews EntityType = "featured_news"
	
	// Research Domain
	EntityTypeResearch EntityType = "research"
	EntityTypeResearchCategory EntityType = "research_category"
	EntityTypeFeaturedResearch EntityType = "featured_research"
	
	// Events Domain
	EntityTypeEvent EntityType = "event"
	EntityTypeEventCategory EntityType = "event_category"
	EntityTypeFeaturedEvent EntityType = "featured_event"
	EntityTypeEventRegistration EntityType = "event_registration"
	
	// Inquiry Domains
	EntityTypeBusinessInquiry EntityType = "business_inquiry"
	EntityTypeDonationsInquiry EntityType = "donations_inquiry"
	EntityTypeMediaInquiry EntityType = "media_inquiry"
	EntityTypeVolunteerApplication EntityType = "volunteer_application"
	
	// System Entities
	EntityTypeUser     EntityType = "user"
	EntityTypeMigration EntityType = "migration"
)

// AuditEvent represents a complete audit event for regulatory compliance and security monitoring.
// This structure captures all necessary information for immutable audit trails that support
// compliance requirements including HIPAA, SOC 2, and other regulatory frameworks.
//
// The audit event includes complete context about the operation, user, environment, and
// data changes to ensure comprehensive audit trail documentation.
type AuditEvent struct {
	AuditID       string                 `json:"audit_id"`         // Unique identifier for the audit event
	EntityType    EntityType             `json:"entity_type"`      // Type of entity being audited
	EntityID      string                 `json:"entity_id"`        // Unique identifier of the audited entity
	OperationType AuditEventType         `json:"operation_type"`   // Type of operation (INSERT, UPDATE, DELETE, etc.)
	AuditTime     time.Time              `json:"audit_timestamp"`  // Timestamp when the audit event occurred
	UserID        string                 `json:"user_id"`          // Identifier of the user performing the operation
	CorrelationID string                 `json:"correlation_id"`   // Request correlation identifier for distributed tracing
	TraceID       string                 `json:"trace_id"`         // Trace identifier for request tracking
	DataSnapshot  *AuditDataSnapshot     `json:"data_snapshot"`    // Before and after data snapshots
	Environment   string                 `json:"environment"`      // Environment where the operation occurred
	AppVersion    string                 `json:"app_version"`      // Application version performing the operation
	RequestURL    string                 `json:"request_url,omitempty"`  // HTTP request URL (if applicable)
	IPAddress     string                 `json:"ip_address,omitempty"`   // Client IP address (if applicable)
	UserAgent     string                 `json:"user_agent,omitempty"`   // Client user agent (if applicable)
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
	case EntityTypeService, EntityTypeServiceCategory, EntityTypeFeaturedCategory,
		 EntityTypeNews, EntityTypeNewsCategory, EntityTypeFeaturedNews,
		 EntityTypeResearch, EntityTypeResearchCategory, EntityTypeFeaturedResearch,
		 EntityTypeEvent, EntityTypeEventCategory, EntityTypeFeaturedEvent, EntityTypeEventRegistration,
		 EntityTypeBusinessInquiry, EntityTypeDonationsInquiry, EntityTypeMediaInquiry, EntityTypeVolunteerApplication,
		 EntityTypeUser, EntityTypeMigration:
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