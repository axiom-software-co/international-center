package volunteers

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	sharedtesting "github.com/axiom-software-co/international-center/src/backend/internal/shared/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockVolunteerRepository provides mock implementation for unit tests
type MockVolunteerRepository struct {
	applications map[string]*VolunteerApplication
	auditEvents  []MockAuditEvent
	failures     map[string]error
}

type MockAuditEvent struct {
	EntityType    domain.EntityType
	EntityID      string
	OperationType domain.AuditEventType
	UserID        string
	Before        interface{}
	After         interface{}
}

func NewMockVolunteerRepository() *MockVolunteerRepository {
	return &MockVolunteerRepository{
		applications: make(map[string]*VolunteerApplication),
		auditEvents:  make([]MockAuditEvent, 0),
		failures:     make(map[string]error),
	}
}

// SetFailure sets a mock failure for specific operations
func (m *MockVolunteerRepository) SetFailure(operation string, err error) {
	m.failures[operation] = err
}

// GetAuditEvents returns all mock audit events
func (m *MockVolunteerRepository) GetAuditEvents() []MockAuditEvent {
	return m.auditEvents
}

// Volunteer application CRUD operations
func (m *MockVolunteerRepository) GetVolunteerApplication(ctx context.Context, applicationID string) (*VolunteerApplication, error) {
	if err, exists := m.failures["GetVolunteerApplication"]; exists {
		return nil, err
	}
	if application, exists := m.applications[applicationID]; exists && !application.IsDeleted {
		return application, nil
	}
	return nil, domain.NewNotFoundError("volunteer_application", applicationID)
}

func (m *MockVolunteerRepository) GetAllVolunteerApplications(ctx context.Context, limit, offset int) ([]*VolunteerApplication, error) {
	if err, exists := m.failures["GetAllVolunteerApplications"]; exists {
		return nil, err
	}
	
	applicationsList := make([]*VolunteerApplication, 0)
	for _, application := range m.applications {
		if !application.IsDeleted {
			applicationsList = append(applicationsList, application)
		}
	}
	return applicationsList, nil
}

func (m *MockVolunteerRepository) GetVolunteerApplicationsByStatus(ctx context.Context, status ApplicationStatus, limit, offset int) ([]*VolunteerApplication, error) {
	if err, exists := m.failures["GetVolunteerApplicationsByStatus"]; exists {
		return nil, err
	}
	
	applicationsList := make([]*VolunteerApplication, 0)
	count := 0
	for _, application := range m.applications {
		if application.Status == status && !application.IsDeleted {
			// Apply offset
			if count < offset {
				count++
				continue
			}
			// Apply limit
			if len(applicationsList) >= limit {
				break
			}
			applicationsList = append(applicationsList, application)
		}
		count++
	}
	return applicationsList, nil
}

func (m *MockVolunteerRepository) GetVolunteerApplicationsByPriority(ctx context.Context, priority ApplicationPriority, limit, offset int) ([]*VolunteerApplication, error) {
	if err, exists := m.failures["GetVolunteerApplicationsByPriority"]; exists {
		return nil, err
	}
	
	applicationsList := make([]*VolunteerApplication, 0)
	for _, application := range m.applications {
		if application.Priority == priority && !application.IsDeleted {
			applicationsList = append(applicationsList, application)
		}
	}
	return applicationsList, nil
}

func (m *MockVolunteerRepository) GetVolunteerApplicationsByInterest(ctx context.Context, interest VolunteerInterest, limit, offset int) ([]*VolunteerApplication, error) {
	if err, exists := m.failures["GetVolunteerApplicationsByInterest"]; exists {
		return nil, err
	}
	
	applicationsList := make([]*VolunteerApplication, 0)
	for _, application := range m.applications {
		if application.VolunteerInterest == interest && !application.IsDeleted {
			applicationsList = append(applicationsList, application)
		}
	}
	return applicationsList, nil
}

