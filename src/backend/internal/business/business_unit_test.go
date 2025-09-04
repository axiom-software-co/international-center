package business

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

// MockBusinessRepository provides mock implementation for unit tests
type MockBusinessRepository struct {
	inquiries   map[string]*BusinessInquiry
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

func NewMockBusinessRepository() BusinessRepositoryInterface {
	return &MockBusinessRepository{
		inquiries:   make(map[string]*BusinessInquiry),
		auditEvents: make([]MockAuditEvent, 0),
		failures:    make(map[string]error),
	}
}

// SetFailure sets a mock failure for specific operations
func (m *MockBusinessRepository) SetFailure(operation string, err error) {
	m.failures[operation] = err
}

// GetAuditEvents returns all mock audit events
func (m *MockBusinessRepository) GetAuditEvents() []MockAuditEvent {
	return m.auditEvents
}

// Repository interface methods
func (m *MockBusinessRepository) SaveInquiry(ctx context.Context, inquiry *BusinessInquiry) error {
	if err := m.failures["SaveInquiry"]; err != nil {
		return err
	}
	m.inquiries[inquiry.InquiryID] = inquiry
	return nil
}

func (m *MockBusinessRepository) GetInquiry(ctx context.Context, inquiryID string) (*BusinessInquiry, error) {
	if err := m.failures["GetInquiry"]; err != nil {
		return nil, err
	}
	inquiry, exists := m.inquiries[inquiryID]
	if !exists {
		return nil, domain.NewNotFoundError("business inquiry", inquiryID)
	}
	return inquiry, nil
}

func (m *MockBusinessRepository) DeleteInquiry(ctx context.Context, inquiryID string, userID string) error {
	if err := m.failures["DeleteInquiry"]; err != nil {
		return err
	}
	inquiry, exists := m.inquiries[inquiryID]
	if !exists {
		return domain.NewNotFoundError("business inquiry", inquiryID)
	}
	inquiry.IsDeleted = true
	inquiry.DeletedAt = &[]time.Time{time.Now()}[0]
	inquiry.UpdatedBy = userID
	return nil
}

func (m *MockBusinessRepository) ListInquiries(ctx context.Context, filters InquiryFilters) ([]*BusinessInquiry, error) {
	if err := m.failures["ListInquiries"]; err != nil {
		return nil, err
	}
	
	var result []*BusinessInquiry
	for _, inquiry := range m.inquiries {
		if !inquiry.IsDeleted {
			result = append(result, inquiry)
		}
	}
	return result, nil
}

func (m *MockBusinessRepository) PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error {
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

// Helper function to create test business inquiry
func createTestBusinessInquiry(inquiryID, organizationName, contactName, userID string) *BusinessInquiry {
	now := time.Now()
	return &BusinessInquiry{
		InquiryID:        inquiryID,
		Status:           InquiryStatusNew,
		Priority:         InquiryPriorityMedium,
		OrganizationName: organizationName,
		ContactName:      contactName,
		Title:            "Director",
		Email:            "contact@organization.com",
		Phone:            &[]string{"5551234567"}[0],
		Industry:         &[]string{"Healthcare"}[0],
		InquiryType:      InquiryTypePartnership,
		Message:          "We are interested in establishing a strategic partnership for healthcare innovation.",
		Source:           "website",
		CreatedAt:        now,
		UpdatedAt:        now,
		CreatedBy:        userID,
		UpdatedBy:        userID,
		IsDeleted:        false,
	}
}

// Admin Test Functions

func TestBusinessService_AdminCreateInquiry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockBusinessRepository)
		userID    string
		request   AdminCreateInquiryRequest
		wantError bool
		errorType string
	}{
		{
			name:      "successfully create inquiry with valid data",
			setupFunc: func(repo *MockBusinessRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			request: AdminCreateInquiryRequest{
				OrganizationName: "Healthcare Partners Inc",
				ContactName:      "John Smith",
				Title:            "Director of Partnerships",
				Email:            "john.smith@healthcarepartners.com",
				Phone:            &[]string{"5551234567"}[0],
				Industry:         &[]string{"Healthcare"}[0],
				InquiryType:      string(InquiryTypePartnership),
				Message:          "We are interested in establishing a strategic partnership for healthcare innovation and research collaboration.",
			},
			wantError: false,
		},
		{
			name:      "return validation error for empty organization name",
			setupFunc: func(repo *MockBusinessRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			request: AdminCreateInquiryRequest{
				OrganizationName: "",
				ContactName:      "John Smith",
				Title:            "Director",
				Email:            "john@example.com",
				InquiryType:      string(InquiryTypePartnership),
				Message:          "We are interested in partnership opportunities for healthcare collaboration.",
			},
			wantError: true,
			errorType: "validation",
		},
		{
			name:      "return validation error for short message",
			setupFunc: func(repo *MockBusinessRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			request: AdminCreateInquiryRequest{
				OrganizationName: "Test Organization",
				ContactName:      "John Smith",
				Title:            "Director",
				Email:            "john@example.com",
				InquiryType:      string(InquiryTypePartnership),
				Message:          "Short message",
			},
			wantError: true,
			errorType: "validation",
		},
		{
			name:      "return unauthorized error for non-admin user",
			setupFunc: func(repo *MockBusinessRepository) {},
			userID:    "regular-user-id",
			request: AdminCreateInquiryRequest{
				OrganizationName: "Test Organization",
				ContactName:      "John Smith",
				Title:            "Director",
				Email:            "john@example.com",
				InquiryType:      string(InquiryTypePartnership),
				Message:          "We are interested in partnership opportunities for healthcare collaboration.",
			},
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockBusinessRepository()
			mockRepo := repo.(*MockBusinessRepository)
			tt.setupFunc(mockRepo)
			
			service := NewBusinessService(repo)
			
			inquiry, err := service.AdminCreateInquiry(ctx, tt.request, tt.userID)
			
			if tt.wantError {
				require.Error(t, err)
				assertErrorType(t, err, tt.errorType)
				assert.Nil(t, inquiry)
			} else {
				require.NoError(t, err)
				require.NotNil(t, inquiry)
				assert.Equal(t, tt.request.OrganizationName, inquiry.OrganizationName)
				assert.Equal(t, InquiryStatusNew, inquiry.Status)
				assert.Equal(t, InquiryPriorityMedium, inquiry.Priority)
			}
		})
	}
}

func TestBusinessService_AdminUpdateInquiry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockBusinessRepository)
		userID    string
		inquiryID string
		request   AdminUpdateInquiryRequest
		wantError bool
		errorType string
	}{
		{
			name: "successfully update inquiry with valid data",
			setupFunc: func(repo *MockBusinessRepository) {
				inquiry := createTestBusinessInquiry("550e8400-e29b-41d4-a716-446655440001", "Test Organization", "John Smith", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.inquiries[inquiry.InquiryID] = inquiry
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			request: AdminUpdateInquiryRequest{
				OrganizationName: &[]string{"Updated Organization"}[0],
				ContactName:      &[]string{"Jane Doe"}[0],
				Title:            &[]string{"Updated Title"}[0],
				Email:            &[]string{"jane@updated.com"}[0],
				InquiryType:      &[]string{string(InquiryTypeLicensing)}[0],
				Message:          &[]string{"Updated message about licensing opportunities for healthcare technology."}[0],
			},
			wantError: false,
		},
		{
			name:      "return not found error for non-existent inquiry",
			setupFunc: func(repo *MockBusinessRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440999",
			request: AdminUpdateInquiryRequest{
				OrganizationName: &[]string{"Updated Organization"}[0],
				ContactName:      &[]string{"Jane Doe"}[0],
				Title:            &[]string{"Updated Title"}[0],
				Email:            &[]string{"jane@updated.com"}[0],
				InquiryType:      &[]string{string(InquiryTypePartnership)}[0],
				Message:          &[]string{"Updated message about partnership opportunities."}[0],
			},
			wantError: true,
			errorType: "not_found",
		},
		{
			name:      "return unauthorized error for non-admin user",
			setupFunc: func(repo *MockBusinessRepository) {},
			userID:    "regular-user-id",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			request: AdminUpdateInquiryRequest{
				OrganizationName: &[]string{"Updated Organization"}[0],
				ContactName:      &[]string{"Jane Doe"}[0],
				Title:            &[]string{"Updated Title"}[0],
				Email:            &[]string{"jane@updated.com"}[0],
				InquiryType:      &[]string{string(InquiryTypePartnership)}[0],
				Message:          &[]string{"Updated message about partnership opportunities."}[0],
			},
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockBusinessRepository()
			mockRepo := repo.(*MockBusinessRepository)
			tt.setupFunc(mockRepo)
			
			service := NewBusinessService(repo)
			
			inquiry, err := service.AdminUpdateInquiry(ctx, tt.inquiryID, tt.request, tt.userID)
			
			if tt.wantError {
				require.Error(t, err)
				assertErrorType(t, err, tt.errorType)
				assert.Nil(t, inquiry)
			} else {
				require.NoError(t, err)
				require.NotNil(t, inquiry)
				assert.Equal(t, *tt.request.OrganizationName, inquiry.OrganizationName)
				assert.Equal(t, *tt.request.ContactName, inquiry.ContactName)
			}
		})
	}
}

func TestBusinessService_AdminDeleteInquiry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockBusinessRepository)
		userID    string
		inquiryID string
		wantError bool
		errorType string
	}{
		{
			name: "successfully delete inquiry",
			setupFunc: func(repo *MockBusinessRepository) {
				inquiry := createTestBusinessInquiry("550e8400-e29b-41d4-a716-446655440001", "Test Organization", "John Smith", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.inquiries[inquiry.InquiryID] = inquiry
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: false,
		},
		{
			name:      "return not found error for non-existent inquiry",
			setupFunc: func(repo *MockBusinessRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440999",
			wantError: true,
			errorType: "not_found",
		},
		{
			name:      "return unauthorized error for non-admin user",
			setupFunc: func(repo *MockBusinessRepository) {},
			userID:    "regular-user-id",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockBusinessRepository()
			mockRepo := repo.(*MockBusinessRepository)
			tt.setupFunc(mockRepo)
			
			service := NewBusinessService(repo)
			
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

func TestBusinessService_AdminAcknowledgeInquiry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockBusinessRepository)
		userID    string
		inquiryID string
		wantError bool
		errorType string
	}{
		{
			name: "successfully acknowledge new inquiry",
			setupFunc: func(repo *MockBusinessRepository) {
				inquiry := createTestBusinessInquiry("550e8400-e29b-41d4-a716-446655440001", "Test Organization", "John Smith", "admin-550e8400-e29b-41d4-a716-446655440003")
				inquiry.Status = InquiryStatusNew
				repo.inquiries[inquiry.InquiryID] = inquiry
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: false,
		},
		{
			name: "return validation error for already acknowledged inquiry",
			setupFunc: func(repo *MockBusinessRepository) {
				inquiry := createTestBusinessInquiry("550e8400-e29b-41d4-a716-446655440001", "Test Organization", "John Smith", "admin-550e8400-e29b-41d4-a716-446655440003")
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
			setupFunc: func(repo *MockBusinessRepository) {},
			userID:    "regular-user-id",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockBusinessRepository()
			mockRepo := repo.(*MockBusinessRepository)
			tt.setupFunc(mockRepo)
			
			service := NewBusinessService(repo)
			
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

func TestBusinessService_AdminResolveInquiry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockBusinessRepository)
		userID    string
		inquiryID string
		wantError bool
		errorType string
	}{
		{
			name: "successfully resolve in progress inquiry",
			setupFunc: func(repo *MockBusinessRepository) {
				inquiry := createTestBusinessInquiry("550e8400-e29b-41d4-a716-446655440001", "Test Organization", "John Smith", "admin-550e8400-e29b-41d4-a716-446655440003")
				inquiry.Status = InquiryStatusInProgress
				repo.inquiries[inquiry.InquiryID] = inquiry
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: false,
		},
		{
			name: "return validation error for new inquiry",
			setupFunc: func(repo *MockBusinessRepository) {
				inquiry := createTestBusinessInquiry("550e8400-e29b-41d4-a716-446655440001", "Test Organization", "John Smith", "admin-550e8400-e29b-41d4-a716-446655440003")
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
			setupFunc: func(repo *MockBusinessRepository) {},
			userID:    "regular-user-id",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockBusinessRepository()
			mockRepo := repo.(*MockBusinessRepository)
			tt.setupFunc(mockRepo)
			
			service := NewBusinessService(repo)
			
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

func TestBusinessService_AdminCloseInquiry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockBusinessRepository)
		userID    string
		inquiryID string
		wantError bool
		errorType string
	}{
		{
			name: "successfully close resolved inquiry",
			setupFunc: func(repo *MockBusinessRepository) {
				inquiry := createTestBusinessInquiry("550e8400-e29b-41d4-a716-446655440001", "Test Organization", "John Smith", "admin-550e8400-e29b-41d4-a716-446655440003")
				inquiry.Status = InquiryStatusResolved
				repo.inquiries[inquiry.InquiryID] = inquiry
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: false,
		},
		{
			name: "successfully close any status inquiry (emergency closure)",
			setupFunc: func(repo *MockBusinessRepository) {
				inquiry := createTestBusinessInquiry("550e8400-e29b-41d4-a716-446655440001", "Test Organization", "John Smith", "admin-550e8400-e29b-41d4-a716-446655440003")
				inquiry.Status = InquiryStatusNew
				repo.inquiries[inquiry.InquiryID] = inquiry
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: false,
		},
		{
			name:      "return unauthorized error for non-admin user",
			setupFunc: func(repo *MockBusinessRepository) {},
			userID:    "regular-user-id",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockBusinessRepository()
			mockRepo := repo.(*MockBusinessRepository)
			tt.setupFunc(mockRepo)
			
			service := NewBusinessService(repo)
			
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

func TestBusinessService_AdminSetPriority(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockBusinessRepository)
		userID    string
		inquiryID string
		priority  InquiryPriority
		wantError bool
		errorType string
	}{
		{
			name: "successfully set priority to high",
			setupFunc: func(repo *MockBusinessRepository) {
				inquiry := createTestBusinessInquiry("550e8400-e29b-41d4-a716-446655440001", "Test Organization", "John Smith", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.inquiries[inquiry.InquiryID] = inquiry
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			priority:  InquiryPriorityHigh,
			wantError: false,
		},
		{
			name: "successfully set priority to urgent",
			setupFunc: func(repo *MockBusinessRepository) {
				inquiry := createTestBusinessInquiry("550e8400-e29b-41d4-a716-446655440001", "Test Organization", "John Smith", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.inquiries[inquiry.InquiryID] = inquiry
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			priority:  InquiryPriorityUrgent,
			wantError: false,
		},
		{
			name:      "return unauthorized error for non-admin user",
			setupFunc: func(repo *MockBusinessRepository) {},
			userID:    "regular-user-id",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			priority:  InquiryPriorityHigh,
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockBusinessRepository()
			mockRepo := repo.(*MockBusinessRepository)
			tt.setupFunc(mockRepo)
			
			service := NewBusinessService(repo)
			
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

func TestBusinessService_AdminListInquiries(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockBusinessRepository)
		userID    string
		filters   InquiryFilters
		wantError bool
		errorType string
		wantCount int
	}{
		{
			name: "successfully list inquiries",
			setupFunc: func(repo *MockBusinessRepository) {
				inquiry1 := createTestBusinessInquiry("550e8400-e29b-41d4-a716-446655440001", "Test Organization 1", "John Smith", "admin-550e8400-e29b-41d4-a716-446655440003")
				inquiry2 := createTestBusinessInquiry("550e8400-e29b-41d4-a716-446655440002", "Test Organization 2", "Jane Doe", "admin-550e8400-e29b-41d4-a716-446655440003")
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
			setupFunc: func(repo *MockBusinessRepository) {},
			userID:    "regular-user-id",
			filters:   InquiryFilters{},
			wantError: true,
			errorType: "unauthorized",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockBusinessRepository()
			mockRepo := repo.(*MockBusinessRepository)
			tt.setupFunc(mockRepo)
			
			service := NewBusinessService(repo)
			
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

func TestBusinessService_AdminGetInquiry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockBusinessRepository)
		userID    string
		inquiryID string
		wantError bool
		errorType string
	}{
		{
			name: "successfully get inquiry",
			setupFunc: func(repo *MockBusinessRepository) {
				inquiry := createTestBusinessInquiry("550e8400-e29b-41d4-a716-446655440001", "Test Organization", "John Smith", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.inquiries[inquiry.InquiryID] = inquiry
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: false,
		},
		{
			name:      "return not found error for non-existent inquiry",
			setupFunc: func(repo *MockBusinessRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440999",
			wantError: true,
			errorType: "not_found",
		},
		{
			name:      "return unauthorized error for non-admin user",
			setupFunc: func(repo *MockBusinessRepository) {},
			userID:    "regular-user-id",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockBusinessRepository()
			mockRepo := repo.(*MockBusinessRepository)
			tt.setupFunc(mockRepo)
			
			service := NewBusinessService(repo)
			
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