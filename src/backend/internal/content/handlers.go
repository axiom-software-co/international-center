package content

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/gorilla/mux"
)

// ContentHandler handles HTTP requests for content operations
type ContentHandler struct {
	service *ContentService
}

// NewContentHandler creates a new content handler
func NewContentHandler(service *ContentService) *ContentHandler {
	return &ContentHandler{
		service: service,
	}
}

// RegisterRoutes registers content routes with the router
func (h *ContentHandler) RegisterRoutes(router *mux.Router) {
	// GET endpoints only (as specified in requirements)
	router.HandleFunc("/api/v1/content", h.GetAllContent).Methods("GET")
	router.HandleFunc("/api/v1/content/{id}", h.GetContent).Methods("GET")
	router.HandleFunc("/api/v1/content/{id}/download", h.GetContentDownload).Methods("GET")
	router.HandleFunc("/api/v1/content/{id}/preview", h.GetContentPreview).Methods("GET")
}

// GetAllContent handles GET /api/v1/content
func (h *ContentHandler) GetAllContent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from context (would be set by authentication middleware)
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "content-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Check for search parameter
	searchTerm := r.URL.Query().Get("search")
	
	var contents []*Content
	var err error
	
	if searchTerm != "" {
		contents, err = h.service.SearchContent(ctx, searchTerm, userID)
	} else {
		contents, err = h.service.GetAllContent(ctx, userID)
	}
	
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"content":        contents,
		"count":          len(contents),
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// GetContent handles GET /api/v1/content/{id}
func (h *ContentHandler) GetContent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	contentID := vars["id"]

	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "content-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	content, err := h.service.GetContent(ctx, contentID, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"content":        content,
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// GetContentDownload handles GET /api/v1/content/{id}/download
func (h *ContentHandler) GetContentDownload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	contentID := vars["id"]

	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "content-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	downloadURL, err := h.service.GetContentDownload(ctx, contentID, userID)
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

// GetContentPreview handles GET /api/v1/content/{id}/preview
func (h *ContentHandler) GetContentPreview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	contentID := vars["id"]

	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "content-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	previewURL, err := h.service.GetContentPreview(ctx, contentID, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"preview_url":    previewURL,
		"expires_in":     1800, // 30 minutes in seconds
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// Helper methods

// getUserIDFromContext extracts user ID from request context
func (h *ContentHandler) getUserIDFromContext(r *http.Request) string {
	// This would be populated by authentication middleware
	// For now, check for a test header
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		return userID
	}
	
	// Return empty string for anonymous access
	return ""
}

// handleError handles different types of domain errors and converts them to HTTP responses
func (h *ContentHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
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
func (h *ContentHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
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

// Admin Content Audit and Analytics Handlers

// GetContentAudit handles GET /admin/api/v1/content/{id}/audit
func (h *ContentHandler) GetContentAudit(w http.ResponseWriter, r *http.Request) {
	// Extract content ID from URL path
	vars := mux.Vars(r)
	contentID := vars["id"]
	
	if contentID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Content ID is required")
		return
	}

	// Extract pagination parameters
	limit, offset := h.extractPaginationParams(r)
	
	// Extract user ID from context (would come from authentication middleware)
	userID := r.Header.Get("X-User-ID")

	// Call service method
	auditEvents, err := h.service.GetContentAudit(r.Context(), contentID, userID, limit, offset)
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

// GetContentProcessingQueue handles GET /admin/api/v1/content/processing-queue
func (h *ContentHandler) GetContentProcessingQueue(w http.ResponseWriter, r *http.Request) {
	// Extract pagination parameters
	limit, offset := h.extractPaginationParams(r)
	
	// Extract user ID from context (would come from authentication middleware)
	userID := r.Header.Get("X-User-ID")

	// Call service method
	queueItems, err := h.service.GetContentProcessingQueue(r.Context(), userID, limit, offset)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	// Return processing queue
	response := map[string]interface{}{
		"processing_queue": queueItems,
		"pagination": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  len(queueItems),
		},
	}
	
	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetContentAnalytics handles GET /admin/api/v1/content/analytics
func (h *ContentHandler) GetContentAnalytics(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context (would come from authentication middleware)
	userID := r.Header.Get("X-User-ID")

	// Call service method
	analytics, err := h.service.GetContentAnalytics(r.Context(), userID)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	// Return analytics
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"analytics": analytics,
	})
}

// HealthCheck provides a health check endpoint
func (h *ContentHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"service":   "content-api",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// ReadinessCheck provides a readiness check endpoint
func (h *ContentHandler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would check:
	// - Dapr connectivity
	// - State store accessibility
	// - Blob storage connectivity
	// For now, just return OK
	
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":  "ready",
		"service": "content-api",
	})
}

// Additional helper methods

// writeErrorResponse writes a simple error response
func (h *ContentHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
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
func (h *ContentHandler) extractPaginationParams(r *http.Request) (limit int, offset int) {
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
func (h *ContentHandler) handleServiceError(w http.ResponseWriter, err error) {
	h.handleError(w, &http.Request{}, err)
}