func (m *MockVolunteerRepository) SearchVolunteerApplications(ctx context.Context, query string, limit, offset int) ([]*VolunteerApplication, error) {
	if err, exists := m.failures["SearchVolunteerApplications"]; exists {
		return nil, err
	}
	
	applicationsList := make([]*VolunteerApplication, 0)
	queryLower := strings.ToLower(query)
	for _, application := range m.applications {
		if !application.IsDeleted {
			if strings.Contains(strings.ToLower(application.FirstName+" "+application.LastName), queryLower) ||
			   strings.Contains(strings.ToLower(application.Email), queryLower) ||
			   strings.Contains(strings.ToLower(application.Motivation), queryLower) {
				applicationsList = append(applicationsList, application)
			}
		}
	}
	return applicationsList, nil
}

func (m *MockVolunteerRepository) SaveVolunteerApplication(ctx context.Context, application *VolunteerApplication) error {
	if err, exists := m.failures["SaveVolunteerApplication"]; exists {
		return err
	}
	m.applications[application.ApplicationID] = application
	return nil
}

func (m *MockVolunteerRepository) DeleteVolunteerApplication(ctx context.Context, applicationID string) error {
	if err, exists := m.failures["DeleteVolunteerApplication"]; exists {
		return err
	}
	if application, exists := m.applications[applicationID]; exists {
		application.IsDeleted = true
		now := time.Now()
		application.DeletedAt = &now
		return nil
	}
	return domain.NewNotFoundError("volunteer_application", applicationID)
}

// Audit repository methods
func (m *MockVolunteerRepository) GetVolunteerApplicationAudit(ctx context.Context, applicationID string, userID string, limit int, offset int) ([]*domain.AuditEvent, error) {
	if err, exists := m.failures["GetVolunteerApplicationAudit"]; exists {
		return nil, err
	}
	// Return mock audit events for this application
	events := make([]*domain.AuditEvent, 0)
	for _, auditEvent := range m.auditEvents {
		if auditEvent.EntityID == applicationID {
			events = append(events, &domain.AuditEvent{
				AuditID:       "audit-" + applicationID + "-1",
				EntityType:    auditEvent.EntityType,
				EntityID:      auditEvent.EntityID,
				OperationType: auditEvent.OperationType,
				AuditTime:     time.Now().Add(-time.Hour),
				UserID:        auditEvent.UserID,
			})
		}
	}
	return events, nil
}

// Audit event publishing method
func (m *MockVolunteerRepository) PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error {
	if err, exists := m.failures["PublishAuditEvent"]; exists {
		return err
	}
	// Record the audit event in memory for testing
	m.auditEvents = append(m.auditEvents, MockAuditEvent{
		EntityType:    entityType,
		EntityID:      entityID,
		OperationType: operationType,
		UserID:        userID,
		Before:        beforeData,
		After:         afterData,
	})
	return nil
}

// Test Volunteer Service Operations

func TestVolunteerService_GetVolunteerApplication(t *testing.T) {
	tests := []struct {
		name          string
		applicationID string
		setupFunc     func(*MockVolunteerRepository)
		wantErr       bool
		wantEmail     string
	}{
		{
			name:          "successfully retrieve existing volunteer application",
			applicationID: "550e8400-e29b-41d4-a716-446655440001",
			setupFunc: func(repo *MockVolunteerRepository) {
				application := &VolunteerApplication{
					ApplicationID:     "550e8400-e29b-41d4-a716-446655440001",
					FirstName:         "John",
					LastName:          "Doe",
					Email:             "john.doe@example.com",
					Phone:             "5551234567",
					Age:               25,
					VolunteerInterest: VolunteerInterestPatientSupport,
					Availability:      Availability4To8Hours,
					Motivation:        "I want to help patients in need and make a difference in their lives.",
					Status:            ApplicationStatusNew,
					Priority:          ApplicationPriorityMedium,
					IsDeleted:         false,
				}
				repo.applications["550e8400-e29b-41d4-a716-446655440001"] = application
			},
			wantErr:   false,
			wantEmail: "john.doe@example.com",
		},
		{
			name:          "return not found error for non-existent application",
			applicationID: "nonexistent",
			setupFunc:     func(repo *MockVolunteerRepository) {},
			wantErr:       true,
		},
		{
			name:          "return not found error for soft deleted application",
			applicationID: "deleted-application",
			setupFunc: func(repo *MockVolunteerRepository) {
				application := &VolunteerApplication{
					ApplicationID: "deleted-application",
					FirstName:     "Jane",
					LastName:      "Smith",
					IsDeleted:     true,
				}
				repo.applications["deleted-application"] = application
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			repo := NewMockVolunteerRepository()
			tt.setupFunc(repo)
			
			service := NewVolunteerService(repo)
			
			result, err := service.GetVolunteerApplication(ctx, tt.applicationID, "550e8400-e29b-41d4-a716-446655440004")
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.wantEmail, result.Email)
			}
		})
	}
}

