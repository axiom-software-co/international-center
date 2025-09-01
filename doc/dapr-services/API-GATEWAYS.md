# API Gateways Specification

## Architecture Overview

**Service Mesh**: Dapr-centric microservices architecture with HTTP middleware pipelines
**Gateway Pattern**: Public (anonymous) and Admin (authenticated) API gateways
**Cross-Cutting Concerns**: CORS, rate limiting, security headers, route validation, observability
**Environment Configuration**: Container-only environment variables with no hardcoded values
**Observability**: Grafana Cloud telemetry collection via Dapr middleware integration

## Gateway Service Overview

### Public API Gateway
**Service Name**: `public-gateway`
**Dapr App ID**: `public-gateway`  
**Access Pattern**: Anonymous public access with IP-based rate limiting
**Primary Function**: Reverse proxy for public APIs with standard cross-cutting concerns
**Rate Limiting**: 1000 requests/minute per IP address
**Security Model**: Anonymous access with security headers and route validation

### Admin API Gateway  
**Service Name**: `admin-gateway`
**Dapr App ID**: `admin-gateway`
**Access Pattern**: Authenticated admin users with role-based access control
**Primary Function**: Reverse proxy for admin APIs with compliance and enhanced security
**Rate Limiting**: 100 requests/minute per authenticated user
**Security Model**: OAuth2 + JWT authentication with OPA policy enforcement
**Compliance**: Audit event sourcing directly to Grafana Cloud Loki for regulatory requirements

## Dapr HTTP Middleware Pipeline Configurations

### Public Gateway Middleware Chain
```yaml
middlewares:
  - name: "cors"
  - name: "rate-limit"
  - name: "security-headers"
  - name: "route-checker"
  - name: "route-alias"
  - name: "telemetry"
  - name: "anonymous-auth"
```

### Admin Gateway Middleware Chain  
```yaml
middlewares:
  - name: "cors"
  - name: "bearer"
  - name: "oauth2"
  - name: "opa-policies"
  - name: "rate-limit"
  - name: "security-headers"
  - name: "route-checker"
  - name: "route-alias"
  - name: "audit-logging"
  - name: "telemetry"
```

## Middleware Configuration Specifications

### CORS Middleware

#### Public Gateway CORS
**Purpose**: Handle cross-origin requests from public website
**Configuration**: Environment-driven allowed origins for public access
```yaml
cors:
  allowedOrigins: "${PUBLIC_WEBSITE_ORIGINS}"
  allowedMethods: ["GET", "POST", "OPTIONS"]
  allowedHeaders: ["Content-Type", "X-Correlation-ID", "X-Request-ID"]
  allowCredentials: false
  maxAge: 3600
```

#### Admin Gateway CORS
**Purpose**: Handle cross-origin requests from admin interfaces
**Configuration**: Restricted origins for admin applications with credential support
```yaml
cors:
  allowedOrigins: "${ADMIN_WEBSITE_ORIGINS}"
  allowedMethods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
  allowedHeaders: ["Content-Type", "Authorization", "X-Correlation-ID", "X-Request-ID"]
  allowCredentials: true
  maxAge: 3600
```

### Authentication Middleware (Admin Gateway Only)

#### Bearer Token Middleware
**Purpose**: Extract and validate JWT bearer tokens from Authorization header
**Configuration**: Token validation with Authentik integration
```yaml
bearer:
  authorizationHeader: "Authorization"
  tokenPattern: "Bearer {token}"
  required: true
  errorOnMissing: true
```

#### OAuth2 Middleware
**Purpose**: Authentik OAuth2 integration and JWT token validation
**Configuration**: JWT validation with issuer, audience, and JWKS validation
```yaml
oauth2:
  provider: "authentik"
  issuer: "${AUTHENTIK_ISSUER_URL}"
  audience: "${AUTHENTIK_ADMIN_AUDIENCE}"
  jwksUrl: "${AUTHENTIK_JWKS_URL}"
  tokenValidation:
    validateExpiry: true
    validateIssuer: true
    validateAudience: true
    clockSkew: "5m"
  userContext:
    userIdClaim: "sub"
    rolesClaim: "groups"
    emailClaim: "email"
```

