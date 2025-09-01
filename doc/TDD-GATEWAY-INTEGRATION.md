# TDD Gateway Integration Agent

## Agent Overview

**Agent Type**: Gateway Integration (Phase 2 - Parallel Execution)
**Primary Responsibility**: Public and admin gateway implementation with Dapr middleware pipelines
**Execution Order**: Runs in parallel with Services Domain and Content Domain agents
**Dependencies**: Infrastructure Provisioning Agent completion
**Success Criteria**: Complete gateway services with OAuth2+OPA authentication, rate limiting, and CORS

## Architecture Context

**Gateway Pattern**: Reverse proxy with cross-cutting concerns handling
**Middleware Stack**: Dapr sidecar middleware configurations for security and performance
**Authentication**: Authentik OAuth2 provider with OPA policy enforcement
**Rate Limiting**: IP-based (public) and user-based (admin) throttling
**Storage Backend**: Azure CosmosDB for gateway state via Dapr state store

## Gateway Integration Scope

### Core Gateway Services
- **Public Gateway**: Anonymous access with IP-based rate limiting (1000 req/min)
- **Admin Gateway**: Role-based access with user-based rate limiting (100 req/min)
- **Middleware Pipeline**: Comprehensive Dapr middleware configuration
- **Health Monitoring**: Gateway health checks and dependency validation

### Dapr Middleware Configuration
- **Name Resolution**: Service discovery and routing
- **Rate Limiting**: Production-ready throttling with overflow protection
- **CORS Policies**: Cross-origin resource sharing configuration
- **OAuth2 Integration**: Token validation and user context management
- **OPA Authorization**: Policy-driven access control and permission enforcement
- **Bearer Token**: Authentication token processing and validation

### Cross-Cutting Concerns Implementation
- **Security Headers**: Comprehensive security header management
- **Request Correlation**: Distributed tracing and request tracking
- **Error Handling**: Standardized error response formatting
- **Performance Monitoring**: Gateway metrics and performance tracking

## Contract Interfaces

### Gateway Routing Contracts
**Public Gateway Routes**: `/api/v1/*` (Services and Content domain routing)
**Admin Gateway Routes**: `/admin/api/v1/*` (Authenticated domain routing)
**Health Check Routes**: `/health` and `/health/ready` (Gateway status validation)

### Authentication Contract
**Identity Provider**: Authentik OAuth2/OIDC integration
**Token Validation**: JWT token processing and claims extraction
**Anonymous Access**: Public gateway supports unauthenticated requests
**Role-Based Access**: Admin gateway enforces role and permission validation

### Policy Enforcement Contract
**Authorization Engine**: Open Policy Agent (OPA) policy evaluation
**Policy Structure**: Role-based access control with resource permissions
**Fallback Policies**: Default security stance for undefined scenarios
**Policy Distribution**: Centralized policy management and distribution

### State Management Contract
**Gateway State Store**: Azure CosmosDB via Dapr state store binding
**Session Management**: User session tracking and state persistence
**Rate Limiting State**: Request counters and throttling state management
**Cache Management**: Response caching and invalidation strategies

## TDD Cycle Structure

### Red Phase: Gateway Integration Contract Tests

#### Test Categories and Files

**Public Gateway Integration Tests**
- **File**: `public-gateway/internal/gateway/public_gateway_test.go`
- **Purpose**: Public gateway routing and middleware validation
- **Timeout**: 15 seconds (Integration test)

```go
func TestPublicGatewayRouting(t *testing.T) {
    // Test: Services API routing to services-api
    // Test: Content API routing to content-api
    // Test: Health check endpoint responsiveness
    // Test: Anonymous access enforcement
    // Test: IP-based rate limiting functionality
}

func TestPublicGatewayMiddleware(t *testing.T) {
    // Test: CORS policy enforcement
    // Test: Security headers application
    // Test: Request correlation ID generation
    // Test: Error response standardization
    // Test: Performance metrics collection
}

func TestPublicGatewayRateLimiting(t *testing.T) {
    // Test: IP-based request throttling (1000 req/min)
    // Test: Rate limit header inclusion
    // Test: Overflow protection and queuing
    // Test: Rate limit reset and recovery
}
```

**Admin Gateway Integration Tests**
- **File**: `admin-gateway/internal/gateway/admin_gateway_test.go`
- **Purpose**: Admin gateway authentication and authorization
- **Timeout**: 15 seconds (Integration test)

```go
func TestAdminGatewayAuthentication(t *testing.T) {
    // Test: OAuth2 token validation via Authentik
    // Test: JWT token claims extraction
    // Test: Authentication failure handling
    // Test: Token refresh functionality
    // Test: User context propagation to downstream services
}

func TestAdminGatewayAuthorization(t *testing.T) {
    // Test: OPA policy evaluation for resource access
    // Test: Role-based permission validation
    // Test: Fallback policy enforcement
    // Test: Authorization failure handling
    // Test: Policy cache management
}

func TestAdminGatewayRateLimiting(t *testing.T) {
    // Test: User-based request throttling (100 req/min)
    // Test: Per-user rate limiting state
    // Test: Role-based rate limit variations
    // Test: Rate limiting bypass for system accounts
}
```

