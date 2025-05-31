<script lang="ts">
	import { Skeleton } from '$lib/components/ui/skeleton';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import CheckCircleIcon from '@lucide/svelte/icons/check-circle';
	import UsersIcon from '@lucide/svelte/icons/users';
	import PhoneIcon from '@lucide/svelte/icons/phone';
	import MessageSquareIcon from '@lucide/svelte/icons/message-square';
	import ShieldAlertIcon from '@lucide/svelte/icons/shield-alert';
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

	// Critical status assessment
	const criticalStatus = $derived.by(() => {
		if (!metrics) return null;

		const issues = [];

		// Today's coverage issues
		if (metrics.fill_rate < 75) {
			issues.push({
				type: 'coverage',
				severity: metrics.fill_rate < 50 ? 'critical' : 'warning',
				message: `${metrics.fill_rate.toFixed(0)}% shifts filled`,
				action: 'Fill gaps'
			});
		}

		// Weekend coverage
		if (metrics.this_weekend_status === 'critical') {
			issues.push({
				type: 'weekend',
				severity: 'critical',
				message: 'Critical weekend shortage',
				action: 'Urgent: Find volunteers'
			});
		} else if (metrics.this_weekend_status === 'partial_coverage') {
			issues.push({
				type: 'weekend',
				severity: 'warning',
				message: 'Partial weekend coverage',
				action: 'Find backup volunteers'
			});
		}

		// No-show issues
		if (qualityMetrics && qualityMetrics.no_show_rate > 15) {
			issues.push({
				type: 'reliability',
				severity: qualityMetrics.no_show_rate > 25 ? 'critical' : 'warning',
				message: `${qualityMetrics.no_show_rate.toFixed(0)}% no-show rate`,
				action: 'Contact unreliable members'
			});
		}

		return {
			overallStatus: issues.some((i) => i.severity === 'critical')
				? 'critical'
				: issues.some((i) => i.severity === 'warning')
					? 'warning'
					: 'good',
			issues: issues.slice(0, 3) // Show max 3 critical issues
		};
	});

	// Today's shift status (simulated - would come from API)
	const todayStatus = $derived.by(() => {
		// This would be real data from API
		const now = new Date();
		const currentHour = now.getHours();

		return {
			currentShift: currentHour >= 18 && currentHour < 6 ? 'active' : 'none',
			currentVolunteer: currentHour >= 18 && currentHour < 6 ? 'Sarah M.' : null,
			nextShift: currentHour < 18 ? '18:00 - Weekend Patrol' : '18:00 Tomorrow',
			nextVolunteer: currentHour < 18 ? 'John D.' : 'TBD'
		};
	});

	// Non-contributors needing attention
	const needsAttention = $derived.by(() => {
		if (!memberContributions) return [];
		return memberContributions
			.filter((c) => c.contribution_category === 'non_contributor')
			.slice(0, 3);
	});

	// Quick actions
	function handleEmergencyBroadcast() {
		window.location.href = '/admin/broadcasts?emergency=true';
	}

	function handleFillGaps() {
		window.location.href = '/admin/shifts';
	}

	function handleContactMembers() {
		window.location.href = '/admin/users';
	}

	function handleViewReports() {
		window.location.href = '/admin/reports';
	}
</script>

