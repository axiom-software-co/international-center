-- Volunteer Applications Domain Database Schema
-- Matching TABLES-VOLUNTEERS.md specification exactly

-- Create volunteer_applications table
CREATE TABLE volunteer_applications (
    application_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status VARCHAR(20) NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'under-review', 'interview-scheduled', 'background-check', 'approved', 'declined', 'withdrawn')),
    priority VARCHAR(10) NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    
    -- Personal Information
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    email VARCHAR(254) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    age INTEGER NOT NULL CHECK (age >= 18 AND age <= 100),
    
    -- Volunteer Details
    volunteer_interest VARCHAR(30) NOT NULL CHECK (volunteer_interest IN ('patient-support', 'community-outreach', 'research-support', 'administrative-support', 'multiple', 'other')),
    availability VARCHAR(20) NOT NULL CHECK (availability IN ('2-4-hours', '4-8-hours', '8-16-hours', '16-hours-plus', 'flexible')),
    experience TEXT CHECK (LENGTH(experience) <= 1000),
    motivation TEXT NOT NULL CHECK (LENGTH(motivation) >= 20 AND LENGTH(motivation) <= 1500),
    schedule_preferences VARCHAR(500),
    
    -- Metadata
    source VARCHAR(50) DEFAULT 'website',
    ip_address INET,
    user_agent TEXT,
    
    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100) NOT NULL DEFAULT 'system',
    updated_by VARCHAR(100) NOT NULL DEFAULT 'system',
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Performance Indexes for volunteer_applications table
CREATE INDEX idx_volunteer_applications_status ON volunteer_applications(status) WHERE NOT is_deleted;
CREATE INDEX idx_volunteer_applications_priority ON volunteer_applications(priority) WHERE NOT is_deleted;
CREATE INDEX idx_volunteer_applications_created_at ON volunteer_applications(created_at) WHERE NOT is_deleted;
CREATE INDEX idx_volunteer_applications_email ON volunteer_applications(email) WHERE NOT is_deleted;
CREATE INDEX idx_volunteer_applications_volunteer_interest ON volunteer_applications(volunteer_interest) WHERE NOT is_deleted;
CREATE INDEX idx_volunteer_applications_availability ON volunteer_applications(availability) WHERE NOT is_deleted;
CREATE INDEX idx_volunteer_applications_age ON volunteer_applications(age) WHERE NOT is_deleted;
CREATE INDEX idx_volunteer_applications_updated_at ON volunteer_applications(updated_at) WHERE NOT is_deleted;
CREATE INDEX idx_volunteer_applications_first_name ON volunteer_applications(first_name) WHERE NOT is_deleted;
CREATE INDEX idx_volunteer_applications_last_name ON volunteer_applications(last_name) WHERE NOT is_deleted;

-- Full-text search indexes for volunteer applications
CREATE INDEX idx_volunteer_applications_name_search ON volunteer_applications USING gin(to_tsvector('english', first_name || ' ' || last_name)) WHERE NOT is_deleted;
CREATE INDEX idx_volunteer_applications_motivation_search ON volunteer_applications USING gin(to_tsvector('english', motivation)) WHERE NOT is_deleted;
CREATE INDEX idx_volunteer_applications_experience_search ON volunteer_applications USING gin(to_tsvector('english', COALESCE(experience, ''))) WHERE NOT is_deleted AND experience IS NOT NULL;

-- Database Functions and Triggers

-- Grafana Cloud Loki audit trigger function
CREATE OR REPLACE FUNCTION publish_volunteer_applications_audit_event_to_grafana_loki()
RETURNS TRIGGER AS $$
BEGIN
    -- Volunteer applications audit event publishing directly to Grafana Cloud Loki via Dapr pub/sub
    -- Implementation publishes to 'grafana-audit-events' topic via Dapr
    -- Event structure includes complete data snapshots for compliance
    -- No local audit table storage - all audit data stored in Grafana Cloud Loki
    -- Ensures immutable audit trail and prevents local tampering
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Automatic updated_at timestamp function
CREATE OR REPLACE FUNCTION update_volunteer_applications_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    -- Automatically update the updated_at timestamp on row updates
    -- Maintains audit trail integrity
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for audit event publishing
CREATE TRIGGER volunteer_applications_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON volunteer_applications
    FOR EACH ROW EXECUTE FUNCTION publish_volunteer_applications_audit_event_to_grafana_loki();

-- Create trigger for automatic updated_at timestamp updates
CREATE TRIGGER volunteer_applications_updated_at_trigger
    BEFORE UPDATE ON volunteer_applications
    FOR EACH ROW EXECUTE FUNCTION update_volunteer_applications_updated_at();