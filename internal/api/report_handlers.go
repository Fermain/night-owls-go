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

	"night-owls-go/internal/service"
)

// ReportHandler handles report-related HTTP requests.
type ReportHandler struct {
	reportService *service.ReportService
	auditService  *service.AuditService
	logger        *slog.Logger
}

// NewReportHandler creates a new ReportHandler.
func NewReportHandler(reportService *service.ReportService, auditService *service.AuditService, logger *slog.Logger) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
		auditService:  auditService,
		logger:        logger.With("handler", "ReportHandler"),
	}
}

// CreateReportRequest is the expected JSON for POST /bookings/{id}/report.
type CreateReportRequest struct {
	Severity          int32    `json:"severity"` // 0, 1, or 2
	Message           string   `json:"message,omitempty"`
	Latitude          *float64 `json:"latitude,omitempty"`
	Longitude         *float64 `json:"longitude,omitempty"`
	Accuracy          *float64 `json:"accuracy,omitempty"`
	LocationTimestamp *string  `json:"location_timestamp,omitempty"`
}

// CreateOffShiftReportRequest is the expected JSON for POST /reports/off-shift.
type CreateOffShiftReportRequest struct {
	Severity          int32    `json:"severity"` // 0, 1, or 2
	Message           string   `json:"message,omitempty"`
	Latitude          *float64 `json:"latitude,omitempty"`
	Longitude         *float64 `json:"longitude,omitempty"`
	Accuracy          *float64 `json:"accuracy,omitempty"`
	LocationTimestamp *string  `json:"location_timestamp,omitempty"`
}

