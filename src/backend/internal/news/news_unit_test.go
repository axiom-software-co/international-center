package news

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	sharedtesting "github.com/axiom-software-co/international-center/src/backend/internal/shared/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockNewsRepository provides mock implementation for unit tests
type MockNewsRepository struct {
	news               map[string]*News
	categories         map[string]*NewsCategory  
	featuredNews       map[string]*FeaturedNews
	auditEvents        []MockAuditEvent
	failures           map[string]error
}

type MockAuditEvent struct {
	EntityType    domain.EntityType
	EntityID      string
	OperationType domain.AuditEventType
	UserID        string
	Before        interface{}
	After         interface{}
}

func NewMockNewsRepository() *MockNewsRepository {
	return &MockNewsRepository{
		news:               make(map[string]*News),
		categories:         make(map[string]*NewsCategory),
		featuredNews:       make(map[string]*FeaturedNews),
		auditEvents:        make([]MockAuditEvent, 0),
		failures:           make(map[string]error),
	}
}

// SetFailure sets a mock failure for specific operations
func (m *MockNewsRepository) SetFailure(operation string, err error) {
	m.failures[operation] = err
}

// GetAuditEvents returns all mock audit events
func (m *MockNewsRepository) GetAuditEvents() []MockAuditEvent {
	return m.auditEvents
}

// News repository methods - these will fail until implemented in GREEN phase
func (m *MockNewsRepository) SaveNews(ctx context.Context, news *News) error {
	if err, exists := m.failures["SaveNews"]; exists {
		return err
	}
	m.news[news.NewsID] = news
	return nil
}

func (m *MockNewsRepository) GetNews(ctx context.Context, newsID string) (*News, error) {
	if err, exists := m.failures["GetNews"]; exists {
		return nil, err
	}
	news, exists := m.news[newsID]
	if !exists || news.IsDeleted {
		return nil, domain.NewNotFoundError("news", newsID)
	}
	return news, nil
}

func (m *MockNewsRepository) GetNewsBySlug(ctx context.Context, slug string) (*News, error) {
	if err, exists := m.failures["GetNewsBySlug"]; exists {
		return nil, err
	}
	for _, news := range m.news {
		if news.Slug == slug && !news.IsDeleted {
			return news, nil
		}
	}
	return nil, domain.NewNotFoundError("news", slug)
}

func (m *MockNewsRepository) GetAllNews(ctx context.Context) ([]*News, error) {
	if err, exists := m.failures["GetAllNews"]; exists {
		return nil, err
	}
	newsList := make([]*News, 0)
	for _, news := range m.news {
		if !news.IsDeleted {
			newsList = append(newsList, news)
		}
	}
	return newsList, nil
}

func (m *MockNewsRepository) GetNewsByCategory(ctx context.Context, categoryID string) ([]*News, error) {
	if err, exists := m.failures["GetNewsByCategory"]; exists {
		return nil, err
	}
	var newsList []*News
	for _, news := range m.news {
		if news.CategoryID == categoryID && !news.IsDeleted {
			newsList = append(newsList, news)
		}
	}
	return newsList, nil
}

func (m *MockNewsRepository) GetNewsByPublishingStatus(ctx context.Context, status PublishingStatus) ([]*News, error) {
	if err, exists := m.failures["GetNewsByPublishingStatus"]; exists {
		return nil, err
	}
	var newsList []*News
	for _, news := range m.news {
		if news.PublishingStatus == status && !news.IsDeleted {
			newsList = append(newsList, news)
		}
	}
	return newsList, nil
}

func (m *MockNewsRepository) DeleteNews(ctx context.Context, newsID string, userID string) error {
	if err, exists := m.failures["DeleteNews"]; exists {
		return err
	}
	news, exists := m.news[newsID]
	if !exists {
		return domain.NewNotFoundError("news", newsID)
	}
	now := time.Now()
	news.IsDeleted = true
	news.DeletedOn = &now
	news.DeletedBy = userID
	return nil
}

