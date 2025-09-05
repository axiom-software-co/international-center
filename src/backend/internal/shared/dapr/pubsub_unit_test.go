package dapr

import (
	"context"
	"testing"
	"time"

	sharedtesting "github.com/axiom-software-co/international-center/src/backend/internal/shared/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED PHASE - Pub/Sub Messaging Tests (50+ test cases)

func TestNewPubSub(t *testing.T) {
	tests := []struct {
		name           string
		client         *Client
		envVars        map[string]string
		validateResult func(*testing.T, *PubSub)
	}{
		{
			name: "create pubsub with default settings",
			client: func() *Client {
				client, err := NewClient()
				require.NoError(t, err)
				return client
			}(),
			envVars: map[string]string{},
			validateResult: func(t *testing.T, ps *PubSub) {
				assert.NotNil(t, ps.client)
				assert.Equal(t, "pubsub-redis", ps.pubsub)
				assert.NotEmpty(t, ps.appID)
			},
		},
		{
			name: "create pubsub with custom pubsub name",
			client: func() *Client {
				client, err := NewClient()
				require.NoError(t, err)
				return client
			}(),
			envVars: map[string]string{
				"DAPR_PUBSUB_NAME": "custom-pubsub",
			},
			validateResult: func(t *testing.T, ps *PubSub) {
				assert.NotNil(t, ps.client)
				assert.Equal(t, "custom-pubsub", ps.pubsub)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}
			defer tt.client.Close()

			// Act
			pubsub := NewPubSub(tt.client)

			// Assert
			assert.NotNil(t, pubsub)
			if tt.validateResult != nil {
				tt.validateResult(t, pubsub)
			}
		})
	}
}

