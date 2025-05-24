<script lang="ts">
	import AdminDashboard from '$lib/components/admin/AdminDashboard.svelte';
	import { createUsersQuery } from '$lib/queries/admin/users/usersQuery';
	import { createSchedulesQuery } from '$lib/queries/admin/schedules/schedulesQuery';
	import { createDashboardShiftsQuery } from '$lib/queries/admin/shifts/dashboardShiftsQuery';

	// Create queries for all admin data
	const usersQuery = $derived(createUsersQuery());
	const schedulesQuery = $derived(createSchedulesQuery());
	const shiftsQuery = $derived(createDashboardShiftsQuery());

	// Combined loading and error states
	const isLoading = $derived(
		$usersQuery.isLoading || $schedulesQuery.isLoading || $shiftsQuery.isLoading
	);
	const isError = $derived($usersQuery.isError || $schedulesQuery.isError || $shiftsQuery.isError);
	const error = $derived(
		$usersQuery.error || $schedulesQuery.error || $shiftsQuery.error || undefined
	);
</script>

<svelte:head>
	<title>Admin Dashboard</title>
</svelte:head>

<AdminDashboard
	{isLoading}
	{isError}
	{error}
	users={$usersQuery.data}
	schedules={$schedulesQuery.data}
	shifts={$shiftsQuery.data}
/>
