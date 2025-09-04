package content

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	sharedtesting "github.com/axiom-software-co/international-center/src/backend/internal/shared/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockContentRepository provides mock implementation for unit tests
type MockContentRepository struct {
	content      map[string]*Content
	auditEvents  []MockAuditEvent
	failures     map[string]error
	blobStorage  map[string][]byte
	accessLogs   []MockAccessLog
}

type MockAuditEvent struct {
	EntityType    domain.EntityType
	EntityID      string
	OperationType domain.AuditEventType
	UserID        string
	Before        interface{}
	After         interface{}
}

type MockAccessLog struct {
	ContentID  string
	UserID     string
	AccessType string
}

func NewMockContentRepository() *MockContentRepository {
	return &MockContentRepository{
		content:     make(map[string]*Content),
		auditEvents: make([]MockAuditEvent, 0),
		failures:    make(map[string]error),
		blobStorage: make(map[string][]byte),
		accessLogs:  make([]MockAccessLog, 0),
	}
}

// SetFailure sets a mock failure for specific operations
func (m *MockContentRepository) SetFailure(operation string, err error) {
	m.failures[operation] = err
}

// GetAuditEvents returns all mock audit events
func (m *MockContentRepository) GetAuditEvents() []MockAuditEvent {
	return m.auditEvents
}

// GetAccessLogs returns all mock access logs
func (m *MockContentRepository) GetAccessLogs() []MockAccessLog {
	return m.accessLogs
}

// Content repository methods
func (m *MockContentRepository) SaveContent(ctx context.Context, content *Content) error {
	if err, exists := m.failures["SaveContent"]; exists {
		return err
	}
	m.content[content.ContentID] = content
	return nil
}

func (m *MockContentRepository) GetContent(ctx context.Context, contentID string) (*Content, error) {
	if err, exists := m.failures["GetContent"]; exists {
		return nil, err
	}
	content, exists := m.content[contentID]
	if !exists || content.IsDeleted {
		return nil, domain.NewNotFoundError("content", contentID)
	}
	return content, nil
}

func (m *MockContentRepository) GetAllContent(ctx context.Context) ([]*Content, error) {
	if err, exists := m.failures["GetAllContent"]; exists {
		return nil, err
	}
	var contents []*Content
	for _, content := range m.content {
		if !content.IsDeleted {
			contents = append(contents, content)
		}
	}
	return contents, nil
}

func (m *MockContentRepository) GetContentByCategory(ctx context.Context, category ContentCategory) ([]*Content, error) {
	if err, exists := m.failures["GetContentByCategory"]; exists {
		return nil, err
	}
	var contents []*Content
	for _, content := range m.content {
		if content.ContentCategory == category && !content.IsDeleted {
			contents = append(contents, content)
		}
	}
	return contents, nil
}

func (m *MockContentRepository) SearchContent(ctx context.Context, searchTerm string) ([]*Content, error) {
	if err, exists := m.failures["SearchContent"]; exists {
		return nil, err
	}
	var contents []*Content
	for _, content := range m.content {
		if !content.IsDeleted {
			contents = append(contents, content)
		}
	}
	return contents, nil
}

func (m *MockContentRepository) DeleteContent(ctx context.Context, contentID string, userID string) error {
	if err, exists := m.failures["DeleteContent"]; exists {
		return err
	}
	content, exists := m.content[contentID]
	if !exists {
		return domain.NewNotFoundError("content", contentID)
	}
	content.Delete(userID)
	return nil
}

// Blob storage methods
func (m *MockContentRepository) UploadContentBlob(ctx context.Context, storagePath string, data []byte, contentType string) error {
	if err, exists := m.failures["UploadContentBlob"]; exists {
		return err
	}
	m.blobStorage[storagePath] = data
	return nil
}

func (m *MockContentRepository) DeleteContentBlob(ctx context.Context, storagePath string) error {
	if err, exists := m.failures["DeleteContentBlob"]; exists {
		return err
	}
	delete(m.blobStorage, storagePath)
	return nil
}

func (m *MockContentRepository) CreateContentBlobURL(ctx context.Context, storagePath string, expiryMinutes int) (string, error) {
	if err, exists := m.failures["CreateContentBlobURL"]; exists {
		return "", err
	}
	return "https://mock-blob-storage.com/download/" + storagePath, nil
}

// Audit repository methods
func (m *MockContentRepository) PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, before interface{}, after interface{}) error {
	if err, exists := m.failures["PublishAuditEvent"]; exists {
		return err
	}
	m.auditEvents = append(m.auditEvents, MockAuditEvent{
		EntityType:    entityType,
		EntityID:      entityID,
		OperationType: operationType,
		UserID:        userID,
		Before:        before,
		After:         after,
	})
	return nil
}

