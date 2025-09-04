# Media Inquiries Tables

This document defines the PostgreSQL table schemas for managing media relations, press inquiries, and journalist communications.

## Overview

The media inquiries domain handles press relations, interview requests, media coverage, and publication inquiries from journalists, reporters, and media organizations requiring timely responses and deadline management.

## Tables

### media_inquiries

Main table for media relations and press inquiries matching the website form fields exactly.

```sql
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

-- Indexes for performance
CREATE INDEX idx_media_inquiries_status ON media_inquiries(status) WHERE NOT is_deleted;
CREATE INDEX idx_media_inquiries_priority ON media_inquiries(priority) WHERE NOT is_deleted;
CREATE INDEX idx_media_inquiries_urgency ON media_inquiries(urgency) WHERE NOT is_deleted;
CREATE INDEX idx_media_inquiries_created_at ON media_inquiries(created_at) WHERE NOT is_deleted;
CREATE INDEX idx_media_inquiries_email ON media_inquiries(email) WHERE NOT is_deleted;
CREATE INDEX idx_media_inquiries_outlet ON media_inquiries(outlet) WHERE NOT is_deleted;
CREATE INDEX idx_media_inquiries_media_type ON media_inquiries(media_type) WHERE NOT is_deleted;
CREATE INDEX idx_media_inquiries_deadline ON media_inquiries(deadline) WHERE NOT is_deleted AND deadline IS NOT NULL;

```

## Constraints and Business Rules

- `outlet` must be at least 2 characters (validated in application layer)
- `contact_name` must contain only letters, spaces, hyphens, and apostrophes (validated in application layer)
- `email` must be a valid business email format (validated in application layer)
- `phone` is required for all media inquiries and must be valid 10-digit USA format (validated in application layer)
- `subject` must be between 20 and 1500 characters (enforced by CHECK constraint)
- `media_type` must be one of: print, digital, television, radio, podcast, medical-journal, other
- `deadline` when provided automatically calculates urgency and may escalate priority
- Urgency calculation: ≤1 day = high, ≤3 days = medium, >3 days = low

## Notes

- Table follows established domain patterns with UUID primary keys and standard audit fields
- Phone numbers are required as per media form requirements for direct contact
- Deadline field drives automatic urgency and priority calculation through database trigger
- Subject field has different semantic meaning than business inquiry message field
- Urgency and priority are automatically calculated based on deadline proximity
- Soft delete pattern maintains audit integrity for media relations history
