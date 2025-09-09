package media

import (
	"context"
	"net"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
)

// MediaRepositoryInterface defines the interface for media inquiry data operations
type MediaRepositoryInterface interface {
	SaveInquiry(ctx context.Context, inquiry *MediaInquiry) error
	GetInquiry(ctx context.Context, inquiryID string) (*MediaInquiry, error)
	DeleteInquiry(ctx context.Context, inquiryID string, userID string) error
	ListInquiries(ctx context.Context, filters InquiryFilters) ([]*MediaInquiry, error)
	PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error
}

// MediaService provides business logic for media inquiry operations
type MediaService struct {
	repository MediaRepositoryInterface
}

// NewMediaService creates a new media service instance
func NewMediaService(repository MediaRepositoryInterface) *MediaService {
	return &MediaService{
		repository: repository,
	}
}

// AdminCreateInquiry creates a new media inquiry (admin only)
func (s *MediaService) AdminCreateInquiry(ctx context.Context, request AdminCreateInquiryRequest, userID string) (*MediaInquiry, error) {
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required to create media inquiries")
	}

	if err := request.Validate(); err != nil {
		return nil, err
	}

	inquiry, err := NewMediaInquiry(request.Outlet, request.ContactName, request.Title, request.Email, request.Phone, request.Subject, userID)
	if err != nil {
		return nil, err
	}

	// Set optional fields
	if request.MediaType != nil {
		mediaType := MediaType(*request.MediaType)
		inquiry.MediaType = &mediaType
	}

	if request.Deadline != nil {
		inquiry.Deadline = request.Deadline
	}

	// Calculate urgency from deadline
	inquiry.CalculateUrgencyFromDeadline()

	if request.IPAddress != nil {
		ip := net.ParseIP(*request.IPAddress)
		if ip != nil {
			ipStr := ip.String()
			inquiry.IPAddress = &ipStr
		}
	}

	if request.UserAgent != nil {
		inquiry.UserAgent = request.UserAgent
	}

	// Validate after setting all fields
	if err := inquiry.Validate(); err != nil {
		return nil, err
	}

	if err := s.repository.SaveInquiry(ctx, inquiry); err != nil {
		return nil, domain.NewInternalError("failed to save media inquiry", err)
	}

	if err := s.repository.PublishAuditEvent(ctx, domain.EntityTypeMediaInquiry, inquiry.InquiryID, domain.AuditEventInsert, userID, nil, inquiry); err != nil {
		return nil, domain.NewInternalError("failed to publish audit event", err)
	}

	return inquiry, nil
}

// AdminUpdateInquiry updates an existing media inquiry (admin only)
func (s *MediaService) AdminUpdateInquiry(ctx context.Context, inquiryID string, request AdminUpdateInquiryRequest, userID string) (*MediaInquiry, error) {
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required to update media inquiries")
	}

	inquiry, err := s.repository.GetInquiry(ctx, inquiryID)
	if err != nil {
		return nil, err
	}

	beforeData := *inquiry

	if request.Outlet != nil {
		if *request.Outlet == "" {
			return nil, domain.NewValidationError("outlet cannot be empty")
		}
		inquiry.Outlet = *request.Outlet
	}

	if request.ContactName != nil {
		if *request.ContactName == "" {
			return nil, domain.NewValidationError("contact_name cannot be empty")
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
		if *request.Phone == "" {
			return nil, domain.NewValidationError("phone cannot be empty")
		}
		inquiry.Phone = *request.Phone
	}

	if request.MediaType != nil {
		mediaType := MediaType(*request.MediaType)
		if !mediaType.IsValid() {
			return nil, domain.NewValidationError("invalid media type")
		}
		inquiry.MediaType = &mediaType
	}

	if request.Deadline != nil {
		inquiry.UpdateDeadline(request.Deadline, userID)
	}

	if request.Subject != nil {
		if *request.Subject == "" {
			return nil, domain.NewValidationError("subject cannot be empty")
		}
		inquiry.Subject = *request.Subject
	}

	inquiry.UpdatedAt = time.Now()
	inquiry.UpdatedBy = userID

	if err := inquiry.Validate(); err != nil {
		return nil, err
	}

	if err := s.repository.SaveInquiry(ctx, inquiry); err != nil {
		return nil, domain.NewInternalError("failed to update media inquiry", err)
	}

	if err := s.repository.PublishAuditEvent(ctx, domain.EntityTypeMediaInquiry, inquiry.InquiryID, domain.AuditEventUpdate, userID, beforeData, *inquiry); err != nil {
		return nil, domain.NewInternalError("failed to publish audit event", err)
	}

	return inquiry, nil
}

