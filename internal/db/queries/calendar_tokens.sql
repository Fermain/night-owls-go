-- Calendar Token Queries

-- name: CreateCalendarToken :one
INSERT INTO calendar_tokens (
    user_id,
    token_hash,
    expires_at
) VALUES (?, ?, ?)
RETURNING *;

-- name: GetCalendarTokenByHash :one
SELECT * FROM calendar_tokens 
WHERE token_hash = ? 
  AND is_revoked = 0 
  AND expires_at > datetime('now');

-- name: ValidateCalendarToken :one
SELECT ct.*, u.name as user_name 
FROM calendar_tokens ct
JOIN users u ON ct.user_id = u.user_id
WHERE ct.user_id = ? 
  AND ct.token_hash = ? 
  AND ct.is_revoked = 0 
  AND ct.expires_at > datetime('now');

-- name: UpdateTokenAccess :exec
UPDATE calendar_tokens 
SET last_accessed_at = datetime('now'),
    access_count = access_count + 1
WHERE token_hash = ?;

-- name: RevokeCalendarToken :exec
UPDATE calendar_tokens 
SET is_revoked = 1
WHERE user_id = ? AND token_hash = ?;

-- name: RevokeAllUserCalendarTokens :exec
UPDATE calendar_tokens 
SET is_revoked = 1
WHERE user_id = ?;

-- name: GetUserCalendarTokens :many
SELECT token_id, created_at, expires_at, last_accessed_at, access_count, is_revoked
FROM calendar_tokens 
WHERE user_id = ?
ORDER BY created_at DESC;

-- name: CleanupExpiredCalendarTokens :exec
DELETE FROM calendar_tokens 
WHERE expires_at < datetime('now', '-30 days'); -- Keep expired tokens for 30 days for audit

-- name: GetCalendarTokenStats :one
SELECT 
    COUNT(*) as total_tokens,
    COUNT(CASE WHEN is_revoked = 0 THEN 1 END) as active_tokens,
    COUNT(CASE WHEN expires_at < datetime('now') THEN 1 END) as expired_tokens,
    COUNT(CASE WHEN last_accessed_at IS NOT NULL THEN 1 END) as used_tokens
FROM calendar_tokens;