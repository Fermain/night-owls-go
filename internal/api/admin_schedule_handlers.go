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

	db "night-owls-go/internal/db/sqlc_generated"
	"night-owls-go/internal/service"

	"github.com/gorhill/cronexpr"
)

// AdminScheduleHandlers holds handlers for admin schedule operations.
type AdminScheduleHandlers struct {
	logger          *slog.Logger
	scheduleService *service.ScheduleService
	auditService    *service.AuditService
}

// NewAdminScheduleHandlers creates a new AdminScheduleHandlers.
func NewAdminScheduleHandlers(logger *slog.Logger, scheduleService *service.ScheduleService, auditService *service.AuditService) *AdminScheduleHandlers {
	return &AdminScheduleHandlers{
		logger:          logger.With("handler", "AdminScheduleHandlers"),
		scheduleService: scheduleService,
		auditService:    auditService,
	}
}

// AdminCreateScheduleRequest defines the expected JSON body for creating a schedule.
type AdminCreateScheduleRequest struct {
	Name      string  `json:"name"`
	CronExpr  string  `json:"cron_expr"`
	StartDate *string `json:"start_date,omitempty"` // Expected format: "YYYY-MM-DD"
	EndDate   *string `json:"end_date,omitempty"`   // Expected format: "YYYY-MM-DD"
	Timezone  *string `json:"timezone,omitempty"`
}

// parseDateToUTC parses a "YYYY-MM-DD" string and returns a time.Time
// representing 00:00:00 UTC on that date.
func parseDateToUTC(dateStr string) (time.Time, error) {
	t, err := time.ParseInLocation("2006-01-02", dateStr, time.UTC) // Parse as UTC directly
	if err != nil {
		return time.Time{}, err
	}
	// Ensure it's specifically 00:00:00 UTC, though ParseInLocation with date-only should do this.
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
}

// AdminCreateSchedule handles POST /api/admin/schedules
func (h *AdminScheduleHandlers) AdminCreateSchedule(w http.ResponseWriter, r *http.Request) {
	var req AdminCreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WarnContext(r.Context(), "Failed to decode request body", "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload", h.logger)
		return
	}

	if req.Name == "" || req.CronExpr == "" {
		RespondWithError(w, http.StatusBadRequest, "Missing required fields (name, cron_expr)", h.logger)
		return
	}

	if _, err := cronexpr.Parse(req.CronExpr); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid CRON expression format", h.logger, "cron_expr", req.CronExpr, "error", err.Error())
		return
	}

	// Validate Timezone if provided
	if req.Timezone != nil && *req.Timezone != "" {
		if _, err := time.LoadLocation(*req.Timezone); err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid timezone string", h.logger, "timezone", *req.Timezone, "error", err.Error())
			return
		}
	}

	params := db.CreateScheduleParams{
		Name:     req.Name,
		CronExpr: req.CronExpr,
	}

	if req.StartDate != nil && *req.StartDate != "" {
		parsedDate, err := parseDateToUTC(*req.StartDate) // Use UTC parser
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid start_date format, expected YYYY-MM-DD", h.logger, "input_date", *req.StartDate, "error", err)
			return
		}
		params.StartDate = sql.NullTime{Time: parsedDate, Valid: true} // Assign to sql.NullTime
	}
	if req.EndDate != nil && *req.EndDate != "" {
		parsedDate, err := parseDateToUTC(*req.EndDate) // Use UTC parser
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid end_date format, expected YYYY-MM-DD", h.logger, "input_date", *req.EndDate, "error", err)
			return
		}
		params.EndDate = sql.NullTime{Time: parsedDate, Valid: true} // Assign to sql.NullTime
	}

	// Set timezone - default to Africa/Johannesburg if not provided
	timezone := "Africa/Johannesburg"
	if req.Timezone != nil && *req.Timezone != "" {
		timezone = *req.Timezone
	}
	params.Timezone = sql.NullString{String: timezone, Valid: true}

	schedule, err := h.scheduleService.AdminCreateSchedule(r.Context(), params)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create schedule", h.logger, "db_params", params, "error", err)
		return
	}

	// Log audit event for schedule creation
	var timezonePtr *string
	if params.Timezone.Valid {
		timezonePtr = &params.Timezone.String
	}

	WithAuditLogging(r.Context(), h.logger, func(userID int64, ipAddress, userAgent string) error {
		return h.auditService.LogScheduleCreated(
			r.Context(),
			userID,
			schedule.ScheduleID,
			schedule.Name,
			schedule.CronExpr,
			timezonePtr,
			schedule.DurationMinutes,
			ipAddress,
			userAgent,
		)
	})

	RespondWithJSON(w, http.StatusCreated, ToScheduleResponse(schedule), h.logger)
}

