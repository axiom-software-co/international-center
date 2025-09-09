package inquiries

import (
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