func TestVolunteerService_GetAllVolunteerApplications(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*MockVolunteerRepository)
		wantCount int
		wantErr   bool
	}{
		{
			name: "successfully retrieve all volunteer applications",
			setupFunc: func(repo *MockVolunteerRepository) {
				repo.applications["550e8400-e29b-41d4-a716-446655440001"] = &VolunteerApplication{
					ApplicationID: "550e8400-e29b-41d4-a716-446655440001",
					FirstName:     "John",
					LastName:      "Doe",
					Email:         "john.doe@example.com",
					IsDeleted:     false,
				}
				repo.applications["application-2"] = &VolunteerApplication{
					ApplicationID: "application-2",
					FirstName:     "Jane",
					LastName:      "Smith",
					Email:         "jane.smith@example.com",
					IsDeleted:     false,
				}
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "return empty array when no applications exist",
			setupFunc: func(repo *MockVolunteerRepository) {},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "exclude soft deleted applications from results",
			setupFunc: func(repo *MockVolunteerRepository) {
				repo.applications["550e8400-e29b-41d4-a716-446655440001"] = &VolunteerApplication{
					ApplicationID: "550e8400-e29b-41d4-a716-446655440001",
					FirstName:     "John",
					LastName:      "Doe",
					IsDeleted:     false,
				}
				repo.applications["application-2"] = &VolunteerApplication{
					ApplicationID: "application-2",
					FirstName:     "Jane",
					LastName:      "Smith",
					IsDeleted:     true,
				}
			},
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			repo := NewMockVolunteerRepository()
			tt.setupFunc(repo)
			
			service := NewVolunteerService(repo)
			
			result, err := service.GetAllVolunteerApplications(ctx, 10, 0)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, tt.wantCount)
			}
		})
	}
}

