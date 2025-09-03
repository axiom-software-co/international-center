package auth

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestPolicyEvaluator_EvaluateAccess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name           string
		userID         string
		roles          []string
		resource       string
		action         string
		expectedAllow  bool
		expectedReason string
	}{
		{
			name:           "admin user can access admin endpoints",
			userID:         "admin-user-123",
			roles:          []string{"admin"},
			resource:       "/admin/api/v1/users",
			action:         "GET",
			expectedAllow:  true,
			expectedReason: "admin role permits all admin operations",
		},
		{
			name:           "regular user cannot access admin endpoints", 
			userID:         "user-456",
			roles:          []string{"user"},
			resource:       "/admin/api/v1/users",
			action:         "GET",
			expectedAllow:  false,
			expectedReason: "insufficient permissions",
		},
		{
			name:           "anonymous user can access public endpoints",
			userID:         "",
			roles:          []string{},
			resource:       "/api/v1/public/health",
			action:         "GET",
			expectedAllow:  true,
			expectedReason: "public endpoint allows anonymous access",
		},
		{
			name:           "anonymous user cannot access protected endpoints",
			userID:         "",
			roles:          []string{},
			resource:       "/api/v1/patients",
			action:         "GET",
			expectedAllow:  false,
			expectedReason: "authentication required",
		},
	}

	evaluator := NewTestPolicyEvaluator()
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Determine gateway based on resource
			gateway := "admin-gateway"
			if !strings.HasPrefix(tt.resource, "/admin/") {
				gateway = "public-gateway"
			}
			
			request := &PolicyRequest{
				UserID:   tt.userID,
				Roles:    tt.roles,
				Resource: tt.resource,
				Action:   tt.action,
				Gateway:  gateway,
			}

			decision, err := evaluator.EvaluateAccess(ctx, request)
			if err != nil {
				t.Fatalf("EvaluateAccess() error = %v", err)
			}

			if decision.Allow != tt.expectedAllow {
				t.Errorf("EvaluateAccess() Allow = %v, want %v", decision.Allow, tt.expectedAllow)
			}

			if decision.Reason != tt.expectedReason {
				t.Errorf("EvaluateAccess() Reason = %v, want %v", decision.Reason, tt.expectedReason)
			}
		})
	}
}

func TestPolicyEvaluator_EvaluateRateLimit(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name           string
		userID         string
		clientIP       string
		gateway        string
		expectedLimit  int
		expectedWindow time.Duration
	}{
		{
			name:           "admin gateway user rate limit",
			userID:         "user-123",
			clientIP:       "192.168.1.1",
			gateway:        "admin-gateway",
			expectedLimit:  100,
			expectedWindow: time.Minute,
		},
		{
			name:           "public gateway IP rate limit",
			userID:         "",
			clientIP:       "192.168.1.2", 
			gateway:        "public-gateway",
			expectedLimit:  1000,
			expectedWindow: time.Minute,
		},
	}

	evaluator := NewTestPolicyEvaluator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &RateLimitRequest{
				UserID:   tt.userID,
				ClientIP: tt.clientIP,
				Gateway:  tt.gateway,
			}

			limits, err := evaluator.EvaluateRateLimit(ctx, request)
			if err != nil {
				t.Fatalf("EvaluateRateLimit() error = %v", err)
			}

			if limits.RequestsPerWindow != tt.expectedLimit {
				t.Errorf("EvaluateRateLimit() RequestsPerWindow = %v, want %v", limits.RequestsPerWindow, tt.expectedLimit)
			}

			if limits.TimeWindow != tt.expectedWindow {
				t.Errorf("EvaluateRateLimit() TimeWindow = %v, want %v", limits.TimeWindow, tt.expectedWindow)
			}
		})
	}
}

func TestNewPolicyEvaluator(t *testing.T) {
	tests := []struct {
		name        string
		opaEndpoint string
		wantErr     bool
	}{
		{
			name:        "valid OPA endpoint",
			opaEndpoint: "http://localhost:8181",
			wantErr:     false,
		},
		{
			name:        "invalid OPA endpoint",
			opaEndpoint: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator := NewPolicyEvaluator(tt.opaEndpoint)
			if tt.wantErr && evaluator != nil {
				t.Errorf("NewPolicyEvaluator() expected error but got evaluator")
			}
			if !tt.wantErr && evaluator == nil {
				t.Errorf("NewPolicyEvaluator() expected evaluator but got nil")
			}
		})
	}
}