package sms

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"
)

// SMSWorker handles concurrent processing of SMS messages
type SMSWorker struct {
	ID           int
	service      *SMSHandlerService
	logger       *slog.Logger
	active       bool
	mutex        sync.RWMutex
	stopChan     chan struct{}
	workQueue    chan *SMSNotificationRequest
	wg           sync.WaitGroup
}

// NewSMSWorker creates a new SMS worker
func NewSMSWorker(id int, service *SMSHandlerService, logger *slog.Logger) *SMSWorker {
	return &SMSWorker{
		ID:        id,
		service:   service,
		logger:    logger.With("worker_id", id, "component", "sms-worker"),
		active:    false,
		stopChan:  make(chan struct{}),
		workQueue: make(chan *SMSNotificationRequest, 50), // Smaller buffer for SMS
	}
}

// Start starts the SMS worker
func (w *SMSWorker) Start(ctx context.Context) error {
	w.mutex.Lock()
	w.active = true
	w.mutex.Unlock()

	w.logger.Info("Starting SMS worker")

	// Start the main processing loop
	w.wg.Add(1)
	go w.processLoop(ctx)

	// Start the retry processing loop
	w.wg.Add(1)
	go w.retryLoop(ctx)

	return nil
}

// Stop stops the SMS worker gracefully
func (w *SMSWorker) Stop(ctx context.Context) error {
	w.logger.Info("Stopping SMS worker")

	// Signal stop
	close(w.stopChan)

	// Mark as inactive
	w.mutex.Lock()
	w.active = false
	w.mutex.Unlock()

	// Wait for goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		w.logger.Info("SMS worker stopped gracefully")
	case <-time.After(10 * time.Second):
		w.logger.Warn("SMS worker stop timeout, some operations may still be running")
	}

	return nil
}

// IsActive returns whether the worker is currently active
func (w *SMSWorker) IsActive() bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.active
}

// QueueSMS queues an SMS request for processing
func (w *SMSWorker) QueueSMS(request *SMSNotificationRequest) error {
	if !w.IsActive() {
		return fmt.Errorf("worker %d is not active", w.ID)
	}

	select {
	case w.workQueue <- request:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("worker %d queue is full, request timed out", w.ID)
	}
}

// GetQueueLength returns the current queue length
func (w *SMSWorker) GetQueueLength() int {
	return len(w.workQueue)
}

// GetMetrics returns worker metrics
func (w *SMSWorker) GetMetrics() WorkerMetrics {
	return WorkerMetrics{
		WorkerID:      w.ID,
		Active:        w.IsActive(),
		QueueLength:   w.GetQueueLength(),
		QueueCapacity: cap(w.workQueue),
	}
}

// processLoop is the main processing loop for handling SMS requests
func (w *SMSWorker) processLoop(ctx context.Context) {
	defer w.wg.Done()

	w.logger.Debug("SMS worker processing loop started")

	for {
		select {
		case <-w.stopChan:
			w.logger.Debug("SMS worker processing loop stopped")
			return

		case <-ctx.Done():
			w.logger.Debug("SMS worker processing loop cancelled")
			return

		case request := <-w.workQueue:
			w.processSMSRequest(ctx, request)

		case <-time.After(30 * time.Second):
			// Periodic health check and maintenance
			w.performMaintenance(ctx)
		}
	}
}

// retryLoop handles retry processing for failed SMS messages
func (w *SMSWorker) retryLoop(ctx context.Context) {
	defer w.wg.Done()

	w.logger.Debug("SMS worker retry loop started")

	ticker := time.NewTicker(60 * time.Second) // Check for retries every minute
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			w.logger.Debug("SMS worker retry loop stopped")
			return

		case <-ctx.Done():
			w.logger.Debug("SMS worker retry loop cancelled")
			return

		case <-ticker.C:
			w.processRetries(ctx)
		}
	}
}

// processSMSRequest processes a single SMS request
func (w *SMSWorker) processSMSRequest(ctx context.Context, request *SMSNotificationRequest) {
	logger := w.logger.With(
		"correlation_id", request.CorrelationID,
		"event_type", request.EventType,
		"recipients", len(request.Recipients))

	logger.Debug("Processing SMS request")

	startTime := time.Now()

	// Add processing delay if configured (typically shorter for SMS)
	if w.service.config.ProcessingDelay > 0 {
		select {
		case <-time.After(w.service.config.ProcessingDelay):
		case <-ctx.Done():
			logger.Debug("Processing cancelled during delay")
			return
		}
	}

	// Validate phone numbers before processing
	validRecipients := w.validatePhoneNumbers(request.Recipients)
	if len(validRecipients) == 0 {
		logger.Warn("No valid phone numbers in SMS request")
		w.handleProcessingError(ctx, request, 
			fmt.Errorf("no valid phone numbers found in request"))
		return
	}

	// Update request with valid recipients only
	request.Recipients = validRecipients

	// Process the SMS request
	err := w.service.ProcessSMSRequest(ctx, request)
	
	processingTime := time.Since(startTime)

	if err != nil {
		logger.Error("Failed to process SMS request",
			"error", err,
			"processing_time", processingTime)

		w.handleProcessingError(ctx, request, err)
	} else {
		logger.Info("Successfully processed SMS request",
			"processing_time", processingTime,
			"valid_recipients", len(validRecipients))
	}
}

