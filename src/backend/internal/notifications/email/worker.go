package email

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// EmailWorker handles concurrent processing of email messages
type EmailWorker struct {
	ID           int
	service      *EmailHandlerService
	logger       *slog.Logger
	active       bool
	mutex        sync.RWMutex
	stopChan     chan struct{}
	workQueue    chan *EmailNotificationRequest
	wg           sync.WaitGroup
}

// NewEmailWorker creates a new email worker
func NewEmailWorker(id int, service *EmailHandlerService, logger *slog.Logger) *EmailWorker {
	return &EmailWorker{
		ID:        id,
		service:   service,
		logger:    logger.With("worker_id", id, "component", "email-worker"),
		active:    false,
		stopChan:  make(chan struct{}),
		workQueue: make(chan *EmailNotificationRequest, 100), // Buffered channel
	}
}

// Start starts the email worker
func (w *EmailWorker) Start(ctx context.Context) error {
	w.mutex.Lock()
	w.active = true
	w.mutex.Unlock()

	w.logger.Info("Starting email worker")

	// Start the main processing loop
	w.wg.Add(1)
	go w.processLoop(ctx)

	// Start the retry processing loop
	w.wg.Add(1)
	go w.retryLoop(ctx)

	return nil
}

// Stop stops the email worker gracefully
func (w *EmailWorker) Stop(ctx context.Context) error {
	w.logger.Info("Stopping email worker")

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
		w.logger.Info("Email worker stopped gracefully")
	case <-time.After(10 * time.Second):
		w.logger.Warn("Email worker stop timeout, some operations may still be running")
	}

	return nil
}

// IsActive returns whether the worker is currently active
func (w *EmailWorker) IsActive() bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.active
}

// QueueEmail queues an email request for processing
func (w *EmailWorker) QueueEmail(request *EmailNotificationRequest) error {
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
func (w *EmailWorker) GetQueueLength() int {
	return len(w.workQueue)
}

// GetMetrics returns worker metrics
func (w *EmailWorker) GetMetrics() WorkerMetrics {
	return WorkerMetrics{
		WorkerID:    w.ID,
		Active:      w.IsActive(),
		QueueLength: w.GetQueueLength(),
		QueueCapacity: cap(w.workQueue),
	}
}

// processLoop is the main processing loop for handling email requests
func (w *EmailWorker) processLoop(ctx context.Context) {
	defer w.wg.Done()

	w.logger.Debug("Email worker processing loop started")

	for {
		select {
		case <-w.stopChan:
			w.logger.Debug("Email worker processing loop stopped")
			return

		case <-ctx.Done():
			w.logger.Debug("Email worker processing loop cancelled")
			return

		case request := <-w.workQueue:
			w.processEmailRequest(ctx, request)

		case <-time.After(30 * time.Second):
			// Periodic health check and maintenance
			w.performMaintenance(ctx)
		}
	}
}

// retryLoop handles retry processing for failed emails
func (w *EmailWorker) retryLoop(ctx context.Context) {
	defer w.wg.Done()

	w.logger.Debug("Email worker retry loop started")

	ticker := time.NewTicker(60 * time.Second) // Check for retries every minute
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			w.logger.Debug("Email worker retry loop stopped")
			return

		case <-ctx.Done():
			w.logger.Debug("Email worker retry loop cancelled")
			return

		case <-ticker.C:
			w.processRetries(ctx)
		}
	}
}

// processEmailRequest processes a single email request
func (w *EmailWorker) processEmailRequest(ctx context.Context, request *EmailNotificationRequest) {
	logger := w.logger.With(
		"correlation_id", request.CorrelationID,
		"event_type", request.EventType,
		"recipients", len(request.Recipients))

	logger.Debug("Processing email request")

	startTime := time.Now()

	// Add processing delay if configured
	if w.service.config.ProcessingDelay > 0 {
		select {
		case <-time.After(w.service.config.ProcessingDelay):
		case <-ctx.Done():
			logger.Debug("Processing cancelled during delay")
			return
		}
	}

	// Process the email request
	err := w.service.ProcessEmailRequest(ctx, request)
	
	processingTime := time.Since(startTime)

	if err != nil {
		logger.Error("Failed to process email request",
			"error", err,
			"processing_time", processingTime)

		// Handle error - could implement dead letter queue logic here
		w.handleProcessingError(ctx, request, err)
	} else {
		logger.Info("Successfully processed email request",
			"processing_time", processingTime)
	}
}

