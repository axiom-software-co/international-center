package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/gorilla/mux"
)

// SubscriberHandler handles HTTP requests for subscriber management
type SubscriberHandler struct {
	subscriberService SubscriberService
	gatewayConfig     *GatewayConfiguration
}

// NewSubscriberHandler creates a new subscriber handler
func NewSubscriberHandler(subscriberService SubscriberService, gatewayConfig *GatewayConfiguration) *SubscriberHandler {
	return &SubscriberHandler{
		subscriberService: subscriberService,
		gatewayConfig:     gatewayConfig,
	}
}

// RegisterSubscriberRoutes registers subscriber management routes
func (h *SubscriberHandler) RegisterSubscriberRoutes(router *mux.Router) {
	// Admin-only routes for subscriber management
	adminRouter := router.PathPrefix("/admin").Subrouter()
	
	// Subscriber CRUD operations
	adminRouter.HandleFunc("/subscribers", h.CreateSubscriber).Methods("POST")
	adminRouter.HandleFunc("/subscribers", h.ListSubscribers).Methods("GET")
	adminRouter.HandleFunc("/subscribers/{id}", h.GetSubscriber).Methods("GET")
	adminRouter.HandleFunc("/subscribers/{id}", h.UpdateSubscriber).Methods("PUT")
	adminRouter.HandleFunc("/subscribers/{id}", h.DeleteSubscriber).Methods("DELETE")
	
	// Additional subscriber endpoints
	adminRouter.HandleFunc("/subscribers/search", h.SearchSubscribers).Methods("GET")
	adminRouter.HandleFunc("/subscribers/events/{eventType}", h.GetSubscribersByEvent).Methods("GET")
}

// CreateSubscriber handles POST /admin/subscribers
func (h *SubscriberHandler) CreateSubscriber(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := domain.GetCorrelationID(ctx)

	// Parse request body
	var req CreateSubscriberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "INVALID_JSON", "invalid JSON format", err)
		return
	}

	// Create subscriber
	subscriber, err := h.subscriberService.CreateSubscriber(ctx, &req)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	// Write successful response
	h.writeJSONResponse(w, r, http.StatusCreated, map[string]interface{}{
		"subscriber":     subscriber,
		"message":        "Subscriber created successfully",
		"correlation_id": correlationID,
	})
}

// GetSubscriber handles GET /admin/subscribers/{id}
func (h *SubscriberHandler) GetSubscriber(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract subscriber ID from URL
	vars := mux.Vars(r)
	subscriberID := vars["id"]
	
	if subscriberID == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "MISSING_PARAMETER", "subscriber ID is required", nil)
		return
	}

	// Get subscriber
	subscriber, err := h.subscriberService.GetSubscriber(ctx, subscriberID)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	// Write response
	h.writeJSONResponse(w, r, http.StatusOK, subscriber)
}

// UpdateSubscriber handles PUT /admin/subscribers/{id}
func (h *SubscriberHandler) UpdateSubscriber(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := domain.GetCorrelationID(ctx)
	
	// Extract subscriber ID from URL
	vars := mux.Vars(r)
	subscriberID := vars["id"]
	
	if subscriberID == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "MISSING_PARAMETER", "subscriber ID is required", nil)
		return
	}

	// Parse request body
	var req UpdateSubscriberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "INVALID_JSON", "invalid JSON format", err)
		return
	}

	// Update subscriber
	subscriber, err := h.subscriberService.UpdateSubscriber(ctx, subscriberID, &req)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	// Write response
	h.writeJSONResponse(w, r, http.StatusOK, map[string]interface{}{
		"subscriber":     subscriber,
		"message":        "Subscriber updated successfully",
		"correlation_id": correlationID,
	})
}

// DeleteSubscriber handles DELETE /admin/subscribers/{id}
func (h *SubscriberHandler) DeleteSubscriber(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := domain.GetCorrelationID(ctx)
	
	// Extract subscriber ID from URL
	vars := mux.Vars(r)
	subscriberID := vars["id"]
	
	if subscriberID == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "MISSING_PARAMETER", "subscriber ID is required", nil)
		return
	}

	// Get deleted by from query parameter or header
	deletedBy := r.URL.Query().Get("deleted_by")
	if deletedBy == "" {
		deletedBy = r.Header.Get("X-User-ID")
	}
	if deletedBy == "" {
		deletedBy = "admin" // Default value for admin operations
	}

	// Delete subscriber
	err := h.subscriberService.DeleteSubscriber(ctx, subscriberID, deletedBy)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	// Write response
	h.writeJSONResponse(w, r, http.StatusOK, map[string]interface{}{
		"message":        "Subscriber deleted successfully",
		"subscriber_id":  subscriberID,
		"correlation_id": correlationID,
	})
}

