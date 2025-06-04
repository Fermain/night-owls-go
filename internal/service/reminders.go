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

// EnqueueShiftReminders schedules -24h and -1h push notification reminders for a booking.
// It uses the outbox pattern by creating entries in the outbox table.
func (s *Scheduler) EnqueueShiftReminders(ctx context.Context, booking db.Booking) error {
	payload24h := fmt.Sprintf(`{"type":"shift_reminder","hours":24,"start_time":"%s","booking_id":%d}`, booking.ShiftStart.Format(time.RFC3339), booking.BookingID)
	remindAt24h := booking.ShiftStart.Add(-24 * time.Hour)

	params24h := db.CreateOutboxItemParams{
		MessageType: "push",
		Recipient:   "",
		Payload:     sql.NullString{String: payload24h, Valid: true},
		UserID:      sql.NullInt64{Int64: booking.UserID, Valid: true},
		SendAt:      remindAt24h,
	}
	s.logger.InfoContext(ctx, "Enqueueing 24h shift reminder", "booking_id", booking.BookingID, "user_id", booking.UserID, "remind_at", remindAt24h, "payload", payload24h)
	_, err := s.querier.CreateOutboxItem(ctx, params24h)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to enqueue 24h shift reminder", "booking_id", booking.BookingID, "user_id", booking.UserID, "error", err)
		return fmt.Errorf("failed to enqueue 24h reminder: %w", err)
	}

	payload1h := fmt.Sprintf(`{"type":"shift_reminder","hours":1,"start_time":"%s","booking_id":%d}`, booking.ShiftStart.Format(time.RFC3339), booking.BookingID)
	remindAt1h := booking.ShiftStart.Add(-1 * time.Hour)

	params1h := db.CreateOutboxItemParams{
		MessageType: "push",
		Recipient:   "",
		Payload:     sql.NullString{String: payload1h, Valid: true},
		UserID:      sql.NullInt64{Int64: booking.UserID, Valid: true},
		SendAt:      remindAt1h,
	}
	s.logger.InfoContext(ctx, "Enqueueing 1h shift reminder", "booking_id", booking.BookingID, "user_id", booking.UserID, "remind_at", remindAt1h, "payload", payload1h)
	_, err = s.querier.CreateOutboxItem(ctx, params1h)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to enqueue 1h shift reminder", "booking_id", booking.BookingID, "user_id", booking.UserID, "error", err)
		return fmt.Errorf("failed to enqueue 1h reminder: %w", err)
	}

	s.logger.InfoContext(ctx, "Shift reminders queued", "booking_id", booking.BookingID)
	return nil
}
