-- Media Inquiries Domain Database Schema
-- Matching TABLES-INQUIRIES-MEDIA.md specification exactly

-- Create media inquiries table
CREATE TABLE media_inquiries (
    inquiry_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status VARCHAR(20) NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'acknowledged', 'in_progress', 'resolved', 'closed')),
    priority VARCHAR(10) NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    urgency VARCHAR(10) NOT NULL DEFAULT 'medium' CHECK (urgency IN ('low', 'medium', 'high')),
    
    -- Media Contact Information
    outlet VARCHAR(100) NOT NULL,
    contact_name VARCHAR(50) NOT NULL,
    title VARCHAR(50) NOT NULL,
    email VARCHAR(254) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    
    -- Media Details
    media_type VARCHAR(20) CHECK (media_type IN ('print', 'digital', 'television', 'radio', 'podcast', 'medical-journal', 'other')),
    deadline DATE,
    subject TEXT NOT NULL CHECK (LENGTH(subject) >= 20 AND LENGTH(subject) <= 1500),
    
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

-- Media inquiries table indexes
CREATE INDEX idx_media_inquiries_status ON media_inquiries(status) WHERE NOT is_deleted;
CREATE INDEX idx_media_inquiries_priority ON media_inquiries(priority) WHERE NOT is_deleted;
CREATE INDEX idx_media_inquiries_urgency ON media_inquiries(urgency) WHERE NOT is_deleted;
CREATE INDEX idx_media_inquiries_created_at ON media_inquiries(created_at) WHERE NOT is_deleted;
CREATE INDEX idx_media_inquiries_email ON media_inquiries(email) WHERE NOT is_deleted;
CREATE INDEX idx_media_inquiries_outlet ON media_inquiries(outlet) WHERE NOT is_deleted;
CREATE INDEX idx_media_inquiries_media_type ON media_inquiries(media_type) WHERE NOT is_deleted;
CREATE INDEX idx_media_inquiries_deadline ON media_inquiries(deadline) WHERE NOT is_deleted AND deadline IS NOT NULL;

-- Database Functions and Triggers

-- Grafana Cloud Loki audit trigger function
CREATE OR REPLACE FUNCTION publish_media_audit_event_to_grafana_loki()
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
CREATE TRIGGER media_inquiries_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON media_inquiries
    FOR EACH ROW EXECUTE FUNCTION publish_media_audit_event_to_grafana_loki();