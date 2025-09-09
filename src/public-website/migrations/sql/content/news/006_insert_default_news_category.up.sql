-- Insert default unassigned news category as required by TABLES-NEWS.md business rules
INSERT INTO news_categories (
    name,
    slug,
    description,
    is_default_unassigned,
    created_by
) VALUES (
    'Unassigned News',
    'unassigned',
    'Default category for news articles that have not been assigned to a specific category',
    TRUE,
    'system'
);