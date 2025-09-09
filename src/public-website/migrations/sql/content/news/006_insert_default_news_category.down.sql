-- Remove default unassigned news category
DELETE FROM news_categories WHERE slug = 'unassigned' AND is_default_unassigned = TRUE;