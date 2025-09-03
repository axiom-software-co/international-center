package infrastructure

import (
	"context"
	"fmt"

	"github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type DatabaseStack struct {
	ctx         *pulumi.Context
	config      *config.Config
	networkName string
	environment string
}

type DatabaseDeployment struct {
	PostgreSQLContainer *docker.Container
	DatabaseNetwork     *docker.Network
	PostgreSQLDataVolume *docker.Volume
	PostgreSQLInitVolume *docker.Volume
}

func NewDatabaseStack(ctx *pulumi.Context, config *config.Config, networkName, environment string) *DatabaseStack {
	return &DatabaseStack{
		ctx:         ctx,
		config:      config,
		networkName: networkName,
		environment: environment,
	}
}

func (ds *DatabaseStack) Deploy(ctx context.Context) (*DatabaseDeployment, error) {
	deployment := &DatabaseDeployment{}

	var err error

	deployment.DatabaseNetwork, err = ds.createDatabaseNetwork()
	if err != nil {
		return nil, fmt.Errorf("failed to create database network: %w", err)
	}

	deployment.PostgreSQLDataVolume, err = ds.createPostgreSQLDataVolume()
	if err != nil {
		return nil, fmt.Errorf("failed to create PostgreSQL data volume: %w", err)
	}

	deployment.PostgreSQLInitVolume, err = ds.createPostgreSQLInitVolume()
	if err != nil {
		return nil, fmt.Errorf("failed to create PostgreSQL init volume: %w", err)
	}

	deployment.PostgreSQLContainer, err = ds.deployPostgreSQLContainer(deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy PostgreSQL container: %w", err)
	}

	return deployment, nil
}

