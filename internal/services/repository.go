package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dapr/go-sdk/client"
	"github.com/google/uuid"
)

type ServicesRepository interface {
	Create(service *Service) error
	GetByID(serviceID string) (*Service, error)
	GetBySlug(slug string) (*Service, error)
	Update(service *Service) error
	Delete(serviceID, userID string) error
	List(offset, limit int) ([]*Service, error)
	ListByCategory(categoryID string, offset, limit int) ([]*Service, error)
	ListPublished(offset, limit int) ([]*Service, error)
}

type PostgreSQLServicesRepository struct {
	db *sql.DB
}

func NewPostgreSQLServicesRepository(db *sql.DB) *PostgreSQLServicesRepository {
	return &PostgreSQLServicesRepository{db: db}
}

func (r *PostgreSQLServicesRepository) Create(service *Service) error {
	query := `
		INSERT INTO services (
			service_id, title, description, slug, category_id, 
			delivery_mode, publishing_status, created_on, is_deleted
		) VALUES (
			gen_random_uuid(), $1, $2, $3, COALESCE($4, (
				SELECT category_id FROM service_categories 
				WHERE is_default_unassigned = true AND is_deleted = false 
				LIMIT 1
			)), $5, $6, $7, false
		) RETURNING service_id, created_on`
	
	err := r.db.QueryRow(query, 
		service.Title, 
		service.Description, 
		service.Slug, 
		nullString(service.CategoryID),
		service.DeliveryMode, 
		service.PublishingStatus,
		time.Now(),
	).Scan(&service.ServiceID, &service.CreatedOn)
	
	return err
}

func (r *PostgreSQLServicesRepository) GetByID(serviceID string) (*Service, error) {
	query := `
		SELECT service_id, title, description, slug, category_id, 
			   delivery_mode, publishing_status, is_deleted, created_on, 
			   modified_on, COALESCE(modified_by, '')
		FROM services 
		WHERE service_id = $1 AND is_deleted = false`
	
	service := &Service{}
	var categoryID sql.NullString
	var modifiedOn sql.NullTime
	
	err := r.db.QueryRow(query, serviceID).Scan(
		&service.ServiceID,
		&service.Title,
		&service.Description,
		&service.Slug,
		&categoryID,
		&service.DeliveryMode,
		&service.PublishingStatus,
		&service.IsDeleted,
		&service.CreatedOn,
		&modifiedOn,
		&service.ModifiedBy,
	)
	
	if categoryID.Valid {
		service.CategoryID = categoryID.String
	}
	
	if modifiedOn.Valid {
		service.ModifiedOn = modifiedOn.Time
	}
	
	return service, err
}

func (r *PostgreSQLServicesRepository) GetBySlug(slug string) (*Service, error) {
	query := `
		SELECT service_id, title, description, slug, category_id, 
			   delivery_mode, publishing_status, is_deleted, created_on, 
			   modified_on, COALESCE(modified_by, '')
		FROM services 
		WHERE slug = $1 AND is_deleted = false`
	
	service := &Service{}
	var categoryID sql.NullString
	var modifiedOn sql.NullTime
	
	err := r.db.QueryRow(query, slug).Scan(
		&service.ServiceID,
		&service.Title,
		&service.Description,
		&service.Slug,
		&categoryID,
		&service.DeliveryMode,
		&service.PublishingStatus,
		&service.IsDeleted,
		&service.CreatedOn,
		&modifiedOn,
		&service.ModifiedBy,
	)
	
	if categoryID.Valid {
		service.CategoryID = categoryID.String
	}
	
	if modifiedOn.Valid {
		service.ModifiedOn = modifiedOn.Time
	}
	
	return service, err
}

func (r *PostgreSQLServicesRepository) Update(service *Service) error {
	query := `
		UPDATE services 
		SET title = $1, description = $2, slug = $3, category_id = $4,
			delivery_mode = $5, publishing_status = $6, modified_on = $7, modified_by = $8
		WHERE service_id = $9 AND is_deleted = false`
	
	_, err := r.db.Exec(query,
		service.Title,
		service.Description,
		service.Slug,
		nullString(service.CategoryID),
		service.DeliveryMode,
		service.PublishingStatus,
		time.Now(),
		service.ModifiedBy,
		service.ServiceID,
	)
	
	return err
}

func (r *PostgreSQLServicesRepository) Delete(serviceID, userID string) error {
	query := `
		UPDATE services 
		SET is_deleted = true, deleted_on = $1, deleted_by = $2
		WHERE service_id = $3 AND is_deleted = false`
	
	_, err := r.db.Exec(query, time.Now(), userID, serviceID)
	return err
}

