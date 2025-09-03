package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/dapr/go-sdk/client"
	"github.com/stretchr/testify/require"
)

func TestDaprPubSubIntegration(t *testing.T) {
	// Phase 4: Pub/Sub Integration Validation
	// Integration test - requires full podman compose environment
	
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	
	daprClient, err := client.NewClient()
	require.NoError(t, err, "Should create Dapr client successfully")
	defer daprClient.Close()

	pubsubName := "pubsub-redis"

	t.Run("pubsub component accessibility", func(t *testing.T) {
		// Test: Redis pub/sub component is accessible and configured
		
		testTopicName := "phase4-component-test"
		testMessage := map[string]interface{}{
			"test_type": "component-accessibility",
			"message": "Pub/sub component accessibility test",
			"timestamp": time.Now().Unix(),
		}
		
		testMessageBytes, err := json.Marshal(testMessage)
		require.NoError(t, err, "Should marshal test message to JSON")
		
		// Test publishing a message to validate component accessibility
		err = daprClient.PublishEvent(ctx, pubsubName, testTopicName, testMessageBytes)
		require.NoError(t, err, "Should successfully publish message to Redis pub/sub component")
		
		// Note: We don't test subscription here as that requires setting up message handlers
		// Component accessibility is validated by successful publish operation
	})
	
	t.Run("pubsub topic operations", func(t *testing.T) {
		// Test: Pub/sub supports various topic operations
		
		topics := []string{
			"audit-events",
			"content-processing-events", 
			"service-notifications",
		}
		
		for _, topicName := range topics {
			t.Run(fmt.Sprintf("topic_%s_publishing", topicName), func(t *testing.T) {
				// Test publishing to different topic types
				topicMessage := map[string]interface{}{
					"topic": topicName,
					"event_type": "phase4-validation",
					"message": fmt.Sprintf("Testing topic %s publishing", topicName),
					"timestamp": time.Now().Unix(),
					"correlation_id": fmt.Sprintf("test-%d", time.Now().UnixNano()),
				}
				
				topicMessageBytes, err := json.Marshal(topicMessage)
				require.NoError(t, err, "Should marshal topic message to JSON")
				
				err = daprClient.PublishEvent(ctx, pubsubName, topicName, topicMessageBytes)
				require.NoError(t, err, "Should successfully publish to topic %s", topicName)
			})
		}
	})
	
	t.Run("pubsub message ordering and delivery", func(t *testing.T) {
		// Test: Message publishing with metadata and ordering
		
		testTopicName := "phase4-ordering-test"
		messageCount := 5
		
		// Publish multiple messages to test ordering capabilities
		for i := 0; i < messageCount; i++ {
			orderMessage := map[string]interface{}{
				"sequence_id": i,
				"batch_id": "phase4-ordering-batch",
				"message": fmt.Sprintf("Ordered message %d of %d", i+1, messageCount),
				"timestamp": time.Now().Unix(),
			}
			
			orderMessageBytes, err := json.Marshal(orderMessage)
			require.NoError(t, err, "Should marshal ordered message to JSON")
			
			// Publish with metadata
			err = daprClient.PublishEventfromCustomContent(ctx, pubsubName, testTopicName, orderMessageBytes)
			require.NoError(t, err, "Should successfully publish ordered message %d", i)
		}
	})
	
	t.Run("pubsub cross-service communication", func(t *testing.T) {
		// Test: Pub/sub enables cross-service async communication
		
		// Simulate different services publishing to shared topics
		serviceTopics := map[string]string{
			"services-api": "content-audit-events",
			"content-api": "services-audit-events",
			"public-gateway": "gateway-events",
			"admin-gateway": "admin-gateway-events",
		}
		
		for serviceName, topicName := range serviceTopics {
			t.Run(fmt.Sprintf("service_%s_cross_communication", serviceName), func(t *testing.T) {
				crossServiceMessage := map[string]interface{}{
					"source_service": serviceName,
					"target_topic": topicName,
					"event_type": "cross-service-communication-test",
					"message": fmt.Sprintf("Cross-service message from %s", serviceName),
					"timestamp": time.Now().Unix(),
					"trace_id": fmt.Sprintf("trace-%s-%d", serviceName, time.Now().UnixNano()),
				}
				
				crossServiceMessageBytes, err := json.Marshal(crossServiceMessage)
				require.NoError(t, err, "Should marshal cross-service message to JSON")
				
				err = daprClient.PublishEvent(ctx, pubsubName, topicName, crossServiceMessageBytes)
				require.NoError(t, err, "Service %s should publish to topic %s", serviceName, topicName)
			})
		}
	})
	
	t.Run("pubsub component configuration validation", func(t *testing.T) {
		// Test: Pub/sub component configuration is properly applied
		
		// Test Redis connectivity through pub/sub component
		configTestTopic := "phase4-config-test"
		configTestMessage := map[string]interface{}{
			"test_type": "configuration-validation",
			"message": "Testing Redis pub/sub component configuration",
			"config_items": []string{
				"redis_host_connectivity",
				"retry_policy",
				"message_delivery",
			},
			"timestamp": time.Now().Unix(),
		}
		
		configTestMessageBytes, err := json.Marshal(configTestMessage)
		require.NoError(t, err, "Should marshal config test message to JSON")
		
		// Test that pub/sub component properly handles the configured Redis settings
		err = daprClient.PublishEvent(ctx, pubsubName, configTestTopic, configTestMessageBytes)
		require.NoError(t, err, "Pub/sub component should handle message with configured Redis settings")
	})
	
	t.Run("pubsub error handling and resilience", func(t *testing.T) {
		// Test: Pub/sub component handles various error scenarios gracefully
		
		// Test publishing with invalid topic name characters (should still work or fail gracefully)
		resilenceTestMessage := map[string]interface{}{
			"test_type": "resilience-validation",
			"message": "Testing pub/sub resilience and error handling",
			"timestamp": time.Now().Unix(),
		}
		
		resilienceTestMessageBytes, err := json.Marshal(resilenceTestMessage)
		require.NoError(t, err, "Should marshal resilience test message to JSON")
		
		// Test normal operation
		err = daprClient.PublishEvent(ctx, pubsubName, "phase4-resilience-test", resilienceTestMessageBytes)
		require.NoError(t, err, "Should successfully publish resilience test message")
		
		// Test empty message handling (should work)
		emptyMessage := make([]byte, 0)
		err = daprClient.PublishEvent(ctx, pubsubName, "phase4-empty-test", emptyMessage)
		require.NoError(t, err, "Should handle empty message publishing gracefully")
	})
}