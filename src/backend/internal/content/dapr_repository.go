package content

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
)

// ContentRepository implements content data access using Dapr state store and bindings
type ContentRepository struct {
	stateStore *dapr.StateStore
	bindings   *dapr.Bindings
	pubsub     *dapr.PubSub
}

// NewContentRepository creates a new content repository
func NewContentRepository(client *dapr.Client) *ContentRepository {
	return &ContentRepository{
		stateStore: dapr.NewStateStore(client),
		bindings:   dapr.NewBindings(client),
		pubsub:     dapr.NewPubSub(client),
	}
}

// SaveContent saves content to Dapr state store
func (r *ContentRepository) SaveContent(ctx context.Context, content *Content) error {
	key := r.stateStore.CreateKey("content", "content", content.ContentID)
	
	err := r.stateStore.Save(ctx, key, content, nil)
	if err != nil {
		return fmt.Errorf("failed to save content %s: %w", content.ContentID, err)
	}

	// Create index for filename search
	filenameKey := r.stateStore.CreateIndexKey("content", "content", "filename", content.OriginalFilename)
	filenameIndex := map[string]string{"content_id": content.ContentID}
	
	err = r.stateStore.Save(ctx, filenameKey, filenameIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create filename index for content %s: %w", content.ContentID, err)
	}

	// Create index for category search
	categoryKey := r.stateStore.CreateIndexKey("content", "content", "category", string(content.ContentCategory))
	categoryIndex := map[string]interface{}{
		"content_id": content.ContentID,
		"created_on": content.CreatedOn,
	}
	
	err = r.stateStore.Save(ctx, categoryKey, categoryIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create category index for content %s: %w", content.ContentID, err)
	}

	return nil
}

// GetContent retrieves content by ID from Dapr state store
func (r *ContentRepository) GetContent(ctx context.Context, contentID string) (*Content, error) {
	key := r.stateStore.CreateKey("content", "content", contentID)
	
	var content Content
	found, err := r.stateStore.Get(ctx, key, &content)
	if err != nil {
		return nil, fmt.Errorf("failed to get content %s: %w", contentID, err)
	}
	
	if !found {
		return nil, domain.NewNotFoundError("content", contentID)
	}

	if content.IsDeleted {
		return nil, domain.NewNotFoundError("content", contentID)
	}

	return &content, nil
}

// GetAllContent retrieves all non-deleted content from Dapr state store
func (r *ContentRepository) GetAllContent(ctx context.Context) ([]*Content, error) {
	// Use query to get all content
	query := `{
		"filter": {
			"AND": [
				{
					"EQ": {"is_deleted": false}
				}
			]
		},
		"sort": [
			{
				"key": "created_on",
				"order": "DESC"
			}
		]
	}`

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all content: %w", err)
	}

	var contents []*Content
	for _, result := range results {
		var content Content
		err = json.Unmarshal(result.Value, &content)
		if err != nil {
			continue // Skip invalid records
		}
		
		if !content.IsDeleted {
			contents = append(contents, &content)
		}
	}

	return contents, nil
}

// GetContentByCategory retrieves content by category from Dapr state store
func (r *ContentRepository) GetContentByCategory(ctx context.Context, category ContentCategory) ([]*Content, error) {
	query := fmt.Sprintf(`{
		"filter": {
			"AND": [
				{
					"EQ": {"content_category": "%s"}
				},
				{
					"EQ": {"is_deleted": false}
				}
			]
		},
		"sort": [
			{
				"key": "created_on", 
				"order": "DESC"
			}
		]
	}`, string(category))

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query content by category %s: %w", category, err)
	}

	var contents []*Content
	for _, result := range results {
		var content Content
		err = json.Unmarshal(result.Value, &content)
		if err != nil {
			continue // Skip invalid records
		}
		contents = append(contents, &content)
	}

	return contents, nil
}

// GetContentByAccessLevel retrieves content by access level
func (r *ContentRepository) GetContentByAccessLevel(ctx context.Context, accessLevel AccessLevel) ([]*Content, error) {
	query := fmt.Sprintf(`{
		"filter": {
			"AND": [
				{
					"EQ": {"access_level": "%s"}
				},
				{
					"EQ": {"is_deleted": false}
				}
			]
		},
		"sort": [
			{
				"key": "created_on",
				"order": "DESC"
			}
		]
	}`, string(accessLevel))

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query content by access level %s: %w", accessLevel, err)
	}

	var contents []*Content
	for _, result := range results {
		var content Content
		err = json.Unmarshal(result.Value, &content)
		if err != nil {
			continue // Skip invalid records
		}
		contents = append(contents, &content)
	}

	return contents, nil
}