// AdminListSchedules handles GET /api/admin/schedules
func (h *AdminScheduleHandlers) AdminListSchedules(w http.ResponseWriter, r *http.Request) {
	schedules, err := h.scheduleService.ListAllSchedules(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to list schedules", h.logger, "error", err)
		return
	}
	RespondWithJSON(w, http.StatusOK, ToScheduleResponses(schedules), h.logger)
}

// AdminGetSchedule handles GET /api/admin/schedules/{id}
func (h *AdminScheduleHandlers) AdminGetSchedule(w http.ResponseWriter, r *http.Request) {
	// Try multiple methods to extract the ID parameter
	scheduleIDStr := r.PathValue("id")
	h.logger.InfoContext(r.Context(), "AdminGetSchedule called", "id_param", scheduleIDStr, "url", r.URL.Path)

	// Alternative method: Parse from URL path directly if r.PathValue fails
	if scheduleIDStr == "" {
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) >= 4 && pathParts[0] == "api" && pathParts[1] == "admin" && pathParts[2] == "schedules" {
			scheduleIDStr = pathParts[3]
			h.logger.InfoContext(r.Context(), "Extracted ID from path manually", "id_param", scheduleIDStr)
		}
	}

	scheduleID, err := strconv.ParseInt(scheduleIDStr, 10, 64)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to parse schedule ID", "id_param", scheduleIDStr, "error", err)
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
	RespondWithJSON(w, http.StatusOK, ToScheduleResponse(schedule), h.logger)
}

// AdminUpdateScheduleRequest reuses AdminCreateScheduleRequest for simplicity.
type AdminUpdateScheduleRequest AdminCreateScheduleRequest

// AdminUpdateSchedule handles PUT /api/admin/schedules/{id}
func (h *AdminScheduleHandlers) AdminUpdateSchedule(w http.ResponseWriter, r *http.Request) {
	// Try multiple methods to extract the ID parameter
	scheduleIDStr := r.PathValue("id")
	h.logger.InfoContext(r.Context(), "AdminUpdateSchedule called", "id_param", scheduleIDStr, "url", r.URL.Path)

	// Alternative method: Parse from URL path directly if r.PathValue fails
	if scheduleIDStr == "" {
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) >= 4 && pathParts[0] == "api" && pathParts[1] == "admin" && pathParts[2] == "schedules" {
			scheduleIDStr = pathParts[3]
			h.logger.InfoContext(r.Context(), "Extracted ID from path manually", "id_param", scheduleIDStr)
		}
	}

	scheduleID, err := strconv.ParseInt(scheduleIDStr, 10, 64)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to parse schedule ID", "id_param", scheduleIDStr, "error", err)
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

	if _, err := cronexpr.Parse(req.CronExpr); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid CRON expression format", h.logger, "cron_expr", req.CronExpr, "error", err.Error())
		return
	}

	// Validate Timezone if provided
	if req.Timezone != nil && *req.Timezone != "" {
		if _, err := time.LoadLocation(*req.Timezone); err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid timezone string", h.logger, "timezone", *req.Timezone, "error", err.Error())
			return
		}
	}

	params := db.UpdateScheduleParams{
		ScheduleID: scheduleID,
		Name:       req.Name,
		CronExpr:   req.CronExpr,
	}

	if req.StartDate != nil && *req.StartDate != "" {
		parsedDate, err := parseDateToUTC(*req.StartDate) // Use UTC parser
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid start_date format, expected YYYY-MM-DD", h.logger, "input_date", *req.StartDate, "error", err)
			return
		}
		params.StartDate = sql.NullTime{Time: parsedDate, Valid: true} // Assign to sql.NullTime
	}
	if req.EndDate != nil && *req.EndDate != "" {
		parsedDate, err := parseDateToUTC(*req.EndDate) // Use UTC parser
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid end_date format, expected YYYY-MM-DD", h.logger, "input_date", *req.EndDate, "error", err)
			return
		}
		params.EndDate = sql.NullTime{Time: parsedDate, Valid: true} // Assign to sql.NullTime
	}

	// Set timezone - default to Africa/Johannesburg if not provided
	timezone := "Africa/Johannesburg"
	if req.Timezone != nil && *req.Timezone != "" {
		timezone = *req.Timezone
	}
	params.Timezone = sql.NullString{String: timezone, Valid: true}

	// Get original schedule for audit logging (before update)
	originalSchedule, originalErr := h.scheduleService.AdminGetScheduleByID(r.Context(), scheduleID)

	schedule, err := h.scheduleService.AdminUpdateSchedule(r.Context(), params)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Schedule not found for update", h.logger, "schedule_id", scheduleID)
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to update schedule", h.logger, "db_params", params, "error", err)
		return
	}

	// Log audit event for schedule update
	if originalErr == nil {
		// Build changes map
		changes := make(map[string]interface{})
		if originalSchedule.Name != schedule.Name {
			changes["name"] = map[string]interface{}{
				"before": originalSchedule.Name,
				"after":  schedule.Name,
			}
		}
		if originalSchedule.CronExpr != schedule.CronExpr {
			changes["cron_expr"] = map[string]interface{}{
				"before": originalSchedule.CronExpr,
				"after":  schedule.CronExpr,
			}
		}
		if originalSchedule.Timezone.String != schedule.Timezone.String {
			changes["timezone"] = map[string]interface{}{
				"before": originalSchedule.Timezone.String,
				"after":  schedule.Timezone.String,
			}
		}
		if originalSchedule.DurationMinutes != schedule.DurationMinutes {
			changes["duration_minutes"] = map[string]interface{}{
				"before": originalSchedule.DurationMinutes,
				"after":  schedule.DurationMinutes,
			}
		}

		WithAuditLogging(r.Context(), h.logger, func(userID int64, ipAddress, userAgent string) error {
			return h.auditService.LogScheduleUpdated(
				r.Context(),
				userID,
				schedule.ScheduleID,
				changes,
				ipAddress,
				userAgent,
			)
		})
	}

	RespondWithJSON(w, http.StatusOK, ToScheduleResponse(schedule), h.logger)
}

