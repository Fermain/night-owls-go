package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"night-owls-go/internal/config"
	db "night-owls-go/internal/db/sqlc_generated"
)

// Error constants for recurring assignment operations
var (
	ErrRecurringAssignmentNotFound      = errors.New("recurring assignment not found")
	ErrRecurringAssignmentInternalError = errors.New("internal error in recurring assignment service")
)

// RecurringAssignmentService handles logic related to recurring shift assignments.
type RecurringAssignmentService struct {
	querier db.Querier
	logger  *slog.Logger
	config  *config.Config
}

// NewRecurringAssignmentService creates a new RecurringAssignmentService.
func NewRecurringAssignmentService(querier db.Querier, logger *slog.Logger, cfg *config.Config) *RecurringAssignmentService {
	return &RecurringAssignmentService{
		querier: querier,
		logger:  logger.With("service", "RecurringAssignmentService"),
		config:  cfg,
	}
}

// CreateRecurringAssignment creates a new recurring assignment.
func (s *RecurringAssignmentService) CreateRecurringAssignment(ctx context.Context, params db.CreateRecurringAssignmentParams) (db.RecurringAssignment, error) {
	// Validate the user exists
	_, err := s.querier.GetUserByID(ctx, params.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.WarnContext(ctx, "User not found for recurring assignment", "user_id", params.UserID)
			return db.RecurringAssignment{}, ErrRecurringAssignmentNotFound
		}
		s.logger.ErrorContext(ctx, "Failed to validate user for recurring assignment", "user_id", params.UserID, "error", err)
		return db.RecurringAssignment{}, ErrRecurringAssignmentInternalError
	}

	// Validate the schedule exists
	_, err = s.querier.GetScheduleByID(ctx, params.ScheduleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.WarnContext(ctx, "Schedule not found for recurring assignment", "schedule_id", params.ScheduleID)
			return db.RecurringAssignment{}, ErrRecurringAssignmentNotFound
		}
		s.logger.ErrorContext(ctx, "Failed to validate schedule for recurring assignment", "schedule_id", params.ScheduleID, "error", err)
		return db.RecurringAssignment{}, ErrRecurringAssignmentInternalError
	}

	assignment, err := s.querier.CreateRecurringAssignment(ctx, params)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to create recurring assignment", "params", params, "error", err)
		return db.RecurringAssignment{}, ErrRecurringAssignmentInternalError
	}

	s.logger.InfoContext(ctx, "Recurring assignment created", "assignment_id", assignment.RecurringAssignmentID, "user_id", assignment.UserID)
	return assignment, nil
}

// GetRecurringAssignmentByID retrieves a recurring assignment by its ID.
func (s *RecurringAssignmentService) GetRecurringAssignmentByID(ctx context.Context, assignmentID int64) (db.RecurringAssignment, error) {
	assignment, err := s.querier.GetRecurringAssignmentByID(ctx, assignmentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.WarnContext(ctx, "Recurring assignment not found", "assignment_id", assignmentID)
			return db.RecurringAssignment{}, ErrRecurringAssignmentNotFound
		}
		s.logger.ErrorContext(ctx, "Failed to get recurring assignment by ID", "assignment_id", assignmentID, "error", err)
		return db.RecurringAssignment{}, ErrRecurringAssignmentInternalError
	}
	return assignment, nil
}

// ListRecurringAssignments retrieves all active recurring assignments.
func (s *RecurringAssignmentService) ListRecurringAssignments(ctx context.Context) ([]db.RecurringAssignment, error) {
	assignments, err := s.querier.ListRecurringAssignments(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to list recurring assignments", "error", err)
		return nil, ErrRecurringAssignmentInternalError
	}
	return assignments, nil
}

// ListRecurringAssignmentsByUserID retrieves all active recurring assignments for a specific user.
func (s *RecurringAssignmentService) ListRecurringAssignmentsByUserID(ctx context.Context, userID int64) ([]db.RecurringAssignment, error) {
	assignments, err := s.querier.ListRecurringAssignmentsByUserID(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to list recurring assignments by user ID", "user_id", userID, "error", err)
		return nil, ErrRecurringAssignmentInternalError
	}
	return assignments, nil
}

// UpdateRecurringAssignment updates an existing recurring assignment.
func (s *RecurringAssignmentService) UpdateRecurringAssignment(ctx context.Context, params db.UpdateRecurringAssignmentParams) (db.RecurringAssignment, error) {
	// Validate the user exists
	_, err := s.querier.GetUserByID(ctx, params.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.WarnContext(ctx, "User not found for recurring assignment update", "user_id", params.UserID)
			return db.RecurringAssignment{}, ErrRecurringAssignmentNotFound
		}
		s.logger.ErrorContext(ctx, "Failed to validate user for recurring assignment update", "user_id", params.UserID, "error", err)
		return db.RecurringAssignment{}, ErrRecurringAssignmentInternalError
	}

	// Validate the schedule exists
	_, err = s.querier.GetScheduleByID(ctx, params.ScheduleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.WarnContext(ctx, "Schedule not found for recurring assignment update", "schedule_id", params.ScheduleID)
			return db.RecurringAssignment{}, ErrRecurringAssignmentNotFound
		}
		s.logger.ErrorContext(ctx, "Failed to validate schedule for recurring assignment update", "schedule_id", params.ScheduleID, "error", err)
		return db.RecurringAssignment{}, ErrRecurringAssignmentInternalError
	}

	assignment, err := s.querier.UpdateRecurringAssignment(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.WarnContext(ctx, "Recurring assignment not found for update", "assignment_id", params.RecurringAssignmentID)
			return db.RecurringAssignment{}, ErrRecurringAssignmentNotFound
		}
		s.logger.ErrorContext(ctx, "Failed to update recurring assignment", "params", params, "error", err)
		return db.RecurringAssignment{}, ErrRecurringAssignmentInternalError
	}

	s.logger.InfoContext(ctx, "Recurring assignment updated", "assignment_id", assignment.RecurringAssignmentID)
	return assignment, nil
}