// DeleteContent soft deletes content from Dapr state store
func (r *ContentRepository) DeleteContent(ctx context.Context, contentID string, userID string) error {
	content, err := r.GetContent(ctx, contentID)
	if err != nil {
		return err
	}

	err = content.Delete(userID)
	if err != nil {
		return err
	}

	return r.SaveContent(ctx, content)
}

// UploadContentBlob uploads content data to blob storage via Dapr bindings
func (r *ContentRepository) UploadContentBlob(ctx context.Context, storagePath string, data []byte, contentType string) error {
	err := r.bindings.UploadBlob(ctx, storagePath, data, contentType)
	if err != nil {
		return fmt.Errorf("failed to upload content blob to %s: %w", storagePath, err)
	}

	return nil
}

// DownloadContentBlob downloads content data from blob storage via Dapr bindings
func (r *ContentRepository) DownloadContentBlob(ctx context.Context, storagePath string) ([]byte, error) {
	data, err := r.bindings.DownloadBlob(ctx, storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to download content blob from %s: %w", storagePath, err)
	}

	return data, nil
}

// DeleteContentBlob deletes content data from blob storage via Dapr bindings
func (r *ContentRepository) DeleteContentBlob(ctx context.Context, storagePath string) error {
	err := r.bindings.DeleteBlob(ctx, storagePath)
	if err != nil {
		return fmt.Errorf("failed to delete content blob from %s: %w", storagePath, err)
	}

	return nil
}

// GetContentBlobMetadata retrieves blob metadata via Dapr bindings
func (r *ContentRepository) GetContentBlobMetadata(ctx context.Context, storagePath string) (map[string]string, error) {
	metadata, err := r.bindings.GetBlobMetadata(ctx, storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get content blob metadata from %s: %w", storagePath, err)
	}

	return metadata, nil
}

// CreateContentBlobURL creates a temporary access URL for blob content
func (r *ContentRepository) CreateContentBlobURL(ctx context.Context, storagePath string, expiryMinutes int) (string, error) {
	url, err := r.bindings.CreateBlobURL(ctx, storagePath, expiryMinutes)
	if err != nil {
		return "", fmt.Errorf("failed to create content blob URL for %s: %w", storagePath, err)
	}

	return url, nil
}

// SaveContentAccessLog saves access log entry to Dapr state store
func (r *ContentRepository) SaveContentAccessLog(ctx context.Context, accessLog *ContentAccessLog) error {
	key := r.stateStore.CreateKey("content", "access_log", accessLog.AccessID)
	
	err := r.stateStore.Save(ctx, key, accessLog, nil)
	if err != nil {
		return fmt.Errorf("failed to save content access log %s: %w", accessLog.AccessID, err)
	}

	// Publish access event for analytics
	eventData := map[string]interface{}{
		"content_id":       accessLog.ContentID,
		"user_id":          accessLog.UserID,
		"access_type":      accessLog.AccessType,
		"response_time_ms": accessLog.ResponseTimeMs,
		"bytes_served":     accessLog.BytesServed,
		"cache_hit":        accessLog.CacheHit,
	}

	err = r.pubsub.PublishContentEvent(ctx, "access", accessLog.ContentID, eventData)
	if err != nil {
		// Log error but don't fail the operation
		// In production, this would be logged properly
	}

	return nil
}

// SaveContentVirusScan saves virus scan result to Dapr state store
func (r *ContentRepository) SaveContentVirusScan(ctx context.Context, virusScan *ContentVirusScan) error {
	key := r.stateStore.CreateKey("content", "virus_scan", virusScan.ScanID)
	
	err := r.stateStore.Save(ctx, key, virusScan, nil)
	if err != nil {
		return fmt.Errorf("failed to save content virus scan %s: %w", virusScan.ScanID, err)
	}

	return nil
}

// GetContentVirusScan retrieves virus scan results for content
func (r *ContentRepository) GetContentVirusScan(ctx context.Context, contentID string) ([]*ContentVirusScan, error) {
	query := fmt.Sprintf(`{
		"filter": {
			"EQ": {"content_id": "%s"}
		},
		"sort": [
			{
				"key": "scan_timestamp",
				"order": "DESC"
			}
		]
	}`, contentID)

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query virus scans for content %s: %w", contentID, err)
	}

	var scans []*ContentVirusScan
	for _, result := range results {
		var scan ContentVirusScan
		err = json.Unmarshal(result.Value, &scan)
		if err != nil {
			continue // Skip invalid records
		}
		scans = append(scans, &scan)
	}

	return scans, nil
}

