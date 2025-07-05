/**
 * Example service demonstrating the new API client and error handling patterns
 * This shows best practices for modern API integration
 */

import { apiClient, createTypedAPI } from '$lib/utils/apiClient';
import { NotificationService, handleNetworkError, handleAuthError } from '$lib/utils/errorHandling';

// === TYPE DEFINITIONS ===

interface User {
	id: number;
	name: string;
	email: string;
	role: 'admin' | 'owl';
	phone: string;
	createdAt: string;
	[key: string]: unknown; // Allow additional properties for API compatibility
}

interface CreateUserRequest {
	name: string;
	email: string;
	phone: string;
	role: 'admin' | 'owl';
	[key: string]: unknown; // Allow additional properties
}

interface UpdateUserRequest extends Partial<CreateUserRequest> {
	id: number;
}

// === TYPED API CLIENT ===

// Create a fully-typed API client for users
const usersAPI = createTypedAPI<User>('/users');

// === SERVICE CLASS ===

export class ModernUserService {
	/**
	 * Get all users with comprehensive error handling
	 */
	static async getAllUsers(): Promise<User[] | null> {
		const response = await usersAPI.list();

		if (response.success) {
			return response.data || [];
		}

		// Handle specific error types
		if (!handleAuthError(response.error) && !handleNetworkError(response.error)) {
			NotificationService.error(response.error);
		}

		return null;
	}

	/**
	 * Get user by ID with retry capability
	 */
	static async getUserById(id: number): Promise<User | null> {
		const retryFn = () => this.getUserById(id);
		const response = await usersAPI.get(id);

		if (response.success) {
			return response.data || null;
		}

		// Handle network errors with retry option
		if (handleNetworkError(response.error, retryFn)) {
			return null;
		}

		// Handle other errors
		if (!handleAuthError(response.error)) {
			NotificationService.error(response.error);
		}

		return null;
	}

	/**
	 * Create user with validation and success feedback
	 */
	static async createUser(userData: CreateUserRequest): Promise<User | null> {
		const response = await usersAPI.create(userData);

		if (response.success) {
			NotificationService.success(`User ${userData.name} created successfully!`);
			return response.data || null;
		}

		// Handle specific error cases
		if (!handleAuthError(response.error)) {
			NotificationService.error(response.error);
		}

		return null;
	}

	/**
	 * Update user with optimistic UI patterns
	 */
	static async updateUser(userData: UpdateUserRequest): Promise<User | null> {
		const response = await usersAPI.update(userData.id, userData);

		if (response.success) {
			NotificationService.success('User updated successfully!');
			return response.data || null;
		}

		// Handle specific error cases
		if (!handleAuthError(response.error)) {
			NotificationService.error(response.error);
		}

		return null;
	}

	/**
	 * Delete user with confirmation
	 */
	static async deleteUser(id: number, userName: string): Promise<boolean> {
		const response = await usersAPI.delete(id);

		if (response.success) {
			NotificationService.success(`User ${userName} deleted successfully!`);
			return true;
		}

		// Handle specific error cases
		if (!handleAuthError(response.error)) {
			NotificationService.error(response.error);
		}

		return false;
	}

	/**
	 * Bulk operations with progress feedback
	 */
	static async bulkUpdateUsers(updates: UpdateUserRequest[]): Promise<User[]> {
		const results: User[] = [];
		let successCount = 0;
		let errorCount = 0;

		for (const update of updates) {
			const result = await this.updateUser(update);
			if (result) {
				results.push(result);
				successCount++;
			} else {
				errorCount++;
			}
		}

		// Show summary notification
		if (successCount > 0 && errorCount === 0) {
			NotificationService.success(`Successfully updated ${successCount} users!`);
		} else if (successCount > 0 && errorCount > 0) {
			NotificationService.warning(`Updated ${successCount} users, ${errorCount} failed`);
		} else {
			NotificationService.error('Failed to update users');
		}

		return results;
	}

	/**
	 * Search users with debouncing and caching
	 */
	static async searchUsers(query: string, filters?: { role?: string }): Promise<User[]> {
		const params = new URLSearchParams();
		if (query) params.set('q', query);
		if (filters?.role) params.set('role', filters.role);

		const endpoint = `/users/search?${params.toString()}`;
		const response = await apiClient.get<User[]>(endpoint);

		if (response.success) {
			return response.data || [];
		}

		// Don't show errors for search - just return empty results
		console.warn('Search failed:', response.error);
		return [];
	}

	/**
	 * Export users data
	 */
	static async exportUsers(format: 'csv' | 'json' = 'csv'): Promise<boolean> {
		const response = await apiClient.get(`/users/export?format=${format}`, {
			timeout: 30000 // Longer timeout for exports
		});

		if (response.success) {
			// Handle file download
			const blob = new Blob([response.data as string], {
				type: format === 'csv' ? 'text/csv' : 'application/json'
			});
			const url = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = `users-export.${format}`;
			a.click();
			URL.revokeObjectURL(url);

			NotificationService.success('Users exported successfully!');
			return true;
		}

		NotificationService.error(response.error);
		return false;
	}
}

// === EXAMPLE USAGE PATTERNS ===

/*
// In a Svelte component:

<script lang="ts">
import { ModernUserService } from '$lib/services/modernApiExample';
import ErrorBoundary from '$lib/components/ErrorBoundary.svelte';

let users = $state<User[]>([]);
let isLoading = $state(false);

async function loadUsers() {
  isLoading = true;
  const result = await ModernUserService.getAllUsers();
  if (result) {
    users = result;
  }
  isLoading = false;
}

async function handleCreateUser(userData: CreateUserRequest) {
  const newUser = await ModernUserService.createUser(userData);
  if (newUser) {
    users = [...users, newUser];
  }
}

async function handleDeleteUser(id: number, name: string) {
  const success = await ModernUserService.deleteUser(id, name);
  if (success) {
    users = users.filter(u => u.id !== id);
  }
}
</script>

<ErrorBoundary fallbackMessage="User management temporarily unavailable">
  {#if isLoading}
    <LoadingState />
  {:else}
    <UserList 
      {users} 
      onCreateUser={handleCreateUser}
      onDeleteUser={handleDeleteUser}
    />
  {/if}
</ErrorBoundary>
*/
