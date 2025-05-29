import { parseISO, isValid } from 'date-fns';
import type { components } from '$lib/types/api';

/**
 * Booking types from OpenAPI spec
 */
export type BookingResponse = components['schemas']['api.BookingResponse'];
export type BookingWithScheduleResponse = components['schemas']['api.BookingWithScheduleResponse'];
export type AvailableShiftSlot = components['schemas']['service.AvailableShiftSlot'];
export type CreateBookingRequest = components['schemas']['api.CreateBookingRequest'];

/**
 * Booking status types
 */
export const BOOKING_STATUS = {
	UPCOMING: 'upcoming',
	ACTIVE: 'active',
	COMPLETED: 'completed',
	MISSED: 'missed',
	PENDING: 'pending'
} as const;

export type BookingStatus = (typeof BOOKING_STATUS)[keyof typeof BOOKING_STATUS];

/**
 * Format date and time for display
 */
export function formatDateTime(dateString: string): string {
	try {
		const date = parseISO(dateString);
		if (!isValid(date)) return 'Invalid Date';

		return date.toLocaleString('en-US', {
			weekday: 'short',
			month: 'short',
			day: 'numeric',
			hour: 'numeric',
			minute: '2-digit',
			hour12: true
		});
	} catch {
		return 'Invalid Date';
	}
}

/**
 * Format time only for display
 */
export function formatTime(timeString: string): string {
	try {
		const date = parseISO(timeString);
		if (!isValid(date)) return 'Invalid Time';

		return date.toLocaleTimeString('en-GB', {
			hour: '2-digit',
			minute: '2-digit'
		});
	} catch {
		return 'Invalid Time';
	}
}

/**
 * Get time until a specific date/time
 */
export function getTimeUntil(timeString: string): string {
	try {
		const date = parseISO(timeString);
		if (!isValid(date)) return 'Invalid Date';

		const now = new Date();
		const diffMs = date.getTime() - now.getTime();

		if (diffMs < 0) return 'Started';

		const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
		const diffMins = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60));

		if (diffHours > 0) return `${diffHours}h ${diffMins}m`;
		return `${diffMins}m`;
	} catch {
		return 'Invalid Date';
	}
}

/**
 * Get booking status based on times and check-in status
 */
export function getShiftStatus(
	startTime: string,
	endTime: string,
	checkedInAt?: string
): BookingStatus {
	try {
		const now = new Date();
		const start = parseISO(startTime);
		const end = parseISO(endTime);

		if (!isValid(start) || !isValid(end)) return BOOKING_STATUS.PENDING;

		if (now < start) return BOOKING_STATUS.UPCOMING;
		if (now >= start && now <= end) return BOOKING_STATUS.ACTIVE;
		if (checkedInAt) return BOOKING_STATUS.COMPLETED;
		if (now > end) return BOOKING_STATUS.MISSED;

		return BOOKING_STATUS.PENDING;
	} catch {
		return BOOKING_STATUS.PENDING;
	}
}

/**
 * Check if a booking can be cancelled (2 hours before start time)
 */
export function canCancelBooking(startTime: string): boolean {
	try {
		const now = new Date();
		const start = parseISO(startTime);

		if (!isValid(start)) return false;

		const cancellationDeadline = new Date(start.getTime() - 2 * 60 * 60 * 1000); // 2 hours before
		return now < cancellationDeadline;
	} catch {
		return false;
	}
}

/**
 * Check if user can check in (30 minutes before start time)
 */
export function canCheckIn(startTime: string, endTime: string): boolean {
	try {
		const now = new Date();
		const start = parseISO(startTime);
		const end = parseISO(endTime);

		if (!isValid(start) || !isValid(end)) return false;

		const checkInWindow = new Date(start.getTime() - 30 * 60 * 1000); // 30 min before
		return now >= checkInWindow && now <= end;
	} catch {
		return false;
	}
}

/**
 * Check if shift is currently active
 */
export function isShiftActive(startTime: string, endTime: string): boolean {
	try {
		const now = new Date();
		const start = parseISO(startTime);
		const end = parseISO(endTime);

		if (!isValid(start) || !isValid(end)) return false;

		return now >= start && now <= end;
	} catch {
		return false;
	}
}
