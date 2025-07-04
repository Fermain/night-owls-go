import type { UserData } from '$lib/schemas/user';
import type { AdminShiftSlot } from '$lib/types';

export interface UserMetrics {
	totalUsers: number;
	adminUsers: number;
	owlUsers: number;
	guestUsers: number;
	recentUsers: number; // Users registered in last 30 days
	activeUsers: number; // Users with recent activity
	roleDistribution: { role: string; count: number; percentage: number }[];
}

export interface UserGrowthData {
	period: string;
	total: number;
	new: number;
}

export interface UserShiftMetrics {
	totalShifts: number;
	averageShiftsPerUser: number;
	usersWithShifts: number;
	usersWithoutShifts: number;
	shiftDistribution: UserShiftDistribution[];
	topVolunteers: UserShiftDistribution[];
	workloadBalance: 'balanced' | 'uneven' | 'concentrated';
}

export interface UserShiftDistribution {
	userId: number;
	userName: string;
	userRole: string;
	shiftCount: number;
	percentage: number;
}

/**
 * Calculate comprehensive user metrics from user data
 */
export function calculateUserMetrics(users: UserData[]): UserMetrics {
	const totalUsers = users.length;
	const adminUsers = users.filter((u) => u.role === 'admin').length;
	const owlUsers = users.filter((u) => u.role === 'owl').length;
	const guestUsers = users.filter((u) => u.role === 'guest').length;

	// Calculate recent users (registered in last 30 days)
	const thirtyDaysAgo = new Date();
	thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30);
	const recentUsers = users.filter((u) => {
		if (!u.created_at) return false;
		const createdDate = new Date(u.created_at);
		return createdDate >= thirtyDaysAgo;
	}).length;

	// For now, assume active users = users with recent activity (can be enhanced later)
	const activeUsers = users.filter((u) => u.role !== 'guest').length;

	// Role distribution with percentages
	const roleDistribution = [
		{
			role: 'Admin',
			count: adminUsers,
			percentage: totalUsers > 0 ? Math.round((adminUsers / totalUsers) * 100) : 0
		},
		{
			role: 'Owl',
			count: owlUsers,
			percentage: totalUsers > 0 ? Math.round((owlUsers / totalUsers) * 100) : 0
		},
		{
			role: 'Guest',
			count: guestUsers,
			percentage: totalUsers > 0 ? Math.round((guestUsers / totalUsers) * 100) : 0
		}
	];

	return {
		totalUsers,
		adminUsers,
		owlUsers,
		guestUsers,
		recentUsers,
		activeUsers,
		roleDistribution
	};
}

/**
 * Calculate shift distribution metrics per user
 */
export function calculateUserShiftMetrics(
	users: UserData[],
	shifts: AdminShiftSlot[]
): UserShiftMetrics {
	// Count shifts per user
	const userShiftCounts = new Map<string, { user: UserData; count: number }>();

	// Initialize all users with 0 shifts (using name as key)
	users.forEach((user) => {
		const userName = user.name || 'Unnamed User';
		userShiftCounts.set(userName, { user, count: 0 });
	});

	// Count actual shifts by matching user names
	const totalShifts = shifts.filter((shift) => shift.is_booked && shift.user_name).length;

	shifts.forEach((shift) => {
		if (shift.is_booked && shift.user_name) {
			const current = userShiftCounts.get(shift.user_name);
			if (current) {
				current.count += 1;
			}
		}
	});

	// Convert to distribution array
	const shiftDistribution: UserShiftDistribution[] = Array.from(userShiftCounts.values())
		.map(({ user, count }) => ({
			userId: user.id,
			userName: user.name || 'Unnamed User',
			userRole: user.role,
			shiftCount: count,
			percentage: totalShifts > 0 ? Math.round((count / totalShifts) * 100) : 0
		}))
		.sort((a, b) => b.shiftCount - a.shiftCount);

	// Calculate metrics
	const usersWithShifts = shiftDistribution.filter((u) => u.shiftCount > 0).length;
	const usersWithoutShifts = shiftDistribution.filter((u) => u.shiftCount === 0).length;
	const averageShiftsPerUser =
		users.length > 0 ? Math.round((totalShifts / users.length) * 10) / 10 : 0;

	// Top volunteers (top 5 by shift count)
	const topVolunteers = shiftDistribution.filter((u) => u.shiftCount > 0).slice(0, 5);

	// Determine workload balance
	let workloadBalance: 'balanced' | 'uneven' | 'concentrated' = 'balanced';
	if (topVolunteers.length > 0) {
		const topUserPercentage = topVolunteers[0]?.percentage || 0;
		if (topUserPercentage > 40) {
			workloadBalance = 'concentrated';
		} else if (topUserPercentage > 25) {
			workloadBalance = 'uneven';
		}
	}

	return {
		totalShifts,
		averageShiftsPerUser,
		usersWithShifts,
		usersWithoutShifts,
		shiftDistribution,
		topVolunteers,
		workloadBalance
	};
}

/**
 * Generate user growth data for charts (simulated for now)
 */
export function generateUserGrowthData(users: UserData[]): UserGrowthData[] {
	// Group users by month for the last 6 months
	const months = [];
	const now = new Date();

	for (let i = 5; i >= 0; i--) {
		const date = new Date(now.getFullYear(), now.getMonth() - i, 1);
		months.push({
			date,
			label: date.toLocaleDateString('en-US', { month: 'short', year: 'numeric' })
		});
	}

	return months.map((month, index) => {
		// Simulate growth data - in real app, you'd calculate from actual creation dates
		const baseUsers = Math.max(10, users.length - (5 - index) * 5);
		const newUsers = Math.floor(Math.random() * 10) + 1;

		return {
			period: month.label,
			total: baseUsers,
			new: newUsers
		};
	});
}

/**
 * Get recently registered users (last 7 days)
 */
export function getRecentUsers(users: UserData[], days: number = 7): UserData[] {
	const cutoffDate = new Date();
	cutoffDate.setDate(cutoffDate.getDate() - days);

	return users
		.filter((u) => {
			if (!u.created_at) return false;
			return new Date(u.created_at) >= cutoffDate;
		})
		.sort((a, b) => {
			const dateA = new Date(a.created_at || 0);
			const dateB = new Date(b.created_at || 0);
			return dateB.getTime() - dateA.getTime();
		})
		.slice(0, 10); // Limit to 10 most recent
}

/**
 * Search and filter users
 */
export function filterUsers(
	users: UserData[],
	searchTerm: string,
	roleFilter?: string
): UserData[] {
	return users.filter((user) => {
		// Role filter
		if (roleFilter && roleFilter !== 'all' && user.role !== roleFilter) {
			return false;
		}

		// Search filter
		if (searchTerm) {
			const term = searchTerm.toLowerCase();
			return (
				user.name?.toLowerCase().includes(term) ||
				user.phone?.toLowerCase().includes(term) ||
				user.role?.toLowerCase().includes(term)
			);
		}

		return true;
	});
}
