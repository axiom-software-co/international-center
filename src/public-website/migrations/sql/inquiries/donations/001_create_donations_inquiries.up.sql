-- Create donations_inquiries table matching TABLES-INQUIRIES-DONATIONS.md specification
CREATE TABLE donations_inquiries (
    inquiry_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status VARCHAR(20) NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'acknowledged', 'in_progress', 'resolved', 'closed')),
    priority VARCHAR(10) NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    
    -- Contact Information
    contact_name VARCHAR(50) NOT NULL,
    email VARCHAR(254) NOT NULL,
    phone VARCHAR(20),
    organization VARCHAR(100),
    
    -- Donor Classification
    donor_type VARCHAR(20) NOT NULL CHECK (donor_type IN ('individual', 'corporate', 'foundation', 'estate', 'other')),
    
    -- Interest Areas
    interest_area VARCHAR(30) CHECK (interest_area IN ('clinic-development', 'research-funding', 'patient-care', 'equipment', 'general-support', 'other')),
    
    -- Donation Intent
    preferred_amount_range VARCHAR(20) CHECK (preferred_amount_range IN ('under-1000', '1000-5000', '5000-25000', '25000-100000', 'over-100000', 'undisclosed')),
    donation_frequency VARCHAR(15) CHECK (donation_frequency IN ('one-time', 'monthly', 'quarterly', 'annually', 'other')),
    message TEXT NOT NULL CHECK (LENGTH(message) >= 20 AND LENGTH(message) <= 2000),
    
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