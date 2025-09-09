package research

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
)

// ResearchRepository implements research data access using Dapr state store and bindings
type ResearchRepository struct {
	stateStore *dapr.StateStore
	bindings   *dapr.Bindings
	pubsub     *dapr.PubSub
}

// NewResearchRepository creates a new research repository with Dapr integration
func NewResearchRepository(stateStore *dapr.StateStore, bindings *dapr.Bindings, pubsub *dapr.PubSub) ResearchRepositoryInterface {
	return &ResearchRepository{
		stateStore: stateStore,
		bindings:   bindings,
		pubsub:     pubsub,
	}
}

// Research operations

// SaveResearch saves a research article to the state store
func (r *ResearchRepository) SaveResearch(ctx context.Context, research *Research) error {
	key := r.stateStore.CreateKey("research", "research", research.ResearchID)
	
	err := r.stateStore.Save(ctx, key, research, nil)
	if err != nil {
		return fmt.Errorf("failed to save research %s: %w", research.ResearchID, err)
	}

	return nil
}

// GetResearch retrieves research by ID from state store
func (r *ResearchRepository) GetResearch(ctx context.Context, researchID string) (*Research, error) {
	key := r.stateStore.CreateKey("research", "research", researchID)
	
	var research Research
	found, err := r.stateStore.Get(ctx, key, &research)
	if err != nil {
		return nil, fmt.Errorf("failed to get research %s: %w", researchID, err)
	}
	
	if !found || research.IsDeleted {
		return nil, domain.NewNotFoundError("research", researchID)
	}
	
	return &research, nil
}

// GetResearchBySlug retrieves research by slug from state store
func (r *ResearchRepository) GetResearchBySlug(ctx context.Context, slug string) (*Research, error) {
	// Query by slug using Dapr state store query
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
		return nil, fmt.Errorf("failed to query research by slug %s: %w", slug, err)
	}

	for _, result := range results {
		var research Research
		err = json.Unmarshal(result.Value, &research)
		if err != nil {
			continue
		}
		if !research.IsDeleted && research.Slug == slug {
			return &research, nil
		}
	}

	return nil, domain.NewNotFoundError("research", slug)
}

