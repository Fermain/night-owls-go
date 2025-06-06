<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';

	// Icons
	import BarChart3Icon from '@lucide/svelte/icons/bar-chart-3';
	import TrendingUpIcon from '@lucide/svelte/icons/trending-up';
	import UsersIcon from '@lucide/svelte/icons/users';
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';

	import { createAdminDashboardQuery } from '$lib/queries/admin/dashboard';
	import { createUsersQuery } from '$lib/queries/admin/users/usersQuery';
	import { goto } from '$app/navigation';
	import {
		calculateIntelligentInsights,
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

	// Calculate insights
	const insights = $derived.by((): IntelligentInsights | null => {
		if (!dashboardData || !usersData.length) return null;
		return calculateIntelligentInsights(dashboardData, usersData);
	});

	// Calculate analytics metrics
	const analytics = $derived.by(() => {
		if (!usersData.length) return null;

		const totalUsers = usersData.length;
		const byRole = {
			admin: usersData.filter((u) => u.role === 'admin').length,
			owl: usersData.filter((u) => u.role === 'owl').length,
			guest: usersData.filter((u) => u.role === 'guest').length
		};

		// Growth analysis (simplified)
		const recentUsers = usersData.filter((u) => {
			if (!u.created_at) return false;
			const createdDate = new Date(u.created_at);
			const thirtyDaysAgo = new Date();
			thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30);
			return createdDate >= thirtyDaysAgo;
		});

		return {
			totalUsers,
			byRole,
			growthRate: totalUsers > 0 ? (recentUsers.length / totalUsers) * 100 : 0,
			recentSignups: recentUsers.length,
			conversionRate: byRole.guest > 0 ? (byRole.owl / (byRole.owl + byRole.guest)) * 100 : 100
		};
	});
</script>

<svelte:head>
	<title>User Analytics - Mount Moreland Night Owls</title>
</svelte:head>

