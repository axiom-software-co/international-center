package automation

import (
	"context"
	"fmt"
	"time"
)

// NewNotificationManager creates a new notification manager
func NewNotificationManager() *NotificationManager {
	nm := &NotificationManager{
		channels: make(map[string]NotificationChannel),
	}
	
	// Register default notification channels
	nm.RegisterChannel("console", &ConsoleNotificationChannel{})
	nm.RegisterChannel("email", &EmailNotificationChannel{})
	nm.RegisterChannel("slack", &SlackNotificationChannel{})
	nm.RegisterChannel("webhook", &WebhookNotificationChannel{})
	
	return nm
}

// RegisterChannel registers a notification channel
func (nm *NotificationManager) RegisterChannel(name string, channel NotificationChannel) {
	nm.channels[name] = channel
}

// SendNotification sends a notification to all configured channels
func (nm *NotificationManager) SendNotification(ctx context.Context, message *NotificationMessage) error {
	var errors []string
	
	for name, channel := range nm.channels {
		if err := channel.Send(ctx, message); err != nil {
			errors = append(errors, fmt.Sprintf("channel %s failed: %v", name, err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("notification failures: %v", errors)
	}
	
	return nil
}

// SendDeploymentStarted sends deployment started notification
func (nm *NotificationManager) SendDeploymentStarted(ctx context.Context, environment, deploymentID string) error {
	message := &NotificationMessage{
		Title:       fmt.Sprintf("Deployment Started - %s", environment),
		Body:        fmt.Sprintf("Deployment %s has started for environment %s", deploymentID, environment),
		Environment: environment,
		Priority:    NotificationPriorityNormal,
		Timestamp:   time.Now(),
		Data: map[string]interface{}{
			"deploymentId": deploymentID,
			"status":       "started",
		},
	}
	
	return nm.SendNotification(ctx, message)
}

// SendDeploymentSucceeded sends deployment success notification
func (nm *NotificationManager) SendDeploymentSucceeded(ctx context.Context, environment, deploymentID string, result *DeploymentResult) error {
	message := &NotificationMessage{
		Title:       fmt.Sprintf("Deployment Succeeded - %s", environment),
		Body:        fmt.Sprintf("Deployment %s completed successfully for environment %s in %v", deploymentID, environment, result.Duration),
		Environment: environment,
		Priority:    NotificationPriorityNormal,
		Timestamp:   time.Now(),
		Data: map[string]interface{}{
			"deploymentId": deploymentID,
			"status":       "succeeded",
			"duration":     result.Duration.String(),
			"resources":    len(result.Resources),
		},
	}
	
	return nm.SendNotification(ctx, message)
}

// SendDeploymentFailed sends deployment failure notification
func (nm *NotificationManager) SendDeploymentFailed(ctx context.Context, environment, deploymentID string, err error) error {
	message := &NotificationMessage{
		Title:       fmt.Sprintf("Deployment Failed - %s", environment),
		Body:        fmt.Sprintf("Deployment %s failed for environment %s: %v", deploymentID, environment, err),
		Environment: environment,
		Priority:    NotificationPriorityCritical,
		Timestamp:   time.Now(),
		Data: map[string]interface{}{
			"deploymentId": deploymentID,
			"status":       "failed",
			"error":        err.Error(),
		},
	}
	
	return nm.SendNotification(ctx, message)
}

// SendValidationFailed sends validation failure notification
func (nm *NotificationManager) SendValidationFailed(ctx context.Context, environment, deploymentID string, err error) error {
	message := &NotificationMessage{
		Title:       fmt.Sprintf("Validation Failed - %s", environment),
		Body:        fmt.Sprintf("Validation failed for deployment %s in environment %s: %v", deploymentID, environment, err),
		Environment: environment,
		Priority:    NotificationPriorityHigh,
		Timestamp:   time.Now(),
		Data: map[string]interface{}{
			"deploymentId": deploymentID,
			"type":         "validation_failed",
			"error":        err.Error(),
		},
	}
	
	return nm.SendNotification(ctx, message)
}

// SendRollbackStarted sends rollback started notification
func (nm *NotificationManager) SendRollbackStarted(ctx context.Context, environment, deploymentID string, originalError error) error {
	message := &NotificationMessage{
		Title:       fmt.Sprintf("Rollback Started - %s", environment),
		Body:        fmt.Sprintf("Rollback initiated for deployment %s in environment %s due to: %v", deploymentID, environment, originalError),
		Environment: environment,
		Priority:    NotificationPriorityHigh,
		Timestamp:   time.Now(),
		Data: map[string]interface{}{
			"deploymentId":   deploymentID,
			"type":           "rollback_started",
			"originalError":  originalError.Error(),
		},
	}
	
	return nm.SendNotification(ctx, message)
}

// SendApprovalRequired sends approval required notification
func (nm *NotificationManager) SendApprovalRequired(ctx context.Context, environment, deploymentID, approvalID string, approvers []string) error {
	message := &NotificationMessage{
		Title:       fmt.Sprintf("Approval Required - %s", environment),
		Body:        fmt.Sprintf("Deployment %s to %s requires approval (ID: %s). Required approvers: %v", deploymentID, environment, approvalID, approvers),
		Environment: environment,
		Priority:    NotificationPriorityHigh,
		Timestamp:   time.Now(),
		Data: map[string]interface{}{
			"deploymentId": deploymentID,
			"approvalId":   approvalID,
			"type":         "approval_required",
			"approvers":    approvers,
		},
	}
	
	return nm.SendNotification(ctx, message)
}

// ConsoleNotificationChannel implementation
type ConsoleNotificationChannel struct{}

func (cnc *ConsoleNotificationChannel) Send(ctx context.Context, message *NotificationMessage) error {
	fmt.Printf("\n=== NOTIFICATION ===\n")
	fmt.Printf("Priority: %s\n", message.Priority)
	fmt.Printf("Environment: %s\n", message.Environment)
	fmt.Printf("Title: %s\n", message.Title)
	fmt.Printf("Body: %s\n", message.Body)
	fmt.Printf("Timestamp: %s\n", message.Timestamp.Format(time.RFC3339))
	
	if len(message.Data) > 0 {
		fmt.Printf("Additional Data:\n")
		for key, value := range message.Data {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}
	fmt.Printf("===================\n\n")
	
	return nil
}

// EmailNotificationChannel implementation
type EmailNotificationChannel struct {
	smtpServer   string
	smtpPort     int
	username     string
	password     string
	fromAddress  string
	toAddresses  []string
}

func NewEmailNotificationChannel(smtpServer string, smtpPort int, username, password, fromAddress string, toAddresses []string) *EmailNotificationChannel {
	return &EmailNotificationChannel{
		smtpServer:  smtpServer,
		smtpPort:    smtpPort,
		username:    username,
		password:    password,
		fromAddress: fromAddress,
		toAddresses: toAddresses,
	}
}

func (enc *EmailNotificationChannel) Send(ctx context.Context, message *NotificationMessage) error {
	// TODO: Implement email sending using SMTP
	fmt.Printf("Email notification: %s - %s\n", message.Title, message.Body)
	return nil
}

// SlackNotificationChannel implementation
type SlackNotificationChannel struct {
	webhookURL string
	channel    string
	username   string
}

func NewSlackNotificationChannel(webhookURL, channel, username string) *SlackNotificationChannel {
	return &SlackNotificationChannel{
		webhookURL: webhookURL,
		channel:    channel,
		username:   username,
	}
}

func (snc *SlackNotificationChannel) Send(ctx context.Context, message *NotificationMessage) error {
	// TODO: Implement Slack webhook integration
	fmt.Printf("Slack notification: %s - %s\n", message.Title, message.Body)
	return nil
}

// WebhookNotificationChannel implementation
type WebhookNotificationChannel struct {
	webhookURL string
	headers    map[string]string
}

func NewWebhookNotificationChannel(webhookURL string, headers map[string]string) *WebhookNotificationChannel {
	return &WebhookNotificationChannel{
		webhookURL: webhookURL,
		headers:    headers,
	}
}

func (wnc *WebhookNotificationChannel) Send(ctx context.Context, message *NotificationMessage) error {
	// TODO: Implement HTTP webhook integration
	fmt.Printf("Webhook notification: %s - %s\n", message.Title, message.Body)
	return nil
}

// NotificationRule defines rules for when to send notifications
type NotificationRule struct {
	Environment  string
	Events       []string
	Priorities   []NotificationPriority
	Channels     []string
	Conditions   []NotificationCondition
	Throttling   *NotificationThrottling
}

// NotificationCondition defines conditions for sending notifications
type NotificationCondition struct {
	Field    string
	Operator string
	Value    interface{}
}

// NotificationThrottling defines throttling rules for notifications
type NotificationThrottling struct {
	MaxNotifications int
	TimeWindow       time.Duration
	CooldownPeriod   time.Duration
}

// NotificationRuleEngine processes notification rules
type NotificationRuleEngine struct {
	rules []NotificationRule
}

// NewNotificationRuleEngine creates a new notification rule engine
func NewNotificationRuleEngine() *NotificationRuleEngine {
	nre := &NotificationRuleEngine{}
	nre.setDefaultRules()
	return nre
}

// setDefaultRules sets default notification rules
func (nre *NotificationRuleEngine) setDefaultRules() {
	nre.rules = []NotificationRule{
		{
			Environment: "production",
			Events:      []string{"deployment_started", "deployment_succeeded", "deployment_failed"},
			Priorities:  []NotificationPriority{NotificationPriorityNormal, NotificationPriorityHigh, NotificationPriorityCritical},
			Channels:    []string{"email", "slack"},
		},
		{
			Environment: "staging",
			Events:      []string{"deployment_failed", "validation_failed"},
			Priorities:  []NotificationPriority{NotificationPriorityHigh, NotificationPriorityCritical},
			Channels:    []string{"slack"},
		},
		{
			Environment: "*", // All environments
			Events:      []string{"approval_required"},
			Priorities:  []NotificationPriority{NotificationPriorityHigh},
			Channels:    []string{"email", "slack"},
		},
	}
}

// ShouldSendNotification determines if a notification should be sent based on rules
func (nre *NotificationRuleEngine) ShouldSendNotification(message *NotificationMessage, eventType string) (bool, []string) {
	var applicableChannels []string
	
	for _, rule := range nre.rules {
		// Check environment match
		if rule.Environment != "*" && rule.Environment != message.Environment {
			continue
		}
		
		// Check event type match
		eventMatches := false
		for _, event := range rule.Events {
			if event == eventType {
				eventMatches = true
				break
			}
		}
		if !eventMatches {
			continue
		}
		
		// Check priority match
		priorityMatches := false
		for _, priority := range rule.Priorities {
			if priority == message.Priority {
				priorityMatches = true
				break
			}
		}
		if !priorityMatches {
			continue
		}
		
		// Rule matches, add channels
		applicableChannels = append(applicableChannels, rule.Channels...)
	}
	
	return len(applicableChannels) > 0, applicableChannels
}

// NotificationHistory tracks notification history
type NotificationHistory struct {
	notifications []NotificationHistoryEntry
}

// NotificationHistoryEntry represents a notification history entry
type NotificationHistoryEntry struct {
	ID          string
	Message     *NotificationMessage
	Channels    []string
	SentAt      time.Time
	Status      string
	Error       string
}

// NewNotificationHistory creates a new notification history tracker
func NewNotificationHistory() *NotificationHistory {
	return &NotificationHistory{
		notifications: []NotificationHistoryEntry{},
	}
}

// AddEntry adds a notification history entry
func (nh *NotificationHistory) AddEntry(id string, message *NotificationMessage, channels []string, status, errorMsg string) {
	entry := NotificationHistoryEntry{
		ID:       id,
		Message:  message,
		Channels: channels,
		SentAt:   time.Now(),
		Status:   status,
		Error:    errorMsg,
	}
	
	nh.notifications = append(nh.notifications, entry)
}

// GetHistory returns notification history for an environment
func (nh *NotificationHistory) GetHistory(environment string) []NotificationHistoryEntry {
	var history []NotificationHistoryEntry
	for _, entry := range nh.notifications {
		if entry.Message.Environment == environment {
			history = append(history, entry)
		}
	}
	return history
}

// GetRecentNotifications returns recent notifications
func (nh *NotificationHistory) GetRecentNotifications(since time.Time) []NotificationHistoryEntry {
	var recent []NotificationHistoryEntry
	for _, entry := range nh.notifications {
		if entry.SentAt.After(since) {
			recent = append(recent, entry)
		}
	}
	return recent
}