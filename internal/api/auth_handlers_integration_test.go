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
	"sort"
	"strings"
	"testing"
	"time"

	"night-owls-go/internal/api"
	"night-owls-go/internal/auth"
	"night-owls-go/internal/config"
	db "night-owls-go/internal/db/sqlc_generated"

	// Removed: "night-owls-go/internal/logging" // logger is created locally
	"night-owls-go/internal/outbox"
	"night-owls-go/internal/service"

	"github.com/go-fuego/fuego"
	"github.com/nyaruka/phonenumbers"
	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert" // For MockMessageSender
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // Pure Go SQLite driver for database/sql
)

// testApp holds all components needed for integration testing the API.
type testApp struct {
	Server          *fuego.Server
	DB              *sql.DB
	Logger          *slog.Logger
	Config          *config.Config
	Querier         db.Querier
	UserService     *service.UserService
	ScheduleService *service.ScheduleService
	BookingService  *service.BookingService
	ReportService   *service.ReportService
	OutboxService   *outbox.DispatcherService
	Cron            *cron.Cron
}

func newTestApp(t *testing.T) *testApp {
	t.Helper()

	cfg := &config.Config{
		ServerPort:           "0",
		DatabasePath:         ":memory:",
		JWTSecret:            "test-jwt-secret",
		DefaultShiftDuration: 2 * time.Hour,
		OTPLogPath:           os.DevNull,
		LogLevel:             "debug", // Ensure debug level for this run
		LogFormat:            "text",  // Text format is easier for quick debug reading
		// Ensure other new config fields have defaults if not set by tests
		JWTExpirationHours: 24,
		OTPValidityMinutes: 5,
		OutboxBatchSize:    10,
		OutboxMaxRetries:   3,
	}

	// logger := slog.New(slog.NewTextHandler(io.Discard, nil)) // Old discarded logger
	// Use a logger that prints to stderr for debugging this test run
	loggerOpts := &slog.HandlerOptions{Level: slog.LevelDebug}
	logger := slog.New(slog.NewTextHandler(os.Stderr, loggerOpts))
	slog.SetDefault(logger) // Also set as default to catch any slog usage from other packages

	dbConn, err := sql.Open("sqlite", cfg.DatabasePath+"?cache=shared&_foreign_keys=on")
	require.NoError(t, err, "Failed to open in-memory DB for app")

	// Manually apply migrations by reading SQL files
	_, currentFilePath, _, ok := runtime.Caller(0)
	require.True(t, ok, "Failed to get current file path")
	currentDir := filepath.Dir(currentFilePath)
	migrationsDir := filepath.Join(currentDir, "..", "db", "migrations")

	migrationFiles, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	require.NoError(t, err, "Failed to glob migration files")
	require.NotEmpty(t, migrationFiles, "No migration files found")

	// CRITICAL: Sort migration files to ensure correct order
	sort.Strings(migrationFiles)

	// Apply migrations in a transaction for safety
	tx, err := dbConn.Begin()
	require.NoError(t, err, "Failed to begin migration transaction")
	defer tx.Rollback()

	for _, migrationFile := range migrationFiles {
		sqlBytes, err := os.ReadFile(migrationFile)
		require.NoError(t, err, fmt.Sprintf("Failed to read migration file: %s", migrationFile))

		// Robust SQL parsing - handle comments and multi-line statements
		sqlContent := string(sqlBytes)
		queries := parseSQLStatements(sqlContent)

		for i, query := range queries {
			trimmedQuery := strings.TrimSpace(query)
			if trimmedQuery == "" {
				continue
			}

			_, err = tx.Exec(trimmedQuery)
			require.NoError(t, err, fmt.Sprintf("Failed to execute migration query %d from %s: %s\nQuery: %s", i, migrationFile, err, trimmedQuery))
		}
		logger.Info("Applied migration", "file", filepath.Base(migrationFile))
	}

	err = tx.Commit()
	require.NoError(t, err, "Failed to commit migration transaction")
	logger.Info("Manually applied migrations for test DB.")

	err = dbConn.Ping() // Verify connection is still good
	require.NoError(t, err, "App's DB connection failed after manual migrations")

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

	// Create fuego server like in production
	server := fuego.NewServer()

	authAPIHandler := api.NewAuthHandler(userService, auditService, logger, cfg, querier, createTestSessionStore())
	scheduleAPIHandler := api.NewScheduleHandler(scheduleService, logger)
	bookingAPIHandler := api.NewBookingHandler(bookingService, auditService, querier, logger)
	reportAPIHandler := api.NewReportHandler(reportService, auditService, logger)
	leaderboardAPIHandler := api.NewLeaderboardHandler(pointsService, logger)

	// Public API routes (no auth required)
	fuego.PostStd(server, "/auth/register", authAPIHandler.RegisterHandler)
	fuego.PostStd(server, "/auth/verify", authAPIHandler.VerifyHandler)
	fuego.GetStd(server, "/auth/validate", authAPIHandler.ValidateHandler)
	fuego.GetStd(server, "/schedules", scheduleAPIHandler.ListSchedulesHandler)
	fuego.GetStd(server, "/shifts/available", scheduleAPIHandler.ListAvailableShiftsHandler)

	// Protected routes (require auth) - creating a group like in production
	protected := fuego.Group(server, "")
	fuego.Use(protected, api.AuthMiddleware(cfg, logger, createTestSessionStore()))
	fuego.PostStd(protected, "/bookings", bookingAPIHandler.CreateBookingHandler)
	fuego.GetStd(protected, "/bookings/my", bookingAPIHandler.GetMyBookingsHandler)
	fuego.PostStd(protected, "/bookings/{id}/checkin", bookingAPIHandler.MarkCheckInHandler)
	fuego.PostStd(protected, "/bookings/{id}/report", reportAPIHandler.CreateReportHandler)
	fuego.PostStd(protected, "/reports/off-shift", reportAPIHandler.CreateOffShiftReportHandler)
	fuego.GetStd(protected, "/user/reports", reportAPIHandler.ListReportsHandler)

	// Leaderboard routes
	fuego.GetStd(protected, "/leaderboard", leaderboardAPIHandler.GetLeaderboardHandler)
	fuego.GetStd(protected, "/leaderboard/shifts", leaderboardAPIHandler.GetStreakLeaderboardHandler)
	fuego.GetStd(protected, "/leaderboard/activity", leaderboardAPIHandler.GetActivityFeedHandler)
	fuego.GetStd(protected, "/user/stats", leaderboardAPIHandler.GetUserStatsHandler)
	fuego.GetStd(protected, "/user/points/history", leaderboardAPIHandler.GetUserPointsHistoryHandler)
	fuego.GetStd(protected, "/user/achievements", leaderboardAPIHandler.GetUserAchievementsHandler)
	fuego.GetStd(protected, "/user/achievements/available", leaderboardAPIHandler.GetAvailableAchievementsHandler)

	return &testApp{
		Server:          server,
		DB:              dbConn,
		Logger:          logger,
		Config:          cfg,
		Querier:         querier,
		UserService:     userService,
		ScheduleService: scheduleService,
		BookingService:  bookingService,
		ReportService:   reportService,
		OutboxService:   outboxService,
		Cron:            cronScheduler,
	}
}

