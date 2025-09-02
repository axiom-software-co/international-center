package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Constants for policy evaluation
const (
	// Gateway identifiers
	AdminGateway  = "admin-gateway"
	PublicGateway = "public-gateway"

	// OPA policy paths
	AdminRBACPath         = "authz/admin_gateway/rbac"
	AdminRateLimitPath    = "authz/admin_gateway/rate_limits"
	PublicAnonymousPath   = "authz/public_gateway/anonymous"
	PublicRateLimitPath   = "authz/public_gateway/rate_limits"

	// Default rate limits
	DefaultAdminRateLimit  = 100
	DefaultPublicRateLimit = 1000
	DefaultTimeWindow      = 60 // seconds

	// HTTP configuration
	DefaultTimeout = 5 * time.Second

	// Policy decision reasons
	ReasonAdminRolePermits       = "admin role permits all admin operations"
	ReasonInsufficientPermissions = "insufficient permissions"
	ReasonAuthenticationRequired  = "authentication required"
	ReasonPublicEndpoint         = "public endpoint allows anonymous access"
	ReasonUnknownGateway         = "unknown gateway"
	ReasonPolicyEvaluationError  = "policy evaluation error"
	ReasonNoMatchingPolicy       = "no matching policy"
)

// Role constants
const (
	RoleAdmin            = "admin"
	RoleUser             = "user"
	RoleHealthcareStaff  = "healthcare_staff"
	RoleUserManager      = "user_manager"
)

// AccessPolicyEvaluator defines the interface for evaluating access control policies
type AccessPolicyEvaluator interface {
	// EvaluateAccess evaluates whether a user/request should be granted access to a resource
	EvaluateAccess(ctx context.Context, request *PolicyRequest) (*PolicyDecision, error)
}

// RateLimitPolicyEvaluator defines the interface for evaluating rate limiting policies
type RateLimitPolicyEvaluator interface {
	// EvaluateRateLimit evaluates rate limiting policies for a user or client IP
	EvaluateRateLimit(ctx context.Context, request *RateLimitRequest) (*RateLimits, error)
}

// PolicyEvaluator defines the combined interface for both access and rate limiting policies
type PolicyEvaluator interface {
	AccessPolicyEvaluator
	RateLimitPolicyEvaluator
}

// PolicyRequest contains the context information needed for policy evaluation
type PolicyRequest struct {
	// UserID is the authenticated user identifier (empty for anonymous requests)
	UserID string `json:"user_id"`
	
	// Roles contains the user's assigned roles
	Roles []string `json:"roles"`
	
	// Resource is the requested API endpoint or resource path  
	Resource string `json:"resource"`
	
	// Action is the HTTP method or operation being performed
	Action string `json:"action"`
	
	// Gateway identifies which gateway is handling the request (admin-gateway or public-gateway)
	Gateway string `json:"gateway"`
	
	// ClientIP is the originating client IP address
	ClientIP string `json:"client_ip,omitempty"`
	
	// Headers contains relevant HTTP headers for policy evaluation
	Headers map[string]string `json:"headers,omitempty"`
	
	// QueryParams contains query parameters that may affect policy decisions
	QueryParams map[string]string `json:"query_params,omitempty"`
}

