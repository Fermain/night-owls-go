package service

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	db "night-owls-go/internal/db/sqlc_generated"
)

var (
	ErrSeverityOutOfRange = errors.New("severity must be between 0 and 2")
	ErrReportBookingAuth  = errors.New("user not authorized to report for this booking or booking does not exist")
)

// ReportService handles logic related to incident reports.
type ReportService struct {
	querier db.Querier
	logger  *slog.Logger
}

// NewReportService creates a new ReportService.
func NewReportService(querier db.Querier, logger *slog.Logger) *ReportService {
	return &ReportService{
		querier: querier,
		logger:  logger.With("service", "ReportService"),
	}
}

// CreateReport handles the logic for creating a new incident report.
func (s *ReportService) CreateReport(ctx context.Context, userIDFromAuth int64, bookingID int64, severity int32, message string) (db.Report, error) {
	// 1. Validate booking exists and user is authorized
	// A user can only report on their own bookings.
	booking, err := s.querier.GetBookingByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.WarnContext(ctx, "Booking not found for report creation", "booking_id", bookingID)
			return db.Report{}, ErrReportBookingAuth // Or ErrBookingNotFound if more specific error needed by handler
		}
		s.logger.ErrorContext(ctx, "Failed to get booking by ID for report creation", "booking_id", bookingID, "error", err)
		return db.Report{}, ErrInternalServer
	}

	if booking.UserID != userIDFromAuth {
		s.logger.WarnContext(ctx, "User forbidden to create report for booking", "booking_id", bookingID, "booking_owner_id", booking.UserID, "auth_user_id", userIDFromAuth)
		return db.Report{}, ErrReportBookingAuth
	}

	// 2. Validate severity (0-2)
	if severity < 0 || severity > 2 {
		s.logger.WarnContext(ctx, "Severity out of range for report", "severity", severity)
		return db.Report{}, ErrSeverityOutOfRange
	}

	// 3. Insert report into DB
	reportParams := db.CreateReportParams{
		BookingID: sql.NullInt64{Int64: bookingID, Valid: true},
		UserID:    sql.NullInt64{Int64: userIDFromAuth, Valid: true},
		Severity:  int64(severity),
		Message:   sql.NullString{String: message, Valid: message != ""},
	}

	createdReport, err := s.querier.CreateReport(ctx, reportParams)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to create report in DB", "params", reportParams, "error", err)
		return db.Report{}, ErrInternalServer
	}

	s.logger.InfoContext(ctx, "Report created successfully", "report_id", createdReport.ReportID, "booking_id", bookingID)
	return createdReport, nil
}

// CreateOffShiftReport handles the logic for creating an off-shift incident report.
func (s *ReportService) CreateOffShiftReport(ctx context.Context, userIDFromAuth int64, severity int32, message string) (db.Report, error) {
	// 1. Validate severity (0-2)
	if severity < 0 || severity > 2 {
		s.logger.WarnContext(ctx, "Severity out of range for off-shift report", "severity", severity)
		return db.Report{}, ErrSeverityOutOfRange
	}

	// 2. Insert off-shift report into DB
	reportParams := db.CreateOffShiftReportParams{
		UserID:   sql.NullInt64{Int64: userIDFromAuth, Valid: true},
		Severity: int64(severity),
		Message:  sql.NullString{String: message, Valid: message != ""},
	}

	createdReport, err := s.querier.CreateOffShiftReport(ctx, reportParams)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to create off-shift report in DB", "params", reportParams, "error", err)
		return db.Report{}, ErrInternalServer
	}

	s.logger.InfoContext(ctx, "Off-shift report created successfully", "report_id", createdReport.ReportID, "user_id", userIDFromAuth)
	return createdReport, nil
} 