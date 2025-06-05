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
	import CheckCircle2Icon from '@lucide/svelte/icons/check-circle-2';

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
		<!-- Status Overview Cards -->
		<div class="grid grid-cols-2 gap-3">
			<!-- Critical Issues Count -->
			<Card.Root
				class="p-4 {statusCounts.criticalIssues > 0 ? 'border-destructive bg-destructive/5' : ''}"
			>
				<div class="flex items-center gap-3">
					<div
						class="p-2 rounded-lg {statusCounts.criticalIssues > 0
							? 'bg-red-100 dark:bg-red-900/20'
							: 'bg-green-100 dark:bg-green-900/20'}"
					>
						{#if statusCounts.criticalIssues > 0}
							<AlertTriangleIcon class="h-5 w-5 text-red-600 dark:text-red-400" />
						{:else}
							<CheckCircle2Icon class="h-5 w-5 text-green-600 dark:text-green-400" />
						{/if}
					</div>
					<div>
						<p class="text-sm font-medium text-muted-foreground">Critical Issues</p>
						<p
							class="text-2xl font-bold {statusCounts.criticalIssues > 0
								? 'text-red-600 dark:text-red-400'
								: ''}"
						>
							{statusCounts.criticalIssues}
						</p>
						<p class="text-xs text-muted-foreground">
							{statusCounts.criticalIssues === 0 ? 'All clear' : 'Need attention'}
						</p>
					</div>
				</div>
			</Card.Root>

			<!-- Pending Guests -->
			<Card.Root
				class="p-4 {statusCounts.pendingGuests > 5
					? 'border-amber-500 bg-amber-50 dark:bg-amber-950/30'
					: ''}"
			>
				<div class="flex items-center gap-3">
					<div class="p-2 bg-amber-100 dark:bg-amber-900/20 rounded-lg">
						<UsersIcon class="h-5 w-5 text-amber-600 dark:text-amber-400" />
					</div>
					<div>
						<p class="text-sm font-medium text-muted-foreground">Pending Guests</p>
						<p class="text-2xl font-bold">
							{statusCounts.pendingGuests}
						</p>
						<p class="text-xs text-muted-foreground">Awaiting approval</p>
					</div>
				</div>
			</Card.Root>

			<!-- Champions Count -->
			<Card.Root class="p-4">
				<div class="flex items-center gap-3">
					<div class="p-2 bg-yellow-100 dark:bg-yellow-900/20 rounded-lg">
						<StarIcon class="h-5 w-5 text-yellow-600 dark:text-yellow-400" />
					</div>
					<div>
						<p class="text-sm font-medium text-muted-foreground">Champions</p>
						<p class="text-2xl font-bold">
							{statusCounts.champions}
						</p>
						<p class="text-xs text-muted-foreground">Reliable volunteers</p>
					</div>
				</div>
			</Card.Root>

			<!-- Free Loaders Count -->
			<Card.Root
				class="p-4 {statusCounts.freeLoaders > 0
					? 'border-orange-500 bg-orange-50 dark:bg-orange-950/30'
					: ''}"
			>
				<div class="flex items-center gap-3">
					<div class="p-2 bg-orange-100 dark:bg-orange-900/20 rounded-lg">
						<UserXIcon class="h-5 w-5 text-orange-600 dark:text-orange-400" />
					</div>
					<div>
						<p class="text-sm font-medium text-muted-foreground">Need Follow-up</p>
						<p
							class="text-2xl font-bold {statusCounts.freeLoaders > 0
								? 'text-orange-600 dark:text-orange-400'
								: ''}"
						>
							{statusCounts.freeLoaders}
						</p>
						<p class="text-xs text-muted-foreground">Poor attendance</p>
					</div>
				</div>
			</Card.Root>
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
			<Card.Root>
				<Card.Header class="pb-3">
					<Card.Title class="flex items-center gap-2">
						<UsersIcon class="h-5 w-5 text-amber-600" />
						Guest Approvals
						<Badge variant="secondary">{insights.pendingGuests.length}</Badge>
					</Card.Title>
				</Card.Header>
				<Card.Content class="space-y-3">
					{#each insights.pendingGuests.slice(0, 4) as guest (guest.id)}
						<div
							class="flex items-center justify-between p-3 bg-amber-50 dark:bg-amber-950/20 rounded-lg hover:bg-amber-100 dark:hover:bg-amber-950/30 transition-colors border border-amber-200 dark:border-amber-800"
						>
							<div class="flex items-center gap-3">
								<div
									class="h-8 w-8 bg-amber-200 dark:bg-amber-800 rounded-full flex items-center justify-center"
								>
									<UsersIcon class="h-4 w-4 text-amber-700 dark:text-amber-300" />
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
									class="h-8"
								>
									Approve
								</Button>
								<Button
									size="sm"
									variant="outline"
									onclick={() => handleViewUser(guest.id)}
									class="h-8"
								>
									View
								</Button>
							</div>
						</div>
					{/each}
					{#if insights.pendingGuests.length > 4}
						<Button
							variant="outline"
							class="w-full"
							onclick={() => goto('/admin/users?role=guest')}
						>
							View All {insights.pendingGuests.length} Pending Guests
						</Button>
					{/if}
				</Card.Content>
			</Card.Root>
		{/if}

		<!-- Champions (Reliable Contributors) -->
		{#if insights.champions.length > 0}
			<Card.Root>
				<Card.Header class="pb-3">
					<Card.Title class="flex items-center gap-2">
						<StarIcon class="h-5 w-5 text-yellow-500" />
						Community Champions
					</Card.Title>
				</Card.Header>
				<Card.Content class="space-y-3">
					{#each insights.champions as champion (champion.user_id)}
						<div
							class="flex items-center justify-between p-3 bg-green-50 dark:bg-green-950/20 rounded-lg hover:bg-green-100 dark:hover:bg-green-950/30 transition-colors border border-green-200 dark:border-green-800"
						>
							<div class="flex items-center gap-3">
								<div
									class="h-8 w-8 bg-yellow-200 dark:bg-yellow-800 rounded-full flex items-center justify-center"
								>
									<StarIcon class="h-4 w-4 text-yellow-600 dark:text-yellow-300" />
								</div>
								<div>
									<p class="font-medium text-sm">{champion.name}</p>
									<p class="text-xs text-muted-foreground">
										{champion.shifts_booked} shifts • {champion.attendance_rate.toFixed(0)}%
										attendance
									</p>
								</div>
							</div>
							<Badge
								variant="default"
								class="bg-green-100 text-green-800 dark:bg-green-800 dark:text-green-100"
							>
								{champion.completion_rate.toFixed(0)}% complete
							</Badge>
						</div>
					{/each}
				</Card.Content>
			</Card.Root>
		{/if}

		<!-- Free Loaders (Needs Attention) -->
		{#if insights.freeLoaders.length > 0}
			<Card.Root class="border-orange-200">
				<Card.Header class="pb-3">
					<Card.Title class="flex items-center gap-2">
						<UserXIcon class="h-5 w-5 text-orange-600" />
						Needs Follow-Up
						<Badge variant="destructive">{insights.freeLoaders.length}</Badge>
					</Card.Title>
				</Card.Header>
				<Card.Content class="space-y-3">
					{#each insights.freeLoaders as freeLoader (freeLoader.user_id)}
						<div
							class="flex items-center justify-between p-3 bg-orange-50 dark:bg-orange-950/20 rounded-lg hover:bg-orange-100 dark:hover:bg-orange-950/30 transition-colors border border-orange-200 dark:border-orange-800"
						>
							<div class="flex items-center gap-3">
								<div
									class="h-8 w-8 bg-orange-200 dark:bg-orange-800 rounded-full flex items-center justify-center"
								>
									<UserXIcon class="h-4 w-4 text-orange-600 dark:text-orange-300" />
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
									class="h-8"
								>
									Contact
								</Button>
								<Button
									size="sm"
									variant="outline"
									onclick={() => handleViewUser(freeLoader.user_id)}
									class="h-8"
								>
									View
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
