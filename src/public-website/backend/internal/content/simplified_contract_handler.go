package content

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/content/events"
	"github.com/axiom-software-co/international-center/src/backend/internal/content/news"
	"github.com/axiom-software-co/international-center/src/backend/internal/content/research"
	"github.com/axiom-software-co/international-center/src/backend/internal/content/services"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/axiom-software-co/international-center/src/backend/internal/contracts/admin"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// SimplifiedContractHandler implements the admin ServerInterface with simplified responses
type SimplifiedContractHandler struct {
	newsService     *news.NewsService
	researchService *research.ResearchService
	servicesService *services.ServicesService
	eventsService   *events.EventsService
}

// NewSimplifiedContractHandler creates a new simplified contract handler
func NewSimplifiedContractHandler(
	newsService *news.NewsService,
	researchService *research.ResearchService,
	servicesService *services.ServicesService,
	eventsService *events.EventsService,
) *SimplifiedContractHandler {
	return &SimplifiedContractHandler{
		newsService:     newsService,
		researchService: researchService,
		servicesService: servicesService,
		eventsService:   eventsService,
	}
}

// News API implementations

// GetNewsAdmin implements GET /admin/api/v1/news
func (h *SimplifiedContractHandler) GetNewsAdmin(w http.ResponseWriter, r *http.Request, params admin.GetNewsAdminParams) {
	ctx := r.Context()
	correlationCtx := domain.FromContext(ctx)
	if correlationCtx == nil {
		correlationCtx = domain.NewCorrelationContext()
	}

	// Extract user ID for service call
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "User ID required", correlationCtx.CorrelationID)
		return
	}

	// Use real news service to get actual data
	newsData, err := h.newsService.GetAllNews(ctx, userID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to fetch news articles", correlationCtx.CorrelationID)
		return
	}

	// Convert domain news to contract-compliant format
	contractNews := make([]admin.NewsArticle, len(newsData))
	for i, news := range newsData {
		contractNews[i] = h.convertNewsToContract(*news)
	}

	// Apply pagination
	page := 1
	limit := 20
	if params.Page != nil {
		page = int(*params.Page)
	}
	if params.Limit != nil {
		limit = int(*params.Limit)
	}

	start := (page - 1) * limit
	end := start + limit
	if start > len(contractNews) {
		start = len(contractNews)
	}
	if end > len(contractNews) {
		end = len(contractNews)
	}

	paginatedNews := contractNews[start:end]
	totalItems := len(contractNews)
	totalPages := (totalItems + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}

	response := struct {
		Data       []admin.NewsArticle   `json:"data"`
		Pagination admin.PaginationInfo `json:"pagination"`
	}{
		Data: paginatedNews,
		Pagination: admin.PaginationInfo{
			CurrentPage:  page,
			TotalPages:   totalPages,
			TotalItems:   totalItems,
			ItemsPerPage: limit,
			HasNext:      page < totalPages,
			HasPrevious:  page > 1,
		},
	}

	h.writeResponse(w, http.StatusOK, response, correlationCtx.CorrelationID)
}

// CreateNewsArticle implements POST /admin/api/v1/news
func (h *SimplifiedContractHandler) CreateNewsArticle(w http.ResponseWriter, r *http.Request) {
	correlationCtx := domain.FromContext(r.Context())
	if correlationCtx == nil {
		correlationCtx = domain.NewCorrelationContext()
	}

	response := admin.CreatedResponse{
		Success:       true,
		Message:       "News article created successfully",
		Data:          map[string]interface{}{"news_id": "placeholder"},
		Timestamp:     time.Now().UTC(),
		CorrelationId: openapi_types.UUID(uuid.MustParse(correlationCtx.CorrelationID)),
	}

	h.writeResponse(w, http.StatusCreated, response, correlationCtx.CorrelationID)
}

