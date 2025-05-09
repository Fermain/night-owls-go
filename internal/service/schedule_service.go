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
// across active schedules within the given time window (or a default window).
func (s *ScheduleService) GetUpcomingAvailableSlots(ctx context.Context, queryFrom *time.Time, queryTo *time.Time, limit *int) ([]AvailableShiftSlot, error) {
	now := time.Now()
	// Default query window: from now to 2 weeks from now if not specified
	defaultFrom := now
	defaultTo := now.AddDate(0, 0, 14) // 2 weeks

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
		return []AvailableShiftSlot{}, nil // Or return an error: errors.New("'from' date cannot be after 'to' date")
	}

	// 1. Fetch active schedules based on today's date (or queryFrom if it makes more sense for overall schedule validity)
	// We use current date for schedule active status for simplicity, assuming schedules are generally long-running.
	// The cron expression itself will be evaluated against actualFrom and actualTo.
	activeSchedules, err := s.querier.ListActiveSchedules(ctx, db.ListActiveSchedulesParams{
		Date:   sql.NullTime{Time: now, Valid: true},         // For checking schedule.start_date
		Date_2: sql.NullTime{Time: now, Valid: true},       // For checking schedule.end_date
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to list active schedules", "error", err)
		return nil, ErrInternalServer
	}

	if len(activeSchedules) == 0 {
		s.logger.InfoContext(ctx, "No active schedules found for the current period")
		return []AvailableShiftSlot{}, nil
	}

	var allPotentialSlots []AvailableShiftSlot

	for _, schedule := range activeSchedules {
		cronExpression, err := cronexpr.Parse(schedule.CronExpr)
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to parse cron expression for schedule", "schedule_id", schedule.ScheduleID, "cron_expr", schedule.CronExpr, "error", err)
			continue // Skip this schedule
		}

		// Determine the effective start and end for this schedule's occurrences
		scheduleActualStart := actualFrom
		if schedule.StartDate.Valid && schedule.StartDate.Time.After(scheduleActualStart) {
			scheduleActualStart = schedule.StartDate.Time
		}

		scheduleActualEnd := actualTo
		if schedule.EndDate.Valid && schedule.EndDate.Time.Before(scheduleActualEnd) {
			scheduleActualEnd = schedule.EndDate.Time
		}

		// Iterate through occurrences within the schedule's effective window
		nextTime := scheduleActualStart
		for {
			nextOccurrence := cronExpression.Next(nextTime)
			if nextOccurrence.IsZero() || nextOccurrence.After(scheduleActualEnd) {
				break // No more occurrences in the window for this schedule
			}
			
			// Ensure the occurrence is also within the overall query window [actualFrom, actualTo]
			// This is important if scheduleActualStart/End were narrowed by schedule.StartDate/EndDate
			if nextOccurrence.Before(actualFrom) || nextOccurrence.After(actualTo) {
				nextTime = nextOccurrence // Continue searching from this occurrence
				continue
			}

			shiftEndTime := nextOccurrence.Add(time.Duration(schedule.DurationMinutes) * time.Minute)
			
			potentialSlot := AvailableShiftSlot{
				ScheduleID:   schedule.ScheduleID,
				ScheduleName: schedule.Name,
				StartTime:    nextOccurrence,
				EndTime:      shiftEndTime,
				Timezone:     schedule.Timezone.String,
				IsBooked:     false, // Assume not booked until checked
			}
			allPotentialSlots = append(allPotentialSlots, potentialSlot)
			nextTime = nextOccurrence // Next search starts after the found occurrence
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
				// Not booked, add to available list
				availableSlots = append(availableSlots, slot)
			} else {
				s.logger.ErrorContext(ctx, "Error checking if slot is booked", "schedule_id", slot.ScheduleID, "start_time", slot.StartTime, "error", err)
				// Potentially skip this slot or handle error more gracefully
			}
		} else {
			// Slot is booked (no error means a row was found)
			// Do nothing, don't add to availableSlots
		}
	}

	// 3. Sort by start time
	sort.Slice(availableSlots, func(i, j int) bool {
		return availableSlots[i].StartTime.Before(availableSlots[j].StartTime)
	})

	// 4. Apply limit if specified
	if limit != nil && len(availableSlots) > *limit {
		availableSlots = availableSlots[:*limit]
	}

	return availableSlots, nil
} 