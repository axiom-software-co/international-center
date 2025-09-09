-- Create events table matching TABLES-EVENTS.md specification
CREATE TABLE events (
    event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    content TEXT,
    slug VARCHAR(255) UNIQUE NOT NULL,
    category_id UUID NOT NULL REFERENCES event_categories(category_id),
    image_url VARCHAR(500),
    organizer_name VARCHAR(255),
    event_date DATE NOT NULL,
    event_time TIME,
    end_date DATE,
    end_time TIME,
    location VARCHAR(500) NOT NULL,
    virtual_link VARCHAR(500),
    max_capacity INTEGER,
    registration_deadline TIMESTAMPTZ,
    registration_status VARCHAR(20) NOT NULL DEFAULT 'open' CHECK (registration_status IN ('open', 'registration_required', 'full', 'cancelled')),
    publishing_status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (publishing_status IN ('draft', 'published', 'archived')),
    
    -- Content metadata
    tags TEXT[],
    event_type VARCHAR(50) NOT NULL CHECK (event_type IN ('workshop', 'seminar', 'webinar', 'conference', 'fundraiser', 'community', 'medical', 'educational')),
    priority_level VARCHAR(20) NOT NULL DEFAULT 'normal' CHECK (priority_level IN ('low', 'normal', 'high', 'urgent')),
    
    -- Audit fields
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    modified_on TIMESTAMPTZ,
    modified_by VARCHAR(255),
    
    -- Soft delete fields
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_on TIMESTAMPTZ,
    deleted_by VARCHAR(255),
    
    -- Business rule constraints
    CONSTRAINT events_end_date_after_start CHECK (end_date IS NULL OR end_date >= event_date),
    CONSTRAINT events_registration_deadline_before_event CHECK (registration_deadline IS NULL OR registration_deadline::DATE <= event_date),
    CONSTRAINT events_virtual_link_https CHECK (virtual_link IS NULL OR virtual_link LIKE 'https://%'),
    CONSTRAINT events_image_url_https CHECK (image_url IS NULL OR image_url LIKE 'https://%')
);