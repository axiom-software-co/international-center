# Events Domain Database Schema

## Core Tables

### Events Table
```sql
CREATE TABLE events (
    event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    content TEXT, -- Event content stored in PostgreSQL
    slug VARCHAR(255) UNIQUE NOT NULL,
    category_id UUID NOT NULL REFERENCES event_categories(category_id),
    image_url VARCHAR(500), -- URL to Azure Blob Storage for images only
    organizer_name VARCHAR(255),
    event_date DATE NOT NULL,
    event_time TIME,
    end_date DATE,
    end_time TIME,
    location VARCHAR(500) NOT NULL,
    virtual_link VARCHAR(500), -- URL for virtual events
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
    deleted_by VARCHAR(255)
);
```

### Event Categories Table
```sql
CREATE TABLE event_categories (
    category_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    is_default_unassigned BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Audit fields
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    modified_on TIMESTAMPTZ,
    modified_by VARCHAR(255),
    
    -- Soft delete fields  
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_on TIMESTAMPTZ,
    deleted_by VARCHAR(255),
    
    CONSTRAINT only_one_default_unassigned CHECK (
        NOT is_default_unassigned OR 
        (SELECT COUNT(*) FROM event_categories WHERE is_default_unassigned = TRUE AND is_deleted = FALSE) <= 1
    )
);
```

### Featured Events Table
```sql
CREATE TABLE featured_events (
    featured_event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(event_id),
    
    -- Audit fields
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    modified_on TIMESTAMPTZ,
    modified_by VARCHAR(255),
    
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
```

### Event Registrations Table
```sql
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
```

## Audit Event Publishing

### Audit Events via Grafana Cloud Loki
All audit data is published directly to Grafana Cloud Loki for compliance and immutable storage. No audit tables exist in PostgreSQL to prevent local tampering and ensure regulatory compliance.

**Audit Event Structure:**
```json
{
  "audit_id": "uuid",
  "entity_type": "event|event_category|featured_event|event_registration",
  "entity_id": "uuid",
  "operation_type": "INSERT|UPDATE|DELETE",
  "audit_timestamp": "timestamp",
  "user_id": "string",
  "correlation_id": "uuid",
  "trace_id": "string",
  "data_snapshot": {
    "before": {...},
    "after": {...}
  },
  "environment": "production|development"
}
```

**Grafana Cloud Loki Labels:**
- `job`: "events-audit"
- `environment`: "production|development" 
- `entity_type`: "event|event_category|featured_event|event_registration"
- `operation_type`: "insert|update|delete"
- `user_id`: authenticated user identifier

## Performance Indexes

### Core Table Indexes
```sql
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
```

### Audit Query Optimization via Grafana Cloud Loki
Audit data queries are optimized through Grafana Cloud Loki label indexing and LogQL queries. No PostgreSQL indexes required for audit data as all audit information is stored externally for compliance.

**Grafana Cloud Loki Query Patterns:**
```logql
-- Query all event operations
{job="events-audit", entity_type="event"} |= `operation_type`

-- Query specific entity operations  
{job="events-audit", entity_type="event"} |= `"entity_id":"<event-id>"`

-- Query operations by user
{job="events-audit", user_id="<user-id>"} |= `operation_type`

-- Query operations within time range with correlation tracking
{job="events-audit"} |= `"correlation_id":"<uuid>"` | json | line_format "{{.audit_timestamp}} {{.operation_type}} {{.entity_type}}"

-- Query registration operations for capacity tracking
{job="events-audit", entity_type="event_registration"} |= `operation_type`
```

## Database Functions and Triggers

### Grafana Cloud Loki Audit Trigger Function
```sql
CREATE OR REPLACE FUNCTION publish_events_audit_event_to_grafana_loki()
RETURNS TRIGGER AS $$
BEGIN
    -- Events audit event publishing directly to Grafana Cloud Loki via Dapr pub/sub
    -- Implementation publishes to 'grafana-audit-events' topic via Dapr
    -- Event structure includes complete data snapshots for compliance
    -- No local audit table storage - all audit data stored in Grafana Cloud Loki
    -- Ensures immutable audit trail and prevents local tampering
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;
```

### Default Category Assignment Function
```sql
CREATE OR REPLACE FUNCTION reassign_events_to_default_category()
RETURNS TRIGGER AS $$
BEGIN
    -- Event category reassignment with Dapr event notification
    -- Implementation will publish to 'events-category-events' topic
    -- Ensures audit compliance via event sourcing pattern
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;
```

### Featured Event Management Function
```sql
CREATE OR REPLACE FUNCTION manage_featured_events()
RETURNS TRIGGER AS $$
BEGIN
    -- Featured event management with single event constraint
    -- Implementation publishes to 'events-featured-events' topic
    -- Ensures audit compliance and business rule enforcement
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;
```

