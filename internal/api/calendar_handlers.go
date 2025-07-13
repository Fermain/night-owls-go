package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"night-owls-go/internal/service"
	"night-owls-go/internal/utils"

	db "night-owls-go/internal/db/sqlc_generated"
)

// CalendarHandler handles calendar-related operations
type CalendarHandler struct {
	bookingService *service.BookingService
	querier        db.Querier
	logger         *slog.Logger
}

// NewCalendarHandler creates a new CalendarHandler
func NewCalendarHandler(bookingService *service.BookingService, querier db.Querier, logger *slog.Logger) *CalendarHandler {
	return &CalendarHandler{
		bookingService: bookingService,
		querier:        querier,
		logger:         logger.With("handler", "CalendarHandler"),
	}
}

// GenerateCalendarFeedToken generates a secure token for WebCal access
// @Summary Generate calendar feed token
// @Description Generates a secure token for accessing user's calendar feed
// @Tags calendar
// @Produce json
// @Success 200 {object} CalendarFeedResponse "Calendar feed information with token"
// @Failure 401 {object} ErrorResponse "Unauthorized - authentication required"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/calendar/generate-token [post]
func (h *CalendarHandler) GenerateCalendarFeedToken(w http.ResponseWriter, r *http.Request) {
	// Get the user ID from context (set by auth middleware)
	userID, ok := r.Context().Value(UserIDKey).(int64)
	if !ok {
		h.logger.ErrorContext(r.Context(), "User ID not found in context")
		RespondWithError(w, http.StatusUnauthorized, "User authentication required", h.logger)
		return
	}

	// Generate secure token
	token, err := generateSecureToken()
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to generate secure token", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to generate calendar feed token", h.logger)
		return
	}

	// Store token in database with proper hashing
	expiresAt := time.Now().Add(365 * 24 * time.Hour) // 1 year expiry

	// Hash the token before storing (never store plain tokens)
	tokenHash := hashToken(token)

	// Store token in database
	_, err = h.querier.CreateCalendarToken(r.Context(), db.CreateCalendarTokenParams{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to store calendar token", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to create calendar feed token", h.logger)
		return
	}

	// Build URLs
	baseURL := getBaseURL(r)
	feedURL := fmt.Sprintf("%s/api/calendar/user/%d/%s", baseURL, userID, token)
	webCalURL := feedURL // Use https:// directly instead of webcal:// for better compatibility

	response := CalendarFeedResponse{
		FeedURL:     feedURL,
		WebCalURL:   webCalURL,
		Token:       token,
		ExpiresAt:   expiresAt,
		Description: "Subscribe to this calendar to automatically sync your Night Owls shifts",
	}

	h.logger.InfoContext(r.Context(), "Generated calendar feed token", "user_id", userID, "expires_at", expiresAt)
	RespondWithJSON(w, http.StatusOK, response, h.logger)
}

// ServeCalendarFeed serves the WebCal ICS feed for a user
// @Summary Serve user calendar feed
// @Description Serves a dynamic ICS calendar feed with user's upcoming shifts
// @Tags calendar
// @Produce text/calendar
// @Param userId path int true "User ID"
// @Param token path string true "Calendar feed token"
// @Success 200 {string} string "ICS calendar feed content"
// @Failure 400 {object} ErrorResponse "Invalid user ID or token"
// @Failure 404 {object} ErrorResponse "User not found or invalid token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/calendar/user/{userId}/{token} [get]
func (h *CalendarHandler) ServeCalendarFeed(w http.ResponseWriter, r *http.Request) {
	// Extract userID and token using Go 1.22's PathValue
	userIDStr := r.PathValue("userId")
	token := r.PathValue("token")

	if userIDStr == "" || token == "" {
		h.logger.WarnContext(r.Context(), "Missing path parameters in calendar feed request", "user_id", userIDStr, "token_present", token != "")
		RespondWithError(w, http.StatusBadRequest, "Invalid calendar feed URL", h.logger)
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		h.logger.WarnContext(r.Context(), "Invalid user ID in calendar feed request", "user_id", userIDStr)
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID", h.logger)
		return
	}

	// Validate token against database
	validToken, err := h.validateCalendarToken(r.Context(), userID, token)
	if err != nil || !validToken {
		h.logger.WarnContext(r.Context(), "Invalid calendar feed token",
			"user_id", userID,
			"token_prefix", token[:min(8, len(token))]+"...",
			"error", err)
		RespondWithError(w, http.StatusNotFound, "Invalid calendar feed", h.logger)
		return
	}

	// Update access tracking
	tokenHash := hashToken(token)
	if updateErr := h.querier.UpdateTokenAccess(r.Context(), tokenHash); updateErr != nil {
		h.logger.WarnContext(r.Context(), "Failed to update token access", "error", updateErr)
		// Don't fail the request for tracking errors
	}

	// Get user's bookings
	bookings, err := h.querier.ListBookingsByUserIDWithSchedule(r.Context(), userID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to fetch user bookings for calendar feed", "user_id", userID, "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to generate calendar feed", h.logger)
		return
	}

	// Generate calendar feed
	calendarData := utils.GenerateUserCalendarFeed(bookings, userID)

	// Set proper headers for calendar feed
	w.Header().Set("Content-Type", calendarData.MIME)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", calendarData.Filename))
	w.Header().Set("Cache-Control", "no-cache, must-revalidate")
	w.Header().Set("Expires", "0")

	h.logger.InfoContext(r.Context(), "Served calendar feed", "user_id", userID, "booking_count", len(bookings))

	// Write calendar content
	w.WriteHeader(http.StatusOK)
	_, writeErr := w.Write([]byte(calendarData.Content))
	if writeErr != nil {
		h.logger.ErrorContext(r.Context(), "Failed to write calendar feed response", "error", writeErr)
	}
}