// GetAllResearch retrieves all research from state store with pagination
func (r *ResearchRepository) GetAllResearch(ctx context.Context, limit, offset int) ([]*Research, error) {
	query := `{
		"filter": {
			"EQ": {"is_deleted": false}
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
		return nil, fmt.Errorf("failed to query all research: %w", err)
	}

	var researchList []*Research
	count := 0
	for _, result := range results {
		// Apply offset
		if count < offset {
			count++
			continue
		}
		
		// Apply limit
		if len(researchList) >= limit {
			break
		}

		var research Research
		err = json.Unmarshal(result.Value, &research)
		if err != nil {
			continue // Skip invalid records
		}
		if !research.IsDeleted {
			researchList = append(researchList, &research)
		}
		count++
	}

	return researchList, nil
}

// GetResearchByCategory retrieves research by category from state store with pagination
func (r *ResearchRepository) GetResearchByCategory(ctx context.Context, categoryID string, limit, offset int) ([]*Research, error) {
	query := fmt.Sprintf(`{
		"filter": {
			"AND": [
				{"EQ": {"category_id": "%s"}},
				{"EQ": {"is_deleted": false}}
			]
		},
		"sort": [
			{
				"key": "created_on", 
				"order": "DESC"
			}
		]
	}`, categoryID)

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query research by category %s: %w", categoryID, err)
	}

	var researchList []*Research
	count := 0
	for _, result := range results {
		// Apply offset
		if count < offset {
			count++
			continue
		}
		
		// Apply limit
		if len(researchList) >= limit {
			break
		}

		var research Research
		err = json.Unmarshal(result.Value, &research)
		if err != nil {
			continue
		}
		if !research.IsDeleted && research.CategoryID == categoryID {
			researchList = append(researchList, &research)
		}
		count++
	}

	return researchList, nil
}

// GetResearchByPublishingStatus retrieves research by publishing status with pagination
func (r *ResearchRepository) GetResearchByPublishingStatus(ctx context.Context, status PublishingStatus, limit, offset int) ([]*Research, error) {
	query := fmt.Sprintf(`{
		"filter": {
			"AND": [
				{"EQ": {"publishing_status": "%s"}},
				{"EQ": {"is_deleted": false}}
			]
		},
		"sort": [
			{
				"key": "created_on",
				"order": "DESC"
			}
		]
	}`, status)

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query research by status %s: %w", status, err)
	}

	var researchList []*Research
	count := 0
	for _, result := range results {
		// Apply offset
		if count < offset {
			count++
			continue
		}
		
		// Apply limit
		if len(researchList) >= limit {
			break
		}

		var research Research
		err = json.Unmarshal(result.Value, &research)
		if err != nil {
			continue
		}
		if !research.IsDeleted && research.PublishingStatus == status {
			researchList = append(researchList, &research)
		}
		count++
	}

	return researchList, nil
}

// DeleteResearch soft deletes research from state store
func (r *ResearchRepository) DeleteResearch(ctx context.Context, researchID string) error {
	// Get existing research
	research, err := r.GetResearch(ctx, researchID)
	if err != nil {
		return err
	}

	// Soft delete
	now := time.Now()
	research.IsDeleted = true
	research.DeletedOn = &now

	return r.SaveResearch(ctx, research)
}

// SearchResearch searches research articles based on search term with pagination
func (r *ResearchRepository) SearchResearch(ctx context.Context, searchTerm string, limit, offset int) ([]*Research, error) {
	// Get all non-deleted research first
	allResearch, err := r.GetAllResearch(ctx, 1000, 0) // Get up to 1000 for search
	if err != nil {
		return nil, err
	}

	if searchTerm == "" {
		return applyPagination(allResearch, limit, offset), nil
	}

	// Simple text search in title, abstract, content, and keywords
	var results []*Research
	searchLower := strings.ToLower(searchTerm)
	
	for _, research := range allResearch {
		if r.matchesSearchTerm(research, searchLower) {
			results = append(results, research)
		}
	}

	return applyPagination(results, limit, offset), nil
}

// Research category operations

// SaveResearchCategory saves a research category to state store
func (r *ResearchRepository) SaveResearchCategory(ctx context.Context, category *ResearchCategory) error {
	key := r.stateStore.CreateKey("research", "category", category.CategoryID)
	
	err := r.stateStore.Save(ctx, key, category, nil)
	if err != nil {
		return fmt.Errorf("failed to save research category %s: %w", category.CategoryID, err)
	}

	return nil
}

// GetResearchCategory retrieves research category by ID
func (r *ResearchRepository) GetResearchCategory(ctx context.Context, categoryID string) (*ResearchCategory, error) {
	key := r.stateStore.CreateKey("research", "category", categoryID)
	
	var category ResearchCategory
	found, err := r.stateStore.Get(ctx, key, &category)
	if err != nil {
		return nil, fmt.Errorf("failed to get research category %s: %w", categoryID, err)
	}
	
	if !found || category.IsDeleted {
		return nil, domain.NewNotFoundError("research category", categoryID)
	}
	
	return &category, nil
}

// GetResearchCategoryBySlug retrieves research category by slug
func (r *ResearchRepository) GetResearchCategoryBySlug(ctx context.Context, slug string) (*ResearchCategory, error) {
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
		return nil, fmt.Errorf("failed to query research category by slug %s: %w", slug, err)
	}

	for _, result := range results {
		var category ResearchCategory
		err = json.Unmarshal(result.Value, &category)
		if err != nil {
			continue
		}
		if !category.IsDeleted && category.Slug == slug {
			return &category, nil
		}
	}

	return nil, domain.NewNotFoundError("research category", slug)
}

// GetAllResearchCategories retrieves all research categories
func (r *ResearchRepository) GetAllResearchCategories(ctx context.Context) ([]*ResearchCategory, error) {
	query := `{
		"filter": {
			"EQ": {"is_deleted": false}
		}
	}`

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all research categories: %w", err)
	}

	var categories []*ResearchCategory
	for _, result := range results {
		var category ResearchCategory
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

// GetDefaultUnassignedCategory retrieves the default unassigned research category
func (r *ResearchRepository) GetDefaultUnassignedCategory(ctx context.Context) (*ResearchCategory, error) {
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
		return nil, fmt.Errorf("failed to query default unassigned research category: %w", err)
	}

	for _, result := range results {
		var category ResearchCategory
		err = json.Unmarshal(result.Value, &category)
		if err != nil {
			continue
		}
		if !category.IsDeleted && category.IsDefaultUnassigned {
			return &category, nil
		}
	}

	return nil, domain.NewNotFoundError("default unassigned research category", "")
}

// DeleteResearchCategory soft deletes a research category
func (r *ResearchRepository) DeleteResearchCategory(ctx context.Context, categoryID string) error {
	// Get existing category
	category, err := r.GetResearchCategory(ctx, categoryID)
	if err != nil {
		return err
	}

	// Soft delete
	now := time.Now()
	category.IsDeleted = true
	category.DeletedOn = &now

	return r.SaveResearchCategory(ctx, category)
}

// Featured research operations

// SaveFeaturedResearch saves featured research to state store
func (r *ResearchRepository) SaveFeaturedResearch(ctx context.Context, featured *FeaturedResearch) error {
	key := r.stateStore.CreateKey("research", "featured", featured.FeaturedResearchID)
	
	err := r.stateStore.Save(ctx, key, featured, nil)
	if err != nil {
		return fmt.Errorf("failed to save featured research %s: %w", featured.FeaturedResearchID, err)
	}

	return nil
}

// GetFeaturedResearch retrieves the current featured research
func (r *ResearchRepository) GetFeaturedResearch(ctx context.Context) (*FeaturedResearch, error) {
	query := `{}`

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query featured research: %w", err)
	}

	// Should only be one featured research
	for _, result := range results {
		var featured FeaturedResearch
		err = json.Unmarshal(result.Value, &featured)
		if err != nil {
			continue
		}
		return &featured, nil
	}

	return nil, domain.NewNotFoundError("featured research", "")
}

// DeleteFeaturedResearch removes featured research
func (r *ResearchRepository) DeleteFeaturedResearch(ctx context.Context, featuredResearchID string) error {
	key := r.stateStore.CreateKey("research", "featured", featuredResearchID)
	
	err := r.stateStore.Delete(ctx, key, nil)
	if err != nil {
		return fmt.Errorf("failed to delete featured research %s: %w", featuredResearchID, err)
	}

	return nil
}

// Audit operations

// PublishAuditEvent publishes audit event to Grafana Loki via Dapr
func (r *ResearchRepository) PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error {
	correlationID := domain.GetCorrelationID(ctx)
	traceID := domain.GetTraceID(ctx)
	
	auditEvent := domain.NewAuditEvent(entityType, entityID, operationType, userID)
	auditEvent.SetTraceContext(correlationID, traceID)
	auditEvent.SetEnvironmentContext("development", "research-api-1.0.0")
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

// GetResearchAudit retrieves audit events for research via Dapr bindings
func (r *ResearchRepository) GetResearchAudit(ctx context.Context, researchID string, userID string, limit int, offset int) ([]*domain.AuditEvent, error) {
	// Query Grafana Loki via Dapr bindings for audit events
	query := fmt.Sprintf(`{
		"query": "{app=\"research-api\"} | json | entity_id=\"%s\" | entity_type=\"research\"",
		"limit": %d,
		"start": %d
	}`, researchID, limit, offset)

	response, err := r.bindings.QueryLoki(ctx, "grafana-loki", query)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit events for research %s: %w", researchID, err)
	}

	// Parse JSON response
	var lokiResponse map[string]interface{}
	if err := json.Unmarshal(response, &lokiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Loki response for research %s: %w", researchID, err)
	}

	// Parse Loki response into audit events
	auditEvents, err := r.parseLokiResearchAuditResponse(lokiResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse audit events for research %s: %w", researchID, err)
	}

	return auditEvents, nil
}

// GetResearchCategoryAudit retrieves audit events for research category via Dapr bindings
func (r *ResearchRepository) GetResearchCategoryAudit(ctx context.Context, categoryID string, userID string, limit int, offset int) ([]*domain.AuditEvent, error) {
	// Query Grafana Loki via Dapr bindings for audit events
	query := fmt.Sprintf(`{
		"query": "{app=\"research-api\"} | json | entity_id=\"%s\" | entity_type=\"research_category\"",
		"limit": %d,
		"start": %d
	}`, categoryID, limit, offset)

	response, err := r.bindings.QueryLoki(ctx, "grafana-loki", query)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit events for research category %s: %w", categoryID, err)
	}

	// Parse JSON response
	var lokiResponse map[string]interface{}
	if err := json.Unmarshal(response, &lokiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Loki response for research category %s: %w", categoryID, err)
	}

	// Parse Loki response into audit events
	auditEvents, err := r.parseLokiResearchAuditResponse(lokiResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse audit events for research category %s: %w", categoryID, err)
	}

	return auditEvents, nil
}

// parseLokiResearchAuditResponse parses Grafana Loki response into AuditEvent objects
func (r *ResearchRepository) parseLokiResearchAuditResponse(response map[string]interface{}) ([]*domain.AuditEvent, error) {
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

// Helper function to apply pagination to research slice
func applyPagination(research []*Research, limit, offset int) []*Research {
	if offset >= len(research) {
		return []*Research{}
	}
	
	startIdx := offset
	endIdx := offset + limit
	if endIdx > len(research) {
		endIdx = len(research)
	}
	
	return research[startIdx:endIdx]
}

// Helper function to check if research matches search term
func (r *ResearchRepository) matchesSearchTerm(research *Research, searchLower string) bool {
	titleMatch := strings.Contains(strings.ToLower(research.Title), searchLower)
	abstractMatch := strings.Contains(strings.ToLower(research.Abstract), searchLower)
	contentMatch := strings.Contains(strings.ToLower(research.Content), searchLower)
	
	// Search in keywords
	keywordMatch := false
	for _, keyword := range research.Keywords {
		if strings.Contains(strings.ToLower(keyword), searchLower) {
			keywordMatch = true
			break
		}
	}
	
	return titleMatch || abstractMatch || contentMatch || keywordMatch
}

// Helper function to safely get string from map
func getString(m map[string]interface{}, key string) string {
	if value, ok := m[key].(string); ok {
		return value
	}
	return ""
}