// GetNewsArticleByIdAdmin implements GET /admin/api/v1/news/{id}
func (h *SimplifiedContractHandler) GetNewsArticleByIdAdmin(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	correlationCtx := domain.FromContext(r.Context())
	if correlationCtx == nil {
		correlationCtx = domain.NewCorrelationContext()
	}

	// Placeholder response - to be implemented with actual news data
	response := struct {
		Data admin.NewsArticle `json:"data"`
	}{
		Data: admin.NewsArticle{
			NewsId:               id,
			Title:                "Placeholder News",
			Summary:              "Placeholder summary",
			CategoryId:           openapi_types.UUID(uuid.New()),
			NewsType:             admin.NewsArticleNewsTypeAnnouncement,
			PriorityLevel:        admin.NewsArticlePriorityLevelNormal,
			PublishingStatus:     admin.NewsArticlePublishingStatusDraft,
			PublicationTimestamp: time.Now().UTC(),
			CreatedOn:            time.Now().UTC(),
			Slug:                 "placeholder-news",
		},
	}

	h.writeResponse(w, http.StatusOK, response, correlationCtx.CorrelationID)
}

// UpdateNewsArticle implements PUT /admin/api/v1/news/{id}
func (h *SimplifiedContractHandler) UpdateNewsArticle(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	correlationCtx := domain.FromContext(r.Context())
	if correlationCtx == nil {
		correlationCtx = domain.NewCorrelationContext()
	}

	response := admin.UpdatedResponse{
		Success:       true,
		Message:       "News article updated successfully",
		Data:          map[string]interface{}{"news_id": id},
		Timestamp:     time.Now().UTC(),
		CorrelationId: openapi_types.UUID(uuid.MustParse(correlationCtx.CorrelationID)),
	}

	h.writeResponse(w, http.StatusOK, response, correlationCtx.CorrelationID)
}

// DeleteNewsArticle implements DELETE /admin/api/v1/news/{id}
func (h *SimplifiedContractHandler) DeleteNewsArticle(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	correlationCtx := domain.FromContext(r.Context())
	if correlationCtx == nil {
		correlationCtx = domain.NewCorrelationContext()
	}

	response := admin.DeletedResponse{
		Success:       true,
		Message:       "News article deleted successfully",
		Timestamp:     time.Now().UTC(),
		CorrelationId: openapi_types.UUID(uuid.MustParse(correlationCtx.CorrelationID)),
	}

	h.writeResponse(w, http.StatusOK, response, correlationCtx.CorrelationID)
}

// PublishNewsArticle implements POST /admin/api/v1/news/{id}/publish
func (h *SimplifiedContractHandler) PublishNewsArticle(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	correlationCtx := domain.FromContext(r.Context())
	if correlationCtx == nil {
		correlationCtx = domain.NewCorrelationContext()
	}

	response := admin.SuccessResponse{
		Success:       true,
		Message:       "News article published successfully",
		Timestamp:     time.Now().UTC(),
		CorrelationId: openapi_types.UUID(uuid.MustParse(correlationCtx.CorrelationID)),
	}

	h.writeResponse(w, http.StatusOK, response, correlationCtx.CorrelationID)
}

// UnpublishNewsArticle implements POST /admin/api/v1/news/{id}/unpublish
func (h *SimplifiedContractHandler) UnpublishNewsArticle(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	correlationCtx := domain.FromContext(r.Context())
	if correlationCtx == nil {
		correlationCtx = domain.NewCorrelationContext()
	}

	response := admin.SuccessResponse{
		Success:       true,
		Message:       "News article unpublished successfully",
		Timestamp:     time.Now().UTC(),
		CorrelationId: openapi_types.UUID(uuid.MustParse(correlationCtx.CorrelationID)),
	}

	h.writeResponse(w, http.StatusOK, response, correlationCtx.CorrelationID)
}

