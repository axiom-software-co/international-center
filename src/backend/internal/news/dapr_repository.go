package news

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
)

// NewsRepository implements news data access using Dapr state store and bindings
type NewsRepository struct {
	stateStore *dapr.StateStore
	bindings   *dapr.Bindings
	pubsub     *dapr.PubSub
}

// NewNewsRepository creates a new news repository with Dapr integration
func NewNewsRepository(stateStore *dapr.StateStore, bindings *dapr.Bindings, pubsub *dapr.PubSub) NewsRepositoryInterface {
	return &NewsRepository{
		stateStore: stateStore,
		bindings:   bindings,
		pubsub:     pubsub,
	}
}

// News operations

// SaveNews saves a news article to the state store
func (r *NewsRepository) SaveNews(ctx context.Context, news *News) error {
	key := r.stateStore.CreateKey("news", "news", news.NewsID)
	
	err := r.stateStore.Save(ctx, key, news, nil)
	if err != nil {
		return fmt.Errorf("failed to save news %s: %w", news.NewsID, err)
	}

	return nil
}

// GetNews retrieves news by ID from state store
func (r *NewsRepository) GetNews(ctx context.Context, newsID string) (*News, error) {
	key := r.stateStore.CreateKey("news", "news", newsID)
	
	var news News
	found, err := r.stateStore.Get(ctx, key, &news)
	if err != nil {
		return nil, fmt.Errorf("failed to get news %s: %w", newsID, err)
	}
	
	if !found || news.IsDeleted {
		return nil, domain.NewNotFoundError("news", newsID)
	}
	
	return &news, nil
}

// GetNewsBySlug retrieves news by slug from state store
func (r *NewsRepository) GetNewsBySlug(ctx context.Context, slug string) (*News, error) {
	// Query by slug using Dapr state store query
	query := fmt.Sprintf(`{
		"filter": {
			"EQ": {"slug": "%s"}
		}
	}`, slug)

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query news by slug %s: %w", slug, err)
	}

	for _, result := range results {
		var news News
		err = json.Unmarshal(result.Value, &news)
		if err != nil {
			continue
		}
		if !news.IsDeleted && news.Slug == slug {
			return &news, nil
		}
	}

	return nil, domain.NewNotFoundError("news", slug)
}

// GetAllNews retrieves all news from state store
func (r *NewsRepository) GetAllNews(ctx context.Context) ([]*News, error) {
	query := `{
		"filter": {
			"EQ": {"is_deleted": false}
		},
		"sort": [
			{
				"key": "publication_timestamp",
				"order": "DESC"
			}
		]
	}`

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all news: %w", err)
	}

	var newsList []*News
	for _, result := range results {
		var news News
		err = json.Unmarshal(result.Value, &news)
		if err != nil {
			continue // Skip invalid records
		}
		if !news.IsDeleted {
			newsList = append(newsList, &news)
		}
	}

	return newsList, nil
}

// GetNewsByCategory retrieves news by category from state store
func (r *NewsRepository) GetNewsByCategory(ctx context.Context, categoryID string) ([]*News, error) {
	query := fmt.Sprintf(`{
		"filter": {
			"AND": [
				{"EQ": {"category_id": "%s"}},
				{"EQ": {"is_deleted": false}}
			]
		},
		"sort": [
			{
				"key": "publication_timestamp", 
				"order": "DESC"
			}
		]
	}`, categoryID)

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query news by category %s: %w", categoryID, err)
	}

	var newsList []*News
	for _, result := range results {
		var news News
		err = json.Unmarshal(result.Value, &news)
		if err != nil {
			continue
		}
		if !news.IsDeleted && news.CategoryID == categoryID {
			newsList = append(newsList, &news)
		}
	}

	return newsList, nil
}

// GetNewsByPublishingStatus retrieves news by publishing status
func (r *NewsRepository) GetNewsByPublishingStatus(ctx context.Context, status PublishingStatus) ([]*News, error) {
	query := fmt.Sprintf(`{
		"filter": {
			"AND": [
				{"EQ": {"publishing_status": "%s"}},
				{"EQ": {"is_deleted": false}}
			]
		},
		"sort": [
			{
				"key": "publication_timestamp",
				"order": "DESC"
			}
		]
	}`, status)

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query news by status %s: %w", status, err)
	}

	var newsList []*News
	for _, result := range results {
		var news News
		err = json.Unmarshal(result.Value, &news)
		if err != nil {
			continue
		}
		if !news.IsDeleted && news.PublishingStatus == status {
			newsList = append(newsList, &news)
		}
	}

	return newsList, nil
}

// DeleteNews soft deletes news from state store
func (r *NewsRepository) DeleteNews(ctx context.Context, newsID string, userID string) error {
	// Get existing news
	news, err := r.GetNews(ctx, newsID)
	if err != nil {
		return err
	}

	// Soft delete
	now := time.Now()
	news.IsDeleted = true
	news.DeletedOn = &now
	news.DeletedBy = userID

	return r.SaveNews(ctx, news)
}

