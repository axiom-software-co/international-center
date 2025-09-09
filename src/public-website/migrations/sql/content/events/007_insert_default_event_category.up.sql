-- Insert default unassigned event category matching domain pattern
INSERT INTO event_categories (
    name,
    slug,
    description,
    is_default_unassigned,
    created_by
) VALUES (
    'Unassigned',
    'unassigned',
    'Default category for events that have not been assigned to a specific category',
    TRUE,
    'system'
);