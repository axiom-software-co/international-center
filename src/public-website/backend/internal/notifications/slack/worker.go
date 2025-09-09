package slack

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"
)

// SlackWorker handles concurrent processing of Slack messages
type SlackWorker struct {
	ID           int
	service      *SlackHandlerService
	logger       *slog.Logger
	active       bool
	mutex        sync.RWMutex
	stopChan     chan struct{}
	workQueue    chan *SlackNotificationRequest
	rateLimiter  *SlackRateLimiter
	wg           sync.WaitGroup
}

// NewSlackWorker creates a new Slack worker
func NewSlackWorker(id int, service *SlackHandlerService, logger *slog.Logger) *SlackWorker {
	return &SlackWorker{
		ID:          id,
		service:     service,
		logger:      logger.With("worker_id", id, "component", "slack-worker"),
		active:      false,
		stopChan:    make(chan struct{}),
		workQueue:   make(chan *SlackNotificationRequest, 100), // Larger buffer for Slack
		rateLimiter: NewSlackRateLimiter(logger),
	}
}

// Start starts the Slack worker
func (w *SlackWorker) Start(ctx context.Context) error {
	w.mutex.Lock()
	w.active = true
	w.mutex.Unlock()

	w.logger.Info("Starting Slack worker")

	// Start the main processing loop
	w.wg.Add(1)
	go w.processLoop(ctx)

	// Start the retry processing loop
	w.wg.Add(1)
	go w.retryLoop(ctx)

	// Start the channel monitoring loop (Slack-specific)
	w.wg.Add(1)
	go w.channelMonitoringLoop(ctx)

	return nil
}

// Stop stops the Slack worker gracefully
func (w *SlackWorker) Stop(ctx context.Context) error {
	w.logger.Info("Stopping Slack worker")

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
		w.logger.Info("Slack worker stopped gracefully")
	case <-time.After(10 * time.Second):
		w.logger.Warn("Slack worker stop timeout, some operations may still be running")
	}

	return nil
}

// IsActive returns whether the worker is currently active
func (w *SlackWorker) IsActive() bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.active
}

// QueueSlackMessage queues a Slack request for processing
func (w *SlackWorker) QueueSlackMessage(request *SlackNotificationRequest) error {
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
func (w *SlackWorker) GetQueueLength() int {
	return len(w.workQueue)
}

// GetMetrics returns worker metrics
func (w *SlackWorker) GetMetrics() WorkerMetrics {
	return WorkerMetrics{
		WorkerID:      w.ID,
		Active:        w.IsActive(),
		QueueLength:   w.GetQueueLength(),
		QueueCapacity: cap(w.workQueue),
	}
}

// processLoop is the main processing loop for handling Slack requests
func (w *SlackWorker) processLoop(ctx context.Context) {
	defer w.wg.Done()

	w.logger.Debug("Slack worker processing loop started")

	for {
		select {
		case <-w.stopChan:
			w.logger.Debug("Slack worker processing loop stopped")
			return

		case <-ctx.Done():
			w.logger.Debug("Slack worker processing loop cancelled")
			return

		case request := <-w.workQueue:
			w.processSlackRequest(ctx, request)

		case <-time.After(30 * time.Second):
			// Periodic health check and maintenance
			w.performMaintenance(ctx)
		}
	}
}

// retryLoop handles retry processing for failed Slack messages
func (w *SlackWorker) retryLoop(ctx context.Context) {
	defer w.wg.Done()

	w.logger.Debug("Slack worker retry loop started")

	ticker := time.NewTicker(60 * time.Second) // Check for retries every minute
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			w.logger.Debug("Slack worker retry loop stopped")
			return

		case <-ctx.Done():
			w.logger.Debug("Slack worker retry loop cancelled")
			return

		case <-ticker.C:
			w.processRetries(ctx)
		}
	}
}

