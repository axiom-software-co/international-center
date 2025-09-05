package volunteers

import (
	"fmt"
	"net"
	"net/mail"
	"strings"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/google/uuid"
)

// Domain types matching TABLES-VOLUNTEERS.md schema
type ApplicationStatus string
type ApplicationPriority string
type VolunteerInterest string
type Availability string

const (
	ApplicationStatusNew                ApplicationStatus = "new"
	ApplicationStatusUnderReview        ApplicationStatus = "under-review"
	ApplicationStatusInterviewScheduled ApplicationStatus = "interview-scheduled"
	ApplicationStatusBackgroundCheck    ApplicationStatus = "background-check"
	ApplicationStatusApproved           ApplicationStatus = "approved"
	ApplicationStatusDeclined           ApplicationStatus = "declined"
	ApplicationStatusWithdrawn          ApplicationStatus = "withdrawn"
)

const (
	ApplicationPriorityLow    ApplicationPriority = "low"
	ApplicationPriorityMedium ApplicationPriority = "medium"
	ApplicationPriorityHigh   ApplicationPriority = "high"
	ApplicationPriorityUrgent ApplicationPriority = "urgent"
)

const (
	VolunteerInterestPatientSupport       VolunteerInterest = "patient-support"
	VolunteerInterestCommunityOutreach    VolunteerInterest = "community-outreach"
	VolunteerInterestResearchSupport      VolunteerInterest = "research-support"
	VolunteerInterestAdministrativeSupport VolunteerInterest = "administrative-support"
	VolunteerInterestMultiple             VolunteerInterest = "multiple"
	VolunteerInterestOther                VolunteerInterest = "other"
)

const (
	Availability2To4Hours   Availability = "2-4-hours"
	Availability4To8Hours   Availability = "4-8-hours"
	Availability8To16Hours  Availability = "8-16-hours"
	Availability16HoursPlus Availability = "16-hours-plus"
	AvailabilityFlexible    Availability = "flexible"
)

// IsValid checks if the application status is valid
func (s ApplicationStatus) IsValid() bool {
	switch s {
	case ApplicationStatusNew, ApplicationStatusUnderReview, ApplicationStatusInterviewScheduled, ApplicationStatusBackgroundCheck, ApplicationStatusApproved, ApplicationStatusDeclined, ApplicationStatusWithdrawn:
		return true
	default:
		return false
	}
}

// IsValid checks if the application priority is valid
func (p ApplicationPriority) IsValid() bool {
	switch p {
	case ApplicationPriorityLow, ApplicationPriorityMedium, ApplicationPriorityHigh, ApplicationPriorityUrgent:
		return true
	default:
		return false
	}
}

// IsValid checks if the volunteer interest is valid
func (v VolunteerInterest) IsValid() bool {
	switch v {
	case VolunteerInterestPatientSupport, VolunteerInterestCommunityOutreach, VolunteerInterestResearchSupport, VolunteerInterestAdministrativeSupport, VolunteerInterestMultiple, VolunteerInterestOther:
		return true
	default:
		return false
	}
}

// IsValid checks if the availability is valid
func (a Availability) IsValid() bool {
	switch a {
	case Availability2To4Hours, Availability4To8Hours, Availability8To16Hours, Availability16HoursPlus, AvailabilityFlexible:
		return true
	default:
		return false
	}
}

// VolunteerApplication represents the main volunteer application entity matching TABLES-VOLUNTEERS.md
type VolunteerApplication struct {
	ApplicationID string              `json:"application_id"`
	Status        ApplicationStatus   `json:"status"`
	Priority      ApplicationPriority `json:"priority"`
	
	// Personal Information
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Age       int    `json:"age"`
	
	// Volunteer Details
	VolunteerInterest     VolunteerInterest `json:"volunteer_interest"`
	Availability          Availability      `json:"availability"`
	Experience            string            `json:"experience,omitempty"`
	Motivation            string            `json:"motivation"`
	SchedulePreferences   string            `json:"schedule_preferences,omitempty"`
	
	// Metadata
	Source    string `json:"source,omitempty"`
	IPAddress net.IP `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
	
	// Audit fields
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy string     `json:"created_by"`
	UpdatedBy string     `json:"updated_by"`
	IsDeleted bool       `json:"is_deleted"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// Validation helper functions

func validateName(name, fieldName string) error {
	if strings.TrimSpace(name) == "" {
		return domain.NewValidationError(fieldName + " is required")
	}
	if len(name) < 2 || len(name) > 50 {
		return domain.NewValidationError(fieldName + " must be between 2 and 50 characters")
	}
	return nil
}

func validateEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return domain.NewValidationError("email is required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return domain.NewValidationError("invalid email format")
	}
	return nil
}

func validateAge(age int) error {
	if age < 18 {
		return domain.NewValidationError("applicant must be at least 18 years old")
	}
	if age > 100 {
		return domain.NewValidationError("invalid age")
	}
	return nil
}

func validateMotivation(motivation string) error {
	if len(strings.TrimSpace(motivation)) < 30 {
		return domain.NewValidationError("motivation must be at least 30 characters")
	}
	if len(motivation) > 2000 {
		return domain.NewValidationError("motivation cannot exceed 2000 characters")
	}
	return nil
}

func validatePhone(phone string, required bool) error {
	if required && phone == "" {
		return domain.NewValidationError("phone number is required")
	}
	if phone != "" && len(phone) != 10 {
		return domain.NewValidationError("phone number must be exactly 10 digits")
	}
	return nil
}

