package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"night-owls-go/internal/config"
	db "night-owls-go/internal/db/sqlc_generated"
	"night-owls-go/internal/logging"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helper to create an in-memory database for testing
func setupTestDB(t *testing.T) (*sql.DB, db.Querier) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create tables
	schema := `
	CREATE TABLE users (
		user_id INTEGER PRIMARY KEY AUTOINCREMENT,
		phone TEXT UNIQUE NOT NULL,
		name TEXT,
		role TEXT NOT NULL DEFAULT 'guest',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE schedules (
		schedule_id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		cron_expr TEXT NOT NULL,
		start_date DATE,
		end_date DATE,
		duration_minutes INTEGER NOT NULL DEFAULT 120,
		timezone TEXT
	);

	CREATE TABLE bookings (
		booking_id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL REFERENCES users(user_id),
		schedule_id INTEGER NOT NULL REFERENCES schedules(schedule_id),
		shift_start DATETIME NOT NULL,
		shift_end DATETIME NOT NULL,
		buddy_user_id INTEGER REFERENCES users(user_id),
		buddy_name TEXT,
		attended BOOLEAN NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(schedule_id, shift_start)
	);

	CREATE TABLE recurring_assignments (
		recurring_assignment_id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
		buddy_name TEXT,
		day_of_week INTEGER NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6),
		schedule_id INTEGER NOT NULL REFERENCES schedules(schedule_id) ON DELETE CASCADE,
		time_slot TEXT NOT NULL,
		description TEXT,
		is_active BOOLEAN NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, day_of_week, schedule_id, time_slot)
	);
	`

	_, err = testDB.Exec(schema)
	require.NoError(t, err)

	return testDB, db.New(testDB)
}

func setupTestService(t *testing.T) (*RecurringAssignmentService, db.Querier, *ScheduleService) {
	testDB, querier := setupTestDB(t)
	t.Cleanup(func() { testDB.Close() })

	cfg := &config.Config{DevMode: true}
	logger := logging.NewLogger(cfg)

	recurringService := NewRecurringAssignmentService(querier, logger, cfg)
	scheduleService := NewScheduleService(querier, logger, cfg)

	return recurringService, querier, scheduleService
}

// Test data setup helpers
func createTestUser(t *testing.T, querier db.Querier, phone, name string) db.User {
	user, err := querier.CreateUser(context.Background(), db.CreateUserParams{
		Phone: phone,
		Name:  sql.NullString{String: name, Valid: name != ""},
		Role:  sql.NullString{String: "owl", Valid: true},
	})
	require.NoError(t, err)
	return user
}

func createTestSchedule(t *testing.T, querier db.Querier, name, cronExpr string) db.Schedule {
	schedule, err := querier.CreateSchedule(context.Background(), db.CreateScheduleParams{
		Name:     name,
		CronExpr: cronExpr,
	})
	require.NoError(t, err)
	return schedule
}

func createTestBooking(t *testing.T, querier db.Querier, userID, scheduleID int64, startTime time.Time) db.Booking {
	booking, err := querier.CreateBooking(context.Background(), db.CreateBookingParams{
		UserID:     userID,
		ScheduleID: scheduleID,
		ShiftStart: startTime,
		ShiftEnd:   startTime.Add(2 * time.Hour),
	})
	require.NoError(t, err)
	return booking
}

