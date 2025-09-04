package research

import (
	"context"
	"strings"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
)

// ResearchRepositoryInterface defines the contract for research data access
type ResearchRepositoryInterface interface {
	// Research operations
	SaveResearch(ctx context.Context, research *Research) error
	GetResearch(ctx context.Context, researchID string) (*Research, error)
	GetResearchBySlug(ctx context.Context, slug string) (*Research, error)
	GetAllResearch(ctx context.Context, limit, offset int) ([]*Research, error)
	GetResearchByCategory(ctx context.Context, categoryID string, limit, offset int) ([]*Research, error)
	GetResearchByPublishingStatus(ctx context.Context, status PublishingStatus, limit, offset int) ([]*Research, error)
	DeleteResearch(ctx context.Context, researchID string) error
	SearchResearch(ctx context.Context, searchTerm string, limit, offset int) ([]*Research, error)

	// Research category operations
	SaveResearchCategory(ctx context.Context, category *ResearchCategory) error
	GetResearchCategory(ctx context.Context, categoryID string) (*ResearchCategory, error)
	GetResearchCategoryBySlug(ctx context.Context, slug string) (*ResearchCategory, error)
	GetAllResearchCategories(ctx context.Context) ([]*ResearchCategory, error)
	GetDefaultUnassignedCategory(ctx context.Context) (*ResearchCategory, error)
	DeleteResearchCategory(ctx context.Context, categoryID string) error

	// Featured research operations
	SaveFeaturedResearch(ctx context.Context, featured *FeaturedResearch) error
	GetFeaturedResearch(ctx context.Context) (*FeaturedResearch, error)
	DeleteFeaturedResearch(ctx context.Context, featuredResearchID string) error

	// Audit operations
	PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error
	GetResearchAudit(ctx context.Context, researchID string, userID string, limit int, offset int) ([]*domain.AuditEvent, error)
	GetResearchCategoryAudit(ctx context.Context, categoryID string, userID string, limit int, offset int) ([]*domain.AuditEvent, error)
}

// ResearchService implements business logic for research operations
type ResearchService struct {
	repository ResearchRepositoryInterface
}

// NewResearchService creates a new research service
func NewResearchService(repository ResearchRepositoryInterface) *ResearchService {
	return &ResearchService{
		repository: repository,
	}
}

// Research operations

func (s *ResearchService) GetResearch(ctx context.Context, researchID string, userID string) (*Research, error) {
	research, err := s.repository.GetResearch(ctx, researchID)
	if err != nil {
		return nil, err
	}

	// Publish audit event for access
	s.repository.PublishAuditEvent(ctx, domain.EntityTypeResearch, researchID, domain.AuditEventAccess, userID, nil, research)

	return research, nil
}

func (s *ResearchService) GetResearchBySlug(ctx context.Context, slug string, userID string) (*Research, error) {
	research, err := s.repository.GetResearchBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	// Publish audit event for access
	s.repository.PublishAuditEvent(ctx, domain.EntityTypeResearch, research.ResearchID, domain.AuditEventAccess, userID, nil, research)

	return research, nil
}

func (s *ResearchService) GetAllResearch(ctx context.Context, limit, offset int) ([]*Research, error) {
	return s.repository.GetAllResearch(ctx, limit, offset)
}

func (s *ResearchService) GetResearchByCategory(ctx context.Context, categoryID string, limit, offset int) ([]*Research, error) {
	return s.repository.GetResearchByCategory(ctx, categoryID, limit, offset)
}

func (s *ResearchService) SearchResearch(ctx context.Context, query string, limit, offset int) ([]*Research, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return make([]*Research, 0), nil
	}

	return s.repository.SearchResearch(ctx, query, limit, offset)
}

func (s *ResearchService) CreateResearch(ctx context.Context, research *Research, userID string) error {
	// Set defaults and validate
	research.SetDefaults()
	research.CreatedBy = userID
	
	if err := research.Validate(); err != nil {
		return err
	}

	// Save research
	if err := s.repository.SaveResearch(ctx, research); err != nil {
		return err
	}

	// Publish audit event
	return s.repository.PublishAuditEvent(ctx, domain.EntityTypeResearch, research.ResearchID, domain.AuditEventInsert, userID, nil, research)
}

