package api_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"night-owls-go/internal/api"
	"night-owls-go/internal/auth"
	"night-owls-go/internal/config"
	db "night-owls-go/internal/db/sqlc_generated"
	"night-owls-go/internal/outbox"
	"night-owls-go/internal/service"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/gorhill/cronexpr"
	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

// MockMessageSender for integration tests (local to this package test)
type MockMessageSender struct {
	mock.Mock
}

func (m *MockMessageSender) Send(recipient, messageType, payload string) error {
	args := m.Called(recipient, messageType, payload)
	return args.Error(0)
}

type adminTestApp struct {
	Router          *chi.Mux
	DB              *sql.DB
	Logger          *slog.Logger
	Config          *config.Config
	Querier         db.Querier
	UserService     *service.UserService
	ScheduleService *service.ScheduleService
	BookingService  *service.BookingService
	ReportService   *service.ReportService
	OutboxService   *outbox.DispatcherService
	PushService     *service.PushSender
	mockSMSSender   *MockMessageSender
	Cron            *cron.Cron
	OTPStore        auth.OTPStore
}

// newAdminTestApp sets up the application for admin-related integration tests.
// It ensures all migrations are run and admin routes are configured.
func newAdminTestApp(t *testing.T) *adminTestApp {
	t.Helper()

	cfg := &config.Config{
		ServerPort:         "0",
		DatabasePath:       ":memory:",
		JWTSecret:          "test-jwt-secret-admin",
		DefaultShiftDuration: 2 * time.Hour,
		OTPLogPath:         os.DevNull,
		LogLevel:           "debug",
		LogFormat:          "text",
		JWTExpirationHours: 1,
		OTPValidityMinutes: 5,
		OutboxBatchSize:    10,
		OutboxMaxRetries:   3,
		StaticDir:          "../../app/build", // Path relative to internal/api
		VAPIDPublic:        "test_public_key",
		VAPIDPrivate:       "test_private_key",
		VAPIDSubject:       "mailto:test@example.com",
	}

	loggerOpts := &slog.HandlerOptions{Level: slog.LevelDebug}
	logger := slog.New(slog.NewTextHandler(io.Discard, loggerOpts)) // Discard logs for cleaner test output unless debugging

	dbConn, err := sql.Open("sqlite", cfg.DatabasePath+"?cache=shared&_foreign_keys=on")
	require.NoError(t, err, "Failed to open in-memory DB for admin app")

	// Apply ALL migrations
	_, currentFilePath, _, ok := runtime.Caller(0)
	require.True(t, ok, "Failed to get current file path")
	currentDir := filepath.Dir(currentFilePath)
	migrationsDir := filepath.Join(currentDir, "..", "db", "migrations") //  internal/api -> internal/db/migrations

	migrationFiles, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	require.NoError(t, err, "Failed to glob migration files")
	require.NotEmpty(t, migrationFiles, "No migration files found")
	
	// Sort migration files to ensure correct order if not already sorted by Glob
	// For simple numeric prefixes, default sort usually works.
	// sort.Strings(migrationFiles) // If more complex naming needed

	for _, migrationFile := range migrationFiles {
		sqlBytes, err := os.ReadFile(migrationFile)
		require.NoError(t, err, fmt.Sprintf("Failed to read migration file: %s", migrationFile))
		
		// Execute the entire migration file as one statement instead of splitting
		// This handles inline comments properly
        sqlContent := strings.TrimSpace(string(sqlBytes))
        if sqlContent == "" {
            continue
        }
        _, err = dbConn.Exec(sqlContent)
        if err != nil {
            // If executing as one statement fails, try splitting by semicolon followed by newline
            queries := strings.Split(sqlContent, ";\n")
            for i, query := range queries {
                trimmedQuery := strings.TrimSpace(query)
                if trimmedQuery == "" || strings.HasPrefix(trimmedQuery, "--") {
                    continue
                }
                // Add semicolon back if it was the last statement and doesn't end with one
                if i == len(queries)-1 && !strings.HasSuffix(trimmedQuery, ";") {
                    trimmedQuery += ";"
                }
                _, err = dbConn.Exec(trimmedQuery)
                require.NoError(t, err, fmt.Sprintf("Failed to execute migration query from %s: %s", migrationFile, trimmedQuery))
            }
        }
	}

	err = dbConn.Ping()
	require.NoError(t, err, "Admin app's DB connection failed after manual migrations")

	querier := db.New(dbConn)
	otpStore := auth.NewInMemoryOTPStore()
	mockSender := new(MockMessageSender)

	userService := service.NewUserService(querier, otpStore, cfg, logger)
	scheduleService := service.NewScheduleService(querier, logger, cfg)
	bookingService := service.NewBookingService(querier, cfg, logger)
	reportService := service.NewReportService(querier, logger)
	pushService := service.NewPushSender(querier, cfg, logger)
	outboxService := outbox.NewDispatcherService(querier, mockSender, pushService, logger, cfg)


	cronScheduler := cron.New()
	// todo: setup cron jobs if they interfere or are needed by test flows

	router := chi.NewRouter()
	router.Use(chiMiddleware.Recoverer)

	// Register all relevant handlers, including admin
	authAPIHandler := api.NewAuthHandler(userService, logger, cfg, querier)
	bookingAPIHandler := api.NewBookingHandler(bookingService, logger)
	adminScheduleAPIHandler := api.NewAdminScheduleHandlers(logger, scheduleService)
	adminUserAPIHandler := api.NewAdminUserHandler(querier, logger)
	adminBookingAPIHandler := api.NewAdminBookingHandler(bookingService, logger)
	adminReportAPIHandler := api.NewAdminReportHandler(reportService, scheduleService, querier, logger)
	adminBroadcastAPIHandler := api.NewAdminBroadcastHandler(querier, logger)
	adminDashboardAPIHandler := api.NewAdminDashboardHandler(nil, logger) // Using nil service for simple implementation
	// pushAPIHandler := api.NewPushHandler(querier, cfg, logger) // If needed

	// Public routes
	router.Post("/auth/register", authAPIHandler.RegisterHandler)
	router.Post("/auth/verify", authAPIHandler.VerifyHandler)
	// ... other public routes if necessary for setup ...

	// Admin routes
	router.Route("/api/admin", func(r chi.Router) {
		r.Use(api.AuthMiddleware(cfg, logger)) // Apply AuthMiddleware to all admin routes
		r.Use(api.AdminMiddleware(logger))     // Apply AdminMiddleware to require admin role

		// Admin Schedules
		r.Route("/schedules", func(sr chi.Router) {
			sr.Get("/", adminScheduleAPIHandler.AdminListSchedules)
			sr.Post("/", adminScheduleAPIHandler.AdminCreateSchedule)
			sr.Get("/{id}", adminScheduleAPIHandler.AdminGetSchedule)
			sr.Put("/{id}", adminScheduleAPIHandler.AdminUpdateSchedule)
			sr.Delete("/{id}", adminScheduleAPIHandler.AdminDeleteSchedule)
			sr.Get("/all-slots", adminScheduleAPIHandler.AdminListAllShiftSlots)
			sr.Post("/bulk-delete", adminScheduleAPIHandler.AdminBulkDeleteSchedules)
		})
		// Admin Users
		r.Route("/users", func(ur chi.Router) {
			ur.Get("/", adminUserAPIHandler.AdminListUsers)
			ur.Get("/{id}", adminUserAPIHandler.AdminGetUser)
			ur.Post("/", adminUserAPIHandler.AdminCreateUser)
			ur.Put("/{id}", adminUserAPIHandler.AdminUpdateUser)
			ur.Delete("/{id}", adminUserAPIHandler.AdminDeleteUser)
			ur.Post("/bulk-delete", adminUserAPIHandler.AdminBulkDeleteUsers)
		})
		// Admin Bookings
		r.Route("/bookings", func(br chi.Router) {
			br.Post("/assign", adminBookingAPIHandler.AssignUserToShiftHandler)
		})
		// Admin Reports
		r.Route("/reports", func(rr chi.Router) {
			rr.Get("/", adminReportAPIHandler.AdminListReportsHandler)
			rr.Get("/{id}", adminReportAPIHandler.AdminGetReportHandler)
			rr.Put("/{id}/archive", adminReportAPIHandler.AdminArchiveReportHandler)
			rr.Put("/{id}/unarchive", adminReportAPIHandler.AdminUnarchiveReportHandler)
			rr.Get("/archived", adminReportAPIHandler.AdminListArchivedReportsHandler)
		})
		// Admin Broadcasts
		r.Route("/broadcasts", func(br chi.Router) {
			br.Get("/", adminBroadcastAPIHandler.AdminListBroadcasts)
			br.Post("/", adminBroadcastAPIHandler.AdminCreateBroadcast)
			br.Get("/{id}", adminBroadcastAPIHandler.AdminGetBroadcast)
		})
		// Admin Dashboard
		r.Get("/dashboard", adminDashboardAPIHandler.GetDashboardHandler)
	})
	
	// Also register protected user routes if admin might interact with them or if setup requires it
	router.Group(func(r chi.Router) {
		r.Use(api.AuthMiddleware(cfg, logger))
		r.Post("/bookings", bookingAPIHandler.CreateBookingHandler)
		// ... other protected routes
	})


	return &adminTestApp{
		Router:          router,
		DB:              dbConn,
		Logger:          logger,
		Config:          cfg,
		Querier:         querier,
		UserService:     userService,
		ScheduleService: scheduleService,
		BookingService:  bookingService,
		ReportService:   reportService,
		PushService:     pushService,
		OutboxService:   outboxService,
		mockSMSSender:   mockSender,
		Cron:            cronScheduler,
		OTPStore:        otpStore,
	}
}

