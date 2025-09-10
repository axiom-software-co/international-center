package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
	
	"github.com/google/uuid"
)

// rateLimitState tracks rate limiting for testing
type rateLimitState struct {
	requests    map[string]int // key -> request count
	lastReset   time.Time
	mutex       sync.Mutex
	maxRequests int
}

// DAPRMiddlewareSimulator simulates DAPR middleware behavior for testing
type DAPRMiddlewareSimulator struct {
	handler     http.Handler
	isAdmin     bool
	rateLimit   *rateLimitState
}

// NewDAPRMiddlewareSimulator creates a new DAPR middleware simulator
func NewDAPRMiddlewareSimulator(handler http.Handler, isAdmin bool) http.Handler {
	// Set rate limits based on gateway type
	maxRequests := 1000 // Public gateway: 1000/min
	if isAdmin {
		maxRequests = 100 // Admin gateway: 100/min
	}
	
	return &DAPRMiddlewareSimulator{
		handler: handler,
		isAdmin: isAdmin,
		rateLimit: &rateLimitState{
			requests:    make(map[string]int),
			lastReset:   time.Now(),
			maxRequests: maxRequests,
		},
	}
}

// ServeHTTP implements http.Handler interface
func (d *DAPRMiddlewareSimulator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check rate limiting first (applies to all requests)
	if d.isRateLimited(r) {
		d.writeRateLimitError(w, r)
		return
	}

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
	
	// Simulate DAPR token validation and role-based access control
	var userID, userRole string
	
	switch token {
	case "valid_admin_token":
		userID = "admin-user-123"
		userRole = "admin"
		// Admin access to all endpoints - no restrictions
		
	case "valid_editor_token":
		userID = "editor-user-456"
		userRole = "content_editor"
		// Editor access to content endpoints but not admin-only operations
		if d.isAdminOnlyEndpoint(r.Method, r.URL.Path) {
			d.writeAuthError(w, r, http.StatusForbidden, "INSUFFICIENT_PERMISSIONS", "insufficient permissions for admin-only endpoints")
			return
		}
		
	case "valid_viewer_token":
		userID = "viewer-user-789"
		userRole = "viewer"
		// Viewer access to read-only endpoints only
		if r.Method != "GET" && r.Method != "HEAD" && r.Method != "OPTIONS" {
			d.writeAuthError(w, r, http.StatusForbidden, "INSUFFICIENT_PERMISSIONS", "insufficient permissions for write operations")
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
	r.Header.Set("X-User-ID", userID)
	r.Header.Set("X-User-Role", userRole)
	
	// Create response wrapper to capture and enhance headers
	responseWrapper := &daprResponseWrapper{
		ResponseWriter: w,
		userRole:       userRole,
	}
	
	// Pass to the actual gateway handler with enhanced response writer
	d.handler.ServeHTTP(responseWrapper, r)
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

// isRateLimited checks if the request should be rate limited
func (d *DAPRMiddlewareSimulator) isRateLimited(r *http.Request) bool {
	d.rateLimit.mutex.Lock()
	defer d.rateLimit.mutex.Unlock()
	
	// Reset counts if a minute has passed
	now := time.Now()
	if now.Sub(d.rateLimit.lastReset) >= time.Minute {
		d.rateLimit.requests = make(map[string]int)
		d.rateLimit.lastReset = now
	}
	
	// Determine rate limiting key
	var key string
	if d.isAdmin {
		// User-based rate limiting for admin gateway
		key = r.Header.Get("X-User-ID")
		if key == "" {
			// For unauthenticated requests, use IP
			key = r.RemoteAddr
		}
	} else {
		// IP-based rate limiting for public gateway
		key = r.RemoteAddr
	}
	
	// Check current count
	currentCount := d.rateLimit.requests[key]
	
	// Increment count
	d.rateLimit.requests[key] = currentCount + 1
	
	// Check if limit exceeded
	return d.rateLimit.requests[key] > d.rateLimit.maxRequests
}

// writeRateLimitError writes a rate limit error response
func (d *DAPRMiddlewareSimulator) writeRateLimitError(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Add rate limit headers
	w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", d.rateLimit.maxRequests))
	w.Header().Set("X-RateLimit-Remaining", "0")
	w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))
	
	// Add security headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	
	// Add correlation ID
	correlationID := uuid.New().String()
	w.Header().Set("X-Correlation-ID", correlationID)
	
	w.WriteHeader(http.StatusTooManyRequests)
	
	response := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    "RATE_LIMIT_EXCEEDED",
			"message": "Rate limit exceeded",
		},
	}
	
	json.NewEncoder(w).Encode(response)
}

// isAdminOnlyEndpoint checks if an endpoint requires admin-only access
func (d *DAPRMiddlewareSimulator) isAdminOnlyEndpoint(method, path string) bool {
	// Admin-only operations based on contract specifications
	adminOnlyPatterns := []struct {
		method string
		path   string
	}{
		{"DELETE", "/admin/api/v1/services/"}, // Service deletion
		{"DELETE", "/admin/api/v1/news/"},     // News deletion  
		{"DELETE", "/admin/api/v1/research/"}, // Research deletion
		{"DELETE", "/admin/api/v1/events/"},   // Event deletion
		{"GET", "/admin/api/v1/services/.*audit"}, // Audit endpoints
		{"GET", "/admin/api/v1/news/.*audit"},     // Audit endpoints
		{"GET", "/admin/api/v1/research/.*audit"}, // Audit endpoints
		{"GET", "/admin/api/v1/events/.*audit"},   // Audit endpoints
	}
	
	for _, pattern := range adminOnlyPatterns {
		if method == pattern.method && strings.Contains(path, strings.TrimSuffix(pattern.path, "/")) {
			return true
		}
	}
	
	// Audit endpoints are admin-only
	if strings.Contains(path, "/audit") {
		return true
	}
	
	return false
}

// daprResponseWrapper wraps ResponseWriter to capture and enhance headers
type daprResponseWrapper struct {
	http.ResponseWriter
	userRole string
	written  bool
}

// WriteHeader captures status code and adds user role header
func (r *daprResponseWrapper) WriteHeader(statusCode int) {
	if !r.written {
		// Add user role to response headers for successful requests
		if statusCode >= 200 && statusCode < 300 {
			r.Header().Set("X-User-Role", r.userRole)
		}
		r.written = true
	}
	r.ResponseWriter.WriteHeader(statusCode)
}

// Write ensures headers are written
func (r *daprResponseWrapper) Write(data []byte) (int, error) {
	if !r.written {
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(data)
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