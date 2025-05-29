# Night Owls Go — Backend Additions for PWA Support (Single‑Preference Model)

> One file, top‑to‑bottom, covering every change you need in the Go service to power the static Svelte PWA with **pure Web‑Push** and existing mock‑SMS OTP delivery. Copy‑paste friendly; trim what you don’t need.

---

## 1  Objectives

| Goal        | Detail                                                                 |
| ----------- | ---------------------------------------------------------------------- |
| Serve SPA   | Host `frontend/dist` from the same binary (or volume).                 |
| Web‑Push    | Store **one** set of subscription keys per user; vendor‑neutral VAPID. |
| Reminders   | Enqueue push rows at −24 h and −1 h for each new booking.              |
| Back‑compat | Keep mock SMS for OTP; swap sender later without touching the API.     |

---

## 2  Config additions (`config.go` & `.env`)

```go
// config.go (add fields)
VAPIDPublic  string `env:"VAPID_PUBLIC"`
VAPIDPrivate string `env:"VAPID_PRIVATE"`
VAPIDSubject string `env:"VAPID_SUBJECT" envDefault:"mailto:admin@example.com"`
StaticDir    string `env:"STATIC_DIR" envDefault:"./frontend/dist"`
```

```env
# .env sample
VAPID_PUBLIC=BNF…k
VAPID_PRIVATE=whU…w
VAPID_SUBJECT=mailto:admin@example.com # Optional
STATIC_DIR=./frontend/dist # Optional
```

Generate VAPID keys once:

```bash
npx web-push generate-vapid-keys --json
```

---

## 3  Database migrations

Two migrations are needed: one for `push_subscriptions` and another to add `user_id` to the `outbox` table (if not already present for other reasons, this is crucial for linking push notifications in the outbox to a user).

### 3.1 `push_subscriptions` table (`internal/db/migrations/YYYYMMDDHHMMSS_create_push_subscriptions.sql`)

```sql
-- +migrate Up
CREATE TABLE push_subscriptions (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id     INTEGER     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    endpoint    TEXT UNIQUE NOT NULL,
    p256dh_key  TEXT        NOT NULL,
    auth_key    TEXT        NOT NULL,
    user_agent  TEXT,
    platform    TEXT,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- +migrate Down
DROP TABLE push_subscriptions;
```

### 3.2 Add `user_id` to `outbox` table (`internal/db/migrations/YYYYMMDDHHMMSS_add_userid_to_outbox.sql`)

This column is needed to associate outbox entries for push notifications with a specific user.

```sql
-- +migrate Up
ALTER TABLE outbox ADD COLUMN user_id INTEGER REFERENCES users(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_outbox_user_id ON outbox(user_id);

-- +migrate Down
ALTER TABLE outbox DROP COLUMN user_id;
-- DROP INDEX IF EXISTS idx_outbox_user_id; -- Adjust if needed for your SQLite version
```

Run migrations:

```bash
migrate -database "sqlite3://$(grep DATABASE_PATH .env | cut -d '=' -f2)" -path internal/db/migrations up
sqlc generate
```
*(Ensure your `.env` file and `DATABASE_PATH` are correctly set up for the `migrate` command.)*

---

## 4  `sqlc` queries

### 4.1 Push subscription queries (`internal/db/queries/push.sql`)

```sql
-- name: UpsertSubscription :exec
INSERT INTO push_subscriptions (user_id, endpoint, p256dh_key, auth_key, user_agent, platform)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(endpoint) DO UPDATE
SET p256dh_key = excluded.p256dh_key,
    auth_key   = excluded.auth_key,
    user_agent = excluded.user_agent,
    platform   = excluded.platform;

-- name: DeleteSubscription :exec
DELETE FROM push_subscriptions WHERE endpoint = ? AND user_id = ?;

-- name: GetSubscriptionsByUser :many
SELECT endpoint, p256dh_key, auth_key FROM push_subscriptions WHERE user_id = ?;
```

### 4.2 Outbox query update (`internal/db/queries/outbox.sql`)

The `CreateOutboxItem` query needs to be updated to include `user_id`.

```sql
-- name: CreateOutboxItem :one
INSERT INTO outbox (
    message_type,
    recipient,
    payload,
    user_id -- Added user_id
) VALUES (
    ?,
    ?,
    ?,
    ? -- Added user_id parameter
)
RETURNING *;

-- Other outbox queries (GetPendingOutboxItems, UpdateOutboxItemStatus) remain largely the same,
-- but GetPendingOutboxItems will now also return the user_id column.
```

