-- name: CreateBooking :one
INSERT INTO bookings (
    user_id,
    schedule_id,
    shift_start,
    shift_end,
    buddy_user_id,
    buddy_name
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
)
RETURNING *;

-- name: GetBookingByScheduleAndStartTime :one
SELECT * FROM bookings
WHERE schedule_id = ? AND shift_start = ?;

-- name: GetBookingByID :one
SELECT * FROM bookings
WHERE booking_id = ?;

-- name: ListBookingsByUserID :many
SELECT * FROM bookings
WHERE user_id = ?
ORDER BY shift_start DESC;

-- name: UpdateBookingAttendance :one
UPDATE bookings
SET attended = ?
WHERE booking_id = ?
RETURNING *; 