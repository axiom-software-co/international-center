package notifications

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// SubscriberRepository provides database operations for notification subscribers
type SubscriberRepository interface {
	CreateSubscriber(ctx context.Context, subscriber *NotificationSubscriber) error
	GetSubscriber(ctx context.Context, subscriberID string) (*NotificationSubscriber, error)
	GetSubscriberByEmail(ctx context.Context, email string) (*NotificationSubscriber, error)
	UpdateSubscriber(ctx context.Context, subscriber *NotificationSubscriber) error
	DeleteSubscriber(ctx context.Context, subscriberID string, deletedBy string) error
	ListSubscribers(ctx context.Context, status *SubscriberStatus, limit, offset int) ([]*NotificationSubscriber, int, error)
	GetSubscribersByEventType(ctx context.Context, eventType EventType) ([]*NotificationSubscriber, error)
	GetActiveSubscribersByPriority(ctx context.Context, priority PriorityThreshold) ([]*NotificationSubscriber, error)
	CheckEmailExists(ctx context.Context, email string, excludeID *string) (bool, error)
	HealthCheck(ctx context.Context) error
}

// PostgreSQLSubscriberRepository implements SubscriberRepository using PostgreSQL
type PostgreSQLSubscriberRepository struct {
	db     *sql.DB
	logger *slog.Logger
	
	// Prepared statements for performance optimization
	createStmt           *sql.Stmt
	getByIDStmt          *sql.Stmt
	getByEmailStmt       *sql.Stmt
	updateStmt           *sql.Stmt
	softDeleteStmt       *sql.Stmt
	listActiveStmt       *sql.Stmt
	getByEventTypeStmt   *sql.Stmt
	getByPriorityStmt    *sql.Stmt
	checkEmailExistsStmt *sql.Stmt
}

// NewPostgreSQLSubscriberRepository creates a new PostgreSQL subscriber repository
func NewPostgreSQLSubscriberRepository(db *sql.DB, logger *slog.Logger) (*PostgreSQLSubscriberRepository, error) {
	repo := &PostgreSQLSubscriberRepository{
		db:     db,
		logger: logger,
	}
	
	// Initialize prepared statements for performance optimization
	if err := repo.initializePreparedStatements(); err != nil {
		return nil, fmt.Errorf("failed to initialize prepared statements: %w", err)
	}
	
	return repo, nil
}

