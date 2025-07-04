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

-- name: ListBookingsByUserIDWithSchedule :many
SELECT 
    b.booking_id,
    b.user_id,
    b.schedule_id,
    b.shift_start,
    b.shift_end,
    b.buddy_user_id,
    b.buddy_name,
    b.checked_in_at,
    b.created_at,
    s.name as schedule_name
FROM bookings b
JOIN schedules s ON b.schedule_id = s.schedule_id
WHERE b.user_id = ?
ORDER BY b.shift_start DESC;

-- name: UpdateBookingCheckIn :one
UPDATE bookings
SET checked_in_at = ?
WHERE booking_id = ?
RETURNING *;

-- name: DeleteBooking :exec
DELETE FROM bookings
WHERE booking_id = ?;

-- Admin Dashboard Metrics Queries

-- name: GetBookingMetrics :one
-- Get booking-based metrics for dashboard (will be combined with slot data in Go)
SELECT 
    COUNT(*) as total_bookings,
    COUNT(b.checked_in_at) as checked_in_bookings,
    COUNT(r.report_id) as completed_bookings,
    CAST(ROUND((CAST(COUNT(b.checked_in_at) AS FLOAT) / COUNT(*)) * 100, 1) AS REAL) as check_in_rate,
    COALESCE(CAST(ROUND((CAST(COUNT(r.report_id) AS FLOAT) / NULLIF(COUNT(b.checked_in_at), 0)) * 100, 1) AS REAL), 0.0) as completion_rate
FROM bookings b
LEFT JOIN reports r ON b.booking_id = r.booking_id
WHERE b.shift_start >= ? AND b.shift_start <= ?;

-- name: GetMemberContributions :many
-- Get member contribution analysis for the past month
SELECT 
    u.user_id,
    u.name,
    u.phone,
    COUNT(b.booking_id) as shifts_booked,
    COUNT(b.checked_in_at) as shifts_attended,
    COUNT(r.report_id) as shifts_completed,
    COALESCE(CAST(ROUND((CAST(COUNT(b.checked_in_at) AS FLOAT) / NULLIF(COUNT(b.booking_id), 0)) * 100, 1) AS REAL), 0.0) as attendance_rate,
    COALESCE(CAST(ROUND((CAST(COUNT(r.report_id) AS FLOAT) / NULLIF(COUNT(b.checked_in_at), 0)) * 100, 1) AS REAL), 0.0) as completion_rate,
    MAX(b.shift_start) as last_shift_date,
    CASE 
        WHEN COUNT(b.booking_id) = 0 THEN 'non_contributor'
        WHEN COUNT(b.booking_id) = 1 THEN 'minimum_contributor' 
        WHEN COUNT(b.booking_id) >= 2 THEN 'fair_contributor'
        WHEN COUNT(b.booking_id) >= 3 THEN 'heavy_lifter'
    END as contribution_category
FROM users u
LEFT JOIN bookings b ON u.user_id = b.user_id 
    AND b.shift_start >= datetime('now', '-30 days')
    AND b.shift_start <= datetime('now')
LEFT JOIN reports r ON b.booking_id = r.booking_id
WHERE u.role IN ('admin', 'owl')
GROUP BY u.user_id, u.name, u.phone
ORDER BY shifts_booked DESC, shifts_completed DESC;

-- name: GetBookingsInDateRange :many
-- Get all bookings in date range with check-in and report status
SELECT 
    b.booking_id,
    b.user_id,
    b.schedule_id,
    b.shift_start,
    b.shift_end,
    b.checked_in_at,
    b.buddy_name,
    COALESCE(u.name, '') as user_name,
    u.phone as user_phone,
    s.name as schedule_name,
    CASE WHEN r.report_id IS NOT NULL THEN 1 ELSE 0 END as has_report,
    CAST(julianday(b.shift_start) - julianday('now') AS INTEGER) as days_from_now,
    CASE 
        WHEN datetime(b.shift_start) <= datetime('now', '+1 day') THEN 'urgent'
        WHEN datetime(b.shift_start) <= datetime('now', '+3 days') THEN 'critical'
        ELSE 'normal'
    END as urgency_level
FROM bookings b
JOIN users u ON b.user_id = u.user_id
JOIN schedules s ON b.schedule_id = s.schedule_id
LEFT JOIN reports r ON b.booking_id = r.booking_id
WHERE b.shift_start >= ? AND b.shift_start <= ?
ORDER BY b.shift_start ASC;

-- name: GetBookingPatternsByTimeSlot :many  
-- Get booking patterns by time slot from historical data
SELECT 
    strftime('%w', b.shift_start) as day_of_week,
    strftime('%H', b.shift_start) as hour_of_day,
    COUNT(*) as total_bookings,
    COUNT(b.checked_in_at) as checked_in_bookings,
    COUNT(r.report_id) as completed_bookings,
    CAST(ROUND((CAST(COUNT(b.checked_in_at) AS FLOAT) / COUNT(*)) * 100, 1) AS REAL) as check_in_rate,
    COALESCE(CAST(ROUND((CAST(COUNT(r.report_id) AS FLOAT) / NULLIF(COUNT(b.checked_in_at), 0)) * 100, 1) AS REAL), 0.0) as completion_rate
FROM bookings b
LEFT JOIN reports r ON b.booking_id = r.booking_id
WHERE b.shift_start >= datetime('now', '-60 days')
    AND b.shift_start <= datetime('now')
GROUP BY day_of_week, hour_of_day
HAVING total_bookings >= 3
ORDER BY check_in_rate ASC, completion_rate ASC; 