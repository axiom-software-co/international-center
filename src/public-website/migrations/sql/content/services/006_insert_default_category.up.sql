-- Insert default unassigned category as required by TABLES-SERVICES.md business rules
INSERT INTO service_categories (
    name,
    slug,
    order_number,
    is_default_unassigned,
    created_by
) VALUES (
    'Unassigned Services',
    'unassigned',
    999999,
    TRUE,
    'system'
);