package email

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/google/uuid"
)

// EmailHandlerService processes email notification requests
type EmailHandlerService struct {
	messageQueue     MessageQueueConsumer
	emailRepository  EmailRepository
	azureClient      AzureEmailClient
	templateRenderer EmailTemplateRenderer
	logger           *slog.Logger
	config           *EmailHandlerConfig
	workers          []*EmailWorker
	stopChan         chan struct{}
}

// EmailHandlerConfig contains configuration for the email handler
type EmailHandlerConfig struct {
	QueueName         string                `json:"queue_name"`
	Workers           int                   `json:"workers"`
	ProcessingDelay   time.Duration         `json:"processing_delay"`
	RetryDelay        time.Duration         `json:"retry_delay"`
	MaxRetries        int                   `json:"max_retries"`
	BatchSize         int                   `json:"batch_size"`
	DeadLetterEnabled bool                  `json:"dead_letter_enabled"`
	Azure             *AzureEmailConfig     `json:"azure"`
	Templates         *EmailTemplateConfig  `json:"templates"`
}

// EmailTemplateConfig contains email template configuration
type EmailTemplateConfig struct {
	DefaultTemplateID string            `json:"default_template_id"`
	TemplateCache     bool              `json:"template_cache"`
	TemplateTimeout   time.Duration     `json:"template_timeout"`
	Variables         map[string]string `json:"variables"`
}

// NewEmailHandlerService creates a new email handler service
func NewEmailHandlerService(
	messageQueue MessageQueueConsumer,
	emailRepository EmailRepository,
	azureClient AzureEmailClient,
	templateRenderer EmailTemplateRenderer,
	logger *slog.Logger,
	config *EmailHandlerConfig,
) *EmailHandlerService {
	return &EmailHandlerService{
		messageQueue:     messageQueue,
		emailRepository:  emailRepository,
		azureClient:      azureClient,
		templateRenderer: templateRenderer,
		logger:           logger,
		config:           config,
		workers:          make([]*EmailWorker, 0, config.Workers),
		stopChan:         make(chan struct{}),
	}
}

