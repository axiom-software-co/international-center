package content

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ContentRepository interface {
	Create(ctx context.Context, content *Content) error
	GetByID(ctx context.Context, contentID string) (*Content, error)
	Update(ctx context.Context, content *Content) error
	Delete(ctx context.Context, contentID, userID string) error
	List(ctx context.Context, offset, limit int) ([]*Content, error)
	ListByCategory(ctx context.Context, contentCategory ContentCategory, offset, limit int) ([]*Content, error)
	ListByTags(ctx context.Context, tags []string, offset, limit int) ([]*Content, error)
	ListPublished(ctx context.Context, offset, limit int) ([]*Content, error)
	ListByType(ctx context.Context, contentCategory ContentCategory, offset, limit int) ([]*Content, error)
}

type MongoContentRepository struct {
	collection *mongo.Collection
}

func NewMongoContentRepository(db *mongo.Database) *MongoContentRepository {
	return &MongoContentRepository{
		collection: db.Collection("content"),
	}
}

func (r *MongoContentRepository) Create(ctx context.Context, content *Content) error {
	content.CreatedOn = time.Now()
	content.IsDeleted = false
	
	result, err := r.collection.InsertOne(ctx, content)
	if err != nil {
		return err
	}
	
	content.ContentID = result.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (r *MongoContentRepository) GetByID(ctx context.Context, contentID string) (*Content, error) {
	objectID, err := primitive.ObjectIDFromHex(contentID)
	if err != nil {
		return nil, err
	}
	
	filter := bson.M{
		"_id":        objectID,
		"is_deleted": false,
	}
	
	var content Content
	err = r.collection.FindOne(ctx, filter).Decode(&content)
	if err != nil {
		return nil, err
	}
	
	return &content, nil
}


func (r *MongoContentRepository) Update(ctx context.Context, content *Content) error {
	objectID, err := primitive.ObjectIDFromHex(content.ContentID)
	if err != nil {
		return err
	}
	
	now := time.Now()
	content.ModifiedOn = &now
	
	filter := bson.M{
		"_id":        objectID,
		"is_deleted": false,
	}
	
	update := bson.M{
		"$set": content,
	}
	
	_, err = r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *MongoContentRepository) Delete(ctx context.Context, contentID, userID string) error {
	objectID, err := primitive.ObjectIDFromHex(contentID)
	if err != nil {
		return err
	}
	
	filter := bson.M{
		"_id":        objectID,
		"is_deleted": false,
	}
	
	update := bson.M{
		"$set": bson.M{
			"is_deleted":  true,
			"deleted_on":  time.Now(),
			"deleted_by":  userID,
			"modified_on": time.Now(),
			"modified_by": userID,
		},
	}
	
	_, err = r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *MongoContentRepository) List(ctx context.Context, offset, limit int) ([]*Content, error) {
	filter := bson.M{"is_deleted": false}
	
	findOptions := options.Find()
	findOptions.SetSkip(int64(offset))
	findOptions.SetLimit(int64(limit))
	findOptions.SetSort(bson.M{"created_on": -1})
	
	return r.findContents(ctx, filter, findOptions)
}

func (r *MongoContentRepository) ListByCategory(ctx context.Context, contentCategory ContentCategory, offset, limit int) ([]*Content, error) {
	filter := bson.M{
		"content_category": contentCategory,
		"is_deleted":       false,
	}
	
	findOptions := options.Find()
	findOptions.SetSkip(int64(offset))
	findOptions.SetLimit(int64(limit))
	findOptions.SetSort(bson.M{"created_on": -1})
	
	return r.findContents(ctx, filter, findOptions)
}

func (r *MongoContentRepository) ListByTags(ctx context.Context, tags []string, offset, limit int) ([]*Content, error) {
	filter := bson.M{
		"tags": bson.M{
			"$in": tags,
		},
		"is_deleted": false,
	}
	
	findOptions := options.Find()
	findOptions.SetSkip(int64(offset))
	findOptions.SetLimit(int64(limit))
	findOptions.SetSort(bson.M{"created_on": -1})
	
	return r.findContents(ctx, filter, findOptions)
}

func (r *MongoContentRepository) ListPublished(ctx context.Context, offset, limit int) ([]*Content, error) {
	filter := bson.M{
		"upload_status": UploadStatusAvailable,
		"is_deleted":        false,
	}
	
	findOptions := options.Find()
	findOptions.SetSkip(int64(offset))
	findOptions.SetLimit(int64(limit))
	findOptions.SetSort(bson.M{"created_on": -1})
	
	return r.findContents(ctx, filter, findOptions)
}

func (r *MongoContentRepository) ListByType(ctx context.Context, contentCategory ContentCategory, offset, limit int) ([]*Content, error) {
	filter := bson.M{
		"content_category": contentCategory,
		"is_deleted":       false,
	}
	
	findOptions := options.Find()
	findOptions.SetSkip(int64(offset))
	findOptions.SetLimit(int64(limit))
	findOptions.SetSort(bson.M{"created_on": -1})
	
	return r.findContents(ctx, filter, findOptions)
}

func (r *MongoContentRepository) findContents(ctx context.Context, filter bson.M, findOptions *options.FindOptions) ([]*Content, error) {
	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var contents []*Content
	for cursor.Next(ctx) {
		var content Content
		if err := cursor.Decode(&content); err != nil {
			return nil, err
		}
		contents = append(contents, &content)
	}
	
	return contents, cursor.Err()
}

type PostgreSQLContentRepository struct {
	db *sql.DB
}

func NewPostgreSQLContentRepository(db *sql.DB) *PostgreSQLContentRepository {
	return &PostgreSQLContentRepository{db: db}
}

func (r *PostgreSQLContentRepository) Create(ctx context.Context, content *Content) error {
	contentUUID := uuid.New()
	content.ContentID = contentUUID.String()
	content.CreatedOn = time.Now()
	content.IsDeleted = false

	query := `
		INSERT INTO content (
			content_id, original_filename, file_size, mime_type, content_hash,
			storage_path, upload_status, alt_text, description, tags,
			content_category, access_level, upload_correlation_id, 
			processing_attempts, created_on, created_by, is_deleted
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`

	_, err := r.db.ExecContext(
		ctx, query,
		content.ContentID,
		content.OriginalFilename,
		content.FileSize,
		content.MimeType,
		content.ContentHash,
		content.StoragePath,
		content.UploadStatus,
		content.AltText,
		content.Description,
		pq.Array(content.Tags),
		content.ContentCategory,
		content.AccessLevel,
		content.UploadCorrelationID,
		content.ProcessingAttempts,
		content.CreatedOn,
		content.CreatedBy,
		content.IsDeleted,
	)

	return err
}

func (r *PostgreSQLContentRepository) GetByID(ctx context.Context, contentID string) (*Content, error) {
	query := `
		SELECT 
			content_id, original_filename, file_size, mime_type, content_hash,
			storage_path, upload_status, alt_text, description, tags,
			content_category, access_level, upload_correlation_id, 
			processing_attempts, last_processed_at, created_on, created_by,
			modified_on, modified_by, is_deleted, deleted_on, deleted_by
		FROM content 
		WHERE content_id = $1 AND is_deleted = false`

	return r.scanContent(ctx, query, contentID)
}


func (r *PostgreSQLContentRepository) Update(ctx context.Context, content *Content) error {
	now := time.Now()
	content.ModifiedOn = &now

	query := `
		UPDATE content SET
			alt_text = $2,
			description = $3,
			tags = $4,
			content_category = $5,
			access_level = $6,
			upload_status = $7,
			processing_attempts = $8,
			last_processed_at = $9,
			modified_on = $10,
			modified_by = $11
		WHERE content_id = $1 AND is_deleted = false`

	_, err := r.db.ExecContext(
		ctx, query,
		content.ContentID,
		content.AltText,
		content.Description,
		pq.Array(content.Tags),
		content.ContentCategory,
		content.AccessLevel,
		content.UploadStatus,
		content.ProcessingAttempts,
		content.LastProcessedAt,
		content.ModifiedOn,
		content.ModifiedBy,
	)

	return err
}

func (r *PostgreSQLContentRepository) Delete(ctx context.Context, contentID, userID string) error {
	now := time.Now()
	query := `
		UPDATE content SET
			is_deleted = true,
			deleted_on = $2,
			deleted_by = $3,
			modified_on = $4,
			modified_by = $5
		WHERE content_id = $1 AND is_deleted = false`

	_, err := r.db.ExecContext(ctx, query, contentID, now, userID, now, userID)
	return err
}

func (r *PostgreSQLContentRepository) List(ctx context.Context, offset, limit int) ([]*Content, error) {
	query := `
		SELECT 
			content_id, original_filename, file_size, mime_type, content_hash,
			storage_path, upload_status, alt_text, description, tags,
			content_category, access_level, upload_correlation_id, 
			processing_attempts, last_processed_at, created_on, created_by,
			modified_on, modified_by, is_deleted, deleted_on, deleted_by
		FROM content 
		WHERE is_deleted = false
		ORDER BY created_on DESC
		LIMIT $1 OFFSET $2`

	return r.scanContents(ctx, query, limit, offset)
}

func (r *PostgreSQLContentRepository) ListByCategory(ctx context.Context, contentCategory ContentCategory, offset, limit int) ([]*Content, error) {
	query := `
		SELECT 
			content_id, original_filename, file_size, mime_type, content_hash,
			storage_path, upload_status, alt_text, description, tags,
			content_category, access_level, upload_correlation_id, 
			processing_attempts, last_processed_at, created_on, created_by,
			modified_on, modified_by, is_deleted, deleted_on, deleted_by
		FROM content 
		WHERE content_category = $1 AND is_deleted = false
		ORDER BY created_on DESC
		LIMIT $2 OFFSET $3`

	return r.scanContents(ctx, query, contentCategory, limit, offset)
}

func (r *PostgreSQLContentRepository) ListByTags(ctx context.Context, tags []string, offset, limit int) ([]*Content, error) {
	query := `
		SELECT 
			content_id, original_filename, file_size, mime_type, content_hash,
			storage_path, upload_status, alt_text, description, tags,
			content_category, access_level, upload_correlation_id, 
			processing_attempts, last_processed_at, created_on, created_by,
			modified_on, modified_by, is_deleted, deleted_on, deleted_by
		FROM content 
		WHERE tags && $1 AND is_deleted = false
		ORDER BY created_on DESC
		LIMIT $2 OFFSET $3`

	return r.scanContents(ctx, query, pq.Array(tags), limit, offset)
}

func (r *PostgreSQLContentRepository) ListPublished(ctx context.Context, offset, limit int) ([]*Content, error) {
	query := `
		SELECT 
			content_id, original_filename, file_size, mime_type, content_hash,
			storage_path, upload_status, alt_text, description, tags,
			content_category, access_level, upload_correlation_id, 
			processing_attempts, last_processed_at, created_on, created_by,
			modified_on, modified_by, is_deleted, deleted_on, deleted_by
		FROM content 
		WHERE upload_status = $1 AND is_deleted = false
		ORDER BY created_on DESC
		LIMIT $2 OFFSET $3`

	return r.scanContents(ctx, query, UploadStatusAvailable, limit, offset)
}

func (r *PostgreSQLContentRepository) ListByType(ctx context.Context, contentCategory ContentCategory, offset, limit int) ([]*Content, error) {
	query := `
		SELECT 
			content_id, original_filename, file_size, mime_type, content_hash,
			storage_path, upload_status, alt_text, description, tags,
			content_category, access_level, upload_correlation_id, 
			processing_attempts, last_processed_at, created_on, created_by,
			modified_on, modified_by, is_deleted, deleted_on, deleted_by
		FROM content 
		WHERE content_category = $1 AND is_deleted = false
		ORDER BY created_on DESC
		LIMIT $2 OFFSET $3`

	return r.scanContents(ctx, query, contentCategory, limit, offset)
}

func (r *PostgreSQLContentRepository) scanContent(ctx context.Context, query string, args ...interface{}) (*Content, error) {
	row := r.db.QueryRowContext(ctx, query, args...)

	var content Content
	var altText sql.NullString
	var description sql.NullString
	var createdBy sql.NullString
	var modifiedOn sql.NullTime
	var modifiedBy sql.NullString
	var deletedOn sql.NullTime
	var deletedBy sql.NullString
	var lastProcessedAt sql.NullTime

	err := row.Scan(
		&content.ContentID,
		&content.OriginalFilename,
		&content.FileSize,
		&content.MimeType,
		&content.ContentHash,
		&content.StoragePath,
		&content.UploadStatus,
		&altText,
		&description,
		pq.Array(&content.Tags),
		&content.ContentCategory,
		&content.AccessLevel,
		&content.UploadCorrelationID,
		&content.ProcessingAttempts,
		&lastProcessedAt,
		&content.CreatedOn,
		&createdBy,
		&modifiedOn,
		&modifiedBy,
		&content.IsDeleted,
		&deletedOn,
		&deletedBy,
	)
	if err != nil {
		return nil, err
	}

	if altText.Valid {
		content.AltText = altText.String
	}
	if description.Valid {
		content.Description = description.String
	}
	if createdBy.Valid {
		content.CreatedBy = createdBy.String
	}
	if modifiedOn.Valid {
		content.ModifiedOn = &modifiedOn.Time
	}
	if modifiedBy.Valid {
		content.ModifiedBy = modifiedBy.String
	}
	if deletedOn.Valid {
		content.DeletedOn = &deletedOn.Time
	}
	if deletedBy.Valid {
		content.DeletedBy = deletedBy.String
	}
	if lastProcessedAt.Valid {
		content.LastProcessedAt = &lastProcessedAt.Time
	}

	return &content, nil
}

func (r *PostgreSQLContentRepository) scanContents(ctx context.Context, query string, args ...interface{}) ([]*Content, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contents []*Content
	for rows.Next() {
		var content Content
		var altText sql.NullString
		var description sql.NullString
		var createdBy sql.NullString
		var modifiedOn sql.NullTime
		var modifiedBy sql.NullString
		var deletedOn sql.NullTime
		var deletedBy sql.NullString
		var lastProcessedAt sql.NullTime

		err := rows.Scan(
			&content.ContentID,
			&content.OriginalFilename,
			&content.FileSize,
			&content.MimeType,
			&content.ContentHash,
			&content.StoragePath,
			&content.UploadStatus,
			&altText,
			&description,
			pq.Array(&content.Tags),
			&content.ContentCategory,
			&content.AccessLevel,
			&content.UploadCorrelationID,
			&content.ProcessingAttempts,
			&lastProcessedAt,
			&content.CreatedOn,
			&createdBy,
			&modifiedOn,
			&modifiedBy,
			&content.IsDeleted,
			&deletedOn,
			&deletedBy,
		)
		if err != nil {
			return nil, err
		}

		if altText.Valid {
			content.AltText = altText.String
		}
		if description.Valid {
			content.Description = description.String
		}
		if createdBy.Valid {
			content.CreatedBy = createdBy.String
		}
		if modifiedOn.Valid {
			content.ModifiedOn = &modifiedOn.Time
		}
		if modifiedBy.Valid {
			content.ModifiedBy = modifiedBy.String
		}
		if deletedOn.Valid {
			content.DeletedOn = &deletedOn.Time
		}
		if deletedBy.Valid {
			content.DeletedBy = deletedBy.String
		}
		if lastProcessedAt.Valid {
			content.LastProcessedAt = &lastProcessedAt.Time
		}

		contents = append(contents, &content)
	}

	return contents, rows.Err()
}