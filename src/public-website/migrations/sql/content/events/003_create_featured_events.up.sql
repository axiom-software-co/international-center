-- Create featured_events table matching TABLES-EVENTS.md specification
CREATE TABLE featured_events (
    featured_event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(event_id),
    
    -- Audit fields
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    modified_on TIMESTAMPTZ,
    modified_by VARCHAR(255),
    
    -- Business rule constraints
    CONSTRAINT only_one_featured_event CHECK (
        (SELECT COUNT(*) FROM featured_events) <= 1
    ),
    CONSTRAINT no_default_unassigned_featured CHECK (
        NOT EXISTS (
            SELECT 1 FROM events e
            JOIN event_categories ec ON e.category_id = ec.category_id
            WHERE e.event_id = featured_events.event_id 
            AND ec.is_default_unassigned = TRUE
        )
    )
);