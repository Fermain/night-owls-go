<script lang="ts">
	import { Skeleton } from '$lib/components/ui/skeleton';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import PlayIcon from '@lucide/svelte/icons/play';
	import PauseIcon from '@lucide/svelte/icons/pause';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import CheckCircleIcon from '@lucide/svelte/icons/check-circle';
	import TrendingUpIcon from '@lucide/svelte/icons/trending-up';
	import type { Schedule, AdminShiftSlot } from '$lib/types';
	import { formatDistanceToNow } from 'date-fns';

	let {
		isLoading,
		isError,
		error,
		schedules,
		shifts
	}: {
		isLoading: boolean;
		isError: boolean;
		error?: Error;
		schedules?: Schedule[];
		shifts?: AdminShiftSlot[];
	} = $props();

	// Calculate schedule metrics
	const scheduleMetrics = $derived.by(() => {
		if (!schedules) return null;

		const total = schedules.length;
		const active = schedules.filter((s) => s.is_active).length;
		const inactive = total - active;

		// Calculate coverage by checking upcoming shifts per schedule
		const scheduleShiftCounts = new Map<string, number>();
		if (shifts) {
			const now = new Date();
			const upcoming = shifts.filter((shift) => new Date(shift.start_time) > now);

			upcoming.forEach((shift) => {
				const current = scheduleShiftCounts.get(shift.schedule_name) || 0;
				scheduleShiftCounts.set(shift.schedule_name, current + 1);
			});
		}

		return {
			total,
			active,
			inactive,
			coverage: scheduleShiftCounts.size,
			averageShiftsPerSchedule:
				scheduleShiftCounts.size > 0
					? Math.round(
							Array.from(scheduleShiftCounts.values()).reduce((a, b) => a + b, 0) /
								scheduleShiftCounts.size
						)
					: 0
		};
	});

	// Recent schedules
	const recentSchedules = $derived.by(() => {
		if (!schedules) return [];
		return [...schedules]
			.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
			.slice(0, 5);
	});

	// Schedule performance analysis
	const schedulePerformance = $derived.by(() => {
		if (!schedules || !shifts) return [];

		const now = new Date();
		const upcoming = shifts.filter((shift) => new Date(shift.start_time) > now);

		const performance = schedules
			.map((schedule) => {
				const scheduleShifts = upcoming.filter((shift) => shift.schedule_name === schedule.name);
				const totalShifts = scheduleShifts.length;
				const filledShifts = scheduleShifts.filter((shift) => shift.is_booked).length;
				const fillRate = totalShifts > 0 ? (filledShifts / totalShifts) * 100 : 0;

				return {
					schedule,
					totalShifts,
					filledShifts,
					fillRate: Math.round(fillRate)
				};
			})
			.sort((a, b) => b.fillRate - a.fillRate);

		return performance.slice(0, 5);
	});

	// Schedule health indicators
	const scheduleHealth = $derived.by(() => {
		if (!scheduleMetrics || !schedulePerformance) return null;

		const warnings = [];
		const issues = [];

		// Check for inactive schedules
		if (scheduleMetrics.inactive > 0) {
			warnings.push(`${scheduleMetrics.inactive} inactive schedules`);
		}

		// Check for schedules with low coverage
		const lowPerformanceSchedules = schedulePerformance.filter((p) => p.fillRate < 50);
		if (lowPerformanceSchedules.length > 0) {
			issues.push(`${lowPerformanceSchedules.length} schedules with low fill rate`);
		}

		// Check for schedules with no upcoming shifts
		const noShiftSchedules = schedulePerformance.filter((p) => p.totalShifts === 0);
		if (noShiftSchedules.length > 0) {
			warnings.push(`${noShiftSchedules.length} schedules with no upcoming shifts`);
		}

		return {
			status: issues.length > 0 ? 'critical' : warnings.length > 0 ? 'warning' : 'healthy',
			issues,
			warnings,
			score: Math.max(0, 100 - issues.length * 25 - warnings.length * 10)
		};
	});

	// Quick action handlers
	function handleCreateSchedule() {
		window.location.href = '/admin/schedules/new';
	}

	function handleViewSlots() {
		window.location.href = '/admin/schedules/slots';
	}

	function handleViewShifts() {
		window.location.href = '/admin/shifts';
	}

	function handleRecurringAssignments() {
		window.location.href = '/admin/shifts/recurring';
	}
</script>