// PolicyDecision represents the result of a policy evaluation
type PolicyDecision struct {
	// Allow indicates whether the request should be permitted
	Allow bool `json:"allow"`
	
	// Reason provides a human-readable explanation for the decision
	Reason string `json:"reason"`
	
	// PolicyID identifies which policy rule made the decision
	PolicyID string `json:"policy_id,omitempty"`
	
	// Metadata contains additional policy-specific information
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// RateLimitRequest contains the context information for rate limit policy evaluation  
type RateLimitRequest struct {
	// UserID is the authenticated user identifier (empty for IP-based limiting)
	UserID string `json:"user_id"`
	
	// ClientIP is the originating client IP address for IP-based limiting
	ClientIP string `json:"client_ip"`
	
	// Gateway identifies which gateway's rate limits to apply
	Gateway string `json:"gateway"`
	
	// Endpoint is the specific API endpoint being accessed
	Endpoint string `json:"endpoint,omitempty"`
}

// RateLimits represents the rate limiting configuration for a request
type RateLimits struct {
	// RequestsPerWindow is the maximum number of requests allowed in the time window
	RequestsPerWindow int `json:"requests_per_window"`
	
	// TimeWindow is the duration of the rate limiting window
	TimeWindow time.Duration `json:"time_window"`
	
	// CurrentCount is the current number of requests in the window (if available)
	CurrentCount int `json:"current_count,omitempty"`
	
	// ResetTime is when the current window resets
	ResetTime time.Time `json:"reset_time,omitempty"`
}

// opaPolicyEvaluator implements the PolicyEvaluator interface using OPA
type opaPolicyEvaluator struct {
	opaEndpoint string
	client      HTTPClient
}

// HTTPClient interface for dependency injection and testing
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewPolicyEvaluator creates a new PolicyEvaluator that connects to an OPA instance
func NewPolicyEvaluator(opaEndpoint string) PolicyEvaluator {
	if opaEndpoint == "" {
		return nil
	}
	
	return &opaPolicyEvaluator{
		opaEndpoint: opaEndpoint,
		client:      &http.Client{Timeout: DefaultTimeout},
	}
}

// EvaluateAccess evaluates access control policies via OPA
func (e *opaPolicyEvaluator) EvaluateAccess(ctx context.Context, request *PolicyRequest) (*PolicyDecision, error) {
	// Build OPA query input
	queryInput := map[string]interface{}{
		"user": map[string]interface{}{
			"user_id": request.UserID,
			"roles":   request.Roles,
		},
		"request": map[string]interface{}{
			"resource":   request.Resource,
			"action":     request.Action,
			"gateway":    request.Gateway,
			"client_ip":  request.ClientIP,
			"headers":    request.Headers,
		},
	}

	// Determine which policy package to query based on gateway
	policyPath, err := e.getPolicyPath(request.Gateway, "access")
	if err != nil {
		return &PolicyDecision{
			Allow:  false,
			Reason: ReasonUnknownGateway,
		}, err
	}

	// Query OPA for access decision
	opaResponse, err := e.queryOPA(ctx, policyPath, queryInput)
	if err != nil {
		return &PolicyDecision{
			Allow:  false,
			Reason: ReasonPolicyEvaluationError,
		}, fmt.Errorf("failed to query OPA for access decision: %w", err)
	}

	// Extract decision and reason from OPA response
	allow := e.extractBoolean(opaResponse, "allow")
	reason := e.extractString(opaResponse, "reason")
	
	return &PolicyDecision{
		Allow:    allow,
		Reason:   reason,
		PolicyID: policyPath,
	}, nil
}

// EvaluateRateLimit evaluates rate limiting policies via OPA  
func (e *opaPolicyEvaluator) EvaluateRateLimit(ctx context.Context, request *RateLimitRequest) (*RateLimits, error) {
	// Build OPA query input
	queryInput := map[string]interface{}{
		"user": map[string]interface{}{
			"user_id": request.UserID,
		},
		"request": map[string]interface{}{
			"client_ip": request.ClientIP,
			"gateway":   request.Gateway,
			"endpoint":  request.Endpoint,
		},
	}

	// Determine which rate limit policy to query based on gateway
	policyPath, err := e.getPolicyPath(request.Gateway, "rate_limit")
	if err != nil {
		return &RateLimits{
			RequestsPerWindow: 0,
			TimeWindow:        0,
		}, err
	}

	// Query OPA for rate limit decision
	opaResponse, err := e.queryOPA(ctx, policyPath, queryInput)
	if err != nil {
		return &RateLimits{
			RequestsPerWindow: 0,
			TimeWindow:        0,
		}, fmt.Errorf("failed to query OPA for rate limit decision: %w", err)
	}

	// Extract rate limits from OPA response
	requestsPerWindow := e.extractInteger(opaResponse, "requests_per_window")
	timeWindowSeconds := e.extractInteger(opaResponse, "time_window_seconds")
	
	return &RateLimits{
		RequestsPerWindow: requestsPerWindow,
		TimeWindow:        time.Duration(timeWindowSeconds) * time.Second,
	}, nil
}

// queryOPA sends a query to OPA and returns the response
func (e *opaPolicyEvaluator) queryOPA(ctx context.Context, policyPath string, input interface{}) (map[string]interface{}, error) {
	// Create OPA query request
	queryRequest := map[string]interface{}{
		"input": input,
	}

	requestBody, err := json.Marshal(queryRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query request: %w", err)
	}

	// Build OPA query URL
	queryURL := fmt.Sprintf("%s/v1/data/%s", e.opaEndpoint, policyPath)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", queryURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute OPA query: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OPA query failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse OPA response
	var opaResponse struct {
		Result map[string]interface{} `json:"result"`
	}
	if err := json.Unmarshal(respBody, &opaResponse); err != nil {
		return nil, fmt.Errorf("failed to parse OPA response: %w", err)
	}

	return opaResponse.Result, nil
}

// getPolicyPath returns the appropriate OPA policy path based on gateway and policy type
func (e *opaPolicyEvaluator) getPolicyPath(gateway, policyType string) (string, error) {
	switch gateway {
	case AdminGateway:
		switch policyType {
		case "access":
			return AdminRBACPath, nil
		case "rate_limit":
			return AdminRateLimitPath, nil
		default:
			return "", fmt.Errorf("unknown policy type: %s", policyType)
		}
	case PublicGateway:
		switch policyType {
		case "access":
			return PublicAnonymousPath, nil
		case "rate_limit":
			return PublicRateLimitPath, nil
		default:
			return "", fmt.Errorf("unknown policy type: %s", policyType)
		}
	default:
		return "", fmt.Errorf("unknown gateway: %s", gateway)
	}
}

// extractBoolean safely extracts a boolean value from OPA response
func (e *opaPolicyEvaluator) extractBoolean(response map[string]interface{}, key string) bool {
	if value, exists := response[key]; exists {
		if boolValue, ok := value.(bool); ok {
			return boolValue
		}
	}
	return false
}

// extractString safely extracts a string value from OPA response
func (e *opaPolicyEvaluator) extractString(response map[string]interface{}, key string) string {
	if value, exists := response[key]; exists {
		if stringValue, ok := value.(string); ok {
			return stringValue
		}
	}
	return ""
}

// extractInteger safely extracts an integer value from OPA response
func (e *opaPolicyEvaluator) extractInteger(response map[string]interface{}, key string) int {
	if value, exists := response[key]; exists {
		if floatValue, ok := value.(float64); ok {
			return int(floatValue)
		}
		if intValue, ok := value.(int); ok {
			return intValue
		}
	}
	return 0
}

// testPolicyEvaluator provides an in-memory policy evaluator for unit testing
type testPolicyEvaluator struct{}

// NewTestPolicyEvaluator creates a policy evaluator that simulates OPA responses for testing
func NewTestPolicyEvaluator() PolicyEvaluator {
	return &testPolicyEvaluator{}
}

// EvaluateAccess implements the policy logic in-memory for unit testing
func (t *testPolicyEvaluator) EvaluateAccess(ctx context.Context, request *PolicyRequest) (*PolicyDecision, error) {
	// Admin gateway policies
	if request.Gateway == AdminGateway {
		// Admin users can access admin endpoints
		if hasRole(request.Roles, RoleAdmin) && strings.HasPrefix(request.Resource, "/admin/") {
			return &PolicyDecision{
				Allow:  true,
				Reason: ReasonAdminRolePermits,
			}, nil
		}
		
		// Regular users cannot access admin endpoints
		if hasRole(request.Roles, RoleUser) && strings.HasPrefix(request.Resource, "/admin/") {
			return &PolicyDecision{
				Allow:  false,
				Reason: ReasonInsufficientPermissions,
			}, nil
		}
		
		// Unauthenticated users cannot access admin endpoints
		if request.UserID == "" && strings.HasPrefix(request.Resource, "/admin/") {
			return &PolicyDecision{
				Allow:  false,
				Reason: ReasonAuthenticationRequired,
			}, nil
		}
	}
	
	// Public gateway policies
	if request.Gateway == PublicGateway {
		// Anonymous access to public endpoints
		if strings.HasPrefix(request.Resource, "/api/v1/public/") {
			return &PolicyDecision{
				Allow:  true,
				Reason: ReasonPublicEndpoint,
			}, nil
		}
		
		// Protected endpoints require authentication
		if request.UserID == "" && (strings.Contains(request.Resource, "/patients") || 
										strings.Contains(request.Resource, "/appointments")) {
			return &PolicyDecision{
				Allow:  false,
				Reason: ReasonAuthenticationRequired,
			}, nil
		}
	}
	
	// Default deny
	return &PolicyDecision{
		Allow:  false,
		Reason: ReasonNoMatchingPolicy,
	}, nil
}

// EvaluateRateLimit implements rate limit logic in-memory for unit testing
func (t *testPolicyEvaluator) EvaluateRateLimit(ctx context.Context, request *RateLimitRequest) (*RateLimits, error) {
	switch request.Gateway {
	case AdminGateway:
		return &RateLimits{
			RequestsPerWindow: DefaultAdminRateLimit,
			TimeWindow:        DefaultTimeWindow * time.Second,
		}, nil
	case PublicGateway:
		return &RateLimits{
			RequestsPerWindow: DefaultPublicRateLimit,
			TimeWindow:        DefaultTimeWindow * time.Second,
		}, nil
	default:
		return &RateLimits{
			RequestsPerWindow: 0,
			TimeWindow:        0,
		}, fmt.Errorf("unknown gateway: %s", request.Gateway)
	}
}

// hasRole checks if a role exists in the roles slice
func hasRole(roles []string, targetRole string) bool {
	for _, role := range roles {
		if role == targetRole {
			return true
		}
	}
	return false
}