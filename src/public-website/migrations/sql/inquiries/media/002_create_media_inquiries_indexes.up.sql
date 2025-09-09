-- Create performance indexes for media_inquiries matching TABLES-INQUIRIES-MEDIA.md specification
CREATE INDEX idx_media_inquiries_status ON media_inquiries(status) WHERE NOT is_deleted;
CREATE INDEX idx_media_inquiries_priority ON media_inquiries(priority) WHERE NOT is_deleted;
CREATE INDEX idx_media_inquiries_urgency ON media_inquiries(urgency) WHERE NOT is_deleted;
CREATE INDEX idx_media_inquiries_created_at ON media_inquiries(created_at) WHERE NOT is_deleted;
CREATE INDEX idx_media_inquiries_email ON media_inquiries(email) WHERE NOT is_deleted;
CREATE INDEX idx_media_inquiries_outlet ON media_inquiries(outlet) WHERE NOT is_deleted;
CREATE INDEX idx_media_inquiries_media_type ON media_inquiries(media_type) WHERE NOT is_deleted;
CREATE INDEX idx_media_inquiries_deadline ON media_inquiries(deadline) WHERE NOT is_deleted AND deadline IS NOT NULL;