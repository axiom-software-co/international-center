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

// Domain Validation Method Tests (TDD RED Phase)

func TestNewsType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		newsType NewsType
		want     bool
	}{
		{"valid announcement", NewsTypeAnnouncement, true},
		{"valid press release", NewsTypePressRelease, true},
		{"valid event", NewsTypeEvent, true},
		{"valid update", NewsTypeUpdate, true},
		{"valid alert", NewsTypeAlert, true},
		{"valid feature", NewsTypeFeature, true},
		{"invalid news type", NewsType("invalid"), false},
		{"empty news type", NewsType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.newsType.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPriorityLevel_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		priority PriorityLevel
		want     bool
	}{
		{"valid low priority", PriorityLevelLow, true},
		{"valid normal priority", PriorityLevelNormal, true},
		{"valid high priority", PriorityLevelHigh, true},
		{"valid urgent priority", PriorityLevelUrgent, true},
		{"invalid priority", PriorityLevel("invalid"), false},
		{"empty priority", PriorityLevel(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.priority.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPublishingStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status PublishingStatus
		want   bool
	}{
		{"valid draft status", PublishingStatusDraft, true},
		{"valid published status", PublishingStatusPublished, true},
		{"valid archived status", PublishingStatusArchived, true},
		{"invalid status", PublishingStatus("invalid"), false},
		{"empty status", PublishingStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewNews(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		summary   string
		newsType  NewsType
		categoryID string
		userID    string
		wantErr   bool
		checkFunc func(*testing.T, *News)
	}{
		{
			name:      "successfully create news with valid data",
			title:     "Breaking News: Medical Center Opens",
			summary:   "Our new medical center is now open to serve the community with advanced healthcare services.",
			newsType:  NewsTypeAnnouncement,
			categoryID: "category-123",
			userID:    "admin-user",
			wantErr:   false,
			checkFunc: func(t *testing.T, news *News) {
				assert.NotEmpty(t, news.NewsID)
				assert.Equal(t, "Breaking News: Medical Center Opens", news.Title)
				assert.Equal(t, NewsTypeAnnouncement, news.NewsType)
				assert.Equal(t, PublishingStatusDraft, news.PublishingStatus)
				assert.Equal(t, PriorityLevelNormal, news.PriorityLevel)
				assert.False(t, news.IsDeleted)
			},
		},
		{
			name:      "return validation error for empty title",
			title:     "",
			summary:   "Valid summary content",
			newsType:  NewsTypeAnnouncement,
			categoryID: "category-123",
			userID:    "admin-user",
			wantErr:   true,
		},
		{
			name:      "return validation error for empty summary",
			title:     "Valid Title",
			summary:   "",
			newsType:  NewsTypeAnnouncement,
			categoryID: "category-123",
			userID:    "admin-user",
			wantErr:   true,
		},
		{
			name:      "return validation error for invalid news type",
			title:     "Valid Title",
			summary:   "Valid summary content",
			newsType:  NewsType("invalid"),
			categoryID: "category-123",
			userID:    "admin-user",
			wantErr:   true,
		},
		{
			name:      "return validation error for empty category ID",
			title:     "Valid Title",
			summary:   "Valid summary content",
			newsType:  NewsTypeAnnouncement,
			categoryID: "",
			userID:    "admin-user",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewNews(tt.title, tt.summary, tt.newsType, tt.categoryID, tt.userID)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

func TestNewNewsCategory(t *testing.T) {
	tests := []struct {
		name    string
		name_param string
		slug    string
		isDefaultUnassigned bool
		userID  string
		wantErr bool
		checkFunc func(*testing.T, *NewsCategory)
	}{
		{
			name: "successfully create news category with valid data",
			name_param: "Health Updates",
			slug: "health-updates",
			isDefaultUnassigned: false,
			userID: "admin-user",
			wantErr: false,
			checkFunc: func(t *testing.T, category *NewsCategory) {
				assert.NotEmpty(t, category.CategoryID)
				assert.Equal(t, "Health Updates", category.Name)
				assert.Equal(t, "health-updates", category.Slug)
				assert.False(t, category.IsDefaultUnassigned)
				assert.False(t, category.IsDeleted)
			},
		},
		{
			name: "return validation error for empty name",
			name_param: "",
			slug: "valid-slug",
			isDefaultUnassigned: false,
			userID: "admin-user",
			wantErr: true,
		},
		{
			name: "return validation error for empty slug",
			name_param: "Valid Name",
			slug: "",
			isDefaultUnassigned: false,
			userID: "admin-user",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewNewsCategory(tt.name_param, tt.slug, tt.isDefaultUnassigned, tt.userID)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

func TestNewFeaturedNews(t *testing.T) {
	tests := []struct {
		name    string
		newsID  string
		userID  string
		wantErr bool
		checkFunc func(*testing.T, *FeaturedNews)
	}{
		{
			name: "successfully create featured news with valid data",
			newsID: "news-123",
			userID: "admin-user",
			wantErr: false,
			checkFunc: func(t *testing.T, featured *FeaturedNews) {
				assert.NotEmpty(t, featured.FeaturedNewsID)
				assert.Equal(t, "news-123", featured.NewsID)
			},
		},
		{
			name: "return validation error for empty news ID",
			newsID: "",
			userID: "admin-user",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewFeaturedNews(tt.newsID, tt.userID)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

func TestNews_Publish(t *testing.T) {
	tests := []struct {
		name        string
		initialStatus PublishingStatus
		userID      string
		wantErr     bool
	}{
		{
			name: "successfully publish news from draft status",
			initialStatus: PublishingStatusDraft,
			userID: "admin-user",
			wantErr: false,
		},
		{
			name: "return error when trying to publish already published news",
			initialStatus: PublishingStatusPublished,
			userID: "admin-user",
			wantErr: true,
		},
		{
			name: "return error when trying to publish archived news",
			initialStatus: PublishingStatusArchived,
			userID: "admin-user",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			news := &News{
				NewsID: "news-123",
				Title: "Test News",
				Summary: "Test summary",
				PublishingStatus: tt.initialStatus,
			}

			err := news.Publish(tt.userID)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, PublishingStatusPublished, news.PublishingStatus)
				assert.NotNil(t, news.ModifiedOn)
				assert.Equal(t, tt.userID, news.ModifiedBy)
			}
		})
	}
}

func TestNews_Archive(t *testing.T) {
	tests := []struct {
		name        string
		initialStatus PublishingStatus
		userID      string
		wantErr     bool
	}{
		{
			name: "successfully archive published news",
			initialStatus: PublishingStatusPublished,
			userID: "admin-user",
			wantErr: false,
		},
		{
			name: "return error when trying to archive draft news",
			initialStatus: PublishingStatusDraft,
			userID: "admin-user",
			wantErr: true,
		},
		{
			name: "return error when trying to archive already archived news",
			initialStatus: PublishingStatusArchived,
			userID: "admin-user",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			news := &News{
				NewsID: "news-123",
				Title: "Test News",
				Summary: "Test summary",
				PublishingStatus: tt.initialStatus,
			}

			err := news.Archive(tt.userID)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, PublishingStatusArchived, news.PublishingStatus)
				assert.NotNil(t, news.ModifiedOn)
				assert.Equal(t, tt.userID, news.ModifiedBy)
			}
		})
	}
}

func TestNews_UnArchive(t *testing.T) {
	tests := []struct {
		name        string
		initialStatus PublishingStatus
		userID      string
		wantErr     bool
	}{
		{
			name: "successfully unarchive archived news",
			initialStatus: PublishingStatusArchived,
			userID: "admin-user",
			wantErr: false,
		},
		{
			name: "return error when trying to unarchive draft news",
			initialStatus: PublishingStatusDraft,
			userID: "admin-user",
			wantErr: true,
		},
		{
			name: "return error when trying to unarchive published news",
			initialStatus: PublishingStatusPublished,
			userID: "admin-user",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			news := &News{
				NewsID: "news-123",
				Title: "Test News",
				Summary: "Test summary",
				PublishingStatus: tt.initialStatus,
			}

			err := news.UnArchive(tt.userID)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, PublishingStatusDraft, news.PublishingStatus)
				assert.NotNil(t, news.ModifiedOn)
				assert.Equal(t, tt.userID, news.ModifiedBy)
			}
		})
	}
}

func TestNews_UpdateDetails(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		summary string
		userID  string
		wantErr bool
	}{
		{
			name: "successfully update news details",
			title: "Updated News Title",
			summary: "Updated news summary content",
			userID: "admin-user",
			wantErr: false,
		},
		{
			name: "return validation error for empty title",
			title: "",
			summary: "Valid summary",
			userID: "admin-user",
			wantErr: true,
		},
		{
			name: "return validation error for empty summary",
			title: "Valid Title",
			summary: "",
			userID: "admin-user",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			news := &News{
				NewsID: "news-123",
				Title: "Original Title",
				Summary: "Original summary",
				PublishingStatus: PublishingStatusDraft,
			}

			err := news.UpdateDetails(tt.title, tt.summary, tt.userID)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.title, news.Title)
				assert.Equal(t, tt.summary, news.Summary)
				assert.NotNil(t, news.ModifiedOn)
				assert.Equal(t, tt.userID, news.ModifiedBy)
			}
		})
	}
}

