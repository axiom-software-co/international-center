package news

import (
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/google/uuid"
)

// Domain types matching TABLES-NEWS.md schema
type NewsType string
type PriorityLevel string
type PublishingStatus string

const (
	NewsTypeAnnouncement  NewsType = "announcement"
	NewsTypePressRelease  NewsType = "press_release"
	NewsTypeEvent         NewsType = "event"
	NewsTypeUpdate        NewsType = "update"
	NewsTypeAlert         NewsType = "alert"
	NewsTypeFeature       NewsType = "feature"
)

const (
	PriorityLevelLow    PriorityLevel = "low"
	PriorityLevelNormal PriorityLevel = "normal"
	PriorityLevelHigh   PriorityLevel = "high"
	PriorityLevelUrgent PriorityLevel = "urgent"
)

const (
	PublishingStatusDraft     PublishingStatus = "draft"
	PublishingStatusPublished PublishingStatus = "published"
	PublishingStatusArchived  PublishingStatus = "archived"
)

// News represents the main news entity matching TABLES-NEWS.md
type News struct {
	NewsID              string           `json:"news_id"`
	Title               string           `json:"title"`
	Summary             string           `json:"summary"`
	Content             string           `json:"content,omitempty"`
	Slug                string           `json:"slug"`
	CategoryID          string           `json:"category_id"`
	ImageURL            string           `json:"image_url,omitempty"`
	AuthorName          string           `json:"author_name,omitempty"`
	PublicationTimestamp time.Time       `json:"publication_timestamp"`
	ExternalSource      string           `json:"external_source,omitempty"`
	ExternalURL         string           `json:"external_url,omitempty"`
	PublishingStatus    PublishingStatus `json:"publishing_status"`
	Tags                []string         `json:"tags,omitempty"`
	NewsType            NewsType         `json:"news_type"`
	PriorityLevel       PriorityLevel    `json:"priority_level"`
	CreatedOn           time.Time        `json:"created_on"`
	CreatedBy           string           `json:"created_by,omitempty"`
	ModifiedOn          *time.Time       `json:"modified_on,omitempty"`
	ModifiedBy          string           `json:"modified_by,omitempty"`
	IsDeleted           bool             `json:"is_deleted"`
	DeletedOn           *time.Time       `json:"deleted_on,omitempty"`
	DeletedBy           string           `json:"deleted_by,omitempty"`
}

