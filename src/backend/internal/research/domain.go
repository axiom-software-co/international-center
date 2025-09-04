package research

import (
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/google/uuid"
)

// Domain types matching TABLES-RESEARCH.md schema
type ResearchType string
type PublishingStatus string

const (
	ResearchTypeClinicalStudy    ResearchType = "clinical_study"
	ResearchTypeCaseReport       ResearchType = "case_report"
	ResearchTypeSystematicReview ResearchType = "systematic_review"
	ResearchTypeMetaAnalysis     ResearchType = "meta_analysis"
	ResearchTypeEditorial        ResearchType = "editorial"
	ResearchTypeCommentary       ResearchType = "commentary"
)

const (
	PublishingStatusDraft     PublishingStatus = "draft"
	PublishingStatusPublished PublishingStatus = "published"
	PublishingStatusArchived  PublishingStatus = "archived"
)

// Research represents the main research entity matching TABLES-RESEARCH.md
type Research struct {
	ResearchID        string           `json:"research_id"`
	Title             string           `json:"title"`
	Abstract          string           `json:"abstract"`
	Content           string           `json:"content,omitempty"`
	Slug              string           `json:"slug"`
	CategoryID        string           `json:"category_id"`
	ImageURL          string           `json:"image_url,omitempty"`
	AuthorNames       string           `json:"author_names"`
	PublicationDate   *time.Time       `json:"publication_date,omitempty"`
	DOI               string           `json:"doi,omitempty"`
	ExternalURL       string           `json:"external_url,omitempty"`
	ReportURL         string           `json:"report_url,omitempty"`
	PublishingStatus  PublishingStatus `json:"publishing_status"`
	Keywords          []string         `json:"keywords,omitempty"`
	ResearchType      ResearchType     `json:"research_type"`
	CreatedOn         time.Time        `json:"created_on"`
	CreatedBy         string           `json:"created_by,omitempty"`
	ModifiedOn        *time.Time       `json:"modified_on,omitempty"`
	ModifiedBy        string           `json:"modified_by,omitempty"`
	IsDeleted         bool             `json:"is_deleted"`
	DeletedOn         *time.Time       `json:"deleted_on,omitempty"`
	DeletedBy         string           `json:"deleted_by,omitempty"`
}

// ResearchCategory represents research categories matching TABLES-RESEARCH.md
type ResearchCategory struct {
	CategoryID          string     `json:"category_id"`
	Name                string     `json:"name"`
	Slug                string     `json:"slug"`
	Description         string     `json:"description,omitempty"`
	IsDefaultUnassigned bool       `json:"is_default_unassigned"`
	CreatedOn           time.Time  `json:"created_on"`
	CreatedBy           string     `json:"created_by,omitempty"`
	ModifiedOn          *time.Time `json:"modified_on,omitempty"`
	ModifiedBy          string     `json:"modified_by,omitempty"`
	IsDeleted           bool       `json:"is_deleted"`
	DeletedOn           *time.Time `json:"deleted_on,omitempty"`
	DeletedBy           string     `json:"deleted_by,omitempty"`
}

// FeaturedResearch represents featured research matching TABLES-RESEARCH.md
type FeaturedResearch struct {
	FeaturedResearchID string     `json:"featured_research_id"`
	ResearchID         string     `json:"research_id"`
	CreatedOn          time.Time  `json:"created_on"`
	CreatedBy          string     `json:"created_by,omitempty"`
	ModifiedOn         *time.Time `json:"modified_on,omitempty"`
	ModifiedBy         string     `json:"modified_by,omitempty"`
}

// Research validation methods

