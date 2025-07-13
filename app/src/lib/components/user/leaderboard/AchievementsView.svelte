<script lang="ts">
	import { onMount } from 'svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { authenticatedFetch } from '$lib/utils/api';

	interface Achievement {
		achievement_id: number;
		name: string;
		description: string;
		icon: string;
		shifts_threshold: number | null;
		earned_at: string | null;
	}

	interface AvailableAchievement {
		achievement_id: number;
		name: string;
		description: string;
		icon: string;
		shifts_threshold: number | null;
	}

	interface UserStats {
		total_points: number;
		shift_count: number;
		rank: number;
	}

	let earnedAchievements: Achievement[] = [];
	let availableAchievements: AvailableAchievement[] = [];
	let userStats: UserStats | null = null;
	let loading = true;
	let error = '';

	async function fetchAchievements() {
		try {
			const [earnedRes, availableRes, statsRes] = await Promise.all([
				authenticatedFetch('/api/user/achievements'),
				authenticatedFetch('/api/user/achievements/available'),
				authenticatedFetch('/api/user/stats')
			]);

			earnedAchievements = await earnedRes.json();
			availableAchievements = await availableRes.json();
			userStats = await statsRes.json();
		} catch (err) {
			error = err instanceof Error ? err.message : 'An error occurred';
		} finally {
			loading = false;
		}
	}

	function formatDate(dateStr: string): string {
		return new Date(dateStr).toLocaleDateString();
	}

	onMount(() => {
		fetchAchievements();
	});
</script>

<div class="space-y-4">
	{#if loading}
		<div class="flex justify-center py-8">
			<div class="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
		</div>
	{:else if error}
		<div class="text-center py-4 text-sm text-destructive">
			{error}
		</div>
	{:else}
		<!-- Earned Achievements -->
		{#if earnedAchievements.length > 0}
			<div class="space-y-2">
				{#each earnedAchievements as achievement (achievement.achievement_id)}
					<div class="flex items-center gap-3 p-2 rounded-lg bg-background border">
						<div class="text-2xl">{achievement.icon}</div>
						<div class="flex-1 min-w-0">
							<div class="font-medium text-sm truncate">{achievement.name}</div>
							<div class="text-xs text-muted-foreground">
								{achievement.earned_at ? formatDate(achievement.earned_at) : 'Earned'}
							</div>
						</div>
						<Badge variant="default" class="text-xs py-0 px-1">Earned</Badge>
					</div>
				{/each}
			</div>
		{/if}

		<!-- Available Achievements -->
		{#if availableAchievements.length > 0}
			<div class="space-y-2">
				{#each availableAchievements as achievement (achievement.achievement_id)}
					{@const current = userStats?.shift_count || 0}
					{@const threshold = achievement.shifts_threshold || 0}
					<div class="flex items-center gap-3 p-2 rounded-lg bg-background border opacity-60">
						<div class="text-2xl">{achievement.icon}</div>
						<div class="flex-1 min-w-0">
							<div class="font-medium text-sm truncate">{achievement.name}</div>
							<div class="text-xs text-muted-foreground">
								{achievement.shifts_threshold
									? `${current}/${threshold} shifts`
									: 'Complete more shifts'}
							</div>
						</div>
						<Badge variant="outline" class="text-xs py-0 px-1">
							{achievement.shifts_threshold ? `${threshold}` : 'Locked'}
						</Badge>
					</div>
				{/each}
			</div>
		{/if}

		<!-- Empty States -->
		{#if earnedAchievements.length === 0 && availableAchievements.length === 0}
			<div class="text-center py-8 text-sm text-muted-foreground">No achievements available</div>
		{:else if earnedAchievements.length === 0}
			<div class="text-center py-8 text-sm text-muted-foreground">No achievements earned yet</div>
		{/if}
	{/if}
</div>
