package services

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

type DeliveryMode string
type PublishingStatus string

const (
	DeliveryMobilService     DeliveryMode = "mobile_service"
	DeliveryOutpatientService DeliveryMode = "outpatient_service"
	DeliveryInpatientService  DeliveryMode = "inpatient_service"
)

const (
	PublishingStatusDraft     PublishingStatus = "draft"
	PublishingStatusPublished PublishingStatus = "published"
	PublishingStatusArchived  PublishingStatus = "archived"
)

type Service struct {
	ServiceID        string
	Title            string
	Description      string
	Slug             string
	ContentURL       string
	CategoryID       string
	ImageURL         string
	OrderNumber      int
	DeliveryMode     DeliveryMode
	PublishingStatus PublishingStatus
	CreatedOn        time.Time
	CreatedBy        string
	ModifiedOn       time.Time
	ModifiedBy       string
	IsDeleted        bool
	DeletedOn        time.Time
	DeletedBy        string
}

type ServiceCategory struct {
	CategoryID            string
	Name                  string
	Slug                  string
	OrderNumber           int
	IsDefaultUnassigned   bool
	CreatedOn             time.Time
	CreatedBy             string
	ModifiedOn            time.Time
	ModifiedBy            string
	IsDeleted             bool
	DeletedOn             time.Time
	DeletedBy             string
}

var slugRegex = regexp.MustCompile(`^[a-z0-9-]+$`)

func NewService(title, description, slug, deliveryMode string) (*Service, error) {
	if strings.TrimSpace(title) == "" {
		return nil, errors.New("title cannot be empty")
	}
	
	if strings.TrimSpace(description) == "" {
		return nil, errors.New("description cannot be empty")
	}
	
	if strings.TrimSpace(slug) == "" {
		return nil, errors.New("slug cannot be empty")
	}
	
	if !isValidSlug(slug) {
		return nil, errors.New("slug must contain only lowercase letters, numbers, and hyphens")
	}
	
	deliveryModeEnum := DeliveryMode(deliveryMode)
	if !isValidDeliveryMode(deliveryModeEnum) {
		return nil, errors.New("invalid delivery mode")
	}
	
	return &Service{
		Title:            title,
		Description:      description,
		Slug:             slug,
		DeliveryMode:     deliveryModeEnum,
		PublishingStatus: PublishingStatusDraft,
		IsDeleted:        false,
		CreatedOn:        time.Now(),
	}, nil
}

func (s *Service) Publish(userID string) error {
	if s.PublishingStatus != PublishingStatusDraft {
		return errors.New("can only publish services with draft status")
	}
	
	s.PublishingStatus = PublishingStatusPublished
	s.ModifiedBy = userID
	s.ModifiedOn = time.Now()
	
	return nil
}

func (s *Service) Archive(userID string) error {
	if s.PublishingStatus != PublishingStatusPublished {
		return errors.New("can only archive services with published status")
	}
	
	s.PublishingStatus = PublishingStatusArchived
	s.ModifiedBy = userID
	s.ModifiedOn = time.Now()
	
	return nil
}

func (s *Service) AssignCategory(categoryID, userID string) error {
	if strings.TrimSpace(categoryID) == "" {
		return errors.New("category ID cannot be empty")
	}
	
	if !isValidUUID(categoryID) {
		return errors.New("category ID must be a valid UUID")
	}
	
	s.CategoryID = categoryID
	s.ModifiedBy = userID
	s.ModifiedOn = time.Now()
	
	return nil
}

func isValidSlug(slug string) bool {
	return slugRegex.MatchString(slug)
}

func isValidDeliveryMode(mode DeliveryMode) bool {
	switch mode {
	case DeliveryMobilService, DeliveryOutpatientService, DeliveryInpatientService:
		return true
	default:
		return false
	}
}

func isValidUUID(uuid string) bool {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return uuidRegex.MatchString(uuid)
}