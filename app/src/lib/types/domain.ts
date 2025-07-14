/**
 * Domain types for Night Owls application
 * These represent core business concepts and should remain stable
 * regardless of API changes.
 */

// === CORE DOMAIN ENTITIES ===

export type UserRole = 'admin' | 'owl' | 'guest';

export interface User {
	id: number;
	name: string;
	phone: string;
	role: UserRole;
	createdAt: string;
	isActive?: boolean;
}

export interface Schedule {
	id: number;
	name: string;
	cronExpression: string;
	timezone: string | null;
	startDate: string | null;
	endDate: string | null;
	durationMinutes: number;
	isActive: boolean;
	createdAt: string;
	updatedAt: string;
	nextRunTime?: string | null;
	slotCount?: number;
}

export interface Shift {
	scheduleId: number;
	scheduleName: string;
	startTime: string;
	endTime: string;
	timezone?: string | null;
	isBooked: boolean;
	bookingId?: number | null;
	userName?: string | null;
	userPhone?: string | null;
	buddyName?: string | null;
}

export interface Booking {
	id: number;
	userId: number;
	scheduleId: number;
	shiftStart: string;
	shiftEnd: string;
	buddyName?: string | null;
	buddyUserId?: number | null;
	checkedInAt?: string | null;
	createdAt: string;
	scheduleName?: string; // For display purposes
}

export interface Report {
	id: number;
	userId: number;
	bookingId?: number | null;
	scheduleId?: number | null;
	message: string;
	severity: ReportSeverity;
	latitude?: number | null;
	longitude?: number | null;
	gpsAccuracy?: number | null;
	gpsTimestamp?: string | null;
	createdAt: string;
	archivedAt?: string | null;
	// Additional context for admin views
	userName?: string;
	userPhone?: string;
	scheduleName?: string;
	shiftStart?: string;
	shiftEnd?: string;
}

export interface EmergencyContact {
	id: number;
	name: string;
	number: string;
	description?: string | null;
	isDefault: boolean;
	displayOrder: number;
}

export interface RecurringAssignment {
	id: number;
	userId: number;
	scheduleId: number;
	dayOfWeek: number; // 0-6, Sunday = 0
	timeSlot: string;
	buddyName?: string | null;
	description?: string | null;
	createdAt: string;
	updatedAt: string;
}

// === ENUMS AND CONSTANTS ===

export enum ReportSeverity {
	Info = 0,
	Warning = 1,
	Critical = 2
}

export const REPORT_SEVERITY_LABELS = {
	[ReportSeverity.Info]: 'Info',
	[ReportSeverity.Warning]: 'Warning',
	[ReportSeverity.Critical]: 'Critical'
} as const;

export const REPORT_SEVERITY_COLORS = {
	[ReportSeverity.Info]: 'blue',
	[ReportSeverity.Warning]: 'yellow',
	[ReportSeverity.Critical]: 'red'
} as const;

export const USER_ROLE_LABELS = {
	admin: 'Administrator',
	owl: 'Night Owl',
	guest: 'Guest'
} as const;

export const DAYS_OF_WEEK = [
	'Sunday',
	'Monday',
	'Tuesday',
	'Wednesday',
	'Thursday',
	'Friday',
	'Saturday'
] as const;

// === VALUE OBJECTS ===

export interface DateRange {
	from: string;
	to: string;
}

export interface TimeSlot {
	start: string;
	end: string;
	timezone?: string;
}

export interface Location {
	latitude: number;
	longitude: number;
	accuracy?: number;
	timestamp?: string;
}

export interface UserSession {
	isAuthenticated: boolean;
	user: User | null;
	token: string | null;
}

// === DASHBOARD AND ANALYTICS ===

export interface DashboardMetrics {
	totalShifts: number;
	bookedShifts: number;
	unfilledShifts: number;
	fillRate: number;
	checkedInShifts: number;
	checkInRate: number;
	completedShifts: number;
	completionRate: number;
	nextWeekUnfilled: number;
	thisWeekendStatus: WeekendStatus;
}

export interface QualityMetrics {
	reliabilityScore: number;
	noShowRate: number;
	incompleteRate: number;
}

