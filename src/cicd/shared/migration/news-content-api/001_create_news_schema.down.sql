-- News Domain Database Schema Rollback
-- Drops all news domain tables, indexes, functions, and triggers

-- Drop triggers first
DROP TRIGGER IF EXISTS featured_news_management_trigger ON featured_news;
DROP TRIGGER IF EXISTS news_publication_status_trigger ON news;
DROP TRIGGER IF EXISTS featured_news_audit_trigger ON featured_news;
DROP TRIGGER IF EXISTS news_categories_audit_trigger ON news_categories;
DROP TRIGGER IF EXISTS news_audit_trigger ON news;

-- Drop functions
DROP FUNCTION IF EXISTS handle_featured_news_changes();
DROP FUNCTION IF EXISTS handle_news_publication_status_change();
DROP FUNCTION IF EXISTS publish_news_audit_event_to_grafana_loki();

-- Drop indexes (PostgreSQL automatically drops indexes when tables are dropped, but being explicit)

-- Featured news indexes
DROP INDEX IF EXISTS idx_featured_news_created_on;
DROP INDEX IF EXISTS idx_featured_news_news_id;

-- News table indexes
DROP INDEX IF EXISTS idx_news_content_fulltext;
DROP INDEX IF EXISTS idx_news_summary_fulltext;
DROP INDEX IF EXISTS idx_news_title_fulltext;
DROP INDEX IF EXISTS idx_news_tags;
DROP INDEX IF EXISTS idx_news_created_on;
DROP INDEX IF EXISTS idx_news_external_source;
DROP INDEX IF EXISTS idx_news_author_name;
DROP INDEX IF EXISTS idx_news_publication_timestamp;
DROP INDEX IF EXISTS idx_news_priority_level;
DROP INDEX IF EXISTS idx_news_news_type;
DROP INDEX IF EXISTS idx_news_publishing_status;
DROP INDEX IF EXISTS idx_news_category_id;
DROP INDEX IF EXISTS idx_news_slug;
DROP INDEX IF EXISTS idx_news_title;

-- News categories indexes
DROP INDEX IF EXISTS idx_news_categories_created_on;
DROP INDEX IF EXISTS idx_news_categories_default;
DROP INDEX IF EXISTS idx_news_categories_slug;
DROP INDEX IF EXISTS idx_news_categories_name;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS featured_news;
DROP TABLE IF EXISTS news;
DROP TABLE IF EXISTS news_categories;