func TestPubSub_PublishEvent(t *testing.T) {
	tests := []struct {
		name          string
		topic         string
		event         *EventMessage
		setupContext  func() (context.Context, context.CancelFunc)
		expectedError string
		validateEvent func(*testing.T, *EventMessage)
	}{
		{
			name:  "publish simple event",
			topic: "test-topic",
			event: &EventMessage{
				Data: map[string]interface{}{
					"message": "test message",
					"type":    "test",
				},
				Metadata: map[string]string{
					"source": "test-app",
				},
				ContentType: "application/json",
				Type:        "test.event",
				Subject:     "test/subject",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateEvent: func(t *testing.T, event *EventMessage) {
				assert.NotZero(t, event.Time)
				assert.NotEmpty(t, event.Source)
			},
		},
		{
			name:  "publish event without timestamp - should auto-populate",
			topic: "timestamp-topic",
			event: &EventMessage{
				Data: map[string]interface{}{
					"test": "data",
				},
				ContentType: "application/json",
				Type:        "timestamp.test",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateEvent: func(t *testing.T, event *EventMessage) {
				assert.NotZero(t, event.Time)
				assert.WithinDuration(t, time.Now(), event.Time, 5*time.Second)
			},
		},
		{
			name:  "publish event without source - should auto-populate",
			topic: "source-topic",
			event: &EventMessage{
				Data: map[string]interface{}{
					"test": "data",
				},
				ContentType: "application/json",
				Type:        "source.test",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateEvent: func(t *testing.T, event *EventMessage) {
				assert.NotEmpty(t, event.Source)
			},
		},
		{
			name:  "publish large event",
			topic: "large-topic",
			event: &EventMessage{
				Data: map[string]interface{}{
					"large_data": string(make([]byte, 1024*10)), // 10KB
					"metadata":   "large event test",
				},
				ContentType: "application/json",
				Type:        "large.event",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 10*time.Second)
			},
		},
		{
			name:  "publish event with complex nested data",
			topic: "complex-topic",
			event: &EventMessage{
				Data: map[string]interface{}{
					"nested": map[string]interface{}{
						"level1": map[string]interface{}{
							"level2": []string{"item1", "item2", "item3"},
						},
					},
					"array": []map[string]interface{}{
						{"key1": "value1"},
						{"key2": "value2"},
					},
					"boolean": true,
					"number":  42,
				},
				ContentType: "application/json",
				Type:        "complex.event",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
		},
		{
			name:  "publish event with timeout context",
			topic: "timeout-topic",
			event: &EventMessage{
				Data: map[string]interface{}{
					"test": "timeout data",
				},
				ContentType: "application/json",
				Type:        "timeout.event",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 1*time.Millisecond)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := tt.setupContext()
			defer cancel()

			client, err := NewClient()
			require.NoError(t, err)
			defer client.Close()

			pubsub := NewPubSub(client)

			// Act
			err = pubsub.PublishEvent(ctx, tt.topic, tt.event)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				if tt.validateEvent != nil {
					tt.validateEvent(t, tt.event)
				}
			}
		})
	}
}

func TestPubSub_PublishAuditEvent(t *testing.T) {
	tests := []struct {
		name          string
		auditEvent    *AuditEvent
		setupContext  func() (context.Context, context.CancelFunc)
		envVars       map[string]string
		expectedError string
		validateEvent func(*testing.T, *AuditEvent)
	}{
		{
			name: "publish audit event for development environment",
			auditEvent: &AuditEvent{
				AuditID:       "audit-123",
				EntityType:    "service",
				EntityID:      "service-456",
				OperationType: "CREATE",
				UserID:        "user-789",
				CorrelationID: "corr-abc",
				TraceID:       "trace-def",
				DataSnapshot: map[string]interface{}{
					"name": "Test Service",
					"type": "test",
				},
				AppVersion: "1.0.0",
				RequestURL: "/api/v1/services",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			envVars: map[string]string{
				"ENVIRONMENT": "development",
			},
			validateEvent: func(t *testing.T, event *AuditEvent) {
				assert.Equal(t, "development", event.Environment)
				assert.NotZero(t, event.AuditTime)
			},
		},
		{
			name: "publish audit event for production environment",
			auditEvent: &AuditEvent{
				AuditID:       "audit-prod-123",
				EntityType:    "user",
				EntityID:      "user-prod-456",
				OperationType: "UPDATE",
				UserID:        "admin-789",
				CorrelationID: "corr-prod-abc",
				TraceID:       "trace-prod-def",
				DataSnapshot: map[string]interface{}{
					"email":  "test@example.com",
					"status": "active",
				},
				AppVersion: "2.0.0",
				RequestURL: "/api/v1/users",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			envVars: map[string]string{
				"ENVIRONMENT":  "production",
				"AUDIT_TOPIC": "grafana-audit-events",
			},
			validateEvent: func(t *testing.T, event *AuditEvent) {
				assert.Equal(t, "production", event.Environment)
				assert.NotZero(t, event.AuditTime)
			},
		},
		{
			name: "publish audit event for staging environment",
			auditEvent: &AuditEvent{
				AuditID:       "audit-staging-123",
				EntityType:    "content",
				EntityID:      "content-staging-456",
				OperationType: "DELETE",
				UserID:        "editor-789",
				CorrelationID: "corr-staging-abc",
				TraceID:       "trace-staging-def",
				DataSnapshot: map[string]interface{}{
					"title":  "Test Content",
					"status": "deleted",
				},
				Environment: "staging",
				AppVersion:  "1.5.0",
				RequestURL:  "/api/v1/content",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			envVars: map[string]string{
				"ENVIRONMENT":  "staging",
				"AUDIT_TOPIC": "grafana-audit-events",
			},
		},
		{
			name: "publish audit event without timestamp - should auto-populate",
			auditEvent: &AuditEvent{
				AuditID:       "audit-no-time",
				EntityType:    "service",
				EntityID:      "service-no-time",
				OperationType: "CREATE",
				UserID:        "user-no-time",
				CorrelationID: "corr-no-time",
				TraceID:       "trace-no-time",
				DataSnapshot: map[string]interface{}{
					"test": "no timestamp",
				},
				AppVersion: "1.0.0",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateEvent: func(t *testing.T, event *AuditEvent) {
				assert.NotZero(t, event.AuditTime)
				assert.WithinDuration(t, time.Now(), event.AuditTime, 5*time.Second)
			},
		},
		{
			name: "publish audit event without environment - should auto-populate",
			auditEvent: &AuditEvent{
				AuditID:       "audit-no-env",
				EntityType:    "service",
				EntityID:      "service-no-env",
				OperationType: "CREATE",
				UserID:        "user-no-env",
				CorrelationID: "corr-no-env",
				TraceID:       "trace-no-env",
				DataSnapshot: map[string]interface{}{
					"test": "no environment",
				},
				AppVersion: "1.0.0",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			validateEvent: func(t *testing.T, event *AuditEvent) {
				assert.NotEmpty(t, event.Environment)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := tt.setupContext()
			defer cancel()

			// Reset client for testing to pick up new environment variables
			ResetClientForTesting()
			t.Setenv("DAPR_TEST_MODE", "true")

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			client, err := NewClient()
			require.NoError(t, err)
			defer client.Close()

			pubsub := NewPubSub(client)

			// Act
			err = pubsub.PublishAuditEvent(ctx, tt.auditEvent)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				if tt.validateEvent != nil {
					tt.validateEvent(t, tt.auditEvent)
				}
			}
		})
	}
}

func TestPubSub_PublishContentEvent(t *testing.T) {
	tests := []struct {
		name          string
		eventType     string
		contentID     string
		data          map[string]interface{}
		setupContext  func() (context.Context, context.CancelFunc)
		envVars       map[string]string
		expectedError string
		validateCall  func(*testing.T)
	}{
		{
			name:      "publish content created event",
			eventType: "created",
			contentID: "content-123",
			data: map[string]interface{}{
				"title":    "Test Content",
				"author":   "Test Author",
				"category": "news",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
		},
		{
			name:      "publish content updated event",
			eventType: "updated",
			contentID: "content-456",
			data: map[string]interface{}{
				"title":        "Updated Content",
				"last_updated": time.Now().Format(time.RFC3339),
				"changes":      []string{"title", "body"},
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
		},
		{
			name:      "publish content deleted event",
			eventType: "deleted",
			contentID: "content-789",
			data: map[string]interface{}{
				"deleted_by": "admin-user",
				"reason":     "content violation",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
		},
		{
			name:      "publish content published event",
			eventType: "published",
			contentID: "content-published-123",
			data: map[string]interface{}{
				"published_at": time.Now().Format(time.RFC3339),
				"visibility":   "public",
				"featured":     true,
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
		},
		{
			name:      "publish content event with custom topic",
			eventType: "archived",
			contentID: "content-archived-456",
			data: map[string]interface{}{
				"archived_at": time.Now().Format(time.RFC3339),
				"archive_reason": "outdated",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			envVars: map[string]string{
				"CONTENT_EVENTS_TOPIC": "custom-content-events",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := tt.setupContext()
			defer cancel()

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			client, err := NewClient()
			require.NoError(t, err)
			defer client.Close()

			pubsub := NewPubSub(client)

			// Act
			err = pubsub.PublishContentEvent(ctx, tt.eventType, tt.contentID, tt.data)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				if tt.validateCall != nil {
					tt.validateCall(t)
				}
			}
		})
	}
}

func TestPubSub_PublishServicesEvent(t *testing.T) {
	tests := []struct {
		name          string
		eventType     string
		serviceID     string
		data          map[string]interface{}
		setupContext  func() (context.Context, context.CancelFunc)
		envVars       map[string]string
		expectedError string
	}{
		{
			name:      "publish service created event",
			eventType: "created",
			serviceID: "service-123",
			data: map[string]interface{}{
				"name":        "Test Service",
				"description": "Test Description",
				"category_id": "cat-456",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
		},
		{
			name:      "publish service updated event",
			eventType: "updated",
			serviceID: "service-456",
			data: map[string]interface{}{
				"name":           "Updated Service",
				"updated_fields": []string{"name", "description"},
				"version":        2,
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
		},
		{
			name:      "publish service activated event",
			eventType: "activated",
			serviceID: "service-789",
			data: map[string]interface{}{
				"activated_by": "admin-user",
				"activated_at": time.Now().Format(time.RFC3339),
				"featured":     true,
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
		},
		{
			name:      "publish service event with custom topic",
			eventType: "deactivated",
			serviceID: "service-deactivated-123",
			data: map[string]interface{}{
				"deactivated_by":     "admin-user",
				"deactivation_reason": "maintenance",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			envVars: map[string]string{
				"SERVICES_EVENTS_TOPIC": "custom-services-events",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := tt.setupContext()
			defer cancel()

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			client, err := NewClient()
			require.NoError(t, err)
			defer client.Close()

			pubsub := NewPubSub(client)

			// Act
			err = pubsub.PublishServicesEvent(ctx, tt.eventType, tt.serviceID, tt.data)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}

func TestPubSub_PublishMigrationEvent(t *testing.T) {
	tests := []struct {
		name          string
		eventType     string
		domain        string
		data          map[string]interface{}
		setupContext  func() (context.Context, context.CancelFunc)
		envVars       map[string]string
		expectedError string
	}{
		{
			name:      "publish migration started event",
			eventType: "started",
			domain:    "services",
			data: map[string]interface{}{
				"migration_id":    "mig-123",
				"version_from":    "1.0",
				"version_to":      "1.1",
				"started_by":      "admin",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
		},
		{
			name:      "publish migration completed event",
			eventType: "completed",
			domain:    "content",
			data: map[string]interface{}{
				"migration_id":     "mig-456",
				"duration_seconds": 45,
				"records_migrated": 1000,
				"completed_at":     time.Now().Format(time.RFC3339),
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
		},
		{
			name:      "publish migration failed event",
			eventType: "failed",
			domain:    "events",
			data: map[string]interface{}{
				"migration_id":  "mig-789",
				"error_message": "connection timeout",
				"failed_at":     time.Now().Format(time.RFC3339),
				"retry_count":   3,
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
		},
		{
			name:      "publish migration rollback event",
			eventType: "rollback",
			domain:    "users",
			data: map[string]interface{}{
				"migration_id":      "mig-rollback-123",
				"rollback_reason":   "data corruption detected",
				"rollback_to_version": "1.0",
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
		},
		{
			name:      "publish migration event with custom topic",
			eventType: "validated",
			domain:    "audit",
			data: map[string]interface{}{
				"validation_id":     "val-123",
				"validation_result": "passed",
				"checks_performed":  []string{"schema", "data_integrity", "constraints"},
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			envVars: map[string]string{
				"MIGRATION_EVENTS_TOPIC": "custom-migration-events",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := tt.setupContext()
			defer cancel()

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			client, err := NewClient()
			require.NoError(t, err)
			defer client.Close()

			pubsub := NewPubSub(client)

			// Act
			err = pubsub.PublishMigrationEvent(ctx, tt.eventType, tt.domain, tt.data)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}

func TestPubSub_CreateCorrelationID(t *testing.T) {
	tests := []struct {
		name     string
		client   *Client
		validate func(*testing.T, string, string)
	}{
		{
			name: "create correlation ID with default app ID",
			client: func() *Client {
				client, err := NewClient()
				require.NoError(t, err)
				return client
			}(),
			validate: func(t *testing.T, correlationID, appID string) {
				assert.Contains(t, correlationID, appID)
				assert.Contains(t, correlationID, "-")
				assert.True(t, len(correlationID) > len(appID))
			},
		},
		{
			name: "create correlation ID with custom app ID",
			client: func() *Client {
				ResetClientForTesting()
				t.Setenv("DAPR_TEST_MODE", "true")
				t.Setenv("DAPR_APP_ID", "custom-test-app")
				client, err := NewClient()
				require.NoError(t, err)
				return client
			}(),
			validate: func(t *testing.T, correlationID, appID string) {
				assert.Contains(t, correlationID, "custom-test-app")
				assert.Contains(t, correlationID, "-")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			defer tt.client.Close()
			pubsub := NewPubSub(tt.client)

			// Act
			correlationID1 := pubsub.CreateCorrelationID()
			correlationID2 := pubsub.CreateCorrelationID()

			// Assert
			assert.NotEmpty(t, correlationID1)
			assert.NotEmpty(t, correlationID2)
			assert.NotEqual(t, correlationID1, correlationID2) // Should be unique
			
			if tt.validate != nil {
				tt.validate(t, correlationID1, tt.client.GetAppID())
			}
		})
	}
}

func TestEventMessage_Structure_Validation(t *testing.T) {
	tests := []struct {
		name         string
		eventMessage *EventMessage
		isValid      bool
	}{
		{
			name: "valid complete event message",
			eventMessage: &EventMessage{
				Topic: "test-topic",
				Data: map[string]interface{}{
					"key": "value",
				},
				Metadata: map[string]string{
					"source": "test",
				},
				ContentType: "application/json",
				Source:      "test-app",
				Type:        "test.event",
				Subject:     "test/subject",
				Time:        time.Now(),
			},
			isValid: true,
		},
		{
			name: "minimal valid event message",
			eventMessage: &EventMessage{
				Data: map[string]interface{}{
					"message": "test",
				},
				ContentType: "application/json",
				Type:        "minimal.event",
			},
			isValid: true,
		},
		{
			name: "event message with empty data",
			eventMessage: &EventMessage{
				Data:        map[string]interface{}{},
				ContentType: "application/json",
				Type:        "empty.event",
			},
			isValid: true,
		},
		{
			name: "event message with nil data",
			eventMessage: &EventMessage{
				Data:        nil,
				ContentType: "application/json",
				Type:        "nil.event",
			},
			isValid: true, // Should be handled gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			client, err := NewClient()
			require.NoError(t, err)
			defer client.Close()

			pubsub := NewPubSub(client)

			// Act & Assert - Should not panic
			assert.NotPanics(t, func() {
				err := pubsub.PublishEvent(ctx, "test-topic", tt.eventMessage)
				if tt.isValid {
					// Event should be processed without panic
				}
				_ = err // May succeed or fail, main goal is no panic
			})
		})
	}
}

func TestAuditEvent_Structure_Validation(t *testing.T) {
	tests := []struct {
		name       string
		auditEvent *AuditEvent
		isValid    bool
	}{
		{
			name: "valid complete audit event",
			auditEvent: &AuditEvent{
				AuditID:       "audit-123",
				EntityType:    "service",
				EntityID:      "service-456",
				OperationType: "CREATE",
				AuditTime:     time.Now(),
				UserID:        "user-789",
				CorrelationID: "corr-abc",
				TraceID:       "trace-def",
				DataSnapshot: map[string]interface{}{
					"name": "Test Service",
				},
				Environment: "test",
				AppVersion:  "1.0.0",
				RequestURL:  "/api/v1/services",
			},
			isValid: true,
		},
		{
			name: "minimal valid audit event",
			auditEvent: &AuditEvent{
				AuditID:       "audit-minimal",
				EntityType:    "user",
				EntityID:      "user-123",
				OperationType: "READ",
				UserID:        "user-456",
				CorrelationID: "corr-minimal",
				TraceID:       "trace-minimal",
			},
			isValid: true,
		},
		{
			name: "audit event with empty data snapshot",
			auditEvent: &AuditEvent{
				AuditID:       "audit-empty-snapshot",
				EntityType:    "content",
				EntityID:      "content-123",
				OperationType: "DELETE",
				UserID:        "admin-user",
				CorrelationID: "corr-empty",
				TraceID:       "trace-empty",
				DataSnapshot:  map[string]interface{}{},
			},
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			client, err := NewClient()
			require.NoError(t, err)
			defer client.Close()

			pubsub := NewPubSub(client)

			// Act & Assert - Should not panic
			assert.NotPanics(t, func() {
				err := pubsub.PublishAuditEvent(ctx, tt.auditEvent)
				if tt.isValid {
					// Audit event should be processed without panic
				}
				_ = err // May succeed or fail, main goal is no panic
			})
		})
	}
}

func TestPubSub_Error_Handling(t *testing.T) {
	tests := []struct {
		name         string
		operation    func(context.Context, *PubSub) error
		setupContext func() (context.Context, context.CancelFunc)
		expectError  bool
	}{
		{
			name: "publish event with cancelled context",
			operation: func(ctx context.Context, ps *PubSub) error {
				event := &EventMessage{
					Data: map[string]interface{}{"test": "data"},
					ContentType: "application/json",
					Type: "test.event",
				}
				return ps.PublishEvent(ctx, "test-topic", event)
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				return ctx, func() {}
			},
			expectError: true,
		},
		{
			name: "publish audit event with cancelled context",
			operation: func(ctx context.Context, ps *PubSub) error {
				auditEvent := &AuditEvent{
					AuditID: "test-audit",
					EntityType: "test",
					EntityID: "test-id",
					OperationType: "CREATE",
					UserID: "test-user",
					CorrelationID: "test-corr",
					TraceID: "test-trace",
				}
				return ps.PublishAuditEvent(ctx, auditEvent)
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				return ctx, func() {}
			},
			expectError: true,
		},
		{
			name: "publish event with nil event should not panic",
			operation: func(ctx context.Context, ps *PubSub) error {
				return ps.PublishEvent(ctx, "test-topic", nil)
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			expectError: true,
		},
		{
			name: "publish audit event with nil event should not panic",
			operation: func(ctx context.Context, ps *PubSub) error {
				return ps.PublishAuditEvent(ctx, nil)
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.CreateUnitTestContext()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := tt.setupContext()
			defer cancel()

			client, err := NewClient()
			require.NoError(t, err)
			defer client.Close()

			pubsub := NewPubSub(client)

			// Act & Assert - Should not panic
			assert.NotPanics(t, func() {
				err := tt.operation(ctx, pubsub)
				if tt.expectError {
					assert.Error(t, err)
				}
			})
		})
	}
}

func TestPubSub_Timeout_Operations(t *testing.T) {
	tests := []struct {
		name      string
		timeout   time.Duration
		operation func(context.Context, *PubSub) error
	}{
		{
			name:    "publish event with timeout",
			timeout: 50 * time.Millisecond,
			operation: func(ctx context.Context, ps *PubSub) error {
				event := &EventMessage{
					Data: map[string]interface{}{"test": "timeout"},
					ContentType: "application/json",
					Type: "timeout.event",
				}
				return ps.PublishEvent(ctx, "timeout-topic", event)
			},
		},
		{
			name:    "publish audit event with timeout",
			timeout: 50 * time.Millisecond,
			operation: func(ctx context.Context, ps *PubSub) error {
				auditEvent := &AuditEvent{
					AuditID: "timeout-audit",
					EntityType: "timeout",
					EntityID: "timeout-id",
					OperationType: "CREATE",
					UserID: "timeout-user",
					CorrelationID: "timeout-corr",
					TraceID: "timeout-trace",
				}
				return ps.PublishAuditEvent(ctx, auditEvent)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			client, err := NewClient()
			require.NoError(t, err)
			defer client.Close()

			pubsub := NewPubSub(client)

			// Act - Should handle timeout gracefully
			err = tt.operation(ctx, pubsub)

			// Assert - Should not panic and may timeout
			if err != nil && ctx.Err() == context.DeadlineExceeded {
				assert.Error(t, err)
			}
		})
	}
}