import { formatDistanceToNow } from 'date-fns';

/**
 * Format a time slot from start and end ISO strings
 */
export function formatTimeSlot(startTimeIso: string, endTimeIso: string): string {
	if (!startTimeIso || !endTimeIso) return 'N/A';
	try {
		const startDate = new Date(startTimeIso);
		const endDate = new Date(endTimeIso);

		const startFormatted = startDate.toLocaleString('en-ZA', {
			weekday: 'short',
			day: 'numeric',
			month: 'short',
			hour: '2-digit',
			minute: '2-digit',
			hour12: false,
			timeZone: 'UTC'
		});

		const endFormatted = endDate.toLocaleTimeString('en-ZA', {
			hour: '2-digit',
			minute: '2-digit',
			hour12: false,
			timeZone: 'UTC'
		});

		return `${startFormatted} - ${endFormatted}`;
	} catch {
		return 'Invalid Date Range';
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
 * Format shift time for reports/displays 
 */
export function formatShiftTime(dateString: string): string {
	try {
		return new Date(dateString).toLocaleString('en-ZA', {
			weekday: 'short',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
			timeZone: 'UTC'
		});
	} catch {
		return 'Invalid Date';
	}
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