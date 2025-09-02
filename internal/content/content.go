package content

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

type ContentType string
type PublishingStatus string

const (
	ContentTypeArticle    ContentType = "article"
	ContentTypePage       ContentType = "page"
	ContentTypeNews       ContentType = "news"
	ContentTypeResource   ContentType = "resource"
)

const (
	PublishingStatusDraft     PublishingStatus = "draft"
	PublishingStatusPublished PublishingStatus = "published"
	PublishingStatusArchived  PublishingStatus = "archived"
)

type Content struct {
	ContentID        string           `bson:"_id,omitempty" json:"content_id"`
	Title            string           `bson:"title" json:"title"`
	Body             string           `bson:"body" json:"body"`
	Slug             string           `bson:"slug" json:"slug"`
	CategoryID       string           `bson:"category_id,omitempty" json:"category_id"`
	ContentType      ContentType      `bson:"content_type" json:"content_type"`
	PublishingStatus PublishingStatus `bson:"publishing_status" json:"publishing_status"`
	Tags             []string         `bson:"tags" json:"tags"`
	MetaDescription  string           `bson:"meta_description,omitempty" json:"meta_description"`
	FeaturedImageURL string           `bson:"featured_image_url,omitempty" json:"featured_image_url"`
	IsDeleted        bool             `bson:"is_deleted" json:"is_deleted"`
	CreatedOn        time.Time        `bson:"created_on" json:"created_on"`
	ModifiedOn       time.Time        `bson:"modified_on,omitempty" json:"modified_on"`
	ModifiedBy       string           `bson:"modified_by,omitempty" json:"modified_by"`
}

var slugRegex = regexp.MustCompile(`^[a-z0-9-]+$`)
var tagRegex = regexp.MustCompile(`^[a-z0-9-]+$`)

func NewContent(title, body, slug, contentType string) (*Content, error) {
	if strings.TrimSpace(title) == "" {
		return nil, errors.New("title cannot be empty")
	}
	
	if strings.TrimSpace(body) == "" {
		return nil, errors.New("body cannot be empty")
	}
	
	if strings.TrimSpace(slug) == "" {
		return nil, errors.New("slug cannot be empty")
	}
	
	if !isValidSlug(slug) {
		return nil, errors.New("slug must contain only lowercase letters, numbers, and hyphens")
	}
	
	contentTypeEnum := ContentType(contentType)
	if !isValidContentType(contentTypeEnum) {
		return nil, errors.New("invalid content type")
	}
	
	return &Content{
		Title:            title,
		Body:             body,
		Slug:             slug,
		ContentType:      contentTypeEnum,
		PublishingStatus: PublishingStatusDraft,
		Tags:             []string{},
		IsDeleted:        false,
		CreatedOn:        time.Now(),
	}, nil
}

func (c *Content) Publish(userID string) error {
	if c.PublishingStatus != PublishingStatusDraft {
		return errors.New("can only publish content with draft status")
	}
	
	c.PublishingStatus = PublishingStatusPublished
	c.ModifiedBy = userID
	c.ModifiedOn = time.Now()
	
	return nil
}

func (c *Content) Archive(userID string) error {
	if c.PublishingStatus != PublishingStatusPublished {
		return errors.New("can only archive content with published status")
	}
	
	c.PublishingStatus = PublishingStatusArchived
	c.ModifiedBy = userID
	c.ModifiedOn = time.Now()
	
	return nil
}

func (c *Content) AssignCategory(categoryID, userID string) error {
	if strings.TrimSpace(categoryID) == "" {
		return errors.New("category ID cannot be empty")
	}
	
	if !isValidUUID(categoryID) {
		return errors.New("category ID must be a valid UUID")
	}
	
	c.CategoryID = categoryID
	c.ModifiedBy = userID
	c.ModifiedOn = time.Now()
	
	return nil
}

func (c *Content) AssignTags(tags []string, userID string) error {
	for _, tag := range tags {
		if !isValidTag(tag) {
			return errors.New("tag must contain only lowercase letters, numbers, and hyphens")
		}
	}
	
	c.Tags = tags
	c.ModifiedBy = userID
	c.ModifiedOn = time.Now()
	
	return nil
}

func isValidSlug(slug string) bool {
	return slugRegex.MatchString(slug)
}

func isValidTag(tag string) bool {
	return tagRegex.MatchString(tag)
}

func isValidContentType(contentType ContentType) bool {
	switch contentType {
	case ContentTypeArticle, ContentTypePage, ContentTypeNews, ContentTypeResource:
		return true
	default:
		return false
	}
}

func isValidUUID(uuid string) bool {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return uuidRegex.MatchString(uuid)
}