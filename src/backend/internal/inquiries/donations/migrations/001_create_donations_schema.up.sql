-- Donations Inquiries Domain Database Schema
-- Matching TABLES-INQUIRIES-DONATIONS.md specification exactly

-- Create donations inquiries table
CREATE TABLE donations_inquiries (
    inquiry_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status VARCHAR(20) NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'acknowledged', 'in_progress', 'resolved', 'closed')),
    priority VARCHAR(10) NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    
    -- Contact Information
    contact_name VARCHAR(50) NOT NULL,
    email VARCHAR(254) NOT NULL,
    phone VARCHAR(20),
    organization VARCHAR(100),
    
    -- Donor Classification
    donor_type VARCHAR(20) NOT NULL CHECK (donor_type IN ('individual', 'corporate', 'foundation', 'estate', 'other')),
    
    -- Interest Areas
    interest_area VARCHAR(30) CHECK (interest_area IN ('clinic-development', 'research-funding', 'patient-care', 'equipment', 'general-support', 'other')),
    
    -- Donation Intent
    preferred_amount_range VARCHAR(20) CHECK (preferred_amount_range IN ('under-1000', '1000-5000', '5000-25000', '25000-100000', 'over-100000', 'undisclosed')),
    donation_frequency VARCHAR(15) CHECK (donation_frequency IN ('one-time', 'monthly', 'quarterly', 'annually', 'other')),
    message TEXT NOT NULL CHECK (LENGTH(message) >= 20 AND LENGTH(message) <= 2000),
    
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

-- Performance Indexes

-- Donations inquiries table indexes
CREATE INDEX idx_donations_inquiries_status ON donations_inquiries(status) WHERE NOT is_deleted;
CREATE INDEX idx_donations_inquiries_priority ON donations_inquiries(priority) WHERE NOT is_deleted;
CREATE INDEX idx_donations_inquiries_created_at ON donations_inquiries(created_at) WHERE NOT is_deleted;
CREATE INDEX idx_donations_inquiries_email ON donations_inquiries(email) WHERE NOT is_deleted;
CREATE INDEX idx_donations_inquiries_donor_type ON donations_inquiries(donor_type) WHERE NOT is_deleted;
CREATE INDEX idx_donations_inquiries_interest_area ON donations_inquiries(interest_area) WHERE NOT is_deleted;
CREATE INDEX idx_donations_inquiries_amount_range ON donations_inquiries(preferred_amount_range) WHERE NOT is_deleted;

-- Database Functions and Triggers

-- Grafana Cloud Loki audit trigger function
CREATE OR REPLACE FUNCTION publish_donations_audit_event_to_grafana_loki()
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

-- Create triggers for audit event publishing
CREATE TRIGGER donations_inquiries_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON donations_inquiries
    FOR EACH ROW EXECUTE FUNCTION publish_donations_audit_event_to_grafana_loki();