func (m *MockContentRepository) LogContentAccess(ctx context.Context, contentID string, userID string, accessType string) error {
	if err, exists := m.failures["LogContentAccess"]; exists {
		return err
	}
	m.accessLogs = append(m.accessLogs, MockAccessLog{
		ContentID:  contentID,
		UserID:     userID,
		AccessType: accessType,
	})
	return nil
}

// Additional repository methods to match interface

func (m *MockContentRepository) GetContentByAccessLevel(ctx context.Context, accessLevel AccessLevel) ([]*Content, error) {
	if err, exists := m.failures["GetContentByAccessLevel"]; exists {
		return nil, err
	}
	var results []*Content
	for _, content := range m.content {
		if content.AccessLevel == accessLevel {
			results = append(results, content)
		}
	}
	if results == nil {
		results = make([]*Content, 0)
	}
	return results, nil
}

func (m *MockContentRepository) DownloadContentBlob(ctx context.Context, storagePath string) ([]byte, error) {
	if err, exists := m.failures["DownloadContentBlob"]; exists {
		return nil, err
	}
	if data, exists := m.blobStorage[storagePath]; exists {
		return data, nil
	}
	return nil, fmt.Errorf("blob not found: %s", storagePath)
}

func (m *MockContentRepository) GetContentBlobMetadata(ctx context.Context, storagePath string) (map[string]string, error) {
	if err, exists := m.failures["GetContentBlobMetadata"]; exists {
		return nil, err
	}
	if _, exists := m.blobStorage[storagePath]; exists {
		return map[string]string{
			"content-type": "application/octet-stream",
			"size":         fmt.Sprintf("%d", len(m.blobStorage[storagePath])),
		}, nil
	}
	return nil, fmt.Errorf("blob not found: %s", storagePath)
}

func (m *MockContentRepository) SaveContentAccessLog(ctx context.Context, accessLog *ContentAccessLog) error {
	if err, exists := m.failures["SaveContentAccessLog"]; exists {
		return err
	}
	m.accessLogs = append(m.accessLogs, MockAccessLog{
		ContentID:  accessLog.ContentID,
		UserID:     accessLog.UserID,
		AccessType: string(accessLog.AccessType),
	})
	return nil
}

func (m *MockContentRepository) SaveContentVirusScan(ctx context.Context, virusScan *ContentVirusScan) error {
	if err, exists := m.failures["SaveContentVirusScan"]; exists {
		return err
	}
	// Mock implementation - just return success
	return nil
}

func (m *MockContentRepository) GetContentVirusScan(ctx context.Context, contentID string) ([]*ContentVirusScan, error) {
	if err, exists := m.failures["GetContentVirusScan"]; exists {
		return nil, err
	}
	// Mock implementation - return empty slice
	return make([]*ContentVirusScan, 0), nil
}

// Admin audit repository methods - these will be implemented in the Green phase

// GetContentAudit mocks getting audit trail for content
func (m *MockContentRepository) GetContentAudit(ctx context.Context, contentID string, userID string, limit int, offset int) ([]*ContentAuditEvent, error) {
	if err, exists := m.failures["GetContentAudit"]; exists {
		return nil, err
	}
	// Return mock audit events for this content
	var events []*ContentAuditEvent
	for _, auditEvent := range m.auditEvents {
		if auditEvent.EntityID == contentID && auditEvent.EntityType == domain.EntityTypeContent {
			events = append(events, &ContentAuditEvent{
				AuditID:       fmt.Sprintf("audit-%s-%d", contentID, len(events)+1),
				EntityType:    string(auditEvent.EntityType),
				EntityID:      auditEvent.EntityID,
				OperationType: string(auditEvent.OperationType),
				AuditTimestamp: time.Now().UTC().Add(-time.Duration(len(events)) * time.Hour),
				UserID:        auditEvent.UserID,
				DataSnapshot: AuditDataSnapshot{
					Before: auditEvent.Before,
					After:  auditEvent.After,
				},
				Environment: "development",
			})
		}
	}
	return events, nil
}

