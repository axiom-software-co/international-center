package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/gorilla/mux"
)

func TestLightweightValidator_ValidateUpdateInquiryRequest(t *testing.T) {
	validator := NewLightweightValidator()

	tests := []struct {
		name    string
		body    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid request",
			body:    `{"status":"in_progress","notes":"Processing inquiry","assigned_to":"550e8400-e29b-41d4-a716-446655440000"}`,
			wantErr: false,
		},
		{
			name:    "valid request without optional fields",
			body:    `{"status":"pending"}`,
			wantErr: false,
		},
		{
			name:    "empty body",
			body:    "",
			wantErr: true,
			errMsg:  "request body is required",
		},
		{
			name:    "invalid JSON",
			body:    `{"status":"pending"`,
			wantErr: true,
			errMsg:  "invalid JSON",
		},
		{
			name:    "invalid status",
			body:    `{"status":"invalid_status"}`,
			wantErr: true,
			errMsg:  "invalid status",
		},
		{
			name:    "invalid UUID format",
			body:    `{"status":"pending","assigned_to":"not-a-uuid"}`,
			wantErr: true,
			errMsg:  "invalid assigned_to UUID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateUpdateInquiryRequest([]byte(tt.body))
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error message to contain %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestLightweightValidator_ValidateInquiryListResponse(t *testing.T) {
	validator := NewLightweightValidator()

	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "valid success response",
			statusCode: 200,
			body: `{
				"data": [
					{
						"inquiry_id": "550e8400-e29b-41d4-a716-446655440000",
						"inquiry_type": "media",
						"status": "pending"
					}
				],
				"pagination": {
					"current_page": 1,
					"total_pages": 1,
					"total_items": 1,
					"items_per_page": 20,
					"has_next": false,
					"has_previous": false
				}
			}`,
			wantErr: false,
		},
		{
			name:       "missing data field",
			statusCode: 200,
			body: `{
				"pagination": {
					"current_page": 1,
					"total_pages": 1,
					"total_items": 1,
					"items_per_page": 20,
					"has_next": false,
					"has_previous": false
				}
			}`,
			wantErr: true,
			errMsg:  "data field is required",
		},
		{
			name:       "missing pagination field",
			statusCode: 200,
			body:       `{"data": []}`,
			wantErr:    true,
			errMsg:     "pagination field is required",
		},
		{
			name:       "incomplete pagination",
			statusCode: 200,
			body: `{
				"data": [],
				"pagination": {
					"current_page": 1
				}
			}`,
			wantErr: true,
			errMsg:  "pagination missing required field",
		},
		{
			name:       "error response",
			statusCode: 400,
			body: `{
				"error": {
					"code": "BAD_REQUEST",
					"message": "Invalid request",
					"correlation_id": "550e8400-e29b-41d4-a716-446655440000",
					"timestamp": "2023-10-01T12:00:00Z"
				}
			}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateInquiryListResponse(tt.statusCode, []byte(tt.body))
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error message to contain %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestLightweightValidator_ValidateInquiryUpdateResponse(t *testing.T) {
	validator := NewLightweightValidator()
	validTimestamp := time.Now().UTC().Format(time.RFC3339)
	validUUID := "550e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "valid update response",
			statusCode: 200,
			body: `{
				"success": true,
				"message": "Inquiry updated successfully",
				"data": {
					"inquiry_id": "550e8400-e29b-41d4-a716-446655440000",
					"status": "in_progress"
				},
				"timestamp": "` + validTimestamp + `",
				"correlation_id": "` + validUUID + `"
			}`,
			wantErr: false,
		},
		{
			name:       "missing success field",
			statusCode: 200,
			body: `{
				"message": "Inquiry updated successfully",
				"data": {},
				"timestamp": "` + validTimestamp + `",
				"correlation_id": "` + validUUID + `"
			}`,
			wantErr: true,
			errMsg:  "success field must be true",
		},
		{
			name:       "invalid timestamp",
			statusCode: 200,
			body: `{
				"success": true,
				"message": "Inquiry updated successfully",
				"data": {},
				"timestamp": "invalid-timestamp",
				"correlation_id": "` + validUUID + `"
			}`,
			wantErr: true,
			errMsg:  "invalid timestamp format",
		},
		{
			name:       "invalid correlation ID",
			statusCode: 200,
			body: `{
				"success": true,
				"message": "Inquiry updated successfully",
				"data": {},
				"timestamp": "` + validTimestamp + `",
				"correlation_id": "not-a-uuid"
			}`,
			wantErr: true,
			errMsg:  "invalid correlation_id UUID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateInquiryUpdateResponse(tt.statusCode, []byte(tt.body))
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error message to contain %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestLightweightValidator_MiddlewareIntegration(t *testing.T) {
	validator := NewLightweightValidator()
	
	// Create test handler that returns a successful response
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add correlation context
		correlationCtx := domain.NewCorrelationContext()
		ctx := correlationCtx.ToContext(r.Context())
		r = r.WithContext(ctx)
		
		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"inquiry_id":      "550e8400-e29b-41d4-a716-446655440000",
					"inquiry_type":    "media",
					"status":          "pending",
					"submitter_name":  "John Doe",
					"submitter_email": "john@example.com",
					"subject":         "Test inquiry",
					"message":         "Test message",
					"submitted_on":    "2023-10-01T12:00:00Z",
				},
			},
			"pagination": map[string]interface{}{
				"current_page":   1,
				"total_pages":    1,
				"total_items":    1,
				"items_per_page": 20,
				"has_next":       false,
				"has_previous":   false,
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})
	
	// Wrap with validation middleware
	wrappedHandler := validator.ValidateRequest(testHandler)
	
	// Create test request
	req := httptest.NewRequest("GET", "/admin/api/v1/inquiries", nil)
	req.Header.Set("Content-Type", "application/json")
	
	// Add correlation context
	correlationCtx := domain.NewCorrelationContext()
	ctx := correlationCtx.ToContext(req.Context())
	req = req.WithContext(ctx)
	
	// Create response recorder
	rr := httptest.NewRecorder()
	
	// Execute request
	wrappedHandler.ServeHTTP(rr, req)
	
	// Check response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	// Verify response is valid JSON
	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}
	
	// Verify response structure
	if _, exists := response["data"]; !exists {
		t.Error("response missing data field")
	}
	
	if _, exists := response["pagination"]; !exists {
		t.Error("response missing pagination field")
	}
}

func TestLightweightValidator_RequestValidationFailure(t *testing.T) {
	validator := NewLightweightValidator()
	
	// Create test handler (should not be called)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called when validation fails")
	})
	
	// Wrap with validation middleware
	wrappedHandler := validator.ValidateRequest(testHandler)
	
	// Create test request with invalid body
	invalidBody := `{"status":"invalid_status"}`
	req := httptest.NewRequest("PUT", "/admin/api/v1/inquiries/550e8400-e29b-41d4-a716-446655440000", 
		bytes.NewBufferString(invalidBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Add correlation context
	correlationCtx := domain.NewCorrelationContext()
	ctx := correlationCtx.ToContext(req.Context())
	req = req.WithContext(ctx)
	
	// Create response recorder
	rr := httptest.NewRecorder()
	
	// Execute request
	wrappedHandler.ServeHTTP(rr, req)
	
	// Check response
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
	
	// Verify error response structure
	var errorResponse map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&errorResponse); err != nil {
		t.Errorf("error response is not valid JSON: %v", err)
	}
	
	if _, exists := errorResponse["error"]; !exists {
		t.Error("error response missing error field")
	}
}

func TestMiddlewareChain(t *testing.T) {
	// Create router
	router := mux.NewRouter()
	
	// Apply middleware chain
	chain := ContractValidationMiddleware()
	chain.Apply(router)
	
	// Add test route
	router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	
	// Execute request
	router.ServeHTTP(rr, req)
	
	// Check that middleware was applied (correlation ID should be present)
	if correlationID := rr.Header().Get("X-Correlation-ID"); correlationID == "" {
		t.Error("expected X-Correlation-ID header to be set by middleware")
	}
	
	// Check security headers
	expectedHeaders := []string{
		"X-Content-Type-Options",
		"X-Frame-Options",
		"X-XSS-Protection",
	}
	
	for _, header := range expectedHeaders {
		if value := rr.Header().Get(header); value == "" {
			t.Errorf("expected security header %s to be set", header)
		}
	}
}