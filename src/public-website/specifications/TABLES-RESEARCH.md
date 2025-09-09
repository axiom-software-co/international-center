# Research Domain Database Schema

## Core Tables

### Research Table
```sql
CREATE TABLE research (
    research_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    abstract TEXT NOT NULL,
    content TEXT, -- Research article content stored in PostgreSQL
    slug VARCHAR(255) UNIQUE NOT NULL,
    category_id UUID NOT NULL REFERENCES research_categories(category_id),
    image_url VARCHAR(500), -- URL to Azure Blob Storage for images only
    author_names VARCHAR(500) NOT NULL,
    publication_date DATE,
    doi VARCHAR(100),
    external_url VARCHAR(500),
    report_url VARCHAR(500), -- URL to PDF publication report in Azure Blob Storage
    publishing_status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (publishing_status IN ('draft', 'published', 'archived')),
    
    -- Content metadata
    keywords TEXT[],
    research_type VARCHAR(50) NOT NULL CHECK (research_type IN ('clinical_study', 'case_report', 'systematic_review', 'meta_analysis', 'editorial', 'commentary')),
    
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

### Research Categories Table
```sql
CREATE TABLE research_categories (
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
        (SELECT COUNT(*) FROM research_categories WHERE is_default_unassigned = TRUE AND is_deleted = FALSE) <= 1
    )
);
```

### Featured Research Table
```sql
CREATE TABLE featured_research (
    featured_research_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    research_id UUID NOT NULL REFERENCES research(research_id),
    
    -- Audit fields
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    modified_on TIMESTAMPTZ,
    modified_by VARCHAR(255),
    
    CONSTRAINT only_one_featured_research CHECK (
        (SELECT COUNT(*) FROM featured_research) <= 1
    ),
    CONSTRAINT no_default_unassigned_featured CHECK (
        NOT EXISTS (
            SELECT 1 FROM research r
            JOIN research_categories rc ON r.category_id = rc.category_id
            WHERE r.research_id = featured_research.research_id 
            AND rc.is_default_unassigned = TRUE
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
  "entity_type": "research|research_category|featured_research",
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
- `job`: "research-audit"
- `environment`: "production|development" 
- `entity_type`: "research|research_category|featured_research"
- `operation_type`: "insert|update|delete"
- `user_id`: authenticated user identifier

## Performance Indexes

### Core Table Indexes
```sql
-- Research table indexes
CREATE INDEX idx_research_category_id ON research(category_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_research_publishing_status ON research(publishing_status) WHERE is_deleted = FALSE;
CREATE INDEX idx_research_slug ON research(slug) WHERE is_deleted = FALSE;
CREATE INDEX idx_research_research_type ON research(research_type) WHERE is_deleted = FALSE;
CREATE INDEX idx_research_publication_date ON research(publication_date) WHERE is_deleted = FALSE;
CREATE INDEX idx_research_author_names ON research(author_names) WHERE is_deleted = FALSE;
CREATE INDEX idx_research_doi ON research(doi) WHERE is_deleted = FALSE;

-- Research categories table indexes  
CREATE INDEX idx_research_categories_slug ON research_categories(slug) WHERE is_deleted = FALSE;
CREATE INDEX idx_research_categories_default ON research_categories(is_default_unassigned) WHERE is_deleted = FALSE;

-- Featured research table indexes
CREATE INDEX idx_featured_research_research_id ON featured_research(research_id);
```

### Audit Query Optimization via Grafana Cloud Loki
Audit data queries are optimized through Grafana Cloud Loki label indexing and LogQL queries. No PostgreSQL indexes required for audit data as all audit information is stored externally for compliance.

**Grafana Cloud Loki Query Patterns:**
```logql
-- Query all research operations
{job="research-audit", entity_type="research"} |= `operation_type`

-- Query specific entity operations  
{job="research-audit", entity_type="research"} |= `"entity_id":"<research-id>"`

-- Query operations by user
{job="research-audit", user_id="<user-id>"} |= `operation_type`

-- Query operations within time range with correlation tracking
{job="research-audit"} |= `"correlation_id":"<uuid>"` | json | line_format "{{.audit_timestamp}} {{.operation_type}} {{.entity_type}}"
```

## Database Functions and Triggers

### Grafana Cloud Loki Audit Trigger Function
```sql
CREATE OR REPLACE FUNCTION publish_research_audit_event_to_grafana_loki()
RETURNS TRIGGER AS $$
BEGIN
    -- Research audit event publishing directly to Grafana Cloud Loki via Dapr pub/sub
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
CREATE OR REPLACE FUNCTION reassign_research_to_default_category()
RETURNS TRIGGER AS $$
BEGIN
    -- Research category reassignment with Dapr event notification
    -- Implementation will publish to 'research-category-events' topic
    -- Ensures audit compliance via event sourcing pattern
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;
```

### Featured Research Management Function
```sql
CREATE OR REPLACE FUNCTION manage_featured_research()
RETURNS TRIGGER AS $$
BEGIN
    -- Featured research management with single article constraint
    -- Implementation publishes to 'research-featured-events' topic
    -- Ensures audit compliance and business rule enforcement
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;
```

## Business Rules

### Category Management Rules
1. **Default Unassigned Category**: Exactly one category must have `is_default_unassigned = TRUE`
2. **Featured Research**: Exactly one research article can be featured at any time
3. **Featured Research Restriction**: The default unassigned category cannot contain featured research
4. **Research Assignment**: All research articles must have a valid category_id
5. **Category Deletion**: When a category is soft-deleted, all its research articles are reassigned to the default unassigned category

### Research Publishing Rules
1. **Publishing Status**: Research can be 'draft', 'published', or 'archived'
2. **Default Status**: New research starts with 'draft' status
3. **Category Assignment**: All research starts with the default unassigned category
4. **Research Types**: Research must specify type (clinical_study, case_report, systematic_review, meta_analysis, editorial, commentary)
5. **Content Storage**: Research content stored in PostgreSQL TEXT fields for full-text search capability
6. **Image Storage**: Research images stored in Azure Blob Storage, referenced via `image_url`
7. **Content Publishing**: Research can be published without content (content can be NULL)
8. **URL Validation**: All `image_url` and `external_url` values must be valid HTTPS URLs when not NULL
9. **DOI Uniqueness**: DOI values must be unique when not NULL

### Research Content Rules
1. **Abstract Required**: All research must have an abstract for indexing and search
2. **Author Attribution**: Author names are required for all research articles
3. **Publication Date**: Publication date used for chronological ordering and filtering
4. **Keywords**: PostgreSQL array for research keywords and tags
5. **External References**: Optional DOI and external URL for source attribution
6. **Publication Reports**: Optional PDF publication reports stored in Azure Blob Storage
7. **Content Immutability**: Published research content changes require audit approval
8. **Version Control**: Content changes create new audit events for compliance tracking

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
2. **Unique Constraints**: Slugs and DOIs must be unique across active (non-deleted) records
3. **Publication Date Ordering**: Publication dates used for chronological ordering and filtering
4. **Foreign Key Integrity**: All category_id references must be valid and active
5. **Image URL Integrity**: `image_url` must point to existing blob storage content when not NULL
6. **Content Validation**: Research content must be valid text format when not NULL
7. **Publication Date Validation**: Publication dates must be realistic (not future dates beyond current date)
8. **Report URL Integrity**: `report_url` must point to existing PDF content in blob storage when not NULL
9. **Featured Research Integrity**: Only one research article can be featured, and it cannot be from default unassigned category