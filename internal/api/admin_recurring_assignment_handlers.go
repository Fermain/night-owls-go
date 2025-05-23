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

	"github.com/go-chi/chi/v5"
)

// AdminRecurringAssignmentHandlers holds handlers for admin recurring assignment operations.
type AdminRecurringAssignmentHandlers struct {
	logger                     *slog.Logger
	recurringAssignmentService *service.RecurringAssignmentService
	scheduleService            *service.ScheduleService
}

// NewAdminRecurringAssignmentHandlers creates a new AdminRecurringAssignmentHandlers.
func NewAdminRecurringAssignmentHandlers(logger *slog.Logger, recurringAssignmentService *service.RecurringAssignmentService, scheduleService *service.ScheduleService) *AdminRecurringAssignmentHandlers {
	return &AdminRecurringAssignmentHandlers{
		logger:                     logger.With("handler", "AdminRecurringAssignmentHandlers"),
		recurringAssignmentService: recurringAssignmentService,
		scheduleService:            scheduleService,
	}
}

// AdminCreateRecurringAssignmentRequest defines the expected JSON body for creating a recurring assignment.
type AdminCreateRecurringAssignmentRequest struct {
	UserID      int64   `json:"user_id"`
	BuddyName   *string `json:"buddy_name,omitempty"`
	DayOfWeek   int64   `json:"day_of_week"`
	ScheduleID  int64   `json:"schedule_id"`
	TimeSlot    string  `json:"time_slot"`
	Description *string `json:"description,omitempty"`
}

// AdminCreateRecurringAssignment handles POST /api/admin/recurring-assignments
func (h *AdminRecurringAssignmentHandlers) AdminCreateRecurringAssignment(w http.ResponseWriter, r *http.Request) {
	var req AdminCreateRecurringAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WarnContext(r.Context(), "Failed to decode request body", "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload", h.logger)
		return
	}

	if req.UserID == 0 || req.DayOfWeek < 0 || req.DayOfWeek > 6 || req.ScheduleID == 0 || req.TimeSlot == "" {
		RespondWithError(w, http.StatusBadRequest, "Missing or invalid required fields", h.logger)
		return
	}

	params := db.CreateRecurringAssignmentParams{
		UserID:     req.UserID,
		DayOfWeek:  req.DayOfWeek,
		ScheduleID: req.ScheduleID,
		TimeSlot:   req.TimeSlot,
	}

	if req.BuddyName != nil {
		params.BuddyName = sql.NullString{String: *req.BuddyName, Valid: *req.BuddyName != ""}
	}

	if req.Description != nil {
		params.Description = sql.NullString{String: *req.Description, Valid: *req.Description != ""}
	}

	assignment, err := h.recurringAssignmentService.CreateRecurringAssignment(r.Context(), params)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			RespondWithError(w, http.StatusBadRequest, "User or schedule not found", h.logger)
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to create recurring assignment", h.logger, "error", err)
		return
	}

	RespondWithJSON(w, http.StatusCreated, assignment, h.logger)
}

// AdminListRecurringAssignments handles GET /api/admin/recurring-assignments
func (h *AdminRecurringAssignmentHandlers) AdminListRecurringAssignments(w http.ResponseWriter, r *http.Request) {
	// Check if this is a materialize request
	if r.URL.Query().Get("materialize") == "true" {
		h.handleMaterializeRequest(w, r)
		return
	}

	assignments, err := h.recurringAssignmentService.ListRecurringAssignments(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to list recurring assignments", h.logger, "error", err)
		return
	}

	RespondWithJSON(w, http.StatusOK, assignments, h.logger)
}

// handleMaterializeRequest handles the materialization logic
func (h *AdminRecurringAssignmentHandlers) handleMaterializeRequest(w http.ResponseWriter, r *http.Request) {
	// Parse optional query parameters for the time window
	now := time.Now().UTC()
	fromTime := now
	toTime := now.AddDate(0, 0, 14) // Default to next 2 weeks

	fromStr := r.URL.Query().Get("from")
	if fromStr != "" {
		parsedFromTime, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid 'from' date format, use RFC3339", h.logger, "input_from", fromStr)
			return
		}
		fromTime = parsedFromTime.UTC()
	}

	toStr := r.URL.Query().Get("to")
	if toStr != "" {
		parsedToTime, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid 'to' date format, use RFC3339", h.logger, "input_to", toStr)
			return
		}
		toTime = parsedToTime.UTC()
	}

	// Trigger materialization
	err := h.recurringAssignmentService.MaterializeUpcomingBookings(r.Context(), h.scheduleService, fromTime, toTime)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to materialize bookings", h.logger, "error", err)
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Bookings materialized successfully",
		"from":    fromTime,
		"to":      toTime,
	}, h.logger)
}

