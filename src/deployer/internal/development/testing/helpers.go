package testing

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const integrationTestTimeout = 15 * time.Second

// GetRequiredEnvVar gets a required environment variable and fails the test if not present
func GetRequiredEnvVar(t *testing.T, key string) string {
	value := os.Getenv(key)
	require.NotEmpty(t, value, "Required environment variable %s must be set for integration tests", key)
	return value
}

// GetEnvVar gets an environment variable with optional default value
func GetEnvVar(key string, defaultValue ...string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

// MakeHTTPRequest creates and executes an HTTP request with timeout
func MakeHTTPRequest(t *testing.T, method, url string) *http.Response {
	ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	require.NoError(t, err, "Failed to create HTTP request")
	
	client := &http.Client{
		Timeout: integrationTestTimeout,
	}
	
	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to execute HTTP request")
	
	return resp
}

// ConnectWithTimeout attempts to establish a network connection with timeout
func ConnectWithTimeout(ctx context.Context, network, address string, timeout time.Duration) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout: timeout,
	}
	return dialer.DialContext(ctx, network, address)
}

// CreateIntegrationTestContext creates a context with timeout for integration tests
func CreateIntegrationTestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), integrationTestTimeout)
}

// makeHTTPRequest creates and executes an HTTP request with context and optional body
func makeHTTPRequest(ctx context.Context, method, url string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		switch v := body.(type) {
		case io.Reader:
			reqBody = v
		case string:
			reqBody = strings.NewReader(v)
		case []byte:
			reqBody = bytes.NewReader(v)
		default:
			reqBody = strings.NewReader(fmt.Sprintf("%v", v))
		}
	}
	
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, err
	}
	
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	client := &http.Client{
		Timeout: integrationTestTimeout,
	}
	
	return client.Do(req)
}