**Dapr Middleware Configuration Tests**
- **File**: `infrastructure/dapr/middleware_test.go`
- **Purpose**: Dapr sidecar middleware pipeline validation
- **Timeout**: 15 seconds (Integration test)

```go
func TestDaprMiddlewareConfiguration(t *testing.T) {
    // Test: Middleware pipeline order and execution
    // Test: Service discovery name resolution
    // Test: Rate limiting middleware configuration
    // Test: CORS middleware policy application
    // Test: OAuth2 middleware token processing
    // Test: OPA middleware policy enforcement
}

func TestMiddlewarePipelineIntegration(t *testing.T) {
    // Test: End-to-end middleware pipeline execution
    // Test: Middleware error handling and propagation
    // Test: Performance impact measurement
    // Test: Middleware configuration hot-reload
}
```

**Gateway State Management Tests**
- **File**: `shared/state/gateway_state_test.go`
- **Purpose**: Azure CosmosDB state store integration
- **Timeout**: 15 seconds (Integration test)

```go
func TestGatewayStateStore(t *testing.T) {
    // Test: CosmosDB state store connectivity via Dapr
    // Test: Session state persistence and retrieval
    // Test: Rate limiting counter management
    // Test: Cache state operations and TTL
    // Test: State store error handling and recovery
}

func TestSessionManagement(t *testing.T) {
    // Test: User session creation and validation
    // Test: Session timeout and cleanup
    // Test: Multi-session user handling
    // Test: Session state synchronization
}
```

**Security Integration Tests**
- **File**: `infrastructure/security/security_integration_test.go`
- **Purpose**: End-to-end security pipeline validation
- **Timeout**: 15 seconds (Integration test)

```go
func TestAuthentikIntegration(t *testing.T) {
    // Test: Authentik OAuth2 provider connectivity
    // Test: OIDC discovery endpoint functionality
    // Test: JWT token validation and claims processing
    // Test: User profile retrieval and caching
}

func TestOPAPolicyEnforcement(t *testing.T) {
    // Test: Policy loading and compilation
    // Test: Policy evaluation for different scenarios
    // Test: Policy decision caching
    // Test: Policy update propagation
}

func TestVaultSecretsIntegration(t *testing.T) {
    // Test: Vault secret retrieval for gateway configuration
    // Test: Secret rotation and cache invalidation
    // Test: Certificate management for TLS termination
}
```

**Health and Monitoring Tests**
- **File**: `shared/health/gateway_health_test.go`
- **Purpose**: Gateway health monitoring and dependency validation
- **Timeout**: 5 seconds

```go
func TestGatewayHealthChecks(t *testing.T) {
    // Test: Gateway liveness check responsiveness
    // Test: Gateway readiness check dependency validation
    // Test: Downstream service health aggregation
    // Test: Health check endpoint performance
}

func TestGatewayMonitoring(t *testing.T) {
    // Test: Metrics collection and export
    // Test: Request tracing and correlation
    // Test: Error rate monitoring
    // Test: Performance metrics accuracy
}
```

### Green Phase: Minimal Gateway Implementation

#### Project Structure Creation
```
public-gateway/
├── cmd/
│   └── public-gateway/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── gateway/
│   │   ├── public_gateway.go
│   │   ├── public_gateway_test.go
│   │   ├── routing.go
│   │   └── middleware.go
│   ├── handlers/
│   │   ├── proxy_handler.go
│   │   ├── health_handler.go
│   │   └── error_handler.go
│   ├── middleware/
│   │   ├── cors.go
│   │   ├── rate_limiting.go
│   │   ├── security_headers.go
│   │   └── correlation.go
│   └── health/
│       ├── health.go
│       └── health_test.go
├── go.mod
├── go.sum
├── Dockerfile
└── README.md

admin-gateway/
├── cmd/
│   └── admin-gateway/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── gateway/
│   │   ├── admin_gateway.go
│   │   ├── admin_gateway_test.go
│   │   ├── routing.go
│   │   └── middleware.go
│   ├── handlers/
│   │   ├── proxy_handler.go
│   │   ├── auth_handler.go
│   │   ├── health_handler.go
│   │   └── error_handler.go
│   ├── middleware/
│   │   ├── authentication.go
│   │   ├── authorization.go
│   │   ├── rate_limiting.go
│   │   └── audit_logging.go
│   ├── auth/
│   │   ├── oauth2_client.go
│   │   ├── jwt_validator.go
│   │   └── opa_client.go
│   └── health/
│       ├── health.go
│       └── health_test.go
├── go.mod
├── go.sum
├── Dockerfile
└── README.md

dapr/
├── components/
│   ├── state-store-cosmosdb.yaml
│   ├── pubsub-servicebus.yaml
│   └── secrets-vault.yaml
├── middleware/
│   ├── rate-limiting.yaml
│   ├── cors.yaml
│   ├── oauth2.yaml
│   ├── opa.yaml
│   └── bearer-token.yaml
└── configuration/
    ├── public-gateway-config.yaml
    └── admin-gateway-config.yaml
```

