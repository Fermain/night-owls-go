import { createMutation, useQueryClient } from '@tanstack/svelte-query';
import { goto } from '$app/navigation';
import { toast } from 'svelte-sonner';
import { selectedScheduleForForm } from '$lib/stores/scheduleEditingStore';
import { authenticatedFetch } from '$lib/utils/api';

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
			const response = await authenticatedFetch(`/api/admin/schedules/${id}`, { method: 'DELETE' });
			if (!response.ok) {
				const errorData = await response
					.json()
					.catch(() => ({ message: 'Failed to delete schedule' }));
				throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
			}
			// Assuming the response for a successful DELETE might be empty or just a status.
			// If it returns JSON, parse it. Otherwise, construct a success object.
			try {
				return (await response.json()) as DeleteResponse;
			} catch (e) {
				// Handle cases where response might not be JSON (e.g., 204 No Content)
				return { ok: response.ok };
			}
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