// GetContentProcessingQueue mocks getting processing queue
func (m *MockContentRepository) GetContentProcessingQueue(ctx context.Context, userID string, limit int, offset int) ([]*ContentProcessingQueueItem, error) {
	if err, exists := m.failures["GetContentProcessingQueue"]; exists {
		return nil, err
	}
	// Return mock processing queue items
	var queueItems []*ContentProcessingQueueItem
	position := 1
	for _, content := range m.content {
		if content.UploadStatus == UploadStatusProcessing {
			queueItems = append(queueItems, &ContentProcessingQueueItem{
				ContentID:             content.ContentID,
				OriginalFilename:      content.OriginalFilename,
				FileSize:              content.FileSize,
				ContentCategory:       string(content.ContentCategory),
				UploadStatus:          string(content.UploadStatus),
				ProcessingAttempts:    content.ProcessingAttempts,
				LastProcessedAt:       content.LastProcessedAt,
				UploadCorrelationID:   content.UploadCorrelationID,
				CreatedOn:             content.CreatedOn,
				QueuePosition:         position,
				EstimatedProcessTime:  30,
			})
			position++
		}
	}
	return queueItems, nil
}

// GetContentAnalytics mocks getting content analytics
func (m *MockContentRepository) GetContentAnalytics(ctx context.Context, userID string) (*ContentAnalytics, error) {
	if err, exists := m.failures["GetContentAnalytics"]; exists {
		return nil, err
	}
	// Return mock analytics
	return &ContentAnalytics{
		TotalContent: int64(len(m.content)),
		ContentByCategory: map[string]int64{
			"document": 5,
			"image":    3,
			"video":    1,
		},
		ContentByAccessLevel: map[string]int64{
			"public":     2,
			"internal":   5,
			"restricted": 2,
		},
		UploadsByDay: map[string]int64{
			"2024-01-01": 3,
			"2024-01-02": 5,
			"2024-01-03": 1,
		},
		ProcessingMetrics: ProcessingMetrics{
			AverageProcessingTime: 2500,
			ProcessingQueue:       2,
			ProcessedToday:        8,
			FailedProcessing:      1,
			ProcessingSuccessRate: 0.89,
		},
		AccessMetrics: AccessMetrics{
			TotalAccesses:       1250,
			UniqueUsers:         45,
			AccessesToday:       125,
			TopContentByAccess:  []ContentAccessStat{},
			AverageResponseTime: 150,
			CacheHitRate:        0.75,
		},
		StorageMetrics: StorageMetrics{
			TotalStorageBytes: 1024 * 1024 * 100, // 100MB
			StorageByBackend: map[string]int64{
				"azure-blob": 1024 * 1024 * 100,
			},
			StorageByCategory: map[string]int64{
				"document": 1024 * 1024 * 80,
				"image":    1024 * 1024 * 20,
			},
			StorageGrowthRate: 0.05,
		},
		VirusScanningMetrics: VirusScanningMetrics{
			TotalScans:      9,
			InfectedFiles:   0,
			SuspiciousFiles: 0,
			ScanFailures:    1,
			AverageScanTime: 1200,
			ScanSuccessRate: 0.89,
		},
		GeneratedAt: time.Now().UTC(),
	}, nil
}

// Test helper functions
func createTestContent(userID string) *Content {
	// Valid SHA-256 hash (64 hex characters)
	validHash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	content, err := NewContent("test.pdf", 1024, "application/pdf", validHash, ContentCategoryDocument, userID)
	if err != nil {
		panic(fmt.Sprintf("Failed to create test content: %v", err))
	}
	return content
}

func createTestContentData() []byte {
	return []byte("test content data for unit testing")
}

// Unit Tests for ContentService

