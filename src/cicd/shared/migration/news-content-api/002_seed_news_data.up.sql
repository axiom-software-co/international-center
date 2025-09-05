-- News Domain Seed Data
-- Initial news categories and sample news data

-- Insert default news categories
INSERT INTO news_categories (category_id, name, slug, description, is_default_unassigned, created_on, created_by) VALUES
    (gen_random_uuid(), 'General', 'general', 'General news and announcements', TRUE, NOW(), 'system'),
    (gen_random_uuid(), 'Press Releases', 'press-releases', 'Official press releases and media announcements', FALSE, NOW(), 'system'),
    (gen_random_uuid(), 'Events', 'events', 'Upcoming events and event announcements', FALSE, NOW(), 'system'),
    (gen_random_uuid(), 'Updates', 'updates', 'System updates and feature announcements', FALSE, NOW(), 'system'),
    (gen_random_uuid(), 'Alerts', 'alerts', 'Important alerts and urgent notifications', FALSE, NOW(), 'system');

-- Insert sample news articles for testing
INSERT INTO news (
    news_id, 
    title, 
    summary, 
    content, 
    slug, 
    category_id, 
    author_name, 
    publishing_status, 
    news_type, 
    priority_level,
    publication_timestamp,
    created_on, 
    created_by
) SELECT 
    gen_random_uuid(),
    'Welcome to International Center News',
    'We are excited to launch our new news platform to keep you updated with the latest announcements and developments.',
    'This is the inaugural post for our new news platform. We will be using this space to share important updates, announcements, and news related to our international center activities. Stay tuned for more exciting content coming soon.',
    'welcome-to-international-center-news',
    category_id,
    'International Center Team',
    'published',
    'announcement',
    'normal',
    NOW(),
    NOW(),
    'system'
FROM news_categories 
WHERE slug = 'general';

-- Insert a sample event announcement
INSERT INTO news (
    news_id, 
    title, 
    summary, 
    content, 
    slug, 
    category_id, 
    author_name, 
    publishing_status, 
    news_type, 
    priority_level,
    publication_timestamp,
    tags,
    created_on, 
    created_by
) SELECT 
    gen_random_uuid(),
    'Upcoming International Conference 2024',
    'Join us for our annual international conference featuring speakers from around the globe.',
    'Our annual international conference brings together thought leaders, innovators, and professionals from various fields. This year''s theme focuses on global collaboration and sustainable development. The event will feature keynote speakers, panel discussions, and networking opportunities. Registration opens next month.',
    'upcoming-international-conference-2024',
    category_id,
    'Events Team',
    'published',
    'event',
    'high',
    NOW() + INTERVAL '1 day',
    ARRAY['conference', 'international', '2024', 'networking'],
    NOW(),
    'system'
FROM news_categories 
WHERE slug = 'events';

-- Insert a sample system update
INSERT INTO news (
    news_id, 
    title, 
    summary, 
    content, 
    slug, 
    category_id, 
    author_name, 
    publishing_status, 
    news_type, 
    priority_level,
    publication_timestamp,
    tags,
    created_on, 
    created_by
) SELECT 
    gen_random_uuid(),
    'Platform Maintenance Scheduled',
    'Scheduled maintenance will occur this weekend to improve system performance and security.',
    'We will be performing routine maintenance on our platform this weekend from Saturday 2:00 AM to Sunday 6:00 AM UTC. During this time, some services may be temporarily unavailable. We apologize for any inconvenience and appreciate your patience as we work to improve your experience.',
    'platform-maintenance-scheduled',
    category_id,
    'Technical Team',
    'published',
    'update',
    'normal',
    NOW() + INTERVAL '2 days',
    ARRAY['maintenance', 'system', 'downtime'],
    NOW(),
    'system'
FROM news_categories 
WHERE slug = 'updates';

-- Make the welcome news article featured
INSERT INTO featured_news (featured_news_id, news_id, created_on, created_by) 
SELECT 
    gen_random_uuid(),
    news_id,
    NOW(),
    'system'
FROM news 
WHERE slug = 'welcome-to-international-center-news';