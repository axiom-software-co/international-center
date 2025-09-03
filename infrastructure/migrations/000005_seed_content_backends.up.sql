-- Seed content storage backends - Required for content domain functionality
-- Per CONTENT-TABLES.md schema requirements

-- Insert Azure Blob Storage backend for production
INSERT INTO content_storage_backend (
    backend_name,
    backend_type,
    is_active,
    priority_order,
    base_url,
    access_key_vault_reference,
    configuration_json,
    health_status,
    created_by
) VALUES (
    'azure-blob-production',
    'azure-blob',
    TRUE,
    1,
    'https://internationalcenter.blob.core.windows.net',
    'vault:azure-blob-access-key',
    '{"container_name": "content", "enable_cdn": false}',
    'unknown',
    'migration-seed'
);

-- Insert Azurite emulator backend for local development
INSERT INTO content_storage_backend (
    backend_name,
    backend_type,
    is_active,
    priority_order,
    base_url,
    access_key_vault_reference,
    configuration_json,
    health_status,
    created_by
) VALUES (
    'azurite-local-development',
    'azure-blob',
    TRUE,
    2,
    'http://azurite:10000/devstoreaccount1',
    'vault:azurite-access-key',
    '{"container_name": "content", "enable_cdn": false, "development_mode": true}',
    'unknown',
    'migration-seed'
);