// Package main is the entry point for the Community Watch API server
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"night-owls-go/internal/api"
	"night-owls-go/internal/auth"
	"night-owls-go/internal/config"
	db "night-owls-go/internal/db/sqlc_generated"
	"night-owls-go/internal/logging"
	"night-owls-go/internal/outbox"
	"night-owls-go/internal/service"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3" // Blank import for DSN-based migration driver registration
	_ "github.com/golang-migrate/migrate/v4/source/file"      // Driver for reading migration files from disk
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/robfig/cron/v3"
	httpSwagger "github.com/swaggo/http-swagger"
	// Import the generated swagger docs when available
	// _ "night-owls-go/docs"
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

	// Instantiate PushSender service
	pushSenderService := service.NewPushSender(querier, cfg, logger)

	outboxDispatcherService := outbox.NewDispatcherService(querier, messageSender, pushSenderService, logger, cfg)

	pushAPIHandler := api.NewPushHandler(querier, cfg, logger)

	// --- Setup Cron Jobs ---
	cronLoggerAdapter := &slogCronLogger{logger: logger.With("component", "cron")}
	cronScheduler := cron.New(cron.WithLogger(cron.PrintfLogger(cronLoggerAdapter)))
	_, err = cronScheduler.AddFunc("@every 1m", func() { // Process outbox every 1 minute
		processed, errors := outboxDispatcherService.ProcessPendingOutboxMessages(context.Background()) // Use background context for cron job
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
	cronScheduler.Start()
	slog.Info("Cron scheduler started for outbox processing.")

	// --- Setup HTTP Router & Handlers ---
	router := chi.NewRouter()
	router.Use(chiMiddleware.RequestID)
	router.Use(chiMiddleware.RealIP)
	// Using slog for logging HTTP requests via middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapper := chiMiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(wrapper, r)
			slog.Info("HTTP request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapper.Status(),
				"latency_ms", time.Since(start).Milliseconds(),
				"bytes_written", wrapper.BytesWritten(),
				"remote_addr", r.RemoteAddr,
				"request_id", chiMiddleware.GetReqID(r.Context()),
			)
		})
	})
	router.Use(chiMiddleware.Recoverer) // Recovers from panics and returns a 500

	// Initialize handlers
	authAPIHandler := api.NewAuthHandler(userService, logger)
	scheduleAPIHandler := api.NewScheduleHandler(scheduleService, logger)
	bookingAPIHandler := api.NewBookingHandler(bookingService, logger)
	reportAPIHandler := api.NewReportHandler(reportService, logger)
	adminScheduleAPIHandler := api.NewAdminScheduleHandlers(logger, scheduleService) // Instantiate AdminScheduleHandlers
	adminUserAPIHandler := api.NewAdminUserHandler(querier, logger) // Instantiate AdminUserHandler

	// Public routes
	router.Post("/auth/register", authAPIHandler.RegisterHandler)
	router.Post("/auth/verify", authAPIHandler.VerifyHandler)
	router.Get("/schedules", scheduleAPIHandler.ListSchedulesHandler)             // Optional, as per guide
	router.Get("/shifts/available", scheduleAPIHandler.ListAvailableShiftsHandler)
	router.Get("/push/vapid-public", pushAPIHandler.VAPIDPublicKey)
	router.Post("/api/ping", api.PingHandler(logger)) // New Ping endpoint

	// Protected routes (require auth)
	router.Group(func(r chi.Router) {
		r.Use(api.AuthMiddleware(cfg, logger)) // Apply AuthMiddleware

		r.Post("/bookings", bookingAPIHandler.CreateBookingHandler)
		r.Patch("/bookings/{id}/attendance", bookingAPIHandler.MarkAttendanceHandler)
		// r.Delete("/bookings/{id}", bookingAPIHandler.CancelBookingHandler) // Optional

		r.Post("/bookings/{id}/report", reportAPIHandler.CreateReportHandler)
		// r.Get("/reports", reportAPIHandler.ListReportsHandler) // Optional

		// Push notification routes (protected)
		r.Post("/push/subscribe", pushAPIHandler.SubscribePush)
		r.Delete("/push/subscribe/{endpoint}", pushAPIHandler.UnsubscribePush)
	})

	// Admin routes (currently unprotected for development of CRUD views)
	router.Route("/api/admin", func(r chi.Router) {
		r.Use(api.AuthMiddleware(cfg, logger))

		// Admin Schedules routes
		router.Route("/api/admin/schedules", func(r chi.Router) {
			r.Get("/", adminScheduleAPIHandler.AdminListSchedules)          // GET /api/admin/schedules
			r.Post("/", adminScheduleAPIHandler.AdminCreateSchedule)         // POST /api/admin/schedules
			r.Get("/{id}", adminScheduleAPIHandler.AdminGetSchedule)          // GET /api/admin/schedules/{id}
			r.Put("/{id}", adminScheduleAPIHandler.AdminUpdateSchedule)        // PUT /api/admin/schedules/{id}
			r.Delete("/{id}", adminScheduleAPIHandler.AdminDeleteSchedule)    // DELETE /api/admin/schedules/{id}
			r.Delete("/", adminScheduleAPIHandler.AdminBulkDeleteSchedules) // DELETE /api/admin/schedules
		})

		// Admin Users routes
		router.Route("/api/admin/users", func(r chi.Router) {
			r.Get("/", adminUserAPIHandler.AdminListUsers)           // GET /api/admin/users
			r.Post("/", adminUserAPIHandler.AdminCreateUser)          // POST /api/admin/users
			r.Get("/{id}", adminUserAPIHandler.AdminGetUser)          // GET /api/admin/users/{id}
			r.Put("/{id}", adminUserAPIHandler.AdminUpdateUser)        // PUT /api/admin/users/{id}
		})

		// Add other admin routes here
	})

	// Swagger documentation
	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // The URL pointing to API definition
	))

	// --- Serve Static Assets & SPA ---
	// Get the absolute path to the static directory (app/build)
	staticPath, err := filepath.Abs(cfg.StaticDir)
	if err != nil {
		slog.Error("Failed to get absolute path for static directory", "path", cfg.StaticDir, "error", err)
		os.Exit(1)
	}
	slog.Info("Serving static files from", "path", staticPath)

	// The NotFound handler will now attempt to serve static files first.
	// If a specific file is not found at the requested path, it serves index.html for SPA routing.
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		// Construct the path to the potential file in the static directory.
		// r.URL.Path is already relative to the server root.
		requestedFilePath := filepath.Join(staticPath, r.URL.Path)

		// Check if the requested path corresponds to an existing file (and not a directory).
		stat, err := os.Stat(requestedFilePath)
		if err == nil && !stat.IsDir() {
			// If the file exists, serve it.
			http.ServeFile(w, r, requestedFilePath)
			return
		}

		// If the file doesn't exist (or it's a directory), serve the main index.html
		// This allows SPA client-side routing to take over.
		indexPage := filepath.Join(staticPath, "index.html")
		// Log if index.html itself is missing, which would be a critical build/config error.
		if _, err := os.Stat(indexPage); os.IsNotExist(err) {
			slog.ErrorContext(r.Context(), "SPA index.html not found", "path", indexPage)
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
		Handler: router,
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