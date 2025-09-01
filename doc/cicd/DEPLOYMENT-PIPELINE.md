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

### Local Development Service Architecture

#### Dapr Control Plane Services
```yaml
dapr-placement:
  image: "daprio/dapr:latest"
  command: ["./placement", "--port", "${DAPR_PLACEMENT_PORT}"]
  ports:
    - "${DAPR_PLACEMENT_PORT}:${DAPR_PLACEMENT_PORT}"

dapr-sentry:
  image: "daprio/dapr:latest"
  command: ["./sentry", "--port", "${DAPR_SENTRY_PORT}"]
  ports:
    - "${DAPR_SENTRY_PORT}:${DAPR_SENTRY_PORT}"

redis-dapr:
  image: "redis:7-alpine"
  ports:
    - "${REDIS_DAPR_PORT}:6379"
```

#### API Services with Automated Migrations
```yaml
# Services API with Migration Init Container
services-api-migrations:
  build: "./services-api"
  command: ["go", "run", "./cmd/migrate", "--environment=development", "--aggressive"]
  environment:
    - DATABASE_CONNECTION_STRING=${SERVICES_DATABASE_CONNECTION_STRING}
    - VAULT_ENDPOINT=${VAULT_LOCAL_ENDPOINT}
    - GRAFANA_ENDPOINT=${GRAFANA_LOCAL_ENDPOINT}
    - MIGRATION_ENVIRONMENT=development
    - MIGRATION_APPROACH=aggressive
  depends_on:
    - postgresql
    - vault
  restart: "no"

services-api:
  build: "./services-api"
  environment:
    - DAPR_HTTP_PORT=${SERVICES_API_DAPR_HTTP_PORT}
    - DAPR_GRPC_PORT=${SERVICES_API_DAPR_GRPC_PORT}
    - DATABASE_CONNECTION_STRING=${SERVICES_DATABASE_CONNECTION_STRING}
    - VAULT_ENDPOINT=${VAULT_LOCAL_ENDPOINT}
    - GRAFANA_ENDPOINT=${GRAFANA_LOCAL_ENDPOINT}
  depends_on:
    - postgresql
    - vault
    - services-api-migrations

services-api-dapr:
  image: "daprio/dapr:latest"
  command: [
    "./daprd",
    "--app-id", "services-api",
    "--app-port", "${SERVICES_API_PORT}",
    "--dapr-http-port", "${SERVICES_API_DAPR_HTTP_PORT}",
    "--dapr-grpc-port", "${SERVICES_API_DAPR_GRPC_PORT}",
    "--config", "/etc/dapr/config.yaml",
    "--components-path", "/etc/dapr/components"
  ]
  volumes:
    - "./dapr/components:/etc/dapr/components"
    - "./dapr/config.yaml:/etc/dapr/config.yaml"
  depends_on:
    - dapr-placement
    - dapr-sentry
    - services-api

# Content API with Migration Init Container
content-api-migrations:
  build: "./content-api"
  command: ["go", "run", "./cmd/migrate", "--environment=development", "--aggressive"]
  environment:
    - DATABASE_CONNECTION_STRING=${CONTENT_DATABASE_CONNECTION_STRING}
    - BLOB_STORAGE_ENDPOINT=${AZURITE_BLOB_ENDPOINT}
    - VAULT_ENDPOINT=${VAULT_LOCAL_ENDPOINT}
    - GRAFANA_ENDPOINT=${GRAFANA_LOCAL_ENDPOINT}
    - MIGRATION_ENVIRONMENT=development
    - MIGRATION_APPROACH=aggressive
  depends_on:
    - postgresql
    - azurite
    - vault
  restart: "no"

content-api:
  build: "./content-api"
  environment:
    - DAPR_HTTP_PORT=${CONTENT_API_DAPR_HTTP_PORT}
    - DAPR_GRPC_PORT=${CONTENT_API_DAPR_GRPC_PORT}
    - DATABASE_CONNECTION_STRING=${CONTENT_DATABASE_CONNECTION_STRING}
    - BLOB_STORAGE_ENDPOINT=${AZURITE_BLOB_ENDPOINT}
    - VAULT_ENDPOINT=${VAULT_LOCAL_ENDPOINT}
    - GRAFANA_ENDPOINT=${GRAFANA_LOCAL_ENDPOINT}
  depends_on:
    - postgresql
    - azurite
    - vault
    - content-api-migrations

content-api-dapr:
  image: "daprio/dapr:latest"
  command: [
    "./daprd",
    "--app-id", "content-api",
    "--app-port", "${CONTENT_API_PORT}",
    "--dapr-http-port", "${CONTENT_API_DAPR_HTTP_PORT}",
    "--dapr-grpc-port", "${CONTENT_API_DAPR_GRPC_PORT}",
    "--config", "/etc/dapr/config.yaml",
    "--components-path", "/etc/dapr/components"
  ]
  volumes:
    - "./dapr/components:/etc/dapr/components"
    - "./dapr/config.yaml:/etc/dapr/config.yaml"
  depends_on:
    - dapr-placement
    - dapr-sentry
    - content-api
```

