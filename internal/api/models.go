package api

import (
	"time"
)

// Models for Swagger documentation
// These are simplified versions of our database models without SQL-specific types

// BookingResponse represents a booking in the API
type BookingResponse struct {
	BookingID   int64      `json:"booking_id"`
	UserID      int64      `json:"user_id"`
	ScheduleID  int64      `json:"schedule_id"`
	ShiftStart  time.Time  `json:"shift_start"`
	ShiftEnd    time.Time  `json:"shift_end"`
	BuddyUserID *int64     `json:"buddy_user_id,omitempty"`
	BuddyName   string     `json:"buddy_name,omitempty"`
	CheckedInAt *time.Time `json:"checked_in_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// BookingWithScheduleResponse includes schedule name for admin views
type BookingWithScheduleResponse struct {
	BookingID    int64      `json:"booking_id"`
	UserID       int64      `json:"user_id"`
	ScheduleID   int64      `json:"schedule_id"`
	ScheduleName string     `json:"schedule_name"`
	ShiftStart   time.Time  `json:"shift_start"`
	ShiftEnd     time.Time  `json:"shift_end"`
	BuddyUserID  *int64     `json:"buddy_user_id,omitempty"`
	BuddyName    string     `json:"buddy_name,omitempty"`
	CheckedInAt  *time.Time `json:"checked_in_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// ReportResponse represents a report in the API
type ReportResponse struct {
	ReportID  int64     `json:"report_id"`
	BookingID int64     `json:"booking_id"`
	Severity  int64     `json:"severity"`
	Message   string    `json:"message,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ScheduleResponse represents a schedule in the API
type ScheduleResponse struct {
	ScheduleID      int64   `json:"schedule_id"`
	Name            string  `json:"name"`
	CronExpr        string  `json:"cron_expr"`
	StartDate       *string `json:"start_date,omitempty"`
	EndDate         *string `json:"end_date,omitempty"`
	DurationMinutes int64   `json:"duration_minutes"`
	Timezone        string  `json:"timezone,omitempty"`
}

// LeaderboardEntry represents a user's position on the leaderboard
type LeaderboardEntry struct {
	UserID           int64  `json:"user_id"`
	Name             string `json:"name"`
	TotalPoints      int64  `json:"total_points"`
	ShiftCount       int64  `json:"shift_count"`
	Rank             int64  `json:"rank"`
	AchievementCount int64  `json:"achievement_count"`
	ActivityStatus   string `json:"activity_status"` // 'active', 'moderate', 'inactive'
}

// UserStatsResponse represents a user's complete points and achievement stats
type UserStatsResponse struct {
	UserID           int64   `json:"user_id"`
	Name             string  `json:"name"`
	TotalPoints      int64   `json:"total_points"`
	ShiftCount       int64   `json:"shift_count"`
	LastActivityDate *string `json:"last_activity_date,omitempty"`
	Rank             int64   `json:"rank"`
}

// PointsHistoryEntry represents a single points transaction
type PointsHistoryEntry struct {
	PointsAwarded int64      `json:"points_awarded"`
	Reason        string     `json:"reason"`
	Multiplier    float64    `json:"multiplier"`
	CreatedAt     time.Time  `json:"created_at"`
	ShiftStart    *time.Time `json:"shift_start,omitempty"`
}

// AchievementResponse represents an achievement badge
type AchievementResponse struct {
	AchievementID   int64      `json:"achievement_id"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	Icon            string     `json:"icon"`
	ShiftsThreshold *int64     `json:"shifts_threshold,omitempty"`
	EarnedAt        *time.Time `json:"earned_at,omitempty"`
}

// ActivityFeedEntry represents recent point-earning activities
type ActivityFeedEntry struct {
	UserName      string    `json:"user_name"`
	PointsAwarded int64     `json:"points_awarded"`
	Reason        string    `json:"reason"`
	ActivityType  string    `json:"activity_type"` // 'major', 'significant', 'standard'
	CreatedAt     time.Time `json:"created_at"`
}
