-- Drop audit functions and triggers for news domain

-- Drop triggers
DROP TRIGGER IF EXISTS news_audit_trigger ON news;
DROP TRIGGER IF EXISTS news_categories_audit_trigger ON news_categories;
DROP TRIGGER IF EXISTS featured_news_audit_trigger ON featured_news;
DROP TRIGGER IF EXISTS news_categories_reassign_trigger ON news_categories;
DROP TRIGGER IF EXISTS featured_news_management_trigger ON featured_news;

-- Drop functions
DROP FUNCTION IF EXISTS publish_news_audit_event_to_grafana_loki();
DROP FUNCTION IF EXISTS reassign_news_to_default_category();
DROP FUNCTION IF EXISTS manage_featured_news();