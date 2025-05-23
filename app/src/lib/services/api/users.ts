import { authenticatedFetch } from '$lib/utils/api';
import type { UserData } from '$lib/schemas/user';

export class UsersApiService {
	/**
	 * Get all users (admin only)
	 */
	static async getAll(): Promise<UserData[]> {
		const response = await authenticatedFetch('/api/admin/users');
		if (!response.ok) {
			throw new Error('Failed to fetch users');
		}
		return response.json();
	}

	/**
	 * Get a specific user by ID
	 */
	static async getById(userId: number): Promise<UserData> {
		const response = await authenticatedFetch(`/api/admin/users/${userId}`);
		if (!response.ok) {
			throw new Error(`Failed to fetch user ${userId}: ${response.status}`);
		}
		return response.json();
	}

	/**
	 * Create a new user
	 */
	static async create(payload: {
		name: string;
		phone: string;
		role: string;
		emergency_contact_name?: string;
		emergency_contact_phone?: string;
		notes?: string;
	}): Promise<{ message: string; user_id?: number }> {
		const response = await authenticatedFetch('/api/admin/users', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});
		if (!response.ok) {
			const errorData = await response
				.json()
				.catch(() => ({ message: 'Failed to create user' }));
			throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
		}
		return response.json();
	}

	/**
	 * Update an existing user
	 */
	static async update(
		userId: number,
		payload: {
			name: string;
			phone: string;
			role: string;
			emergency_contact_name?: string;
			emergency_contact_phone?: string;
			notes?: string;
		}
	): Promise<{ message: string; user_id?: number }> {
		const response = await authenticatedFetch(`/api/admin/users/${userId}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});
		if (!response.ok) {
			const errorData = await response
				.json()
				.catch(() => ({ message: 'Failed to update user' }));
			throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
		}
		return response.json();
	}

	/**
	 * Delete a user
	 */
	static async delete(userId: number): Promise<{ ok: boolean; message?: string }> {
		const response = await authenticatedFetch(`/api/admin/users/${userId}`, {
			method: 'DELETE'
		});
		if (!response.ok) {
			const errorData = await response
				.json()
				.catch(() => ({ message: 'Failed to delete user' }));
			throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
		}
		// Handle both JSON response and empty response
		try {
			return await response.json();
		} catch (e) {
			return { ok: response.ok };
		}
	}

	/**
	 * Update user role
	 */
	static async updateRole(
		userId: number,
		role: string
	): Promise<{ message: string }> {
		const response = await authenticatedFetch(`/api/admin/users/${userId}/role`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ role })
		});
		if (!response.ok) {
			const errorData = await response
				.json()
				.catch(() => ({ message: 'Failed to update user role' }));
			throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
		}
		return response.json();
	}

	/**
	 * Bulk delete users
	 */
	static async bulkDelete(userIds: number[]): Promise<{ message: string; deleted_count: number }> {
		const response = await authenticatedFetch('/api/admin/users/bulk-delete', {
			method: 'DELETE',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ user_ids: userIds })
		});
		if (!response.ok) {
			const errorData = await response
				.json()
				.catch(() => ({ message: 'Failed to bulk delete users' }));
			throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
		}
		return response.json();
	}
} 