func (s *ResearchService) UpdateResearch(ctx context.Context, research *Research, userID string) error {
	// Get existing research for audit
	existing, err := s.repository.GetResearch(ctx, research.ResearchID)
	if err != nil {
		return err
	}

	// Set modification fields and validate
	research.ModifiedBy = userID
	now := research.CreatedOn // Use existing created time
	research.ModifiedOn = &now
	
	if err := research.Validate(); err != nil {
		return err
	}

	// Save updated research
	if err := s.repository.SaveResearch(ctx, research); err != nil {
		return err
	}

	// Publish audit event
	return s.repository.PublishAuditEvent(ctx, domain.EntityTypeResearch, research.ResearchID, domain.AuditEventUpdate, userID, existing, research)
}

func (s *ResearchService) PublishResearch(ctx context.Context, researchID string, userID string) error {
	// Get existing research
	research, err := s.repository.GetResearch(ctx, researchID)
	if err != nil {
		return err
	}

	existing := *research // Copy for audit

	// Validate can be published
	if err := research.CanBePublished(); err != nil {
		return err
	}

	// Update publishing status
	research.PublishingStatus = PublishingStatusPublished
	research.ModifiedBy = userID
	now := research.CreatedOn
	research.ModifiedOn = &now

	// Save updated research
	if err := s.repository.SaveResearch(ctx, research); err != nil {
		return err
	}

	// Publish audit event
	return s.repository.PublishAuditEvent(ctx, domain.EntityTypeResearch, researchID, domain.AuditEventPublish, userID, &existing, research)
}

func (s *ResearchService) ArchiveResearch(ctx context.Context, researchID string, userID string) error {
	// Get existing research
	research, err := s.repository.GetResearch(ctx, researchID)
	if err != nil {
		return err
	}

	existing := *research // Copy for audit

	// Update publishing status
	research.PublishingStatus = PublishingStatusArchived
	research.ModifiedBy = userID
	now := research.CreatedOn
	research.ModifiedOn = &now

	// Save updated research
	if err := s.repository.SaveResearch(ctx, research); err != nil {
		return err
	}

	// Publish audit event
	return s.repository.PublishAuditEvent(ctx, domain.EntityTypeResearch, researchID, domain.AuditEventArchive, userID, &existing, research)
}

func (s *ResearchService) DeleteResearch(ctx context.Context, researchID string, userID string) error {
	// Get existing research for audit
	existing, err := s.repository.GetResearch(ctx, researchID)
	if err != nil {
		return err
	}

	// Soft delete research
	if err := s.repository.DeleteResearch(ctx, researchID); err != nil {
		return err
	}

	// Publish audit event
	return s.repository.PublishAuditEvent(ctx, domain.EntityTypeResearch, researchID, domain.AuditEventDelete, userID, existing, nil)
}

// Research category operations

func (s *ResearchService) GetAllResearchCategories(ctx context.Context) ([]*ResearchCategory, error) {
	return s.repository.GetAllResearchCategories(ctx)
}

func (s *ResearchService) GetResearchCategory(ctx context.Context, categoryID string) (*ResearchCategory, error) {
	return s.repository.GetResearchCategory(ctx, categoryID)
}

func (s *ResearchService) GetResearchCategoryBySlug(ctx context.Context, slug string) (*ResearchCategory, error) {
	return s.repository.GetResearchCategoryBySlug(ctx, slug)
}

func (s *ResearchService) CreateResearchCategory(ctx context.Context, category *ResearchCategory, userID string) error {
	// Set defaults and validate
	category.SetDefaults()
	category.CreatedBy = userID
	
	if err := category.Validate(); err != nil {
		return err
	}

	// Save category
	if err := s.repository.SaveResearchCategory(ctx, category); err != nil {
		return err
	}

	// Publish audit event
	return s.repository.PublishAuditEvent(ctx, domain.EntityTypeResearchCategory, category.CategoryID, domain.AuditEventInsert, userID, nil, category)
}

