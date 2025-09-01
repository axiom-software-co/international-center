-- Create services database tables - Exact match to SERVICES-SCHEMA.md

-- Services table
CREATE TABLE services (
    service_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    content_url VARCHAR(500), -- URL to Azure Blob Storage content
    category_id UUID NOT NULL,
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

-- Service categories table
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
    deleted_by VARCHAR(255),
    
    -- Complex constraint for single default category enforced at application level
);

-- Featured categories table
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
    -- Note: Complex constraint for preventing default_unassigned categories from being featured
    -- is enforced at application level due to PostgreSQL subquery limitation in CHECK constraints
);

-- Add foreign key constraint for services table
ALTER TABLE services ADD CONSTRAINT fk_services_category_id FOREIGN KEY (category_id) REFERENCES service_categories(category_id);

-- Performance indexes per SERVICES-SCHEMA.md
CREATE INDEX idx_services_category_id ON services(category_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_services_publishing_status ON services(publishing_status) WHERE is_deleted = FALSE;
CREATE INDEX idx_services_slug ON services(slug) WHERE is_deleted = FALSE;
CREATE INDEX idx_services_order_category ON services(category_id, order_number) WHERE is_deleted = FALSE;
CREATE INDEX idx_services_delivery_mode ON services(delivery_mode) WHERE is_deleted = FALSE;

CREATE INDEX idx_service_categories_slug ON service_categories(slug) WHERE is_deleted = FALSE;
CREATE INDEX idx_service_categories_order ON service_categories(order_number) WHERE is_deleted = FALSE;
CREATE INDEX idx_service_categories_default ON service_categories(is_default_unassigned) WHERE is_deleted = FALSE;

CREATE INDEX idx_featured_categories_category_id ON featured_categories(category_id);
CREATE INDEX idx_featured_categories_position ON featured_categories(feature_position);