package service

import (
	"context"
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
// payload should be a byte slice representing the JSON payload for the push notification.
// ttl is the time-to-live in seconds for the push message.
func (s *PushSender) Send(ctx context.Context, userID int64, payload []byte, ttl int) {
	subs, err := s.db.GetSubscriptionsByUser(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get subscriptions by user", "user_id", userID, "error", err)
		return
	}

	if len(subs) == 0 {
		s.logger.InfoContext(ctx, "no subscriptions found for user, skipping push send", "user_id", userID)
		return
	}

	for _, sub := range subs {
		// Construct the webpush.Subscription object
		subscription := &webpush.Subscription{
			Endpoint: sub.Endpoint,
			Keys:     webpush.Keys{P256dh: sub.P256dhKey, Auth: sub.AuthKey},
		}

		// Send the notification
		// The webpush library handles sending to different push services based on the endpoint.
		resp, err := webpush.SendNotification(payload, subscription, &webpush.Options{
			VAPIDPublicKey:  s.config.VAPIDPublic,
			VAPIDPrivateKey: s.config.VAPIDPrivate,
			TTL:             ttl,
			Subscriber:      s.config.VAPIDSubject, // Typically an email address or URL
		})
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to send web push notification", "user_id", userID, "endpoint", sub.Endpoint, "error", err)
			// Consider if we should remove the subscription if it's permanently failed (e.g., 404, 410)
			// For now, just log the error.
		} else {
			// Closing the response body is important to avoid resource leaks
			if resp != nil && resp.Body != nil {
				_ = resp.Body.Close()
			}
			s.logger.InfoContext(ctx, "web push notification sent successfully", "user_id", userID, "endpoint", sub.Endpoint, "status_code", resp.StatusCode)
		}
	}
} 