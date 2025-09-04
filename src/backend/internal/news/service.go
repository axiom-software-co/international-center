package news

import (
	"context"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
)

// NewsRepositoryInterface defines the contract for news data access
type NewsRepositoryInterface interface {
	// News operations
	SaveNews(ctx context.Context, news *News) error
	GetNews(ctx context.Context, newsID string) (*News, error)
	GetNewsBySlug(ctx context.Context, slug string) (*News, error)
	GetAllNews(ctx context.Context) ([]*News, error)
	GetNewsByCategory(ctx context.Context, categoryID string) ([]*News, error)
	GetNewsByPublishingStatus(ctx context.Context, status PublishingStatus) ([]*News, error)
	DeleteNews(ctx context.Context, newsID string, userID string) error
	SearchNews(ctx context.Context, searchTerm string) ([]*News, error)

	// News category operations
	SaveNewsCategory(ctx context.Context, category *NewsCategory) error
	GetNewsCategory(ctx context.Context, categoryID string) (*NewsCategory, error)
	GetNewsCategoryBySlug(ctx context.Context, slug string) (*NewsCategory, error)
	GetAllNewsCategories(ctx context.Context) ([]*NewsCategory, error)
	GetDefaultUnassignedCategory(ctx context.Context) (*NewsCategory, error)
	DeleteNewsCategory(ctx context.Context, categoryID string, userID string) error

	// Featured news operations
	SaveFeaturedNews(ctx context.Context, featured *FeaturedNews) error
	GetFeaturedNews(ctx context.Context) (*FeaturedNews, error)
	DeleteFeaturedNews(ctx context.Context, featuredNewsID string) error

	// Audit operations
	PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error
	GetNewsAudit(ctx context.Context, newsID string, userID string, limit int, offset int) ([]*domain.AuditEvent, error)
	GetNewsCategoryAudit(ctx context.Context, categoryID string, userID string, limit int, offset int) ([]*domain.AuditEvent, error)
}

// NewsService implements business logic for news operations
type NewsService struct {
	repository NewsRepositoryInterface
}

// NewNewsService creates a new news service
func NewNewsService(repository NewsRepositoryInterface) *NewsService {
	return &NewsService{
		repository: repository,
	}
}

// GetNews retrieves news by ID
func (s *NewsService) GetNews(ctx context.Context, newsID string, userID string) (*News, error) {
	if newsID == "" {
		return nil, domain.NewValidationError("news ID cannot be empty")
	}

	news, err := s.repository.GetNews(ctx, newsID)
	if err != nil {
		// Don't wrap domain errors that are already properly categorized
		if domain.IsNotFoundError(err) {
			return nil, err
		}
		return nil, domain.WrapError(err, "failed to get news")
	}

	return news, nil
}

// GetNewsBySlug retrieves news by slug
func (s *NewsService) GetNewsBySlug(ctx context.Context, slug string, userID string) (*News, error) {
	if slug == "" {
		return nil, domain.NewValidationError("slug cannot be empty")
	}

	news, err := s.repository.GetNewsBySlug(ctx, slug)
	if err != nil {
		if domain.IsNotFoundError(err) {
			return nil, err
		}
		return nil, domain.WrapError(err, "failed to get news by slug")
	}

	return news, nil
}

// GetAllNews retrieves all news
func (s *NewsService) GetAllNews(ctx context.Context, userID string) ([]*News, error) {
	newsList, err := s.repository.GetAllNews(ctx)
	if err != nil {
		return nil, domain.WrapError(err, "failed to get all news")
	}

	return newsList, nil
}

// GetNewsByCategory retrieves news by category
func (s *NewsService) GetNewsByCategory(ctx context.Context, categoryID string, userID string) ([]*News, error) {
	if categoryID == "" {
		return nil, domain.NewValidationError("category ID cannot be empty")
	}

	newsList, err := s.repository.GetNewsByCategory(ctx, categoryID)
	if err != nil {
		return nil, domain.WrapError(err, "failed to get news by category")
	}

	return newsList, nil
}

