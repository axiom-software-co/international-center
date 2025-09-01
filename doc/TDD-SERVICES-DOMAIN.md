# TDD Services Domain Agent

## Agent Overview

**Agent Type**: Services Domain (Phase 2 - Parallel Execution)
**Primary Responsibility**: Services API implementation, business logic, and domain-specific workflows
**Execution Order**: Runs in parallel with Content Domain and Gateway Integration agents
**Dependencies**: Infrastructure Provisioning Agent completion
**Success Criteria**: Complete services-api with PostgreSQL integration and audit compliance

## Architecture Context

**Domain**: Services management with categories and publishing workflows
**Database**: PostgreSQL with services domain schema compliance
**API Patterns**: Public (anonymous) and Admin (authenticated) endpoints
**Integration**: Dapr service mesh, Grafana Cloud Loki audit events
**Schema Authority**: SERVICES-SCHEMA.md is authoritative for all database changes

## Services Domain Scope

### Core Business Entities
- **Services**: Primary domain entity with publishing workflow
- **Service Categories**: Organization and categorization logic
- **Featured Categories**: Highlighting and positioning logic

### Service Layers Implementation
- **Handler Layer**: HTTP request/response, routing, middleware integration
- **Service Layer**: Business logic, domain rules, workflow orchestration
- **Repository Layer**: Data access, PostgreSQL integration, audit event publishing

### Database Schema Compliance
- **Services Table**: Complete CRUD with business rule enforcement
- **Service Categories Table**: Category management with default assignment logic  
- **Featured Categories Table**: Position constraints and business rule validation
- **Audit Integration**: Direct Grafana Cloud Loki event publishing (no local audit tables)

## Contract Interfaces

### API Endpoint Contracts
**Public Services API**: `/api/v1/services/*` (Anonymous access via public-gateway)
**Admin Services API**: `/admin/api/v1/services/*` (Authenticated access via admin-gateway)

### Database Schema Contract
**Authoritative Source**: `/doc/dapr-services/SERVICES-SCHEMA.md`
**Compliance Requirement**: Implementation must exactly match schema specification

### Audit Event Contract
**Destination**: Grafana Cloud Loki via Dapr pub/sub
**Structure**: Unified audit event format across all domains
**Immutability**: Audit events never modified or deleted

## TDD Cycle Structure

### Red Phase: Services Domain Contract Tests

#### Test Categories and Files

**Services Entity Tests**
- **File**: `services-api/internal/domain/services/service_test.go`
- **Purpose**: Services entity behavior validation
- **Timeout**: 5 seconds

```go
func TestServiceEntityCreation(t *testing.T) {
    // Test: Service creation with required fields
    // Test: Slug generation and uniqueness validation
    // Test: Publishing status state machine
    // Test: Category assignment logic
    // Test: Business rule validation
}

func TestServicePublishingWorkflow(t *testing.T) {
    // Test: Draft → Published transition
    // Test: Published → Archived transition
    // Test: Business rule enforcement for each state
    // Test: Audit event generation for state changes
}

func TestServiceCategoryAssignment(t *testing.T) {
    // Test: Valid category assignment
    // Test: Default unassigned category fallback
    // Test: Category deletion impact on services
    // Test: Category reassignment workflows
}
```

**Service Categories Tests**
- **File**: `services-api/internal/domain/categories/category_test.go`
- **Purpose**: Category management and business rules
- **Timeout**: 5 seconds

```go
func TestCategoryManagement(t *testing.T) {
    // Test: Category creation and validation
    // Test: Default unassigned category constraint (exactly one)
    // Test: Category deletion with service reassignment
    // Test: Order number management
}

func TestFeaturedCategoryConstraints(t *testing.T) {
    // Test: Featured category position constraints (1 and 2)
    // Test: Default unassigned category cannot be featured
    // Test: Featured category business rule enforcement
}
```

**Repository Layer Tests**
- **File**: `services-api/internal/repositories/services_repository_test.go`
- **Purpose**: Database integration and schema compliance
- **Timeout**: 15 seconds (Integration test)

```go
func TestServicesRepositoryIntegration(t *testing.T) {
    // Test: CRUD operations with PostgreSQL
    // Test: Schema compliance with SERVICES-SCHEMA.md
    // Test: Database constraints and indexes
    // Test: Transaction management
    // Test: Connection pooling and error handling
}

func TestAuditEventPublishing(t *testing.T) {
    // Test: Audit event creation for all operations
    // Test: Grafana Cloud Loki event publishing via Dapr
    // Test: Event correlation and tracing
    // Test: No local audit table storage (compliance)
}
```

**Service Layer Tests**
- **File**: `services-api/internal/services/services_service_test.go`
- **Purpose**: Business logic coordination and workflow management
- **Timeout**: 5 seconds

