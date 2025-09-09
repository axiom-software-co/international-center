-- Drop research domain indexes

-- Research table indexes
DROP INDEX IF EXISTS idx_research_category_id;
DROP INDEX IF EXISTS idx_research_publishing_status;
DROP INDEX IF EXISTS idx_research_slug;
DROP INDEX IF EXISTS idx_research_research_type;
DROP INDEX IF EXISTS idx_research_publication_date;
DROP INDEX IF EXISTS idx_research_author_names;
DROP INDEX IF EXISTS idx_research_doi;

-- Research categories table indexes
DROP INDEX IF EXISTS idx_research_categories_slug;
DROP INDEX IF EXISTS idx_research_categories_default;

-- Featured research table indexes
DROP INDEX IF EXISTS idx_featured_research_research_id;