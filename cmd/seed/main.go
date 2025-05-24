package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"night-owls-go/internal/config"
	db "night-owls-go/internal/db/sqlc_generated"
	"night-owls-go/internal/logging"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

// SeedData represents the data to be seeded
type SeedData struct {
	Users               []UserSeed
	Schedules           []ScheduleSeed
	RecurringAssignments []RecurringAssignmentSeed
	Bookings            []BookingSeed
}

type UserSeed struct {
	Name  string
	Phone string
	Role  string
}

type ScheduleSeed struct {
	Name            string
	CronExpr        string
	StartDate       string
	EndDate         string
	DurationMinutes int64
	Timezone        string
}

type RecurringAssignmentSeed struct {
	UserPhone   string
	ScheduleName string
	DayOfWeek   int64
	TimeSlot    string
	BuddyName   string
	Description string
}

type BookingSeed struct {
	UserPhone    string
	ScheduleName string
	ShiftStart   string
	BuddyName    string
	Attended     bool
}

func main() {
	var (
		dbPath = flag.String("db", "", "Database path (if empty, uses config)")
		reset  = flag.Bool("reset", false, "Reset database before seeding")
		dryRun = flag.Bool("dry-run", false, "Show what would be seeded without actually doing it")
	)
	flag.Parse()

	// Load environment and config
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	logger := logging.NewLogger(cfg)
	slog.SetDefault(logger)

	// Determine database path
	databasePath := cfg.DatabasePath
	if *dbPath != "" {
		databasePath = *dbPath
	}

	logger.Info("Starting seeding process", "database", databasePath, "reset", *reset, "dry_run", *dryRun)

	if *dryRun {
		logger.Info("DRY RUN MODE - No actual changes will be made")
		showSeedData(logger)
		return
	}

	// Handle database reset
	if *reset {
		if err := os.Remove(databasePath); err != nil && !os.IsNotExist(err) {
			logger.Error("Failed to remove existing database", "path", databasePath, "error", err)
			os.Exit(1)
		}
		logger.Info("Database reset completed", "path", databasePath)
	}

	// Run migrations first
	if err := runMigrations(databasePath, logger); err != nil {
		logger.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Open database connection
	dbConn, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		logger.Error("Failed to open database", "path", databasePath, "error", err)
		os.Exit(1)
	}
	defer dbConn.Close()

	if err = dbConn.Ping(); err != nil {
		logger.Error("Failed to ping database", "error", err)
		os.Exit(1)
	}

	// Perform seeding
	querier := db.New(dbConn)
	if err := seedDatabase(context.Background(), querier, logger); err != nil {
		logger.Error("Seeding failed", "error", err)
		os.Exit(1)
	}

	logger.Info("Seeding completed successfully")
}

func runMigrations(databasePath string, logger *slog.Logger) error {
	migrationDSN := "sqlite3://" + databasePath
	logger.Info("Running migrations", "dsn", migrationDSN)

	m, err := migrate.New("file://internal/db/migrations", migrationDSN)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer func() {
		if srcErr, dbErr := m.Close(); srcErr != nil || dbErr != nil {
			logger.Warn("Error closing migration instance", "src_err", srcErr, "db_err", dbErr)
		}
	}()

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		logger.Info("No new migrations to apply")
	} else {
		logger.Info("Migrations applied successfully")
	}

	return nil
}

