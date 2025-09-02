package content

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED PHASE - Content Domain Entity Tests

func TestContentEntityCreation(t *testing.T) {
	t.Run("create content with required fields", func(t *testing.T) {
		content, err := NewContent(
			"Test Article",
			"Article content body",
			"test-article",
			"article",
		)

		require.NoError(t, err, "Should create content with valid required fields")
		assert.Equal(t, "Test Article", content.Title)
		assert.Equal(t, "Article content body", content.Body)
		assert.Equal(t, "test-article", content.Slug)
		assert.Equal(t, ContentTypeArticle, content.ContentType)
		assert.Equal(t, PublishingStatusDraft, content.PublishingStatus)
		assert.False(t, content.IsDeleted)
		assert.NotZero(t, content.CreatedOn)
	})

	t.Run("validate required fields", func(t *testing.T) {
		_, err := NewContent("", "body", "slug", "article")
		assert.Error(t, err, "Should fail when title is empty")

		_, err = NewContent("Title", "", "slug", "article")
		assert.Error(t, err, "Should fail when body is empty")

		_, err = NewContent("Title", "body", "", "article")
		assert.Error(t, err, "Should fail when slug is empty")

		_, err = NewContent("Title", "body", "slug", "invalid_type")
		assert.Error(t, err, "Should fail when content type is invalid")
	})

	t.Run("validate slug format", func(t *testing.T) {
		validSlugs := []string{"test-article", "article-123", "my-health-content"}
		for _, slug := range validSlugs {
			_, err := NewContent("Title", "Body", slug, "article")
			assert.NoError(t, err, "Should accept valid slug: %s", slug)
		}

		invalidSlugs := []string{"Test Article", "article_123", "article!", ""}
		for _, slug := range invalidSlugs {
			_, err := NewContent("Title", "Body", slug, "article")
			assert.Error(t, err, "Should reject invalid slug: %s", slug)
		}
	})
}

func TestContentPublishingWorkflow(t *testing.T) {
	t.Run("draft to published transition", func(t *testing.T) {
		content, _ := NewContent("Test Article", "Body", "test-article", "article")
		
		err := content.Publish("admin-user")
		require.NoError(t, err, "Should allow publishing from draft status")
		assert.Equal(t, PublishingStatusPublished, content.PublishingStatus)
		assert.Equal(t, "admin-user", content.ModifiedBy)
		assert.NotZero(t, content.ModifiedOn)
	})

	t.Run("published to archived transition", func(t *testing.T) {
		content, _ := NewContent("Test Article", "Body", "test-article", "article")
		content.Publish("admin-user")
		
		err := content.Archive("admin-user")
		require.NoError(t, err, "Should allow archiving from published status")
		assert.Equal(t, PublishingStatusArchived, content.PublishingStatus)
	})

	t.Run("invalid state transitions", func(t *testing.T) {
		content, _ := NewContent("Test Article", "Body", "test-article", "article")
		
		err := content.Archive("admin-user")
		assert.Error(t, err, "Should not allow archiving from draft status")
	})
}

func TestContentCategoryAssignment(t *testing.T) {
	t.Run("assign valid category", func(t *testing.T) {
		content, _ := NewContent("Test Article", "Body", "test-article", "article")
		categoryID := "650e8400-e29b-41d4-a716-446655440000"
		
		err := content.AssignCategory(categoryID, "admin-user")
		require.NoError(t, err, "Should assign valid category")
		assert.Equal(t, categoryID, content.CategoryID)
	})

	t.Run("validate category ID format", func(t *testing.T) {
		content, _ := NewContent("Test Article", "Body", "test-article", "article")
		
		err := content.AssignCategory("invalid-uuid", "admin-user")
		assert.Error(t, err, "Should reject invalid UUID format")
	})
}

func TestContentTagging(t *testing.T) {
	t.Run("assign valid tags", func(t *testing.T) {
		content, _ := NewContent("Test Article", "Body", "test-article", "article")
		tags := []string{"health", "wellness", "medical"}
		
		err := content.AssignTags(tags, "admin-user")
		require.NoError(t, err, "Should assign valid tags")
		assert.Equal(t, tags, content.Tags)
	})

	t.Run("validate tag format", func(t *testing.T) {
		content, _ := NewContent("Test Article", "Body", "test-article", "article")
		
		validTags := []string{"health", "wellness-tips", "medical123"}
		err := content.AssignTags(validTags, "admin-user")
		assert.NoError(t, err, "Should accept valid tags")

		invalidTags := []string{"health care", "wellness_tips", "medical!"}
		err = content.AssignTags(invalidTags, "admin-user")
		assert.Error(t, err, "Should reject invalid tag format")
	})
}