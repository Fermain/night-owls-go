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
	ScheduleID int64     `json:"schedule_id"`
	StartTime  time.Time `json:"start_time"` // Expected in RFC3339 format e.g. "2025-05-10T18:00:00Z"
	BuddyPhone string    `json:"buddy_phone,omitempty"`
	BuddyName  string    `json:"buddy_name,omitempty"`
}

// CreateBookingHandler handles POST /bookings
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

	RespondWithJSON(w, http.StatusCreated, booking, h.logger)
}

// MarkAttendanceRequest is the expected JSON for PATCH /bookings/{id}/attendance
type MarkAttendanceRequest struct {
	Attended bool `json:"attended"`
}

// MarkAttendanceHandler handles PATCH /bookings/{id}/attendance
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

	RespondWithJSON(w, http.StatusOK, updatedBooking, h.logger)
} 