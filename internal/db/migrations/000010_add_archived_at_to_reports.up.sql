-- Add archived_at column to reports table for soft-delete functionality
ALTER TABLE reports ADD COLUMN archived_at TIMESTAMP NULL;

-- Create index for efficient querying of non-archived reports
CREATE INDEX idx_reports_archived_at ON reports(archived_at);

-- Create index for efficient archiving queries by severity and date
CREATE INDEX idx_reports_severity_created_archived ON reports(severity, created_at, archived_at); 