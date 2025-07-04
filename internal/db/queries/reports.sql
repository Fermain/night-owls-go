-- name: CreateReport :one
INSERT INTO reports (
    booking_id,
    user_id,
    severity,
    message,
    latitude,
    longitude,
    gps_accuracy,
    gps_timestamp,
    archived_at
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    NULL
)
RETURNING *;

-- name: CreateOffShiftReport :one
INSERT INTO reports (
    user_id,
    severity,
    message,
    latitude,
    longitude,
    gps_accuracy,
    gps_timestamp,
    archived_at
) VALUES (
    ?,
    ?,
    ?,
    ?,
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
WHERE r.user_id = ? AND r.archived_at IS NULL
ORDER BY r.created_at DESC;

-- name: AdminGetReportWithContext :one
SELECT 
    r.report_id,
    r.booking_id,
    r.user_id,
    r.severity,
    r.message,
    r.created_at,
    r.archived_at,
    r.latitude,
    r.longitude,
    r.gps_accuracy,
    r.gps_timestamp,
    COALESCE(u.name, '') as user_name,
    u.phone as user_phone,
    COALESCE(b.schedule_id, 0) as schedule_id,
    COALESCE(s.name, 'Off-Shift Report') as schedule_name,
    COALESCE(datetime(b.shift_start), datetime(r.created_at)) as shift_start,
    COALESCE(datetime(b.shift_end), datetime(r.created_at)) as shift_end
FROM reports r
JOIN users u ON r.user_id = u.user_id
LEFT JOIN bookings b ON r.booking_id = b.booking_id
LEFT JOIN schedules s ON b.schedule_id = s.schedule_id
WHERE r.report_id = ?;

-- name: AdminListReportsWithContext :many
SELECT 
    r.report_id,
    r.booking_id,
    r.user_id,
    r.severity,
    r.message,
    r.created_at,
    r.archived_at,
    r.latitude,
    r.longitude,
    r.gps_accuracy,
    r.gps_timestamp,
    COALESCE(u.name, '') as user_name,
    u.phone as user_phone,
    COALESCE(b.schedule_id, 0) as schedule_id,
    COALESCE(s.name, 'Off-Shift Report') as schedule_name,
    COALESCE(datetime(b.shift_start), datetime(r.created_at)) as shift_start,
    COALESCE(datetime(b.shift_end), datetime(r.created_at)) as shift_end
FROM reports r
JOIN users u ON r.user_id = u.user_id
LEFT JOIN bookings b ON r.booking_id = b.booking_id
LEFT JOIN schedules s ON b.schedule_id = s.schedule_id
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
    r.user_id,
    r.severity,
    r.message,
    r.created_at,
    r.archived_at,
    r.latitude,
    r.longitude,
    r.gps_accuracy,
    r.gps_timestamp,
    COALESCE(u.name, '') as user_name,
    u.phone as user_phone,
    COALESCE(b.schedule_id, 0) as schedule_id,
    COALESCE(s.name, 'Off-Shift Report') as schedule_name,
    COALESCE(datetime(b.shift_start), datetime(r.created_at)) as shift_start,
    COALESCE(datetime(b.shift_end), datetime(r.created_at)) as shift_end
FROM reports r
JOIN users u ON r.user_id = u.user_id
LEFT JOIN bookings b ON r.booking_id = b.booking_id
LEFT JOIN schedules s ON b.schedule_id = s.schedule_id
WHERE r.archived_at IS NOT NULL
ORDER BY r.archived_at DESC;

-- name: GetReportsForAutoArchiving :many
SELECT report_id, severity, created_at
FROM reports 
WHERE archived_at IS NULL 
AND (
    -- Normal reports older than 1 month
    (severity = 0 AND created_at < datetime('now', '-1 month'))
    OR
    -- Suspicion reports older than 1 year  
    (severity = 1 AND created_at < datetime('now', '-1 year'))
    -- Incident reports (severity = 2) are never auto-archived
);

-- name: BulkArchiveReports :exec
UPDATE reports 
SET archived_at = CURRENT_TIMESTAMP 
WHERE report_id IN (sqlc.slice('report_ids')) AND archived_at IS NULL;

-- name: DeleteReport :exec
DELETE FROM reports 
WHERE report_id = ?;

-- name: UpdateReportPhotoCount :exec
UPDATE reports 
SET photo_count = (
    SELECT COUNT(*) FROM report_photos WHERE report_photos.report_id = reports.report_id
)
WHERE reports.report_id = ?;

-- Photo operations
-- name: CreateReportPhoto :one
INSERT INTO report_photos (
    report_id,
    filename,
    original_filename,
    file_size_bytes,
    mime_type,
    width_pixels,
    height_pixels,
    storage_path,
    thumbnail_path,
    checksum_sha256,
    is_processed
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING *;

-- name: GetReportPhotos :many
SELECT * FROM report_photos 
WHERE report_id = ? 
ORDER BY upload_timestamp ASC;

-- name: GetReportPhoto :one
SELECT * FROM report_photos 
WHERE photo_id = ? AND report_id = ?;

-- name: DeleteReportPhoto :exec
DELETE FROM report_photos 
WHERE photo_id = ? AND report_id = ?; 