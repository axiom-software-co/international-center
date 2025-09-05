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

// RED PHASE TESTS: Domain Model Standardization

func TestInquiryPriority_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		priority InquiryPriority
		want     bool
	}{
		{
			name:     "valid low priority",
			priority: InquiryPriorityLow,
			want:     true,
		},
		{
			name:     "valid medium priority",
			priority: InquiryPriorityMedium,
			want:     true,
		},
		{
			name:     "valid high priority",
			priority: InquiryPriorityHigh,
			want:     true,
		},
		{
			name:     "valid urgent priority",
			priority: InquiryPriorityUrgent,
			want:     true,
		},
		{
			name:     "invalid priority",
			priority: InquiryPriority("invalid"),
			want:     false,
		},
		{
			name:     "empty priority",
			priority: InquiryPriority(""),
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.priority.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInquiryStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status InquiryStatus
		want   bool
	}{
		{
			name:   "valid new status",
			status: InquiryStatusNew,
			want:   true,
		},
		{
			name:   "valid acknowledged status",
			status: InquiryStatusAcknowledged,
			want:   true,
		},
		{
			name:   "valid in progress status",
			status: InquiryStatusInProgress,
			want:   true,
		},
		{
			name:   "valid resolved status",
			status: InquiryStatusResolved,
			want:   true,
		},
		{
			name:   "valid closed status",
			status: InquiryStatusClosed,
			want:   true,
		},
		{
			name:   "invalid status",
			status: InquiryStatus("invalid"),
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInquiryType_IsValid(t *testing.T) {
	tests := []struct {
		name        string
		inquiryType InquiryType
		want        bool
	}{
		{
			name:        "valid partnership type",
			inquiryType: InquiryTypePartnership,
			want:        true,
		},
		{
			name:        "valid licensing type",
			inquiryType: InquiryTypeLicensing,
			want:        true,
		},
		{
			name:        "valid research type",
			inquiryType: InquiryTypeResearch,
			want:        true,
		},
		{
			name:        "valid technology type",
			inquiryType: InquiryTypeTechnology,
			want:        true,
		},
		{
			name:        "valid regulatory type",
			inquiryType: InquiryTypeRegulatory,
			want:        true,
		},
		{
			name:        "valid other type",
			inquiryType: InquiryTypeOther,
			want:        true,
		},
		{
			name:        "invalid type",
			inquiryType: InquiryType("invalid"),
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.inquiryType.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewBusinessInquiry(t *testing.T) {
	userID := "admin-550e8400-e29b-41d4-a716-446655440003"

	tests := []struct {
		name    string
		request AdminCreateInquiryRequest
		want    *BusinessInquiry
		wantErr bool
	}{
		{
			name: "successfully create business inquiry with valid data",
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
			wantErr: false,
		},
		{
			name: "return validation error for empty organization name",
			request: AdminCreateInquiryRequest{
				OrganizationName: "",
				ContactName:      "John Smith",
				Title:            "Director",
				Email:            "john@example.com",
				InquiryType:      string(InquiryTypePartnership),
				Message:          "We are interested in partnership opportunities for healthcare collaboration.",
			},
			wantErr: true,
		},
		{
			name: "return validation error for short message",
			request: AdminCreateInquiryRequest{
				OrganizationName: "Test Organization",
				ContactName:      "John Smith",
				Title:            "Director",
				Email:            "john@example.com",
				InquiryType:      string(InquiryTypePartnership),
				Message:          "Short message",
			},
			wantErr: true,
		},
		{
			name: "return validation error for invalid inquiry type",
			request: AdminCreateInquiryRequest{
				OrganizationName: "Test Organization",
				ContactName:      "John Smith",
				Title:            "Director",
				Email:            "john@example.com",
				InquiryType:      "invalid_type",
				Message:          "We are interested in partnership opportunities for healthcare collaboration.",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inquiry, err := NewBusinessInquiry(tt.request, userID)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, inquiry)
			} else {
				require.NoError(t, err)
				require.NotNil(t, inquiry)
				assert.Equal(t, tt.request.OrganizationName, inquiry.OrganizationName)
				assert.Equal(t, tt.request.ContactName, inquiry.ContactName)
				assert.Equal(t, InquiryStatusNew, inquiry.Status)
				assert.Equal(t, InquiryPriorityMedium, inquiry.Priority)
				assert.NotEmpty(t, inquiry.InquiryID)
				assert.False(t, inquiry.CreatedAt.IsZero())
				assert.Equal(t, userID, inquiry.CreatedBy)
			}
		})
	}
}

func TestBusinessInquiry_Validate(t *testing.T) {
	tests := []struct {
		name    string
		inquiry *BusinessInquiry
		wantErr bool
	}{
		{
			name: "valid business inquiry",
			inquiry: &BusinessInquiry{
				InquiryID:        "550e8400-e29b-41d4-a716-446655440001",
				OrganizationName: "Test Organization",
				ContactName:      "John Smith",
				Title:            "Director",
				Email:            "john@example.com",
				InquiryType:      InquiryTypePartnership,
				Message:          "We are interested in partnership opportunities for healthcare collaboration.",
				Status:           InquiryStatusNew,
				Priority:         InquiryPriorityMedium,
			},
			wantErr: false,
		},
		{
			name: "invalid inquiry with empty organization name",
			inquiry: &BusinessInquiry{
				InquiryID:        "550e8400-e29b-41d4-a716-446655440001",
				OrganizationName: "",
				ContactName:      "John Smith",
				Title:            "Director",
				Email:            "john@example.com",
				InquiryType:      InquiryTypePartnership,
				Message:          "We are interested in partnership opportunities.",
				Status:           InquiryStatusNew,
				Priority:         InquiryPriorityMedium,
			},
			wantErr: true,
		},
		{
			name: "invalid inquiry with invalid email",
			inquiry: &BusinessInquiry{
				InquiryID:        "550e8400-e29b-41d4-a716-446655440001",
				OrganizationName: "Test Organization",
				ContactName:      "John Smith",
				Title:            "Director",
				Email:            "invalid-email",
				InquiryType:      InquiryTypePartnership,
				Message:          "We are interested in partnership opportunities.",
				Status:           InquiryStatusNew,
				Priority:         InquiryPriorityMedium,
			},
			wantErr: true,
		},
		{
			name: "invalid inquiry with invalid priority",
			inquiry: &BusinessInquiry{
				InquiryID:        "550e8400-e29b-41d4-a716-446655440001",
				OrganizationName: "Test Organization",
				ContactName:      "John Smith",
				Title:            "Director",
				Email:            "john@example.com",
				InquiryType:      InquiryTypePartnership,
				Message:          "We are interested in partnership opportunities.",
				Status:           InquiryStatusNew,
				Priority:         InquiryPriority("invalid"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.inquiry.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBusinessInquiry_SetPriority(t *testing.T) {
	inquiry := createTestBusinessInquiry("550e8400-e29b-41d4-a716-446655440001", "Test Organization", "John Smith", "admin-550e8400-e29b-41d4-a716-446655440003")
	userID := "admin-550e8400-e29b-41d4-a716-446655440003"

	tests := []struct {
		name     string
		priority InquiryPriority
		wantErr  bool
	}{
		{
			name:     "successfully set priority to high",
			priority: InquiryPriorityHigh,
			wantErr:  false,
		},
		{
			name:     "successfully set priority to urgent",
			priority: InquiryPriorityUrgent,
			wantErr:  false,
		},
		{
			name:     "return validation error for invalid priority",
			priority: InquiryPriority("invalid"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := inquiry.SetPriority(tt.priority, userID)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.priority, inquiry.Priority)
				assert.Equal(t, userID, inquiry.UpdatedBy)
			}
		})
	}
}

func TestBusinessInquiry_Acknowledge(t *testing.T) {
	userID := "admin-550e8400-e29b-41d4-a716-446655440003"

	tests := []struct {
		name           string
		initialStatus  InquiryStatus
		wantErr        bool
		expectedStatus InquiryStatus
	}{
		{
			name:           "successfully acknowledge new inquiry",
			initialStatus:  InquiryStatusNew,
			wantErr:        false,
			expectedStatus: InquiryStatusAcknowledged,
		},
		{
			name:          "return validation error for already acknowledged inquiry",
			initialStatus: InquiryStatusAcknowledged,
			wantErr:       true,
		},
		{
			name:          "return validation error for closed inquiry",
			initialStatus: InquiryStatusClosed,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inquiry := createTestBusinessInquiry("550e8400-e29b-41d4-a716-446655440001", "Test Organization", "John Smith", userID)
			inquiry.Status = tt.initialStatus

			err := inquiry.Acknowledge(userID)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, inquiry.Status)
				assert.Equal(t, userID, inquiry.UpdatedBy)
			}
		})
	}
}

func TestBusinessInquiry_Resolve(t *testing.T) {
	userID := "admin-550e8400-e29b-41d4-a716-446655440003"

	tests := []struct {
		name           string
		initialStatus  InquiryStatus
		wantErr        bool
		expectedStatus InquiryStatus
	}{
		{
			name:           "successfully resolve acknowledged inquiry",
			initialStatus:  InquiryStatusAcknowledged,
			wantErr:        false,
			expectedStatus: InquiryStatusResolved,
		},
		{
			name:           "successfully resolve in progress inquiry",
			initialStatus:  InquiryStatusInProgress,
			wantErr:        false,
			expectedStatus: InquiryStatusResolved,
		},
		{
			name:          "return validation error for new inquiry",
			initialStatus: InquiryStatusNew,
			wantErr:       true,
		},
		{
			name:          "return validation error for already resolved inquiry",
			initialStatus: InquiryStatusResolved,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inquiry := createTestBusinessInquiry("550e8400-e29b-41d4-a716-446655440001", "Test Organization", "John Smith", userID)
			inquiry.Status = tt.initialStatus

			err := inquiry.Resolve(userID)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, inquiry.Status)
				assert.Equal(t, userID, inquiry.UpdatedBy)
			}
		})
	}
}

func TestBusinessInquiry_Close(t *testing.T) {
	userID := "admin-550e8400-e29b-41d4-a716-446655440003"

	tests := []struct {
		name           string
		initialStatus  InquiryStatus
		expectedStatus InquiryStatus
	}{
		{
			name:           "successfully close resolved inquiry",
			initialStatus:  InquiryStatusResolved,
			expectedStatus: InquiryStatusClosed,
		},
		{
			name:           "successfully close new inquiry (emergency closure)",
			initialStatus:  InquiryStatusNew,
			expectedStatus: InquiryStatusClosed,
		},
		{
			name:           "successfully close acknowledged inquiry",
			initialStatus:  InquiryStatusAcknowledged,
			expectedStatus: InquiryStatusClosed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inquiry := createTestBusinessInquiry("550e8400-e29b-41d4-a716-446655440001", "Test Organization", "John Smith", userID)
			inquiry.Status = tt.initialStatus

			err := inquiry.Close(userID)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, inquiry.Status)
			assert.Equal(t, userID, inquiry.UpdatedBy)
		})
	}
}