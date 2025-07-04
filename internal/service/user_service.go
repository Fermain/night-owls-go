package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"night-owls-go/internal/auth"
	"night-owls-go/internal/config"
	db "night-owls-go/internal/db/sqlc_generated"
	"night-owls-go/internal/otp"
)

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrOTPValidationFailed = errors.New("otp validation failed")
	ErrInternalServer      = errors.New("internal server error")
)

// immediateSendAt returns a time that ensures immediate dispatch by the outbox processor.
// Uses a small negative offset to ensure the message is processed immediately.
// Returns UTC time to match SQLite's CURRENT_TIMESTAMP behavior.
func immediateSendAt() time.Time {
	return time.Now().UTC().Add(-1 * time.Second)
}

// JWTGenerator defines a function that generates a JWT token
type JWTGenerator func(userID int64, phone string, name string, role string, secret string, expiryHours int) (string, error)

// UserService handles user registration, login, and OTP verification.
type UserService struct {
	querier           db.Querier // From sqlc
	otpStore          auth.OTPStore
	twilioOTP         *otp.Client // Twilio OTP client for real SMS
	jwtSecret         string
	otpLogPath        string
	logger            *slog.Logger
	cfg               *config.Config
	jwtGenerator      JWTGenerator // Function for JWT generation, allows mocking in tests
	otpRateLimitSvc   *OTPRateLimitingService // Rate limiting service
}

// NewUserService creates a new UserService.
func NewUserService(querier db.Querier, otpStore auth.OTPStore, cfg *config.Config, logger *slog.Logger) *UserService {
	var twilioOTP *otp.Client

	// Initialize Twilio OTP client if credentials are provided
	if cfg.TwilioAccountSID != "" && cfg.TwilioAuthToken != "" && cfg.TwilioVerifySID != "" {
		twilioOTP = otp.New(cfg.TwilioAccountSID, cfg.TwilioAuthToken, cfg.TwilioVerifySID)
		logger.Info("Twilio OTP client initialized for real SMS verification")
	} else {
		logger.Info("Twilio credentials not configured, using mock OTP flow")
	}

	// Initialize OTP rate limiting service
	otpRateLimitSvc := NewOTPRateLimitingService(querier, logger)

	return &UserService{
		querier:           querier,
		otpStore:          otpStore,
		twilioOTP:         twilioOTP,
		jwtSecret:         cfg.JWTSecret,
		otpLogPath:        cfg.OTPLogPath,
		logger:            logger.With("service", "UserService"),
		cfg:               cfg,
		jwtGenerator:      auth.GenerateJWT, // Use the real implementation by default
		otpRateLimitSvc:   otpRateLimitSvc,
	}
}

// SetJWTGenerator allows tests to inject a custom JWT generator
func (s *UserService) SetJWTGenerator(generator JWTGenerator) {
	s.jwtGenerator = generator
}

