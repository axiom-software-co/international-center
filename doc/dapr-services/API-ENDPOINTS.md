# API Endpoints Specification

## Architecture Overview

**Service Mesh**: Dapr-centric microservices architecture
**Gateway Pattern**: Public and admin API gateways with middleware pipelines
**Authentication**: Anonymous (public) / OAuth2+JWT (admin) via Authentik
**Observability**: Grafana Cloud telemetry and audit logging integration
**Environment Configuration**: Container-only environment variables

## Services Domain API

### Public Services API (Public Gateway)

**Base Path**: `/api/v1/services`  
**Gateway**: Public Gateway (Dapr Service: `public-gateway`)  
**Service Invocation**: `public-gateway/api/v1/services`  
**Authentication**: Anonymous (Authentik anonymous access)  
**Rate Limiting**: 1000 requests/minute per IP (Dapr rate limit middleware)  
**Backing Store**: Azure CosmosDB (Dapr state store: `cosmosdb-services`)  
**CORS**: Public website origins (Dapr CORS middleware)  

#### Service Retrieval Endpoints

##### GET /api/v1/services
**Description**: Retrieve paginated list of published services  
**Access Level**: Anonymous  
**Query Parameters**:
- `page` (int, optional): Page number, default 1
- `limit` (int, optional): Items per page, default 20, max 100
- `category_id` (UUID, optional): Filter by category
- `delivery_mode` (string, optional): Filter by delivery mode

**Response**: 200 OK
```json
{
  "services": [
    {
      "service_id": "uuid",
      "title": "string",
      "description": "string",
      "slug": "string",
      "long_description_url": "string|null",
      "category_id": "uuid",
      "category_name": "string",
      "image_url": "string|null",
      "delivery_mode": "mobile_service|outpatient_service|inpatient_service",
      "created_on": "timestamp"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total_items": 150,
    "total_pages": 8,
    "has_next": true,
    "has_previous": false
  }
}
```

##### GET /api/v1/services/{id}
**Description**: Retrieve service by ID (published services only)  
**Access Level**: Anonymous  
**Path Parameters**: `id` (UUID): Service identifier

**Response**: 200 OK
```json
{
  "service_id": "uuid",
  "title": "string",
  "description": "string",
  "slug": "string",
  "long_description_url": "string|null",
  "category_id": "uuid",
  "category_name": "string",
  "category_slug": "string",
  "image_url": "string|null",
  "delivery_mode": "mobile_service|outpatient_service|inpatient_service",
  "publishing_status": "published",
  "created_on": "timestamp"
}
```

##### GET /api/v1/services/slug/{slug}
**Description**: Retrieve service by URL-friendly slug  
**Access Level**: Anonymous  
**Path Parameters**: `slug` (string): URL-friendly service identifier

**Response**: 200 OK (same as GET /api/v1/services/{id})

##### GET /api/v1/services/featured
**Description**: Retrieve services from featured categories (positions 1 and 2)  
**Access Level**: Anonymous

**Response**: 200 OK
```json
{
  "featured_categories": [
    {
      "category_id": "uuid",
      "name": "string",
      "slug": "string",
      "feature_position": 1,
      "services": [
        {
          "service_id": "uuid",
          "title": "string",
          "description": "string",
          "slug": "string",
          "image_url": "string|null",
          "delivery_mode": "mobile_service|outpatient_service|inpatient_service"
        }
      ]
    }
  ]
}
```

#### Category Endpoints

##### GET /api/v1/services/categories
**Description**: Retrieve list of active service categories (excluding default unassigned)  
**Access Level**: Anonymous

**Response**: 200 OK
```json
{
  "categories": [
    {
      "category_id": "uuid",
      "name": "string",
      "slug": "string",
      "order_number": 1,
      "service_count": 15
    }
  ]
}
```

##### GET /api/v1/services/categories/{id}/services
**Description**: Retrieve published services by category  
**Access Level**: Anonymous  
**Path Parameters**: `id` (UUID): Category identifier  
**Query Parameters**: Standard pagination parameters

**Response**: 200 OK (same structure as GET /api/v1/services)

#### Search Endpoints

