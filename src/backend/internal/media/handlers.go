package media

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/gorilla/mux"
)

// MediaHandler handles HTTP requests for media inquiry operations
type MediaHandler struct {
	service *MediaService
}

// NewMediaHandler creates a new media handler
func NewMediaHandler(service *MediaService) *MediaHandler {
	return &MediaHandler{
		service: service,
	}
}

// RegisterRoutes registers media inquiry routes with the router
func (h *MediaHandler) RegisterRoutes(router *mux.Router) {
	// Admin endpoints - will be handled by admin gateway
	router.HandleFunc("/admin/api/v1/media/inquiries", h.CreateInquiry).Methods("POST")
	router.HandleFunc("/admin/api/v1/media/inquiries/{id}", h.UpdateInquiry).Methods("PUT")
	router.HandleFunc("/admin/api/v1/media/inquiries/{id}", h.DeleteInquiry).Methods("DELETE")
	router.HandleFunc("/admin/api/v1/media/inquiries/{id}", h.GetInquiry).Methods("GET")
	router.HandleFunc("/admin/api/v1/media/inquiries", h.ListInquiries).Methods("GET")
	router.HandleFunc("/admin/api/v1/media/inquiries/{id}/acknowledge", h.AcknowledgeInquiry).Methods("POST")
	router.HandleFunc("/admin/api/v1/media/inquiries/{id}/resolve", h.ResolveInquiry).Methods("POST")
	router.HandleFunc("/admin/api/v1/media/inquiries/{id}/close", h.CloseInquiry).Methods("POST")
	router.HandleFunc("/admin/api/v1/media/inquiries/{id}/priority", h.SetPriority).Methods("PUT")
}

// Admin media inquiry endpoints

