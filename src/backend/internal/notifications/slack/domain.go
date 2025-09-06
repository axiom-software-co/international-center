package slack

import (
	"fmt"
	"strings"
	"time"
)

// Slack notification domain models

type SlackMessage struct {
	MessageID     string            `json:"message_id"`
	SubscriberID  string            `json:"subscriber_id"`
	Channels      []string          `json:"channels"`
	Content       string            `json:"content"`
	EventType     string            `json:"event_type"`
	Priority      string            `json:"priority"`
	EventData     map[string]interface{} `json:"event_data"`
	Attachments   []SlackAttachment `json:"attachments,omitempty"`
	Blocks        []SlackBlock      `json:"blocks,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	CorrelationID string            `json:"correlation_id"`
}

type SlackAttachment struct {
	Color     string       `json:"color,omitempty"`
	Title     string       `json:"title,omitempty"`
	Text      string       `json:"text,omitempty"`
	Fields    []SlackField `json:"fields,omitempty"`
	Footer    string       `json:"footer,omitempty"`
	Timestamp int64        `json:"ts,omitempty"`
}

type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type SlackBlock struct {
	Type     string                 `json:"type"`
	Text     *SlackTextElement      `json:"text,omitempty"`
	Elements []SlackElement         `json:"elements,omitempty"`
	Fields   []SlackTextElement     `json:"fields,omitempty"`
	Accessory *SlackElement         `json:"accessory,omitempty"`
}

type SlackTextElement struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type SlackElement struct {
	Type     string            `json:"type"`
	Text     *SlackTextElement `json:"text,omitempty"`
	Value    string            `json:"value,omitempty"`
	URL      string            `json:"url,omitempty"`
	ActionID string            `json:"action_id,omitempty"`
}

type SlackDeliveryStatus struct {
	MessageID      string                `json:"message_id"`
	SubscriberID   string                `json:"subscriber_id"`
	Channels       []SlackChannelStatus  `json:"channels"`
	Status         SlackDeliveryStatusType `json:"status"`
	AttemptCount   int                   `json:"attempt_count"`
	LastAttemptAt  time.Time             `json:"last_attempt_at"`
	DeliveredAt    *time.Time            `json:"delivered_at,omitempty"`
	ErrorMessage   *string               `json:"error_message,omitempty"`
	NextRetryAt    *time.Time            `json:"next_retry_at,omitempty"`
}

type SlackChannelStatus struct {
	Channel      string                  `json:"channel"`
	Status       SlackDeliveryStatusType `json:"status"`
	MessageTS    string                  `json:"message_ts,omitempty"`
	DeliveredAt  *time.Time              `json:"delivered_at,omitempty"`
	ErrorMessage *string                 `json:"error_message,omitempty"`
}

type SlackDeliveryStatusType string

const (
	SlackStatusPending   SlackDeliveryStatusType = "pending"
	SlackStatusSent      SlackDeliveryStatusType = "sent"
	SlackStatusDelivered SlackDeliveryStatusType = "delivered"
	SlackStatusFailed    SlackDeliveryStatusType = "failed"
	SlackStatusRateLimit SlackDeliveryStatusType = "rate_limited"
	SlackStatusBlocked   SlackDeliveryStatusType = "blocked"
)

// Slack API configuration
type SlackConfig struct {
	BotToken       string `json:"bot_token"`
	AppToken       string `json:"app_token"`
	DefaultChannel string `json:"default_channel"`
	MaxRetries     int    `json:"max_retries"`
	RetryDelay     int    `json:"retry_delay_seconds"`
	RequestTimeout int    `json:"request_timeout_seconds"`
}

// Slack notification request from notification router
type SlackNotificationRequest struct {
	SubscriberID  string                 `json:"subscriber_id"`
	EventType     string                 `json:"event_type"`
	Priority      string                 `json:"priority"`
	Channels      []string               `json:"channels"`
	EventData     map[string]interface{} `json:"event_data"`
	Schedule      string                 `json:"schedule"`
	CreatedAt     time.Time              `json:"created_at"`
	CorrelationID string                 `json:"correlation_id"`
}

// Slack message character limits
const (
	MaxSlackMessageLength = 4000 // Slack text message limit
	MaxSlackBlockLength   = 3000 // Individual block text limit
	MaxSlackFieldLength   = 2000 // Field value limit
)

// Channel routing by event type
var EventChannelMap = map[string][]string{
	"inquiry-business":       {"#inquiries", "#business"},
	"inquiry-media":          {"#inquiries", "#media"},
	"inquiry-donations":      {"#inquiries", "#donations"},
	"inquiry-volunteers":     {"#inquiries", "#volunteers"},
	"event-registration":     {"#content", "#events"},
	"system-error":           {"#alerts", "#critical"},
	"capacity-alert":         {"#alerts", "#monitoring"},
	"admin-action-required":  {"#admin", "#urgent"},
	"compliance-alert":       {"#compliance", "#alerts"},
}

// Priority colors for Slack attachments
var PriorityColorMap = map[string]string{
	"critical": "#ff0000", // Red
	"high":     "#ff9900", // Orange
	"medium":   "#ffcc00", // Yellow
	"low":      "#36c5f0", // Slack blue
	"info":     "#2eb886", // Green
}

// Slack message validation
func (s *SlackMessage) IsValid() bool {
	return s.MessageID != "" &&
		s.SubscriberID != "" &&
		len(s.Channels) > 0 &&
		s.Content != "" &&
		len(s.Content) <= MaxSlackMessageLength &&
		s.validateChannels()
}

func (s *SlackMessage) validateChannels() bool {
	for _, channel := range s.Channels {
		if !IsValidSlackChannel(channel) {
			return false
		}
	}
	return true
}

func (r *SlackNotificationRequest) IsValid() bool {
	return r.SubscriberID != "" &&
		r.EventType != "" &&
		len(r.Channels) > 0 &&
		r.EventData != nil &&
		r.validateChannels()
}

func (r *SlackNotificationRequest) validateChannels() bool {
	for _, channel := range r.Channels {
		if !IsValidSlackChannel(channel) {
			return false
		}
	}
	return true
}

func (s SlackDeliveryStatusType) IsValid() bool {
	switch s {
	case SlackStatusPending, SlackStatusSent, SlackStatusDelivered,
		SlackStatusFailed, SlackStatusRateLimit, SlackStatusBlocked:
		return true
	default:
		return false
	}
}

func (s SlackDeliveryStatusType) IsFinalStatus() bool {
	return s == SlackStatusDelivered ||
		s == SlackStatusFailed ||
		s == SlackStatusBlocked
}

// Slack channel validation
func IsValidSlackChannel(channel string) bool {
	if channel == "" {
		return false
	}
	
	// Channel should start with # or @ or be a valid channel ID
	if strings.HasPrefix(channel, "#") || strings.HasPrefix(channel, "@") {
		return len(channel) >= 2 && len(channel) <= 80
	}
	
	// Channel ID format (starts with C for channels, D for DMs, G for groups)
	if strings.HasPrefix(channel, "C") || strings.HasPrefix(channel, "D") || strings.HasPrefix(channel, "G") {
		return len(channel) >= 9 && len(channel) <= 11
	}
	
	return false
}

// Generate Slack message content by event type
func GenerateSlackContent(eventType string, eventData map[string]interface{}) string {
	switch eventType {
	case "inquiry-business":
		return generateBusinessInquirySlack(eventData)
	case "inquiry-media":
		return generateMediaInquirySlack(eventData)
	case "inquiry-donations":
		return generateDonationInquirySlack(eventData)
	case "inquiry-volunteers":
		return generateVolunteerInquirySlack(eventData)
	case "event-registration":
		return generateContentPublicationSlack(eventData)
	case "system-error":
		return generateSystemErrorSlack(eventData)
	case "capacity-alert":
		return generateCapacityAlertSlack(eventData)
	case "admin-action-required":
		return generateAdminActionSlack(eventData)
	case "compliance-alert":
		return generateComplianceAlertSlack(eventData)
	default:
		return "New notification alert. Check admin dashboard for details."
	}
}

func generateBusinessInquirySlack(eventData map[string]interface{}) string {
	entityID := extractString(eventData, "entity_id")
	if entityID != "" {
		return fmt.Sprintf("🏢 New business inquiry received: %s\nReview in admin dashboard.", entityID)
	}
	return "🏢 New business inquiry received. Check admin dashboard."
}

func generateMediaInquirySlack(eventData map[string]interface{}) string {
	entityID := extractString(eventData, "entity_id")
	if entityID != "" {
		return fmt.Sprintf("📺 New media inquiry received: %s\nReview in admin dashboard.", entityID)
	}
	return "📺 New media inquiry received. Check admin dashboard."
}

func generateDonationInquirySlack(eventData map[string]interface{}) string {
	entityID := extractString(eventData, "entity_id")
	if entityID != "" {
		return fmt.Sprintf("💰 New donation inquiry received: %s\nReview in admin dashboard.", entityID)
	}
	return "💰 New donation inquiry received. Check admin dashboard."
}

func generateVolunteerInquirySlack(eventData map[string]interface{}) string {
	entityID := extractString(eventData, "entity_id")
	if entityID != "" {
		return fmt.Sprintf("🤝 New volunteer application received: %s\nReview in admin dashboard.", entityID)
	}
	return "🤝 New volunteer application received. Check admin dashboard."
}

func generateContentPublicationSlack(eventData map[string]interface{}) string {
	entityType := extractString(eventData, "entity_type")
	entityID := extractString(eventData, "entity_id")
	
	if entityType != "" && entityID != "" {
		return fmt.Sprintf("📝 New %s content published: %s\nView details in dashboard.", entityType, entityID)
	} else if entityType != "" {
		return fmt.Sprintf("📝 New %s content published. Check dashboard for details.", entityType)
	}
	return "📝 New content published. Check dashboard for details."
}

func generateSystemErrorSlack(eventData map[string]interface{}) string {
	errorType := extractString(eventData, "error_type")
	if errorType != "" {
		return fmt.Sprintf("🚨 URGENT: System error detected (%s)\nImmediate action required!", errorType)
	}
	return "🚨 URGENT: System error detected. Immediate action required!"
}

func generateCapacityAlertSlack(eventData map[string]interface{}) string {
	resourceType := extractString(eventData, "resource_type")
	if resourceType != "" {
		return fmt.Sprintf("⚠️ WARNING: %s capacity alert\nCheck system resources immediately.", resourceType)
	}
	return "⚠️ WARNING: Capacity alert detected. Check system resources."
}

func generateAdminActionSlack(eventData map[string]interface{}) string {
	actionType := extractString(eventData, "action_type")
	if actionType != "" {
		return fmt.Sprintf("👨‍💼 Admin action required: %s\nCheck dashboard for details.", actionType)
	}
	return "👨‍💼 Admin action required. Check dashboard for details."
}

func generateComplianceAlertSlack(eventData map[string]interface{}) string {
	alertType := extractString(eventData, "alert_type")
	if alertType != "" {
		return fmt.Sprintf("⚖️ COMPLIANCE ALERT: %s\nImmediate review required!", alertType)
	}
	return "⚖️ COMPLIANCE ALERT detected. Immediate review required!"
}

// Generate Slack attachment for event data
func GenerateSlackAttachment(eventType string, eventData map[string]interface{}, priority string) SlackAttachment {
	attachment := SlackAttachment{
		Color:     getPriorityColor(priority),
		Footer:    "International Center Notification System",
		Timestamp: time.Now().Unix(),
	}
	
	switch eventType {
	case "inquiry-business", "inquiry-media", "inquiry-donations", "inquiry-volunteers":
		attachment.Title = "Inquiry Details"
		attachment.Fields = []SlackField{
			{Title: "Type", Value: eventType, Short: true},
			{Title: "Priority", Value: priority, Short: true},
			{Title: "Entity ID", Value: extractString(eventData, "entity_id"), Short: true},
		}
	case "event-registration":
		attachment.Title = "Content Publication"
		attachment.Fields = []SlackField{
			{Title: "Content Type", Value: extractString(eventData, "entity_type"), Short: true},
			{Title: "Entity ID", Value: extractString(eventData, "entity_id"), Short: true},
			{Title: "Priority", Value: priority, Short: true},
		}
	case "system-error", "capacity-alert", "admin-action-required", "compliance-alert":
		attachment.Title = "Alert Details"
		attachment.Fields = []SlackField{
			{Title: "Alert Type", Value: eventType, Short: true},
			{Title: "Priority", Value: priority, Short: true},
		}
		
		if errorType := extractString(eventData, "error_type"); errorType != "" {
			attachment.Fields = append(attachment.Fields, SlackField{
				Title: "Error Type", Value: errorType, Short: true,
			})
		}
		
		if resourceType := extractString(eventData, "resource_type"); resourceType != "" {
			attachment.Fields = append(attachment.Fields, SlackField{
				Title: "Resource", Value: resourceType, Short: true,
			})
		}
	}
	
	return attachment
}

// Get channels for event type
func GetChannelsForEventType(eventType string) []string {
	if channels, exists := EventChannelMap[eventType]; exists {
		return channels
	}
	return []string{"#general"}
}

// Get priority color
func getPriorityColor(priority string) string {
	if color, exists := PriorityColorMap[priority]; exists {
		return color
	}
	return PriorityColorMap["info"]
}

// Truncate Slack content to character limit
func TruncateSlackContent(content string, maxLength int) string {
	if len(content) <= maxLength {
		return content
	}
	
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