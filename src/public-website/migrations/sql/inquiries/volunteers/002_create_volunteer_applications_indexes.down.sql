-- Drop volunteer_applications indexes
DROP INDEX IF EXISTS idx_volunteer_applications_status;
DROP INDEX IF EXISTS idx_volunteer_applications_priority;
DROP INDEX IF EXISTS idx_volunteer_applications_created_at;
DROP INDEX IF EXISTS idx_volunteer_applications_email;
DROP INDEX IF EXISTS idx_volunteer_applications_volunteer_interest;
DROP INDEX IF EXISTS idx_volunteer_applications_availability;
DROP INDEX IF EXISTS idx_volunteer_applications_age;