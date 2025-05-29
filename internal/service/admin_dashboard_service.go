package service

import (
	"context"
	"log/slog"
	"time"

	db "night-owls-go/internal/db/sqlc_generated"
)

// Helper function to safely convert interface{} to float64
func toFloat64(v interface{}) float64 {
	if v == nil {
		return 0.0
	}
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	default:
		return 0.0
	}
}

type AdminDashboardService struct {
	querier         db.Querier
	scheduleService *ScheduleService
	logger          *slog.Logger
}

// Dashboard metric types
type DashboardMetrics struct {
	TotalShifts       int     `json:"total_shifts"`
	BookedShifts      int     `json:"booked_shifts"`
	UnfilledShifts    int     `json:"unfilled_shifts"`
	CheckedInShifts   int     `json:"checked_in_shifts"`
	CompletedShifts   int     `json:"completed_shifts"`
	FillRate          float64 `json:"fill_rate"`
	CheckInRate       float64 `json:"check_in_rate"`
	CompletionRate    float64 `json:"completion_rate"`
	NextWeekUnfilled  int     `json:"next_week_unfilled"`
	ThisWeekendStatus string  `json:"this_weekend_status"`
}

type MemberContribution struct {
	UserID               int64      `json:"user_id"`
	Name                 string     `json:"name"`
	Phone                string     `json:"phone"`
	ShiftsBooked         int        `json:"shifts_booked"`
	ShiftsAttended       int        `json:"shifts_attended"`
	ShiftsCompleted      int        `json:"shifts_completed"`
	AttendanceRate       float64    `json:"attendance_rate"`
	CompletionRate       float64    `json:"completion_rate"`
	LastShiftDate        *time.Time `json:"last_shift_date"`
	ContributionCategory string     `json:"contribution_category"`
}

type QualityMetrics struct {
	NoShowRate       float64 `json:"no_show_rate"`
	IncompleteRate   float64 `json:"incomplete_rate"`
	ReliabilityScore float64 `json:"reliability_score"`
}

type TimeSlotPattern struct {
	DayOfWeek      string  `json:"day_of_week"`
	HourOfDay      string  `json:"hour_of_day"`
	TotalBookings  int     `json:"total_bookings"`
	CheckInRate    float64 `json:"check_in_rate"`
	CompletionRate float64 `json:"completion_rate"`
}

type AdminDashboard struct {
	Metrics             DashboardMetrics     `json:"metrics"`
	MemberContributions []MemberContribution `json:"member_contributions"`
	QualityMetrics      QualityMetrics       `json:"quality_metrics"`
	ProblematicSlots    []TimeSlotPattern    `json:"problematic_slots"`
	GeneratedAt         time.Time            `json:"generated_at"`
}

func NewAdminDashboardService(querier db.Querier, scheduleService *ScheduleService, logger *slog.Logger) *AdminDashboardService {
	return &AdminDashboardService{
		querier:         querier,
		scheduleService: scheduleService,
		logger:          logger,
	}
}

// GetDashboard generates comprehensive admin dashboard metrics
func (s *AdminDashboardService) GetDashboard(ctx context.Context) (*AdminDashboard, error) {
	now := time.Now().UTC()
	twoWeeksFromNow := now.AddDate(0, 0, 14)

	// Calculate metrics for next 2 weeks
	metrics := s.calculateDashboardMetrics(ctx, now, twoWeeksFromNow)

	// Get member contributions (past 30 days)
	contributions := s.getMemberContributions(ctx)

	// Calculate quality metrics
	qualityMetrics, err := s.calculateQualityMetrics(ctx, now.AddDate(0, 0, -30), now)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to calculate quality metrics", "error", err)
		qualityMetrics = &QualityMetrics{} // Return default metrics instead of failing
	}

	// Get problematic time slots
	problematicSlots, err := s.getProblematicTimeSlots(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to get problematic time slots", "error", err)
		problematicSlots = []TimeSlotPattern{} // Return empty slice instead of failing
	}

	return &AdminDashboard{
		Metrics:             *metrics,
		MemberContributions: contributions,
		QualityMetrics:      *qualityMetrics,
		ProblematicSlots:    problematicSlots,
		GeneratedAt:         now,
	}, nil
}

func (s *AdminDashboardService) calculateDashboardMetrics(ctx context.Context, from, to time.Time) *DashboardMetrics {
	// Get all available slots for the period using existing schedule service
	allSlots, err := s.scheduleService.AdminGetAllShiftSlots(ctx, &from, &to, nil)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to get all shift slots", "error", err)
		// Return default metrics instead of failing
		allSlots = []AdminAvailableShiftSlot{}
	}

	// Get booking metrics for the same period
	bookingMetrics, err := s.querier.GetBookingMetrics(ctx, db.GetBookingMetricsParams{
		ShiftStart:   from,
		ShiftStart_2: to,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to get booking metrics", "error", err)
		// Return default metrics instead of failing
		bookingMetrics = db.GetBookingMetricsRow{
			TotalBookings:     0,
			CheckedInBookings: 0,
			CompletedBookings: 0,
			CheckInRate:       0.0,
			CompletionRate:    0.0,
		}
	}

	totalShifts := len(allSlots)
	bookedShifts := int(bookingMetrics.TotalBookings)
	unfilledShifts := totalShifts - bookedShifts
	checkedInShifts := int(bookingMetrics.CheckedInBookings)
	completedShifts := int(bookingMetrics.CompletedBookings)

	// Calculate fill rate
	fillRate := 0.0
	if totalShifts > 0 {
		fillRate = float64(bookedShifts) / float64(totalShifts) * 100
	}

	// Calculate weekend status
	weekendStatus := s.getWeekendStatus(allSlots)

	// Count unfilled shifts in next week
	nextWeek := from.AddDate(0, 0, 7)
	nextWeekUnfilled := 0
	for _, slot := range allSlots {
		if slot.StartTime.Before(nextWeek) && !slot.IsBooked {
			nextWeekUnfilled++
		}
	}

	return &DashboardMetrics{
		TotalShifts:       totalShifts,
		BookedShifts:      bookedShifts,
		UnfilledShifts:    unfilledShifts,
		CheckedInShifts:   checkedInShifts,
		CompletedShifts:   completedShifts,
		FillRate:          fillRate,
		CheckInRate:       float64(bookingMetrics.CheckInRate),
		CompletionRate:    toFloat64(bookingMetrics.CompletionRate),
		NextWeekUnfilled:  nextWeekUnfilled,
		ThisWeekendStatus: weekendStatus,
	}
}

