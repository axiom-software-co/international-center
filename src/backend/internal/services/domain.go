package services

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

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
	if !isValidDeliveryMode(deliveryMode) {
		return errors.New("invalid delivery mode")
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

func validateNewServiceParams(title, description, slug string, categoryID string, deliveryMode DeliveryMode) error {
	if strings.TrimSpace(title) == "" {
		return errors.New("title cannot be empty")
	}

	if strings.TrimSpace(description) == "" {
		return errors.New("description cannot be empty")
	}

	if strings.TrimSpace(slug) == "" {
		return errors.New("slug cannot be empty")
	}

	if !isValidSlug(slug) {
		return errors.New("slug must contain only lowercase letters, numbers, and hyphens")
	}

	if strings.TrimSpace(categoryID) == "" {
		return errors.New("category ID cannot be empty")
	}

	if !isValidDeliveryMode(deliveryMode) {
		return errors.New("invalid delivery mode")
	}

	return nil
}

func validateNewServiceCategoryParams(name, slug string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("category name cannot be empty")
	}

	if strings.TrimSpace(slug) == "" {
		return errors.New("slug cannot be empty")
	}

	if !isValidSlug(slug) {
		return errors.New("slug must contain only lowercase letters, numbers, and hyphens")
	}

	return nil
}

func validateNewFeaturedCategoryParams(categoryID string, featurePosition int) error {
	if strings.TrimSpace(categoryID) == "" {
		return errors.New("category ID cannot be empty")
	}

	if featurePosition != 1 && featurePosition != 2 {
		return errors.New("feature position must be 1 or 2")
	}

	return nil
}

func isValidSlug(slug string) bool {
	return slugRegex.MatchString(slug)
}

func isValidDeliveryMode(mode DeliveryMode) bool {
	switch mode {
	case DeliveryModeMobile, DeliveryModeOutpatient, DeliveryModeInpatient:
		return true
	default:
		return false
	}
}

func isValidPublishingStatus(status PublishingStatus) bool {
	switch status {
	case PublishingStatusDraft, PublishingStatusPublished, PublishingStatusArchived:
		return true
	default:
		return false
	}
}

func generateContentBlobPath(environment, serviceID, contentHash string) string {
	return fmt.Sprintf("blob-storage://%s/services/content/%s/%s.html", environment, serviceID, contentHash)
}