#### Gateway Service Implementation

**Public Gateway Service**
- **File**: `public-gateway/internal/gateway/public_gateway.go`
- **Purpose**: Anonymous access gateway with IP-based rate limiting
- **Features**: Service routing, CORS handling, performance monitoring

**Admin Gateway Service**
- **File**: `admin-gateway/internal/gateway/admin_gateway.go`
- **Purpose**: Authenticated access gateway with role-based authorization
- **Features**: OAuth2 integration, OPA policy enforcement, audit logging

#### Middleware Pipeline Implementation

**Rate Limiting Middleware**
- **File**: `shared/middleware/rate_limiting.go`
- **Purpose**: Request throttling with IP and user-based strategies
- **Features**: Token bucket algorithm, overflow protection, header management

**Authentication Middleware**
- **File**: `admin-gateway/internal/middleware/authentication.go`
- **Purpose**: OAuth2 token validation and user context extraction
- **Features**: JWT validation, claims processing, token refresh handling

**Authorization Middleware**
- **File**: `admin-gateway/internal/middleware/authorization.go`
- **Purpose**: OPA policy-driven access control
- **Features**: Policy evaluation, permission caching, fallback handling

#### Dapr Configuration Implementation

**Middleware Configuration**
- **Files**: `dapr/middleware/*.yaml`
- **Purpose**: Dapr sidecar middleware pipeline configuration
- **Features**: Service-specific middleware chains, policy definitions

**State Store Configuration**
- **File**: `dapr/components/state-store-cosmosdb.yaml`
- **Purpose**: Azure CosmosDB integration for gateway state
- **Features**: Session management, rate limiting state, cache storage

### Refactor Phase: Gateway Optimization

#### Performance Optimization
- Connection pooling and keep-alive optimization
- Response caching strategies and cache invalidation
- Middleware pipeline performance tuning
- Request routing optimization and load balancing

#### Security Enhancement
- Advanced threat detection and mitigation
- Security header optimization and compliance
- Certificate management and TLS configuration
- Audit logging enhancement and compliance

#### Monitoring and Observability Enhancement
- Advanced metrics collection and dashboards
- Distributed tracing integration and correlation
- Error tracking and alerting configuration
- Performance benchmarking and optimization

## Gateway Integration Points

### Service Discovery Integration
- **Dapr Service Invocation**: Automatic service discovery and load balancing
- **Health Check Aggregation**: Downstream service health monitoring
- **Circuit Breaker**: Service failure detection and recovery
- **Retry Logic**: Request retry with exponential backoff

### Authentication Provider Integration
- **Authentik OAuth2**: Identity provider integration and token validation
- **JWT Processing**: Token claims extraction and user context management
- **Session Management**: User session tracking and timeout handling
- **Multi-Factor Authentication**: Enhanced security for admin access

### Authorization Engine Integration
- **Open Policy Agent**: Policy-driven authorization decisions
- **Policy Distribution**: Centralized policy management and updates
- **Permission Caching**: Performance optimization for authorization checks
- **Fallback Policies**: Default security stance for undefined scenarios

### State Management Integration
- **Azure CosmosDB**: Gateway state persistence via Dapr state store
- **Session Storage**: User session tracking and management
- **Rate Limiting Counters**: Request throttling state management
- **Cache Management**: Response caching and TTL management

## Gateway Configuration

