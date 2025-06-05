package api_test

import (
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
	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

type leaderboardTestApp struct {
	Router         *chi.Mux
	DB             *sql.DB
	Logger         *slog.Logger
	Config         *config.Config
	Querier        db.Querier
	UserService    *service.UserService
	PointsService  *service.PointsService
	BookingService *service.BookingService
	ReportService  *service.ReportService
	OutboxService  *outbox.DispatcherService
	PushService    *service.PushSender
	mockSMSSender  *MockMessageSender
	Cron           *cron.Cron
	OTPStore       auth.OTPStore
}

func newLeaderboardTestApp(t *testing.T) *leaderboardTestApp {
	t.Helper()

	cfg := &config.Config{
		ServerPort:           "0",
		DatabasePath:         ":memory:",
		JWTSecret:            "test-jwt-secret-leaderboard",
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
	require.NoError(t, err, "Failed to open in-memory DB for leaderboard app")

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
	require.NoError(t, err, "Leaderboard app's DB connection failed after migrations")

	querier := db.New(dbConn)
	otpStore := auth.NewInMemoryOTPStore()
	mockSender := new(MockMessageSender)

	userService := service.NewUserService(querier, otpStore, cfg, logger)
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
	authAPIHandler := api.NewAuthHandler(userService, auditService, logger, cfg, querier)
	leaderboardAPIHandler := api.NewLeaderboardHandler(pointsService, logger)

	// Public routes
	router.Post("/auth/register", authAPIHandler.RegisterHandler)
	router.Post("/auth/verify", authAPIHandler.VerifyHandler)

	// Protected leaderboard routes
	router.Group(func(r chi.Router) {
		r.Use(api.AuthMiddleware(cfg, logger))
		r.Get("/api/leaderboard", leaderboardAPIHandler.GetLeaderboardHandler)
		r.Get("/api/leaderboard/shifts", leaderboardAPIHandler.GetStreakLeaderboardHandler)
		r.Get("/api/leaderboard/activity", leaderboardAPIHandler.GetActivityFeedHandler)
		r.Get("/api/user/stats", leaderboardAPIHandler.GetUserStatsHandler)
		r.Get("/api/user/points/history", leaderboardAPIHandler.GetUserPointsHistoryHandler)
		r.Get("/api/user/achievements", leaderboardAPIHandler.GetUserAchievementsHandler)
		r.Get("/api/user/achievements/available", leaderboardAPIHandler.GetAvailableAchievementsHandler)
	})

	return &leaderboardTestApp{
		Router:         router,
		DB:             dbConn,
		Logger:         logger,
		Config:         cfg,
		Querier:        querier,
		UserService:    userService,
		PointsService:  pointsService,
		BookingService: bookingService,
		ReportService:  reportService,
		PushService:    pushService,
		OutboxService:  outboxService,
		mockSMSSender:  mockSender,
		Cron:           cronScheduler,
		OTPStore:       otpStore,
	}
}

func (app *leaderboardTestApp) makeRequest(t *testing.T, method, path string, body io.Reader, token string) *httptest.ResponseRecorder {
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

func (app *leaderboardTestApp) createTestUserAndLogin(t *testing.T, phone, name, role string) (db.CreateUserRow, string) {
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

func TestLeaderboardEndpoints_GetLeaderboard_Success(t *testing.T) {
	app := newLeaderboardTestApp(t)
	defer app.DB.Close()

	// Create test users
	user1, token1 := app.createTestUserAndLogin(t, "+15550001001", "Alice", "owl")
	user2, _ := app.createTestUserAndLogin(t, "+15550001002", "Bob", "owl")

	t.Logf("Created test users: %v, %v", user1.UserID, user2.UserID)

	// Test leaderboard endpoint with authentication
	rr := app.makeRequest(t, "GET", "/api/leaderboard", nil, token1)
	require.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())

	// Verify response structure
	var leaderboard []api.LeaderboardEntry
	err := json.Unmarshal(rr.Body.Bytes(), &leaderboard)
	require.NoError(t, err, "Failed to parse leaderboard response")

	// Should return empty leaderboard initially, but with proper structure
	assert.IsType(t, []api.LeaderboardEntry{}, leaderboard)
	t.Logf("Leaderboard response: %+v", leaderboard)
}

func TestLeaderboardEndpoints_GetShiftsLeaderboard_Success(t *testing.T) {
	app := newLeaderboardTestApp(t)
	defer app.DB.Close()

	user, token := app.createTestUserAndLogin(t, "+15550001001", "Test User", "owl")
	t.Logf("Created test user: %v", user.UserID)

	rr := app.makeRequest(t, "GET", "/api/leaderboard/shifts", nil, token)
	require.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())

	var shiftsLeaderboard []api.LeaderboardEntry
	err := json.Unmarshal(rr.Body.Bytes(), &shiftsLeaderboard)
	require.NoError(t, err)

	assert.IsType(t, []api.LeaderboardEntry{}, shiftsLeaderboard)
}

func TestLeaderboardEndpoints_GetUserStats_Success(t *testing.T) {
	app := newLeaderboardTestApp(t)
	defer app.DB.Close()

	user, token := app.createTestUserAndLogin(t, "+15550001001", "Test User", "owl")
	t.Logf("Created test user: %v", user.UserID)

	rr := app.makeRequest(t, "GET", "/api/user/stats", nil, token)
	require.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())

	var userStats api.UserStatsResponse
	err := json.Unmarshal(rr.Body.Bytes(), &userStats)
	require.NoError(t, err)

	// Verify initial stats structure
	assert.GreaterOrEqual(t, userStats.TotalPoints, int64(0))
	assert.GreaterOrEqual(t, userStats.ShiftCount, int64(0))
	assert.GreaterOrEqual(t, userStats.Rank, int64(0))
	assert.GreaterOrEqual(t, userStats.UserID, int64(1))
}

func TestLeaderboardEndpoints_GetUserPointsHistory_Success(t *testing.T) {
	app := newLeaderboardTestApp(t)
	defer app.DB.Close()

	user, token := app.createTestUserAndLogin(t, "+15550001001", "Test User", "owl")
	t.Logf("Created test user: %v", user.UserID)

	rr := app.makeRequest(t, "GET", "/api/user/points/history?limit=10", nil, token)
	require.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())

	var pointsHistory []api.PointsHistoryEntry
	err := json.Unmarshal(rr.Body.Bytes(), &pointsHistory)
	require.NoError(t, err)

	assert.IsType(t, []api.PointsHistoryEntry{}, pointsHistory)
}