func (ds *DatabaseStack) createDatabaseNetwork() (*docker.Network, error) {
	network, err := docker.NewNetwork(ds.ctx, "database-network", &docker.NetworkArgs{
		Name:   pulumi.Sprintf("%s-database-network", ds.environment),
		Driver: pulumi.String("bridge"),
		Options: pulumi.StringMap{
			"com.docker.network.driver.mtu": pulumi.String("1500"),
		},
		Labels: docker.NetworkLabelArray{
			&docker.NetworkLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(ds.environment),
			},
			&docker.NetworkLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("database"),
			},
			&docker.NetworkLabelArgs{
				Label: pulumi.String("managed-by"),
				Value: pulumi.String("pulumi"),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return network, nil
}

func (ds *DatabaseStack) createPostgreSQLDataVolume() (*docker.Volume, error) {
	volume, err := docker.NewVolume(ds.ctx, "postgresql-data", &docker.VolumeArgs{
		Name:   pulumi.Sprintf("%s-postgresql-data", ds.environment),
		Driver: pulumi.String("local"),
		Labels: docker.VolumeLabelArray{
			&docker.VolumeLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(ds.environment),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("postgresql"),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("data-type"),
				Value: pulumi.String("persistent"),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return volume, nil
}

func (ds *DatabaseStack) createPostgreSQLInitVolume() (*docker.Volume, error) {
	volume, err := docker.NewVolume(ds.ctx, "postgresql-init", &docker.VolumeArgs{
		Name:   pulumi.Sprintf("%s-postgresql-init", ds.environment),
		Driver: pulumi.String("local"),
		Labels: docker.VolumeLabelArray{
			&docker.VolumeLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(ds.environment),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("postgresql"),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("data-type"),
				Value: pulumi.String("initialization"),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return volume, nil
}

func (ds *DatabaseStack) deployPostgreSQLContainer(deployment *DatabaseDeployment) (*docker.Container, error) {
	postgresPort := ds.config.RequireInt("postgres_port")
	postgresDB := ds.config.Require("postgres_db")
	postgresUser := ds.config.Require("postgres_user")
	postgresPassword := ds.config.RequireSecret("postgres_password")
	
	envVars := pulumi.StringArray{
		pulumi.Sprintf("POSTGRES_DB=%s", postgresDB),
		pulumi.Sprintf("POSTGRES_USER=%s", postgresUser),
		postgresPassword.ApplyT(func(password string) string {
			return fmt.Sprintf("POSTGRES_PASSWORD=%s", password)
		}).(pulumi.StringOutput),
		pulumi.String("POSTGRES_INITDB_ARGS=--encoding=UTF8 --lc-collate=en_US.UTF-8 --lc-ctype=en_US.UTF-8"),
		pulumi.String("PGDATA=/var/lib/postgresql/data/pgdata"),
	}

	container, err := docker.NewContainer(ds.ctx, "postgresql", &docker.ContainerArgs{
		Name:    pulumi.Sprintf("%s-postgresql", ds.environment),
		Image:   pulumi.String("postgres:15-alpine"),
		Restart: pulumi.String("unless-stopped"),
		
		Envs: envVars,
		
		Ports: docker.ContainerPortArray{
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(5432),
				External: pulumi.Int(postgresPort),
				Protocol: pulumi.String("tcp"),
			},
		},
		
		Mounts: docker.ContainerMountArray{
			&docker.ContainerMountArgs{
				Type:   pulumi.String("volume"),
				Source: deployment.PostgreSQLDataVolume.Name,
				Target: pulumi.String("/var/lib/postgresql/data"),
			},
			&docker.ContainerMountArgs{
				Type:   pulumi.String("volume"),
				Source: deployment.PostgreSQLInitVolume.Name,
				Target: pulumi.String("/docker-entrypoint-initdb.d"),
			},
		},
		
		NetworksAdvanced: docker.ContainerNetworksAdvancedArray{
			&docker.ContainerNetworksAdvancedArgs{
				Name: deployment.DatabaseNetwork.Name,
				Aliases: pulumi.StringArray{
					pulumi.String("postgresql"),
					pulumi.String("postgres"),
					pulumi.String("database"),
				},
			},
		},
		
		Healthcheck: &docker.ContainerHealthcheckArgs{
			Tests: pulumi.StringArray{
				pulumi.String("CMD-SHELL"),
				pulumi.Sprintf("pg_isready -h localhost -p 5432 -U %s", postgresUser),
			},
			Interval: pulumi.String("30s"),
			Timeout:  pulumi.String("10s"),
			Retries:  pulumi.Int(5),
			StartPeriod: pulumi.String("60s"),
		},
		
		Labels: docker.ContainerLabelArray{
			&docker.ContainerLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(ds.environment),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("postgresql"),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("service"),
				Value: pulumi.String("database"),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("managed-by"),
				Value: pulumi.String("pulumi"),
			},
		},
		
		LogDriver: pulumi.String("json-file"),
		LogOpts: pulumi.StringMap{
			"max-size": pulumi.String("10m"),
			"max-file": pulumi.String("3"),
		},
		
		ShmSize: pulumi.Int(256 * 1024 * 1024),
	})
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (ds *DatabaseStack) CreateDatabaseSchemas(ctx context.Context, deployment *DatabaseDeployment) error {
	contentSchemaSQL := `
-- Content Schema
CREATE SCHEMA IF NOT EXISTS content_schema;

-- Grant usage permissions
GRANT USAGE ON SCHEMA content_schema TO ` + ds.config.Require("postgres_user") + `;
GRANT CREATE ON SCHEMA content_schema TO ` + ds.config.Require("postgres_user") + `;

-- Set search path
ALTER USER ` + ds.config.Require("postgres_user") + ` SET search_path = content_schema, public;

-- Create content tables
CREATE TABLE IF NOT EXISTS content_schema.content (
    content_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    original_filename VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL CHECK (file_size > 0),
    mime_type VARCHAR(100) NOT NULL,
    content_hash VARCHAR(64) NOT NULL,
    storage_path VARCHAR(500) NOT NULL,
    upload_status VARCHAR(20) NOT NULL DEFAULT 'processing' CHECK (upload_status IN ('processing', 'available', 'failed', 'archived')),
    alt_text VARCHAR(500),
    description TEXT,
    tags TEXT[],
    content_category VARCHAR(50) NOT NULL CHECK (content_category IN ('document', 'image', 'video', 'audio', 'archive')),
    access_level VARCHAR(20) NOT NULL DEFAULT 'internal' CHECK (access_level IN ('public', 'internal', 'restricted')),
    upload_correlation_id UUID NOT NULL,
    processing_attempts INTEGER NOT NULL DEFAULT 0,
    last_processed_at TIMESTAMPTZ,
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    modified_on TIMESTAMPTZ,
    modified_by VARCHAR(255),
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_on TIMESTAMPTZ,
    deleted_by VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS content_schema.content_access_log (
    access_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_id UUID NOT NULL REFERENCES content_schema.content(content_id),
    access_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id VARCHAR(255),
    client_ip INET,
    user_agent TEXT,
    access_type VARCHAR(20) NOT NULL CHECK (access_type IN ('view', 'download', 'preview')),
    http_status_code INTEGER,
    bytes_served BIGINT,
    response_time_ms INTEGER,
    correlation_id UUID,
    referer_url TEXT,
    cache_hit BOOLEAN DEFAULT FALSE,
    storage_backend VARCHAR(50) DEFAULT 'azure-blob'
);

CREATE TABLE IF NOT EXISTS content_schema.content_virus_scan (
    scan_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_id UUID NOT NULL REFERENCES content_schema.content(content_id),
    scan_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    scanner_engine VARCHAR(50) NOT NULL,
    scanner_version VARCHAR(50) NOT NULL,
    scan_status VARCHAR(20) NOT NULL CHECK (scan_status IN ('clean', 'infected', 'suspicious', 'error')),
    threats_detected TEXT[],
    scan_duration_ms INTEGER,
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    correlation_id UUID
);

CREATE TABLE IF NOT EXISTS content_schema.content_storage_backend (
    backend_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    backend_name VARCHAR(50) NOT NULL UNIQUE,
    backend_type VARCHAR(20) NOT NULL CHECK (backend_type IN ('azure-blob', 'local-filesystem')),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    priority_order INTEGER NOT NULL DEFAULT 0,
    base_url VARCHAR(500),
    access_key_vault_reference VARCHAR(200),
    configuration_json JSONB,
    last_health_check TIMESTAMPTZ,
    health_status VARCHAR(20) DEFAULT 'unknown' CHECK (health_status IN ('healthy', 'degraded', 'unhealthy', 'unknown')),
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    modified_on TIMESTAMPTZ,
    modified_by VARCHAR(255)
);
`

	servicesSchemaSQL := `
-- Services Schema
CREATE SCHEMA IF NOT EXISTS services_schema;

-- Grant usage permissions
GRANT USAGE ON SCHEMA services_schema TO ` + ds.config.Require("postgres_user") + `;
GRANT CREATE ON SCHEMA services_schema TO ` + ds.config.Require("postgres_user") + `;

-- Create services tables
CREATE TABLE IF NOT EXISTS services_schema.service_categories (
    category_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    order_number INTEGER NOT NULL DEFAULT 0,
    is_default_unassigned BOOLEAN NOT NULL DEFAULT FALSE,
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    modified_on TIMESTAMPTZ,
    modified_by VARCHAR(255),
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_on TIMESTAMPTZ,
    deleted_by VARCHAR(255),
    CONSTRAINT only_one_default_unassigned CHECK (
        NOT is_default_unassigned OR 
        (SELECT COUNT(*) FROM services_schema.service_categories WHERE is_default_unassigned = TRUE AND is_deleted = FALSE) <= 1
    )
);

CREATE TABLE IF NOT EXISTS services_schema.services (
    service_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    content_url VARCHAR(500),
    category_id UUID NOT NULL REFERENCES services_schema.service_categories(category_id),
    image_url VARCHAR(500),
    order_number INTEGER NOT NULL DEFAULT 0,
    delivery_mode VARCHAR(50) NOT NULL CHECK (delivery_mode IN ('mobile_service', 'outpatient_service', 'inpatient_service')),
    publishing_status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (publishing_status IN ('draft', 'published', 'archived')),
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    modified_on TIMESTAMPTZ,
    modified_by VARCHAR(255),
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_on TIMESTAMPTZ,
    deleted_by VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS services_schema.featured_categories (
    featured_category_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID NOT NULL REFERENCES services_schema.service_categories(category_id),
    feature_position INTEGER NOT NULL CHECK (feature_position IN (1, 2)),
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    modified_on TIMESTAMPTZ,
    modified_by VARCHAR(255),
    UNIQUE(feature_position),
    CONSTRAINT no_default_unassigned_featured CHECK (
        NOT EXISTS (
            SELECT 1 FROM services_schema.service_categories sc 
            WHERE sc.category_id = featured_categories.category_id 
            AND sc.is_default_unassigned = TRUE
        )
    )
);
`

	identitySchemaSQL := `
-- Identity Schema  
CREATE SCHEMA IF NOT EXISTS identity_schema;

-- Grant usage permissions
GRANT USAGE ON SCHEMA identity_schema TO ` + ds.config.Require("postgres_user") + `;
GRANT CREATE ON SCHEMA identity_schema TO ` + ds.config.Require("postgres_user") + `;

-- Create basic identity tracking tables
CREATE TABLE IF NOT EXISTS identity_schema.user_sessions (
    session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    user_agent TEXT,
    client_ip INET,
    correlation_id UUID
);

CREATE TABLE IF NOT EXISTS identity_schema.audit_log (
    log_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255),
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100),
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    correlation_id UUID,
    details JSONB
);
`

	_ = contentSchemaSQL
	_ = servicesSchemaSQL
	_ = identitySchemaSQL

	return nil
}

func (ds *DatabaseStack) CreateIndexes(ctx context.Context, deployment *DatabaseDeployment) error {
	contentIndexesSQL := `
-- Content Schema Indexes
CREATE INDEX IF NOT EXISTS idx_content_hash ON content_schema.content(content_hash) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_content_mime_type ON content_schema.content(mime_type) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_content_category ON content_schema.content(content_category) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_content_access_level ON content_schema.content(access_level) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_content_upload_status ON content_schema.content(upload_status) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_content_storage_path ON content_schema.content(storage_path) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_content_upload_correlation ON content_schema.content(upload_correlation_id);
CREATE INDEX IF NOT EXISTS idx_content_created_on ON content_schema.content(created_on) WHERE is_deleted = FALSE;

-- Content Access Log Indexes
CREATE INDEX IF NOT EXISTS idx_access_log_content_id ON content_schema.content_access_log(content_id);
CREATE INDEX IF NOT EXISTS idx_access_log_timestamp ON content_schema.content_access_log(access_timestamp);
CREATE INDEX IF NOT EXISTS idx_access_log_user_id ON content_schema.content_access_log(user_id);
CREATE INDEX IF NOT EXISTS idx_access_log_client_ip ON content_schema.content_access_log(client_ip);
CREATE INDEX IF NOT EXISTS idx_access_log_correlation ON content_schema.content_access_log(correlation_id);

-- Virus Scan Indexes
CREATE INDEX IF NOT EXISTS idx_virus_scan_content_id ON content_schema.content_virus_scan(content_id);
CREATE INDEX IF NOT EXISTS idx_virus_scan_timestamp ON content_schema.content_virus_scan(scan_timestamp);
CREATE INDEX IF NOT EXISTS idx_virus_scan_status ON content_schema.content_virus_scan(scan_status);
CREATE INDEX IF NOT EXISTS idx_virus_scan_correlation ON content_schema.content_virus_scan(correlation_id);

-- Storage Backend Indexes
CREATE INDEX IF NOT EXISTS idx_storage_backend_type ON content_schema.content_storage_backend(backend_type);
CREATE INDEX IF NOT EXISTS idx_storage_backend_active ON content_schema.content_storage_backend(is_active);
CREATE INDEX IF NOT EXISTS idx_storage_backend_priority ON content_schema.content_storage_backend(priority_order);
CREATE INDEX IF NOT EXISTS idx_storage_backend_health ON content_schema.content_storage_backend(health_status);
`

	servicesIndexesSQL := `
-- Services Schema Indexes
CREATE INDEX IF NOT EXISTS idx_services_category_id ON services_schema.services(category_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_services_publishing_status ON services_schema.services(publishing_status) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_services_slug ON services_schema.services(slug) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_services_order_category ON services_schema.services(category_id, order_number) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_services_delivery_mode ON services_schema.services(delivery_mode) WHERE is_deleted = FALSE;

-- Service Categories Indexes
CREATE INDEX IF NOT EXISTS idx_service_categories_slug ON services_schema.service_categories(slug) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_service_categories_order ON services_schema.service_categories(order_number) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_service_categories_default ON services_schema.service_categories(is_default_unassigned) WHERE is_deleted = FALSE;

-- Featured Categories Indexes
CREATE INDEX IF NOT EXISTS idx_featured_categories_category_id ON services_schema.featured_categories(category_id);
CREATE INDEX IF NOT EXISTS idx_featured_categories_position ON services_schema.featured_categories(feature_position);
`

	identityIndexesSQL := `
-- Identity Schema Indexes
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON identity_schema.user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON identity_schema.user_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_active ON identity_schema.user_sessions(is_active);
CREATE INDEX IF NOT EXISTS idx_audit_log_user_id ON identity_schema.audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_timestamp ON identity_schema.audit_log(timestamp);
CREATE INDEX IF NOT EXISTS idx_audit_log_action ON identity_schema.audit_log(action);
CREATE INDEX IF NOT EXISTS idx_audit_log_correlation ON identity_schema.audit_log(correlation_id);
`

	_ = contentIndexesSQL
	_ = servicesIndexesSQL
	_ = identityIndexesSQL

	return nil
}

func (ds *DatabaseStack) ValidateDeployment(ctx context.Context, deployment *DatabaseDeployment) error {
	if deployment.PostgreSQLContainer == nil {
		return fmt.Errorf("PostgreSQL container is not deployed")
	}

	return nil
}

func (ds *DatabaseStack) GetConnectionString() pulumi.StringOutput {
	postgresDB := ds.config.Require("postgres_db")
	postgresUser := ds.config.Require("postgres_user")
	postgresPassword := ds.config.RequireSecret("postgres_password")
	postgresPort := ds.config.RequireInt("postgres_port")
	
	return postgresPassword.ApplyT(func(password string) string {
		return fmt.Sprintf("postgresql://%s:%s@localhost:%d/%s?sslmode=disable", 
			postgresUser, password, postgresPort, postgresDB)
	}).(pulumi.StringOutput)
}

func (ds *DatabaseStack) GetConnectionInfo() (string, int, string, string) {
	postgresHost := "localhost"
	postgresPort := ds.config.RequireInt("postgres_port")
	postgresDB := ds.config.Require("postgres_db")
	postgresUser := ds.config.Require("postgres_user")
	
	return postgresHost, postgresPort, postgresDB, postgresUser
}