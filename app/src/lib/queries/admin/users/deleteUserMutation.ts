import { createMutation, useQueryClient } from '@tanstack/svelte-query';
import { goto } from '$app/navigation';
import { toast } from 'svelte-sonner';

// Assuming the API returns a success message or just a status code
interface DeleteUserResponse {
    message?: string;
    // Potentially other fields if the API returns more info
}

export function createDeleteUserMutation(onSettled?: () => void) {
    const queryClient = useQueryClient();
    return createMutation<DeleteUserResponse, Error, number>({
        mutationFn: async (userIdToDelete) => {
            const response = await fetch(`/api/admin/users/${userIdToDelete}`, {
                method: 'DELETE'
            });
            if (!response.ok) {
                const errorData = await response.json().catch(() => ({ message: 'Failed to delete user' }));
                throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
            }
            // For DELETE, response might be empty (204) or have a message (200)
            // Try to parse JSON, but handle cases where it might be empty.
            const contentType = response.headers.get("content-type");
            if (contentType && contentType.indexOf("application/json") !== -1) {
                return response.json() as Promise<DeleteUserResponse>;
            }
            return { message: 'User deleted successfully' }; // Or an empty object if preferred for non-JSON responses
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