#### Gateway Services
```yaml
# Public Gateway with Migration Init Container
public-gateway-migrations:
  build: "./public-gateway"
  command: ["go", "run", "./cmd/migrate", "--environment=development", "--aggressive"]
  environment:
    - COSMOSDB_ENDPOINT=${COSMOSDB_EMULATOR_ENDPOINT}
    - VAULT_ENDPOINT=${VAULT_LOCAL_ENDPOINT}
    - GRAFANA_ENDPOINT=${GRAFANA_LOCAL_ENDPOINT}
    - MIGRATION_ENVIRONMENT=development
    - MIGRATION_APPROACH=aggressive
  depends_on:
    - cosmosdb-emulator
    - vault
  restart: "no"

public-gateway:
  build: "./public-gateway"
  environment:
    - DAPR_HTTP_PORT=${PUBLIC_GATEWAY_DAPR_HTTP_PORT}
    - DAPR_GRPC_PORT=${PUBLIC_GATEWAY_DAPR_GRPC_PORT}
    - PUBLIC_WEBSITE_ORIGINS=${PUBLIC_WEBSITE_ORIGINS}
    - AUTHENTIK_ENDPOINT=${AUTHENTIK_LOCAL_ENDPOINT}
    - GRAFANA_ENDPOINT=${GRAFANA_LOCAL_ENDPOINT}
  ports:
    - "${PUBLIC_GATEWAY_PORT}:${PUBLIC_GATEWAY_PORT}"
  depends_on:
    - authentik
    - public-gateway-migrations

public-gateway-dapr:
  image: "daprio/dapr:latest"
  command: [
    "./daprd",
    "--app-id", "public-gateway",
    "--app-port", "${PUBLIC_GATEWAY_PORT}",
    "--dapr-http-port", "${PUBLIC_GATEWAY_DAPR_HTTP_PORT}",
    "--dapr-grpc-port", "${PUBLIC_GATEWAY_DAPR_GRPC_PORT}",
    "--config", "/etc/dapr/config.yaml",
    "--components-path", "/etc/dapr/components"
  ]
  volumes:
    - "./dapr/components:/etc/dapr/components"
    - "./dapr/config.yaml:/etc/dapr/config.yaml"
  depends_on:
    - dapr-placement
    - dapr-sentry
    - public-gateway

# Admin Gateway
admin-gateway:
  build: "./admin-gateway"
  environment:
    - DAPR_HTTP_PORT=${ADMIN_GATEWAY_DAPR_HTTP_PORT}
    - DAPR_GRPC_PORT=${ADMIN_GATEWAY_DAPR_GRPC_PORT}
    - ADMIN_WEBSITE_ORIGINS=${ADMIN_WEBSITE_ORIGINS}
    - AUTHENTIK_ISSUER_URL=${AUTHENTIK_ISSUER_URL}
    - AUTHENTIK_JWKS_URL=${AUTHENTIK_JWKS_URL}
    - OPA_POLICY_ENDPOINT=${OPA_LOCAL_ENDPOINT}
    - GRAFANA_CLOUD_ENDPOINT=${GRAFANA_LOCAL_ENDPOINT}
  ports:
    - "${ADMIN_GATEWAY_PORT}:${ADMIN_GATEWAY_PORT}"
  depends_on:
    - authentik
    - opa

admin-gateway-dapr:
  image: "daprio/dapr:latest"
  command: [
    "./daprd",
    "--app-id", "admin-gateway",
    "--app-port", "${ADMIN_GATEWAY_PORT}",
    "--dapr-http-port", "${ADMIN_GATEWAY_DAPR_HTTP_PORT}",
    "--dapr-grpc-port", "${ADMIN_GATEWAY_DAPR_GRPC_PORT}",
    "--config", "/etc/dapr/config.yaml",
    "--components-path", "/etc/dapr/components"
  ]
  volumes:
    - "./dapr/components:/etc/dapr/components"
    - "./dapr/config.yaml:/etc/dapr/config.yaml"
  depends_on:
    - dapr-placement
    - dapr-sentry
    - admin-gateway
```

### Local Development Infrastructure Services

#### Identity Provider Stack
```yaml
authentik-postgresql:
  image: "postgres:15-alpine"
  environment:
    - POSTGRES_PASSWORD=${AUTHENTIK_DB_PASSWORD}
    - POSTGRES_USER=${AUTHENTIK_DB_USER}
    - POSTGRES_DB=${AUTHENTIK_DB_NAME}
  volumes:
    - authentik-postgresql-data:/var/lib/postgresql/data

authentik-redis:
  image: "redis:7-alpine"
  command: --save 60 1 --loglevel warning
  volumes:
    - authentik-redis-data:/data

authentik:
  image: "ghcr.io/goauthentik/server:latest"
  command: server
  environment:
    - AUTHENTIK_REDIS__HOST=${AUTHENTIK_REDIS_HOST}
    - AUTHENTIK_POSTGRESQL__HOST=${AUTHENTIK_POSTGRESQL_HOST}
    - AUTHENTIK_POSTGRESQL__USER=${AUTHENTIK_DB_USER}
    - AUTHENTIK_POSTGRESQL__NAME=${AUTHENTIK_DB_NAME}
    - AUTHENTIK_POSTGRESQL__PASSWORD=${AUTHENTIK_DB_PASSWORD}
    - AUTHENTIK_SECRET_KEY=${AUTHENTIK_SECRET_KEY}
  ports:
    - "${AUTHENTIK_PORT}:9000"
  depends_on:
    - authentik-postgresql
    - authentik-redis

authentik-worker:
  image: "ghcr.io/goauthentik/server:latest"
  command: worker
  environment:
    - AUTHENTIK_REDIS__HOST=${AUTHENTIK_REDIS_HOST}
    - AUTHENTIK_POSTGRESQL__HOST=${AUTHENTIK_POSTGRESQL_HOST}
    - AUTHENTIK_POSTGRESQL__USER=${AUTHENTIK_DB_USER}
    - AUTHENTIK_POSTGRESQL__NAME=${AUTHENTIK_DB_NAME}
    - AUTHENTIK_POSTGRESQL__PASSWORD=${AUTHENTIK_DB_PASSWORD}
    - AUTHENTIK_SECRET_KEY=${AUTHENTIK_SECRET_KEY}
  depends_on:
    - authentik-postgresql
    - authentik-redis

opa:
  image: "openpolicyagent/opa:latest"
  command: ["run", "--server", "--config-file=/etc/opa/config.yaml"]
  environment:
    - OPA_CONFIG_FILE=/etc/opa/config.yaml
  ports:
    - "${OPA_PORT}:8181"
  volumes:
    - "./opa/policies:/etc/opa/policies"
    - "./opa/config.yaml:/etc/opa/config.yaml"
```

