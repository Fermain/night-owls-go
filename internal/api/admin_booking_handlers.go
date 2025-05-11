package api

import (
	"encoding/json"
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

	// In a real scenario, the booking service method would be designed to accept a target UserID.
	// For now, let's assume bookingService.AdminCreateBookingForUser exists or will be created.
	// This service method would handle all business logic (validating user, schedule, slot, conflicts) 
	// and create the booking for the target req.UserID.

	// booking, err := h.bookingService.AdminCreateBookingForUser(r.Context(), req.UserID, req.ScheduleID, req.StartTime)
	// For now, we'll respond with a placeholder success, assuming the service call would work.
	// The actual implementation will depend on changes in booking_service.go

	h.logger.InfoContext(r.Context(), "Placeholder: Admin assigning user to shift", 
		"target_user_id", req.UserID, 
		"schedule_id", req.ScheduleID, 
		"start_time", req.StartTime.Format(time.RFC3339))

	// Placeholder response - replace with actual booking response after service implementation
	RespondWithJSON(w, http.StatusCreated, map[string]string{"message": "Shift assigned to user (placeholder)"}, h.logger)
} 