import { authenticatedFetch } from '$lib/utils/api';
import type { Schedule, AdminShiftSlot } from '$lib/types';

export class SchedulesApiService {
	/**
	 * Get all schedules
	 */
	static async getAll(): Promise<Schedule[]> {
		const response = await authenticatedFetch('/api/admin/schedules');
		if (!response.ok) {
			throw new Error('Failed to fetch schedules');
		}
		return response.json();
	}

	/**
	 * Get a specific schedule by ID
	 */
	static async getById(scheduleId: number): Promise<Schedule> {
		const response = await authenticatedFetch(`/api/admin/schedules/${scheduleId}`);
		if (!response.ok) {
			throw new Error(`Failed to fetch schedule ${scheduleId}: ${response.status}`);
		}
		return response.json();
	}

	/**
	 * Create a new schedule
	 */
	static async create(payload: {
		name: string;
		cron_expr: string;
		start_date: string | null;
		end_date: string | null;
	}): Promise<{ message: string; schedule_id?: number }> {
		const response = await authenticatedFetch('/api/admin/schedules', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});
		if (!response.ok) {
			const errorData = await response
				.json()
				.catch(() => ({ message: 'Failed to create schedule' }));
			throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
		}
		return response.json();
	}

	/**
	 * Update an existing schedule
	 */
	static async update(
		scheduleId: number,
		payload: {
			name: string;
			cron_expr: string;
			start_date: string | null;
			end_date: string | null;
		}
	): Promise<{ message: string; schedule_id?: number }> {
		const response = await authenticatedFetch(`/api/admin/schedules/${scheduleId}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});
		if (!response.ok) {
			const errorData = await response
				.json()
				.catch(() => ({ message: 'Failed to update schedule' }));
			throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
		}
		return response.json();
	}

	/**
	 * Delete a schedule
	 */
	static async delete(scheduleId: number): Promise<{ ok: boolean; message?: string }> {
		const response = await authenticatedFetch(`/api/admin/schedules/${scheduleId}`, {
			method: 'DELETE'
		});
		if (!response.ok) {
			const errorData = await response
				.json()
				.catch(() => ({ message: 'Failed to delete schedule' }));
			throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
		}
		// Handle both JSON response and empty response
		try {
			return await response.json();
		} catch (_e) {
			return { ok: response.ok };
		}
	}

	/**
	 * Get all shift slots across schedules with optional date filtering
	 */
	static async getAllSlots(params?: { from?: string; to?: string }): Promise<AdminShiftSlot[]> {
		const searchParams = new URLSearchParams();
		if (params?.from) searchParams.append('from', params.from);
		if (params?.to) searchParams.append('to', params.to);

		const url = `/api/admin/schedules/all-slots${searchParams.toString() ? `?${searchParams.toString()}` : ''}`;
		const response = await authenticatedFetch(url);

		if (!response.ok) {
			let errorMsg = `HTTP error ${response.status}`;
			try {
				const errorData = await response.json();
				errorMsg = errorData.message || errorData.error || errorMsg;
			} catch {
				/* ignore */
			}
			throw new Error(errorMsg);
		}
		return response.json();
	}
}
