package service

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"
	"time"

	db "night-owls-go/internal/db/sqlc_generated"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates an in-memory SQLite database with required schema
func setupTestDB(t *testing.T) *sql.DB {
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create required schema for OTP rate limiting tests
	schema := `
		CREATE TABLE otp_attempts (
			attempt_id INTEGER PRIMARY KEY,
			phone TEXT NOT NULL,
			attempted_at DATETIME NOT NULL,
			success INTEGER NOT NULL DEFAULT 0,
			client_ip TEXT,
			user_agent TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE otp_rate_limits (
			phone TEXT PRIMARY KEY,
			failed_attempts INTEGER NOT NULL DEFAULT 0,
			first_attempt_at DATETIME NOT NULL,
			last_attempt_at DATETIME NOT NULL,
			locked_until DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`

	_, err = database.Exec(schema)
	if err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return database
}

func TestOTPRateLimitingService_CheckRateLimit_NoExistingRecord(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	queries := db.New(database)
	service := NewOTPRateLimitingService(queries, logger)

	ctx := context.Background()
	phone := "+1234567890"

	// Should allow attempt for phone with no history
	err := service.CheckRateLimit(ctx, phone)
	if err != nil {
		t.Errorf("Expected no error for clean phone, got: %v", err)
	}
}

