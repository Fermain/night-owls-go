package api

import (
	db "night-owls-go/internal/db/sqlc_generated"
	"time"
)

// Converter functions to transform database models to API-friendly responses
// This allows us to use the database models as the source of truth while
// providing clean API documentation with Swagger

// ToBookingResponse converts a database Booking to an API-friendly response
func ToBookingResponse(booking db.Booking) BookingResponse {
	var buddyUserID *int64
	if booking.BuddyUserID.Valid {
		value := booking.BuddyUserID.Int64
		buddyUserID = &value
	}

	var buddyName string
	if booking.BuddyName.Valid {
		buddyName = booking.BuddyName.String
	}

	// Handle CreatedAt
	var createdAt time.Time
	if booking.CreatedAt.Valid {
		createdAt = booking.CreatedAt.Time
	} else {
		createdAt = time.Now() // Fallback, though this shouldn't happen
	}

	return BookingResponse{
		BookingID:    booking.BookingID,
		UserID:       booking.UserID,
		ScheduleID:   booking.ScheduleID,
		ShiftStart:   booking.ShiftStart,
		ShiftEnd:     booking.ShiftEnd,
		BuddyUserID:  buddyUserID,
		BuddyName:    buddyName,
		Attended:     booking.Attended,
		CreatedAt:    createdAt,
	}
}

// ToReportResponse converts a database Report to an API-friendly response
func ToReportResponse(report db.Report) ReportResponse {
	var message string
	if report.Message.Valid {
		message = report.Message.String
	}

	// Handle CreatedAt
	var createdAt time.Time
	if report.CreatedAt.Valid {
		createdAt = report.CreatedAt.Time
	} else {
		createdAt = time.Now() // Fallback, though this shouldn't happen
	}

	return ReportResponse{
		ReportID:  report.ReportID,
		BookingID: report.BookingID,
		Severity:  report.Severity,
		Message:   message,
		CreatedAt: createdAt,
	}
}

// ToScheduleResponse converts a database Schedule to an API-friendly response
func ToScheduleResponse(schedule db.Schedule) ScheduleResponse {
	var startDate, endDate *time.Time
	if schedule.StartDate.Valid {
		value := schedule.StartDate.Time
		startDate = &value
	}
	if schedule.EndDate.Valid {
		value := schedule.EndDate.Time
		endDate = &value
	}

	var timezone string
	if schedule.Timezone.Valid {
		timezone = schedule.Timezone.String
	}

	return ScheduleResponse{
		ScheduleID:      schedule.ScheduleID,
		Name:            schedule.Name,
		CronExpr:        schedule.CronExpr,
		StartDate:       startDate,
		EndDate:         endDate,
		DurationMinutes: schedule.DurationMinutes,
		Timezone:        timezone,
	}
}

// ToScheduleResponses converts a slice of database Schedules to API-friendly responses
func ToScheduleResponses(schedules []db.Schedule) []ScheduleResponse {
	responses := make([]ScheduleResponse, len(schedules))
	for i, schedule := range schedules {
		responses[i] = ToScheduleResponse(schedule)
	}
	return responses
}

// ToBookingResponses converts a slice of database Bookings to API-friendly responses
func ToBookingResponses(bookings []db.Booking) []BookingResponse {
	responses := make([]BookingResponse, len(bookings))
	for i, booking := range bookings {
		responses[i] = ToBookingResponse(booking)
	}
	return responses
} 