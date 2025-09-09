package volunteers

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/gorilla/mux"
)

// VolunteerHandler handles HTTP requests for volunteer operations
type VolunteerHandler struct {
	service *VolunteerService
}

// NewVolunteerHandler creates a new volunteer handler
func NewVolunteerHandler(service *VolunteerService) *VolunteerHandler {
	return &VolunteerHandler{
		service: service,
	}
}

// RegisterRoutes registers volunteer routes with the router
func (h *VolunteerHandler) RegisterRoutes(router *mux.Router) {
	// Public endpoints for volunteer applications
	router.HandleFunc("/api/v1/volunteers/applications", h.CreateVolunteerApplication).Methods("POST")
	
	// Admin endpoints - will be handled by admin gateway
	// Volunteer application management
	router.HandleFunc("/admin/api/v1/volunteers/applications", h.GetAllVolunteerApplications).Methods("GET")
	router.HandleFunc("/admin/api/v1/volunteers/applications/{id}", h.GetVolunteerApplication).Methods("GET")
	router.HandleFunc("/admin/api/v1/volunteers/applications/search", h.SearchVolunteerApplications).Methods("GET")
	router.HandleFunc("/admin/api/v1/volunteers/applications/status/{status}", h.GetVolunteerApplicationsByStatus).Methods("GET")
	router.HandleFunc("/admin/api/v1/volunteers/applications/interest/{interest}", h.GetVolunteerApplicationsByInterest).Methods("GET")
	router.HandleFunc("/admin/api/v1/volunteers/applications/{id}/status", h.UpdateVolunteerApplicationStatus).Methods("PUT")
	router.HandleFunc("/admin/api/v1/volunteers/applications/{id}/priority", h.UpdateVolunteerApplicationPriority).Methods("PUT")
	router.HandleFunc("/admin/api/v1/volunteers/applications/{id}", h.DeleteVolunteerApplication).Methods("DELETE")
	router.HandleFunc("/admin/api/v1/volunteers/applications/{id}/audit", h.GetVolunteerApplicationAudit).Methods("GET")
}

// Public API endpoints

