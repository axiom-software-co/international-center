-- Create performance indexes for news domain matching TABLES-NEWS.md specification

-- News table indexes
CREATE INDEX idx_news_category_id ON news(category_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_publishing_status ON news(publishing_status) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_slug ON news(slug) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_news_type ON news(news_type) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_priority_level ON news(priority_level) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_publication_timestamp ON news(publication_timestamp) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_author_name ON news(author_name) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_external_source ON news(external_source) WHERE is_deleted = FALSE;

-- News categories table indexes  
CREATE INDEX idx_news_categories_slug ON news_categories(slug) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_categories_default ON news_categories(is_default_unassigned) WHERE is_deleted = FALSE;

-- Featured news table indexes
CREATE INDEX idx_featured_news_news_id ON featured_news(news_id);