##### GET /api/v1/services/search
**Description**: Search published services with full-text search and filters  
**Access Level**: Anonymous  
**Query Parameters**:
- `q` (string, required): Search query
- `category_id` (UUID, optional): Filter by category
- `delivery_mode` (string, optional): Filter by delivery mode
- `page` (int, optional): Page number, default 1
- `limit` (int, optional): Items per page, default 20, max 50

**Response**: 200 OK
```json
{
  "query": "search terms",
  "services": [...],
  "pagination": {...},
  "search_metadata": {
    "total_matches": 25,
    "search_time_ms": 45
  }
}
```

### Admin Services API (Admin Gateway)

**Base Path**: `/admin/api/v1/services`  
**Gateway**: Admin Gateway (Dapr Service: `admin-gateway`)  
**Service Invocation**: `admin-gateway/admin/api/v1/services`  
**Authentication**: OAuth2/JWT (Authentik OAuth2 middleware)  
**Authorization**: Role-based access control (OPA policies via Dapr)  
**Rate Limiting**: 100 requests/minute per user (Dapr rate limit middleware)  
**Audit Logging**: Event sourcing via Dapr pub/sub to audit topics  

#### Service Management Endpoints

##### GET /admin/api/v1/services
**Description**: Retrieve all services including drafts and archived  
**Access Level**: Admin role required  
**Query Parameters**:
- `page` (int, optional): Page number, default 1
- `limit` (int, optional): Items per page, default 20, max 100
- `status` (string, optional): Filter by publishing status
- `category_id` (UUID, optional): Filter by category
- `include_deleted` (bool, optional): Include soft-deleted services, default false

**Response**: 200 OK
```json
{
  "services": [
    {
      "service_id": "uuid",
      "title": "string",
      "description": "string",
      "slug": "string",
      "long_description_url": "string|null",
      "category_id": "uuid",
      "category_name": "string",
      "image_url": "string|null",
      "order_number": 1,
      "delivery_mode": "mobile_service|outpatient_service|inpatient_service",
      "publishing_status": "draft|published|archived",
      "created_on": "timestamp",
      "created_by": "string|null",
      "modified_on": "timestamp|null",
      "modified_by": "string|null",
      "is_deleted": false
    }
  ],
  "pagination": {...}
}
```

##### POST /admin/api/v1/services
**Description**: Create new service  
**Access Level**: Admin role required  
**Request Body**:
```json
{
  "title": "string",
  "description": "string",
  "slug": "string|null",
  "long_description_url": "string|null",
  "category_id": "uuid|null",
  "image_url": "string|null",
  "order_number": 0,
  "delivery_mode": "mobile_service|outpatient_service|inpatient_service"
}
```

**Response**: 201 Created
```json
{
  "service_id": "uuid",
  "title": "string",
  "description": "string",
  "slug": "string",
  "publishing_status": "draft",
  "category_id": "uuid",
  "created_on": "timestamp",
  "created_by": "string"
}
```

##### PUT /admin/api/v1/services/{id}
**Description**: Update existing service  
**Access Level**: Admin role required  
**Path Parameters**: `id` (UUID): Service identifier  
**Request Body**: (same as POST, all fields optional except those being updated)

**Response**: 200 OK (updated service object)

##### DELETE /admin/api/v1/services/{id}
**Description**: Soft delete service (sets is_deleted = true)  
**Access Level**: Admin role required  
**Path Parameters**: `id` (UUID): Service identifier

**Response**: 204 No Content

#### Publishing Workflow Endpoints

##### POST /admin/api/v1/services/{id}/publish
**Description**: Change service status to published  
**Access Level**: Admin role required  
**Path Parameters**: `id` (UUID): Service identifier

**Response**: 200 OK
```json
{
  "service_id": "uuid",
  "previous_status": "draft",
  "current_status": "published",
  "published_at": "timestamp",
  "published_by": "string"
}
```

##### POST /admin/api/v1/services/{id}/archive
**Description**: Change service status to archived  
**Access Level**: Admin role required  
**Path Parameters**: `id` (UUID): Service identifier

**Response**: 200 OK
```json
{
  "service_id": "uuid",
  "previous_status": "published",
  "current_status": "archived",
  "archived_at": "timestamp",
  "archived_by": "string"
}
```

#### Category Management Endpoints

##### POST /admin/api/v1/services/categories
**Description**: Create new service category  
**Access Level**: Admin role required  
**Request Body**:
```json
{
  "name": "string",
  "slug": "string|null",
  "order_number": 0
}
```

