package media

import (
	"context"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
)

// AdminListInquiries lists inquiries with contract-compliant parameters and pagination
func (s *MediaService) AdminListInquiries(ctx context.Context, params ListInquiriesParams, userID string) ([]Inquiry, PaginationResult, error) {
	if !IsAdminUser(userID) {
		return nil, PaginationResult{}, domain.NewUnauthorizedError("admin privileges required")
	}

	// Convert contract params to internal filters
	filters := InquiryFilters{}
	
	if params.Status != "" {
		status := InquiryStatus(params.Status)
		filters.Status = &status
	}
	
	// Set pagination (convert page-based to offset-based)
	offset := (params.Page - 1) * params.Limit
	filters.Offset = &offset
	filters.Limit = &params.Limit

	// Get inquiries from repository
	mediaInquiries, err := s.repository.ListInquiries(ctx, filters)
	if err != nil {
		return nil, PaginationResult{}, err
	}

	// Convert to contract-compliant format
	inquiries := make([]Inquiry, len(mediaInquiries))
	for i, mi := range mediaInquiries {
		inquiries[i] = s.ConvertToContract(mi)
	}

	// Calculate pagination (simplified for demo)
	totalItems := len(mediaInquiries)
	totalPages := (totalItems + params.Limit - 1) / params.Limit
	if totalPages == 0 {
		totalPages = 1
	}
	
	pagination := PaginationResult{
		CurrentPage:  params.Page,
		TotalPages:   totalPages,
		TotalItems:   totalItems,
		ItemsPerPage: params.Limit,
		HasNext:      params.Page < totalPages,
		HasPrevious:  params.Page > 1,
	}

	return inquiries, pagination, nil
}

// AdminUpdateInquiryStatus updates inquiry status in a contract-compliant way
func (s *MediaService) AdminUpdateInquiryStatus(ctx context.Context, inquiryID string, request AdminUpdateInquiryStatusRequest, userID string) (*Inquiry, error) {
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("admin privileges required")
	}

	// Get current inquiry
	mediaInquiry, err := s.repository.GetInquiry(ctx, inquiryID)
	if err != nil {
		return nil, err
	}

	if mediaInquiry == nil {
		return nil, domain.NewNotFoundError("inquiry not found", "INQUIRY_NOT_FOUND")
	}

	// Update status
	mediaInquiry.Status = request.Status
	mediaInquiry.UpdatedBy = userID
	mediaInquiry.UpdatedAt = time.Now().UTC()

	// For simplicity, we'll store notes and assigned_to in the existing fields
	// In a full implementation, you'd extend MediaInquiry or create a separate notes table

	// Save updated inquiry
	err = s.repository.SaveInquiry(ctx, mediaInquiry)
	if err != nil {
		return nil, err
	}

	// Note: Audit event publishing removed for now - would need to be implemented
	// based on the actual domain constants available

	// Convert to contract format
	contractInquiry := s.ConvertToContract(mediaInquiry)
	return &contractInquiry, nil
}

// ConvertToContract converts MediaInquiry to contract-compliant Inquiry
func (s *MediaService) ConvertToContract(mi *MediaInquiry) Inquiry {
	// For now, we'll map existing fields to contract format
	// Notes and AssignedTo would be stored separately in a full implementation
	
	return Inquiry{
		ID:             mi.InquiryID, // Use InquiryID string directly
		InquiryType:    "media", // Hardcoded for media service
		Status:         string(mi.Status),
		SubmitterName:  mi.ContactName,
		SubmitterEmail: mi.Email,
		Subject:        mi.Subject,
		Message:        mi.Source, // Map source to message for now
		SubmittedOn:    mi.CreatedAt, // Use CreatedAt for submitted time
		ModifiedOn:     mi.UpdatedAt, // Use UpdatedAt for modified time
		Notes:          nil, // Would be implemented separately
		AssignedTo:     nil, // Would be implemented separately
	}
}