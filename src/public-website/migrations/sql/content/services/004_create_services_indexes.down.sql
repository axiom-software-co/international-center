-- Drop services domain indexes

-- Services table indexes
DROP INDEX IF EXISTS idx_services_category_id;
DROP INDEX IF EXISTS idx_services_publishing_status;
DROP INDEX IF EXISTS idx_services_slug;
DROP INDEX IF EXISTS idx_services_order_category;
DROP INDEX IF EXISTS idx_services_delivery_mode;

-- Service categories table indexes
DROP INDEX IF EXISTS idx_service_categories_slug;
DROP INDEX IF EXISTS idx_service_categories_order;
DROP INDEX IF EXISTS idx_service_categories_default;

-- Featured categories table indexes
DROP INDEX IF EXISTS idx_featured_categories_category_id;
DROP INDEX IF EXISTS idx_featured_categories_position;