package donations

import (
	"context"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/google/uuid"
)

// Inquiry Status Enum
type InquiryStatus string

const (
	InquiryStatusNew          InquiryStatus = "new"
	InquiryStatusAcknowledged InquiryStatus = "acknowledged"
	InquiryStatusInProgress   InquiryStatus = "in_progress"
	InquiryStatusResolved     InquiryStatus = "resolved"
	InquiryStatusClosed       InquiryStatus = "closed"
)

func (s InquiryStatus) IsValid() bool {
	switch s {
	case InquiryStatusNew, InquiryStatusAcknowledged, InquiryStatusInProgress, InquiryStatusResolved, InquiryStatusClosed:
		return true
	default:
		return false
	}
}

// Inquiry Priority Enum
type InquiryPriority string

const (
	InquiryPriorityLow    InquiryPriority = "low"
	InquiryPriorityMedium InquiryPriority = "medium"
	InquiryPriorityHigh   InquiryPriority = "high"
	InquiryPriorityUrgent InquiryPriority = "urgent"
)

func (p InquiryPriority) IsValid() bool {
	switch p {
	case InquiryPriorityLow, InquiryPriorityMedium, InquiryPriorityHigh, InquiryPriorityUrgent:
		return true
	default:
		return false
	}
}

// Donor Type Enum
type DonorType string

const (
	DonorTypeIndividual  DonorType = "individual"
	DonorTypeCorporate   DonorType = "corporate"
	DonorTypeFoundation  DonorType = "foundation"
	DonorTypeEstate      DonorType = "estate"
	DonorTypeOther       DonorType = "other"
)

func (d DonorType) IsValid() bool {
	switch d {
	case DonorTypeIndividual, DonorTypeCorporate, DonorTypeFoundation, DonorTypeEstate, DonorTypeOther:
		return true
	default:
		return false
	}
}

func (d DonorType) RequiresOrganization() bool {
	switch d {
	case DonorTypeCorporate, DonorTypeFoundation:
		return true
	default:
		return false
	}
}

// Interest Area Enum
type InterestArea string

const (
	InterestAreaClinicDevelopment InterestArea = "clinic-development"
	InterestAreaResearchFunding   InterestArea = "research-funding"
	InterestAreaPatientCare       InterestArea = "patient-care"
	InterestAreaEquipment         InterestArea = "equipment"
	InterestAreaGeneralSupport    InterestArea = "general-support"
	InterestAreaOther             InterestArea = "other"
)

func (i InterestArea) IsValid() bool {
	switch i {
	case InterestAreaClinicDevelopment, InterestAreaResearchFunding, InterestAreaPatientCare, InterestAreaEquipment, InterestAreaGeneralSupport, InterestAreaOther:
		return true
	default:
		return false
	}
}

// Amount Range Enum
type AmountRange string

const (
	AmountRangeUnder1000     AmountRange = "under-1000"
	AmountRange1000To5000    AmountRange = "1000-5000"
	AmountRange5000To25000   AmountRange = "5000-25000"
	AmountRange25000To100000 AmountRange = "25000-100000"
	AmountRangeOver100000    AmountRange = "over-100000"
	AmountRangeUndisclosed   AmountRange = "undisclosed"
)

func (a AmountRange) IsValid() bool {
	switch a {
	case AmountRangeUnder1000, AmountRange1000To5000, AmountRange5000To25000, AmountRange25000To100000, AmountRangeOver100000, AmountRangeUndisclosed:
		return true
	default:
		return false
	}
}

// Donation Frequency Enum
type DonationFrequency string

const (
	DonationFrequencyOneTime   DonationFrequency = "one-time"
	DonationFrequencyMonthly   DonationFrequency = "monthly"
	DonationFrequencyQuarterly DonationFrequency = "quarterly"
	DonationFrequencyAnnually  DonationFrequency = "annually"
	DonationFrequencyOther     DonationFrequency = "other"
)

func (d DonationFrequency) IsValid() bool {
	switch d {
	case DonationFrequencyOneTime, DonationFrequencyMonthly, DonationFrequencyQuarterly, DonationFrequencyAnnually, DonationFrequencyOther:
		return true
	default:
		return false
	}
}