#### OPA Policies Middleware
**Purpose**: Role-based access control via Open Policy Agent
**Configuration**: Policy enforcement for admin operations with fallback deny policy
```yaml
opaPolicies:
  endpoint: "${OPA_POLICY_ENDPOINT}"
  policies:
    - name: "admin-services-policy"
      path: "/admin/api/v1/services"
      requiredRoles: ["admin", "services-manager"]
    - name: "admin-content-policy"
      path: "/admin/api/v1/content"
      requiredRoles: ["admin", "content-manager"]
    - name: "audit-read-policy"
      path: "/admin/api/v1/*/audit"
      requiredRoles: ["admin", "audit-viewer"]
  fallbackPolicy: "deny"
  contextInclusion:
    - "user.id"
    - "user.roles"
    - "request.method"
    - "request.path"
```

#### Anonymous Authentication Middleware (Public Gateway Only)
**Purpose**: Handle anonymous access patterns via Authentik integration
**Configuration**: Anonymous user context creation for public access
```yaml
anonymousAuth:
  enabled: true
  userContext:
    userId: "anonymous"
    roles: ["public"]
  authentikIntegration:
    endpoint: "${AUTHENTIK_ENDPOINT}"
    realm: "public"
```

### Rate Limiting Middleware

#### Public Gateway Rate Limiting
**Purpose**: IP-based rate limiting for anonymous users
**Configuration**: 1000 requests/minute per IP address with Redis storage
```yaml
rateLimit:
  type: "ip-based"
  requests: 1000
  duration: "1m"
  storage: "redis-rate-limit"
  headers: true
  keyPattern: "public-gateway:{client-ip}"
```

#### Admin Gateway Rate Limiting
**Purpose**: User-based rate limiting for admin operations
**Configuration**: 100 requests/minute per authenticated user with Redis storage
```yaml
rateLimit:
  type: "user-based"
  requests: 100
  duration: "1m"
  storage: "redis-rate-limit"
  headers: true
  keyPattern: "admin-gateway:{user-id}"
  userIdExtraction: "jwt.sub"
```

### Security Headers Middleware

#### Public Gateway Security Headers
**Purpose**: Apply security headers for public content access
**Configuration**: Standard security header enforcement for anonymous access
```yaml
securityHeaders:
  contentTypeOptions: "nosniff"
  frameOptions: "DENY"
  contentSecurityPolicy: "default-src 'self'"
  strictTransportSecurity: "max-age=31536000"
  referrerPolicy: "strict-origin-when-cross-origin"
```

#### Admin Gateway Security Headers
**Purpose**: Enhanced security headers for admin interfaces
**Configuration**: Strict security policy enforcement for authenticated admin operations
```yaml
securityHeaders:
  contentTypeOptions: "nosniff"
  frameOptions: "DENY"
  contentSecurityPolicy: "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'"
  strictTransportSecurity: "max-age=31536000; includeSubDomains"
  referrerPolicy: "strict-origin"
  permissionsPolicy: "geolocation=(), microphone=(), camera=()"
```

### Route Validation and Aliasing

#### Public Gateway Route Checker
**Purpose**: Validate incoming routes against allowed public API patterns
**Configuration**: Public API route validation for anonymous access
```yaml
routeChecker:
  allowedRoutes:
    - "/api/v1/services"
    - "/api/v1/services/{id}"
    - "/api/v1/services/slug/{slug}"
    - "/api/v1/services/featured"
    - "/api/v1/services/categories"
    - "/api/v1/services/categories/{id}/services"
    - "/api/v1/services/search"
    - "/api/v1/content"
    - "/api/v1/content/{id}"
    - "/api/v1/content/{id}/download"
    - "/api/v1/content/{id}/preview"
    - "/health"
    - "/health/ready"
```

