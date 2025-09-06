package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
)

// SlackAPIClient provides Slack API functionality
type SlackAPIClient interface {
	Initialize(ctx context.Context, config *SlackConfig) error
	SendMessage(ctx context.Context, request *SlackSendMessageRequest) (*SlackSendMessageResponse, error)
	UpdateMessage(ctx context.Context, request *SlackUpdateMessageRequest) (*SlackUpdateMessageResponse, error)
	DeleteMessage(ctx context.Context, channel, messageTS string) error
	GetChannelInfo(ctx context.Context, channel string) (*SlackChannelInfo, error)
	HealthCheck(ctx context.Context) error
}

// SlackWebAPIClient implements SlackAPIClient using Slack Web API
type SlackWebAPIClient struct {
	config     *SlackConfig
	httpClient *http.Client
	logger     *slog.Logger
	baseURL    string
}

// NewSlackWebAPIClient creates a new Slack Web API client
func NewSlackWebAPIClient(logger *slog.Logger) *SlackWebAPIClient {
	return &SlackWebAPIClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:  logger,
		baseURL: "https://slack.com/api",
	}
}

// Initialize initializes the Slack client with configuration
func (s *SlackWebAPIClient) Initialize(ctx context.Context, config *SlackConfig) error {
	if config == nil {
		return domain.NewValidationError("Slack configuration cannot be nil")
	}

	if config.BotToken == "" {
		return domain.NewValidationError("Slack bot token is required")
	}

	if config.DefaultChannel == "" {
		return domain.NewValidationError("default channel is required")
	}

	// Validate default channel format
	if !IsValidSlackChannel(config.DefaultChannel) {
		return domain.NewValidationError("default channel format is invalid")
	}

	s.config = config
	s.httpClient.Timeout = time.Duration(config.RequestTimeout) * time.Second

	s.logger.Info("Slack Web API client initialized",
		"default_channel", config.DefaultChannel,
		"timeout", s.httpClient.Timeout)

	// Test the bot token with a simple API call
	if err := s.testConnection(ctx); err != nil {
		return fmt.Errorf("failed to test Slack connection: %w", err)
	}

	return nil
}

// SendMessage sends a message to a Slack channel
func (s *SlackWebAPIClient) SendMessage(ctx context.Context, request *SlackSendMessageRequest) (*SlackSendMessageResponse, error) {
	if err := s.validateSendRequest(request); err != nil {
		return nil, fmt.Errorf("invalid send request: %w", err)
	}

	logger := s.logger.With(
		"channel", request.Channel,
		"text_length", len(request.Text))

	logger.Debug("Sending message to Slack")

	// Prepare API request
	url := fmt.Sprintf("%s/chat.postMessage", s.baseURL)
	
	// Create request payload
	payload := map[string]interface{}{
		"channel":  request.Channel,
		"text":     request.Text,
		"username": request.Username,
	}

	// Add optional fields
	if request.IconEmoji != "" {
		payload["icon_emoji"] = request.IconEmoji
	}
	if request.IconURL != "" {
		payload["icon_url"] = request.IconURL
	}
	if request.ThreadTS != "" {
		payload["thread_ts"] = request.ThreadTS
	}

	// Add attachments if present
	if len(request.Attachments) > 0 {
		payload["attachments"] = request.Attachments
	}

	// Add blocks if present (modern Slack message format)
	if len(request.Blocks) > 0 {
		payload["blocks"] = request.Blocks
	}

	// Make HTTP request
	response, err := s.makeAPIRequest(ctx, "POST", url, payload)
	if err != nil {
		logger.Error("Failed to send Slack message", "error", err)
		return nil, fmt.Errorf("Slack API request failed: %w", err)
	}

	// Parse response
	var slackResponse SlackSendMessageResponse
	if err := json.Unmarshal(response, &slackResponse); err != nil {
		return nil, fmt.Errorf("failed to parse Slack response: %w", err)
	}

	if !slackResponse.OK {
		logger.Error("Slack API returned error", "error", slackResponse.Error)
		return nil, fmt.Errorf("Slack API error: %s", slackResponse.Error)
	}

	logger.Info("Message sent to Slack successfully",
		"channel", slackResponse.Channel,
		"message_ts", slackResponse.MessageTS)

	return &slackResponse, nil
}

