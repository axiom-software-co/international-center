package sms

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
)

// SMS notification domain models

type SMSMessage struct {
	MessageID     string            `json:"message_id"`
	SubscriberID  string            `json:"subscriber_id"`
	Recipients    []string          `json:"recipients"`
	Content       string            `json:"content"`
	EventType     string            `json:"event_type"`
	Priority      string            `json:"priority"`
	EventData     map[string]interface{} `json:"event_data"`
	CreatedAt     time.Time         `json:"created_at"`
	CorrelationID string            `json:"correlation_id"`
}

type SMSDeliveryStatus struct {
	MessageID      string                `json:"message_id"`
	SubscriberID   string                `json:"subscriber_id"`
	Recipients     []SMSRecipientStatus  `json:"recipients"`
	Status         SMSDeliveryStatus     `json:"status"`
	AttemptCount   int                   `json:"attempt_count"`
	LastAttemptAt  time.Time             `json:"last_attempt_at"`
	DeliveredAt    *time.Time            `json:"delivered_at,omitempty"`
	ErrorMessage   *string               `json:"error_message,omitempty"`
	NextRetryAt    *time.Time            `json:"next_retry_at,omitempty"`
}

type SMSRecipientStatus struct {
	PhoneNumber  string                `json:"phone_number"`
	Status       SMSDeliveryStatusType `json:"status"`
	DeliveredAt  *time.Time            `json:"delivered_at,omitempty"`
	ErrorMessage *string               `json:"error_message,omitempty"`
}

type SMSDeliveryStatusType string

const (
	SMSStatusPending   SMSDeliveryStatusType = "pending"
	SMSStatusSent      SMSDeliveryStatusType = "sent"
	SMSStatusDelivered SMSDeliveryStatusType = "delivered"
	SMSStatusFailed    SMSDeliveryStatusType = "failed"
	SMSStatusBlocked   SMSDeliveryStatusType = "blocked"
	SMSStatusOptedOut  SMSDeliveryStatusType = "opted_out"
)

// Azure Communication Service SMS configuration
type AzureSMSConfig struct {
	ConnectionString string `json:"connection_string"`
	FromNumber       string `json:"from_number"`
	MaxRetries       int    `json:"max_retries"`
	RetryDelay       int    `json:"retry_delay_seconds"`
	RequestTimeout   int    `json:"request_timeout_seconds"`
}

// SMS notification request from notification router
type SMSNotificationRequest struct {
	SubscriberID  string                 `json:"subscriber_id"`
	EventType     string                 `json:"event_type"`
	Priority      string                 `json:"priority"`
	Recipients    []string               `json:"recipients"`
	EventData     map[string]interface{} `json:"event_data"`
	Schedule      string                 `json:"schedule"`
	CreatedAt     time.Time              `json:"created_at"`
	CorrelationID string                 `json:"correlation_id"`
}

// SMS character limits and formatting constraints
const (
	MaxSMSLength         = 160 // Standard SMS character limit
	MaxConcatSMSLength   = 1600 // Maximum for concatenated SMS
	MaxPhoneNumberLength = 15   // E.164 format maximum
	MinPhoneNumberLength = 10   // US format minimum
)

// SMS message validation and formatting
func (s *SMSMessage) IsValid() bool {
	return s.MessageID != "" &&
		s.SubscriberID != "" &&
		len(s.Recipients) > 0 &&
		s.Content != "" &&
		len(s.Content) <= MaxConcatSMSLength &&
		s.validatePhoneNumbers()
}

func (s *SMSMessage) validatePhoneNumbers() bool {
	for _, phone := range s.Recipients {
		if !IsValidUSPhoneNumber(phone) {
			return false
		}
	}
	return true
}

func (r *SMSNotificationRequest) IsValid() bool {
	return r.SubscriberID != "" &&
		r.EventType != "" &&
		len(r.Recipients) > 0 &&
		r.EventData != nil &&
		r.validatePhoneNumbers()
}

func (r *SMSNotificationRequest) validatePhoneNumbers() bool {
	for _, phone := range r.Recipients {
		if !IsValidUSPhoneNumber(phone) {
			return false
		}
	}
	return true
}

func (s SMSDeliveryStatusType) IsValid() bool {
	switch s {
	case SMSStatusPending, SMSStatusSent, SMSStatusDelivered,
		SMSStatusFailed, SMSStatusBlocked, SMSStatusOptedOut:
		return true
	default:
		return false
	}
}

func (s SMSDeliveryStatusType) IsFinalStatus() bool {
	return s == SMSStatusDelivered ||
		s == SMSStatusFailed ||
		s == SMSStatusBlocked ||
		s == SMSStatusOptedOut
}

// Phone number validation for US format
func IsValidUSPhoneNumber(phone string) bool {
	// Remove non-numeric characters
	cleaned := regexp.MustCompile(`[^\d]`).ReplaceAllString(phone, "")
	
	// Check length
	if len(cleaned) < MinPhoneNumberLength || len(cleaned) > MaxPhoneNumberLength {
		return false
	}
	
	// US phone numbers: 10 digits (with optional country code +1)
	if len(cleaned) == 10 {
		// Validate US area code (first digit 2-9, second digit 0-9)
		if cleaned[0] < '2' || cleaned[0] > '9' {
			return false
		}
		return true
	}
	
	if len(cleaned) == 11 && cleaned[0] == '1' {
		// US number with country code, validate area code
		if cleaned[1] < '2' || cleaned[1] > '9' {
			return false
		}
		return true
	}
	
	return false
}

