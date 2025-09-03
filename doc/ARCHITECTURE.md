# Development Architecture Reference

## Architectural Philosophy

### Core Principles
- **Cohesion over Coupling**: Organize code by feature rather than technical layer
- **Test-Driven Development**: Tests drive and validate architectural design
- **Contract-First Design**: Interfaces define system boundaries and behavior
- **Dependency Inversion**: Interfaces for variable concerns, concrete types for stable concerns
- **Warnings as Errors**: Enforce code quality through compiler validation

### Design Methodology
- **Red-Green-Refactor Cycle**: TDD phases that validate architectural decisions
- **Vertical Slice Architecture**: Feature-oriented organization with shared kernels
- **Property-Based Testing**: Behavior verification across input domains

## Layer Architecture

### Hierarchical Structure
```
Frontend Layer (Highest)
├── Gateway Layer
├── Application Layer  
├── Domain Layer
└── Infrastructure Layer (Lowest)
```

### Architectural Rules
- Lower layers never depend on nor are aware of higher layers
- Each layer has distinct responsibilities and boundaries
- Cross-cutting concerns handled at appropriate abstraction levels

## Domain Design Patterns

### Vertical Slice Architecture
- **Feature Organization**: Code organized by business capability
- **Slice Boundaries**: Complete feature implementation within bounded context  
- **Minimal Coupling**: Slices interact through well-defined interfaces
- **Independent Evolution**: Slices can evolve independently

### Shared Kernel Pattern
- **Common Domain Logic**: Shared business concepts across bounded contexts
- **Controlled Coupling**: Explicit agreement on shared model evolution
- **Version Management**: Coordinated changes across dependent contexts
- **Minimal Surface Area**: Only essential concepts shared

### Service Patterns
- **Handler Pattern**: Request/response coordination and orchestration
- **Service Pattern**: Business logic coordination and workflow management  
- **Repository Pattern**: Data access abstraction and persistence concerns
- **Health Check Pattern**: Service readiness and liveness validation

## Testing Architecture

### Test Categories and Timeouts
- **Unit Tests**: Domain logic isolation with dependency injection (5 seconds timeout)
- **Integration Tests**: Full environment testing (15 seconds timeout)  
- **End-to-End Tests**: Complete system workflow validation (30 seconds timeout)

### Testing Principles
- **Arrange-Act-Assert**: Clear test structure and intention
- **Contract-First Testing**: Interface compliance over implementation details
- **Property-Based Testing**: Behavior verification across input ranges
- **Timeout Constraints**: Fast failure for problematic scenarios

### Test Isolation Strategies
- **Unit Test Isolation**: Dependency injection with interface boundaries
- **Integration Test Environment**: Full infrastructure stack availability
- **Test Framework Usage**: Structured testing over command-line tools

## Development Workflow

### Test-Driven Development Cycles
- **Red Phase**: Write failing test that defines desired behavior
- **Green Phase**: Implement minimal code to satisfy test requirements  
- **Refactor Phase**: Improve code quality without changing behavior

### Contract-First Design
- **Interface Definition**: Public contracts defined before implementation
- **Behavior Specification**: Expected interactions and outcomes documented
- **Implementation Validation**: Tests verify contract compliance

### Environment Management
- **Container-Driven Configuration**: All settings via environment variables
- **Infrastructure Parity**: Local development mirrors production environment
- **Migration Strategies**: Environment-specific database evolution approaches

## Infrastructure Patterns

### Configuration Management
- **Environment Variables**: Container-only configuration approach
- **No Hardcoding**: Network settings always from environment
- **No Fallback Configuration**: Explicit configuration requirements

### Migration Strategies
- **Development Approach**: Aggressive migration with easy rollback
- **Staging Approach**: Careful validation with supported rollback  
- **Production Approach**: Conservative validation with manual approval

### Dependency Management
- **Singleton Dependencies**: Singletons only depend on other singletons
- **Interface Injection**: Variable concerns abstracted through interfaces
- **Concrete Types**: Stable concerns use concrete implementations

## Observability Patterns

### Structured Logging
- **Key Information**: User ID, Correlation ID, Request URL, Application Version
- **Log Levels**: Debug, Information, Warning, Error, Critical
- **Developer Focus**: Logs designed for development and troubleshooting
- **Delivery Tolerance**: Acceptable to lose some log entries

### Distributed Tracing  
- **Request Correlation**: Track requests across service boundaries
- **Performance Monitoring**: Identify bottlenecks and optimization opportunities
- **Debugging Support**: Trace execution flow for issue resolution

### Health Monitoring
- **Readiness Checks**: Service capability to handle requests
- **Liveness Checks**: Service operational status verification
- **Dependency Validation**: Downstream service availability

## Security Architecture

### Defense Principles
- **Layered Security**: Multiple security controls at different levels
- **Fallback Policies**: Default security stance when specific policies absent
- **Least Privilege**: Minimal required access permissions
- **Audit Compliance**: Immutable event logging for regulatory requirements

### Identity and Access Management
- **Authentication**: Identity verification and validation
- **Authorization**: Permission-based access control
- **Policy Enforcement**: Rule-based access decisions

## Anti-Patterns to Avoid

