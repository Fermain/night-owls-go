import { authenticatedFetch } from '$lib/utils/api';
import type { components } from '$lib/types/api';

type BookingResponse = components['schemas']['api.BookingResponse'];

export class ShiftsApiService {
	/**
	 * Assign a user to a shift slot
	 */
	static async assignUser(payload: {
		schedule_id: number;
		start_time: string;
		user_id: number;
	}): Promise<{ success: boolean; message?: string }> {
		const response = await authenticatedFetch('/api/admin/bookings/assign', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});
		if (!response.ok) {
			const errorData = await response.json().catch(() => ({ message: 'Failed to assign shift' }));
			throw new Error(errorData.message || `HTTP error ${response.status}`);
		}
		return response.json();
	}

	/**
	 * Unassign a user from a shift slot
	 */
	static async unassignUser(payload: {
		schedule_id: number;
		start_time: string;
	}): Promise<{ success: boolean; message?: string }> {
		const response = await authenticatedFetch('/api/admin/bookings/unassign', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});
		if (!response.ok) {
			const errorData = await response
				.json()
				.catch(() => ({ message: 'Failed to unassign shift' }));
			throw new Error(errorData.message || `HTTP error ${response.status}`);
		}
		return response.json();
	}

	/**
	 * Get shift bookings for a user
	 */
	static async getUserBookings(userId: number): Promise<BookingResponse[]> {
		const response = await authenticatedFetch(`/api/admin/users/${userId}/bookings`);
		if (!response.ok) {
			throw new Error(`Failed to fetch user bookings: ${response.status}`);
		}
		return response.json();
	}

	/**
	 * Get all bookings (admin view)
	 */
	static async getAllBookings(params?: {
		from?: string;
		to?: string;
		user_id?: number;
		schedule_id?: number;
	}): Promise<BookingResponse[]> {
		const searchParams = new URLSearchParams();
		if (params?.from) searchParams.append('from', params.from);
		if (params?.to) searchParams.append('to', params.to);
		if (params?.user_id) searchParams.append('user_id', params.user_id.toString());
		if (params?.schedule_id) searchParams.append('schedule_id', params.schedule_id.toString());

		const url = `/api/admin/bookings${searchParams.toString() ? `?${searchParams.toString()}` : ''}`;
		const response = await authenticatedFetch(url);

		if (!response.ok) {
			throw new Error(`Failed to fetch bookings: ${response.status}`);
		}
		return response.json();
	}

	/**
	 * Cancel a booking
	 */
	static async cancelBooking(bookingId: number): Promise<{ success: boolean; message?: string }> {
		const response = await authenticatedFetch(`/api/admin/bookings/${bookingId}/cancel`, {
			method: 'POST'
		});
		if (!response.ok) {
			const errorData = await response
				.json()
				.catch(() => ({ message: 'Failed to cancel booking' }));
			throw new Error(errorData.message || `HTTP error ${response.status}`);
		}
		return response.json();
	}
}
