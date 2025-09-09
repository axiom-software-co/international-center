-- Create audit_events table for cross-cutting audit infrastructure
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

-- Partition table by audit_timestamp for performance
-- Note: This would typically be implemented with range partitioning in production