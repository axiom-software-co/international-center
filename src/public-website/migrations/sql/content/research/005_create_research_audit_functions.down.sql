-- Drop audit functions and triggers for research domain

DROP TRIGGER IF EXISTS research_audit_trigger ON research;
DROP TRIGGER IF EXISTS research_categories_audit_trigger ON research_categories;
DROP TRIGGER IF EXISTS featured_research_audit_trigger ON featured_research;
DROP TRIGGER IF EXISTS research_categories_reassign_trigger ON research_categories;
DROP TRIGGER IF EXISTS featured_research_management_trigger ON featured_research;

DROP FUNCTION IF EXISTS publish_research_audit_event_to_grafana_loki();
DROP FUNCTION IF EXISTS reassign_research_to_default_category();
DROP FUNCTION IF EXISTS manage_featured_research();