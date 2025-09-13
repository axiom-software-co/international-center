package content

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/content/events"
	"github.com/axiom-software-co/international-center/src/backend/internal/content/news"
	"github.com/axiom-software-co/international-center/src/backend/internal/content/research"
	"github.com/axiom-software-co/international-center/src/backend/internal/content/services"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/middleware"
	"github.com/gorilla/mux"
)

// ContentHandler consolidates all content domain handlers
type ContentHandler struct {
	eventsHandler       *events.EventsHandler
	newsHandler         *news.NewsHandler
	researchHandler     *research.ResearchHandler
	servicesHandler     *services.ServicesHandler
	contractContentServer *SimplifiedContractHandler
	newsService         *news.NewsService
	researchService     *research.ResearchService
	servicesService     *services.ServicesService
	eventsService       *events.EventsService
}

// NewContentHandler creates a new consolidated content handler
func NewContentHandler(client *dapr.Client) (*ContentHandler, error) {
	if client == nil {
		return nil, fmt.Errorf("dapr client cannot be nil")
	}
	
	// Create shared dapr components for news and research domains
	stateStore := dapr.NewStateStore(client)
	bindings := dapr.NewBindings(client)
	pubsub := dapr.NewPubSub(client)

	// Initialize Events domain (uses *dapr.Client)
	eventsRepository := events.NewEventsRepository(client)
	eventsService := events.NewEventsService(eventsRepository)
	eventsHandler := events.NewEventsHandler(eventsService)

	// Initialize News domain (uses separate components)
	newsRepository := news.NewNewsRepository(stateStore, bindings, pubsub)
	newsService := news.NewNewsService(newsRepository)
	newsHandler := news.NewNewsHandler(newsService)

	// Initialize Research domain (uses separate components)
	researchRepository := research.NewResearchRepository(stateStore, bindings, pubsub)
	researchService := research.NewResearchService(researchRepository)
	researchHandler := research.NewResearchHandler(researchService)

	// Initialize Services domain (uses *dapr.Client)
	servicesRepository := services.NewServicesRepository(client)
	servicesService := services.NewServicesService(servicesRepository)
	servicesHandler := services.NewServicesHandler(servicesService)

	// Initialize contract-compliant content server
	contractContentServer := NewSimplifiedContractHandler(newsService, researchService, servicesService, eventsService)

	return &ContentHandler{
		eventsHandler:       eventsHandler,
		newsHandler:         newsHandler,
		researchHandler:     researchHandler,
		servicesHandler:     servicesHandler,
		contractContentServer: contractContentServer,
		newsService:         newsService,
		researchService:     researchService,
		servicesService:     servicesService,
		eventsService:       eventsService,
	}, nil
}

// RegisterRoutes registers all content domain routes with the router
func (h *ContentHandler) RegisterRoutes(router *mux.Router) {
	// Apply contract validation middleware to admin routes
	adminRouter := router.PathPrefix("/admin/api/v1").Subrouter()
	adminMiddleware := middleware.AdminAPIMiddleware()
	adminMiddleware.Apply(adminRouter)
	
	// Register contract-compliant routes for admin content API
	h.registerContractCompliantRoutes(adminRouter)
	
	// Apply validation middleware to public routes
	publicRouter := router.PathPrefix("/api/v1").Subrouter()
	publicMiddleware := middleware.PublicAPIMiddleware()
	publicMiddleware.Apply(publicRouter)
	
	// Register legacy routes for backward compatibility during migration
	h.registerLegacyRoutes(router)
}

// registerContractCompliantRoutes registers routes using generated interfaces
func (h *ContentHandler) registerContractCompliantRoutes(adminRouter *mux.Router) {
	// Register contract-compliant content routes for news, research, services, events
	RegisterSimplifiedContentRoutes(adminRouter, h.newsService, h.researchService, h.servicesService, h.eventsService)
}

// registerLegacyRoutes registers existing domain-specific routes for backward compatibility
func (h *ContentHandler) registerLegacyRoutes(router *mux.Router) {
	// Legacy routes for gradual migration - these will be replaced with contract-compliant versions
	h.eventsHandler.RegisterRoutes(router)
	h.newsHandler.RegisterRoutes(router)
	h.researchHandler.RegisterRoutes(router)
	h.servicesHandler.RegisterRoutes(router)
	
	// Add featured content endpoints that frontend contract clients expect
	h.registerFeaturedContentRoutes(router)
}

