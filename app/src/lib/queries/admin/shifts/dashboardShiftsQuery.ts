import { createQuery } from '@tanstack/svelte-query';
import { SchedulesApiService } from '$lib/services/api';
import type { AdminShiftSlot } from '$lib/types';
import { createDateRange } from '$lib/utils/dateFormatting';

/**
 * Query for dashboard shifts data (next 90 days)
 */
export function createDashboardShiftsQuery(enabled: boolean = true) {
	return createQuery<AdminShiftSlot[], Error>({
		queryKey: ['dashboardShifts'],
		queryFn: async () => {
			const { from, to } = createDateRange(90); // 90 days for dashboard
			return await SchedulesApiService.getAllSlots({ from, to });
		},
		enabled,
		staleTime: 1000 * 60 * 5, // 5 minutes
		gcTime: 1000 * 60 * 10, // 10 minutes
		retry: 2
	});
}
