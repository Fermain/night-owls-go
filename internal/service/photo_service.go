package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	db "night-owls-go/internal/db/sqlc_generated"
)

type PhotoService struct {
	querier     db.Querier
	logger      *slog.Logger
	storagePath string
	maxFileSize int64
}

type PhotoUploadRequest struct {
	File     multipart.File
	Header   *multipart.FileHeader
	ReportID int64
	UserID   int64
}

type PhotoMetadata struct {
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

func NewPhotoService(querier db.Querier, logger *slog.Logger) *PhotoService {
	storagePath := "./data/report-photos"
	
	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		logger.Error("Failed to create photo storage directory", "error", err)
	}

	return &PhotoService{
		querier:     querier,
		logger:      logger,
		storagePath: storagePath,
		maxFileSize: 10 * 1024 * 1024, // 10MB
	}
}

func (ps *PhotoService) UploadPhoto(ctx context.Context, req PhotoUploadRequest) (*PhotoMetadata, error) {
	// Validate file size
	if req.Header.Size > ps.maxFileSize {
		return nil, fmt.Errorf("file size %d exceeds maximum allowed size %d", req.Header.Size, ps.maxFileSize)
	}

	// Validate mime type
	contentType := req.Header.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		return nil, fmt.Errorf("invalid file type: %s", contentType)
	}

	// Read file content
	fileBytes, err := io.ReadAll(req.File)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Calculate checksum
	checksum := fmt.Sprintf("%x", sha256.Sum256(fileBytes))

	// Get image dimensions
	width, height := getImageDimensions(fileBytes)

	// Generate filename
	timestamp := time.Now()
	filename := fmt.Sprintf("%d_%d_%s.%s", 
		req.ReportID, 
		timestamp.Unix(), 
		checksum[:8], 
		getFileExtension(contentType))

	// Create storage path with date structure
	dateDir := timestamp.Format("2006/01/02")
	fullStorageDir := filepath.Join(ps.storagePath, dateDir)
	if err := os.MkdirAll(fullStorageDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Save file
	fullPath := filepath.Join(fullStorageDir, filename)
	if err := os.WriteFile(fullPath, fileBytes, 0644); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Store in database
	photoData, err := ps.querier.CreateReportPhoto(ctx, db.CreateReportPhotoParams{
		ReportID:         req.ReportID,
		Filename:         filename,
		OriginalFilename: sql.NullString{String: req.Header.Filename, Valid: true},
		FileSizeBytes:    req.Header.Size,
		MimeType:         contentType,
		WidthPixels:      convertIntToNullInt64(width),
		HeightPixels:     convertIntToNullInt64(height),
		StoragePath:      fullPath,
		ThumbnailPath:    sql.NullString{String: "", Valid: false},
		ChecksumSha256:   checksum,
		IsProcessed:      sql.NullBool{Bool: false, Valid: true},
	})
	if err != nil {
		// Clean up file if database insert fails
		os.Remove(fullPath)
		return nil, fmt.Errorf("failed to save photo metadata: %w", err)
	}

	// Update report photo count
	if err := ps.querier.UpdateReportPhotoCount(ctx, req.ReportID); err != nil {
		ps.logger.Warn("Failed to update report photo count", "report_id", req.ReportID, "error", err)
	}

	ps.logger.Info("Photo uploaded successfully", 
		"photo_id", photoData.PhotoID, 
		"report_id", req.ReportID, 
		"filename", filename,
		"size_bytes", req.Header.Size)

	return &PhotoMetadata{
		PhotoID:       photoData.PhotoID,
		ReportID:      photoData.ReportID,
		Filename:      photoData.Filename,
		OriginalName:  req.Header.Filename,
		FileSizeBytes: photoData.FileSizeBytes,
		MimeType:      photoData.MimeType,
		Width:         convertNullableInt(photoData.WidthPixels),
		Height:        convertNullableInt(photoData.HeightPixels),
		UploadTime:    photoData.UploadTimestamp.Time,
		PhotoURL:      fmt.Sprintf("/api/reports/%d/photos/%d", req.ReportID, photoData.PhotoID),
	}, nil
}

func (ps *PhotoService) GetReportPhotos(ctx context.Context, reportID int64) ([]PhotoMetadata, error) {
	photos, err := ps.querier.GetReportPhotos(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report photos: %w", err)
	}

	result := make([]PhotoMetadata, len(photos))
	for i, photo := range photos {
		result[i] = PhotoMetadata{
			PhotoID:       photo.PhotoID,
			ReportID:      photo.ReportID,
			Filename:      photo.Filename,
			OriginalName:  photo.OriginalFilename.String,
			FileSizeBytes: photo.FileSizeBytes,
			MimeType:      photo.MimeType,
			Width:         convertNullableInt(photo.WidthPixels),
			Height:        convertNullableInt(photo.HeightPixels),
			UploadTime:    photo.UploadTimestamp.Time,
			PhotoURL:      fmt.Sprintf("/api/reports/%d/photos/%d", reportID, photo.PhotoID),
		}
	}

	return result, nil
}

func (ps *PhotoService) DeletePhoto(ctx context.Context, photoID, reportID int64) error {
	// Get photo info first
	photo, err := ps.querier.GetReportPhoto(ctx, db.GetReportPhotoParams{
		PhotoID:  photoID,
		ReportID: reportID,
	})
	if err != nil {
		return fmt.Errorf("photo not found: %w", err)
	}

	// Delete file
	if err := os.Remove(photo.StoragePath); err != nil {
		ps.logger.Warn("Failed to delete photo file", "path", photo.StoragePath, "error", err)
	}

	// Delete from database
	if err := ps.querier.DeleteReportPhoto(ctx, db.DeleteReportPhotoParams{
		PhotoID:  photoID,
		ReportID: reportID,
	}); err != nil {
		return fmt.Errorf("failed to delete photo from database: %w", err)
	}

	// Update report photo count
	if err := ps.querier.UpdateReportPhotoCount(ctx, reportID); err != nil {
		ps.logger.Warn("Failed to update report photo count", "report_id", reportID, "error", err)
	}

	ps.logger.Info("Photo deleted successfully", "photo_id", photoID, "report_id", reportID)
	return nil
}

// Helper functions

func isValidImageType(contentType string) bool {
	validTypes := []string{"image/jpeg", "image/png", "image/webp"}
	for _, validType := range validTypes {
		if contentType == validType {
			return true
		}
	}
	return false
}

func getFileExtension(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/webp":
		return "webp"
	default:
		return "jpg"
	}
}

func getImageDimensions(data []byte) (*int, *int) {
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, nil
	}
	width := config.Width
	height := config.Height
	return &width, &height
}

func convertNullableInt(val sql.NullInt64) *int {
	if val.Valid {
		intVal := int(val.Int64)
		return &intVal
	}
	return nil
}

func convertIntToNullInt64(val *int) sql.NullInt64 {
	if val == nil {
		return sql.NullInt64{Int64: 0, Valid: false}
	}
	return sql.NullInt64{Int64: int64(*val), Valid: true}
} 