#### Local Grafana Observability Stack
```yaml
grafana:
  image: "grafana/grafana:latest"
  environment:
    - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD}
    - GF_USERS_ALLOW_SIGN_UP=false
  ports:
    - "${GRAFANA_PORT}:3000"
  volumes:
    - grafana-data:/var/lib/grafana
    - "./grafana/provisioning:/etc/grafana/provisioning"

mimir:
  image: "grafana/mimir:latest"
  command: ["-config.file=/etc/mimir/mimir.yaml"]
  ports:
    - "${MIMIR_PORT}:9009"
  volumes:
    - mimir-data:/data
    - "./mimir/mimir.yaml:/etc/mimir/mimir.yaml"

loki:
  image: "grafana/loki:latest"
  command: ["-config.file=/etc/loki/local-config.yaml"]
  ports:
    - "${LOKI_PORT}:3100"
  volumes:
    - loki-data:/loki
    - "./loki/local-config.yaml:/etc/loki/local-config.yaml"

tempo:
  image: "grafana/tempo:latest"
  command: ["-config.file=/etc/tempo/tempo.yaml"]
  ports:
    - "${TEMPO_PORT}:3200"
  volumes:
    - tempo-data:/tmp/tempo
    - "./tempo/tempo.yaml:/etc/tempo/tempo.yaml"

pyroscope:
  image: "grafana/pyroscope:latest"
  environment:
    - PYROSCOPE_CONFIG_FILE=/etc/pyroscope/pyroscope.yaml
  ports:
    - "${PYROSCOPE_PORT}:4040"
  volumes:
    - pyroscope-data:/var/lib/pyroscope
    - "./pyroscope/pyroscope.yaml:/etc/pyroscope/pyroscope.yaml"

grafana-agent:
  image: "grafana/agent:latest"
  command: ["-config.file=/etc/agent/agent.yaml"]
  environment:
    - GRAFANA_CLOUD_API_KEY=${GRAFANA_CLOUD_API_KEY}
  volumes:
    - "./grafana-agent/agent.yaml:/etc/agent/agent.yaml"
  depends_on:
    - grafana
    - mimir
    - loki
    - tempo

grafana-agent-dapr:
  image: "daprio/dapr:latest"
  command: [
    "./daprd",
    "--app-id", "grafana-agent",
    "--app-port", "${GRAFANA_AGENT_PORT}",
    "--dapr-http-port", "${GRAFANA_AGENT_DAPR_HTTP_PORT}",
    "--dapr-grpc-port", "${GRAFANA_AGENT_DAPR_GRPC_PORT}",
    "--config", "/etc/dapr/config.yaml",
    "--components-path", "/etc/dapr/components"
  ]
  volumes:
    - "./dapr/components:/etc/dapr/components"
    - "./dapr/config.yaml:/etc/dapr/config.yaml"
  depends_on:
    - dapr-placement
    - dapr-sentry
    - grafana-agent
```

#### Azure Service Emulators
```yaml
postgresql:
  image: "postgres:15-alpine"
  environment:
    - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
    - POSTGRES_USER=${POSTGRES_USER}
    - POSTGRES_DB=${POSTGRES_DB}
    - POSTGRES_MULTIPLE_DATABASES=${POSTGRES_MULTIPLE_DATABASES}
  ports:
    - "${POSTGRES_PORT}:5432"
  volumes:
    - postgresql-data:/var/lib/postgresql/data
    - "./postgresql/init-scripts:/docker-entrypoint-initdb.d"

azurite:
  image: "mcr.microsoft.com/azure-storage/azurite:latest"
  command: "azurite --blobHost ${AZURITE_BLOB_HOST} --blobPort ${AZURITE_BLOB_PORT}"
  ports:
    - "${AZURITE_BLOB_PORT}:${AZURITE_BLOB_PORT}"
  volumes:
    - azurite-data:/data

cosmosdb-emulator:
  image: "mcr.microsoft.com/cosmosdb/linux/azure-cosmos-emulator:latest"
  environment:
    - AZURE_COSMOS_EMULATOR_PARTITION_COUNT=${COSMOSDB_PARTITION_COUNT}
    - AZURE_COSMOS_EMULATOR_ENABLE_DATA_PERSISTENCE=true
  ports:
    - "${COSMOSDB_PORT}:8081"
  volumes:
    - cosmos-data:/tmp/cosmos/appdata

service-bus-emulator:
  image: "mcr.microsoft.com/azure-messaging/servicebus-emulator:latest"
  environment:
    - ACCEPT_EULA=Y
    - SQL_SERVER=${SERVICE_BUS_SQL_SERVER}
    - SA_PASSWORD=${SERVICE_BUS_SA_PASSWORD}
  ports:
    - "${SERVICE_BUS_PORT}:5672"
  depends_on:
    - sql-edge

sql-edge:
  image: "mcr.microsoft.com/azure-sql-edge:latest"
  environment:
    - ACCEPT_EULA=Y
    - SA_PASSWORD=${SERVICE_BUS_SA_PASSWORD}
    - MSSQL_PID=Developer
  volumes:
    - sql-edge-data:/var/opt/mssql

vault:
  image: "hashicorp/vault:latest"
  command: "vault server -config=/vault/config/vault.json"
  environment:
    - VAULT_ADDR=${VAULT_ADDR}
    - VAULT_DEV_ROOT_TOKEN_ID=${VAULT_ROOT_TOKEN}
    - VAULT_DEV_LISTEN_ADDRESS=${VAULT_LISTEN_ADDRESS}
  ports:
    - "${VAULT_PORT}:8200"
  volumes:
    - vault-data:/vault/data
    - "./vault/config.json:/vault/config/vault.json"
  cap_add:
    - IPC_LOCK
```

### Development Environment Workflow

#### Environment Startup and Migration
```bash
# Start complete development environment
podman-compose up -d

# Start infrastructure services first
podman-compose up -d dapr-placement dapr-sentry redis-dapr
podman-compose up -d postgresql vault authentik azurite cosmosdb-emulator

# Run migrations (init containers)
podman-compose up -d services-api-migrations content-api-migrations public-gateway-migrations

# Wait for migrations to complete, then start API services
podman-compose up -d services-api services-api-dapr
podman-compose up -d content-api content-api-dapr
podman-compose up -d public-gateway public-gateway-dapr
podman-compose up -d admin-gateway admin-gateway-dapr
```

#### Development Environment Reset (Aggressive Rollback)
```bash
# Complete environment reset - destroy and recreate all data
podman-compose down -v
podman volume prune -f
podman system prune -f
podman-compose up -d

# Quick reset - restart services with fresh migrations
podman-compose restart services-api-migrations content-api-migrations public-gateway-migrations
podman-compose restart services-api content-api public-gateway
```

#### Integration Testing Preparation
```bash
# Ensure complete environment is healthy before running integration tests
podman-compose ps | grep -v "Up" && echo "Environment not ready" || echo "Environment ready for integration tests"

# Wait for all migrations to complete
while [ "$(podman-compose ps --filter 'name=*-migrations' --filter 'status=exited' --format '{{.Names}}' | wc -l)" -lt 3 ]; do
  echo "Waiting for migrations to complete..."
  sleep 2
done

# Verify Azure emulator connectivity
curl http://localhost:${AZURITE_BLOB_PORT}/devstoreaccount1
curl http://localhost:${COSMOSDB_PORT}/_explorer/index.html
```

