package business

import (
	"context"
	"net"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/google/uuid"
)

// BusinessRepositoryInterface defines the interface for business inquiry data operations
type BusinessRepositoryInterface interface {
	// Business inquiry operations
	SaveInquiry(ctx context.Context, inquiry *BusinessInquiry) error
	GetInquiry(ctx context.Context, inquiryID string) (*BusinessInquiry, error)
	DeleteInquiry(ctx context.Context, inquiryID string, userID string) error
	ListInquiries(ctx context.Context, filters InquiryFilters) ([]*BusinessInquiry, error)
	
	// Audit operations
	PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error
}

// BusinessService provides business logic for business inquiry operations
type BusinessService struct {
	repository BusinessRepositoryInterface
}

// NewBusinessService creates a new business service instance
func NewBusinessService(repository BusinessRepositoryInterface) *BusinessService {
	return &BusinessService{
		repository: repository,
	}
}

// AdminCreateInquiry creates a new business inquiry (admin only)
func (s *BusinessService) AdminCreateInquiry(ctx context.Context, request AdminCreateInquiryRequest, userID string) (*BusinessInquiry, error) {
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required to create business inquiries")
	}

	if request.OrganizationName == "" {
		return nil, domain.NewValidationError("organization name is required")
	}

	if len(request.Message) < 20 || len(request.Message) > 1500 {
		return nil, domain.NewValidationError("message must be between 20 and 1500 characters")
	}

	if request.ContactName == "" {
		return nil, domain.NewValidationError("contact name is required")
	}

	if request.Email == "" {
		return nil, domain.NewValidationError("email is required")
	}

	if request.Title == "" {
		return nil, domain.NewValidationError("title is required")
	}

	if request.InquiryType == "" {
		return nil, domain.NewValidationError("inquiry type is required")
	}

	inquiryType := InquiryType(request.InquiryType)
	if !IsValidInquiryType(inquiryType) {
		return nil, domain.NewValidationError("invalid inquiry type")
	}

	now := time.Now().UTC()
	inquiry := &BusinessInquiry{
		InquiryID:        uuid.New().String(),
		Status:           InquiryStatusNew,
		Priority:         InquiryPriorityMedium,
		OrganizationName: request.OrganizationName,
		ContactName:      request.ContactName,
		Title:            request.Title,
		Email:            request.Email,
		Phone:            request.Phone,
		Industry:         request.Industry,
		InquiryType:      inquiryType,
		Message:          request.Message,
		Source:           "website",
		CreatedAt:        now,
		UpdatedAt:        now,
		CreatedBy:        userID,
		UpdatedBy:        userID,
		IsDeleted:        false,
	}

	if request.Source != "" {
		inquiry.Source = request.Source
	}

	if request.IPAddress != nil {
		ip := net.ParseIP(*request.IPAddress)
		if ip != nil {
			inquiry.IPAddress = &ip
		}
	}

	if request.UserAgent != nil {
		inquiry.UserAgent = request.UserAgent
	}

	if err := inquiry.Validate(); err != nil {
		return nil, err
	}

	if err := s.repository.SaveInquiry(ctx, inquiry); err != nil {
		return nil, domain.NewInternalError("failed to save business inquiry", err)
	}

	if err := s.repository.PublishAuditEvent(ctx, domain.EntityTypeBusinessInquiry, inquiry.InquiryID, domain.AuditEventInsert, userID, nil, inquiry); err != nil {
		return nil, domain.NewInternalError("failed to publish audit event", err)
	}

	return inquiry, nil
}

