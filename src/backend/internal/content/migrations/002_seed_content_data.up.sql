-- Content Domain Seed Data
-- Initial data required for content domain functionality

-- Insert default storage backends
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
) VALUES 
(
    'azure-blob-production',
    'azure-blob',
    true,
    1,
    'https://production.blob.core.windows.net',
    'azure-blob-storage-key',
    '{"container_name": "content", "connection_timeout": 30}',
    'unknown',
    'system'
),
(
    'azure-blob-development',
    'azure-blob',
    true,
    2,
    'http://localhost:10000/devstoreaccount1',
    'azurite-development-key',
    '{"container_name": "content", "connection_timeout": 10}',
    'unknown',
    'system'
),
(
    'local-filesystem-fallback',
    'local-filesystem',
    false,
    3,
    '/tmp/content-storage',
    null,
    '{"base_directory": "/tmp/content-storage", "create_directories": true}',
    'unknown',
    'system'
);