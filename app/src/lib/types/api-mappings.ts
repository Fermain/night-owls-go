/**
 * API mapping utilities for Night Owls application
 * Bridges auto-generated API types with domain types
 */

import type { components } from './api';
import type {
	User,
	Schedule,
	Booking,
	Report,
	EmergencyContact,
	AuditEvent,
	DashboardMetrics,
	QualityMetrics,
	MemberContribution,
	TimeSlotPattern,
	UserRole,
	AuditEventType,
	WeekendStatus,
	ContributionCategory,
	Broadcast,
	CreateBroadcastData,
	BroadcastAudience,
	BroadcastStatus
} from './domain';
import { ReportSeverity } from './domain'; // Import as value, not type

// === TYPE ALIASES FROM GENERATED API ===

type APIUser = components['schemas']['api.UserAPIResponse'];
type APISchedule = components['schemas']['api.ScheduleResponse'];
type APIBooking = components['schemas']['api.BookingResponse'];
type APIBookingWithSchedule = components['schemas']['api.BookingWithScheduleResponse'];
type APIReport = components['schemas']['api.AdminReportResponse'];
type APIEmergencyContact = components['schemas']['api.EmergencyContactResponse'];
type APIDashboardMetrics = components['schemas']['service.DashboardMetrics'];
type APIQualityMetrics = components['schemas']['service.QualityMetrics'];
type APIMemberContribution = components['schemas']['service.MemberContribution'];
type APITimeSlotPattern = components['schemas']['service.TimeSlotPattern'];

// === USER MAPPINGS ===

export function mapAPIUserToDomain(apiUser: APIUser): User {
	return {
		id: apiUser.id ?? 0,
		name: apiUser.name ?? '',
		phone: apiUser.phone ?? '',
		role: (apiUser.role as UserRole) ?? 'guest',
		createdAt: apiUser.created_at ?? new Date().toISOString(),
		isActive: true // API doesn't provide this field yet
	};
}

// === SCHEDULE MAPPINGS ===

export function mapAPIScheduleToDomain(apiSchedule: APISchedule): Schedule {
	return {
		id: apiSchedule.schedule_id ?? 0,
		name: apiSchedule.name ?? '',
		cronExpression: apiSchedule.cron_expr ?? '',
		timezone: apiSchedule.timezone ?? null,
		startDate: apiSchedule.start_date ?? null,
		endDate: apiSchedule.end_date ?? null,
		durationMinutes: apiSchedule.duration_minutes ?? 0,
		isActive: true, // API doesn't provide this field
		createdAt: new Date().toISOString(), // API doesn't provide this field
		updatedAt: new Date().toISOString() // API doesn't provide this field
	};
}

// === BOOKING MAPPINGS ===

export function mapAPIBookingToDomain(apiBooking: APIBooking): Booking {
	return {
		id: apiBooking.booking_id ?? 0,
		userId: apiBooking.user_id ?? 0,
		scheduleId: apiBooking.schedule_id ?? 0,
		shiftStart: apiBooking.shift_start ?? '',
		shiftEnd: apiBooking.shift_end ?? '',
		buddyName: apiBooking.buddy_name ?? null,
		buddyUserId: apiBooking.buddy_user_id ?? null,
		checkedInAt: apiBooking.checked_in_at ?? null,
		createdAt: apiBooking.created_at ?? new Date().toISOString()
	};
}

export function mapAPIBookingWithScheduleToDomain(apiBooking: APIBookingWithSchedule): Booking {
	return {
		id: apiBooking.booking_id ?? 0,
		userId: apiBooking.user_id ?? 0,
		scheduleId: apiBooking.schedule_id ?? 0,
		shiftStart: apiBooking.shift_start ?? '',
		shiftEnd: apiBooking.shift_end ?? '',
		buddyName: apiBooking.buddy_name ?? null,
		buddyUserId: apiBooking.buddy_user_id ?? null,
		checkedInAt: apiBooking.checked_in_at ?? null,
		createdAt: apiBooking.created_at ?? new Date().toISOString(),
		scheduleName: apiBooking.schedule_name ?? undefined
	};
}

// === REPORT MAPPINGS ===

export function mapAPIReportToDomain(apiReport: APIReport): Report {
	return {
		id: apiReport.report_id ?? 0,
		userId: apiReport.user_id ?? 0,
		bookingId: apiReport.booking_id ?? null,
		scheduleId: apiReport.schedule_id ?? null,
		message: apiReport.message ?? '',
		severity: (apiReport.severity as ReportSeverity) ?? ReportSeverity.Info,
		latitude: apiReport.latitude ?? null,
		longitude: apiReport.longitude ?? null,
		gpsAccuracy: apiReport.gps_accuracy ?? null,
		gpsTimestamp: apiReport.gps_timestamp ?? null,
		createdAt: apiReport.created_at ?? new Date().toISOString(),
		archivedAt: apiReport.archived_at ?? null,
		// Admin context fields - convert null to undefined
		userName: apiReport.user_name || undefined,
		userPhone: apiReport.user_phone || undefined,
		scheduleName: apiReport.schedule_name || undefined,
		shiftStart: apiReport.shift_start || undefined,
		shiftEnd: apiReport.shift_end || undefined
	};
}

