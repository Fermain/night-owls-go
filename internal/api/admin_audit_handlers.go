package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	db "night-owls-go/internal/db/sqlc_generated"
	"night-owls-go/internal/service"
)

// AdminAuditHandler handles admin audit trail API requests
type AdminAuditHandler struct {
	auditService *service.AuditService
	querier      db.Querier
	logger       *slog.Logger
}

// NewAdminAuditHandler creates a new AdminAuditHandler
func NewAdminAuditHandler(auditService *service.AuditService, querier db.Querier, logger *slog.Logger) *AdminAuditHandler {
	return &AdminAuditHandler{
		auditService: auditService,
		querier:      querier,
		logger:       logger.With("handler", "AdminAuditHandler"),
	}
}

// AuditEventResponse represents an audit event for API responses
type AuditEventResponse struct {
	EventID      int64                  `json:"event_id"`
	EventType    string                 `json:"event_type"`
	ActorUserID  *int64                 `json:"actor_user_id"`
	ActorName    string                 `json:"actor_name"`
	ActorPhone   string                 `json:"actor_phone"`
	TargetUserID *int64                 `json:"target_user_id"`
	TargetName   string                 `json:"target_name"`
	TargetPhone  string                 `json:"target_phone"`
	EntityType   string                 `json:"entity_type"`
	EntityID     *int64                 `json:"entity_id"`
	Action       string                 `json:"action"`
	Details      map[string]interface{} `json:"details"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
	CreatedAt    string                 `json:"created_at"`
}

// AuditStatsResponse represents audit statistics
type AuditStatsResponse struct {
	TotalEvents      int64  `json:"total_events"`
	UniqueActors     int64  `json:"unique_actors"`
	UniqueEventTypes int64  `json:"unique_event_types"`
	EarliestEvent    string `json:"earliest_event"`
	LatestEvent      string `json:"latest_event"`
}

// EventTypeStatsResponse represents statistics by event type
type EventTypeStatsResponse struct {
	EventType   string `json:"event_type"`
	EventCount  int64  `json:"event_count"`
	LatestEvent string `json:"latest_event"`
}

// parseUserIDs parses comma-separated user IDs and returns them as int64 slice
func parseUserIDs(userIDStr string) ([]int64, error) {
	if userIDStr == "" {
		return nil, nil
	}

	parts := strings.Split(userIDStr, ",")
	userIDs := make([]int64, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}

		// Try to parse as number first
		if userID, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
			userIDs = append(userIDs, userID)
		} else {
			// If not a number, try to find user by phone
			// This would require a user lookup, for now skip invalid entries
			continue
		}
	}

	return userIDs, nil
}

// queryAuditEventsWithFilters executes a dynamic query with filters
func (h *AdminAuditHandler) queryAuditEventsWithFilters(ctx context.Context, eventType string, actorUserIDs, targetUserIDs []int64, limit, offset int64) ([]db.ListAuditEventsRow, error) {
	// For now, use a simpler approach with the existing SQLC queries
	// and handle multiple user IDs differently
	return h.queryWithMultipleFilters(ctx, eventType, actorUserIDs, targetUserIDs, limit, offset)
}

// queryWithMultipleFilters handles multiple user IDs by making multiple queries and merging results
func (h *AdminAuditHandler) queryWithMultipleFilters(ctx context.Context, eventType string, actorUserIDs, targetUserIDs []int64, limit, offset int64) ([]db.ListAuditEventsRow, error) {
	// For now, use the first user ID from each list and fall back to regular queries
	// This is a simplified implementation that works with the existing SQLC queries

	if eventType != "" && len(actorUserIDs) == 0 && len(targetUserIDs) == 0 {
		// Event type only
		dbEvents, err := h.querier.ListAuditEventsByType(ctx, db.ListAuditEventsByTypeParams{
			EventType: eventType,
			Limit:     limit,
			Offset:    offset,
		})
		if err != nil {
			return nil, err
		}

		// Convert to ListAuditEventsRow format
		var result []db.ListAuditEventsRow
		for _, event := range dbEvents {
			result = append(result, db.ListAuditEventsRow(event))
		}
		return result, nil
	}

	if len(actorUserIDs) == 1 && len(targetUserIDs) == 0 && eventType == "" {
		// Single actor only
		dbEvents, err := h.querier.ListAuditEventsByActor(ctx, db.ListAuditEventsByActorParams{
			ActorUserID: sql.NullInt64{Int64: actorUserIDs[0], Valid: true},
			Limit:       limit,
			Offset:      offset,
		})
		if err != nil {
			return nil, err
		}

		// Convert to ListAuditEventsRow format
		var result []db.ListAuditEventsRow
		for _, event := range dbEvents {
			result = append(result, db.ListAuditEventsRow(event))
		}
		return result, nil
	}

	if len(targetUserIDs) == 1 && len(actorUserIDs) == 0 && eventType == "" {
		// Single target only
		dbEvents, err := h.querier.ListAuditEventsByTarget(ctx, db.ListAuditEventsByTargetParams{
			TargetUserID: sql.NullInt64{Int64: targetUserIDs[0], Valid: true},
			Limit:        limit,
			Offset:       offset,
		})
		if err != nil {
			return nil, err
		}

		// Convert to ListAuditEventsRow format
		var result []db.ListAuditEventsRow
		for _, event := range dbEvents {
			result = append(result, db.ListAuditEventsRow(event))
		}
		return result, nil
	}

	// For complex cases with multiple user IDs, fall back to all events
	// and filter in Go (not ideal for performance but works for now)
	dbEvents, err := h.querier.ListAuditEvents(ctx, db.ListAuditEventsParams{
		Limit:  limit * 2, // Get more events to account for filtering
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	// Filter the results in Go
	var filteredEvents []db.ListAuditEventsRow
	for _, event := range dbEvents {
		// Check event type filter
		if eventType != "" && event.EventType != eventType {
			continue
		}

		// Check actor user ID filter
		if len(actorUserIDs) > 0 {
			found := false
			if event.ActorUserID.Valid {
				for _, id := range actorUserIDs {
					if event.ActorUserID.Int64 == id {
						found = true
						break
					}
				}
			}
			if !found {
				continue
			}
		}

		// Check target user ID filter
		if len(targetUserIDs) > 0 {
			found := false
			if event.TargetUserID.Valid {
				for _, id := range targetUserIDs {
					if event.TargetUserID.Int64 == id {
						found = true
						break
					}
				}
			}
			if !found {
				continue
			}
		}

		filteredEvents = append(filteredEvents, event)

		// Limit results
		if int64(len(filteredEvents)) >= limit {
			break
		}
	}

	return filteredEvents, nil
}

// AdminListAuditEvents handles GET /api/admin/audit-events
// @Summary List audit events (Admin)
// @Description Get a paginated list of audit events with optional filtering. Supports comma-separated user IDs.
// @Tags admin/audit
// @Produce json
// @Param event_type query string false "Filter by event type"
// @Param actor_user_id query string false "Filter by actor user ID(s) - comma-separated"
// @Param target_user_id query string false "Filter by target user ID(s) - comma-separated"
// @Param limit query int false "Number of events to return (default 50, max 100)"
// @Param offset query int false "Number of events to skip (default 0)"
// @Success 200 {array} AuditEventResponse "List of audit events"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/admin/audit-events [get]
func (h *AdminAuditHandler) AdminListAuditEvents(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	eventType := r.URL.Query().Get("event_type")
	actorUserIDStr := r.URL.Query().Get("actor_user_id")
	targetUserIDStr := r.URL.Query().Get("target_user_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Set defaults
	limit := int64(50)
	offset := int64(0)

	// Parse limit
	if limitStr != "" {
		if parsedLimit, err := strconv.ParseInt(limitStr, 10, 64); err == nil {
			if parsedLimit > 0 && parsedLimit <= 100 {
				limit = parsedLimit
			}
		}
	}

	// Parse offset
	if offsetStr != "" {
		if parsedOffset, err := strconv.ParseInt(offsetStr, 10, 64); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Parse user IDs
	actorUserIDs, err := parseUserIDs(actorUserIDStr)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to parse actor user IDs", "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid actor_user_id parameter", h.logger)
		return
	}

	targetUserIDs, err := parseUserIDs(targetUserIDStr)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to parse target user IDs", "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid target_user_id parameter", h.logger)
		return
	}

	// Query with filters
	var apiEvents []AuditEventResponse

	// Check if we need to use the flexible query or simple queries
	needsFlexibleQuery := len(actorUserIDs) > 1 || len(targetUserIDs) > 1 ||
		(eventType != "" && (len(actorUserIDs) > 0 || len(targetUserIDs) > 0))

	if needsFlexibleQuery {
		// Use dynamic query for complex filtering
		dbEvents, err := h.queryAuditEventsWithFilters(r.Context(), eventType, actorUserIDs, targetUserIDs, limit, offset)
		if err != nil {
			h.logger.ErrorContext(r.Context(), "Failed to query audit events with filters", "error", err)
			RespondWithError(w, http.StatusInternalServerError, "Failed to fetch audit events", h.logger)
			return
		}
		apiEvents = h.convertAuditEventsToAPI(dbEvents)
	} else {
		// Use existing optimized queries for simple cases
		if eventType != "" {
			dbEvents, err := h.querier.ListAuditEventsByType(r.Context(), db.ListAuditEventsByTypeParams{
				EventType: eventType,
				Limit:     limit,
				Offset:    offset,
			})
			if err != nil {
				h.logger.ErrorContext(r.Context(), "Failed to list audit events by type", "error", err)
				RespondWithError(w, http.StatusInternalServerError, "Failed to fetch audit events", h.logger)
				return
			}
			apiEvents = h.convertAuditEventsToAPI(dbEvents)
		} else if len(actorUserIDs) == 1 {
			dbEvents, err := h.querier.ListAuditEventsByActor(r.Context(), db.ListAuditEventsByActorParams{
				ActorUserID: sql.NullInt64{Int64: actorUserIDs[0], Valid: true},
				Limit:       limit,
				Offset:      offset,
			})
			if err != nil {
				h.logger.ErrorContext(r.Context(), "Failed to list audit events by actor", "error", err)
				RespondWithError(w, http.StatusInternalServerError, "Failed to fetch audit events", h.logger)
				return
			}
			apiEvents = h.convertAuditEventsToAPI(dbEvents)
		} else if len(targetUserIDs) == 1 {
			dbEvents, err := h.querier.ListAuditEventsByTarget(r.Context(), db.ListAuditEventsByTargetParams{
				TargetUserID: sql.NullInt64{Int64: targetUserIDs[0], Valid: true},
				Limit:        limit,
				Offset:       offset,
			})
			if err != nil {
				h.logger.ErrorContext(r.Context(), "Failed to list audit events by target", "error", err)
				RespondWithError(w, http.StatusInternalServerError, "Failed to fetch audit events", h.logger)
				return
			}
			apiEvents = h.convertAuditEventsToAPI(dbEvents)
		} else {
			dbEvents, err := h.querier.ListAuditEvents(r.Context(), db.ListAuditEventsParams{
				Limit:  limit,
				Offset: offset,
			})
			if err != nil {
				h.logger.ErrorContext(r.Context(), "Failed to list audit events", "error", err)
				RespondWithError(w, http.StatusInternalServerError, "Failed to fetch audit events", h.logger)
				return
			}
			apiEvents = h.convertAuditEventsToAPI(dbEvents)
		}
	}

	RespondWithJSON(w, http.StatusOK, apiEvents, h.logger)
}

// convertAuditEventsToAPI is a generic helper to convert any audit event row type to API response
func (h *AdminAuditHandler) convertAuditEventsToAPI(dbEvents interface{}) []AuditEventResponse {
	var apiEvents []AuditEventResponse

	// Use type assertions to handle different row types
	switch events := dbEvents.(type) {
	case []db.ListAuditEventsRow:
		apiEvents = make([]AuditEventResponse, 0, len(events))
		for _, event := range events {
			apiEvents = append(apiEvents, h.convertSingleEventToAPI(
				event.EventID, event.EventType, event.ActorUserID, event.ActorName, event.ActorPhone,
				event.TargetUserID, event.TargetName, event.TargetPhone, event.EntityType, event.EntityID,
				event.Action, event.Details, event.IpAddress, event.UserAgent, event.CreatedAt,
			))
		}
	case []db.ListAuditEventsByTypeRow:
		apiEvents = make([]AuditEventResponse, 0, len(events))
		for _, event := range events {
			apiEvents = append(apiEvents, h.convertSingleEventToAPI(
				event.EventID, event.EventType, event.ActorUserID, event.ActorName, event.ActorPhone,
				event.TargetUserID, event.TargetName, event.TargetPhone, event.EntityType, event.EntityID,
				event.Action, event.Details, event.IpAddress, event.UserAgent, event.CreatedAt,
			))
		}
	case []db.ListAuditEventsByActorRow:
		apiEvents = make([]AuditEventResponse, 0, len(events))
		for _, event := range events {
			apiEvents = append(apiEvents, h.convertSingleEventToAPI(
				event.EventID, event.EventType, event.ActorUserID, event.ActorName, event.ActorPhone,
				event.TargetUserID, event.TargetName, event.TargetPhone, event.EntityType, event.EntityID,
				event.Action, event.Details, event.IpAddress, event.UserAgent, event.CreatedAt,
			))
		}
	case []db.ListAuditEventsByTargetRow:
		apiEvents = make([]AuditEventResponse, 0, len(events))
		for _, event := range events {
			apiEvents = append(apiEvents, h.convertSingleEventToAPI(
				event.EventID, event.EventType, event.ActorUserID, event.ActorName, event.ActorPhone,
				event.TargetUserID, event.TargetName, event.TargetPhone, event.EntityType, event.EntityID,
				event.Action, event.Details, event.IpAddress, event.UserAgent, event.CreatedAt,
			))
		}
	}

	return apiEvents
}

// convertSingleEventToAPI converts a single audit event with all nullable fields to API response
func (h *AdminAuditHandler) convertSingleEventToAPI(
	eventID int64, eventType string, actorUserID sql.NullInt64, actorName string, actorPhone sql.NullString,
	targetUserID sql.NullInt64, targetName string, targetPhone sql.NullString, entityType string, entityID sql.NullInt64,
	action string, details sql.NullString, ipAddress, userAgent sql.NullString, createdAt sql.NullTime,
) AuditEventResponse {
	apiEvent := AuditEventResponse{
		EventID:    eventID,
		EventType:  eventType,
		EntityType: entityType,
		Action:     action,
		ActorName:  actorName,  // Already a string from COALESCE
		TargetName: targetName, // Already a string from COALESCE
	}

	// Handle nullable fields
	if actorUserID.Valid {
		apiEvent.ActorUserID = &actorUserID.Int64
	}
	if actorPhone.Valid {
		apiEvent.ActorPhone = actorPhone.String
	}
	if targetUserID.Valid {
		apiEvent.TargetUserID = &targetUserID.Int64
	}
	if targetPhone.Valid {
		apiEvent.TargetPhone = targetPhone.String
	}
	if entityID.Valid {
		apiEvent.EntityID = &entityID.Int64
	}
	if ipAddress.Valid {
		apiEvent.IPAddress = ipAddress.String
	}
	if userAgent.Valid {
		apiEvent.UserAgent = userAgent.String
	}
	if createdAt.Valid {
		apiEvent.CreatedAt = createdAt.Time.Format(time.RFC3339)
	}

	// Parse JSON details
	if details.Valid {
		var detailsMap map[string]interface{}
		if err := json.Unmarshal([]byte(details.String), &detailsMap); err == nil {
			apiEvent.Details = detailsMap
		}
	}

	return apiEvent
}

// AdminGetAuditStats handles GET /api/admin/audit-events/stats
// @Summary Get audit statistics (Admin)
// @Description Get overall audit trail statistics
// @Tags admin/audit
// @Produce json
// @Success 200 {object} AuditStatsResponse "Audit statistics"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/admin/audit-events/stats [get]
func (h *AdminAuditHandler) AdminGetAuditStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.querier.GetAuditEventStats(r.Context())
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to get audit stats", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch audit statistics", h.logger)
		return
	}

	apiStats := AuditStatsResponse{
		TotalEvents:      stats.TotalEvents,
		UniqueActors:     stats.UniqueActors,
		UniqueEventTypes: stats.UniqueEventTypes,
	}

	// Handle interface{} timestamp fields - they should be time strings from SQLite
	if stats.EarliestEvent != nil {
		if earliestStr, ok := stats.EarliestEvent.(string); ok {
			if parsed, parseErr := time.Parse("2006-01-02 15:04:05", earliestStr); parseErr == nil {
				apiStats.EarliestEvent = parsed.Format(time.RFC3339)
			}
		}
	}
	if stats.LatestEvent != nil {
		if latestStr, ok := stats.LatestEvent.(string); ok {
			if parsed, parseErr := time.Parse("2006-01-02 15:04:05", latestStr); parseErr == nil {
				apiStats.LatestEvent = parsed.Format(time.RFC3339)
			}
		}
	}

	RespondWithJSON(w, http.StatusOK, apiStats, h.logger)
}

// AdminGetAuditEventTypeStats handles GET /api/admin/audit-events/type-stats
// @Summary Get audit statistics by event type (Admin)
// @Description Get audit event counts grouped by event type
// @Tags admin/audit
// @Produce json
// @Success 200 {array} EventTypeStatsResponse "Event type statistics"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/admin/audit-events/type-stats [get]
func (h *AdminAuditHandler) AdminGetAuditEventTypeStats(w http.ResponseWriter, r *http.Request) {
	typeStats, err := h.querier.GetAuditEventsByTypeStats(r.Context())
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to get audit type stats", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch event type statistics", h.logger)
		return
	}

	apiTypeStats := make([]EventTypeStatsResponse, 0, len(typeStats))
	for _, stat := range typeStats {
		apiTypeStat := EventTypeStatsResponse{
			EventType:  stat.EventType,
			EventCount: stat.EventCount,
		}

		// Handle interface{} timestamp field
		if stat.LatestEvent != nil {
			if latestStr, ok := stat.LatestEvent.(string); ok {
				if parsed, parseErr := time.Parse("2006-01-02 15:04:05", latestStr); parseErr == nil {
					apiTypeStat.LatestEvent = parsed.Format(time.RFC3339)
				}
			}
		}

		apiTypeStats = append(apiTypeStats, apiTypeStat)
	}

	RespondWithJSON(w, http.StatusOK, apiTypeStats, h.logger)
}