After modifying SQL queries, always run `sqlc generate`.

---

## 5  HTTP handlers (`internal/api/push_handlers.go`)

These handlers will be methods on a `PushHandler` struct, which is instantiated with dependencies like `db.Querier`, `*config.Config`, and `*slog.Logger`.

```go
package api

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"

	"night-owls-go/internal/config" // Adjusted path
	db "night-owls-go/internal/db/sqlc_generated" // Adjusted path

	"github.com/go-chi/chi/v5"
)

// PushHandler handles push notification related HTTP requests.
type PushHandler struct {
	DB     db.Querier
	Config *config.Config
	Logger *slog.Logger
}

// NewPushHandler creates a new PushHandler.
func NewPushHandler(querier db.Querier, cfg *config.Config, logger *slog.Logger) *PushHandler {
	return &PushHandler{
		DB:     querier,
		Config: cfg,
		Logger: logger.With("handler", "PushHandler"),
	}
}

// PushSubscriptionRequest defines the structure for the subscription request body.
// Used for request decoding and Swagger schema generation.
type PushSubscriptionRequest struct {
	Endpoint  string `json:"endpoint" validate:"required"`
	P256dhKey string `json:"p256dh_key" validate:"required"`
	AuthKey   string `json:"auth_key" validate:"required"`
	UserAgent string `json:"user_agent,omitempty"`
	Platform  string `json:"platform,omitempty"`
}

// swagger:route POST /push/subscribe push subscribePush
// Store or update the caller's Web-Push subscription.
// responses:
//   200: OK
//   400: Bad Request
//   401: Unauthorized
//   500: Internal Server Error
func (h *PushHandler) SubscribePush(w http.ResponseWriter, r *http.Request) {
    var req PushSubscriptionRequest // Using defined struct
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        RespondWithError(w, http.StatusBadRequest, "invalid body", h.Logger, "error", err)
        return
    }

    userIDVal := r.Context().Value(UserIDKey) // UserIDKey from api package
    userID, ok := userIDVal.(int64)
    if !ok {
        RespondWithError(w, http.StatusUnauthorized, "unauthorized - user ID not in context", h.Logger)
        return
    }

    params := db.UpsertSubscriptionParams{
        UserID:    userID,
        Endpoint:  req.Endpoint,
        P256dhKey: req.P256dhKey,
        AuthKey:   req.AuthKey,
        UserAgent: sql.NullString{String: req.UserAgent, Valid: req.UserAgent != ""},
        Platform:  sql.NullString{String: req.Platform,  Valid: req.Platform  != ""},
    }
    if err := h.DB.UpsertSubscription(r.Context(), params); err != nil {
        h.Logger.ErrorContext(r.Context(), "failed to upsert subscription", "error", err, "user_id", userID, "endpoint", req.Endpoint)
        RespondWithError(w, http.StatusInternalServerError, "db error", h.Logger, "error", err)
        return
    }
    w.WriteHeader(http.StatusOK)
}

// swagger:route DELETE /push/subscribe/{endpoint} push unsubscribePush
// Unsubscribes a push notification endpoint.
// responses:
//   204: No Content
//   401: Unauthorized
//   500: Internal Server Error
func (h *PushHandler) UnsubscribePush(w http.ResponseWriter, r *http.Request) {
    endpoint := chi.URLParam(r, "endpoint")
    userIDVal := r.Context().Value(UserIDKey) // UserIDKey from api package
    userID, ok := userIDVal.(int64)
    if !ok {
        RespondWithError(w, http.StatusUnauthorized, "unauthorized - user ID not in context", h.Logger)
        return
    }

    deleteParams := db.DeleteSubscriptionParams{ // Ensure this param struct is correct
        Endpoint: endpoint,
        UserID:   userID,
    }
    if err := h.DB.DeleteSubscription(r.Context(), deleteParams); err != nil { // Pass params struct
        h.Logger.ErrorContext(r.Context(), "failed to delete subscription", "error", err, "user_id", userID, "endpoint", endpoint)
        RespondWithError(w, http.StatusInternalServerError, "db error", h.Logger, "error", err)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}

// swagger:route GET /push/vapid-public push getVAPID
// Returns the VAPID public key.
// responses:
//   200: OK
//   500: Internal Server Error (if key not configured)
func (h *PushHandler) VAPIDPublicKey(w http.ResponseWriter, r *http.Request) {
    if h.Config.VAPIDPublic == "" {
        h.Logger.ErrorContext(r.Context(), "VAPID public key is not configured")
        RespondWithError(w, http.StatusInternalServerError, "VAPID public key not configured", h.Logger)
        return
    }
    RespondWithJSON(w, http.StatusOK, map[string]string{"key": h.Config.VAPIDPublic}, h.Logger)
}

// Ensure UserIDKey is defined in the api package (e.g., in middleware.go)
// type ContextKey string
// const UserIDKey ContextKey = "userID"

// Ensure RespondWithError and RespondWithJSON are available in the api package (e.g., in jsonutils.go)
// func RespondWithError(w http.ResponseWriter, code int, message string, logger *slog.Logger, details ...any)
// func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}, logger *slog.Logger, details ...any)
```