// SearchNews searches news articles based on search term
func (r *NewsRepository) SearchNews(ctx context.Context, searchTerm string) ([]*News, error) {
	// Get all non-deleted news first
	allNews, err := r.GetAllNews(ctx)
	if err != nil {
		return nil, err
	}

	if searchTerm == "" {
		return allNews, nil
	}

	// Simple text search in title, summary, and content
	var results []*News
	searchLower := strings.ToLower(searchTerm)
	
	for _, news := range allNews {
		titleMatch := strings.Contains(strings.ToLower(news.Title), searchLower)
		summaryMatch := strings.Contains(strings.ToLower(news.Summary), searchLower)
		contentMatch := strings.Contains(strings.ToLower(news.Content), searchLower)
		
		if titleMatch || summaryMatch || contentMatch {
			results = append(results, news)
		}
	}

	return results, nil
}

// News category operations

// SaveNewsCategory saves a news category to state store
func (r *NewsRepository) SaveNewsCategory(ctx context.Context, category *NewsCategory) error {
	key := r.stateStore.CreateKey("news", "category", category.CategoryID)
	
	err := r.stateStore.Save(ctx, key, category, nil)
	if err != nil {
		return fmt.Errorf("failed to save news category %s: %w", category.CategoryID, err)
	}

	return nil
}

// GetNewsCategory retrieves news category by ID
func (r *NewsRepository) GetNewsCategory(ctx context.Context, categoryID string) (*NewsCategory, error) {
	key := r.stateStore.CreateKey("news", "category", categoryID)
	
	var category NewsCategory
	found, err := r.stateStore.Get(ctx, key, &category)
	if err != nil {
		return nil, fmt.Errorf("failed to get news category %s: %w", categoryID, err)
	}
	
	if !found || category.IsDeleted {
		return nil, domain.NewNotFoundError("news category", categoryID)
	}
	
	return &category, nil
}

// GetNewsCategoryBySlug retrieves news category by slug
func (r *NewsRepository) GetNewsCategoryBySlug(ctx context.Context, slug string) (*NewsCategory, error) {
	query := fmt.Sprintf(`{
		"filter": {
			"AND": [
				{"EQ": {"slug": "%s"}},
				{"EQ": {"is_deleted": false}}
			]
		}
	}`, slug)

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query news category by slug %s: %w", slug, err)
	}

	for _, result := range results {
		var category NewsCategory
		err = json.Unmarshal(result.Value, &category)
		if err != nil {
			continue
		}
		if !category.IsDeleted && category.Slug == slug {
			return &category, nil
		}
	}

	return nil, domain.NewNotFoundError("news category", slug)
}

// GetAllNewsCategories retrieves all news categories
func (r *NewsRepository) GetAllNewsCategories(ctx context.Context) ([]*NewsCategory, error) {
	query := `{
		"filter": {
			"EQ": {"is_deleted": false}
		}
	}`

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all news categories: %w", err)
	}

	var categories []*NewsCategory
	for _, result := range results {
		var category NewsCategory
		err = json.Unmarshal(result.Value, &category)
		if err != nil {
			continue
		}
		if !category.IsDeleted {
			categories = append(categories, &category)
		}
	}

	return categories, nil
}

// GetDefaultUnassignedCategory retrieves the default unassigned news category
func (r *NewsRepository) GetDefaultUnassignedCategory(ctx context.Context) (*NewsCategory, error) {
	query := `{
		"filter": {
			"AND": [
				{"EQ": {"is_default_unassigned": true}},
				{"EQ": {"is_deleted": false}}
			]
		}
	}`

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query default unassigned news category: %w", err)
	}

	for _, result := range results {
		var category NewsCategory
		err = json.Unmarshal(result.Value, &category)
		if err != nil {
			continue
		}
		if !category.IsDeleted && category.IsDefaultUnassigned {
			return &category, nil
		}
	}

	return nil, domain.NewNotFoundError("default unassigned news category", "")
}

// DeleteNewsCategory soft deletes a news category
func (r *NewsRepository) DeleteNewsCategory(ctx context.Context, categoryID string, userID string) error {
	// Get existing category
	category, err := r.GetNewsCategory(ctx, categoryID)
	if err != nil {
		return err
	}

	// Soft delete
	now := time.Now()
	category.IsDeleted = true
	category.DeletedOn = &now
	category.DeletedBy = userID

	return r.SaveNewsCategory(ctx, category)
}

// Featured news operations

// SaveFeaturedNews saves featured news to state store
func (r *NewsRepository) SaveFeaturedNews(ctx context.Context, featured *FeaturedNews) error {
	key := r.stateStore.CreateKey("news", "featured", featured.FeaturedNewsID)
	
	err := r.stateStore.Save(ctx, key, featured, nil)
	if err != nil {
		return fmt.Errorf("failed to save featured news %s: %w", featured.FeaturedNewsID, err)
	}

	return nil
}

// GetFeaturedNews retrieves the current featured news
func (r *NewsRepository) GetFeaturedNews(ctx context.Context) (*FeaturedNews, error) {
	query := `{}`

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query featured news: %w", err)
	}

	// Should only be one featured news
	for _, result := range results {
		var featured FeaturedNews
		err = json.Unmarshal(result.Value, &featured)
		if err != nil {
			continue
		}
		return &featured, nil
	}

	return nil, domain.NewNotFoundError("featured news", "")
}