#### Admin Gateway Route Checker
**Purpose**: Validate admin routes against allowed administrative patterns
**Configuration**: Admin API route validation for authenticated access
```yaml
routeChecker:
  allowedRoutes:
    - "/admin/api/v1/services"
    - "/admin/api/v1/services/{id}"
    - "/admin/api/v1/services/{id}/publish"
    - "/admin/api/v1/services/{id}/archive"
    - "/admin/api/v1/services/{id}/audit"
    - "/admin/api/v1/services/categories"
    - "/admin/api/v1/services/categories/{id}"
    - "/admin/api/v1/services/categories/{id}/audit"
    - "/admin/api/v1/services/featured-categories"
    - "/admin/api/v1/content"
    - "/admin/api/v1/content/{id}"
    - "/admin/api/v1/content/upload"
    - "/admin/api/v1/content/{id}/reprocess"
    - "/admin/api/v1/content/{id}/status"
    - "/admin/api/v1/content/{id}/audit"
    - "/admin/api/v1/content/processing-queue"
    - "/admin/api/v1/content/analytics"
    - "/health"
    - "/health/ready"
```

#### Route Alias Middleware

**Public Gateway Route Aliases**: API versioning and legacy support
```yaml
routeAlias:
  aliases:
    - from: "/services"
      to: "/api/v1/services"
    - from: "/content"
      to: "/api/v1/content"
```

**Admin Gateway Route Aliases**: Admin route aliasing and versioning
```yaml
routeAlias:
  aliases:
    - from: "/admin/services"
      to: "/admin/api/v1/services"
    - from: "/admin/content"
      to: "/admin/api/v1/content"
```

### Audit Logging Middleware (Admin Gateway Only)

**Purpose**: Compliance audit logging directly to Grafana Cloud Loki
**Configuration**: Direct audit event publishing for immutable compliance trail
```yaml
auditLogging:
  enabled: true
  destination: "grafana-cloud-loki"
  lokiEndpoint: "${GRAFANA_CLOUD_LOKI_ENDPOINT}"
  apiKey: "${GRAFANA_CLOUD_API_KEY}"
  includeRequestBody: true
  includeResponseBody: false
  sensitiveFields:
    - "password"
    - "token"
    - "secret"
  auditContext:
    - "user.id"
    - "user.email"
    - "user.roles"
    - "request.method"
    - "request.path"
    - "correlation.id"
    - "trace.id"
  lokiLabels:
    job: "admin-audit"
    environment: "${ENVIRONMENT}"
    service: "admin-gateway"
```

### Telemetry Middleware

#### Public Gateway Telemetry
**Purpose**: Grafana observability integration for anonymous access patterns
**Configuration**: Distributed tracing and metrics collection for public APIs
```yaml
telemetry:
  tracing:
    enabled: true
    samplingRate: 1.0
    headers:
      - "X-Trace-ID"
      - "X-Span-ID"
      - "X-Request-ID"
  metrics:
    enabled: true
    namespace: "public_gateway"
    labels:
      - "method"
      - "route"
      - "status_code"
```

#### Admin Gateway Telemetry
**Purpose**: Enhanced observability integration for admin operations
**Configuration**: Enhanced telemetry with user context and audit metrics
```yaml
telemetry:
  tracing:
    enabled: true
    samplingRate: 1.0
    headers:
      - "X-Trace-ID"
      - "X-Span-ID"
      - "X-Request-ID"
      - "X-User-ID"
  metrics:
    enabled: true
    namespace: "admin_gateway"
    labels:
      - "method"
      - "route"
      - "status_code"
      - "user_id"
      - "user_role"
  auditMetrics:
    enabled: true
    destination: "grafana-cloud"
```

## Backend Service Routing Patterns

### Services API Routing

#### Public Gateway Services API
**Target Service**: `services-api`
**Base Path**: `/api/v1/services`
**Service Invocation**: `services-api/api/v1/services`
**Health Check**: `services-api/health`
**Access Pattern**: Anonymous access to published services

