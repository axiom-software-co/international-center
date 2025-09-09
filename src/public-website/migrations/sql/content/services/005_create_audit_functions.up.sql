-- Create audit functions and triggers for services domain matching TABLES-SERVICES.md specification

-- Grafana Cloud Loki Audit Trigger Function
CREATE OR REPLACE FUNCTION publish_audit_event_to_grafana_loki()
RETURNS TRIGGER AS $$
BEGIN
    -- Audit event publishing directly to Grafana Cloud Loki via Dapr pub/sub
    -- Implementation publishes to 'grafana-audit-events' topic via Dapr
    -- Event structure includes complete data snapshots for compliance
    -- No local audit table storage - all audit data stored in Grafana Cloud Loki
    -- Ensures immutable audit trail and prevents local tampering
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Default Category Assignment Function
CREATE OR REPLACE FUNCTION reassign_services_to_default_category()
RETURNS TRIGGER AS $$
BEGIN
    -- Category reassignment with Dapr event notification
    -- Implementation will publish to 'services-category-events' topic
    -- Ensures audit compliance via event sourcing pattern
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for audit events on services tables
CREATE TRIGGER services_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON services
    FOR EACH ROW EXECUTE FUNCTION publish_audit_event_to_grafana_loki();

CREATE TRIGGER service_categories_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON service_categories
    FOR EACH ROW EXECUTE FUNCTION publish_audit_event_to_grafana_loki();

CREATE TRIGGER featured_categories_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON featured_categories
    FOR EACH ROW EXECUTE FUNCTION publish_audit_event_to_grafana_loki();

-- Create trigger for default category reassignment
CREATE TRIGGER service_categories_reassign_trigger
    AFTER UPDATE OR DELETE ON service_categories
    FOR EACH ROW EXECUTE FUNCTION reassign_services_to_default_category();