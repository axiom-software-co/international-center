package media

import (
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

// Inquiry Urgency Enum
type InquiryUrgency string

const (
	InquiryUrgencyLow    InquiryUrgency = "low"
	InquiryUrgencyMedium InquiryUrgency = "medium"
	InquiryUrgencyHigh   InquiryUrgency = "high"
)

func (u InquiryUrgency) IsValid() bool {
	switch u {
	case InquiryUrgencyLow, InquiryUrgencyMedium, InquiryUrgencyHigh:
		return true
	default:
		return false
	}
}

// Media Type Enum
type MediaType string

const (
	MediaTypePrint          MediaType = "print"
	MediaTypeDigital        MediaType = "digital"
	MediaTypeTelevision     MediaType = "television"
	MediaTypeRadio          MediaType = "radio"
	MediaTypePodcast        MediaType = "podcast"
	MediaTypeMedicalJournal MediaType = "medical-journal"
	MediaTypeOther          MediaType = "other"
)

func (m MediaType) IsValid() bool {
	switch m {
	case MediaTypePrint, MediaTypeDigital, MediaTypeTelevision, MediaTypeRadio, MediaTypePodcast, MediaTypeMedicalJournal, MediaTypeOther:
		return true
	default:
		return false
	}
}

// MediaInquiry represents a media inquiry for press relations and communications
type MediaInquiry struct {
	InquiryID   string          `json:"inquiry_id"`
	Status      InquiryStatus   `json:"status"`
	Priority    InquiryPriority `json:"priority"`
	Urgency     InquiryUrgency  `json:"urgency"`
	Outlet      string          `json:"outlet"`
	ContactName string          `json:"contact_name"`
	Title       string          `json:"title"`
	Email       string          `json:"email"`
	Phone       string          `json:"phone"`
	MediaType   *MediaType      `json:"media_type,omitempty"`
	Deadline    *time.Time      `json:"deadline,omitempty"`
	Subject     string          `json:"subject"`
	Source      string          `json:"source"`
	IPAddress   *string         `json:"ip_address,omitempty"`
	UserAgent   *string         `json:"user_agent,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	CreatedBy   string          `json:"created_by"`
	UpdatedBy   string          `json:"updated_by"`
	IsDeleted   bool            `json:"is_deleted"`
	DeletedAt   *time.Time      `json:"deleted_at,omitempty"`
}

// NewMediaInquiry creates a new media inquiry with default values
func NewMediaInquiry(outlet, contactName, title, email, phone, subject string, createdBy string) (*MediaInquiry, error) {
	now := time.Now()
	inquiryID := uuid.New().String()

	inquiry := &MediaInquiry{
		InquiryID:   inquiryID,
		Status:      InquiryStatusNew,
		Priority:    InquiryPriorityMedium,
		Urgency:     InquiryUrgencyMedium,
		Outlet:      outlet,
		ContactName: contactName,
		Title:       title,
		Email:       email,
		Phone:       phone,
		Subject:     subject,
		Source:      "website",
		CreatedAt:   now,
		UpdatedAt:   now,
		CreatedBy:   createdBy,
		UpdatedBy:   createdBy,
		IsDeleted:   false,
	}

	return inquiry, nil
}

// CalculateUrgencyFromDeadline calculates urgency based on deadline proximity
func (m *MediaInquiry) CalculateUrgencyFromDeadline() {
	if m.Deadline == nil {
		m.Urgency = InquiryUrgencyMedium
		return
	}

	timeUntilDeadline := m.Deadline.Sub(time.Now())
	
	if timeUntilDeadline <= 24*time.Hour {
		m.Urgency = InquiryUrgencyHigh
	} else if timeUntilDeadline <= 72*time.Hour { // 3 days
		m.Urgency = InquiryUrgencyMedium
	} else {
		m.Urgency = InquiryUrgencyLow
	}
}

// Validate validates the media inquiry data
func (m *MediaInquiry) Validate() error {
	if m.InquiryID == "" {
		return domain.NewValidationError("inquiry_id is required")
	}

	if !m.Status.IsValid() {
		return domain.NewValidationError("status must be one of: new, acknowledged, in_progress, resolved, closed")
	}

	if !m.Priority.IsValid() {
		return domain.NewValidationError("priority must be one of: low, medium, high, urgent")
	}

	if !m.Urgency.IsValid() {
		return domain.NewValidationError("urgency must be one of: low, medium, high")
	}

	if err := m.validateOutlet(); err != nil {
		return err
	}

	if err := m.validateContactName(); err != nil {
		return err
	}

	if err := m.validateTitle(); err != nil {
		return err
	}

	if err := m.validateEmail(); err != nil {
		return err
	}

	if err := m.validatePhone(); err != nil {
		return err
	}

	if err := m.validateMediaType(); err != nil {
		return err
	}

	if err := m.validateSubject(); err != nil {
		return err
	}

	if m.Source == "" {
		return domain.NewValidationError("source is required")
	}

	if m.CreatedBy == "" {
		return domain.NewValidationError("created_by is required")
	}

	if m.UpdatedBy == "" {
		return domain.NewValidationError("updated_by is required")
	}

	return nil
}

func (m *MediaInquiry) validateOutlet() error {
	if m.Outlet == "" {
		return domain.NewValidationError("outlet is required")
	}

	if len(m.Outlet) < 2 {
		return domain.NewValidationError("outlet must be at least 2 characters")
	}

	if len(m.Outlet) > 100 {
		return domain.NewValidationError("outlet cannot exceed 100 characters")
	}

	return nil
}

func (m *MediaInquiry) validateContactName() error {
	if m.ContactName == "" {
		return domain.NewValidationError("contact_name is required")
	}

	if len(m.ContactName) > 50 {
		return domain.NewValidationError("contact_name cannot exceed 50 characters")
	}

	nameRegex := regexp.MustCompile(`^[a-zA-Z\s\-']+$`)
	if !nameRegex.MatchString(m.ContactName) {
		return domain.NewValidationError("contact_name must contain only letters, spaces, hyphens, and apostrophes")
	}

	return nil
}

func (m *MediaInquiry) validateTitle() error {
	if m.Title == "" {
		return domain.NewValidationError("title is required")
	}

	if len(m.Title) > 50 {
		return domain.NewValidationError("title cannot exceed 50 characters")
	}

	return nil
}

func (m *MediaInquiry) validateEmail() error {
	if m.Email == "" {
		return domain.NewValidationError("email is required")
	}

	if len(m.Email) > 254 {
		return domain.NewValidationError("email cannot exceed 254 characters")
	}

	_, err := mail.ParseAddress(m.Email)
	if err != nil {
		return domain.NewValidationError("email must be a valid email address")
	}

	return nil
}

func (m *MediaInquiry) validatePhone() error {
	if m.Phone == "" {
		return domain.NewValidationError("phone is required")
	}

	if len(m.Phone) > 20 {
		return domain.NewValidationError("phone cannot exceed 20 characters")
	}

	phoneRegex := regexp.MustCompile(`^\d{10}$`)
	if !phoneRegex.MatchString(m.Phone) {
		return domain.NewValidationError("phone must be a valid 10-digit USA format")
	}

	return nil
}

func (m *MediaInquiry) validateMediaType() error {
	if m.MediaType != nil && !m.MediaType.IsValid() {
		return domain.NewValidationError("media_type must be one of: print, digital, television, radio, podcast, medical-journal, other")
	}
	return nil
}

func (m *MediaInquiry) validateSubject() error {
	if m.Subject == "" {
		return domain.NewValidationError("subject is required")
	}

	subjectLength := len(strings.TrimSpace(m.Subject))
	if subjectLength < 20 {
		return domain.NewValidationError("subject must be at least 20 characters")
	}

	if subjectLength > 1500 {
		return domain.NewValidationError("subject cannot exceed 1500 characters")
	}

	return nil
}

// CanTransitionTo checks if the inquiry can transition to the target status
func (m *MediaInquiry) CanTransitionTo(targetStatus InquiryStatus) error {
	switch targetStatus {
	case InquiryStatusAcknowledged:
		if m.Status != InquiryStatusNew {
			return domain.NewValidationError("inquiry can only be acknowledged from new status")
		}
	case InquiryStatusInProgress:
		if m.Status != InquiryStatusAcknowledged {
			return domain.NewValidationError("inquiry can only be set to in_progress from acknowledged status")
		}
	case InquiryStatusResolved:
		if m.Status != InquiryStatusInProgress {
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
func (m *MediaInquiry) UpdateStatus(newStatus InquiryStatus, userID string) error {
	if err := m.CanTransitionTo(newStatus); err != nil {
		return err
	}

	m.Status = newStatus
	m.UpdatedAt = time.Now()
	m.UpdatedBy = userID
	return nil
}

// UpdatePriority updates the inquiry priority
func (m *MediaInquiry) UpdatePriority(newPriority InquiryPriority, userID string) error {
	if !newPriority.IsValid() {
		return domain.NewValidationError("priority must be one of: low, medium, high, urgent")
	}

	m.Priority = newPriority
	m.UpdatedAt = time.Now()
	m.UpdatedBy = userID
	return nil
}

// UpdateDeadline updates the deadline and recalculates urgency
func (m *MediaInquiry) UpdateDeadline(newDeadline *time.Time, userID string) {
	m.Deadline = newDeadline
	m.CalculateUrgencyFromDeadline()
	m.UpdatedAt = time.Now()
	m.UpdatedBy = userID
}

// MarkAsDeleted performs soft delete on the inquiry
func (m *MediaInquiry) MarkAsDeleted(userID string) {
	now := time.Now()
	m.IsDeleted = true
	m.DeletedAt = &now
	m.UpdatedAt = now
	m.UpdatedBy = userID
}

// Request DTOs for Admin Operations

type AdminCreateInquiryRequest struct {
	Outlet      string     `json:"outlet"`
	ContactName string     `json:"contact_name"`
	Title       string     `json:"title"`
	Email       string     `json:"email"`
	Phone       string     `json:"phone"`
	MediaType   *string    `json:"media_type,omitempty"`
	Deadline    *time.Time `json:"deadline,omitempty"`
	Subject     string     `json:"subject"`
	IPAddress   *string    `json:"ip_address,omitempty"`
	UserAgent   *string    `json:"user_agent,omitempty"`
}

func (r *AdminCreateInquiryRequest) Validate() error {
	if r.Outlet == "" {
		return domain.NewValidationError("outlet is required")
	}

	if len(r.Outlet) < 2 {
		return domain.NewValidationError("outlet must be at least 2 characters")
	}

	if r.ContactName == "" {
		return domain.NewValidationError("contact_name is required")
	}

	if r.Title == "" {
		return domain.NewValidationError("title is required")
	}

	if r.Email == "" {
		return domain.NewValidationError("email is required")
	}

	if r.Phone == "" {
		return domain.NewValidationError("phone is required")
	}

	phoneRegex := regexp.MustCompile(`^\d{10}$`)
	if !phoneRegex.MatchString(r.Phone) {
		return domain.NewValidationError("phone must be a valid 10-digit USA format")
	}

	if r.MediaType != nil {
		mediaType := MediaType(*r.MediaType)
		if !mediaType.IsValid() {
			return domain.NewValidationError("media_type must be one of: print, digital, television, radio, podcast, medical-journal, other")
		}
	}

	if r.Subject == "" {
		return domain.NewValidationError("subject is required")
	}

	subjectLength := len(strings.TrimSpace(r.Subject))
	if subjectLength < 20 {
		return domain.NewValidationError("subject must be at least 20 characters")
	}

	if subjectLength > 1500 {
		return domain.NewValidationError("subject cannot exceed 1500 characters")
	}

	return nil
}

type AdminUpdateInquiryRequest struct {
	Outlet      *string    `json:"outlet,omitempty"`
	ContactName *string    `json:"contact_name,omitempty"`
	Title       *string    `json:"title,omitempty"`
	Email       *string    `json:"email,omitempty"`
	Phone       *string    `json:"phone,omitempty"`
	MediaType   *string    `json:"media_type,omitempty"`
	Deadline    *time.Time `json:"deadline,omitempty"`
	Subject     *string    `json:"subject,omitempty"`
}

// InquiryFilters represents filters for listing inquiries
type InquiryFilters struct {
	Status    *InquiryStatus   `json:"status,omitempty"`
	Priority  *InquiryPriority `json:"priority,omitempty"`
	Urgency   *InquiryUrgency  `json:"urgency,omitempty"`
	MediaType *MediaType       `json:"media_type,omitempty"`
	Outlet    *string          `json:"outlet,omitempty"`
	Limit     *int             `json:"limit,omitempty"`
	Offset    *int             `json:"offset,omitempty"`
}

// IsAdminUser checks if the user ID represents an admin user
func IsAdminUser(userID string) bool {
	return len(userID) > 6 && userID[:6] == "admin-"
}

