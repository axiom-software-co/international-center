package research

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

// MockResearchRepository provides mock implementation for unit tests
type MockResearchRepository struct {
	research           map[string]*Research
	categories         map[string]*ResearchCategory  
	featuredResearch   map[string]*FeaturedResearch
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

func NewMockResearchRepository() *MockResearchRepository {
	return &MockResearchRepository{
		research:           make(map[string]*Research),
		categories:         make(map[string]*ResearchCategory),
		featuredResearch:   make(map[string]*FeaturedResearch),
		auditEvents:        make([]MockAuditEvent, 0),
		failures:           make(map[string]error),
	}
}

// SetFailure sets a mock failure for specific operations
func (m *MockResearchRepository) SetFailure(operation string, err error) {
	m.failures[operation] = err
}

// GetAuditEvents returns all mock audit events
func (m *MockResearchRepository) GetAuditEvents() []MockAuditEvent {
	return m.auditEvents
}

// Research CRUD operations - these will be implemented in Green phase
func (m *MockResearchRepository) GetResearch(ctx context.Context, researchID string) (*Research, error) {
	if err, exists := m.failures["GetResearch"]; exists {
		return nil, err
	}
	if research, exists := m.research[researchID]; exists && !research.IsDeleted {
		return research, nil
	}
	return nil, domain.NewNotFoundError("research", researchID)
}

func (m *MockResearchRepository) GetAllResearch(ctx context.Context, limit, offset int) ([]*Research, error) {
	if err, exists := m.failures["GetAllResearch"]; exists {
		return nil, err
	}
	
	researchList := make([]*Research, 0)
	for _, research := range m.research {
		if !research.IsDeleted {
			researchList = append(researchList, research)
		}
	}
	return researchList, nil
}

func (m *MockResearchRepository) GetResearchBySlug(ctx context.Context, slug string) (*Research, error) {
	if err, exists := m.failures["GetResearchBySlug"]; exists {
		return nil, err
	}
	for _, research := range m.research {
		if research.Slug == slug && !research.IsDeleted {
			return research, nil
		}
	}
	return nil, domain.NewNotFoundError("research", slug)
}

func (m *MockResearchRepository) SearchResearch(ctx context.Context, query string, limit, offset int) ([]*Research, error) {
	if err, exists := m.failures["SearchResearch"]; exists {
		return nil, err
	}
	
	researchList := make([]*Research, 0)
	queryLower := strings.ToLower(query)
	for _, research := range m.research {
		if !research.IsDeleted {
			if strings.Contains(strings.ToLower(research.Title), queryLower) ||
			   strings.Contains(strings.ToLower(research.Abstract), queryLower) {
				researchList = append(researchList, research)
			}
		}
	}
	return researchList, nil
}

func (m *MockResearchRepository) GetResearchByPublishingStatus(ctx context.Context, status PublishingStatus, limit, offset int) ([]*Research, error) {
	if err, exists := m.failures["GetResearchByPublishingStatus"]; exists {
		return nil, err
	}
	
	researchList := make([]*Research, 0)
	count := 0
	for _, research := range m.research {
		if research.PublishingStatus == status && !research.IsDeleted {
			// Apply offset
			if count < offset {
				count++
				continue
			}
			// Apply limit
			if len(researchList) >= limit {
				break
			}
			researchList = append(researchList, research)
		}
		count++
	}
	return researchList, nil
}

func (m *MockResearchRepository) SaveResearch(ctx context.Context, research *Research) error {
	if err, exists := m.failures["SaveResearch"]; exists {
		return err
	}
	m.research[research.ResearchID] = research
	return nil
}

func (m *MockResearchRepository) DeleteResearch(ctx context.Context, researchID string) error {
	if err, exists := m.failures["DeleteResearch"]; exists {
		return err
	}
	if research, exists := m.research[researchID]; exists {
		research.IsDeleted = true
		now := time.Now()
		research.DeletedOn = &now
		return nil
	}
	return domain.NewNotFoundError("research", researchID)
}

// Research category operations
func (m *MockResearchRepository) GetAllResearchCategories(ctx context.Context) ([]*ResearchCategory, error) {
	if err, exists := m.failures["GetAllResearchCategories"]; exists {
		return nil, err
	}
	
	categories := make([]*ResearchCategory, 0)
	for _, category := range m.categories {
		if !category.IsDeleted {
			categories = append(categories, category)
		}
	}
	return categories, nil
}

func (m *MockResearchRepository) GetResearchCategory(ctx context.Context, categoryID string) (*ResearchCategory, error) {
	if err, exists := m.failures["GetResearchCategory"]; exists {
		return nil, err
	}
	if category, exists := m.categories[categoryID]; exists && !category.IsDeleted {
		return category, nil
	}
	return nil, domain.NewNotFoundError("research_category", categoryID)
}

func (m *MockResearchRepository) GetResearchCategoryBySlug(ctx context.Context, slug string) (*ResearchCategory, error) {
	if err, exists := m.failures["GetResearchCategoryBySlug"]; exists {
		return nil, err
	}
	for _, category := range m.categories {
		if category.Slug == slug && !category.IsDeleted {
			return category, nil
		}
	}
	return nil, domain.NewNotFoundError("research_category", slug)
}

func (m *MockResearchRepository) SaveResearchCategory(ctx context.Context, category *ResearchCategory) error {
	if err, exists := m.failures["SaveResearchCategory"]; exists {
		return err
	}
	m.categories[category.CategoryID] = category
	return nil
}

func (m *MockResearchRepository) GetDefaultUnassignedCategory(ctx context.Context) (*ResearchCategory, error) {
	if err, exists := m.failures["GetDefaultUnassignedCategory"]; exists {
		return nil, err
	}
	for _, category := range m.categories {
		if category.IsDefaultUnassigned && !category.IsDeleted {
			return category, nil
		}
	}
	return nil, domain.NewNotFoundError("default_unassigned_category", "")
}

func (m *MockResearchRepository) DeleteResearchCategory(ctx context.Context, categoryID string) error {
	if err, exists := m.failures["DeleteResearchCategory"]; exists {
		return err
	}
	if category, exists := m.categories[categoryID]; exists {
		category.IsDeleted = true
		now := time.Now()
		category.DeletedOn = &now
		return nil
	}
	return domain.NewNotFoundError("research_category", categoryID)
}

func (m *MockResearchRepository) GetResearchByCategory(ctx context.Context, categoryID string, limit, offset int) ([]*Research, error) {
	if err, exists := m.failures["GetResearchByCategory"]; exists {
		return nil, err
	}
	
	researchList := make([]*Research, 0)
	for _, research := range m.research {
		if research.CategoryID == categoryID && !research.IsDeleted {
			researchList = append(researchList, research)
		}
	}
	return researchList, nil
}

// Featured research operations
func (m *MockResearchRepository) GetFeaturedResearch(ctx context.Context) (*FeaturedResearch, error) {
	if err, exists := m.failures["GetFeaturedResearch"]; exists {
		return nil, err
	}
	
	for _, featured := range m.featuredResearch {
		return featured, nil // Only one featured research allowed
	}
	return nil, domain.NewNotFoundError("featured_research", "featured")
}

func (m *MockResearchRepository) SaveFeaturedResearch(ctx context.Context, featured *FeaturedResearch) error {
	if err, exists := m.failures["SaveFeaturedResearch"]; exists {
		return err
	}
	
	// Clear existing featured research (only one allowed)
	m.featuredResearch = make(map[string]*FeaturedResearch)
	m.featuredResearch[featured.FeaturedResearchID] = featured
	return nil
}

func (m *MockResearchRepository) DeleteFeaturedResearch(ctx context.Context, featuredResearchID string) error {
	if err, exists := m.failures["DeleteFeaturedResearch"]; exists {
		return err
	}
	delete(m.featuredResearch, featuredResearchID)
	return nil
}

// Audit repository methods - these will be implemented in Green phase
func (m *MockResearchRepository) GetResearchAudit(ctx context.Context, researchID string, userID string, limit int, offset int) ([]*domain.AuditEvent, error) {
	if err, exists := m.failures["GetResearchAudit"]; exists {
		return nil, err
	}
	// Return mock audit events for this research
	events := make([]*domain.AuditEvent, 0)
	for _, auditEvent := range m.auditEvents {
		if auditEvent.EntityID == researchID {
			events = append(events, &domain.AuditEvent{
				AuditID:       "audit-" + researchID + "-1",
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

func (m *MockResearchRepository) GetResearchCategoryAudit(ctx context.Context, categoryID string, userID string, limit int, offset int) ([]*domain.AuditEvent, error) {
	if err, exists := m.failures["GetResearchCategoryAudit"]; exists {
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
func (m *MockResearchRepository) PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error {
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

// Test Research Service Operations
func TestResearchService_GetResearch(t *testing.T) {
	tests := []struct {
		name       string
		researchID string
		setupFunc  func(*MockResearchRepository)
		wantErr    bool
		wantTitle  string
	}{
		{
			name:       "successfully retrieve existing research",
			researchID: "550e8400-e29b-41d4-a716-446655440001",
			setupFunc: func(repo *MockResearchRepository) {
				research := &Research{
					ResearchID: "550e8400-e29b-41d4-a716-446655440001",
					Title:      "Clinical Study Results",
					Abstract:   "This study examines clinical outcomes",
					Slug:       "clinical-study-results",
					IsDeleted:  false,
				}
				repo.research["550e8400-e29b-41d4-a716-446655440001"] = research
			},
			wantErr:   false,
			wantTitle: "Clinical Study Results",
		},
		{
			name:       "return not found error for non-existent research",
			researchID: "nonexistent",
			setupFunc:  func(repo *MockResearchRepository) {},
			wantErr:    true,
		},
		{
			name:       "return not found error for soft deleted research",
			researchID: "deleted-research",
			setupFunc: func(repo *MockResearchRepository) {
				research := &Research{
					ResearchID: "deleted-research",
					Title:      "Deleted Research",
					IsDeleted:  true,
				}
				repo.research["deleted-research"] = research
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			repo := NewMockResearchRepository()
			tt.setupFunc(repo)
			
			service := NewResearchService(repo)
			
			result, err := service.GetResearch(ctx, tt.researchID, "550e8400-e29b-41d4-a716-446655440004")
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.wantTitle, result.Title)
			}
		})
	}
}

func TestResearchService_GetAllResearch(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*MockResearchRepository)
		wantCount int
		wantErr   bool
	}{
		{
			name: "successfully retrieve all research",
			setupFunc: func(repo *MockResearchRepository) {
				repo.research["550e8400-e29b-41d4-a716-446655440001"] = &Research{ResearchID: "550e8400-e29b-41d4-a716-446655440001", Title: "Research 1", IsDeleted: false}
				repo.research["research-2"] = &Research{ResearchID: "research-2", Title: "Research 2", IsDeleted: false}
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "return empty array when no research exists",
			setupFunc: func(repo *MockResearchRepository) {},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "exclude soft deleted research from results",
			setupFunc: func(repo *MockResearchRepository) {
				repo.research["550e8400-e29b-41d4-a716-446655440001"] = &Research{ResearchID: "550e8400-e29b-41d4-a716-446655440001", Title: "Active Research", IsDeleted: false}
				repo.research["research-2"] = &Research{ResearchID: "research-2", Title: "Deleted Research", IsDeleted: true}
			},
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			repo := NewMockResearchRepository()
			tt.setupFunc(repo)
			
			service := NewResearchService(repo)
			
			result, err := service.GetAllResearch(ctx, 10, 0)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, tt.wantCount)
			}
		})
	}
}

func TestResearchService_SearchResearch(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		setupFunc func(*MockResearchRepository)
		wantCount int
		wantErr   bool
	}{
		{
			name:  "successfully search research by title",
			query: "clinical",
			setupFunc: func(repo *MockResearchRepository) {
				repo.research["550e8400-e29b-41d4-a716-446655440001"] = &Research{
					ResearchID: "550e8400-e29b-41d4-a716-446655440001",
					Title:      "Clinical Study Results",
					Abstract:   "Study abstract",
					IsDeleted:  false,
				}
				repo.research["research-2"] = &Research{
					ResearchID: "research-2", 
					Title:      "Meta Analysis Report",
					Abstract:   "Analysis abstract",
					IsDeleted:  false,
				}
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:  "successfully search research by abstract",
			query: "outcomes",
			setupFunc: func(repo *MockResearchRepository) {
				repo.research["550e8400-e29b-41d4-a716-446655440001"] = &Research{
					ResearchID: "550e8400-e29b-41d4-a716-446655440001",
					Title:      "Study Report", 
					Abstract:   "This study examines patient outcomes",
					IsDeleted:  false,
				}
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:  "return empty results when no matches found",
			query: "nonexistent",
			setupFunc: func(repo *MockResearchRepository) {
				repo.research["550e8400-e29b-41d4-a716-446655440001"] = &Research{
					ResearchID: "550e8400-e29b-41d4-a716-446655440001",
					Title:      "Clinical Study",
					Abstract:   "Study abstract",
					IsDeleted:  false,
				}
			},
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			repo := NewMockResearchRepository()
			tt.setupFunc(repo)
			
			service := NewResearchService(repo)
			
			result, err := service.SearchResearch(ctx, tt.query, 10, 0)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, tt.wantCount)
			}
		})
	}
}

func TestResearchService_GetFeaturedResearch(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*MockResearchRepository)
		wantErr   bool
	}{
		{
			name: "successfully retrieve featured research",
			setupFunc: func(repo *MockResearchRepository) {
				featured := &FeaturedResearch{
					FeaturedResearchID: "featured-1",
					ResearchID:         "550e8400-e29b-41d4-a716-446655440001", 
				}
				repo.featuredResearch["featured-1"] = featured
			},
			wantErr: false,
		},
		{
			name:      "return not found when no featured research exists",
			setupFunc: func(repo *MockResearchRepository) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			repo := NewMockResearchRepository()
			tt.setupFunc(repo)
			
			service := NewResearchService(repo)
			
			result, err := service.GetFeaturedResearch(ctx)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestResearchService_GetResearchAudit(t *testing.T) {
	tests := []struct {
		name       string
		researchID string
		userID     string
		setupFunc  func(*MockResearchRepository)
		wantErr    bool
		wantCount  int
	}{
		{
			name:       "successfully retrieve research audit events",
			researchID: "550e8400-e29b-41d4-a716-446655440001",
			userID:     "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockResearchRepository) {
				repo.auditEvents = append(repo.auditEvents, MockAuditEvent{
					EntityType:    domain.EntityTypeResearch,
					EntityID:      "550e8400-e29b-41d4-a716-446655440001",
					OperationType: domain.AuditEventInsert,
					UserID:        "admin-550e8400-e29b-41d4-a716-446655440003",
				})
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name:       "return empty array for research with no audit events",
			researchID: "research-2",
			userID:     "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc:  func(repo *MockResearchRepository) {},
			wantErr:    false,
			wantCount:  0,
		},
		{
			name:       "return unauthorized error for non-admin user",
			researchID: "550e8400-e29b-41d4-a716-446655440001",
			userID:     "550e8400-e29b-41d4-a716-446655440004",
			setupFunc:  func(repo *MockResearchRepository) {},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			repo := NewMockResearchRepository()
			tt.setupFunc(repo)
			
			service := NewResearchService(repo)
			
			result, err := service.GetResearchAudit(ctx, tt.researchID, tt.userID, 10, 0)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, tt.wantCount)
			}
		})
	}
}

