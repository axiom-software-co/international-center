package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/gorilla/mux"
)

// ContractValidator handles OpenAPI contract validation for requests and responses
type ContractValidator struct {
	adminSpec  *openapi3.T
	publicSpec *openapi3.T
	adminRouter routers.Router
	publicRouter routers.Router
}

// NewContractValidator creates a new contract validator with OpenAPI specifications
func NewContractValidator(adminSpecPath, publicSpecPath string) (*ContractValidator, error) {
	ctx := context.Background()
	
	// Load admin API specification
	adminLoader := &openapi3.Loader{Context: ctx, IsExternalRefsAllowed: true}
	adminSpec, err := adminLoader.LoadFromFile(adminSpecPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load admin API spec: %w", err)
	}
	
	// Load public API specification
	publicLoader := &openapi3.Loader{Context: ctx, IsExternalRefsAllowed: true}
	publicSpec, err := publicLoader.LoadFromFile(publicSpecPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load public API spec: %w", err)
	}
	
	// Validate specifications
	if err := adminSpec.Validate(ctx); err != nil {
		return nil, fmt.Errorf("invalid admin API spec: %w", err)
	}
	
	if err := publicSpec.Validate(ctx); err != nil {
		return nil, fmt.Errorf("invalid public API spec: %w", err)
	}
	
	// Create routers for route matching
	adminRouter, err := gorillamux.NewRouter(adminSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin router: %w", err)
	}
	
	publicRouter, err := gorillamux.NewRouter(publicSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to create public router: %w", err)
	}
	
	return &ContractValidator{
		adminSpec:    adminSpec,
		publicSpec:   publicSpec,
		adminRouter:  adminRouter,
		publicRouter: publicRouter,
	}, nil
}

// ValidateRequest validates an incoming HTTP request against the OpenAPI specification
func (cv *ContractValidator) ValidateRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Determine API type from path
		isAdminAPI := strings.HasPrefix(r.URL.Path, "/admin/api/")
		
		var router routers.Router
		if isAdminAPI {
			router = cv.adminRouter
		} else {
			router = cv.publicRouter
		}
		
		// Find matching route
		route, pathParams, err := router.FindRoute(r)
		if err != nil {
			// Route not found in spec - let it pass through (might be handled by other middleware)
			next.ServeHTTP(w, r)
			return
		}
		
		// Read request body for validation (we'll need to restore it)
		var requestBody []byte
		if r.Body != nil {
			requestBody, err = io.ReadAll(r.Body)
			if err != nil {
				cv.handleValidationError(w, r, fmt.Errorf("failed to read request body: %w", err))
				return
			}
			// Restore body for downstream handlers
			r.Body = io.NopCloser(bytes.NewReader(requestBody))
		}
		
		// Create validation input
		requestValidationInput := &openapi3filter.RequestValidationInput{
			Request:    r,
			PathParams: pathParams,
			Route:      route,
		}
		
		// Validate request
		ctx := r.Context()
		if err := openapi3filter.ValidateRequest(ctx, requestValidationInput); err != nil {
			cv.handleValidationError(w, r, fmt.Errorf("request validation failed: %w", err))
			return
		}
		
		// Store validation context for response validation
		ctx = context.WithValue(ctx, "contract_route", route)
		ctx = context.WithValue(ctx, "contract_path_params", pathParams)
		ctx = context.WithValue(ctx, "contract_is_admin", isAdminAPI)
		
		// Create response writer wrapper for response validation
		wrappedWriter := &responseValidatorWriter{
			ResponseWriter: w,
			route:         route,
			request:       r,
			validator:     cv,
		}
		
		// Continue to next handler with wrapped writer and updated context
		next.ServeHTTP(wrappedWriter, r.WithContext(ctx))
	})
}

// responseValidatorWriter wraps http.ResponseWriter to validate responses
type responseValidatorWriter struct {
	http.ResponseWriter
	route     *routers.Route
	request   *http.Request
	validator *ContractValidator
	
	statusCode int
	body       bytes.Buffer
	written    bool
}

// Write captures response body for validation
func (w *responseValidatorWriter) Write(data []byte) (int, error) {
	if !w.written {
		// Default status code if not explicitly set
		if w.statusCode == 0 {
			w.statusCode = http.StatusOK
		}
		w.written = true
	}
	
	// Buffer response body for validation
	w.body.Write(data)
	
	// Write to actual response
	return w.ResponseWriter.Write(data)
}

// WriteHeader captures status code and validates response
func (w *responseValidatorWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	
	// Validate response before writing headers
	if err := w.validateResponse(); err != nil {
		// Log validation error but don't fail the response
		// In production, you might want to handle this differently
		fmt.Printf("Response validation error: %v\n", err)
	}
	
	w.ResponseWriter.WriteHeader(statusCode)
}

// validateResponse validates the response against the OpenAPI specification
func (w *responseValidatorWriter) validateResponse() error {
	// Skip validation for certain status codes that might not have schema definitions
	if w.statusCode == http.StatusNoContent || w.statusCode == http.StatusNotModified {
		return nil
	}
	
	// Get content type
	contentType := w.Header().Get("Content-Type")
	if contentType == "" {
		contentType = "application/json" // Default assumption
	}
	
	// Create response for validation
	response := &http.Response{
		StatusCode: w.statusCode,
		Header:     w.Header(),
		Body:       io.NopCloser(bytes.NewReader(w.body.Bytes())),
	}
	
	// Create validation input
	responseValidationInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request: w.request,
			Route:   w.route,
		},
		Status: w.statusCode,
		Header: response.Header,
	}
	
	// Only validate body if there's content
	if w.body.Len() > 0 {
		responseValidationInput.SetBodyBytes(w.body.Bytes())
	}
	
	// Validate response
	ctx := context.Background()
	if err := openapi3filter.ValidateResponse(ctx, responseValidationInput); err != nil {
		return fmt.Errorf("response validation failed: %w", err)
	}
	
	return nil
}

// handleValidationError handles contract validation errors
func (cv *ContractValidator) handleValidationError(w http.ResponseWriter, r *http.Request, err error) {
	// Extract correlation ID from context if available
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
			"message":        "Contract validation failed",
			"correlation_id": correlationID,
			"timestamp":      fmt.Sprintf("%v", time.Now().UTC().Format(time.RFC3339)),
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

// ValidationConfig holds configuration for contract validation middleware
type ValidationConfig struct {
	AdminSpecPath  string
	PublicSpecPath string
	EnableLogging  bool
	StrictMode     bool // If true, fails on response validation errors
}

// NewContractValidationMiddleware creates a new contract validation middleware
func NewContractValidationMiddleware(config ValidationConfig) (mux.MiddlewareFunc, error) {
	validator, err := NewContractValidator(config.AdminSpecPath, config.PublicSpecPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract validator: %w", err)
	}
	
	return validator.ValidateRequest, nil
}

// ValidationStats tracks validation metrics
type ValidationStats struct {
	RequestsValidated   int64
	RequestsValid       int64
	RequestsInvalid     int64
	ResponsesValidated  int64
	ResponsesValid      int64
	ResponsesInvalid    int64
}

// GetStats returns current validation statistics
func (cv *ContractValidator) GetStats() ValidationStats {
	// In a real implementation, this would use atomic counters
	// For now, return empty stats
	return ValidationStats{}
}