// ListSubscribers handles GET /admin/subscribers
func (h *SubscriberHandler) ListSubscribers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Parse query parameters
	queryParams := r.URL.Query()
	
	// Parse page (default: 1)
	page := 1
	if pageStr := queryParams.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err != nil {
			h.writeErrorResponse(w, r, http.StatusBadRequest, "INVALID_PARAMETER", "invalid page parameter", err)
			return
		} else if p < 1 {
			h.writeErrorResponse(w, r, http.StatusBadRequest, "INVALID_PARAMETER", "invalid page parameter", nil)
			return
		} else {
			page = p
		}
	}
	
	// Parse page size (default: 20, max: 100)
	pageSize := 20
	if sizeStr := queryParams.Get("page_size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err != nil {
			h.writeErrorResponse(w, r, http.StatusBadRequest, "INVALID_PARAMETER", "invalid page size parameter", err)
			return
		} else if s < 1 || s > 100 {
			h.writeErrorResponse(w, r, http.StatusBadRequest, "INVALID_PARAMETER", "invalid page size parameter", nil)
			return
		} else {
			pageSize = s
		}
	}
	
	// Parse status filter
	var status *SubscriberStatus
	if statusStr := queryParams.Get("status"); statusStr != "" {
		s := SubscriberStatus(statusStr)
		if s != SubscriberStatusActive && s != SubscriberStatusInactive && s != SubscriberStatusSuspended {
			h.writeErrorResponse(w, r, http.StatusBadRequest, "INVALID_PARAMETER", "invalid status parameter", nil)
			return
		}
		status = &s
	}

	// List subscribers
	subscribers, total, err := h.subscriberService.ListSubscribers(ctx, status, page, pageSize)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	// Calculate pagination metadata
	totalPages := (total + pageSize - 1) / pageSize
	hasNext := page < totalPages
	hasPrev := page > 1

	// Write response
	response := map[string]interface{}{
		"subscribers": subscribers,
		"pagination": map[string]interface{}{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": totalPages,
			"has_next":    hasNext,
			"has_prev":    hasPrev,
		},
	}
	
	if status != nil {
		response["filter"] = map[string]interface{}{
			"status": *status,
		}
	}

	h.writeJSONResponse(w, r, http.StatusOK, response)
}

// SearchSubscribers handles GET /admin/subscribers/search
func (h *SubscriberHandler) SearchSubscribers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queryParams := r.URL.Query()
	
	// Get search parameters
	email := queryParams.Get("email")
	name := queryParams.Get("name")
	eventType := queryParams.Get("event_type")
	
	if email == "" && name == "" && eventType == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "MISSING_PARAMETER", 
			"at least one search parameter (email, name, or event_type) is required", nil)
		return
	}

	var subscribers []*NotificationSubscriber
	var err error

	// Search by email (exact match)
	if email != "" {
		subscriber, err := h.subscriberService.GetSubscribersByEvent(ctx, EventType(eventType), PriorityThresholdLow)
		if err != nil && !domain.IsNotFoundError(err) {
			h.handleServiceError(w, r, err)
			return
		}
		subscribers = subscriber
	} else if eventType != "" {
		// Search by event type
		subscribers, err = h.subscriberService.GetSubscribersByEvent(ctx, EventType(eventType), PriorityThresholdLow)
		if err != nil {
			h.handleServiceError(w, r, err)
			return
		}
	} else {
		// For name search, we'd need to implement a search method in the service
		// For now, return empty results
		subscribers = []*NotificationSubscriber{}
	}

	// Write response
	h.writeJSONResponse(w, r, http.StatusOK, map[string]interface{}{
		"subscribers": subscribers,
		"total":       len(subscribers),
		"search_criteria": map[string]interface{}{
			"email":      email,
			"name":       name,
			"event_type": eventType,
		},
	})
}

// GetSubscribersByEvent handles GET /admin/subscribers/events/{eventType}
func (h *SubscriberHandler) GetSubscribersByEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract event type from URL
	vars := mux.Vars(r)
	eventTypeStr := vars["eventType"]
	
	if eventTypeStr == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "MISSING_PARAMETER", "event type is required", nil)
		return
	}

	eventType := EventType(eventTypeStr)

	// Get priority from query parameter (default: low)
	priority := PriorityThresholdLow
	if priorityStr := r.URL.Query().Get("priority"); priorityStr != "" {
		priority = PriorityThreshold(priorityStr)
	}

	// Get subscribers by event
	subscribers, err := h.subscriberService.GetSubscribersByEvent(ctx, eventType, priority)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	// Write response
	h.writeJSONResponse(w, r, http.StatusOK, map[string]interface{}{
		"subscribers": subscribers,
		"total":       len(subscribers),
		"event_type":  eventType,
		"priority":    priority,
	})
}

