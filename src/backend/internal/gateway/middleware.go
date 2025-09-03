package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"golang.org/x/time/rate"
)

// Middleware represents gateway middleware
type Middleware struct {
	config        *GatewayConfiguration
	rateLimiters  map[string]*rate.Limiter
}

// NewMiddleware creates a new middleware instance
func NewMiddleware(config *GatewayConfiguration) *Middleware {
	return &Middleware{
		config:       config,
		rateLimiters: make(map[string]*rate.Limiter),
	}
}

// ApplyMiddleware applies all configured middleware to the handler
func (m *Middleware) ApplyMiddleware(handler http.Handler) http.Handler {
	// Apply middleware in reverse order (outermost first)
	
	// Apply observability middleware (tracing, logging)
	if m.config.Observability.Enabled {
		handler = m.observabilityMiddleware(handler)
	}
	
	// Apply security headers middleware
	if m.config.Security.SecurityHeaders.Enabled {
		handler = m.securityHeadersMiddleware(handler)
	}
	
	// Apply CORS middleware
	if m.config.CORS.Enabled {
		handler = m.corsMiddleware(handler)
	}
	
	// Apply rate limiting middleware
	if m.config.RateLimit.Enabled {
		handler = m.rateLimitMiddleware(handler)
	}
	
	// Apply authentication middleware (admin gateway only)
	if m.config.ShouldRequireAuth() {
		handler = m.authenticationMiddleware(handler)
	}
	
	// Apply correlation context middleware (always first)
	handler = m.correlationContextMiddleware(handler)
	
	return handler
}

// correlationContextMiddleware ensures all requests have correlation context
func (m *Middleware) correlationContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		
		// Create or extract correlation context
		correlationCtx := domain.FromContext(ctx)
		if correlationCtx.CorrelationID == "" {
			correlationCtx.CorrelationID = domain.GenerateCorrelationID()
		}
		
		// Set trace ID if present in headers
		if traceID := r.Header.Get("X-Trace-ID"); traceID != "" {
			correlationCtx.TraceID = traceID
		}
		
		// Update context
		ctx = correlationCtx.ToContext(ctx)
		
		// Add correlation ID to response headers
		w.Header().Set("X-Correlation-ID", correlationCtx.CorrelationID)
		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// authenticationMiddleware validates authentication for admin gateway
func (m *Middleware) authenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health checks and metrics
		if m.isHealthOrMetricsPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		
		// Extract authentication token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.writeErrorResponse(w, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "Authentication is required")
			return
		}
		
		// Validate token format (Bearer token)
		if !strings.HasPrefix(authHeader, "Bearer ") {
			m.writeErrorResponse(w, http.StatusUnauthorized, "INVALID_TOKEN_FORMAT", "Invalid authentication token format")
			return
		}
		
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			m.writeErrorResponse(w, http.StatusUnauthorized, "EMPTY_TOKEN", "Authentication token cannot be empty")
			return
		}
		
		// In a real implementation, this would validate the token with Authentik
		// For now, just check for a test user ID header
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			m.writeErrorResponse(w, http.StatusUnauthorized, "USER_ID_REQUIRED", "User ID is required for authenticated requests")
			return
		}
		
		// Token validation would happen here
		// For testing purposes, accept any non-empty token
		
		next.ServeHTTP(w, r)
	})
}

