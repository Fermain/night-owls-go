import { createMutation, useQueryClient } from '@tanstack/svelte-query';
import { goto } from '$app/navigation';
import { toast } from 'svelte-sonner';
import { UsersApiService } from '$lib/services/api';

// Assuming the API returns a success message or just a status code
interface DeleteUserResponse {
	message?: string;
	// Potentially other fields if the API returns more info
}

export function createDeleteUserMutation(onSettled?: () => void) {
	const queryClient = useQueryClient();
	return createMutation<DeleteUserResponse, Error, number>({
		mutationFn: async (userIdToDelete) => {
			return await UsersApiService.delete(userIdToDelete);
		},
		onSuccess: async () => {
			toast.success('User deleted successfully!');
			await queryClient.invalidateQueries({ queryKey: ['adminUsers'] });
			goto('/admin/users');
		},
		onError: (error) => {
			toast.error(`Error deleting user: ${error.message}`);
		},
		onSettled: () => {
			onSettled?.();
		}
	});
}
