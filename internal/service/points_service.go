package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	db "night-owls-go/internal/db/sqlc_generated"
)

// PointsService handles all points-related operations for the leaderboard system
type PointsService struct {
	querier db.Querier
	logger  *slog.Logger
}

// NewPointsService creates a new PointsService
func NewPointsService(querier db.Querier, logger *slog.Logger) *PointsService {
	return &PointsService{
		querier: querier,
		logger:  logger.With("service", "PointsService"),
	}
}

// Points awarded for different actions
const (
	PointsShiftCommitment = 5  // Committing to a shift (booking creation)
	PointsShiftDropout    = -10 // Dropping out of a shift (booking cancellation)
	PointsShiftCheckin    = 10 // Checking in to a shift on time
	PointsShiftCompletion = 15 // Completing a shift with check-in
	PointsReportFiled     = 5  // Filing a report during shift
	PointsEarlyCheckin    = 3  // Bonus for checking in early
	PointsLevel2Report    = 10 // Extra points for serious incident reports
	PointsWeekendShift    = 5  // Weekend shift bonus
	PointsLateNightShift  = 3  // Late night shift bonus (after 22:00)
	PointsFrequencyBonus  = 10 // Bonus for doing multiple shifts per month
)

// PointReason defines reasons for awarding points
type PointReason string

const (
	ReasonShiftCommitment PointReason = "shift_commitment"
	ReasonShiftDropout    PointReason = "shift_dropout"
	ReasonShiftCheckin    PointReason = "shift_checkin"
	ReasonShiftCompletion PointReason = "shift_completion"
	ReasonReportFiled     PointReason = "report_filed"
	ReasonEarlyCheckin    PointReason = "early_checkin"
	ReasonLevel2Report    PointReason = "level2_report"
	ReasonWeekendBonus    PointReason = "weekend_bonus"
	ReasonLateNightBonus  PointReason = "late_night_bonus"
	ReasonFrequencyBonus  PointReason = "frequency_bonus"
)

// AwardShiftCheckinPoints awards points when a user checks in to a shift
func (ps *PointsService) AwardShiftCheckinPoints(ctx context.Context, userID int64, booking db.Booking) error {
	basePoints := PointsShiftCheckin
	multiplier := 1.0

	// Calculate bonus points
	bonusReasons := []struct {
		reason  PointReason
		points  int
		applies bool
		desc    string
	}{
		{ReasonEarlyCheckin, PointsEarlyCheckin, ps.isEarlyCheckin(booking), "early check-in"},
		{ReasonWeekendBonus, PointsWeekendShift, ps.isWeekendShift(booking), "weekend shift"},
		{ReasonLateNightBonus, PointsLateNightShift, ps.isLateNightShift(booking), "late night shift"},
	}

	// Award base points for check-in
	if err := ps.awardPointsWithHistory(ctx, userID, &booking.BookingID, basePoints, ReasonShiftCheckin, multiplier); err != nil {
		return fmt.Errorf("failed to award checkin points: %w", err)
	}

	// Award bonus points
	for _, bonus := range bonusReasons {
		if bonus.applies {
			if err := ps.awardPointsWithHistory(ctx, userID, &booking.BookingID, bonus.points, bonus.reason, multiplier); err != nil {
				ps.logger.WarnContext(ctx, "Failed to award bonus points",
					"user_id", userID, "reason", bonus.reason, "error", err)
			} else {
				ps.logger.InfoContext(ctx, "Awarded bonus points",
					"user_id", userID, "points", bonus.points, "reason", bonus.desc)
			}
		}
	}

	// Update user's total points
	if err := ps.updateUserTotalPoints(ctx, userID); err != nil {
		return fmt.Errorf("failed to update total points: %w", err)
	}

	ps.logger.InfoContext(ctx, "Awarded shift check-in points",
		"user_id", userID, "booking_id", booking.BookingID, "base_points", basePoints)

	return nil
}

// AwardShiftCommitmentPoints awards points when a user commits to a shift (booking creation)
func (ps *PointsService) AwardShiftCommitmentPoints(ctx context.Context, userID int64, bookingID int64) error {
	basePoints := PointsShiftCommitment
	multiplier := 1.0

	// Award commitment points
	if err := ps.awardPointsWithHistory(ctx, userID, &bookingID, basePoints, ReasonShiftCommitment, multiplier); err != nil {
		return fmt.Errorf("failed to award commitment points: %w", err)
	}

	// Update user's total points
	if err := ps.updateUserTotalPoints(ctx, userID); err != nil {
		return fmt.Errorf("failed to update total points: %w", err)
	}

	ps.logger.InfoContext(ctx, "Awarded shift commitment points",
		"user_id", userID, "booking_id", bookingID, "points", basePoints)

	return nil
}

