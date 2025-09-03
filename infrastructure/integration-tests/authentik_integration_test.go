package tests

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthentikIdentityProvider(t *testing.T) {
	// Integration test - requires full podman compose environment

	t.Run("authentik service startup and configuration", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Authentik service starts and is accessible
		authentikPort := requireEnv(t, "AUTHENTIK_PORT")
		authentikURL := fmt.Sprintf("http://localhost:%s/if/flow/default-authentication-flow/", authentikPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", authentikURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err != nil {
			t.Logf("Authentik not accessible at port %s: %v", authentikPort, err)
		}
		require.NoError(t, err, "Authentik should be accessible")
		defer resp.Body.Close()
		
		// Should return authentication flow page or redirect
		assert.True(t, resp.StatusCode == http.StatusOK || 
			resp.StatusCode == http.StatusFound ||
			resp.StatusCode == http.StatusTemporaryRedirect,
			"Authentik should serve authentication flow")
	})

	t.Run("oauth2 endpoint availability", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: OAuth2 endpoints are available
		authentikPort := requireEnv(t, "AUTHENTIK_PORT")
		
		// Test OAuth2 discovery endpoint
		discoveryURL := fmt.Sprintf("http://localhost:%s/application/o/.well-known/openid_configuration", authentikPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", discoveryURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err != nil {
			t.Logf("OAuth2 discovery endpoint not accessible: %v", err)
		}
		require.NoError(t, err, "OAuth2 discovery should be available")
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"OAuth2 discovery should return configuration")
		
		// Verify it returns JSON configuration
		contentType := resp.Header.Get("Content-Type")
		assert.Contains(t, contentType, "application/json",
			"Discovery should return JSON configuration")
	})

	t.Run("jwt token validation capability", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: JWT validation endpoints are available
		authentikPort := requireEnv(t, "AUTHENTIK_PORT")
		
		// Test token introspection endpoint
		tokenURL := fmt.Sprintf("http://localhost:%s/application/o/introspect/", authentikPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		
		// Create a dummy token introspection request
		tokenData := "token=dummy_token_for_test"
		req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, 
			bytes.NewReader([]byte(tokenData)))
		require.NoError(t, err)
		
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		
		resp, err := client.Do(req)
		if err != nil {
			t.Logf("Token introspection endpoint not accessible: %v", err)
		}
		require.NoError(t, err, "Token introspection should be available")
		defer resp.Body.Close()
		
		// Should handle request (even if token is invalid)
		assert.True(t, resp.StatusCode == http.StatusOK || 
			resp.StatusCode == http.StatusUnauthorized ||
			resp.StatusCode == http.StatusBadRequest,
			"Token introspection endpoint should handle requests")
	})

	t.Run("anonymous access configuration", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Anonymous access is properly configured
		authentikPort := requireEnv(t, "AUTHENTIK_PORT")
		
		// Test public endpoints that should allow anonymous access
		healthURL := fmt.Sprintf("http://localhost:%s/-/health/live/", authentikPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err != nil {
			t.Logf("Authentik health endpoint not accessible: %v", err)
		}
		require.NoError(t, err, "Health endpoint should be accessible")
		defer resp.Body.Close()
		
		assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent,
			"Health endpoint should allow anonymous access (200 or 204)")
	})
}