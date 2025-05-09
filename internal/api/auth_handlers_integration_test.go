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

	// Removed: "night-owls-go/internal/logging" // logger is created locally
	"night-owls-go/internal/outbox"
	"night-owls-go/internal/service"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock" // For MockMessageSender
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // Pure Go SQLite driver for database/sql
)

// MockMessageSender for integration tests (local to this package test)
type MockMessageSender struct {
	mock.Mock
}

func (m *MockMessageSender) Send(recipient, messageType, payload string) error {
	args := m.Called(recipient, messageType, payload)
	return args.Error(0)
}


// testApp holds all components needed for integration testing the API.
type testApp struct {
	Router *chi.Mux
	DB     *sql.DB 
	Logger *slog.Logger
	Config *config.Config
    Querier db.Querier 
	UserService    *service.UserService
	ScheduleService *service.ScheduleService
	BookingService  *service.BookingService
	ReportService   *service.ReportService
	OutboxService   *outbox.DispatcherService
	mockSMSSender *MockMessageSender // Use local mock
	Cron          *cron.Cron
}

func newTestApp(t *testing.T) *testApp {
	t.Helper()

	cfg := &config.Config{
		ServerPort:         "0", 
		DatabasePath:       ":memory:", 
		JWTSecret:          "test-jwt-secret",
		DefaultShiftDuration: 2 * time.Hour,
		OTPLogPath:         os.DevNull, 
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	dbConn, err := sql.Open("sqlite", cfg.DatabasePath+"?cache=shared&_foreign_keys=on")
	require.NoError(t, err, "Failed to open in-memory DB for app")

	// Manually apply migrations by reading SQL files
	migrationFiles := []string{
		"../db/migrations/000001_init_schema.up.sql",
		"../db/migrations/000002_seed_schedules.up.sql",
	}
	for _, migrationFile := range migrationFiles {
		// Construct path relative to this test file's directory (internal/api)
		// This assumes the test CWD is the package directory.
		// For robustness, one might determine project root and build absolute path.
		// For simplicity, using direct relative path.
		// Path needs to be from `auth_handlers_integration_test.go` to the migration files.
		// If test file is in `internal/api`, and migrations in `internal/db/migrations`,
		// then path is `../db/migrations/filename.sql`.

		// Let's get the directory of the current test file to make relative paths more robust.
		_, currentFilePath, _, ok := runtime.Caller(0)
        require.True(t, ok, "Failed to get current file path")
        currentDir := filepath.Dir(currentFilePath)

		absMigrationPath := filepath.Join(currentDir, migrationFile) 
		sqlBytes, err := os.ReadFile(absMigrationPath)
		require.NoError(t, err, fmt.Sprintf("Failed to read migration file: %s (abs: %s)", migrationFile, absMigrationPath))
		
		// Split SQL statements if the file contains multiple (SQLite often processes one by one)
        // For these simple .up.sql files, they are usually fine to execute as a whole if no GO batch separators
        // or we can split by ";\n"
        queries := strings.Split(string(sqlBytes), ";\n")
        for _, query := range queries {
            trimmedQuery := strings.TrimSpace(query)
            if trimmedQuery == "" {
                continue
            }
		    _, err = dbConn.Exec(trimmedQuery)
		    require.NoError(t, err, fmt.Sprintf("Failed to execute migration query from %s: %s", migrationFile, trimmedQuery))
        }
	}
	logger.Info("Manually applied migrations for test DB.")

    err = dbConn.Ping() // Verify connection is still good
    require.NoError(t, err, "App's DB connection failed after manual migrations")

	querier := db.New(dbConn)
	otpStore := auth.NewInMemoryOTPStore()
    mockSender := new(MockMessageSender) 

	userService := service.NewUserService(querier, otpStore, cfg, logger)
	scheduleService := service.NewScheduleService(querier, logger)
	bookingService := service.NewBookingService(querier, cfg, logger)
	reportService := service.NewReportService(querier, logger)
	outboxService := outbox.NewDispatcherService(querier, mockSender, logger)

	cronScheduler := cron.New()

	router := chi.NewRouter()
	router.Use(chiMiddleware.Recoverer) 

	authAPIHandler := api.NewAuthHandler(userService, logger)
	scheduleAPIHandler := api.NewScheduleHandler(scheduleService, logger)
	bookingAPIHandler := api.NewBookingHandler(bookingService, logger)
	reportAPIHandler := api.NewReportHandler(reportService, logger)

	router.Post("/auth/register", authAPIHandler.RegisterHandler)
	router.Post("/auth/verify", authAPIHandler.VerifyHandler)
	router.Get("/schedules", scheduleAPIHandler.ListSchedulesHandler)
	router.Get("/shifts/available", scheduleAPIHandler.ListAvailableShiftsHandler)
	router.Group(func(r chi.Router) {
		r.Use(api.AuthMiddleware(cfg, logger))
		r.Post("/bookings", bookingAPIHandler.CreateBookingHandler)
		r.Patch("/bookings/{id}/attendance", bookingAPIHandler.MarkAttendanceHandler)
		r.Post("/bookings/{id}/report", reportAPIHandler.CreateReportHandler)
	})

	return &testApp{
		Router: router,
		DB:     dbConn,
		Logger: logger,
		Config: cfg,
        Querier: querier,
		UserService:    userService,
		ScheduleService: scheduleService,
		BookingService:  bookingService,
		ReportService:   reportService,
		OutboxService:   outboxService,
		mockSMSSender: mockSender,
		Cron: cronScheduler,
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
	app.Router.ServeHTTP(rr, req)
	return rr
}

func TestAuthEndpoints_RegisterAndVerify_Success(t *testing.T) {
	app := newTestApp(t)
	defer app.DB.Close()

	phone := "+1234567001"
	name := "Integration User"

	registerPayload := api.RegisterRequest{Phone: phone, Name: name}
	payloadBytes, _ := json.Marshal(registerPayload)
	rr := app.makeRequest(t, "POST", "/auth/register", bytes.NewBuffer(payloadBytes), "")

	assert.Equal(t, http.StatusOK, rr.Code)
	var regResp api.RegisterResponse
	err := json.Unmarshal(rr.Body.Bytes(), &regResp)
	require.NoError(t, err)
	assert.Contains(t, regResp.Message, "OTP sent")

	user, err := app.Querier.GetUserByPhone(context.Background(), phone)
	require.NoError(t, err)
	assert.Equal(t, phone, user.Phone)
	assert.Equal(t, name, user.Name.String)

    ctx := context.Background()
    outboxItems, err := app.Querier.GetPendingOutboxItems(ctx, 10)
    require.NoError(t, err)
    require.NotEmpty(t, outboxItems, "Expected an OTP outbox message")
    
    var otpValue string
    for _, item := range outboxItems {
        if item.Recipient == phone && item.MessageType == "OTP_VERIFICATION" {
            var otpPayload struct { OTP string `json:"otp"` }
            err = json.Unmarshal([]byte(item.Payload.String), &otpPayload)
            require.NoError(t, err)
            otpValue = otpPayload.OTP
            break
        }
    }
    require.NotEmpty(t, otpValue, "Could not find OTP for user in outbox")

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
	phone := "+1234567002"
	// Register first to store an OTP
	err := app.UserService.RegisterOrLoginUser(context.Background(), phone, sql.NullString{String:"TestUser", Valid:true})
	require.NoError(t, err) 

	verifyPayload := api.VerifyRequest{Phone: phone, Code: "000000"} 
	payloadBytes, _ := json.Marshal(verifyPayload)
	rr := app.makeRequest(t, "POST", "/auth/verify", bytes.NewBuffer(payloadBytes), "")
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// TODO: More integration tests here 