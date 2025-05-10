import {
	CalendarDate,
	type DateValue,
	parseAbsoluteToLocal,
	parseDate as parseDateStringToDateValue
} from '@internationalized/date';

/**
 * Parses a YYYY-MM-DD string into a DateValue object (specifically CalendarDate).
 * Returns undefined if the input string is null, undefined, or invalid.
 */
export function parseYyyyMmDdToCalendarDate(
	dateString: string | null | undefined
): CalendarDate | undefined {
	if (!dateString) {
		return undefined;
	}
	try {
		// parseDateStringToDateValue expects "YYYY-MM-DD"
		const parsed = parseDateStringToDateValue(dateString);
		return new CalendarDate(parsed.year, parsed.month, parsed.day);
	} catch (e) {
		console.error(`[parseYyyyMmDdToCalendarDate] Failed to parse date string "${dateString}":`, e);
		return undefined;
	}
}

/**
 * Formats a DateValue object (like CalendarDate) into a YYYY-MM-DD string.
 * Returns null if the input DateValue is null or undefined.
 */
export function formatCalendarDateToYyyyMmDd(
	dateValue: DateValue | null | undefined
): string | null {
	if (!dateValue) {
		return null;
	}
	// DateValue objects have a toString() method that returns "YYYY-MM-DD"
	return dateValue.toString();
}

/**
 * Parses an ISO 8601 string (e.g., "YYYY-MM-DDTHH:mm:ssZ") and returns only the "YYYY-MM-DD" part.
 * Returns null if the input string is null, undefined, or doesn't contain 'T'.
 * If the string doesn't contain 'T' but is otherwise valid, it's returned as is.
 */
export function parseIsoToYyyyMmDd(isoString: string | null | undefined): string | null {
	if (isoString === null || isoString === undefined) {
		return null;
	}
	if (typeof isoString === 'string') {
		if (isoString.includes('T')) {
			return isoString.split('T')[0];
		}
		// If it's already in YYYY-MM-DD format or some other non-ISO string, return as is for now.
		// Further validation might be needed depending on source.
		return isoString;
	}
	return null;
}

/**
 * Converts a JavaScript Date object to a YYYY-MM-DD string.
 * Returns null if the input Date is null or undefined.
 */
export function formatJsDateToYyyyMmDd(date: Date | null | undefined): string | null {
	if (!date) {
		return null;
	}
	const year = date.getFullYear();
	const month = (date.getMonth() + 1).toString().padStart(2, '0');
	const day = date.getDate().toString().padStart(2, '0');
	return `${year}-${month}-${day}`;
}

/**
 * Converts an ISO date string (potentially with time and timezone) to a DateValue (CalendarDate)
 * that represents the local date.
 * This is useful when the backend sends a full ISO timestamp, but the calendar should
 * operate on the date part in the user's local timezone interpretation of that timestamp.
 */
export function parseAbsoluteIsoToLocalDate(
	isoString: string | null | undefined
): CalendarDate | undefined {
	if (!isoString) {
		return undefined;
	}
	try {
		// parseAbsoluteToLocal converts the ISO string (absolute time) to the local time zone
		// and returns a ZonedDateTime. We then extract the date part.
		const localDateTime = parseAbsoluteToLocal(isoString);
		return new CalendarDate(localDateTime.year, localDateTime.month, localDateTime.day);
	} catch (e) {
		console.error(`[parseAbsoluteIsoToLocalDate] Failed to parse ISO string "${isoString}":`, e);
		// Fallback or alternative parsing if direct parseAbsoluteToLocal fails or is not desired.
		// For instance, if we strictly want to truncate to YYYY-MM-DD from the string:
		const yyyyMmDdPart = parseIsoToYyyyMmDd(isoString);
		if (yyyyMmDdPart) {
			return parseYyyyMmDdToCalendarDate(yyyyMmDdPart);
		}
		return undefined;
	}
}
