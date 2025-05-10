<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import type { Schedule } from '$lib/components/schedules_table/columns'; // Reusing Schedule type
	import { columns as publicColumns } from '$lib/components/schedules_table/columns';
	import SchedulesDataTable from '$lib/components/schedules_table/schedules-data-table.svelte';
	import { page } from '$app/state'; // To read URL params
	import ScheduleForm from '$lib/components/admin/schedules/ScheduleForm.svelte'; // Import existing form

	type AdminSchedulesAPIResponse = Schedule[];
	// Type for single schedule detail, assuming backend returns the same Schedule structure
	type ScheduleDetailAPIResponse = Schedule;

	// --- Query for the list of all schedules (for the table) ---
	const fetchAdminSchedules = async (): Promise<AdminSchedulesAPIResponse> => {
		const response = await fetch('/api/admin/schedules');
		if (!response.ok) {
			const errorText = await response.text();
			throw new Error(`API request failed: ${response.status} ${response.statusText} - ${errorText}`);
		}
		return response.json();
	};

	const adminSchedulesQuery = createQuery<AdminSchedulesAPIResponse, Error, AdminSchedulesAPIResponse, string[]>({
		queryKey: ['adminSchedules'],
		queryFn: fetchAdminSchedules
	});

	let tableColumns = publicColumns;
	let tableData: AdminSchedulesAPIResponse = $derived($adminSchedulesQuery.data ?? []);

	// --- Logic for fetching and displaying a single selected schedule ---
	let selectedScheduleId = $derived(page.url.searchParams.get('scheduleId'));

	const fetchScheduleDetail = async (id: string): Promise<ScheduleDetailAPIResponse> => {
		const response = await fetch(`/api/admin/schedules/${id}`);
		if (!response.ok) {
			const errorText = await response.text();
			throw new Error(`API request failed for schedule ${id}: ${response.status} ${response.statusText} - ${errorText}`);
		}
		return response.json();
	};

	const scheduleDetailQuery = $derived(createQuery<ScheduleDetailAPIResponse, Error, ScheduleDetailAPIResponse, [string, string | null]>({
		queryKey: ['adminScheduleDetail', selectedScheduleId], // Reactive query key
		queryFn: () => fetchScheduleDetail(selectedScheduleId!), // Assert non-null as it's enabled only when id is present
		enabled: !!selectedScheduleId // Only run query if selectedScheduleId has a value
	}));

</script>

<svelte:head>
	<title>Admin - {selectedScheduleId ? `Schedule Details` : 'All Schedules'}</title>
</svelte:head>

<div class="container mx-auto p-4">
	{#if selectedScheduleId}
		<!-- Displaying detail/form for a selected schedule -->
		{#if $scheduleDetailQuery.isLoading}
			<p>Loading schedule details for ID: {selectedScheduleId}...</p>
		{:else if $scheduleDetailQuery.isError}
			<p class="text-red-500">Error fetching schedule details: {$scheduleDetailQuery.error?.message}</p>
		{:else if $scheduleDetailQuery.data}
			<!-- Use ScheduleForm to display the selected schedule -->
			<ScheduleForm schedule={$scheduleDetailQuery.data} />
		{:else}
			<p>No data for schedule ID: {selectedScheduleId}.</p>
		{/if}
	{:else}
		<!-- Displaying the table of all schedules -->
		{#if $adminSchedulesQuery.isLoading}
			<p>Loading schedules...</p>
		{:else if $adminSchedulesQuery.isError}
			<p class="text-red-500">Error fetching schedules: {$adminSchedulesQuery.error?.message}</p>
			{#if $adminSchedulesQuery.error?.message?.includes('Failed to decode request body')}
				<p class="text-sm text-gray-600 mt-1">
					This might indicate an issue with the request sent by the client or how the server expects
					the data.
				</p>
			{/if}
			{#if $adminSchedulesQuery.error?.message?.includes('Failed to create schedule') || $adminSchedulesQuery.error?.message?.includes('Failed to list schedules')}
				<p class="text-sm text-gray-600 mt-1">
					This often points to a server-side or database issue. Check server logs.
				</p>
			{/if}
		{:else if $adminSchedulesQuery.data}
			{#if tableData.length === 0}
				<p>
					No schedules found. <a href="/admin/schedules/new" class="text-blue-600 hover:underline">Create the first one!</a>
				</p>
			{:else}
				<SchedulesDataTable columns={tableColumns} data={tableData} />
			{/if}
		{:else}
			<p>No schedule data available.</p>
		{/if}
	{/if}
</div>
