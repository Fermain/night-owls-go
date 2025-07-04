package config

import (
	"fmt"
	"os"
	"strconv"
	"strings" // For ToLower on log level/format
	"time"
)

// Config holds the application configuration.
// Values are typically loaded from environment variables.
type Config struct {
	ServerPort           string
	DatabasePath         string
	JWTSecret            string
	DefaultShiftDuration time.Duration
	OTPLogPath           string // Path for logging OTPs instead of sending them
	LogLevel             string // e.g., "debug", "info", "warn", "error"
	LogFormat            string // e.g., "json", "text"

	JWTExpirationHours int
	OTPValidityMinutes int
	// OTPLength          int // Usually fixed by implementation, less often configured
	OutboxBatchSize  int
	OutboxMaxRetries int

	// Development mode - enables dev features like OTP in responses
	DevMode bool

	// PWA / WebPush specific
	VAPIDPublic  string
	VAPIDPrivate string
	VAPIDSubject string

	// Twilio configuration for SMS OTP
	TwilioAccountSID string
	TwilioAuthToken  string
	TwilioVerifySID  string
	TwilioFromNumber string
}

// Security validation constants
const (
	DefaultJWTSecret = "super-secret-jwt-key-please-change-in-prod"
)

// ValidateSecurityConfig validates critical security configurations
func (c *Config) ValidateSecurityConfig() error {
	// Check for default JWT secret
	if c.JWTSecret == DefaultJWTSecret {
		if isProductionEnvironment() {
			return fmt.Errorf("CRITICAL SECURITY ERROR: Default JWT secret detected in production environment. Set JWT_SECRET environment variable")
		}
		// Warn in development but don't fail
		fmt.Printf("WARNING: Using default JWT secret. Set JWT_SECRET environment variable for production\n")
	}

	// Check for weak JWT secrets (too short)
	if len(c.JWTSecret) < 32 {
		if isProductionEnvironment() {
			return fmt.Errorf("CRITICAL SECURITY ERROR: JWT secret too short (%d chars). Use at least 32 characters", len(c.JWTSecret))
		}
		fmt.Printf("WARNING: JWT secret is short (%d chars). Use at least 32 characters for production\n", len(c.JWTSecret))
	}

	// Validate dev mode in production
	if c.DevMode && isProductionEnvironment() {
		return fmt.Errorf("CRITICAL SECURITY ERROR: Development mode cannot be enabled in production environment")
	}

	return nil
}

// isProductionEnvironment detects if we're running in a production environment
func isProductionEnvironment() bool {
	env := strings.ToLower(os.Getenv("ENVIRONMENT"))
	goEnv := strings.ToLower(os.Getenv("GO_ENV"))
	nodeEnv := strings.ToLower(os.Getenv("NODE_ENV"))
	
	// Check common production environment indicators
	return env == "production" || env == "prod" ||
		   goEnv == "production" || goEnv == "prod" ||
		   nodeEnv == "production" || nodeEnv == "prod" ||
		   os.Getenv("RAILWAY_ENVIRONMENT") == "production" ||
		   os.Getenv("VERCEL_ENV") == "production" ||
		   os.Getenv("HEROKU_APP_NAME") != "" ||
		   os.Getenv("PORT") != "" // Common production indicator
}

// LoadConfig loads configuration from environment variables
// and sets defaults for missing values.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		// Default values
		ServerPort:           "5888",
		DatabasePath:         "./community_watch.db",
		JWTSecret:            "super-secret-jwt-key-please-change-in-prod", // IMPORTANT: Change this in production!
		DefaultShiftDuration: 2 * time.Hour,
		OTPLogPath:           "./sms_outbox.log",
		LogLevel:             "info", // Default log level
		LogFormat:            "json", // Default log format

		JWTExpirationHours: 24, // Default 24 hours
		OTPValidityMinutes: 5,  // Default 5 minutes
		// OTPLength:          6,     // Default 6 digits (if we make it configurable)
		OutboxBatchSize:  10, // Default 10 messages per batch
		OutboxMaxRetries: 3,  // Default 3 retries

		// Development mode - enables dev features like OTP in responses
		DevMode: false,

		// PWA / WebPush defaults
		VAPIDSubject: "mailto:admin@example.com", // Default VAPID subject
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

	// Load development mode
	if val := os.Getenv("DEV_MODE"); val != "" {
		if devMode, err := strconv.ParseBool(val); err == nil {
			cfg.DevMode = devMode
		}
	}

	// Load PWA / WebPush specific environment variables
	if vapidPublic := os.Getenv("VAPID_PUBLIC_KEY"); vapidPublic != "" {
		cfg.VAPIDPublic = vapidPublic
	}
	if vapidPrivate := os.Getenv("VAPID_PRIVATE_KEY"); vapidPrivate != "" {
		cfg.VAPIDPrivate = vapidPrivate
	}
	if vapidSubject := os.Getenv("VAPID_SUBJECT"); vapidSubject != "" {
		cfg.VAPIDSubject = vapidSubject
	}

	// Load Twilio configuration
	if twilioAccountSID := os.Getenv("TWILIO_ACCOUNT_SID"); twilioAccountSID != "" {
		cfg.TwilioAccountSID = twilioAccountSID
	}
	if twilioAuthToken := os.Getenv("TWILIO_AUTH_TOKEN"); twilioAuthToken != "" {
		cfg.TwilioAuthToken = twilioAuthToken
	}
	if twilioVerifySID := os.Getenv("TWILIO_VERIFY_SID"); twilioVerifySID != "" {
		cfg.TwilioVerifySID = twilioVerifySID
	}
	if twilioFromNumber := os.Getenv("TWILIO_FROM_NUMBER"); twilioFromNumber != "" {
		cfg.TwilioFromNumber = twilioFromNumber
	}

	return cfg, nil
}