func TestContentService_GetContent(t *testing.T) {
	tests := []struct {
		name           string
		contentID      string
		userID         string
		setupMock      func(*MockContentRepository)
		expectedError  string
		validateResult func(*testing.T, *Content)
	}{
		{
			name:      "successfully get public content without auth",
			contentID: "content-1",
			userID:    "",
			setupMock: func(repo *MockContentRepository) {
				content := createTestContent("creator-1")
				content.ContentID = "content-1"
				content.AccessLevel = AccessLevelPublic
				content.UploadStatus = UploadStatusAvailable
				repo.content["content-1"] = content
			},
			validateResult: func(t *testing.T, content *Content) {
				assert.Equal(t, "content-1", content.ContentID)
				assert.Equal(t, AccessLevelPublic, content.AccessLevel)
			},
		},
		{
			name:      "successfully get internal content with auth",
			contentID: "content-2",
			userID:    "user-1",
			setupMock: func(repo *MockContentRepository) {
				content := createTestContent("creator-1")
				content.ContentID = "content-2"
				content.AccessLevel = AccessLevelInternal
				content.UploadStatus = UploadStatusAvailable
				repo.content["content-2"] = content
			},
			validateResult: func(t *testing.T, content *Content) {
				assert.Equal(t, "content-2", content.ContentID)
				assert.Equal(t, AccessLevelInternal, content.AccessLevel)
			},
		},
		{
			name:      "successfully get restricted content with creator auth",
			contentID: "content-3",
			userID:    "creator-1",
			setupMock: func(repo *MockContentRepository) {
				content := createTestContent("creator-1")
				content.ContentID = "content-3"
				content.AccessLevel = AccessLevelRestricted
				content.UploadStatus = UploadStatusAvailable
				repo.content["content-3"] = content
			},
			validateResult: func(t *testing.T, content *Content) {
				assert.Equal(t, "content-3", content.ContentID)
				assert.Equal(t, AccessLevelRestricted, content.AccessLevel)
			},
		},
		{
			name:          "fail with empty content ID",
			contentID:     "",
			userID:        "user-1",
			setupMock:     func(repo *MockContentRepository) {},
			expectedError: "content ID cannot be empty",
		},
		{
			name:      "fail accessing internal content without auth",
			contentID: "content-4",
			userID:    "",
			setupMock: func(repo *MockContentRepository) {
				content := createTestContent("creator-1")
				content.ContentID = "content-4"
				content.AccessLevel = AccessLevelInternal
				repo.content["content-4"] = content
			},
			expectedError: "authentication required",
		},
		{
			name:      "fail accessing restricted content with wrong user",
			contentID: "content-5",
			userID:    "different-user",
			setupMock: func(repo *MockContentRepository) {
				content := createTestContent("creator-1")
				content.ContentID = "content-5"
				content.AccessLevel = AccessLevelRestricted
				repo.content["content-5"] = content
			},
			expectedError: "insufficient permissions",
		},
		{
			name:      "fail when content not found",
			contentID: "nonexistent",
			userID:    "user-1",
			setupMock: func(repo *MockContentRepository) {},
			expectedError: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			
			mockRepo := NewMockContentRepository()
			tt.setupMock(mockRepo)
			service := NewContentService(mockRepo)

			// Act
			result, err := service.GetContent(ctx, tt.contentID, tt.userID)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

func TestContentService_CreateContent(t *testing.T) {
	tests := []struct {
		name           string
		filename       string
		data           []byte
		contentCategory ContentCategory
		userID         string
		setupMock      func(*MockContentRepository)
		expectedError  string
		validateResult func(*testing.T, *Content, *MockContentRepository)
	}{
		{
			name:            "successfully create document content",
			filename:        "document.pdf",
			data:            createTestContentData(),
			contentCategory: ContentCategoryDocument,
			userID:          "creator-1",
			setupMock:       func(repo *MockContentRepository) {},
			validateResult: func(t *testing.T, content *Content, repo *MockContentRepository) {
				assert.Equal(t, "document.pdf", content.OriginalFilename)
				assert.Equal(t, int64(len(createTestContentData())), content.FileSize)
				assert.Equal(t, ContentCategoryDocument, content.ContentCategory)
				assert.Equal(t, "creator-1", content.CreatedBy)
				assert.Equal(t, UploadStatusProcessing, content.UploadStatus)
				assert.Equal(t, AccessLevelInternal, content.AccessLevel)
				assert.NotEmpty(t, content.ContentHash)
				
				// Verify content was saved
				savedContent, exists := repo.content[content.ContentID]
				assert.True(t, exists)
				assert.Equal(t, content.ContentID, savedContent.ContentID)
				
				// Verify blob was uploaded
				assert.Contains(t, repo.blobStorage, content.StoragePath)
				assert.Equal(t, createTestContentData(), repo.blobStorage[content.StoragePath])
				
				// Verify audit event was published
				auditEvents := repo.GetAuditEvents()
				assert.Len(t, auditEvents, 1)
				assert.Equal(t, domain.EntityTypeContent, auditEvents[0].EntityType)
				assert.Equal(t, content.ContentID, auditEvents[0].EntityID)
				assert.Equal(t, domain.AuditEventInsert, auditEvents[0].OperationType)
				assert.Equal(t, "creator-1", auditEvents[0].UserID)
			},
		},
		{
			name:            "successfully create image content",
			filename:        "image.jpg",
			data:            createTestContentData(),
			contentCategory: ContentCategoryImage,
			userID:          "creator-1",
			setupMock:       func(repo *MockContentRepository) {},
			validateResult: func(t *testing.T, content *Content, repo *MockContentRepository) {
				assert.Equal(t, "image.jpg", content.OriginalFilename)
				assert.Equal(t, ContentCategoryImage, content.ContentCategory)
			},
		},
		{
			name:            "fail with empty filename",
			filename:        "",
			data:            createTestContentData(),
			contentCategory: ContentCategoryDocument,
			userID:          "user-1",
			setupMock:       func(repo *MockContentRepository) {},
			expectedError:   "filename cannot be empty",
		},
		{
			name:            "fail with empty data",
			filename:        "document.pdf",
			data:            []byte{},
			contentCategory: ContentCategoryDocument,
			userID:          "user-1",
			setupMock:       func(repo *MockContentRepository) {},
			expectedError:   "content data cannot be empty",
		},
		{
			name:            "fail with invalid content category",
			filename:        "document.pdf",
			data:            createTestContentData(),
			contentCategory: "invalid_category",
			userID:          "user-1",
			setupMock:       func(repo *MockContentRepository) {},
			expectedError:   "invalid content category",
		},
		{
			name:            "fail with empty user ID",
			filename:        "document.pdf",
			data:            createTestContentData(),
			contentCategory: ContentCategoryDocument,
			userID:          "",
			setupMock:       func(repo *MockContentRepository) {},
			expectedError:   "user ID cannot be empty",
		},
		{
			name:            "fail when blob upload fails",
			filename:        "document.pdf",
			data:            createTestContentData(),
			contentCategory: ContentCategoryDocument,
			userID:          "user-1",
			setupMock: func(repo *MockContentRepository) {
				repo.SetFailure("UploadContentBlob", domain.NewInternalError("blob upload failed", nil))
			},
			expectedError: "failed to upload content blob",
		},
		{
			name:            "fail when metadata save fails",
			filename:        "document.pdf",
			data:            createTestContentData(),
			contentCategory: ContentCategoryDocument,
			userID:          "user-1",
			setupMock: func(repo *MockContentRepository) {
				repo.SetFailure("SaveContent", domain.NewInternalError("metadata save failed", nil))
			},
			expectedError: "failed to save content metadata",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			
			mockRepo := NewMockContentRepository()
			tt.setupMock(mockRepo)
			service := NewContentService(mockRepo)

			// Act
			result, err := service.CreateContent(ctx, tt.filename, tt.data, tt.contentCategory, tt.userID)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result, mockRepo)
				}
			}
		})
	}
}

