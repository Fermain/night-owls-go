import { createQuery } from '@tanstack/svelte-query';
import { SchedulesApiService } from '$lib/services/api';
import type { AdminShiftSlot } from '$lib/types';
import { createDateRange } from '$lib/utils/dateFormatting';

/**
 * Query for upcoming shifts in the next 2 weeks (for sidebar display)
 */
export function createUpcomingShiftsQuery(enabled: boolean = true) {
	return createQuery<AdminShiftSlot[], Error>({
		queryKey: ['upcomingShifts'],
		queryFn: async () => {
			const { from, to } = createDateRange(14); // 14 days (2 weeks)
			const shifts = await SchedulesApiService.getAllSlots({ from, to });

			// Filter to only upcoming shifts (from now onwards) and sort by start time
			const now = new Date();
			return shifts
				.filter((shift) => new Date(shift.start_time) >= now)
				.sort((a, b) => new Date(a.start_time).getTime() - new Date(b.start_time).getTime())
				.slice(0, 10); // Limit to 10 most upcoming shifts for sidebar
		},
		enabled,
		staleTime: 1000 * 60 * 2, // 2 minutes (more frequent updates for upcoming shifts)
		gcTime: 1000 * 60 * 5, // 5 minutes
		retry: 2
	});
}
