<script lang="ts">
	import { Skeleton } from '$lib/components/ui/skeleton';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { goto } from '$app/navigation';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import CheckCircleIcon from '@lucide/svelte/icons/check-circle';
	import TrendingUpIcon from '@lucide/svelte/icons/trending-up';
	import UsersIcon from '@lucide/svelte/icons/users';
	import UserCheckIcon from '@lucide/svelte/icons/user-check';
	import UserXIcon from '@lucide/svelte/icons/user-x';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import type { ShiftAnalytics } from '$lib/queries/admin/shifts/shiftsAnalyticsQuery';
	import { formatDistanceToNow } from 'date-fns';

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

	const metrics = $derived(analytics?.metrics);

	// Calculate system health for shifts
	const systemHealth = $derived.by(() => {
		if (!metrics) return null;

		const warnings = [];
		const issues = [];

		// Check fill rate
		if (metrics.fillRate < 50) {
			issues.push('Low shift fill rate (< 50%)');
		} else if (metrics.fillRate < 75) {
			warnings.push('Moderate shift fill rate (< 75%)');
		}

		// Check urgent unfilled shifts
		if (metrics.urgentUnfilled > 3) {
			issues.push(`${metrics.urgentUnfilled} urgent shifts unfilled (< 24h)`);
		} else if (metrics.urgentUnfilled > 0) {
			warnings.push(`${metrics.urgentUnfilled} urgent shifts unfilled (< 24h)`);
		}

		// Check critical unfilled shifts
		if (metrics.criticalUnfilled > 5) {
			issues.push(`${metrics.criticalUnfilled} critical shifts unfilled (< 72h)`);
		} else if (metrics.criticalUnfilled > 2) {
			warnings.push(`${metrics.criticalUnfilled} critical shifts unfilled (< 72h)`);
		}

		// Check schedule balance
		const imbalancedSchedules = metrics.scheduleBreakdown.filter((s) => s.fillRate < 40);
		if (imbalancedSchedules.length > 1) {
			warnings.push(`${imbalancedSchedules.length} schedules have low fill rates`);
		}

		return {
			status: issues.length > 0 ? 'critical' : warnings.length > 0 ? 'warning' : 'healthy',
			issues,
			warnings,
			score: Math.max(0, 100 - issues.length * 30 - warnings.length * 15)
		};
	});

	// Format shift title using watch tradition
	function formatShiftTitle(startTime: string, endTime: string): string {
		const start = new Date(startTime);
		const end = new Date(endTime);

		const previousDay = new Date(start);
		previousDay.setDate(previousDay.getDate() - 1);

		const dayName = previousDay.toLocaleDateString('en-US', { weekday: 'long', timeZone: 'UTC' });

		const startHour = start.getUTCHours();
		const endHour = end.getUTCHours();

		const formatHour = (hour: number) => (hour === 0 ? 12 : hour > 12 ? hour - 12 : hour);
		const getAmPm = (hour: number) => (hour < 12 ? 'AM' : 'PM');

		const startHour12 = formatHour(startHour);
		const endHour12 = formatHour(endHour);
		const endAmPm = getAmPm(endHour);

		const timeRange = `${startHour12}-${endHour12}${endAmPm}`;

		return `${dayName} Night ${timeRange}`;
	}

	function getUrgencyColor(shift: { start_time: string }): string {
		const shiftDate = new Date(shift.start_time);
		const now = new Date();
		const hoursUntil = (shiftDate.getTime() - now.getTime()) / (1000 * 60 * 60);

		if (hoursUntil <= 24) return 'red';
		if (hoursUntil <= 72) return 'orange';
		return 'green';
	}

	function getUrgencyText(shift: { start_time: string }): string {
		const shiftDate = new Date(shift.start_time);
		const now = new Date();
		const hoursUntil = (shiftDate.getTime() - now.getTime()) / (1000 * 60 * 60);

		if (hoursUntil <= 24) return 'Urgent';
		if (hoursUntil <= 72) return 'Critical';
		return 'Normal';
	}

	// Quick action handlers
	function handleBulkAssignment() {
		goto('/admin/shifts/bulk-signup');
	}

	function handleViewSchedules() {
		goto('/admin/schedules');
	}

	function handleManageUsers() {
		goto('/admin/users');
	}

	function handleCreateSchedule() {
		goto('/admin/schedules/new');
	}

	// Navigation function for shift details
	function navigateToShiftDetail(shift: { start_time: string; schedule_id: number }) {
		const shiftStartTime = encodeURIComponent(shift.start_time);
		goto(`/admin/shifts?shiftStartTime=${shiftStartTime}`);
	}
