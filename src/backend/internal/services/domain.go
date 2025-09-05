package services

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/google/uuid"
)

// Domain types matching TABLES-SERVICES.md schema
type DeliveryMode string
type PublishingStatus string

const (
	DeliveryModeMobile    DeliveryMode = "mobile_service"
	DeliveryModeOutpatient DeliveryMode = "outpatient_service"
	DeliveryModeInpatient  DeliveryMode = "inpatient_service"
)

const (
	PublishingStatusDraft     PublishingStatus = "draft"
	PublishingStatusPublished PublishingStatus = "published"
	PublishingStatusArchived  PublishingStatus = "archived"
)

// IsValid checks if the delivery mode is valid
func (d DeliveryMode) IsValid() bool {
	switch d {
	case DeliveryModeMobile, DeliveryModeOutpatient, DeliveryModeInpatient:
		return true
	default:
		return false
	}
}

// IsValid checks if the publishing status is valid
func (p PublishingStatus) IsValid() bool {
	switch p {
	case PublishingStatusDraft, PublishingStatusPublished, PublishingStatusArchived:
		return true
	default:
		return false
	}
}

// Service represents the main services entity matching TABLES-SERVICES.md
type Service struct {
	ServiceID        string           `json:"service_id"`
	Title            string           `json:"title"`
	Description      string           `json:"description"`
	Slug             string           `json:"slug"`
	ContentURL       string           `json:"content_url,omitempty"`
	CategoryID       string           `json:"category_id"`
	ImageURL         string           `json:"image_url,omitempty"`
	OrderNumber      int              `json:"order_number"`
	DeliveryMode     DeliveryMode     `json:"delivery_mode"`
	PublishingStatus PublishingStatus `json:"publishing_status"`
	CreatedOn        time.Time        `json:"created_on"`
	CreatedBy        string           `json:"created_by,omitempty"`
	ModifiedOn       *time.Time       `json:"modified_on,omitempty"`
	ModifiedBy       string           `json:"modified_by,omitempty"`
	IsDeleted        bool             `json:"is_deleted"`
	DeletedOn        *time.Time       `json:"deleted_on,omitempty"`
	DeletedBy        string           `json:"deleted_by,omitempty"`
}

// ServiceCategory represents service categories matching TABLES-SERVICES.md
type ServiceCategory struct {
	CategoryID            string     `json:"category_id"`
	Name                  string     `json:"name"`
	Slug                  string     `json:"slug"`
	OrderNumber           int        `json:"order_number"`
	IsDefaultUnassigned   bool       `json:"is_default_unassigned"`
	CreatedOn             time.Time  `json:"created_on"`
	CreatedBy             string     `json:"created_by,omitempty"`
	ModifiedOn            *time.Time `json:"modified_on,omitempty"`
	ModifiedBy            string     `json:"modified_by,omitempty"`
	IsDeleted             bool       `json:"is_deleted"`
	DeletedOn             *time.Time `json:"deleted_on,omitempty"`
	DeletedBy             string     `json:"deleted_by,omitempty"`
}

// FeaturedCategory represents featured categories matching TABLES-SERVICES.md
type FeaturedCategory struct {
	FeaturedCategoryID string     `json:"featured_category_id"`
	CategoryID         string     `json:"category_id"`
	FeaturePosition    int        `json:"feature_position"`
	CreatedOn          time.Time  `json:"created_on"`
	CreatedBy          string     `json:"created_by,omitempty"`
	ModifiedOn         *time.Time `json:"modified_on,omitempty"`
	ModifiedBy         string     `json:"modified_by,omitempty"`
}

// ServiceAuditEvent represents audit events for services domain
type ServiceAuditEvent struct {
	AuditID        string            `json:"audit_id"`
	EntityType     string            `json:"entity_type"`
	EntityID       string            `json:"entity_id"`
	OperationType  string            `json:"operation_type"`
	AuditTimestamp time.Time         `json:"audit_timestamp"`
	UserID         string            `json:"user_id"`
	CorrelationID  string            `json:"correlation_id"`
	TraceID        string            `json:"trace_id"`
	DataSnapshot   AuditDataSnapshot `json:"data_snapshot"`
	Environment    string            `json:"environment"`
}

// AuditDataSnapshot represents before/after data in audit events
type AuditDataSnapshot struct {
	Before interface{} `json:"before"`
	After  interface{} `json:"after"`
}

// Domain validation patterns
var (
	slugRegex = regexp.MustCompile(`^[a-z0-9-]+$`)
)

// Domain factory functions

