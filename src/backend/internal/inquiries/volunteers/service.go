package volunteers

import (
	"context"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/google/uuid"
)

// VolunteerRepositoryInterface defines the contract for volunteer data access
type VolunteerRepositoryInterface interface {
	// Volunteer application operations
	SaveVolunteerApplication(ctx context.Context, application *VolunteerApplication) error
	GetVolunteerApplication(ctx context.Context, applicationID string) (*VolunteerApplication, error)
	GetAllVolunteerApplications(ctx context.Context, limit, offset int) ([]*VolunteerApplication, error)
	GetVolunteerApplicationsByStatus(ctx context.Context, status ApplicationStatus, limit, offset int) ([]*VolunteerApplication, error)
	GetVolunteerApplicationsByPriority(ctx context.Context, priority ApplicationPriority, limit, offset int) ([]*VolunteerApplication, error)
	GetVolunteerApplicationsByInterest(ctx context.Context, interest VolunteerInterest, limit, offset int) ([]*VolunteerApplication, error)
	SearchVolunteerApplications(ctx context.Context, searchTerm string, limit, offset int) ([]*VolunteerApplication, error)
	DeleteVolunteerApplication(ctx context.Context, applicationID string) error

	// Audit operations
	PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error
	GetVolunteerApplicationAudit(ctx context.Context, applicationID string, userID string, limit int, offset int) ([]*domain.AuditEvent, error)
}

// VolunteerService implements business logic for volunteer operations
type VolunteerService struct {
	repository VolunteerRepositoryInterface
}

// NewVolunteerService creates a new volunteer service
func NewVolunteerService(repository VolunteerRepositoryInterface) *VolunteerService {
	return &VolunteerService{
		repository: repository,
	}
}

// ValidateVolunteerApplication validates a volunteer application
func ValidateVolunteerApplication(application *VolunteerApplication) error {
	if application.FirstName == "" {
		return domain.NewValidationFieldError("first_name", "first name cannot be empty")
	}
	if len(application.FirstName) < 2 || len(application.FirstName) > 50 {
		return domain.NewValidationFieldError("first_name", "first name must be between 2 and 50 characters")
	}

	if application.LastName == "" {
		return domain.NewValidationFieldError("last_name", "last name cannot be empty")
	}
	if len(application.LastName) < 2 || len(application.LastName) > 50 {
		return domain.NewValidationFieldError("last_name", "last name must be between 2 and 50 characters")
	}

	if application.Email == "" {
		return domain.NewValidationFieldError("email", "email cannot be empty")
	}
	if _, err := mail.ParseAddress(application.Email); err != nil {
		return domain.NewValidationFieldError("email", "invalid email format")
	}

	if application.Phone == "" {
		return domain.NewValidationFieldError("phone", "phone number cannot be empty")
	}
	phoneRegex := regexp.MustCompile(`^\d{10}$`)
	if !phoneRegex.MatchString(application.Phone) {
		return domain.NewValidationFieldError("phone", "phone number must be 10 digits")
	}

	if application.Age < 18 || application.Age > 100 {
		return domain.NewValidationFieldError("age", "age must be between 18 and 100")
	}

	if application.Motivation == "" {
		return domain.NewValidationFieldError("motivation", "motivation cannot be empty")
	}
	if len(application.Motivation) < 20 || len(application.Motivation) > 1500 {
		return domain.NewValidationFieldError("motivation", "motivation must be between 20 and 1500 characters")
	}

	if application.Experience != "" && len(application.Experience) > 1000 {
		return domain.NewValidationFieldError("experience", "experience cannot exceed 1000 characters")
	}

	if application.SchedulePreferences != "" && len(application.SchedulePreferences) > 500 {
		return domain.NewValidationFieldError("schedule_preferences", "schedule preferences cannot exceed 500 characters")
	}

	// Validate enum values
	if !IsValidApplicationStatus(application.Status) {
		return domain.NewValidationFieldError("status", "invalid application status")
	}

	if !IsValidApplicationPriority(application.Priority) {
		return domain.NewValidationFieldError("priority", "invalid application priority")
	}

	if !IsValidVolunteerInterest(application.VolunteerInterest) {
		return domain.NewValidationFieldError("volunteer_interest", "invalid volunteer interest")
	}

	if !IsValidAvailability(application.Availability) {
		return domain.NewValidationFieldError("availability", "invalid availability")
	}

	return nil
}

