package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"night-owls-go/internal/service"
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
	if errJSON := json.NewDecoder(r.Body).Decode(&req); errJSON != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload", h.logger, "error", errJSON.Error())
		return
	}
	defer r.Body.Close()

	trimmedPhone := strings.TrimSpace(req.Phone)
	if trimmedPhone == "" {
		RespondWithError(w, http.StatusBadRequest, "Phone number cannot be empty", h.logger)
		return
	}

	if !phoneRegex.MatchString(trimmedPhone) {
		RespondWithError(w, http.StatusBadRequest, "Invalid phone number format (e.g., +12223334444).", h.logger, "phone_provided", trimmedPhone)
		return
	}

	var sqlName sql.NullString
	if req.Name != "" {
		sqlName.String = req.Name
		sqlName.Valid = true
	}

	err := h.userService.RegisterOrLoginUser(r.Context(), trimmedPhone, sqlName)
	if err != nil {
		if errors.Is(err, service.ErrInternalServer) { 
			RespondWithError(w, http.StatusInternalServerError, "Failed to process registration", h.logger, "phone", trimmedPhone)
		} else {
			RespondWithError(w, http.StatusInternalServerError, "An unexpected error occurred", h.logger, "phone", trimmedPhone, "service_error", err.Error())
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
	if errJSON := json.NewDecoder(r.Body).Decode(&req); errJSON != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload", h.logger, "error", errJSON.Error())
		return
	}
	defer r.Body.Close()

	trimmedPhone := strings.TrimSpace(req.Phone)
	trimmedCode := strings.TrimSpace(req.Code)

	if trimmedPhone == "" || trimmedCode == "" {
		RespondWithError(w, http.StatusBadRequest, "Phone number and code cannot be empty", h.logger)
		return
	}
	
	if !phoneRegex.MatchString(trimmedPhone) {
		RespondWithError(w, http.StatusBadRequest, "Invalid phone number format for verification.", h.logger, "phone_provided", trimmedPhone)
		return
	}

	token, err := h.userService.VerifyOTP(r.Context(), trimmedPhone, trimmedCode)
	if err != nil {
		if errors.Is(err, service.ErrOTPValidationFailed) {
			RespondWithError(w, http.StatusUnauthorized, "Invalid or expired OTP", h.logger, "phone", trimmedPhone)
		} else if errors.Is(err, service.ErrUserNotFound) {
			RespondWithError(w, http.StatusNotFound, "User not found for this phone number", h.logger, "phone", trimmedPhone)
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to verify OTP", h.logger, "phone", trimmedPhone, "service_error", err.Error())
		}
		return
	}

	RespondWithJSON(w, http.StatusOK, VerifyResponse{Token: token}, h.logger)
} 