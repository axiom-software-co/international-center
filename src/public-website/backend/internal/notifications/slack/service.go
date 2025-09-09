package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/google/uuid"
)

// SlackHandlerService processes Slack notification requests
type SlackHandlerService struct {
	messageQueue     MessageQueueConsumer
	slackRepository  SlackRepository
	slackClient      SlackAPIClient
	logger           *slog.Logger
	config           *SlackHandlerConfig
	workers          []*SlackWorker
	stopChan         chan struct{}
}

// SlackHandlerConfig contains configuration for the Slack handler
type SlackHandlerConfig struct {
	QueueName         string          `json:"queue_name"`
	Workers           int             `json:"workers"`
	ProcessingDelay   time.Duration   `json:"processing_delay"`
	RetryDelay        time.Duration   `json:"retry_delay"`
	MaxRetries        int             `json:"max_retries"`
	BatchSize         int             `json:"batch_size"`
	DeadLetterEnabled bool            `json:"dead_letter_enabled"`
	Slack             *SlackConfig    `json:"slack"`
}

// NewSlackHandlerService creates a new Slack handler service
func NewSlackHandlerService(
	messageQueue MessageQueueConsumer,
	slackRepository SlackRepository,
	slackClient SlackAPIClient,
	logger *slog.Logger,
	config *SlackHandlerConfig,
) *SlackHandlerService {
	return &SlackHandlerService{
		messageQueue:    messageQueue,
		slackRepository: slackRepository,
		slackClient:     slackClient,
		logger:          logger,
		config:          config,
		workers:         make([]*SlackWorker, 0, config.Workers),
		stopChan:        make(chan struct{}),
	}
}

