-- Rollback User Points System

-- Drop indexes
DROP INDEX IF EXISTS idx_users_current_streak;
DROP INDEX IF EXISTS idx_users_total_points;
DROP INDEX IF EXISTS idx_user_achievements_user_id;
DROP INDEX IF EXISTS idx_points_history_created_at;
DROP INDEX IF EXISTS idx_points_history_user_id;

-- Drop tables (in reverse order of dependencies)
DROP TABLE IF EXISTS user_achievements;
DROP TABLE IF EXISTS achievements;
DROP TABLE IF EXISTS points_history;

-- Remove columns from users table
ALTER TABLE users DROP COLUMN IF EXISTS last_activity_date;
ALTER TABLE users DROP COLUMN IF EXISTS longest_streak;
ALTER TABLE users DROP COLUMN IF EXISTS current_streak;
ALTER TABLE users DROP COLUMN IF EXISTS total_points; 