func (app *testApp) makeRequest(t *testing.T, method, path string, body io.Reader, token string) *httptest.ResponseRecorder {
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
	app.Server.Mux.ServeHTTP(rr, req)
	return rr
}

func TestAuthEndpoints_RegisterAndVerify_Success(t *testing.T) {
	app := newTestApp(t)
	defer app.DB.Close()

	phone := "+442079460001" // Valid UK-style number
	name := "Integration User UK"

	registerPayload := api.RegisterRequest{Phone: phone, Name: name}
	payloadBytes, _ := json.Marshal(registerPayload)
	rr := app.makeRequest(t, "POST", "/auth/register", bytes.NewBuffer(payloadBytes), "")

	assert.Equal(t, http.StatusOK, rr.Code)
	var regResp api.RegisterResponse
	err := json.Unmarshal(rr.Body.Bytes(), &regResp)
	require.NoError(t, err)
	assert.Contains(t, regResp.Message, "Verification code sent")

	user, err := app.Querier.GetUserByPhone(context.Background(), phone)
	require.NoError(t, err)
	assert.Equal(t, phone, user.Phone)
	assert.Equal(t, name, user.Name.String)

	ctx := context.Background()
	outboxItems, err := app.Querier.GetPendingOutboxItems(ctx, 10)
	require.NoError(t, err)
	require.NotEmpty(t, outboxItems, "Expected an OTP outbox message")

	// We need to check against the E.164 version of the phone number
	parsedNumForOutboxCheck, _ := phonenumbers.Parse(phone, "GB") // Same parsing as handler
	e164PhoneForOutbox := phonenumbers.Format(parsedNumForOutboxCheck, phonenumbers.E164)

	var otpValue string
	foundInOutbox := false
	for _, item := range outboxItems {
		if item.Recipient == e164PhoneForOutbox && item.MessageType == "OTP_VERIFICATION" {
			var otpPayload struct {
				OTP string `json:"otp"`
			}
			err = json.Unmarshal([]byte(item.Payload.String), &otpPayload)
			require.NoError(t, err)
			otpValue = otpPayload.OTP
			foundInOutbox = true
			break
		}
	}
	require.True(t, foundInOutbox, "Could not find OTP for user %s in outbox", e164PhoneForOutbox)
	require.NotEmpty(t, otpValue, "OTP value is empty")

	verifyPayload := api.VerifyRequest{Phone: phone, Code: otpValue}
	payloadBytes, _ = json.Marshal(verifyPayload)
	rr = app.makeRequest(t, "POST", "/auth/verify", bytes.NewBuffer(payloadBytes), "")

	assert.Equal(t, http.StatusOK, rr.Code)
	var verifyResp api.VerifyResponse
	err = json.Unmarshal(rr.Body.Bytes(), &verifyResp)
	require.NoError(t, err)
	assert.NotEmpty(t, verifyResp.Token, "Expected a JWT token")

	claims, err := auth.ValidateJWT(verifyResp.Token, app.Config.JWTSecret)
	require.NoError(t, err)
	assert.Equal(t, user.UserID, claims.UserID)
	assert.Equal(t, user.Phone, claims.Phone)
}

