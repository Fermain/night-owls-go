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

	return cfg, nil
} 