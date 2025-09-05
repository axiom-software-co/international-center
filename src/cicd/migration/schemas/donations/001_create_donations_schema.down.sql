-- Rollback Donations Inquiries Domain Database Schema
-- Drop in reverse order of creation to handle dependencies

-- Drop triggers first
DROP TRIGGER IF EXISTS donations_inquiries_audit_trigger ON donations_inquiries;

-- Drop functions
DROP FUNCTION IF EXISTS publish_donations_audit_event_to_grafana_loki();

-- Drop indexes
DROP INDEX IF EXISTS idx_donations_inquiries_amount_range;
DROP INDEX IF EXISTS idx_donations_inquiries_interest_area;
DROP INDEX IF EXISTS idx_donations_inquiries_donor_type;
DROP INDEX IF EXISTS idx_donations_inquiries_email;
DROP INDEX IF EXISTS idx_donations_inquiries_created_at;
DROP INDEX IF EXISTS idx_donations_inquiries_priority;
DROP INDEX IF EXISTS idx_donations_inquiries_status;

-- Drop tables
DROP TABLE IF EXISTS donations_inquiries;