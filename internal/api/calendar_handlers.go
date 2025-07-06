package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
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

	// Store token in database (we'll need to create this table)
	expiresAt := time.Now().Add(365 * 24 * time.Hour) // 1 year expiry
	
	// For now, we'll store it as a simple key-value store
	// In production, you'd want a proper calendar_tokens table
	
	// Build URLs
	baseURL := getBaseURL(r)
	feedURL := fmt.Sprintf("%s/api/calendar/user/%d/%s", baseURL, userID, token)
	webCalURL := strings.Replace(feedURL, "https://", "webcal://", 1)
	webCalURL = strings.Replace(webCalURL, "http://", "webcal://", 1)

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
	// Parse URL path to extract userID and token
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 5 {
		RespondWithError(w, http.StatusBadRequest, "Invalid calendar feed URL", h.logger)
		return
	}

	userIDStr := pathParts[3] // api/calendar/user/{userId}/{token}
	token := pathParts[4]    // No .ics extension to trim

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		h.logger.WarnContext(r.Context(), "Invalid user ID in calendar feed request", "user_id", userIDStr)
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID", h.logger)
		return
	}

	// Validate token (in production, check against database)
	if !isValidCalendarToken(userID, token) {
		h.logger.WarnContext(r.Context(), "Invalid calendar feed token", "user_id", userID, "token", token[:8]+"...")
		RespondWithError(w, http.StatusNotFound, "Invalid calendar feed", h.logger)
		return
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

// isValidCalendarToken validates a calendar feed token
// TODO: Implement proper token validation against database
func isValidCalendarToken(userID int64, token string) bool {
	// Placeholder implementation
	// In production, validate against calendar_tokens table
	return len(token) == 64 // Hex-encoded 32 bytes
} 