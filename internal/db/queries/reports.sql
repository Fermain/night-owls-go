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