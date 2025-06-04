-- Create audit_events table for comprehensive system activity tracking
CREATE TABLE audit_events (
    event_id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_type TEXT NOT NULL, -- 'user.created', 'user.login', 'booking.assigned', etc.
    actor_user_id INTEGER REFERENCES users(user_id), -- Who performed the action (null for system events)
    target_user_id INTEGER REFERENCES users(user_id), -- Who was affected (optional)
    entity_type TEXT NOT NULL, -- 'user', 'booking', 'schedule', 'report', etc.
    entity_id INTEGER, -- ID of the affected entity (optional for some events)
    action TEXT NOT NULL, -- 'created', 'updated', 'deleted', 'login', 'assigned', etc.
    details TEXT, -- JSON with before/after values and context
    ip_address TEXT, -- For security tracking
    user_agent TEXT, -- Browser/device info
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for efficient querying
CREATE INDEX idx_audit_events_type ON audit_events(event_type);
CREATE INDEX idx_audit_events_actor ON audit_events(actor_user_id);
CREATE INDEX idx_audit_events_target ON audit_events(target_user_id);
CREATE INDEX idx_audit_events_entity ON audit_events(entity_type, entity_id);
CREATE INDEX idx_audit_events_created_at ON audit_events(created_at);
CREATE INDEX idx_audit_events_action ON audit_events(action); 