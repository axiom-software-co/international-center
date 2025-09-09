package services

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

// ServicesRepository implements services data access using Dapr state store and bindings
type ServicesRepository struct {
	stateStore *dapr.StateStore
	bindings   *dapr.Bindings
	pubsub     *dapr.PubSub
}

// NewServicesRepository creates a new services repository
func NewServicesRepository(client *dapr.Client) *ServicesRepository {
	return &ServicesRepository{
		stateStore: dapr.NewStateStore(client),
		bindings:   dapr.NewBindings(client),
		pubsub:     dapr.NewPubSub(client),
	}
}

// Service operations

// SaveService saves service to Dapr state store
func (r *ServicesRepository) SaveService(ctx context.Context, service *Service) error {
	key := r.stateStore.CreateKey("services", "service", service.ServiceID)
	
	err := r.stateStore.Save(ctx, key, service, nil)
	if err != nil {
		return fmt.Errorf("failed to save service %s: %w", service.ServiceID, err)
	}

	// Create index for slug search
	slugKey := r.stateStore.CreateIndexKey("services", "service", "slug", service.Slug)
	slugIndex := map[string]string{"service_id": service.ServiceID}
	
	err = r.stateStore.Save(ctx, slugKey, slugIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create slug index for service %s: %w", service.ServiceID, err)
	}

	// Create index for category search
	categoryKey := r.stateStore.CreateIndexKey("services", "service", "category", service.CategoryID)
	categoryIndex := map[string]interface{}{
		"service_id":   service.ServiceID,
		"created_on":   service.CreatedOn,
		"order_number": service.OrderNumber,
	}
	
	err = r.stateStore.Save(ctx, categoryKey, categoryIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create category index for service %s: %w", service.ServiceID, err)
	}

	// Create index for publishing status search
	statusKey := r.stateStore.CreateIndexKey("services", "service", "status", string(service.PublishingStatus))
	statusIndex := map[string]interface{}{
		"service_id": service.ServiceID,
		"created_on": service.CreatedOn,
	}
	
	err = r.stateStore.Save(ctx, statusKey, statusIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create status index for service %s: %w", service.ServiceID, err)
	}

	return nil
}

// GetService retrieves service by ID from Dapr state store
func (r *ServicesRepository) GetService(ctx context.Context, serviceID string) (*Service, error) {
	key := r.stateStore.CreateKey("services", "service", serviceID)
	
	var service Service
	found, err := r.stateStore.Get(ctx, key, &service)
	if err != nil {
		return nil, fmt.Errorf("failed to get service %s: %w", serviceID, err)
	}
	
	if !found {
		return nil, domain.NewNotFoundError("service", serviceID)
	}

	if service.IsDeleted {
		return nil, domain.NewNotFoundError("service", serviceID)
	}

	return &service, nil
}

// GetServiceBySlug retrieves service by slug from Dapr state store
func (r *ServicesRepository) GetServiceBySlug(ctx context.Context, slug string) (*Service, error) {
	query := fmt.Sprintf(`{
		"filter": {
			"AND": [
				{
					"EQ": {"slug": "%s"}
				},
				{
					"EQ": {"is_deleted": false}
				}
			]
		}
	}`, slug)

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query service by slug %s: %w", slug, err)
	}

	if len(results) == 0 {
		return nil, domain.NewNotFoundError("service with slug", slug)
	}

	var service Service
	err = json.Unmarshal(results[0].Value, &service)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal service: %w", err)
	}

	return &service, nil
}

// GetAllServices retrieves all non-deleted services from Dapr state store
func (r *ServicesRepository) GetAllServices(ctx context.Context) ([]*Service, error) {
	query := `{
		"filter": {
			"EQ": {"is_deleted": false}
		},
		"sort": [
			{
				"key": "order_number",
				"order": "ASC"
			},
			{
				"key": "created_on",
				"order": "DESC"
			}
		]
	}`

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all services: %w", err)
	}

	var services []*Service
	for _, result := range results {
		var service Service
		err = json.Unmarshal(result.Value, &service)
		if err != nil {
			continue // Skip invalid records
		}
		
		if !service.IsDeleted {
			services = append(services, &service)
		}
	}

	return services, nil
}

