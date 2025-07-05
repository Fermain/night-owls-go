package api

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"log/slog"
	"math/big"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"night-owls-go/internal/auth"
	"night-owls-go/internal/config"
	db "night-owls-go/internal/db/sqlc_generated"
	"night-owls-go/internal/service"

	"github.com/gorilla/sessions"
	"github.com/nyaruka/phonenumbers"
)

var phoneRegex = regexp.MustCompile(`^\+[1-9]\d{6,14}$`)

// Session configuration constants
const (
	SessionName = "night-owls-session"
	// SessionMaxAge removed - now calculated from JWT expiration config
)

// Standardized error messages to prevent user enumeration
const (
	// Generic authentication error - used for all auth failures
	AuthenticationFailedMessage = "Authentication failed"
	
	// Generic validation error - used for all validation failures  
	ValidationFailedMessage = "Invalid request"
	
	// Generic internal error - used for all server errors
	InternalErrorMessage = "Service temporarily unavailable"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	userService  *service.UserService
	auditService *service.AuditService
	logger       *slog.Logger
	config       *config.Config
	querier      db.Querier
	sessionStore sessions.Store
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(userService *service.UserService, auditService *service.AuditService, logger *slog.Logger, cfg *config.Config, querier db.Querier, sessionStore sessions.Store) *AuthHandler {
	logger.Info("AuthHandler created with config", "dev_mode", cfg.DevMode, "server_port", cfg.ServerPort)
	return &AuthHandler{
		userService:  userService,
		auditService: auditService,
		logger:       logger.With("handler", "AuthHandler"),
		config:       cfg,
		querier:      querier,
		sessionStore: sessionStore,
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

// addTimingRandomization adds a small random delay to prevent timing attacks
func addTimingRandomization() {
	// Add 50-150ms random delay to normalize response times
	randomMs, err := rand.Int(rand.Reader, big.NewInt(100))
	if err != nil {
		// Fallback to fixed delay if randomization fails
		time.Sleep(100 * time.Millisecond)
		return
	}
	delay := time.Duration(50+randomMs.Int64()) * time.Millisecond
	time.Sleep(delay)
}

// setUserSession creates a secure session for the authenticated user
func (h *AuthHandler) setUserSession(w http.ResponseWriter, r *http.Request, user db.GetUserByPhoneRow, token string) error {
	session, err := h.sessionStore.Get(r, SessionName)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to get session", "error", err)
		return err
	}

	// Set session values
	session.Values["user_id"] = user.UserID
	session.Values["phone"] = user.Phone
	session.Values["role"] = user.Role
	session.Values["token"] = token // Keep token for backward compatibility
	
	userName := ""
	if user.Name.Valid {
		userName = user.Name.String
	}
	session.Values["name"] = userName

	// Configure session options
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   h.config.JWTExpirationHours * 3600, // Convert hours to seconds, sync with JWT expiry
		HttpOnly: true,
		Secure:   !h.config.DevMode, // Secure in production, allow HTTP in dev
		SameSite: http.SameSiteStrictMode,
	}

	// Save session
	return session.Save(r, w)
}

// clearUserSession removes the user session (logout)
func (h *AuthHandler) clearUserSession(w http.ResponseWriter, r *http.Request) error {
	session, err := h.sessionStore.Get(r, SessionName)
	if err != nil {
		h.logger.WarnContext(r.Context(), "Failed to get session for clearing", "error", err)
		// Continue with clearing anyway
	}

	// Clear session values
	session.Values = make(map[interface{}]interface{})
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   -1, // Expire immediately
		HttpOnly: true,
		Secure:   !h.config.DevMode,
		SameSite: http.SameSiteStrictMode,
	}

	return session.Save(r, w)
}

