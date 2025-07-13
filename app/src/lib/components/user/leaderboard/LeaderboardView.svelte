<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Trophy, Medal, Award } from 'lucide-svelte';
	import { authenticatedFetch } from '$lib/utils/api';

	interface LeaderboardEntry {
		user_id: number;
		name: string;
		phone?: string;
		total_points: number;
		shift_count: number;
		rank: number;
	}

	interface UserStats {
		total_points: number;
		shift_count: number;
		rank: number;
		recent_achievements: Achievement[];
	}

	interface Achievement {
		achievement_id: number;
		name: string;
		description: string;
		icon: string;
		earned_at: string;
	}

	let activeTab: 'points' | 'shifts' = 'points';
	let pointsLeaderboard: LeaderboardEntry[] = [];
	let shiftsLeaderboard: LeaderboardEntry[] = [];
	let userStats: UserStats | null = null;
	let loading = true;
	let error = '';

	async function fetchLeaderboards() {
		try {
			const [pointsRes, shiftsRes, statsRes] = await Promise.all([
				authenticatedFetch('/api/leaderboard'),
				authenticatedFetch('/api/leaderboard/shifts'),
				authenticatedFetch('/api/user/stats')
			]);

			pointsLeaderboard = await pointsRes.json();
			shiftsLeaderboard = await shiftsRes.json();
			userStats = await statsRes.json();
		} catch (err) {
			error = err instanceof Error ? err.message : 'An error occurred';
		} finally {
			loading = false;
		}
	}

	function getRankIcon(rank: number) {
		switch (rank) {
			case 1:
				return { icon: Trophy, class: 'text-yellow-500' };
			case 2:
				return { icon: Medal, class: 'text-gray-400' };
			case 3:
				return { icon: Award, class: 'text-amber-600' };
			default:
				return null;
		}
	}

	function getInitials(name: string): string {
		return name
			.split(' ')
			.map((n) => n[0])
			.join('')
			.toUpperCase()
			.slice(0, 2);
	}

	onMount(() => {
		fetchLeaderboards();
	});
</script>

<div class="space-y-4">
	<!-- Your Stats - Compact -->
	{#if userStats}
		<div class="bg-muted/50 rounded-lg p-3">
			<div class="grid grid-cols-3 gap-4 text-center">
				<div>
					<div class="text-lg font-bold">{userStats.total_points}</div>
					<div class="text-xs text-muted-foreground">Points</div>
				</div>
				<div>
					<div class="text-lg font-bold">{userStats.shift_count}</div>
					<div class="text-xs text-muted-foreground">Shifts</div>
				</div>
				<div>
					<div class="text-lg font-bold">#{userStats.rank}</div>
					<div class="text-xs text-muted-foreground">Rank</div>
				</div>
			</div>
		</div>
	{/if}

	<!-- Tab Toggle - Minimal -->
	<div class="flex rounded-lg bg-muted p-1">
		<Button
			variant={activeTab === 'points' ? 'default' : 'ghost'}
			size="sm"
			class="flex-1 h-8"
			onclick={() => (activeTab = 'points')}
		>
			Points
		</Button>
		<Button
			variant={activeTab === 'shifts' ? 'default' : 'ghost'}
			size="sm"
			class="flex-1 h-8"
			onclick={() => (activeTab = 'shifts')}
		>
			Shifts
		</Button>
	</div>

	<!-- Leaderboard List - Compact -->
	{#if loading}
		<div class="flex justify-center py-8">
			<div class="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
		</div>
	{:else if error}
		<div class="text-center py-4 text-sm text-destructive">
			{error}
		</div>
	{:else}
		<div class="space-y-1">
			{#each activeTab === 'points' ? pointsLeaderboard : shiftsLeaderboard as entry (entry.user_id)}
				{@const rankInfo = getRankIcon(entry.rank)}
				<div class="flex items-center gap-3 p-2 rounded-lg bg-background border">
					<!-- Rank -->
					<div class="w-8 flex justify-center">
						{#if rankInfo}
							<svelte:component this={rankInfo.icon} class="h-4 w-4 {rankInfo.class}" />
						{:else}
							<span class="text-sm font-medium text-muted-foreground">#{entry.rank}</span>
						{/if}
					</div>

					<!-- Avatar -->
					<div
						class="w-8 h-8 bg-muted rounded-full flex items-center justify-center text-xs font-medium"
					>
						{getInitials(entry.name)}
					</div>

					<!-- Name & Stats -->
					<div class="flex-1 min-w-0">
						<div class="font-medium text-sm truncate">{entry.name}</div>
						<div class="text-xs text-muted-foreground">
							{activeTab === 'points'
								? `${entry.shift_count} shifts`
								: `${entry.total_points} points`}
						</div>
					</div>

					<!-- Main Stat -->
					<div class="text-right">
						<div class="font-bold text-sm">
							{activeTab === 'points' ? entry.total_points : entry.shift_count}
						</div>
					</div>
				</div>
			{/each}
		</div>

		<!-- Empty State -->
		{#if (activeTab === 'points' ? pointsLeaderboard : shiftsLeaderboard).length === 0}
			<div class="text-center py-8 text-sm text-muted-foreground">No data available</div>
		{/if}
	{/if}

	<!-- Recent Achievements - Only if present -->
	{#if userStats && userStats.recent_achievements && userStats.recent_achievements.length > 0}
		<div class="space-y-2">
			<h3 class="text-sm font-medium">Recent Achievements</h3>
			{#each userStats.recent_achievements.slice(0, 3) as achievement (achievement.achievement_id)}
				<div class="flex items-center gap-2 p-2 bg-muted/50 rounded">
					<span class="text-lg">{achievement.icon}</span>
					<div class="flex-1 min-w-0">
						<div class="text-sm font-medium truncate">{achievement.name}</div>
						<div class="text-xs text-muted-foreground truncate">{achievement.description}</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
