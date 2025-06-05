-- Revert Achievement Icons Migration (Actual Database)
-- Reverts achievement icons back to original emoji set

-- Revert the achievement icons to their original values
UPDATE achievements SET icon = '🦉' WHERE name = 'First Steps';
UPDATE achievements SET icon = '🛡️' WHERE name = 'Night Guardian';
UPDATE achievements SET icon = '⭐' WHERE name = 'Dedicated Owl';
UPDATE achievements SET icon = '💎' WHERE name = 'Elite Guardian';
UPDATE achievements SET icon = '🏆' WHERE name = 'Community Hero'; 