func (s *AdminDashboardService) getMemberContributions(ctx context.Context) []MemberContribution {
	contributions, err := s.querier.GetMemberContributions(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to get member contributions", "error", err)
		return []MemberContribution{}
	}

	result := make([]MemberContribution, len(contributions))
	for i, contrib := range contributions {
		var lastShiftDate *time.Time
		if contrib.LastShiftDate != nil {
			if t, ok := contrib.LastShiftDate.(time.Time); ok {
				lastShiftDate = &t
			}
		}

		contributionCategory := ""
		if contrib.ContributionCategory != nil {
			if cat, ok := contrib.ContributionCategory.(string); ok {
				contributionCategory = cat
			}
		}

		result[i] = MemberContribution{
			UserID:               contrib.UserID,
			Name:                 contrib.Name.String,
			Phone:                contrib.Phone,
			ShiftsBooked:         int(contrib.ShiftsBooked),
			ShiftsAttended:       int(contrib.ShiftsAttended),
			ShiftsCompleted:      int(contrib.ShiftsCompleted),
			AttendanceRate:       toFloat64(contrib.AttendanceRate),
			CompletionRate:       toFloat64(contrib.CompletionRate),
			LastShiftDate:        lastShiftDate,
			ContributionCategory: contributionCategory,
		}
	}

	return result
}

func (s *AdminDashboardService) calculateQualityMetrics(ctx context.Context, from, to time.Time) (*QualityMetrics, error) {
	bookings, err := s.querier.GetBookingsInDateRange(ctx, db.GetBookingsInDateRangeParams{
		ShiftStart:   from,
		ShiftStart_2: to,
	})
	if err != nil {
		return nil, err
	}

	if len(bookings) == 0 {
		return &QualityMetrics{}, nil
	}

	totalBookings := len(bookings)
	noShows := 0
	incomplete := 0
	complete := 0

	for _, booking := range bookings {
		if !booking.CheckedInAt.Valid {
			noShows++
		} else if booking.HasReport == 0 {
			incomplete++
		} else {
			complete++
		}
	}

	noShowRate := float64(noShows) / float64(totalBookings) * 100
	incompleteRate := float64(incomplete) / float64(totalBookings-noShows) * 100
	reliabilityScore := float64(complete) / float64(totalBookings) * 100

	return &QualityMetrics{
		NoShowRate:       noShowRate,
		IncompleteRate:   incompleteRate,
		ReliabilityScore: reliabilityScore,
	}, nil
}

func (s *AdminDashboardService) getProblematicTimeSlots(ctx context.Context) ([]TimeSlotPattern, error) {
	patterns, err := s.querier.GetBookingPatternsByTimeSlot(ctx)
	if err != nil {
		return nil, err
	}

	dayNames := map[string]string{
		"0": "Sunday",
		"1": "Monday",
		"2": "Tuesday",
		"3": "Wednesday",
		"4": "Thursday",
		"5": "Friday",
		"6": "Saturday",
	}

	result := make([]TimeSlotPattern, len(patterns))
	for i, pattern := range patterns {
		dayOfWeekStr := ""
		if pattern.DayOfWeek != nil {
			if dow, ok := pattern.DayOfWeek.(string); ok {
				dayOfWeekStr = dow
			}
		}

		dayName, exists := dayNames[dayOfWeekStr]
		if !exists {
			dayName = "Unknown"
		}

		hourOfDayStr := ""
		if pattern.HourOfDay != nil {
			if hod, ok := pattern.HourOfDay.(string); ok {
				hourOfDayStr = hod
			}
		}

		result[i] = TimeSlotPattern{
			DayOfWeek:      dayName,
			HourOfDay:      hourOfDayStr + ":00",
			TotalBookings:  int(pattern.TotalBookings),
			CheckInRate:    float64(pattern.CheckInRate),
			CompletionRate: toFloat64(pattern.CompletionRate),
		}
	}

	return result, nil
}

func (s *AdminDashboardService) getWeekendStatus(slots []AdminAvailableShiftSlot) string {
	now := time.Now()
	thisWeekend := now
	for thisWeekend.Weekday() != time.Saturday {
		thisWeekend = thisWeekend.AddDate(0, 0, 1)
	}
	nextSunday := thisWeekend.AddDate(0, 0, 1)

	unfilledCount := 0
	totalWeekendShifts := 0

	for _, slot := range slots {
		if slot.StartTime.After(thisWeekend) && slot.StartTime.Before(nextSunday.AddDate(0, 0, 1)) {
			totalWeekendShifts++
			if !slot.IsBooked {
				unfilledCount++
			}
		}
	}

	if totalWeekendShifts == 0 {
		return "no_shifts"
	}
	if unfilledCount == 0 {
		return "fully_covered"
	}
	if unfilledCount >= totalWeekendShifts/2 {
		return "critical"
	}
	return "partial_coverage"
}
