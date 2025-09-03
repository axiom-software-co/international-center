-- Remove seed service categories data

-- Remove featured categories first (due to foreign key constraints)
DELETE FROM featured_categories WHERE created_by = 'migration-seed';

-- Remove service categories (maintaining referential integrity)
DELETE FROM service_categories WHERE created_by = 'migration-seed';