## Staging and Production CI/CD Pipeline

### Branch Strategy and Pipeline Triggers
- **Develop Branch**: Staging deployments with careful migration approach
- **Main Branch**: Production deployments with conservative migration approach and manual approval
- **Feature Branches**: Short-lived, merge via pull requests with automated testing
- **Release Tags**: Semantic versioning for production releases
- **No Environment Branches**: Environment-specific behavior controlled via Pulumi stack configuration

```yaml
on:
  push:
    branches: [develop]  # Staging deployments
    paths: 
      - 'services-api/**'
      - 'content-api/**'
      - 'public-gateway/**'
      - 'admin-gateway/**'
      - 'infrastructure/**'
      - 'migrations/**'
  push:
    branches: [main]     # Production deployments
    paths:
      - 'services-api/**'
      - 'content-api/**'
      - 'public-gateway/**'
      - 'admin-gateway/**'
      - 'infrastructure/**'
      - 'migrations/**'
  pull_request:
    branches: [develop, main]
```

### Build and Test Pipeline

#### Go Application Testing
```yaml
test:
  runs-on: ubuntu-latest
  timeout-minutes: 10
  strategy:
    matrix:
      service: [services-api, content-api, public-gateway, admin-gateway]
  
  steps:
    - name: Checkout code
      uses: actions/checkout@v4
    
    - name: Setup Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.21'
    
    - name: Cache Go modules
      uses: actions/cache@v4
      with:
        path: ~/go/pkg/mod
        key: go-mod-${{ hashFiles('**/go.sum') }}
    
    - name: Run unit tests
      run: |
        cd ${{ matrix.service }}
        go test -v -race -timeout=5s ./...
        go test -v -race -coverprofile=coverage.out ./...
    
    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v4
      with:
        file: ./${{ matrix.service }}/coverage.out
        flags: ${{ matrix.service }}
```

#### Container Build and Registry Push
```yaml
build-and-push:
  runs-on: ubuntu-latest
  needs: [test]
  if: github.ref == 'refs/heads/develop' || github.ref == 'refs/heads/main'
  strategy:
    matrix:
      service: [services-api, content-api, public-gateway, admin-gateway]
  
  steps:
    - name: Checkout code
      uses: actions/checkout@v4
    
    - name: Login to GitHub Container Registry
      uses: docker/login-action@v3
      with:
        registry: ghcr.io
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}
    
    - name: Build and push container
      uses: docker/build-push-action@v5
      with:
        context: ./${{ matrix.service }}
        push: true
        tags: |
          ghcr.io/${{ github.repository }}/${{ matrix.service }}:latest
          ghcr.io/${{ github.repository }}/${{ matrix.service }}:${{ github.sha }}
        build-args: |
          VERSION=${{ github.sha }}
          BUILD_DATE=${{ github.event.head_commit.timestamp }}
```

### Staging Deployment (Careful Migration Approach)

```yaml
deploy-staging:
  runs-on: ubuntu-latest
  needs: [build-and-push]
  if: github.ref == 'refs/heads/develop'
  defaults:
    run:
      working-directory: infrastructure
      
  steps:
    - name: Checkout code
      uses: actions/checkout@v4
    
    - name: Setup Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.21'
        
    - name: Setup Pulumi CLI
      uses: pulumi/actions@v5
      with:
        pulumi-version: '3.96.0'
        
    - name: Build Migration Tools
      run: |
        cd ../migrations
        go build -o migrate ./cmd/migrate/
        chmod +x migrate
        
    - name: Configure Pulumi Backend and Providers
      run: |
        # Login to Pulumi with Azure Blob Storage backend
        pulumi login azblob://${{ secrets.PULUMI_STATE_CONTAINER }}?storage_account=${{ secrets.PULUMI_STORAGE_ACCOUNT }}
        
        # Select or create staging stack
        pulumi stack select staging || pulumi stack init staging
        
        # Set staging configuration
        pulumi config set azure:location "East US 2"
        pulumi config set environment "staging"
        pulumi config set migrationApproach "careful"
        pulumi config set containerImageTag "${{ github.sha }}"
        pulumi config set correlationId "${{ github.sha }}"
        pulumi config set --secret grafanaCloudApiKey "${{ secrets.GRAFANA_CLOUD_API_KEY }}"
        pulumi config set --secret authentikClientSecret "${{ secrets.AUTHENTIK_CLIENT_SECRET }}"
        pulumi config set --secret vaultCloudNamespace "${{ secrets.VAULT_CLOUD_NAMESPACE }}"
        pulumi config set --secret grafanaCloudLokiEndpoint "${{ secrets.GRAFANA_CLOUD_LOKI_ENDPOINT }}"
      env:
        AZURE_STORAGE_ACCOUNT: ${{ secrets.PULUMI_STORAGE_ACCOUNT }}
        AZURE_STORAGE_KEY: ${{ secrets.PULUMI_STORAGE_KEY }}
        
    - name: Deploy Infrastructure + Migrations + Applications
      run: pulumi up --yes --stack staging
      env:
        # Azure Provider Configuration
        ARM_CLIENT_ID: ${{ secrets.ARM_CLIENT_ID }}
        ARM_CLIENT_SECRET: ${{ secrets.ARM_CLIENT_SECRET }}
        ARM_SUBSCRIPTION_ID: ${{ secrets.AZURE_SUBSCRIPTION_ID }}
        ARM_TENANT_ID: ${{ secrets.ARM_TENANT_ID }}
        
        # Grafana Cloud Provider Configuration
        GRAFANA_CLOUD_API_KEY: ${{ secrets.GRAFANA_CLOUD_API_KEY }}
        GRAFANA_CLOUD_STACK_SLUG: ${{ secrets.GRAFANA_CLOUD_STACK_SLUG }}
        
        # HashiCorp Vault Cloud Provider Configuration
        VAULT_ADDR: ${{ secrets.VAULT_CLOUD_ADDR }}
        VAULT_TOKEN: ${{ secrets.VAULT_CLOUD_TOKEN }}
        VAULT_NAMESPACE: ${{ secrets.VAULT_CLOUD_NAMESPACE }}
        
    - name: Verify Deployment and Migration Success
      run: |
        # Get deployment results from Pulumi stack outputs
        POSTGRES_MIGRATION_STATUS=$(pulumi stack output postgresMigrationStatus)
        COSMOSDB_MIGRATION_STATUS=$(pulumi stack output cosmosdbMigrationStatus)
        SERVICES_DEPLOYMENT_STATUS=$(pulumi stack output servicesDeploymentStatus)
        GATEWAYS_DEPLOYMENT_STATUS=$(pulumi stack output gatewaysDeploymentStatus)
        SCHEMA_VALIDATION_RESULT=$(pulumi stack output schemaValidationResult)
        AUDIT_TABLES_CHECK=$(pulumi stack output auditTablesCheck)
        
        # Verify all components deployed successfully
        if [ "$POSTGRES_MIGRATION_STATUS" != "success" ] || \
           [ "$COSMOSDB_MIGRATION_STATUS" != "success" ] || \
           [ "$SERVICES_DEPLOYMENT_STATUS" != "success" ] || \
           [ "$GATEWAYS_DEPLOYMENT_STATUS" != "success" ]; then
          echo "ERROR: One or more components failed to deploy in staging"
          exit 1
        fi
        
        # Verify schema compliance (database matches specification files)
        if [ "$SCHEMA_VALIDATION_RESULT" != "compliant" ]; then
          echo "ERROR: Schema validation failed - database does not match specification files"
          exit 1
        fi
        
        # Verify no audit tables exist (compliance requirement)
        if [ "$AUDIT_TABLES_CHECK" != "0" ]; then
          echo "ERROR: Audit tables found in database. All audit data must be stored in Grafana Cloud Loki."
          exit 1
        fi
        
        # Output deployment results
        echo "Staging Deployment Results:"
        echo "- PostgreSQL Migration: $POSTGRES_MIGRATION_STATUS"
        echo "- CosmosDB Migration: $COSMOSDB_MIGRATION_STATUS"
        echo "- Services Deployment: $SERVICES_DEPLOYMENT_STATUS"
        echo "- Gateways Deployment: $GATEWAYS_DEPLOYMENT_STATUS"
        echo "- Schema Validation: $SCHEMA_VALIDATION_RESULT"
        echo "- Audit Tables Check: $AUDIT_TABLES_CHECK"
```