// RegisterOrLoginUser handles creating/finding a user, generating an OTP, storing it,
// and queuing an OTP message to the (mocked) outbox.
// If name is provided, this is a registration attempt and will create a new user.
// If name is empty, this is a login attempt and will fail if user doesn't exist.
// Now includes registration rate limiting to prevent abuse.
func (s *UserService) RegisterOrLoginUser(ctx context.Context, phone string, name sql.NullString, clientIP, userAgent string) error {
	// Use provided client info, fallback to unknown if not provided
	if clientIP == "" {
		clientIP = "unknown"
	}
	if userAgent == "" {
		userAgent = "unknown"
	}
	
	// Check registration rate limits first
	if err := s.otpRateLimitSvc.CheckRegistrationRateLimit(ctx, phone, clientIP); err != nil {
		s.logger.WarnContext(ctx, "Registration attempt blocked by rate limiting", "phone", phone, "error", err)
		// Record failed attempt for audit
		_ = s.otpRateLimitSvc.RecordRegistrationAttempt(ctx, phone, clientIP, userAgent, false)
		return fmt.Errorf("registration rate limit exceeded: %w", err)
	}

	// Debug logging to understand the name parameter
	s.logger.InfoContext(ctx, "RegisterOrLoginUser called", "phone", phone, "name_valid", name.Valid, "name_string", name.String)

	user, err := s.querier.GetUserByPhone(ctx, phone)
	registrationSuccess := false // Track if registration/login was successful
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// User does not exist
			s.logger.InfoContext(ctx, "User does not exist", "phone", phone, "name_valid", name.Valid, "name_string", name.String)

			if !name.Valid || name.String == "" {
				// This is a login attempt (no name provided) but user doesn't exist
				s.logger.WarnContext(ctx, "Login attempt for non-existent user", "phone", phone)
				// Record failed registration attempt
				_ = s.otpRateLimitSvc.RecordRegistrationAttempt(ctx, phone, clientIP, userAgent, false)
				return errors.New("user not found - please register first")
			}

			// This is a registration attempt (name provided), create new user
			s.logger.InfoContext(ctx, "Creating new user during registration", "phone", phone, "name", name.String)

			// Smart role assignment: First two users get admin, everyone else gets guest
			defaultRole, err := s.determineRoleForNewUser(ctx)
			if err != nil {
				s.logger.ErrorContext(ctx, "Failed to determine role for new user", "phone", phone, "error", err)
				// Record failed registration attempt
				_ = s.otpRateLimitSvc.RecordRegistrationAttempt(ctx, phone, clientIP, userAgent, false)
				return ErrInternalServer
			}

			createUserParams := db.CreateUserParams{
				Phone: phone,
				Name:  name,        // sql.NullString handles optional name
				Role:  defaultRole, // Pass role as interface{} (string)
			}
			createResult, err := s.querier.CreateUser(ctx, createUserParams)
			if err != nil {
				s.logger.ErrorContext(ctx, "Failed to create user", "phone", phone, "error", err)
				// Record failed registration attempt
				_ = s.otpRateLimitSvc.RecordRegistrationAttempt(ctx, phone, clientIP, userAgent, false)
				return ErrInternalServer
			}

			// Convert CreateUserRow to GetUserByPhoneRow since they have the same structure
			user = db.GetUserByPhoneRow{
				UserID:    createResult.UserID,
				Phone:     createResult.Phone,
				Name:      createResult.Name,
				CreatedAt: createResult.CreatedAt,
				Role:      createResult.Role,
			}

			s.logger.InfoContext(ctx, "New user created during registration", "phone", phone, "user_id", user.UserID, "role", defaultRole, "name", name.String)
			registrationSuccess = true
		} else {
			s.logger.ErrorContext(ctx, "Failed to get user by phone", "phone", phone, "error", err)
			// Record failed registration attempt
			_ = s.otpRateLimitSvc.RecordRegistrationAttempt(ctx, phone, clientIP, userAgent, false)
			return ErrInternalServer
		}
	} else {
		// User exists - log whether this is login or registration attempt
		if name.Valid && name.String != "" {
			s.logger.InfoContext(ctx, "Registration attempt for existing user", "phone", phone, "user_id", user.UserID)
		} else {
			s.logger.InfoContext(ctx, "Login attempt for existing user", "phone", phone, "user_id", user.UserID)
		}
		registrationSuccess = true // User exists, so registration/login is valid
	}

	// Send OTP - use Twilio if configured, otherwise fall back to mock flow
	otpSendSuccess := false
	if s.twilioOTP != nil {
		// Use Twilio Verify to send real SMS OTP
		s.logger.InfoContext(ctx, "Attempting to send OTP via Twilio", "phone", phone)
		err = s.twilioOTP.StartSMS(ctx, phone)
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to send Twilio OTP", "phone", phone, "error", err)
			// Record failed registration attempt (OTP send failed)
			_ = s.otpRateLimitSvc.RecordRegistrationAttempt(ctx, phone, clientIP, userAgent, false)
			return fmt.Errorf("failed to send SMS: %w", err)
		}
		s.logger.InfoContext(ctx, "Twilio OTP sent successfully", "phone", phone)
		otpSendSuccess = true

		// For Twilio, we don't store OTP locally or use outbox since Twilio manages it
	} else {
		// Fall back to mock OTP flow for development/testing
		s.logger.InfoContext(ctx, "Using mock OTP flow (Twilio not configured)", "phone", phone)
		otp, err := auth.GenerateOTP()
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to generate OTP", "phone", phone, "error", err)
			// Record failed registration attempt (OTP generation failed)
			_ = s.otpRateLimitSvc.RecordRegistrationAttempt(ctx, phone, clientIP, userAgent, false)
			return ErrInternalServer
		}

		otpValidityDuration := time.Duration(s.cfg.OTPValidityMinutes) * time.Minute
		s.otpStore.StoreOTP(phone, otp, otpValidityDuration)
		s.logger.DebugContext(ctx, "Mock OTP generated and stored for user", "phone", phone, "validity_minutes", s.cfg.OTPValidityMinutes)

		// Queue OTP message to outbox for mock SMS
		outboxPayload := fmt.Sprintf(`{"otp": "%s"}`, otp)
		_, err = s.querier.CreateOutboxItem(ctx, db.CreateOutboxItemParams{
			MessageType: "OTP_VERIFICATION",
			Recipient:   phone,
			Payload:     sql.NullString{String: outboxPayload, Valid: true},
			SendAt:      immediateSendAt(),
		})
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to create outbox item for mock OTP", "phone", phone, "error", err)
			// Non-fatal for OTP sending itself, but log it
		}
		otpSendSuccess = true
	}

	// Record successful registration attempt (both user creation/validation and OTP send succeeded)
	if registrationSuccess && otpSendSuccess {
		_ = s.otpRateLimitSvc.RecordRegistrationAttempt(ctx, phone, clientIP, userAgent, true)
	}

	return nil
}

