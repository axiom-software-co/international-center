package gateway

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/database"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// PostgreSQLSubscriberRepository implements SubscriberRepository using PostgreSQL
type PostgreSQLSubscriberRepository struct {
	db *sql.DB
}

// NewPostgreSQLSubscriberRepository creates a new PostgreSQL subscriber repository
func NewPostgreSQLSubscriberRepository(db *sql.DB) *PostgreSQLSubscriberRepository {
	return &PostgreSQLSubscriberRepository{
		db: db,
	}
}

// CreateSubscriber creates a new notification subscriber
func (r *PostgreSQLSubscriberRepository) CreateSubscriber(ctx context.Context, subscriber *NotificationSubscriber) error {
	if subscriber == nil {
		return domain.NewValidationError("subscriber cannot be nil", nil)
	}

	// Validate subscriber ID format
	if _, err := uuid.Parse(subscriber.SubscriberID); err != nil {
		return domain.NewValidationError("invalid subscriber ID format", err)
	}

	// Check for duplicate email
	exists, err := r.CheckEmailExists(ctx, subscriber.Email, nil)
	if err != nil {
		return fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return domain.NewConflictError("duplicate email address", nil)
	}

	query := `
		INSERT INTO notification_subscribers (
			subscriber_id, status, subscriber_name, email, phone, event_types, 
			notification_methods, notification_schedule, priority_threshold, 
			notes, created_at, updated_at, created_by, updated_by, is_deleted
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`

	_, err = r.db.ExecContext(
		ctx,
		query,
		subscriber.SubscriberID,
		subscriber.Status,
		subscriber.SubscriberName,
		subscriber.Email,
		subscriber.Phone,
		pq.Array(subscriber.EventTypes),
		pq.Array(subscriber.NotificationMethods),
		subscriber.NotificationSchedule,
		subscriber.PriorityThreshold,
		subscriber.Notes,
		subscriber.CreatedAt,
		subscriber.UpdatedAt,
		subscriber.CreatedBy,
		subscriber.UpdatedBy,
		subscriber.IsDeleted,
	)

	if err != nil {
		if database.IsDuplicateKeyError(err) {
			return domain.NewConflictError("duplicate email address", err)
		}
		return fmt.Errorf("failed to create subscriber: %w", err)
	}

	return nil
}

// GetSubscriber retrieves a subscriber by ID
func (r *PostgreSQLSubscriberRepository) GetSubscriber(ctx context.Context, subscriberID string) (*NotificationSubscriber, error) {
	if subscriberID == "" {
		return nil, domain.NewValidationError("subscriber ID cannot be empty", nil)
	}

	// Validate subscriber ID format
	if _, err := uuid.Parse(subscriberID); err != nil {
		return nil, domain.NewValidationError("invalid subscriber ID format", err)
	}

	query := `
		SELECT subscriber_id, status, subscriber_name, email, phone, event_types, 
			   notification_methods, notification_schedule, priority_threshold, 
			   notes, created_at, updated_at, created_by, updated_by, is_deleted, deleted_at
		FROM notification_subscribers 
		WHERE subscriber_id = $1 AND is_deleted = false
	`

	row := r.db.QueryRowContext(ctx, query, subscriberID)

	subscriber := &NotificationSubscriber{}
	var eventTypes pq.StringArray
	var notificationMethods pq.StringArray

	err := row.Scan(
		&subscriber.SubscriberID,
		&subscriber.Status,
		&subscriber.SubscriberName,
		&subscriber.Email,
		&subscriber.Phone,
		&eventTypes,
		&notificationMethods,
		&subscriber.NotificationSchedule,
		&subscriber.PriorityThreshold,
		&subscriber.Notes,
		&subscriber.CreatedAt,
		&subscriber.UpdatedAt,
		&subscriber.CreatedBy,
		&subscriber.UpdatedBy,
		&subscriber.IsDeleted,
		&subscriber.DeletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewNotFoundError("subscriber not found", nil)
		}
		return nil, fmt.Errorf("failed to get subscriber: %w", err)
	}

	// Convert arrays to slices
	subscriber.EventTypes = make([]EventType, len(eventTypes))
	for i, et := range eventTypes {
		subscriber.EventTypes[i] = EventType(et)
	}

	subscriber.NotificationMethods = make([]NotificationMethod, len(notificationMethods))
	for i, nm := range notificationMethods {
		subscriber.NotificationMethods[i] = NotificationMethod(nm)
	}

	return subscriber, nil
}

