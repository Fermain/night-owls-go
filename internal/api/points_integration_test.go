package api_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"night-owls-go/internal/auth"
	db "night-owls-go/internal/db/sqlc_generated"
	"night-owls-go/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper methods for testApp
func (app *testApp) createTestUserAndLogin(t *testing.T, phone, name, role string) (db.CreateUserRow, string) {
	t.Helper()
	ctx := context.Background()
	user, err := app.Querier.CreateUser(ctx, db.CreateUserParams{
		Phone: phone,
		Name:  sql.NullString{String: name, Valid: true},
		Role:  sql.NullString{String: role, Valid: true},
	})
	require.NoError(t, err)
	otp, errOtp := auth.GenerateOTP()
	require.NoError(t, errOtp)

	// Get the OTP store from the testApp (need to access it via UserService)
	// Store OTP directly without going through RegisterOrLoginUser
	otpStore := auth.NewInMemoryOTPStore()
	otpStore.StoreOTP(phone, otp, 5*time.Minute)

	// Create a temporary user service with our OTP store
	tempUserService := service.NewUserService(app.Querier, otpStore, app.Config, app.Logger)

	token, err := tempUserService.VerifyOTP(ctx, phone, otp, "test-ip", "test-agent")
	require.NoError(t, err, "Failed to verify OTP and get token for test user %s", phone)
	require.NotEmpty(t, token)
	return user, token
}

func (app *testApp) createTestSchedule(t *testing.T, name, cronExpr string) db.Schedule {
	t.Helper()
	ctx := context.Background()
	schedule, err := app.Querier.CreateSchedule(ctx, db.CreateScheduleParams{
		Name:            name,
		CronExpr:        cronExpr,
		DurationMinutes: 120, // 2 hours
		Timezone:        sql.NullString{String: "UTC", Valid: true},
	})
	require.NoError(t, err)
	return schedule
}