#### Admin Gateway Services API
**Target Service**: `services-api`
**Base Path**: `/admin/api/v1/services`
**Service Invocation**: `services-api/admin/api/v1/services`
**Health Check**: `services-api/health`
**Access Pattern**: Authenticated admin access with role-based permissions
**Audit Context**: Full request/response logging to Grafana Cloud

### Content API Routing

#### Public Gateway Content API
**Target Service**: `content-api`
**Base Path**: `/api/v1/content`
**Service Invocation**: `content-api/api/v1/content`
**Health Check**: `content-api/health`
**Access Pattern**: Anonymous access to approved content

#### Admin Gateway Content API
**Target Service**: `content-api`
**Base Path**: `/admin/api/v1/content`
**Service Invocation**: `content-api/admin/api/v1/content`
**Health Check**: `content-api/health`
**Access Pattern**: Authenticated admin access with role-based permissions
**Audit Context**: Full request/response logging to Grafana Cloud

## Environment Configuration

### Public Gateway Environment Variables
```bash
# Network Configuration
PUBLIC_GATEWAY_PORT=${PUBLIC_GATEWAY_PORT}
PUBLIC_GATEWAY_HOST=${PUBLIC_GATEWAY_HOST}

# CORS Configuration
PUBLIC_WEBSITE_ORIGINS=${PUBLIC_WEBSITE_ORIGINS}

# Authentik Integration
AUTHENTIK_ENDPOINT=${AUTHENTIK_ENDPOINT}
AUTHENTIK_PUBLIC_REALM=${AUTHENTIK_PUBLIC_REALM}

# Rate Limiting
RATE_LIMIT_REDIS_ENDPOINT=${RATE_LIMIT_REDIS_ENDPOINT}

# Backend Services
SERVICES_API_ENDPOINT=${SERVICES_API_ENDPOINT}
CONTENT_API_ENDPOINT=${CONTENT_API_ENDPOINT}

# Observability
GRAFANA_ENDPOINT=${GRAFANA_ENDPOINT}
TELEMETRY_SAMPLING_RATE=${TELEMETRY_SAMPLING_RATE}
```

### Admin Gateway Environment Variables
```bash
# Network Configuration
ADMIN_GATEWAY_PORT=${ADMIN_GATEWAY_PORT}
ADMIN_GATEWAY_HOST=${ADMIN_GATEWAY_HOST}

# CORS Configuration
ADMIN_WEBSITE_ORIGINS=${ADMIN_WEBSITE_ORIGINS}

# Authentik OAuth2 Integration
AUTHENTIK_ISSUER_URL=${AUTHENTIK_ISSUER_URL}
AUTHENTIK_ADMIN_AUDIENCE=${AUTHENTIK_ADMIN_AUDIENCE}
AUTHENTIK_JWKS_URL=${AUTHENTIK_JWKS_URL}
AUTHENTIK_CLIENT_ID=${AUTHENTIK_ADMIN_CLIENT_ID}

# OPA Policy Engine
OPA_POLICY_ENDPOINT=${OPA_POLICY_ENDPOINT}

# Rate Limiting
RATE_LIMIT_REDIS_ENDPOINT=${RATE_LIMIT_REDIS_ENDPOINT}

# Backend Services
SERVICES_API_ENDPOINT=${SERVICES_API_ENDPOINT}
CONTENT_API_ENDPOINT=${CONTENT_API_ENDPOINT}

# Observability
GRAFANA_CLOUD_ENDPOINT=${GRAFANA_CLOUD_ENDPOINT}
GRAFANA_CLOUD_API_KEY=${GRAFANA_CLOUD_API_KEY}
TELEMETRY_SAMPLING_RATE=${TELEMETRY_SAMPLING_RATE}

# Audit Logging - Direct to Grafana Cloud Loki
GRAFANA_CLOUD_LOKI_ENDPOINT=${GRAFANA_CLOUD_LOKI_ENDPOINT}
GRAFANA_CLOUD_LOKI_API_KEY=${GRAFANA_CLOUD_API_KEY}
```

