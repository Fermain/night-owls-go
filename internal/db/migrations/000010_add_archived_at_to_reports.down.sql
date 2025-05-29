-- Remove indexes
DROP INDEX IF EXISTS idx_reports_severity_created_archived;
DROP INDEX IF EXISTS idx_reports_archived_at;

-- Remove archived_at column from reports table
ALTER TABLE reports DROP COLUMN archived_at; 