package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"night-owls-go/internal/service"
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

	if strings.TrimSpace(req.Phone) == "" {
		RespondWithError(w, http.StatusBadRequest, "Phone number cannot be empty", h.logger)
		return
	}
	// TODO: Add more robust phone number validation (e.g., regex, library)

	var sqlName sql.NullString
	if req.Name != "" {
		sqlName.String = req.Name
		sqlName.Valid = true
	}

	err := h.userService.RegisterOrLoginUser(r.Context(), req.Phone, sqlName)
	if err != nil {
		if errors.Is(err, service.ErrInternalServer) { // Example of mapping service errors
			RespondWithError(w, http.StatusInternalServerError, "Failed to process registration", h.logger, "phone", req.Phone)
		} else {
			RespondWithError(w, http.StatusInternalServerError, "An unexpected error occurred", h.logger, "phone", req.Phone, "service_error", err.Error())
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

	if strings.TrimSpace(req.Phone) == "" || strings.TrimSpace(req.Code) == "" {
		RespondWithError(w, http.StatusBadRequest, "Phone number and code cannot be empty", h.logger)
		return
	}

	token, err := h.userService.VerifyOTP(r.Context(), req.Phone, req.Code)
	if err != nil {
		if errors.Is(err, service.ErrOTPValidationFailed) {
			RespondWithError(w, http.StatusUnauthorized, "Invalid or expired OTP", h.logger, "phone", req.Phone)
		} else if errors.Is(err, service.ErrUserNotFound) {
			RespondWithError(w, http.StatusNotFound, "User not found for this phone number", h.logger, "phone", req.Phone)
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to verify OTP", h.logger, "phone", req.Phone, "service_error", err.Error())
		}
		return
	}

	RespondWithJSON(w, http.StatusOK, VerifyResponse{Token: token}, h.logger)
} 