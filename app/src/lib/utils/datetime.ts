/**
 * Consolidated datetime utilities for the Night Owls application
 * Handles timezone conversion, formatting, parsing, and validation
 *
 * Primary timezone: South Africa Standard Time (SAST = UTC+2)
 */

import { formatDistanceToNow } from 'date-fns';
import {
	CalendarDate,
	type DateValue,
	parseAbsoluteToLocal,
	parseDate as parseDateStringToDateValue
} from '@internationalized/date';

// === CONSTANTS ===
export const SAST_TIMEZONE = 'Africa/Johannesburg';
export const SAST_LOCALE = 'en-ZA';
export const FALLBACK_LOCALE = 'en-GB';

// === TYPE DEFINITIONS ===
export interface DateRange {
	from: string;
	to: string;
}

export interface TimeSlotFormatOptions {
	includeDate?: boolean;
	use12Hour?: boolean;
	showTimezone?: boolean;
}

// === CORE VALIDATION ===

/**
 * Safely parse a date string and return a Date object
 * @param dateString - ISO string, date string, or any valid date input
 * @returns Date object or null if invalid
 */
export function safeParseDate(dateString: string | null | undefined): Date | null {
	if (!dateString) return null;

	try {
		const date = new Date(dateString);
		return isNaN(date.getTime()) ? null : date;
	} catch {
		return null;
	}
}

/**
 * Check if a date string is valid
 */
export function isValidDate(dateString: string | null | undefined): boolean {
	return safeParseDate(dateString) !== null;
}

// === SAST FORMATTING ===

/**
 * Format time only in SAST (HH:mm)
 */
export function formatTime(timeString: string, use12Hour = false): string {
	const date = safeParseDate(timeString);
	if (!date) return 'Invalid Time';

	return date.toLocaleTimeString(SAST_LOCALE, {
		hour: '2-digit',
		minute: '2-digit',
		hour12: use12Hour,
		timeZone: SAST_TIMEZONE
	});
}

/**
 * Format date only in SAST (e.g., "Mon, 15 Jan")
 */
export function formatDate(dateString: string, options: Intl.DateTimeFormatOptions = {}): string {
	const date = safeParseDate(dateString);
	if (!date) return 'Invalid Date';

	const defaultOptions: Intl.DateTimeFormatOptions = {
		weekday: 'short',
		month: 'short',
		day: 'numeric',
		timeZone: SAST_TIMEZONE,
		...options
	};

	return date.toLocaleDateString(SAST_LOCALE, defaultOptions);
}

/**
 * Format date and time in SAST (e.g., "Mon, 15 Jan 14:30")
 */
export function formatDateTime(dateString: string, use12Hour = false): string {
	const date = safeParseDate(dateString);
	if (!date) return 'Invalid Date';

	return date.toLocaleString(SAST_LOCALE, {
		weekday: 'short',
		month: 'short',
		day: 'numeric',
		hour: '2-digit',
		minute: '2-digit',
		hour12: use12Hour,
		timeZone: SAST_TIMEZONE
	});
}

/**
 * Format time slot range (e.g., "Mon, 15 Jan 14:30 - 16:30")
 */
export function formatTimeSlot(
	startTimeIso: string,
	endTimeIso: string,
	options: TimeSlotFormatOptions = {}
): string {
	const startDate = safeParseDate(startTimeIso);
	const endDate = safeParseDate(endTimeIso);

	if (!startDate || !endDate) return 'Invalid Time Range';

	const { includeDate = true, use12Hour = false } = options;

	if (includeDate) {
		const startFormatted = startDate.toLocaleString(SAST_LOCALE, {
			weekday: 'short',
			day: 'numeric',
			month: 'short',
			hour: '2-digit',
			minute: '2-digit',
			hour12: use12Hour,
			timeZone: SAST_TIMEZONE
		});

		const endFormatted = endDate.toLocaleTimeString(SAST_LOCALE, {
			hour: '2-digit',
			minute: '2-digit',
			hour12: use12Hour,
			timeZone: SAST_TIMEZONE
		});

		return `${startFormatted} - ${endFormatted}`;
	} else {
		// Time only format
		const startTime = formatTime(startTimeIso, use12Hour);
		const endTime = formatTime(endTimeIso, use12Hour);
		return `${startTime} - ${endTime}`;
	}
}

/**
 * Format shift time for displays (includes day and time)
 */
export function formatShiftTime(dateString: string, use12Hour = false): string {
	return formatDateTime(dateString, use12Hour);
}

/**
 * Format relative time (e.g., "2 hours ago", "in 3 days")
 */
export function formatRelativeTime(timeIso: string): string {
	const date = safeParseDate(timeIso);
	if (!date) return 'Invalid Date';

	try {
		return formatDistanceToNow(date, { addSuffix: true });
	} catch {
		return 'Invalid Date';
	}
}

// === DATE RANGE OPERATIONS ===

/**
 * Create a date range for API queries
 */
