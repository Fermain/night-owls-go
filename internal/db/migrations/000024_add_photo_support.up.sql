-- Add photo support to reports
-- Add photo count to reports table
ALTER TABLE reports ADD COLUMN photo_count INTEGER DEFAULT 0;

-- Create report_photos table
CREATE TABLE report_photos (
    photo_id INTEGER PRIMARY KEY AUTOINCREMENT,
    report_id INTEGER NOT NULL REFERENCES reports(report_id) ON DELETE CASCADE,
    filename TEXT NOT NULL,
    original_filename TEXT,
    file_size_bytes INTEGER NOT NULL,
    mime_type TEXT NOT NULL,
    width_pixels INTEGER,
    height_pixels INTEGER,
    upload_timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    storage_path TEXT NOT NULL,
    thumbnail_path TEXT,
    
    -- Security and validation
    checksum_sha256 TEXT NOT NULL,
    is_processed BOOLEAN DEFAULT FALSE,
    
    CONSTRAINT valid_mime_type CHECK (mime_type IN ('image/jpeg', 'image/png', 'image/webp')),
    CONSTRAINT valid_file_size CHECK (file_size_bytes > 0 AND file_size_bytes <= 10485760) -- 10MB max
);

CREATE INDEX idx_report_photos_report_id ON report_photos(report_id);
CREATE INDEX idx_report_photos_upload_timestamp ON report_photos(upload_timestamp); 