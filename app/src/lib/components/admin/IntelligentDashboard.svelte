<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Skeleton } from '$lib/components/ui/skeleton';

	// Icons
	import UserXIcon from '@lucide/svelte/icons/user-x';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import StarIcon from '@lucide/svelte/icons/star';
	import ShieldAlertIcon from '@lucide/svelte/icons/shield-alert';
	import UsersIcon from '@lucide/svelte/icons/users';
	import MessageSquareIcon from '@lucide/svelte/icons/message-square';

	import { createAdminDashboardQuery } from '$lib/queries/admin/dashboard';
	import { createUsersQuery } from '$lib/queries/admin/users/usersQuery';
	import { goto } from '$app/navigation';
	import { toast } from 'svelte-sonner';
	import { UsersApiService } from '$lib/services/api';
	import {
		calculateIntelligentInsights,
		getStatusCounts,
		type IntelligentInsights
	} from '$lib/utils/intelligentDashboard';

	// Reactive queries
	const dashboardQuery = $derived(createAdminDashboardQuery());
	const usersQuery = $derived(createUsersQuery());

	// Derived states
	const isLoading = $derived($dashboardQuery.isLoading || $usersQuery.isLoading);
	const isError = $derived($dashboardQuery.isError || $usersQuery.isError);
	const dashboardData = $derived($dashboardQuery.data);
	const usersData = $derived($usersQuery.data || []);

	// Calculate insights using utils
	const insights = $derived.by((): IntelligentInsights | null => {
		if (!dashboardData || !usersData.length) return null;
		return calculateIntelligentInsights(dashboardData, usersData);
	});

	// Calculate status counts for overview cards
	const statusCounts = $derived.by(() => {
		if (!insights) return null;
		return getStatusCounts(insights);
	});

	// Quick actions - following your patterns
	async function handleApproveGuest(userId: number, userName?: string) {
		try {
			// Get current user data first since update requires all fields
			const currentUser = await UsersApiService.getById(userId);

			// Update user with new role while keeping existing data
			await UsersApiService.update(userId, {
				name: currentUser.name || '',
				phone: currentUser.phone,
				role: 'owl'
			});

			toast.success(`${userName ?? 'User'} has been approved as an Owl volunteer!`);
			window.location.reload();
		} catch (_error) {
			toast.error('Failed to approve guest user');
		}
	}

	function handleViewUser(userId: number) {
		goto(`/admin/users?userId=${userId}`);
	}

	function handleContactUser(userId: number, reason: string, userName?: string) {
		toast.info(`Contact ${userName ?? 'user'} about: ${reason}`);
	}

	function handleCriticalAction(type: string) {
		switch (type) {
			case 'unfilled_shifts':
				window.location.href = '/admin/shifts';
				break;
			case 'high_no_show':
				window.location.href = '/admin/users';
				break;
			case 'guest_backlog':
				goto('/admin/users?role=guest');
				break;
		}
	}
</script>

