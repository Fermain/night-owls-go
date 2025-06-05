package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"night-owls-go/internal/auth"
	"night-owls-go/internal/config"
	db "night-owls-go/internal/db/sqlc_generated"
	"night-owls-go/internal/service"

	"github.com/nyaruka/phonenumbers"
)

var phoneRegex = regexp.MustCompile(`^\+[1-9]\d{6,14}$`)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	userService  *service.UserService
	auditService *service.AuditService
	logger       *slog.Logger
	config       *config.Config
	querier      db.Querier
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(userService *service.UserService, auditService *service.AuditService, logger *slog.Logger, cfg *config.Config, querier db.Querier) *AuthHandler {
	logger.Info("AuthHandler created with config", "dev_mode", cfg.DevMode, "server_port", cfg.ServerPort)
	return &AuthHandler{
		userService:  userService,
		auditService: auditService,
		logger:       logger.With("handler", "AuthHandler"),
		config:       cfg,
		querier:      querier,
	}
}

// RegisterRequest is the expected JSON for POST /auth/register.
type RegisterRequest struct {
	Phone string `json:"phone"`
	Name  string `json:"name,omitempty"`
}

// RegisterResponse is the JSON response from POST /auth/register.
type RegisterResponse struct {
	Message string `json:"message"`
	// Development only - include OTP for easier testing
	DevOTP string `json:"dev_otp,omitempty"`
}

// VerifyRequest is the expected JSON for POST /auth/verify.
type VerifyRequest struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

// VerifyResponse is the JSON response from POST /auth/verify.
type VerifyResponse struct {
	Token string `json:"token"`
}

// DevLoginRequest is the expected JSON for POST /auth/dev-login (dev mode only).
type DevLoginRequest struct {
	Phone string `json:"phone"`
}

// DevLoginResponse is the JSON response from POST /auth/dev-login (dev mode only).
type DevLoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID    int64  `json:"id"`
		Phone string `json:"phone"`
		Name  string `json:"name"`
		Role  string `json:"role"`
	} `json:"user"`
}

// RegisterHandler handles POST /auth/register
// @Summary Register a new user or request OTP for existing user
// @Description Registers a new user with phone number or starts login flow for existing user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "User registration details"
// @Success 200 {object} RegisterResponse "OTP sent successfully"
// @Failure 400 {object} ErrorResponse "Invalid phone number or request format"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.InfoContext(r.Context(), "RegisterHandler called with dev_mode", "dev_mode", h.config.DevMode)

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload", h.logger, "error", err.Error())
		return
	}
	defer r.Body.Close()

	if req.Phone == "" {
		RespondWithError(w, http.StatusBadRequest, "Phone number is required", h.logger)
		return
	}

	// Parse the phone number to E.164 format
	parsedNum, err := phonenumbers.Parse(req.Phone, "")
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid phone number format", h.logger, "error", err.Error())
		return
	}
	if !phonenumbers.IsValidNumber(parsedNum) {
		RespondWithError(w, http.StatusBadRequest, "Invalid phone number", h.logger)
		return
	}
	phoneE164 := phonenumbers.Format(parsedNum, phonenumbers.E164)

	// Create a NullString for name
	var sqlName sql.NullString
	if req.Name != "" {
		sqlName.String = req.Name
		sqlName.Valid = true
	}

	err = h.userService.RegisterOrLoginUser(r.Context(), phoneE164, sqlName)
	if err != nil {
		h.logger.InfoContext(r.Context(), "RegisterOrLoginUser error", "error_message", err.Error())

		// Check for specific user not found error
		if err.Error() == "user not found - please register first" {
			RespondWithError(w, http.StatusBadRequest, "user not found - please register first", h.logger, "error", err.Error())
		} else if strings.Contains(err.Error(), "failed to send SMS") {
			// Twilio-specific error
			RespondWithError(w, http.StatusInternalServerError, "Failed to send SMS verification code", h.logger, "error", err.Error())
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to register/login user", h.logger, "error", err.Error())
		}
		return
	}

	// Determine response message based on configuration
	var response RegisterResponse
	cfg := h.config
	if cfg.TwilioAccountSID != "" && cfg.TwilioAuthToken != "" && cfg.TwilioVerifySID != "" {
		response.Message = "Verification code sent via SMS"
		h.logger.InfoContext(r.Context(), "SMS sent via Twilio", "phone", phoneE164)
	} else {
		response.Message = "Verification code sent (development mode)"
		h.logger.InfoContext(r.Context(), "Using mock OTP flow", "phone", phoneE164)
	}

	// In development mode, provide helpful information about the OTP flow
	if h.config.DevMode {
		if cfg.TwilioAccountSID != "" && cfg.TwilioAuthToken != "" && cfg.TwilioVerifySID != "" {
			// Twilio is configured in dev mode
			h.logger.InfoContext(r.Context(), "Twilio OTP sent in development mode", "phone", phoneE164)
		} else {
			// Mock flow in dev mode - include OTP for testing
			h.logger.InfoContext(r.Context(), "Mock OTP flow used in development mode", "phone", phoneE164)
			// For mock flow, get the latest OTP from outbox
			if otp := h.getLatestOTPFromOutbox(r.Context(), phoneE164); otp != "" {
				h.logger.InfoContext(r.Context(), "Found mock OTP for development response", "phone", phoneE164, "otp", otp)
				response.DevOTP = otp
			} else {
				h.logger.WarnContext(r.Context(), "No mock OTP found for development response", "phone", phoneE164)
			}
		}
	} else {
		h.logger.InfoContext(r.Context(), "Production mode - OTP details not included in response")
	}

	RespondWithJSON(w, http.StatusOK, response, h.logger)
}