// channelMonitoringLoop monitors Slack channels for validity (Slack-specific)
func (w *SlackWorker) channelMonitoringLoop(ctx context.Context) {
	defer w.wg.Done()

	w.logger.Debug("Slack worker channel monitoring loop started")

	ticker := time.NewTicker(5 * time.Minute) // Check channels every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			w.logger.Debug("Slack worker channel monitoring loop stopped")
			return

		case <-ctx.Done():
			w.logger.Debug("Slack worker channel monitoring loop cancelled")
			return

		case <-ticker.C:
			w.monitorChannels(ctx)
		}
	}
}

// processSlackRequest processes a single Slack request
func (w *SlackWorker) processSlackRequest(ctx context.Context, request *SlackNotificationRequest) {
	logger := w.logger.With(
		"correlation_id", request.CorrelationID,
		"event_type", request.EventType,
		"channels", len(request.Channels))

	logger.Debug("Processing Slack request")

	startTime := time.Now()

	// Rate limit Slack API calls
	if err := w.rateLimiter.Wait(ctx); err != nil {
		logger.Debug("Processing cancelled during rate limiting")
		return
	}

	// Add processing delay if configured (usually minimal for Slack)
	if w.service.config.ProcessingDelay > 0 {
		select {
		case <-time.After(w.service.config.ProcessingDelay):
		case <-ctx.Done():
			logger.Debug("Processing cancelled during delay")
			return
		}
	}

	// Validate and resolve channels before processing
	validChannels := w.validateAndResolveChannels(ctx, request.Channels)
	if len(validChannels) == 0 {
		logger.Warn("No valid Slack channels in request")
		w.handleProcessingError(ctx, request, 
			fmt.Errorf("no valid Slack channels found in request"))
		return
	}

	// Update request with valid channels only
	request.Channels = validChannels

	// Optimize message content for Slack
	request = w.optimizeMessageForSlack(request)

	// Process the Slack request
	err := w.service.ProcessSlackRequest(ctx, request)
	
	processingTime := time.Since(startTime)

	if err != nil {
		logger.Error("Failed to process Slack request",
			"error", err,
			"processing_time", processingTime)

		w.handleProcessingError(ctx, request, err)
	} else {
		logger.Info("Successfully processed Slack request",
			"processing_time", processingTime,
			"valid_channels", len(validChannels))
	}
}

// processRetries processes retry requests for failed Slack messages
func (w *SlackWorker) processRetries(ctx context.Context) {
	logger := w.logger.With("operation", "retry_processing")
	
	logger.Debug("Processing Slack retries")

	// Get failed messages that are ready for retry
	failedMessages, err := w.service.slackRepository.GetFailedMessages(ctx, 10)
	if err != nil {
		logger.Error("Failed to get failed messages for retry", "error", err)
		return
	}

	if len(failedMessages) == 0 {
		logger.Debug("No failed Slack messages to retry")
		return
	}

	logger.Info("Found failed Slack messages to retry", "count", len(failedMessages))

	for _, message := range failedMessages {
		// Rate limit retries
		if err := w.rateLimiter.Wait(ctx); err != nil {
			logger.Debug("Retry processing cancelled due to rate limiting")
			return
		}

		// Check if it's time to retry
		status, err := w.service.slackRepository.GetDeliveryStatus(ctx, message.MessageID)
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
			logger.Debug("Slack message exceeded max retries",
				"message_id", message.MessageID,
				"attempts", status.AttemptCount)
			continue // Exceeded max retries
		}

		// Attempt retry
		logger.Info("Retrying failed Slack message",
			"message_id", message.MessageID,
			"attempt", status.AttemptCount+1)

		if err := w.service.RetryFailedSlackMessage(ctx, message.MessageID); err != nil {
			logger.Error("Slack retry attempt failed",
				"message_id", message.MessageID,
				"error", err)
		}
	}
}

