package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/google/uuid"
)

// Middleware represents gateway middleware
type Middleware struct {
	config        *GatewayConfiguration
}

// NewMiddleware creates a new middleware instance
func NewMiddleware(config *GatewayConfiguration) *Middleware {
	return &Middleware{
		config: config,
	}
}

// ApplyMiddleware applies all configured middleware to the handler
func (m *Middleware) ApplyMiddleware(handler http.Handler) http.Handler {
	// DAPR middleware chain - DAPR sidecar handles authentication, rate limiting, CORS
	// We only apply minimal gateway-specific middleware here
	
	// Apply observability middleware (tracing, logging) - still needed for gateway metrics
	if m.config.Observability.Enabled {
		handler = m.observabilityMiddleware(handler)
	}
	
	// Apply security headers middleware - basic security headers still needed
	if m.config.Security.SecurityHeaders.Enabled {
		handler = m.securityHeadersMiddleware(handler)
	}
	
	// Apply correlation context middleware (always first) - needed for tracing
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
			correlationCtx.CorrelationID = uuid.New().String()
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
				w.Header().Set("X-Trace-ID", domain.GetTraceID(r.Context()))
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
	
	// Add security headers to error responses
	if m.config.Security.SecurityHeaders.Enabled {
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
	}
	
	// Add WWW-Authenticate header for 401 responses
	if statusCode == http.StatusUnauthorized {
		w.Header().Set("WWW-Authenticate", "Bearer realm=\"gateway\"")
	}
	
	// Add gateway identification header
	w.Header().Set("X-Gateway", m.config.Name)
	w.Header().Set("X-Gateway-Version", m.config.Version)
	
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