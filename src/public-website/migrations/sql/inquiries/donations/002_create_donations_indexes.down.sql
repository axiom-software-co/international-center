-- Drop donations inquiries indexes
DROP INDEX IF EXISTS idx_donations_inquiries_status;
DROP INDEX IF EXISTS idx_donations_inquiries_priority;
DROP INDEX IF EXISTS idx_donations_inquiries_created_at;
DROP INDEX IF EXISTS idx_donations_inquiries_email;
DROP INDEX IF EXISTS idx_donations_inquiries_donor_type;
DROP INDEX IF EXISTS idx_donations_inquiries_interest_area;
DROP INDEX IF EXISTS idx_donations_inquiries_amount_range;