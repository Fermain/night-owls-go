<script lang="ts">
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Avatar, AvatarFallback } from '$lib/components/ui/avatar';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { goto } from '$app/navigation';
	import {
		AlertTriangleIcon,
		CheckCircleIcon,
		UserXIcon,
		UserCheckIcon,
		TrophyIcon,
		CalendarIcon,
		TrendingUpIcon,
		UsersIcon
	} from 'lucide-svelte';
	import type { ShiftAnalytics } from '$lib/queries/admin/shifts/shiftsAnalyticsQuery';
	import { createLeaderboardQuery } from '$lib/queries/admin/leaderboardQuery';
	import { generateThumbnailDateInfo } from '$lib/utils/adminDialogs';
	import { formatShiftTimeRangeLocal } from '$lib/utils/shifts';
	import { formatDayNight } from '$lib/utils/shiftFormatting';

	let {
		isLoading,
		isError,
		error,
		analytics
	}: {
		isLoading: boolean;
		isError: boolean;
		error?: Error;
		analytics?: ShiftAnalytics;
	} = $props();

	// Create leaderboard query for real data
	const leaderboardQuery = $derived(createLeaderboardQuery());

	// Extract key data
	const metrics = $derived(analytics?.metrics);
	const shiftsToFill = $derived(
		metrics?.upcomingShifts?.filter((s) => !s.is_booked).slice(0, 8) || []
	);
	const shiftsFilled = $derived(metrics?.recentBookings?.slice(0, 8) || []);

	// Use real leaderboard data (combined from both points and shifts rankings)
	const leaderboardData = $derived.by(() => {
		if (!$leaderboardQuery.data || $leaderboardQuery.isError) {
			// Fallback data when API fails - use meaningful demo data
			return [
				{ name: 'Alice Admin', shifts: 5, points: 100, rank: 1, user_id: 1 },
				{ name: 'Charlie Volunteer', shifts: 3, points: 75, rank: 2, user_id: 3 },
				{ name: 'Diana Scout', shifts: 2, points: 50, rank: 3, user_id: 4 },
				{ name: 'Eve Patrol', shifts: 1, points: 25, rank: 4, user_id: 5 },
				{ name: 'Frank Night Owl', shifts: 4, points: 60, rank: 5, user_id: 6 },
				{ name: 'Grace Guardian', shifts: 2, points: 40, rank: 6, user_id: 7 },
				{ name: 'Harry Helper', shifts: 1, points: 15, rank: 7, user_id: 8 },
				{ name: 'Ivy Watcher', shifts: 3, points: 45, rank: 8, user_id: 9 }
			];
		}

		// Use points leaderboard as primary source, limit to top 8
		return $leaderboardQuery.data.pointsLeaderboard.slice(0, 8).map((entry) => ({
			name: entry.name,
			shifts: entry.shift_count,
			points: entry.total_points,
			rank: entry.rank,
			user_id: entry.user_id
		}));
	});

	// Navigation handlers
	function handleShiftClick(shift: { start_time: string; schedule_id: number }) {
		const shiftStartTime = encodeURIComponent(shift.start_time);
		goto(`/admin/shifts?shiftStartTime=${shiftStartTime}`);
	}

	function handleBulkAssignment() {
		goto('/admin/shifts/bulk-signup');
	}

	function handleViewAllUsers() {
		goto('/admin/users');
	}

	function getInitials(name: string): string {
		return name
			.split(' ')
			.map((n) => n[0])
			.join('')
			.toUpperCase()
			.slice(0, 2);
	}

	function getRankIcon(rank: number) {
		switch (rank) {
			case 1:
				return { class: 'text-yellow-500', symbol: '🥇' };
			case 2:
				return { class: 'text-gray-400', symbol: '🥈' };
			case 3:
				return { class: 'text-amber-600', symbol: '🥉' };
			default:
				return { class: 'text-muted-foreground', symbol: `#${rank}` };
		}
	}
