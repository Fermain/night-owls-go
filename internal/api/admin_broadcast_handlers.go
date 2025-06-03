package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	db "night-owls-go/internal/db/sqlc_generated"

	"github.com/go-chi/chi/v5"
)

// AdminBroadcastHandler handles admin broadcast operations.
type AdminBroadcastHandler struct {
	querier db.Querier
	logger  *slog.Logger
}

// NewAdminBroadcastHandler creates a new AdminBroadcastHandler.
func NewAdminBroadcastHandler(querier db.Querier, logger *slog.Logger) *AdminBroadcastHandler {
	return &AdminBroadcastHandler{
		querier: querier,
		logger:  logger.With("handler", "AdminBroadcastHandler"),
	}
}

// CreateBroadcastRequest defines the expected JSON body for creating a broadcast.
type CreateBroadcastRequest struct {
	Title       string     `json:"title"`
	Message     string     `json:"message"`
	Audience    string     `json:"audience"`
	PushEnabled bool       `json:"push_enabled"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
}

// BroadcastResponse represents a broadcast in API responses.
type BroadcastResponse struct {
	BroadcastID    int64      `json:"broadcast_id"`
	Title          string     `json:"title"`
	Message        string     `json:"message"`
	Audience       string     `json:"audience"`
	SenderUserID   int64      `json:"sender_user_id"`
	SenderName     string     `json:"sender_name,omitempty"`
	PushEnabled    bool       `json:"push_enabled"`
	ScheduledAt    *time.Time `json:"scheduled_at"`
	SentAt         *time.Time `json:"sent_at"`
	Status         string     `json:"status"`
	RecipientCount int64      `json:"recipient_count"`
	SentCount      int64      `json:"sent_count"`
	FailedCount    int64      `json:"failed_count"`
	CreatedAt      time.Time  `json:"created_at"`
}

// AdminCreateBroadcast handles POST /api/admin/broadcasts
func (h *AdminBroadcastHandler) AdminCreateBroadcast(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value(UserIDKey)
	userID, ok := userIDVal.(int64)
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User ID not found in context", h.logger)
		return
	}

	var req CreateBroadcastRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload", h.logger, "error", err)
		return
	}

	if req.Message == "" {
		RespondWithError(w, http.StatusBadRequest, "Message is required", h.logger)
		return
	}

	if req.Title == "" {
		RespondWithError(w, http.StatusBadRequest, "Title is required", h.logger)
		return
	}

	if req.Audience == "" {
		RespondWithError(w, http.StatusBadRequest, "Audience is required", h.logger)
		return
	}

	// Validate audience values
	validAudiences := map[string]bool{
		"all":    true,
		"admins": true,
		"owls":   true,
		"active": true,
	}
	if !validAudiences[req.Audience] {
		RespondWithError(w, http.StatusBadRequest, "Invalid audience", h.logger)
		return
	}

	// Calculate recipient count based on audience
	recipientCount, err := h.calculateRecipientCount(r.Context(), req.Audience)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to calculate recipient count", h.logger, "error", err)
		return
	}

	var scheduledAt sql.NullTime
	if req.ScheduledAt != nil {
		scheduledAt = sql.NullTime{Time: *req.ScheduledAt, Valid: true}
	}

	params := db.CreateBroadcastParams{
		Title:          req.Title,
		Message:        req.Message,
		Audience:       req.Audience,
		SenderUserID:   userID,
		PushEnabled:    req.PushEnabled,
		ScheduledAt:    scheduledAt,
		RecipientCount: sql.NullInt64{Int64: recipientCount, Valid: true},
	}

	broadcast, err := h.querier.CreateBroadcast(r.Context(), params)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create broadcast", h.logger, "error", err)
		return
	}

	response := BroadcastResponse{
		BroadcastID:    broadcast.BroadcastID,
		Title:          broadcast.Title,
		Message:        broadcast.Message,
		Audience:       broadcast.Audience,
		SenderUserID:   broadcast.SenderUserID,
		PushEnabled:    broadcast.PushEnabled,
		ScheduledAt:    nullTimeToPointer(broadcast.ScheduledAt),
		SentAt:         nullTimeToPointer(broadcast.SentAt),
		Status:         broadcast.Status,
		RecipientCount: nullInt64ToInt64(broadcast.RecipientCount),
		SentCount:      nullInt64ToInt64(broadcast.SentCount),
		FailedCount:    nullInt64ToInt64(broadcast.FailedCount),
		CreatedAt:      broadcast.CreatedAt.Time,
	}

	RespondWithJSON(w, http.StatusCreated, response, h.logger)
}

// AdminListBroadcasts handles GET /api/admin/broadcasts
func (h *AdminBroadcastHandler) AdminListBroadcasts(w http.ResponseWriter, r *http.Request) {
	broadcasts, err := h.querier.ListBroadcastsWithSender(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to list broadcasts", h.logger, "error", err)
		return
	}

	var response []BroadcastResponse
	for _, broadcast := range broadcasts {
		response = append(response, BroadcastResponse{
			BroadcastID:    broadcast.BroadcastID,
			Title:          broadcast.Title,
			Message:        broadcast.Message,
			Audience:       broadcast.Audience,
			SenderUserID:   broadcast.SenderUserID,
			SenderName:     broadcast.SenderName,
			PushEnabled:    broadcast.PushEnabled,
			ScheduledAt:    nullTimeToPointer(broadcast.ScheduledAt),
			SentAt:         nullTimeToPointer(broadcast.SentAt),
			Status:         broadcast.Status,
			RecipientCount: nullInt64ToInt64(broadcast.RecipientCount),
			SentCount:      nullInt64ToInt64(broadcast.SentCount),
			FailedCount:    nullInt64ToInt64(broadcast.FailedCount),
			CreatedAt:      broadcast.CreatedAt.Time,
		})
	}

	RespondWithJSON(w, http.StatusOK, response, h.logger)
}

// AdminGetBroadcast handles GET /api/admin/broadcasts/{id}
func (h *AdminBroadcastHandler) AdminGetBroadcast(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid broadcast ID", h.logger, "id", idStr)
		return
	}

	broadcast, err := h.querier.GetBroadcastByID(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondWithError(w, http.StatusNotFound, "Broadcast not found", h.logger, "id", id)
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to get broadcast", h.logger, "error", err)
		}
		return
	}

	response := BroadcastResponse{
		BroadcastID:    broadcast.BroadcastID,
		Title:          broadcast.Title,
		Message:        broadcast.Message,
		Audience:       broadcast.Audience,
		SenderUserID:   broadcast.SenderUserID,
		PushEnabled:    broadcast.PushEnabled,
		ScheduledAt:    nullTimeToPointer(broadcast.ScheduledAt),
		SentAt:         nullTimeToPointer(broadcast.SentAt),
		Status:         broadcast.Status,
		RecipientCount: nullInt64ToInt64(broadcast.RecipientCount),
		SentCount:      nullInt64ToInt64(broadcast.SentCount),
		FailedCount:    nullInt64ToInt64(broadcast.FailedCount),
		CreatedAt:      broadcast.CreatedAt.Time,
	}

	RespondWithJSON(w, http.StatusOK, response, h.logger)
}

// calculateRecipientCount calculates how many users would receive the broadcast based on audience
func (h *AdminBroadcastHandler) calculateRecipientCount(ctx context.Context, audience string) (int64, error) {
	users, err := h.querier.ListUsers(ctx, nil)
	if err != nil {
		return 0, err
	}

	switch audience {
	case "all":
		return int64(len(users)), nil
	case "admins":
		count := int64(0)
		for _, user := range users {
			if user.Role == "admin" {
				count++
			}
		}
		return count, nil
	case "owls":
		count := int64(0)
		for _, user := range users {
			if user.Role == "owl" || user.Role == "" {
				count++
			}
		}
		return count, nil
	case "active":
		// For now, return all users. In future, filter by last activity
		return int64(len(users)), nil
	default:
		return 0, nil
	}
}

// nullTimeToPointer converts sql.NullTime to *time.Time
func nullTimeToPointer(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

// nullInt64ToInt64 converts sql.NullInt64 to int64
func nullInt64ToInt64(ni sql.NullInt64) int64 {
	if ni.Valid {
		return ni.Int64
	}
	return 0
}
