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