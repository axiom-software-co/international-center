package gateway

import (
	"encoding/json"
	"net/http"
	"strings"
	
	"github.com/google/uuid"
)

// DAPRMiddlewareSimulator simulates DAPR middleware behavior for testing
type DAPRMiddlewareSimulator struct {
	handler   http.Handler
	isAdmin   bool
}

// NewDAPRMiddlewareSimulator creates a new DAPR middleware simulator
func NewDAPRMiddlewareSimulator(handler http.Handler, isAdmin bool) http.Handler {
	return &DAPRMiddlewareSimulator{
		handler: handler,
		isAdmin: isAdmin,
	}
}

// ServeHTTP implements http.Handler interface
func (d *DAPRMiddlewareSimulator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Skip DAPR authentication for health, readiness, metrics, and gateway info endpoints
	if d.isSystemEndpoint(r.URL.Path) {
		d.handler.ServeHTTP(w, r)
		return
	}
	
	// Handle admin routes based on gateway type
	if strings.HasPrefix(r.URL.Path, "/admin/") {
		// For public gateways, admin routes don't exist - let underlying handler return 404
		if !d.isAdmin {
			d.handler.ServeHTTP(w, r)
			return
		}
		// For admin gateways, apply DAPR authentication middleware
	} else {
		// Non-admin routes get CORS headers and pass through
		d.addCORSHeaders(w, r)
		d.handler.ServeHTTP(w, r)
		return
	}
	
	// Simulate DAPR bearer authentication middleware for admin routes on admin gateways
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		d.writeAuthError(w, r, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "Authentication is required")
		return
	}
	
	if !strings.HasPrefix(authHeader, "Bearer ") {
		d.writeAuthError(w, r, http.StatusUnauthorized, "INVALID_TOKEN_FORMAT", "Invalid authentication token format")
		return
	}
	
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		d.writeAuthError(w, r, http.StatusUnauthorized, "EMPTY_TOKEN", "Authentication token cannot be empty")
		return
	}
	
	// Simulate DAPR token validation - these match the test tokens from auth_middleware_contract_test.go
	switch token {
	case "valid_admin_token":
		// Allow admin access to all endpoints
	case "valid_editor_token":
		// Allow editor access to content creation endpoints
	case "valid_viewer_token":
		// Allow viewer access but restrict certain operations
		if r.Method == "DELETE" {
			d.writeAuthError(w, r, http.StatusForbidden, "INSUFFICIENT_PERMISSIONS", "Insufficient permissions for this operation")
			return
		}
	case "expired_admin_token":
		d.writeAuthError(w, r, http.StatusUnauthorized, "TOKEN_EXPIRED", "Authentication token has expired")
		return
	default:
		d.writeAuthError(w, r, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid authentication token")
		return
	}
	
	// Set user context headers that would be added by DAPR
	switch token {
	case "valid_admin_token":
		r.Header.Set("X-User-ID", "admin-user-123")
		r.Header.Set("X-User-Role", "admin")
	case "valid_editor_token":
		r.Header.Set("X-User-ID", "editor-user-456")
		r.Header.Set("X-User-Role", "editor")
	case "valid_viewer_token":
		r.Header.Set("X-User-ID", "viewer-user-789")
		r.Header.Set("X-User-Role", "viewer")
	}
	
	// Pass to the actual gateway handler
	d.handler.ServeHTTP(w, r)
}

// isSystemEndpoint checks if the path is a system endpoint that shouldn't require auth
func (d *DAPRMiddlewareSimulator) isSystemEndpoint(path string) bool {
	systemPaths := []string{
		"/health",
		"/ready", 
		"/metrics",
		"/gateway/info",
	}
	
	for _, systemPath := range systemPaths {
		if path == systemPath {
			return true
		}
	}
	
	return false
}

// writeAuthError writes a standardized authentication error response
func (d *DAPRMiddlewareSimulator) writeAuthError(w http.ResponseWriter, r *http.Request, statusCode int, errorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	
	// Add security headers that would normally be set by gateway middleware
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; object-src 'none'")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	
	// Add gateway identification headers
	w.Header().Set("X-Gateway", "admin-gateway")
	w.Header().Set("X-Gateway-Version", "1.0.0")
	
	// Add correlation ID header - extract from request context or generate if missing
	var correlationID string
	if existingID := r.Header.Get("X-Correlation-ID"); existingID != "" {
		correlationID = existingID
	} else if contextID := r.Context().Value("correlationID"); contextID != nil {
		correlationID = contextID.(string)
	} else {
		// Generate new correlation ID if none exists (required for audit trail)
		correlationID = uuid.New().String()
	}
	w.Header().Set("X-Correlation-ID", correlationID)
	
	// Add WWW-Authenticate header for 401 responses (DAPR middleware behavior)
	if statusCode == http.StatusUnauthorized {
		w.Header().Set("WWW-Authenticate", "Bearer realm=\"dapr\"")
	}
	
	w.WriteHeader(statusCode)
	
	response := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    errorCode,
			"message": message,
		},
	}
	
	json.NewEncoder(w).Encode(response)
}

// addCORSHeaders adds CORS headers as would be done by DAPR CORS middleware
func (d *DAPRMiddlewareSimulator) addCORSHeaders(w http.ResponseWriter, r *http.Request) {
	// Basic CORS headers that would be set by DAPR CORS middleware
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	w.Header().Set("Access-Control-Max-Age", "86400")
}

// GetHandler returns the underlying handler for testing and debugging purposes
func (d *DAPRMiddlewareSimulator) GetHandler() http.Handler {
	return d.handler
}