func TestVolunteerService_SearchVolunteerApplications(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		setupFunc func(*MockVolunteerRepository)
		wantCount int
		wantErr   bool
	}{
		{
			name:  "successfully search applications by name",
			query: "john",
			setupFunc: func(repo *MockVolunteerRepository) {
				repo.applications["550e8400-e29b-41d4-a716-446655440001"] = &VolunteerApplication{
					ApplicationID: "550e8400-e29b-41d4-a716-446655440001",
					FirstName:     "John",
					LastName:      "Doe",
					Email:         "john.doe@example.com",
					Motivation:    "I want to help patients",
					IsDeleted:     false,
				}
				repo.applications["application-2"] = &VolunteerApplication{
					ApplicationID: "application-2",
					FirstName:     "Jane",
					LastName:      "Smith",
					Email:         "jane.smith@example.com",
					Motivation:    "Community service",
					IsDeleted:     false,
				}
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:  "successfully search applications by email",
			query: "jane.smith",
			setupFunc: func(repo *MockVolunteerRepository) {
				repo.applications["550e8400-e29b-41d4-a716-446655440001"] = &VolunteerApplication{
					ApplicationID: "550e8400-e29b-41d4-a716-446655440001",
					FirstName:     "John",
					LastName:      "Doe",
					Email:         "john.doe@example.com",
					IsDeleted:     false,
				}
				repo.applications["application-2"] = &VolunteerApplication{
					ApplicationID: "application-2",
					FirstName:     "Jane",
					LastName:      "Smith",
					Email:         "jane.smith@example.com",
					IsDeleted:     false,
				}
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:  "successfully search applications by motivation",
			query: "help patients",
			setupFunc: func(repo *MockVolunteerRepository) {
				repo.applications["550e8400-e29b-41d4-a716-446655440001"] = &VolunteerApplication{
					ApplicationID: "550e8400-e29b-41d4-a716-446655440001",
					FirstName:     "John",
					LastName:      "Doe",
					Email:         "john.doe@example.com",
					Motivation:    "I want to help patients and make a difference",
					IsDeleted:     false,
				}
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:  "return empty results when no matches found",
			query: "nonexistent",
			setupFunc: func(repo *MockVolunteerRepository) {
				repo.applications["550e8400-e29b-41d4-a716-446655440001"] = &VolunteerApplication{
					ApplicationID: "550e8400-e29b-41d4-a716-446655440001",
					FirstName:     "John",
					LastName:      "Doe",
					Email:         "john.doe@example.com",
					IsDeleted:     false,
				}
			},
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			repo := NewMockVolunteerRepository()
			tt.setupFunc(repo)
			
			service := NewVolunteerService(repo)
			
			result, err := service.SearchVolunteerApplications(ctx, tt.query, 10, 0)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, tt.wantCount)
			}
		})
	}
}

func TestVolunteerService_GetVolunteerApplicationAudit(t *testing.T) {
	tests := []struct {
		name          string
		applicationID string
		userID        string
		setupFunc     func(*MockVolunteerRepository)
		wantErr       bool
		wantCount     int
	}{
		{
			name:          "successfully retrieve volunteer application audit events",
			applicationID: "550e8400-e29b-41d4-a716-446655440001",
			userID:        "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockVolunteerRepository) {
				repo.auditEvents = append(repo.auditEvents, MockAuditEvent{
					EntityType:    domain.EntityTypeVolunteerApplication,
					EntityID:      "550e8400-e29b-41d4-a716-446655440001",
					OperationType: domain.AuditEventInsert,
					UserID:        "admin-550e8400-e29b-41d4-a716-446655440003",
				})
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name:          "return empty array for application with no audit events",
			applicationID: "application-2",
			userID:        "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc:     func(repo *MockVolunteerRepository) {},
			wantErr:       false,
			wantCount:     0,
		},
		{
			name:          "return unauthorized error for non-admin user",
			applicationID: "550e8400-e29b-41d4-a716-446655440001",
			userID:        "550e8400-e29b-41d4-a716-446655440004",
			setupFunc:     func(repo *MockVolunteerRepository) {},
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			repo := NewMockVolunteerRepository()
			tt.setupFunc(repo)
			
			service := NewVolunteerService(repo)
			
			result, err := service.GetVolunteerApplicationAudit(ctx, tt.applicationID, tt.userID, 10, 0)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, tt.wantCount)
			}
		})
	}
}

// Admin CRUD Operation Tests

