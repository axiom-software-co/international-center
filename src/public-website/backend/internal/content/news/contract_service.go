package news

import (
	"context"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/google/uuid"
)

// Contract-compliant types for API integration
type ListNewsParams struct {
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	Search     string `json:"search,omitempty"`
	CategoryID string `json:"category_id,omitempty"`
	Status     string `json:"status,omitempty"`
}

type CreateNewsArticleRequest struct {
	Title            string    `json:"title"`
	Summary          string    `json:"summary"`
	Content          *string   `json:"content,omitempty"`
	CategoryID       string    `json:"category_id"`
	AuthorName       *string   `json:"author_name,omitempty"`
	ImageURL         *string   `json:"image_url,omitempty"`
	NewsType         string    `json:"news_type"`
	PriorityLevel    string    `json:"priority_level"`
	Tags             *[]string `json:"tags,omitempty"`
	PublishingStatus string    `json:"publishing_status"`
}

type UpdateNewsArticleRequest struct {
	Title            *string   `json:"title,omitempty"`
	Summary          *string   `json:"summary,omitempty"`
	Content          *string   `json:"content,omitempty"`
	CategoryID       *string   `json:"category_id,omitempty"`
	AuthorName       *string   `json:"author_name,omitempty"`
	ImageURL         *string   `json:"image_url,omitempty"`
	NewsType         *string   `json:"news_type,omitempty"`
	PriorityLevel    *string   `json:"priority_level,omitempty"`
	Tags             *[]string `json:"tags,omitempty"`
	PublishingStatus *string   `json:"publishing_status,omitempty"`
}

type CreateCategoryRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type PaginationResult struct {
	CurrentPage  int  `json:"current_page"`
	TotalPages   int  `json:"total_pages"`
	TotalItems   int  `json:"total_items"`
	ItemsPerPage int  `json:"items_per_page"`
	HasNext      bool `json:"has_next"`
	HasPrevious  bool `json:"has_previous"`
}

// Contract-compliant service methods for news operations

// AdminListNews lists news articles with contract-compliant parameters and pagination
func (s *NewsService) AdminListNews(ctx context.Context, params ListNewsParams) ([]*News, PaginationResult, error) {
	// Convert to internal status format
	var status PublishingStatus
	if params.Status != "" {
		switch params.Status {
		case "draft":
			status = PublishingStatusDraft
		case "published":
			status = PublishingStatusPublished
		case "archived":
			status = PublishingStatusArchived
		}
	}

	// Get news articles from repository
	var newsItems []*News
	var err error
	
	if params.Status != "" {
		newsItems, err = s.repository.GetNewsByPublishingStatus(ctx, status)
	} else if params.CategoryID != "" {
		newsItems, err = s.repository.GetNewsByCategory(ctx, params.CategoryID)
	} else if params.Search != "" {
		newsItems, err = s.repository.SearchNews(ctx, params.Search)
	} else {
		newsItems, err = s.repository.GetAllNews(ctx)
	}
	
	if err != nil {
		return nil, PaginationResult{}, err
	}

	// Apply pagination (simplified for demo - in production, implement proper DB pagination)  
	start := (params.Page - 1) * params.Limit
	end := start + params.Limit
	if start > len(newsItems) {
		start = len(newsItems)
	}
	if end > len(newsItems) {
		end = len(newsItems)
	}
	
	paginatedArticles := newsItems[start:end]
	totalItems := len(newsItems)
	totalPages := (totalItems + params.Limit - 1) / params.Limit
	if totalPages == 0 {
		totalPages = 1
	}
	
	pagination := PaginationResult{
		CurrentPage:  params.Page,
		TotalPages:   totalPages,
		TotalItems:   totalItems,
		ItemsPerPage: params.Limit,
		HasNext:      params.Page < totalPages,
		HasPrevious:  params.Page > 1,
	}

	return paginatedArticles, pagination, nil
}

// AdminCreateNews creates a new news article in a contract-compliant way  
func (s *NewsService) AdminCreateNews(ctx context.Context, request CreateNewsArticleRequest, userID string) (*News, error) {
	// Create internal news entity
	newsID := uuid.New().String()
	now := time.Now().UTC()
	
	// Generate slug from title
	slug := generateSlugFromTitle(request.Title)
	
	news := &News{
		NewsID:              newsID,
		Title:               request.Title,
		Summary:             request.Summary,
		Content:             getStringValue(request.Content),
		Slug:                slug,
		CategoryID:          request.CategoryID,
		AuthorName:          getStringValue(request.AuthorName),
		ImageURL:            getStringValue(request.ImageURL),
		NewsType:            NewsType(request.NewsType),
		PriorityLevel:       PriorityLevel(request.PriorityLevel),
		PublishingStatus:    PublishingStatus(request.PublishingStatus),
		PublicationTimestamp: now,
		Tags:                getStringSliceValue(request.Tags),
		CreatedOn:           now,
		CreatedBy:           userID,
	}

	// Save to repository
	err := s.repository.SaveNews(ctx, news)
	if err != nil {
		return nil, err
	}

	// Return domain entity directly
	return news, nil
}