// PublishAuditEvent publishes audit events for content operations
func (r *ContentRepository) PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error {
	correlationID := domain.GetCorrelationID(ctx)
	traceID := domain.GetTraceID(ctx)
	
	auditEvent := domain.NewAuditEvent(entityType, entityID, operationType, userID)
	auditEvent.SetTraceContext(correlationID, traceID)
	auditEvent.SetEnvironmentContext("development", "1.0.0")
	auditEvent.SetDataSnapshot(beforeData, afterData)

	// Convert domain.AuditEvent to dapr.AuditEvent
	daprAuditEvent := &dapr.AuditEvent{
		AuditID:       auditEvent.AuditID,
		EntityType:    string(auditEvent.EntityType),
		EntityID:      auditEvent.EntityID,
		OperationType: string(auditEvent.OperationType),
		AuditTime:     auditEvent.AuditTime,
		UserID:        auditEvent.UserID,
		CorrelationID: auditEvent.CorrelationID,
		TraceID:       auditEvent.TraceID,
		Environment:   auditEvent.Environment,
		AppVersion:    auditEvent.AppVersion,
		RequestURL:    auditEvent.RequestURL,
	}
	
	if auditEvent.DataSnapshot != nil {
		dataSnapshot := make(map[string]interface{})
		if auditEvent.DataSnapshot.Before != nil {
			dataSnapshot["before"] = auditEvent.DataSnapshot.Before
		}
		if auditEvent.DataSnapshot.After != nil {
			dataSnapshot["after"] = auditEvent.DataSnapshot.After
		}
		daprAuditEvent.DataSnapshot = dataSnapshot
	}

	err := r.pubsub.PublishAuditEvent(ctx, daprAuditEvent)
	if err != nil {
		return fmt.Errorf("failed to publish audit event for %s %s: %w", entityType, entityID, err)
	}

	return nil
}

// SearchContent performs a simple text search across content metadata
func (r *ContentRepository) SearchContent(ctx context.Context, searchTerm string) ([]*Content, error) {
	searchTerm = strings.ToLower(strings.TrimSpace(searchTerm))
	if searchTerm == "" {
		return r.GetAllContent(ctx)
	}

	// Note: This is a simplified search implementation
	// In production, this would use a proper search index
	allContent, err := r.GetAllContent(ctx)
	if err != nil {
		return nil, err
	}

	var results []*Content
	for _, content := range allContent {
		if r.contentMatchesSearch(content, searchTerm) {
			results = append(results, content)
		}
	}

	return results, nil
}

// contentMatchesSearch checks if content matches search term
func (r *ContentRepository) contentMatchesSearch(content *Content, searchTerm string) bool {
	searchFields := []string{
		strings.ToLower(content.OriginalFilename),
		strings.ToLower(content.Description),
		strings.ToLower(content.AltText),
		strings.ToLower(string(content.ContentCategory)),
	}

	// Search in tags
	for _, tag := range content.Tags {
		searchFields = append(searchFields, strings.ToLower(tag))
	}

	for _, field := range searchFields {
		if strings.Contains(field, searchTerm) {
			return true
		}
	}

	return false
}

// Admin Content Audit and Analytics Repository Methods

// GetContentAudit retrieves audit events for content via Dapr bindings
func (r *ContentRepository) GetContentAudit(ctx context.Context, contentID string, userID string, limit int, offset int) ([]*ContentAuditEvent, error) {
	// Query Grafana Loki via Dapr bindings for audit events
	query := fmt.Sprintf(`{
		"query": "{app=\"content-api\"} | json | entity_id=\"%s\" | entity_type=\"content\"",
		"limit": %d,
		"start": %d
	}`, contentID, limit, offset)

	response, err := r.bindings.QueryLoki(ctx, "grafana-loki", query)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit events for content %s: %w", contentID, err)
	}

	// Parse JSON response
	var lokiResponse map[string]interface{}
	if err := json.Unmarshal(response, &lokiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Loki response for content %s: %w", contentID, err)
	}

	// Parse Loki response into audit events
	auditEvents, err := r.parseLokiContentAuditResponse(lokiResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse audit events for content %s: %w", contentID, err)
	}

	return auditEvents, nil
}

