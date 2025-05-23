package api

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"night-owls-go/internal/config"
	db "night-owls-go/internal/db/sqlc_generated"

	"github.com/go-chi/chi/v5"
)

// PushHandler handles push notification related HTTP requests.
type PushHandler struct {
	DB     db.Querier // Or *sql.DB if direct access is preferred and Querier is not sufficient
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

// swagger:route POST /push/subscribe push subscribePush
// Store or update the caller's Web-Push subscription.
// responses:
//   200: OK
//   400: Bad Request
//   500: Internal Server Error
func (h *PushHandler) SubscribePush(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Endpoint  string `json:"endpoint" validate:"required"`
		P256dhKey string `json:"p256dh_key" validate:"required"`
		AuthKey   string `json:"auth_key" validate:"required"`
		UserAgent string `json:"user_agent"`
		Platform  string `json:"platform"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid body", h.Logger, "error", err)
		return
	}

	userIDVal := r.Context().Value(UserIDKey) // Use UserIDKey from the api package (middleware.go)
	userID, ok := userIDVal.(int64)
	if !ok {
		// This should ideally be caught by AuthMiddleware, but double check.
		RespondWithError(w, http.StatusUnauthorized, "unauthorized - user ID not in context or invalid type", h.Logger)
		return
	}

	params := db.UpsertSubscriptionParams{
		UserID:    userID,
		Endpoint:  req.Endpoint,
		P256dhKey: req.P256dhKey,
		AuthKey:   req.AuthKey,
		UserAgent: sql.NullString{String: req.UserAgent, Valid: req.UserAgent != ""},
		Platform:  sql.NullString{String: req.Platform, Valid: req.Platform != ""},
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
//   500: Internal Server Error
func (h *PushHandler) UnsubscribePush(w http.ResponseWriter, r *http.Request) {
	// Try multiple methods to extract the endpoint parameter
	endpoint := chi.URLParam(r, "endpoint")
	h.Logger.InfoContext(r.Context(), "UnsubscribePush called", "endpoint_param", endpoint, "url", r.URL.Path)
	
	// Alternative method: Parse from URL path directly if chi.URLParam fails
	if endpoint == "" {
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) >= 3 && pathParts[0] == "push" && pathParts[1] == "subscribe" {
			endpoint = pathParts[2]
			h.Logger.InfoContext(r.Context(), "Extracted endpoint from path manually", "endpoint_param", endpoint)
		}
	}
	
	// Alternative method 2: Check request context for route values
	if endpoint == "" {
		if rctx := chi.RouteContext(r.Context()); rctx != nil {
			for i, param := range rctx.URLParams.Keys {
				if param == "endpoint" && i < len(rctx.URLParams.Values) {
					endpoint = rctx.URLParams.Values[i]
					h.Logger.InfoContext(r.Context(), "Found endpoint in route context", "endpoint_param", endpoint)
					break
				}
			}
		}
	}

	userIDVal := r.Context().Value(UserIDKey) // Use UserIDKey from the api package (middleware.go)
	userID, ok := userIDVal.(int64)
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "unauthorized - user ID not in context or invalid type", h.Logger)
		return
	}

	params := db.DeleteSubscriptionParams{
		Endpoint: endpoint,
		UserID:   userID,
	}

	if err := h.DB.DeleteSubscription(r.Context(), params); err != nil {
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
func (h *PushHandler) VAPIDPublicKey(w http.ResponseWriter, r *http.Request) {
	if h.Config.VAPIDPublic == "" {
		h.Logger.WarnContext(r.Context(), "VAPID public key is not configured")
		RespondWithError(w, http.StatusInternalServerError, "VAPID public key not configured", h.Logger)
		return
	}
	RespondWithJSON(w, http.StatusOK, map[string]string{"key": h.Config.VAPIDPublic}, h.Logger)
} 