</script>

<div class="p-6">
	<div class="max-w-7xl mx-auto">
		<!-- Header -->
		<div class="mb-6">
			<h1 class="text-3xl font-bold tracking-tight mb-2">Shifts Dashboard</h1>
		</div>

		{#if isLoading}
			<!-- Loading State -->
			<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
				{#each Array(3) as _, i (i)}
					<Card>
						<CardHeader>
							<Skeleton class="h-6 w-32" />
						</CardHeader>
						<CardContent>
							<div class="space-y-3">
								{#each Array(5) as _, j (j)}
									<div class="flex items-center gap-3">
										<Skeleton class="h-10 w-10 rounded-full" />
										<div class="flex-1">
											<Skeleton class="h-4 w-full mb-2" />
											<Skeleton class="h-3 w-3/4" />
										</div>
									</div>
								{/each}
							</div>
						</CardContent>
					</Card>
				{/each}
			</div>
		{:else if isError}
			<!-- Error State -->
			<div class="text-center py-16">
				<AlertTriangleIcon class="h-12 w-12 text-destructive mx-auto mb-4" />
				<p class="text-destructive text-lg mb-2">Error Loading Dashboard</p>
				<p class="text-muted-foreground">{error?.message || 'Unknown error occurred'}</p>
			</div>
		{:else}
			<!-- Main Content Grid -->
			<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
				<!-- Column 1: Shifts to Fill -->
				<Card>
					<CardHeader class="pb-4">
						<div class="flex items-center justify-between">
							<CardTitle class="flex items-center gap-2">
								<AlertTriangleIcon class="h-5 w-5 text-red-500" />
								Shifts to Fill
							</CardTitle>
							<Badge variant="destructive" class="text-xs">
								{shiftsToFill.length} urgent
							</Badge>
						</div>
					</CardHeader>
					<CardContent>
						{#if shiftsToFill.length > 0}
							<div class="space-y-3">
								{#each shiftsToFill as shift (shift.schedule_id + '-' + shift.start_time)}
									{@const thumbnailDate = generateThumbnailDateInfo(shift)}
									<div
										class="flex items-center gap-3 p-3 bg-red-50 dark:bg-red-900/10 rounded-lg border border-red-200 dark:border-red-800 hover:bg-red-100 dark:hover:bg-red-900/20 transition-colors cursor-pointer"
										onclick={() => handleShiftClick(shift)}
										onkeydown={(e) => {
											if (e.key === 'Enter' || e.key === ' ') {
												e.preventDefault();
												handleShiftClick(shift);
											}
										}}
										role="button"
										tabindex="0"
									>
										<div class="flex-shrink-0">
											<div
												class="w-10 h-10 rounded-full bg-red-100 dark:bg-red-900/20 flex items-center justify-center"
											>
												<UserXIcon class="h-5 w-5 text-red-600" />
											</div>
										</div>
										<div class="flex-grow min-w-0">
											<p class="font-medium text-sm truncate">
												{formatDayNight(shift.start_time)}
											</p>
											<div class="flex items-center gap-2 text-xs text-muted-foreground mt-1">
												<span class="font-mono">{thumbnailDate.shortDate}</span>
												<span>•</span>
												<span>{formatShiftTimeRangeLocal(shift.start_time, shift.end_time)}</span>
											</div>
											<p class="text-xs text-muted-foreground truncate">{shift.schedule_name}</p>
										</div>
										<div class="flex-shrink-0">
											<Badge variant="destructive" class="text-xs">
												{thumbnailDate.relativeTime}
											</Badge>
										</div>
									</div>
								{/each}
								<div class="pt-2">
									<Button variant="outline" size="sm" class="w-full" onclick={handleBulkAssignment}>
										<UsersIcon class="h-4 w-4 mr-2" />
										Bulk Assignment
									</Button>
								</div>
							</div>
						{:else}
							<div class="text-center py-8">
								<CheckCircleIcon class="h-12 w-12 text-green-500 mx-auto mb-3" />
								<p class="font-medium text-green-700 dark:text-green-400 mb-1">
									All shifts filled!
								</p>
								<p class="text-sm text-muted-foreground">No urgent assignments needed</p>
							</div>
						{/if}
					</CardContent>
				</Card>

				<!-- Column 2: Recently Filled Shifts -->
				<Card>
					<CardHeader class="pb-4">
						<div class="flex items-center justify-between">
							<CardTitle class="flex items-center gap-2">
								<CheckCircleIcon class="h-5 w-5 text-green-500" />
								Recently Filled
							</CardTitle>
							<Badge variant="secondary" class="text-xs">
								{metrics?.fillRate?.toFixed(1) || 0}% filled
							</Badge>
						</div>
					</CardHeader>
					<CardContent>
						{#if shiftsFilled.length > 0}
							<div class="space-y-3">
								{#each shiftsFilled as shift (shift.schedule_id + '-' + shift.start_time)}
									{@const thumbnailDate = generateThumbnailDateInfo(shift)}
									<div
										class="flex items-center gap-3 p-3 bg-green-50 dark:bg-green-900/10 rounded-lg border border-green-200 dark:border-green-800 hover:bg-green-100 dark:hover:bg-green-900/20 transition-colors cursor-pointer"
										onclick={() => handleShiftClick(shift)}
										onkeydown={(e) => {
											if (e.key === 'Enter' || e.key === ' ') {
												e.preventDefault();
												handleShiftClick(shift);
											}
										}}
										role="button"
										tabindex="0"
									>
										<div class="flex-shrink-0">
											<div
												class="w-10 h-10 rounded-full bg-green-100 dark:bg-green-900/20 flex items-center justify-center"
											>
												<UserCheckIcon class="h-5 w-5 text-green-600" />
											</div>
										</div>
										<div class="flex-grow min-w-0">
											<p class="font-medium text-sm truncate">
												{formatDayNight(shift.start_time)}
											</p>
											<div class="flex items-center gap-2 text-xs text-muted-foreground mt-1">
												<span class="font-mono">{thumbnailDate.shortDate}</span>
												<span>•</span>
												<span>{formatShiftTimeRangeLocal(shift.start_time, shift.end_time)}</span>
											</div>
											<p class="text-xs text-green-600 font-medium truncate">
												{shift.user_name || 'Assigned'}
												{#if shift.buddy_name}
													+ {shift.buddy_name}
												{/if}
											</p>
										</div>
										<div class="flex-shrink-0">
											<Badge variant="secondary" class="text-xs bg-green-100 text-green-700">
												{thumbnailDate.relativeTime}
											</Badge>
										</div>
									</div>
								{/each}
							</div>
						{:else}
							<div class="text-center py-8">
								<CalendarIcon class="h-12 w-12 text-muted-foreground mx-auto mb-3" />
								<p class="font-medium text-muted-foreground mb-1">No recent assignments</p>
								<p class="text-sm text-muted-foreground">Shift assignments will appear here</p>
							</div>
						{/if}
					</CardContent>
				</Card>

				<!-- Column 3: Top Performers Leaderboard - Now with Real Data -->
				<Card>
					<CardHeader class="pb-4">
						<div class="flex items-center justify-between">
							<CardTitle class="flex items-center gap-2">
								<TrophyIcon class="h-5 w-5 text-yellow-500" />
								Top Performers
							</CardTitle>
							<Badge variant="secondary" class="text-xs">
								{$leaderboardQuery.isLoading
									? 'Loading...'
									: $leaderboardQuery.isError || !$leaderboardQuery.data
										? 'Fallback data'
										: 'Live data'}
							</Badge>
						</div>
					</CardHeader>
					<CardContent>
						{#if $leaderboardQuery.isLoading}
							<!-- Loading leaderboard skeleton -->
							<div class="space-y-3">
								{#each Array(8) as _, i (i)}
									<div class="flex items-center gap-3 p-3 bg-muted/50 rounded-lg">
										<Skeleton class="h-8 w-8" />
										<Skeleton class="h-10 w-10 rounded-full" />
										<div class="flex-grow">
											<Skeleton class="h-4 w-full mb-2" />
											<Skeleton class="h-3 w-3/4" />
										</div>
										<Skeleton class="h-4 w-4" />
									</div>
								{/each}
							</div>
						{:else if $leaderboardQuery.isError}
							<!-- Error state for leaderboard -->
							<div class="text-center py-8">
								<AlertTriangleIcon class="h-8 w-8 text-destructive mx-auto mb-2" />
								<p class="text-sm text-destructive">Failed to load leaderboard</p>
							</div>
						{:else if leaderboardData.length > 0}
							<!-- Real leaderboard data -->
							<div class="space-y-3">
								{#each leaderboardData as user (user.user_id)}
									{@const rankInfo = getRankIcon(user.rank)}
									<div
										class="flex items-center gap-3 p-3 bg-muted/50 rounded-lg hover:bg-muted transition-colors"
									>
										<div class="flex-shrink-0">
											<div class="w-8 h-8 flex items-center justify-center">
												<span class="text-lg font-bold {rankInfo.class}">
													{rankInfo.symbol}
												</span>
											</div>
										</div>
										<Avatar class="h-10 w-10">
											<AvatarFallback class="text-xs font-medium">
												{getInitials(user.name)}
											</AvatarFallback>
										</Avatar>
										<div class="flex-grow min-w-0">
											<p class="font-medium text-sm truncate">{user.name}</p>
											<div class="flex items-center gap-3 text-xs text-muted-foreground">
												<span>{user.shifts} shifts</span>
												<span>•</span>
												<span>{user.points} points</span>
											</div>
										</div>
										<div class="flex-shrink-0">
											<TrendingUpIcon class="h-4 w-4 text-green-500" />
										</div>
									</div>
								{/each}
								<div class="pt-2">
									<Button variant="outline" size="sm" class="w-full" onclick={handleViewAllUsers}>
										<UsersIcon class="h-4 w-4 mr-2" />
										View All Users
									</Button>
								</div>
							</div>
						{:else}
							<!-- Empty state -->
							<div class="text-center py-8">
								<TrophyIcon class="h-12 w-12 text-muted-foreground mx-auto mb-3" />
								<p class="font-medium text-muted-foreground mb-1">No leaderboard data</p>
								<p class="text-sm text-muted-foreground">User performance data will appear here</p>
							</div>
						{/if}
					</CardContent>
				</Card>
			</div>

			<!-- Summary Stats Row -->
			<div class="mt-6 grid grid-cols-1 md:grid-cols-4 gap-4">
				<Card class="p-4">
					<div class="text-center">
						<div class="text-2xl font-bold text-primary">{metrics?.totalShifts || 0}</div>
						<div class="text-sm text-muted-foreground">Total Shifts</div>
					</div>
				</Card>
				<Card class="p-4">
					<div class="text-center">
						<div class="text-2xl font-bold text-green-600">{metrics?.filledShifts || 0}</div>
						<div class="text-sm text-muted-foreground">Filled</div>
					</div>
				</Card>
				<Card class="p-4">
					<div class="text-center">
						<div class="text-2xl font-bold text-red-600">{metrics?.availableShifts || 0}</div>
						<div class="text-sm text-muted-foreground">Available</div>
					</div>
				</Card>
				<Card class="p-4">
					<div class="text-center">
						<div class="text-2xl font-bold text-primary">{metrics?.fillRate?.toFixed(1) || 0}%</div>
						<div class="text-sm text-muted-foreground">Fill Rate</div>
					</div>
				</Card>
			</div>
		{/if}
	</div>
</div>
