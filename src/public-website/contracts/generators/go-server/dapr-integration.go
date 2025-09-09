package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// DAPRServiceInterface defines the contract for DAPR service implementations
type DAPRServiceInterface interface {
	// Service identification
	GetServiceName() string
	GetServiceVersion() string
	
	// Health check methods
	HealthCheck(ctx context.Context) error
	ReadinessCheck(ctx context.Context) error
	
	// DAPR-specific methods
	HandleServiceInvocation(ctx context.Context, method string, data []byte) ([]byte, error)
	HandlePubSubEvent(ctx context.Context, topic string, data []byte) error
}

// PublicGatewayService defines the interface for public gateway service implementations
type PublicGatewayService interface {
	DAPRServiceInterface
	
	// Services domain
	GetServices(ctx context.Context, page, limit int, search string) (interface{}, error)
	GetServiceByID(ctx context.Context, id string) (interface{}, error)
	GetServiceBySlug(ctx context.Context, slug string) (interface{}, error)
	GetFeaturedServices(ctx context.Context) (interface{}, error)
	
	// News domain  
	GetNews(ctx context.Context, page, limit int, search string) (interface{}, error)
	GetNewsByID(ctx context.Context, id string) (interface{}, error)
	GetNewsBySlug(ctx context.Context, slug string) (interface{}, error)
	GetFeaturedNews(ctx context.Context) (interface{}, error)
	
	// Research domain
	GetResearch(ctx context.Context, page, limit int, search string) (interface{}, error)
	GetResearchByID(ctx context.Context, id string) (interface{}, error)
	GetResearchBySlug(ctx context.Context, slug string) (interface{}, error)
	
	// Events domain
	GetEvents(ctx context.Context, page, limit int, search string) (interface{}, error)
	GetEventByID(ctx context.Context, id string) (interface{}, error)
	RegisterForEvent(ctx context.Context, eventID string, registration interface{}) (interface{}, error)
	
	// Inquiries
	SubmitMediaInquiry(ctx context.Context, inquiry interface{}) (interface{}, error)
	SubmitBusinessInquiry(ctx context.Context, inquiry interface{}) (interface{}, error)
	SubmitDonationInquiry(ctx context.Context, inquiry interface{}) (interface{}, error)
	SubmitVolunteerInquiry(ctx context.Context, inquiry interface{}) (interface{}, error)
}

// AdminGatewayService defines the interface for admin gateway service implementations  
type AdminGatewayService interface {
	DAPRServiceInterface
	
	// Authentication
	Login(ctx context.Context, credentials interface{}) (interface{}, error)
	RefreshToken(ctx context.Context, refreshToken string) (interface{}, error)
	Logout(ctx context.Context) error
	
	// User management
	GetAdminUsers(ctx context.Context, page, limit int, search string) (interface{}, error)
	CreateAdminUser(ctx context.Context, user interface{}) (interface{}, error)
	UpdateAdminUser(ctx context.Context, id string, user interface{}) (interface{}, error)
	DeleteAdminUser(ctx context.Context, id string) error
	
	// Content management
	CreateNewsArticle(ctx context.Context, article interface{}) (interface{}, error)
	UpdateNewsArticle(ctx context.Context, id string, article interface{}) (interface{}, error)
	DeleteNewsArticle(ctx context.Context, id string) error
	PublishNewsArticle(ctx context.Context, id string) error
	UnpublishNewsArticle(ctx context.Context, id string) error
	
	// Inquiry management
	GetInquiries(ctx context.Context, page, limit int, inquiryType, status string) (interface{}, error)
	UpdateInquiryStatus(ctx context.Context, id, status string) error
	
	// Analytics
	GetDashboardAnalytics(ctx context.Context, period string) (interface{}, error)
}

// DAPRHTTPHandler wraps the generated HTTP handlers with DAPR middleware
type DAPRHTTPHandler struct {
	service DAPRServiceInterface
	router  *mux.Router
}

func NewDAPRHTTPHandler(service DAPRServiceInterface) *DAPRHTTPHandler {
	handler := &DAPRHTTPHandler{
		service: service,
		router:  mux.NewRouter(),
	}
	
	handler.setupRoutes()
	return handler
}

func (h *DAPRHTTPHandler) setupRoutes() {
	// DAPR health endpoints
	h.router.HandleFunc("/health", h.handleHealth).Methods("GET")
	h.router.HandleFunc("/health/ready", h.handleReady).Methods("GET")
	
	// DAPR service invocation endpoint
	h.router.HandleFunc("/invoke/{method}", h.handleServiceInvocation).Methods("POST")
	
	// DAPR pubsub endpoint
	h.router.HandleFunc("/pubsub/{topic}", h.handlePubSubEvent).Methods("POST")
}

func (h *DAPRHTTPHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	if err := h.service.HealthCheck(ctx); err != nil {
		http.Error(w, fmt.Sprintf("Health check failed: %v", err), http.StatusServiceUnavailable)
		return
	}
	
	response := map[string]interface{}{
		"status":    "healthy",
		"service":   h.service.GetServiceName(),
		"version":   h.service.GetServiceVersion(),
		"timestamp": "2024-01-01T00:00:00Z", // Should use actual timestamp
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *DAPRHTTPHandler) handleReady(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	if err := h.service.ReadinessCheck(ctx); err != nil {
		http.Error(w, fmt.Sprintf("Readiness check failed: %v", err), http.StatusServiceUnavailable)
		return
	}
	
	response := map[string]interface{}{
		"status":  "ready",
		"service": h.service.GetServiceName(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *DAPRHTTPHandler) handleServiceInvocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	method := vars["method"]
	
	var data []byte
	if r.Body != nil {
		defer r.Body.Close()
		var err error
		data, err = json.Marshal(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
			return
		}
	}
	
	result, err := h.service.HandleServiceInvocation(r.Context(), method, data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Service invocation failed: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

func (h *DAPRHTTPHandler) handlePubSubEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topic := vars["topic"]
	
	var data []byte
	if r.Body != nil {
		defer r.Body.Close()
		var err error
		data, err = json.Marshal(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
			return
		}
	}
	
	if err := h.service.HandlePubSubEvent(r.Context(), topic, data); err != nil {
		http.Error(w, fmt.Sprintf("PubSub event handling failed: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}

func (h *DAPRHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Apply CORS middleware for development
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
	
	handler := c.Handler(h.router)
	handler.ServeHTTP(w, r)
}

// BaseService provides a base implementation of DAPRServiceInterface
type BaseService struct {
	serviceName    string
	serviceVersion string
}

func NewBaseService(name, version string) *BaseService {
	return &BaseService{
		serviceName:    name,
		serviceVersion: version,
	}
}

func (s *BaseService) GetServiceName() string {
	return s.serviceName
}

func (s *BaseService) GetServiceVersion() string {
	return s.serviceVersion
}

func (s *BaseService) HealthCheck(ctx context.Context) error {
	// Base health check - override in concrete implementations
	return nil
}

func (s *BaseService) ReadinessCheck(ctx context.Context) error {
	// Base readiness check - override in concrete implementations
	return nil
}

func (s *BaseService) HandleServiceInvocation(ctx context.Context, method string, data []byte) ([]byte, error) {
	// Base implementation - override in concrete implementations
	return nil, fmt.Errorf("service invocation method %s not implemented", method)
}

func (s *BaseService) HandlePubSubEvent(ctx context.Context, topic string, data []byte) error {
	// Base implementation - override in concrete implementations
	return fmt.Errorf("pubsub topic %s not implemented", topic)
}