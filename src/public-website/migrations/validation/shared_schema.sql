-- Shared Domain Validation Schema - Final Desired State
-- This file represents the authoritative schema state after all shared domain migrations are complete
-- Used for validation against deployed schema, not for direct migration execution

-- Audit Events Table
CREATE TABLE audit_events (
    audit_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(100) NOT NULL,
    entity_id UUID NOT NULL,
    operation_type VARCHAR(20) NOT NULL CHECK (operation_type IN ('INSERT', 'UPDATE', 'DELETE')),
    
    -- Audit Metadata
    audit_timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id VARCHAR(100) NOT NULL,
    correlation_id UUID NOT NULL,
    trace_id VARCHAR(100),
    
    -- Data Snapshots for Compliance
    data_before JSONB,
    data_after JSONB,
    
    -- Environment Context
    environment VARCHAR(50) NOT NULL DEFAULT 'development',
    source_service VARCHAR(100) NOT NULL,
    source_ip INET,
    user_agent TEXT,
    
    -- Additional Context
    notes TEXT
);

-- Correlation Tracking Table
CREATE TABLE correlation_tracking (
    correlation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trace_id VARCHAR(100) NOT NULL,
    parent_correlation_id UUID REFERENCES correlation_tracking(correlation_id),
    
    -- Request Context
    service_name VARCHAR(100) NOT NULL,
    operation_name VARCHAR(100) NOT NULL,
    request_path VARCHAR(500),
    http_method VARCHAR(10),
    
    -- Timing Information
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    duration_ms INTEGER,
    
    -- User Context
    user_id VARCHAR(100),
    session_id UUID,
    
    -- Request Metadata
    client_ip INET,
    user_agent TEXT,
    request_headers JSONB,
    
    -- Response Information
    status_code INTEGER,
    error_message TEXT,
    response_size_bytes INTEGER,
    
    -- Environment Context
    environment VARCHAR(50) NOT NULL DEFAULT 'development',
    instance_id VARCHAR(100),
    
    -- Additional Context
    custom_properties JSONB DEFAULT '{}'
);

-- Performance Indexes
CREATE INDEX idx_audit_events_entity_type ON audit_events(entity_type);
CREATE INDEX idx_audit_events_entity_id ON audit_events(entity_id);
CREATE INDEX idx_audit_events_operation ON audit_events(operation_type);
CREATE INDEX idx_audit_events_timestamp ON audit_events(audit_timestamp);
CREATE INDEX idx_audit_events_user_id ON audit_events(user_id);
CREATE INDEX idx_audit_events_correlation_id ON audit_events(correlation_id);
CREATE INDEX idx_audit_events_environment ON audit_events(environment);

CREATE INDEX idx_correlation_tracking_trace_id ON correlation_tracking(trace_id);
CREATE INDEX idx_correlation_tracking_parent ON correlation_tracking(parent_correlation_id);
CREATE INDEX idx_correlation_tracking_service ON correlation_tracking(service_name);
CREATE INDEX idx_correlation_tracking_operation ON correlation_tracking(operation_name);
CREATE INDEX idx_correlation_tracking_started ON correlation_tracking(started_at);
CREATE INDEX idx_correlation_tracking_user_id ON correlation_tracking(user_id);
CREATE INDEX idx_correlation_tracking_environment ON correlation_tracking(environment);