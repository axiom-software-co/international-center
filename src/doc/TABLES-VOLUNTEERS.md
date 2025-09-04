# Volunteer Applications Tables

This document defines the PostgreSQL table schemas for managing volunteer applications and community engagement inquiries.

## Overview

The volunteer applications domain handles recruitment inquiries from community members interested in patient support, research assistance, administrative support, and community outreach programs requiring background screening and skills assessment.

## Tables

### volunteer_applications

Main table for volunteer recruitment applications matching the website form fields exactly.

```sql
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

-- Indexes for performance
CREATE INDEX idx_volunteer_applications_status ON volunteer_applications(status) WHERE NOT is_deleted;
CREATE INDEX idx_volunteer_applications_priority ON volunteer_applications(priority) WHERE NOT is_deleted;
CREATE INDEX idx_volunteer_applications_created_at ON volunteer_applications(created_at) WHERE NOT is_deleted;
CREATE INDEX idx_volunteer_applications_email ON volunteer_applications(email) WHERE NOT is_deleted;
CREATE INDEX idx_volunteer_applications_volunteer_interest ON volunteer_applications(volunteer_interest) WHERE NOT is_deleted;
CREATE INDEX idx_volunteer_applications_availability ON volunteer_applications(availability) WHERE NOT is_deleted;
CREATE INDEX idx_volunteer_applications_age ON volunteer_applications(age) WHERE NOT is_deleted;

```




## Constraints and Business Rules

- `first_name` and `last_name` must be at least 2 characters and contain only letters, spaces, hyphens, and apostrophes (validated in application layer)
- `email` must be a valid email format (validated in application layer)
- `phone` is required and must be a valid 10-digit USA format (validated in application layer)
- `age` must be 18 or older and not exceed 100 (enforced by CHECK constraint)
- `volunteer_interest` must be one of: patient-support, community-outreach, research-support, administrative-support, multiple, other
- `availability` must be one of: 2-4-hours, 4-8-hours, 8-16-hours, 16-hours-plus, flexible
- `experience` is optional but if provided must not exceed 1000 characters (enforced by CHECK constraint)
- `motivation` must be between 20 and 1500 characters (enforced by CHECK constraint)
- `schedule_preferences` is optional and limited to 500 characters for scheduling details

## Notes

- Table follows established domain patterns with UUID primary keys and standard audit fields
- Phone numbers are required as per volunteer form requirements for direct contact during screening process
- Age verification is enforced at database level to ensure legal compliance for volunteer programs
- Status progression supports complete volunteer lifecycle from application through approval/decline
- Experience field accommodates optional background information without requiring structured data
- Motivation field enforces minimum content requirement to ensure quality applications
- Soft delete pattern maintains audit integrity for volunteer program compliance requirements