func seedDatabase(ctx context.Context, querier db.Querier, logger *slog.Logger) error {
	seedData := getSeedData()

	// Seed users first
	userMap := make(map[string]int64)
	for _, userSeed := range seedData.Users {
		user, err := querier.CreateUser(ctx, db.CreateUserParams{
			Phone: userSeed.Phone,
			Name:  sql.NullString{String: userSeed.Name, Valid: true},
			Role:  sql.NullString{String: userSeed.Role, Valid: true},
		})
		if err != nil {
			logger.Error("Failed to create user", "phone", userSeed.Phone, "error", err)
			return err
		}
		userMap[userSeed.Phone] = user.UserID
		logger.Info("Created user", "name", userSeed.Name, "phone", userSeed.Phone, "role", userSeed.Role)
	}

	// Seed additional schedules
	scheduleMap := make(map[string]int64)
	for _, scheduleSeed := range seedData.Schedules {
		startDate, _ := time.Parse("2006-01-02", scheduleSeed.StartDate)
		endDate, _ := time.Parse("2006-01-02", scheduleSeed.EndDate)

		schedule, err := querier.CreateSchedule(ctx, db.CreateScheduleParams{
			Name:            scheduleSeed.Name,
			CronExpr:        scheduleSeed.CronExpr,
			StartDate:       sql.NullTime{Time: startDate, Valid: true},
			EndDate:         sql.NullTime{Time: endDate, Valid: true},
			DurationMinutes: scheduleSeed.DurationMinutes,
			Timezone:        sql.NullString{String: scheduleSeed.Timezone, Valid: true},
		})
		if err != nil {
			logger.Error("Failed to create schedule", "name", scheduleSeed.Name, "error", err)
			return err
		}
		scheduleMap[scheduleSeed.Name] = schedule.ScheduleID
		logger.Info("Created schedule", "name", scheduleSeed.Name, "cron", scheduleSeed.CronExpr)
	}

	// Get existing schedules from migrations
	existingSchedules, err := querier.ListAllSchedules(ctx)
	if err != nil {
		return fmt.Errorf("failed to list existing schedules: %w", err)
	}
	for _, schedule := range existingSchedules {
		scheduleMap[schedule.Name] = schedule.ScheduleID
	}

	// Seed recurring assignments
	for _, assignmentSeed := range seedData.RecurringAssignments {
		userID, ok := userMap[assignmentSeed.UserPhone]
		if !ok {
			logger.Warn("User not found for recurring assignment", "phone", assignmentSeed.UserPhone)
			continue
		}

		scheduleID, ok := scheduleMap[assignmentSeed.ScheduleName]
		if !ok {
			logger.Warn("Schedule not found for recurring assignment", "schedule", assignmentSeed.ScheduleName)
			continue
		}

		params := db.CreateRecurringAssignmentParams{
			UserID:     userID,
			DayOfWeek:  assignmentSeed.DayOfWeek,
			ScheduleID: scheduleID,
			TimeSlot:   assignmentSeed.TimeSlot,
		}

		if assignmentSeed.BuddyName != "" {
			params.BuddyName = sql.NullString{String: assignmentSeed.BuddyName, Valid: true}
		}

		if assignmentSeed.Description != "" {
			params.Description = sql.NullString{String: assignmentSeed.Description, Valid: true}
		}

		assignment, err := querier.CreateRecurringAssignment(ctx, params)
		if err != nil {
			logger.Error("Failed to create recurring assignment", 
				"user", assignmentSeed.UserPhone, 
				"schedule", assignmentSeed.ScheduleName, 
				"error", err)
			return err
		}
		logger.Info("Created recurring assignment", 
			"user", assignmentSeed.UserPhone, 
			"schedule", assignmentSeed.ScheduleName,
			"day", assignmentSeed.DayOfWeek,
			"time_slot", assignmentSeed.TimeSlot,
			"id", assignment.RecurringAssignmentID)
	}

	// Seed sample bookings
	for _, bookingSeed := range seedData.Bookings {
		userID, ok := userMap[bookingSeed.UserPhone]
		if !ok {
			logger.Warn("User not found for booking", "phone", bookingSeed.UserPhone)
			continue
		}

		scheduleID, ok := scheduleMap[bookingSeed.ScheduleName]
		if !ok {
			logger.Warn("Schedule not found for booking", "schedule", bookingSeed.ScheduleName)
			continue
		}

		shiftStart, err := time.Parse(time.RFC3339, bookingSeed.ShiftStart)
		if err != nil {
			logger.Error("Invalid shift start time", "time", bookingSeed.ShiftStart, "error", err)
			continue
		}

		// Get schedule to calculate end time
		schedule, err := querier.GetScheduleByID(ctx, scheduleID)
		if err != nil {
			logger.Error("Failed to get schedule for booking", "schedule_id", scheduleID, "error", err)
			continue
		}

		shiftEnd := shiftStart.Add(time.Duration(schedule.DurationMinutes) * time.Minute)

		params := db.CreateBookingParams{
			UserID:     userID,
			ScheduleID: scheduleID,
			ShiftStart: shiftStart,
			ShiftEnd:   shiftEnd,
		}

		if bookingSeed.BuddyName != "" {
			params.BuddyName = sql.NullString{String: bookingSeed.BuddyName, Valid: true}
		}

		booking, err := querier.CreateBooking(ctx, params)
		if err != nil {
			logger.Error("Failed to create booking", 
				"user", bookingSeed.UserPhone, 
				"schedule", bookingSeed.ScheduleName,
				"shift_start", bookingSeed.ShiftStart,
				"error", err)
			// Continue with other bookings rather than failing
			continue
		}

		// Update attendance if specified
		if bookingSeed.Attended {
			_, err = querier.UpdateBookingAttendance(ctx, db.UpdateBookingAttendanceParams{
				BookingID: booking.BookingID,
				Attended:  bookingSeed.Attended,
			})
			if err != nil {
				logger.Warn("Failed to update booking attendance", 
					"booking_id", booking.BookingID, 
					"error", err)
			}
		}

		logger.Info("Created booking", 
			"user", bookingSeed.UserPhone, 
			"schedule", bookingSeed.ScheduleName,
			"shift_start", shiftStart.Format(time.RFC3339),
			"id", booking.BookingID)
	}

	return nil
}