// RevokeCalendarToken revokes a user's calendar feed token
// @Summary Revoke calendar feed token
// @Description Revokes the current user's calendar feed token, requiring regeneration for future access
// @Tags calendar
// @Produce json
// @Success 200 {object} map[string]string "Token revoked successfully"
// @Failure 401 {object} ErrorResponse "Unauthorized - authentication required"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/calendar/revoke-token [post]
func (h *CalendarHandler) RevokeCalendarToken(w http.ResponseWriter, r *http.Request) {
	// Get the user ID from context (set by auth middleware)
	userID, ok := r.Context().Value(UserIDKey).(int64)
	if !ok {
		h.logger.ErrorContext(r.Context(), "User ID not found in context")
		RespondWithError(w, http.StatusUnauthorized, "User authentication required", h.logger)
		return
	}

	// Revoke all tokens for the user
	err := h.querier.RevokeAllUserCalendarTokens(r.Context(), userID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to revoke calendar tokens", "user_id", userID, "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to revoke calendar tokens", h.logger)
		return
	}

	h.logger.InfoContext(r.Context(), "Revoked all calendar tokens", "user_id", userID)

	response := map[string]string{
		"message": "All calendar feed tokens have been revoked successfully",
		"status":  "revoked",
	}

	RespondWithJSON(w, http.StatusOK, response, h.logger)
}

// GetCalendarTokenInfo returns information about the user's calendar tokens
// @Summary Get calendar token information
// @Description Returns information about the user's current calendar tokens
// @Tags calendar
// @Produce json
// @Success 200 {object} map[string]interface{} "Token information"
// @Failure 401 {object} ErrorResponse "Unauthorized - authentication required"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/calendar/token-info [get]
func (h *CalendarHandler) GetCalendarTokenInfo(w http.ResponseWriter, r *http.Request) {
	// Get the user ID from context (set by auth middleware)
	userID, ok := r.Context().Value(UserIDKey).(int64)
	if !ok {
		h.logger.ErrorContext(r.Context(), "User ID not found in context")
		RespondWithError(w, http.StatusUnauthorized, "User authentication required", h.logger)
		return
	}

	// Get user's calendar tokens
	tokens, err := h.querier.GetUserCalendarTokens(r.Context(), userID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to fetch calendar tokens", "user_id", userID, "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch token information", h.logger)
		return
	}

	// Count active tokens
	activeCount := 0
	for _, token := range tokens {
		if !token.IsRevoked.Bool && token.ExpiresAt.After(time.Now()) {
			activeCount++
		}
	}

	response := map[string]interface{}{
		"total_tokens":  len(tokens),
		"active_tokens": activeCount,
		"has_active":    activeCount > 0,
		"tokens":        tokens, // Include full token info for admin purposes
	}

	RespondWithJSON(w, http.StatusOK, response, h.logger)
}

// Helper functions

// generateSecureToken creates a cryptographically secure random token
func generateSecureToken() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// getBaseURL extracts the base URL from the request
func getBaseURL(r *http.Request) string {
	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}

	// Check for forwarded headers in case of reverse proxy
	if forwardedProto := r.Header.Get("X-Forwarded-Proto"); forwardedProto != "" {
		scheme = forwardedProto
	}

	host := r.Host
	if forwardedHost := r.Header.Get("X-Forwarded-Host"); forwardedHost != "" {
		host = forwardedHost
	}

	return fmt.Sprintf("%s://%s", scheme, host)
}

// validateCalendarToken validates a calendar feed token against the database
func (h *CalendarHandler) validateCalendarToken(ctx context.Context, userID int64, token string) (bool, error) {
	// Basic validation: ensure token is properly formatted
	if len(token) != 64 {
		return false, fmt.Errorf("invalid token format")
	}

	// Hash the token to match against stored hash
	tokenHash := hashToken(token)

	// Validate against database
	validationResult, err := h.querier.ValidateCalendarToken(ctx, db.ValidateCalendarTokenParams{
		UserID:    userID,
		TokenHash: tokenHash,
	})
	if err != nil {
		return false, err
	}

	// Additional security check: ensure token belongs to the correct user
	if validationResult.UserID != userID {
		return false, fmt.Errorf("token user mismatch")
	}

	return true, nil
}

// hashToken creates a SHA-256 hash of the token for secure storage
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
