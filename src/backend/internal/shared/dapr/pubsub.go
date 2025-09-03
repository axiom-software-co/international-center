package dapr

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/dapr/go-sdk/client"
)

// PubSub wraps Dapr pub/sub operations
type PubSub struct {
	client  *Client
	pubsub  string
	appID   string
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
}

// NewPubSub creates a new pub/sub instance
func NewPubSub(client *Client) *PubSub {
	pubsubName := getEnv("DAPR_PUBSUB_NAME", "pubsub-redis")
	
	return &PubSub{
		client: client,
		pubsub: pubsubName,
		appID:  client.GetAppID(),
	}
}

// PublishEvent publishes a generic event to a topic
func (p *PubSub) PublishEvent(ctx context.Context, topic string, event *EventMessage) error {
	if event.Time.IsZero() {
		event.Time = time.Now()
	}
	
	if event.Source == "" {
		event.Source = p.appID
	}

	data, err := json.Marshal(event)
	if err != nil {
		return domain.WrapError(err, fmt.Sprintf("failed to marshal event for pub/sub topic %s", topic))
	}

	err = p.client.GetClient().PublishEvent(ctx, p.pubsub, topic, data)
	if err != nil {
		return domain.NewDependencyError("pub/sub", domain.WrapError(err, fmt.Sprintf("failed to publish event to topic %s", topic)))
	}

	return nil
}

// PublishAuditEvent publishes an audit event to Grafana Loki
func (p *PubSub) PublishAuditEvent(ctx context.Context, event *AuditEvent) error {
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

// PublishContentEvent publishes content-related events
func (p *PubSub) PublishContentEvent(ctx context.Context, eventType string, contentID string, data map[string]interface{}) error {
	topic := getEnv("CONTENT_EVENTS_TOPIC", "content-events")
	
	event := &EventMessage{
		Topic: topic,
		Data:  data,
		Metadata: map[string]string{
			"content_id":  contentID,
			"event_type":  eventType,
			"environment": p.client.GetEnvironment(),
		},
		ContentType: "application/json",
		Source:      p.appID,
		Type:        fmt.Sprintf("content.%s", eventType),
		Subject:     fmt.Sprintf("content/%s", contentID),
	}

	return p.PublishEvent(ctx, topic, event)
}

// PublishServicesEvent publishes services-related events
func (p *PubSub) PublishServicesEvent(ctx context.Context, eventType string, serviceID string, data map[string]interface{}) error {
	topic := getEnv("SERVICES_EVENTS_TOPIC", "services-events")
	
	event := &EventMessage{
		Topic: topic,
		Data:  data,
		Metadata: map[string]string{
			"service_id":  serviceID,
			"event_type":  eventType,
			"environment": p.client.GetEnvironment(),
		},
		ContentType: "application/json",
		Source:      p.appID,
		Type:        fmt.Sprintf("services.%s", eventType),
		Subject:     fmt.Sprintf("services/%s", serviceID),
	}

	return p.PublishEvent(ctx, topic, event)
}

// PublishMigrationEvent publishes migration-related events
func (p *PubSub) PublishMigrationEvent(ctx context.Context, eventType string, domain string, data map[string]interface{}) error {
	topic := getEnv("MIGRATION_EVENTS_TOPIC", "migration-events")
	
	event := &EventMessage{
		Topic: topic,
		Data:  data,
		Metadata: map[string]string{
			"domain":      domain,
			"event_type":  eventType,
			"environment": p.client.GetEnvironment(),
		},
		ContentType: "application/json",
		Source:      p.appID,
		Type:        fmt.Sprintf("migration.%s", eventType),
		Subject:     fmt.Sprintf("migration/%s", domain),
	}

	return p.PublishEvent(ctx, topic, event)
}

// CreateCorrelationID creates a unique correlation ID for request tracing
func (p *PubSub) CreateCorrelationID() string {
	return fmt.Sprintf("%s-%d", p.appID, time.Now().UnixNano())
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}