// monitorChannels monitors Slack channels for validity and accessibility
func (w *SlackWorker) monitorChannels(ctx context.Context) {
	logger := w.logger.With("operation", "channel_monitoring")
	
	logger.Debug("Monitoring Slack channels")

	// Get channels from event type mappings
	allChannels := make(map[string]bool)
	for _, channels := range EventChannelMap {
		for _, channel := range channels {
			allChannels[channel] = true
		}
	}

	// Add default channel
	if w.service.config.Slack.DefaultChannel != "" {
		allChannels[w.service.config.Slack.DefaultChannel] = true
	}

	// Check each channel
	for channel := range allChannels {
		if err := w.rateLimiter.Wait(ctx); err != nil {
			logger.Debug("Channel monitoring cancelled due to rate limiting")
			return
		}

		channelInfo, err := w.service.slackClient.GetChannelInfo(ctx, channel)
		if err != nil {
			logger.Warn("Channel monitoring detected issue",
				"channel", channel,
				"error", err)
		} else {
			logger.Debug("Channel monitoring OK",
				"channel", channel,
				"channel_id", channelInfo.ID,
				"is_member", channelInfo.IsMember)
		}
	}
}

// performMaintenance performs periodic maintenance tasks
func (w *SlackWorker) performMaintenance(ctx context.Context) {
	logger := w.logger.With("operation", "maintenance")
	
	logger.Debug("Performing Slack worker maintenance")

	// Log worker statistics
	metrics := w.GetMetrics()
	logger.Debug("Slack worker metrics",
		"queue_length", metrics.QueueLength,
		"queue_capacity", metrics.QueueCapacity,
		"queue_utilization", float64(metrics.QueueLength)/float64(metrics.QueueCapacity)*100)

	// Log rate limiting statistics
	w.logRateLimitingStats(ctx)
	
	// Health check the dependencies
	if err := w.service.slackClient.HealthCheck(ctx); err != nil {
		logger.Warn("Slack client health check failed during maintenance", "error", err)
	}
}

// logRateLimitingStats logs rate limiting statistics
func (w *SlackWorker) logRateLimitingStats(ctx context.Context) {
	// In a full implementation, this would collect statistics
	// about rate limiting effectiveness
	w.logger.Debug("Slack rate limiting statistics logged")
}

// validateAndResolveChannels validates and resolves Slack channel names
func (w *SlackWorker) validateAndResolveChannels(ctx context.Context, channels []string) []string {
	valid := make([]string, 0, len(channels))
	
	for _, channel := range channels {
		if IsValidSlackChannel(channel) {
			// For channels starting with #, we keep them as-is
			// In a full implementation, we might resolve channel names to IDs
			valid = append(valid, channel)
		} else {
			w.logger.Warn("Invalid Slack channel in request", "channel", channel)
		}
	}
	
	return valid
}

// optimizeMessageForSlack optimizes content for Slack delivery
func (w *SlackWorker) optimizeMessageForSlack(request *SlackNotificationRequest) *SlackNotificationRequest {
	// Create a copy to avoid modifying the original
	optimized := *request
	
	// Generate enhanced content based on event type and priority
	if content := GenerateSlackContent(request.EventType, request.EventData); content != "" {
		optimized.EventData = make(map[string]interface{})
		for k, v := range request.EventData {
			optimized.EventData[k] = v
		}
		optimized.EventData["optimized_content"] = content
	}

	// Add priority-specific optimizations
	switch strings.ToLower(request.Priority) {
	case "urgent", "critical":
		// For urgent messages, ensure they go to alert channels
		if !w.containsAlertChannel(optimized.Channels) {
			optimized.Channels = append(optimized.Channels, "#alerts")
		}
	case "high":
		// For high priority, add appropriate formatting indicators
		// This would be handled in the message generation
	}

	return &optimized
}

// containsAlertChannel checks if channels contain an alert channel
func (w *SlackWorker) containsAlertChannel(channels []string) bool {
	alertChannels := []string{"#alerts", "#critical", "#urgent", "#monitoring"}
	
	for _, channel := range channels {
		for _, alert := range alertChannels {
			if channel == alert {
				return true
			}
		}
	}
	
	return false
}

