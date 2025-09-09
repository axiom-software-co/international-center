-- Create performance indexes matching TABLES-EVENTS.md specification

-- Events table indexes
CREATE INDEX idx_events_category_id ON events(category_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_events_publishing_status ON events(publishing_status) WHERE is_deleted = FALSE;
CREATE INDEX idx_events_registration_status ON events(registration_status) WHERE is_deleted = FALSE;
CREATE INDEX idx_events_slug ON events(slug) WHERE is_deleted = FALSE;
CREATE INDEX idx_events_event_type ON events(event_type) WHERE is_deleted = FALSE;
CREATE INDEX idx_events_priority_level ON events(priority_level) WHERE is_deleted = FALSE;
CREATE INDEX idx_events_event_date ON events(event_date) WHERE is_deleted = FALSE;
CREATE INDEX idx_events_organizer_name ON events(organizer_name) WHERE is_deleted = FALSE;
CREATE INDEX idx_events_registration_deadline ON events(registration_deadline) WHERE is_deleted = FALSE;

-- Event categories table indexes
CREATE INDEX idx_event_categories_slug ON event_categories(slug) WHERE is_deleted = FALSE;
CREATE INDEX idx_event_categories_default ON event_categories(is_default_unassigned) WHERE is_deleted = FALSE;

-- Featured events table indexes
CREATE INDEX idx_featured_events_event_id ON featured_events(event_id);

-- Event registrations table indexes
CREATE INDEX idx_event_registrations_event_id ON event_registrations(event_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_event_registrations_participant_email ON event_registrations(participant_email) WHERE is_deleted = FALSE;
CREATE INDEX idx_event_registrations_registration_status ON event_registrations(registration_status) WHERE is_deleted = FALSE;
CREATE INDEX idx_event_registrations_timestamp ON event_registrations(registration_timestamp) WHERE is_deleted = FALSE;