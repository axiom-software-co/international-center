package sms

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/google/uuid"
)

// SMSHandlerService processes SMS notification requests
type SMSHandlerService struct {
	messageQueue    MessageQueueConsumer
	smsRepository   SMSRepository
	azureClient     AzureSMSClient
	logger          *slog.Logger
	config          *SMSHandlerConfig
	workers         []*SMSWorker
	stopChan        chan struct{}
}

// SMSHandlerConfig contains configuration for the SMS handler
type SMSHandlerConfig struct {
	QueueName         string            `json:"queue_name"`
	Workers           int               `json:"workers"`
	ProcessingDelay   time.Duration     `json:"processing_delay"`
	RetryDelay        time.Duration     `json:"retry_delay"`
	MaxRetries        int               `json:"max_retries"`
	BatchSize         int               `json:"batch_size"`
	DeadLetterEnabled bool              `json:"dead_letter_enabled"`
	Azure             *AzureSMSConfig   `json:"azure"`
}

// NewSMSHandlerService creates a new SMS handler service
func NewSMSHandlerService(
	messageQueue MessageQueueConsumer,
	smsRepository SMSRepository,
	azureClient AzureSMSClient,
	logger *slog.Logger,
	config *SMSHandlerConfig,
) *SMSHandlerService {
	return &SMSHandlerService{
		messageQueue:  messageQueue,
		smsRepository: smsRepository,
		azureClient:   azureClient,
		logger:        logger,
		config:        config,
		workers:       make([]*SMSWorker, 0, config.Workers),
		stopChan:      make(chan struct{}),
	}
}

