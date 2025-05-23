package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	db "night-owls-go/internal/db/sqlc_generated"
	"night-owls-go/internal/service"

	"github.com/go-chi/chi/v5"
)

// AdminRecurringAssignmentHandlers holds handlers for admin recurring assignment operations.
type AdminRecurringAssignmentHandlers struct {
	logger                     *slog.Logger
	recurringAssignmentService *service.RecurringAssignmentService
}

// NewAdminRecurringAssignmentHandlers creates a new AdminRecurringAssignmentHandlers.
func NewAdminRecurringAssignmentHandlers(logger *slog.Logger, recurringAssignmentService *service.RecurringAssignmentService) *AdminRecurringAssignmentHandlers {
	return &AdminRecurringAssignmentHandlers{
		logger:                     logger.With("handler", "AdminRecurringAssignmentHandlers"),
		recurringAssignmentService: recurringAssignmentService,
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
	assignments, err := h.recurringAssignmentService.ListRecurringAssignments(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to list recurring assignments", h.logger, "error", err)
		return
	}

	RespondWithJSON(w, http.StatusOK, assignments, h.logger)
}

// AdminGetRecurringAssignment handles GET /api/admin/recurring-assignments/{id}
func (h *AdminRecurringAssignmentHandlers) AdminGetRecurringAssignment(w http.ResponseWriter, r *http.Request) {
	assignmentIDStr := chi.URLParam(r, "id")
	assignmentID, err := strconv.ParseInt(assignmentIDStr, 10, 64)
	if err != nil {
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
	assignmentIDStr := chi.URLParam(r, "id")
	assignmentID, err := strconv.ParseInt(assignmentIDStr, 10, 64)
	if err != nil {
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
	assignmentIDStr := chi.URLParam(r, "id")
	assignmentID, err := strconv.ParseInt(assignmentIDStr, 10, 64)
	if err != nil {
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