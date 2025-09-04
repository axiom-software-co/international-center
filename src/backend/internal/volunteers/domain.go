package volunteers

import (
	"context"
	"net"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
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

