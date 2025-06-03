-- name: CreateAuditEvent :one
INSERT INTO audit_events (
    event_type,
    actor_user_id,
    target_user_id,
    entity_type,
    entity_id,
    action,
    details,
    ip_address,
    user_agent
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
)
RETURNING *;

-- name: ListAuditEvents :many
SELECT 
    ae.*,
    COALESCE(actor.name, '') as actor_name,
    actor.phone as actor_phone,
    COALESCE(target.name, '') as target_name,
    target.phone as target_phone
FROM audit_events ae
LEFT JOIN users actor ON ae.actor_user_id = actor.user_id
LEFT JOIN users target ON ae.target_user_id = target.user_id
ORDER BY ae.created_at DESC
LIMIT ? OFFSET ?;

-- name: ListAuditEventsByType :many
SELECT 
    ae.*,
    COALESCE(actor.name, '') as actor_name,
    actor.phone as actor_phone,
    COALESCE(target.name, '') as target_name,
    target.phone as target_phone
FROM audit_events ae
LEFT JOIN users actor ON ae.actor_user_id = actor.user_id
LEFT JOIN users target ON ae.target_user_id = target.user_id
WHERE ae.event_type = ?
ORDER BY ae.created_at DESC
LIMIT ? OFFSET ?;

-- name: ListAuditEventsByActor :many
SELECT 
    ae.*,
    COALESCE(actor.name, '') as actor_name,
    actor.phone as actor_phone,
    COALESCE(target.name, '') as target_name,
    target.phone as target_phone
FROM audit_events ae
LEFT JOIN users actor ON ae.actor_user_id = actor.user_id
LEFT JOIN users target ON ae.target_user_id = target.user_id
WHERE ae.actor_user_id = ?
ORDER BY ae.created_at DESC
LIMIT ? OFFSET ?;

-- name: ListAuditEventsByTarget :many
SELECT 
    ae.*,
    COALESCE(actor.name, '') as actor_name,
    actor.phone as actor_phone,
    COALESCE(target.name, '') as target_name,
    target.phone as target_phone
FROM audit_events ae
LEFT JOIN users actor ON ae.actor_user_id = actor.user_id
LEFT JOIN users target ON ae.target_user_id = target.user_id
WHERE ae.target_user_id = ?
ORDER BY ae.created_at DESC
LIMIT ? OFFSET ?;

-- name: ListAuditEventsByDateRange :many
SELECT 
    ae.*,
    COALESCE(actor.name, '') as actor_name,
    actor.phone as actor_phone,
    COALESCE(target.name, '') as target_name,
    target.phone as target_phone
FROM audit_events ae
LEFT JOIN users actor ON ae.actor_user_id = actor.user_id
LEFT JOIN users target ON ae.target_user_id = target.user_id
WHERE ae.created_at >= ? AND ae.created_at <= ?
ORDER BY ae.created_at DESC
LIMIT ? OFFSET ?;

-- name: GetAuditEventStats :one
SELECT 
    COUNT(*) as total_events,
    COUNT(DISTINCT ae.actor_user_id) as unique_actors,
    COUNT(DISTINCT ae.event_type) as unique_event_types,
    MIN(ae.created_at) as earliest_event,
    MAX(ae.created_at) as latest_event
FROM audit_events ae;

-- name: GetAuditEventsByTypeStats :many
SELECT 
    event_type,
    COUNT(*) as event_count,
    MAX(created_at) as latest_event
FROM audit_events 
GROUP BY event_type
ORDER BY event_count DESC; 