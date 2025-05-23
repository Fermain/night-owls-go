import { createMutation, useQueryClient } from '@tanstack/svelte-query';
import { goto } from '$app/navigation';
import { toast } from 'svelte-sonner';
import { selectedScheduleForForm } from '$lib/stores/scheduleEditingStore';
import { SchedulesApiService } from '$lib/services/api';

// Define more specific types if possible
interface DeleteResponse {
	// Define based on actual API response, assuming it might just be a success status or simple message
	ok: boolean;
	message?: string;
}

export function createDeleteScheduleMutation(onFinally?: () => void, onSuccessCallback?: () => void) {
	const queryClient = useQueryClient();
	return createMutation<DeleteResponse, Error, number>({
		mutationFn: async (id) => {
			return await SchedulesApiService.delete(id);
		},
		onSuccess: async () => {
			toast.success('Schedule deleted successfully!');
			await queryClient.invalidateQueries({ queryKey: ['adminSchedules'] });
			await queryClient.invalidateQueries({ queryKey: ['adminSchedulesForLayout'] });
			selectedScheduleForForm.set(undefined);
			
			// Use callback if provided, otherwise default navigation
			if (onSuccessCallback) {
				onSuccessCallback();
			} else {
				goto('/admin/schedules');
			}
			onFinally?.();
		},
		onError: (error) => {
			toast.error(`Delete Error: ${error.message}`);
			onFinally?.();
		}
	});
}
