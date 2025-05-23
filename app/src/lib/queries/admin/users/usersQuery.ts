import { createQuery } from '@tanstack/svelte-query';
import { UsersApiService } from '$lib/services/api';
import type { UserData } from '$lib/schemas/user';

/**
 * Query for all users data
 */
export function createUsersQuery(searchTerm?: string) {
	return createQuery<UserData[], Error>({
		queryKey: ['adminUsers', searchTerm || ''],
		queryFn: () => UsersApiService.getAll(),
		staleTime: 1000 * 60 * 5, // 5 minutes
		gcTime: 1000 * 60 * 10, // 10 minutes
		retry: 2
	});
} 