# TDD Infrastructure Provisioning Agent

## Agent Overview

**Agent Type**: Infrastructure Provisioning (Phase 1 - Solo Execution)
**Primary Responsibility**: Establish shared infrastructure foundation for all integration testing
**Execution Order**: Runs first before all parallel agents
**Success Criteria**: Complete Podman Compose environment ready for parallel agent integration testing

## Architecture Context

**Environment**: Development with aggressive migration approach
**Runtime**: Podman Compose with containerd runtime
**Infrastructure Scope**: Complete local development environment with production parity
**Migration Strategy**: Aggressive - always migrate to latest version with automatic rollback via environment reset
**Timeout**: 30 seconds maximum for all migration operations

## Infrastructure Components Scope

### Dapr Control Plane Services
- **dapr-placement**: Service placement coordination
- **dapr-sentry**: Certificate authority and identity management
- **redis-dapr**: State store for Dapr control plane

### Core Infrastructure Services  
- **postgresql**: Primary relational database with multiple database support
- **azurite**: Azure Blob Storage emulator for content storage
- **cosmosdb-emulator**: Azure CosmosDB emulator for gateway state
- **service-bus-emulator**: Azure Service Bus emulator for pub/sub
- **sql-edge**: SQL Server for Service Bus emulator dependency

### Identity and Security Stack
- **authentik**: OAuth2/OIDC identity provider
- **authentik-postgresql**: Authentik database backend
- **authentik-redis**: Authentik session storage
- **authentik-worker**: Authentik background task processor
- **opa**: Open Policy Agent for authorization policies
- **vault**: HashiCorp Vault for secrets management

### Observability Stack
- **grafana**: Dashboards and visualization
- **mimir**: Metrics collection and storage
- **loki**: Log aggregation and querying
- **tempo**: Distributed tracing
- **pyroscope**: Continuous profiling
- **grafana-agent**: Telemetry collection agent

## TDD Cycle Structure

### Red Phase: Infrastructure Health Tests

#### Test Categories and Files

**Environment Readiness Tests**
- **File**: `infrastructure/tests/environment_test.go`
- **Purpose**: Validate Podman Compose environment startup
- **Timeout**: 30 seconds

```go
func TestPodmanComposeEnvironmentStartup(t *testing.T) {
    // Test: All infrastructure services start successfully
    // Test: All services report healthy status
    // Test: Network connectivity between services established
}

func TestEnvironmentVariableConfiguration(t *testing.T) {
    // Test: All required environment variables present
    // Test: No hardcoded network configuration
    // Test: Container-only configuration compliance
}
```

**Database Migration Tests**
- **File**: `infrastructure/migrations/migration_test.go`
- **Purpose**: Validate aggressive migration execution
- **Timeout**: 30 seconds

```go
func TestDatabaseMigrationExecution(t *testing.T) {
    // Test: PostgreSQL migrations execute successfully
    // Test: Schema compliance with SERVICES-SCHEMA.md
    // Test: Schema compliance with CONTENT-SCHEMA.md
    // Test: Migration rollback capability via environment reset
}

func TestMultiDatabaseSupport(t *testing.T) {
    // Test: Services database initialization
    // Test: Content database initialization
    // Test: Authentik database initialization
    // Test: Database isolation and access control
}
```

**Dapr Control Plane Tests**
- **File**: `infrastructure/tests/dapr_test.go`
- **Purpose**: Validate Dapr control plane functionality
- **Timeout**: 15 seconds

```go
func TestDaprControlPlaneHealth(t *testing.T) {
    // Test: Dapr placement service operational
    // Test: Dapr sentry service operational
    // Test: Redis state store connectivity
    // Test: Service discovery functionality
}

func TestDaprServiceRegistration(t *testing.T) {
    // Test: Service registration capability
    // Test: Health check endpoint responsiveness
    // Test: Service invocation readiness
}
```

**Azure Emulator Integration Tests**
- **File**: `infrastructure/tests/azure_emulators_test.go`
- **Purpose**: Validate Azure service emulation
- **Timeout**: 15 seconds

```go
func TestAzuriteStorageEmulator(t *testing.T) {
    // Test: Azurite blob storage container creation
    // Test: Blob storage read/write operations
    // Test: Storage backend health monitoring
}

func TestCosmosDBEmulator(t *testing.T) {
    // Test: CosmosDB emulator database creation
    // Test: Document operations functionality
    // Test: Query execution capability
}

func TestServiceBusEmulator(t *testing.T) {
    // Test: Service Bus queue creation
    // Test: Message publishing capability
    // Test: Message consumption functionality
}
```

**Security Infrastructure Tests**
- **File**: `infrastructure/tests/security_test.go`
- **Purpose**: Validate security service integration
- **Timeout**: 15 seconds

```go
func TestAuthentikIdentityProvider(t *testing.T) {
    // Test: Authentik service startup and configuration
    // Test: OAuth2 endpoint availability
    // Test: JWT token validation capability
    // Test: Anonymous access configuration
}

func TestVaultSecretsManagement(t *testing.T) {
    // Test: Vault service initialization
    // Test: Secret storage and retrieval
    // Test: Policy enforcement
}

func TestOPAPolicyEngine(t *testing.T) {
    // Test: OPA service startup
    // Test: Policy loading and compilation
    // Test: Policy evaluation endpoint availability
}
```

**Observability Infrastructure Tests**
- **File**: `infrastructure/tests/observability_test.go`
- **Purpose**: Validate monitoring and observability stack
- **Timeout**: 15 seconds

