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
			researchID: "research-1",
			setupFunc: func(repo *MockResearchRepository) {
				research := &Research{
					ResearchID: "research-1",
					Title:      "Clinical Study Results",
					Abstract:   "This study examines clinical outcomes",
					Slug:       "clinical-study-results",
					IsDeleted:  false,
				}
				repo.research["research-1"] = research
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
			
			result, err := service.GetResearch(ctx, tt.researchID, "user-1")
			
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
				repo.research["research-1"] = &Research{ResearchID: "research-1", Title: "Research 1", IsDeleted: false}
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
				repo.research["research-1"] = &Research{ResearchID: "research-1", Title: "Active Research", IsDeleted: false}
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
				repo.research["research-1"] = &Research{
					ResearchID: "research-1",
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
				repo.research["research-1"] = &Research{
					ResearchID: "research-1",
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
				repo.research["research-1"] = &Research{
					ResearchID: "research-1",
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
					ResearchID:         "research-1", 
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
			researchID: "research-1",
			userID:     "admin-1",
			setupFunc: func(repo *MockResearchRepository) {
				repo.auditEvents = append(repo.auditEvents, MockAuditEvent{
					EntityType:    domain.EntityTypeResearch,
					EntityID:      "research-1",
					OperationType: domain.AuditEventInsert,
					UserID:        "admin-1",
				})
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name:       "return empty array for research with no audit events",
			researchID: "research-2",
			userID:     "admin-1",
			setupFunc:  func(repo *MockResearchRepository) {},
			wantErr:    false,
			wantCount:  0,
		},
		{
			name:       "return unauthorized error for non-admin user",
			researchID: "research-1",
			userID:     "user-1",
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
			categoryID: "category-1",
			userID:     "admin-1",
			setupFunc: func(repo *MockResearchRepository) {
				repo.auditEvents = append(repo.auditEvents, MockAuditEvent{
					EntityType:    domain.EntityTypeCategory,
					EntityID:      "category-1",
					OperationType: domain.AuditEventUpdate,
					UserID:        "admin-1",
				})
			},
			wantErr: false,
		},
		{
			name:       "return unauthorized error for non-admin user",
			categoryID: "category-1",
			userID:     "user-1",
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