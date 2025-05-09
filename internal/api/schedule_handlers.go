package api

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"night-owls-go/internal/service"
)

// ScheduleHandler handles schedule-related HTTP requests.
type ScheduleHandler struct {
	scheduleService *service.ScheduleService
	logger          *slog.Logger
}

// NewScheduleHandler creates a new ScheduleHandler.
func NewScheduleHandler(scheduleService *service.ScheduleService, logger *slog.Logger) *ScheduleHandler {
	return &ScheduleHandler{
		scheduleService: scheduleService,
		logger:          logger.With("handler", "ScheduleHandler"),
	}
}

// ListSchedulesHandler handles GET /schedules
// @Summary List all schedules
// @Description Returns a list of all defined schedules
// @Tags schedules
// @Produce json
// @Success 200 {array} ScheduleResponse "List of schedules"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /schedules [get]
func (h *ScheduleHandler) ListSchedulesHandler(w http.ResponseWriter, r *http.Request) {
	// Fetch all schedules from the database
	schedules, err := h.scheduleService.ListAllSchedules(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve schedules", h.logger, "error", err.Error())
		return
	}

	// Convert to API response format
	scheduleResponses := ToScheduleResponses(schedules)
	RespondWithJSON(w, http.StatusOK, scheduleResponses, h.logger)
}

// ListAvailableShiftsHandler handles GET /shifts/available
// @Summary List available shift slots
// @Description Returns a list of available shift slots based on schedule definitions
// @Tags shifts
// @Produce json
// @Param from query string false "Start date for shift window (RFC3339 format)"
// @Param to query string false "End date for shift window (RFC3339 format)"
// @Param limit query int false "Maximum number of shifts to return"
// @Success 200 {array} service.AvailableShiftSlot "List of available shift slots"
// @Failure 400 {object} ErrorResponse "Invalid query parameters"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /shifts/available [get]
func (h *ScheduleHandler) ListAvailableShiftsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse optional query parameters for date range and limit
	var fromTime *time.Time
	var toTime *time.Time
	var limit *int

	fromStr := r.URL.Query().Get("from")
	if fromStr != "" {
		parsedFromTime, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid 'from' date format, use RFC3339", h.logger)
			return
		}
		fromTime = &parsedFromTime
	}

	toStr := r.URL.Query().Get("to")
	if toStr != "" {
		parsedToTime, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid 'to' date format, use RFC3339", h.logger)
			return
		}
		toTime = &parsedToTime
	}

	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit <= 0 {
			RespondWithError(w, http.StatusBadRequest, "Invalid 'limit' parameter, must be a positive integer", h.logger)
			return
		}
		limit = &parsedLimit
	}

	// Get available shifts from the service
	shifts, err := h.scheduleService.GetUpcomingAvailableSlots(r.Context(), fromTime, toTime, limit)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve available shifts", h.logger)
		return
	}

	RespondWithJSON(w, http.StatusOK, shifts, h.logger)
} 