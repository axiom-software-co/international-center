-- Create event_registrations table matching TABLES-EVENTS.md specification
CREATE TABLE event_registrations (
    registration_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(event_id),
    participant_name VARCHAR(255) NOT NULL,
    participant_email VARCHAR(254) NOT NULL,
    participant_phone VARCHAR(20),
    registration_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    registration_status VARCHAR(20) NOT NULL DEFAULT 'registered' CHECK (registration_status IN ('registered', 'confirmed', 'cancelled', 'no_show')),
    
    -- Special requirements or notes
    special_requirements TEXT,
    dietary_restrictions TEXT,
    accessibility_needs TEXT,
    
    -- Audit fields
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    modified_on TIMESTAMPTZ,
    modified_by VARCHAR(255),
    
    -- Soft delete fields
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_on TIMESTAMPTZ,
    deleted_by VARCHAR(255),
    
    -- Unique constraint to prevent duplicate registrations
    CONSTRAINT unique_event_participant UNIQUE (event_id, participant_email)
);