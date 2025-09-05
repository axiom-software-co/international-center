-- Rollback Media Inquiries Domain Database Schema
-- Drop in reverse order of creation to handle dependencies

-- Drop triggers first
DROP TRIGGER IF EXISTS media_inquiries_audit_trigger ON media_inquiries;

-- Drop functions
DROP FUNCTION IF EXISTS publish_media_audit_event_to_grafana_loki();

-- Drop indexes
DROP INDEX IF EXISTS idx_media_inquiries_deadline;
DROP INDEX IF EXISTS idx_media_inquiries_media_type;
DROP INDEX IF EXISTS idx_media_inquiries_outlet;
DROP INDEX IF EXISTS idx_media_inquiries_email;
DROP INDEX IF EXISTS idx_media_inquiries_created_at;
DROP INDEX IF EXISTS idx_media_inquiries_urgency;
DROP INDEX IF EXISTS idx_media_inquiries_priority;
DROP INDEX IF EXISTS idx_media_inquiries_status;

-- Drop tables
DROP TABLE IF EXISTS media_inquiries;