```go
func TestServicesBusinessLogic(t *testing.T) {
    // Test: Service creation workflow with validation
    // Test: Publishing workflow orchestration
    // Test: Category assignment business rules
    // Test: Content URL management (Azure Blob Storage integration)
}

func TestServiceQueryOperations(t *testing.T) {
    // Test: Service listing with pagination
    // Test: Service filtering by category and status
    // Test: Service search functionality
    // Test: Featured services retrieval
}
```

**Handler Layer Tests**
- **File**: `services-api/internal/handlers/services_handler_test.go`
- **Purpose**: HTTP API contract validation
- **Timeout**: 15 seconds (Integration test)

```go
func TestPublicServicesAPIEndpoints(t *testing.T) {
    // Test: GET /api/v1/services (anonymous access)
    // Test: GET /api/v1/services/{id} (anonymous access)
    // Test: GET /api/v1/services/slug/{slug} (anonymous access)
    // Test: GET /api/v1/services/featured (anonymous access)
    // Test: Error response format compliance
}

func TestAdminServicesAPIEndpoints(t *testing.T) {
    // Test: POST /admin/api/v1/services (authenticated)
    // Test: PUT /admin/api/v1/services/{id} (authenticated)
    // Test: POST /admin/api/v1/services/{id}/publish (authenticated)
    // Test: DELETE /admin/api/v1/services/{id} (authenticated)
    // Test: Authorization and role-based access control
}
```

**Health Check Tests**
- **File**: `services-api/internal/health/health_test.go`
- **Purpose**: Service health and readiness validation
- **Timeout**: 5 seconds

```go
func TestServicesAPIHealthChecks(t *testing.T) {
    // Test: /health endpoint response format
    // Test: /health/ready endpoint dependency checking
    // Test: PostgreSQL connection health validation
    // Test: Dapr sidecar connectivity verification
}
```

### Green Phase: Minimal Services Implementation

#### Project Structure Creation
```
services-api/
├── cmd/
│   └── services-api/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── domain/
│   │   ├── services/
│   │   │   ├── service.go
│   │   │   └── service_test.go
│   │   └── categories/
│   │       ├── category.go
│   │       └── category_test.go
│   ├── repositories/
│   │   ├── services_repository.go
│   │   ├── services_repository_test.go
│   │   ├── categories_repository.go
│   │   └── categories_repository_test.go
│   ├── services/
│   │   ├── services_service.go
│   │   ├── services_service_test.go
│   │   ├── categories_service.go
│   │   └── categories_service_test.go
│   ├── handlers/
│   │   ├── services_handler.go
│   │   ├── services_handler_test.go
│   │   ├── categories_handler.go
│   │   └── categories_handler_test.go
│   ├── middleware/
│   │   └── logging.go
│   └── health/
│       ├── health.go
│       └── health_test.go
├── migrations/
│   ├── 000001_services_schema.up.sql
│   └── 000001_services_schema.down.sql
├── go.mod
├── go.sum
├── Dockerfile
└── README.md
```

#### Domain Entity Implementation

**Service Entity**
- **File**: `services-api/internal/domain/services/service.go`
- **Purpose**: Core business entity with validation and behavior

**Category Entity**  
- **File**: `services-api/internal/domain/categories/category.go`
- **Purpose**: Category management with business rules

#### Repository Implementation

**Services Repository**
- **File**: `services-api/internal/repositories/services_repository.go`
- **Purpose**: PostgreSQL integration with schema compliance
- **Features**: CRUD operations, audit event publishing, connection management

#### Service Layer Implementation

**Services Business Service**
- **File**: `services-api/internal/services/services_service.go`
- **Purpose**: Business logic orchestration and workflow management
- **Features**: Publishing workflow, category assignment, validation

#### Handler Implementation

**Services HTTP Handlers**
- **File**: `services-api/internal/handlers/services_handler.go`
- **Purpose**: HTTP API implementation with Dapr integration
- **Features**: Public and admin endpoints, error handling, request validation

### Refactor Phase: Services Domain Optimization

#### Performance Optimization
- Database query optimization with proper indexing
- Connection pooling configuration and tuning
- Caching strategy for frequently accessed data
- Pagination optimization for large result sets

#### Business Logic Enhancement
- Advanced validation rules and constraints
- Workflow automation and state management
- Content URL management optimization
- Category management feature enhancement

#### Integration Improvements  
- Dapr service invocation optimization
- Audit event publishing reliability
- Error handling and recovery mechanisms
- Observability and monitoring integration

## Services Domain Integration Points

