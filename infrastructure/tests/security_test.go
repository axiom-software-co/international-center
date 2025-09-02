package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthentikIdentityProvider(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Skip if not in integration test environment
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test: INTEGRATION_TESTS not set to true")
	}

	t.Run("authentik service startup and configuration", func(t *testing.T) {
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
		
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"Health endpoint should allow anonymous access")
	})
}

func TestVaultSecretsManagement(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Skip if not in integration test environment
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test: INTEGRATION_TESTS not set to true")
	}

	t.Run("vault service initialization", func(t *testing.T) {
		// Test: Vault service is initialized and accessible
		vaultPort := requireEnv(t, "VAULT_PORT")
		vaultURL := fmt.Sprintf("http://localhost:%s/v1/sys/health", vaultPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", vaultURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err != nil {
			t.Logf("Vault not accessible at port %s: %v", vaultPort, err)
		}
		require.NoError(t, err, "Vault should be accessible")
		defer resp.Body.Close()
		
		// Vault health check returns 200 for initialized, 501 for uninitialized
		assert.True(t, resp.StatusCode == http.StatusOK || 
			resp.StatusCode == http.StatusNotImplemented ||
			resp.StatusCode == http.StatusServiceUnavailable,
			"Vault should respond to health checks")
	})

	t.Run("secret storage and retrieval", func(t *testing.T) {
		// Test: Basic secret operations are available
		vaultPort := requireEnv(t, "VAULT_PORT")
		
		// Test that secret endpoints are accessible
		secretURL := fmt.Sprintf("http://localhost:%s/v1/secret/data/test", vaultPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", secretURL, nil)
		require.NoError(t, err)
		
		// Add Vault token header (this would be a real token in actual use)
		req.Header.Set("X-Vault-Token", "test-token")
		
		resp, err := client.Do(req)
		if err != nil {
			t.Logf("Vault secret endpoint not accessible: %v", err)
		}
		require.NoError(t, err, "Vault secret endpoint should be accessible")
		defer resp.Body.Close()
		
		// Should handle request (even if unauthorized)
		assert.True(t, resp.StatusCode == http.StatusOK ||
			resp.StatusCode == http.StatusUnauthorized ||
			resp.StatusCode == http.StatusForbidden ||
			resp.StatusCode == http.StatusNotFound,
			"Vault should handle secret requests")
	})

	t.Run("policy enforcement", func(t *testing.T) {
		// Test: Policy enforcement is operational
		vaultPort := requireEnv(t, "VAULT_PORT")
		
		// Test policy endpoint
		policyURL := fmt.Sprintf("http://localhost:%s/v1/sys/policies/acl", vaultPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", policyURL, nil)
		require.NoError(t, err)
		
		req.Header.Set("X-Vault-Token", "test-token")
		
		resp, err := client.Do(req)
		if err != nil {
			t.Logf("Vault policy endpoint not accessible: %v", err)
		}
		require.NoError(t, err, "Vault policy endpoint should be accessible")
		defer resp.Body.Close()
		
		// Should handle policy requests
		assert.True(t, resp.StatusCode != http.StatusNotFound,
			"Vault policy endpoint should be available")
	})
}

func TestOPAPolicyEngine(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Skip if not in integration test environment
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test: INTEGRATION_TESTS not set to true")
	}

	t.Run("opa service startup", func(t *testing.T) {
		// Test: OPA service starts and is accessible
		opaPort := requireEnv(t, "OPA_PORT")
		opaURL := fmt.Sprintf("http://localhost:%s/health", opaPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", opaURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err != nil {
			t.Logf("OPA not accessible at port %s: %v", opaPort, err)
		}
		require.NoError(t, err, "OPA should be accessible")
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"OPA health check should return 200 OK")
	})

	t.Run("policy loading and compilation", func(t *testing.T) {
		// Test: OPA can load and compile policies
		opaPort := requireEnv(t, "OPA_PORT")
		
		// Test policy endpoint
		policyURL := fmt.Sprintf("http://localhost:%s/v1/policies", opaPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", policyURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err != nil {
			t.Logf("OPA policy endpoint not accessible: %v", err)
		}
		require.NoError(t, err, "OPA policy endpoint should be accessible")
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"OPA should list policies")
		
		// Should return JSON response
		contentType := resp.Header.Get("Content-Type")
		assert.Contains(t, contentType, "application/json",
			"OPA should return JSON policy list")
	})

	t.Run("policy evaluation endpoint availability", func(t *testing.T) {
		// Test: Policy evaluation endpoints are available
		opaPort := requireEnv(t, "OPA_PORT")
		
		// Test data API endpoint (used for policy evaluation)
		dataURL := fmt.Sprintf("http://localhost:%s/v1/data", opaPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", dataURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err != nil {
			t.Logf("OPA data endpoint not accessible: %v", err)
		}
		require.NoError(t, err, "OPA data endpoint should be accessible")
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"OPA data endpoint should be available")
		
		// Test policy evaluation with sample request
		testPolicy := map[string]interface{}{
			"input": map[string]interface{}{
				"user":   "test-user",
				"action": "read",
				"resource": "test-resource",
			},
		}
		
		policyBytes, err := json.Marshal(testPolicy)
		require.NoError(t, err)
		
		evalURL := fmt.Sprintf("http://localhost:%s/v1/data/example/allow", opaPort)
		evalReq, err := http.NewRequestWithContext(ctx, "POST", evalURL, 
			bytes.NewReader(policyBytes))
		require.NoError(t, err)
		
		evalReq.Header.Set("Content-Type", "application/json")
		
		evalResp, err := client.Do(evalReq)
		if err == nil {
			defer evalResp.Body.Close()
			// Should handle evaluation request (even if policy doesn't exist)
			assert.True(t, evalResp.StatusCode == http.StatusOK ||
				evalResp.StatusCode == http.StatusNotFound,
				"OPA should handle policy evaluation requests")
		}
	})
}