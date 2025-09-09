-- Insert default unassigned research category as required by TABLES-RESEARCH.md business rules
INSERT INTO research_categories (
    name,
    slug,
    description,
    is_default_unassigned,
    created_by
) VALUES (
    'Unassigned Research',
    'unassigned',
    'Default category for research articles that have not been assigned to a specific category',
    TRUE,
    'system'
);