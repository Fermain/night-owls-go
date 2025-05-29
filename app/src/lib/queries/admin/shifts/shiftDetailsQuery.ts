import { createQuery } from '@tanstack/svelte-query';
import { SchedulesApiService } from '$lib/services/api';
import type { AdminShiftSlot } from '$lib/types';

/**
 * Fetch shift details by matching start time
 */
async function fetchShiftDetails(shiftStartTime: string): Promise<AdminShiftSlot | null> {
	// Get shifts for a wider range to find the specific shift
	const shiftDate = new Date(shiftStartTime);
	const dayStart = new Date(shiftDate);
	dayStart.setUTCHours(0, 0, 0, 0);
	const dayEnd = new Date(shiftDate);
	dayEnd.setUTCHours(23, 59, 59, 999);

	const shifts = await SchedulesApiService.getAllSlots({
		from: dayStart.toISOString(),
		to: dayEnd.toISOString()
	});

	// Find the exact shift by start time
	const matchingShift = shifts.find((shift) => shift.start_time === shiftStartTime);

	return matchingShift || null;
}

/**
 * Query for specific shift details by start time
 */
export function createShiftDetailsQuery(shiftStartTime: string | null) {
	return createQuery<AdminShiftSlot | null, Error>({
		queryKey: ['shiftDetails', shiftStartTime || ''],
		queryFn: () => fetchShiftDetails(shiftStartTime!),
		enabled: !!shiftStartTime,
		staleTime: 1000 * 60 * 2, // 2 minutes
		retry: 1
	});
}
