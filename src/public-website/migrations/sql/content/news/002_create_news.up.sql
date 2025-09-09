-- Create news table matching TABLES-NEWS.md specification
CREATE TABLE news (
    news_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    summary TEXT NOT NULL,
    content TEXT, -- News article content stored in PostgreSQL
    slug VARCHAR(255) UNIQUE NOT NULL,
    category_id UUID NOT NULL REFERENCES news_categories(category_id),
    image_url VARCHAR(500), -- URL to Azure Blob Storage for images only
    author_name VARCHAR(255),
    publication_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    external_source VARCHAR(255),
    external_url VARCHAR(500),
    publishing_status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (publishing_status IN ('draft', 'published', 'archived')),
    
    -- Content metadata
    tags TEXT[],
    news_type VARCHAR(50) NOT NULL CHECK (news_type IN ('announcement', 'press_release', 'event', 'update', 'alert', 'feature')),
    priority_level VARCHAR(20) NOT NULL DEFAULT 'normal' CHECK (priority_level IN ('low', 'normal', 'high', 'urgent')),
    
    -- Audit fields
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    modified_on TIMESTAMPTZ,
    modified_by VARCHAR(255),
    
    -- Soft delete fields
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_on TIMESTAMPTZ,
    deleted_by VARCHAR(255)
);