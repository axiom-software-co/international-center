-- Research Domain Seed Data Rollback
-- Removes all seed data for research domain

-- Remove all system-created research categories
DELETE FROM research_categories WHERE created_by = 'system';

-- Alternative approach for more specific cleanup if needed:
-- DELETE FROM research_categories WHERE slug IN (
--     'unassigned', 
--     'clinical-studies', 
--     'case-reports', 
--     'systematic-reviews', 
--     'meta-analysis', 
--     'editorial', 
--     'commentary'
-- ) AND created_by = 'system';