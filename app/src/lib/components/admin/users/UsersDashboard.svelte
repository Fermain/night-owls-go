<script lang="ts">
	import { Skeleton } from '$lib/components/ui/skeleton';
	import UserMetrics from './UserMetrics.svelte';
	import UserRoleChart from './UserRoleChart.svelte';
	import UserGrowthChart from './UserGrowthChart.svelte';
	import RecentUsers from './RecentUsers.svelte';
	import UserShiftMetrics from './UserShiftMetrics.svelte';
	import TopVolunteers from './TopVolunteers.svelte';
	import ShiftDistributionChart from './ShiftDistributionChart.svelte';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import type { UserData } from '$lib/schemas/user';
	import type { AdminShiftSlot } from '$lib/types';
	import { 
		calculateUserMetrics, 
		generateUserGrowthData, 
		getRecentUsers,
		calculateUserShiftMetrics
	} from '$lib/utils/userProcessing';

	let { 
		isLoading, 
		isError, 
		error, 
		users,
		shifts
	}: { 
		isLoading: boolean;
		isError: boolean;
		error?: Error;
		users?: UserData[];
		shifts?: AdminShiftSlot[];
	} = $props();

	// Calculate metrics and data for charts
	const userMetrics = $derived.by(() => {
		if (!users || users.length === 0) return null;
		return calculateUserMetrics(users);
	});

	const shiftMetrics = $derived.by(() => {
		if (!users || !shifts) return null;
		return calculateUserShiftMetrics(users, shifts);
	});

	const growthData = $derived.by(() => {
		if (!users) return [];
		return generateUserGrowthData(users);
	});

	const recentUsers = $derived.by(() => {
		if (!users) return [];
		return getRecentUsers(users, 7);
	});
</script>

<div class="p-8">
	<div class="max-w-full mx-auto">
		<div class="mb-8">
			<h1 class="text-3xl font-semibold mb-3">Users Dashboard</h1>
			<p class="text-muted-foreground text-lg">
				Overview of user registrations, roles, and shift distribution
			</p>
		</div>

		{#if isLoading}
			<!-- Loading Dashboard Skeletons -->
			<div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6 mb-8">
				{#each Array(4) as _, i (i)}
					<div class="p-6 border rounded-lg">
						<Skeleton class="h-4 w-24 mb-2" />
						<Skeleton class="h-8 w-16 mb-1" />
						<Skeleton class="h-3 w-20" />
					</div>
				{/each}
			</div>
			<!-- Shift metrics skeleton -->
			<div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6 mb-8">
				{#each Array(4) as _, i (i)}
					<div class="p-6 border rounded-lg">
						<Skeleton class="h-4 w-24 mb-2" />
						<Skeleton class="h-8 w-16 mb-1" />
						<Skeleton class="h-3 w-20" />
					</div>
				{/each}
			</div>
			<div class="grid grid-cols-1 xl:grid-cols-2 2xl:grid-cols-3 gap-8">
				{#each Array(4) as _, i (i)}
					<div class="p-6 border rounded-lg">
						<Skeleton class="h-6 w-32 mb-4" />
						<Skeleton class="h-64 w-full" />
					</div>
				{/each}
			</div>
		{:else if isError}
			<div class="text-center py-16">
				<p class="text-destructive text-lg mb-2">Error Loading Dashboard</p>
				<p class="text-muted-foreground">
					{error?.message || 'Unknown error occurred'}
				</p>
			</div>
		{:else if !users || users.length === 0}
			<div class="text-center py-16">
				<div class="max-w-md mx-auto">
					<CalendarIcon class="h-16 w-16 text-muted-foreground mx-auto mb-6" />
					<h2 class="text-2xl font-semibold mb-3">No Users Found</h2>
					<p class="text-muted-foreground mb-8">
						No users have been registered yet. Create the first user to get started.
					</p>
					<a 
						href="/admin/users/new"
						class="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 bg-primary text-primary-foreground hover:bg-primary/90 h-11 px-8 py-2"
					>
						Create First User
					</a>
				</div>
			</div>
		{:else if userMetrics}
			<!-- User Registration Metrics -->
			<UserMetrics metrics={userMetrics} />

			<!-- Shift Distribution Metrics -->
			{#if shiftMetrics}
				<UserShiftMetrics metrics={shiftMetrics} />
			{/if}

			<!-- Charts Grid with enhanced layout -->
			<div class="grid grid-cols-1 xl:grid-cols-12 gap-8">
				<!-- Top row: Role chart and Growth chart -->
				<div class="xl:col-span-4">
					<UserRoleChart metrics={userMetrics} />
				</div>
				<div class="xl:col-span-4">
					<UserGrowthChart data={growthData} />
				</div>
				<div class="xl:col-span-4">
					<RecentUsers users={recentUsers} />
				</div>
				
				<!-- Bottom row: Shift distribution components -->
				{#if shiftMetrics}
					<div class="xl:col-span-8">
						<ShiftDistributionChart distribution={shiftMetrics.shiftDistribution} />
					</div>
					<div class="xl:col-span-4">
						<TopVolunteers volunteers={shiftMetrics.topVolunteers} />
					</div>
				{/if}
			</div>
		{:else}
			<!-- Fallback for unexpected states -->
			<div class="text-center py-16">
				<p class="text-muted-foreground">Unable to display dashboard data.</p>
			</div>
		{/if}
	</div>
</div> 