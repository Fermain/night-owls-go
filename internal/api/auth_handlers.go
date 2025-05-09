package api

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"night-owls-go/internal/service"

	"github.com/nyaruka/phonenumbers"
)

var phoneRegex = regexp.MustCompile(`^\+[1-9]\d{6,14}$`)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	userService *service.UserService
	logger      *slog.Logger
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(userService *service.UserService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		logger:      logger.With("handler", "AuthHandler"),
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
		RespondWithError(w, http.StatusInternalServerError, "Failed to register/login user", h.logger, "error", err.Error())
		return
	}

	RespondWithJSON(w, http.StatusOK, RegisterResponse{Message: "OTP sent to sms_outbox.log"}, h.logger)
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

	RespondWithJSON(w, http.StatusOK, VerifyResponse{Token: token}, h.logger)
} 