// processRetries processes retry requests for failed emails
func (w *EmailWorker) processRetries(ctx context.Context) {
	logger := w.logger.With("operation", "retry_processing")
	
	logger.Debug("Processing email retries")

	// Get failed messages that are ready for retry
	failedMessages, err := w.service.emailRepository.GetFailedMessages(ctx, 10)
	if err != nil {
		logger.Error("Failed to get failed messages for retry", "error", err)
		return
	}

	if len(failedMessages) == 0 {
		logger.Debug("No failed messages to retry")
		return
	}

	logger.Info("Found failed messages to retry", "count", len(failedMessages))

	for _, message := range failedMessages {
		// Check if it's time to retry
		status, err := w.service.emailRepository.GetDeliveryStatus(ctx, message.MessageID)
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
			logger.Debug("Message exceeded max retries",
				"message_id", message.MessageID,
				"attempts", status.AttemptCount)
			continue // Exceeded max retries
		}

		// Attempt retry
		logger.Info("Retrying failed email",
			"message_id", message.MessageID,
			"attempt", status.AttemptCount+1)

		if err := w.service.RetryFailedEmail(ctx, message.MessageID); err != nil {
			logger.Error("Retry attempt failed",
				"message_id", message.MessageID,
				"error", err)
		}
	}
}

// performMaintenance performs periodic maintenance tasks
func (w *EmailWorker) performMaintenance(ctx context.Context) {
	logger := w.logger.With("operation", "maintenance")
	
	logger.Debug("Performing worker maintenance")

	// Log worker statistics
	metrics := w.GetMetrics()
	logger.Debug("Worker metrics",
		"queue_length", metrics.QueueLength,
		"queue_capacity", metrics.QueueCapacity,
		"queue_utilization", float64(metrics.QueueLength)/float64(metrics.QueueCapacity)*100)

	// Check for stale messages or other maintenance tasks
	// This could include cleanup operations, metrics collection, etc.
	
	// Health check the dependencies
	if err := w.service.azureClient.HealthCheck(ctx); err != nil {
		logger.Warn("Azure client health check failed during maintenance", "error", err)
	}
}

// handleProcessingError handles errors that occur during email processing
func (w *EmailWorker) handleProcessingError(ctx context.Context, request *EmailNotificationRequest, processingError error) {
	logger := w.logger.With(
		"correlation_id", request.CorrelationID,
		"error_type", fmt.Sprintf("%T", processingError))

	// Categorize the error to determine handling strategy
	switch {
	case isDependencyError(processingError):
		// Dependency errors (Azure API down, database issues) should be retried
		logger.Warn("Dependency error encountered, will retry",
			"error", processingError)
		// Error will be handled by retry logic
		
	case isValidationError(processingError):
		// Validation errors are permanent and should not be retried
		logger.Error("Validation error encountered, moving to dead letter queue",
			"error", processingError)
		w.sendToDeadLetterQueue(ctx, request, processingError)
		
	case isRateLimitError(processingError):
		// Rate limit errors should be retried with backoff
		logger.Warn("Rate limit error encountered, will retry with backoff",
			"error", processingError)
		// Implement exponential backoff in retry logic
		
	default:
		// Unknown errors should be logged and retried with caution
		logger.Error("Unknown error encountered during processing",
			"error", processingError)
	}
}

// sendToDeadLetterQueue sends a message to the dead letter queue
func (w *EmailWorker) sendToDeadLetterQueue(ctx context.Context, request *EmailNotificationRequest, err error) {
	logger := w.logger.With("correlation_id", request.CorrelationID)
	
	logger.Info("Sending message to dead letter queue")

	// In a full implementation, this would publish to a dead letter queue
	// For now, we just log the action
	deadLetterMessage := DeadLetterMessage{
		OriginalRequest: *request,
		Error:          err.Error(),
		Timestamp:      time.Now().UTC(),
		WorkerID:       w.ID,
		Reason:         "processing_failed",
	}

	logger.Error("Message sent to dead letter queue",
		"message_id", deadLetterMessage.OriginalRequest.CorrelationID,
		"error", deadLetterMessage.Error,
		"reason", deadLetterMessage.Reason)
}

// Error classification helper functions
func isDependencyError(err error) bool {
	// Check if error is a dependency-related error
	// This would check the error type or message
	return false // Simplified for now
}

func isValidationError(err error) bool {
	// Check if error is a validation error
	// This would check the error type or message
	return false // Simplified for now
}

func isRateLimitError(err error) bool {
	// Check if error is a rate limit error
	// This would check the error type or message
	return false // Simplified for now
}

// Supporting Types

// WorkerMetrics represents metrics for a worker
type WorkerMetrics struct {
	WorkerID      int  `json:"worker_id"`
	Active        bool `json:"active"`
	QueueLength   int  `json:"queue_length"`
	QueueCapacity int  `json:"queue_capacity"`
}

// DeadLetterMessage represents a message that failed processing
type DeadLetterMessage struct {
	OriginalRequest EmailNotificationRequest `json:"original_request"`
	Error          string                   `json:"error"`
	Timestamp      time.Time                `json:"timestamp"`
	WorkerID       int                      `json:"worker_id"`
	Reason         string                   `json:"reason"`
}

// EmailNotificationRequest represents a request to send an email notification
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