// VerifyHandler handles POST /auth/verify
// @Summary Verify OTP and get authentication token
// @Description Verifies the one-time password (OTP) and returns a JWT token on success
// @Tags auth
// @Accept json
// @Produce json
// @Param request body VerifyRequest true "Verification details"
// @Success 200 {object} VerifyResponse "Verified successfully, returns JWT token"
// @Failure 400 {object} ErrorResponse "Invalid request format"
// @Failure 401 {object} ErrorResponse "Invalid OTP or verification failed"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/verify [post]
func (h *AuthHandler) VerifyHandler(w http.ResponseWriter, r *http.Request) {
	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload", h.logger, "error", err.Error())
		return
	}
	defer r.Body.Close()

	if req.Phone == "" || req.Code == "" {
		RespondWithError(w, http.StatusBadRequest, "Phone number and code are required", h.logger)
		return
	}

	// Parse the phone number to E.164 format
	parsedNum, err := phonenumbers.Parse(req.Phone, "")
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid phone number format", h.logger, "error", err.Error())
		return
	}
	if !phonenumbers.IsValidNumber(parsedNum) {
		RespondWithError(w, http.StatusBadRequest, "Invalid phone number", h.logger)
		return
	}
	phoneE164 := phonenumbers.Format(parsedNum, phonenumbers.E164)

	token, err := h.userService.VerifyOTP(r.Context(), phoneE164, strings.TrimSpace(req.Code))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == service.ErrOTPValidationFailed || err == service.ErrUserNotFound {
			statusCode = http.StatusUnauthorized
		}
		RespondWithError(w, statusCode, "OTP verification failed", h.logger, "error", err.Error())
		return
	}

	// Get user details for audit logging
	user, err := h.querier.GetUserByPhone(r.Context(), phoneE164)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to get user for audit logging", "error", err, "phone", phoneE164)
		// Continue without audit logging rather than failing the login
	} else {
		// Log successful login event
		userName := ""
		if user.Name.Valid {
			userName = user.Name.String
		}

		ipAddress, userAgent := GetAuditInfoFromContext(r.Context())
		if err := h.auditService.LogUserLogin(r.Context(), user.UserID, userName, user.Phone, ipAddress, userAgent); err != nil {
			h.logger.ErrorContext(r.Context(), "Failed to log user login audit event", "error", err, "user_id", user.UserID)
		}
	}

	RespondWithJSON(w, http.StatusOK, VerifyResponse{Token: token}, h.logger)
}

