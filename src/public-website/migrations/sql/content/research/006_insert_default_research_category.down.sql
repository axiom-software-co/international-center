-- Remove default unassigned research category
DELETE FROM research_categories WHERE slug = 'unassigned' AND is_default_unassigned = TRUE;