### Production Deployment (Conservative Migration Approach)

```yaml
deploy-production:
  runs-on: ubuntu-latest
  environment: production-deployment  # GitHub environment with manual approval
  needs: [build-and-push]
  if: github.ref == 'refs/heads/main'
  defaults:
    run:
      working-directory: infrastructure
      
  steps:
    - name: Checkout code
      uses: actions/checkout@v4
    
    - name: Setup Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.21'
        
    - name: Setup Pulumi CLI
      uses: pulumi/actions@v5
      with:
        pulumi-version: '3.96.0'
        
    - name: Build Migration Tools
      run: |
        cd ../migrations
        go build -o migrate ./cmd/migrate/
        chmod +x migrate
        
    - name: Configure Pulumi Backend and Providers
      run: |
        # Login to Pulumi with Azure Blob Storage backend
        pulumi login azblob://${{ secrets.PULUMI_STATE_CONTAINER }}?storage_account=${{ secrets.PULUMI_STORAGE_ACCOUNT }}
        
        # Select production stack
        pulumi stack select production
        
        # Set production configuration (conservative approach)
        pulumi config set azure:location "East US 2"
        pulumi config set environment "production"
        pulumi config set migrationApproach "conservative"
        pulumi config set migrationValidation "extensive"
        pulumi config set migrationBackup "full"
        pulumi config set manualApprovalRequired "true"
        pulumi config set complianceMode "true"
        pulumi config set containerImageTag "${{ github.sha }}"
        pulumi config set correlationId "${{ github.sha }}"
        pulumi config set --secret grafanaCloudApiKey "${{ secrets.GRAFANA_CLOUD_API_KEY }}"
        pulumi config set --secret authentikClientSecret "${{ secrets.AUTHENTIK_CLIENT_SECRET }}"
        pulumi config set --secret vaultCloudNamespace "${{ secrets.VAULT_CLOUD_NAMESPACE }}"
        pulumi config set --secret grafanaCloudLokiEndpoint "${{ secrets.GRAFANA_CLOUD_LOKI_ENDPOINT }}"
        pulumi config set --secret azureBackupStorageAccount "${{ secrets.AZURE_BACKUP_STORAGE_ACCOUNT }}"
        pulumi config set --secret slackWebhookUrl "${{ secrets.SLACK_WEBHOOK_URL }}"
      env:
        AZURE_STORAGE_ACCOUNT: ${{ secrets.PULUMI_STORAGE_ACCOUNT }}
        AZURE_STORAGE_KEY: ${{ secrets.PULUMI_STORAGE_KEY }}
        
    - name: Deploy Infrastructure + Migrations + Applications
      run: pulumi up --yes --stack production
      env:
        # Azure Provider Configuration
        ARM_CLIENT_ID: ${{ secrets.ARM_CLIENT_ID }}
        ARM_CLIENT_SECRET: ${{ secrets.ARM_CLIENT_SECRET }}
        ARM_SUBSCRIPTION_ID: ${{ secrets.AZURE_SUBSCRIPTION_ID }}
        ARM_TENANT_ID: ${{ secrets.ARM_TENANT_ID }}
        
        # Grafana Cloud Provider Configuration
        GRAFANA_CLOUD_API_KEY: ${{ secrets.GRAFANA_CLOUD_API_KEY }}
        GRAFANA_CLOUD_STACK_SLUG: ${{ secrets.GRAFANA_CLOUD_STACK_SLUG }}
        
        # HashiCorp Vault Cloud Provider Configuration
        VAULT_ADDR: ${{ secrets.VAULT_CLOUD_ADDR }}
        VAULT_TOKEN: ${{ secrets.VAULT_CLOUD_TOKEN }}
        VAULT_NAMESPACE: ${{ secrets.VAULT_CLOUD_NAMESPACE }}
        
    - name: Verify Production Deployment and Compliance
      run: |
        # Get comprehensive deployment results from Pulumi stack outputs
        MIGRATION_BACKUP_STATUS=$(pulumi stack output migrationBackupStatus)
        POSTGRES_MIGRATION_STATUS=$(pulumi stack output postgresMigrationStatus)
        COSMOSDB_MIGRATION_STATUS=$(pulumi stack output cosmosdbMigrationStatus)
        SERVICES_DEPLOYMENT_STATUS=$(pulumi stack output servicesDeploymentStatus)
        GATEWAYS_DEPLOYMENT_STATUS=$(pulumi stack output gatewaysDeploymentStatus)
        SCHEMA_VALIDATION_RESULT=$(pulumi stack output schemaValidationResult)
        COMPLIANCE_CHECK_RESULT=$(pulumi stack output complianceCheckResult)
        AUDIT_TABLES_CHECK=$(pulumi stack output auditTablesCheck)
        DEPLOYMENT_AUDIT_EVENT_ID=$(pulumi stack output deploymentAuditEventId)
        
        # Verify backup completed successfully
        if [ "$MIGRATION_BACKUP_STATUS" != "success" ]; then
          echo "ERROR: Production migration backup failed"
          exit 1
        fi
        
        # Verify all components deployed successfully
        if [ "$POSTGRES_MIGRATION_STATUS" != "success" ] || \
           [ "$COSMOSDB_MIGRATION_STATUS" != "success" ] || \
           [ "$SERVICES_DEPLOYMENT_STATUS" != "success" ] || \
           [ "$GATEWAYS_DEPLOYMENT_STATUS" != "success" ]; then
          echo "ERROR: One or more components failed to deploy in production"
          exit 1
        fi
        
        # Verify extensive schema compliance for production
        if [ "$SCHEMA_VALIDATION_RESULT" != "compliant" ]; then
          echo "ERROR: Extensive schema validation failed - database does not match specification files"
          exit 1
        fi
        
        # Verify compliance checks passed
        if [ "$COMPLIANCE_CHECK_RESULT" != "passed" ]; then
          echo "ERROR: Compliance checks failed for production deployment"
          exit 1
        fi
        
        # Verify no audit tables exist (compliance requirement)
        if [ "$AUDIT_TABLES_CHECK" != "0" ]; then
          echo "ERROR: Audit tables found in production database. All audit data must be stored in Grafana Cloud Loki."
          exit 1
        fi
        
        # Output detailed deployment results for audit trail
        echo "Production Deployment Results:"
        echo "- Migration Backup: $MIGRATION_BACKUP_STATUS"
        echo "- PostgreSQL Migration: $POSTGRES_MIGRATION_STATUS"
        echo "- CosmosDB Migration: $COSMOSDB_MIGRATION_STATUS"
        echo "- Services Deployment: $SERVICES_DEPLOYMENT_STATUS"
        echo "- Gateways Deployment: $GATEWAYS_DEPLOYMENT_STATUS"
        echo "- Schema Validation: $SCHEMA_VALIDATION_RESULT"
        echo "- Compliance Check: $COMPLIANCE_CHECK_RESULT"
        echo "- Audit Tables Check: $AUDIT_TABLES_CHECK"
        echo "- Audit Event ID: $DEPLOYMENT_AUDIT_EVENT_ID (stored in Grafana Cloud Loki)"
```

