package tests

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRabbitMQConnectivity(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Skip if not in integration test environment
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test: INTEGRATION_TESTS not set to true")
	}

	t.Run("rabbitmq amqp service connectivity", func(t *testing.T) {
		// Test: RabbitMQ AMQP service is accessible
		rabbitmqPort := getEnvWithDefault("RABBITMQ_PORT", "5672")
		
		// Test RabbitMQ AMQP connectivity using TCP connection
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%s", rabbitmqPort), 5*time.Second)
		if err != nil {
			t.Logf("RabbitMQ AMQP not accessible on port %s: %v", rabbitmqPort, err)
		}
		require.NoError(t, err, "RabbitMQ AMQP should be accessible on port %s", rabbitmqPort)
		
		if conn != nil {
			conn.Close()
			t.Logf("RabbitMQ AMQP successfully connected on port %s", rabbitmqPort)
		}
	})

	t.Run("rabbitmq management interface accessible", func(t *testing.T) {
		// Test: RabbitMQ Management interface is accessible
		managementPort := getEnvWithDefault("RABBITMQ_MANAGEMENT_PORT", "15672")
		
		// Test RabbitMQ Management interface
		managementURL := fmt.Sprintf("http://localhost:%s", managementPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", managementURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err != nil {
			t.Logf("RabbitMQ Management interface not accessible on port %s: %v", managementPort, err)
		}
		require.NoError(t, err, "RabbitMQ Management interface should be accessible")
		defer resp.Body.Close()
		
		// Management interface should respond with authentication prompt or dashboard
		assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusUnauthorized,
			"RabbitMQ Management interface should respond")
	})

	t.Run("rabbitmq authentication configured", func(t *testing.T) {
		// Test: RabbitMQ authentication is properly configured
		rabbitmqUsername := os.Getenv("RABBITMQ_USERNAME")
		rabbitmqPassword := os.Getenv("RABBITMQ_PASSWORD")
		
		assert.NotEmpty(t, rabbitmqUsername, "RabbitMQ username should be configured")
		assert.NotEmpty(t, rabbitmqPassword, "RabbitMQ password should be configured")
	})

	t.Run("rabbitmq data persistence configured", func(t *testing.T) {
		// Test: RabbitMQ data persistence volume is configured
		rabbitmqDataVolume := os.Getenv("RABBITMQ_DATA_VOLUME")
		assert.NotEmpty(t, rabbitmqDataVolume, "RabbitMQ data volume should be configured")
	})
}