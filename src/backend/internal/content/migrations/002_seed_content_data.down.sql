-- Rollback Content Domain Seed Data
-- Remove all seed data inserted during migration

-- Remove default storage backends
DELETE FROM content_storage_backend WHERE backend_name IN (
    'azure-blob-production',
    'azure-blob-development',
    'local-filesystem-fallback'
);