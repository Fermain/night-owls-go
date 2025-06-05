-- Remove photo support
DROP INDEX IF EXISTS idx_report_photos_upload_timestamp;
DROP INDEX IF EXISTS idx_report_photos_report_id;
DROP TABLE IF EXISTS report_photos;
ALTER TABLE reports DROP COLUMN photo_count; 