func (r *PostgreSQLServicesRepository) List(offset, limit int) ([]*Service, error) {
	query := `
		SELECT service_id, title, description, slug, category_id, 
			   delivery_mode, publishing_status, is_deleted, created_on, 
			   modified_on, COALESCE(modified_by, '')
		FROM services 
		WHERE is_deleted = false
		ORDER BY created_on DESC
		LIMIT $1 OFFSET $2`
	
	return r.scanServices(query, limit, offset)
}

func (r *PostgreSQLServicesRepository) ListByCategory(categoryID string, offset, limit int) ([]*Service, error) {
	query := `
		SELECT service_id, title, description, slug, category_id, 
			   delivery_mode, publishing_status, is_deleted, created_on, 
			   modified_on, COALESCE(modified_by, '')
		FROM services 
		WHERE category_id = $1 AND is_deleted = false
		ORDER BY order_number ASC, created_on DESC
		LIMIT $2 OFFSET $3`
	
	return r.scanServices(query, categoryID, limit, offset)
}

func (r *PostgreSQLServicesRepository) ListPublished(offset, limit int) ([]*Service, error) {
	query := `
		SELECT service_id, title, description, slug, category_id, 
			   delivery_mode, publishing_status, is_deleted, created_on, 
			   modified_on, COALESCE(modified_by, '')
		FROM services 
		WHERE publishing_status = 'published' AND is_deleted = false
		ORDER BY order_number ASC, created_on DESC
		LIMIT $1 OFFSET $2`
	
	return r.scanServices(query, limit, offset)
}

func (r *PostgreSQLServicesRepository) scanServices(query string, args ...interface{}) ([]*Service, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var services []*Service
	for rows.Next() {
		service := &Service{}
		var categoryID sql.NullString
		var modifiedOn sql.NullTime
		
		err := rows.Scan(
			&service.ServiceID,
			&service.Title,
			&service.Description,
			&service.Slug,
			&categoryID,
			&service.DeliveryMode,
			&service.PublishingStatus,
			&service.IsDeleted,
			&service.CreatedOn,
			&modifiedOn,
			&service.ModifiedBy,
		)
		if err != nil {
			return nil, err
		}
		
		if categoryID.Valid {
			service.CategoryID = categoryID.String
		}
		
		if modifiedOn.Valid {
			service.ModifiedOn = modifiedOn.Time
		}
		
		services = append(services, service)
	}
	
	return services, rows.Err()
}

func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

type DaprStateStoreRepository struct {
	client    client.Client
	storeName string
}

func NewDaprStateStoreRepository(daprClient client.Client, storeName string) *DaprStateStoreRepository {
	return &DaprStateStoreRepository{
		client:    daprClient,
		storeName: storeName,
	}
}

func (r *DaprStateStoreRepository) Create(service *Service) error {
	ctx := context.Background()
	
	service.ServiceID = uuid.New().String()
	service.CreatedOn = time.Now()
	service.IsDeleted = false
	
	// Set default category if not provided
	if service.CategoryID == "" {
		defaultCategory, err := r.getDefaultCategory(ctx)
		if err != nil {
			return fmt.Errorf("failed to get default category: %w", err)
		}
		service.CategoryID = defaultCategory.CategoryID
	}
	
	serviceData, err := json.Marshal(service)
	if err != nil {
		return fmt.Errorf("failed to marshal service: %w", err)
	}
	
	err = r.client.SaveState(ctx, r.storeName, service.ServiceID, serviceData, nil)
	if err != nil {
		return fmt.Errorf("failed to save service: %w", err)
	}
	
	// Also save by slug for retrieval
	slugKey := fmt.Sprintf("slug:%s", service.Slug)
	slugData, _ := json.Marshal(map[string]string{"service_id": service.ServiceID})
	err = r.client.SaveState(ctx, r.storeName, slugKey, slugData, nil)
	if err != nil {
		return fmt.Errorf("failed to save service slug mapping: %w", err)
	}
	
	return nil
}

func (r *DaprStateStoreRepository) GetByID(serviceID string) (*Service, error) {
	ctx := context.Background()
	
	item, err := r.client.GetState(ctx, r.storeName, serviceID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}
	
	if len(item.Value) == 0 {
		return nil, fmt.Errorf("service not found")
	}
	
	var service Service
	err = json.Unmarshal(item.Value, &service)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal service: %w", err)
	}
	
	if service.IsDeleted {
		return nil, fmt.Errorf("service not found")
	}
	
	return &service, nil
}

func (r *DaprStateStoreRepository) GetBySlug(slug string) (*Service, error) {
	ctx := context.Background()
	
	slugKey := fmt.Sprintf("slug:%s", slug)
	item, err := r.client.GetState(ctx, r.storeName, slugKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get service by slug: %w", err)
	}
	
	if len(item.Value) == 0 {
		return nil, fmt.Errorf("service not found")
	}
	
	var slugMapping map[string]string
	err = json.Unmarshal(item.Value, &slugMapping)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal slug mapping: %w", err)
	}
	
	serviceID, exists := slugMapping["service_id"]
	if !exists {
		return nil, fmt.Errorf("service not found")
	}
	
	return r.GetByID(serviceID)
}

