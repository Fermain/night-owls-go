package api

import (
	"log/slog"
	"net/http"
	"time"

	db "night-owls-go/internal/db/sqlc_generated"
)

// BroadcastHandler handles user-facing broadcast operations.
type BroadcastHandler struct {
	querier db.Querier
	logger  *slog.Logger
}

// NewBroadcastHandler creates a new BroadcastHandler.
func NewBroadcastHandler(querier db.Querier, logger *slog.Logger) *BroadcastHandler {
	return &BroadcastHandler{
		querier: querier,
		logger:  logger.With("handler", "BroadcastHandler"),
	}
}

// UserBroadcastResponse represents a broadcast for user consumption.
type UserBroadcastResponse struct {
	ID        int64     `json:"id"`
	Message   string    `json:"message"`
	Audience  string    `json:"audience"`
	CreatedAt time.Time `json:"created_at"`
}

// ListUserBroadcasts handles GET /api/broadcasts
// Returns broadcasts that the authenticated user should see
func (h *BroadcastHandler) ListUserBroadcasts(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by auth middleware)
	userIDVal := r.Context().Value(UserIDKey)
	userID, ok := userIDVal.(int64)
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User ID not found in context", h.logger)
		return
	}

	// Get user details to determine role
	user, err := h.querier.GetUserByID(r.Context(), userID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get user details", h.logger, "error", err)
		return
	}

	// Get all broadcasts
	broadcasts, err := h.querier.ListBroadcasts(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to list broadcasts", h.logger, "error", err)
		return
	}

	// Filter broadcasts based on user role and audience
	var userBroadcasts []UserBroadcastResponse
	for _, broadcast := range broadcasts {
		if h.shouldUserSeeBroadcast(user, broadcast) {
			// Handle sql.NullTime properly
			var createdAt time.Time
			if broadcast.CreatedAt.Valid {
				createdAt = broadcast.CreatedAt.Time
			} else {
				createdAt = time.Now() // Fallback to current time if null
			}
			
			userBroadcasts = append(userBroadcasts, UserBroadcastResponse{
				ID:        broadcast.BroadcastID,
				Message:   broadcast.Message,
				Audience:  broadcast.Audience,
				CreatedAt: createdAt,
			})
		}
	}

	RespondWithJSON(w, http.StatusOK, userBroadcasts, h.logger)
}

// shouldUserSeeBroadcast determines if a user should see a particular broadcast
func (h *BroadcastHandler) shouldUserSeeBroadcast(user db.User, broadcast db.Broadcast) bool {
	// Only show sent broadcasts
	if broadcast.Status != "sent" {
		return false
	}

	switch broadcast.Audience {
	case "all":
		return true
	case "admins":
		return user.Role == "admin"
	case "owls":
		return user.Role == "owl" || user.Role == ""
	case "active":
		// For now, show to all users. In future, check last activity
		return true
	default:
		return false
	}
}