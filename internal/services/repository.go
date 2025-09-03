package services

import (
	"context"
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


type servicesRepository struct {
	client    client.Client
	storeName string
}

func NewServicesRepository(daprClient client.Client, storeName string) ServicesRepository {
	return &servicesRepository{
		client:    daprClient,
		storeName: storeName,
	}
}

func (r *servicesRepository) Create(service *Service) error {
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

func (r *servicesRepository) GetByID(serviceID string) (*Service, error) {
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

func (r *servicesRepository) GetBySlug(slug string) (*Service, error) {
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

func (r *servicesRepository) Update(service *Service) error {
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

func (r *servicesRepository) Delete(serviceID, userID string) error {
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

func (r *servicesRepository) List(offset, limit int) ([]*Service, error) {
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
	
	services := make([]*Service, 0)
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

func (r *servicesRepository) ListByCategory(categoryID string, offset, limit int) ([]*Service, error) {
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
	
	services := make([]*Service, 0)
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

func (r *servicesRepository) ListPublished(offset, limit int) ([]*Service, error) {
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
	
	services := make([]*Service, 0)
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

func (r *servicesRepository) getDefaultCategory(ctx context.Context) (*ServiceCategory, error) {
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