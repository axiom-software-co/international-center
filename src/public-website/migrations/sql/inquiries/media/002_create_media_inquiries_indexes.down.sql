-- Drop media_inquiries indexes
DROP INDEX IF EXISTS idx_media_inquiries_status;
DROP INDEX IF EXISTS idx_media_inquiries_priority;
DROP INDEX IF EXISTS idx_media_inquiries_urgency;
DROP INDEX IF EXISTS idx_media_inquiries_created_at;
DROP INDEX IF EXISTS idx_media_inquiries_email;
DROP INDEX IF EXISTS idx_media_inquiries_outlet;
DROP INDEX IF EXISTS idx_media_inquiries_media_type;
DROP INDEX IF EXISTS idx_media_inquiries_deadline;