// GetSubscriberByEmail retrieves a subscriber by email address
func (r *PostgreSQLSubscriberRepository) GetSubscriberByEmail(ctx context.Context, email string) (*NotificationSubscriber, error) {
	if email == "" {
		return nil, domain.NewValidationError("email cannot be empty", nil)
	}

	query := `
		SELECT subscriber_id, status, subscriber_name, email, phone, event_types, 
			   notification_methods, notification_schedule, priority_threshold, 
			   notes, created_at, updated_at, created_by, updated_by, is_deleted, deleted_at
		FROM notification_subscribers 
		WHERE email = $1 AND is_deleted = false
	`

	row := r.db.QueryRowContext(ctx, query, email)

	subscriber := &NotificationSubscriber{}
	var eventTypes pq.StringArray
	var notificationMethods pq.StringArray

	err := row.Scan(
		&subscriber.SubscriberID,
		&subscriber.Status,
		&subscriber.SubscriberName,
		&subscriber.Email,
		&subscriber.Phone,
		&eventTypes,
		&notificationMethods,
		&subscriber.NotificationSchedule,
		&subscriber.PriorityThreshold,
		&subscriber.Notes,
		&subscriber.CreatedAt,
		&subscriber.UpdatedAt,
		&subscriber.CreatedBy,
		&subscriber.UpdatedBy,
		&subscriber.IsDeleted,
		&subscriber.DeletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewNotFoundError("subscriber not found", nil)
		}
		return nil, fmt.Errorf("failed to get subscriber by email: %w", err)
	}

	// Convert arrays to slices
	subscriber.EventTypes = make([]EventType, len(eventTypes))
	for i, et := range eventTypes {
		subscriber.EventTypes[i] = EventType(et)
	}

	subscriber.NotificationMethods = make([]NotificationMethod, len(notificationMethods))
	for i, nm := range notificationMethods {
		subscriber.NotificationMethods[i] = NotificationMethod(nm)
	}

	return subscriber, nil
}

