package inquiries

import (
	"fmt"
	"net/http"

	"github.com/axiom-software-co/international-center/src/backend/internal/inquiries/business"
	"github.com/axiom-software-co/international-center/src/backend/internal/inquiries/donations"
	"github.com/axiom-software-co/international-center/src/backend/internal/inquiries/media"
	"github.com/axiom-software-co/international-center/src/backend/internal/inquiries/volunteers"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
	"github.com/gorilla/mux"
)

// InquiriesHandler consolidates all inquiries domain handlers
type InquiriesHandler struct {
	businessHandler   *business.BusinessHandler
	donationsHandler  *donations.DonationsHandler
	mediaHandler      *media.MediaHandler
	volunteersHandler *volunteers.VolunteerHandler
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

	return &InquiriesHandler{
		businessHandler:   businessHandler,
		donationsHandler:  donationsHandler,
		mediaHandler:      mediaHandler,
		volunteersHandler: volunteersHandler,
	}, nil
}

// RegisterRoutes registers all inquiries domain routes with the router
func (h *InquiriesHandler) RegisterRoutes(router *mux.Router) {
	// Register routes for each domain
	h.businessHandler.RegisterRoutes(router)
	h.donationsHandler.RegisterRoutes(router)
	h.mediaHandler.RegisterRoutes(router)
	h.volunteersHandler.RegisterRoutes(router)
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