All three routes mount under an authenticated group in `cmd/server/main.go`, except the `/push/vapid-public` key route which is public. The `PushHandler` is instantiated in `main.go` and its methods are registered as route handlers.

---

## 6  Static file server (`cmd/server/main.go`)

```go
// In main.go, after other initializations (config, logger, db, services):
// ...
// querier := db.New(dbConn)
// cfg := // ... load config
// logger := // ... create logger
// pushAPIHandler := api.NewPushHandler(querier, cfg, logger)
// ...

// In router setup, after API routes are mounted:

// Public push route
// router.Get("/push/vapid-public", pushAPIHandler.VAPIDPublicKey)

// Authenticated push routes
// router.Group(func(r chi.Router) {
//    r.Use(api.AuthMiddleware(cfg, logger))
//    r.Post("/push/subscribe", pushAPIHandler.SubscribePush)
//    r.Delete("/push/subscribe/{endpoint}", pushAPIHandler.UnsubscribePush)
// })

// Static file serving
fs := http.FileServer(http.Dir(cfg.StaticDir))
router.Handle("/static/*", http.StripPrefix("/static/", fs)) // Note: ensure prefix ends with / if static files are in root of StaticDir

// SPA fallback for client-side routing
router.NotFound(func(w http.ResponseWriter, r *http.Request) {
    // Exclude API paths from SPA fallback if necessary, e.g. by checking r.URL.Path
    // if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/swagger/") {
    //    http.NotFound(w, r)
    //    return
    // }
    http.ServeFile(w, r, filepath.Join(cfg.StaticDir, "index.html"))
})

// MIME tweak for webmanifest (add once, e.g., during main setup)
// import "mime"
mime.AddExtensionType(".webmanifest", "application/manifest+json")
```

`frontend/dist` should contain `index.html`, `assets/*`, `service-worker.js`.

---

## 7  Web‑Push sender (`internal/service/push.go`)

```go
package service

import (
	"context"
	"log/slog"

	"night-owls-go/internal/config" // Adjusted path
	db "night-owls-go/internal/db/sqlc_generated" // Adjusted path

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
func (s *PushSender) Send(ctx context.Context, userID int64, payload []byte, ttl int) {
	subs, err := s.db.GetSubscriptionsByUser(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get subscriptions by user", "user_id", userID, "error", err)
		return
	}
    if len(subs) == 0 {
        s.logger.InfoContext(ctx, "no subscriptions found for user", "user_id", userID)
        return
    }

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
		})
		if err != nil {
			s.logger.WarnContext(ctx, "push failed, consider unsubscribing", "endpoint", sub.Endpoint, "user_id", userID, "error", err)
            // TODO: Handle specific errors like 404 or 410 to remove dead subscriptions.
		} else {
            if resp != nil && resp.Body != nil {
                _ = resp.Body.Close() // Important to close response body
            }
            s.logger.InfoContext(ctx, "push sent successfully", "endpoint", sub.Endpoint, "user_id", userID, "status_code", resp.StatusCode)
        }
	}
}
```
The `PushSender` is instantiated in `cmd/server/main.go` and passed to the `DispatcherService`.

---

## 8  Outbox integration (`internal/outbox/dispatcher.go`)

The `DispatcherService` needs to be updated to handle "push" `MessageType` entries from the outbox.

