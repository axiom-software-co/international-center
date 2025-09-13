-- Create featured_events table matching TABLES-EVENTS.md specification
CREATE TABLE featured_events (
    featured_event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(event_id),
    
    -- Audit fields
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    modified_on TIMESTAMPTZ,
    modified_by VARCHAR(255),
    
    -- TODO: Implement constraint validation using database triggers or application logic
    -- Business rule: only one featured event allowed
    -- Business rule: featured events cannot reference default unassigned categories
);