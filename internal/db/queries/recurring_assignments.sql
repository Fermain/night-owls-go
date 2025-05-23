-- name: CreateRecurringAssignment :one
INSERT INTO recurring_assignments (
    user_id,
    buddy_name,
    day_of_week,
    schedule_id,
    time_slot,
    description
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
)
RETURNING *;

-- name: GetRecurringAssignmentByID :one
SELECT * FROM recurring_assignments
WHERE recurring_assignment_id = ?;

-- name: ListRecurringAssignments :many
SELECT * FROM recurring_assignments
WHERE is_active = 1
ORDER BY day_of_week, time_slot;

-- name: ListRecurringAssignmentsByUserID :many
SELECT * FROM recurring_assignments
WHERE user_id = ? AND is_active = 1
ORDER BY day_of_week, time_slot;

-- name: UpdateRecurringAssignment :one
UPDATE recurring_assignments
SET
    user_id = ?,
    buddy_name = ?,
    day_of_week = ?,
    schedule_id = ?,
    time_slot = ?,
    description = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE
    recurring_assignment_id = ?
RETURNING *;

-- name: DeleteRecurringAssignment :exec
UPDATE recurring_assignments
SET is_active = 0, updated_at = CURRENT_TIMESTAMP
WHERE recurring_assignment_id = ?;

-- name: GetRecurringAssignmentsByPattern :many
SELECT ra.*, u.name as user_name, u.phone as user_phone, s.name as schedule_name
FROM recurring_assignments ra
JOIN users u ON ra.user_id = u.user_id
JOIN schedules s ON ra.schedule_id = s.schedule_id
WHERE ra.day_of_week = ? 
  AND ra.schedule_id = ? 
  AND ra.time_slot = ? 
  AND ra.is_active = 1; 