```go
// In internal/outbox/dispatcher.go

// Update DispatcherService struct:
type DispatcherService struct {
	querier    db.Querier
	smsSender  MessageSender // Assuming this is your existing SMS sender interface/type
	pushSender *service.PushSender // Add PushSender
	logger     *slog.Logger
	cfg        *config.Config
}

// Update NewDispatcherService constructor:
func NewDispatcherService(
    querier db.Querier, 
    smsSender MessageSender, 
    pushSender *service.PushSender, // Add PushSender
    logger *slog.Logger, 
    cfg *config.Config,
) *DispatcherService {
	return &DispatcherService{
		querier:    querier,
		smsSender:  smsSender,
		pushSender: pushSender, // Initialize PushSender
		logger:     logger.With("service", "OutboxDispatcher"),
		cfg:        cfg,
	}
}

// In ProcessPendingOutboxMessages method:
// ...
for _, item := range pendingItems {
    sendCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
    var dispatchErr error

    switch item.MessageType { // Assuming MessageType serves as the "channel"
    case "sms":
        if s.smsSender != nil {
            dispatchErr = s.smsSender.Send(item.Recipient, item.MessageType, item.Payload.String)
        } else {
            dispatchErr = errors.New("smsSender not configured")
            s.logger.ErrorContext(sendCtx, dispatchErr.Error(), "outbox_id", item.OutboxID)
        }
    case "push":
        if s.pushSender != nil {
            if !item.UserID.Valid { // UserID is now available on item from GetPendingOutboxItems
                dispatchErr = errors.New("user_id is null for push notification")
                s.logger.ErrorContext(sendCtx, dispatchErr.Error(), "outbox_id", item.OutboxID)
            } else {
                // Payload is expected to be JSON string, e.g., {"type":"shift_reminder","booking_id":123}
                // TTL is set to 60s as per an example, adjust if needed.
                s.pushSender.Send(sendCtx, item.UserID.Int64, []byte(item.Payload.String), 60)
                // pushSender.Send logs its own errors internally.
                // If it returned an error, dispatchErr should be set.
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

    // Update outbox item status (failed, sent, permanently_failed)
    // ... (existing logic for updating status based on dispatchErr) ...
}
// ...
```

Define `item.Payload` as compact JSON the SW can parse: `{ "type":"shift_reminder","booking_id":123 }`.

---

## 9  Reminder scheduling (`internal/service/reminders.go`)

A `Scheduler` service can be responsible for creating outbox entries for reminders.

```go
package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	db "night-owls-go/internal/db/sqlc_generated" // Adjusted path
)

type Scheduler struct {
	querier db.Querier
	logger  *slog.Logger
}

func NewScheduler(querier db.Querier, logger *slog.Logger) *Scheduler {
	return &Scheduler{
		querier: querier,
		logger:  logger.With("service", "Scheduler"),
	}
}

func (s *Scheduler) EnqueueShiftReminders(ctx context.Context, booking db.Booking) error {
	// 24-hour reminder
	payload24h := fmt.Sprintf(`{"type":"shift_reminder","hours":24,"start_time":"%s","booking_id":%d}`,
		booking.ShiftStart.Format(time.RFC3339), booking.BookingID)
	
	params24h := db.CreateOutboxItemParams{
		MessageType: "push",
		Recipient:   "", // Not used for push; UserID is primary
		Payload:     sql.NullString{String: payload24h, Valid: true},
		UserID:      sql.NullInt64{Int64: booking.UserID, Valid: true},
        // SendAt: To be implemented (see note below)
	}
	_, err := s.querier.CreateOutboxItem(ctx, params24h)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to enqueue 24h shift reminder", "booking_id", booking.BookingID, "error", err)
		return fmt.Errorf("enqueue 24h reminder: %w", err)
	}
    s.logger.InfoContext(ctx, "Enqueued 24h shift reminder", "booking_id", booking.BookingID, "user_id", booking.UserID)

	// 1-hour reminder
	payload1h := fmt.Sprintf(`{"type":"shift_reminder","hours":1,"start_time":"%s","booking_id":%d}`,
		booking.ShiftStart.Format(time.RFC3339), booking.BookingID)

	params1h := db.CreateOutboxItemParams{
		MessageType: "push",
		Recipient:   "",
		Payload:     sql.NullString{String: payload1h, Valid: true},
		UserID:      sql.NullInt64{Int64: booking.UserID, Valid: true},
        // SendAt: To be implemented (see note below)
	}
	_, err = s.querier.CreateOutboxItem(ctx, params1h)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to enqueue 1h shift reminder", "booking_id", booking.BookingID, "error", err)
		return fmt.Errorf("enqueue 1h reminder: %w", err)
	}
    s.logger.InfoContext(ctx, "Enqueued 1h shift reminder", "booking_id", booking.BookingID, "user_id", booking.UserID)

    s.logger.WarnContext(ctx, "SendAt CAVEAT: Reminder scheduling currently does not use a 'SendAt' field in the outbox table. Reminders will be dispatched on the next outbox processing cycle (e.g., every minute). True scheduled delivery requires adding a 'send_at' timestamp column to the 'outbox' table, updating sqlc queries to use it for creating items and for fetching pending items (WHERE send_at <= NOW()), and setting this field in CreateOutboxItemParams above (e.g., with booking.ShiftStart.Add(-24 * time.Hour)). This will be addressed in a future iteration.")
	return nil
}
```

