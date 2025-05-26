<script lang="ts">
	import { Skeleton } from '$lib/components/ui/skeleton';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import UsersIcon from '@lucide/svelte/icons/users';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import SettingsIcon from '@lucide/svelte/icons/settings';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import CheckCircleIcon from '@lucide/svelte/icons/check-circle';
	import TrendingUpIcon from '@lucide/svelte/icons/trending-up';
	import UserCheckIcon from '@lucide/svelte/icons/user-check';
	import UserXIcon from '@lucide/svelte/icons/user-x';
	import TrendingDownIcon from '@lucide/svelte/icons/trending-down';
	import type { AdminDashboardData } from '$lib/queries/admin/dashboard';

	let {
		isLoading,
		isError,
		error,
		data
	}: {
		isLoading: boolean;
		isError: boolean;
		error?: Error;
		data?: AdminDashboardData;
	} = $props();

	// Extract metrics from dashboard data
	const metrics = $derived(data?.metrics);
	const memberContributions = $derived(data?.member_contributions || []);
	const qualityMetrics = $derived(data?.quality_metrics);
	const problematicSlots = $derived(data?.problematic_slots || []);

	// Calculate derived metrics
	const weekendStatusColor = $derived.by(() => {
		if (!metrics?.this_weekend_status) return 'gray';
		switch (metrics.this_weekend_status) {
			case 'fully_covered': return 'green';
			case 'partial_coverage': return 'yellow';
			case 'critical': return 'red';
			case 'no_shifts': return 'gray';
			default: return 'gray';
		}
	});

	const weekendStatusText = $derived.by(() => {
		if (!metrics?.this_weekend_status) return 'Unknown';
		switch (metrics.this_weekend_status) {
			case 'fully_covered': return 'Fully Covered';
			case 'partial_coverage': return 'Partial Coverage';
			case 'critical': return 'Critical';
			case 'no_shifts': return 'No Shifts';
			default: return 'Unknown';
		}
	});

	// System health based on backend data
	const systemHealth = $derived.by(() => {
		if (!metrics || !qualityMetrics) return null;

		const warnings = [];
		const issues = [];

		// Check fill rate
		if (metrics.fill_rate < 50) {
			issues.push('Low shift fill rate (< 50%)');
		} else if (metrics.fill_rate < 75) {
			warnings.push('Moderate shift fill rate (< 75%)');
		}

		// Check next week coverage
		if (metrics.next_week_unfilled > 5) {
			issues.push(`${metrics.next_week_unfilled} unfilled shifts next week`);
		} else if (metrics.next_week_unfilled > 2) {
			warnings.push(`${metrics.next_week_unfilled} unfilled shifts next week`);
		}

		// Check weekend status
		if (metrics.this_weekend_status === 'critical') {
			issues.push('Critical weekend coverage shortage');
		} else if (metrics.this_weekend_status === 'partial_coverage') {
			warnings.push('Partial weekend coverage');
		}

		// Check reliability
		if (qualityMetrics.reliability_score < 60) {
			issues.push('Low reliability score (< 60%)');
		} else if (qualityMetrics.reliability_score < 80) {
			warnings.push('Moderate reliability score (< 80%)');
		}

		// Check no-show rate
		if (qualityMetrics.no_show_rate > 20) {
			issues.push('High no-show rate (> 20%)');
		} else if (qualityMetrics.no_show_rate > 10) {
			warnings.push('Elevated no-show rate (> 10%)');
		}

		return {
			status: issues.length > 0 ? 'critical' : warnings.length > 0 ? 'warning' : 'healthy',
			issues,
			warnings,
			score: Math.max(0, 100 - issues.length * 30 - warnings.length * 15)
		};
	});

	// Top contributors (sorted by contribution level)
	const topContributors = $derived.by(() => {
		if (!memberContributions) return [];
		const contributors = [...memberContributions]
			.filter(c => c.shifts_booked > 0)
			.sort((a, b) => b.shifts_booked - a.shifts_booked)
			.slice(0, 5);
		return contributors;
	});

	// Non-contributors who need encouragement
	const nonContributors = $derived.by(() => {
		if (!memberContributions) return [];
		return memberContributions.filter(c => c.contribution_category === 'non_contributor');
	});

	// Quick action handlers
	function handleCreateUser() {
		window.location.href = '/admin/users/new';
	}

	function handleCreateSchedule() {
		window.location.href = '/admin/schedules/new';
	}

	function handleViewShifts() {
		window.location.href = '/admin/shifts';
	}

	function handleViewReports() {
		window.location.href = '/admin/reports';
	}

	function formatContributionCategory(category: string): string {
		switch (category) {
			case 'non_contributor': return 'Not Contributing';
			case 'minimum_contributor': return 'Minimal';
			case 'fair_contributor': return 'Fair Share';
			case 'heavy_lifter': return 'Heavy Lifter';
			default: return category;
		}
	}

	function getContributionColor(category: string): string {
		switch (category) {
			case 'non_contributor': return 'text-red-600 dark:text-red-400';
			case 'minimum_contributor': return 'text-yellow-600 dark:text-yellow-400';
			case 'fair_contributor': return 'text-green-600 dark:text-green-400';
			case 'heavy_lifter': return 'text-blue-600 dark:text-blue-400';
			default: return 'text-gray-600 dark:text-gray-400';
		}
	}