// Helper methods

// handleServiceError converts service errors to HTTP responses
func (h *SubscriberHandler) handleServiceError(w http.ResponseWriter, r *http.Request, err error) {
	var statusCode int
	var errorCode string
	var message string

	switch {
	case domain.IsValidationError(err):
		statusCode = http.StatusBadRequest
		errorCode = "VALIDATION_ERROR"
		message = err.Error()
	case domain.IsNotFoundError(err):
		statusCode = http.StatusNotFound
		errorCode = "SUBSCRIBER_NOT_FOUND"
		message = err.Error()
	case domain.IsConflictError(err):
		statusCode = http.StatusConflict
		errorCode = "SUBSCRIBER_CONFLICT"
		message = err.Error()
	case domain.IsUnauthorizedError(err):
		statusCode = http.StatusUnauthorized
		errorCode = "UNAUTHORIZED"
		message = err.Error()
	case domain.IsForbiddenError(err):
		statusCode = http.StatusForbidden
		errorCode = "FORBIDDEN"
		message = err.Error()
	case domain.IsDependencyError(err):
		statusCode = http.StatusServiceUnavailable
		errorCode = "SERVICE_UNAVAILABLE"
		message = "Database service temporarily unavailable"
	default:
		statusCode = http.StatusInternalServerError
		errorCode = "INTERNAL_ERROR"
		message = "An internal error occurred while processing the request"
	}

	h.writeErrorResponse(w, r, statusCode, errorCode, message, err)
}

// writeErrorResponse writes a standardized error response
func (h *SubscriberHandler) writeErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, errorCode string, message string, err error) {
	correlationID := domain.GetCorrelationID(r.Context())

	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":           errorCode,
			"message":        message,
			"correlation_id": correlationID,
		},
		"gateway": map[string]interface{}{
			"name":    h.gatewayConfig.Name,
			"version": h.gatewayConfig.Version,
		},
	}

	// Add additional error details for debugging (only in development)
	if h.gatewayConfig.Environment == "development" && err != nil {
		errorResponse["debug"] = map[string]interface{}{
			"error_detail": err.Error(),
			"error_type":   fmt.Sprintf("%T", err),
		}
	}

	h.writeJSONResponse(w, r, statusCode, errorResponse)
}

// writeJSONResponse writes a JSON response with proper headers
func (h *SubscriberHandler) writeJSONResponse(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}) {
	// Set correlation ID header
	if correlationID := domain.GetCorrelationID(r.Context()); correlationID != "" {
		w.Header().Set("X-Correlation-ID", correlationID)
	}

	// Set standard headers
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")

	// Set cache control based on response type
	if statusCode >= 200 && statusCode < 300 {
		// Success responses - allow caching for GET requests
		if r.Method == "GET" {
			w.Header().Set("Cache-Control", "private, max-age=300") // 5 minutes
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
	} else {
		// Error responses - no caching
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	}

	// Set admin-specific security headers
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

	w.WriteHeader(statusCode)

	if data != nil {
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ") // Pretty print for admin interface
		if err := encoder.Encode(data); err != nil {
			// Log error but don't expose it to client
			fmt.Printf("Failed to encode JSON response: %v\n", err)
		}
	}
}

// Health check endpoint for subscriber management
func (h *SubscriberHandler) SubscriberHealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	health := map[string]interface{}{
		"status":         "ok",
		"service":        "subscriber-management",
		"version":        h.gatewayConfig.Version,
		"timestamp":      r.Header.Get("X-Request-Time"),
	}

	// Perform basic service health check
	// In a full implementation, this would check database connectivity
	// For now, we assume the service is healthy if it can respond

	h.writeJSONResponse(w, r, http.StatusOK, health)
}

// Validation helper methods

// validateSubscriberID validates subscriber ID from URL parameter
func (h *SubscriberHandler) validateSubscriberID(subscriberID string) error {
	if subscriberID == "" {
		return domain.NewValidationError("subscriber ID cannot be empty", nil)
	}
	
	if strings.TrimSpace(subscriberID) == "" {
		return domain.NewValidationError("subscriber ID cannot be empty", nil)
	}
	
	// Additional validation can be added here
	return nil
}

// extractUserID extracts user ID from request context or headers
func (h *SubscriberHandler) extractUserID(r *http.Request) string {
	// Try to get user ID from context (set by authentication middleware)
	if userID := r.Context().Value("user_id"); userID != nil {
		if uid, ok := userID.(string); ok {
			return uid
		}
	}
	
	// Fallback to header
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		return userID
	}
	
	// Default for admin operations
	return "admin"
}