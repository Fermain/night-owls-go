import type { AdminDashboardData, DashboardMetrics } from '$lib/queries/admin/dashboard';
import type { UserData } from '$lib/schemas/user';

export interface CriticalIssue {
	type: 'unfilled_shifts' | 'high_no_show' | 'guest_backlog';
	severity: 'high' | 'medium' | 'low';
	message: string;
	action: string;
}

export interface Champion {
	user_id: number;
	name: string;
	shifts_booked: number;
	attendance_rate: number;
	completion_rate: number;
}

export interface FreeLoader {
	user_id: number;
	name: string;
	shifts_booked: number;
	attendance_rate: number;
}

export interface TopReporter {
	user_id: number;
	name: string;
	shifts_completed: number;
}

export interface IntelligentInsights {
	pendingGuests: UserData[];
	freeLoaders: FreeLoader[];
	champions: Champion[];
	topReporters: TopReporter[];
	criticalIssues: CriticalIssue[];
	metrics: DashboardMetrics;
}

/**
 * Calculate intelligent insights from dashboard and user data
 */
export function calculateIntelligentInsights(
	dashboardData: AdminDashboardData,
	usersData: UserData[]
): IntelligentInsights {
	const { member_contributions: contributions, quality_metrics, metrics } = dashboardData;

	// 1. Guest Approval Queue
	const pendingGuests = usersData.filter((u: UserData) => u.role === 'guest');

	// 2. Free Loaders (book but don't show up)
	const freeLoaders: FreeLoader[] = contributions
		.filter((c) => c.shifts_booked > 0 && c.attendance_rate < 60)
		.slice(0, 3)
		.map((c) => ({
			user_id: c.user_id,
			name: c.name,
			shifts_booked: c.shifts_booked,
			attendance_rate: c.attendance_rate
		}));

	// 3. Champions (reliable and active)
	const champions: Champion[] = contributions
		.filter((c) => c.shifts_booked >= 2 && c.attendance_rate >= 90 && c.completion_rate >= 80)
		.slice(0, 3)
		.map((c) => ({
			user_id: c.user_id,
			name: c.name,
			shifts_booked: c.shifts_booked,
			attendance_rate: c.attendance_rate,
			completion_rate: c.completion_rate
		}));

	// 4. Top Reporters
	const topReporters: TopReporter[] = contributions
		.filter((c) => c.shifts_completed >= 3)
		.slice(0, 3)
		.map((c) => ({
			user_id: c.user_id,
			name: c.name,
			shifts_completed: c.shifts_completed
		}));

	// 5. Critical Issues
	const criticalIssues: CriticalIssue[] = [];

	if (metrics.next_week_unfilled > 5) {
		criticalIssues.push({
			type: 'unfilled_shifts',
			severity: 'high',
			message: `${metrics.next_week_unfilled} shifts unfilled next week`,
			action: 'Schedule more volunteers'
		});
	}

	if (quality_metrics.no_show_rate > 25) {
		criticalIssues.push({
			type: 'high_no_show',
			severity: 'medium',
			message: `${quality_metrics.no_show_rate.toFixed(1)}% no-show rate`,
			action: 'Contact unreliable volunteers'
		});
	}

	if (pendingGuests.length > 10) {
		criticalIssues.push({
			type: 'guest_backlog',
			severity: 'low',
			message: `${pendingGuests.length} guests awaiting approval`,
			action: 'Review guest applications'
		});
	}

	return {
		pendingGuests,
		freeLoaders,
		champions,
		topReporters,
		criticalIssues,
		metrics
	};
}

/**
 * Get status counts for overview cards
 */
export function getStatusCounts(insights: IntelligentInsights) {
	return {
		criticalIssues: insights.criticalIssues.length,
		pendingGuests: insights.pendingGuests.length,
		champions: insights.champions.length,
		freeLoaders: insights.freeLoaders.length
	};
}

/**
 * Determine if there are critical issues requiring immediate attention
 */
export function hasCriticalIssues(insights: IntelligentInsights): boolean {
	return insights.criticalIssues.some((issue) => issue.severity === 'high');
}

/**
 * Get the most urgent issue for display
 */
export function getMostUrgentIssue(insights: IntelligentInsights): CriticalIssue | null {
	const highPriority = insights.criticalIssues.filter((issue) => issue.severity === 'high');
	if (highPriority.length > 0) return highPriority[0];

	const mediumPriority = insights.criticalIssues.filter((issue) => issue.severity === 'medium');
	if (mediumPriority.length > 0) return mediumPriority[0];

	const lowPriority = insights.criticalIssues.filter((issue) => issue.severity === 'low');
	if (lowPriority.length > 0) return lowPriority[0];

	return null;
}