// DeleteFeaturedNews removes featured news
func (r *NewsRepository) DeleteFeaturedNews(ctx context.Context, featuredNewsID string) error {
	key := r.stateStore.CreateKey("news", "featured", featuredNewsID)
	
	err := r.stateStore.Delete(ctx, key, nil)
	if err != nil {
		return fmt.Errorf("failed to delete featured news %s: %w", featuredNewsID, err)
	}

	return nil
}

// Audit operations

// PublishAuditEvent publishes audit event to Grafana Loki via Dapr
func (r *NewsRepository) PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error {
	correlationID := domain.GetCorrelationID(ctx)
	traceID := domain.GetTraceID(ctx)
	
	auditEvent := domain.NewAuditEvent(entityType, entityID, operationType, userID)
	auditEvent.SetTraceContext(correlationID, traceID)
	auditEvent.SetEnvironmentContext("development", "news-api-1.0.0")
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
	
	// Convert data snapshot to map[string]interface{}
	if auditEvent.DataSnapshot != nil {
		daprAuditEvent.DataSnapshot = map[string]interface{}{
			"before": auditEvent.DataSnapshot.Before,
			"after":  auditEvent.DataSnapshot.After,
		}
	}

	err := r.pubsub.PublishAuditEvent(ctx, daprAuditEvent)
	if err != nil {
		return fmt.Errorf("failed to publish audit event for %s %s: %w", entityType, entityID, err)
	}

	return nil
}

// GetNewsAudit retrieves audit events for news via Dapr bindings
func (r *NewsRepository) GetNewsAudit(ctx context.Context, newsID string, userID string, limit int, offset int) ([]*domain.AuditEvent, error) {
	// Query Grafana Loki via Dapr bindings for audit events
	query := fmt.Sprintf(`{
		"query": "{app=\"news-api\"} | json | entity_id=\"%s\" | entity_type=\"news\"",
		"limit": %d,
		"start": %d
	}`, newsID, limit, offset)

	response, err := r.bindings.QueryLoki(ctx, "grafana-loki", query)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit events for news %s: %w", newsID, err)
	}

	// Parse JSON response
	var lokiResponse map[string]interface{}
	if err := json.Unmarshal(response, &lokiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Loki response for news %s: %w", newsID, err)
	}

	// Parse Loki response into audit events
	auditEvents, err := r.parseLokiNewsAuditResponse(lokiResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse audit events for news %s: %w", newsID, err)
	}

	return auditEvents, nil
}

// GetNewsCategoryAudit retrieves audit events for news category via Dapr bindings
func (r *NewsRepository) GetNewsCategoryAudit(ctx context.Context, categoryID string, userID string, limit int, offset int) ([]*domain.AuditEvent, error) {
	// Query Grafana Loki via Dapr bindings for audit events
	query := fmt.Sprintf(`{
		"query": "{app=\"news-api\"} | json | entity_id=\"%s\" | entity_type=\"news_category\"",
		"limit": %d,
		"start": %d
	}`, categoryID, limit, offset)

	response, err := r.bindings.QueryLoki(ctx, "grafana-loki", query)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit events for news category %s: %w", categoryID, err)
	}

	// Parse JSON response
	var lokiResponse map[string]interface{}
	if err := json.Unmarshal(response, &lokiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Loki response for news category %s: %w", categoryID, err)
	}

	// Parse Loki response into audit events
	auditEvents, err := r.parseLokiNewsAuditResponse(lokiResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse audit events for news category %s: %w", categoryID, err)
	}

	return auditEvents, nil
}

// parseLokiNewsAuditResponse parses Grafana Loki response into AuditEvent objects
func (r *NewsRepository) parseLokiNewsAuditResponse(response map[string]interface{}) ([]*domain.AuditEvent, error) {
	var auditEvents []*domain.AuditEvent

	// Parse Loki response structure
	// This is a simplified implementation - in production you'd handle the full Loki response format
	if data, ok := response["data"].(map[string]interface{}); ok {
		if result, ok := data["result"].([]interface{}); ok {
			for _, item := range result {
				if itemMap, ok := item.(map[string]interface{}); ok {
					event := &domain.AuditEvent{
						AuditID:       fmt.Sprintf("audit-%d", time.Now().UnixNano()),
						EntityType:    domain.EntityType(getString(itemMap, "entity_type")),
						EntityID:      getString(itemMap, "entity_id"),
						OperationType: domain.AuditEventType(getString(itemMap, "operation_type")),
						AuditTime:     time.Now(),
						UserID:        getString(itemMap, "user_id"),
						CorrelationID: getString(itemMap, "correlation_id"),
						Environment:   getString(itemMap, "environment"),
					}
					auditEvents = append(auditEvents, event)
				}
			}
		}
	}

	return auditEvents, nil
}

// Helper function to safely get string from map
func getString(m map[string]interface{}, key string) string {
	if value, ok := m[key].(string); ok {
		return value
	}
	return ""
}