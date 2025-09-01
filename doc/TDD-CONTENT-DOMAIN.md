# TDD Content Domain Agent

## Agent Overview

**Agent Type**: Content Domain (Phase 2 - Parallel Execution)
**Primary Responsibility**: Content API implementation, file processing, storage integration, and content delivery workflows
**Execution Order**: Runs in parallel with Services Domain and Gateway Integration agents
**Dependencies**: Infrastructure Provisioning Agent completion
**Success Criteria**: Complete content-api with Azure Blob Storage integration, virus scanning, and audit compliance

## Architecture Context

**Domain**: Content management with upload, processing, and delivery workflows
**Database**: PostgreSQL with content domain schema compliance
**Storage**: Azure Blob Storage via Dapr bindings for cloud-native portability
**Processing**: Async virus scanning, hash-based deduplication, metadata extraction
**Schema Authority**: CONTENT-SCHEMA.md is authoritative for all database changes

## Content Domain Scope

### Core Business Entities
- **Content**: Primary domain entity with upload and processing workflow
- **Content Access Log**: Analytics and monitoring for content access patterns
- **Content Virus Scan**: Security scanning results and status tracking
- **Content Storage Backend**: Multi-backend storage management with health monitoring

### Service Layers Implementation
- **Handler Layer**: HTTP multipart upload, download, preview endpoints
- **Service Layer**: Content processing orchestration, workflow management
- **Repository Layer**: PostgreSQL integration, Azure Blob Storage via Dapr bindings

### Content Processing Pipeline
- **Upload Phase**: Multipart form handling, metadata extraction, hash calculation
- **Processing Phase**: Async virus scanning via Dapr pub/sub
- **Storage Phase**: Azure Blob Storage via Dapr bindings with path management
- **Access Phase**: Content delivery with access logging and analytics

## Contract Interfaces

### API Endpoint Contracts
**Public Content API**: `/api/v1/content/*` (Anonymous access to approved content)
**Admin Content API**: `/admin/api/v1/content/*` (Authenticated access with upload capabilities)

### Database Schema Contract
**Authoritative Source**: `/doc/dapr-services/CONTENT-SCHEMA.md`
**Compliance Requirement**: Implementation must exactly match schema specification

### Storage Integration Contract
**Azure Blob Storage**: Production via Dapr binding `blob-storage`
**Azurite Emulator**: Development via Dapr binding `blob-storage-local`
**Path Structure**: `{environment}/content/{year}/{month}/{content-id}/{content-hash}.{ext}`

### Audit Event Contract
**Destination**: Grafana Cloud Loki via Dapr pub/sub
**Structure**: Unified audit event format for content domain entities
**Content Integrity**: SHA-256 hash verification and deduplication

## TDD Cycle Structure

### Red Phase: Content Domain Contract Tests

#### Test Categories and Files

**Content Entity Tests**
- **File**: `content-api/internal/domain/content/content_test.go`
- **Purpose**: Content entity behavior and validation
- **Timeout**: 5 seconds

```go
func TestContentEntityCreation(t *testing.T) {
    // Test: Content creation with required metadata
    // Test: Hash calculation and integrity validation
    // Test: Upload status state machine
    // Test: Access level enforcement
    // Test: Content classification logic
}

func TestContentProcessingWorkflow(t *testing.T) {
    // Test: Processing → Available transition
    // Test: Processing → Failed transition
    // Test: Processing → Archived transition
    // Test: Virus scanning integration
    // Test: Audit event generation for state changes
}

func TestContentIntegrityValidation(t *testing.T) {
    // Test: SHA-256 hash calculation and storage
    // Test: Hash-based deduplication logic
    // Test: File size validation and constraints
    // Test: MIME type detection and validation
}
```

**Content Processing Tests**
- **File**: `content-api/internal/domain/processing/processing_test.go`
- **Purpose**: Content processing workflow validation
- **Timeout**: 5 seconds

```go
func TestContentUploadProcessing(t *testing.T) {
    // Test: Multipart form parsing and validation
    // Test: Metadata extraction from uploaded files
    // Test: Correlation ID assignment and tracking
    // Test: Processing attempt management
}

func TestVirusScanningIntegration(t *testing.T) {
    // Test: Virus scan request creation
    // Test: Scan result processing and storage
    // Test: Threat detection handling
    // Test: Scan status updates and notifications
}
```

**Storage Backend Tests**
- **File**: `content-api/internal/domain/storage/storage_backend_test.go`
- **Purpose**: Storage backend management and health monitoring
- **Timeout**: 5 seconds

