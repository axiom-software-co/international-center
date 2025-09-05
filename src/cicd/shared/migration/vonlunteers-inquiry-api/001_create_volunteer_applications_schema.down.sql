-- Volunteer Applications Domain Database Schema Rollback
-- Removes all volunteer applications domain tables and related database objects

-- Drop triggers first (to avoid dependency issues)
DROP TRIGGER IF EXISTS volunteer_applications_updated_at_trigger ON volunteer_applications;
DROP TRIGGER IF EXISTS volunteer_applications_audit_trigger ON volunteer_applications;

-- Drop functions
DROP FUNCTION IF EXISTS update_volunteer_applications_updated_at();
DROP FUNCTION IF EXISTS publish_volunteer_applications_audit_event_to_grafana_loki();

-- Drop indexes (PostgreSQL will auto-drop with tables, but explicit for clarity)
DROP INDEX IF EXISTS idx_volunteer_applications_experience_search;
DROP INDEX IF EXISTS idx_volunteer_applications_motivation_search;
DROP INDEX IF EXISTS idx_volunteer_applications_name_search;
DROP INDEX IF EXISTS idx_volunteer_applications_last_name;
DROP INDEX IF EXISTS idx_volunteer_applications_first_name;
DROP INDEX IF EXISTS idx_volunteer_applications_updated_at;
DROP INDEX IF EXISTS idx_volunteer_applications_age;
DROP INDEX IF EXISTS idx_volunteer_applications_availability;
DROP INDEX IF EXISTS idx_volunteer_applications_volunteer_interest;
DROP INDEX IF EXISTS idx_volunteer_applications_email;
DROP INDEX IF EXISTS idx_volunteer_applications_created_at;
DROP INDEX IF EXISTS idx_volunteer_applications_priority;
DROP INDEX IF EXISTS idx_volunteer_applications_status;

-- Drop tables
DROP TABLE IF EXISTS volunteer_applications;