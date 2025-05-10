package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	db "night-owls-go/internal/db/sqlc_generated"
	"night-owls-go/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/gorhill/cronexpr"
)

// AdminScheduleHandlers holds handlers for admin schedule operations.
type AdminScheduleHandlers struct {
	logger          *slog.Logger
	scheduleService *service.ScheduleService
}

// NewAdminScheduleHandlers creates a new AdminScheduleHandlers.
func NewAdminScheduleHandlers(logger *slog.Logger, scheduleService *service.ScheduleService) *AdminScheduleHandlers {
	return &AdminScheduleHandlers{
		logger:          logger.With("handler", "AdminScheduleHandlers"),
		scheduleService: scheduleService,
	}
}

// AdminCreateScheduleRequest defines the expected JSON body for creating a schedule.
type AdminCreateScheduleRequest struct {
	Name            string  `json:"name"`
	CronExpr        string  `json:"cron_expr"`
	StartDate       *string `json:"start_date,omitempty"` // Expected format: "YYYY-MM-DD"
	EndDate         *string `json:"end_date,omitempty"`   // Expected format: "YYYY-MM-DD"
	Timezone        *string `json:"timezone,omitempty"`
}

func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

// AdminCreateSchedule handles POST /api/admin/schedules
func (h *AdminScheduleHandlers) AdminCreateSchedule(w http.ResponseWriter, r *http.Request) {
	var req AdminCreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WarnContext(r.Context(), "Failed to decode request body", "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload", h.logger)
		return
	}

	// Basic validation (more can be added)
	if req.Name == "" || req.CronExpr == "" {
		RespondWithError(w, http.StatusBadRequest, "Missing or invalid required fields (name, cron_expr)", h.logger)
		return
	}

	// Validate CRON expression format
	if _, err := cronexpr.Parse(req.CronExpr); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid CRON expression format", h.logger, "cron_expr", req.CronExpr, "error", err.Error())
		return
	}

	params := db.CreateScheduleParams{
		Name:            req.Name,
		CronExpr:        req.CronExpr,
	}

	if req.StartDate != nil && *req.StartDate != "" {
		parsedDate, err := parseDate(*req.StartDate)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid start_date format, expected YYYY-MM-DD", h.logger, "input_date", *req.StartDate, "error", err)
			return
		}
		params.StartDate.Time = parsedDate
		params.StartDate.Valid = true
	}
	if req.EndDate != nil && *req.EndDate != "" {
		parsedDate, err := parseDate(*req.EndDate)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid end_date format, expected YYYY-MM-DD", h.logger, "input_date", *req.EndDate, "error", err)
			return
		}
		params.EndDate.Time = parsedDate
		params.EndDate.Valid = true
	}
	if req.Timezone != nil && *req.Timezone != "" {
		params.Timezone.String = *req.Timezone
		params.Timezone.Valid = true
	}

	schedule, err := h.scheduleService.AdminCreateSchedule(r.Context(), params)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create schedule", h.logger, "db_params", params, "error", err)
		return
	}
	RespondWithJSON(w, http.StatusCreated, schedule, h.logger)
}

// AdminListSchedules handles GET /api/admin/schedules
func (h *AdminScheduleHandlers) AdminListSchedules(w http.ResponseWriter, r *http.Request) {
	schedules, err := h.scheduleService.ListAllSchedules(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to list schedules", h.logger, "error", err)
		return
	}
	RespondWithJSON(w, http.StatusOK, schedules, h.logger)
}

// AdminGetSchedule handles GET /api/admin/schedules/{id}
func (h *AdminScheduleHandlers) AdminGetSchedule(w http.ResponseWriter, r *http.Request) {
	scheduleIDStr := chi.URLParam(r, "id")
	scheduleID, err := strconv.ParseInt(scheduleIDStr, 10, 64)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid schedule ID format", h.logger, "schedule_id_str", scheduleIDStr, "error", err)
		return
	}

	schedule, err := h.scheduleService.AdminGetScheduleByID(r.Context(), scheduleID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Schedule not found", h.logger, "schedule_id", scheduleID)
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to get schedule", h.logger, "schedule_id", scheduleID, "error", err)
		return
	}
	RespondWithJSON(w, http.StatusOK, schedule, h.logger)
}

// AdminUpdateScheduleRequest reuses AdminCreateScheduleRequest for simplicity.
type AdminUpdateScheduleRequest AdminCreateScheduleRequest

