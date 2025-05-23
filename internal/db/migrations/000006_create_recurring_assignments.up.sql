CREATE TABLE recurring_assignments (
    recurring_assignment_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    buddy_name TEXT,
    day_of_week INTEGER NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6),
    schedule_id INTEGER NOT NULL REFERENCES schedules(schedule_id) ON DELETE CASCADE,
    time_slot TEXT NOT NULL, -- Format: "HH:MM-HH:MM"
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, day_of_week, schedule_id, time_slot)
); 