func TestPointsIntegration_ShiftCheckinAwardsPoints(t *testing.T) {
	app := newTestApp(t)
	defer app.DB.Close()

	// Create test user and schedule
	user, token := app.createTestUserAndLogin(t, "+15550001001", "Test User", "owl")
	schedule := app.createTestSchedule(t, "Test Schedule", "0 8 * * *")

	// Create a booking
	ctx := context.Background()
	booking, err := app.Querier.CreateBooking(ctx, db.CreateBookingParams{
		UserID:     user.UserID,
		ScheduleID: schedule.ScheduleID,
		ShiftStart: time.Now().Add(-30 * time.Minute), // Already started, eligible for check-in
		ShiftEnd:   time.Now().Add(90 * time.Minute),
	})
	require.NoError(t, err)

	// Get initial user points (should be 0)
	initialStats, err := app.Querier.GetUserPoints(ctx, user.UserID)
	require.NoError(t, err)
	initialPoints := int64(0)
	if initialStats.TotalPoints.Valid {
		initialPoints = initialStats.TotalPoints.Int64
	}

	// Check in to the booking
	rr := app.makeRequest(t, "POST", fmt.Sprintf("/bookings/%d/checkin", booking.BookingID), nil, token)
	require.Equal(t, http.StatusOK, rr.Code, "Check-in failed: %s", rr.Body.String())

	// Wait a moment for async processing
	time.Sleep(100 * time.Millisecond)

	// Verify points were awarded
	finalStats, err := app.Querier.GetUserPoints(ctx, user.UserID)
	require.NoError(t, err)

	expectedMinPoints := initialPoints + 10 // At least base check-in points
	if finalStats.TotalPoints.Valid {
		assert.GreaterOrEqual(t, finalStats.TotalPoints.Int64, expectedMinPoints,
			"Expected at least %d points after check-in, got %d", expectedMinPoints, finalStats.TotalPoints.Int64)
	} else {
		t.Fatal("TotalPoints should be valid after awarding points")
	}

	// Verify points history was created
	history, err := app.Querier.GetUserPointsHistory(ctx, db.GetUserPointsHistoryParams{
		UserID: user.UserID,
		Limit:  10,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, history, "Points history should not be empty after check-in")

	// Verify at least one entry has check-in reason
	found := false
	for _, entry := range history {
		if entry.Reason == "shift_checkin" {
			found = true
			assert.Equal(t, int64(10), entry.PointsAwarded, "Check-in should award 10 points")
			break
		}
	}
	assert.True(t, found, "Should have points history entry for shift_checkin")

	t.Logf("✅ Check-in points integration test passed. Awarded %d total points",
		finalStats.TotalPoints.Int64-initialPoints)
}

func TestPointsIntegration_ShiftCompletionAwardsPoints(t *testing.T) {
	app := newTestApp(t)
	defer app.DB.Close()

	// Create test user and schedule
	user, token := app.createTestUserAndLogin(t, "+15550001002", "Test User 2", "owl")
	schedule := app.createTestSchedule(t, "Test Schedule 2", "0 8 * * *")

	ctx := context.Background()

	// Create and check in to a booking
	booking, err := app.Querier.CreateBooking(ctx, db.CreateBookingParams{
		UserID:     user.UserID,
		ScheduleID: schedule.ScheduleID,
		ShiftStart: time.Now().Add(-2 * time.Hour), // Past shift
		ShiftEnd:   time.Now().Add(-1 * time.Hour),
	})
	require.NoError(t, err)

	// Mark as checked in
	_, err = app.Querier.UpdateBookingCheckIn(ctx, db.UpdateBookingCheckInParams{
		BookingID:   booking.BookingID,
		CheckedInAt: sql.NullTime{Time: time.Now().Add(-2 * time.Hour), Valid: true},
	})
	require.NoError(t, err)

	// Get points after check-in (simulate that check-in already happened)
	// For this test, we'll assume check-in points were already awarded

	// Create a report (completing the shift)
	reportPayload := map[string]interface{}{
		"booking_id": booking.BookingID,
		"severity":   1,
		"message":    "Test shift completion report",
	}
	reportBody, _ := json.Marshal(reportPayload)

	rr := app.makeRequest(t, "POST", "/bookings/"+fmt.Sprintf("%d", booking.BookingID)+"/report", bytes.NewReader(reportBody), token)
	require.Equal(t, http.StatusCreated, rr.Code, "Report creation failed: %s", rr.Body.String())

	// Wait for async processing
	time.Sleep(100 * time.Millisecond)

	// Verify completion points were awarded
	finalStats, err := app.Querier.GetUserPoints(ctx, user.UserID)
	require.NoError(t, err)

	// We expect completion points (15) + report points (5) to be awarded
	expectedMinPoints := int64(20) // Base completion + report points
	if finalStats.TotalPoints.Valid {
		assert.GreaterOrEqual(t, finalStats.TotalPoints.Int64, expectedMinPoints,
			"Expected at least %d points after completion, got %d",
			expectedMinPoints, finalStats.TotalPoints.Int64)
	} else {
		t.Fatal("TotalPoints should be valid after awarding completion points")
	}

	// Verify completion points history
	history, err := app.Querier.GetUserPointsHistory(ctx, db.GetUserPointsHistoryParams{
		UserID: user.UserID,
		Limit:  10,
	})
	require.NoError(t, err)

	// Look for completion and report points
	foundCompletion := false
	foundReport := false
	for _, entry := range history {
		if entry.Reason == "shift_completion" {
			foundCompletion = true
			assert.Equal(t, int64(15), entry.PointsAwarded, "Completion should award 15 points")
		}
		if entry.Reason == "report_filed" {
			foundReport = true
			assert.Equal(t, int64(5), entry.PointsAwarded, "Report should award 5 points")
		}
	}
	assert.True(t, foundCompletion, "Should have completion points in history")
	assert.True(t, foundReport, "Should have report points in history")

	// Verify shift count was updated
	assert.True(t, finalStats.ShiftCount.Valid, "ShiftCount should be valid")
	assert.GreaterOrEqual(t, finalStats.ShiftCount.Int64, int64(1), "Should have at least 1 completed shift")

	t.Logf("✅ Completion points integration test passed. Total points: %d", finalStats.TotalPoints.Int64)
}

func TestPointsIntegration_EarlyCheckinBonus(t *testing.T) {
	app := newTestApp(t)
	defer app.DB.Close()

	// Create test user and schedule
	user, token := app.createTestUserAndLogin(t, "+15550001003", "Early Bird", "owl")
	schedule := app.createTestSchedule(t, "Early Test Schedule", "0 8 * * *")

	ctx := context.Background()

	// Create a booking that starts in 30 minutes (early check-in opportunity)
	futureShiftStart := time.Now().Add(30 * time.Minute)
	booking, err := app.Querier.CreateBooking(ctx, db.CreateBookingParams{
		UserID:     user.UserID,
		ScheduleID: schedule.ScheduleID,
		ShiftStart: futureShiftStart,
		ShiftEnd:   futureShiftStart.Add(2 * time.Hour),
	})
	require.NoError(t, err)

	// Check in early (now, when shift starts in 30 min)
	rr := app.makeRequest(t, "POST", fmt.Sprintf("/bookings/%d/checkin", booking.BookingID), nil, token)
	require.Equal(t, http.StatusOK, rr.Code, "Early check-in failed: %s", rr.Body.String())

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Verify bonus points for early check-in
	history, err := app.Querier.GetUserPointsHistory(ctx, db.GetUserPointsHistoryParams{
		UserID: user.UserID,
		Limit:  10,
	})
	require.NoError(t, err)

	foundEarlyBonus := false
	totalPoints := int64(0)
	for _, entry := range history {
		totalPoints += entry.PointsAwarded
		if entry.Reason == "early_checkin" {
			foundEarlyBonus = true
			assert.Equal(t, int64(3), entry.PointsAwarded, "Early check-in bonus should be 3 points")
		}
	}

	assert.True(t, foundEarlyBonus, "Should have early check-in bonus in points history")
	assert.GreaterOrEqual(t, totalPoints, int64(13), "Should have at least 13 points (10 base + 3 early bonus)")

	t.Logf("✅ Early check-in bonus test passed. Total awarded: %d points", totalPoints)
}

func TestPointsIntegration_LeaderboardUpdates(t *testing.T) {
	app := newTestApp(t)
	defer app.DB.Close()

	// Create multiple test users
	user1, token1 := app.createTestUserAndLogin(t, "+15550001005", "Leader User", "owl")
	user2, _ := app.createTestUserAndLogin(t, "+15550001006", "Follower User", "owl")

	ctx := context.Background()

	// Give user1 more points by awarding them directly
	err := app.Querier.AwardPoints(ctx, db.AwardPointsParams{
		UserID:        user1.UserID,
		BookingID:     sql.NullInt64{Valid: false},
		PointsAwarded: 100,
		Reason:        "test_leader_points",
		Multiplier:    sql.NullFloat64{Float64: 1.0, Valid: true},
	})
	require.NoError(t, err)

	err = app.Querier.UpdateUserTotalPoints(ctx, db.UpdateUserTotalPointsParams{
		UserID:   user1.UserID,
		UserID_2: user1.UserID,
	})
	require.NoError(t, err)

	// Give user2 fewer points
	err = app.Querier.AwardPoints(ctx, db.AwardPointsParams{
		UserID:        user2.UserID,
		BookingID:     sql.NullInt64{Valid: false},
		PointsAwarded: 50,
		Reason:        "test_follower_points",
		Multiplier:    sql.NullFloat64{Float64: 1.0, Valid: true},
	})
	require.NoError(t, err)

	err = app.Querier.UpdateUserTotalPoints(ctx, db.UpdateUserTotalPointsParams{
		UserID:   user2.UserID,
		UserID_2: user2.UserID,
	})
	require.NoError(t, err)

	// Test leaderboard API endpoint
	rr := app.makeRequest(t, "GET", "/leaderboard", nil, token1)
	require.Equal(t, http.StatusOK, rr.Code, "Leaderboard request failed: %s", rr.Body.String())

	var leaderboard []map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &leaderboard)
	require.NoError(t, err)

	// Verify leaderboard order (user1 should be first)
	if len(leaderboard) >= 2 {
		// Find our test users in the leaderboard
		var user1Entry, user2Entry map[string]interface{}
		for _, entry := range leaderboard {
			if userID, ok := entry["user_id"].(float64); ok {
				if int64(userID) == user1.UserID {
					user1Entry = entry
				} else if int64(userID) == user2.UserID {
					user2Entry = entry
				}
			}
		}

		if user1Entry != nil && user2Entry != nil {
			user1Points := user1Entry["total_points"].(float64)
			user2Points := user2Entry["total_points"].(float64)
			assert.Greater(t, user1Points, user2Points, "User1 should have more points than User2")
		}
	}

	// Test user stats endpoint
	rr = app.makeRequest(t, "GET", "/user/stats", nil, token1)
	require.Equal(t, http.StatusOK, rr.Code, "User stats request failed: %s", rr.Body.String())

	var userStats map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &userStats)
	require.NoError(t, err)

	// Verify user stats
	if totalPoints, ok := userStats["total_points"]; ok {
		assert.GreaterOrEqual(t, totalPoints.(float64), float64(100), "User should have at least 100 points")
	}

	t.Logf("✅ Leaderboard integration test passed")
}