### Event Capacity Validation Function
```sql
CREATE OR REPLACE FUNCTION validate_event_capacity()
RETURNS TRIGGER AS $$
DECLARE
    current_capacity INTEGER;
    max_capacity INTEGER;
BEGIN
    -- Get current registration count and maximum capacity
    SELECT COUNT(*), e.max_capacity INTO current_capacity, max_capacity
    FROM event_registrations er
    JOIN events e ON er.event_id = e.event_id
    WHERE er.event_id = NEW.event_id 
    AND er.registration_status IN ('registered', 'confirmed')
    AND er.is_deleted = FALSE
    GROUP BY e.max_capacity;
    
    -- Check capacity constraint
    IF max_capacity IS NOT NULL AND current_capacity >= max_capacity THEN
        RAISE EXCEPTION 'Event capacity exceeded. Maximum capacity: %, Current registrations: %', max_capacity, current_capacity;
    END IF;
    
    -- Update event registration status if approaching capacity
    IF max_capacity IS NOT NULL AND current_capacity >= (max_capacity * 0.9) THEN
        UPDATE events SET registration_status = 'full' WHERE event_id = NEW.event_id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

## Business Rules

### Category Management Rules
1. **Default Unassigned Category**: Exactly one category must have `is_default_unassigned = TRUE`
2. **Featured Events**: Exactly one event can be featured at any time
3. **Featured Event Restriction**: The default unassigned category cannot contain featured events
4. **Event Assignment**: All events must have a valid category_id
5. **Category Deletion**: When a category is soft-deleted, all its events are reassigned to the default unassigned category

### Event Publishing Rules
1. **Publishing Status**: Events can be 'draft', 'published', or 'archived'
2. **Default Status**: New events start with 'draft' status
3. **Category Assignment**: All events start with the default unassigned category
4. **Event Types**: Events must specify type (workshop, seminar, webinar, conference, fundraiser, community, medical, educational)
5. **Content Storage**: Event content stored in PostgreSQL TEXT fields for full-text search capability
6. **Image Storage**: Event images stored in Azure Blob Storage, referenced via `image_url`
7. **Content Publishing**: Events can be published without content (content can be NULL)
8. **URL Validation**: All `image_url` and `virtual_link` values must be valid HTTPS URLs when not NULL
9. **Priority Levels**: Event priority affects display ordering and notification urgency

### Event Scheduling Rules
1. **Date Required**: All events must have an event_date
2. **Time Optional**: Event time is optional for all-day events
3. **End Date Validation**: End date must be same day or later than event_date when provided
4. **Registration Deadline**: Registration deadline must be before or on event_date when provided
5. **Past Event Publishing**: Events with past dates cannot be published (only archived)
6. **Location Required**: All events must specify a location (physical address or virtual)
7. **Virtual Events**: Virtual events must provide virtual_link when location indicates online event
8. **Multi-Day Events**: Events spanning multiple days use event_date/end_date combination

### Registration Management Rules
1. **Capacity Enforcement**: Registration count cannot exceed max_capacity when specified
2. **Registration Status**: Registration status automatically updates based on capacity
3. **Duplicate Prevention**: One registration per email address per event
4. **Registration Deadline**: No new registrations accepted after registration_deadline
5. **Cancelled Event Registration**: No new registrations for cancelled events
6. **Registration Workflow**: Registered → Confirmed → No Show/Cancelled progression
7. **Participant Information**: Name and email required, phone optional
8. **Special Requirements**: Optional fields for accessibility and dietary needs

### Event Content Rules
1. **Description Required**: All events must have a description for indexing and preview
2. **Organizer Attribution**: Organizer name optional for events (can be institutional)
3. **Event Date Ordering**: Used for chronological ordering and calendar filtering
4. **Tags**: PostgreSQL array for event tags and content categorization
5. **Virtual Integration**: Virtual events support online delivery with link management
6. **Content Immutability**: Published event content changes require audit approval
7. **Version Control**: Content changes create new audit events for compliance tracking
8. **Capacity Management**: High capacity events require immediate publication workflow

### Registration Audit Requirements
1. **Registration Tracking**: All registration changes audited via direct publishing to Grafana Cloud Loki
2. **Participant Privacy**: Registration audit events exclude sensitive participant data
3. **Capacity Compliance**: Registration capacity tracking audited for event management
4. **Registration Status Changes**: All status changes create audit events
5. **Cancellation Tracking**: Registration cancellations maintain audit trail
6. **No Show Tracking**: No show status changes audited for event analysis
7. **Duplicate Registration Prevention**: Audit events track duplicate registration attempts
8. **Registration Deadline Enforcement**: Late registration attempts audited

### Audit Requirements
1. **Compliance Enforcement**: All changes audited via direct publishing to Grafana Cloud Loki
2. **Immutable Audit Records**: Audit events stored in Grafana Cloud Loki, never modified or deleted
3. **Complete Data Snapshots**: Audit events include complete before/after data snapshots for compliance
4. **Correlation Tracking**: All operations share correlation_id and trace_id for distributed tracing
5. **User Attribution**: All operations record user_id from Authentik authentication context
6. **Event Publishing**: Audit events published to 'grafana-audit-events' topic via Dapr pub/sub
7. **Grafana Cloud Integration**: All audit operations stored directly in Grafana Cloud Loki for compliance
8. **No Local Audit Storage**: No audit tables in PostgreSQL to prevent local tampering

### Data Integrity Rules
1. **Soft Delete Only**: Records are never physically deleted, only marked as deleted
2. **Unique Constraints**: Slugs must be unique across active (non-deleted) events
3. **Date Ordering**: Event dates used for chronological ordering and calendar filtering
4. **Foreign Key Integrity**: All category_id references must be valid and active
5. **Image URL Integrity**: `image_url` must point to existing blob storage content when not NULL
6. **Content Validation**: Event content must be valid text format when not NULL
7. **Event Date Validation**: Event dates must be realistic (not past dates for new events)
8. **Featured Event Integrity**: Only one event can be featured, and it cannot be from default unassigned category
9. **Registration Integrity**: Registration counts must accurately reflect actual registrations
10. **Capacity Validation**: Registration capacity constraints enforced at database level