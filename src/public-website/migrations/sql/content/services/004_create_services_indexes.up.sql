-- Create performance indexes for services domain matching TABLES-SERVICES.md specification

-- Services table indexes
CREATE INDEX idx_services_category_id ON services(category_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_services_publishing_status ON services(publishing_status) WHERE is_deleted = FALSE;
CREATE INDEX idx_services_slug ON services(slug) WHERE is_deleted = FALSE;
CREATE INDEX idx_services_order_category ON services(category_id, order_number) WHERE is_deleted = FALSE;
CREATE INDEX idx_services_delivery_mode ON services(delivery_mode) WHERE is_deleted = FALSE;

-- Service categories table indexes  
CREATE INDEX idx_service_categories_slug ON service_categories(slug) WHERE is_deleted = FALSE;
CREATE INDEX idx_service_categories_order ON service_categories(order_number) WHERE is_deleted = FALSE;
CREATE INDEX idx_service_categories_default ON service_categories(is_default_unassigned) WHERE is_deleted = FALSE;

-- Featured categories table indexes
CREATE INDEX idx_featured_categories_category_id ON featured_categories(category_id);
CREATE INDEX idx_featured_categories_position ON featured_categories(feature_position);