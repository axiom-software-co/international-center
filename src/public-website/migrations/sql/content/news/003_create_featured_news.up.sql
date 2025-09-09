-- Create featured_news table matching TABLES-NEWS.md specification
CREATE TABLE featured_news (
    featured_news_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    news_id UUID NOT NULL REFERENCES news(news_id),
    
    -- Audit fields
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    modified_on TIMESTAMPTZ,
    modified_by VARCHAR(255),
    
    CONSTRAINT only_one_featured_news CHECK (
        (SELECT COUNT(*) FROM featured_news) <= 1
    ),
    CONSTRAINT no_default_unassigned_featured CHECK (
        NOT EXISTS (
            SELECT 1 FROM news n
            JOIN news_categories nc ON n.category_id = nc.category_id
            WHERE n.news_id = featured_news.news_id 
            AND nc.is_default_unassigned = TRUE
        )
    )
);