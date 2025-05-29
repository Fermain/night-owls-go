package service

import (
	"context"
	"log/slog"
	"time"

	db "night-owls-go/internal/db/sqlc_generated"
)

// ReportArchivingService handles automatic archiving of reports based on retention policies
type ReportArchivingService struct {
	querier db.Querier
	logger  *slog.Logger
}

// NewReportArchivingService creates a new ReportArchivingService
func NewReportArchivingService(querier db.Querier, logger *slog.Logger) *ReportArchivingService {
	return &ReportArchivingService{
		querier: querier,
		logger:  logger.With("service", "ReportArchivingService"),
	}
}

// ArchiveOldReports automatically archives reports based on retention policies:
// - Normal reports (severity 0): archived after 1 month
// - Suspicion reports (severity 1): archived after 1 year
// - Incident reports (severity 2): never auto-archived
func (s *ReportArchivingService) ArchiveOldReports(ctx context.Context) (int, error) {
	s.logger.InfoContext(ctx, "Starting automatic report archiving process")

	// Get reports that should be archived
	reportsToArchive, err := s.querier.GetReportsForAutoArchiving(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to get reports for auto-archiving", "error", err)
		return 0, err
	}

	if len(reportsToArchive) == 0 {
		s.logger.InfoContext(ctx, "No reports found for auto-archiving")
		return 0, nil
	}

	// Extract report IDs
	reportIDs := make([]int64, len(reportsToArchive))
	for i, report := range reportsToArchive {
		reportIDs[i] = report.ReportID
	}

	// Bulk archive the reports
	err = s.querier.BulkArchiveReports(ctx, reportIDs)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to bulk archive reports", "error", err, "report_count", len(reportIDs))
		return 0, err
	}

	// Log details about what was archived
	severityCounts := make(map[int64]int)
	for _, report := range reportsToArchive {
		severityCounts[report.Severity]++
	}

	s.logger.InfoContext(ctx, "Successfully auto-archived reports",
		"total_archived", len(reportIDs),
		"normal_reports", severityCounts[0],
		"suspicion_reports", severityCounts[1],
		"incident_reports", severityCounts[2])

	return len(reportIDs), nil
}

// GetArchivingStats returns statistics about archivable reports
func (s *ReportArchivingService) GetArchivingStats(ctx context.Context) (map[string]interface{}, error) {
	reportsToArchive, err := s.querier.GetReportsForAutoArchiving(ctx)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_archivable": len(reportsToArchive),
		"by_severity": map[string]int{
			"normal":    0,
			"suspicion": 0,
			"incident":  0,
		},
		"oldest_archivable": nil,
	}

	if len(reportsToArchive) == 0 {
		return stats, nil
	}

	var oldestTime *time.Time
	severityMap := stats["by_severity"].(map[string]int)

	for _, report := range reportsToArchive {
		switch report.Severity {
		case 0:
			severityMap["normal"]++
		case 1:
			severityMap["suspicion"]++
		case 2:
			severityMap["incident"]++
		}

		if report.CreatedAt.Valid {
			if oldestTime == nil || report.CreatedAt.Time.Before(*oldestTime) {
				oldestTime = &report.CreatedAt.Time
			}
		}
	}

	if oldestTime != nil {
		stats["oldest_archivable"] = oldestTime.Format(time.RFC3339)
	}

	return stats, nil
}
