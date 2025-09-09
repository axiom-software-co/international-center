-- Create correlation_tracking table for distributed tracing and request correlation
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