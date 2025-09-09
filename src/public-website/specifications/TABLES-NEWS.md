# News Domain Database Schema

## Core Tables

### News Table
```sql
CREATE TABLE news (
    news_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    summary TEXT NOT NULL,
    content TEXT, -- News article content stored in PostgreSQL
    slug VARCHAR(255) UNIQUE NOT NULL,
    category_id UUID NOT NULL REFERENCES news_categories(category_id),
    image_url VARCHAR(500), -- URL to Azure Blob Storage for images only
    author_name VARCHAR(255),
    publication_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    external_source VARCHAR(255),
    external_url VARCHAR(500),
    publishing_status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (publishing_status IN ('draft', 'published', 'archived')),
    
    -- Content metadata
    tags TEXT[],
    news_type VARCHAR(50) NOT NULL CHECK (news_type IN ('announcement', 'press_release', 'event', 'update', 'alert', 'feature')),
    priority_level VARCHAR(20) NOT NULL DEFAULT 'normal' CHECK (priority_level IN ('low', 'normal', 'high', 'urgent')),
    
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

### News Categories Table
```sql
CREATE TABLE news_categories (
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
        (SELECT COUNT(*) FROM news_categories WHERE is_default_unassigned = TRUE AND is_deleted = FALSE) <= 1
    )
);
```

### Featured News Table
```sql
CREATE TABLE featured_news (
    featured_news_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    news_id UUID NOT NULL REFERENCES news(news_id),
    
    -- Audit fields
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    modified_on TIMESTAMPTZ,
    modified_by VARCHAR(255),
    
    CONSTRAINT only_one_featured_news CHECK (
        (SELECT COUNT(*) FROM featured_news) <= 1
    ),
    CONSTRAINT no_default_unassigned_featured CHECK (
        NOT EXISTS (
            SELECT 1 FROM news n
            JOIN news_categories nc ON n.category_id = nc.category_id
            WHERE n.news_id = featured_news.news_id 
            AND nc.is_default_unassigned = TRUE
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
  "entity_type": "news|news_category|featured_news",
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
- `job`: "news-audit"
- `environment`: "production|development" 
- `entity_type`: "news|news_category|featured_news"
- `operation_type`: "insert|update|delete"
- `user_id`: authenticated user identifier

## Performance Indexes

### Core Table Indexes
```sql
-- News table indexes
CREATE INDEX idx_news_category_id ON news(category_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_publishing_status ON news(publishing_status) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_slug ON news(slug) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_news_type ON news(news_type) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_priority_level ON news(priority_level) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_publication_timestamp ON news(publication_timestamp) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_author_name ON news(author_name) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_external_source ON news(external_source) WHERE is_deleted = FALSE;

-- News categories table indexes  
CREATE INDEX idx_news_categories_slug ON news_categories(slug) WHERE is_deleted = FALSE;
CREATE INDEX idx_news_categories_default ON news_categories(is_default_unassigned) WHERE is_deleted = FALSE;

-- Featured news table indexes
CREATE INDEX idx_featured_news_news_id ON featured_news(news_id);
```

### Audit Query Optimization via Grafana Cloud Loki
Audit data queries are optimized through Grafana Cloud Loki label indexing and LogQL queries. No PostgreSQL indexes required for audit data as all audit information is stored externally for compliance.

**Grafana Cloud Loki Query Patterns:**
```logql
-- Query all news operations
{job="news-audit", entity_type="news"} |= `operation_type`

-- Query specific entity operations  
{job="news-audit", entity_type="news"} |= `"entity_id":"<news-id>"`

-- Query operations by user
{job="news-audit", user_id="<user-id>"} |= `operation_type`

-- Query operations within time range with correlation tracking
{job="news-audit"} |= `"correlation_id":"<uuid>"` | json | line_format "{{.audit_timestamp}} {{.operation_type}} {{.entity_type}}"
```

## Database Functions and Triggers

### Grafana Cloud Loki Audit Trigger Function
```sql
CREATE OR REPLACE FUNCTION publish_news_audit_event_to_grafana_loki()
RETURNS TRIGGER AS $$
BEGIN
    -- News audit event publishing directly to Grafana Cloud Loki via Dapr pub/sub
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
CREATE OR REPLACE FUNCTION reassign_news_to_default_category()
RETURNS TRIGGER AS $$
BEGIN
    -- News category reassignment with Dapr event notification
    -- Implementation will publish to 'news-category-events' topic
    -- Ensures audit compliance via event sourcing pattern
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;
```

### Featured News Management Function
```sql
CREATE OR REPLACE FUNCTION manage_featured_news()
RETURNS TRIGGER AS $$
BEGIN
    -- Featured news management with single article constraint
    -- Implementation publishes to 'news-featured-events' topic
    -- Ensures audit compliance and business rule enforcement
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;
```

## Business Rules

### Category Management Rules
1. **Default Unassigned Category**: Exactly one category must have `is_default_unassigned = TRUE`
2. **Featured News**: Exactly one news article can be featured at any time
3. **Featured News Restriction**: The default unassigned category cannot contain featured news
4. **News Assignment**: All news articles must have a valid category_id
5. **Category Deletion**: When a category is soft-deleted, all its news articles are reassigned to the default unassigned category

### News Publishing Rules
1. **Publishing Status**: News can be 'draft', 'published', or 'archived'
2. **Default Status**: New news starts with 'draft' status
3. **Category Assignment**: All news starts with the default unassigned category
4. **News Types**: News must specify type (announcement, press_release, event, update, alert, feature)
5. **Content Storage**: News content stored in PostgreSQL TEXT fields for full-text search capability
6. **Image Storage**: News images stored in Azure Blob Storage, referenced via `image_url`
7. **Content Publishing**: News can be published without content (content can be NULL)
8. **URL Validation**: All `image_url` and `external_url` values must be valid HTTPS URLs when not NULL
9. **Priority Levels**: News priority affects display ordering and notification urgency

### News Content Rules
1. **Summary Required**: All news must have a summary for indexing and preview
2. **Author Attribution**: Author name optional for news articles (can be institutional)
3. **Publication Timestamp**: Used for chronological ordering and recency filtering
4. **Tags**: PostgreSQL array for news tags and content categorization
5. **External Sources**: Optional external source and URL for attribution
6. **Content Immutability**: Published news content changes require audit approval
7. **Version Control**: Content changes create new audit events for compliance tracking
8. **Urgency Management**: High and urgent priority news require immediate publication workflow

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
3. **Timestamp Ordering**: Publication timestamps used for chronological ordering and filtering
4. **Foreign Key Integrity**: All category_id references must be valid and active
5. **Image URL Integrity**: `image_url` must point to existing blob storage content when not NULL
6. **Content Validation**: News content must be valid text format when not NULL
7. **Publication Timestamp Validation**: Publication timestamps must be realistic (not future dates beyond reasonable scheduling)
8. **Featured News Integrity**: Only one news article can be featured, and it cannot be from default unassigned category
9. **Priority Level Enforcement**: Priority levels affect notification and display systems consistently