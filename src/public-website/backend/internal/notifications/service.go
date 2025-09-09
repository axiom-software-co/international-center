package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/google/uuid"
)

// NotificationRouterService orchestrates notification routing from domain events
type NotificationRouterService struct {
	subscriberRepo  SubscriberRepository
	messageQueue    MessageQueueClient
	emailPublisher  EmailNotificationPublisher
	smsPublisher    SMSNotificationPublisher
	slackPublisher  SlackNotificationPublisher
	logger          *slog.Logger
	config          *NotificationConfig
}

// NewNotificationRouterService creates a new notification router service
func NewNotificationRouterService(
	subscriberRepo SubscriberRepository,
	messageQueue MessageQueueClient,
	emailPublisher EmailNotificationPublisher,
	smsPublisher SMSNotificationPublisher,
	slackPublisher SlackNotificationPublisher,
	logger *slog.Logger,
	config *NotificationConfig,
) *NotificationRouterService {
	return &NotificationRouterService{
		subscriberRepo:  subscriberRepo,
		messageQueue:    messageQueue,
		emailPublisher:  emailPublisher,
		smsPublisher:    smsPublisher,
		slackPublisher:  slackPublisher,
		logger:          logger,
		config:          config,
	}
}

// Start initializes the notification router and begins processing events
func (n *NotificationRouterService) Start(ctx context.Context) error {
	n.logger.Info("Starting notification router service",
		"service", "notification-router",
		"version", n.config.Version)

	// Validate configuration
	if err := n.validateConfiguration(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Subscribe to domain events
	if err := n.subscribeToEvents(ctx); err != nil {
		return fmt.Errorf("failed to subscribe to events: %w", err)
	}

	n.logger.Info("Notification router service started successfully")
	return nil
}

// Stop gracefully shuts down the notification router service
func (n *NotificationRouterService) Stop(ctx context.Context) error {
	n.logger.Info("Stopping notification router service")
	
	if err := n.messageQueue.Close(ctx); err != nil {
		n.logger.Error("Error closing message queue connection", "error", err)
		return fmt.Errorf("failed to close message queue: %w", err)
	}
	
	n.logger.Info("Notification router service stopped successfully")
	return nil
}

// ProcessDomainEvent processes a single domain event and routes notifications
func (n *NotificationRouterService) ProcessDomainEvent(ctx context.Context, event *DomainEvent) error {
	correlationID := extractCorrelationID(ctx, event)
	logger := n.logger.With(
		"correlation_id", correlationID,
		"event_type", event.EventType,
		"topic", event.Topic,
	)

	logger.Debug("Processing domain event for notifications")

	// Classify the domain event
	eventType := ClassifyDomainEvent(event.Topic, event.EventData)
	priority := DeterminePriority(event.EventType, event.EventData)

	logger = logger.With(
		"classified_event_type", eventType,
		"priority", priority,
	)

	// Get matching subscribers from database
	subscribers, err := n.getMatchingSubscribers(ctx, eventType, priority)
	if err != nil {
		logger.Error("Failed to get matching subscribers", "error", err)
		return fmt.Errorf("failed to get subscribers: %w", err)
	}

	if len(subscribers) == 0 {
		logger.Debug("No matching subscribers found for event")
		return nil
	}

	logger.Info("Found matching subscribers for event", "subscriber_count", len(subscribers))

	// Route notifications to handlers based on subscriber preferences
	return n.routeNotifications(ctx, eventType, priority, event.EventData, subscribers, correlationID)
}

// GetSubscribersForEvent retrieves subscribers for a specific event type and priority
func (n *NotificationRouterService) GetSubscribersForEvent(ctx context.Context, eventType EventType, priority PriorityThreshold) ([]*NotificationSubscriber, error) {
	return n.getMatchingSubscribers(ctx, eventType, priority)
}

// ValidateConfiguration validates the service configuration
func (n *NotificationRouterService) ValidateConfiguration() error {
	return n.validateConfiguration()
}

// GetHealthStatus returns the health status of the notification router
func (n *NotificationRouterService) GetHealthStatus(ctx context.Context) (*HealthStatus, error) {
	status := &HealthStatus{
		ServiceName: "notification-router",
		Status:      "healthy",
		Timestamp:   time.Now().UTC(),
		Checks:      make(map[string]CheckResult),
	}

	// Check subscriber repository
	if err := n.subscriberRepo.HealthCheck(ctx); err != nil {
		status.Checks["subscriber_repository"] = CheckResult{
			Status: "unhealthy",
			Error:  err.Error(),
		}
		status.Status = "unhealthy"
	} else {
		status.Checks["subscriber_repository"] = CheckResult{
			Status: "healthy",
		}
	}

	// Check message queue
	if err := n.messageQueue.HealthCheck(ctx); err != nil {
		status.Checks["message_queue"] = CheckResult{
			Status: "unhealthy",
			Error:  err.Error(),
		}
		status.Status = "unhealthy"
	} else {
		status.Checks["message_queue"] = CheckResult{
			Status: "healthy",
		}
	}

	return status, nil
}

// Private helper methods

// validateConfiguration validates the service configuration
func (n *NotificationRouterService) validateConfiguration() error {
	if n.config == nil {
		return domain.NewValidationError("configuration cannot be nil")
	}

	if n.config.MaxRetries < 0 {
		return domain.NewValidationError("max retries cannot be negative")
	}

	if n.config.RetryDelay < time.Second {
		return domain.NewValidationError("retry delay must be at least 1 second")
	}

	if n.config.BatchSize <= 0 {
		return domain.NewValidationError("batch size must be positive")
	}

	if n.config.ProcessingTimeout < 5*time.Second {
		return domain.NewValidationError("processing timeout must be at least 5 seconds")
	}

	return nil
}

// subscribeToEvents subscribes to domain events from all relevant topics
func (n *NotificationRouterService) subscribeToEvents(ctx context.Context) error {
	topics := []string{
		"business-inquiry-events",
		"media-inquiry-events", 
		"donation-inquiry-events",
		"volunteer-inquiry-events",
		"content-publication-events",
		"system-alert-events",
		"capacity-alert-events",
		"admin-action-events",
		"compliance-alert-events",
	}

	for _, topic := range topics {
		if err := n.messageQueue.Subscribe(ctx, topic, n.handleDomainEvent); err != nil {
			return fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
		}
		n.logger.Info("Subscribed to domain event topic", "topic", topic)
	}

	return nil
}

// handleDomainEvent handles incoming domain events from message queue
func (n *NotificationRouterService) handleDomainEvent(ctx context.Context, message *Message) error {
	correlationID := message.CorrelationID
	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	logger := n.logger.With("correlation_id", correlationID, "topic", message.Topic)
	ctx = domain.WithCorrelationID(ctx, correlationID)

	// Parse domain event
	var event DomainEvent
	if err := json.Unmarshal(message.Data, &event); err != nil {
		logger.Error("Failed to parse domain event", "error", err)
		return fmt.Errorf("failed to parse domain event: %w", err)
	}

	// Add message metadata to event
	event.Topic = message.Topic
	event.CorrelationID = correlationID

	// Process the event with timeout
	processCtx, cancel := context.WithTimeout(ctx, n.config.ProcessingTimeout)
	defer cancel()

	return n.ProcessDomainEvent(processCtx, &event)
}

// getMatchingSubscribers retrieves subscribers that match the event criteria
func (n *NotificationRouterService) getMatchingSubscribers(ctx context.Context, eventType EventType, priority PriorityThreshold) ([]*NotificationSubscriber, error) {
	// Get subscribers that have this event type in their preferences
	allSubscribers, err := n.subscriberRepo.GetSubscribersByEventType(ctx, eventType)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscribers by event type: %w", err)
	}

	// Filter by priority threshold and active status
	var matchingSubscribers []*NotificationSubscriber
	for _, subscriber := range allSubscribers {
		if subscriber.Status != SubscriberStatusActive {
			continue
		}

		if !n.meetsPriorityThreshold(priority, subscriber.PriorityThreshold) {
			continue
		}

		matchingSubscribers = append(matchingSubscribers, subscriber)
	}

	return matchingSubscribers, nil
}

