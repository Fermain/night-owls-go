import { createMutation, useQueryClient } from '@tanstack/svelte-query';
import { toast } from 'svelte-sonner';
import { UsersApiService } from '$lib/services/api';

interface BulkDeleteResponse {
	message: string;
	deleted_count: number;
}

export function createBulkDeleteUsersMutation() {
	const queryClient = useQueryClient();
	return createMutation<BulkDeleteResponse, Error, number[]>({
		mutationFn: async (userIds) => {
			return await UsersApiService.bulkDelete(userIds);
		},
		onSuccess: (data) => {
			toast.success(`Successfully deleted ${data.deleted_count} users!`);
			queryClient.invalidateQueries({ queryKey: ['adminUsers'] });
		},
		onError: (error) => {
			toast.error(`Bulk delete error: ${error.message}`);
		}
	});
}