func TestNews_ValidateComprehensive(t *testing.T) {
	tests := []struct {
		name        string
		news        *News
		wantErr     bool
		wantMsgContains string
	}{
		{
			name: "valid news entity",
			news: &News{
				NewsID:              "550e8400-e29b-41d4-a716-446655440001",
				Title:               "Valid News Title",
				Summary:             "Valid news summary content",
				Slug:                "valid-news-title",
				CategoryID:          "550e8400-e29b-41d4-a716-446655440002",
				NewsType:            NewsTypeAnnouncement,
				PriorityLevel:       PriorityLevelNormal,
				PublishingStatus:    PublishingStatusDraft,
				PublicationTimestamp: time.Now(),
				CreatedOn:           time.Now(),
				CreatedBy:           "system",
			},
			wantErr: false,
		},
		{
			name: "invalid news with empty title",
			news: &News{
				NewsID:              "550e8400-e29b-41d4-a716-446655440001",
				Title:               "",
				Summary:             "Valid summary",
				NewsType:            NewsTypeAnnouncement,
			},
			wantErr:             true,
			wantMsgContains:     "title",
		},
		{
			name: "invalid news with invalid news type",
			news: &News{
				NewsID:              "550e8400-e29b-41d4-a716-446655440001",
				Title:               "Valid Title",
				Summary:             "Valid news summary content",
				Slug:                "valid-title",
				CategoryID:          "550e8400-e29b-41d4-a716-446655440002",
				NewsType:            NewsType("invalid"),
				PriorityLevel:       PriorityLevelNormal,
				PublishingStatus:    PublishingStatusDraft,
				CreatedBy:           "system",
			},
			wantErr:             true,
			wantMsgContains:     "news type",
		},
		{
			name: "invalid news with invalid publishing status",
			news: &News{
				NewsID:              "550e8400-e29b-41d4-a716-446655440001",
				Title:               "Valid Title",
				Summary:             "Valid news summary content",
				Slug:                "valid-title",
				CategoryID:          "550e8400-e29b-41d4-a716-446655440002",
				NewsType:            NewsTypeAnnouncement,
				PriorityLevel:       PriorityLevelNormal,
				PublishingStatus:    PublishingStatus("invalid"),
				CreatedBy:           "system",
			},
			wantErr:             true,
			wantMsgContains:     "publishing status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.news.ValidateComprehensive()
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantMsgContains != "" {
					assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.wantMsgContains))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNews_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus PublishingStatus
		targetStatus  PublishingStatus
		wantErr       bool
	}{
		{
			name: "can transition from draft to published",
			currentStatus: PublishingStatusDraft,
			targetStatus: PublishingStatusPublished,
			wantErr: false,
		},
		{
			name: "can transition from published to archived",
			currentStatus: PublishingStatusPublished,
			targetStatus: PublishingStatusArchived,
			wantErr: false,
		},
		{
			name: "can transition from archived to draft",
			currentStatus: PublishingStatusArchived,
			targetStatus: PublishingStatusDraft,
			wantErr: false,
		},
		{
			name: "cannot transition from draft to archived",
			currentStatus: PublishingStatusDraft,
			targetStatus: PublishingStatusArchived,
			wantErr: true,
		},
		{
			name: "cannot transition from published to draft",
			currentStatus: PublishingStatusPublished,
			targetStatus: PublishingStatusDraft,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			news := &News{
				NewsID: "news-123",
				PublishingStatus: tt.currentStatus,
			}
			
			err := news.CanTransitionTo(tt.targetStatus)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

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
		if auditEvent.EntityID == categoryID && auditEvent.EntityType == domain.EntityTypeNewsCategory {
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
			newsID: "550e8400-e29b-41d4-a716-446655440001", 
			userID: "550e8400-e29b-41d4-a716-446655440004",
			setupFn: func(repo *MockNewsRepository) {
				repo.news["550e8400-e29b-41d4-a716-446655440001"] = &News{
					NewsID:           "550e8400-e29b-41d4-a716-446655440001",
					Title:            "Test News",
					Summary:          "Test summary",
					Slug:             "test-news",
					CategoryID:       "550e8400-e29b-41d4-a716-446655440002",
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
			userID:  "550e8400-e29b-41d4-a716-446655440004", 
			setupFn: func(repo *MockNewsRepository) {},
			wantErr: true,
			errType: "not_found",
		},
		{
			name:   "return not found error for soft deleted news", 
			newsID: "deleted-news",
			userID: "550e8400-e29b-41d4-a716-446655440004",
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
			userID: "550e8400-e29b-41d4-a716-446655440004",
			setupFn: func(repo *MockNewsRepository) {
				repo.news["550e8400-e29b-41d4-a716-446655440001"] = &News{
					NewsID:           "550e8400-e29b-41d4-a716-446655440001",
					Title:            "News 1",
					Summary:          "Summary 1", 
					PublishingStatus: PublishingStatusPublished,
				}
				repo.news["550e8400-e29b-41d4-a716-446655440006"] = &News{
					NewsID:           "550e8400-e29b-41d4-a716-446655440006",
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
			userID: "550e8400-e29b-41d4-a716-446655440004", 
			setupFn: func(repo *MockNewsRepository) {},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:   "exclude soft deleted news from results",
			userID: "550e8400-e29b-41d4-a716-446655440004",
			setupFn: func(repo *MockNewsRepository) {
				now := time.Now()
				repo.news["550e8400-e29b-41d4-a716-446655440001"] = &News{
					NewsID:           "550e8400-e29b-41d4-a716-446655440001",
					Title:            "Active News",
					Summary:          "Active summary",
					PublishingStatus: PublishingStatusPublished,
				}
				repo.news["550e8400-e29b-41d4-a716-446655440006"] = &News{
					NewsID:           "550e8400-e29b-41d4-a716-446655440006", 
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
			userID:     "550e8400-e29b-41d4-a716-446655440004",
			setupFn: func(repo *MockNewsRepository) {
				repo.news["550e8400-e29b-41d4-a716-446655440001"] = &News{
					NewsID:  "550e8400-e29b-41d4-a716-446655440001",
					Title:   "Breaking News Today",
					Summary: "Summary 1",
				}
				repo.news["550e8400-e29b-41d4-a716-446655440006"] = &News{
					NewsID:  "550e8400-e29b-41d4-a716-446655440006", 
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
			userID:     "550e8400-e29b-41d4-a716-446655440004",
			setupFn: func(repo *MockNewsRepository) {
				repo.news["550e8400-e29b-41d4-a716-446655440001"] = &News{
					NewsID:  "550e8400-e29b-41d4-a716-446655440001",
					Title:   "News Title",
					Summary: "This is an important announcement",
				}
				repo.news["550e8400-e29b-41d4-a716-446655440006"] = &News{
					NewsID:  "550e8400-e29b-41d4-a716-446655440006",
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
			userID:     "550e8400-e29b-41d4-a716-446655440004",
			setupFn: func(repo *MockNewsRepository) {
				repo.news["550e8400-e29b-41d4-a716-446655440001"] = &News{
					NewsID:  "550e8400-e29b-41d4-a716-446655440001",
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
			userID: "550e8400-e29b-41d4-a716-446655440004",
			setupFn: func(repo *MockNewsRepository) {
				repo.news["550e8400-e29b-41d4-a716-446655440001"] = &News{
					NewsID:           "550e8400-e29b-41d4-a716-446655440001",
					Title:            "Featured News",
					Summary:          "Featured summary",
					PublishingStatus: PublishingStatusPublished,
				}
				repo.featuredNews["550e8400-e29b-41d4-a716-446655440005"] = &FeaturedNews{
					FeaturedNewsID: "550e8400-e29b-41d4-a716-446655440005",
					NewsID:         "550e8400-e29b-41d4-a716-446655440001",
				}
			},
			wantErr: false,
		},
		{
			name:   "return not found when no featured news exists",
			userID: "550e8400-e29b-41d4-a716-446655440004",
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
			newsID: "550e8400-e29b-41d4-a716-446655440001",
			userID: "admin-550e8400-e29b-41d4-a716-446655440003",
			limit:  10,
			offset: 0,
			setupFn: func(repo *MockNewsRepository) {
				// Add audit event
				repo.auditEvents = append(repo.auditEvents, MockAuditEvent{
					EntityType:    domain.EntityType("news"),
					EntityID:      "550e8400-e29b-41d4-a716-446655440001",
					OperationType: domain.AuditEventInsert,
					UserID:        "admin-550e8400-e29b-41d4-a716-446655440003",
				})
			},
			wantErr: false,
		},
		{
			name:   "return empty array for news with no audit events",
			newsID: "news-without-audit",
			userID: "admin-550e8400-e29b-41d4-a716-446655440003", 
			limit:  10,
			offset: 0,
			setupFn: func(repo *MockNewsRepository) {},
			wantErr: false,
		},
		{
			name:   "return unauthorized error for non-admin user",
			newsID: "550e8400-e29b-41d4-a716-446655440001",
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
			categoryID: "550e8400-e29b-41d4-a716-446655440002",
			userID:     "admin-550e8400-e29b-41d4-a716-446655440003",
			limit:      10,
			offset:     0,
			setupFn: func(repo *MockNewsRepository) {
				// Add audit event for category
				repo.auditEvents = append(repo.auditEvents, MockAuditEvent{
					EntityType:    domain.EntityTypeNewsCategory,
					EntityID:      "550e8400-e29b-41d4-a716-446655440002", 
					OperationType: domain.AuditEventUpdate,
					UserID:        "admin-550e8400-e29b-41d4-a716-446655440003",
				})
			},
			wantErr: false,
		},
		{
			name:       "return unauthorized error for non-admin user",
			categoryID: "550e8400-e29b-41d4-a716-446655440002",
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

// Admin CRUD Operations Tests

func TestNewsService_CreateNews(t *testing.T) {
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	tests := []struct {
		name      string
		news      *News
		userID    string
		setupFunc func(*MockNewsRepository)
		wantErr   bool
	}{
		{
			name: "successfully create new news",
			news: &News{
				Title:            "Breaking News Story",
				Content:          "Content for breaking news story with sufficient length to meet validation requirements",
				Summary:          "Breaking news summary",
				Slug:             "breaking-news-story",
				CategoryID:       "550e8400-e29b-41d4-a716-446655440002",
				AuthorName:       "John Doe",
				PublishingStatus: PublishingStatusDraft,
				NewsType:         NewsTypeAnnouncement,
				PriorityLevel:    PriorityLevelNormal,
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockNewsRepository) {},
			wantErr:   false,
		},
		{
			name: "return validation error for invalid news",
			news: &News{
				Title: "", // Invalid: empty title
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockNewsRepository) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockNewsRepository()
			tt.setupFunc(repo)
			service := NewNewsService(repo)

			err := service.CreateNews(ctx, tt.news, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, repo.auditEvents, 1)
				assert.Equal(t, domain.AuditEventInsert, repo.auditEvents[0].OperationType)
			}
		})
	}
}

func TestNewsService_UpdateNews(t *testing.T) {
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	tests := []struct {
		name      string
		news      *News
		userID    string
		setupFunc func(*MockNewsRepository)
		wantErr   bool
	}{
		{
			name: "successfully update existing news",
			news: &News{
				NewsID:           "550e8400-e29b-41d4-a716-446655440001",
				Title:            "Updated News Title",
				Content:          "Updated content for news article with sufficient length to meet validation requirements",
				Summary:          "Updated summary",
				Slug:             "updated-news-title",
				CategoryID:       "550e8400-e29b-41d4-a716-446655440002",
				AuthorName:       "John Doe",
				PublishingStatus: PublishingStatusDraft,
				NewsType:         NewsTypeAnnouncement,
				PriorityLevel:    PriorityLevelNormal,
			},
			userID: "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockNewsRepository) {
				repo.news["550e8400-e29b-41d4-a716-446655440001"] = &News{
					NewsID:           "550e8400-e29b-41d4-a716-446655440001",
					Title:            "Original Title",
					Content:          "Original content",
					Summary:          "Original summary",
					PublishingStatus: PublishingStatusDraft,
				}
			},
			wantErr: false,
		},
		{
			name: "return not found error for non-existent news",
			news: &News{
				NewsID: "550e8400-e29b-41d4-a716-446655440999",
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockNewsRepository) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockNewsRepository()
			tt.setupFunc(repo)
			service := NewNewsService(repo)

			err := service.UpdateNews(ctx, tt.news, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, repo.auditEvents, 1)
				assert.Equal(t, domain.AuditEventUpdate, repo.auditEvents[0].OperationType)
			}
		})
	}
}

func TestNewsService_DeleteNews(t *testing.T) {
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	tests := []struct {
		name      string
		newsID    string
		userID    string
		setupFunc func(*MockNewsRepository)
		wantErr   bool
	}{
		{
			name:   "successfully delete existing news",
			newsID: "550e8400-e29b-41d4-a716-446655440001",
			userID: "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockNewsRepository) {
				repo.news["550e8400-e29b-41d4-a716-446655440001"] = &News{
					NewsID:           "550e8400-e29b-41d4-a716-446655440001",
					Title:            "News to Delete",
					PublishingStatus: PublishingStatusDraft,
				}
			},
			wantErr: false,
		},
		{
			name:      "return not found error for non-existent news",
			newsID:    "550e8400-e29b-41d4-a716-446655440999",
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockNewsRepository) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockNewsRepository()
			tt.setupFunc(repo)
			service := NewNewsService(repo)

			err := service.DeleteNews(ctx, tt.newsID, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, repo.auditEvents, 1)
				assert.Equal(t, domain.AuditEventDelete, repo.auditEvents[0].OperationType)
			}
		})
	}
}

func TestNewsService_PublishNews(t *testing.T) {
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	tests := []struct {
		name      string
		newsID    string
		userID    string
		setupFunc func(*MockNewsRepository)
		wantErr   bool
	}{
		{
			name:   "successfully publish draft news",
			newsID: "550e8400-e29b-41d4-a716-446655440001",
			userID: "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockNewsRepository) {
				repo.news["550e8400-e29b-41d4-a716-446655440001"] = &News{
					NewsID:           "550e8400-e29b-41d4-a716-446655440001",
					Title:            "Complete News Article",
					Content:          "Complete content for publication",
					Summary:          "Complete summary",
					AuthorName:       "Author Name",
					PublishingStatus: PublishingStatusDraft,
				}
			},
			wantErr: false,
		},
		{
			name:   "return validation error for news without required fields",
			newsID: "550e8400-e29b-41d4-a716-446655440001",
			userID: "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockNewsRepository) {
				repo.news["550e8400-e29b-41d4-a716-446655440001"] = &News{
					NewsID:           "550e8400-e29b-41d4-a716-446655440001",
					Title:            "Incomplete News",
					Content:          "", // Missing content
					PublishingStatus: PublishingStatusDraft,
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockNewsRepository()
			tt.setupFunc(repo)
			service := NewNewsService(repo)

			err := service.PublishNews(ctx, tt.newsID, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, repo.auditEvents, 1)
				assert.Equal(t, domain.AuditEventPublish, repo.auditEvents[0].OperationType)
			}
		})
	}
}

func TestNewsService_ArchiveNews(t *testing.T) {
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	tests := []struct {
		name      string
		newsID    string
		userID    string
		setupFunc func(*MockNewsRepository)
		wantErr   bool
	}{
		{
			name:   "successfully archive published news",
			newsID: "550e8400-e29b-41d4-a716-446655440001",
			userID: "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockNewsRepository) {
				repo.news["550e8400-e29b-41d4-a716-446655440001"] = &News{
					NewsID:           "550e8400-e29b-41d4-a716-446655440001",
					Title:            "Published News",
					PublishingStatus: PublishingStatusPublished,
				}
			},
			wantErr: false,
		},
		{
			name:      "return not found error for non-existent news",
			newsID:    "550e8400-e29b-41d4-a716-446655440999",
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockNewsRepository) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockNewsRepository()
			tt.setupFunc(repo)
			service := NewNewsService(repo)

			err := service.ArchiveNews(ctx, tt.newsID, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, repo.auditEvents, 1)
				assert.Equal(t, domain.AuditEventArchive, repo.auditEvents[0].OperationType)
			}
		})
	}
}

func TestNewsService_CreateNewsCategory(t *testing.T) {
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	tests := []struct {
		name      string
		category  *NewsCategory
		userID    string
		setupFunc func(*MockNewsRepository)
		wantErr   bool
	}{
		{
			name: "successfully create new news category",
			category: &NewsCategory{
				Name:        "Breaking News",
				Slug:        "breaking-news",
				Description: "Category for breaking news articles",
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockNewsRepository) {},
			wantErr:   false,
		},
		{
			name: "return validation error for invalid category",
			category: &NewsCategory{
				Name: "", // Invalid: empty name
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockNewsRepository) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockNewsRepository()
			tt.setupFunc(repo)
			service := NewNewsService(repo)

			err := service.CreateNewsCategory(ctx, tt.category, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, repo.auditEvents, 1)
				assert.Equal(t, domain.AuditEventInsert, repo.auditEvents[0].OperationType)
			}
		})
	}
}