// UpdateSubscriber updates an existing subscriber
func (r *PostgreSQLSubscriberRepository) UpdateSubscriber(ctx context.Context, subscriber *NotificationSubscriber) error {
	if subscriber == nil {
		return domain.NewValidationError("subscriber cannot be nil", nil)
	}

	// Validate subscriber ID format
	if _, err := uuid.Parse(subscriber.SubscriberID); err != nil {
		return domain.NewValidationError("invalid subscriber ID format", err)
	}

	// Check for duplicate email (excluding this subscriber)
	exists, err := r.CheckEmailExists(ctx, subscriber.Email, &subscriber.SubscriberID)
	if err != nil {
		return fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return domain.NewConflictError("duplicate email address", nil)
	}

	query := `
		UPDATE notification_subscribers 
		SET status = $2, subscriber_name = $3, email = $4, phone = $5, event_types = $6, 
			notification_methods = $7, notification_schedule = $8, priority_threshold = $9, 
			notes = $10, updated_at = $11, updated_by = $12
		WHERE subscriber_id = $1 AND is_deleted = false
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		subscriber.SubscriberID,
		subscriber.Status,
		subscriber.SubscriberName,
		subscriber.Email,
		subscriber.Phone,
		pq.Array(subscriber.EventTypes),
		pq.Array(subscriber.NotificationMethods),
		subscriber.NotificationSchedule,
		subscriber.PriorityThreshold,
		subscriber.Notes,
		subscriber.UpdatedAt,
		subscriber.UpdatedBy,
	)

	if err != nil {
		if database.IsDuplicateKeyError(err) {
			return domain.NewConflictError("duplicate email address", err)
		}
		return fmt.Errorf("failed to update subscriber: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.NewNotFoundError("subscriber not found", nil)
	}

	return nil
}

// DeleteSubscriber soft deletes a subscriber
func (r *PostgreSQLSubscriberRepository) DeleteSubscriber(ctx context.Context, subscriberID string, deletedBy string) error {
	if subscriberID == "" {
		return domain.NewValidationError("subscriber ID cannot be empty", nil)
	}

	if deletedBy == "" {
		return domain.NewValidationError("deleted by cannot be empty", nil)
	}

	// Validate subscriber ID format
	if _, err := uuid.Parse(subscriberID); err != nil {
		return domain.NewValidationError("invalid subscriber ID format", err)
	}

	query := `
		UPDATE notification_subscribers 
		SET is_deleted = true, deleted_at = $2, updated_by = $3, updated_at = $2
		WHERE subscriber_id = $1 AND is_deleted = false
	`

	result, err := r.db.ExecContext(ctx, query, subscriberID, time.Now().UTC(), deletedBy)
	if err != nil {
		return fmt.Errorf("failed to delete subscriber: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.NewNotFoundError("subscriber not found", nil)
	}

	return nil
}

// ListSubscribers retrieves subscribers with pagination and filtering
func (r *PostgreSQLSubscriberRepository) ListSubscribers(ctx context.Context, status *SubscriberStatus, limit, offset int) ([]*NotificationSubscriber, int, error) {
	if limit < 0 {
		return nil, 0, domain.NewValidationError("invalid limit parameter", nil)
	}

	if offset < 0 {
		return nil, 0, domain.NewValidationError("invalid offset parameter", nil)
	}

	// Build query with optional status filter
	baseQuery := `
		FROM notification_subscribers 
		WHERE is_deleted = false
	`
	args := []interface{}{}
	argCount := 0

	if status != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, *status)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get subscriber count: %w", err)
	}

	// Get paginated results
	selectQuery := `
		SELECT subscriber_id, status, subscriber_name, email, phone, event_types, 
			   notification_methods, notification_schedule, priority_threshold, 
			   notes, created_at, updated_at, created_by, updated_by, is_deleted, deleted_at
	` + baseQuery + `
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`

	argCount++
	limitArg := argCount
	argCount++
	offsetArg := argCount

	args = append(args, limit, offset)
	selectQuery = fmt.Sprintf(selectQuery, limitArg, offsetArg)

	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list subscribers: %w", err)
	}
	defer rows.Close()

	var subscribers []*NotificationSubscriber

	for rows.Next() {
		subscriber := &NotificationSubscriber{}
		var eventTypes pq.StringArray
		var notificationMethods pq.StringArray

		err := rows.Scan(
			&subscriber.SubscriberID,
			&subscriber.Status,
			&subscriber.SubscriberName,
			&subscriber.Email,
			&subscriber.Phone,
			&eventTypes,
			&notificationMethods,
			&subscriber.NotificationSchedule,
			&subscriber.PriorityThreshold,
			&subscriber.Notes,
			&subscriber.CreatedAt,
			&subscriber.UpdatedAt,
			&subscriber.CreatedBy,
			&subscriber.UpdatedBy,
			&subscriber.IsDeleted,
			&subscriber.DeletedAt,
		)

		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan subscriber: %w", err)
		}

		// Convert arrays to slices
		subscriber.EventTypes = make([]EventType, len(eventTypes))
		for i, et := range eventTypes {
			subscriber.EventTypes[i] = EventType(et)
		}

		subscriber.NotificationMethods = make([]NotificationMethod, len(notificationMethods))
		for i, nm := range notificationMethods {
			subscriber.NotificationMethods[i] = NotificationMethod(nm)
		}

		subscribers = append(subscribers, subscriber)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate subscribers: %w", err)
	}

	return subscribers, total, nil
}

// GetSubscribersByEventType retrieves subscribers for a specific event type
func (r *PostgreSQLSubscriberRepository) GetSubscribersByEventType(ctx context.Context, eventType EventType) ([]*NotificationSubscriber, error) {
	query := `
		SELECT subscriber_id, status, subscriber_name, email, phone, event_types, 
			   notification_methods, notification_schedule, priority_threshold, 
			   notes, created_at, updated_at, created_by, updated_by, is_deleted, deleted_at
		FROM notification_subscribers 
		WHERE is_deleted = false 
		  AND status = 'active'
		  AND $1 = ANY(event_types)
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, eventType)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscribers by event type: %w", err)
	}
	defer rows.Close()

	var subscribers []*NotificationSubscriber

	for rows.Next() {
		subscriber := &NotificationSubscriber{}
		var eventTypes pq.StringArray
		var notificationMethods pq.StringArray

		err := rows.Scan(
			&subscriber.SubscriberID,
			&subscriber.Status,
			&subscriber.SubscriberName,
			&subscriber.Email,
			&subscriber.Phone,
			&eventTypes,
			&notificationMethods,
			&subscriber.NotificationSchedule,
			&subscriber.PriorityThreshold,
			&subscriber.Notes,
			&subscriber.CreatedAt,
			&subscriber.UpdatedAt,
			&subscriber.CreatedBy,
			&subscriber.UpdatedBy,
			&subscriber.IsDeleted,
			&subscriber.DeletedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan subscriber: %w", err)
		}

		// Convert arrays to slices
		subscriber.EventTypes = make([]EventType, len(eventTypes))
		for i, et := range eventTypes {
			subscriber.EventTypes[i] = EventType(et)
		}

		subscriber.NotificationMethods = make([]NotificationMethod, len(notificationMethods))
		for i, nm := range notificationMethods {
			subscriber.NotificationMethods[i] = NotificationMethod(nm)
		}

		subscribers = append(subscribers, subscriber)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate subscribers: %w", err)
	}

	return subscribers, nil
}