### Local Development Environment Overrides
```bash
# Public Gateway Local Development
BLOB_STORAGE_ENDPOINT=${AZURITE_BLOB_ENDPOINT}
COSMOSDB_ENDPOINT=${COSMOSDB_EMULATOR_ENDPOINT}
SERVICE_BUS_ENDPOINT=${SERVICE_BUS_EMULATOR_ENDPOINT}
VAULT_ENDPOINT=${VAULT_LOCAL_ENDPOINT}

# Admin Gateway Local Development
AUTHENTIK_LOCAL_ENDPOINT=${AUTHENTIK_LOCAL_ENDPOINT}
OPA_LOCAL_ENDPOINT=${OPA_LOCAL_ENDPOINT}
GRAFANA_LOCAL_ENDPOINT=${GRAFANA_LOCAL_ENDPOINT}
SERVICE_BUS_LOCAL_ENDPOINT=${SERVICE_BUS_LOCAL_ENDPOINT}
```

## Health Check Endpoints

### Public Gateway Health Checks

#### Gateway Health Check
**Endpoint**: `/health`
**Response**:
```json
{
  "status": "healthy",
  "timestamp": "timestamp",
  "service": "public-gateway",
  "version": "string",
  "dapr_app_id": "public-gateway",
  "middleware_status": {
    "cors": "healthy",
    "rate_limit": "healthy",
    "security_headers": "healthy",
    "telemetry": "healthy"
  }
}
```

#### Gateway Readiness Check
**Endpoint**: `/health/ready`
**Response**:
```json
{
  "status": "ready",
  "checks": {
    "dapr_sidecar": "healthy",
    "services_api": "healthy",
    "content_api": "healthy",
    "authentik": "healthy",
    "rate_limit_store": "healthy",
    "grafana_telemetry": "healthy"
  },
  "environment": "${ENVIRONMENT}",
  "dapr_app_id": "public-gateway"
}
```

### Admin Gateway Health Checks

#### Gateway Health Check
**Endpoint**: `/health`
**Response**:
```json
{
  "status": "healthy",
  "timestamp": "timestamp",
  "service": "admin-gateway",
  "version": "string",
  "dapr_app_id": "admin-gateway",
  "middleware_status": {
    "oauth2": "healthy",
    "opa_policies": "healthy",
    "rate_limit": "healthy",
    "audit_logging": "healthy",
    "telemetry": "healthy"
  }
}
```

#### Gateway Readiness Check
**Endpoint**: `/health/ready`
**Response**:
```json
{
  "status": "ready",
  "checks": {
    "dapr_sidecar": "healthy",
    "services_api": "healthy",
    "content_api": "healthy",
    "authentik_oauth2": "healthy",
    "opa_policies": "healthy",
    "rate_limit_store": "healthy",
    "grafana_cloud": "healthy",
    "grafana_cloud_loki": "healthy"
  },
  "environment": "${ENVIRONMENT}",
  "dapr_app_id": "admin-gateway"
}
```

## Error Handling and Response Patterns

### Public Gateway Error Response
```json
{
  "error": {
    "code": "GATEWAY_ERROR",
    "message": "Gateway processing error",
    "details": "Middleware or routing error details",
    "correlation_id": "uuid",
    "trace_id": "trace_id",
    "span_id": "span_id",
    "service_name": "public-gateway",
    "middleware": "rate-limit|cors|security|routing",
    "timestamp": "timestamp"
  }
}
```

### Admin Gateway Error Response
```json
{
  "error": {
    "code": "ADMIN_GATEWAY_ERROR",
    "message": "Admin gateway processing error",
    "details": "Authentication, authorization, or routing error",
    "correlation_id": "uuid",
    "trace_id": "trace_id",
    "span_id": "span_id",
    "service_name": "admin-gateway",
    "middleware": "oauth2|opa|rate-limit|audit|routing",
    "user_context": {
      "user_id": "user_id",
      "roles": ["role1", "role2"]
    },
    "timestamp": "timestamp"
  }
}
```

