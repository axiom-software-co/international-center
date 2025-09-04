package events

import (
	"encoding/json"
	"net/http"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/gorilla/mux"
)

// EventsHandler handles HTTP requests for events operations
type EventsHandler struct {
	service *EventsService
}

// NewEventsHandler creates a new events handler
func NewEventsHandler(service *EventsService) *EventsHandler {
	return &EventsHandler{
		service: service,
	}
}

// RegisterRoutes registers events routes with the router
func (h *EventsHandler) RegisterRoutes(router *mux.Router) {
	// GET endpoints only (as specified in requirements)
	
	// Event endpoints
	router.HandleFunc("/api/v1/events", h.GetAllEvents).Methods("GET")
	router.HandleFunc("/api/v1/events/{id}", h.GetEvent).Methods("GET")
	router.HandleFunc("/api/v1/events/slug/{slug}", h.GetEventBySlug).Methods("GET")
	
	// Event category endpoints
	router.HandleFunc("/api/v1/events/categories", h.GetAllEventCategories).Methods("GET")
	router.HandleFunc("/api/v1/events/categories/{id}", h.GetEventCategory).Methods("GET")
	router.HandleFunc("/api/v1/events/categories/slug/{slug}", h.GetEventCategoryBySlug).Methods("GET")
	router.HandleFunc("/api/v1/events/categories/{id}/events", h.GetEventsByCategory).Methods("GET")
	
	// Featured event endpoints
	router.HandleFunc("/api/v1/events/featured", h.GetFeaturedEvent).Methods("GET")
	
	// Published events endpoint
	router.HandleFunc("/api/v1/events/published", h.GetPublishedEvents).Methods("GET")
	
	// Upcoming events endpoint
	router.HandleFunc("/api/v1/events/upcoming", h.GetUpcomingEvents).Methods("GET")
	
	// Event registrations endpoint (public view)
	router.HandleFunc("/api/v1/events/{id}/registrations/status", h.GetEventRegistrationStatus).Methods("GET")
	
	// Admin endpoints - will be handled by admin gateway
	// Event admin endpoints
	router.HandleFunc("/admin/api/v1/events", h.CreateEvent).Methods("POST")
	router.HandleFunc("/admin/api/v1/events/{id}", h.UpdateEvent).Methods("PUT")
	router.HandleFunc("/admin/api/v1/events/{id}", h.DeleteEvent).Methods("DELETE")
	router.HandleFunc("/admin/api/v1/events/{id}/publish", h.PublishEvent).Methods("POST")
	router.HandleFunc("/admin/api/v1/events/{id}/archive", h.ArchiveEvent).Methods("POST")
	
	// Event category admin endpoints
	router.HandleFunc("/admin/api/v1/events/categories", h.CreateEventCategory).Methods("POST")
	router.HandleFunc("/admin/api/v1/events/categories/{id}", h.DeleteEventCategory).Methods("DELETE")
	
	// Featured event admin endpoint
	router.HandleFunc("/admin/api/v1/events/featured", h.SetFeaturedEvent).Methods("PUT")
	
	// Event registrations admin endpoint
	router.HandleFunc("/admin/api/v1/events/{id}/registrations", h.GetEventRegistrations).Methods("GET")
}

// Public event endpoints