func getSeedData() SeedData {
	return SeedData{
		Users: []UserSeed{
			// Admin users
			{Name: "Alice Admin", Phone: "+27821234567", Role: "admin"},
			{Name: "Bob Manager", Phone: "+27821234568", Role: "admin"},

			// Owl volunteers
			{Name: "Charlie Volunteer", Phone: "+27821234569", Role: "owl"},
			{Name: "Diana Scout", Phone: "+27821234570", Role: "owl"},
			{Name: "Eve Patrol", Phone: "+27821234571", Role: "owl"},
			{Name: "Frank Guard", Phone: "+27821234572", Role: "owl"},
			{Name: "Grace Watch", Phone: "+27821234573", Role: "owl"},
			{Name: "Henry Security", Phone: "+27821234574", Role: "owl"},

			// Guest users
			{Name: "Iris Guest", Phone: "+27821234575", Role: "guest"},
			{Name: "Jack Visitor", Phone: "+27821234576", Role: "guest"},
		},

		Schedules: []ScheduleSeed{
			// Development schedules with more frequent shifts for testing
			{
				Name:            "Daily Evening Patrol",
				CronExpr:        "0 18 * * *", // Every day at 6 PM
				StartDate:       "2024-01-01",
				EndDate:         "2024-12-31",
				DurationMinutes: 120,
				Timezone:        "Africa/Johannesburg",
			},
			{
				Name:            "Weekend Morning Watch",
				CronExpr:        "0 6,10 * * 6,0", // Sat/Sun at 6 AM and 10 AM
				StartDate:       "2024-01-01",
				EndDate:         "2024-12-31",
				DurationMinutes: 240, // 4 hours
				Timezone:        "Africa/Johannesburg",
			},
			{
				Name:            "Weekday Lunch Security",
				CronExpr:        "0 12 * * 1-5", // Mon-Fri at noon
				StartDate:       "2024-01-01",
				EndDate:         "2024-12-31",
				DurationMinutes: 60,
				Timezone:        "Africa/Johannesburg",
			},
		},

		RecurringAssignments: []RecurringAssignmentSeed{
			// Charlie on weekend mornings
			{
				UserPhone:    "+27821234569", // Charlie
				ScheduleName: "Weekend Morning Watch",
				DayOfWeek:    6, // Saturday
				TimeSlot:     "06:00-10:00",
				BuddyName:    "Diana Scout",
				Description:  "Regular Saturday morning patrol",
			},
			{
				UserPhone:    "+27821234570", // Diana
				ScheduleName: "Weekend Morning Watch", 
				DayOfWeek:    0, // Sunday
				TimeSlot:     "10:00-14:00",
				BuddyName:    "Charlie Volunteer",
				Description:  "Sunday morning community watch",
			},

			// Eve on daily evening patrol
			{
				UserPhone:    "+27821234571", // Eve
				ScheduleName: "Daily Evening Patrol",
				DayOfWeek:    1, // Monday
				TimeSlot:     "18:00-20:00",
				Description:  "Monday evening patrol",
			},
			{
				UserPhone:    "+27821234571", // Eve
				ScheduleName: "Daily Evening Patrol",
				DayOfWeek:    3, // Wednesday
				TimeSlot:     "18:00-20:00",
				Description:  "Wednesday evening patrol",
			},

			// Frank on weekday lunch
			{
				UserPhone:    "+27821234572", // Frank
				ScheduleName: "Weekday Lunch Security",
				DayOfWeek:    2, // Tuesday
				TimeSlot:     "12:00-13:00",
				Description:  "Tuesday lunch security",
			},
			{
				UserPhone:    "+27821234572", // Frank
				ScheduleName: "Weekday Lunch Security",
				DayOfWeek:    4, // Thursday
				TimeSlot:     "12:00-13:00",
				Description:  "Thursday lunch security",
			},

			// Grace on summer patrol (from migration)
			{
				UserPhone:    "+27821234573", // Grace
				ScheduleName: "Summer Patrol (Nov-Apr)",
				DayOfWeek:    6, // Saturday
				TimeSlot:     "00:00-02:00",
				BuddyName:    "Henry Security",
				Description:  "Summer Saturday night patrol",
			},
		},

		Bookings: []BookingSeed{
			// Some historical bookings for testing
			{
				UserPhone:    "+27821234569", // Charlie
				ScheduleName: "Daily Evening Patrol",
				ShiftStart:   "2024-11-25T18:00:00Z", // Recent Monday
				BuddyName:    "Diana Scout",
				Attended:     true,
			},
			{
				UserPhone:    "+27821234570", // Diana
				ScheduleName: "Weekend Morning Watch",
				ShiftStart:   "2024-11-24T06:00:00Z", // Recent Sunday
				BuddyName:    "Charlie Volunteer",
				Attended:     true,
			},
			{
				UserPhone:    "+27821234571", // Eve
				ScheduleName: "Daily Evening Patrol",
				ShiftStart:   "2024-11-26T18:00:00Z", // Recent Tuesday
				Attended:     false, // Missed shift
			},
			{
				UserPhone:    "+27821234572", // Frank
				ScheduleName: "Weekday Lunch Security",
				ShiftStart:   "2024-11-26T12:00:00Z", // Recent Tuesday
				Attended:     true,
			},
		},
	}
}

func showSeedData(logger *slog.Logger) {
	seedData := getSeedData()
	
	logger.Info("=== SEED DATA PREVIEW ===")
	
	logger.Info("Users to be created", "count", len(seedData.Users))
	for _, user := range seedData.Users {
		logger.Info("  User", "name", user.Name, "phone", user.Phone, "role", user.Role)
	}
	
	logger.Info("Schedules to be created", "count", len(seedData.Schedules))
	for _, schedule := range seedData.Schedules {
		logger.Info("  Schedule", "name", schedule.Name, "cron", schedule.CronExpr)
	}
	
	logger.Info("Recurring assignments to be created", "count", len(seedData.RecurringAssignments))
	for _, assignment := range seedData.RecurringAssignments {
		logger.Info("  Assignment", 
			"user", assignment.UserPhone, 
			"schedule", assignment.ScheduleName,
			"day", assignment.DayOfWeek,
			"time", assignment.TimeSlot)
	}
	
	logger.Info("Bookings to be created", "count", len(seedData.Bookings))
	for _, booking := range seedData.Bookings {
		logger.Info("  Booking", 
			"user", booking.UserPhone, 
			"schedule", booking.ScheduleName,
			"shift", booking.ShiftStart,
			"attended", booking.Attended)
	}
} 