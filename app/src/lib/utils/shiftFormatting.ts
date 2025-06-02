export interface ShiftTimeRange {
	startTime: string;
	endTime: string;
}

/**
 * Formats a shift title using community watch tradition.
 * Early morning shifts (midnight to 6 AM) are shown as "previous night".
 * Example: Saturday 2:00 AM becomes "Fri night 02:00-04:00"
 */
export function formatShiftTitle(startTime: string, endTime: string): string {
	try {
		const start = new Date(startTime);
		const end = new Date(endTime);
		const hours = start.getHours();

		// If it's early morning (midnight to 6 AM), show as "previous night"
		if (hours >= 0 && hours < 6) {
			const previousDay = new Date(start);
			previousDay.setDate(start.getDate() - 1);
			const dayName = previousDay.toLocaleDateString('en-GB', { weekday: 'short' });

			// Format time range
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

			return `${dayName} night ${startFormatted}-${endFormatted}`;
		} else {
			// Otherwise show the actual day
			const dayName = start.toLocaleDateString('en-GB', { weekday: 'short' });

			// Format time range
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

			return `${dayName} ${startFormatted}-${endFormatted}`;
		}
	} catch (error) {
		console.error('Error formatting shift title:', error);
		return 'Invalid time';
	}
}

/**
 * Formats shift title in uppercase for condensed display.
 * Used in admin thumbnails and compact views.
 */
export function formatShiftTitleCondensed(startTime: string, endTime: string): string {
	if (!startTime || !endTime) return 'N/A';
	try {
		const start = new Date(startTime);
		const end = new Date(endTime);
		const hours = start.getHours();

		// If it's early morning (midnight to 6 AM), show as "previous night"
		if (hours >= 0 && hours < 6) {
			const previousDay = new Date(start);
			previousDay.setDate(start.getDate() - 1);
			const dayName = previousDay.toLocaleDateString('en-GB', { weekday: 'short' }).toUpperCase();

			// Format time range
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

			return `${dayName} NIGHT ${startFormatted}-${endFormatted}`;
		} else {
			// Otherwise show the actual day
			const dayName = start.toLocaleDateString('en-GB', { weekday: 'short' }).toUpperCase();

			// Format time range
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

			return `${dayName} ${startFormatted}-${endFormatted}`;
		}
	} catch (error) {
		console.error('Error formatting condensed shift title:', error);
		return 'Invalid Time';
	}
}

/**
 * Formats just the day/night portion without time.
 * Used in compact shift cards where time is displayed separately.
 */
export function formatDayNight(timeString: string | undefined): string {
	if (!timeString) return 'Unknown day';
	try {
		const date = new Date(timeString);
		const hours = date.getHours();

		// If it's early morning (midnight to 6 AM), show as "previous night"
		if (hours >= 0 && hours < 6) {
			const previousDay = new Date(date);
			previousDay.setDate(date.getDate() - 1);
			const dayName = previousDay.toLocaleDateString('en-GB', { weekday: 'short' });
			return `${dayName} night`;
		} else {
			// Otherwise show the actual day
			const dayName = date.toLocaleDateString('en-GB', { weekday: 'short' });
			return dayName;
		}
	} catch (error) {
		console.error('Error formatting day/night:', error);
		return 'Unknown day';
	}
}

/**
 * Formats time only (HH:mm format).
 * Helper function for displaying time separately from day/night.
 */
export function formatTime(timeString: string | undefined): string {
	if (!timeString) return '--:--';
	try {
		return new Date(timeString).toLocaleTimeString('en-GB', {
			hour: '2-digit',
			minute: '2-digit',
			hour12: false
		});
	} catch (error) {
		console.error('Error formatting time:', error);
		return '--:--';
	}
}

/**
 * Formats a time range (start-end).
 * Helper function for displaying time ranges.
 */
export function formatTimeRange(startTime: string, endTime: string): string {
	try {
		const startFormatted = formatTime(startTime);
		const endFormatted = formatTime(endTime);
		return `${startFormatted}-${endFormatted}`;
	} catch (error) {
		console.error('Error formatting time range:', error);
		return '--:-- - --:--';
	}
}

/**
 * Formats a time slot in HH:mm - HH:mm format (with spaces).
 * Used for pattern matching and simple time display.
 */
export function formatTimeSlot(startTime: string, endTime: string): string {
	try {
		const start = new Date(startTime);
		const end = new Date(endTime);
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
	} catch (error) {
		console.error('Error formatting time slot:', error);
		return '--:-- - --:--';
	}
}

/**
 * Checks if a shift time is in the early morning (midnight to 6 AM).
 * Used for determining if "night of" logic should apply.
 */
export function isEarlyMorningShift(timeString: string): boolean {
	try {
		const date = new Date(timeString);
		const hours = date.getHours();
		return hours >= 0 && hours < 6;
	} catch (error) {
		console.error('Error checking early morning shift:', error);
		return false;
	}
}