### Environment Variables
```bash
# Public Gateway Configuration
PUBLIC_GATEWAY_PORT=${PUBLIC_GATEWAY_PORT}
PUBLIC_GATEWAY_DAPR_HTTP_PORT=${PUBLIC_GATEWAY_DAPR_HTTP_PORT}
PUBLIC_GATEWAY_DAPR_GRPC_PORT=${PUBLIC_GATEWAY_DAPR_GRPC_PORT}

# Admin Gateway Configuration
ADMIN_GATEWAY_PORT=${ADMIN_GATEWAY_PORT}
ADMIN_GATEWAY_DAPR_HTTP_PORT=${ADMIN_GATEWAY_DAPR_HTTP_PORT}
ADMIN_GATEWAY_DAPR_GRPC_PORT=${ADMIN_GATEWAY_DAPR_GRPC_PORT}

# Rate Limiting Configuration
PUBLIC_RATE_LIMIT_RPM=1000
ADMIN_RATE_LIMIT_RPM=100
RATE_LIMIT_BURST_SIZE=50
RATE_LIMIT_WINDOW_SIZE=60s

# Authentication Configuration
AUTHENTIK_ENDPOINT=${AUTHENTIK_ENDPOINT}
AUTHENTIK_CLIENT_ID=${AUTHENTIK_CLIENT_ID}
AUTHENTIK_CLIENT_SECRET=${AUTHENTIK_CLIENT_SECRET}
JWT_SIGNING_KEY=${JWT_SIGNING_KEY}
TOKEN_REFRESH_THRESHOLD=300s

# Authorization Configuration
OPA_ENDPOINT=${OPA_ENDPOINT}
OPA_POLICY_PACKAGE=gateway.authz
POLICY_CACHE_TTL=300s
FALLBACK_POLICY_ALLOW=false

# State Store Configuration
COSMOSDB_ENDPOINT=${COSMOSDB_ENDPOINT}
COSMOSDB_KEY=${COSMOSDB_KEY}
COSMOSDB_DATABASE=gateway_state
STATE_STORE_TTL=3600s

# Security Configuration
SECURITY_HEADERS_ENABLED=true
CORS_ALLOWED_ORIGINS=${CORS_ALLOWED_ORIGINS}
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization,X-Requested-With
TLS_CERT_PATH=${TLS_CERT_PATH}
TLS_KEY_PATH=${TLS_KEY_PATH}

# Monitoring Configuration
METRICS_ENABLED=true
TRACING_ENABLED=true
AUDIT_LOGGING_ENABLED=true
HEALTH_CHECK_INTERVAL=30s
```

### Service Routing Configuration
```bash
# Service Discovery
SERVICES_API_ENDPOINT=services-api
CONTENT_API_ENDPOINT=content-api
SERVICE_INVOCATION_TIMEOUT=30s

# Load Balancing
LOAD_BALANCING_STRATEGY=round_robin
CONNECTION_POOL_SIZE=100
KEEP_ALIVE_TIMEOUT=120s

# Circuit Breaker
CIRCUIT_BREAKER_THRESHOLD=10
CIRCUIT_BREAKER_TIMEOUT=60s
CIRCUIT_BREAKER_RECOVERY_TIME=300s
```

## Agent Success Criteria

### Public Gateway Validation
- IP-based rate limiting operational (1000 req/min)
- Anonymous access to services and content APIs functional
- CORS policies properly configured for public website origins
- Health check endpoints responding within timeout constraints
- Security headers applied to all responses

### Admin Gateway Validation
- OAuth2 authentication integration with Authentik operational
- OPA policy enforcement for role-based access control
- User-based rate limiting functional (100 req/min)
- Audit logging to Grafana Cloud Loki successful
- JWT token validation and refresh functionality operational

### Dapr Middleware Pipeline Validation
- All middleware components properly configured and operational
- Middleware execution order correct and performance optimized
- Service discovery and routing functional across all endpoints
- State store integration for session and rate limiting operational
- Error handling and fallback mechanisms tested and validated

### Integration Test Compliance
- End-to-end authentication and authorization workflows functional
- Gateway routing to services and content APIs validated
- Rate limiting enforcement tested under load conditions
- Health check aggregation and monitoring operational
- Performance benchmarks established and documented

### Security Compliance Validation
- All security headers properly applied and validated
- Authentication bypass attempts properly blocked
- Authorization policy enforcement comprehensive
- Audit trail completeness for compliance requirements
- TLS termination and certificate management operational

## Coordination with Parallel Agents

### Shared Infrastructure Dependencies
- **Dapr Control Plane**: Service discovery and middleware coordination
- **Azure CosmosDB**: Gateway state management and session storage
- **Authentik Identity Provider**: Authentication and user management
- **Open Policy Agent**: Authorization policy enforcement
- **Grafana Stack**: Audit logging and observability integration

### Services Domain Integration
- **Service Discovery**: Automatic routing to services-api endpoints
- **Health Check Integration**: Gateway health depends on services-api health
- **Error Response Consistency**: Standardized error formats across domains
- **Correlation Tracking**: Request correlation propagation to services

### Content Domain Integration
- **Content API Routing**: Gateway routing for content upload and delivery
- **Large File Handling**: Streaming support for content operations
- **Health Check Integration**: Gateway health depends on content-api health
- **Storage Backend Coordination**: State management coordination with content storage

### End-to-End Validation Preparation
- **Complete Gateway Implementation**: All endpoints and middleware operational
- **Authentication/Authorization Integration**: Full security pipeline functional
- **Performance Baselines**: Gateway performance metrics established
- **Audit Trail Completeness**: Full audit event coverage
- **Monitoring Integration**: Gateway metrics available in observability stack