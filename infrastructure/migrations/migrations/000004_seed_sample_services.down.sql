-- Remove seed sample services data

DELETE FROM services WHERE created_by = 'migration-seed';