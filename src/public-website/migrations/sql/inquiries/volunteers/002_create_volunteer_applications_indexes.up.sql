-- Create performance indexes for volunteer_applications matching TABLES-VOLUNTEERS.md specification
CREATE INDEX idx_volunteer_applications_status ON volunteer_applications(status) WHERE NOT is_deleted;
CREATE INDEX idx_volunteer_applications_priority ON volunteer_applications(priority) WHERE NOT is_deleted;
CREATE INDEX idx_volunteer_applications_created_at ON volunteer_applications(created_at) WHERE NOT is_deleted;
CREATE INDEX idx_volunteer_applications_email ON volunteer_applications(email) WHERE NOT is_deleted;
CREATE INDEX idx_volunteer_applications_volunteer_interest ON volunteer_applications(volunteer_interest) WHERE NOT is_deleted;
CREATE INDEX idx_volunteer_applications_availability ON volunteer_applications(availability) WHERE NOT is_deleted;
CREATE INDEX idx_volunteer_applications_age ON volunteer_applications(age) WHERE NOT is_deleted;