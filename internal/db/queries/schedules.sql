-- name: CreateSchedule :one
INSERT INTO schedules (
    name,
    cron_expr,
    start_date,
    end_date,
    duration_minutes,
    timezone
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
)
RETURNING *;

-- name: GetScheduleByID :one
SELECT * FROM schedules
WHERE schedule_id = ?;

-- name: ListActiveSchedules :many
SELECT * FROM schedules
WHERE 
    (start_date IS NULL OR date(?) >= start_date) 
AND 
    (end_date IS NULL OR date(?) <= end_date)
ORDER BY name;

-- name: ListAllSchedules :many
SELECT * FROM schedules
ORDER BY name;

-- name: UpdateSchedule :one
UPDATE schedules
SET
    name = ?,
    cron_expr = ?,
    start_date = ?,
    end_date = ?,
    duration_minutes = ?,
    timezone = ?
WHERE
    schedule_id = ?
RETURNING *;

-- name: DeleteSchedule :exec
DELETE FROM schedules
WHERE schedule_id = ?;

-- name: AdminBulkDeleteSchedules :exec
DELETE FROM schedules
WHERE schedule_id IN (sqlc.slice('schedule_ids')); 