// meetsPriorityThreshold checks if event priority meets subscriber threshold
func (n *NotificationRouterService) meetsPriorityThreshold(eventPriority PriorityThreshold, subscriberThreshold PriorityThreshold) bool {
	priorityLevels := map[PriorityThreshold]int{
		PriorityLow:    1,
		PriorityMedium: 2,
		PriorityHigh:   3,
		PriorityUrgent: 4,
	}

	eventLevel, exists := priorityLevels[eventPriority]
	if !exists {
		return false
	}

	thresholdLevel, exists := priorityLevels[subscriberThreshold]
	if !exists {
		return false
	}

	return eventLevel >= thresholdLevel
}

// routeNotifications routes notifications to appropriate handlers
func (n *NotificationRouterService) routeNotifications(
	ctx context.Context,
	eventType EventType,
	priority PriorityThreshold,
	eventData map[string]interface{},
	subscribers []*NotificationSubscriber,
	correlationID string,
) error {
	logger := n.logger.With("correlation_id", correlationID, "event_type", eventType)

	var publishErrors []error

	// Group subscribers by notification method preferences
	emailSubscribers := make([]*NotificationSubscriber, 0)
	smsSubscribers := make([]*NotificationSubscriber, 0)
	slackSubscribers := make([]*NotificationSubscriber, 0)

	for _, subscriber := range subscribers {
		for _, method := range subscriber.NotificationMethods {
			switch method {
			case NotificationMethodEmail:
				emailSubscribers = append(emailSubscribers, subscriber)
			case NotificationMethodSMS:
				smsSubscribers = append(smsSubscribers, subscriber)
			case NotificationMethodBoth:
				emailSubscribers = append(emailSubscribers, subscriber)
				smsSubscribers = append(smsSubscribers, subscriber)
			}
		}
	}

	// Always send to Slack for admin monitoring (all events)
	slackSubscribers = subscribers

	// Publish email notifications
	if len(emailSubscribers) > 0 {
		if err := n.publishEmailNotifications(ctx, eventType, priority, eventData, emailSubscribers, correlationID); err != nil {
			logger.Error("Failed to publish email notifications", "error", err)
			publishErrors = append(publishErrors, fmt.Errorf("email notifications: %w", err))
		}
	}

	// Publish SMS notifications
	if len(smsSubscribers) > 0 {
		if err := n.publishSMSNotifications(ctx, eventType, priority, eventData, smsSubscribers, correlationID); err != nil {
			logger.Error("Failed to publish SMS notifications", "error", err)
			publishErrors = append(publishErrors, fmt.Errorf("SMS notifications: %w", err))
		}
	}

	// Publish Slack notifications
	if len(slackSubscribers) > 0 {
		if err := n.publishSlackNotifications(ctx, eventType, priority, eventData, slackSubscribers, correlationID); err != nil {
			logger.Error("Failed to publish Slack notifications", "error", err)
			publishErrors = append(publishErrors, fmt.Errorf("Slack notifications: %w", err))
		}
	}

	// Return combined errors if any
	if len(publishErrors) > 0 {
		return domain.NewDependencyError("failed to publish some notifications", publishErrors[0])
	}

	logger.Info("Successfully routed notifications",
		"email_count", len(emailSubscribers),
		"sms_count", len(smsSubscribers),
		"slack_count", len(slackSubscribers))

	return nil
}

