CREATE TABLE emergency_contacts (
    contact_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    number TEXT NOT NULL,
    description TEXT,
    is_default BOOLEAN NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Insert RUSA as the default emergency contact
INSERT INTO emergency_contacts (name, number, description, is_default, display_order) 
VALUES ('RUSA', '086 123 4333', 'Private Security Response Unit', 1, 1);

-- Insert SAPS as secondary option
INSERT INTO emergency_contacts (name, number, description, is_default, display_order) 
VALUES ('SAPS', '10111', 'South African Police Service', 0, 2);

-- Insert Medical Emergency
INSERT INTO emergency_contacts (name, number, description, is_default, display_order) 
VALUES ('ER24', '084 124', 'Emergency Medical Services', 0, 3); 