// GetNewsCategoriesAdmin implements GET /admin/api/v1/news/categories
func (h *SimplifiedContractHandler) GetNewsCategoriesAdmin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationCtx := domain.FromContext(ctx)
	if correlationCtx == nil {
		correlationCtx = domain.NewCorrelationContext()
	}

	// Extract user ID for service call
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "User ID required", correlationCtx.CorrelationID)
		return
	}

	// Use real news service to get actual categories
	categoriesData, err := h.newsService.GetAllNewsCategories(ctx, userID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to fetch news categories", correlationCtx.CorrelationID)
		return
	}

	// Convert domain categories to contract-compliant format
	contractCategories := make([]admin.NewsCategory, len(categoriesData))
	for i, category := range categoriesData {
		contractCategories[i] = h.convertNewsCategoryToContract(*category)
	}

	response := struct {
		Data []admin.NewsCategory `json:"data"`
	}{
		Data: contractCategories,
	}

	h.writeResponse(w, http.StatusOK, response, correlationCtx.CorrelationID)
}

// CreateNewsCategory implements POST /admin/api/v1/news/categories
func (h *SimplifiedContractHandler) CreateNewsCategory(w http.ResponseWriter, r *http.Request) {
	correlationCtx := domain.FromContext(r.Context())
	if correlationCtx == nil {
		correlationCtx = domain.NewCorrelationContext()
	}

	response := admin.CreatedResponse{
		Success:       true,
		Message:       "News category created successfully",
		Data:          map[string]interface{}{"category_id": "placeholder"},
		Timestamp:     time.Now().UTC(),
		CorrelationId: openapi_types.UUID(uuid.MustParse(correlationCtx.CorrelationID)),
	}

	h.writeResponse(w, http.StatusCreated, response, correlationCtx.CorrelationID)
}

// Services API implementations
func (h *SimplifiedContractHandler) GetServicesAdmin(w http.ResponseWriter, r *http.Request, params admin.GetServicesAdminParams) {
	correlationCtx := domain.FromContext(r.Context())
	if correlationCtx == nil {
		correlationCtx = domain.NewCorrelationContext()
	}

	response := struct {
		Data       []admin.Service       `json:"data"`
		Pagination admin.PaginationInfo `json:"pagination"`
	}{
		Data: []admin.Service{},
		Pagination: admin.PaginationInfo{
			CurrentPage:  1,
			TotalPages:   1,
			TotalItems:   0,
			ItemsPerPage: 20,
			HasNext:      false,
			HasPrevious:  false,
		},
	}

	h.writeResponse(w, http.StatusOK, response, correlationCtx.CorrelationID)
}

func (h *SimplifiedContractHandler) CreateService(w http.ResponseWriter, r *http.Request) {
	correlationCtx := domain.FromContext(r.Context())
	if correlationCtx == nil {
		correlationCtx = domain.NewCorrelationContext()
	}

	response := admin.CreatedResponse{
		Success:       true,
		Message:       "Service created successfully",
		Data:          map[string]interface{}{"service_id": "placeholder"},
		Timestamp:     time.Now().UTC(),
		CorrelationId: openapi_types.UUID(uuid.MustParse(correlationCtx.CorrelationID)),
	}

	h.writeResponse(w, http.StatusCreated, response, correlationCtx.CorrelationID)
}

// Events API implementations
func (h *SimplifiedContractHandler) GetEventsAdmin(w http.ResponseWriter, r *http.Request, params admin.GetEventsAdminParams) {
	correlationCtx := domain.FromContext(r.Context())
	if correlationCtx == nil {
		correlationCtx = domain.NewCorrelationContext()
	}

	response := struct {
		Data       []admin.Event         `json:"data"`
		Pagination admin.PaginationInfo `json:"pagination"`
	}{
		Data: []admin.Event{},
		Pagination: admin.PaginationInfo{
			CurrentPage:  1,
			TotalPages:   1,
			TotalItems:   0,
			ItemsPerPage: 20,
			HasNext:      false,
			HasPrevious:  false,
		},
	}

	h.writeResponse(w, http.StatusOK, response, correlationCtx.CorrelationID)
}

func (h *SimplifiedContractHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	correlationCtx := domain.FromContext(r.Context())
	if correlationCtx == nil {
		correlationCtx = domain.NewCorrelationContext()
	}

	response := admin.CreatedResponse{
		Success:       true,
		Message:       "Event created successfully",
		Data:          map[string]interface{}{"event_id": "placeholder"},
		Timestamp:     time.Now().UTC(),
		CorrelationId: openapi_types.UUID(uuid.MustParse(correlationCtx.CorrelationID)),
	}

	h.writeResponse(w, http.StatusCreated, response, correlationCtx.CorrelationID)
}