### Common HTTP Status Codes
- `400`: Bad Request - Route not allowed by route checker middleware
- `401`: Unauthorized - Invalid or missing JWT token (admin gateway)
- `403`: Forbidden - Valid token but insufficient permissions via OPA (admin gateway)
- `429`: Too Many Requests - Rate limit exceeded (per IP for public, per user for admin)
- `500`: Internal Server Error - Gateway or upstream service error
- `503`: Service Unavailable - Backend service unhealthy

## Security Configuration

### Public Gateway Security Model
- **Authentication**: Anonymous access with security headers
- **Rate Limiting**: IP-based rate limiting for abuse prevention  
- **Route Validation**: Request validation via route checker middleware
- **Security Headers**: XSS/CSRF protection via security headers
- **Content Security Policy**: Restrictive CSP for public content access

### Admin Gateway Security Model

#### OAuth2 Authentication Flow
1. Client sends request with `Authorization: Bearer <jwt>`
2. Bearer middleware extracts token
3. OAuth2 middleware validates with Authentik
4. User context extracted from JWT claims
5. Request proceeds to OPA policy evaluation

#### Role-Based Access Control
**Policy Engine**: Open Policy Agent via Dapr middleware
**Policy Storage**: HashiCorp Vault via Dapr secret store
**Policy Evaluation**: Real-time per request evaluation
**Fallback Policy**: Explicit deny for compliance requirements

#### Audit Compliance Requirements
- All admin operations logged to Grafana Cloud Loki
- Event sourcing via Dapr pub/sub to `admin-audit-events` topic
- Immutable audit trail for regulatory compliance
- User attribution and correlation tracking throughout request lifecycle
- Request/response body logging for sensitive administrative operations

## Performance Configuration

### Connection Management
- **HTTP/2 Support**: Both gateways support HTTP/2 for client connections
- **Connection Pooling**: Backend service invocation uses connection pooling
- **Dapr Service Mesh**: Optimized service-to-service communication
- **TLS Termination**: TLS termination at gateway level for security

### Caching Strategy

#### Public Gateway Caching
- **Gateway Level**: No caching at gateway level (handled by Azure Blob Storage)
- **Cache Headers**: Cache-control header passthrough from backend services
- **ETag Support**: Conditional requests supported for content optimization
- **Compression**: Gzip compression enabled for text content

#### Admin Gateway Caching
- **Admin Operations**: No caching for admin operations (compliance requirement)
- **Real-Time Data**: Real-time data requirements for administrative functions
- **Conditional Requests**: ETag support for conditional requests where appropriate
- **Compression**: Gzip compression for large admin payloads

## Monitoring and Alerting

### Grafana Cloud Integration

#### Public Gateway Monitoring
- Request/response metrics collection and analysis
- Error rate monitoring and alerting for anonymous access patterns
- Response time percentiles tracking for public API performance
- Rate limiting metrics and alerts for abuse detection

#### Admin Gateway Monitoring
- User activity monitoring and anomaly detection
- Performance metrics for admin operations with user attribution
- Security event monitoring and alerting for compliance
- Failed authentication attempt tracking and alerting

### Compliance Monitoring (Admin Gateway)
- Failed authentication attempt tracking and alerting
- Permission violation monitoring via OPA policy failures
- Rate limit breach alerting for user-based limits
- Audit log delivery verification to Grafana Cloud Loki

### Log Aggregation Patterns
- **Structured Logging**: Grafana Loki integration for both gateways
- **Security Events**: Dedicated log streams for security-related events
- **Performance Metrics**: Grafana Mimir integration for performance tracking
- **Distributed Tracing**: Grafana Tempo integration for request correlation
- **Audit Events**: Compliance log retention for admin gateway operations