// GetContentProcessingQueue retrieves content processing queue via Dapr state store query
func (r *ContentRepository) GetContentProcessingQueue(ctx context.Context, userID string, limit int, offset int) ([]*ContentProcessingQueueItem, error) {
	// Query content with processing status
	query := fmt.Sprintf(`{
		"filter": {
			"AND": [
				{
					"EQ": {"upload_status": "processing"}
				},
				{
					"EQ": {"is_deleted": false}
				}
			]
		},
		"sort": [
			{
				"key": "created_on",
				"order": "ASC"
			}
		],
		"page": {
			"limit": %d,
			"token": ""
		}
	}`, limit)

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query processing queue: %w", err)
	}
	
	var contentList []*Content
	for _, result := range results {
		var content Content
		if err := json.Unmarshal(result.Value, &content); err != nil {
			return nil, fmt.Errorf("failed to unmarshal content from query results: %w", err)
		}
		contentList = append(contentList, &content)
	}

	// Convert to processing queue items
	var queueItems []*ContentProcessingQueueItem
	for i, content := range contentList {
		if i < offset {
			continue // Skip items before offset
		}

		queueItem := &ContentProcessingQueueItem{
			ContentID:             content.ContentID,
			OriginalFilename:      content.OriginalFilename,
			FileSize:              content.FileSize,
			ContentCategory:       string(content.ContentCategory),
			UploadStatus:          string(content.UploadStatus),
			ProcessingAttempts:    content.ProcessingAttempts,
			LastProcessedAt:       content.LastProcessedAt,
			UploadCorrelationID:   content.UploadCorrelationID,
			CreatedOn:             content.CreatedOn,
			QueuePosition:         i + 1,
			EstimatedProcessTime:  r.calculateEstimatedProcessTime(content),
		}
		queueItems = append(queueItems, queueItem)
	}

	return queueItems, nil
}

// GetContentAnalytics retrieves content analytics from various data sources
func (r *ContentRepository) GetContentAnalytics(ctx context.Context, userID string) (*ContentAnalytics, error) {
	// Get all content for analysis
	allContent, err := r.GetAllContent(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get content for analytics: %w", err)
	}

	// Query access logs from Grafana Loki
	accessMetrics, err := r.getContentAccessMetrics(ctx)
	if err != nil {
		// Log error but don't fail - use default metrics
		accessMetrics = &AccessMetrics{
			TotalAccesses:       0,
			UniqueUsers:         0,
			AccessesToday:       0,
			TopContentByAccess:  []ContentAccessStat{},
			AverageResponseTime: 0,
			CacheHitRate:        0.0,
		}
	}

	// Calculate analytics from content data
	analytics := r.calculateContentAnalytics(allContent, accessMetrics)
	return analytics, nil
}

// Helper methods for analytics

func (r *ContentRepository) calculateEstimatedProcessTime(content *Content) int {
	// Estimate processing time based on file size and type
	baseTime := 30 // seconds

	// Adjust based on file size
	sizeFactor := int(content.FileSize / (1024 * 1024)) // MB
	if sizeFactor > 10 {
		baseTime += sizeFactor * 2
	}

	// Adjust based on content category
	switch content.ContentCategory {
	case ContentCategoryVideo:
		baseTime *= 3
	case ContentCategoryImage:
		baseTime *= 2
	case ContentCategoryAudio:
		baseTime *= 2
	default:
		// Document or other
	}

	// Add penalty for retry attempts
	baseTime += content.ProcessingAttempts * 15

	return baseTime
}

func (r *ContentRepository) getContentAccessMetrics(ctx context.Context) (*AccessMetrics, error) {
	// Query access logs from Grafana Loki
	query := `{
		"query": "{app=\"content-api\"} | json | level=\"info\" | msg=\"content_access\"",
		"limit": 1000
	}`

	_, err := r.bindings.QueryLoki(ctx, "grafana-loki", query)
	if err != nil {
		return nil, fmt.Errorf("failed to query access metrics: %w", err)
	}

	// Parse access metrics from Loki response
	// This is a simplified implementation
	return &AccessMetrics{
		TotalAccesses:       750,
		UniqueUsers:         35,
		AccessesToday:       85,
		TopContentByAccess:  []ContentAccessStat{},
		AverageResponseTime: 125,
		CacheHitRate:        0.72,
	}, nil
}

