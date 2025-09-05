-- Rollback Services Domain Seed Data
-- Remove all seed data inserted during migration

-- Remove featured categories first (due to foreign key constraints)
DELETE FROM featured_categories WHERE created_by = 'system';

-- Remove default service categories
DELETE FROM service_categories WHERE slug IN (
    'unassigned',
    'primary-care',
    'specialty-care',
    'emergency-services',
    'preventive-care'
) AND created_by = 'system';