// GetServicesByCategory retrieves services by category from Dapr state store
func (r *ServicesRepository) GetServicesByCategory(ctx context.Context, categoryID string) ([]*Service, error) {
	query := fmt.Sprintf(`{
		"filter": {
			"AND": [
				{
					"EQ": {"category_id": "%s"}
				},
				{
					"EQ": {"is_deleted": false}
				}
			]
		},
		"sort": [
			{
				"key": "order_number",
				"order": "ASC"
			},
			{
				"key": "created_on",
				"order": "DESC"
			}
		]
	}`, categoryID)

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query services by category %s: %w", categoryID, err)
	}

	var services []*Service
	for _, result := range results {
		var service Service
		err = json.Unmarshal(result.Value, &service)
		if err != nil {
			continue // Skip invalid records
		}
		services = append(services, &service)
	}

	return services, nil
}

// GetServicesByPublishingStatus retrieves services by publishing status
func (r *ServicesRepository) GetServicesByPublishingStatus(ctx context.Context, status PublishingStatus) ([]*Service, error) {
	query := fmt.Sprintf(`{
		"filter": {
			"AND": [
				{
					"EQ": {"publishing_status": "%s"}
				},
				{
					"EQ": {"is_deleted": false}
				}
			]
		},
		"sort": [
			{
				"key": "order_number",
				"order": "ASC"
			},
			{
				"key": "created_on",
				"order": "DESC"
			}
		]
	}`, string(status))

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query services by publishing status %s: %w", status, err)
	}

	var services []*Service
	for _, result := range results {
		var service Service
		err = json.Unmarshal(result.Value, &service)
		if err != nil {
			continue // Skip invalid records
		}
		services = append(services, &service)
	}

	return services, nil
}

// DeleteService soft deletes service from Dapr state store
func (r *ServicesRepository) DeleteService(ctx context.Context, serviceID string, userID string) error {
	service, err := r.GetService(ctx, serviceID)
	if err != nil {
		return err
	}

	err = service.Delete(userID)
	if err != nil {
		return err
	}

	return r.SaveService(ctx, service)
}

// ServiceCategory operations

// SaveServiceCategory saves service category to Dapr state store
func (r *ServicesRepository) SaveServiceCategory(ctx context.Context, category *ServiceCategory) error {
	key := r.stateStore.CreateKey("services", "category", category.CategoryID)
	
	err := r.stateStore.Save(ctx, key, category, nil)
	if err != nil {
		return fmt.Errorf("failed to save service category %s: %w", category.CategoryID, err)
	}

	// Create index for slug search
	slugKey := r.stateStore.CreateIndexKey("services", "category", "slug", category.Slug)
	slugIndex := map[string]string{"category_id": category.CategoryID}
	
	err = r.stateStore.Save(ctx, slugKey, slugIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create slug index for service category %s: %w", category.CategoryID, err)
	}

	return nil
}

// GetServiceCategory retrieves service category by ID from Dapr state store
func (r *ServicesRepository) GetServiceCategory(ctx context.Context, categoryID string) (*ServiceCategory, error) {
	key := r.stateStore.CreateKey("services", "category", categoryID)
	
	var category ServiceCategory
	found, err := r.stateStore.Get(ctx, key, &category)
	if err != nil {
		return nil, fmt.Errorf("failed to get service category %s: %w", categoryID, err)
	}
	
	if !found {
		return nil, domain.NewNotFoundError("service category", categoryID)
	}

	if category.IsDeleted {
		return nil, domain.NewNotFoundError("service category", categoryID)
	}

	return &category, nil
}

// GetServiceCategoryBySlug retrieves service category by slug from Dapr state store
func (r *ServicesRepository) GetServiceCategoryBySlug(ctx context.Context, slug string) (*ServiceCategory, error) {
	query := fmt.Sprintf(`{
		"filter": {
			"AND": [
				{
					"EQ": {"slug": "%s"}
				},
				{
					"EQ": {"is_deleted": false}
				}
			]
		}
	}`, slug)

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query service category by slug %s: %w", slug, err)
	}

	if len(results) == 0 {
		return nil, domain.NewNotFoundError("service category with slug", slug)
	}

	var category ServiceCategory
	err = json.Unmarshal(results[0].Value, &category)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal service category: %w", err)
	}

	return &category, nil
}