// NewsCategory represents news categories matching TABLES-NEWS.md
type NewsCategory struct {
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

// FeaturedNews represents featured news matching TABLES-NEWS.md
type FeaturedNews struct {
	FeaturedNewsID string     `json:"featured_news_id"`
	NewsID         string     `json:"news_id"`
	CreatedOn      time.Time  `json:"created_on"`
	CreatedBy      string     `json:"created_by,omitempty"`
	ModifiedOn     *time.Time `json:"modified_on,omitempty"`
	ModifiedBy     string     `json:"modified_by,omitempty"`
}


// News validation methods

func (n *News) Validate() error {
	if err := domain.ValidateUUID(n.NewsID); err != nil {
		return domain.NewValidationFieldError("news_id", "news_id "+err.Error())
	}
	
	if err := domain.ValidateTitle(n.Title); err != nil {
		return err
	}
	
	if err := domain.ValidateRequiredString("summary", n.Summary); err != nil {
		return err
	}
	
	if err := domain.ValidateSlug(n.Slug); err != nil {
		return err
	}
	
	if err := domain.ValidateUUID(n.CategoryID); err != nil {
		return domain.NewValidationFieldError("category_id", "category_id "+err.Error())
	}
	
	if err := domain.ValidateEnum("news_type", string(n.NewsType), getValidNewsTypes()); err != nil {
		return err
	}
	
	if err := domain.ValidateEnum("priority_level", string(n.PriorityLevel), getValidPriorityLevels()); err != nil {
		return err
	}
	
	if err := domain.ValidateEnum("publishing_status", string(n.PublishingStatus), getValidPublishingStatuses()); err != nil {
		return err
	}
	
	if err := domain.ValidateHTTPSURL("image_url", n.ImageURL); err != nil {
		return err
	}
	
	if err := domain.ValidateHTTPSURL("external_url", n.ExternalURL); err != nil {
		return err
	}
	
	return nil
}

func (n *News) GenerateSlug() {
	if n.Slug == "" && n.Title != "" {
		n.Slug = domain.GenerateSlug(n.Title)
	}
}

func (n *News) SetDefaults() {
	if n.NewsID == "" {
		n.NewsID = uuid.New().String()
	}
	
	if n.PublishingStatus == "" {
		n.PublishingStatus = PublishingStatusDraft
	}
	
	if n.NewsType == "" {
		n.NewsType = NewsTypeAnnouncement
	}
	
	if n.PriorityLevel == "" {
		n.PriorityLevel = PriorityLevelNormal
	}
	
	if n.PublicationTimestamp.IsZero() {
		n.PublicationTimestamp = time.Now()
	}
	
	if n.CreatedOn.IsZero() {
		n.CreatedOn = time.Now()
	}
	
	n.GenerateSlug()
}

func (n *News) CanBePublished() error {
	if n.Title == "" {
		return domain.NewValidationError("cannot publish news without title")
	}
	
	if n.Summary == "" {
		return domain.NewValidationError("cannot publish news without summary")  
	}
	
	if n.IsDeleted {
		return domain.NewValidationError("cannot publish deleted news")
	}
	
	return nil
}

// NewsCategory validation methods

func (nc *NewsCategory) Validate() error {
	if err := domain.ValidateUUID(nc.CategoryID); err != nil {
		return domain.NewValidationFieldError("category_id", "category_id "+err.Error())
	}
	
	if err := domain.ValidateRequiredStringWithLength("name", nc.Name, 255); err != nil {
		return err
	}
	
	if err := domain.ValidateSlug(nc.Slug); err != nil {
		return err
	}
	
	return nil
}

func (nc *NewsCategory) GenerateSlug() {
	if nc.Slug == "" && nc.Name != "" {
		nc.Slug = domain.GenerateSlug(nc.Name)
	}
}

func (nc *NewsCategory) SetDefaults() {
	if nc.CategoryID == "" {
		nc.CategoryID = uuid.New().String()
	}
	
	if nc.CreatedOn.IsZero() {
		nc.CreatedOn = time.Now()
	}
	
	nc.GenerateSlug()
}

// FeaturedNews validation methods

func (fn *FeaturedNews) Validate() error {
	if err := domain.ValidateUUID(fn.FeaturedNewsID); err != nil {
		return domain.NewValidationFieldError("featured_news_id", "featured_news_id "+err.Error())
	}
	
	if err := domain.ValidateUUID(fn.NewsID); err != nil {
		return domain.NewValidationFieldError("news_id", "news_id "+err.Error())
	}
	
	return nil
}

func (fn *FeaturedNews) SetDefaults() {
	if fn.FeaturedNewsID == "" {
		fn.FeaturedNewsID = uuid.New().String()
	}
	
	if fn.CreatedOn.IsZero() {
		fn.CreatedOn = time.Now()
	}
}

// Utility functions for enum validation

func getValidNewsTypes() []string {
	return []string{
		string(NewsTypeAnnouncement),
		string(NewsTypePressRelease),
		string(NewsTypeEvent),
		string(NewsTypeUpdate),
		string(NewsTypeAlert),
		string(NewsTypeFeature),
	}
}

func getValidPriorityLevels() []string {
	return []string{
		string(PriorityLevelLow),
		string(PriorityLevelNormal),
		string(PriorityLevelHigh),
		string(PriorityLevelUrgent),
	}
}

func getValidPublishingStatuses() []string {
	return []string{
		string(PublishingStatusDraft),
		string(PublishingStatusPublished),
		string(PublishingStatusArchived),
	}
}