// NewService creates a new service entity
func NewService(title, description, slug string, categoryID string, deliveryMode DeliveryMode, userID string) (*Service, error) {
	if err := validateNewServiceParams(title, description, slug, categoryID, deliveryMode); err != nil {
		return nil, err
	}

	serviceID := uuid.New().String()
	now := time.Now().UTC()

	return &Service{
		ServiceID:        serviceID,
		Title:            title,
		Description:      description,
		Slug:             slug,
		CategoryID:       categoryID,
		OrderNumber:      0,
		DeliveryMode:     deliveryMode,
		PublishingStatus: PublishingStatusDraft, // Default to draft
		CreatedOn:        now,
		CreatedBy:        userID,
		IsDeleted:        false,
	}, nil
}

// NewServiceCategory creates a new service category entity
func NewServiceCategory(name, slug string, isDefaultUnassigned bool, userID string) (*ServiceCategory, error) {
	if err := validateNewServiceCategoryParams(name, slug); err != nil {
		return nil, err
	}

	categoryID := uuid.New().String()
	now := time.Now().UTC()

	return &ServiceCategory{
		CategoryID:          categoryID,
		Name:                name,
		Slug:                slug,
		OrderNumber:         0,
		IsDefaultUnassigned: isDefaultUnassigned,
		CreatedOn:           now,
		CreatedBy:           userID,
		IsDeleted:           false,
	}, nil
}

// NewFeaturedCategory creates a new featured category entity
func NewFeaturedCategory(categoryID string, featurePosition int, userID string) (*FeaturedCategory, error) {
	if err := validateNewFeaturedCategoryParams(categoryID, featurePosition); err != nil {
		return nil, err
	}

	featuredCategoryID := uuid.New().String()
	now := time.Now().UTC()

	return &FeaturedCategory{
		FeaturedCategoryID: featuredCategoryID,
		CategoryID:         categoryID,
		FeaturePosition:    featurePosition,
		CreatedOn:          now,
		CreatedBy:          userID,
	}, nil
}

// Domain business logic methods for Service

func (s *Service) Publish(userID string) error {
	if s.PublishingStatus != PublishingStatusDraft {
		return errors.New("can only publish services with draft status")
	}

	s.PublishingStatus = PublishingStatusPublished
	s.ModifiedBy = userID
	now := time.Now().UTC()
	s.ModifiedOn = &now

	return nil
}

func (s *Service) Archive(userID string) error {
	if s.PublishingStatus != PublishingStatusPublished {
		return errors.New("can only archive published services")
	}

	s.PublishingStatus = PublishingStatusArchived
	s.ModifiedBy = userID
	now := time.Now().UTC()
	s.ModifiedOn = &now

	return nil
}

func (s *Service) UnArchive(userID string) error {
	if s.PublishingStatus != PublishingStatusArchived {
		return errors.New("can only unarchive archived services")
	}

	s.PublishingStatus = PublishingStatusDraft
	s.ModifiedBy = userID
	now := time.Now().UTC()
	s.ModifiedOn = &now

	return nil
}

func (s *Service) UpdateDetails(title, description string, userID string) error {
	if strings.TrimSpace(title) == "" {
		return errors.New("title cannot be empty")
	}

	if strings.TrimSpace(description) == "" {
		return errors.New("description cannot be empty")
	}

	s.Title = title
	s.Description = description
	s.ModifiedBy = userID
	now := time.Now().UTC()
	s.ModifiedOn = &now

	return nil
}

func (s *Service) SetContentURL(contentURL string, userID string) error {
	if contentURL != "" {
		if !strings.HasPrefix(contentURL, "https://") {
			return errors.New("content URL must be a valid HTTPS URL")
		}
	}

	s.ContentURL = contentURL
	s.ModifiedBy = userID
	now := time.Now().UTC()
	s.ModifiedOn = &now

	return nil
}

func (s *Service) SetImageURL(imageURL string, userID string) error {
	if imageURL != "" {
		if !strings.HasPrefix(imageURL, "https://") {
			return errors.New("image URL must be a valid HTTPS URL")
		}
	}

	s.ImageURL = imageURL
	s.ModifiedBy = userID
	now := time.Now().UTC()
	s.ModifiedOn = &now

	return nil
}

func (s *Service) SetOrderNumber(orderNumber int, userID string) error {
	s.OrderNumber = orderNumber
	s.ModifiedBy = userID
	now := time.Now().UTC()
	s.ModifiedOn = &now

	return nil
}

func (s *Service) ChangeCategory(categoryID string, userID string) error {
	if strings.TrimSpace(categoryID) == "" {
		return errors.New("category ID cannot be empty")
	}

	s.CategoryID = categoryID
	s.ModifiedBy = userID
	now := time.Now().UTC()
	s.ModifiedOn = &now

	return nil
}