export interface MemberContribution {
	userId: number;
	name: string;
	phone: string;
	shiftsBooked: number;
	shiftsAttended: number;
	shiftsCompleted: number;
	attendanceRate: number;
	completionRate: number;
	contributionCategory: ContributionCategory;
	lastShiftDate?: string | null;
}

export interface TimeSlotPattern {
	dayOfWeek: string;
	hourOfDay: string;
	totalBookings: number;
	checkInRate: number;
	completionRate: number;
}

export interface SystemHealth {
	status: 'healthy' | 'warning' | 'critical';
	score: number;
	issues: string[];
	warnings: string[];
}

export type WeekendStatus = 'fully_covered' | 'partial_coverage' | 'critical' | 'no_shifts';

export type ContributionCategory =
	| 'non_contributor'
	| 'minimum_contributor'
	| 'fair_contributor'
	| 'heavy_lifter';

// === AUDIT SYSTEM ===

export interface AuditEvent {
	id: number;
	eventType: AuditEventType;
	userId?: number | null;
	targetUserId?: number | null;
	entityType?: string;
	entityId?: number | null;
	action?: string;
	ipAddress?: string | null;
	userAgent?: string | null;
	details: Record<string, unknown>;
	createdAt: string;
	// Populated for display
	userName?: string;
	userPhone?: string;
	targetUserName?: string;
	targetUserPhone?: string;
}

export type AuditEventType =
	| 'user.login'
	| 'user.registered'
	| 'user.created'
	| 'user.updated'
	| 'user.deleted'
	| 'user.bulk_deleted'
	| 'user.role_changed'
	| 'schedule.created'
	| 'schedule.updated'
	| 'schedule.deleted'
	| 'schedule.bulk_deleted'
	| 'booking.created'
	| 'booking.cancelled'
	| 'booking.checked_in'
	| 'booking.admin_assigned'
	| 'report.created'
	| 'report.archived'
	| 'report.unarchived'
	| 'report.viewed'
	| 'report.deleted'
	| 'auth.logout'
	| 'auth.failed_login'
	| 'auth.session_expired';

export const AUDIT_EVENT_LABELS = {
	'user.login': 'User Login',
	'user.registered': 'User Registered',
	'user.created': 'User Created',
	'user.updated': 'User Updated',
	'user.deleted': 'User Deleted',
	'user.bulk_deleted': 'Bulk User Deletion',
	'user.role_changed': 'Role Changed',
	'schedule.created': 'Schedule Created',
	'schedule.updated': 'Schedule Updated',
	'schedule.deleted': 'Schedule Deleted',
	'schedule.bulk_deleted': 'Bulk Schedule Deletion',
	'booking.created': 'Booking Created',
	'booking.cancelled': 'Booking Cancelled',
	'booking.checked_in': 'Booking Checked In',
	'booking.admin_assigned': 'Admin Assigned Booking',
	'report.created': 'Report Submitted',
	'report.archived': 'Report Archived',
	'report.unarchived': 'Report Unarchived',
	'report.viewed': 'Report Viewed',
	'report.deleted': 'Report Deleted',
	'auth.logout': 'User Logout',
	'auth.failed_login': 'Failed Login',
	'auth.session_expired': 'Session Expired'
} as const;

export const AUDIT_EVENT_COLORS = {
	'user.login': 'green',
	'user.registered': 'green',
	'user.created': 'blue',
	'user.updated': 'yellow',
	'user.deleted': 'red',
	'user.bulk_deleted': 'red',
	'user.role_changed': 'purple',
	'schedule.created': 'blue',
	'schedule.updated': 'yellow',
	'schedule.deleted': 'red',
	'schedule.bulk_deleted': 'red',
	'booking.created': 'green',
	'booking.cancelled': 'orange',
	'booking.checked_in': 'green',
	'booking.admin_assigned': 'blue',
	'report.created': 'blue',
	'report.archived': 'gray',
	'report.unarchived': 'yellow',
	'report.viewed': 'gray',
	'report.deleted': 'red',
	'auth.logout': 'gray',
	'auth.failed_login': 'red',
	'auth.session_expired': 'orange'
} as const;

// === FORM DATA TYPES ===

