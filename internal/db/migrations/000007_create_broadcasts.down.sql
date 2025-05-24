-- +migrate Down
DROP INDEX IF EXISTS idx_broadcasts_created_at;
DROP INDEX IF EXISTS idx_broadcasts_sender;
DROP INDEX IF EXISTS idx_broadcasts_status;
DROP TABLE IF EXISTS broadcasts; 