```go
func TestGrafanaObservabilityStack(t *testing.T) {
    // Test: Grafana service availability
    // Test: Mimir metrics storage connectivity
    // Test: Loki log aggregation functionality
    // Test: Tempo tracing capability
    // Test: Pyroscope profiling integration
}

func TestTelemetryCollection(t *testing.T) {
    // Test: Grafana Agent configuration
    // Test: Telemetry endpoint connectivity
    // Test: Data pipeline functionality
}
```

### Green Phase: Minimal Infrastructure Implementation

#### Infrastructure Service Configuration

**Podman Compose Service Definitions**
- **File**: `podman-compose.yml`
- **Purpose**: Define all infrastructure services with proper dependencies
- **Environment**: Container-only configuration

**Database Migration Implementation**
- **File**: `infrastructure/migrations/migrator.go`
- **Purpose**: Implement aggressive migration strategy
- **Features**: Schema compliance validation, rollback via environment reset

**Dapr Configuration**
- **File**: `dapr/config.yaml`
- **Purpose**: Dapr control plane configuration
- **Components**: Service discovery, state management, pub/sub

**Environment Variable Management**
- **File**: `.env.development`
- **Purpose**: Development environment configuration
- **Compliance**: Container-only configuration pattern

### Refactor Phase: Infrastructure Optimization

#### Performance Optimization
- Container startup sequence optimization
- Resource allocation tuning
- Network configuration optimization
- Volume mount performance improvements

#### Reliability Improvements
- Health check implementation and tuning
- Retry logic for service dependencies
- Graceful shutdown handling
- Error recovery mechanisms

#### Monitoring and Observability Enhancement
- Infrastructure metrics collection
- Performance monitoring dashboards
- Alert configuration for infrastructure failures
- Log aggregation optimization

## Infrastructure Validation Checklist

### Service Health Validation
```bash
# All services must report healthy status
podman-compose ps --filter health=healthy

# Expected healthy services count: 15+
# Services: postgresql, azurite, cosmosdb-emulator, service-bus-emulator, 
#          sql-edge, vault, authentik, authentik-postgresql, authentik-redis,
#          authentik-worker, opa, grafana, mimir, loki, tempo, pyroscope,
#          grafana-agent, dapr-placement, dapr-sentry, redis-dapr
```

### Database Migration Validation
```bash
# Verify migrations completed successfully
# Check schema compliance with authoritative schema files
# Validate multiple database initialization
# Confirm no audit tables in PostgreSQL (compliance requirement)
```

### Network Connectivity Validation
```bash
# Verify service-to-service connectivity
# Validate port configuration from environment variables
# Confirm no hardcoded network configuration
# Test Dapr service invocation functionality
```

### Integration Test Readiness Validation
```bash
# Confirm all infrastructure services accessible
# Validate configuration consistency
# Verify audit event pipeline to Grafana Cloud Loki
# Test complete environment reset capability
```

## Agent Completion Signal

### Success Criteria
1. All infrastructure services healthy and operational
2. Database migrations completed with schema compliance
3. Dapr control plane functional for service discovery
4. Azure emulators operational for integration testing
5. Security services configured for authentication/authorization
6. Observability stack operational for telemetry collection
7. Environment reset capability verified for rollback scenarios

### Handoff to Parallel Agents
- **Signal**: Infrastructure health validation complete
- **Shared Resources**: Complete Podman Compose environment
- **Integration Points**: Database connections, Dapr service discovery, Azure emulator endpoints
- **Monitoring**: Grafana dashboards available for parallel agent monitoring

### Failure Handling
- **Rollback Strategy**: Complete environment destruction and recreation
- **Timeout Enforcement**: 30 seconds maximum for any infrastructure operation
- **Error Reporting**: Structured logging to development console
- **Recovery Process**: Automated retry with exponential backoff

## Environment Configuration Reference

### Required Environment Variables
```bash
# Database Configuration
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=development
POSTGRES_DB=international_center
POSTGRES_MULTIPLE_DATABASES=services_db,content_db

# Dapr Configuration
DAPR_PLACEMENT_PORT=6050
DAPR_SENTRY_PORT=6051
REDIS_DAPR_PORT=6379

# Azure Emulators
AZURITE_BLOB_PORT=10000
COSMOSDB_PORT=8081
SERVICE_BUS_PORT=5672

# Security Services
AUTHENTIK_PORT=9000
VAULT_PORT=8200
OPA_PORT=8181

# Observability
GRAFANA_PORT=3000
MIMIR_PORT=9009
LOKI_PORT=3100
TEMPO_PORT=3200
PYROSCOPE_PORT=4040

# Migration Configuration
MIGRATION_ENVIRONMENT=development
MIGRATION_APPROACH=aggressive
MIGRATION_TIMEOUT=30s
```

## Integration with Parallel Agents

### Contract Definitions
- **Database Schema Contracts**: SERVICES-SCHEMA.md and CONTENT-SCHEMA.md authoritative
- **Service Discovery**: Dapr control plane provides service registration
- **Configuration Management**: Container-only environment variables
- **Health Monitoring**: Grafana stack available for all agents

### Shared Infrastructure Services
- **PostgreSQL**: Services and Content domain databases
- **Azure Emulators**: Storage, CosmosDB, Service Bus integration
- **Dapr Control Plane**: Service mesh and middleware functionality
- **Security Services**: Authentik OAuth2, OPA policies, Vault secrets
- **Observability**: Grafana stack for metrics, logs, traces, profiling

### Agent Coordination Points
- **Infrastructure Readiness**: All parallel agents wait for completion signal
- **Environment Reset**: Infrastructure agent handles full environment rollback
- **Configuration Updates**: Infrastructure agent manages environment variable changes
- **Health Monitoring**: Shared observability stack for all agent monitoring