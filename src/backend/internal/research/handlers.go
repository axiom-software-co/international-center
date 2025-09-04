package research

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/gorilla/mux"
)

// ResearchHandler handles HTTP requests for research operations
type ResearchHandler struct {
	service *ResearchService
}

// NewResearchHandler creates a new research handler
func NewResearchHandler(service *ResearchService) *ResearchHandler {
	return &ResearchHandler{
		service: service,
	}
}

// RegisterRoutes registers research routes with the router
func (h *ResearchHandler) RegisterRoutes(router *mux.Router) {
	// Public GET endpoints
	router.HandleFunc("/api/v1/research", h.GetAllResearch).Methods("GET")
	router.HandleFunc("/api/v1/research/{id}", h.GetResearch).Methods("GET")
	router.HandleFunc("/api/v1/research/slug/{slug}", h.GetResearchBySlug).Methods("GET")
	router.HandleFunc("/api/v1/research/featured", h.GetFeaturedResearch).Methods("GET")
	router.HandleFunc("/api/v1/research/categories", h.GetAllResearchCategories).Methods("GET")
	router.HandleFunc("/api/v1/research/categories/{id}/research", h.GetResearchByCategory).Methods("GET")
	router.HandleFunc("/api/v1/research/search", h.SearchResearch).Methods("GET")
	router.HandleFunc("/api/v1/research/{id}/report", h.GetResearchReport).Methods("GET")
	
	// Admin endpoints - will be handled by admin gateway
	router.HandleFunc("/admin/api/v1/research/{id}/audit", h.GetResearchAudit).Methods("GET")
	router.HandleFunc("/admin/api/v1/research/categories/{id}/audit", h.GetResearchCategoryAudit).Methods("GET")
}

// Public API endpoints

// GetAllResearch handles GET /api/v1/research
func (h *ResearchHandler) GetAllResearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from context (would be set by authentication middleware)
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "research-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Extract pagination parameters
	limit, offset := h.extractPaginationParams(r)

	// Check for search parameter
	searchTerm := r.URL.Query().Get("search")
	
	var researchList []*Research
	var err error
	
	if searchTerm != "" {
		researchList, err = h.service.SearchResearch(ctx, searchTerm, limit, offset)
	} else {
		researchList, err = h.service.GetAllResearch(ctx, limit, offset)
	}
	
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"research":       researchList,
		"count":          len(researchList),
		"pagination": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
		},
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// GetResearch handles GET /api/v1/research/{id}
func (h *ResearchHandler) GetResearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	researchID := vars["id"]

	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "research-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	research, err := h.service.GetResearch(ctx, researchID, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"research":       research,
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// GetResearchBySlug handles GET /api/v1/research/slug/{slug}
func (h *ResearchHandler) GetResearchBySlug(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	slug := vars["slug"]

	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "research-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	research, err := h.service.GetResearchBySlug(ctx, slug, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"research":       research,
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// GetFeaturedResearch handles GET /api/v1/research/featured
func (h *ResearchHandler) GetFeaturedResearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "research-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	featuredResearch, err := h.service.GetFeaturedResearch(ctx)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"featured_research": featuredResearch,
		"correlation_id":    correlationCtx.CorrelationID,
	})
}

// GetAllResearchCategories handles GET /api/v1/research/categories
func (h *ResearchHandler) GetAllResearchCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "research-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	categories, err := h.service.GetAllResearchCategories(ctx)
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

// GetResearchByCategory handles GET /api/v1/research/categories/{id}/research
func (h *ResearchHandler) GetResearchByCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	categoryID := vars["id"]

	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "research-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Extract pagination parameters
	limit, offset := h.extractPaginationParams(r)

	researchList, err := h.service.GetResearchByCategory(ctx, categoryID, limit, offset)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"research":       researchList,
		"count":          len(researchList),
		"category_id":    categoryID,
		"pagination": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
		},
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// SearchResearch handles GET /api/v1/research/search
func (h *ResearchHandler) SearchResearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	searchTerm := r.URL.Query().Get("q")
	
	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "research-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Extract pagination parameters
	limit, offset := h.extractPaginationParams(r)

	results, err := h.service.SearchResearch(ctx, searchTerm, limit, offset)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"research":       results,
		"count":          len(results),
		"search_term":    searchTerm,
		"pagination": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
		},
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// GetResearchReport handles GET /api/v1/research/{id}/report
func (h *ResearchHandler) GetResearchReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	researchID := vars["id"]

	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "research-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// First get the research to validate existence and get report URL
	research, err := h.service.GetResearch(ctx, researchID, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	// Check if research has a report URL
	if research.ReportURL == "" {
		h.writeErrorResponse(w, http.StatusNotFound, "No report available for this research")
		return
	}

	// For now, redirect to the report URL
	// In a full implementation, this would fetch from Azure Blob Storage through Dapr
	http.Redirect(w, r, research.ReportURL, http.StatusTemporaryRedirect)
}

// Admin API endpoints

// GetResearchAudit handles GET /admin/api/v1/research/{id}/audit
func (h *ResearchHandler) GetResearchAudit(w http.ResponseWriter, r *http.Request) {
	// Extract research ID from URL path
	vars := mux.Vars(r)
	researchID := vars["id"]
	
	if researchID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Research ID is required")
		return
	}

	// Extract pagination parameters
	limit, offset := h.extractPaginationParams(r)
	
	// Extract user ID from context (would come from authentication middleware)
	userID := r.Header.Get("X-User-ID")

	// Call service method
	auditEvents, err := h.service.GetResearchAudit(r.Context(), researchID, userID, limit, offset)
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

// GetResearchCategoryAudit handles GET /admin/api/v1/research/categories/{id}/audit
func (h *ResearchHandler) GetResearchCategoryAudit(w http.ResponseWriter, r *http.Request) {
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
	auditEvents, err := h.service.GetResearchCategoryAudit(r.Context(), categoryID, userID, limit, offset)
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
func (h *ResearchHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"service":   "research-api",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// ReadinessCheck provides a readiness check endpoint
func (h *ResearchHandler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would check:
	// - Dapr connectivity
	// - State store accessibility
	// - Blob storage connectivity
	// For now, just return OK
	
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":  "ready",
		"service": "research-api",
	})
}

// Helper methods

// getUserIDFromContext extracts user ID from request context
func (h *ResearchHandler) getUserIDFromContext(r *http.Request) string {
	// This would be populated by authentication middleware
	// For now, check for a test header
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		return userID
	}
	
	// Return empty string for anonymous access
	return ""
}

// handleError handles different types of domain errors and converts them to HTTP responses
func (h *ResearchHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
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
func (h *ResearchHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
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
func (h *ResearchHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
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
func (h *ResearchHandler) extractPaginationParams(r *http.Request) (limit int, offset int) {
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
func (h *ResearchHandler) handleServiceError(w http.ResponseWriter, err error) {
	h.handleError(w, &http.Request{}, err)
}