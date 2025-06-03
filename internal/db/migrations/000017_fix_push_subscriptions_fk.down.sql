-- +migrate Down
-- Rollback: Revert to the original (incorrect) foreign key constraint
-- 
-- IMPACT: This rollback preserves existing subscription data while reverting the FK constraint
-- NOTE: Rolling back will restore the original FK issue, but data is preserved

-- Disable foreign key checks temporarily for the migration
PRAGMA foreign_keys = OFF;

-- Create backup table with current data
CREATE TABLE push_subscriptions_temp AS SELECT * FROM push_subscriptions;

-- Drop the current table (with correct foreign key)
DROP TABLE push_subscriptions;

-- Recreate the table with original (incorrect) foreign key constraint
CREATE TABLE push_subscriptions (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id     INTEGER     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    endpoint    TEXT UNIQUE NOT NULL,
    p256dh_key  TEXT        NOT NULL,
    auth_key    TEXT        NOT NULL,
    user_agent  TEXT,
    platform    TEXT,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Restore data from backup
INSERT INTO push_subscriptions (id, user_id, endpoint, p256dh_key, auth_key, user_agent, platform, created_at)
SELECT id, user_id, endpoint, p256dh_key, auth_key, user_agent, platform, created_at 
FROM push_subscriptions_temp;

-- Clean up temporary table
DROP TABLE push_subscriptions_temp;

-- Re-enable foreign key checks (note: this will restore the original FK issue)
PRAGMA foreign_keys = ON; 