func TestContentService_GetContentDownload(t *testing.T) {
	tests := []struct {
		name           string
		contentID      string
		userID         string
		setupMock      func(*MockContentRepository)
		expectedError  string
		validateResult func(*testing.T, string, *MockContentRepository)
	}{
		{
			name:      "successfully get download URL for available content",
			contentID: "content-1",
			userID:    "",
			setupMock: func(repo *MockContentRepository) {
				content := createTestContent("creator-1")
				content.ContentID = "content-1"
				content.AccessLevel = AccessLevelPublic
				content.UploadStatus = UploadStatusAvailable
				repo.content["content-1"] = content
			},
			validateResult: func(t *testing.T, downloadURL string, repo *MockContentRepository) {
				assert.Contains(t, downloadURL, "https://mock-blob-storage.com/download/")
				
				// Verify access was logged
				accessLogs := repo.GetAccessLogs()
				assert.Len(t, accessLogs, 1)
				assert.Equal(t, "content-1", accessLogs[0].ContentID)
				assert.Equal(t, "download", accessLogs[0].AccessType)
			},
		},
		{
			name:      "fail when content not available",
			contentID: "content-2",
			userID:    "",
			setupMock: func(repo *MockContentRepository) {
				content := createTestContent("creator-1")
				content.ContentID = "content-2"
				content.AccessLevel = AccessLevelPublic
				content.UploadStatus = UploadStatusProcessing
				repo.content["content-2"] = content
			},
			expectedError: "content is not available for download",
		},
		{
			name:      "fail when content not found",
			contentID: "nonexistent",
			userID:    "",
			setupMock: func(repo *MockContentRepository) {},
			expectedError: "not found",
		},
		{
			name:      "fail when blob URL creation fails",
			contentID: "content-3",
			userID:    "",
			setupMock: func(repo *MockContentRepository) {
				content := createTestContent("creator-1")
				content.ContentID = "content-3"
				content.AccessLevel = AccessLevelPublic
				content.UploadStatus = UploadStatusAvailable
				repo.content["content-3"] = content
				repo.SetFailure("CreateContentBlobURL", domain.NewInternalError("blob URL creation failed", nil))
			},
			expectedError: "failed to create download URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			
			mockRepo := NewMockContentRepository()
			tt.setupMock(mockRepo)
			service := NewContentService(mockRepo)

			// Act
			result, err := service.GetContentDownload(ctx, tt.contentID, tt.userID)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Empty(t, result)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result, mockRepo)
				}
			}
		})
	}
}