func (s *Service) SetDeliveryMode(deliveryMode DeliveryMode, userID string) error {
	if !deliveryMode.IsValid() {
		return domain.NewValidationError("invalid delivery mode")
	}

	s.DeliveryMode = deliveryMode
	s.ModifiedBy = userID
	now := time.Now().UTC()
	s.ModifiedOn = &now

	return nil
}

func (s *Service) Delete(userID string) error {
	s.IsDeleted = true
	s.DeletedBy = userID
	now := time.Now().UTC()
	s.DeletedOn = &now

	return nil
}

// Domain business logic methods for ServiceCategory

func (sc *ServiceCategory) UpdateDetails(name string, userID string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("category name cannot be empty")
	}

	sc.Name = name
	sc.ModifiedBy = userID
	now := time.Now().UTC()
	sc.ModifiedOn = &now

	return nil
}

func (sc *ServiceCategory) SetOrderNumber(orderNumber int, userID string) error {
	sc.OrderNumber = orderNumber
	sc.ModifiedBy = userID
	now := time.Now().UTC()
	sc.ModifiedOn = &now

	return nil
}

func (sc *ServiceCategory) Delete(userID string) error {
	if sc.IsDefaultUnassigned {
		return errors.New("cannot delete the default unassigned category")
	}

	sc.IsDeleted = true
	sc.DeletedBy = userID
	now := time.Now().UTC()
	sc.DeletedOn = &now

	return nil
}

// Domain business logic methods for FeaturedCategory

func (fc *FeaturedCategory) ChangeCategory(categoryID string, userID string) error {
	if strings.TrimSpace(categoryID) == "" {
		return errors.New("category ID cannot be empty")
	}

	fc.CategoryID = categoryID
	fc.ModifiedBy = userID
	now := time.Now().UTC()
	fc.ModifiedOn = &now

	return nil
}

func (fc *FeaturedCategory) SetFeaturePosition(featurePosition int, userID string) error {
	if featurePosition != 1 && featurePosition != 2 {
		return errors.New("feature position must be 1 or 2")
	}

	fc.FeaturePosition = featurePosition
	fc.ModifiedBy = userID
	now := time.Now().UTC()
	fc.ModifiedOn = &now

	return nil
}

// Domain validation functions

// Validation helper functions matching Volunteers/News domain patterns

func validateServiceTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return domain.NewValidationError("title is required")
	}
	if len(title) < 2 || len(title) > 255 {
		return domain.NewValidationError("title must be between 2 and 255 characters")
	}
	return nil
}

func validateServiceDescription(description string) error {
	if strings.TrimSpace(description) == "" {
		return domain.NewValidationError("description is required")
	}
	if len(description) < 10 || len(description) > 2000 {
		return domain.NewValidationError("description must be between 10 and 2000 characters")
	}
	return nil
}

func validateServiceSlug(slug string) error {
	if strings.TrimSpace(slug) == "" {
		return domain.NewValidationError("slug is required")
	}
	if len(slug) < 2 || len(slug) > 255 {
		return domain.NewValidationError("slug must be between 2 and 255 characters")
	}
	if !slugRegex.MatchString(slug) {
		return domain.NewValidationError("slug must contain only lowercase letters, numbers, and hyphens")
	}
	return nil
}

func validateNewServiceParams(title, description, slug string, categoryID string, deliveryMode DeliveryMode) error {
	if err := validateServiceTitle(title); err != nil {
		return err
	}

	if err := validateServiceDescription(description); err != nil {
		return err
	}

	if err := validateServiceSlug(slug); err != nil {
		return err
	}

	if strings.TrimSpace(categoryID) == "" {
		return domain.NewValidationError("category ID cannot be empty")
	}

	if !deliveryMode.IsValid() {
		return domain.NewValidationError("invalid delivery mode")
	}

	return nil
}

func validateNewServiceCategoryParams(name, slug string) error {
	if strings.TrimSpace(name) == "" {
		return domain.NewValidationError("category name cannot be empty")
	}

	if err := validateServiceSlug(slug); err != nil {
		return err
	}

	return nil
}

func validateNewFeaturedCategoryParams(categoryID string, featurePosition int) error {
	if strings.TrimSpace(categoryID) == "" {
		return domain.NewValidationError("category ID cannot be empty")
	}

	if featurePosition != 1 && featurePosition != 2 {
		return domain.NewValidationError("feature position must be 1 or 2")
	}

	return nil
}

// Removed redundant validation functions:
// - isValidSlug: replaced by comprehensive validateServiceSlug helper
// - isValidDeliveryMode: replaced by DeliveryMode.IsValid() method  
// - isValidPublishingStatus: replaced by PublishingStatus.IsValid() method

func generateContentBlobPath(environment, serviceID, contentHash string) string {
	return fmt.Sprintf("blob-storage://%s/services/content/%s/%s.html", environment, serviceID, contentHash)
}