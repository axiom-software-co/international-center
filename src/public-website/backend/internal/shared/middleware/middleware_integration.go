package middleware

import (
	"net/http"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/gorilla/mux"
)

// MiddlewareChain represents a chain of HTTP middleware
type MiddlewareChain struct {
	middlewares []mux.MiddlewareFunc
}

// NewMiddlewareChain creates a new middleware chain
func NewMiddlewareChain() *MiddlewareChain {
	return &MiddlewareChain{
		middlewares: make([]mux.MiddlewareFunc, 0),
	}
}

// Add adds middleware to the chain
func (mc *MiddlewareChain) Add(middleware mux.MiddlewareFunc) *MiddlewareChain {
	mc.middlewares = append(mc.middlewares, middleware)
	return mc
}

// Apply applies all middleware to a router
func (mc *MiddlewareChain) Apply(router *mux.Router) {
	for _, middleware := range mc.middlewares {
		router.Use(middleware)
	}
}

// ContractValidationMiddleware creates the contract validation middleware chain
func ContractValidationMiddleware() *MiddlewareChain {
	chain := NewMiddlewareChain()
	
	// Add correlation ID middleware first
	chain.Add(CorrelationMiddleware())
	
	// Add lightweight contract validation
	validator := NewLightweightValidationMiddleware()
	chain.Add(validator.ValidateRequest)
	
	// Add security headers
	chain.Add(SecurityHeadersMiddleware())
	
	return chain
}

// CorrelationMiddleware adds correlation ID to request context
func CorrelationMiddleware() mux.MiddlewareFunc {
	return mux.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get or create correlation context
			correlationCtx := domain.FromContext(r.Context())
			if correlationCtx == nil {
				correlationCtx = domain.NewCorrelationContext()
			}
			
			// Check for existing correlation ID in headers
			if existingID := r.Header.Get("X-Correlation-ID"); existingID != "" {
				correlationCtx.CorrelationID = existingID
			}
			
			// Add correlation ID to response headers
			w.Header().Set("X-Correlation-ID", correlationCtx.CorrelationID)
			
			// Continue with updated context
			ctx := correlationCtx.ToContext(r.Context())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
}

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware() mux.MiddlewareFunc {
	return mux.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Content-Security-Policy", "default-src 'self'")
			
			next.ServeHTTP(w, r)
		})
	})
}

// RateLimitingMiddleware creates a simple rate limiting middleware
func RateLimitingMiddleware(requestsPerMinute int) mux.MiddlewareFunc {
	// In production, use a proper rate limiting implementation
	// This is a simplified example
	return mux.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: Implement actual rate limiting logic
			// For now, just pass through
			next.ServeHTTP(w, r)
		})
	})
}

// LoggingMiddleware creates structured logging middleware
func LoggingMiddleware() mux.MiddlewareFunc {
	return mux.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get correlation context
			correlationCtx := domain.FromContext(r.Context())
			correlationID := "unknown"
			if correlationCtx != nil {
				correlationID = correlationCtx.CorrelationID
			}
			
			// Log request start
			// In production, use structured logging (slog)
			println("Request started:", r.Method, r.URL.Path, "correlation_id:", correlationID)
			
			// Create response recorder to capture status
			recorder := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
			
			// Process request
			next.ServeHTTP(recorder, r)
			
			// Log request completion
			println("Request completed:", r.Method, r.URL.Path, "status:", recorder.statusCode, "correlation_id:", correlationID)
		})
	})
}

// responseRecorder captures response status for logging
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rr *responseRecorder) WriteHeader(statusCode int) {
	rr.statusCode = statusCode
	rr.ResponseWriter.WriteHeader(statusCode)
}

// AdminAPIMiddleware creates middleware chain for admin API endpoints
func AdminAPIMiddleware() *MiddlewareChain {
	chain := NewMiddlewareChain()
	
	// Add logging first
	chain.Add(LoggingMiddleware())
	
	// Add correlation ID
	chain.Add(CorrelationMiddleware())
	
	// Add rate limiting for admin API (lower limit)
	chain.Add(RateLimitingMiddleware(100)) // 100 requests per minute for admin
	
	// Add contract validation
	validator := NewLightweightValidationMiddleware()
	chain.Add(validator.ValidateRequest)
	
	// Add security headers
	chain.Add(SecurityHeadersMiddleware())
	
	return chain
}

// PublicAPIMiddleware creates middleware chain for public API endpoints
func PublicAPIMiddleware() *MiddlewareChain {
	chain := NewMiddlewareChain()
	
	// Add logging first
	chain.Add(LoggingMiddleware())
	
	// Add correlation ID
	chain.Add(CorrelationMiddleware())
	
	// Add rate limiting for public API (higher limit)
	chain.Add(RateLimitingMiddleware(1000)) // 1000 requests per minute for public
	
	// Add security headers
	chain.Add(SecurityHeadersMiddleware())
	
	return chain
}

// ApplyMiddlewareToInquiryRoutes shows how to apply middleware to inquiry routes
func ApplyMiddlewareToInquiryRoutes(router *mux.Router) {
	// Create admin subrouter with admin middleware
	adminRouter := router.PathPrefix("/admin/api/v1").Subrouter()
	AdminAPIMiddleware().Apply(adminRouter)
	
	// Create public subrouter with public middleware  
	publicRouter := router.PathPrefix("/api/v1").Subrouter()
	PublicAPIMiddleware().Apply(publicRouter)
}

// ValidationConfiguration holds contract validation settings
type ValidationConfiguration struct {
	EnableRequestValidation  bool
	EnableResponseValidation bool
	StrictMode              bool
	LogValidationErrors     bool
}

// DefaultValidationConfig returns default validation configuration
func DefaultValidationConfig() ValidationConfiguration {
	return ValidationConfiguration{
		EnableRequestValidation:  true,
		EnableResponseValidation: true,
		StrictMode:              false, // Don't fail requests on response validation errors
		LogValidationErrors:     true,
	}
}