// AdminUpdateInquiry updates an existing business inquiry (admin only)
func (s *BusinessService) AdminUpdateInquiry(ctx context.Context, inquiryID string, request AdminUpdateInquiryRequest, userID string) (*BusinessInquiry, error) {
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required to update business inquiries")
	}

	inquiry, err := s.repository.GetInquiry(ctx, inquiryID)
	if err != nil {
		return nil, err
	}

	beforeData := *inquiry

	if request.OrganizationName != nil {
		if *request.OrganizationName == "" {
			return nil, domain.NewValidationError("organization name cannot be empty")
		}
		inquiry.OrganizationName = *request.OrganizationName
	}

	if request.ContactName != nil {
		if *request.ContactName == "" {
			return nil, domain.NewValidationError("contact name cannot be empty")
		}
		inquiry.ContactName = *request.ContactName
	}

	if request.Title != nil {
		if *request.Title == "" {
			return nil, domain.NewValidationError("title cannot be empty")
		}
		inquiry.Title = *request.Title
	}

	if request.Email != nil {
		if *request.Email == "" {
			return nil, domain.NewValidationError("email cannot be empty")
		}
		inquiry.Email = *request.Email
	}

	if request.Phone != nil {
		inquiry.Phone = request.Phone
	}

	if request.Industry != nil {
		inquiry.Industry = request.Industry
	}

	if request.InquiryType != nil {
		inquiryType := InquiryType(*request.InquiryType)
		if !IsValidInquiryType(inquiryType) {
			return nil, domain.NewValidationError("invalid inquiry type")
		}
		inquiry.InquiryType = inquiryType
	}

	if request.Message != nil {
		if len(*request.Message) < 20 || len(*request.Message) > 1500 {
			return nil, domain.NewValidationError("message must be between 20 and 1500 characters")
		}
		inquiry.Message = *request.Message
	}

	inquiry.UpdatedAt = time.Now().UTC()
	inquiry.UpdatedBy = userID

	if err := inquiry.Validate(); err != nil {
		return nil, err
	}

	if err := s.repository.SaveInquiry(ctx, inquiry); err != nil {
		return nil, domain.NewInternalError("failed to update business inquiry", err)
	}

	if err := s.repository.PublishAuditEvent(ctx, domain.EntityTypeBusinessInquiry, inquiry.InquiryID, domain.AuditEventUpdate, userID, beforeData, *inquiry); err != nil {
		return nil, domain.NewInternalError("failed to publish audit event", err)
	}

	return inquiry, nil
}

// AdminDeleteInquiry soft deletes a business inquiry (admin only)
func (s *BusinessService) AdminDeleteInquiry(ctx context.Context, inquiryID string, userID string) error {
	if !IsAdminUser(userID) {
		return domain.NewUnauthorizedError("admin privileges required to delete business inquiries")
	}

	inquiry, err := s.repository.GetInquiry(ctx, inquiryID)
	if err != nil {
		return err
	}

	beforeData := *inquiry

	if err := s.repository.DeleteInquiry(ctx, inquiryID, userID); err != nil {
		return domain.NewInternalError("failed to delete business inquiry", err)
	}

	if err := s.repository.PublishAuditEvent(ctx, domain.EntityTypeBusinessInquiry, inquiryID, domain.AuditEventDelete, userID, beforeData, nil); err != nil {
		return domain.NewInternalError("failed to publish audit event", err)
	}

	return nil
}

// AdminAcknowledgeInquiry acknowledges a new business inquiry (admin only)
func (s *BusinessService) AdminAcknowledgeInquiry(ctx context.Context, inquiryID string, userID string) (*BusinessInquiry, error) {
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required to acknowledge business inquiries")
	}

	inquiry, err := s.repository.GetInquiry(ctx, inquiryID)
	if err != nil {
		return nil, err
	}

	if inquiry.Status != InquiryStatusNew {
		return nil, domain.NewValidationError("only new inquiries can be acknowledged")
	}

	beforeData := *inquiry

	inquiry.Status = InquiryStatusAcknowledged
	inquiry.UpdatedAt = time.Now().UTC()
	inquiry.UpdatedBy = userID

	if err := s.repository.SaveInquiry(ctx, inquiry); err != nil {
		return nil, domain.NewInternalError("failed to acknowledge business inquiry", err)
	}

	if err := s.repository.PublishAuditEvent(ctx, domain.EntityTypeBusinessInquiry, inquiry.InquiryID, domain.AuditEventUpdate, userID, beforeData, *inquiry); err != nil {
		return nil, domain.NewInternalError("failed to publish audit event", err)
	}

	return inquiry, nil
}