// IsValidApplicationStatus checks if application status is valid
func IsValidApplicationStatus(status ApplicationStatus) bool {
	switch status {
	case ApplicationStatusNew, ApplicationStatusUnderReview, ApplicationStatusInterviewScheduled, ApplicationStatusBackgroundCheck, ApplicationStatusApproved, ApplicationStatusDeclined, ApplicationStatusWithdrawn:
		return true
	default:
		return false
	}
}

// IsValidApplicationPriority checks if application priority is valid
func IsValidApplicationPriority(priority ApplicationPriority) bool {
	switch priority {
	case ApplicationPriorityLow, ApplicationPriorityMedium, ApplicationPriorityHigh, ApplicationPriorityUrgent:
		return true
	default:
		return false
	}
}

// IsValidVolunteerInterest checks if volunteer interest is valid
func IsValidVolunteerInterest(interest VolunteerInterest) bool {
	switch interest {
	case VolunteerInterestPatientSupport, VolunteerInterestCommunityOutreach, VolunteerInterestResearchSupport, VolunteerInterestAdministrativeSupport, VolunteerInterestMultiple, VolunteerInterestOther:
		return true
	default:
		return false
	}
}

// IsValidAvailability checks if availability is valid
func IsValidAvailability(availability Availability) bool {
	switch availability {
	case Availability2To4Hours, Availability4To8Hours, Availability8To16Hours, Availability16HoursPlus, AvailabilityFlexible:
		return true
	default:
		return false
	}
}

// IsAdminUser checks if user has admin privileges for audit operations
func IsAdminUser(userID string) bool {
	return strings.HasPrefix(userID, "admin-")
}

// GetVolunteerApplication retrieves a volunteer application by ID
func (s *VolunteerService) GetVolunteerApplication(ctx context.Context, applicationID string, userID string) (*VolunteerApplication, error) {
	return s.repository.GetVolunteerApplication(ctx, applicationID)
}

// GetAllVolunteerApplications retrieves all volunteer applications with pagination
func (s *VolunteerService) GetAllVolunteerApplications(ctx context.Context, limit, offset int) ([]*VolunteerApplication, error) {
	return s.repository.GetAllVolunteerApplications(ctx, limit, offset)
}

// SearchVolunteerApplications searches volunteer applications by query
func (s *VolunteerService) SearchVolunteerApplications(ctx context.Context, query string, limit, offset int) ([]*VolunteerApplication, error) {
	if query == "" {
		return s.GetAllVolunteerApplications(ctx, limit, offset)
	}
	return s.repository.SearchVolunteerApplications(ctx, query, limit, offset)
}

// GetVolunteerApplicationsByStatus retrieves applications by status
func (s *VolunteerService) GetVolunteerApplicationsByStatus(ctx context.Context, status ApplicationStatus, limit, offset int) ([]*VolunteerApplication, error) {
	return s.repository.GetVolunteerApplicationsByStatus(ctx, status, limit, offset)
}

// GetVolunteerApplicationsByInterest retrieves applications by volunteer interest
func (s *VolunteerService) GetVolunteerApplicationsByInterest(ctx context.Context, interest VolunteerInterest, limit, offset int) ([]*VolunteerApplication, error) {
	return s.repository.GetVolunteerApplicationsByInterest(ctx, interest, limit, offset)
}

