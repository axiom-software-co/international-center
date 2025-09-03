-- Seed service categories - Required for schema compliance and testing
-- Exact match to SERVICES-TABLES.md business rules

-- Insert default unassigned category (exactly one required)
INSERT INTO service_categories (
    name, 
    slug, 
    order_number, 
    is_default_unassigned,
    created_by
) VALUES (
    'Unassigned',
    'unassigned', 
    999,
    TRUE,
    'migration-seed'
);

-- Insert primary service categories
INSERT INTO service_categories (
    name, 
    slug, 
    order_number, 
    is_default_unassigned,
    created_by
) VALUES 
(
    'Emergency Services',
    'emergency-services',
    1,
    FALSE,
    'migration-seed'
),
(
    'Outpatient Care',
    'outpatient-care',
    2,
    FALSE,
    'migration-seed'
),
(
    'Specialized Treatment',
    'specialized-treatment',
    3,
    FALSE,
    'migration-seed'
),
(
    'Preventive Care',
    'preventive-care',
    4,
    FALSE,
    'migration-seed'
);

-- Insert featured categories (exactly positions 1 and 2 as per schema rules)
INSERT INTO featured_categories (
    category_id,
    feature_position,
    created_by
) VALUES 
(
    (SELECT category_id FROM service_categories WHERE slug = 'emergency-services'),
    1,
    'migration-seed'
),
(
    (SELECT category_id FROM service_categories WHERE slug = 'outpatient-care'),
    2,
    'migration-seed'
);