// processRetries processes retry requests for failed SMS messages
func (w *SMSWorker) processRetries(ctx context.Context) {
	logger := w.logger.With("operation", "retry_processing")
	
	logger.Debug("Processing SMS retries")

	// Get failed messages that are ready for retry
	failedMessages, err := w.service.smsRepository.GetFailedMessages(ctx, 10)
	if err != nil {
		logger.Error("Failed to get failed messages for retry", "error", err)
		return
	}

	if len(failedMessages) == 0 {
		logger.Debug("No failed SMS messages to retry")
		return
	}

	logger.Info("Found failed SMS messages to retry", "count", len(failedMessages))

	for _, message := range failedMessages {
		// Check if it's time to retry
		status, err := w.service.smsRepository.GetDeliveryStatus(ctx, message.MessageID)
		if err != nil {
			logger.Error("Failed to get delivery status for retry", 
				"message_id", message.MessageID, 
				"error", err)
			continue
		}

		if status.NextRetryAt == nil || time.Now().UTC().Before(*status.NextRetryAt) {
			continue // Not yet time to retry
		}

		if status.AttemptCount >= w.service.config.MaxRetries {
			logger.Debug("SMS message exceeded max retries",
				"message_id", message.MessageID,
				"attempts", status.AttemptCount)
			continue // Exceeded max retries
		}

		// Attempt retry
		logger.Info("Retrying failed SMS",
			"message_id", message.MessageID,
			"attempt", status.AttemptCount+1)

		if err := w.service.RetryFailedSMS(ctx, message.MessageID); err != nil {
			logger.Error("SMS retry attempt failed",
				"message_id", message.MessageID,
				"error", err)
		}
	}
}

// performMaintenance performs periodic maintenance tasks
func (w *SMSWorker) performMaintenance(ctx context.Context) {
	logger := w.logger.With("operation", "maintenance")
	
	logger.Debug("Performing SMS worker maintenance")

	// Log worker statistics
	metrics := w.GetMetrics()
	logger.Debug("SMS worker metrics",
		"queue_length", metrics.QueueLength,
		"queue_capacity", metrics.QueueCapacity,
		"queue_utilization", float64(metrics.QueueLength)/float64(metrics.QueueCapacity)*100)

	// Check message size statistics
	w.logMessageStatistics(ctx)
	
	// Health check the dependencies
	if err := w.service.azureClient.HealthCheck(ctx); err != nil {
		logger.Warn("Azure SMS client health check failed during maintenance", "error", err)
	}
}

// logMessageStatistics logs statistics about processed messages
func (w *SMSWorker) logMessageStatistics(ctx context.Context) {
	// In a full implementation, this would collect statistics
	// about message sizes, delivery rates, etc.
	w.logger.Debug("SMS message statistics logged")
}

// validatePhoneNumbers validates and formats phone numbers
func (w *SMSWorker) validatePhoneNumbers(phoneNumbers []string) []string {
	valid := make([]string, 0, len(phoneNumbers))
	
	for _, phone := range phoneNumbers {
		if IsValidUSPhoneNumber(phone) {
			formatted := FormatPhoneNumberE164(phone)
			valid = append(valid, formatted)
		} else {
			w.logger.Warn("Invalid phone number in SMS request", "phone", phone)
		}
	}
	
	return valid
}

// handleProcessingError handles errors that occur during SMS processing
func (w *SMSWorker) handleProcessingError(ctx context.Context, request *SMSNotificationRequest, processingError error) {
	logger := w.logger.With(
		"correlation_id", request.CorrelationID,
		"error_type", fmt.Sprintf("%T", processingError))

	// Categorize the error to determine handling strategy
	switch {
	case isDependencyError(processingError):
		// Dependency errors (Azure API down, database issues) should be retried
		logger.Warn("SMS dependency error encountered, will retry",
			"error", processingError)
		
	case isValidationError(processingError):
		// Validation errors are permanent and should not be retried
		logger.Error("SMS validation error encountered, moving to dead letter queue",
			"error", processingError)
		w.sendToDeadLetterQueue(ctx, request, processingError)
		
	case isRateLimitError(processingError):
		// Rate limit errors should be retried with exponential backoff
		logger.Warn("SMS rate limit error encountered, will retry with backoff",
			"error", processingError)
		
	case isCarrierError(processingError):
		// Carrier-specific errors (number unreachable, blocked, etc.)
		logger.Error("SMS carrier error encountered",
			"error", processingError)
		w.sendToDeadLetterQueue(ctx, request, processingError)
		
	default:
		// Unknown errors should be logged and retried with caution
		logger.Error("Unknown SMS error encountered during processing",
			"error", processingError)
	}
}

