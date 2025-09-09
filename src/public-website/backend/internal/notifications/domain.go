package notifications

import (
	"strings"
	"time"
)

// Domain types matching TABLES-INTERNAL-NOTIFICATIONS-SUBSCRIBERS.md schema
type SubscriberStatus string
type EventType string
type NotificationMethod string
type NotificationSchedule string
type PriorityThreshold string

const (
	SubscriberStatusActive    SubscriberStatus = "active"
	SubscriberStatusInactive  SubscriberStatus = "inactive"
	SubscriberStatusSuspended SubscriberStatus = "suspended"
)

const (
	EventTypeInquiryMedia        EventType = "inquiry-media"
	EventTypeInquiryBusiness     EventType = "inquiry-business"
	EventTypeInquiryDonations    EventType = "inquiry-donations"
	EventTypeInquiryVolunteers   EventType = "inquiry-volunteers"
	EventTypeEventRegistration   EventType = "event-registration"
	EventTypeSystemError         EventType = "system-error"
	EventTypeCapacityAlert       EventType = "capacity-alert"
	EventTypeAdminActionRequired EventType = "admin-action-required"
	EventTypeComplianceAlert     EventType = "compliance-alert"
)

const (
	NotificationMethodEmail NotificationMethod = "email"
	NotificationMethodSMS   NotificationMethod = "sms"
	NotificationMethodBoth  NotificationMethod = "both"
)

const (
	ScheduleImmediate NotificationSchedule = "immediate"
	ScheduleHourly    NotificationSchedule = "hourly"
	ScheduleDaily     NotificationSchedule = "daily"
)

const (
	PriorityLow    PriorityThreshold = "low"
	PriorityMedium PriorityThreshold = "medium"
	PriorityHigh   PriorityThreshold = "high"
	PriorityUrgent PriorityThreshold = "urgent"
)

// NotificationSubscriber matches database schema exactly
type NotificationSubscriber struct {
	SubscriberID         string                `json:"subscriber_id"`
	Status               SubscriberStatus      `json:"status"`
	SubscriberName       string                `json:"subscriber_name"`
	Email                string                `json:"email"`
	Phone                *string               `json:"phone,omitempty"`
	EventTypes           []EventType           `json:"event_types"`
	NotificationMethods  []NotificationMethod  `json:"notification_methods"`
	NotificationSchedule NotificationSchedule  `json:"notification_schedule"`
	PriorityThreshold    PriorityThreshold     `json:"priority_threshold"`
	Notes                *string               `json:"notes,omitempty"`
	CreatedAt            time.Time             `json:"created_at"`
	UpdatedAt            time.Time             `json:"updated_at"`
	CreatedBy            string                `json:"created_by"`
	UpdatedBy            string                `json:"updated_by"`
	IsDeleted            bool                  `json:"is_deleted"`
	DeletedAt            *time.Time            `json:"deleted_at,omitempty"`
}

// DomainEvent represents events from existing domain APIs
type DomainEvent struct {
	EventID       string                 `json:"event_id"`
	Topic         string                 `json:"topic"`
	EventType     string                 `json:"event_type"`
	Priority      PriorityThreshold      `json:"priority"`
	EntityType    string                 `json:"entity_type"`
	EntityID      string                 `json:"entity_id"`
	OperationType string                 `json:"operation_type"`
	UserID        string                 `json:"user_id"`
	CorrelationID string                 `json:"correlation_id"`
	EventData     map[string]interface{} `json:"event_data"`
	Timestamp     time.Time              `json:"timestamp"`
	Environment   string                 `json:"environment"`
}

// NotificationMessage represents messages sent to channel handlers
type NotificationMessage struct {
	MessageID     string                 `json:"message_id"`
	SubscriberID  string                 `json:"subscriber_id"`
	EventType     EventType              `json:"event_type"`
	Priority      PriorityThreshold      `json:"priority"`
	EventData     map[string]interface{} `json:"event_data"`
	Recipients    NotificationRecipients `json:"recipients"`
	Schedule      NotificationSchedule   `json:"schedule"`
	CreatedAt     time.Time              `json:"created_at"`
	CorrelationID string                 `json:"correlation_id"`
}

type NotificationRecipients struct {
	Email []string `json:"email,omitempty"`
	SMS   []string `json:"sms,omitempty"`
	Slack []string `json:"slack,omitempty"`
}

type ChannelRoute struct {
	Channel      NotificationChannel `json:"channel"`
	Topic        string              `json:"topic"`
	SubscriberID string              `json:"subscriber_id"`
	Recipient    string              `json:"recipient"`
}

type NotificationChannel string

const (
	ChannelEmail NotificationChannel = "email"
	ChannelSMS   NotificationChannel = "sms"
	ChannelSlack NotificationChannel = "slack"
)

// IsValid validates subscriber status
func (s SubscriberStatus) IsValid() bool {
	switch s {
	case SubscriberStatusActive, SubscriberStatusInactive, SubscriberStatusSuspended:
		return true
	default:
		return false
	}
}

// IsValid validates event type
func (e EventType) IsValid() bool {
	switch e {
	case EventTypeInquiryMedia, EventTypeInquiryBusiness, EventTypeInquiryDonations,
		EventTypeInquiryVolunteers, EventTypeEventRegistration, EventTypeSystemError,
		EventTypeCapacityAlert, EventTypeAdminActionRequired, EventTypeComplianceAlert:
		return true
	default:
		return false
	}
}

