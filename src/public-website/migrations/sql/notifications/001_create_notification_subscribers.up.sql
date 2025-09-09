-- Create notification_subscribers table matching TABLES-INTERNAL-NOTIFICATIONS-SUBSCRIBERS.md specification
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