// === EMERGENCY CONTACT MAPPINGS ===

export function mapAPIEmergencyContactToDomain(apiContact: APIEmergencyContact): EmergencyContact {
	return {
		id: apiContact.id ?? 0,
		name: apiContact.name ?? '',
		number: apiContact.number ?? '',
		description: apiContact.description ?? null,
		isDefault: apiContact.is_default ?? false,
		displayOrder: apiContact.display_order ?? 0
	};
}

// === DASHBOARD MAPPINGS ===

export function mapAPIDashboardMetricsToDomain(apiMetrics: APIDashboardMetrics): DashboardMetrics {
	return {
		totalShifts: apiMetrics.total_shifts ?? 0,
		bookedShifts: apiMetrics.booked_shifts ?? 0,
		unfilledShifts: apiMetrics.unfilled_shifts ?? 0,
		fillRate: apiMetrics.fill_rate ?? 0,
		checkedInShifts: apiMetrics.checked_in_shifts ?? 0,
		checkInRate: apiMetrics.check_in_rate ?? 0,
		completedShifts: apiMetrics.completed_shifts ?? 0,
		completionRate: apiMetrics.completion_rate ?? 0,
		nextWeekUnfilled: apiMetrics.next_week_unfilled ?? 0,
		thisWeekendStatus: (apiMetrics.this_weekend_status as WeekendStatus) ?? 'no_shifts'
	};
}

export function mapAPIQualityMetricsToDomain(apiMetrics: APIQualityMetrics): QualityMetrics {
	return {
		reliabilityScore: apiMetrics.reliability_score ?? 0,
		noShowRate: apiMetrics.no_show_rate ?? 0,
		incompleteRate: apiMetrics.incomplete_rate ?? 0
	};
}

export function mapAPIMemberContributionToDomain(
	apiContribution: APIMemberContribution
): MemberContribution {
	return {
		userId: apiContribution.user_id ?? 0,
		name: apiContribution.name ?? '',
		phone: apiContribution.phone ?? '',
		shiftsBooked: apiContribution.shifts_booked ?? 0,
		shiftsAttended: apiContribution.shifts_attended ?? 0,
		shiftsCompleted: apiContribution.shifts_completed ?? 0,
		attendanceRate: apiContribution.attendance_rate ?? 0,
		completionRate: apiContribution.completion_rate ?? 0,
		contributionCategory:
			(apiContribution.contribution_category as ContributionCategory) ?? 'non_contributor',
		lastShiftDate: apiContribution.last_shift_date ?? null
	};
}

export function mapAPITimeSlotPatternToDomain(apiPattern: APITimeSlotPattern): TimeSlotPattern {
	return {
		dayOfWeek: apiPattern.day_of_week ?? '',
		hourOfDay: apiPattern.hour_of_day ?? '',
		totalBookings: apiPattern.total_bookings ?? 0,
		checkInRate: apiPattern.check_in_rate ?? 0,
		completionRate: apiPattern.completion_rate ?? 0
	};
}

// === AUDIT MAPPINGS ===

// Note: API types for audit events are not yet in the generated schema
// These will need to be added when the audit API is included in the OpenAPI spec

export function mapAPIAuditEventToDomain(apiEvent: Record<string, unknown>): AuditEvent {
	return {
		id: (apiEvent.event_id as number) ?? 0,
		eventType: (apiEvent.event_type as AuditEventType) ?? 'user.login',
		userId: (apiEvent.actor_user_id as number | null) ?? null,
		targetUserId: (apiEvent.target_user_id as number | null) ?? null,
		entityType: (apiEvent.entity_type as string) ?? undefined,
		entityId: (apiEvent.entity_id as number | null) ?? null,
		action: (apiEvent.action as string) ?? undefined,
		ipAddress: (apiEvent.ip_address as string | null) ?? null,
		userAgent: (apiEvent.user_agent as string | null) ?? null,
		details: (apiEvent.details as Record<string, unknown>) ?? {},
		createdAt: (apiEvent.created_at as string) ?? new Date().toISOString(),
		userName: (apiEvent.actor_name as string) ?? 'Unknown',
		userPhone: (apiEvent.actor_phone as string) ?? undefined,
		targetUserName: (apiEvent.target_name as string) ?? undefined,
		targetUserPhone: (apiEvent.target_phone as string) ?? undefined
	};
}