// Format phone number to E.164 format
func FormatPhoneNumberE164(phone string) string {
	cleaned := regexp.MustCompile(`[^\d]`).ReplaceAllString(phone, "")
	
	if len(cleaned) == 10 {
		return "+1" + cleaned
	}
	
	if len(cleaned) == 11 && cleaned[0] == '1' {
		return "+" + cleaned
	}
	
	return phone // Return original if formatting fails
}

// Generate SMS content by event type with character limit compliance
func GenerateSMSContent(eventType string, eventData map[string]interface{}) string {
	switch eventType {
	case "inquiry-business":
		return generateBusinessInquirySMS(eventData)
	case "inquiry-media":
		return generateMediaInquirySMS(eventData)
	case "inquiry-donations":
		return generateDonationInquirySMS(eventData)
	case "inquiry-volunteers":
		return generateVolunteerInquirySMS(eventData)
	case "event-registration":
		return generateContentPublicationSMS(eventData)
	case "system-error":
		return generateSystemErrorSMS(eventData)
	case "capacity-alert":
		return generateCapacityAlertSMS(eventData)
	case "admin-action-required":
		return generateAdminActionSMS(eventData)
	case "compliance-alert":
		return generateComplianceAlertSMS(eventData)
	default:
		return "New notification alert. Check admin dashboard for details."
	}
}

func generateBusinessInquirySMS(eventData map[string]interface{}) string {
	entityID := extractString(eventData, "entity_id")
	if entityID != "" {
		return fmt.Sprintf("New business inquiry %s received. Review in admin dashboard.", entityID)
	}
	return "New business inquiry received. Check admin dashboard."
}

func generateMediaInquirySMS(eventData map[string]interface{}) string {
	entityID := extractString(eventData, "entity_id")
	if entityID != "" {
		return fmt.Sprintf("New media inquiry %s received. Review in admin dashboard.", entityID)
	}
	return "New media inquiry received. Check admin dashboard."
}

func generateDonationInquirySMS(eventData map[string]interface{}) string {
	entityID := extractString(eventData, "entity_id")
	if entityID != "" {
		return fmt.Sprintf("New donation inquiry %s received. Review in admin dashboard.", entityID)
	}
	return "New donation inquiry received. Check admin dashboard."
}

func generateVolunteerInquirySMS(eventData map[string]interface{}) string {
	entityID := extractString(eventData, "entity_id")
	if entityID != "" {
		return fmt.Sprintf("New volunteer application %s received. Review in admin dashboard.", entityID)
	}
	return "New volunteer application received. Check admin dashboard."
}

func generateContentPublicationSMS(eventData map[string]interface{}) string {
	entityType := extractString(eventData, "entity_type")
	entityID := extractString(eventData, "entity_id")
	
	if entityType != "" && entityID != "" {
		return fmt.Sprintf("New %s content %s published. View details in dashboard.", entityType, entityID)
	} else if entityType != "" {
		return fmt.Sprintf("New %s content published. Check dashboard for details.", entityType)
	}
	return "New content published. Check dashboard for details."
}

func generateSystemErrorSMS(eventData map[string]interface{}) string {
	errorType := extractString(eventData, "error_type")
	if errorType != "" {
		return fmt.Sprintf("URGENT: System error (%s) detected. Immediate action required.", errorType)
	}
	return "URGENT: System error detected. Immediate action required."
}

func generateCapacityAlertSMS(eventData map[string]interface{}) string {
	resourceType := extractString(eventData, "resource_type")
	if resourceType != "" {
		return fmt.Sprintf("WARNING: %s capacity alert. Check system resources immediately.", resourceType)
	}
	return "WARNING: Capacity alert detected. Check system resources."
}

func generateAdminActionSMS(eventData map[string]interface{}) string {
	actionType := extractString(eventData, "action_type")
	if actionType != "" {
		return fmt.Sprintf("Admin action required: %s. Check dashboard for details.", actionType)
	}
	return "Admin action required. Check dashboard for details."
}

func generateComplianceAlertSMS(eventData map[string]interface{}) string {
	alertType := extractString(eventData, "alert_type")
	if alertType != "" {
		return fmt.Sprintf("COMPLIANCE: %s alert. Immediate review required.", alertType)
	}
	return "COMPLIANCE: Alert detected. Immediate review required."
}

// Truncate SMS content to character limit
func TruncateSMSContent(content string, maxLength int) string {
	if len(content) <= maxLength {
		return content
	}
	
	// Try to truncate at word boundary
	if maxLength > 3 {
		truncated := content[:maxLength-3]
		if lastSpace := strings.LastIndex(truncated, " "); lastSpace > maxLength/2 {
			return truncated[:lastSpace] + "..."
		}
		return truncated + "..."
	}
	
	return content[:maxLength]
}

// Helper function to extract string value from event data
func extractString(data map[string]interface{}, key string) string {
	if value, exists := data[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}