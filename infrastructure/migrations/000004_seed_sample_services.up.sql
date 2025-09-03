-- Seed sample services - Required for integration testing
-- Services with different delivery modes and publishing statuses per schema

INSERT INTO services (
    title,
    description,
    slug,
    category_id,
    delivery_mode,
    publishing_status,
    order_number,
    created_by
) VALUES 
(
    'Emergency Room Treatment',
    'Comprehensive emergency medical care available 24/7 for urgent health conditions requiring immediate attention.',
    'emergency-room-treatment',
    (SELECT category_id FROM service_categories WHERE slug = 'emergency-services'),
    'inpatient_service',
    'published',
    1,
    'migration-seed'
),
(
    'Mobile Health Screening',
    'Convenient health screening services brought directly to your location with comprehensive diagnostic capabilities.',
    'mobile-health-screening',
    (SELECT category_id FROM service_categories WHERE slug = 'preventive-care'),
    'mobile_service',
    'published',
    1,
    'migration-seed'
),
(
    'Outpatient Surgery Center',
    'Advanced outpatient surgical procedures with same-day discharge for minimally invasive treatments.',
    'outpatient-surgery-center',
    (SELECT category_id FROM service_categories WHERE slug = 'outpatient-care'),
    'outpatient_service',
    'published',
    1,
    'migration-seed'
),
(
    'Specialized Cardiac Care',
    'Comprehensive cardiac treatment including diagnostics, intervention, and ongoing management for heart conditions.',
    'specialized-cardiac-care',
    (SELECT category_id FROM service_categories WHERE slug = 'specialized-treatment'),
    'inpatient_service',
    'draft',
    1,
    'migration-seed'
),
(
    'Community Wellness Program',
    'Community-based wellness and health education program focused on preventive care and health promotion.',
    'community-wellness-program',
    (SELECT category_id FROM service_categories WHERE slug = 'preventive-care'),
    'mobile_service',
    'archived',
    2,
    'migration-seed'
);