-- Research Domain Seed Data
-- Initial data required for research domain functionality

-- Insert default unassigned research category (required by business rules)
INSERT INTO research_categories (
    category_id,
    name,
    slug,
    description,
    is_default_unassigned,
    created_by
) VALUES (
    gen_random_uuid(),
    'Unassigned',
    'unassigned',
    'Default category for research articles that have not been assigned to a specific category. This category is required by the system and cannot be deleted.',
    true,
    'system'
);

-- Insert common research categories for initial setup
INSERT INTO research_categories (
    category_id,
    name,
    slug,
    description,
    is_default_unassigned,
    created_by
) VALUES 
(
    gen_random_uuid(),
    'Clinical Studies',
    'clinical-studies',
    'Research articles focused on clinical trials and studies involving human subjects',
    false,
    'system'
),
(
    gen_random_uuid(),
    'Case Reports',
    'case-reports', 
    'Detailed reports of individual patient cases or case series',
    false,
    'system'
),
(
    gen_random_uuid(),
    'Systematic Reviews',
    'systematic-reviews',
    'Comprehensive reviews of existing research literature on specific topics',
    false,
    'system'
),
(
    gen_random_uuid(),
    'Meta-Analysis',
    'meta-analysis',
    'Statistical analysis combining results from multiple independent studies',
    false,
    'system'
),
(
    gen_random_uuid(),
    'Editorial',
    'editorial',
    'Editorial articles, commentary, and opinion pieces',
    false,
    'system'
),
(
    gen_random_uuid(),
    'Commentary',
    'commentary',
    'Expert commentary and perspective articles',
    false,
    'system'
);