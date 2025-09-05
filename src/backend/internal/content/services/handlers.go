package services

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
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
	
	// Admin endpoints - will be handled by admin gateway
	// Service admin endpoints
	router.HandleFunc("/admin/api/v1/services", h.CreateService).Methods("POST")
	router.HandleFunc("/admin/api/v1/services/{id}", h.UpdateService).Methods("PUT")
	router.HandleFunc("/admin/api/v1/services/{id}", h.DeleteService).Methods("DELETE")
	router.HandleFunc("/admin/api/v1/services/{id}/publish", h.PublishService).Methods("POST")
	router.HandleFunc("/admin/api/v1/services/{id}/archive", h.ArchiveService).Methods("POST")
	
	// Service category admin endpoints
	router.HandleFunc("/admin/api/v1/services/categories", h.CreateServiceCategory).Methods("POST")
	router.HandleFunc("/admin/api/v1/services/categories/{id}", h.UpdateServiceCategory).Methods("PUT")
	router.HandleFunc("/admin/api/v1/services/categories/{id}", h.DeleteServiceCategory).Methods("DELETE")
	
	// Featured categories admin endpoint
	router.HandleFunc("/admin/api/v1/services/featured", h.SetFeaturedCategories).Methods("PUT")
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

// Admin Audit and Analytics Handlers

// GetServiceAudit handles GET /admin/api/v1/services/{id}/audit
func (h *ServicesHandler) GetServiceAudit(w http.ResponseWriter, r *http.Request) {
	// Extract service ID from URL path
	vars := mux.Vars(r)
	serviceID := vars["id"]
	
	if serviceID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Service ID is required")
		return
	}

	// Extract pagination parameters
	limit, offset := h.extractPaginationParams(r)
	
	// Extract user ID from context (would come from authentication middleware)
	userID := r.Header.Get("X-User-ID")

	// Call service method
	auditEvents, err := h.service.GetServiceAudit(r.Context(), serviceID, userID, limit, offset)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	// Return audit events
	response := map[string]interface{}{
		"audit_events": auditEvents,
		"pagination": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  len(auditEvents),
		},
	}
	
	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetServiceCategoryAudit handles GET /admin/api/v1/services/categories/{id}/audit
func (h *ServicesHandler) GetServiceCategoryAudit(w http.ResponseWriter, r *http.Request) {
	// Extract category ID from URL path
	vars := mux.Vars(r)
	categoryID := vars["id"]
	
	if categoryID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Category ID is required")
		return
	}

	// Extract pagination parameters
	limit, offset := h.extractPaginationParams(r)
	
	// Extract user ID from context (would come from authentication middleware)
	userID := r.Header.Get("X-User-ID")

	// Call service method
	auditEvents, err := h.service.GetServiceCategoryAudit(r.Context(), categoryID, userID, limit, offset)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	// Return audit events
	response := map[string]interface{}{
		"audit_events": auditEvents,
		"pagination": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  len(auditEvents),
		},
	}
	
	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetAdminFeaturedCategories handles GET /admin/api/v1/services/featured-categories
func (h *ServicesHandler) GetAdminFeaturedCategories(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context (would come from authentication middleware)
	userID := r.Header.Get("X-User-ID")

	// Call service method
	featuredCategories, err := h.service.GetAdminFeaturedCategories(r.Context(), userID)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	// Return featured categories with admin details
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"featured_categories": featuredCategories,
	})
}

// HealthCheck provides a health check endpoint
func (h *ServicesHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"service":   "services-api",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
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

// Additional helper methods

// writeErrorResponse writes a simple error response
func (h *ServicesHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := map[string]interface{}{
		"error": map[string]interface{}{
			"message": message,
		},
	}
	
	json.NewEncoder(w).Encode(response)
}

// extractPaginationParams extracts limit and offset from query parameters
func (h *ServicesHandler) extractPaginationParams(r *http.Request) (limit int, offset int) {
	limit = 20 // default limit
	offset = 0 // default offset
	
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	
	return limit, offset
}

// handleServiceError handles service errors (alias for handleError for consistency)
func (h *ServicesHandler) handleServiceError(w http.ResponseWriter, err error) {
	h.handleError(w, &http.Request{}, err)
}

// Admin Service CRUD Handlers

// CreateService handles POST /admin/api/v1/services
func (h *ServicesHandler) CreateService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from header (set by authentication middleware)
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "User ID is required")
		return
	}

	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "services-admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Parse request body
	var service Service
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Create service through service
	if err := h.service.AdminCreateService(ctx, &service, userID); err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, map[string]interface{}{
		"service": &service,
		"message": "Service created successfully",
	})
}