// VerifyOTP validates the OTP for a given phone number and if valid, generates a JWT.
// Now includes rate limiting and brute force protection.
func (s *UserService) VerifyOTP(ctx context.Context, phone string, otpToValidate string) (string, error) {
	// Extract client info for audit logging (basic implementation)
	clientIP := "unknown"
	userAgent := "unknown"
	
	var otpValid bool

	// Verify OTP - use Twilio if configured, otherwise use local store with rate limiting
	if s.twilioOTP != nil {
		// For Twilio, check rate limit first
		if err := s.otpRateLimitSvc.CheckRateLimit(ctx, phone); err != nil {
			s.logger.WarnContext(ctx, "OTP verification blocked by rate limiting", "phone", phone, "error", err)
			return "", err
		}

		// Use Twilio Verify to check the OTP
		valid, err := s.twilioOTP.Check(ctx, phone, otpToValidate)
		if err != nil {
			s.logger.ErrorContext(ctx, "Twilio OTP verification failed", "phone", phone, "error", err)
			// Record failed attempt
			_ = s.otpRateLimitSvc.RecordOTPAttempt(ctx, phone, false, clientIP, userAgent)
			return "", ErrOTPValidationFailed
		}
		
		otpValid = valid
		// Record the attempt (success or failure)
		_ = s.otpRateLimitSvc.RecordOTPAttempt(ctx, phone, otpValid, clientIP, userAgent)
		
		if otpValid {
			s.logger.InfoContext(ctx, "Twilio OTP verified successfully", "phone", phone)
		} else {
			s.logger.WarnContext(ctx, "Twilio OTP validation failed - invalid code", "phone", phone)
		}
	} else {
		// Fall back to local OTP store for development/testing with rate limiting
		
		// Check rate limit first
		if err := s.otpRateLimitSvc.CheckRateLimit(ctx, phone); err != nil {
			s.logger.WarnContext(ctx, "Mock OTP verification blocked by rate limiting", "phone", phone, "error", err)
			return "", err
		}

		// Use existing ValidateOTP method (which already does constant-time comparison internally)
		otpValid = s.otpStore.ValidateOTP(phone, otpToValidate)
		
		// Record the attempt (success or failure)
		_ = s.otpRateLimitSvc.RecordOTPAttempt(ctx, phone, otpValid, clientIP, userAgent)
		
		if otpValid {
			s.logger.InfoContext(ctx, "Mock OTP verified successfully", "phone", phone)
			// Note: ValidateOTP already clears the OTP on success
		} else {
			s.logger.WarnContext(ctx, "Mock OTP validation failed", "phone", phone)
		}
	}

	if !otpValid {
		return "", ErrOTPValidationFailed
	}

	// OTP is valid, get user details to include in JWT
	user, err := s.querier.GetUserByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.ErrorContext(ctx, "User not found after OTP validation", "phone", phone, "error", err)
			return "", ErrUserNotFound // Should not happen if OTP was stored for this phone
		}
		s.logger.ErrorContext(ctx, "Failed to get user by phone after OTP validation", "phone", phone, "error", err)
		return "", ErrInternalServer
	}

	// Generate JWT (e.g., valid for 24 hours)
	// The expiration duration could also come from config.
	userName := ""
	if user.Name.Valid {
		userName = user.Name.String
	}
	tokenString, err := s.jwtGenerator(user.UserID, user.Phone, userName, user.Role, s.jwtSecret, s.cfg.JWTExpirationHours) // Use from config
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to generate JWT", "phone", phone, "user_id", user.UserID, "error", err)
		return "", ErrInternalServer
	}

	s.logger.InfoContext(ctx, "OTP validated and JWT generated", "phone", phone, "user_id", user.UserID)
	return tokenString, nil
}

// determineRoleForNewUser determines the role for a new user based on business logic:
// - First two users get admin role (for initial setup)
// - Everyone else gets guest role (must be manually promoted to owl by admins)
func (s *UserService) determineRoleForNewUser(ctx context.Context) (string, error) {
	// Count existing users to determine if this should be an admin
	userCount, err := s.querier.CountUsers(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to count users for role determination", "error", err)
		return "", err
	}

	// First two users get admin role for initial system setup
	if userCount < 2 {
		s.logger.InfoContext(ctx, "Assigning admin role to initial user", "user_count", userCount)
		return "admin", nil
	}

	// Everyone else gets guest role by default
	s.logger.InfoContext(ctx, "Assigning guest role to new user", "user_count", userCount)
	return "guest", nil
}
