ALTER TABLE users
ADD COLUMN role TEXT NOT NULL DEFAULT 'guest' CHECK (role IN ('admin', 'owl', 'guest')); 