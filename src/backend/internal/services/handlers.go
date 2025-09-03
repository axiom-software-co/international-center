package services

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/axiom-software-co/international-center/src/internal/shared/domain"
	"github.com/gorilla/mux"
)

// ServicesHandler handles HTTP requests for services operations
type ServicesHandler struct {
	service *ServicesService
}

// NewServicesHandler creates a new services handler
func NewServicesHandler(service *ServicesService) *ServicesHandler {
	return &ServicesHandler{
		service: service,
	}
}

// RegisterRoutes registers services routes with the router
func (h *ServicesHandler) RegisterRoutes(router *mux.Router) {
	// GET endpoints only (as specified in requirements)
	
	// Service endpoints
	router.HandleFunc("/api/v1/services", h.GetAllServices).Methods("GET")
	router.HandleFunc("/api/v1/services/{id}", h.GetService).Methods("GET")
	router.HandleFunc("/api/v1/services/slug/{slug}", h.GetServiceBySlug).Methods("GET")
	router.HandleFunc("/api/v1/services/{id}/content/download", h.GetServiceContentDownload).Methods("GET")
	
	// Service category endpoints
	router.HandleFunc("/api/v1/services/categories", h.GetAllServiceCategories).Methods("GET")
	router.HandleFunc("/api/v1/services/categories/{id}", h.GetServiceCategory).Methods("GET")
	router.HandleFunc("/api/v1/services/categories/slug/{slug}", h.GetServiceCategoryBySlug).Methods("GET")
	router.HandleFunc("/api/v1/services/categories/{id}/services", h.GetServicesByCategory).Methods("GET")
	
	// Featured category endpoints
	router.HandleFunc("/api/v1/services/featured", h.GetAllFeaturedCategories).Methods("GET")
	router.HandleFunc("/api/v1/services/featured/{position}", h.GetFeaturedCategoryByPosition).Methods("GET")
	
	// Published services endpoint
	router.HandleFunc("/api/v1/services/published", h.GetPublishedServices).Methods("GET")
}

// Service endpoints

// GetAllServices handles GET /api/v1/services
func (h *ServicesHandler) GetAllServices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from context (would be set by authentication middleware)
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "services-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Check for search parameter
	searchTerm := r.URL.Query().Get("search")
	
	var services []*Service
	var err error
	
	if searchTerm != "" {
		services, err = h.service.SearchServices(ctx, searchTerm, userID)
	} else {
		services, err = h.service.GetAllServices(ctx, userID)
	}
	
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"services":       services,
		"count":          len(services),
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// GetService handles GET /api/v1/services/{id}
func (h *ServicesHandler) GetService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	serviceID := vars["id"]

	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "services-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	service, err := h.service.GetService(ctx, serviceID, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"service":        service,
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// GetServiceBySlug handles GET /api/v1/services/slug/{slug}
func (h *ServicesHandler) GetServiceBySlug(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	slug := vars["slug"]

	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "services-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	service, err := h.service.GetServiceBySlug(ctx, slug, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"service":        service,
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// GetServiceContentDownload handles GET /api/v1/services/{id}/content/download
func (h *ServicesHandler) GetServiceContentDownload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	serviceID := vars["id"]

	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "services-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	downloadURL, err := h.service.GetServiceContentDownload(ctx, serviceID, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"download_url":   downloadURL,
		"expires_in":     3600, // 1 hour in seconds
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// GetPublishedServices handles GET /api/v1/services/published
func (h *ServicesHandler) GetPublishedServices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "services-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	services, err := h.service.GetPublishedServices(ctx, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"services":       services,
		"count":          len(services),
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// Service category endpoints

// GetAllServiceCategories handles GET /api/v1/services/categories
func (h *ServicesHandler) GetAllServiceCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "services-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	categories, err := h.service.GetAllServiceCategories(ctx, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"categories":     categories,
		"count":          len(categories),
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// GetServiceCategory handles GET /api/v1/services/categories/{id}
func (h *ServicesHandler) GetServiceCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	categoryID := vars["id"]

	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "services-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	category, err := h.service.GetServiceCategory(ctx, categoryID, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"category":       category,
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// GetServiceCategoryBySlug handles GET /api/v1/services/categories/slug/{slug}
func (h *ServicesHandler) GetServiceCategoryBySlug(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	slug := vars["slug"]

	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "services-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	category, err := h.service.GetServiceCategoryBySlug(ctx, slug, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"category":       category,
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// GetServicesByCategory handles GET /api/v1/services/categories/{id}/services
func (h *ServicesHandler) GetServicesByCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	categoryID := vars["id"]

	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "services-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	services, err := h.service.GetServicesByCategory(ctx, categoryID, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"services":       services,
		"category_id":    categoryID,
		"count":          len(services),
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// Featured category endpoints

// GetAllFeaturedCategories handles GET /api/v1/services/featured
func (h *ServicesHandler) GetAllFeaturedCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "services-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	featured, err := h.service.GetAllFeaturedCategories(ctx, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"featured_categories": featured,
		"count":               len(featured),
		"correlation_id":      correlationCtx.CorrelationID,
	})
}

// GetFeaturedCategoryByPosition handles GET /api/v1/services/featured/{position}
func (h *ServicesHandler) GetFeaturedCategoryByPosition(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	positionStr := vars["position"]

	// Parse position
	position, err := strconv.Atoi(positionStr)
	if err != nil {
		h.handleError(w, r, domain.NewValidationError("invalid position parameter"))
		return
	}

	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "services-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	featured, err := h.service.GetFeaturedCategoryByPosition(ctx, position, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"featured_category": featured,
		"correlation_id":    correlationCtx.CorrelationID,
	})
}

// Helper methods

// getUserIDFromContext extracts user ID from request context
func (h *ServicesHandler) getUserIDFromContext(r *http.Request) string {
	// This would be populated by authentication middleware
	// For now, check for a test header
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		return userID
	}
	
	// Return empty string for anonymous access
	return ""
}

// handleError handles different types of domain errors and converts them to HTTP responses
func (h *ServicesHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
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
	default:
		statusCode = http.StatusInternalServerError
		errorCode = "INTERNAL_ERROR"
		message = "An internal error occurred"
	}

	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":           errorCode,
			"message":        message,
			"correlation_id": correlationID,
		},
	}

	h.writeJSONResponse(w, statusCode, errorResponse)
}

// writeJSONResponse writes a JSON response with proper headers
func (h *ServicesHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	
	// Add security headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	
	w.WriteHeader(statusCode)
	
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// HealthCheck provides a health check endpoint
func (h *ServicesHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"service": "services-api",
	})
}

// ReadinessCheck provides a readiness check endpoint
func (h *ServicesHandler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would check:
	// - Dapr connectivity
	// - State store accessibility
	// - Blob storage connectivity
	// For now, just return OK
	
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":  "ready",
		"service": "services-api",
	})
}