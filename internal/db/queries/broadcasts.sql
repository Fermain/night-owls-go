-- name: CreateBroadcast :one
INSERT INTO broadcasts (
    message,
    audience,
    sender_user_id,
    push_enabled,
    scheduled_at,
    recipient_count
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
)
RETURNING *;

-- name: GetBroadcastByID :one
SELECT * FROM broadcasts
WHERE broadcast_id = ?;

-- name: ListBroadcasts :many
SELECT * FROM broadcasts
ORDER BY created_at DESC;

-- name: ListBroadcastsWithSender :many
SELECT 
    b.broadcast_id,
    b.message,
    b.audience,
    b.sender_user_id,
    b.push_enabled,
    b.scheduled_at,
    b.sent_at,
    b.status,
    b.recipient_count,
    b.sent_count,
    b.failed_count,
    b.created_at,
    COALESCE(u.name, '') as sender_name
FROM broadcasts b
JOIN users u ON b.sender_user_id = u.user_id
ORDER BY b.created_at DESC;

-- name: UpdateBroadcastStatus :one
UPDATE broadcasts
SET
    status = ?,
    sent_at = ?,
    sent_count = ?,
    failed_count = ?
WHERE
    broadcast_id = ?
RETURNING *;

-- name: ListPendingBroadcasts :many
SELECT * FROM broadcasts
WHERE status = 'pending'
AND (scheduled_at IS NULL OR scheduled_at <= datetime('now'))
ORDER BY created_at ASC; 