-- Add support for off-shift reports
-- Make booking_id nullable and add user_id for off-shift reports

-- Create new reports table with updated schema
CREATE TABLE reports_new (
    report_id INTEGER PRIMARY KEY AUTOINCREMENT,
    booking_id INTEGER REFERENCES bookings(booking_id) ON DELETE CASCADE, -- Now nullable
    user_id INTEGER REFERENCES users(user_id) ON DELETE CASCADE, -- For off-shift reports
    severity INTEGER NOT NULL, -- 0=normal, 1=suspicion, 2=incident
    message TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    latitude REAL,
    longitude REAL,
    gps_accuracy REAL,
    gps_timestamp DATETIME,
    archived_at DATETIME,
    -- Ensure either booking_id or user_id is provided
    CHECK ((booking_id IS NOT NULL) OR (user_id IS NOT NULL))
);

-- Copy existing data from old table
INSERT INTO reports_new (
    report_id, booking_id, user_id, severity, message, created_at, 
    latitude, longitude, gps_accuracy, gps_timestamp, archived_at
)
SELECT 
    r.report_id, 
    r.booking_id, 
    b.user_id, -- Get user_id from booking for existing reports
    r.severity, 
    r.message, 
    r.created_at,
    r.latitude, 
    r.longitude, 
    r.gps_accuracy, 
    r.gps_timestamp, 
    r.archived_at
FROM reports r
JOIN bookings b ON r.booking_id = b.booking_id;

-- Drop old table and rename new one
DROP TABLE reports;
ALTER TABLE reports_new RENAME TO reports; 