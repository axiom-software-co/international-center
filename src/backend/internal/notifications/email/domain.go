package email

import (
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
)

// Email notification domain models

type EmailMessage struct {
	MessageID     string            `json:"message_id"`
	SubscriberID  string            `json:"subscriber_id"`
	Recipients    []string          `json:"recipients"`
	Subject       string            `json:"subject"`
	HtmlContent   string            `json:"html_content"`
	TextContent   string            `json:"text_content"`
	EventType     string            `json:"event_type"`
	Priority      string            `json:"priority"`
	EventData     map[string]interface{} `json:"event_data"`
	TemplateID    string            `json:"template_id"`
	PersonalizationData map[string]interface{} `json:"personalization_data"`
	CreatedAt     time.Time         `json:"created_at"`
	CorrelationID string            `json:"correlation_id"`
}

type EmailTemplate struct {
	TemplateID   string            `json:"template_id"`
	EventType    string            `json:"event_type"`
	Subject      string            `json:"subject"`
	HtmlTemplate string            `json:"html_template"`
	TextTemplate string            `json:"text_template"`
	Variables    []string          `json:"variables"`
}

type EmailDeliveryStatus struct {
	MessageID      string                `json:"message_id"`
	SubscriberID   string                `json:"subscriber_id"`
	Recipients     []RecipientStatus     `json:"recipients"`
	Status         DeliveryStatus        `json:"status"`
	AttemptCount   int                   `json:"attempt_count"`
	LastAttemptAt  time.Time             `json:"last_attempt_at"`
	DeliveredAt    *time.Time            `json:"delivered_at,omitempty"`
	ErrorMessage   *string               `json:"error_message,omitempty"`
	NextRetryAt    *time.Time            `json:"next_retry_at,omitempty"`
}

type RecipientStatus struct {
	Email        string         `json:"email"`
	Status       DeliveryStatus `json:"status"`
	DeliveredAt  *time.Time     `json:"delivered_at,omitempty"`
	ErrorMessage *string        `json:"error_message,omitempty"`
}

type DeliveryStatus string

const (
	DeliveryStatusPending   DeliveryStatus = "pending"
	DeliveryStatusSent      DeliveryStatus = "sent"
	DeliveryStatusDelivered DeliveryStatus = "delivered"
	DeliveryStatusFailed    DeliveryStatus = "failed"
	DeliveryStatusBounced   DeliveryStatus = "bounced"
	DeliveryStatusSpam      DeliveryStatus = "spam"
)

// Azure Communication Service configuration
type AzureEmailConfig struct {
	ConnectionString string `json:"connection_string"`
	SenderAddress    string `json:"sender_address"`
	SenderName       string `json:"sender_name"`
	ReplyToAddress   string `json:"reply_to_address"`
	MaxRetries       int    `json:"max_retries"`
	RetryDelay       int    `json:"retry_delay_seconds"`
	RequestTimeout   int    `json:"request_timeout_seconds"`
}

// Email notification request from notification router
type EmailNotificationRequest struct {
	SubscriberID  string                 `json:"subscriber_id"`
	EventType     string                 `json:"event_type"`
	Priority      string                 `json:"priority"`
	Recipients    []string               `json:"recipients"`
	EventData     map[string]interface{} `json:"event_data"`
	Schedule      string                 `json:"schedule"`
	CreatedAt     time.Time              `json:"created_at"`
	CorrelationID string                 `json:"correlation_id"`
}

// Template rendering data
type EmailTemplateData struct {
	SubscriberName    string                 `json:"subscriber_name"`
	EventType         string                 `json:"event_type"`
	Priority          string                 `json:"priority"`
	EventDescription  string                 `json:"event_description"`
	EntityID          string                 `json:"entity_id"`
	UserID            string                 `json:"user_id"`
	Timestamp         string                 `json:"timestamp"`
	CorrelationID     string                 `json:"correlation_id"`
	EventData         map[string]interface{} `json:"event_data"`
	ActionURL         string                 `json:"action_url"`
	UnsubscribeURL    string                 `json:"unsubscribe_url"`
}

// Email validation and formatting
func (e *EmailMessage) IsValid() bool {
	return e.MessageID != "" &&
		e.SubscriberID != "" &&
		len(e.Recipients) > 0 &&
		e.Subject != "" &&
		(e.HtmlContent != "" || e.TextContent != "")
}

func (r *EmailNotificationRequest) IsValid() bool {
	return r.SubscriberID != "" &&
		r.EventType != "" &&
		len(r.Recipients) > 0 &&
		r.EventData != nil
}

func (s DeliveryStatus) IsValid() bool {
	switch s {
	case DeliveryStatusPending, DeliveryStatusSent, DeliveryStatusDelivered,
		DeliveryStatusFailed, DeliveryStatusBounced, DeliveryStatusSpam:
		return true
	default:
		return false
	}
}

func (s DeliveryStatus) IsFinalStatus() bool {
	return s == DeliveryStatusDelivered ||
		s == DeliveryStatusFailed ||
		s == DeliveryStatusBounced ||
		s == DeliveryStatusSpam
}

// Email template mappings by event type
func GetTemplateIDByEventType(eventType string) string {
	templateMap := map[string]string{
		"inquiry-business":         "business-inquiry-template",
		"inquiry-media":           "media-inquiry-template",
		"inquiry-donations":       "donation-inquiry-template",
		"inquiry-volunteers":      "volunteer-inquiry-template",
		"event-registration":      "content-publication-template",
		"system-error":            "system-alert-template",
		"capacity-alert":          "capacity-warning-template",
		"admin-action-required":   "admin-action-template",
		"compliance-alert":        "compliance-alert-template",
	}
	
	if templateID, exists := templateMap[eventType]; exists {
		return templateID
	}
	return "default-notification-template"
}

// Subject line generation by event type
func GenerateSubjectByEventType(eventType string, eventData map[string]interface{}) string {
	switch eventType {
	case "inquiry-business":
		return "New Business Inquiry Received"
	case "inquiry-media":
		return "New Media Inquiry Received"
	case "inquiry-donations":
		return "New Donation Inquiry Received"
	case "inquiry-volunteers":
		return "New Volunteer Application Received"
	case "event-registration":
		if entityType, exists := eventData["entity_type"]; exists {
			return "New " + entityType.(string) + " Content Published"
		}
		return "New Content Published"
	case "system-error":
		return "System Alert: Error Detected"
	case "capacity-alert":
		return "Capacity Warning Alert"
	case "admin-action-required":
		return "Admin Action Required"
	case "compliance-alert":
		return "Compliance Alert - Review Required"
	default:
		return "Notification Alert"
	}
}