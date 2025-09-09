-- Inquiries Domain Validation Schema - Final Desired State
-- This file represents the authoritative schema state after all inquiry migrations are complete
-- Used for validation against deployed schema, not for direct migration execution

-- Donations Inquiries Table
CREATE TABLE donations_inquiries (
    inquiry_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status VARCHAR(20) NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'acknowledged', 'in_progress', 'resolved', 'closed')),
    priority VARCHAR(10) NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    
    -- Personal Information
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    email VARCHAR(254) NOT NULL,
    phone VARCHAR(20),
    
    -- Donation Details
    donation_type VARCHAR(20) NOT NULL CHECK (donation_type IN ('one-time', 'monthly', 'corporate', 'memorial', 'tribute', 'other')),
    amount_intention VARCHAR(20) CHECK (amount_intention IN ('under-100', '100-500', '500-1000', '1000-plus', 'discuss')),
    message TEXT CHECK (LENGTH(message) <= 1500),
    
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

-- Business Inquiries Table
CREATE TABLE business_inquiries (
    inquiry_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status VARCHAR(20) NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'acknowledged', 'in_progress', 'resolved', 'closed')),
    priority VARCHAR(10) NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    
    -- Organization Information
    organization_name VARCHAR(100) NOT NULL,
    contact_name VARCHAR(50) NOT NULL,
    title VARCHAR(50) NOT NULL,
    email VARCHAR(254) NOT NULL,
    phone VARCHAR(20),
    industry VARCHAR(50),
    
    -- Inquiry Details
    inquiry_type VARCHAR(20) NOT NULL CHECK (inquiry_type IN ('partnership', 'licensing', 'research', 'technology', 'regulatory', 'other')),
    message TEXT NOT NULL CHECK (LENGTH(message) >= 20 AND LENGTH(message) <= 1500),
    
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

-- Media Inquiries Table
CREATE TABLE media_inquiries (
    inquiry_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status VARCHAR(20) NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'acknowledged', 'in_progress', 'resolved', 'closed')),
    priority VARCHAR(10) NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    urgency VARCHAR(10) NOT NULL DEFAULT 'medium' CHECK (urgency IN ('low', 'medium', 'high')),
    
    -- Media Contact Information
    outlet VARCHAR(100) NOT NULL,
    contact_name VARCHAR(50) NOT NULL,
    title VARCHAR(50) NOT NULL,
    email VARCHAR(254) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    
    -- Media Details
    media_type VARCHAR(20) CHECK (media_type IN ('print', 'digital', 'television', 'radio', 'podcast', 'medical-journal', 'other')),
    deadline DATE,
    subject TEXT NOT NULL CHECK (LENGTH(subject) >= 20 AND LENGTH(subject) <= 1500),
    
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

-- Volunteer Applications Table
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

-- Donations Inquiries Indexes
CREATE INDEX idx_donations_inquiries_status ON donations_inquiries(status) WHERE NOT is_deleted;
CREATE INDEX idx_donations_inquiries_priority ON donations_inquiries(priority) WHERE NOT is_deleted;
CREATE INDEX idx_donations_inquiries_created_at ON donations_inquiries(created_at) WHERE NOT is_deleted;
CREATE INDEX idx_donations_inquiries_email ON donations_inquiries(email) WHERE NOT is_deleted;
CREATE INDEX idx_donations_inquiries_donation_type ON donations_inquiries(donation_type) WHERE NOT is_deleted;
CREATE INDEX idx_donations_inquiries_amount_intention ON donations_inquiries(amount_intention) WHERE NOT is_deleted;

-- Business Inquiries Indexes
CREATE INDEX idx_business_inquiries_status ON business_inquiries(status) WHERE NOT is_deleted;
CREATE INDEX idx_business_inquiries_priority ON business_inquiries(priority) WHERE NOT is_deleted;
CREATE INDEX idx_business_inquiries_created_at ON business_inquiries(created_at) WHERE NOT is_deleted;
CREATE INDEX idx_business_inquiries_email ON business_inquiries(email) WHERE NOT is_deleted;
CREATE INDEX idx_business_inquiries_organization ON business_inquiries(organization_name) WHERE NOT is_deleted;
CREATE INDEX idx_business_inquiries_inquiry_type ON business_inquiries(inquiry_type) WHERE NOT is_deleted;

-- Media Inquiries Indexes
CREATE INDEX idx_media_inquiries_status ON media_inquiries(status) WHERE NOT is_deleted;
CREATE INDEX idx_media_inquiries_priority ON media_inquiries(priority) WHERE NOT is_deleted;
CREATE INDEX idx_media_inquiries_urgency ON media_inquiries(urgency) WHERE NOT is_deleted;
CREATE INDEX idx_media_inquiries_created_at ON media_inquiries(created_at) WHERE NOT is_deleted;
CREATE INDEX idx_media_inquiries_email ON media_inquiries(email) WHERE NOT is_deleted;
CREATE INDEX idx_media_inquiries_outlet ON media_inquiries(outlet) WHERE NOT is_deleted;
CREATE INDEX idx_media_inquiries_media_type ON media_inquiries(media_type) WHERE NOT is_deleted;
CREATE INDEX idx_media_inquiries_deadline ON media_inquiries(deadline) WHERE NOT is_deleted AND deadline IS NOT NULL;

-- Volunteer Applications Indexes
CREATE INDEX idx_volunteer_applications_status ON volunteer_applications(status) WHERE NOT is_deleted;
CREATE INDEX idx_volunteer_applications_priority ON volunteer_applications(priority) WHERE NOT is_deleted;
CREATE INDEX idx_volunteer_applications_created_at ON volunteer_applications(created_at) WHERE NOT is_deleted;
CREATE INDEX idx_volunteer_applications_email ON volunteer_applications(email) WHERE NOT is_deleted;
CREATE INDEX idx_volunteer_applications_volunteer_interest ON volunteer_applications(volunteer_interest) WHERE NOT is_deleted;
CREATE INDEX idx_volunteer_applications_availability ON volunteer_applications(availability) WHERE NOT is_deleted;
CREATE INDEX idx_volunteer_applications_age ON volunteer_applications(age) WHERE NOT is_deleted;