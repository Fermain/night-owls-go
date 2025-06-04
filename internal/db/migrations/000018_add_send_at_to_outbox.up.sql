-- +migrate Up
-- Add send_at column to outbox for scheduling message dispatch
-- Existing rows default to immediate dispatch via CURRENT_TIMESTAMP

ALTER TABLE outbox ADD COLUMN send_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_outbox_send_at ON outbox(send_at);
