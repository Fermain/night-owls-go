package api

import (
	"time"
)

// Models for Swagger documentation
// These are simplified versions of our database models without SQL-specific types

// BookingResponse represents a booking in the API
type BookingResponse struct {
	BookingID    int64      `json:"booking_id"`
	UserID       int64      `json:"user_id"`
	ScheduleID   int64      `json:"schedule_id"`
	ShiftStart   time.Time  `json:"shift_start"`
	ShiftEnd     time.Time  `json:"shift_end"`
	BuddyUserID  *int64     `json:"buddy_user_id,omitempty"`
	BuddyName    string     `json:"buddy_name,omitempty"`
	CheckedInAt  *time.Time `json:"checked_in_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
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
	ReportID     int64     `json:"report_id"`
	BookingID    int64     `json:"booking_id"`
	Severity     int64     `json:"severity"`
	Message      string    `json:"message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// ScheduleResponse represents a schedule in the API
type ScheduleResponse struct {
	ScheduleID      int64     `json:"schedule_id"`
	Name            string    `json:"name"`
	CronExpr        string    `json:"cron_expr"`
	StartDate       *string   `json:"start_date,omitempty"`
	EndDate         *string   `json:"end_date,omitempty"`
	DurationMinutes int64     `json:"duration_minutes"`
	Timezone        string    `json:"timezone,omitempty"`
} 