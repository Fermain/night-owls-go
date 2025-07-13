package migration

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	db "night-owls-go/internal/db/sqlc_generated"
	"night-owls-go/internal/service"
)

type PointsMigrator struct {
	queries       db.Querier
	pointsService *service.PointsService
	logger        *slog.Logger
	dryRun        bool
	dbConn        *sql.DB
}

type MigrationSummary struct {
	TotalBookings         int
	BookingsNeedingPoints int
	CheckInPoints         int
	CompletionPoints      int
	BonusOpportunities    int
	EstimatedTotalPoints  int
	AffectedUsers         int
}

func NewPointsMigrator(queries db.Querier, pointsService *service.PointsService, dbConn *sql.DB, logger *slog.Logger, dryRun bool) *PointsMigrator {
	return &PointsMigrator{
		queries:       queries,
		pointsService: pointsService,
		logger:        logger,
		dryRun:        dryRun,
		dbConn:        dbConn,
	}
}

func (m *PointsMigrator) Preview(ctx context.Context) (*MigrationSummary, error) {
	m.logger.Info("🔍 Analyzing historical bookings...")

	// Get all bookings without existing points
	bookingsWithoutPoints, err := m.getBookingsWithoutPoints(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get bookings: %w", err)
	}

	summary := &MigrationSummary{
		TotalBookings: len(bookingsWithoutPoints),
	}

	userMap := make(map[int64]bool)
	estimatedPoints := 0

	for _, booking := range bookingsWithoutPoints {
		summary.BookingsNeedingPoints++
		userMap[booking.UserID] = true

		// Check-in points (10 points)
		if booking.CheckedInAt.Valid {
			summary.CheckInPoints++
			estimatedPoints += 10

			// Bonus calculations
			if m.isEarlyCheckIn(booking) {
				summary.BonusOpportunities++
				estimatedPoints += 3
			}
			if m.isWeekendShift(booking) {
				summary.BonusOpportunities++
				estimatedPoints += 5
			}
			if m.isLateNightShift(booking) {
				summary.BonusOpportunities++
				estimatedPoints += 3
			}
		}

		// Completion points (15 + 5 for report)
		reports, err := m.getReportsForBooking(ctx, booking.BookingID)
		if err == nil && len(reports) > 0 {
			summary.CompletionPoints++
			estimatedPoints += 20 // 15 completion + 5 report

			// Level 2 bonus
			for _, report := range reports {
				if report.Severity >= 2 {
					summary.BonusOpportunities++
					estimatedPoints += 10
					break
				}
			}
		}
	}

	summary.EstimatedTotalPoints = estimatedPoints
	summary.AffectedUsers = len(userMap)

	return summary, nil
}

func (m *PointsMigrator) Execute(ctx context.Context) error {
	bookingsWithoutPoints, err := m.getBookingsWithoutPoints(ctx)
	if err != nil {
		return fmt.Errorf("failed to get bookings: %w", err)
	}

	m.logger.Info("📈 Processing bookings...", "count", len(bookingsWithoutPoints))

	processed := 0
	errors := 0

	for _, booking := range bookingsWithoutPoints {
		if err := m.processBooking(ctx, booking); err != nil {
			m.logger.Error("Failed to process booking", "booking_id", booking.BookingID, "error", err)
			errors++
		} else {
			processed++
		}

		// Progress indicator
		if processed%10 == 0 {
			m.logger.Info("Progress", "processed", processed, "total", len(bookingsWithoutPoints))
		}
	}

	m.logger.Info("🎉 Processing complete", "successful", processed, "errors", errors)

	// Recalculate all user totals
	m.logger.Info("📊 Recalculating user totals...")
	if err := m.recalculateAllUserTotals(ctx); err != nil {
		return fmt.Errorf("failed to recalculate totals: %w", err)
	}

	// Check achievements for all users
	m.logger.Info("🏆 Checking achievements...")
	if err := m.checkAllAchievements(ctx); err != nil {
		return fmt.Errorf("failed to check achievements: %w", err)
	}

	return nil
}

