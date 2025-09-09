package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/google/uuid"
)

// RequestValidator defines the interface for request validation
type RequestValidator func(body []byte) error

// ResponseValidator defines the interface for response validation  
type ResponseValidator func(statusCode int, body []byte) error

// ValidationRule represents a validation rule for a specific endpoint
type ValidationRule struct {
	Method            string
	Path              string
	RequestValidator  RequestValidator
	ResponseValidator ResponseValidator
}

// LightweightValidator provides efficient runtime validation using generated types
type LightweightValidator struct {
	rules map[string]ValidationRule
}

// NewLightweightValidator creates a new lightweight validator
func NewLightweightValidator() *LightweightValidator {
	validator := &LightweightValidator{
		rules: make(map[string]ValidationRule),
	}
	
	// Register validation rules for inquiry endpoints
	validator.registerInquiryRules()
	
	return validator
}

// registerInquiryRules registers validation rules for inquiry management endpoints
func (v *LightweightValidator) registerInquiryRules() {
	// GET /admin/api/v1/inquiries
	v.AddRule("GET", "/admin/api/v1/inquiries", nil, v.validateInquiryListResponse)
	
	// GET /admin/api/v1/inquiries/{id}
	v.AddRule("GET", "/admin/api/v1/inquiries/", nil, v.validateInquiryResponse)
	
	// PUT /admin/api/v1/inquiries/{id}
	v.AddRule("PUT", "/admin/api/v1/inquiries/", v.validateUpdateInquiryRequest, v.validateInquiryUpdateResponse)
}

// AddRule adds a validation rule for a specific endpoint
func (v *LightweightValidator) AddRule(method, path string, reqValidator RequestValidator, respValidator ResponseValidator) {
	key := v.ruleKey(method, path)
	v.rules[key] = ValidationRule{
		Method:            method,
		Path:              path,
		RequestValidator:  reqValidator,
		ResponseValidator: respValidator,
	}
}

// ruleKey generates a key for storing validation rules
func (v *LightweightValidator) ruleKey(method, path string) string {
	return fmt.Sprintf("%s:%s", method, path)
}

// findRule finds a validation rule for the given request
func (v *LightweightValidator) findRule(method, path string) *ValidationRule {
	// Try exact match first
	if rule, exists := v.rules[v.ruleKey(method, path)]; exists {
		return &rule
	}
	
	// Try pattern matching for paths with parameters
	for key, rule := range v.rules {
		if strings.HasPrefix(key, method+":") {
			rulePath := strings.TrimPrefix(key, method+":")
			if v.pathMatches(rulePath, path) {
				return &rule
			}
		}
	}
	
	return nil
}

// pathMatches checks if a path matches a rule pattern
func (v *LightweightValidator) pathMatches(pattern, path string) bool {
	// Simple pattern matching - in production, use a more robust approach
	if strings.HasSuffix(pattern, "/") {
		return strings.HasPrefix(path, pattern)
	}
	return pattern == path
}

// ValidateRequest validates a request against contract rules
func (v *LightweightValidator) ValidateRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Find validation rule
		rule := v.findRule(r.Method, r.URL.Path)
		if rule == nil || rule.RequestValidator == nil {
			// No validation rule - pass through
			next.ServeHTTP(w, r)
			return
		}
		
		// Read and validate request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			v.writeValidationError(w, r, fmt.Errorf("failed to read request body: %w", err))
			return
		}
		
		// Restore request body for downstream handlers
		r.Body = io.NopCloser(bytes.NewReader(body))
		
		// Validate request
		if err := rule.RequestValidator(body); err != nil {
			v.writeValidationError(w, r, fmt.Errorf("request validation failed: %w", err))
			return
		}
		
		// Create response validator wrapper
		wrapper := &validationResponseWriter{
			ResponseWriter: w,
			request:       r,
			validator:     v,
			rule:          rule,
		}
		
		next.ServeHTTP(wrapper, r)
	})
}

// validationResponseWriter wraps ResponseWriter for response validation
type validationResponseWriter struct {
	http.ResponseWriter
	request    *http.Request
	validator  *LightweightValidator
	rule       *ValidationRule
	statusCode int
	body       bytes.Buffer
	written    bool
}

// Write captures response body
func (w *validationResponseWriter) Write(data []byte) (int, error) {
	if !w.written {
		if w.statusCode == 0 {
			w.statusCode = http.StatusOK
		}
		w.written = true
	}
	
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

// WriteHeader validates response and writes status
func (w *validationResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	
	// Validate response if validator exists
	if w.rule.ResponseValidator != nil {
		if err := w.rule.ResponseValidator(statusCode, w.body.Bytes()); err != nil {
			// Log validation error - in production, handle according to policy
			fmt.Printf("Response validation failed for %s %s: %v\n", 
				w.request.Method, w.request.URL.Path, err)
		}
	}
	
	w.ResponseWriter.WriteHeader(statusCode)
}

// Validation functions for specific inquiry endpoints

// validateUpdateInquiryRequest validates inquiry status update requests
func (v *LightweightValidator) validateUpdateInquiryRequest(body []byte) error {
	if len(body) == 0 {
		return fmt.Errorf("request body is required")
	}
	
	var request struct {
		Status     string  `json:"status"`
		Notes      *string `json:"notes,omitempty"`
		AssignedTo *string `json:"assigned_to,omitempty"`
	}
	
	if err := json.Unmarshal(body, &request); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	
	// Validate status field
	validStatuses := []string{"pending", "in_progress", "completed", "closed"}
	if !v.contains(validStatuses, request.Status) {
		return fmt.Errorf("invalid status: %s, must be one of %v", request.Status, validStatuses)
	}
	
	// Validate UUID format for assigned_to if present
	if request.AssignedTo != nil && *request.AssignedTo != "" {
		if _, err := uuid.Parse(*request.AssignedTo); err != nil {
			return fmt.Errorf("invalid assigned_to UUID format: %w", err)
		}
	}
	
	return nil
}

// validateInquiryListResponse validates inquiry list responses
func (v *LightweightValidator) validateInquiryListResponse(statusCode int, body []byte) error {
	if statusCode != http.StatusOK {
		return v.validateErrorResponse(statusCode, body)
	}
	
	if len(body) == 0 {
		return fmt.Errorf("response body is required for successful inquiry list")
	}
	
	var response struct {
		Data       []map[string]interface{} `json:"data"`
		Pagination map[string]interface{}   `json:"pagination"`
	}
	
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("invalid JSON response: %w", err)
	}
	
	// Validate data structure
	if response.Data == nil {
		return fmt.Errorf("data field is required")
	}
	
	// Validate pagination structure
	if response.Pagination == nil {
		return fmt.Errorf("pagination field is required")
	}
	
	// Validate pagination fields
	requiredPagFields := []string{"current_page", "total_pages", "total_items", "items_per_page", "has_next", "has_previous"}
	for _, field := range requiredPagFields {
		if _, exists := response.Pagination[field]; !exists {
			return fmt.Errorf("pagination missing required field: %s", field)
		}
	}
	
	return nil
}