// rateLimitMiddleware implements rate limiting
func (m *Middleware) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip rate limiting for health checks and metrics
		if m.isHealthOrMetricsPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		
		// Determine rate limiting key
		var key string
		switch m.config.RateLimit.KeyExtractor {
		case "user":
			if userID := r.Header.Get("X-User-ID"); userID != "" {
				key = "user:" + userID
			} else {
				key = "ip:" + m.extractClientIP(r)
			}
		case "ip":
			key = "ip:" + m.extractClientIP(r)
		default:
			key = "default"
		}
		
		// Get or create rate limiter for key
		limiter := m.getRateLimiter(key)
		
		// Check if request is allowed
		if !limiter.Allow() {
			m.writeErrorResponse(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Request rate limit exceeded")
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// corsMiddleware implements CORS handling
func (m *Middleware) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		
		// Set CORS headers if origin is allowed
		if m.isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", fmt.Sprintf("%t", m.config.CORS.AllowCredentials))
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(m.config.CORS.AllowedMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(m.config.CORS.AllowedHeaders, ", "))
			w.Header().Set("Access-Control-Expose-Headers", strings.Join(m.config.CORS.ExposedHeaders, ", "))
			w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", m.config.CORS.MaxAge))
		}
		
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// securityHeadersMiddleware adds security headers
func (m *Middleware) securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := m.config.Security.SecurityHeaders
		
		if headers.ContentTypeOptions != "" {
			w.Header().Set("X-Content-Type-Options", headers.ContentTypeOptions)
		}
		if headers.FrameOptions != "" {
			w.Header().Set("X-Frame-Options", headers.FrameOptions)
		}
		if headers.XSSProtection != "" {
			w.Header().Set("X-XSS-Protection", headers.XSSProtection)
		}
		if headers.StrictTransportSecurity != "" {
			w.Header().Set("Strict-Transport-Security", headers.StrictTransportSecurity)
		}
		if headers.ContentSecurityPolicy != "" {
			w.Header().Set("Content-Security-Policy", headers.ContentSecurityPolicy)
		}
		if headers.ReferrerPolicy != "" {
			w.Header().Set("Referrer-Policy", headers.ReferrerPolicy)
		}
		
		// Add gateway identification header
		w.Header().Set("X-Gateway", m.config.Name)
		w.Header().Set("X-Gateway-Version", m.config.Version)
		
		next.ServeHTTP(w, r)
	})
}

// observabilityMiddleware adds observability features
func (m *Middleware) observabilityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create response writer wrapper to capture status
		ww := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}
		
		// Add tracing headers if enabled
		if m.config.Observability.TracingEnabled {
			if traceID := r.Header.Get("X-Trace-ID"); traceID == "" {
				w.Header().Set("X-Trace-ID", domain.GenerateTraceID())
			}
		}
		
		// Process request
		next.ServeHTTP(ww, r)
		
		// Log request if enabled
		if m.config.Observability.LoggingEnabled {
			duration := time.Since(start)
			m.logRequest(r, ww.statusCode, duration)
		}
	})
}

// Helper methods

// getRateLimiter gets or creates a rate limiter for a key
func (m *Middleware) getRateLimiter(key string) *rate.Limiter {
	if limiter, exists := m.rateLimiters[key]; exists {
		return limiter
	}
	
	// Create new rate limiter
	limiter := rate.NewLimiter(
		rate.Limit(m.config.RateLimit.RequestsPerMinute)/60, // requests per second
		m.config.RateLimit.BurstSize,
	)
	
	m.rateLimiters[key] = limiter
	return limiter
}

// extractClientIP extracts the client IP from request
func (m *Middleware) extractClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
		ip = ip[:colonIndex]
	}
	
	return ip
}

// isOriginAllowed checks if origin is allowed for CORS
func (m *Middleware) isOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}
	
	for _, allowedOrigin := range m.config.CORS.AllowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return true
		}
		
		// Support wildcard subdomains
		if strings.Contains(allowedOrigin, "*") {
			pattern := strings.Replace(allowedOrigin, "*", ".*", -1)
			// Simple pattern matching - in production would use regex
			if strings.Contains(origin, strings.Replace(pattern, ".*", "", -1)) {
				return true
			}
		}
	}
	
	return false
}

// isHealthOrMetricsPath checks if path is a health or metrics endpoint
func (m *Middleware) isHealthOrMetricsPath(path string) bool {
	healthPaths := []string{
		m.config.Observability.HealthCheckPath,
		m.config.Observability.ReadinessPath,
		m.config.Observability.MetricsPath,
		m.config.ServiceRouting.HealthCheckPath,
		m.config.ServiceRouting.MetricsPath,
	}
	
	for _, healthPath := range healthPaths {
		if path == healthPath {
			return true
		}
	}
	
	return false
}

// writeErrorResponse writes a standardized error response
func (m *Middleware) writeErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    errorCode,
			"message": message,
		},
	}
	
	// Don't handle encoding errors here - keep it simple
	json.NewEncoder(w).Encode(response)
}

// logRequest logs the request details
func (m *Middleware) logRequest(r *http.Request, statusCode int, duration time.Duration) {
	correlationID := domain.GetCorrelationID(r.Context())
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "anonymous"
	}
	
	// In a real implementation, this would use structured logging
	// For now, we'll keep it simple
	fmt.Printf("[%s] %s %s %d %v - User: %s, Correlation: %s\n",
		time.Now().Format(time.RFC3339),
		r.Method,
		r.URL.Path,
		statusCode,
		duration,
		userID,
		correlationID,
	)
}

// responseWriterWrapper wraps http.ResponseWriter to capture status code
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterWrapper) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}