func TestResearchService_GetResearchCategoryAudit(t *testing.T) {
	tests := []struct {
		name       string
		categoryID string
		userID     string
		setupFunc  func(*MockResearchRepository)
		wantErr    bool
	}{
		{
			name:       "successfully retrieve research category audit events",
			categoryID: "550e8400-e29b-41d4-a716-446655440002",
			userID:     "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockResearchRepository) {
				repo.auditEvents = append(repo.auditEvents, MockAuditEvent{
					EntityType:    domain.EntityTypeCategory,
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
			userID:     "550e8400-e29b-41d4-a716-446655440004",
			setupFunc:  func(repo *MockResearchRepository) {},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			repo := NewMockResearchRepository()
			tt.setupFunc(repo)
			
			service := NewResearchService(repo)
			
			result, err := service.GetResearchCategoryAudit(ctx, tt.categoryID, tt.userID, 10, 0)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// Admin CRUD Operation Tests

func TestResearchService_CreateResearch(t *testing.T) {
	tests := []struct {
		name      string
		research  *Research
		userID    string
		setupFunc func(*MockResearchRepository)
		wantErr   bool
	}{
		{
			name: "successfully create new research",
			research: &Research{
				Title:            "New Clinical Study",
				Abstract:         "Abstract for new clinical study with sufficient length to meet validation requirements",
				Slug:             "new-clinical-study",
				CategoryID:       "550e8400-e29b-41d4-a716-446655440002",
				AuthorNames:      "Dr. Smith, Dr. Johnson",
				PublishingStatus: PublishingStatusDraft,
				ResearchType:     ResearchTypeClinicalStudy,
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockResearchRepository) {},
			wantErr:   false,
		},
		{
			name: "return validation error for invalid research",
			research: &Research{
				Title: "", // Invalid: empty title
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockResearchRepository) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			repo := NewMockResearchRepository()
			tt.setupFunc(repo)
			
			service := NewResearchService(repo)
			
			err := service.CreateResearch(ctx, tt.research, tt.userID)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify audit event was published
				events := repo.GetAuditEvents()
				assert.Len(t, events, 1)
				assert.Equal(t, domain.EntityTypeResearch, events[0].EntityType)
				assert.Equal(t, domain.AuditEventInsert, events[0].OperationType)
			}
		})
	}
}

func TestResearchService_UpdateResearch(t *testing.T) {
	tests := []struct {
		name      string
		research  *Research
		userID    string
		setupFunc func(*MockResearchRepository)
		wantErr   bool
	}{
		{
			name: "successfully update existing research",
			research: &Research{
				ResearchID:       "550e8400-e29b-41d4-a716-446655440001",
				Title:            "Updated Clinical Study",
				Abstract:         "Updated abstract for clinical study with sufficient length to meet validation requirements",
				Slug:             "updated-clinical-study", 
				CategoryID:       "550e8400-e29b-41d4-a716-446655440002",
				AuthorNames:      "Dr. Smith, Dr. Johnson",
				PublishingStatus: PublishingStatusDraft,
				ResearchType:     ResearchTypeClinicalStudy,
			},
			userID: "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockResearchRepository) {
				existing := &Research{
					ResearchID:       "550e8400-e29b-41d4-a716-446655440001",
					Title:            "Original Clinical Study",
					Abstract:         "Original abstract for clinical study with sufficient length to meet validation requirements",
					Slug:             "original-clinical-study",
					CategoryID:       "550e8400-e29b-41d4-a716-446655440002", 
					AuthorNames:      "Dr. Smith",
					PublishingStatus: PublishingStatusDraft,
					ResearchType:     ResearchTypeClinicalStudy,
					CreatedOn:        time.Now(),
				}
				repo.research["550e8400-e29b-41d4-a716-446655440001"] = existing
			},
			wantErr: false,
		},
		{
			name: "return not found error for non-existent research",
			research: &Research{
				ResearchID: "nonexistent",
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockResearchRepository) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			repo := NewMockResearchRepository()
			tt.setupFunc(repo)
			
			service := NewResearchService(repo)
			
			err := service.UpdateResearch(ctx, tt.research, tt.userID)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify audit event was published
				events := repo.GetAuditEvents()
				assert.Len(t, events, 1)
				assert.Equal(t, domain.EntityTypeResearch, events[0].EntityType)
				assert.Equal(t, domain.AuditEventUpdate, events[0].OperationType)
			}
		})
	}
}

