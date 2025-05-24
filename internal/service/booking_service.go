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

	// Ensure times are stored in UTC for consistency with availability lookup
	utcStartTime := startTime.UTC()
	utcEndTime := shiftEndTime.UTC()

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
		ShiftStart:  utcStartTime,
		ShiftEnd:    utcEndTime,
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

// AdminAssignUserToShift handles the logic for an admin assigning a user to a specific shift slot.
func (s *BookingService) AdminAssignUserToShift(ctx context.Context, targetUserID int64, scheduleID int64, shiftStartTime time.Time) (db.Booking, error) {
	// 1. Validate target user
	_, err := s.querier.GetUserByID(ctx, targetUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.WarnContext(ctx, "Target user not found for admin assignment", "target_user_id", targetUserID)
			return db.Booking{}, ErrUserNotFound // Assuming ErrUserNotFound is defined, or use a generic error
		}
		s.logger.ErrorContext(ctx, "Failed to get target user by ID for admin assignment", "target_user_id", targetUserID, "error", err)
		return db.Booking{}, ErrInternalServer
	}

	// 2. Validate schedule and start time
	schedule, err := s.querier.GetScheduleByID(ctx, scheduleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.WarnContext(ctx, "Schedule not found for admin assignment", "schedule_id", scheduleID)
			return db.Booking{}, ErrScheduleNotFound
		}
		s.logger.ErrorContext(ctx, "Failed to get schedule by ID for admin assignment", "schedule_id", scheduleID, "error", err)
		return db.Booking{}, ErrInternalServer
	}

	// Ensure shiftStartTime is UTC before any comparisons or calculations that assume UTC
	// The shiftStartTime from the request should ideally be parsed as UTC or converted if it has timezone info.
	// For this service method, we'll assume it's provided as UTC, consistent with how bookings are stored.
	utcShiftStartTime := shiftStartTime.UTC()


	// Check if shiftStartTime is within the schedule's overall active period (start_date and end_date)
	// Dates in DB are YYYY-MM-DD, effectively UTC. startTime is also UTC.
	if (schedule.StartDate.Valid && utcShiftStartTime.Before(schedule.StartDate.Time)) ||
		(schedule.EndDate.Valid && utcShiftStartTime.After(schedule.EndDate.Time.AddDate(0,0,1).Add(-time.Nanosecond))) { // end_date is inclusive
		s.logger.WarnContext(ctx, "Admin assignment attempt for shift time outside schedule active dates",
			"schedule_id", scheduleID, "start_time", utcShiftStartTime,
			"schedule_start_date", schedule.StartDate, "schedule_end_date", schedule.EndDate)
		return db.Booking{}, ErrShiftTimeInvalid
	}
	
	// Handle timezone for cron expression validation
	// Load the schedule's timezone, or default to UTC if not specified or invalid
	loc := time.UTC
	if schedule.Timezone.Valid && schedule.Timezone.String != "" {
		loadedLoc, locErr := time.LoadLocation(schedule.Timezone.String)
		if locErr == nil {
			loc = loadedLoc
		} else {
			s.logger.WarnContext(ctx, "Failed to load timezone for schedule during admin assignment, using UTC",
				"schedule_id", schedule.ScheduleID, "timezone_str", schedule.Timezone.String, "error", locErr)
		}
	}
	// Convert the UTC shiftStartTime to the schedule's local time for cron validation
	localShiftStartTimeForCron := utcShiftStartTime.In(loc)


	cronExpression, err := cronexpr.Parse(schedule.CronExpr)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to parse cron expression for schedule during admin assignment", "schedule_id", scheduleID, "cron_expr", schedule.CronExpr, "error", err)
		return db.Booking{}, ErrInternalServer
	}

	// Validate the localShiftStartTimeForCron against the schedule's cron expression.
	nextOccurrenceFromAlmostStartTime := cronExpression.Next(localShiftStartTimeForCron.Add(-1 * time.Second))
	if nextOccurrenceFromAlmostStartTime.IsZero() || !nextOccurrenceFromAlmostStartTime.Equal(localShiftStartTimeForCron) {
		s.logger.WarnContext(ctx, "Requested start_time does not match a cron expression occurrence in schedule's timezone",
			"schedule_id", scheduleID, "start_time_utc", utcShiftStartTime, "start_time_local", localShiftStartTimeForCron, "cron_expr", schedule.CronExpr, "calculated_next_local", nextOccurrenceFromAlmostStartTime)
		return db.Booking{}, ErrShiftTimeInvalid
	}

	// 3. Check for booking conflicts (using UTC start time)
	_, err = s.querier.GetBookingByScheduleAndStartTime(ctx, db.GetBookingByScheduleAndStartTimeParams{
		ScheduleID: scheduleID,
		ShiftStart: utcShiftStartTime,
	})
	if err == nil {
		// A booking was found, so it's a conflict
		s.logger.WarnContext(ctx, "Admin assignment conflict: Slot already booked", "schedule_id", scheduleID, "start_time", utcShiftStartTime)
		return db.Booking{}, ErrBookingConflict
	}
	if !errors.Is(err, sql.ErrNoRows) {
		// An unexpected error occurred while checking for conflicts
		s.logger.ErrorContext(ctx, "Failed to check for booking conflict during admin assignment", "schedule_id", scheduleID, "start_time", utcShiftStartTime, "error", err)
		return db.Booking{}, ErrInternalServer
	}
	// If err is sql.ErrNoRows, the slot is available, proceed.

	// 4. Calculate shift_end (using UTC start time)
	shiftEndTime := utcShiftStartTime.Add(time.Duration(schedule.DurationMinutes) * time.Minute)

	// 5. Create booking
	bookingParams := db.CreateBookingParams{
		UserID:      targetUserID,
		ScheduleID:  scheduleID,
		ShiftStart:  utcShiftStartTime, // Store in UTC
		ShiftEnd:    shiftEndTime,   // Store in UTC
		BuddyUserID: sql.NullInt64{Valid: false}, // No buddy for admin assignment by default
		BuddyName:   sql.NullString{Valid: false},// No buddy for admin assignment by default
	}

	createdBooking, err := s.querier.CreateBooking(ctx, bookingParams)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to create booking in DB during admin assignment", "params", bookingParams, "error", err)
		if isUniqueConstraintError(err) {
			return db.Booking{}, ErrBookingConflict // Should have been caught above, but as a safeguard
		}
		return db.Booking{}, ErrInternalServer
	}
	s.logger.InfoContext(ctx, "Booking created successfully by admin", "booking_id", createdBooking.BookingID, "assigned_user_id", targetUserID, "schedule_id", scheduleID)

	// 6. (Optional) Queue confirmation message to outbox for the assigned user
	outboxPayload := fmt.Sprintf(`{"booking_id": %d, "user_id": %d, "shift_start": "%s", "assigned_by": "admin"}`,
		createdBooking.BookingID, createdBooking.UserID, createdBooking.ShiftStart.Format(time.RFC3339))
	_, err = s.querier.CreateOutboxItem(ctx, db.CreateOutboxItemParams{
		MessageType: "ADMIN_SHIFT_ASSIGNMENT",
		Recipient:   fmt.Sprintf("%d", createdBooking.UserID), // Or user's phone if preferred for notification
		Payload:     sql.NullString{String: outboxPayload, Valid: true},
		UserID:      sql.NullInt64{Int64: targetUserID, Valid: true},
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to create outbox item for admin assignment notification", "booking_id", createdBooking.BookingID, "error", err)
		// Non-fatal for booking creation itself, but log it.
	}

	return createdBooking, nil
}

// GetUserBookings retrieves all bookings for a specific user.
func (s *BookingService) GetUserBookings(ctx context.Context, userID int64) ([]db.ListBookingsByUserIDWithScheduleRow, error) {
	// Validate that user exists
	_, err := s.querier.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.WarnContext(ctx, "User not found for booking retrieval", "user_id", userID)
			return nil, ErrUserNotFound
		}
		s.logger.ErrorContext(ctx, "Failed to get user by ID for booking retrieval", "user_id", userID, "error", err)
		return nil, ErrInternalServer
	}

	// Get all bookings for the user with schedule names
	bookings, err := s.querier.ListBookingsByUserIDWithSchedule(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to get user bookings with schedule", "user_id", userID, "error", err)
		return nil, ErrInternalServer
	}

	return bookings, nil
} 