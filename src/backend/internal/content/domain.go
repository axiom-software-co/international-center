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

// Domain types matching TABLES-CONTENT.md schema
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

// Content represents the main content entity matching TABLES-CONTENT.md
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

// ContentAccessLog represents access logging matching TABLES-CONTENT.md
type ContentAccessLog struct {
	AccessID         string     `json:"access_id"`
	ContentID        string     `json:"content_id"`
	AccessTimestamp  time.Time  `json:"access_timestamp"`
	UserID           string     `json:"user_id,omitempty"`
	ClientIP         string     `json:"client_ip"`
	UserAgent        string     `json:"user_agent"`
	AccessType       string     `json:"access_type"`
	HTTPStatusCode   int        `json:"http_status_code"`
	BytesServed      int64      `json:"bytes_served"`
	ResponseTimeMs   int        `json:"response_time_ms"`
	CorrelationID    string     `json:"correlation_id,omitempty"`
	RefererURL       string     `json:"referer_url,omitempty"`
	CacheHit         bool       `json:"cache_hit"`
	StorageBackend   string     `json:"storage_backend"`
}

// ContentVirusScan represents virus scanning matching TABLES-CONTENT.md
type ContentVirusScan struct {
	ScanID           string     `json:"scan_id"`
	ContentID        string     `json:"content_id"`
	ScanTimestamp    time.Time  `json:"scan_timestamp"`
	ScannerEngine    string     `json:"scanner_engine"`
	ScannerVersion   string     `json:"scanner_version"`
	ScanStatus       string     `json:"scan_status"`
	ThreatsDetected  []string   `json:"threats_detected"`
	ScanDurationMs   int        `json:"scan_duration_ms"`
	CreatedOn        time.Time  `json:"created_on"`
	CorrelationID    string     `json:"correlation_id,omitempty"`
}

// ContentStorageBackend represents storage backends matching TABLES-CONTENT.md
type ContentStorageBackend struct {
	BackendID                  string                 `json:"backend_id"`
	BackendName                string                 `json:"backend_name"`
	BackendType                string                 `json:"backend_type"`
	IsActive                   bool                   `json:"is_active"`
	PriorityOrder             int                    `json:"priority_order"`
	BaseURL                   string                 `json:"base_url,omitempty"`
	AccessKeyVaultReference   string                 `json:"access_key_vault_reference,omitempty"`
	ConfigurationJSON         map[string]interface{} `json:"configuration_json,omitempty"`
	LastHealthCheck           *time.Time             `json:"last_health_check,omitempty"`
	HealthStatus              string                 `json:"health_status"`
	CreatedOn                 time.Time              `json:"created_on"`
	CreatedBy                 string                 `json:"created_by,omitempty"`
	ModifiedOn                *time.Time             `json:"modified_on,omitempty"`
	ModifiedBy                string                 `json:"modified_by,omitempty"`
}

// ContentAuditEvent represents audit events for content domain
type ContentAuditEvent struct {
	AuditID        string            `json:"audit_id"`
	EntityType     string            `json:"entity_type"`
	EntityID       string            `json:"entity_id"`
	OperationType  string            `json:"operation_type"`
	AuditTimestamp time.Time         `json:"audit_timestamp"`
	UserID         string            `json:"user_id"`
	CorrelationID  string            `json:"correlation_id"`
	TraceID        string            `json:"trace_id"`
	DataSnapshot   AuditDataSnapshot `json:"data_snapshot"`
	Environment    string            `json:"environment"`
}

// AuditDataSnapshot represents before/after data in audit events
type AuditDataSnapshot struct {
	Before interface{} `json:"before"`
	After  interface{} `json:"after"`
}

// ContentProcessingQueueItem represents items in the processing queue
type ContentProcessingQueueItem struct {
	ContentID             string      `json:"content_id"`
	OriginalFilename      string      `json:"original_filename"`
	FileSize              int64       `json:"file_size"`
	ContentCategory       string      `json:"content_category"`
	UploadStatus          string      `json:"upload_status"`
	ProcessingAttempts    int         `json:"processing_attempts"`
	LastProcessedAt       *time.Time  `json:"last_processed_at,omitempty"`
	UploadCorrelationID   string      `json:"upload_correlation_id"`
	CreatedOn             time.Time   `json:"created_on"`
	QueuePosition         int         `json:"queue_position"`
	EstimatedProcessTime  int         `json:"estimated_process_time_seconds"`
}