// isDevelopmentMode checks if we're in development mode
func isDevelopmentMode() bool {
	env := strings.ToLower(os.Getenv("ENVIRONMENT"))
	devMode := os.Getenv("DEV_MODE") == "true"
	return env == "development" || env == "dev" || devMode
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

	// Extract client information for rate limiting
	clientIP, userAgent := extractClientInfo(r)

	err = h.userService.RegisterOrLoginUser(r.Context(), phoneE164, sqlName, clientIP, userAgent)
	if err != nil {
		h.logger.InfoContext(r.Context(), "RegisterOrLoginUser error", "error_message", err.Error())
		
		// Add timing randomization to prevent enumeration via timing attacks
		addTimingRandomization()

		// Differentiate between client-side and server-side errors
		if strings.Contains(err.Error(), "rate limit") {
			// Rate limiting errors should be more specific to help legitimate users
			RespondWithError(w, http.StatusTooManyRequests, "Too many requests. Please try again later.", h.logger, "error", err.Error())
		} else if strings.Contains(err.Error(), "internal server error") || strings.Contains(err.Error(), "failed to send SMS") || strings.Contains(err.Error(), "database") || strings.Contains(err.Error(), "Twilio") {
			// Server-side errors (e.g., database, SMS/Twilio issues) should return 500
			RespondWithError(w, http.StatusInternalServerError, InternalErrorMessage, h.logger, "error", err.Error())
		} else {
			// Client-side errors (user not found, validation failures, etc.) get generic message
			RespondWithError(w, http.StatusBadRequest, ValidationFailedMessage, h.logger, "error", err.Error())
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

	// Extract client information for rate limiting and audit logging
	clientIP, userAgent := extractClientInfo(r)

	token, err := h.userService.VerifyOTP(r.Context(), phoneE164, strings.TrimSpace(req.Code), clientIP, userAgent)
	if err != nil {
		// Add timing randomization to prevent enumeration via timing attacks
		addTimingRandomization()
		
		// Differentiate between client-side and server-side errors
		if strings.Contains(err.Error(), "rate limit") {
			// Rate limiting errors should be more specific to help legitimate users
			RespondWithError(w, http.StatusTooManyRequests, "Too many requests. Please try again later.", h.logger, "error", err.Error())
		} else if strings.Contains(err.Error(), "internal server error") || strings.Contains(err.Error(), "database") || strings.Contains(err.Error(), "Twilio") {
			// Server-side errors (e.g., database or Twilio issues) should return 500
			RespondWithError(w, http.StatusInternalServerError, InternalErrorMessage, h.logger, "error", err.Error())
		} else {
			// Client-side errors (invalid OTP, user not found, etc.) get generic message
			RespondWithError(w, http.StatusUnauthorized, AuthenticationFailedMessage, h.logger, "error", err.Error())
		}
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

	// Set secure HTTP-only cookie for the JWT token
	err = h.setUserSession(w, r, user, token)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to set session", h.logger, "error", err.Error())
		return
	}
	
	// For now, also return token in response for backward compatibility
	// TODO: Remove token from JSON response in future version once all clients use cookies
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
		// Add timing randomization to prevent enumeration via timing attacks
		addTimingRandomization()
		
		// Always return the same generic error regardless of the specific issue
		// This prevents attackers from determining if a phone number is registered
		if err == sql.ErrNoRows {
			RespondWithError(w, http.StatusUnauthorized, AuthenticationFailedMessage, h.logger, "phone", phoneE164)
		} else {
			RespondWithError(w, http.StatusInternalServerError, InternalErrorMessage, h.logger, "error", err.Error())
		}
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

	// Set secure HTTP-only cookie for the JWT token
	err = h.setUserSession(w, r, user, token)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to set session", h.logger, "error", err.Error())
		return
	}

	response := DevLoginResponse{
		Token: token, // Keep for backward compatibility
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

// ValidateHandler handles GET /auth/validate
// @Summary Validate JWT token and return user info
// @Description Validates the JWT token and returns user information for server-side route protection
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "User info if token is valid"
// @Failure 401 {object} ErrorResponse "Invalid or expired token"
// @Router /auth/validate [get]
func (h *AuthHandler) ValidateHandler(w http.ResponseWriter, r *http.Request) {
	// Extract token from Authorization header or session
	var token string
	
	// First try session (from cookie)
	if session, err := h.sessionStore.Get(r, SessionName); err == nil {
		if sessionToken, ok := session.Values["token"].(string); ok {
			token = sessionToken
		}
	}
	
	// Fall back to Authorization header
	if token == "" {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}
	
	if token == "" {
		RespondWithError(w, http.StatusUnauthorized, "No token provided", h.logger)
		return
	}
	
	// Validate JWT token
	claims, err := auth.ValidateJWT(token, h.config.JWTSecret)
	if err != nil {
		h.logger.WarnContext(r.Context(), "Invalid JWT token in validate endpoint", "error", err.Error())
		RespondWithError(w, http.StatusUnauthorized, "Invalid token", h.logger)
		return
	}
	
	// Return user information
	userInfo := map[string]interface{}{
		"id":    claims.UserID,
		"phone": claims.Phone,
		"name":  claims.Name,
		"role":  claims.Role,
	}
	
	RespondWithJSON(w, http.StatusOK, userInfo, h.logger)
}

// LogoutHandler handles POST /auth/logout
// @Summary Logout and clear authentication cookies
// @Description Clears the JWT authentication cookie and logs out the user
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Successfully logged out"
// @Router /auth/logout [post]
func (h *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear the JWT cookie
	err := h.clearUserSession(w, r)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to clear session", h.logger, "error", err.Error())
		return
	}
	
	// Log logout event if user context is available
	if userIDVal := r.Context().Value(UserIDKey); userIDVal != nil {
		if userID, ok := userIDVal.(int64); ok {
			ipAddress, userAgent := GetAuditInfoFromContext(r.Context())
			if err := h.auditService.LogUserLogout(r.Context(), userID, ipAddress, userAgent); err != nil {
				h.logger.ErrorContext(r.Context(), "Failed to log user logout audit event", "error", err, "user_id", userID)
			}
		}
	}
	
	h.logger.InfoContext(r.Context(), "User logged out successfully")
	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Successfully logged out"}, h.logger)
}

// extractClientInfo extracts client IP and user agent from HTTP request
func extractClientInfo(r *http.Request) (clientIP, userAgent string) {
	// Extract IP address (handle proxy headers)
	clientIP = r.Header.Get("X-Forwarded-For")
	if clientIP != "" {
		// Extract the first IP from the comma-separated list
		clientIP = strings.TrimSpace(strings.Split(clientIP, ",")[0])
	} else {
		clientIP = r.Header.Get("X-Real-IP")
	}
	if clientIP == "" {
		clientIP = r.RemoteAddr
	}
	// Clean up IP (remove port if present) - IPv6 safe
	host, _, err := net.SplitHostPort(clientIP)
	if err == nil {
		clientIP = host
	} else {
		// Fallback: sanitize clientIP by removing port manually if possible
		if strings.Contains(clientIP, ":") {
			clientIP = strings.Split(clientIP, ":")[0]
		}
		// If still problematic, use a safe default
		if clientIP == "" || strings.Contains(clientIP, " ") {
			clientIP = "unknown"
		}
	}
	
	// Extract user agent
	userAgent = r.UserAgent()
	if userAgent == "" {
		userAgent = "unknown"
	}
	
	return clientIP, userAgent
}
