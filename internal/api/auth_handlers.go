package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"night-owls-go/internal/service"

	"github.com/nyaruka/phonenumbers"
)

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

// RegisterRequest is the expected JSON request body for /auth/register.
type RegisterRequest struct {
	Phone string `json:"phone"`
	Name  string `json:"name,omitempty"` // Optional name
}

// RegisterResponse is the JSON response for a successful /auth/register.
type RegisterResponse struct {
	Message string `json:"message"`
}

// RegisterHandler handles requests to /auth/register.
func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload", h.logger, "error", err.Error())
		return
	}
	defer r.Body.Close()

	trimmedPhone := strings.TrimSpace(req.Phone)
	if trimmedPhone == "" {
		RespondWithError(w, http.StatusBadRequest, "Phone number cannot be empty", h.logger)
		return
	}

	// Validate phone number using phonenumbers library
	// For numbers with a leading "+", the defaultRegion is usually ignored by Parse.
	// If we expect numbers without a country code, a defaultRegion (e.g., "ZA" for South Africa) would be crucial.
	num, err := phonenumbers.Parse(trimmedPhone, "") // Using empty defaultRegion, relies on "+" for international
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid phone number format.", h.logger, "phone_provided", trimmedPhone, "parse_error", err.Error())
		return
	}
	if !phonenumbers.IsValidNumber(num) {
		RespondWithError(w, http.StatusBadRequest, "Phone number is not valid.", h.logger, "phone_provided", trimmedPhone)
		return
	}

	// Optionally, format to E.164 for consistent storage, though not strictly enforced here yet
	e164Phone := phonenumbers.Format(num, phonenumbers.E164)

	var sqlName sql.NullString
	if req.Name != "" {
		sqlName.String = req.Name
		sqlName.Valid = true
	}

	// Use the E164 formatted phone number for registration
	err = h.userService.RegisterOrLoginUser(r.Context(), e164Phone, sqlName)
	if err != nil {
		if errors.Is(err, service.ErrInternalServer) { 
			RespondWithError(w, http.StatusInternalServerError, "Failed to process registration", h.logger, "phone", e164Phone)
		} else {
			RespondWithError(w, http.StatusInternalServerError, "An unexpected error occurred", h.logger, "phone", e164Phone, "service_error", err.Error())
		}
		return
	}

	RespondWithJSON(w, http.StatusOK, RegisterResponse{Message: "OTP sent. For development, check server logs or OTP log file."}, h.logger)
}

// VerifyRequest is the expected JSON request body for /auth/verify.
type VerifyRequest struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

// VerifyResponse is the JSON response for a successful /auth/verify.
type VerifyResponse struct {
	Token string `json:"token"`
}

// VerifyHandler handles requests to /auth/verify.
func (h *AuthHandler) VerifyHandler(w http.ResponseWriter, r *http.Request) {
	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload", h.logger, "error", err.Error())
		return
	}
	defer r.Body.Close()

	trimmedPhone := strings.TrimSpace(req.Phone)
	trimmedCode := strings.TrimSpace(req.Code)

	if trimmedPhone == "" || trimmedCode == "" {
		RespondWithError(w, http.StatusBadRequest, "Phone number and code cannot be empty", h.logger)
		return
	}
	
	// It might be good to parse/validate phone here too, or ensure it's stored in E164
	// For now, assume phone from request is what user initially entered for register.
	// If we decide to always use E164 internally, this might need adjustment.
	numVerify, err := phonenumbers.Parse(trimmedPhone, "")
	if err != nil || !phonenumbers.IsValidNumber(numVerify) {
		RespondWithError(w, http.StatusBadRequest, "Invalid phone number format for verification.", h.logger, "phone_provided", trimmedPhone)
		return
	}
	e164PhoneVerify := phonenumbers.Format(numVerify, phonenumbers.E164)

	token, err := h.userService.VerifyOTP(r.Context(), e164PhoneVerify, trimmedCode)
	if err != nil {
		if errors.Is(err, service.ErrOTPValidationFailed) {
			RespondWithError(w, http.StatusUnauthorized, "Invalid or expired OTP", h.logger, "phone", e164PhoneVerify)
		} else if errors.Is(err, service.ErrUserNotFound) {
			RespondWithError(w, http.StatusNotFound, "User not found for this phone number", h.logger, "phone", e164PhoneVerify)
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to verify OTP", h.logger, "phone", e164PhoneVerify, "service_error", err.Error())
		}
		return
	}

	RespondWithJSON(w, http.StatusOK, VerifyResponse{Token: token}, h.logger)
} 