import { createMutation, useQueryClient } from '@tanstack/svelte-query';
import { toast } from 'svelte-sonner';
import { authenticatedFetch } from '$lib/utils/api';

interface BulkDeleteUsersResponse {
	message: string;
}

interface BulkDeleteUsersRequest {
	user_ids: number[];
}

export function createBulkDeleteUsersMutation(onSettled?: () => void) {
	const queryClient = useQueryClient();
	return createMutation<BulkDeleteUsersResponse, Error, number[]>({
		mutationFn: async (userIds) => {
			const payload: BulkDeleteUsersRequest = { user_ids: userIds };
			
			const response = await authenticatedFetch('/api/admin/users/bulk-delete', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify(payload)
			});
			
			if (!response.ok) {
				const errorData = await response.json().catch(() => ({ 
					message: 'Failed to bulk delete users' 
				}));
				throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
			}
			
			return (await response.json()) as BulkDeleteUsersResponse;
		},
		onSuccess: async (data, userIds) => {
			toast.success(data.message || `Successfully deleted ${userIds.length} users`);
			await queryClient.invalidateQueries({ queryKey: ['adminUsers'] });
		},
		onError: (error, userIds) => {
			toast.error(`Error deleting ${userIds.length} users: ${error.message}`);
		},
		onSettled: () => {
			onSettled?.();
		}
	});
} 