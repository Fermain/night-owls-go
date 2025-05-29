import { createQuery } from '@tanstack/svelte-query';
import { SchedulesApiService } from '$lib/services/api';
import type { Schedule } from '$lib/types';

export function createSchedulesQuery() {
	return createQuery<Schedule[], Error>({
		queryKey: ['adminSchedules'],
		queryFn: () => SchedulesApiService.getAll(),
		staleTime: 1000 * 60 * 5, // 5 minutes
		gcTime: 1000 * 60 * 10, // 10 minutes
		retry: 2
	});
}
