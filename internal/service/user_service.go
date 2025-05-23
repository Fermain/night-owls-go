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
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrOTPValidationFailed = errors.New("otp validation failed")
	ErrInternalServer    = errors.New("internal server error")
)

// JWTGenerator defines a function that generates a JWT token
type JWTGenerator func(userID int64, phone string, role string, secret string, expiryHours int) (string, error)

// UserService handles user registration, login, and OTP verification.
type UserService struct {
	querier      db.Querier // From sqlc
	otpStore     auth.OTPStore
	jwtSecret    string
	otpLogPath   string
	logger       *slog.Logger
	cfg          *config.Config
	jwtGenerator JWTGenerator // Function for JWT generation, allows mocking in tests
}

// NewUserService creates a new UserService.
func NewUserService(querier db.Querier, otpStore auth.OTPStore, cfg *config.Config, logger *slog.Logger) *UserService {
	return &UserService{
		querier:      querier,
		otpStore:     otpStore,
		jwtSecret:    cfg.JWTSecret,
		otpLogPath:   cfg.OTPLogPath,
		logger:       logger.With("service", "UserService"),
		cfg:          cfg,
		jwtGenerator: auth.GenerateJWT, // Use the real implementation by default
	}
}

// SetJWTGenerator allows tests to inject a custom JWT generator
func (s *UserService) SetJWTGenerator(generator JWTGenerator) {
	s.jwtGenerator = generator
}

// RegisterOrLoginUser handles creating/finding a user, generating an OTP, storing it,
// and queuing an OTP message to the (mocked) outbox.
func (s *UserService) RegisterOrLoginUser(ctx context.Context, phone string, name sql.NullString) error {
	user, err := s.querier.GetUserByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// User does not exist, create new user
			// For development, make all new users admin by default
			// In production, you'd want more sophisticated role assignment logic
			defaultRole := "admin" // For development - all users are admin

			createUserParams := db.CreateUserParams{
				Phone: phone,
				Name:  name, // sql.NullString handles optional name
				Role:  defaultRole, // Pass role as interface{} (string)
			}
			user, err = s.querier.CreateUser(ctx, createUserParams)
			if err != nil {
				s.logger.ErrorContext(ctx, "Failed to create user", "phone", phone, "error", err)
				return ErrInternalServer
			}
			s.logger.InfoContext(ctx, "New user created", "phone", phone, "user_id", user.UserID, "role", defaultRole)
		} else {
			s.logger.ErrorContext(ctx, "Failed to get user by phone", "phone", phone, "error", err)
			return ErrInternalServer
		}
	}

	otp, err := auth.GenerateOTP()
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to generate OTP", "phone", phone, "error", err)
		return ErrInternalServer
	}

	otpValidityDuration := time.Duration(s.cfg.OTPValidityMinutes) * time.Minute // Use from config
	s.otpStore.StoreOTP(phone, otp, otpValidityDuration) 
	s.logger.DebugContext(ctx, "OTP generated and stored for user", "phone", phone, "validity_minutes", s.cfg.OTPValidityMinutes)

	// Queue OTP message to outbox (actual DB write)
	outboxPayload := fmt.Sprintf(`{"otp": "%s"}`, otp)
	_, err = s.querier.CreateOutboxItem(ctx, db.CreateOutboxItemParams{
		MessageType: "OTP_VERIFICATION",
		Recipient:   phone,
		Payload:     sql.NullString{String: outboxPayload, Valid: true},
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to create outbox item for OTP", "phone", phone, "error", err)
		// Non-fatal for OTP sending itself, but log it. The OTP is still in memory.
		// Depending on requirements, this could be a fatal error for the request.
	}

	// TODO: In a real system, we would NOT log the OTP here in production.
	// For development, one might write it to a specific file or use a very verbose debug log.
	// The plan mentions writing to `sms_outbox.log` - this service should not directly write to files.
	// The Outbox dispatcher will handle the "sending" (logging to file in our mock case).
	// The logging of OTP above (s.logger.InfoContext) is for dev console visibility.

	return nil
}

// VerifyOTP validates the OTP for a given phone number and if valid, generates a JWT.
func (s *UserService) VerifyOTP(ctx context.Context, phone string, otpToValidate string) (string, error) {
	if !s.otpStore.ValidateOTP(phone, otpToValidate) {
		s.logger.WarnContext(ctx, "OTP validation failed", "phone", phone)
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
	tokenString, err := s.jwtGenerator(user.UserID, user.Phone, user.Role, s.jwtSecret, s.cfg.JWTExpirationHours) // Use from config
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to generate JWT", "phone", phone, "user_id", user.UserID, "error", err)
		return "", ErrInternalServer
	}

	s.logger.InfoContext(ctx, "OTP validated and JWT generated", "phone", phone, "user_id", user.UserID)
	return tokenString, nil
} 