import { createQuery } from '@tanstack/svelte-query';
import { getCurrentUser } from '$lib/utils/auth';

export interface LeaderboardEntry {
	user_id: number;
	name: string;
	phone?: string;
	total_points: number;
	shift_count: number;
	rank: number;
}

export interface LeaderboardData {
	pointsLeaderboard: LeaderboardEntry[];
	shiftsLeaderboard: LeaderboardEntry[];
}

async function fetchLeaderboardData(): Promise<LeaderboardData> {
	const user = getCurrentUser();
	if (!user?.token) {
		throw new Error('Not authenticated');
	}

	const headers = {
		Authorization: `Bearer ${user.token}`,
		'Content-Type': 'application/json'
	};

	try {
		// Fetch both leaderboards in parallel
		const [pointsRes, shiftsRes] = await Promise.all([
			fetch('/api/leaderboard', { headers }),
			fetch('/api/leaderboard/shifts', { headers })
		]);

		// Handle individual endpoint failures gracefully
		let pointsLeaderboard: LeaderboardEntry[] = [];
		let shiftsLeaderboard: LeaderboardEntry[] = [];

		if (pointsRes.ok) {
			pointsLeaderboard = await pointsRes.json();
		} else {
			console.warn('Failed to fetch points leaderboard:', pointsRes.status, pointsRes.statusText);
		}

		if (shiftsRes.ok) {
			shiftsLeaderboard = await shiftsRes.json();
		} else {
			console.warn('Failed to fetch shifts leaderboard:', shiftsRes.status, shiftsRes.statusText);
		}

		return {
			pointsLeaderboard: pointsLeaderboard || [],
			shiftsLeaderboard: shiftsLeaderboard || []
		};
	} catch (error) {
		console.error('Error fetching leaderboard data:', error);
		// Return empty data instead of throwing to allow dashboard to still load
		return {
			pointsLeaderboard: [],
			shiftsLeaderboard: []
		};
	}
}

export function createLeaderboardQuery() {
	return createQuery({
		queryKey: ['leaderboard'],
		queryFn: fetchLeaderboardData,
		refetchInterval: 2 * 60 * 1000, // Refetch every 2 minutes (more frequent for points consistency)
		staleTime: 30 * 1000, // Consider data stale after 30 seconds (faster updates)
		retry: 3, // Retry failed requests 3 times (improved reliability)
		retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 5000) // Exponential backoff with max 5s
	});
}
