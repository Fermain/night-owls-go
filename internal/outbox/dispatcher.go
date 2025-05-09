package outbox

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"night-owls-go/internal/config" // For config values
	db "night-owls-go/internal/db/sqlc_generated"
)

// const (
// 	defaultBatchSize = 10 // Moved to config
// 	maxRetryCount    = 3  // Moved to config
// )

// DispatcherService processes pending messages from the outbox.
type DispatcherService struct {
	querier  db.Querier
	sender   MessageSender
	logger   *slog.Logger
	cfg      *config.Config // Added config
}

// NewDispatcherService creates a new DispatcherService.
func NewDispatcherService(querier db.Querier, sender MessageSender, logger *slog.Logger, cfg *config.Config) *DispatcherService {
	return &DispatcherService{
		querier:  querier,
		sender:   sender,
		logger:   logger.With("service", "OutboxDispatcher"),
		cfg:      cfg, // Store config
	}
}

// ProcessPendingOutboxMessages fetches and processes pending outbox messages.
func (s *DispatcherService) ProcessPendingOutboxMessages(ctx context.Context) (processedCount int, errCount int) {
	s.logger.InfoContext(ctx, "Starting to process pending outbox messages...")

	pendingItems, err := s.querier.GetPendingOutboxItems(ctx, int64(s.cfg.OutboxBatchSize)) // Use from config
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.InfoContext(ctx, "No pending outbox messages to process.")
			return 0, 0
		}
		s.logger.ErrorContext(ctx, "Failed to get pending outbox items", "error", err)
		return 0, 1 
	}

	if len(pendingItems) == 0 {
		s.logger.InfoContext(ctx, "No pending outbox messages found in this batch.")
		return 0, 0
	}

	s.logger.InfoContext(ctx, "Fetched pending outbox messages", "count", len(pendingItems))

	for _, item := range pendingItems {
		sendCtx, cancel := context.WithTimeout(ctx, 30*time.Second) 
		err = s.sender.Send(item.Recipient, item.MessageType, item.Payload.String)
		cancel() 

		updateParams := db.UpdateOutboxItemStatusParams{
			OutboxID:   item.OutboxID,
			RetryCount: item.RetryCount, 
		}

		if err != nil {
			s.logger.ErrorContext(sendCtx, "Failed to send message from outbox", "outbox_id", item.OutboxID, "error", err)
			errCount++
			updateParams.Status = "failed"
			updateParams.RetryCount.Int64 = item.RetryCount.Int64 + 1
			updateParams.RetryCount.Valid = true
			if updateParams.RetryCount.Int64 >= int64(s.cfg.OutboxMaxRetries) { // Use from config
				s.logger.WarnContext(sendCtx, "Message reached max retry count, marking as permanently failed", "outbox_id", item.OutboxID)
				updateParams.Status = "permanently_failed" 
			}
		} else {
			s.logger.InfoContext(sendCtx, "Successfully sent message from outbox", "outbox_id", item.OutboxID)
			updateParams.Status = "sent"
			updateParams.SentAt = sql.NullTime{Time: time.Now(), Valid: true}
			processedCount++
		}

		_, updateErr := s.querier.UpdateOutboxItemStatus(ctx, updateParams)
		if updateErr != nil {
			s.logger.ErrorContext(ctx, "Failed to update outbox item status", "outbox_id", item.OutboxID, "error", updateErr)
			errCount++ 
		}
	}
	return processedCount, errCount
} 