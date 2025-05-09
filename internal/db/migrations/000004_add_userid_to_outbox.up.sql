-- +migrate Up
ALTER TABLE outbox ADD COLUMN user_id INTEGER REFERENCES users(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_outbox_user_id ON outbox(user_id);
