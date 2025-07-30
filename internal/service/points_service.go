package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	db "night-owls-go/internal/db/sqlc_generated"
)

// PointsService handles all points-related operations for the leaderboard system
type PointsService struct {
	querier *db.Queries
	db      *sql.DB
	logger  *slog.Logger
}

// NewPointsService creates a new PointsService
func NewPointsService(querier *db.Queries, database *sql.DB, logger *slog.Logger) *PointsService {
	return &PointsService{
		querier: querier,
		db:      database,
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

// CheckAndAwardAchievements is a public wrapper for checking and awarding achievements
func (ps *PointsService) CheckAndAwardAchievements(ctx context.Context, userID int64) error {
	return ps.checkAndAwardAchievements(ctx, userID)
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

// ===== ATOMIC OPERATIONS =====
// These methods use database transactions to ensure atomicity and consistency

// AtomicAwardShiftCheckinPoints atomically awards points when a user checks in to a shift
// This replaces AwardShiftCheckinPoints with proper transaction handling
func (ps *PointsService) AtomicAwardShiftCheckinPoints(ctx context.Context, userID int64, booking db.Booking) error {
	tx, err := ps.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	txQuerier := ps.querier.WithTx(tx)

	// Calculate all points and bonuses
	basePoints := PointsShiftCheckin
	totalPoints := basePoints
	pointEntries := []struct {
		points int
		reason PointReason
	}{
		{basePoints, ReasonShiftCheckin},
	}

	// Add bonuses
	if ps.isEarlyCheckin(booking) {
		pointEntries = append(pointEntries, struct {
			points int
			reason PointReason
		}{PointsEarlyCheckin, ReasonEarlyCheckin})
		totalPoints += PointsEarlyCheckin
	}

	if ps.isWeekendShift(booking) {
		pointEntries = append(pointEntries, struct {
			points int
			reason PointReason
		}{PointsWeekendShift, ReasonWeekendBonus})
		totalPoints += PointsWeekendShift
	}

	if ps.isLateNightShift(booking) {
		pointEntries = append(pointEntries, struct {
			points int
			reason PointReason
		}{PointsLateNightShift, ReasonLateNightBonus})
		totalPoints += PointsLateNightShift
	}

	// Insert all point history entries
	for _, entry := range pointEntries {
		if err := txQuerier.AwardPoints(ctx, db.AwardPointsParams{
			UserID:        userID,
			BookingID:     sql.NullInt64{Int64: booking.BookingID, Valid: true},
			PointsAwarded: int64(entry.points),
			Reason:        string(entry.reason),
			Multiplier:    sql.NullFloat64{Float64: 1.0, Valid: true},
		}); err != nil {
			return fmt.Errorf("failed to award points for %s: %w", entry.reason, err)
		}
	}

	// Update user's total points incrementally
	if err := txQuerier.UpdateUserTotalPointsIncremental(ctx, db.UpdateUserTotalPointsIncrementalParams{
		TotalPoints: sql.NullInt64{Int64: int64(totalPoints), Valid: true},
		UserID:      userID,
	}); err != nil {
		return fmt.Errorf("failed to update total points: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	ps.logger.InfoContext(ctx, "Atomically awarded shift check-in points",
		"user_id", userID, "booking_id", booking.BookingID, "total_points", totalPoints)

	return nil
}

// AtomicAwardShiftCommitmentPoints atomically awards points when a user commits to a shift
func (ps *PointsService) AtomicAwardShiftCommitmentPoints(ctx context.Context, userID int64, bookingID int64) error {
	tx, err := ps.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	txQuerier := ps.querier.WithTx(tx)

	// Award commitment points
	if err := txQuerier.AwardPoints(ctx, db.AwardPointsParams{
		UserID:        userID,
		BookingID:     sql.NullInt64{Int64: bookingID, Valid: true},
		PointsAwarded: int64(PointsShiftCommitment),
		Reason:        string(ReasonShiftCommitment),
		Multiplier:    sql.NullFloat64{Float64: 1.0, Valid: true},
	}); err != nil {
		return fmt.Errorf("failed to award commitment points: %w", err)
	}

	// Update user's total points incrementally
	if err := txQuerier.UpdateUserTotalPointsIncremental(ctx, db.UpdateUserTotalPointsIncrementalParams{
		TotalPoints: sql.NullInt64{Int64: int64(PointsShiftCommitment), Valid: true},
		UserID:      userID,
	}); err != nil {
		return fmt.Errorf("failed to update total points: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	ps.logger.InfoContext(ctx, "Atomically awarded shift commitment points",
		"user_id", userID, "booking_id", bookingID, "points", PointsShiftCommitment)

	return nil
}

// AtomicAwardShiftDropoutPoints atomically deducts points when a user drops out of a shift
func (ps *PointsService) AtomicAwardShiftDropoutPoints(ctx context.Context, userID int64, bookingID int64) error {
	tx, err := ps.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	txQuerier := ps.querier.WithTx(tx)

	// Award dropout points (negative)
	if err := txQuerier.AwardPoints(ctx, db.AwardPointsParams{
		UserID:        userID,
		BookingID:     sql.NullInt64{Int64: bookingID, Valid: true},
		PointsAwarded: int64(PointsShiftDropout), // This is negative (-10)
		Reason:        string(ReasonShiftDropout),
		Multiplier:    sql.NullFloat64{Float64: 1.0, Valid: true},
	}); err != nil {
		return fmt.Errorf("failed to award dropout points: %w", err)
	}

	// Update user's total points incrementally
	if err := txQuerier.UpdateUserTotalPointsIncremental(ctx, db.UpdateUserTotalPointsIncrementalParams{
		TotalPoints: sql.NullInt64{Int64: int64(PointsShiftDropout), Valid: true}, // Negative value
		UserID:      userID,
	}); err != nil {
		return fmt.Errorf("failed to update total points: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	ps.logger.InfoContext(ctx, "Atomically awarded shift dropout points",
		"user_id", userID, "booking_id", bookingID, "points", PointsShiftDropout)

	return nil
}

// AtomicAwardShiftCompletionPoints atomically awards points when a user completes a shift
func (ps *PointsService) AtomicAwardShiftCompletionPoints(ctx context.Context, userID int64, bookingID int64, reportSeverity int) error {
	tx, err := ps.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	txQuerier := ps.querier.WithTx(tx)

	totalPoints := PointsShiftCompletion
	pointEntries := []struct {
		points int
		reason PointReason
	}{
		{PointsShiftCompletion, ReasonShiftCompletion},
		{PointsReportFiled, ReasonReportFiled},
	}

	// Add level 2 report bonus
	if reportSeverity >= 2 {
		pointEntries = append(pointEntries, struct {
			points int
			reason PointReason
		}{PointsLevel2Report, ReasonLevel2Report})
		totalPoints += PointsReportFiled + PointsLevel2Report
	} else {
		totalPoints += PointsReportFiled
	}

	// Check for frequency bonus
	frequencyEligible, err := txQuerier.CheckFrequencyBonusEligibility(ctx, userID)
	if err != nil {
		ps.logger.WarnContext(ctx, "Failed to check frequency bonus eligibility", "user_id", userID, "error", err)
	} else if frequencyEligible > 0 { // Already has completed shifts this month
		pointEntries = append(pointEntries, struct {
			points int
			reason PointReason
		}{PointsFrequencyBonus, ReasonFrequencyBonus})
		totalPoints += PointsFrequencyBonus
	}

	// Insert all point history entries
	for _, entry := range pointEntries {
		if err := txQuerier.AwardPoints(ctx, db.AwardPointsParams{
			UserID:        userID,
			BookingID:     sql.NullInt64{Int64: bookingID, Valid: true},
			PointsAwarded: int64(entry.points),
			Reason:        string(entry.reason),
			Multiplier:    sql.NullFloat64{Float64: 1.0, Valid: true},
		}); err != nil {
			return fmt.Errorf("failed to award points for %s: %w", entry.reason, err)
		}
	}

	// Atomically update both points and shift count
	if err := txQuerier.UpdateUserPointsAndShiftCount(ctx, db.UpdateUserPointsAndShiftCountParams{
		TotalPoints: sql.NullInt64{Int64: int64(totalPoints), Valid: true},
		ShiftCount:  sql.NullInt64{Int64: 1, Valid: true}, // Increment by 1
		UserID:      userID,
	}); err != nil {
		return fmt.Errorf("failed to update points and shift count: %w", err)
	}

	// Check for achievements
	if err := ps.checkAndAwardAchievementsInTx(ctx, txQuerier, userID); err != nil {
		ps.logger.WarnContext(ctx, "Failed to check achievements", "user_id", userID, "error", err)
		// Non-fatal - don't fail the entire operation
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	ps.logger.InfoContext(ctx, "Atomically awarded shift completion points",
		"user_id", userID, "booking_id", bookingID, "total_points", totalPoints)

	return nil
}

// checkAndAwardAchievementsInTx checks and awards achievements within an existing transaction
func (ps *PointsService) checkAndAwardAchievementsInTx(ctx context.Context, txQuerier db.Querier, userID int64) error {
	// Get user's current stats
	userStats, err := txQuerier.GetUserCurrentPoints(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user stats: %w", err)
	}

	// Get available achievements
	availableAchievements, err := txQuerier.GetAvailableAchievements(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get available achievements: %w", err)
	}

	// Award achievements based on shift count
	for _, achievement := range availableAchievements {
		if achievement.ShiftsThreshold.Valid && userStats.ShiftCount.Valid &&
		   userStats.ShiftCount.Int64 >= achievement.ShiftsThreshold.Int64 {
			if err := txQuerier.AwardAchievement(ctx, db.AwardAchievementParams{
				UserID:        userID,
				AchievementID: achievement.AchievementID,
			}); err != nil {
				ps.logger.WarnContext(ctx, "Failed to award achievement",
					"user_id", userID, "achievement_id", achievement.AchievementID, "error", err)
				// Continue with other achievements
			} else {
				ps.logger.InfoContext(ctx, "Awarded achievement",
					"user_id", userID, "achievement", achievement.Name)
			}
		}
	}

	return nil
}

// ===== ERROR RECOVERY AND RETRY MECHANISMS =====

// PointsOperationWithRetry executes a points operation with retry logic for transient failures
func (ps *PointsService) PointsOperationWithRetry(ctx context.Context, operation func() error, operationName string) error {
	const maxRetries = 3
	const baseDelay = 100 * time.Millisecond

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff with jitter - use safe calculation to prevent overflow
			// Cap the exponent to prevent integer overflow
			exponent := attempt - 1
			if exponent > 10 { // Cap at 2^10 = 1024x multiplier
				exponent = 10
			}
			delay := baseDelay * time.Duration(1<<exponent)
			ps.logger.InfoContext(ctx, "Retrying points operation",
				"operation", operationName, "attempt", attempt+1, "delay_ms", delay.Milliseconds())
			
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		err := operation()
		if err == nil {
			if attempt > 0 {
				ps.logger.InfoContext(ctx, "Points operation succeeded after retry",
					"operation", operationName, "attempt", attempt+1)
			}
			return nil
		}

		lastErr = err
		
		// Check if error is retryable (database lock, temporary network issues, etc.)
		if !ps.isRetryableError(err) {
			ps.logger.WarnContext(ctx, "Points operation failed with non-retryable error",
				"operation", operationName, "error", err)
			return err
		}

		ps.logger.WarnContext(ctx, "Points operation failed, will retry",
			"operation", operationName, "attempt", attempt+1, "error", err)
	}

	ps.logger.ErrorContext(ctx, "Points operation failed after all retries",
		"operation", operationName, "max_retries", maxRetries, "final_error", lastErr)
	return fmt.Errorf("points operation '%s' failed after %d retries: %w", operationName, maxRetries, lastErr)
}

// isRetryableError determines if an error is worth retrying
func (ps *PointsService) isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	
	// SQLite specific retryable errors
	retryablePatterns := []string{
		"database is locked",
		"database is busy",
		"no such table", // Could be during migration
		"constraint failed", // Could be temporary constraint violations during concurrent operations
		"connection reset",
		"connection refused",
		"timeout",
		"temporary",
	}
	
	for _, pattern := range retryablePatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return true
		}
	}
	
	return false
}

// AtomicAwardShiftCheckinPointsWithRetry wraps the atomic operation with retry logic
func (ps *PointsService) AtomicAwardShiftCheckinPointsWithRetry(ctx context.Context, userID int64, booking db.Booking) error {
	return ps.PointsOperationWithRetry(ctx, func() error {
		return ps.AtomicAwardShiftCheckinPoints(ctx, userID, booking)
	}, "shift_checkin_points")
}

// AtomicAwardShiftCommitmentPointsWithRetry wraps the atomic operation with retry logic
func (ps *PointsService) AtomicAwardShiftCommitmentPointsWithRetry(ctx context.Context, userID int64, bookingID int64) error {
	return ps.PointsOperationWithRetry(ctx, func() error {
		return ps.AtomicAwardShiftCommitmentPoints(ctx, userID, bookingID)
	}, "shift_commitment_points")
}

// AtomicAwardShiftDropoutPointsWithRetry wraps the atomic operation with retry logic
func (ps *PointsService) AtomicAwardShiftDropoutPointsWithRetry(ctx context.Context, userID int64, bookingID int64) error {
	return ps.PointsOperationWithRetry(ctx, func() error {
		return ps.AtomicAwardShiftDropoutPoints(ctx, userID, bookingID)
	}, "shift_dropout_points")
}

// AtomicAwardShiftCompletionPointsWithRetry wraps the atomic operation with retry logic
func (ps *PointsService) AtomicAwardShiftCompletionPointsWithRetry(ctx context.Context, userID int64, bookingID int64, reportSeverity int) error {
	return ps.PointsOperationWithRetry(ctx, func() error {
		return ps.AtomicAwardShiftCompletionPoints(ctx, userID, bookingID, reportSeverity)
	}, "shift_completion_points")
}
