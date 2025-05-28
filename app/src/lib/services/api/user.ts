import { authenticatedFetch } from '$lib/utils/api';

export interface AvailableShiftSlot {
	schedule_id: number;
	schedule_name: string;
	start_time: string;
	end_time: string;
	timezone?: string;
	is_booked: boolean;
}

export interface UserBooking {
	booking_id: number;
	user_id: number;
	schedule_id: number;
	schedule_name: string;
	shift_start: string;
	shift_end: string;
	buddy_user_id?: number;
	buddy_name?: string;
	checked_in_at?: string;
	created_at: string;
}

export interface CreateBookingRequest {
	schedule_id: number;
	start_time: string;
	buddy_name?: string;
	buddy_phone?: string;
}

export interface CreateReportRequest {
	severity: number; // 0=low, 1=normal, 2=high
	message: string;
}

export interface ReportResponse {
	report_id: number;
	booking_id: number;
	severity: number;
	message: string;
	created_at: string;
}

export class UserApiService {
	/**
	 * Get available shift slots
	 */
	static async getAvailableShifts(params?: {
		from?: string;
		to?: string;
		limit?: number;
	}): Promise<AvailableShiftSlot[]> {
		const searchParams = new URLSearchParams();
		if (params?.from) searchParams.append('from', params.from);
		if (params?.to) searchParams.append('to', params.to);
		if (params?.limit) searchParams.append('limit', params.limit.toString());

		const url = `/shifts/available${searchParams.toString() ? `?${searchParams.toString()}` : ''}`;
		const response = await fetch(url);

		if (!response.ok) {
			throw new Error(`Failed to fetch available shifts: ${response.status}`);
		}
		return response.json();
	}

	/**
	 * Create a new booking
	 */
	static async createBooking(request: CreateBookingRequest): Promise<UserBooking> {
		const response = await authenticatedFetch('/bookings', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify(request)
		});

		if (!response.ok) {
			const errorText = await response.text();
			throw new Error(`Failed to create booking: ${errorText}`);
		}
		return response.json();
	}

	/**
	 * Mark check-in for a booking
	 */
	static async markCheckIn(bookingId: number): Promise<UserBooking> {
		const response = await authenticatedFetch(`/bookings/${bookingId}/checkin`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			}
		});

		if (!response.ok) {
			const errorText = await response.text();
			throw new Error(`Failed to check in: ${errorText}`);
		}
		return response.json();
	}

	/**
	 * Submit an incident report for a booking
	 */
	static async submitReport(
		bookingId: number,
		request: CreateReportRequest
	): Promise<ReportResponse> {
		const response = await authenticatedFetch(`/bookings/${bookingId}/report`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify(request)
		});

		if (!response.ok) {
			const errorText = await response.text();
			throw new Error(`Failed to submit report: ${errorText}`);
		}
		return response.json();
	}

	/**
	 * Get current user's bookings
	 */
	static async getMyBookings(): Promise<UserBooking[]> {
		const response = await authenticatedFetch('/bookings/my');

		if (!response.ok) {
			throw new Error(`Failed to fetch user bookings: ${response.status}`);
		}
		return response.json();
	}
}
