<script lang="ts">
	import { onMount } from 'svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Activity, Trophy, Star, Clock, Calendar } from 'lucide-svelte';
	import { authenticatedFetch } from '$lib/utils/api';

	interface ActivityEntry {
		user_name: string;
		activity_type: string;
		points_awarded: number;
		reason: string;
		created_at: string;
	}

	let activityFeed: ActivityEntry[] = [];
	let loading = true;
	let error = '';

	async function fetchActivity() {
		try {
			const response = await authenticatedFetch('/api/leaderboard/activity');
			activityFeed = await response.json();
		} catch (err) {
			error = err instanceof Error ? err.message : 'An error occurred';
		} finally {
			loading = false;
		}
	}

	function getActivityIcon(activityType: string) {
		switch (activityType) {
			case 'checkin':
				return { icon: Clock, color: 'text-blue-600' };
			case 'completion':
				return { icon: Star, color: 'text-green-600' };
			case 'achievement':
				return { icon: Trophy, color: 'text-yellow-600' };
			case 'bonus':
				return { icon: Calendar, color: 'text-purple-600' };
			default:
				return { icon: Activity, color: 'text-gray-600' };
		}
	}

	function formatTimeAgo(dateStr: string): string {
		const date = new Date(dateStr);
		const now = new Date();
		const diffMs = now.getTime() - date.getTime();
		const diffMinutes = Math.floor(diffMs / (1000 * 60));
		const diffHours = Math.floor(diffMinutes / 60);
		const diffDays = Math.floor(diffHours / 24);

		if (diffMinutes < 60) return `${diffMinutes}m ago`;
		if (diffHours < 24) return `${diffHours}h ago`;
		if (diffDays < 7) return `${diffDays}d ago`;
		return date.toLocaleDateString();
	}

	function getActionDescription(activityType: string, reason: string): string {
		switch (activityType) {
			case 'checkin':
				return 'checked in';
			case 'completion':
				return 'completed shift';
			case 'achievement':
				return 'earned achievement';
			case 'bonus':
				return getReasonDescription(reason);
			default:
				return 'earned points';
		}
	}

	function getReasonDescription(reason: string): string {
		const reasonMap: Record<string, string> = {
			early_checkin: 'early check-in',
			weekend_bonus: 'weekend shift',
			late_night_bonus: 'late night shift',
			frequency_bonus: 'frequency bonus',
			level2_report: 'serious incident'
		};
		return reasonMap[reason] || 'bonus points';
	}

	onMount(() => {
		fetchActivity();
		const interval = setInterval(fetchActivity, 30000);
		return () => clearInterval(interval);
	});
</script>

<div class="space-y-4">
	<!-- Activity Feed -->
	{#if loading}
		<div class="flex justify-center py-8">
			<div class="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
		</div>
	{:else if error}
		<div class="text-center py-4 text-sm text-destructive">
			{error}
		</div>
	{:else if activityFeed.length === 0}
		<div class="text-center py-8 text-sm text-muted-foreground">No recent activity</div>
	{:else}
		<div class="space-y-2">
			{#each activityFeed as activity, index (`${index}-${activity.user_name}-${activity.created_at}-${activity.reason}`)}
				{@const iconInfo = getActivityIcon(activity.activity_type)}
				<div class="flex items-center gap-3 p-2 rounded-lg bg-background border">
					<!-- Content -->
					<div class="flex-1 min-w-0">
						<div class="flex items-center gap-2">
							<span class="font-medium text-sm truncate">{activity.user_name}</span>
							<Badge variant="outline" class="text-xs py-0 px-1">
								+{activity.points_awarded}
							</Badge>
							<span class="text-xs text-muted-foreground">
								{formatTimeAgo(activity.created_at)}
							</span>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
