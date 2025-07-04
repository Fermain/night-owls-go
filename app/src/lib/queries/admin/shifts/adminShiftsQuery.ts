import { createQuery } from '@tanstack/svelte-query';
import { SchedulesApiService } from '$lib/services/api';
import type { AdminShiftSlot } from '$lib/types';
import { createDateRange } from '$lib/utils/dateFormatting';

/**
 * Query for admin shifts calendar view - shows all shifts (filled and unfilled)
 */
export function createAdminShiftsQuery(dayRange: string = '14', enabled: boolean = true) {
	return createQuery<AdminShiftSlot[], Error>({
		queryKey: ['adminShifts', dayRange],
		queryFn: async () => {
			const days = parseInt(dayRange);
			const { from, to } = createDateRange(days);
			const shifts = await SchedulesApiService.getAllSlots({ from, to });

			// Sort by start time for consistent display
			return shifts.sort(
				(a, b) => new Date(a.start_time).getTime() - new Date(b.start_time).getTime()
			);
		},
		enabled,
		staleTime: 1000 * 60 * 3, // 3 minutes
		gcTime: 1000 * 60 * 10, // 10 minutes
		retry: 2
	});
}
