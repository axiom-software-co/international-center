package donations

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

// MockDonationsRepository provides mock implementation for unit tests
type MockDonationsRepository struct {
	inquiries   map[string]*DonationsInquiry
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

func NewMockDonationsRepository() DonationsRepositoryInterface {
	return &MockDonationsRepository{
		inquiries:   make(map[string]*DonationsInquiry),
		auditEvents: make([]MockAuditEvent, 0),
		failures:    make(map[string]error),
	}
}

// SetFailure sets a mock failure for specific operations
func (m *MockDonationsRepository) SetFailure(operation string, err error) {
	m.failures[operation] = err
}

// GetAuditEvents returns all mock audit events
func (m *MockDonationsRepository) GetAuditEvents() []MockAuditEvent {
	return m.auditEvents
}

// Repository interface methods
func (m *MockDonationsRepository) SaveInquiry(ctx context.Context, inquiry *DonationsInquiry) error {
	if err := m.failures["SaveInquiry"]; err != nil {
		return err
	}
	m.inquiries[inquiry.InquiryID] = inquiry
	return nil
}

func (m *MockDonationsRepository) GetInquiry(ctx context.Context, inquiryID string) (*DonationsInquiry, error) {
	if err := m.failures["GetInquiry"]; err != nil {
		return nil, err
	}
	inquiry, exists := m.inquiries[inquiryID]
	if !exists {
		return nil, domain.NewNotFoundError("donations inquiry", inquiryID)
	}
	return inquiry, nil
}

func (m *MockDonationsRepository) DeleteInquiry(ctx context.Context, inquiryID string, userID string) error {
	if err := m.failures["DeleteInquiry"]; err != nil {
		return err
	}
	inquiry, exists := m.inquiries[inquiryID]
	if !exists {
		return domain.NewNotFoundError("donations inquiry", inquiryID)
	}
	inquiry.IsDeleted = true
	inquiry.DeletedAt = &[]time.Time{time.Now()}[0]
	inquiry.UpdatedBy = userID
	return nil
}

func (m *MockDonationsRepository) ListInquiries(ctx context.Context, filters InquiryFilters) ([]*DonationsInquiry, error) {
	if err := m.failures["ListInquiries"]; err != nil {
		return nil, err
	}
	
	var result []*DonationsInquiry
	for _, inquiry := range m.inquiries {
		if !inquiry.IsDeleted {
			result = append(result, inquiry)
		}
	}
	return result, nil
}

func (m *MockDonationsRepository) PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error {
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

// Helper function to create test donations inquiry
func createTestDonationsInquiry(inquiryID, contactName, userID string) *DonationsInquiry {
	now := time.Now()
	return &DonationsInquiry{
		InquiryID:             inquiryID,
		Status:                InquiryStatusNew,
		Priority:              InquiryPriorityMedium,
		ContactName:           contactName,
		Email:                 "contact@donor.com",
		Phone:                 &[]string{"5551234567"}[0],
		Organization:          &[]string{"Donor Foundation"}[0],
		DonorType:             DonorTypeFoundation,
		InterestArea:          &[]InterestArea{InterestAreaResearchFunding}[0],
		PreferredAmountRange:  &[]AmountRange{AmountRange1000To5000}[0],
		DonationFrequency:     &[]DonationFrequency{DonationFrequencyOneTime}[0],
		Message:               "We are interested in supporting your research initiatives through targeted funding opportunities.",
		Source:                "website",
		CreatedAt:             now,
		UpdatedAt:             now,
		CreatedBy:             userID,
		UpdatedBy:             userID,
		IsDeleted:             false,
	}
}

// Admin Test Functions

func TestDonationsService_AdminCreateInquiry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockDonationsRepository)
		userID    string
		request   AdminCreateInquiryRequest
		wantError bool
		errorType string
	}{
		{
			name:      "successfully create individual donor inquiry",
			setupFunc: func(repo *MockDonationsRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			request: AdminCreateInquiryRequest{
				ContactName:          "John Smith",
				Email:                "john.smith@email.com",
				Phone:                &[]string{"5551234567"}[0],
				DonorType:            string(DonorTypeIndividual),
				InterestArea:         &[]string{string(InterestAreaPatientCare)}[0],
				PreferredAmountRange: &[]string{string(AmountRangeUnder1000)}[0],
				DonationFrequency:    &[]string{string(DonationFrequencyMonthly)}[0],
				Message:              "I would like to support patient care programs through monthly donations.",
			},
			wantError: false,
		},
		{
			name:      "successfully create corporate donor inquiry with organization",
			setupFunc: func(repo *MockDonationsRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			request: AdminCreateInquiryRequest{
				ContactName:          "Corporate Representative",
				Email:                "contact@corporation.com",
				Organization:         &[]string{"Healthcare Corporation"}[0],
				DonorType:            string(DonorTypeCorporate),
				InterestArea:         &[]string{string(InterestAreaClinicDevelopment)}[0],
				PreferredAmountRange: &[]string{string(AmountRangeOver100000)}[0],
				DonationFrequency:    &[]string{string(DonationFrequencyAnnually)}[0],
				Message:              "Our corporation is interested in supporting clinic development through substantial annual contributions.",
			},
			wantError: false,
		},
		{
			name:      "return validation error for empty contact name",
			setupFunc: func(repo *MockDonationsRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			request: AdminCreateInquiryRequest{
				ContactName: "",
				Email:       "john@email.com",
				DonorType:   string(DonorTypeIndividual),
				Message:     "I would like to support your organization through donations.",
			},
			wantError: true,
			errorType: "validation",
		},
		{
			name:      "return validation error for short message",
			setupFunc: func(repo *MockDonationsRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			request: AdminCreateInquiryRequest{
				ContactName: "John Smith",
				Email:       "john@email.com",
				DonorType:   string(DonorTypeIndividual),
				Message:     "Short message",
			},
			wantError: true,
			errorType: "validation",
		},
		{
			name:      "return validation error for corporate donor without organization",
			setupFunc: func(repo *MockDonationsRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			request: AdminCreateInquiryRequest{
				ContactName: "Corporate Rep",
				Email:       "contact@corp.com",
				DonorType:   string(DonorTypeCorporate),
				Message:     "Corporate donation inquiry without organization field provided.",
			},
			wantError: true,
			errorType: "validation",
		},
		{
			name:      "return unauthorized error for non-admin user",
			setupFunc: func(repo *MockDonationsRepository) {},
			userID:    "regular-user-id",
			request: AdminCreateInquiryRequest{
				ContactName: "John Smith",
				Email:       "john@email.com",
				DonorType:   string(DonorTypeIndividual),
				Message:     "I would like to support your organization through donations.",
			},
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockDonationsRepository()
			mockRepo := repo.(*MockDonationsRepository)
			tt.setupFunc(mockRepo)
			
			service := NewDonationsService(repo)
			
			inquiry, err := service.AdminCreateInquiry(ctx, tt.request, tt.userID)
			
			if tt.wantError {
				require.Error(t, err)
				assertErrorType(t, err, tt.errorType)
				assert.Nil(t, inquiry)
			} else {
				require.NoError(t, err)
				require.NotNil(t, inquiry)
				assert.Equal(t, tt.request.ContactName, inquiry.ContactName)
				assert.Equal(t, InquiryStatusNew, inquiry.Status)
				assert.Equal(t, InquiryPriorityMedium, inquiry.Priority)
			}
		})
	}
}

func TestDonationsService_AdminUpdateInquiry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockDonationsRepository)
		userID    string
		inquiryID string
		request   AdminUpdateInquiryRequest
		wantError bool
		errorType string
	}{
		{
			name: "successfully update inquiry with valid data",
			setupFunc: func(repo *MockDonationsRepository) {
				inquiry := createTestDonationsInquiry("550e8400-e29b-41d4-a716-446655440001", "John Smith", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.inquiries[inquiry.InquiryID] = inquiry
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			request: AdminUpdateInquiryRequest{
				ContactName:          &[]string{"Jane Doe"}[0],
				Email:                &[]string{"jane@donor.com"}[0],
				DonorType:            &[]string{string(DonorTypeIndividual)}[0],
				InterestArea:         &[]string{string(InterestAreaEquipment)}[0],
				PreferredAmountRange: &[]string{string(AmountRange5000To25000)}[0],
				Message:              &[]string{"Updated message about supporting equipment acquisition through individual donations."}[0],
			},
			wantError: false,
		},
		{
			name:      "return not found error for non-existent inquiry",
			setupFunc: func(repo *MockDonationsRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440999",
			request: AdminUpdateInquiryRequest{
				ContactName: &[]string{"Jane Doe"}[0],
				Email:       &[]string{"jane@donor.com"}[0],
				Message:     &[]string{"Updated message about supporting our organization through donations."}[0],
			},
			wantError: true,
			errorType: "not_found",
		},
		{
			name:      "return unauthorized error for non-admin user",
			setupFunc: func(repo *MockDonationsRepository) {},
			userID:    "regular-user-id",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			request: AdminUpdateInquiryRequest{
				ContactName: &[]string{"Jane Doe"}[0],
				Email:       &[]string{"jane@donor.com"}[0],
				Message:     &[]string{"Updated message about supporting our organization through donations."}[0],
			},
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockDonationsRepository()
			mockRepo := repo.(*MockDonationsRepository)
			tt.setupFunc(mockRepo)
			
			service := NewDonationsService(repo)
			
			inquiry, err := service.AdminUpdateInquiry(ctx, tt.inquiryID, tt.request, tt.userID)
			
			if tt.wantError {
				require.Error(t, err)
				assertErrorType(t, err, tt.errorType)
				assert.Nil(t, inquiry)
			} else {
				require.NoError(t, err)
				require.NotNil(t, inquiry)
				assert.Equal(t, *tt.request.ContactName, inquiry.ContactName)
				assert.Equal(t, *tt.request.Email, inquiry.Email)
			}
		})
	}
}

func TestDonationsService_AdminDeleteInquiry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockDonationsRepository)
		userID    string
		inquiryID string
		wantError bool
		errorType string
	}{
		{
			name: "successfully delete inquiry",
			setupFunc: func(repo *MockDonationsRepository) {
				inquiry := createTestDonationsInquiry("550e8400-e29b-41d4-a716-446655440001", "John Smith", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.inquiries[inquiry.InquiryID] = inquiry
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: false,
		},
		{
			name:      "return not found error for non-existent inquiry",
			setupFunc: func(repo *MockDonationsRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440999",
			wantError: true,
			errorType: "not_found",
		},
		{
			name:      "return unauthorized error for non-admin user",
			setupFunc: func(repo *MockDonationsRepository) {},
			userID:    "regular-user-id",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockDonationsRepository()
			mockRepo := repo.(*MockDonationsRepository)
			tt.setupFunc(mockRepo)
			
			service := NewDonationsService(repo)
			
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

func TestDonationsService_AdminAcknowledgeInquiry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockDonationsRepository)
		userID    string
		inquiryID string
		wantError bool
		errorType string
	}{
		{
			name: "successfully acknowledge new inquiry",
			setupFunc: func(repo *MockDonationsRepository) {
				inquiry := createTestDonationsInquiry("550e8400-e29b-41d4-a716-446655440001", "John Smith", "admin-550e8400-e29b-41d4-a716-446655440003")
				inquiry.Status = InquiryStatusNew
				repo.inquiries[inquiry.InquiryID] = inquiry
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: false,
		},
		{
			name: "return validation error for already acknowledged inquiry",
			setupFunc: func(repo *MockDonationsRepository) {
				inquiry := createTestDonationsInquiry("550e8400-e29b-41d4-a716-446655440001", "John Smith", "admin-550e8400-e29b-41d4-a716-446655440003")
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
			setupFunc: func(repo *MockDonationsRepository) {},
			userID:    "regular-user-id",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockDonationsRepository()
			mockRepo := repo.(*MockDonationsRepository)
			tt.setupFunc(mockRepo)
			
			service := NewDonationsService(repo)
			
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

func TestDonationsService_AdminResolveInquiry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockDonationsRepository)
		userID    string
		inquiryID string
		wantError bool
		errorType string
	}{
		{
			name: "successfully resolve in progress inquiry",
			setupFunc: func(repo *MockDonationsRepository) {
				inquiry := createTestDonationsInquiry("550e8400-e29b-41d4-a716-446655440001", "John Smith", "admin-550e8400-e29b-41d4-a716-446655440003")
				inquiry.Status = InquiryStatusInProgress
				repo.inquiries[inquiry.InquiryID] = inquiry
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: false,
		},
		{
			name: "return validation error for new inquiry",
			setupFunc: func(repo *MockDonationsRepository) {
				inquiry := createTestDonationsInquiry("550e8400-e29b-41d4-a716-446655440001", "John Smith", "admin-550e8400-e29b-41d4-a716-446655440003")
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
			setupFunc: func(repo *MockDonationsRepository) {},
			userID:    "regular-user-id",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockDonationsRepository()
			mockRepo := repo.(*MockDonationsRepository)
			tt.setupFunc(mockRepo)
			
			service := NewDonationsService(repo)
			
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

func TestDonationsService_AdminCloseInquiry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockDonationsRepository)
		userID    string
		inquiryID string
		wantError bool
		errorType string
	}{
		{
			name: "successfully close resolved inquiry",
			setupFunc: func(repo *MockDonationsRepository) {
				inquiry := createTestDonationsInquiry("550e8400-e29b-41d4-a716-446655440001", "John Smith", "admin-550e8400-e29b-41d4-a716-446655440003")
				inquiry.Status = InquiryStatusResolved
				repo.inquiries[inquiry.InquiryID] = inquiry
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: false,
		},
		{
			name: "successfully close any status inquiry (emergency closure)",
			setupFunc: func(repo *MockDonationsRepository) {
				inquiry := createTestDonationsInquiry("550e8400-e29b-41d4-a716-446655440001", "John Smith", "admin-550e8400-e29b-41d4-a716-446655440003")
				inquiry.Status = InquiryStatusNew
				repo.inquiries[inquiry.InquiryID] = inquiry
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: false,
		},
		{
			name:      "return unauthorized error for non-admin user",
			setupFunc: func(repo *MockDonationsRepository) {},
			userID:    "regular-user-id",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockDonationsRepository()
			mockRepo := repo.(*MockDonationsRepository)
			tt.setupFunc(mockRepo)
			
			service := NewDonationsService(repo)
			
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

func TestDonationsService_AdminSetPriority(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockDonationsRepository)
		userID    string
		inquiryID string
		priority  InquiryPriority
		wantError bool
		errorType string
	}{
		{
			name: "successfully set priority to high",
			setupFunc: func(repo *MockDonationsRepository) {
				inquiry := createTestDonationsInquiry("550e8400-e29b-41d4-a716-446655440001", "John Smith", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.inquiries[inquiry.InquiryID] = inquiry
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			priority:  InquiryPriorityHigh,
			wantError: false,
		},
		{
			name: "successfully set priority to urgent",
			setupFunc: func(repo *MockDonationsRepository) {
				inquiry := createTestDonationsInquiry("550e8400-e29b-41d4-a716-446655440001", "John Smith", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.inquiries[inquiry.InquiryID] = inquiry
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			priority:  InquiryPriorityUrgent,
			wantError: false,
		},
		{
			name:      "return unauthorized error for non-admin user",
			setupFunc: func(repo *MockDonationsRepository) {},
			userID:    "regular-user-id",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			priority:  InquiryPriorityHigh,
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockDonationsRepository()
			mockRepo := repo.(*MockDonationsRepository)
			tt.setupFunc(mockRepo)
			
			service := NewDonationsService(repo)
			
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

func TestDonationsService_AdminListInquiries(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockDonationsRepository)
		userID    string
		filters   InquiryFilters
		wantError bool
		errorType string
		wantCount int
	}{
		{
			name: "successfully list inquiries",
			setupFunc: func(repo *MockDonationsRepository) {
				inquiry1 := createTestDonationsInquiry("550e8400-e29b-41d4-a716-446655440001", "John Smith", "admin-550e8400-e29b-41d4-a716-446655440003")
				inquiry2 := createTestDonationsInquiry("550e8400-e29b-41d4-a716-446655440002", "Jane Doe", "admin-550e8400-e29b-41d4-a716-446655440003")
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
			setupFunc: func(repo *MockDonationsRepository) {},
			userID:    "regular-user-id",
			filters:   InquiryFilters{},
			wantError: true,
			errorType: "unauthorized",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockDonationsRepository()
			mockRepo := repo.(*MockDonationsRepository)
			tt.setupFunc(mockRepo)
			
			service := NewDonationsService(repo)
			
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

func TestDonationsService_AdminGetInquiry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		setupFunc func(repo *MockDonationsRepository)
		userID    string
		inquiryID string
		wantError bool
		errorType string
	}{
		{
			name: "successfully get inquiry",
			setupFunc: func(repo *MockDonationsRepository) {
				inquiry := createTestDonationsInquiry("550e8400-e29b-41d4-a716-446655440001", "John Smith", "admin-550e8400-e29b-41d4-a716-446655440003")
				repo.inquiries[inquiry.InquiryID] = inquiry
			},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: false,
		},
		{
			name:      "return not found error for non-existent inquiry",
			setupFunc: func(repo *MockDonationsRepository) {},
			userID:    "admin-550e8400-e29b-41d4-a716-446655440003",
			inquiryID: "550e8400-e29b-41d4-a716-446655440999",
			wantError: true,
			errorType: "not_found",
		},
		{
			name:      "return unauthorized error for non-admin user",
			setupFunc: func(repo *MockDonationsRepository) {},
			userID:    "regular-user-id",
			inquiryID: "550e8400-e29b-41d4-a716-446655440001",
			wantError: true,
			errorType: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockDonationsRepository()
			mockRepo := repo.(*MockDonationsRepository)
			tt.setupFunc(mockRepo)
			
			service := NewDonationsService(repo)
			
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