// Start initializes the email handler service and starts processing messages
func (e *EmailHandlerService) Start(ctx context.Context) error {
	e.logger.Info("Starting email handler service",
		"service", "email-handler",
		"workers", e.config.Workers,
		"queue", e.config.QueueName)

	// Validate configuration
	if err := e.validateConfiguration(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Initialize Azure client
	if err := e.azureClient.Initialize(ctx, e.config.Azure); err != nil {
		return fmt.Errorf("failed to initialize Azure client: %w", err)
	}

	// Start workers
	for i := 0; i < e.config.Workers; i++ {
		worker := NewEmailWorker(i, e, e.logger)
		e.workers = append(e.workers, worker)
		
		go func(w *EmailWorker) {
			if err := w.Start(ctx); err != nil {
				e.logger.Error("Worker failed", "worker_id", w.ID, "error", err)
			}
		}(worker)
	}

	// Subscribe to email notification queue
	if err := e.messageQueue.Subscribe(ctx, e.config.QueueName, e.handleEmailRequest); err != nil {
		return fmt.Errorf("failed to subscribe to queue %s: %w", e.config.QueueName, err)
	}

	e.logger.Info("Email handler service started successfully")
	return nil
}

// Stop gracefully shuts down the email handler service
func (e *EmailHandlerService) Stop(ctx context.Context) error {
	e.logger.Info("Stopping email handler service")
	
	// Signal stop to all workers
	close(e.stopChan)
	
	// Wait for workers to complete current processing
	for _, worker := range e.workers {
		worker.Stop(ctx)
	}

	// Unsubscribe from queue
	if err := e.messageQueue.Unsubscribe(ctx, e.config.QueueName); err != nil {
		e.logger.Error("Error unsubscribing from queue", "error", err)
	}

	e.logger.Info("Email handler service stopped successfully")
	return nil
}

// ProcessEmailRequest processes a single email notification request
func (e *EmailHandlerService) ProcessEmailRequest(ctx context.Context, request *EmailNotificationRequest) error {
	correlationID := request.CorrelationID
	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	logger := e.logger.With(
		"correlation_id", correlationID,
		"event_type", request.EventType,
		"recipients", len(request.Recipients),
	)

	ctx = domain.WithCorrelationID(ctx, correlationID)

	logger.Debug("Processing email notification request")

	// Create email message from request
	emailMessage, err := e.createEmailMessage(ctx, request)
	if err != nil {
		logger.Error("Failed to create email message", "error", err)
		return fmt.Errorf("failed to create email message: %w", err)
	}

	// Validate email message
	if !emailMessage.IsValid() {
		logger.Error("Invalid email message created")
		return domain.NewValidationError("invalid email message")
	}

	// Save email message to database
	if err := e.emailRepository.SaveMessage(ctx, emailMessage); err != nil {
		logger.Error("Failed to save email message", "error", err)
		return fmt.Errorf("failed to save email message: %w", err)
	}

	logger.Info("Created email message", "message_id", emailMessage.MessageID)

	// Send email via Azure Communication Services
	deliveryStatus, err := e.sendEmail(ctx, emailMessage)
	if err != nil {
		logger.Error("Failed to send email", "message_id", emailMessage.MessageID, "error", err)
		
		// Update delivery status as failed
		deliveryStatus = &EmailDeliveryStatus{
			MessageID:      emailMessage.MessageID,
			SubscriberID:   emailMessage.SubscriberID,
			Status:         DeliveryStatusFailed,
			AttemptCount:   1,
			LastAttemptAt:  time.Now().UTC(),
			ErrorMessage:   stringPtr(err.Error()),
		}
	}

	// Save delivery status
	if err := e.emailRepository.SaveDeliveryStatus(ctx, deliveryStatus); err != nil {
		logger.Error("Failed to save delivery status", "error", err)
		return fmt.Errorf("failed to save delivery status: %w", err)
	}

	// Schedule retry if needed
	if deliveryStatus.Status == DeliveryStatusFailed && deliveryStatus.AttemptCount < e.config.MaxRetries {
		if err := e.scheduleRetry(ctx, emailMessage, deliveryStatus); err != nil {
			logger.Error("Failed to schedule retry", "error", err)
		}
	}

	logger.Info("Email processing completed",
		"message_id", emailMessage.MessageID,
		"status", deliveryStatus.Status,
		"attempt_count", deliveryStatus.AttemptCount)

	return nil
}

// GetDeliveryStatus retrieves the delivery status for an email
func (e *EmailHandlerService) GetDeliveryStatus(ctx context.Context, messageID string) (*EmailDeliveryStatus, error) {
	return e.emailRepository.GetDeliveryStatus(ctx, messageID)
}

// RetryFailedEmail retries sending a failed email
func (e *EmailHandlerService) RetryFailedEmail(ctx context.Context, messageID string) error {
	logger := e.logger.With("message_id", messageID)

	// Get email message
	emailMessage, err := e.emailRepository.GetMessage(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to get email message: %w", err)
	}

	// Get current delivery status
	deliveryStatus, err := e.emailRepository.GetDeliveryStatus(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to get delivery status: %w", err)
	}

	// Check if retry is allowed
	if deliveryStatus.Status.IsFinalStatus() && deliveryStatus.Status != DeliveryStatusFailed {
		return domain.NewValidationError("email already delivered or cannot be retried")
	}

	if deliveryStatus.AttemptCount >= e.config.MaxRetries {
		return domain.NewValidationError("maximum retry attempts exceeded")
	}

	logger.Info("Retrying failed email", "attempt", deliveryStatus.AttemptCount+1)

	// Send email
	newDeliveryStatus, err := e.sendEmail(ctx, emailMessage)
	if err != nil {
		logger.Error("Retry attempt failed", "error", err)
		
		// Update failure status
		deliveryStatus.AttemptCount++
		deliveryStatus.LastAttemptAt = time.Now().UTC()
		deliveryStatus.ErrorMessage = stringPtr(err.Error())
		deliveryStatus.Status = DeliveryStatusFailed
		
		if err := e.emailRepository.UpdateDeliveryStatus(ctx, deliveryStatus); err != nil {
			logger.Error("Failed to update delivery status after retry failure", "error", err)
		}
		
		return fmt.Errorf("retry failed: %w", err)
	}

	// Update delivery status
	if err := e.emailRepository.UpdateDeliveryStatus(ctx, newDeliveryStatus); err != nil {
		logger.Error("Failed to update delivery status after successful retry", "error", err)
		return fmt.Errorf("failed to update delivery status: %w", err)
	}

	logger.Info("Email retry successful", "status", newDeliveryStatus.Status)
	return nil
}

// GetHealthStatus returns the health status of the email handler
func (e *EmailHandlerService) GetHealthStatus(ctx context.Context) (*HealthStatus, error) {
	status := &HealthStatus{
		ServiceName: "email-handler",
		Status:      "healthy",
		Timestamp:   time.Now().UTC(),
		Checks:      make(map[string]CheckResult),
	}

	// Check email repository
	if err := e.emailRepository.HealthCheck(ctx); err != nil {
		status.Checks["email_repository"] = CheckResult{
			Status: "unhealthy",
			Error:  err.Error(),
		}
		status.Status = "unhealthy"
	} else {
		status.Checks["email_repository"] = CheckResult{
			Status: "healthy",
		}
	}

	// Check Azure Communication Services
	if err := e.azureClient.HealthCheck(ctx); err != nil {
		status.Checks["azure_email_client"] = CheckResult{
			Status: "unhealthy",
			Error:  err.Error(),
		}
		status.Status = "unhealthy"
	} else {
		status.Checks["azure_email_client"] = CheckResult{
			Status: "healthy",
		}
	}

	// Check message queue
	if err := e.messageQueue.HealthCheck(ctx); err != nil {
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
	for _, worker := range e.workers {
		if worker.IsActive() {
			activeWorkers++
		}
	}
	
	status.Checks["workers"] = CheckResult{
		Status: "healthy",
		Details: map[string]interface{}{
			"configured_workers": len(e.workers),
			"active_workers":     activeWorkers,
		},
	}

	if activeWorkers == 0 && len(e.workers) > 0 {
		status.Checks["workers"] = CheckResult{
			Status: "unhealthy",
			Error:  "no active workers",
		}
		status.Status = "unhealthy"
	}

	return status, nil
}

// Private helper methods

// handleEmailRequest handles incoming email requests from the message queue
func (e *EmailHandlerService) handleEmailRequest(ctx context.Context, message *QueueMessage) error {
	// Parse email notification request
	var request EmailNotificationRequest
	if err := json.Unmarshal(message.Data, &request); err != nil {
		e.logger.Error("Failed to parse email request", "error", err)
		return fmt.Errorf("failed to parse email request: %w", err)
	}

	// Set correlation ID from message
	if request.CorrelationID == "" {
		request.CorrelationID = message.CorrelationID
	}

	// Process with timeout
	processCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return e.ProcessEmailRequest(processCtx, &request)
}

// createEmailMessage creates an email message from a notification request
func (e *EmailHandlerService) createEmailMessage(ctx context.Context, request *EmailNotificationRequest) (*EmailMessage, error) {
	// Generate unique message ID
	messageID := uuid.New().String()

	// Get template ID for event type
	templateID := GetTemplateIDByEventType(request.EventType)

	// Generate subject
	subject := GenerateSubjectByEventType(request.EventType, request.EventData)

	// Create template data
	templateData := &EmailTemplateData{
		EventType:         request.EventType,
		Priority:          request.Priority,
		EntityID:          domain.ExtractString(request.EventData, "entity_id"),
		UserID:            domain.ExtractString(request.EventData, "user_id"),
		Timestamp:         request.CreatedAt.Format(time.RFC3339),
		CorrelationID:     request.CorrelationID,
		EventData:         request.EventData,
		ActionURL:         generateActionURL(request.EventType, request.EventData),
		UnsubscribeURL:    generateUnsubscribeURL(request.SubscriberID),
	}

	// Render email content
	htmlContent, textContent, err := e.templateRenderer.RenderTemplate(ctx, templateID, templateData)
	if err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	// Create email message
	emailMessage := &EmailMessage{
		MessageID:           messageID,
		SubscriberID:        request.SubscriberID,
		Recipients:          request.Recipients,
		Subject:             subject,
		HtmlContent:         htmlContent,
		TextContent:         textContent,
		EventType:           request.EventType,
		Priority:            request.Priority,
		EventData:           request.EventData,
		TemplateID:          templateID,
		PersonalizationData: map[string]interface{}{
			"event_type":     request.EventType,
			"priority":       request.Priority,
			"correlation_id": request.CorrelationID,
		},
		CreatedAt:     time.Now().UTC(),
		CorrelationID: request.CorrelationID,
	}

	return emailMessage, nil
}

// sendEmail sends an email via Azure Communication Services
func (e *EmailHandlerService) sendEmail(ctx context.Context, message *EmailMessage) (*EmailDeliveryStatus, error) {
	logger := e.logger.With("message_id", message.MessageID)

	// Create Azure send request
	sendRequest := &AzureSendEmailRequest{
		SenderAddress: e.config.Azure.SenderAddress,
		Recipients: AzureRecipients{
			To: make([]AzureRecipient, len(message.Recipients)),
		},
		Content: AzureEmailContent{
			Subject:   message.Subject,
			PlainText: message.TextContent,
			Html:      message.HtmlContent,
		},
		Headers:       make(map[string]string),
		ReplyTo:       e.config.Azure.ReplyToAddress,
		AttachmentIds: []string{}, // No attachments for notifications
	}

	// Add recipients
	for i, recipient := range message.Recipients {
		sendRequest.Recipients.To[i] = AzureRecipient{
			Address:     recipient,
			DisplayName: "", // Use email address as display name
		}
	}

	// Add custom headers
	sendRequest.Headers["X-Message-ID"] = message.MessageID
	sendRequest.Headers["X-Correlation-ID"] = message.CorrelationID
	sendRequest.Headers["X-Event-Type"] = message.EventType
	sendRequest.Headers["X-Priority"] = message.Priority

	// Send email
	azureResponse, err := e.azureClient.SendEmail(ctx, sendRequest)
	if err != nil {
		logger.Error("Azure send email failed", "error", err)
		return nil, fmt.Errorf("Azure send email failed: %w", err)
	}

	// Create recipient statuses
	recipientStatuses := make([]RecipientStatus, len(message.Recipients))
	for i, recipient := range message.Recipients {
		recipientStatuses[i] = RecipientStatus{
			Email:       recipient,
			Status:      DeliveryStatusSent,
			DeliveredAt: nil, // Will be updated by webhook
		}
	}

	// Create delivery status
	deliveryStatus := &EmailDeliveryStatus{
		MessageID:     message.MessageID,
		SubscriberID:  message.SubscriberID,
		Recipients:    recipientStatuses,
		Status:        DeliveryStatusSent,
		AttemptCount:  1,
		LastAttemptAt: time.Now().UTC(),
	}

	logger.Info("Email sent successfully",
		"azure_message_id", azureResponse.MessageID,
		"recipients", len(message.Recipients))

	return deliveryStatus, nil
}

// scheduleRetry schedules a retry for a failed email
func (e *EmailHandlerService) scheduleRetry(ctx context.Context, message *EmailMessage, status *EmailDeliveryStatus) error {
	nextRetryAt := time.Now().UTC().Add(e.config.RetryDelay * time.Duration(status.AttemptCount))
	
	status.NextRetryAt = &nextRetryAt

	if err := e.emailRepository.UpdateDeliveryStatus(ctx, status); err != nil {
		return fmt.Errorf("failed to update delivery status with retry time: %w", err)
	}

	e.logger.Info("Scheduled email retry",
		"message_id", message.MessageID,
		"next_retry_at", nextRetryAt,
		"attempt", status.AttemptCount)

	// In a full implementation, this would schedule the retry with a job scheduler
	// For now, we just update the database with the retry time
	return nil
}

// validateConfiguration validates the email handler configuration
func (e *EmailHandlerService) validateConfiguration() error {
	if e.config == nil {
		return domain.NewValidationError("configuration cannot be nil")
	}

	if e.config.QueueName == "" {
		return domain.NewValidationError("queue name is required")
	}

	if e.config.Workers <= 0 {
		return domain.NewValidationError("workers must be positive")
	}

	if e.config.MaxRetries < 0 {
		return domain.NewValidationError("max retries cannot be negative")
	}

	if e.config.Azure == nil {
		return domain.NewValidationError("Azure configuration is required")
	}

	if e.config.Azure.ConnectionString == "" {
		return domain.NewValidationError("Azure connection string is required")
	}

	if e.config.Azure.SenderAddress == "" {
		return domain.NewValidationError("Azure sender address is required")
	}

	return nil
}

// Helper functions

func stringPtr(s string) *string {
	return &s
}

func generateActionURL(eventType string, eventData map[string]interface{}) string {
	// Generate appropriate action URL based on event type
	switch eventType {
	case "inquiry-business", "inquiry-media", "inquiry-donations", "inquiry-volunteers":
		if entityID := domain.ExtractString(eventData, "entity_id"); entityID != "" {
			return fmt.Sprintf("https://admin.international-center.app/inquiries/%s", entityID)
		}
		return "https://admin.international-center.app/inquiries"
	case "event-registration":
		if entityID := domain.ExtractString(eventData, "entity_id"); entityID != "" {
			return fmt.Sprintf("https://admin.international-center.app/content/%s", entityID)
		}
		return "https://admin.international-center.app/content"
	case "system-error", "capacity-alert", "compliance-alert":
		return "https://admin.international-center.app/alerts"
	case "admin-action-required":
		return "https://admin.international-center.app/admin"
	default:
		return "https://admin.international-center.app"
	}
}

func generateUnsubscribeURL(subscriberID string) string {
	return fmt.Sprintf("https://admin.international-center.app/notifications/unsubscribe/%s", subscriberID)
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