func TestAuthEndpoints_Register_InvalidPayload(t *testing.T) {
	app := newTestApp(t)
	defer app.DB.Close()
	rr := app.makeRequest(t, "POST", "/auth/register", strings.NewReader("not a json"), "")
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAuthEndpoints_Verify_InvalidOTP(t *testing.T) {
	app := newTestApp(t)
	defer app.DB.Close()
	phone := "+14155550102"
	// Register first to store an OTP
	err := app.UserService.RegisterOrLoginUser(context.Background(), phone, sql.NullString{String: "TestUser", Valid: true}, "test-ip", "test-agent")
	require.NoError(t, err)

	verifyPayload := api.VerifyRequest{Phone: phone, Code: "000000"}
	payloadBytes, _ := json.Marshal(verifyPayload)
	rr := app.makeRequest(t, "POST", "/auth/verify", bytes.NewBuffer(payloadBytes), "")
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthEndpoints_ValidateToken_Success(t *testing.T) {
	app := newTestApp(t)
	defer app.DB.Close()
	phone := "+14155550105"
	userName := "Validate Test User"
	ctx := context.Background()

	// Register and verify user to get a valid token
	err := app.UserService.RegisterOrLoginUser(ctx, phone,
		sql.NullString{String: userName, Valid: true}, "test-ip", "test-agent")
	require.NoError(t, err)

	outboxItems, err := app.Querier.GetPendingOutboxItems(ctx, 10)
	require.NoError(t, err)
	var otpValue string
	for _, item := range outboxItems {
		if item.Recipient == phone && item.MessageType == "OTP_VERIFICATION" {
			var otpPayload struct {
				OTP string `json:"otp"`
			}
			err = json.Unmarshal([]byte(item.Payload.String), &otpPayload)
			require.NoError(t, err)
			otpValue = otpPayload.OTP
			break
		}
	}
	require.NotEmpty(t, otpValue)

	token, err := app.UserService.VerifyOTP(ctx, phone, otpValue, "test-ip", "test-agent")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Test the validate endpoint with valid token
	rr := app.makeRequest(t, "GET", "/auth/validate", nil, token)
	assert.Equal(t, http.StatusOK, rr.Code)

	var validateResp map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &validateResp)
	require.NoError(t, err)

	assert.NotEmpty(t, validateResp["id"])
	assert.Equal(t, phone, validateResp["phone"])
	assert.Equal(t, userName, validateResp["name"])
	assert.Equal(t, "admin", validateResp["role"]) // First user gets admin role
}

func TestAuthEndpoints_ValidateToken_NoToken(t *testing.T) {
	app := newTestApp(t)
	defer app.DB.Close()

	rr := app.makeRequest(t, "GET", "/auth/validate", nil, "")
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthEndpoints_ValidateToken_InvalidToken(t *testing.T) {
	app := newTestApp(t)
	defer app.DB.Close()

	rr := app.makeRequest(t, "GET", "/auth/validate", nil, "invalid-token")
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// TODO: More integration tests here