<div class="space-y-4 p-4">
	{#if isLoading}
		<!-- Loading state optimized for mobile -->
		<div class="space-y-4">
			<Skeleton class="h-24 w-full rounded-lg" />
			<Skeleton class="h-32 w-full rounded-lg" />
			<div class="grid grid-cols-2 gap-3">
				<Skeleton class="h-20 rounded-lg" />
				<Skeleton class="h-20 rounded-lg" />
			</div>
		</div>
	{:else if isError}
		<Card.Root class="p-4 border-destructive bg-destructive/5">
			<div class="flex items-center gap-3">
				<AlertTriangleIcon class="h-5 w-5 text-destructive" />
				<div>
					<p class="font-medium text-destructive">Dashboard Error</p>
					<p class="text-sm text-muted-foreground">{error?.message || 'Failed to load'}</p>
				</div>
			</div>
		</Card.Root>
	{:else if data}
		<!-- CRITICAL: Current Status (Always visible at top) -->
		<Card.Root
			class="p-4 {criticalStatus?.overallStatus === 'critical'
				? 'border-red-500 bg-red-50 dark:bg-red-950/30'
				: criticalStatus?.overallStatus === 'warning'
					? 'border-yellow-500 bg-yellow-50 dark:bg-yellow-950/30'
					: 'border-green-500 bg-green-50 dark:bg-green-950/30'}"
		>
			<div class="flex items-start justify-between">
				<div class="flex-1">
					<div class="flex items-center gap-2 mb-2">
						{#if criticalStatus?.overallStatus === 'critical'}
							<ShieldAlertIcon class="h-5 w-5 text-red-600" />
							<h2 class="font-semibold text-red-900 dark:text-red-100">Action Required</h2>
						{:else if criticalStatus?.overallStatus === 'warning'}
							<AlertTriangleIcon class="h-5 w-5 text-yellow-600" />
							<h2 class="font-semibold text-yellow-900 dark:text-yellow-100">Attention Needed</h2>
						{:else}
							<CheckCircleIcon class="h-5 w-5 text-green-600" />
							<h2 class="font-semibold text-green-900 dark:text-green-100">All Good</h2>
						{/if}
					</div>

					{#if criticalStatus?.issues && criticalStatus.issues.length > 0}
						<div class="space-y-2">
							{#each criticalStatus.issues as issue, index (index)}
								<div class="flex items-center justify-between">
									<span class="text-sm">{issue.message}</span>
									<Badge
										variant={issue.severity === 'critical' ? 'destructive' : 'secondary'}
										class="text-xs"
									>
										{issue.action}
									</Badge>
								</div>
							{/each}
						</div>
					{:else}
						<p class="text-sm text-green-700 dark:text-green-300">
							Community watch operations running smoothly
						</p>
					{/if}
				</div>
			</div>
		</Card.Root>

		<!-- CRITICAL: Right Now Status -->
		<Card.Root class="p-4">
			<h3 class="font-medium mb-3 flex items-center gap-2">
				<ClockIcon class="h-4 w-4" />
				Right Now
			</h3>
			<div class="space-y-3">
				<div class="flex items-center justify-between p-3 bg-muted/50 rounded-lg">
					<div>
						<p class="text-sm font-medium">Current Shift</p>
						{#if todayStatus.currentShift === 'active'}
							<p class="text-xs text-green-600">Active • {todayStatus.currentVolunteer}</p>
						{:else}
							<p class="text-xs text-muted-foreground">No active shift</p>
						{/if}
					</div>
					<div class="text-right">
						<p class="text-sm font-medium">Next Shift</p>
						<p class="text-xs text-muted-foreground">{todayStatus.nextShift}</p>
						<p
							class="text-xs {todayStatus.nextVolunteer === 'TBD'
								? 'text-red-600'
								: 'text-green-600'}"
						>
							{todayStatus.nextVolunteer}
						</p>
					</div>
				</div>
			</div>
		</Card.Root>

		<!-- IMPORTANT: Weekend & Week Ahead -->
		<div class="grid grid-cols-2 gap-3">
			<Card.Root class="p-4">
				<div class="text-center">
					<p class="text-sm font-medium mb-1">This Weekend</p>
					{#if metrics?.this_weekend_status === 'fully_covered'}
						<CheckCircleIcon class="h-8 w-8 text-green-600 mx-auto mb-1" />
						<p class="text-xs text-green-600">Fully Covered</p>
					{:else if metrics?.this_weekend_status === 'partial_coverage'}
						<AlertTriangleIcon class="h-8 w-8 text-yellow-600 mx-auto mb-1" />
						<p class="text-xs text-yellow-600">Partial Coverage</p>
					{:else}
						<AlertTriangleIcon class="h-8 w-8 text-red-600 mx-auto mb-1" />
						<p class="text-xs text-red-600">Critical</p>
					{/if}
				</div>
			</Card.Root>

			<Card.Root class="p-4">
				<div class="text-center">
					<p class="text-sm font-medium mb-1">Next Week</p>
					<div
						class="text-2xl font-bold mb-1 {(metrics?.next_week_unfilled || 0) > 3
							? 'text-red-600'
							: (metrics?.next_week_unfilled || 0) > 1
								? 'text-yellow-600'
								: 'text-green-600'}"
					>
						{metrics?.next_week_unfilled || 0}
					</div>
					<p class="text-xs text-muted-foreground">unfilled shifts</p>
				</div>
			</Card.Root>
		</div>

		<!-- IMPORTANT: Quick Actions -->
		<Card.Root class="p-4">
			<h3 class="font-medium mb-3">Quick Actions</h3>
			<div class="grid grid-cols-2 gap-3">
				{#if criticalStatus?.overallStatus === 'critical'}
					<Button
						variant="destructive"
						class="h-16 flex-col gap-1 col-span-2"
						onclick={handleEmergencyBroadcast}
					>
						<MessageSquareIcon class="h-5 w-5" />
						<span class="text-xs">🚨 Emergency Broadcast</span>
					</Button>
				{/if}

				<Button variant="outline" class="h-16 flex-col gap-1" onclick={handleFillGaps}>
					<CalendarIcon class="h-5 w-5" />
					<span class="text-xs">Fill Gaps</span>
				</Button>

				<Button variant="outline" class="h-16 flex-col gap-1" onclick={handleContactMembers}>
					<PhoneIcon class="h-5 w-5" />
					<span class="text-xs">Contact Members</span>
				</Button>

				<Button variant="outline" class="h-16 flex-col gap-1" onclick={handleViewReports}>
					<ShieldAlertIcon class="h-5 w-5" />
					<span class="text-xs">View Reports</span>
				</Button>

				<Button
					variant="outline"
					class="h-16 flex-col gap-1"
					onclick={() => (window.location.href = '/admin/emergency-contacts')}
				>
					<PhoneIcon class="h-5 w-5" />
					<span class="text-xs">Emergency Contacts</span>
				</Button>
			</div>
		</Card.Root>

		<!-- LIKE TO KNOW: Members Needing Attention (Collapsed by default) -->
		{#if needsAttention.length > 0}
			<Card.Root class="p-4">
				<h3 class="font-medium mb-3 flex items-center gap-2">
					<UsersIcon class="h-4 w-4" />
					Members Needing Attention
					<Badge variant="secondary" class="text-xs">{needsAttention.length}</Badge>
				</h3>
				<div class="space-y-2">
					{#each needsAttention as member (member.name)}
						<div
							class="flex items-center justify-between p-2 bg-yellow-50 dark:bg-yellow-950/20 rounded-lg"
						>
							<span class="text-sm font-medium">{member.name}</span>
							<Badge variant="outline" class="text-xs">No contributions</Badge>
						</div>
					{/each}
				</div>
			</Card.Root>
		{/if}

		<!-- LIKE TO KNOW: Key Stats (Minimal, collapsed) -->
		<Card.Root class="p-4">
			<h3 class="font-medium mb-3">Key Stats</h3>
			<div class="grid grid-cols-3 gap-4 text-center">
				<div>
					<p class="text-lg font-bold">{metrics?.fill_rate.toFixed(0) || 0}%</p>
					<p class="text-xs text-muted-foreground">Fill Rate</p>
				</div>
				<div>
					<p class="text-lg font-bold">{metrics?.check_in_rate.toFixed(0) || 0}%</p>
					<p class="text-xs text-muted-foreground">Show Up</p>
				</div>
				<div>
					<p class="text-lg font-bold">{qualityMetrics?.reliability_score.toFixed(0) || 0}%</p>
					<p class="text-xs text-muted-foreground">Reliable</p>
				</div>
			</div>
		</Card.Root>
	{/if}
</div>
