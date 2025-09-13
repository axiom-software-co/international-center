-- Create featured_news table matching TABLES-NEWS.md specification
CREATE TABLE featured_news (
    featured_news_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    news_id UUID NOT NULL REFERENCES news(news_id),
    
    -- Audit fields
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    modified_on TIMESTAMPTZ,
    modified_by VARCHAR(255),
    
    -- TODO: Implement constraint validation using database triggers or application logic
    -- Business rule: only one featured news item allowed
    -- Business rule: featured news cannot reference default unassigned categories
);