func TestResearchService_DeleteResearch(t *testing.T) {
	tests := []struct {
		name       string
		researchID string
		userID     string
		setupFunc  func(*MockResearchRepository)
		wantErr    bool
	}{
		{
			name:       "successfully delete existing research",
			researchID: "550e8400-e29b-41d4-a716-446655440001",
			userID:     "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockResearchRepository) {
				research := &Research{
					ResearchID:       "550e8400-e29b-41d4-a716-446655440001",
					Title:            "Test Research",
					Abstract:         "Abstract for test research with sufficient length to meet validation requirements",
					Slug:             "test-research",
					CategoryID:       "550e8400-e29b-41d4-a716-446655440002",
					AuthorNames:      "Dr. Test",
					PublishingStatus: PublishingStatusDraft,
					ResearchType:     ResearchTypeClinicalStudy,
					CreatedOn:        time.Now(),
				}
				repo.research["550e8400-e29b-41d4-a716-446655440001"] = research
			},
			wantErr: false,
		},
		{
			name:       "return not found error for non-existent research",
			researchID: "nonexistent",
			userID:     "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc:  func(repo *MockResearchRepository) {},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			repo := NewMockResearchRepository()
			tt.setupFunc(repo)
			
			service := NewResearchService(repo)
			
			err := service.DeleteResearch(ctx, tt.researchID, tt.userID)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify research is soft deleted
				if research, exists := repo.research[tt.researchID]; exists {
					assert.True(t, research.IsDeleted)
				}
				// Verify audit event was published
				events := repo.GetAuditEvents()
				assert.Len(t, events, 1)
				assert.Equal(t, domain.EntityTypeResearch, events[0].EntityType)
				assert.Equal(t, domain.AuditEventDelete, events[0].OperationType)
			}
		})
	}
}

