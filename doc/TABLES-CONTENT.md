# Content Domain Database Schema

## Core Tables

### Content Table
```sql
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
```

### Content Access Log Table
```sql
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
```

### Content Virus Scan Table
```sql
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
```

### Content Storage Backend Table
```sql
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
```

## Audit Event Publishing

### Audit Events via Grafana Cloud Loki
All audit data is published directly to Grafana Cloud Loki for compliance and immutable storage. No audit tables exist in PostgreSQL to prevent local tampering and ensure regulatory compliance.

**Audit Event Structure:**
```json
{
  "audit_id": "uuid",
  "entity_type": "content|storage_backend",
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
- `job`: "content-audit"
- `environment`: "production|development" 
- `entity_type`: "content|storage_backend"
- `operation_type`: "insert|update|delete"
- `user_id`: authenticated user identifier

## Performance Indexes

### Core Table Indexes
```sql
-- Content table indexes
CREATE INDEX idx_content_hash ON content(content_hash) WHERE is_deleted = FALSE;
CREATE INDEX idx_content_mime_type ON content(mime_type) WHERE is_deleted = FALSE;
CREATE INDEX idx_content_category ON content(content_category) WHERE is_deleted = FALSE;
CREATE INDEX idx_content_access_level ON content(access_level) WHERE is_deleted = FALSE;
CREATE INDEX idx_content_upload_status ON content(upload_status) WHERE is_deleted = FALSE;
CREATE INDEX idx_content_storage_path ON content(storage_path) WHERE is_deleted = FALSE;
CREATE INDEX idx_content_upload_correlation ON content(upload_correlation_id);
CREATE INDEX idx_content_created_on ON content(created_on) WHERE is_deleted = FALSE;
CREATE INDEX idx_content_file_size ON content(file_size) WHERE is_deleted = FALSE;

-- Content access log indexes
CREATE INDEX idx_access_log_content_id ON content_access_log(content_id);
CREATE INDEX idx_access_log_timestamp ON content_access_log(access_timestamp);
CREATE INDEX idx_access_log_user_id ON content_access_log(user_id);
CREATE INDEX idx_access_log_client_ip ON content_access_log(client_ip);
CREATE INDEX idx_access_log_access_type ON content_access_log(access_type);
CREATE INDEX idx_access_log_correlation ON content_access_log(correlation_id);
CREATE INDEX idx_access_log_cache_performance ON content_access_log(cache_hit, response_time_ms);

-- Content virus scan indexes
CREATE INDEX idx_virus_scan_content_id ON content_virus_scan(content_id);
CREATE INDEX idx_virus_scan_timestamp ON content_virus_scan(scan_timestamp);
CREATE INDEX idx_virus_scan_status ON content_virus_scan(scan_status);
CREATE INDEX idx_virus_scan_correlation ON content_virus_scan(correlation_id);

-- Content storage backend indexes
CREATE INDEX idx_storage_backend_type ON content_storage_backend(backend_type);
CREATE INDEX idx_storage_backend_active ON content_storage_backend(is_active);
CREATE INDEX idx_storage_backend_priority ON content_storage_backend(priority_order);
CREATE INDEX idx_storage_backend_health ON content_storage_backend(health_status);
```

### Audit Query Optimization via Grafana Cloud Loki
Audit data queries are optimized through Grafana Cloud Loki label indexing and LogQL queries. No PostgreSQL indexes required for audit data as all audit information is stored externally for compliance.

**Grafana Cloud Loki Query Patterns:**
```logql
-- Query all content operations
{job="content-audit", entity_type="content"} |= `operation_type`

-- Query specific content operations  
{job="content-audit", entity_type="content"} |= `"entity_id":"<content-id>"`

-- Query storage backend operations
{job="content-audit", entity_type="storage_backend"} |= `operation_type`

-- Query operations by user with correlation tracking
{job="content-audit"} |= `"user_id":"<user-id>"` | json | line_format "{{.audit_timestamp}} {{.operation_type}} {{.entity_type}}"
```

## Storage Architecture

### Azure Blob Storage Integration
Content files are stored in Azure Blob Storage accessed via Dapr bindings for cloud-native performance and portability.

**Storage Pattern:**
- **Database**: Stores metadata, hash, and Dapr binding reference
- **Azure Blob Storage**: Production storage via Dapr binding `blob-storage`  
- **Azurite Emulator**: Local development via Dapr binding `blob-storage-local`
- **Dapr Bindings**: Abstracts storage access for multi-environment portability

**Path Structure:**
```
{dapr-binding}://{environment}/content/{year}/{month}/{content-id}/{content-hash}.{ext}
```

**Content Processing Pipeline:**
1. File upload via multipart form to Dapr HTTP endpoint
2. Async virus scanning via Dapr pub/sub to `content-virus-scan` topic
3. Hash calculation for integrity verification
4. Metadata extraction and PostgreSQL storage
5. Azure Blob Storage via Dapr binding
6. Database record creation with correlation tracking
7. Audit event generation via Dapr pub/sub to `content-audit-events` topic
8. Grafana telemetry collection throughout pipeline

**Content Integrity:**
- SHA-256 hash stored for integrity verification
- Content immutable after successful upload
- Hash-based deduplication prevents duplicate storage
- Virus scanning via Azure Service Bus integration

### Content Delivery Strategy

**Access Patterns:**
- **Public Content**: Served via Dapr HTTP middleware with access control
- **Internal Content**: Authentication via Authentik OAuth2 middleware
- **Restricted Content**: Role-based access control via OPA policies

**Performance Optimization:**
- Content access logging via Dapr pub/sub to `content-access-events` topic
- Grafana telemetry for response time and cache performance monitoring
- PostgreSQL query optimization for metadata access
- Azure Blob Storage integration for large file delivery

**Security Measures:**
- Virus scanning via Dapr pub/sub integration with Azure Service Bus
- Access level enforcement via Dapr middleware chain
- IP-based rate limiting via Dapr rate limit middleware
- User agent analysis via Grafana telemetry collection

## Database Functions and Triggers

### Content Processing Trigger Function
```sql
CREATE OR REPLACE FUNCTION process_content_upload_with_dapr_events()
RETURNS TRIGGER AS $$
BEGIN
    -- Content upload processing with Dapr pub/sub integration
    -- Publishes to 'content-upload-events' topic for async virus scanning
    -- Publishes to 'content-processing-events' topic for metadata extraction
    -- Integration with Grafana telemetry for correlation tracking
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

