-- Create performance indexes for research domain matching TABLES-RESEARCH.md specification

-- Research table indexes
CREATE INDEX idx_research_category_id ON research(category_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_research_publishing_status ON research(publishing_status) WHERE is_deleted = FALSE;
CREATE INDEX idx_research_slug ON research(slug) WHERE is_deleted = FALSE;
CREATE INDEX idx_research_research_type ON research(research_type) WHERE is_deleted = FALSE;
CREATE INDEX idx_research_publication_date ON research(publication_date) WHERE is_deleted = FALSE;
CREATE INDEX idx_research_author_names ON research(author_names) WHERE is_deleted = FALSE;
CREATE INDEX idx_research_doi ON research(doi) WHERE is_deleted = FALSE;

-- Research categories table indexes  
CREATE INDEX idx_research_categories_slug ON research_categories(slug) WHERE is_deleted = FALSE;
CREATE INDEX idx_research_categories_default ON research_categories(is_default_unassigned) WHERE is_deleted = FALSE;

-- Featured research table indexes
CREATE INDEX idx_featured_research_research_id ON featured_research(research_id);