<!-- Mobile-First Intelligent Dashboard -->
<div class="space-y-4">
	{#if isLoading}
		<!-- Loading state -->
		<div class="space-y-4">
			<Skeleton class="h-24 w-full rounded-lg" />
			<div class="grid grid-cols-2 gap-3">
				{#each Array(4) as _, i (i)}
					<Skeleton class="h-20 rounded-lg" />
				{/each}
			</div>
			<Skeleton class="h-32 w-full rounded-lg" />
		</div>
	{:else if isError}
		<Card.Root class="p-4 border-destructive bg-destructive/5">
			<div class="flex items-center gap-3">
				<AlertTriangleIcon class="h-5 w-5 text-destructive" />
				<div>
					<p class="font-medium text-destructive">Dashboard Error</p>
					<p class="text-sm text-muted-foreground">Failed to load dashboard data</p>
				</div>
			</div>
		</Card.Root>
	{:else if insights && statusCounts}
		<!-- Compact Status Overview -->
		<div class="flex gap-4 text-center">
			<div class="flex-1">
				<p
					class="text-lg font-bold {statusCounts.criticalIssues > 0
						? 'text-red-600'
						: 'text-green-600'}"
				>
					{statusCounts.criticalIssues}
				</p>
				<p class="text-xs text-muted-foreground">Critical</p>
			</div>
			<div class="flex-1">
				<p class="text-lg font-bold text-amber-600">{statusCounts.pendingGuests}</p>
				<p class="text-xs text-muted-foreground">Pending</p>
			</div>
			<div class="flex-1">
				<p class="text-lg font-bold text-green-600">{statusCounts.champions}</p>
				<p class="text-xs text-muted-foreground">Champions</p>
			</div>
			<div class="flex-1">
				<p class="text-lg font-bold {statusCounts.freeLoaders > 0 ? 'text-orange-600' : ''}">
					{statusCounts.freeLoaders}
				</p>
				<p class="text-xs text-muted-foreground">Follow-up</p>
			</div>
		</div>

		<!-- Critical Issues (Always Show First) -->
		{#if insights.criticalIssues.length > 0}
			<Card.Root class="border-destructive">
				<Card.Header class="pb-3">
					<Card.Title class="flex items-center gap-2 text-destructive">
						<ShieldAlertIcon class="h-5 w-5" />
						Critical Issues
					</Card.Title>
				</Card.Header>
				<Card.Content class="space-y-3">
					{#each insights.criticalIssues as issue (issue.type)}
						<div class="border border-destructive/50 rounded-lg p-3 bg-destructive/5">
							<div class="flex items-start gap-2">
								<AlertTriangleIcon class="h-4 w-4 text-destructive mt-0.5" />
								<div class="flex-1">
									<p class="text-sm font-medium text-destructive">{issue.message}</p>
									<Button
										size="sm"
										variant="destructive"
										class="mt-2"
										onclick={() => handleCriticalAction(issue.type)}
									>
										{issue.action}
									</Button>
								</div>
							</div>
						</div>
					{/each}
				</Card.Content>
			</Card.Root>
		{/if}

		<!-- Pending Guest Approvals -->
		{#if insights.pendingGuests.length > 0}
			<Card.Root class="border-amber-200 bg-amber-50 dark:bg-amber-950/20">
				<Card.Header class="pb-3">
					<Card.Title class="text-base flex items-center gap-2">
						<UsersIcon class="h-4 w-4 text-amber-600" />
						Guest Approvals
						<Badge class="bg-amber-600 text-white">{insights.pendingGuests.length}</Badge>
					</Card.Title>
				</Card.Header>
				<Card.Content class="space-y-3">
					{#each insights.pendingGuests.slice(0, 4) as guest (guest.id)}
						<div
							class="flex items-center justify-between p-3 bg-white dark:bg-amber-900/10 rounded border"
						>
							<div class="flex items-center gap-3">
								<div
									class="h-8 w-8 bg-amber-100 dark:bg-amber-800 rounded-full flex items-center justify-center"
								>
									<UsersIcon class="h-4 w-4 text-amber-600" />
								</div>
								<div>
									<p class="font-medium text-sm">{guest.name ?? 'Unnamed User'}</p>
									<p class="text-xs text-muted-foreground">{guest.phone}</p>
								</div>
							</div>
							<div class="flex gap-1">
								<Button
									size="sm"
									onclick={() => handleApproveGuest(guest.id, guest.name ?? undefined)}
									class="h-8 bg-green-600 hover:bg-green-700 text-white"
								>
									✓
								</Button>
								<Button
									size="sm"
									variant="outline"
									onclick={() => handleViewUser(guest.id)}
									class="h-8 text-xs"
								>
									Edit
								</Button>
							</div>
						</div>
					{/each}
					{#if insights.pendingGuests.length > 4}
						<Button
							variant="outline"
							class="w-full text-xs"
							onclick={() => goto('/admin/users?role=guest')}
						>
							View All {insights.pendingGuests.length} Pending
						</Button>
					{/if}
				</Card.Content>
			</Card.Root>
		{/if}

		<!-- Champions (Reliable Contributors) -->
		{#if insights.champions.length > 0}
			<Card.Root>
				<Card.Header class="pb-3">
					<Card.Title class="text-base flex items-center gap-2">
						<StarIcon class="h-4 w-4 text-green-600" />
						Community Champions
					</Card.Title>
				</Card.Header>
				<Card.Content class="space-y-2">
					{#each insights.champions as champion (champion.user_id)}
						<div class="flex items-center justify-between p-3 rounded border">
							<div class="flex items-center gap-3">
								<div
									class="h-8 w-8 bg-green-100 dark:bg-green-800 rounded-full flex items-center justify-center"
								>
									<StarIcon class="h-4 w-4 text-green-600" />
								</div>
								<div>
									<p class="font-medium text-sm">{champion.name}</p>
									<p class="text-xs text-muted-foreground">
										{champion.shifts_booked} shifts • {champion.attendance_rate.toFixed(0)}%
										attendance
									</p>
								</div>
							</div>
							<div class="text-right">
								<p class="text-sm font-medium text-green-600">
									{champion.completion_rate.toFixed(0)}%
								</p>
							</div>
						</div>
					{/each}
				</Card.Content>
			</Card.Root>
		{/if}

		<!-- Free Loaders (Needs Attention) -->
		{#if insights.freeLoaders.length > 0}
			<Card.Root class="border-orange-200 bg-orange-50 dark:bg-orange-950/20">
				<Card.Header class="pb-3">
					<Card.Title class="text-base flex items-center gap-2">
						<UserXIcon class="h-4 w-4 text-orange-600" />
						Needs Follow-Up
						<Badge class="bg-orange-600 text-white">{insights.freeLoaders.length}</Badge>
					</Card.Title>
				</Card.Header>
				<Card.Content class="space-y-2">
					{#each insights.freeLoaders as freeLoader (freeLoader.user_id)}
						<div
							class="flex items-center justify-between p-3 bg-white dark:bg-orange-900/10 rounded border"
						>
							<div class="flex items-center gap-3">
								<div
									class="h-8 w-8 bg-orange-100 dark:bg-orange-800 rounded-full flex items-center justify-center"
								>
									<UserXIcon class="h-4 w-4 text-orange-600" />
								</div>
								<div>
									<p class="font-medium text-sm">{freeLoader.name}</p>
									<p class="text-xs text-muted-foreground">
										{freeLoader.shifts_booked} booked • {freeLoader.attendance_rate.toFixed(0)}%
										attendance
									</p>
								</div>
							</div>
							<div class="flex gap-1">
								<Button
									size="sm"
									variant="outline"
									onclick={() =>
										handleContactUser(freeLoader.user_id, 'poor_attendance', freeLoader.name)}
									class="h-8 text-xs"
								>
									Contact
								</Button>
								<Button
									size="sm"
									variant="outline"
									onclick={() => handleViewUser(freeLoader.user_id)}
									class="h-8 text-xs"
								>
									Edit
								</Button>
							</div>
						</div>
					{/each}
				</Card.Content>
			</Card.Root>
		{/if}

		<!-- Top Reporters -->
		{#if insights.topReporters.length > 0}
			<Card.Root>
				<Card.Header class="pb-3">
					<Card.Title class="flex items-center gap-2">
						<MessageSquareIcon class="h-5 w-5 text-blue-600" />
						Top Reporters
					</Card.Title>
				</Card.Header>
				<Card.Content class="space-y-3">
					{#each insights.topReporters as reporter (reporter.user_id)}
						<div
							class="flex items-center justify-between p-3 bg-blue-50 dark:bg-blue-950/20 rounded-lg hover:bg-blue-100 dark:hover:bg-blue-950/30 transition-colors border border-blue-200 dark:border-blue-800"
						>
							<div class="flex items-center gap-3">
								<div
									class="h-8 w-8 bg-blue-200 dark:bg-blue-800 rounded-full flex items-center justify-center"
								>
									<MessageSquareIcon class="h-4 w-4 text-blue-600 dark:text-blue-300" />
								</div>
								<div>
									<p class="font-medium text-sm">{reporter.name}</p>
									<p class="text-xs text-muted-foreground">
										{reporter.shifts_completed} reports completed
									</p>
								</div>
							</div>
							<Badge
								variant="secondary"
								class="bg-blue-100 text-blue-800 dark:bg-blue-800 dark:text-blue-100"
							>
								{reporter.shifts_completed} reports
							</Badge>
						</div>
					{/each}
				</Card.Content>
			</Card.Root>
		{/if}
	{/if}
</div>
