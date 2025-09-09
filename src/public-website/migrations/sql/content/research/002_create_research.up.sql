-- Create research table matching TABLES-RESEARCH.md specification
CREATE TABLE research (
    research_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    abstract TEXT NOT NULL,
    content TEXT, -- Research article content stored in PostgreSQL
    slug VARCHAR(255) UNIQUE NOT NULL,
    category_id UUID NOT NULL REFERENCES research_categories(category_id),
    image_url VARCHAR(500), -- URL to Azure Blob Storage for images only
    author_names VARCHAR(500) NOT NULL,
    publication_date DATE,
    doi VARCHAR(100),
    external_url VARCHAR(500),
    report_url VARCHAR(500), -- URL to PDF publication report in Azure Blob Storage
    publishing_status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (publishing_status IN ('draft', 'published', 'archived')),
    
    -- Content metadata
    keywords TEXT[],
    research_type VARCHAR(50) NOT NULL CHECK (research_type IN ('clinical_study', 'case_report', 'systematic_review', 'meta_analysis', 'editorial', 'commentary')),
    
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