package sms

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	sharedtesting "github.com/axiom-software-co/international-center/src/backend/internal/shared/testing"
	"github.com/stretchr/testify/assert"
)

// RED PHASE - SMS Handler Tests (35+ test cases)

func TestIsValidUSPhoneNumber(t *testing.T) {
	tests := []struct {
		name     string
		phone    string
		expected bool
	}{
		{
			name:     "valid 10-digit US phone number",
			phone:    "2125551234",
			expected: true,
		},
		{
			name:     "valid 11-digit US phone number with country code",
			phone:    "12125551234",
			expected: true,
		},
		{
			name:     "valid phone number with formatting",
			phone:    "(212) 555-1234",
			expected: true,
		},
		{
			name:     "valid phone number with dashes",
			phone:    "212-555-1234",
			expected: true,
		},
		{
			name:     "valid phone number with dots",
			phone:    "212.555.1234",
			expected: true,
		},
		{
			name:     "valid phone number with spaces",
			phone:    "212 555 1234",
			expected: true,
		},
		{
			name:     "valid phone number with +1 country code",
			phone:    "+1 212 555 1234",
			expected: true,
		},
		{
			name:     "invalid phone number - too short",
			phone:    "555123",
			expected: false,
		},
		{
			name:     "invalid phone number - too long",
			phone:    "123456789012345",
			expected: false,
		},
		{
			name:     "invalid phone number - invalid area code starting with 0",
			phone:    "0125551234",
			expected: false,
		},
		{
			name:     "invalid phone number - invalid area code starting with 1",
			phone:    "1125551234",
			expected: false,
		},
		{
			name:     "invalid phone number - non-numeric characters",
			phone:    "212abcd1234",
			expected: false,
		},
		{
			name:     "invalid phone number - empty string",
			phone:    "",
			expected: false,
		},
		{
			name:     "invalid 11-digit number not starting with 1",
			phone:    "22125551234",
			expected: false,
		},
		{
			name:     "valid edge case - area code starting with 9",
			phone:    "9125551234",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			// Act
			result := IsValidUSPhoneNumber(tt.phone)

			// Assert
			assert.Equal(t, tt.expected, result)

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestFormatPhoneNumberE164(t *testing.T) {
	tests := []struct {
		name     string
		phone    string
		expected string
	}{
		{
			name:     "format 10-digit number to E164",
			phone:    "2125551234",
			expected: "+12125551234",
		},
		{
			name:     "format number with formatting to E164",
			phone:    "(212) 555-1234",
			expected: "+12125551234",
		},
		{
			name:     "format 11-digit number to E164",
			phone:    "12125551234",
			expected: "+12125551234",
		},
		{
			name:     "format number with +1 already present",
			phone:    "+1 212 555 1234",
			expected: "+12125551234",
		},
		{
			name:     "format number with dashes",
			phone:    "212-555-1234",
			expected: "+12125551234",
		},
		{
			name:     "return original for invalid number",
			phone:    "invalid",
			expected: "invalid",
		},
		{
			name:     "return original for too short number",
			phone:    "555123",
			expected: "555123",
		},
		{
			name:     "return original for too long number",
			phone:    "123456789012345",
			expected: "123456789012345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			// Act
			result := FormatPhoneNumberE164(tt.phone)

			// Assert
			assert.Equal(t, tt.expected, result)

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestSMSMessage_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		message  *SMSMessage
		expected bool
	}{
		{
			name: "valid SMS message",
			message: &SMSMessage{
				MessageID:     "sms-001",
				SubscriberID:  "sub-001",
				Recipients:    []string{"2125551234"},
				Content:       "Test SMS message",
				EventType:     "inquiry-business",
				Priority:      "high",
				CreatedAt:     time.Now(),
				CorrelationID: "corr-001",
			},
			expected: true,
		},
		{
			name: "valid SMS message with multiple recipients",
			message: &SMSMessage{
				MessageID:     "sms-002",
				SubscriberID:  "sub-002",
				Recipients:    []string{"2125551234", "3125559876"},
				Content:       "Test SMS message with multiple recipients",
				EventType:     "inquiry-media",
				Priority:      "medium",
				CreatedAt:     time.Now(),
				CorrelationID: "corr-002",
			},
			expected: true,
		},
		{
			name: "valid SMS message at character limit",
			message: &SMSMessage{
				MessageID:     "sms-003",
				SubscriberID:  "sub-003",
				Recipients:    []string{"2125551234"},
				Content:       strings.Repeat("a", MaxSMSLength),
				EventType:     "inquiry-donations",
				Priority:      "low",
				CreatedAt:     time.Now(),
				CorrelationID: "corr-003",
			},
			expected: true,
		},
		{
			name: "invalid SMS message - missing message ID",
			message: &SMSMessage{
				MessageID:     "",
				SubscriberID:  "sub-004",
				Recipients:    []string{"2125551234"},
				Content:       "Test message",
				EventType:     "inquiry-volunteers",
				CorrelationID: "corr-004",
			},
			expected: false,
		},
		{
			name: "invalid SMS message - missing subscriber ID",
			message: &SMSMessage{
				MessageID:     "sms-005",
				SubscriberID:  "",
				Recipients:    []string{"2125551234"},
				Content:       "Test message",
				EventType:     "event-registration",
				CorrelationID: "corr-005",
			},
			expected: false,
		},
		{
			name: "invalid SMS message - empty recipients",
			message: &SMSMessage{
				MessageID:     "sms-006",
				SubscriberID:  "sub-006",
				Recipients:    []string{},
				Content:       "Test message",
				EventType:     "system-error",
				CorrelationID: "corr-006",
			},
			expected: false,
		},
		{
			name: "invalid SMS message - empty content",
			message: &SMSMessage{
				MessageID:     "sms-007",
				SubscriberID:  "sub-007",
				Recipients:    []string{"2125551234"},
				Content:       "",
				EventType:     "capacity-alert",
				CorrelationID: "corr-007",
			},
			expected: false,
		},
		{
			name: "invalid SMS message - content too long",
			message: &SMSMessage{
				MessageID:     "sms-008",
				SubscriberID:  "sub-008",
				Recipients:    []string{"2125551234"},
				Content:       strings.Repeat("a", MaxConcatSMSLength+1),
				EventType:     "admin-action-required",
				CorrelationID: "corr-008",
			},
			expected: false,
		},
		{
			name: "invalid SMS message - invalid phone number",
			message: &SMSMessage{
				MessageID:     "sms-009",
				SubscriberID:  "sub-009",
				Recipients:    []string{"invalid-phone"},
				Content:       "Test message",
				EventType:     "compliance-alert",
				CorrelationID: "corr-009",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			// Act
			result := tt.message.IsValid()

			// Assert
			assert.Equal(t, tt.expected, result)

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestSMSNotificationRequest_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		request  *SMSNotificationRequest
		expected bool
	}{
		{
			name: "valid SMS notification request",
			request: &SMSNotificationRequest{
				SubscriberID:  "sub-001",
				EventType:     "inquiry-business",
				Priority:      "high",
				Recipients:    []string{"2125551234"},
				EventData:     map[string]interface{}{"entity_id": "biz-001"},
				Schedule:      "immediate",
				CreatedAt:     time.Now(),
				CorrelationID: "corr-001",
			},
			expected: true,
		},
		{
			name: "valid request with multiple phone numbers",
			request: &SMSNotificationRequest{
				SubscriberID:  "sub-002",
				EventType:     "event-registration",
				Priority:      "medium",
				Recipients:    []string{"2125551234", "3125559876"},
				EventData:     map[string]interface{}{"entity_id": "event-001"},
				Schedule:      "immediate",
				CreatedAt:     time.Now(),
				CorrelationID: "corr-002",
			},
			expected: true,
		},
		{
			name: "invalid request - missing subscriber ID",
			request: &SMSNotificationRequest{
				SubscriberID:  "",
				EventType:     "inquiry-media",
				Priority:      "low",
				Recipients:    []string{"2125551234"},
				EventData:     map[string]interface{}{"entity_id": "media-001"},
				CorrelationID: "corr-003",
			},
			expected: false,
		},
		{
			name: "invalid request - missing event type",
			request: &SMSNotificationRequest{
				SubscriberID:  "sub-004",
				EventType:     "",
				Priority:      "high",
				Recipients:    []string{"2125551234"},
				EventData:     map[string]interface{}{"entity_id": "donation-001"},
				CorrelationID: "corr-004",
			},
			expected: false,
		},
		{
			name: "invalid request - empty recipients",
			request: &SMSNotificationRequest{
				SubscriberID:  "sub-005",
				EventType:     "inquiry-volunteers",
				Priority:      "medium",
				Recipients:    []string{},
				EventData:     map[string]interface{}{"entity_id": "volunteer-001"},
				CorrelationID: "corr-005",
			},
			expected: false,
		},
		{
			name: "invalid request - nil event data",
			request: &SMSNotificationRequest{
				SubscriberID:  "sub-006",
				EventType:     "system-error",
				Priority:      "urgent",
				Recipients:    []string{"2125551234"},
				EventData:     nil,
				CorrelationID: "corr-006",
			},
			expected: false,
		},
		{
			name: "invalid request - invalid phone number",
			request: &SMSNotificationRequest{
				SubscriberID:  "sub-007",
				EventType:     "capacity-alert",
				Priority:      "high",
				Recipients:    []string{"invalid-phone"},
				EventData:     map[string]interface{}{"entity_id": "capacity-001"},
				CorrelationID: "corr-007",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			// Act
			result := tt.request.IsValid()

			// Assert
			assert.Equal(t, tt.expected, result)

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestSMSDeliveryStatusType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   SMSDeliveryStatusType
		expected bool
	}{
		{
			name:     "valid pending status",
			status:   SMSStatusPending,
			expected: true,
		},
		{
			name:     "valid sent status",
			status:   SMSStatusSent,
			expected: true,
		},
		{
			name:     "valid delivered status",
			status:   SMSStatusDelivered,
			expected: true,
		},
		{
			name:     "valid failed status",
			status:   SMSStatusFailed,
			expected: true,
		},
		{
			name:     "valid blocked status",
			status:   SMSStatusBlocked,
			expected: true,
		},
		{
			name:     "valid opted out status",
			status:   SMSStatusOptedOut,
			expected: true,
		},
		{
			name:     "invalid status",
			status:   SMSDeliveryStatusType("invalid"),
			expected: false,
		},
		{
			name:     "empty status",
			status:   SMSDeliveryStatusType(""),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			// Act
			result := tt.status.IsValid()

			// Assert
			assert.Equal(t, tt.expected, result)

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestSMSDeliveryStatusType_IsFinalStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   SMSDeliveryStatusType
		expected bool
	}{
		{
			name:     "delivered is final status",
			status:   SMSStatusDelivered,
			expected: true,
		},
		{
			name:     "failed is final status",
			status:   SMSStatusFailed,
			expected: true,
		},
		{
			name:     "blocked is final status",
			status:   SMSStatusBlocked,
			expected: true,
		},
		{
			name:     "opted out is final status",
			status:   SMSStatusOptedOut,
			expected: true,
		},
		{
			name:     "pending is not final status",
			status:   SMSStatusPending,
			expected: false,
		},
		{
			name:     "sent is not final status",
			status:   SMSStatusSent,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			// Act
			result := tt.status.IsFinalStatus()

			// Assert
			assert.Equal(t, tt.expected, result)

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestGenerateSMSContent(t *testing.T) {
	tests := []struct {
		name              string
		eventType         string
		eventData         map[string]interface{}
		expectedContains  []string
		maxLength         int
	}{
		{
			name:      "generate business inquiry SMS",
			eventType: "inquiry-business",
			eventData: map[string]interface{}{"entity_id": "biz-001"},
			expectedContains: []string{"business inquiry", "biz-001", "admin dashboard"},
			maxLength: MaxSMSLength,
		},
		{
			name:      "generate media inquiry SMS",
			eventType: "inquiry-media",
			eventData: map[string]interface{}{"entity_id": "media-001"},
			expectedContains: []string{"media inquiry", "media-001", "admin dashboard"},
			maxLength: MaxSMSLength,
		},
		{
			name:      "generate donation inquiry SMS",
			eventType: "inquiry-donations",
			eventData: map[string]interface{}{"entity_id": "donation-001"},
			expectedContains: []string{"donation inquiry", "donation-001", "admin dashboard"},
			maxLength: MaxSMSLength,
		},
		{
			name:      "generate volunteer inquiry SMS",
			eventType: "inquiry-volunteers",
			eventData: map[string]interface{}{"entity_id": "volunteer-001"},
			expectedContains: []string{"volunteer application", "volunteer-001", "admin dashboard"},
			maxLength: MaxSMSLength,
		},
		{
			name:      "generate content publication SMS with entity type",
			eventType: "event-registration",
			eventData: map[string]interface{}{"entity_type": "research", "entity_id": "research-001"},
			expectedContains: []string{"research content", "research-001", "published"},
			maxLength: MaxSMSLength,
		},
		{
			name:      "generate content publication SMS without entity ID",
			eventType: "event-registration",
			eventData: map[string]interface{}{"entity_type": "news"},
			expectedContains: []string{"news content", "published", "dashboard"},
			maxLength: MaxSMSLength,
		},
		{
			name:      "generate system error SMS",
			eventType: "system-error",
			eventData: map[string]interface{}{"error_type": "database_connection"},
			expectedContains: []string{"URGENT", "System error", "database_connection", "action required"},
			maxLength: MaxSMSLength,
		},
		{
			name:      "generate capacity alert SMS",
			eventType: "capacity-alert",
			eventData: map[string]interface{}{"resource_type": "memory"},
			expectedContains: []string{"WARNING", "memory capacity", "check system"},
			maxLength: MaxSMSLength,
		},
		{
			name:      "generate admin action SMS",
			eventType: "admin-action-required",
			eventData: map[string]interface{}{"action_type": "approval_needed"},
			expectedContains: []string{"Admin action required", "approval_needed", "dashboard"},
			maxLength: MaxSMSLength,
		},
		{
			name:      "generate compliance alert SMS",
			eventType: "compliance-alert",
			eventData: map[string]interface{}{"alert_type": "audit_due"},
			expectedContains: []string{"COMPLIANCE", "audit_due", "review required"},
			maxLength: MaxSMSLength,
		},
		{
			name:      "generate default SMS for unknown event type",
			eventType: "unknown-event",
			eventData: map[string]interface{}{},
			expectedContains: []string{"notification alert", "admin dashboard"},
			maxLength: MaxSMSLength,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			// Act
			result := GenerateSMSContent(tt.eventType, tt.eventData)

			// Assert
			assert.LessOrEqual(t, len(result), tt.maxLength, "SMS content should not exceed maximum length")
			
			for _, expectedText := range tt.expectedContains {
				assert.Contains(t, strings.ToLower(result), strings.ToLower(expectedText),
					"SMS content should contain expected text: %s", expectedText)
			}

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestTruncateSMSContent(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		maxLength int
		expected  string
	}{
		{
			name:      "content within limit unchanged",
			content:   "Short message",
			maxLength: MaxSMSLength,
			expected:  "Short message",
		},
		{
			name:      "content exactly at limit unchanged",
			content:   strings.Repeat("a", MaxSMSLength),
			maxLength: MaxSMSLength,
			expected:  strings.Repeat("a", MaxSMSLength),
		},
		{
			name:      "long content truncated with ellipsis",
			content:   "This is a very long SMS message that needs to be truncated because it exceeds the character limit",
			maxLength: 50,
			expected:  "This is a very long SMS message that needs...",
		},
		{
			name:      "truncate at word boundary when possible",
			content:   "This is a test message with multiple words that should be truncated",
			maxLength: 30,
			expected:  "This is a test message...",
		},
		{
			name:      "handle very short max length",
			content:   "Test message",
			maxLength: 5,
			expected:  "Te...",
		},
		{
			name:      "handle max length of 3 or less",
			content:   "Test message",
			maxLength: 3,
			expected:   "Tes",
		},
		{
			name:      "handle empty content",
			content:   "",
			maxLength: MaxSMSLength,
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			// Act
			result := TruncateSMSContent(tt.content, tt.maxLength)

			// Assert
			assert.LessOrEqual(t, len(result), tt.maxLength)
			assert.Equal(t, tt.expected, result)

			_ = ctx // Use context to avoid linting issues
		})
	}
}

// SMSRepositoryInterface for testing
type SMSRepositoryInterface interface {
	// RabbitMQ Operations
	SubscribeToSMSRequests(ctx context.Context, handler func(context.Context, *SMSNotificationRequest) error) error
	PublishToDeadLetterQueue(ctx context.Context, request *SMSNotificationRequest, errorMsg string) error
	AcknowledgeMessage(ctx context.Context, messageID string) error
	RejectMessage(ctx context.Context, messageID string, requeue bool) error

	// Azure Communication Service Operations
	SendSMS(ctx context.Context, message *SMSMessage) (*SMSDeliveryStatus, error)
	GetDeliveryStatus(ctx context.Context, messageID string) (*SMSDeliveryStatus, error)
	
	// Configuration Operations
	GetAzureSMSConfig(ctx context.Context) (*AzureSMSConfig, error)
	
	// Retry and Tracking Operations
	SaveDeliveryStatus(ctx context.Context, status *SMSDeliveryStatus) error
	GetFailedDeliveries(ctx context.Context, limit int) ([]*SMSDeliveryStatus, error)
	ScheduleRetry(ctx context.Context, messageID string, retryAt time.Time) error

	// Audit Operations
	PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error
}

// SMSServiceInterface for testing
type SMSServiceInterface interface {
	// Message Processing
	ProcessSMSNotificationRequest(ctx context.Context, request *SMSNotificationRequest) error
	FormatSMSMessage(ctx context.Context, request *SMSNotificationRequest) (*SMSMessage, error)
	SendSMSMessage(ctx context.Context, message *SMSMessage) error

	// Content Generation
	GenerateMessageContent(ctx context.Context, eventType string, eventData map[string]interface{}) string
	ValidatePhoneNumbers(ctx context.Context, phoneNumbers []string) ([]string, []string) // valid, invalid

	// Delivery Management
	HandleDeliveryFailure(ctx context.Context, message *SMSMessage, errorMsg string) error
	RetryFailedDelivery(ctx context.Context, messageID string) error
	ProcessDeliveryCallback(ctx context.Context, messageID string, status SMSDeliveryStatusType, errorMsg *string) error

	// Health and Status
	HealthCheck(ctx context.Context) error
	GetDeliveryMetrics(ctx context.Context) (map[string]interface{}, error)
}

func TestSMSService_ProcessSMSNotificationRequest(t *testing.T) {
	tests := []struct {
		name         string
		request      *SMSNotificationRequest
		mockConfig   *AzureSMSConfig
		expectError  bool
		expectedContent []string
	}{
		{
			name: "process business inquiry SMS successfully",
			request: &SMSNotificationRequest{
				SubscriberID:  "sub-001",
				EventType:     "inquiry-business",
				Priority:      "high",
				Recipients:    []string{"2125551234"},
				EventData:     map[string]interface{}{"entity_id": "biz-001"},
				Schedule:      "immediate",
				CreatedAt:     time.Now(),
				CorrelationID: "corr-001",
			},
			mockConfig: &AzureSMSConfig{
				FromNumber:       "+15551234567",
				MaxRetries:       3,
				RetryDelay:       30,
				RequestTimeout:   60,
			},
			expectError:     false,
			expectedContent: []string{"business inquiry", "biz-001"},
		},
		{
			name: "process content publication SMS successfully",
			request: &SMSNotificationRequest{
				SubscriberID:  "sub-002",
				EventType:     "event-registration",
				Priority:      "medium",
				Recipients:    []string{"3125559876", "4155558765"},
				EventData:     map[string]interface{}{"entity_type": "research", "entity_id": "research-001"},
				Schedule:      "immediate",
				CreatedAt:     time.Now(),
				CorrelationID: "corr-002",
			},
			mockConfig: &AzureSMSConfig{
				FromNumber:     "+15551234567",
				MaxRetries:     3,
				RetryDelay:     30,
			},
			expectError:     false,
			expectedContent: []string{"research content", "published"},
		},
		{
			name: "process urgent system error SMS",
			request: &SMSNotificationRequest{
				SubscriberID:  "sub-003",
				EventType:     "system-error",
				Priority:      "urgent",
				Recipients:    []string{"5551234567"},
				EventData:     map[string]interface{}{"error_type": "database_failure"},
				Schedule:      "immediate",
				CreatedAt:     time.Now(),
				CorrelationID: "corr-003",
			},
			mockConfig: &AzureSMSConfig{
				FromNumber:     "+15551234567",
				MaxRetries:     3,
			},
			expectError:     false,
			expectedContent: []string{"URGENT", "System error", "database_failure"},
		},
		{
			name: "process invalid request should fail",
			request: &SMSNotificationRequest{
				SubscriberID:  "",
				EventType:     "inquiry-media",
				Priority:      "low",
				Recipients:    []string{},
				EventData:     nil,
				CorrelationID: "corr-004",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			// Validate request
			if !tt.expectError {
				assert.True(t, tt.request.IsValid(), "Request should be valid for success test cases")
				
				// Validate content generation
				content := GenerateSMSContent(tt.request.EventType, tt.request.EventData)
				for _, expectedText := range tt.expectedContent {
					assert.Contains(t, strings.ToLower(content), strings.ToLower(expectedText))
				}
				
				// Validate phone number formatting
				for _, phone := range tt.request.Recipients {
					assert.True(t, IsValidUSPhoneNumber(phone), "Phone number should be valid: %s", phone)
					formatted := FormatPhoneNumberE164(phone)
					assert.True(t, strings.HasPrefix(formatted, "+1"), "Phone should be formatted to E164: %s", formatted)
				}
			} else {
				assert.False(t, tt.request.IsValid(), "Request should be invalid for error test cases")
			}

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestSMSService_HandleDeliveryFailure(t *testing.T) {
	tests := []struct {
		name            string
		message         *SMSMessage
		errorMsg        string
		attemptCount    int
		maxRetries      int
		expectRetry     bool
		expectDeadLetter bool
	}{
		{
			name: "retry failed delivery within retry limit",
			message: &SMSMessage{
				MessageID:     "sms-001",
				SubscriberID:  "sub-001",
				Recipients:    []string{"+12125551234"},
				Content:       "Test SMS message",
				EventType:     "inquiry-business",
				Priority:      "high",
				CorrelationID: "corr-001",
			},
			errorMsg:        "Temporary delivery failure",
			attemptCount:    1,
			maxRetries:      3,
			expectRetry:     true,
			expectDeadLetter: false,
		},
		{
			name: "send to dead letter queue after max retries",
			message: &SMSMessage{
				MessageID:     "sms-002",
				SubscriberID:  "sub-002",
				Recipients:    []string{"+13125559876"},
				Content:       "Test SMS message",
				EventType:     "event-registration",
				Priority:      "medium",
				CorrelationID: "corr-002",
			},
			errorMsg:        "Persistent delivery failure",
			attemptCount:    3,
			maxRetries:      3,
			expectRetry:     false,
			expectDeadLetter: true,
		},
		{
			name: "send to dead letter queue for opted out recipient",
			message: &SMSMessage{
				MessageID:     "sms-003",
				SubscriberID:  "sub-003",
				Recipients:    []string{"+14155558765"},
				Content:       "Test SMS message",
				EventType:     "system-error",
				Priority:      "urgent",
				CorrelationID: "corr-003",
			},
			errorMsg:        "Recipient opted out",
			attemptCount:    1,
			maxRetries:      3,
			expectRetry:     false,
			expectDeadLetter: true,
		},
		{
			name: "send to dead letter queue for invalid phone number",
			message: &SMSMessage{
				MessageID:     "sms-004",
				SubscriberID:  "sub-004",
				Recipients:    []string{"+1invalid"},
				Content:       "Test SMS message",
				EventType:     "inquiry-media",
				Priority:      "low",
				CorrelationID: "corr-004",
			},
			errorMsg:        "Invalid phone number",
			attemptCount:    1,
			maxRetries:      3,
			expectRetry:     false,
			expectDeadLetter: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			// Validate failure handling logic
			shouldRetry := tt.attemptCount < tt.maxRetries && 
				!isPermanentSMSFailure(tt.errorMsg)
			shouldDeadLetter := tt.attemptCount >= tt.maxRetries || 
				isPermanentSMSFailure(tt.errorMsg)

			// Assert
			assert.Equal(t, tt.expectRetry, shouldRetry)
			assert.Equal(t, tt.expectDeadLetter, shouldDeadLetter)
			assert.True(t, tt.message.IsValid())

			_ = ctx // Use context to avoid linting issues
		})
	}
}

// Helper functions for testing

func isPermanentSMSFailure(errorMsg string) bool {
	permanentErrors := []string{
		"Invalid phone number",
		"Recipient opted out",
		"Phone number blocked",
		"Carrier blocked",
	}
	
	for _, permError := range permanentErrors {
		if strings.Contains(errorMsg, permError) {
			return true
		}
	}
	return false
}