// DonationsInquiry represents a donor contact and fundraising inquiry
type DonationsInquiry struct {
	InquiryID             string             `json:"inquiry_id"`
	Status                InquiryStatus      `json:"status"`
	Priority              InquiryPriority    `json:"priority"`
	ContactName           string             `json:"contact_name"`
	Email                 string             `json:"email"`
	Phone                 *string            `json:"phone,omitempty"`
	Organization          *string            `json:"organization,omitempty"`
	DonorType             DonorType          `json:"donor_type"`
	InterestArea          *InterestArea      `json:"interest_area,omitempty"`
	PreferredAmountRange  *AmountRange       `json:"preferred_amount_range,omitempty"`
	DonationFrequency     *DonationFrequency `json:"donation_frequency,omitempty"`
	Message               string             `json:"message"`
	Source                string             `json:"source"`
	IPAddress             *string            `json:"ip_address,omitempty"`
	UserAgent             *string            `json:"user_agent,omitempty"`
	CreatedAt             time.Time          `json:"created_at"`
	UpdatedAt             time.Time          `json:"updated_at"`
	CreatedBy             string             `json:"created_by"`
	UpdatedBy             string             `json:"updated_by"`
	IsDeleted             bool               `json:"is_deleted"`
	DeletedAt             *time.Time         `json:"deleted_at,omitempty"`
}

// NewDonationsInquiry creates a new donations inquiry with default values
func NewDonationsInquiry(contactName, email, message string, donorType DonorType, createdBy string) (*DonationsInquiry, error) {
	now := time.Now()
	inquiryID := uuid.New().String()

	inquiry := &DonationsInquiry{
		InquiryID:   inquiryID,
		Status:      InquiryStatusNew,
		Priority:    InquiryPriorityMedium,
		ContactName: contactName,
		Email:       email,
		DonorType:   donorType,
		Message:     message,
		Source:      "website",
		CreatedAt:   now,
		UpdatedAt:   now,
		CreatedBy:   createdBy,
		UpdatedBy:   createdBy,
		IsDeleted:   false,
	}

	return inquiry, nil
}

// Validate validates the donations inquiry data
func (d *DonationsInquiry) Validate() error {
	if d.InquiryID == "" {
		return domain.NewValidationError("inquiry_id is required")
	}

	if !d.Status.IsValid() {
		return domain.NewValidationError("status must be one of: new, acknowledged, in_progress, resolved, closed")
	}

	if !d.Priority.IsValid() {
		return domain.NewValidationError("priority must be one of: low, medium, high, urgent")
	}

	if err := d.validateContactName(); err != nil {
		return err
	}

	if err := d.validateEmail(); err != nil {
		return err
	}

	if err := d.validatePhone(); err != nil {
		return err
	}

	if !d.DonorType.IsValid() {
		return domain.NewValidationError("donor_type must be one of: individual, corporate, foundation, estate, other")
	}

	if err := d.validateOrganization(); err != nil {
		return err
	}

	if err := d.validateInterestArea(); err != nil {
		return err
	}

	if err := d.validateAmountRange(); err != nil {
		return err
	}

	if err := d.validateDonationFrequency(); err != nil {
		return err
	}

	if err := d.validateMessage(); err != nil {
		return err
	}

	if d.Source == "" {
		return domain.NewValidationError("source is required")
	}

	if d.CreatedBy == "" {
		return domain.NewValidationError("created_by is required")
	}

	if d.UpdatedBy == "" {
		return domain.NewValidationError("updated_by is required")
	}

	return nil
}

func (d *DonationsInquiry) validateContactName() error {
	if d.ContactName == "" {
		return domain.NewValidationError("contact_name is required")
	}

	if len(d.ContactName) < 2 {
		return domain.NewValidationError("contact_name must be at least 2 characters")
	}

	if len(d.ContactName) > 50 {
		return domain.NewValidationError("contact_name cannot exceed 50 characters")
	}

	nameRegex := regexp.MustCompile(`^[a-zA-Z\s\-']+$`)
	if !nameRegex.MatchString(d.ContactName) {
		return domain.NewValidationError("contact_name must contain only letters, spaces, hyphens, and apostrophes")
	}

	return nil
}

