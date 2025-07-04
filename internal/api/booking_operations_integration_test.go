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
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

type bookingTestApp struct {
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

func newBookingTestApp(t *testing.T) *bookingTestApp {
	t.Helper()

	cfg := &config.Config{
		ServerPort:           "0",
		DatabasePath:         ":memory:",
		JWTSecret:            "test-jwt-secret-booking",
		DefaultShiftDuration: 2 * time.Hour,
		OTPLogPath:           os.DevNull,
		LogLevel:             "debug",
		LogFormat:            "text",
		JWTExpirationHours:   1,
		OTPValidityMinutes:   5,
		OutboxBatchSize:      10,
		OutboxMaxRetries:     3,
		VAPIDPublic:          "test_public_key",
		VAPIDPrivate:         "test_private_key",
		VAPIDSubject:         "mailto:test@example.com",
	}

	loggerOpts := &slog.HandlerOptions{Level: slog.LevelDebug}
	logger := slog.New(slog.NewTextHandler(io.Discard, loggerOpts))

	dbConn, err := sql.Open("sqlite", cfg.DatabasePath+"?cache=shared&_foreign_keys=on")
	require.NoError(t, err, "Failed to open in-memory DB for booking app")

	// Apply migrations
	_, currentFilePath, _, ok := runtime.Caller(0)
	require.True(t, ok, "Failed to get current file path")
	currentDir := filepath.Dir(currentFilePath)
	migrationsDir := filepath.Join(currentDir, "..", "db", "migrations")

	migrationFiles, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	require.NoError(t, err, "Failed to glob migration files")
	require.NotEmpty(t, migrationFiles, "No migration files found")

	for _, migrationFile := range migrationFiles {
		sqlBytes, err := os.ReadFile(migrationFile)
		require.NoError(t, err, fmt.Sprintf("Failed to read migration file: %s", migrationFile))

		sqlContent := strings.TrimSpace(string(sqlBytes))
		if sqlContent == "" {
			continue
		}
		_, err = dbConn.Exec(sqlContent)
		if err != nil {
			queries := strings.Split(sqlContent, ";\n")
			for i, query := range queries {
				trimmedQuery := strings.TrimSpace(query)
				if trimmedQuery == "" || strings.HasPrefix(trimmedQuery, "--") {
					continue
				}
				if i == len(queries)-1 && !strings.HasSuffix(trimmedQuery, ";") {
					trimmedQuery += ";"
				}
				_, err = dbConn.Exec(trimmedQuery)
				require.NoError(t, err, fmt.Sprintf("Failed to execute migration query from %s: %s", migrationFile, trimmedQuery))
			}
		}
	}

	err = dbConn.Ping()
	require.NoError(t, err, "Booking app's DB connection failed after migrations")

	querier := db.New(dbConn)
	otpStore := auth.NewInMemoryOTPStore()
	mockSender := new(MockMessageSender)

	userService := service.NewUserService(querier, otpStore, cfg, logger)
	scheduleService := service.NewScheduleService(querier, logger, cfg)
	pointsService := service.NewPointsService(querier, logger)
	bookingService := service.NewBookingService(querier, cfg, logger, pointsService)
	reportService := service.NewReportService(querier, logger, pointsService)
	auditService := service.NewAuditService(querier, logger)
	pushService := service.NewPushSender(querier, cfg, logger)
	outboxService := outbox.NewDispatcherService(querier, mockSender, pushService, logger, cfg)

	cronScheduler := cron.New()

	router := chi.NewRouter()
	router.Use(chiMiddleware.Recoverer)

	// Register handlers
	authAPIHandler := api.NewAuthHandler(userService, auditService, logger, cfg, querier, createTestSessionStore())
	bookingAPIHandler := api.NewBookingHandler(bookingService, auditService, querier, logger)
	adminBookingAPIHandler := api.NewAdminBookingHandler(bookingService, logger)

	// Public routes
	router.Post("/auth/register", authAPIHandler.RegisterHandler)
	router.Post("/auth/verify", authAPIHandler.VerifyHandler)

	// Protected user routes
	router.Group(func(r chi.Router) {
		r.Use(api.AuthMiddleware(cfg, logger, createTestSessionStore()))
		r.Post("/bookings", bookingAPIHandler.CreateBookingHandler)
		r.Get("/bookings/my", bookingAPIHandler.GetMyBookingsHandler)
		r.Post("/bookings/{id}/checkin", bookingAPIHandler.MarkCheckInHandler)
	})

	// Admin routes
	router.Route("/api/admin", func(r chi.Router) {
		r.Use(api.AuthMiddleware(cfg, logger, createTestSessionStore()))
		r.Use(api.AdminMiddleware(logger))
		r.Route("/bookings", func(br chi.Router) {
			br.Post("/assign", adminBookingAPIHandler.AssignUserToShiftHandler)
		})
	})

	return &bookingTestApp{
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

func (app *bookingTestApp) makeRequest(t *testing.T, method, path string, body io.Reader, token string) *httptest.ResponseRecorder {
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

func (app *bookingTestApp) createTestUserAndLogin(t *testing.T, phone, name, role string) (db.CreateUserRow, string) {
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

func (app *bookingTestApp) createTestSchedule(t *testing.T) db.Schedule {
	t.Helper()
	ctx := context.Background()
	schedule, err := app.Querier.CreateSchedule(ctx, db.CreateScheduleParams{
		Name:            "Test Night Watch",
		CronExpr:        "0 22 * * *", // Every day at 10 PM
		DurationMinutes: 120,          // 2 hours
		Timezone:        sql.NullString{String: "UTC", Valid: true},
	})
	require.NoError(t, err)
	return schedule
}

func TestBookingOperations_CreateBooking_Success(t *testing.T) {
	app := newBookingTestApp(t)
	defer app.DB.Close()

	user, token := app.createTestUserAndLogin(t, "+15550001001", "Test User", "owl")
	schedule := app.createTestSchedule(t)

	// Calculate next shift time (for future reference, but not used in new API)
	locUTC, _ := time.LoadLocation("UTC")
	tomorrow := time.Now().In(locUTC).AddDate(0, 0, 1)
	startOfTomorrow := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, locUTC)

	parsedCron, err := cronexpr.Parse(schedule.CronExpr)
	require.NoError(t, err)
	shiftStartTime := parsedCron.Next(startOfTomorrow)

	createRequest := api.CreateBookingRequest{
		ScheduleID: schedule.ScheduleID,
		StartTime:  shiftStartTime,
		BuddyName:  &user.Name.String,
	}
	payloadBytes, _ := json.Marshal(createRequest)

	rr := app.makeRequest(t, "POST", "/bookings", bytes.NewBuffer(payloadBytes), token)

	require.Equal(t, http.StatusCreated, rr.Code, "Response: %s", rr.Body.String())

	var bookingResp api.BookingResponse
	err = json.Unmarshal(rr.Body.Bytes(), &bookingResp)
	require.NoError(t, err)

	assert.Equal(t, user.UserID, bookingResp.UserID)
	assert.Equal(t, schedule.ScheduleID, bookingResp.ScheduleID)
	assert.True(t, shiftStartTime.Equal(bookingResp.ShiftStart.In(locUTC)))
	assert.Nil(t, bookingResp.CheckedInAt)

	// Verify booking exists in database
	ctx := context.Background()
	dbBooking, err := app.Querier.GetBookingByID(ctx, bookingResp.BookingID)
	require.NoError(t, err)
	assert.Equal(t, user.UserID, dbBooking.UserID)
	assert.Equal(t, schedule.ScheduleID, dbBooking.ScheduleID)
}

func TestBookingOperations_CreateBooking_DuplicateSlot(t *testing.T) {
	app := newBookingTestApp(t)
	defer app.DB.Close()

	user1, token1 := app.createTestUserAndLogin(t, "+15550001001", "User One", "owl")
	_, token2 := app.createTestUserAndLogin(t, "+15550001002", "User Two", "owl")
	schedule := app.createTestSchedule(t)

	// Calculate next shift time (not used but kept for future use)
	locUTC, _ := time.LoadLocation("UTC")
	tomorrow := time.Now().In(locUTC).AddDate(0, 0, 1)
	startOfTomorrow := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, locUTC)

	parsedCron, err := cronexpr.Parse(schedule.CronExpr)
	require.NoError(t, err)
	shiftStartTime := parsedCron.Next(startOfTomorrow)

	createRequest := api.CreateBookingRequest{
		ScheduleID: schedule.ScheduleID,
		StartTime:  shiftStartTime,
		BuddyName:  &user1.Name.String,
	}
	payloadBytes, _ := json.Marshal(createRequest)

	// First user books successfully
	rr1 := app.makeRequest(t, "POST", "/bookings", bytes.NewBuffer(payloadBytes), token1)
	require.Equal(t, http.StatusCreated, rr1.Code)

	// Second user tries to book same slot - should fail
	rr2 := app.makeRequest(t, "POST", "/bookings", bytes.NewBuffer(payloadBytes), token2)
	assert.Equal(t, http.StatusConflict, rr2.Code)

	var errorResp api.ErrorResponse
	err = json.Unmarshal(rr2.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Contains(t, errorResp.Error, "already booked")
}

func TestBookingOperations_GetMyBookings_Success(t *testing.T) {
	app := newBookingTestApp(t)
	defer app.DB.Close()

	user, token := app.createTestUserAndLogin(t, "+15550001001", "Test User", "owl")
	schedule := app.createTestSchedule(t)

	// Create a booking first
	ctx := context.Background()
	locUTC, _ := time.LoadLocation("UTC")
	shiftStart := time.Now().In(locUTC).Add(24 * time.Hour)
	shiftEnd := shiftStart.Add(2 * time.Hour)

	booking, err := app.Querier.CreateBooking(ctx, db.CreateBookingParams{
		UserID:     user.UserID,
		ScheduleID: schedule.ScheduleID,
		ShiftStart: shiftStart,
		ShiftEnd:   shiftEnd,
	})
	require.NoError(t, err)

	// Get user's bookings
	rr := app.makeRequest(t, "GET", "/bookings/my", nil, token)
	require.Equal(t, http.StatusOK, rr.Code)

	var bookings []api.BookingWithScheduleResponse
	err = json.Unmarshal(rr.Body.Bytes(), &bookings)
	require.NoError(t, err)

	require.Len(t, bookings, 1)
	assert.Equal(t, booking.BookingID, bookings[0].BookingID)
	assert.Equal(t, user.UserID, bookings[0].UserID)
	assert.Equal(t, schedule.ScheduleID, bookings[0].ScheduleID)
}

func TestBookingOperations_CheckIn_Success(t *testing.T) {
	app := newBookingTestApp(t)
	defer app.DB.Close()

	user, token := app.createTestUserAndLogin(t, "+15550001001", "Test User", "owl")
	schedule := app.createTestSchedule(t)

	// Create a booking for current time (eligible for check-in)
	ctx := context.Background()
	locUTC, _ := time.LoadLocation("UTC")
	shiftStart := time.Now().In(locUTC).Add(-30 * time.Minute) // Started 30 minutes ago
	shiftEnd := shiftStart.Add(2 * time.Hour)

	booking, err := app.Querier.CreateBooking(ctx, db.CreateBookingParams{
		UserID:     user.UserID,
		ScheduleID: schedule.ScheduleID,
		ShiftStart: shiftStart,
		ShiftEnd:   shiftEnd,
	})
	require.NoError(t, err)

	rr := app.makeRequest(t, "POST", fmt.Sprintf("/bookings/%d/checkin", booking.BookingID), nil, token)
	require.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())

	var bookingResp api.BookingResponse
	err = json.Unmarshal(rr.Body.Bytes(), &bookingResp)
	require.NoError(t, err)

	assert.NotNil(t, bookingResp.CheckedInAt)

	// Verify in database
	dbBooking, err := app.Querier.GetBookingByID(ctx, booking.BookingID)
	require.NoError(t, err)
	assert.True(t, dbBooking.CheckedInAt.Valid)
}

func TestBookingOperations_CheckIn_TooEarly(t *testing.T) {
	app := newBookingTestApp(t)
	defer app.DB.Close()

	user, token := app.createTestUserAndLogin(t, "+15550001001", "Test User", "owl")
	schedule := app.createTestSchedule(t)

	// Create a booking for future time (not eligible for check-in yet)
	ctx := context.Background()
	locUTC, _ := time.LoadLocation("UTC")
	shiftStart := time.Now().In(locUTC).Add(2 * time.Hour) // Starts in 2 hours
	shiftEnd := shiftStart.Add(2 * time.Hour)

	booking, err := app.Querier.CreateBooking(ctx, db.CreateBookingParams{
		UserID:     user.UserID,
		ScheduleID: schedule.ScheduleID,
		ShiftStart: shiftStart,
		ShiftEnd:   shiftEnd,
	})
	require.NoError(t, err)

	rr := app.makeRequest(t, "POST", fmt.Sprintf("/bookings/%d/checkin", booking.BookingID), nil, token)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errorResp api.ErrorResponse
	err = json.Unmarshal(rr.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Contains(t, errorResp.Error, "too early")
}

func TestBookingOperations_Unauthorized_NoToken(t *testing.T) {
	app := newBookingTestApp(t)
	defer app.DB.Close()

	// Test all protected endpoints without token
	endpoints := []struct {
		method string
		path   string
	}{
		{"POST", "/bookings"},
		{"GET", "/bookings/my"},
		{"POST", "/bookings/1/checkin"},
	}

	for _, endpoint := range endpoints {
		rr := app.makeRequest(t, endpoint.method, endpoint.path, nil, "")
		assert.Equal(t, http.StatusUnauthorized, rr.Code, "Endpoint %s %s should require authentication", endpoint.method, endpoint.path)
	}
}

func TestBookingOperations_Unauthorized_WrongUser(t *testing.T) {
	app := newBookingTestApp(t)
	defer app.DB.Close()

	user1, _ := app.createTestUserAndLogin(t, "+15550001001", "User One", "owl")
	_, token2 := app.createTestUserAndLogin(t, "+15550001002", "User Two", "owl")
	schedule := app.createTestSchedule(t)

	// User1 creates a booking
	ctx := context.Background()
	locUTC, _ := time.LoadLocation("UTC")
	shiftStart := time.Now().In(locUTC).Add(24 * time.Hour)
	shiftEnd := shiftStart.Add(2 * time.Hour)

	booking, err := app.Querier.CreateBooking(ctx, db.CreateBookingParams{
		UserID:     user1.UserID,
		ScheduleID: schedule.ScheduleID,
		ShiftStart: shiftStart,
		ShiftEnd:   shiftEnd,
	})
	require.NoError(t, err)

	// User2 tries to check in to User1's booking
	rr := app.makeRequest(t, "POST", fmt.Sprintf("/bookings/%d/checkin", booking.BookingID), nil, token2)
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestBookingOperations_InvalidBookingID(t *testing.T) {
	app := newBookingTestApp(t)
	defer app.DB.Close()

	_, token := app.createTestUserAndLogin(t, "+15550001001", "Test User", "owl")

	// Test with non-existent booking ID
	rr := app.makeRequest(t, "POST", "/bookings/99999/checkin", nil, token)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestBookingOperations_InvalidRequestPayload(t *testing.T) {
	app := newBookingTestApp(t)
	defer app.DB.Close()

	_, token := app.createTestUserAndLogin(t, "+15550001001", "Test User", "owl")

	// Test with invalid JSON
	invalidJSON := bytes.NewBufferString(`{"invalid": json}`)
	rr := app.makeRequest(t, "POST", "/bookings", invalidJSON, token)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// Test with missing required fields
	incompleteRequest := bytes.NewBufferString(`{"schedule_id": 1}`) // Missing start_time
	rr = app.makeRequest(t, "POST", "/bookings", incompleteRequest, token)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
