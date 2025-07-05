package service

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	db "night-owls-go/internal/db/sqlc_generated"
)

var (
	ErrOTPRateLimited = errors.New("too many OTP attempts, account temporarily locked")
	ErrOTPLocked      = errors.New("account locked due to suspicious activity")
)

// OTP Rate Limiting Configuration
const (
	MaxOTPAttemptsPerWindow = 3        // Max attempts before lockout
	OTPAttemptWindow        = 15 * time.Minute // Time window for counting attempts
	InitialLockoutDuration  = 30 * time.Minute // First lockout duration
	MaxLockoutDuration      = 24 * time.Hour   // Maximum lockout duration
	LockoutMultiplier       = 2                // Exponential backoff multiplier
	
	// Registration rate limiting
	MaxRegistrationAttemptsPerIP    = 10       // Max registration attempts per IP per hour
	MaxRegistrationAttemptsPerPhone = 3        // Max registration attempts per phone per hour  
	RegistrationWindow              = 1 * time.Hour // Time window for registration attempts
)

type OTPRateLimitingService struct {
	querier db.Querier
	logger  *slog.Logger
}

func NewOTPRateLimitingService(querier db.Querier, logger *slog.Logger) *OTPRateLimitingService {
	return &OTPRateLimitingService{
		querier: querier,
		logger:  logger.With("service", "OTPRateLimiting"),
	}
}

// CheckRateLimit verifies if a phone number is allowed to attempt OTP verification
func (s *OTPRateLimitingService) CheckRateLimit(ctx context.Context, phone string) error {
	// Clean up expired locks first
	err := s.querier.CleanupExpiredLocks(ctx)
	if err != nil {
		s.logger.WarnContext(ctx, "Failed to cleanup expired locks", "error", err)
		// Continue anyway, don't block on cleanup failure
	}

	// Check if phone is currently locked
	rateLimit, err := s.querier.GetOTPRateLimit(ctx, phone)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.logger.ErrorContext(ctx, "Failed to get OTP rate limit", "phone", phone, "error", err)
		// Allow attempt on database error to avoid false lockouts
		return nil
	}

	// If no rate limit record exists, allow the attempt
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}

	// Check if currently locked
	if rateLimit.LockedUntil.Valid && rateLimit.LockedUntil.Time.After(time.Now().UTC()) {
		s.logger.WarnContext(ctx, "OTP attempt blocked - account locked", 
			"phone", phone, 
			"locked_until", rateLimit.LockedUntil.Time,
			"failed_attempts", rateLimit.FailedAttempts,
		)
		return ErrOTPLocked
	}

	// Check attempts in current window
	windowStart := time.Now().UTC().Add(-OTPAttemptWindow)
	failedCount, err := s.querier.GetFailedOTPAttemptsInWindow(ctx, db.GetFailedOTPAttemptsInWindowParams{
		Phone:        phone,
		AttemptedAt: windowStart,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to get failed OTP attempts", "phone", phone, "error", err)
		// Allow attempt on database error
		return nil
	}

	if failedCount >= MaxOTPAttemptsPerWindow {
		s.logger.WarnContext(ctx, "OTP attempt blocked - rate limit exceeded", 
			"phone", phone, 
			"failed_count", failedCount,
			"max_attempts", MaxOTPAttemptsPerWindow,
		)
		return ErrOTPRateLimited
	}

	return nil
}