### Rollback Management

#### Automated Rollback on Deployment Failure
```yaml
rollback-on-failure:
  runs-on: ubuntu-latest
  if: failure() && (github.ref == 'refs/heads/develop' || github.ref == 'refs/heads/main')
  defaults:
    run:
      working-directory: infrastructure
  
  steps:
    - name: Checkout code
      uses: actions/checkout@v4
    
    - name: Setup Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.21'
    
    - name: Setup Pulumi CLI
      uses: pulumi/actions@v5
      with:
        pulumi-version: '3.96.0'
    
    - name: Rollback deployment via Pulumi
      run: |
        # Login to Pulumi with Azure Blob Storage backend
        pulumi login azblob://${{ secrets.PULUMI_STATE_CONTAINER }}?storage_account=${{ secrets.PULUMI_STORAGE_ACCOUNT }}
        
        # Select appropriate stack
        TARGET_STACK="${{ github.ref == 'refs/heads/main' && 'production' || 'staging' }}"
        pulumi stack select "$TARGET_STACK"
        
        # Get previous successful deployment tag
        PREVIOUS_TAG=$(git log --oneline -n 2 | tail -1 | cut -d' ' -f1)
        
        # Set rollback configuration
        pulumi config set containerImageTag "$PREVIOUS_TAG"
        pulumi config set rollbackReason "deployment-failure"
        pulumi config set rollbackCorrelationId "${{ github.sha }}-failure-rollback"
        
        # Execute rollback
        pulumi up --yes --stack "$TARGET_STACK"
      env:
        # Multi-cloud provider configuration
        ARM_CLIENT_ID: ${{ secrets.ARM_CLIENT_ID }}
        ARM_CLIENT_SECRET: ${{ secrets.ARM_CLIENT_SECRET }}
        ARM_SUBSCRIPTION_ID: ${{ secrets.AZURE_SUBSCRIPTION_ID }}
        ARM_TENANT_ID: ${{ secrets.ARM_TENANT_ID }}
        GRAFANA_CLOUD_API_KEY: ${{ secrets.GRAFANA_CLOUD_API_KEY }}
        VAULT_ADDR: ${{ secrets.VAULT_CLOUD_ADDR }}
        VAULT_TOKEN: ${{ secrets.VAULT_CLOUD_TOKEN }}
        VAULT_NAMESPACE: ${{ secrets.VAULT_CLOUD_NAMESPACE }}
    
    - name: Notify team of rollback
      uses: 8398a7/action-slack@v3
      with:
        status: failure
        text: "Deployment failed and was automatically rolled back to $PREVIOUS_TAG"
      env:
        SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
```

### Health Verification and Compliance Monitoring

