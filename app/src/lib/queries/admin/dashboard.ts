import { createQuery } from '@tanstack/svelte-query';
import { getCurrentUser } from '$lib/utils/auth';

export interface DashboardMetrics {
	total_shifts: number;
	booked_shifts: number;
	unfilled_shifts: number;
	checked_in_shifts: number;
	completed_shifts: number;
	fill_rate: number;
	check_in_rate: number;
	completion_rate: number;
	next_week_unfilled: number;
	this_weekend_status: string;
}

export interface MemberContribution {
	user_id: number;
	name: string;
	phone: string;
	shifts_booked: number;
	shifts_attended: number;
	shifts_completed: number;
	attendance_rate: number;
	completion_rate: number;
	last_shift_date: string | null;
	contribution_category: string;
}

export interface QualityMetrics {
	no_show_rate: number;
	incomplete_rate: number;
	reliability_score: number;
}

export interface TimeSlotPattern {
	day_of_week: string;
	hour_of_day: string;
	total_bookings: number;
	check_in_rate: number;
	completion_rate: number;
}

export interface AdminDashboardData {
	metrics: DashboardMetrics;
	member_contributions: MemberContribution[];
	quality_metrics: QualityMetrics;
	problematic_slots: TimeSlotPattern[];
	generated_at: string;
}

async function fetchAdminDashboard(): Promise<AdminDashboardData> {
	const user = getCurrentUser();
	if (!user?.token) {
		throw new Error('Not authenticated');
	}

	const response = await fetch('/api/admin/dashboard', {
		method: 'GET',
		headers: {
			'Content-Type': 'application/json',
			Authorization: `Bearer ${user.token}`
		}
	});

	if (!response.ok) {
		const error = await response.json().catch(() => ({}));
		throw new Error(error.error || `Failed to fetch dashboard: ${response.statusText}`);
	}

	return response.json();
}

export function createAdminDashboardQuery() {
	return createQuery({
		queryKey: ['admin', 'dashboard'],
		queryFn: fetchAdminDashboard,
		refetchInterval: 60000, // Refetch every minute
		staleTime: 30000 // Consider data stale after 30 seconds
	});
} 