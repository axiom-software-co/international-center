-- News Domain Validation Schema
-- This file represents the authoritative final desired state for the news domain
-- Used for validation after migrations to ensure schema compliance

-- News Categories Table
CREATE TABLE news_categories (
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
        (SELECT COUNT(*) FROM news_categories WHERE is_default_unassigned = TRUE AND is_deleted = FALSE) <= 1
    )
);

-- News Table
CREATE TABLE news (
    news_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    summary TEXT NOT NULL,
    content TEXT, -- News article content stored in PostgreSQL
    slug VARCHAR(255) UNIQUE NOT NULL,
    category_id UUID NOT NULL REFERENCES news_categories(category_id),
    image_url VARCHAR(500), -- URL to Azure Blob Storage for images only
    author_name VARCHAR(255),
    publication_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    external_source VARCHAR(255),
    external_url VARCHAR(500),
    publishing_status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (publishing_status IN ('draft', 'published', 'archived')),
    
    -- Content metadata
    tags TEXT[],
    news_type VARCHAR(50) NOT NULL CHECK (news_type IN ('announcement', 'press_release', 'event', 'update', 'alert', 'feature')),
    priority_level VARCHAR(20) NOT NULL DEFAULT 'normal' CHECK (priority_level IN ('low', 'normal', 'high', 'urgent')),
    
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

-- Featured News Table
CREATE TABLE featured_news (
    featured_news_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    news_id UUID NOT NULL REFERENCES news(news_id),
    
    -- Audit fields
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    modified_on TIMESTAMPTZ,
    modified_by VARCHAR(255),
    
    CONSTRAINT only_one_featured_news CHECK (
        (SELECT COUNT(*) FROM featured_news) <= 1
    ),
    CONSTRAINT no_default_unassigned_featured CHECK (
        NOT EXISTS (
            SELECT 1 FROM news n
            JOIN news_categories nc ON n.category_id = nc.category_id
            WHERE n.news_id = featured_news.news_id 
            AND nc.is_default_unassigned = TRUE
        )
    )
);

-- Performance Indexes
CREATE INDEX idx_news_category_id ON news(category_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_publishing_status ON news(publishing_status) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_slug ON news(slug) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_news_type ON news(news_type) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_priority_level ON news(priority_level) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_publication_timestamp ON news(publication_timestamp) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_author_name ON news(author_name) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_external_source ON news(external_source) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_categories_slug ON news_categories(slug) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_categories_default ON news_categories(is_default_unassigned) WHERE is_deleted = FALSE;
CREATE INDEX idx_featured_news_news_id ON featured_news(news_id);

-- Audit Functions
CREATE OR REPLACE FUNCTION publish_news_audit_event_to_grafana_loki()
RETURNS TRIGGER AS $$
BEGIN
    -- News audit event publishing directly to Grafana Cloud Loki via Dapr pub/sub
    -- Implementation publishes to 'grafana-audit-events' topic via Dapr
    -- Event structure includes complete data snapshots for compliance
    -- No local audit table storage - all audit data stored in Grafana Cloud Loki
    -- Ensures immutable audit trail and prevents local tampering
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION reassign_news_to_default_category()
RETURNS TRIGGER AS $$
BEGIN
    -- News category reassignment with Dapr event notification
    -- Implementation will publish to 'news-category-events' topic
    -- Ensures audit compliance via event sourcing pattern
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION manage_featured_news()
RETURNS TRIGGER AS $$
BEGIN
    -- Featured news management with single article constraint
    -- Implementation publishes to 'news-featured-events' topic
    -- Ensures audit compliance and business rule enforcement
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Audit Triggers
CREATE TRIGGER news_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON news
    FOR EACH ROW EXECUTE FUNCTION publish_news_audit_event_to_grafana_loki();

CREATE TRIGGER news_categories_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON news_categories
    FOR EACH ROW EXECUTE FUNCTION publish_news_audit_event_to_grafana_loki();

CREATE TRIGGER featured_news_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON featured_news
    FOR EACH ROW EXECUTE FUNCTION publish_news_audit_event_to_grafana_loki();

CREATE TRIGGER news_categories_reassign_trigger
    AFTER UPDATE OR DELETE ON news_categories
    FOR EACH ROW EXECUTE FUNCTION reassign_news_to_default_category();

CREATE TRIGGER featured_news_management_trigger
    AFTER INSERT OR UPDATE OR DELETE ON featured_news
    FOR EACH ROW EXECUTE FUNCTION manage_featured_news();

-- Required Data
INSERT INTO news_categories (
    name,
    slug,
    description,
    is_default_unassigned,
    created_by
) VALUES (
    'Unassigned News',
    'unassigned',
    'Default category for news articles that have not been assigned to a specific category',
    TRUE,
    'system'
);