func (m *MockNewsRepository) SearchNews(ctx context.Context, searchTerm string) ([]*News, error) {
	if err, exists := m.failures["SearchNews"]; exists {
		return nil, err
	}
	results := make([]*News, 0)
	for _, news := range m.news {
		if !news.IsDeleted {
			// Simple mock search - check title and summary
			if len(searchTerm) == 0 || 
				strings.Contains(strings.ToLower(news.Title), strings.ToLower(searchTerm)) ||
				strings.Contains(strings.ToLower(news.Summary), strings.ToLower(searchTerm)) {
				results = append(results, news)
			}
		}
	}
	return results, nil
}

// News category repository methods
func (m *MockNewsRepository) SaveNewsCategory(ctx context.Context, category *NewsCategory) error {
	if err, exists := m.failures["SaveNewsCategory"]; exists {
		return err
	}
	m.categories[category.CategoryID] = category
	return nil
}

func (m *MockNewsRepository) GetNewsCategory(ctx context.Context, categoryID string) (*NewsCategory, error) {
	if err, exists := m.failures["GetNewsCategory"]; exists {
		return nil, err
	}
	category, exists := m.categories[categoryID]
	if !exists || category.IsDeleted {
		return nil, domain.NewNotFoundError("news category", categoryID)
	}
	return category, nil
}

func (m *MockNewsRepository) GetNewsCategoryBySlug(ctx context.Context, slug string) (*NewsCategory, error) {
	if err, exists := m.failures["GetNewsCategoryBySlug"]; exists {
		return nil, err
	}
	for _, category := range m.categories {
		if category.Slug == slug && !category.IsDeleted {
			return category, nil
		}
	}
	return nil, domain.NewNotFoundError("news category", slug)
}

func (m *MockNewsRepository) GetAllNewsCategories(ctx context.Context) ([]*NewsCategory, error) {
	if err, exists := m.failures["GetAllNewsCategories"]; exists {
		return nil, err
	}
	var categories []*NewsCategory
	for _, category := range m.categories {
		if !category.IsDeleted {
			categories = append(categories, category)
		}
	}
	return categories, nil
}

func (m *MockNewsRepository) GetDefaultUnassignedCategory(ctx context.Context) (*NewsCategory, error) {
	if err, exists := m.failures["GetDefaultUnassignedCategory"]; exists {
		return nil, err
	}
	for _, category := range m.categories {
		if category.IsDefaultUnassigned && !category.IsDeleted {
			return category, nil
		}
	}
	return nil, domain.NewNotFoundError("default unassigned news category", "")
}

func (m *MockNewsRepository) DeleteNewsCategory(ctx context.Context, categoryID string, userID string) error {
	if err, exists := m.failures["DeleteNewsCategory"]; exists {
		return err
	}
	category, exists := m.categories[categoryID]
	if !exists {
		return domain.NewNotFoundError("news category", categoryID)
	}
	now := time.Now()
	category.IsDeleted = true
	category.DeletedOn = &now
	category.DeletedBy = userID
	return nil
}

// Featured news repository methods
func (m *MockNewsRepository) SaveFeaturedNews(ctx context.Context, featured *FeaturedNews) error {
	if err, exists := m.failures["SaveFeaturedNews"]; exists {
		return err
	}
	m.featuredNews[featured.FeaturedNewsID] = featured
	return nil
}

func (m *MockNewsRepository) GetFeaturedNews(ctx context.Context) (*FeaturedNews, error) {
	if err, exists := m.failures["GetFeaturedNews"]; exists {
		return nil, err
	}
	for _, featured := range m.featuredNews {
		return featured, nil // Only one featured news allowed
	}
	return nil, domain.NewNotFoundError("featured news", "")
}

func (m *MockNewsRepository) DeleteFeaturedNews(ctx context.Context, featuredNewsID string) error {
	if err, exists := m.failures["DeleteFeaturedNews"]; exists {
		return err
	}
	delete(m.featuredNews, featuredNewsID)
	return nil
}

