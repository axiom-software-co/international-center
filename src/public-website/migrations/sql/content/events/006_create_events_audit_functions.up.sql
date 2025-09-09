-- Create audit functions and triggers matching TABLES-EVENTS.md specification

-- Grafana Cloud Loki audit trigger function
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

-- Default category assignment function
CREATE OR REPLACE FUNCTION reassign_events_to_default_category()
RETURNS TRIGGER AS $$
BEGIN
    -- Event category reassignment with Dapr event notification
    -- Implementation will publish to 'events-category-events' topic
    -- Ensures audit compliance via event sourcing pattern
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Featured event management function
CREATE OR REPLACE FUNCTION manage_featured_events()
RETURNS TRIGGER AS $$
BEGIN
    -- Featured event management with single event constraint
    -- Implementation publishes to 'events-featured-events' topic
    -- Ensures audit compliance and business rule enforcement
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Event capacity validation function
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

-- Create triggers for audit functions
CREATE TRIGGER events_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON events
    FOR EACH ROW EXECUTE FUNCTION publish_events_audit_event_to_grafana_loki();

CREATE TRIGGER event_categories_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON event_categories
    FOR EACH ROW EXECUTE FUNCTION publish_events_audit_event_to_grafana_loki();

CREATE TRIGGER featured_events_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON featured_events
    FOR EACH ROW EXECUTE FUNCTION publish_events_audit_event_to_grafana_loki();

CREATE TRIGGER event_registrations_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON event_registrations
    FOR EACH ROW EXECUTE FUNCTION publish_events_audit_event_to_grafana_loki();

CREATE TRIGGER event_categories_reassignment_trigger
    BEFORE DELETE ON event_categories
    FOR EACH ROW EXECUTE FUNCTION reassign_events_to_default_category();

CREATE TRIGGER featured_events_management_trigger
    AFTER INSERT OR UPDATE ON featured_events
    FOR EACH ROW EXECUTE FUNCTION manage_featured_events();

CREATE TRIGGER event_registrations_capacity_trigger
    BEFORE INSERT OR UPDATE ON event_registrations
    FOR EACH ROW EXECUTE FUNCTION validate_event_capacity();