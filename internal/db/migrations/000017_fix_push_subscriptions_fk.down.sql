-- +migrate Down
-- Revert back to the incorrect foreign key (for rollback purposes)

DROP TABLE IF EXISTS push_subscriptions;

CREATE TABLE push_subscriptions (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id     INTEGER     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    endpoint    TEXT UNIQUE NOT NULL,
    p256dh_key  TEXT        NOT NULL,
    auth_key    TEXT        NOT NULL,
    user_agent  TEXT,
    platform    TEXT,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
); 