func TestResearchService_PublishResearch(t *testing.T) {
	tests := []struct {
		name       string
		researchID string
		userID     string
		setupFunc  func(*MockResearchRepository)
		wantErr    bool
	}{
		{
			name:       "successfully publish draft research",
			researchID: "550e8400-e29b-41d4-a716-446655440001",
			userID:     "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockResearchRepository) {
				research := &Research{
					ResearchID:       "550e8400-e29b-41d4-a716-446655440001",
					Title:            "Test Research", 
					Abstract:         "Abstract for test research with sufficient length to meet validation requirements",
					Slug:             "test-research",
					CategoryID:       "550e8400-e29b-41d4-a716-446655440002",
					AuthorNames:      "Dr. Test",
					PublishingStatus: PublishingStatusDraft,
					ResearchType:     ResearchTypeClinicalStudy,
					CreatedOn:        time.Now(),
				}
				repo.research["550e8400-e29b-41d4-a716-446655440001"] = research
			},
			wantErr: false,
		},
		{
			name:       "return validation error for research without required fields",
			researchID: "research-2",
			userID:     "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockResearchRepository) {
				research := &Research{
					ResearchID:       "research-2",
					Title:            "", // Missing required field
					PublishingStatus: PublishingStatusDraft,
					ResearchType:     ResearchTypeClinicalStudy,
					CreatedOn:        time.Now(),
				}
				repo.research["research-2"] = research
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			repo := NewMockResearchRepository()
			tt.setupFunc(repo)
			
			service := NewResearchService(repo)
			
			err := service.PublishResearch(ctx, tt.researchID, tt.userID)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify status changed to published
				research := repo.research[tt.researchID]
				assert.Equal(t, PublishingStatusPublished, research.PublishingStatus)
				// Verify audit event was published
				events := repo.GetAuditEvents()
				assert.Len(t, events, 1)
				assert.Equal(t, domain.EntityTypeResearch, events[0].EntityType)
				assert.Equal(t, domain.AuditEventPublish, events[0].OperationType)
			}
		})
	}
}

