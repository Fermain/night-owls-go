package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
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
		userCount = flag.Int("users", 10, "Number of users to create (default: 10)")
		futureBookings = flag.Bool("future-bookings", false, "Generate future bookings for next 30 days")
		exportJSON = flag.String("export", "", "Export seeded data to JSON file")
		verbose = flag.Bool("verbose", false, "Enable verbose logging")
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
	if *verbose {
		// Set log level to debug for verbose output
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	slog.SetDefault(logger)

	// Determine database path
	databasePath := cfg.DatabasePath
	if *dbPath != "" {
		databasePath = *dbPath
	}

	logger.Info("Starting seeding process", 
		"database", databasePath, 
		"reset", *reset, 
		"dry_run", *dryRun,
		"user_count", *userCount,
		"future_bookings", *futureBookings,
		"export", *exportJSON)

	if *dryRun {
		logger.Info("DRY RUN MODE - No actual changes will be made")
		showSeedData(logger, *userCount, *futureBookings)
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
	seededData, err := seedDatabase(context.Background(), querier, logger, *userCount, *futureBookings)
	if err != nil {
		logger.Error("Seeding failed", "error", err)
		os.Exit(1)
	}

	// Export data if requested
	if *exportJSON != "" {
		if err := exportSeededData(seededData, *exportJSON, logger); err != nil {
			logger.Error("Failed to export data", "file", *exportJSON, "error", err)
			os.Exit(1)
		}
		logger.Info("Data exported successfully", "file", *exportJSON)
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

func seedDatabase(ctx context.Context, querier db.Querier, logger *slog.Logger, userCount int, futureBookings bool) (SeedData, error) {
	seedData := getSeedDataWithOptions(userCount, futureBookings)

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
			return SeedData{}, err
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
			return SeedData{}, err
		}
		scheduleMap[scheduleSeed.Name] = schedule.ScheduleID
		logger.Info("Created schedule", "name", scheduleSeed.Name, "cron", scheduleSeed.CronExpr)
	}

	// Get existing schedules from migrations
	existingSchedules, err := querier.ListAllSchedules(ctx)
	if err != nil {
		return SeedData{}, fmt.Errorf("failed to list existing schedules: %w", err)
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
			return SeedData{}, err
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

		// Update check-in status if specified
		if bookingSeed.Attended {
			checkInTime := sql.NullTime{Time: shiftStart.Add(10 * time.Minute), Valid: true} // Check in 10 minutes after shift start
			_, err = querier.UpdateBookingCheckIn(ctx, db.UpdateBookingCheckInParams{
				BookingID:   booking.BookingID,
				CheckedInAt: checkInTime,
			})
			if err != nil {
				logger.Warn("Failed to update booking check-in", 
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

	// Seed reports after bookings are created
	logger.Info("Seeding reports...")
	if err := seedReports(querier); err != nil {
		logger.Error("Failed to seed reports", "error", err)
		return SeedData{}, err
	}

	// Seed some recent critical reports for demo purposes
	if err := seedRecentCriticalReports(querier); err != nil {
		logger.Error("Failed to seed recent critical reports", "error", err)
		return SeedData{}, err
	}

	return seedData, nil
}

func getSeedData() SeedData {
	return getSeedDataWithOptions(10, false)
}

func getSeedDataWithOptions(userCount int, includeFutureBookings bool) SeedData {
	users := generateUsers(userCount)
	
	// No additional schedules - rely only on migration schedules
	schedules := []ScheduleSeed{}
	
	recurringAssignments := []RecurringAssignmentSeed{
		// Charlie on Old schedule
		{
			UserPhone:    "+27821234569", // Charlie
			ScheduleName: "Old schedule",
			DayOfWeek:    1, // Monday
			TimeSlot:     "00:00-02:00",
			BuddyName:    "Diana Scout",
			Description:  "Monday midnight shift",
		},
		{
			UserPhone:    "+27821234570", // Diana
			ScheduleName: "Old schedule", 
			DayOfWeek:    2, // Tuesday
			TimeSlot:     "00:00-02:00",
			BuddyName:    "Charlie Volunteer",
			Description:  "Tuesday midnight shift",
		},

		// Eve on Old schedule
		{
			UserPhone:    "+27821234571", // Eve
			ScheduleName: "Old schedule",
			DayOfWeek:    3, // Wednesday
			TimeSlot:     "00:00-02:00",
			Description:  "Wednesday midnight shift",
		},
		{
			UserPhone:    "+27821234571", // Eve
			ScheduleName: "Old schedule",
			DayOfWeek:    4, // Thursday
			TimeSlot:     "00:00-02:00",
			Description:  "Thursday midnight shift",
		},

		// Frank on Old schedule
		{
			UserPhone:    "+27821234572", // Frank
			ScheduleName: "Old schedule",
			DayOfWeek:    5, // Friday
			TimeSlot:     "00:00-02:00",
			Description:  "Friday midnight shift",
		},
		{
			UserPhone:    "+27821234572", // Frank
			ScheduleName: "Old schedule",
			DayOfWeek:    6, // Saturday
			TimeSlot:     "00:00-02:00",
			Description:  "Saturday midnight shift",
		},

		// Grace on Old schedule
		{
			UserPhone:    "+27821234573", // Grace
			ScheduleName: "Old schedule",
			DayOfWeek:    0, // Sunday
			TimeSlot:     "00:00-02:00",
			BuddyName:    "Henry Security",
			Description:  "Sunday midnight shift",
		},
	}
	
	// Filter recurring assignments based on available users
	var filteredAssignments []RecurringAssignmentSeed
	userPhones := make(map[string]bool)
	for _, user := range users {
		userPhones[user.Phone] = true
	}
	
	for _, assignment := range recurringAssignments {
		if userPhones[assignment.UserPhone] {
			filteredAssignments = append(filteredAssignments, assignment)
		}
	}
	
	historicalBookings := []BookingSeed{
		// Historical bookings for Old schedule with some reports
		{
			UserPhone:    "+27821234569", // Charlie
			ScheduleName: "Old schedule",
			ShiftStart:   "2024-11-25T00:00:00Z", // Recent Monday
			BuddyName:    "Diana Scout",
			Attended:     true,
		},
		{
			UserPhone:    "+27821234570", // Diana
			ScheduleName: "Old schedule",
			ShiftStart:   "2024-11-26T00:00:00Z", // Recent Tuesday
			BuddyName:    "Charlie Volunteer",
			Attended:     true,
		},
		{
			UserPhone:    "+27821234571", // Eve
			ScheduleName: "Old schedule",
			ShiftStart:   "2024-11-27T00:00:00Z", // Recent Wednesday
			Attended:     false, // Missed shift
		},
		{
			UserPhone:    "+27821234571", // Eve
			ScheduleName: "Old schedule",
			ShiftStart:   "2024-11-28T00:00:00Z", // Recent Thursday
			Attended:     true,
		},
		{
			UserPhone:    "+27821234572", // Frank
			ScheduleName: "Old schedule",
			ShiftStart:   "2024-11-29T00:00:00Z", // Recent Friday
			Attended:     true,
		},
		{
			UserPhone:    "+27821234572", // Frank
			ScheduleName: "Old schedule",
			ShiftStart:   "2024-11-30T00:00:00Z", // Recent Saturday
			Attended:     false, // Missed shift
		},
		{
			UserPhone:    "+27821234573", // Grace
			ScheduleName: "Old schedule",
			ShiftStart:   "2024-12-01T00:00:00Z", // Recent Sunday
			BuddyName:    "Henry Security",
			Attended:     true,
		},
		// More historical bookings for better report data
		{
			UserPhone:    "+27821234569", // Charlie
			ScheduleName: "Old schedule",
			ShiftStart:   "2024-11-18T00:00:00Z", // Previous Monday
			BuddyName:    "Diana Scout",
			Attended:     true,
		},
		{
			UserPhone:    "+27821234570", // Diana
			ScheduleName: "Old schedule",
			ShiftStart:   "2024-11-19T00:00:00Z", // Previous Tuesday
			BuddyName:    "Charlie Volunteer",
			Attended:     true,
		},
		{
			UserPhone:    "+27821234571", // Eve
			ScheduleName: "Old schedule",
			ShiftStart:   "2024-11-20T00:00:00Z", // Previous Wednesday
			Attended:     true,
		},
		{
			UserPhone:    "+27821234572", // Frank
			ScheduleName: "Old schedule",
			ShiftStart:   "2024-11-21T00:00:00Z", // Previous Thursday
			Attended:     true,
		},
		{
			UserPhone:    "+27821234573", // Grace
			ScheduleName: "Old schedule",
			ShiftStart:   "2024-11-22T00:00:00Z", // Previous Friday
			BuddyName:    "Henry Security",
			Attended:     false, // Missed shift
		},
		{
			UserPhone:    "+27821234574", // Henry
			ScheduleName: "Old schedule",
			ShiftStart:   "2024-11-23T00:00:00Z", // Previous Saturday
			Attended:     true,
		},
		{
			UserPhone:    "+27821234577", // Leo
			ScheduleName: "Old schedule",
			ShiftStart:   "2024-11-24T00:00:00Z", // Previous Sunday
			Attended:     true,
		},
		// Even older bookings for more data
		{
			UserPhone:    "+27821234578", // Zoe
			ScheduleName: "Old schedule",
			ShiftStart:   "2024-11-11T00:00:00Z",
			Attended:     true,
		},
		{
			UserPhone:    "+27821234580", // Ivy
			ScheduleName: "Old schedule",
			ShiftStart:   "2024-11-12T00:00:00Z",
			Attended:     true,
		},
		{
			UserPhone:    "+27821234581", // Sam
			ScheduleName: "Old schedule",
			ShiftStart:   "2024-11-13T00:00:00Z",
			Attended:     false,
		},
		{
			UserPhone:    "+27821234569", // Charlie
			ScheduleName: "Old schedule",
			ShiftStart:   "2024-11-14T00:00:00Z",
			Attended:     true,
		},
		{
			UserPhone:    "+27821234570", // Diana
			ScheduleName: "Old schedule",
			ShiftStart:   "2024-11-15T00:00:00Z",
			Attended:     true,
		},
	}
	
	// Filter historical bookings based on available users
	var filteredBookings []BookingSeed
	for _, booking := range historicalBookings {
		if userPhones[booking.UserPhone] {
			filteredBookings = append(filteredBookings, booking)
		}
	}
	
	// Add future bookings if requested
	if includeFutureBookings {
		// Create user and schedule maps for future booking generation
		userMap := make(map[string]int64)
		scheduleMap := make(map[string]int64)
		
		// These would be populated during actual seeding
		// For now, we'll generate future bookings based on known users
		futureBookings := generateFutureBookings(userMap, scheduleMap)
		
		// Filter future bookings based on available users
		for _, booking := range futureBookings {
			if userPhones[booking.UserPhone] {
				filteredBookings = append(filteredBookings, booking)
			}
		}
	}

	return SeedData{
		Users:               users,
		Schedules:           schedules,
		RecurringAssignments: filteredAssignments,
		Bookings:            filteredBookings,
	}
}

func showSeedData(logger *slog.Logger, userCount int, futureBookings bool) {
	seedData := getSeedDataWithOptions(userCount, futureBookings)
	
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
	
	if futureBookings {
		futureCount := 0
		for _, booking := range seedData.Bookings {
			// Count future bookings (approximate)
			if strings.Contains(booking.ShiftStart, "2024-12") || strings.Contains(booking.ShiftStart, "2025") {
				futureCount++
			}
		}
		logger.Info("Future bookings included", "estimated_count", futureCount)
	}
}

func exportSeededData(seedData SeedData, filePath string, logger *slog.Logger) error {
	logger.Info("Exporting seeded data", "file", filePath)
	
	// Create export structure with metadata
	exportData := struct {
		ExportedAt time.Time `json:"exported_at"`
		Version    string    `json:"version"`
		Database   string    `json:"database"`
		Data       SeedData  `json:"data"`
	}{
		ExportedAt: time.Now().UTC(),
		Version:    "1.0",
		Database:   "Night Owls Go",
		Data:       seedData,
	}
	
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create export file: %w", err)
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(exportData); err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}
	
	return nil
}

func generateFutureBookings(userMap map[string]int64, scheduleMap map[string]int64) []BookingSeed {
	var futureBookings []BookingSeed
	now := time.Now()
	
	// Generate bookings for the next 30 days
	for i := 1; i <= 30; i++ {
		futureDate := now.AddDate(0, 0, i)
		weekday := int(futureDate.Weekday())
		
		// Add Old schedule bookings (continue existing pattern)
		if weekday >= 1 && weekday <= 5 { // Monday to Friday
			futureBookings = append(futureBookings, BookingSeed{
				UserPhone:    "+27821234571", // Eve Patrol
				ScheduleName: "Old schedule",
				ShiftStart:   futureDate.Format("2006-01-02") + "T00:00:00Z",
				BuddyName:    "",
				Attended:     false, // Future bookings default to not attended
			})
		}
		
		// Add weekend Old schedule bookings
		if weekday == 6 || weekday == 0 { // Saturday or Sunday
			userPhone := "+27821234572" // Frank
			if weekday == 0 { // Sunday
				userPhone = "+27821234573" // Grace
			}
			
			futureBookings = append(futureBookings, BookingSeed{
				UserPhone:    userPhone,
				ScheduleName: "Old schedule",
				ShiftStart:   futureDate.Format("2006-01-02") + "T00:00:00Z",
				BuddyName:    "Henry Security",
				Attended:     false,
			})
		}
		
		// Occasionally add New schedule bookings (for variety)
		if i%7 == 0 { // Every 7th day
			futureBookings = append(futureBookings, BookingSeed{
				UserPhone:    "+27821234574", // Henry Security
				ScheduleName: "New schedule",
				ShiftStart:   futureDate.Format("2006-01-02") + "T00:00:00Z",
				BuddyName:    "",
				Attended:     false,
			})
		}
	}
	
	return futureBookings
}

func generateUsers(count int) []UserSeed {
	if count <= 10 {
		// Return the default user set for small counts
		return []UserSeed{
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
		}[:count]
	}
	
	// Generate additional users for larger counts
	users := generateUsers(10) // Start with base 10
	
	owlNames := []string{"Leo", "Zoe", "Max", "Ivy", "Sam", "Ruby", "Alex", "Nova", "Finn", "Luna"}
	guestNames := []string{"Maya", "Ryan", "Aria", "Dean", "Nora", "Kyle", "Sage", "Troy", "Vale", "Reed"}
	
	phoneBase := 27821234577 // Continue from last default user
	
	for i := 10; i < count; i++ {
		var name, role string
		
		if i%4 == 0 { // Every 4th user is a guest
			role = "guest"
			name = fmt.Sprintf("%s Guest", guestNames[(i-10)%len(guestNames)])
		} else { // Rest are owls
			role = "owl"
			name = fmt.Sprintf("%s Owl", owlNames[(i-10)%len(owlNames)])
		}
		
		users = append(users, UserSeed{
			Name:  name,
			Phone: fmt.Sprintf("+%d", phoneBase+i-10),
			Role:  role,
		})
	}
	
	return users
} 