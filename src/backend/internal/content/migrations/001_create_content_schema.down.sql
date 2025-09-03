-- Rollback Content Domain Database Schema
-- Drop in reverse order of creation to handle dependencies

-- Drop triggers first
DROP TRIGGER IF EXISTS storage_health_monitoring_trigger ON content_storage_backend;
DROP TRIGGER IF EXISTS content_upload_processing_trigger ON content;
DROP TRIGGER IF EXISTS content_storage_backend_audit_trigger ON content_storage_backend;
DROP TRIGGER IF EXISTS content_audit_trigger ON content;

-- Drop functions
DROP FUNCTION IF EXISTS update_storage_health_with_notifications();
DROP FUNCTION IF EXISTS publish_content_audit_event_to_grafana_loki();
DROP FUNCTION IF EXISTS process_content_upload_with_dapr_events();

-- Drop indexes
DROP INDEX IF EXISTS idx_storage_backend_health;
DROP INDEX IF EXISTS idx_storage_backend_priority;
DROP INDEX IF EXISTS idx_storage_backend_active;
DROP INDEX IF EXISTS idx_storage_backend_type;
DROP INDEX IF EXISTS idx_virus_scan_correlation;
DROP INDEX IF EXISTS idx_virus_scan_status;
DROP INDEX IF EXISTS idx_virus_scan_timestamp;
DROP INDEX IF EXISTS idx_virus_scan_content_id;
DROP INDEX IF EXISTS idx_access_log_cache_performance;
DROP INDEX IF EXISTS idx_access_log_correlation;
DROP INDEX IF EXISTS idx_access_log_access_type;
DROP INDEX IF EXISTS idx_access_log_client_ip;
DROP INDEX IF EXISTS idx_access_log_user_id;
DROP INDEX IF EXISTS idx_access_log_timestamp;
DROP INDEX IF EXISTS idx_access_log_content_id;
DROP INDEX IF EXISTS idx_content_file_size;
DROP INDEX IF EXISTS idx_content_created_on;
DROP INDEX IF EXISTS idx_content_upload_correlation;
DROP INDEX IF EXISTS idx_content_storage_path;
DROP INDEX IF EXISTS idx_content_upload_status;
DROP INDEX IF EXISTS idx_content_access_level;
DROP INDEX IF EXISTS idx_content_category;
DROP INDEX IF EXISTS idx_content_mime_type;
DROP INDEX IF EXISTS idx_content_hash;

-- Drop tables in dependency order
DROP TABLE IF EXISTS content_virus_scan;
DROP TABLE IF EXISTS content_access_log;
DROP TABLE IF EXISTS content_storage_backend;
DROP TABLE IF EXISTS content;