func (d *DonationsInquiry) validateEmail() error {
	if d.Email == "" {
		return domain.NewValidationError("email is required")
	}

	if len(d.Email) > 254 {
		return domain.NewValidationError("email cannot exceed 254 characters")
	}

	_, err := mail.ParseAddress(d.Email)
	if err != nil {
		return domain.NewValidationError("email must be a valid email address")
	}

	return nil
}

func (d *DonationsInquiry) validatePhone() error {
	if d.Phone == nil {
		return nil
	}

	if len(*d.Phone) > 20 {
		return domain.NewValidationError("phone cannot exceed 20 characters")
	}

	phoneRegex := regexp.MustCompile(`^\d{10}$`)
	if !phoneRegex.MatchString(*d.Phone) {
		return domain.NewValidationError("phone must be a valid 10-digit USA format")
	}

	return nil
}

func (d *DonationsInquiry) validateOrganization() error {
	if d.DonorType.RequiresOrganization() {
		if d.Organization == nil || *d.Organization == "" {
			return domain.NewValidationError(fmt.Sprintf("organization is required for %s donors", d.DonorType))
		}
	}

	if d.Organization != nil {
		if len(*d.Organization) > 100 {
			return domain.NewValidationError("organization cannot exceed 100 characters")
		}
		
		if len(*d.Organization) < 2 {
			return domain.NewValidationError("organization must be at least 2 characters")
		}
	}

	return nil
}

func (d *DonationsInquiry) validateInterestArea() error {
	if d.InterestArea != nil && !d.InterestArea.IsValid() {
		return domain.NewValidationError("interest_area must be one of: clinic-development, research-funding, patient-care, equipment, general-support, other")
	}
	return nil
}

func (d *DonationsInquiry) validateAmountRange() error {
	if d.PreferredAmountRange != nil && !d.PreferredAmountRange.IsValid() {
		return domain.NewValidationError("preferred_amount_range must be one of: under-1000, 1000-5000, 5000-25000, 25000-100000, over-100000, undisclosed")
	}
	return nil
}

func (d *DonationsInquiry) validateDonationFrequency() error {
	if d.DonationFrequency != nil && !d.DonationFrequency.IsValid() {
		return domain.NewValidationError("donation_frequency must be one of: one-time, monthly, quarterly, annually, other")
	}
	return nil
}

func (d *DonationsInquiry) validateMessage() error {
	if d.Message == "" {
		return domain.NewValidationError("message is required")
	}

	messageLength := len(strings.TrimSpace(d.Message))
	if messageLength < 20 {
		return domain.NewValidationError("message must be at least 20 characters")
	}

	if messageLength > 2000 {
		return domain.NewValidationError("message cannot exceed 2000 characters")
	}

	return nil
}

// CanTransitionTo checks if the inquiry can transition to the target status
func (d *DonationsInquiry) CanTransitionTo(targetStatus InquiryStatus) error {
	switch targetStatus {
	case InquiryStatusAcknowledged:
		if d.Status != InquiryStatusNew {
			return domain.NewValidationError("inquiry can only be acknowledged from new status")
		}
	case InquiryStatusInProgress:
		if d.Status != InquiryStatusAcknowledged {
			return domain.NewValidationError("inquiry can only be set to in_progress from acknowledged status")
		}
	case InquiryStatusResolved:
		if d.Status != InquiryStatusInProgress {
			return domain.NewValidationError("inquiry can only be resolved from in_progress status")
		}
	case InquiryStatusClosed:
		// Closed can be set from any status (emergency closure)
		return nil
	default:
		return domain.NewValidationError(fmt.Sprintf("invalid target status: %s", targetStatus))
	}
	return nil
}

// UpdateStatus updates the inquiry status with validation
func (d *DonationsInquiry) UpdateStatus(newStatus InquiryStatus, userID string) error {
	if err := d.CanTransitionTo(newStatus); err != nil {
		return err
	}

	d.Status = newStatus
	d.UpdatedAt = time.Now()
	d.UpdatedBy = userID
	return nil
}