func (r *Research) Validate() error {
	if err := domain.ValidateUUID(r.ResearchID); err != nil {
		return domain.NewValidationFieldError("research_id", "research_id "+err.Error())
	}
	
	if err := domain.ValidateTitle(r.Title); err != nil {
		return err
	}
	
	if err := domain.ValidateRequiredString("abstract", r.Abstract); err != nil {
		return err
	}
	
	if err := domain.ValidateSlug(r.Slug); err != nil {
		return err
	}
	
	if err := domain.ValidateUUID(r.CategoryID); err != nil {
		return domain.NewValidationFieldError("category_id", "category_id "+err.Error())
	}
	
	if err := domain.ValidateRequiredStringWithLength("author_names", r.AuthorNames, 500); err != nil {
		return err
	}
	
	if err := domain.ValidateEnum("research_type", string(r.ResearchType), getValidResearchTypes()); err != nil {
		return err
	}
	
	if err := domain.ValidateEnum("publishing_status", string(r.PublishingStatus), getValidPublishingStatuses()); err != nil {
		return err
	}
	
	if err := domain.ValidateHTTPSURL("image_url", r.ImageURL); err != nil {
		return err
	}
	
	if err := domain.ValidateHTTPSURL("external_url", r.ExternalURL); err != nil {
		return err
	}
	
	if err := domain.ValidateHTTPSURL("report_url", r.ReportURL); err != nil {
		return err
	}
	
	return nil
}

func (r *Research) GenerateSlug() {
	if r.Slug == "" && r.Title != "" {
		r.Slug = domain.GenerateSlug(r.Title)
	}
}

func (r *Research) SetDefaults() {
	if r.ResearchID == "" {
		r.ResearchID = uuid.New().String()
	}
	
	if r.PublishingStatus == "" {
		r.PublishingStatus = PublishingStatusDraft
	}
	
	if r.ResearchType == "" {
		r.ResearchType = ResearchTypeClinicalStudy
	}
	
	if r.CreatedOn.IsZero() {
		r.CreatedOn = time.Now()
	}
	
	r.GenerateSlug()
}

func (r *Research) CanBePublished() error {
	if r.Title == "" {
		return domain.NewValidationError("cannot publish research without title")
	}
	
	if r.Abstract == "" {
		return domain.NewValidationError("cannot publish research without abstract")  
	}
	
	if r.AuthorNames == "" {
		return domain.NewValidationError("cannot publish research without author names")
	}
	
	if r.IsDeleted {
		return domain.NewValidationError("cannot publish deleted research")
	}
	
	return nil
}

// ResearchCategory validation methods

func (rc *ResearchCategory) Validate() error {
	if err := domain.ValidateUUID(rc.CategoryID); err != nil {
		return domain.NewValidationFieldError("category_id", "category_id "+err.Error())
	}
	
	if err := domain.ValidateRequiredStringWithLength("name", rc.Name, 255); err != nil {
		return err
	}
	
	if err := domain.ValidateSlug(rc.Slug); err != nil {
		return err
	}
	
	return nil
}

func (rc *ResearchCategory) GenerateSlug() {
	if rc.Slug == "" && rc.Name != "" {
		rc.Slug = domain.GenerateSlug(rc.Name)
	}
}

func (rc *ResearchCategory) SetDefaults() {
	if rc.CategoryID == "" {
		rc.CategoryID = uuid.New().String()
	}
	
	if rc.CreatedOn.IsZero() {
		rc.CreatedOn = time.Now()
	}
	
	rc.GenerateSlug()
}

// FeaturedResearch validation methods

func (fr *FeaturedResearch) Validate() error {
	if err := domain.ValidateUUID(fr.FeaturedResearchID); err != nil {
		return domain.NewValidationFieldError("featured_research_id", "featured_research_id "+err.Error())
	}
	
	if err := domain.ValidateUUID(fr.ResearchID); err != nil {
		return domain.NewValidationFieldError("research_id", "research_id "+err.Error())
	}
	
	return nil
}

func (fr *FeaturedResearch) SetDefaults() {
	if fr.FeaturedResearchID == "" {
		fr.FeaturedResearchID = uuid.New().String()
	}
	
	if fr.CreatedOn.IsZero() {
		fr.CreatedOn = time.Now()
	}
}

// Utility functions for enum validation

func getValidResearchTypes() []string {
	return []string{
		string(ResearchTypeClinicalStudy),
		string(ResearchTypeCaseReport),
		string(ResearchTypeSystematicReview),
		string(ResearchTypeMetaAnalysis),
		string(ResearchTypeEditorial),
		string(ResearchTypeCommentary),
	}
}

func getValidPublishingStatuses() []string {
	return []string{
		string(PublishingStatusDraft),
		string(PublishingStatusPublished),
		string(PublishingStatusArchived),
	}
}