func TestResearchService_ArchiveResearch(t *testing.T) {
	tests := []struct {
		name       string
		researchID string
		userID     string
		setupFunc  func(*MockResearchRepository)
		wantErr    bool
	}{
		{
			name:       "successfully archive published research",
			researchID: "550e8400-e29b-41d4-a716-446655440001",
			userID:     "admin-550e8400-e29b-41d4-a716-446655440003", 
			setupFunc: func(repo *MockResearchRepository) {
				research := &Research{
					ResearchID:       "550e8400-e29b-41d4-a716-446655440001",
					Title:            "Test Research",
					Abstract:         "Abstract for test research with sufficient length to meet validation requirements",
					Slug:             "test-research",
					CategoryID:       "550e8400-e29b-41d4-a716-446655440002",
					AuthorNames:      "Dr. Test",
					PublishingStatus: PublishingStatusPublished,
					ResearchType:     ResearchTypeClinicalStudy,
					CreatedOn:        time.Now(),
				}
				repo.research["550e8400-e29b-41d4-a716-446655440001"] = research
			},
			wantErr: false,
		},
		{
			name:       "return not found error for non-existent research",
			researchID: "nonexistent",
			userID:     "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc:  func(repo *MockResearchRepository) {},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			repo := NewMockResearchRepository()
			tt.setupFunc(repo)
			
			service := NewResearchService(repo)
			
			err := service.ArchiveResearch(ctx, tt.researchID, tt.userID)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify status changed to archived
				research := repo.research[tt.researchID]
				assert.Equal(t, PublishingStatusArchived, research.PublishingStatus)
				// Verify audit event was published
				events := repo.GetAuditEvents()
				assert.Len(t, events, 1)
				assert.Equal(t, domain.EntityTypeResearch, events[0].EntityType)
				assert.Equal(t, domain.AuditEventArchive, events[0].OperationType)
			}
		})
	}
}

