package service

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"sort"
	"time"

	"night-owls-go/internal/config"
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
	config  *config.Config
}

// NewScheduleService creates a new ScheduleService.
func NewScheduleService(querier db.Querier, logger *slog.Logger, cfg *config.Config) *ScheduleService {
	return &ScheduleService{
		querier: querier,
		logger:  logger.With("service", "ScheduleService"),
		config:  cfg,
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

// AdminAvailableShiftSlot represents a shift slot with booking details for admin view.
type AdminAvailableShiftSlot struct {
	ScheduleID   int64     `json:"schedule_id"`
	ScheduleName string    `json:"schedule_name"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Timezone     string    `json:"timezone,omitempty"`
	IsBooked     bool      `json:"is_booked"`
	BookingID    *int64    `json:"booking_id,omitempty"`
	UserName     *string   `json:"user_name,omitempty"`
	UserPhone    *string   `json:"user_phone,omitempty"`
}

// GetUpcomingAvailableSlots finds all available (not booked) shift slots
// across schedules that are active within the given time window.
func (s *ScheduleService) GetUpcomingAvailableSlots(ctx context.Context, queryFrom *time.Time, queryTo *time.Time, limit *int) ([]AvailableShiftSlot, error) {
	now := time.Now().UTC() // Use UTC for baseline "now"
	defaultFrom := now
	defaultTo := now.AddDate(0, 0, 14)

	actualFrom := defaultFrom
	if queryFrom != nil {
		actualFrom = (*queryFrom).UTC() // Ensure query parameters are treated as UTC
	}
	actualTo := defaultTo
	if queryTo != nil {
		actualTo = (*queryTo).UTC()
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
		loc := time.UTC // Default to UTC
		if schedule.Timezone.Valid && schedule.Timezone.String != "" {
			loadedLoc, errLoadLoc := time.LoadLocation(schedule.Timezone.String)
			if errLoadLoc != nil {
				s.logger.WarnContext(ctx, "Failed to load timezone for schedule, defaulting to UTC",
					"schedule_id", schedule.ScheduleID, "timezone_str", schedule.Timezone.String, "error", errLoadLoc)
			} else {
				loc = loadedLoc
			}
		}

		queryFromInLoc := actualFrom.In(loc)
		queryToInLoc := actualTo.In(loc)

		scheduleActiveStartInLoc := time.Time{}
		if schedule.StartDate.Valid {
			y, m, d := schedule.StartDate.Time.Date()
			scheduleActiveStartInLoc = time.Date(y, m, d, 0, 0, 0, 0, loc)
		}
		scheduleActiveEndInLoc := time.Time{}
		if schedule.EndDate.Valid {
			y, m, d := schedule.EndDate.Time.Date()
			scheduleActiveEndInLoc = time.Date(y, m, d, 23, 59, 59, 999999999, loc)
		}

		iterationStartInLoc := queryFromInLoc
		if !scheduleActiveStartInLoc.IsZero() && scheduleActiveStartInLoc.After(iterationStartInLoc) {
			iterationStartInLoc = scheduleActiveStartInLoc
		}

		iterationEndInLoc := queryToInLoc
		if !scheduleActiveEndInLoc.IsZero() && scheduleActiveEndInLoc.Before(iterationEndInLoc) {
			iterationEndInLoc = scheduleActiveEndInLoc
		}

		if iterationStartInLoc.After(iterationEndInLoc) {
			continue
		}

		cronExpression, err := cronexpr.Parse(schedule.CronExpr)
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to parse cron expression", "schedule_id", schedule.ScheduleID, "error", err)
			continue
		}

		nextTime := iterationStartInLoc
		firstPossibleOccurrence := cronExpression.Next(iterationStartInLoc.Add(-time.Second))
		if firstPossibleOccurrence.Equal(iterationStartInLoc) {
			if !firstPossibleOccurrence.After(iterationEndInLoc) {
				shiftEndTime := firstPossibleOccurrence.Add(time.Duration(schedule.DurationMinutes) * time.Minute)
				potentialSlot := AvailableShiftSlot{
					ScheduleID:   schedule.ScheduleID,
					ScheduleName: schedule.Name,
					StartTime:    firstPossibleOccurrence,
					EndTime:      shiftEndTime,
					Timezone:     loc.String(),
					IsBooked:     false,
				}
				allPotentialSlots = append(allPotentialSlots, potentialSlot)
			}
			nextTime = firstPossibleOccurrence
		}

		for {
			nextOccurrence := cronExpression.Next(nextTime)
			if nextOccurrence.IsZero() || nextOccurrence.After(iterationEndInLoc) {
				break
			}

			shiftEndTime := nextOccurrence.Add(time.Duration(schedule.DurationMinutes) * time.Minute)
			potentialSlot := AvailableShiftSlot{
				ScheduleID:   schedule.ScheduleID,
				ScheduleName: schedule.Name,
				StartTime:    nextOccurrence,
				EndTime:      shiftEndTime,
				Timezone:     loc.String(),
				IsBooked:     false,
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

// AdminGetAllShiftSlots finds all shift slots (booked or not)
// across all schedules that are active within the given time window,
// and includes booking details if a slot is booked.
func (s *ScheduleService) AdminGetAllShiftSlots(ctx context.Context, queryFrom *time.Time, queryTo *time.Time, limit *int) ([]AdminAvailableShiftSlot, error) {
	now := time.Now().UTC() // Use UTC for baseline "now"
	// Default window for admin view might be shorter or configurable, e.g., 7 days
	defaultFrom := now
	defaultTo := now.AddDate(0, 0, 7)

	actualFrom := defaultFrom
	if queryFrom != nil {
		actualFrom = (*queryFrom).UTC() // Ensure query parameters are treated as UTC if no TZ info
	}
	actualTo := defaultTo
	if queryTo != nil {
		actualTo = (*queryTo).UTC()
	}

	if actualFrom.After(actualTo) {
		s.logger.WarnContext(ctx, "Query 'from' date is after 'to' date for admin slots", "from", actualFrom, "to", actualTo)
		return []AdminAvailableShiftSlot{}, nil
	}

	allSchedules, err := s.querier.ListAllSchedules(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to list all schedules for admin slots", "error", err)
		return nil, ErrInternalServer
	}

	if len(allSchedules) == 0 {
		s.logger.InfoContext(ctx, "No schedules defined in the system for admin slots")
		return []AdminAvailableShiftSlot{}, nil
	}

	var allSlots []AdminAvailableShiftSlot

	for _, schedule := range allSchedules {
		loc := time.UTC // Default to UTC
		if schedule.Timezone.Valid && schedule.Timezone.String != "" {
			loadedLoc, errLoadLoc := time.LoadLocation(schedule.Timezone.String)
			if errLoadLoc != nil {
				s.logger.WarnContext(ctx, "Failed to load timezone for schedule, defaulting to UTC",
					"schedule_id", schedule.ScheduleID, "timezone_str", schedule.Timezone.String, "error", errLoadLoc)
				// loc remains time.UTC
			} else {
				loc = loadedLoc
			}
		}

		// Convert query window to schedule's location
		queryFromInLoc := actualFrom.In(loc)
		queryToInLoc := actualTo.In(loc)

		// Determine schedule's own active start/end in its location
		scheduleActiveStartInLoc := time.Time{} // Zero time if not set
		if schedule.StartDate.Valid {
			// Assuming StartDate from DB is date-only, interpret it in schedule's loc as start of day
			y, m, d := schedule.StartDate.Time.Date()
			scheduleActiveStartInLoc = time.Date(y, m, d, 0, 0, 0, 0, loc)
		}
		scheduleActiveEndInLoc := time.Time{} // Zero time, effectively no upper bound if not set
		if schedule.EndDate.Valid {
			// Assuming EndDate from DB is date-only, interpret it in schedule's loc as end of day
			y, m, d := schedule.EndDate.Time.Date()
			scheduleActiveEndInLoc = time.Date(y, m, d, 23, 59, 59, 999999999, loc)
		}

		// Determine the effective iteration window for this schedule in its location
		iterationStartInLoc := queryFromInLoc
		if !scheduleActiveStartInLoc.IsZero() && scheduleActiveStartInLoc.After(iterationStartInLoc) {
			iterationStartInLoc = scheduleActiveStartInLoc
		}

		iterationEndInLoc := queryToInLoc
		if !scheduleActiveEndInLoc.IsZero() && scheduleActiveEndInLoc.Before(iterationEndInLoc) {
			iterationEndInLoc = scheduleActiveEndInLoc
		}

		// Check if schedule itself is active/relevant within the query window
		// This logic effectively replaces the previous scheduleStartsBeforeOrAtQueryEnd/scheduleEndsAfterOrAtQueryStart
		if iterationStartInLoc.After(iterationEndInLoc) {
			continue // No overlap between query window and schedule's active period
		}

		cronExpression, err := cronexpr.Parse(schedule.CronExpr)
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to parse cron expression for admin slots", "schedule_id", schedule.ScheduleID, "error", err)
			continue
		}
		
		// nextTime must be in the schedule's location for cronexpr.Next() to work as intended
		nextTime := iterationStartInLoc
		
		// Handle iterationStartInLoc potentially being a valid cron time.
		// cronexpr.Next() finds the time *after* the given time. To include iterationStartInLoc if it's a hit:
		// Check if the cron's next time from (iterationStartInLoc - 1 sec) is iterationStartInLoc itself.
		firstPossibleOccurrence := cronExpression.Next(iterationStartInLoc.Add(-time.Second))
		if firstPossibleOccurrence.Equal(iterationStartInLoc) {
			if !firstPossibleOccurrence.After(iterationEndInLoc) {
				shiftEndTime := firstPossibleOccurrence.Add(time.Duration(schedule.DurationMinutes) * time.Minute)
				slot := AdminAvailableShiftSlot{
					ScheduleID:   schedule.ScheduleID,
					ScheduleName: schedule.Name,
					StartTime:    firstPossibleOccurrence, // This is in schedule's loc
					EndTime:      shiftEndTime,          // This is also in schedule's loc
					Timezone:     loc.String(),          // Store the location string used
					IsBooked:     false, 
				}
				allSlots = append(allSlots, slot)
			}
			// Set nextTime to firstPossibleOccurrence to ensure the loop starts correctly
			// if iterationStartInLoc was indeed a hit.
			nextTime = firstPossibleOccurrence
		}

		for {
			// nextTime is already in loc. cronExpression.Next will return time in loc.
			nextOccurrence := cronExpression.Next(nextTime)
			if nextOccurrence.IsZero() || nextOccurrence.After(iterationEndInLoc) {
				break
			}

			shiftEndTime := nextOccurrence.Add(time.Duration(schedule.DurationMinutes) * time.Minute)
			slot := AdminAvailableShiftSlot{
				ScheduleID:   schedule.ScheduleID,
				ScheduleName: schedule.Name,
				StartTime:    nextOccurrence, // This is in schedule's loc
				EndTime:      shiftEndTime,   // This is also in schedule's loc
				Timezone:     loc.String(),   // Store the location string used
				IsBooked:     false, 
			}
			allSlots = append(allSlots, slot)
			nextTime = nextOccurrence
		}
	}

	// Populate booking details
	var populatedSlots []AdminAvailableShiftSlot
	for _, slot := range allSlots {
		populatedSlot := slot // Make a copy to modify
		booking, err := s.querier.GetBookingByScheduleAndStartTime(ctx, db.GetBookingByScheduleAndStartTimeParams{
			ScheduleID: slot.ScheduleID,
			ShiftStart: slot.StartTime,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				populatedSlot.IsBooked = false
			} else {
				s.logger.ErrorContext(ctx, "Error checking booking for admin slot", "schedule_id", slot.ScheduleID, "start_time", slot.StartTime, "error", err)
				// Decide if we should skip this slot or return it as not booked
				// For now, assume not booked on error, but log it
				populatedSlot.IsBooked = false
			}
		} else {
			populatedSlot.IsBooked = true
			populatedSlot.BookingID = &booking.BookingID
			// Fetch user details
			user, userErr := s.querier.GetUserByID(ctx, booking.UserID)
			if userErr != nil {
				if !errors.Is(userErr, sql.ErrNoRows) {
					s.logger.ErrorContext(ctx, "Error fetching user for booked admin slot", "user_id", booking.UserID, "booking_id", booking.BookingID, "error", userErr)
				}
				// User details might be missing or an error occurred, leave UserName/UserPhone as nil
			} else {
				if user.Name.Valid {
					populatedSlot.UserName = &user.Name.String
				}
				// Assuming User struct has a Phone field of type string
				populatedSlot.UserPhone = &user.Phone
			}
		}
		populatedSlots = append(populatedSlots, populatedSlot)
	}

	sort.Slice(populatedSlots, func(i, j int) bool {
		return populatedSlots[i].StartTime.Before(populatedSlots[j].StartTime)
	})

	if limit != nil && len(populatedSlots) > *limit {
		populatedSlots = populatedSlots[:*limit]
	}

	return populatedSlots, nil
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
	params.DurationMinutes = int64(s.config.DefaultShiftDuration.Minutes())

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
	params.DurationMinutes = int64(s.config.DefaultShiftDuration.Minutes())

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

// AdminBulkDeleteSchedules deletes multiple schedules by their IDs.
func (s *ScheduleService) AdminBulkDeleteSchedules(ctx context.Context, scheduleIDs []int64) error {
	return s.querier.AdminBulkDeleteSchedules(ctx, scheduleIDs)
} 