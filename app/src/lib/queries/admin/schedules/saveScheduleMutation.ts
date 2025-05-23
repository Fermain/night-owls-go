import { createMutation, useQueryClient } from '@tanstack/svelte-query';
import { goto } from '$app/navigation';
import { toast } from 'svelte-sonner';
import { selectedScheduleForForm } from '$lib/stores/scheduleEditingStore';
import { SchedulesApiService } from '$lib/services/api';

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

export function createSaveScheduleMutation(onSuccessCallback?: () => void) {
	const queryClient = useQueryClient();
	return createMutation<ScheduleResponse, Error, { payload: SchedulePayload; scheduleId?: number }>(
		{
			mutationFn: async ({ payload, scheduleId }) => {
				if (scheduleId) {
					return await SchedulesApiService.update(scheduleId, payload);
				} else {
					return await SchedulesApiService.create(payload);
				}
			},
			onSuccess: async (_data, { scheduleId }) => {
				toast.success(`Schedule ${scheduleId ? 'updated' : 'created'} successfully!`);
				await queryClient.invalidateQueries({ queryKey: ['adminSchedules'] });
				await queryClient.invalidateQueries({ queryKey: ['adminSchedulesForLayout'] });
				if (scheduleId) {
					selectedScheduleForForm.set(undefined);
				}
				
				// Use callback if provided, otherwise default navigation
				if (onSuccessCallback) {
					onSuccessCallback();
				} else {
					goto('/admin/schedules');
				}
			},
			onError: (error) => {
				toast.error(`Save Error: ${error.message}`);
			}
		}
	);
}