// AdminGetRecurringAssignment handles GET /api/admin/recurring-assignments/{id}
func (h *AdminRecurringAssignmentHandlers) AdminGetRecurringAssignment(w http.ResponseWriter, r *http.Request) {
	// Try multiple methods to extract the ID parameter
	assignmentIDStr := chi.URLParam(r, "id")
	h.logger.InfoContext(r.Context(), "AdminGetRecurringAssignment called", "id_param", assignmentIDStr, "url", r.URL.Path)
	
	// Alternative method: Parse from URL path directly if chi.URLParam fails
	if assignmentIDStr == "" {
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) >= 4 && pathParts[0] == "api" && pathParts[1] == "admin" && pathParts[2] == "recurring-assignments" {
			assignmentIDStr = pathParts[3]
			h.logger.InfoContext(r.Context(), "Extracted ID from path manually", "id_param", assignmentIDStr)
		}
	}
	
	// Alternative method 2: Check request context for route values
	if assignmentIDStr == "" {
		if rctx := chi.RouteContext(r.Context()); rctx != nil {
			for i, param := range rctx.URLParams.Keys {
				if param == "id" && i < len(rctx.URLParams.Values) {
					assignmentIDStr = rctx.URLParams.Values[i]
					h.logger.InfoContext(r.Context(), "Found ID in route context", "id_param", assignmentIDStr)
					break
				}
			}
		}
	}
	
	assignmentID, err := strconv.ParseInt(assignmentIDStr, 10, 64)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to parse assignment ID", "id_param", assignmentIDStr, "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid assignment ID format", h.logger, "assignment_id_str", assignmentIDStr, "error", err)
		return
	}

	assignment, err := h.recurringAssignmentService.GetRecurringAssignmentByID(r.Context(), assignmentID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Recurring assignment not found", h.logger, "assignment_id", assignmentID)
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to get recurring assignment", h.logger, "assignment_id", assignmentID, "error", err)
		return
	}

	RespondWithJSON(w, http.StatusOK, assignment, h.logger)
}

// AdminUpdateRecurringAssignmentRequest reuses AdminCreateRecurringAssignmentRequest for simplicity.
type AdminUpdateRecurringAssignmentRequest AdminCreateRecurringAssignmentRequest

// AdminUpdateRecurringAssignment handles PUT /api/admin/recurring-assignments/{id}
func (h *AdminRecurringAssignmentHandlers) AdminUpdateRecurringAssignment(w http.ResponseWriter, r *http.Request) {
	// Try multiple methods to extract the ID parameter
	assignmentIDStr := chi.URLParam(r, "id")
	h.logger.InfoContext(r.Context(), "AdminUpdateRecurringAssignment called", "id_param", assignmentIDStr, "url", r.URL.Path)
	
	// Alternative method: Parse from URL path directly if chi.URLParam fails
	if assignmentIDStr == "" {
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) >= 4 && pathParts[0] == "api" && pathParts[1] == "admin" && pathParts[2] == "recurring-assignments" {
			assignmentIDStr = pathParts[3]
			h.logger.InfoContext(r.Context(), "Extracted ID from path manually", "id_param", assignmentIDStr)
		}
	}
	
	// Alternative method 2: Check request context for route values
	if assignmentIDStr == "" {
		if rctx := chi.RouteContext(r.Context()); rctx != nil {
			for i, param := range rctx.URLParams.Keys {
				if param == "id" && i < len(rctx.URLParams.Values) {
					assignmentIDStr = rctx.URLParams.Values[i]
					h.logger.InfoContext(r.Context(), "Found ID in route context", "id_param", assignmentIDStr)
					break
				}
			}
		}
	}
	
	assignmentID, err := strconv.ParseInt(assignmentIDStr, 10, 64)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to parse assignment ID", "id_param", assignmentIDStr, "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid assignment ID format", h.logger, "assignment_id_str", assignmentIDStr, "error", err)
		return
	}

	var req AdminUpdateRecurringAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WarnContext(r.Context(), "Failed to decode request body for update", "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload", h.logger)
		return
	}

	if req.UserID == 0 || req.DayOfWeek < 0 || req.DayOfWeek > 6 || req.ScheduleID == 0 || req.TimeSlot == "" {
		RespondWithError(w, http.StatusBadRequest, "Missing or invalid required fields", h.logger)
		return
	}

	params := db.UpdateRecurringAssignmentParams{
		RecurringAssignmentID: assignmentID,
		UserID:                req.UserID,
		DayOfWeek:             req.DayOfWeek,
		ScheduleID:            req.ScheduleID,
		TimeSlot:              req.TimeSlot,
	}

	if req.BuddyName != nil {
		params.BuddyName = sql.NullString{String: *req.BuddyName, Valid: *req.BuddyName != ""}
	}

	if req.Description != nil {
		params.Description = sql.NullString{String: *req.Description, Valid: *req.Description != ""}
	}

	assignment, err := h.recurringAssignmentService.UpdateRecurringAssignment(r.Context(), params)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Recurring assignment, user, or schedule not found", h.logger, "assignment_id", assignmentID)
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to update recurring assignment", h.logger, "assignment_id", assignmentID, "error", err)
		return
	}

	RespondWithJSON(w, http.StatusOK, assignment, h.logger)
}

