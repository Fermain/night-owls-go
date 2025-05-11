package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"night-owls-go/internal/service"
)

// AdminBookingHandler handles admin-specific booking operations.
type AdminBookingHandler struct {
	bookingService *service.BookingService
	logger         *slog.Logger
}

// NewAdminBookingHandler creates a new AdminBookingHandler.
func NewAdminBookingHandler(bookingService *service.BookingService, logger *slog.Logger) *AdminBookingHandler {
	return &AdminBookingHandler{
		bookingService: bookingService,
		logger:         logger.With("handler", "AdminBookingHandler"),
	}
}

// AssignUserToShiftRequest is the expected JSON for POST /api/admin/bookings/assign.
type AssignUserToShiftRequest struct {
	ScheduleID int64     `json:"schedule_id"`
	StartTime  time.Time `json:"start_time"`
	UserID     int64     `json:"user_id"`
}

// AssignUserToShiftHandler handles POST /api/admin/bookings/assign
// @Summary Assign a user to a specific shift slot (Admin)
// @Description Allows an admin to book a specific user for a given schedule ID and start time.
// @Tags admin-bookings
// @Accept json
// @Produce json
// @Param request body AssignUserToShiftRequest true "Assignment details (schedule_id, start_time, user_id)"
// @Success 201 {object} BookingResponse "Booking created successfully for the user"
// @Failure 400 {object} ErrorResponse "Invalid request format or data (e.g., missing fields, invalid user/schedule)"
// @Failure 401 {object} ErrorResponse "Unauthorized - admin authentication required"
// @Failure 403 {object} ErrorResponse "Forbidden - user does not have admin privileges"
// @Failure 404 {object} ErrorResponse "User or Schedule not found"
// @Failure 409 {object} ErrorResponse "Slot already booked or other conflict"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/admin/bookings/assign [post]
func (h *AdminBookingHandler) AssignUserToShiftHandler(w http.ResponseWriter, r *http.Request) {
	// Placeholder: Implement admin role check here if not handled by middleware more broadly
	// For example, check if r.Context().Value(UserRoleKey) == "admin"

	var req AssignUserToShiftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload", h.logger, "error", err.Error())
		return
	}
	defer r.Body.Close()

	if req.ScheduleID <= 0 || req.UserID <= 0 || req.StartTime.IsZero() {
		RespondWithError(w, http.StatusBadRequest, "Missing or invalid schedule_id, user_id, or start_time", h.logger)
		return
	}

	// Ensure StartTime is in UTC, as service layer and DB expect UTC.
	// The request JSON unmarshaler for time.Time typically parses RFC3339, which includes offset info.
	// Converting to UTC standardizes it.
	utcStartTime := req.StartTime.UTC()

	createdBooking, err := h.bookingService.AdminAssignUserToShift(r.Context(), req.UserID, req.ScheduleID, utcStartTime)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			RespondWithError(w, http.StatusNotFound, "Target user not found", h.logger, "user_id", req.UserID, "error", err.Error())
		case errors.Is(err, service.ErrScheduleNotFound):
			RespondWithError(w, http.StatusNotFound, "Schedule not found", h.logger, "schedule_id", req.ScheduleID, "error", err.Error())
		case errors.Is(err, service.ErrShiftTimeInvalid):
			RespondWithError(w, http.StatusBadRequest, "Invalid shift start time for the schedule", h.logger, "schedule_id", req.ScheduleID, "start_time", req.StartTime, "error", err.Error())
		case errors.Is(err, service.ErrBookingConflict):
			RespondWithError(w, http.StatusConflict, "Shift slot is already booked or a conflict exists", h.logger, "schedule_id", req.ScheduleID, "start_time", req.StartTime, "error", err.Error())
		case errors.Is(err, service.ErrInternalServer):
			RespondWithError(w, http.StatusInternalServerError, "Internal server error processing assignment", h.logger, "error", err.Error())
		default:
			RespondWithError(w, http.StatusInternalServerError, "An unexpected error occurred", h.logger, "error", err.Error())
		}
		return
	}

	bookingResponse := ToBookingResponse(createdBooking)
	RespondWithJSON(w, http.StatusCreated, bookingResponse, h.logger)
} 