func TestResearchService_CreateResearchCategory(t *testing.T) {
	tests := []struct {
		name      string
		category  *ResearchCategory
		userID    string
		setupFunc func(*MockResearchRepository)
		wantErr   bool
	}{
		{
			name: "successfully create new research category",
			category: &ResearchCategory{
				Name:        "New Category",
				Slug:        "new-category",
				Description: "Description for new category",
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockResearchRepository) {},
			wantErr:   false,
		},
		{
			name: "return validation error for invalid category",
			category: &ResearchCategory{
				Name: "", // Invalid: empty name
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockResearchRepository) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			repo := NewMockResearchRepository()
			tt.setupFunc(repo)
			
			service := NewResearchService(repo)
			
			err := service.CreateResearchCategory(ctx, tt.category, tt.userID)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify audit event was published
				events := repo.GetAuditEvents()
				assert.Len(t, events, 1)
				assert.Equal(t, domain.EntityTypeResearchCategory, events[0].EntityType)
				assert.Equal(t, domain.AuditEventInsert, events[0].OperationType)
			}
		})
	}
}

func TestResearchService_UpdateResearchCategory(t *testing.T) {
	tests := []struct {
		name      string
		category  *ResearchCategory
		userID    string
		setupFunc func(*MockResearchRepository)
		wantErr   bool
	}{
		{
			name: "successfully update existing research category",
			category: &ResearchCategory{
				CategoryID:  "550e8400-e29b-41d4-a716-446655440002",
				Name:        "Updated Category",
				Slug:        "updated-category",
				Description: "Updated description",
			},
			userID: "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockResearchRepository) {
				existing := &ResearchCategory{
					CategoryID:  "550e8400-e29b-41d4-a716-446655440002",
					Name:        "Original Category",
					Slug:        "original-category",
					Description: "Original description",
					CreatedOn:   time.Now(),
				}
				repo.categories["550e8400-e29b-41d4-a716-446655440002"] = existing
			},
			wantErr: false,
		},
		{
			name: "return not found error for non-existent category",
			category: &ResearchCategory{
				CategoryID: "nonexistent",
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockResearchRepository) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			repo := NewMockResearchRepository()
			tt.setupFunc(repo)
			
			service := NewResearchService(repo)
			
			err := service.UpdateResearchCategory(ctx, tt.category, tt.userID)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify audit event was published
				events := repo.GetAuditEvents()
				assert.Len(t, events, 1)
				assert.Equal(t, domain.EntityTypeResearchCategory, events[0].EntityType)
				assert.Equal(t, domain.AuditEventUpdate, events[0].OperationType)
			}
		})
	}
}

