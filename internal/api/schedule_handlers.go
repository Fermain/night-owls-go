package api

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"night-owls-go/internal/service"
)

// ScheduleHandler handles schedule and shift availability related HTTP requests.
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
// It retrieves and returns all defined schedules (or active ones, current implementation fetches all via service).
func (h *ScheduleHandler) ListSchedulesHandler(w http.ResponseWriter, r *http.Request) {
	// For now, this directly calls a method that might be added to ScheduleService
	// or directly uses the querier if simple enough. The plan was ListActiveSchedules from querier.
	// Let's assume ScheduleService will have a method to list schedules (perhaps all for now for simplicity).

	// This endpoint is optional in the guide and not fully detailed for ScheduleService.
	// For now, let's return a placeholder or a direct query if that's easier.
	// To align with the service layer pattern, we would ideally add a `ListAllSchedules` to `ScheduleService`.
	// As a simplification for now, we'll acknowledge it's planned and might be empty or basic.
	
	// Placeholder response as the service method for this wasn't fully detailed in GetUpcomingAvailableSlots phase.
	// A full implementation would fetch from s.scheduleService.GetAllSchedules(r.Context())
	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Listing schedules - TBD or use GetUpcomingAvailableSlots"}, h.logger)
	// Example of how it *could* look if service.ListAllSchedules existed and returned []db.Schedule:
	// schedules, err := h.scheduleService.ListAllSchedules(r.Context())
	// if err != nil {
	// 	RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve schedules", h.logger, "error", err.Error())
	// 	return
	// }
	// RespondWithJSON(w, http.StatusOK, schedules, h.logger)
}

// ListAvailableShiftsHandler handles GET /shifts/available
func (h *ScheduleHandler) ListAvailableShiftsHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	var fromTime, toTime *time.Time
	var limit *int

	if fromStr := queryParams.Get("from"); fromStr != "" {
		t, err := time.Parse(time.RFC3339, fromStr) // Expect ISO 8601 / RFC3339 like "2024-05-10T00:00:00Z"
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid 'from' date format. Use RFC3339.", h.logger, "value", fromStr, "error", err.Error())
			return
		}
		fromTime = &t
	}

	if toStr := queryParams.Get("to"); toStr != "" {
		t, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid 'to' date format. Use RFC3339.", h.logger, "value", toStr, "error", err.Error())
			return
		}
		toTime = &t
	}

	if limitStr := queryParams.Get("limit"); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l <= 0 {
			RespondWithError(w, http.StatusBadRequest, "Invalid 'limit' value. Must be a positive integer.", h.logger, "value", limitStr, "error", err)
			return
		}
		limit = &l
	}

	availableSlots, err := h.scheduleService.GetUpcomingAvailableSlots(r.Context(), fromTime, toTime, limit)
	if err != nil {
		// The service layer already logs specifics, so we send a generic server error.
		// Specific mapping could be done if service returns typed errors (e.g., ErrInvalidInput)
		RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve available shifts", h.logger, "service_error", err.Error())
		return
	}

	if len(availableSlots) == 0 {
		RespondWithJSON(w, http.StatusOK, []service.AvailableShiftSlot{}, h.logger) // Return empty array, not null
		return
	}

	RespondWithJSON(w, http.StatusOK, availableSlots, h.logger)
} 