**Response**: 201 Created (category object)

##### PUT /admin/api/v1/services/categories/{id}
**Description**: Update service category  
**Access Level**: Admin role required  
**Path Parameters**: `id` (UUID): Category identifier

**Response**: 200 OK (updated category object)

##### DELETE /admin/api/v1/services/categories/{id}
**Description**: Soft delete category (reassigns services to default unassigned)  
**Access Level**: Admin role required  
**Path Parameters**: `id` (UUID): Category identifier

**Response**: 204 No Content

#### Featured Category Management

##### GET /admin/api/v1/services/featured-categories
**Description**: Retrieve current featured category assignments  
**Access Level**: Admin role required

**Response**: 200 OK
```json
{
  "featured_categories": [
    {
      "featured_category_id": "uuid",
      "category_id": "uuid",
      "category_name": "string",
      "feature_position": 1,
      "created_on": "timestamp",
      "created_by": "string"
    }
  ]
}
```

##### POST /admin/api/v1/services/featured-categories
**Description**: Set category as featured (positions 1 or 2 only)  
**Access Level**: Admin role required  
**Request Body**:
```json
{
  "category_id": "uuid",
  "feature_position": 1
}
```

**Response**: 201 Created

##### DELETE /admin/api/v1/services/featured-categories/{position}
**Description**: Remove category from featured position  
**Access Level**: Admin role required  
**Path Parameters**: `position` (int): Feature position (1 or 2)

**Response**: 204 No Content

## Content Domain API

### Public Content API (Public Gateway)

**Base Path**: `/api/v1/content`  
**Gateway**: Public Gateway (Dapr Service: `public-gateway`)  
**Service Invocation**: `public-gateway/api/v1/content`  
**Authentication**: Anonymous (public content) / Optional JWT (internal content) via Authentik  
**Rate Limiting**: 1000 requests/minute per IP (Dapr rate limit middleware)  
**Backing Store**: Azure CosmosDB (Dapr state store: `cosmosdb-content`)  
**Storage**: Azure Blob Storage (Dapr binding: `blob-storage`)  

#### Content Access Endpoints

##### GET /api/v1/content/{id}
**Description**: Retrieve content metadata and access information  
**Access Level**: Public/Internal content (restricted requires admin access)  
**Path Parameters**: `id` (UUID): Content identifier

**Response**: 200 OK
```json
{
  "content_id": "uuid",
  "original_filename": "string",
  "file_size": 1048576,
  "mime_type": "string",
  "content_hash": "string",
  "alt_text": "string|null",
  "description": "string|null",
  "tags": ["tag1", "tag2"],
  "content_category": "document|image|video|audio|archive",
  "access_level": "public|internal",
  "download_url": "string",
  "preview_url": "string|null",
  "created_on": "timestamp"
}
```

##### GET /api/v1/content/{id}/download
**Description**: Download content file with access logging  
**Access Level**: Based on content access_level  
**Path Parameters**: `id` (UUID): Content identifier  
**Query Parameters**: `inline` (bool, optional): Return as inline content, default false

**Response**: 200 OK
- **Headers**:
  - `Content-Type: <mime_type>`
  - `Content-Length: <file_size>`
  - `Content-Disposition: attachment; filename="<original_filename>"` (or inline)
  - `Content-Security-Policy: default-src 'none'`
  - `X-Content-Type-Options: nosniff`
  - `X-Content-Hash: <sha256_hash>`
- **Body**: Binary content data

##### GET /api/v1/content/{id}/preview
**Description**: Generate or retrieve content preview/thumbnail  
**Access Level**: Based on content access_level  
**Path Parameters**: `id` (UUID): Content identifier  
**Query Parameters**:
- `size` (string, optional): Preview size (small, medium, large), default medium
- `format` (string, optional): Preview format (webp, jpeg, png), default webp

**Response**: 200 OK (binary image data or redirect to preview URL)

#### Content Discovery Endpoints

##### GET /api/v1/content
**Description**: List public content with pagination and filters  
**Access Level**: Anonymous (public content only)  
**Query Parameters**:
- `page` (int, optional): Page number, default 1
- `limit` (int, optional): Items per page, default 20, max 100
- `category` (string, optional): Filter by content category
- `mime_type` (string, optional): Filter by MIME type pattern
- `tags` (string[], optional): Filter by tags (comma-separated)