// GetAllServiceCategories retrieves all non-deleted service categories from Dapr state store
func (r *ServicesRepository) GetAllServiceCategories(ctx context.Context) ([]*ServiceCategory, error) {
	query := `{
		"filter": {
			"EQ": {"is_deleted": false}
		},
		"sort": [
			{
				"key": "order_number",
				"order": "ASC"
			},
			{
				"key": "name",
				"order": "ASC"
			}
		]
	}`

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all service categories: %w", err)
	}

	var categories []*ServiceCategory
	for _, result := range results {
		var category ServiceCategory
		err = json.Unmarshal(result.Value, &category)
		if err != nil {
			continue // Skip invalid records
		}
		
		if !category.IsDeleted {
			categories = append(categories, &category)
		}
	}

	return categories, nil
}

// GetDefaultUnassignedCategory retrieves the default unassigned category
func (r *ServicesRepository) GetDefaultUnassignedCategory(ctx context.Context) (*ServiceCategory, error) {
	query := `{
		"filter": {
			"AND": [
				{
					"EQ": {"is_default_unassigned": true}
				},
				{
					"EQ": {"is_deleted": false}
				}
			]
		}
	}`

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query default unassigned category: %w", err)
	}

	if len(results) == 0 {
		return nil, domain.NewNotFoundError("default unassigned category", "")
	}

	var category ServiceCategory
	err = json.Unmarshal(results[0].Value, &category)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal default unassigned category: %w", err)
	}

	return &category, nil
}

// DeleteServiceCategory soft deletes service category from Dapr state store
func (r *ServicesRepository) DeleteServiceCategory(ctx context.Context, categoryID string, userID string) error {
	category, err := r.GetServiceCategory(ctx, categoryID)
	if err != nil {
		return err
	}

	err = category.Delete(userID)
	if err != nil {
		return err
	}

	return r.SaveServiceCategory(ctx, category)
}

// FeaturedCategory operations

// SaveFeaturedCategory saves featured category to Dapr state store
func (r *ServicesRepository) SaveFeaturedCategory(ctx context.Context, featured *FeaturedCategory) error {
	key := r.stateStore.CreateKey("services", "featured", featured.FeaturedCategoryID)
	
	err := r.stateStore.Save(ctx, key, featured, nil)
	if err != nil {
		return fmt.Errorf("failed to save featured category %s: %w", featured.FeaturedCategoryID, err)
	}

	// Create index for position search
	positionKey := r.stateStore.CreateIndexKey("services", "featured", "position", fmt.Sprintf("%d", featured.FeaturePosition))
	positionIndex := map[string]interface{}{
		"featured_category_id": featured.FeaturedCategoryID,
		"category_id":         featured.CategoryID,
	}
	
	err = r.stateStore.Save(ctx, positionKey, positionIndex, nil)
	if err != nil {
		return fmt.Errorf("failed to create position index for featured category %s: %w", featured.FeaturedCategoryID, err)
	}

	return nil
}

// GetFeaturedCategory retrieves featured category by ID from Dapr state store
func (r *ServicesRepository) GetFeaturedCategory(ctx context.Context, featuredCategoryID string) (*FeaturedCategory, error) {
	key := r.stateStore.CreateKey("services", "featured", featuredCategoryID)
	
	var featured FeaturedCategory
	found, err := r.stateStore.Get(ctx, key, &featured)
	if err != nil {
		return nil, fmt.Errorf("failed to get featured category %s: %w", featuredCategoryID, err)
	}
	
	if !found {
		return nil, domain.NewNotFoundError("featured category", featuredCategoryID)
	}

	return &featured, nil
}

// GetAllFeaturedCategories retrieves all featured categories ordered by position
func (r *ServicesRepository) GetAllFeaturedCategories(ctx context.Context) ([]*FeaturedCategory, error) {
	query := `{
		"sort": [
			{
				"key": "feature_position",
				"order": "ASC"
			}
		]
	}`

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all featured categories: %w", err)
	}

	var featured []*FeaturedCategory
	for _, result := range results {
		var featuredCategory FeaturedCategory
		err = json.Unmarshal(result.Value, &featuredCategory)
		if err != nil {
			continue // Skip invalid records
		}
		featured = append(featured, &featuredCategory)
	}

	return featured, nil
}

