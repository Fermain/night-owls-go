-- Revert off-shift reports support
-- Make booking_id NOT NULL again and remove user_id

-- Create old reports table schema
CREATE TABLE reports_old (
    report_id INTEGER PRIMARY KEY AUTOINCREMENT,
    booking_id INTEGER NOT NULL REFERENCES bookings(booking_id) ON DELETE CASCADE,
    severity INTEGER NOT NULL,
    message TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    latitude REAL,
    longitude REAL,
    gps_accuracy REAL,
    gps_timestamp DATETIME,
    archived_at DATETIME
);

-- Copy back only reports that have booking_id (skip off-shift reports)
INSERT INTO reports_old (
    report_id, booking_id, severity, message, created_at,
    latitude, longitude, gps_accuracy, gps_timestamp, archived_at
)
SELECT 
    report_id, booking_id, severity, message, created_at,
    latitude, longitude, gps_accuracy, gps_timestamp, archived_at
FROM reports
WHERE booking_id IS NOT NULL;

-- Drop new table and rename old one
DROP TABLE reports;
ALTER TABLE reports_old RENAME TO reports; 