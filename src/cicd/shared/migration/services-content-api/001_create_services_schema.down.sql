-- Rollback Services Domain Database Schema
-- Drop in reverse order of creation to handle dependencies

-- Drop triggers first
DROP TRIGGER IF EXISTS service_category_deletion_trigger ON service_categories;
DROP TRIGGER IF EXISTS featured_categories_audit_trigger ON featured_categories;
DROP TRIGGER IF EXISTS service_categories_audit_trigger ON service_categories;
DROP TRIGGER IF EXISTS services_audit_trigger ON services;

-- Drop functions
DROP FUNCTION IF EXISTS reassign_services_to_default_category();
DROP FUNCTION IF EXISTS publish_audit_event_to_grafana_loki();

-- Drop indexes
DROP INDEX IF EXISTS idx_featured_categories_position;
DROP INDEX IF EXISTS idx_featured_categories_category_id;
DROP INDEX IF EXISTS idx_service_categories_default;
DROP INDEX IF EXISTS idx_service_categories_order;
DROP INDEX IF EXISTS idx_service_categories_slug;
DROP INDEX IF EXISTS idx_services_delivery_mode;
DROP INDEX IF EXISTS idx_services_order_category;
DROP INDEX IF EXISTS idx_services_slug;
DROP INDEX IF EXISTS idx_services_publishing_status;
DROP INDEX IF EXISTS idx_services_category_id;

-- Drop tables in dependency order
DROP TABLE IF EXISTS featured_categories;
DROP TABLE IF EXISTS services;
DROP TABLE IF EXISTS service_categories;