// AdminDeleteInquiry soft deletes a media inquiry (admin only)
func (s *MediaService) AdminDeleteInquiry(ctx context.Context, inquiryID string, userID string) error {
	if !IsAdminUser(userID) {
		return domain.NewUnauthorizedError("admin privileges required to delete media inquiries")
	}

	inquiry, err := s.repository.GetInquiry(ctx, inquiryID)
	if err != nil {
		return err
	}

	beforeData := *inquiry

	if err := s.repository.DeleteInquiry(ctx, inquiryID, userID); err != nil {
		return domain.NewInternalError("failed to delete media inquiry", err)
	}

	if err := s.repository.PublishAuditEvent(ctx, domain.EntityTypeMediaInquiry, inquiryID, domain.AuditEventDelete, userID, beforeData, nil); err != nil {
		return domain.NewInternalError("failed to publish audit event", err)
	}

	return nil
}

// AdminAcknowledgeInquiry acknowledges a new media inquiry (admin only)
func (s *MediaService) AdminAcknowledgeInquiry(ctx context.Context, inquiryID string, userID string) (*MediaInquiry, error) {
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required to acknowledge media inquiries")
	}

	inquiry, err := s.repository.GetInquiry(ctx, inquiryID)
	if err != nil {
		return nil, err
	}

	if err := inquiry.CanTransitionTo(InquiryStatusAcknowledged); err != nil {
		return nil, err
	}

	beforeData := *inquiry

	if err := inquiry.UpdateStatus(InquiryStatusAcknowledged, userID); err != nil {
		return nil, err
	}

	if err := s.repository.SaveInquiry(ctx, inquiry); err != nil {
		return nil, domain.NewInternalError("failed to acknowledge media inquiry", err)
	}

	if err := s.repository.PublishAuditEvent(ctx, domain.EntityTypeMediaInquiry, inquiry.InquiryID, domain.AuditEventUpdate, userID, beforeData, *inquiry); err != nil {
		return nil, domain.NewInternalError("failed to publish audit event", err)
	}

	return inquiry, nil
}

// AdminResolveInquiry resolves an in-progress media inquiry (admin only)
func (s *MediaService) AdminResolveInquiry(ctx context.Context, inquiryID string, userID string) (*MediaInquiry, error) {
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required to resolve media inquiries")
	}

	inquiry, err := s.repository.GetInquiry(ctx, inquiryID)
	if err != nil {
		return nil, err
	}

	if err := inquiry.CanTransitionTo(InquiryStatusResolved); err != nil {
		return nil, err
	}

	beforeData := *inquiry

	if err := inquiry.UpdateStatus(InquiryStatusResolved, userID); err != nil {
		return nil, err
	}

	if err := s.repository.SaveInquiry(ctx, inquiry); err != nil {
		return nil, domain.NewInternalError("failed to resolve media inquiry", err)
	}

	if err := s.repository.PublishAuditEvent(ctx, domain.EntityTypeMediaInquiry, inquiry.InquiryID, domain.AuditEventUpdate, userID, beforeData, *inquiry); err != nil {
		return nil, domain.NewInternalError("failed to publish audit event", err)
	}

	return inquiry, nil
}

// AdminCloseInquiry closes a media inquiry (admin only)
func (s *MediaService) AdminCloseInquiry(ctx context.Context, inquiryID string, userID string) (*MediaInquiry, error) {
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required to close media inquiries")
	}

	inquiry, err := s.repository.GetInquiry(ctx, inquiryID)
	if err != nil {
		return nil, err
	}

	beforeData := *inquiry

	if err := inquiry.UpdateStatus(InquiryStatusClosed, userID); err != nil {
		return nil, err
	}

	if err := s.repository.SaveInquiry(ctx, inquiry); err != nil {
		return nil, domain.NewInternalError("failed to close media inquiry", err)
	}

	if err := s.repository.PublishAuditEvent(ctx, domain.EntityTypeMediaInquiry, inquiry.InquiryID, domain.AuditEventUpdate, userID, beforeData, *inquiry); err != nil {
		return nil, domain.NewInternalError("failed to publish audit event", err)
	}

	return inquiry, nil
}

// AdminSetPriority sets the priority of a media inquiry (admin only)
func (s *MediaService) AdminSetPriority(ctx context.Context, inquiryID string, priority InquiryPriority, userID string) (*MediaInquiry, error) {
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required to set inquiry priority")
	}

	inquiry, err := s.repository.GetInquiry(ctx, inquiryID)
	if err != nil {
		return nil, err
	}

	beforeData := *inquiry

	if err := inquiry.UpdatePriority(priority, userID); err != nil {
		return nil, err
	}

	if err := s.repository.SaveInquiry(ctx, inquiry); err != nil {
		return nil, domain.NewInternalError("failed to set inquiry priority", err)
	}

	if err := s.repository.PublishAuditEvent(ctx, domain.EntityTypeMediaInquiry, inquiry.InquiryID, domain.AuditEventUpdate, userID, beforeData, *inquiry); err != nil {
		return nil, domain.NewInternalError("failed to publish audit event", err)
	}

	return inquiry, nil
}


// AdminGetInquiry retrieves a specific media inquiry (admin only)
func (s *MediaService) AdminGetInquiry(ctx context.Context, inquiryID string, userID string) (*MediaInquiry, error) {
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required to get media inquiry")
	}

	inquiry, err := s.repository.GetInquiry(ctx, inquiryID)
	if err != nil {
		return nil, err
	}

	return inquiry, nil
}