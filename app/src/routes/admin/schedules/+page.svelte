<script lang="ts">
	// import { createQuery } from '@tanstack/svelte-query'; // No longer needed for adminSchedulesQuery here
	// import { columns as publicColumns } from '$lib/components/schedules_table/columns'; // Not used for dashboard
	// import SchedulesDataTable from '$lib/components/schedules_table/schedules-data-table.svelte'; // Not used for dashboard
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import ScheduleForm from '$lib/components/admin/schedules/ScheduleForm.svelte';
	import SchedulesDashboard from '$lib/components/admin/schedules/SchedulesDashboard.svelte';
	import { selectedScheduleForForm } from '$lib/stores/scheduleEditingStore';
	import { createSchedulesQuery } from '$lib/queries/admin/schedules/schedulesQuery';
	import { createDashboardShiftsQuery } from '$lib/queries/admin/shifts/dashboardShiftsQuery';

	// Get the currently selected schedule from the store
	let currentScheduleToEdit = $derived($selectedScheduleForForm);

	// Create queries for dashboard data
	const schedulesQuery = $derived(createSchedulesQuery());
	const shiftsQuery = $derived(createDashboardShiftsQuery());

	// Combined loading and error states for dashboard
	const isLoading = $derived($schedulesQuery.isLoading || $shiftsQuery.isLoading);
	const isError = $derived($schedulesQuery.isError || $shiftsQuery.isError);
	const error = $derived($schedulesQuery.error || $shiftsQuery.error || undefined);

	function handleCreateNew() {
		selectedScheduleForForm.set(undefined);
		goto('/admin/schedules/new');
	}
</script>

<svelte:head>
	<title>
		Admin - {currentScheduleToEdit
			? 'Edit Schedule'
			: page.url.pathname.endsWith('/new')
				? 'New Schedule'
				: 'Schedules Dashboard'}
	</title>
</svelte:head>

<div class="container mx-auto p-4">
	{#if currentScheduleToEdit}
		<ScheduleForm schedule={currentScheduleToEdit} />
	{:else if page.url.pathname.endsWith('/new')}
		<ScheduleForm />
	{:else}
		<!-- Schedules Dashboard View -->
		<SchedulesDashboard
			{isLoading}
			{isError}
			{error}
			schedules={$schedulesQuery.data}
			shifts={$shiftsQuery.data}
		/>
	{/if}
</div>
