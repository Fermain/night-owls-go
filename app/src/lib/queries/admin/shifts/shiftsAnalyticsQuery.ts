import { createQuery } from '@tanstack/svelte-query';
import { getCurrentUser } from '$lib/utils/auth';
import { SchedulesApiService } from '$lib/services/api';
import type { AdminShiftSlot } from '$lib/types';
import { createDateRange } from '$lib/utils/dateFormatting';

export interface ShiftAnalytics {
	// Slot data from schedules API
	shifts: AdminShiftSlot[];
	
	// Calculated metrics
	metrics: {
		totalShifts: number;
		filledShifts: number;
		availableShifts: number;
		fillRate: number;
		
		// Time-based analysis
		todayShifts: number;
		tomorrowShifts: number;
		thisWeekShifts: number;
		nextWeekShifts: number;
		
		// Urgency analysis
		urgentUnfilled: number; // Unfilled shifts within 24 hours
		criticalUnfilled: number; // Unfilled shifts within 72 hours
		
		// Schedule breakdown
		scheduleBreakdown: Array<{
			schedule_id: number;
			schedule_name: string;
			total: number;
			filled: number;
			available: number;
			fillRate: number;
		}>;
		
		// Time slot patterns
		timeSlotAnalysis: Array<{
			hour: number;
			dayOfWeek: number;
			dayName: string;
			timeSlot: string;
			total: number;
			filled: number;
			fillRate: number;
		}>;
		
		// Recent activity
		recentBookings: AdminShiftSlot[];
		upcomingShifts: AdminShiftSlot[];
	};
}

async function fetchShiftsAnalytics(days: number = 30): Promise<ShiftAnalytics> {
	const user = getCurrentUser();
	if (!user?.token) {
		throw new Error('Not authenticated');
	}

	// Get comprehensive shift data
	const { from, to } = createDateRange(days);
	const shifts = await SchedulesApiService.getAllSlots({ from, to });

	// Calculate comprehensive metrics
	const now = new Date();
	const tomorrow = new Date(now.getTime() + 24 * 60 * 60 * 1000);
	const thisWeekEnd = new Date(now.getTime() + 7 * 24 * 60 * 60 * 1000);
	const nextWeekEnd = new Date(now.getTime() + 14 * 24 * 60 * 60 * 1000);

	const totalShifts = shifts.length;
	const filledShifts = shifts.filter(s => s.is_booked).length;
	const availableShifts = totalShifts - filledShifts;
	const fillRate = totalShifts > 0 ? (filledShifts / totalShifts) * 100 : 0;

	// Time-based metrics
	const todayShifts = shifts.filter(s => {
		const shiftDate = new Date(s.start_time);
		return shiftDate.toDateString() === now.toDateString();
	}).length;

	const tomorrowShifts = shifts.filter(s => {
		const shiftDate = new Date(s.start_time);
		return shiftDate.toDateString() === tomorrow.toDateString();
	}).length;

	const thisWeekShifts = shifts.filter(s => {
		const shiftDate = new Date(s.start_time);
		return shiftDate >= now && shiftDate <= thisWeekEnd;
	}).length;

	const nextWeekShifts = shifts.filter(s => {
		const shiftDate = new Date(s.start_time);
		return shiftDate > thisWeekEnd && shiftDate <= nextWeekEnd;
	}).length;

	// Urgency analysis
	const urgentUnfilled = shifts.filter(s => {
		const shiftDate = new Date(s.start_time);
		return !s.is_booked && shiftDate >= now && shiftDate <= tomorrow;
	}).length;

	const criticalTime = new Date(now.getTime() + 72 * 60 * 60 * 1000);
	const criticalUnfilled = shifts.filter(s => {
		const shiftDate = new Date(s.start_time);
		return !s.is_booked && shiftDate >= now && shiftDate <= criticalTime;
	}).length;

	// Schedule breakdown
	const scheduleMap = new Map<string, {
		schedule_id: number;
		schedule_name: string;
		total: number;
		filled: number;
	}>();

	shifts.forEach(shift => {
		const key = `${shift.schedule_id}-${shift.schedule_name}`;
		if (!scheduleMap.has(key)) {
			scheduleMap.set(key, {
				schedule_id: shift.schedule_id,
				schedule_name: shift.schedule_name,
				total: 0,
				filled: 0
			});
		}
		const schedule = scheduleMap.get(key)!;
		schedule.total++;
		if (shift.is_booked) schedule.filled++;
	});

	const scheduleBreakdown = Array.from(scheduleMap.values()).map(s => ({
		...s,
		available: s.total - s.filled,
		fillRate: s.total > 0 ? (s.filled / s.total) * 100 : 0
	})).sort((a, b) => a.schedule_name.localeCompare(b.schedule_name));

	// Time slot analysis
	const timeSlotMap = new Map<string, {
		hour: number;
		dayOfWeek: number;
		dayName: string;
		timeSlot: string;
		total: number;
		filled: number;
	}>();

	const dayNames = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];

	shifts.forEach(shift => {
		const shiftDate = new Date(shift.start_time);
		const hour = shiftDate.getUTCHours();
		const dayOfWeek = shiftDate.getUTCDay();
		const key = `${dayOfWeek}-${hour}`;
		
		if (!timeSlotMap.has(key)) {
			const formatHour = (h: number) => {
				if (h === 0) return '12AM';
				if (h < 12) return `${h}AM`;
				if (h === 12) return '12PM';
				return `${h - 12}PM`;
			};
			
			timeSlotMap.set(key, {
				hour,
				dayOfWeek,
				dayName: dayNames[dayOfWeek],
				timeSlot: formatHour(hour),
				total: 0,
				filled: 0
			});
		}
		const slot = timeSlotMap.get(key)!;
		slot.total++;
		if (shift.is_booked) slot.filled++;
	});

	const timeSlotAnalysis = Array.from(timeSlotMap.values()).map(s => ({
		...s,
		fillRate: s.total > 0 ? (s.filled / s.total) * 100 : 0
	})).sort((a, b) => {
		if (a.dayOfWeek !== b.dayOfWeek) return a.dayOfWeek - b.dayOfWeek;
		return a.hour - b.hour;
	});

	// Recent activity
	const recentBookings = shifts
		.filter(s => s.is_booked)
		.sort((a, b) => new Date(b.start_time).getTime() - new Date(a.start_time).getTime())
		.slice(0, 10);

	const upcomingShifts = shifts
		.filter(s => {
			const shiftDate = new Date(s.start_time);
			return shiftDate >= now;
		})
		.sort((a, b) => new Date(a.start_time).getTime() - new Date(b.start_time).getTime())
		.slice(0, 15);

	return {
		shifts,
		metrics: {
			totalShifts,
			filledShifts,
			availableShifts,
			fillRate,
			todayShifts,
			tomorrowShifts,
			thisWeekShifts,
			nextWeekShifts,
			urgentUnfilled,
			criticalUnfilled,
			scheduleBreakdown,
			timeSlotAnalysis,
			recentBookings,
			upcomingShifts
		}
	};
}

export function createShiftsAnalyticsQuery(days: number = 30) {
	return createQuery({
		queryKey: ['shifts', 'analytics', days],
		queryFn: () => fetchShiftsAnalytics(days),
		refetchInterval: 30000, // Refetch every 30 seconds for real-time updates
		staleTime: 15000 // Consider data stale after 15 seconds
	});
} 