func TestNewsService_UpdateNewsCategory(t *testing.T) {
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	tests := []struct {
		name      string
		category  *NewsCategory
		userID    string
		setupFunc func(*MockNewsRepository)
		wantErr   bool
	}{
		{
			name: "successfully update existing news category",
			category: &NewsCategory{
				CategoryID:  "550e8400-e29b-41d4-a716-446655440002",
				Name:        "Updated Category Name",
				Slug:        "updated-category-name",
				Description: "Updated description",
			},
			userID: "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockNewsRepository) {
				repo.categories["550e8400-e29b-41d4-a716-446655440002"] = &NewsCategory{
					CategoryID:  "550e8400-e29b-41d4-a716-446655440002",
					Name:        "Original Category",
					Description: "Original description",
				}
			},
			wantErr: false,
		},
		{
			name: "return not found error for non-existent category",
			category: &NewsCategory{
				CategoryID: "550e8400-e29b-41d4-a716-446655440999",
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockNewsRepository) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockNewsRepository()
			tt.setupFunc(repo)
			service := NewNewsService(repo)

			err := service.UpdateNewsCategory(ctx, tt.category, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, repo.auditEvents, 1)
				assert.Equal(t, domain.AuditEventUpdate, repo.auditEvents[0].OperationType)
			}
		})
	}
}