func (app *adminTestApp) makeRequest(t *testing.T, method, path string, body io.Reader, token string) *httptest.ResponseRecorder {
	t.Helper()
	req, err := http.NewRequest(method, path, body)
	require.NoError(t, err)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rr := httptest.NewRecorder()
	app.Router.ServeHTTP(rr, req)
	return rr
}

func (app *adminTestApp) createTestUserAndLogin(t *testing.T, phone, name, role string) (db.User, string) {
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
	app.OTPStore.StoreOTP(phone, otp, 5*time.Minute)

	token, err := app.UserService.VerifyOTP(ctx, phone, otp)
	require.NoError(t, err, "Failed to verify OTP and get token for test user %s", phone)
	require.NotEmpty(t, token)
	return user, token
}

func TestAdminAssignUserToShift_Success(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	adminUser, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")
	require.Equal(t, "admin", adminUser.Role)

	targetUserParams := db.CreateUserParams{
		Phone: "+15550001002",
		Name:  sql.NullString{String: "Target Owl", Valid: true},
		Role:  sql.NullString{String: "owl", Valid: true},
	}
	targetUser, err := app.Querier.CreateUser(context.Background(), targetUserParams)
	require.NoError(t, err)

	ctx := context.Background()
	scheduleName := "Nightly Watch Test Schedule"
	cronExpr := "0 22 * * *" // Every day at 10 PM UTC
	durationMinutes := int64(120)

	testSchedule, err := app.Querier.CreateSchedule(ctx, db.CreateScheduleParams{
		Name:            scheduleName,
		CronExpr:        cronExpr,
		DurationMinutes: durationMinutes,
		Timezone:        sql.NullString{String: "UTC", Valid: true},
	})
	require.NoError(t, err)

	locUTC, _ := time.LoadLocation("UTC")
	tomorrow := time.Now().In(locUTC).AddDate(0, 0, 1)
	startOfTomorrow := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, locUTC)

	parsedCron, err := cronexpr.Parse(testSchedule.CronExpr)
	require.NoError(t, err)
	shiftStartTime := parsedCron.Next(startOfTomorrow)
	require.False(t, shiftStartTime.IsZero(), "Could not determine next shift start time")

	assignRequest := api.AssignUserToShiftRequest{
		ScheduleID: testSchedule.ScheduleID,
		StartTime:  shiftStartTime,
		UserID:     targetUser.UserID,
	}
	payloadBytes, _ := json.Marshal(assignRequest)

	rr := app.makeRequest(t, "POST", "/api/admin/bookings/assign", bytes.NewBuffer(payloadBytes), adminToken)

	require.Equal(t, http.StatusCreated, rr.Code, "Expected StatusCreated. Response: %s", rr.Body.String())

	var bookingResp api.BookingResponse
	err = json.Unmarshal(rr.Body.Bytes(), &bookingResp)
	require.NoError(t, err, "Failed to unmarshal BookingResponse")

	assert.Equal(t, targetUser.UserID, bookingResp.UserID)
	assert.Equal(t, testSchedule.ScheduleID, bookingResp.ScheduleID)
	assert.True(t, shiftStartTime.Equal(bookingResp.ShiftStart.In(locUTC)), "ShiftStartTime mismatch. Expected %v (UTC), Got %v (UTC)", shiftStartTime, bookingResp.ShiftStart.In(locUTC))
	assert.Nil(t, bookingResp.CheckedInAt)

	dbBooking, err := app.Querier.GetBookingByID(ctx, bookingResp.BookingID)
	require.NoError(t, err, "Failed to get booking from DB by ID")
	assert.Equal(t, targetUser.UserID, dbBooking.UserID)
	assert.Equal(t, testSchedule.ScheduleID, dbBooking.ScheduleID)
	assert.True(t, shiftStartTime.Equal(dbBooking.ShiftStart.In(locUTC)), "DB ShiftStartTime mismatch")

	outboxItems, err := app.Querier.GetPendingOutboxItems(ctx, 5)
	require.NoError(t, err)

	foundOutboxMsg := false
	for _, item := range outboxItems {
		if item.MessageType == "ADMIN_SHIFT_ASSIGNMENT" && item.UserID.Valid && item.UserID.Int64 == targetUser.UserID {
			var outboxPayload struct {
				BookingID  int64  `json:"booking_id"`
				UserID     int64  `json:"user_id"`
				AssignedBy string `json:"assigned_by"`
			}
			errJson := json.Unmarshal([]byte(item.Payload.String), &outboxPayload)
			require.NoError(t, errJson)
			assert.Equal(t, bookingResp.BookingID, outboxPayload.BookingID)
			assert.Equal(t, targetUser.UserID, outboxPayload.UserID)
			assert.Equal(t, "admin", outboxPayload.AssignedBy)
			foundOutboxMsg = true
			break
		}
	}
	assert.True(t, foundOutboxMsg, "Expected ADMIN_SHIFT_ASSIGNMENT outbox message for the target user")
}

