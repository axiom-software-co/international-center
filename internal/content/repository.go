package content

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ContentRepository interface {
	Create(ctx context.Context, content *Content) error
	GetByID(ctx context.Context, contentID string) (*Content, error)
	GetBySlug(ctx context.Context, slug string) (*Content, error)
	Update(ctx context.Context, content *Content) error
	Delete(ctx context.Context, contentID, userID string) error
	List(ctx context.Context, offset, limit int) ([]*Content, error)
	ListByCategory(ctx context.Context, categoryID string, offset, limit int) ([]*Content, error)
	ListByTags(ctx context.Context, tags []string, offset, limit int) ([]*Content, error)
	ListPublished(ctx context.Context, offset, limit int) ([]*Content, error)
	ListByType(ctx context.Context, contentType ContentType, offset, limit int) ([]*Content, error)
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

func (r *MongoContentRepository) GetBySlug(ctx context.Context, slug string) (*Content, error) {
	filter := bson.M{
		"slug":       slug,
		"is_deleted": false,
	}
	
	var content Content
	err := r.collection.FindOne(ctx, filter).Decode(&content)
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
	
	content.ModifiedOn = time.Now()
	
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

func (r *MongoContentRepository) ListByCategory(ctx context.Context, categoryID string, offset, limit int) ([]*Content, error) {
	filter := bson.M{
		"category_id": categoryID,
		"is_deleted":  false,
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
		"publishing_status": PublishingStatusPublished,
		"is_deleted":        false,
	}
	
	findOptions := options.Find()
	findOptions.SetSkip(int64(offset))
	findOptions.SetLimit(int64(limit))
	findOptions.SetSort(bson.M{"created_on": -1})
	
	return r.findContents(ctx, filter, findOptions)
}

func (r *MongoContentRepository) ListByType(ctx context.Context, contentType ContentType, offset, limit int) ([]*Content, error) {
	filter := bson.M{
		"content_type": contentType,
		"is_deleted":   false,
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