func TestResearchService_DeleteResearchCategory(t *testing.T) {
	tests := []struct {
		name       string
		categoryID string
		userID     string
		setupFunc  func(*MockResearchRepository)
		wantErr    bool
	}{
		{
			name:       "successfully delete existing research category",
			categoryID: "550e8400-e29b-41d4-a716-446655440002",
			userID:     "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockResearchRepository) {
				category := &ResearchCategory{
					CategoryID:            "550e8400-e29b-41d4-a716-446655440002",
					Name:                  "Test Category",
					Slug:                  "test-category",
					Description:           "Test description",
					IsDefaultUnassigned:   false,
					CreatedOn:             time.Now(),
				}
				repo.categories["550e8400-e29b-41d4-a716-446655440002"] = category
			},
			wantErr: false,
		},
		{
			name:       "return validation error for default unassigned category",
			categoryID: "default-category",
			userID:     "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockResearchRepository) {
				category := &ResearchCategory{
					CategoryID:            "default-category",
					Name:                  "Unassigned",
					Slug:                  "unassigned",
					IsDefaultUnassigned:   true,
					CreatedOn:             time.Now(),
				}
				repo.categories["default-category"] = category
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			repo := NewMockResearchRepository()
			tt.setupFunc(repo)
			
			service := NewResearchService(repo)
			
			err := service.DeleteResearchCategory(ctx, tt.categoryID, tt.userID)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify category is soft deleted
				if category, exists := repo.categories[tt.categoryID]; exists {
					assert.True(t, category.IsDeleted)
				}
				// Verify audit event was published
				events := repo.GetAuditEvents()
				assert.Len(t, events, 1)
				assert.Equal(t, domain.EntityTypeResearchCategory, events[0].EntityType)
				assert.Equal(t, domain.AuditEventDelete, events[0].OperationType)
			}
		})
	}
}