// TODO: Add more tests for failure cases:
// - Non-admin user trying to assign
// - Assigning to non-existent user
// - Assigning to non-existent schedule
// - Assigning to invalid shift time
// - Assigning to already booked slot
// - Invalid request payload

// parseSQLStatements parses SQL content and splits it into individual statements,
// handling comments and multi-line statements correctly.
func parseSQLStatements(sqlContent string) []string {
	var statements []string
	var currentStatement strings.Builder
	inSingleLineComment := false
	inMultiLineComment := false
	inString := false
	var stringDelimiter rune
	
	lines := strings.Split(sqlContent, "\n")
	
	for _, line := range lines {
		originalLine := line
		line = strings.TrimSpace(line)
		
		// Skip migration directives
		if strings.HasPrefix(strings.TrimSpace(line), "-- +migrate") {
			continue
		}
		
		// Handle single-line comments that start the line
		if strings.HasPrefix(strings.TrimSpace(line), "--") && !inString && !inMultiLineComment {
			// Skip this line entirely
			continue
		}
		
		for i, char := range line {
			if inSingleLineComment {
				// Single line comments end at line end, will be reset below
				break
			} else if inMultiLineComment {
				if char == '*' && i+1 < len(line) && rune(line[i+1]) == '/' {
					inMultiLineComment = false
					// Skip the '/' as well
					continue
				}
				continue
			} else if inString {
				currentStatement.WriteRune(char)
				if char == stringDelimiter {
					// Check if it's escaped
					if i == 0 || rune(line[i-1]) != '\\' {
						inString = false
					}
				}
			} else {
				// Not in comment or string
				switch char {
				case '\'', '"':
					inString = true
					stringDelimiter = char
					currentStatement.WriteRune(char)
				case '/':
					if i+1 < len(line) && rune(line[i+1]) == '*' {
						inMultiLineComment = true
						// Skip the '*' as well
						continue
					} else {
						currentStatement.WriteRune(char)
					}
				case '-':
					if i+1 < len(line) && rune(line[i+1]) == '-' {
						inSingleLineComment = true
						break // Skip rest of line
					} else {
						currentStatement.WriteRune(char)
					}
				case ';':
					// End of statement
					stmt := strings.TrimSpace(currentStatement.String())
					if stmt != "" {
						statements = append(statements, stmt)
					}
					currentStatement.Reset()
				default:
					currentStatement.WriteRune(char)
				}
			}
		}
		
		// Reset single-line comment flag at end of line
		inSingleLineComment = false
		
		// Add newline to preserve formatting (except for comment lines)
		if !strings.HasPrefix(strings.TrimSpace(originalLine), "--") {
			currentStatement.WriteRune('\n')
		}
	}
	
	// Handle final statement if no trailing semicolon
	if finalStmt := strings.TrimSpace(currentStatement.String()); finalStmt != "" {
		statements = append(statements, finalStmt)
	}
	
	return statements
}