export function createDateRange(days = 30): DateRange {
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
	defaultDays = 30
): DateRange {
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

// === CALENDAR/FORM INTEGRATION ===

/**
 * Parse YYYY-MM-DD string to CalendarDate for form components
 */
export function parseYyyyMmDdToCalendarDate(
	dateString: string | null | undefined
): CalendarDate | undefined {
	if (!dateString) return undefined;

	try {
		const parsed = parseDateStringToDateValue(dateString);
		return new CalendarDate(parsed.year, parsed.month, parsed.day);
	} catch (e) {
		console.error(`Failed to parse date string "${dateString}":`, e);
		return undefined;
	}
}

/**
 * Format CalendarDate to YYYY-MM-DD string
 */
export function formatCalendarDateToYyyyMmDd(
	dateValue: DateValue | null | undefined
): string | null {
	if (!dateValue) return null;
	return dateValue.toString();
}

/**
 * Extract YYYY-MM-DD from ISO string
 */
export function parseIsoToYyyyMmDd(isoString: string | null | undefined): string | null {
	if (!isoString) return null;

	if (typeof isoString === 'string' && isoString.includes('T')) {
		return isoString.split('T')[0];
	}

	return isoString;
}

/**
 * Convert JavaScript Date to YYYY-MM-DD string
 */
export function formatJsDateToYyyyMmDd(date: Date | null | undefined): string | null {
	if (!date) return null;

	const year = date.getFullYear();
	const month = (date.getMonth() + 1).toString().padStart(2, '0');
	const day = date.getDate().toString().padStart(2, '0');

	return `${year}-${month}-${day}`;
}

/**
 * Parse YYYY-MM-DD to JavaScript Date (UTC)
 */
export function parseYyyyMmDdToJsDate(dateStr: string | null | undefined): Date | null {
	if (!dateStr) return null;

	const [year, month, day] = dateStr.split('-').map(Number);
	if (year && month && day) {
		const date = new Date(Date.UTC(year, month - 1, day));

		// Validate the constructed date
		if (
			!isNaN(date.valueOf()) &&
			date.getUTCFullYear() === year &&
			date.getUTCMonth() === month - 1 &&
			date.getUTCDate() === day
		) {
			return date;
		}
	}

	console.warn(`Failed to parse date string "${dateStr}" into a valid JS Date.`);
	return null;
}

/**
 * Convert ISO string to local CalendarDate
 */
export function parseAbsoluteIsoToLocalDate(
	isoString: string | null | undefined
): CalendarDate | undefined {
	if (!isoString) return undefined;

	try {
		const localDateTime = parseAbsoluteToLocal(isoString);
		return new CalendarDate(localDateTime.year, localDateTime.month, localDateTime.day);
	} catch (e) {
		console.error(`Failed to parse ISO string "${isoString}":`, e);

		// Fallback: extract YYYY-MM-DD part
		const yyyyMmDdPart = parseIsoToYyyyMmDd(isoString);
		if (yyyyMmDdPart) {
			return parseYyyyMmDdToCalendarDate(yyyyMmDdPart);
		}

		return undefined;
	}
}

// === TIME CALCULATIONS ===

/**
 * Get current SAST time
 */
export function getCurrentSASTTime(): Date {
	return new Date();
}

/**
 * Calculate time until a specific date/time
 */
export function getTimeUntil(timeString: string): string {
	const date = safeParseDate(timeString);
	if (!date) return 'Invalid Date';

	const now = getCurrentSASTTime();
	const diffMs = date.getTime() - now.getTime();

	if (diffMs < 0) return 'Started';

	const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
	const diffMins = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60));

	if (diffHours > 0) return `in ${diffHours}h ${diffMins}m`;
	return `in ${diffMins}m`;
}

/**
 * Check if a date is today (in SAST)
 */
export function isToday(dateString: string): boolean {
	const date = safeParseDate(dateString);
	if (!date) return false;

	const today = new Date();
	return formatJsDateToYyyyMmDd(date) === formatJsDateToYyyyMmDd(today);
}

/**
 * Check if a date is tomorrow (in SAST)
 */
export function isTomorrow(dateString: string): boolean {
	const date = safeParseDate(dateString);
	if (!date) return false;

	const tomorrow = new Date();
	tomorrow.setDate(tomorrow.getDate() + 1);
	return formatJsDateToYyyyMmDd(date) === formatJsDateToYyyyMmDd(tomorrow);
}

// === TIMEZONE UTILITIES ===

/**
 * Convert UTC date to SAST (for debugging)
 */
export function convertUTCToSAST(utcDateString: string): string {
	const date = safeParseDate(utcDateString);
	if (!date) return 'Invalid Date';

	// Add 2 hours for SAST offset
	const sastDate = new Date(date.getTime() + 2 * 60 * 60 * 1000);
	return sastDate.toISOString();
}

/**
 * Check if timezone operations should use SAST
 */
export function shouldUseSAST(): boolean {
	// Always use SAST for this South African application
	return true;
}

// === LEGACY COMPATIBILITY ===
// Export aliases for backward compatibility during migration

/** @deprecated Use formatTimeSlot instead */
export const formatTimeSlotRange = formatTimeSlot;

/** @deprecated Use formatShiftTime instead */
export const formatShiftTimeSAST = formatShiftTime;
