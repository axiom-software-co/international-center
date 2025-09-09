package news

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
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

// IsValid checks if the news type is valid
func (n NewsType) IsValid() bool {
	switch n {
	case NewsTypeAnnouncement, NewsTypePressRelease, NewsTypeEvent, NewsTypeUpdate, NewsTypeAlert, NewsTypeFeature:
		return true
	default:
		return false
	}
}

// IsValid checks if the priority level is valid  
func (p PriorityLevel) IsValid() bool {
	switch p {
	case PriorityLevelLow, PriorityLevelNormal, PriorityLevelHigh, PriorityLevelUrgent:
		return true
	default:
		return false
	}
}

// IsValid checks if the publishing status is valid
func (s PublishingStatus) IsValid() bool {
	switch s {
	case PublishingStatusDraft, PublishingStatusPublished, PublishingStatusArchived:
		return true
	default:
		return false
	}
}

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

// Validation helper functions

var (
	slugRegex = regexp.MustCompile(`^[a-z0-9-]+$`)
)

func validateStringField(value, fieldName string, minLen, maxLen int) error {
	if strings.TrimSpace(value) == "" {
		return domain.NewValidationError(fieldName + " is required")
	}
	if len(value) < minLen || len(value) > maxLen {
		return domain.NewValidationError(fmt.Sprintf("%s must be between %d and %d characters", fieldName, minLen, maxLen))
	}
	return nil
}

func validateNewsTitle(title string) error {
	return validateStringField(title, "title", 3, 255)
}

func validateNewsSummary(summary string) error {
	return validateStringField(summary, "summary", 10, 1000)
}

func validateNewsSlug(slug string) error {
	if err := validateStringField(slug, "slug", 3, 255); err != nil {
		return err
	}
	if !slugRegex.MatchString(slug) {
		return domain.NewValidationError("slug must contain only lowercase letters, numbers, and hyphens")
	}
	return nil
}

func validateCategoryName(name string) error {
	return validateStringField(name, "category name", 2, 255)
}

func validateUserID(userID, fieldName string) error {
	if strings.TrimSpace(userID) == "" {
		return domain.NewValidationError(fieldName + " is required")
	}
	return nil
}

func generateSlugFromTitle(title string) string {
	slug := strings.ToLower(strings.TrimSpace(title))
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(slug, "")
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}

// Factory functions

// NewNews creates a new news entity with validation
func NewNews(title, summary string, newsType NewsType, categoryID, userID string) (*News, error) {
	if err := validateNewsTitle(title); err != nil {
		return nil, err
	}
	if err := validateNewsSummary(summary); err != nil {
		return nil, err
	}
	if !newsType.IsValid() {
		return nil, domain.NewValidationError("invalid news type")
	}
	if strings.TrimSpace(categoryID) == "" {
		return nil, domain.NewValidationError("category ID is required")
	}
	if err := validateUserID(userID, "created by"); err != nil {
		return nil, err
	}

	now := time.Now()
	slug := generateSlugFromTitle(title)

	return &News{
		NewsID:               uuid.New().String(),
		Title:                title,
		Summary:              summary,
		Slug:                 slug,
		CategoryID:           categoryID,
		NewsType:             newsType,
		PriorityLevel:        PriorityLevelNormal,
		PublishingStatus:     PublishingStatusDraft,
		PublicationTimestamp: now,
		CreatedOn:            now,
		CreatedBy:            userID,
		IsDeleted:            false,
	}, nil
}

// NewNewsCategory creates a new news category with validation
func NewNewsCategory(name, slug string, isDefaultUnassigned bool, userID string) (*NewsCategory, error) {
	if err := validateCategoryName(name); err != nil {
		return nil, err
	}
	if err := validateNewsSlug(slug); err != nil {
		return nil, err
	}
	if err := validateUserID(userID, "created by"); err != nil {
		return nil, err
	}

	now := time.Now()

	return &NewsCategory{
		CategoryID:          uuid.New().String(),
		Name:                name,
		Slug:                slug,
		IsDefaultUnassigned: isDefaultUnassigned,
		CreatedOn:           now,
		CreatedBy:           userID,
		IsDeleted:           false,
	}, nil
}

