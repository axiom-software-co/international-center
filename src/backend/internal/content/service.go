package content

import (
	"context"
	"crypto/sha256"
	"fmt"
	"mime"
	"path/filepath"
	"strings"

	"github.com/axiom-software-co/international-center/src/internal/shared/domain"
)

// ContentService implements business logic for content operations
type ContentService struct {
	repository *ContentRepository
}

// NewContentService creates a new content service
func NewContentService(repository *ContentRepository) *ContentService {
	return &ContentService{
		repository: repository,
	}
}

// GetContent retrieves content by ID
func (s *ContentService) GetContent(ctx context.Context, contentID string, userID string) (*Content, error) {
	if contentID == "" {
		return nil, domain.NewValidationError("content ID cannot be empty")
	}

	content, err := s.repository.GetContent(ctx, contentID)
	if err != nil {
		return nil, err
	}

	// Check access permissions
	if err := s.checkContentAccess(content, userID); err != nil {
		return nil, err
	}

	// Log access for analytics
	go s.logContentAccess(context.Background(), content, userID, "view")

	return content, nil
}

// GetAllContent retrieves all content accessible to the user
func (s *ContentService) GetAllContent(ctx context.Context, userID string) ([]*Content, error) {
	allContent, err := s.repository.GetAllContent(ctx)
	if err != nil {
		return nil, err
	}

	// Filter based on access permissions
	var accessibleContent []*Content
	for _, content := range allContent {
		if s.checkContentAccess(content, userID) == nil {
			accessibleContent = append(accessibleContent, content)
		}
	}

	return accessibleContent, nil
}

// GetContentByCategory retrieves content by category
func (s *ContentService) GetContentByCategory(ctx context.Context, category ContentCategory, userID string) ([]*Content, error) {
	if !isValidContentCategory(category) {
		return nil, domain.NewValidationError("invalid content category")
	}

	allContent, err := s.repository.GetContentByCategory(ctx, category)
	if err != nil {
		return nil, err
	}

	// Filter based on access permissions
	var accessibleContent []*Content
	for _, content := range allContent {
		if s.checkContentAccess(content, userID) == nil {
			accessibleContent = append(accessibleContent, content)
		}
	}

	return accessibleContent, nil
}

// GetContentDownload retrieves content download URL
func (s *ContentService) GetContentDownload(ctx context.Context, contentID string, userID string) (string, error) {
	content, err := s.GetContent(ctx, contentID, userID)
	if err != nil {
		return "", err
	}

	// Check if content is available for download
	if content.UploadStatus != UploadStatusAvailable {
		return "", domain.NewConflictError("content is not available for download")
	}

	// Create temporary download URL (expires in 1 hour)
	downloadURL, err := s.repository.CreateContentBlobURL(ctx, content.StoragePath, 60)
	if err != nil {
		return "", domain.NewInternalError("failed to create download URL", err)
	}

	// Log download access
	go s.logContentAccess(context.Background(), content, userID, "download")

	return downloadURL, nil
}

// GetContentPreview retrieves content preview URL
func (s *ContentService) GetContentPreview(ctx context.Context, contentID string, userID string) (string, error) {
	content, err := s.GetContent(ctx, contentID, userID)
	if err != nil {
		return "", err
	}

	// Check if content supports preview
	if !s.supportsPreview(content) {
		return "", domain.NewConflictError("content type does not support preview")
	}

	// Create temporary preview URL (expires in 30 minutes)
	previewURL, err := s.repository.CreateContentBlobURL(ctx, content.StoragePath, 30)
	if err != nil {
		return "", domain.NewInternalError("failed to create preview URL", err)
	}

	// Log preview access
	go s.logContentAccess(context.Background(), content, userID, "preview")

	return previewURL, nil
}