// Start initializes the SMS handler service and starts processing messages
func (s *SMSHandlerService) Start(ctx context.Context) error {
	s.logger.Info("Starting SMS handler service",
		"service", "sms-handler",
		"workers", s.config.Workers,
		"queue", s.config.QueueName)

	// Validate configuration
	if err := s.validateConfiguration(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Initialize Azure client
	if err := s.azureClient.Initialize(ctx, s.config.Azure); err != nil {
		return fmt.Errorf("failed to initialize Azure client: %w", err)
	}

	// Start workers
	for i := 0; i < s.config.Workers; i++ {
		worker := NewSMSWorker(i, s, s.logger)
		s.workers = append(s.workers, worker)
		
		go func(w *SMSWorker) {
			if err := w.Start(ctx); err != nil {
				s.logger.Error("Worker failed", "worker_id", w.ID, "error", err)
			}
		}(worker)
	}

	// Subscribe to SMS notification queue
	if err := s.messageQueue.Subscribe(ctx, s.config.QueueName, s.handleSMSRequest); err != nil {
		return fmt.Errorf("failed to subscribe to queue %s: %w", s.config.QueueName, err)
	}

	s.logger.Info("SMS handler service started successfully")
	return nil
}

// Stop gracefully shuts down the SMS handler service
func (s *SMSHandlerService) Stop(ctx context.Context) error {
	s.logger.Info("Stopping SMS handler service")
	
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

	s.logger.Info("SMS handler service stopped successfully")
	return nil
}

// ProcessSMSRequest processes a single SMS notification request
func (s *SMSHandlerService) ProcessSMSRequest(ctx context.Context, request *SMSNotificationRequest) error {
	correlationID := request.CorrelationID
	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	logger := s.logger.With(
		"correlation_id", correlationID,
		"event_type", request.EventType,
		"recipients", len(request.Recipients),
	)

	ctx = domain.WithCorrelationID(ctx, correlationID)

	logger.Debug("Processing SMS notification request")

	// Create SMS message from request
	smsMessage, err := s.createSMSMessage(ctx, request)
	if err != nil {
		logger.Error("Failed to create SMS message", "error", err)
		return fmt.Errorf("failed to create SMS message: %w", err)
	}

	// Validate SMS message
	if !smsMessage.IsValid() {
		logger.Error("Invalid SMS message created")
		return domain.NewValidationError("invalid SMS message")
	}

	// Save SMS message to database
	if err := s.smsRepository.SaveMessage(ctx, smsMessage); err != nil {
		logger.Error("Failed to save SMS message", "error", err)
		return fmt.Errorf("failed to save SMS message: %w", err)
	}

	logger.Info("Created SMS message", "message_id", smsMessage.MessageID)

	// Send SMS via Azure Communication Services
	deliveryStatus, err := s.sendSMS(ctx, smsMessage)
	if err != nil {
		logger.Error("Failed to send SMS", "message_id", smsMessage.MessageID, "error", err)
		
		// Update delivery status as failed
		deliveryStatus = &SMSDeliveryStatus{
			MessageID:      smsMessage.MessageID,
			SubscriberID:   smsMessage.SubscriberID,
			Status:         SMSStatusFailed,
			AttemptCount:   1,
			LastAttemptAt:  time.Now().UTC(),
			ErrorMessage:   stringPtr(err.Error()),
		}
	}

	// Save delivery status
	if err := s.smsRepository.SaveDeliveryStatus(ctx, deliveryStatus); err != nil {
		logger.Error("Failed to save delivery status", "error", err)
		return fmt.Errorf("failed to save delivery status: %w", err)
	}

	// Schedule retry if needed
	if deliveryStatus.Status == SMSStatusFailed && deliveryStatus.AttemptCount < s.config.MaxRetries {
		if err := s.scheduleRetry(ctx, smsMessage, deliveryStatus); err != nil {
			logger.Error("Failed to schedule retry", "error", err)
		}
	}

	logger.Info("SMS processing completed",
		"message_id", smsMessage.MessageID,
		"status", deliveryStatus.Status,
		"attempt_count", deliveryStatus.AttemptCount)

	return nil
}

// GetDeliveryStatus retrieves the delivery status for an SMS
func (s *SMSHandlerService) GetDeliveryStatus(ctx context.Context, messageID string) (*SMSDeliveryStatus, error) {
	return s.smsRepository.GetDeliveryStatus(ctx, messageID)
}

// RetryFailedSMS retries sending a failed SMS
func (s *SMSHandlerService) RetryFailedSMS(ctx context.Context, messageID string) error {
	logger := s.logger.With("message_id", messageID)

	// Get SMS message
	smsMessage, err := s.smsRepository.GetMessage(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to get SMS message: %w", err)
	}

	// Get current delivery status
	deliveryStatus, err := s.smsRepository.GetDeliveryStatus(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to get delivery status: %w", err)
	}

	// Check if retry is allowed
	if deliveryStatus.Status.IsFinalStatus() && deliveryStatus.Status != SMSStatusFailed {
		return domain.NewValidationError("SMS already delivered or cannot be retried")
	}

	if deliveryStatus.AttemptCount >= s.config.MaxRetries {
		return domain.NewValidationError("maximum retry attempts exceeded")
	}

	logger.Info("Retrying failed SMS", "attempt", deliveryStatus.AttemptCount+1)

	// Send SMS
	newDeliveryStatus, err := s.sendSMS(ctx, smsMessage)
	if err != nil {
		logger.Error("Retry attempt failed", "error", err)
		
		// Update failure status
		deliveryStatus.AttemptCount++
		deliveryStatus.LastAttemptAt = time.Now().UTC()
		deliveryStatus.ErrorMessage = stringPtr(err.Error())
		deliveryStatus.Status = SMSStatusFailed
		
		if err := s.smsRepository.UpdateDeliveryStatus(ctx, deliveryStatus); err != nil {
			logger.Error("Failed to update delivery status after retry failure", "error", err)
		}
		
		return fmt.Errorf("retry failed: %w", err)
	}

	// Update delivery status
	if err := s.smsRepository.UpdateDeliveryStatus(ctx, newDeliveryStatus); err != nil {
		logger.Error("Failed to update delivery status after successful retry", "error", err)
		return fmt.Errorf("failed to update delivery status: %w", err)
	}

	logger.Info("SMS retry successful", "status", newDeliveryStatus.Status)
	return nil
}

// GetHealthStatus returns the health status of the SMS handler
func (s *SMSHandlerService) GetHealthStatus(ctx context.Context) (*HealthStatus, error) {
	status := &HealthStatus{
		ServiceName: "sms-handler",
		Status:      "healthy",
		Timestamp:   time.Now().UTC(),
		Checks:      make(map[string]CheckResult),
	}

	// Check SMS repository
	if err := s.smsRepository.HealthCheck(ctx); err != nil {
		status.Checks["sms_repository"] = CheckResult{
			Status: "unhealthy",
			Error:  err.Error(),
		}
		status.Status = "unhealthy"
	} else {
		status.Checks["sms_repository"] = CheckResult{
			Status: "healthy",
		}
	}

	// Check Azure Communication Services
	if err := s.azureClient.HealthCheck(ctx); err != nil {
		status.Checks["azure_sms_client"] = CheckResult{
			Status: "unhealthy",
			Error:  err.Error(),
		}
		status.Status = "unhealthy"
	} else {
		status.Checks["azure_sms_client"] = CheckResult{
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

// handleSMSRequest handles incoming SMS requests from the message queue
func (s *SMSHandlerService) handleSMSRequest(ctx context.Context, message *QueueMessage) error {
	// Parse SMS notification request
	var request SMSNotificationRequest
	if err := json.Unmarshal(message.Data, &request); err != nil {
		s.logger.Error("Failed to parse SMS request", "error", err)
		return fmt.Errorf("failed to parse SMS request: %w", err)
	}

	// Set correlation ID from message
	if request.CorrelationID == "" {
		request.CorrelationID = message.CorrelationID
	}

	// Process with timeout
	processCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return s.ProcessSMSRequest(processCtx, &request)
}

// createSMSMessage creates an SMS message from a notification request
func (s *SMSHandlerService) createSMSMessage(ctx context.Context, request *SMSNotificationRequest) (*SMSMessage, error) {
	// Generate unique message ID
	messageID := uuid.New().String()

	// Validate and format phone numbers
	validRecipients := make([]string, 0, len(request.Recipients))
	for _, phone := range request.Recipients {
		if IsValidUSPhoneNumber(phone) {
			formatted := FormatPhoneNumberE164(phone)
			validRecipients = append(validRecipients, formatted)
		} else {
			s.logger.Warn("Invalid phone number in request", 
				"phone", phone, 
				"correlation_id", request.CorrelationID)
		}
	}

	if len(validRecipients) == 0 {
		return nil, domain.NewValidationError("no valid phone numbers found in request")
	}

	// Generate SMS content based on event type
	content := GenerateSMSContent(request.EventType, request.EventData)

	// Truncate content to SMS limits
	content = TruncateSMSContent(content, MaxSMSLength)

	// Create SMS message
	smsMessage := &SMSMessage{
		MessageID:     messageID,
		SubscriberID:  request.SubscriberID,
		Recipients:    validRecipients,
		Content:       content,
		EventType:     request.EventType,
		Priority:      request.Priority,
		EventData:     request.EventData,
		CreatedAt:     time.Now().UTC(),
		CorrelationID: request.CorrelationID,
	}

	return smsMessage, nil
}

// sendSMS sends an SMS via Azure Communication Services
func (s *SMSHandlerService) sendSMS(ctx context.Context, message *SMSMessage) (*SMSDeliveryStatus, error) {
	logger := s.logger.With("message_id", message.MessageID)

	// Create Azure send request
	sendRequest := &AzureSendSMSRequest{
		From:    s.config.Azure.FromNumber,
		To:      message.Recipients,
		Message: message.Content,
		DeliveryReportEnabled: true,
		Tags: map[string]string{
			"message-id":     message.MessageID,
			"correlation-id": message.CorrelationID,
			"event-type":     message.EventType,
			"priority":       message.Priority,
		},
	}

	// Send SMS
	azureResponse, err := s.azureClient.SendSMS(ctx, sendRequest)
	if err != nil {
		logger.Error("Azure send SMS failed", "error", err)
		return nil, fmt.Errorf("Azure send SMS failed: %w", err)
	}

	// Create recipient statuses
	recipientStatuses := make([]SMSRecipientStatus, len(message.Recipients))
	for i, recipient := range message.Recipients {
		recipientStatuses[i] = SMSRecipientStatus{
			PhoneNumber: recipient,
			Status:      SMSStatusSent,
			DeliveredAt: nil, // Will be updated by webhook
		}
	}

	// Create delivery status
	deliveryStatus := &SMSDeliveryStatus{
		MessageID:     message.MessageID,
		SubscriberID:  message.SubscriberID,
		Recipients:    recipientStatuses,
		Status:        SMSStatusSent,
		AttemptCount:  1,
		LastAttemptAt: time.Now().UTC(),
	}

	logger.Info("SMS sent successfully",
		"azure_message_id", azureResponse.MessageID,
		"recipients", len(message.Recipients),
		"content_length", len(message.Content))

	return deliveryStatus, nil
}

// scheduleRetry schedules a retry for a failed SMS
func (s *SMSHandlerService) scheduleRetry(ctx context.Context, message *SMSMessage, status *SMSDeliveryStatus) error {
	nextRetryAt := time.Now().UTC().Add(s.config.RetryDelay * time.Duration(status.AttemptCount))
	
	status.NextRetryAt = &nextRetryAt

	if err := s.smsRepository.UpdateDeliveryStatus(ctx, status); err != nil {
		return fmt.Errorf("failed to update delivery status with retry time: %w", err)
	}

	s.logger.Info("Scheduled SMS retry",
		"message_id", message.MessageID,
		"next_retry_at", nextRetryAt,
		"attempt", status.AttemptCount)

	return nil
}

// validateConfiguration validates the SMS handler configuration
func (s *SMSHandlerService) validateConfiguration() error {
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

	if s.config.Azure == nil {
		return domain.NewValidationError("Azure configuration is required")
	}

	if s.config.Azure.ConnectionString == "" {
		return domain.NewValidationError("Azure connection string is required")
	}

	if s.config.Azure.FromNumber == "" {
		return domain.NewValidationError("Azure from number is required")
	}

	// Validate from number format
	if !IsValidUSPhoneNumber(s.config.Azure.FromNumber) {
		return domain.NewValidationError("Azure from number is not a valid US phone number")
	}

	return nil
}

// Helper functions

func stringPtr(s string) *string {
	return &s
}

// Supporting Interfaces and Types

// MessageQueueConsumer interface for consuming messages
type MessageQueueConsumer interface {
	Subscribe(ctx context.Context, queueName string, handler func(context.Context, *QueueMessage) error) error
	Unsubscribe(ctx context.Context, queueName string) error
	HealthCheck(ctx context.Context) error
}

// SMSRepository interface for SMS data persistence
type SMSRepository interface {
	SaveMessage(ctx context.Context, message *SMSMessage) error
	GetMessage(ctx context.Context, messageID string) (*SMSMessage, error)
	UpdateMessage(ctx context.Context, message *SMSMessage) error
	DeleteMessage(ctx context.Context, messageID string) error
	SaveDeliveryStatus(ctx context.Context, status *SMSDeliveryStatus) error
	GetDeliveryStatus(ctx context.Context, messageID string) (*SMSDeliveryStatus, error)
	UpdateDeliveryStatus(ctx context.Context, status *SMSDeliveryStatus) error
	GetPendingMessages(ctx context.Context) ([]*SMSMessage, error)
	GetFailedMessages(ctx context.Context, limit int) ([]*SMSMessage, error)
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

// Azure SMS API Types

// AzureSendSMSRequest represents a send SMS request to Azure
type AzureSendSMSRequest struct {
	From                  string            `json:"from"`
	To                    []string          `json:"to"`
	Message               string            `json:"message"`
	DeliveryReportEnabled bool              `json:"enableDeliveryReport"`
	Tags                  map[string]string `json:"tags,omitempty"`
}

// AzureSendSMSResponse represents a send SMS response from Azure
type AzureSendSMSResponse struct {
	MessageID string                     `json:"messageId"`
	To        []AzureSMSRecipientResult  `json:"to"`
}

// AzureSMSRecipientResult represents recipient result from Azure
type AzureSMSRecipientResult struct {
	To               string `json:"to"`
	MessageID        string `json:"messageId"`
	HttpStatusCode   int    `json:"httpStatusCode"`
	ErrorMessage     string `json:"errorMessage,omitempty"`
	RepeatabilityResult string `json:"repeatabilityResult,omitempty"`
	Successful       bool   `json:"successful"`
}

// AzureSMSDeliveryStatus represents delivery status from Azure
type AzureSMSDeliveryStatus struct {
	MessageID    string                     `json:"messageId"`
	From         string                     `json:"from"`
	To           string                     `json:"to"`
	DeliveryStatus string                   `json:"deliveryStatus"`
	DeliveryStatusDetails AzureSMSDeliveryDetails `json:"deliveryStatusDetails"`
	ReceivedTimestamp     time.Time              `json:"receivedTimestamp"`
}

// AzureSMSDeliveryDetails represents detailed SMS delivery information
type AzureSMSDeliveryDetails struct {
	StatusMessage string    `json:"statusMessage"`
	Timestamp     time.Time `json:"timestamp"`
}