func TestVolunteerService_CreateVolunteerApplication(t *testing.T) {
	tests := []struct {
		name        string
		application *VolunteerApplication
		userID      string
		setupFunc   func(*MockVolunteerRepository)
		wantErr     bool
	}{
		{
			name: "successfully create new volunteer application",
			application: &VolunteerApplication{
				FirstName:         "John",
				LastName:          "Doe",
				Email:             "john.doe@example.com",
				Phone:             "5551234567",
				Age:               25,
				VolunteerInterest: VolunteerInterestPatientSupport,
				Availability:      Availability4To8Hours,
				Motivation:        "I want to help patients in need and make a difference in their lives.",
				Status:            ApplicationStatusNew,
				Priority:          ApplicationPriorityMedium,
				Source:            "website",
			},
			userID:    "system",
			setupFunc: func(repo *MockVolunteerRepository) {},
			wantErr:   false,
		},
		{
			name: "return validation error for invalid application",
			application: &VolunteerApplication{
				FirstName: "", // Invalid: empty name
			},
			userID:    "system",
			setupFunc: func(repo *MockVolunteerRepository) {},
			wantErr:   true,
		},
		{
			name: "return validation error for invalid age",
			application: &VolunteerApplication{
				FirstName:         "John",
				LastName:          "Doe",
				Email:             "john.doe@example.com",
				Phone:             "5551234567",
				Age:               17, // Invalid: under 18
				VolunteerInterest: VolunteerInterestPatientSupport,
				Availability:      Availability4To8Hours,
				Motivation:        "I want to help patients in need and make a difference.",
			},
			userID:    "system",
			setupFunc: func(repo *MockVolunteerRepository) {},
			wantErr:   true,
		},
		{
			name: "return validation error for short motivation",
			application: &VolunteerApplication{
				FirstName:         "John",
				LastName:          "Doe",
				Email:             "john.doe@example.com",
				Phone:             "5551234567",
				Age:               25,
				VolunteerInterest: VolunteerInterestPatientSupport,
				Availability:      Availability4To8Hours,
				Motivation:        "Short", // Invalid: too short
			},
			userID:    "system",
			setupFunc: func(repo *MockVolunteerRepository) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			repo := NewMockVolunteerRepository()
			tt.setupFunc(repo)
			
			service := NewVolunteerService(repo)
			
			err := service.CreateVolunteerApplication(ctx, tt.application, tt.userID)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify audit event was published
				events := repo.GetAuditEvents()
				assert.Len(t, events, 1)
				assert.Equal(t, domain.EntityTypeVolunteerApplication, events[0].EntityType)
				assert.Equal(t, domain.AuditEventInsert, events[0].OperationType)
			}
		})
	}
}

func TestVolunteerService_UpdateVolunteerApplicationStatus(t *testing.T) {
	tests := []struct {
		name          string
		applicationID string
		status        ApplicationStatus
		userID        string
		setupFunc     func(*MockVolunteerRepository)
		wantErr       bool
	}{
		{
			name:          "successfully update application status to under review",
			applicationID: "550e8400-e29b-41d4-a716-446655440001",
			status:        ApplicationStatusUnderReview,
			userID:        "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockVolunteerRepository) {
				existing := &VolunteerApplication{
					ApplicationID:     "550e8400-e29b-41d4-a716-446655440001",
					FirstName:         "John",
					LastName:          "Doe",
					Email:             "john.doe@example.com",
					Phone:             "5551234567",
					Age:               25,
					VolunteerInterest: VolunteerInterestPatientSupport,
					Availability:      Availability4To8Hours,
					Motivation:        "I want to help patients in need.",
					Status:            ApplicationStatusNew,
					Priority:          ApplicationPriorityMedium,
					CreatedAt:         time.Now(),
				}
				repo.applications["550e8400-e29b-41d4-a716-446655440001"] = existing
			},
			wantErr: false,
		},
		{
			name:          "successfully update application status to approved",
			applicationID: "550e8400-e29b-41d4-a716-446655440001",
			status:        ApplicationStatusApproved,
			userID:        "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockVolunteerRepository) {
				existing := &VolunteerApplication{
					ApplicationID:     "550e8400-e29b-41d4-a716-446655440001",
					FirstName:         "John",
					LastName:          "Doe",
					Email:             "john.doe@example.com",
					Status:            ApplicationStatusBackgroundCheck,
					CreatedAt:         time.Now(),
				}
				repo.applications["550e8400-e29b-41d4-a716-446655440001"] = existing
			},
			wantErr: false,
		},
		{
			name:          "return not found error for non-existent application",
			applicationID: "nonexistent",
			status:        ApplicationStatusUnderReview,
			userID:        "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc:     func(repo *MockVolunteerRepository) {},
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			repo := NewMockVolunteerRepository()
			tt.setupFunc(repo)
			
			service := NewVolunteerService(repo)
			
			err := service.UpdateVolunteerApplicationStatus(ctx, tt.applicationID, tt.status, tt.userID)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify status was updated
				application := repo.applications[tt.applicationID]
				assert.Equal(t, tt.status, application.Status)
				// Verify audit event was published
				events := repo.GetAuditEvents()
				assert.Len(t, events, 1)
				assert.Equal(t, domain.EntityTypeVolunteerApplication, events[0].EntityType)
				assert.Equal(t, domain.AuditEventUpdate, events[0].OperationType)
			}
		})
	}
}

