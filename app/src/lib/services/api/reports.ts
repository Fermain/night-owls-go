import { authenticatedFetch } from '$lib/utils/api';
import type { components } from '$lib/types/api';

type ReportResponse = components['schemas']['api.ReportResponse'];
type AdminReportResponse = components['schemas']['api.AdminReportResponse'];

// Define report severity levels
export type ReportSeverity = 0 | 1 | 2; // 0=Normal, 1=Suspicion, 2=Incident

export class ReportsApiService {
	/**
	 * Get all reports (admin only)
	 */
	static async getAll(params?: {
		from?: string;
		to?: string;
		severity?: number;
		user_id?: number;
		booking_id?: number;
	}): Promise<AdminReportResponse[]> {
		const searchParams = new URLSearchParams();
		if (params?.from) searchParams.append('from', params.from);
		if (params?.to) searchParams.append('to', params.to);
		if (params?.severity !== undefined) searchParams.append('severity', params.severity.toString());
		if (params?.user_id) searchParams.append('user_id', params.user_id.toString());
		if (params?.booking_id) searchParams.append('booking_id', params.booking_id.toString());

		const url = `/api/admin/reports${searchParams.toString() ? `?${searchParams.toString()}` : ''}`;
		const response = await authenticatedFetch(url);

		if (!response.ok) {
			throw new Error(`Failed to fetch reports: ${response.status}`);
		}
		return response.json();
	}

	/**
	 * Get a specific report by ID
	 */
	static async getById(reportId: number): Promise<AdminReportResponse> {
		const response = await authenticatedFetch(`/api/admin/reports/${reportId}`);
		if (!response.ok) {
			throw new Error(`Failed to fetch report ${reportId}: ${response.status}`);
		}
		return response.json();
	}

	/**
	 * Create a new report (typically from volunteers during shifts)
	 */
	static async create(
		bookingId: number,
		payload: {
			message: string;
			severity: number;
		}
	): Promise<ReportResponse> {
		const response = await authenticatedFetch(`/api/bookings/${bookingId}/report`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});
		if (!response.ok) {
			const errorData = await response.json().catch(() => ({ message: 'Failed to create report' }));
			throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
		}
		return response.json();
	}

	/**
	 * Create an off-shift report (when not on a scheduled shift)
	 */
	static async createOffShift(payload: {
		message: string;
		severity: number;
	}): Promise<ReportResponse> {
		const response = await authenticatedFetch('/api/reports/off-shift', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});
		if (!response.ok) {
			const errorData = await response
				.json()
				.catch(() => ({ message: 'Failed to create off-shift report' }));
			throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
		}
		return response.json();
	}

	/**
	 * Update an existing report (admin only)
	 */
	static async update(
		reportId: number,
		payload: {
			message: string;
			severity: number;
		}
	): Promise<AdminReportResponse> {
		const response = await authenticatedFetch(`/api/admin/reports/${reportId}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});
		if (!response.ok) {
			const errorData = await response.json().catch(() => ({ message: 'Failed to update report' }));
			throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
		}
		return response.json();
	}

	/**
	 * Delete a report (admin only)
	 */
	static async delete(reportId: number): Promise<{ ok: boolean; message?: string }> {
		const response = await authenticatedFetch(`/api/admin/reports/${reportId}`, {
			method: 'DELETE'
		});
		if (!response.ok) {
			const errorData = await response.json().catch(() => ({ message: 'Failed to delete report' }));
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
	 * Archive a report (admin only) - soft delete
	 */
	static async archive(reportId: number): Promise<{ ok: boolean; message?: string }> {
		const response = await authenticatedFetch(`/api/admin/reports/${reportId}/archive`, {
			method: 'PUT'
		});
		if (!response.ok) {
			const errorData = await response
				.json()
				.catch(() => ({ message: 'Failed to archive report' }));
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
	 * Unarchive a report (admin only) - restore from archive
	 */
	static async unarchive(reportId: number): Promise<{ ok: boolean; message?: string }> {
		const response = await authenticatedFetch(`/api/admin/reports/${reportId}/unarchive`, {
			method: 'PUT'
		});
		if (!response.ok) {
			const errorData = await response
				.json()
				.catch(() => ({ message: 'Failed to unarchive report' }));
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
	 * Get reports for a specific booking
	 */
	static async getByBookingId(bookingId: number): Promise<AdminReportResponse[]> {
		const response = await authenticatedFetch(`/api/admin/bookings/${bookingId}/reports`);
		if (!response.ok) {
			throw new Error(`Failed to fetch reports for booking ${bookingId}: ${response.status}`);
		}
		return response.json();
	}

	/**
	 * Get reports by a specific user
	 */
	static async getByUserId(userId: number): Promise<AdminReportResponse[]> {
		const response = await authenticatedFetch(`/api/admin/users/${userId}/reports`);
		if (!response.ok) {
			throw new Error(`Failed to fetch reports for user ${userId}: ${response.status}`);
		}
		return response.json();
	}
}