</script>

<div class="p-8">
	<div class="max-w-full mx-auto">
		<div class="mb-8">
			<h1 class="text-3xl font-semibold mb-3">Admin Dashboard</h1>
			<p class="text-muted-foreground text-lg">
				Central hub for managing volunteers, schedules, and community watch operations
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
		{:else if data}
			<!-- Main Metrics Cards -->
			<div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6 mb-8">
				<!-- Shift Fill Rate -->
				<Card.Root class="p-6">
					<div class="flex items-center gap-3">
						<div class="p-2 bg-purple-100 dark:bg-purple-900/20 rounded-lg">
							<ClockIcon class="h-6 w-6 text-purple-600 dark:text-purple-400" />
						</div>
						<div>
							<p class="text-sm font-medium text-muted-foreground">Fill Rate</p>
							<p class="text-2xl font-bold">{metrics?.fill_rate.toFixed(1) || 0}%</p>
							<p class="text-xs text-muted-foreground">
								{metrics?.booked_shifts || 0} of {metrics?.total_shifts || 0} filled
							</p>
						</div>
					</div>
				</Card.Root>

				<!-- Check-in Rate -->
				<Card.Root class="p-6">
					<div class="flex items-center gap-3">
						<div class="p-2 bg-green-100 dark:bg-green-900/20 rounded-lg">
							<UserCheckIcon class="h-6 w-6 text-green-600 dark:text-green-400" />
						</div>
						<div>
							<p class="text-sm font-medium text-muted-foreground">Check-in Rate</p>
							<p class="text-2xl font-bold">{metrics?.check_in_rate.toFixed(1) || 0}%</p>
							<p class="text-xs text-muted-foreground">
								{metrics?.checked_in_shifts || 0} checked in
							</p>
						</div>
					</div>
				</Card.Root>

				<!-- Weekend Status -->
				<Card.Root class="p-6">
					<div class="flex items-center gap-3">
						<div class="p-2 rounded-lg {weekendStatusColor === 'green' 
							? 'bg-green-100 dark:bg-green-900/20'
							: weekendStatusColor === 'yellow'
								? 'bg-yellow-100 dark:bg-yellow-900/20'
								: weekendStatusColor === 'red'
									? 'bg-red-100 dark:bg-red-900/20'
									: 'bg-gray-100 dark:bg-gray-900/20'}">
							{#if weekendStatusColor === 'green'}
								<CheckCircleIcon class="h-6 w-6 text-green-600 dark:text-green-400" />
							{:else}
								<AlertTriangleIcon class="h-6 w-6 text-{weekendStatusColor}-600 dark:text-{weekendStatusColor}-400" />
							{/if}
						</div>
						<div>
							<p class="text-sm font-medium text-muted-foreground">This Weekend</p>
							<p class="text-lg font-bold">{weekendStatusText}</p>
							<p class="text-xs text-muted-foreground">
								{metrics?.next_week_unfilled || 0} unfilled next week
							</p>
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

			<!-- Quality Metrics Row -->
			<div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
				<Card.Root class="p-6">
					<div class="flex items-center gap-3">
						<div class="p-2 bg-blue-100 dark:bg-blue-900/20 rounded-lg">
							<TrendingUpIcon class="h-6 w-6 text-blue-600 dark:text-blue-400" />
						</div>
						<div>
							<p class="text-sm font-medium text-muted-foreground">Reliability Score</p>
							<p class="text-2xl font-bold">{qualityMetrics?.reliability_score.toFixed(1) || 0}%</p>
							<p class="text-xs text-muted-foreground">Overall completion rate</p>
						</div>
					</div>
				</Card.Root>

				<Card.Root class="p-6">
					<div class="flex items-center gap-3">
						<div class="p-2 bg-red-100 dark:bg-red-900/20 rounded-lg">
							<UserXIcon class="h-6 w-6 text-red-600 dark:text-red-400" />
						</div>
						<div>
							<p class="text-sm font-medium text-muted-foreground">No-Show Rate</p>
							<p class="text-2xl font-bold">{qualityMetrics?.no_show_rate.toFixed(1) || 0}%</p>
							<p class="text-xs text-muted-foreground">Didn't check in</p>
						</div>
					</div>
				</Card.Root>

				<Card.Root class="p-6">
					<div class="flex items-center gap-3">
						<div class="p-2 bg-yellow-100 dark:bg-yellow-900/20 rounded-lg">
							<TrendingDownIcon class="h-6 w-6 text-yellow-600 dark:text-yellow-400" />
						</div>
						<div>
							<p class="text-sm font-medium text-muted-foreground">Incomplete Rate</p>
							<p class="text-2xl font-bold">{qualityMetrics?.incomplete_rate.toFixed(1) || 0}%</p>
							<p class="text-xs text-muted-foreground">No report submitted</p>
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
					<Button variant="outline" class="h-20 flex-col gap-2" onclick={handleCreateUser}>
						<PlusIcon class="h-5 w-5" />
						<span class="text-sm">Add User</span>
					</Button>
					<Button variant="outline" class="h-20 flex-col gap-2" onclick={handleCreateSchedule}>
						<CalendarIcon class="h-5 w-5" />
						<span class="text-sm">New Schedule</span>
					</Button>
					<Button variant="outline" class="h-20 flex-col gap-2" onclick={handleViewShifts}>
						<ClockIcon class="h-5 w-5" />
						<span class="text-sm">Manage Shifts</span>
					</Button>
					<Button variant="outline" class="h-20 flex-col gap-2" onclick={handleViewReports}>
						<SettingsIcon class="h-5 w-5" />
						<span class="text-sm">View Reports</span>
					</Button>
				</div>
			</Card.Root>

			<!-- Content Grid -->
			<div class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
				<!-- Top Contributors -->
				<Card.Root class="p-6">
					<div class="flex items-center justify-between mb-4">
						<h2 class="text-lg font-semibold">Top Contributors</h2>
						<Button
							variant="ghost"
							size="sm"
							onclick={() => (window.location.href = '/admin/users')}
						>
							View All
						</Button>
					</div>
					<div class="space-y-3">
						{#if topContributors && topContributors.length > 0}
							{#each topContributors as contributor (contributor.user_id)}
								<div class="flex items-center gap-3 p-3 bg-muted/50 rounded-lg">
									<div class="flex-shrink-0">
										<div
											class="w-8 h-8 bg-blue-100 dark:bg-blue-900/20 rounded-full flex items-center justify-center"
										>
											<UsersIcon class="h-4 w-4 text-blue-600 dark:text-blue-400" />
										</div>
									</div>
									<div class="flex-grow min-w-0">
										<p class="font-medium text-sm truncate">{contributor.name}</p>
										<p class="text-xs {getContributionColor(contributor.contribution_category)}">
											{formatContributionCategory(contributor.contribution_category)} • {contributor.shifts_booked} shifts
										</p>
									</div>
									<div class="flex-shrink-0 text-right">
										<p class="text-xs text-muted-foreground">
											{contributor.attendance_rate.toFixed(0)}% attend
										</p>
									</div>
								</div>
							{/each}
						{:else}
							<div class="text-center py-8 text-muted-foreground">
								<UsersIcon class="h-8 w-8 mx-auto mb-2" />
								<p class="text-sm">No active contributors</p>
							</div>
						{/if}
					</div>
				</Card.Root>

				<!-- Problematic Time Slots -->
				<Card.Root class="p-6">
					<div class="flex items-center justify-between mb-4">
						<h2 class="text-lg font-semibold">Problem Time Slots</h2>
						<Button
							variant="ghost"
							size="sm"
							onclick={() => (window.location.href = '/admin/shifts')}
						>
							View Shifts
						</Button>
					</div>
					<div class="space-y-3">
						{#if problematicSlots && problematicSlots.length > 0}
							{#each problematicSlots as slot (slot.day_of_week + slot.hour_of_day)}
								<div class="flex items-center gap-3 p-3 bg-muted/50 rounded-lg">
									<div class="flex-shrink-0">
										<div
											class="w-8 h-8 bg-red-100 dark:bg-red-900/20 rounded-full flex items-center justify-center"
										>
											<AlertTriangleIcon class="h-4 w-4 text-red-600 dark:text-red-400" />
										</div>
									</div>
									<div class="flex-grow min-w-0">
										<p class="font-medium text-sm truncate">{slot.day_of_week} {slot.hour_of_day}</p>
										<p class="text-xs text-muted-foreground">
											{slot.total_bookings} bookings
										</p>
									</div>
									<div class="flex-shrink-0 text-right">
										<p class="text-xs text-red-600 dark:text-red-400">
											{slot.completion_rate.toFixed(0)}% complete
										</p>
									</div>
								</div>
							{/each}
						{:else}
							<div class="text-center py-8 text-muted-foreground">
								<CheckCircleIcon class="h-8 w-8 mx-auto mb-2 text-green-500" />
								<p class="text-sm">No problematic time slots</p>
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
							{#each systemHealth.issues as issue}
								<div
									class="flex items-center gap-3 p-3 bg-red-50 dark:bg-red-900/10 rounded-lg border border-red-200 dark:border-red-800"
								>
									<AlertTriangleIcon class="h-4 w-4 text-red-600 dark:text-red-400 flex-shrink-0" />
									<p class="text-sm text-red-700 dark:text-red-400">{issue}</p>
								</div>
							{/each}
						{/if}
						{#if systemHealth?.warnings && systemHealth.warnings.length > 0}
							{#each systemHealth.warnings as warning}
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

						<!-- Show non-contributors alert if any -->
						{#if nonContributors.length > 0}
							<div
								class="flex items-center gap-3 p-3 bg-blue-50 dark:bg-blue-900/10 rounded-lg border border-blue-200 dark:border-blue-800"
							>
								<UsersIcon class="h-4 w-4 text-blue-600 dark:text-blue-400 flex-shrink-0" />
								<p class="text-sm text-blue-700 dark:text-blue-400">
									{nonContributors.length} volunteers haven't taken shifts recently
								</p>
							</div>
						{/if}
					</div>
				</Card.Root>
			</div>
		{/if}
	</div>
</div>
