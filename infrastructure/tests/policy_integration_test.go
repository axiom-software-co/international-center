package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/internal/auth"
)

func TestPolicyIntegration_AdminGatewayRBAC(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Arrange - Set up test user and policies in OPA
	setupTestPolicies(t, ctx)
	defer cleanupTestPolicies(t, ctx)

	tests := []struct {
		name           string
		userID         string
		roles          []string
		endpoint       string
		method         string
		expectedStatus int
		expectedAllow  bool
	}{
		{
			name:           "admin user accesses user management endpoint",
			userID:         "admin-001",
			roles:          []string{"admin", "user_manager"},
			endpoint:       "/admin/api/v1/users",
			method:         "GET",
			expectedStatus: 200,
			expectedAllow:  true,
		},
		{
			name:           "regular user blocked from admin endpoint", 
			userID:         "user-002",
			roles:          []string{"user"},
			endpoint:       "/admin/api/v1/users",
			method:         "GET",
			expectedStatus: 403,
			expectedAllow:  false,
		},
		{
			name:           "healthcare staff accesses patient endpoint",
			userID:         "staff-003",
			roles:          []string{"healthcare_staff"},
			endpoint:       "/admin/api/v1/patients",
			method:         "GET",
			expectedStatus: 200,
			expectedAllow:  true,
		},
	}

	evaluator := auth.NewPolicyEvaluator("http://opa:8181")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act - Evaluate access policy
			request := &auth.PolicyRequest{
				UserID:   tt.userID,
				Roles:    tt.roles,
				Resource: tt.endpoint,
				Action:   tt.method,
				Gateway:  "admin-gateway",
			}

			decision, err := evaluator.EvaluateAccess(ctx, request)

			// Assert - Verify policy decision
			if err != nil {
				t.Fatalf("Policy evaluation failed: %v", err)
			}

			if decision.Allow != tt.expectedAllow {
				t.Errorf("Expected Allow=%v, got Allow=%v", tt.expectedAllow, decision.Allow)
			}

			// Verify OPA was called correctly by checking the decision reason
			if decision.Allow && decision.Reason == "" {
				t.Error("Expected non-empty reason for allowed access")
			}
			if !decision.Allow && decision.Reason == "" {
				t.Error("Expected non-empty reason for denied access")
			}
		})
	}
}

func TestPolicyIntegration_PublicGatewayAnonymousAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Arrange - Set up anonymous access policies
	setupTestPolicies(t, ctx)
	defer cleanupTestPolicies(t, ctx)

	tests := []struct {
		name           string
		endpoint       string
		method         string
		expectedAllow  bool
		expectedReason string
	}{
		{
			name:           "public health endpoint allows anonymous access",
			endpoint:       "/api/v1/public/health", 
			method:         "GET",
			expectedAllow:  true,
			expectedReason: "public endpoint",
		},
		{
			name:           "public information endpoint allows anonymous access",
			endpoint:       "/api/v1/public/info",
			method:         "GET", 
			expectedAllow:  true,
			expectedReason: "public endpoint",
		},
		{
			name:           "protected endpoint requires authentication",
			endpoint:       "/api/v1/patients",
			method:         "GET",
			expectedAllow:  false,
			expectedReason: "authentication required",
		},
	}

	evaluator := auth.NewPolicyEvaluator("http://opa:8181")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act - Evaluate anonymous access
			request := &auth.PolicyRequest{
				UserID:   "", // Anonymous user
				Roles:    []string{},
				Resource: tt.endpoint,
				Action:   tt.method,
				Gateway:  "public-gateway",
			}

			decision, err := evaluator.EvaluateAccess(ctx, request)

			// Assert - Verify anonymous access policy
			if err != nil {
				t.Fatalf("Anonymous policy evaluation failed: %v", err)
			}

			if decision.Allow != tt.expectedAllow {
				t.Errorf("Expected Allow=%v, got Allow=%v", tt.expectedAllow, decision.Allow)
			}

			if !contains(decision.Reason, tt.expectedReason) {
				t.Errorf("Expected reason to contain '%v', got '%v'", tt.expectedReason, decision.Reason)
			}
		})
	}
}

