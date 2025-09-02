package tests

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMongoDBConnectivity(t *testing.T) {
	// Skip if not in integration test environment
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test: INTEGRATION_TESTS not set to true")
	}

	t.Run("mongodb service connectivity", func(t *testing.T) {
		// Test: MongoDB service is accessible
		mongoPort := requireEnv(t, "MONGO_PORT")
		
		// Test MongoDB connectivity using TCP connection
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%s", mongoPort), 5*time.Second)
		if err != nil {
			t.Logf("MongoDB not accessible on port %s: %v", mongoPort, err)
		}
		require.NoError(t, err, "MongoDB should be accessible on port %s", mongoPort)
		
		if conn != nil {
			conn.Close()
			t.Logf("MongoDB successfully connected on port %s", mongoPort)
		}
	})

	t.Run("mongodb authentication configured", func(t *testing.T) {
		// Test: MongoDB authentication is properly configured
		mongoUsername := os.Getenv("MONGO_USERNAME")
		mongoPassword := os.Getenv("MONGO_PASSWORD")
		mongoDatabase := os.Getenv("MONGO_DB")
		
		assert.NotEmpty(t, mongoUsername, "MongoDB username should be configured")
		assert.NotEmpty(t, mongoPassword, "MongoDB password should be configured")
		assert.NotEmpty(t, mongoDatabase, "MongoDB database should be configured")
	})

	t.Run("mongodb data persistence configured", func(t *testing.T) {
		// Test: MongoDB data persistence volume is configured
		mongoDataVolume := os.Getenv("MONGO_DATA_VOLUME")
		assert.NotEmpty(t, mongoDataVolume, "MongoDB data volume should be configured")
	})
}