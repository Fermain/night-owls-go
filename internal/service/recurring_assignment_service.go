package service

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"night-owls-go/internal/config"
	db "night-owls-go/internal/db/sqlc_generated"
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
			return db.RecurringAssignment{}, ErrNotFound
		}
		s.logger.ErrorContext(ctx, "Failed to validate user for recurring assignment", "user_id", params.UserID, "error", err)
		return db.RecurringAssignment{}, ErrInternalServer
	}

	// Validate the schedule exists
	_, err = s.querier.GetScheduleByID(ctx, params.ScheduleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.WarnContext(ctx, "Schedule not found for recurring assignment", "schedule_id", params.ScheduleID)
			return db.RecurringAssignment{}, ErrNotFound
		}
		s.logger.ErrorContext(ctx, "Failed to validate schedule for recurring assignment", "schedule_id", params.ScheduleID, "error", err)
		return db.RecurringAssignment{}, ErrInternalServer
	}

	assignment, err := s.querier.CreateRecurringAssignment(ctx, params)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to create recurring assignment", "params", params, "error", err)
		return db.RecurringAssignment{}, ErrInternalServer
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
			return db.RecurringAssignment{}, ErrNotFound
		}
		s.logger.ErrorContext(ctx, "Failed to get recurring assignment by ID", "assignment_id", assignmentID, "error", err)
		return db.RecurringAssignment{}, ErrInternalServer
	}
	return assignment, nil
}

// ListRecurringAssignments retrieves all active recurring assignments.
func (s *RecurringAssignmentService) ListRecurringAssignments(ctx context.Context) ([]db.RecurringAssignment, error) {
	assignments, err := s.querier.ListRecurringAssignments(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to list recurring assignments", "error", err)
		return nil, ErrInternalServer
	}
	return assignments, nil
}

// ListRecurringAssignmentsByUserID retrieves all active recurring assignments for a specific user.
func (s *RecurringAssignmentService) ListRecurringAssignmentsByUserID(ctx context.Context, userID int64) ([]db.RecurringAssignment, error) {
	assignments, err := s.querier.ListRecurringAssignmentsByUserID(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to list recurring assignments by user ID", "user_id", userID, "error", err)
		return nil, ErrInternalServer
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
			return db.RecurringAssignment{}, ErrNotFound
		}
		s.logger.ErrorContext(ctx, "Failed to validate user for recurring assignment update", "user_id", params.UserID, "error", err)
		return db.RecurringAssignment{}, ErrInternalServer
	}

	// Validate the schedule exists
	_, err = s.querier.GetScheduleByID(ctx, params.ScheduleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.WarnContext(ctx, "Schedule not found for recurring assignment update", "schedule_id", params.ScheduleID)
			return db.RecurringAssignment{}, ErrNotFound
		}
		s.logger.ErrorContext(ctx, "Failed to validate schedule for recurring assignment update", "schedule_id", params.ScheduleID, "error", err)
		return db.RecurringAssignment{}, ErrInternalServer
	}

	assignment, err := s.querier.UpdateRecurringAssignment(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.WarnContext(ctx, "Recurring assignment not found for update", "assignment_id", params.RecurringAssignmentID)
			return db.RecurringAssignment{}, ErrNotFound
		}
		s.logger.ErrorContext(ctx, "Failed to update recurring assignment", "params", params, "error", err)
		return db.RecurringAssignment{}, ErrInternalServer
	}

	s.logger.InfoContext(ctx, "Recurring assignment updated", "assignment_id", assignment.RecurringAssignmentID)
	return assignment, nil
}

// DeleteRecurringAssignment soft-deletes a recurring assignment by setting is_active to false.
func (s *RecurringAssignmentService) DeleteRecurringAssignment(ctx context.Context, assignmentID int64) error {
	err := s.querier.DeleteRecurringAssignment(ctx, assignmentID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to delete recurring assignment", "assignment_id", assignmentID, "error", err)
		return ErrInternalServer
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
		return nil, ErrInternalServer
	}
	return assignments, nil
} 