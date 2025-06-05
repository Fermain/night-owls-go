-- Points System Queries

-- name: AwardPoints :exec
-- Award points to a user for a specific reason
INSERT INTO points_history (user_id, booking_id, points_awarded, reason, multiplier)
VALUES (?, ?, ?, ?, ?);

-- name: UpdateUserTotalPoints :exec
-- Update user's total points (should be called after AwardPoints)
UPDATE users 
SET total_points = (
    SELECT COALESCE(SUM(ph.points_awarded * ph.multiplier), 0) 
    FROM points_history ph
    WHERE ph.user_id = ?
)
WHERE users.user_id = ?;

-- name: UpdateUserShiftCount :exec
-- Update user's shift count and last activity
UPDATE users 
SET shift_count = shift_count + 1,
    last_activity_date = DATE('now')
WHERE user_id = ?;

-- name: GetUserPoints :one
-- Get a user's current points and shift information
SELECT 
    user_id,
    name,
    total_points,
    shift_count,
    last_activity_date
FROM users 
WHERE user_id = ?;

-- name: GetTopUsers :many
-- Get leaderboard of top users by points
SELECT 
    u.user_id,
    u.name,
    u.total_points,
    u.shift_count,
    COUNT(DISTINCT ua.achievement_id) as achievement_count,
    -- Recent activity indicator
    CASE 
        WHEN u.last_activity_date >= DATE('now', '-7 days') THEN 'active'
        WHEN u.last_activity_date >= DATE('now', '-30 days') THEN 'moderate' 
        ELSE 'inactive'
    END as activity_status
FROM users u
LEFT JOIN user_achievements ua ON u.user_id = ua.user_id
WHERE u.role IN ('admin', 'owl') AND u.total_points > 0
GROUP BY u.user_id, u.name, u.total_points, u.shift_count, u.last_activity_date
ORDER BY u.total_points DESC, u.shift_count DESC
LIMIT ?;

-- name: GetTopUsersByShifts :many
-- Get leaderboard of top users by shift count
SELECT 
    u.user_id,
    u.name,
    u.total_points,
    u.shift_count,
    COUNT(DISTINCT ua.achievement_id) as achievement_count,
    -- Recent activity indicator
    CASE 
        WHEN u.last_activity_date >= DATE('now', '-7 days') THEN 'active'
        WHEN u.last_activity_date >= DATE('now', '-30 days') THEN 'moderate' 
        ELSE 'inactive'
    END as activity_status
FROM users u
LEFT JOIN user_achievements ua ON u.user_id = ua.user_id
WHERE u.role IN ('admin', 'owl') AND u.shift_count > 0
GROUP BY u.user_id, u.name, u.total_points, u.shift_count, u.last_activity_date
ORDER BY u.shift_count DESC, u.total_points DESC
LIMIT ?;

-- name: GetUserRank :one
-- Get a specific user's rank by points
SELECT 
    COUNT(*) + 1 as user_rank
FROM users 
WHERE role IN ('admin', 'owl') 
    AND total_points > (SELECT u2.total_points FROM users u2 WHERE u2.user_id = ?)
    AND total_points > 0;

-- name: GetUserPointsHistory :many
-- Get recent points history for a user
SELECT 
    ph.points_awarded,
    ph.reason,
    ph.multiplier,
    ph.created_at,
    b.shift_start
FROM points_history ph
LEFT JOIN bookings b ON ph.booking_id = b.booking_id
WHERE ph.user_id = ?
ORDER BY ph.created_at DESC
LIMIT ?;

-- name: GetUserAchievements :many
-- Get all achievements earned by a user
SELECT 
    a.achievement_id,
    a.name,
    a.description,
    a.icon,
    ua.earned_at
FROM achievements a
JOIN user_achievements ua ON a.achievement_id = ua.achievement_id
WHERE ua.user_id = ?
ORDER BY ua.earned_at DESC;

-- name: GetAvailableAchievements :many
-- Get achievements a user hasn't earned yet
SELECT 
    a.achievement_id,
    a.name,
    a.description,
    a.icon,
    a.shifts_threshold
FROM achievements a
WHERE a.achievement_id NOT IN (
    SELECT ua.achievement_id 
    FROM user_achievements ua 
    WHERE ua.user_id = ?
)
ORDER BY a.shifts_threshold;

-- name: AwardAchievement :exec
-- Award an achievement to a user
INSERT OR IGNORE INTO user_achievements (user_id, achievement_id)
VALUES (?, ?);

-- name: GetRecentActivity :many
-- Get recent point-earning activities across all users for activity feed
SELECT 
    u.name,
    ph.points_awarded,
    ph.reason,
    ph.created_at,
    -- Anonymize for privacy while keeping engagement
    CASE 
        WHEN ph.points_awarded >= 50 THEN 'major'
        WHEN ph.points_awarded >= 20 THEN 'significant'
        ELSE 'standard'
    END as activity_type
FROM points_history ph
JOIN users u ON ph.user_id = u.user_id
WHERE ph.created_at >= datetime('now', '-24 hours')
    AND u.role IN ('admin', 'owl')
ORDER BY ph.created_at DESC
LIMIT ?;

-- name: GetUserShiftCountForMonth :one
-- Get the number of shifts a user has completed in a specific month
SELECT 
    COUNT(*) as shift_count
FROM bookings b
WHERE b.user_id = ?
    AND b.checked_in_at IS NOT NULL
    AND strftime('%Y-%m', b.shift_start) = ?; 