# Services Domain Database Schema

## Core Tables

### Services Table
```sql
CREATE TABLE services (
    service_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    content_url VARCHAR(500), -- URL to Azure Blob Storage content
    category_id UUID NOT NULL REFERENCES service_categories(category_id),
    image_url VARCHAR(500),
    order_number INTEGER NOT NULL DEFAULT 0,
    delivery_mode VARCHAR(50) NOT NULL CHECK (delivery_mode IN ('mobile_service', 'outpatient_service', 'inpatient_service')),
    publishing_status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (publishing_status IN ('draft', 'published', 'archived')),
    
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

### Service Categories Table
```sql
CREATE TABLE service_categories (
    category_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    order_number INTEGER NOT NULL DEFAULT 0,
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
        (SELECT COUNT(*) FROM service_categories WHERE is_default_unassigned = TRUE AND is_deleted = FALSE) <= 1
    )
);
```

### Featured Categories Table
```sql
CREATE TABLE featured_categories (
    featured_category_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID NOT NULL REFERENCES service_categories(category_id),
    feature_position INTEGER NOT NULL CHECK (feature_position IN (1, 2)),
    
    -- Audit fields
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    modified_on TIMESTAMPTZ,
    modified_by VARCHAR(255),
    
    UNIQUE(feature_position),
    CONSTRAINT no_default_unassigned_featured CHECK (
        NOT EXISTS (
            SELECT 1 FROM service_categories sc 
            WHERE sc.category_id = featured_categories.category_id 
            AND sc.is_default_unassigned = TRUE
        )
    )
);
```

## Audit Event Publishing

### Audit Events via Grafana Cloud Loki
All audit data is published directly to Grafana Cloud Loki for compliance and immutable storage. No audit tables exist in PostgreSQL to prevent local tampering and ensure regulatory compliance.

**Audit Event Structure:**
```json
{
  "audit_id": "uuid",
  "entity_type": "service|category|featured_category",
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
- `job`: "services-audit"
- `environment`: "production|development" 
- `entity_type`: "service|category|featured_category"
- `operation_type`: "insert|update|delete"
- `user_id`: authenticated user identifier

## Performance Indexes

### Core Table Indexes
```sql
-- Services table indexes
CREATE INDEX idx_services_category_id ON services(category_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_services_publishing_status ON services(publishing_status) WHERE is_deleted = FALSE;
CREATE INDEX idx_services_slug ON services(slug) WHERE is_deleted = FALSE;
CREATE INDEX idx_services_order_category ON services(category_id, order_number) WHERE is_deleted = FALSE;
CREATE INDEX idx_services_delivery_mode ON services(delivery_mode) WHERE is_deleted = FALSE;

-- Service categories table indexes  
CREATE INDEX idx_service_categories_slug ON service_categories(slug) WHERE is_deleted = FALSE;
CREATE INDEX idx_service_categories_order ON service_categories(order_number) WHERE is_deleted = FALSE;
CREATE INDEX idx_service_categories_default ON service_categories(is_default_unassigned) WHERE is_deleted = FALSE;

-- Featured categories table indexes
CREATE INDEX idx_featured_categories_category_id ON featured_categories(category_id);
CREATE INDEX idx_featured_categories_position ON featured_categories(feature_position);
```

### Audit Query Optimization via Grafana Cloud Loki
Audit data queries are optimized through Grafana Cloud Loki label indexing and LogQL queries. No PostgreSQL indexes required for audit data as all audit information is stored externally for compliance.

**Grafana Cloud Loki Query Patterns:**
```logql
-- Query all service operations
{job="services-audit", entity_type="service"} |= `operation_type`

-- Query specific entity operations  
{job="services-audit", entity_type="service"} |= `"entity_id":"<service-id>"`

-- Query operations by user
{job="services-audit", user_id="<user-id>"} |= `operation_type`

-- Query operations within time range with correlation tracking
{job="services-audit"} |= `"correlation_id":"<uuid>"` | json | line_format "{{.audit_timestamp}} {{.operation_type}} {{.entity_type}}"
```

## Content Storage Strategy

### Rich Content Management
Services rich content (detailed descriptions, articles) is stored in Azure Blob Storage via Dapr bindings for cloud-native portability.

**Storage Pattern:**
- **Database**: Stores `content_url` pointing to blob storage content
- **Azure Blob Storage**: Stores HTML/Markdown files accessed via Dapr binding `blob-storage`
- **Local Development**: Azurite emulator via Dapr binding `blob-storage-local`

**URL Structure:**
```
blob-storage://{environment}/services/content/{service-id}/{content-hash}.html
```

**Content Versioning:**
- Content files use hash-based naming for immutability
- URL changes trigger audit events via Dapr pub/sub to `services-content-events` topic
- Previous content versions retained for audit compliance

**Performance Benefits:**
- Smaller database rows with PostgreSQL optimization
- Dapr binding abstraction enables multi-cloud portability
- Blob storage optimized for large content delivery
- Cost-effective storage with Azure managed services

**Compliance:**
- Content URL changes audited in `services_audit` table
- Content integrity verified via hash-based storage
- Immutable content versions for regulatory requirements
- Audit trail published to Grafana via Dapr telemetry middleware

## Database Functions and Triggers

### Grafana Cloud Loki Audit Trigger Function
```sql
CREATE OR REPLACE FUNCTION publish_audit_event_to_grafana_loki()
RETURNS TRIGGER AS $$
BEGIN
    -- Audit event publishing directly to Grafana Cloud Loki via Dapr pub/sub
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
CREATE OR REPLACE FUNCTION reassign_services_to_default_category()
RETURNS TRIGGER AS $$
BEGIN
    -- Category reassignment with Dapr event notification
    -- Implementation will publish to 'services-category-events' topic
    -- Ensures audit compliance via event sourcing pattern
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;
```

## Business Rules

### Category Management Rules
1. **Default Unassigned Category**: Exactly one category must have `is_default_unassigned = TRUE`
2. **Featured Categories**: Exactly two categories can be featured (positions 1 and 2)
3. **Featured Category Restriction**: The default unassigned category cannot be featured
4. **Service Assignment**: All services must have a valid category_id
5. **Category Deletion**: When a category is soft-deleted, all its services are reassigned to the default unassigned category

### Service Publishing Rules
1. **Publishing Status**: Services can be 'draft', 'published', or 'archived'
2. **Default Status**: New services start with 'draft' status
3. **Category Assignment**: All services start with the default unassigned category
4. **Delivery Modes**: Services must specify delivery mode (mobile_service, outpatient_service, inpatient_service)
5. **Content Storage**: Rich content stored in Azure Blob Storage, referenced via `long_description_url`
6. **Content Versioning**: Content URLs use hash-based naming for immutable versions
7. **Content Publishing**: Services can be published without rich content (URL can be NULL)
8. **URL Validation**: All `content_url` values must be valid HTTPS URLs when not NULL

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
2. **Unique Constraints**: Slugs must be unique across active (non-deleted) records
3. **Order Numbers**: Used for consistent ordering within categories and featured positions
4. **Foreign Key Integrity**: All category_id references must be valid and active
5. **Content URL Integrity**: `content_url` must point to existing blob storage content when not NULL
6. **Content Immutability**: Blob storage content files are never modified, only new versions created
7. **URL Format Validation**: Content URLs must follow the standard pattern with service-id and content-hash
8. **Content Lifecycle**: Orphaned content files cleaned up after audit retention period expires
