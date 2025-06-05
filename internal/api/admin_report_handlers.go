package api

import (
	"net/http"
	"time"

	db "night-owls-go/internal/db/sqlc_generated"
	"night-owls-go/internal/service"

	"database/sql"
	"log/slog"
	"strconv"
	"strings"
)

// AdminReportHandler handles admin-specific report operations.
type AdminReportHandler struct {
	reportService   *service.ReportService
	scheduleService *service.ScheduleService
	querier         db.Querier
	auditService    *service.AuditService
	logger          *slog.Logger
}

// NewAdminReportHandler creates a new AdminReportHandler.
func NewAdminReportHandler(reportService *service.ReportService, scheduleService *service.ScheduleService, querier db.Querier, auditService *service.AuditService, logger *slog.Logger) *AdminReportHandler {
	return &AdminReportHandler{
		reportService:   reportService,
		scheduleService: scheduleService,
		querier:         querier,
		auditService:    auditService,
		logger:          logger.With("handler", "AdminReportHandler"),
	}
}

// AdminReportResponse extends the basic ReportResponse with admin context
type AdminReportResponse struct {
	ReportID     int64      `json:"report_id"`
	BookingID    int64      `json:"booking_id"`
	Severity     int64      `json:"severity"`
	Message      string     `json:"message,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	ArchivedAt   *time.Time `json:"archived_at,omitempty"`
	UserID       int64      `json:"user_id"`
	UserName     string     `json:"user_name,omitempty"`
	UserPhone    string     `json:"user_phone"`
	ScheduleID   int64      `json:"schedule_id"`
	ScheduleName string     `json:"schedule_name"`
	ShiftStart   time.Time  `json:"shift_start"`
	ShiftEnd     time.Time  `json:"shift_end"`
	Latitude     *float64   `json:"latitude,omitempty"`
	Longitude    *float64   `json:"longitude,omitempty"`
	GPSAccuracy  *float64   `json:"gps_accuracy,omitempty"`
	GPSTimestamp *time.Time `json:"gps_timestamp,omitempty"`
}

// AdminListReportsHandler handles GET /api/admin/reports
// @Summary List all reports (Admin)
// @Description Get all incident reports with full context including user and schedule information
// @Tags admin/reports
// @Produce json
// @Param from query string false "Start date (YYYY-MM-DD)"
// @Param to query string false "End date (YYYY-MM-DD)"
// @Param severity query int false "Filter by severity (0=info, 1=warning, 2=critical)"
// @Param schedule_id query int false "Filter by schedule ID"
// @Param user_id query int false "Filter by user ID"
// @Success 200 {array} AdminReportResponse "List of reports with full context"
// @Failure 400 {object} ErrorResponse "Invalid query parameters"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/admin/reports [get]
func (h *AdminReportHandler) AdminListReportsHandler(w http.ResponseWriter, r *http.Request) {
	// For now, get all reports. Filtering can be added later with additional SQL queries
	reports, err := h.querier.AdminListReportsWithContext(r.Context())
	if err != nil {
		h.logger.Error("Failed to fetch reports with context", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch reports", h.logger)
		return
	}

	// Convert to API response format
	apiReports := make([]AdminReportResponse, 0, len(reports))
	for _, report := range reports {
		apiReport := AdminReportResponse{
			ReportID:     report.ReportID,
			Severity:     report.Severity,
			UserName:     report.UserName,
			UserPhone:    report.UserPhone,
			ScheduleID:   report.ScheduleID,
			ScheduleName: report.ScheduleName,
		}

		// Handle nullable BookingID field
		if report.BookingID.Valid {
			apiReport.BookingID = report.BookingID.Int64
		}

		// Handle nullable UserID field
		if report.UserID.Valid {
			apiReport.UserID = report.UserID.Int64
		}

		// Handle interface{} ShiftStart field
		if shiftStartStr, ok := report.ShiftStart.(string); ok {
			if parsedTime, err := time.Parse("2006-01-02 15:04:05", shiftStartStr); err == nil {
				apiReport.ShiftStart = parsedTime
			}
		}

		// Handle interface{} ShiftEnd field
		if shiftEndStr, ok := report.ShiftEnd.(string); ok {
			if parsedTime, err := time.Parse("2006-01-02 15:04:05", shiftEndStr); err == nil {
				apiReport.ShiftEnd = parsedTime
			}
		}

		// Handle nullable CreatedAt field
		if report.CreatedAt.Valid {
			apiReport.CreatedAt = report.CreatedAt.Time
		}

		// Handle nullable ArchivedAt field
		if report.ArchivedAt.Valid {
			apiReport.ArchivedAt = &report.ArchivedAt.Time
		}

		// Handle nullable message field
		if report.Message.Valid {
			apiReport.Message = report.Message.String
		}

		// Handle GPS fields
		if report.Latitude.Valid {
			apiReport.Latitude = &report.Latitude.Float64
		}
		if report.Longitude.Valid {
			apiReport.Longitude = &report.Longitude.Float64
		}
		if report.GpsAccuracy.Valid {
			apiReport.GPSAccuracy = &report.GpsAccuracy.Float64
		}
		if report.GpsTimestamp.Valid {
			apiReport.GPSTimestamp = &report.GpsTimestamp.Time
		}

		apiReports = append(apiReports, apiReport)
	}

	RespondWithJSON(w, http.StatusOK, apiReports, h.logger)
}

// AdminGetReportHandler handles GET /api/admin/reports/{id}
// @Summary Get a specific report (Admin)
// @Description Get a specific report with full context by ID
// @Tags admin/reports
// @Produce json
// @Param id path int true "Report ID"
// @Success 200 {object} AdminReportResponse "Report with full context"
// @Failure 400 {object} ErrorResponse "Invalid report ID"
// @Failure 404 {object} ErrorResponse "Report not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/admin/reports/{id} [get]
func (h *AdminReportHandler) AdminGetReportHandler(w http.ResponseWriter, r *http.Request) {
	// Try multiple methods to extract the ID parameter (following users pattern)
	idStr := r.PathValue("id")
	h.logger.InfoContext(r.Context(), "AdminGetReportHandler called", "id_param", idStr, "url", r.URL.Path)

	// Alternative method: Parse from URL path directly if r.PathValue fails
	if idStr == "" {
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) >= 4 && pathParts[0] == "api" && pathParts[1] == "admin" && pathParts[2] == "reports" {
			idStr = pathParts[3]
			h.logger.InfoContext(r.Context(), "Extracted ID from path manually", "id_param", idStr)
		}
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to parse report ID", "id_param", idStr, "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid report ID", h.logger, "error", err)
		return
	}

	report, err := h.querier.AdminGetReportWithContext(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondWithError(w, http.StatusNotFound, "Report not found", h.logger)
		} else {
			h.logger.ErrorContext(r.Context(), "Failed to get report by ID", "report_id", id, "error", err)
			RespondWithError(w, http.StatusInternalServerError, "Failed to fetch report", h.logger)
		}
		return
	}

	// Convert to API response format
	apiReport := AdminReportResponse{
		ReportID:     report.ReportID,
		Severity:     report.Severity,
		UserName:     report.UserName,
		UserPhone:    report.UserPhone,
		ScheduleID:   report.ScheduleID,
		ScheduleName: report.ScheduleName,
	}

	// Handle nullable BookingID field
	if report.BookingID.Valid {
		apiReport.BookingID = report.BookingID.Int64
	}

	// Handle nullable UserID field
	if report.UserID.Valid {
		apiReport.UserID = report.UserID.Int64
	}

	// Handle interface{} ShiftStart field
	if shiftStartStr, ok := report.ShiftStart.(string); ok {
		if parsedTime, err := time.Parse("2006-01-02 15:04:05", shiftStartStr); err == nil {
			apiReport.ShiftStart = parsedTime
		}
	}

	// Handle interface{} ShiftEnd field
	if shiftEndStr, ok := report.ShiftEnd.(string); ok {
		if parsedTime, err := time.Parse("2006-01-02 15:04:05", shiftEndStr); err == nil {
			apiReport.ShiftEnd = parsedTime
		}
	}

	// Handle nullable CreatedAt field
	if report.CreatedAt.Valid {
		apiReport.CreatedAt = report.CreatedAt.Time
	}

	// Handle nullable ArchivedAt field
	if report.ArchivedAt.Valid {
		apiReport.ArchivedAt = &report.ArchivedAt.Time
	}

	// Handle nullable message field
	if report.Message.Valid {
		apiReport.Message = report.Message.String
	}

	// Handle GPS fields
	if report.Latitude.Valid {
		apiReport.Latitude = &report.Latitude.Float64
	}
	if report.Longitude.Valid {
		apiReport.Longitude = &report.Longitude.Float64
	}
	if report.GpsAccuracy.Valid {
		apiReport.GPSAccuracy = &report.GpsAccuracy.Float64
	}
	if report.GpsTimestamp.Valid {
		apiReport.GPSTimestamp = &report.GpsTimestamp.Time
	}

	RespondWithJSON(w, http.StatusOK, apiReport, h.logger)
}

// AdminArchiveReportHandler handles PUT /api/admin/reports/{id}/archive
// @Summary Archive a report (Admin)
// @Description Archive a specific report by ID (soft delete)
// @Tags admin/reports
// @Param id path int true "Report ID"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} ErrorResponse "Invalid report ID"
// @Failure 404 {object} ErrorResponse "Report not found or already archived"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/admin/reports/{id}/archive [put]
func (h *AdminReportHandler) AdminArchiveReportHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) >= 4 && pathParts[0] == "api" && pathParts[1] == "admin" && pathParts[2] == "reports" {
			idStr = pathParts[3]
		}
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid report ID", h.logger)
		return
	}

	err = h.querier.ArchiveReport(r.Context(), id)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to archive report", "report_id", id, "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to archive report", h.logger)
		return
	}

	// Get report details for audit logging
	report, reportErr := h.querier.AdminGetReportWithContext(r.Context(), id)
	if reportErr == nil {
		// Get user ID from auth context for audit logging
		userIDFromAuth, ok := r.Context().Value(UserIDKey).(int64)
		if ok {
			ipAddress, userAgent := GetAuditInfoFromContext(r.Context())

			var reporterUserID *int64
			if report.UserID.Valid {
				reporterUserID = &report.UserID.Int64
			}

			auditErr := h.auditService.LogReportArchived(
				r.Context(),
				userIDFromAuth,
				id,
				reporterUserID,
				report.Severity,
				ipAddress,
				userAgent,
			)
			if auditErr != nil {
				h.logger.WarnContext(r.Context(), "Failed to log report archive audit event", "report_id", id, "error", auditErr)
			}
		}
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Report archived successfully"}, h.logger)
}

// AdminUnarchiveReportHandler handles PUT /api/admin/reports/{id}/unarchive
// @Summary Unarchive a report (Admin)
// @Description Unarchive a specific report by ID
// @Tags admin/reports
// @Param id path int true "Report ID"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} ErrorResponse "Invalid report ID"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/admin/reports/{id}/unarchive [put]
func (h *AdminReportHandler) AdminUnarchiveReportHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) >= 4 && pathParts[0] == "api" && pathParts[1] == "admin" && pathParts[2] == "reports" {
			idStr = pathParts[3]
		}
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid report ID", h.logger)
		return
	}

	err = h.querier.UnarchiveReport(r.Context(), id)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to unarchive report", "report_id", id, "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to unarchive report", h.logger)
		return
	}

	// Get report details for audit logging
	report, reportErr := h.querier.AdminGetReportWithContext(r.Context(), id)
	if reportErr == nil {
		// Get user ID from auth context for audit logging
		userIDFromAuth, ok := r.Context().Value(UserIDKey).(int64)
		if ok {
			ipAddress, userAgent := GetAuditInfoFromContext(r.Context())

			var reporterUserID *int64
			if report.UserID.Valid {
				reporterUserID = &report.UserID.Int64
			}

			auditErr := h.auditService.LogReportUnarchived(
				r.Context(),
				userIDFromAuth,
				id,
				reporterUserID,
				report.Severity,
				ipAddress,
				userAgent,
			)
			if auditErr != nil {
				h.logger.WarnContext(r.Context(), "Failed to log report unarchive audit event", "report_id", id, "error", auditErr)
			}
		}
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Report unarchived successfully"}, h.logger)
}

// AdminListArchivedReportsHandler handles GET /api/admin/reports/archived
// @Summary List archived reports (Admin)
// @Description Get all archived reports with full context
// @Tags admin/reports
// @Produce json
// @Success 200 {array} AdminReportResponse "List of archived reports"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/admin/reports/archived [get]
func (h *AdminReportHandler) AdminListArchivedReportsHandler(w http.ResponseWriter, r *http.Request) {
	reports, err := h.querier.AdminListArchivedReportsWithContext(r.Context())
	if err != nil {
		h.logger.Error("Failed to fetch archived reports", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch archived reports", h.logger)
		return
	}

	// Convert to API response format
	apiReports := make([]AdminReportResponse, 0, len(reports))
	for _, report := range reports {
		apiReport := AdminReportResponse{
			ReportID:     report.ReportID,
			Severity:     report.Severity,
			UserName:     report.UserName,
			UserPhone:    report.UserPhone,
			ScheduleID:   report.ScheduleID,
			ScheduleName: report.ScheduleName,
		}

		// Handle nullable BookingID field
		if report.BookingID.Valid {
			apiReport.BookingID = report.BookingID.Int64
		}

		// Handle nullable UserID field
		if report.UserID.Valid {
			apiReport.UserID = report.UserID.Int64
		}

		// Handle interface{} ShiftStart field
		if shiftStartStr, ok := report.ShiftStart.(string); ok {
			if parsedTime, err := time.Parse("2006-01-02 15:04:05", shiftStartStr); err == nil {
				apiReport.ShiftStart = parsedTime
			}
		}

		// Handle interface{} ShiftEnd field
		if shiftEndStr, ok := report.ShiftEnd.(string); ok {
			if parsedTime, err := time.Parse("2006-01-02 15:04:05", shiftEndStr); err == nil {
				apiReport.ShiftEnd = parsedTime
			}
		}

		// Handle nullable CreatedAt field
		if report.CreatedAt.Valid {
			apiReport.CreatedAt = report.CreatedAt.Time
		}

		// Handle nullable ArchivedAt field
		if report.ArchivedAt.Valid {
			apiReport.ArchivedAt = &report.ArchivedAt.Time
		}

		// Handle nullable message field
		if report.Message.Valid {
			apiReport.Message = report.Message.String
		}

		// Handle GPS fields
		if report.Latitude.Valid {
			apiReport.Latitude = &report.Latitude.Float64
		}
		if report.Longitude.Valid {
			apiReport.Longitude = &report.Longitude.Float64
		}
		if report.GpsAccuracy.Valid {
			apiReport.GPSAccuracy = &report.GpsAccuracy.Float64
		}
		if report.GpsTimestamp.Valid {
			apiReport.GPSTimestamp = &report.GpsTimestamp.Time
		}

		apiReports = append(apiReports, apiReport)
	}

	RespondWithJSON(w, http.StatusOK, apiReports, h.logger)
}

// AdminDeleteReportHandler handles DELETE /api/admin/reports/{id}
// @Summary Delete a report (Admin)
// @Description Permanently delete a specific report by ID
// @Tags admin/reports
// @Param id path int true "Report ID"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} ErrorResponse "Invalid report ID"
// @Failure 404 {object} ErrorResponse "Report not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/admin/reports/{id} [delete]
func (h *AdminReportHandler) AdminDeleteReportHandler(w http.ResponseWriter, r *http.Request) {
	// Try multiple methods to extract the ID parameter
	idStr := r.PathValue("id")
	h.logger.InfoContext(r.Context(), "AdminDeleteReportHandler called", "id_param", idStr, "url", r.URL.Path)

	// Alternative method: Parse from URL path directly if r.PathValue fails
	if idStr == "" {
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) >= 4 && pathParts[0] == "api" && pathParts[1] == "admin" && pathParts[2] == "reports" {
			idStr = pathParts[3]
			h.logger.InfoContext(r.Context(), "Extracted ID from path manually", "id_param", idStr)
		}
	}



	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to parse report ID", "id_param", idStr, "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid report ID", h.logger, "error", err)
		return
	}

	// Check if report exists before deleting
	report, err := h.querier.AdminGetReportWithContext(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondWithError(w, http.StatusNotFound, "Report not found", h.logger)
		} else {
			h.logger.ErrorContext(r.Context(), "Failed to check report existence", "report_id", id, "error", err)
			RespondWithError(w, http.StatusInternalServerError, "Failed to delete report", h.logger)
		}
		return
	}

	err = h.querier.DeleteReport(r.Context(), id)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to delete report", "report_id", id, "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete report", h.logger)
		return
	}

	// Log audit event for report deletion
	userIDFromAuth, ok := r.Context().Value(UserIDKey).(int64)
	if ok {
		ipAddress, userAgent := GetAuditInfoFromContext(r.Context())

		var reporterUserID *int64
		if report.UserID.Valid {
			reporterUserID = &report.UserID.Int64
		}

		auditErr := h.auditService.LogReportDeleted(
			r.Context(),
			userIDFromAuth,
			id,
			reporterUserID,
			report.Severity,
			ipAddress,
			userAgent,
		)
		if auditErr != nil {
			h.logger.WarnContext(r.Context(), "Failed to log report deletion audit event", "report_id", id, "error", auditErr)
		}
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Report deleted successfully"}, h.logger)
}
