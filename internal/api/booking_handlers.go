package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"night-owls-go/internal/service"

	"github.com/go-chi/chi/v5"
)

// ErrorResponse represents an error response in the API
// Used for Swagger documentation
type ErrorResponse struct {
	Error string `json:"error"`
}

// BookingHandler handles user-facing booking operations.
type BookingHandler struct {
	bookingService *service.BookingService
	logger         *slog.Logger
}

// NewBookingHandler creates a new BookingHandler.
func NewBookingHandler(bookingService *service.BookingService, logger *slog.Logger) *BookingHandler {
	return &BookingHandler{
		bookingService: bookingService,
		logger:         logger.With("handler", "BookingHandler"),
	}
}

// CreateBookingRequest is the expected JSON for POST /bookings.
type CreateBookingRequest struct {
	ScheduleID int64     `json:"schedule_id"`
	StartTime  time.Time `json:"start_time"`
	BuddyName  *string   `json:"buddy_name,omitempty"`
	BuddyPhone *string   `json:"buddy_phone,omitempty"`
}

// CreateBookingHandler handles POST /bookings
// @Summary Create a new booking
// @Description Books a shift slot for a user
// @Tags bookings
// @Accept json
// @Produce json
// @Param request body CreateBookingRequest true "Booking details"
// @Success 201 {object} BookingResponse "Booking created successfully"
// @Failure 400 {object} ErrorResponse "Invalid request format or data"
// @Failure 401 {object} ErrorResponse "Unauthorized - authentication required"
// @Failure 404 {object} ErrorResponse "Schedule not found"
// @Failure 409 {object} ErrorResponse "Slot already booked (conflict)"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /bookings [post]
func (h *BookingHandler) CreateBookingHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from JWT context
	userID, ok := r.Context().Value(UserIDKey).(int64)
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User not authenticated", h.logger)
		return
	}

	var req CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload", h.logger, "error", err.Error())
		return
	}
	defer r.Body.Close()

	if req.ScheduleID <= 0 || req.StartTime.IsZero() {
		RespondWithError(w, http.StatusBadRequest, "Missing or invalid schedule_id or start_time", h.logger)
		return
	}

	// Ensure StartTime is in UTC
	utcStartTime := req.StartTime.UTC()

	// Convert pointer strings to sql.NullString
	var buddyName sql.NullString
	if req.BuddyName != nil && *req.BuddyName != "" {
		buddyName = sql.NullString{String: *req.BuddyName, Valid: true}
	}

	var buddyPhone sql.NullString
	if req.BuddyPhone != nil && *req.BuddyPhone != "" {
		buddyPhone = sql.NullString{String: *req.BuddyPhone, Valid: true}
	}

	createdBooking, err := h.bookingService.CreateBooking(r.Context(), userID, req.ScheduleID, utcStartTime, buddyPhone, buddyName)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrScheduleNotFound):
			RespondWithError(w, http.StatusNotFound, "Schedule not found", h.logger, "schedule_id", req.ScheduleID, "error", err.Error())
		case errors.Is(err, service.ErrShiftTimeInvalid):
			RespondWithError(w, http.StatusBadRequest, "Invalid shift start time for the schedule", h.logger, "schedule_id", req.ScheduleID, "start_time", req.StartTime, "error", err.Error())
		case errors.Is(err, service.ErrBookingConflict):
			RespondWithError(w, http.StatusConflict, "Shift slot is already booked", h.logger, "schedule_id", req.ScheduleID, "start_time", req.StartTime, "error", err.Error())
		case errors.Is(err, service.ErrInternalServer):
			RespondWithError(w, http.StatusInternalServerError, "Internal server error processing booking", h.logger, "error", err.Error())
		default:
			RespondWithError(w, http.StatusInternalServerError, "An unexpected error occurred", h.logger, "error", err.Error())
		}
		return
	}

	bookingResponse := ToBookingResponse(createdBooking)
	RespondWithJSON(w, http.StatusCreated, bookingResponse, h.logger)
}

// GetMyBookingsHandler handles GET /bookings/my
// @Summary Get current user's bookings
// @Description Returns all bookings for the authenticated user
// @Tags bookings
// @Produce json
// @Success 200 {array} BookingWithScheduleResponse "List of user's bookings with schedule names"
// @Failure 401 {object} ErrorResponse "Unauthorized - authentication required"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /bookings/my [get]
func (h *BookingHandler) GetMyBookingsHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from JWT context
	userID, ok := r.Context().Value(UserIDKey).(int64)
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User not authenticated", h.logger)
		return
	}

	bookings, err := h.bookingService.GetUserBookings(r.Context(), userID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch user bookings", h.logger, "error", err.Error())
		return
	}

	// Convert to API response format
	bookingResponses := make([]BookingWithScheduleResponse, len(bookings))
	for i, booking := range bookings {
		bookingResponses[i] = ToBookingWithScheduleResponse(booking)
	}

	RespondWithJSON(w, http.StatusOK, bookingResponses, h.logger)
}

