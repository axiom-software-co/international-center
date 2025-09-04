# Business Inquiries Tables

This document defines the PostgreSQL table schemas for managing business partnership and collaboration inquiries.

## Overview

The business inquiries domain handles partnership opportunities, licensing requests, research collaborations, technology integrations, and regulatory consultations from organizations seeking professional relationships.

## Tables

### business_inquiries

Main table for business partnership and collaboration requests matching the website form fields exactly.

```sql
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

-- Indexes for performance
CREATE INDEX idx_business_inquiries_status ON business_inquiries(status) WHERE NOT is_deleted;
CREATE INDEX idx_business_inquiries_priority ON business_inquiries(priority) WHERE NOT is_deleted;
CREATE INDEX idx_business_inquiries_created_at ON business_inquiries(created_at) WHERE NOT is_deleted;
CREATE INDEX idx_business_inquiries_email ON business_inquiries(email) WHERE NOT is_deleted;
CREATE INDEX idx_business_inquiries_organization ON business_inquiries(organization_name) WHERE NOT is_deleted;
CREATE INDEX idx_business_inquiries_inquiry_type ON business_inquiries(inquiry_type) WHERE NOT is_deleted;

```




## Constraints and Business Rules

- `organization_name` must be at least 2 characters (validated in application layer)
- `contact_name` must contain only letters, spaces, hyphens, and apostrophes (validated in application layer)
- `email` must be a valid business email format (validated in application layer)
- `phone` is optional but if provided must be a valid 10-digit USA format (validated in application layer)
- `message` must be between 20 and 1500 characters (enforced by CHECK constraint)
- `inquiry_type` must be one of: partnership, licensing, research, technology, regulatory, other

## Notes

- Table follows established domain patterns with UUID primary keys and standard audit fields
- Phone numbers are optional as per the business form requirements
- Industry field is optional to accommodate inquiries from various sectors
- Soft delete pattern maintains audit integrity for compliance requirements
