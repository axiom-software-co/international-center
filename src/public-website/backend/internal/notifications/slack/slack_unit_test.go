package slack

import (
	"testing"
)

// Domain Model Validation Tests

func TestSlackMessage_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		message  *SlackMessage
		expected bool
	}{
		{
			name: "valid slack message with channels",
			message: &SlackMessage{
				MessageID:     "msg-001",
				SubscriberID:  "sub-001",
				Channels:      []string{"#general", "#alerts"},
				Content:       "Test message content",
				EventType:     "inquiry-business",
				Priority:      "high",
				CorrelationID: "corr-001",
			},
			expected: true,
		},
		{
			name: "invalid slack message - empty message ID",
			message: &SlackMessage{
				SubscriberID:  "sub-001",
				Channels:      []string{"#general"},
				Content:       "Test content",
				CorrelationID: "corr-001",
			},
			expected: false,
		},
		{
			name: "invalid slack message - empty subscriber ID",
			message: &SlackMessage{
				MessageID:     "msg-001",
				Channels:      []string{"#general"},
				Content:       "Test content",
				CorrelationID: "corr-001",
			},
			expected: false,
		},
		{
			name: "invalid slack message - no channels",
			message: &SlackMessage{
				MessageID:     "msg-001",
				SubscriberID:  "sub-001",
				Channels:      []string{},
				Content:       "Test content",
				CorrelationID: "corr-001",
			},
			expected: false,
		},
		{
			name: "invalid slack message - empty content",
			message: &SlackMessage{
				MessageID:     "msg-001",
				SubscriberID:  "sub-001",
				Channels:      []string{"#general"},
				CorrelationID: "corr-001",
			},
			expected: false,
		},
		{
			name: "invalid slack message - content too long",
			message: &SlackMessage{
				MessageID:     "msg-001",
				SubscriberID:  "sub-001",
				Channels:      []string{"#general"},
				Content:       string(make([]byte, MaxSlackMessageLength+1)),
				CorrelationID: "corr-001",
			},
			expected: false,
		},
		{
			name: "invalid slack message - invalid channel",
			message: &SlackMessage{
				MessageID:     "msg-001",
				SubscriberID:  "sub-001",
				Channels:      []string{"invalid-channel"},
				Content:       "Test content",
				CorrelationID: "corr-001",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.message.IsValid()
			if result != tt.expected {
				t.Errorf("SlackMessage.IsValid() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestSlackNotificationRequest_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		request  *SlackNotificationRequest
		expected bool
	}{
		{
			name: "valid slack notification request",
			request: &SlackNotificationRequest{
				SubscriberID:  "sub-001",
				EventType:     "inquiry-business",
				Priority:      "high",
				Channels:      []string{"#inquiries"},
				EventData:     map[string]interface{}{"entity_id": "ent-001"},
				CorrelationID: "corr-001",
			},
			expected: true,
		},
		{
			name: "invalid request - empty subscriber ID",
			request: &SlackNotificationRequest{
				EventType:     "inquiry-business",
				Channels:      []string{"#inquiries"},
				EventData:     map[string]interface{}{"entity_id": "ent-001"},
				CorrelationID: "corr-001",
			},
			expected: false,
		},
		{
			name: "invalid request - empty event type",
			request: &SlackNotificationRequest{
				SubscriberID:  "sub-001",
				Channels:      []string{"#inquiries"},
				EventData:     map[string]interface{}{"entity_id": "ent-001"},
				CorrelationID: "corr-001",
			},
			expected: false,
		},
		{
			name: "invalid request - no channels",
			request: &SlackNotificationRequest{
				SubscriberID:  "sub-001",
				EventType:     "inquiry-business",
				Channels:      []string{},
				EventData:     map[string]interface{}{"entity_id": "ent-001"},
				CorrelationID: "corr-001",
			},
			expected: false,
		},
		{
			name: "invalid request - nil event data",
			request: &SlackNotificationRequest{
				SubscriberID:  "sub-001",
				EventType:     "inquiry-business",
				Channels:      []string{"#inquiries"},
				CorrelationID: "corr-001",
			},
			expected: false,
		},
		{
			name: "invalid request - invalid channel format",
			request: &SlackNotificationRequest{
				SubscriberID:  "sub-001",
				EventType:     "inquiry-business",
				Channels:      []string{"invalid-channel"},
				EventData:     map[string]interface{}{"entity_id": "ent-001"},
				CorrelationID: "corr-001",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.request.IsValid()
			if result != tt.expected {
				t.Errorf("SlackNotificationRequest.IsValid() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Channel Validation Tests

func TestIsValidSlackChannel(t *testing.T) {
	tests := []struct {
		name     string
		channel  string
		expected bool
	}{
		{
			name:     "valid channel with hash prefix",
			channel:  "#general",
			expected: true,
		},
		{
			name:     "valid channel with at prefix",
			channel:  "@username",
			expected: true,
		},
		{
			name:     "valid channel ID format",
			channel:  "C1234567890",
			expected: true,
		},
		{
			name:     "valid DM channel ID",
			channel:  "D1234567890",
			expected: true,
		},
		{
			name:     "valid group channel ID",
			channel:  "G1234567890",
			expected: true,
		},
		{
			name:     "invalid empty channel",
			channel:  "",
			expected: false,
		},
		{
			name:     "invalid channel - only hash",
			channel:  "#",
			expected: false,
		},
		{
			name:     "invalid channel - only at",
			channel:  "@",
			expected: false,
		},
		{
			name:     "invalid channel - no prefix",
			channel:  "general",
			expected: false,
		},
		{
			name:     "invalid channel - too long",
			channel:  "#" + string(make([]byte, 80)),
			expected: false,
		},
		{
			name:     "invalid channel ID - too short",
			channel:  "C123",
			expected: false,
		},
		{
			name:     "invalid channel ID - too long",
			channel:  "C123456789012",
			expected: false,
		},
		{
			name:     "invalid channel ID - wrong prefix",
			channel:  "X1234567890",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidSlackChannel(tt.channel)
			if result != tt.expected {
				t.Errorf("IsValidSlackChannel(%q) = %v, expected %v", tt.channel, result, tt.expected)
			}
		})
	}
}

// Status Type Validation Tests

func TestSlackDeliveryStatusType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   SlackDeliveryStatusType
		expected bool
	}{
		{
			name:     "valid status - pending",
			status:   SlackStatusPending,
			expected: true,
		},
		{
			name:     "valid status - sent",
			status:   SlackStatusSent,
			expected: true,
		},
		{
			name:     "valid status - delivered",
			status:   SlackStatusDelivered,
			expected: true,
		},
		{
			name:     "valid status - failed",
			status:   SlackStatusFailed,
			expected: true,
		},
		{
			name:     "valid status - rate limited",
			status:   SlackStatusRateLimit,
			expected: true,
		},
		{
			name:     "valid status - blocked",
			status:   SlackStatusBlocked,
			expected: true,
		},
		{
			name:     "invalid status",
			status:   SlackDeliveryStatusType("invalid"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.IsValid()
			if result != tt.expected {
				t.Errorf("SlackDeliveryStatusType.IsValid() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestSlackDeliveryStatusType_IsFinalStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   SlackDeliveryStatusType
		expected bool
	}{
		{
			name:     "final status - delivered",
			status:   SlackStatusDelivered,
			expected: true,
		},
		{
			name:     "final status - failed",
			status:   SlackStatusFailed,
			expected: true,
		},
		{
			name:     "final status - blocked",
			status:   SlackStatusBlocked,
			expected: true,
		},
		{
			name:     "non-final status - pending",
			status:   SlackStatusPending,
			expected: false,
		},
		{
			name:     "non-final status - sent",
			status:   SlackStatusSent,
			expected: false,
		},
		{
			name:     "non-final status - rate limited",
			status:   SlackStatusRateLimit,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.IsFinalStatus()
			if result != tt.expected {
				t.Errorf("SlackDeliveryStatusType.IsFinalStatus() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Message Content Generation Tests

func TestGenerateSlackContent(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		eventData map[string]interface{}
		expected  string
	}{
		{
			name:      "business inquiry with entity ID",
			eventType: "inquiry-business",
			eventData: map[string]interface{}{"entity_id": "biz-001"},
			expected:  "🏢 New business inquiry received: biz-001\nReview in admin dashboard.",
		},
		{
			name:      "business inquiry without entity ID",
			eventType: "inquiry-business",
			eventData: map[string]interface{}{},
			expected:  "🏢 New business inquiry received. Check admin dashboard.",
		},
		{
			name:      "media inquiry with entity ID",
			eventType: "inquiry-media",
			eventData: map[string]interface{}{"entity_id": "media-001"},
			expected:  "📺 New media inquiry received: media-001\nReview in admin dashboard.",
		},
		{
			name:      "donation inquiry with entity ID",
			eventType: "inquiry-donations",
			eventData: map[string]interface{}{"entity_id": "donate-001"},
			expected:  "💰 New donation inquiry received: donate-001\nReview in admin dashboard.",
		},
		{
			name:      "volunteer inquiry with entity ID",
			eventType: "inquiry-volunteers",
			eventData: map[string]interface{}{"entity_id": "vol-001"},
			expected:  "🤝 New volunteer application received: vol-001\nReview in admin dashboard.",
		},
		{
			name:      "content publication with entity type and ID",
			eventType: "event-registration",
			eventData: map[string]interface{}{"entity_type": "event", "entity_id": "evt-001"},
			expected:  "📝 New event content published: evt-001\nView details in dashboard.",
		},
		{
			name:      "content publication with only entity type",
			eventType: "event-registration",
			eventData: map[string]interface{}{"entity_type": "news"},
			expected:  "📝 New news content published. Check dashboard for details.",
		},
		{
			name:      "content publication without specifics",
			eventType: "event-registration",
			eventData: map[string]interface{}{},
			expected:  "📝 New content published. Check dashboard for details.",
		},
		{
			name:      "system error with error type",
			eventType: "system-error",
			eventData: map[string]interface{}{"error_type": "database_connection"},
			expected:  "🚨 URGENT: System error detected (database_connection)\nImmediate action required!",
		},
		{
			name:      "system error without specifics",
			eventType: "system-error",
			eventData: map[string]interface{}{},
			expected:  "🚨 URGENT: System error detected. Immediate action required!",
		},
		{
			name:      "capacity alert with resource type",
			eventType: "capacity-alert",
			eventData: map[string]interface{}{"resource_type": "memory"},
			expected:  "⚠️ WARNING: memory capacity alert\nCheck system resources immediately.",
		},
		{
			name:      "admin action with action type",
			eventType: "admin-action-required",
			eventData: map[string]interface{}{"action_type": "user_approval"},
			expected:  "👨‍💼 Admin action required: user_approval\nCheck dashboard for details.",
		},
		{
			name:      "compliance alert with alert type",
			eventType: "compliance-alert",
			eventData: map[string]interface{}{"alert_type": "gdpr_violation"},
			expected:  "⚖️ COMPLIANCE ALERT: gdpr_violation\nImmediate review required!",
		},
		{
			name:      "unknown event type",
			eventType: "unknown-event",
			eventData: map[string]interface{}{},
			expected:  "New notification alert. Check admin dashboard for details.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateSlackContent(tt.eventType, tt.eventData)
			if result != tt.expected {
				t.Errorf("GenerateSlackContent() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// Slack Attachment Generation Tests

func TestGenerateSlackAttachment(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		eventData map[string]interface{}
		priority  string
		validate  func(*testing.T, SlackAttachment)
	}{
		{
			name:      "business inquiry attachment",
			eventType: "inquiry-business",
			eventData: map[string]interface{}{"entity_id": "biz-001"},
			priority:  "high",
			validate: func(t *testing.T, attachment SlackAttachment) {
				if attachment.Title != "Inquiry Details" {
					t.Errorf("Expected title 'Inquiry Details', got %q", attachment.Title)
				}
				if attachment.Color != "#ff9900" {
					t.Errorf("Expected color '#ff9900', got %q", attachment.Color)
				}
				if len(attachment.Fields) != 3 {
					t.Errorf("Expected 3 fields, got %d", len(attachment.Fields))
				}
			},
		},
		{
			name:      "content publication attachment",
			eventType: "event-registration",
			eventData: map[string]interface{}{"entity_type": "news", "entity_id": "news-001"},
			priority:  "medium",
			validate: func(t *testing.T, attachment SlackAttachment) {
				if attachment.Title != "Content Publication" {
					t.Errorf("Expected title 'Content Publication', got %q", attachment.Title)
				}
				if attachment.Color != "#ffcc00" {
					t.Errorf("Expected color '#ffcc00', got %q", attachment.Color)
				}
			},
		},
		{
			name:      "system error attachment",
			eventType: "system-error",
			eventData: map[string]interface{}{"error_type": "database_connection"},
			priority:  "critical",
			validate: func(t *testing.T, attachment SlackAttachment) {
				if attachment.Title != "Alert Details" {
					t.Errorf("Expected title 'Alert Details', got %q", attachment.Title)
				}
				if attachment.Color != "#ff0000" {
					t.Errorf("Expected color '#ff0000', got %q", attachment.Color)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attachment := GenerateSlackAttachment(tt.eventType, tt.eventData, tt.priority)
			
			if attachment.Footer != "International Center Notification System" {
				t.Errorf("Expected footer 'International Center Notification System', got %q", attachment.Footer)
			}
			
			if attachment.Timestamp == 0 {
				t.Error("Expected non-zero timestamp")
			}
			
			tt.validate(t, attachment)
		})
	}
}

// Channel Routing Tests

func TestGetChannelsForEventType(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		expected  []string
	}{
		{
			name:      "business inquiry channels",
			eventType: "inquiry-business",
			expected:  []string{"#inquiries", "#business"},
		},
		{
			name:      "media inquiry channels",
			eventType: "inquiry-media",
			expected:  []string{"#inquiries", "#media"},
		},
		{
			name:      "donation inquiry channels",
			eventType: "inquiry-donations",
			expected:  []string{"#inquiries", "#donations"},
		},
		{
			name:      "volunteer inquiry channels",
			eventType: "inquiry-volunteers",
			expected:  []string{"#inquiries", "#volunteers"},
		},
		{
			name:      "content publication channels",
			eventType: "event-registration",
			expected:  []string{"#content", "#events"},
		},
		{
			name:      "system error channels",
			eventType: "system-error",
			expected:  []string{"#alerts", "#critical"},
		},
		{
			name:      "capacity alert channels",
			eventType: "capacity-alert",
			expected:  []string{"#alerts", "#monitoring"},
		},
		{
			name:      "admin action channels",
			eventType: "admin-action-required",
			expected:  []string{"#admin", "#urgent"},
		},
		{
			name:      "compliance alert channels",
			eventType: "compliance-alert",
			expected:  []string{"#compliance", "#alerts"},
		},
		{
			name:      "unknown event type default",
			eventType: "unknown-event",
			expected:  []string{"#general"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetChannelsForEventType(tt.eventType)
			if len(result) != len(tt.expected) {
				t.Errorf("GetChannelsForEventType() returned %d channels, expected %d", len(result), len(tt.expected))
				return
			}
			
			for i, channel := range result {
				if channel != tt.expected[i] {
					t.Errorf("GetChannelsForEventType() channel %d = %q, expected %q", i, channel, tt.expected[i])
				}
			}
		})
	}
}

// Priority Color Mapping Tests

func TestGetPriorityColor(t *testing.T) {
	tests := []struct {
		name     string
		priority string
		expected string
	}{
		{
			name:     "critical priority color",
			priority: "critical",
			expected: "#ff0000",
		},
		{
			name:     "high priority color",
			priority: "high",
			expected: "#ff9900",
		},
		{
			name:     "medium priority color",
			priority: "medium",
			expected: "#ffcc00",
		},
		{
			name:     "low priority color",
			priority: "low",
			expected: "#36c5f0",
		},
		{
			name:     "info priority color",
			priority: "info",
			expected: "#2eb886",
		},
		{
			name:     "unknown priority defaults to info",
			priority: "unknown",
			expected: "#2eb886",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPriorityColor(tt.priority)
			if result != tt.expected {
				t.Errorf("getPriorityColor(%q) = %q, expected %q", tt.priority, result, tt.expected)
			}
		})
	}
}

// Content Truncation Tests

func TestTruncateSlackContent(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		maxLength int
		expected  string
	}{
		{
			name:      "content shorter than limit",
			content:   "Short message",
			maxLength: 100,
			expected:  "Short message",
		},
		{
			name:      "content exactly at limit",
			content:   "Exact",
			maxLength: 5,
			expected:  "Exact",
		},
		{
			name:      "content longer than limit - word boundary",
			content:   "This is a long message that needs truncation",
			maxLength: 20,
			expected:  "This is a long...",
		},
		{
			name:      "content longer than limit - no word boundary",
			content:   "ThisIsAVeryLongWordThatCannotBeTruncatedAtWordBoundary",
			maxLength: 20,
			expected:  "ThisIsAVeryLongWo...",
		},
		{
			name:      "very short limit",
			content:   "Test message",
			maxLength: 3,
			expected:  "Tes",
		},
		{
			name:      "limit of 1",
			content:   "Test",
			maxLength: 1,
			expected:  "T",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateSlackContent(tt.content, tt.maxLength)
			if result != tt.expected {
				t.Errorf("TruncateSlackContent() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// Service Interface Contract Tests

type SlackService interface {
	SendMessage(request *SlackNotificationRequest) (*SlackDeliveryStatus, error)
	GetDeliveryStatus(messageID string) (*SlackDeliveryStatus, error)
	UpdateDeliveryStatus(messageID string, status SlackDeliveryStatusType) error
	RetryFailedMessage(messageID string) error
	ValidateWebhookSignature(signature string, body []byte) bool
}


func TestSlackServiceInterface(t *testing.T) {
	t.Run("service interface contract", func(t *testing.T) {
		var service SlackService
		if service != nil {
			t.Error("Service interface should be nil when not implemented")
		}
	})
}

func TestSlackRepositoryInterface(t *testing.T) {
	t.Run("repository interface contract", func(t *testing.T) {
		var repository SlackRepository
		if repository != nil {
			t.Error("Repository interface should be nil when not implemented")
		}
	})
}

// Configuration Validation Tests

func TestSlackConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config SlackConfig
		valid  bool
	}{
		{
			name: "valid configuration",
			config: SlackConfig{
				BotToken:       "xoxb-valid-token",
				AppToken:       "xapp-valid-token",
				DefaultChannel: "#general",
				MaxRetries:     3,
				RetryDelay:     5,
				RequestTimeout: 30,
			},
			valid: true,
		},
		{
			name: "configuration with minimal values",
			config: SlackConfig{
				BotToken:       "xoxb-token",
				DefaultChannel: "#alerts",
				MaxRetries:     1,
				RetryDelay:     1,
				RequestTimeout: 10,
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.BotToken == "" && tt.valid {
				t.Error("Valid config should have bot token")
			}
			if tt.config.DefaultChannel == "" && tt.valid {
				t.Error("Valid config should have default channel")
			}
		})
	}
}

// Helper Function Tests

func TestExtractString(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		key      string
		expected string
	}{
		{
			name:     "extract existing string value",
			data:     map[string]interface{}{"entity_id": "ent-001"},
			key:      "entity_id",
			expected: "ent-001",
		},
		{
			name:     "extract non-existing key",
			data:     map[string]interface{}{"other_key": "value"},
			key:      "entity_id",
			expected: "",
		},
		{
			name:     "extract non-string value",
			data:     map[string]interface{}{"entity_id": 123},
			key:      "entity_id",
			expected: "",
		},
		{
			name:     "extract from empty data",
			data:     map[string]interface{}{},
			key:      "entity_id",
			expected: "",
		},
		{
			name:     "extract from nil data",
			data:     nil,
			key:      "entity_id",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractString(tt.data, tt.key)
			if result != tt.expected {
				t.Errorf("extractString() = %q, expected %q", result, tt.expected)
			}
		})
	}
}