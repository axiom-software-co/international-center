-- Create featured_categories table matching TABLES-SERVICES.md specification
CREATE TABLE featured_categories (
    featured_category_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID NOT NULL REFERENCES service_categories(category_id),
    feature_position INTEGER NOT NULL CHECK (feature_position IN (1, 2)),
    
    -- Audit fields
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    modified_on TIMESTAMPTZ,
    modified_by VARCHAR(255),
    
    UNIQUE(feature_position)
    -- TODO: Implement constraint validation using database triggers or application logic
    -- Business rule: featured categories cannot reference default unassigned categories
);