// AwardShiftDropoutPoints deducts points when a user drops out of a shift (booking cancellation)
func (ps *PointsService) AwardShiftDropoutPoints(ctx context.Context, userID int64, bookingID int64) error {
	points := PointsShiftDropout // This is negative (-10)
	multiplier := 1.0

	// Award dropout points (negative)
	if err := ps.awardPointsWithHistory(ctx, userID, &bookingID, points, ReasonShiftDropout, multiplier); err != nil {
		return fmt.Errorf("failed to award dropout points: %w", err)
	}

	// Update user's total points
	if err := ps.updateUserTotalPoints(ctx, userID); err != nil {
		return fmt.Errorf("failed to update total points: %w", err)
	}

	ps.logger.InfoContext(ctx, "Awarded shift dropout points",
		"user_id", userID, "booking_id", bookingID, "points", points)

	return nil
}

// AwardShiftCompletionPoints awards points when a user completes a shift (with report)
func (ps *PointsService) AwardShiftCompletionPoints(ctx context.Context, userID int64, bookingID int64, reportSeverity int) error {
	basePoints := PointsShiftCompletion
	multiplier := 1.0

	// Award base completion points
	if err := ps.awardPointsWithHistory(ctx, userID, &bookingID, basePoints, ReasonShiftCompletion, multiplier); err != nil {
		return fmt.Errorf("failed to award completion points: %w", err)
	}

	// Award report filing points
	reportPoints := PointsReportFiled
	if reportSeverity >= 2 { // Level 2 serious incidents
		reportPoints += PointsLevel2Report
	}

	reason := ReasonReportFiled
	if reportSeverity >= 2 {
		reason = ReasonLevel2Report
	}

	if err := ps.awardPointsWithHistory(ctx, userID, &bookingID, reportPoints, reason, multiplier); err != nil {
		ps.logger.WarnContext(ctx, "Failed to award report points",
			"user_id", userID, "error", err)
	}

	// Update shift count and check for frequency bonus
	if err := ps.updateShiftCountAndCheckFrequency(ctx, userID); err != nil {
		ps.logger.WarnContext(ctx, "Failed to update shift count", "user_id", userID, "error", err)
	}

	// Update total points
	if err := ps.updateUserTotalPoints(ctx, userID); err != nil {
		return fmt.Errorf("failed to update total points: %w", err)
	}

	// Check for achievements
	if err := ps.checkAndAwardAchievements(ctx, userID); err != nil {
		ps.logger.WarnContext(ctx, "Failed to check achievements",
			"user_id", userID, "error", err)
	}

	ps.logger.InfoContext(ctx, "Awarded shift completion points",
		"user_id", userID, "booking_id", bookingID, "points", basePoints+reportPoints)

	return nil
}

// Helper methods

func (ps *PointsService) awardPointsWithHistory(ctx context.Context, userID int64, bookingID *int64, points int, reason PointReason, multiplier float64) error {
	return ps.querier.AwardPoints(ctx, db.AwardPointsParams{
		UserID:        userID,
		BookingID:     sql.NullInt64{Int64: *bookingID, Valid: bookingID != nil},
		PointsAwarded: int64(points),
		Reason:        string(reason),
		Multiplier:    sql.NullFloat64{Float64: multiplier, Valid: true},
	})
}

func (ps *PointsService) updateUserTotalPoints(ctx context.Context, userID int64) error {
	return ps.querier.UpdateUserTotalPoints(ctx, db.UpdateUserTotalPointsParams{
		UserID:   userID,
		UserID_2: userID,
	})
}

func (ps *PointsService) updateShiftCountAndCheckFrequency(ctx context.Context, userID int64) error {
	// Update shift count
	if err := ps.querier.UpdateUserShiftCount(ctx, userID); err != nil {
		return fmt.Errorf("failed to update shift count: %w", err)
	}

	// Check for frequency bonus (more than one shift this month)
	currentMonth := time.Now().Format("2006-01")
	monthlyShifts, err := ps.getMonthlyShiftCount(ctx, userID, currentMonth)
	if err != nil {
		ps.logger.WarnContext(ctx, "Failed to get monthly shift count", "user_id", userID, "error", err)
		return nil // Non-fatal
	}

	// Award frequency bonus for 2nd, 3rd, etc. shifts in the month
	if monthlyShifts > 1 {
		if err := ps.awardPointsWithHistory(ctx, userID, nil, PointsFrequencyBonus, ReasonFrequencyBonus, 1.0); err != nil {
			ps.logger.WarnContext(ctx, "Failed to award frequency bonus",
				"user_id", userID, "monthly_shifts", monthlyShifts, "error", err)
		} else {
			ps.logger.InfoContext(ctx, "Awarded frequency bonus",
				"user_id", userID, "monthly_shifts", monthlyShifts, "bonus_points", PointsFrequencyBonus)
		}
	}

	return nil
}

