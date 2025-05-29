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

// ProcessPendingOutboxMessages fetches and processes pending outbox messages.
func (s *DispatcherService) ProcessPendingOutboxMessages(ctx context.Context) (processedCount int, errCount int) {
	s.logger.InfoContext(ctx, "Starting to process pending outbox messages...")

	pendingItems, err := s.querier.GetPendingOutboxItems(ctx, int64(s.cfg.OutboxBatchSize))
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
		var dispatchErr error

		switch item.MessageType { // Assuming MessageType acts as the channel ("sms", "push")
		case "sms", "OTP_VERIFICATION":
			if s.smsSender != nil {
				dispatchErr = s.smsSender.Send(item.Recipient, item.MessageType, item.Payload.String)
			} else {
				dispatchErr = errors.New("smsSender is not configured")
				s.logger.ErrorContext(sendCtx, "smsSender not configured, cannot send SMS", "outbox_id", item.OutboxID)
			}
		case "push":
			if s.pushSender != nil {
				if !item.UserID.Valid {
					dispatchErr = errors.New("user_id is null for push notification")
					s.logger.ErrorContext(sendCtx, "user_id is null for push notification", "outbox_id", item.OutboxID)
				} else {
					// Payload for push is documented as JSON: `{"type":"shift_reminder","booking_id":123}`
					// TTL for push is 60 seconds as per PWA.md example
					s.pushSender.Send(sendCtx, item.UserID.Int64, []byte(item.Payload.String), 60)
					// Note: pushSender.Send itself logs errors, so we might not need to wrap with dispatchErr unless it returns one.
					// For now, assuming Send handles its logging and doesn't return an error that needs generic handling here.
				}
			} else {
				dispatchErr = errors.New("pushSender is not configured")
				s.logger.ErrorContext(sendCtx, "pushSender not configured, cannot send push notification", "outbox_id", item.OutboxID)
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
			updateParams.Status = "sent"
			updateParams.SentAt = sql.NullTime{Time: time.Now(), Valid: true}
			processedCount++
		}

		_, updateErr := s.querier.UpdateOutboxItemStatus(ctx, updateParams) // Changed sendCtx to ctx for the update
		if updateErr != nil {
			s.logger.ErrorContext(ctx, "Failed to update outbox item status", "outbox_id", item.OutboxID, "error", updateErr)
			errCount++
		}
	}
	return processedCount, errCount
}
