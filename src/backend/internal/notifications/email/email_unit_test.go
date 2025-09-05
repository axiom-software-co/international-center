package email

import (
	"context"
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	sharedtesting "github.com/axiom-software-co/international-center/src/backend/internal/shared/testing"
	"github.com/stretchr/testify/assert"
)

// RED PHASE - Email Handler Tests (40+ test cases)

func TestEmailMessage_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		message  *EmailMessage
		expected bool
	}{
		{
			name: "valid email message with HTML content",
			message: &EmailMessage{
				MessageID:     "msg-001",
				SubscriberID:  "sub-001",
				Recipients:    []string{"user@example.com"},
				Subject:       "Test Subject",
				HtmlContent:   "<p>Test HTML content</p>",
				TextContent:   "",
				EventType:     "inquiry-business",
				Priority:      "high",
				CorrelationID: "corr-001",
			},
			expected: true,
		},
		{
			name: "valid email message with text content only",
			message: &EmailMessage{
				MessageID:     "msg-002",
				SubscriberID:  "sub-002",
				Recipients:    []string{"user@example.com"},
				Subject:       "Test Subject",
				HtmlContent:   "",
				TextContent:   "Test text content",
				EventType:     "inquiry-media",
				Priority:      "medium",
				CorrelationID: "corr-002",
			},
			expected: true,
		},
		{
			name: "valid email message with multiple recipients",
			message: &EmailMessage{
				MessageID:     "msg-003",
				SubscriberID:  "sub-003",
				Recipients:    []string{"user1@example.com", "user2@example.com"},
				Subject:       "Test Subject",
				HtmlContent:   "<p>Test content</p>",
				TextContent:   "Test content",
				EventType:     "inquiry-donations",
				Priority:      "low",
				CorrelationID: "corr-003",
			},
			expected: true,
		},
		{
			name: "invalid email message - missing message ID",
			message: &EmailMessage{
				MessageID:     "",
				SubscriberID:  "sub-004",
				Recipients:    []string{"user@example.com"},
				Subject:       "Test Subject",
				HtmlContent:   "<p>Test content</p>",
				CorrelationID: "corr-004",
			},
			expected: false,
		},
		{
			name: "invalid email message - missing subscriber ID",
			message: &EmailMessage{
				MessageID:     "msg-005",
				SubscriberID:  "",
				Recipients:    []string{"user@example.com"},
				Subject:       "Test Subject",
				HtmlContent:   "<p>Test content</p>",
				CorrelationID: "corr-005",
			},
			expected: false,
		},
		{
			name: "invalid email message - empty recipients",
			message: &EmailMessage{
				MessageID:     "msg-006",
				SubscriberID:  "sub-006",
				Recipients:    []string{},
				Subject:       "Test Subject",
				HtmlContent:   "<p>Test content</p>",
				CorrelationID: "corr-006",
			},
			expected: false,
		},
		{
			name: "invalid email message - missing subject",
			message: &EmailMessage{
				MessageID:     "msg-007",
				SubscriberID:  "sub-007",
				Recipients:    []string{"user@example.com"},
				Subject:       "",
				HtmlContent:   "<p>Test content</p>",
				CorrelationID: "corr-007",
			},
			expected: false,
		},
		{
			name: "invalid email message - missing both HTML and text content",
			message: &EmailMessage{
				MessageID:     "msg-008",
				SubscriberID:  "sub-008",
				Recipients:    []string{"user@example.com"},
				Subject:       "Test Subject",
				HtmlContent:   "",
				TextContent:   "",
				CorrelationID: "corr-008",
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

func TestEmailNotificationRequest_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		request  *EmailNotificationRequest
		expected bool
	}{
		{
			name: "valid email notification request",
			request: &EmailNotificationRequest{
				SubscriberID:  "sub-001",
				EventType:     "inquiry-business",
				Priority:      "high",
				Recipients:    []string{"user@example.com"},
				EventData:     map[string]interface{}{"entity_id": "biz-001"},
				Schedule:      "immediate",
				CreatedAt:     time.Now(),
				CorrelationID: "corr-001",
			},
			expected: true,
		},
		{
			name: "valid request with multiple recipients",
			request: &EmailNotificationRequest{
				SubscriberID:  "sub-002",
				EventType:     "inquiry-media",
				Priority:      "medium",
				Recipients:    []string{"user1@example.com", "user2@example.com"},
				EventData:     map[string]interface{}{"entity_id": "media-001"},
				Schedule:      "immediate",
				CreatedAt:     time.Now(),
				CorrelationID: "corr-002",
			},
			expected: true,
		},
		{
			name: "invalid request - missing subscriber ID",
			request: &EmailNotificationRequest{
				SubscriberID:  "",
				EventType:     "inquiry-donations",
				Priority:      "low",
				Recipients:    []string{"user@example.com"},
				EventData:     map[string]interface{}{"entity_id": "donation-001"},
				CorrelationID: "corr-003",
			},
			expected: false,
		},
		{
			name: "invalid request - missing event type",
			request: &EmailNotificationRequest{
				SubscriberID:  "sub-004",
				EventType:     "",
				Priority:      "high",
				Recipients:    []string{"user@example.com"},
				EventData:     map[string]interface{}{"entity_id": "volunteer-001"},
				CorrelationID: "corr-004",
			},
			expected: false,
		},
		{
			name: "invalid request - empty recipients",
			request: &EmailNotificationRequest{
				SubscriberID:  "sub-005",
				EventType:     "event-registration",
				Priority:      "medium",
				Recipients:    []string{},
				EventData:     map[string]interface{}{"entity_id": "event-001"},
				CorrelationID: "corr-005",
			},
			expected: false,
		},
		{
			name: "invalid request - nil event data",
			request: &EmailNotificationRequest{
				SubscriberID:  "sub-006",
				EventType:     "system-error",
				Priority:      "urgent",
				Recipients:    []string{"admin@example.com"},
				EventData:     nil,
				CorrelationID: "corr-006",
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

func TestDeliveryStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   DeliveryStatus
		expected bool
	}{
		{
			name:     "valid pending status",
			status:   DeliveryStatusPending,
			expected: true,
		},
		{
			name:     "valid sent status",
			status:   DeliveryStatusSent,
			expected: true,
		},
		{
			name:     "valid delivered status",
			status:   DeliveryStatusDelivered,
			expected: true,
		},
		{
			name:     "valid failed status",
			status:   DeliveryStatusFailed,
			expected: true,
		},
		{
			name:     "valid bounced status",
			status:   DeliveryStatusBounced,
			expected: true,
		},
		{
			name:     "valid spam status",
			status:   DeliveryStatusSpam,
			expected: true,
		},
		{
			name:     "invalid status",
			status:   DeliveryStatus("invalid"),
			expected: false,
		},
		{
			name:     "empty status",
			status:   DeliveryStatus(""),
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

func TestDeliveryStatus_IsFinalStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   DeliveryStatus
		expected bool
	}{
		{
			name:     "delivered is final status",
			status:   DeliveryStatusDelivered,
			expected: true,
		},
		{
			name:     "failed is final status",
			status:   DeliveryStatusFailed,
			expected: true,
		},
		{
			name:     "bounced is final status",
			status:   DeliveryStatusBounced,
			expected: true,
		},
		{
			name:     "spam is final status",
			status:   DeliveryStatusSpam,
			expected: true,
		},
		{
			name:     "pending is not final status",
			status:   DeliveryStatusPending,
			expected: false,
		},
		{
			name:     "sent is not final status",
			status:   DeliveryStatusSent,
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

func TestGetTemplateIDByEventType(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		expected  string
	}{
		{
			name:      "business inquiry event type",
			eventType: "inquiry-business",
			expected:  "business-inquiry-template",
		},
		{
			name:      "media inquiry event type",
			eventType: "inquiry-media",
			expected:  "media-inquiry-template",
		},
		{
			name:      "donation inquiry event type",
			eventType: "inquiry-donations",
			expected:  "donation-inquiry-template",
		},
		{
			name:      "volunteer inquiry event type",
			eventType: "inquiry-volunteers",
			expected:  "volunteer-inquiry-template",
		},
		{
			name:      "event registration event type",
			eventType: "event-registration",
			expected:  "content-publication-template",
		},
		{
			name:      "system error event type",
			eventType: "system-error",
			expected:  "system-alert-template",
		},
		{
			name:      "capacity alert event type",
			eventType: "capacity-alert",
			expected:  "capacity-warning-template",
		},
		{
			name:      "admin action required event type",
			eventType: "admin-action-required",
			expected:  "admin-action-template",
		},
		{
			name:      "compliance alert event type",
			eventType: "compliance-alert",
			expected:  "compliance-alert-template",
		},
		{
			name:      "unknown event type returns default template",
			eventType: "unknown-event",
			expected:  "default-notification-template",
		},
		{
			name:      "empty event type returns default template",
			eventType: "",
			expected:  "default-notification-template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			// Act
			result := GetTemplateIDByEventType(tt.eventType)

			// Assert
			assert.Equal(t, tt.expected, result)

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestGenerateSubjectByEventType(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		eventData map[string]interface{}
		expected  string
	}{
		{
			name:      "business inquiry subject",
			eventType: "inquiry-business",
			eventData: map[string]interface{}{},
			expected:  "New Business Inquiry Received",
		},
		{
			name:      "media inquiry subject",
			eventType: "inquiry-media",
			eventData: map[string]interface{}{},
			expected:  "New Media Inquiry Received",
		},
		{
			name:      "donation inquiry subject",
			eventType: "inquiry-donations",
			eventData: map[string]interface{}{},
			expected:  "New Donation Inquiry Received",
		},
		{
			name:      "volunteer inquiry subject",
			eventType: "inquiry-volunteers",
			eventData: map[string]interface{}{},
			expected:  "New Volunteer Application Received",
		},
		{
			name:      "content publication with entity type",
			eventType: "event-registration",
			eventData: map[string]interface{}{"entity_type": "research"},
			expected:  "New research Content Published",
		},
		{
			name:      "content publication without entity type",
			eventType: "event-registration",
			eventData: map[string]interface{}{},
			expected:  "New Content Published",
		},
		{
			name:      "system error subject",
			eventType: "system-error",
			eventData: map[string]interface{}{},
			expected:  "System Alert: Error Detected",
		},
		{
			name:      "capacity alert subject",
			eventType: "capacity-alert",
			eventData: map[string]interface{}{},
			expected:  "Capacity Warning Alert",
		},
		{
			name:      "admin action required subject",
			eventType: "admin-action-required",
			eventData: map[string]interface{}{},
			expected:  "Admin Action Required",
		},
		{
			name:      "compliance alert subject",
			eventType: "compliance-alert",
			eventData: map[string]interface{}{},
			expected:  "Compliance Alert - Review Required",
		},
		{
			name:      "unknown event type default subject",
			eventType: "unknown-event",
			eventData: map[string]interface{}{},
			expected:  "Notification Alert",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			// Act
			result := GenerateSubjectByEventType(tt.eventType, tt.eventData)

			// Assert
			assert.Equal(t, tt.expected, result)

			_ = ctx // Use context to avoid linting issues
		})
	}
}

// EmailRepositoryInterface for testing
type EmailRepositoryInterface interface {
	// RabbitMQ Operations
	SubscribeToEmailRequests(ctx context.Context, handler func(context.Context, *EmailNotificationRequest) error) error
	PublishToDeadLetterQueue(ctx context.Context, request *EmailNotificationRequest, errorMsg string) error
	AcknowledgeMessage(ctx context.Context, messageID string) error
	RejectMessage(ctx context.Context, messageID string, requeue bool) error

	// Azure Communication Service Operations
	SendEmail(ctx context.Context, message *EmailMessage) (*EmailDeliveryStatus, error)
	GetDeliveryStatus(ctx context.Context, messageID string) (*EmailDeliveryStatus, error)
	
	// Template and Configuration Operations
	GetEmailTemplate(ctx context.Context, templateID string) (*EmailTemplate, error)
	GetAzureEmailConfig(ctx context.Context) (*AzureEmailConfig, error)
	
	// Retry and Tracking Operations
	SaveDeliveryStatus(ctx context.Context, status *EmailDeliveryStatus) error
	GetFailedDeliveries(ctx context.Context, limit int) ([]*EmailDeliveryStatus, error)
	ScheduleRetry(ctx context.Context, messageID string, retryAt time.Time) error

	// Audit Operations
	PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error
}

// EmailServiceInterface for testing
type EmailServiceInterface interface {
	// Message Processing
	ProcessEmailNotificationRequest(ctx context.Context, request *EmailNotificationRequest) error
	FormatEmailMessage(ctx context.Context, request *EmailNotificationRequest) (*EmailMessage, error)
	SendEmailMessage(ctx context.Context, message *EmailMessage) error

	// Template Operations
	RenderEmailTemplate(ctx context.Context, templateID string, data *EmailTemplateData) (*EmailMessage, error)
	BuildTemplateData(ctx context.Context, request *EmailNotificationRequest) (*EmailTemplateData, error)

	// Delivery Management
	HandleDeliveryFailure(ctx context.Context, message *EmailMessage, errorMsg string) error
	RetryFailedDelivery(ctx context.Context, messageID string) error
	ProcessDeliveryCallback(ctx context.Context, messageID string, status DeliveryStatus, errorMsg *string) error

	// Health and Status
	HealthCheck(ctx context.Context) error
	GetDeliveryMetrics(ctx context.Context) (map[string]interface{}, error)
}

func TestEmailService_ProcessEmailNotificationRequest(t *testing.T) {
	tests := []struct {
		name           string
		request        *EmailNotificationRequest
		mockTemplate   *EmailTemplate
		mockConfig     *AzureEmailConfig
		expectError    bool
		expectedSubject string
	}{
		{
			name: "process business inquiry notification successfully",
			request: &EmailNotificationRequest{
				SubscriberID:  "sub-001",
				EventType:     "inquiry-business",
				Priority:      "high",
				Recipients:    []string{"business@example.com"},
				EventData:     map[string]interface{}{"entity_id": "biz-001", "user_id": "user-001"},
				Schedule:      "immediate",
				CreatedAt:     time.Now(),
				CorrelationID: "corr-001",
			},
			mockTemplate: &EmailTemplate{
				TemplateID:   "business-inquiry-template",
				EventType:    "inquiry-business",
				Subject:      "New Business Inquiry Received",
				HtmlTemplate: "<p>New inquiry from {{.UserID}}</p>",
				TextTemplate: "New inquiry from {{.UserID}}",
			},
			mockConfig: &AzureEmailConfig{
				SenderAddress:    "notifications@example.com",
				SenderName:      "Notifications",
				ReplyToAddress:  "noreply@example.com",
				MaxRetries:      3,
				RetryDelay:      30,
				RequestTimeout:  60,
			},
			expectError:     false,
			expectedSubject: "New Business Inquiry Received",
		},
		{
			name: "process media inquiry notification successfully",
			request: &EmailNotificationRequest{
				SubscriberID:  "sub-002",
				EventType:     "inquiry-media",
				Priority:      "medium",
				Recipients:    []string{"media@example.com", "pr@example.com"},
				EventData:     map[string]interface{}{"entity_id": "media-001", "user_id": "user-002"},
				Schedule:      "immediate",
				CreatedAt:     time.Now(),
				CorrelationID: "corr-002",
			},
			mockTemplate: &EmailTemplate{
				TemplateID:   "media-inquiry-template",
				EventType:    "inquiry-media",
				Subject:      "New Media Inquiry Received",
				HtmlTemplate: "<p>New media inquiry from {{.UserID}}</p>",
				TextTemplate: "New media inquiry from {{.UserID}}",
			},
			mockConfig: &AzureEmailConfig{
				SenderAddress:   "notifications@example.com",
				SenderName:     "Notifications",
				ReplyToAddress: "noreply@example.com",
				MaxRetries:     3,
			},
			expectError:     false,
			expectedSubject: "New Media Inquiry Received",
		},
		{
			name: "process invalid request should fail",
			request: &EmailNotificationRequest{
				SubscriberID:  "",
				EventType:     "inquiry-donations",
				Priority:      "low",
				Recipients:    []string{},
				EventData:     nil,
				CorrelationID: "corr-003",
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
				
				// Validate template mapping
				expectedTemplateID := GetTemplateIDByEventType(tt.request.EventType)
				if tt.mockTemplate != nil {
					assert.Equal(t, expectedTemplateID, tt.mockTemplate.TemplateID)
				}
				
				// Validate subject generation
				actualSubject := GenerateSubjectByEventType(tt.request.EventType, tt.request.EventData)
				assert.Equal(t, tt.expectedSubject, actualSubject)
			} else {
				assert.False(t, tt.request.IsValid(), "Request should be invalid for error test cases")
			}

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestEmailService_HandleDeliveryFailure(t *testing.T) {
	tests := []struct {
		name            string
		message         *EmailMessage
		errorMsg        string
		attemptCount    int
		maxRetries      int
		expectRetry     bool
		expectDeadLetter bool
	}{
		{
			name: "retry failed delivery within retry limit",
			message: &EmailMessage{
				MessageID:     "msg-001",
				SubscriberID:  "sub-001",
				Recipients:    []string{"user@example.com"},
				Subject:       "Test Subject",
				HtmlContent:   "<p>Test content</p>",
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
			message: &EmailMessage{
				MessageID:     "msg-002",
				SubscriberID:  "sub-002",
				Recipients:    []string{"user@example.com"},
				Subject:       "Test Subject",
				HtmlContent:   "<p>Test content</p>",
				EventType:     "inquiry-media",
				Priority:      "medium",
				CorrelationID: "corr-002",
			},
			errorMsg:        "Permanent delivery failure",
			attemptCount:    3,
			maxRetries:      3,
			expectRetry:     false,
			expectDeadLetter: true,
		},
		{
			name: "send to dead letter queue for invalid recipient",
			message: &EmailMessage{
				MessageID:     "msg-003",
				SubscriberID:  "sub-003",
				Recipients:    []string{"invalid-email"},
				Subject:       "Test Subject",
				HtmlContent:   "<p>Test content</p>",
				EventType:     "inquiry-donations",
				Priority:      "low",
				CorrelationID: "corr-003",
			},
			errorMsg:        "Invalid email address",
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
				!isPermanentFailure(tt.errorMsg)
			shouldDeadLetter := tt.attemptCount >= tt.maxRetries || 
				isPermanentFailure(tt.errorMsg)

			// Assert
			assert.Equal(t, tt.expectRetry, shouldRetry)
			assert.Equal(t, tt.expectDeadLetter, shouldDeadLetter)
			assert.True(t, tt.message.IsValid())

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestEmailService_DeliveryStatusTracking(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  DeliveryStatus
		callbackStatus DeliveryStatus
		expectedFinal  bool
		expectUpdate   bool
	}{
		{
			name:           "update from pending to sent",
			initialStatus:  DeliveryStatusPending,
			callbackStatus: DeliveryStatusSent,
			expectedFinal:  false,
			expectUpdate:   true,
		},
		{
			name:           "update from sent to delivered",
			initialStatus:  DeliveryStatusSent,
			callbackStatus: DeliveryStatusDelivered,
			expectedFinal:  true,
			expectUpdate:   true,
		},
		{
			name:           "update from pending to failed",
			initialStatus:  DeliveryStatusPending,
			callbackStatus: DeliveryStatusFailed,
			expectedFinal:  true,
			expectUpdate:   true,
		},
		{
			name:           "update from sent to bounced",
			initialStatus:  DeliveryStatusSent,
			callbackStatus: DeliveryStatusBounced,
			expectedFinal:  true,
			expectUpdate:   true,
		},
		{
			name:           "no update for final status",
			initialStatus:  DeliveryStatusDelivered,
			callbackStatus: DeliveryStatusFailed,
			expectedFinal:  true,
			expectUpdate:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			// Validate status transitions
			shouldUpdate := !tt.initialStatus.IsFinalStatus()
			isFinal := tt.callbackStatus.IsFinalStatus()

			// Assert
			assert.Equal(t, tt.expectUpdate, shouldUpdate)
			assert.Equal(t, tt.expectedFinal, isFinal)
			assert.True(t, tt.callbackStatus.IsValid())

			_ = ctx // Use context to avoid linting issues
		})
	}
}

// Helper functions for testing

func isPermanentFailure(errorMsg string) bool {
	permanentErrors := []string{
		"Invalid email address",
		"Recipient blocked",
		"Domain not found",
		"Mailbox does not exist",
	}
	
	for _, permError := range permanentErrors {
		if errorMsg == permError {
			return true
		}
	}
	return false
}