// RecordOTPAttempt records an OTP verification attempt and handles rate limiting
func (s *OTPRateLimitingService) RecordOTPAttempt(ctx context.Context, phone string, success bool, clientIP, userAgent string) error {
	now := time.Now().UTC()

	// Record the attempt
	_, err := s.querier.CreateOTPAttempt(ctx, db.CreateOTPAttemptParams{
		Phone:       phone,
		AttemptedAt: now,
		Success:     boolToInt(success),
		ClientIp:    sql.NullString{String: clientIP, Valid: clientIP != ""},
		UserAgent:   sql.NullString{String: userAgent, Valid: userAgent != ""},
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to record OTP attempt", "phone", phone, "error", err)
		// Continue with rate limiting logic even if recording fails
	}

	if success {
		// Reset rate limit on successful verification
		err = s.querier.ResetOTPRateLimit(ctx, phone)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			s.logger.WarnContext(ctx, "Failed to reset OTP rate limit after success", "phone", phone, "error", err)
		}
		s.logger.InfoContext(ctx, "OTP verification successful, rate limit reset", "phone", phone)
		return nil
	}

	// Handle failed attempt - update or create rate limit record
	rateLimit, err := s.querier.GetOTPRateLimit(ctx, phone)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.logger.ErrorContext(ctx, "Failed to get OTP rate limit for failed attempt", "phone", phone, "error", err)
		return nil // Don't block user on database error
	}

	var newFailedAttempts int64
	var lockoutDuration time.Duration

	if errors.Is(err, sql.ErrNoRows) {
		// Create new rate limit record
		newFailedAttempts = 1
		_, err = s.querier.CreateOTPRateLimit(ctx, db.CreateOTPRateLimitParams{
			Phone:           phone,
			FailedAttempts:  newFailedAttempts,
			FirstAttemptAt:  now,
			LastAttemptAt:   now,
		})
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to create OTP rate limit", "phone", phone, "error", err)
		}
	} else {
		// Update existing record
		newFailedAttempts = rateLimit.FailedAttempts + 1
		
		// Calculate lockout duration with exponential backoff
		if newFailedAttempts >= MaxOTPAttemptsPerWindow {
			// Calculate progressive lockout duration
			lockoutMultiplier := int(newFailedAttempts - MaxOTPAttemptsPerWindow + 1)
			lockoutDuration = InitialLockoutDuration
			for i := 1; i < lockoutMultiplier; i++ {
				lockoutDuration *= LockoutMultiplier
				if lockoutDuration > MaxLockoutDuration {
					lockoutDuration = MaxLockoutDuration
					break
				}
			}
		}

		var lockedUntil sql.NullTime
		if lockoutDuration > 0 {
			lockedUntil = sql.NullTime{
				Time:  now.Add(lockoutDuration),
				Valid: true,
			}
		}

		err = s.querier.UpdateOTPRateLimit(ctx, db.UpdateOTPRateLimitParams{
			FailedAttempts: newFailedAttempts,
			LockedUntil:    lockedUntil,
			LastAttemptAt:  now,
			Phone:          phone,
		})
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to update OTP rate limit", "phone", phone, "error", err)
		}
	}

	if lockoutDuration > 0 {
		s.logger.WarnContext(ctx, "Phone locked due to failed OTP attempts", 
			"phone", phone,
			"failed_attempts", newFailedAttempts,
			"lockout_duration", lockoutDuration,
			"locked_until", now.Add(lockoutDuration),
		)
	} else {
		s.logger.InfoContext(ctx, "Failed OTP attempt recorded", 
			"phone", phone,
			"failed_attempts", newFailedAttempts,
		)
	}

	return nil
}

// VerifyOTPWithRateLimit performs constant-time OTP verification with rate limiting
func (s *OTPRateLimitingService) VerifyOTPWithRateLimit(ctx context.Context, phone string, providedOTP, expectedOTP string, clientIP, userAgent string) (bool, error) {
	// Check rate limit first
	if err := s.CheckRateLimit(ctx, phone); err != nil {
		// Still record the attempt for audit purposes
		_ = s.RecordOTPAttempt(ctx, phone, false, clientIP, userAgent)
		return false, err
	}

	// Perform constant-time comparison to prevent timing attacks
	otpValid := subtle.ConstantTimeCompare([]byte(expectedOTP), []byte(providedOTP)) == 1

	// Record the attempt
	if err := s.RecordOTPAttempt(ctx, phone, otpValid, clientIP, userAgent); err != nil {
		s.logger.WarnContext(ctx, "Failed to record OTP attempt", "phone", phone, "error", err)
	}

	return otpValid, nil
}

// CleanupOldAttempts removes old OTP attempts for maintenance
func (s *OTPRateLimitingService) CleanupOldAttempts(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().UTC().Add(-olderThan)
	err := s.querier.CleanupOldOTPAttempts(ctx, cutoff)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to cleanup old OTP attempts", "cutoff", cutoff, "error", err)
		return err
	}
	
	s.logger.InfoContext(ctx, "Cleaned up old OTP attempts", "cutoff", cutoff)
	return nil
}

// GetLockoutInfo returns current lockout information for a phone number
func (s *OTPRateLimitingService) GetLockoutInfo(ctx context.Context, phone string) (isLocked bool, lockedUntil *time.Time, failedAttempts int64, err error) {
	rateLimit, err := s.querier.GetOTPRateLimit(ctx, phone)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil, 0, nil
		}
		return false, nil, 0, err
	}

	isLocked = rateLimit.LockedUntil.Valid && rateLimit.LockedUntil.Time.After(time.Now().UTC())
	var lockedUntilPtr *time.Time
	if rateLimit.LockedUntil.Valid {
		lockedUntilPtr = &rateLimit.LockedUntil.Time
	}

	return isLocked, lockedUntilPtr, rateLimit.FailedAttempts, nil
}

