import {
	differenceInDays,
	isToday,
	isTomorrow,
	format,
	formatDistanceToNow,
	isValid,
	parseISO,
	differenceInHours
} from 'date-fns';
import { SAST_TIMEZONE, SAST_LOCALE } from './timezone';

/**
 * Get relative time description for a shift using date-fns
 */
export function getRelativeTime(dateString: string): string {
	try {
		const date = parseISO(dateString);

		if (!isValid(date)) {
			return 'Invalid Date';
		}

		if (isToday(date)) return 'Today';
		if (isTomorrow(date)) return 'Tomorrow';

		const daysDiff = differenceInDays(date, new Date());

		if (daysDiff <= 7 && daysDiff > 0) {
			return `In ${daysDiff} days`;
		}

		// For dates further out, use a short format
		return format(date, 'MMM d');
	} catch {
		return 'Invalid Date';
	}
}

/**
 * Format shift time range for display in SAST timezone
 */
export function formatShiftTimeRange(startTime: string, endTime: string): string {
	try {
		const start = new Date(startTime);
		const end = new Date(endTime);

		if (isNaN(start.getTime()) || isNaN(end.getTime())) {
			return 'Invalid Time Range';
		}

		const startFormatted = start.toLocaleTimeString(SAST_LOCALE, {
			hour: '2-digit',
			minute: '2-digit',
			hour12: false,
			timeZone: SAST_TIMEZONE
		});

		const endFormatted = end.toLocaleTimeString(SAST_LOCALE, {
			hour: '2-digit',
			minute: '2-digit',
			hour12: false,
			timeZone: SAST_TIMEZONE
		});

		return `${startFormatted} - ${endFormatted}`;
	} catch {
		return 'Invalid Time Range';
	}
}

/**
 * Get a more detailed relative time description (alternative to getRelativeTime)
 */
export function getDetailedRelativeTime(dateString: string): string {
	try {
		const date = parseISO(dateString);

		if (!isValid(date)) {
			return 'Invalid Date';
		}

		return formatDistanceToNow(date, { addSuffix: true });
	} catch {
		return 'Invalid Date';
	}
}

/**
 * Format shift date and time for detailed display in SAST timezone
 */
export function formatShiftDateTime(dateString: string): string {
	try {
		const date = new Date(dateString);

		if (isNaN(date.getTime())) {
			return 'Invalid Date';
		}

		return date.toLocaleString(SAST_LOCALE, {
			weekday: 'short',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
			hour12: false,
			timeZone: SAST_TIMEZONE
		});
	} catch {
		return 'Invalid Date';
	}
}

/**
 * Check if a shift is happening soon (within next 2 hours)
 */
export function isShiftSoon(startTime: string): boolean {
	try {
		const start = parseISO(startTime);
		if (!isValid(start)) return false;

		const now = new Date();
		const hoursUntilShift = differenceInHours(start, now);

		return hoursUntilShift <= 2 && hoursUntilShift >= 0;
	} catch {
		return false;
	}
}

/**
 * Get booking status information for a shift
 */
export function getShiftBookingStatus(shift: {
	is_booked: boolean;
	user_name?: string | null;
	buddy_name?: string | null;
}) {
	if (!shift.is_booked) {
		return {
			status: 'available' as const,
			label: 'Available',
			color: 'text-orange-600'
		};
	}

	if (shift.buddy_name) {
		return {
			status: 'buddy' as const,
			label: 'Team Assignment',
			color: 'text-green-600',
			primary: shift.user_name || 'Primary',
			buddy: shift.buddy_name
		};
	}

	return {
		status: 'single' as const,
		label: shift.user_name || 'Booked',
		color: 'text-green-600'
	};
}