func TestNewsService_DeleteNewsCategory(t *testing.T) {
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	tests := []struct {
		name       string
		categoryID string
		userID     string
		setupFunc  func(*MockNewsRepository)
		wantErr    bool
	}{
		{
			name:       "successfully delete existing news category",
			categoryID: "550e8400-e29b-41d4-a716-446655440002",
			userID:     "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockNewsRepository) {
				repo.categories["550e8400-e29b-41d4-a716-446655440002"] = &NewsCategory{
					CategoryID:            "550e8400-e29b-41d4-a716-446655440002",
					Name:                  "Category to Delete",
					IsDefaultUnassigned:   false,
				}
			},
			wantErr: false,
		},
		{
			name:       "return validation error for default unassigned category",
			categoryID: "550e8400-e29b-41d4-a716-446655440002",
			userID:     "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockNewsRepository) {
				repo.categories["550e8400-e29b-41d4-a716-446655440002"] = &NewsCategory{
					CategoryID:            "550e8400-e29b-41d4-a716-446655440002",
					Name:                  "Unassigned",
					IsDefaultUnassigned:   true,
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockNewsRepository()
			tt.setupFunc(repo)
			service := NewNewsService(repo)

			err := service.DeleteNewsCategory(ctx, tt.categoryID, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, repo.auditEvents, 1)
				assert.Equal(t, domain.AuditEventDelete, repo.auditEvents[0].OperationType)
			}
		})
	}
}

