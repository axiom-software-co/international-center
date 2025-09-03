package content

import (
	"context"
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/testing"
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
func (m *MockContentRepository) UploadContentBlob(ctx context.Context, storagePath string, data []byte, mimeType string) error {
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

func (m *MockContentRepository) CreateContentBlobURL(ctx context.Context, storagePath string, expirationMinutes int) (string, error) {
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

// Test helper functions
func createTestContent(userID string) *Content {
	content, _ := NewContent("test.pdf", 1024, "application/pdf", "testhash123", ContentCategoryDocument, userID)
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
			ctx, cancel := testing.CreateUnitTestContext()
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
			ctx, cancel := testing.CreateUnitTestContext()
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
			ctx, cancel := testing.CreateUnitTestContext()
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
			ctx, cancel := testing.CreateUnitTestContext()
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
			ctx, cancel := testing.CreateUnitTestContext()
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
	ctx, cancel := testing.CreateUnitTestContext()
	defer cancel()
	
	// Verify context has 5 second timeout
	deadline, hasDeadline := ctx.Deadline()
	require.True(t, hasDeadline)
	assert.True(t, time.Until(deadline) <= 5*time.Second)
	assert.True(t, time.Until(deadline) > 4*time.Second) // Allow some margin
}