// GetAdminFeaturedCategories retrieves featured categories with admin details
func (r *ServicesRepository) GetAdminFeaturedCategories(ctx context.Context) ([]*FeaturedCategory, error) {
	// For admin endpoint, return all featured categories (same as GetAllFeaturedCategories for now)
	// In a real implementation, this might include additional admin fields
	return r.GetAllFeaturedCategories(ctx)
}

// GetFeaturedCategoryByPosition retrieves featured category by position
func (r *ServicesRepository) GetFeaturedCategoryByPosition(ctx context.Context, position int) (*FeaturedCategory, error) {
	query := fmt.Sprintf(`{
		"filter": {
			"EQ": {"feature_position": %d}
		}
	}`, position)

	results, err := r.stateStore.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query featured category by position %d: %w", position, err)
	}

	if len(results) == 0 {
		return nil, domain.NewNotFoundError("featured category at position", fmt.Sprintf("%d", position))
	}

	var featured FeaturedCategory
	err = json.Unmarshal(results[0].Value, &featured)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal featured category: %w", err)
	}

	return &featured, nil
}

// DeleteFeaturedCategory deletes featured category from Dapr state store
func (r *ServicesRepository) DeleteFeaturedCategory(ctx context.Context, featuredCategoryID string) error {
	key := r.stateStore.CreateKey("services", "featured", featuredCategoryID)
	
	err := r.stateStore.Delete(ctx, key, nil)
	if err != nil {
		return fmt.Errorf("failed to delete featured category %s: %w", featuredCategoryID, err)
	}

	return nil
}

// Content storage operations

// UploadServiceContentBlob uploads service content to blob storage via Dapr bindings
func (r *ServicesRepository) UploadServiceContentBlob(ctx context.Context, storagePath string, data []byte, contentType string) error {
	err := r.bindings.UploadBlob(ctx, storagePath, data, contentType)
	if err != nil {
		return fmt.Errorf("failed to upload service content blob to %s: %w", storagePath, err)
	}

	return nil
}

// DownloadServiceContentBlob downloads service content from blob storage via Dapr bindings
func (r *ServicesRepository) DownloadServiceContentBlob(ctx context.Context, storagePath string) ([]byte, error) {
	data, err := r.bindings.DownloadBlob(ctx, storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to download service content blob from %s: %w", storagePath, err)
	}

	return data, nil
}

// CreateServiceContentBlobURL creates a temporary access URL for service content blob
func (r *ServicesRepository) CreateServiceContentBlobURL(ctx context.Context, storagePath string, expiryMinutes int) (string, error) {
	url, err := r.bindings.CreateBlobURL(ctx, storagePath, expiryMinutes)
	if err != nil {
		return "", fmt.Errorf("failed to create service content blob URL for %s: %w", storagePath, err)
	}

	return url, nil
}

// PublishAuditEvent publishes audit events for services operations
func (r *ServicesRepository) PublishAuditEvent(ctx context.Context, entityType domain.EntityType, entityID string, operationType domain.AuditEventType, userID string, beforeData, afterData interface{}) error {
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

// SearchServices performs a simple text search across services metadata
func (r *ServicesRepository) SearchServices(ctx context.Context, searchTerm string) ([]*Service, error) {
	searchTerm = strings.ToLower(strings.TrimSpace(searchTerm))
	if searchTerm == "" {
		return r.GetAllServices(ctx)
	}

	// Note: This is a simplified search implementation
	// In production, this would use a proper search index
	allServices, err := r.GetAllServices(ctx)
	if err != nil {
		return nil, err
	}

	var results []*Service
	for _, service := range allServices {
		if r.serviceMatchesSearch(service, searchTerm) {
			results = append(results, service)
		}
	}

	return results, nil
}

// serviceMatchesSearch checks if service matches search term
func (r *ServicesRepository) serviceMatchesSearch(service *Service, searchTerm string) bool {
	searchFields := []string{
		strings.ToLower(service.Title),
		strings.ToLower(service.Description),
		strings.ToLower(service.Slug),
		strings.ToLower(string(service.DeliveryMode)),
		strings.ToLower(string(service.PublishingStatus)),
	}

	for _, field := range searchFields {
		if strings.Contains(field, searchTerm) {
			return true
		}
	}

	return false
}

// Admin Audit Repository Methods

// GetServiceAudit retrieves audit events for a service via Dapr bindings
func (r *ServicesRepository) GetServiceAudit(ctx context.Context, serviceID string, limit int, offset int) ([]*ServiceAuditEvent, error) {
	// Query Grafana Loki via Dapr bindings for audit events
	query := fmt.Sprintf(`{
		"query": "{app=\"services-api\"} | json | entity_id=\"%s\" | entity_type=\"service\"",
		"limit": %d,
		"start": %d
	}`, serviceID, limit, offset)

	response, err := r.bindings.QueryLoki(ctx, "grafana-loki", query)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit events for service %s: %w", serviceID, err)
	}

	// Parse JSON response
	var lokiResponse map[string]interface{}
	if err := json.Unmarshal(response, &lokiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Loki response for service %s: %w", serviceID, err)
	}

	// Parse Loki response into audit events
	auditEvents, err := r.parseLokiAuditResponse(lokiResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse audit events for service %s: %w", serviceID, err)
	}

	return auditEvents, nil
}

