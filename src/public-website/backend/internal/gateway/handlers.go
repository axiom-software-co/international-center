package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/gorilla/mux"
)

// GatewayHandler handles HTTP requests for the gateway
type GatewayHandler struct {
	config            *GatewayConfiguration
	serviceProxy      *ServiceProxy
	middleware        *Middleware
	auditService      *AuditService
	subscriberHandler *SubscriberHandler
}

// NewGatewayHandler creates a new gateway handler
func NewGatewayHandler(config *GatewayConfiguration, serviceProxy *ServiceProxy, middleware *Middleware) *GatewayHandler {
	handler := &GatewayHandler{
		config:       config,
		serviceProxy: serviceProxy,
		middleware:   middleware,
	}
	
	// Initialize subscriber handler for admin gateways
	if config.IsAdmin() {
		// Note: In a full implementation, database connection would be injected
		// For now, we create a placeholder that will be initialized when needed
		handler.subscriberHandler = nil // Will be set in gateway service initialization
	}
	
	return handler
}

// RegisterRoutes registers gateway routes with the router
func (h *GatewayHandler) RegisterRoutes(router *mux.Router) {
	// Apply middleware to all routes
	router.Use(h.middleware.ApplyMiddleware)
	
	// Apply audit middleware for admin gateways after other middleware
	if h.auditService != nil {
		router.Use(h.auditService.AuditMiddleware())
	}
	
	// Health and metrics endpoints (bypass proxy)
	if h.config.Observability.HealthCheckPath != "" {
		router.HandleFunc(h.config.Observability.HealthCheckPath, h.HealthCheck).Methods("GET")
	}
	
	if h.config.Observability.ReadinessPath != "" {
		router.HandleFunc(h.config.Observability.ReadinessPath, h.ReadinessCheck).Methods("GET")
	}
	
	if h.config.Observability.MetricsPath != "" {
		router.HandleFunc(h.config.Observability.MetricsPath, h.MetricsEndpoint).Methods("GET")
	}
	
	// Service proxy routes - admin gateways use /admin prefix, public gateways use /api prefix
	apiPrefix := "/api"
	if h.config.IsAdmin() {
		apiPrefix = "/admin/api"
	}
	
	if h.config.ServiceRouting.ContentAPIEnabled {
		// Content API routes (includes research and events)
		router.PathPrefix(apiPrefix + "/v1/content").HandlerFunc(h.ProxyToContentAPI).Methods("GET", "OPTIONS")
		router.PathPrefix(apiPrefix + "/v1/research").HandlerFunc(h.ProxyToContentAPI).Methods("GET", "PUT", "POST", "DELETE", "OPTIONS")
		router.PathPrefix(apiPrefix + "/v1/events").HandlerFunc(h.ProxyToContentAPI).Methods("GET", "OPTIONS")
		
		// Simple API routes for development (without v1 prefix)
		router.HandleFunc("/api/news", h.ProxyToContentAPI).Methods("GET", "OPTIONS")
		router.HandleFunc("/api/events", h.ProxyToContentAPI).Methods("GET", "OPTIONS")
		router.HandleFunc("/api/research", h.ProxyToContentAPI).Methods("GET", "OPTIONS")
	}
	
	if h.config.ServiceRouting.ServicesAPIEnabled {
		// Services API routes
		router.PathPrefix(apiPrefix + "/v1/services").HandlerFunc(h.ProxyToServicesAPI).Methods("GET", "POST", "PUT", "DELETE", "OPTIONS")
		
		// Simple API route for development
		router.HandleFunc("/api/services", h.ProxyToServicesAPI).Methods("GET", "OPTIONS")
	}
	
	if h.config.ServiceRouting.NewsAPIEnabled {
		// News API routes
		router.PathPrefix(apiPrefix + "/v1/news").HandlerFunc(h.ProxyToNewsAPI).Methods("GET", "POST", "PUT", "DELETE", "OPTIONS")
	}
	
	// Inquiries API routes (always enabled for public gateway, admin routes for admin gateway)
	if h.config.IsPublic() {
		router.PathPrefix("/api/v1/inquiries").HandlerFunc(h.ProxyToInquiriesAPI).Methods("POST", "OPTIONS")
	}
	if h.config.IsAdmin() {
		router.HandleFunc("/api/admin/inquiries", h.ProxyToInquiriesAPI).Methods("GET", "POST", "PUT", "DELETE", "OPTIONS")
		router.HandleFunc("/api/admin/subscribers", h.ProxyToNotificationAPI).Methods("GET", "POST", "PUT", "DELETE", "OPTIONS")
	}
	
	if h.config.ServiceRouting.NotificationAPIEnabled {
		// Notification API routes (admin-only)
		router.PathPrefix(apiPrefix + "/v1/notifications").HandlerFunc(h.ProxyToNotificationAPI).Methods("GET", "POST", "PUT", "DELETE")
	}
	
	// Gateway information endpoint
	router.HandleFunc("/gateway/info", h.GatewayInfo).Methods("GET")
	
	// Admin-specific routes (subscriber management)
	if h.config.IsAdmin() && h.subscriberHandler != nil {
		h.subscriberHandler.RegisterSubscriberRoutes(router)
		
		// Additional admin health check endpoint for subscriber management
		router.HandleFunc("/admin/subscribers/health", h.subscriberHandler.SubscriberHealthCheck).Methods("GET")
	}
}

