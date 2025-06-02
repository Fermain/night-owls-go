<script lang="ts">
	import MobileAdminDashboard from '$lib/components/admin/MobileAdminDashboard.svelte';
	import SidebarPage from '$lib/components/sidebar-page.svelte';
	import UpcomingShifts from '$lib/components/admin/shifts/UpcomingShifts.svelte';
	import { createAdminDashboardQuery } from '$lib/queries/admin/dashboard';

	// Create the comprehensive dashboard query
	const dashboardQuery = $derived(createAdminDashboardQuery());

	const isLoading = $derived($dashboardQuery.isLoading);
	const isError = $derived($dashboardQuery.isError);
	const error = $derived($dashboardQuery.error || undefined);
	const dashboardData = $derived($dashboardQuery.data);
</script>

<svelte:head>
	<title>Admin Dashboard - Mount Moreland Night Owls</title>
</svelte:head>

<SidebarPage title="Upcoming Shifts">
	{#snippet listContent()}
		<UpcomingShifts maxItems={8} />
	{/snippet}

	<div class="p-6 space-y-6">
		<!-- Page Header with proper admin styling -->
		<div class="border-b pb-4">
			<h1 class="text-3xl font-bold tracking-tight">Admin Dashboard</h1>
			<p class="text-lg text-muted-foreground mt-2">
				Quick insights and actions for community watch operations
			</p>
		</div>

		<!-- Mobile-First Dashboard Content with consistent padding -->
		<MobileAdminDashboard {isLoading} {isError} {error} data={dashboardData} />
	</div>
</SidebarPage>