// CreateInquiry handles POST /admin/api/v1/media/inquiries
func (h *MediaHandler) CreateInquiry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from context (would come from authentication middleware)
	userID := r.Header.Get("X-User-ID")
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "media-admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Parse request body
	var request AdminCreateInquiryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.handleError(w, r, domain.NewValidationError("invalid request body"))
		return
	}

	// Call service method
	inquiry, err := h.service.AdminCreateInquiry(ctx, request, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	// Return created inquiry
	h.writeJSONResponse(w, http.StatusCreated, map[string]interface{}{
		"inquiry":        inquiry,
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// UpdateInquiry handles PUT /admin/api/v1/media/inquiries/{id}
func (h *MediaHandler) UpdateInquiry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	inquiryID := vars["id"]
	
	// Extract user ID from context (would come from authentication middleware)
	userID := r.Header.Get("X-User-ID")
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "media-admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Parse request body
	var request AdminUpdateInquiryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.handleError(w, r, domain.NewValidationError("invalid request body"))
		return
	}

	// Call service method
	inquiry, err := h.service.AdminUpdateInquiry(ctx, inquiryID, request, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	// Return updated inquiry
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"inquiry":        inquiry,
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// DeleteInquiry handles DELETE /admin/api/v1/media/inquiries/{id}
func (h *MediaHandler) DeleteInquiry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	inquiryID := vars["id"]
	
	// Extract user ID from context (would come from authentication middleware)
	userID := r.Header.Get("X-User-ID")
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "media-admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Call service method
	err := h.service.AdminDeleteInquiry(ctx, inquiryID, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	// Return success response
	h.writeJSONResponse(w, http.StatusNoContent, map[string]interface{}{
		"message":        "inquiry deleted successfully",
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// GetInquiry handles GET /admin/api/v1/media/inquiries/{id}
func (h *MediaHandler) GetInquiry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	inquiryID := vars["id"]
	
	// Extract user ID from context (would come from authentication middleware)
	userID := r.Header.Get("X-User-ID")
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "media-admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Call service method
	inquiry, err := h.service.AdminGetInquiry(ctx, inquiryID, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	// Return inquiry
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"inquiry":        inquiry,
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// ListInquiries handles GET /admin/api/v1/media/inquiries
func (h *MediaHandler) ListInquiries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from context (would come from authentication middleware)
	userID := r.Header.Get("X-User-ID")
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "media-admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Parse query parameters for filters
	filters := InquiryFilters{}
	
	if status := r.URL.Query().Get("status"); status != "" {
		inquiryStatus := InquiryStatus(status)
		filters.Status = &inquiryStatus
	}
	
	if priority := r.URL.Query().Get("priority"); priority != "" {
		inquiryPriority := InquiryPriority(priority)
		filters.Priority = &inquiryPriority
	}
	
	if urgency := r.URL.Query().Get("urgency"); urgency != "" {
		inquiryUrgency := InquiryUrgency(urgency)
		filters.Urgency = &inquiryUrgency
	}
	
	if mediaType := r.URL.Query().Get("media_type"); mediaType != "" {
		mediaTypeVal := MediaType(mediaType)
		filters.MediaType = &mediaTypeVal
	}
	
	if outlet := r.URL.Query().Get("outlet"); outlet != "" {
		filters.Outlet = &outlet
	}
	
	if limit := r.URL.Query().Get("limit"); limit != "" {
		if limitInt, err := strconv.Atoi(limit); err == nil && limitInt > 0 {
			filters.Limit = &limitInt
		}
	}
	
	if offset := r.URL.Query().Get("offset"); offset != "" {
		if offsetInt, err := strconv.Atoi(offset); err == nil && offsetInt >= 0 {
			filters.Offset = &offsetInt
		}
	}

	// Call service method
	inquiries, err := h.service.AdminListInquiries(ctx, filters, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	// Return inquiries
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"inquiries":      inquiries,
		"count":          len(inquiries),
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// AcknowledgeInquiry handles POST /admin/api/v1/media/inquiries/{id}/acknowledge
func (h *MediaHandler) AcknowledgeInquiry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	inquiryID := vars["id"]
	
	// Extract user ID from context (would come from authentication middleware)
	userID := r.Header.Get("X-User-ID")
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "media-admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Call service method
	inquiry, err := h.service.AdminAcknowledgeInquiry(ctx, inquiryID, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	// Return acknowledged inquiry
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"inquiry":        inquiry,
		"message":        "inquiry acknowledged successfully",
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// ResolveInquiry handles POST /admin/api/v1/media/inquiries/{id}/resolve
func (h *MediaHandler) ResolveInquiry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	inquiryID := vars["id"]
	
	// Extract user ID from context (would come from authentication middleware)
	userID := r.Header.Get("X-User-ID")
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "media-admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Call service method
	inquiry, err := h.service.AdminResolveInquiry(ctx, inquiryID, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	// Return resolved inquiry
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"inquiry":        inquiry,
		"message":        "inquiry resolved successfully",
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// CloseInquiry handles POST /admin/api/v1/media/inquiries/{id}/close
func (h *MediaHandler) CloseInquiry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	inquiryID := vars["id"]
	
	// Extract user ID from context (would come from authentication middleware)
	userID := r.Header.Get("X-User-ID")
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "media-admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Call service method
	inquiry, err := h.service.AdminCloseInquiry(ctx, inquiryID, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	// Return closed inquiry
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"inquiry":        inquiry,
		"message":        "inquiry closed successfully",
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// SetPriority handles PUT /admin/api/v1/media/inquiries/{id}/priority
func (h *MediaHandler) SetPriority(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	inquiryID := vars["id"]
	
	// Extract user ID from context (would come from authentication middleware)
	userID := r.Header.Get("X-User-ID")
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "media-admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Parse request body
	var request struct {
		Priority string `json:"priority" validate:"required,oneof=low medium high urgent"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.handleError(w, r, domain.NewValidationError("invalid request body"))
		return
	}

	priority := InquiryPriority(request.Priority)
	if !priority.IsValid() {
		h.handleError(w, r, domain.NewValidationError("invalid priority value"))
		return
	}

	// Call service method
	inquiry, err := h.service.AdminSetPriority(ctx, inquiryID, priority, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	// Return updated inquiry
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"inquiry":        inquiry,
		"message":        "inquiry priority updated successfully",
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// Helper functions

// handleError handles domain errors and returns appropriate HTTP responses
func (h *MediaHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
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
	case domain.IsTimeoutError(err):
		statusCode = http.StatusRequestTimeout
		errorCode = "TIMEOUT"
		message = err.Error()
	default:
		statusCode = http.StatusInternalServerError
		errorCode = "INTERNAL_ERROR"
		message = "An internal error occurred"
	}

	h.writeJSONResponse(w, statusCode, map[string]interface{}{
		"error": map[string]interface{}{
			"code":    errorCode,
			"message": message,
		},
		"correlation_id": correlationID,
	})
}

// writeJSONResponse writes a JSON response with proper headers
func (h *MediaHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
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
func (h *MediaHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"service":   "media-api",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// ReadinessCheck provides a readiness check endpoint
func (h *MediaHandler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would check:
	// - Dapr connectivity
	// - State store accessibility
	// - Blob storage connectivity
	// For now, just return OK
	
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":  "ready",
		"service": "media-api",
	})
}