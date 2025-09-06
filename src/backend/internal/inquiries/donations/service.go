package donations

import (
	"context"
	"net"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
)

// DonationsRepositoryInterface defines the interface for donations inquiry data operations
type DonationsRepositoryInterface interface {
	SaveInquiry(ctx context.Context, inquiry *DonationsInquiry) error
	GetInquiry(ctx context.Context, inquiryID string) (*DonationsInquiry, error)
	DeleteInquiry(ctx context.Context, inquiryID string, userID string) error
	ListInquiries(ctx context.Context, filters InquiryFilters) ([]*DonationsInquiry, error)
	PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error
}

// DonationsService provides business logic for donations inquiry operations
type DonationsService struct {
	repository DonationsRepositoryInterface
}

// NewDonationsService creates a new donations service instance
func NewDonationsService(repository DonationsRepositoryInterface) *DonationsService {
	return &DonationsService{
		repository: repository,
	}
}

// IsAdminUser checks if the user ID represents an admin user
func IsAdminUser(userID string) bool {
	return len(userID) > 6 && userID[:6] == "admin-"
}

// AdminCreateInquiry creates a new donations inquiry (admin only)
func (s *DonationsService) AdminCreateInquiry(ctx context.Context, request AdminCreateInquiryRequest, userID string) (*DonationsInquiry, error) {
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required to create donations inquiries")
	}

	if err := request.Validate(); err != nil {
		return nil, err
	}

	donorType := DonorType(request.DonorType)
	inquiry, err := NewDonationsInquiry(request.ContactName, request.Email, request.Message, donorType, userID)
	if err != nil {
		return nil, err
	}

	// Set optional fields
	if request.Phone != nil {
		inquiry.Phone = request.Phone
	}

	if request.Organization != nil {
		inquiry.Organization = request.Organization
	}

	// Validate after setting all fields
	if err := inquiry.Validate(); err != nil {
		return nil, err
	}

	if request.InterestArea != nil {
		interestArea := InterestArea(*request.InterestArea)
		if !interestArea.IsValid() {
			return nil, domain.NewValidationError("invalid interest area")
		}
		inquiry.InterestArea = &interestArea
	}

	if request.PreferredAmountRange != nil {
		amountRange := AmountRange(*request.PreferredAmountRange)
		if !amountRange.IsValid() {
			return nil, domain.NewValidationError("invalid preferred amount range")
		}
		inquiry.PreferredAmountRange = &amountRange
	}

	if request.DonationFrequency != nil {
		donationFreq := DonationFrequency(*request.DonationFrequency)
		if !donationFreq.IsValid() {
			return nil, domain.NewValidationError("invalid donation frequency")
		}
		inquiry.DonationFrequency = &donationFreq
	}

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

	if err := s.repository.SaveInquiry(ctx, inquiry); err != nil {
		return nil, domain.NewInternalError("failed to save donations inquiry", err)
	}

	if err := s.repository.PublishAuditEvent(ctx, domain.EntityTypeDonationsInquiry, inquiry.InquiryID, domain.AuditEventInsert, userID, nil, inquiry); err != nil {
		return nil, domain.NewInternalError("failed to publish audit event", err)
	}

	return inquiry, nil
}

// AdminUpdateInquiry updates an existing donations inquiry (admin only)
func (s *DonationsService) AdminUpdateInquiry(ctx context.Context, inquiryID string, request AdminUpdateInquiryRequest, userID string) (*DonationsInquiry, error) {
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required to update donations inquiries")
	}

	inquiry, err := s.repository.GetInquiry(ctx, inquiryID)
	if err != nil {
		return nil, err
	}

	beforeData := *inquiry

	if request.ContactName != nil {
		if *request.ContactName == "" {
			return nil, domain.NewValidationError("contact name cannot be empty")
		}
		inquiry.ContactName = *request.ContactName
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

	if request.Organization != nil {
		inquiry.Organization = request.Organization
	}

	if request.DonorType != nil {
		donorType := DonorType(*request.DonorType)
		if !donorType.IsValid() {
			return nil, domain.NewValidationError("invalid donor type")
		}
		inquiry.DonorType = donorType
	}

	if request.InterestArea != nil {
		interestArea := InterestArea(*request.InterestArea)
		if !interestArea.IsValid() {
			return nil, domain.NewValidationError("invalid interest area")
		}
		inquiry.InterestArea = &interestArea
	}

	if request.PreferredAmountRange != nil {
		amountRange := AmountRange(*request.PreferredAmountRange)
		if !amountRange.IsValid() {
			return nil, domain.NewValidationError("invalid preferred amount range")
		}
		inquiry.PreferredAmountRange = &amountRange
	}

	if request.DonationFrequency != nil {
		donationFreq := DonationFrequency(*request.DonationFrequency)
		if !donationFreq.IsValid() {
			return nil, domain.NewValidationError("invalid donation frequency")
		}
		inquiry.DonationFrequency = &donationFreq
	}

	if request.Message != nil {
		if *request.Message == "" {
			return nil, domain.NewValidationError("message cannot be empty")
		}
		inquiry.Message = *request.Message
	}

	inquiry.UpdatedAt = time.Now()
	inquiry.UpdatedBy = userID

	if err := inquiry.Validate(); err != nil {
		return nil, err
	}

	if err := s.repository.SaveInquiry(ctx, inquiry); err != nil {
		return nil, domain.NewInternalError("failed to update donations inquiry", err)
	}

	if err := s.repository.PublishAuditEvent(ctx, domain.EntityTypeDonationsInquiry, inquiry.InquiryID, domain.AuditEventUpdate, userID, beforeData, *inquiry); err != nil {
		return nil, domain.NewInternalError("failed to publish audit event", err)
	}

	return inquiry, nil
}