<div class="p-4 md:p-6 space-y-6">
	<!-- Header with back button -->
	<div class="border-b pb-4">
		<div class="flex items-center gap-3 mb-2">
			<Button variant="ghost" size="sm" onclick={() => goto('/admin/users')}>
				<ArrowLeftIcon class="h-4 w-4 mr-2" />
				Back
			</Button>
			<BarChart3Icon class="h-8 w-8 text-primary" />
			<h1 class="text-2xl md:text-3xl font-bold tracking-tight">User Analytics</h1>
		</div>
		<p class="text-base md:text-lg text-muted-foreground">
			Detailed insights and trends for Night Owls community members
		</p>
	</div>

	{#if isLoading}
		<!-- Loading state -->
		<div class="space-y-4">
			<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
				{#each Array(3) as _, i (i)}
					<Skeleton class="h-32 rounded-lg" />
				{/each}
			</div>
			<Skeleton class="h-64 w-full rounded-lg" />
		</div>
	{:else if isError}
		<Card.Root class="p-4 border-destructive bg-destructive/5">
			<div class="flex items-center gap-3">
				<AlertTriangleIcon class="h-5 w-5 text-destructive" />
				<div>
					<p class="font-medium text-destructive">Analytics Error</p>
					<p class="text-sm text-muted-foreground">Failed to load analytics data</p>
				</div>
			</div>
		</Card.Root>
	{:else if analytics}
		<!-- Key Metrics -->
		<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
			<Card.Root class="p-6">
				<div class="flex items-center justify-between">
					<div>
						<p class="text-sm font-medium text-muted-foreground">Total Users</p>
						<p class="text-3xl font-bold">{analytics.totalUsers}</p>
						<p class="text-sm text-muted-foreground">
							{analytics.recentSignups} new in last 30 days
						</p>
					</div>
					<UsersIcon class="h-8 w-8 text-blue-600" />
				</div>
			</Card.Root>

			<Card.Root class="p-6">
				<div class="flex items-center justify-between">
					<div>
						<p class="text-sm font-medium text-muted-foreground">Growth Rate</p>
						<p class="text-3xl font-bold">{analytics.growthRate.toFixed(1)}%</p>
						<p class="text-sm text-muted-foreground">Monthly growth</p>
					</div>
					<TrendingUpIcon class="h-8 w-8 text-green-600" />
				</div>
			</Card.Root>

			<Card.Root class="p-6">
				<div class="flex items-center justify-between">
					<div>
						<p class="text-sm font-medium text-muted-foreground">Volunteer Conversion</p>
						<p class="text-3xl font-bold">{analytics.conversionRate.toFixed(1)}%</p>
						<p class="text-sm text-muted-foreground">Guest to volunteer rate</p>
					</div>
					<BarChart3Icon class="h-8 w-8 text-purple-600" />
				</div>
			</Card.Root>
		</div>

		<!-- Role Distribution -->
		<Card.Root>
			<Card.Header>
				<Card.Title>User Role Distribution</Card.Title>
				<Card.Description>
					Breakdown of users by their current role in the community
				</Card.Description>
			</Card.Header>
			<Card.Content>
				<div class="space-y-4">
					<!-- Admins -->
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-3">
							<div class="w-4 h-4 bg-red-500 rounded"></div>
							<span class="font-medium">Administrators</span>
						</div>
						<div class="text-right">
							<span class="text-lg font-bold">{analytics.byRole.admin}</span>
							<span class="text-sm text-muted-foreground ml-2">
								({((analytics.byRole.admin / analytics.totalUsers) * 100).toFixed(1)}%)
							</span>
						</div>
					</div>

					<!-- Night Owls -->
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-3">
							<div class="w-4 h-4 bg-yellow-500 rounded"></div>
							<span class="font-medium">Night Owl Volunteers</span>
						</div>
						<div class="text-right">
							<span class="text-lg font-bold">{analytics.byRole.owl}</span>
							<span class="text-sm text-muted-foreground ml-2">
								({((analytics.byRole.owl / analytics.totalUsers) * 100).toFixed(1)}%)
							</span>
						</div>
					</div>

					<!-- Guests -->
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-3">
							<div class="w-4 h-4 bg-amber-500 rounded"></div>
							<span class="font-medium">Pending Guests</span>
						</div>
						<div class="text-right">
							<span class="text-lg font-bold">{analytics.byRole.guest}</span>
							<span class="text-sm text-muted-foreground ml-2">
								({((analytics.byRole.guest / analytics.totalUsers) * 100).toFixed(1)}%)
							</span>
						</div>
					</div>
				</div>
			</Card.Content>
		</Card.Root>

		<!-- Performance Insights (if available) -->
		{#if insights}
			<Card.Root>
				<Card.Header>
					<Card.Title>Community Performance</Card.Title>
					<Card.Description>Insights based on shift participation and reliability</Card.Description>
				</Card.Header>
				<Card.Content>
					<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
						<!-- Champions -->
						<div>
							<h4 class="font-medium mb-3">Top Performers ({insights.champions.length})</h4>
							{#if insights.champions.length > 0}
								<div class="space-y-2">
									{#each insights.champions.slice(0, 5) as champion (champion.user_id)}
										<div class="flex items-center justify-between text-sm">
											<span>{champion.name}</span>
											<span class="text-muted-foreground">
												{champion.attendance_rate.toFixed(0)}% attendance
											</span>
										</div>
									{/each}
								</div>
							{:else}
								<p class="text-sm text-muted-foreground">No performance data available yet</p>
							{/if}
						</div>

						<!-- Needs Attention -->
						<div>
							<h4 class="font-medium mb-3">Needs Follow-up ({insights.freeLoaders.length})</h4>
							{#if insights.freeLoaders.length > 0}
								<div class="space-y-2">
									{#each insights.freeLoaders.slice(0, 5) as user (user.user_id)}
										<div class="flex items-center justify-between text-sm">
											<span>{user.name}</span>
											<span class="text-muted-foreground">
												{user.attendance_rate.toFixed(0)}% attendance
											</span>
										</div>
									{/each}
								</div>
							{:else}
								<p class="text-sm text-muted-foreground">All users are performing well!</p>
							{/if}
						</div>
					</div>
				</Card.Content>
			</Card.Root>
		{/if}

		<!-- Note about future features -->
		<Card.Root class="border-dashed">
			<Card.Content class="p-6 text-center">
				<BarChart3Icon class="h-8 w-8 mx-auto mb-2 text-muted-foreground" />
				<p class="text-sm text-muted-foreground">
					📊 Advanced charts and trend analysis will be added once we have more production data
				</p>
			</Card.Content>
		</Card.Root>
	{/if}
</div>
