<script lang="ts">
	import AdminDashboard from '$lib/components/admin/AdminDashboard.svelte';
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
	<title>Admin Dashboard</title>
</svelte:head>

<SidebarPage title="Upcoming Shifts">
	{#snippet listContent()}
		<UpcomingShifts maxItems={8} />
	{/snippet}

	{#snippet children()}
		<AdminDashboard {isLoading} {isError} {error} data={dashboardData} />
	{/snippet}
</SidebarPage>