export interface CreateUserData {
	name: string;
	phone: string;
	role?: UserRole;
}

export interface UpdateUserData extends CreateUserData {
	id: number;
}

export interface CreateBookingData {
	scheduleId: number;
	startTime: string;
	buddyName?: string;
	buddyPhone?: string;
}

export interface CreateReportData {
	message: string;
	severity: ReportSeverity;
	location?: Location;
}

export type CreateOffShiftReportData = CreateReportData;

export interface CreateEmergencyContactData {
	name: string;
	number: string;
	description?: string;
	isDefault?: boolean;
	displayOrder?: number;
}

export interface UpdateEmergencyContactData extends CreateEmergencyContactData {
	id: number;
}

export interface CreateBroadcastData {
	title: string;
	message: string;
	audience: BroadcastAudience;
	pushEnabled: boolean;
	scheduledAt?: string;
}

export type UpdateBroadcastData = CreateBroadcastData;

// === FILTER AND SEARCH TYPES ===

export interface ShiftFilters {
	dateRange?: DateRange;
	scheduleIds?: number[];
	showBooked?: boolean;
	showAvailable?: boolean;
}

export interface UserFilters {
	search?: string;
	role?: UserRole;
	isActive?: boolean;
}

export interface ReportFilters {
	dateRange?: DateRange;
	severity?: ReportSeverity;
	scheduleId?: number;
	userId?: number;
	includeArchived?: boolean;
}

export interface AuditFilters {
	dateRange?: DateRange;
	eventTypes?: AuditEventType[];
	userIds?: number[];
	limit?: number;
	offset?: number;
}

// === PAGINATION ===

export interface PaginationParams {
	limit?: number;
	offset?: number;
}

export interface PaginatedResponse<T> {
	data: T[];
	total: number;
	limit: number;
	offset: number;
	hasMore: boolean;
}

// === BROADCAST SYSTEM ===

export interface Broadcast {
	id: number;
	title: string;
	message: string;
	audience: BroadcastAudience;
	senderUserId: number;
	senderName?: string;
	pushEnabled: boolean;
	scheduledAt?: string | null;
	sentAt?: string | null;
	status: BroadcastStatus;
	recipientCount: number;
	sentCount: number;
	failedCount: number;
	createdAt: string;
}

export type BroadcastAudience = 'all' | 'admins' | 'owls' | 'active';

export type BroadcastStatus = 'draft' | 'scheduled' | 'sending' | 'sent' | 'failed';

export const BROADCAST_AUDIENCE_LABELS = {
	all: 'All Users',
	admins: 'Administrators Only',
	owls: 'Night Owls Only',
	active: 'Active Users Only'
} as const;

export const BROADCAST_STATUS_LABELS = {
	draft: 'Draft',
	scheduled: 'Scheduled',
	sending: 'Sending',
	sent: 'Sent',
	failed: 'Failed'
} as const;

export const BROADCAST_STATUS_COLORS = {
	draft: 'gray',
	scheduled: 'blue',
	sending: 'yellow',
	sent: 'green',
	failed: 'red'
} as const;

// =============================================================================
// DATABASE & MIGRATION TYPES
// =============================================================================

export interface DatabaseHealth {
	status: 'healthy' | 'unhealthy' | 'degraded';
	connectionStatus: 'connected' | 'disconnected' | 'error';
	migrationStatus: 'up-to-date' | 'pending' | 'dirty' | 'failed';
	lastChecked: string;
	version?: number;
	error?: string;
}

export interface MigrationState {
	version: number;
	dirty: boolean;
	description?: string;
	appliedAt?: string;
	error?: string;
}

export interface MigrationResult {
	success: boolean;
	previousVersion?: number;
	currentVersion?: number;
	migrationsApplied: number;
	error?: string;
	requiresManualIntervention?: boolean;
	recoveryInstructions?: string[];
}

export const DATABASE_STATUSES = ['healthy', 'unhealthy', 'degraded'] as const;
export const CONNECTION_STATUSES = ['connected', 'disconnected', 'error'] as const;
export const MIGRATION_STATUSES = ['up-to-date', 'pending', 'dirty', 'failed'] as const;
