-- Research Domain Database Schema Rollback
-- Removes all research domain tables and related database objects

-- Drop triggers first (to avoid dependency issues)
DROP TRIGGER IF EXISTS featured_research_management_trigger ON featured_research;
DROP TRIGGER IF EXISTS research_category_reassignment_trigger ON research_categories;
DROP TRIGGER IF EXISTS featured_research_audit_trigger ON featured_research;
DROP TRIGGER IF EXISTS research_categories_audit_trigger ON research_categories;
DROP TRIGGER IF EXISTS research_audit_trigger ON research;

-- Drop functions
DROP FUNCTION IF EXISTS manage_featured_research();
DROP FUNCTION IF EXISTS reassign_research_to_default_category();
DROP FUNCTION IF EXISTS publish_research_audit_event_to_grafana_loki();

-- Drop indexes (PostgreSQL will auto-drop with tables, but explicit for clarity)
DROP INDEX IF EXISTS idx_featured_research_created_on;
DROP INDEX IF EXISTS idx_featured_research_research_id;
DROP INDEX IF EXISTS idx_research_categories_name;
DROP INDEX IF EXISTS idx_research_categories_default;
DROP INDEX IF EXISTS idx_research_categories_slug;
DROP INDEX IF EXISTS idx_research_keywords;
DROP INDEX IF EXISTS idx_research_content;
DROP INDEX IF EXISTS idx_research_abstract;
DROP INDEX IF EXISTS idx_research_title;
DROP INDEX IF EXISTS idx_research_created_on;
DROP INDEX IF EXISTS idx_research_doi;
DROP INDEX IF EXISTS idx_research_author_names;
DROP INDEX IF EXISTS idx_research_publication_date;
DROP INDEX IF EXISTS idx_research_research_type;
DROP INDEX IF EXISTS idx_research_slug;
DROP INDEX IF EXISTS idx_research_publishing_status;
DROP INDEX IF EXISTS idx_research_category_id;

-- Drop tables (in reverse dependency order)
DROP TABLE IF EXISTS featured_research;
DROP TABLE IF EXISTS research;
DROP TABLE IF EXISTS research_categories;