**Response**: 200 OK
```json
{
  "content": [
    {
      "content_id": "uuid",
      "original_filename": "string",
      "file_size": 1048576,
      "mime_type": "string",
      "content_category": "document|image|video|audio|archive",
      "alt_text": "string|null",
      "tags": ["tag1", "tag2"],
      "download_url": "string",
      "preview_url": "string|null",
      "created_on": "timestamp"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total_items": 150,
    "total_pages": 8,
    "has_next": true,
    "has_previous": false
  }
}
```

### Admin Content API (Admin Gateway)

**Base Path**: `/admin/api/v1/content`  
**Gateway**: Admin Gateway (Dapr Service: `admin-gateway`)  
**Service Invocation**: `admin-gateway/admin/api/v1/content`  
**Authentication**: OAuth2/JWT (Authentik OAuth2 middleware)  
**Authorization**: Role-based access control (OPA policies via Dapr)  
**Rate Limiting**: 100 requests/minute per user (Dapr rate limit middleware)  
**Storage**: Azure Blob Storage (Dapr binding: `blob-storage`)  
**Virus Scanning**: Async processing via Dapr pub/sub to `content-virus-scan` topic  

#### Content Management Endpoints

##### GET /admin/api/v1/content
**Description**: Retrieve all content including restricted and processing items  
**Access Level**: Admin role required  
**Query Parameters**:
- `page` (int, optional): Page number, default 1
- `limit` (int, optional): Items per page, default 20, max 100
- `status` (string, optional): Filter by upload status
- `access_level` (string, optional): Filter by access level
- `category` (string, optional): Filter by content category
- `include_deleted` (bool, optional): Include soft-deleted content, default false

**Response**: 200 OK
```json
{
  "content": [
    {
      "content_id": "uuid",
      "original_filename": "string",
      "file_size": 1048576,
      "mime_type": "string",
      "content_hash": "string",
      "storage_path": "string",
      "upload_status": "processing|available|failed|archived",
      "alt_text": "string|null",
      "description": "string|null",
      "tags": ["tag1", "tag2"],
      "content_category": "document|image|video|audio|archive",
      "access_level": "public|internal|restricted",
      "upload_correlation_id": "uuid",
      "processing_attempts": 1,
      "last_processed_at": "timestamp|null",
      "created_on": "timestamp",
      "created_by": "string|null",
      "modified_on": "timestamp|null",
      "modified_by": "string|null",
      "is_deleted": false
    }
  ],
  "pagination": {...}
}
```

##### POST /admin/api/v1/content/upload
**Description**: Upload new content file with virus scanning  
**Access Level**: Admin role required  
**Content-Type**: multipart/form-data  
**Request Body**:
- `file` (file): Content file to upload
- `alt_text` (string, optional): Alternative text description
- `description` (string, optional): Content description
- `tags` (string[], optional): Content tags (JSON array)
- `access_level` (string, optional): Access level (public, internal, restricted), default internal
- `content_category` (string, optional): Auto-detected from MIME type if not provided

**Response**: 202 Accepted
```json
{
  "content_id": "uuid",
  "original_filename": "string",
  "file_size": 1048576,
  "mime_type": "string",
  "upload_status": "processing",
  "upload_correlation_id": "uuid",
  "estimated_processing_time": "30s",
  "created_on": "timestamp"
}
```

##### PUT /admin/api/v1/content/{id}
**Description**: Update content metadata (not the file itself)  
**Access Level**: Admin role required  
**Path Parameters**: `id` (UUID): Content identifier  
**Request Body**:
```json
{
  "alt_text": "string|null",
  "description": "string|null",
  "tags": ["tag1", "tag2"],
  "access_level": "public|internal|restricted"
}
```

**Response**: 200 OK (updated content object)

##### DELETE /admin/api/v1/content/{id}
**Description**: Soft delete content (sets is_deleted = true, preserves file)  
**Access Level**: Admin role required  
**Path Parameters**: `id` (UUID): Content identifier

**Response**: 204 No Content

#### Content Processing Endpoints

##### POST /admin/api/v1/content/{id}/reprocess
**Description**: Trigger content reprocessing (virus scan, metadata extraction)  
**Access Level**: Admin role required  
**Path Parameters**: `id` (UUID): Content identifier

