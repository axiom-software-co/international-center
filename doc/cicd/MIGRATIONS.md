# Database Migrations and Schema Management

## Migration Architecture Overview

**Migration Tool**: golang-migrate with custom Go migrator wrapper
**Execution Context**: Integrated with Pulumi orchestration in CI/CD pipeline
**Environment Strategies**: Aggressive (development), Careful (staging), Conservative (production)
**Schema Compliance**: Database schemas must exactly match specification files

## Migration Tool Integration

### Go Migrator Implementation
```go
// infrastructure/migrations/migrator.go
type Migrator struct {
    DatabaseURL string
    Environment string
    Logger      *slog.Logger
}

func (m *Migrator) RunMigrations(ctx context.Context) error {
    // Environment-specific migration execution
    switch m.Environment {
    case "development":
        return m.aggressiveMigration(ctx)
    case "staging":
        return m.carefulMigration(ctx)
    case "production":
        return m.conservativeMigration(ctx)
    }
}
```

### Migration Tool Building in CI/CD
Migration tools are built during deployment process via Pulumi orchestration:

```bash
# Built as part of CI/CD workflow
cd ../migrations
go build -o migrate ./cmd/migrate/
chmod +x migrate
```

### Environment-Specific Migration Strategies

#### Development Environment (Aggressive Approach)
- **Execution**: Always migrate to latest version
- **Rollback**: Easy - destroy and recreate database
- **Safety Checks**: Minimal validation
- **Automation**: Full automation via Podman Compose init containers
- **Timeout**: 30 seconds maximum
- **Error Handling**: Fail fast, recreate if needed

#### Staging Environment (Careful Approach)
- **Execution**: Migrate with validation and confirmation
- **Rollback**: Supported with confirmation prompts
- **Safety Checks**: Moderate validation of schema integrity
- **Automation**: Pulumi orchestrated via GitHub Actions
- **Backup**: Incremental backup before migration
- **Timeout**: 300 seconds maximum

#### Production Environment (Conservative Approach)
- **Execution**: Extensive validation before migration
- **Rollback**: Manual approval required for rollback operations
- **Safety Checks**: Full validation and backup verification
- **Automation**: Pulumi orchestrated with human approval gates
- **Backup**: Full backup before any schema changes
- **Timeout**: 900 seconds maximum

## Migration Execution Patterns

### Pulumi Integration Pattern
Migrations are executed as Pulumi Command resources with proper dependencies:

```yaml
# Migration execution within Pulumi deployment
migration_command:
  depends_on: [database_infrastructure]
  command: "./migrate --environment=${environment} --database-url=${database_url}"
  environment_variables:
    MIGRATION_ENVIRONMENT: ${environment}
    DATABASE_CONNECTION_STRING: ${database_url}
    GRAFANA_ENDPOINT: ${grafana_endpoint}
```

### Local Development Pattern
```yaml
# Podman Compose migration init containers
services-api-migrations:
  build: "./services-api"
  command: ["go", "run", "./cmd/migrate", "--environment=development", "--aggressive"]
  environment:
    - DATABASE_CONNECTION_STRING=${SERVICES_DATABASE_CONNECTION_STRING}
    - MIGRATION_ENVIRONMENT=development
    - MIGRATION_APPROACH=aggressive
  depends_on: [postgresql, vault]
  restart: "no"
```

## Database Schema References

### Schema Compliance Validation
All database schemas must exactly match specifications defined in the authoritative schema files. Schema validation is performed during migration execution in staging and production environments.

### Services Domain Schema
**Authoritative Schema Definition**: [`/doc/dapr-services/SERVICES-SCHEMA.md`](../dapr-services/SERVICES-SCHEMA.md)

The Services domain includes:
- Services table with publishing workflow support
- Service categories table with default assignment logic
- Featured categories table with position constraints
- PostgreSQL performance indexes for query optimization
- Grafana Cloud Loki audit event integration
- Azure Blob Storage content management

### Content Domain Schema  
**Authoritative Schema Definition**: [`/doc/dapr-services/CONTENT-SCHEMA.md`](../dapr-services/CONTENT-SCHEMA.md)

The Content domain includes:
- Content table with virus scanning integration
- Content access logging for analytics
- Content virus scan tracking
- Storage backend management with health monitoring
- PostgreSQL performance indexes for content queries
- Grafana Cloud Loki audit event integration

### Migration Naming Convention
```
{version}_{description}.{direction}.sql

Examples:
- 000001_initial_schema.up.sql
- 000001_initial_schema.down.sql
- 000002_add_services_table.up.sql
- 000002_add_services_table.down.sql
```

### Migration File Requirements
1. **Atomic Operations**: Each migration file contains atomic database operations
2. **Idempotent**: Migrations can be run multiple times safely
3. **Backward Compatibility**: Up migrations maintain data integrity
4. **Rollback Support**: Down migrations provide clean rollback path
5. **Schema Compliance**: All changes documented in this specification file

## Rollback Strategies

### Development Environment Rollback
```bash
# Development: Aggressive rollback (destroy and recreate)
podman-compose down -v
podman volume prune -f
podman-compose up -d
```

### Staging Environment Rollback
```bash
# Staging: Careful rollback with confirmation
./migrate --environment=staging --rollback-to-version={version}
# Requires confirmation prompt and validation
```

### Production Environment Rollback
```bash
# Production: Conservative rollback with manual approval
./migrate --environment=production --rollback-to-version={version} --require-approval
# Requires manual approval via GitHub Actions environment
```

### Rollback Validation Process
1. **Pre-Rollback Backup**: Full database backup before rollback
2. **Schema Validation**: Verify target schema matches specification
3. **Data Integrity Check**: Validate data consistency after rollback
4. **Application Compatibility**: Ensure application works with rolled-back schema
5. **Audit Logging**: All rollback operations logged to Grafana Cloud Loki

