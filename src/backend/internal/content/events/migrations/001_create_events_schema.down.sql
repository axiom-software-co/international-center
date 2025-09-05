-- Rollback Events Domain Database Schema
-- Drop in reverse order of creation to handle dependencies

-- Drop triggers first
DROP TRIGGER IF EXISTS event_category_deletion_trigger ON event_categories;
DROP TRIGGER IF EXISTS event_registrations_audit_trigger ON event_registrations;
DROP TRIGGER IF EXISTS featured_events_audit_trigger ON featured_events;
DROP TRIGGER IF EXISTS event_categories_audit_trigger ON event_categories;
DROP TRIGGER IF EXISTS events_audit_trigger ON events;

-- Drop functions
DROP FUNCTION IF EXISTS reassign_events_to_default_category();
DROP FUNCTION IF EXISTS publish_events_audit_event_to_grafana_loki();

-- Drop indexes
DROP INDEX IF EXISTS idx_event_registrations_timestamp;
DROP INDEX IF EXISTS idx_event_registrations_registration_status;
DROP INDEX IF EXISTS idx_event_registrations_participant_email;
DROP INDEX IF EXISTS idx_event_registrations_event_id;
DROP INDEX IF EXISTS idx_featured_events_event_id;
DROP INDEX IF EXISTS idx_event_categories_default;
DROP INDEX IF EXISTS idx_event_categories_slug;
DROP INDEX IF EXISTS idx_events_registration_deadline;
DROP INDEX IF EXISTS idx_events_organizer_name;
DROP INDEX IF EXISTS idx_events_event_date;
DROP INDEX IF EXISTS idx_events_priority_level;
DROP INDEX IF EXISTS idx_events_event_type;
DROP INDEX IF EXISTS idx_events_slug;
DROP INDEX IF EXISTS idx_events_registration_status;
DROP INDEX IF EXISTS idx_events_publishing_status;
DROP INDEX IF EXISTS idx_events_category_id;

-- Drop tables in dependency order
DROP TABLE IF EXISTS event_registrations;
DROP TABLE IF EXISTS featured_events;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS event_categories;