func TestNewsService_SetFeaturedNews(t *testing.T) {
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	tests := []struct {
		name      string
		newsID    string
		userID    string
		setupFunc func(*MockNewsRepository)
		wantErr   bool
	}{
		{
			name:   "successfully set featured news",
			newsID: "550e8400-e29b-41d4-a716-446655440001",
			userID: "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockNewsRepository) {
				repo.news["550e8400-e29b-41d4-a716-446655440001"] = &News{
					NewsID:           "550e8400-e29b-41d4-a716-446655440001",
					Title:            "Featured News Article",
					PublishingStatus: PublishingStatusPublished,
				}
			},
			wantErr: false,
		},
		{
			name:   "return validation error for unpublished news",
			newsID: "550e8400-e29b-41d4-a716-446655440001",
			userID: "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockNewsRepository) {
				repo.news["550e8400-e29b-41d4-a716-446655440001"] = &News{
					NewsID:           "550e8400-e29b-41d4-a716-446655440001",
					Title:            "Draft News Article",
					PublishingStatus: PublishingStatusDraft,
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockNewsRepository()
			tt.setupFunc(repo)
			service := NewNewsService(repo)

			err := service.SetFeaturedNews(ctx, tt.newsID, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, repo.auditEvents, 1)
				assert.Equal(t, domain.AuditEventInsert, repo.auditEvents[0].OperationType)
			}
		})
	}
}

func TestNewsService_RemoveFeaturedNews(t *testing.T) {
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()

	tests := []struct {
		name      string
		userID    string
		setupFunc func(*MockNewsRepository)
		wantErr   bool
	}{
		{
			name:   "successfully remove featured news",
			userID: "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockNewsRepository) {
				repo.featuredNews["550e8400-e29b-41d4-a716-446655440005"] = &FeaturedNews{
					FeaturedNewsID: "550e8400-e29b-41d4-a716-446655440005",
					NewsID:         "550e8400-e29b-41d4-a716-446655440001",
				}
			},
			wantErr: false,
		},
		{
			name:      "return not found error when no featured news exists",
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockNewsRepository) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockNewsRepository()
			tt.setupFunc(repo)
			service := NewNewsService(repo)

			err := service.RemoveFeaturedNews(ctx, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, repo.auditEvents, 1)
				assert.Equal(t, domain.AuditEventDelete, repo.auditEvents[0].OperationType)
			}
		})
	}
}