// AdminDeleteSchedule handles DELETE /api/admin/schedules/{id}
func (h *AdminScheduleHandlers) AdminDeleteSchedule(w http.ResponseWriter, r *http.Request) {
	// Try multiple methods to extract the ID parameter
	scheduleIDStr := r.PathValue("id")
	h.logger.InfoContext(r.Context(), "AdminDeleteSchedule called", "id_param", scheduleIDStr, "url", r.URL.Path)

	// Alternative method: Parse from URL path directly if r.PathValue fails
	if scheduleIDStr == "" {
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) >= 4 && pathParts[0] == "api" && pathParts[1] == "admin" && pathParts[2] == "schedules" {
			scheduleIDStr = pathParts[3]
			h.logger.InfoContext(r.Context(), "Extracted ID from path manually", "id_param", scheduleIDStr)
		}
	}

	scheduleID, err := strconv.ParseInt(scheduleIDStr, 10, 64)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to parse schedule ID", "id_param", scheduleIDStr, "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid schedule ID format", h.logger, "schedule_id_str", scheduleIDStr, "error", err)
		return
	}

	// Get schedule details before deletion for audit logging
	schedule, scheduleErr := h.scheduleService.AdminGetScheduleByID(r.Context(), scheduleID)

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

	// Log audit event for schedule deletion
	if scheduleErr == nil {
		WithAuditLogging(r.Context(), h.logger, func(userID int64, ipAddress, userAgent string) error {
			return h.auditService.LogScheduleDeleted(
				r.Context(),
				userID,
				scheduleID,
				schedule.Name,
				ipAddress,
				userAgent,
			)
		})
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Schedule deleted successfully"}, h.logger)
}

// AdminListAllShiftSlots handles GET /api/admin/schedules/all-slots
func (h *AdminScheduleHandlers) AdminListAllShiftSlots(w http.ResponseWriter, r *http.Request) {
	// Parse optional query parameters for date range and limit
	var fromTime *time.Time
	var toTime *time.Time
	var limit *int

	fromStr := r.URL.Query().Get("from")
	if fromStr != "" {
		parsedFromTime, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid 'from' date format, use RFC3339 (e.g., YYYY-MM-DDTHH:MM:SSZ)", h.logger, "input_from", fromStr)
			return
		}
		fromTime = &parsedFromTime
	}

	toStr := r.URL.Query().Get("to")
	if toStr != "" {
		parsedToTime, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid 'to' date format, use RFC3339 (e.g., YYYY-MM-DDTHH:MM:SSZ)", h.logger, "input_to", toStr)
			return
		}
		toTime = &parsedToTime
	}

	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit <= 0 {
			RespondWithError(w, http.StatusBadRequest, "Invalid 'limit' parameter, must be a positive integer", h.logger, "input_limit", limitStr)
			return
		}
		limit = &parsedLimit
	}

	// Get all shift slots (booked or not) from the service
	slots, err := h.scheduleService.AdminGetAllShiftSlots(r.Context(), fromTime, toTime, limit)
	if err != nil {
		// The service layer should return specific errors like ErrNotFound or ErrInternalServer
		if errors.Is(err, service.ErrInternalServer) { // Assuming service might return this
			RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve shift slots due to an internal error", h.logger, "error", err)
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve shift slots", h.logger, "error", err)
		}
		return
	}

	RespondWithJSON(w, http.StatusOK, slots, h.logger)
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

	// Log audit event for bulk schedule deletion
	WithAuditLogging(r.Context(), h.logger, func(userID int64, ipAddress, userAgent string) error {
		return h.auditService.LogScheduleBulkDeleted(
			r.Context(),
			userID,
			req.ScheduleIDs,
			len(req.ScheduleIDs),
			ipAddress,
			userAgent,
		)
	})

	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Schedules deleted successfully"}, h.logger)
}
