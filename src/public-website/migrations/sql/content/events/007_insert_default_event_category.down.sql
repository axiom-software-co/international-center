-- Remove default unassigned event category
DELETE FROM event_categories WHERE slug = 'unassigned' AND is_default_unassigned = TRUE;