```go
func TestStorageBackendManagement(t *testing.T) {
    // Test: Backend registration and configuration
    // Test: Priority-based backend selection
    // Test: Health monitoring and status tracking
    // Test: Failover logic and backend switching
}

func TestBlobStorageIntegration(t *testing.T) {
    // Test: Dapr binding configuration
    // Test: Path generation and management
    // Test: Storage operation error handling
    // Test: Configuration security (Azure Key Vault references)
}
```

**Repository Layer Tests**
- **File**: `content-api/internal/repositories/content_repository_test.go`
- **Purpose**: Database integration and schema compliance
- **Timeout**: 15 seconds (Integration test)

```go
func TestContentRepositoryIntegration(t *testing.T) {
    // Test: CRUD operations with PostgreSQL
    // Test: Schema compliance with CONTENT-SCHEMA.md
    // Test: Database constraints and indexes
    // Test: Transaction management and rollback
    // Test: Connection pooling and error handling
}

func TestContentAccessLogging(t *testing.T) {
    // Test: Access log entry creation
    // Test: Analytics data collection
    // Test: Performance metrics tracking
    // Test: IP address and user agent logging
}

func TestAuditEventPublishing(t *testing.T) {
    // Test: Audit event creation for all operations
    // Test: Grafana Cloud Loki event publishing via Dapr
    // Test: Event correlation and tracing
    // Test: No local audit table storage (compliance)
}
```

**Service Layer Tests**
- **File**: `content-api/internal/services/content_service_test.go`
- **Purpose**: Business logic coordination and workflow management
- **Timeout**: 5 seconds

```go
func TestContentBusinessLogic(t *testing.T) {
    // Test: Content upload workflow orchestration
    // Test: Processing pipeline coordination
    // Test: Storage backend selection logic
    // Test: Access control and permission validation
}

func TestContentDeliveryService(t *testing.T) {
    // Test: Content retrieval and access validation
    // Test: Download and preview functionality
    // Test: Cache header management
    // Test: Content streaming and optimization
}

func TestContentAnalyticsService(t *testing.T) {
    // Test: Access pattern analysis
    // Test: Usage metrics collection
    // Test: Performance monitoring data
    // Test: Storage backend performance tracking
}
```

**Handler Layer Tests**
- **File**: `content-api/internal/handlers/content_handler_test.go`
- **Purpose**: HTTP API contract validation
- **Timeout**: 15 seconds (Integration test)

```go
func TestPublicContentAPIEndpoints(t *testing.T) {
    // Test: GET /api/v1/content (anonymous access to approved content)
    // Test: GET /api/v1/content/{id} (anonymous access)
    // Test: GET /api/v1/content/{id}/download (anonymous access)
    // Test: GET /api/v1/content/{id}/preview (anonymous access)
    // Test: Error response format compliance
}

func TestAdminContentAPIEndpoints(t *testing.T) {
    // Test: POST /admin/api/v1/content/upload (authenticated)
    // Test: PUT /admin/api/v1/content/{id} (authenticated)
    // Test: POST /admin/api/v1/content/{id}/reprocess (authenticated)
    // Test: GET /admin/api/v1/content/{id}/status (authenticated)
    // Test: GET /admin/api/v1/content/analytics (authenticated)
}

func TestContentUploadHandler(t *testing.T) {
    // Test: Multipart form upload handling
    // Test: File size limit enforcement
    // Test: MIME type validation
    // Test: Upload correlation tracking
}
```

**Azure Blob Storage Integration Tests**
- **File**: `content-api/internal/storage/blob_storage_test.go`
- **Purpose**: Azure Blob Storage integration via Dapr bindings
- **Timeout**: 15 seconds (Integration test)

```go
func TestDaprBlobStorageBinding(t *testing.T) {
    // Test: Dapr binding configuration and connectivity
    // Test: Blob upload operations via Dapr HTTP API
    // Test: Blob download and streaming
    // Test: Blob metadata and properties management
}

func TestStoragePathManagement(t *testing.T) {
    // Test: Path generation and formatting
    // Test: Content organization by date and ID
    // Test: Hash-based file naming
    // Test: Path immutability enforcement
}

func TestStorageErrorHandling(t *testing.T) {
    // Test: Network failure recovery
    // Test: Storage capacity issues
    // Test: Authentication and authorization errors
    // Test: Retry logic and exponential backoff
}
```

### Green Phase: Minimal Content Implementation