// ProxyToContentAPI proxies requests to content API service
func (h *GatewayHandler) ProxyToContentAPI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Add request timeout
	ctx, cancel := context.WithTimeout(ctx, h.config.Timeouts.RequestTimeout)
	defer cancel()
	
	// Proxy request to content service (matching deployed service name)
	err := h.serviceProxy.ProxyRequest(ctx, w, r, "content")
	if err != nil {
		h.handleError(w, r, err)
		return
	}
}

// ProxyToServicesAPI proxies requests to services API service
func (h *GatewayHandler) ProxyToServicesAPI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Add request timeout
	ctx, cancel := context.WithTimeout(ctx, h.config.Timeouts.RequestTimeout)
	defer cancel()
	
	// Proxy request to content service (services domain is consolidated)
	err := h.serviceProxy.ProxyRequest(ctx, w, r, "content")
	if err != nil {
		h.handleError(w, r, err)
		return
	}
}

// ProxyToNewsAPI proxies requests to news API service
func (h *GatewayHandler) ProxyToNewsAPI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Add request timeout
	ctx, cancel := context.WithTimeout(ctx, h.config.Timeouts.RequestTimeout)
	defer cancel()
	
	// Proxy request to content service (news domain is consolidated)
	err := h.serviceProxy.ProxyRequest(ctx, w, r, "content")
	if err != nil {
		h.handleError(w, r, err)
		return
	}
}

// ProxyToNotificationAPI proxies requests to notification API service
func (h *GatewayHandler) ProxyToNotificationAPI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Add request timeout
	ctx, cancel := context.WithTimeout(ctx, h.config.Timeouts.RequestTimeout)
	defer cancel()
	
	// Proxy request to notifications service
	err := h.serviceProxy.ProxyRequest(ctx, w, r, "notifications")
	if err != nil {
		h.handleError(w, r, err)
		return
	}
}

// ProxyToInquiriesAPI proxies requests to inquiries API service
func (h *GatewayHandler) ProxyToInquiriesAPI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Add request timeout
	ctx, cancel := context.WithTimeout(ctx, h.config.Timeouts.RequestTimeout)
	defer cancel()
	
	// Proxy request to inquiries service
	err := h.serviceProxy.ProxyRequest(ctx, w, r, "inquiries")
	if err != nil {
		h.handleError(w, r, err)
		return
	}
}

