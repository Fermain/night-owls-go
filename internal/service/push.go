package service

import (
	"context"
	"fmt"
	"log/slog"

	"night-owls-go/internal/config"
	db "night-owls-go/internal/db/sqlc_generated"

	webpush "github.com/SherClockHolmes/webpush-go"
)

// PushSender sends web push notifications.
type PushSender struct {
	db     db.Querier
	config *config.Config
	logger *slog.Logger
}

// NewPushSender creates a new PushSender.
func NewPushSender(db db.Querier, cfg *config.Config, logger *slog.Logger) *PushSender {
	return &PushSender{
		db:     db,
		config: cfg,
		logger: logger.With("service", "PushSender"),
	}
}

// Send sends a push notification to all registered subscriptions for a user.
// Now returns an error if any send fails, for better upstream handling.
func (s *PushSender) Send(ctx context.Context, userID int64, payload []byte, ttl int) error {
	if ttl == 0 {
		ttl = 604800 // Default to 1 week if not specified
	}

	subs, err := s.db.GetSubscriptionsByUser(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get subscriptions by user", "user_id", userID, "error", err)
		return err
	}

	if len(subs) == 0 {
		s.logger.InfoContext(ctx, "no subscriptions found for user, skipping push send", "user_id", userID)
		return nil
	}

	var lastErr error
	for _, sub := range subs {
		subscription := &webpush.Subscription{
			Endpoint: sub.Endpoint,
			Keys:     webpush.Keys{P256dh: sub.P256dhKey, Auth: sub.AuthKey},
		}

		resp, err := webpush.SendNotification(payload, subscription, &webpush.Options{
			VAPIDPublicKey:  s.config.VAPIDPublic,
			VAPIDPrivateKey: s.config.VAPIDPrivate,
			TTL:             ttl,
			Subscriber:      s.config.VAPIDSubject,
			Urgency:         "high", // Add for FCM priority
		})
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}

		if err != nil {
			s.logger.ErrorContext(ctx, "failed to send web push notification", "user_id", userID, "endpoint", sub.Endpoint, "error", err)
			lastErr = err // Capture last error
			// Clean up expired subscriptions
			if resp != nil && (resp.StatusCode == 404 || resp.StatusCode == 410) {
				params := db.DeleteSubscriptionParams{Endpoint: sub.Endpoint, UserID: userID}
				if delErr := s.db.DeleteSubscription(ctx, params); delErr != nil {
					s.logger.ErrorContext(ctx, "failed to remove expired subscription", "user_id", userID, "endpoint", sub.Endpoint, "error", delErr)
				} else {
					s.logger.InfoContext(ctx, "removed expired push subscription", "user_id", userID, "endpoint", sub.Endpoint)
				}
			}
			continue // Continue to next sub, don't fail all
		}
		s.logger.InfoContext(ctx, "web push notification sent successfully", "user_id", userID, "endpoint", sub.Endpoint, "status_code", resp.StatusCode)
	}

	if lastErr != nil {
		return fmt.Errorf("failed to send to at least one subscription: %w", lastErr)
	}
	return nil
}