</script>

<div class="p-6">
	<div class="max-w-7xl mx-auto">
		<div class="mb-6">
			<h1 class="text-2xl font-bold flex items-center gap-2 mb-3">
				<CalendarIcon class="h-6 w-6" />
				Shifts Dashboard
			</h1>
			<p class="text-muted-foreground text-lg">
				Real-time overview of shift coverage, alerts, and assignment opportunities
			</p>
		</div>

		{#if isLoading}
			<!-- Loading Dashboard Skeletons -->
			<div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6 mb-8">
				{#each Array(4) as _, i (i)}
					<Card.Root class="p-6">
						<Skeleton class="h-4 w-24 mb-2" />
						<Skeleton class="h-8 w-16 mb-1" />
						<Skeleton class="h-3 w-20" />
					</Card.Root>
				{/each}
			</div>
			<div class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
				{#each Array(3) as _, i (i)}
					<Card.Root class="p-6">
						<Skeleton class="h-6 w-32 mb-4" />
						<Skeleton class="h-48 w-full" />
					</Card.Root>
				{/each}
			</div>
		{:else if isError}
			<div class="text-center py-16">
				<p class="text-destructive text-lg mb-2">Error Loading Shifts Dashboard</p>
				<p class="text-muted-foreground">
					{error?.message || 'Unknown error occurred'}
				</p>
			</div>
		{:else if analytics}
			<!-- Main Metrics Cards - Hidden on mobile -->
			<div class="hidden lg:grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6 mb-8">
				<!-- Fill Rate -->
				<Card.Root class="p-6">
					<div class="flex items-center gap-3">
						<div class="p-2 bg-purple-100 dark:bg-purple-900/20 rounded-lg">
							<TrendingUpIcon class="h-6 w-6 text-purple-600 dark:text-purple-400" />
						</div>
						<div>
							<p class="text-sm font-medium text-muted-foreground">Fill Rate</p>
							<p class="text-2xl font-bold">{metrics?.fillRate.toFixed(1) || 0}%</p>
							<p class="text-xs text-muted-foreground">
								{metrics?.filledShifts || 0} of {metrics?.totalShifts || 0} filled
							</p>
						</div>
					</div>
				</Card.Root>

				<!-- Urgent Unfilled -->
				<Card.Root class="p-6">
					<div class="flex items-center gap-3">
						<div
							class="p-2 rounded-lg {(metrics?.urgentUnfilled || 0) > 0
								? 'bg-red-100 dark:bg-red-900/20'
								: 'bg-green-100 dark:bg-green-900/20'}"
						>
							{#if (metrics?.urgentUnfilled || 0) > 0}
								<AlertTriangleIcon class="h-6 w-6 text-red-600 dark:text-red-400" />
							{:else}
								<CheckCircleIcon class="h-6 w-6 text-green-600 dark:text-green-400" />
							{/if}
						</div>
						<div>
							<p class="text-sm font-medium text-muted-foreground">Urgent (24h)</p>
							<p
								class="text-2xl font-bold {(metrics?.urgentUnfilled || 0) > 0
									? 'text-red-600 dark:text-red-400'
									: ''}"
							>
								{metrics?.urgentUnfilled || 0}
							</p>
							<p class="text-xs text-muted-foreground">Unfilled shifts</p>
						</div>
					</div>
				</Card.Root>

				<!-- This Week -->
				<Card.Root class="p-6">
					<div class="flex items-center gap-3">
						<div class="p-2 bg-blue-100 dark:bg-blue-900/20 rounded-lg">
							<CalendarIcon class="h-6 w-6 text-blue-600 dark:text-blue-400" />
						</div>
						<div>
							<p class="text-sm font-medium text-muted-foreground">This Week</p>
							<p class="text-2xl font-bold">{metrics?.thisWeekShifts || 0}</p>
							<p class="text-xs text-muted-foreground">Scheduled shifts</p>
						</div>
					</div>
				</Card.Root>

				<!-- System Health -->
				<Card.Root class="p-6">
					<div class="flex items-center gap-3">
						<div
							class="p-2 rounded-lg {systemHealth?.status === 'healthy'
								? 'bg-green-100 dark:bg-green-900/20'
								: systemHealth?.status === 'warning'
									? 'bg-yellow-100 dark:bg-yellow-900/20'
									: 'bg-red-100 dark:bg-red-900/20'}"
						>
							{#if systemHealth?.status === 'healthy'}
								<CheckCircleIcon class="h-6 w-6 text-green-600 dark:text-green-400" />
							{:else}
								<AlertTriangleIcon class="h-6 w-6 text-yellow-600 dark:text-yellow-400" />
							{/if}
						</div>
						<div>
							<p class="text-sm font-medium text-muted-foreground">System Health</p>
							<p class="text-2xl font-bold">{systemHealth?.score || 0}</p>
							<p class="text-xs text-muted-foreground capitalize">
								{systemHealth?.status || 'unknown'} status
							</p>
						</div>
					</div>
				</Card.Root>
			</div>

			<!-- Mobile-only compact metrics -->
			<div class="lg:hidden mb-6">
				<div class="grid grid-cols-2 gap-4 mb-4">
					<Card.Root class="p-4">
						<div class="text-center">
							<p class="text-xs text-muted-foreground">Fill Rate</p>
							<p class="text-xl font-bold">{metrics?.fillRate.toFixed(1) || 0}%</p>
						</div>
					</Card.Root>
					<Card.Root class="p-4">
						<div class="text-center">
							<p class="text-xs text-muted-foreground">Urgent (24h)</p>
							<p class="text-xl font-bold text-red-600">{metrics?.urgentUnfilled || 0}</p>
						</div>
					</Card.Root>
				</div>
				{#if (metrics?.urgentUnfilled || 0) > 0}
					<div
						class="bg-red-50 dark:bg-red-900/10 border border-red-200 dark:border-red-800 rounded-lg p-3 mb-4"
					>
						<div class="flex items-center gap-2">
							<AlertTriangleIcon class="h-4 w-4 text-red-600" />
							<p class="text-sm font-medium text-red-800 dark:text-red-200">
								{metrics?.urgentUnfilled || 0} urgent shift{(metrics?.urgentUnfilled || 0) > 1
									? 's'
									: ''} need immediate attention
							</p>
						</div>
					</div>
				{/if}
			</div>

			<!-- Time-based Metrics Row - Hidden on mobile -->
			<div class="hidden lg:grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
				<Card.Root class="p-6">
					<div class="flex items-center gap-3">
						<div class="p-2 bg-green-100 dark:bg-green-900/20 rounded-lg">
							<ClockIcon class="h-6 w-6 text-green-600 dark:text-green-400" />
						</div>
						<div>
							<p class="text-sm font-medium text-muted-foreground">Today</p>
							<p class="text-xl font-bold">{metrics?.todayShifts || 0}</p>
						</div>
					</div>
				</Card.Root>

				<Card.Root class="p-6">
					<div class="flex items-center gap-3">
						<div class="p-2 bg-blue-100 dark:bg-blue-900/20 rounded-lg">
							<ClockIcon class="h-6 w-6 text-blue-600 dark:text-blue-400" />
						</div>
						<div>
							<p class="text-sm font-medium text-muted-foreground">Tomorrow</p>
							<p class="text-xl font-bold">{metrics?.tomorrowShifts || 0}</p>
						</div>
					</div>
				</Card.Root>

				<Card.Root class="p-6">
					<div class="flex items-center gap-3">
						<div class="p-2 bg-purple-100 dark:bg-purple-900/20 rounded-lg">
							<CalendarIcon class="h-6 w-6 text-purple-600 dark:text-purple-400" />
						</div>
						<div>
							<p class="text-sm font-medium text-muted-foreground">Next Week</p>
							<p class="text-xl font-bold">{metrics?.nextWeekShifts || 0}</p>
						</div>
					</div>
				</Card.Root>

				<Card.Root class="p-6">
					<div class="flex items-center gap-3">
						<div class="p-2 bg-orange-100 dark:bg-orange-900/20 rounded-lg">
							<AlertTriangleIcon class="h-6 w-6 text-orange-600 dark:text-orange-400" />
						</div>
						<div>
							<p class="text-sm font-medium text-muted-foreground">Critical (72h)</p>
							<p class="text-xl font-bold">{metrics?.criticalUnfilled || 0}</p>
						</div>
					</div>
				</Card.Root>
			</div>

			<!-- Quick Actions - Hidden on mobile -->
			<Card.Root class="hidden lg:block p-6 mb-8">
				<div class="flex items-center gap-2 mb-4">
					<PlusIcon class="h-5 w-5 text-muted-foreground" />
					<h2 class="text-lg font-semibold">Quick Actions</h2>
				</div>
				<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
					<Button variant="outline" class="h-20 flex-col gap-2" onclick={handleBulkAssignment}>
						<UsersIcon class="h-5 w-5" />
						<span class="text-sm">Bulk Assignment</span>
					</Button>
					<Button variant="outline" class="h-20 flex-col gap-2" onclick={handleViewSchedules}>
						<CalendarIcon class="h-5 w-5" />
						<span class="text-sm">Manage Schedules</span>
					</Button>
					<Button variant="outline" class="h-20 flex-col gap-2" onclick={handleManageUsers}>
						<UserCheckIcon class="h-5 w-5" />
						<span class="text-sm">Manage Users</span>
					</Button>
					<Button variant="outline" class="h-20 flex-col gap-2" onclick={handleCreateSchedule}>
						<PlusIcon class="h-5 w-5" />
						<span class="text-sm">New Schedule</span>
					</Button>
				</div>
			</Card.Root>

			<!-- Content Grid -->
			<div class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
				<!-- Upcoming Unfilled Shifts -->
				<Card.Root class="p-6">
					<div class="flex items-center justify-between mb-4">
						<h2 class="text-lg font-semibold">Urgent Assignments Needed</h2>
						<Button variant="ghost" size="sm" onclick={handleBulkAssignment}>Assign</Button>
					</div>
					<div class="space-y-3">
						{#if metrics?.upcomingShifts}
							{@const urgentShifts = metrics.upcomingShifts.filter((s) => !s.is_booked).slice(0, 8)}
							{#if urgentShifts.length > 0}
								{#each urgentShifts as shift (shift.schedule_id + '-' + shift.start_time)}
									<div
										class="flex items-center gap-3 p-3 bg-muted/50 rounded-lg hover:bg-muted/70 transition-colors cursor-pointer group"
										onclick={() => navigateToShiftDetail(shift)}
										onkeydown={(e) => {
											if (e.key === 'Enter' || e.key === ' ') {
												e.preventDefault();
												navigateToShiftDetail(shift);
											}
										}}
										role="button"
										tabindex="0"
										aria-label={`View details for ${formatShiftTitle(shift.start_time, shift.end_time)} shift`}
									>
										<div class="flex-shrink-0">
											<div
												class="w-8 h-8 rounded-full flex items-center justify-center
												{getUrgencyColor(shift) === 'red'
													? 'bg-red-100 dark:bg-red-900/20'
													: getUrgencyColor(shift) === 'orange'
														? 'bg-orange-100 dark:bg-orange-900/20'
														: 'bg-green-100 dark:bg-green-900/20'}"
											>
												<UserXIcon
													class="h-4 w-4 text-{getUrgencyColor(
														shift
													)}-600 dark:text-{getUrgencyColor(shift)}-400"
												/>
											</div>
										</div>
										<div class="flex-grow min-w-0">
											<p
												class="font-medium text-sm truncate group-hover:text-primary transition-colors"
											>
												{formatShiftTitle(shift.start_time, shift.end_time)}
											</p>
											<p class="text-xs text-muted-foreground">
												{shift.schedule_name}
											</p>
										</div>
										<div class="flex-shrink-0 text-right">
											<Badge
												variant={getUrgencyColor(shift) === 'red' ? 'destructive' : 'secondary'}
												class="text-xs"
											>
												{getUrgencyText(shift)}
											</Badge>
											<p class="text-xs text-muted-foreground mt-1">
												{formatDistanceToNow(new Date(shift.start_time), { addSuffix: true })}
											</p>
										</div>
									</div>
								{/each}
							{:else}
								<div class="text-center py-8 text-muted-foreground">
									<CheckCircleIcon class="h-8 w-8 mx-auto mb-2 text-green-500" />
									<p class="text-sm">All upcoming shifts are filled!</p>
								</div>
							{/if}
						{/if}
					</div>
				</Card.Root>

				<!-- Schedule Performance -->
				<Card.Root class="p-6">
					<div class="flex items-center justify-between mb-4">
						<h2 class="text-lg font-semibold">Schedule Performance</h2>
						<Button variant="ghost" size="sm" onclick={handleViewSchedules}>View All</Button>
					</div>
					<div class="space-y-3">
						{#if metrics?.scheduleBreakdown && metrics.scheduleBreakdown.length > 0}
							{#each metrics.scheduleBreakdown as schedule (schedule.schedule_id)}
								<div class="flex items-center gap-3 p-3 bg-muted/50 rounded-lg">
									<div class="flex-shrink-0">
										<div
											class="w-8 h-8 rounded-full flex items-center justify-center
											{schedule.fillRate >= 75
												? 'bg-green-100 dark:bg-green-900/20'
												: schedule.fillRate >= 50
													? 'bg-yellow-100 dark:bg-yellow-900/20'
													: 'bg-red-100 dark:bg-red-900/20'}"
										>
											<CalendarIcon
												class="h-4 w-4 text-{schedule.fillRate >= 75
													? 'green'
													: schedule.fillRate >= 50
														? 'yellow'
														: 'red'}-600 dark:text-{schedule.fillRate >= 75
													? 'green'
													: schedule.fillRate >= 50
														? 'yellow'
														: 'red'}-400"
											/>
										</div>
									</div>
									<div class="flex-grow min-w-0">
										<p class="font-medium text-sm truncate">{schedule.schedule_name}</p>
										<p class="text-xs text-muted-foreground">
											{schedule.filled} of {schedule.total} filled
										</p>
									</div>
									<div class="flex-shrink-0 text-right">
										<p class="text-sm font-medium">{schedule.fillRate.toFixed(0)}%</p>
										<p class="text-xs text-muted-foreground">
											{schedule.available} available
										</p>
									</div>
								</div>
							{/each}
						{:else}
							<div class="text-center py-8 text-muted-foreground">
								<CalendarIcon class="h-8 w-8 mx-auto mb-2" />
								<p class="text-sm">No schedules found</p>
							</div>
						{/if}
					</div>
				</Card.Root>

				<!-- System Alerts -->
				<Card.Root class="p-6">
					<div class="flex items-center gap-2 mb-4">
						<AlertTriangleIcon class="h-5 w-5 text-muted-foreground" />
						<h2 class="text-lg font-semibold">System Alerts</h2>
					</div>
					<div class="space-y-3">
						{#if systemHealth?.issues && systemHealth.issues.length > 0}
							{#each systemHealth.issues as issue, index (index)}
								<div
									class="flex items-center gap-3 p-3 bg-red-50 dark:bg-red-900/10 rounded-lg border border-red-200 dark:border-red-800"
								>
									<AlertTriangleIcon class="h-4 w-4 text-red-600 dark:text-red-400 flex-shrink-0" />
									<p class="text-sm text-red-700 dark:text-red-400">{issue}</p>
								</div>
							{/each}
						{/if}
						{#if systemHealth?.warnings && systemHealth.warnings.length > 0}
							{#each systemHealth.warnings as warning, index (index)}
								<div
									class="flex items-center gap-3 p-3 bg-yellow-50 dark:bg-yellow-900/10 rounded-lg border border-yellow-200 dark:border-yellow-800"
								>
									<AlertTriangleIcon
										class="h-4 w-4 text-yellow-600 dark:text-yellow-400 flex-shrink-0"
									/>
									<p class="text-sm text-yellow-700 dark:text-yellow-400">{warning}</p>
								</div>
							{/each}
						{/if}
						{#if (!systemHealth?.issues || systemHealth.issues.length === 0) && (!systemHealth?.warnings || systemHealth.warnings.length === 0)}
							<div class="text-center py-8 text-muted-foreground">
								<CheckCircleIcon class="h-8 w-8 mx-auto mb-2 text-green-500" />
								<p class="text-sm">All systems operational</p>
							</div>
						{/if}
					</div>
				</Card.Root>
			</div>
		{/if}
	</div>
</div>
