package news

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/gorilla/mux"
)

// NewsHandler handles HTTP requests for news operations
type NewsHandler struct {
	service *NewsService
}

// NewNewsHandler creates a new news handler
func NewNewsHandler(service *NewsService) *NewsHandler {
	return &NewsHandler{
		service: service,
	}
}

// RegisterRoutes registers news routes with the router
func (h *NewsHandler) RegisterRoutes(router *mux.Router) {
	// Public GET endpoints
	router.HandleFunc("/api/v1/news", h.GetAllNews).Methods("GET")
	router.HandleFunc("/api/v1/news/{id}", h.GetNews).Methods("GET")
	router.HandleFunc("/api/v1/news/slug/{slug}", h.GetNewsBySlug).Methods("GET")
	router.HandleFunc("/api/v1/news/featured", h.GetFeaturedNews).Methods("GET")
	router.HandleFunc("/api/v1/news/categories", h.GetAllNewsCategories).Methods("GET")
	router.HandleFunc("/api/v1/news/categories/{id}/news", h.GetNewsByCategory).Methods("GET")
	router.HandleFunc("/api/v1/news/search", h.SearchNews).Methods("GET")
	
	// Admin endpoints - will be handled by admin gateway
	router.HandleFunc("/admin/api/v1/news/{id}/audit", h.GetNewsAudit).Methods("GET")
	router.HandleFunc("/admin/api/v1/news/categories/{id}/audit", h.GetNewsCategoryAudit).Methods("GET")
}

// Public API endpoints

// GetAllNews handles GET /api/v1/news
func (h *NewsHandler) GetAllNews(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from context (would be set by authentication middleware)
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "news-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Check for search parameter
	searchTerm := r.URL.Query().Get("search")
	
	var newsList []*News
	var err error
	
	if searchTerm != "" {
		newsList, err = h.service.SearchNews(ctx, searchTerm, userID)
	} else {
		newsList, err = h.service.GetAllNews(ctx, userID)
	}
	
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"news":           newsList,
		"count":          len(newsList),
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// GetNews handles GET /api/v1/news/{id}
func (h *NewsHandler) GetNews(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	newsID := vars["id"]

	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "news-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	news, err := h.service.GetNews(ctx, newsID, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"news":           news,
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// GetNewsBySlug handles GET /api/v1/news/slug/{slug}
func (h *NewsHandler) GetNewsBySlug(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	slug := vars["slug"]

	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "news-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	news, err := h.service.GetNewsBySlug(ctx, slug, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"news":           news,
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// GetFeaturedNews handles GET /api/v1/news/featured
func (h *NewsHandler) GetFeaturedNews(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "news-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	news, err := h.service.GetFeaturedNews(ctx, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"featured_news":  news,
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// GetAllNewsCategories handles GET /api/v1/news/categories
func (h *NewsHandler) GetAllNewsCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "news-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	categories, err := h.service.GetAllNewsCategories(ctx, userID)
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

// GetNewsByCategory handles GET /api/v1/news/categories/{id}/news
func (h *NewsHandler) GetNewsByCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	categoryID := vars["id"]

	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "news-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	newsList, err := h.service.GetNewsByCategory(ctx, categoryID, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"news":           newsList,
		"count":          len(newsList),
		"category_id":    categoryID,
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// SearchNews handles GET /api/v1/news/search
func (h *NewsHandler) SearchNews(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	searchTerm := r.URL.Query().Get("q")
	
	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "news-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	results, err := h.service.SearchNews(ctx, searchTerm, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"news":           results,
		"count":          len(results),
		"search_term":    searchTerm,
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// Admin API endpoints

// GetNewsAudit handles GET /admin/api/v1/news/{id}/audit
func (h *NewsHandler) GetNewsAudit(w http.ResponseWriter, r *http.Request) {
	// Extract news ID from URL path
	vars := mux.Vars(r)
	newsID := vars["id"]
	
	if newsID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "News ID is required")
		return
	}

	// Extract pagination parameters
	limit, offset := h.extractPaginationParams(r)
	
	// Extract user ID from context (would come from authentication middleware)
	userID := r.Header.Get("X-User-ID")

	// Call service method
	auditEvents, err := h.service.GetNewsAudit(r.Context(), newsID, userID, limit, offset)
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

// GetNewsCategoryAudit handles GET /admin/api/v1/news/categories/{id}/audit
func (h *NewsHandler) GetNewsCategoryAudit(w http.ResponseWriter, r *http.Request) {
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
	auditEvents, err := h.service.GetNewsCategoryAudit(r.Context(), categoryID, userID, limit, offset)
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

// HealthCheck provides a health check endpoint
func (h *NewsHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"service":   "news-api",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// ReadinessCheck provides a readiness check endpoint
func (h *NewsHandler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would check:
	// - Dapr connectivity
	// - State store accessibility
	// - Blob storage connectivity
	// For now, just return OK
	
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":  "ready",
		"service": "news-api",
	})
}

// Helper methods

// getUserIDFromContext extracts user ID from request context
func (h *NewsHandler) getUserIDFromContext(r *http.Request) string {
	// This would be populated by authentication middleware
	// For now, check for a test header
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		return userID
	}
	
	// Return empty string for anonymous access
	return ""
}

// handleError handles different types of domain errors and converts them to HTTP responses
func (h *NewsHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
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
func (h *NewsHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
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

// writeErrorResponse writes a simple error response
func (h *NewsHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
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
func (h *NewsHandler) extractPaginationParams(r *http.Request) (limit int, offset int) {
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
func (h *NewsHandler) handleServiceError(w http.ResponseWriter, err error) {
	h.handleError(w, &http.Request{}, err)
}