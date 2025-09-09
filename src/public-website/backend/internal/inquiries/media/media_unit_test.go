package media

import (
	"context"
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to assert error type
func assertErrorType(t *testing.T, err error, expectedType string) {
	switch expectedType {
	case "validation":
		assert.True(t, domain.IsValidationError(err), "expected validation error")
	case "not_found":
		assert.True(t, domain.IsNotFoundError(err), "expected not found error")
	case "unauthorized":
		assert.True(t, domain.IsUnauthorizedError(err), "expected unauthorized error")
	case "forbidden":
		assert.True(t, domain.IsForbiddenError(err), "expected forbidden error")
	case "conflict":
		assert.True(t, domain.IsConflictError(err), "expected conflict error")
	default:
		t.Fatalf("unknown error type: %s", expectedType)
	}
}

// MockMediaRepository provides mock implementation for unit tests
type MockMediaRepository struct {
	inquiries   map[string]*MediaInquiry
	auditEvents []MockAuditEvent
	failures    map[string]error
}

type MockAuditEvent struct {
	EntityType    domain.EntityType
	EntityID      string
	OperationType domain.AuditEventType
	UserID        string
	Before        interface{}
	After         interface{}
}

func NewMockMediaRepository() MediaRepositoryInterface {
	return &MockMediaRepository{
		inquiries:   make(map[string]*MediaInquiry),
		auditEvents: make([]MockAuditEvent, 0),
		failures:    make(map[string]error),
	}
}

// SetFailure sets a mock failure for specific operations
func (m *MockMediaRepository) SetFailure(operation string, err error) {
	m.failures[operation] = err
}

// GetAuditEvents returns all mock audit events
func (m *MockMediaRepository) GetAuditEvents() []MockAuditEvent {
	return m.auditEvents
}

// Repository interface methods
func (m *MockMediaRepository) SaveInquiry(ctx context.Context, inquiry *MediaInquiry) error {
	if err := m.failures["SaveInquiry"]; err != nil {
		return err
	}
	m.inquiries[inquiry.InquiryID] = inquiry
	return nil
}

func (m *MockMediaRepository) GetInquiry(ctx context.Context, inquiryID string) (*MediaInquiry, error) {
	if err := m.failures["GetInquiry"]; err != nil {
		return nil, err
	}
	inquiry, exists := m.inquiries[inquiryID]
	if !exists {
		return nil, domain.NewNotFoundError("media inquiry", inquiryID)
	}
	return inquiry, nil
}

func (m *MockMediaRepository) DeleteInquiry(ctx context.Context, inquiryID string, userID string) error {
	if err := m.failures["DeleteInquiry"]; err != nil {
		return err
	}
	inquiry, exists := m.inquiries[inquiryID]
	if !exists {
		return domain.NewNotFoundError("media inquiry", inquiryID)
	}
	inquiry.IsDeleted = true
	inquiry.DeletedAt = &[]time.Time{time.Now()}[0]
	inquiry.UpdatedBy = userID
	return nil
}

func (m *MockMediaRepository) ListInquiries(ctx context.Context, filters InquiryFilters) ([]*MediaInquiry, error) {
	if err := m.failures["ListInquiries"]; err != nil {
		return nil, err
	}
	
	var result []*MediaInquiry
	for _, inquiry := range m.inquiries {
		if !inquiry.IsDeleted {
			result = append(result, inquiry)
		}
	}
	return result, nil
}

func (m *MockMediaRepository) PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error {
	if err := m.failures["PublishAuditEvent"]; err != nil {
		return err
	}
	
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

// Helper function to create test media inquiry
func createTestMediaInquiry(inquiryID, outlet, contactName, userID string) *MediaInquiry {
	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)
	return &MediaInquiry{
		InquiryID:   inquiryID,
		Status:      InquiryStatusNew,
		Priority:    InquiryPriorityMedium,
		Urgency:     InquiryUrgencyMedium,
		Outlet:      outlet,
		ContactName: contactName,
		Title:       "Senior Reporter",
		Email:       "reporter@newsoutlet.com",
		Phone:       "5551234567",
		MediaType:   &[]MediaType{MediaTypeDigital}[0],
		Deadline:    &[]time.Time{tomorrow}[0],
		Subject:     "Request for interview regarding recent clinic developments and patient care improvements.",
		Source:      "website",
		CreatedAt:   now,
		UpdatedAt:   now,
		CreatedBy:   userID,
		UpdatedBy:   userID,
		IsDeleted:   false,
	}
}

// Helper function to get pointer to time for yesterday (high urgency)
func getYesterday() *time.Time {
	yesterday := time.Now().Add(-24 * time.Hour)
	return &yesterday
}

// Helper function to get pointer to time for tomorrow (medium urgency)
func getTomorrow() *time.Time {
	tomorrow := time.Now().Add(24 * time.Hour)
	return &tomorrow
}

// Helper function to get pointer to time for next week (low urgency)
func getNextWeek() *time.Time {
	nextWeek := time.Now().Add(7 * 24 * time.Hour)
	return &nextWeek
}

// Admin Test Functions

func TestMediaService_AdminCreateInquiry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockMediaRepository)
		userID    string
		request   AdminCreateInquiryRequest
		wantError bool
		errorType string
	}{
		{
			name:      "successfully create digital media inquiry with deadline",
			setupFunc: func(repo *MockMediaRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			request: AdminCreateInquiryRequest{
				Outlet:      "TechNews Daily",
				ContactName: "Jane Reporter",
				Title:       "Senior Health Correspondent",
				Email:       "jane@technews.com",
				Phone:       "5551234567",
				MediaType:   &[]string{string(MediaTypeDigital)}[0],
				Deadline:    getTomorrow(),
				Subject:     "Request for interview about recent medical breakthroughs and patient care innovations.",
			},
			wantError: false,
		},
		{
			name:      "successfully create television inquiry with urgent deadline",
			setupFunc: func(repo *MockMediaRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			request: AdminCreateInquiryRequest{
				Outlet:      "Health News Network",
				ContactName: "Bob Anchor",
				Title:       "News Anchor",
				Email:       "bob@healthnews.tv",
				Phone:       "5559876543",
				MediaType:   &[]string{string(MediaTypeTelevision)}[0],
				Deadline:    &[]time.Time{time.Now().Add(12 * time.Hour)}[0], // Less than 1 day - should be high urgency
				Subject:     "Urgent request for live interview regarding breaking health news and clinic response.",
			},
			wantError: false,
		},
		{
			name:      "successfully create medical journal inquiry without deadline",
			setupFunc: func(repo *MockMediaRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			request: AdminCreateInquiryRequest{
				Outlet:      "Medical Research Quarterly",
				ContactName: "Sarah Editor",
				Title:       "Chief Medical Editor",
				Email:       "editor@medresearch.com",
				Phone:       "5555678901",
				MediaType:   &[]string{string(MediaTypeMedicalJournal)}[0],
				Subject:     "Request for research data and expert commentary for upcoming medical journal article.",
			},
			wantError: false,
		},
		{
			name:      "return validation error for empty outlet",
			setupFunc: func(repo *MockMediaRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			request: AdminCreateInquiryRequest{
				Outlet:      "",
				ContactName: "Jane Reporter",
				Title:       "Reporter",
				Email:       "jane@news.com",
				Phone:       "5551234567",
				Subject:     "Request for interview about clinic developments and patient care.",
			},
			wantError: true,
			errorType: "validation",
		},
		{
			name:      "return validation error for empty phone number",
			setupFunc: func(repo *MockMediaRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			request: AdminCreateInquiryRequest{
				Outlet:      "News Channel",
				ContactName: "Jane Reporter",
				Title:       "Reporter",
				Email:       "jane@news.com",
				Phone:       "",
				Subject:     "Request for interview about clinic developments and patient care.",
			},
			wantError: true,
			errorType: "validation",
		},
		{
			name:      "return validation error for short subject",
			setupFunc: func(repo *MockMediaRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			request: AdminCreateInquiryRequest{
				Outlet:      "News Channel",
				ContactName: "Jane Reporter",
				Title:       "Reporter",
				Email:       "jane@news.com",
				Phone:       "5551234567",
				Subject:     "Short subject",
			},
			wantError: true,
			errorType: "validation",
		},
		{
			name:      "return validation error for invalid media type",
			setupFunc: func(repo *MockMediaRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			request: AdminCreateInquiryRequest{
				Outlet:      "News Channel",
				ContactName: "Jane Reporter",
				Title:       "Reporter",
				Email:       "jane@news.com",
				Phone:       "5551234567",
				MediaType:   &[]string{"invalid-media-type"}[0],
				Subject:     "Request for interview about clinic developments and patient care.",
			},
			wantError: true,
			errorType: "validation",
		},
		{
			name:      "return unauthorized error for non-admin user",
			setupFunc: func(repo *MockMediaRepository) {},
			userID:    "regular-user-id",
			request: AdminCreateInquiryRequest{
				Outlet:      "News Channel",
				ContactName: "Jane Reporter",
				Title:       "Reporter",
				Email:       "jane@news.com",
				Phone:       "5551234567",
				Subject:     "Request for interview about clinic developments and patient care.",
			},
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockMediaRepository()
			mockRepo := repo.(*MockMediaRepository)
			tt.setupFunc(mockRepo)
			
			service := NewMediaService(repo)
			
			inquiry, err := service.AdminCreateInquiry(ctx, tt.request, tt.userID)
			
			if tt.wantError {
				require.Error(t, err)
				assertErrorType(t, err, tt.errorType)
				assert.Nil(t, inquiry)
			} else {
				require.NoError(t, err)
				require.NotNil(t, inquiry)
				assert.Equal(t, tt.request.Outlet, inquiry.Outlet)
				assert.Equal(t, tt.request.ContactName, inquiry.ContactName)
				assert.Equal(t, InquiryStatusNew, inquiry.Status)
				assert.Equal(t, InquiryPriorityMedium, inquiry.Priority)
				
				// Test urgency calculation based on deadline
				if tt.request.Deadline != nil {
					timeUntilDeadline := tt.request.Deadline.Sub(time.Now())
					if timeUntilDeadline <= 24*time.Hour {
						assert.Equal(t, InquiryUrgencyHigh, inquiry.Urgency)
					} else if timeUntilDeadline <= 72*time.Hour {
						assert.Equal(t, InquiryUrgencyMedium, inquiry.Urgency)
					} else {
						assert.Equal(t, InquiryUrgencyLow, inquiry.Urgency)
					}
				} else {
					assert.Equal(t, InquiryUrgencyMedium, inquiry.Urgency)
				}
			}
		})
	}
}

func TestMediaService_AdminUpdateInquiry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockMediaRepository)
		userID    string
		inquiryID string
		request   AdminUpdateInquiryRequest
		wantError bool
		errorType string
	}{
		{
			name: "successfully update inquiry with valid data",
			setupFunc: func(repo *MockMediaRepository) {
				inquiry := createTestMediaInquiry("550e8400-e29b-41d4-a716-446655440001", "Original News", "John Reporter", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.inquiries[inquiry.InquiryID] = inquiry
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			request: AdminUpdateInquiryRequest{
				Outlet:      &[]string{"Updated News Network"}[0],
				ContactName: &[]string{"Jane Doe"}[0],
				Title:       &[]string{"Chief Editor"}[0],
				MediaType:   &[]string{string(MediaTypePrint)}[0],
				Deadline:    getNextWeek(),
				Subject:     &[]string{"Updated request for comprehensive interview about medical facility developments."}[0],
			},
			wantError: false,
		},
		{
			name:      "return not found error for non-existent inquiry",
			setupFunc: func(repo *MockMediaRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440999",
			request: AdminUpdateInquiryRequest{
				Outlet:      &[]string{"Updated News"}[0],
				ContactName: &[]string{"Jane Doe"}[0],
				Subject:     &[]string{"Updated request for interview about medical developments and patient care."}[0],
			},
			wantError: true,
			errorType: "not_found",
		},
		{
			name:      "return unauthorized error for non-admin user",
			setupFunc: func(repo *MockMediaRepository) {},
			userID:    "regular-user-id",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			request: AdminUpdateInquiryRequest{
				Outlet:      &[]string{"Updated News"}[0],
				ContactName: &[]string{"Jane Doe"}[0],
				Subject:     &[]string{"Updated request for interview about medical developments and patient care."}[0],
			},
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockMediaRepository()
			mockRepo := repo.(*MockMediaRepository)
			tt.setupFunc(mockRepo)
			
			service := NewMediaService(repo)
			
			inquiry, err := service.AdminUpdateInquiry(ctx, tt.inquiryID, tt.request, tt.userID)
			
			if tt.wantError {
				require.Error(t, err)
				assertErrorType(t, err, tt.errorType)
				assert.Nil(t, inquiry)
			} else {
				require.NoError(t, err)
				require.NotNil(t, inquiry)
				if tt.request.Outlet != nil {
					assert.Equal(t, *tt.request.Outlet, inquiry.Outlet)
				}
				if tt.request.ContactName != nil {
					assert.Equal(t, *tt.request.ContactName, inquiry.ContactName)
				}
			}
		})
	}
}

func TestMediaService_AdminDeleteInquiry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockMediaRepository)
		userID    string
		inquiryID string
		wantError bool
		errorType string
	}{
		{
			name: "successfully delete inquiry",
			setupFunc: func(repo *MockMediaRepository) {
				inquiry := createTestMediaInquiry("550e8400-e29b-41d4-a716-446655440001", "News Channel", "John Reporter", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.inquiries[inquiry.InquiryID] = inquiry
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: false,
		},
		{
			name:      "return not found error for non-existent inquiry",
			setupFunc: func(repo *MockMediaRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440999",
			wantError: true,
			errorType: "not_found",
		},
		{
			name:      "return unauthorized error for non-admin user",
			setupFunc: func(repo *MockMediaRepository) {},
			userID:    "regular-user-id",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockMediaRepository()
			mockRepo := repo.(*MockMediaRepository)
			tt.setupFunc(mockRepo)
			
			service := NewMediaService(repo)
			
			err := service.AdminDeleteInquiry(ctx, tt.inquiryID, tt.userID)
			
			if tt.wantError {
				require.Error(t, err)
				assertErrorType(t, err, tt.errorType)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMediaService_AdminAcknowledgeInquiry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockMediaRepository)
		userID    string
		inquiryID string
		wantError bool
		errorType string
	}{
		{
			name: "successfully acknowledge new inquiry",
			setupFunc: func(repo *MockMediaRepository) {
				inquiry := createTestMediaInquiry("550e8400-e29b-41d4-a716-446655440001", "News Channel", "John Reporter", "admin-550e8400-e29b-41d4-a716-446655440003")
				inquiry.Status = InquiryStatusNew
				repo.inquiries[inquiry.InquiryID] = inquiry
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: false,
		},
		{
			name: "return validation error for already acknowledged inquiry",
			setupFunc: func(repo *MockMediaRepository) {
				inquiry := createTestMediaInquiry("550e8400-e29b-41d4-a716-446655440001", "News Channel", "John Reporter", "admin-550e8400-e29b-41d4-a716-446655440003")
				inquiry.Status = InquiryStatusAcknowledged
				repo.inquiries[inquiry.InquiryID] = inquiry
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: true,
			errorType: "validation",
		},
		{
			name:      "return unauthorized error for non-admin user",
			setupFunc: func(repo *MockMediaRepository) {},
			userID:    "regular-user-id",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockMediaRepository()
			mockRepo := repo.(*MockMediaRepository)
			tt.setupFunc(mockRepo)
			
			service := NewMediaService(repo)
			
			inquiry, err := service.AdminAcknowledgeInquiry(ctx, tt.inquiryID, tt.userID)
			
			if tt.wantError {
				require.Error(t, err)
				assertErrorType(t, err, tt.errorType)
				assert.Nil(t, inquiry)
			} else {
				require.NoError(t, err)
				require.NotNil(t, inquiry)
				assert.Equal(t, InquiryStatusAcknowledged, inquiry.Status)
			}
		})
	}
}

func TestMediaService_AdminResolveInquiry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockMediaRepository)
		userID    string
		inquiryID string
		wantError bool
		errorType string
	}{
		{
			name: "successfully resolve in progress inquiry",
			setupFunc: func(repo *MockMediaRepository) {
				inquiry := createTestMediaInquiry("550e8400-e29b-41d4-a716-446655440001", "News Channel", "John Reporter", "admin-550e8400-e29b-41d4-a716-446655440003")
				inquiry.Status = InquiryStatusInProgress
				repo.inquiries[inquiry.InquiryID] = inquiry
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: false,
		},
		{
			name: "return validation error for new inquiry",
			setupFunc: func(repo *MockMediaRepository) {
				inquiry := createTestMediaInquiry("550e8400-e29b-41d4-a716-446655440001", "News Channel", "John Reporter", "admin-550e8400-e29b-41d4-a716-446655440003")
				inquiry.Status = InquiryStatusNew
				repo.inquiries[inquiry.InquiryID] = inquiry
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: true,
			errorType: "validation",
		},
		{
			name:      "return unauthorized error for non-admin user",
			setupFunc: func(repo *MockMediaRepository) {},
			userID:    "regular-user-id",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockMediaRepository()
			mockRepo := repo.(*MockMediaRepository)
			tt.setupFunc(mockRepo)
			
			service := NewMediaService(repo)
			
			inquiry, err := service.AdminResolveInquiry(ctx, tt.inquiryID, tt.userID)
			
			if tt.wantError {
				require.Error(t, err)
				assertErrorType(t, err, tt.errorType)
				assert.Nil(t, inquiry)
			} else {
				require.NoError(t, err)
				require.NotNil(t, inquiry)
				assert.Equal(t, InquiryStatusResolved, inquiry.Status)
			}
		})
	}
}

func TestMediaService_AdminCloseInquiry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockMediaRepository)
		userID    string
		inquiryID string
		wantError bool
		errorType string
	}{
		{
			name: "successfully close resolved inquiry",
			setupFunc: func(repo *MockMediaRepository) {
				inquiry := createTestMediaInquiry("550e8400-e29b-41d4-a716-446655440001", "News Channel", "John Reporter", "admin-550e8400-e29b-41d4-a716-446655440003")
				inquiry.Status = InquiryStatusResolved
				repo.inquiries[inquiry.InquiryID] = inquiry
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: false,
		},
		{
			name: "successfully close any status inquiry (emergency closure)",
			setupFunc: func(repo *MockMediaRepository) {
				inquiry := createTestMediaInquiry("550e8400-e29b-41d4-a716-446655440001", "News Channel", "John Reporter", "admin-550e8400-e29b-41d4-a716-446655440003")
				inquiry.Status = InquiryStatusNew
				repo.inquiries[inquiry.InquiryID] = inquiry
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: false,
		},
		{
			name:      "return unauthorized error for non-admin user",
			setupFunc: func(repo *MockMediaRepository) {},
			userID:    "regular-user-id",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockMediaRepository()
			mockRepo := repo.(*MockMediaRepository)
			tt.setupFunc(mockRepo)
			
			service := NewMediaService(repo)
			
			inquiry, err := service.AdminCloseInquiry(ctx, tt.inquiryID, tt.userID)
			
			if tt.wantError {
				require.Error(t, err)
				assertErrorType(t, err, tt.errorType)
				assert.Nil(t, inquiry)
			} else {
				require.NoError(t, err)
				require.NotNil(t, inquiry)
				assert.Equal(t, InquiryStatusClosed, inquiry.Status)
			}
		})
	}
}

func TestMediaService_AdminSetPriority(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockMediaRepository)
		userID    string
		inquiryID string
		priority  InquiryPriority
		wantError bool
		errorType string
	}{
		{
			name: "successfully set priority to high",
			setupFunc: func(repo *MockMediaRepository) {
				inquiry := createTestMediaInquiry("550e8400-e29b-41d4-a716-446655440001", "News Channel", "John Reporter", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.inquiries[inquiry.InquiryID] = inquiry
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			priority:  InquiryPriorityHigh,
			wantError: false,
		},
		{
			name: "successfully set priority to urgent",
			setupFunc: func(repo *MockMediaRepository) {
				inquiry := createTestMediaInquiry("550e8400-e29b-41d4-a716-446655440001", "News Channel", "John Reporter", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.inquiries[inquiry.InquiryID] = inquiry
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			priority:  InquiryPriorityUrgent,
			wantError: false,
		},
		{
			name:      "return unauthorized error for non-admin user",
			setupFunc: func(repo *MockMediaRepository) {},
			userID:    "regular-user-id",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			priority:  InquiryPriorityHigh,
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockMediaRepository()
			mockRepo := repo.(*MockMediaRepository)
			tt.setupFunc(mockRepo)
			
			service := NewMediaService(repo)
			
			inquiry, err := service.AdminSetPriority(ctx, tt.inquiryID, tt.priority, tt.userID)
			
			if tt.wantError {
				require.Error(t, err)
				assertErrorType(t, err, tt.errorType)
				assert.Nil(t, inquiry)
			} else {
				require.NoError(t, err)
				require.NotNil(t, inquiry)
				assert.Equal(t, tt.priority, inquiry.Priority)
			}
		})
	}
}

func TestMediaService_AdminListInquiries(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockMediaRepository)
		userID    string
		filters   InquiryFilters
		wantError bool
		errorType string
		wantCount int
	}{
		{
			name: "successfully list inquiries",
			setupFunc: func(repo *MockMediaRepository) {
				inquiry1 := createTestMediaInquiry("550e8400-e29b-41d4-a716-446655440001", "News Channel", "John Reporter", "admin-550e8400-e29b-41d4-a716-446655440003")
				inquiry2 := createTestMediaInquiry("550e8400-e29b-41d4-a716-446655440002", "Radio Station", "Jane Host", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.inquiries[inquiry1.InquiryID] = inquiry1
				repo.inquiries[inquiry2.InquiryID] = inquiry2
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			filters:   InquiryFilters{},
			wantError: false,
			wantCount: 2,
		},
		{
			name:      "return unauthorized error for non-admin user",
			setupFunc: func(repo *MockMediaRepository) {},
			userID:    "regular-user-id",
			filters:   InquiryFilters{},
			wantError: true,
			errorType: "unauthorized",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockMediaRepository()
			mockRepo := repo.(*MockMediaRepository)
			tt.setupFunc(mockRepo)
			
			service := NewMediaService(repo)
			
			inquiries, err := service.AdminListInquiries(ctx, tt.filters, tt.userID)
			
			if tt.wantError {
				require.Error(t, err)
				assertErrorType(t, err, tt.errorType)
				assert.Nil(t, inquiries)
			} else {
				require.NoError(t, err)
				assert.Len(t, inquiries, tt.wantCount)
			}
		})
	}
}

func TestMediaService_AdminGetInquiry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockMediaRepository)
		userID    string
		inquiryID string
		wantError bool
		errorType string
	}{
		{
			name: "successfully get inquiry",
			setupFunc: func(repo *MockMediaRepository) {
				inquiry := createTestMediaInquiry("550e8400-e29b-41d4-a716-446655440001", "News Channel", "John Reporter", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.inquiries[inquiry.InquiryID] = inquiry
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: false,
		},
		{
			name:      "return not found error for non-existent inquiry",
			setupFunc: func(repo *MockMediaRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440999",
			wantError: true,
			errorType: "not_found",
		},
		{
			name:      "return unauthorized error for non-admin user",
			setupFunc: func(repo *MockMediaRepository) {},
			userID:    "regular-user-id",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockMediaRepository()
			mockRepo := repo.(*MockMediaRepository)
			tt.setupFunc(mockRepo)
			
			service := NewMediaService(repo)
			
			inquiry, err := service.AdminGetInquiry(ctx, tt.inquiryID, tt.userID)
			
			if tt.wantError {
				require.Error(t, err)
				assertErrorType(t, err, tt.errorType)
				assert.Nil(t, inquiry)
			} else {
				require.NoError(t, err)
				require.NotNil(t, inquiry)
				assert.Equal(t, tt.inquiryID, inquiry.InquiryID)
			}
		})
	}
}

// Test deadline-based urgency calculation logic specifically
func TestMediaService_DeadlineUrgencyCalculation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name            string
		deadline        *time.Time
		expectedUrgency InquiryUrgency
		description     string
	}{
		{
			name:            "deadline in 12 hours should be high urgency",
			deadline:        &[]time.Time{time.Now().Add(12 * time.Hour)}[0],
			expectedUrgency: InquiryUrgencyHigh,
			description:     "Less than 1 day",
		},
		{
			name:            "deadline in 2 days should be medium urgency",
			deadline:        &[]time.Time{time.Now().Add(48 * time.Hour)}[0],
			expectedUrgency: InquiryUrgencyMedium,
			description:     "Between 1-3 days",
		},
		{
			name:            "deadline in 5 days should be low urgency",
			deadline:        &[]time.Time{time.Now().Add(5 * 24 * time.Hour)}[0],
			expectedUrgency: InquiryUrgencyLow,
			description:     "More than 3 days",
		},
		{
			name:            "no deadline should default to medium urgency",
			deadline:        nil,
			expectedUrgency: InquiryUrgencyMedium,
			description:     "No deadline provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockMediaRepository()
			service := NewMediaService(repo)

			request := AdminCreateInquiryRequest{
				Outlet:      "Test News",
				ContactName: "Test Reporter",
				Title:       "Reporter",
				Email:       "test@news.com",
				Phone:       "5551234567",
				Deadline:    tt.deadline,
				Subject:     "Test request for interview about medical facility developments and patient care.",
			}

			inquiry, err := service.AdminCreateInquiry(ctx, request, "admin-550e8400-e29b-41d4-a716-446655440003")

			require.NoError(t, err)
			require.NotNil(t, inquiry)
			assert.Equal(t, tt.expectedUrgency, inquiry.Urgency, "Urgency calculation failed: %s", tt.description)
		})
	}
}