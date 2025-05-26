-- name: CreateReport :one
INSERT INTO reports (
    booking_id,
    severity,
    message
) VALUES (
    ?,
    ?,
    ?
)
RETURNING *;

-- name: GetReportByBookingID :one
SELECT * FROM reports
WHERE booking_id = ?;

-- name: ListReportsByUserID :many
SELECT r.* 
FROM reports r
JOIN bookings b ON r.booking_id = b.booking_id
WHERE b.user_id = ?
ORDER BY r.created_at DESC;

-- name: AdminGetReportWithContext :one
SELECT 
    r.report_id,
    r.booking_id,
    r.severity,
    r.message,
    r.created_at,
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
ORDER BY r.created_at DESC; 