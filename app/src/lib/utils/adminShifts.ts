import type { AdminShiftSlot } from '$lib/types';

/**
 * Admin-specific shift status analysis
 * Provides semantic meaning for admin workflow context
 */
export interface AdminShiftStatus {
	status: 'filled' | 'unfilled' | 'inconsistent' | 'past';
	color: 'green' | 'red' | 'yellow' | 'gray';
	urgency: 'none' | 'low' | 'medium' | 'high';
	label: string;
	icon: '✓' | '!' | '?' | '⏷';
	needsAttention: boolean;
}

/**
 * Analyze shift status from admin perspective
 * Different from user perspective - focuses on operational needs
 */
export function getAdminShiftStatus(shift: AdminShiftSlot): AdminShiftStatus {
	// Check if shift is in the past
	const now = new Date();
	const shiftTime = new Date(shift.start_time);
	const isPast = shiftTime < now;

	if (isPast) {
		return {
			status: 'past',
			color: 'gray',
			urgency: 'none',
			label: 'Past Shift',
			icon: '⏷',
			needsAttention: false
		};
	}

	// Check if shift is properly filled
	if (shift.is_booked && shift.user_name) {
		const label = shift.buddy_name ? `${shift.user_name} + ${shift.buddy_name}` : shift.user_name;

		return {
			status: 'filled',
			color: 'green',
			urgency: 'none',
			label,
			icon: '✓',
			needsAttention: false
		};
	}

	// Check for data inconsistency (marked as booked but no user assigned)
	if (shift.is_booked && !shift.user_name) {
		return {
			status: 'inconsistent',
			color: 'yellow',
			urgency: 'medium',
			label: 'Data Issue - Review Needed',
			icon: '?',
			needsAttention: true
		};
	}

	// Shift is unfilled - determine urgency based on timing
	const hoursUntilShift = (shiftTime.getTime() - now.getTime()) / (1000 * 60 * 60);

	let urgency: 'low' | 'medium' | 'high' = 'low';
	if (hoursUntilShift <= 24) {
		urgency = 'high';
	} else if (hoursUntilShift <= 72) {
		urgency = 'medium';
	}

	return {
		status: 'unfilled',
		color: 'red',
		urgency,
		label: 'Needs Assignment',
		icon: '!',
		needsAttention: true
	};
}

/**
 * Get CSS classes for admin shift display
 */
export function getAdminShiftClasses(status: AdminShiftStatus, baseClasses: string = ''): string {
	const colorClasses = {
		green: 'bg-green-500 text-white border-green-600 hover:bg-green-600',
		red: 'bg-red-500 text-white border-red-600 hover:bg-red-600',
		yellow: 'bg-yellow-500 text-white border-yellow-600 hover:bg-yellow-600',
		gray: 'bg-gray-400 text-gray-600 opacity-50'
	};

	return `${baseClasses} ${colorClasses[status.color]} border shadow-sm`;
}

/**
 * Filter shifts that need admin attention
 */
export function getShiftsNeedingAttention(shifts: AdminShiftSlot[]): AdminShiftSlot[] {
	return shifts.filter((shift) => {
		const status = getAdminShiftStatus(shift);
		return status.needsAttention;
	});
}

/**
 * Group shifts by status for admin dashboard
 */
export function groupShiftsByStatus(shifts: AdminShiftSlot[]) {
	const groups = {
		filled: [] as AdminShiftSlot[],
		unfilled: [] as AdminShiftSlot[],
		inconsistent: [] as AdminShiftSlot[],
		past: [] as AdminShiftSlot[]
	};

	shifts.forEach((shift) => {
		const status = getAdminShiftStatus(shift);
		groups[status.status].push(shift);
	});

	return groups;
}

/**
 * Calculate admin dashboard metrics
 */
export function calculateAdminMetrics(shifts: AdminShiftSlot[]) {
	const groups = groupShiftsByStatus(shifts);
	const total = shifts.length;

	if (total === 0) {
		return {
			totalShifts: 0,
			filledCount: 0,
			unfilledCount: 0,
			inconsistentCount: 0,
			fillRate: 0,
			urgentCount: 0
		};
	}

	const filledCount = groups.filled.length;
	const unfilledCount = groups.unfilled.length;
	const inconsistentCount = groups.inconsistent.length;

	// Count urgent unfilled shifts (within 24 hours)
	const urgentCount = groups.unfilled.filter((shift) => {
		const status = getAdminShiftStatus(shift);
		return status.urgency === 'high';
	}).length;

	return {
		totalShifts: total,
		filledCount,
		unfilledCount,
		inconsistentCount,
		fillRate: Math.round((filledCount / total) * 100),
		urgentCount
	};
}