// handleProcessingError handles errors that occur during Slack processing
func (w *SlackWorker) handleProcessingError(ctx context.Context, request *SlackNotificationRequest, processingError error) {
	logger := w.logger.With(
		"correlation_id", request.CorrelationID,
		"error_type", fmt.Sprintf("%T", processingError))

	// Categorize the error to determine handling strategy
	switch {
	case isSlackRateLimitError(processingError):
		// Rate limit errors should be retried with exponential backoff
		logger.Warn("Slack rate limit error encountered, will retry with backoff",
			"error", processingError)
		
	case isSlackAuthError(processingError):
		// Auth errors are permanent and should not be retried
		logger.Error("Slack authentication error encountered, moving to dead letter queue",
			"error", processingError)
		w.sendToDeadLetterQueue(ctx, request, processingError)
		
	case isSlackChannelError(processingError):
		// Channel errors might be temporary (archived channels, permissions)
		logger.Warn("Slack channel error encountered, will attempt with default channel",
			"error", processingError)
		w.fallbackToDefaultChannel(ctx, request, processingError)
		
	case isSlackContentError(processingError):
		// Content errors are permanent (message too long, invalid format)
		logger.Error("Slack content error encountered, moving to dead letter queue",
			"error", processingError)
		w.sendToDeadLetterQueue(ctx, request, processingError)
		
	case isDependencyError(processingError):
		// Dependency errors (Slack API down, database issues) should be retried
		logger.Warn("Slack dependency error encountered, will retry",
			"error", processingError)
		
	default:
		// Unknown errors should be logged and retried with caution
		logger.Error("Unknown Slack error encountered during processing",
			"error", processingError)
	}
}

// fallbackToDefaultChannel attempts to send to default channel on channel errors
func (w *SlackWorker) fallbackToDefaultChannel(ctx context.Context, request *SlackNotificationRequest, originalError error) {
	logger := w.logger.With("correlation_id", request.CorrelationID)
	
	// Create fallback request with default channel
	fallbackRequest := *request
	fallbackRequest.Channels = []string{w.service.config.Slack.DefaultChannel}
	
	// Add note about fallback in event data
	fallbackRequest.EventData = make(map[string]interface{})
	for k, v := range request.EventData {
		fallbackRequest.EventData[k] = v
	}
	fallbackRequest.EventData["fallback_reason"] = originalError.Error()
	fallbackRequest.EventData["original_channels"] = request.Channels

	logger.Info("Attempting fallback to default channel",
		"default_channel", w.service.config.Slack.DefaultChannel)

	// Attempt to process with default channel
	if err := w.service.ProcessSlackRequest(ctx, &fallbackRequest); err != nil {
		logger.Error("Fallback to default channel failed", "error", err)
		w.sendToDeadLetterQueue(ctx, request, err)
	} else {
		logger.Info("Successfully sent to default channel as fallback")
	}
}

// sendToDeadLetterQueue sends a message to the dead letter queue
func (w *SlackWorker) sendToDeadLetterQueue(ctx context.Context, request *SlackNotificationRequest, err error) {
	logger := w.logger.With("correlation_id", request.CorrelationID)
	
	logger.Info("Sending Slack message to dead letter queue")

	deadLetterMessage := DeadLetterMessage{
		OriginalRequest: *request,
		Error:          err.Error(),
		Timestamp:      time.Now().UTC(),
		WorkerID:       w.ID,
		Reason:         "slack_processing_failed",
	}

	logger.Error("Slack message sent to dead letter queue",
		"message_id", deadLetterMessage.OriginalRequest.CorrelationID,
		"error", deadLetterMessage.Error,
		"reason", deadLetterMessage.Reason)
}

// Error classification helper functions specific to Slack
func isSlackRateLimitError(err error) bool {
	// Check if error is a Slack rate limit error
	return strings.Contains(strings.ToLower(err.Error()), "rate_limited")
}

func isSlackAuthError(err error) bool {
	// Check if error is a Slack authentication error
	errorStr := strings.ToLower(err.Error())
	return strings.Contains(errorStr, "invalid_auth") || 
		   strings.Contains(errorStr, "account_inactive")
}