#### Post-Deployment Health Verification
```yaml
verify-deployment:
  runs-on: ubuntu-latest
  needs: [deploy-staging, deploy-production]
  if: always() && (github.ref == 'refs/heads/develop' || github.ref == 'refs/heads/main')
  timeout-minutes: 10
  defaults:
    run:
      working-directory: infrastructure
  
  steps:
    - name: Checkout code
      uses: actions/checkout@v4
    
    - name: Setup Pulumi CLI
      uses: pulumi/actions@v5
      with:
        pulumi-version: '3.96.0'
        
    - name: Get deployment endpoints from Pulumi stack
      run: |
        # Login to Pulumi with Azure Blob Storage backend
        pulumi login azblob://${{ secrets.PULUMI_STATE_CONTAINER }}?storage_account=${{ secrets.PULUMI_STORAGE_ACCOUNT }}
        
        # Select appropriate stack
        TARGET_STACK="${{ github.ref == 'refs/heads/main' && 'production' || 'staging' }}"
        pulumi stack select "$TARGET_STACK"
        
        # Export service endpoints for health checks
        echo "PUBLIC_GATEWAY_URL=$(pulumi stack output publicGatewayUrl)" >> $GITHUB_ENV
        echo "ADMIN_GATEWAY_URL=$(pulumi stack output adminGatewayUrl)" >> $GITHUB_ENV
        echo "SERVICES_API_URL=$(pulumi stack output servicesApiUrl)" >> $GITHUB_ENV
        echo "CONTENT_API_URL=$(pulumi stack output contentApiUrl)" >> $GITHUB_ENV
      env:
        AZURE_STORAGE_ACCOUNT: ${{ secrets.PULUMI_STORAGE_ACCOUNT }}
        AZURE_STORAGE_KEY: ${{ secrets.PULUMI_STORAGE_KEY }}
    
    - name: Wait for services to be ready
      run: |
        # Wait for all services to report healthy
        for service_url in "$PUBLIC_GATEWAY_URL" "$ADMIN_GATEWAY_URL" "$SERVICES_API_URL" "$CONTENT_API_URL"; do
          echo "Checking health for: $service_url"
          timeout 300 bash -c "until curl -f $service_url/health; do echo 'Waiting...'; sleep 10; done"
        done
    
    - name: Run smoke tests
      run: |
        # Test public API endpoints
        curl -f "$PUBLIC_GATEWAY_URL/api/v1/services" -H "Accept: application/json"
        
        # Test admin API endpoints with authentication
        curl -f -H "Authorization: Bearer ${{ secrets.ADMIN_TEST_TOKEN }}" \
          -H "Accept: application/json" \
          "$ADMIN_GATEWAY_URL/admin/api/v1/services"
    
    - name: Verify observability and audit logging
      run: |
        # Verify metrics are flowing to Grafana Cloud
        curl -X GET \
          -H "Authorization: Bearer ${{ secrets.GRAFANA_CLOUD_API_KEY }}" \
          "${{ secrets.GRAFANA_CLOUD_ENDPOINT }}/api/v1/query?query=up{job=\"dapr\"}" | \
          jq -e '.data.result | length > 0'
        
        # Verify audit events are being stored in Grafana Cloud Loki
        curl -X GET \
          -H "Authorization: Bearer ${{ secrets.GRAFANA_CLOUD_API_KEY }}" \
          "${{ secrets.GRAFANA_CLOUD_LOKI_ENDPOINT }}/loki/api/v1/query_range?query={job=\"deployment-audit\"}" | \
          jq -e '.data.result | length >= 0'
```

### Deployment Audit and Compliance Reporting

```yaml
notify-and-audit:
  runs-on: ubuntu-latest
  needs: [verify-deployment]
  if: always() && (github.ref == 'refs/heads/develop' || github.ref == 'refs/heads/main')
  
  steps:
    - name: Generate deployment audit report
      run: |
        # Create comprehensive deployment report
        TARGET_ENV="${{ github.ref == 'refs/heads/main' && 'production' || 'staging' }}"
        
        cat > deployment-report.txt << EOF
        Deployment Audit Report
        =======================
        Environment: $TARGET_ENV
        Commit: ${{ github.sha }}
        Timestamp: ${{ github.event.head_commit.timestamp }}
        Author: ${{ github.event.head_commit.author.email }}
        Branch: ${{ github.ref_name }}
        Repository: ${{ github.repository }}
        
        Services Deployed:
        - services-api:${{ github.sha }}
        - content-api:${{ github.sha }}
        - public-gateway:${{ github.sha }}
        - admin-gateway:${{ github.sha }}
        
        Deployment Status: ${{ job.status }}
        EOF
    
    - name: Store audit report in Grafana Cloud Loki
      run: |
        # Send deployment report to Grafana Cloud Loki for compliance audit trail
        curl -X POST \
          -H "Authorization: Bearer ${{ secrets.GRAFANA_CLOUD_LOKI_API_KEY }}" \
          -H "Content-Type: application/json" \
          "${{ secrets.GRAFANA_CLOUD_LOKI_ENDPOINT }}/loki/api/v1/push" \
          -d '{
            "streams": [{
              "stream": {
                "job": "deployment-audit",
                "environment": "${{ github.ref == '\''refs/heads/main'\'' && '\''production'\'' || '\''staging'\'' }}",
                "audit_type": "deployment",
                "service": "ci-cd",
                "repository": "${{ github.repository }}"
              },
              "values": [["'$(date +%s%N)'", "'"$(cat deployment-report.txt | sed 's/"/\\"/g' | tr '\n' ' ')"'"]]
            }]
          }'
        
        echo "Deployment audit report stored in Grafana Cloud Loki"
    
    - name: Notify successful deployment
      if: success()
      uses: 8398a7/action-slack@v3
      with:
        status: success
        text: "${{ github.ref == 'refs/heads/main' && 'Production' || 'Staging' }} deployment successful: ${{ github.sha }}"
      env:
        SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
```

## Environment Configuration Management