// initializePreparedStatements creates prepared statements for common queries
func (r *PostgreSQLSubscriberRepository) initializePreparedStatements() error {
	var err error
	
	// Create subscriber statement
	createQuery := `
		INSERT INTO notification_subscribers (
			subscriber_id, status, subscriber_name, email, phone,
			event_types, notification_methods, notification_schedule, 
			priority_threshold, notes, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	if r.createStmt, err = r.db.Prepare(createQuery); err != nil {
		return fmt.Errorf("failed to prepare create statement: %w", err)
	}
	
	// Get by ID statement
	getByIDQuery := `
		SELECT subscriber_id, status, subscriber_name, email, phone,
		       event_types, notification_methods, notification_schedule,
		       priority_threshold, notes, created_at, updated_at,
		       created_by, updated_by, is_deleted, deleted_at
		FROM notification_subscribers 
		WHERE subscriber_id = $1 AND is_deleted = false`
	if r.getByIDStmt, err = r.db.Prepare(getByIDQuery); err != nil {
		return fmt.Errorf("failed to prepare getByID statement: %w", err)
	}
	
	// Get by email statement
	getByEmailQuery := `
		SELECT subscriber_id, status, subscriber_name, email, phone,
		       event_types, notification_methods, notification_schedule,
		       priority_threshold, notes, created_at, updated_at,
		       created_by, updated_by, is_deleted, deleted_at
		FROM notification_subscribers 
		WHERE email = $1 AND is_deleted = false`
	if r.getByEmailStmt, err = r.db.Prepare(getByEmailQuery); err != nil {
		return fmt.Errorf("failed to prepare getByEmail statement: %w", err)
	}
	
	// Check email exists statement
	checkEmailQuery := `
		SELECT COUNT(*) > 0 
		FROM notification_subscribers 
		WHERE email = $1 AND is_deleted = false AND ($2::text IS NULL OR subscriber_id != $2)`
	if r.checkEmailExistsStmt, err = r.db.Prepare(checkEmailQuery); err != nil {
		return fmt.Errorf("failed to prepare checkEmailExists statement: %w", err)
	}
	
	return nil
}

// Close cleans up prepared statements and connections
func (r *PostgreSQLSubscriberRepository) Close() error {
	statements := []*sql.Stmt{
		r.createStmt, r.getByIDStmt, r.getByEmailStmt, r.updateStmt,
		r.softDeleteStmt, r.listActiveStmt, r.getByEventTypeStmt,
		r.getByPriorityStmt, r.checkEmailExistsStmt,
	}
	
	for _, stmt := range statements {
		if stmt != nil {
			if err := stmt.Close(); err != nil {
				r.logger.Error("Failed to close prepared statement", "error", err)
			}
		}
	}
	
	return nil
}

// CreateSubscriber creates a new notification subscriber
func (r *PostgreSQLSubscriberRepository) CreateSubscriber(ctx context.Context, subscriber *NotificationSubscriber) error {
	query := `
		INSERT INTO notification_subscribers (
			subscriber_id, status, subscriber_name, email, phone,
			event_types, notification_methods, notification_schedule, 
			priority_threshold, notes, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	// Convert arrays to PostgreSQL array format
	eventTypes := make([]string, len(subscriber.EventTypes))
	for i, et := range subscriber.EventTypes {
		eventTypes[i] = string(et)
	}

	notificationMethods := make([]string, len(subscriber.NotificationMethods))
	for i, nm := range subscriber.NotificationMethods {
		notificationMethods[i] = string(nm)
	}

	_, err := r.db.ExecContext(ctx, query,
		subscriber.SubscriberID,
		string(subscriber.Status),
		subscriber.SubscriberName,
		subscriber.Email,
		subscriber.Phone,
		pq.Array(eventTypes),
		pq.Array(notificationMethods),
		string(subscriber.NotificationSchedule),
		string(subscriber.PriorityThreshold),
		subscriber.Notes,
		subscriber.CreatedBy,
		subscriber.UpdatedBy,
	)

	if err != nil {
		r.logger.Error("Failed to create subscriber",
			"subscriber_id", subscriber.SubscriberID,
			"email", subscriber.Email,
			"error", err)

		if isUniqueViolation(err) {
			return domain.NewValidationError("email address already exists")
		}

		return domain.NewDependencyError("failed to create subscriber", err)
	}

	r.logger.Info("Created new subscriber",
		"subscriber_id", subscriber.SubscriberID,
		"email", subscriber.Email)

	return nil
}

// GetSubscriber retrieves a subscriber by ID
func (r *PostgreSQLSubscriberRepository) GetSubscriber(ctx context.Context, subscriberID string) (*NotificationSubscriber, error) {
	// Validate UUID format
	if _, err := uuid.Parse(subscriberID); err != nil {
		return nil, domain.NewValidationError("invalid subscriber ID format")
	}

	query := `
		SELECT subscriber_id, status, subscriber_name, email, phone,
			   event_types, notification_methods, notification_schedule,
			   priority_threshold, notes, created_at, updated_at,
			   created_by, updated_by, is_deleted, deleted_at
		FROM notification_subscribers 
		WHERE subscriber_id = $1 AND NOT is_deleted
	`

	var subscriber NotificationSubscriber
	var eventTypes pq.StringArray
	var notificationMethods pq.StringArray

	err := r.db.QueryRowContext(ctx, query, subscriberID).Scan(
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
			return nil, domain.NewNotFoundError("subscriber", subscriberID)
		}

		r.logger.Error("Failed to get subscriber",
			"subscriber_id", subscriberID,
			"error", err)

		return nil, domain.NewDependencyError("failed to get subscriber", err)
	}

	// Convert arrays
	subscriber.EventTypes = make([]EventType, len(eventTypes))
	for i, et := range eventTypes {
		subscriber.EventTypes[i] = EventType(et)
	}

	subscriber.NotificationMethods = make([]NotificationMethod, len(notificationMethods))
	for i, nm := range notificationMethods {
		subscriber.NotificationMethods[i] = NotificationMethod(nm)
	}

	return &subscriber, nil
}

// GetSubscriberByEmail retrieves a subscriber by email address
func (r *PostgreSQLSubscriberRepository) GetSubscriberByEmail(ctx context.Context, email string) (*NotificationSubscriber, error) {
	query := `
		SELECT subscriber_id, status, subscriber_name, email, phone,
			   event_types, notification_methods, notification_schedule,
			   priority_threshold, notes, created_at, updated_at,
			   created_by, updated_by, is_deleted, deleted_at
		FROM notification_subscribers 
		WHERE email = $1 AND NOT is_deleted
	`

	var subscriber NotificationSubscriber
	var eventTypes pq.StringArray
	var notificationMethods pq.StringArray

	err := r.db.QueryRowContext(ctx, query, email).Scan(
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
			return nil, domain.NewNotFoundError("subscriber", email)
		}

		return nil, domain.NewDependencyError("failed to get subscriber by email", err)
	}

	// Convert arrays
	subscriber.EventTypes = make([]EventType, len(eventTypes))
	for i, et := range eventTypes {
		subscriber.EventTypes[i] = EventType(et)
	}

	subscriber.NotificationMethods = make([]NotificationMethod, len(notificationMethods))
	for i, nm := range notificationMethods {
		subscriber.NotificationMethods[i] = NotificationMethod(nm)
	}

	return &subscriber, nil
}

