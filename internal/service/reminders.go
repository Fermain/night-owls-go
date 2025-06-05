package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	db "night-owls-go/internal/db/sqlc_generated"
	// "night-owls-go/internal/model" // Assuming model.Notification might be defined here or is simple enough to inline
)

// Scheduler handles scheduling of notifications, like shift reminders.
type Scheduler struct {
	querier db.Querier
	logger  *slog.Logger
}

// NewScheduler creates a new Scheduler.
func NewScheduler(querier db.Querier, logger *slog.Logger) *Scheduler {
	return &Scheduler{
		querier: querier,
		logger:  logger.With("service", "Scheduler"),
	}
}

// scheduleReminder creates a scheduled outbox item for a shift reminder.
// This helper reduces duplication in scheduling logic.
func (s *Scheduler) scheduleReminder(ctx context.Context, booking db.Booking, hours int, sendAt time.Time) error {
	payload := fmt.Sprintf(`{"type":"shift_reminder","hours":%d,"start_time":"%s","booking_id":%d}`,
		hours, booking.ShiftStart.Format(time.RFC3339), booking.BookingID)

	params := db.CreateOutboxItemParams{
		MessageType: "push",
		Recipient:   "",
		Payload:     sql.NullString{String: payload, Valid: true},
		UserID:      sql.NullInt64{Int64: booking.UserID, Valid: true},
		SendAt:      sendAt,
	}

	s.logger.InfoContext(ctx, "Enqueueing shift reminder",
		"booking_id", booking.BookingID,
		"user_id", booking.UserID,
		"hours", hours,
		"remind_at", sendAt,
		"payload", payload)

	_, err := s.querier.CreateOutboxItem(ctx, params)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to enqueue shift reminder",
			"booking_id", booking.BookingID,
			"user_id", booking.UserID,
			"hours", hours,
			"error", err)
		return fmt.Errorf("failed to enqueue %dh reminder: %w", hours, err)
	}

	return nil
}

// EnqueueShiftReminders schedules -24h and -1h push notification reminders for a booking.
// It uses the outbox pattern by creating entries in the outbox table.
func (s *Scheduler) EnqueueShiftReminders(ctx context.Context, booking db.Booking) error {
	// Schedule 24-hour reminder
	remindAt24h := booking.ShiftStart.Add(-24 * time.Hour)
	if err := s.scheduleReminder(ctx, booking, 24, remindAt24h); err != nil {
		return err
	}

	// Schedule 1-hour reminder
	remindAt1h := booking.ShiftStart.Add(-1 * time.Hour)
	if err := s.scheduleReminder(ctx, booking, 1, remindAt1h); err != nil {
		return err
	}

	s.logger.InfoContext(ctx, "Shift reminders queued", "booking_id", booking.BookingID)
	return nil
}
