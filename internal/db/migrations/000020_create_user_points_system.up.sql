-- User Points System for Leaderboard
-- Tracks points, achievements, and streaks for Night Owls volunteers

-- Add points columns to users table
ALTER TABLE users ADD COLUMN total_points INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN current_streak INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN longest_streak INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN last_activity_date DATE;

-- Create points_history table to track all point awards
CREATE TABLE points_history (
    history_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    booking_id INTEGER REFERENCES bookings(booking_id) ON DELETE SET NULL,
    points_awarded INTEGER NOT NULL,
    reason TEXT NOT NULL, -- 'shift_completion', 'check_in_on_time', 'report_filed', 'streak_bonus', etc.
    multiplier REAL DEFAULT 1.0, -- For bonus multipliers during special events
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create achievements table for badges/milestones
CREATE TABLE achievements (
    achievement_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL,
    icon TEXT, -- Icon name or emoji
    points_threshold INTEGER, -- Points needed to unlock (if points-based)
    streak_threshold INTEGER, -- Streak needed to unlock (if streak-based)
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

-- Insert initial achievements
INSERT INTO achievements (name, description, icon, points_threshold) VALUES
('First Steps', 'Complete your first shift', '🦉', 10),
('Night Guardian', 'Earn 100 points', '🛡️', 100),
('Dedicated Owl', 'Earn 500 points', '⭐', 500),
('Elite Guardian', 'Earn 1000 points', '💎', 1000),
('Community Hero', 'Earn 2500 points', '🏆', 2500);

INSERT INTO achievements (name, description, icon, streak_threshold) VALUES
('Consistent', 'Maintain a 3-shift streak', '🔥', 3),
('Reliable', 'Maintain a 7-shift streak', '💪', 7),
('Unwavering', 'Maintain a 15-shift streak', '⚡', 15),
('Legendary', 'Maintain a 30-shift streak', '👑', 30);

-- Create indexes for performance
CREATE INDEX idx_points_history_user_id ON points_history(user_id);
CREATE INDEX idx_points_history_created_at ON points_history(created_at);
CREATE INDEX idx_user_achievements_user_id ON user_achievements(user_id);
CREATE INDEX idx_users_total_points ON users(total_points DESC);
CREATE INDEX idx_users_current_streak ON users(current_streak DESC); 