func TestContentService_UpdateContentMetadata(t *testing.T) {
	tests := []struct {
		name           string
		contentID      string
		description    string
		altText        string
		tags           []string
		accessLevel    AccessLevel
		userID         string
		setupMock      func(*MockContentRepository)
		expectedError  string
		validateResult func(*testing.T, *Content, *MockContentRepository)
	}{
		{
			name:        "successfully update all metadata fields",
			contentID:   "content-1",
			description: "Updated description",
			altText:     "Updated alt text",
			tags:        []string{"tag1", "tag2"},
			accessLevel: AccessLevelPublic,
			userID:      "creator-1",
			setupMock: func(repo *MockContentRepository) {
				content := createTestContent("creator-1")
				content.ContentID = "content-1"
				repo.content["content-1"] = content
			},
			validateResult: func(t *testing.T, content *Content, repo *MockContentRepository) {
				assert.Equal(t, "Updated description", content.Description)
				assert.Equal(t, "Updated alt text", content.AltText)
				assert.Equal(t, []string{"tag1", "tag2"}, content.Tags)
				assert.Equal(t, AccessLevelPublic, content.AccessLevel)
				assert.Equal(t, "creator-1", content.ModifiedBy)
				assert.NotNil(t, content.ModifiedOn)
				
				// Verify audit event was published
				auditEvents := repo.GetAuditEvents()
				assert.Len(t, auditEvents, 1)
				assert.Equal(t, domain.AuditEventUpdate, auditEvents[0].OperationType)
			},
		},
		{
			name:        "fail when user lacks permission",
			contentID:   "content-2",
			description: "Updated description",
			userID:      "different-user",
			setupMock: func(repo *MockContentRepository) {
				content := createTestContent("creator-1")
				content.ContentID = "content-2"
				repo.content["content-2"] = content
			},
			expectedError: "insufficient permissions to modify content",
		},
		{
			name:        "fail when content not found",
			contentID:   "nonexistent",
			description: "Updated description",
			userID:      "user-1",
			setupMock:   func(repo *MockContentRepository) {},
			expectedError: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			
			mockRepo := NewMockContentRepository()
			tt.setupMock(mockRepo)
			service := NewContentService(mockRepo)

			// Act
			result, err := service.UpdateContentMetadata(ctx, tt.contentID, tt.description, tt.altText, tt.tags, tt.accessLevel, tt.userID)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result, mockRepo)
				}
			}
		})
	}
}

