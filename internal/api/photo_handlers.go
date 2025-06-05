package api

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"night-owls-go/internal/service"
)

// PhotoHandler handles photo-related HTTP requests.
type PhotoHandler struct {
	photoService *service.PhotoService
	logger       *slog.Logger
}

// NewPhotoHandler creates a new PhotoHandler.
func NewPhotoHandler(photoService *service.PhotoService, logger *slog.Logger) *PhotoHandler {
	return &PhotoHandler{
		photoService: photoService,
		logger:       logger.With("handler", "PhotoHandler"),
	}
}

// PhotoResponse represents a photo in API responses.
type PhotoResponse struct {
	PhotoID       int64     `json:"photo_id"`
	ReportID      int64     `json:"report_id"`
	Filename      string    `json:"filename"`
	OriginalName  string    `json:"original_filename"`
	FileSizeBytes int64     `json:"file_size_bytes"`
	MimeType      string    `json:"mime_type"`
	Width         *int      `json:"width_pixels,omitempty"`
	Height        *int      `json:"height_pixels,omitempty"`
	UploadTime    time.Time `json:"upload_timestamp"`
	PhotoURL      string    `json:"photo_url"`
}

// GetReportPhotosHandler handles GET /api/admin/report-photos/{reportId}
// @Summary Get photos for a report
// @Description Get all photos attached to a specific report
// @Tags photos
// @Produce json
// @Param id path int true "Report ID"
// @Success 200 {array} PhotoResponse "List of photos for the report"
// @Failure 400 {object} ErrorResponse "Invalid report ID"
// @Failure 403 {object} ErrorResponse "Forbidden - not authorized to view this report's photos"
// @Failure 404 {object} ErrorResponse "Report not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/admin/report-photos/{reportId} [get]
func (h *PhotoHandler) GetReportPhotosHandler(w http.ResponseWriter, r *http.Request) {
	// Extract report ID from URL
	idStr := r.PathValue("reportId")

	reportID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to parse report ID", "id_param", idStr, "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid report ID", h.logger)
		return
	}

	// Get photos for the report
	photos, err := h.photoService.GetReportPhotos(r.Context(), reportID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to get report photos", "report_id", reportID, "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch photos", h.logger)
		return
	}

	// Convert to API response format
	photoResponses := make([]PhotoResponse, len(photos))
	for i, photo := range photos {
		photoResponses[i] = PhotoResponse{
			PhotoID:       photo.PhotoID,
			ReportID:      photo.ReportID,
			Filename:      photo.Filename,
			OriginalName:  photo.OriginalName,
			FileSizeBytes: photo.FileSizeBytes,
			MimeType:      photo.MimeType,
			Width:         photo.Width,
			Height:        photo.Height,
			UploadTime:    photo.UploadTime,
			PhotoURL:      photo.PhotoURL,
		}
	}

	RespondWithJSON(w, http.StatusOK, photoResponses, h.logger)
}

// DeleteReportPhotoHandler handles DELETE /api/admin/report-photos/{reportId}/{photoId}
// @Summary Delete a photo from a report (Admin only)
// @Description Delete a specific photo from a report
// @Tags photos
// @Param id path int true "Report ID"
// @Param photoId path int true "Photo ID"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} ErrorResponse "Invalid ID parameters"
// @Failure 403 {object} ErrorResponse "Forbidden - admin access required"
// @Failure 404 {object} ErrorResponse "Photo not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/admin/report-photos/{reportId}/{photoId} [delete]
func (h *PhotoHandler) DeleteReportPhotoHandler(w http.ResponseWriter, r *http.Request) {
	// Extract report ID from URL
	idStr := r.PathValue("reportId")
	reportID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to parse report ID", "id_param", idStr, "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid report ID", h.logger)
		return
	}

	// Extract photo ID from URL
	photoIdStr := r.PathValue("photoId")
	photoID, err := strconv.ParseInt(photoIdStr, 10, 64)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to parse photo ID", "photo_id_param", photoIdStr, "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid photo ID", h.logger)
		return
	}

	// Delete the photo
	err = h.photoService.DeletePhoto(r.Context(), photoID, reportID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to delete photo", "report_id", reportID, "photo_id", photoID, "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete photo", h.logger)
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Photo deleted successfully"}, h.logger)
} 