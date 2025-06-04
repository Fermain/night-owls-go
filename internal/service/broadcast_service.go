package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"night-owls-go/internal/config"
	db "night-owls-go/internal/db/sqlc_generated"
)

// BroadcastService handles broadcast processing and delivery
type BroadcastService struct {
	querier db.Querier
	logger  *slog.Logger
	cfg     *config.Config
}

// NewBroadcastService creates a new BroadcastService
func NewBroadcastService(querier db.Querier, logger *slog.Logger, cfg *config.Config) *BroadcastService {
	return &BroadcastService{
		querier: querier,
		logger:  logger.With("service", "BroadcastService"),
		cfg:     cfg,
	}
}

// ProcessPendingBroadcasts processes all pending broadcasts and creates outbox entries
func (s *BroadcastService) ProcessPendingBroadcasts(ctx context.Context) (int, error) {
	pendingBroadcasts, err := s.querier.ListPendingBroadcasts(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to get pending broadcasts", "error", err)
		return 0, err
	}

	if len(pendingBroadcasts) == 0 {
		s.logger.InfoContext(ctx, "No pending broadcasts to process")
		return 0, nil
	}

	s.logger.InfoContext(ctx, "Processing pending broadcasts", "count", len(pendingBroadcasts))

	processed := 0
	for _, broadcast := range pendingBroadcasts {
		if err := s.processBroadcast(ctx, broadcast); err != nil {
			s.logger.ErrorContext(ctx, "Failed to process broadcast", "broadcast_id", broadcast.BroadcastID, "error", err)
			// Mark as failed
			_, updateErr := s.querier.UpdateBroadcastStatus(ctx, db.UpdateBroadcastStatusParams{
				BroadcastID: broadcast.BroadcastID,
				Status:      "failed",
				SentAt:      sql.NullTime{},
				SentCount:   sql.NullInt64{Int64: 0, Valid: true},
				FailedCount: sql.NullInt64{Int64: 1, Valid: true},
			})
			if updateErr != nil {
				s.logger.ErrorContext(ctx, "Failed to update broadcast status to failed", "broadcast_id", broadcast.BroadcastID, "error", updateErr)
			}
			continue
		}
		processed++
	}

	s.logger.InfoContext(ctx, "Completed processing broadcasts", "processed", processed, "failed", len(pendingBroadcasts)-processed)
	return processed, nil
}

