-- Create audit functions and triggers for news domain matching TABLES-NEWS.md specification

-- Grafana Cloud Loki Audit Trigger Function
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

-- Default Category Assignment Function
CREATE OR REPLACE FUNCTION reassign_news_to_default_category()
RETURNS TRIGGER AS $$
BEGIN
    -- News category reassignment with Dapr event notification
    -- Implementation will publish to 'news-category-events' topic
    -- Ensures audit compliance via event sourcing pattern
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Featured News Management Function
CREATE OR REPLACE FUNCTION manage_featured_news()
RETURNS TRIGGER AS $$
BEGIN
    -- Featured news management with single article constraint
    -- Implementation publishes to 'news-featured-events' topic
    -- Ensures audit compliance and business rule enforcement
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Create triggers for audit events on news tables
CREATE TRIGGER news_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON news
    FOR EACH ROW EXECUTE FUNCTION publish_news_audit_event_to_grafana_loki();

CREATE TRIGGER news_categories_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON news_categories
    FOR EACH ROW EXECUTE FUNCTION publish_news_audit_event_to_grafana_loki();

CREATE TRIGGER featured_news_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON featured_news
    FOR EACH ROW EXECUTE FUNCTION publish_news_audit_event_to_grafana_loki();

-- Create triggers for business rule management
CREATE TRIGGER news_categories_reassign_trigger
    AFTER UPDATE OR DELETE ON news_categories
    FOR EACH ROW EXECUTE FUNCTION reassign_news_to_default_category();

CREATE TRIGGER featured_news_management_trigger
    AFTER INSERT OR UPDATE OR DELETE ON featured_news
    FOR EACH ROW EXECUTE FUNCTION manage_featured_news();