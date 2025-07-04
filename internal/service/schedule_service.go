package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"night-owls-go/internal/config"
	db "night-owls-go/internal/db/sqlc_generated"

	"github.com/robfig/cron/v3"
)

// Service specific errors
var (
	ErrNotFound     = errors.New("requested resource not found")
	ErrInvalidInput = errors.New("invalid input")
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
// Enhanced to include assignment details for community roster experience.
type AvailableShiftSlot struct {
	ScheduleID   int64     `json:"schedule_id"`
	ScheduleName string    `json:"schedule_name"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Timezone     string    `json:"timezone,omitempty"`
	IsBooked     bool      `json:"is_booked"`
	BookingID    *int64    `json:"booking_id,omitempty"`
	UserName     *string   `json:"user_name,omitempty"`
	UserPhone    *string   `json:"user_phone,omitempty"`
	BuddyName    *string   `json:"buddy_name,omitempty"`
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
	BuddyName    *string   `json:"buddy_name,omitempty"`
}

// calculateScheduleBoundaryTimesInLocation determines the effective start and end times
// of a schedule in its specific timezone, and returns the loaded location.
// It uses the YYYY-MM-DD from the schedule's StartDate/EndDate (which are UTC in the DB)
// and interprets them as 00:00:00 and 23:59:59 in the schedule's timezone.
func calculateScheduleBoundaryTimesInLocation(schedule db.Schedule, defaultLoc *time.Location, logger *slog.Logger, ctx context.Context) (effectiveStartInLoc, effectiveEndInLoc time.Time, loc *time.Location, err error) {
	loc = defaultLoc // Start with default (usually UTC)
	if schedule.Timezone.Valid && schedule.Timezone.String != "" {
		loadedLoc, errLoadLoc := time.LoadLocation(schedule.Timezone.String)
		if errLoadLoc != nil {
			logger.WarnContext(ctx, "Failed to load timezone for schedule, using default",
				"schedule_id", schedule.ScheduleID, "timezone_str", schedule.Timezone.String, "error", errLoadLoc)
			// loc remains defaultLoc, return an error to indicate potential issue
			err = fmt.Errorf("failed to load location '%s': %w", schedule.Timezone.String, errLoadLoc)
			// Still return calculated times using defaultLoc, but caller should be aware of the error.
		} else {
			loc = loadedLoc
		}
	}

	var startDate, endDate time.Time // Will be zero if not set

	if schedule.StartDate.Valid {
		y, m, d := schedule.StartDate.Time.Date() // .Time is UTC from DB
		startDate = time.Date(y, m, d, 0, 0, 0, 0, loc)
	}

	if schedule.EndDate.Valid {
		y, m, d := schedule.EndDate.Time.Date() // .Time is UTC from DB
		endDate = time.Date(y, m, d, 23, 59, 59, 999999999, loc)
	}

	return startDate, endDate, loc, err // err will be nil if location loaded successfully or no timezone string
}

// GetUpcomingAvailableSlots finds only available (unbooked) shift slots
// across schedules that are active within the given time window.
// This method filters out already booked slots to show only available opportunities.
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

	allSchedules, err := s.querier.ListAllSchedules(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to list all schedules", "error", err)
		return nil, ErrInternalServer
	}

	if len(allSchedules) == 0 {
		s.logger.InfoContext(ctx, "No schedules defined in the system")
		return []AvailableShiftSlot{}, nil
	}

	var allSlots []AvailableShiftSlot

	for _, schedule := range allSchedules {
		scheduleActiveStartInLoc, scheduleActiveEndInLoc, loc, locErr := calculateScheduleBoundaryTimesInLocation(schedule, time.UTC, s.logger, ctx)
		if locErr != nil {
			s.logger.WarnContext(ctx, "Proceeding with default location for schedule due to load error", "schedule_id", schedule.ScheduleID, "error", locErr)
		}

		queryFromInLoc := actualFrom.In(loc)
		queryToInLoc := actualTo.In(loc)

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

		cronOccurrences, err := parseScheduleInTimezone(schedule.CronExpr, loc.String(), iterationStartInLoc, iterationEndInLoc)
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to parse cron expression", "schedule_id", schedule.ScheduleID, "cron_expr", schedule.CronExpr, "error", err)
			continue
		}

		for _, nextTime := range cronOccurrences {
			shiftEndTime := nextTime.Add(time.Duration(schedule.DurationMinutes) * time.Minute)
			slot := AvailableShiftSlot{
				ScheduleID:   schedule.ScheduleID,
				ScheduleName: schedule.Name,
				StartTime:    nextTime, // Already in UTC from parseScheduleInTimezone
				EndTime:      shiftEndTime,
				Timezone:     loc.String(),
				IsBooked:     false,
			}
			allSlots = append(allSlots, slot)
		}
	}

	// Batch retrieve bookings for the entire time range to avoid N+1 queries
	var populatedSlots []AvailableShiftSlot
	if len(allSlots) == 0 {
		return populatedSlots, nil
	}

	// Find the overall time range for batch booking query
	minTime := allSlots[0].StartTime
	maxTime := allSlots[0].EndTime
	for _, slot := range allSlots {
		if slot.StartTime.Before(minTime) {
			minTime = slot.StartTime
		}
		if slot.EndTime.After(maxTime) {
			maxTime = slot.EndTime
		}
	}

	// Batch retrieve all bookings in the time range
	allBookings, err := s.querier.GetBookingsInDateRange(ctx, db.GetBookingsInDateRangeParams{
		ShiftStart:   minTime.UTC(),
		ShiftStart_2: maxTime.UTC(),
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to batch retrieve bookings", "error", err)
		// Continue with unbooked slots if booking query fails
		for _, slot := range allSlots {
			populatedSlots = append(populatedSlots, slot)
		}
		return populatedSlots, nil
	}

	// Create lookup map: schedule_id + start_time -> booking details
	bookingMap := make(map[string]db.GetBookingsInDateRangeRow)
	for _, booking := range allBookings {
		key := fmt.Sprintf("%d_%s", booking.ScheduleID, booking.ShiftStart.UTC().Format(time.RFC3339))
		bookingMap[key] = booking
	}

	// Process each slot using the booking lookup map
	for _, slot := range allSlots {
		populatedSlot := slot
		slotKey := fmt.Sprintf("%d_%s", slot.ScheduleID, slot.StartTime.UTC().Format(time.RFC3339))
		
		if booking, exists := bookingMap[slotKey]; exists {
			// Booking exists - populate assignment details
			populatedSlot.IsBooked = true
			populatedSlot.BookingID = &booking.BookingID
			
			// User details are already included in the joined query
			if booking.UserName != "" {
				populatedSlot.UserName = &booking.UserName
			}
			if booking.UserPhone != "" {
				populatedSlot.UserPhone = &booking.UserPhone
			}
			
			// Handle buddy information from the joined query
			if booking.BuddyName.Valid && booking.BuddyName.String != "" {
				populatedSlot.BuddyName = &booking.BuddyName.String
			}
		} else {
			// No booking exists - keep as available slot
			populatedSlot.IsBooked = false
		}

		// Only include unbooked (available) slots in the result
		if !populatedSlot.IsBooked {
			populatedSlots = append(populatedSlots, populatedSlot)
		}
	}

	sort.Slice(populatedSlots, func(i, j int) bool {
		return populatedSlots[i].StartTime.Before(populatedSlots[j].StartTime)
	})

	if limit != nil && len(populatedSlots) > *limit {
		populatedSlots = populatedSlots[:*limit]
	}

	return populatedSlots, nil
}

// AdminGetAllShiftSlots finds all shift slots (booked or not)
// across all schedules that are active within the given time window,
// and includes booking details if a slot is booked OR reserved by recurring assignment.
func (s *ScheduleService) AdminGetAllShiftSlots(ctx context.Context, queryFrom *time.Time, queryTo *time.Time, limit *int) ([]AdminAvailableShiftSlot, error) {
	now := time.Now().UTC()
	defaultFrom := now
	defaultTo := now.AddDate(0, 0, 7)

	actualFrom := defaultFrom
	if queryFrom != nil {
		actualFrom = (*queryFrom).UTC()
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
		scheduleActiveStartInLoc, scheduleActiveEndInLoc, loc, locErr := calculateScheduleBoundaryTimesInLocation(schedule, time.UTC, s.logger, ctx)
		if locErr != nil {
			s.logger.WarnContext(ctx, "Proceeding with default location for schedule due to load error", "schedule_id", schedule.ScheduleID, "error", locErr)
		}

		queryFromInLoc := actualFrom.In(loc)
		queryToInLoc := actualTo.In(loc)

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

		cronOccurrences, err := parseScheduleInTimezone(schedule.CronExpr, loc.String(), iterationStartInLoc, iterationEndInLoc)
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to parse cron expression for admin slots", "schedule_id", schedule.ScheduleID, "cron_expr", schedule.CronExpr, "error", err)
			continue
		}

		for _, nextTime := range cronOccurrences {
			shiftEndTime := nextTime.Add(time.Duration(schedule.DurationMinutes) * time.Minute)
			slot := AdminAvailableShiftSlot{
				ScheduleID:   schedule.ScheduleID,
				ScheduleName: schedule.Name,
				StartTime:    nextTime, // Already in UTC from parseScheduleInTimezone
				EndTime:      shiftEndTime,
				Timezone:     loc.String(),
				IsBooked:     false,
			}
			allSlots = append(allSlots, slot)
		}
	}

	// Batch retrieve bookings for the entire time range to avoid N+1 queries
	var populatedSlots []AdminAvailableShiftSlot
	if len(allSlots) == 0 {
		return populatedSlots, nil
	}

	// Find the overall time range for batch booking query
	minTime := allSlots[0].StartTime
	maxTime := allSlots[0].EndTime
	for _, slot := range allSlots {
		if slot.StartTime.Before(minTime) {
			minTime = slot.StartTime
		}
		if slot.EndTime.After(maxTime) {
			maxTime = slot.EndTime
		}
	}

	// Batch retrieve all bookings in the time range
	allBookings, err := s.querier.GetBookingsInDateRange(ctx, db.GetBookingsInDateRangeParams{
		ShiftStart:   minTime.UTC(),
		ShiftStart_2: maxTime.UTC(),
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to batch retrieve bookings for admin", "error", err)
		// Continue with unbooked slots if booking query fails
		for _, slot := range allSlots {
			populatedSlots = append(populatedSlots, AdminAvailableShiftSlot(slot))
		}
		return populatedSlots, nil
	}

	// Create lookup map: schedule_id + start_time -> booking details
	bookingMap := make(map[string]db.GetBookingsInDateRangeRow)
	for _, booking := range allBookings {
		key := fmt.Sprintf("%d_%s", booking.ScheduleID, booking.ShiftStart.UTC().Format(time.RFC3339))
		bookingMap[key] = booking
	}

	// Process each slot using the booking lookup map
	for _, slot := range allSlots {
		populatedSlot := AdminAvailableShiftSlot(slot)
		slotKey := fmt.Sprintf("%d_%s", slot.ScheduleID, slot.StartTime.UTC().Format(time.RFC3339))
		
		if booking, exists := bookingMap[slotKey]; exists {
			// Booking exists - populate assignment details
			populatedSlot.IsBooked = true
			populatedSlot.BookingID = &booking.BookingID
			
			// User details are already included in the joined query
			if booking.UserName != "" {
				populatedSlot.UserName = &booking.UserName
			}
			if booking.UserPhone != "" {
				populatedSlot.UserPhone = &booking.UserPhone
			}
			
			// Handle buddy information from the joined query
			if booking.BuddyName.Valid && booking.BuddyName.String != "" {
				populatedSlot.BuddyName = &booking.BuddyName.String
			}
		} else {
			// No booking exists
			populatedSlot.IsBooked = false
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

// parseScheduleInTimezone parses a cron expression in the specified timezone
// and returns the next occurrence times as UTC timestamps.
// This follows ChatGPT's recommendation for proper timezone-aware cron parsing.
func parseScheduleInTimezone(cronExpr string, timezone string, fromTime, toTime time.Time) ([]time.Time, error) {
	// Load the timezone location
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("failed to load timezone %s: %w", timezone, err)
	}

	// Create a timezone-aware cron parser
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(cronExpr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cron expression %s: %w", cronExpr, err)
	}

	var occurrences []time.Time
	current := fromTime.In(loc)
	end := toTime.In(loc)

	// Generate occurrences in the target timezone
	for len(occurrences) < 1000 { // Prevent infinite loops
		next := schedule.Next(current)
		if next.IsZero() || next.After(end) {
			break
		}

		// Convert to UTC for storage/comparison
		occurrences = append(occurrences, next.UTC())
		current = next.Add(time.Minute) // Move past this occurrence
	}

	return occurrences, nil
}
