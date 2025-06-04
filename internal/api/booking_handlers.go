package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"night-owls-go/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-fuego/fuego"

	db "night-owls-go/internal/db/sqlc_generated"
)

// ErrorResponse represents an error response in the API
// Used for Swagger documentation
type ErrorResponse struct {
	Error string `json:"error"`
}

// BookingHandler handles user-facing booking operations.
type BookingHandler struct {
	service      *service.BookingService
	auditService *service.AuditService
	querier      db.Querier
	logger       *slog.Logger
}

// NewBookingHandler creates a new BookingHandler.
func NewBookingHandler(service *service.BookingService, auditService *service.AuditService, querier db.Querier, logger *slog.Logger) *BookingHandler {
	return &BookingHandler{
		service:      service,
		auditService: auditService,
		querier:      querier,
		logger:       logger,
	}
}

// Request types for Fuego
type CreateBookingRequest struct {
	ScheduleID int64     `json:"schedule_id" validate:"required"`
	StartTime  time.Time `json:"start_time" validate:"required"`
	BuddyName  *string   `json:"buddy_name,omitempty"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

// Helper function to convert db.Booking to BookingResponse
func toBookingResponse(booking db.Booking) *BookingResponse {
	resp := &BookingResponse{
		BookingID:  booking.BookingID,
		UserID:     booking.UserID,
		ScheduleID: booking.ScheduleID,
		ShiftStart: booking.ShiftStart,
		ShiftEnd:   booking.ShiftEnd,
	}

	// Handle buddy user ID
	if booking.BuddyUserID.Valid {
		resp.BuddyUserID = &booking.BuddyUserID.Int64
	}

	// Handle buddy name
	if booking.BuddyName.Valid {
		resp.BuddyName = booking.BuddyName.String
	}

	// Handle checked in at
	if booking.CheckedInAt.Valid {
		resp.CheckedInAt = &booking.CheckedInAt.Time
	}

	// Handle created at
	if booking.CreatedAt.Valid {
		resp.CreatedAt = booking.CreatedAt.Time
	}

	return resp
}

// mapServiceErrorToHTTP maps service layer errors to appropriate HTTP status codes
func mapServiceErrorToHTTP(err error) (int, string) {
	switch {
	case errors.Is(err, service.ErrBookingConflict):
		return http.StatusConflict, err.Error()
	case errors.Is(err, service.ErrBookingNotFound):
		return http.StatusNotFound, err.Error()
	case errors.Is(err, service.ErrScheduleNotFound):
		return http.StatusNotFound, err.Error()
	case errors.Is(err, service.ErrUserNotFound):
		return http.StatusNotFound, err.Error()
	case errors.Is(err, service.ErrForbiddenUpdate):
		return http.StatusForbidden, err.Error()
	case errors.Is(err, service.ErrShiftTimeInvalid):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, service.ErrCheckInTooEarly):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, service.ErrBookingCannotBeCancelled):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, service.ErrInternalServer):
		return http.StatusInternalServerError, "Internal server error"
	default:
		return http.StatusBadRequest, err.Error()
	}
}

// CreateBookingFuego handles POST /bookings with Fuego typed handler
// @Summary Create a new booking
// @Description Create a new booking for a shift
// @Tags bookings
// @Accept json
// @Produce json
// @Param booking body CreateBookingRequest true "Booking details"
// @Success 201 {object} BookingResponse "Booking created successfully"
// @Failure 400 {object} ErrorResponse "Invalid request data"
// @Failure 401 {object} ErrorResponse "Unauthorized - authentication required"
// @Failure 409 {object} ErrorResponse "Shift already booked or not available"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /bookings [post]
func (h *BookingHandler) CreateBookingFuego(c fuego.ContextWithBody[CreateBookingRequest]) (*BookingResponse, error) {
	// Get request body
	req, err := c.Body()
	if err != nil {
		h.logger.ErrorContext(c.Context(), "Failed to parse request body", "error", err)
		return nil, fuego.BadRequestError{
			Err:    err,
			Detail: "Invalid request body",
		}
	}

	// Get the user ID from context (set by auth middleware)
	userID, ok := c.Context().Value(UserIDKey).(int64)
	if !ok {
		h.logger.ErrorContext(c.Context(), "User ID not found in context", "handler", "BookingHandler")
		return nil, fuego.UnauthorizedError{
			Err:    nil,
			Detail: "User authentication required",
		}
	}

	h.logger.InfoContext(c.Context(), "Creating booking", "schedule_id", req.ScheduleID, "start_time", req.StartTime, "user_id", userID, "buddy_name", req.BuddyName)

	// Convert buddy name to sql.NullString
	var buddyName sql.NullString
	if req.BuddyName != nil && *req.BuddyName != "" {
		buddyName = sql.NullString{String: *req.BuddyName, Valid: true}
	}

	booking, err := h.service.CreateBooking(c.Context(), userID, req.ScheduleID, req.StartTime, sql.NullString{}, buddyName)
	if err != nil {
		h.logger.ErrorContext(c.Context(), "Failed to create booking", "schedule_id", req.ScheduleID, "start_time", req.StartTime, "user_id", userID, "error", err)
		status, detail := mapServiceErrorToHTTP(err)
		return nil, fuego.HTTPError{
			Err:    err,
			Status: status,
			Detail: detail,
		}
	}

	h.logger.InfoContext(c.Context(), "Booking created successfully", "booking_id", booking.BookingID, "schedule_id", req.ScheduleID, "start_time", req.StartTime, "user_id", userID)

	// Log audit event for booking creation
	ipAddress := c.Request().Header.Get("X-Forwarded-For")
	if ipAddress == "" {
		ipAddress = c.Request().Header.Get("X-Real-IP")
	}
	if ipAddress == "" {
		ipAddress = c.Request().RemoteAddr
	}
	userAgent := c.Request().Header.Get("User-Agent")

	// Get schedule name for audit logging (we need to look it up)
	// For now, we'll use the schedule ID as a string since we don't have the schedule name easily available
	scheduleName := fmt.Sprintf("Schedule %d", req.ScheduleID)

	var buddyNamePtr *string
	if buddyName.Valid && buddyName.String != "" {
		buddyNamePtr = &buddyName.String
	}

	auditErr := h.auditService.LogBookingCreated(
		c.Context(),
		userID,
		booking.BookingID,
		req.ScheduleID,
		scheduleName,
		booking.ShiftStart.Format(time.RFC3339),
		booking.ShiftEnd.Format(time.RFC3339),
		buddyNamePtr,
		ipAddress,
		userAgent,
	)
	if auditErr != nil {
		h.logger.WarnContext(c.Context(), "Failed to log booking creation audit event", "booking_id", booking.BookingID, "error", auditErr)
	}

	// Set the success status code for creation
	c.SetStatus(http.StatusCreated)

	return toBookingResponse(booking), nil
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

	bookings, err := h.service.GetUserBookings(r.Context(), userID)
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

// MarkCheckInFuego handles POST /bookings/{id}/checkin with Fuego typed handler
// @Summary Mark check-in for a booking
// @Description Mark a user as checked in for their booking
// @Tags bookings
// @Produce json
// @Param id path int true "Booking ID"
// @Success 200 {object} SuccessResponse "Check-in recorded successfully"
// @Failure 400 {object} ErrorResponse "Invalid booking ID or check-in not allowed"
// @Failure 401 {object} ErrorResponse "Unauthorized - authentication required"
// @Failure 403 {object} ErrorResponse "Not authorized to check in for this booking"
// @Failure 404 {object} ErrorResponse "Booking not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /bookings/{id}/checkin [post]
func (h *BookingHandler) MarkCheckInFuego(c fuego.ContextNoBody) (SuccessResponse, error) {
	// Extract booking ID from URL using Fuego's context
	bookingIDStr := c.PathParam("id")
	h.logger.InfoContext(c.Context(), "MarkCheckInFuego called", "id_param", bookingIDStr, "url", c.Request().URL.Path)

	if bookingIDStr == "" {
		h.logger.ErrorContext(c.Context(), "Booking ID not found in path", "handler", "BookingHandler")
		return SuccessResponse{}, fuego.BadRequestError{
			Err:    nil,
			Detail: "Invalid booking ID",
		}
	}

	bookingID, err := strconv.ParseInt(bookingIDStr, 10, 64)
	if err != nil {
		h.logger.ErrorContext(c.Context(), "Failed to parse booking ID", "handler", "BookingHandler", "id_param", bookingIDStr, "error", err)
		return SuccessResponse{}, fuego.BadRequestError{
			Err:    err,
			Detail: "Invalid booking ID",
		}
	}

	// Get the user ID from context (set by auth middleware)
	userID, ok := c.Context().Value(UserIDKey).(int64)
	if !ok {
		h.logger.ErrorContext(c.Context(), "User ID not found in context", "handler", "BookingHandler")
		return SuccessResponse{}, fuego.UnauthorizedError{
			Err:    nil,
			Detail: "User authentication required",
		}
	}

	h.logger.InfoContext(c.Context(), "Processing check-in", "booking_id", bookingID, "user_id", userID)

	// Note: MarkCheckIn returns (db.Booking, error) but we don't need the booking for the response
	updatedBooking, err := h.service.MarkCheckIn(c.Context(), bookingID, userID)
	if err != nil {
		h.logger.ErrorContext(c.Context(), "Failed to mark check-in", "booking_id", bookingID, "user_id", userID, "error", err)
		status, detail := mapServiceErrorToHTTP(err)
		return SuccessResponse{}, fuego.HTTPError{
			Err:    err,
			Status: status,
			Detail: detail,
		}
	}

	h.logger.InfoContext(c.Context(), "Check-in recorded successfully", "booking_id", bookingID, "user_id", userID)

	// Log audit event for booking check-in
	ipAddress := c.Request().Header.Get("X-Forwarded-For")
	if ipAddress == "" {
		ipAddress = c.Request().Header.Get("X-Real-IP")
	}
	if ipAddress == "" {
		ipAddress = c.Request().RemoteAddr
	}
	userAgent := c.Request().Header.Get("User-Agent")

	scheduleName := fmt.Sprintf("Schedule %d", updatedBooking.ScheduleID)
	auditErr := h.auditService.LogBookingCheckedIn(
		c.Context(),
		userID,
		bookingID,
		updatedBooking.ScheduleID,
		scheduleName,
		updatedBooking.ShiftStart.Format(time.RFC3339),
		ipAddress,
		userAgent,
	)
	if auditErr != nil {
		h.logger.WarnContext(c.Context(), "Failed to log booking check-in audit event", "booking_id", bookingID, "error", auditErr)
	}

	return SuccessResponse{Message: "Check-in recorded successfully"}, nil
}

// CancelBookingFuego handles DELETE /bookings/{id} with Fuego typed handler
// @Summary Cancel a booking
// @Description Cancel a user's booking if it's not too close to the shift start time
// @Tags bookings
// @Param id path int true "Booking ID"
// @Success 204 "Booking cancelled successfully"
// @Failure 400 {object} ErrorResponse "Invalid booking ID or booking cannot be cancelled"
// @Failure 401 {object} ErrorResponse "Unauthorized - authentication required"
// @Failure 403 {object} ErrorResponse "Not authorized to cancel this booking"
// @Failure 404 {object} ErrorResponse "Booking not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /bookings/{id} [delete]
func (h *BookingHandler) CancelBookingFuego(c fuego.ContextNoBody) (any, error) {
	// Extract booking ID from URL using Fuego's context
	bookingIDStr := c.PathParam("id")
	h.logger.InfoContext(c.Context(), "CancelBookingFuego called", "id_param", bookingIDStr, "url", c.Request().URL.Path)

	if bookingIDStr == "" {
		h.logger.ErrorContext(c.Context(), "Booking ID not found in path", "handler", "BookingHandler")
		return nil, fuego.BadRequestError{
			Err:    nil,
			Detail: "Invalid booking ID",
		}
	}

	bookingID, err := strconv.ParseInt(bookingIDStr, 10, 64)
	if err != nil {
		h.logger.ErrorContext(c.Context(), "Failed to parse booking ID", "handler", "BookingHandler", "id_param", bookingIDStr, "error", err)
		return nil, fuego.BadRequestError{
			Err:    err,
			Detail: "Invalid booking ID",
		}
	}

	// Get the user ID from context (set by auth middleware)
	userID, ok := c.Context().Value(UserIDKey).(int64)
	if !ok {
		h.logger.ErrorContext(c.Context(), "User ID not found in context", "handler", "BookingHandler")
		return nil, fuego.UnauthorizedError{
			Err:    nil,
			Detail: "User authentication required",
		}
	}

	h.logger.InfoContext(c.Context(), "Processing booking cancellation", "booking_id", bookingID, "user_id", userID)

	// Get booking details before cancellation for audit logging
	bookingDetails, bookingErr := h.querier.GetBookingByID(c.Context(), bookingID)
	var scheduleID int64
	var scheduleName string
	var shiftStart, shiftEnd string
	if bookingErr == nil {
		scheduleID = bookingDetails.ScheduleID
		scheduleName = fmt.Sprintf("Schedule %d", scheduleID)
		shiftStart = bookingDetails.ShiftStart.Format(time.RFC3339)
		shiftEnd = bookingDetails.ShiftEnd.Format(time.RFC3339)
	}

	err = h.service.CancelBooking(c.Context(), bookingID, userID)
	if err != nil {
		h.logger.ErrorContext(c.Context(), "Failed to cancel booking", "booking_id", bookingID, "user_id", userID, "error", err)
		status, detail := mapServiceErrorToHTTP(err)
		return nil, fuego.HTTPError{
			Err:    err,
			Status: status,
			Detail: detail,
		}
	}

	h.logger.InfoContext(c.Context(), "Booking cancelled successfully", "booking_id", bookingID, "user_id", userID)

	// Log audit event for booking cancellation
	if bookingErr == nil {
		ipAddress := c.Request().Header.Get("X-Forwarded-For")
		if ipAddress == "" {
			ipAddress = c.Request().Header.Get("X-Real-IP")
		}
		if ipAddress == "" {
			ipAddress = c.Request().RemoteAddr
		}
		userAgent := c.Request().Header.Get("User-Agent")

		auditErr := h.auditService.LogBookingCancelled(
			c.Context(),
			userID,
			bookingID,
			scheduleID,
			scheduleName,
			shiftStart,
			shiftEnd,
			ipAddress,
			userAgent,
		)
		if auditErr != nil {
			h.logger.WarnContext(c.Context(), "Failed to log booking cancellation audit event", "booking_id", bookingID, "error", auditErr)
		}
	}

	// For 204 No Content, we manually set the status and don't return any content
	c.Response().WriteHeader(http.StatusNoContent)

	// Return nil with no error to prevent Fuego from trying to serialize anything
	return nil, nil
}

// Legacy handlers - keeping for backwards compatibility during migration
// These will be removed once the migration is complete

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
	// Debug: Log the full request details
	h.logger.InfoContext(r.Context(), "CancelBookingHandler - Full request details",
		"method", r.Method,
		"url", r.URL.String(),
		"path", r.URL.Path,
		"raw_path", r.URL.RawPath)

	// Extract booking ID from URL
	bookingIDStr := chi.URLParam(r, "id")
	h.logger.InfoContext(r.Context(), "CancelBookingHandler called", "id_param", bookingIDStr, "url", r.URL.Path)

	// Alternative method: Parse from URL path directly if chi.URLParam fails
	if bookingIDStr == "" {
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		h.logger.InfoContext(r.Context(), "Path parts for manual extraction", "path_parts", pathParts, "path", r.URL.Path)

		// Handle both /bookings/14 and /api/bookings/14 patterns
		for i, part := range pathParts {
			if part == "bookings" && i+1 < len(pathParts) {
				bookingIDStr = pathParts[i+1]
				h.logger.InfoContext(r.Context(), "Extracted ID from path manually", "id_param", bookingIDStr)
				break
			}
		}

		// If still empty, try a simpler approach - just get the last part of the path
		if bookingIDStr == "" && len(pathParts) > 0 {
			bookingIDStr = pathParts[len(pathParts)-1]
			h.logger.InfoContext(r.Context(), "Using last path segment as ID", "id_param", bookingIDStr)
		}
	}

	if bookingIDStr == "" {
		h.logger.ErrorContext(r.Context(), "Could not extract booking ID from path", "url", r.URL.Path)
		RespondWithError(w, http.StatusBadRequest, "Invalid booking ID", h.logger)
		return
	}

	bookingID, err := strconv.ParseInt(bookingIDStr, 10, 64)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to parse booking ID", "handler", "BookingHandler", "id_param", bookingIDStr, "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid booking ID", h.logger)
		return
	}

	// Get the user ID from context (set by auth middleware)
	userID, ok := r.Context().Value(UserIDKey).(int64)
	if !ok {
		h.logger.ErrorContext(r.Context(), "User ID not found in context", "handler", "BookingHandler")
		RespondWithError(w, http.StatusUnauthorized, "User authentication required", h.logger)
		return
	}

	h.logger.InfoContext(r.Context(), "Processing booking cancellation", "booking_id", bookingID, "user_id", userID)

	err = h.service.CancelBooking(r.Context(), bookingID, userID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to cancel booking", "booking_id", bookingID, "user_id", userID, "error", err)
		status, detail := mapServiceErrorToHTTP(err)
		RespondWithError(w, status, detail, h.logger)
		return
	}

	h.logger.InfoContext(r.Context(), "Booking cancelled successfully", "booking_id", bookingID, "user_id", userID)
	w.WriteHeader(http.StatusNoContent)
}

// MarkCheckInHandler handles POST /bookings/{id}/checkin
// @Summary Mark check-in for a booking
// @Description Mark a user as checked in for their booking
// @Tags bookings
// @Produce json
// @Param id path int true "Booking ID"
// @Success 200 {object} BookingResponse "Check-in recorded successfully"
// @Failure 400 {object} ErrorResponse "Invalid booking ID or check-in not allowed"
// @Failure 401 {object} ErrorResponse "Unauthorized - authentication required"
// @Failure 403 {object} ErrorResponse "Not authorized to check in for this booking"
// @Failure 404 {object} ErrorResponse "Booking not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /bookings/{id}/checkin [post]
func (h *BookingHandler) MarkCheckInHandler(w http.ResponseWriter, r *http.Request) {
	// Extract booking ID from URL
	bookingIDStr := chi.URLParam(r, "id")
	h.logger.InfoContext(r.Context(), "MarkCheckInHandler called", "id_param", bookingIDStr, "url", r.URL.Path)

	// Alternative method: Parse from URL path directly if chi.URLParam fails
	if bookingIDStr == "" {
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) >= 3 && pathParts[0] == "bookings" && pathParts[2] == "checkin" {
			bookingIDStr = pathParts[1]
			h.logger.InfoContext(r.Context(), "Extracted ID from path manually", "id_param", bookingIDStr)
		}
	}

	bookingID, err := strconv.ParseInt(bookingIDStr, 10, 64)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to parse booking ID", "handler", "BookingHandler", "id_param", bookingIDStr, "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid booking ID", h.logger)
		return
	}

	// Get the user ID from context (set by auth middleware)
	userID, ok := r.Context().Value(UserIDKey).(int64)
	if !ok {
		h.logger.ErrorContext(r.Context(), "User ID not found in context", "handler", "BookingHandler")
		RespondWithError(w, http.StatusUnauthorized, "User authentication required", h.logger)
		return
	}

	h.logger.InfoContext(r.Context(), "Processing check-in", "booking_id", bookingID, "user_id", userID)

	// MarkCheckIn returns the updated booking with check-in timestamp
	updatedBooking, err := h.service.MarkCheckIn(r.Context(), bookingID, userID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to mark check-in", "booking_id", bookingID, "user_id", userID, "error", err)
		status, detail := mapServiceErrorToHTTP(err)
		RespondWithError(w, status, detail, h.logger)
		return
	}

	h.logger.InfoContext(r.Context(), "Check-in recorded successfully", "booking_id", bookingID, "user_id", userID)
	RespondWithJSON(w, http.StatusOK, toBookingResponse(updatedBooking), h.logger)
}

// CreateBookingHandler handles POST /bookings (legacy chi handler for tests)
// @Summary Create a new booking
// @Description Create a new booking for a shift
// @Tags bookings
// @Accept json
// @Produce json
// @Param booking body CreateBookingRequest true "Booking details"
// @Success 201 {object} BookingResponse "Booking created successfully"
// @Failure 400 {object} ErrorResponse "Invalid request data"
// @Failure 401 {object} ErrorResponse "Unauthorized - authentication required"
// @Failure 409 {object} ErrorResponse "Shift already booked or not available"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /bookings [post]
func (h *BookingHandler) CreateBookingHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to parse request body", "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", h.logger)
		return
	}

	// Get the user ID from context (set by auth middleware)
	userID, ok := r.Context().Value(UserIDKey).(int64)
	if !ok {
		h.logger.ErrorContext(r.Context(), "User ID not found in context", "handler", "BookingHandler")
		RespondWithError(w, http.StatusUnauthorized, "User authentication required", h.logger)
		return
	}

	h.logger.InfoContext(r.Context(), "Creating booking", "schedule_id", req.ScheduleID, "start_time", req.StartTime, "user_id", userID, "buddy_name", req.BuddyName)

	// Convert buddy name to sql.NullString
	var buddyName sql.NullString
	if req.BuddyName != nil && *req.BuddyName != "" {
		buddyName = sql.NullString{String: *req.BuddyName, Valid: true}
	}

	booking, err := h.service.CreateBooking(r.Context(), userID, req.ScheduleID, req.StartTime, sql.NullString{}, buddyName)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to create booking", "schedule_id", req.ScheduleID, "start_time", req.StartTime, "user_id", userID, "error", err)
		status, detail := mapServiceErrorToHTTP(err)
		RespondWithError(w, status, detail, h.logger)
		return
	}

	h.logger.InfoContext(r.Context(), "Booking created successfully", "booking_id", booking.BookingID, "schedule_id", req.ScheduleID, "start_time", req.StartTime, "user_id", userID)

	RespondWithJSON(w, http.StatusCreated, toBookingResponse(booking), h.logger)
}