// CreateVolunteerApplication creates a new volunteer application
func (s *VolunteerService) CreateVolunteerApplication(ctx context.Context, application *VolunteerApplication, userID string) error {
	// Validate the application
	if err := ValidateVolunteerApplication(application); err != nil {
		return err
	}

	// Generate UUID if not provided
	if application.ApplicationID == "" {
		application.ApplicationID = uuid.New().String()
	}

	// Set defaults
	if application.Status == "" {
		application.Status = ApplicationStatusNew
	}
	if application.Priority == "" {
		application.Priority = ApplicationPriorityMedium
	}
	if application.Source == "" {
		application.Source = "website"
	}

	// Set audit fields
	now := time.Now()
	application.CreatedAt = now
	application.UpdatedAt = now
	application.CreatedBy = userID
	application.UpdatedBy = userID
	application.IsDeleted = false

	// Save to repository
	if err := s.repository.SaveVolunteerApplication(ctx, application); err != nil {
		return domain.WrapError(err, "failed to save volunteer application")
	}

	// Publish audit event
	return s.repository.PublishAuditEvent(ctx, domain.EntityTypeVolunteerApplication, application.ApplicationID, domain.AuditEventInsert, userID, nil, application)
}

// UpdateVolunteerApplicationStatus updates the status of a volunteer application
func (s *VolunteerService) UpdateVolunteerApplicationStatus(ctx context.Context, applicationID string, status ApplicationStatus, userID string) error {
	// Validate status
	if !IsValidApplicationStatus(status) {
		return domain.NewValidationFieldError("status", "invalid application status")
	}

	// Get existing application
	existing, err := s.repository.GetVolunteerApplication(ctx, applicationID)
	if err != nil {
		return err
	}

	// Create a copy for before snapshot
	beforeData := *existing

	// Update status and audit fields
	existing.Status = status
	existing.UpdatedAt = time.Now()
	existing.UpdatedBy = userID

	// Save updated application
	if err := s.repository.SaveVolunteerApplication(ctx, existing); err != nil {
		return domain.WrapError(err, "failed to update volunteer application status")
	}

	// Publish audit event
	return s.repository.PublishAuditEvent(ctx, domain.EntityTypeVolunteerApplication, applicationID, domain.AuditEventUpdate, userID, &beforeData, existing)
}

// UpdateVolunteerApplicationPriority updates the priority of a volunteer application
func (s *VolunteerService) UpdateVolunteerApplicationPriority(ctx context.Context, applicationID string, priority ApplicationPriority, userID string) error {
	// Validate priority
	if !IsValidApplicationPriority(priority) {
		return domain.NewValidationFieldError("priority", "invalid application priority")
	}

	// Get existing application
	existing, err := s.repository.GetVolunteerApplication(ctx, applicationID)
	if err != nil {
		return err
	}

	// Create a copy for before snapshot
	beforeData := *existing

	// Update priority and audit fields
	existing.Priority = priority
	existing.UpdatedAt = time.Now()
	existing.UpdatedBy = userID

	// Save updated application
	if err := s.repository.SaveVolunteerApplication(ctx, existing); err != nil {
		return domain.WrapError(err, "failed to update volunteer application priority")
	}

	// Publish audit event
	return s.repository.PublishAuditEvent(ctx, domain.EntityTypeVolunteerApplication, applicationID, domain.AuditEventUpdate, userID, &beforeData, existing)
}

// DeleteVolunteerApplication soft deletes a volunteer application
func (s *VolunteerService) DeleteVolunteerApplication(ctx context.Context, applicationID string, userID string) error {
	// Get existing application for audit
	existing, err := s.repository.GetVolunteerApplication(ctx, applicationID)
	if err != nil {
		return err
	}

	// Create a copy for before snapshot
	beforeData := *existing

	// Perform soft delete
	if err := s.repository.DeleteVolunteerApplication(ctx, applicationID); err != nil {
		return domain.WrapError(err, "failed to delete volunteer application")
	}

	// Publish audit event
	return s.repository.PublishAuditEvent(ctx, domain.EntityTypeVolunteerApplication, applicationID, domain.AuditEventDelete, userID, &beforeData, nil)
}

// GetVolunteerApplicationAudit retrieves audit events for a volunteer application
func (s *VolunteerService) GetVolunteerApplicationAudit(ctx context.Context, applicationID string, userID string, limit, offset int) ([]*domain.AuditEvent, error) {
	// Check if user is admin
	if !IsAdminUser(userID) {
		return nil, domain.NewUnauthorizedError("only admin users can access audit data")
	}

	return s.repository.GetVolunteerApplicationAudit(ctx, applicationID, userID, limit, offset)
}