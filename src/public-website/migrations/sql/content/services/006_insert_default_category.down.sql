-- Remove default unassigned category
DELETE FROM service_categories WHERE slug = 'unassigned' AND is_default_unassigned = TRUE;