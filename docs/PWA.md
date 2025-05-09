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
STATIC_DIR=./frontend/dist
```

Generate VAPID keys once:

```bash
npx web-push generate-vapid-keys --json
```

---

## 3  Database migration (`internal/db/migrations/20250509_push.sql`)

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

Run:

```bash
migrate -database "sqlite3://$(grep DATABASE_PATH .env | cut -d '=' -f2)" -path internal/db/migrations up
sqlc generate
```

---

## 4  `sqlc` queries (`internal/db/queries/push.sql`)

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

---

## 5  HTTP handlers (`internal/api/push_handlers.go`)

```go
// swagger:route POST /push/subscribe push subscribePush
// Store or update the caller's Web‑Push subscription.
// responses:
//   200: OK
func (a *API) SubscribePush(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Endpoint   string `json:"endpoint" validate:"required"`
        P256dhKey   string `json:"p256dh_key" validate:"required"`
        AuthKey     string `json:"auth_key" validate:"required"`
        UserAgent   string `json:"user_agent"`
        Platform    string `json:"platform"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        api.RespondErr(w, http.StatusBadRequest, "invalid body")
        return
    }
    userID := auth.UserIDFromCtx(r.Context())
    if err := a.DB.UpsertSubscription(r.Context(), db.UpsertSubscriptionParams{
        UserID:    userID,
        Endpoint:  req.Endpoint,
        P256dhKey: req.P256dhKey,
        AuthKey:   req.AuthKey,
        UserAgent: sql.NullString{String: req.UserAgent, Valid: req.UserAgent != ""},
        Platform:  sql.NullString{String: req.Platform,  Valid: req.Platform  != ""},
    }); err != nil {
        api.RespondErr(w, 500, "db error")
        return
    }
    w.WriteHeader(http.StatusOK)
}

// swagger:route DELETE /push/subscribe/{endpoint} push unsubscribePush
func (a *API) UnsubscribePush(w http.ResponseWriter, r *http.Request) {
    endpoint := chi.URLParam(r, "endpoint")
    userID := auth.UserIDFromCtx(r.Context())
    a.DB.DeleteSubscription(r.Context(), endpoint, userID)
    w.WriteHeader(http.StatusNoContent)
}

// swagger:route GET /push/vapid-public push getVAPID
func (a *API) VAPIDPublicKey(w http.ResponseWriter, r *http.Request) {
    api.RespondJSON(w, 200, map[string]string{"key": a.cfg.VAPIDPublic})
}
```

All three routes mount under authenticated group except the public‑key one.

---

## 6  Static file server (`cmd/server/main.go`)

```go
// after API routes are mounted
fs := http.FileServer(http.Dir(cfg.StaticDir))
r.Handle("/static/*", http.StripPrefix("/static", fs))

// SPA fallback
r.NotFound(func(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, filepath.Join(cfg.StaticDir, "index.html"))
})

// MIME tweak for webmanifest
mime.AddExtensionType(".webmanifest", "application/manifest+json")
```

`frontend/dist` should contain `index.html`, `assets/*`, `service-worker.js`.

---

## 7  Web‑Push sender (`internal/service/push.go`)

```go
import (
    webpush "github.com/SherClockHolmes/webpush-go"
)

type PushSender struct {
    db     db.Querier
    cfg    *config.Config
    logger *slog.Logger
}

func (s *PushSender) Send(ctx context.Context, userID int64, payload []byte, ttl int) {
    subs, _ := s.db.GetSubscriptionsByUser(ctx, userID)
    for _, sub := range subs {
        subscription := &webpush.Subscription{
            Endpoint: sub.Endpoint,
            Keys: webpush.Keys{P256dh: sub.P256dhKey, Auth: sub.AuthKey},
        }
        _, err := webpush.SendNotification(payload, subscription, &webpush.Options{
            VapIDPublicKey:  s.cfg.VAPIDPublic,
            VapIDPrivateKey: s.cfg.VAPIDPrivate,
            TTL:             ttl,
            Subscriber:      s.cfg.VAPIDSubject,
        })
        if err != nil {
            s.logger.Warn("push failed", "endpoint", sub.Endpoint, "err", err)
        }
    }
}
```

---

## 8  Outbox integration (`internal/service/dispatcher.go`)

```go
switch n.Channel {
case "sms":
    s.smsSender.Send(...)
case "push":
    s.pushSender.Send(ctx, n.UserID, []byte(n.Payload), 60)
}
```

Define `n.Payload` as compact JSON the SW can parse: `{ "type":"shift_reminder","booking_id":123 }`.

---

## 9  Reminder scheduling (`internal/service/reminders.go`)

```go
func (s *Scheduler) EnqueueShiftReminders(ctx context.Context, b db.Booking) error {
    // 24‑hour
    s.outbox.Enqueue(ctx, model.Notification{
        UserID:  b.UserID,
        Channel: "push",
        SendAt:  b.ShiftStart.Add(-24 * time.Hour),
        Payload: fmt.Sprintf(`{"type":"shift_reminder","hours":24,"start":"%s"}`, b.ShiftStart.Format(time.RFC3339)),
    })
    // 1‑hour
    s.outbox.Enqueue(ctx, model.Notification{ ...chem changed... })
    return nil
}
```

Call `EnqueueShiftReminders` from the booking‑creation flow.

---

## 10  SMS abstraction (kept mock for now)

```go
// internal/service/sms.go
//type SMSSender interface { Send(to, template, payload string) error }
// NewMockSMSSender returns a console‑logger implementation.
```

Switching to Twilio later means writing `twilioSMSSender` and wiring via config.

---

## 11  Swagger updates (`docs/swagger/*`)

Add **push** tag, three paths, and `PushSubscription` schema:

```yaml
PushSubscription:
  type: object
  required: [endpoint,p256dh_key,auth_key]
  properties:
    endpoint:   { type: string }
    p256dh_key: { type: string }
    auth_key:   { type: string }
    user_agent: { type: string }
    platform:   { type: string }
```

Run:

```bash
swag init -g cmd/server/main.go -o ./docs/swagger
```

---

## 12  Test checklist

* [ ] Migration applies & `sqlc` code compiles.
* [ ] `POST /push/subscribe` stores row.
* [ ] `GET /push/vapid-public` returns key.
* [ ] Booking creation enqueues push rows at expected times.
* [ ] Dispatcher sends `webpush` to test subscription (use [https://web-push-codelab.glitch.me/](https://web-push-codelab.glitch.me/)).
* [ ] Static `/` serves SPA and deep links work (404 fallback).

