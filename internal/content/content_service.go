package content

import (
	"context"
)

type ContentService struct {
	repository ContentRepository
}

func NewContentService(repository ContentRepository) *ContentService {
	return &ContentService{
		repository: repository,
	}
}

func (s *ContentService) CreateContent(ctx context.Context, title, body, slug, contentType string) (*Content, error) {
	content, err := NewContent(title, body, slug, contentType)
	if err != nil {
		return nil, err
	}
	
	err = s.repository.Create(ctx, content)
	if err != nil {
		return nil, err
	}
	
	return content, nil
}

func (s *ContentService) GetContent(ctx context.Context, contentID string) (*Content, error) {
	return s.repository.GetByID(ctx, contentID)
}

func (s *ContentService) GetContentBySlug(ctx context.Context, slug string) (*Content, error) {
	return s.repository.GetBySlug(ctx, slug)
}

func (s *ContentService) UpdateContent(ctx context.Context, contentID, title, body, slug, contentType, userID string) (*Content, error) {
	content, err := s.repository.GetByID(ctx, contentID)
	if err != nil {
		return nil, err
	}
	
	content.Title = title
	content.Body = body
	content.Slug = slug
	content.ContentType = ContentType(contentType)
	content.ModifiedBy = userID
	
	if !isValidContentType(content.ContentType) {
		return nil, err
	}
	
	err = s.repository.Update(ctx, content)
	if err != nil {
		return nil, err
	}
	
	return content, nil
}

func (s *ContentService) PublishContent(ctx context.Context, contentID, userID string) error {
	content, err := s.repository.GetByID(ctx, contentID)
	if err != nil {
		return err
	}
	
	err = content.Publish(userID)
	if err != nil {
		return err
	}
	
	return s.repository.Update(ctx, content)
}

func (s *ContentService) ArchiveContent(ctx context.Context, contentID, userID string) error {
	content, err := s.repository.GetByID(ctx, contentID)
	if err != nil {
		return err
	}
	
	err = content.Archive(userID)
	if err != nil {
		return err
	}
	
	return s.repository.Update(ctx, content)
}

func (s *ContentService) AssignContentCategory(ctx context.Context, contentID, categoryID, userID string) error {
	content, err := s.repository.GetByID(ctx, contentID)
	if err != nil {
		return err
	}
	
	err = content.AssignCategory(categoryID, userID)
	if err != nil {
		return err
	}
	
	return s.repository.Update(ctx, content)
}

func (s *ContentService) AssignContentTags(ctx context.Context, contentID string, tags []string, userID string) error {
	content, err := s.repository.GetByID(ctx, contentID)
	if err != nil {
		return err
	}
	
	err = content.AssignTags(tags, userID)
	if err != nil {
		return err
	}
	
	return s.repository.Update(ctx, content)
}

func (s *ContentService) DeleteContent(ctx context.Context, contentID, userID string) error {
	return s.repository.Delete(ctx, contentID, userID)
}

func (s *ContentService) ListContent(ctx context.Context, offset, limit int) ([]*Content, error) {
	return s.repository.List(ctx, offset, limit)
}

func (s *ContentService) ListContentByCategory(ctx context.Context, categoryID string, offset, limit int) ([]*Content, error) {
	return s.repository.ListByCategory(ctx, categoryID, offset, limit)
}

func (s *ContentService) ListContentByTags(ctx context.Context, tags []string, offset, limit int) ([]*Content, error) {
	return s.repository.ListByTags(ctx, tags, offset, limit)
}

func (s *ContentService) ListPublishedContent(ctx context.Context, offset, limit int) ([]*Content, error) {
	return s.repository.ListPublished(ctx, offset, limit)
}

func (s *ContentService) ListContentByType(ctx context.Context, contentType ContentType, offset, limit int) ([]*Content, error) {
	return s.repository.ListByType(ctx, contentType, offset, limit)
}