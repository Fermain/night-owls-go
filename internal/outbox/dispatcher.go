package outbox

import (
	"context"
	"database/sql"
	"errors"
	"log/slog" // For converting UserID from string if necessary, though it should be int64 from DB
	"time"

	"night-owls-go/internal/config" // For config values
	db "night-owls-go/internal/db/sqlc_generated"
	"night-owls-go/internal/service" // For PushSender
)

// const (
// 	defaultBatchSize = 10 // Moved to config
// 	maxRetryCount    = 3  // Moved to config
// )

// DispatcherService processes pending messages from the outbox.
type DispatcherService struct {
	querier    db.Querier
	smsSender  MessageSender // Renaming 'sender' to 'smsSender' for clarity
	pushSender *service.PushSender
	logger     *slog.Logger
	cfg        *config.Config
}

// NewDispatcherService creates a new DispatcherService.
func NewDispatcherService(querier db.Querier, smsSender MessageSender, pushSender *service.PushSender, logger *slog.Logger, cfg *config.Config) *DispatcherService {
	return &DispatcherService{
		querier:    querier,
		smsSender:  smsSender,
		pushSender: pushSender,
		logger:     logger.With("service", "OutboxDispatcher"),
		cfg:        cfg,
	}
}

// ProcessPendingOutboxItems processes pending outbox items and dispatches them via the appropriate sender.
func (s *DispatcherService) ProcessPendingOutboxItems(ctx context.Context) (int, int) {
	pendingItems, err := s.querier.GetPendingOutboxItems(ctx, int64(s.cfg.OutboxBatchSize))
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to get pending outbox items", "error", err)
		return 0, 1 // Increment errCount for fetch failure
	}

	if len(pendingItems) == 0 {
		s.logger.InfoContext(ctx, "No pending outbox items to process")
		return 0, 0
	}

	s.logger.InfoContext(ctx, "Processing pending outbox items", "count", len(pendingItems))

	processedCount := 0
	errCount := 0

	for _, item := range pendingItems {
		sendCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		var dispatchErr error

		switch item.MessageType {
		case "sms":
			if s.smsSender != nil {
				dispatchErr = s.smsSender.Send(item.Recipient, item.MessageType, item.Payload.String)
			} else {
				dispatchErr = errors.New("smsSender not configured")
				s.logger.ErrorContext(sendCtx, dispatchErr.Error(), "outbox_id", item.OutboxID)
			}
		case "push":
			if s.pushSender != nil {
				if !item.UserID.Valid {
					dispatchErr = errors.New("user_id is null for push notification")
					s.logger.ErrorContext(sendCtx, dispatchErr.Error(), "outbox_id", item.OutboxID)
				} else {
					dispatchErr = s.pushSender.Send(sendCtx, item.UserID.Int64, []byte(item.Payload.String), 604800) // Use 1 week TTL
					if dispatchErr != nil {
						s.logger.ErrorContext(sendCtx, "PushSender failed", "outbox_id", item.OutboxID, "error", dispatchErr)
					}
				}
			} else {
				dispatchErr = errors.New("pushSender not configured")
				s.logger.ErrorContext(sendCtx, dispatchErr.Error(), "outbox_id", item.OutboxID)
			}
		default:
			dispatchErr = errors.New("unknown message type: " + item.MessageType)
			s.logger.WarnContext(sendCtx, "Unknown message type in outbox", "outbox_id", item.OutboxID, "message_type", item.MessageType)
		}
		cancel()

		updateParams := db.UpdateOutboxItemStatusParams{
			OutboxID:   item.OutboxID,
			RetryCount: item.RetryCount,
		}

		if dispatchErr != nil {
			s.logger.ErrorContext(ctx, "Failed to dispatch message from outbox", "outbox_id", item.OutboxID, "message_type", item.MessageType, "error", dispatchErr)
			errCount++
			updateParams.Status = "failed"
			updateParams.RetryCount.Int64 = item.RetryCount.Int64 + 1
			updateParams.RetryCount.Valid = true
			if updateParams.RetryCount.Int64 >= int64(s.cfg.OutboxMaxRetries) {
				s.logger.WarnContext(ctx, "Message reached max retry count, marking as permanently failed", "outbox_id", item.OutboxID)
				updateParams.Status = "permanently_failed"
			}
		} else {
			s.logger.InfoContext(ctx, "Successfully dispatched message from outbox", "outbox_id", item.OutboxID, "message_type", item.MessageType)
			processedCount++
			updateParams.Status = "sent"
			updateParams.SentAt = sql.NullTime{Time: time.Now(), Valid: true}
		}

		_, updateErr := s.querier.UpdateOutboxItemStatus(ctx, updateParams)
		if updateErr != nil {
			s.logger.ErrorContext(ctx, "Failed to update outbox item status", "outbox_id", item.OutboxID, "error", updateErr)
			errCount++
		}
	}
	return processedCount, errCount
}
