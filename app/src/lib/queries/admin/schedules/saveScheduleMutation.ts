import { createMutation, useQueryClient } from '@tanstack/svelte-query';
import { goto } from '$app/navigation';
import { toast } from 'svelte-sonner';
import { selectedScheduleForForm } from '$lib/stores/scheduleEditingStore';
import { authenticatedFetch } from '$lib/utils/api';

// Define more specific types if possible
interface SchedulePayload {
	name: string;
	cron_expr: string;
	start_date: string | null;
	end_date: string | null;
}

interface ScheduleResponse {
	// Define based on actual API response
	message: string;
	schedule_id?: number;
}

export function createSaveScheduleMutation() {
	const queryClient = useQueryClient();
	return createMutation<ScheduleResponse, Error, { payload: SchedulePayload; scheduleId?: number }>(
		{
			mutationFn: async ({ payload, scheduleId }) => {
				const url = scheduleId ? `/api/admin/schedules/${scheduleId}` : '/api/admin/schedules';
				const method = scheduleId ? 'PUT' : 'POST';
				const response = await authenticatedFetch(url, {
					method,
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify(payload)
				});
				if (!response.ok) {
					const errorData = await response
						.json()
						.catch(() => ({ message: `Failed to ${scheduleId ? 'update' : 'create'} schedule` }));
					throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
				}
				return response.json() as Promise<ScheduleResponse>;
			},
			onSuccess: async (_data, { scheduleId }) => {
				toast.success(`Schedule ${scheduleId ? 'updated' : 'created'} successfully!`);
				await queryClient.invalidateQueries({ queryKey: ['adminSchedules'] });
				await queryClient.invalidateQueries({ queryKey: ['adminSchedulesForLayout'] });
				if (scheduleId) {
					selectedScheduleForForm.set(undefined);
				}
				goto('/admin/schedules');
			},
			onError: (error) => {
				toast.error(`Save Error: ${error.message}`);
			}
		}
	);
}
