-- Update Achievement Icons Migration (Actual Database)
-- Updates achievement icons to new emoji set for better visual progression

-- Update the achievement icons that actually exist in the database
-- Following progression: hatching chick -> owl -> eagle -> dragon -> trophy
UPDATE achievements SET icon = '🐣' WHERE name = 'First Steps';
UPDATE achievements SET icon = '🦉' WHERE name = 'Night Guardian';
UPDATE achievements SET icon = '🦅' WHERE name = 'Dedicated Owl';
UPDATE achievements SET icon = '🐉' WHERE name = 'Elite Guardian';
UPDATE achievements SET icon = '🏆' WHERE name = 'Community Hero'; 