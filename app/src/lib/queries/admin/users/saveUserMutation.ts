import { createMutation, useQueryClient } from '@tanstack/svelte-query';
import { goto } from '$app/navigation';
import { toast } from 'svelte-sonner';
import type { E164Number } from 'svelte-tel-input/types';
import { UsersApiService } from '$lib/services/api';

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

			// Convert E164Number to string
			const apiPayload = {
				name: payload.name || '',
				phone: payload.phone.toString(),
				role: payload.role
			};

			console.log('saveUserMutation - API Payload:', apiPayload);

			if (isEditMode && userId) {
				return await UsersApiService.update(userId, apiPayload);
			} else {
				return await UsersApiService.create(apiPayload);
			}
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