func validateUserID(userID, fieldName string) error {
	if strings.TrimSpace(userID) == "" {
		return domain.NewValidationError(fieldName + " is required")
	}
	return nil
}

// NewVolunteerApplication creates a new volunteer application with validation
func NewVolunteerApplication(firstName, lastName, email, phone string, age int, interest VolunteerInterest, availability Availability, motivation, userID string) (*VolunteerApplication, error) {
	// Validate all input parameters using helper functions
	if err := validateName(firstName, "first name"); err != nil {
		return nil, err
	}
	if err := validateName(lastName, "last name"); err != nil {
		return nil, err
	}
	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if err := validatePhone(phone, true); err != nil {
		return nil, err
	}
	if err := validateAge(age); err != nil {
		return nil, err
	}
	if err := validateMotivation(motivation); err != nil {
		return nil, err
	}

	if !interest.IsValid() {
		return nil, domain.NewValidationError("invalid volunteer interest")
	}
	if !availability.IsValid() {
		return nil, domain.NewValidationError("invalid availability")
	}

	// Create the application
	now := time.Now()
	application := &VolunteerApplication{
		ApplicationID:     uuid.New().String(),
		Status:            ApplicationStatusNew,
		Priority:          ApplicationPriorityMedium,
		FirstName:         firstName,
		LastName:          lastName,
		Email:             email,
		Phone:             phone,
		Age:               age,
		VolunteerInterest: interest,
		Availability:      availability,
		Motivation:        motivation,
		Source:            "website",
		CreatedAt:         now,
		UpdatedAt:         now,
		CreatedBy:         userID,
		UpdatedBy:         userID,
		IsDeleted:         false,
	}

	return application, nil
}

// Validate validates the volunteer application data
func (v *VolunteerApplication) Validate() error {
	if v.ApplicationID == "" {
		return domain.NewValidationError("application ID is required")
	}

	// Use helper functions for common validation logic
	if err := validateName(v.FirstName, "first name"); err != nil {
		return err
	}
	if err := validateName(v.LastName, "last name"); err != nil {
		return err
	}
	if err := validateEmail(v.Email); err != nil {
		return err
	}
	if err := validateAge(v.Age); err != nil {
		return err
	}
	if err := validateMotivation(v.Motivation); err != nil {
		return err
	}
	if err := validatePhone(v.Phone, false); err != nil {
		return err
	}
	if err := validateUserID(v.CreatedBy, "created by"); err != nil {
		return err
	}
	if err := validateUserID(v.UpdatedBy, "updated by"); err != nil {
		return err
	}

	// Validate enum fields
	if !v.Status.IsValid() {
		return domain.NewValidationError("invalid application status")
	}
	if !v.Priority.IsValid() {
		return domain.NewValidationError("invalid application priority")
	}
	if !v.VolunteerInterest.IsValid() {
		return domain.NewValidationError("invalid volunteer interest")
	}
	if !v.Availability.IsValid() {
		return domain.NewValidationError("invalid availability")
	}

	return nil
}

// SetPriority updates the priority of the application
func (v *VolunteerApplication) SetPriority(priority ApplicationPriority, userID string) error {
	if !priority.IsValid() {
		return domain.NewValidationError("invalid priority value")
	}

	v.Priority = priority
	v.UpdatedBy = userID
	v.UpdatedAt = time.Now()

	return nil
}

// UpdateStatus updates the status of the application with validation
func (v *VolunteerApplication) UpdateStatus(newStatus ApplicationStatus, userID string) error {
	if !newStatus.IsValid() {
		return domain.NewValidationError("invalid status value")
	}

	// Check if transition is allowed
	if err := v.CanTransitionTo(newStatus); err != nil {
		return err
	}

	v.Status = newStatus
	v.UpdatedBy = userID
	v.UpdatedAt = time.Now()

	return nil
}

// CanTransitionTo checks if the application can transition to the target status
func (v *VolunteerApplication) CanTransitionTo(targetStatus ApplicationStatus) error {
	switch v.Status {
	case ApplicationStatusNew:
		if targetStatus == ApplicationStatusUnderReview || targetStatus == ApplicationStatusDeclined || targetStatus == ApplicationStatusWithdrawn {
			return nil
		}
	case ApplicationStatusUnderReview:
		if targetStatus == ApplicationStatusInterviewScheduled || targetStatus == ApplicationStatusDeclined || targetStatus == ApplicationStatusWithdrawn {
			return nil
		}
	case ApplicationStatusInterviewScheduled:
		if targetStatus == ApplicationStatusBackgroundCheck || targetStatus == ApplicationStatusDeclined || targetStatus == ApplicationStatusWithdrawn {
			return nil
		}
	case ApplicationStatusBackgroundCheck:
		if targetStatus == ApplicationStatusApproved || targetStatus == ApplicationStatusDeclined || targetStatus == ApplicationStatusWithdrawn {
			return nil
		}
	case ApplicationStatusApproved:
		if targetStatus == ApplicationStatusWithdrawn {
			return nil
		}
	case ApplicationStatusDeclined:
		// Cannot transition from declined
		return domain.NewValidationError("cannot transition from declined status")
	case ApplicationStatusWithdrawn:
		// Cannot transition from withdrawn
		return domain.NewValidationError("cannot transition from withdrawn status")
	}

	return domain.NewValidationError(fmt.Sprintf("cannot transition from %s to %s", v.Status, targetStatus))
}