// HealthCheck provides a health check endpoint
func (h *GatewayHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Check gateway health
	health := map[string]interface{}{
		"status":    "ok",
		"gateway":   h.config.Name,
		"version":   h.config.Version,
		"timestamp": time.Now().UTC(),
	}
	
	// Check backend service health
	if err := h.serviceProxy.HealthCheck(ctx); err != nil {
		health["status"] = "degraded"
		health["backend_services"] = "unhealthy"
		health["error"] = err.Error()
		
		h.writeJSONResponse(w, r, http.StatusServiceUnavailable, health)
		return
	}
	
	health["backend_services"] = "healthy"
	h.writeJSONResponse(w, r, http.StatusOK, health)
}

// ReadinessCheck provides a readiness check endpoint
func (h *GatewayHandler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	readiness := map[string]interface{}{
		"status":    "ready",
		"gateway":   h.config.Name,
		"version":   h.config.Version,
		"timestamp": time.Now().UTC(),
	}
	
	// Check if gateway is ready to accept traffic
	ready := true
	reasons := []string{}
	
	// Check backend service readiness
	if err := h.serviceProxy.HealthCheck(ctx); err != nil {
		ready = false
		reasons = append(reasons, "backend services not ready")
	}
	
	// Check configuration validity
	if h.config.Port == 0 {
		ready = false
		reasons = append(reasons, "invalid configuration")
	}
	
	if !ready {
		readiness["status"] = "not_ready"
		readiness["reasons"] = reasons
		h.writeJSONResponse(w, r, http.StatusServiceUnavailable, readiness)
		return
	}
	
	h.writeJSONResponse(w, r, http.StatusOK, readiness)
}

// MetricsEndpoint provides metrics information
func (h *GatewayHandler) MetricsEndpoint(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Get service metrics from proxy
	serviceMetrics, err := h.serviceProxy.GetServiceMetrics(ctx)
	if err != nil {
		h.handleError(w, r, domain.NewInternalError("failed to get service metrics", err))
		return
	}
	
	// Add gateway-specific metrics
	metrics := map[string]interface{}{
		"gateway": map[string]interface{}{
			"name":        h.config.Name,
			"type":        h.config.Type,
			"version":     h.config.Version,
			"environment": h.config.Environment,
			"uptime":      time.Now().UTC(),
			"configuration": map[string]interface{}{
				"rate_limit_enabled":       h.config.RateLimit.Enabled,
				"cors_enabled":             h.config.CORS.Enabled,
				"auth_required":            h.config.ShouldRequireAuth(),
				"content_api_enabled":      h.config.ServiceRouting.ContentAPIEnabled,
				"services_api_enabled":     h.config.ServiceRouting.ServicesAPIEnabled,
				"notification_api_enabled": h.config.ServiceRouting.NotificationAPIEnabled,
			},
		},
		"services": serviceMetrics,
	}
	
	h.writeJSONResponse(w, r, http.StatusOK, metrics)
}

// GatewayInfo provides gateway information
func (h *GatewayHandler) GatewayInfo(w http.ResponseWriter, r *http.Request) {
	info := map[string]interface{}{
		"name":        h.config.Name,
		"type":        h.config.Type,
		"version":     h.config.Version,
		"environment": h.config.Environment,
		"capabilities": map[string]interface{}{
			"content_api":       h.config.ServiceRouting.ContentAPIEnabled,
			"services_api":      h.config.ServiceRouting.ServicesAPIEnabled,
			"notification_api":  h.config.ServiceRouting.NotificationAPIEnabled,
			"rate_limiting":     h.config.RateLimit.Enabled,
			"cors":              h.config.CORS.Enabled,
			"authentication":    h.config.ShouldRequireAuth(),
		},
		"endpoints": map[string]interface{}{
			"health":    h.config.Observability.HealthCheckPath,
			"readiness": h.config.Observability.ReadinessPath,
			"metrics":   h.config.Observability.MetricsPath,
		},
	}
	
	// Add CORS information for public gateway
	if h.config.IsPublic() {
		info["cors"] = map[string]interface{}{
			"allowed_origins": h.config.CORS.AllowedOrigins,
			"allowed_methods": h.config.CORS.AllowedMethods,
		}
	}
	
	h.writeJSONResponse(w, r, http.StatusOK, info)
}

