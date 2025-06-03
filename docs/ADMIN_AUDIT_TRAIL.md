# Admin Audit Trail System

## Overview

The Night Owls Admin Audit Trail provides comprehensive tracking of all administrative and user actions within the system. This is essential for security operations, compliance, and system monitoring.

## Phase 1: Database Design & Core Infrastructure

### 1.1 Audit Events Table Schema

```sql
CREATE TABLE audit_events (
    event_id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_type TEXT NOT NULL, -- 'user.created', 'booking.assigned', etc.
    actor_user_id INTEGER REFERENCES users(user_id), -- Who performed the action
    target_user_id INTEGER REFERENCES users(user_id), -- Who was affected (optional)
    entity_type TEXT NOT NULL, -- 'user', 'booking', 'schedule', 'report', etc.
    entity_id INTEGER, -- ID of the affected entity
    action TEXT NOT NULL, -- 'created', 'updated', 'deleted', 'assigned', etc.
    details TEXT, -- JSON with before/after values and context
    ip_address TEXT, -- For security tracking
    user_agent TEXT, -- Browser/device info
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_events_type ON audit_events(event_type);
CREATE INDEX idx_audit_events_actor ON audit_events(actor_user_id);
CREATE INDEX idx_audit_events_target ON audit_events(target_user_id);
CREATE INDEX idx_audit_events_created_at ON audit_events(created_at);
```

### 1.2 Event Types to Track

#### User Management (Phase 1 - Authentication Focus)
- `user.registered` - User self-registration
- `user.login` - User login (including admin)
- `user.created` - Admin creates user
- `user.updated` - Admin updates user profile
- `user.role_changed` - Admin changes user role
- `user.deleted` - Admin deletes user
- `user.bulk_deleted` - Admin bulk deletes users

#### Booking/Shift Management (Phase 2)
- `booking.created` - User books shift
- `booking.assigned` - Admin assigns user to shift
- `booking.checked_in` - User checks in to shift
- `booking.cancelled` - User/admin cancels booking

#### Schedule Management (Phase 3)
- `schedule.created` - Admin creates schedule
- `schedule.updated` - Admin updates schedule
- `schedule.deleted` - Admin deletes schedule
- `schedule.bulk_deleted` - Admin bulk deletes schedules

#### Report Management (Phase 4)
- `report.submitted` - User submits shift report
- `report.off_shift_submitted` - User submits off-shift report
- `report.archived` - Admin archives report
- `report.unarchived` - Admin unarchives report

#### Broadcast Management (Phase 5)
- `broadcast.created` - Admin creates broadcast
- `broadcast.sent` - System sends broadcast

#### Emergency Contacts (Phase 6)
- `emergency_contact.created` - Admin creates emergency contact
- `emergency_contact.updated` - Admin updates emergency contact
- `emergency_contact.deleted` - Admin deletes emergency contact

#### System Events (Phase 7)
- `push.subscribed` - User subscribes to notifications
- `push.unsubscribed` - User unsubscribes

## Phase 2: Backend Implementation

### 2.1 Audit Service

```go
// internal/service/audit_service.go
type AuditService struct {
    querier db.Querier
    logger  *slog.Logger
}

type AuditEvent struct {
    EventType    string
    ActorUserID  *int64
    TargetUserID *int64
    EntityType   string
    EntityID     *int64
    Action       string
    Details      map[string]interface{}
    IPAddress    string
    UserAgent    string
}

func (s *AuditService) LogEvent(ctx context.Context, event AuditEvent) error
```

### 2.2 Audit Middleware

```go
// Middleware to capture IP and User-Agent for all audited requests
func AuditContextMiddleware(next http.Handler) http.Handler
```

### 2.3 Integration Points

- Modify all admin handlers to log events
- Add audit logging to service layer methods
- Hook into authentication for login tracking

### 2.4 API Endpoints

```go
// GET /api/admin/audit-events
// GET /api/admin/audit-events?event_type=user.created&user_id=123&limit=50&offset=0
func (h *AdminAuditHandler) ListAuditEvents(w http.ResponseWriter, r *http.Request)

// GET /api/admin/audit-events/stats
func (h *AdminAuditHandler) GetAuditStats(w http.ResponseWriter, r *http.Request)
```

## Phase 3: Frontend Implementation

### 3.1 History Page Structure

