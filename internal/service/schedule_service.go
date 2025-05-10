package service

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"sort"
	"time"

	db "night-owls-go/internal/db/sqlc_generated"

	"github.com/gorhill/cronexpr"
)

// Service specific errors
var (
	ErrNotFound       = errors.New("requested resource not found")
	// ErrInternalServer is assumed to be defined globally or in another service package.
	// Add other common service errors here if needed, e.g., ErrInvalidInput
)

// ScheduleService handles logic related to schedules and shift availability.
type ScheduleService struct {
	querier db.Querier
	logger  *slog.Logger
}

// NewScheduleService creates a new ScheduleService.
func NewScheduleService(querier db.Querier, logger *slog.Logger) *ScheduleService {
	return &ScheduleService{
		querier: querier,
		logger:  logger.With("service", "ScheduleService"),
	}
}

// AvailableShiftSlot represents a shift slot that can be booked.
// It combines information from a schedule and a specific occurrence.
type AvailableShiftSlot struct {
	ScheduleID    int64     `json:"schedule_id"`
	ScheduleName  string    `json:"schedule_name"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Timezone      string    `json:"timezone,omitempty"`
	IsBooked      bool      `json:"is_booked"` // Should always be false when returned by GetUpcomingAvailableSlots
}

// GetUpcomingAvailableSlots finds all available (not booked) shift slots
// across schedules that are active within the given time window.
func (s *ScheduleService) GetUpcomingAvailableSlots(ctx context.Context, queryFrom *time.Time, queryTo *time.Time, limit *int) ([]AvailableShiftSlot, error) {
	now := time.Now()
	defaultFrom := now
	defaultTo := now.AddDate(0, 0, 14) 

	actualFrom := defaultFrom
	if queryFrom != nil {
		actualFrom = *queryFrom
	}
	actualTo := defaultTo
	if queryTo != nil {
		actualTo = *queryTo
	}

	if actualFrom.After(actualTo) {
		s.logger.WarnContext(ctx, "Query 'from' date is after 'to' date", "from", actualFrom, "to", actualTo)
		return []AvailableShiftSlot{}, nil 
	}

	// 1. Fetch all schedules
	allSchedules, err := s.querier.ListAllSchedules(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to list all schedules", "error", err)
		return nil, ErrInternalServer
	}

	if len(allSchedules) == 0 {
		s.logger.InfoContext(ctx, "No schedules defined in the system")
		return []AvailableShiftSlot{}, nil
	}

	var allPotentialSlots []AvailableShiftSlot

	for _, schedule := range allSchedules {
		// Check if schedule itself is active/relevant within the query window [actualFrom, actualTo]
		scheduleStartsBeforeOrAtQueryEnd := !schedule.StartDate.Valid || !schedule.StartDate.Time.After(actualTo)
		scheduleEndsAfterOrAtQueryStart := !schedule.EndDate.Valid || !schedule.EndDate.Time.Before(actualFrom)

		if !(scheduleStartsBeforeOrAtQueryEnd && scheduleEndsAfterOrAtQueryStart) {
			// This schedule does not overlap with the query window at all
			continue
		}

		cronExpression, err := cronexpr.Parse(schedule.CronExpr)
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to parse cron expression", "schedule_id", schedule.ScheduleID, "error", err)
			continue 
		}

		// Determine the effective start and end for iterating occurrences for this schedule
		// It must be within the query window AND within the schedule's own active dates.
		iterationStart := actualFrom
		if schedule.StartDate.Valid && schedule.StartDate.Time.After(iterationStart) {
			iterationStart = schedule.StartDate.Time
		}

		iterationEnd := actualTo
		if schedule.EndDate.Valid && schedule.EndDate.Time.Before(iterationEnd) {
			iterationEnd = schedule.EndDate.Time
		}

		// If iteration window is invalid (e.g. start is after end), skip
        if iterationStart.After(iterationEnd) {
            continue
        }

		nextTime := iterationStart // Start generating from the beginning of the intersection window
        // cronexpr.Next gives time strictly *after* nextTime. To include iterationStart if it's a valid cron time:
        if cronexpr.MustParse(schedule.CronExpr).Next(iterationStart.Add(-time.Second)).Equal(iterationStart) {
            // iterationStart is a valid cron time, process it first
            if !iterationStart.After(iterationEnd) { // Check if it's within iterationEnd
                shiftEndTime := iterationStart.Add(time.Duration(schedule.DurationMinutes) * time.Minute)
                potentialSlot := AvailableShiftSlot{
                    ScheduleID:   schedule.ScheduleID, ScheduleName: schedule.Name, 
                    StartTime: iterationStart, EndTime: shiftEndTime, 
                    Timezone: schedule.Timezone.String, IsBooked: false,
                }
                allPotentialSlots = append(allPotentialSlots, potentialSlot)
            }
        }

		for {
			nextOccurrence := cronExpression.Next(nextTime)
			if nextOccurrence.IsZero() || nextOccurrence.After(iterationEnd) {
				break 
			}
			
			shiftEndTime := nextOccurrence.Add(time.Duration(schedule.DurationMinutes) * time.Minute)
			potentialSlot := AvailableShiftSlot{
				ScheduleID:   schedule.ScheduleID, ScheduleName: schedule.Name, 
				StartTime: nextOccurrence, EndTime:      shiftEndTime, 
				Timezone: schedule.Timezone.String, IsBooked:     false,
			}
			allPotentialSlots = append(allPotentialSlots, potentialSlot)
			nextTime = nextOccurrence 
		}
	}

	// 2. Filter out booked slots
	// This could be optimized by fetching all relevant bookings in one go, but for simplicity:
	var availableSlots []AvailableShiftSlot
	for _, slot := range allPotentialSlots {
		_, err := s.querier.GetBookingByScheduleAndStartTime(ctx, db.GetBookingByScheduleAndStartTimeParams{
			ScheduleID:  slot.ScheduleID,
			ShiftStart: slot.StartTime,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				availableSlots = append(availableSlots, slot)
			} else {
				s.logger.ErrorContext(ctx, "Error checking if slot is booked", "schedule_id", slot.ScheduleID, "error", err)
			}
		} 
	}
	sort.Slice(availableSlots, func(i, j int) bool {
		return availableSlots[i].StartTime.Before(availableSlots[j].StartTime)
	})
	if limit != nil && len(availableSlots) > *limit {
		availableSlots = availableSlots[:*limit]
	}

	return availableSlots, nil
}

// Add the ListAllSchedules method to retrieve all schedules
func (s *ScheduleService) ListAllSchedules(ctx context.Context) ([]db.Schedule, error) {
	schedules, err := s.querier.ListAllSchedules(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to list all schedules", "error", err)
		return nil, ErrInternalServer
	}
	return schedules, nil
}

// AdminCreateSchedule creates a new schedule (admin operation).
func (s *ScheduleService) AdminCreateSchedule(ctx context.Context, params db.CreateScheduleParams) (db.Schedule, error) {
	schedule, err := s.querier.CreateSchedule(ctx, params)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to create schedule (admin)", "params", params, "error", err)
		return db.Schedule{}, ErrInternalServer
	}
	s.logger.InfoContext(ctx, "Schedule created (admin)", "schedule_id", schedule.ScheduleID, "name", schedule.Name)
	return schedule, nil
}

// AdminGetScheduleByID retrieves a specific schedule by its ID (admin operation).
func (s *ScheduleService) AdminGetScheduleByID(ctx context.Context, scheduleID int64) (db.Schedule, error) {
	schedule, err := s.querier.GetScheduleByID(ctx, scheduleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.WarnContext(ctx, "Schedule not found (admin)", "schedule_id", scheduleID, "error", err)
			return db.Schedule{}, ErrNotFound
		}
		s.logger.ErrorContext(ctx, "Failed to get schedule by ID (admin)", "schedule_id", scheduleID, "error", err)
		return db.Schedule{}, ErrInternalServer
	}
	return schedule, nil
}

// AdminUpdateSchedule updates an existing schedule (admin operation).
func (s *ScheduleService) AdminUpdateSchedule(ctx context.Context, params db.UpdateScheduleParams) (db.Schedule, error) {
	schedule, err := s.querier.UpdateSchedule(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { 
			s.logger.WarnContext(ctx, "Schedule not found for update (admin)", "schedule_id", params.ScheduleID, "error", err)
			return db.Schedule{}, ErrNotFound
		}
		s.logger.ErrorContext(ctx, "Failed to update schedule (admin)", "params", params, "error", err)
		return db.Schedule{}, ErrInternalServer
	}
	s.logger.InfoContext(ctx, "Schedule updated (admin)", "schedule_id", schedule.ScheduleID, "name", schedule.Name)
	return schedule, nil
}

// AdminDeleteSchedule deletes a schedule by its ID (admin operation).
func (s *ScheduleService) AdminDeleteSchedule(ctx context.Context, scheduleID int64) error {
	err := s.querier.DeleteSchedule(ctx, scheduleID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to delete schedule (admin)", "schedule_id", scheduleID, "error", err)
		return ErrInternalServer
	}
	s.logger.InfoContext(ctx, "Schedule deleted (admin)", "schedule_id", scheduleID)
	return nil
} 