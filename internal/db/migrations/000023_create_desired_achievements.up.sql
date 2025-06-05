-- Create User's Desired Achievement System
-- Replace existing achievements with the user's requested Owl-themed progression

-- Clear existing achievements first
DELETE FROM user_achievements;
DELETE FROM achievements;

-- Insert the user's desired achievements 
-- Using special_condition to store shift requirements as JSON
INSERT INTO achievements (name, description, icon, special_condition) VALUES
('Owlet', 'Complete your first shift', '🐣', '{"shifts_required": 1}'),
('Solid Owl', 'Complete 20 shifts', '🦉', '{"shifts_required": 20}'),
('Wise Owl', 'Complete 50 shifts', '🦅', '{"shifts_required": 50}'),
('Super Owl', 'Complete 100 shifts', '🐉', '{"shifts_required": 100}'); 