// publishEmailNotifications publishes email notifications to the email handler queue
func (n *NotificationRouterService) publishEmailNotifications(
	ctx context.Context,
	eventType EventType,
	priority PriorityThreshold,
	eventData map[string]interface{},
	subscribers []*NotificationSubscriber,
	correlationID string,
) error {
	// Group subscribers by schedule preference for batching
	immediateSubscribers := make([]*NotificationSubscriber, 0)
	scheduledSubscribers := make(map[NotificationSchedule][]*NotificationSubscriber)

	for _, subscriber := range subscribers {
		if subscriber.NotificationSchedule == ScheduleImmediate {
			immediateSubscribers = append(immediateSubscribers, subscriber)
		} else {
			if scheduledSubscribers[subscriber.NotificationSchedule] == nil {
				scheduledSubscribers[subscriber.NotificationSchedule] = make([]*NotificationSubscriber, 0)
			}
			scheduledSubscribers[subscriber.NotificationSchedule] = append(
				scheduledSubscribers[subscriber.NotificationSchedule], subscriber)
		}
	}

	// Send immediate notifications
	if len(immediateSubscribers) > 0 {
		recipients := make([]string, len(immediateSubscribers))
		for i, subscriber := range immediateSubscribers {
			recipients[i] = subscriber.Email
		}

		emailRequest := &EmailNotificationRequest{
			SubscriberID:  "router-batch",
			EventType:     string(eventType),
			Priority:      string(priority),
			Recipients:    recipients,
			EventData:     eventData,
			Schedule:      string(ScheduleImmediate),
			CreatedAt:     time.Now().UTC(),
			CorrelationID: correlationID,
		}

		if err := n.emailPublisher.PublishEmailNotification(ctx, emailRequest); err != nil {
			return fmt.Errorf("failed to publish immediate email notifications: %w", err)
		}
	}

	// Send scheduled notifications (would be handled by scheduler service)
	for schedule, scheduleSubscribers := range scheduledSubscribers {
		recipients := make([]string, len(scheduleSubscribers))
		for i, subscriber := range scheduleSubscribers {
			recipients[i] = subscriber.Email
		}

		emailRequest := &EmailNotificationRequest{
			SubscriberID:  "router-scheduled",
			EventType:     string(eventType),
			Priority:      string(priority),
			Recipients:    recipients,
			EventData:     eventData,
			Schedule:      string(schedule),
			CreatedAt:     time.Now().UTC(),
			CorrelationID: correlationID,
		}

		if err := n.emailPublisher.PublishEmailNotification(ctx, emailRequest); err != nil {
			return fmt.Errorf("failed to publish scheduled email notifications for %s: %w", schedule, err)
		}
	}

	return nil
}

