package content

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dapr/go-sdk/client"
	"github.com/google/uuid"
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




type contentRepository struct {
	client        client.Client
	storeName     string
	bindingName   string
}

func NewContentRepository(daprClient client.Client, storeName, bindingName string) ContentRepository {
	return &contentRepository{
		client:      daprClient,
		storeName:   storeName,
		bindingName: bindingName,
	}
}

func (r *contentRepository) Create(ctx context.Context, content *Content) error {
	content.ContentID = uuid.New().String()
	content.CreatedOn = time.Now()
	content.IsDeleted = false
	content.UploadStatus = UploadStatusProcessing
	content.ProcessingAttempts = 0

	contentData, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("failed to marshal content: %w", err)
	}

	err = r.client.SaveState(ctx, r.storeName, content.ContentID, contentData, nil)
	if err != nil {
		return fmt.Errorf("failed to save content: %w", err)
	}

	// Create hash-based key for deduplication
	hashKey := fmt.Sprintf("hash:%s", content.ContentHash)
	hashData, _ := json.Marshal(map[string]string{"content_id": content.ContentID})
	err = r.client.SaveState(ctx, r.storeName, hashKey, hashData, nil)
	if err != nil {
		return fmt.Errorf("failed to save content hash mapping: %w", err)
	}

	return nil
}

func (r *contentRepository) GetByID(ctx context.Context, contentID string) (*Content, error) {
	item, err := r.client.GetState(ctx, r.storeName, contentID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get content: %w", err)
	}

	if len(item.Value) == 0 {
		return nil, fmt.Errorf("content not found")
	}

	var content Content
	err = json.Unmarshal(item.Value, &content)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal content: %w", err)
	}

	if content.IsDeleted {
		return nil, fmt.Errorf("content not found")
	}

	return &content, nil
}

func (r *contentRepository) Update(ctx context.Context, content *Content) error {
	now := time.Now()
	content.ModifiedOn = &now

	contentData, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("failed to marshal content: %w", err)
	}

	err = r.client.SaveState(ctx, r.storeName, content.ContentID, contentData, nil)
	if err != nil {
		return fmt.Errorf("failed to update content: %w", err)
	}

	return nil
}

func (r *contentRepository) Delete(ctx context.Context, contentID, userID string) error {
	content, err := r.GetByID(ctx, contentID)
	if err != nil {
		return err
	}

	now := time.Now()
	content.IsDeleted = true
	content.DeletedOn = &now
	content.DeletedBy = userID
	content.ModifiedOn = &now
	content.ModifiedBy = userID

	return r.Update(ctx, content)
}

func (r *contentRepository) List(ctx context.Context, offset, limit int) ([]*Content, error) {
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
		return nil, fmt.Errorf("failed to query content: %w", err)
	}

	contents := make([]*Content, 0)
	for _, result := range results.Results {
		var content Content
		err = json.Unmarshal(result.Value, &content)
		if err != nil {
			continue
		}
		if !content.IsDeleted {
			contents = append(contents, &content)
		}
	}

	return contents, nil
}

func (r *contentRepository) ListByCategory(ctx context.Context, contentCategory ContentCategory, offset, limit int) ([]*Content, error) {
	query := fmt.Sprintf(`{
		"filter": {
			"AND": [
				{ "EQ": { "content_category": "%s" } },
				{ "EQ": { "is_deleted": false } }
			]
		},
		"sort": [
			{ "key": "created_on", "order": "DESC" }
		],
		"page": {
			"limit": %d
		}
	}`, contentCategory, limit)

	results, err := r.client.QueryStateAlpha1(ctx, r.storeName, query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query content by category: %w", err)
	}

	contents := make([]*Content, 0)
	for _, result := range results.Results {
		var content Content
		err = json.Unmarshal(result.Value, &content)
		if err != nil {
			continue
		}
		if !content.IsDeleted && content.ContentCategory == contentCategory {
			contents = append(contents, &content)
		}
	}

	return contents, nil
}

func (r *contentRepository) ListByTags(ctx context.Context, tags []string, offset, limit int) ([]*Content, error) {
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
		return nil, fmt.Errorf("failed to query content: %w", err)
	}

	contents := make([]*Content, 0)
	for _, result := range results.Results {
		var content Content
		err = json.Unmarshal(result.Value, &content)
		if err != nil {
			continue
		}
		if !content.IsDeleted && r.hasAnyTag(content.Tags, tags) {
			contents = append(contents, &content)
		}
	}

	return contents, nil
}

func (r *contentRepository) ListPublished(ctx context.Context, offset, limit int) ([]*Content, error) {
	query := fmt.Sprintf(`{
		"filter": {
			"AND": [
				{ "EQ": { "upload_status": "%s" } },
				{ "EQ": { "is_deleted": false } }
			]
		},
		"sort": [
			{ "key": "created_on", "order": "DESC" }
		],
		"page": {
			"limit": %d
		}
	}`, UploadStatusAvailable, limit)

	results, err := r.client.QueryStateAlpha1(ctx, r.storeName, query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query published content: %w", err)
	}

	contents := make([]*Content, 0)
	for _, result := range results.Results {
		var content Content
		err = json.Unmarshal(result.Value, &content)
		if err != nil {
			continue
		}
		if !content.IsDeleted && content.UploadStatus == UploadStatusAvailable {
			contents = append(contents, &content)
		}
	}

	return contents, nil
}

func (r *contentRepository) ListByType(ctx context.Context, contentCategory ContentCategory, offset, limit int) ([]*Content, error) {
	return r.ListByCategory(ctx, contentCategory, offset, limit)
}

func (r *contentRepository) hasAnyTag(contentTags, searchTags []string) bool {
	for _, searchTag := range searchTags {
		for _, contentTag := range contentTags {
			if contentTag == searchTag {
				return true
			}
		}
	}
	return false
}