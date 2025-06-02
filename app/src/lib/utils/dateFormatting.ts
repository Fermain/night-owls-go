import { formatDistanceToNow } from 'date-fns';
import { formatShiftTime as formatShiftTimeSAST } from './timezone';

/**
 * Format time slot range for times that are already in SAST (no conversion needed)
 */
export function formatTimeSlot(startTimeIso: string, endTimeIso: string): string {
	try {
		const start = new Date(startTimeIso);
		const end = new Date(endTimeIso);

		if (isNaN(start.getTime()) || isNaN(end.getTime())) {
			return 'Invalid Time Range';
		}

		// Use local time formatting since API already returns SAST times
		const startFormatted = start.toLocaleTimeString('en-GB', {
			hour: '2-digit',
			minute: '2-digit',
			hour12: false
		});

		const endFormatted = end.toLocaleTimeString('en-GB', {
			hour: '2-digit',
			minute: '2-digit',
			hour12: false
		});

		return `${startFormatted} - ${endFormatted}`;
	} catch {
		return 'Invalid Time Range';
	}
}

/**
 * Format relative time from ISO string (e.g., "2 hours ago", "in 3 days")
 */
export function formatRelativeTime(timeIso: string): string {
	if (!timeIso) return 'N/A';
	try {
		return formatDistanceToNow(new Date(timeIso), { addSuffix: true });
	} catch {
		return 'Invalid Date';
	}
}

/**
 * Format shift time for reports/displays (SAST)
 */
export function formatShiftTime(dateString: string): string {
	return formatShiftTimeSAST(dateString);
}

/**
 * Create date range for API queries
 */
export function createDateRange(days: number = 30): { from: string; to: string } {
	const now = new Date();
	const futureDate = new Date(now.getTime() + days * 24 * 60 * 60 * 1000);
	return {
		from: now.toISOString(),
		to: futureDate.toISOString()
	};
}

/**
 * Validate and normalize date range inputs
 */
export function normalizeDateRange(
	startDate: string | null,
	endDate: string | null,
	defaultDays: number = 30
): { from: string; to: string } {
	if (startDate && endDate) {
		const fromDate = new Date(startDate + 'T00:00:00Z').toISOString();
		const toDate = new Date(endDate + 'T23:59:59Z').toISOString();

		// Safety check to prevent invalid ranges
		if (new Date(fromDate) <= new Date(toDate)) {
			return { from: fromDate, to: toDate };
		}
	}

	// Default fallback
	return createDateRange(defaultDays);
}