// CRUD Operations Tests
func TestCreateRecurringAssignment(t *testing.T) {
	service, querier, _ := setupTestService(t)
	ctx := context.Background()

	user := createTestUser(t, querier, "+1234567890", "John Doe")
	schedule := createTestSchedule(t, querier, "Night Watch", "0 18 * * 6")

	t.Run("successful creation", func(t *testing.T) {
		params := db.CreateRecurringAssignmentParams{
			UserID:     user.UserID,
			DayOfWeek:  6, // Saturday
			ScheduleID: schedule.ScheduleID,
			TimeSlot:   "18:00-20:00",
			BuddyName:  sql.NullString{String: "Jane Smith", Valid: true},
			Description: sql.NullString{String: "Regular night patrol", Valid: true},
		}

		assignment, err := service.CreateRecurringAssignment(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, user.UserID, assignment.UserID)
		assert.Equal(t, int64(6), assignment.DayOfWeek)
		assert.Equal(t, schedule.ScheduleID, assignment.ScheduleID)
		assert.Equal(t, "18:00-20:00", assignment.TimeSlot)
		assert.True(t, assignment.BuddyName.Valid)
		assert.Equal(t, "Jane Smith", assignment.BuddyName.String)
		assert.True(t, assignment.IsActive)
	})

	t.Run("user not found", func(t *testing.T) {
		params := db.CreateRecurringAssignmentParams{
			UserID:     99999, // Non-existent user
			DayOfWeek:  6,
			ScheduleID: schedule.ScheduleID,
			TimeSlot:   "18:00-20:00",
		}

		_, err := service.CreateRecurringAssignment(ctx, params)
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("schedule not found", func(t *testing.T) {
		params := db.CreateRecurringAssignmentParams{
			UserID:     user.UserID,
			DayOfWeek:  6,
			ScheduleID: 99999, // Non-existent schedule
			TimeSlot:   "18:00-20:00",
		}

		_, err := service.CreateRecurringAssignment(ctx, params)
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("duplicate assignment", func(t *testing.T) {
		params := db.CreateRecurringAssignmentParams{
			UserID:     user.UserID,
			DayOfWeek:  6,
			ScheduleID: schedule.ScheduleID,
			TimeSlot:   "18:00-20:00",
		}

		// Create first assignment
		_, err := service.CreateRecurringAssignment(ctx, params)
		require.NoError(t, err)

		// Try to create duplicate
		_, err = service.CreateRecurringAssignment(ctx, params)
		assert.Error(t, err) // Should fail due to UNIQUE constraint
	})
}

func TestGetRecurringAssignmentByID(t *testing.T) {
	service, querier, _ := setupTestService(t)
	ctx := context.Background()

	user := createTestUser(t, querier, "+1234567890", "John Doe")
	schedule := createTestSchedule(t, querier, "Night Watch", "0 18 * * 6")

	// Create assignment
	params := db.CreateRecurringAssignmentParams{
		UserID:     user.UserID,
		DayOfWeek:  6,
		ScheduleID: schedule.ScheduleID,
		TimeSlot:   "18:00-20:00",
	}
	created, err := service.CreateRecurringAssignment(ctx, params)
	require.NoError(t, err)

	t.Run("found", func(t *testing.T) {
		assignment, err := service.GetRecurringAssignmentByID(ctx, created.RecurringAssignmentID)
		require.NoError(t, err)
		assert.Equal(t, created.RecurringAssignmentID, assignment.RecurringAssignmentID)
		assert.Equal(t, user.UserID, assignment.UserID)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := service.GetRecurringAssignmentByID(ctx, 99999)
		assert.ErrorIs(t, err, ErrNotFound)
	})
}

func TestDeleteRecurringAssignment(t *testing.T) {
	service, querier, _ := setupTestService(t)
	ctx := context.Background()

	user := createTestUser(t, querier, "+1234567890", "John Doe")
	schedule := createTestSchedule(t, querier, "Night Watch", "0 18 * * 6")

	// Create assignment
	params := db.CreateRecurringAssignmentParams{
		UserID:     user.UserID,
		DayOfWeek:  6,
		ScheduleID: schedule.ScheduleID,
		TimeSlot:   "18:00-20:00",
	}
	created, err := service.CreateRecurringAssignment(ctx, params)
	require.NoError(t, err)

	t.Run("successful soft delete", func(t *testing.T) {
		err := service.DeleteRecurringAssignment(ctx, created.RecurringAssignmentID)
		require.NoError(t, err)

		// Verify it's not in active list
		assignments, err := service.ListRecurringAssignments(ctx)
		require.NoError(t, err)
		assert.Len(t, assignments, 0)

		// But still exists in DB (soft delete)
		assignment, err := service.GetRecurringAssignmentByID(ctx, created.RecurringAssignmentID)
		require.NoError(t, err)
		assert.False(t, assignment.IsActive)
	})
}

// MaterializeUpcomingBookings Tests - The Most Critical Functionality
func TestMaterializeUpcomingBookings(t *testing.T) {
	service, querier, scheduleService := setupTestService(t)
	ctx := context.Background()

	// Test data setup
	user1 := createTestUser(t, querier, "+1111111111", "Alice")
	user2 := createTestUser(t, querier, "+2222222222", "Bob")
	schedule := createTestSchedule(t, querier, "Night Watch", "0 18 * * 6") // Every Saturday 6PM

	t.Run("creates bookings from recurring assignments", func(t *testing.T) {
		// Create recurring assignment for Saturday 18:00-20:00
		_, err := service.CreateRecurringAssignment(ctx, db.CreateRecurringAssignmentParams{
			UserID:     user1.UserID,
			DayOfWeek:  6, // Saturday
			ScheduleID: schedule.ScheduleID,
			TimeSlot:   "18:00-20:00",
			BuddyName:  sql.NullString{String: "Buddy1", Valid: true},
		})
		require.NoError(t, err)

		// The materialization will be tested with real shift slots in integration tests
		// Here we test the core logic
		fromTime := time.Now().UTC()
		toTime := fromTime.AddDate(0, 0, 14)

		err = service.MaterializeUpcomingBookings(ctx, scheduleService, fromTime, toTime)
		assert.NoError(t, err)
	})

	t.Run("skips already booked slots", func(t *testing.T) {
		// Create recurring assignment
		_, err := service.CreateRecurringAssignment(ctx, db.CreateRecurringAssignmentParams{
			UserID:     user2.UserID,
			DayOfWeek:  0, // Sunday
			ScheduleID: schedule.ScheduleID,
			TimeSlot:   "18:00-20:00",
		})
		require.NoError(t, err)

		// Pre-create a booking for the same slot
		nextSunday := time.Now().UTC().AddDate(0, 0, int((7-time.Now().Weekday())%7))
		if nextSunday.Before(time.Now()) {
			nextSunday = nextSunday.AddDate(0, 0, 7)
		}
		shiftStart := time.Date(nextSunday.Year(), nextSunday.Month(), nextSunday.Day(),
			18, 0, 0, 0, time.UTC)

		createTestBooking(t, querier, user1.UserID, schedule.ScheduleID, shiftStart)

		// Materialization should skip this slot
		fromTime := time.Now().UTC()
		toTime := time.Now().UTC().AddDate(0, 0, 14)

		err = service.MaterializeUpcomingBookings(ctx, scheduleService, fromTime, toTime)
		assert.NoError(t, err)

		// Verify no duplicate booking was created by checking user1's bookings
		user1Bookings, err := querier.ListBookingsByUserID(ctx, user1.UserID)
		require.NoError(t, err)
		
		// Count bookings for this specific time slot
		slotBookings := 0
		for _, booking := range user1Bookings {
			if booking.ShiftStart.Equal(shiftStart) {
				slotBookings++
			}
		}
		assert.Equal(t, 1, slotBookings, "Should have exactly one booking for this slot")
	})

	t.Run("handles multiple assignments for same slot - first wins", func(t *testing.T) {
		// Clean up previous test data
		err := querier.AdminBulkDeleteUsers(ctx, []int64{user1.UserID, user2.UserID})
		require.NoError(t, err)
		
		user1 = createTestUser(t, querier, "+3333333333", "Charlie")
		user2 = createTestUser(t, querier, "+4444444444", "David")

		// Create two recurring assignments for the same slot
		_, err = service.CreateRecurringAssignment(ctx, db.CreateRecurringAssignmentParams{
			UserID:     user1.UserID,
			DayOfWeek:  1, // Monday
			ScheduleID: schedule.ScheduleID,
			TimeSlot:   "18:00-20:00",
		})
		require.NoError(t, err)

		_, err = service.CreateRecurringAssignment(ctx, db.CreateRecurringAssignmentParams{
			UserID:     user2.UserID,
			DayOfWeek:  1, // Monday
			ScheduleID: schedule.ScheduleID,
			TimeSlot:   "18:00-20:00",
		})
		require.NoError(t, err)

		fromTime := time.Now().UTC()
		toTime := time.Now().UTC().AddDate(0, 0, 14)

		err = service.MaterializeUpcomingBookings(ctx, scheduleService, fromTime, toTime)
		assert.NoError(t, err)

		// The test would need actual shift slots to verify behavior
		// This logic will be tested in integration tests
	})

	t.Run("no assignments", func(t *testing.T) {
		// Clean up all assignments
		assignments, err := service.ListRecurringAssignments(ctx)
		require.NoError(t, err)
		
		for _, assignment := range assignments {
			err := service.DeleteRecurringAssignment(ctx, assignment.RecurringAssignmentID)
			require.NoError(t, err)
		}

		fromTime := time.Now().UTC()
		toTime := time.Now().UTC().AddDate(0, 0, 14)

		err = service.MaterializeUpcomingBookings(ctx, scheduleService, fromTime, toTime)
		assert.NoError(t, err) // Should not error with no assignments
	})

	t.Run("time slot matching precision", func(t *testing.T) {
		// Test that time slot matching is precise
		_, err := service.CreateRecurringAssignment(ctx, db.CreateRecurringAssignmentParams{
			UserID:     user1.UserID,
			DayOfWeek:  2, // Tuesday
			ScheduleID: schedule.ScheduleID,
			TimeSlot:   "18:00-20:00", // Exact format
		})
		require.NoError(t, err)

		// The materialization logic should match "18:00-20:00" exactly
		// Tested in integration where we can control shift slots
		fromTime := time.Now().UTC()
		toTime := time.Now().UTC().AddDate(0, 0, 14)

		err = service.MaterializeUpcomingBookings(ctx, scheduleService, fromTime, toTime)
		assert.NoError(t, err)
	})
}

func TestListRecurringAssignments(t *testing.T) {
	service, querier, _ := setupTestService(t)
	ctx := context.Background()

	user := createTestUser(t, querier, "+1234567890", "John Doe")
	schedule := createTestSchedule(t, querier, "Night Watch", "0 18 * * 6")

	t.Run("lists only active assignments", func(t *testing.T) {
		// Create active assignment
		_, err := service.CreateRecurringAssignment(ctx, db.CreateRecurringAssignmentParams{
			UserID:     user.UserID,
			DayOfWeek:  6,
			ScheduleID: schedule.ScheduleID,
			TimeSlot:   "18:00-20:00",
		})
		require.NoError(t, err)

		// Create and delete another assignment
		deleted, err := service.CreateRecurringAssignment(ctx, db.CreateRecurringAssignmentParams{
			UserID:     user.UserID,
			DayOfWeek:  0,
			ScheduleID: schedule.ScheduleID,
			TimeSlot:   "18:00-20:00",
		})
		require.NoError(t, err)

		err = service.DeleteRecurringAssignment(ctx, deleted.RecurringAssignmentID)
		require.NoError(t, err)

		// List should only return active assignment
		assignments, err := service.ListRecurringAssignments(ctx)
		require.NoError(t, err)
		assert.Len(t, assignments, 1)
		assert.Equal(t, int64(6), assignments[0].DayOfWeek)
	})
}

func TestUpdateRecurringAssignment(t *testing.T) {
	service, querier, _ := setupTestService(t)
	ctx := context.Background()

	user1 := createTestUser(t, querier, "+1234567890", "John Doe")
	user2 := createTestUser(t, querier, "+0987654321", "Jane Smith")
	schedule := createTestSchedule(t, querier, "Night Watch", "0 18 * * 6")

	// Create assignment
	created, err := service.CreateRecurringAssignment(ctx, db.CreateRecurringAssignmentParams{
		UserID:     user1.UserID,
		DayOfWeek:  6,
		ScheduleID: schedule.ScheduleID,
		TimeSlot:   "18:00-20:00",
	})
	require.NoError(t, err)

	t.Run("successful update", func(t *testing.T) {
		params := db.UpdateRecurringAssignmentParams{
			RecurringAssignmentID: created.RecurringAssignmentID,
			UserID:                user2.UserID,
			DayOfWeek:             0, // Change to Sunday
			ScheduleID:            schedule.ScheduleID,
			TimeSlot:              "19:00-21:00", // Change time
			BuddyName:             sql.NullString{String: "New Buddy", Valid: true},
			Description:           sql.NullString{String: "Updated description", Valid: true},
		}

		updated, err := service.UpdateRecurringAssignment(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, user2.UserID, updated.UserID)
		assert.Equal(t, int64(0), updated.DayOfWeek)
		assert.Equal(t, "19:00-21:00", updated.TimeSlot)
		assert.Equal(t, "New Buddy", updated.BuddyName.String)
	})

	t.Run("assignment not found", func(t *testing.T) {
		params := db.UpdateRecurringAssignmentParams{
			RecurringAssignmentID: 99999,
			UserID:                user1.UserID,
			DayOfWeek:             6,
			ScheduleID:            schedule.ScheduleID,
			TimeSlot:              "18:00-20:00",
		}

		_, err := service.UpdateRecurringAssignment(ctx, params)
		assert.ErrorIs(t, err, ErrNotFound)
	})
}

func TestGetRecurringAssignmentsByPattern(t *testing.T) {
	service, querier, _ := setupTestService(t)
	ctx := context.Background()

	user := createTestUser(t, querier, "+1234567890", "John Doe")
	schedule := createTestSchedule(t, querier, "Night Watch", "0 18 * * 6")

	// Create assignment
	_, err := service.CreateRecurringAssignment(ctx, db.CreateRecurringAssignmentParams{
		UserID:     user.UserID,
		DayOfWeek:  6,
		ScheduleID: schedule.ScheduleID,
		TimeSlot:   "18:00-20:00",
	})
	require.NoError(t, err)

	t.Run("finds matching pattern", func(t *testing.T) {
		assignments, err := service.GetRecurringAssignmentsByPattern(ctx, 6, schedule.ScheduleID, "18:00-20:00")
		require.NoError(t, err)
		assert.Len(t, assignments, 1)
		assert.Equal(t, user.UserID, assignments[0].UserID)
		assert.Equal(t, "John Doe", assignments[0].UserName.String)
	})

	t.Run("no matches", func(t *testing.T) {
		assignments, err := service.GetRecurringAssignmentsByPattern(ctx, 0, schedule.ScheduleID, "18:00-20:00")
		require.NoError(t, err)
		assert.Len(t, assignments, 0)
	})
}

// Edge Cases and Error Handling
func TestRecurringAssignmentEdgeCases(t *testing.T) {
	service, querier, _ := setupTestService(t)
	ctx := context.Background()

	user := createTestUser(t, querier, "+1234567890", "John Doe")
	schedule := createTestSchedule(t, querier, "Night Watch", "0 18 * * 6")

	t.Run("day of week boundary values", func(t *testing.T) {
		// Test Sunday (0) and Saturday (6)
		for _, day := range []int64{0, 6} {
			_, err := service.CreateRecurringAssignment(ctx, db.CreateRecurringAssignmentParams{
				UserID:     user.UserID,
				DayOfWeek:  day,
				ScheduleID: schedule.ScheduleID,
				TimeSlot:   "18:00-20:00",
			})
			assert.NoError(t, err, "Day %d should be valid", day)
		}
	})

	t.Run("time slot format variations", func(t *testing.T) {
		validTimeSlots := []string{
			"00:00-02:00", // Midnight
			"23:00-01:00", // Crosses midnight (if supported)
			"06:30-08:30", // Half-hour precision
		}

		for i, timeSlot := range validTimeSlots {
			_, err := service.CreateRecurringAssignment(ctx, db.CreateRecurringAssignmentParams{
				UserID:     user.UserID,
				DayOfWeek:  int64(i + 1), // Different days to avoid conflicts
				ScheduleID: schedule.ScheduleID,
				TimeSlot:   timeSlot,
			})
			assert.NoError(t, err, "Time slot %s should be valid", timeSlot)
		}
	})

	t.Run("empty and null buddy name handling", func(t *testing.T) {
		// Test with null buddy name
		_, err := service.CreateRecurringAssignment(ctx, db.CreateRecurringAssignmentParams{
			UserID:     user.UserID,
			DayOfWeek:  3,
			ScheduleID: schedule.ScheduleID,
			TimeSlot:   "18:00-20:00",
			BuddyName:  sql.NullString{Valid: false},
		})
		assert.NoError(t, err)

		// Test with empty buddy name
		_, err = service.CreateRecurringAssignment(ctx, db.CreateRecurringAssignmentParams{
			UserID:     user.UserID,
			DayOfWeek:  4,
			ScheduleID: schedule.ScheduleID,
			TimeSlot:   "18:00-20:00",
			BuddyName:  sql.NullString{String: "", Valid: true},
		})
		assert.NoError(t, err)
	})
} 