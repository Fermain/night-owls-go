package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"night-owls-go/internal/config"
	db "night-owls-go/internal/db/sqlc_generated"

	"github.com/gorhill/cronexpr"
)

var (
	ErrScheduleNotFound      = errors.New("schedule not found")
	ErrShiftTimeInvalid    = errors.New("requested shift time is invalid for the schedule or outside its active window")
	ErrBookingConflict       = errors.New("shift slot is already booked") // Corresponds to HTTP 409
	ErrBookingNotFound     = errors.New("booking not found")
	ErrForbiddenUpdate     = errors.New("user not authorized to update this booking")
)

// BookingService handles logic related to bookings.
type BookingService struct {
	querier db.Querier
	cfg     *config.Config
	logger  *slog.Logger
}

// NewBookingService creates a new BookingService.
func NewBookingService(querier db.Querier, cfg *config.Config, logger *slog.Logger) *BookingService {
	return &BookingService{
		querier: querier,
		cfg:     cfg,
		logger:  logger.With("service", "BookingService"),
	}
}

// CreateBooking handles the logic for creating a new booking.
func (s *BookingService) CreateBooking(ctx context.Context, userID int64, scheduleID int64, startTime time.Time, buddyPhone, buddyName sql.NullString) (db.Booking, error) {
	// 1. Validate schedule and start time
	schedule, err := s.querier.GetScheduleByID(ctx, scheduleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.WarnContext(ctx, "Schedule not found for booking attempt", "schedule_id", scheduleID)
			return db.Booking{}, ErrScheduleNotFound
		}
		s.logger.ErrorContext(ctx, "Failed to get schedule by ID for booking", "schedule_id", scheduleID, "error", err)
		return db.Booking{}, ErrInternalServer
	}

	// Check if startTime is within the schedule's overall active period (start_date and end_date)
	if (schedule.StartDate.Valid && startTime.Before(schedule.StartDate.Time)) || 
	   (schedule.EndDate.Valid && startTime.After(schedule.EndDate.Time)) {
		s.logger.WarnContext(ctx, "Booking attempt for shift time outside schedule active dates", 
			"schedule_id", scheduleID, "start_time", startTime, 
			"schedule_start_date", schedule.StartDate, "schedule_end_date", schedule.EndDate)
		return db.Booking{}, ErrShiftTimeInvalid
	}

	// Validate startTime against the schedule's cron expression
	cronExpression, err := cronexpr.Parse(schedule.CronExpr)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to parse cron expression for schedule during booking", "schedule_id", scheduleID, "cron_expr", schedule.CronExpr, "error", err)
		return db.Booking{}, ErrInternalServer // This is a data integrity issue with the schedule itself
	}

	// Check if the provided startTime is an actual occurrence of the cron expression.
	// We find the next occurrence *from one second before* the startTime.
	// If this next occurrence is not exactly startTime, then startTime is not a valid point.
	// This also implicitly handles that startTime must not be in the past relative to cron's cycle if not careful,
	// but GetAvailableSlots should prevent past times.
	nextOccurrenceFromAlmostStartTime := cronExpression.Next(startTime.Add(-1 * time.Second))
	if nextOccurrenceFromAlmostStartTime.IsZero() || !nextOccurrenceFromAlmostStartTime.Equal(startTime) {
		s.logger.WarnContext(ctx, "Requested start_time does not match a cron expression occurrence", 
			"schedule_id", scheduleID, "start_time", startTime, "cron_expr", schedule.CronExpr, "calculated_next", nextOccurrenceFromAlmostStartTime)
		return db.Booking{}, ErrShiftTimeInvalid
	}

	// 2. Calculate shift_end
	shiftEndTime := startTime.Add(time.Duration(schedule.DurationMinutes) * time.Minute)

	// 3. Handle buddy logic
	var buddyUserID sql.NullInt64
	actualBuddyName := buddyName // Use provided buddyName by default

	if buddyPhone.Valid && buddyPhone.String != "" {
		buddyUser, err := s.querier.GetUserByPhone(ctx, buddyPhone.String)
		if err == nil {
			// Buddy is a registered user
			buddyUserID.Int64 = buddyUser.UserID
			buddyUserID.Valid = true
			if buddyUser.Name.Valid && buddyUser.Name.String != "" {
				actualBuddyName.String = buddyUser.Name.String // Prefer registered name
				actualBuddyName.Valid = true
			} else if !actualBuddyName.Valid { // if no name provided in request and user has no registered name
                 actualBuddyName.Valid = false // ensure it remains null
            }
		} else if !errors.Is(err, sql.ErrNoRows) {
			s.logger.ErrorContext(ctx, "Error looking up buddy by phone", "buddy_phone", buddyPhone.String, "error", err)
			// Non-fatal, proceed with buddyName if provided
		}
	}

	// 4. Insert booking into DB
	bookingParams := db.CreateBookingParams{
		UserID:      userID,
		ScheduleID:  scheduleID,
		ShiftStart:  startTime,
		ShiftEnd:    shiftEndTime,
		BuddyUserID: buddyUserID,
		BuddyName:   actualBuddyName,
	}

	createdBooking, err := s.querier.CreateBooking(ctx, bookingParams)
	if err != nil {
		// Check for unique constraint violation (duplicate booking for same schedule_id and shift_start)
		// SQLite error code for unique constraint is 1555 (SQLITE_CONSTRAINT_UNIQUE) or 2067 (SQLITE_CONSTRAINT_PRIMARYKEY on an UPSERT) or 19
		// A more portable way might be to check the error string if driver allows, or rely on specific DB error types.
		// For now, assuming sqlc or the driver might wrap this in a recognizable way, or we catch a generic error and let handler map to 409.
		// The Guide.md suggests the DB unique constraint will cause the insert to fail, and we catch that error to return 409.
		// This often manifests as a generic error containing "UNIQUE constraint failed".
		s.logger.ErrorContext(ctx, "Failed to create booking in DB", "params", bookingParams, "error", err)
		// A simple check for now, this might need refinement based on actual errors from sqlite driver
        if isUniqueConstraintError(err) { // Renamed placeholder for actual error check
            return db.Booking{}, ErrBookingConflict
        }
		return db.Booking{}, ErrInternalServer
	}
	s.logger.InfoContext(ctx, "Booking created successfully", "booking_id", createdBooking.BookingID, "user_id", userID)

	// 5. Queue confirmation message to outbox
	outboxPayload := fmt.Sprintf(`{"booking_id": %d, "user_id": %d, "shift_start": "%s"}`, 
		createdBooking.BookingID, createdBooking.UserID, createdBooking.ShiftStart.Format(time.RFC3339))
	_, err = s.querier.CreateOutboxItem(ctx, db.CreateOutboxItemParams{
		MessageType: "BOOKING_CONFIRMATION",
		Recipient:   fmt.Sprintf("%d", createdBooking.UserID), // Could be phone number or user ID
		Payload:     sql.NullString{String: outboxPayload, Valid: true},
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to create outbox item for booking confirmation", "booking_id", createdBooking.BookingID, "error", err)
		// Non-fatal for booking creation itself, but log it.
	}

	return createdBooking, nil
}

