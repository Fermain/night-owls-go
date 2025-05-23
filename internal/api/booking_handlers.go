package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"night-owls-go/internal/service"

	"github.com/go-chi/chi/v5"
)

// ErrorResponse represents an error response in the API
// Used for Swagger documentation
type ErrorResponse struct {
	Error string `json:"error"`
}

// BookingHandler handles booking-related HTTP requests.
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
	// Schedule ID for the booking
	ScheduleID int64 `json:"schedule_id" example:"42" validate:"required"`
	// Start time for the shift (RFC3339 format)
	StartTime time.Time `json:"start_time" example:"2025-05-10T18:00:00Z" validate:"required"`
	// Optional buddy's phone number
	BuddyPhone string `json:"buddy_phone,omitempty" example:"+1234567890"`
	// Optional buddy's name
	BuddyName string `json:"buddy_name,omitempty" example:"Bob"`
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
	userIDVal := r.Context().Value(UserIDKey)
	userID, ok := userIDVal.(int64)
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User ID not found in context or invalid type", h.logger)
		return
	}

	var req CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload", h.logger, "error", err.Error())
		return
	}
	defer r.Body.Close()

	if req.ScheduleID <= 0 {
		RespondWithError(w, http.StatusBadRequest, "Invalid schedule_id", h.logger, "schedule_id", req.ScheduleID)
		return
	}
	if req.StartTime.IsZero() {
		RespondWithError(w, http.StatusBadRequest, "Invalid start_time", h.logger, "start_time", req.StartTime)
		return
	}

	var buddyPhone sql.NullString
	if strings.TrimSpace(req.BuddyPhone) != "" {
		buddyPhone.String = strings.TrimSpace(req.BuddyPhone)
		buddyPhone.Valid = true
	}
	var buddyName sql.NullString
	if strings.TrimSpace(req.BuddyName) != "" {
		buddyName.String = strings.TrimSpace(req.BuddyName)
		buddyName.Valid = true
	}

	booking, err := h.bookingService.CreateBooking(r.Context(), userID, req.ScheduleID, req.StartTime, buddyPhone, buddyName)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrScheduleNotFound):
			RespondWithError(w, http.StatusNotFound, "Schedule not found", h.logger, "schedule_id", req.ScheduleID)
		case errors.Is(err, service.ErrShiftTimeInvalid):
			RespondWithError(w, http.StatusBadRequest, "Requested shift time is invalid for the schedule", h.logger, "schedule_id", req.ScheduleID, "start_time", req.StartTime)
		case errors.Is(err, service.ErrBookingConflict):
			RespondWithError(w, http.StatusConflict, "Shift slot is already booked", h.logger, "schedule_id", req.ScheduleID, "start_time", req.StartTime)
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to create booking", h.logger, "error", err.Error())
		}
		return
	}

	// Convert to API response format
	bookingResponse := ToBookingResponse(booking)
	RespondWithJSON(w, http.StatusCreated, bookingResponse, h.logger)
}

// MarkAttendanceRequest is the expected JSON for PATCH /bookings/{id}/attendance
type MarkAttendanceRequest struct {
	// Whether the volunteer attended
	Attended bool `json:"attended" example:"true" validate:"required"`
}

// MarkAttendanceHandler handles PATCH /bookings/{id}/attendance
// @Summary Mark attendance for a booking
// @Description Updates a booking to record whether the volunteer attended
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path int true "Booking ID" example(101)
// @Param request body MarkAttendanceRequest true "Attendance status"
// @Success 200 {object} BookingResponse "Attendance marked successfully"
// @Failure 400 {object} ErrorResponse "Invalid request format"
// @Failure 401 {object} ErrorResponse "Unauthorized - authentication required"
// @Failure 403 {object} ErrorResponse "Forbidden - not authorized to mark this booking"
// @Failure 404 {object} ErrorResponse "Booking not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /bookings/{id}/attendance [patch]
func (h *BookingHandler) MarkAttendanceHandler(w http.ResponseWriter, r *http.Request) {
	userIDFromAuthVal := r.Context().Value(UserIDKey)
	userIDFromAuth, ok := userIDFromAuthVal.(int64)
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User ID not found in context or invalid type for auth", h.logger)
		return
	}

	bookingIDStr := chi.URLParam(r, "id")
	bookingID, err := strconv.ParseInt(bookingIDStr, 10, 64)
	if err != nil || bookingID <= 0 {
		RespondWithError(w, http.StatusBadRequest, "Invalid booking ID in path", h.logger, "booking_id_str", bookingIDStr)
		return
	}

	var req MarkAttendanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload", h.logger, "error", err.Error())
		return
	}
	defer r.Body.Close()

	updatedBooking, err := h.bookingService.MarkAttendance(r.Context(), bookingID, userIDFromAuth, req.Attended)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBookingNotFound):
			RespondWithError(w, http.StatusNotFound, "Booking not found", h.logger, "booking_id", bookingID)
		case errors.Is(err, service.ErrForbiddenUpdate):
			RespondWithError(w, http.StatusForbidden, "You are not authorized to update this booking's attendance", h.logger, "booking_id", bookingID)
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to mark attendance", h.logger, "error", err.Error())
		}
		return
	}

	// Convert to API response format
	bookingResponse := ToBookingResponse(updatedBooking)
	RespondWithJSON(w, http.StatusOK, bookingResponse, h.logger)
} 