// boolToInt converts boolean to integer for database storage
func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// CheckRegistrationRateLimit verifies if registration attempts are allowed from IP and for phone
// 
// DESIGN NOTE: This function reuses the existing otp_attempts table for both phone and IP rate limiting
// by storing IP addresses in the phone column for IP-based records. This avoids creating separate tables
// while maintaining the same rate limiting logic. IP records are distinguishable by context and the 
// client_ip field always contains the actual IP address for audit purposes.
func (s *OTPRateLimitingService) CheckRegistrationRateLimit(ctx context.Context, phone, clientIP string) error {
	windowStart := time.Now().UTC().Add(-RegistrationWindow)

	// Check IP-based rate limiting (prevent spam from single source)
	// NOTE: Using "IP:" prefix to clearly distinguish IP records from phone records
	if clientIP != "" {
		ipKey := "IP:" + clientIP // Use clear prefix to distinguish IP records from phone records
		ipAttempts, err := s.querier.GetOTPAttemptsInWindow(ctx, db.GetOTPAttemptsInWindowParams{
			Phone:       ipKey,
			AttemptedAt: windowStart,
		})
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to get IP registration attempts", "ip", clientIP, "error", err)
			// Allow on database error to avoid false lockouts
		} else if len(ipAttempts) >= MaxRegistrationAttemptsPerIP {
			s.logger.WarnContext(ctx, "Registration rate limit exceeded for IP", 
				"ip", clientIP, 
				"attempts", len(ipAttempts),
				"max_attempts", MaxRegistrationAttemptsPerIP,
			)
			return fmt.Errorf("too many registration attempts from this IP address, try again later")
		}
	}

	// Check phone-based rate limiting (prevent targeting specific numbers)  
	phoneAttempts, err := s.querier.GetOTPAttemptsInWindow(ctx, db.GetOTPAttemptsInWindowParams{
		Phone:       phone,
		AttemptedAt: windowStart,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to get phone registration attempts", "phone", phone, "error", err)
		// Allow on database error
	} else if len(phoneAttempts) >= MaxRegistrationAttemptsPerPhone {
		s.logger.WarnContext(ctx, "Registration rate limit exceeded for phone", 
			"phone", phone, 
			"attempts", len(phoneAttempts),
			"max_attempts", MaxRegistrationAttemptsPerPhone,
		)
		return fmt.Errorf("too many registration attempts for this phone number, try again later")
	}

	return nil
}

// RecordRegistrationAttempt records a registration attempt for both IP and phone rate limiting
// 
// DESIGN NOTE: This creates two records - one for the phone number and one for the IP address.
// The IP record stores the IP address in the phone field to reuse the existing schema.
// Both records maintain the actual IP address in the client_ip field for proper audit trails.
func (s *OTPRateLimitingService) RecordRegistrationAttempt(ctx context.Context, phone, clientIP, userAgent string, success bool) error {
	now := time.Now().UTC()

	// Record phone-based attempt
	_, err := s.querier.CreateOTPAttempt(ctx, db.CreateOTPAttemptParams{
		Phone:       phone,
		AttemptedAt: now,
		Success:     boolToInt(success),
		ClientIp:    sql.NullString{String: clientIP, Valid: clientIP != ""},
		UserAgent:   sql.NullString{String: userAgent, Valid: userAgent != ""},
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to record phone registration attempt", "phone", phone, "error", err)
	}

	// Record IP-based attempt (if IP is available)
	if clientIP != "" {
		ipKey := "IP:" + clientIP // Use clear prefix to distinguish IP records from phone records
		_, err = s.querier.CreateOTPAttempt(ctx, db.CreateOTPAttemptParams{
			Phone:       ipKey,
			AttemptedAt: now,
			Success:     boolToInt(success),
			ClientIp:    sql.NullString{String: clientIP, Valid: true},
			UserAgent:   sql.NullString{String: userAgent, Valid: userAgent != ""},
		})
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to record IP registration attempt", "ip", clientIP, "error", err)
		}
	}

	s.logger.InfoContext(ctx, "Registration attempt recorded", 
		"phone", phone, 
		"ip", clientIP, 
		"success", success,
	)

	return nil
} 