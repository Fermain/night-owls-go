/**
 * Timezone utilities for South Africa Standard Time (SAST = UTC+2)
 * This ensures all times are displayed correctly for South African users
 */

// South Africa Standard Time is UTC+2 (no daylight saving)
export const SAST_TIMEZONE = 'Africa/Johannesburg';
export const SAST_LOCALE = 'en-ZA';

/**
 * Format time only for display in SAST
 */
export function formatTime(timeString: string): string {
	try {
		const date = new Date(timeString);
		if (isNaN(date.getTime())) return 'Invalid Time';

		return date.toLocaleTimeString(SAST_LOCALE, {
			hour: '2-digit',
			minute: '2-digit',
			hour12: false,
			timeZone: SAST_TIMEZONE
		});
	} catch {
		return 'Invalid Time';
	}
}

/**
 * Format date and time for display in SAST
 */
export function formatDateTime(dateString: string): string {
	try {
		const date = new Date(dateString);
		if (isNaN(date.getTime())) return 'Invalid Date';

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
 * Format date only for display in SAST
 */
export function formatDate(dateString: string): string {
	try {
		const date = new Date(dateString);
		if (isNaN(date.getTime())) return 'Invalid Date';

		return date.toLocaleDateString(SAST_LOCALE, {
			weekday: 'short',
			month: 'short',
			day: 'numeric',
			timeZone: SAST_TIMEZONE
		});
	} catch {
		return 'Invalid Date';
	}
}

/**
 * Format a time slot from start and end ISO strings in SAST
 */
export function formatTimeSlot(startTimeIso: string, endTimeIso: string): string {
	if (!startTimeIso || !endTimeIso) return 'N/A';
	try {
		const startDate = new Date(startTimeIso);
		const endDate = new Date(endTimeIso);

		if (isNaN(startDate.getTime()) || isNaN(endDate.getTime())) {
			return 'Invalid Date Range';
		}

		const startFormatted = startDate.toLocaleString(SAST_LOCALE, {
			weekday: 'short',
			day: 'numeric',
			month: 'short',
			hour: '2-digit',
			minute: '2-digit',
			hour12: false,
			timeZone: SAST_TIMEZONE
		});

		const endFormatted = endDate.toLocaleTimeString(SAST_LOCALE, {
			hour: '2-digit',
			minute: '2-digit',
			hour12: false,
			timeZone: SAST_TIMEZONE
		});

		return `${startFormatted} - ${endFormatted}`;
	} catch {
		return 'Invalid Date Range';
	}
}

/**
 * Format shift time for reports/displays in SAST
 */
export function formatShiftTime(dateString: string): string {
	try {
		const date = new Date(dateString);
		if (isNaN(date.getTime())) return 'Invalid Date';

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
 * Get current SAST time
 */
export function getCurrentSASTTime(): Date {
	return new Date();
}

/**
 * Check if time operations should use SAST (for future server-side timezone logic)
 */
export function shouldUseSAST(): boolean {
	// For now, always use SAST since this app is only used in South Africa
	return true;
}

/**
 * Convert UTC date to SAST for display (useful for debugging)
 */
export function convertUTCToSAST(utcDateString: string): string {
	try {
		const utcDate = new Date(utcDateString);
		if (isNaN(utcDate.getTime())) return 'Invalid Date';

		// Add 2 hours for SAST offset
		const sastDate = new Date(utcDate.getTime() + 2 * 60 * 60 * 1000);
		return sastDate.toISOString();
	} catch {
		return 'Invalid Date';
	}
}