// ContentAnalytics represents content usage analytics
type ContentAnalytics struct {
	TotalContent            int64                   `json:"total_content"`
	ContentByCategory       map[string]int64        `json:"content_by_category"`
	ContentByAccessLevel    map[string]int64        `json:"content_by_access_level"`
	UploadsByDay           map[string]int64        `json:"uploads_by_day"`
	ProcessingMetrics      ProcessingMetrics       `json:"processing_metrics"`
	AccessMetrics          AccessMetrics           `json:"access_metrics"`
	StorageMetrics         StorageMetrics          `json:"storage_metrics"`
	VirusScanningMetrics   VirusScanningMetrics    `json:"virus_scanning_metrics"`
	GeneratedAt            time.Time               `json:"generated_at"`
}

// ProcessingMetrics represents processing performance metrics
type ProcessingMetrics struct {
	AverageProcessingTime   int         `json:"average_processing_time_ms"`
	ProcessingQueue         int         `json:"processing_queue_size"`
	ProcessedToday          int64       `json:"processed_today"`
	FailedProcessing        int64       `json:"failed_processing"`
	ProcessingSuccessRate   float64     `json:"processing_success_rate"`
}

// AccessMetrics represents content access metrics
type AccessMetrics struct {
	TotalAccesses          int64       `json:"total_accesses"`
	UniqueUsers            int64       `json:"unique_users"`
	AccessesToday          int64       `json:"accesses_today"`
	TopContentByAccess     []ContentAccessStat `json:"top_content_by_access"`
	AverageResponseTime    int         `json:"average_response_time_ms"`
	CacheHitRate          float64     `json:"cache_hit_rate"`
}

// StorageMetrics represents storage usage metrics
type StorageMetrics struct {
	TotalStorageBytes     int64       `json:"total_storage_bytes"`
	StorageByBackend      map[string]int64 `json:"storage_by_backend"`
	StorageByCategory     map[string]int64 `json:"storage_by_category"`
	StorageGrowthRate     float64     `json:"storage_growth_rate_daily"`
}

// VirusScanningMetrics represents virus scanning metrics
type VirusScanningMetrics struct {
	TotalScans           int64       `json:"total_scans"`
	InfectedFiles        int64       `json:"infected_files"`
	SuspiciousFiles      int64       `json:"suspicious_files"`
	ScanFailures         int64       `json:"scan_failures"`
	AverageScanTime      int         `json:"average_scan_time_ms"`
	ScanSuccessRate      float64     `json:"scan_success_rate"`
}

// ContentAccessStat represents statistics for content access
type ContentAccessStat struct {
	ContentID          string    `json:"content_id"`
	OriginalFilename   string    `json:"original_filename"`
	AccessCount        int64     `json:"access_count"`
	UniqueUsers        int64     `json:"unique_users"`
	LastAccess         time.Time `json:"last_access"`
}

// Domain validation patterns
var (
	filenameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	hashRegex     = regexp.MustCompile(`^[a-f0-9]{64}$`)
)

// Domain factory function
func NewContent(originalFilename string, fileSize int64, mimeType string, contentHash string, contentCategory ContentCategory, userID string) (*Content, error) {
	if err := validateNewContentParams(originalFilename, fileSize, mimeType, contentHash, contentCategory); err != nil {
		return nil, err
	}

	contentID := uuid.New().String()
	correlationID := uuid.New().String()
	now := time.Now().UTC()

	// Generate storage path based on current time and content ID
	storagePath := generateStoragePath("development", "content", now, contentID, contentHash, getFileExtension(originalFilename))

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
		CreatedBy:           userID,
		IsDeleted:           false,
	}, nil
}

// Domain business logic methods
func (c *Content) MarkAsAvailable(userID string) error {
	if c.UploadStatus != UploadStatusProcessing {
		return errors.New("can only mark content as available when processing")
	}

	c.UploadStatus = UploadStatusAvailable
	c.ModifiedBy = userID
	now := time.Now().UTC()
	c.ModifiedOn = &now

	return nil
}

func (c *Content) MarkAsFailed(userID string) error {
	c.UploadStatus = UploadStatusFailed
	c.ModifiedBy = userID
	now := time.Now().UTC()
	c.ModifiedOn = &now

	return nil
}