func (h *AuthHandler) getLatestOTPFromOutbox(ctx context.Context, phoneE164 string) string {
	// For development mode, get recent outbox items for this specific recipient
	items, err := h.querier.GetRecentOutboxItemsByRecipient(ctx, db.GetRecentOutboxItemsByRecipientParams{
		Recipient: phoneE164,
		Limit:     5,
	})
	if err != nil {
		h.logger.WarnContext(ctx, "Failed to get outbox items for dev OTP", "error", err)
		return ""
	}

	h.logger.InfoContext(ctx, "Retrieved outbox items for OTP lookup", "phone", phoneE164, "count", len(items))

	// Find the latest OTP for this phone number
	for _, item := range items {
		h.logger.InfoContext(ctx, "Examining outbox item", "outbox_id", item.OutboxID, "message_type", item.MessageType, "status", item.Status)
		if item.MessageType == "OTP_VERIFICATION" && item.Payload.Valid {
			// Parse the JSON payload to extract the OTP
			var payload struct {
				OTP string `json:"otp"`
			}
			if err := json.Unmarshal([]byte(item.Payload.String), &payload); err == nil {
				h.logger.InfoContext(ctx, "Successfully extracted OTP from outbox", "phone", phoneE164, "otp", payload.OTP)
				return payload.OTP
			} else {
				h.logger.WarnContext(ctx, "Failed to parse OTP payload", "error", err, "payload", item.Payload.String)
			}
		}
	}

	return ""
}

// DevLoginHandler handles POST /auth/dev-login (development mode only)
// @Summary Development-only direct login endpoint
// @Description Bypasses OTP and directly generates JWT token for testing (DEV MODE ONLY)
// @Tags auth
// @Accept json
// @Produce json
// @Param request body DevLoginRequest true "Phone number for dev login"
// @Success 200 {object} DevLoginResponse "JWT token generated successfully"
// @Failure 400 {object} ErrorResponse "Invalid request or dev mode disabled"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/dev-login [post]
func (h *AuthHandler) DevLoginHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow in development mode
	if !h.config.DevMode {
		RespondWithError(w, http.StatusForbidden, "Development login endpoint is only available in dev mode", h.logger)
		return
	}

	var req DevLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload", h.logger, "error", err.Error())
		return
	}
	defer r.Body.Close()

	if req.Phone == "" {
		RespondWithError(w, http.StatusBadRequest, "Phone number is required", h.logger)
		return
	}

	// Parse the phone number to E.164 format
	parsedNum, err := phonenumbers.Parse(req.Phone, "")
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid phone number format", h.logger, "error", err.Error())
		return
	}
	if !phonenumbers.IsValidNumber(parsedNum) {
		RespondWithError(w, http.StatusBadRequest, "Invalid phone number", h.logger)
		return
	}
	phoneE164 := phonenumbers.Format(parsedNum, phonenumbers.E164)

	h.logger.InfoContext(r.Context(), "Dev login attempt", "phone", phoneE164)

	// Get user details from database
	user, err := h.querier.GetUserByPhone(r.Context(), phoneE164)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondWithError(w, http.StatusNotFound, "User not found", h.logger, "phone", phoneE164)
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to get user", h.logger, "error", err.Error())
		return
	}

	// Generate JWT token directly using the auth package
	userName := ""
	if user.Name.Valid {
		userName = user.Name.String
	}
	token, err := auth.GenerateJWT(user.UserID, user.Phone, userName, user.Role, h.config.JWTSecret, h.config.JWTExpirationHours)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to generate token", h.logger, "error", err.Error())
		return
	}

	// Log successful dev login event
	ipAddress, userAgent := GetAuditInfoFromContext(r.Context())
	if err := h.auditService.LogUserLogin(r.Context(), user.UserID, userName, user.Phone, ipAddress, userAgent); err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to log dev login audit event", "error", err, "user_id", user.UserID)
	}

	response := DevLoginResponse{
		Token: token,
		User: struct {
			ID    int64  `json:"id"`
			Phone string `json:"phone"`
			Name  string `json:"name"`
			Role  string `json:"role"`
		}{
			ID:    user.UserID,
			Phone: user.Phone,
			Name:  user.Name.String,
			Role:  user.Role,
		},
	}

	h.logger.InfoContext(r.Context(), "Dev login successful", "phone", phoneE164, "user_id", user.UserID, "role", user.Role)
	RespondWithJSON(w, http.StatusOK, response, h.logger)
}
