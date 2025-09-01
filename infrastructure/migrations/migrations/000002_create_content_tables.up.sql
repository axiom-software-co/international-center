-- Create content database tables - Exact match to CONTENT-SCHEMA.md

-- Content table
CREATE TABLE content (
    content_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    original_filename VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL CHECK (file_size > 0),
    mime_type VARCHAR(100) NOT NULL,
    content_hash VARCHAR(64) NOT NULL, -- SHA-256 hash for integrity
    storage_path VARCHAR(500) NOT NULL, -- Azure Blob Storage path
    upload_status VARCHAR(20) NOT NULL DEFAULT 'processing' CHECK (upload_status IN ('processing', 'available', 'failed', 'archived')),
    
    -- Content metadata
    alt_text VARCHAR(500),
    description TEXT,
    tags TEXT[], -- PostgreSQL array for content tags
    
    -- Content classification
    content_category VARCHAR(50) NOT NULL CHECK (content_category IN ('document', 'image', 'video', 'audio', 'archive')),
    access_level VARCHAR(20) NOT NULL DEFAULT 'internal' CHECK (access_level IN ('public', 'internal', 'restricted')),
    
    -- Upload tracking
    upload_correlation_id UUID NOT NULL,
    processing_attempts INTEGER NOT NULL DEFAULT 0,
    last_processed_at TIMESTAMPTZ,
    
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

-- Content access log table
CREATE TABLE content_access_log (
    access_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_id UUID NOT NULL REFERENCES content(content_id),
    access_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id VARCHAR(255),
    client_ip INET,
    user_agent TEXT,
    access_type VARCHAR(20) NOT NULL CHECK (access_type IN ('view', 'download', 'preview')),
    http_status_code INTEGER,
    bytes_served BIGINT,
    response_time_ms INTEGER,
    
    -- Request context
    correlation_id UUID,
    referer_url TEXT,
    
    -- Performance tracking
    cache_hit BOOLEAN DEFAULT FALSE,
    storage_backend VARCHAR(50) DEFAULT 'azure-blob'
);

-- Content virus scan table
CREATE TABLE content_virus_scan (
    scan_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_id UUID NOT NULL REFERENCES content(content_id),
    scan_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    scanner_engine VARCHAR(50) NOT NULL,
    scanner_version VARCHAR(50) NOT NULL,
    scan_status VARCHAR(20) NOT NULL CHECK (scan_status IN ('clean', 'infected', 'suspicious', 'error')),
    threats_detected TEXT[],
    scan_duration_ms INTEGER,
    
    -- Audit fields
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    correlation_id UUID
);

-- Content storage backend table
CREATE TABLE content_storage_backend (
    backend_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    backend_name VARCHAR(50) NOT NULL UNIQUE,
    backend_type VARCHAR(20) NOT NULL CHECK (backend_type IN ('azure-blob', 'local-filesystem')),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    priority_order INTEGER NOT NULL DEFAULT 0,
    
    -- Configuration
    base_url VARCHAR(500),
    access_key_vault_reference VARCHAR(200), -- Azure Key Vault reference
    configuration_json JSONB, -- Backend-specific config
    
    -- Health tracking
    last_health_check TIMESTAMPTZ,
    health_status VARCHAR(20) DEFAULT 'unknown' CHECK (health_status IN ('healthy', 'degraded', 'unhealthy', 'unknown')),
    
    -- Audit fields
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    modified_on TIMESTAMPTZ,
    modified_by VARCHAR(255)
);

-- Performance indexes per CONTENT-SCHEMA.md
CREATE INDEX idx_content_hash ON content(content_hash) WHERE is_deleted = FALSE;
CREATE INDEX idx_content_mime_type ON content(mime_type) WHERE is_deleted = FALSE;
CREATE INDEX idx_content_category ON content(content_category) WHERE is_deleted = FALSE;
CREATE INDEX idx_content_access_level ON content(access_level) WHERE is_deleted = FALSE;
CREATE INDEX idx_content_upload_status ON content(upload_status) WHERE is_deleted = FALSE;
CREATE INDEX idx_content_storage_path ON content(storage_path) WHERE is_deleted = FALSE;
CREATE INDEX idx_content_upload_correlation ON content(upload_correlation_id);
CREATE INDEX idx_content_created_on ON content(created_on) WHERE is_deleted = FALSE;
CREATE INDEX idx_content_file_size ON content(file_size) WHERE is_deleted = FALSE;

CREATE INDEX idx_access_log_content_id ON content_access_log(content_id);
CREATE INDEX idx_access_log_timestamp ON content_access_log(access_timestamp);
CREATE INDEX idx_access_log_user_id ON content_access_log(user_id);
CREATE INDEX idx_access_log_client_ip ON content_access_log(client_ip);
CREATE INDEX idx_access_log_access_type ON content_access_log(access_type);
CREATE INDEX idx_access_log_correlation ON content_access_log(correlation_id);
CREATE INDEX idx_access_log_cache_performance ON content_access_log(cache_hit, response_time_ms);

CREATE INDEX idx_virus_scan_content_id ON content_virus_scan(content_id);
CREATE INDEX idx_virus_scan_timestamp ON content_virus_scan(scan_timestamp);
CREATE INDEX idx_virus_scan_status ON content_virus_scan(scan_status);
CREATE INDEX idx_virus_scan_correlation ON content_virus_scan(correlation_id);

CREATE INDEX idx_storage_backend_type ON content_storage_backend(backend_type);
CREATE INDEX idx_storage_backend_active ON content_storage_backend(is_active);
CREATE INDEX idx_storage_backend_priority ON content_storage_backend(priority_order);
CREATE INDEX idx_storage_backend_health ON content_storage_backend(health_status);