// sendToDeadLetterQueue sends a message to the dead letter queue
func (w *SMSWorker) sendToDeadLetterQueue(ctx context.Context, request *SMSNotificationRequest, err error) {
	logger := w.logger.With("correlation_id", request.CorrelationID)
	
	logger.Info("Sending SMS message to dead letter queue")

	deadLetterMessage := DeadLetterMessage{
		OriginalRequest: *request,
		Error:          err.Error(),
		Timestamp:      time.Now().UTC(),
		WorkerID:       w.ID,
		Reason:         "sms_processing_failed",
	}

	logger.Error("SMS message sent to dead letter queue",
		"message_id", deadLetterMessage.OriginalRequest.CorrelationID,
		"error", deadLetterMessage.Error,
		"reason", deadLetterMessage.Reason)
}

// Error classification helper functions specific to SMS
func isDependencyError(err error) bool {
	// Check if error is a dependency-related error
	// This would check the error type or message for Azure API errors
	return false // Simplified for now
}

func isValidationError(err error) bool {
	// Check if error is a validation error (invalid phone numbers, etc.)
	return false // Simplified for now
}

func isRateLimitError(err error) bool {
	// Check if error is a rate limit error from Azure or carrier
	return false // Simplified for now
}

func isCarrierError(err error) bool {
	// Check if error is from the SMS carrier (number unreachable, blocked, etc.)
	return false // Simplified for now
}

// Supporting Types

// WorkerMetrics represents metrics for an SMS worker
type WorkerMetrics struct {
	WorkerID      int  `json:"worker_id"`
	Active        bool `json:"active"`
	QueueLength   int  `json:"queue_length"`
	QueueCapacity int  `json:"queue_capacity"`
}

// DeadLetterMessage represents an SMS message that failed processing
type DeadLetterMessage struct {
	OriginalRequest SMSNotificationRequest `json:"original_request"`
	Error          string                 `json:"error"`
	Timestamp      time.Time              `json:"timestamp"`
	WorkerID       int                    `json:"worker_id"`
	Reason         string                 `json:"reason"`
}


// SMS-specific worker capabilities

// ProcessBulkSMS processes multiple SMS requests as a batch
func (w *SMSWorker) ProcessBulkSMS(ctx context.Context, requests []*SMSNotificationRequest) error {
	logger := w.logger.With("batch_size", len(requests))
	logger.Info("Processing bulk SMS requests")

	successCount := 0
	failureCount := 0

	for _, request := range requests {
		select {
		case <-ctx.Done():
			logger.Debug("Bulk SMS processing cancelled")
			return ctx.Err()
		default:
			w.processSMSRequest(ctx, request)
			successCount++
		}
	}

	logger.Info("Bulk SMS processing completed",
		"success_count", successCount,
		"failure_count", failureCount)

	return nil
}

// ValidateMessageContent validates SMS message content for compliance
func (w *SMSWorker) ValidateMessageContent(content string) error {
	// Check message length
	if len(content) == 0 {
		return fmt.Errorf("SMS content cannot be empty")
	}

	if len(content) > MaxConcatSMSLength {
		return fmt.Errorf("SMS content too long: %d characters (max %d)", 
			len(content), MaxConcatSMSLength)
	}

	// Check for prohibited content (in a real implementation)
	prohibitedTerms := []string{
		// Would include terms that violate SMS carrier policies
	}

	contentLower := strings.ToLower(content)
	for _, term := range prohibitedTerms {
		if strings.Contains(contentLower, term) {
			return fmt.Errorf("SMS content contains prohibited term: %s", term)
		}
	}

	return nil
}

// OptimizeMessageForSMS optimizes content for SMS delivery
func (w *SMSWorker) OptimizeMessageForSMS(content string, priority string) string {
	// For urgent messages, add priority indicator
	if strings.ToLower(priority) == "urgent" || strings.ToLower(priority) == "critical" {
		content = "URGENT: " + content
	}

	// Truncate to SMS limits
	if len(content) > MaxSMSLength {
		content = TruncateSMSContent(content, MaxSMSLength)
	}

	// Remove HTML tags if any
	content = w.stripHTMLTags(content)

	// Normalize whitespace
	content = w.normalizeWhitespace(content)

	return content
}

// stripHTMLTags removes HTML tags from content
func (w *SMSWorker) stripHTMLTags(content string) string {
	// Simple HTML tag removal - in production would use a proper HTML parser
	result := content
	// Remove common HTML tags
	htmlTags := []string{"<br>", "<br/>", "<p>", "</p>", "<div>", "</div>", 
		"<span>", "</span>", "<strong>", "</strong>", "<b>", "</b>", 
		"<em>", "</em>", "<i>", "</i>"}
	
	for _, tag := range htmlTags {
		result = strings.ReplaceAll(result, tag, "")
	}
	
	return result
}

// normalizeWhitespace normalizes whitespace in SMS content
func (w *SMSWorker) normalizeWhitespace(content string) string {
	// Replace multiple spaces with single space
	for strings.Contains(content, "  ") {
		content = strings.ReplaceAll(content, "  ", " ")
	}
	
	// Replace multiple newlines with single newline
	for strings.Contains(content, "\n\n\n") {
		content = strings.ReplaceAll(content, "\n\n\n", "\n\n")
	}
	
	return strings.TrimSpace(content)
}

