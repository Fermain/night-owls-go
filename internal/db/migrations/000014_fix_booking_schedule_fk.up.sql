-- Clean up orphaned bookings first
DELETE FROM bookings 
WHERE schedule_id NOT IN (SELECT schedule_id FROM schedules);

-- Recreate the bookings table with proper CASCADE DELETE
-- First, create new table with correct constraints
CREATE TABLE bookings_new (
    booking_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(user_id),
    schedule_id INTEGER NOT NULL REFERENCES schedules(schedule_id) ON DELETE CASCADE,
    shift_start DATETIME NOT NULL,
    shift_end DATETIME NOT NULL,
    buddy_user_id INTEGER REFERENCES users(user_id),
    buddy_name TEXT,
    checked_in_at DATETIME, -- Updated from old 'attended' boolean
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(schedule_id, shift_start)
);

-- Copy valid data from old table
INSERT INTO bookings_new (
    booking_id, user_id, schedule_id, shift_start, shift_end, 
    buddy_user_id, buddy_name, checked_in_at, created_at
)
SELECT 
    booking_id, user_id, schedule_id, shift_start, shift_end, 
    buddy_user_id, buddy_name, checked_in_at, created_at
FROM bookings 
WHERE schedule_id IN (SELECT schedule_id FROM schedules);

-- Drop old table and rename new one
DROP TABLE bookings;
ALTER TABLE bookings_new RENAME TO bookings;

-- Recreate indexes if any existed
CREATE UNIQUE INDEX IF NOT EXISTS idx_bookings_schedule_start 
ON bookings(schedule_id, shift_start); 