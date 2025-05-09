CREATE TABLE users (
    user_id INTEGER PRIMARY KEY AUTOINCREMENT,
    phone TEXT UNIQUE NOT NULL,
    name TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE schedules (
    schedule_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    cron_expr TEXT NOT NULL,
    start_date DATE,
    end_date DATE,
    duration_minutes INTEGER NOT NULL DEFAULT 120,
    timezone TEXT -- e.g., "Africa/Johannesburg", useful if DST affects cron interpretation
);

CREATE TABLE bookings (
    booking_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(user_id),
    schedule_id INTEGER NOT NULL REFERENCES schedules(schedule_id),
    shift_start DATETIME NOT NULL,
    shift_end DATETIME NOT NULL, -- Calculated as shift_start + schedules.duration_minutes
    buddy_user_id INTEGER REFERENCES users(user_id),
    buddy_name TEXT,
    attended BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(schedule_id, shift_start)
);

CREATE TABLE reports (
    report_id INTEGER PRIMARY KEY AUTOINCREMENT,
    booking_id INTEGER NOT NULL REFERENCES bookings(booking_id) ON DELETE CASCADE,
    severity INTEGER NOT NULL, -- 0=low, 1=moderate, 2=serious. Consider CHECK constraint.
    message TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE outbox (
    outbox_id INTEGER PRIMARY KEY AUTOINCREMENT,
    message_type TEXT NOT NULL, -- e.g., "OTP_VERIFICATION", "BOOKING_CONFIRMATION"
    recipient TEXT NOT NULL, -- e.g., phone number or user_id for internal routing
    payload TEXT, -- JSON payload for the message content
    status TEXT NOT NULL DEFAULT 'pending', -- pending, sent, failed
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    sent_at DATETIME,
    retry_count INTEGER DEFAULT 0
); 