func TestVolunteerService_UpdateVolunteerApplicationPriority(t *testing.T) {
	tests := []struct {
		name          string
		applicationID string
		priority      ApplicationPriority
		userID        string
		setupFunc     func(*MockVolunteerRepository)
		wantErr       bool
	}{
		{
			name:          "successfully update application priority to high",
			applicationID: "550e8400-e29b-41d4-a716-446655440001",
			priority:      ApplicationPriorityHigh,
			userID:        "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockVolunteerRepository) {
				existing := &VolunteerApplication{
					ApplicationID: "550e8400-e29b-41d4-a716-446655440001",
					FirstName:     "John",
					LastName:      "Doe",
					Email:         "john.doe@example.com",
					Priority:      ApplicationPriorityMedium,
					CreatedAt:     time.Now(),
				}
				repo.applications["550e8400-e29b-41d4-a716-446655440001"] = existing
			},
			wantErr: false,
		},
		{
			name:          "return not found error for non-existent application",
			applicationID: "nonexistent",
			priority:      ApplicationPriorityHigh,
			userID:        "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc:     func(repo *MockVolunteerRepository) {},
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			repo := NewMockVolunteerRepository()
			tt.setupFunc(repo)
			
			service := NewVolunteerService(repo)
			
			err := service.UpdateVolunteerApplicationPriority(ctx, tt.applicationID, tt.priority, tt.userID)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify priority was updated
				application := repo.applications[tt.applicationID]
				assert.Equal(t, tt.priority, application.Priority)
				// Verify audit event was published
				events := repo.GetAuditEvents()
				assert.Len(t, events, 1)
				assert.Equal(t, domain.EntityTypeVolunteerApplication, events[0].EntityType)
				assert.Equal(t, domain.AuditEventUpdate, events[0].OperationType)
			}
		})
	}
}

func TestVolunteerService_DeleteVolunteerApplication(t *testing.T) {
	tests := []struct {
		name          string
		applicationID string
		userID        string
		setupFunc     func(*MockVolunteerRepository)
		wantErr       bool
	}{
		{
			name:          "successfully delete existing volunteer application",
			applicationID: "550e8400-e29b-41d4-a716-446655440001",
			userID:        "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc: func(repo *MockVolunteerRepository) {
				application := &VolunteerApplication{
					ApplicationID:     "550e8400-e29b-41d4-a716-446655440001",
					FirstName:         "John",
					LastName:          "Doe",
					Email:             "john.doe@example.com",
					Phone:             "5551234567",
					Age:               25,
					VolunteerInterest: VolunteerInterestPatientSupport,
					Availability:      Availability4To8Hours,
					Motivation:        "I want to help patients in need.",
					Status:            ApplicationStatusNew,
					Priority:          ApplicationPriorityMedium,
					CreatedAt:         time.Now(),
				}
				repo.applications["550e8400-e29b-41d4-a716-446655440001"] = application
			},
			wantErr: false,
		},
		{
			name:          "return not found error for non-existent application",
			applicationID: "nonexistent",
			userID:        "admin-550e8400-e29b-41d4-a716-446655440003",
			setupFunc:     func(repo *MockVolunteerRepository) {},
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			repo := NewMockVolunteerRepository()
			tt.setupFunc(repo)
			
			service := NewVolunteerService(repo)
			
			err := service.DeleteVolunteerApplication(ctx, tt.applicationID, tt.userID)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify application is soft deleted
				if application, exists := repo.applications[tt.applicationID]; exists {
					assert.True(t, application.IsDeleted)
				}
				// Verify audit event was published
				events := repo.GetAuditEvents()
				assert.Len(t, events, 1)
				assert.Equal(t, domain.EntityTypeVolunteerApplication, events[0].EntityType)
				assert.Equal(t, domain.AuditEventDelete, events[0].OperationType)
			}
		})
	}
}