// GetActiveSubscribersByPriority retrieves active subscribers for a priority level
func (r *PostgreSQLSubscriberRepository) GetActiveSubscribersByPriority(ctx context.Context, priority PriorityThreshold) ([]*NotificationSubscriber, error) {
	query := `
		SELECT subscriber_id, status, subscriber_name, email, phone, event_types, 
			   notification_methods, notification_schedule, priority_threshold, 
			   notes, created_at, updated_at, created_by, updated_by, is_deleted, deleted_at
		FROM notification_subscribers 
		WHERE is_deleted = false 
		  AND status = 'active'
		  AND (
		    priority_threshold = $1 OR
		    (priority_threshold = 'low' AND $1 IN ('medium', 'high', 'urgent')) OR
		    (priority_threshold = 'medium' AND $1 IN ('high', 'urgent')) OR
		    (priority_threshold = 'high' AND $1 = 'urgent')
		  )
		ORDER BY priority_threshold DESC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, priority)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscribers by priority: %w", err)
	}
	defer rows.Close()

	var subscribers []*NotificationSubscriber

	for rows.Next() {
		subscriber := &NotificationSubscriber{}
		var eventTypes pq.StringArray
		var notificationMethods pq.StringArray

		err := rows.Scan(
			&subscriber.SubscriberID,
			&subscriber.Status,
			&subscriber.SubscriberName,
			&subscriber.Email,
			&subscriber.Phone,
			&eventTypes,
			&notificationMethods,
			&subscriber.NotificationSchedule,
			&subscriber.PriorityThreshold,
			&subscriber.Notes,
			&subscriber.CreatedAt,
			&subscriber.UpdatedAt,
			&subscriber.CreatedBy,
			&subscriber.UpdatedBy,
			&subscriber.IsDeleted,
			&subscriber.DeletedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan subscriber: %w", err)
		}

		// Convert arrays to slices
		subscriber.EventTypes = make([]EventType, len(eventTypes))
		for i, et := range eventTypes {
			subscriber.EventTypes[i] = EventType(et)
		}

		subscriber.NotificationMethods = make([]NotificationMethod, len(notificationMethods))
		for i, nm := range notificationMethods {
			subscriber.NotificationMethods[i] = NotificationMethod(nm)
		}

		subscribers = append(subscribers, subscriber)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate subscribers: %w", err)
	}

	return subscribers, nil
}

// CheckEmailExists checks if an email address already exists
func (r *PostgreSQLSubscriberRepository) CheckEmailExists(ctx context.Context, email string, excludeID *string) (bool, error) {
	if email == "" {
		return false, domain.NewValidationError("email cannot be empty", nil)
	}

	query := `
		SELECT COUNT(*) 
		FROM notification_subscribers 
		WHERE email = $1 AND is_deleted = false
	`
	args := []interface{}{email}

	if excludeID != nil {
		query += " AND subscriber_id != $2"
		args = append(args, *excludeID)
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return count > 0, nil
}