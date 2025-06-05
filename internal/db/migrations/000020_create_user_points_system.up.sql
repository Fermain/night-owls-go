-- User Points System for Leaderboard
-- Tracks points, achievements, and streaks for Night Owls volunteers

-- Add points columns to users table
ALTER TABLE users ADD COLUMN total_points INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN shift_count INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN last_activity_date DATE;

-- Create points_history table to track all point awards
CREATE TABLE points_history (
    history_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    booking_id INTEGER REFERENCES bookings(booking_id) ON DELETE SET NULL,
    points_awarded INTEGER NOT NULL,
    reason TEXT NOT NULL, -- 'shift_completion', 'check_in_on_time', 'report_filed', 'weekend_bonus', etc.
    multiplier REAL DEFAULT 1.0, -- For bonus multipliers during special events
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create achievements table for badges/milestones
CREATE TABLE achievements (
    achievement_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL,
    icon TEXT, -- Icon name or emoji
    shifts_threshold INTEGER, -- Number of shifts needed to unlock
    special_condition TEXT, -- JSON for complex conditions
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create user_achievements to track which users have earned which achievements
CREATE TABLE user_achievements (
    user_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    achievement_id INTEGER NOT NULL REFERENCES achievements(achievement_id) ON DELETE CASCADE,
    earned_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, achievement_id)
);

-- Insert achievements based on shift count
INSERT INTO achievements (name, description, icon, shifts_threshold) VALUES
('Owlet', 'Complete your first shift', '🐣', 1),
('Solid Owl', 'Complete 20 shifts', '🦉', 20),
('Wise Owl', 'Complete 50 shifts', '🦅', 50),
('Super Owl', 'Complete 100 shifts', '🐉', 100);

-- Create indexes for performance
CREATE INDEX idx_points_history_user_id ON points_history(user_id);
CREATE INDEX idx_points_history_created_at ON points_history(created_at);
CREATE INDEX idx_user_achievements_user_id ON user_achievements(user_id);
CREATE INDEX idx_users_total_points ON users(total_points DESC);
CREATE INDEX idx_users_shift_count ON users(shift_count DESC); 