package outbox

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	db "night-owls-go/internal/db/sqlc_generated"
)

const (
	defaultBatchSize = 10 // Number of pending messages to fetch at once
	maxRetryCount    = 3  // Max number of retries for a message
)

// DispatcherService processes pending messages from the outbox.
type DispatcherService struct {
	querier  db.Querier
	sender   MessageSender
	logger   *slog.Logger
	// Potentially add a Ticker or similar for periodic execution if not managed externally by cron
}

// NewDispatcherService creates a new DispatcherService.
func NewDispatcherService(querier db.Querier, sender MessageSender, logger *slog.Logger) *DispatcherService {
	return &DispatcherService{
		querier:  querier,
		sender:   sender,
		logger:   logger.With("service", "OutboxDispatcher"),
	}
}

// ProcessPendingOutboxMessages fetches and processes pending outbox messages.
func (s *DispatcherService) ProcessPendingOutboxMessages(ctx context.Context) (processedCount int, errCount int) {
	s.logger.InfoContext(ctx, "Starting to process pending outbox messages...")

	pendingItems, err := s.querier.GetPendingOutboxItems(ctx, defaultBatchSize)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.InfoContext(ctx, "No pending outbox messages to process.")
			return 0, 0
		}
		s.logger.ErrorContext(ctx, "Failed to get pending outbox items", "error", err)
		return 0, 1 // Indicate an error occurred during fetching
	}

	if len(pendingItems) == 0 {
		s.logger.InfoContext(ctx, "No pending outbox messages found in this batch.")
		return 0, 0
	}

	s.logger.InfoContext(ctx, "Fetched pending outbox messages", "count", len(pendingItems))

	for _, item := range pendingItems {
		// Use a new context for each send attempt to allow individual timeouts if needed.
		sendCtx, cancel := context.WithTimeout(ctx, 30*time.Second) // Example timeout for send operation

		err = s.sender.Send(item.Recipient, item.MessageType, item.Payload.String)
		
		cancel() // Release context resources

		updateParams := db.UpdateOutboxItemStatusParams{
			OutboxID:   item.OutboxID,
			RetryCount: item.RetryCount, // Keep current retry count if send was successful or retries exhausted
		}

		if err != nil {
			s.logger.ErrorContext(sendCtx, "Failed to send message from outbox", "outbox_id", item.OutboxID, "recipient", item.Recipient, "type", item.MessageType, "error", err)
			errCount++
			updateParams.Status = "failed"
			updateParams.RetryCount.Int64 = item.RetryCount.Int64 + 1
			updateParams.RetryCount.Valid = true
			if updateParams.RetryCount.Int64 >= maxRetryCount {
				s.logger.WarnContext(sendCtx, "Message reached max retry count, marking as permanently failed", "outbox_id", item.OutboxID)
				updateParams.Status = "permanently_failed" // Or some other terminal status
			}
		} else {
			s.logger.InfoContext(sendCtx, "Successfully sent message from outbox", "outbox_id", item.OutboxID, "recipient", item.Recipient, "type", item.MessageType)
			updateParams.Status = "sent"
			updateParams.SentAt = sql.NullTime{Time: time.Now(), Valid: true}
			processedCount++
		}

		_, updateErr := s.querier.UpdateOutboxItemStatus(ctx, updateParams)
		if updateErr != nil {
			s.logger.ErrorContext(ctx, "Failed to update outbox item status", "outbox_id", item.OutboxID, "new_status", updateParams.Status, "error", updateErr)
			// This is a more critical error as the outbox state might become inconsistent.
			// Depending on strategy, might stop processing or just log and continue.
			errCount++ // Count this as an error too
		}
	}
	return processedCount, errCount
} 