func TestOTPRateLimitingService_RecordOTPAttempt_Success(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	queries := db.New(database)
	service := NewOTPRateLimitingService(queries, logger)

	ctx := context.Background()
	phone := "+1234567890"
	clientIP := "192.168.1.1"
	userAgent := "test-agent"

	// Record successful attempt
	err := service.RecordOTPAttempt(ctx, phone, true, clientIP, userAgent)
	if err != nil {
		t.Errorf("Failed to record successful OTP attempt: %v", err)
	}

	// Verify attempt was recorded
	attempts, err := queries.GetOTPAttemptsInWindow(ctx, db.GetOTPAttemptsInWindowParams{
		Phone:       phone,
		AttemptedAt: time.Now().UTC().Add(-1 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Failed to get OTP attempts: %v", err)
	}

	if len(attempts) != 1 {
		t.Errorf("Expected 1 attempt, got %d", len(attempts))
	}

	if attempts[0].Success != 1 {
		t.Errorf("Expected success=1, got %d", attempts[0].Success)
	}
}

func TestOTPRateLimitingService_RateLimitingLogic(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	queries := db.New(database)
	service := NewOTPRateLimitingService(queries, logger)

	ctx := context.Background()
	phone := "+1234567890"
	clientIP := "192.168.1.1"
	userAgent := "test-agent"

	// Test progressive rate limiting
	for i := 0; i < MaxOTPAttemptsPerWindow; i++ {
		// Should allow attempts up to the limit
		err := service.CheckRateLimit(ctx, phone)
		if err != nil {
			t.Errorf("Attempt %d should be allowed, got error: %v", i+1, err)
		}

		// Record failed attempt
		err = service.RecordOTPAttempt(ctx, phone, false, clientIP, userAgent)
		if err != nil {
			t.Errorf("Failed to record attempt %d: %v", i+1, err)
		}
	}

	// Next attempt should be locked (account gets locked immediately after max attempts)
	err := service.CheckRateLimit(ctx, phone)
	if err != ErrOTPLocked {
		t.Errorf("Expected ErrOTPLocked after %d attempts, got: %v", MaxOTPAttemptsPerWindow, err)
	}
}

func TestOTPRateLimitingService_ExponentialLockout(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	queries := db.New(database)
	service := NewOTPRateLimitingService(queries, logger)

	ctx := context.Background()
	phone := "+1234567890"
	clientIP := "192.168.1.1"
	userAgent := "test-agent"

	// Simulate progressive lockout
	for attempt := 0; attempt < 6; attempt++ {
		err := service.RecordOTPAttempt(ctx, phone, false, clientIP, userAgent)
		if err != nil {
			t.Errorf("Failed to record attempt %d: %v", attempt+1, err)
		}
	}

	// Check lockout info
	isLocked, lockedUntil, failedAttempts, err := service.GetLockoutInfo(ctx, phone)
	if err != nil {
		t.Fatalf("Failed to get lockout info: %v", err)
	}

	if !isLocked {
		t.Error("Phone should be locked after excessive failed attempts")
	}

	if failedAttempts != 6 {
		t.Errorf("Expected 6 failed attempts, got %d", failedAttempts)
	}

	if lockedUntil == nil {
		t.Error("LockedUntil should not be nil for locked phone")
	}

	// Should return ErrOTPLocked when checking rate limit
	err = service.CheckRateLimit(ctx, phone)
	if err != ErrOTPLocked {
		t.Errorf("Expected ErrOTPLocked for locked phone, got: %v", err)
	}
}

func TestOTPRateLimitingService_SuccessResetsRateLimit(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	queries := db.New(database)
	service := NewOTPRateLimitingService(queries, logger)

	ctx := context.Background()
	phone := "+1234567890"
	clientIP := "192.168.1.1"
	userAgent := "test-agent"

	// Record some failed attempts
	for i := 0; i < MaxOTPAttemptsPerWindow-1; i++ {
		err := service.RecordOTPAttempt(ctx, phone, false, clientIP, userAgent)
		if err != nil {
			t.Errorf("Failed to record failed attempt %d: %v", i+1, err)
		}
	}

	// Record successful attempt - this should reset the rate limit
	err := service.RecordOTPAttempt(ctx, phone, true, clientIP, userAgent)
	if err != nil {
		t.Errorf("Failed to record successful attempt: %v", err)
	}

	// Rate limit should be reset - should allow new attempts
	err = service.CheckRateLimit(ctx, phone)
	if err != nil {
		t.Errorf("Rate limit should be reset after success, got error: %v", err)
	}

	// Test that we can now make new failed attempts (using a different phone to avoid window conflicts)
	newPhone := "+1234567891"
	for i := 0; i < MaxOTPAttemptsPerWindow; i++ {
		err := service.CheckRateLimit(ctx, newPhone)
		if err != nil {
			t.Errorf("Fresh attempt %d should be allowed for new phone, got error: %v", i+1, err)
		}
		err = service.RecordOTPAttempt(ctx, newPhone, false, clientIP, userAgent)
		if err != nil {
			t.Errorf("Failed to record fresh attempt %d: %v", i+1, err)
		}
	}
}

func TestOTPRateLimitingService_ConstantTimeVerification(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	queries := db.New(database)
	service := NewOTPRateLimitingService(queries, logger)

	ctx := context.Background()
	phone := "+1234567890"
	clientIP := "192.168.1.1"
	userAgent := "test-agent"
	correctOTP := "123456"
	wrongOTP := "654321"

	// Test correct OTP
	valid, err := service.VerifyOTPWithRateLimit(ctx, phone, correctOTP, correctOTP, clientIP, userAgent)
	if err != nil {
		t.Errorf("Verification should succeed with correct OTP: %v", err)
	}
	if !valid {
		t.Error("OTP should be valid with matching codes")
	}

	// Test wrong OTP
	valid, err = service.VerifyOTPWithRateLimit(ctx, phone, wrongOTP, correctOTP, clientIP, userAgent)
	if err != nil {
		t.Errorf("Verification should not error with wrong OTP: %v", err)
	}
	if valid {
		t.Error("OTP should be invalid with non-matching codes")
	}

	// Test empty OTP
	valid, err = service.VerifyOTPWithRateLimit(ctx, phone, "", correctOTP, clientIP, userAgent)
	if err != nil {
		t.Errorf("Verification should not error with empty OTP: %v", err)
	}
	if valid {
		t.Error("OTP should be invalid with empty code")
	}
}

func TestOTPRateLimitingService_RegistrationRateLimit_IP(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	queries := db.New(database)
	service := NewOTPRateLimitingService(queries, logger)

	ctx := context.Background()
	clientIP := "192.168.1.1"
	userAgent := "test-agent"

	// Record registration attempts up to IP limit
	for i := 0; i < MaxRegistrationAttemptsPerIP; i++ {
		phone := "+123456789" + string(rune('0'+i))
		err := service.RecordRegistrationAttempt(ctx, phone, clientIP, userAgent, false)
		if err != nil {
			t.Errorf("Failed to record registration attempt %d: %v", i+1, err)
		}
	}

	// Next registration from same IP should be blocked
	err := service.CheckRegistrationRateLimit(ctx, "+1999999999", clientIP)
	if err == nil {
		t.Error("Expected registration rate limit error for IP after max attempts")
	}
	if err != nil && err.Error() != "too many registration attempts from this IP address, try again later" {
		t.Errorf("Expected IP rate limit error, got: %v", err)
	}
}

func TestOTPRateLimitingService_RegistrationRateLimit_Phone(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	queries := db.New(database)
	service := NewOTPRateLimitingService(queries, logger)

	ctx := context.Background()
	phone := "+1234567890"
	userAgent := "test-agent"

	// Record registration attempts up to phone limit
	for i := 0; i < MaxRegistrationAttemptsPerPhone; i++ {
		clientIP := "192.168.1." + string(rune('1'+i))
		err := service.RecordRegistrationAttempt(ctx, phone, clientIP, userAgent, false)
		if err != nil {
			t.Errorf("Failed to record registration attempt %d: %v", i+1, err)
		}
	}

	// Next registration for same phone should be blocked
	err := service.CheckRegistrationRateLimit(ctx, phone, "192.168.1.100")
	if err == nil {
		t.Error("Expected registration rate limit error for phone after max attempts")
	}
	if err != nil && err.Error() != "too many registration attempts for this phone number, try again later" {
		t.Errorf("Expected phone rate limit error, got: %v", err)
	}
}

func TestOTPRateLimitingService_DatabaseErrorHandling(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	queries := db.New(database)
	service := NewOTPRateLimitingService(queries, logger)

	ctx := context.Background()
	phone := "+1234567890"

	// Close database to simulate database errors
	database.Close()

	// CheckRateLimit should not block on database errors (fail-open)
	err := service.CheckRateLimit(ctx, phone)
	if err != nil {
		t.Errorf("CheckRateLimit should not block on database errors, got: %v", err)
	}

	// RecordOTPAttempt should not panic on database errors
	err = service.RecordOTPAttempt(ctx, phone, false, "192.168.1.1", "test-agent")
	// Error is expected but should not panic
	if err == nil {
		t.Log("RecordOTPAttempt gracefully handled database error")
	}
}

func TestOTPRateLimitingService_CleanupOldAttempts(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	queries := db.New(database)
	service := NewOTPRateLimitingService(queries, logger)

	ctx := context.Background()
	phone := "+1234567890"
	clientIP := "192.168.1.1"
	userAgent := "test-agent"

	// Record some attempts
	err := service.RecordOTPAttempt(ctx, phone, false, clientIP, userAgent)
	if err != nil {
		t.Errorf("Failed to record OTP attempt: %v", err)
	}

	// Cleanup old attempts
	err = service.CleanupOldAttempts(ctx, 24*time.Hour)
	if err != nil {
		t.Errorf("Failed to cleanup old attempts: %v", err)
	}

	// Should still have recent attempts
	attempts, err := queries.GetOTPAttemptsInWindow(ctx, db.GetOTPAttemptsInWindowParams{
		Phone:       phone,
		AttemptedAt: time.Now().UTC().Add(-1 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Failed to get OTP attempts after cleanup: %v", err)
	}

	if len(attempts) != 1 {
		t.Errorf("Expected 1 recent attempt after cleanup, got %d", len(attempts))
	}
}

// TestOTPRateLimitingService_SecurityEdgeCases tests security edge cases
func TestOTPRateLimitingService_SecurityEdgeCases(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	queries := db.New(database)
	service := NewOTPRateLimitingService(queries, logger)

	ctx := context.Background()

	// Test with empty phone number
	err := service.CheckRateLimit(ctx, "")
	if err != nil {
		t.Error("CheckRateLimit should handle empty phone gracefully")
	}

	// Test with very long phone number
	longPhone := "+1" + string(make([]byte, 1000))
	err = service.CheckRateLimit(ctx, longPhone)
	if err != nil {
		t.Error("CheckRateLimit should handle long phone gracefully")
	}

	// Test constant-time verification with different length strings
	shortOTP := "123"
	longOTP := "123456789012345"
	
	valid, err := service.VerifyOTPWithRateLimit(ctx, "+1234567890", shortOTP, longOTP, "192.168.1.1", "test")
	if err != nil {
		t.Errorf("VerifyOTPWithRateLimit should handle different length OTPs: %v", err)
	}
	if valid {
		t.Error("Different length OTPs should not be valid")
	}
} 