func TestPolicyIntegration_RateLimitEvaluation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Arrange - Set up rate limiting policies
	setupTestPolicies(t, ctx)
	defer cleanupTestPolicies(t, ctx)

	tests := []struct {
		name           string
		userID         string
		clientIP       string
		gateway        string
		expectedLimit  int
		expectedWindow time.Duration
	}{
		{
			name:           "admin gateway user-based rate limiting",
			userID:         "user-123",
			clientIP:       "192.168.1.100",
			gateway:        "admin-gateway", 
			expectedLimit:  100,
			expectedWindow: time.Minute,
		},
		{
			name:           "public gateway IP-based rate limiting",
			userID:         "",
			clientIP:       "192.168.1.200",
			gateway:        "public-gateway",
			expectedLimit:  1000, 
			expectedWindow: time.Minute,
		},
	}

	evaluator := auth.NewPolicyEvaluator("http://opa:8181")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act - Evaluate rate limits
			request := &auth.RateLimitRequest{
				UserID:   tt.userID,
				ClientIP: tt.clientIP,
				Gateway:  tt.gateway,
			}

			limits, err := evaluator.EvaluateRateLimit(ctx, request)

			// Assert - Verify rate limit policy 
			if err != nil {
				t.Fatalf("Rate limit evaluation failed: %v", err)
			}

			if limits.RequestsPerWindow != tt.expectedLimit {
				t.Errorf("Expected RequestsPerWindow=%v, got %v", tt.expectedLimit, limits.RequestsPerWindow)
			}

			if limits.TimeWindow != tt.expectedWindow {
				t.Errorf("Expected TimeWindow=%v, got %v", tt.expectedWindow, limits.TimeWindow)
			}
		})
	}
}

// setupTestPolicies loads test policies into OPA for integration testing
func setupTestPolicies(t *testing.T, ctx context.Context) {
	// Load admin gateway RBAC policies
	adminRBACPolicy := map[string]interface{}{
		"admin_gateway": map[string]interface{}{
			"rbac": map[string]interface{}{
				"admin_role_permissions": []string{
					"/admin/api/v1/users",
					"/admin/api/v1/roles", 
					"/admin/api/v1/audit",
				},
				"healthcare_staff_permissions": []string{
					"/admin/api/v1/patients",
					"/admin/api/v1/appointments",
				},
			},
		},
	}

	// Load public gateway anonymous access policies
	publicAccessPolicy := map[string]interface{}{
		"public_gateway": map[string]interface{}{
			"anonymous": map[string]interface{}{
				"allowed_endpoints": []string{
					"/api/v1/public/health",
					"/api/v1/public/info",
				},
			},
		},
	}

	// Load rate limiting policies  
	rateLimitPolicy := map[string]interface{}{
		"rate_limits": map[string]interface{}{
			"admin_gateway": map[string]interface{}{
				"requests_per_minute": 100,
				"window_seconds": 60,
			},
			"public_gateway": map[string]interface{}{
				"requests_per_minute": 1000,
				"window_seconds": 60,
			},
		},
	}

	// Send policies to OPA
	sendPolicyToOPA(t, ctx, "admin_rbac", adminRBACPolicy)
	sendPolicyToOPA(t, ctx, "public_access", publicAccessPolicy) 
	sendPolicyToOPA(t, ctx, "rate_limits", rateLimitPolicy)
}

// cleanupTestPolicies removes test policies from OPA after integration testing
func cleanupTestPolicies(t *testing.T, ctx context.Context) {
	deletePolicyFromOPA(t, ctx, "admin_rbac")
	deletePolicyFromOPA(t, ctx, "public_access")
	deletePolicyFromOPA(t, ctx, "rate_limits")
}

// sendPolicyToOPA uploads a policy document to OPA via REST API
func sendPolicyToOPA(t *testing.T, ctx context.Context, policyID string, policy interface{}) {
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		t.Fatalf("Failed to marshal policy: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", "http://localhost:8181/v1/data/"+policyID, bytes.NewBuffer(policyJSON))
	if err != nil {
		t.Fatalf("Failed to create policy request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send policy to OPA: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("OPA policy upload failed with status: %v", resp.StatusCode)
	}
}

// deletePolicyFromOPA removes a policy document from OPA via REST API
func deletePolicyFromOPA(t *testing.T, ctx context.Context, policyID string) {
	req, err := http.NewRequestWithContext(ctx, "DELETE", "http://localhost:8181/v1/data/"+policyID, nil)
	if err != nil {
		t.Logf("Failed to create policy delete request: %v", err)
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Logf("Failed to delete policy from OPA: %v", err)
		return
	}
	defer resp.Body.Close()
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(substr) > 0 && contains(s[1:], substr)) || (len(s) > 0 && s[0:len(substr)] == substr))
}