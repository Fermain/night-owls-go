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

// CalendarData represents calendar file information for downloads
type CalendarData struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
	MimeType string `json:"mime_type"`
}

// CalendarFeedToken represents a secure WebCal feed access token
type CalendarFeedToken struct {
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// CalendarFeedResponse contains WebCal subscription information
type CalendarFeedResponse struct {
	FeedURL     string    `json:"feed_url"`
	WebCalURL   string    `json:"webcal_url"`
	Token       string    `json:"token"`
	ExpiresAt   time.Time `json:"expires_at"`
	Description string    `json:"description"`
}

// PublicShiftSlot represents a shift slot for public schedule viewing
// Provides community visibility while protecting user privacy
type PublicShiftSlot struct {
	ScheduleID   int64   `json:"schedule_id"`
	ScheduleName string  `json:"schedule_name"`
	StartTime    string  `json:"start_time"`
	EndTime      string  `json:"end_time"`
	Timezone     *string `json:"timezone,omitempty"`
	IsBooked     bool    `json:"is_booked"`
	BookedBy     *string `json:"booked_by,omitempty"`  // Privacy-masked: "John D." or "Booked"
	HasBuddy     bool    `json:"has_buddy"`            // Boolean instead of buddy name
	BookingID    *int64  `json:"booking_id,omitempty"` // For frontend compatibility
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
