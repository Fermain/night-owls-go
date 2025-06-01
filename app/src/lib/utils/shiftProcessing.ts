import type { AdminShiftSlot } from '$lib/types';

export interface ScheduleMetric {
	schedule: string;
	total: number;
	filled: number;
	fillRate: number;
}

export interface TimeSlotMetric {
	timeSlot: string;
	total: number;
	fillRate: number;
}

export interface DashboardMetrics {
	totalShifts: number;
	filledShifts: number;
	availableShifts: number;
	fillRate: number;
	scheduleData: ScheduleMetric[];
	timeSlotData: TimeSlotMetric[];
	fillRateData: { label: string; value: number }[];
}

/**
 * Process shift data to generate schedule-based metrics
 */
export function processScheduleData(shifts: AdminShiftSlot[]): ScheduleMetric[] {
	const scheduleStats = new Map<string, { total: number; filled: number; name: string }>();

	shifts.forEach((shift) => {
		const key = shift.schedule_id.toString();
		if (!scheduleStats.has(key)) {
			scheduleStats.set(key, { total: 0, filled: 0, name: shift.schedule_name });
		}
		const stats = scheduleStats.get(key)!;
		stats.total += 1;
		if (shift.is_booked) stats.filled += 1;
	});

	return Array.from(scheduleStats.entries())
		.map(([_id, stats]) => ({
			schedule: stats.name,
			total: stats.total,
			filled: stats.filled,
			fillRate: stats.total > 0 ? Math.round((stats.filled / stats.total) * 100) : 0
		}))
		.sort((a, b) => b.total - a.total);
}

/**
 * Process shift data to generate time slot metrics
 */
export function processTimeSlotData(shifts: AdminShiftSlot[]): TimeSlotMetric[] {
	const timeSlots = new Map<string, { total: number; filled: number }>();

	shifts.forEach((shift) => {
		const start = new Date(shift.start_time);
		const hour = start.getUTCHours();
		const timeSlotLabels = [
			'00:00-02:00',
			'02:00-04:00',
			'04:00-06:00',
			'06:00-08:00',
			'08:00-10:00',
			'10:00-12:00',
			'12:00-14:00',
			'14:00-16:00',
			'16:00-18:00',
			'18:00-20:00',
			'20:00-22:00',
			'22:00-24:00'
		];
		const slotIndex = Math.floor(hour / 2);
		const slot = timeSlotLabels[slotIndex] || timeSlotLabels[timeSlotLabels.length - 1];

		if (!timeSlots.has(slot)) {
			timeSlots.set(slot, { total: 0, filled: 0 });
		}
		const stats = timeSlots.get(slot)!;
		stats.total += 1;
		if (shift.is_booked) stats.filled += 1;
	});

	return Array.from(timeSlots.entries())
		.map(([timeSlot, stats]) => ({
			timeSlot,
			total: stats.total,
			fillRate: stats.total > 0 ? Math.round((stats.filled / stats.total) * 100) : 0
		}))
		.filter((item) => item.total > 0)
		.sort((a, b) => a.timeSlot.localeCompare(b.timeSlot));
}

/**
 * Calculate comprehensive dashboard metrics from shift data
 */
export function calculateDashboardMetrics(shifts: AdminShiftSlot[]): DashboardMetrics | null {
	if (shifts.length === 0) return null;

	const totalShifts = shifts.length;
	const filledShifts = shifts.filter((s) => s.is_booked).length;
	const availableShifts = totalShifts - filledShifts;
	const fillRate = totalShifts > 0 ? Math.round((filledShifts / totalShifts) * 100) : 0;

	return {
		totalShifts,
		filledShifts,
		availableShifts,
		fillRate,
		scheduleData: processScheduleData(shifts),
		timeSlotData: processTimeSlotData(shifts),
		fillRateData: [
			{ label: 'Filled', value: filledShifts },
			{ label: 'Available', value: availableShifts }
		]
	};
}