func TestVolunteerService_GetVolunteerApplicationsByStatus(t *testing.T) {
	tests := []struct {
		name      string
		status    ApplicationStatus
		setupFunc func(*MockVolunteerRepository)
		wantCount int
		wantErr   bool
	}{
		{
			name:   "successfully retrieve applications by status",
			status: ApplicationStatusNew,
			setupFunc: func(repo *MockVolunteerRepository) {
				repo.applications["550e8400-e29b-41d4-a716-446655440001"] = &VolunteerApplication{
					ApplicationID: "550e8400-e29b-41d4-a716-446655440001",
					FirstName:     "John",
					LastName:      "Doe",
					Status:        ApplicationStatusNew,
					IsDeleted:     false,
				}
				repo.applications["application-2"] = &VolunteerApplication{
					ApplicationID: "application-2",
					FirstName:     "Jane",
					LastName:      "Smith",
					Status:        ApplicationStatusUnderReview,
					IsDeleted:     false,
				}
				repo.applications["application-3"] = &VolunteerApplication{
					ApplicationID: "application-3",
					FirstName:     "Bob",
					LastName:      "Johnson",
					Status:        ApplicationStatusNew,
					IsDeleted:     false,
				}
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:   "return empty array when no applications match status",
			status: ApplicationStatusApproved,
			setupFunc: func(repo *MockVolunteerRepository) {
				repo.applications["550e8400-e29b-41d4-a716-446655440001"] = &VolunteerApplication{
					ApplicationID: "550e8400-e29b-41d4-a716-446655440001",
					Status:        ApplicationStatusNew,
					IsDeleted:     false,
				}
			},
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			repo := NewMockVolunteerRepository()
			tt.setupFunc(repo)
			
			service := NewVolunteerService(repo)
			
			result, err := service.GetVolunteerApplicationsByStatus(ctx, tt.status, 10, 0)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, tt.wantCount)
			}
		})
	}
}

func TestVolunteerService_GetVolunteerApplicationsByInterest(t *testing.T) {
	tests := []struct {
		name      string
		interest  VolunteerInterest
		setupFunc func(*MockVolunteerRepository)
		wantCount int
		wantErr   bool
	}{
		{
			name:     "successfully retrieve applications by volunteer interest",
			interest: VolunteerInterestPatientSupport,
			setupFunc: func(repo *MockVolunteerRepository) {
				repo.applications["550e8400-e29b-41d4-a716-446655440001"] = &VolunteerApplication{
					ApplicationID:     "550e8400-e29b-41d4-a716-446655440001",
					FirstName:         "John",
					LastName:          "Doe",
					VolunteerInterest: VolunteerInterestPatientSupport,
					IsDeleted:         false,
				}
				repo.applications["application-2"] = &VolunteerApplication{
					ApplicationID:     "application-2",
					FirstName:         "Jane",
					LastName:          "Smith",
					VolunteerInterest: VolunteerInterestCommunityOutreach,
					IsDeleted:         false,
				}
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:     "return empty array when no applications match interest",
			interest: VolunteerInterestResearchSupport,
			setupFunc: func(repo *MockVolunteerRepository) {
				repo.applications["550e8400-e29b-41d4-a716-446655440001"] = &VolunteerApplication{
					ApplicationID:     "550e8400-e29b-41d4-a716-446655440001",
					VolunteerInterest: VolunteerInterestPatientSupport,
					IsDeleted:         false,
				}
			},
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			repo := NewMockVolunteerRepository()
			tt.setupFunc(repo)
			
			service := NewVolunteerService(repo)
			
			result, err := service.GetVolunteerApplicationsByInterest(ctx, tt.interest, 10, 0)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, tt.wantCount)
			}
		})
	}
}