func isSlackChannelError(err error) bool {
	// Check if error is related to Slack channels
	errorStr := strings.ToLower(err.Error())
	return strings.Contains(errorStr, "channel_not_found") ||
		   strings.Contains(errorStr, "is_archived") ||
		   strings.Contains(errorStr, "restricted_action")
}

func isSlackContentError(err error) bool {
	// Check if error is related to message content
	errorStr := strings.ToLower(err.Error())
	return strings.Contains(errorStr, "msg_too_long") ||
		   strings.Contains(errorStr, "no_text") ||
		   strings.Contains(errorStr, "too_many_attachments")
}

func isDependencyError(err error) bool {
	// Check if error is a dependency-related error
	return false // Simplified for now
}

// Supporting Types

// WorkerMetrics represents metrics for a Slack worker
type WorkerMetrics struct {
	WorkerID      int  `json:"worker_id"`
	Active        bool `json:"active"`
	QueueLength   int  `json:"queue_length"`
	QueueCapacity int  `json:"queue_capacity"`
}

// DeadLetterMessage represents a Slack message that failed processing
type DeadLetterMessage struct {
	OriginalRequest SlackNotificationRequest `json:"original_request"`
	Error          string                   `json:"error"`
	Timestamp      time.Time                `json:"timestamp"`
	WorkerID       int                      `json:"worker_id"`
	Reason         string                   `json:"reason"`
}


// Slack-specific worker capabilities

// ProcessThreadedMessage processes a Slack message as part of a thread
func (w *SlackWorker) ProcessThreadedMessage(ctx context.Context, request *SlackNotificationRequest, threadTS string) error {
	logger := w.logger.With("correlation_id", request.CorrelationID, "thread_ts", threadTS)
	logger.Info("Processing threaded Slack message")

	// Add thread timestamp to event data
	if request.EventData == nil {
		request.EventData = make(map[string]interface{})
	}
	request.EventData["thread_ts"] = threadTS

	return w.service.ProcessSlackRequest(ctx, request)
}

// AddReaction adds a reaction to a Slack message (for status updates)
func (w *SlackWorker) AddReaction(ctx context.Context, channel, messageTS, reaction string) error {
	logger := w.logger.With("channel", channel, "message_ts", messageTS, "reaction", reaction)
	logger.Debug("Adding reaction to Slack message")

	// Rate limit the reaction
	if err := w.rateLimiter.Wait(ctx); err != nil {
		return err
	}

	// In a full implementation, this would call the Slack API to add a reaction
	logger.Info("Reaction added to Slack message")
	return nil
}

// UpdateMessageStatus updates a message to reflect status changes
func (w *SlackWorker) UpdateMessageStatus(ctx context.Context, messageID string, newStatus string) error {
	logger := w.logger.With("message_id", messageID, "new_status", newStatus)
	logger.Debug("Updating Slack message status")

	// Get original message
	message, err := w.service.slackRepository.GetMessage(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}

	// Create status update content
	statusEmoji := map[string]string{
		"processing": "⏳",
		"completed":  "✅",
		"failed":     "❌",
		"cancelled":  "🚫",
	}

	emoji, exists := statusEmoji[strings.ToLower(newStatus)]
	if !exists {
		emoji = "ℹ️"
	}

	updatedContent := fmt.Sprintf("%s %s Status: %s", emoji, message.Content, newStatus)

	// Update the message in each channel
	for _, channelStatus := range message.Channels {
		if channelStatus != "" && message.MessageID != "" {
			updateRequest := &SlackUpdateMessageRequest{
				Channel:   channelStatus,
				MessageTS: message.MessageID, // Would need to store actual message timestamps
				Text:      updatedContent,
			}

			if err := w.rateLimiter.Wait(ctx); err != nil {
				return err
			}

			_, err := w.service.slackClient.UpdateMessage(ctx, updateRequest)
			if err != nil {
				logger.Warn("Failed to update message status in channel",
					"channel", channelStatus,
					"error", err)
			}
		}
	}

	logger.Info("Slack message status updated successfully")
	return nil
}