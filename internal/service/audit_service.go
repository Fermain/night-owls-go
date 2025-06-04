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
	if len(event.Details) > 0 {
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

// ===== BOOKING LIFECYCLE EVENTS =====

// LogBookingCreated logs when a user creates a booking
func (s *AuditService) LogBookingCreated(ctx context.Context, userID, bookingID, scheduleID int64, scheduleName string, shiftStart, shiftEnd string, buddyName *string, ipAddress, userAgent string) error {
	details := map[string]interface{}{
		"schedule_id":   scheduleID,
		"schedule_name": scheduleName,
		"shift_start":   shiftStart,
		"shift_end":     shiftEnd,
	}
	if buddyName != nil && *buddyName != "" {
		details["buddy_name"] = *buddyName
	}

	return s.LogEvent(ctx, AuditEvent{
		EventType:   "booking.created",
		ActorUserID: &userID,
		EntityType:  "booking",
		EntityID:    &bookingID,
		Action:      "created",
		Details:     details,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	})
}

// LogBookingCancelled logs when a user cancels their booking
func (s *AuditService) LogBookingCancelled(ctx context.Context, userID, bookingID, scheduleID int64, scheduleName string, shiftStart, shiftEnd string, ipAddress, userAgent string) error {
	return s.LogEvent(ctx, AuditEvent{
		EventType:   "booking.cancelled",
		ActorUserID: &userID,
		EntityType:  "booking",
		EntityID:    &bookingID,
		Action:      "cancelled",
		Details: map[string]interface{}{
			"schedule_id":   scheduleID,
			"schedule_name": scheduleName,
			"shift_start":   shiftStart,
			"shift_end":     shiftEnd,
		},
		IPAddress: ipAddress,
		UserAgent: userAgent,
	})
}

// LogBookingCheckedIn logs when a user checks in for their booking
func (s *AuditService) LogBookingCheckedIn(ctx context.Context, userID, bookingID, scheduleID int64, scheduleName string, shiftStart string, ipAddress, userAgent string) error {
	return s.LogEvent(ctx, AuditEvent{
		EventType:   "booking.checked_in",
		ActorUserID: &userID,
		EntityType:  "booking",
		EntityID:    &bookingID,
		Action:      "checked_in",
		Details: map[string]interface{}{
			"schedule_id":   scheduleID,
			"schedule_name": scheduleName,
			"shift_start":   shiftStart,
		},
		IPAddress: ipAddress,
		UserAgent: userAgent,
	})
}

// LogBookingAdminAssigned logs when an admin assigns a user to a booking
func (s *AuditService) LogBookingAdminAssigned(ctx context.Context, actorUserID, targetUserID, bookingID, scheduleID int64, scheduleName string, shiftStart, shiftEnd string, ipAddress, userAgent string) error {
	return s.LogEvent(ctx, AuditEvent{
		EventType:    "booking.admin_assigned",
		ActorUserID:  &actorUserID,
		TargetUserID: &targetUserID,
		EntityType:   "booking",
		EntityID:     &bookingID,
		Action:       "admin_assigned",
		Details: map[string]interface{}{
			"schedule_id":   scheduleID,
			"schedule_name": scheduleName,
			"shift_start":   shiftStart,
			"shift_end":     shiftEnd,
		},
		IPAddress: ipAddress,
		UserAgent: userAgent,
	})
}

// ===== REPORT MANAGEMENT EVENTS =====

// LogReportCreated logs when a user creates an incident report
func (s *AuditService) LogReportCreated(ctx context.Context, userID, reportID int64, bookingID *int64, severity int64, hasLocation bool, ipAddress, userAgent string) error {
	details := map[string]interface{}{
		"severity":     severity,
		"has_location": hasLocation,
	}
	if bookingID != nil {
		details["booking_id"] = *bookingID
		details["report_type"] = "on_shift"
	} else {
		details["report_type"] = "off_shift"
	}

	return s.LogEvent(ctx, AuditEvent{
		EventType:   "report.created",
		ActorUserID: &userID,
		EntityType:  "report",
		EntityID:    &reportID,
		Action:      "created",
		Details:     details,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	})
}

// LogReportArchived logs when an admin archives a report
func (s *AuditService) LogReportArchived(ctx context.Context, actorUserID, reportID int64, reporterUserID *int64, severity int64, ipAddress, userAgent string) error {
	details := map[string]interface{}{
		"severity": severity,
	}
	if reporterUserID != nil {
		details["reporter_user_id"] = *reporterUserID
	}

	return s.LogEvent(ctx, AuditEvent{
		EventType:   "report.archived",
		ActorUserID: &actorUserID,
		EntityType:  "report",
		EntityID:    &reportID,
		Action:      "archived",
		Details:     details,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	})
}

// LogReportUnarchived logs when an admin unarchives a report
func (s *AuditService) LogReportUnarchived(ctx context.Context, actorUserID, reportID int64, reporterUserID *int64, severity int64, ipAddress, userAgent string) error {
	details := map[string]interface{}{
		"severity": severity,
	}
	if reporterUserID != nil {
		details["reporter_user_id"] = *reporterUserID
	}

	return s.LogEvent(ctx, AuditEvent{
		EventType:   "report.unarchived",
		ActorUserID: &actorUserID,
		EntityType:  "report",
		EntityID:    &reportID,
		Action:      "unarchived",
		Details:     details,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	})
}

// LogReportViewed logs when an admin views a specific report
func (s *AuditService) LogReportViewed(ctx context.Context, actorUserID, reportID int64, reporterUserID *int64, severity int64, ipAddress, userAgent string) error {
	details := map[string]interface{}{
		"severity": severity,
	}
	if reporterUserID != nil {
		details["reporter_user_id"] = *reporterUserID
	}

	return s.LogEvent(ctx, AuditEvent{
		EventType:   "report.viewed",
		ActorUserID: &actorUserID,
		EntityType:  "report",
		EntityID:    &reportID,
		Action:      "viewed",
		Details:     details,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	})
}

// LogReportDeleted logs when an admin permanently deletes a report
func (s *AuditService) LogReportDeleted(ctx context.Context, actorUserID, reportID int64, reporterUserID *int64, severity int64, ipAddress, userAgent string) error {
	details := map[string]interface{}{
		"severity": severity,
	}
	if reporterUserID != nil {
		details["reporter_user_id"] = *reporterUserID
	}

	return s.LogEvent(ctx, AuditEvent{
		EventType:   "report.deleted",
		ActorUserID: &actorUserID,
		EntityType:  "report",
		EntityID:    &reportID,
		Action:      "deleted",
		Details:     details,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	})
}

// ===== SCHEDULE MANAGEMENT EVENTS =====

// LogScheduleCreated logs when an admin creates a schedule
func (s *AuditService) LogScheduleCreated(ctx context.Context, actorUserID, scheduleID int64, scheduleName, cronExpr string, timezone *string, durationMinutes int64, ipAddress, userAgent string) error {
	details := map[string]interface{}{
		"schedule_name":    scheduleName,
		"cron_expr":        cronExpr,
		"duration_minutes": durationMinutes,
	}
	if timezone != nil {
		details["timezone"] = *timezone
	}

	return s.LogEvent(ctx, AuditEvent{
		EventType:   "schedule.created",
		ActorUserID: &actorUserID,
		EntityType:  "schedule",
		EntityID:    &scheduleID,
		Action:      "created",
		Details:     details,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	})
}

// LogScheduleUpdated logs when an admin updates a schedule
func (s *AuditService) LogScheduleUpdated(ctx context.Context, actorUserID, scheduleID int64, changes map[string]interface{}, ipAddress, userAgent string) error {
	return s.LogEvent(ctx, AuditEvent{
		EventType:   "schedule.updated",
		ActorUserID: &actorUserID,
		EntityType:  "schedule",
		EntityID:    &scheduleID,
		Action:      "updated",
		Details:     changes,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	})
}

// LogScheduleDeleted logs when an admin deletes a schedule
func (s *AuditService) LogScheduleDeleted(ctx context.Context, actorUserID, scheduleID int64, scheduleName string, ipAddress, userAgent string) error {
	return s.LogEvent(ctx, AuditEvent{
		EventType:   "schedule.deleted",
		ActorUserID: &actorUserID,
		EntityType:  "schedule",
		EntityID:    &scheduleID,
		Action:      "deleted",
		Details: map[string]interface{}{
			"schedule_name": scheduleName,
		},
		IPAddress: ipAddress,
		UserAgent: userAgent,
	})
}

// LogScheduleBulkDeleted logs when an admin bulk deletes schedules
func (s *AuditService) LogScheduleBulkDeleted(ctx context.Context, actorUserID int64, deletedScheduleIDs []int64, deletedCount int, ipAddress, userAgent string) error {
	return s.LogEvent(ctx, AuditEvent{
		EventType:   "schedule.bulk_deleted",
		ActorUserID: &actorUserID,
		EntityType:  "schedule",
		Action:      "bulk_deleted",
		Details: map[string]interface{}{
			"deleted_schedule_ids": deletedScheduleIDs,
			"deleted_count":        deletedCount,
		},
		IPAddress: ipAddress,
		UserAgent: userAgent,
	})
}

// ===== AUTHENTICATION & SECURITY EVENTS =====

// LogUserLogout logs when a user explicitly logs out
func (s *AuditService) LogUserLogout(ctx context.Context, userID int64, ipAddress, userAgent string) error {
	return s.LogEvent(ctx, AuditEvent{
		EventType:   "auth.logout",
		ActorUserID: &userID,
		EntityType:  "user",
		EntityID:    &userID,
		Action:      "logout",
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	})
}

// LogFailedLogin logs failed login attempts
func (s *AuditService) LogFailedLogin(ctx context.Context, phone, reason, ipAddress, userAgent string) error {
	return s.LogEvent(ctx, AuditEvent{
		EventType:  "auth.failed_login",
		EntityType: "user",
		Action:     "failed_login",
		Details: map[string]interface{}{
			"phone":  phone,
			"reason": reason,
		},
		IPAddress: ipAddress,
		UserAgent: userAgent,
	})
}

// LogSessionExpired logs when a user's session expires
func (s *AuditService) LogSessionExpired(ctx context.Context, userID int64, reason string, ipAddress, userAgent string) error {
	return s.LogEvent(ctx, AuditEvent{
		EventType:   "auth.session_expired",
		ActorUserID: &userID,
		EntityType:  "user",
		EntityID:    &userID,
		Action:      "session_expired",
		Details: map[string]interface{}{
			"reason": reason,
		},
		IPAddress: ipAddress,
		UserAgent: userAgent,
	})
} 