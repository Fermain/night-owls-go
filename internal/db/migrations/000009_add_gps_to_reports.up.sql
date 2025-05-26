-- +migrate Up
-- Add GPS location fields to reports table
ALTER TABLE reports ADD COLUMN latitude REAL;
ALTER TABLE reports ADD COLUMN longitude REAL;
ALTER TABLE reports ADD COLUMN gps_accuracy REAL;
ALTER TABLE reports ADD COLUMN gps_timestamp DATETIME;

-- Add index for location-based queries
CREATE INDEX IF NOT EXISTS idx_reports_location ON reports(latitude, longitude);

-- Add comments for documentation
-- latitude: GPS latitude coordinate (-90 to 90)
-- longitude: GPS longitude coordinate (-180 to 180)  
-- gps_accuracy: GPS accuracy in meters
-- gps_timestamp: When the GPS location was captured 