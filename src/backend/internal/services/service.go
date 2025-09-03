package services

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/axiom-software-co/international-center/src/internal/shared/domain"
)

// ServicesService implements business logic for services operations
type ServicesService struct {
	repository *ServicesRepository
}

// NewServicesService creates a new services service
func NewServicesService(repository *ServicesRepository) *ServicesService {
	return &ServicesService{
		repository: repository,
	}
}

// Service operations

// GetService retrieves service by ID
func (s *ServicesService) GetService(ctx context.Context, serviceID string, userID string) (*Service, error) {
	if serviceID == "" {
		return nil, domain.NewValidationError("service ID cannot be empty")
	}

	service, err := s.repository.GetService(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	// Check access permissions - only published services are publicly accessible
	if err := s.checkServiceAccess(service, userID); err != nil {
		return nil, err
	}

	return service, nil
}

// GetServiceBySlug retrieves service by slug
func (s *ServicesService) GetServiceBySlug(ctx context.Context, slug string, userID string) (*Service, error) {
	if slug == "" {
		return nil, domain.NewValidationError("service slug cannot be empty")
	}

	service, err := s.repository.GetServiceBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	// Check access permissions
	if err := s.checkServiceAccess(service, userID); err != nil {
		return nil, err
	}

	return service, nil
}

// GetAllServices retrieves all services accessible to the user
func (s *ServicesService) GetAllServices(ctx context.Context, userID string) ([]*Service, error) {
	allServices, err := s.repository.GetAllServices(ctx)
	if err != nil {
		return nil, err
	}

	// Filter based on access permissions
	var accessibleServices []*Service
	for _, service := range allServices {
		if s.checkServiceAccess(service, userID) == nil {
			accessibleServices = append(accessibleServices, service)
		}
	}

	return accessibleServices, nil
}

// GetServicesByCategory retrieves services by category
func (s *ServicesService) GetServicesByCategory(ctx context.Context, categoryID string, userID string) ([]*Service, error) {
	if categoryID == "" {
		return nil, domain.NewValidationError("category ID cannot be empty")
	}

	// Verify category exists
	_, err := s.repository.GetServiceCategory(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	services, err := s.repository.GetServicesByCategory(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	// Filter based on access permissions
	var accessibleServices []*Service
	for _, service := range services {
		if s.checkServiceAccess(service, userID) == nil {
			accessibleServices = append(accessibleServices, service)
		}
	}

	return accessibleServices, nil
}

// GetPublishedServices retrieves only published services
func (s *ServicesService) GetPublishedServices(ctx context.Context, userID string) ([]*Service, error) {
	services, err := s.repository.GetServicesByPublishingStatus(ctx, PublishingStatusPublished)
	if err != nil {
		return nil, err
	}

	return services, nil
}

// CreateService creates new service
func (s *ServicesService) CreateService(ctx context.Context, title, description, slug string, categoryID string, deliveryMode DeliveryMode, userID string) (*Service, error) {
	// Validate input parameters
	if err := s.validateCreateServiceParams(title, description, slug, categoryID, deliveryMode, userID); err != nil {
		return nil, err
	}

	// Verify category exists
	_, err := s.repository.GetServiceCategory(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	// Check if slug is already in use
	existingService, err := s.repository.GetServiceBySlug(ctx, slug)
	if err == nil && existingService != nil {
		return nil, domain.NewConflictError(fmt.Sprintf("service with slug '%s' already exists", slug))
	}

	// Create service entity
	service, err := NewService(title, description, slug, categoryID, deliveryMode, userID)
	if err != nil {
		return nil, domain.NewInternalError("failed to create service entity", err)
	}

	// Save service metadata to state store
	err = s.repository.SaveService(ctx, service)
	if err != nil {
		return nil, domain.NewInternalError("failed to save service metadata", err)
	}

	// Publish audit event
	err = s.repository.PublishAuditEvent(ctx, domain.EntityTypeService, service.ServiceID, domain.AuditEventInsert, userID, nil, service)
	if err != nil {
		// Log error but don't fail the operation
	}

	return service, nil
}

// UpdateServiceDetails updates service title and description
func (s *ServicesService) UpdateServiceDetails(ctx context.Context, serviceID string, title, description string, userID string) (*Service, error) {
	// Get existing service
	service, err := s.repository.GetService(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	// Check permissions
	if !s.canModifyService(service, userID) {
		return nil, domain.NewForbiddenError("insufficient permissions to modify service")
	}

	// Store original data for audit
	originalService := *service

	// Update details
	err = service.UpdateDetails(title, description, userID)
	if err != nil {
		return nil, domain.NewInternalError("failed to update service details", err)
	}

	// Save updated service
	err = s.repository.SaveService(ctx, service)
	if err != nil {
		return nil, domain.NewInternalError("failed to save updated service", err)
	}

	// Publish audit event
	err = s.repository.PublishAuditEvent(ctx, domain.EntityTypeService, service.ServiceID, domain.AuditEventUpdate, userID, &originalService, service)
	if err != nil {
		// Log error but don't fail the operation
	}

	return service, nil
}

// SetServiceContentURL sets the content URL for a service
func (s *ServicesService) SetServiceContentURL(ctx context.Context, serviceID string, contentURL string, userID string) (*Service, error) {
	// Get existing service
	service, err := s.repository.GetService(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	// Check permissions
	if !s.canModifyService(service, userID) {
		return nil, domain.NewForbiddenError("insufficient permissions to modify service")
	}

	// Store original data for audit
	originalService := *service

	// Set content URL
	err = service.SetContentURL(contentURL, userID)
	if err != nil {
		return nil, domain.NewInternalError("failed to set content URL", err)
	}

	// Save updated service
	err = s.repository.SaveService(ctx, service)
	if err != nil {
		return nil, domain.NewInternalError("failed to save updated service", err)
	}

	// Publish audit event
	err = s.repository.PublishAuditEvent(ctx, domain.EntityTypeService, service.ServiceID, domain.AuditEventUpdate, userID, &originalService, service)
	if err != nil {
		// Log error but don't fail the operation
	}

	return service, nil
}

// PublishService publishes a service
func (s *ServicesService) PublishService(ctx context.Context, serviceID string, userID string) (*Service, error) {
	// Get existing service
	service, err := s.repository.GetService(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	// Check permissions
	if !s.canModifyService(service, userID) {
		return nil, domain.NewForbiddenError("insufficient permissions to publish service")
	}

	// Store original data for audit
	originalService := *service

	// Publish service
	err = service.Publish(userID)
	if err != nil {
		return nil, domain.NewInternalError("failed to publish service", err)
	}

	// Save updated service
	err = s.repository.SaveService(ctx, service)
	if err != nil {
		return nil, domain.NewInternalError("failed to save published service", err)
	}

	// Publish audit event
	err = s.repository.PublishAuditEvent(ctx, domain.EntityTypeService, service.ServiceID, domain.AuditEventUpdate, userID, &originalService, service)
	if err != nil {
		// Log error but don't fail the operation
	}

	return service, nil
}

// DeleteService soft deletes service
func (s *ServicesService) DeleteService(ctx context.Context, serviceID string, userID string) error {
	// Get existing service
	service, err := s.repository.GetService(ctx, serviceID)
	if err != nil {
		return err
	}

	// Check permissions
	if !s.canModifyService(service, userID) {
		return domain.NewForbiddenError("insufficient permissions to delete service")
	}

	// Store original data for audit
	originalService := *service

	// Delete service
	err = s.repository.DeleteService(ctx, serviceID, userID)
	if err != nil {
		return domain.NewInternalError("failed to delete service", err)
	}

	// Publish audit event
	err = s.repository.PublishAuditEvent(ctx, domain.EntityTypeService, service.ServiceID, domain.AuditEventDelete, userID, &originalService, nil)
	if err != nil {
		// Log error but don't fail the operation
	}

	return nil
}

// SearchServices performs service search
func (s *ServicesService) SearchServices(ctx context.Context, searchTerm string, userID string) ([]*Service, error) {
	if strings.TrimSpace(searchTerm) == "" {
		return s.GetAllServices(ctx, userID)
	}

	searchResults, err := s.repository.SearchServices(ctx, searchTerm)
	if err != nil {
		return nil, domain.NewInternalError("failed to search services", err)
	}

	// Filter based on access permissions
	var accessibleResults []*Service
	for _, service := range searchResults {
		if s.checkServiceAccess(service, userID) == nil {
			accessibleResults = append(accessibleResults, service)
		}
	}

	return accessibleResults, nil
}

// ServiceCategory operations

// GetServiceCategory retrieves service category by ID
func (s *ServicesService) GetServiceCategory(ctx context.Context, categoryID string, userID string) (*ServiceCategory, error) {
	if categoryID == "" {
		return nil, domain.NewValidationError("category ID cannot be empty")
	}

	category, err := s.repository.GetServiceCategory(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	return category, nil
}

// GetServiceCategoryBySlug retrieves service category by slug
func (s *ServicesService) GetServiceCategoryBySlug(ctx context.Context, slug string, userID string) (*ServiceCategory, error) {
	if slug == "" {
		return nil, domain.NewValidationError("category slug cannot be empty")
	}

	category, err := s.repository.GetServiceCategoryBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	return category, nil
}

// GetAllServiceCategories retrieves all service categories
func (s *ServicesService) GetAllServiceCategories(ctx context.Context, userID string) ([]*ServiceCategory, error) {
	categories, err := s.repository.GetAllServiceCategories(ctx)
	if err != nil {
		return nil, err
	}

	return categories, nil
}

// FeaturedCategory operations

// GetAllFeaturedCategories retrieves all featured categories
func (s *ServicesService) GetAllFeaturedCategories(ctx context.Context, userID string) ([]*FeaturedCategory, error) {
	featured, err := s.repository.GetAllFeaturedCategories(ctx)
	if err != nil {
		return nil, err
	}

	return featured, nil
}

// GetFeaturedCategoryByPosition retrieves featured category by position
func (s *ServicesService) GetFeaturedCategoryByPosition(ctx context.Context, position int, userID string) (*FeaturedCategory, error) {
	if position != 1 && position != 2 {
		return nil, domain.NewValidationError("feature position must be 1 or 2")
	}

	featured, err := s.repository.GetFeaturedCategoryByPosition(ctx, position)
	if err != nil {
		return nil, err
	}

	return featured, nil
}

// Content operations

// GetServiceContentDownload retrieves service content download URL
func (s *ServicesService) GetServiceContentDownload(ctx context.Context, serviceID string, userID string) (string, error) {
	service, err := s.GetService(ctx, serviceID, userID)
	if err != nil {
		return "", err
	}

	// Check if service has content
	if service.ContentURL == "" {
		return "", domain.NewConflictError("service does not have content available")
	}

	// Create temporary download URL (expires in 1 hour)
	downloadURL, err := s.repository.CreateServiceContentBlobURL(ctx, service.ContentURL, 60)
	if err != nil {
		return "", domain.NewInternalError("failed to create content download URL", err)
	}

	return downloadURL, nil
}

// UploadServiceContent uploads service content
func (s *ServicesService) UploadServiceContent(ctx context.Context, serviceID string, content []byte, contentType string, userID string) (string, error) {
	// Get service
	service, err := s.repository.GetService(ctx, serviceID)
	if err != nil {
		return "", err
	}

	// Check permissions
	if !s.canModifyService(service, userID) {
		return "", domain.NewForbiddenError("insufficient permissions to upload content")
	}

	// Validate content
	if len(content) == 0 {
		return "", domain.NewValidationError("content cannot be empty")
	}

	// Calculate content hash
	contentHash := s.calculateContentHash(content)

	// Generate storage path
	storagePath := generateContentBlobPath("development", serviceID, contentHash)

	// Upload to blob storage
	err = s.repository.UploadServiceContentBlob(ctx, storagePath, content, contentType)
	if err != nil {
		return "", domain.NewInternalError("failed to upload service content", err)
	}

	return storagePath, nil
}

// Private helper methods

// validateCreateServiceParams validates service creation parameters
func (s *ServicesService) validateCreateServiceParams(title, description, slug string, categoryID string, deliveryMode DeliveryMode, userID string) error {
	if strings.TrimSpace(title) == "" {
		return domain.NewValidationError("title cannot be empty")
	}

	if strings.TrimSpace(description) == "" {
		return domain.NewValidationError("description cannot be empty")
	}

	if strings.TrimSpace(slug) == "" {
		return domain.NewValidationError("slug cannot be empty")
	}

	if !isValidSlug(slug) {
		return domain.NewValidationError("slug must contain only lowercase letters, numbers, and hyphens")
	}

	if strings.TrimSpace(categoryID) == "" {
		return domain.NewValidationError("category ID cannot be empty")
	}

	if !isValidDeliveryMode(deliveryMode) {
		return domain.NewValidationError("invalid delivery mode")
	}

	if strings.TrimSpace(userID) == "" {
		return domain.NewValidationError("user ID cannot be empty")
	}

	return nil
}

// calculateContentHash calculates SHA-256 hash of content data
func (s *ServicesService) calculateContentHash(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// checkServiceAccess checks if user can access service based on publishing status
func (s *ServicesService) checkServiceAccess(service *Service, userID string) error {
	switch service.PublishingStatus {
	case PublishingStatusPublished:
		return nil // Anyone can access published services
	case PublishingStatusDraft, PublishingStatusArchived:
		if userID == "" {
			return domain.NewUnauthorizedError("authentication required for draft/archived services")
		}
		// In a real implementation, this would check user roles/permissions
		// For now, only the creator can access draft/archived services
		if service.CreatedBy != userID {
			return domain.NewForbiddenError("insufficient permissions to access draft/archived service")
		}
		return nil
	default:
		return domain.NewInternalError("invalid publishing status", nil)
	}
}

// canModifyService checks if user can modify service
func (s *ServicesService) canModifyService(service *Service, userID string) bool {
	if userID == "" {
		return false
	}

	// Only the creator can modify service
	// In a real implementation, this would check for admin roles too
	return service.CreatedBy == userID
}