// UpdateMessage updates an existing message in Slack
func (s *SlackWebAPIClient) UpdateMessage(ctx context.Context, request *SlackUpdateMessageRequest) (*SlackUpdateMessageResponse, error) {
	if err := s.validateUpdateRequest(request); err != nil {
		return nil, fmt.Errorf("invalid update request: %w", err)
	}

	logger := s.logger.With(
		"channel", request.Channel,
		"message_ts", request.MessageTS)

	logger.Debug("Updating Slack message")

	// Prepare API request
	url := fmt.Sprintf("%s/chat.update", s.baseURL)
	
	payload := map[string]interface{}{
		"channel": request.Channel,
		"ts":      request.MessageTS,
		"text":    request.Text,
	}

	// Add attachments and blocks if present
	if len(request.Attachments) > 0 {
		payload["attachments"] = request.Attachments
	}
	if len(request.Blocks) > 0 {
		payload["blocks"] = request.Blocks
	}

	// Make HTTP request
	response, err := s.makeAPIRequest(ctx, "POST", url, payload)
	if err != nil {
		logger.Error("Failed to update Slack message", "error", err)
		return nil, fmt.Errorf("Slack API request failed: %w", err)
	}

	// Parse response
	var slackResponse SlackUpdateMessageResponse
	if err := json.Unmarshal(response, &slackResponse); err != nil {
		return nil, fmt.Errorf("failed to parse Slack response: %w", err)
	}

	if !slackResponse.OK {
		logger.Error("Slack API returned error", "error", slackResponse.Error)
		return nil, fmt.Errorf("Slack API error: %s", slackResponse.Error)
	}

	logger.Info("Message updated in Slack successfully")
	return &slackResponse, nil
}

// DeleteMessage deletes a message from Slack
func (s *SlackWebAPIClient) DeleteMessage(ctx context.Context, channel, messageTS string) error {
	if channel == "" {
		return domain.NewValidationError("channel is required")
	}
	if messageTS == "" {
		return domain.NewValidationError("message timestamp is required")
	}

	logger := s.logger.With("channel", channel, "message_ts", messageTS)
	logger.Debug("Deleting Slack message")

	// Prepare API request
	url := fmt.Sprintf("%s/chat.delete", s.baseURL)
	
	payload := map[string]interface{}{
		"channel": channel,
		"ts":      messageTS,
	}

	// Make HTTP request
	response, err := s.makeAPIRequest(ctx, "POST", url, payload)
	if err != nil {
		logger.Error("Failed to delete Slack message", "error", err)
		return fmt.Errorf("Slack API request failed: %w", err)
	}

	// Parse response
	var slackResponse struct {
		OK    bool   `json:"ok"`
		Error string `json:"error,omitempty"`
	}
	if err := json.Unmarshal(response, &slackResponse); err != nil {
		return fmt.Errorf("failed to parse Slack response: %w", err)
	}

	if !slackResponse.OK {
		logger.Error("Slack API returned error", "error", slackResponse.Error)
		return fmt.Errorf("Slack API error: %s", slackResponse.Error)
	}

	logger.Info("Message deleted from Slack successfully")
	return nil
}

// GetChannelInfo retrieves information about a Slack channel
func (s *SlackWebAPIClient) GetChannelInfo(ctx context.Context, channel string) (*SlackChannelInfo, error) {
	if channel == "" {
		return nil, domain.NewValidationError("channel is required")
	}

	logger := s.logger.With("channel", channel)
	logger.Debug("Getting Slack channel info")

	// Determine the appropriate API method based on channel format
	var url string
	var payload map[string]interface{}

	if strings.HasPrefix(channel, "#") || strings.HasPrefix(channel, "C") {
		// Public channel
		url = fmt.Sprintf("%s/conversations.info", s.baseURL)
		payload = map[string]interface{}{
			"channel": strings.TrimPrefix(channel, "#"),
		}
	} else if strings.HasPrefix(channel, "@") || strings.HasPrefix(channel, "D") {
		// Direct message
		url = fmt.Sprintf("%s/conversations.info", s.baseURL)
		payload = map[string]interface{}{
			"channel": strings.TrimPrefix(channel, "@"),
		}
	} else {
		return nil, domain.NewValidationError("unsupported channel format")
	}

	// Make HTTP request
	response, err := s.makeAPIRequest(ctx, "GET", url, payload)
	if err != nil {
		logger.Error("Failed to get Slack channel info", "error", err)
		return nil, fmt.Errorf("Slack API request failed: %w", err)
	}

	// Parse response
	var slackResponse struct {
		OK      bool `json:"ok"`
		Channel struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			IsIM     bool   `json:"is_im"`
			IsGroup  bool   `json:"is_group"`
			IsMember bool   `json:"is_member"`
		} `json:"channel"`
		Error string `json:"error,omitempty"`
	}
	
	if err := json.Unmarshal(response, &slackResponse); err != nil {
		return nil, fmt.Errorf("failed to parse Slack response: %w", err)
	}

	if !slackResponse.OK {
		logger.Error("Slack API returned error", "error", slackResponse.Error)
		return nil, fmt.Errorf("Slack API error: %s", slackResponse.Error)
	}

	channelInfo := &SlackChannelInfo{
		ID:       slackResponse.Channel.ID,
		Name:     slackResponse.Channel.Name,
		IsIM:     slackResponse.Channel.IsIM,
		IsGroup:  slackResponse.Channel.IsGroup,
		IsMember: slackResponse.Channel.IsMember,
	}

	logger.Debug("Retrieved channel info", "channel_id", channelInfo.ID)
	return channelInfo, nil
}

// HealthCheck performs a health check on the Slack API
func (s *SlackWebAPIClient) HealthCheck(ctx context.Context) error {
	if s.config == nil {
		return domain.NewDependencyError("Slack client not initialized", nil)
	}

	return s.testConnection(ctx)
}

// Private helper methods