// DeleteRecurringAssignment soft-deletes a recurring assignment by setting is_active to false.
func (s *RecurringAssignmentService) DeleteRecurringAssignment(ctx context.Context, assignmentID int64) error {
	err := s.querier.DeleteRecurringAssignment(ctx, assignmentID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to delete recurring assignment", "assignment_id", assignmentID, "error", err)
		return ErrRecurringAssignmentInternalError
	}

	s.logger.InfoContext(ctx, "Recurring assignment deleted", "assignment_id", assignmentID)
	return nil
}

// GetRecurringAssignmentsByPattern retrieves recurring assignments that match a specific pattern.
func (s *RecurringAssignmentService) GetRecurringAssignmentsByPattern(ctx context.Context, dayOfWeek int64, scheduleID int64, timeSlot string) ([]db.GetRecurringAssignmentsByPatternRow, error) {
	assignments, err := s.querier.GetRecurringAssignmentsByPattern(ctx, db.GetRecurringAssignmentsByPatternParams{
		DayOfWeek:  dayOfWeek,
		ScheduleID: scheduleID,
		TimeSlot:   timeSlot,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to get recurring assignments by pattern", "day_of_week", dayOfWeek, "schedule_id", scheduleID, "time_slot", timeSlot, "error", err)
		return nil, ErrRecurringAssignmentInternalError
	}
	return assignments, nil
}

// MaterializeUpcomingBookings creates individual bookings for recurring assignments
// within the specified time window. This should be called by a cron job.
func (s *RecurringAssignmentService) MaterializeUpcomingBookings(ctx context.Context, scheduleService *ScheduleService, fromTime time.Time, toTime time.Time) error {
	s.logger.InfoContext(ctx, "Starting to materialize upcoming bookings from recurring assignments", "from", fromTime, "to", toTime)

	// Get all active recurring assignments
	recurringAssignments, err := s.ListRecurringAssignments(ctx)
	if err != nil {
		return err
	}

	if len(recurringAssignments) == 0 {
		s.logger.InfoContext(ctx, "No recurring assignments found")
		return nil
	}

	// Get all shift slots in the time window
	limit := 1000 // Reasonable limit for the window
	slots, err := scheduleService.AdminGetAllShiftSlots(ctx, &fromTime, &toTime, &limit)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to get shift slots for materialization", "error", err)
		return err
	}

	materializedCount := 0
	skippedCount := 0

	// For each slot, check if it matches any recurring assignment
	for _, slot := range slots {
		// Skip if already booked
		if slot.IsBooked {
			continue
		}

		// Extract day of week and time slot from the shift
		dayOfWeek := int64(slot.StartTime.Weekday())
		startHour := slot.StartTime.Hour()
		startMin := slot.StartTime.Minute()
		endHour := slot.EndTime.Hour()
		endMin := slot.EndTime.Minute()
		timeSlot := fmt.Sprintf("%02d:%02d-%02d:%02d", startHour, startMin, endHour, endMin)

		// Find matching recurring assignments
		for _, assignment := range recurringAssignments {
			if assignment.DayOfWeek == dayOfWeek && 
			   assignment.ScheduleID == slot.ScheduleID && 
			   assignment.TimeSlot == timeSlot {
				
				// Check if booking already exists (in case of race conditions)
				_, err := s.querier.GetBookingByScheduleAndStartTime(ctx, db.GetBookingByScheduleAndStartTimeParams{
					ScheduleID: slot.ScheduleID,
					ShiftStart: slot.StartTime.UTC(),
				})
				
				if err == nil {
					// Booking already exists, skip
					skippedCount++
					continue
				} else if !errors.Is(err, sql.ErrNoRows) {
					// Unexpected error
					s.logger.ErrorContext(ctx, "Error checking existing booking", "error", err)
					continue
				}

				// Create the booking
				bookingParams := db.CreateBookingParams{
					UserID:     assignment.UserID,
					ScheduleID: assignment.ScheduleID,
					ShiftStart: slot.StartTime.UTC(),
					ShiftEnd:   slot.EndTime.UTC(),
				}

				// Add buddy if specified
				if assignment.BuddyName.Valid && assignment.BuddyName.String != "" {
					bookingParams.BuddyName = assignment.BuddyName
				}

				_, err = s.querier.CreateBooking(ctx, bookingParams)
				if err != nil {
					s.logger.ErrorContext(ctx, "Failed to create booking from recurring assignment", 
						"assignment_id", assignment.RecurringAssignmentID,
						"user_id", assignment.UserID,
						"schedule_id", assignment.ScheduleID,
						"shift_start", slot.StartTime,
						"error", err)
					continue
				}

				materializedCount++
				s.logger.InfoContext(ctx, "Created booking from recurring assignment",
					"assignment_id", assignment.RecurringAssignmentID,
					"user_id", assignment.UserID,
					"schedule_id", assignment.ScheduleID,
					"shift_start", slot.StartTime)
				
				// Only create one booking per slot (first matching assignment wins)
				break
			}
		}
	}

	s.logger.InfoContext(ctx, "Completed materializing bookings", 
		"materialized", materializedCount, 
		"skipped", skippedCount,
		"total_slots", len(slots))
	
	return nil
} 