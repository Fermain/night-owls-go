-- Update Achievement Icons Migration
-- Updates achievement icons to new emoji set for better visual progression

-- Update achievement icons to new progression: hatching chick -> owl -> eagle -> dragon
UPDATE achievements SET icon = '🐣' WHERE name = 'Owlet';
UPDATE achievements SET icon = '🦉' WHERE name = 'Solid Owl';
UPDATE achievements SET icon = '🦅' WHERE name = 'Wise Owl';
UPDATE achievements SET icon = '🐉' WHERE name = 'Super Owl'; 