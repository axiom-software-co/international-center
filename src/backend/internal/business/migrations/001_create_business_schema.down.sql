-- Rollback Business Inquiries Domain Database Schema
-- Drop in reverse order of creation to handle dependencies

-- Drop triggers first
DROP TRIGGER IF EXISTS business_inquiries_audit_trigger ON business_inquiries;

-- Drop functions
DROP FUNCTION IF EXISTS publish_business_audit_event_to_grafana_loki();

-- Drop indexes
DROP INDEX IF EXISTS idx_business_inquiries_inquiry_type;
DROP INDEX IF EXISTS idx_business_inquiries_organization;
DROP INDEX IF EXISTS idx_business_inquiries_email;
DROP INDEX IF EXISTS idx_business_inquiries_created_at;
DROP INDEX IF EXISTS idx_business_inquiries_priority;
DROP INDEX IF EXISTS idx_business_inquiries_status;

-- Drop tables
DROP TABLE IF EXISTS business_inquiries;