func TestResearchService_SetFeaturedResearch(t *testing.T) {
	tests := []struct {
		name       string
		researchID string
		userID     string
		setupFunc  func(*MockResearchRepository)
		wantErr    bool
	}{
		{
			name:       "successfully set featured research",
			researchID: "550e8400-e29b-41d4-a716-446655440001",
			userID:     "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockResearchRepository) {
				research := &Research{
					ResearchID:       "550e8400-e29b-41d4-a716-446655440001",
					Title:            "Test Research",
					Abstract:         "Abstract for test research with sufficient length to meet validation requirements",
					Slug:             "test-research",
					CategoryID:       "550e8400-e29b-41d4-a716-446655440002",
					AuthorNames:      "Dr. Test",
					PublishingStatus: PublishingStatusPublished,
					ResearchType:     ResearchTypeClinicalStudy,
					CreatedOn:        time.Now(),
				}
				repo.research["550e8400-e29b-41d4-a716-446655440001"] = research
			},
			wantErr: false,
		},
		{
			name:       "return validation error for unpublished research",
			researchID: "research-2", 
			userID:     "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockResearchRepository) {
				research := &Research{
					ResearchID:       "research-2",
					Title:            "Draft Research",
					Abstract:         "Abstract for draft research with sufficient length to meet validation requirements",
					Slug:             "draft-research",
					CategoryID:       "550e8400-e29b-41d4-a716-446655440002",
					AuthorNames:      "Dr. Test",
					PublishingStatus: PublishingStatusDraft,
					ResearchType:     ResearchTypeClinicalStudy,
					CreatedOn:        time.Now(),
				}
				repo.research["research-2"] = research
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			repo := NewMockResearchRepository()
			tt.setupFunc(repo)
			
			service := NewResearchService(repo)
			
			err := service.SetFeaturedResearch(ctx, tt.researchID, tt.userID)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify featured research was created
				assert.Len(t, repo.featuredResearch, 1)
				// Verify audit event was published
				events := repo.GetAuditEvents()
				assert.Len(t, events, 1)
				assert.Equal(t, domain.EntityTypeFeaturedResearch, events[0].EntityType)
				assert.Equal(t, domain.AuditEventInsert, events[0].OperationType)
			}
		})
	}
}

func TestResearchService_RemoveFeaturedResearch(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		setupFunc func(*MockResearchRepository)
		wantErr   bool
	}{
		{
			name:   "successfully remove featured research",
			userID: "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockResearchRepository) {
				featured := &FeaturedResearch{
					FeaturedResearchID: "featured-1",
					ResearchID:         "550e8400-e29b-41d4-a716-446655440001",
					CreatedOn:          time.Now(),
				}
				repo.featuredResearch["featured-1"] = featured
			},
			wantErr: false,
		},
		{
			name:      "return not found error when no featured research exists",
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockResearchRepository) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			repo := NewMockResearchRepository()
			tt.setupFunc(repo)
			
			service := NewResearchService(repo)
			
			err := service.RemoveFeaturedResearch(ctx, tt.userID)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify featured research was removed
				assert.Len(t, repo.featuredResearch, 0)
				// Verify audit event was published
				events := repo.GetAuditEvents()
				assert.Len(t, events, 1)
				assert.Equal(t, domain.EntityTypeFeaturedResearch, events[0].EntityType)
				assert.Equal(t, domain.AuditEventDelete, events[0].OperationType)
			}
		})
	}
}