// processBroadcast processes a single broadcast by creating outbox entries for all recipients
func (s *BroadcastService) processBroadcast(ctx context.Context, broadcast db.Broadcast) error {
	s.logger.InfoContext(ctx, "Processing broadcast", "broadcast_id", broadcast.BroadcastID, "audience", broadcast.Audience)

	// Mark as sending
	_, err := s.querier.UpdateBroadcastStatus(ctx, db.UpdateBroadcastStatusParams{
		BroadcastID: broadcast.BroadcastID,
		Status:      "sending",
		SentAt:      sql.NullTime{},
		SentCount:   sql.NullInt64{Int64: 0, Valid: true},
		FailedCount: sql.NullInt64{Int64: 0, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to update broadcast status to sending: %w", err)
	}

	// Get recipients based on audience
	recipients, err := s.getRecipients(ctx, broadcast.Audience)
	if err != nil {
		return fmt.Errorf("failed to get recipients: %w", err)
	}

	if len(recipients) == 0 {
		s.logger.WarnContext(ctx, "No recipients found for broadcast", "broadcast_id", broadcast.BroadcastID, "audience", broadcast.Audience)
		// Mark as sent with 0 recipients
		_, err := s.querier.UpdateBroadcastStatus(ctx, db.UpdateBroadcastStatusParams{
			BroadcastID: broadcast.BroadcastID,
			Status:      "sent",
			SentAt:      sql.NullTime{Time: time.Now(), Valid: true},
			SentCount:   sql.NullInt64{Int64: 0, Valid: true},
			FailedCount: sql.NullInt64{Int64: 0, Valid: true},
		})
		return err
	}

	// Create outbox entries for push notifications (if enabled)
	var outboxCount int64
	if broadcast.PushEnabled {
		outboxCount, err = s.createPushOutboxEntries(ctx, broadcast, recipients)
		if err != nil {
			return fmt.Errorf("failed to create push outbox entries: %w", err)
		}
	}

	// Update broadcast status to sent
	_, err = s.querier.UpdateBroadcastStatus(ctx, db.UpdateBroadcastStatusParams{
		BroadcastID: broadcast.BroadcastID,
		Status:      "sent",
		SentAt:      sql.NullTime{Time: time.Now(), Valid: true},
		SentCount:   sql.NullInt64{Int64: outboxCount, Valid: true},
		FailedCount: sql.NullInt64{Int64: 0, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to update broadcast status to sent: %w", err)
	}

	s.logger.InfoContext(ctx, "Successfully processed broadcast",
		"broadcast_id", broadcast.BroadcastID,
		"recipients", len(recipients),
		"outbox_entries", outboxCount)

	return nil
}

// getRecipients gets the list of users based on the audience filter
func (s *BroadcastService) getRecipients(ctx context.Context, audience string) ([]db.ListUsersRow, error) {
	allUsers, err := s.querier.ListUsers(ctx, nil)
	if err != nil {
		return nil, err
	}

	var recipients []db.ListUsersRow
	switch audience {
	case "all":
		recipients = allUsers
	case "admins":
		for _, user := range allUsers {
			if user.Role == "admin" {
				recipients = append(recipients, user)
			}
		}
	case "owls":
		for _, user := range allUsers {
			if user.Role == "owl" || user.Role == "" {
				recipients = append(recipients, user)
			}
		}
	case "active":
		// For now, return all users. In future, filter by last activity
		recipients = allUsers
	default:
		return nil, fmt.Errorf("unknown audience: %s", audience)
	}

	return recipients, nil
}

// createPushOutboxEntries creates outbox entries for push notifications
func (s *BroadcastService) createPushOutboxEntries(ctx context.Context, broadcast db.Broadcast, recipients []db.ListUsersRow) (int64, error) {
	// Create push notification payload
	pushPayload := map[string]interface{}{
		"type":  "broadcast",
		"title": "New Message",
		"body":  broadcast.Message,
		"data": map[string]interface{}{
			"broadcast_id": broadcast.BroadcastID,
			"type":         "broadcast",
		},
	}

	payloadBytes, err := json.Marshal(pushPayload)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal push payload: %w", err)
	}

	var count int64
	for _, recipient := range recipients {
		// Create outbox entry for push notification
		_, err := s.querier.CreateOutboxItem(ctx, db.CreateOutboxItemParams{
			UserID:      sql.NullInt64{Int64: recipient.UserID, Valid: true},
			Recipient:   "", // Not used for push notifications
			MessageType: "push",
			Payload:     sql.NullString{String: string(payloadBytes), Valid: true},
			SendAt:      time.Now().Add(-1 * time.Second),
		})
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to create outbox entry for push notification",
				"user_id", recipient.UserID,
				"broadcast_id", broadcast.BroadcastID,
				"error", err)
			continue
		}
		count++
	}

	return count, nil
}

// ScheduleBroadcast schedules a broadcast for future delivery
func (s *BroadcastService) ScheduleBroadcast(ctx context.Context, broadcastID int64, scheduledAt time.Time) error {
	// This is handled by the ProcessPendingBroadcasts method which checks the scheduled_at field
	// The cron job should call ProcessPendingBroadcasts regularly to handle scheduled broadcasts
	s.logger.InfoContext(ctx, "Broadcast scheduled", "broadcast_id", broadcastID, "scheduled_at", scheduledAt)
	return nil
}
