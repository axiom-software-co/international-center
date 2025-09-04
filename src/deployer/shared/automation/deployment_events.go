package automation

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// DeploymentEventHandler handles deployment events for monitoring and logging
type DeploymentEventHandler struct {
	logger         *log.Logger
	progressWriter io.Writer
	eventHandlers  []EventHandler
	metrics        *DeploymentMetrics
}

// EventHandler defines interface for handling deployment events
type EventHandler interface {
	OnEvent(event *DeploymentEvent)
}

// DeploymentEvent represents a deployment event
type DeploymentEvent struct {
	Type        EventType
	Environment string
	StackName   string
	Message     string
	Timestamp   time.Time
	Data        map[string]interface{}
}

// EventType defines the type of deployment event
type EventType string

const (
	EventStackCreated        EventType = "stack_created"
	EventDeploymentStarted   EventType = "deployment_started"
	EventDeploymentProgress  EventType = "deployment_progress"
	EventDeploymentCompleted EventType = "deployment_completed"
	EventDestroyStarted      EventType = "destroy_started"
	EventDestroyCompleted    EventType = "destroy_completed"
	EventValidationStarted   EventType = "validation_started"
	EventValidationCompleted EventType = "validation_completed"
	EventError               EventType = "error"
)

// DeploymentMetrics tracks deployment metrics
type DeploymentMetrics struct {
	TotalDeployments      int
	SuccessfulDeployments int
	FailedDeployments     int
	AverageDeployTime     time.Duration
	DeploymentsByEnvironment map[string]int
	LastDeploymentTime    time.Time
}

// ConsoleEventHandler logs events to console
type ConsoleEventHandler struct {
	logger *log.Logger
}

// WebhookEventHandler sends events to webhook endpoints
type WebhookEventHandler struct {
	webhookURL string
	client     interface{} // HTTP client would be defined here
}

// NewDeploymentEventHandler creates a new deployment event handler
func NewDeploymentEventHandler() *DeploymentEventHandler {
	return &DeploymentEventHandler{
		logger:         log.New(os.Stdout, "[DEPLOYMENT] ", log.LstdFlags),
		progressWriter: os.Stdout,
		eventHandlers:  []EventHandler{},
		metrics: &DeploymentMetrics{
			DeploymentsByEnvironment: make(map[string]int),
		},
	}
}

// AddEventHandler adds an event handler
func (deh *DeploymentEventHandler) AddEventHandler(handler EventHandler) {
	deh.eventHandlers = append(deh.eventHandlers, handler)
}

// GetProgressStreams returns progress streams for Pulumi operations
func (deh *DeploymentEventHandler) GetProgressStreams() io.Writer {
	return deh.progressWriter
}

// OnStackCreated handles stack creation events
func (deh *DeploymentEventHandler) OnStackCreated(environment, stackName string) {
	event := &DeploymentEvent{
		Type:        EventStackCreated,
		Environment: environment,
		StackName:   stackName,
		Message:     fmt.Sprintf("Stack %s created for environment %s", stackName, environment),
		Timestamp:   time.Now(),
		Data:        map[string]interface{}{},
	}
	
	deh.handleEvent(event)
}

// OnDeploymentStarted handles deployment start events
func (deh *DeploymentEventHandler) OnDeploymentStarted(environment string, result *DeploymentResult) {
	event := &DeploymentEvent{
		Type:        EventDeploymentStarted,
		Environment: environment,
		StackName:   result.StackName,
		Message:     fmt.Sprintf("Deployment started for environment %s", environment),
		Timestamp:   time.Now(),
		Data: map[string]interface{}{
			"operation": result.Operation,
			"status":    result.Status,
		},
	}
	
	deh.handleEvent(event)
}

// OnDeploymentCompleted handles deployment completion events
func (deh *DeploymentEventHandler) OnDeploymentCompleted(environment string, result *DeploymentResult) {
	event := &DeploymentEvent{
		Type:        EventDeploymentCompleted,
		Environment: environment,
		StackName:   result.StackName,
		Message:     fmt.Sprintf("Deployment completed for environment %s with status %s", environment, result.Status),
		Timestamp:   time.Now(),
		Data: map[string]interface{}{
			"operation": result.Operation,
			"status":    result.Status,
			"duration":  result.Duration.String(),
			"resources": len(result.Resources),
		},
	}
	
	if result.Error != nil {
		event.Data["error"] = result.Error.Error()
	}
	
	// Update metrics
	deh.updateMetrics(environment, result)
	
	deh.handleEvent(event)
}

// OnDestroyStarted handles destroy start events
func (deh *DeploymentEventHandler) OnDestroyStarted(environment string, result *DeploymentResult) {
	event := &DeploymentEvent{
		Type:        EventDestroyStarted,
		Environment: environment,
		StackName:   result.StackName,
		Message:     fmt.Sprintf("Destroy started for environment %s", environment),
		Timestamp:   time.Now(),
		Data: map[string]interface{}{
			"operation": result.Operation,
		},
	}
	
	deh.handleEvent(event)
}

