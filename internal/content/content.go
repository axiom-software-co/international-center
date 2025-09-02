package content

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ContentCategory string
type AccessLevel string
type UploadStatus string

const (
	ContentCategoryDocument ContentCategory = "document"
	ContentCategoryImage    ContentCategory = "image"
	ContentCategoryVideo    ContentCategory = "video"
	ContentCategoryAudio    ContentCategory = "audio"
	ContentCategoryArchive  ContentCategory = "archive"
)

const (
	AccessLevelPublic     AccessLevel = "public"
	AccessLevelInternal   AccessLevel = "internal"
	AccessLevelRestricted AccessLevel = "restricted"
)

const (
	UploadStatusProcessing UploadStatus = "processing"
	UploadStatusAvailable  UploadStatus = "available"
	UploadStatusFailed     UploadStatus = "failed"
	UploadStatusArchived   UploadStatus = "archived"
)

type Content struct {
	ContentID             string          `json:"content_id"`
	OriginalFilename      string          `json:"original_filename"`
	FileSize              int64           `json:"file_size"`
	MimeType              string          `json:"mime_type"`
	ContentHash           string          `json:"content_hash"`
	StoragePath           string          `json:"storage_path"`
	UploadStatus          UploadStatus    `json:"upload_status"`
	AltText               string          `json:"alt_text,omitempty"`
	Description           string          `json:"description,omitempty"`
	Tags                  []string        `json:"tags"`
	ContentCategory       ContentCategory `json:"content_category"`
	AccessLevel           AccessLevel     `json:"access_level"`
	UploadCorrelationID   string          `json:"upload_correlation_id"`
	ProcessingAttempts    int             `json:"processing_attempts"`
	LastProcessedAt       *time.Time      `json:"last_processed_at,omitempty"`
	CreatedOn             time.Time       `json:"created_on"`
	CreatedBy             string          `json:"created_by,omitempty"`
	ModifiedOn            *time.Time      `json:"modified_on,omitempty"`
	ModifiedBy            string          `json:"modified_by,omitempty"`
	IsDeleted             bool            `json:"is_deleted"`
	DeletedOn             *time.Time      `json:"deleted_on,omitempty"`
	DeletedBy             string          `json:"deleted_by,omitempty"`
}

var filenameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
var hashRegex = regexp.MustCompile(`^[a-f0-9]{64}$`)

func NewContent(originalFilename string, fileSize int64, mimeType string, contentHash string, contentCategory ContentCategory) (*Content, error) {
	if strings.TrimSpace(originalFilename) == "" {
		return nil, errors.New("original filename cannot be empty")
	}
	
	if fileSize <= 0 {
		return nil, errors.New("file size must be greater than 0")
	}
	
	if strings.TrimSpace(mimeType) == "" {
		return nil, errors.New("mime type cannot be empty")
	}
	
	if strings.TrimSpace(contentHash) == "" {
		return nil, errors.New("content hash cannot be empty")
	}
	
	if !isValidHash(contentHash) {
		return nil, errors.New("content hash must be a valid SHA-256 hex string")
	}
	
	if !isValidContentCategory(contentCategory) {
		return nil, errors.New("invalid content category")
	}
	
	contentID := uuid.New().String()
	correlationID := uuid.New().String()
	
	// Generate storage path
	now := time.Now()
	storagePath := fmt.Sprintf("development/content/%d/%02d/%s/%s.%s",
		now.Year(), now.Month(), contentID, contentHash, getFileExtension(originalFilename))
	
	return &Content{
		ContentID:           contentID,
		OriginalFilename:    originalFilename,
		FileSize:            fileSize,
		MimeType:            mimeType,
		ContentHash:         contentHash,
		StoragePath:         storagePath,
		UploadStatus:        UploadStatusProcessing,
		Tags:                []string{},
		ContentCategory:     contentCategory,
		AccessLevel:         AccessLevelInternal, // Default to internal
		UploadCorrelationID: correlationID,
		ProcessingAttempts:  0,
		CreatedOn:           now,
		IsDeleted:           false,
	}, nil
}

func (c *Content) MarkAsAvailable(userID string) error {
	if c.UploadStatus != UploadStatusProcessing {
		return errors.New("can only mark content as available when processing")
	}
	
	c.UploadStatus = UploadStatusAvailable
	c.ModifiedBy = userID
	now := time.Now()
	c.ModifiedOn = &now
	
	return nil
}

func (c *Content) MarkAsFailed(userID string) error {
	c.UploadStatus = UploadStatusFailed
	c.ModifiedBy = userID
	now := time.Now()
	c.ModifiedOn = &now
	
	return nil
}

func (c *Content) Archive(userID string) error {
	if c.UploadStatus != UploadStatusAvailable {
		return errors.New("can only archive available content")
	}
	
	c.UploadStatus = UploadStatusArchived
	c.ModifiedBy = userID
	now := time.Now()
	c.ModifiedOn = &now
	
	return nil
}

func (c *Content) SetAccessLevel(accessLevel AccessLevel, userID string) error {
	if !isValidAccessLevel(accessLevel) {
		return errors.New("invalid access level")
	}
	
	c.AccessLevel = accessLevel
	c.ModifiedBy = userID
	now := time.Now()
	c.ModifiedOn = &now
	
	return nil
}

func (c *Content) AssignTags(tags []string, userID string) error {
	c.Tags = tags
	c.ModifiedBy = userID
	now := time.Now()
	c.ModifiedOn = &now
	
	return nil
}

func (c *Content) SetDescription(description string, userID string) error {
	c.Description = description
	c.ModifiedBy = userID
	now := time.Now()
	c.ModifiedOn = &now
	
	return nil
}

func isValidHash(hash string) bool {
	return hashRegex.MatchString(hash)
}

func isValidContentCategory(category ContentCategory) bool {
	switch category {
	case ContentCategoryDocument, ContentCategoryImage, ContentCategoryVideo, ContentCategoryAudio, ContentCategoryArchive:
		return true
	default:
		return false
	}
}

func isValidAccessLevel(level AccessLevel) bool {
	switch level {
	case AccessLevelPublic, AccessLevelInternal, AccessLevelRestricted:
		return true
	default:
		return false
	}
}

func getFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	if ext != "" && len(ext) > 1 {
		return ext[1:] // Remove the leading dot
	}
	return "bin" // Default extension for files without extension
}