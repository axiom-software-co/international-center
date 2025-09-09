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

// Message represents a message in the message queue system
type Message struct {
	ID            string            `json:"id"`
	Topic         string            `json:"topic"`
	Data          []byte            `json:"data"`
	Headers       map[string]string `json:"headers"`
	CorrelationID string            `json:"correlation_id"`
	Timestamp     time.Time         `json:"timestamp"`
	RetryCount    int               `json:"retry_count"`
	MaxRetries    int               `json:"max_retries"`
}

// MessageHandler represents a function that handles incoming messages
type MessageHandler func(ctx context.Context, message *Message) error

// MessageQueueClient provides message queue operations for RabbitMQ integration
type MessageQueueClient interface {
	// Publish a message to a topic
	Publish(ctx context.Context, topic string, data []byte, headers map[string]string) error
	
	// Subscribe to messages from a topic
	Subscribe(ctx context.Context, topic string, handler MessageHandler) error
	
	// Unsubscribe from a topic
	Unsubscribe(ctx context.Context, topic string) error
	
	// Create a dead letter queue for failed messages
	CreateDeadLetterQueue(ctx context.Context, topic string) error
	
	// Health check
	HealthCheck(ctx context.Context) error
	
	// Close the connection
	Close(ctx context.Context) error
}

// RabbitMQClient implements MessageQueueClient using RabbitMQ
type RabbitMQClient struct {
	connectionString string
	logger          *slog.Logger
	subscriptions   map[string]MessageHandler
	config          *MessageQueueConfig
}

// MessageQueueConfig is defined in config.go

// NewRabbitMQClient creates a new RabbitMQ client
func NewRabbitMQClient(connectionString string, logger *slog.Logger, config *MessageQueueConfig) *RabbitMQClient {
	return &RabbitMQClient{
		connectionString: connectionString,
		logger:          logger,
		subscriptions:   make(map[string]MessageHandler),
		config:          config,
	}
}

// Publish publishes a message to a topic
func (r *RabbitMQClient) Publish(ctx context.Context, topic string, data []byte, headers map[string]string) error {
	message := &Message{
		ID:            uuid.New().String(),
		Topic:         topic,
		Data:          data,
		Headers:       headers,
		CorrelationID: domain.GetCorrelationID(ctx),
		Timestamp:     time.Now().UTC(),
		RetryCount:    0,
		MaxRetries:    r.config.MaxRetries,
	}

	r.logger.Debug("Publishing message to topic",
		"topic", topic,
		"message_id", message.ID,
		"correlation_id", message.CorrelationID)

	// In a real implementation, this would publish to RabbitMQ
	// For now, we simulate the operation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Simulate publishing delay
		time.Sleep(10 * time.Millisecond)
	}

	r.logger.Info("Message published successfully",
		"topic", topic,
		"message_id", message.ID)

	return nil
}

// Subscribe subscribes to messages from a topic
func (r *RabbitMQClient) Subscribe(ctx context.Context, topic string, handler MessageHandler) error {
	r.logger.Info("Subscribing to topic", "topic", topic)

	r.subscriptions[topic] = handler

	// In a real implementation, this would set up RabbitMQ consumer
	// For now, we simulate the subscription
	go r.simulateMessageConsumption(ctx, topic, handler)

	return nil
}

// Unsubscribe unsubscribes from a topic
func (r *RabbitMQClient) Unsubscribe(ctx context.Context, topic string) error {
	r.logger.Info("Unsubscribing from topic", "topic", topic)

	delete(r.subscriptions, topic)

	return nil
}

// CreateDeadLetterQueue creates a dead letter queue for failed messages
func (r *RabbitMQClient) CreateDeadLetterQueue(ctx context.Context, topic string) error {
	deadLetterTopic := topic + ".dead-letter"
	
	r.logger.Info("Creating dead letter queue",
		"original_topic", topic,
		"dead_letter_topic", deadLetterTopic)

	// In a real implementation, this would create DLQ in RabbitMQ
	return nil
}

// HealthCheck performs a health check on the RabbitMQ connection
func (r *RabbitMQClient) HealthCheck(ctx context.Context) error {
	// In a real implementation, this would check RabbitMQ connection
	// For now, we simulate a successful health check
	return nil
}

// Close closes the RabbitMQ connection
func (r *RabbitMQClient) Close(ctx context.Context) error {
	r.logger.Info("Closing RabbitMQ connection")

	// Clear subscriptions
	r.subscriptions = make(map[string]MessageHandler)

	return nil
}

// simulateMessageConsumption simulates consuming messages from RabbitMQ
func (r *RabbitMQClient) simulateMessageConsumption(ctx context.Context, topic string, handler MessageHandler) {
	// This would be implemented with actual RabbitMQ consumer logic
	// For unit testing, we can inject mock messages through test helpers
}

// Notification Publishers

// EmailNotificationPublisher publishes email notification requests
type EmailNotificationPublisher interface {
	PublishEmailNotification(ctx context.Context, request *EmailNotificationRequest) error
}

// SMSNotificationPublisher publishes SMS notification requests  
type SMSNotificationPublisher interface {
	PublishSMSNotification(ctx context.Context, request *SMSNotificationRequest) error
}

// SlackNotificationPublisher publishes Slack notification requests
type SlackNotificationPublisher interface {
	PublishSlackNotification(ctx context.Context, request *SlackNotificationRequest) error
}

