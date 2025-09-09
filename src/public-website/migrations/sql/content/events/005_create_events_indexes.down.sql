-- Drop all events domain indexes

-- Events table indexes
DROP INDEX IF EXISTS idx_events_category_id;
DROP INDEX IF EXISTS idx_events_publishing_status;
DROP INDEX IF EXISTS idx_events_registration_status;
DROP INDEX IF EXISTS idx_events_slug;
DROP INDEX IF EXISTS idx_events_event_type;
DROP INDEX IF EXISTS idx_events_priority_level;
DROP INDEX IF EXISTS idx_events_event_date;
DROP INDEX IF EXISTS idx_events_organizer_name;
DROP INDEX IF EXISTS idx_events_registration_deadline;

-- Event categories table indexes
DROP INDEX IF EXISTS idx_event_categories_slug;
DROP INDEX IF EXISTS idx_event_categories_default;

-- Featured events table indexes
DROP INDEX IF EXISTS idx_featured_events_event_id;

-- Event registrations table indexes
DROP INDEX IF EXISTS idx_event_registrations_event_id;
DROP INDEX IF EXISTS idx_event_registrations_participant_email;
DROP INDEX IF EXISTS idx_event_registrations_registration_status;
DROP INDEX IF EXISTS idx_event_registrations_timestamp;