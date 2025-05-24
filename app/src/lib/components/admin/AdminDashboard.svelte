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
	import type { UserData } from '$lib/schemas/user';
	import type { AdminShiftSlot, Schedule } from '$lib/types';
	import { calculateUserMetrics } from '$lib/utils/userProcessing';
	import { calculateDashboardMetrics } from '$lib/utils/shiftProcessing';
	import { formatDistanceToNow } from 'date-fns';

	let {
		isLoading,
		isError,
		error,
		users,
		schedules,
		shifts
	}: {
		isLoading: boolean;
		isError: boolean;
		error?: Error;
		users?: UserData[];
		schedules?: Schedule[];
		shifts?: AdminShiftSlot[];
	} = $props();

	// Calculate metrics
	const userMetrics = $derived.by(() => {
		if (!users || users.length === 0) return null;
		return calculateUserMetrics(users);
	});

	const shiftMetrics = $derived.by(() => {
		if (!shifts) return null;
		return calculateDashboardMetrics(shifts);
	});

	const scheduleCount = $derived(schedules?.length || 0);

	// Recent activity calculations
	const recentUsers = $derived.by(() => {
		if (!users) return [];
		const sorted = [...users].sort(
			(a, b) => new Date(b.created_at || 0).getTime() - new Date(a.created_at || 0).getTime()
		);
		return sorted.slice(0, 5);
	});

	const upcomingShifts = $derived.by(() => {
		if (!shifts) return [];
		const now = new Date();
		const upcoming = shifts
			.filter((shift) => new Date(shift.start_time) > now)
			.sort((a, b) => new Date(a.start_time).getTime() - new Date(b.start_time).getTime());
		return upcoming.slice(0, 5);
	});

	// System health indicators
	const systemHealth = $derived.by(() => {
		if (!userMetrics || !shiftMetrics) return null;

		const warnings = [];
		const issues = [];

		// Check fill rate
		if (shiftMetrics.fillRate < 50) {
			issues.push('Low shift fill rate (< 50%)');
		} else if (shiftMetrics.fillRate < 75) {
			warnings.push('Moderate shift fill rate (< 75%)');
		}

		// Check user engagement
		const activeUsers = userMetrics.owlUsers + userMetrics.adminUsers;
		if (activeUsers < 5) {
			warnings.push('Few active volunteers (< 5)');
		}

		// Check upcoming shifts
		const upcomingCount = upcomingShifts?.length || 0;
		if (upcomingCount < 3) {
			warnings.push('Few upcoming shifts scheduled');
		}

		return {
			status: issues.length > 0 ? 'critical' : warnings.length > 0 ? 'warning' : 'healthy',
			issues,
			warnings,
			score: Math.max(0, 100 - issues.length * 30 - warnings.length * 15)
		};
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
		{:else}
			<!-- Main Metrics Cards -->
			<div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6 mb-8">
				<!-- Total Users -->
				<Card.Root class="p-6">
					<div class="flex items-center gap-3">
						<div class="p-2 bg-blue-100 dark:bg-blue-900/20 rounded-lg">
							<UsersIcon class="h-6 w-6 text-blue-600 dark:text-blue-400" />
						</div>
						<div>
							<p class="text-sm font-medium text-muted-foreground">Total Users</p>
							<p class="text-2xl font-bold">{userMetrics?.totalUsers || 0}</p>
							<p class="text-xs text-muted-foreground">
								{userMetrics?.adminUsers || 0} admins, {userMetrics?.owlUsers || 0} owls
							</p>
						</div>
					</div>
				</Card.Root>

				<!-- Active Schedules -->
				<Card.Root class="p-6">
					<div class="flex items-center gap-3">
						<div class="p-2 bg-green-100 dark:bg-green-900/20 rounded-lg">
							<CalendarIcon class="h-6 w-6 text-green-600 dark:text-green-400" />
						</div>
						<div>
							<p class="text-sm font-medium text-muted-foreground">Active Schedules</p>
							<p class="text-2xl font-bold">{scheduleCount}</p>
							<p class="text-xs text-muted-foreground">Generating shifts automatically</p>
						</div>
					</div>
				</Card.Root>

				<!-- Shift Fill Rate -->
				<Card.Root class="p-6">
					<div class="flex items-center gap-3">
						<div class="p-2 bg-purple-100 dark:bg-purple-900/20 rounded-lg">
							<ClockIcon class="h-6 w-6 text-purple-600 dark:text-purple-400" />
						</div>
						<div>
							<p class="text-sm font-medium text-muted-foreground">Fill Rate</p>
							<p class="text-2xl font-bold">{shiftMetrics?.fillRate.toFixed(1) || 0}%</p>
							<p class="text-xs text-muted-foreground">
								{shiftMetrics?.filledShifts || 0} of {shiftMetrics?.totalShifts || 0} filled
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
				<!-- Recent Users -->
				<Card.Root class="p-6">
					<div class="flex items-center justify-between mb-4">
						<h2 class="text-lg font-semibold">Recent Users</h2>
						<Button
							variant="ghost"
							size="sm"
							onclick={() => (window.location.href = '/admin/users')}
						>
							View All
						</Button>
					</div>
					<div class="space-y-3">
						{#if recentUsers && recentUsers.length > 0}
							{#each recentUsers as user (user.id)}
								<div class="flex items-center gap-3 p-3 bg-muted/50 rounded-lg">
									<div class="flex-shrink-0">
										{#if user.role === 'admin'}
											<div
												class="w-8 h-8 bg-red-100 dark:bg-red-900/20 rounded-full flex items-center justify-center"
											>
												<UsersIcon class="h-4 w-4 text-red-600 dark:text-red-400" />
											</div>
										{:else if user.role === 'owl'}
											<div
												class="w-8 h-8 bg-blue-100 dark:bg-blue-900/20 rounded-full flex items-center justify-center"
											>
												<UsersIcon class="h-4 w-4 text-blue-600 dark:text-blue-400" />
											</div>
										{:else}
											<div
												class="w-8 h-8 bg-gray-100 dark:bg-gray-900/20 rounded-full flex items-center justify-center"
											>
												<UsersIcon class="h-4 w-4 text-gray-600 dark:text-gray-400" />
											</div>
										{/if}
									</div>
									<div class="flex-grow min-w-0">
										<p class="font-medium text-sm truncate">{user.name || 'Unnamed User'}</p>
										<p class="text-xs text-muted-foreground capitalize">{user.role}</p>
									</div>
									<div class="flex-shrink-0 text-right">
										<p class="text-xs text-muted-foreground">
											{user.created_at
												? formatDistanceToNow(new Date(user.created_at), { addSuffix: true })
												: 'Recently'}
										</p>
									</div>
								</div>
							{/each}
						{:else}
							<div class="text-center py-8 text-muted-foreground">
								<UsersIcon class="h-8 w-8 mx-auto mb-2" />
								<p class="text-sm">No users found</p>
							</div>
						{/if}
					</div>
				</Card.Root>

				<!-- Upcoming Shifts -->
				<Card.Root class="p-6">
					<div class="flex items-center justify-between mb-4">
						<h2 class="text-lg font-semibold">Upcoming Shifts</h2>
						<Button
							variant="ghost"
							size="sm"
							onclick={() => (window.location.href = '/admin/shifts')}
						>
							View All
						</Button>
					</div>
					<div class="space-y-3">
						{#if upcomingShifts && upcomingShifts.length > 0}
							{#each upcomingShifts as shift (shift.start_time)}
								<div class="flex items-center gap-3 p-3 bg-muted/50 rounded-lg">
									<div class="flex-shrink-0">
										{#if shift.user_name}
											<div
												class="w-8 h-8 bg-green-100 dark:bg-green-900/20 rounded-full flex items-center justify-center"
											>
												<CheckCircleIcon class="h-4 w-4 text-green-600 dark:text-green-400" />
											</div>
										{:else}
											<div
												class="w-8 h-8 bg-yellow-100 dark:bg-yellow-900/20 rounded-full flex items-center justify-center"
											>
												<ClockIcon class="h-4 w-4 text-yellow-600 dark:text-yellow-400" />
											</div>
										{/if}
									</div>
									<div class="flex-grow min-w-0">
										<p class="font-medium text-sm truncate">{shift.schedule_name}</p>
										<p class="text-xs text-muted-foreground">
											{shift.user_name || 'Unassigned'}
										</p>
									</div>
									<div class="flex-shrink-0 text-right">
										<p class="text-xs text-muted-foreground">
											{formatDistanceToNow(new Date(shift.start_time), { addSuffix: true })}
										</p>
									</div>
								</div>
							{/each}
						{:else}
							<div class="text-center py-8 text-muted-foreground">
								<ClockIcon class="h-8 w-8 mx-auto mb-2" />
								<p class="text-sm">No upcoming shifts</p>
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
					</div>
				</Card.Root>
			</div>
		{/if}
	</div>
</div>