Call `EnqueueShiftReminders` from the booking‑creation flow (e.g., in `BookingService.CreateBooking`).
The `Scheduler` service would be instantiated in `cmd/server/main.go` and passed to `BookingService`.

---

## 10  SMS abstraction (kept mock for now)

```go
// internal/outbox/outbox.go (or similar, defines MessageSender interface)
// type MessageSender interface { Send(recipient, messageType, payload string) error }

// internal/outbox/sms_mock.go (or similar)
// NewLogFileMessageSender returns a console/file-logger implementation.
```

Switching to Twilio later means writing `twilioSMSSender` and wiring via config.

---

## 11  Swagger updates (`docs/swagger/*`)

Add **push** tag, three paths, and `PushSubscription` schema:

```yaml
# This is conceptual YAML. Actual schema generation is best handled by `swag init`
# after defining a Go struct for the request body.

# Example of defining PushSubscriptionRequest in internal/api/models.go:
# type PushSubscriptionRequest struct {
#    Endpoint  string `json:"endpoint" validate:"required"`
#    P256dhKey string `json:"p256dh_key" validate:"required"`
#    AuthKey   string `json:"auth_key" validate:"required"`
#    UserAgent string `json:"user_agent,omitempty"`
#    Platform  string `json:"platform,omitempty"`
# }
# Then use this struct in the SubscribePush handler. `swag init` will pick it up.

# Expected schema in generated swagger.yaml (under components/schemas or definitions):
PushSubscription: # Or PushSubscriptionRequest if named that way
  type: object
  required: [endpoint,p256dh_key,auth_key]
  properties:
    endpoint:   { type: string }
    p256dh_key: { type: string }
    auth_key:   { type: string }
    user_agent: { type: string }
    platform:   { type: string }
```

Run (after Go code changes, especially if adding/changing handler structs or comments):

```bash
swag init -g cmd/server/main.go -o ./docs/swagger
```

Verify `docs/swagger/swagger.yaml` (or `.json`) includes the "push" tag, the three new paths under it, and the `PushSubscription` (or `PushSubscriptionRequest`) schema definition.

---

## 12  Test checklist

* [X] Config additions for VAPID and StaticDir are in `config.go`.
* [X] Migration for `push_subscriptions` table created and applied.
* [X] Migration for adding `user_id` to `outbox` table created and applied.
* [X] `sqlc` queries for `push.sql` created.
* [X] `sqlc` query `CreateOutboxItem` in `outbox.sql` updated for `user_id`.
* [X] `sqlc generate` run successfully after schema and query changes.
* [X] HTTP handlers for push (`SubscribePush`, `UnsubscribePush`, `VAPIDPublicKey`) created in `internal/api/push_handlers.go` using `PushHandler`.
* [X] Static file server and SPA fallback configured in `cmd/server/main.go`.
* [X] `PushHandler` instantiated and routes mounted in `cmd/server/main.go`.
* [X] Web Push `PushSender` service created in `internal/service/push.go`.
* [X] `PushSender` integrated into `DispatcherService` in `internal/outbox/dispatcher.go`.
* [X] `DispatcherService` constructor updated and used in `cmd/server/main.go`.
* [X] Reminder scheduling `Scheduler` service with `EnqueueShiftReminders` created in `internal/service/reminders.go`. (Caveat: `SendAt` functionality for precise timing is TBD).
* [X] `swag init` run to update Swagger documentation.
* [ ] Test: `POST /push/subscribe` stores row correctly.
* [ ] Test: `GET /push/vapid-public` returns VAPID public key.
* [ ] Test: Booking creation enqueues push rows in `outbox` table with correct `user_id` and `message_type="push"`. (Note: They will be dispatched on next cron cycle, not delayed by `SendAt` yet).
* [ ] Test: Dispatcher sends `webpush` to a test subscription (e.g., using [https://web-push-codelab.glitch.me/](https://web-push-codelab.glitch.me/)).
* [ ] Test: Static `/` serves SPA, and deep links (e.g., `/some/spa/route`) fall back to `index.html`.
* [ ] Test: `.webmanifest` file is served with `application/manifest+json` MIME type.

