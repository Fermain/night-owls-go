-- +migrate Down
-- Remove GPS location fields from reports table
DROP INDEX IF EXISTS idx_reports_location;

ALTER TABLE reports DROP COLUMN latitude;
ALTER TABLE reports DROP COLUMN longitude;
ALTER TABLE reports DROP COLUMN gps_accuracy;
ALTER TABLE reports DROP COLUMN gps_timestamp; 