// AdminGetNews retrieves a specific news article (admin only)
func (s *NewsService) AdminGetNews(ctx context.Context, newsID string, userID string) (*News, error) {
	news, err := s.repository.GetNews(ctx, newsID)
	if err != nil {
		return nil, err
	}

	if news == nil {
		return nil, domain.NewNotFoundError("news article", newsID)
	}

	// Return domain entity directly
	return news, nil
}

// AdminUpdateNews updates a news article in a contract-compliant way
func (s *NewsService) AdminUpdateNews(ctx context.Context, newsID string, request UpdateNewsArticleRequest, userID string) (*News, error) {
	// Get existing news article
	news, err := s.repository.GetNews(ctx, newsID)
	if err != nil {
		return nil, err
	}

	if news == nil {
		return nil, domain.NewNotFoundError("news article", newsID)
	}

	// Update fields that are provided
	now := time.Now().UTC()
	news.ModifiedOn = &now
	news.ModifiedBy = userID

	if request.Title != nil {
		news.Title = *request.Title
		news.Slug = generateSlugFromTitle(*request.Title) // Regenerate slug if title changes
	}
	if request.Summary != nil {
		news.Summary = *request.Summary
	}
	if request.Content != nil {
		news.Content = *request.Content
	}
	if request.CategoryID != nil {
		news.CategoryID = *request.CategoryID
	}
	if request.AuthorName != nil {
		news.AuthorName = *request.AuthorName
	}
	if request.ImageURL != nil {
		news.ImageURL = *request.ImageURL
	}
	if request.NewsType != nil {
		news.NewsType = NewsType(*request.NewsType)
	}
	if request.PriorityLevel != nil {
		news.PriorityLevel = PriorityLevel(*request.PriorityLevel)
	}
	if request.Tags != nil {
		news.Tags = *request.Tags
	}
	if request.PublishingStatus != nil {
		news.PublishingStatus = PublishingStatus(*request.PublishingStatus)
	}

	// Save updated news
	err = s.repository.SaveNews(ctx, news)
	if err != nil {
		return nil, err
	}

	// Return domain entity directly
	return news, nil
}

// AdminDeleteNews deletes a news article in a contract-compliant way
func (s *NewsService) AdminDeleteNews(ctx context.Context, newsID string, userID string) error {
	return s.repository.DeleteNews(ctx, newsID, userID)
}

// AdminPublishNews publishes a news article
func (s *NewsService) AdminPublishNews(ctx context.Context, newsID string, userID string) (*News, error) {
	// Get existing news article
	news, err := s.repository.GetNews(ctx, newsID)
	if err != nil {
		return nil, err
	}

	if news == nil {
		return nil, domain.NewNotFoundError("news article", newsID)
	}

	// Update publishing status
	now := time.Now().UTC()
	news.PublishingStatus = PublishingStatusPublished
	news.PublicationTimestamp = now
	news.ModifiedOn = &now
	news.ModifiedBy = userID

	// Save updated news
	err = s.repository.SaveNews(ctx, news)
	if err != nil {
		return nil, err
	}

	// Return domain entity directly
	return news, nil
}

// AdminUnpublishNews unpublishes a news article
func (s *NewsService) AdminUnpublishNews(ctx context.Context, newsID string, userID string) (*News, error) {
	// Get existing news article
	news, err := s.repository.GetNews(ctx, newsID)
	if err != nil {
		return nil, err
	}

	if news == nil {
		return nil, domain.NewNotFoundError("news article", newsID)
	}

	// Update publishing status
	now := time.Now().UTC()
	news.PublishingStatus = PublishingStatusDraft
	news.ModifiedOn = &now
	news.ModifiedBy = userID

	// Save updated news
	err = s.repository.SaveNews(ctx, news)
	if err != nil {
		return nil, err
	}

	// Return domain entity directly
	return news, nil
}

// AdminListCategories lists news categories for admin interface
func (s *NewsService) AdminListCategories(ctx context.Context, userID string) ([]*NewsCategory, error) {
	categories, err := s.repository.GetAllNewsCategories(ctx)
	if err != nil {
		return nil, err
	}

	// Return domain entities directly
	return categories, nil
}

// AdminCreateCategory creates a new news category
func (s *NewsService) AdminCreateCategory(ctx context.Context, request CreateCategoryRequest, userID string) (*NewsCategory, error) {
	// Create internal category entity
	categoryID := uuid.New().String()
	now := time.Now().UTC()
	
	// Generate slug from name
	slug := generateSlugFromTitle(request.Name)
	
	category := &NewsCategory{
		CategoryID:          categoryID,
		Name:                request.Name,
		Slug:                slug,
		Description:         getStringValue(request.Description),
		IsDefaultUnassigned: false, // New categories are not default
		CreatedOn:           now,
		CreatedBy:           userID,
	}

	// Save to repository
	err := s.repository.SaveNewsCategory(ctx, category)
	if err != nil {
		return nil, err
	}

	// Return domain entity directly
	return category, nil
}

// Helper functions

// getStringValue safely gets string value from pointer
func getStringValue(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

// getStringSliceValue safely gets string slice value from pointer
func getStringSliceValue(s *[]string) []string {
	if s != nil {
		return *s
	}
	return []string{}
}