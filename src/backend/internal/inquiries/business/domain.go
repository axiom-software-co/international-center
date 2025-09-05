package business

import (
	"fmt"
	"net"
	"net/mail"
	"strings"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/google/uuid"
)

// InquiryStatus represents the status of a business inquiry
type InquiryStatus string

const (
	InquiryStatusNew          InquiryStatus = "new"
	InquiryStatusAcknowledged InquiryStatus = "acknowledged"
	InquiryStatusInProgress   InquiryStatus = "in_progress"
	InquiryStatusResolved     InquiryStatus = "resolved"
	InquiryStatusClosed       InquiryStatus = "closed"
)

// InquiryPriority represents the priority level of a business inquiry
type InquiryPriority string

const (
	InquiryPriorityLow    InquiryPriority = "low"
	InquiryPriorityMedium InquiryPriority = "medium"
	InquiryPriorityHigh   InquiryPriority = "high"
	InquiryPriorityUrgent InquiryPriority = "urgent"
)

// InquiryType represents the type of business inquiry
type InquiryType string

const (
	InquiryTypePartnership InquiryType = "partnership"
	InquiryTypeLicensing   InquiryType = "licensing"
	InquiryTypeResearch    InquiryType = "research"
	InquiryTypeTechnology  InquiryType = "technology"
	InquiryTypeRegulatory  InquiryType = "regulatory"
	InquiryTypeOther       InquiryType = "other"
)

