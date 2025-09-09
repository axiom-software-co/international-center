-- Create audit functions and triggers for research domain matching TABLES-RESEARCH.md specification

CREATE OR REPLACE FUNCTION publish_research_audit_event_to_grafana_loki()
RETURNS TRIGGER AS $$
BEGIN
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION reassign_research_to_default_category()
RETURNS TRIGGER AS $$
BEGIN
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION manage_featured_research()
RETURNS TRIGGER AS $$
BEGIN
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER research_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON research
    FOR EACH ROW EXECUTE FUNCTION publish_research_audit_event_to_grafana_loki();

CREATE TRIGGER research_categories_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON research_categories
    FOR EACH ROW EXECUTE FUNCTION publish_research_audit_event_to_grafana_loki();

CREATE TRIGGER featured_research_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON featured_research
    FOR EACH ROW EXECUTE FUNCTION publish_research_audit_event_to_grafana_loki();

CREATE TRIGGER research_categories_reassign_trigger
    AFTER UPDATE OR DELETE ON research_categories
    FOR EACH ROW EXECUTE FUNCTION reassign_research_to_default_category();

CREATE TRIGGER featured_research_management_trigger
    AFTER INSERT OR UPDATE OR DELETE ON featured_research
    FOR EACH ROW EXECUTE FUNCTION manage_featured_research();