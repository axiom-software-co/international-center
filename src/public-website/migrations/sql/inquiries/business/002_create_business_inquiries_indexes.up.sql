-- Create performance indexes for business_inquiries matching TABLES-INQUIRIES-BUSINESS.md specification
CREATE INDEX idx_business_inquiries_status ON business_inquiries(status) WHERE NOT is_deleted;
CREATE INDEX idx_business_inquiries_priority ON business_inquiries(priority) WHERE NOT is_deleted;
CREATE INDEX idx_business_inquiries_created_at ON business_inquiries(created_at) WHERE NOT is_deleted;
CREATE INDEX idx_business_inquiries_email ON business_inquiries(email) WHERE NOT is_deleted;
CREATE INDEX idx_business_inquiries_organization ON business_inquiries(organization_name) WHERE NOT is_deleted;
CREATE INDEX idx_business_inquiries_inquiry_type ON business_inquiries(inquiry_type) WHERE NOT is_deleted;