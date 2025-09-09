# Donations Inquiries Tables

This document defines the PostgreSQL table schemas for managing donor contact inquiries and fundraising communications.

## Overview

The donations inquiries domain handles contact requests from potential donors, philanthropic organizations, corporate sponsors, and estate planners interested in supporting clinic development, research initiatives, patient care programs, and equipment acquisition.

## Tables

### donations_inquiries

Main table for donor contact and fundraising inquiries matching the website form fields exactly.

```sql
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

-- Indexes for performance
CREATE INDEX idx_donations_inquiries_status ON donations_inquiries(status) WHERE NOT is_deleted;
CREATE INDEX idx_donations_inquiries_priority ON donations_inquiries(priority) WHERE NOT is_deleted;
CREATE INDEX idx_donations_inquiries_created_at ON donations_inquiries(created_at) WHERE NOT is_deleted;
CREATE INDEX idx_donations_inquiries_email ON donations_inquiries(email) WHERE NOT is_deleted;
CREATE INDEX idx_donations_inquiries_donor_type ON donations_inquiries(donor_type) WHERE NOT is_deleted;
CREATE INDEX idx_donations_inquiries_interest_area ON donations_inquiries(interest_area) WHERE NOT is_deleted;
CREATE INDEX idx_donations_inquiries_amount_range ON donations_inquiries(preferred_amount_range) WHERE NOT is_deleted;

```




## Constraints and Business Rules

- `contact_name` must be at least 2 characters (validated in application layer)
- `email` must be a valid email format (validated in application layer)
- `phone` is optional but if provided must be a valid 10-digit USA format (validated in application layer)
- `organization` is optional for individual donors but required for corporate/foundation donors (validated in application layer)
- `message` must be between 20 and 2000 characters (enforced by CHECK constraint)
- `donor_type` must be one of: individual, corporate, foundation, estate, other
- `interest_area` must be one of: clinic-development, research-funding, patient-care, equipment, general-support, other
- `preferred_amount_range` must be one of: under-1000, 1000-5000, 5000-25000, 25000-100000, over-100000, undisclosed
- `donation_frequency` must be one of: one-time, monthly, quarterly, annually, other

## Notes

- Table follows established domain patterns with UUID primary keys and standard audit fields
- Phone numbers are optional to accommodate international donors and privacy preferences
- Organization field supports corporate and foundation donor identification
- Amount ranges provide fundraising insights while respecting donor privacy
- Interest areas align with clinic strategic priorities and funding categories
- Soft delete pattern maintains audit integrity for fundraising compliance requirements