// isUniqueConstraintError checks if the error is likely a unique constraint violation.
func isUniqueConstraintError(err error) bool {
    if err == nil { return false }
    // This is a simplistic check. Real applications might use specific error codes or types
    // provided by the database driver or an ORM.
    // For SQLite with mattn/go-sqlite3, check for sqlite3.Error.Code == sqlite3.CONSTRAINT_UNIQUE (19 or 1555 or 2067)
    // or if err.Error() contains "UNIQUE constraint failed"
    // For now, a simple string check is used.
    return strings.Contains(err.Error(), "UNIQUE constraint failed")
}


// MarkAttendance handles marking a booking as attended or not.
func (s *BookingService) MarkAttendance(ctx context.Context, bookingID int64, userIDFromAuth int64, attendedStatus bool) (db.Booking, error) {
	booking, err := s.querier.GetBookingByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.WarnContext(ctx, "Booking not found for attendance marking", "booking_id", bookingID)
			return db.Booking{}, ErrBookingNotFound
		}
		s.logger.ErrorContext(ctx, "Failed to get booking by ID for attendance", "booking_id", bookingID, "error", err)
		return db.Booking{}, ErrInternalServer
	}

	// Authorization: Only the user who booked can mark attendance (or an admin, not implemented yet)
	if booking.UserID != userIDFromAuth {
		s.logger.WarnContext(ctx, "User forbidden to mark attendance for booking", "booking_id", bookingID, "booking_owner_id", booking.UserID, "auth_user_id", userIDFromAuth)
		return db.Booking{}, ErrForbiddenUpdate
	}

	updatedBooking, err := s.querier.UpdateBookingAttendance(ctx, db.UpdateBookingAttendanceParams{
		BookingID: bookingID,
		Attended:  attendedStatus,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to update booking attendance in DB", "booking_id", bookingID, "attended_status", attendedStatus, "error", err)
		return db.Booking{}, ErrInternalServer
	}

	s.logger.InfoContext(ctx, "Booking attendance marked successfully", "booking_id", updatedBooking.BookingID, "attended", updatedBooking.Attended)
	return updatedBooking, nil
} 