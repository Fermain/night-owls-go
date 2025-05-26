<script lang="ts">
	import AdminDashboard from '$lib/components/admin/AdminDashboard.svelte';
	import { createAdminDashboardQuery } from '$lib/queries/admin/dashboard';

	// Create the comprehensive dashboard query
	const dashboardQuery = $derived(createAdminDashboardQuery());

	const isLoading = $derived($dashboardQuery.isLoading);
	const isError = $derived($dashboardQuery.isError);
	const error = $derived($dashboardQuery.error);
	const dashboardData = $derived($dashboardQuery.data);
</script>

<svelte:head>
	<title>Admin Dashboard</title>
</svelte:head>

<AdminDashboard
	{isLoading}
	{isError}
	{error}
	data={dashboardData}
/>