## CI/CD Pipeline Integration

### Migration Execution Order in Pulumi
```go
// Pulumi resource dependency chain
foundationResources := deployFoundation(ctx, environment)
migrations := deployMigrations(ctx, environment, foundationResources)
applications := deployApplications(ctx, environment, []pulumi.Resource{migrations})
```

### GitHub Actions Integration
```yaml
# Migration tool built during deployment
- name: Build Migration Tools
  run: |
    cd ../migrations
    go build -o migrate ./cmd/migrate/
    chmod +x migrate

# Migration execution via Pulumi
- name: Deploy Infrastructure + Migrations + Applications
  run: pulumi up --yes --stack staging
```

### Migration Status Verification
```yaml
# Post-deployment verification
- name: Verify Migration Success
  run: |
    POSTGRES_MIGRATION_STATUS=$(pulumi stack output postgresMigrationStatus)
    COSMOSDB_MIGRATION_STATUS=$(pulumi stack output cosmosdbMigrationStatus)
    SCHEMA_VALIDATION_RESULT=$(pulumi stack output schemaValidationResult)
    
    if [ "$POSTGRES_MIGRATION_STATUS" != "success" ] || 
       [ "$COSMOSDB_MIGRATION_STATUS" != "success" ] || 
       [ "$SCHEMA_VALIDATION_RESULT" != "compliant" ]; then
      echo "ERROR: Migration failed or schema non-compliant"
      exit 1
    fi
```

## Environment Configuration

### Migration Tool Configuration
All migration configuration via environment variables (no hardcoded values):

```bash
# Database Configuration
DATABASE_CONNECTION_STRING=${POSTGRES_CONNECTION_STRING}
COSMOSDB_ENDPOINT=${COSMOSDB_ENDPOINT}

# Migration Behavior
MIGRATION_ENVIRONMENT=${ENVIRONMENT}  # development|staging|production
MIGRATION_APPROACH=${APPROACH}        # aggressive|careful|conservative
MIGRATION_TIMEOUT=${TIMEOUT}          # Environment-specific timeout

# Observability
GRAFANA_ENDPOINT=${GRAFANA_ENDPOINT}
VAULT_ENDPOINT=${VAULT_ENDPOINT}

# Audit Logging
CORRELATION_ID=${CORRELATION_ID}
TRACE_ID=${TRACE_ID}
USER_ID=${USER_ID}
```

### Environment-Specific Overrides
```yaml
# Development overrides
MIGRATION_TIMEOUT=30s
MIGRATION_BACKUP=false
MIGRATION_VALIDATION=minimal

# Staging overrides
MIGRATION_TIMEOUT=300s
MIGRATION_BACKUP=incremental
MIGRATION_VALIDATION=moderate

# Production overrides
MIGRATION_TIMEOUT=900s
MIGRATION_BACKUP=full
MIGRATION_VALIDATION=extensive
MIGRATION_MANUAL_APPROVAL=true
```

## Migration Testing Patterns

### Unit Testing Migration Logic
```go
// migrations/migrator_test.go
func TestMigrator_DevelopmentApproach(t *testing.T) {
    // Arrange
    migrator := &Migrator{
        Environment: "development",
        DatabaseURL: testDB.ConnectionString(),
    }
    
    // Act
    err := migrator.RunMigrations(context.Background())
    
    // Assert
    assert.NoError(t, err)
    // Verify schema matches specification
}
```

### Integration Testing Requirements
1. **Full Environment**: Only run integration tests when complete Podman Compose environment is running
2. **Infrastructure Dependencies**: Integration with PostgreSQL, CosmosDB emulator, Vault
3. **Schema Validation**: Verify migrated schema exactly matches specification files
4. **Audit Verification**: Confirm audit events published to Grafana Cloud Loki
5. **Timeout Compliance**: Migration tests must complete within environment-specific timeouts

### Migration Test Execution
```bash
# Integration tests require full environment
podman-compose ps | grep -v "Up" && echo "Environment not ready" || echo "Ready for migration tests"

# Wait for all services healthy
while [ "$(podman-compose ps --filter 'health=healthy' --format '{{.Names}}' | wc -l)" -lt 10 ]; do
  echo "Waiting for services to be healthy..."
  sleep 2
done

# Run migration integration tests
go test ./migrations/... -tags=integration -timeout=15s
```

## Compliance and Audit Requirements

### Migration Audit Events
All migration operations generate audit events with complete data snapshots:

```json
{
  "audit_id": "uuid",
  "operation_type": "MIGRATION_EXECUTE|MIGRATION_ROLLBACK",
  "migration_version": "000005",
  "environment": "production|staging|development", 
  "correlation_id": "uuid",
  "trace_id": "string",
  "user_id": "system|user-id",
  "migration_details": {
    "approach": "conservative|careful|aggressive",
    "backup_status": "success|failed|skipped",
    "validation_result": "passed|failed",
    "duration_seconds": 45,
    "affected_tables": ["services", "service_categories"]
  },
  "audit_timestamp": "timestamp"
}
```

### Compliance Validation Checklist
1. ✅ **Schema Compliance**: Database schema matches specification files exactly
2. ✅ **Audit Completeness**: All migration operations logged to Grafana Cloud Loki
3. ✅ **No Local Audit Storage**: No audit tables in PostgreSQL (compliance requirement)
4. ✅ **Backup Verification**: Production migrations include full backup verification
5. ✅ **Manual Approval**: Production rollbacks require manual approval gates
6. ✅ **Immutable Audit Trail**: Audit events never modified or deleted
7. ✅ **Environment Isolation**: Migration tools configuration via environment variables only
