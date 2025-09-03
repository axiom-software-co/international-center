-- Remove seed content storage backends data

DELETE FROM content_storage_backend WHERE created_by = 'migration-seed';