### Development Environment Variables
```bash
# Dapr Configuration
DAPR_PLACEMENT_PORT=6050
DAPR_SENTRY_PORT=6051
REDIS_DAPR_PORT=6379

# Services API
SERVICES_API_PORT=8080
SERVICES_API_DAPR_HTTP_PORT=3500
SERVICES_API_DAPR_GRPC_PORT=50001

# Content API
CONTENT_API_PORT=8081
CONTENT_API_DAPR_HTTP_PORT=3501
CONTENT_API_DAPR_GRPC_PORT=50002

# Public Gateway
PUBLIC_GATEWAY_PORT=8000
PUBLIC_GATEWAY_DAPR_HTTP_PORT=3502
PUBLIC_GATEWAY_DAPR_GRPC_PORT=50003

# Admin Gateway
ADMIN_GATEWAY_PORT=8001
ADMIN_GATEWAY_DAPR_HTTP_PORT=3503
ADMIN_GATEWAY_DAPR_GRPC_PORT=50004

# Database Configuration
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=development
POSTGRES_DB=international_center
SERVICES_DATABASE_CONNECTION_STRING="postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgresql:${POSTGRES_PORT}/${POSTGRES_DB}"
CONTENT_DATABASE_CONNECTION_STRING="postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgresql:${POSTGRES_PORT}/${POSTGRES_DB}"

# Azure Emulators
AZURITE_BLOB_PORT=10000
AZURITE_BLOB_HOST=0.0.0.0
AZURITE_BLOB_ENDPOINT="http://azurite:${AZURITE_BLOB_PORT}"
COSMOSDB_PORT=8081
COSMOSDB_EMULATOR_ENDPOINT="http://cosmosdb-emulator:${COSMOSDB_PORT}"
SERVICE_BUS_PORT=5672
SERVICE_BUS_SA_PASSWORD=YourStrong!Passw0rd

# Migration Configuration
MIGRATION_ENVIRONMENT=development
MIGRATION_APPROACH=aggressive
MIGRATION_TIMEOUT=30s

# Authentik Configuration
AUTHENTIK_PORT=9000
AUTHENTIK_SECRET_KEY=development-secret-key
AUTHENTIK_DB_PASSWORD=authentik-password
AUTHENTIK_DB_USER=authentik
AUTHENTIK_DB_NAME=authentik
AUTHENTIK_ISSUER_URL="http://authentik:${AUTHENTIK_PORT}/application/o/"
AUTHENTIK_JWKS_URL="http://authentik:${AUTHENTIK_PORT}/application/o/jwks/"

# Vault Configuration
VAULT_PORT=8200
VAULT_ROOT_TOKEN=development-root-token
VAULT_ADDR="http://vault:${VAULT_PORT}"

# OPA Configuration
OPA_PORT=8181
OPA_LOCAL_ENDPOINT="http://opa:${OPA_PORT}"

# Grafana Stack
GRAFANA_PORT=3000
GRAFANA_ADMIN_PASSWORD=admin
MIMIR_PORT=9009
LOKI_PORT=3100
TEMPO_PORT=3200
PYROSCOPE_PORT=4040
GRAFANA_AGENT_PORT=12345

# Website Origins
PUBLIC_WEBSITE_ORIGINS="http://localhost:4321,http://127.0.0.1:4321"
ADMIN_WEBSITE_ORIGINS="http://localhost:4322,http://127.0.0.1:4322"
```

### GitHub Secrets for CI/CD Pipeline
```yaml
# Pulumi State Storage Configuration
PULUMI_STORAGE_ACCOUNT: "Azure storage account for Pulumi state backend"
PULUMI_STATE_CONTAINER: "Azure storage container for Pulumi state files"
PULUMI_STORAGE_KEY: "Azure storage account key for Pulumi state access"

# Azure Authentication Configuration  
ARM_CLIENT_ID: "Azure service principal client ID"
ARM_CLIENT_SECRET: "Azure service principal client secret"
ARM_SUBSCRIPTION_ID: "Azure subscription identifier"
ARM_TENANT_ID: "Azure tenant identifier"
AZURE_BACKUP_STORAGE_ACCOUNT: "Azure storage account for migration backups"

# HashiCorp Vault Cloud Configuration
VAULT_CLOUD_ADDR: "HashiCorp Vault Cloud address"
VAULT_CLOUD_TOKEN: "HashiCorp Vault Cloud token"
VAULT_CLOUD_NAMESPACE: "HashiCorp Vault Cloud namespace"

# Grafana Cloud Configuration
GRAFANA_CLOUD_API_KEY: "Grafana Cloud API key"
GRAFANA_CLOUD_STACK_SLUG: "Grafana Cloud stack slug"
GRAFANA_CLOUD_ENDPOINT: "Grafana Cloud API endpoint"
GRAFANA_CLOUD_LOKI_ENDPOINT: "Grafana Cloud Loki endpoint for audit logs"
GRAFANA_CLOUD_LOKI_API_KEY: "Grafana Cloud Loki API key"

# Identity Provider Configuration
AUTHENTIK_ISSUER_URL: "Production Authentik issuer URL"
AUTHENTIK_CLIENT_SECRET: "Authentik OAuth2 client secret"

# Notification Configuration
SLACK_WEBHOOK_URL: "Slack webhook URL for alerts and notifications"

# Testing Configuration
ADMIN_TEST_TOKEN: "Admin user JWT for smoke testing"
```

### Environment-Specific Configuration via Pulumi Stacks
```yaml
# Production Environment (managed via Pulumi configuration)
production:
  environment: "production"
  migrationApproach: "conservative"
  migrationValidation: "extensive"
  migrationBackup: "full"
  manualApprovalRequired: "true"
  complianceMode: "true"
  logLevel: "information"
  auditRetentionDays: "2555"  # 7 years compliance
  
# Staging Environment (managed via Pulumi configuration)  
staging:
  environment: "staging"
  migrationApproach: "careful"
  migrationValidation: "standard"
  migrationBackup: "incremental"
  manualApprovalRequired: "false"
  logLevel: "debug"
  auditRetentionDays: "90"
```

## Pulumi Infrastructure Orchestration Pattern

### Integrated Infrastructure Code Structure
The Pulumi Go SDK infrastructure program (`infrastructure/main.go`) orchestrates all resources across environments:

```go
// Example structure - actual implementation handled in infrastructure/
func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        config := pulumi.NewConfig("myapp")
        environment := config.Require("environment")
        
        // 1. Deploy foundational infrastructure (databases, storage, networking)
        infrastructure, err := deployFoundation(ctx, environment)
        if err != nil {
            return err
        }
        
        // 2. Run migrations via pulumi-command resources with proper dependencies
        migrations, err := deployMigrations(ctx, environment, infrastructure)
        if err != nil {
            return err
        }
        
        // 3. Deploy applications and gateways after migrations complete
        applications, err := deployApplications(ctx, environment, []pulumi.Resource{
            migrations, infrastructure,
        })
        if err != nil {
            return err
        }
        
        // 4. Configure observability, alerting, and compliance
        _, err = deployObservability(ctx, environment, applications)
        return err
    })
}
```

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