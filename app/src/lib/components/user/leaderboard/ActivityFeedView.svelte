<script lang="ts">
	import { onMount } from 'svelte';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Avatar, AvatarFallback } from '$lib/components/ui/avatar';
	import { Activity, Trophy, Star, Clock, Calendar } from 'lucide-svelte';
	import { authenticatedFetch } from '$lib/utils/api';

	interface ActivityEntry {
		activity_id: number;
		user_name: string;
		action_type: string;
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

	function getActivityIcon(actionType: string) {
		switch (actionType) {
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

	function getInitials(name: string): string {
		return name
			.split(' ')
			.map((n) => n[0])
			.join('')
			.toUpperCase()
			.slice(0, 2);
	}

	function formatTimeAgo(dateStr: string): string {
		const date = new Date(dateStr);
		const now = new Date();
		const diffMs = now.getTime() - date.getTime();
		const diffMinutes = Math.floor(diffMs / (1000 * 60));
		const diffHours = Math.floor(diffMinutes / 60);
		const diffDays = Math.floor(diffHours / 24);

		if (diffMinutes < 60) {
			return `${diffMinutes}m ago`;
		} else if (diffHours < 24) {
			return `${diffHours}h ago`;
		} else if (diffDays < 7) {
			return `${diffDays}d ago`;
		} else {
			return date.toLocaleDateString();
		}
	}

	function getActionDescription(actionType: string, reason: string): string {
		switch (actionType) {
			case 'checkin':
				return 'checked in to their shift';
			case 'completion':
				return 'completed a shift';
			case 'achievement':
				return 'earned a new achievement';
			case 'bonus':
				return getReasonDescription(reason);
			default:
				return 'earned points';
		}
	}

	function getReasonDescription(reason: string): string {
		const reasonMap: Record<string, string> = {
			early_checkin: 'got early check-in bonus',
			weekend_bonus: 'completed weekend shift',
			late_night_bonus: 'worked late night shift',
			frequency_bonus: 'earned frequency bonus',
			level2_report: 'reported serious incident'
		};
		return reasonMap[reason] || 'earned bonus points';
	}

	onMount(() => {
		fetchActivity();
		// Refresh activity every 30 seconds
		const interval = setInterval(fetchActivity, 30000);
		return () => clearInterval(interval);
	});
</script>

<div class="space-y-6">
	<!-- Header -->
	<Card>
		<CardHeader>
			<CardTitle class="flex items-center gap-2">
				<Activity class="h-5 w-5" />
				Community Activity
			</CardTitle>
		</CardHeader>
		<CardContent>
			<div class="text-center text-muted-foreground">
				<p>See what your fellow Night Owls have been up to!</p>
				<p class="text-sm mt-1">Updates every 30 seconds</p>
			</div>
		</CardContent>
	</Card>

	<!-- Activity Feed -->
	<Card>
		<CardHeader>
			<CardTitle class="flex items-center gap-2">
				<Trophy class="h-5 w-5" />
				Recent Activity
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
			{:else if activityFeed.length === 0}
				<div class="text-center py-8 text-muted-foreground">
					<Activity class="h-12 w-12 mx-auto mb-4 opacity-50" />
					<p>No recent activity.</p>
					<p class="text-sm">Be the first to check in to a shift!</p>
				</div>
			{:else}
				<div class="space-y-4 max-h-96 overflow-y-auto">
					{#each activityFeed as activity (activity.activity_id)}
						{@const iconInfo = getActivityIcon(activity.action_type)}
						<div
							class="flex items-start gap-3 p-3 rounded-lg bg-muted/30 hover:bg-muted/50 transition-colors"
						>
							<Avatar class="h-8 w-8 flex-shrink-0">
								<AvatarFallback class="text-xs">
									{getInitials(activity.user_name)}
								</AvatarFallback>
							</Avatar>

							<div class="flex-1 min-w-0">
								<div class="flex items-center gap-2 mb-1">
									<span class="font-medium text-sm">{activity.user_name}</span>
									<span class="text-xs text-muted-foreground">
										{getActionDescription(activity.action_type, activity.reason)}
									</span>
								</div>

								<div class="flex items-center gap-2">
									<svelte:component this={iconInfo.icon} class="h-3 w-3 {iconInfo.color}" />
									<Badge variant="outline" class="text-xs">
										+{activity.points_awarded} pts
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
		</CardContent>
	</Card>

	<!-- Activity Stats -->
	<Card>
		<CardHeader>
			<CardTitle class="flex items-center gap-2">
				<Star class="h-5 w-5" />
				Today's Community Stats
			</CardTitle>
		</CardHeader>
		<CardContent>
			<div class="grid grid-cols-2 gap-4 text-center">
				<div>
					<div class="text-2xl font-bold text-primary">
						{activityFeed.filter((a) => a.action_type === 'checkin').length}
					</div>
					<div class="text-sm text-muted-foreground">Check-ins Today</div>
				</div>
				<div>
					<div class="text-2xl font-bold text-primary">
						{activityFeed.reduce((sum, a) => sum + a.points_awarded, 0)}
					</div>
					<div class="text-sm text-muted-foreground">Points Earned</div>
				</div>
			</div>
		</CardContent>
	</Card>
</div>
