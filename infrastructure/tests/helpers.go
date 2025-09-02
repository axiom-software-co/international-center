package tests

import (
	"os"
	"testing"
)

// requireEnv retrieves environment variable or fails test if missing
func requireEnv(t *testing.T, key string) string {
	value := os.Getenv(key)
	if value == "" {
		t.Fatalf("Required environment variable %s is not set", key)
	}
	return value
}