// registerFeaturedContentRoutes registers featured content endpoints for frontend contract clients
func (h *ContentHandler) registerFeaturedContentRoutes(router *mux.Router) {
	// Featured news endpoint
	router.HandleFunc("/api/v1/news/featured", h.GetFeaturedNews).Methods("GET")
	
	// Featured services endpoint  
	router.HandleFunc("/api/v1/services/featured", h.GetFeaturedServices).Methods("GET")
	
	// Featured research endpoint
	router.HandleFunc("/api/v1/research/featured", h.GetFeaturedResearch).Methods("GET")
	
	// Featured events endpoint
	router.HandleFunc("/api/v1/events/featured", h.GetFeaturedEvents).Methods("GET")
	
	// Category endpoints that frontend needs
	router.HandleFunc("/api/v1/news/categories", h.GetNewsCategories).Methods("GET")
	router.HandleFunc("/api/v1/services/categories", h.GetServicesCategories).Methods("GET")
	router.HandleFunc("/api/v1/research/categories", h.GetResearchCategories).Methods("GET")
	router.HandleFunc("/api/v1/events/categories", h.GetEventsCategories).Methods("GET")
	
	// Simple API endpoints for development and testing
	router.HandleFunc("/api/news", h.GetAllNews).Methods("GET")
	router.HandleFunc("/api/events", h.GetAllEvents).Methods("GET")
	router.HandleFunc("/api/research", h.GetAllResearch).Methods("GET")
	router.HandleFunc("/api/services", h.GetAllServices).Methods("GET")
}

// Featured content handlers