// UpdateSubscriber updates an existing subscriber
func (r *PostgreSQLSubscriberRepository) UpdateSubscriber(ctx context.Context, subscriber *NotificationSubscriber) error {
	query := `
		UPDATE notification_subscribers 
		SET status = $2, subscriber_name = $3, email = $4, phone = $5,
			event_types = $6, notification_methods = $7, 
			notification_schedule = $8, priority_threshold = $9,
			notes = $10, updated_by = $11, updated_at = CURRENT_TIMESTAMP
		WHERE subscriber_id = $1 AND NOT is_deleted
	`

	// Convert arrays
	eventTypes := make([]string, len(subscriber.EventTypes))
	for i, et := range subscriber.EventTypes {
		eventTypes[i] = string(et)
	}

	notificationMethods := make([]string, len(subscriber.NotificationMethods))
	for i, nm := range subscriber.NotificationMethods {
		notificationMethods[i] = string(nm)
	}

	result, err := r.db.ExecContext(ctx, query,
		subscriber.SubscriberID,
		string(subscriber.Status),
		subscriber.SubscriberName,
		subscriber.Email,
		subscriber.Phone,
		pq.Array(eventTypes),
		pq.Array(notificationMethods),
		string(subscriber.NotificationSchedule),
		string(subscriber.PriorityThreshold),
		subscriber.Notes,
		subscriber.UpdatedBy,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return domain.NewValidationError("email address already exists")
		}

		return domain.NewDependencyError("failed to update subscriber", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return domain.NewDependencyError("failed to check update result", err)
	}

	if rowsAffected == 0 {
		return domain.NewNotFoundError("subscriber", subscriber.SubscriberID)
	}

	r.logger.Info("Updated subscriber",
		"subscriber_id", subscriber.SubscriberID,
		"email", subscriber.Email)

	return nil
}

// DeleteSubscriber soft deletes a subscriber
func (r *PostgreSQLSubscriberRepository) DeleteSubscriber(ctx context.Context, subscriberID string, deletedBy string) error {
	query := `
		UPDATE notification_subscribers 
		SET is_deleted = true, deleted_at = CURRENT_TIMESTAMP, updated_by = $2
		WHERE subscriber_id = $1 AND NOT is_deleted
	`

	result, err := r.db.ExecContext(ctx, query, subscriberID, deletedBy)
	if err != nil {
		return domain.NewDependencyError("failed to delete subscriber", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return domain.NewDependencyError("failed to check delete result", err)
	}

	if rowsAffected == 0 {
		return domain.NewNotFoundError("subscriber", subscriberID)
	}

	r.logger.Info("Deleted subscriber", "subscriber_id", subscriberID)
	return nil
}

// ListSubscribers retrieves a paginated list of subscribers
func (r *PostgreSQLSubscriberRepository) ListSubscribers(ctx context.Context, status *SubscriberStatus, limit, offset int) ([]*NotificationSubscriber, int, error) {
	if limit <= 0 {
		return nil, 0, domain.NewValidationError("invalid limit parameter")
	}

	if offset < 0 {
		return nil, 0, domain.NewValidationError("invalid offset parameter")
	}

	// Build query with optional status filter
	whereClause := "WHERE NOT is_deleted"
	args := []interface{}{limit, offset}
	argIndex := 3

	if status != nil {
		whereClause += " AND status = $3"
		args = append([]interface{}{*status}, args...)
		argIndex = 4
	}

	query := fmt.Sprintf(`
		SELECT subscriber_id, status, subscriber_name, email, phone,
			   event_types, notification_methods, notification_schedule,
			   priority_threshold, notes, created_at, updated_at,
			   created_by, updated_by, is_deleted, deleted_at
		FROM notification_subscribers 
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex-2, argIndex-1)

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM notification_subscribers %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		return nil, 0, domain.NewDependencyError("failed to count subscribers", err)
	}

	// Get subscribers
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, domain.NewDependencyError("failed to list subscribers", err)
	}
	defer rows.Close()

	var subscribers []*NotificationSubscriber
	for rows.Next() {
		var subscriber NotificationSubscriber
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
			return nil, 0, domain.NewDependencyError("failed to scan subscriber", err)
		}

		// Convert arrays
		subscriber.EventTypes = make([]EventType, len(eventTypes))
		for i, et := range eventTypes {
			subscriber.EventTypes[i] = EventType(et)
		}

		subscriber.NotificationMethods = make([]NotificationMethod, len(notificationMethods))
		for i, nm := range notificationMethods {
			subscriber.NotificationMethods[i] = NotificationMethod(nm)
		}

		subscribers = append(subscribers, &subscriber)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, domain.NewDependencyError("failed to iterate subscribers", err)
	}

	return subscribers, total, nil
}

// GetSubscribersByEventType retrieves active subscribers for a specific event type
func (r *PostgreSQLSubscriberRepository) GetSubscribersByEventType(ctx context.Context, eventType EventType) ([]*NotificationSubscriber, error) {
	query := `
		SELECT subscriber_id, status, subscriber_name, email, phone,
			   event_types, notification_methods, notification_schedule,
			   priority_threshold, notes, created_at, updated_at,
			   created_by, updated_by, is_deleted, deleted_at
		FROM notification_subscribers 
		WHERE status = 'active' 
		  AND NOT is_deleted
		  AND $1 = ANY(event_types)
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, string(eventType))
	if err != nil {
		return nil, domain.NewDependencyError("failed to get subscribers by event type", err)
	}
	defer rows.Close()

	var subscribers []*NotificationSubscriber
	for rows.Next() {
		var subscriber NotificationSubscriber
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
			return nil, domain.NewDependencyError("failed to scan subscriber", err)
		}

		// Convert arrays
		subscriber.EventTypes = make([]EventType, len(eventTypes))
		for i, et := range eventTypes {
			subscriber.EventTypes[i] = EventType(et)
		}

		subscriber.NotificationMethods = make([]NotificationMethod, len(notificationMethods))
		for i, nm := range notificationMethods {
			subscriber.NotificationMethods[i] = NotificationMethod(nm)
		}

		subscribers = append(subscribers, &subscriber)
	}

	if err = rows.Err(); err != nil {
		return nil, domain.NewDependencyError("failed to iterate subscribers", err)
	}

	return subscribers, nil
}