// CreateVolunteerApplication handles POST /api/v1/volunteers/applications
func (h *VolunteerHandler) CreateVolunteerApplication(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Parse request body
	var application VolunteerApplication
	if err := json.NewDecoder(r.Body).Decode(&application); err != nil {
		h.handleError(w, r, domain.NewValidationError("invalid request body: "+err.Error()))
		return
	}

	// Extract IP address for metadata
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		application.IPAddress = net.ParseIP(ip)
	}

	// Extract user agent
	application.UserAgent = r.UserAgent()

	// System user ID for public applications
	userID := "system"
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "volunteers-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Create the application
	if err := h.service.CreateVolunteerApplication(ctx, &application, userID); err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, map[string]interface{}{
		"application_id": application.ApplicationID,
		"status":         application.Status,
		"message":        "Volunteer application submitted successfully",
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// Admin API endpoints

// GetAllVolunteerApplications handles GET /admin/api/v1/volunteers/applications
func (h *VolunteerHandler) GetAllVolunteerApplications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "volunteers-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Extract pagination parameters
	limit, offset := h.extractPaginationParams(r)

	applications, err := h.service.GetAllVolunteerApplications(ctx, limit, offset)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"applications": applications,
		"count":        len(applications),
		"pagination": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
		},
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// GetVolunteerApplication handles GET /admin/api/v1/volunteers/applications/{id}
func (h *VolunteerHandler) GetVolunteerApplication(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	applicationID := vars["id"]

	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "volunteers-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	application, err := h.service.GetVolunteerApplication(ctx, applicationID, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"application":    application,
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// SearchVolunteerApplications handles GET /admin/api/v1/volunteers/applications/search
func (h *VolunteerHandler) SearchVolunteerApplications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "volunteers-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Extract search query
	query := r.URL.Query().Get("q")
	if query == "" {
		h.handleError(w, r, domain.NewValidationError("search query parameter 'q' is required"))
		return
	}

	// Extract pagination parameters
	limit, offset := h.extractPaginationParams(r)

	applications, err := h.service.SearchVolunteerApplications(ctx, query, limit, offset)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"applications": applications,
		"count":        len(applications),
		"query":        query,
		"pagination": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
		},
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// GetVolunteerApplicationsByStatus handles GET /admin/api/v1/volunteers/applications/status/{status}
func (h *VolunteerHandler) GetVolunteerApplicationsByStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	status := ApplicationStatus(vars["status"])

	// Validate status
	if !IsValidApplicationStatus(status) {
		h.handleError(w, r, domain.NewValidationError("invalid application status"))
		return
	}

	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "volunteers-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Extract pagination parameters
	limit, offset := h.extractPaginationParams(r)

	applications, err := h.service.GetVolunteerApplicationsByStatus(ctx, status, limit, offset)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"applications": applications,
		"count":        len(applications),
		"status":       status,
		"pagination": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
		},
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// GetVolunteerApplicationsByInterest handles GET /admin/api/v1/volunteers/applications/interest/{interest}
func (h *VolunteerHandler) GetVolunteerApplicationsByInterest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	interest := VolunteerInterest(vars["interest"])

	// Validate interest
	if !IsValidVolunteerInterest(interest) {
		h.handleError(w, r, domain.NewValidationError("invalid volunteer interest"))
		return
	}

	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "volunteers-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Extract pagination parameters
	limit, offset := h.extractPaginationParams(r)

	applications, err := h.service.GetVolunteerApplicationsByInterest(ctx, interest, limit, offset)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"applications": applications,
		"count":        len(applications),
		"interest":     interest,
		"pagination": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
		},
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// UpdateVolunteerApplicationStatus handles PUT /admin/api/v1/volunteers/applications/{id}/status
func (h *VolunteerHandler) UpdateVolunteerApplicationStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	applicationID := vars["id"]

	// Parse request body
	var request struct {
		Status ApplicationStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.handleError(w, r, domain.NewValidationError("invalid request body: "+err.Error()))
		return
	}

	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "volunteers-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Update status
	if err := h.service.UpdateVolunteerApplicationStatus(ctx, applicationID, request.Status, userID); err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message":        "Volunteer application status updated successfully",
		"application_id": applicationID,
		"status":         request.Status,
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// UpdateVolunteerApplicationPriority handles PUT /admin/api/v1/volunteers/applications/{id}/priority
func (h *VolunteerHandler) UpdateVolunteerApplicationPriority(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	applicationID := vars["id"]

	// Parse request body
	var request struct {
		Priority ApplicationPriority `json:"priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.handleError(w, r, domain.NewValidationError("invalid request body: "+err.Error()))
		return
	}

	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "volunteers-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Update priority
	if err := h.service.UpdateVolunteerApplicationPriority(ctx, applicationID, request.Priority, userID); err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message":        "Volunteer application priority updated successfully",
		"application_id": applicationID,
		"priority":       request.Priority,
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// DeleteVolunteerApplication handles DELETE /admin/api/v1/volunteers/applications/{id}
func (h *VolunteerHandler) DeleteVolunteerApplication(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	applicationID := vars["id"]

	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "volunteers-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Delete application
	if err := h.service.DeleteVolunteerApplication(ctx, applicationID, userID); err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message":        "Volunteer application deleted successfully",
		"application_id": applicationID,
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// GetVolunteerApplicationAudit handles GET /admin/api/v1/volunteers/applications/{id}/audit
func (h *VolunteerHandler) GetVolunteerApplicationAudit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	applicationID := vars["id"]

	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "volunteers-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Extract pagination parameters
	limit, offset := h.extractPaginationParams(r)

	auditEvents, err := h.service.GetVolunteerApplicationAudit(ctx, applicationID, userID, limit, offset)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"audit_events":   auditEvents,
		"count":          len(auditEvents),
		"application_id": applicationID,
		"pagination": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
		},
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// Health check endpoints

// HealthCheck handles GET /health
func (h *VolunteerHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "volunteers-api",
		"version": "1.0.0",
	})
}

// ReadinessCheck handles GET /health/ready
func (h *VolunteerHandler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	// In production, this would check dependencies like database connectivity
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":  "ready",
		"service": "volunteers-api",
		"version": "1.0.0",
	})
}

// Helper methods

// getUserIDFromContext extracts user ID from request context
func (h *VolunteerHandler) getUserIDFromContext(r *http.Request) string {
	// This would be populated by authentication middleware
	// For now, check for a test header
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		return userID
	}
	
	// Return empty string for anonymous access
	return ""
}

// extractPaginationParams extracts limit and offset from query parameters
func (h *VolunteerHandler) extractPaginationParams(r *http.Request) (int, int) {
	limit := 10 // default
	offset := 0 // default

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	return limit, offset
}

// handleError handles HTTP errors with proper status codes and responses
func (h *VolunteerHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
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
	default:
		statusCode = http.StatusInternalServerError
		errorCode = "INTERNAL_ERROR"
		message = "An internal server error occurred"
	}

	// Log error for debugging (in production, would use structured logging)
	// log.Printf("Error processing request %s %s: %v", r.Method, r.URL.Path, err)

	h.writeJSONResponse(w, statusCode, map[string]interface{}{
		"error": map[string]interface{}{
			"code":           errorCode,
			"message":        message,
			"correlation_id": correlationID,
		},
	})
}

// writeJSONResponse writes a JSON response with proper headers
func (h *VolunteerHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	
	// Add security headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	
	w.WriteHeader(statusCode)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			// If we can't encode the response, there's not much we can do
			// In production, would log this error
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}