func TestContentService_GetAllContent(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		setupMock      func(*MockContentRepository)
		expectedError  string
		validateResult func(*testing.T, []*Content)
	}{
		{
			name:   "successfully get public content for anonymous user",
			userID: "",
			setupMock: func(repo *MockContentRepository) {
				// Public content - should be accessible
				publicContent := createTestContent("creator-1")
				publicContent.ContentID = "public"
				publicContent.AccessLevel = AccessLevelPublic
				repo.content["public"] = publicContent
				
				// Internal content - should not be accessible anonymously
				internalContent := createTestContent("creator-1")
				internalContent.ContentID = "internal"
				internalContent.AccessLevel = AccessLevelInternal
				repo.content["internal"] = internalContent
			},
			validateResult: func(t *testing.T, contents []*Content) {
				assert.Len(t, contents, 1)
				assert.Equal(t, "public", contents[0].ContentID)
				assert.Equal(t, AccessLevelPublic, contents[0].AccessLevel)
			},
		},
		{
			name:   "successfully get all accessible content for authenticated user",
			userID: "creator-1",
			setupMock: func(repo *MockContentRepository) {
				// Public content - should be accessible
				publicContent := createTestContent("creator-1")
				publicContent.ContentID = "public"
				publicContent.AccessLevel = AccessLevelPublic
				repo.content["public"] = publicContent
				
				// Internal content - should be accessible
				internalContent := createTestContent("creator-1")
				internalContent.ContentID = "internal"
				internalContent.AccessLevel = AccessLevelInternal
				repo.content["internal"] = internalContent
				
				// Own restricted content - should be accessible
				restrictedContent := createTestContent("creator-1")
				restrictedContent.ContentID = "restricted"
				restrictedContent.AccessLevel = AccessLevelRestricted
				repo.content["restricted"] = restrictedContent
				
				// Other user's restricted content - should not be accessible
				otherRestrictedContent := createTestContent("other-creator")
				otherRestrictedContent.ContentID = "other-restricted"
				otherRestrictedContent.AccessLevel = AccessLevelRestricted
				repo.content["other-restricted"] = otherRestrictedContent
			},
			validateResult: func(t *testing.T, contents []*Content) {
				assert.Len(t, contents, 3)
				contentIDs := make([]string, len(contents))
				for i, content := range contents {
					contentIDs[i] = content.ContentID
				}
				assert.Contains(t, contentIDs, "public")
				assert.Contains(t, contentIDs, "internal")
				assert.Contains(t, contentIDs, "restricted")
			},
		},
		{
			name:   "return empty array when no accessible content",
			userID: "",
			setupMock: func(repo *MockContentRepository) {
				// Only internal content - not accessible anonymously
				internalContent := createTestContent("creator-1")
				internalContent.ContentID = "internal"
				internalContent.AccessLevel = AccessLevelInternal
				repo.content["internal"] = internalContent
			},
			validateResult: func(t *testing.T, contents []*Content) {
				assert.Len(t, contents, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			
			mockRepo := NewMockContentRepository()
			tt.setupMock(mockRepo)
			service := NewContentService(mockRepo)

			// Act
			result, err := service.GetAllContent(ctx, tt.userID)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

func TestContentService_Timeout(t *testing.T) {
	// Test that context timeout is respected (5 seconds for unit tests)
	ctx, cancel := sharedtesting.CreateUnitTestContext()
	defer cancel()
	
	// Verify context has 5 second timeout
	deadline, hasDeadline := ctx.Deadline()
	require.True(t, hasDeadline)
	assert.True(t, time.Until(deadline) <= 5*time.Second)
	assert.True(t, time.Until(deadline) > 4*time.Second) // Allow some margin
}

// Admin Content Audit Endpoint Unit Tests - RED PHASE (Failing Tests)

func TestContentService_GetContentAudit(t *testing.T) {
	tests := []struct {
		name           string
		contentID      string
		userID         string
		limit          int
		offset         int
		setupMock      func(*MockContentRepository)
		expectedError  string
		validateResult func(*testing.T, []*ContentAuditEvent)
	}{
		{
			name:      "successfully get content audit trail",
			contentID: "content-1",
			userID:    "admin-1",
			limit:     10,
			offset:    0,
			setupMock: func(repo *MockContentRepository) {
				// Create content with some audit events
				content := createTestContent("creator-1")
				content.ContentID = "content-1"
				repo.content["content-1"] = content
				
				// Add audit events
				repo.auditEvents = []MockAuditEvent{
					{
						EntityType:    domain.EntityTypeContent,
						EntityID:      "content-1",
						OperationType: domain.AuditEventInsert,
						UserID:        "creator-1",
						Before:        nil,
						After:         content,
					},
					{
						EntityType:    domain.EntityTypeContent,
						EntityID:      "content-1",
						OperationType: domain.AuditEventUpdate,
						UserID:        "creator-1",
						Before:        content,
						After:         content,
					},
				}
			},
			validateResult: func(t *testing.T, events []*ContentAuditEvent) {
				assert.Len(t, events, 2)
				assert.Equal(t, "content-1", events[0].EntityID)
				assert.Equal(t, "content", events[0].EntityType)
				assert.Equal(t, "creator-1", events[0].UserID)
				assert.Equal(t, "development", events[0].Environment)
			},
		},
		{
			name:          "fail with empty content ID",
			contentID:     "",
			userID:        "admin-1",
			limit:         10,
			offset:        0,
			setupMock:     func(repo *MockContentRepository) {},
			expectedError: "content ID cannot be empty",
		},
		{
			name:          "fail without admin authentication",
			contentID:     "content-1",
			userID:        "",
			limit:         10,
			offset:        0,
			setupMock:     func(repo *MockContentRepository) {},
			expectedError: "admin authentication required",
		},
		{
			name:      "return empty array for content with no audit events",
			contentID: "content-2",
			userID:    "admin-1",
			limit:     10,
			offset:    0,
			setupMock: func(repo *MockContentRepository) {
				content := createTestContent("creator-1")
				content.ContentID = "content-2"
				repo.content["content-2"] = content
			},
			validateResult: func(t *testing.T, events []*ContentAuditEvent) {
				assert.Len(t, events, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			
			mockRepo := NewMockContentRepository()
			tt.setupMock(mockRepo)
			service := NewContentService(mockRepo)

			// Act - this will fail until we implement GetContentAudit method
			result, err := service.GetContentAudit(ctx, tt.contentID, tt.userID, tt.limit, tt.offset)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

func TestContentService_GetContentProcessingQueue(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		limit          int
		offset         int
		setupMock      func(*MockContentRepository)
		expectedError  string
		validateResult func(*testing.T, []*ContentProcessingQueueItem)
	}{
		{
			name:   "successfully get processing queue",
			userID: "admin-1",
			limit:  10,
			offset: 0,
			setupMock: func(repo *MockContentRepository) {
				// Create content in processing status
				processingContent1 := createTestContent("creator-1")
				processingContent1.ContentID = "processing-1"
				processingContent1.UploadStatus = UploadStatusProcessing
				processingContent1.ProcessingAttempts = 1
				repo.content["processing-1"] = processingContent1
				
				processingContent2 := createTestContent("creator-2")
				processingContent2.ContentID = "processing-2"
				processingContent2.UploadStatus = UploadStatusProcessing
				processingContent2.ProcessingAttempts = 2
				repo.content["processing-2"] = processingContent2
				
				// Create available content (should not appear in queue)
				availableContent := createTestContent("creator-1")
				availableContent.ContentID = "available-1"
				availableContent.UploadStatus = UploadStatusAvailable
				repo.content["available-1"] = availableContent
			},
			validateResult: func(t *testing.T, items []*ContentProcessingQueueItem) {
				assert.Len(t, items, 2)
				assert.Equal(t, "processing", items[0].UploadStatus)
				assert.Greater(t, items[0].QueuePosition, 0)
				assert.Greater(t, items[0].EstimatedProcessTime, 0)
			},
		},
		{
			name:          "fail without admin authentication",
			userID:        "",
			limit:         10,
			offset:        0,
			setupMock:     func(repo *MockContentRepository) {},
			expectedError: "admin authentication required",
		},
		{
			name:   "return empty queue when no processing content",
			userID: "admin-1",
			limit:  10,
			offset: 0,
			setupMock: func(repo *MockContentRepository) {
				// Only available content
				availableContent := createTestContent("creator-1")
				availableContent.ContentID = "available-1"
				availableContent.UploadStatus = UploadStatusAvailable
				repo.content["available-1"] = availableContent
			},
			validateResult: func(t *testing.T, items []*ContentProcessingQueueItem) {
				assert.Len(t, items, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			
			mockRepo := NewMockContentRepository()
			tt.setupMock(mockRepo)
			service := NewContentService(mockRepo)

			// Act - this will fail until we implement GetContentProcessingQueue method
			result, err := service.GetContentProcessingQueue(ctx, tt.userID, tt.limit, tt.offset)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

func TestContentService_GetContentAnalytics(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		setupMock      func(*MockContentRepository)
		expectedError  string
		validateResult func(*testing.T, *ContentAnalytics)
	}{
		{
			name:   "successfully get content analytics",
			userID: "admin-1",
			setupMock: func(repo *MockContentRepository) {
				// Create various content types
				docContent := createTestContent("creator-1")
				docContent.ContentCategory = ContentCategoryDocument
				docContent.AccessLevel = AccessLevelPublic
				repo.content["doc-1"] = docContent
				
				imgContent := createTestContent("creator-2")
				imgContent.ContentCategory = ContentCategoryImage
				imgContent.AccessLevel = AccessLevelInternal
				repo.content["img-1"] = imgContent
				
				videoContent := createTestContent("creator-3")
				videoContent.ContentCategory = ContentCategoryVideo
				videoContent.AccessLevel = AccessLevelRestricted
				repo.content["video-1"] = videoContent
			},
			validateResult: func(t *testing.T, analytics *ContentAnalytics) {
				assert.Greater(t, analytics.TotalContent, int64(0))
				assert.NotEmpty(t, analytics.ContentByCategory)
				assert.NotEmpty(t, analytics.ContentByAccessLevel)
				assert.NotEmpty(t, analytics.UploadsByDay)
				assert.Greater(t, analytics.ProcessingMetrics.AverageProcessingTime, 0)
				assert.GreaterOrEqual(t, analytics.ProcessingMetrics.ProcessingSuccessRate, 0.0)
				assert.LessOrEqual(t, analytics.ProcessingMetrics.ProcessingSuccessRate, 1.0)
				assert.GreaterOrEqual(t, analytics.AccessMetrics.CacheHitRate, 0.0)
				assert.LessOrEqual(t, analytics.AccessMetrics.CacheHitRate, 1.0)
				assert.Greater(t, analytics.StorageMetrics.TotalStorageBytes, int64(0))
				assert.NotNil(t, analytics.GeneratedAt)
			},
		},
		{
			name:          "fail without admin authentication",
			userID:        "",
			setupMock:     func(repo *MockContentRepository) {},
			expectedError: "admin authentication required",
		},
		{
			name:   "successfully get analytics with no content",
			userID: "admin-1",
			setupMock: func(repo *MockContentRepository) {
				// No content in repository
			},
			validateResult: func(t *testing.T, analytics *ContentAnalytics) {
				assert.Equal(t, int64(0), analytics.TotalContent)
				assert.NotNil(t, analytics.GeneratedAt)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()
			
			mockRepo := NewMockContentRepository()
			tt.setupMock(mockRepo)
			service := NewContentService(mockRepo)

			// Act - this will fail until we implement GetContentAnalytics method
			result, err := service.GetContentAnalytics(ctx, tt.userID)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}