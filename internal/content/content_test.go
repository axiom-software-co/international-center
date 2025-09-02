package content

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED PHASE - Content Domain Entity Tests for File-Based Content Management

func TestContentEntityCreation(t *testing.T) {
	t.Run("create content with required fields", func(t *testing.T) {
		content, err := NewContent(
			"test-document.pdf",
			1024,
			"application/pdf",
			"abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			ContentCategoryDocument,
		)

		require.NoError(t, err, "Should create content with valid required fields")
		assert.Equal(t, "test-document.pdf", content.OriginalFilename)
		assert.Equal(t, int64(1024), content.FileSize)
		assert.Equal(t, "application/pdf", content.MimeType)
		assert.Equal(t, "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890", content.ContentHash)
		assert.Equal(t, ContentCategoryDocument, content.ContentCategory)
		assert.Equal(t, UploadStatusProcessing, content.UploadStatus)
		assert.Equal(t, AccessLevelInternal, content.AccessLevel) // Default
		assert.False(t, content.IsDeleted)
		assert.NotZero(t, content.CreatedOn)
		assert.NotEmpty(t, content.ContentID)
		assert.NotEmpty(t, content.StoragePath)
		assert.NotEmpty(t, content.UploadCorrelationID)
	})

	t.Run("validate required fields", func(t *testing.T) {
		validHash := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
		
		_, err := NewContent("", 1024, "application/pdf", validHash, ContentCategoryDocument)
		assert.Error(t, err, "Should fail when filename is empty")

		_, err = NewContent("test.pdf", 0, "application/pdf", validHash, ContentCategoryDocument)
		assert.Error(t, err, "Should fail when file size is zero")

		_, err = NewContent("test.pdf", -1, "application/pdf", validHash, ContentCategoryDocument)
		assert.Error(t, err, "Should fail when file size is negative")

		_, err = NewContent("test.pdf", 1024, "", validHash, ContentCategoryDocument)
		assert.Error(t, err, "Should fail when mime type is empty")

		_, err = NewContent("test.pdf", 1024, "application/pdf", "", ContentCategoryDocument)
		assert.Error(t, err, "Should fail when content hash is empty")

		_, err = NewContent("test.pdf", 1024, "application/pdf", "invalid-hash", ContentCategoryDocument)
		assert.Error(t, err, "Should fail when content hash is invalid format")

		_, err = NewContent("test.pdf", 1024, "application/pdf", validHash, "invalid_category")
		assert.Error(t, err, "Should fail when content category is invalid")
	})

	t.Run("validate content hash format", func(t *testing.T) {
		validHashes := []string{
			"abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		}
		for _, hash := range validHashes {
			_, err := NewContent("test.pdf", 1024, "application/pdf", hash, ContentCategoryDocument)
			assert.NoError(t, err, "Should accept valid hash: %s", hash)
		}

		invalidHashes := []string{
			"short", // Too short
			"abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890X", // Too long
			"ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890", // Uppercase
			"abcdef1234567890abcdef1234567890abcdef1234567890abcdef123456789g", // Invalid character
		}
		for _, hash := range invalidHashes {
			_, err := NewContent("test.pdf", 1024, "application/pdf", hash, ContentCategoryDocument)
			assert.Error(t, err, "Should reject invalid hash: %s", hash)
		}
	})
}

func TestContentUploadWorkflow(t *testing.T) {
	t.Run("processing to available transition", func(t *testing.T) {
		content, _ := NewContent(
			"document.pdf", 
			1024, 
			"application/pdf", 
			"abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			ContentCategoryDocument,
		)
		
		err := content.MarkAsAvailable("admin-user")
		require.NoError(t, err, "Should allow marking as available from processing status")
		assert.Equal(t, UploadStatusAvailable, content.UploadStatus)
		assert.Equal(t, "admin-user", content.ModifiedBy)
		assert.NotNil(t, content.ModifiedOn)
	})

	t.Run("available to archived transition", func(t *testing.T) {
		content, _ := NewContent(
			"document.pdf", 
			1024, 
			"application/pdf", 
			"abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			ContentCategoryDocument,
		)
		content.MarkAsAvailable("admin-user")
		
		err := content.Archive("admin-user")
		require.NoError(t, err, "Should allow archiving from available status")
		assert.Equal(t, UploadStatusArchived, content.UploadStatus)
	})

	t.Run("mark processing as failed", func(t *testing.T) {
		content, _ := NewContent(
			"document.pdf", 
			1024, 
			"application/pdf", 
			"abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			ContentCategoryDocument,
		)
		
		err := content.MarkAsFailed("admin-user")
		require.NoError(t, err, "Should allow marking as failed from processing status")
		assert.Equal(t, UploadStatusFailed, content.UploadStatus)
		assert.Equal(t, "admin-user", content.ModifiedBy)
	})

	t.Run("invalid state transitions", func(t *testing.T) {
		content, _ := NewContent(
			"document.pdf", 
			1024, 
			"application/pdf", 
			"abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			ContentCategoryDocument,
		)
		
		// Can only archive available content
		err := content.Archive("admin-user")
		assert.Error(t, err, "Should not allow archiving from processing status")

		// Mark as available first
		content.MarkAsAvailable("admin-user")
		
		// Can't mark as available when already available
		err = content.MarkAsAvailable("admin-user")
		assert.Error(t, err, "Should not allow marking as available when already available")
	})
}

func TestContentAccessLevelManagement(t *testing.T) {
	t.Run("set valid access levels", func(t *testing.T) {
		content, _ := NewContent(
			"document.pdf", 
			1024, 
			"application/pdf", 
			"abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			ContentCategoryDocument,
		)
		
		err := content.SetAccessLevel(AccessLevelPublic, "admin-user")
		require.NoError(t, err, "Should set public access level")
		assert.Equal(t, AccessLevelPublic, content.AccessLevel)
		assert.Equal(t, "admin-user", content.ModifiedBy)
		
		err = content.SetAccessLevel(AccessLevelRestricted, "admin-user")
		require.NoError(t, err, "Should set restricted access level")
		assert.Equal(t, AccessLevelRestricted, content.AccessLevel)
	})

	t.Run("reject invalid access level", func(t *testing.T) {
		content, _ := NewContent(
			"document.pdf", 
			1024, 
			"application/pdf", 
			"abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			ContentCategoryDocument,
		)
		
		err := content.SetAccessLevel("invalid_level", "admin-user")
		assert.Error(t, err, "Should reject invalid access level")
	})
}

func TestContentTagging(t *testing.T) {
	t.Run("assign tags to content", func(t *testing.T) {
		content, _ := NewContent(
			"document.pdf", 
			1024, 
			"application/pdf", 
			"abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			ContentCategoryDocument,
		)
		tags := []string{"medical", "compliance", "document"}
		
		err := content.AssignTags(tags, "admin-user")
		require.NoError(t, err, "Should assign tags to content")
		assert.Equal(t, tags, content.Tags)
		assert.Equal(t, "admin-user", content.ModifiedBy)
	})

	t.Run("set content description", func(t *testing.T) {
		content, _ := NewContent(
			"document.pdf", 
			1024, 
			"application/pdf", 
			"abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			ContentCategoryDocument,
		)
		description := "Important compliance document for medical procedures"
		
		err := content.SetDescription(description, "admin-user")
		require.NoError(t, err, "Should set content description")
		assert.Equal(t, description, content.Description)
		assert.Equal(t, "admin-user", content.ModifiedBy)
	})
}