func (c *Content) Archive(userID string) error {
	if c.UploadStatus != UploadStatusAvailable {
		return errors.New("can only archive available content")
	}

	c.UploadStatus = UploadStatusArchived
	c.ModifiedBy = userID
	now := time.Now().UTC()
	c.ModifiedOn = &now

	return nil
}

func (c *Content) SetAccessLevel(accessLevel AccessLevel, userID string) error {
	if !isValidAccessLevel(accessLevel) {
		return errors.New("invalid access level")
	}

	c.AccessLevel = accessLevel
	c.ModifiedBy = userID
	now := time.Now().UTC()
	c.ModifiedOn = &now

	return nil
}

func (c *Content) AssignTags(tags []string, userID string) error {
	// Validate and clean tags
	var cleanTags []string
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			cleanTags = append(cleanTags, tag)
		}
	}

	c.Tags = cleanTags
	c.ModifiedBy = userID
	now := time.Now().UTC()
	c.ModifiedOn = &now

	return nil
}

func (c *Content) SetDescription(description string, userID string) error {
	c.Description = description
	c.ModifiedBy = userID
	now := time.Now().UTC()
	c.ModifiedOn = &now

	return nil
}

func (c *Content) SetAltText(altText string, userID string) error {
	c.AltText = altText
	c.ModifiedBy = userID
	now := time.Now().UTC()
	c.ModifiedOn = &now

	return nil
}

func (c *Content) Delete(userID string) error {
	c.IsDeleted = true
	c.DeletedBy = userID
	now := time.Now().UTC()
	c.DeletedOn = &now

	return nil
}

func (c *Content) IncrementProcessingAttempts() {
	c.ProcessingAttempts++
	now := time.Now().UTC()
	c.LastProcessedAt = &now
}

// NewContentAccessLog creates a new access log entry
func NewContentAccessLog(contentID, userID, clientIP, userAgent, accessType string, httpStatusCode int, bytesServed int64, responseTimeMs int, correlationID string) *ContentAccessLog {
	return &ContentAccessLog{
		AccessID:        uuid.New().String(),
		ContentID:       contentID,
		AccessTimestamp: time.Now().UTC(),
		UserID:          userID,
		ClientIP:        clientIP,
		UserAgent:       userAgent,
		AccessType:      accessType,
		HTTPStatusCode:  httpStatusCode,
		BytesServed:     bytesServed,
		ResponseTimeMs:  responseTimeMs,
		CorrelationID:   correlationID,
		CacheHit:        false,
		StorageBackend:  "azure-blob",
	}
}

// NewContentVirusScan creates a new virus scan record
func NewContentVirusScan(contentID, scannerEngine, scannerVersion, scanStatus string, threatsDetected []string, scanDurationMs int, correlationID string) *ContentVirusScan {
	return &ContentVirusScan{
		ScanID:          uuid.New().String(),
		ContentID:       contentID,
		ScanTimestamp:   time.Now().UTC(),
		ScannerEngine:   scannerEngine,
		ScannerVersion:  scannerVersion,
		ScanStatus:      scanStatus,
		ThreatsDetected: threatsDetected,
		ScanDurationMs:  scanDurationMs,
		CreatedOn:       time.Now().UTC(),
		CorrelationID:   correlationID,
	}
}

// Domain validation functions
func validateNewContentParams(originalFilename string, fileSize int64, mimeType string, contentHash string, contentCategory ContentCategory) error {
	if strings.TrimSpace(originalFilename) == "" {
		return errors.New("original filename cannot be empty")
	}

	if fileSize <= 0 {
		return errors.New("file size must be greater than 0")
	}

	if strings.TrimSpace(mimeType) == "" {
		return errors.New("mime type cannot be empty")
	}

	if strings.TrimSpace(contentHash) == "" {
		return errors.New("content hash cannot be empty")
	}

	if !isValidHash(contentHash) {
		return errors.New("content hash must be a valid SHA-256 hex string")
	}

	if !isValidContentCategory(contentCategory) {
		return errors.New("invalid content category")
	}

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

func generateStoragePath(environment, domain string, timestamp time.Time, contentID, hash, extension string) string {
	year := fmt.Sprintf("%04d", timestamp.Year())
	month := fmt.Sprintf("%02d", timestamp.Month())
	return fmt.Sprintf("%s/%s/%s/%s/%s/%s.%s", environment, domain, year, month, contentID, hash, extension)
}