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

func (s *ContentService) CreateContent(ctx context.Context, originalFilename string, fileSize int64, mimeType string, contentHash string, contentCategory ContentCategory) (*Content, error) {
	content, err := NewContent(originalFilename, fileSize, mimeType, contentHash, contentCategory)
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

func (s *ContentService) UpdateContentMetadata(ctx context.Context, contentID, altText, description string, tags []string, userID string) (*Content, error) {
	content, err := s.repository.GetByID(ctx, contentID)
	if err != nil {
		return nil, err
	}
	
	if altText != "" {
		content.AltText = altText
	}
	
	if description != "" {
		err = content.SetDescription(description, userID)
		if err != nil {
			return nil, err
		}
	}
	
	if tags != nil {
		err = content.AssignTags(tags, userID)
		if err != nil {
			return nil, err
		}
	}
	
	err = s.repository.Update(ctx, content)
	if err != nil {
		return nil, err
	}
	
	return content, nil
}

func (s *ContentService) MarkContentAsAvailable(ctx context.Context, contentID, userID string) error {
	content, err := s.repository.GetByID(ctx, contentID)
	if err != nil {
		return err
	}
	
	err = content.MarkAsAvailable(userID)
	if err != nil {
		return err
	}
	
	return s.repository.Update(ctx, content)
}

func (s *ContentService) MarkContentAsFailed(ctx context.Context, contentID, userID string) error {
	content, err := s.repository.GetByID(ctx, contentID)
	if err != nil {
		return err
	}
	
	err = content.MarkAsFailed(userID)
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

func (s *ContentService) SetContentAccessLevel(ctx context.Context, contentID string, accessLevel AccessLevel, userID string) error {
	content, err := s.repository.GetByID(ctx, contentID)
	if err != nil {
		return err
	}
	
	err = content.SetAccessLevel(accessLevel, userID)
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

func (s *ContentService) ListContentByCategory(ctx context.Context, contentCategory ContentCategory, offset, limit int) ([]*Content, error) {
	return s.repository.ListByCategory(ctx, contentCategory, offset, limit)
}

func (s *ContentService) ListContentByTags(ctx context.Context, tags []string, offset, limit int) ([]*Content, error) {
	return s.repository.ListByTags(ctx, tags, offset, limit)
}

func (s *ContentService) ListAvailableContent(ctx context.Context, offset, limit int) ([]*Content, error) {
	return s.repository.ListPublished(ctx, offset, limit)
}

func (s *ContentService) ListContentByType(ctx context.Context, contentCategory ContentCategory, offset, limit int) ([]*Content, error) {
	return s.repository.ListByType(ctx, contentCategory, offset, limit)
}