// AdminUpdateSchedule handles PUT /api/admin/schedules/{id}
func (h *AdminScheduleHandlers) AdminUpdateSchedule(w http.ResponseWriter, r *http.Request) {
	scheduleIDStr := chi.URLParam(r, "id")
	scheduleID, err := strconv.ParseInt(scheduleIDStr, 10, 64)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid schedule ID format", h.logger, "schedule_id_str", scheduleIDStr, "error", err)
		return
	}

	var req AdminUpdateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WarnContext(r.Context(), "Failed to decode request body for update", "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload", h.logger)
		return
	}

	if req.Name == "" || req.CronExpr == "" {
		RespondWithError(w, http.StatusBadRequest, "Missing or invalid required fields (name, cron_expr)", h.logger)
		return
	}

	// Validate CRON expression format
	if _, err := cronexpr.Parse(req.CronExpr); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid CRON expression format", h.logger, "cron_expr", req.CronExpr, "error", err.Error())
		return
	}

	params := db.UpdateScheduleParams{
		ScheduleID:      scheduleID,
		Name:            req.Name,
		CronExpr:        req.CronExpr,
	}

	if req.StartDate != nil && *req.StartDate != "" {
		parsedDate, err := parseDate(*req.StartDate)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid start_date format, expected YYYY-MM-DD", h.logger, "input_date", *req.StartDate, "error", err)
			return
		}
		params.StartDate.Time = parsedDate
		params.StartDate.Valid = true
	}
	if req.EndDate != nil && *req.EndDate != "" {
		parsedDate, err := parseDate(*req.EndDate)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid end_date format, expected YYYY-MM-DD", h.logger, "input_date", *req.EndDate, "error", err)
			return
		}
		params.EndDate.Time = parsedDate
		params.EndDate.Valid = true
	}
	if req.Timezone != nil && *req.Timezone != "" {
		params.Timezone.String = *req.Timezone
		params.Timezone.Valid = true
	}

	schedule, err := h.scheduleService.AdminUpdateSchedule(r.Context(), params)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Schedule not found for update", h.logger, "schedule_id", scheduleID)
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to update schedule", h.logger, "db_params", params, "error", err)
		return
	}
	RespondWithJSON(w, http.StatusOK, schedule, h.logger)
}

// AdminDeleteSchedule handles DELETE /api/admin/schedules/{id}
func (h *AdminScheduleHandlers) AdminDeleteSchedule(w http.ResponseWriter, r *http.Request) {
	scheduleIDStr := chi.URLParam(r, "id")
	scheduleID, err := strconv.ParseInt(scheduleIDStr, 10, 64)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid schedule ID format", h.logger, "schedule_id_str", scheduleIDStr, "error", err)
		return
	}

	err = h.scheduleService.AdminDeleteSchedule(r.Context(), scheduleID)
	if err != nil {
		// Assuming service.ErrNotFound might not be returned by a simple delete if 0 rows affected.
		// If it can be, this is a more specific error for the client.
		if errors.Is(err, service.ErrNotFound) { // Defensive check
			RespondWithError(w, http.StatusNotFound, "Schedule not found for deletion", h.logger, "schedule_id", scheduleID)
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete schedule", h.logger, "schedule_id", scheduleID, "error", err)
		return
	}
	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Schedule deleted successfully"}, h.logger)
}

// AdminBulkDeleteSchedules handles requests to bulk delete schedules.
func (h *AdminScheduleHandlers) AdminBulkDeleteSchedules(w http.ResponseWriter, r *http.Request) {
	type BulkDeleteRequest struct {
		ScheduleIDs []int64 `json:"schedule_ids"`
	}
	var req BulkDeleteRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload", h.logger)
		return
	}

	if len(req.ScheduleIDs) == 0 {
		RespondWithError(w, http.StatusBadRequest, "No schedule IDs provided for deletion", h.logger)
		return
	}

	err := h.scheduleService.AdminBulkDeleteSchedules(r.Context(), req.ScheduleIDs)
	if err != nil {
		// Log the error for server-side observability
		slog.ErrorContext(r.Context(), "Error bulk deleting schedules", "error", err, "schedule_ids", req.ScheduleIDs)
		
		// Check for specific errors if needed, e.g., if some IDs were not found (though DELETE is often idempotent)
		// For a simple bulk delete, a generic server error might be sufficient if any part fails.
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete schedules", h.logger)
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Schedules deleted successfully"}, h.logger)
} 