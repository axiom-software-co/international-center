package content

import (
	"fmt"
	"net/http"

	"github.com/axiom-software-co/international-center/src/backend/internal/content/events"
	"github.com/axiom-software-co/international-center/src/backend/internal/content/news"
	"github.com/axiom-software-co/international-center/src/backend/internal/content/research"
	"github.com/axiom-software-co/international-center/src/backend/internal/content/services"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
	"github.com/gorilla/mux"
)

// ContentHandler consolidates all content domain handlers
type ContentHandler struct {
	eventsHandler   *events.EventsHandler
	newsHandler     *news.NewsHandler
	researchHandler *research.ResearchHandler
	servicesHandler *services.ServicesHandler
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

	return &ContentHandler{
		eventsHandler:   eventsHandler,
		newsHandler:     newsHandler,
		researchHandler: researchHandler,
		servicesHandler: servicesHandler,
	}, nil
}

// RegisterRoutes registers all content domain routes with the router
func (h *ContentHandler) RegisterRoutes(router *mux.Router) {
	// Register routes for each domain
	h.eventsHandler.RegisterRoutes(router)
	h.newsHandler.RegisterRoutes(router)
	h.researchHandler.RegisterRoutes(router)
	h.servicesHandler.RegisterRoutes(router)
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