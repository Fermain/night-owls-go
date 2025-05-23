import { createMutation, useQueryClient } from '@tanstack/svelte-query';
import { goto } from '$app/navigation';
import { toast } from 'svelte-sonner';
import type { E164Number } from 'svelte-tel-input/types';
import { authenticatedFetch } from '$lib/utils/api';

interface UserPayload {
	phone: E164Number;
	name: string | null;
	role: 'admin' | 'owl' | 'guest';
}

// Assuming the API returns some user data or a success message
interface UserResponse {
	// Define based on actual API response
	message?: string;
	id?: number;
	// ... other user fields if returned
}

interface SaveUserVariables {
	payload: UserPayload;
	userId?: number;
}

export function createSaveUserMutation() {
	const queryClient = useQueryClient();
	return createMutation<UserResponse, Error, SaveUserVariables>({
		mutationFn: async (vars) => {
			const { payload, userId } = vars;
			const isEditMode = userId !== undefined;

			console.log('saveUserMutation - Variables:', vars);
			console.log('saveUserMutation - userId:', userId);
			console.log('saveUserMutation - isEditMode:', isEditMode);

			const url = isEditMode ? `/api/admin/users/${userId}` : '/api/admin/users';
			const method = isEditMode ? 'PUT' : 'POST';

			console.log('saveUserMutation - URL:', url);
			console.log('saveUserMutation - Method:', method);
			console.log('saveUserMutation - Payload:', payload);

			const response = await authenticatedFetch(url, {
				method: method,
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify(payload)
			});

			if (!response.ok) {
				const errorData = await response.json().catch(() => ({
					message: `Failed to ${isEditMode ? 'update' : 'create'} user`
				}));
				throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
			}
			// It's good practice to return the parsed JSON response
			// even if not strictly used by all onSuccess handlers immediately.
			return response.json() as Promise<UserResponse>;
		},
		onSuccess: async (_data, vars) => {
			const { userId } = vars;
			const isEditMode = userId !== undefined;
			toast.success(`User ${isEditMode ? 'updated' : 'created'} successfully!`);
			await queryClient.invalidateQueries({ queryKey: ['adminUsers'] });
			if (isEditMode && userId) {
				await queryClient.invalidateQueries({ queryKey: ['adminUser', userId] });
			}
			goto('/admin/users');
		},
		onError: (error) => {
			toast.error(`Error: ${error.message}`);
		}
	});
}
