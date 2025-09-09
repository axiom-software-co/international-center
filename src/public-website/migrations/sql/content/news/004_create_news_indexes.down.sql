-- Drop news domain indexes

-- News table indexes
DROP INDEX IF EXISTS idx_news_category_id;
DROP INDEX IF EXISTS idx_news_publishing_status;
DROP INDEX IF EXISTS idx_news_slug;
DROP INDEX IF EXISTS idx_news_news_type;
DROP INDEX IF EXISTS idx_news_priority_level;
DROP INDEX IF EXISTS idx_news_publication_timestamp;
DROP INDEX IF EXISTS idx_news_author_name;
DROP INDEX IF EXISTS idx_news_external_source;

-- News categories table indexes
DROP INDEX IF EXISTS idx_news_categories_slug;
DROP INDEX IF EXISTS idx_news_categories_default;

-- Featured news table indexes
DROP INDEX IF EXISTS idx_featured_news_news_id;