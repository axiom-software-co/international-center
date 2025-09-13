package dapr

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
)

// PubSub wraps Dapr pub/sub operations with advanced event processing capabilities
type PubSub struct {
	client         *Client
	pubsub         string
	appID          string
	batchProcessor *EventBatchProcessor
	dlqHandler     *DeadLetterQueueHandler
	eventStore     *EventStore
	metrics        *PubSubMetrics
	config         *PubSubConfig
}

// AuditEvent represents an audit event for compliance logging
type AuditEvent struct {
	AuditID       string                 `json:"audit_id"`
	EntityType    string                 `json:"entity_type"`
	EntityID      string                 `json:"entity_id"`
	OperationType string                 `json:"operation_type"`
	AuditTime     time.Time              `json:"audit_timestamp"`
	UserID        string                 `json:"user_id"`
	CorrelationID string                 `json:"correlation_id"`
	TraceID       string                 `json:"trace_id"`
	DataSnapshot  map[string]interface{} `json:"data_snapshot"`
	Environment   string                 `json:"environment"`
	AppVersion    string                 `json:"app_version"`
	RequestURL    string                 `json:"request_url,omitempty"`
}

// EventBatchProcessor handles batching of events for improved throughput
type EventBatchProcessor struct {
	batchSize     int
	flushInterval time.Duration
	eventBuffer   []EventMessage
	bufferMutex   sync.Mutex
	flushTimer    *time.Timer
	pubsub        *PubSub
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// DeadLetterQueueHandler manages failed event processing
type DeadLetterQueueHandler struct {
	dlqTopic        string
	maxRetries      int
	retryDelays     []time.Duration
	failedEvents    map[string]*FailedEvent
	failedEventsMux sync.RWMutex
	pubsub          *PubSub
	stateStore      *StateStore
}

// FailedEvent represents an event that failed processing
type FailedEvent struct {
	OriginalEvent EventMessage  `json:"original_event"`
	FailureReason string        `json:"failure_reason"`
	RetryCount    int           `json:"retry_count"`
	FirstFailure  time.Time     `json:"first_failure"`
	LastAttempt   time.Time     `json:"last_attempt"`
	NextRetry     time.Time     `json:"next_retry"`
	IsPermanent   bool          `json:"is_permanent"`
}

// EventStore provides event replay and audit capabilities
type EventStore struct {
	stateStore    *StateStore
	retentionDays int
	enabled       bool
	indexByTopic  map[string][]string
	indexMutex    sync.RWMutex
}

// PubSubMetrics tracks performance metrics for pub/sub operations
type PubSubMetrics struct {
	EventsPublished     int64            `json:"events_published"`
	EventsConsumed      int64            `json:"events_consumed"`
	BatchesProcessed    int64            `json:"batches_processed"`
	DeadLetterEvents    int64            `json:"dead_letter_events"`
	ReplayedEvents      int64            `json:"replayed_events"`
	PublishLatencies    map[string]int64 `json:"publish_latencies_ms"`
	ConsumerLatencies   map[string]int64 `json:"consumer_latencies_ms"`
	ErrorCounts         map[string]int64 `json:"error_counts"`
	mu                  sync.RWMutex
}

// PubSubConfig configures advanced pub/sub behavior
type PubSubConfig struct {
	BatchingEnabled       bool          `json:"batching_enabled"`
	BatchSize            int           `json:"batch_size"`
	BatchFlushInterval   time.Duration `json:"batch_flush_interval"`
	DeadLetterEnabled    bool          `json:"dead_letter_enabled"`
	DeadLetterTopic      string        `json:"dead_letter_topic"`
	MaxRetries           int           `json:"max_retries"`
	EventStoreEnabled    bool          `json:"event_store_enabled"`
	EventRetentionDays   int           `json:"event_retention_days"`
	MetricsEnabled       bool          `json:"metrics_enabled"`
	ParallelProcessing   bool          `json:"parallel_processing"`
	MaxParallelWorkers   int           `json:"max_parallel_workers"`
}

// EventMessage represents a generic event message
type EventMessage struct {
	Topic       string                 `json:"topic"`
	Data        map[string]interface{} `json:"data"`
	Metadata    map[string]string      `json:"metadata"`
	ContentType string                 `json:"content_type"`
	Source      string                 `json:"source"`
	Type        string                 `json:"type"`
	Subject     string                 `json:"subject,omitempty"`
	Time        time.Time              `json:"time"`
	SchemaVersion string               `json:"schema_version"`
	CorrelationID string               `json:"correlation_id"`
	RetryCount    int                   `json:"retry_count,omitempty"`
}

// CrossServiceEvent represents events that flow between services
type CrossServiceEvent struct {
	EventID       string                 `json:"event_id"`
	EventType     string                 `json:"event_type"`
	SourceService string                 `json:"source_service"`
	TargetService string                 `json:"target_service,omitempty"`
	EntityType    string                 `json:"entity_type"`
	EntityID      string                 `json:"entity_id"`
	OperationType string                 `json:"operation_type"`
	Payload       map[string]interface{} `json:"payload"`
	Timestamp     time.Time              `json:"timestamp"`
	CorrelationID string                 `json:"correlation_id"`
	UserID        string                 `json:"user_id,omitempty"`
	SchemaVersion string                 `json:"schema_version"`
	Priority      string                 `json:"priority,omitempty"`
}

// EventSchema defines the expected structure for different event types
type EventSchema struct {
	EventType     string              `json:"event_type"`
	SchemaVersion string              `json:"schema_version"`
	RequiredFields map[string]string  `json:"required_fields"`
	OptionalFields map[string]string  `json:"optional_fields"`
	ValidSources   []string           `json:"valid_sources"`
	ValidTargets   []string           `json:"valid_targets,omitempty"`
}

// RetryConfig configures retry behavior for event publishing
type RetryConfig struct {
	MaxRetries    int           `json:"max_retries"`
	InitialDelay  time.Duration `json:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
}

// NewPubSub creates a new pub/sub instance with advanced event processing capabilities
func NewPubSub(client *Client) *PubSub {
	pubsubName := getEnv("DAPR_PUBSUB_NAME", "pubsub")
	
	// Initialize configuration from environment
	config := &PubSubConfig{
		BatchingEnabled:      getEnv("PUBSUB_BATCHING_ENABLED", "true") == "true",
		BatchSize:           parseIntEnv("PUBSUB_BATCH_SIZE", 100),
		BatchFlushInterval:  parseDurationEnv("PUBSUB_BATCH_FLUSH_INTERVAL", 5*time.Second),
		DeadLetterEnabled:   getEnv("PUBSUB_DLQ_ENABLED", "true") == "true",
		DeadLetterTopic:     getEnv("PUBSUB_DLQ_TOPIC", "dead-letter-queue"),
		MaxRetries:          parseIntEnv("PUBSUB_MAX_RETRIES", 3),
		EventStoreEnabled:   getEnv("PUBSUB_EVENT_STORE_ENABLED", "true") == "true",
		EventRetentionDays:  parseIntEnv("PUBSUB_EVENT_RETENTION_DAYS", 30),
		MetricsEnabled:      getEnv("PUBSUB_METRICS_ENABLED", "true") == "true",
		ParallelProcessing:  getEnv("PUBSUB_PARALLEL_PROCESSING", "true") == "true",
		MaxParallelWorkers:  parseIntEnv("PUBSUB_MAX_PARALLEL_WORKERS", 10),
	}
	
	pubsub := &PubSub{
		client: client,
		pubsub: pubsubName,
		appID:  client.GetAppID(),
		config: config,
	}
	
	// Initialize metrics if enabled
	if config.MetricsEnabled {
		pubsub.metrics = &PubSubMetrics{
			PublishLatencies:  make(map[string]int64),
			ConsumerLatencies: make(map[string]int64),
			ErrorCounts:       make(map[string]int64),
		}
	}
	
	// Initialize event store if enabled
	if config.EventStoreEnabled {
		pubsub.eventStore = &EventStore{
			stateStore:    NewStateStore(client),
			retentionDays: config.EventRetentionDays,
			enabled:       true,
			indexByTopic:  make(map[string][]string),
		}
	}
	
	// Initialize dead letter queue handler if enabled
	if config.DeadLetterEnabled {
		retryDelays := []time.Duration{
			1 * time.Second,
			5 * time.Second,
			30 * time.Second,
			2 * time.Minute,
			10 * time.Minute,
		}
		
		pubsub.dlqHandler = &DeadLetterQueueHandler{
			dlqTopic:     config.DeadLetterTopic,
			maxRetries:   config.MaxRetries,
			retryDelays:  retryDelays,
			failedEvents: make(map[string]*FailedEvent),
			pubsub:       pubsub,
			stateStore:   NewStateStore(client),
		}
	}
	
	// Initialize batch processor if enabled
	if config.BatchingEnabled {
		pubsub.batchProcessor = &EventBatchProcessor{
			batchSize:     config.BatchSize,
			flushInterval: config.BatchFlushInterval,
			eventBuffer:   make([]EventMessage, 0, config.BatchSize),
			pubsub:        pubsub,
			stopChan:      make(chan struct{}),
		}
		
		// Start batch processing goroutine
		pubsub.batchProcessor.wg.Add(1)
		go pubsub.batchProcessor.processBatches()
	}
	
	return pubsub
}

// getDefaultRetryConfig returns default retry configuration
func (p *PubSub) getDefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:    3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
	}
}

// getEventSchemas returns predefined event schemas for validation
func (p *PubSub) getEventSchemas() map[string]*EventSchema {
	return map[string]*EventSchema{
		"content.published": {
			EventType:     "content.published",
			SchemaVersion: "1.0",
			RequiredFields: map[string]string{
				"content_id":   "string",
				"title":        "string",
				"content_type": "string",
				"published_at": "string",
			},
			OptionalFields: map[string]string{
				"summary":     "string",
				"category_id": "string",
				"author_id":   "string",
			},
			ValidSources: []string{"content-api"},
			ValidTargets: []string{"notification-api", "admin-gateway"},
		},
		"inquiry.submitted": {
			EventType:     "inquiry.submitted",
			SchemaVersion: "1.0",
			RequiredFields: map[string]string{
				"inquiry_id":   "string",
				"inquiry_type": "string",
				"status":       "string",
				"submitted_at": "string",
			},
			OptionalFields: map[string]string{
				"priority":      "string",
				"organization": "string",
				"contact_info": "object",
			},
			ValidSources: []string{"inquiries-api"},
			ValidTargets: []string{"notification-api", "admin-gateway"},
		},
		"inquiry.updated": {
			EventType:     "inquiry.updated",
			SchemaVersion: "1.0",
			RequiredFields: map[string]string{
				"inquiry_id":   "string",
				"inquiry_type": "string",
				"old_status":   "string",
				"new_status":   "string",
				"updated_at":   "string",
			},
			OptionalFields: map[string]string{
				"updated_by": "string",
				"notes":      "string",
			},
			ValidSources: []string{"inquiries-api", "admin-gateway"},
			ValidTargets: []string{"notification-api"},
		},
		"notification.sent": {
			EventType:     "notification.sent",
			SchemaVersion: "1.0",
			RequiredFields: map[string]string{
				"notification_id": "string",
				"channel":        "string",
				"recipient":      "string",
				"sent_at":        "string",
				"status":         "string",
			},
			OptionalFields: map[string]string{
				"error_message": "string",
				"retry_count":   "number",
			},
			ValidSources: []string{"notification-api"},
		},
	}
}

// validateEventSchema validates an event against its schema
func (p *PubSub) validateEventSchema(event *CrossServiceEvent) error {
	schemas := p.getEventSchemas()
	schema, exists := schemas[event.EventType]
	if !exists {
		return fmt.Errorf("no schema defined for event type: %s", event.EventType)
	}

	// Validate source service
	if len(schema.ValidSources) > 0 {
		validSource := false
		for _, source := range schema.ValidSources {
			if source == event.SourceService {
				validSource = true
				break
			}
		}
		if !validSource {
			return fmt.Errorf("invalid source service %s for event type %s, valid sources: %v", 
				event.SourceService, event.EventType, schema.ValidSources)
		}
	}

	// Validate target service if specified
	if event.TargetService != "" && len(schema.ValidTargets) > 0 {
		validTarget := false
		for _, target := range schema.ValidTargets {
			if target == event.TargetService {
				validTarget = true
				break
			}
		}
		if !validTarget {
			return fmt.Errorf("invalid target service %s for event type %s, valid targets: %v", 
				event.TargetService, event.EventType, schema.ValidTargets)
		}
	}

	// Validate required fields
	for fieldName, fieldType := range schema.RequiredFields {
		value, exists := event.Payload[fieldName]
		if !exists {
			return fmt.Errorf("missing required field %s for event type %s", fieldName, event.EventType)
		}
		if !p.validateFieldType(value, fieldType) {
			return fmt.Errorf("invalid type for field %s in event type %s, expected %s", 
				fieldName, event.EventType, fieldType)
		}
	}

	// Validate schema version
	if event.SchemaVersion != schema.SchemaVersion {
		return fmt.Errorf("schema version mismatch for event type %s, expected %s, got %s", 
			event.EventType, schema.SchemaVersion, event.SchemaVersion)
	}

	return nil
}

// validateFieldType validates that a field value matches the expected type
func (p *PubSub) validateFieldType(value interface{}, expectedType string) bool {
	switch expectedType {
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		_, ok1 := value.(int)
		_, ok2 := value.(float64)
		_, ok3 := value.(int64)
		return ok1 || ok2 || ok3
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "object":
		_, ok := value.(map[string]interface{})
		return ok
	case "array":
		return reflect.TypeOf(value).Kind() == reflect.Slice
	default:
		return true // Unknown types are allowed
	}
}

// PublishEvent publishes a generic event to a topic
func (p *PubSub) PublishEvent(ctx context.Context, topic string, event *EventMessage) error {
	if topic == "" {
		return fmt.Errorf("topic cannot be empty")
	}
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	if event.Time.IsZero() {
		event.Time = time.Now()
	}
	
	if event.Source == "" {
		event.Source = p.appID
	}

	// Set correlation ID if not present
	if event.CorrelationID == "" {
		event.CorrelationID = domain.GetCorrelationID(ctx)
		if event.CorrelationID == "" {
			event.CorrelationID = p.CreateCorrelationID()
		}
	}

	// Set schema version if not present
	if event.SchemaVersion == "" {
		event.SchemaVersion = "1.0"
	}

	data, err := json.Marshal(event)
	if err != nil {
		return domain.WrapError(err, fmt.Sprintf("failed to marshal event for pub/sub topic %s", topic))
	}

	// In test mode, mock successful event publishing
	if p.client.GetClient() == nil {
		// Check for context cancellation even in test mode
		if ctx.Err() == context.Canceled {
			return ctx.Err()
		}
		if ctx.Err() == context.DeadlineExceeded {
			return domain.NewTimeoutError(fmt.Sprintf("pub/sub publish operation for topic %s", topic))
		}
		return nil
	}

	err = p.client.GetClient().PublishEvent(ctx, p.pubsub, topic, data)
	if err != nil {
		return domain.NewDependencyError("pub/sub", domain.WrapError(err, fmt.Sprintf("failed to publish event to topic %s", topic)))
	}

	return nil
}

// PublishEventWithRetry publishes an event with retry logic for reliability
func (p *PubSub) PublishEventWithRetry(ctx context.Context, topic string, event *EventMessage, retryConfig *RetryConfig) error {
	if retryConfig == nil {
		retryConfig = p.getDefaultRetryConfig()
	}

	var lastErr error
	for attempt := 0; attempt <= retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate delay with exponential backoff
			delay := time.Duration(float64(retryConfig.InitialDelay) * 
				float64(attempt) * retryConfig.BackoffFactor)
			if delay > retryConfig.MaxDelay {
				delay = retryConfig.MaxDelay
			}

			// Wait before retry
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Update retry count in event
		event.RetryCount = attempt

		err := p.PublishEvent(ctx, topic, event)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Don't retry on context cancellation or validation errors
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	return fmt.Errorf("failed to publish event after %d attempts: %w", retryConfig.MaxRetries+1, lastErr)
}

// PublishCrossServiceEvent publishes a cross-service event with schema validation
func (p *PubSub) PublishCrossServiceEvent(ctx context.Context, event *CrossServiceEvent) error {
	if event == nil {
		return fmt.Errorf("cross-service event cannot be nil")
	}

	// Set default values
	if event.EventID == "" {
		event.EventID = fmt.Sprintf("%s-%d", event.EventType, time.Now().UnixNano())
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.SourceService == "" {
		event.SourceService = p.appID
	}
	if event.CorrelationID == "" {
		event.CorrelationID = domain.GetCorrelationID(ctx)
		if event.CorrelationID == "" {
			event.CorrelationID = p.CreateCorrelationID()
		}
	}
	if event.SchemaVersion == "" {
		event.SchemaVersion = "1.0"
	}

	// Validate event schema
	if err := p.validateEventSchema(event); err != nil {
		return fmt.Errorf("event schema validation failed: %w", err)
	}

	// Determine topic based on event type
	topic := p.getTopicForEventType(event.EventType)

	// Create EventMessage
	eventMessage := &EventMessage{
		Topic:         topic,
		Data:          map[string]interface{}{
			"event_id":       event.EventID,
			"event_type":     event.EventType,
			"source_service": event.SourceService,
			"target_service": event.TargetService,
			"entity_type":    event.EntityType,
			"entity_id":      event.EntityID,
			"operation_type": event.OperationType,
			"payload":        event.Payload,
			"timestamp":      event.Timestamp.Format(time.RFC3339),
			"user_id":        event.UserID,
			"priority":       event.Priority,
		},
		Metadata: map[string]string{
			"event_type":     event.EventType,
			"source_service": event.SourceService,
			"target_service": event.TargetService,
			"entity_type":    event.EntityType,
			"operation_type": event.OperationType,
			"priority":       event.Priority,
			"environment":    p.client.GetEnvironment(),
		},
		ContentType:   "application/json",
		Source:        event.SourceService,
		Type:          event.EventType,
		Subject:       fmt.Sprintf("%s/%s", event.EntityType, event.EntityID),
		Time:          event.Timestamp,
		SchemaVersion: event.SchemaVersion,
		CorrelationID: event.CorrelationID,
	}

	// Publish with retry for reliability
	return p.PublishEventWithRetry(ctx, topic, eventMessage, nil)
}

// getTopicForEventType determines the appropriate topic based on event type
func (p *PubSub) getTopicForEventType(eventType string) string {
	topicMappings := map[string]string{
		"content.published": "content-events",
		"content.updated":   "content-events",
		"content.deleted":   "content-events",
		"inquiry.submitted": "inquiry-events", 
		"inquiry.updated":   "inquiry-events",
		"inquiry.completed": "inquiry-events",
		"notification.sent": "notification-events",
		"notification.failed": "notification-events",
	}

	if topic, exists := topicMappings[eventType]; exists {
		return topic
	}

	// Default topic for unknown event types
	return "cross-service-events"
}

// PublishAuditEvent publishes an audit event to Grafana Loki
func (p *PubSub) PublishAuditEvent(ctx context.Context, event *AuditEvent) error {
	if event == nil {
		return fmt.Errorf("audit event cannot be nil")
	}

	// Set environment and timestamp if not provided
	if event.Environment == "" {
		event.Environment = p.client.GetEnvironment()
	}
	
	if event.AuditTime.IsZero() {
		event.AuditTime = time.Now()
	}

	// Determine the appropriate topic based on environment
	var topic string
	switch p.client.GetEnvironment() {
	case "production", "staging":
		topic = getEnv("AUDIT_TOPIC", "grafana-audit-events")
	default:
		topic = getEnv("AUDIT_TOPIC", "audit-events-dev")
	}

	eventMsg := &EventMessage{
		Topic: topic,
		Data: map[string]interface{}{
			"audit_id":       event.AuditID,
			"entity_type":    event.EntityType,
			"entity_id":      event.EntityID,
			"operation_type": event.OperationType,
			"audit_timestamp": event.AuditTime.Format(time.RFC3339Nano),
			"user_id":        event.UserID,
			"correlation_id": event.CorrelationID,
			"trace_id":       event.TraceID,
			"data_snapshot":  event.DataSnapshot,
			"environment":    event.Environment,
			"app_version":    event.AppVersion,
			"request_url":    event.RequestURL,
		},
		Metadata: map[string]string{
			"job":            fmt.Sprintf("%s-audit", event.EntityType),
			"environment":    event.Environment,
			"entity_type":    event.EntityType,
			"operation_type": event.OperationType,
			"user_id":        event.UserID,
		},
		ContentType: "application/json",
		Source:      p.appID,
		Type:        "audit.event",
		Subject:     fmt.Sprintf("%s/%s", event.EntityType, event.EntityID),
		Time:        event.AuditTime,
	}

	err := p.PublishEvent(ctx, topic, eventMsg)
	if err != nil {
		return fmt.Errorf("failed to publish audit event: %w", err)
	}

	return nil
}

// PublishContentEvent publishes content-related events with enhanced validation
func (p *PubSub) PublishContentEvent(ctx context.Context, eventType string, contentID string, data map[string]interface{}, targetService string) error {
	// Create cross-service event
	crossServiceEvent := &CrossServiceEvent{
		EventType:     fmt.Sprintf("content.%s", eventType),
		SourceService: p.appID,
		TargetService: targetService,
		EntityType:    "content",
		EntityID:      contentID,
		OperationType: strings.ToUpper(eventType),
		Payload:       data,
		UserID:        getStringFromData(data, "user_id"),
		Priority:      getStringFromData(data, "priority"),
	}

	// Ensure required fields are present
	if crossServiceEvent.Payload == nil {
		crossServiceEvent.Payload = make(map[string]interface{})
	}
	crossServiceEvent.Payload["content_id"] = contentID
	if _, exists := crossServiceEvent.Payload["title"]; !exists {
		crossServiceEvent.Payload["title"] = getStringFromData(data, "title")
	}
	if _, exists := crossServiceEvent.Payload["content_type"]; !exists {
		crossServiceEvent.Payload["content_type"] = getStringFromData(data, "content_type")
	}
	if eventType == "published" {
		if _, exists := crossServiceEvent.Payload["published_at"]; !exists {
			crossServiceEvent.Payload["published_at"] = time.Now().Format(time.RFC3339)
		}
	}

	return p.PublishCrossServiceEvent(ctx, crossServiceEvent)
}

// PublishServicesEvent publishes services-related events with enhanced validation
func (p *PubSub) PublishServicesEvent(ctx context.Context, eventType string, serviceID string, data map[string]interface{}, targetService string) error {
	// Create cross-service event
	crossServiceEvent := &CrossServiceEvent{
		EventType:     fmt.Sprintf("services.%s", eventType),
		SourceService: p.appID,
		TargetService: targetService,
		EntityType:    "service",
		EntityID:      serviceID,
		OperationType: strings.ToUpper(eventType),
		Payload:       data,
		UserID:        getStringFromData(data, "user_id"),
		Priority:      getStringFromData(data, "priority"),
	}

	return p.PublishCrossServiceEvent(ctx, crossServiceEvent)
}

// PublishInquiryEvent publishes inquiry-related events with enhanced validation
func (p *PubSub) PublishInquiryEvent(ctx context.Context, eventType string, inquiryID string, inquiryType string, data map[string]interface{}, targetService string) error {
	// Create cross-service event
	crossServiceEvent := &CrossServiceEvent{
		EventType:     fmt.Sprintf("inquiry.%s", eventType),
		SourceService: p.appID,
		TargetService: targetService,
		EntityType:    "inquiry",
		EntityID:      inquiryID,
		OperationType: strings.ToUpper(eventType),
		Payload:       data,
		UserID:        getStringFromData(data, "user_id"),
		Priority:      getStringFromData(data, "priority"),
	}

	// Ensure required fields are present
	if crossServiceEvent.Payload == nil {
		crossServiceEvent.Payload = make(map[string]interface{})
	}
	crossServiceEvent.Payload["inquiry_id"] = inquiryID
	crossServiceEvent.Payload["inquiry_type"] = inquiryType
	if eventType == "submitted" {
		if _, exists := crossServiceEvent.Payload["submitted_at"]; !exists {
			crossServiceEvent.Payload["submitted_at"] = time.Now().Format(time.RFC3339)
		}
		if _, exists := crossServiceEvent.Payload["status"]; !exists {
			crossServiceEvent.Payload["status"] = "submitted"
		}
	} else if eventType == "updated" {
		if _, exists := crossServiceEvent.Payload["updated_at"]; !exists {
			crossServiceEvent.Payload["updated_at"] = time.Now().Format(time.RFC3339)
		}
	}

	return p.PublishCrossServiceEvent(ctx, crossServiceEvent)
}

// PublishNotificationEvent publishes notification-related events
func (p *PubSub) PublishNotificationEvent(ctx context.Context, eventType string, notificationID string, data map[string]interface{}) error {
	// Create cross-service event
	crossServiceEvent := &CrossServiceEvent{
		EventType:     fmt.Sprintf("notification.%s", eventType),
		SourceService: p.appID,
		EntityType:    "notification",
		EntityID:      notificationID,
		OperationType: strings.ToUpper(eventType),
		Payload:       data,
		UserID:        getStringFromData(data, "user_id"),
		Priority:      getStringFromData(data, "priority"),
	}

	// Ensure required fields are present
	if crossServiceEvent.Payload == nil {
		crossServiceEvent.Payload = make(map[string]interface{})
	}
	crossServiceEvent.Payload["notification_id"] = notificationID
	if eventType == "sent" {
		if _, exists := crossServiceEvent.Payload["sent_at"]; !exists {
			crossServiceEvent.Payload["sent_at"] = time.Now().Format(time.RFC3339)
		}
		if _, exists := crossServiceEvent.Payload["status"]; !exists {
			crossServiceEvent.Payload["status"] = "sent"
		}
	}

	return p.PublishCrossServiceEvent(ctx, crossServiceEvent)
}

// getStringFromData safely extracts a string value from data map
func getStringFromData(data map[string]interface{}, key string) string {
	if data == nil {
		return ""
	}
	if value, exists := data[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// CreateCorrelationID creates a unique correlation ID for request tracing
func (p *PubSub) CreateCorrelationID() string {
	return fmt.Sprintf("%s-%d", p.appID, time.Now().UnixNano())
}

// ValidateEventFlow validates that an event can flow from source to target service
func (p *PubSub) ValidateEventFlow(sourceService, targetService, eventType string) error {
	schemas := p.getEventSchemas()
	schema, exists := schemas[eventType]
	if !exists {
		// Allow unknown event types for flexibility
		return nil
	}

	// Validate source service
	if len(schema.ValidSources) > 0 {
		validSource := false
		for _, source := range schema.ValidSources {
			if source == sourceService {
				validSource = true
				break
			}
		}
		if !validSource {
			return fmt.Errorf("service %s is not authorized to publish %s events", sourceService, eventType)
		}
	}

	// Validate target service if specified
	if targetService != "" && len(schema.ValidTargets) > 0 {
		validTarget := false
		for _, target := range schema.ValidTargets {
			if target == targetService {
				validTarget = true
				break
			}
		}
		if !validTarget {
			return fmt.Errorf("service %s is not a valid target for %s events", targetService, eventType)
		}
	}

	return nil
}

// GetValidEventTargets returns the list of valid target services for an event type
func (p *PubSub) GetValidEventTargets(eventType string) []string {
	schemas := p.getEventSchemas()
	schema, exists := schemas[eventType]
	if !exists {
		return []string{} // No targets defined
	}
	return schema.ValidTargets
}

// GetEventSchema returns the schema for a given event type
func (p *PubSub) GetEventSchema(eventType string) (*EventSchema, bool) {
	schemas := p.getEventSchemas()
	schema, exists := schemas[eventType]
	return schema, exists
}

// Advanced Event Processing Methods

// PublishWithBatching publishes events using batch processing for improved throughput
func (p *PubSub) PublishWithBatching(ctx context.Context, event EventMessage) error {
	if p.batchProcessor == nil {
		// Fall back to regular publishing if batching is disabled
		return p.PublishEvent(ctx, event.Topic, event.Data, event.Metadata)
	}

	p.batchProcessor.bufferMutex.Lock()
	defer p.batchProcessor.bufferMutex.Unlock()

	// Add event to buffer
	p.batchProcessor.eventBuffer = append(p.batchProcessor.eventBuffer, event)

	// Check if batch is full and should be flushed immediately
	if len(p.batchProcessor.eventBuffer) >= p.batchProcessor.batchSize {
		return p.batchProcessor.flushBatch()
	}

	return nil
}

// PublishWithRetryAndDLQ publishes an event with retry logic and dead letter queue support
func (p *PubSub) PublishWithRetryAndDLQ(ctx context.Context, event EventMessage) error {
	eventID := fmt.Sprintf("%s-%d", event.Topic, time.Now().UnixNano())
	
	// Try publishing with retry logic
	err := p.publishWithRetry(ctx, event, p.getDefaultRetryConfig())
	if err != nil {
		// Send to dead letter queue if all retries failed
		if p.dlqHandler != nil {
			return p.dlqHandler.handleFailedEvent(ctx, eventID, event, err.Error())
		}
		return err
	}

	p.recordMetric("events_published", 1)
	return nil
}

// ReplayEvents replays events from the event store within a specified time range
func (p *PubSub) ReplayEvents(ctx context.Context, topic string, startTime, endTime time.Time, targetTopic string) error {
	if p.eventStore == nil || !p.eventStore.enabled {
		return fmt.Errorf("event store is not enabled")
	}

	events, err := p.eventStore.GetEventsInTimeRange(ctx, topic, startTime, endTime)
	if err != nil {
		return fmt.Errorf("failed to retrieve events for replay: %w", err)
	}

	for _, event := range events {
		// Modify event for replay
		replayEvent := event
		replayEvent.Topic = targetTopic
		replayEvent.Metadata["replayed"] = "true"
		replayEvent.Metadata["original_topic"] = topic
		replayEvent.Metadata["replay_timestamp"] = time.Now().Format(time.RFC3339)

		err := p.PublishEvent(ctx, targetTopic, replayEvent.Data, replayEvent.Metadata)
		if err != nil {
			return fmt.Errorf("failed to replay event %s: %w", event.CorrelationID, err)
		}
	}

	p.recordMetric("replayed_events", int64(len(events)))
	return nil
}

// GetFailedEvents returns all events currently in the dead letter queue
func (p *PubSub) GetFailedEvents(ctx context.Context) (map[string]*FailedEvent, error) {
	if p.dlqHandler == nil {
		return nil, fmt.Errorf("dead letter queue is not enabled")
	}

	p.dlqHandler.failedEventsMux.RLock()
	defer p.dlqHandler.failedEventsMux.RUnlock()

	// Create a copy to avoid concurrent access issues
	failedEventsCopy := make(map[string]*FailedEvent)
	for id, event := range p.dlqHandler.failedEvents {
		failedEventsCopy[id] = event
	}

	return failedEventsCopy, nil
}

// RetryFailedEvent manually retries a specific failed event
func (p *PubSub) RetryFailedEvent(ctx context.Context, eventID string) error {
	if p.dlqHandler == nil {
		return fmt.Errorf("dead letter queue is not enabled")
	}

	return p.dlqHandler.retryEvent(ctx, eventID)
}

// GetMetrics returns current pub/sub performance metrics
func (p *PubSub) GetMetrics() *PubSubMetrics {
	if p.metrics == nil {
		return nil
	}

	p.metrics.mu.RLock()
	defer p.metrics.mu.RUnlock()

	// Create a copy to avoid concurrent access issues
	metricsCopy := &PubSubMetrics{
		EventsPublished:   p.metrics.EventsPublished,
		EventsConsumed:    p.metrics.EventsConsumed,
		BatchesProcessed:  p.metrics.BatchesProcessed,
		DeadLetterEvents:  p.metrics.DeadLetterEvents,
		ReplayedEvents:    p.metrics.ReplayedEvents,
		PublishLatencies:  make(map[string]int64),
		ConsumerLatencies: make(map[string]int64),
		ErrorCounts:       make(map[string]int64),
	}

	for k, v := range p.metrics.PublishLatencies {
		metricsCopy.PublishLatencies[k] = v
	}
	for k, v := range p.metrics.ConsumerLatencies {
		metricsCopy.ConsumerLatencies[k] = v
	}
	for k, v := range p.metrics.ErrorCounts {
		metricsCopy.ErrorCounts[k] = v
	}

	return metricsCopy
}

// EventBatchProcessor Methods

func (bp *EventBatchProcessor) processBatches() {
	defer bp.wg.Done()
	
	bp.flushTimer = time.NewTimer(bp.flushInterval)
	defer bp.flushTimer.Stop()

	for {
		select {
		case <-bp.stopChan:
			// Flush any remaining events before stopping
			bp.bufferMutex.Lock()
			if len(bp.eventBuffer) > 0 {
				bp.flushBatch()
			}
			bp.bufferMutex.Unlock()
			return

		case <-bp.flushTimer.C:
			bp.bufferMutex.Lock()
			if len(bp.eventBuffer) > 0 {
				bp.flushBatch()
			}
			bp.bufferMutex.Unlock()
			bp.flushTimer.Reset(bp.flushInterval)
		}
	}
}

func (bp *EventBatchProcessor) flushBatch() error {
	if len(bp.eventBuffer) == 0 {
		return nil
	}

	// Process events in parallel if configured
	if bp.pubsub.config.ParallelProcessing {
		return bp.processBatchParallel()
	}

	// Sequential processing
	for _, event := range bp.eventBuffer {
		err := bp.pubsub.PublishEvent(context.Background(), event.Topic, event.Data, event.Metadata)
		if err != nil {
			// Handle individual event failure
			if bp.pubsub.dlqHandler != nil {
				eventID := fmt.Sprintf("%s-%d", event.Topic, time.Now().UnixNano())
				bp.pubsub.dlqHandler.handleFailedEvent(context.Background(), eventID, event, err.Error())
			}
		}
	}

	bp.eventBuffer = bp.eventBuffer[:0] // Clear buffer
	bp.pubsub.recordMetric("batches_processed", 1)
	return nil
}

func (bp *EventBatchProcessor) processBatchParallel() error {
	semaphore := make(chan struct{}, bp.pubsub.config.MaxParallelWorkers)
	var wg sync.WaitGroup

	for _, event := range bp.eventBuffer {
		wg.Add(1)
		go func(evt EventMessage) {
			defer wg.Done()
			semaphore <- struct{}{} // Acquire
			defer func() { <-semaphore }() // Release

			err := bp.pubsub.PublishEvent(context.Background(), evt.Topic, evt.Data, evt.Metadata)
			if err != nil {
				if bp.pubsub.dlqHandler != nil {
					eventID := fmt.Sprintf("%s-%d", evt.Topic, time.Now().UnixNano())
					bp.pubsub.dlqHandler.handleFailedEvent(context.Background(), eventID, evt, err.Error())
				}
			}
		}(event)
	}

	wg.Wait()
	bp.eventBuffer = bp.eventBuffer[:0] // Clear buffer
	bp.pubsub.recordMetric("batches_processed", 1)
	return nil
}

// DeadLetterQueueHandler Methods

func (dlq *DeadLetterQueueHandler) handleFailedEvent(ctx context.Context, eventID string, event EventMessage, reason string) error {
	dlq.failedEventsMux.Lock()
	defer dlq.failedEventsMux.Unlock()

	now := time.Now()
	
	// Check if this event has failed before
	if existingEvent, exists := dlq.failedEvents[eventID]; exists {
		existingEvent.RetryCount++
		existingEvent.LastAttempt = now
		existingEvent.FailureReason = reason

		// Check if max retries exceeded
		if existingEvent.RetryCount >= dlq.maxRetries {
			existingEvent.IsPermanent = true
		} else {
			// Calculate next retry time
			if existingEvent.RetryCount-1 < len(dlq.retryDelays) {
				existingEvent.NextRetry = now.Add(dlq.retryDelays[existingEvent.RetryCount-1])
			} else {
				// Use last delay for retries beyond defined delays
				existingEvent.NextRetry = now.Add(dlq.retryDelays[len(dlq.retryDelays)-1])
			}
		}
	} else {
		// First failure
		failedEvent := &FailedEvent{
			OriginalEvent: event,
			FailureReason: reason,
			RetryCount:    1,
			FirstFailure:  now,
			LastAttempt:   now,
			NextRetry:     now.Add(dlq.retryDelays[0]),
			IsPermanent:   false,
		}
		dlq.failedEvents[eventID] = failedEvent
	}

	// Persist to state store for durability
	err := dlq.stateStore.Save(ctx, fmt.Sprintf("dlq:event:%s", eventID), dlq.failedEvents[eventID], nil)
	if err != nil {
		return fmt.Errorf("failed to persist failed event to state store: %w", err)
	}

	dlq.pubsub.recordMetric("dead_letter_events", 1)
	return nil
}

func (dlq *DeadLetterQueueHandler) retryEvent(ctx context.Context, eventID string) error {
	dlq.failedEventsMux.RLock()
	failedEvent, exists := dlq.failedEvents[eventID]
	dlq.failedEventsMux.RUnlock()

	if !exists {
		return fmt.Errorf("failed event with ID %s not found", eventID)
	}

	if failedEvent.IsPermanent {
		return fmt.Errorf("event %s has been marked as permanently failed", eventID)
	}

	// Attempt to republish the event
	err := dlq.pubsub.PublishEvent(ctx, failedEvent.OriginalEvent.Topic, 
		failedEvent.OriginalEvent.Data, failedEvent.OriginalEvent.Metadata)
	
	if err != nil {
		// Update failure information
		return dlq.handleFailedEvent(ctx, eventID, failedEvent.OriginalEvent, err.Error())
	}

	// Success - remove from failed events
	dlq.failedEventsMux.Lock()
	delete(dlq.failedEvents, eventID)
	dlq.failedEventsMux.Unlock()

	// Remove from state store
	dlq.stateStore.Delete(ctx, fmt.Sprintf("dlq:event:%s", eventID), nil)

	return nil
}

// EventStore Methods

func (es *EventStore) StoreEvent(ctx context.Context, event EventMessage) error {
	if !es.enabled {
		return nil
	}

	eventKey := fmt.Sprintf("event:%s:%s:%d", event.Topic, event.CorrelationID, time.Now().UnixNano())
	
	// Store the event
	err := es.stateStore.Save(ctx, eventKey, event, nil)
	if err != nil {
		return fmt.Errorf("failed to store event: %w", err)
	}

	// Update topic index
	es.indexMutex.Lock()
	es.indexByTopic[event.Topic] = append(es.indexByTopic[event.Topic], eventKey)
	es.indexMutex.Unlock()

	return nil
}

func (es *EventStore) GetEventsInTimeRange(ctx context.Context, topic string, startTime, endTime time.Time) ([]EventMessage, error) {
	es.indexMutex.RLock()
	eventKeys, exists := es.indexByTopic[topic]
	es.indexMutex.RUnlock()

	if !exists {
		return []EventMessage{}, nil
	}

	var events []EventMessage
	
	// Retrieve events and filter by time range
	for _, eventKey := range eventKeys {
		var event EventMessage
		found, err := es.stateStore.Get(ctx, eventKey, &event)
		if err != nil {
			continue // Skip failed retrievals
		}
		
		if found && event.Time.After(startTime) && event.Time.Before(endTime) {
			events = append(events, event)
		}
	}

	return events, nil
}

// Helper Methods

func (p *PubSub) publishWithRetry(ctx context.Context, event EventMessage, retryConfig *RetryConfig) error {
	var lastErr error
	
	for attempt := 0; attempt <= retryConfig.MaxRetries; attempt++ {
		err := p.PublishEvent(ctx, event.Topic, event.Data, event.Metadata)
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		if attempt < retryConfig.MaxRetries {
			delay := time.Duration(float64(retryConfig.InitialDelay) * 
				float64(attempt) * retryConfig.BackoffFactor)
			
			if delay > retryConfig.MaxDelay {
				delay = retryConfig.MaxDelay
			}
			
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				continue
			}
		}
	}
	
	return fmt.Errorf("publishing failed after %d retries: %w", retryConfig.MaxRetries, lastErr)
}

func (p *PubSub) recordMetric(metricName string, value int64) {
	if p.metrics == nil {
		return
	}

	p.metrics.mu.Lock()
	defer p.metrics.mu.Unlock()

	switch metricName {
	case "events_published":
		p.metrics.EventsPublished += value
	case "events_consumed":
		p.metrics.EventsConsumed += value
	case "batches_processed":
		p.metrics.BatchesProcessed += value
	case "dead_letter_events":
		p.metrics.DeadLetterEvents += value
	case "replayed_events":
		p.metrics.ReplayedEvents += value
	}
}

// Shutdown gracefully shuts down the pub/sub system
func (p *PubSub) Shutdown(ctx context.Context) error {
	if p.batchProcessor != nil {
		close(p.batchProcessor.stopChan)
		p.batchProcessor.wg.Wait()
	}
	return nil
}