func (r *DaprStateStoreRepository) Update(service *Service) error {
	ctx := context.Background()
	
	service.ModifiedOn = time.Now()
	
	serviceData, err := json.Marshal(service)
	if err != nil {
		return fmt.Errorf("failed to marshal service: %w", err)
	}
	
	err = r.client.SaveState(ctx, r.storeName, service.ServiceID, serviceData, nil)
	if err != nil {
		return fmt.Errorf("failed to update service: %w", err)
	}
	
	// Update slug mapping if changed
	slugKey := fmt.Sprintf("slug:%s", service.Slug)
	slugData, _ := json.Marshal(map[string]string{"service_id": service.ServiceID})
	err = r.client.SaveState(ctx, r.storeName, slugKey, slugData, nil)
	if err != nil {
		return fmt.Errorf("failed to update service slug mapping: %w", err)
	}
	
	return nil
}

func (r *DaprStateStoreRepository) Delete(serviceID, userID string) error {
	service, err := r.GetByID(serviceID)
	if err != nil {
		return err
	}
	
	service.IsDeleted = true
	service.DeletedOn = time.Now()
	service.DeletedBy = userID
	service.ModifiedOn = time.Now()
	service.ModifiedBy = userID
	
	return r.Update(service)
}

func (r *DaprStateStoreRepository) List(offset, limit int) ([]*Service, error) {
	ctx := context.Background()
	
	// For state store, we'll use a query operation
	query := fmt.Sprintf(`{
		"filter": {
			"EQ": { "is_deleted": false }
		},
		"sort": [
			{ "key": "created_on", "order": "DESC" }
		],
		"page": {
			"limit": %d
		}
	}`, limit)
	
	results, err := r.client.QueryStateAlpha1(ctx, r.storeName, query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query services: %w", err)
	}
	
	var services []*Service
	for _, result := range results.Results {
		var service Service
		err = json.Unmarshal(result.Value, &service)
		if err != nil {
			continue // Skip malformed entries
		}
		if !service.IsDeleted {
			services = append(services, &service)
		}
	}
	
	return services, nil
}

func (r *DaprStateStoreRepository) ListByCategory(categoryID string, offset, limit int) ([]*Service, error) {
	ctx := context.Background()
	
	query := fmt.Sprintf(`{
		"filter": {
			"AND": [
				{ "EQ": { "category_id": "%s" } },
				{ "EQ": { "is_deleted": false } }
			]
		},
		"sort": [
			{ "key": "order_number", "order": "ASC" },
			{ "key": "created_on", "order": "DESC" }
		],
		"page": {
			"limit": %d
		}
	}`, categoryID, limit)
	
	results, err := r.client.QueryStateAlpha1(ctx, r.storeName, query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query services by category: %w", err)
	}
	
	var services []*Service
	for _, result := range results.Results {
		var service Service
		err = json.Unmarshal(result.Value, &service)
		if err != nil {
			continue
		}
		if !service.IsDeleted {
			services = append(services, &service)
		}
	}
	
	return services, nil
}

func (r *DaprStateStoreRepository) ListPublished(offset, limit int) ([]*Service, error) {
	ctx := context.Background()
	
	query := fmt.Sprintf(`{
		"filter": {
			"AND": [
				{ "EQ": { "publishing_status": "published" } },
				{ "EQ": { "is_deleted": false } }
			]
		},
		"sort": [
			{ "key": "order_number", "order": "ASC" },
			{ "key": "created_on", "order": "DESC" }
		],
		"page": {
			"limit": %d
		}
	}`, limit)
	
	results, err := r.client.QueryStateAlpha1(ctx, r.storeName, query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query published services: %w", err)
	}
	
	var services []*Service
	for _, result := range results.Results {
		var service Service
		err = json.Unmarshal(result.Value, &service)
		if err != nil {
			continue
		}
		if !service.IsDeleted {
			services = append(services, &service)
		}
	}
	
	return services, nil
}

func (r *DaprStateStoreRepository) getDefaultCategory(ctx context.Context) (*ServiceCategory, error) {
	// Query for default unassigned category
	query := `{
		"filter": {
			"AND": [
				{ "EQ": { "is_default_unassigned": true } },
				{ "EQ": { "is_deleted": false } }
			]
		}
	}`
	
	results, err := r.client.QueryStateAlpha1(ctx, "categories-store", query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query default category: %w", err)
	}
	
	if len(results.Results) == 0 {
		return nil, fmt.Errorf("no default category found")
	}
	
	var category ServiceCategory
	err = json.Unmarshal(results.Results[0].Value, &category)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal category: %w", err)
	}
	
	return &category, nil
}