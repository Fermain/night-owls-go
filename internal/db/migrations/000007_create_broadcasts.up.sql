-- +migrate Up
CREATE TABLE broadcasts (
    broadcast_id INTEGER PRIMARY KEY AUTOINCREMENT,
    message TEXT NOT NULL,
    audience TEXT NOT NULL, -- 'all', 'admins', 'owls', 'active'
    sender_user_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    push_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    scheduled_at DATETIME,
    sent_at DATETIME,
    status TEXT NOT NULL DEFAULT 'pending', -- 'pending', 'sending', 'sent', 'failed'
    recipient_count INTEGER DEFAULT 0,
    sent_count INTEGER DEFAULT 0,
    failed_count INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_broadcasts_status ON broadcasts(status);
CREATE INDEX IF NOT EXISTS idx_broadcasts_sender ON broadcasts(sender_user_id);
CREATE INDEX IF NOT EXISTS idx_broadcasts_created_at ON broadcasts(created_at); 