**Response**: 202 Accepted
```json
{
  "content_id": "uuid",
  "processing_status": "queued",
  "correlation_id": "uuid",
  "estimated_processing_time": "30s"
}
```

##### GET /admin/api/v1/content/{id}/status
**Description**: Get current processing status and details  
**Access Level**: Admin role required  
**Path Parameters**: `id` (UUID): Content identifier

**Response**: 200 OK
```json
{
  "content_id": "uuid",
  "upload_status": "processing|available|failed|archived",
  "processing_attempts": 2,
  "last_processed_at": "timestamp|null",
  "processing_details": {
    "virus_scan_status": "clean|infected|suspicious|error",
    "metadata_extracted": true,
    "thumbnail_generated": true,
    "storage_verified": true
  },
  "error_message": "string|null"
}
```

##### GET /admin/api/v1/content/processing-queue
**Description**: View content processing queue status  
**Access Level**: Admin role required

**Response**: 200 OK
```json
{
  "queue_status": {
    "pending_items": 5,
    "processing_items": 2,
    "failed_items": 1,
    "average_processing_time": "45s"
  },
  "recent_activity": [
    {
      "content_id": "uuid",
      "filename": "string",
      "status": "processing|completed|failed",
      "started_at": "timestamp",
      "completed_at": "timestamp|null"
    }
  ]
}
```

#### Content Analytics Endpoints

##### GET /admin/api/v1/content/{id}/access-log
**Description**: Retrieve access logs for specific content  
**Access Level**: Admin role required  
**Path Parameters**: `id` (UUID): Content identifier  
**Query Parameters**:
- `page` (int, optional): Page number, default 1
- `limit` (int, optional): Items per page, default 50, max 200
- `start_date` (datetime, optional): Filter from date
- `end_date` (datetime, optional): Filter to date
- `access_type` (string, optional): Filter by access type

**Response**: 200 OK
```json
{
  "content_id": "uuid",
  "access_logs": [
    {
      "access_id": "uuid",
      "access_timestamp": "timestamp",
      "user_id": "string|null",
      "client_ip": "192.168.1.1",
      "user_agent": "string",
      "access_type": "view|download|preview",
      "http_status_code": 200,
      "bytes_served": 1048576,
      "response_time_ms": 150,
      "correlation_id": "uuid",
      "referer_url": "string|null",
      "cache_hit": true
    }
  ],
  "pagination": {...},
  "analytics": {
    "total_accesses": 150,
    "unique_visitors": 45,
    "total_bytes_served": 157286400,
    "average_response_time": 125.5,
    "cache_hit_rate": 0.85
  }
}
```

##### GET /admin/api/v1/content/analytics
**Description**: Content access analytics and reporting  
**Access Level**: Admin role required  
**Query Parameters**:
- `start_date` (datetime, optional): Analysis period start
- `end_date` (datetime, optional): Analysis period end
- `granularity` (string, optional): Data granularity (hour, day, week), default day

**Response**: 200 OK
```json
{
  "period": {
    "start_date": "timestamp",
    "end_date": "timestamp",
    "granularity": "day"
  },
  "summary": {
    "total_accesses": 5000,
    "unique_visitors": 1200,
    "total_bytes_served": 5368709120,
    "average_response_time": 145.2,
    "cache_hit_rate": 0.82
  },
  "top_content": [
    {
      "content_id": "uuid",
      "filename": "string",
      "access_count": 150,
      "unique_visitors": 45,
      "bytes_served": 157286400
    }
  ],
  "access_patterns": [
    {
      "timestamp": "timestamp",
      "access_count": 245,
      "bytes_served": 268435456,
      "average_response_time": 132.1
    }
  ]
}
```

#### Virus Scanning Endpoints

##### GET /admin/api/v1/content/{id}/virus-scans
**Description**: Retrieve virus scan history for content  
**Access Level**: Admin role required  
**Path Parameters**: `id` (UUID): Content identifier

**Response**: 200 OK
```json
{
  "content_id": "uuid",
  "virus_scans": [
    {
      "scan_id": "uuid",
      "scan_timestamp": "timestamp",
      "scanner_engine": "ClamAV",
      "scanner_version": "1.0.0",
      "scan_status": "clean|infected|suspicious|error",
      "threats_detected": ["threat1", "threat2"],
      "scan_duration_ms": 1500,
      "correlation_id": "uuid"
    }
  ]
}
```

