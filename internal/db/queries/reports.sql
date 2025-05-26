-- name: CreateReport :one
INSERT INTO reports (
    booking_id,
    severity,
    message,
    archived_at
) VALUES (
    ?,
    ?,
    ?,
    NULL
)
RETURNING *;

-- name: GetReportByBookingID :one
SELECT * FROM reports
WHERE booking_id = ? AND archived_at IS NULL;

-- name: ListReportsByUserID :many
SELECT r.* 
FROM reports r
JOIN bookings b ON r.booking_id = b.booking_id
WHERE b.user_id = ? AND r.archived_at IS NULL
ORDER BY r.created_at DESC;

-- name: AdminGetReportWithContext :one
SELECT 
    r.report_id,
    r.booking_id,
    r.severity,
    r.message,
    r.created_at,
    r.archived_at,
    b.user_id,
    COALESCE(u.name, '') as user_name,
    u.phone as user_phone,
    b.schedule_id,
    s.name as schedule_name,
    b.shift_start,
    b.shift_end
FROM reports r
JOIN bookings b ON r.booking_id = b.booking_id
JOIN users u ON b.user_id = u.user_id
JOIN schedules s ON b.schedule_id = s.schedule_id
WHERE r.report_id = ?;

-- name: AdminListReportsWithContext :many
SELECT 
    r.report_id,
    r.booking_id,
    r.severity,
    r.message,
    r.created_at,
    r.archived_at,
    b.user_id,
    COALESCE(u.name, '') as user_name,
    u.phone as user_phone,
    b.schedule_id,
    s.name as schedule_name,
    b.shift_start,
    b.shift_end
FROM reports r
JOIN bookings b ON r.booking_id = b.booking_id
JOIN users u ON b.user_id = u.user_id
JOIN schedules s ON b.schedule_id = s.schedule_id
WHERE r.archived_at IS NULL
ORDER BY r.created_at DESC;

-- name: ArchiveReport :exec
UPDATE reports 
SET archived_at = CURRENT_TIMESTAMP 
WHERE report_id = ? AND archived_at IS NULL;

-- name: UnarchiveReport :exec
UPDATE reports 
SET archived_at = NULL 
WHERE report_id = ?;

-- name: AdminListArchivedReportsWithContext :many
SELECT 
    r.report_id,
    r.booking_id,
    r.severity,
    r.message,
    r.created_at,
    r.archived_at,
    b.user_id,
    COALESCE(u.name, '') as user_name,
    u.phone as user_phone,
    b.schedule_id,
    s.name as schedule_name,
    b.shift_start,
    b.shift_end
FROM reports r
JOIN bookings b ON r.booking_id = b.booking_id
JOIN users u ON b.user_id = u.user_id
JOIN schedules s ON b.schedule_id = s.schedule_id
WHERE r.archived_at IS NOT NULL
ORDER BY r.archived_at DESC;

-- name: GetReportsForAutoArchiving :many
SELECT report_id, severity, created_at
FROM reports 
WHERE archived_at IS NULL 
AND (
    -- Info reports older than 1 month
    (severity = 0 AND created_at < datetime('now', '-1 month'))
    OR
    -- Warning reports older than 1 year  
    (severity = 1 AND created_at < datetime('now', '-1 year'))
    -- Critical reports (severity = 2) are never auto-archived
);

-- name: BulkArchiveReports :exec
UPDATE reports 
SET archived_at = CURRENT_TIMESTAMP 
WHERE report_id IN (sqlc.slice('report_ids')) AND archived_at IS NULL; 