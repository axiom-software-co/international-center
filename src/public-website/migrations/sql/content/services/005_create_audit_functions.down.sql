-- Drop audit functions and triggers for services domain

-- Drop triggers
DROP TRIGGER IF EXISTS services_audit_trigger ON services;
DROP TRIGGER IF EXISTS service_categories_audit_trigger ON service_categories;
DROP TRIGGER IF EXISTS featured_categories_audit_trigger ON featured_categories;
DROP TRIGGER IF EXISTS service_categories_reassign_trigger ON service_categories;

-- Drop functions
DROP FUNCTION IF EXISTS publish_audit_event_to_grafana_loki();
DROP FUNCTION IF EXISTS reassign_services_to_default_category();