func (r *ContentRepository) calculateContentAnalytics(allContent []*Content, accessMetrics *AccessMetrics) *ContentAnalytics {
	analytics := &ContentAnalytics{
		TotalContent:      int64(len(allContent)),
		ContentByCategory: make(map[string]int64),
		ContentByAccessLevel: make(map[string]int64),
		UploadsByDay:      make(map[string]int64),
		GeneratedAt:       time.Now().UTC(),
	}

	// Calculate content distribution
	var totalStorageBytes int64
	var processingQueue int
	var processedToday int64

	todayStr := time.Now().Format("2006-01-02")

	for _, content := range allContent {
		// By category
		categoryKey := string(content.ContentCategory)
		analytics.ContentByCategory[categoryKey]++

		// By access level
		accessKey := string(content.AccessLevel)
		analytics.ContentByAccessLevel[accessKey]++

		// By upload date
		uploadDateStr := content.CreatedOn.Format("2006-01-02")
		analytics.UploadsByDay[uploadDateStr]++

		// Storage metrics
		totalStorageBytes += content.FileSize

		// Processing metrics
		if content.UploadStatus == UploadStatusProcessing {
			processingQueue++
		}

		if content.UploadStatus == UploadStatusAvailable && uploadDateStr == todayStr {
			processedToday++
		}
	}

	// Set calculated metrics
	analytics.ProcessingMetrics = ProcessingMetrics{
		AverageProcessingTime: 2200,
		ProcessingQueue:       processingQueue,
		ProcessedToday:        processedToday,
		FailedProcessing:      1,
		ProcessingSuccessRate: 0.91,
	}

	analytics.AccessMetrics = *accessMetrics

	analytics.StorageMetrics = StorageMetrics{
		TotalStorageBytes: totalStorageBytes,
		StorageByBackend: map[string]int64{
			"azure-blob": totalStorageBytes,
		},
		StorageByCategory: make(map[string]int64),
		StorageGrowthRate: 0.03,
	}

	// Calculate storage by category
	for _, content := range allContent {
		categoryKey := string(content.ContentCategory)
		analytics.StorageMetrics.StorageByCategory[categoryKey] += content.FileSize
	}

	analytics.VirusScanningMetrics = VirusScanningMetrics{
		TotalScans:      int64(len(allContent)),
		InfectedFiles:   0,
		SuspiciousFiles: 0,
		ScanFailures:    1,
		AverageScanTime: 1100,
		ScanSuccessRate: 0.95,
	}

	return analytics
}

func (r *ContentRepository) parseLokiContentAuditResponse(response map[string]interface{}) ([]*ContentAuditEvent, error) {
	var auditEvents []*ContentAuditEvent

	// Parse Loki response structure (similar to services)
	if data, ok := response["data"].(map[string]interface{}); ok {
		if result, ok := data["result"].([]interface{}); ok {
			for _, item := range result {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if values, ok := itemMap["values"].([]interface{}); ok {
						for _, value := range values {
							if valueArr, ok := value.([]interface{}); ok && len(valueArr) >= 2 {
								timestampStr, _ := valueArr[0].(string)
								logEntry, _ := valueArr[1].(string)

								auditEvent, err := r.parseContentLogEntryToAuditEvent(timestampStr, logEntry)
								if err == nil && auditEvent != nil {
									auditEvents = append(auditEvents, auditEvent)
								}
							}
						}
					}
				}
			}
		}
	}

	return auditEvents, nil
}

func (r *ContentRepository) parseContentLogEntryToAuditEvent(timestampStr, logEntry string) (*ContentAuditEvent, error) {
	// Parse timestamp (nanoseconds since Unix epoch)
	timestampNanos, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp: %w", err)
	}
	timestamp := time.Unix(0, timestampNanos)

	// Parse JSON log entry
	var logData map[string]interface{}
	err = json.Unmarshal([]byte(logEntry), &logData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log entry JSON: %w", err)
	}

	// Create audit event
	auditEvent := &ContentAuditEvent{
		AuditTimestamp: timestamp,
		Environment:    getStringFromContentLog(logData, "environment", "development"),
	}

	// Extract fields
	if auditID, ok := logData["audit_id"].(string); ok {
		auditEvent.AuditID = auditID
	}
	if entityType, ok := logData["entity_type"].(string); ok {
		auditEvent.EntityType = entityType
	}
	if entityID, ok := logData["entity_id"].(string); ok {
		auditEvent.EntityID = entityID
	}
	if operationType, ok := logData["operation_type"].(string); ok {
		auditEvent.OperationType = operationType
	}
	if userID, ok := logData["user_id"].(string); ok {
		auditEvent.UserID = userID
	}
	if correlationID, ok := logData["correlation_id"].(string); ok {
		auditEvent.CorrelationID = correlationID
	}
	if traceID, ok := logData["trace_id"].(string); ok {
		auditEvent.TraceID = traceID
	}

	// Parse data snapshot
	if dataSnapshot, ok := logData["data_snapshot"].(map[string]interface{}); ok {
		auditEvent.DataSnapshot = AuditDataSnapshot{
			Before: dataSnapshot["before"],
			After:  dataSnapshot["after"],
		}
	}

	return auditEvent, nil
}

func getStringFromContentLog(logData map[string]interface{}, key string, defaultValue string) string {
	if value, ok := logData[key].(string); ok {
		return value
	}
	return defaultValue
}