-- OTP Attempts Queries

-- name: CreateOTPAttempt :one
INSERT INTO otp_attempts (phone, attempted_at, success, client_ip, user_agent)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetOTPAttemptsInWindow :many
SELECT * FROM otp_attempts 
WHERE phone = ? AND attempted_at >= ?
ORDER BY attempted_at DESC;

-- name: GetFailedOTPAttemptsInWindow :one
SELECT COUNT(*) as failed_count
FROM otp_attempts 
WHERE phone = ? AND attempted_at >= ? AND success = 0;

-- name: CleanupOldOTPAttempts :exec
DELETE FROM otp_attempts 
WHERE created_at < ?;

-- OTP Rate Limits Queries

-- name: GetOTPRateLimit :one
SELECT * FROM otp_rate_limits 
WHERE phone = ?;

-- name: CreateOTPRateLimit :one
INSERT INTO otp_rate_limits (phone, failed_attempts, first_attempt_at, last_attempt_at)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: UpdateOTPRateLimit :exec
UPDATE otp_rate_limits 
SET failed_attempts = ?, 
    locked_until = ?, 
    last_attempt_at = ?, 
    updated_at = CURRENT_TIMESTAMP
WHERE phone = ?;

-- name: ResetOTPRateLimit :exec
UPDATE otp_rate_limits 
SET failed_attempts = 0, 
    locked_until = NULL, 
    updated_at = CURRENT_TIMESTAMP
WHERE phone = ?;

-- name: DeleteOTPRateLimit :exec
DELETE FROM otp_rate_limits 
WHERE phone = ?;

-- name: GetLockedPhones :many
SELECT phone, locked_until, failed_attempts 
FROM otp_rate_limits 
WHERE locked_until IS NOT NULL AND locked_until > CURRENT_TIMESTAMP;

-- name: CleanupExpiredLocks :exec
UPDATE otp_rate_limits 
SET failed_attempts = 0, locked_until = NULL, updated_at = CURRENT_TIMESTAMP
WHERE locked_until IS NOT NULL AND locked_until <= CURRENT_TIMESTAMP; 