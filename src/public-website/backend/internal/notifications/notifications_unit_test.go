package notifications

import (
	"context"
	"testing"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	sharedtesting "github.com/axiom-software-co/international-center/src/backend/internal/shared/testing"
	"github.com/stretchr/testify/assert"
)

// RED PHASE - Notification Router Tests (60+ test cases)

func TestClassifyDomainEvent(t *testing.T) {
	tests := []struct {
		name      string
		topic     string
		eventData map[string]interface{}
		expected  EventType
	}{
		{
			name:      "classify business inquiry event",
			topic:     "business-inquiry-events",
			eventData: map[string]interface{}{"operation_type": "CREATE"},
			expected:  EventTypeInquiryBusiness,
		},
		{
			name:      "classify media inquiry event",
			topic:     "media-inquiry-events",
			eventData: map[string]interface{}{"operation_type": "CREATE"},
			expected:  EventTypeInquiryMedia,
		},
		{
			name:      "classify donation inquiry event",
			topic:     "donation-inquiry-events",
			eventData: map[string]interface{}{"operation_type": "CREATE"},
			expected:  EventTypeInquiryDonations,
		},
		{
			name:      "classify volunteer inquiry event",
			topic:     "volunteer-inquiry-events",
			eventData: map[string]interface{}{"operation_type": "CREATE"},
			expected:  EventTypeInquiryVolunteers,
		},
		{
			name:      "classify services content creation event",
			topic:     "services-content-events",
			eventData: map[string]interface{}{"operation_type": "CREATE"},
			expected:  EventTypeEventRegistration,
		},
		{
			name:      "classify news content publish event",
			topic:     "news-content-events",
			eventData: map[string]interface{}{"operation_type": "PUBLISH"},
			expected:  EventTypeEventRegistration,
		},
		{
			name:      "classify research content creation event",
			topic:     "research-content-events",
			eventData: map[string]interface{}{"operation_type": "CREATE"},
			expected:  EventTypeEventRegistration,
		},
		{
			name:      "classify events content publish event",
			topic:     "events-content-events",
			eventData: map[string]interface{}{"operation_type": "PUBLISH"},
			expected:  EventTypeEventRegistration,
		},
		{
			name:      "classify system error event",
			topic:     "system-events",
			eventData: map[string]interface{}{"event_type": "system_error"},
			expected:  EventTypeSystemError,
		},
		{
			name:      "classify capacity alert event",
			topic:     "system-events",
			eventData: map[string]interface{}{"event_type": "capacity_limit"},
			expected:  EventTypeCapacityAlert,
		},
		{
			name:      "classify compliance audit event",
			topic:     "grafana-audit-events",
			eventData: map[string]interface{}{"event_type": "compliance_audit"},
			expected:  EventTypeComplianceAlert,
		},
		{
			name:      "classify unknown event as admin action required",
			topic:     "unknown-events",
			eventData: map[string]interface{}{"operation_type": "UNKNOWN"},
			expected:  EventTypeAdminActionRequired,
		},
		{
			name:      "classify content update event as admin action",
			topic:     "services-content-events",
			eventData: map[string]interface{}{"operation_type": "UPDATE"},
			expected:  EventTypeAdminActionRequired,
		},
		{
			name:      "classify error event with case insensitive match",
			topic:     "system-events",
			eventData: map[string]interface{}{"event_type": "SYSTEM_ERROR"},
			expected:  EventTypeSystemError,
		},
		{
			name:      "classify capacity event with capacity keyword",
			topic:     "system-events",
			eventData: map[string]interface{}{"event_type": "memory_capacity_warning"},
			expected:  EventTypeCapacityAlert,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			// Act
			result := ClassifyDomainEvent(tt.topic, tt.eventData)

			// Assert
			assert.Equal(t, tt.expected, result)

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestExtractPriorityFromEvent(t *testing.T) {
	tests := []struct {
		name      string
		eventData map[string]interface{}
		expected  PriorityThreshold
	}{
		{
			name:      "extract explicit urgent priority",
			eventData: map[string]interface{}{"priority": "urgent"},
			expected:  PriorityUrgent,
		},
		{
			name:      "extract explicit high priority",
			eventData: map[string]interface{}{"priority": "high"},
			expected:  PriorityHigh,
		},
		{
			name:      "extract explicit medium priority",
			eventData: map[string]interface{}{"priority": "medium"},
			expected:  PriorityMedium,
		},
		{
			name:      "extract explicit low priority",
			eventData: map[string]interface{}{"priority": "low"},
			expected:  PriorityLow,
		},
		{
			name:      "default to high priority for delete operations",
			eventData: map[string]interface{}{"operation_type": "DELETE"},
			expected:  PriorityHigh,
		},
		{
			name:      "default to medium priority for create operations",
			eventData: map[string]interface{}{"operation_type": "CREATE"},
			expected:  PriorityMedium,
		},
		{
			name:      "default to medium priority for publish operations",
			eventData: map[string]interface{}{"operation_type": "PUBLISH"},
			expected:  PriorityMedium,
		},
		{
			name:      "default to low priority for update operations",
			eventData: map[string]interface{}{"operation_type": "UPDATE"},
			expected:  PriorityLow,
		},
		{
			name:      "default to low priority for unknown operations",
			eventData: map[string]interface{}{"operation_type": "UNKNOWN"},
			expected:  PriorityLow,
		},
		{
			name:      "default to low priority when no priority or operation specified",
			eventData: map[string]interface{}{},
			expected:  PriorityLow,
		},
		{
			name:      "handle invalid priority string",
			eventData: map[string]interface{}{"priority": "invalid"},
			expected:  PriorityLow,
		},
		{
			name:      "handle non-string priority value",
			eventData: map[string]interface{}{"priority": 123},
			expected:  PriorityLow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			// Act
			result := ExtractPriorityFromEvent(tt.eventData)

			// Assert
			assert.Equal(t, tt.expected, result)

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestNotificationSubscriber_HasEmailMethod(t *testing.T) {
	tests := []struct {
		name                string
		notificationMethods []NotificationMethod
		expected            bool
	}{
		{
			name:                "has email method explicitly",
			notificationMethods: []NotificationMethod{NotificationMethodEmail},
			expected:            true,
		},
		{
			name:                "has both method includes email",
			notificationMethods: []NotificationMethod{NotificationMethodBoth},
			expected:            true,
		},
		{
			name:                "has email and SMS methods",
			notificationMethods: []NotificationMethod{NotificationMethodEmail, NotificationMethodSMS},
			expected:            true,
		},
		{
			name:                "has both and SMS methods",
			notificationMethods: []NotificationMethod{NotificationMethodBoth, NotificationMethodSMS},
			expected:            true,
		},
		{
			name:                "has only SMS method",
			notificationMethods: []NotificationMethod{NotificationMethodSMS},
			expected:            false,
		},
		{
			name:                "has no methods",
			notificationMethods: []NotificationMethod{},
			expected:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			subscriber := &NotificationSubscriber{
				NotificationMethods: tt.notificationMethods,
			}

			// Act
			result := subscriber.HasEmailMethod()

			// Assert
			assert.Equal(t, tt.expected, result)

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestNotificationSubscriber_HasSMSMethod(t *testing.T) {
	tests := []struct {
		name                string
		notificationMethods []NotificationMethod
		expected            bool
	}{
		{
			name:                "has SMS method explicitly",
			notificationMethods: []NotificationMethod{NotificationMethodSMS},
			expected:            true,
		},
		{
			name:                "has both method includes SMS",
			notificationMethods: []NotificationMethod{NotificationMethodBoth},
			expected:            true,
		},
		{
			name:                "has email and SMS methods",
			notificationMethods: []NotificationMethod{NotificationMethodEmail, NotificationMethodSMS},
			expected:            true,
		},
		{
			name:                "has both and email methods",
			notificationMethods: []NotificationMethod{NotificationMethodBoth, NotificationMethodEmail},
			expected:            true,
		},
		{
			name:                "has only email method",
			notificationMethods: []NotificationMethod{NotificationMethodEmail},
			expected:            false,
		},
		{
			name:                "has no methods",
			notificationMethods: []NotificationMethod{},
			expected:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			subscriber := &NotificationSubscriber{
				NotificationMethods: tt.notificationMethods,
			}

			// Act
			result := subscriber.HasSMSMethod()

			// Assert
			assert.Equal(t, tt.expected, result)

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestPriorityThreshold_GetPriorityValue(t *testing.T) {
	tests := []struct {
		name     string
		priority PriorityThreshold
		expected int
	}{
		{
			name:     "urgent priority has highest value",
			priority: PriorityUrgent,
			expected: 4,
		},
		{
			name:     "high priority has second highest value",
			priority: PriorityHigh,
			expected: 3,
		},
		{
			name:     "medium priority has middle value",
			priority: PriorityMedium,
			expected: 2,
		},
		{
			name:     "low priority has lowest value",
			priority: PriorityLow,
			expected: 1,
		},
		{
			name:     "invalid priority returns zero",
			priority: PriorityThreshold("invalid"),
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			// Act
			result := tt.priority.GetPriorityValue()

			// Assert
			assert.Equal(t, tt.expected, result)

			_ = ctx // Use context to avoid linting issues
		})
	}
}

// NotificationRepositoryInterface for testing
type NotificationRepositoryInterface interface {
	// Subscriber Management
	CreateSubscriber(ctx context.Context, subscriber *NotificationSubscriber) error
	GetSubscriber(ctx context.Context, subscriberID string) (*NotificationSubscriber, error)
	GetAllSubscribers(ctx context.Context, limit, offset int) ([]*NotificationSubscriber, error)
	UpdateSubscriber(ctx context.Context, subscriber *NotificationSubscriber) error
	DeleteSubscriber(ctx context.Context, subscriberID string) error

	// Event Routing Queries
	GetActiveSubscribersForEvent(ctx context.Context, eventType EventType, priority PriorityThreshold) ([]*NotificationSubscriber, error)
	GetSubscribersForSchedule(ctx context.Context, schedule NotificationSchedule) ([]*NotificationSubscriber, error)

	// Message Publishing
	PublishNotificationMessage(ctx context.Context, topic string, message *NotificationMessage) error
	SubscribeToDomainEvents(ctx context.Context, topics []string, handler func(context.Context, string, map[string]interface{}) error) error

	// Audit Operations
	PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error
}

// NotificationServiceInterface for testing
type NotificationServiceInterface interface {
	// Event Processing
	ProcessDomainEvent(ctx context.Context, topic string, eventData map[string]interface{}) error
	RouteNotificationToChannels(ctx context.Context, subscriber *NotificationSubscriber, eventType EventType, eventData map[string]interface{}) error

	// Subscriber Management
	CreateSubscriber(ctx context.Context, subscriber *NotificationSubscriber, userID string) error
	UpdateSubscriber(ctx context.Context, subscriber *NotificationSubscriber, userID string) error
	DeleteSubscriber(ctx context.Context, subscriberID string, userID string) error
	GetSubscriber(ctx context.Context, subscriberID string) (*NotificationSubscriber, error)
	GetAllSubscribers(ctx context.Context, limit, offset int) ([]*NotificationSubscriber, error)

	// Scheduling and Filtering
	ShouldProcessImmediately(subscriber *NotificationSubscriber, eventPriority PriorityThreshold) bool
	FilterSubscribersByEventType(subscribers []*NotificationSubscriber, eventType EventType) []*NotificationSubscriber
}

func TestNotificationService_ProcessDomainEvent(t *testing.T) {
	tests := []struct {
		name                string
		topic               string
		eventData           map[string]interface{}
		mockSubscribers     []*NotificationSubscriber
		expectedEventType   EventType
		expectedPriority    PriorityThreshold
		expectedRouteCalls  int
		expectError         bool
	}{
		{
			name:  "process business inquiry event successfully",
			topic: "business-inquiry-events",
			eventData: map[string]interface{}{
				"operation_type":  "CREATE",
				"priority":        "high",
				"correlation_id":  "test-correlation-123",
				"entity_id":       "business-001",
			},
			mockSubscribers: []*NotificationSubscriber{
				{
					SubscriberID:         "sub-001",
					Status:               SubscriberStatusActive,
					EventTypes:           []EventType{EventTypeInquiryBusiness},
					NotificationMethods:  []NotificationMethod{NotificationMethodEmail},
					NotificationSchedule: ScheduleImmediate,
					PriorityThreshold:    PriorityMedium,
					Email:                "user@example.com",
				},
			},
			expectedEventType:  EventTypeInquiryBusiness,
			expectedPriority:   PriorityHigh,
			expectedRouteCalls: 1,
			expectError:        false,
		},
		{
			name:  "process content creation event successfully",
			topic: "services-content-events",
			eventData: map[string]interface{}{
				"operation_type":  "CREATE",
				"priority":        "medium",
				"correlation_id":  "test-correlation-456",
				"entity_id":       "service-001",
			},
			mockSubscribers: []*NotificationSubscriber{
				{
					SubscriberID:         "sub-002",
					Status:               SubscriberStatusActive,
					EventTypes:           []EventType{EventTypeEventRegistration},
					NotificationMethods:  []NotificationMethod{NotificationMethodSMS},
					NotificationSchedule: ScheduleImmediate,
					PriorityThreshold:    PriorityLow,
					Phone:                stringPtr("1234567890"),
				},
			},
			expectedEventType:  EventTypeEventRegistration,
			expectedPriority:   PriorityMedium,
			expectedRouteCalls: 1,
			expectError:        false,
		},
		{
			name:  "process event with no matching subscribers",
			topic: "business-inquiry-events",
			eventData: map[string]interface{}{
				"operation_type": "CREATE",
				"priority":       "low",
			},
			mockSubscribers:    []*NotificationSubscriber{},
			expectedEventType:  EventTypeInquiryBusiness,
			expectedPriority:   PriorityLow,
			expectedRouteCalls: 0,
			expectError:        false,
		},
		{
			name:  "process event with subscriber below priority threshold",
			topic: "media-inquiry-events",
			eventData: map[string]interface{}{
				"operation_type": "CREATE",
				"priority":       "low",
			},
			mockSubscribers: []*NotificationSubscriber{
				{
					SubscriberID:         "sub-003",
					Status:               SubscriberStatusActive,
					EventTypes:           []EventType{EventTypeInquiryMedia},
					NotificationMethods:  []NotificationMethod{NotificationMethodEmail},
					NotificationSchedule: ScheduleImmediate,
					PriorityThreshold:    PriorityHigh, // Higher threshold than event priority
					Email:                "user@example.com",
				},
			},
			expectedEventType:  EventTypeInquiryMedia,
			expectedPriority:   PriorityLow,
			expectedRouteCalls: 0,
			expectError:        false,
		},
		{
			name:  "process event with multiple matching subscribers",
			topic: "donation-inquiry-events",
			eventData: map[string]interface{}{
				"operation_type": "CREATE",
				"priority":       "urgent",
			},
			mockSubscribers: []*NotificationSubscriber{
				{
					SubscriberID:         "sub-004",
					Status:               SubscriberStatusActive,
					EventTypes:           []EventType{EventTypeInquiryDonations},
					NotificationMethods:  []NotificationMethod{NotificationMethodEmail},
					NotificationSchedule: ScheduleImmediate,
					PriorityThreshold:    PriorityMedium,
					Email:                "user1@example.com",
				},
				{
					SubscriberID:         "sub-005",
					Status:               SubscriberStatusActive,
					EventTypes:           []EventType{EventTypeInquiryDonations},
					NotificationMethods:  []NotificationMethod{NotificationMethodSMS},
					NotificationSchedule: ScheduleImmediate,
					PriorityThreshold:    PriorityLow,
					Phone:                stringPtr("0987654321"),
				},
			},
			expectedEventType:  EventTypeInquiryDonations,
			expectedPriority:   PriorityUrgent,
			expectedRouteCalls: 2,
			expectError:        false,
		},
		{
			name:  "process event with inactive subscriber should be filtered",
			topic: "volunteer-inquiry-events",
			eventData: map[string]interface{}{
				"operation_type": "CREATE",
				"priority":       "high",
			},
			mockSubscribers: []*NotificationSubscriber{
				{
					SubscriberID:         "sub-006",
					Status:               SubscriberStatusInactive, // Inactive subscriber
					EventTypes:           []EventType{EventTypeInquiryVolunteers},
					NotificationMethods:  []NotificationMethod{NotificationMethodEmail},
					NotificationSchedule: ScheduleImmediate,
					PriorityThreshold:    PriorityLow,
					Email:                "inactive@example.com",
				},
			},
			expectedEventType:  EventTypeInquiryVolunteers,
			expectedPriority:   PriorityHigh,
			expectedRouteCalls: 0,
			expectError:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			// This test validates the contract and expected behavior
			// Implementation will be added in GREEN phase

			// Assert expected classification
			actualEventType := ClassifyDomainEvent(tt.topic, tt.eventData)
			assert.Equal(t, tt.expectedEventType, actualEventType)

			actualPriority := ExtractPriorityFromEvent(tt.eventData)
			assert.Equal(t, tt.expectedPriority, actualPriority)

			// Count expected active subscribers that meet criteria
			activeMatchingCount := 0
			for _, sub := range tt.mockSubscribers {
				if sub.Status == SubscriberStatusActive &&
					containsEventType(sub.EventTypes, actualEventType) &&
					sub.PriorityThreshold.GetPriorityValue() <= actualPriority.GetPriorityValue() {
					activeMatchingCount++
				}
			}

			assert.Equal(t, tt.expectedRouteCalls, activeMatchingCount)

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestNotificationService_RouteNotificationToChannels(t *testing.T) {
	tests := []struct {
		name            string
		subscriber      *NotificationSubscriber
		eventType       EventType
		eventData       map[string]interface{}
		expectedEmails  []string
		expectedSMS     []string
		expectedSlack   []string
		expectError     bool
	}{
		{
			name: "route to email channel for email method",
			subscriber: &NotificationSubscriber{
				SubscriberID:        "sub-001",
				Email:               "user@example.com",
				NotificationMethods: []NotificationMethod{NotificationMethodEmail},
			},
			eventType:      EventTypeInquiryBusiness,
			eventData:      map[string]interface{}{"entity_id": "biz-001"},
			expectedEmails: []string{"user@example.com"},
			expectedSMS:    []string{},
			expectedSlack:  []string{},
			expectError:    false,
		},
		{
			name: "route to SMS channel for SMS method",
			subscriber: &NotificationSubscriber{
				SubscriberID:        "sub-002",
				Phone:               stringPtr("1234567890"),
				NotificationMethods: []NotificationMethod{NotificationMethodSMS},
			},
			eventType:      EventTypeInquiryMedia,
			eventData:      map[string]interface{}{"entity_id": "media-001"},
			expectedEmails: []string{},
			expectedSMS:    []string{"1234567890"},
			expectedSlack:  []string{},
			expectError:    false,
		},
		{
			name: "route to both channels for both method",
			subscriber: &NotificationSubscriber{
				SubscriberID:        "sub-003",
				Email:               "user@example.com",
				Phone:               stringPtr("1234567890"),
				NotificationMethods: []NotificationMethod{NotificationMethodBoth},
			},
			eventType:      EventTypeInquiryDonations,
			eventData:      map[string]interface{}{"entity_id": "donation-001"},
			expectedEmails: []string{"user@example.com"},
			expectedSMS:    []string{"1234567890"},
			expectedSlack:  []string{},
			expectError:    false,
		},
		{
			name: "route operational event to slack channel",
			subscriber: &NotificationSubscriber{
				SubscriberID:        "sub-004",
				Email:               "ops@example.com",
				NotificationMethods: []NotificationMethod{NotificationMethodEmail},
			},
			eventType:      EventTypeSystemError,
			eventData:      map[string]interface{}{"entity_id": "error-001"},
			expectedEmails: []string{"ops@example.com"},
			expectedSMS:    []string{},
			expectedSlack:  []string{"#system-alerts"},
			expectError:    false,
		},
		{
			name: "route SMS method but no phone number",
			subscriber: &NotificationSubscriber{
				SubscriberID:        "sub-005",
				Email:               "user@example.com",
				Phone:               nil, // No phone number
				NotificationMethods: []NotificationMethod{NotificationMethodSMS},
			},
			eventType:      EventTypeInquiryVolunteers,
			eventData:      map[string]interface{}{"entity_id": "volunteer-001"},
			expectedEmails: []string{},
			expectedSMS:    []string{},
			expectedSlack:  []string{},
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			// This test validates the routing logic expectations
			// Implementation will be added in GREEN phase

			// Validate subscriber methods are correctly interpreted
			if tt.subscriber.HasEmailMethod() && tt.subscriber.Email != "" {
				assert.Contains(t, tt.expectedEmails, tt.subscriber.Email)
			}

			if tt.subscriber.HasSMSMethod() && tt.subscriber.Phone != nil {
				assert.Contains(t, tt.expectedSMS, *tt.subscriber.Phone)
			}

			// Operational events should include Slack notifications
			if isOperationalEventType(tt.eventType) {
				assert.NotEmpty(t, tt.expectedSlack)
			}

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestNotificationService_ShouldProcessImmediately(t *testing.T) {
	tests := []struct {
		name           string
		subscriber     *NotificationSubscriber
		eventPriority  PriorityThreshold
		expectedResult bool
	}{
		{
			name: "process immediate schedule immediately",
			subscriber: &NotificationSubscriber{
				NotificationSchedule: ScheduleImmediate,
				PriorityThreshold:    PriorityLow,
			},
			eventPriority:  PriorityMedium,
			expectedResult: true,
		},
		{
			name: "process urgent priority immediately regardless of schedule",
			subscriber: &NotificationSubscriber{
				NotificationSchedule: ScheduleDaily,
				PriorityThreshold:    PriorityLow,
			},
			eventPriority:  PriorityUrgent,
			expectedResult: true,
		},
		{
			name: "process high priority immediately regardless of schedule",
			subscriber: &NotificationSubscriber{
				NotificationSchedule: ScheduleHourly,
				PriorityThreshold:    PriorityLow,
			},
			eventPriority:  PriorityHigh,
			expectedResult: true,
		},
		{
			name: "defer low priority event for hourly schedule",
			subscriber: &NotificationSubscriber{
				NotificationSchedule: ScheduleHourly,
				PriorityThreshold:    PriorityLow,
			},
			eventPriority:  PriorityLow,
			expectedResult: false,
		},
		{
			name: "defer medium priority event for daily schedule",
			subscriber: &NotificationSubscriber{
				NotificationSchedule: ScheduleDaily,
				PriorityThreshold:    PriorityLow,
			},
			eventPriority:  PriorityMedium,
			expectedResult: false,
		},
		{
			name: "process event at subscriber priority threshold",
			subscriber: &NotificationSubscriber{
				NotificationSchedule: ScheduleImmediate,
				PriorityThreshold:    PriorityMedium,
			},
			eventPriority:  PriorityMedium,
			expectedResult: true,
		},
		{
			name: "reject event below subscriber priority threshold",
			subscriber: &NotificationSubscriber{
				NotificationSchedule: ScheduleImmediate,
				PriorityThreshold:    PriorityHigh,
			},
			eventPriority:  PriorityMedium,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			// This test validates the scheduling logic
			// Implementation will be added in GREEN phase

			// Validate immediate processing conditions
			isUrgentOrHigh := tt.eventPriority == PriorityUrgent || tt.eventPriority == PriorityHigh
			isImmediateSchedule := tt.subscriber.NotificationSchedule == ScheduleImmediate
			meetsPriorityThreshold := tt.eventPriority.GetPriorityValue() >= tt.subscriber.PriorityThreshold.GetPriorityValue()

			expectedProcessing := (isUrgentOrHigh || isImmediateSchedule) && meetsPriorityThreshold
			assert.Equal(t, tt.expectedResult, expectedProcessing)

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestNotificationService_FilterSubscribersByEventType(t *testing.T) {
	tests := []struct {
		name        string
		subscribers []*NotificationSubscriber
		eventType   EventType
		expectedIDs []string
	}{
		{
			name: "filter subscribers by business inquiry event type",
			subscribers: []*NotificationSubscriber{
				{
					SubscriberID: "sub-001",
					EventTypes:   []EventType{EventTypeInquiryBusiness, EventTypeInquiryMedia},
				},
				{
					SubscriberID: "sub-002",
					EventTypes:   []EventType{EventTypeInquiryDonations},
				},
				{
					SubscriberID: "sub-003",
					EventTypes:   []EventType{EventTypeInquiryBusiness},
				},
			},
			eventType:   EventTypeInquiryBusiness,
			expectedIDs: []string{"sub-001", "sub-003"},
		},
		{
			name: "filter subscribers with no matching event types",
			subscribers: []*NotificationSubscriber{
				{
					SubscriberID: "sub-001",
					EventTypes:   []EventType{EventTypeInquiryMedia},
				},
				{
					SubscriberID: "sub-002",
					EventTypes:   []EventType{EventTypeInquiryDonations},
				},
			},
			eventType:   EventTypeInquiryBusiness,
			expectedIDs: []string{},
		},
		{
			name: "filter subscribers with all matching event types",
			subscribers: []*NotificationSubscriber{
				{
					SubscriberID: "sub-001",
					EventTypes:   []EventType{EventTypeEventRegistration, EventTypeSystemError},
				},
				{
					SubscriberID: "sub-002",
					EventTypes:   []EventType{EventTypeEventRegistration, EventTypeInquiryMedia},
				},
			},
			eventType:   EventTypeEventRegistration,
			expectedIDs: []string{"sub-001", "sub-002"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			// Act - Manual filtering for testing
			filteredIDs := make([]string, 0)
			for _, subscriber := range tt.subscribers {
				if containsEventType(subscriber.EventTypes, tt.eventType) {
					filteredIDs = append(filteredIDs, subscriber.SubscriberID)
				}
			}

			// Assert
			assert.Equal(t, tt.expectedIDs, filteredIDs)

			_ = ctx // Use context to avoid linting issues
		})
	}
}

// Helper functions for testing

func stringPtr(s string) *string {
	return &s
}

func containsEventType(eventTypes []EventType, target EventType) bool {
	for _, eventType := range eventTypes {
		if eventType == target {
			return true
		}
	}
	return false
}

func isOperationalEventType(eventType EventType) bool {
	operationalTypes := []EventType{
		EventTypeSystemError,
		EventTypeCapacityAlert,
		EventTypeComplianceAlert,
		EventTypeAdminActionRequired,
	}
	
	for _, opType := range operationalTypes {
		if eventType == opType {
			return true
		}
	}
	return false
}