// BusinessInquiry represents a business partnership and collaboration inquiry
type BusinessInquiry struct {
	// Primary identifiers
	InquiryID string          `json:"inquiry_id"`
	Status    InquiryStatus   `json:"status"`
	Priority  InquiryPriority `json:"priority"`

	// Organization Information
	OrganizationName string  `json:"organization_name"`
	ContactName      string  `json:"contact_name"`
	Title            string  `json:"title"`
	Email            string  `json:"email"`
	Phone            *string `json:"phone,omitempty"`
	Industry         *string `json:"industry,omitempty"`

	// Inquiry Details
	InquiryType InquiryType `json:"inquiry_type"`
	Message     string      `json:"message"`

	// Metadata
	Source    string  `json:"source"`
	IPAddress *net.IP `json:"ip_address,omitempty"`
	UserAgent *string `json:"user_agent,omitempty"`

	// Audit fields
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy string     `json:"created_by"`
	UpdatedBy string     `json:"updated_by"`
	IsDeleted bool       `json:"is_deleted"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// InquiryFilters represents filters for querying business inquiries
type InquiryFilters struct {
	Status      *InquiryStatus   `json:"status,omitempty"`
	Priority    *InquiryPriority `json:"priority,omitempty"`
	InquiryType *InquiryType     `json:"inquiry_type,omitempty"`
	Industry    *string          `json:"industry,omitempty"`
	CreatedFrom *time.Time       `json:"created_from,omitempty"`
	CreatedTo   *time.Time       `json:"created_to,omitempty"`
	Limit       *int             `json:"limit,omitempty"`
	Offset      *int             `json:"offset,omitempty"`
}

// Admin Request/Response types
type AdminCreateInquiryRequest struct {
	OrganizationName string  `json:"organization_name" validate:"required,min=2,max=100"`
	ContactName      string  `json:"contact_name" validate:"required,min=2,max=50"`
	Title            string  `json:"title" validate:"required,min=2,max=50"`
	Email            string  `json:"email" validate:"required,email,max=254"`
	Phone            *string `json:"phone,omitempty" validate:"omitempty,len=10"`
	Industry         *string `json:"industry,omitempty" validate:"omitempty,max=50"`
	InquiryType      string  `json:"inquiry_type" validate:"required,oneof=partnership licensing research technology regulatory other"`
	Message          string  `json:"message" validate:"required,min=20,max=1500"`
	Source           string  `json:"source,omitempty"`
	IPAddress        *string `json:"ip_address,omitempty"`
	UserAgent        *string `json:"user_agent,omitempty"`
}

type AdminUpdateInquiryRequest struct {
	OrganizationName *string `json:"organization_name,omitempty" validate:"omitempty,min=2,max=100"`
	ContactName      *string `json:"contact_name,omitempty" validate:"omitempty,min=2,max=50"`
	Title            *string `json:"title,omitempty" validate:"omitempty,min=2,max=50"`
	Email            *string `json:"email,omitempty" validate:"omitempty,email,max=254"`
	Phone            *string `json:"phone,omitempty" validate:"omitempty,len=10"`
	Industry         *string `json:"industry,omitempty" validate:"omitempty,max=50"`
	InquiryType      *string `json:"inquiry_type,omitempty" validate:"omitempty,oneof=partnership licensing research technology regulatory other"`
	Message          *string `json:"message,omitempty" validate:"omitempty,min=20,max=1500"`
}

// Domain validation functions
func (bi *BusinessInquiry) Validate() error {
	if bi.InquiryID == "" {
		return domain.NewValidationError("inquiry ID is required")
	}

	if bi.OrganizationName == "" {
		return domain.NewValidationError("organization name is required")
	}

	if len(bi.OrganizationName) < 2 || len(bi.OrganizationName) > 100 {
		return domain.NewValidationError("organization name must be between 2 and 100 characters")
	}

	if bi.ContactName == "" {
		return domain.NewValidationError("contact name is required")
	}

	if len(bi.ContactName) < 2 || len(bi.ContactName) > 50 {
		return domain.NewValidationError("contact name must be between 2 and 50 characters")
	}

	if bi.Email == "" {
		return domain.NewValidationError("email is required")
	}

	if len(bi.Email) > 254 {
		return domain.NewValidationError("email must not exceed 254 characters")
	}

	if err := validateBusinessEmail(bi.Email); err != nil {
		return err
	}

	if bi.Title == "" {
		return domain.NewValidationError("title is required")
	}

	if len(bi.Title) > 50 {
		return domain.NewValidationError("title must not exceed 50 characters")
	}

	if bi.Message == "" {
		return domain.NewValidationError("message is required")
	}

	if len(bi.Message) < 20 || len(bi.Message) > 1500 {
		return domain.NewValidationError("message must be between 20 and 1500 characters")
	}

	if !bi.InquiryType.IsValid() {
		return domain.NewValidationError("invalid inquiry type")
	}

	if !bi.Status.IsValid() {
		return domain.NewValidationError("invalid inquiry status")
	}

	if !bi.Priority.IsValid() {
		return domain.NewValidationError("invalid inquiry priority")
	}

	if bi.Phone != nil && len(*bi.Phone) != 10 {
		return domain.NewValidationError("phone number must be exactly 10 digits")
	}

	if bi.Industry != nil && len(*bi.Industry) > 50 {
		return domain.NewValidationError("industry must not exceed 50 characters")
	}

	return nil
}

// IsValidInquiryType checks if the inquiry type is valid
func IsValidInquiryType(inquiryType InquiryType) bool {
	switch inquiryType {
	case InquiryTypePartnership, InquiryTypeLicensing, InquiryTypeResearch, InquiryTypeTechnology, InquiryTypeRegulatory, InquiryTypeOther:
		return true
	default:
		return false
	}
}

// IsValidInquiryStatus checks if the inquiry status is valid
func IsValidInquiryStatus(status InquiryStatus) bool {
	switch status {
	case InquiryStatusNew, InquiryStatusAcknowledged, InquiryStatusInProgress, InquiryStatusResolved, InquiryStatusClosed:
		return true
	default:
		return false
	}
}

// IsValidInquiryPriority checks if the inquiry priority is valid
func IsValidInquiryPriority(priority InquiryPriority) bool {
	return priority.IsValid()
}

// IsValid checks if the inquiry priority is valid
func (p InquiryPriority) IsValid() bool {
	switch p {
	case InquiryPriorityLow, InquiryPriorityMedium, InquiryPriorityHigh, InquiryPriorityUrgent:
		return true
	default:
		return false
	}
}

// IsValid checks if the inquiry status is valid
func (s InquiryStatus) IsValid() bool {
	switch s {
	case InquiryStatusNew, InquiryStatusAcknowledged, InquiryStatusInProgress, InquiryStatusResolved, InquiryStatusClosed:
		return true
	default:
		return false
	}
}

// IsValid checks if the inquiry type is valid
func (t InquiryType) IsValid() bool {
	switch t {
	case InquiryTypePartnership, InquiryTypeLicensing, InquiryTypeResearch, InquiryTypeTechnology, InquiryTypeRegulatory, InquiryTypeOther:
		return true
	default:
		return false
	}
}

// Factory Functions

// NewBusinessInquiry creates a new business inquiry with validation
func NewBusinessInquiry(request AdminCreateInquiryRequest, userID string) (*BusinessInquiry, error) {
	// Validate required fields
	if request.OrganizationName == "" {
		return nil, domain.NewValidationError("organization name is required")
	}
	if len(request.OrganizationName) < 2 || len(request.OrganizationName) > 100 {
		return nil, domain.NewValidationError("organization name must be between 2 and 100 characters")
	}

	if request.ContactName == "" {
		return nil, domain.NewValidationError("contact name is required")
	}
	if len(request.ContactName) < 2 || len(request.ContactName) > 50 {
		return nil, domain.NewValidationError("contact name must be between 2 and 50 characters")
	}

	if request.Email == "" {
		return nil, domain.NewValidationError("email is required")
	}
	if _, err := mail.ParseAddress(request.Email); err != nil {
		return nil, domain.NewValidationError("invalid email format")
	}

	if request.Title == "" {
		return nil, domain.NewValidationError("title is required")
	}
	if len(request.Title) > 50 {
		return nil, domain.NewValidationError("title must not exceed 50 characters")
	}

	if request.Message == "" {
		return nil, domain.NewValidationError("message is required")
	}
	if len(request.Message) < 20 || len(request.Message) > 1500 {
		return nil, domain.NewValidationError("message must be between 20 and 1500 characters")
	}

	inquiryType := InquiryType(request.InquiryType)
	if !inquiryType.IsValid() {
		return nil, domain.NewValidationError("invalid inquiry type")
	}

	if request.Phone != nil && len(*request.Phone) != 10 {
		return nil, domain.NewValidationError("phone number must be exactly 10 digits")
	}

	if request.Industry != nil && len(*request.Industry) > 50 {
		return nil, domain.NewValidationError("industry must not exceed 50 characters")
	}

	// Parse IP address if provided
	var ipAddress *net.IP
	if request.IPAddress != nil {
		if ip := net.ParseIP(*request.IPAddress); ip != nil {
			ipAddress = &ip
		}
	}

	// Create the inquiry
	now := time.Now()
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
		Source:           getSourceOrDefault(request.Source),
		IPAddress:        ipAddress,
		UserAgent:        request.UserAgent,
		CreatedAt:        now,
		UpdatedAt:        now,
		CreatedBy:        userID,
		UpdatedBy:        userID,
		IsDeleted:        false,
	}

	return inquiry, nil
}

// Business Logic Methods

// SetPriority updates the priority of the inquiry
func (bi *BusinessInquiry) SetPriority(priority InquiryPriority, userID string) error {
	if !priority.IsValid() {
		return domain.NewValidationError("invalid priority value")
	}

	bi.Priority = priority
	bi.UpdatedBy = userID
	bi.UpdatedAt = time.Now()

	return nil
}

// Acknowledge transitions the inquiry from new to acknowledged status
func (bi *BusinessInquiry) Acknowledge(userID string) error {
	if bi.Status != InquiryStatusNew {
		return domain.NewValidationError(fmt.Sprintf("cannot acknowledge inquiry with status %s", bi.Status))
	}

	if bi.IsDeleted {
		return domain.NewValidationError("cannot acknowledge deleted inquiry")
	}

	bi.Status = InquiryStatusAcknowledged
	bi.UpdatedBy = userID
	bi.UpdatedAt = time.Now()

	return nil
}

// Resolve transitions the inquiry to resolved status
func (bi *BusinessInquiry) Resolve(userID string) error {
	if bi.Status != InquiryStatusAcknowledged && bi.Status != InquiryStatusInProgress {
		return domain.NewValidationError(fmt.Sprintf("cannot resolve inquiry with status %s", bi.Status))
	}

	if bi.IsDeleted {
		return domain.NewValidationError("cannot resolve deleted inquiry")
	}

	bi.Status = InquiryStatusResolved
	bi.UpdatedBy = userID
	bi.UpdatedAt = time.Now()

	return nil
}

// Close transitions the inquiry to closed status (can be done from any status)
func (bi *BusinessInquiry) Close(userID string) error {
	if bi.Status == InquiryStatusClosed {
		return domain.NewValidationError("inquiry is already closed")
	}

	if bi.IsDeleted {
		return domain.NewValidationError("cannot close deleted inquiry")
	}

	bi.Status = InquiryStatusClosed
	bi.UpdatedBy = userID
	bi.UpdatedAt = time.Now()

	return nil
}

// Helper functions

// getSourceOrDefault returns the source or defaults to "website"
func getSourceOrDefault(source string) string {
	if source == "" {
		return "website"
	}
	return source
}

// validateBusinessEmail validates email format with stricter business requirements
func validateBusinessEmail(email string) error {
	if email == "" {
		return domain.NewValidationError("email is required")
	}

	// Must contain @ symbol
	if !strings.Contains(email, "@") {
		return domain.NewValidationError("invalid email format")
	}

	// Use Go's built-in email parser
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return domain.NewValidationError("invalid email format")
	}

	// Additional check: must have domain part
	parts := strings.Split(addr.Address, "@")
	if len(parts) != 2 || len(parts[0]) == 0 || len(parts[1]) == 0 {
		return domain.NewValidationError("invalid email format")
	}

	// Domain must contain a dot
	if !strings.Contains(parts[1], ".") {
		return domain.NewValidationError("invalid email format")
	}

	return nil
}

// CanTransitionTo checks if the inquiry can transition to the target status
func (bi *BusinessInquiry) CanTransitionTo(targetStatus InquiryStatus) bool {
	switch bi.Status {
	case InquiryStatusNew:
		return targetStatus == InquiryStatusAcknowledged || targetStatus == InquiryStatusClosed
	case InquiryStatusAcknowledged:
		return targetStatus == InquiryStatusInProgress || targetStatus == InquiryStatusClosed
	case InquiryStatusInProgress:
		return targetStatus == InquiryStatusResolved || targetStatus == InquiryStatusClosed
	case InquiryStatusResolved:
		return targetStatus == InquiryStatusClosed || targetStatus == InquiryStatusInProgress
	case InquiryStatusClosed:
		return false // Cannot transition from closed
	default:
		return false
	}
}

// IsAdminUser checks if the user ID represents an admin user
func IsAdminUser(userID string) bool {
	return len(userID) > 6 && userID[:6] == "admin-"
}

