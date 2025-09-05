-- Services Domain Seed Data
-- Initial data required for services domain functionality

-- Insert default unassigned category (required by business rules)
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
    true,
    'system'
);

-- Insert additional default categories
INSERT INTO service_categories (
    name,
    slug,
    order_number,
    is_default_unassigned,
    created_by
) VALUES 
(
    'Primary Care',
    'primary-care',
    1,
    false,
    'system'
),
(
    'Specialty Care',
    'specialty-care',
    2,
    false,
    'system'
),
(
    'Emergency Services',
    'emergency-services',
    3,
    false,
    'system'
),
(
    'Preventive Care',
    'preventive-care',
    4,
    false,
    'system'
);

-- Set featured categories (positions 1 and 2)
INSERT INTO featured_categories (
    category_id,
    feature_position,
    created_by
) VALUES 
(
    (SELECT category_id FROM service_categories WHERE slug = 'primary-care' LIMIT 1),
    1,
    'system'
),
(
    (SELECT category_id FROM service_categories WHERE slug = 'emergency-services' LIMIT 1),
    2,
    'system'
);