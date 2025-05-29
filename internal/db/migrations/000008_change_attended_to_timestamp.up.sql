-- +migrate Up
-- Change bookings.attended from boolean to timestamp
-- First add the new column
ALTER TABLE bookings ADD COLUMN checked_in_at DATETIME;

-- Copy existing data: if attended=true, set to current timestamp, otherwise NULL
UPDATE bookings SET checked_in_at = CURRENT_TIMESTAMP WHERE attended = 1;

-- Drop the old column
ALTER TABLE bookings DROP COLUMN attended; 