// RabbitMQEmailPublisher implements EmailNotificationPublisher
type RabbitMQEmailPublisher struct {
	messageQueue MessageQueueClient
	logger       *slog.Logger
	topic        string
}

// NewRabbitMQEmailPublisher creates a new email publisher
func NewRabbitMQEmailPublisher(messageQueue MessageQueueClient, logger *slog.Logger) *RabbitMQEmailPublisher {
	return &RabbitMQEmailPublisher{
		messageQueue: messageQueue,
		logger:       logger,
		topic:        "email-notifications",
	}
}

// PublishEmailNotification publishes an email notification request
func (p *RabbitMQEmailPublisher) PublishEmailNotification(ctx context.Context, request *EmailNotificationRequest) error {
	data, err := json.Marshal(request)
	if err != nil {
		p.logger.Error("Failed to marshal email notification request", "error", err)
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	headers := map[string]string{
		"content-type":   "application/json",
		"event-type":     request.EventType,
		"priority":       request.Priority,
		"correlation-id": request.CorrelationID,
	}

	err = p.messageQueue.Publish(ctx, p.topic, data, headers)
	if err != nil {
		p.logger.Error("Failed to publish email notification", 
			"correlation_id", request.CorrelationID,
			"recipients", len(request.Recipients),
			"error", err)
		return fmt.Errorf("failed to publish email notification: %w", err)
	}

	p.logger.Info("Published email notification",
		"correlation_id", request.CorrelationID,
		"recipients", len(request.Recipients),
		"event_type", request.EventType)

	return nil
}

// RabbitMQSMSPublisher implements SMSNotificationPublisher
type RabbitMQSMSPublisher struct {
	messageQueue MessageQueueClient
	logger       *slog.Logger
	topic        string
}

// NewRabbitMQSMSPublisher creates a new SMS publisher
func NewRabbitMQSMSPublisher(messageQueue MessageQueueClient, logger *slog.Logger) *RabbitMQSMSPublisher {
	return &RabbitMQSMSPublisher{
		messageQueue: messageQueue,
		logger:       logger,
		topic:        "sms-notifications",
	}
}

// PublishSMSNotification publishes an SMS notification request
func (p *RabbitMQSMSPublisher) PublishSMSNotification(ctx context.Context, request *SMSNotificationRequest) error {
	data, err := json.Marshal(request)
	if err != nil {
		p.logger.Error("Failed to marshal SMS notification request", "error", err)
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	headers := map[string]string{
		"content-type":   "application/json",
		"event-type":     request.EventType,
		"priority":       request.Priority,
		"correlation-id": request.CorrelationID,
	}

	err = p.messageQueue.Publish(ctx, p.topic, data, headers)
	if err != nil {
		p.logger.Error("Failed to publish SMS notification",
			"correlation_id", request.CorrelationID,
			"recipients", len(request.Recipients),
			"error", err)
		return fmt.Errorf("failed to publish SMS notification: %w", err)
	}

	p.logger.Info("Published SMS notification",
		"correlation_id", request.CorrelationID,
		"recipients", len(request.Recipients),
		"event_type", request.EventType)

	return nil
}

// RabbitMQSlackPublisher implements SlackNotificationPublisher
type RabbitMQSlackPublisher struct {
	messageQueue MessageQueueClient
	logger       *slog.Logger
	topic        string
}

// NewRabbitMQSlackPublisher creates a new Slack publisher
func NewRabbitMQSlackPublisher(messageQueue MessageQueueClient, logger *slog.Logger) *RabbitMQSlackPublisher {
	return &RabbitMQSlackPublisher{
		messageQueue: messageQueue,
		logger:       logger,
		topic:        "slack-notifications",
	}
}

// PublishSlackNotification publishes a Slack notification request
func (p *RabbitMQSlackPublisher) PublishSlackNotification(ctx context.Context, request *SlackNotificationRequest) error {
	data, err := json.Marshal(request)
	if err != nil {
		p.logger.Error("Failed to marshal Slack notification request", "error", err)
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	headers := map[string]string{
		"content-type":   "application/json",
		"event-type":     request.EventType,
		"priority":       request.Priority,
		"correlation-id": request.CorrelationID,
	}

	err = p.messageQueue.Publish(ctx, p.topic, data, headers)
	if err != nil {
		p.logger.Error("Failed to publish Slack notification",
			"correlation_id", request.CorrelationID,
			"channels", len(request.Channels),
			"error", err)
		return fmt.Errorf("failed to publish Slack notification: %w", err)
	}

	p.logger.Info("Published Slack notification",
		"correlation_id", request.CorrelationID,
		"channels", len(request.Channels),
		"event_type", request.EventType)

	return nil
}

// Notification Request Types (used for publishing to handlers)

// EmailNotificationRequest represents a request to send email notifications
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

// SMSNotificationRequest represents a request to send SMS notifications
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

// SlackNotificationRequest represents a request to send Slack notifications
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

// Health Status Types

// HealthStatus represents the health status of a service
type HealthStatus struct {
	ServiceName string                 `json:"service_name"`
	Status      string                 `json:"status"`
	Timestamp   time.Time              `json:"timestamp"`
	Checks      map[string]CheckResult `json:"checks"`
}

// CheckResult represents the result of a health check
type CheckResult struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}