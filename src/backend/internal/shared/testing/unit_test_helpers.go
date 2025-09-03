package testing

import (
	"context"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
)

// CreateUnitTestContext creates a context with timeout for unit tests
func CreateUnitTestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

// MockDaprComponents provides mock implementations for unit testing
type MockDaprComponents struct {
	StateStore   map[string]interface{}
	PubSubEvents []MockPubSubEvent
	Secrets      map[string]string
	Bindings     map[string][]byte
}

// MockPubSubEvent represents a mock pub/sub event
type MockPubSubEvent struct {
	Topic string
	Data  interface{}
}

// NewMockDaprComponents creates new mock Dapr components
func NewMockDaprComponents() *MockDaprComponents {
	return &MockDaprComponents{
		StateStore: make(map[string]interface{}),
		Secrets:    make(map[string]string),
		Bindings:   make(map[string][]byte),
	}
}

// AddMockSecret adds a mock secret
func (m *MockDaprComponents) AddMockSecret(key, value string) {
	m.Secrets[key] = value
}

// AddMockBinding adds a mock binding response
func (m *MockDaprComponents) AddMockBinding(operation string, data []byte) {
	m.Bindings[operation] = data
}

// GetMockPubSubEvents returns all published events
func (m *MockDaprComponents) GetMockPubSubEvents(topic string) []MockPubSubEvent {
	var events []MockPubSubEvent
	for _, event := range m.PubSubEvents {
		if event.Topic == topic {
			events = append(events, event)
		}
	}
	return events
}

// MockStateStore provides mock state store operations for unit tests
type MockStateStore struct {
	data map[string]interface{}
}

// NewMockStateStore creates a new mock state store
func NewMockStateStore() *MockStateStore {
	return &MockStateStore{
		data: make(map[string]interface{}),
	}
}

// Save mocks saving state
func (m *MockStateStore) Save(ctx context.Context, key string, value interface{}, metadata map[string]string) error {
	m.data[key] = value
	return nil
}

// Get mocks getting state
func (m *MockStateStore) Get(ctx context.Context, key string, target interface{}) (bool, error) {
	if _, exists := m.data[key]; exists {
		// Simple mock implementation - in real tests you'd do proper unmarshaling
		return true, nil
	}
	return false, nil
}

// Delete mocks deleting state
func (m *MockStateStore) Delete(ctx context.Context, key string, metadata map[string]string) error {
	delete(m.data, key)
	return nil
}

// MockPubSub provides mock pub/sub operations for unit tests
type MockPubSub struct {
	publishedEvents []MockPubSubEvent
}

// NewMockPubSub creates a new mock pub/sub
func NewMockPubSub() *MockPubSub {
	return &MockPubSub{
		publishedEvents: make([]MockPubSubEvent, 0),
	}
}

// PublishEvent mocks publishing an event
func (m *MockPubSub) PublishEvent(ctx context.Context, topic string, data interface{}) error {
	m.publishedEvents = append(m.publishedEvents, MockPubSubEvent{
		Topic: topic,
		Data:  data,
	})
	return nil
}

// GetPublishedEvents returns all published events for a topic
func (m *MockPubSub) GetPublishedEvents(topic string) []MockPubSubEvent {
	var events []MockPubSubEvent
	for _, event := range m.publishedEvents {
		if event.Topic == topic {
			events = append(events, event)
		}
	}
	return events
}

// MockServiceInvocation provides mock service invocation for unit tests
type MockServiceInvocation struct {
	responses map[string]interface{}
}

// NewMockServiceInvocation creates a new mock service invocation
func NewMockServiceInvocation() *MockServiceInvocation {
	return &MockServiceInvocation{
		responses: make(map[string]interface{}),
	}
}

// SetMockResponse sets a mock response for a service call
func (m *MockServiceInvocation) SetMockResponse(appID, method string, response interface{}) {
	key := appID + "/" + method
	m.responses[key] = response
}

// InvokeService mocks invoking a service
func (m *MockServiceInvocation) InvokeService(ctx context.Context, appID, method string, data []byte) (interface{}, error) {
	key := appID + "/" + method
	if response, exists := m.responses[key]; exists {
		return response, nil
	}
	return nil, nil
}

// CreateTestCorrelationContext creates a test correlation context
func CreateTestCorrelationContext(userID string) *domain.CorrelationContext {
	ctx := domain.NewCorrelationContext()
	ctx.SetUserContext(userID, "test-1.0.0")
	return ctx
}

// CreateTestAuditEvent creates a test audit event
func CreateTestAuditEvent(entityType domain.EntityType, entityID, userID string, operationType domain.AuditEventType) *domain.AuditEvent {
	event := domain.NewAuditEvent(entityType, entityID, operationType, userID)
	event.SetEnvironmentContext("test", "test-1.0.0")
	return event
}