import { authenticatedFetch } from '$lib/utils/api';

export interface BookingResponse {
	booking_id: number;
	user_id: number;
	schedule_id: number;
	schedule_name: string;
	shift_start: string;
	shift_end: string;
	buddy_user_id?: number;
	buddy_name?: string;
	attended?: boolean;
	created_at: string;
}

export class BookingsApiService {
	/**
	 * Get all bookings for a specific user (admin only)
	 */
	static async getUserBookings(userId: number): Promise<BookingResponse[]> {
		const response = await authenticatedFetch(`/api/admin/users/${userId}/bookings`);

		if (!response.ok) {
			const error = await response.text();
			throw new Error(`Failed to fetch user bookings: ${error}`);
		}

		return response.json();
	}

	/**
	 * Assign user to shift (admin only)
	 */
	static async assignUserToShift(params: {
		user_id: number;
		schedule_id: number;
		start_time: string;
	}): Promise<BookingResponse> {
		const response = await authenticatedFetch('/api/admin/bookings/assign', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify(params)
		});

		if (!response.ok) {
			const error = await response.text();
			throw new Error(`Failed to assign user to shift: ${error}`);
		}

		return response.json();
	}
}