// AdminDeleteRecurringAssignment handles DELETE /api/admin/recurring-assignments/{id}
func (h *AdminRecurringAssignmentHandlers) AdminDeleteRecurringAssignment(w http.ResponseWriter, r *http.Request) {
	// Try multiple methods to extract the ID parameter
	assignmentIDStr := chi.URLParam(r, "id")
	h.logger.InfoContext(r.Context(), "AdminDeleteRecurringAssignment called", "id_param", assignmentIDStr, "url", r.URL.Path)
	
	// Alternative method: Parse from URL path directly if chi.URLParam fails
	if assignmentIDStr == "" {
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) >= 4 && pathParts[0] == "api" && pathParts[1] == "admin" && pathParts[2] == "recurring-assignments" {
			assignmentIDStr = pathParts[3]
			h.logger.InfoContext(r.Context(), "Extracted ID from path manually", "id_param", assignmentIDStr)
		}
	}
	
	// Alternative method 2: Check request context for route values
	if assignmentIDStr == "" {
		if rctx := chi.RouteContext(r.Context()); rctx != nil {
			for i, param := range rctx.URLParams.Keys {
				if param == "id" && i < len(rctx.URLParams.Values) {
					assignmentIDStr = rctx.URLParams.Values[i]
					h.logger.InfoContext(r.Context(), "Found ID in route context", "id_param", assignmentIDStr)
					break
				}
			}
		}
	}
	
	assignmentID, err := strconv.ParseInt(assignmentIDStr, 10, 64)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to parse assignment ID", "id_param", assignmentIDStr, "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid assignment ID format", h.logger, "assignment_id_str", assignmentIDStr, "error", err)
		return
	}

	err = h.recurringAssignmentService.DeleteRecurringAssignment(r.Context(), assignmentID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete recurring assignment", h.logger, "assignment_id", assignmentID, "error", err)
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Recurring assignment deleted successfully"}, h.logger)
}

// AdminMaterializeBookings handles POST /api/admin/recurring-assignments/materialize
// This endpoint manually triggers the materialization of bookings from recurring assignments
func (h *AdminRecurringAssignmentHandlers) AdminMaterializeBookings(w http.ResponseWriter, r *http.Request, scheduleService *service.ScheduleService) {
	// Parse optional query parameters for the time window
	now := time.Now().UTC()
	fromTime := now
	toTime := now.AddDate(0, 0, 14) // Default to next 2 weeks

	fromStr := r.URL.Query().Get("from")
	if fromStr != "" {
		parsedFromTime, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid 'from' date format, use RFC3339", h.logger, "input_from", fromStr)
			return
		}
		fromTime = parsedFromTime.UTC()
	}

	toStr := r.URL.Query().Get("to")
	if toStr != "" {
		parsedToTime, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid 'to' date format, use RFC3339", h.logger, "input_to", toStr)
			return
		}
		toTime = parsedToTime.UTC()
	}

	// Trigger materialization
	err := h.recurringAssignmentService.MaterializeUpcomingBookings(r.Context(), scheduleService, fromTime, toTime)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to materialize bookings", h.logger, "error", err)
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Bookings materialized successfully",
		"from":    fromTime,
		"to":      toTime,
	}, h.logger)
}

// AdminMaterializeBookingsHandler returns a handler function for materializing bookings
func (h *AdminRecurringAssignmentHandlers) AdminMaterializeBookingsHandler(scheduleService *service.ScheduleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.AdminMaterializeBookings(w, r, scheduleService)
	}
} 