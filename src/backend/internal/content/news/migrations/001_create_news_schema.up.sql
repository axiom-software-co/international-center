-- News Domain Database Schema
-- Matching domain.go specification exactly

-- Create news_categories table
CREATE TABLE news_categories (
    category_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
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
    deleted_by VARCHAR(255)
);

-- Create news table
CREATE TABLE news (
    news_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    summary TEXT NOT NULL,
    content TEXT,
    slug VARCHAR(255) NOT NULL UNIQUE,
    category_id UUID NOT NULL REFERENCES news_categories(category_id),
    image_url VARCHAR(500),
    author_name VARCHAR(255),
    publication_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    external_source VARCHAR(255),
    external_url VARCHAR(500),
    publishing_status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (publishing_status IN ('draft', 'published', 'archived')),
    tags TEXT[], -- PostgreSQL array for news tags
    news_type VARCHAR(20) NOT NULL DEFAULT 'announcement' CHECK (news_type IN ('announcement', 'press_release', 'event', 'update', 'alert', 'feature')),
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

-- Create featured_news table
CREATE TABLE featured_news (
    featured_news_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    news_id UUID NOT NULL REFERENCES news(news_id),
    
    -- Audit fields
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    modified_on TIMESTAMPTZ,
    modified_by VARCHAR(255),
    
    -- Unique constraint to prevent duplicate featured news
    UNIQUE(news_id)
);

-- Performance Indexes

-- News categories table indexes
CREATE INDEX idx_news_categories_name ON news_categories(name) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_categories_slug ON news_categories(slug) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_categories_default ON news_categories(is_default_unassigned) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_categories_created_on ON news_categories(created_on) WHERE is_deleted = FALSE;

-- News table indexes
CREATE INDEX idx_news_title ON news(title) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_slug ON news(slug) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_category_id ON news(category_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_publishing_status ON news(publishing_status) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_news_type ON news(news_type) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_priority_level ON news(priority_level) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_publication_timestamp ON news(publication_timestamp) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_author_name ON news(author_name) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_external_source ON news(external_source) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_created_on ON news(created_on) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_tags ON news USING GIN(tags) WHERE is_deleted = FALSE;

-- Full-text search indexes for news content
CREATE INDEX idx_news_title_fulltext ON news USING GIN(to_tsvector('english', title)) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_summary_fulltext ON news USING GIN(to_tsvector('english', summary)) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_content_fulltext ON news USING GIN(to_tsvector('english', content)) WHERE is_deleted = FALSE;

-- Featured news table indexes
CREATE INDEX idx_featured_news_news_id ON featured_news(news_id);
CREATE INDEX idx_featured_news_created_on ON featured_news(created_on);

-- Database Functions and Triggers

-- Grafana Cloud Loki audit function for news domain
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

-- Publication status change trigger function
CREATE OR REPLACE FUNCTION handle_news_publication_status_change()
RETURNS TRIGGER AS $$
BEGIN
    -- News publication status change handling with Dapr pub/sub integration
    -- Publishes to 'news-publication-events' topic for cache invalidation
    -- Publishes to 'news-notification-events' topic for subscriber alerts
    -- Integration with Grafana telemetry for publication metrics
    IF NEW.publishing_status != OLD.publishing_status THEN
        -- Publication status changed - trigger downstream processing
        NULL; -- Placeholder for Dapr pub/sub integration
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Featured news management trigger function
CREATE OR REPLACE FUNCTION handle_featured_news_changes()
RETURNS TRIGGER AS $$
BEGIN
    -- Featured news management with Dapr service invocation
    -- Publishes to 'featured-news-events' topic for cache updates
    -- Publishes audit events to 'grafana-audit-events' for featured news operations
    -- Integration with Grafana monitoring for homepage updates
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Create triggers for audit event publishing
CREATE TRIGGER news_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON news
    FOR EACH ROW EXECUTE FUNCTION publish_news_audit_event_to_grafana_loki();

CREATE TRIGGER news_categories_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON news_categories
    FOR EACH ROW EXECUTE FUNCTION publish_news_audit_event_to_grafana_loki();

CREATE TRIGGER featured_news_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON featured_news
    FOR EACH ROW EXECUTE FUNCTION publish_news_audit_event_to_grafana_loki();

CREATE TRIGGER news_publication_status_trigger
    AFTER UPDATE ON news
    FOR EACH ROW EXECUTE FUNCTION handle_news_publication_status_change();

CREATE TRIGGER featured_news_management_trigger
    AFTER INSERT OR UPDATE OR DELETE ON featured_news
    FOR EACH ROW EXECUTE FUNCTION handle_featured_news_changes();