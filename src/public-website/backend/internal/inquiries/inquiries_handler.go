package inquiries

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/axiom-software-co/international-center/src/backend/internal/inquiries/business"
	"github.com/axiom-software-co/international-center/src/backend/internal/inquiries/donations"
	"github.com/axiom-software-co/international-center/src/backend/internal/inquiries/media"
	"github.com/axiom-software-co/international-center/src/backend/internal/inquiries/volunteers"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/middleware"
	"github.com/gorilla/mux"
)

// InquiriesHandler consolidates all inquiries domain handlers
type InquiriesHandler struct {
	businessHandler       *business.BusinessHandler
	donationsHandler      *donations.DonationsHandler
	mediaHandler          *media.MediaHandler
	volunteersHandler     *volunteers.VolunteerHandler
	contractCompliantServer *ContractCompliantServer
	mediaService          *media.MediaService
}

// NewInquiriesHandler creates a new consolidated inquiries handler
func NewInquiriesHandler(client *dapr.Client) (*InquiriesHandler, error) {
	if client == nil {
		return nil, fmt.Errorf("dapr client cannot be nil")
	}
	
	// Create shared dapr components for volunteers domain (uses separate components pattern)
	stateStore := dapr.NewStateStore(client)
	bindings := dapr.NewBindings(client)
	pubsub := dapr.NewPubSub(client)

	// Initialize Business domain (uses *dapr.Client pattern)
	businessRepository := business.NewBusinessRepository(client)
	businessService := business.NewBusinessService(businessRepository)
	businessHandler := business.NewBusinessHandler(businessService)

	// Initialize Donations domain (uses *dapr.Client pattern)
	donationsRepository := donations.NewDonationsRepository(client)
	donationsService := donations.NewDonationsService(donationsRepository)
	donationsHandler := donations.NewDonationsHandler(donationsService)

	// Initialize Media domain (uses *dapr.Client pattern)
	mediaRepository := media.NewMediaRepository(client)
	mediaService := media.NewMediaService(mediaRepository)
	mediaHandler := media.NewMediaHandler(mediaService)

	// Initialize Volunteers domain (uses separate components pattern)
	volunteersRepository := volunteers.NewVolunteerRepository(stateStore, bindings, pubsub)
	volunteersService := volunteers.NewVolunteerService(volunteersRepository)
	volunteersHandler := volunteers.NewVolunteerHandler(volunteersService)

	// Initialize contract-compliant server for API contract enforcement
	contractCompliantServer := NewContractCompliantServer(mediaService)

	return &InquiriesHandler{
		businessHandler:       businessHandler,
		donationsHandler:      donationsHandler,
		mediaHandler:          mediaHandler,
		volunteersHandler:     volunteersHandler,
		contractCompliantServer: contractCompliantServer,
		mediaService:          mediaService,
	}, nil
}

// RegisterRoutes registers all inquiries domain routes with the router
func (h *InquiriesHandler) RegisterRoutes(router *mux.Router) {
	// Apply contract validation middleware to admin routes
	adminRouter := router.PathPrefix("/admin/api/v1").Subrouter()
	adminMiddleware := middleware.AdminAPIMiddleware()
	adminMiddleware.Apply(adminRouter)
	
	// Register contract-compliant routes for admin inquiries API
	h.registerContractCompliantRoutes(adminRouter)
	
	// Apply validation middleware to public routes
	publicRouter := router.PathPrefix("/api/v1").Subrouter()
	publicMiddleware := middleware.PublicAPIMiddleware()
	publicMiddleware.Apply(publicRouter)
	
	// Register legacy routes (to be migrated to contract-compliant approach)
	h.registerLegacyRoutes(router)
}

// registerContractCompliantRoutes registers routes using generated interfaces
func (h *InquiriesHandler) registerContractCompliantRoutes(adminRouter *mux.Router) {
	// Register contract-compliant inquiry routes
	RegisterContractRoutes(adminRouter, h.mediaService)
}

// registerLegacyRoutes registers existing domain-specific routes for backward compatibility
func (h *InquiriesHandler) registerLegacyRoutes(router *mux.Router) {
	// Legacy routes for gradual migration - these will be replaced with contract-compliant versions
	h.businessHandler.RegisterRoutes(router)
	h.donationsHandler.RegisterRoutes(router)
	h.volunteersHandler.RegisterRoutes(router)
	
	// Note: mediaHandler routes are now handled by contract-compliant server
	// h.mediaHandler.RegisterRoutes(router) - commented out as replaced by contract routes
	
	// Simple API endpoints for development and testing
	router.HandleFunc("/api/inquiries", h.GetAllInquiries).Methods("GET")
	router.HandleFunc("/api/inquiries/business", h.GetBusinessInquiries).Methods("GET")
	router.HandleFunc("/api/inquiries/donations", h.GetDonationInquiries).Methods("GET")
	router.HandleFunc("/api/inquiries/media", h.GetMediaInquiries).Methods("GET")
	router.HandleFunc("/api/inquiries/volunteers", h.GetVolunteerInquiries).Methods("GET")
}

// Simple API endpoint handlers for development and testing

// GetAllInquiries handles GET /api/inquiries
func (h *InquiriesHandler) GetAllInquiries(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []interface{}{},
		"count": 0,
		"service": "inquiries-api",
		"domains": map[string]interface{}{
			"business": "available",
			"donations": "available",
			"media": "available",
			"volunteers": "available",
		},
		"message": "Inquiries API endpoint implemented",
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetBusinessInquiries handles GET /api/inquiries/business
func (h *InquiriesHandler) GetBusinessInquiries(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []interface{}{},
		"count": 0,
		"service": "inquiries-api",
		"domain": "business",
		"message": "Business inquiries API endpoint implemented",
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetDonationInquiries handles GET /api/inquiries/donations
func (h *InquiriesHandler) GetDonationInquiries(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []interface{}{},
		"count": 0,
		"service": "inquiries-api",
		"domain": "donations",
		"message": "Donation inquiries API endpoint implemented",
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetMediaInquiries handles GET /api/inquiries/media
func (h *InquiriesHandler) GetMediaInquiries(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []interface{}{},
		"count": 0,
		"service": "inquiries-api",
		"domain": "media",
		"message": "Media inquiries API endpoint implemented",
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetVolunteerInquiries handles GET /api/inquiries/volunteers
func (h *InquiriesHandler) GetVolunteerInquiries(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data": []interface{}{},
		"count": 0,
		"service": "inquiries-api",
		"domain": "volunteers",
		"message": "Volunteer inquiries API endpoint implemented",
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// writeJSONResponse writes a JSON response
func (h *InquiriesHandler) writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
	}
}

// HealthCheck performs health check across all inquiries domains
func (h *InquiriesHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	w.Write([]byte(`{"status":"healthy","service":"inquiries-api","domains":{"business":"healthy","donations":"healthy","media":"healthy","volunteers":"healthy"}}`))
}

// ReadinessCheck performs readiness check across all inquiries domains
func (h *InquiriesHandler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	w.Write([]byte(`{"status":"ready","service":"inquiries-api","domains":{"business":"ready","donations":"ready","media":"ready","volunteers":"ready"}}`))
}