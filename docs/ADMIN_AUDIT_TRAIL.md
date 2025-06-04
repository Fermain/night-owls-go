# Admin Audit Trail System

## Overview

The Night Owls Admin Audit Trail provides comprehensive tracking of all administrative and user actions within the system. This is essential for security operations, compliance, and system monitoring.

## Implementation Status

- [x] Phase 1: Database schema and audit service
- [x] Phase 2: User authentication audit events  
- [x] Phase 3: User management audit events
- [x] Phase 4: Frontend timeline view
- [ ] Phase 5: Additional entity types (shifts, reports, schedules)
- [ ] Phase 6: Advanced features (export, retention, analytics)

## Current Implementation

### Database Schema (✅)

**audit_events table**:
- `event_id` - Primary key
- `event_type` - Structured event identifier (e.g., 'user.login', 'user.created')
- `actor_user_id` - Who performed the action (NULL for system events)
- `target_user_id` - Who was affected (optional)
- `entity_type` - Type of entity affected ('user', 'booking', etc.)
- `entity_id` - ID of affected entity (optional)
- `action` - Action performed ('created', 'updated', 'login', etc.)
- `details` - JSON with context data (before/after values, etc.)
- `ip_address` - Source IP address
- `user_agent` - Browser/client information
- `created_at` - Timestamp of the event

**Indexes**: Optimized for queries by event type, actor, target, and timestamp.

### Backend Components (✅)

**AuditService** (`internal/service/audit_service.go`):
- Structured event logging with JSON details
- Helper methods for common event types
- Error handling and logging

**AuditMiddleware** (`internal/api/audit_middleware.go`):
- Captures IP address (proxy-aware)
- Captures User-Agent information
- Context propagation for audit data

**API Endpoints**:
- `GET /api/admin/audit-events` - Paginated events with filtering
- `GET /api/admin/audit-events/stats` - Overall statistics
- `GET /api/admin/audit-events/type-stats` - Event type breakdown

### Frontend Components (✅)

**Admin History Page** (`/admin/history`):
- **AuditTimeline.svelte** - Beautiful chronological timeline
- **AuditFilters.svelte** - Filtering by event type, user, pagination
- **AuditStats.svelte** - Statistical overview with charts
- **Event Details Modal** - Detailed view with JSON data

**Features**:
- Color-coded events: Login (green), Create (blue), Update (yellow), Role Change (purple), Delete (red)
- Relative timestamps with hover for exact time
- IP address and User-Agent display
- Detailed event information in modal dialogs
- Real-time API integration

### Events Currently Tracked

**User Authentication**:
- `user.login` - Successful logins (OTP and dev mode)

**User Management**:
- `user.created` - Admin creates new user
- `user.updated` - Admin updates user profile
- `user.role_changed` - Admin changes user role
- `user.deleted` - Admin deletes user (when not blocked by constraints)
- `user.bulk_deleted` - Admin bulk deletes users

**Data Captured**:
- Actor information (who performed the action)
- Target information (who was affected)
- Before/after values for updates
- IP address and User-Agent
- Structured JSON details
- Precise timestamps

## Usage Examples

### Backend - Adding New Event Types

```go
// In a handler function
func (h *SomeHandler) SomeAction(w http.ResponseWriter, r *http.Request) {
    // ... perform action ...
    
    // Log audit event
    if actorUserID, ok := r.Context().Value(UserIDKey).(int64); ok {
        ipAddress, userAgent := GetAuditInfoFromContext(r.Context())
        
        err := h.auditService.LogSomeEvent(
            r.Context(),
            actorUserID,
            targetID,
            "details",
            ipAddress,
            userAgent,
        )
        if err != nil {
            h.logger.ErrorContext(r.Context(), "Failed to log audit event", "error", err)
        }
    }
}
```

### Frontend - Accessing Audit Data

```typescript
// Fetch audit events
const response = await fetch('/api/admin/audit-events?limit=50&event_type=user.login', {
    headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
    }
});
const events = await response.json();
```

## Next Steps

### Phase 5: Additional Entity Types
- Shifts and schedules management events
- Report creation, archiving, deletion events
- Broadcast management events
- Emergency contact changes

### Phase 6: Advanced Features
- Export functionality (CSV, JSON)
- Advanced filtering (date ranges, complex queries)
- Retention policies and log rotation
- Integration with external logging systems
- Real-time notifications for critical events

## Security Considerations

1. **Access Control**: Only admin users can access audit logs
2. **Data Integrity**: Events are immutable once logged
3. **Performance**: Async logging to avoid blocking operations
4. **Retention**: Consider implementing automatic archiving
5. **Privacy**: Sensitive data is structured but not exposed unnecessarily

## Technical Notes

- Built with Go backend (SQLC + SQLite) and SvelteKit frontend
- Uses structured logging with JSON details for flexibility
- Proxy-aware IP detection for accurate source tracking
- Type-safe with comprehensive error handling
- Optimized database queries with proper indexing