#### Project Structure Creation
```
content-api/
├── cmd/
│   └── content-api/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── domain/
│   │   ├── content/
│   │   │   ├── content.go
│   │   │   └── content_test.go
│   │   ├── processing/
│   │   │   ├── processing.go
│   │   │   └── processing_test.go
│   │   └── storage/
│   │       ├── storage_backend.go
│   │       └── storage_backend_test.go
│   ├── repositories/
│   │   ├── content_repository.go
│   │   ├── content_repository_test.go
│   │   ├── access_log_repository.go
│   │   └── virus_scan_repository.go
│   ├── services/
│   │   ├── content_service.go
│   │   ├── content_service_test.go
│   │   ├── upload_service.go
│   │   ├── processing_service.go
│   │   └── analytics_service.go
│   ├── handlers/
│   │   ├── content_handler.go
│   │   ├── content_handler_test.go
│   │   ├── upload_handler.go
│   │   └── admin_handler.go
│   ├── storage/
│   │   ├── blob_storage.go
│   │   ├── blob_storage_test.go
│   │   └── storage_manager.go
│   ├── middleware/
│   │   ├── upload_middleware.go
│   │   └── access_logging.go
│   └── health/
│       ├── health.go
│       └── health_test.go
├── migrations/
│   ├── 000001_content_schema.up.sql
│   └── 000001_content_schema.down.sql
├── go.mod
├── go.sum
├── Dockerfile
└── README.md
```

#### Domain Entity Implementation

**Content Entity**
- **File**: `content-api/internal/domain/content/content.go`
- **Purpose**: Core content entity with validation and processing workflow

**Storage Backend Entity**
- **File**: `content-api/internal/domain/storage/storage_backend.go`
- **Purpose**: Storage backend management with health monitoring

#### Repository Implementation

**Content Repository**
- **File**: `content-api/internal/repositories/content_repository.go`
- **Purpose**: PostgreSQL integration with schema compliance
- **Features**: CRUD operations, access logging, audit event publishing

#### Service Layer Implementation

**Content Service**
- **File**: `content-api/internal/services/content_service.go`
- **Purpose**: Business logic orchestration and workflow management
- **Features**: Upload processing, virus scanning coordination, delivery optimization

**Upload Service**
- **File**: `content-api/internal/services/upload_service.go`
- **Purpose**: File upload processing and validation
- **Features**: Multipart handling, metadata extraction, hash calculation

#### Storage Integration Implementation

**Blob Storage Service**
- **File**: `content-api/internal/storage/blob_storage.go`
- **Purpose**: Azure Blob Storage integration via Dapr bindings
- **Features**: Upload/download operations, path management, error handling

#### Handler Implementation

**Content HTTP Handlers**
- **File**: `content-api/internal/handlers/content_handler.go`
- **Purpose**: HTTP API implementation with Dapr integration
- **Features**: Public and admin endpoints, multipart upload, streaming download

### Refactor Phase: Content Domain Optimization

#### Performance Optimization
- Streaming upload/download implementation
- Content compression and optimization
- Caching strategy for frequently accessed content
- Storage backend performance tuning

#### Processing Pipeline Enhancement
- Async processing workflow optimization
- Virus scanning performance improvements
- Metadata extraction enhancement
- Error handling and retry logic improvement

#### Storage Management Improvements
- Multi-backend storage optimization
- Health monitoring and alerting
- Storage cost optimization
- Content lifecycle management

## Content Domain Integration Points

### Database Integration
- **PostgreSQL Connection**: Via environment-configured connection string
- **Schema Compliance**: Exact match with CONTENT-SCHEMA.md specification
- **Migration Management**: Handled by Infrastructure Provisioning Agent
- **Audit Compliance**: No local audit tables, direct Grafana Cloud Loki integration

### Azure Blob Storage Integration
- **Dapr Bindings**: Cloud-native storage abstraction
- **Development Environment**: Azurite emulator integration
- **Production Environment**: Azure Blob Storage with authentication
- **Path Management**: Organized content structure with immutable paths

### Dapr Service Mesh Integration
- **Service Registration**: Dapr app-id: `content-api`
- **Pub/Sub Integration**: Virus scanning events, audit events
- **Bindings Integration**: Azure Blob Storage operations
- **Service Invocation**: Available for gateway routing

### Processing Pipeline Integration
- **Virus Scanning**: Async processing via Dapr pub/sub
- **Metadata Extraction**: File analysis and classification
- **Hash Calculation**: SHA-256 integrity verification
- **Deduplication**: Hash-based duplicate detection

## Content Configuration