// MarkCheckInHandler handles POST /bookings/{id}/checkin
// @Summary Check in to a booking
// @Description Mark the user as checked in to their booked shift with a timestamp
// @Tags bookings
// @Produce json
// @Param id path int true "Booking ID"
// @Success 200 {object} BookingResponse "Check-in marked successfully"
// @Failure 400 {object} ErrorResponse "Invalid booking ID"
// @Failure 401 {object} ErrorResponse "Unauthorized - authentication required"
// @Failure 403 {object} ErrorResponse "Not authorized to check in to this booking"
// @Failure 404 {object} ErrorResponse "Booking not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /bookings/{id}/checkin [post]
func (h *BookingHandler) MarkCheckInHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from JWT context
	userID, ok := r.Context().Value(UserIDKey).(int64)
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User not authenticated", h.logger)
		return
	}

	// Extract booking ID from URL
	bookingIDStr := chi.URLParam(r, "id")
	h.logger.InfoContext(r.Context(), "MarkCheckInHandler called", "id_param", bookingIDStr, "url", r.URL.Path)
	
	bookingID, err := strconv.ParseInt(bookingIDStr, 10, 64)
	if err != nil || bookingID <= 0 {
		h.logger.ErrorContext(r.Context(), "Failed to parse booking ID", "id_param", bookingIDStr, "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid booking ID", h.logger, "booking_id", bookingIDStr)
		return
	}

	updatedBooking, err := h.bookingService.MarkCheckIn(r.Context(), bookingID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBookingNotFound):
			RespondWithError(w, http.StatusNotFound, "Booking not found", h.logger, "booking_id", bookingID, "error", err.Error())
		case errors.Is(err, service.ErrForbiddenUpdate):
			RespondWithError(w, http.StatusForbidden, "Not authorized to check in to this booking", h.logger, "booking_id", bookingID, "user_id", userID, "error", err.Error())
		case errors.Is(err, service.ErrCheckInTooEarly):
			RespondWithError(w, http.StatusBadRequest, "Check-in is too early - can only check in up to 30 minutes before shift starts", h.logger, "booking_id", bookingID, "error", err.Error())
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to mark check-in", h.logger, "error", err.Error())
		}
		return
	}

	bookingResponse := ToBookingResponse(updatedBooking)
	RespondWithJSON(w, http.StatusOK, bookingResponse, h.logger)
}

// CancelBookingHandler handles DELETE /bookings/{id}
// @Summary Cancel a booking
// @Description Cancel a user's booking if it's not too close to the shift start time
// @Tags bookings
// @Produce json
// @Param id path int true "Booking ID"
// @Success 204 "Booking cancelled successfully"
// @Failure 400 {object} ErrorResponse "Invalid booking ID or booking cannot be cancelled"
// @Failure 401 {object} ErrorResponse "Unauthorized - authentication required"
// @Failure 403 {object} ErrorResponse "Not authorized to cancel this booking"
// @Failure 404 {object} ErrorResponse "Booking not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /bookings/{id} [delete]
func (h *BookingHandler) CancelBookingHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from JWT context
	userID, ok := r.Context().Value(UserIDKey).(int64)
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User not authenticated", h.logger)
		return
	}

	// Extract booking ID from URL
	bookingIDStr := chi.URLParam(r, "id")
	h.logger.InfoContext(r.Context(), "CancelBookingHandler called", "id_param", bookingIDStr, "url", r.URL.Path)
	
	bookingID, err := strconv.ParseInt(bookingIDStr, 10, 64)
	if err != nil || bookingID <= 0 {
		h.logger.ErrorContext(r.Context(), "Failed to parse booking ID", "id_param", bookingIDStr, "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid booking ID", h.logger, "booking_id", bookingIDStr)
		return
	}

	err = h.bookingService.CancelBooking(r.Context(), bookingID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBookingNotFound):
			RespondWithError(w, http.StatusNotFound, "Booking not found", h.logger, "booking_id", bookingID, "error", err.Error())
		case errors.Is(err, service.ErrForbiddenUpdate):
			RespondWithError(w, http.StatusForbidden, "Not authorized to cancel this booking", h.logger, "booking_id", bookingID, "user_id", userID, "error", err.Error())
		case errors.Is(err, service.ErrBookingCannotBeCancelled):
			RespondWithError(w, http.StatusBadRequest, "Booking cannot be cancelled - shift has already started or is too close to start time", h.logger, "booking_id", bookingID, "error", err.Error())
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to cancel booking", h.logger, "error", err.Error())
		}
		return
	}

	// Return 204 No Content for successful deletion
	w.WriteHeader(http.StatusNoContent)
} 