func (m *PointsMigrator) processBooking(ctx context.Context, booking db.Booking) error {
	// Award check-in points if user checked in
	if booking.CheckedInAt.Valid {
		if err := m.awardCheckInPoints(ctx, booking); err != nil {
			return fmt.Errorf("failed to award check-in points: %w", err)
		}
	}

	// Award completion points if there are reports
	reports, err := m.getReportsForBooking(ctx, booking.BookingID)
	if err != nil {
		return fmt.Errorf("failed to get reports: %w", err)
	}

	if len(reports) > 0 {
		maxSeverity := 0
		for _, report := range reports {
			if report.Severity > maxSeverity {
				maxSeverity = report.Severity
			}
		}

		if err := m.awardCompletionPoints(ctx, booking.UserID, booking.BookingID, maxSeverity, booking.ShiftStart); err != nil {
			return fmt.Errorf("failed to award completion points: %w", err)
		}
	}

	return nil
}

func (m *PointsMigrator) awardCheckInPoints(ctx context.Context, booking db.Booking) error {
	basePoints := 10
	userID := booking.UserID
	bookingID := booking.BookingID

	// Award base check-in points
	if err := m.awardPointsWithHistory(ctx, userID, &bookingID, basePoints, "shift_checkin", booking.CheckedInAt.Time); err != nil {
		return err
	}

	// Award bonuses
	if m.isEarlyCheckIn(booking) {
		if err := m.awardPointsWithHistory(ctx, userID, &bookingID, 3, "early_checkin", booking.CheckedInAt.Time); err != nil {
			return err
		}
	}

	if m.isWeekendShift(booking) {
		if err := m.awardPointsWithHistory(ctx, userID, &bookingID, 5, "weekend_bonus", booking.CheckedInAt.Time); err != nil {
			return err
		}
	}

	if m.isLateNightShift(booking) {
		if err := m.awardPointsWithHistory(ctx, userID, &bookingID, 3, "late_night_bonus", booking.CheckedInAt.Time); err != nil {
			return err
		}
	}

	return nil
}

func (m *PointsMigrator) awardCompletionPoints(ctx context.Context, userID, bookingID int64, maxSeverity int, shiftStart time.Time) error {
	// Award completion points (15)
	awardTime := shiftStart.Add(2 * time.Hour) // Assume completed 2 hours after start
	if err := m.awardPointsWithHistory(ctx, userID, &bookingID, 15, "shift_completion", awardTime); err != nil {
		return err
	}

	// Award report points (5 + potential level 2 bonus)
	reportPoints := 5
	reason := "report_filed"
	if maxSeverity >= 2 {
		reportPoints += 10
		reason = "level2_report"
	}

	if err := m.awardPointsWithHistory(ctx, userID, &bookingID, reportPoints, reason, awardTime.Add(5*time.Minute)); err != nil {
		return err
	}

	return nil
}

func (m *PointsMigrator) awardPointsWithHistory(ctx context.Context, userID int64, bookingID *int64, points int, reason string, awardedAt time.Time) error {
	if m.dryRun {
		m.logger.Info("DRY-RUN: Would award points", 
			"user_id", userID, 
			"booking_id", bookingID, 
			"points", points, 
			"reason", reason,
			"awarded_at", awardedAt.Format("2006-01-02 15:04:05"))
		return nil
	}

	// Insert with historical timestamp
	query := `INSERT INTO points_history (user_id, booking_id, points_awarded, reason, multiplier, created_at) 
			  VALUES (?, ?, ?, ?, 1.0, ?)`
	
	_, err := m.dbConn.ExecContext(ctx, query, userID, bookingID, points, reason, awardedAt)
	if err != nil {
		return fmt.Errorf("failed to insert points history: %w", err)
	}
	
	return nil
}

// Helper methods (same logic as points service)
func (m *PointsMigrator) isEarlyCheckIn(booking db.Booking) bool {
	if !booking.CheckedInAt.Valid {
		return false
	}
	return booking.ShiftStart.Sub(booking.CheckedInAt.Time) > 15*time.Minute
}

