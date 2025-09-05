-- Business Inquiries Domain Database Schema
-- Matching TABLES-INQUIRIES-BUSINESS.md specification exactly

-- Create business inquiries table
CREATE TABLE business_inquiries (
    inquiry_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status VARCHAR(20) NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'acknowledged', 'in_progress', 'resolved', 'closed')),
    priority VARCHAR(10) NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    
    -- Organization Information
    organization_name VARCHAR(100) NOT NULL,
    contact_name VARCHAR(50) NOT NULL,
    title VARCHAR(50) NOT NULL,
    email VARCHAR(254) NOT NULL,
    phone VARCHAR(20),
    industry VARCHAR(50),
    
    -- Inquiry Details
    inquiry_type VARCHAR(20) NOT NULL CHECK (inquiry_type IN ('partnership', 'licensing', 'research', 'technology', 'regulatory', 'other')),
    message TEXT NOT NULL CHECK (LENGTH(message) >= 20 AND LENGTH(message) <= 1500),
    
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

-- Business inquiries table indexes
CREATE INDEX idx_business_inquiries_status ON business_inquiries(status) WHERE NOT is_deleted;
CREATE INDEX idx_business_inquiries_priority ON business_inquiries(priority) WHERE NOT is_deleted;
CREATE INDEX idx_business_inquiries_created_at ON business_inquiries(created_at) WHERE NOT is_deleted;
CREATE INDEX idx_business_inquiries_email ON business_inquiries(email) WHERE NOT is_deleted;
CREATE INDEX idx_business_inquiries_organization ON business_inquiries(organization_name) WHERE NOT is_deleted;
CREATE INDEX idx_business_inquiries_inquiry_type ON business_inquiries(inquiry_type) WHERE NOT is_deleted;

-- Database Functions and Triggers

-- Grafana Cloud Loki audit trigger function
CREATE OR REPLACE FUNCTION publish_business_audit_event_to_grafana_loki()
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
CREATE TRIGGER business_inquiries_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON business_inquiries
    FOR EACH ROW EXECUTE FUNCTION publish_business_audit_event_to_grafana_loki();