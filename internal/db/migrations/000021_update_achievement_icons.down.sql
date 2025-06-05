-- Revert Achievement Icons Migration
-- Reverts achievement icons back to original emoji set

-- Revert achievement icons to original icons
UPDATE achievements SET icon = '🦉' WHERE name = 'Owlet';
UPDATE achievements SET icon = '🦉' WHERE name = 'Solid Owl';
UPDATE achievements SET icon = '🦜' WHERE name = 'Wise Owl';
UPDATE achievements SET icon = '🦅' WHERE name = 'Super Owl'; 