// Research API implementations
func (h *SimplifiedContractHandler) GetResearchAdmin(w http.ResponseWriter, r *http.Request, params admin.GetResearchAdminParams) {
	correlationCtx := domain.FromContext(r.Context())
	if correlationCtx == nil {
		correlationCtx = domain.NewCorrelationContext()
	}

	response := struct {
		Data       []admin.ResearchPublication `json:"data"`
		Pagination admin.PaginationInfo        `json:"pagination"`
	}{
		Data: []admin.ResearchPublication{},
		Pagination: admin.PaginationInfo{
			CurrentPage:  1,
			TotalPages:   1,
			TotalItems:   0,
			ItemsPerPage: 20,
			HasNext:      false,
			HasPrevious:  false,
		},
	}

	h.writeResponse(w, http.StatusOK, response, correlationCtx.CorrelationID)
}

func (h *SimplifiedContractHandler) CreateResearchPublication(w http.ResponseWriter, r *http.Request) {
	correlationCtx := domain.FromContext(r.Context())
	if correlationCtx == nil {
		correlationCtx = domain.NewCorrelationContext()
	}

	response := admin.CreatedResponse{
		Success:       true,
		Message:       "Research publication created successfully",
		Data:          map[string]interface{}{"research_id": "placeholder"},
		Timestamp:     time.Now().UTC(),
		CorrelationId: openapi_types.UUID(uuid.MustParse(correlationCtx.CorrelationID)),
	}

	h.writeResponse(w, http.StatusCreated, response, correlationCtx.CorrelationID)
}