// === BATCH MAPPING UTILITIES ===

export function mapAPIUserArrayToDomain(apiUsers: APIUser[]): User[] {
	return apiUsers.map(mapAPIUserToDomain);
}

export function mapAPIScheduleArrayToDomain(apiSchedules: APISchedule[]): Schedule[] {
	return apiSchedules.map(mapAPIScheduleToDomain);
}

export function mapAPIBookingArrayToDomain(apiBookings: APIBooking[]): Booking[] {
	return apiBookings.map(mapAPIBookingToDomain);
}

export function mapAPIBookingWithScheduleArrayToDomain(
	apiBookings: APIBookingWithSchedule[]
): Booking[] {
	return apiBookings.map(mapAPIBookingWithScheduleToDomain);
}

export function mapAPIReportArrayToDomain(apiReports: APIReport[]): Report[] {
	return apiReports.map(mapAPIReportToDomain);
}

export function mapAPIEmergencyContactArrayToDomain(
	apiContacts: APIEmergencyContact[]
): EmergencyContact[] {
	return apiContacts.map(mapAPIEmergencyContactToDomain);
}

export function mapAPIMemberContributionArrayToDomain(
	apiContributions: APIMemberContribution[]
): MemberContribution[] {
	return apiContributions.map(mapAPIMemberContributionToDomain);
}

export function mapAPITimeSlotPatternArrayToDomain(
	apiPatterns: APITimeSlotPattern[]
): TimeSlotPattern[] {
	return apiPatterns.map(mapAPITimeSlotPatternToDomain);
}

// === REQUEST BODY MAPPING UTILITIES ===

export function mapCreateUserToAPIRequest(userData: {
	name: string;
	phone: string;
	role?: UserRole;
}): components['schemas']['api.createUserRequest'] {
	return {
		name: userData.name,
		phone: userData.phone,
		role: userData.role
	};
}

export function mapUpdateUserToAPIRequest(userData: {
	name: string;
	phone: string;
	role: UserRole;
}): components['schemas']['api.updateUserRequest'] {
	return {
		name: userData.name,
		phone: userData.phone,
		role: userData.role
	};
}

export function mapCreateEmergencyContactToAPIRequest(contactData: {
	name: string;
	number: string;
	description?: string;
	isDefault?: boolean;
	displayOrder?: number;
}): components['schemas']['api.CreateEmergencyContactRequest'] {
	return {
		name: contactData.name,
		number: contactData.number,
		description: contactData.description,
		is_default: contactData.isDefault,
		display_order: contactData.displayOrder
	};
}

export function mapUpdateEmergencyContactToAPIRequest(contactData: {
	name: string;
	number: string;
	description?: string;
	isDefault?: boolean;
	displayOrder?: number;
}): components['schemas']['api.UpdateEmergencyContactRequest'] {
	return {
		name: contactData.name,
		number: contactData.number,
		description: contactData.description,
		is_default: contactData.isDefault,
		display_order: contactData.displayOrder
	};
}

// === BROADCAST MAPPINGS ===

// Note: Using the Zod schema types since broadcasts don't have OpenAPI types yet
export function mapCreateBroadcastToAPIRequest(broadcastData: CreateBroadcastData): {
	title: string;
	message: string;
	audience: BroadcastAudience;
	push_enabled: boolean;
	scheduled_at?: string;
} {
	return {
		title: broadcastData.title,
		message: broadcastData.message,
		audience: broadcastData.audience,
		push_enabled: broadcastData.pushEnabled,
		scheduled_at: broadcastData.scheduledAt
	};
}

// Map from API response to domain type
export function mapAPIBroadcastToDomain(apiBroadcast: {
	broadcast_id: number;
	title: string;
	message: string;
	audience: BroadcastAudience;
	sender_user_id: number;
	sender_name?: string;
	push_enabled: boolean;
	scheduled_at?: string | null;
	sent_at?: string | null;
	status: string;
	recipient_count: number;
	sent_count: number;
	failed_count: number;
	created_at: string;
}): Broadcast {
	return {
		id: apiBroadcast.broadcast_id,
		title: apiBroadcast.title,
		message: apiBroadcast.message,
		audience: apiBroadcast.audience,
		senderUserId: apiBroadcast.sender_user_id,
		senderName: apiBroadcast.sender_name,
		pushEnabled: apiBroadcast.push_enabled,
		scheduledAt: apiBroadcast.scheduled_at,
		sentAt: apiBroadcast.sent_at,
		status: apiBroadcast.status as BroadcastStatus,
		recipientCount: apiBroadcast.recipient_count,
		sentCount: apiBroadcast.sent_count,
		failedCount: apiBroadcast.failed_count,
		createdAt: apiBroadcast.created_at
	};
}