// AdminDeleteInquiry soft deletes a donations inquiry (admin only)
func (s *DonationsService) AdminDeleteInquiry(ctx context.Context, inquiryID string, userID string) error {
	if !IsAdminUser(userID) {
		return domain.NewUnauthorizedError("admin privileges required to delete donations inquiries")
	}

	inquiry, err := s.repository.GetInquiry(ctx, inquiryID)
	if err != nil {
		return err
	}

	beforeData := *inquiry

	if err := s.repository.DeleteInquiry(ctx, inquiryID, userID); err != nil {
		return domain.NewInternalError("failed to delete donations inquiry", err)
	}

	if err := s.repository.PublishAuditEvent(ctx, domain.EntityTypeDonationsInquiry, inquiryID, domain.AuditEventDelete, userID, beforeData, nil); err != nil {
		return domain.NewInternalError("failed to publish audit event", err)
	}

	return nil
}

// AdminAcknowledgeInquiry acknowledges a new donations inquiry (admin only)
func (s *DonationsService) AdminAcknowledgeInquiry(ctx context.Context, inquiryID string, userID string) (*DonationsInquiry, error) {
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required to acknowledge donations inquiries")
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
		return nil, domain.NewInternalError("failed to acknowledge donations inquiry", err)
	}

	if err := s.repository.PublishAuditEvent(ctx, domain.EntityTypeDonationsInquiry, inquiry.InquiryID, domain.AuditEventUpdate, userID, beforeData, *inquiry); err != nil {
		return nil, domain.NewInternalError("failed to publish audit event", err)
	}

	return inquiry, nil
}

// AdminResolveInquiry resolves an in-progress donations inquiry (admin only)
func (s *DonationsService) AdminResolveInquiry(ctx context.Context, inquiryID string, userID string) (*DonationsInquiry, error) {
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required to resolve donations inquiries")
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
		return nil, domain.NewInternalError("failed to resolve donations inquiry", err)
	}

	if err := s.repository.PublishAuditEvent(ctx, domain.EntityTypeDonationsInquiry, inquiry.InquiryID, domain.AuditEventUpdate, userID, beforeData, *inquiry); err != nil {
		return nil, domain.NewInternalError("failed to publish audit event", err)
	}

	return inquiry, nil
}

// AdminCloseInquiry closes a donations inquiry (admin only)
func (s *DonationsService) AdminCloseInquiry(ctx context.Context, inquiryID string, userID string) (*DonationsInquiry, error) {
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required to close donations inquiries")
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
		return nil, domain.NewInternalError("failed to close donations inquiry", err)
	}

	if err := s.repository.PublishAuditEvent(ctx, domain.EntityTypeDonationsInquiry, inquiry.InquiryID, domain.AuditEventUpdate, userID, beforeData, *inquiry); err != nil {
		return nil, domain.NewInternalError("failed to publish audit event", err)
	}

	return inquiry, nil
}

// AdminSetPriority sets the priority of a donations inquiry (admin only)
func (s *DonationsService) AdminSetPriority(ctx context.Context, inquiryID string, priority InquiryPriority, userID string) (*DonationsInquiry, error) {
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

	if err := s.repository.PublishAuditEvent(ctx, domain.EntityTypeDonationsInquiry, inquiry.InquiryID, domain.AuditEventUpdate, userID, beforeData, *inquiry); err != nil {
		return nil, domain.NewInternalError("failed to publish audit event", err)
	}

	return inquiry, nil
}

// AdminListInquiries lists donations inquiries with filters (admin only)
func (s *DonationsService) AdminListInquiries(ctx context.Context, filters InquiryFilters, userID string) ([]*DonationsInquiry, error) {
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required to list donations inquiries")
	}

	inquiries, err := s.repository.ListInquiries(ctx, filters)
	if err != nil {
		return nil, domain.NewInternalError("failed to list donations inquiries", err)
	}

	return inquiries, nil
}

// AdminGetInquiry retrieves a specific donations inquiry (admin only)
func (s *DonationsService) AdminGetInquiry(ctx context.Context, inquiryID string, userID string) (*DonationsInquiry, error) {
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required to get donations inquiry")
	}

	inquiry, err := s.repository.GetInquiry(ctx, inquiryID)
	if err != nil {
		return nil, err
	}

	return inquiry, nil
}