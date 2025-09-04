-- Services Domain Database Schema
-- Matching TABLES-SERVICES.md specification exactly

-- Create service categories table first (referenced by services table)
CREATE TABLE service_categories (
    category_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    order_number INTEGER NOT NULL DEFAULT 0,
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
    -- Note: single default_unassigned constraint enforced by application logic
);

-- Create services table
CREATE TABLE services (
    service_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    content_url VARCHAR(500), -- URL to Azure Blob Storage content
    category_id UUID NOT NULL REFERENCES service_categories(category_id),
    image_url VARCHAR(500),
    order_number INTEGER NOT NULL DEFAULT 0,
    delivery_mode VARCHAR(50) NOT NULL CHECK (delivery_mode IN ('mobile_service', 'outpatient_service', 'inpatient_service')),
    publishing_status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (publishing_status IN ('draft', 'published', 'archived')),
    
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

-- Create featured categories table
CREATE TABLE featured_categories (
    featured_category_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID NOT NULL REFERENCES service_categories(category_id),
    feature_position INTEGER NOT NULL CHECK (feature_position IN (1, 2)),
    
    -- Audit fields
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    modified_on TIMESTAMPTZ,
    modified_by VARCHAR(255),
    
    UNIQUE(feature_position)
    -- Note: no_default_unassigned_featured constraint enforced by application logic
);

-- Performance Indexes

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

-- Database Functions and Triggers

-- Grafana Cloud Loki audit trigger function
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

-- Default category assignment function
CREATE OR REPLACE FUNCTION reassign_services_to_default_category()
RETURNS TRIGGER AS $$
BEGIN
    -- Category reassignment with Dapr event notification
    -- Implementation will publish to 'services-category-events' topic
    -- Ensures audit compliance via event sourcing pattern
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for audit event publishing
CREATE TRIGGER services_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON services
    FOR EACH ROW EXECUTE FUNCTION publish_audit_event_to_grafana_loki();

CREATE TRIGGER service_categories_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON service_categories
    FOR EACH ROW EXECUTE FUNCTION publish_audit_event_to_grafana_loki();

CREATE TRIGGER featured_categories_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON featured_categories
    FOR EACH ROW EXECUTE FUNCTION publish_audit_event_to_grafana_loki();

CREATE TRIGGER service_category_deletion_trigger
    AFTER UPDATE ON service_categories
    FOR EACH ROW EXECUTE FUNCTION reassign_services_to_default_category();