// IsValid validates notification method
func (n NotificationMethod) IsValid() bool {
	switch n {
	case NotificationMethodEmail, NotificationMethodSMS, NotificationMethodBoth:
		return true
	default:
		return false
	}
}

// IsValid validates notification schedule
func (n NotificationSchedule) IsValid() bool {
	switch n {
	case ScheduleImmediate, ScheduleHourly, ScheduleDaily:
		return true
	default:
		return false
	}
}

// IsValid validates priority threshold
func (p PriorityThreshold) IsValid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent:
		return true
	default:
		return false
	}
}

// GetPriorityValue returns numeric value for priority comparison
func (p PriorityThreshold) GetPriorityValue() int {
	switch p {
	case PriorityUrgent:
		return 4
	case PriorityHigh:
		return 3
	case PriorityMedium:
		return 2
	case PriorityLow:
		return 1
	default:
		return 0
	}
}

// HasEmailMethod checks if subscriber has email notification method
func (s *NotificationSubscriber) HasEmailMethod() bool {
	for _, method := range s.NotificationMethods {
		if method == NotificationMethodEmail || method == NotificationMethodBoth {
			return true
		}
	}
	return false
}

// HasSMSMethod checks if subscriber has SMS notification method
func (s *NotificationSubscriber) HasSMSMethod() bool {
	for _, method := range s.NotificationMethods {
		if method == NotificationMethodSMS || method == NotificationMethodBoth {
			return true
		}
	}
	return false
}

// ClassifyDomainEvent maps domain events to schema event types
func ClassifyDomainEvent(topic string, eventData map[string]interface{}) EventType {
	switch {
	case strings.Contains(topic, "business-inquiry"):
		return EventTypeInquiryBusiness
	case strings.Contains(topic, "media-inquiry"):
		return EventTypeInquiryMedia
	case strings.Contains(topic, "donation-inquiry"):
		return EventTypeInquiryDonations
	case strings.Contains(topic, "volunteer-inquiry"):
		return EventTypeInquiryVolunteers
	case isContentCreationEvent(topic, eventData):
		return EventTypeEventRegistration
	case isSystemErrorEvent(eventData):
		return EventTypeSystemError
	case isCapacityEvent(eventData):
		return EventTypeCapacityAlert
	case isComplianceEvent(eventData):
		return EventTypeComplianceAlert
	default:
		return EventTypeAdminActionRequired
	}
}

// Helper functions for event classification
func isContentCreationEvent(topic string, eventData map[string]interface{}) bool {
	contentTopics := []string{"services-content", "news-content", "research-content", "events-content"}
	for _, contentTopic := range contentTopics {
		if strings.Contains(topic, contentTopic) {
			if operation, exists := eventData["operation_type"]; exists {
				return operation == "CREATE" || operation == "PUBLISH"
			}
		}
	}
	return false
}

func isSystemErrorEvent(eventData map[string]interface{}) bool {
	if eventType, exists := eventData["event_type"]; exists {
		return strings.Contains(strings.ToLower(eventType.(string)), "error")
	}
	return false
}

func isCapacityEvent(eventData map[string]interface{}) bool {
	if eventType, exists := eventData["event_type"]; exists {
		eventStr := strings.ToLower(eventType.(string))
		return strings.Contains(eventStr, "capacity") || strings.Contains(eventStr, "limit")
	}
	return false
}

func isComplianceEvent(eventData map[string]interface{}) bool {
	if eventType, exists := eventData["event_type"]; exists {
		eventStr := strings.ToLower(eventType.(string))
		return strings.Contains(eventStr, "audit") || strings.Contains(eventStr, "compliance")
	}
	return false
}

// ExtractPriorityFromEvent extracts priority from domain event
func ExtractPriorityFromEvent(eventData map[string]interface{}) PriorityThreshold {
	if priority, exists := eventData["priority"]; exists {
		if priorityStr, ok := priority.(string); ok {
			switch PriorityThreshold(priorityStr) {
			case PriorityUrgent, PriorityHigh, PriorityMedium, PriorityLow:
				return PriorityThreshold(priorityStr)
			}
		}
	}
	
	// Default priority based on operation type
	if operation, exists := eventData["operation_type"]; exists {
		switch operation {
		case "DELETE":
			return PriorityHigh
		case "CREATE", "PUBLISH":
			return PriorityMedium
		case "UPDATE":
			return PriorityLow
		}
	}
	
	return PriorityLow
}

// DeterminePriority determines the priority of a domain event
func DeterminePriority(eventType string, eventData map[string]interface{}) PriorityThreshold {
	// First, try to extract explicit priority from event data
	if priority := ExtractPriorityFromEvent(eventData); priority != PriorityLow {
		return priority
	}
	
	// Determine priority based on event type
	switch eventType {
	case "system-error":
		return PriorityUrgent
	case "capacity-alert", "compliance-alert":
		return PriorityHigh
	case "admin-action-required":
		return PriorityMedium
	default:
		// For inquiries and content updates, use medium priority
		if strings.Contains(eventType, "inquiry") {
			return PriorityMedium
		}
		return PriorityLow
	}
}