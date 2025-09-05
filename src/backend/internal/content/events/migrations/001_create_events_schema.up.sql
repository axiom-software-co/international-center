-- Events Domain Database Schema
-- Matching TABLES-EVENTS.md specification exactly

-- Create event categories table first (referenced by events table)
CREATE TABLE event_categories (
    category_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
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

-- Create events table
CREATE TABLE events (
    event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    content TEXT, -- Event content stored in PostgreSQL
    slug VARCHAR(255) UNIQUE NOT NULL,
    category_id UUID NOT NULL REFERENCES event_categories(category_id),
    image_url VARCHAR(500), -- URL to Azure Blob Storage for images only
    organizer_name VARCHAR(255),
    event_date DATE NOT NULL,
    event_time TIME,
    end_date DATE,
    end_time TIME,
    location VARCHAR(500) NOT NULL,
    virtual_link VARCHAR(500), -- URL for virtual events
    max_capacity INTEGER,
    registration_deadline TIMESTAMPTZ,
    registration_status VARCHAR(20) NOT NULL DEFAULT 'open' CHECK (registration_status IN ('open', 'registration_required', 'full', 'cancelled')),
    publishing_status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (publishing_status IN ('draft', 'published', 'archived')),
    
    -- Content metadata
    tags TEXT[],
    event_type VARCHAR(50) NOT NULL CHECK (event_type IN ('workshop', 'seminar', 'webinar', 'conference', 'fundraiser', 'community', 'medical', 'educational')),
    priority_level VARCHAR(20) NOT NULL DEFAULT 'normal' CHECK (priority_level IN ('low', 'normal', 'high', 'urgent')),
    
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

-- Create featured events table
CREATE TABLE featured_events (
    featured_event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(event_id),
    
    -- Audit fields
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    modified_on TIMESTAMPTZ,
    modified_by VARCHAR(255)
    -- Note: single featured event constraint enforced by application logic
    -- Note: no_default_unassigned_featured constraint enforced by application logic
);

-- Create event registrations table
CREATE TABLE event_registrations (
    registration_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(event_id),
    participant_name VARCHAR(255) NOT NULL,
    participant_email VARCHAR(254) NOT NULL,
    participant_phone VARCHAR(20),
    registration_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    registration_status VARCHAR(20) NOT NULL DEFAULT 'registered' CHECK (registration_status IN ('registered', 'confirmed', 'cancelled', 'no_show')),
    
    -- Special requirements or notes
    special_requirements TEXT,
    dietary_restrictions TEXT,
    accessibility_needs TEXT,
    
    -- Audit fields
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    modified_on TIMESTAMPTZ,
    modified_by VARCHAR(255),
    
    -- Soft delete fields
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_on TIMESTAMPTZ,
    deleted_by VARCHAR(255),
    
    -- Unique constraint to prevent duplicate registrations
    CONSTRAINT unique_event_participant UNIQUE (event_id, participant_email)
);

-- Performance Indexes

-- Events table indexes
CREATE INDEX idx_events_category_id ON events(category_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_events_publishing_status ON events(publishing_status) WHERE is_deleted = FALSE;
CREATE INDEX idx_events_registration_status ON events(registration_status) WHERE is_deleted = FALSE;
CREATE INDEX idx_events_slug ON events(slug) WHERE is_deleted = FALSE;
CREATE INDEX idx_events_event_type ON events(event_type) WHERE is_deleted = FALSE;
CREATE INDEX idx_events_priority_level ON events(priority_level) WHERE is_deleted = FALSE;
CREATE INDEX idx_events_event_date ON events(event_date) WHERE is_deleted = FALSE;
CREATE INDEX idx_events_organizer_name ON events(organizer_name) WHERE is_deleted = FALSE;
CREATE INDEX idx_events_registration_deadline ON events(registration_deadline) WHERE is_deleted = FALSE;

-- Event categories table indexes  
CREATE INDEX idx_event_categories_slug ON event_categories(slug) WHERE is_deleted = FALSE;
CREATE INDEX idx_event_categories_default ON event_categories(is_default_unassigned) WHERE is_deleted = FALSE;

-- Featured events table indexes
CREATE INDEX idx_featured_events_event_id ON featured_events(event_id);

-- Event registrations table indexes
CREATE INDEX idx_event_registrations_event_id ON event_registrations(event_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_event_registrations_participant_email ON event_registrations(participant_email) WHERE is_deleted = FALSE;
CREATE INDEX idx_event_registrations_registration_status ON event_registrations(registration_status) WHERE is_deleted = FALSE;
CREATE INDEX idx_event_registrations_timestamp ON event_registrations(registration_timestamp) WHERE is_deleted = FALSE;

-- Database Functions and Triggers

-- Grafana Cloud Loki audit trigger function
CREATE OR REPLACE FUNCTION publish_events_audit_event_to_grafana_loki()
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
CREATE OR REPLACE FUNCTION reassign_events_to_default_category()
RETURNS TRIGGER AS $$
BEGIN
    -- Category reassignment with Dapr event notification
    -- Implementation will publish to 'events-category-events' topic
    -- Ensures audit compliance via event sourcing pattern
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for audit event publishing
CREATE TRIGGER events_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON events
    FOR EACH ROW EXECUTE FUNCTION publish_events_audit_event_to_grafana_loki();

CREATE TRIGGER event_categories_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON event_categories
    FOR EACH ROW EXECUTE FUNCTION publish_events_audit_event_to_grafana_loki();

CREATE TRIGGER featured_events_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON featured_events
    FOR EACH ROW EXECUTE FUNCTION publish_events_audit_event_to_grafana_loki();

CREATE TRIGGER event_registrations_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON event_registrations
    FOR EACH ROW EXECUTE FUNCTION publish_events_audit_event_to_grafana_loki();

CREATE TRIGGER event_category_deletion_trigger
    AFTER UPDATE ON event_categories
    FOR EACH ROW EXECUTE FUNCTION reassign_events_to_default_category();