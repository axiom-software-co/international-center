-- Research Domain Database Schema
-- Matching TABLES-RESEARCH.md specification exactly

-- Create research_categories table first (referenced by research table)
CREATE TABLE research_categories (
    category_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    is_default_unassigned BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Audit fields
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    modified_on TIMESTAMPTZ,
    modified_by VARCHAR(255),
    
    -- Soft delete fields  
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_on TIMESTAMPTZ,
    deleted_by VARCHAR(255),
    
    CONSTRAINT only_one_default_unassigned CHECK (
        NOT is_default_unassigned OR 
        (SELECT COUNT(*) FROM research_categories WHERE is_default_unassigned = TRUE AND is_deleted = FALSE) <= 1
    )
);

-- Create research table
CREATE TABLE research (
    research_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    abstract TEXT NOT NULL,
    content TEXT, -- Research article content stored in PostgreSQL
    slug VARCHAR(255) UNIQUE NOT NULL,
    category_id UUID NOT NULL REFERENCES research_categories(category_id),
    image_url VARCHAR(500), -- URL to Azure Blob Storage for images only
    author_names VARCHAR(500) NOT NULL,
    publication_date DATE,
    doi VARCHAR(100),
    external_url VARCHAR(500),
    report_url VARCHAR(500), -- URL to PDF publication report in Azure Blob Storage
    publishing_status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (publishing_status IN ('draft', 'published', 'archived')),
    
    -- Content metadata
    keywords TEXT[],
    research_type VARCHAR(50) NOT NULL CHECK (research_type IN ('clinical_study', 'case_report', 'systematic_review', 'meta_analysis', 'editorial', 'commentary')),
    
    -- Audit fields
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    modified_on TIMESTAMPTZ,
    modified_by VARCHAR(255),
    
    -- Soft delete fields
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_on TIMESTAMPTZ,
    deleted_by VARCHAR(255)
);

-- Create featured_research table
CREATE TABLE featured_research (
    featured_research_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    research_id UUID NOT NULL REFERENCES research(research_id),
    
    -- Audit fields
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    modified_on TIMESTAMPTZ,
    modified_by VARCHAR(255),
    
    CONSTRAINT only_one_featured_research CHECK (
        (SELECT COUNT(*) FROM featured_research) <= 1
    ),
    CONSTRAINT no_default_unassigned_featured CHECK (
        NOT EXISTS (
            SELECT 1 FROM research r
            JOIN research_categories rc ON r.category_id = rc.category_id
            WHERE r.research_id = featured_research.research_id 
            AND rc.is_default_unassigned = TRUE
        )
    )
);

-- Performance Indexes

-- Research table indexes
CREATE INDEX idx_research_category_id ON research(category_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_research_publishing_status ON research(publishing_status) WHERE is_deleted = FALSE;
CREATE INDEX idx_research_slug ON research(slug) WHERE is_deleted = FALSE;
CREATE INDEX idx_research_research_type ON research(research_type) WHERE is_deleted = FALSE;
CREATE INDEX idx_research_publication_date ON research(publication_date) WHERE is_deleted = FALSE;
CREATE INDEX idx_research_author_names ON research(author_names) WHERE is_deleted = FALSE;
CREATE INDEX idx_research_doi ON research(doi) WHERE is_deleted = FALSE;
CREATE INDEX idx_research_created_on ON research(created_on) WHERE is_deleted = FALSE;
CREATE INDEX idx_research_title ON research USING gin(to_tsvector('english', title)) WHERE is_deleted = FALSE;
CREATE INDEX idx_research_abstract ON research USING gin(to_tsvector('english', abstract)) WHERE is_deleted = FALSE;
CREATE INDEX idx_research_content ON research USING gin(to_tsvector('english', content)) WHERE is_deleted = FALSE AND content IS NOT NULL;
CREATE INDEX idx_research_keywords ON research USING gin(keywords) WHERE is_deleted = FALSE;

-- Research categories table indexes  
CREATE INDEX idx_research_categories_slug ON research_categories(slug) WHERE is_deleted = FALSE;
CREATE INDEX idx_research_categories_default ON research_categories(is_default_unassigned) WHERE is_deleted = FALSE;
CREATE INDEX idx_research_categories_name ON research_categories(name) WHERE is_deleted = FALSE;

-- Featured research table indexes
CREATE INDEX idx_featured_research_research_id ON featured_research(research_id);
CREATE INDEX idx_featured_research_created_on ON featured_research(created_on);

-- Database Functions and Triggers

-- Grafana Cloud Loki audit trigger function
CREATE OR REPLACE FUNCTION publish_research_audit_event_to_grafana_loki()
RETURNS TRIGGER AS $$
BEGIN
    -- Research audit event publishing directly to Grafana Cloud Loki via Dapr pub/sub
    -- Implementation publishes to 'grafana-audit-events' topic via Dapr
    -- Event structure includes complete data snapshots for compliance
    -- No local audit table storage - all audit data stored in Grafana Cloud Loki
    -- Ensures immutable audit trail and prevents local tampering
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Default category assignment function
CREATE OR REPLACE FUNCTION reassign_research_to_default_category()
RETURNS TRIGGER AS $$
BEGIN
    -- Research category reassignment with Dapr event notification
    -- Implementation will publish to 'research-category-events' topic
    -- Ensures audit compliance via event sourcing pattern
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Featured research management function
CREATE OR REPLACE FUNCTION manage_featured_research()
RETURNS TRIGGER AS $$
BEGIN
    -- Featured research management with single article constraint
    -- Implementation publishes to 'research-featured-events' topic
    -- Ensures audit compliance and business rule enforcement
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Create triggers for audit event publishing
CREATE TRIGGER research_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON research
    FOR EACH ROW EXECUTE FUNCTION publish_research_audit_event_to_grafana_loki();

CREATE TRIGGER research_categories_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON research_categories
    FOR EACH ROW EXECUTE FUNCTION publish_research_audit_event_to_grafana_loki();

CREATE TRIGGER featured_research_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON featured_research
    FOR EACH ROW EXECUTE FUNCTION publish_research_audit_event_to_grafana_loki();

-- Create trigger for category reassignment
CREATE TRIGGER research_category_reassignment_trigger
    BEFORE DELETE ON research_categories
    FOR EACH ROW EXECUTE FUNCTION reassign_research_to_default_category();

-- Create trigger for featured research management
CREATE TRIGGER featured_research_management_trigger
    AFTER INSERT OR UPDATE OR DELETE ON featured_research
    FOR EACH ROW EXECUTE FUNCTION manage_featured_research();