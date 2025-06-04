-- +migrate Down
-- Remove send_at column from outbox

PRAGMA foreign_keys = OFF;

CREATE TABLE outbox_temp AS SELECT outbox_id, message_type, recipient, payload, status, created_at, sent_at, retry_count, user_id FROM outbox;
DROP TABLE outbox;
CREATE TABLE outbox (
    outbox_id INTEGER PRIMARY KEY AUTOINCREMENT,
    message_type TEXT NOT NULL,
    recipient TEXT NOT NULL,
    payload TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    sent_at DATETIME,
    retry_count INTEGER DEFAULT 0,
    user_id INTEGER REFERENCES users(user_id) ON DELETE CASCADE
);
INSERT INTO outbox (outbox_id, message_type, recipient, payload, status, created_at, sent_at, retry_count, user_id)
SELECT outbox_id, message_type, recipient, payload, status, created_at, sent_at, retry_count, user_id FROM outbox_temp;
CREATE INDEX IF NOT EXISTS idx_outbox_user_id ON outbox(user_id);
DROP TABLE outbox_temp;

PRAGMA foreign_keys = ON;