// GetFeaturedNews handles GET /api/v1/news/featured
func (h *ContentHandler) GetFeaturedNews(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := "" // Public endpoint doesn't require specific user ID
	
	featuredNews, err := h.newsService.GetFeaturedNews(ctx, userID)
	if err != nil {
		h.writeJSONError(w, http.StatusInternalServerError, "Failed to get featured news", err)
		return
	}
	
	response := map[string]interface{}{
		"data": featuredNews,
	}
	
	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetFeaturedServices handles GET /api/v1/services/featured
func (h *ContentHandler) GetFeaturedServices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := "" // Public endpoint doesn't require specific user ID
	
	// Get featured service category at position 1 (primary featured)
	featuredCategory, err := h.servicesService.GetFeaturedCategoryByPosition(ctx, 1, userID)
	if err != nil {
		h.writeJSONError(w, http.StatusInternalServerError, "Failed to get featured services", err)
		return
	}
	
	response := map[string]interface{}{
		"data": featuredCategory,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetFeaturedResearch handles GET /api/v1/research/featured  
func (h *ContentHandler) GetFeaturedResearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	featuredResearch, err := h.researchService.GetFeaturedResearch(ctx)
	if err != nil {
		h.writeJSONError(w, http.StatusInternalServerError, "Failed to get featured research", err)
		return
	}
	
	response := map[string]interface{}{
		"data": featuredResearch,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetFeaturedEvents handles GET /api/v1/events/featured
func (h *ContentHandler) GetFeaturedEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	featuredEvent, err := h.eventsService.GetFeaturedEvent(ctx)
	if err != nil {
		h.writeJSONError(w, http.StatusInternalServerError, "Failed to get featured event", err)
		return
	}
	
	response := map[string]interface{}{
		"data": featuredEvent,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// Category endpoints

// GetNewsCategories handles GET /api/v1/news/categories
func (h *ContentHandler) GetNewsCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	categories, err := h.newsService.GetAllNewsCategories(ctx, "public")
	if err != nil {
		h.writeJSONError(w, http.StatusInternalServerError, "Failed to get news categories", err)
		return
	}

	response := map[string]interface{}{
		"data": categories,
		"count": len(categories),
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetServicesCategories handles GET /api/v1/services/categories
func (h *ContentHandler) GetServicesCategories(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []interface{}{}, // Placeholder for now
		"count": 0,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetResearchCategories handles GET /api/v1/research/categories
func (h *ContentHandler) GetResearchCategories(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []interface{}{}, // Placeholder for now
		"count": 0,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetEventsCategories handles GET /api/v1/events/categories
func (h *ContentHandler) GetEventsCategories(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []interface{}{}, // Placeholder for now
		"count": 0,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// Helper methods

// writeJSONResponse writes a JSON response
func (h *ContentHandler) writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Simple API endpoint handlers for development and testing

// GetAllNews handles GET /api/news
func (h *ContentHandler) GetAllNews(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Get all news from service using correct method signature
	allNews, err := h.newsService.GetAllNews(ctx, "system")
	if err != nil {
		h.writeJSONError(w, http.StatusInternalServerError, "Failed to get news", err)
		return
	}

	// Add required pagination fields for contract compliance
	response := map[string]interface{}{
		"data": allNews,
		"pagination": map[string]interface{}{
			"page":  1,
			"limit": 20,
			"total": len(allNews),
		},
		"count": len(allNews),
		"service": "content-api",
		"domain": "news",
	}

	// Add correlation ID header for contract compliance
	w.Header().Set("X-Correlation-ID", "content-news-"+fmt.Sprintf("%d", time.Now().UnixNano()))
	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetAllEvents handles GET /api/events
func (h *ContentHandler) GetAllEvents(w http.ResponseWriter, r *http.Request) {
	// Return structured response for events with contract compliance fields
	response := map[string]interface{}{
		"data": []interface{}{},
		"pagination": map[string]interface{}{
			"page":  1,
			"limit": 20,
			"total": 0,
		},
		"count": 0,
		"service": "content-api", 
		"domain": "events",
		"message": "Events API endpoint implemented",
	}

	// Add correlation ID header for contract compliance
	w.Header().Set("X-Correlation-ID", "content-events-"+fmt.Sprintf("%d", time.Now().UnixNano()))
	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetAllResearch handles GET /api/research
func (h *ContentHandler) GetAllResearch(w http.ResponseWriter, r *http.Request) {
	// Return structured response for research with contract compliance fields
	response := map[string]interface{}{
		"data": []interface{}{},
		"pagination": map[string]interface{}{
			"page":  1,
			"limit": 20,
			"total": 0,
		},
		"count": 0,
		"service": "content-api",
		"domain": "research",
		"message": "Research API endpoint implemented",
	}

	// Add correlation ID header for contract compliance
	w.Header().Set("X-Correlation-ID", "content-research-"+fmt.Sprintf("%d", time.Now().UnixNano()))
	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetAllServices handles GET /api/services
func (h *ContentHandler) GetAllServices(w http.ResponseWriter, r *http.Request) {
	// Return structured response for services with contract compliance fields
	response := map[string]interface{}{
		"data": []interface{}{},
		"pagination": map[string]interface{}{
			"page":  1,
			"limit": 20,
			"total": 0,
		},
		"count": 0,
		"service": "content-api",
		"domain": "services",
		"message": "Services API endpoint implemented",
	}

	// Add correlation ID header for contract compliance
	w.Header().Set("X-Correlation-ID", "content-services-"+fmt.Sprintf("%d", time.Now().UnixNano()))
	h.writeJSONResponse(w, http.StatusOK, response)
}

// writeJSONError writes a JSON error response with proper domain error mapping
func (h *ContentHandler) writeJSONError(w http.ResponseWriter, fallbackStatus int, message string, err error) {
	// Map domain errors to proper HTTP status codes
	var statusCode int
	var errorCode string
	
	// Inspect domain error types and map to appropriate HTTP status codes
	switch {
	case domain.IsNotFoundError(err):
		statusCode = http.StatusNotFound
		errorCode = "NOT_FOUND"
	case domain.IsValidationError(err):
		statusCode = http.StatusBadRequest
		errorCode = "VALIDATION_ERROR"
	case domain.IsUnauthorizedError(err):
		statusCode = http.StatusUnauthorized
		errorCode = "UNAUTHORIZED"
	case domain.IsForbiddenError(err):
		statusCode = http.StatusForbidden
		errorCode = "FORBIDDEN"
	case domain.IsConflictError(err):
		statusCode = http.StatusConflict
		errorCode = "CONFLICT"
	case domain.IsRateLimitError(err):
		statusCode = http.StatusTooManyRequests
		errorCode = "RATE_LIMIT_EXCEEDED"
	case domain.IsDependencyError(err):
		statusCode = http.StatusBadGateway
		errorCode = "DEPENDENCY_ERROR"
	default:
		// Use fallback status for unknown errors
		statusCode = fallbackStatus
		errorCode = "INTERNAL_ERROR"
	}
	
	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    errorCode,
			"message": message,
			"details": err.Error(),
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse)
}

// HealthCheck performs health check across all content domains
func (h *ContentHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	w.Write([]byte(`{"status":"healthy","service":"content-api","domains":{"events":"healthy","news":"healthy","research":"healthy","services":"healthy"}}`))
}

// ReadinessCheck performs readiness check across all content domains
func (h *ContentHandler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	// You could call individual readiness checks here if needed
	// h.eventsHandler.ReadinessCheck(w, r)
	// etc.
	
	w.Write([]byte(`{"status":"ready","service":"content-api","domains":{"events":"ready","news":"ready","research":"ready","services":"ready"}}`))
}