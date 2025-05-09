package config

import (
	"os"
	"strconv"
	"strings" // For ToLower on log level/format
	"time"
)

// Config holds the application configuration.
// Values are typically loaded from environment variables.
type Config struct {
	ServerPort         string
	DatabasePath       string
	JWTSecret          string
	DefaultShiftDuration time.Duration
	OTPLogPath         string // Path for logging OTPs instead of sending them
	LogLevel           string // e.g., "debug", "info", "warn", "error"
	LogFormat          string // e.g., "json", "text"

	JWTExpirationHours int
	OTPValidityMinutes int
	// OTPLength          int // Usually fixed by implementation, less often configured
	OutboxBatchSize    int
	OutboxMaxRetries   int

	// PWA / WebPush specific
	VAPIDPublic  string
	VAPIDPrivate string
	VAPIDSubject string
	StaticDir    string
}

// LoadConfig loads configuration from environment variables
// and sets defaults for missing values.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		// Default values
		ServerPort:         "8080",
		DatabasePath:       "./community_watch.db",
		JWTSecret:          "super-secret-jwt-key-please-change-in-prod", // IMPORTANT: Change this in production!
		DefaultShiftDuration: 2 * time.Hour,
		OTPLogPath:         "./sms_outbox.log",
		LogLevel:           "info", // Default log level
		LogFormat:          "json", // Default log format

		JWTExpirationHours: 24,    // Default 24 hours
		OTPValidityMinutes: 5,     // Default 5 minutes
		// OTPLength:          6,     // Default 6 digits (if we make it configurable)
		OutboxBatchSize:    10,    // Default 10 messages per batch
		OutboxMaxRetries:   3,     // Default 3 retries

		// PWA / WebPush defaults
		VAPIDSubject: "mailto:admin@example.com", // Default VAPID subject
		StaticDir:    "app/build",          // Default static file directory - Hardcoded as per user request
	}

	if port := os.Getenv("SERVER_PORT"); port != "" {
		cfg.ServerPort = port
	}

	if dbPath := os.Getenv("DATABASE_PATH"); dbPath != "" {
		cfg.DatabasePath = dbPath
	}

	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		cfg.JWTSecret = jwtSecret
	}

	if shiftDurationHours := os.Getenv("DEFAULT_SHIFT_DURATION_HOURS"); shiftDurationHours != "" {
		duration, err := strconv.Atoi(shiftDurationHours)
		if err == nil && duration > 0 {
			cfg.DefaultShiftDuration = time.Duration(duration) * time.Hour
		}
		// We could log an error here if parsing fails, but for now, we'll use the default.
	}

	if otpLogPath := os.Getenv("OTP_LOG_PATH"); otpLogPath != "" {
		cfg.OTPLogPath = otpLogPath
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.LogLevel = strings.ToLower(logLevel)
	}

	if logFormat := os.Getenv("LOG_FORMAT"); logFormat != "" {
		cfg.LogFormat = strings.ToLower(logFormat)
	}

	// Load new integer fields
	if val := os.Getenv("JWT_EXPIRATION_HOURS"); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil && intVal > 0 {
			cfg.JWTExpirationHours = intVal
		}
	}
	if val := os.Getenv("OTP_VALIDITY_MINUTES"); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil && intVal > 0 {
			cfg.OTPValidityMinutes = intVal
		}
	}
	// if val := os.Getenv("OTP_LENGTH"); val != "" { ... cfg.OTPLength = intVal ... }
	if val := os.Getenv("OUTBOX_BATCH_SIZE"); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil && intVal > 0 {
			cfg.OutboxBatchSize = intVal
		}
	}
	if val := os.Getenv("OUTBOX_MAX_RETRIES"); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil && intVal >= 0 { // Max retries can be 0
			cfg.OutboxMaxRetries = intVal
		}
	}

	// Load PWA / WebPush specific environment variables
	if vapidPublic := os.Getenv("VAPID_PUBLIC"); vapidPublic != "" {
		cfg.VAPIDPublic = vapidPublic
	}
	if vapidPrivate := os.Getenv("VAPID_PRIVATE"); vapidPrivate != "" {
		cfg.VAPIDPrivate = vapidPrivate
	}
	if vapidSubject := os.Getenv("VAPID_SUBJECT"); vapidSubject != "" {
		cfg.VAPIDSubject = vapidSubject
	}
	if staticDir := os.Getenv("STATIC_DIR"); staticDir != "" {
		cfg.StaticDir = staticDir
	}

	return cfg, nil
} 