// UpdatePriority updates the inquiry priority
func (d *DonationsInquiry) UpdatePriority(newPriority InquiryPriority, userID string) error {
	if !newPriority.IsValid() {
		return domain.NewValidationError("priority must be one of: low, medium, high, urgent")
	}

	d.Priority = newPriority
	d.UpdatedAt = time.Now()
	d.UpdatedBy = userID
	return nil
}

// MarkAsDeleted performs soft delete on the inquiry
func (d *DonationsInquiry) MarkAsDeleted(userID string) {
	now := time.Now()
	d.IsDeleted = true
	d.DeletedAt = &now
	d.UpdatedAt = now
	d.UpdatedBy = userID
}

// Request DTOs for Admin Operations

type AdminCreateInquiryRequest struct {
	ContactName          string  `json:"contact_name"`
	Email                string  `json:"email"`
	Phone                *string `json:"phone,omitempty"`
	Organization         *string `json:"organization,omitempty"`
	DonorType            string  `json:"donor_type"`
	InterestArea         *string `json:"interest_area,omitempty"`
	PreferredAmountRange *string `json:"preferred_amount_range,omitempty"`
	DonationFrequency    *string `json:"donation_frequency,omitempty"`
	Message              string  `json:"message"`
	IPAddress            *string `json:"ip_address,omitempty"`
	UserAgent            *string `json:"user_agent,omitempty"`
}

func (r *AdminCreateInquiryRequest) Validate() error {
	if r.ContactName == "" {
		return domain.NewValidationError("contact_name is required")
	}

	if r.Email == "" {
		return domain.NewValidationError("email is required")
	}

	if r.DonorType == "" {
		return domain.NewValidationError("donor_type is required")
	}

	donorType := DonorType(r.DonorType)
	if !donorType.IsValid() {
		return domain.NewValidationError("donor_type must be one of: individual, corporate, foundation, estate, other")
	}

	if donorType.RequiresOrganization() && (r.Organization == nil || *r.Organization == "") {
		return domain.NewValidationError(fmt.Sprintf("organization is required for %s donors", donorType))
	}

	if r.Message == "" {
		return domain.NewValidationError("message is required")
	}

	messageLength := len(strings.TrimSpace(r.Message))
	if messageLength < 20 {
		return domain.NewValidationError("message must be at least 20 characters")
	}

	if messageLength > 2000 {
		return domain.NewValidationError("message cannot exceed 2000 characters")
	}

	return nil
}

type AdminUpdateInquiryRequest struct {
	ContactName          *string `json:"contact_name,omitempty"`
	Email                *string `json:"email,omitempty"`
	Phone                *string `json:"phone,omitempty"`
	Organization         *string `json:"organization,omitempty"`
	DonorType            *string `json:"donor_type,omitempty"`
	InterestArea         *string `json:"interest_area,omitempty"`
	PreferredAmountRange *string `json:"preferred_amount_range,omitempty"`
	DonationFrequency    *string `json:"donation_frequency,omitempty"`
	Message              *string `json:"message,omitempty"`
}

// InquiryFilters represents filters for listing inquiries
type InquiryFilters struct {
	Status       *InquiryStatus    `json:"status,omitempty"`
	Priority     *InquiryPriority  `json:"priority,omitempty"`
	DonorType    *DonorType        `json:"donor_type,omitempty"`
	InterestArea *InterestArea     `json:"interest_area,omitempty"`
	AmountRange  *AmountRange      `json:"amount_range,omitempty"`
	Limit        *int              `json:"limit,omitempty"`
	Offset       *int              `json:"offset,omitempty"`
}

// Repository Interface

type DonationsRepositoryInterface interface {
	SaveInquiry(ctx context.Context, inquiry *DonationsInquiry) error
	GetInquiry(ctx context.Context, inquiryID string) (*DonationsInquiry, error)
	DeleteInquiry(ctx context.Context, inquiryID string, userID string) error
	ListInquiries(ctx context.Context, filters InquiryFilters) ([]*DonationsInquiry, error)
	PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error
}