package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"

	db "night-owls-go/internal/db/sqlc_generated"
)

// AuditService handles logging of audit events for security and compliance
type AuditService struct {
	querier db.Querier
	logger  *slog.Logger
}

// NewAuditService creates a new AuditService
func NewAuditService(querier db.Querier, logger *slog.Logger) *AuditService {
	return &AuditService{
		querier: querier,
		logger:  logger.With("service", "AuditService"),
	}
}

// AuditEvent represents an auditable event in the system
type AuditEvent struct {
	EventType    string                 `json:"event_type"`
	ActorUserID  *int64                 `json:"actor_user_id,omitempty"`
	TargetUserID *int64                 `json:"target_user_id,omitempty"`
	EntityType   string                 `json:"entity_type"`
	EntityID     *int64                 `json:"entity_id,omitempty"`
	Action       string                 `json:"action"`
	Details      map[string]interface{} `json:"details,omitempty"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
}

// LogEvent logs an audit event to the database
func (s *AuditService) LogEvent(ctx context.Context, event AuditEvent) error {
	// Convert details to JSON
	var detailsJSON sql.NullString
	if event.Details != nil && len(event.Details) > 0 {
		detailsBytes, err := json.Marshal(event.Details)
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to marshal audit event details", "error", err, "event_type", event.EventType)
			// Continue without details rather than failing the audit
		} else {
			detailsJSON = sql.NullString{String: string(detailsBytes), Valid: true}
		}
	}

	// Convert actor user ID
	var actorUserID sql.NullInt64
	if event.ActorUserID != nil {
		actorUserID = sql.NullInt64{Int64: *event.ActorUserID, Valid: true}
	}

	// Convert target user ID
	var targetUserID sql.NullInt64
	if event.TargetUserID != nil {
		targetUserID = sql.NullInt64{Int64: *event.TargetUserID, Valid: true}
	}

	// Convert entity ID
	var entityID sql.NullInt64
	if event.EntityID != nil {
		entityID = sql.NullInt64{Int64: *event.EntityID, Valid: true}
	}

	// Convert IP address
	var ipAddress sql.NullString
	if event.IPAddress != "" {
		ipAddress = sql.NullString{String: event.IPAddress, Valid: true}
	}

	// Convert user agent
	var userAgent sql.NullString
	if event.UserAgent != "" {
		userAgent = sql.NullString{String: event.UserAgent, Valid: true}
	}

	// Create audit event in database
	_, err := s.querier.CreateAuditEvent(ctx, db.CreateAuditEventParams{
		EventType:    event.EventType,
		ActorUserID:  actorUserID,
		TargetUserID: targetUserID,
		EntityType:   event.EntityType,
		EntityID:     entityID,
		Action:       event.Action,
		Details:      detailsJSON,
		IpAddress:    ipAddress,
		UserAgent:    userAgent,
	})

	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to create audit event", 
			"error", err,
			"event_type", event.EventType,
			"action", event.Action,
			"entity_type", event.EntityType,
		)
		return err
	}

	s.logger.InfoContext(ctx, "Audit event logged",
		"event_type", event.EventType,
		"action", event.Action,
		"entity_type", event.EntityType,
		"actor_user_id", event.ActorUserID,
		"target_user_id", event.TargetUserID,
	)

	return nil
}

// LogUserLogin logs a user login event
func (s *AuditService) LogUserLogin(ctx context.Context, userID int64, userName, userPhone, ipAddress, userAgent string) error {
	return s.LogEvent(ctx, AuditEvent{
		EventType:   "user.login",
		ActorUserID: &userID,
		EntityType:  "user",
		EntityID:    &userID,
		Action:      "login",
		Details: map[string]interface{}{
			"user_name":  userName,
			"user_phone": userPhone,
		},
		IPAddress: ipAddress,
		UserAgent: userAgent,
	})
}

// LogUserRegistration logs a user registration event
func (s *AuditService) LogUserRegistration(ctx context.Context, userID int64, userName, userPhone, ipAddress, userAgent string) error {
	return s.LogEvent(ctx, AuditEvent{
		EventType:  "user.registered",
		EntityType: "user",
		EntityID:   &userID,
		Action:     "registered",
		Details: map[string]interface{}{
			"user_name":  userName,
			"user_phone": userPhone,
		},
		IPAddress: ipAddress,
		UserAgent: userAgent,
	})
}

// LogUserCreated logs when an admin creates a new user
func (s *AuditService) LogUserCreated(ctx context.Context, actorUserID, targetUserID int64, targetUserName, targetUserPhone string, targetRole string, ipAddress, userAgent string) error {
	return s.LogEvent(ctx, AuditEvent{
		EventType:    "user.created",
		ActorUserID:  &actorUserID,
		TargetUserID: &targetUserID,
		EntityType:   "user",
		EntityID:     &targetUserID,
		Action:       "created",
		Details: map[string]interface{}{
			"target_user_name":  targetUserName,
			"target_user_phone": targetUserPhone,
			"target_role":       targetRole,
		},
		IPAddress: ipAddress,
		UserAgent: userAgent,
	})
}

// LogUserUpdated logs when an admin updates a user
func (s *AuditService) LogUserUpdated(ctx context.Context, actorUserID, targetUserID int64, changes map[string]interface{}, ipAddress, userAgent string) error {
	return s.LogEvent(ctx, AuditEvent{
		EventType:    "user.updated",
		ActorUserID:  &actorUserID,
		TargetUserID: &targetUserID,
		EntityType:   "user",
		EntityID:     &targetUserID,
		Action:       "updated",
		Details:      changes,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	})
}

// LogUserRoleChanged logs when an admin changes a user's role
func (s *AuditService) LogUserRoleChanged(ctx context.Context, actorUserID, targetUserID int64, oldRole, newRole string, ipAddress, userAgent string) error {
	return s.LogEvent(ctx, AuditEvent{
		EventType:    "user.role_changed",
		ActorUserID:  &actorUserID,
		TargetUserID: &targetUserID,
		EntityType:   "user",
		EntityID:     &targetUserID,
		Action:       "role_changed",
		Details: map[string]interface{}{
			"old_role": oldRole,
			"new_role": newRole,
		},
		IPAddress: ipAddress,
		UserAgent: userAgent,
	})
}

// LogUserDeleted logs when an admin deletes a user
func (s *AuditService) LogUserDeleted(ctx context.Context, actorUserID, targetUserID int64, targetUserName, targetUserPhone string, ipAddress, userAgent string) error {
	return s.LogEvent(ctx, AuditEvent{
		EventType:    "user.deleted",
		ActorUserID:  &actorUserID,
		TargetUserID: &targetUserID,
		EntityType:   "user",
		EntityID:     &targetUserID,
		Action:       "deleted",
		Details: map[string]interface{}{
			"target_user_name":  targetUserName,
			"target_user_phone": targetUserPhone,
		},
		IPAddress: ipAddress,
		UserAgent: userAgent,
	})
}

// LogUserBulkDeleted logs when an admin bulk deletes users
func (s *AuditService) LogUserBulkDeleted(ctx context.Context, actorUserID int64, deletedUserIDs []int64, deletedCount int, ipAddress, userAgent string) error {
	return s.LogEvent(ctx, AuditEvent{
		EventType:   "user.bulk_deleted",
		ActorUserID: &actorUserID,
		EntityType:  "user",
		Action:      "bulk_deleted",
		Details: map[string]interface{}{
			"deleted_user_ids": deletedUserIDs,
			"deleted_count":    deletedCount,
		},
		IPAddress: ipAddress,
		UserAgent: userAgent,
	})
} 