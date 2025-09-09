# Internal Notifications Subscribers Tables

This document defines the PostgreSQL table schemas for managing internal staff notification subscriber configurations and system event routing.

## Overview

The internal notifications subscribers domain handles configuration of staff members who receive system notifications based on events occurring within the platform, including inquiry submissions, system alerts, capacity warnings, and administrative actions requiring attention.

## Tables

### notification_subscribers

Main table for internal staff notification subscription configuration used by dapr APIs for event routing through pub/sub notification router.

```sql
CREATE TABLE notification_subscribers (
    subscriber_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended')),
    
    -- Subscriber Information
    subscriber_name VARCHAR(100) NOT NULL,
    email VARCHAR(254) NOT NULL,
    phone VARCHAR(20),
    
    -- Notification Configuration
    event_types TEXT[] NOT NULL CHECK (array_length(event_types, 1) > 0),
    notification_methods TEXT[] NOT NULL CHECK (array_length(notification_methods, 1) > 0),
    notification_schedule VARCHAR(20) NOT NULL DEFAULT 'immediate' CHECK (notification_schedule IN ('immediate', 'hourly', 'daily')),
    priority_threshold VARCHAR(10) NOT NULL DEFAULT 'low' CHECK (priority_threshold IN ('low', 'medium', 'high', 'urgent')),
    
    -- Metadata
    notes TEXT,
    
    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100) NOT NULL DEFAULT 'system',
    updated_by VARCHAR(100) NOT NULL DEFAULT 'system',
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for performance
CREATE INDEX idx_notification_subscribers_status ON notification_subscribers(status) WHERE NOT is_deleted;
CREATE INDEX idx_notification_subscribers_email ON notification_subscribers(email) WHERE NOT is_deleted;
CREATE INDEX idx_notification_subscribers_event_types ON notification_subscribers USING GIN(event_types) WHERE NOT is_deleted;
CREATE INDEX idx_notification_subscribers_priority_threshold ON notification_subscribers(priority_threshold) WHERE NOT is_deleted;
CREATE INDEX idx_notification_subscribers_notification_methods ON notification_subscribers USING GIN(notification_methods) WHERE NOT is_deleted;
CREATE INDEX idx_notification_subscribers_schedule ON notification_subscribers(notification_schedule) WHERE NOT is_deleted;
CREATE INDEX idx_notification_subscribers_created_at ON notification_subscribers(created_at) WHERE NOT is_deleted;

```




## Constraints and Business Rules

- `subscriber_name` must be at least 2 characters (validated in application layer)
- `email` must be a valid email format (validated in application layer)
- `phone` is optional but if provided must be a valid 10-digit USA format (validated in application layer)
- `event_types` must contain at least one valid event type from the allowed values
- `notification_methods` must contain at least one valid method from: 'email', 'sms', 'both'
- `event_types` must be from valid enum values: 'inquiry-media', 'inquiry-business', 'inquiry-donations', 'inquiry-volunteers', 'event-registration', 'system-error', 'capacity-alert', 'admin-action-required', 'compliance-alert'
- `notification_schedule` must be one of: 'immediate', 'hourly', 'daily'
- `priority_threshold` must be one of: 'low', 'medium', 'high', 'urgent'
- Only active subscribers receive notifications (enforced by application layer)

## Notes

- Table follows established domain patterns with UUID primary keys and standard audit fields
- PostgreSQL arrays used for event_types and notification_methods to support multiple values
- GIN indexes on array fields provide efficient querying for event routing
- Phone numbers are optional to accommodate email-only notification preferences
- Admin gateway provides full CRUD operations for subscriber management
- Dapr APIs consume subscriber configurations via read-only queries for notification routing
- Pub/sub notification router queries subscribers based on event type and priority threshold
- Soft delete pattern maintains audit integrity for notification history tracking