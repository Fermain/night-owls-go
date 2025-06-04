package api

import (
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
	defer func() {
		if closeErr := r.Body.Close(); closeErr != nil {
			h.logger.ErrorContext(r.Context(), "Failed to close request body", "error", closeErr)
		}
	}()

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

// GetUserBookingsHandler handles GET /api/admin/users/{userId}/bookings
// @Summary Get all bookings for a specific user (Admin)
// @Description Allows an admin to view all bookings (past and future) for a specific user.
// @Tags admin-bookings
// @Produce json
// @Param userId path int true "User ID" example(42)
// @Success 200 {array} BookingWithScheduleResponse "List of bookings for the user with schedule names"
// @Failure 400 {object} ErrorResponse "Invalid user ID"
// @Failure 401 {object} ErrorResponse "Unauthorized - admin authentication required"
// @Failure 403 {object} ErrorResponse "Forbidden - user does not have admin privileges"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/admin/users/{userId}/bookings [get]
func (h *AdminBookingHandler) GetUserBookingsHandler(w http.ResponseWriter, r *http.Request) {
	// Try multiple methods to extract the ID parameter
	userIDStr := chi.URLParam(r, "userId")
	h.logger.InfoContext(r.Context(), "GetUserBookingsHandler called", "id_param", userIDStr, "url", r.URL.Path)

	// Alternative method: Parse from URL path directly if chi.URLParam fails
	if userIDStr == "" {
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) >= 5 && pathParts[0] == "api" && pathParts[1] == "admin" && pathParts[2] == "users" && pathParts[4] == "bookings" {
			userIDStr = pathParts[3]
			h.logger.InfoContext(r.Context(), "Extracted ID from path manually", "id_param", userIDStr)
		}
	}

	// Alternative method 2: Check request context for route values
	if userIDStr == "" {
		if rctx := chi.RouteContext(r.Context()); rctx != nil {
			for i, param := range rctx.URLParams.Keys {
				if param == "userId" && i < len(rctx.URLParams.Values) {
					userIDStr = rctx.URLParams.Values[i]
					h.logger.InfoContext(r.Context(), "Found ID in route context", "id_param", userIDStr)
					break
				}
			}
		}
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		h.logger.ErrorContext(r.Context(), "Failed to parse user ID", "id_param", userIDStr, "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID", h.logger, "user_id", userIDStr)
		return
	}

	bookings, err := h.bookingService.GetUserBookings(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			RespondWithError(w, http.StatusNotFound, "User not found", h.logger, "user_id", userID)
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to fetch user bookings", h.logger, "error", err.Error())
		}
		return
	}

	// Convert to API response format
	bookingResponses := make([]BookingWithScheduleResponse, len(bookings))
	for i, booking := range bookings {
		bookingResponses[i] = ToBookingWithScheduleResponse(booking)
	}

	RespondWithJSON(w, http.StatusOK, bookingResponses, h.logger)
}