// publishSMSNotifications publishes SMS notifications to the SMS handler queue
func (n *NotificationRouterService) publishSMSNotifications(
	ctx context.Context,
	eventType EventType,
	priority PriorityThreshold,
	eventData map[string]interface{},
	subscribers []*NotificationSubscriber,
	correlationID string,
) error {
	// SMS notifications are typically immediate only
	var recipients []string
	for _, subscriber := range subscribers {
		if subscriber.Phone != nil && *subscriber.Phone != "" {
			recipients = append(recipients, *subscriber.Phone)
		}
	}

	if len(recipients) == 0 {
		return nil // No valid phone numbers
	}

	smsRequest := &SMSNotificationRequest{
		SubscriberID:  "router-sms",
		EventType:     string(eventType),
		Priority:      string(priority),
		Recipients:    recipients,
		EventData:     eventData,
		Schedule:      string(ScheduleImmediate),
		CreatedAt:     time.Now().UTC(),
		CorrelationID: correlationID,
	}

	return n.smsPublisher.PublishSMSNotification(ctx, smsRequest)
}

// publishSlackNotifications publishes Slack notifications to the Slack handler queue
func (n *NotificationRouterService) publishSlackNotifications(
	ctx context.Context,
	eventType EventType,
	priority PriorityThreshold,
	eventData map[string]interface{},
	subscribers []*NotificationSubscriber,
	correlationID string,
) error {
	// Slack notifications go to predefined channels based on event type
	channels := GetChannelsForEventType(string(eventType))

	slackRequest := &SlackNotificationRequest{
		SubscriberID:  "router-slack",
		EventType:     string(eventType),
		Priority:      string(priority),
		Channels:      channels,
		EventData:     eventData,
		Schedule:      string(ScheduleImmediate),
		CreatedAt:     time.Now().UTC(),
		CorrelationID: correlationID,
	}

	return n.slackPublisher.PublishSlackNotification(ctx, slackRequest)
}

// extractCorrelationID extracts or generates correlation ID
func extractCorrelationID(ctx context.Context, event *DomainEvent) string {
	if correlationID := domain.GetCorrelationID(ctx); correlationID != "" {
		return correlationID
	}
	
	if event != nil && event.CorrelationID != "" {
		return event.CorrelationID
	}
	
	return uuid.New().String()
}

// GetChannelsForEventType returns Slack channels for an event type
func GetChannelsForEventType(eventType string) []string {
	eventChannelMap := map[string][]string{
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
	
	if channels, exists := eventChannelMap[eventType]; exists {
		return channels
	}
	return []string{"#general"}
}