// Audit repository methods - these will be implemented in Green phase
func (m *MockNewsRepository) GetNewsAudit(ctx context.Context, newsID string, userID string, limit int, offset int) ([]*domain.AuditEvent, error) {
	if err, exists := m.failures["GetNewsAudit"]; exists {
		return nil, err
	}
	// Return mock audit events for this news
	events := make([]*domain.AuditEvent, 0)
	for _, auditEvent := range m.auditEvents {
		if auditEvent.EntityID == newsID {
			events = append(events, &domain.AuditEvent{
				AuditID:       "audit-" + newsID + "-1",
				EntityType:    auditEvent.EntityType,
				EntityID:      auditEvent.EntityID,
				OperationType: auditEvent.OperationType,
				AuditTime:     time.Now().Add(-time.Hour),
				UserID:        auditEvent.UserID,
			})
		}
	}
	return events, nil
}

func (m *MockNewsRepository) GetNewsCategoryAudit(ctx context.Context, categoryID string, userID string, limit int, offset int) ([]*domain.AuditEvent, error) {
	if err, exists := m.failures["GetNewsCategoryAudit"]; exists {
		return nil, err
	}
	// Return mock audit events for this category
	var events []*domain.AuditEvent
	for _, auditEvent := range m.auditEvents {
		if auditEvent.EntityID == categoryID && auditEvent.EntityType == domain.EntityTypeCategory {
			events = append(events, &domain.AuditEvent{
				AuditID:       "audit-" + categoryID + "-1", 
				EntityType:    auditEvent.EntityType,
				EntityID:      auditEvent.EntityID,
				OperationType: auditEvent.OperationType,
				AuditTime:     time.Now().Add(-time.Hour),
				UserID:        auditEvent.UserID,
			})
		}
	}
	return events, nil
}

// Audit event publishing method
func (m *MockNewsRepository) PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error {
	if err, exists := m.failures["PublishAuditEvent"]; exists {
		return err
	}
	// Record the audit event in memory for testing
	m.auditEvents = append(m.auditEvents, MockAuditEvent{
		EntityType:    entityType,
		EntityID:      entityID,
		OperationType: operationType,
		UserID:        userID,
		Before:        beforeData,
		After:         afterData,
	})
	return nil
}

// Unit tests for News Service - RED PHASE (will fail until GREEN phase implementation)

func TestNewsService_GetNews(t *testing.T) {
	ctx, cancel := sharedtesting.CreateUnitTestContext()
		defer cancel()

	tests := []struct {
		name     string
		newsID   string
		userID   string
		setupFn  func(*MockNewsRepository)
		wantErr  bool
		errType  string
	}{
		{
			name:   "successfully retrieve existing news",
			newsID: "news-1", 
			userID: "user-1",
			setupFn: func(repo *MockNewsRepository) {
				repo.news["news-1"] = &News{
					NewsID:           "news-1",
					Title:            "Test News",
					Summary:          "Test summary",
					Slug:             "test-news",
					CategoryID:       "cat-1",
					PublishingStatus: PublishingStatusPublished,
					NewsType:         NewsTypeAnnouncement,
					PriorityLevel:    PriorityLevelNormal,
				}
			},
			wantErr: false,
		},
		{
			name:    "return not found error for non-existent news",
			newsID:  "non-existent",
			userID:  "user-1", 
			setupFn: func(repo *MockNewsRepository) {},
			wantErr: true,
			errType: "not_found",
		},
		{
			name:   "return not found error for soft deleted news", 
			newsID: "deleted-news",
			userID: "user-1",
			setupFn: func(repo *MockNewsRepository) {
				now := time.Now()
				repo.news["deleted-news"] = &News{
					NewsID:    "deleted-news",
					Title:     "Deleted News",
					Summary:   "Deleted summary",
					IsDeleted: true,
					DeletedOn: &now,
				}
			},
			wantErr: true,
			errType: "not_found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := NewMockNewsRepository()
			tt.setupFn(mockRepo)
			service := NewNewsService(mockRepo)

			// Act 
			result, err := service.GetNews(ctx, tt.newsID, tt.userID)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
				if tt.errType == "not_found" {
					assert.True(t, domain.IsNotFoundError(err))
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.newsID, result.NewsID)
			}
		})
	}
}

