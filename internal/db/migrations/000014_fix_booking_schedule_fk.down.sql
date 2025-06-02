-- Revert back to the original bookings table structure without CASCADE DELETE
CREATE TABLE bookings_old (
    booking_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(user_id),
    schedule_id INTEGER NOT NULL REFERENCES schedules(schedule_id),
    shift_start DATETIME NOT NULL,
    shift_end DATETIME NOT NULL,
    buddy_user_id INTEGER REFERENCES users(user_id),
    buddy_name TEXT,
    checked_in_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(schedule_id, shift_start)
);

-- Copy data back
INSERT INTO bookings_old (
    booking_id, user_id, schedule_id, shift_start, shift_end, 
    buddy_user_id, buddy_name, checked_in_at, created_at
)
SELECT 
    booking_id, user_id, schedule_id, shift_start, shift_end, 
    buddy_user_id, buddy_name, checked_in_at, created_at
FROM bookings;

-- Replace table
DROP TABLE bookings;
ALTER TABLE bookings_old RENAME TO bookings; 