// validateInquiryResponse validates single inquiry responses
func (v *LightweightValidator) validateInquiryResponse(statusCode int, body []byte) error {
	if statusCode != http.StatusOK {
		return v.validateErrorResponse(statusCode, body)
	}
	
	if len(body) == 0 {
		return fmt.Errorf("response body is required for successful inquiry response")
	}
	
	var response struct {
		Data map[string]interface{} `json:"data"`
	}
	
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("invalid JSON response: %w", err)
	}
	
	// Validate inquiry data structure
	if response.Data == nil {
		return fmt.Errorf("data field is required")
	}
	
	// Validate required inquiry fields
	requiredFields := []string{"inquiry_id", "inquiry_type", "status", "submitter_name", "submitter_email", "subject", "message", "submitted_on"}
	for _, field := range requiredFields {
		if _, exists := response.Data[field]; !exists {
			return fmt.Errorf("inquiry data missing required field: %s", field)
		}
	}
	
	return nil
}

// validateInquiryUpdateResponse validates inquiry update responses
func (v *LightweightValidator) validateInquiryUpdateResponse(statusCode int, body []byte) error {
	if statusCode != http.StatusOK {
		return v.validateErrorResponse(statusCode, body)
	}
	
	if len(body) == 0 {
		return fmt.Errorf("response body is required for successful inquiry update")
	}
	
	var response struct {
		Success       bool                   `json:"success"`
		Message       string                 `json:"message"`
		Data          map[string]interface{} `json:"data"`
		Timestamp     string                 `json:"timestamp"`
		CorrelationId string                 `json:"correlation_id"`
	}
	
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("invalid JSON response: %w", err)
	}
	
	// Validate required fields
	if !response.Success {
		return fmt.Errorf("success field must be true for successful update")
	}
	
	if response.Message == "" {
		return fmt.Errorf("message field is required")
	}
	
	if response.Data == nil {
		return fmt.Errorf("data field is required")
	}
	
	// Validate timestamp format
	if _, err := time.Parse(time.RFC3339, response.Timestamp); err != nil {
		return fmt.Errorf("invalid timestamp format: %w", err)
	}
	
	// Validate correlation ID format
	if _, err := uuid.Parse(response.CorrelationId); err != nil {
		return fmt.Errorf("invalid correlation_id UUID format: %w", err)
	}
	
	return nil
}

// validateErrorResponse validates error responses
func (v *LightweightValidator) validateErrorResponse(statusCode int, body []byte) error {
	// Error responses should have body
	if len(body) == 0 && statusCode >= 400 {
		return fmt.Errorf("error responses should include body")
	}
	
	if len(body) > 0 {
		var errorResp struct {
			Error map[string]interface{} `json:"error"`
		}
		
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return fmt.Errorf("invalid error response JSON: %w", err)
		}
		
		// Validate error structure
		if errorResp.Error == nil {
			return fmt.Errorf("error field is required in error responses")
		}
		
		// Validate required error fields
		requiredFields := []string{"code", "message", "correlation_id", "timestamp"}
		for _, field := range requiredFields {
			if _, exists := errorResp.Error[field]; !exists {
				return fmt.Errorf("error missing required field: %s", field)
			}
		}
	}
	
	return nil
}

// Helper functions

// contains checks if a slice contains a value
func (v *LightweightValidator) contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// writeValidationError writes a validation error response
func (v *LightweightValidator) writeValidationError(w http.ResponseWriter, r *http.Request, err error) {
	// Extract correlation ID
	correlationID := "unknown"
	if ctx := r.Context(); ctx != nil {
		if corrCtx := domain.FromContext(ctx); corrCtx != nil {
			correlationID = corrCtx.CorrelationID
		}
	}
	
	// Create standardized error response
	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":           "CONTRACT_VALIDATION_FAILED",
			"message":        "Request validation failed",
			"correlation_id": correlationID,
			"timestamp":      time.Now().UTC().Format(time.RFC3339),
			"details": map[string]interface{}{
				"validation_error": err.Error(),
				"path":            r.URL.Path,
				"method":          r.Method,
			},
		},
	}
	
	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Correlation-ID", correlationID)
	w.WriteHeader(http.StatusBadRequest)
	
	// Write error response
	json.NewEncoder(w).Encode(errorResponse)
}

// NewLightweightValidationMiddleware creates lightweight contract validation middleware
func NewLightweightValidationMiddleware() *LightweightValidator {
	return NewLightweightValidator()
}