### Architectural Anti-Patterns
- **Result Pattern**: Go has built-in error handling mechanisms
- **Infrastructure Abstraction**: Integration tests use actual infrastructure components
- **Component Replacement**: Address root infrastructure issues directly

### Development Anti-Patterns
- **Script Files**: Scripts violate proper application architecture
- **Hardcoded Configuration**: All settings must come from environment
- **Test Shortcuts**: CLI tools instead of proper test frameworks

## Quality Assurance

### Code Quality
- **Warnings as Errors**: Compiler warnings treated as build failures
- **Static Analysis**: Automated code quality validation
- **Technical Debt Awareness**: Continuous attention to code maintainability

### Architecture Validation
- **Contract Compliance**: Interface adherence verification
- **Dependency Direction**: Layer dependency rule enforcement
- **Coupling Measurement**: Monitor and minimize inter-component coupling

## Version Control Strategy

### Trunk-Based Development
- **Single Main Branch**: All development flows through main branch
- **Short-Lived Features**: Minimal branch lifetime and scope
- **Continuous Integration**: Automated validation of all changes
- **Feature Flags**: Runtime behavior control instead of branching

# End-to-End Deployment Pipeline - Multi-Environment Orchestration

## Architecture Overview

**Development Environment**: Podman Compose with containerd runtime for complete local development
**Staging Environment**: Azure Container Apps with Dapr integration via Pulumi orchestration (careful migration approach)
**Production Environment**: Azure Container Apps with Dapr integration via Pulumi orchestration (conservative migration approach)
**Version Control**: Trunk-based development with GitHub and GitHub Container Registry
**Infrastructure as Code**: Single Pulumi Go SDK program orchestrating multi-cloud deployment
**State Management**: Azure Blob Storage for Pulumi state backend with environment isolation
**Migration Management**: Go migration tools built and executed via Pulumi Command resources
**Observability**: Grafana Cloud integration for staging/production, local Grafana stack for development
**Security**: HashiCorp Vault Cloud for production secrets, local Vault for development

## Environment-Specific Deployment Strategies

### Development Environment (Aggressive Migration Approach)
- **Runtime**: Podman Compose with containerd
- **Purpose**: Complete local development environment with production parity
- **Migration Strategy**: Aggressive - always migrate to latest version with automatic rollback via environment reset
- **Rollback**: Easy - destroy and recreate database and volumes
- **Safety Checks**: Minimal validation for fast iteration
- **Automation**: Full automation via Podman Compose init containers
- **Timeout**: 30 seconds maximum for migration operations

### Staging Environment (Careful Migration Approach)
- **Runtime**: Azure Container Apps with Dapr service mesh
- **Purpose**: Production-like validation environment for integration testing
- **Migration Strategy**: Careful - migrate with validation and confirmation prompts
- **Rollback**: Supported with automated confirmation via Pulumi orchestration
- **Safety Checks**: Moderate validation of schema integrity and compliance
- **Automation**: Pulumi orchestrated via GitHub Actions on develop branch
- **Backup**: Incremental backup before migration operations
- **Timeout**: 300 seconds maximum for migration operations

### Production Environment (Conservative Migration Approach)
- **Runtime**: Azure Container Apps with Dapr service mesh and compliance controls
- **Purpose**: Production deployment with extensive validation and audit requirements
- **Migration Strategy**: Conservative - extensive validation before any schema changes
- **Rollback**: Manual approval required for rollback operations via GitHub environment
- **Safety Checks**: Full validation, backup verification, and compliance validation
- **Automation**: Pulumi orchestrated with human approval gates on main branch
- **Backup**: Full backup with verification before any schema changes
- **Timeout**: 900 seconds maximum for migration operations

## Development Environment Configuration

This integrated approach ensures:
- **Single Source of Truth**: All infrastructure, migrations, and applications defined in one program
- **Proper Dependencies**: Pulumi manages resource creation order and dependencies across environments
- **Environment-Specific Behavior**: Configuration-driven approach for development/staging/production
- **Compliance and Audit**: Built-in audit logging and compliance validation for all environments
- **State Consistency**: All changes tracked in single Pulumi state file per environment
- **Migration Integration**: Database migrations executed as Pulumi Command resources with proper orchestration

## Compliance and Audit Requirements

### Multi-Environment Audit Trail
- **Development**: Basic logging to local Grafana stack for debugging and development workflow
- **Staging**: Enhanced audit logging to Grafana Cloud Loki for validation and integration testing
- **Production**: Full compliance audit logging to Grafana Cloud Loki with immutable event sourcing
- **Cross-Environment**: All deployment operations logged with correlation IDs and environment context

### Schema Compliance Validation
- **Development**: Minimal validation for fast iteration with aggressive rollback via environment reset
- **Staging**: Moderate schema validation ensuring database matches specification files exactly
- **Production**: Extensive schema validation with full backup verification and compliance checking
- **No Audit Tables**: Compliance requirement enforced across all environments - no audit data in databases

### Migration Backup and Recovery Strategy
- **Development**: No backup required - aggressive rollback via complete environment recreation
- **Staging**: Incremental backup before migration operations with automated rollback capability
- **Production**: Full backup with verification before any schema changes, manual approval for rollbacks
- **Cross-Environment**: All backup operations logged to Grafana Cloud for audit trail compliance