// UserReportResponse is the user-facing report response (fewer fields than admin)
type UserReportResponse struct {
	ReportID     int64      `json:"report_id"`
	BookingID    *int64     `json:"booking_id,omitempty"`
	Severity     int64      `json:"severity"`
	Message      string     `json:"message,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	Latitude     *float64   `json:"latitude,omitempty"`
	Longitude    *float64   `json:"longitude,omitempty"`
	GPSAccuracy  *float64   `json:"gps_accuracy,omitempty"`
	GPSTimestamp *time.Time `json:"gps_timestamp,omitempty"`
	ScheduleName *string    `json:"schedule_name,omitempty"`
	ShiftStart   *time.Time `json:"shift_start,omitempty"`
	ShiftEnd     *time.Time `json:"shift_end,omitempty"`
}

// CreateReportHandler handles POST /bookings/{id}/report
// @Summary Create a report for a booking
// @Description Submits an incident report for a specific booking
// @Tags reports
// @Accept json
// @Produce json
// @Param id path int true "Booking ID"
// @Param request body CreateReportRequest true "Report details"
// @Success 201 {object} ReportResponse "Report created successfully"
// @Failure 400 {object} ErrorResponse "Invalid request format or severity out of range"
// @Failure 401 {object} ErrorResponse "Unauthorized - authentication required"
// @Failure 403 {object} ErrorResponse "Forbidden - not authorized to report on this booking"
// @Failure 404 {object} ErrorResponse "Booking not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /bookings/{id}/report [post]
func (h *ReportHandler) CreateReportHandler(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value(UserIDKey)
	userID, ok := userIDVal.(int64)
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User ID not found in context or invalid type", h.logger)
		return
	}

	// Try multiple methods to extract the ID parameter
	bookingIDStr := r.PathValue("id")
	h.logger.InfoContext(r.Context(), "CreateReportHandler called", "id_param", bookingIDStr, "url", r.URL.Path)

	// Alternative method: Parse from URL path directly if r.PathValue fails
	if bookingIDStr == "" {
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) >= 3 && pathParts[0] == "bookings" && pathParts[2] == "report" {
			bookingIDStr = pathParts[1]
			h.logger.InfoContext(r.Context(), "Extracted ID from path manually", "id_param", bookingIDStr)
		}
	}



	bookingID, err := strconv.ParseInt(bookingIDStr, 10, 64)
	if err != nil || bookingID <= 0 {
		h.logger.ErrorContext(r.Context(), "Failed to parse booking ID", "id_param", bookingIDStr, "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid booking ID in path", h.logger, "booking_id_str", bookingIDStr)
		return
	}

	var req CreateReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload", h.logger, "error", err.Error())
		return
	}
	defer r.Body.Close()

	// Create a NullString for message
	var messageSQL sql.NullString
	if strings.TrimSpace(req.Message) != "" {
		messageSQL.String = strings.TrimSpace(req.Message)
		messageSQL.Valid = true
	}

	// Parse GPS location data
	var gpsLocation *service.GPSLocation
	if req.Latitude != nil && req.Longitude != nil {
		gpsLocation = &service.GPSLocation{
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
			Accuracy:  req.Accuracy,
		}

		// Parse timestamp if provided
		if req.LocationTimestamp != nil {
			if parsedTime, err := time.Parse(time.RFC3339, *req.LocationTimestamp); err == nil {
				gpsLocation.Timestamp = &parsedTime
			}
		}
	}

	report, err := h.reportService.CreateReport(r.Context(), userID, bookingID, req.Severity, messageSQL.String, gpsLocation)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrReportBookingAuth):
			// Includes both not found and not authorized (combined by design)
			RespondWithError(w, http.StatusForbidden, "Not authorized to create report for this booking", h.logger, "booking_id", bookingID)
		case errors.Is(err, service.ErrSeverityOutOfRange):
			RespondWithError(w, http.StatusBadRequest, "Severity must be 0, 1, or 2", h.logger, "severity", req.Severity)
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to create report", h.logger, "error", err.Error())
		}
		return
	}

	// Log audit event for report creation
	ipAddress, userAgent := GetAuditInfoFromContext(r.Context())

	hasLocation := gpsLocation != nil
	auditErr := h.auditService.LogReportCreated(
		r.Context(),
		userID,
		report.ReportID,
		&bookingID,
		int64(req.Severity),
		hasLocation,
		ipAddress,
		userAgent,
	)
	if auditErr != nil {
		h.logger.WarnContext(r.Context(), "Failed to log report creation audit event", "report_id", report.ReportID, "error", auditErr)
	}

	// Convert to API response format
	reportResponse := ToReportResponse(report)
	RespondWithJSON(w, http.StatusCreated, reportResponse, h.logger)
}

// ListReportsHandler handles GET /api/user/reports
// @Summary Get current user's reports
// @Description Get all reports submitted by the current authenticated user
// @Tags reports
// @Produce json
// @Success 200 {array} UserReportResponse "List of user's reports"
// @Failure 401 {object} ErrorResponse "Unauthorized - authentication required"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/user/reports [get]
func (h *ReportHandler) ListReportsHandler(w http.ResponseWriter, r *http.Request) {
	userIDFromAuthVal := r.Context().Value(UserIDKey)
	userIDFromAuth, ok := userIDFromAuthVal.(int64)
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User ID not found in context", h.logger)
		return
	}

	reports, err := h.reportService.ListReportsByUser(r.Context(), userIDFromAuth)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to list reports for user", "user_id", userIDFromAuth, "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch reports", h.logger)
		return
	}

	// Convert to user-facing API response format
	userReports := make([]UserReportResponse, 0, len(reports))
	for _, report := range reports {
		userReport := UserReportResponse{
			ReportID:  report.ReportID,
			Severity:  report.Severity,
			CreatedAt: report.CreatedAt.Time,
		}

		// Handle nullable BookingID
		if report.BookingID.Valid {
			userReport.BookingID = &report.BookingID.Int64
		}

		// Handle nullable Message
		if report.Message.Valid {
			userReport.Message = report.Message.String
		}

		// Handle GPS fields
		if report.Latitude.Valid {
			userReport.Latitude = &report.Latitude.Float64
		}
		if report.Longitude.Valid {
			userReport.Longitude = &report.Longitude.Float64
		}
		if report.GpsAccuracy.Valid {
			userReport.GPSAccuracy = &report.GpsAccuracy.Float64
		}
		if report.GpsTimestamp.Valid {
			userReport.GPSTimestamp = &report.GpsTimestamp.Time
		}

		// Get schedule info if this is a shift report
		if report.BookingID.Valid {
			booking, err := h.reportService.GetBookingDetails(r.Context(), report.BookingID.Int64)
			if err == nil && booking != nil {
				userReport.ScheduleName = &booking.ScheduleName
				userReport.ShiftStart = &booking.ShiftStart
				userReport.ShiftEnd = &booking.ShiftEnd
			}
		}

		userReports = append(userReports, userReport)
	}

	RespondWithJSON(w, http.StatusOK, userReports, h.logger)
}

// CreateOffShiftReportHandler handles POST /reports/off-shift
// @Summary Create an off-shift report
// @Description Submits an incident report when not on a scheduled shift
// @Tags reports
// @Accept json
// @Produce json
// @Param request body CreateOffShiftReportRequest true "Report details"
// @Success 201 {object} ReportResponse "Report created successfully"
// @Failure 400 {object} ErrorResponse "Invalid request format or severity out of range"
// @Failure 401 {object} ErrorResponse "Unauthorized - authentication required"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /reports/off-shift [post]
func (h *ReportHandler) CreateOffShiftReportHandler(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value(UserIDKey)
	userID, ok := userIDVal.(int64)
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User ID not found in context or invalid type", h.logger)
		return
	}

	var req CreateOffShiftReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload", h.logger, "error", err.Error())
		return
	}
	defer r.Body.Close()

	// Create a NullString for message
	var messageSQL sql.NullString
	if strings.TrimSpace(req.Message) != "" {
		messageSQL.String = strings.TrimSpace(req.Message)
		messageSQL.Valid = true
	}

	// Parse GPS location data
	var gpsLocation *service.GPSLocation
	if req.Latitude != nil && req.Longitude != nil {
		gpsLocation = &service.GPSLocation{
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
			Accuracy:  req.Accuracy,
		}

		// Parse timestamp if provided
		if req.LocationTimestamp != nil {
			if parsedTime, err := time.Parse(time.RFC3339, *req.LocationTimestamp); err == nil {
				gpsLocation.Timestamp = &parsedTime
			}
		}
	}

	report, err := h.reportService.CreateOffShiftReport(r.Context(), userID, req.Severity, messageSQL.String, gpsLocation)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSeverityOutOfRange):
			RespondWithError(w, http.StatusBadRequest, "Severity must be 0, 1, or 2", h.logger, "severity", req.Severity)
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to create off-shift report", h.logger, "error", err.Error())
		}
		return
	}

	// Log audit event for off-shift report creation
	ipAddress, userAgent := GetAuditInfoFromContext(r.Context())

	hasLocation := gpsLocation != nil
	auditErr := h.auditService.LogReportCreated(
		r.Context(),
		userID,
		report.ReportID,
		nil, // No booking ID for off-shift reports
		int64(req.Severity),
		hasLocation,
		ipAddress,
		userAgent,
	)
	if auditErr != nil {
		h.logger.WarnContext(r.Context(), "Failed to log off-shift report creation audit event", "report_id", report.ReportID, "error", auditErr)
	}

	// Convert to API response format
	reportResponse := ToReportResponse(report)
	RespondWithJSON(w, http.StatusCreated, reportResponse, h.logger)
}
