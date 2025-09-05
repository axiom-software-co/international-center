package volunteers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Domain Validation Method Tests (TDD RED Phase)

func TestApplicationStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status ApplicationStatus
		want   bool
	}{
		{"valid new status", ApplicationStatusNew, true},
		{"valid under review status", ApplicationStatusUnderReview, true},
		{"valid interview scheduled status", ApplicationStatusInterviewScheduled, true},
		{"valid background check status", ApplicationStatusBackgroundCheck, true},
		{"valid approved status", ApplicationStatusApproved, true},
		{"valid declined status", ApplicationStatusDeclined, true},
		{"valid withdrawn status", ApplicationStatusWithdrawn, true},
		{"invalid status", ApplicationStatus("invalid"), false},
		{"empty status", ApplicationStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestApplicationPriority_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		priority ApplicationPriority
		want     bool
	}{
		{"valid low priority", ApplicationPriorityLow, true},
		{"valid medium priority", ApplicationPriorityMedium, true},
		{"valid high priority", ApplicationPriorityHigh, true},
		{"valid urgent priority", ApplicationPriorityUrgent, true},
		{"invalid priority", ApplicationPriority("invalid"), false},
		{"empty priority", ApplicationPriority(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.priority.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestVolunteerInterest_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		interest VolunteerInterest
		want     bool
	}{
		{"valid patient support", VolunteerInterestPatientSupport, true},
		{"valid community outreach", VolunteerInterestCommunityOutreach, true},
		{"valid research support", VolunteerInterestResearchSupport, true},
		{"valid administrative support", VolunteerInterestAdministrativeSupport, true},
		{"valid multiple", VolunteerInterestMultiple, true},
		{"valid other", VolunteerInterestOther, true},
		{"invalid interest", VolunteerInterest("invalid"), false},
		{"empty interest", VolunteerInterest(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.interest.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAvailability_IsValid(t *testing.T) {
	tests := []struct {
		name         string
		availability Availability
		want         bool
	}{
		{"valid 2-4 hours", Availability2To4Hours, true},
		{"valid 4-8 hours", Availability4To8Hours, true},
		{"valid 8-16 hours", Availability8To16Hours, true},
		{"valid 16+ hours", Availability16HoursPlus, true},
		{"valid flexible", AvailabilityFlexible, true},
		{"invalid availability", Availability("invalid"), false},
		{"empty availability", Availability(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.availability.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewVolunteerApplication(t *testing.T) {
	tests := []struct {
		name      string
		firstName string
		lastName  string
		email     string
		phone     string
		age       int
		interest  VolunteerInterest
		availability Availability
		motivation string
		userID    string
		wantErr   bool
		checkFunc func(*testing.T, *VolunteerApplication)
	}{
		{
			name:         "successfully create volunteer application with valid data",
			firstName:    "John",
			lastName:     "Doe",
			email:        "john.doe@example.com",
			phone:        "5551234567",
			age:          25,
			interest:     VolunteerInterestPatientSupport,
			availability: Availability4To8Hours,
			motivation:   "I want to help patients in need and make a meaningful difference in their lives.",
			userID:       "system",
			wantErr:      false,
			checkFunc: func(t *testing.T, app *VolunteerApplication) {
				assert.NotEmpty(t, app.ApplicationID)
				assert.Equal(t, "John", app.FirstName)
				assert.Equal(t, "Doe", app.LastName)
				assert.Equal(t, "john.doe@example.com", app.Email)
				assert.Equal(t, ApplicationStatusNew, app.Status)
				assert.Equal(t, ApplicationPriorityMedium, app.Priority)
				assert.False(t, app.IsDeleted)
			},
		},
		{
			name:         "return validation error for empty first name",
			firstName:    "",
			lastName:     "Doe",
			email:        "john.doe@example.com",
			phone:        "5551234567",
			age:          25,
			interest:     VolunteerInterestPatientSupport,
			availability: Availability4To8Hours,
			motivation:   "I want to help patients in need.",
			userID:       "system",
			wantErr:      true,
		},
		{
			name:         "return validation error for short motivation",
			firstName:    "John",
			lastName:     "Doe",
			email:        "john.doe@example.com",
			phone:        "5551234567",
			age:          25,
			interest:     VolunteerInterestPatientSupport,
			availability: Availability4To8Hours,
			motivation:   "Short",
			userID:       "system",
			wantErr:      true,
		},
		{
			name:         "return validation error for invalid volunteer interest",
			firstName:    "John",
			lastName:     "Doe",
			email:        "john.doe@example.com",
			phone:        "5551234567",
			age:          25,
			interest:     VolunteerInterest("invalid"),
			availability: Availability4To8Hours,
			motivation:   "I want to help patients in need and make a difference.",
			userID:       "system",
			wantErr:      true,
		},
		{
			name:         "return validation error for underage applicant",
			firstName:    "John",
			lastName:     "Doe",
			email:        "john.doe@example.com",
			phone:        "5551234567",
			age:          17,
			interest:     VolunteerInterestPatientSupport,
			availability: Availability4To8Hours,
			motivation:   "I want to help patients in need and make a difference.",
			userID:       "system",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewVolunteerApplication(tt.firstName, tt.lastName, tt.email, tt.phone, tt.age, tt.interest, tt.availability, tt.motivation, tt.userID)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

func TestVolunteerApplication_Validate(t *testing.T) {
	tests := []struct {
		name        string
		application *VolunteerApplication
		wantErr     bool
		wantMsgContains string
	}{
		{
			name: "valid volunteer application",
			application: &VolunteerApplication{
				ApplicationID:     "550e8400-e29b-41d4-a716-446655440001",
				FirstName:         "John",
				LastName:          "Doe", 
				Email:             "john.doe@example.com",
				Phone:             "5551234567",
				Age:               25,
				VolunteerInterest: VolunteerInterestPatientSupport,
				Availability:      Availability4To8Hours,
				Motivation:        "I want to help patients in need and make a meaningful difference in their lives.",
				Status:            ApplicationStatusNew,
				Priority:          ApplicationPriorityMedium,
				Source:            "website",
				CreatedBy:         "system",
				UpdatedBy:         "system",
			},
			wantErr: false,
		},
		{
			name: "invalid application with empty first name",
			application: &VolunteerApplication{
				ApplicationID: "550e8400-e29b-41d4-a716-446655440001",
				FirstName:     "",
				LastName:      "Doe",
				Email:         "john.doe@example.com",
				Age:           25,
			},
			wantErr:             true,
			wantMsgContains:     "first name",
		},
		{
			name: "invalid application with invalid email",
			application: &VolunteerApplication{
				ApplicationID: "550e8400-e29b-41d4-a716-446655440001",
				FirstName:     "John",
				LastName:      "Doe",
				Email:         "invalid-email",
				Age:           25,
			},
			wantErr:             true,
			wantMsgContains:     "email",
		},
		{
			name: "invalid application with invalid priority",
			application: &VolunteerApplication{
				ApplicationID:     "550e8400-e29b-41d4-a716-446655440001",
				FirstName:         "John",
				LastName:          "Doe",
				Email:             "john.doe@example.com",
				Phone:             "5551234567",
				Age:               25,
				Status:            ApplicationStatusNew,
				VolunteerInterest: VolunteerInterestPatientSupport,
				Availability:      Availability4To8Hours,
				Motivation:        "I want to help patients in need and make a meaningful difference in their lives.",
				Priority:          ApplicationPriority("invalid"),
				CreatedBy:         "system",
				UpdatedBy:         "system",
			},
			wantErr:             true,
			wantMsgContains:     "priority",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.application.Validate()
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantMsgContains != "" {
					assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.wantMsgContains))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVolunteerApplication_SetPriority(t *testing.T) {
	tests := []struct {
		name     string
		priority ApplicationPriority
		userID   string
		wantErr  bool
	}{
		{
			name:     "successfully set priority to high",
			priority: ApplicationPriorityHigh,
			userID:   "admin-user",
			wantErr:  false,
		},
		{
			name:     "successfully set priority to urgent",
			priority: ApplicationPriorityUrgent,
			userID:   "admin-user",
			wantErr:  false,
		},
		{
			name:     "return validation error for invalid priority",
			priority: ApplicationPriority("invalid"),
			userID:   "admin-user",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &VolunteerApplication{
				ApplicationID: "550e8400-e29b-41d4-a716-446655440001",
				Priority:      ApplicationPriorityMedium,
				UpdatedBy:     "original-user",
			}
			originalUpdatedAt := app.UpdatedAt
			
			err := app.SetPriority(tt.priority, tt.userID)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.priority, app.Priority)
				assert.Equal(t, tt.userID, app.UpdatedBy)
				assert.True(t, app.UpdatedAt.After(originalUpdatedAt))
			}
		})
	}
}

func TestVolunteerApplication_UpdateStatus(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus ApplicationStatus
		newStatus     ApplicationStatus
		userID        string
		wantErr       bool
	}{
		{
			name:          "successfully update status to under review from new",
			initialStatus: ApplicationStatusNew,
			newStatus:     ApplicationStatusUnderReview,
			userID:        "admin-user",
			wantErr:       false,
		},
		{
			name:          "successfully update status to approved from background check",
			initialStatus: ApplicationStatusBackgroundCheck,
			newStatus:     ApplicationStatusApproved,
			userID:        "admin-user",
			wantErr:       false,
		},
		{
			name:          "return validation error for invalid status",
			initialStatus: ApplicationStatusNew,
			newStatus:     ApplicationStatus("invalid"),
			userID:        "admin-user",
			wantErr:       true,
		},
		{
			name:          "return validation error for invalid transition",
			initialStatus: ApplicationStatusNew,
			newStatus:     ApplicationStatusApproved,
			userID:        "admin-user",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &VolunteerApplication{
				ApplicationID: "550e8400-e29b-41d4-a716-446655440001",
				Status:        tt.initialStatus,
				UpdatedBy:     "original-user",
			}
			originalUpdatedAt := app.UpdatedAt
			
			err := app.UpdateStatus(tt.newStatus, tt.userID)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.newStatus, app.Status)
				assert.Equal(t, tt.userID, app.UpdatedBy)
				assert.True(t, app.UpdatedAt.After(originalUpdatedAt))
			}
		})
	}
}

func TestVolunteerApplication_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name         string
		currentStatus ApplicationStatus
		targetStatus ApplicationStatus
		wantErr      bool
	}{
		{
			name:          "can transition from new to under review",
			currentStatus: ApplicationStatusNew,
			targetStatus:  ApplicationStatusUnderReview,
			wantErr:       false,
		},
		{
			name:          "can transition from under review to interview scheduled",
			currentStatus: ApplicationStatusUnderReview,
			targetStatus:  ApplicationStatusInterviewScheduled,
			wantErr:       false,
		},
		{
			name:          "can transition from interview scheduled to background check",
			currentStatus: ApplicationStatusInterviewScheduled,
			targetStatus:  ApplicationStatusBackgroundCheck,
			wantErr:       false,
		},
		{
			name:          "can transition from background check to approved",
			currentStatus: ApplicationStatusBackgroundCheck,
			targetStatus:  ApplicationStatusApproved,
			wantErr:       false,
		},
		{
			name:          "can transition from any status to declined",
			currentStatus: ApplicationStatusUnderReview,
			targetStatus:  ApplicationStatusDeclined,
			wantErr:       false,
		},
		{
			name:          "can transition from any status to withdrawn",
			currentStatus: ApplicationStatusInterviewScheduled,
			targetStatus:  ApplicationStatusWithdrawn,
			wantErr:       false,
		},
		{
			name:          "cannot transition from new directly to approved",
			currentStatus: ApplicationStatusNew,
			targetStatus:  ApplicationStatusApproved,
			wantErr:       true,
		},
		{
			name:          "cannot transition from approved to new",
			currentStatus: ApplicationStatusApproved,
			targetStatus:  ApplicationStatusNew,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &VolunteerApplication{
				Status: tt.currentStatus,
			}
			
			err := app.CanTransitionTo(tt.targetStatus)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}