func (m *PointsMigrator) isWeekendShift(booking db.Booking) bool {
	weekday := booking.ShiftStart.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

func (m *PointsMigrator) isLateNightShift(booking db.Booking) bool {
	hour := booking.ShiftStart.Hour()
	return hour >= 22 || hour <= 5
}

func (m *PointsMigrator) getBookingsWithoutPoints(ctx context.Context) ([]db.Booking, error) {
	// Query for bookings that don't have corresponding points_history entries
	query := `
		SELECT b.booking_id, b.user_id, b.schedule_id, b.shift_start, b.shift_end, 
		       b.buddy_user_id, b.buddy_name, b.checked_in_at, b.created_at
		FROM bookings b 
		LEFT JOIN points_history ph ON b.booking_id = ph.booking_id 
		WHERE ph.booking_id IS NULL
		ORDER BY b.shift_start ASC`
	
	rows, err := m.dbConn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query bookings without points: %w", err)
	}
	defer rows.Close()
	
	var bookings []db.Booking
	for rows.Next() {
		var b db.Booking
		
		err := rows.Scan(
			&b.BookingID,
			&b.UserID,
			&b.ScheduleID,
			&b.ShiftStart,
			&b.ShiftEnd,
			&b.BuddyUserID,
			&b.BuddyName,
			&b.CheckedInAt,
			&b.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan booking row: %w", err)
		}
		
		bookings = append(bookings, b)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating booking rows: %w", err)
	}
	
	return bookings, nil
}

func (m *PointsMigrator) getReportsForBooking(ctx context.Context, bookingID int64) ([]struct{ Severity int }, error) {
	// Query for reports associated with this booking
	query := `SELECT severity FROM reports WHERE booking_id = ?`
	
	rows, err := m.dbConn.QueryContext(ctx, query, bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to query reports for booking: %w", err)
	}
	defer rows.Close()
	
	var reports []struct{ Severity int }
	for rows.Next() {
		var severity int
		if err := rows.Scan(&severity); err != nil {
			return nil, fmt.Errorf("failed to scan report severity: %w", err)
		}
		reports = append(reports, struct{ Severity int }{Severity: severity})
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating report rows: %w", err)
	}
	
	return reports, nil
}

func (m *PointsMigrator) recalculateAllUserTotals(ctx context.Context) error {
	if m.dryRun {
		m.logger.Info("DRY-RUN: Would recalculate user totals")
		return nil
	}
	
	// Update all user total points
	query := `
		UPDATE users 
		SET total_points = (
			SELECT COALESCE(SUM(points_awarded * multiplier), 0) 
			FROM points_history 
			WHERE points_history.user_id = users.user_id
		)`
	
	_, err := m.dbConn.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to recalculate user totals: %w", err)
	}
	
	m.logger.Info("Successfully recalculated user totals")
	return nil
}

func (m *PointsMigrator) checkAllAchievements(ctx context.Context) error {
	if m.dryRun {
		m.logger.Info("DRY-RUN: Would check achievements for all users")
		return nil
	}
	
	// Get all users and check achievements
	usersQuery := `SELECT user_id FROM users`
	rows, err := m.dbConn.QueryContext(ctx, usersQuery)
	if err != nil {
		return fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()
	
	var userIDs []int64
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return fmt.Errorf("failed to scan user ID: %w", err)
		}
		userIDs = append(userIDs, userID)
	}
	
	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating user rows: %w", err)
	}
	
	// Check achievements for each user using the points service
	for _, userID := range userIDs {
		if err := m.pointsService.CheckAndAwardAchievements(ctx, userID); err != nil {
			m.logger.Error("Failed to check achievements for user", "user_id", userID, "error", err)
			// Continue with other users rather than failing completely
		}
	}
	
	m.logger.Info("Successfully checked achievements for all users", "user_count", len(userIDs))
	return nil
}

func (m *PointsMigrator) PrintSummary(summary *MigrationSummary) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 HISTORICAL POINTS MIGRATION SUMMARY")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("📋 Total historical bookings: %d\n", summary.TotalBookings)
	fmt.Printf("🎯 Bookings needing points: %d\n", summary.BookingsNeedingPoints)
	fmt.Printf("👥 Users affected: %d\n", summary.AffectedUsers)
	fmt.Println("")
	fmt.Printf("✅ Check-in points to award: %d bookings × 10 pts = %d pts\n", 
		summary.CheckInPoints, summary.CheckInPoints*10)
	fmt.Printf("🎯 Completion points to award: %d bookings × ~20 pts = %d pts\n", 
		summary.CompletionPoints, summary.CompletionPoints*20)
	fmt.Printf("🎁 Bonus opportunities: %d\n", summary.BonusOpportunities)
	fmt.Println("")
	fmt.Printf("💰 ESTIMATED TOTAL POINTS: %d\n", summary.EstimatedTotalPoints)
	fmt.Println(strings.Repeat("=", 60))
	
	if summary.BookingsNeedingPoints == 0 {
		fmt.Println("✨ All historical bookings already have points awarded!")
		fmt.Println("   No migration needed.")
	}
}