func (s *ResearchService) UpdateResearchCategory(ctx context.Context, category *ResearchCategory, userID string) error {
	// Get existing category for audit
	existing, err := s.repository.GetResearchCategory(ctx, category.CategoryID)
	if err != nil {
		return err
	}

	// Set modification fields and validate
	category.ModifiedBy = userID
	now := category.CreatedOn
	category.ModifiedOn = &now
	
	if err := category.Validate(); err != nil {
		return err
	}

	// Save updated category
	if err := s.repository.SaveResearchCategory(ctx, category); err != nil {
		return err
	}

	// Publish audit event
	return s.repository.PublishAuditEvent(ctx, domain.EntityTypeResearchCategory, category.CategoryID, domain.AuditEventUpdate, userID, existing, category)
}

func (s *ResearchService) DeleteResearchCategory(ctx context.Context, categoryID string, userID string) error {
	// Get existing category for audit
	existing, err := s.repository.GetResearchCategory(ctx, categoryID)
	if err != nil {
		return err
	}

	// Cannot delete default unassigned category
	if existing.IsDefaultUnassigned {
		return domain.NewValidationError("cannot delete default unassigned category")
	}

	// Soft delete category
	if err := s.repository.DeleteResearchCategory(ctx, categoryID); err != nil {
		return err
	}

	// Publish audit event
	return s.repository.PublishAuditEvent(ctx, domain.EntityTypeResearchCategory, categoryID, domain.AuditEventDelete, userID, existing, nil)
}

// Featured research operations

func (s *ResearchService) GetFeaturedResearch(ctx context.Context) (*FeaturedResearch, error) {
	return s.repository.GetFeaturedResearch(ctx)
}

func (s *ResearchService) SetFeaturedResearch(ctx context.Context, researchID string, userID string) error {
	// Validate research exists and can be featured
	research, err := s.repository.GetResearch(ctx, researchID)
	if err != nil {
		return err
	}

	if research.PublishingStatus != PublishingStatusPublished {
		return domain.NewValidationError("can only feature published research")
	}

	// Create featured research
	featured := &FeaturedResearch{
		ResearchID: researchID,
		CreatedBy:  userID,
	}
	featured.SetDefaults()

	// Save featured research (will replace existing due to constraint)
	if err := s.repository.SaveFeaturedResearch(ctx, featured); err != nil {
		return err
	}

	// Publish audit event
	return s.repository.PublishAuditEvent(ctx, domain.EntityTypeFeaturedResearch, featured.FeaturedResearchID, domain.AuditEventInsert, userID, nil, featured)
}

func (s *ResearchService) RemoveFeaturedResearch(ctx context.Context, userID string) error {
	// Get existing featured research
	existing, err := s.repository.GetFeaturedResearch(ctx)
	if err != nil {
		return err
	}

	// Delete featured research
	if err := s.repository.DeleteFeaturedResearch(ctx, existing.FeaturedResearchID); err != nil {
		return err
	}

	// Publish audit event
	return s.repository.PublishAuditEvent(ctx, domain.EntityTypeFeaturedResearch, existing.FeaturedResearchID, domain.AuditEventDelete, userID, existing, nil)
}

// Audit operations

func (s *ResearchService) GetResearchAudit(ctx context.Context, researchID string, userID string, limit int, offset int) ([]*domain.AuditEvent, error) {
	// Check if user has admin privileges
	if !strings.HasPrefix(userID, "admin-") {
		return nil, domain.NewUnauthorizedError("audit access requires admin privileges")
	}

	return s.repository.GetResearchAudit(ctx, researchID, userID, limit, offset)
}

func (s *ResearchService) GetResearchCategoryAudit(ctx context.Context, categoryID string, userID string, limit int, offset int) ([]*domain.AuditEvent, error) {
	// Check if user has admin privileges
	if !strings.HasPrefix(userID, "admin-") {
		return nil, domain.NewUnauthorizedError("audit access requires admin privileges")
	}

	return s.repository.GetResearchCategoryAudit(ctx, categoryID, userID, limit, offset)
}