// CreateContent creates new content from uploaded data
func (s *ContentService) CreateContent(ctx context.Context, filename string, data []byte, contentCategory ContentCategory, userID string) (*Content, error) {
	// Validate input parameters
	if err := s.validateCreateContentParams(filename, data, contentCategory, userID); err != nil {
		return nil, err
	}

	// Calculate content hash
	contentHash := s.calculateContentHash(data)

	// Detect MIME type
	mimeType := s.detectMimeType(filename, data)

	// Create content entity
	content, err := NewContent(filename, int64(len(data)), mimeType, contentHash, contentCategory, userID)
	if err != nil {
		return nil, domain.NewInternalError("failed to create content entity", err)
	}

	// Upload to blob storage
	err = s.repository.UploadContentBlob(ctx, content.StoragePath, data, mimeType)
	if err != nil {
		return nil, domain.NewInternalError("failed to upload content blob", err)
	}

	// Save content metadata to state store
	err = s.repository.SaveContent(ctx, content)
	if err != nil {
		// If save fails, attempt to cleanup blob
		go s.repository.DeleteContentBlob(context.Background(), content.StoragePath)
		return nil, domain.NewInternalError("failed to save content metadata", err)
	}

	// Publish audit event
	err = s.repository.PublishAuditEvent(ctx, domain.EntityTypeContent, content.ContentID, domain.AuditEventInsert, userID, nil, content)
	if err != nil {
		// Log error but don't fail the operation
	}

	// Trigger async processing (virus scan, thumbnail generation, etc.)
	go s.triggerAsyncProcessing(context.Background(), content)

	return content, nil
}

// UpdateContentMetadata updates content metadata
func (s *ContentService) UpdateContentMetadata(ctx context.Context, contentID string, description, altText string, tags []string, accessLevel AccessLevel, userID string) (*Content, error) {
	// Get existing content
	content, err := s.repository.GetContent(ctx, contentID)
	if err != nil {
		return nil, err
	}

	// Check permissions
	if !s.canModifyContent(content, userID) {
		return nil, domain.NewForbiddenError("insufficient permissions to modify content")
	}

	// Store original data for audit
	originalContent := *content

	// Update metadata
	if description != "" {
		err = content.SetDescription(description, userID)
		if err != nil {
			return nil, domain.NewInternalError("failed to set description", err)
		}
	}

	if altText != "" {
		err = content.SetAltText(altText, userID)
		if err != nil {
			return nil, domain.NewInternalError("failed to set alt text", err)
		}
	}

	if len(tags) > 0 {
		err = content.AssignTags(tags, userID)
		if err != nil {
			return nil, domain.NewInternalError("failed to assign tags", err)
		}
	}

	if accessLevel != "" {
		err = content.SetAccessLevel(accessLevel, userID)
		if err != nil {
			return nil, domain.NewInternalError("failed to set access level", err)
		}
	}

	// Save updated content
	err = s.repository.SaveContent(ctx, content)
	if err != nil {
		return nil, domain.NewInternalError("failed to save updated content", err)
	}

	// Publish audit event
	err = s.repository.PublishAuditEvent(ctx, domain.EntityTypeContent, content.ContentID, domain.AuditEventUpdate, userID, &originalContent, content)
	if err != nil {
		// Log error but don't fail the operation
	}

	return content, nil
}

// DeleteContent soft deletes content
func (s *ContentService) DeleteContent(ctx context.Context, contentID string, userID string) error {
	// Get existing content
	content, err := s.repository.GetContent(ctx, contentID)
	if err != nil {
		return err
	}

	// Check permissions
	if !s.canModifyContent(content, userID) {
		return domain.NewForbiddenError("insufficient permissions to delete content")
	}

	// Store original data for audit
	originalContent := *content

	// Delete content
	err = s.repository.DeleteContent(ctx, contentID, userID)
	if err != nil {
		return domain.NewInternalError("failed to delete content", err)
	}

	// Publish audit event
	err = s.repository.PublishAuditEvent(ctx, domain.EntityTypeContent, content.ContentID, domain.AuditEventDelete, userID, &originalContent, nil)
	if err != nil {
		// Log error but don't fail the operation
	}

	return nil
}

// SearchContent performs content search
func (s *ContentService) SearchContent(ctx context.Context, searchTerm string, userID string) ([]*Content, error) {
	if strings.TrimSpace(searchTerm) == "" {
		return s.GetAllContent(ctx, userID)
	}

	searchResults, err := s.repository.SearchContent(ctx, searchTerm)
	if err != nil {
		return nil, domain.NewInternalError("failed to search content", err)
	}

	// Filter based on access permissions
	var accessibleResults []*Content
	for _, content := range searchResults {
		if s.checkContentAccess(content, userID) == nil {
			accessibleResults = append(accessibleResults, content)
		}
	}

	return accessibleResults, nil
}

