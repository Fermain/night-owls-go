// Package main is the entry point for the Community Watch API server
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
	"path/filepath"
	"strings"
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
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	httpSwagger "github.com/swaggo/http-swagger"

	// Import the generated swagger docs when available
	_ "night-owls-go/docs/swagger"

	"github.com/go-fuego/fuego"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

// @title Community Watch Shift Scheduler API
// @version 1.0
// @description API for managing community watch shifts, bookings, and reports
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
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
	err := godotenv.Load()
	if err != nil {
		// Log this initial finding using standard log before slog is fully set up
		log.Println("No .env file found, using environment variables or defaults")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Critical: Error loading configuration: %v", err)
	}

	logger := logging.NewLogger(cfg) // Initialize logger with config
	slog.SetDefault(logger)          // Set as global default

	slog.Info("Configuration loaded successfully")
	slog.Info("Development mode status", "dev_mode", cfg.DevMode)

	dbConn, err := sql.Open("sqlite3", cfg.DatabasePath)
	if err != nil {
		slog.Error("Failed to open database connection", "path", cfg.DatabasePath, "error", err)
		os.Exit(1)
	}
	defer dbConn.Close()
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

	userService := service.NewUserService(querier, otpStore, cfg, logger)
	scheduleService := service.NewScheduleService(querier, logger, cfg)
	bookingService := service.NewBookingService(querier, cfg, logger)
	reportService := service.NewReportService(querier, logger)
	recurringAssignmentService := service.NewRecurringAssignmentService(querier, logger, cfg)
	adminDashboardService := service.NewAdminDashboardService(querier, scheduleService, logger)

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

	// Materialize recurring assignments every hour
	_, err = cronScheduler.AddFunc("@hourly", func() {
		ctx := context.Background()
		now := time.Now().UTC()
		fromTime := now
		toTime := now.AddDate(0, 0, 14) // Next 2 weeks
		
		err := recurringAssignmentService.MaterializeUpcomingBookings(ctx, scheduleService, fromTime, toTime)
		if err != nil {
			slog.Error("Failed to materialize recurring assignment bookings", "error", err)
		}
	})
	if err != nil {
		slog.Error("Failed to add recurring assignment materialization job to cron", "error", err)
		os.Exit(1)
	}

	cronScheduler.Start()
	slog.Info("Cron scheduler started for outbox processing and recurring assignments.")

	// --- Setup HTTP Router & Handlers ---
	s := fuego.NewServer(
		fuego.WithAddr(":" + cfg.ServerPort),
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
	authAPIHandler := api.NewAuthHandler(userService, logger, cfg, querier)
	scheduleAPIHandler := api.NewScheduleHandler(scheduleService, logger)
	bookingAPIHandler := api.NewBookingHandler(bookingService, logger)
	reportAPIHandler := api.NewReportHandler(reportService, logger)
	adminScheduleAPIHandler := api.NewAdminScheduleHandlers(logger, scheduleService)
	adminUserAPIHandler := api.NewAdminUserHandler(querier, logger)
	adminBookingAPIHandler := api.NewAdminBookingHandler(bookingService, logger)
	adminRecurringAssignmentAPIHandler := api.NewAdminRecurringAssignmentHandlers(logger, recurringAssignmentService, scheduleService)
	adminReportAPIHandler := api.NewAdminReportHandler(reportService, scheduleService, querier, logger)
	adminBroadcastAPIHandler := api.NewAdminBroadcastHandler(querier, logger)
	adminDashboardAPIHandler := api.NewAdminDashboardHandler(adminDashboardService, logger)

	// Debug: Check handler initialization
	logger.Info("Handler initialization", "booking_handler_nil", bookingAPIHandler == nil, "report_handler_nil", reportAPIHandler == nil)

	// Public routes
	fuego.PostStd(s, "/api/auth/register", authAPIHandler.RegisterHandler)
	fuego.PostStd(s, "/api/auth/verify", authAPIHandler.VerifyHandler)
	
	// Development-only auth endpoints
	if cfg.DevMode {
		fuego.PostStd(s, "/api/auth/dev-login", authAPIHandler.DevLoginHandler)
		slog.Info("Development mode: dev-login endpoint enabled")
	}
	
	fuego.GetStd(s, "/schedules", scheduleAPIHandler.ListSchedulesHandler)
	fuego.GetStd(s, "/shifts/available", scheduleAPIHandler.ListAvailableShiftsHandler)
	fuego.GetStd(s, "/push/vapid-public", pushAPIHandler.VAPIDPublicKey)
	fuego.PostStd(s, "/api/ping", api.PingHandler(logger))
	
	// Protected routes (require auth)
	protected := fuego.Group(s, "")
	fuego.Use(protected, api.AuthMiddleware(cfg, logger))
	fuego.PostStd(protected, "/bookings", bookingAPIHandler.CreateBookingHandler)
	fuego.GetStd(protected, "/bookings/my", bookingAPIHandler.GetMyBookingsHandler)
	fuego.PostStd(protected, "/bookings/{id}/checkin", bookingAPIHandler.MarkCheckInHandler)
	fuego.PostStd(protected, "/bookings/{id}/report", reportAPIHandler.CreateReportHandler)
	fuego.PostStd(protected, "/push/subscribe", pushAPIHandler.SubscribePush)
	fuego.DeleteStd(protected, "/push/subscribe/{endpoint}", pushAPIHandler.UnsubscribePush)

	// Admin routes
	admin := fuego.Group(s, "/api/admin")
	fuego.Use(admin, api.AuthMiddleware(cfg, logger))

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

	// Admin Recurring Assignments
	fuego.GetStd(admin, "/recurring-assignments", adminRecurringAssignmentAPIHandler.AdminListRecurringAssignments)
	fuego.PostStd(admin, "/recurring-assignments", adminRecurringAssignmentAPIHandler.AdminCreateRecurringAssignment)
	fuego.GetStd(admin, "/recurring-assignments/{id}", adminRecurringAssignmentAPIHandler.AdminGetRecurringAssignment)
	fuego.PutStd(admin, "/recurring-assignments/{id}", adminRecurringAssignmentAPIHandler.AdminUpdateRecurringAssignment)
	fuego.DeleteStd(admin, "/recurring-assignments/{id}", adminRecurringAssignmentAPIHandler.AdminDeleteRecurringAssignment)

	// Admin Reports
	fuego.GetStd(admin, "/reports", adminReportAPIHandler.AdminListReportsHandler)
	fuego.GetStd(admin, "/reports/{id}", adminReportAPIHandler.AdminGetReportHandler)

	// Admin Broadcasts
	fuego.GetStd(admin, "/broadcasts", adminBroadcastAPIHandler.AdminListBroadcasts)
	fuego.PostStd(admin, "/broadcasts", adminBroadcastAPIHandler.AdminCreateBroadcast)
	fuego.GetStd(admin, "/broadcasts/{id}", adminBroadcastAPIHandler.AdminGetBroadcast)

	// Admin Dashboard
	fuego.GetStd(admin, "/dashboard", adminDashboardAPIHandler.GetDashboardHandler)

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
		spec := s.Engine.OutputOpenAPISpec()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(spec)
	})

	// --- Serve Static Assets & SPA ---
	// Static file serving
	staticPath, err := filepath.Abs(cfg.StaticDir)
	if err != nil {
		logger.Error("Failed to get absolute path for static directory", "path", cfg.StaticDir, "error", err)
		os.Exit(1)
	}
	logger.Info("Serving static files from", "path", staticPath)

	// Serve index.html for the root path
	fuego.GetStd(s, "/", func(w http.ResponseWriter, r *http.Request) {
		indexPage := filepath.Join(staticPath, "index.html")
		if _, err := os.Stat(indexPage); os.IsNotExist(err) {
			logger.Error("SPA index.html not found", "path", indexPage)
			http.Error(w, "Internal Server Error: Application not found", http.StatusInternalServerError)
			return
		}
		http.ServeFile(w, r, indexPage)
	})

	// Serve static files and fallback to index.html for SPA routes
	fuego.GetStd(s, "/*filepath", func(w http.ResponseWriter, r *http.Request) {
		// Don't serve SPA for API requests - let them 404 if not found
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		
		requestedFilePath := filepath.Join(staticPath, r.URL.Path)
		stat, err := os.Stat(requestedFilePath)
		if err == nil && !stat.IsDir() {
			http.ServeFile(w, r, requestedFilePath)
			return
		}
		indexPage := filepath.Join(staticPath, "index.html")
		if _, err := os.Stat(indexPage); os.IsNotExist(err) {
			logger.Error("SPA index.html not found", "path", indexPage)
			http.Error(w, "Internal Server Error: Application not found", http.StatusInternalServerError)
			return
		}
		http.ServeFile(w, r, indexPage)
	})

	// MIME tweak for webmanifest
	// This ensures .webmanifest files are served with the correct Content-Type.
	mime.AddExtensionType(".webmanifest", "application/manifest+json")

	// --- Start HTTP Server ---
	httpServer := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: s.Mux,
		ReadTimeout: 5 * time.Second, 
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  120 * time.Second,
	}

	slog.Info("Community Watch Backend Starting HTTP server...", "port", cfg.ServerPort)

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
		"file://internal/db/migrations",
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