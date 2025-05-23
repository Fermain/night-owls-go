-- name: CreateOutboxItem :one
INSERT INTO outbox (
    message_type,
    recipient,
    payload,
    user_id
) VALUES (
    ?,
    ?,
    ?,
    ?
)
RETURNING *;

-- name: GetPendingOutboxItems :many
SELECT * FROM outbox
WHERE status = 'pending'
ORDER BY created_at ASC
LIMIT ?; -- Limit to prevent processing too many at once

-- name: GetRecentOutboxItemsByRecipient :many
SELECT * FROM outbox
WHERE recipient = ?
ORDER BY created_at DESC
LIMIT ?;

-- name: UpdateOutboxItemStatus :one
UPDATE outbox
SET status = ?,
    sent_at = ?,
    retry_count = ?
WHERE outbox_id = ?
RETURNING *; 