-- Revert User's Desired Achievement System
-- Restore original achievements

-- Clear the user's achievements
DELETE FROM user_achievements;
DELETE FROM achievements;

-- Restore original achievements
INSERT INTO achievements (name, description, icon, points_threshold) VALUES
('First Steps', 'Complete your first shift', '🦉', 100),
('Night Guardian', 'Earn 100 points', '🛡️', 500),
('Dedicated Owl', 'Earn 500 points', '⭐', 1000),
('Elite Guardian', 'Earn 1000 points', '💎', 2500),
('Community Hero', 'Earn 2500 points', '🏆', NULL);

INSERT INTO achievements (name, description, icon, streak_threshold) VALUES
('Consistent', 'Maintain a 3-shift streak', '🔥', 3),
('Reliable', 'Maintain a 7-shift streak', '💪', 7),
('Unwavering', 'Maintain a 15-shift streak', '⚡', 15),
('Legendary', 'Maintain a 30-shift streak', '👑', 30); 