### Environment Variables
```bash
# Service Configuration
CONTENT_API_PORT=${CONTENT_API_PORT}
DAPR_HTTP_PORT=${CONTENT_API_DAPR_HTTP_PORT}
DAPR_GRPC_PORT=${CONTENT_API_DAPR_GRPC_PORT}

# Database Configuration
DATABASE_CONNECTION_STRING=${CONTENT_DATABASE_CONNECTION_STRING}
DATABASE_MAX_CONNECTIONS=${DATABASE_MAX_CONNECTIONS}
DATABASE_CONNECTION_TIMEOUT=${DATABASE_CONNECTION_TIMEOUT}

# Storage Configuration
BLOB_STORAGE_ENDPOINT=${BLOB_STORAGE_ENDPOINT}
BLOB_STORAGE_BINDING_NAME=blob-storage
STORAGE_CONTAINER_NAME=content
STORAGE_PATH_PREFIX=${ENVIRONMENT}

# Content Processing
MAX_FILE_SIZE=${MAX_FILE_SIZE}
ALLOWED_MIME_TYPES=${ALLOWED_MIME_TYPES}
UPLOAD_TIMEOUT=${UPLOAD_TIMEOUT}
PROCESSING_TIMEOUT=${PROCESSING_TIMEOUT}

# Virus Scanning
VIRUS_SCANNING_ENABLED=${VIRUS_SCANNING_ENABLED}
VIRUS_SCAN_TOPIC=content-virus-scan
VIRUS_SCAN_TIMEOUT=300s

# Audit Configuration
GRAFANA_CLOUD_LOKI_ENDPOINT=${GRAFANA_CLOUD_LOKI_ENDPOINT}
AUDIT_EVENTS_TOPIC=content-audit-events

# Observability
GRAFANA_ENDPOINT=${GRAFANA_ENDPOINT}
TELEMETRY_SAMPLING_RATE=${TELEMETRY_SAMPLING_RATE}
LOG_LEVEL=${LOG_LEVEL}
```

### Content Access Configuration
```bash
# Access Control
PUBLIC_ACCESS_ENABLED=true
INTERNAL_ACCESS_ENABLED=true
RESTRICTED_ACCESS_ENABLED=true

# Analytics and Logging
ACCESS_LOGGING_ENABLED=true
ANALYTICS_ENABLED=true
PERFORMANCE_MONITORING=true

# Cache Configuration
CONTENT_CACHE_TTL=3600
CACHE_HEADERS_ENABLED=true
ETAG_SUPPORT_ENABLED=true
```

## Agent Success Criteria

### Unit Test Compliance
- All business logic tests pass with 5 seconds timeout
- Content entity validation and processing workflow tests complete
- Service layer orchestration tests validated
- Mock-based isolation for external dependencies

### Integration Test Compliance
- PostgreSQL integration tests pass with 15 seconds timeout
- Azure Blob Storage integration via Dapr bindings functional
- Schema compliance validation with CONTENT-SCHEMA.md
- Content upload, processing, and delivery workflows operational

### Storage Integration Validation
- Azure Blob Storage operations via Dapr bindings successful
- Content path management and organization implemented
- Hash-based deduplication functionality verified
- Storage backend health monitoring operational

### Processing Pipeline Validation
- Content upload and processing workflow complete
- Virus scanning integration functional (if enabled)
- Metadata extraction and classification working
- Audit event publishing to Grafana Cloud Loki successful

### API Contract Compliance
- Public content endpoints accessible via public-gateway
- Admin content endpoints accessible via admin-gateway
- Multipart file upload handling operational
- Content download and streaming functional
- Error response format matches standardized structure

## Coordination with Parallel Agents

### Shared Infrastructure Dependencies
- **PostgreSQL**: Database connection and schema compliance
- **Azure Blob Storage**: Content storage via Dapr bindings
- **Dapr Control Plane**: Service registration and pub/sub functionality
- **Grafana Stack**: Audit event publishing and observability

### Gateway Integration Readiness
- **Service Discovery**: content-api registered with Dapr
- **Health Endpoints**: Available for gateway health checks
- **API Contracts**: Public and admin endpoints ready for routing
- **Upload Handling**: Multipart form processing ready

### Services Domain Coordination
- **Shared Audit Format**: Consistent event structure with services domain
- **Database Independence**: No cross-domain database dependencies
- **Content References**: Services can reference content via URLs

### End-to-End Validation Preparation
- **Complete API Implementation**: All content endpoints functional
- **Processing Workflow Validation**: Upload to delivery pipeline operational
- **Audit Trail Completeness**: Full audit event coverage
- **Performance Metrics**: Upload/download performance baselines established
- **Storage Health**: Multi-backend storage monitoring operational