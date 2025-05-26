-- +migrate Down
-- Revert bookings.checked_in_at timestamp back to attended boolean
-- First add the old column back
ALTER TABLE bookings ADD COLUMN attended BOOLEAN NOT NULL DEFAULT 0;

-- Copy existing data: if checked_in_at is not NULL, set attended=true
UPDATE bookings SET attended = 1 WHERE checked_in_at IS NOT NULL;

-- Drop the new column
ALTER TABLE bookings DROP COLUMN checked_in_at; 