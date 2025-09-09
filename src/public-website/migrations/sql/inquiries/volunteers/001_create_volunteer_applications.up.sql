-- Create volunteer_applications table matching TABLES-VOLUNTEERS.md specification
CREATE TABLE volunteer_applications (
    application_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status VARCHAR(20) NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'under-review', 'interview-scheduled', 'background-check', 'approved', 'declined', 'withdrawn')),
    priority VARCHAR(10) NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    
    -- Personal Information
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    email VARCHAR(254) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    age INTEGER NOT NULL CHECK (age >= 18 AND age <= 100),
    
    -- Volunteer Details
    volunteer_interest VARCHAR(30) NOT NULL CHECK (volunteer_interest IN ('patient-support', 'community-outreach', 'research-support', 'administrative-support', 'multiple', 'other')),
    availability VARCHAR(20) NOT NULL CHECK (availability IN ('2-4-hours', '4-8-hours', '8-16-hours', '16-hours-plus', 'flexible')),
    experience TEXT CHECK (LENGTH(experience) <= 1000),
    motivation TEXT NOT NULL CHECK (LENGTH(motivation) >= 20 AND LENGTH(motivation) <= 1500),
    schedule_preferences VARCHAR(500),
    
    -- Metadata
    source VARCHAR(50) DEFAULT 'website',
    ip_address INET,
    user_agent TEXT,
    
    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100) NOT NULL DEFAULT 'system',
    updated_by VARCHAR(100) NOT NULL DEFAULT 'system',
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMP WITH TIME ZONE
);