func TestLeaderboardEndpoints_GetUserAchievements_Success(t *testing.T) {
	app := newLeaderboardTestApp(t)
	defer app.DB.Close()

	user, token := app.createTestUserAndLogin(t, "+15550001001", "Test User", "owl")
	t.Logf("Created test user: %v", user.UserID)

	rr := app.makeRequest(t, "GET", "/api/user/achievements", nil, token)
	require.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())

	var achievements []api.AchievementResponse
	err := json.Unmarshal(rr.Body.Bytes(), &achievements)
	require.NoError(t, err)

	assert.IsType(t, []api.AchievementResponse{}, achievements)
}

func TestLeaderboardEndpoints_GetAvailableAchievements_Success(t *testing.T) {
	app := newLeaderboardTestApp(t)
	defer app.DB.Close()

	user, token := app.createTestUserAndLogin(t, "+15550001001", "Test User", "owl")
	t.Logf("Created test user: %v", user.UserID)

	rr := app.makeRequest(t, "GET", "/api/user/achievements/available", nil, token)
	require.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())

	var availableAchievements []api.AchievementResponse
	err := json.Unmarshal(rr.Body.Bytes(), &availableAchievements)
	require.NoError(t, err)

	assert.IsType(t, []api.AchievementResponse{}, availableAchievements)
}

func TestLeaderboardEndpoints_GetActivityFeed_Success(t *testing.T) {
	app := newLeaderboardTestApp(t)
	defer app.DB.Close()

	user, token := app.createTestUserAndLogin(t, "+15550001001", "Test User", "owl")
	t.Logf("Created test user: %v", user.UserID)

	rr := app.makeRequest(t, "GET", "/api/leaderboard/activity", nil, token)
	require.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())

	var activityFeed []api.ActivityFeedEntry
	err := json.Unmarshal(rr.Body.Bytes(), &activityFeed)
	require.NoError(t, err)

	assert.IsType(t, []api.ActivityFeedEntry{}, activityFeed)
}

func TestLeaderboardEndpoints_RequireAuthentication(t *testing.T) {
	app := newLeaderboardTestApp(t)
	defer app.DB.Close()

	// Test all leaderboard endpoints without authentication
	endpoints := []string{
		"/api/leaderboard",
		"/api/leaderboard/shifts",
		"/api/leaderboard/activity",
		"/api/user/stats",
		"/api/user/points/history",
		"/api/user/achievements",
		"/api/user/achievements/available",
	}

	for _, endpoint := range endpoints {
		rr := app.makeRequest(t, "GET", endpoint, nil, "")
		assert.Equal(t, http.StatusUnauthorized, rr.Code, "Endpoint %s should require authentication", endpoint)
	}
}

func TestLeaderboardEndpoints_InvalidToken(t *testing.T) {
	app := newLeaderboardTestApp(t)
	defer app.DB.Close()

	endpoints := []string{
		"/api/leaderboard",
		"/api/user/stats",
	}

	for _, endpoint := range endpoints {
		rr := app.makeRequest(t, "GET", endpoint, nil, "invalid-token")
		assert.Equal(t, http.StatusUnauthorized, rr.Code, "Endpoint %s should reject invalid tokens", endpoint)
	}
}