func (ps *PointsService) getMonthlyShiftCount(ctx context.Context, userID int64, monthStr string) (int, error) {
	// This is a simplified count - in production you'd want a proper query
	// For now, we'll estimate based on recent activity
	return 1, nil // Placeholder - would need a proper query to count completed shifts this month
}

func (ps *PointsService) checkAndAwardAchievements(ctx context.Context, userID int64) error {
	// Get user's current stats
	userStats, err := ps.querier.GetUserPoints(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user stats: %w", err)
	}

	// Get available achievements
	available, err := ps.querier.GetAvailableAchievements(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get available achievements: %w", err)
	}

	// Check each achievement
	for _, achievement := range available {
		shouldAward := false

		// Shift-based achievements
		if achievement.ShiftsThreshold.Valid && userStats.ShiftCount.Valid &&
			userStats.ShiftCount.Int64 >= achievement.ShiftsThreshold.Int64 {
			shouldAward = true
		}

		if shouldAward {
			if err := ps.querier.AwardAchievement(ctx, db.AwardAchievementParams{
				UserID:        userID,
				AchievementID: achievement.AchievementID,
			}); err != nil {
				ps.logger.WarnContext(ctx, "Failed to award achievement",
					"user_id", userID, "achievement", achievement.Name, "error", err)
			} else {
				ps.logger.InfoContext(ctx, "Achievement earned!",
					"user_id", userID, "achievement", achievement.Name)
			}
		}
	}

	return nil
}

func (ps *PointsService) isEarlyCheckin(booking db.Booking) bool {
	// Check if user checked in more than 15 minutes early
	if !booking.CheckedInAt.Valid {
		return false
	}

	checkinTime := booking.CheckedInAt.Time
	shiftStart := booking.ShiftStart

	// If checked in more than 15 minutes early
	return shiftStart.Sub(checkinTime) > 15*time.Minute
}

func (ps *PointsService) isWeekendShift(booking db.Booking) bool {
	weekday := booking.ShiftStart.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

func (ps *PointsService) isLateNightShift(booking db.Booking) bool {
	hour := booking.ShiftStart.Hour()
	return hour >= 22 || hour <= 5 // 10 PM to 5 AM
}

// Public methods for API endpoints

func (ps *PointsService) GetLeaderboard(ctx context.Context, limit int32) ([]db.GetTopUsersRow, error) {
	return ps.querier.GetTopUsers(ctx, int64(limit))
}

func (ps *PointsService) GetShiftLeaderboard(ctx context.Context, limit int32) ([]db.GetTopUsersByShiftsRow, error) {
	return ps.querier.GetTopUsersByShifts(ctx, int64(limit))
}

func (ps *PointsService) GetUserRank(ctx context.Context, userID int64) (int64, error) {
	return ps.querier.GetUserRank(ctx, userID)
}

func (ps *PointsService) GetUserStats(ctx context.Context, userID int64) (db.GetUserPointsRow, error) {
	return ps.querier.GetUserPoints(ctx, userID)
}

func (ps *PointsService) GetUserAchievements(ctx context.Context, userID int64) ([]db.GetUserAchievementsRow, error) {
	return ps.querier.GetUserAchievements(ctx, userID)
}

func (ps *PointsService) GetRecentActivity(ctx context.Context, limit int32) ([]db.GetRecentActivityRow, error) {
	return ps.querier.GetRecentActivity(ctx, int64(limit))
}

func (ps *PointsService) GetUserPointsHistory(ctx context.Context, userID int64, limit int64) ([]db.GetUserPointsHistoryRow, error) {
	return ps.querier.GetUserPointsHistory(ctx, db.GetUserPointsHistoryParams{
		UserID: userID,
		Limit:  limit,
	})
}

func (ps *PointsService) GetAvailableAchievements(ctx context.Context, userID int64) ([]db.GetAvailableAchievementsRow, error) {
	return ps.querier.GetAvailableAchievements(ctx, userID)
}