// GetAllEvents handles GET /api/v1/events
func (h *EventsHandler) GetAllEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from context (would be set by authentication middleware)
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "events-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// For public endpoint, return only published events
	// This would be implemented with appropriate service method
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message":        "Public events endpoint - implementation pending",
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// GetEvent handles GET /api/v1/events/{id}
func (h *EventsHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	eventID := vars["id"]

	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "events-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// For public endpoint, only return published events
	// This would be implemented with appropriate service method
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message":        "Public event endpoint - implementation pending",
		"event_id":       eventID,
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// GetFeaturedEvent handles GET /api/v1/events/featured
func (h *EventsHandler) GetFeaturedEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "events-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// This would be implemented with appropriate service method
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message":        "Featured event endpoint - implementation pending",
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// GetUpcomingEvents handles GET /api/v1/events/upcoming
func (h *EventsHandler) GetUpcomingEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "events-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// This would be implemented with appropriate service method
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message":        "Upcoming events endpoint - implementation pending",
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// GetEventRegistrationStatus handles GET /api/v1/events/{id}/registrations/status
func (h *EventsHandler) GetEventRegistrationStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	eventID := vars["id"]
	
	// Extract user ID from context
	userID := h.getUserIDFromContext(r)
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "events-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// This would show registration status without personal details
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message":        "Event registration status endpoint - implementation pending",
		"event_id":       eventID,
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// Admin event endpoints

// CreateEvent handles POST /admin/api/v1/events
func (h *EventsHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from context (would come from authentication middleware)
	userID := r.Header.Get("X-User-ID")
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "events-admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Parse request body
	var request AdminCreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.handleError(w, r, domain.NewValidationError("invalid request body"))
		return
	}

	// Call service method
	event, err := h.service.AdminCreateEvent(ctx, request, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	// Return created event
	h.writeJSONResponse(w, http.StatusCreated, map[string]interface{}{
		"event":          event,
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// UpdateEvent handles PUT /admin/api/v1/events/{id}
func (h *EventsHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	eventID := vars["id"]
	
	// Extract user ID from context (would come from authentication middleware)
	userID := r.Header.Get("X-User-ID")
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "events-admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Parse request body
	var request AdminUpdateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.handleError(w, r, domain.NewValidationError("invalid request body"))
		return
	}

	// Call service method
	event, err := h.service.AdminUpdateEvent(ctx, eventID, request, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	// Return updated event
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"event":          event,
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// DeleteEvent handles DELETE /admin/api/v1/events/{id}
func (h *EventsHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	eventID := vars["id"]
	
	// Extract user ID from context (would come from authentication middleware)
	userID := r.Header.Get("X-User-ID")
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "events-admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Call service method
	err := h.service.AdminDeleteEvent(ctx, eventID, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	// Return success response
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message":        "Event deleted successfully",
		"event_id":       eventID,
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// PublishEvent handles POST /admin/api/v1/events/{id}/publish
func (h *EventsHandler) PublishEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	eventID := vars["id"]
	
	// Extract user ID from context (would come from authentication middleware)
	userID := r.Header.Get("X-User-ID")
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "events-admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Call service method
	event, err := h.service.AdminPublishEvent(ctx, eventID, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	// Return published event
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"event":          event,
		"message":        "Event published successfully",
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// ArchiveEvent handles POST /admin/api/v1/events/{id}/archive
func (h *EventsHandler) ArchiveEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	eventID := vars["id"]
	
	// Extract user ID from context (would come from authentication middleware)
	userID := r.Header.Get("X-User-ID")
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "events-admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Call service method
	event, err := h.service.AdminArchiveEvent(ctx, eventID, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	// Return archived event
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"event":          event,
		"message":        "Event archived successfully",
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// CreateEventCategory handles POST /admin/api/v1/events/categories
func (h *EventsHandler) CreateEventCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from context (would come from authentication middleware)
	userID := r.Header.Get("X-User-ID")
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "events-admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Parse request body
	var request AdminCreateEventCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.handleError(w, r, domain.NewValidationError("invalid request body"))
		return
	}

	// Call service method
	category, err := h.service.AdminCreateEventCategory(ctx, request, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	// Return created category
	h.writeJSONResponse(w, http.StatusCreated, map[string]interface{}{
		"category":       category,
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// DeleteEventCategory handles DELETE /admin/api/v1/events/categories/{id}
func (h *EventsHandler) DeleteEventCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	categoryID := vars["id"]
	
	// Extract user ID from context (would come from authentication middleware)
	userID := r.Header.Get("X-User-ID")
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "events-admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Call service method
	err := h.service.AdminDeleteEventCategory(ctx, categoryID, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	// Return success response
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message":        "Event category deleted successfully",
		"category_id":    categoryID,
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// SetFeaturedEvent handles PUT /admin/api/v1/events/featured
func (h *EventsHandler) SetFeaturedEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from context (would come from authentication middleware)
	userID := r.Header.Get("X-User-ID")
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "events-admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Parse request body
	var request struct {
		EventID string `json:"event_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.handleError(w, r, domain.NewValidationError("invalid request body"))
		return
	}

	// Call service method
	featuredEvent, err := h.service.AdminSetFeaturedEvent(ctx, request.EventID, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	// Return featured event
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"featured_event": featuredEvent,
		"message":        "Featured event set successfully",
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// GetEventRegistrations handles GET /admin/api/v1/events/{id}/registrations
func (h *EventsHandler) GetEventRegistrations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	eventID := vars["id"]
	
	// Extract user ID from context (would come from authentication middleware)
	userID := r.Header.Get("X-User-ID")
	
	// Add correlation context
	correlationCtx := domain.FromContext(ctx)
	correlationCtx.SetUserContext(userID, "events-admin-api-1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	// Call service method
	registrations, err := h.service.AdminGetEventRegistrations(ctx, eventID, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	// Return event registrations
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"registrations":  registrations,
		"count":          len(registrations),
		"event_id":       eventID,
		"correlation_id": correlationCtx.CorrelationID,
	})
}

// Placeholder endpoints for category operations

// GetAllEventCategories handles GET /api/v1/events/categories
func (h *EventsHandler) GetAllEventCategories(w http.ResponseWriter, r *http.Request) {
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Event categories endpoint - implementation pending",
	})
}

// GetEventCategory handles GET /api/v1/events/categories/{id}
func (h *EventsHandler) GetEventCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID := vars["id"]
	
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message":     "Event category endpoint - implementation pending",
		"category_id": categoryID,
	})
}

// GetEventCategoryBySlug handles GET /api/v1/events/categories/slug/{slug}
func (h *EventsHandler) GetEventCategoryBySlug(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]
	
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Event category by slug endpoint - implementation pending",
		"slug":    slug,
	})
}

// GetEventsByCategory handles GET /api/v1/events/categories/{id}/events
func (h *EventsHandler) GetEventsByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID := vars["id"]
	
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message":     "Events by category endpoint - implementation pending",
		"category_id": categoryID,
	})
}

// GetEventBySlug handles GET /api/v1/events/slug/{slug}
func (h *EventsHandler) GetEventBySlug(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]
	
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Event by slug endpoint - implementation pending",
		"slug":    slug,
	})
}

// GetPublishedEvents handles GET /api/v1/events/published
func (h *EventsHandler) GetPublishedEvents(w http.ResponseWriter, r *http.Request) {
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Published events endpoint - implementation pending",
	})
}

// Helper methods

// getUserIDFromContext extracts user ID from request context
func (h *EventsHandler) getUserIDFromContext(r *http.Request) string {
	// This would be populated by authentication middleware
	// For now, check for a test header
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		return userID
	}
	
	// Return empty string for anonymous access
	return ""
}

// handleError handles different types of domain errors and converts them to HTTP responses
func (h *EventsHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
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
func (h *EventsHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
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