// GetServiceCategoryAudit retrieves audit events for a service category via Dapr bindings
func (r *ServicesRepository) GetServiceCategoryAudit(ctx context.Context, categoryID string, limit int, offset int) ([]*ServiceAuditEvent, error) {
	// Query Grafana Loki via Dapr bindings for audit events
	query := fmt.Sprintf(`{
		"query": "{app=\"services-api\"} | json | entity_id=\"%s\" | entity_type=\"service_category\"",
		"limit": %d,
		"start": %d
	}`, categoryID, limit, offset)

	response, err := r.bindings.QueryLoki(ctx, "grafana-loki", query)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit events for category %s: %w", categoryID, err)
	}

	// Parse JSON response
	var lokiResponse map[string]interface{}
	if err := json.Unmarshal(response, &lokiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Loki response for category %s: %w", categoryID, err)
	}

	// Parse Loki response into audit events
	auditEvents, err := r.parseLokiAuditResponse(lokiResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse audit events for category %s: %w", categoryID, err)
	}

	return auditEvents, nil
}

// parseLokiAuditResponse parses Grafana Loki response into ServiceAuditEvent objects
func (r *ServicesRepository) parseLokiAuditResponse(response map[string]interface{}) ([]*ServiceAuditEvent, error) {
	var auditEvents []*ServiceAuditEvent

	// Parse Loki response structure
	if data, ok := response["data"].(map[string]interface{}); ok {
		if result, ok := data["result"].([]interface{}); ok {
			for _, item := range result {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if values, ok := itemMap["values"].([]interface{}); ok {
						for _, value := range values {
							if valueArr, ok := value.([]interface{}); ok && len(valueArr) >= 2 {
								// Parse timestamp and log entry
								timestampStr, _ := valueArr[0].(string)
								logEntry, _ := valueArr[1].(string)

								auditEvent, err := r.parseLogEntryToAuditEvent(timestampStr, logEntry)
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

// parseLogEntryToAuditEvent parses a Loki log entry into a ServiceAuditEvent
func (r *ServicesRepository) parseLogEntryToAuditEvent(timestampStr, logEntry string) (*ServiceAuditEvent, error) {
	// Parse timestamp
	timestamp, err := parseTimestampFromLoki(timestampStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	// Parse JSON log entry
	var logData map[string]interface{}
	err = json.Unmarshal([]byte(logEntry), &logData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log entry JSON: %w", err)
	}

	// Extract audit event fields
	auditEvent := &ServiceAuditEvent{
		AuditTimestamp: timestamp,
		Environment:    getStringFromLog(logData, "environment", "development"),
	}

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

// Helper functions

func parseTimestampFromLoki(timestampStr string) (time.Time, error) {
	// Loki timestamps are in nanoseconds since Unix epoch
	timestampNanos, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	
	return time.Unix(0, timestampNanos), nil
}

func getStringFromLog(logData map[string]interface{}, key string, defaultValue string) string {
	if value, ok := logData[key].(string); ok {
		return value
	}
	return defaultValue
}