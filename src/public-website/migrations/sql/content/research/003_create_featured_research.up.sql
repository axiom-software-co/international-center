-- Create featured_research table matching TABLES-RESEARCH.md specification
CREATE TABLE featured_research (
    featured_research_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    research_id UUID NOT NULL REFERENCES research(research_id),
    
    -- Audit fields
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    modified_on TIMESTAMPTZ,
    modified_by VARCHAR(255),
    
    -- TODO: Implement constraint validation using database triggers or application logic
    -- Business rule: only one featured research item allowed
    -- Business rule: featured research cannot reference default unassigned categories
);