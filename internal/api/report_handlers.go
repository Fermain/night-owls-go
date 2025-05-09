package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"night-owls-go/internal/service"

	"github.com/go-chi/chi/v5"
)

// ReportHandler handles incident report related HTTP requests.
type ReportHandler struct {
	reportService *service.ReportService
	logger        *slog.Logger
}

// NewReportHandler creates a new ReportHandler.
func NewReportHandler(reportService *service.ReportService, logger *slog.Logger) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
		logger:        logger.With("handler", "ReportHandler"),
	}
}

// CreateReportRequest is the expected JSON for POST /bookings/{id}/report
type CreateReportRequest struct {
	Severity int32  `json:"severity"` // Expect 0, 1, or 2
	Message  string `json:"message,omitempty"`
}

// CreateReportHandler handles POST /bookings/{id}/report
func (h *ReportHandler) CreateReportHandler(w http.ResponseWriter, r *http.Request) {
	userIDFromAuthVal := r.Context().Value(UserIDKey)
	userIDFromAuth, ok := userIDFromAuthVal.(int64)
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User ID not found in context or invalid type for auth", h.logger)
		return
	}

	bookingIDStr := chi.URLParam(r, "id") // "id" is the placeholder for booking ID in the route
	bookingID, err := strconv.ParseInt(bookingIDStr, 10, 64)
	if err != nil || bookingID <= 0 {
		RespondWithError(w, http.StatusBadRequest, "Invalid booking ID in path", h.logger, "booking_id_str", bookingIDStr)
		return
	}

	var req CreateReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload", h.logger, "error", err.Error())
		return
	}
	defer r.Body.Close()

	// Validate severity here at the handler level as well, though service also validates
	if req.Severity < 0 || req.Severity > 2 {
		RespondWithError(w, http.StatusBadRequest, "Severity must be between 0 and 2", h.logger, "severity_provided", req.Severity)
		return
	}

	createdReport, err := h.reportService.CreateReport(r.Context(), userIDFromAuth, bookingID, req.Severity, req.Message)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrReportBookingAuth):
			RespondWithError(w, http.StatusForbidden, "Not authorized to create report for this booking or booking not found", h.logger, "booking_id", bookingID)
		case errors.Is(err, service.ErrSeverityOutOfRange): // Should be caught by handler, but good to have service error mapping
			RespondWithError(w, http.StatusBadRequest, "Severity value is out of range (0-2)", h.logger, "severity", req.Severity)
		case errors.Is(err, service.ErrBookingNotFound): // More specific if service distinguishes from general auth error
			RespondWithError(w, http.StatusNotFound, "Booking not found to create report against", h.logger, "booking_id", bookingID)
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to create report", h.logger, "error", err.Error())
		}
		return
	}

	RespondWithJSON(w, http.StatusCreated, createdReport, h.logger)
}

// ListReportsHandler (Optional as per guide) - Placeholder
func (h *ReportHandler) ListReportsHandler(w http.ResponseWriter, r *http.Request) {
	// userIDFromAuthVal := r.Context().Value(UserIDKey)
	// userIDFromAuth, ok := userIDFromAuthVal.(int64)
	// if !ok { ... }
	// Implementation would involve calling a service method like h.reportService.ListReportsByUser(r.Context(), userIDFromAuth)
	// and then RespondWithJSON.
	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Listing reports - TBD"}, h.logger)
} 