// GetActiveSubscribersByPriority retrieves active subscribers by priority threshold
func (r *PostgreSQLSubscriberRepository) GetActiveSubscribersByPriority(ctx context.Context, priority PriorityThreshold) ([]*NotificationSubscriber, error) {
	// Priority levels: low=1, medium=2, high=3, urgent=4
	priorityLevels := map[PriorityThreshold]int{
		PriorityLow:    1,
		PriorityMedium: 2,
		PriorityHigh:   3,
		PriorityUrgent: 4,
	}

	eventLevel, exists := priorityLevels[priority]
	if !exists {
		return nil, domain.NewValidationError("invalid priority threshold")
	}

	query := `
		SELECT subscriber_id, status, subscriber_name, email, phone,
			   event_types, notification_methods, notification_schedule,
			   priority_threshold, notes, created_at, updated_at,
			   created_by, updated_by, is_deleted, deleted_at
		FROM notification_subscribers 
		WHERE status = 'active' 
		  AND NOT is_deleted
		  AND CASE priority_threshold
		      WHEN 'low' THEN 1
		      WHEN 'medium' THEN 2
		      WHEN 'high' THEN 3
		      WHEN 'urgent' THEN 4
		      ELSE 0
		      END <= $1
		ORDER BY priority_threshold DESC, created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, eventLevel)
	if err != nil {
		return nil, domain.NewDependencyError("failed to get subscribers by priority", err)
	}
	defer rows.Close()

	var subscribers []*NotificationSubscriber
	for rows.Next() {
		var subscriber NotificationSubscriber
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
			return nil, domain.NewDependencyError("failed to scan subscriber", err)
		}

		// Convert arrays
		subscriber.EventTypes = make([]EventType, len(eventTypes))
		for i, et := range eventTypes {
			subscriber.EventTypes[i] = EventType(et)
		}

		subscriber.NotificationMethods = make([]NotificationMethod, len(notificationMethods))
		for i, nm := range notificationMethods {
			subscriber.NotificationMethods[i] = NotificationMethod(nm)
		}

		subscribers = append(subscribers, &subscriber)
	}

	if err = rows.Err(); err != nil {
		return nil, domain.NewDependencyError("failed to iterate subscribers", err)
	}

	return subscribers, nil
}

// CheckEmailExists checks if an email address already exists
func (r *PostgreSQLSubscriberRepository) CheckEmailExists(ctx context.Context, email string, excludeID *string) (bool, error) {
	query := "SELECT COUNT(*) FROM notification_subscribers WHERE email = $1 AND NOT is_deleted"
	args := []interface{}{email}

	if excludeID != nil {
		query += " AND subscriber_id != $2"
		args = append(args, *excludeID)
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return false, domain.NewDependencyError("failed to check email existence", err)
	}

	return count > 0, nil
}

// HealthCheck performs a health check on the database connection
func (r *PostgreSQLSubscriberRepository) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := r.db.PingContext(ctx)
	if err != nil {
		return domain.NewDependencyError("database health check failed", err)
	}

	return nil
}

// Helper functions

// isUniqueViolation checks if the error is a unique constraint violation
func isUniqueViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505" // unique_violation
	}
	return false
}