### Grafana Cloud Loki Audit Function
```sql
CREATE OR REPLACE FUNCTION publish_content_audit_event_to_grafana_loki()
RETURNS TRIGGER AS $$
BEGIN
    -- Content audit event publishing directly to Grafana Cloud Loki via Dapr pub/sub
    -- Implementation publishes to 'grafana-audit-events' topic via Dapr
    -- Event structure includes complete data snapshots for compliance
    -- No local audit table storage - all audit data stored in Grafana Cloud Loki
    -- Ensures immutable audit trail and prevents local tampering
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;
```

### Storage Health Check Function
```sql
CREATE OR REPLACE FUNCTION update_storage_health_with_notifications()
RETURNS TRIGGER AS $$
BEGIN
    -- Storage health monitoring with Dapr service invocation
    -- Publishes to 'storage-health-events' topic for alerting
    -- Publishes audit events to 'grafana-audit-events' for storage operations
    -- Integration with Grafana monitoring for dashboard updates
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

## Business Rules

### Content Upload Rules
1. **File Size Limits**: Maximum file size enforced at application level
2. **Mime Type Validation**: Only approved mime types accepted
3. **Virus Scanning**: All uploads scanned before storage (if enabled)
4. **Hash Integrity**: SHA-256 hash calculated and verified
5. **Duplicate Prevention**: Hash-based deduplication prevents storage waste
6. **Processing Status**: Content marked as 'processing' until fully uploaded
7. **Correlation Tracking**: All upload operations share correlation_id

### Content Access Rules  
1. **Access Level Enforcement**: Public, internal, restricted access levels enforced
2. **Authentication Required**: Internal and restricted content requires authentication
3. **Role-Based Access**: Restricted content requires specific roles
4. **Access Logging**: All content access logged for audit compliance
5. **Rate Limiting**: IP-based rate limiting for anonymous access
6. **Cache Headers**: Appropriate cache headers for performance optimization

### Content Security Rules
1. **Virus Scanning Integration**: Optional virus scanning for all uploads
2. **Content Classification**: Automatic classification based on mime type
3. **Access Monitoring**: Suspicious access patterns logged and alerted
4. **IP Tracking**: Client IP addresses logged for security analysis
5. **User Agent Analysis**: User agent strings analyzed for bot detection

### Storage Management Rules
1. **Backend Failover**: Multiple storage backends with priority ordering
2. **Health Monitoring**: Regular health checks of storage backends
3. **Configuration Security**: Storage credentials stored in Azure Key Vault
4. **Path Immutability**: Storage paths never changed after creation
5. **Retention Policy**: Content retention based on access level and compliance requirements

### Audit Requirements
1. **Compliance Logging**: All content operations audited via direct publishing to Grafana Cloud Loki
2. **Immutable Audit Records**: Audit events stored in Grafana Cloud Loki, never modified or deleted
3. **Complete Snapshots**: Audit events include complete before/after data snapshots for compliance
4. **Access Tracking**: Every content access logged with full context to Grafana Cloud Loki
5. **Correlation IDs**: All operations share correlation_id and trace_id for distributed tracing
6. **User Attribution**: All operations record user_id from Authentik authentication context
7. **Grafana Cloud Integration**: All audit operations stored directly in Grafana Cloud Loki for compliance
8. **No Local Audit Storage**: No audit tables in PostgreSQL to prevent local tampering

### Data Integrity Rules
1. **Soft Delete Only**: Content records never physically deleted
2. **Hash Uniqueness**: Content hash ensures no duplicate storage
3. **Foreign Key Integrity**: All content_id references must be valid
4. **Storage Path Validation**: All storage paths must follow standard format
5. **File Size Accuracy**: File size must match actual stored content
6. **Mime Type Validation**: Mime type must match file content analysis
7. **Processing State**: Content status accurately reflects processing state
8. **Access Level Consistency**: Access level enforced across all access points
9. **Azure Integration**: Content stored in Azure Blob Storage with Azurite emulator for testing