// UpdateService handles PUT /admin/api/v1/services/{id}
func (h *ServicesHandler) UpdateService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from header
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "User ID is required")
		return
	}

	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "services-admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Parse request body
	var service Service
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Get service ID from URL parameters
	vars := mux.Vars(r)
	serviceID := vars["id"]
	service.ServiceID = serviceID

	// Update service through service
	if err := h.service.AdminUpdateService(ctx, &service, userID); err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"service": &service,
		"message": "Service updated successfully",
	})
}

// DeleteService handles DELETE /admin/api/v1/services/{id}
func (h *ServicesHandler) DeleteService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from header
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "User ID is required")
		return
	}

	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "services-admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Get service ID from URL parameters
	vars := mux.Vars(r)
	serviceID := vars["id"]

	// Delete service through service
	if err := h.service.AdminDeleteService(ctx, serviceID, userID); err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Service deleted successfully",
	})
}

// PublishService handles POST /admin/api/v1/services/{id}/publish
func (h *ServicesHandler) PublishService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from header
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "User ID is required")
		return
	}

	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "services-admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Get service ID from URL parameters
	vars := mux.Vars(r)
	serviceID := vars["id"]

	// Publish service through service
	if err := h.service.AdminPublishService(ctx, serviceID, userID); err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Service published successfully",
	})
}

// ArchiveService handles POST /admin/api/v1/services/{id}/archive
func (h *ServicesHandler) ArchiveService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from header
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "User ID is required")
		return
	}

	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "services-admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Get service ID from URL parameters
	vars := mux.Vars(r)
	serviceID := vars["id"]

	// Archive service through service
	if err := h.service.AdminArchiveService(ctx, serviceID, userID); err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Service archived successfully",
	})
}

// Admin Service Category CRUD Handlers

// CreateServiceCategory handles POST /admin/api/v1/services/categories
func (h *ServicesHandler) CreateServiceCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from header
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "User ID is required")
		return
	}

	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "services-admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Parse request body
	var category ServiceCategory
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Create category through service
	if err := h.service.AdminCreateServiceCategory(ctx, &category, userID); err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, map[string]interface{}{
		"category": &category,
		"message":  "Service category created successfully",
	})
}

// UpdateServiceCategory handles PUT /admin/api/v1/services/categories/{id}
func (h *ServicesHandler) UpdateServiceCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from header
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "User ID is required")
		return
	}

	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "services-admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Parse request body
	var category ServiceCategory
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Get category ID from URL parameters
	vars := mux.Vars(r)
	categoryID := vars["id"]
	category.CategoryID = categoryID

	// Update category through service
	if err := h.service.AdminUpdateServiceCategory(ctx, &category, userID); err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"category": &category,
		"message":  "Service category updated successfully",
	})
}

// DeleteServiceCategory handles DELETE /admin/api/v1/services/categories/{id}
func (h *ServicesHandler) DeleteServiceCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from header
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "User ID is required")
		return
	}

	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "services-admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Get category ID from URL parameters
	vars := mux.Vars(r)
	categoryID := vars["id"]

	// Delete category through service
	if err := h.service.AdminDeleteServiceCategory(ctx, categoryID, userID); err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Service category deleted successfully",
	})
}

// SetFeaturedCategories handles PUT /admin/api/v1/services/featured
func (h *ServicesHandler) SetFeaturedCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from header
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "User ID is required")
		return
	}

	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "services-admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Parse request body
	var request struct {
		CategoryIDs []string `json:"category_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Set featured categories through service
	if err := h.service.AdminSetFeaturedCategories(ctx, request.CategoryIDs, userID); err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message":      "Featured categories updated successfully",
		"category_ids": request.CategoryIDs,
	})
}