### Database Integration
- **PostgreSQL Connection**: Via environment-configured connection string
- **Schema Compliance**: Exact match with SERVICES-SCHEMA.md specification
- **Migration Management**: Handled by Infrastructure Provisioning Agent
- **Audit Compliance**: No local audit tables, direct Grafana Cloud Loki integration

### Dapr Service Mesh Integration
- **Service Registration**: Dapr app-id: `services-api`
- **Service Invocation**: Available for gateway routing
- **Pub/Sub Integration**: Audit events to `services-audit-events` topic
- **State Management**: Optional caching via Dapr state store

### Gateway Integration Contracts
- **Public Gateway**: Anonymous access to published services
- **Admin Gateway**: Authenticated access with role-based permissions
- **Route Definitions**: Specified in API-ENDPOINTS.md
- **Error Response Format**: Standardized across all services

### Audit and Compliance Integration
- **Event Publishing**: Direct to Grafana Cloud Loki via Dapr pub/sub
- **Event Structure**: Unified format with services domain entity types
- **Correlation Tracking**: Request correlation and distributed tracing
- **Compliance Requirements**: Immutable audit trail, no local storage

## Service Configuration

### Environment Variables
```bash
# Service Configuration
SERVICES_API_PORT=${SERVICES_API_PORT}
DAPR_HTTP_PORT=${SERVICES_API_DAPR_HTTP_PORT}
DAPR_GRPC_PORT=${SERVICES_API_DAPR_GRPC_PORT}

# Database Configuration
DATABASE_CONNECTION_STRING=${SERVICES_DATABASE_CONNECTION_STRING}
DATABASE_MAX_CONNECTIONS=${DATABASE_MAX_CONNECTIONS}
DATABASE_CONNECTION_TIMEOUT=${DATABASE_CONNECTION_TIMEOUT}

# Audit Configuration  
GRAFANA_CLOUD_LOKI_ENDPOINT=${GRAFANA_CLOUD_LOKI_ENDPOINT}
AUDIT_EVENTS_TOPIC=services-audit-events

# Observability
GRAFANA_ENDPOINT=${GRAFANA_ENDPOINT}
TELEMETRY_SAMPLING_RATE=${TELEMETRY_SAMPLING_RATE}
LOG_LEVEL=${LOG_LEVEL}

# Content Storage
BLOB_STORAGE_ENDPOINT=${BLOB_STORAGE_ENDPOINT}
CONTENT_BASE_URL=${CONTENT_BASE_URL}
```

### Health Check Configuration
```bash
# Health Check Endpoints
HEALTH_CHECK_INTERVAL=30s
READINESS_CHECK_TIMEOUT=10s
LIVENESS_CHECK_TIMEOUT=5s

# Dependency Health Checks
POSTGRESQL_HEALTH_CHECK=true
DAPR_HEALTH_CHECK=true
BLOB_STORAGE_HEALTH_CHECK=true
```

## Agent Success Criteria

### Unit Test Compliance
- All business logic tests pass with 5 seconds timeout
- Domain entity validation and behavior tests complete
- Service layer orchestration tests validated
- Mock-based isolation for external dependencies

### Integration Test Compliance
- PostgreSQL integration tests pass with 15 seconds timeout
- Schema compliance validation with SERVICES-SCHEMA.md
- Audit event publishing verification to Grafana Cloud Loki
- HTTP API contract validation for public and admin endpoints

### Schema Compliance Validation
- Services table implementation matches specification exactly
- Service categories table with proper constraints
- Featured categories table with business rule enforcement
- Database indexes and performance optimization implemented

### API Contract Compliance
- Public services endpoints accessible via public-gateway
- Admin services endpoints accessible via admin-gateway  
- Error response format matches standardized structure
- Authentication and authorization integration functional

### Audit and Compliance Verification
- All CRUD operations generate audit events
- Audit events published to Grafana Cloud Loki successfully
- No local audit tables present in database
- Correlation and tracing integration operational

## Coordination with Parallel Agents

### Shared Infrastructure Dependencies
- **PostgreSQL**: Database connection and schema compliance
- **Dapr Control Plane**: Service registration and discovery
- **Grafana Stack**: Audit event publishing and observability

### Gateway Integration Readiness
- **Service Discovery**: services-api registered with Dapr
- **Health Endpoints**: Available for gateway health checks
- **API Contracts**: Public and admin endpoints ready for routing

### Content Domain Coordination
- **Shared Audit Format**: Consistent event structure with content domain
- **Database Independence**: No cross-domain database dependencies
- **Service Mesh Integration**: Independent service registration

### End-to-End Validation Preparation
- **Complete API Implementation**: All endpoints functional
- **Business Workflow Validation**: Publishing workflows operational
- **Audit Trail Completeness**: Full audit event coverage
- **Performance Baseline**: Response time and throughput metrics established