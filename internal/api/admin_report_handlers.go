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

	"github.com/go-chi/chi/v5"
)

// AdminReportHandler handles admin-specific report operations.
type AdminReportHandler struct {
	reportService   *service.ReportService
	scheduleService *service.ScheduleService
	querier         db.Querier
	logger          *slog.Logger
}

// NewAdminReportHandler creates a new AdminReportHandler.
func NewAdminReportHandler(reportService *service.ReportService, scheduleService *service.ScheduleService, querier db.Querier, logger *slog.Logger) *AdminReportHandler {
	return &AdminReportHandler{
		reportService:   reportService,
		scheduleService: scheduleService,
		querier:         querier,
		logger:          logger.With("handler", "AdminReportHandler"),
	}
}

// AdminReportResponse extends the basic ReportResponse with admin context
type AdminReportResponse struct {
	ReportID     int64     `json:"report_id"`
	BookingID    int64     `json:"booking_id"`
	Severity     int64     `json:"severity"`
	Message      string    `json:"message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UserID       int64     `json:"user_id"`
	UserName     string    `json:"user_name,omitempty"`
	UserPhone    string    `json:"user_phone"`
	ScheduleID   int64     `json:"schedule_id"`
	ScheduleName string    `json:"schedule_name"`
	ShiftStart   time.Time `json:"shift_start"`
	ShiftEnd     time.Time `json:"shift_end"`
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
			BookingID:    report.BookingID,
			Severity:     report.Severity,
			UserID:       report.UserID,
			UserName:     report.UserName,
			UserPhone:    report.UserPhone,
			ScheduleID:   report.ScheduleID,
			ScheduleName: report.ScheduleName,
			ShiftStart:   report.ShiftStart,
			ShiftEnd:     report.ShiftEnd,
		}

		// Handle nullable CreatedAt field
		if report.CreatedAt.Valid {
			apiReport.CreatedAt = report.CreatedAt.Time
		}

		// Handle nullable message field
		if report.Message.Valid {
			apiReport.Message = report.Message.String
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
	idStr := chi.URLParam(r, "id")
	h.logger.InfoContext(r.Context(), "AdminGetReportHandler called", "id_param", idStr, "url", r.URL.Path)
	
	// Alternative method: Parse from URL path directly if chi.URLParam fails
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
		BookingID:    report.BookingID,
		Severity:     report.Severity,
		UserID:       report.UserID,
		UserName:     report.UserName,
		UserPhone:    report.UserPhone,
		ScheduleID:   report.ScheduleID,
		ScheduleName: report.ScheduleName,
		ShiftStart:   report.ShiftStart,
		ShiftEnd:     report.ShiftEnd,
	}

	// Handle nullable CreatedAt field
	if report.CreatedAt.Valid {
		apiReport.CreatedAt = report.CreatedAt.Time
	}

	// Handle nullable message field
	if report.Message.Valid {
		apiReport.Message = report.Message.String
	}

	RespondWithJSON(w, http.StatusOK, apiReport, h.logger)
} 