// testConnection tests the connection to Slack API
func (s *SlackWebAPIClient) testConnection(ctx context.Context) error {
	url := fmt.Sprintf("%s/auth.test", s.baseURL)
	
	response, err := s.makeAPIRequest(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("Slack connection test failed: %w", err)
	}

	var testResponse struct {
		OK    bool   `json:"ok"`
		Error string `json:"error,omitempty"`
	}
	
	if err := json.Unmarshal(response, &testResponse); err != nil {
		return fmt.Errorf("failed to parse test response: %w", err)
	}

	if !testResponse.OK {
		return fmt.Errorf("Slack auth test failed: %s", testResponse.Error)
	}

	s.logger.Debug("Slack connection test successful")
	return nil
}

// makeAPIRequest makes an HTTP request to the Slack API
func (s *SlackWebAPIClient) makeAPIRequest(ctx context.Context, method, url string, payload map[string]interface{}) ([]byte, error) {
	var body io.Reader

	if payload != nil {
		if method == "GET" {
			// For GET requests, add parameters to URL
			// This is simplified - in production would properly encode query parameters
			url += "?"
			for key, value := range payload {
				url += fmt.Sprintf("%s=%v&", key, value)
			}
			url = strings.TrimSuffix(url, "&")
		} else {
			// For POST requests, send JSON body
			jsonData, err := json.Marshal(payload)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal payload: %w", err)
			}
			body = bytes.NewBuffer(jsonData)
		}
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	req.Header.Set("Authorization", "Bearer "+s.config.BotToken)
	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Make request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	return responseBody, nil
}

// validateSendRequest validates a send message request
func (s *SlackWebAPIClient) validateSendRequest(request *SlackSendMessageRequest) error {
	if request == nil {
		return domain.NewValidationError("send request cannot be nil")
	}

	if request.Channel == "" {
		return domain.NewValidationError("channel is required")
	}

	if request.Text == "" && len(request.Attachments) == 0 && len(request.Blocks) == 0 {
		return domain.NewValidationError("text, attachments, or blocks are required")
	}

	// Validate channel format
	if !IsValidSlackChannel(request.Channel) {
		return domain.NewValidationError(fmt.Sprintf("invalid channel format: %s", request.Channel))
	}

	// Validate text length
	if len(request.Text) > MaxSlackMessageLength {
		return domain.NewValidationError(fmt.Sprintf("text too long: %d characters (max %d)", 
			len(request.Text), MaxSlackMessageLength))
	}

	return nil
}

// validateUpdateRequest validates an update message request
func (s *SlackWebAPIClient) validateUpdateRequest(request *SlackUpdateMessageRequest) error {
	if request == nil {
		return domain.NewValidationError("update request cannot be nil")
	}

	if request.Channel == "" {
		return domain.NewValidationError("channel is required")
	}

	if request.MessageTS == "" {
		return domain.NewValidationError("message timestamp is required")
	}

	if request.Text == "" && len(request.Attachments) == 0 && len(request.Blocks) == 0 {
		return domain.NewValidationError("text, attachments, or blocks are required")
	}

	// Validate channel format
	if !IsValidSlackChannel(request.Channel) {
		return domain.NewValidationError(fmt.Sprintf("invalid channel format: %s", request.Channel))
	}

	return nil
}

// Rate Limiting and Error Handling

// SlackRateLimiter handles Slack API rate limiting
type SlackRateLimiter struct {
	lastRequest time.Time
	minInterval time.Duration
	logger      *slog.Logger
}

// NewSlackRateLimiter creates a new rate limiter
func NewSlackRateLimiter(logger *slog.Logger) *SlackRateLimiter {
	return &SlackRateLimiter{
		minInterval: 1 * time.Second, // Slack allows ~1 message per second
		logger:      logger,
	}
}

// Wait waits for the appropriate interval before allowing the next request
func (rl *SlackRateLimiter) Wait(ctx context.Context) error {
	elapsed := time.Since(rl.lastRequest)
	if elapsed < rl.minInterval {
		waitTime := rl.minInterval - elapsed
		rl.logger.Debug("Rate limiting Slack API request", "wait_time", waitTime)
		
		select {
		case <-time.After(waitTime):
			rl.lastRequest = time.Now()
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	
	rl.lastRequest = time.Now()
	return nil
}

// SlackError represents a Slack API error
type SlackError struct {
	Code    string `json:"error"`
	Message string `json:"message,omitempty"`
}

// Error implements the error interface
func (e SlackError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("Slack API error %s: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("Slack API error: %s", e.Code)
}

// IsRetryable determines if a Slack error is retryable
func (e SlackError) IsRetryable() bool {
	retryableErrors := map[string]bool{
		"rate_limited":     true,
		"fatal_error":      false,
		"invalid_auth":     false,
		"account_inactive": false,
		"channel_not_found": false,
		"is_archived":      false,
		"msg_too_long":     false,
		"no_text":          false,
		"restricted_action": false,
		"thread_not_found": false,
		"too_many_attachments": false,
		"user_not_found":   false,
	}

	return retryableErrors[e.Code]
}