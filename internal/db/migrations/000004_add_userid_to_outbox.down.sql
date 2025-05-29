-- +migrate Down
ALTER TABLE outbox DROP COLUMN user_id;
DROP INDEX IF EXISTS idx_outbox_user_id;