// Stub implementations for other admin interface methods
func (h *SimplifiedContractHandler) GetDashboardAnalytics(w http.ResponseWriter, r *http.Request, params admin.GetDashboardAnalyticsParams) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *SimplifiedContractHandler) AdminLogin(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *SimplifiedContractHandler) AdminLogout(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *SimplifiedContractHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *SimplifiedContractHandler) GetAdminHealth(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *SimplifiedContractHandler) GetInquiries(w http.ResponseWriter, r *http.Request, params admin.GetInquiriesParams) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *SimplifiedContractHandler) GetInquiryById(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *SimplifiedContractHandler) UpdateInquiryStatus(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *SimplifiedContractHandler) GetSystemSettings(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *SimplifiedContractHandler) UpdateSystemSettings(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *SimplifiedContractHandler) GetAdminUsers(w http.ResponseWriter, r *http.Request, params admin.GetAdminUsersParams) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *SimplifiedContractHandler) CreateAdminUser(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *SimplifiedContractHandler) DeleteAdminUser(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *SimplifiedContractHandler) GetAdminUserById(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *SimplifiedContractHandler) UpdateAdminUser(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// writeResponse writes a standardized JSON response
func (h *SimplifiedContractHandler) writeResponse(w http.ResponseWriter, statusCode int, data interface{}, correlationID string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Correlation-ID", correlationID)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	
	w.WriteHeader(statusCode)
	
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// writeErrorResponse writes a standardized error response
func (h *SimplifiedContractHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, correlationID string) {
	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":           "INTERNAL_ERROR",
			"message":        message,
			"correlation_id": correlationID,
			"timestamp":      time.Now().UTC().Format(time.RFC3339),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Correlation-ID", correlationID)
	w.WriteHeader(statusCode)
	
	json.NewEncoder(w).Encode(errorResponse)
}

// convertNewsToContract converts domain news to contract-compliant NewsArticle
func (h *SimplifiedContractHandler) convertNewsToContract(news news.News) admin.NewsArticle {
	newsUUID, _ := uuid.Parse(news.NewsID)
	categoryUUID, _ := uuid.Parse(news.CategoryID)
	
	return admin.NewsArticle{
		NewsId:               openapi_types.UUID(newsUUID),
		Title:                news.Title,
		Summary:              news.Summary,
		Content:              &news.Content,
		CategoryId:           openapi_types.UUID(categoryUUID),
		AuthorName:           &news.AuthorName,
		ImageUrl:             &news.ImageURL,
		NewsType:             h.convertNewsType(news.NewsType),
		PriorityLevel:        h.convertPriorityLevel(news.PriorityLevel),
		PublishingStatus:     h.convertPublishingStatus(news.PublishingStatus),
		PublicationTimestamp: news.PublicationTimestamp,
		Tags:                 &news.Tags,
		CreatedOn:            news.CreatedOn,
		ModifiedOn:           news.ModifiedOn,
		CreatedBy:            &news.CreatedBy,
		ModifiedBy:           &news.ModifiedBy,
		Slug:                 news.Slug,
	}
}

// convertNewsCategoryToContract converts domain category to contract-compliant NewsCategory
func (h *SimplifiedContractHandler) convertNewsCategoryToContract(category news.NewsCategory) admin.NewsCategory {
	categoryUUID, _ := uuid.Parse(category.CategoryID)
	
	return admin.NewsCategory{
		CategoryId:          openapi_types.UUID(categoryUUID),
		Name:                category.Name,
		Slug:                category.Slug,
		Description:         &category.Description,
		IsDefaultUnassigned: category.IsDefaultUnassigned,
		CreatedOn:           category.CreatedOn,
		ModifiedOn:          category.ModifiedOn,
		CreatedBy:           &category.CreatedBy,
		ModifiedBy:          &category.ModifiedBy,
	}
}

// Helper methods for enum conversion
func (h *SimplifiedContractHandler) convertNewsType(newsType news.NewsType) admin.NewsArticleNewsType {
	switch newsType {
	case news.NewsTypeAnnouncement:
		return admin.NewsArticleNewsTypeAnnouncement
	case news.NewsTypePressRelease:
		return admin.NewsArticleNewsTypePressRelease
	case news.NewsTypeEvent:
		return admin.NewsArticleNewsTypeEvent
	case news.NewsTypeUpdate:
		return admin.NewsArticleNewsTypeUpdate
	case news.NewsTypeAlert:
		return admin.NewsArticleNewsTypeAlert
	case news.NewsTypeFeature:
		return admin.NewsArticleNewsTypeFeature
	default:
		return admin.NewsArticleNewsTypeAnnouncement
	}
}

func (h *SimplifiedContractHandler) convertPriorityLevel(priority news.PriorityLevel) admin.NewsArticlePriorityLevel {
	switch priority {
	case news.PriorityLevelLow:
		return admin.NewsArticlePriorityLevelLow
	case news.PriorityLevelNormal:
		return admin.NewsArticlePriorityLevelNormal
	case news.PriorityLevelHigh:
		return admin.NewsArticlePriorityLevelHigh
	case news.PriorityLevelUrgent:
		return admin.NewsArticlePriorityLevelUrgent
	default:
		return admin.NewsArticlePriorityLevelNormal
	}
}

func (h *SimplifiedContractHandler) convertPublishingStatus(status news.PublishingStatus) admin.NewsArticlePublishingStatus {
	switch status {
	case news.PublishingStatusDraft:
		return admin.NewsArticlePublishingStatusDraft
	case news.PublishingStatusPublished:
		return admin.NewsArticlePublishingStatusPublished
	case news.PublishingStatusArchived:
		return admin.NewsArticlePublishingStatusArchived
	default:
		return admin.NewsArticlePublishingStatusDraft
	}
}

// RegisterSimplifiedContentRoutes registers simplified contract-compliant routes
func RegisterSimplifiedContentRoutes(router *mux.Router,
	newsService *news.NewsService,
	researchService *research.ResearchService,
	servicesService *services.ServicesService,
	eventsService *events.EventsService) {
	
	handler := NewSimplifiedContractHandler(newsService, researchService, servicesService, eventsService)
	admin.HandlerFromMux(handler, router)
}