// AdminResolveInquiry resolves an in-progress business inquiry (admin only)
func (s *BusinessService) AdminResolveInquiry(ctx context.Context, inquiryID string, userID string) (*BusinessInquiry, error) {
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required to resolve business inquiries")
	}

	inquiry, err := s.repository.GetInquiry(ctx, inquiryID)
	if err != nil {
		return nil, err
	}

	if inquiry.Status != InquiryStatusInProgress {
		return nil, domain.NewValidationError("only in-progress inquiries can be resolved")
	}

	beforeData := *inquiry

	inquiry.Status = InquiryStatusResolved
	inquiry.UpdatedAt = time.Now().UTC()
	inquiry.UpdatedBy = userID

	if err := s.repository.SaveInquiry(ctx, inquiry); err != nil {
		return nil, domain.NewInternalError("failed to resolve business inquiry", err)
	}

	if err := s.repository.PublishAuditEvent(ctx, domain.EntityTypeBusinessInquiry, inquiry.InquiryID, domain.AuditEventUpdate, userID, beforeData, *inquiry); err != nil {
		return nil, domain.NewInternalError("failed to publish audit event", err)
	}

	return inquiry, nil
}

// AdminCloseInquiry closes a business inquiry (admin only)
func (s *BusinessService) AdminCloseInquiry(ctx context.Context, inquiryID string, userID string) (*BusinessInquiry, error) {
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required to close business inquiries")
	}

	inquiry, err := s.repository.GetInquiry(ctx, inquiryID)
	if err != nil {
		return nil, err
	}

	beforeData := *inquiry

	inquiry.Status = InquiryStatusClosed
	inquiry.UpdatedAt = time.Now().UTC()
	inquiry.UpdatedBy = userID

	if err := s.repository.SaveInquiry(ctx, inquiry); err != nil {
		return nil, domain.NewInternalError("failed to close business inquiry", err)
	}

	if err := s.repository.PublishAuditEvent(ctx, domain.EntityTypeBusinessInquiry, inquiry.InquiryID, domain.AuditEventUpdate, userID, beforeData, *inquiry); err != nil {
		return nil, domain.NewInternalError("failed to publish audit event", err)
	}

	return inquiry, nil
}

// AdminSetPriority sets the priority of a business inquiry (admin only)
func (s *BusinessService) AdminSetPriority(ctx context.Context, inquiryID string, priority InquiryPriority, userID string) (*BusinessInquiry, error) {
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required to set inquiry priority")
	}

	if !IsValidInquiryPriority(priority) {
		return nil, domain.NewValidationError("invalid priority level")
	}

	inquiry, err := s.repository.GetInquiry(ctx, inquiryID)
	if err != nil {
		return nil, err
	}

	beforeData := *inquiry

	inquiry.Priority = priority
	inquiry.UpdatedAt = time.Now().UTC()
	inquiry.UpdatedBy = userID

	if err := s.repository.SaveInquiry(ctx, inquiry); err != nil {
		return nil, domain.NewInternalError("failed to set inquiry priority", err)
	}

	if err := s.repository.PublishAuditEvent(ctx, domain.EntityTypeBusinessInquiry, inquiry.InquiryID, domain.AuditEventUpdate, userID, beforeData, *inquiry); err != nil {
		return nil, domain.NewInternalError("failed to publish audit event", err)
	}

	return inquiry, nil
}

// AdminListInquiries lists business inquiries with filters (admin only)
func (s *BusinessService) AdminListInquiries(ctx context.Context, filters InquiryFilters, userID string) ([]*BusinessInquiry, error) {
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required to list business inquiries")
	}

	inquiries, err := s.repository.ListInquiries(ctx, filters)
	if err != nil {
		return nil, domain.NewInternalError("failed to list business inquiries", err)
	}

	return inquiries, nil
}

// AdminGetInquiry retrieves a specific business inquiry (admin only)
func (s *BusinessService) AdminGetInquiry(ctx context.Context, inquiryID string, userID string) (*BusinessInquiry, error) {
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required to get business inquiry")
	}

	inquiry, err := s.repository.GetInquiry(ctx, inquiryID)
	if err != nil {
		return nil, err
	}

	return inquiry, nil
}