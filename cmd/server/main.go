// Package main is the entry point for the Night Owls Control API server
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"night-owls-go/internal/api"
	"night-owls-go/internal/auth"
	"night-owls-go/internal/config"
	db "night-owls-go/internal/db/sqlc_generated"
	"night-owls-go/internal/logging"
	"night-owls-go/internal/outbox"
	"night-owls-go/internal/service"

	"github.com/golang-migrate/migrate/v4"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	httpSwagger "github.com/swaggo/http-swagger"

	// Import the generated swagger docs when available
	// _ "night-owls-go/docs/swagger"

	"github.com/go-fuego/fuego"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

// @title Night Owls Control Shift Scheduler API
// @version 1.0
// @description API for managing community watch shifts, bookings, and reports
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:5888
// @BasePath /
// @schemes http

// slogCronLogger is an adapter to use slog.Logger with cron.PrintfLogger.
type slogCronLogger struct {
	logger *slog.Logger
}

// Printf implements the cron.Logger interface.
func (scl *slogCronLogger) Printf(format string, args ...interface{}) {
	// Note: This simplistic Printf won't parse format string for structured key-value pairs.
	// It will log the entire formatted string as a message.
	// For more structured cron logs, one might need a more sophisticated adapter or a custom cron logger.
	scl.logger.Info(fmt.Sprintf(format, args...))
}