// Private helper methods

// validateCreateContentParams validates content creation parameters
func (s *ContentService) validateCreateContentParams(filename string, data []byte, category ContentCategory, userID string) error {
	if strings.TrimSpace(filename) == "" {
		return domain.NewValidationError("filename cannot be empty")
	}

	if len(data) == 0 {
		return domain.NewValidationError("content data cannot be empty")
	}

	if !isValidContentCategory(category) {
		return domain.NewValidationError("invalid content category")
	}

	if strings.TrimSpace(userID) == "" {
		return domain.NewValidationError("user ID cannot be empty")
	}

	// Check file size limits (example: 100MB max)
	maxFileSize := int64(100 * 1024 * 1024)
	if int64(len(data)) > maxFileSize {
		return domain.NewValidationError("file size exceeds maximum allowed size")
	}

	return nil
}

// calculateContentHash calculates SHA-256 hash of content data
func (s *ContentService) calculateContentHash(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// detectMimeType detects MIME type from filename and data
func (s *ContentService) detectMimeType(filename string, data []byte) string {
	// Try to detect from filename extension first
	mimeType := mime.TypeByExtension(filepath.Ext(filename))
	if mimeType != "" {
		return mimeType
	}

	// Fallback to detecting from content (simplified implementation)
	// In production, this would use a proper content detection library
	return "application/octet-stream"
}

// checkContentAccess checks if user can access content based on access level
func (s *ContentService) checkContentAccess(content *Content, userID string) error {
	switch content.AccessLevel {
	case AccessLevelPublic:
		return nil // Anyone can access public content
	case AccessLevelInternal:
		if userID == "" {
			return domain.NewUnauthorizedError("authentication required for internal content")
		}
		return nil // Any authenticated user can access internal content
	case AccessLevelRestricted:
		if userID == "" {
			return domain.NewUnauthorizedError("authentication required for restricted content")
		}
		// In a real implementation, this would check user roles/permissions
		// For now, only the creator can access restricted content
		if content.CreatedBy != userID {
			return domain.NewForbiddenError("insufficient permissions to access restricted content")
		}
		return nil
	default:
		return domain.NewInternalError("invalid access level", nil)
	}
}

// canModifyContent checks if user can modify content
func (s *ContentService) canModifyContent(content *Content, userID string) bool {
	if userID == "" {
		return false
	}

	// Only the creator can modify content
	// In a real implementation, this would check for admin roles too
	return content.CreatedBy == userID
}

// supportsPreview checks if content type supports preview
func (s *ContentService) supportsPreview(content *Content) bool {
	switch content.ContentCategory {
	case ContentCategoryImage, ContentCategoryDocument:
		return true
	case ContentCategoryVideo, ContentCategoryAudio:
		// Depends on format - simplified for now
		return true
	default:
		return false
	}
}

// logContentAccess logs content access for analytics
func (s *ContentService) logContentAccess(ctx context.Context, content *Content, userID, accessType string) {
	correlationID := domain.GetCorrelationID(ctx)
	
	accessLog := NewContentAccessLog(
		content.ContentID,
		userID,
		"", // Client IP would be extracted from HTTP request
		"", // User agent would be extracted from HTTP request
		accessType,
		200,
		content.FileSize,
		0, // Response time would be calculated
		correlationID,
	)

	// Save access log (don't block on this)
	_ = s.repository.SaveContentAccessLog(ctx, accessLog)
}

// triggerAsyncProcessing triggers background processing for new content
func (s *ContentService) triggerAsyncProcessing(ctx context.Context, content *Content) {
	// This would trigger async processing like:
	// - Virus scanning
	// - Thumbnail generation
	// - Metadata extraction
	// - Content optimization
	
	// For now, just mark as available after a delay
	// In production, this would be handled by separate services
	content.MarkAsAvailable("system")
	_ = s.repository.SaveContent(ctx, content)
}