-- Create performance indexes for donations_inquiries matching TABLES-INQUIRIES-DONATIONS.md specification
CREATE INDEX idx_donations_inquiries_status ON donations_inquiries(status) WHERE NOT is_deleted;
CREATE INDEX idx_donations_inquiries_priority ON donations_inquiries(priority) WHERE NOT is_deleted;
CREATE INDEX idx_donations_inquiries_created_at ON donations_inquiries(created_at) WHERE NOT is_deleted;
CREATE INDEX idx_donations_inquiries_email ON donations_inquiries(email) WHERE NOT is_deleted;
CREATE INDEX idx_donations_inquiries_donor_type ON donations_inquiries(donor_type) WHERE NOT is_deleted;
CREATE INDEX idx_donations_inquiries_interest_area ON donations_inquiries(interest_area) WHERE NOT is_deleted;
CREATE INDEX idx_donations_inquiries_amount_range ON donations_inquiries(preferred_amount_range) WHERE NOT is_deleted;