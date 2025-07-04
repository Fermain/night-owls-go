/**
 * Admin Dialog Utilities
 *
 * Shared utilities for admin dialog components, focusing on date/time
 * formatting and display logic for shift assignment and details dialogs.
 */

import { formatTimeSlot } from '$lib/utils/dateFormatting';
import { getRelativeTime } from '$lib/utils/shifts';
import type { AdminShiftSlot } from '$lib/types';

// === DIALOG DATE FORMATTING ===

/**
 * Comprehensive date/time information for dialog headers
 */
export interface DialogDateInfo {
	fullDate: string; // "Wednesday, January 15, 2025"
	timeRange: string; // "18:00 - 06:00"
	relative: string; // "in 2 days", "tomorrow", "in 3 hours"
}

/**
 * Generate comprehensive date/time information for admin dialogs
 * Provides consistent formatting across all admin dialog components
 */
export function generateDialogDateInfo(shift: AdminShiftSlot): DialogDateInfo {
	const date = new Date(shift.start_time);

	return {
		// Full date: "Wednesday, January 15, 2025"
		fullDate: date.toLocaleDateString('en-US', {
			weekday: 'long',
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		}),
		// Time range: "18:00 - 06:00"
		timeRange: formatTimeSlot(shift.start_time, shift.end_time),
		// Relative: "in 2 days", "tomorrow", "in 3 hours"
		relative: getRelativeTime(shift.start_time)
	};
}

// === SHIFT STATUS ANALYSIS ===

/**
 * Determine if a shift is properly filled for admin context
 */
export function isShiftFilled(shift: AdminShiftSlot): boolean {
	return shift.is_booked && !!shift.user_name;
}

/**
 * Determine if a shift has data inconsistencies
 */
export function hasDataInconsistency(shift: AdminShiftSlot): boolean {
	return shift.is_booked && !shift.user_name;
}

/**
 * Get appropriate status badge info for admin dialogs
 */
export interface StatusBadgeInfo {
	text: string;
	variant: 'default' | 'secondary' | 'destructive' | 'outline';
	className: string;
}

export function getStatusBadgeInfo(shift: AdminShiftSlot): StatusBadgeInfo {
	if (isShiftFilled(shift)) {
		return {
			text: 'Assigned',
			variant: 'default',
			className: 'bg-green-100 text-green-700 border-green-200'
		};
	}

	if (hasDataInconsistency(shift)) {
		return {
			text: 'Data Issue',
			variant: 'secondary',
			className: 'bg-yellow-100 text-yellow-700 border-yellow-200'
		};
	}

	return {
		text: 'Needs Assignment',
		variant: 'secondary',
		className: 'bg-orange-100 text-orange-700 border-orange-200'
	};
}

// === DIALOG VALIDATION ===

/**
 * Validate shift data for dialog display
 */
export function validateShiftForDialog(shift: AdminShiftSlot | null): boolean {
	if (!shift) return false;

	// Basic required fields
	if (!shift.start_time || !shift.end_time || !shift.schedule_name) {
		return false;
	}

	// Validate dates
	const startDate = new Date(shift.start_time);
	const endDate = new Date(shift.end_time);

	if (isNaN(startDate.getTime()) || isNaN(endDate.getTime())) {
		return false;
	}

	if (startDate >= endDate) {
		return false;
	}

	return true;
}

/**
 * Generate error message for invalid shifts
 */
export function getShiftValidationError(shift: AdminShiftSlot | null): string | null {
	if (!shift) return 'No shift data provided';

	if (!shift.start_time || !shift.end_time) {
		return 'Missing shift timing information';
	}

	if (!shift.schedule_name) {
		return 'Missing schedule information';
	}

	const startDate = new Date(shift.start_time);
	const endDate = new Date(shift.end_time);

	if (isNaN(startDate.getTime()) || isNaN(endDate.getTime())) {
		return 'Invalid shift timing data';
	}

	if (startDate >= endDate) {
		return 'Shift end time must be after start time';
	}

	return null;
}
