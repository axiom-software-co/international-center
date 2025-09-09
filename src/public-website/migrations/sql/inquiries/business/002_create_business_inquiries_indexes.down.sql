-- Drop business_inquiries indexes
DROP INDEX IF EXISTS idx_business_inquiries_status;
DROP INDEX IF EXISTS idx_business_inquiries_priority;
DROP INDEX IF EXISTS idx_business_inquiries_created_at;
DROP INDEX IF EXISTS idx_business_inquiries_email;
DROP INDEX IF EXISTS idx_business_inquiries_organization;
DROP INDEX IF EXISTS idx_business_inquiries_inquiry_type;