// NotFoundHandler handles undefined routes
func (h *GatewayHandler) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	correlationID := domain.GetCorrelationID(r.Context())
	
	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":           "ROUTE_NOT_FOUND",
			"message":        "The requested route was not found",
			"path":           r.URL.Path,
			"method":         r.Method,
			"correlation_id": correlationID,
		},
		"gateway": map[string]interface{}{
			"name":    h.config.Name,
			"version": h.config.Version,
		},
	}
	
	h.writeJSONResponse(w, r, http.StatusNotFound, errorResponse)
}

// Helper methods

// handleError handles different types of domain errors and converts them to HTTP responses
func (h *GatewayHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	correlationID := domain.GetCorrelationID(r.Context())
	
	var statusCode int
	var errorCode string
	var message string

	// Handle domain errors
	switch {
	case domain.IsValidationError(err):
		statusCode = http.StatusBadRequest
		errorCode = "VALIDATION_ERROR"
		message = err.Error()
	case domain.IsNotFoundError(err):
		statusCode = http.StatusNotFound
		errorCode = "NOT_FOUND"
		message = err.Error()
	case domain.IsUnauthorizedError(err):
		statusCode = http.StatusUnauthorized
		errorCode = "UNAUTHORIZED"
		message = err.Error()
	case domain.IsForbiddenError(err):
		statusCode = http.StatusForbidden
		errorCode = "FORBIDDEN"
		message = err.Error()
	case domain.IsConflictError(err):
		statusCode = http.StatusConflict
		errorCode = "CONFLICT"
		message = err.Error()
	case domain.IsRateLimitError(err):
		statusCode = http.StatusTooManyRequests
		errorCode = "RATE_LIMIT_EXCEEDED"
		message = err.Error()
	case domain.IsDependencyError(err):
		statusCode = http.StatusBadGateway
		errorCode = "DEPENDENCY_ERROR"
		message = err.Error()
	default:
		statusCode = http.StatusBadGateway
		errorCode = "GATEWAY_ERROR"
		message = "Gateway processing error occurred"
	}

	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":           errorCode,
			"message":        message,
			"correlation_id": correlationID,
		},
		"gateway": map[string]interface{}{
			"name":    h.config.Name,
			"version": h.config.Version,
		},
	}

	h.writeJSONResponse(w, r, statusCode, errorResponse)
}

// writeJSONResponse writes a JSON response with proper headers
func (h *GatewayHandler) writeJSONResponse(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}) {
	// Set correlation ID header
	if correlationID := domain.GetCorrelationID(r.Context()); correlationID != "" {
		w.Header().Set("X-Correlation-ID", correlationID)
	}
	w.Header().Set("Content-Type", "application/json")
	
	// Set cache control based on gateway configuration
	if h.config.CacheControl.Enabled && statusCode == http.StatusOK {
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", h.config.CacheControl.MaxAge))
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}
	
	w.WriteHeader(statusCode)
	
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// CreateRouter creates a configured router for the gateway
func (h *GatewayHandler) CreateRouter() *mux.Router {
	router := mux.NewRouter()
	h.RegisterRoutes(router)
	
	// Register 404 handler with middleware applied
	router.NotFoundHandler = h.middleware.ApplyMiddleware(http.HandlerFunc(h.NotFoundHandler))
	
	return router
}

// GetConfiguration returns the gateway configuration
func (h *GatewayHandler) GetConfiguration() *GatewayConfiguration {
	return h.config
}

// SetSubscriberHandler sets the subscriber handler for admin gateways
func (h *GatewayHandler) SetSubscriberHandler(subscriberHandler *SubscriberHandler) {
	h.subscriberHandler = subscriberHandler
}

// GetSubscriberHandler returns the subscriber handler
func (h *GatewayHandler) GetSubscriberHandler() *SubscriberHandler {
	return h.subscriberHandler
}

// SetAuditService sets the audit service for admin gateways
func (h *GatewayHandler) SetAuditService(auditService *AuditService) {
	h.auditService = auditService
}