package services

import (
	"database/sql"
	"time"
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
		SELECT service_id, title, description, slug, COALESCE(category_id, ''), 
			   delivery_mode, publishing_status, is_deleted, created_on, 
			   COALESCE(modified_on, '0001-01-01'::timestamptz), COALESCE(modified_by, '')
		FROM services 
		WHERE service_id = $1 AND is_deleted = false`
	
	service := &Service{}
	var modifiedOn time.Time
	
	err := r.db.QueryRow(query, serviceID).Scan(
		&service.ServiceID,
		&service.Title,
		&service.Description,
		&service.Slug,
		&service.CategoryID,
		&service.DeliveryMode,
		&service.PublishingStatus,
		&service.IsDeleted,
		&service.CreatedOn,
		&modifiedOn,
		&service.ModifiedBy,
	)
	
	if !modifiedOn.IsZero() && modifiedOn.Year() > 1 {
		service.ModifiedOn = modifiedOn
	}
	
	return service, err
}

func (r *PostgreSQLServicesRepository) GetBySlug(slug string) (*Service, error) {
	query := `
		SELECT service_id, title, description, slug, COALESCE(category_id, ''), 
			   delivery_mode, publishing_status, is_deleted, created_on, 
			   COALESCE(modified_on, '0001-01-01'::timestamptz), COALESCE(modified_by, '')
		FROM services 
		WHERE slug = $1 AND is_deleted = false`
	
	service := &Service{}
	var modifiedOn time.Time
	
	err := r.db.QueryRow(query, slug).Scan(
		&service.ServiceID,
		&service.Title,
		&service.Description,
		&service.Slug,
		&service.CategoryID,
		&service.DeliveryMode,
		&service.PublishingStatus,
		&service.IsDeleted,
		&service.CreatedOn,
		&modifiedOn,
		&service.ModifiedBy,
	)
	
	if !modifiedOn.IsZero() && modifiedOn.Year() > 1 {
		service.ModifiedOn = modifiedOn
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
		SELECT service_id, title, description, slug, COALESCE(category_id, ''), 
			   delivery_mode, publishing_status, is_deleted, created_on, 
			   COALESCE(modified_on, '0001-01-01'::timestamptz), COALESCE(modified_by, '')
		FROM services 
		WHERE is_deleted = false
		ORDER BY created_on DESC
		LIMIT $1 OFFSET $2`
	
	return r.scanServices(query, limit, offset)
}

func (r *PostgreSQLServicesRepository) ListByCategory(categoryID string, offset, limit int) ([]*Service, error) {
	query := `
		SELECT service_id, title, description, slug, COALESCE(category_id, ''), 
			   delivery_mode, publishing_status, is_deleted, created_on, 
			   COALESCE(modified_on, '0001-01-01'::timestamptz), COALESCE(modified_by, '')
		FROM services 
		WHERE category_id = $1 AND is_deleted = false
		ORDER BY order_number ASC, created_on DESC
		LIMIT $2 OFFSET $3`
	
	return r.scanServices(query, categoryID, limit, offset)
}

func (r *PostgreSQLServicesRepository) ListPublished(offset, limit int) ([]*Service, error) {
	query := `
		SELECT service_id, title, description, slug, COALESCE(category_id, ''), 
			   delivery_mode, publishing_status, is_deleted, created_on, 
			   COALESCE(modified_on, '0001-01-01'::timestamptz), COALESCE(modified_by, '')
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
		var modifiedOn time.Time
		
		err := rows.Scan(
			&service.ServiceID,
			&service.Title,
			&service.Description,
			&service.Slug,
			&service.CategoryID,
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
		
		if !modifiedOn.IsZero() && modifiedOn.Year() > 1 {
			service.ModifiedOn = modifiedOn
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