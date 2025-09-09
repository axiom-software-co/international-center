-- Create featured_research table matching TABLES-RESEARCH.md specification
CREATE TABLE featured_research (
    featured_research_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    research_id UUID NOT NULL REFERENCES research(research_id),
    
    -- Audit fields
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    modified_on TIMESTAMPTZ,
    modified_by VARCHAR(255),
    
    CONSTRAINT only_one_featured_research CHECK (
        (SELECT COUNT(*) FROM featured_research) <= 1
    ),
    CONSTRAINT no_default_unassigned_featured CHECK (
        NOT EXISTS (
            SELECT 1 FROM research r
            JOIN research_categories rc ON r.category_id = rc.category_id
            WHERE r.research_id = featured_research.research_id 
            AND rc.is_default_unassigned = TRUE
        )
    )
);