func TestNewsService_GetAllNews(t *testing.T) {
	ctx, cancel := sharedtesting.CreateUnitTestContext()
		defer cancel()

	tests := []struct {
		name       string
		userID     string
		setupFn    func(*MockNewsRepository)
		wantCount  int
		wantErr    bool
	}{
		{
			name:   "successfully retrieve all news",
			userID: "user-1",
			setupFn: func(repo *MockNewsRepository) {
				repo.news["news-1"] = &News{
					NewsID:           "news-1",
					Title:            "News 1",
					Summary:          "Summary 1", 
					PublishingStatus: PublishingStatusPublished,
				}
				repo.news["news-2"] = &News{
					NewsID:           "news-2",
					Title:            "News 2", 
					Summary:          "Summary 2",
					PublishingStatus: PublishingStatusDraft,
				}
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:   "return empty array when no news exists",
			userID: "user-1", 
			setupFn: func(repo *MockNewsRepository) {},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:   "exclude soft deleted news from results",
			userID: "user-1",
			setupFn: func(repo *MockNewsRepository) {
				now := time.Now()
				repo.news["news-1"] = &News{
					NewsID:           "news-1",
					Title:            "Active News",
					Summary:          "Active summary",
					PublishingStatus: PublishingStatusPublished,
				}
				repo.news["news-2"] = &News{
					NewsID:           "news-2", 
					Title:            "Deleted News",
					Summary:          "Deleted summary",
					PublishingStatus: PublishingStatusPublished,
					IsDeleted:        true,
					DeletedOn:        &now,
				}
			},
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := NewMockNewsRepository()
			tt.setupFn(mockRepo)
			service := NewNewsService(mockRepo)

			// Act
			result, err := service.GetAllNews(ctx, tt.userID)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, tt.wantCount)
			}
		})
	}
}

func TestNewsService_SearchNews(t *testing.T) {
	ctx, cancel := sharedtesting.CreateUnitTestContext()
		defer cancel()

	tests := []struct {
		name       string
		searchTerm string
		userID     string
		setupFn    func(*MockNewsRepository)
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "successfully search news by title",
			searchTerm: "breaking",
			userID:     "user-1",
			setupFn: func(repo *MockNewsRepository) {
				repo.news["news-1"] = &News{
					NewsID:  "news-1",
					Title:   "Breaking News Today",
					Summary: "Summary 1",
				}
				repo.news["news-2"] = &News{
					NewsID:  "news-2", 
					Title:   "Regular Update",
					Summary: "Summary 2",
				}
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "successfully search news by summary",
			searchTerm: "important",
			userID:     "user-1",
			setupFn: func(repo *MockNewsRepository) {
				repo.news["news-1"] = &News{
					NewsID:  "news-1",
					Title:   "News Title",
					Summary: "This is an important announcement",
				}
				repo.news["news-2"] = &News{
					NewsID:  "news-2",
					Title:   "Other News", 
					Summary: "Regular summary",
				}
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "return empty results when no matches found",
			searchTerm: "nonexistent",
			userID:     "user-1",
			setupFn: func(repo *MockNewsRepository) {
				repo.news["news-1"] = &News{
					NewsID:  "news-1",
					Title:   "News Title",
					Summary: "Summary",
				}
			},
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := NewMockNewsRepository()
			tt.setupFn(mockRepo)
			service := NewNewsService(mockRepo)

			// Act
			result, err := service.SearchNews(ctx, tt.searchTerm, tt.userID)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, tt.wantCount)
			}
		})
	}
}