// SearchNews searches for news based on search term
func (s *NewsService) SearchNews(ctx context.Context, searchTerm string, userID string) ([]*News, error) {
	results, err := s.repository.SearchNews(ctx, searchTerm)
	if err != nil {
		return nil, domain.WrapError(err, "failed to search news")
	}

	return results, nil
}

// GetPublishedNews retrieves only published news
func (s *NewsService) GetPublishedNews(ctx context.Context, userID string) ([]*News, error) {
	publishedNews, err := s.repository.GetNewsByPublishingStatus(ctx, PublishingStatusPublished)
	if err != nil {
		return nil, domain.WrapError(err, "failed to get published news")
	}

	return publishedNews, nil
}

// GetFeaturedNews retrieves the featured news
func (s *NewsService) GetFeaturedNews(ctx context.Context, userID string) (*News, error) {
	featured, err := s.repository.GetFeaturedNews(ctx)
	if err != nil {
		if domain.IsNotFoundError(err) {
			return nil, err
		}
		return nil, domain.WrapError(err, "failed to get featured news")
	}

	// Get the actual news article
	news, err := s.repository.GetNews(ctx, featured.NewsID)
	if err != nil {
		if domain.IsNotFoundError(err) {
			return nil, err
		}
		return nil, domain.WrapError(err, "failed to get featured news article")
	}

	return news, nil
}

// News Category operations

// GetNewsCategory retrieves news category by ID
func (s *NewsService) GetNewsCategory(ctx context.Context, categoryID string, userID string) (*NewsCategory, error) {
	if categoryID == "" {
		return nil, domain.NewValidationError("category ID cannot be empty")
	}

	category, err := s.repository.GetNewsCategory(ctx, categoryID)
	if err != nil {
		if domain.IsNotFoundError(err) {
			return nil, err
		}
		return nil, domain.WrapError(err, "failed to get news category")
	}

	return category, nil
}

// GetNewsCategoryBySlug retrieves news category by slug
func (s *NewsService) GetNewsCategoryBySlug(ctx context.Context, slug string, userID string) (*NewsCategory, error) {
	if slug == "" {
		return nil, domain.NewValidationError("slug cannot be empty")
	}

	category, err := s.repository.GetNewsCategoryBySlug(ctx, slug)
	if err != nil {
		if domain.IsNotFoundError(err) {
			return nil, err
		}
		return nil, domain.WrapError(err, "failed to get news category by slug")
	}

	return category, nil
}

// GetAllNewsCategories retrieves all news categories
func (s *NewsService) GetAllNewsCategories(ctx context.Context, userID string) ([]*NewsCategory, error) {
	categories, err := s.repository.GetAllNewsCategories(ctx)
	if err != nil {
		return nil, domain.WrapError(err, "failed to get all news categories")
	}

	return categories, nil
}

// Admin-only operations

// GetNewsAudit retrieves audit trail for news (admin only)
func (s *NewsService) GetNewsAudit(ctx context.Context, newsID string, userID string, limit int, offset int) ([]*domain.AuditEvent, error) {
	// Validate admin authentication
	if userID == "" {
		return nil, domain.NewUnauthorizedError("admin authentication required")
	}

	if newsID == "" {
		return nil, domain.NewValidationError("news ID cannot be empty")
	}

	// Get audit events from repository
	auditEvents, err := s.repository.GetNewsAudit(ctx, newsID, userID, limit, offset)
	if err != nil {
		return nil, domain.WrapError(err, "failed to get audit trail for news " + newsID)
	}

	return auditEvents, nil
}

// GetNewsCategoryAudit retrieves audit trail for news category (admin only)
func (s *NewsService) GetNewsCategoryAudit(ctx context.Context, categoryID string, userID string, limit int, offset int) ([]*domain.AuditEvent, error) {
	// Validate admin authentication
	if userID == "" {
		return nil, domain.NewUnauthorizedError("admin authentication required")
	}

	if categoryID == "" {
		return nil, domain.NewValidationError("category ID cannot be empty")
	}

	// Get audit events from repository
	auditEvents, err := s.repository.GetNewsCategoryAudit(ctx, categoryID, userID, limit, offset)
	if err != nil {
		return nil, domain.WrapError(err, "failed to get audit trail for news category " + categoryID)
	}

	return auditEvents, nil
}