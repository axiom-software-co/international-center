-- Drop audit functions and triggers

-- Drop triggers first
DROP TRIGGER IF EXISTS events_audit_trigger ON events;
DROP TRIGGER IF EXISTS event_categories_audit_trigger ON event_categories;
DROP TRIGGER IF EXISTS featured_events_audit_trigger ON featured_events;
DROP TRIGGER IF EXISTS event_registrations_audit_trigger ON event_registrations;
DROP TRIGGER IF EXISTS event_categories_reassignment_trigger ON event_categories;
DROP TRIGGER IF EXISTS featured_events_management_trigger ON featured_events;
DROP TRIGGER IF EXISTS event_registrations_capacity_trigger ON event_registrations;

-- Drop functions
DROP FUNCTION IF EXISTS publish_events_audit_event_to_grafana_loki();
DROP FUNCTION IF EXISTS reassign_events_to_default_category();
DROP FUNCTION IF EXISTS manage_featured_events();
DROP FUNCTION IF EXISTS validate_event_capacity();