```
/admin/history
├── Timeline View (default)
├── Event Type Filter
├── User Filter  
├── Date Range Filter
├── Search
└── Export functionality
```

### 3.2 Components to Create

- `AdminHistoryPage.svelte` - Main history page
- `AuditTimeline.svelte` - Timeline component
- `AuditEventCard.svelte` - Individual event display
- `AuditFilters.svelte` - Filtering controls
- `AuditEventDetails.svelte` - Detailed event modal

### 3.3 Event Display Format

```
[Time] [Actor] [Action] [Target/Entity]
├── "2 hours ago"
├── "John Admin" 
├── "assigned shift to"
└── "Sarah Volunteer for Daily Evening Patrol (Jan 25, 6:00 PM)"

Details: Previous assignment was "Mike Security"
IP: 192.168.1.100 | Browser: Chrome on Windows
```

## Phase 4: Implementation Priority

### High Priority Events (Security Critical)
1. User role changes (admin privileges)
2. User deletion/bulk deletion
3. Admin logins
4. Report archiving/unarchiving
5. Emergency contact changes

### Medium Priority Events
1. User creation/updates
2. Schedule management
3. Shift assignments
4. Broadcast creation

### Low Priority Events
1. Regular user bookings
2. Check-ins
3. Report submissions
4. Push notification subscriptions

## Phase 5: Advanced Features

### 5.1 Security Enhancements
- Audit log integrity (checksums)
- Audit log retention policies
- Suspicious activity detection
- Export to external logging systems

### 5.2 Analytics & Insights
- Activity heatmaps
- User behavior patterns
- System usage statistics
- Performance impact tracking

## Implementation Status

- [x] Phase 1: Database schema and audit service
- [x] Phase 2: User authentication audit events  
- [x] Phase 3: User management audit events
- [ ] Phase 4: Frontend timeline view
- [ ] Phase 5: Filtering and search
- [ ] Phase 6: Additional entity types
- [ ] Phase 7: Advanced features

## Current Implementation Details

### Completed Features

**Database Schema (✅)**
- `audit_events` table created with proper indexing
- Migration 000018_create_audit_events implemented
- SQLC queries generated for audit operations

**Audit Service (✅)**
- `AuditService` with structured event logging
- JSON serialization for event details
- Helper methods for user authentication events:
  - `LogUserLogin` - tracks successful logins
  - `LogUserRegistration` - tracks new registrations
  - `LogUserCreated` - tracks admin user creation
  - `LogUserUpdated` - tracks user updates
  - `LogUserRoleChanged` - tracks role changes
  - `LogUserDeleted` - tracks user deletion
  - `LogUserBulkDeleted` - tracks bulk operations

**Audit Middleware (✅)**
- `AuditContextMiddleware` captures IP address and User-Agent
- Proxy-aware IP detection (X-Forwarded-For, X-Real-IP, etc.)
- Context propagation for audit information

**Authentication Audit Integration (✅)**
- `VerifyHandler` logs successful OTP logins
- `DevLoginHandler` logs development mode logins
- Comprehensive error handling for audit failures

### User Authentication Events Currently Tracked

1. **user.login** - When users successfully authenticate via OTP or dev login
   - Captures: user ID, name, phone, IP address, User-Agent
   - Triggered: After successful OTP verification or dev login

### Next Steps

**Phase 3: User Management Audit Events**
- Integrate audit logging into admin user handlers:
  - User creation (`user.created`)
  - User updates (`user.updated`) 
  - Role changes (`user.role_changed`)
  - User deletion (`user.deleted`)
  - Bulk deletion (`user.bulk_deleted`)

**Phase 4: Frontend Timeline View**
- Create admin history page at `/admin/history`
- Implement audit event API endpoints
- Build timeline UI components

## Security Considerations

1. **Data Sensitivity**: Audit logs contain sensitive information and must be protected
2. **Access Control**: Only admins should access audit logs
3. **Retention**: Implement proper retention policies for compliance
4. **Integrity**: Ensure audit logs cannot be tampered with
5. **Performance**: Audit logging should not significantly impact system performance

## Technical Notes

- Use SQLite triggers for critical system events as backup
- Implement async logging to avoid performance impact
- Consider separate database for audit logs in production
- Implement log rotation and archiving
- Add monitoring for audit system health 