func TestNewsService_GetFeaturedNews(t *testing.T) {
	ctx, cancel := sharedtesting.CreateUnitTestContext()
		defer cancel()

	tests := []struct {
		name    string
		userID  string
		setupFn func(*MockNewsRepository)
		wantErr bool
		errType string
	}{
		{
			name:   "successfully retrieve featured news",
			userID: "user-1",
			setupFn: func(repo *MockNewsRepository) {
				repo.news["news-1"] = &News{
					NewsID:           "news-1",
					Title:            "Featured News",
					Summary:          "Featured summary",
					PublishingStatus: PublishingStatusPublished,
				}
				repo.featuredNews["featured-1"] = &FeaturedNews{
					FeaturedNewsID: "featured-1",
					NewsID:         "news-1",
				}
			},
			wantErr: false,
		},
		{
			name:   "return not found when no featured news exists",
			userID: "user-1",
			setupFn: func(repo *MockNewsRepository) {},
			wantErr: true,
			errType: "not_found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := NewMockNewsRepository()
			tt.setupFn(mockRepo) 
			service := NewNewsService(mockRepo)

			// Act
			result, err := service.GetFeaturedNews(ctx, tt.userID)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
				if tt.errType == "not_found" {
					assert.True(t, domain.IsNotFoundError(err))
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// Admin audit functionality tests

func TestNewsService_GetNewsAudit(t *testing.T) {
	ctx, cancel := sharedtesting.CreateUnitTestContext()
		defer cancel()

	tests := []struct {
		name     string
		newsID   string
		userID   string
		limit    int
		offset   int
		setupFn  func(*MockNewsRepository)
		wantErr  bool
		errType  string
	}{
		{
			name:   "successfully retrieve news audit events",
			newsID: "news-1",
			userID: "admin-1",
			limit:  10,
			offset: 0,
			setupFn: func(repo *MockNewsRepository) {
				// Add audit event
				repo.auditEvents = append(repo.auditEvents, MockAuditEvent{
					EntityType:    domain.EntityType("news"),
					EntityID:      "news-1",
					OperationType: domain.AuditEventInsert,
					UserID:        "admin-1",
				})
			},
			wantErr: false,
		},
		{
			name:   "return empty array for news with no audit events",
			newsID: "news-without-audit",
			userID: "admin-1", 
			limit:  10,
			offset: 0,
			setupFn: func(repo *MockNewsRepository) {},
			wantErr: false,
		},
		{
			name:   "return unauthorized error for non-admin user",
			newsID: "news-1",
			userID: "",
			limit:  10,
			offset: 0,
			setupFn: func(repo *MockNewsRepository) {},
			wantErr: true,
			errType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange  
			mockRepo := NewMockNewsRepository()
			tt.setupFn(mockRepo)
			service := NewNewsService(mockRepo)

			// Act - this will fail until we implement GetNewsAudit method
			result, err := service.GetNewsAudit(ctx, tt.newsID, tt.userID, tt.limit, tt.offset)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
				if tt.errType == "unauthorized" {
					assert.True(t, domain.IsUnauthorizedError(err))
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestNewsService_GetNewsCategoryAudit(t *testing.T) {
	ctx, cancel := sharedtesting.CreateUnitTestContext()
		defer cancel()

	tests := []struct {
		name       string
		categoryID string
		userID     string
		limit      int
		offset     int
		setupFn    func(*MockNewsRepository)
		wantErr    bool
		errType    string
	}{
		{
			name:       "successfully retrieve news category audit events", 
			categoryID: "cat-1",
			userID:     "admin-1",
			limit:      10,
			offset:     0,
			setupFn: func(repo *MockNewsRepository) {
				// Add audit event for category
				repo.auditEvents = append(repo.auditEvents, MockAuditEvent{
					EntityType:    domain.EntityTypeCategory,
					EntityID:      "cat-1", 
					OperationType: domain.AuditEventUpdate,
					UserID:        "admin-1",
				})
			},
			wantErr: false,
		},
		{
			name:       "return unauthorized error for non-admin user",
			categoryID: "cat-1",
			userID:     "",
			limit:      10,
			offset:     0,
			setupFn:    func(repo *MockNewsRepository) {},
			wantErr:    true,
			errType:    "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := NewMockNewsRepository()
			tt.setupFn(mockRepo)
			service := NewNewsService(mockRepo)

			// Act - this will fail until we implement GetNewsCategoryAudit method  
			result, err := service.GetNewsCategoryAudit(ctx, tt.categoryID, tt.userID, tt.limit, tt.offset)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
				if tt.errType == "unauthorized" {
					assert.True(t, domain.IsUnauthorizedError(err))
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}