// NewFeaturedNews creates a new featured news with validation
func NewFeaturedNews(newsID, userID string) (*FeaturedNews, error) {
	if strings.TrimSpace(newsID) == "" {
		return nil, domain.NewValidationError("news ID is required")
	}
	if err := validateUserID(userID, "created by"); err != nil {
		return nil, err
	}

	now := time.Now()

	return &FeaturedNews{
		FeaturedNewsID: uuid.New().String(),
		NewsID:         newsID,
		CreatedOn:      now,
		CreatedBy:      userID,
	}, nil
}

// Business logic methods for News

// Publish transitions news from draft to published status
func (n *News) Publish(userID string) error {
	if n.PublishingStatus != PublishingStatusDraft {
		return errors.New("can only publish news with draft status")
	}

	n.PublishingStatus = PublishingStatusPublished
	n.ModifiedBy = userID
	now := time.Now()
	n.ModifiedOn = &now

	return nil
}

// Archive transitions news from published to archived status
func (n *News) Archive(userID string) error {
	if n.PublishingStatus != PublishingStatusPublished {
		return errors.New("can only archive published news")
	}

	n.PublishingStatus = PublishingStatusArchived
	n.ModifiedBy = userID
	now := time.Now()
	n.ModifiedOn = &now

	return nil
}

// UnArchive transitions news from archived to draft status
func (n *News) UnArchive(userID string) error {
	if n.PublishingStatus != PublishingStatusArchived {
		return errors.New("can only unarchive archived news")
	}

	n.PublishingStatus = PublishingStatusDraft
	n.ModifiedBy = userID
	now := time.Now()
	n.ModifiedOn = &now

	return nil
}

// UpdateDetails updates the title and summary of news
func (n *News) UpdateDetails(title, summary, userID string) error {
	if err := validateNewsTitle(title); err != nil {
		return err
	}
	if err := validateNewsSummary(summary); err != nil {
		return err
	}

	n.Title = title
	n.Summary = summary
	n.ModifiedBy = userID
	now := time.Now()
	n.ModifiedOn = &now

	return nil
}

// ValidateComprehensive validates the news entity comprehensively
func (n *News) ValidateComprehensive() error {
	if n.NewsID == "" {
		return domain.NewValidationError("news ID is required")
	}

	if err := validateNewsTitle(n.Title); err != nil {
		return err
	}

	if err := validateNewsSummary(n.Summary); err != nil {
		return err
	}

	if n.Slug != "" {
		if err := validateNewsSlug(n.Slug); err != nil {
			return err
		}
	}

	if strings.TrimSpace(n.CategoryID) == "" {
		return domain.NewValidationError("category ID is required")
	}

	if !n.NewsType.IsValid() {
		return domain.NewValidationError("invalid news type")
	}

	if !n.PriorityLevel.IsValid() {
		return domain.NewValidationError("invalid priority level")
	}

	if !n.PublishingStatus.IsValid() {
		return domain.NewValidationError("invalid publishing status")
	}

	if n.CreatedBy != "" {
		if err := validateUserID(n.CreatedBy, "created by"); err != nil {
			return err
		}
	}

	return nil
}

// CanTransitionTo checks if news can transition to target publishing status
func (n *News) CanTransitionTo(targetStatus PublishingStatus) error {
	if !targetStatus.IsValid() {
		return domain.NewValidationError("invalid target status")
	}

	switch n.PublishingStatus {
	case PublishingStatusDraft:
		if targetStatus == PublishingStatusPublished {
			return nil
		}
		return domain.NewValidationError(fmt.Sprintf("cannot transition from %s to %s", n.PublishingStatus, targetStatus))
	case PublishingStatusPublished:
		if targetStatus == PublishingStatusArchived {
			return nil
		}
		return domain.NewValidationError(fmt.Sprintf("cannot transition from %s to %s", n.PublishingStatus, targetStatus))
	case PublishingStatusArchived:
		if targetStatus == PublishingStatusDraft {
			return nil
		}
		return domain.NewValidationError(fmt.Sprintf("cannot transition from %s to %s", n.PublishingStatus, targetStatus))
	}

	return domain.NewValidationError(fmt.Sprintf("cannot transition from %s to %s", n.PublishingStatus, targetStatus))
}