##### POST /admin/api/v1/content/{id}/virus-scan
**Description**: Trigger manual virus scan for content  
**Access Level**: Admin role required  
**Path Parameters**: `id` (UUID): Content identifier

**Response**: 202 Accepted
```json
{
  "scan_id": "uuid",
  "content_id": "uuid",
  "scan_status": "queued",
  "estimated_completion": "timestamp"
}
```

#### Storage Backend Management

##### GET /admin/api/v1/content/storage-backends
**Description**: List configured storage backends with health status  
**Access Level**: Admin role required

**Response**: 200 OK
```json
{
  "storage_backends": [
    {
      "backend_id": "uuid",
      "backend_name": "primary-blob-storage",
      "backend_type": "azure-blob|local-filesystem",
      "is_active": true,
      "priority_order": 1,
      "health_status": "healthy|degraded|unhealthy|unknown",
      "last_health_check": "timestamp",
      "configuration": {
        "base_url": "https://storage.blob.core.windows.net",
        "container": "content"
      },
      "statistics": {
        "total_content_items": 1500,
        "total_storage_used": 5368709120,
        "average_response_time": 125.5
      }
    }
  ]
}
```

##### POST /admin/api/v1/content/storage-backends/{id}/health-check
**Description**: Trigger manual health check for storage backend  
**Access Level**: Admin role required  
**Path Parameters**: `id` (UUID): Backend identifier

**Response**: 200 OK
```json
{
  "backend_id": "uuid",
  "health_status": "healthy|degraded|unhealthy",
  "response_time_ms": 150,
  "last_health_check": "timestamp",
  "health_details": {
    "connectivity": "ok",
    "authentication": "ok",
    "write_test": "ok",
    "read_test": "ok"
  }
}
```

#### Content Lifecycle Management

##### GET /admin/api/v1/content/lifecycle/orphaned
**Description**: List content files without database references (cleanup candidates)  
**Access Level**: Admin role required  
**Query Parameters**:
- `older_than_days` (int, optional): Files older than X days, default 30
- `storage_backend` (UUID, optional): Specific storage backend

**Response**: 200 OK
```json
{
  "orphaned_content": [
    {
      "storage_path": "string",
      "file_size": 1048576,
      "last_modified": "timestamp",
      "storage_backend": "string",
      "estimated_cleanup_impact": "low|medium|high"
    }
  ],
  "summary": {
    "total_orphaned_files": 25,
    "total_wasted_storage": 26843545600,
    "oldest_orphaned_file": "timestamp"
  }
}
```

##### POST /admin/api/v1/content/lifecycle/cleanup
**Description**: Execute orphaned content cleanup  
**Access Level**: Admin role required  
**Request Body**:
```json
{
  "older_than_days": 30,
  "dry_run": true,
  "storage_backend": "uuid|null"
}
```

**Response**: 200 OK
```json
{
  "cleanup_results": {
    "files_processed": 25,
    "files_deleted": 20,
    "files_skipped": 5,
    "storage_reclaimed": 21474836480,
    "errors": []
  }
}
```

## Audit Endpoints

### Services Audit

##### GET /admin/api/v1/services/{id}/audit
**Description**: Retrieve complete audit history for service from Grafana Cloud Loki  
**Access Level**: Admin role required  
**Path Parameters**: `id` (UUID): Service identifier  
**Query Parameters**: Standard pagination and time range parameters

**Response**: 200 OK
```json
{
  "service_id": "uuid",
  "audit_records": [
    {
      "audit_id": "uuid",
      "entity_type": "service",
      "operation_type": "INSERT|UPDATE|DELETE",
      "audit_timestamp": "timestamp",
      "user_id": "string|null",
      "correlation_id": "uuid",
      "trace_id": "string",
      "data_snapshot": {
        "before": {...},
        "after": {...}
      },
      "environment": "production|development"
    }
  ],
  "pagination": {...},
  "grafana_query_info": {
    "query_duration_ms": 150,
    "total_log_entries": 1500,
    "query_range": "24h"
  }
}
```

### Content Audit

