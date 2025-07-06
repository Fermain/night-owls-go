<script lang="ts">
	import { onMount } from 'svelte';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Avatar, AvatarFallback } from '$lib/components/ui/avatar';
	import { Trophy, Medal, Award, Target, TrendingUp, Users } from 'lucide-svelte';
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
			// Fetch both leaderboards in parallel using authenticatedFetch
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
				return { icon: Trophy, class: 'text-yellow-600' };
			case 2:
				return { icon: Medal, class: 'text-gray-400' };
			case 3:
				return { icon: Award, class: 'text-amber-600' };
			default:
				return { icon: Target, class: 'text-muted-foreground' };
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

	function formatDate(dateStr: string): string {
		return new Date(dateStr).toLocaleDateString();
	}

	onMount(() => {
		fetchLeaderboards();
	});
</script>

<div class="space-y-6">
	<!-- Header with User Stats -->
	{#if userStats}
		<Card>
			<CardHeader>
				<CardTitle class="flex items-center gap-2">
					<TrendingUp class="h-5 w-5" />
					Your Performance
				</CardTitle>
			</CardHeader>
			<CardContent>
				<div class="grid grid-cols-3 gap-4 text-center">
					<div>
						<div class="text-2xl font-bold text-primary">{userStats.total_points}</div>
						<div class="text-sm text-muted-foreground">Total Points</div>
					</div>
					<div>
						<div class="text-2xl font-bold text-primary">{userStats.shift_count}</div>
						<div class="text-sm text-muted-foreground">Shifts Completed</div>
					</div>
					<div>
						<div class="text-2xl font-bold text-primary">#{userStats.rank}</div>
						<div class="text-sm text-muted-foreground">Current Rank</div>
					</div>
				</div>
			</CardContent>
		</Card>
	{/if}

	<!-- Tab Navigation -->
	<div class="flex space-x-1 rounded-lg bg-muted p-1">
		<Button
			variant={activeTab === 'points' ? 'default' : 'ghost'}
			size="sm"
			class="flex-1"
			onclick={() => (activeTab = 'points')}
		>
			<Trophy class="h-4 w-4 mr-2" />
			Points Leaderboard
		</Button>
		<Button
			variant={activeTab === 'shifts' ? 'default' : 'ghost'}
			size="sm"
			class="flex-1"
			onclick={() => (activeTab = 'shifts')}
		>
			<Users class="h-4 w-4 mr-2" />
			Shifts Leaderboard
		</Button>
	</div>

	<!-- Leaderboard Content -->
	<Card>
		<CardHeader>
			<CardTitle class="flex items-center gap-2">
				{#if activeTab === 'points'}
					<Trophy class="h-5 w-5" />
					Points Leaderboard
				{:else}
					<Users class="h-5 w-5" />
					Shifts Leaderboard
				{/if}
			</CardTitle>
		</CardHeader>
		<CardContent>
			{#if loading}
				<div class="flex items-center justify-center py-8">
					<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
				</div>
			{:else if error}
				<div class="text-center py-8 text-destructive">
					{error}
				</div>
			{:else}
				<div class="space-y-2">
					{#each activeTab === 'points' ? pointsLeaderboard : shiftsLeaderboard as entry (entry.user_id)}
						{@const rankInfo = getRankIcon(entry.rank)}
						<div
							class="flex items-center justify-between p-3 rounded-lg bg-muted/50 hover:bg-muted transition-colors"
						>
							<div class="flex items-center gap-3">
								<div class="flex items-center justify-center w-8 h-8">
									{#if entry.rank <= 3}
										<svelte:component this={rankInfo.icon} class="h-5 w-5 {rankInfo.class}" />
									{:else}
										<span class="text-sm font-medium text-muted-foreground">#{entry.rank}</span>
									{/if}
								</div>

								<Avatar class="h-10 w-10">
									<AvatarFallback>
										{getInitials(entry.name)}
									</AvatarFallback>
								</Avatar>

								<div>
									<div class="font-medium">{entry.name}</div>
									<div class="text-sm text-muted-foreground">
										{activeTab === 'points'
											? `${entry.shift_count} shifts`
											: `${entry.total_points} points`}
									</div>
								</div>
							</div>

							<div class="text-right">
								<div class="font-bold text-lg">
									{activeTab === 'points' ? entry.total_points : entry.shift_count}
								</div>
								<div class="text-xs text-muted-foreground">
									{activeTab === 'points' ? 'points' : 'shifts'}
								</div>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</CardContent>
	</Card>

	<!-- Recent Achievements -->
	{#if userStats && userStats.recent_achievements && userStats.recent_achievements.length > 0}
		<Card>
			<CardHeader>
				<CardTitle class="flex items-center gap-2">
					<Award class="h-5 w-5" />
					Recent Achievements
				</CardTitle>
			</CardHeader>
			<CardContent>
				<div class="space-y-2">
					{#each userStats.recent_achievements as achievement (achievement.achievement_id)}
						<div class="flex items-center gap-3 p-2 rounded-lg bg-muted/50">
							<div class="text-2xl">{achievement.icon}</div>
							<div class="flex-1">
								<div class="font-medium">{achievement.name}</div>
								<div class="text-sm text-muted-foreground">{achievement.description}</div>
							</div>
							<Badge variant="secondary" class="text-xs">
								{formatDate(achievement.earned_at)}
							</Badge>
						</div>
					{/each}
				</div>
			</CardContent>
		</Card>
	{/if}
</div>
