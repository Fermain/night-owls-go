-- name: UpsertSubscription :exec
INSERT INTO push_subscriptions (user_id, endpoint, p256dh_key, auth_key, user_agent, platform)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(endpoint) DO UPDATE
SET p256dh_key = excluded.p256dh_key,
    auth_key   = excluded.auth_key,
    user_agent = excluded.user_agent,
    platform   = excluded.platform;

-- name: DeleteSubscription :exec
DELETE FROM push_subscriptions WHERE endpoint = ? AND user_id = ?;

-- name: GetSubscriptionsByUser :many
SELECT endpoint, p256dh_key, auth_key FROM push_subscriptions WHERE user_id = ?; 