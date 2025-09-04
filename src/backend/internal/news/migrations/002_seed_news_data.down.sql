-- News Domain Seed Data Rollback
-- Removes all seeded news data

-- Remove featured news entries
DELETE FROM featured_news WHERE created_by = 'system';

-- Remove sample news articles
DELETE FROM news WHERE created_by = 'system';

-- Remove default news categories
DELETE FROM news_categories WHERE created_by = 'system';