<div class="p-8">
	<div class="max-w-full mx-auto">
		<div class="mb-8">
			<h1 class="text-3xl font-semibold mb-3">Schedules Dashboard</h1>
			<p class="text-muted-foreground text-lg">
				Manage and monitor all automated shift schedules and their performance
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
				<p class="text-destructive text-lg mb-2">Error Loading Dashboard</p>
				<p class="text-muted-foreground">
					{error?.message || 'Unknown error occurred'}
				</p>
			</div>
		{:else if !schedules || schedules.length === 0}
			<div class="text-center py-16">
				<div class="max-w-md mx-auto">
					<CalendarIcon class="h-16 w-16 text-muted-foreground mx-auto mb-6" />
					<h2 class="text-2xl font-semibold mb-3">No Schedules Found</h2>
					<p class="text-muted-foreground mb-8">
						No schedules have been created yet. Create your first schedule to start generating
						automatic shifts.
					</p>
					<Button onclick={handleCreateSchedule} class="inline-flex items-center gap-2">
						<PlusIcon class="h-4 w-4" />
						Create First Schedule
					</Button>
				</div>
			</div>
		{:else}
			<!-- Main Metrics Cards -->
			<div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6 mb-8">
				<!-- Total Schedules -->
				<Card.Root class="p-6">
					<div class="flex items-center gap-3">
						<div class="p-2 bg-blue-100 dark:bg-blue-900/20 rounded-lg">
							<CalendarIcon class="h-6 w-6 text-blue-600 dark:text-blue-400" />
						</div>
						<div>
							<p class="text-sm font-medium text-muted-foreground">Total Schedules</p>
							<p class="text-2xl font-bold">{scheduleMetrics?.total || 0}</p>
							<p class="text-xs text-muted-foreground">
								{scheduleMetrics?.active || 0} active, {scheduleMetrics?.inactive || 0} inactive
							</p>
						</div>
					</div>
				</Card.Root>

				<!-- Active Schedules -->
				<Card.Root class="p-6">
					<div class="flex items-center gap-3">
						<div class="p-2 bg-green-100 dark:bg-green-900/20 rounded-lg">
							<PlayIcon class="h-6 w-6 text-green-600 dark:text-green-400" />
						</div>
						<div>
							<p class="text-sm font-medium text-muted-foreground">Active Schedules</p>
							<p class="text-2xl font-bold">{scheduleMetrics?.active || 0}</p>
							<p class="text-xs text-muted-foreground">Generating shifts automatically</p>
						</div>
					</div>
				</Card.Root>

				<!-- Schedule Coverage -->
				<Card.Root class="p-6">
					<div class="flex items-center gap-3">
						<div class="p-2 bg-purple-100 dark:bg-purple-900/20 rounded-lg">
							<ClockIcon class="h-6 w-6 text-purple-600 dark:text-purple-400" />
						</div>
						<div>
							<p class="text-sm font-medium text-muted-foreground">Coverage</p>
							<p class="text-2xl font-bold">{scheduleMetrics?.coverage || 0}</p>
							<p class="text-xs text-muted-foreground">Schedules with upcoming shifts</p>
						</div>
					</div>
				</Card.Root>

				<!-- Schedule Health -->
				<Card.Root class="p-6">
					<div class="flex items-center gap-3">
						<div
							class="p-2 rounded-lg {scheduleHealth?.status === 'healthy'
								? 'bg-green-100 dark:bg-green-900/20'
								: scheduleHealth?.status === 'warning'
									? 'bg-yellow-100 dark:bg-yellow-900/20'
									: 'bg-red-100 dark:bg-red-900/20'}"
						>
							{#if scheduleHealth?.status === 'healthy'}
								<CheckCircleIcon class="h-6 w-6 text-green-600 dark:text-green-400" />
							{:else}
								<AlertTriangleIcon class="h-6 w-6 text-yellow-600 dark:text-yellow-400" />
							{/if}
						</div>
						<div>
							<p class="text-sm font-medium text-muted-foreground">Health Score</p>
							<p class="text-2xl font-bold">{scheduleHealth?.score || 0}</p>
							<p class="text-xs text-muted-foreground capitalize">
								{scheduleHealth?.status || 'unknown'} status
							</p>
						</div>
					</div>
				</Card.Root>
			</div>

			<!-- Quick Actions -->
			<Card.Root class="p-6 mb-8">
				<div class="flex items-center gap-2 mb-4">
					<TrendingUpIcon class="h-5 w-5 text-muted-foreground" />
					<h2 class="text-lg font-semibold">Quick Actions</h2>
				</div>
				<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
					<Button variant="outline" class="h-20 flex-col gap-2" onclick={handleCreateSchedule}>
						<PlusIcon class="h-5 w-5" />
						<span class="text-sm">New Schedule</span>
					</Button>
					<Button variant="outline" class="h-20 flex-col gap-2" onclick={handleViewSlots}>
						<CalendarIcon class="h-5 w-5" />
						<span class="text-sm">View Slots</span>
					</Button>
					<Button variant="outline" class="h-20 flex-col gap-2" onclick={handleViewShifts}>
						<ClockIcon class="h-5 w-5" />
						<span class="text-sm">Manage Shifts</span>
					</Button>
					<Button
						variant="outline"
						class="h-20 flex-col gap-2"
						onclick={handleRecurringAssignments}
					>
						<PlayIcon class="h-5 w-5" />
						<span class="text-sm">Recurring</span>
					</Button>
				</div>
			</Card.Root>

			<!-- Content Grid -->
			<div class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
				<!-- Recent Schedules -->
				<Card.Root class="p-6">
					<div class="flex items-center justify-between mb-4">
						<h2 class="text-lg font-semibold">Recent Schedules</h2>
						<Button
							variant="ghost"
							size="sm"
							onclick={() => (window.location.href = '/admin/schedules')}
						>
							View All
						</Button>
					</div>
					<div class="space-y-3">
						{#if recentSchedules && recentSchedules.length > 0}
							{#each recentSchedules as schedule (schedule.schedule_id)}
								<div class="flex items-center gap-3 p-3 bg-muted/50 rounded-lg">
									<div class="flex-shrink-0">
										{#if schedule.is_active}
											<div
												class="w-8 h-8 bg-green-100 dark:bg-green-900/20 rounded-full flex items-center justify-center"
											>
												<PlayIcon class="h-4 w-4 text-green-600 dark:text-green-400" />
											</div>
										{:else}
											<div
												class="w-8 h-8 bg-gray-100 dark:bg-gray-900/20 rounded-full flex items-center justify-center"
											>
												<PauseIcon class="h-4 w-4 text-gray-600 dark:text-gray-400" />
											</div>
										{/if}
									</div>
									<div class="flex-grow min-w-0">
										<p class="font-medium text-sm truncate">{schedule.name}</p>
										<p class="text-xs text-muted-foreground">{schedule.cron_expr}</p>
									</div>
									<div class="flex-shrink-0 text-right">
										<p class="text-xs text-muted-foreground">
											{formatDistanceToNow(new Date(schedule.created_at), { addSuffix: true })}
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

				<!-- Schedule Performance -->
				<Card.Root class="p-6">
					<div class="flex items-center justify-between mb-4">
						<h2 class="text-lg font-semibold">Performance</h2>
						<Button
							variant="ghost"
							size="sm"
							onclick={() => (window.location.href = '/admin/schedules/slots')}
						>
							View Details
						</Button>
					</div>
					<div class="space-y-3">
						{#if schedulePerformance && schedulePerformance.length > 0}
							{#each schedulePerformance as perf (perf.schedule.schedule_id)}
								<div class="flex items-center gap-3 p-3 bg-muted/50 rounded-lg">
									<div class="flex-shrink-0">
										<div
											class="w-8 h-8 bg-blue-100 dark:bg-blue-900/20 rounded-full flex items-center justify-center"
										>
											<span class="text-xs font-bold text-blue-600 dark:text-blue-400">
												{perf.fillRate}%
											</span>
										</div>
									</div>
									<div class="flex-grow min-w-0">
										<p class="font-medium text-sm truncate">{perf.schedule.name}</p>
										<p class="text-xs text-muted-foreground">
											{perf.filledShifts} of {perf.totalShifts} shifts filled
										</p>
									</div>
									<div class="flex-shrink-0">
										<div class="w-16 bg-gray-200 dark:bg-gray-700 rounded-full h-2">
											<div
												class="h-2 rounded-full {perf.fillRate >= 75
													? 'bg-green-500'
													: perf.fillRate >= 50
														? 'bg-yellow-500'
														: 'bg-red-500'}"
												style="width: {perf.fillRate}%"
											></div>
										</div>
									</div>
								</div>
							{/each}
						{:else}
							<div class="text-center py-8 text-muted-foreground">
								<ClockIcon class="h-8 w-8 mx-auto mb-2" />
								<p class="text-sm">No performance data</p>
							</div>
						{/if}
					</div>
				</Card.Root>

				<!-- Schedule Health Alerts -->
				<Card.Root class="p-6">
					<div class="flex items-center gap-2 mb-4">
						<AlertTriangleIcon class="h-5 w-5 text-muted-foreground" />
						<h2 class="text-lg font-semibold">Health Alerts</h2>
					</div>
					<div class="space-y-3">
						{#if scheduleHealth?.issues && scheduleHealth.issues.length > 0}
							{#each scheduleHealth.issues as issue, index (index)}
								<div
									class="flex items-center gap-3 p-3 bg-red-50 dark:bg-red-900/10 rounded-lg border border-red-200 dark:border-red-800"
								>
									<AlertTriangleIcon class="h-4 w-4 text-red-600 dark:text-red-400 flex-shrink-0" />
									<p class="text-sm text-red-700 dark:text-red-400">{issue}</p>
								</div>
							{/each}
						{/if}
						{#if scheduleHealth?.warnings && scheduleHealth.warnings.length > 0}
							{#each scheduleHealth.warnings as warning, index (index)}
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
						{#if (!scheduleHealth?.issues || scheduleHealth.issues.length === 0) && (!scheduleHealth?.warnings || scheduleHealth.warnings.length === 0)}
							<div class="text-center py-8 text-muted-foreground">
								<CheckCircleIcon class="h-8 w-8 mx-auto mb-2 text-green-500" />
								<p class="text-sm">All schedules healthy</p>
							</div>
						{/if}
					</div>
				</Card.Root>
			</div>
		{/if}
	</div>
</div>