func main() {
	startTime := time.Now() // Track server start time for health check uptime

	// API route constants
	const apiPrefix = "/api"

	// Force timezone to UTC
	if err := os.Setenv("TZ", "UTC"); err != nil {
		log.Printf("Warning: Failed to set timezone to UTC: %v", err)
	}

	// Use Overload to force .env file values to override any existing environment variables
	err := godotenv.Overload()
	if err != nil {
		// Log this initial finding using standard log before slog is fully set up
		log.Println("No .env file found, using environment variables or defaults")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Critical: Error loading configuration: %v", err)
	}

	// SECURITY: Validate critical security configurations before starting
	if err := cfg.ValidateSecurityConfig(); err != nil {
		log.Fatalf("Critical: %v", err)
	}

	logger := logging.NewLogger(cfg) // Initialize logger with config
	slog.SetDefault(logger)          // Set as global default

	slog.Info("Configuration loaded successfully")
	slog.Info("Development mode status", "dev_mode", cfg.DevMode)

	dbConn, err := sql.Open("sqlite3", cfg.DatabasePath+"?_foreign_keys=on")
	if err != nil {
		slog.Error("Failed to open database connection", "path", cfg.DatabasePath, "error", err)
		os.Exit(1)
	}
	defer func() {
		if closeErr := dbConn.Close(); closeErr != nil {
			slog.Error("Failed to close database connection", "error", closeErr)
		}
	}()
	if err = dbConn.Ping(); err != nil {
		slog.Error("Failed to ping database", "path", cfg.DatabasePath, "error", err)
		os.Exit(1)
	}
	slog.Info("Successfully connected to the database", "path", cfg.DatabasePath)

	runMigrations(dbConn, cfg, logger)

	// --- Initialize Dependencies & Services ---
	querier := db.New(dbConn) // sqlc generated querier
	otpStore := auth.NewInMemoryOTPStore()
	messageSender, err := outbox.NewLogFileMessageSender(cfg.OTPLogPath, logger)
	if err != nil {
		slog.Error("Failed to create LogFileMessageSender", "path", cfg.OTPLogPath, "error", err)
		os.Exit(1)
	}

	// Initialize audit service for security logging
	auditService := service.NewAuditService(querier, logger)

	// Initialize session store for secure session management
	sessionStore := sessions.NewCookieStore([]byte(cfg.JWTSecret))
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   cfg.JWTExpirationHours * 3600, // Convert hours to seconds, sync with JWT expiry
		HttpOnly: true,
		Secure:   !cfg.DevMode, // Use secure cookies in production
		SameSite: http.SameSiteStrictMode,
	}

	userService := service.NewUserService(querier, otpStore, cfg, logger)
	scheduleService := service.NewScheduleService(querier, logger, cfg)
	pointsService := service.NewPointsService(querier, logger)
	bookingService := service.NewBookingService(querier, cfg, logger, pointsService)
	reportService := service.NewReportService(querier, logger, pointsService)
	reportArchivingService := service.NewReportArchivingService(querier, logger)
	adminDashboardService := service.NewAdminDashboardService(querier, scheduleService, logger)
	broadcastService := service.NewBroadcastService(querier, logger, cfg)
	emergencyContactService := service.NewEmergencyContactService(querier, logger)

	// Instantiate PushSender service
	pushSenderService := service.NewPushSender(querier, cfg, logger)

	outboxDispatcherService := outbox.NewDispatcherService(querier, messageSender, pushSenderService, logger, cfg)

	pushAPIHandler := api.NewPushHandler(querier, cfg, logger)

	// --- Setup Cron Jobs ---
	cronLoggerAdapter := &slogCronLogger{logger: logger.With("component", "cron")}
	cronScheduler := cron.New(cron.WithLogger(cron.PrintfLogger(cronLoggerAdapter)))

	// Process outbox every 1 minute
	_, err = cronScheduler.AddFunc("@every 1m", func() {
		processed, errors := outboxDispatcherService.ProcessPendingOutboxMessages(context.Background())
		if errors > 0 {
			slog.Warn("Outbox dispatcher finished with errors", "processed_count", processed, "error_count", errors)
		} else if processed > 0 {
			slog.Info("Outbox dispatcher finished successfully", "processed_count", processed)
		}
	})
	if err != nil {
		slog.Error("Failed to add outbox dispatcher job to cron", "error", err)
		os.Exit(1)
	}

	// Process pending broadcasts every 30 seconds
	_, err = cronScheduler.AddFunc("@every 30s", func() {
		processed, err := broadcastService.ProcessPendingBroadcasts(context.Background())
		if err != nil {
			slog.Error("Failed to process pending broadcasts", "error", err)
		} else if processed > 0 {
			slog.Info("Successfully processed pending broadcasts", "processed_count", processed)
		}
	})
	if err != nil {
		slog.Error("Failed to add broadcast processing job to cron", "error", err)
		os.Exit(1)
	}

	// Auto-archive old reports daily at 2 AM
	_, err = cronScheduler.AddFunc("0 2 * * *", func() {
		ctx := context.Background()
		archived, err := reportArchivingService.ArchiveOldReports(ctx)
		if err != nil {
			slog.Error("Failed to auto-archive old reports", "error", err)
		} else if archived > 0 {
			slog.Info("Successfully auto-archived old reports", "archived_count", archived)
		}
	})
	if err != nil {
		slog.Error("Failed to add report archiving job to cron", "error", err)
		os.Exit(1)
	}

	cronScheduler.Start()
	slog.Info("Cron scheduler started for outbox processing, broadcasts, and report archiving.")

	// --- Setup HTTP Router & Handlers ---
	s := fuego.NewServer(
		fuego.WithAddr(":"+cfg.ServerPort),
		fuego.WithEngineOptions(
			fuego.WithOpenAPIConfig(fuego.OpenAPIConfig{
				UIHandler: func(specURL string) http.Handler {
					return httpSwagger.Handler(
						httpSwagger.URL(specURL),
						httpSwagger.Layout(httpSwagger.BaseLayout),
						httpSwagger.PersistAuthorization(true),
					)
				},
				SwaggerURL:       "/swagger",
				SpecURL:          "/swagger/doc.json",
				JSONFilePath:     "openapi.json",
				Disabled:         false,
				DisableSwaggerUI: false,
				DisableMessages:  false,
			}),
		),
	)

	// Global middlewares
	fuego.Use(s, api.SecurityHeadersMiddleware()) // Add security headers first
	fuego.Use(s, api.AuditContextMiddleware) // Add audit context middleware first
	fuego.Use(s, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapper := w // fuego does not have chi's WrapResponseWriter, so use w directly
			next.ServeHTTP(wrapper, r)
			slog.Info("HTTP request",
				"method", r.Method,
				"path", r.URL.Path,
				// No status or bytes_written without wrapper, unless custom ResponseWriter is implemented
				"latency_ms", time.Since(start).Milliseconds(),
				"remote_addr", r.RemoteAddr,
			)
		})
	})

	// Initialize handlers
	authAPIHandler := api.NewAuthHandler(userService, auditService, logger, cfg, querier, sessionStore)
	scheduleAPIHandler := api.NewScheduleHandler(scheduleService, logger)
	bookingAPIHandler := api.NewBookingHandler(bookingService, auditService, querier, logger)
	reportAPIHandler := api.NewReportHandler(reportService, auditService, logger)
	adminScheduleAPIHandler := api.NewAdminScheduleHandlers(logger, scheduleService, auditService)
	adminUserAPIHandler := api.NewAdminUserHandler(querier, auditService, logger)
	adminBookingAPIHandler := api.NewAdminBookingHandler(bookingService, logger)
	adminReportAPIHandler := api.NewAdminReportHandler(reportService, scheduleService, querier, auditService, logger)
	adminBroadcastAPIHandler := api.NewAdminBroadcastHandler(querier, logger)
	broadcastAPIHandler := api.NewBroadcastHandler(querier, logger)
	adminDashboardAPIHandler := api.NewAdminDashboardHandler(adminDashboardService, logger)
	emergencyContactAPIHandler := api.NewEmergencyContactHandler(emergencyContactService, logger)
	adminAuditAPIHandler := api.NewAdminAuditHandler(auditService, querier, logger)
	leaderboardAPIHandler := api.NewLeaderboardHandler(pointsService, logger)
	calendarAPIHandler := api.NewCalendarHandler(bookingService, querier, logger)

	// Debug: Check handler initialization
	logger.Info("Handler initialization", "booking_handler_nil", bookingAPIHandler == nil, "report_handler_nil", reportAPIHandler == nil, "calendar_handler_nil", calendarAPIHandler == nil)

	// Public API routes
	publicAPI := fuego.Group(s, apiPrefix)
	fuego.PostStd(publicAPI, "/auth/register", authAPIHandler.RegisterHandler)
	fuego.PostStd(publicAPI, "/auth/verify", authAPIHandler.VerifyHandler)
	fuego.GetStd(publicAPI, "/auth/validate", authAPIHandler.ValidateHandler)

	// Development-only auth endpoints
	if cfg.DevMode {
		fuego.PostStd(publicAPI, "/auth/dev-login", authAPIHandler.DevLoginHandler)
		slog.Info("Development mode: dev-login endpoint enabled")
	}

	fuego.GetStd(publicAPI, "/schedules", scheduleAPIHandler.ListSchedulesHandler)
	fuego.GetStd(publicAPI, "/shifts/available", scheduleAPIHandler.ListAvailableShiftsHandler)
	fuego.GetStd(publicAPI, "/shifts/schedule", scheduleAPIHandler.GetPublicScheduleSlotsHandler)
	fuego.GetStd(publicAPI, "/push/vapid-public", pushAPIHandler.VAPIDPublicKey)
	fuego.PostStd(publicAPI, "/ping", api.PingHandler(logger))

	// Emergency contacts (public access)
	fuego.GetStd(publicAPI, "/emergency-contacts", emergencyContactAPIHandler.GetEmergencyContactsHandler)
	fuego.GetStd(publicAPI, "/emergency-contacts/default", emergencyContactAPIHandler.GetDefaultEmergencyContactHandler)

	// Health check endpoints for monitoring
	fuego.GetStd(s, "/health", func(w http.ResponseWriter, r *http.Request) {
		// Check database connectivity
		dbStatus := "up"
		if err := dbConn.Ping(); err != nil {
			dbStatus = "down"
			w.WriteHeader(http.StatusServiceUnavailable)
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"status":   "unhealthy",
				"database": dbStatus,
				"error":    err.Error(),
			}); err != nil {
				slog.Error("Failed to encode health check error response", "error", err)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "healthy",
			"database": dbStatus,
			"uptime":   time.Since(startTime).String(),
			"version":  "1.0.0", // TODO: Use build version
		}); err != nil {
			slog.Error("Failed to encode health check response", "error", err)
		}
	})

	// API health check endpoint
	fuego.GetStd(publicAPI, "/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"service": "night-owls-api",
		}); err != nil {
			slog.Error("Failed to encode API health check response", "error", err)
		}
	})

	// Logout endpoint (requires auth to log out properly, but could also be public)
	fuego.PostStd(publicAPI, "/auth/logout", authAPIHandler.LogoutHandler)

	// Protected routes (require auth)
	protected := fuego.Group(s, apiPrefix)
	fuego.Use(protected, api.AuthMiddleware(cfg, logger, sessionStore))
	fuego.Post(protected, "/bookings", bookingAPIHandler.CreateBookingFuego)
	fuego.GetStd(protected, "/bookings/my", bookingAPIHandler.GetMyBookingsHandler)
	fuego.Post(protected, "/bookings/{id}/checkin", bookingAPIHandler.MarkCheckInFuego)
	fuego.Delete(protected, "/bookings/{id}", bookingAPIHandler.CancelBookingFuego)
	fuego.PostStd(protected, "/bookings/{id}/report", reportAPIHandler.CreateReportHandler)
	fuego.PostStd(protected, "/reports/off-shift", reportAPIHandler.CreateOffShiftReportHandler)
	fuego.GetStd(protected, "/user/reports", reportAPIHandler.ListReportsHandler)
	fuego.GetStd(protected, "/broadcasts", broadcastAPIHandler.ListUserBroadcasts)
	fuego.PostStd(protected, "/push/subscribe", pushAPIHandler.SubscribePush)
	fuego.DeleteStd(protected, "/push/subscribe/{endpoint}", pushAPIHandler.UnsubscribePush)

	// Leaderboard routes (require auth)
	fuego.GetStd(protected, "/leaderboard", leaderboardAPIHandler.GetLeaderboardHandler)
	fuego.GetStd(protected, "/leaderboard/shifts", leaderboardAPIHandler.GetStreakLeaderboardHandler)
	fuego.GetStd(protected, "/leaderboard/activity", leaderboardAPIHandler.GetActivityFeedHandler)
	fuego.GetStd(protected, "/user/stats", leaderboardAPIHandler.GetUserStatsHandler)
	fuego.GetStd(protected, "/user/points/history", leaderboardAPIHandler.GetUserPointsHistoryHandler)
	fuego.GetStd(protected, "/user/achievements", leaderboardAPIHandler.GetUserAchievementsHandler)
	fuego.GetStd(protected, "/user/achievements/available", leaderboardAPIHandler.GetAvailableAchievementsHandler)

	// Calendar routes (require auth)
	logger.Info("Registering calendar routes", "handler_nil", calendarAPIHandler == nil)
	fuego.PostStd(protected, "/calendar/generate-token", calendarAPIHandler.GenerateCalendarFeedToken)
	fuego.GetStd(publicAPI, "/calendar/user/{userId}/{token}", calendarAPIHandler.ServeCalendarFeed)
	logger.Info("Calendar routes registered successfully")

	// Admin routes
	admin := fuego.Group(s, apiPrefix+"/admin")
	fuego.Use(admin, api.AuthMiddleware(cfg, logger, sessionStore))
	fuego.Use(admin, api.AdminMiddleware(logger))

	// Admin Schedules
	fuego.GetStd(admin, "/schedules", adminScheduleAPIHandler.AdminListSchedules)
	fuego.PostStd(admin, "/schedules", adminScheduleAPIHandler.AdminCreateSchedule)
	fuego.GetStd(admin, "/schedules/all-slots", adminScheduleAPIHandler.AdminListAllShiftSlots)
	fuego.GetStd(admin, "/schedules/{id}", adminScheduleAPIHandler.AdminGetSchedule)
	fuego.PutStd(admin, "/schedules/{id}", adminScheduleAPIHandler.AdminUpdateSchedule)
	fuego.DeleteStd(admin, "/schedules/{id}", adminScheduleAPIHandler.AdminDeleteSchedule)
	fuego.DeleteStd(admin, "/schedules", adminScheduleAPIHandler.AdminBulkDeleteSchedules)

	// Admin Users
	fuego.GetStd(admin, "/users", adminUserAPIHandler.AdminListUsers)
	fuego.PostStd(admin, "/users", adminUserAPIHandler.AdminCreateUser)
	fuego.GetStd(admin, "/users/{id}", adminUserAPIHandler.AdminGetUser)
	fuego.GetStd(admin, "/users/{userId}/bookings", adminBookingAPIHandler.GetUserBookingsHandler)
	fuego.PutStd(admin, "/users/{id}", adminUserAPIHandler.AdminUpdateUser)
	fuego.DeleteStd(admin, "/users/{id}", adminUserAPIHandler.AdminDeleteUser)
	fuego.PostStd(admin, "/users/bulk-delete", adminUserAPIHandler.AdminBulkDeleteUsers)

	// Admin Bookings
	fuego.PostStd(admin, "/bookings/assign", adminBookingAPIHandler.AssignUserToShiftHandler)
	fuego.PostStd(admin, "/bookings/unassign", adminBookingAPIHandler.UnassignUserFromShiftHandler)

	// Admin Reports
	fuego.GetStd(admin, "/reports", adminReportAPIHandler.AdminListReportsHandler)
	fuego.GetStd(admin, "/reports/archived", adminReportAPIHandler.AdminListArchivedReportsHandler)
	fuego.GetStd(admin, "/reports/{id}", adminReportAPIHandler.AdminGetReportHandler)
	fuego.PutStd(admin, "/reports/{id}/archive", adminReportAPIHandler.AdminArchiveReportHandler)
	fuego.PutStd(admin, "/reports/{id}/unarchive", adminReportAPIHandler.AdminUnarchiveReportHandler)
	fuego.DeleteStd(admin, "/reports/{id}", adminReportAPIHandler.AdminDeleteReportHandler)

	// Admin Broadcasts
	fuego.GetStd(admin, "/broadcasts", adminBroadcastAPIHandler.AdminListBroadcasts)
	fuego.PostStd(admin, "/broadcasts", adminBroadcastAPIHandler.AdminCreateBroadcast)
	fuego.GetStd(admin, "/broadcasts/{id}", adminBroadcastAPIHandler.AdminGetBroadcast)
	fuego.DeleteStd(admin, "/broadcasts/{id}", adminBroadcastAPIHandler.AdminDeleteBroadcast)

	// Test user broadcasts under admin for debugging
	fuego.GetStd(admin, "/test-broadcasts", broadcastAPIHandler.ListUserBroadcasts)

	// Debug endpoint to manually trigger broadcast processing
	fuego.PostStd(admin, "/debug/process-broadcasts", func(w http.ResponseWriter, r *http.Request) {
		processed, err := broadcastService.ProcessPendingBroadcasts(r.Context())
		if err != nil {
			logger.Error("Failed to process broadcasts", "error", err)
			http.Error(w, "Failed to process broadcasts: "+err.Error(), http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"processed": processed,
			"message":   fmt.Sprintf("Successfully processed %d broadcasts", processed),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Error("Failed to encode broadcast processing response", "error", err)
		}
	})

	// Debug endpoint to manually trigger report archiving
	fuego.PostStd(admin, "/debug/archive-reports", func(w http.ResponseWriter, r *http.Request) {
		archived, err := reportArchivingService.ArchiveOldReports(r.Context())
		if err != nil {
			logger.Error("Failed to archive reports", "error", err)
			http.Error(w, "Failed to archive reports: "+err.Error(), http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"archived": archived,
			"message":  fmt.Sprintf("Successfully archived %d reports", archived),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Error("Failed to encode archive reports response", "error", err)
		}
	})

	// Debug endpoint to get archiving statistics
	fuego.GetStd(admin, "/debug/archiving-stats", func(w http.ResponseWriter, r *http.Request) {
		stats, err := reportArchivingService.GetArchivingStats(r.Context())
		if err != nil {
			logger.Error("Failed to get archiving stats", "error", err)
			http.Error(w, "Failed to get archiving stats: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(stats); err != nil {
			logger.Error("Failed to encode archiving stats response", "error", err)
		}
	})

	// Debug endpoint to test push notifications
	fuego.PostStd(admin, "/debug/test-push", func(w http.ResponseWriter, r *http.Request) {
		userIDVal := r.Context().Value(api.UserIDKey)
		userID, ok := userIDVal.(int64)
		if !ok {
			http.Error(w, "User ID not found in context", http.StatusUnauthorized)
			return
		}

		// Create test push notification payload
		testPayload := map[string]interface{}{
			"type":  "test",
			"title": "Test Push Notification",
			"body":  "This is a test push notification from the admin panel",
			"data": map[string]interface{}{
				"test": true,
			},
		}

		payloadBytes, err := json.Marshal(testPayload)
		if err != nil {
			logger.Error("Failed to marshal test push payload", "error", err)
			http.Error(w, "Failed to create test payload: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Send test push notification
		pushSenderService.Send(r.Context(), userID, payloadBytes, 300) // 5 minutes TTL

		response := map[string]interface{}{
			"message": "Test push notification sent",
			"user_id": userID,
			"payload": testPayload,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Error("Failed to encode test push response", "error", err)
		}
	})

	// Simple test handler - mimicking working admin handlers
	fuego.GetStd(admin, "/simple-test", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Simple test handler called")
		// Mimic the exact response pattern of working admin handlers
		broadcasts := []map[string]interface{}{
			{
				"id":         999,
				"message":    "Test broadcast from simple handler",
				"audience":   "all",
				"created_at": "2025-05-26T12:00:00Z",
			},
		}

		// Use the same response pattern as other handlers
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(broadcasts); err != nil {
			logger.Error("Failed to encode simple test response", "error", err)
		}
	})

	// Admin Dashboard
	fuego.GetStd(admin, "/dashboard", adminDashboardAPIHandler.GetDashboardHandler)

	// Admin Emergency Contacts
	fuego.GetStd(admin, "/emergency-contacts", emergencyContactAPIHandler.AdminGetEmergencyContactsHandler)
	fuego.PostStd(admin, "/emergency-contacts", emergencyContactAPIHandler.AdminCreateEmergencyContactHandler)
	fuego.GetStd(admin, "/emergency-contacts/{id}", emergencyContactAPIHandler.AdminGetEmergencyContactHandler)
	fuego.PutStd(admin, "/emergency-contacts/{id}", emergencyContactAPIHandler.AdminUpdateEmergencyContactHandler)
	fuego.DeleteStd(admin, "/emergency-contacts/{id}", emergencyContactAPIHandler.AdminDeleteEmergencyContactHandler)
	fuego.PutStd(admin, "/emergency-contacts/{id}/default", emergencyContactAPIHandler.AdminSetDefaultEmergencyContactHandler)

	// Admin Audit Trail
	fuego.GetStd(admin, "/audit-events", adminAuditAPIHandler.AdminListAuditEvents)
	fuego.GetStd(admin, "/audit-events/stats", adminAuditAPIHandler.AdminGetAuditStats)
	fuego.GetStd(admin, "/audit-events/type-stats", adminAuditAPIHandler.AdminGetAuditEventTypeStats)

	// Explicit Swagger routes (must be before SPA fallback)
	fuego.GetStd(s, "/swagger", func(w http.ResponseWriter, r *http.Request) {
		handler := httpSwagger.Handler(
			httpSwagger.URL("/swagger/doc.json"),
			httpSwagger.Layout(httpSwagger.BaseLayout),
			httpSwagger.PersistAuthorization(true),
		)
		handler.ServeHTTP(w, r)
	})
	fuego.GetStd(s, "/swagger/", func(w http.ResponseWriter, r *http.Request) {
		handler := httpSwagger.Handler(
			httpSwagger.URL("/swagger/doc.json"),
			httpSwagger.Layout(httpSwagger.BaseLayout),
			httpSwagger.PersistAuthorization(true),
		)
		handler.ServeHTTP(w, r)
	})
	fuego.GetStd(s, "/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		// Generate OpenAPI spec and serve it
		spec := s.OutputOpenAPISpec()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(spec); err != nil {
			logger.Error("Failed to encode OpenAPI spec", "error", err)
		}
	})

	// --- Static File Serving ---
	// NOTE: Static files and SPA routing are now handled by Caddy
	// Go server only handles API routes - no static file serving needed

	// MIME tweak for webmanifest (keeping for any remaining Go-served content)
	if err := mime.AddExtensionType(".webmanifest", "application/manifest+json"); err != nil {
		logger.Warn("Failed to add MIME type for webmanifest", "error", err)
	}

	// --- Start HTTP Server ---
	httpServer := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      s.Mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	slog.Info("Night Owls Control Backend Starting HTTP server...", "port", cfg.ServerPort)

	// Goroutine for graceful shutdown
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server ListenAndServe error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")

	// Shutdown cron scheduler
	slog.Info("Stopping cron scheduler...")
	cronCtx := cronScheduler.Stop() // Stop() returns a context that is done when all running jobs are complete.
	select {
	case <-cronCtx.Done():
		slog.Info("Cron scheduler stopped gracefully.")
	case <-time.After(10 * time.Second): // Timeout for cron jobs to finish
		slog.Warn("Cron scheduler shutdown timed out.")
	}

	// Shutdown HTTP server
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelShutdown()

	if err := httpServer.Shutdown(ctxShutdown); err != nil {
		slog.Error("HTTP server Shutdown error", "error", err)
	}

	slog.Info("Server gracefully stopped.")
}

func runMigrations(dbConn *sql.DB, cfg *config.Config, logger *slog.Logger) {
	// For migrations, it's cleaner to let migrate manage its own DB connection
	// based on the DSN, rather than sharing and potentially closing the main app's dbConn.
	migrationDSN := "sqlite3://" + cfg.DatabasePath
	logger.Info("Preparing to run migrations using DSN", "dsn", migrationDSN)

	m, err := migrate.New(
		"file://./internal/db/migrations",
		migrationDSN)
	if err != nil {
		logger.Error("Failed to create migrate instance with DSN", "dsn", migrationDSN, "error", err)
		os.Exit(1)
	}
	// It's important to defer Close on the migrate instance created with New()
	// to clean up its own database connection and source file handles.
	defer func() {
		if srcErr, dbErr := m.Close(); srcErr != nil || dbErr != nil {
			if srcErr != nil {
				logger.Warn("Error closing migration source after DSN-based migration", "error", srcErr)
			}
			if dbErr != nil {
				logger.Warn("Error closing migration database connection after DSN-based migration", "error", dbErr)
			}
		} else {
			logger.Info("Migration instance (DSN-based) closed successfully.")
		}
	}()

	logger.Info("Running database migrations...")
	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Error("Failed to apply migrations using DSN", "dsn", migrationDSN, "error", err)
		os.Exit(1)
	} else if err == migrate.ErrNoChange {
		logger.Info("No new migrations to apply.")
	} else {
		logger.Info("Database migrations applied successfully.")
	}
	// The main dbConn (passed as an argument but no longer directly used here) remains untouched and managed by main().
}