// Start initializes the Slack handler service and starts processing messages
func (s *SlackHandlerService) Start(ctx context.Context) error {
	s.logger.Info("Starting Slack handler service",
		"service", "slack-handler",
		"workers", s.config.Workers,
		"queue", s.config.QueueName)

	// Validate configuration
	if err := s.validateConfiguration(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Initialize Slack client
	if err := s.slackClient.Initialize(ctx, s.config.Slack); err != nil {
		return fmt.Errorf("failed to initialize Slack client: %w", err)
	}

	// Start workers
	for i := 0; i < s.config.Workers; i++ {
		worker := NewSlackWorker(i, s, s.logger)
		s.workers = append(s.workers, worker)
		
		go func(w *SlackWorker) {
			if err := w.Start(ctx); err != nil {
				s.logger.Error("Worker failed", "worker_id", w.ID, "error", err)
			}
		}(worker)
	}

	// Subscribe to Slack notification queue
	if err := s.messageQueue.Subscribe(ctx, s.config.QueueName, s.handleSlackRequest); err != nil {
		return fmt.Errorf("failed to subscribe to queue %s: %w", s.config.QueueName, err)
	}

	s.logger.Info("Slack handler service started successfully")
	return nil
}

// Stop gracefully shuts down the Slack handler service
func (s *SlackHandlerService) Stop(ctx context.Context) error {
	s.logger.Info("Stopping Slack handler service")
	
	// Signal stop to all workers
	close(s.stopChan)
	
	// Wait for workers to complete current processing
	for _, worker := range s.workers {
		worker.Stop(ctx)
	}

	// Unsubscribe from queue
	if err := s.messageQueue.Unsubscribe(ctx, s.config.QueueName); err != nil {
		s.logger.Error("Error unsubscribing from queue", "error", err)
	}

	s.logger.Info("Slack handler service stopped successfully")
	return nil
}

// ProcessSlackRequest processes a single Slack notification request
func (s *SlackHandlerService) ProcessSlackRequest(ctx context.Context, request *SlackNotificationRequest) error {
	correlationID := request.CorrelationID
	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	logger := s.logger.With(
		"correlation_id", correlationID,
		"event_type", request.EventType,
		"channels", len(request.Channels),
	)

	ctx = domain.WithCorrelationID(ctx, correlationID)

	logger.Debug("Processing Slack notification request")

	// Create Slack message from request
	slackMessage, err := s.createSlackMessage(ctx, request)
	if err != nil {
		logger.Error("Failed to create Slack message", "error", err)
		return fmt.Errorf("failed to create Slack message: %w", err)
	}

	// Validate Slack message
	if !slackMessage.IsValid() {
		logger.Error("Invalid Slack message created")
		return domain.NewValidationError("invalid Slack message")
	}

	// Save Slack message to database
	if err := s.slackRepository.SaveMessage(ctx, slackMessage); err != nil {
		logger.Error("Failed to save Slack message", "error", err)
		return fmt.Errorf("failed to save Slack message: %w", err)
	}

	logger.Info("Created Slack message", "message_id", slackMessage.MessageID)

	// Send message via Slack API
	deliveryStatus, err := s.sendSlackMessage(ctx, slackMessage)
	if err != nil {
		logger.Error("Failed to send Slack message", "message_id", slackMessage.MessageID, "error", err)
		
		// Update delivery status as failed
		deliveryStatus = &SlackDeliveryStatus{
			MessageID:      slackMessage.MessageID,
			SubscriberID:   slackMessage.SubscriberID,
			Status:         SlackStatusFailed,
			AttemptCount:   1,
			LastAttemptAt:  time.Now().UTC(),
			ErrorMessage:   stringPtr(err.Error()),
		}
	}

	// Save delivery status
	if err := s.slackRepository.SaveDeliveryStatus(ctx, deliveryStatus); err != nil {
		logger.Error("Failed to save delivery status", "error", err)
		return fmt.Errorf("failed to save delivery status: %w", err)
	}

	// Schedule retry if needed
	if deliveryStatus.Status == SlackStatusFailed && deliveryStatus.AttemptCount < s.config.MaxRetries {
		if err := s.scheduleRetry(ctx, slackMessage, deliveryStatus); err != nil {
			logger.Error("Failed to schedule retry", "error", err)
		}
	}

	logger.Info("Slack processing completed",
		"message_id", slackMessage.MessageID,
		"status", deliveryStatus.Status,
		"attempt_count", deliveryStatus.AttemptCount)

	return nil
}

// GetDeliveryStatus retrieves the delivery status for a Slack message
func (s *SlackHandlerService) GetDeliveryStatus(ctx context.Context, messageID string) (*SlackDeliveryStatus, error) {
	return s.slackRepository.GetDeliveryStatus(ctx, messageID)
}

// RetryFailedSlackMessage retries sending a failed Slack message
func (s *SlackHandlerService) RetryFailedSlackMessage(ctx context.Context, messageID string) error {
	logger := s.logger.With("message_id", messageID)

	// Get Slack message
	slackMessage, err := s.slackRepository.GetMessage(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to get Slack message: %w", err)
	}

	// Get current delivery status
	deliveryStatus, err := s.slackRepository.GetDeliveryStatus(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to get delivery status: %w", err)
	}

	// Check if retry is allowed
	if deliveryStatus.Status.IsFinalStatus() && deliveryStatus.Status != SlackStatusFailed {
		return domain.NewValidationError("Slack message already delivered or cannot be retried")
	}

	if deliveryStatus.AttemptCount >= s.config.MaxRetries {
		return domain.NewValidationError("maximum retry attempts exceeded")
	}

	logger.Info("Retrying failed Slack message", "attempt", deliveryStatus.AttemptCount+1)

	// Send Slack message
	newDeliveryStatus, err := s.sendSlackMessage(ctx, slackMessage)
	if err != nil {
		logger.Error("Retry attempt failed", "error", err)
		
		// Update failure status
		deliveryStatus.AttemptCount++
		deliveryStatus.LastAttemptAt = time.Now().UTC()
		deliveryStatus.ErrorMessage = stringPtr(err.Error())
		deliveryStatus.Status = SlackStatusFailed
		
		if err := s.slackRepository.UpdateDeliveryStatus(ctx, deliveryStatus); err != nil {
			logger.Error("Failed to update delivery status after retry failure", "error", err)
		}
		
		return fmt.Errorf("retry failed: %w", err)
	}

	// Update delivery status
	if err := s.slackRepository.UpdateDeliveryStatus(ctx, newDeliveryStatus); err != nil {
		logger.Error("Failed to update delivery status after successful retry", "error", err)
		return fmt.Errorf("failed to update delivery status: %w", err)
	}

	logger.Info("Slack retry successful", "status", newDeliveryStatus.Status)
	return nil
}

// GetHealthStatus returns the health status of the Slack handler
func (s *SlackHandlerService) GetHealthStatus(ctx context.Context) (*HealthStatus, error) {
	status := &HealthStatus{
		ServiceName: "slack-handler",
		Status:      "healthy",
		Timestamp:   time.Now().UTC(),
		Checks:      make(map[string]CheckResult),
	}

	// Check Slack repository
	if err := s.slackRepository.HealthCheck(ctx); err != nil {
		status.Checks["slack_repository"] = CheckResult{
			Status: "unhealthy",
			Error:  err.Error(),
		}
		status.Status = "unhealthy"
	} else {
		status.Checks["slack_repository"] = CheckResult{
			Status: "healthy",
		}
	}

	// Check Slack API client
	if err := s.slackClient.HealthCheck(ctx); err != nil {
		status.Checks["slack_api_client"] = CheckResult{
			Status: "unhealthy",
			Error:  err.Error(),
		}
		status.Status = "unhealthy"
	} else {
		status.Checks["slack_api_client"] = CheckResult{
			Status: "healthy",
		}
	}

	// Check message queue
	if err := s.messageQueue.HealthCheck(ctx); err != nil {
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

	// Add worker status
	activeWorkers := 0
	for _, worker := range s.workers {
		if worker.IsActive() {
			activeWorkers++
		}
	}
	
	status.Checks["workers"] = CheckResult{
		Status: "healthy",
		Details: map[string]interface{}{
			"configured_workers": len(s.workers),
			"active_workers":     activeWorkers,
		},
	}

	if activeWorkers == 0 && len(s.workers) > 0 {
		status.Checks["workers"] = CheckResult{
			Status: "unhealthy",
			Error:  "no active workers",
		}
		status.Status = "unhealthy"
	}

	return status, nil
}

// Private helper methods

// handleSlackRequest handles incoming Slack requests from the message queue
func (s *SlackHandlerService) handleSlackRequest(ctx context.Context, message *QueueMessage) error {
	// Parse Slack notification request
	var request SlackNotificationRequest
	if err := json.Unmarshal(message.Data, &request); err != nil {
		s.logger.Error("Failed to parse Slack request", "error", err)
		return fmt.Errorf("failed to parse Slack request: %w", err)
	}

	// Set correlation ID from message
	if request.CorrelationID == "" {
		request.CorrelationID = message.CorrelationID
	}

	// Process with timeout
	processCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return s.ProcessSlackRequest(processCtx, &request)
}

// createSlackMessage creates a Slack message from a notification request
func (s *SlackHandlerService) createSlackMessage(ctx context.Context, request *SlackNotificationRequest) (*SlackMessage, error) {
	// Generate unique message ID
	messageID := uuid.New().String()

	// Validate and resolve channels
	validChannels := make([]string, 0, len(request.Channels))
	for _, channel := range request.Channels {
		if IsValidSlackChannel(channel) {
			validChannels = append(validChannels, channel)
		} else {
			s.logger.Warn("Invalid Slack channel in request", 
				"channel", channel, 
				"correlation_id", request.CorrelationID)
		}
	}

	// Add default channels if none are valid or specified
	if len(validChannels) == 0 {
		defaultChannels := GetChannelsForEventType(request.EventType)
		validChannels = defaultChannels
	}

	// Generate Slack content based on event type
	content := GenerateSlackContent(request.EventType, request.EventData)

	// Truncate content if needed
	content = TruncateSlackContent(content, MaxSlackMessageLength)

	// Generate rich message attachment
	attachment := GenerateSlackAttachment(request.EventType, request.EventData, request.Priority)

	// Create Slack message
	slackMessage := &SlackMessage{
		MessageID:     messageID,
		SubscriberID:  request.SubscriberID,
		Channels:      validChannels,
		Content:       content,
		EventType:     request.EventType,
		Priority:      request.Priority,
		EventData:     request.EventData,
		Attachments:   []SlackAttachment{attachment},
		CreatedAt:     time.Now().UTC(),
		CorrelationID: request.CorrelationID,
	}

	return slackMessage, nil
}

// sendSlackMessage sends a message via Slack API
func (s *SlackHandlerService) sendSlackMessage(ctx context.Context, message *SlackMessage) (*SlackDeliveryStatus, error) {
	logger := s.logger.With("message_id", message.MessageID)

	// Create channel statuses
	channelStatuses := make([]SlackChannelStatus, len(message.Channels))
	
	// Send to each channel
	for i, channel := range message.Channels {
		channelLogger := logger.With("channel", channel)
		
		// Create Slack API request
		sendRequest := &SlackSendMessageRequest{
			Channel:     channel,
			Text:        message.Content,
			Attachments: message.Attachments,
			Blocks:      message.Blocks,
			Username:    "International Center Bot",
			IconEmoji:   ":bell:",
			Metadata: map[string]string{
				"message-id":     message.MessageID,
				"correlation-id": message.CorrelationID,
				"event-type":     message.EventType,
				"priority":       message.Priority,
			},
		}

		// Send to Slack
		slackResponse, err := s.slackClient.SendMessage(ctx, sendRequest)
		if err != nil {
			channelLogger.Error("Failed to send Slack message to channel", "error", err)
			
			channelStatuses[i] = SlackChannelStatus{
				Channel:      channel,
				Status:       SlackStatusFailed,
				ErrorMessage: stringPtr(err.Error()),
			}
		} else {
			channelLogger.Info("Slack message sent successfully", "slack_ts", slackResponse.MessageTS)
			
			channelStatuses[i] = SlackChannelStatus{
				Channel:     channel,
				Status:      SlackStatusDelivered,
				MessageTS:   slackResponse.MessageTS,
				DeliveredAt: timePtr(time.Now().UTC()),
			}
		}
	}

	// Determine overall status
	overallStatus := SlackStatusDelivered
	failedCount := 0
	for _, channelStatus := range channelStatuses {
		if channelStatus.Status == SlackStatusFailed {
			failedCount++
		}
	}

	if failedCount == len(channelStatuses) {
		overallStatus = SlackStatusFailed
	} else if failedCount > 0 {
		overallStatus = SlackStatusSent // Partial delivery
	}

	// Create delivery status
	deliveryStatus := &SlackDeliveryStatus{
		MessageID:     message.MessageID,
		SubscriberID:  message.SubscriberID,
		Channels:      channelStatuses,
		Status:        overallStatus,
		AttemptCount:  1,
		LastAttemptAt: time.Now().UTC(),
	}

	if overallStatus == SlackStatusDelivered {
		deliveryStatus.DeliveredAt = timePtr(time.Now().UTC())
	}

	logger.Info("Slack message processing completed",
		"channels", len(message.Channels),
		"successful", len(channelStatuses)-failedCount,
		"failed", failedCount,
		"overall_status", overallStatus)

	return deliveryStatus, nil
}

// scheduleRetry schedules a retry for a failed Slack message
func (s *SlackHandlerService) scheduleRetry(ctx context.Context, message *SlackMessage, status *SlackDeliveryStatus) error {
	nextRetryAt := time.Now().UTC().Add(s.config.RetryDelay * time.Duration(status.AttemptCount))
	
	status.NextRetryAt = &nextRetryAt

	if err := s.slackRepository.UpdateDeliveryStatus(ctx, status); err != nil {
		return fmt.Errorf("failed to update delivery status with retry time: %w", err)
	}

	s.logger.Info("Scheduled Slack retry",
		"message_id", message.MessageID,
		"next_retry_at", nextRetryAt,
		"attempt", status.AttemptCount)

	return nil
}

// validateConfiguration validates the Slack handler configuration
func (s *SlackHandlerService) validateConfiguration() error {
	if s.config == nil {
		return domain.NewValidationError("configuration cannot be nil")
	}

	if s.config.QueueName == "" {
		return domain.NewValidationError("queue name is required")
	}

	if s.config.Workers <= 0 {
		return domain.NewValidationError("workers must be positive")
	}

	if s.config.MaxRetries < 0 {
		return domain.NewValidationError("max retries cannot be negative")
	}

	if s.config.Slack == nil {
		return domain.NewValidationError("Slack configuration is required")
	}

	if s.config.Slack.BotToken == "" {
		return domain.NewValidationError("Slack bot token is required")
	}

	if s.config.Slack.DefaultChannel == "" {
		return domain.NewValidationError("Slack default channel is required")
	}

	// Validate default channel format
	if !IsValidSlackChannel(s.config.Slack.DefaultChannel) {
		return domain.NewValidationError("Slack default channel is not valid")
	}

	return nil
}

// Helper functions

func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// Supporting Interfaces and Types

// MessageQueueConsumer interface for consuming messages
type MessageQueueConsumer interface {
	Subscribe(ctx context.Context, queueName string, handler func(context.Context, *QueueMessage) error) error
	Unsubscribe(ctx context.Context, queueName string) error
	HealthCheck(ctx context.Context) error
}

// SlackRepository interface for Slack data persistence
type SlackRepository interface {
	SaveMessage(ctx context.Context, message *SlackMessage) error
	GetMessage(ctx context.Context, messageID string) (*SlackMessage, error)
	UpdateMessage(ctx context.Context, message *SlackMessage) error
	DeleteMessage(ctx context.Context, messageID string) error
	SaveDeliveryStatus(ctx context.Context, status *SlackDeliveryStatus) error
	GetDeliveryStatus(ctx context.Context, messageID string) (*SlackDeliveryStatus, error)
	UpdateDeliveryStatus(ctx context.Context, status *SlackDeliveryStatus) error
	GetPendingMessages(ctx context.Context) ([]*SlackMessage, error)
	GetFailedMessages(ctx context.Context, limit int) ([]*SlackMessage, error)
	HealthCheck(ctx context.Context) error
}


// Queue Message Types

// QueueMessage represents a message from the queue
type QueueMessage struct {
	ID            string            `json:"id"`
	Data          []byte            `json:"data"`
	Headers       map[string]string `json:"headers"`
	CorrelationID string            `json:"correlation_id"`
	Timestamp     time.Time         `json:"timestamp"`
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
	Status  string                 `json:"status"`
	Error   string                 `json:"error,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Slack API Types

// SlackSendMessageRequest represents a send message request to Slack
type SlackSendMessageRequest struct {
	Channel     string            `json:"channel"`
	Text        string            `json:"text"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
	Blocks      []SlackBlock      `json:"blocks,omitempty"`
	Username    string            `json:"username,omitempty"`
	IconEmoji   string            `json:"icon_emoji,omitempty"`
	IconURL     string            `json:"icon_url,omitempty"`
	ThreadTS    string            `json:"thread_ts,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// SlackSendMessageResponse represents a send message response from Slack
type SlackSendMessageResponse struct {
	OK        bool   `json:"ok"`
	Channel   string `json:"channel"`
	MessageTS string `json:"ts"`
	Error     string `json:"error,omitempty"`
}

// SlackUpdateMessageRequest represents an update message request
type SlackUpdateMessageRequest struct {
	Channel     string            `json:"channel"`
	MessageTS   string            `json:"ts"`
	Text        string            `json:"text"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
	Blocks      []SlackBlock      `json:"blocks,omitempty"`
}

// SlackUpdateMessageResponse represents an update message response
type SlackUpdateMessageResponse struct {
	OK        bool   `json:"ok"`
	Channel   string `json:"channel"`
	MessageTS string `json:"ts"`
	Error     string `json:"error,omitempty"`
}

// SlackChannelInfo represents channel information from Slack
type SlackChannelInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	IsIM    bool   `json:"is_im"`
	IsGroup bool   `json:"is_group"`
	IsMember bool  `json:"is_member"`
}