##### GET /admin/api/v1/content/{id}/audit
**Description**: Retrieve complete audit history for content from Grafana Cloud Loki  
**Access Level**: Admin role required  
**Path Parameters**: `id` (UUID): Content identifier  
**Query Parameters**: Standard pagination and time range parameters

**Response**: 200 OK (similar structure to services audit with entity_type: "content")

## Health Check Endpoints

##### GET /health
**Description**: Basic health check for Dapr service mesh  
**Access Level**: Anonymous

**Response**: 200 OK
```json
{
  "status": "healthy",
  "timestamp": "timestamp",
  "service": "services-api|content-api",
  "version": "string",
  "dapr_app_id": "services-api|content-api",
  "correlation_id": "uuid"
}
```

##### GET /health/ready
**Description**: Readiness probe for Podman container orchestration  
**Access Level**: Anonymous

**Response**: 200 OK
```json
{
  "status": "ready",
  "checks": {
    "database": "healthy",
    "dapr_sidecar": "healthy",
    "cosmosdb_state_store": "healthy",
    "azure_service_bus": "healthy",
    "vault_secret_store": "healthy",
    "grafana_telemetry": "healthy"
  },
  "environment": "${ENVIRONMENT}",
  "dapr_app_id": "services-api|content-api"
}
```

## Standard Error Response Format

All error responses follow consistent structure with Dapr tracing:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message",
    "details": "Additional error details",
    "correlation_id": "uuid",
    "trace_id": "trace_id",
    "span_id": "span_id",
    "service_name": "services-api|content-api",
    "timestamp": "timestamp"
  }
}
```

## Standard HTTP Error Codes

- `400`: Bad Request - Invalid parameters or request format
- `401`: Unauthorized - Invalid or missing authentication
- `403`: Forbidden - Valid authentication but insufficient permissions
- `404`: Not Found - Resource not found or not accessible
- `409`: Conflict - Resource conflict (e.g., duplicate slug)
- `413`: Payload Too Large - File upload exceeds size limit
- `415`: Unsupported Media Type - Invalid content type for upload
- `422`: Unprocessable Entity - Business rule violation or virus detected
- `429`: Too Many Requests - Rate limit exceeded
- `500`: Internal Server Error - Server processing error
- `503`: Service Unavailable - Service or storage backend unavailable

## Request/Response Headers

### Required Request Headers (Admin APIs)
- `Authorization: Bearer <jwt_token>` (Authentik OAuth2)
- `Content-Type: application/json` (or multipart/form-data for uploads)
- `X-Correlation-ID: <uuid>` (optional, generated by Dapr if not provided)

### Observability Headers
- `X-Trace-ID: <trace_id>` (Grafana distributed tracing)
- `X-Span-ID: <span_id>` (Grafana span tracking)
- `X-Request-ID: <uuid>` (Request tracking across Dapr services)

### Standard Response Headers
- `X-Correlation-ID: <uuid>` (Dapr correlation tracking)
- `X-Trace-ID: <trace_id>` (Grafana trace propagation)
- `X-RateLimit-Remaining: <count>` (Dapr rate limit middleware)
- `X-RateLimit-Reset: <timestamp>` (Dapr rate limit middleware)
- `Content-Type: application/json`

### Content Delivery Headers
- `Content-Security-Policy: default-src 'none'`
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-Content-Hash: <sha256_hash>` (Azure Blob Storage integrity)
- `X-Storage-Backend: azure-blob` (Dapr binding identifier)
- `Cache-Control: public, max-age=31536000` (for immutable content)

## Business Rules Enforcement

1. **Access Level Enforcement**: Content access based on authentication and access_level
2. **Content Integrity**: SHA-256 hash verification on upload and download
3. **Virus Scanning**: All uploads scanned before marking as available
4. **Storage Immutability**: Content files never modified after successful upload
5. **Deduplication**: Identical content (same hash) stored only once
6. **Soft Delete**: No physical deletion of content or metadata
7. **Audit Trail**: All admin operations generate immutable audit records
8. **Access Logging**: Every content access logged for analytics and security
9. **Category Assignment**: All services must have valid category_id
10. **Slug Uniqueness**: Service slugs must be unique across active services
11. **Featured Categories**: Maximum 2 categories can be featured
12. **Publishing Rules**: Only published services visible via public API
13. **Correlation Tracking**: All operations tracked with correlation IDs for compliance