// OnValidationStarted handles validation start events
func (deh *DeploymentEventHandler) OnValidationStarted(environment string, validationType string) {
	event := &DeploymentEvent{
		Type:        EventValidationStarted,
		Environment: environment,
		Message:     fmt.Sprintf("Validation started for environment %s: %s", environment, validationType),
		Timestamp:   time.Now(),
		Data: map[string]interface{}{
			"validationType": validationType,
		},
	}
	
	deh.handleEvent(event)
}

// OnValidationCompleted handles validation completion events
func (deh *DeploymentEventHandler) OnValidationCompleted(environment string, validationType string, success bool, errors []string) {
	event := &DeploymentEvent{
		Type:        EventValidationCompleted,
		Environment: environment,
		Message:     fmt.Sprintf("Validation completed for environment %s: %s (success: %v)", environment, validationType, success),
		Timestamp:   time.Now(),
		Data: map[string]interface{}{
			"validationType": validationType,
			"success":        success,
			"errors":         errors,
		},
	}
	
	deh.handleEvent(event)
}

// OnError handles error events
func (deh *DeploymentEventHandler) OnError(environment string, operation string, err error) {
	event := &DeploymentEvent{
		Type:        EventError,
		Environment: environment,
		Message:     fmt.Sprintf("Error in %s for environment %s: %v", operation, environment, err),
		Timestamp:   time.Now(),
		Data: map[string]interface{}{
			"operation": operation,
			"error":     err.Error(),
		},
	}
	
	deh.handleEvent(event)
}

// handleEvent processes an event through all registered handlers
func (deh *DeploymentEventHandler) handleEvent(event *DeploymentEvent) {
	// Log to console
	deh.logger.Printf("[%s] %s: %s", event.Type, event.Environment, event.Message)
	
	// Send to all registered handlers
	for _, handler := range deh.eventHandlers {
		handler.OnEvent(event)
	}
}

// updateMetrics updates deployment metrics
func (deh *DeploymentEventHandler) updateMetrics(environment string, result *DeploymentResult) {
	deh.metrics.TotalDeployments++
	deh.metrics.DeploymentsByEnvironment[environment]++
	deh.metrics.LastDeploymentTime = result.Timestamp
	
	if result.Status == DeploymentStatusSucceeded {
		deh.metrics.SuccessfulDeployments++
	} else {
		deh.metrics.FailedDeployments++
	}
	
	// Update average deploy time
	if deh.metrics.TotalDeployments > 0 {
		totalTime := deh.metrics.AverageDeployTime * time.Duration(deh.metrics.TotalDeployments-1)
		totalTime += result.Duration
		deh.metrics.AverageDeployTime = totalTime / time.Duration(deh.metrics.TotalDeployments)
	} else {
		deh.metrics.AverageDeployTime = result.Duration
	}
}

// GetMetrics returns current deployment metrics
func (deh *DeploymentEventHandler) GetMetrics() *DeploymentMetrics {
	return deh.metrics
}

// ConsoleEventHandler implementation
func NewConsoleEventHandler() *ConsoleEventHandler {
	return &ConsoleEventHandler{
		logger: log.New(os.Stdout, "[EVENT] ", log.LstdFlags),
	}
}

func (ceh *ConsoleEventHandler) OnEvent(event *DeploymentEvent) {
	ceh.logger.Printf("[%s] %s - %s", event.Type, event.Environment, event.Message)
	
	// Log additional data for certain events
	if event.Type == EventDeploymentCompleted {
		if duration, ok := event.Data["duration"]; ok {
			ceh.logger.Printf("  Duration: %s", duration)
		}
		if resources, ok := event.Data["resources"]; ok {
			ceh.logger.Printf("  Resources: %d", resources)
		}
	}
}

// WebhookEventHandler implementation
func NewWebhookEventHandler(webhookURL string) *WebhookEventHandler {
	return &WebhookEventHandler{
		webhookURL: webhookURL,
		// client would be initialized here
	}
}

func (weh *WebhookEventHandler) OnEvent(event *DeploymentEvent) {
	// TODO: Implement webhook posting
	// This would send the event to the configured webhook endpoint
	fmt.Printf("Webhook event: %s - %s\n", event.Type, event.Message)
}

// DeploymentEventCollector collects events for analysis
type DeploymentEventCollector struct {
	events []DeploymentEvent
}

func NewDeploymentEventCollector() *DeploymentEventCollector {
	return &DeploymentEventCollector{
		events: []DeploymentEvent{},
	}
}

func (dec *DeploymentEventCollector) OnEvent(event *DeploymentEvent) {
	dec.events = append(dec.events, *event)
}

func (dec *DeploymentEventCollector) GetEvents() []DeploymentEvent {
	return dec.events
}

func (dec *DeploymentEventCollector) GetEventsByType(eventType EventType) []DeploymentEvent {
	var filtered []DeploymentEvent
	for _, event := range dec.events {
		if event.Type == eventType {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

func (dec *DeploymentEventCollector) GetEventsByEnvironment(environment string) []DeploymentEvent {
	var filtered []DeploymentEvent
	for _, event := range dec.events {
		if event.Environment == environment {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// ClearEvents clears collected events
func (dec *DeploymentEventCollector) ClearEvents() {
	dec.events = []DeploymentEvent{}
}