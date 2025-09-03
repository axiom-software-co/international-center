package content

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/axiom-software-co/international-center/src/internal/shared/dapr"
	"github.com/axiom-software-co/international-center/src/internal/shared/domain"
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

	err := r.pubsub.PublishAuditEvent(ctx, auditEvent)
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